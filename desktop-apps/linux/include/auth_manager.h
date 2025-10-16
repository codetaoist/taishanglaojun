#ifndef AUTH_MANAGER_H
#define AUTH_MANAGER_H

#include "http_client.h"
#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// 用户结构
typedef struct {
    char* id;
    char* username;
    char* email;
    char* avatar_url;
    char* created_at;
    char* updated_at;
} user_t;

// 登录请求结构
typedef struct {
    char* username;
    char* password;
} login_request_t;

// 注册请求结构
typedef struct {
    char* username;
    char* email;
    char* password;
    char* confirm_password;
} register_request_t;

// 认证响应结构
typedef struct {
    bool success;
    char* message;
    char* access_token;
    char* refresh_token;
    user_t user;
    int expires_in;
} auth_response_t;

// 认证管理器结构
typedef struct {
    http_client_t* http_client;
    char* auth_server_url;
    char* access_token;
    char* refresh_token;
    user_t current_user;
    bool logged_in;
    bool auto_refresh_enabled;
} auth_manager_t;

// 回调函数类型
typedef void (*auth_callback_t)(const auth_response_t* response, void* user_data);
typedef void (*logout_callback_t)(bool success, void* user_data);

// 全局认证管理器实例
extern auth_manager_t* g_auth_manager;

// 初始化和清理函数
bool auth_manager_init(void);
void auth_manager_cleanup(void);

// 认证管理器创建和销毁
auth_manager_t* auth_manager_create(void);
void auth_manager_destroy(auth_manager_t* manager);

// 同步认证方法
auth_response_t* auth_manager_login(auth_manager_t* manager, const login_request_t* request);
auth_response_t* auth_manager_register(auth_manager_t* manager, const register_request_t* request);
bool auth_manager_logout(auth_manager_t* manager);
bool auth_manager_refresh_token(auth_manager_t* manager);

// 异步认证方法
bool auth_manager_login_async(auth_manager_t* manager, const login_request_t* request, 
                             auth_callback_t callback, void* user_data);
bool auth_manager_register_async(auth_manager_t* manager, const register_request_t* request, 
                                auth_callback_t callback, void* user_data);
bool auth_manager_logout_async(auth_manager_t* manager, logout_callback_t callback, void* user_data);
bool auth_manager_refresh_token_async(auth_manager_t* manager, logout_callback_t callback, void* user_data);

// Token管理
bool auth_manager_is_logged_in(const auth_manager_t* manager);
const char* auth_manager_get_access_token(const auth_manager_t* manager);
const char* auth_manager_get_refresh_token(const auth_manager_t* manager);
user_t auth_manager_get_current_user(const auth_manager_t* manager);

// 配置方法
void auth_manager_set_server_url(auth_manager_t* manager, const char* url);
void auth_manager_enable_auto_refresh(auth_manager_t* manager, bool enable);
void auth_manager_clear_auth_data(auth_manager_t* manager);

// 请求和响应处理
login_request_t* login_request_create(const char* username, const char* password);
void login_request_free(login_request_t* request);

register_request_t* register_request_create(const char* username, const char* email, 
                                           const char* password, const char* confirm_password);
void register_request_free(register_request_t* request);

void auth_response_free(auth_response_t* response);
void user_free(user_t* user);

// 辅助函数
char* auth_get_config_dir(void);
bool auth_save_to_file(const char* filename, const char* data);
char* auth_load_from_file(const char* filename);

#ifdef __cplusplus
}
#endif

#endif // AUTH_MANAGER_H