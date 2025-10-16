#ifndef FRIEND_MANAGER_H
#define FRIEND_MANAGER_H

#include "http_client.h"
#include "auth_manager.h"
#include <stdbool.h>
#include <stddef.h>
#include <pthread.h>

#ifdef __cplusplus
extern "C" {
#endif

// 好友状态枚举
typedef enum {
    FRIEND_STATUS_PENDING,    // 待确认
    FRIEND_STATUS_ACCEPTED,   // 已接受
    FRIEND_STATUS_BLOCKED,    // 已屏蔽
    FRIEND_STATUS_DECLINED    // 已拒绝
} friend_status_t;

// 在线状态枚举
typedef enum {
    ONLINE_STATUS_ONLINE,     // 在线
    ONLINE_STATUS_OFFLINE,    // 离线
    ONLINE_STATUS_AWAY,       // 离开
    ONLINE_STATUS_BUSY        // 忙碌
} online_status_t;

// 好友信息结构
typedef struct {
    char* id;
    char* username;
    char* email;
    char* avatar_url;
    friend_status_t status;
    online_status_t online_status;
    char* last_seen;
    char* created_at;
    char* updated_at;
} friend_t;

// 好友请求结构
typedef struct {
    char* id;
    char* from_user_id;
    char* to_user_id;
    char* from_username;
    char* to_username;
    char* message;
    friend_status_t status;
    char* created_at;
    char* updated_at;
} friend_request_t;

// 添加好友请求结构
typedef struct {
    char* username;
    char* message;
} add_friend_request_t;

// 好友响应结构
typedef struct {
    bool success;
    char* message;
    friend_t* friends;
    size_t friends_count;
    friend_request_t* requests;
    size_t requests_count;
} friend_response_t;

// 好友管理器结构
typedef struct {
    http_client_t* http_client;
    char* server_url;
    friend_t* friends;
    size_t friends_count;
    friend_request_t* pending_requests;
    size_t pending_requests_count;
    online_status_t current_online_status;
    
    // 配置
    bool auto_refresh_enabled;
    int refresh_interval;
    bool is_running;
    
    // 线程相关
    pthread_t refresh_thread;
    pthread_mutex_t data_mutex;
} friend_manager_t;

// 回调函数类型
typedef void (*friend_list_callback_t)(const friend_response_t* response, void* user_data);
typedef void (*friend_request_callback_t)(const friend_response_t* response, void* user_data);
typedef void (*add_friend_callback_t)(bool success, const char* message, void* user_data);
typedef void (*respond_friend_callback_t)(bool success, const char* message, void* user_data);
typedef void (*remove_friend_callback_t)(bool success, const char* message, void* user_data);
typedef void (*friend_status_changed_callback_t)(const friend_t* friend, void* user_data);

// 全局好友管理器实例
extern friend_manager_t* g_friend_manager;

// 初始化和清理函数
bool friend_manager_init(void);
void friend_manager_cleanup(void);

// 好友管理器创建和销毁
friend_manager_t* friend_manager_create(void);
void friend_manager_destroy(friend_manager_t* manager);

// 同步方法
friend_response_t* friend_manager_get_friend_list(friend_manager_t* manager);
friend_response_t* friend_manager_get_friend_requests(friend_manager_t* manager);
bool friend_manager_add_friend(friend_manager_t* manager, const char* username, const char* message);
bool friend_manager_respond_to_request(friend_manager_t* manager, const char* request_id, bool accept);
bool friend_manager_remove_friend(friend_manager_t* manager, const char* friend_id);
bool friend_manager_block_friend(friend_manager_t* manager, const char* friend_id);
bool friend_manager_unblock_friend(friend_manager_t* manager, const char* friend_id);

// 异步方法
bool friend_manager_get_friend_list_async(friend_manager_t* manager, friend_list_callback_t callback, void* user_data);
bool friend_manager_get_friend_requests_async(friend_manager_t* manager, friend_request_callback_t callback, void* user_data);
bool friend_manager_add_friend_async(friend_manager_t* manager, const char* username, const char* message, 
                                    add_friend_callback_t callback, void* user_data);
bool friend_manager_respond_to_request_async(friend_manager_t* manager, const char* request_id, bool accept,
                                           respond_friend_callback_t callback, void* user_data);
bool friend_manager_remove_friend_async(friend_manager_t* manager, const char* friend_id,
                                       remove_friend_callback_t callback, void* user_data);
bool friend_manager_block_friend_async(friend_manager_t* manager, const char* friend_id,
                                      respond_friend_callback_t callback, void* user_data);
bool friend_manager_unblock_friend_async(friend_manager_t* manager, const char* friend_id,
                                        respond_friend_callback_t callback, void* user_data);

// 好友状态管理
void friend_manager_update_online_status(friend_manager_t* manager, online_status_t status);
online_status_t friend_manager_get_online_status(const friend_manager_t* manager);
friend_t* friend_manager_find_friend_by_id(friend_manager_t* manager, const char* friend_id);
friend_t* friend_manager_find_friend_by_username(friend_manager_t* manager, const char* username);

// 配置方法
void friend_manager_set_server_url(friend_manager_t* manager, const char* url);
void friend_manager_enable_auto_refresh(friend_manager_t* manager, bool enable);
void friend_manager_set_refresh_interval(friend_manager_t* manager, int seconds);

// 事件回调设置
void friend_manager_set_on_friend_list_updated(friend_manager_t* manager, friend_list_callback_t callback, void* user_data);
void friend_manager_set_on_friend_request_received(friend_manager_t* manager, friend_request_callback_t callback, void* user_data);
void friend_manager_set_on_friend_status_changed(friend_manager_t* manager, friend_status_changed_callback_t callback, void* user_data);

// 内存管理
friend_t* friend_create_from_json(const char* json_str);
void friend_free(friend_t* friend);
friend_request_t* friend_request_create_from_json(const char* json_str);
void friend_request_free(friend_request_t* request);
add_friend_request_t* add_friend_request_create(const char* username, const char* message);
void add_friend_request_free(add_friend_request_t* request);
friend_response_t* friend_response_create_from_json(const char* json_str);
void friend_response_free(friend_response_t* response);

// 辅助函数
const char* friend_status_to_string(friend_status_t status);
friend_status_t string_to_friend_status(const char* status);
const char* online_status_to_string(online_status_t status);
online_status_t string_to_online_status(const char* status);

char* friend_build_url(const friend_manager_t* manager, const char* endpoint);
http_request_t* friend_create_authenticated_request(const friend_manager_t* manager, const char* url, const char* method);

#ifdef __cplusplus
}
#endif

#endif // FRIEND_MANAGER_H