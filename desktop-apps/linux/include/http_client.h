#ifndef HTTP_CLIENT_H
#define HTTP_CLIENT_H

#include <curl/curl.h>
#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// HTTP响应结构
typedef struct {
    long status_code;
    char* body;
    size_t body_size;
    char* headers;
    size_t headers_size;
    bool success;
    char* error_message;
} http_response_t;

// HTTP请求结构
typedef struct {
    char* method;
    char* url;
    char* body;
    char** headers;
    int header_count;
    long timeout_ms;
} http_request_t;

// HTTP客户端结构
typedef struct {
    CURL* curl;
    char* base_url;
    char** default_headers;
    int default_header_count;
} http_client_t;

// 回调函数类型
typedef void (*http_callback_t)(const http_response_t* response, void* user_data);

// 内部数据结构
typedef struct {
    char* data;
    size_t size;
} http_memory_t;

// 全局HTTP客户端实例
extern http_client_t* g_http_client;

// 初始化和清理函数
bool http_client_init(void);
void http_client_cleanup(void);

// HTTP客户端创建和销毁
http_client_t* http_client_create(void);
void http_client_destroy(http_client_t* client);

// 同步HTTP请求
http_response_t* http_client_request(http_client_t* client, const http_request_t* request);

// 异步HTTP请求
bool http_client_request_async(http_client_t* client, const http_request_t* request, 
                              http_callback_t callback, void* user_data);

// 便捷方法
http_response_t* http_client_get(http_client_t* client, const char* url, 
                                char** headers, int header_count);
http_response_t* http_client_post(http_client_t* client, const char* url, 
                                 const char* body, char** headers, int header_count);
http_response_t* http_client_put(http_client_t* client, const char* url, 
                                const char* body, char** headers, int header_count);
http_response_t* http_client_delete(http_client_t* client, const char* url, 
                                   char** headers, int header_count);

// 配置方法
void http_client_set_base_url(http_client_t* client, const char* base_url);
void http_client_add_default_header(http_client_t* client, const char* key, const char* value);
void http_client_remove_default_header(http_client_t* client, const char* key);

// 响应处理
void http_response_free(http_response_t* response);

// 请求处理
http_request_t* http_request_create(const char* method, const char* url);
void http_request_set_body(http_request_t* request, const char* body);
void http_request_add_header(http_request_t* request, const char* key, const char* value);
void http_request_set_timeout(http_request_t* request, long timeout_ms);
void http_request_free(http_request_t* request);

// 辅助函数
char* http_build_url(const char* base_url, const char* path);
char* http_escape_string(const char* string);
void http_unescape_string(char* string);

#ifdef __cplusplus
}
#endif

#endif // HTTP_CLIENT_H