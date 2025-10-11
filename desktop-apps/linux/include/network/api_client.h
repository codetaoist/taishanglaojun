#ifndef TAISHANG_API_CLIENT_H
#define TAISHANG_API_CLIENT_H

#include <glib.h>
#include "client.h"

G_BEGIN_DECLS

// API Response structure
typedef struct {
    gboolean success;
    int status_code;
    char *data;
    char *error_message;
} TaishangApiResponse;

// Authentication functions
TaishangApiResponse *taishang_api_client_login(const char *username, const char *password);
TaishangApiResponse *taishang_api_client_logout(void);
TaishangApiResponse *taishang_api_client_register(const char *username, const char *email, const char *password);

// Chat functions
TaishangApiResponse *taishang_api_client_send_message(const char *recipient, const char *message, const char *message_type);
TaishangApiResponse *taishang_api_client_get_chat_history(const char *contact, int limit, int offset);

// Project functions
TaishangApiResponse *taishang_api_client_create_project(const char *name, const char *description, const char *project_type);
TaishangApiResponse *taishang_api_client_get_projects(void);
TaishangApiResponse *taishang_api_client_get_project(const char *project_id);
TaishangApiResponse *taishang_api_client_delete_project(const char *project_id);

// File transfer functions
TaishangApiResponse *taishang_api_client_upload_file(const char *file_path, const char *destination);
TaishangApiResponse *taishang_api_client_download_file(const char *file_id, const char *local_path);
TaishangApiResponse *taishang_api_client_get_files(void);

// Friend management functions
TaishangApiResponse *taishang_api_client_get_friends(void);
TaishangApiResponse *taishang_api_client_add_friend(const char *username);
TaishangApiResponse *taishang_api_client_remove_friend(const char *username);

// WebSocket functions
gboolean taishang_api_client_connect_chat_websocket(TaishangWebSocketOpenCallback on_open,
                                                    TaishangWebSocketMessageCallback on_message,
                                                    TaishangWebSocketCloseCallback on_close,
                                                    TaishangWebSocketErrorCallback on_error,
                                                    void *user_data);

gboolean taishang_api_client_connect_notifications_websocket(TaishangWebSocketOpenCallback on_open,
                                                             TaishangWebSocketMessageCallback on_message,
                                                             TaishangWebSocketCloseCallback on_close,
                                                             TaishangWebSocketErrorCallback on_error,
                                                             void *user_data);

gboolean taishang_api_client_send_chat_websocket_message(const char *message);
void taishang_api_client_disconnect_websockets(void);

// Utility functions
void taishang_api_response_free(TaishangApiResponse *response);

G_END_DECLS

#endif // TAISHANG_API_CLIENT_H