#include "../include/http_client.h"
#include <iostream>
#include <sstream>
#include <thread>

namespace TaishangLaojun {

std::unique_ptr<HttpClient> g_httpClient = nullptr;

HttpClient::HttpClient() : hSession(nullptr) {
    // 初始化WinHTTP会话
    hSession = WinHttpOpen(L"TaishangLaojun Desktop Client/1.0",
                          WINHTTP_ACCESS_TYPE_DEFAULT_PROXY,
                          WINHTTP_NO_PROXY_NAME,
                          WINHTTP_NO_PROXY_BYPASS,
                          0);
    
    if (!hSession) {
        std::wcerr << L"Failed to initialize WinHTTP session. Error: " << GetLastError() << std::endl;
    }
}

HttpClient::~HttpClient() {
    if (hSession) {
        WinHttpCloseHandle(hSession);
    }
}

HttpResponse HttpClient::request(const HttpRequest& req) {
    return performRequest(req);
}

void HttpClient::requestAsync(const HttpRequest& req, std::function<void(const HttpResponse&)> callback) {
    std::thread([this, req, callback]() {
        HttpResponse response = performRequest(req);
        callback(response);
    }).detach();
}

HttpResponse HttpClient::get(const std::string& url, const std::map<std::string, std::string>& headers) {
    HttpRequest req;
    req.method = "GET";
    req.url = baseUrl.empty() ? url : baseUrl + url;
    req.headers = headers;
    
    // 添加默认头部
    for (const auto& header : defaultHeaders) {
        if (req.headers.find(header.first) == req.headers.end()) {
            req.headers[header.first] = header.second;
        }
    }
    
    return request(req);
}

HttpResponse HttpClient::post(const std::string& url, const std::string& body, const std::map<std::string, std::string>& headers) {
    HttpRequest req;
    req.method = "POST";
    req.url = baseUrl.empty() ? url : baseUrl + url;
    req.body = body;
    req.headers = headers;
    
    // 添加默认头部
    for (const auto& header : defaultHeaders) {
        if (req.headers.find(header.first) == req.headers.end()) {
            req.headers[header.first] = header.second;
        }
    }
    
    // 设置Content-Type如果没有指定
    if (req.headers.find("Content-Type") == req.headers.end() && !body.empty()) {
        req.headers["Content-Type"] = "application/json";
    }
    
    return request(req);
}

HttpResponse HttpClient::put(const std::string& url, const std::string& body, const std::map<std::string, std::string>& headers) {
    HttpRequest req;
    req.method = "PUT";
    req.url = baseUrl.empty() ? url : baseUrl + url;
    req.body = body;
    req.headers = headers;
    
    // 添加默认头部
    for (const auto& header : defaultHeaders) {
        if (req.headers.find(header.first) == req.headers.end()) {
            req.headers[header.first] = header.second;
        }
    }
    
    if (req.headers.find("Content-Type") == req.headers.end() && !body.empty()) {
        req.headers["Content-Type"] = "application/json";
    }
    
    return request(req);
}

HttpResponse HttpClient::del(const std::string& url, const std::map<std::string, std::string>& headers) {
    HttpRequest req;
    req.method = "DELETE";
    req.url = baseUrl.empty() ? url : baseUrl + url;
    req.headers = headers;
    
    // 添加默认头部
    for (const auto& header : defaultHeaders) {
        if (req.headers.find(header.first) == req.headers.end()) {
            req.headers[header.first] = header.second;
        }
    }
    
    return request(req);
}

void HttpClient::setDefaultHeader(const std::string& key, const std::string& value) {
    defaultHeaders[key] = value;
}

void HttpClient::removeDefaultHeader(const std::string& key) {
    defaultHeaders.erase(key);
}

void HttpClient::setBaseUrl(const std::string& base_url) {
    baseUrl = base_url;
}

HttpResponse HttpClient::performRequest(const HttpRequest& req) {
    HttpResponse response;
    response.success = false;
    
    if (!hSession) {
        response.error_message = "HTTP session not initialized";
        return response;
    }
    
    try {
        std::wstring host, path;
        INTERNET_PORT port;
        bool isHttps;
        parseUrl(req.url, host, path, port, isHttps);
        
        // 连接到服务器
        HINTERNET hConnect = WinHttpConnect(hSession, host.c_str(), port, 0);
        if (!hConnect) {
            response.error_message = "Failed to connect to server";
            return response;
        }
        
        // 创建请求
        DWORD flags = isHttps ? WINHTTP_FLAG_SECURE : 0;
        std::wstring method = stringToWString(req.method);
        
        HINTERNET hRequest = WinHttpOpenRequest(hConnect,
                                               method.c_str(),
                                               path.c_str(),
                                               NULL,
                                               WINHTTP_NO_REFERER,
                                               WINHTTP_DEFAULT_ACCEPT_TYPES,
                                               flags);
        
        if (!hRequest) {
            WinHttpCloseHandle(hConnect);
            response.error_message = "Failed to create request";
            return response;
        }
        
        // 设置超时
        WinHttpSetTimeouts(hRequest, req.timeout_ms, req.timeout_ms, req.timeout_ms, req.timeout_ms);
        
        // 添加头部
        std::wstring headers;
        for (const auto& header : req.headers) {
            headers += stringToWString(header.first + ": " + header.second + "\r\n");
        }
        
        if (!headers.empty()) {
            WinHttpAddRequestHeaders(hRequest, headers.c_str(), -1, WINHTTP_ADDREQ_FLAG_ADD);
        }
        
        // 发送请求
        BOOL result = WinHttpSendRequest(hRequest,
                                        WINHTTP_NO_ADDITIONAL_HEADERS,
                                        0,
                                        (LPVOID)req.body.c_str(),
                                        req.body.length(),
                                        req.body.length(),
                                        0);
        
        if (!result) {
            WinHttpCloseHandle(hRequest);
            WinHttpCloseHandle(hConnect);
            response.error_message = "Failed to send request";
            return response;
        }
        
        // 接收响应
        result = WinHttpReceiveResponse(hRequest, NULL);
        if (!result) {
            WinHttpCloseHandle(hRequest);
            WinHttpCloseHandle(hConnect);
            response.error_message = "Failed to receive response";
            return response;
        }
        
        // 获取状态码
        DWORD statusCode = 0;
        DWORD statusCodeSize = sizeof(statusCode);
        WinHttpQueryHeaders(hRequest,
                           WINHTTP_QUERY_STATUS_CODE | WINHTTP_QUERY_FLAG_NUMBER,
                           WINHTTP_HEADER_NAME_BY_INDEX,
                           &statusCode,
                           &statusCodeSize,
                           WINHTTP_NO_HEADER_INDEX);
        
        response.status_code = statusCode;
        
        // 读取响应体
        std::string responseBody;
        DWORD bytesAvailable = 0;
        
        do {
            bytesAvailable = 0;
            if (!WinHttpQueryDataAvailable(hRequest, &bytesAvailable)) {
                break;
            }
            
            if (bytesAvailable > 0) {
                std::vector<char> buffer(bytesAvailable + 1);
                DWORD bytesRead = 0;
                
                if (WinHttpReadData(hRequest, buffer.data(), bytesAvailable, &bytesRead)) {
                    buffer[bytesRead] = '\0';
                    responseBody.append(buffer.data(), bytesRead);
                }
            }
        } while (bytesAvailable > 0);
        
        response.body = responseBody;
        response.success = true;
        
        // 清理资源
        WinHttpCloseHandle(hRequest);
        WinHttpCloseHandle(hConnect);
        
    } catch (const std::exception& e) {
        response.error_message = e.what();
    }
    
    return response;
}

std::wstring HttpClient::stringToWString(const std::string& str) {
    if (str.empty()) return std::wstring();
    
    int size = MultiByteToWideChar(CP_UTF8, 0, str.c_str(), -1, NULL, 0);
    std::wstring wstr(size, 0);
    MultiByteToWideChar(CP_UTF8, 0, str.c_str(), -1, &wstr[0], size);
    return wstr;
}

std::string HttpClient::wstringToString(const std::wstring& wstr) {
    if (wstr.empty()) return std::string();
    
    int size = WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, NULL, 0, NULL, NULL);
    std::string str(size, 0);
    WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, &str[0], size, NULL, NULL);
    return str;
}

void HttpClient::parseUrl(const std::string& url, std::wstring& host, std::wstring& path, INTERNET_PORT& port, bool& isHttps) {
    std::string urlCopy = url;
    
    // 检查协议
    isHttps = false;
    if (urlCopy.find("https://") == 0) {
        isHttps = true;
        urlCopy = urlCopy.substr(8);
        port = INTERNET_DEFAULT_HTTPS_PORT;
    } else if (urlCopy.find("http://") == 0) {
        urlCopy = urlCopy.substr(7);
        port = INTERNET_DEFAULT_HTTP_PORT;
    } else {
        port = INTERNET_DEFAULT_HTTP_PORT;
    }
    
    // 分离主机和路径
    size_t pathPos = urlCopy.find('/');
    if (pathPos != std::string::npos) {
        host = stringToWString(urlCopy.substr(0, pathPos));
        path = stringToWString(urlCopy.substr(pathPos));
    } else {
        host = stringToWString(urlCopy);
        path = L"/";
    }
    
    // 检查端口
    size_t portPos = host.find(L':');
    if (portPos != std::wstring::npos) {
        std::wstring portStr = host.substr(portPos + 1);
        host = host.substr(0, portPos);
        port = static_cast<INTERNET_PORT>(_wtoi(portStr.c_str()));
    }
}

bool initHttpClient() {
    try {
        g_httpClient = std::make_unique<HttpClient>();
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Failed to initialize HTTP client: " << e.what() << std::endl;
        return false;
    }
}

void cleanupHttpClient() {
    g_httpClient.reset();
}

} // namespace TaishangLaojun