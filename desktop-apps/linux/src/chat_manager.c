#include "chat_manager.h"
#include "auth_manager.h"
#include <unistd.h>
#include <sys/time.h>
#include <errno.h>

// 全局实例
chat_manager_t* g_chat_manager = NULL;

// 内部辅助函数声明
static char* build_chat_url(const chat_manager_t* manager, const char* endpoint);
static char* build_websocket_url(const chat_manager_t* manager);
static json_object* message_to_json(const message_t* message);
static json_object* chat_to_json(const chat_t* chat);
static json_object* send_message_request_to_json(const send_message_request_t* request);
static json_object* create_chat_request_to_json(const create_chat_request_t* request);
static void* websocket_thread_func(void* arg);
static void* reconnect_thread_func(void* arg);
static void handle_websocket_message(chat_manager_t* manager, const char* message);
static void add_auth_headers(http_request_t* request);
static void update_local_chats(chat_manager_t* manager, const chat_t* chats, size_t count);
static void update_local_messages(chat_manager_t* manager, const char* chat_id, const message_t* messages, size_t count);
static void add_local_message(chat_manager_t* manager, const message_t* message);
static size_t websocket_write_callback(void* contents, size_t size, size_t nmemb, void* userp);

// 聊天管理器创建和销毁
chat_manager_t* chat_manager_create(void) {
    chat_manager_t* manager = (chat_manager_t*)calloc(1, sizeof(chat_manager_t));
    if (!manager) {
        return NULL;
    }
    
    // 初始化互斥锁
    if (pthread_mutex_init(&manager->data_mutex, NULL) != 0) {
        free(manager);
        return NULL;
    }
    
    // 设置默认配置
    manager->server_url = strdup("http://localhost:8080");
    manager->websocket_url = strdup("ws://localhost:8080/ws");
    manager->auto_reconnect_enabled = true;
    manager->reconnect_interval = 5;
    
    // 初始化数据存储
    manager->chats = NULL;
    manager->chats_count = 0;
    manager->chat_messages = NULL;
    manager->chat_message_counts = NULL;
    manager->chat_messages_capacity = 0;
    
    // 初始化WebSocket状态
    manager->websocket_handle = NULL;
    manager->websocket_connected = false;
    manager->should_stop_websocket = false;
    manager->should_stop_reconnect = false;
    
    manager->initialized = false;
    
    return manager;
}

void chat_manager_destroy(chat_manager_t* manager) {
    if (!manager) return;
    
    chat_manager_cleanup(manager);
    
    // 清理数据
    if (manager->chats) {
        for (size_t i = 0; i < manager->chats_count; i++) {
            chat_free(&manager->chats[i]);
        }
        free(manager->chats);
    }
    
    if (manager->chat_messages) {
        for (size_t i = 0; i < manager->chat_messages_capacity; i++) {
            if (manager->chat_messages[i]) {
                for (size_t j = 0; j < manager->chat_message_counts[i]; j++) {
                    message_free(&manager->chat_messages[i][j]);
                }
                free(manager->chat_messages[i]);
            }
        }
        free(manager->chat_messages);
        free(manager->chat_message_counts);
    }
    
    // 清理配置
    free(manager->server_url);
    free(manager->websocket_url);
    
    // 销毁互斥锁
    pthread_mutex_destroy(&manager->data_mutex);
    
    free(manager);
}

// 初始化和清理
bool chat_manager_initialize(chat_manager_t* manager) {
    if (!manager) return false;
    
    if (manager->initialized) {
        return true;
    }
    
    // 初始化HTTP客户端
    manager->http_client = http_client_create();
    if (!manager->http_client) {
        return false;
    }
    
    // 设置基础URL
    http_client_set_base_url(manager->http_client, manager->server_url);
    
    manager->initialized = true;
    return true;
}

void chat_manager_cleanup(chat_manager_t* manager) {
    if (!manager || !manager->initialized) return;
    
    // 断开WebSocket连接
    chat_manager_disconnect_websocket(manager);
    
    // 停止重连线程
    manager->should_stop_reconnect = true;
    if (manager->reconnect_thread) {
        pthread_join(manager->reconnect_thread, NULL);
        manager->reconnect_thread = 0;
    }
    
    // 清理HTTP客户端
    if (manager->http_client) {
        http_client_destroy(manager->http_client);
        manager->http_client = NULL;
    }
    
    manager->initialized = false;
}

// 聊天列表管理
bool chat_manager_get_chat_list(chat_manager_t* manager) {
    if (!manager || !manager->initialized) return false;
    
    char* url = build_chat_url(manager, "/chats");
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("GET");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    if (!response || response->status_code != 200) {
        if (response) http_response_free(response);
        return false;
    }
    
    // 解析响应
    chat_response_t* chat_response = chat_response_create_from_json(response->body);
    http_response_free(response);
    
    if (!chat_response || !chat_response->success) {
        if (chat_response) chat_response_free(chat_response);
        return false;
    }
    
    // 更新本地数据
    update_local_chats(manager, chat_response->chats, chat_response->chats_count);
    
    // 触发回调
    if (manager->on_chats_updated) {
        manager->on_chats_updated(manager->chats, manager->chats_count);
    }
    
    chat_response_free(chat_response);
    return true;
}

bool chat_manager_get_chat_list_async(chat_manager_t* manager) {
    // TODO: 实现异步版本
    return chat_manager_get_chat_list(manager);
}

// 消息管理
bool chat_manager_get_messages(chat_manager_t* manager, const char* chat_id, int page, int limit) {
    if (!manager || !manager->initialized || !chat_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/chats/%s/messages?page=%d&limit=%d", chat_id, page, limit);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("GET");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    if (!response || response->status_code != 200) {
        if (response) http_response_free(response);
        return false;
    }
    
    // 解析响应
    chat_response_t* chat_response = chat_response_create_from_json(response->body);
    http_response_free(response);
    
    if (!chat_response || !chat_response->success) {
        if (chat_response) chat_response_free(chat_response);
        return false;
    }
    
    // 更新本地消息
    update_local_messages(manager, chat_id, chat_response->messages, chat_response->messages_count);
    
    // 触发回调
    if (manager->on_messages_updated) {
        manager->on_messages_updated(chat_response->messages, chat_response->messages_count);
    }
    
    chat_response_free(chat_response);
    return true;
}

bool chat_manager_get_messages_async(chat_manager_t* manager, const char* chat_id, int page, int limit) {
    // TODO: 实现异步版本
    return chat_manager_get_messages(manager, chat_id, page, limit);
}

bool chat_manager_send_message(chat_manager_t* manager, const send_message_request_t* request) {
    if (!manager || !manager->initialized || !request) return false;
    
    char* url = build_chat_url(manager, "/messages");
    if (!url) return false;
    
    json_object* json_data = send_message_request_to_json(request);
    const char* json_string = json_object_to_json_string(json_data);
    
    http_request_t http_request = {0};
    http_request.url = url;
    http_request.method = strdup("POST");
    http_request.body = strdup(json_string);
    add_auth_headers(&http_request);
    
    // 添加Content-Type头
    int header_count = 0;
    if (http_request.headers) {
        while (http_request.headers[header_count]) header_count++;
    }
    
    http_request.headers = realloc(http_request.headers, (header_count + 2) * sizeof(char*));
    http_request.headers[header_count] = strdup("Content-Type: application/json");
    http_request.headers[header_count + 1] = NULL;
    
    http_response_t* response = http_client_request(manager->http_client, &http_request);
    
    free(url);
    free(http_request.method);
    free(http_request.body);
    if (http_request.headers) {
        for (int i = 0; http_request.headers[i]; i++) {
            free(http_request.headers[i]);
        }
        free(http_request.headers);
    }
    json_object_put(json_data);
    
    if (!response || response->status_code != 200) {
        if (response) http_response_free(response);
        return false;
    }
    
    // 解析响应
    chat_response_t* chat_response = chat_response_create_from_json(response->body);
    http_response_free(response);
    
    if (!chat_response || !chat_response->success) {
        if (chat_response) chat_response_free(chat_response);
        return false;
    }
    
    // 添加到本地消息
    add_local_message(manager, &chat_response->message_data);
    
    // 触发回调
    if (manager->on_new_message) {
        manager->on_new_message(&chat_response->message_data);
    }
    
    chat_response_free(chat_response);
    return true;
}

bool chat_manager_send_message_async(chat_manager_t* manager, const send_message_request_t* request) {
    // TODO: 实现异步版本
    return chat_manager_send_message(manager, request);
}

bool chat_manager_mark_message_as_read(chat_manager_t* manager, const char* message_id) {
    if (!manager || !manager->initialized || !message_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/messages/%s/read", message_id);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("PUT");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    bool success = response && response->status_code == 200;
    if (response) http_response_free(response);
    
    return success;
}

bool chat_manager_mark_chat_as_read(chat_manager_t* manager, const char* chat_id) {
    if (!manager || !manager->initialized || !chat_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/chats/%s/read", chat_id);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("PUT");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    bool success = response && response->status_code == 200;
    if (response) http_response_free(response);
    
    return success;
}

// 聊天会话管理
bool chat_manager_create_chat(chat_manager_t* manager, const create_chat_request_t* request) {
    if (!manager || !manager->initialized || !request) return false;
    
    char* url = build_chat_url(manager, "/chats");
    if (!url) return false;
    
    json_object* json_data = create_chat_request_to_json(request);
    const char* json_string = json_object_to_json_string(json_data);
    
    http_request_t http_request = {0};
    http_request.url = url;
    http_request.method = strdup("POST");
    http_request.body = strdup(json_string);
    add_auth_headers(&http_request);
    
    // 添加Content-Type头
    int header_count = 0;
    if (http_request.headers) {
        while (http_request.headers[header_count]) header_count++;
    }
    
    http_request.headers = realloc(http_request.headers, (header_count + 2) * sizeof(char*));
    http_request.headers[header_count] = strdup("Content-Type: application/json");
    http_request.headers[header_count + 1] = NULL;
    
    http_response_t* response = http_client_request(manager->http_client, &http_request);
    
    free(url);
    free(http_request.method);
    free(http_request.body);
    if (http_request.headers) {
        for (int i = 0; http_request.headers[i]; i++) {
            free(http_request.headers[i]);
        }
        free(http_request.headers);
    }
    json_object_put(json_data);
    
    if (!response || response->status_code != 201) {
        if (response) http_response_free(response);
        return false;
    }
    
    // 解析响应并更新本地数据
    chat_response_t* chat_response = chat_response_create_from_json(response->body);
    http_response_free(response);
    
    if (!chat_response || !chat_response->success) {
        if (chat_response) chat_response_free(chat_response);
        return false;
    }
    
    // 添加新聊天到本地列表
    pthread_mutex_lock(&manager->data_mutex);
    manager->chats = realloc(manager->chats, (manager->chats_count + 1) * sizeof(chat_t));
    manager->chats[manager->chats_count] = chat_response->chat;
    manager->chats_count++;
    pthread_mutex_unlock(&manager->data_mutex);
    
    // 触发回调
    if (manager->on_chats_updated) {
        manager->on_chats_updated(manager->chats, manager->chats_count);
    }
    
    chat_response_free(chat_response);
    return true;
}

bool chat_manager_create_chat_async(chat_manager_t* manager, const create_chat_request_t* request) {
    // TODO: 实现异步版本
    return chat_manager_create_chat(manager, request);
}

bool chat_manager_delete_chat(chat_manager_t* manager, const char* chat_id) {
    if (!manager || !manager->initialized || !chat_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/chats/%s", chat_id);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("DELETE");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    bool success = response && response->status_code == 200;
    if (response) http_response_free(response);
    
    if (success) {
        // 从本地列表中移除聊天
        pthread_mutex_lock(&manager->data_mutex);
        for (size_t i = 0; i < manager->chats_count; i++) {
            if (strcmp(manager->chats[i].id, chat_id) == 0) {
                chat_free(&manager->chats[i]);
                memmove(&manager->chats[i], &manager->chats[i + 1], 
                       (manager->chats_count - i - 1) * sizeof(chat_t));
                manager->chats_count--;
                break;
            }
        }
        pthread_mutex_unlock(&manager->data_mutex);
        
        // 触发回调
        if (manager->on_chats_updated) {
            manager->on_chats_updated(manager->chats, manager->chats_count);
        }
    }
    
    return success;
}

bool chat_manager_leave_chat(chat_manager_t* manager, const char* chat_id) {
    if (!manager || !manager->initialized || !chat_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/chats/%s/leave", chat_id);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("POST");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    bool success = response && response->status_code == 200;
    if (response) http_response_free(response);
    
    return success;
}

bool chat_manager_add_participant(chat_manager_t* manager, const char* chat_id, const char* user_id) {
    if (!manager || !manager->initialized || !chat_id || !user_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/chats/%s/participants", chat_id);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    json_object* json_data = json_object_new_object();
    json_object* user_id_obj = json_object_new_string(user_id);
    json_object_object_add(json_data, "user_id", user_id_obj);
    
    const char* json_string = json_object_to_json_string(json_data);
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("POST");
    request.body = strdup(json_string);
    add_auth_headers(&request);
    
    // 添加Content-Type头
    int header_count = 0;
    if (request.headers) {
        while (request.headers[header_count]) header_count++;
    }
    
    request.headers = realloc(request.headers, (header_count + 2) * sizeof(char*));
    request.headers[header_count] = strdup("Content-Type: application/json");
    request.headers[header_count + 1] = NULL;
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    free(request.body);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    json_object_put(json_data);
    
    bool success = response && response->status_code == 200;
    if (response) http_response_free(response);
    
    return success;
}

bool chat_manager_remove_participant(chat_manager_t* manager, const char* chat_id, const char* user_id) {
    if (!manager || !manager->initialized || !chat_id || !user_id) return false;
    
    char endpoint[256];
    snprintf(endpoint, sizeof(endpoint), "/chats/%s/participants/%s", chat_id, user_id);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("DELETE");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    bool success = response && response->status_code == 200;
    if (response) http_response_free(response);
    
    return success;
}

// 实时功能
bool chat_manager_connect_websocket(chat_manager_t* manager) {
    if (!manager || !manager->initialized) return false;
    
    if (manager->websocket_connected) {
        return true;
    }
    
    manager->should_stop_websocket = false;
    
    // 创建WebSocket线程
    if (pthread_create(&manager->websocket_thread, NULL, websocket_thread_func, manager) != 0) {
        return false;
    }
    
    // 启动自动重连线程
    if (manager->auto_reconnect_enabled) {
        manager->should_stop_reconnect = false;
        if (pthread_create(&manager->reconnect_thread, NULL, reconnect_thread_func, manager) != 0) {
            manager->should_stop_websocket = true;
            pthread_join(manager->websocket_thread, NULL);
            return false;
        }
    }
    
    return true;
}

void chat_manager_disconnect_websocket(chat_manager_t* manager) {
    if (!manager) return;
    
    manager->should_stop_websocket = true;
    manager->websocket_connected = false;
    
    if (manager->websocket_handle) {
        curl_easy_cleanup(manager->websocket_handle);
        manager->websocket_handle = NULL;
    }
    
    if (manager->websocket_thread) {
        pthread_join(manager->websocket_thread, NULL);
        manager->websocket_thread = 0;
    }
}

bool chat_manager_is_websocket_connected(const chat_manager_t* manager) {
    return manager && manager->websocket_connected;
}

bool chat_manager_send_typing_status(chat_manager_t* manager, const char* chat_id, bool is_typing) {
    if (!manager || !manager->initialized || !chat_id) return false;
    
    // TODO: 通过WebSocket发送打字状态
    // 这里需要实现WebSocket消息发送功能
    
    return true;
}

// 文件传输
bool chat_manager_send_file(chat_manager_t* manager, const char* chat_id, const char* file_path) {
    if (!manager || !manager->initialized || !chat_id || !file_path) return false;
    
    // TODO: 实现文件上传功能
    // 1. 先上传文件到服务器
    // 2. 获取文件URL
    // 3. 发送文件消息
    
    return false;
}

bool chat_manager_download_file(chat_manager_t* manager, const char* file_url, const char* save_path) {
    if (!manager || !manager->initialized || !file_url || !save_path) return false;
    
    // TODO: 实现文件下载功能
    
    return false;
}

// 搜索功能
bool chat_manager_search_messages(chat_manager_t* manager, const char* query, const char* chat_id) {
    if (!manager || !manager->initialized || !query) return false;
    
    char endpoint[512];
    if (chat_id) {
        snprintf(endpoint, sizeof(endpoint), "/messages/search?q=%s&chat_id=%s", query, chat_id);
    } else {
        snprintf(endpoint, sizeof(endpoint), "/messages/search?q=%s", query);
    }
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("GET");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    if (!response || response->status_code != 200) {
        if (response) http_response_free(response);
        return false;
    }
    
    // 解析搜索结果
    chat_response_t* chat_response = chat_response_create_from_json(response->body);
    http_response_free(response);
    
    if (!chat_response || !chat_response->success) {
        if (chat_response) chat_response_free(chat_response);
        return false;
    }
    
    // 触发回调
    if (manager->on_messages_updated) {
        manager->on_messages_updated(chat_response->messages, chat_response->messages_count);
    }
    
    chat_response_free(chat_response);
    return true;
}

bool chat_manager_search_chats(chat_manager_t* manager, const char* query) {
    if (!manager || !manager->initialized || !query) return false;
    
    char endpoint[512];
    snprintf(endpoint, sizeof(endpoint), "/chats/search?q=%s", query);
    
    char* url = build_chat_url(manager, endpoint);
    if (!url) return false;
    
    http_request_t request = {0};
    request.url = url;
    request.method = strdup("GET");
    add_auth_headers(&request);
    
    http_response_t* response = http_client_request(manager->http_client, &request);
    
    free(url);
    free(request.method);
    if (request.headers) {
        for (int i = 0; request.headers[i]; i++) {
            free(request.headers[i]);
        }
        free(request.headers);
    }
    
    if (!response || response->status_code != 200) {
        if (response) http_response_free(response);
        return false;
    }
    
    // 解析搜索结果
    chat_response_t* chat_response = chat_response_create_from_json(response->body);
    http_response_free(response);
    
    if (!chat_response || !chat_response->success) {
        if (chat_response) chat_response_free(chat_response);
        return false;
    }
    
    // 触发回调
    if (manager->on_chats_updated) {
        manager->on_chats_updated(chat_response->chats, chat_response->chats_count);
    }
    
    chat_response_free(chat_response);
    return true;
}

// 本地数据管理
chat_t* chat_manager_find_chat_by_id(chat_manager_t* manager, const char* chat_id) {
    if (!manager || !chat_id) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->chats_count; i++) {
        if (strcmp(manager->chats[i].id, chat_id) == 0) {
            pthread_mutex_unlock(&manager->data_mutex);
            return &manager->chats[i];
        }
    }
    pthread_mutex_unlock(&manager->data_mutex);
    
    return NULL;
}

chat_t* chat_manager_find_chat_by_participant(chat_manager_t* manager, const char* user_id) {
    if (!manager || !user_id) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->chats_count; i++) {
        if (manager->chats[i].type == CHAT_TYPE_PRIVATE) {
            for (size_t j = 0; j < manager->chats[i].participants_count; j++) {
                if (strcmp(manager->chats[i].participants[j], user_id) == 0) {
                    pthread_mutex_unlock(&manager->data_mutex);
                    return &manager->chats[i];
                }
            }
        }
    }
    pthread_mutex_unlock(&manager->data_mutex);
    
    return NULL;
}

message_t* chat_manager_find_message_by_id(chat_manager_t* manager, const char* message_id) {
    if (!manager || !message_id) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->chat_messages_capacity; i++) {
        if (manager->chat_messages[i]) {
            for (size_t j = 0; j < manager->chat_message_counts[i]; j++) {
                if (strcmp(manager->chat_messages[i][j].id, message_id) == 0) {
                    pthread_mutex_unlock(&manager->data_mutex);
                    return &manager->chat_messages[i][j];
                }
            }
        }
    }
    pthread_mutex_unlock(&manager->data_mutex);
    
    return NULL;
}

message_t* chat_manager_get_chat_messages(chat_manager_t* manager, const char* chat_id, size_t* count) {
    if (!manager || !chat_id || !count) return NULL;
    
    *count = 0;
    
    // 找到聊天索引
    size_t chat_index = SIZE_MAX;
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->chats_count; i++) {
        if (strcmp(manager->chats[i].id, chat_id) == 0) {
            chat_index = i;
            break;
        }
    }
    
    if (chat_index == SIZE_MAX || chat_index >= manager->chat_messages_capacity || 
        !manager->chat_messages[chat_index]) {
        pthread_mutex_unlock(&manager->data_mutex);
        return NULL;
    }
    
    *count = manager->chat_message_counts[chat_index];
    message_t* messages = manager->chat_messages[chat_index];
    pthread_mutex_unlock(&manager->data_mutex);
    
    return messages;
}

// 事件回调设置
void chat_manager_set_on_chats_updated_callback(chat_manager_t* manager, on_chats_updated_callback_t callback) {
    if (manager) {
        manager->on_chats_updated = callback;
    }
}

void chat_manager_set_on_messages_updated_callback(chat_manager_t* manager, on_messages_updated_callback_t callback) {
    if (manager) {
        manager->on_messages_updated = callback;
    }
}

void chat_manager_set_on_new_message_callback(chat_manager_t* manager, on_new_message_callback_t callback) {
    if (manager) {
        manager->on_new_message = callback;
    }
}

void chat_manager_set_on_message_status_updated_callback(chat_manager_t* manager, on_message_status_updated_callback_t callback) {
    if (manager) {
        manager->on_message_status_updated = callback;
    }
}

void chat_manager_set_on_typing_status_callback(chat_manager_t* manager, on_typing_status_callback_t callback) {
    if (manager) {
        manager->on_typing_status = callback;
    }
}

void chat_manager_set_on_error_callback(chat_manager_t* manager, on_error_callback_t callback) {
    if (manager) {
        manager->on_error = callback;
    }
}

// 配置方法
void chat_manager_set_server_url(chat_manager_t* manager, const char* url) {
    if (manager && url) {
        free(manager->server_url);
        manager->server_url = strdup(url);
        if (manager->http_client) {
            http_client_set_base_url(manager->http_client, url);
        }
    }
}

void chat_manager_set_websocket_url(chat_manager_t* manager, const char* url) {
    if (manager && url) {
        free(manager->websocket_url);
        manager->websocket_url = strdup(url);
    }
}

void chat_manager_enable_auto_reconnect(chat_manager_t* manager, bool enable) {
    if (manager) {
        manager->auto_reconnect_enabled = enable;
    }
}

void chat_manager_set_reconnect_interval(chat_manager_t* manager, int seconds) {
    if (manager && seconds > 0) {
        manager->reconnect_interval = seconds;
    }
}

// 状态查询
bool chat_manager_is_initialized(const chat_manager_t* manager) {
    return manager && manager->initialized;
}

int chat_manager_get_unread_message_count(const chat_manager_t* manager) {
    if (!manager) return 0;
    
    int total_unread = 0;
    pthread_mutex_lock((pthread_mutex_t*)&manager->data_mutex);
    for (size_t i = 0; i < manager->chats_count; i++) {
        total_unread += manager->chats[i].unread_count;
    }
    pthread_mutex_unlock((pthread_mutex_t*)&manager->data_mutex);
    
    return total_unread;
}

int chat_manager_get_chat_count(const chat_manager_t* manager) {
    return manager ? (int)manager->chats_count : 0;
}

// 辅助函数实现
const char* message_type_to_string(message_type_t type) {
    switch (type) {
        case MESSAGE_TYPE_TEXT: return "text";
        case MESSAGE_TYPE_IMAGE: return "image";
        case MESSAGE_TYPE_FILE: return "file";
        case MESSAGE_TYPE_SYSTEM: return "system";
        case MESSAGE_TYPE_EMOJI: return "emoji";
        default: return "text";
    }
}

message_type_t string_to_message_type(const char* type) {
    if (!type) return MESSAGE_TYPE_TEXT;
    if (strcmp(type, "text") == 0) return MESSAGE_TYPE_TEXT;
    if (strcmp(type, "image") == 0) return MESSAGE_TYPE_IMAGE;
    if (strcmp(type, "file") == 0) return MESSAGE_TYPE_FILE;
    if (strcmp(type, "system") == 0) return MESSAGE_TYPE_SYSTEM;
    if (strcmp(type, "emoji") == 0) return MESSAGE_TYPE_EMOJI;
    return MESSAGE_TYPE_TEXT;
}

const char* chat_type_to_string(chat_type_t type) {
    switch (type) {
        case CHAT_TYPE_PRIVATE: return "private";
        case CHAT_TYPE_GROUP: return "group";
        default: return "private";
    }
}

chat_type_t string_to_chat_type(const char* type) {
    if (!type) return CHAT_TYPE_PRIVATE;
    if (strcmp(type, "private") == 0) return CHAT_TYPE_PRIVATE;
    if (strcmp(type, "group") == 0) return CHAT_TYPE_GROUP;
    return CHAT_TYPE_PRIVATE;
}

const char* message_status_to_string(message_status_t status) {
    switch (status) {
        case MESSAGE_STATUS_SENDING: return "sending";
        case MESSAGE_STATUS_SENT: return "sent";
        case MESSAGE_STATUS_DELIVERED: return "delivered";
        case MESSAGE_STATUS_READ: return "read";
        case MESSAGE_STATUS_FAILED: return "failed";
        default: return "sent";
    }
}

message_status_t string_to_message_status(const char* status) {
    if (!status) return MESSAGE_STATUS_SENT;
    if (strcmp(status, "sending") == 0) return MESSAGE_STATUS_SENDING;
    if (strcmp(status, "sent") == 0) return MESSAGE_STATUS_SENT;
    if (strcmp(status, "delivered") == 0) return MESSAGE_STATUS_DELIVERED;
    if (strcmp(status, "read") == 0) return MESSAGE_STATUS_READ;
    if (strcmp(status, "failed") == 0) return MESSAGE_STATUS_FAILED;
    return MESSAGE_STATUS_SENT;
}

// 内存管理函数实现
message_t* message_create_from_json(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    message_t* message = (message_t*)calloc(1, sizeof(message_t));
    if (!message) {
        json_object_put(root);
        return NULL;
    }
    
    json_object* obj;
    
    if (json_object_object_get_ex(root, "id", &obj)) {
        message->id = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "chat_id", &obj)) {
        message->chat_id = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "sender_id", &obj)) {
        message->sender_id = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "sender_username", &obj)) {
        message->sender_username = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "content", &obj)) {
        message->content = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "type", &obj)) {
        message->type = string_to_message_type(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "status", &obj)) {
        message->status = string_to_message_status(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "timestamp", &obj)) {
        message->timestamp = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "created_at", &obj)) {
        message->created_at = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "updated_at", &obj)) {
        message->updated_at = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "file_name", &obj)) {
        message->file_name = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "file_url", &obj)) {
        message->file_url = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "file_size", &obj)) {
        message->file_size = json_object_get_int64(obj);
    }
    
    if (json_object_object_get_ex(root, "reply_to_message_id", &obj)) {
        message->reply_to_message_id = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "reply_to_content", &obj)) {
        message->reply_to_content = strdup(json_object_get_string(obj));
    }
    
    json_object_put(root);
    return message;
}

void message_free(message_t* message) {
    if (!message) return;
    
    free(message->id);
    free(message->chat_id);
    free(message->sender_id);
    free(message->sender_username);
    free(message->content);
    free(message->timestamp);
    free(message->created_at);
    free(message->updated_at);
    free(message->file_name);
    free(message->file_url);
    free(message->reply_to_message_id);
    free(message->reply_to_content);
    
    memset(message, 0, sizeof(message_t));
}

chat_t* chat_create_from_json(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    chat_t* chat = (chat_t*)calloc(1, sizeof(chat_t));
    if (!chat) {
        json_object_put(root);
        return NULL;
    }
    
    json_object* obj;
    
    if (json_object_object_get_ex(root, "id", &obj)) {
        chat->id = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "name", &obj)) {
        chat->name = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "type", &obj)) {
        chat->type = string_to_chat_type(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "avatar_url", &obj)) {
        chat->avatar_url = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "last_message", &obj)) {
        chat->last_message = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "last_message_time", &obj)) {
        chat->last_message_time = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "unread_count", &obj)) {
        chat->unread_count = json_object_get_int(obj);
    }
    
    if (json_object_object_get_ex(root, "participants", &obj)) {
        int array_len = json_object_array_length(obj);
        chat->participants = (char**)calloc(array_len, sizeof(char*));
        chat->participants_count = array_len;
        
        for (int i = 0; i < array_len; i++) {
            json_object* participant = json_object_array_get_idx(obj, i);
            chat->participants[i] = strdup(json_object_get_string(participant));
        }
    }
    
    if (json_object_object_get_ex(root, "created_at", &obj)) {
        chat->created_at = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "updated_at", &obj)) {
        chat->updated_at = strdup(json_object_get_string(obj));
    }
    
    json_object_put(root);
    return chat;
}

void chat_free(chat_t* chat) {
    if (!chat) return;
    
    free(chat->id);
    free(chat->name);
    free(chat->avatar_url);
    free(chat->last_message);
    free(chat->last_message_time);
    free(chat->created_at);
    free(chat->updated_at);
    
    if (chat->participants) {
        for (size_t i = 0; i < chat->participants_count; i++) {
            free(chat->participants[i]);
        }
        free(chat->participants);
    }
    
    memset(chat, 0, sizeof(chat_t));
}

send_message_request_t* send_message_request_create(const char* chat_id, const char* content, message_type_t type) {
    if (!chat_id || !content) return NULL;
    
    send_message_request_t* request = (send_message_request_t*)calloc(1, sizeof(send_message_request_t));
    if (!request) return NULL;
    
    request->chat_id = strdup(chat_id);
    request->content = strdup(content);
    request->type = type;
    
    return request;
}

void send_message_request_free(send_message_request_t* request) {
    if (!request) return;
    
    free(request->chat_id);
    free(request->content);
    free(request->reply_to_message_id);
    
    free(request);
}

create_chat_request_t* create_chat_request_create(chat_type_t type, const char* name, const char** participants, size_t participants_count) {
    create_chat_request_t* request = (create_chat_request_t*)calloc(1, sizeof(create_chat_request_t));
    if (!request) return NULL;
    
    request->type = type;
    if (name) {
        request->name = strdup(name);
    }
    
    if (participants && participants_count > 0) {
        request->participants = (char**)calloc(participants_count, sizeof(char*));
        request->participants_count = participants_count;
        
        for (size_t i = 0; i < participants_count; i++) {
            request->participants[i] = strdup(participants[i]);
        }
    }
    
    return request;
}

void create_chat_request_free(create_chat_request_t* request) {
    if (!request) return;
    
    free(request->name);
    
    if (request->participants) {
        for (size_t i = 0; i < request->participants_count; i++) {
            free(request->participants[i]);
        }
        free(request->participants);
    }
    
    free(request);
}

chat_response_t* chat_response_create_from_json(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    chat_response_t* response = (chat_response_t*)calloc(1, sizeof(chat_response_t));
    if (!response) {
        json_object_put(root);
        return NULL;
    }
    
    json_object* obj;
    
    if (json_object_object_get_ex(root, "success", &obj)) {
        response->success = json_object_get_boolean(obj);
    }
    
    if (json_object_object_get_ex(root, "message", &obj)) {
        response->message = strdup(json_object_get_string(obj));
    }
    
    // 解析聊天列表
    if (json_object_object_get_ex(root, "chats", &obj)) {
        int array_len = json_object_array_length(obj);
        response->chats = (chat_t*)calloc(array_len, sizeof(chat_t));
        response->chats_count = array_len;
        
        for (int i = 0; i < array_len; i++) {
            json_object* chat_obj = json_object_array_get_idx(obj, i);
            const char* chat_json = json_object_to_json_string(chat_obj);
            chat_t* chat = chat_create_from_json(chat_json);
            if (chat) {
                response->chats[i] = *chat;
                free(chat);
            }
        }
    }
    
    // 解析消息列表
    if (json_object_object_get_ex(root, "messages", &obj)) {
        int array_len = json_object_array_length(obj);
        response->messages = (message_t*)calloc(array_len, sizeof(message_t));
        response->messages_count = array_len;
        
        for (int i = 0; i < array_len; i++) {
            json_object* message_obj = json_object_array_get_idx(obj, i);
            const char* message_json = json_object_to_json_string(message_obj);
            message_t* message = message_create_from_json(message_json);
            if (message) {
                response->messages[i] = *message;
                free(message);
            }
        }
    }
    
    // 解析单个聊天
    if (json_object_object_get_ex(root, "chat", &obj)) {
        const char* chat_json = json_object_to_json_string(obj);
        chat_t* chat = chat_create_from_json(chat_json);
        if (chat) {
            response->chat = *chat;
            free(chat);
        }
    }
    
    // 解析单个消息
    if (json_object_object_get_ex(root, "message_data", &obj)) {
        const char* message_json = json_object_to_json_string(obj);
        message_t* message = message_create_from_json(message_json);
        if (message) {
            response->message_data = *message;
            free(message);
        }
    }
    
    json_object_put(root);
    return response;
}

void chat_response_free(chat_response_t* response) {
    if (!response) return;
    
    free(response->message);
    
    if (response->chats) {
        for (size_t i = 0; i < response->chats_count; i++) {
            chat_free(&response->chats[i]);
        }
        free(response->chats);
    }
    
    if (response->messages) {
        for (size_t i = 0; i < response->messages_count; i++) {
            message_free(&response->messages[i]);
        }
        free(response->messages);
    }
    
    chat_free(&response->chat);
    message_free(&response->message_data);
    
    free(response);
}

websocket_message_t* websocket_message_create_from_json(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    websocket_message_t* ws_message = (websocket_message_t*)calloc(1, sizeof(websocket_message_t));
    if (!ws_message) {
        json_object_put(root);
        return NULL;
    }
    
    json_object* obj;
    
    if (json_object_object_get_ex(root, "type", &obj)) {
        ws_message->type = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "chat_id", &obj)) {
        ws_message->chat_id = strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "data", &obj)) {
        ws_message->data = strdup(json_object_to_json_string(obj));
    }
    
    if (json_object_object_get_ex(root, "timestamp", &obj)) {
        ws_message->timestamp = strdup(json_object_get_string(obj));
    }
    
    json_object_put(root);
    return ws_message;
}

void websocket_message_free(websocket_message_t* ws_message) {
    if (!ws_message) return;
    
    free(ws_message->type);
    free(ws_message->chat_id);
    free(ws_message->data);
    free(ws_message->timestamp);
    
    free(ws_message);
}

// 全局函数实现
bool chat_manager_init(void) {
    if (g_chat_manager) {
        return true;
    }
    
    g_chat_manager = chat_manager_create();
    if (!g_chat_manager) {
        return false;
    }
    
    return chat_manager_initialize(g_chat_manager);
}

void chat_manager_cleanup_global(void) {
    if (g_chat_manager) {
        chat_manager_destroy(g_chat_manager);
        g_chat_manager = NULL;
    }
}

chat_manager_t* chat_manager_get_instance(void) {
    return g_chat_manager;
}

// 内部辅助函数实现
static char* build_chat_url(const chat_manager_t* manager, const char* endpoint) {
    if (!manager || !manager->server_url || !endpoint) return NULL;
    
    size_t url_len = strlen(manager->server_url) + strlen(endpoint) + 1;
    char* url = (char*)malloc(url_len);
    if (!url) return NULL;
    
    snprintf(url, url_len, "%s%s", manager->server_url, endpoint);
    return url;
}

static char* build_websocket_url(const chat_manager_t* manager) {
    if (!manager || !manager->websocket_url) return NULL;
    
    // 添加认证token到WebSocket URL
    auth_manager_t* auth = auth_manager_get_instance();
    if (!auth) return strdup(manager->websocket_url);
    
    const char* token = auth_manager_get_access_token(auth);
    if (!token) return strdup(manager->websocket_url);
    
    size_t url_len = strlen(manager->websocket_url) + strlen(token) + 20;
    char* url = (char*)malloc(url_len);
    if (!url) return NULL;
    
    snprintf(url, url_len, "%s?token=%s", manager->websocket_url, token);
    return url;
}

static json_object* message_to_json(const message_t* message) {
    if (!message) return NULL;
    
    json_object* json = json_object_new_object();
    
    if (message->id) {
        json_object_object_add(json, "id", json_object_new_string(message->id));
    }
    if (message->chat_id) {
        json_object_object_add(json, "chat_id", json_object_new_string(message->chat_id));
    }
    if (message->sender_id) {
        json_object_object_add(json, "sender_id", json_object_new_string(message->sender_id));
    }
    if (message->sender_username) {
        json_object_object_add(json, "sender_username", json_object_new_string(message->sender_username));
    }
    if (message->content) {
        json_object_object_add(json, "content", json_object_new_string(message->content));
    }
    
    json_object_object_add(json, "type", json_object_new_string(message_type_to_string(message->type)));
    json_object_object_add(json, "status", json_object_new_string(message_status_to_string(message->status)));
    
    if (message->timestamp) {
        json_object_object_add(json, "timestamp", json_object_new_string(message->timestamp));
    }
    if (message->created_at) {
        json_object_object_add(json, "created_at", json_object_new_string(message->created_at));
    }
    if (message->updated_at) {
        json_object_object_add(json, "updated_at", json_object_new_string(message->updated_at));
    }
    
    if (message->file_name) {
        json_object_object_add(json, "file_name", json_object_new_string(message->file_name));
    }
    if (message->file_url) {
        json_object_object_add(json, "file_url", json_object_new_string(message->file_url));
    }
    if (message->file_size > 0) {
        json_object_object_add(json, "file_size", json_object_new_int64(message->file_size));
    }
    
    if (message->reply_to_message_id) {
        json_object_object_add(json, "reply_to_message_id", json_object_new_string(message->reply_to_message_id));
    }
    if (message->reply_to_content) {
        json_object_object_add(json, "reply_to_content", json_object_new_string(message->reply_to_content));
    }
    
    return json;
}

static json_object* chat_to_json(const chat_t* chat) {
    if (!chat) return NULL;
    
    json_object* json = json_object_new_object();
    
    if (chat->id) {
        json_object_object_add(json, "id", json_object_new_string(chat->id));
    }
    if (chat->name) {
        json_object_object_add(json, "name", json_object_new_string(chat->name));
    }
    
    json_object_object_add(json, "type", json_object_new_string(chat_type_to_string(chat->type)));
    
    if (chat->avatar_url) {
        json_object_object_add(json, "avatar_url", json_object_new_string(chat->avatar_url));
    }
    if (chat->last_message) {
        json_object_object_add(json, "last_message", json_object_new_string(chat->last_message));
    }
    if (chat->last_message_time) {
        json_object_object_add(json, "last_message_time", json_object_new_string(chat->last_message_time));
    }
    
    json_object_object_add(json, "unread_count", json_object_new_int(chat->unread_count));
    
    if (chat->participants && chat->participants_count > 0) {
        json_object* participants_array = json_object_new_array();
        for (size_t i = 0; i < chat->participants_count; i++) {
            json_object_array_add(participants_array, json_object_new_string(chat->participants[i]));
        }
        json_object_object_add(json, "participants", participants_array);
    }
    
    if (chat->created_at) {
        json_object_object_add(json, "created_at", json_object_new_string(chat->created_at));
    }
    if (chat->updated_at) {
        json_object_object_add(json, "updated_at", json_object_new_string(chat->updated_at));
    }
    
    return json;
}

static json_object* send_message_request_to_json(const send_message_request_t* request) {
    if (!request) return NULL;
    
    json_object* json = json_object_new_object();
    
    if (request->chat_id) {
        json_object_object_add(json, "chat_id", json_object_new_string(request->chat_id));
    }
    if (request->content) {
        json_object_object_add(json, "content", json_object_new_string(request->content));
    }
    
    json_object_object_add(json, "type", json_object_new_string(message_type_to_string(request->type)));
    
    if (request->reply_to_message_id) {
        json_object_object_add(json, "reply_to_message_id", json_object_new_string(request->reply_to_message_id));
    }
    
    return json;
}

static json_object* create_chat_request_to_json(const create_chat_request_t* request) {
    if (!request) return NULL;
    
    json_object* json = json_object_new_object();
    
    json_object_object_add(json, "type", json_object_new_string(chat_type_to_string(request->type)));
    
    if (request->name) {
        json_object_object_add(json, "name", json_object_new_string(request->name));
    }
    
    if (request->participants && request->participants_count > 0) {
        json_object* participants_array = json_object_new_array();
        for (size_t i = 0; i < request->participants_count; i++) {
            json_object_array_add(participants_array, json_object_new_string(request->participants[i]));
        }
        json_object_object_add(json, "participants", participants_array);
    }
    
    return json;
}

static void* websocket_thread_func(void* arg) {
    chat_manager_t* manager = (chat_manager_t*)arg;
    if (!manager) return NULL;
    
    while (!manager->should_stop_websocket) {
        char* ws_url = build_websocket_url(manager);
        if (!ws_url) {
            sleep(manager->reconnect_interval);
            continue;
        }
        
        CURL* curl = curl_easy_init();
        if (!curl) {
            free(ws_url);
            sleep(manager->reconnect_interval);
            continue;
        }
        
        curl_easy_setopt(curl, CURLOPT_URL, ws_url);
        curl_easy_setopt(curl, CURLOPT_CONNECT_ONLY, 2L); // WebSocket
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, websocket_write_callback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, manager);
        
        CURLcode res = curl_easy_perform(curl);
        
        if (res == CURLE_OK) {
            manager->websocket_connected = true;
            manager->websocket_handle = curl;
            
            // WebSocket消息接收循环
            while (!manager->should_stop_websocket && manager->websocket_connected) {
                size_t rlen;
                char buffer[1024];
                
                res = curl_ws_recv(curl, buffer, sizeof(buffer), &rlen, NULL);
                if (res == CURLE_OK && rlen > 0) {
                    buffer[rlen] = '\0';
                    handle_websocket_message(manager, buffer);
                } else if (res != CURLE_AGAIN) {
                    manager->websocket_connected = false;
                    break;
                }
                
                usleep(10000); // 10ms
            }
        }
        
        curl_easy_cleanup(curl);
        manager->websocket_handle = NULL;
        manager->websocket_connected = false;
        free(ws_url);
        
        if (!manager->should_stop_websocket && manager->auto_reconnect_enabled) {
            sleep(manager->reconnect_interval);
        }
    }
    
    return NULL;
}

static void* reconnect_thread_func(void* arg) {
    chat_manager_t* manager = (chat_manager_t*)arg;
    if (!manager) return NULL;
    
    while (!manager->should_stop_reconnect) {
        sleep(manager->reconnect_interval);
        
        if (!manager->websocket_connected && !manager->should_stop_websocket) {
            // 尝试重连
            if (manager->on_error) {
                manager->on_error("WebSocket disconnected, attempting to reconnect...");
            }
        }
    }
    
    return NULL;
}

static void handle_websocket_message(chat_manager_t* manager, const char* message) {
    if (!manager || !message) return;
    
    websocket_message_t* ws_message = websocket_message_create_from_json(message);
    if (!ws_message) return;
    
    if (strcmp(ws_message->type, "new_message") == 0) {
        message_t* new_message = message_create_from_json(ws_message->data);
        if (new_message) {
            add_local_message(manager, new_message);
            
            if (manager->on_new_message) {
                manager->on_new_message(new_message);
            }
            
            message_free(new_message);
            free(new_message);
        }
    } else if (strcmp(ws_message->type, "message_status_updated") == 0) {
        message_t* updated_message = message_create_from_json(ws_message->data);
        if (updated_message) {
            // 更新本地消息状态
            message_t* local_message = chat_manager_find_message_by_id(manager, updated_message->id);
            if (local_message) {
                local_message->status = updated_message->status;
                
                if (manager->on_message_status_updated) {
                    manager->on_message_status_updated(local_message);
                }
            }
            
            message_free(updated_message);
            free(updated_message);
        }
    } else if (strcmp(ws_message->type, "typing_status") == 0) {
        json_object* data = json_tokener_parse(ws_message->data);
        if (data) {
            json_object* user_id_obj, *is_typing_obj;
            
            if (json_object_object_get_ex(data, "user_id", &user_id_obj) &&
                json_object_object_get_ex(data, "is_typing", &is_typing_obj)) {
                
                const char* user_id = json_object_get_string(user_id_obj);
                bool is_typing = json_object_get_boolean(is_typing_obj);
                
                if (manager->on_typing_status) {
                    manager->on_typing_status(ws_message->chat_id, user_id, is_typing);
                }
            }
            
            json_object_put(data);
        }
    }
    
    websocket_message_free(ws_message);
}

static void add_auth_headers(http_request_t* request) {
    if (!request) return;
    
    auth_manager_t* auth = auth_manager_get_instance();
    if (!auth) return;
    
    const char* token = auth_manager_get_access_token(auth);
    if (!token) return;
    
    // 计算现有头部数量
    int header_count = 0;
    if (request->headers) {
        while (request->headers[header_count]) header_count++;
    }
    
    // 添加Authorization头
    request->headers = realloc(request->headers, (header_count + 2) * sizeof(char*));
    
    size_t auth_header_len = strlen("Authorization: Bearer ") + strlen(token) + 1;
    char* auth_header = (char*)malloc(auth_header_len);
    snprintf(auth_header, auth_header_len, "Authorization: Bearer %s", token);
    
    request->headers[header_count] = auth_header;
    request->headers[header_count + 1] = NULL;
}

static void update_local_chats(chat_manager_t* manager, const chat_t* chats, size_t count) {
    if (!manager || !chats) return;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // 清理旧数据
    if (manager->chats) {
        for (size_t i = 0; i < manager->chats_count; i++) {
            chat_free(&manager->chats[i]);
        }
        free(manager->chats);
    }
    
    // 复制新数据
    manager->chats = (chat_t*)calloc(count, sizeof(chat_t));
    manager->chats_count = count;
    
    for (size_t i = 0; i < count; i++) {
        manager->chats[i] = chats[i];
        
        // 深拷贝字符串字段
        if (chats[i].id) manager->chats[i].id = strdup(chats[i].id);
        if (chats[i].name) manager->chats[i].name = strdup(chats[i].name);
        if (chats[i].avatar_url) manager->chats[i].avatar_url = strdup(chats[i].avatar_url);
        if (chats[i].last_message) manager->chats[i].last_message = strdup(chats[i].last_message);
        if (chats[i].last_message_time) manager->chats[i].last_message_time = strdup(chats[i].last_message_time);
        if (chats[i].created_at) manager->chats[i].created_at = strdup(chats[i].created_at);
        if (chats[i].updated_at) manager->chats[i].updated_at = strdup(chats[i].updated_at);
        
        // 深拷贝参与者列表
        if (chats[i].participants && chats[i].participants_count > 0) {
            manager->chats[i].participants = (char**)calloc(chats[i].participants_count, sizeof(char*));
            for (size_t j = 0; j < chats[i].participants_count; j++) {
                manager->chats[i].participants[j] = strdup(chats[i].participants[j]);
            }
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
}

static void update_local_messages(chat_manager_t* manager, const char* chat_id, const message_t* messages, size_t count) {
    if (!manager || !chat_id || !messages) return;
    
    // 找到聊天索引
    size_t chat_index = SIZE_MAX;
    pthread_mutex_lock(&manager->data_mutex);
    
    for (size_t i = 0; i < manager->chats_count; i++) {
        if (strcmp(manager->chats[i].id, chat_id) == 0) {
            chat_index = i;
            break;
        }
    }
    
    if (chat_index == SIZE_MAX) {
        pthread_mutex_unlock(&manager->data_mutex);
        return;
    }
    
    // 确保消息数组有足够空间
    if (chat_index >= manager->chat_messages_capacity) {
        size_t new_capacity = chat_index + 1;
        manager->chat_messages = realloc(manager->chat_messages, new_capacity * sizeof(message_t*));
        manager->chat_message_counts = realloc(manager->chat_message_counts, new_capacity * sizeof(size_t));
        
        // 初始化新空间
        for (size_t i = manager->chat_messages_capacity; i < new_capacity; i++) {
            manager->chat_messages[i] = NULL;
            manager->chat_message_counts[i] = 0;
        }
        
        manager->chat_messages_capacity = new_capacity;
    }
    
    // 清理旧消息
    if (manager->chat_messages[chat_index]) {
        for (size_t i = 0; i < manager->chat_message_counts[chat_index]; i++) {
            message_free(&manager->chat_messages[chat_index][i]);
        }
        free(manager->chat_messages[chat_index]);
    }
    
    // 复制新消息
    manager->chat_messages[chat_index] = (message_t*)calloc(count, sizeof(message_t));
    manager->chat_message_counts[chat_index] = count;
    
    for (size_t i = 0; i < count; i++) {
        manager->chat_messages[chat_index][i] = messages[i];
        
        // 深拷贝字符串字段
        if (messages[i].id) manager->chat_messages[chat_index][i].id = strdup(messages[i].id);
        if (messages[i].chat_id) manager->chat_messages[chat_index][i].chat_id = strdup(messages[i].chat_id);
        if (messages[i].sender_id) manager->chat_messages[chat_index][i].sender_id = strdup(messages[i].sender_id);
        if (messages[i].sender_username) manager->chat_messages[chat_index][i].sender_username = strdup(messages[i].sender_username);
        if (messages[i].content) manager->chat_messages[chat_index][i].content = strdup(messages[i].content);
        if (messages[i].timestamp) manager->chat_messages[chat_index][i].timestamp = strdup(messages[i].timestamp);
        if (messages[i].created_at) manager->chat_messages[chat_index][i].created_at = strdup(messages[i].created_at);
        if (messages[i].updated_at) manager->chat_messages[chat_index][i].updated_at = strdup(messages[i].updated_at);
        if (messages[i].file_name) manager->chat_messages[chat_index][i].file_name = strdup(messages[i].file_name);
        if (messages[i].file_url) manager->chat_messages[chat_index][i].file_url = strdup(messages[i].file_url);
        if (messages[i].reply_to_message_id) manager->chat_messages[chat_index][i].reply_to_message_id = strdup(messages[i].reply_to_message_id);
        if (messages[i].reply_to_content) manager->chat_messages[chat_index][i].reply_to_content = strdup(messages[i].reply_to_content);
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
}

static void add_local_message(chat_manager_t* manager, const message_t* message) {
    if (!manager || !message || !message->chat_id) return;
    
    // 找到聊天索引
    size_t chat_index = SIZE_MAX;
    pthread_mutex_lock(&manager->data_mutex);
    
    for (size_t i = 0; i < manager->chats_count; i++) {
        if (strcmp(manager->chats[i].id, message->chat_id) == 0) {
            chat_index = i;
            break;
        }
    }
    
    if (chat_index == SIZE_MAX) {
        pthread_mutex_unlock(&manager->data_mutex);
        return;
    }
    
    // 确保消息数组有足够空间
    if (chat_index >= manager->chat_messages_capacity) {
        size_t new_capacity = chat_index + 1;
        manager->chat_messages = realloc(manager->chat_messages, new_capacity * sizeof(message_t*));
        manager->chat_message_counts = realloc(manager->chat_message_counts, new_capacity * sizeof(size_t));
        
        for (size_t i = manager->chat_messages_capacity; i < new_capacity; i++) {
            manager->chat_messages[i] = NULL;
            manager->chat_message_counts[i] = 0;
        }
        
        manager->chat_messages_capacity = new_capacity;
    }
    
    // 添加新消息
    size_t current_count = manager->chat_message_counts[chat_index];
    manager->chat_messages[chat_index] = realloc(manager->chat_messages[chat_index], 
                                                (current_count + 1) * sizeof(message_t));
    
    // 复制消息数据
    manager->chat_messages[chat_index][current_count] = *message;
    
    // 深拷贝字符串字段
    if (message->id) manager->chat_messages[chat_index][current_count].id = strdup(message->id);
    if (message->chat_id) manager->chat_messages[chat_index][current_count].chat_id = strdup(message->chat_id);
    if (message->sender_id) manager->chat_messages[chat_index][current_count].sender_id = strdup(message->sender_id);
    if (message->sender_username) manager->chat_messages[chat_index][current_count].sender_username = strdup(message->sender_username);
    if (message->content) manager->chat_messages[chat_index][current_count].content = strdup(message->content);
    if (message->timestamp) manager->chat_messages[chat_index][current_count].timestamp = strdup(message->timestamp);
    if (message->created_at) manager->chat_messages[chat_index][current_count].created_at = strdup(message->created_at);
    if (message->updated_at) manager->chat_messages[chat_index][current_count].updated_at = strdup(message->updated_at);
    if (message->file_name) manager->chat_messages[chat_index][current_count].file_name = strdup(message->file_name);
    if (message->file_url) manager->chat_messages[chat_index][current_count].file_url = strdup(message->file_url);
    if (message->reply_to_message_id) manager->chat_messages[chat_index][current_count].reply_to_message_id = strdup(message->reply_to_message_id);
    if (message->reply_to_content) manager->chat_messages[chat_index][current_count].reply_to_content = strdup(message->reply_to_content);
    
    manager->chat_message_counts[chat_index]++;
    
    pthread_mutex_unlock(&manager->data_mutex);
}

static size_t websocket_write_callback(void* contents, size_t size, size_t nmemb, void* userp) {
    size_t realsize = size * nmemb;
    chat_manager_t* manager = (chat_manager_t*)userp;
    
    if (manager && contents && realsize > 0) {
        char* message = (char*)malloc(realsize + 1);
        if (message) {
            memcpy(message, contents, realsize);
            message[realsize] = '\0';
            handle_websocket_message(manager, message);
            free(message);
        }
    }
    
    return realsize;
}