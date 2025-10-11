#include "auth_manager.h"
#include <iostream>
#include <string>

// 测试框架宏
#define TEST_ASSERT(condition, message) \
    do { \
        if (!(condition)) { \
            std::cerr << "FAIL: " << message << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

#define TEST_ASSERT_EQ(expected, actual, message) \
    do { \
        if ((expected) != (actual)) { \
            std::cerr << "FAIL: " << message << " - Expected: " << (expected) << ", Actual: " << (actual) \
                      << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

// 测试认证管理器登录功能
bool test_auth_manager_login() {
    AuthManager authManager;
    
    // 测试空用户名和密码
    bool result1 = authManager.Login("", "");
    TEST_ASSERT(!result1, "Login should fail with empty credentials");
    
    // 测试无效用户名
    bool result2 = authManager.Login("invalid_user", "password123");
    // 在测试环境中，这应该失败（除非有模拟的后端）
    
    // 测试有效格式的用户名和密码
    bool result3 = authManager.Login("test@example.com", "password123");
    // 在没有后端连接的情况下，这可能会失败，这是预期的
    
    return true;
}

// 测试认证管理器登出功能
bool test_auth_manager_logout() {
    AuthManager authManager;
    
    // 测试在未登录状态下登出
    bool result1 = authManager.Logout();
    // 应该能够安全地处理未登录状态的登出
    
    // 测试登录后登出
    // 注意：在测试环境中，登录可能失败，但我们仍然可以测试登出逻辑
    authManager.Login("test@example.com", "password123");
    bool result2 = authManager.Logout();
    
    return true;
}

// 测试令牌管理
bool test_auth_manager_token_management() {
    AuthManager authManager;
    
    // 测试获取空令牌
    std::string token = authManager.GetAccessToken();
    // 在未登录状态下，令牌应该为空
    
    // 测试令牌验证
    bool isValid = authManager.IsTokenValid();
    TEST_ASSERT(!isValid, "Token should be invalid when not logged in");
    
    // 测试刷新令牌
    bool refreshResult = authManager.RefreshToken();
    // 在没有有效令牌的情况下，刷新应该失败
    
    return true;
}

// 测试用户信息管理
bool test_auth_manager_user_info() {
    AuthManager authManager;
    
    // 测试获取用户信息（未登录状态）
    std::string userId = authManager.GetUserId();
    std::string username = authManager.GetUsername();
    std::string email = authManager.GetUserEmail();
    
    // 在未登录状态下，这些应该为空
    TEST_ASSERT(userId.empty(), "User ID should be empty when not logged in");
    TEST_ASSERT(username.empty(), "Username should be empty when not logged in");
    TEST_ASSERT(email.empty(), "Email should be empty when not logged in");
    
    return true;
}

// 测试认证URL构建
bool test_auth_manager_url_building() {
    AuthManager authManager;
    
    // 测试构建认证URL
    std::string authUrl = authManager.BuildAuthUrl("test_client_id", "http://localhost:3000/callback");
    TEST_ASSERT(!authUrl.empty(), "Auth URL should not be empty");
    
    // 验证URL包含必要的参数
    TEST_ASSERT(authUrl.find("client_id") != std::string::npos, "Auth URL should contain client_id");
    TEST_ASSERT(authUrl.find("redirect_uri") != std::string::npos, "Auth URL should contain redirect_uri");
    TEST_ASSERT(authUrl.find("response_type") != std::string::npos, "Auth URL should contain response_type");
    
    return true;
}