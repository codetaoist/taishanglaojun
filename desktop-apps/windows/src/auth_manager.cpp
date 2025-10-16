#include "../include/auth_manager.h"
#include <json/json.h>
#include <fstream>
#include <iostream>
#include <thread>
#include <chrono>
#include <shlobj.h>

namespace TaishangLaojun {

std::unique_ptr<AuthManager> g_authManager = nullptr;

AuthManager::AuthManager() 
    : loggedIn(false), autoRefreshEnabled(true) {
    httpClient = std::make_unique<HttpClient>();
    authServerUrl = "http://localhost:8082"; // 默认认证服务器地址
    loadAuthData();
}

AuthManager::~AuthManager() {
    // 析构函数会自动清理资源
}

AuthResponse AuthManager::login(const LoginRequest& request) {
    std::string url = buildAuthUrl("/auth/login");
    std::string jsonBody = serializeLoginRequest(request);
    
    std::map<std::string, std::string> headers;
    headers["Content-Type"] = "application/json";
    
    HttpResponse response = httpClient->post(url, jsonBody, headers);
    AuthResponse authResponse = parseAuthResponse(response);
    
    if (authResponse.success) {
        saveAuthData(authResponse);
        if (autoRefreshEnabled) {
            scheduleTokenRefresh();
        }
    }
    
    return authResponse;
}

AuthResponse AuthManager::registerUser(const RegisterRequest& request) {
    std::string url = buildAuthUrl("/auth/register");
    std::string jsonBody = serializeRegisterRequest(request);
    
    std::map<std::string, std::string> headers;
    headers["Content-Type"] = "application/json";
    
    HttpResponse response = httpClient->post(url, jsonBody, headers);
    return parseAuthResponse(response);
}

bool AuthManager::logout() {
    std::string url = buildAuthUrl("/auth/logout");
    
    std::map<std::string, std::string> headers;
    headers["Authorization"] = "Bearer " + accessToken;
    
    HttpResponse response = httpClient->post(url, "", headers);
    
    // 无论服务器响应如何，都清除本地认证数据
    clearAuthData();
    
    return response.success;
}

bool AuthManager::refreshToken() {
    if (refreshToken.empty()) {
        return false;
    }
    
    std::string url = buildAuthUrl("/auth/refresh");
    
    Json::Value jsonBody;
    jsonBody["refresh_token"] = refreshToken;
    
    Json::StreamWriterBuilder builder;
    std::string jsonString = Json::writeString(builder, jsonBody);
    
    std::map<std::string, std::string> headers;
    headers["Content-Type"] = "application/json";
    
    HttpResponse response = httpClient->post(url, jsonString, headers);
    
    if (response.success) {
        AuthResponse authResponse = parseAuthResponse(response);
        if (authResponse.success) {
            saveAuthData(authResponse);
            return true;
        }
    }
    
    // 刷新失败，清除认证数据
    clearAuthData();
    return false;
}

void AuthManager::loginAsync(const LoginRequest& request, std::function<void(const AuthResponse&)> callback) {
    std::thread([this, request, callback]() {
        AuthResponse response = login(request);
        callback(response);
    }).detach();
}

void AuthManager::registerAsync(const RegisterRequest& request, std::function<void(const AuthResponse&)> callback) {
    std::thread([this, request, callback]() {
        AuthResponse response = registerUser(request);
        callback(response);
    }).detach();
}

void AuthManager::logoutAsync(std::function<void(bool)> callback) {
    std::thread([this, callback]() {
        bool success = logout();
        callback(success);
    }).detach();
}

void AuthManager::refreshTokenAsync(std::function<void(bool)> callback) {
    std::thread([this, callback]() {
        bool success = refreshToken();
        callback(success);
    }).detach();
}

bool AuthManager::isLoggedIn() const {
    return loggedIn && !accessToken.empty();
}

std::string AuthManager::getAccessToken() const {
    return accessToken;
}

std::string AuthManager::getRefreshToken() const {
    return refreshToken;
}

User AuthManager::getCurrentUser() const {
    return currentUser;
}

void AuthManager::setAuthServerUrl(const std::string& url) {
    authServerUrl = url;
}

void AuthManager::enableAutoRefresh(bool enable) {
    autoRefreshEnabled = enable;
}

void AuthManager::clearAuthData() {
    accessToken.clear();
    refreshToken.clear();
    currentUser = User{};
    loggedIn = false;
    
    // 删除本地存储的认证文件
    char* appDataPath;
    if (SHGetKnownFolderPath(FOLDERID_RoamingAppData, 0, NULL, (PWSTR*)&appDataPath) == S_OK) {
        std::string authFilePath = std::string((char*)appDataPath) + "\\TaishangLaojun\\auth.json";
        DeleteFileA(authFilePath.c_str());
        CoTaskMemFree(appDataPath);
    }
}

void AuthManager::saveAuthData(const AuthResponse& response) {
    accessToken = response.access_token;
    refreshToken = response.refresh_token;
    currentUser = response.user;
    loggedIn = true;
    
    // 保存到本地文件
    Json::Value authData;
    authData["access_token"] = accessToken;
    authData["refresh_token"] = refreshToken;
    authData["user"]["id"] = currentUser.id;
    authData["user"]["username"] = currentUser.username;
    authData["user"]["email"] = currentUser.email;
    authData["user"]["avatar_url"] = currentUser.avatar_url;
    authData["user"]["created_at"] = currentUser.created_at;
    authData["user"]["updated_at"] = currentUser.updated_at;
    
    char* appDataPath;
    if (SHGetKnownFolderPath(FOLDERID_RoamingAppData, 0, NULL, (PWSTR*)&appDataPath) == S_OK) {
        std::string appDir = std::string((char*)appDataPath) + "\\TaishangLaojun";
        CreateDirectoryA(appDir.c_str(), NULL);
        
        std::string authFilePath = appDir + "\\auth.json";
        std::ofstream file(authFilePath);
        if (file.is_open()) {
            Json::StreamWriterBuilder builder;
            std::unique_ptr<Json::StreamWriter> writer(builder.newStreamWriter());
            writer->write(authData, &file);
            file.close();
        }
        
        CoTaskMemFree(appDataPath);
    }
    
    // 设置HTTP客户端的默认认证头
    httpClient->setDefaultHeader("Authorization", "Bearer " + accessToken);
}

void AuthManager::loadAuthData() {
    char* appDataPath;
    if (SHGetKnownFolderPath(FOLDERID_RoamingAppData, 0, NULL, (PWSTR*)&appDataPath) != S_OK) {
        return;
    }
    
    std::string authFilePath = std::string((char*)appDataPath) + "\\TaishangLaojun\\auth.json";
    CoTaskMemFree(appDataPath);
    
    std::ifstream file(authFilePath);
    if (!file.is_open()) {
        return;
    }
    
    Json::Value authData;
    Json::CharReaderBuilder builder;
    std::string errors;
    
    if (Json::parseFromStream(builder, file, &authData, &errors)) {
        accessToken = authData.get("access_token", "").asString();
        refreshToken = authData.get("refresh_token", "").asString();
        
        if (authData.isMember("user")) {
            Json::Value userData = authData["user"];
            currentUser.id = userData.get("id", "").asString();
            currentUser.username = userData.get("username", "").asString();
            currentUser.email = userData.get("email", "").asString();
            currentUser.avatar_url = userData.get("avatar_url", "").asString();
            currentUser.created_at = userData.get("created_at", "").asString();
            currentUser.updated_at = userData.get("updated_at", "").asString();
        }
        
        if (!accessToken.empty() && validateToken(accessToken)) {
            loggedIn = true;
            httpClient->setDefaultHeader("Authorization", "Bearer " + accessToken);
            
            if (autoRefreshEnabled) {
                scheduleTokenRefresh();
            }
        }
    }
    
    file.close();
}

bool AuthManager::validateToken(const std::string& token) {
    // 简单的token格式验证
    return !token.empty() && token.length() > 10;
}

std::string AuthManager::buildAuthUrl(const std::string& endpoint) {
    std::string url = authServerUrl;
    if (!url.empty() && url.back() == '/') {
        url.pop_back();
    }
    
    if (!endpoint.empty() && endpoint.front() != '/') {
        url += "/";
    }
    
    url += endpoint;
    return url;
}

void AuthManager::scheduleTokenRefresh() {
    // 在后台线程中定期刷新token（每25分钟刷新一次，假设token有效期30分钟）
    std::thread([this]() {
        while (loggedIn && autoRefreshEnabled) {
            std::this_thread::sleep_for(std::chrono::minutes(25));
            
            if (loggedIn && autoRefreshEnabled) {
                refreshToken();
            }
        }
    }).detach();
}

std::string AuthManager::serializeLoginRequest(const LoginRequest& request) {
    Json::Value json;
    json["username"] = request.username;
    json["password"] = request.password;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, json);
}

std::string AuthManager::serializeRegisterRequest(const RegisterRequest& request) {
    Json::Value json;
    json["username"] = request.username;
    json["email"] = request.email;
    json["password"] = request.password;
    json["confirm_password"] = request.confirm_password;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, json);
}

AuthResponse AuthManager::parseAuthResponse(const HttpResponse& response) {
    AuthResponse authResponse;
    authResponse.success = false;
    
    if (!response.success) {
        authResponse.message = response.error_message;
        return authResponse;
    }
    
    try {
        Json::Value json;
        Json::CharReaderBuilder builder;
        std::string errors;
        std::istringstream stream(response.body);
        
        if (Json::parseFromStream(builder, stream, &json, &errors)) {
            authResponse.success = json.get("success", false).asBool();
            authResponse.message = json.get("message", "").asString();
            
            if (json.isMember("data")) {
                Json::Value data = json["data"];
                authResponse.access_token = data.get("access_token", "").asString();
                authResponse.refresh_token = data.get("refresh_token", "").asString();
                authResponse.expires_in = data.get("expires_in", 0).asInt();
                
                if (data.isMember("user")) {
                    Json::Value userData = data["user"];
                    authResponse.user.id = userData.get("id", "").asString();
                    authResponse.user.username = userData.get("username", "").asString();
                    authResponse.user.email = userData.get("email", "").asString();
                    authResponse.user.avatar_url = userData.get("avatar_url", "").asString();
                    authResponse.user.created_at = userData.get("created_at", "").asString();
                    authResponse.user.updated_at = userData.get("updated_at", "").asString();
                }
            }
        } else {
            authResponse.message = "Failed to parse response: " + errors;
        }
    } catch (const std::exception& e) {
        authResponse.message = "Exception parsing response: " + std::string(e.what());
    }
    
    return authResponse;
}

bool initAuthManager() {
    try {
        g_authManager = std::make_unique<AuthManager>();
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Failed to initialize auth manager: " << e.what() << std::endl;
        return false;
    }
}

void cleanupAuthManager() {
    g_authManager.reset();
}

} // namespace TaishangLaojun