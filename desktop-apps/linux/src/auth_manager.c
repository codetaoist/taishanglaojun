#include "auth_manager.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <pwd.h>
#include <json-c/json.h>

// 全局认证管理器实例
auth_manager_t* g_auth_manager = NULL;

// 初始化和清理函数
bool auth_manager_init(void) {
    if (g_auth_manager != NULL) {
        return true; // 已经初始化
    }
    
    g_auth_manager = auth_manager_create();
    return g_auth_manager != NULL;
}

void auth_manager_cleanup(void) {
    if (g_auth_manager != NULL) {
        auth_manager_destroy(g_auth_manager);
        g_auth_manager = NULL;
    }
}

// 认证管理器创建和销毁
auth_manager_t* auth_manager_create(void) {
    auth_manager_t* manager = (auth_manager_t*)malloc(sizeof(auth_manager_t));
    if (!manager) return NULL;
    
    memset(manager, 0, sizeof(auth_manager_t));
    
    manager->http_client = http_client_create();
    if (!manager->http_client) {
        free(manager);
        return NULL;
    }
    
    manager->auth_server_url = strdup("http://localhost:8082");
    manager->logged_in = false;
    manager->auto_refresh_enabled = true;
    
    return manager;
}

void auth_manager_destroy(auth_manager_t* manager) {
    if (!manager) return;
    
    if (manager->http_client) {
        http_client_destroy(manager->http_client);
    }
    
    free(manager->auth_server_url);
    free(manager->access_token);
    free(manager->refresh_token);
    user_free(&manager->current_user);
    
    free(manager);
}

// JSON辅助函数
static json_object* create_login_json(const login_request_t* request) {
    json_object* json = json_object_new_object();
    json_object* username = json_object_new_string(request->username);
    json_object* password = json_object_new_string(request->password);
    
    json_object_object_add(json, "username", username);
    json_object_object_add(json, "password", password);
    
    return json;
}

static json_object* create_register_json(const register_request_t* request) {
    json_object* json = json_object_new_object();
    json_object* username = json_object_new_string(request->username);
    json_object* email = json_object_new_string(request->email);
    json_object* password = json_object_new_string(request->password);
    json_object* confirm_password = json_object_new_string(request->confirm_password);
    
    json_object_object_add(json, "username", username);
    json_object_object_add(json, "email", email);
    json_object_object_add(json, "password", password);
    json_object_object_add(json, "confirm_password", confirm_password);
    
    return json;
}

static auth_response_t* parse_auth_response(const char* json_str) {
    if (!json_str) return NULL;
    
    json_object* root = json_tokener_parse(json_str);
    if (!root) return NULL;
    
    auth_response_t* response = (auth_response_t*)malloc(sizeof(auth_response_t));
    memset(response, 0, sizeof(auth_response_t));
    
    json_object* success_obj;
    if (json_object_object_get_ex(root, "success", &success_obj)) {
        response->success = json_object_get_boolean(success_obj);
    }
    
    json_object* message_obj;
    if (json_object_object_get_ex(root, "message", &message_obj)) {
        const char* message = json_object_get_string(message_obj);
        response->message = strdup(message);
    }
    
    json_object* access_token_obj;
    if (json_object_object_get_ex(root, "access_token", &access_token_obj)) {
        const char* token = json_object_get_string(access_token_obj);
        response->access_token = strdup(token);
    }
    
    json_object* refresh_token_obj;
    if (json_object_object_get_ex(root, "refresh_token", &refresh_token_obj)) {
        const char* token = json_object_get_string(refresh_token_obj);
        response->refresh_token = strdup(token);
    }
    
    json_object* expires_in_obj;
    if (json_object_object_get_ex(root, "expires_in", &expires_in_obj)) {
        response->expires_in = json_object_get_int(expires_in_obj);
    }
    
    json_object* user_obj;
    if (json_object_object_get_ex(root, "user", &user_obj)) {
        json_object* id_obj;
        if (json_object_object_get_ex(user_obj, "id", &id_obj)) {
            response->user.id = strdup(json_object_get_string(id_obj));
        }
        
        json_object* username_obj;
        if (json_object_object_get_ex(user_obj, "username", &username_obj)) {
            response->user.username = strdup(json_object_get_string(username_obj));
        }
        
        json_object* email_obj;
        if (json_object_object_get_ex(user_obj, "email", &email_obj)) {
            response->user.email = strdup(json_object_get_string(email_obj));
        }
        
        json_object* avatar_url_obj;
        if (json_object_object_get_ex(user_obj, "avatar_url", &avatar_url_obj)) {
            response->user.avatar_url = strdup(json_object_get_string(avatar_url_obj));
        }
        
        json_object* created_at_obj;
        if (json_object_object_get_ex(user_obj, "created_at", &created_at_obj)) {
            response->user.created_at = strdup(json_object_get_string(created_at_obj));
        }
        
        json_object* updated_at_obj;
        if (json_object_object_get_ex(user_obj, "updated_at", &updated_at_obj)) {
            response->user.updated_at = strdup(json_object_get_string(updated_at_obj));
        }
    }
    
    json_object_put(root);
    return response;
}

// 同步认证方法
auth_response_t* auth_manager_login(auth_manager_t* manager, const login_request_t* request) {
    if (!manager || !request) return NULL;
    
    json_object* json = create_login_json(request);
    const char* json_str = json_object_to_json_string(json);
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/auth/login", manager->auth_server_url);
    
    http_request_t* http_req = http_request_create();
    http_request_set_url(http_req, url);
    http_request_set_method(http_req, "POST");
    http_request_add_header(http_req, "Content-Type", "application/json");
    http_request_set_body(http_req, json_str);
    
    http_response_t* http_resp = http_client_send_request(manager->http_client, http_req);
    
    auth_response_t* response = NULL;
    if (http_resp && http_resp->status_code == 200) {
        response = parse_auth_response(http_resp->body);
        
        if (response && response->success) {
            // 保存认证信息
            free(manager->access_token);
            free(manager->refresh_token);
            user_free(&manager->current_user);
            
            manager->access_token = strdup(response->access_token);
            manager->refresh_token = strdup(response->refresh_token);
            manager->current_user = response->user;
            manager->logged_in = true;
            
            // 设置默认Authorization头
            char auth_header[1024];
            snprintf(auth_header, sizeof(auth_header), "Bearer %s", manager->access_token);
            http_client_set_default_header(manager->http_client, "Authorization", auth_header);
        }
    }
    
    json_object_put(json);
    http_request_free(http_req);
    http_response_free(http_resp);
    
    return response;
}

auth_response_t* auth_manager_register(auth_manager_t* manager, const register_request_t* request) {
    if (!manager || !request) return NULL;
    
    json_object* json = create_register_json(request);
    const char* json_str = json_object_to_json_string(json);
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/auth/register", manager->auth_server_url);
    
    http_request_t* http_req = http_request_create();
    http_request_set_url(http_req, url);
    http_request_set_method(http_req, "POST");
    http_request_add_header(http_req, "Content-Type", "application/json");
    http_request_set_body(http_req, json_str);
    
    http_response_t* http_resp = http_client_send_request(manager->http_client, http_req);
    
    auth_response_t* response = NULL;
    if (http_resp && http_resp->status_code == 201) {
        response = parse_auth_response(http_resp->body);
    }
    
    json_object_put(json);
    http_request_free(http_req);
    http_response_free(http_resp);
    
    return response;
}

bool auth_manager_logout(auth_manager_t* manager) {
    if (!manager || !manager->logged_in) return false;
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/auth/logout", manager->auth_server_url);
    
    http_request_t* http_req = http_request_create();
    http_request_set_url(http_req, url);
    http_request_set_method(http_req, "POST");
    
    http_response_t* http_resp = http_client_send_request(manager->http_client, http_req);
    
    bool success = (http_resp && http_resp->status_code == 200);
    
    // 清除本地认证信息
    auth_manager_clear_auth_data(manager);
    
    http_request_free(http_req);
    http_response_free(http_resp);
    
    return success;
}

bool auth_manager_refresh_token(auth_manager_t* manager) {
    if (!manager || !manager->refresh_token) return false;
    
    char url[512];
    snprintf(url, sizeof(url), "%s/api/auth/refresh", manager->auth_server_url);
    
    json_object* json = json_object_new_object();
    json_object* refresh_token = json_object_new_string(manager->refresh_token);
    json_object_object_add(json, "refresh_token", refresh_token);
    
    const char* json_str = json_object_to_json_string(json);
    
    http_request_t* http_req = http_request_create();
    http_request_set_url(http_req, url);
    http_request_set_method(http_req, "POST");
    http_request_add_header(http_req, "Content-Type", "application/json");
    http_request_set_body(http_req, json_str);
    
    http_response_t* http_resp = http_client_send_request(manager->http_client, http_req);
    
    bool success = false;
    if (http_resp && http_resp->status_code == 200) {
        auth_response_t* response = parse_auth_response(http_resp->body);
        if (response && response->success) {
            free(manager->access_token);
            manager->access_token = strdup(response->access_token);
            
            // 更新Authorization头
            char auth_header[1024];
            snprintf(auth_header, sizeof(auth_header), "Bearer %s", manager->access_token);
            http_client_set_default_header(manager->http_client, "Authorization", auth_header);
            
            success = true;
        }
        auth_response_free(response);
    }
    
    json_object_put(json);
    http_request_free(http_req);
    http_response_free(http_resp);
    
    return success;
}

// Token管理
bool auth_manager_is_logged_in(const auth_manager_t* manager) {
    return manager && manager->logged_in && manager->access_token;
}

const char* auth_manager_get_access_token(const auth_manager_t* manager) {
    return manager ? manager->access_token : NULL;
}

const char* auth_manager_get_refresh_token(const auth_manager_t* manager) {
    return manager ? manager->refresh_token : NULL;
}

user_t auth_manager_get_current_user(const auth_manager_t* manager) {
    user_t empty_user = {0};
    return manager ? manager->current_user : empty_user;
}

// 配置方法
void auth_manager_set_server_url(auth_manager_t* manager, const char* url) {
    if (!manager || !url) return;
    
    free(manager->auth_server_url);
    manager->auth_server_url = strdup(url);
}

void auth_manager_enable_auto_refresh(auth_manager_t* manager, bool enable) {
    if (manager) {
        manager->auto_refresh_enabled = enable;
    }
}

void auth_manager_clear_auth_data(auth_manager_t* manager) {
    if (!manager) return;
    
    free(manager->access_token);
    free(manager->refresh_token);
    user_free(&manager->current_user);
    
    manager->access_token = NULL;
    manager->refresh_token = NULL;
    memset(&manager->current_user, 0, sizeof(user_t));
    manager->logged_in = false;
    
    // 移除Authorization头
    http_client_remove_default_header(manager->http_client, "Authorization");
}

// 请求和响应处理
login_request_t* login_request_create(const char* username, const char* password) {
    if (!username || !password) return NULL;
    
    login_request_t* request = (login_request_t*)malloc(sizeof(login_request_t));
    request->username = strdup(username);
    request->password = strdup(password);
    
    return request;
}

void login_request_free(login_request_t* request) {
    if (!request) return;
    
    free(request->username);
    free(request->password);
    free(request);
}

register_request_t* register_request_create(const char* username, const char* email, 
                                           const char* password, const char* confirm_password) {
    if (!username || !email || !password || !confirm_password) return NULL;
    
    register_request_t* request = (register_request_t*)malloc(sizeof(register_request_t));
    request->username = strdup(username);
    request->email = strdup(email);
    request->password = strdup(password);
    request->confirm_password = strdup(confirm_password);
    
    return request;
}

void register_request_free(register_request_t* request) {
    if (!request) return;
    
    free(request->username);
    free(request->email);
    free(request->password);
    free(request->confirm_password);
    free(request);
}

void auth_response_free(auth_response_t* response) {
    if (!response) return;
    
    free(response->message);
    free(response->access_token);
    free(response->refresh_token);
    user_free(&response->user);
    free(response);
}

void user_free(user_t* user) {
    if (!user) return;
    
    free(user->id);
    free(user->username);
    free(user->email);
    free(user->avatar_url);
    free(user->created_at);
    free(user->updated_at);
    
    memset(user, 0, sizeof(user_t));
}

// 辅助函数
char* auth_get_config_dir(void) {
    struct passwd* pw = getpwuid(getuid());
    if (!pw) return NULL;
    
    char* config_dir = (char*)malloc(512);
    snprintf(config_dir, 512, "%s/.config/taishanglaojun", pw->pw_dir);
    
    // 创建目录如果不存在
    mkdir(config_dir, 0755);
    
    return config_dir;
}

bool auth_save_to_file(const char* filename, const char* data) {
    if (!filename || !data) return false;
    
    char* config_dir = auth_get_config_dir();
    if (!config_dir) return false;
    
    char filepath[1024];
    snprintf(filepath, sizeof(filepath), "%s/%s", config_dir, filename);
    
    FILE* file = fopen(filepath, "w");
    if (!file) {
        free(config_dir);
        return false;
    }
    
    fprintf(file, "%s", data);
    fclose(file);
    free(config_dir);
    
    return true;
}

char* auth_load_from_file(const char* filename) {
    if (!filename) return NULL;
    
    char* config_dir = auth_get_config_dir();
    if (!config_dir) return NULL;
    
    char filepath[1024];
    snprintf(filepath, sizeof(filepath), "%s/%s", config_dir, filename);
    
    FILE* file = fopen(filepath, "r");
    if (!file) {
        free(config_dir);
        return NULL;
    }
    
    fseek(file, 0, SEEK_END);
    long length = ftell(file);
    fseek(file, 0, SEEK_SET);
    
    char* data = (char*)malloc(length + 1);
    fread(data, 1, length, file);
    data[length] = '\0';
    
    fclose(file);
    free(config_dir);
    
    return data;
}