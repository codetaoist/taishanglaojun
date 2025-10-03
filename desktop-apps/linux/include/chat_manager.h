#ifndef CHAT_MANAGER_H
#define CHAT_MANAGER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <pthread.h>
#include <time.h>
#include <curl/curl.h>
#include <json-c/json.h>
#include "http_client.h"

#ifdef __cplusplus
extern "C" {
#endif

// 消息类型枚举
typedef enum {
    MESSAGE_TYPE_TEXT,
    MESSAGE_TYPE_IMAGE,
    MESSAGE_TYPE_FILE,
    MESSAGE_TYPE_SYSTEM,
    MESSAGE_TYPE_EMOJI
} message_type_t;

// 聊天类型枚举
typedef enum {
    CHAT_TYPE_PRIVATE,
    CHAT_TYPE_GROUP
} chat_type_t;

// 消息状态枚举
typedef enum {
    MESSAGE_STATUS_SENDING,
    MESSAGE_STATUS_SENT,
    MESSAGE_STATUS_DELIVERED,
    MESSAGE_STATUS_READ,
    MESSAGE_STATUS_FAILED
} message_status_t;

// 消息结构体
typedef struct {
    char* id;
    char* chat_id;
    char* sender_id;
    char* sender_username;
    char* content;
    message_type_t type;
    message_status_t status;
    char* timestamp;
    char* created_at;
    char* updated_at;
    
    // 文件消息相关
    char* file_name;
    char* file_url;
    size_t file_size;
    
    // 回复消息相关
    char* reply_to_message_id;
    char* reply_to_content;
} message_t;

// 聊天会话结构体
typedef struct {
    char* id;
    char* name;
    chat_type_t type;
    char* avatar_url;
    char* last_message;
    char* last_message_time;
    int unread_count;
    char** participants;
    size_t participants_count;
    char* created_at;
    char* updated_at;
} chat_t;

// 发送消息请求结构体
typedef struct {
    char* chat_id;
    char* content;
    message_type_t type;
    char* reply_to_message_id;
} send_message_request_t;

// 创建聊天请求结构体
typedef struct {
    chat_type_t type;
    char* name;
    char** participants;
    size_t participants_count;
} create_chat_request_t;

// 聊天响应结构体
typedef struct {
    bool success;
    char* message;
    chat_t* chats;
    size_t chats_count;
    message_t* messages;
    size_t messages_count;
    chat_t chat;
    message_t message_data;
} chat_response_t;

// WebSocket消息结构体
typedef struct {
    char* type;
    char* chat_id;
    char* data;
    char* timestamp;
} websocket_message_t;

// 事件回调函数类型
typedef void (*on_chats_updated_callback_t)(const chat_t* chats, size_t count);
typedef void (*on_messages_updated_callback_t)(const message_t* messages, size_t count);
typedef void (*on_new_message_callback_t)(const message_t* message);
typedef void (*on_message_status_updated_callback_t)(const message_t* message);
typedef void (*on_typing_status_callback_t)(const char* chat_id, const char* user_id, bool is_typing);
typedef void (*on_error_callback_t)(const char* error_message);

// 聊天管理器结构体
typedef struct {
    http_client_t* http_client;
    
    // 配置
    char* server_url;
    char* websocket_url;
    bool auto_reconnect_enabled;
    int reconnect_interval;
    
    // 数据存储
    chat_t* chats;
    size_t chats_count;
    message_t** chat_messages; // 每个聊天的消息数组
    size_t* chat_message_counts; // 每个聊天的消息数量
    size_t chat_messages_capacity;
    pthread_mutex_t data_mutex;
    
    // WebSocket相关
    CURL* websocket_handle;
    pthread_t websocket_thread;
    bool websocket_connected;
    bool should_stop_websocket;
    
    // 自动重连
    pthread_t reconnect_thread;
    bool should_stop_reconnect;
    
    // 事件回调
    on_chats_updated_callback_t on_chats_updated;
    on_messages_updated_callback_t on_messages_updated;
    on_new_message_callback_t on_new_message;
    on_message_status_updated_callback_t on_message_status_updated;
    on_typing_status_callback_t on_typing_status;
    on_error_callback_t on_error;
    
    // 状态
    bool initialized;
} chat_manager_t;

// 聊天管理器创建和销毁
chat_manager_t* chat_manager_create(void);
void chat_manager_destroy(chat_manager_t* manager);

// 初始化和清理
bool chat_manager_initialize(chat_manager_t* manager);
void chat_manager_cleanup(chat_manager_t* manager);

// 聊天列表管理
bool chat_manager_get_chat_list(chat_manager_t* manager);
bool chat_manager_get_chat_list_async(chat_manager_t* manager);

// 消息管理
bool chat_manager_get_messages(chat_manager_t* manager, const char* chat_id, int page, int limit);
bool chat_manager_get_messages_async(chat_manager_t* manager, const char* chat_id, int page, int limit);
bool chat_manager_send_message(chat_manager_t* manager, const send_message_request_t* request);
bool chat_manager_send_message_async(chat_manager_t* manager, const send_message_request_t* request);
bool chat_manager_mark_message_as_read(chat_manager_t* manager, const char* message_id);
bool chat_manager_mark_chat_as_read(chat_manager_t* manager, const char* chat_id);

// 聊天会话管理
bool chat_manager_create_chat(chat_manager_t* manager, const create_chat_request_t* request);
bool chat_manager_create_chat_async(chat_manager_t* manager, const create_chat_request_t* request);
bool chat_manager_delete_chat(chat_manager_t* manager, const char* chat_id);
bool chat_manager_leave_chat(chat_manager_t* manager, const char* chat_id);
bool chat_manager_add_participant(chat_manager_t* manager, const char* chat_id, const char* user_id);
bool chat_manager_remove_participant(chat_manager_t* manager, const char* chat_id, const char* user_id);

// 实时功能
bool chat_manager_connect_websocket(chat_manager_t* manager);
void chat_manager_disconnect_websocket(chat_manager_t* manager);
bool chat_manager_is_websocket_connected(const chat_manager_t* manager);
bool chat_manager_send_typing_status(chat_manager_t* manager, const char* chat_id, bool is_typing);

// 文件传输
bool chat_manager_send_file(chat_manager_t* manager, const char* chat_id, const char* file_path);
bool chat_manager_download_file(chat_manager_t* manager, const char* file_url, const char* save_path);

// 搜索功能
bool chat_manager_search_messages(chat_manager_t* manager, const char* query, const char* chat_id);
bool chat_manager_search_chats(chat_manager_t* manager, const char* query);

// 本地数据管理
chat_t* chat_manager_find_chat_by_id(chat_manager_t* manager, const char* chat_id);
chat_t* chat_manager_find_chat_by_participant(chat_manager_t* manager, const char* user_id);
message_t* chat_manager_find_message_by_id(chat_manager_t* manager, const char* message_id);
message_t* chat_manager_get_chat_messages(chat_manager_t* manager, const char* chat_id, size_t* count);

// 事件回调设置
void chat_manager_set_on_chats_updated_callback(chat_manager_t* manager, on_chats_updated_callback_t callback);
void chat_manager_set_on_messages_updated_callback(chat_manager_t* manager, on_messages_updated_callback_t callback);
void chat_manager_set_on_new_message_callback(chat_manager_t* manager, on_new_message_callback_t callback);
void chat_manager_set_on_message_status_updated_callback(chat_manager_t* manager, on_message_status_updated_callback_t callback);
void chat_manager_set_on_typing_status_callback(chat_manager_t* manager, on_typing_status_callback_t callback);
void chat_manager_set_on_error_callback(chat_manager_t* manager, on_error_callback_t callback);

// 配置方法
void chat_manager_set_server_url(chat_manager_t* manager, const char* url);
void chat_manager_set_websocket_url(chat_manager_t* manager, const char* url);
void chat_manager_enable_auto_reconnect(chat_manager_t* manager, bool enable);
void chat_manager_set_reconnect_interval(chat_manager_t* manager, int seconds);

// 状态查询
bool chat_manager_is_initialized(const chat_manager_t* manager);
int chat_manager_get_unread_message_count(const chat_manager_t* manager);
int chat_manager_get_chat_count(const chat_manager_t* manager);

// 辅助函数
const char* message_type_to_string(message_type_t type);
message_type_t string_to_message_type(const char* type);
const char* chat_type_to_string(chat_type_t type);
chat_type_t string_to_chat_type(const char* type);
const char* message_status_to_string(message_status_t status);
message_status_t string_to_message_status(const char* status);

// 内存管理函数
message_t* message_create_from_json(const char* json_str);
void message_free(message_t* message);
chat_t* chat_create_from_json(const char* json_str);
void chat_free(chat_t* chat);
send_message_request_t* send_message_request_create(const char* chat_id, const char* content, message_type_t type);
void send_message_request_free(send_message_request_t* request);
create_chat_request_t* create_chat_request_create(chat_type_t type, const char* name, const char** participants, size_t participants_count);
void create_chat_request_free(create_chat_request_t* request);
chat_response_t* chat_response_create_from_json(const char* json_str);
void chat_response_free(chat_response_t* response);
websocket_message_t* websocket_message_create_from_json(const char* json_str);
void websocket_message_free(websocket_message_t* ws_message);

// 全局实例管理
extern chat_manager_t* g_chat_manager;

// 全局函数
bool chat_manager_init(void);
void chat_manager_cleanup_global(void);
chat_manager_t* chat_manager_get_instance(void);

#ifdef __cplusplus
}
#endif

#endif // CHAT_MANAGER_H