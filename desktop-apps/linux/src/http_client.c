#include "../include/http_client.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>

// 全局HTTP客户端实例
http_client_t* g_http_client = NULL;

// 内部回调数据结构
typedef struct {
    http_callback_t callback;
    void* user_data;
    http_request_t* request;
    http_client_t* client;
} async_request_data_t;

// libcurl写入回调函数
static size_t write_callback(void* contents, size_t size, size_t nmemb, http_memory_t* mem) {
    size_t realsize = size * nmemb;
    char* ptr = realloc(mem->data, mem->size + realsize + 1);
    
    if (!ptr) {
        printf("Not enough memory (realloc returned NULL)\n");
        return 0;
    }
    
    mem->data = ptr;
    memcpy(&(mem->data[mem->size]), contents, realsize);
    mem->size += realsize;
    mem->data[mem->size] = 0;
    
    return realsize;
}

// 头部写入回调函数
static size_t header_callback(void* contents, size_t size, size_t nmemb, http_memory_t* mem) {
    return write_callback(contents, size, nmemb, mem);
}

bool http_client_init(void) {
    if (curl_global_init(CURL_GLOBAL_DEFAULT) != CURLE_OK) {
        return false;
    }
    
    g_http_client = http_client_create();
    return g_http_client != NULL;
}

void http_client_cleanup(void) {
    if (g_http_client) {
        http_client_destroy(g_http_client);
        g_http_client = NULL;
    }
    curl_global_cleanup();
}

http_client_t* http_client_create(void) {
    http_client_t* client = malloc(sizeof(http_client_t));
    if (!client) {
        return NULL;
    }
    
    client->curl = curl_easy_init();
    if (!client->curl) {
        free(client);
        return NULL;
    }
    
    client->base_url = NULL;
    client->default_headers = NULL;
    client->default_header_count = 0;
    
    return client;
}

void http_client_destroy(http_client_t* client) {
    if (!client) return;
    
    if (client->curl) {
        curl_easy_cleanup(client->curl);
    }
    
    if (client->base_url) {
        free(client->base_url);
    }
    
    if (client->default_headers) {
        for (int i = 0; i < client->default_header_count; i++) {
            free(client->default_headers[i]);
        }
        free(client->default_headers);
    }
    
    free(client);
}

http_response_t* http_client_request(http_client_t* client, const http_request_t* request) {
    if (!client || !client->curl || !request) {
        return NULL;
    }
    
    http_response_t* response = malloc(sizeof(http_response_t));
    if (!response) {
        return NULL;
    }
    
    // 初始化响应结构
    memset(response, 0, sizeof(http_response_t));
    response->success = false;
    
    // 准备内存结构
    http_memory_t body_mem = {0};
    http_memory_t header_mem = {0};
    
    // 构建完整URL
    char* full_url;
    if (client->base_url && strstr(request->url, "http") != request->url) {
        full_url = http_build_url(client->base_url, request->url);
    } else {
        full_url = strdup(request->url);
    }
    
    // 设置基本选项
    curl_easy_setopt(client->curl, CURLOPT_URL, full_url);
    curl_easy_setopt(client->curl, CURLOPT_WRITEFUNCTION, write_callback);
    curl_easy_setopt(client->curl, CURLOPT_WRITEDATA, &body_mem);
    curl_easy_setopt(client->curl, CURLOPT_HEADERFUNCTION, header_callback);
    curl_easy_setopt(client->curl, CURLOPT_HEADERDATA, &header_mem);
    
    // 设置超时
    if (request->timeout_ms > 0) {
        curl_easy_setopt(client->curl, CURLOPT_TIMEOUT_MS, request->timeout_ms);
    }
    
    // 设置HTTP方法
    if (strcmp(request->method, "POST") == 0) {
        curl_easy_setopt(client->curl, CURLOPT_POST, 1L);
        if (request->body) {
            curl_easy_setopt(client->curl, CURLOPT_POSTFIELDS, request->body);
        }
    } else if (strcmp(request->method, "PUT") == 0) {
        curl_easy_setopt(client->curl, CURLOPT_CUSTOMREQUEST, "PUT");
        if (request->body) {
            curl_easy_setopt(client->curl, CURLOPT_POSTFIELDS, request->body);
        }
    } else if (strcmp(request->method, "DELETE") == 0) {
        curl_easy_setopt(client->curl, CURLOPT_CUSTOMREQUEST, "DELETE");
    } else {
        curl_easy_setopt(client->curl, CURLOPT_HTTPGET, 1L);
    }
    
    // 设置头部
    struct curl_slist* headers = NULL;
    
    // 添加默认头部
    for (int i = 0; i < client->default_header_count; i++) {
        headers = curl_slist_append(headers, client->default_headers[i]);
    }
    
    // 添加请求头部
    for (int i = 0; i < request->header_count; i++) {
        headers = curl_slist_append(headers, request->headers[i]);
    }
    
    if (headers) {
        curl_easy_setopt(client->curl, CURLOPT_HTTPHEADER, headers);
    }
    
    // 执行请求
    CURLcode res = curl_easy_perform(client->curl);
    
    if (res == CURLE_OK) {
        // 获取状态码
        curl_easy_getinfo(client->curl, CURLINFO_RESPONSE_CODE, &response->status_code);
        
        // 复制响应体
        if (body_mem.data) {
            response->body = malloc(body_mem.size + 1);
            if (response->body) {
                memcpy(response->body, body_mem.data, body_mem.size);
                response->body[body_mem.size] = '\0';
                response->body_size = body_mem.size;
            }
        }
        
        // 复制头部
        if (header_mem.data) {
            response->headers = malloc(header_mem.size + 1);
            if (response->headers) {
                memcpy(response->headers, header_mem.data, header_mem.size);
                response->headers[header_mem.size] = '\0';
                response->headers_size = header_mem.size;
            }
        }
        
        response->success = true;
    } else {
        // 错误处理
        const char* error_str = curl_easy_strerror(res);
        response->error_message = strdup(error_str);
    }
    
    // 清理
    if (headers) {
        curl_slist_free_all(headers);
    }
    if (body_mem.data) {
        free(body_mem.data);
    }
    if (header_mem.data) {
        free(header_mem.data);
    }
    free(full_url);
    
    return response;
}

// 异步请求线程函数
static void* async_request_thread(void* arg) {
    async_request_data_t* data = (async_request_data_t*)arg;
    
    http_response_t* response = http_client_request(data->client, data->request);
    
    if (data->callback) {
        data->callback(response, data->user_data);
    }
    
    http_response_free(response);
    http_request_free(data->request);
    free(data);
    
    return NULL;
}

bool http_client_request_async(http_client_t* client, const http_request_t* request, 
                              http_callback_t callback, void* user_data) {
    if (!client || !request) {
        return false;
    }
    
    async_request_data_t* data = malloc(sizeof(async_request_data_t));
    if (!data) {
        return false;
    }
    
    data->callback = callback;
    data->user_data = user_data;
    data->client = client;
    
    // 复制请求数据
    data->request = http_request_create(request->method, request->url);
    if (request->body) {
        http_request_set_body(data->request, request->body);
    }
    for (int i = 0; i < request->header_count; i++) {
        // 解析头部键值对
        char* header_copy = strdup(request->headers[i]);
        char* colon = strchr(header_copy, ':');
        if (colon) {
            *colon = '\0';
            char* key = header_copy;
            char* value = colon + 1;
            while (*value == ' ') value++; // 跳过空格
            http_request_add_header(data->request, key, value);
        }
        free(header_copy);
    }
    http_request_set_timeout(data->request, request->timeout_ms);
    
    pthread_t thread;
    int result = pthread_create(&thread, NULL, async_request_thread, data);
    if (result != 0) {
        http_request_free(data->request);
        free(data);
        return false;
    }
    
    pthread_detach(thread);
    return true;
}

http_response_t* http_client_get(http_client_t* client, const char* url, 
                                char** headers, int header_count) {
    http_request_t* request = http_request_create("GET", url);
    if (!request) return NULL;
    
    for (int i = 0; i < header_count; i++) {
        request->headers = realloc(request->headers, (request->header_count + 1) * sizeof(char*));
        request->headers[request->header_count] = strdup(headers[i]);
        request->header_count++;
    }
    
    http_response_t* response = http_client_request(client, request);
    http_request_free(request);
    
    return response;
}

http_response_t* http_client_post(http_client_t* client, const char* url, 
                                 const char* body, char** headers, int header_count) {
    http_request_t* request = http_request_create("POST", url);
    if (!request) return NULL;
    
    if (body) {
        http_request_set_body(request, body);
    }
    
    for (int i = 0; i < header_count; i++) {
        request->headers = realloc(request->headers, (request->header_count + 1) * sizeof(char*));
        request->headers[request->header_count] = strdup(headers[i]);
        request->header_count++;
    }
    
    http_response_t* response = http_client_request(client, request);
    http_request_free(request);
    
    return response;
}

http_response_t* http_client_put(http_client_t* client, const char* url, 
                                const char* body, char** headers, int header_count) {
    http_request_t* request = http_request_create("PUT", url);
    if (!request) return NULL;
    
    if (body) {
        http_request_set_body(request, body);
    }
    
    for (int i = 0; i < header_count; i++) {
        request->headers = realloc(request->headers, (request->header_count + 1) * sizeof(char*));
        request->headers[request->header_count] = strdup(headers[i]);
        request->header_count++;
    }
    
    http_response_t* response = http_client_request(client, request);
    http_request_free(request);
    
    return response;
}

http_response_t* http_client_delete(http_client_t* client, const char* url, 
                                   char** headers, int header_count) {
    http_request_t* request = http_request_create("DELETE", url);
    if (!request) return NULL;
    
    for (int i = 0; i < header_count; i++) {
        request->headers = realloc(request->headers, (request->header_count + 1) * sizeof(char*));
        request->headers[request->header_count] = strdup(headers[i]);
        request->header_count++;
    }
    
    http_response_t* response = http_client_request(client, request);
    http_request_free(request);
    
    return response;
}

void http_client_set_base_url(http_client_t* client, const char* base_url) {
    if (!client) return;
    
    if (client->base_url) {
        free(client->base_url);
    }
    
    client->base_url = base_url ? strdup(base_url) : NULL;
}

void http_client_add_default_header(http_client_t* client, const char* key, const char* value) {
    if (!client || !key || !value) return;
    
    char* header = malloc(strlen(key) + strlen(value) + 3);
    sprintf(header, "%s: %s", key, value);
    
    client->default_headers = realloc(client->default_headers, 
                                     (client->default_header_count + 1) * sizeof(char*));
    client->default_headers[client->default_header_count] = header;
    client->default_header_count++;
}

void http_client_remove_default_header(http_client_t* client, const char* key) {
    if (!client || !key) return;
    
    for (int i = 0; i < client->default_header_count; i++) {
        if (strncmp(client->default_headers[i], key, strlen(key)) == 0) {
            free(client->default_headers[i]);
            
            // 移动后续元素
            for (int j = i; j < client->default_header_count - 1; j++) {
                client->default_headers[j] = client->default_headers[j + 1];
            }
            
            client->default_header_count--;
            client->default_headers = realloc(client->default_headers, 
                                             client->default_header_count * sizeof(char*));
            break;
        }
    }
}

void http_response_free(http_response_t* response) {
    if (!response) return;
    
    if (response->body) {
        free(response->body);
    }
    if (response->headers) {
        free(response->headers);
    }
    if (response->error_message) {
        free(response->error_message);
    }
    
    free(response);
}

http_request_t* http_request_create(const char* method, const char* url) {
    if (!method || !url) return NULL;
    
    http_request_t* request = malloc(sizeof(http_request_t));
    if (!request) return NULL;
    
    memset(request, 0, sizeof(http_request_t));
    
    request->method = strdup(method);
    request->url = strdup(url);
    request->timeout_ms = 30000; // 默认30秒超时
    
    return request;
}

void http_request_set_body(http_request_t* request, const char* body) {
    if (!request) return;
    
    if (request->body) {
        free(request->body);
    }
    
    request->body = body ? strdup(body) : NULL;
}

void http_request_add_header(http_request_t* request, const char* key, const char* value) {
    if (!request || !key || !value) return;
    
    char* header = malloc(strlen(key) + strlen(value) + 3);
    sprintf(header, "%s: %s", key, value);
    
    request->headers = realloc(request->headers, (request->header_count + 1) * sizeof(char*));
    request->headers[request->header_count] = header;
    request->header_count++;
}

void http_request_set_timeout(http_request_t* request, long timeout_ms) {
    if (!request) return;
    request->timeout_ms = timeout_ms;
}

void http_request_free(http_request_t* request) {
    if (!request) return;
    
    if (request->method) {
        free(request->method);
    }
    if (request->url) {
        free(request->url);
    }
    if (request->body) {
        free(request->body);
    }
    if (request->headers) {
        for (int i = 0; i < request->header_count; i++) {
            free(request->headers[i]);
        }
        free(request->headers);
    }
    
    free(request);
}

char* http_build_url(const char* base_url, const char* path) {
    if (!base_url || !path) return NULL;
    
    size_t base_len = strlen(base_url);
    size_t path_len = strlen(path);
    
    // 检查base_url是否以'/'结尾，path是否以'/'开头
    bool base_ends_slash = (base_len > 0 && base_url[base_len - 1] == '/');
    bool path_starts_slash = (path_len > 0 && path[0] == '/');
    
    size_t total_len = base_len + path_len + 1;
    if (!base_ends_slash && !path_starts_slash) {
        total_len++; // 需要添加'/'
    } else if (base_ends_slash && path_starts_slash) {
        total_len--; // 需要去掉一个'/'
    }
    
    char* result = malloc(total_len);
    if (!result) return NULL;
    
    strcpy(result, base_url);
    
    if (!base_ends_slash && !path_starts_slash) {
        strcat(result, "/");
        strcat(result, path);
    } else if (base_ends_slash && path_starts_slash) {
        strcat(result, path + 1);
    } else {
        strcat(result, path);
    }
    
    return result;
}

char* http_escape_string(const char* string) {
    if (!string) return NULL;
    
    CURL* curl = curl_easy_init();
    if (!curl) return NULL;
    
    char* escaped = curl_easy_escape(curl, string, 0);
    char* result = escaped ? strdup(escaped) : NULL;
    
    if (escaped) {
        curl_free(escaped);
    }
    curl_easy_cleanup(curl);
    
    return result;
}

void http_unescape_string(char* string) {
    if (!string) return;
    
    CURL* curl = curl_easy_init();
    if (!curl) return;
    
    int outlength;
    char* unescaped = curl_easy_unescape(curl, string, 0, &outlength);
    
    if (unescaped) {
        strcpy(string, unescaped);
        curl_free(unescaped);
    }
    
    curl_easy_cleanup(curl);
}