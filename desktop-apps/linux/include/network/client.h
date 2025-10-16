#ifndef TAISHANG_NETWORK_CLIENT_H
#define TAISHANG_NETWORK_CLIENT_H

#include <glib.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangNetworkClient TaishangNetworkClient;

// HTTP Response structure
typedef struct {
    int status_code;
    char *data;
    size_t size;
    char *error;
    gboolean success;
} TaishangHttpResponse;

// WebSocket callback types
typedef void (*TaishangWebSocketOpenCallback)(void *user_data);
typedef void (*TaishangWebSocketMessageCallback)(const char *message, void *user_data);
typedef void (*TaishangWebSocketCloseCallback)(int code, const char *reason, void *user_data);
typedef void (*TaishangWebSocketErrorCallback)(const char *error, void *user_data);

// Initialization and cleanup
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
                                                   TaishangWebSocketOpenCallback on_open,
                                                   TaishangWebSocketMessageCallback on_message,
                                                   TaishangWebSocketCloseCallback on_close,
                                                   TaishangWebSocketErrorCallback on_error,
                                                   void *user_data);
gboolean taishang_network_client_websocket_send(const char *url, const char *message);
void taishang_network_client_websocket_close(const char *url);

// Configuration functions
void taishang_network_client_set_auth_token(const char *token);
void taishang_network_client_set_timeout(long timeout_seconds);
void taishang_network_client_set_verify_ssl(gboolean verify);
void taishang_network_client_add_header(const char *header);

// Utility macros
#define TAISHANG_HTTP_CONTENT_TYPE_JSON "application/json"
#define TAISHANG_HTTP_CONTENT_TYPE_FORM "application/x-www-form-urlencoded"
#define TAISHANG_HTTP_CONTENT_TYPE_TEXT "text/plain"

G_END_DECLS

#endif // TAISHANG_NETWORK_CLIENT_H