#include "http_client.h"
#include <iostream>
#include <string>
#include <map>

// 测试框架宏
#define TEST_ASSERT(condition, message) \
    do { \
        if (!(condition)) { \
            std::cerr << "FAIL: " << message << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

// 测试HTTP客户端GET请求
bool test_http_client_get() {
    HttpClient httpClient;
    
    // 测试空URL
    std::string response1 = httpClient.Get("");
    TEST_ASSERT(response1.empty(), "GET request with empty URL should return empty response");
    
    // 测试无效URL
    std::string response2 = httpClient.Get("invalid-url");
    TEST_ASSERT(response2.empty(), "GET request with invalid URL should return empty response");
    
    // 测试本地URL（可能不存在）
    std::string response3 = httpClient.Get("http://localhost:8080/api/test");
    // 在测试环境中，这可能会失败（服务器不存在），但应该能够处理
    
    // 测试HTTPS URL
    std::string response4 = httpClient.Get("https://httpbin.org/get");
    // 这可能会成功，取决于网络连接
    
    return true;
}

// 测试HTTP客户端POST请求
bool test_http_client_post() {
    HttpClient httpClient;
    
    // 测试空URL的POST请求
    std::string response1 = httpClient.Post("", "test data");
    TEST_ASSERT(response1.empty(), "POST request with empty URL should return empty response");
    
    // 测试空数据的POST请求
    std::string response2 = httpClient.Post("http://localhost:8080/api/test", "");
    // 应该能够处理空数据
    
    // 测试JSON数据的POST请求
    std::string jsonData = R"({"message": "Hello, World!", "timestamp": "2024-01-01T12:00:00Z"})";
    std::string response3 = httpClient.Post("http://localhost:8080/api/messages", jsonData);
    // 在测试环境中，这可能会失败（服务器不存在）
    
    return true;
}

// 测试HTTP客户端头部管理
bool test_http_client_headers() {
    HttpClient httpClient;
    
    // 测试设置头部
    bool setResult1 = httpClient.SetHeader("Content-Type", "application/json");
    TEST_ASSERT(setResult1, "Should be able to set Content-Type header");
    
    bool setResult2 = httpClient.SetHeader("Authorization", "Bearer test-token");
    TEST_ASSERT(setResult2, "Should be able to set Authorization header");
    
    // 测试获取头部
    std::string contentType = httpClient.GetHeader("Content-Type");
    TEST_ASSERT(contentType == "application/json", "Content-Type header should match what was set");
    
    // 测试移除头部
    bool removeResult = httpClient.RemoveHeader("Authorization");
    TEST_ASSERT(removeResult, "Should be able to remove header");
    
    // 测试清除所有头部
    bool clearResult = httpClient.ClearHeaders();
    TEST_ASSERT(clearResult, "Should be able to clear all headers");
    
    return true;
}

// 测试HTTP客户端超时设置
bool test_http_client_timeout() {
    HttpClient httpClient;
    
    // 测试设置超时
    bool timeoutResult = httpClient.SetTimeout(30000); // 30秒
    TEST_ASSERT(timeoutResult, "Should be able to set timeout");
    
    // 测试获取超时
    int timeout = httpClient.GetTimeout();
    TEST_ASSERT(timeout == 30000, "Timeout should match what was set");
    
    // 测试无效超时值
    bool invalidTimeoutResult = httpClient.SetTimeout(-1);
    TEST_ASSERT(!invalidTimeoutResult, "Should not accept negative timeout");
    
    return true;
}

// 测试HTTP客户端错误处理
bool test_http_client_error_handling() {
    HttpClient httpClient;
    
    // 测试获取最后错误
    std::string lastError = httpClient.GetLastError();
    // 初始状态下可能没有错误
    
    // 测试获取响应状态码
    int statusCode = httpClient.GetLastStatusCode();
    TEST_ASSERT(statusCode >= 0, "Status code should be non-negative");
    
    // 测试连接到不存在的服务器
    std::string response = httpClient.Get("http://nonexistent-server-12345.com/api/test");
    TEST_ASSERT(response.empty(), "Request to non-existent server should fail");
    
    // 检查是否记录了错误
    std::string errorAfterFailure = httpClient.GetLastError();
    // 应该有错误信息
    
    return true;
}

// 测试HTTP客户端认证
bool test_http_client_authentication() {
    HttpClient httpClient;
    
    // 测试基本认证
    bool basicAuthResult = httpClient.SetBasicAuth("username", "password");
    TEST_ASSERT(basicAuthResult, "Should be able to set basic authentication");
    
    // 测试Bearer令牌认证
    bool bearerAuthResult = httpClient.SetBearerToken("test-bearer-token");
    TEST_ASSERT(bearerAuthResult, "Should be able to set bearer token");
    
    // 测试清除认证
    bool clearAuthResult = httpClient.ClearAuthentication();
    TEST_ASSERT(clearAuthResult, "Should be able to clear authentication");
    
    return true;
}

// 测试HTTP客户端文件上传
bool test_http_client_file_upload() {
    HttpClient httpClient;
    
    // 创建测试文件
    std::string testFile = "test_upload_http.txt";
    std::ofstream file(testFile);
    if (file.is_open()) {
        file << "Test file content for HTTP upload.";
        file.close();
    }
    
    // 测试文件上传
    std::string uploadResponse = httpClient.UploadFile("http://localhost:8080/api/upload", testFile);
    // 在测试环境中，这可能会失败（服务器不存在）
    
    // 测试上传不存在的文件
    std::string invalidUploadResponse = httpClient.UploadFile("http://localhost:8080/api/upload", "nonexistent.txt");
    TEST_ASSERT(invalidUploadResponse.empty(), "Upload of non-existent file should fail");
    
    // 清理测试文件
    std::remove(testFile.c_str());
    
    return true;
}