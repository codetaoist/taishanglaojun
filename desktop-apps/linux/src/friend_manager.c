#include "friend_manager.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <json-c/json.h>

// 全局好友管理器实例
friend_manager_t* g_friend_manager = NULL;

// 初始化和清理函数
bool friend_manager_init(void) {
    if (g_friend_manager != NULL) {
        return true; // 已经初始化
    }
    
    g_friend_manager = friend_manager_create();
    return g_friend_manager != NULL;
}

void friend_manager_cleanup(void) {
    if (g_friend_manager != NULL) {
        friend_manager_destroy(g_friend_manager);
        g_friend_manager = NULL;
    }
}

// 好友管理器创建和销毁
friend_manager_t* friend_manager_create(void) {
    friend_manager_t* manager = (friend_manager_t*)malloc(sizeof(friend_manager_t));
    if (!manager) return NULL;
    
    memset(manager, 0, sizeof(friend_manager_t));
    
    manager->http_client = http_client_create();
    if (!manager->http_client) {
        free(manager);
        return NULL;
    }
    
    manager->server_url = strdup("http://localhost:8081");
    manager->current_online_status = ONLINE_STATUS_OFFLINE;
    manager->auto_refresh_enabled = true;
    manager->refresh_interval = 30;
    manager->is_running = false;
    
    // 初始化互斥锁
    if (pthread_mutex_init(&manager->data_mutex, NULL) != 0) {
        http_client_destroy(manager->http_client);
        free(manager->server_url);
        free(manager);
        return NULL;
    }
    
    return manager;
}

void friend_manager_destroy(friend_manager_t* manager) {
    if (!manager) return;
    
    // 停止刷新线程
    manager->is_running = false;
    if (manager->refresh_thread) {
        pthread_join(manager->refresh_thread, NULL);
    }
    
    // 清理HTTP客户端
    if (manager->http_client) {
        http_client_destroy(manager->http_client);
    }
    
    // 清理好友列表
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->friends_count; i++) {
        friend_free(&manager->friends[i]);
    }
    free(manager->friends);
    
    // 清理好友请求
    for (size_t i = 0; i < manager->pending_requests_count; i++) {
        friend_request_free(&manager->pending_requests[i]);
    }
    free(manager->pending_requests);
    pthread_mutex_unlock(&manager->data_mutex);
    
    // 清理其他资源
    free(manager->server_url);
    pthread_mutex_destroy(&manager->data_mutex);
    free(manager);
}

// JSON解析辅助函数
static json_object* create_add_friend_json(const char* username, const char* message) {
    json_object* json = json_object_new_object();
    json_object* username_obj = json_object_new_string(username);
    json_object* message_obj = json_object_new_string(message ? message : "");
    
    json_object_object_add(json, "username", username_obj);
    json_object_object_add(json, "message", message_obj);
    
    return json;
}

static json_object* create_respond_request_json(bool accept) {
    json_object* json = json_object_new_object();
    json_object* action_obj = json_object_new_string(accept ? "accept" : "decline");
    
    json_object_object_add(json, "action", action_obj);
    
    return json;
}

static friend_response_t* parse_friend_response(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    friend_response_t* response = (friend_response_t*)malloc(sizeof(friend_response_t));
    memset(response, 0, sizeof(friend_response_t));
    
    json_object* success_obj;
    if (json_object_object_get_ex(root, "success", &success_obj)) {
        response->success = json_object_get_boolean(success_obj);
    }
    
    json_object* message_obj;
    if (json_object_object_get_ex(root, "message", &message_obj)) {
        const char* message = json_object_get_string(message_obj);
        response->message = strdup(message);
    }
    
    // 解析好友列表
    json_object* friends_array;
    if (json_object_object_get_ex(root, "friends", &friends_array) && json_object_is_type(friends_array, json_type_array)) {
        int friends_count = json_object_array_length(friends_array);
        if (friends_count > 0) {
            response->friends = (friend_t*)malloc(sizeof(friend_t) * friends_count);
            response->friends_count = friends_count;
            
            for (int i = 0; i < friends_count; i++) {
                json_object* friend_obj = json_object_array_get_idx(friends_array, i);
                const char* friend_json = json_object_to_json_string(friend_obj);
                friend_t* friend = friend_create_from_json(friend_json);
                if (friend) {
                    response->friends[i] = *friend;
                    free(friend);
                }
            }
        }
    }
    
    // 解析好友请求
    json_object* requests_array;
    if (json_object_object_get_ex(root, "requests", &requests_array) && json_object_is_type(requests_array, json_type_array)) {
        int requests_count = json_object_array_length(requests_array);
        if (requests_count > 0) {
            response->requests = (friend_request_t*)malloc(sizeof(friend_request_t) * requests_count);
            response->requests_count = requests_count;
            
            for (int i = 0; i < requests_count; i++) {
                json_object* request_obj = json_object_array_get_idx(requests_array, i);
                const char* request_json = json_object_to_json_string(request_obj);
                friend_request_t* request = friend_request_create_from_json(request_json);
                if (request) {
                    response->requests[i] = *request;
                    free(request);
                }
            }
        }
    }
    
    json_object_put(root);
    return response;
}

// 同步方法实现
friend_response_t* friend_manager_get_friend_list(friend_manager_t* manager) {
    if (!manager) return NULL;
    
    char* url = friend_build_url(manager, "/api/friends");
    http_request_t* request = friend_create_authenticated_request(manager, url, "GET");
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    friend_response_t* friend_response = NULL;
    if (response && response->status_code == 200) {
        friend_response = parse_friend_response(response->body);
        
        // 更新本地好友列表
        if (friend_response && friend_response->success) {
            pthread_mutex_lock(&manager->data_mutex);
            
            // 清理旧数据
            for (size_t i = 0; i < manager->friends_count; i++) {
                friend_free(&manager->friends[i]);
            }
            free(manager->friends);
            
            // 复制新数据
            manager->friends_count = friend_response->friends_count;
            if (manager->friends_count > 0) {
                manager->friends = (friend_t*)malloc(sizeof(friend_t) * manager->friends_count);
                memcpy(manager->friends, friend_response->friends, sizeof(friend_t) * manager->friends_count);
            } else {
                manager->friends = NULL;
            }
            
            pthread_mutex_unlock(&manager->data_mutex);
        }
    }
    
    free(url);
    http_request_free(request);
    http_response_free(response);
    
    return friend_response;
}

friend_response_t* friend_manager_get_friend_requests(friend_manager_t* manager) {
    if (!manager) return NULL;
    
    char* url = friend_build_url(manager, "/api/friends/requests");
    http_request_t* request = friend_create_authenticated_request(manager, url, "GET");
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    friend_response_t* friend_response = NULL;
    if (response && response->status_code == 200) {
        friend_response = parse_friend_response(response->body);
        
        // 更新本地请求列表
        if (friend_response && friend_response->success) {
            pthread_mutex_lock(&manager->data_mutex);
            
            // 清理旧数据
            for (size_t i = 0; i < manager->pending_requests_count; i++) {
                friend_request_free(&manager->pending_requests[i]);
            }
            free(manager->pending_requests);
            
            // 复制新数据
            manager->pending_requests_count = friend_response->requests_count;
            if (manager->pending_requests_count > 0) {
                manager->pending_requests = (friend_request_t*)malloc(sizeof(friend_request_t) * manager->pending_requests_count);
                memcpy(manager->pending_requests, friend_response->requests, sizeof(friend_request_t) * manager->pending_requests_count);
            } else {
                manager->pending_requests = NULL;
            }
            
            pthread_mutex_unlock(&manager->data_mutex);
        }
    }
    
    free(url);
    http_request_free(request);
    http_response_free(response);
    
    return friend_response;
}

bool friend_manager_add_friend(friend_manager_t* manager, const char* username, const char* message) {
    if (!manager || !username) return false;
    
    json_object* json = create_add_friend_json(username, message);
    const char* json_str = json_object_to_json_string(json);
    
    char* url = friend_build_url(manager, "/api/friends/add");
    http_request_t* request = friend_create_authenticated_request(manager, url, "POST");
    http_request_set_body(request, json_str);
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    bool success = false;
    if (response && (response->status_code == 200 || response->status_code == 201)) {
        friend_response_t* friend_response = parse_friend_response(response->body);
        if (friend_response) {
            success = friend_response->success;
            friend_response_free(friend_response);
        }
    }
    
    json_object_put(json);
    free(url);
    http_request_free(request);
    http_response_free(response);
    
    return success;
}

bool friend_manager_respond_to_request(friend_manager_t* manager, const char* request_id, bool accept) {
    if (!manager || !request_id) return false;
    
    json_object* json = create_respond_request_json(accept);
    const char* json_str = json_object_to_json_string(json);
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/friends/requests/%s", manager->server_url, request_id);
    
    http_request_t* request = friend_create_authenticated_request(manager, url, "PUT");
    http_request_set_body(request, json_str);
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    bool success = false;
    if (response && response->status_code == 200) {
        friend_response_t* friend_response = parse_friend_response(response->body);
        if (friend_response) {
            success = friend_response->success;
            friend_response_free(friend_response);
        }
    }
    
    json_object_put(json);
    http_request_free(request);
    http_response_free(response);
    
    return success;
}

bool friend_manager_remove_friend(friend_manager_t* manager, const char* friend_id) {
    if (!manager || !friend_id) return false;
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/friends/%s", manager->server_url, friend_id);
    
    http_request_t* request = friend_create_authenticated_request(manager, url, "DELETE");
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    bool success = false;
    if (response && response->status_code == 200) {
        friend_response_t* friend_response = parse_friend_response(response->body);
        if (friend_response) {
            success = friend_response->success;
            friend_response_free(friend_response);
        }
    }
    
    http_request_free(request);
    http_response_free(response);
    
    return success;
}

bool friend_manager_block_friend(friend_manager_t* manager, const char* friend_id) {
    if (!manager || !friend_id) return false;
    
    json_object* json = json_object_new_object();
    json_object* action_obj = json_object_new_string("block");
    json_object_object_add(json, "action", action_obj);
    
    const char* json_str = json_object_to_json_string(json);
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/friends/%s/block", manager->server_url, friend_id);
    
    http_request_t* request = friend_create_authenticated_request(manager, url, "PUT");
    http_request_set_body(request, json_str);
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    bool success = false;
    if (response && response->status_code == 200) {
        friend_response_t* friend_response = parse_friend_response(response->body);
        if (friend_response) {
            success = friend_response->success;
            friend_response_free(friend_response);
        }
    }
    
    json_object_put(json);
    http_request_free(request);
    http_response_free(response);
    
    return success;
}

bool friend_manager_unblock_friend(friend_manager_t* manager, const char* friend_id) {
    if (!manager || !friend_id) return false;
    
    json_object* json = json_object_new_object();
    json_object* action_obj = json_object_new_string("unblock");
    json_object_object_add(json, "action", action_obj);
    
    const char* json_str = json_object_to_json_string(json);
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/friends/%s/unblock", manager->server_url, friend_id);
    
    http_request_t* request = friend_create_authenticated_request(manager, url, "PUT");
    http_request_set_body(request, json_str);
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    bool success = false;
    if (response && response->status_code == 200) {
        friend_response_t* friend_response = parse_friend_response(response->body);
        if (friend_response) {
            success = friend_response->success;
            friend_response_free(friend_response);
        }
    }
    
    json_object_put(json);
    http_request_free(request);
    http_response_free(response);
    
    return success;
}

// 好友状态管理
void friend_manager_update_online_status(friend_manager_t* manager, online_status_t status) {
    if (!manager) return;
    
    manager->current_online_status = status;
    
    json_object* json = json_object_new_object();
    json_object* status_obj = json_object_new_string(online_status_to_string(status));
    json_object_object_add(json, "status", status_obj);
    
    const char* json_str = json_object_to_json_string(json);
    
    char* url = friend_build_url(manager, "/api/user/status");
    http_request_t* request = friend_create_authenticated_request(manager, url, "PUT");
    http_request_set_body(request, json_str);
    
    http_response_t* response = http_client_send_request(manager->http_client, request);
    
    json_object_put(json);
    free(url);
    http_request_free(request);
    http_response_free(response);
}

online_status_t friend_manager_get_online_status(const friend_manager_t* manager) {
    return manager ? manager->current_online_status : ONLINE_STATUS_OFFLINE;
}

friend_t* friend_manager_find_friend_by_id(friend_manager_t* manager, const char* friend_id) {
    if (!manager || !friend_id) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->friends_count; i++) {
        if (manager->friends[i].id && strcmp(manager->friends[i].id, friend_id) == 0) {
            pthread_mutex_unlock(&manager->data_mutex);
            return &manager->friends[i];
        }
    }
    pthread_mutex_unlock(&manager->data_mutex);
    
    return NULL;
}

friend_t* friend_manager_find_friend_by_username(friend_manager_t* manager, const char* username) {
    if (!manager || !username) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    for (size_t i = 0; i < manager->friends_count; i++) {
        if (manager->friends[i].username && strcmp(manager->friends[i].username, username) == 0) {
            pthread_mutex_unlock(&manager->data_mutex);
            return &manager->friends[i];
        }
    }
    pthread_mutex_unlock(&manager->data_mutex);
    
    return NULL;
}

// 配置方法
void friend_manager_set_server_url(friend_manager_t* manager, const char* url) {
    if (!manager || !url) return;
    
    free(manager->server_url);
    manager->server_url = strdup(url);
}

void friend_manager_enable_auto_refresh(friend_manager_t* manager, bool enable) {
    if (!manager) return;
    
    manager->auto_refresh_enabled = enable;
    // TODO: 实现自动刷新线程启动/停止逻辑
}

void friend_manager_set_refresh_interval(friend_manager_t* manager, int seconds) {
    if (manager) {
        manager->refresh_interval = seconds;
    }
}

// 内存管理函数
friend_t* friend_create_from_json(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    friend_t* friend = (friend_t*)malloc(sizeof(friend_t));
    memset(friend, 0, sizeof(friend_t));
    
    json_object* id_obj;
    if (json_object_object_get_ex(root, "id", &id_obj)) {
        friend->id = strdup(json_object_get_string(id_obj));
    }
    
    json_object* username_obj;
    if (json_object_object_get_ex(root, "username", &username_obj)) {
        friend->username = strdup(json_object_get_string(username_obj));
    }
    
    json_object* email_obj;
    if (json_object_object_get_ex(root, "email", &email_obj)) {
        friend->email = strdup(json_object_get_string(email_obj));
    }
    
    json_object* avatar_url_obj;
    if (json_object_object_get_ex(root, "avatar_url", &avatar_url_obj)) {
        friend->avatar_url = strdup(json_object_get_string(avatar_url_obj));
    }
    
    json_object* status_obj;
    if (json_object_object_get_ex(root, "status", &status_obj)) {
        friend->status = string_to_friend_status(json_object_get_string(status_obj));
    }
    
    json_object* online_status_obj;
    if (json_object_object_get_ex(root, "online_status", &online_status_obj)) {
        friend->online_status = string_to_online_status(json_object_get_string(online_status_obj));
    }
    
    json_object* last_seen_obj;
    if (json_object_object_get_ex(root, "last_seen", &last_seen_obj)) {
        friend->last_seen = strdup(json_object_get_string(last_seen_obj));
    }
    
    json_object* created_at_obj;
    if (json_object_object_get_ex(root, "created_at", &created_at_obj)) {
        friend->created_at = strdup(json_object_get_string(created_at_obj));
    }
    
    json_object* updated_at_obj;
    if (json_object_object_get_ex(root, "updated_at", &updated_at_obj)) {
        friend->updated_at = strdup(json_object_get_string(updated_at_obj));
    }
    
    json_object_put(root);
    return friend;
}

void friend_free(friend_t* friend) {
    if (!friend) return;
    
    free(friend->id);
    free(friend->username);
    free(friend->email);
    free(friend->avatar_url);
    free(friend->last_seen);
    free(friend->created_at);
    free(friend->updated_at);
    
    memset(friend, 0, sizeof(friend_t));
}

friend_request_t* friend_request_create_from_json(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    friend_request_t* request = (friend_request_t*)malloc(sizeof(friend_request_t));
    memset(request, 0, sizeof(friend_request_t));
    
    json_object* id_obj;
    if (json_object_object_get_ex(root, "id", &id_obj)) {
        request->id = strdup(json_object_get_string(id_obj));
    }
    
    json_object* from_user_id_obj;
    if (json_object_object_get_ex(root, "from_user_id", &from_user_id_obj)) {
        request->from_user_id = strdup(json_object_get_string(from_user_id_obj));
    }
    
    json_object* to_user_id_obj;
    if (json_object_object_get_ex(root, "to_user_id", &to_user_id_obj)) {
        request->to_user_id = strdup(json_object_get_string(to_user_id_obj));
    }
    
    json_object* from_username_obj;
    if (json_object_object_get_ex(root, "from_username", &from_username_obj)) {
        request->from_username = strdup(json_object_get_string(from_username_obj));
    }
    
    json_object* to_username_obj;
    if (json_object_object_get_ex(root, "to_username", &to_username_obj)) {
        request->to_username = strdup(json_object_get_string(to_username_obj));
    }
    
    json_object* message_obj;
    if (json_object_object_get_ex(root, "message", &message_obj)) {
        request->message = strdup(json_object_get_string(message_obj));
    }
    
    json_object* status_obj;
    if (json_object_object_get_ex(root, "status", &status_obj)) {
        request->status = string_to_friend_status(json_object_get_string(status_obj));
    }
    
    json_object* created_at_obj;
    if (json_object_object_get_ex(root, "created_at", &created_at_obj)) {
        request->created_at = strdup(json_object_get_string(created_at_obj));
    }
    
    json_object* updated_at_obj;
    if (json_object_object_get_ex(root, "updated_at", &updated_at_obj)) {
        request->updated_at = strdup(json_object_get_string(updated_at_obj));
    }
    
    json_object_put(root);
    return request;
}

void friend_request_free(friend_request_t* request) {
    if (!request) return;
    
    free(request->id);
    free(request->from_user_id);
    free(request->to_user_id);
    free(request->from_username);
    free(request->to_username);
    free(request->message);
    free(request->created_at);
    free(request->updated_at);
    
    memset(request, 0, sizeof(friend_request_t));
}

void friend_response_free(friend_response_t* response) {
    if (!response) return;
    
    free(response->message);
    
    for (size_t i = 0; i < response->friends_count; i++) {
        friend_free(&response->friends[i]);
    }
    free(response->friends);
    
    for (size_t i = 0; i < response->requests_count; i++) {
        friend_request_free(&response->requests[i]);
    }
    free(response->requests);
    
    free(response);
}

// 辅助函数实现
const char* friend_status_to_string(friend_status_t status) {
    switch (status) {
        case FRIEND_STATUS_PENDING: return "pending";
        case FRIEND_STATUS_ACCEPTED: return "accepted";
        case FRIEND_STATUS_BLOCKED: return "blocked";
        case FRIEND_STATUS_DECLINED: return "declined";
        default: return "unknown";
    }
}

friend_status_t string_to_friend_status(const char* status) {
    if (!status) return FRIEND_STATUS_PENDING;
    
    if (strcmp(status, "pending") == 0) return FRIEND_STATUS_PENDING;
    if (strcmp(status, "accepted") == 0) return FRIEND_STATUS_ACCEPTED;
    if (strcmp(status, "blocked") == 0) return FRIEND_STATUS_BLOCKED;
    if (strcmp(status, "declined") == 0) return FRIEND_STATUS_DECLINED;
    
    return FRIEND_STATUS_PENDING;
}

const char* online_status_to_string(online_status_t status) {
    switch (status) {
        case ONLINE_STATUS_ONLINE: return "online";
        case ONLINE_STATUS_OFFLINE: return "offline";
        case ONLINE_STATUS_AWAY: return "away";
        case ONLINE_STATUS_BUSY: return "busy";
        default: return "offline";
    }
}

online_status_t string_to_online_status(const char* status) {
    if (!status) return ONLINE_STATUS_OFFLINE;
    
    if (strcmp(status, "online") == 0) return ONLINE_STATUS_ONLINE;
    if (strcmp(status, "offline") == 0) return ONLINE_STATUS_OFFLINE;
    if (strcmp(status, "away") == 0) return ONLINE_STATUS_AWAY;
    if (strcmp(status, "busy") == 0) return ONLINE_STATUS_BUSY;
    
    return ONLINE_STATUS_OFFLINE;
}

char* friend_build_url(const friend_manager_t* manager, const char* endpoint) {
    if (!manager || !endpoint) return NULL;
    
    size_t url_len = strlen(manager->server_url) + strlen(endpoint) + 1;
    char* url = (char*)malloc(url_len);
    snprintf(url, url_len, "%s%s", manager->server_url, endpoint);
    
    return url;
}

http_request_t* friend_create_authenticated_request(const friend_manager_t* manager, const char* url, const char* method) {
    if (!manager || !url || !method) return NULL;
    
    http_request_t* request = http_request_create();
    http_request_set_url(request, url);
    http_request_set_method(request, method);
    
    // 添加认证头
    if (g_auth_manager && auth_manager_is_logged_in(g_auth_manager)) {
        const char* token = auth_manager_get_access_token(g_auth_manager);
        if (token) {
            char auth_header[1024];
            snprintf(auth_header, sizeof(auth_header), "Bearer %s", token);
            http_request_add_header(request, "Authorization", auth_header);
        }
    }
    
    http_request_add_header(request, "Content-Type", "application/json");
    
    return request;
}