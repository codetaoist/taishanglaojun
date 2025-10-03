#pragma once

#include <windows.h>
#include <winhttp.h>
#include <string>
#include <map>
#include <functional>
#include <memory>
#include <vector>

#pragma comment(lib, "winhttp.lib")

namespace TaishangLaojun {

struct HttpResponse {
    int status_code;
    std::string body;
    std::map<std::string, std::string> headers;
    bool success;
    std::string error_message;
};

struct HttpRequest {
    std::string method;
    std::string url;
    std::string body;
    std::map<std::string, std::string> headers;
    int timeout_ms = 30000; // 30秒超时
};

class HttpClient {
public:
    HttpClient();
    ~HttpClient();

    // 同步HTTP请求
    HttpResponse request(const HttpRequest& req);
    
    // 异步HTTP请求
    void requestAsync(const HttpRequest& req, std::function<void(const HttpResponse&)> callback);

    // 便捷方法
    HttpResponse get(const std::string& url, const std::map<std::string, std::string>& headers = {});
    HttpResponse post(const std::string& url, const std::string& body, const std::map<std::string, std::string>& headers = {});
    HttpResponse put(const std::string& url, const std::string& body, const std::map<std::string, std::string>& headers = {});
    HttpResponse del(const std::string& url, const std::map<std::string, std::string>& headers = {});

    // 设置默认头部（如认证token）
    void setDefaultHeader(const std::string& key, const std::string& value);
    void removeDefaultHeader(const std::string& key);

    // 设置基础URL
    void setBaseUrl(const std::string& base_url);

private:
    HINTERNET hSession;
    std::map<std::string, std::string> defaultHeaders;
    std::string baseUrl;

    // 内部辅助方法
    HttpResponse performRequest(const HttpRequest& req);
    std::wstring stringToWString(const std::string& str);
    std::string wstringToString(const std::wstring& wstr);
    void parseUrl(const std::string& url, std::wstring& host, std::wstring& path, INTERNET_PORT& port, bool& isHttps);
};

// 全局HTTP客户端实例
extern std::unique_ptr<HttpClient> g_httpClient;

// 初始化和清理函数
bool initHttpClient();
void cleanupHttpClient();

} // namespace TaishangLaojun