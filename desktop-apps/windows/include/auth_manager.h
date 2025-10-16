#pragma once

#include "http_client.h"
#include <string>
#include <functional>
#include <memory>

namespace TaishangLaojun {

struct User {
    std::string id;
    std::string username;
    std::string email;
    std::string avatar_url;
    std::string created_at;
    std::string updated_at;
};

struct LoginRequest {
    std::string username;
    std::string password;
};

struct RegisterRequest {
    std::string username;
    std::string email;
    std::string password;
    std::string confirm_password;
};

struct AuthResponse {
    bool success;
    std::string message;
    std::string access_token;
    std::string refresh_token;
    User user;
    int expires_in; // token过期时间（秒）
};

class AuthManager {
public:
    AuthManager();
    ~AuthManager();

    // 同步认证方法
    AuthResponse login(const LoginRequest& request);
    AuthResponse registerUser(const RegisterRequest& request);
    bool logout();
    bool refreshToken();
    
    // 异步认证方法
    void loginAsync(const LoginRequest& request, std::function<void(const AuthResponse&)> callback);
    void registerAsync(const RegisterRequest& request, std::function<void(const AuthResponse&)> callback);
    void logoutAsync(std::function<void(bool)> callback);
    void refreshTokenAsync(std::function<void(bool)> callback);

    // Token管理
    bool isLoggedIn() const;
    std::string getAccessToken() const;
    std::string getRefreshToken() const;
    User getCurrentUser() const;
    
    // 设置认证服务器地址
    void setAuthServerUrl(const std::string& url);
    
    // 自动刷新token
    void enableAutoRefresh(bool enable);
    
    // 清除认证信息
    void clearAuthData();

private:
    std::unique_ptr<HttpClient> httpClient;
    std::string authServerUrl;
    std::string accessToken;
    std::string refreshToken;
    User currentUser;
    bool loggedIn;
    bool autoRefreshEnabled;
    
    // 内部方法
    void saveAuthData(const AuthResponse& response);
    void loadAuthData();
    bool validateToken(const std::string& token);
    std::string buildAuthUrl(const std::string& endpoint);
    void scheduleTokenRefresh();
    
    // JSON处理
    std::string serializeLoginRequest(const LoginRequest& request);
    std::string serializeRegisterRequest(const RegisterRequest& request);
    AuthResponse parseAuthResponse(const HttpResponse& response);
};

// 全局认证管理器实例
extern std::unique_ptr<AuthManager> g_authManager;

// 初始化和清理函数
bool initAuthManager();
void cleanupAuthManager();

} // namespace TaishangLaojun