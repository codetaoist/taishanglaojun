#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <json-c/json.h>
#include <glib.h>
#include "../../include/network/api_client.h"
#include "../../include/network/client.h"

// API endpoints
#define API_AUTH_LOGIN "/api/auth/login"
#define API_AUTH_LOGOUT "/api/auth/logout"
#define API_AUTH_REFRESH "/api/auth/refresh"
#define API_AUTH_REGISTER "/api/auth/register"

#define API_CHAT_MESSAGES "/api/chat/messages"
#define API_CHAT_SEND "/api/chat/send"
#define API_CHAT_HISTORY "/api/chat/history"

#define API_PROJECTS_LIST "/api/projects"
#define API_PROJECTS_CREATE "/api/projects"
#define API_PROJECTS_GET "/api/projects/%s"
#define API_PROJECTS_UPDATE "/api/projects/%s"
#define API_PROJECTS_DELETE "/api/projects/%s"

#define API_FILES_UPLOAD "/api/files/upload"
#define API_FILES_DOWNLOAD "/api/files/download/%s"
#define API_FILES_LIST "/api/files"
#define API_FILES_DELETE "/api/files/%s"

#define API_FRIENDS_LIST "/api/friends"
#define API_FRIENDS_ADD "/api/friends/add"
#define API_FRIENDS_REMOVE "/api/friends/remove/%s"
#define API_FRIENDS_REQUESTS "/api/friends/requests"

// WebSocket endpoints
#define WS_CHAT "/ws/chat"
#define WS_NOTIFICATIONS "/ws/notifications"
#define WS_PRESENCE "/ws/presence"

// Authentication functions
TaishangApiResponse *taishang_api_client_login(const char *username, const char *password) {
    if (!username || !password) {
        return NULL;
    }
    
    // Create login request
    json_object *request = json_object_new_object();
    json_object *username_obj = json_object_new_string(username);
    json_object *password_obj = json_object_new_string(password);
    
    json_object_object_add(request, "username", username_obj);
    json_object_object_add(request, "password", password_obj);
    
    const char *request_data = json_object_to_json_string(request);
    
    // Send HTTP request
    TaishangHttpResponse *http_response = taishang_network_client_post(API_AUTH_LOGIN, request_data, TAISHANG_HTTP_CONTENT_TYPE_JSON);
    
    // Create API response
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    
    if (http_response->success && http_response->data) {
        // Parse response JSON
        json_object *response_json = json_tokener_parse(http_response->data);
        if (response_json) {
            json_object *token_obj;
            if (json_object_object_get_ex(response_json, "token", &token_obj)) {
                const char *token = json_object_get_string(token_obj);
                taishang_network_client_set_auth_token(token);
            }
            
            api_response->data = g_strdup(http_response->data);
            json_object_put(response_json);
        }
    }
    
    // Cleanup
    json_object_put(request);
    taishang_http_response_free(http_response);
    
    g_print("Login attempt for user: %s -> %s\n", username, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_logout(void) {
    TaishangHttpResponse *http_response = taishang_network_client_post(API_AUTH_LOGOUT, NULL, NULL);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    
    if (http_response->success) {
        // Clear auth token
        taishang_network_client_set_auth_token(NULL);
    }
    
    taishang_http_response_free(http_response);
    
    g_print("Logout -> %s\n", api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_register(const char *username, const char *email, const char *password) {
    if (!username || !email || !password) {
        return NULL;
    }
    
    // Create registration request
    json_object *request = json_object_new_object();
    json_object *username_obj = json_object_new_string(username);
    json_object *email_obj = json_object_new_string(email);
    json_object *password_obj = json_object_new_string(password);
    
    json_object_object_add(request, "username", username_obj);
    json_object_object_add(request, "email", email_obj);
    json_object_object_add(request, "password", password_obj);
    
    const char *request_data = json_object_to_json_string(request);
    
    // Send HTTP request
    TaishangHttpResponse *http_response = taishang_network_client_post(API_AUTH_REGISTER, request_data, TAISHANG_HTTP_CONTENT_TYPE_JSON);
    
    // Create API response
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    // Cleanup
    json_object_put(request);
    taishang_http_response_free(http_response);
    
    g_print("Registration attempt for user: %s -> %s\n", username, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

// Chat functions
TaishangApiResponse *taishang_api_client_send_message(const char *recipient, const char *message, const char *message_type) {
    if (!recipient || !message) {
        return NULL;
    }
    
    // Create message request
    json_object *request = json_object_new_object();
    json_object *recipient_obj = json_object_new_string(recipient);
    json_object *message_obj = json_object_new_string(message);
    json_object *type_obj = json_object_new_string(message_type ? message_type : "text");
    json_object *timestamp_obj = json_object_new_int64(g_get_real_time());
    
    json_object_object_add(request, "recipient", recipient_obj);
    json_object_object_add(request, "message", message_obj);
    json_object_object_add(request, "type", type_obj);
    json_object_object_add(request, "timestamp", timestamp_obj);
    
    const char *request_data = json_object_to_json_string(request);
    
    // Send HTTP request
    TaishangHttpResponse *http_response = taishang_network_client_post(API_CHAT_SEND, request_data, TAISHANG_HTTP_CONTENT_TYPE_JSON);
    
    // Create API response
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    // Cleanup
    json_object_put(request);
    taishang_http_response_free(http_response);
    
    g_print("Message sent to %s -> %s\n", recipient, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_get_chat_history(const char *contact, int limit, int offset) {
    // Create query parameters
    GHashTable *params = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, g_free);
    
    if (contact) {
        g_hash_table_insert(params, g_strdup("contact"), g_strdup(contact));
    }
    
    if (limit > 0) {
        g_hash_table_insert(params, g_strdup("limit"), g_strdup_printf("%d", limit));
    }
    
    if (offset > 0) {
        g_hash_table_insert(params, g_strdup("offset"), g_strdup_printf("%d", offset));
    }
    
    // Send HTTP request
    TaishangHttpResponse *http_response = taishang_network_client_get(API_CHAT_HISTORY, params);
    
    // Create API response
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    // Cleanup
    g_hash_table_destroy(params);
    taishang_http_response_free(http_response);
    
    g_print("Chat history retrieved for %s -> %s\n", contact ? contact : "all", api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

// Project functions
TaishangApiResponse *taishang_api_client_create_project(const char *name, const char *description, const char *project_type) {
    if (!name) {
        return NULL;
    }
    
    // Create project request
    json_object *request = json_object_new_object();
    json_object *name_obj = json_object_new_string(name);
    json_object *desc_obj = json_object_new_string(description ? description : "");
    json_object *type_obj = json_object_new_string(project_type ? project_type : "general");
    json_object *created_obj = json_object_new_int64(g_get_real_time());
    
    json_object_object_add(request, "name", name_obj);
    json_object_object_add(request, "description", desc_obj);
    json_object_object_add(request, "type", type_obj);
    json_object_object_add(request, "created_at", created_obj);
    
    const char *request_data = json_object_to_json_string(request);
    
    // Send HTTP request
    TaishangHttpResponse *http_response = taishang_network_client_post(API_PROJECTS_CREATE, request_data, TAISHANG_HTTP_CONTENT_TYPE_JSON);
    
    // Create API response
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    // Cleanup
    json_object_put(request);
    taishang_http_response_free(http_response);
    
    g_print("Project created: %s -> %s\n", name, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_get_projects(void) {
    TaishangHttpResponse *http_response = taishang_network_client_get(API_PROJECTS_LIST, NULL);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    taishang_http_response_free(http_response);
    
    g_print("Projects retrieved -> %s\n", api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_get_project(const char *project_id) {
    if (!project_id) {
        return NULL;
    }
    
    char *endpoint = g_strdup_printf(API_PROJECTS_GET, project_id);
    TaishangHttpResponse *http_response = taishang_network_client_get(endpoint, NULL);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    g_free(endpoint);
    taishang_http_response_free(http_response);
    
    g_print("Project %s retrieved -> %s\n", project_id, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_delete_project(const char *project_id) {
    if (!project_id) {
        return NULL;
    }
    
    char *endpoint = g_strdup_printf(API_PROJECTS_DELETE, project_id);
    TaishangHttpResponse *http_response = taishang_network_client_delete(endpoint);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    g_free(endpoint);
    taishang_http_response_free(http_response);
    
    g_print("Project %s deleted -> %s\n", project_id, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

// File transfer functions
TaishangApiResponse *taishang_api_client_upload_file(const char *file_path, const char *destination) {
    if (!file_path) {
        return NULL;
    }
    
    // TODO: Implement multipart file upload
    // For now, create a simple JSON request with file info
    json_object *request = json_object_new_object();
    json_object *path_obj = json_object_new_string(file_path);
    json_object *dest_obj = json_object_new_string(destination ? destination : "");
    json_object *timestamp_obj = json_object_new_int64(g_get_real_time());
    
    json_object_object_add(request, "file_path", path_obj);
    json_object_object_add(request, "destination", dest_obj);
    json_object_object_add(request, "timestamp", timestamp_obj);
    
    const char *request_data = json_object_to_json_string(request);
    
    TaishangHttpResponse *http_response = taishang_network_client_post(API_FILES_UPLOAD, request_data, TAISHANG_HTTP_CONTENT_TYPE_JSON);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    json_object_put(request);
    taishang_http_response_free(http_response);
    
    g_print("File upload: %s -> %s\n", file_path, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_download_file(const char *file_id, const char *local_path) {
    if (!file_id) {
        return NULL;
    }
    
    char *endpoint = g_strdup_printf(API_FILES_DOWNLOAD, file_id);
    TaishangHttpResponse *http_response = taishang_network_client_get(endpoint, NULL);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    
    // If successful and local_path is provided, save the file
    if (http_response->success && local_path && http_response->data) {
        GError *error = NULL;
        if (g_file_set_contents(local_path, http_response->data, http_response->size, &error)) {
            api_response->data = g_strdup_printf("File saved to: %s", local_path);
        } else {
            api_response->success = FALSE;
            api_response->error_message = g_strdup(error->message);
            g_error_free(error);
        }
    } else {
        api_response->data = g_strdup(http_response->data);
    }
    
    g_free(endpoint);
    taishang_http_response_free(http_response);
    
    g_print("File download: %s -> %s\n", file_id, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_get_files(void) {
    TaishangHttpResponse *http_response = taishang_network_client_get(API_FILES_LIST, NULL);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    taishang_http_response_free(http_response);
    
    g_print("Files list retrieved -> %s\n", api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

// Friend management functions
TaishangApiResponse *taishang_api_client_get_friends(void) {
    TaishangHttpResponse *http_response = taishang_network_client_get(API_FRIENDS_LIST, NULL);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    taishang_http_response_free(http_response);
    
    g_print("Friends list retrieved -> %s\n", api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_add_friend(const char *username) {
    if (!username) {
        return NULL;
    }
    
    json_object *request = json_object_new_object();
    json_object *username_obj = json_object_new_string(username);
    
    json_object_object_add(request, "username", username_obj);
    
    const char *request_data = json_object_to_json_string(request);
    
    TaishangHttpResponse *http_response = taishang_network_client_post(API_FRIENDS_ADD, request_data, TAISHANG_HTTP_CONTENT_TYPE_JSON);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    json_object_put(request);
    taishang_http_response_free(http_response);
    
    g_print("Friend add request: %s -> %s\n", username, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

TaishangApiResponse *taishang_api_client_remove_friend(const char *username) {
    if (!username) {
        return NULL;
    }
    
    char *endpoint = g_strdup_printf(API_FRIENDS_REMOVE, username);
    TaishangHttpResponse *http_response = taishang_network_client_delete(endpoint);
    
    TaishangApiResponse *api_response = g_new0(TaishangApiResponse, 1);
    api_response->success = http_response->success;
    api_response->status_code = http_response->status_code;
    api_response->error_message = g_strdup(http_response->error);
    api_response->data = g_strdup(http_response->data);
    
    g_free(endpoint);
    taishang_http_response_free(http_response);
    
    g_print("Friend removed: %s -> %s\n", username, api_response->success ? "SUCCESS" : "FAILED");
    return api_response;
}

// WebSocket functions
gboolean taishang_api_client_connect_chat_websocket(TaishangWebSocketOpenCallback on_open,
                                                    TaishangWebSocketMessageCallback on_message,
                                                    TaishangWebSocketCloseCallback on_close,
                                                    TaishangWebSocketErrorCallback on_error,
                                                    void *user_data) {
    return taishang_network_client_websocket_connect(WS_CHAT, "chat", on_open, on_message, on_close, on_error, user_data);
}

gboolean taishang_api_client_connect_notifications_websocket(TaishangWebSocketOpenCallback on_open,
                                                             TaishangWebSocketMessageCallback on_message,
                                                             TaishangWebSocketCloseCallback on_close,
                                                             TaishangWebSocketErrorCallback on_error,
                                                             void *user_data) {
    return taishang_network_client_websocket_connect(WS_NOTIFICATIONS, "notifications", on_open, on_message, on_close, on_error, user_data);
}

gboolean taishang_api_client_send_chat_websocket_message(const char *message) {
    return taishang_network_client_websocket_send(WS_CHAT, message);
}

void taishang_api_client_disconnect_websockets(void) {
    taishang_network_client_websocket_close(WS_CHAT);
    taishang_network_client_websocket_close(WS_NOTIFICATIONS);
    taishang_network_client_websocket_close(WS_PRESENCE);
}

// Utility functions
void taishang_api_response_free(TaishangApiResponse *response) {
    if (!response) return;
    
    g_free(response->data);
    g_free(response->error_message);
    g_free(response);
}