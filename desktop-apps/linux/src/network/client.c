#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
#include <json-c/json.h>
#include <glib.h>
#include <pthread.h>
#include "../../include/network/client.h"

// Response data structure
typedef struct {
    char *data;
    size_t size;
} HttpResponse;

// WebSocket connection structure
typedef struct {
    char *url;
    char *protocol;
    gboolean connected;
    pthread_t thread;
    
    // Callbacks
    void (*on_open)(void *user_data);
    void (*on_message)(const char *message, void *user_data);
    void (*on_close)(int code, const char *reason, void *user_data);
    void (*on_error)(const char *error, void *user_data);
    void *user_data;
} WebSocketConnection;

// Network client structure
typedef struct {
    CURL *curl_handle;
    struct curl_slist *headers;
    char *base_url;
    char *auth_token;
    
    // WebSocket connections
    GHashTable *websocket_connections;
    
    // Configuration
    long timeout;
    gboolean verify_ssl;
    char *user_agent;
} TaishangNetworkClient;

static TaishangNetworkClient *network_client = NULL;

// Forward declarations
static size_t write_callback(void *contents, size_t size, size_t nmemb, HttpResponse *response);
static void *websocket_thread_func(void *arg);
static void websocket_connection_free(WebSocketConnection *conn);

// Public functions
gboolean taishang_network_client_init(const char *base_url);
void taishang_network_client_cleanup(void);
TaishangNetworkClient *taishang_network_client_get_instance(void);

// HTTP functions
TaishangHttpResponse *taishang_network_client_get(const char *endpoint, GHashTable *params);
TaishangHttpResponse *taishang_network_client_post(const char *endpoint, const char *data, const char *content_type);
TaishangHttpResponse *taishang_network_client_put(const char *endpoint, const char *data, const char *content_type);
TaishangHttpResponse *taishang_network_client_delete(const char *endpoint);
void taishang_http_response_free(TaishangHttpResponse *response);

// WebSocket functions
gboolean taishang_network_client_websocket_connect(const char *url, const char *protocol,
                                                   void (*on_open)(void *user_data),
                                                   void (*on_message)(const char *message, void *user_data),
                                                   void (*on_close)(int code, const char *reason, void *user_data),
                                                   void (*on_error)(const char *error, void *user_data),
                                                   void *user_data);
gboolean taishang_network_client_websocket_send(const char *url, const char *message);
void taishang_network_client_websocket_close(const char *url);

// Configuration functions
void taishang_network_client_set_auth_token(const char *token);
void taishang_network_client_set_timeout(long timeout_seconds);
void taishang_network_client_set_verify_ssl(gboolean verify);
void taishang_network_client_add_header(const char *header);

// Implementation
gboolean taishang_network_client_init(const char *base_url) {
    if (network_client != NULL) {
        g_warning("Network client already initialized");
        return FALSE;
    }
    
    // Initialize curl
    if (curl_global_init(CURL_GLOBAL_DEFAULT) != CURLE_OK) {
        g_error("Failed to initialize curl");
        return FALSE;
    }
    
    network_client = g_new0(TaishangNetworkClient, 1);
    network_client->curl_handle = curl_easy_init();
    
    if (!network_client->curl_handle) {
        g_error("Failed to initialize curl handle");
        g_free(network_client);
        network_client = NULL;
        return FALSE;
    }
    
    // Set default configuration
    network_client->base_url = g_strdup(base_url ? base_url : "http://localhost:8080");
    network_client->timeout = 30;
    network_client->verify_ssl = TRUE;
    network_client->user_agent = g_strdup("TaishangApp/1.0");
    network_client->websocket_connections = g_hash_table_new_full(g_str_hash, g_str_equal, 
                                                                  g_free, (GDestroyNotify)websocket_connection_free);
    
    // Set curl options
    curl_easy_setopt(network_client->curl_handle, CURLOPT_USERAGENT, network_client->user_agent);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_TIMEOUT, network_client->timeout);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_SSL_VERIFYPEER, network_client->verify_ssl ? 1L : 0L);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_WRITEFUNCTION, write_callback);
    
    g_print("Network client initialized with base URL: %s\n", network_client->base_url);
    return TRUE;
}

void taishang_network_client_cleanup(void) {
    if (network_client == NULL) {
        return;
    }
    
    // Close all WebSocket connections
    if (network_client->websocket_connections) {
        g_hash_table_destroy(network_client->websocket_connections);
    }
    
    // Cleanup curl
    if (network_client->headers) {
        curl_slist_free_all(network_client->headers);
    }
    
    if (network_client->curl_handle) {
        curl_easy_cleanup(network_client->curl_handle);
    }
    
    curl_global_cleanup();
    
    // Free memory
    g_free(network_client->base_url);
    g_free(network_client->auth_token);
    g_free(network_client->user_agent);
    g_free(network_client);
    network_client = NULL;
    
    g_print("Network client cleaned up\n");
}

TaishangNetworkClient *taishang_network_client_get_instance(void) {
    return network_client;
}

// HTTP implementation
static size_t write_callback(void *contents, size_t size, size_t nmemb, HttpResponse *response) {
    size_t real_size = size * nmemb;
    
    response->data = realloc(response->data, response->size + real_size + 1);
    if (response->data == NULL) {
        g_error("Failed to allocate memory for HTTP response");
        return 0;
    }
    
    memcpy(&(response->data[response->size]), contents, real_size);
    response->size += real_size;
    response->data[response->size] = 0;
    
    return real_size;
}

TaishangHttpResponse *taishang_network_client_get(const char *endpoint, GHashTable *params) {
    if (!network_client || !endpoint) {
        return NULL;
    }
    
    // Build URL
    GString *url = g_string_new(network_client->base_url);
    if (endpoint[0] != '/') {
        g_string_append_c(url, '/');
    }
    g_string_append(url, endpoint);
    
    // Add query parameters
    if (params && g_hash_table_size(params) > 0) {
        g_string_append_c(url, '?');
        gboolean first = TRUE;
        
        GHashTableIter iter;
        gpointer key, value;
        g_hash_table_iter_init(&iter, params);
        
        while (g_hash_table_iter_next(&iter, &key, &value)) {
            if (!first) {
                g_string_append_c(url, '&');
            }
            
            char *encoded_key = curl_easy_escape(network_client->curl_handle, (char *)key, 0);
            char *encoded_value = curl_easy_escape(network_client->curl_handle, (char *)value, 0);
            
            g_string_append_printf(url, "%s=%s", encoded_key, encoded_value);
            
            curl_free(encoded_key);
            curl_free(encoded_value);
            first = FALSE;
        }
    }
    
    // Prepare response
    HttpResponse response = {0};
    
    // Set curl options
    curl_easy_setopt(network_client->curl_handle, CURLOPT_URL, url->str);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_HTTPGET, 1L);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_HTTPHEADER, network_client->headers);
    
    // Perform request
    CURLcode res = curl_easy_perform(network_client->curl_handle);
    
    // Get response info
    long response_code = 0;
    curl_easy_getinfo(network_client->curl_handle, CURLINFO_RESPONSE_CODE, &response_code);
    
    // Create response object
    TaishangHttpResponse *http_response = g_new0(TaishangHttpResponse, 1);
    http_response->status_code = (int)response_code;
    http_response->data = response.data;
    http_response->size = response.size;
    http_response->success = (res == CURLE_OK && response_code >= 200 && response_code < 300);
    
    if (res != CURLE_OK) {
        http_response->error = g_strdup(curl_easy_strerror(res));
    }
    
    g_string_free(url, TRUE);
    g_print("GET %s -> %d\n", endpoint, http_response->status_code);
    
    return http_response;
}

TaishangHttpResponse *taishang_network_client_post(const char *endpoint, const char *data, const char *content_type) {
    if (!network_client || !endpoint) {
        return NULL;
    }
    
    // Build URL
    GString *url = g_string_new(network_client->base_url);
    if (endpoint[0] != '/') {
        g_string_append_c(url, '/');
    }
    g_string_append(url, endpoint);
    
    // Prepare response
    HttpResponse response = {0};
    
    // Set content type header
    struct curl_slist *headers = network_client->headers;
    if (content_type) {
        char *content_type_header = g_strdup_printf("Content-Type: %s", content_type);
        headers = curl_slist_append(headers, content_type_header);
        g_free(content_type_header);
    }
    
    // Set curl options
    curl_easy_setopt(network_client->curl_handle, CURLOPT_URL, url->str);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_POST, 1L);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_POSTFIELDS, data);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_POSTFIELDSIZE, data ? strlen(data) : 0);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_HTTPHEADER, headers);
    
    // Perform request
    CURLcode res = curl_easy_perform(network_client->curl_handle);
    
    // Get response info
    long response_code = 0;
    curl_easy_getinfo(network_client->curl_handle, CURLINFO_RESPONSE_CODE, &response_code);
    
    // Create response object
    TaishangHttpResponse *http_response = g_new0(TaishangHttpResponse, 1);
    http_response->status_code = (int)response_code;
    http_response->data = response.data;
    http_response->size = response.size;
    http_response->success = (res == CURLE_OK && response_code >= 200 && response_code < 300);
    
    if (res != CURLE_OK) {
        http_response->error = g_strdup(curl_easy_strerror(res));
    }
    
    // Cleanup temporary headers
    if (content_type && headers != network_client->headers) {
        curl_slist_free_all(headers);
    }
    
    g_string_free(url, TRUE);
    g_print("POST %s -> %d\n", endpoint, http_response->status_code);
    
    return http_response;
}

TaishangHttpResponse *taishang_network_client_put(const char *endpoint, const char *data, const char *content_type) {
    if (!network_client || !endpoint) {
        return NULL;
    }
    
    // Build URL
    GString *url = g_string_new(network_client->base_url);
    if (endpoint[0] != '/') {
        g_string_append_c(url, '/');
    }
    g_string_append(url, endpoint);
    
    // Prepare response
    HttpResponse response = {0};
    
    // Set content type header
    struct curl_slist *headers = network_client->headers;
    if (content_type) {
        char *content_type_header = g_strdup_printf("Content-Type: %s", content_type);
        headers = curl_slist_append(headers, content_type_header);
        g_free(content_type_header);
    }
    
    // Set curl options
    curl_easy_setopt(network_client->curl_handle, CURLOPT_URL, url->str);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_CUSTOMREQUEST, "PUT");
    curl_easy_setopt(network_client->curl_handle, CURLOPT_POSTFIELDS, data);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_POSTFIELDSIZE, data ? strlen(data) : 0);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_HTTPHEADER, headers);
    
    // Perform request
    CURLcode res = curl_easy_perform(network_client->curl_handle);
    
    // Get response info
    long response_code = 0;
    curl_easy_getinfo(network_client->curl_handle, CURLINFO_RESPONSE_CODE, &response_code);
    
    // Create response object
    TaishangHttpResponse *http_response = g_new0(TaishangHttpResponse, 1);
    http_response->status_code = (int)response_code;
    http_response->data = response.data;
    http_response->size = response.size;
    http_response->success = (res == CURLE_OK && response_code >= 200 && response_code < 300);
    
    if (res != CURLE_OK) {
        http_response->error = g_strdup(curl_easy_strerror(res));
    }
    
    // Cleanup temporary headers
    if (content_type && headers != network_client->headers) {
        curl_slist_free_all(headers);
    }
    
    g_string_free(url, TRUE);
    g_print("PUT %s -> %d\n", endpoint, http_response->status_code);
    
    return http_response;
}

TaishangHttpResponse *taishang_network_client_delete(const char *endpoint) {
    if (!network_client || !endpoint) {
        return NULL;
    }
    
    // Build URL
    GString *url = g_string_new(network_client->base_url);
    if (endpoint[0] != '/') {
        g_string_append_c(url, '/');
    }
    g_string_append(url, endpoint);
    
    // Prepare response
    HttpResponse response = {0};
    
    // Set curl options
    curl_easy_setopt(network_client->curl_handle, CURLOPT_URL, url->str);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_CUSTOMREQUEST, "DELETE");
    curl_easy_setopt(network_client->curl_handle, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(network_client->curl_handle, CURLOPT_HTTPHEADER, network_client->headers);
    
    // Perform request
    CURLcode res = curl_easy_perform(network_client->curl_handle);
    
    // Get response info
    long response_code = 0;
    curl_easy_getinfo(network_client->curl_handle, CURLINFO_RESPONSE_CODE, &response_code);
    
    // Create response object
    TaishangHttpResponse *http_response = g_new0(TaishangHttpResponse, 1);
    http_response->status_code = (int)response_code;
    http_response->data = response.data;
    http_response->size = response.size;
    http_response->success = (res == CURLE_OK && response_code >= 200 && response_code < 300);
    
    if (res != CURLE_OK) {
        http_response->error = g_strdup(curl_easy_strerror(res));
    }
    
    g_string_free(url, TRUE);
    g_print("DELETE %s -> %d\n", endpoint, http_response->status_code);
    
    return http_response;
}

void taishang_http_response_free(TaishangHttpResponse *response) {
    if (!response) return;
    
    g_free(response->data);
    g_free(response->error);
    g_free(response);
}

// WebSocket implementation (simplified - would need a proper WebSocket library)
gboolean taishang_network_client_websocket_connect(const char *url, const char *protocol,
                                                   void (*on_open)(void *user_data),
                                                   void (*on_message)(const char *message, void *user_data),
                                                   void (*on_close)(int code, const char *reason, void *user_data),
                                                   void (*on_error)(const char *error, void *user_data),
                                                   void *user_data) {
    if (!network_client || !url) {
        return FALSE;
    }
    
    // Check if connection already exists
    if (g_hash_table_contains(network_client->websocket_connections, url)) {
        g_warning("WebSocket connection already exists for URL: %s", url);
        return FALSE;
    }
    
    // Create connection
    WebSocketConnection *conn = g_new0(WebSocketConnection, 1);
    conn->url = g_strdup(url);
    conn->protocol = g_strdup(protocol);
    conn->connected = FALSE;
    conn->on_open = on_open;
    conn->on_message = on_message;
    conn->on_close = on_close;
    conn->on_error = on_error;
    conn->user_data = user_data;
    
    // Start connection thread
    if (pthread_create(&conn->thread, NULL, websocket_thread_func, conn) != 0) {
        g_error("Failed to create WebSocket thread");
        websocket_connection_free(conn);
        return FALSE;
    }
    
    // Store connection
    g_hash_table_insert(network_client->websocket_connections, g_strdup(url), conn);
    
    g_print("WebSocket connection initiated for: %s\n", url);
    return TRUE;
}

gboolean taishang_network_client_websocket_send(const char *url, const char *message) {
    if (!network_client || !url || !message) {
        return FALSE;
    }
    
    WebSocketConnection *conn = g_hash_table_lookup(network_client->websocket_connections, url);
    if (!conn || !conn->connected) {
        g_warning("WebSocket connection not found or not connected: %s", url);
        return FALSE;
    }
    
    // TODO: Implement actual WebSocket message sending
    g_print("WebSocket send to %s: %s\n", url, message);
    return TRUE;
}

void taishang_network_client_websocket_close(const char *url) {
    if (!network_client || !url) {
        return;
    }
    
    WebSocketConnection *conn = g_hash_table_lookup(network_client->websocket_connections, url);
    if (conn) {
        conn->connected = FALSE;
        pthread_cancel(conn->thread);
        g_hash_table_remove(network_client->websocket_connections, url);
        g_print("WebSocket connection closed: %s\n", url);
    }
}

static void *websocket_thread_func(void *arg) {
    WebSocketConnection *conn = (WebSocketConnection *)arg;
    
    // TODO: Implement actual WebSocket connection logic
    // This is a simplified simulation
    
    // Simulate connection
    g_usleep(1000000); // 1 second delay
    conn->connected = TRUE;
    
    if (conn->on_open) {
        conn->on_open(conn->user_data);
    }
    
    // Simulate periodic messages
    while (conn->connected) {
        g_usleep(5000000); // 5 seconds
        
        if (conn->connected && conn->on_message) {
            conn->on_message("{\"type\":\"ping\",\"timestamp\":\"" G_STRINGIFY(G_USEC_PER_SEC) "\"}", conn->user_data);
        }
    }
    
    if (conn->on_close) {
        conn->on_close(1000, "Connection closed", conn->user_data);
    }
    
    return NULL;
}

static void websocket_connection_free(WebSocketConnection *conn) {
    if (!conn) return;
    
    conn->connected = FALSE;
    pthread_cancel(conn->thread);
    
    g_free(conn->url);
    g_free(conn->protocol);
    g_free(conn);
}

// Configuration functions
void taishang_network_client_set_auth_token(const char *token) {
    if (!network_client) return;
    
    g_free(network_client->auth_token);
    network_client->auth_token = g_strdup(token);
    
    // Update authorization header
    if (token) {
        char *auth_header = g_strdup_printf("Authorization: Bearer %s", token);
        network_client->headers = curl_slist_append(network_client->headers, auth_header);
        g_free(auth_header);
    }
    
    g_print("Auth token updated\n");
}

void taishang_network_client_set_timeout(long timeout_seconds) {
    if (!network_client) return;
    
    network_client->timeout = timeout_seconds;
    curl_easy_setopt(network_client->curl_handle, CURLOPT_TIMEOUT, timeout_seconds);
    
    g_print("Timeout set to %ld seconds\n", timeout_seconds);
}

void taishang_network_client_set_verify_ssl(gboolean verify) {
    if (!network_client) return;
    
    network_client->verify_ssl = verify;
    curl_easy_setopt(network_client->curl_handle, CURLOPT_SSL_VERIFYPEER, verify ? 1L : 0L);
    
    g_print("SSL verification %s\n", verify ? "enabled" : "disabled");
}

void taishang_network_client_add_header(const char *header) {
    if (!network_client || !header) return;
    
    network_client->headers = curl_slist_append(network_client->headers, header);
    g_print("Header added: %s\n", header);
}