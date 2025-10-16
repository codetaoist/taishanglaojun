#ifndef TEST_NETWORK_H
#define TEST_NETWORK_H

#include <glib.h>

G_BEGIN_DECLS

// Test registration function
void register_network_tests(void);

// Network client tests
void test_network_client_init(void);
void test_network_client_cleanup(void);
void test_network_client_http_get(void);
void test_network_client_http_post(void);
void test_network_client_http_put(void);
void test_network_client_http_delete(void);
void test_network_client_websocket_connect(void);
void test_network_client_websocket_send(void);
void test_network_client_websocket_close(void);
void test_network_client_ssl_verification(void);
void test_network_client_timeout(void);
void test_network_client_headers(void);
void test_network_client_auth_token(void);

// API client tests
void test_api_client_login(void);
void test_api_client_logout(void);
void test_api_client_register(void);
void test_api_client_send_message(void);
void test_api_client_get_chat_history(void);
void test_api_client_create_project(void);
void test_api_client_get_projects(void);
void test_api_client_delete_project(void);
void test_api_client_upload_file(void);
void test_api_client_download_file(void);
void test_api_client_get_files(void);
void test_api_client_get_friends(void);
void test_api_client_add_friend(void);
void test_api_client_remove_friend(void);
void test_api_client_websocket_chat(void);
void test_api_client_websocket_notifications(void);

// Error handling tests
void test_network_error_handling(void);
void test_network_connection_failure(void);
void test_network_timeout_handling(void);
void test_network_ssl_error(void);
void test_network_invalid_response(void);

// Performance tests
void test_network_concurrent_requests(void);
void test_network_large_file_transfer(void);
void test_network_websocket_stress(void);

// Mock server helpers
typedef struct _MockServer MockServer;

MockServer *mock_server_new(guint port);
void mock_server_free(MockServer *server);
void mock_server_start(MockServer *server);
void mock_server_stop(MockServer *server);
void mock_server_add_route(MockServer *server, const char *method, const char *path, const char *response);
void mock_server_set_delay(MockServer *server, guint delay_ms);
void mock_server_set_error_rate(MockServer *server, gdouble error_rate);

G_END_DECLS

#endif // TEST_NETWORK_H