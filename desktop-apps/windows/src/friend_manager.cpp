#include "friend_manager.h"
#include <json/json.h>
#include <thread>
#include <chrono>
#include <sstream>

// 全局好友管理器实例
std::unique_ptr<FriendManager> g_friendManager = nullptr;

// 全局初始化和清理函数
bool initFriendManager() {
    if (g_friendManager) {
        return true; // 已经初始化
    }
    
    g_friendManager = std::make_unique<FriendManager>();
    return g_friendManager->initialize();
}

void cleanupFriendManager() {
    if (g_friendManager) {
        g_friendManager->cleanup();
        g_friendManager.reset();
    }
}

// Friend 结构实现
Friend::Friend(const std::string& json) {
    Json::Value root;
    Json::Reader reader;
    
    if (reader.parse(json, root)) {
        id = root.get("id", "").asString();
        username = root.get("username", "").asString();
        email = root.get("email", "").asString();
        avatarUrl = root.get("avatar_url", "").asString();
        status = stringToFriendStatus(root.get("status", "").asString());
        onlineStatus = stringToOnlineStatus(root.get("online_status", "").asString());
        lastSeen = root.get("last_seen", "").asString();
        createdAt = root.get("created_at", "").asString();
        updatedAt = root.get("updated_at", "").asString();
    }
}

std::string Friend::toJson() const {
    Json::Value root;
    root["id"] = id;
    root["username"] = username;
    root["email"] = email;
    root["avatar_url"] = avatarUrl;
    root["status"] = friendStatusToString(status);
    root["online_status"] = onlineStatusToString(onlineStatus);
    root["last_seen"] = lastSeen;
    root["created_at"] = createdAt;
    root["updated_at"] = updatedAt;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, root);
}

// FriendRequest 结构实现
FriendRequest::FriendRequest(const std::string& json) {
    Json::Value root;
    Json::Reader reader;
    
    if (reader.parse(json, root)) {
        id = root.get("id", "").asString();
        fromUserId = root.get("from_user_id", "").asString();
        toUserId = root.get("to_user_id", "").asString();
        fromUsername = root.get("from_username", "").asString();
        toUsername = root.get("to_username", "").asString();
        message = root.get("message", "").asString();
        status = stringToFriendStatus(root.get("status", "").asString());
        createdAt = root.get("created_at", "").asString();
        updatedAt = root.get("updated_at", "").asString();
    }
}

std::string FriendRequest::toJson() const {
    Json::Value root;
    root["id"] = id;
    root["from_user_id"] = fromUserId;
    root["to_user_id"] = toUserId;
    root["from_username"] = fromUsername;
    root["to_username"] = toUsername;
    root["message"] = message;
    root["status"] = friendStatusToString(status);
    root["created_at"] = createdAt;
    root["updated_at"] = updatedAt;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, root);
}

// AddFriendRequest 结构实现
std::string AddFriendRequest::toJson() const {
    Json::Value root;
    root["username"] = username;
    root["message"] = message;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, root);
}

// FriendResponse 结构实现
FriendResponse::FriendResponse(const std::string& json) {
    Json::Value root;
    Json::Reader reader;
    
    if (reader.parse(json, root)) {
        success = root.get("success", false).asBool();
        message = root.get("message", "").asString();
        
        const Json::Value& friendsArray = root["friends"];
        if (friendsArray.isArray()) {
            for (const auto& friendJson : friendsArray) {
                Json::StreamWriterBuilder builder;
                std::string friendStr = Json::writeString(builder, friendJson);
                friends.emplace_back(friendStr);
            }
        }
        
        const Json::Value& requestsArray = root["requests"];
        if (requestsArray.isArray()) {
            for (const auto& requestJson : requestsArray) {
                Json::StreamWriterBuilder builder;
                std::string requestStr = Json::writeString(builder, requestJson);
                requests.emplace_back(requestStr);
            }
        }
    }
}

// FriendManager 实现
FriendManager::FriendManager() 
    : serverUrl("http://localhost:8081")
    , currentOnlineStatus(OnlineStatus::OFFLINE)
    , autoRefreshEnabled(true)
    , refreshInterval(30)
    , isRunning(false) {
}

FriendManager::~FriendManager() {
    cleanup();
}

bool FriendManager::initialize() {
    httpClient = std::make_unique<HttpClient>();
    if (!httpClient) {
        return false;
    }
    
    currentOnlineStatus = OnlineStatus::ONLINE;
    isRunning = true;
    
    if (autoRefreshEnabled) {
        startAutoRefresh();
    }
    
    return true;
}

void FriendManager::cleanup() {
    isRunning = false;
    stopAutoRefresh();
    
    if (refreshThread.joinable()) {
        refreshThread.join();
    }
    
    httpClient.reset();
}

// 同步方法实现
FriendResponse FriendManager::getFriendList() {
    std::string url = buildUrl("/api/friends");
    HttpRequest request = createAuthenticatedRequest(url, "GET");
    
    HttpResponse response = httpClient->get(url);
    return parseResponse(response);
}

FriendResponse FriendManager::getFriendRequests() {
    std::string url = buildUrl("/api/friends/requests");
    HttpRequest request = createAuthenticatedRequest(url, "GET");
    
    HttpResponse response = httpClient->get(url);
    return parseResponse(response);
}

bool FriendManager::addFriend(const std::string& username, const std::string& message) {
    AddFriendRequest request;
    request.username = username;
    request.message = message;
    
    std::string url = buildUrl("/api/friends/add");
    std::string jsonData = request.toJson();
    
    HttpResponse response = httpClient->post(url, jsonData);
    
    if (response.statusCode == 200 || response.statusCode == 201) {
        FriendResponse friendResponse(response.body);
        return friendResponse.success;
    }
    
    return false;
}

bool FriendManager::respondToFriendRequest(const std::string& requestId, bool accept) {
    Json::Value root;
    root["action"] = accept ? "accept" : "decline";
    
    Json::StreamWriterBuilder builder;
    std::string jsonData = Json::writeString(builder, root);
    
    std::string url = buildUrl("/api/friends/requests/" + requestId);
    HttpResponse response = httpClient->put(url, jsonData);
    
    if (response.statusCode == 200) {
        FriendResponse friendResponse(response.body);
        return friendResponse.success;
    }
    
    return false;
}

bool FriendManager::removeFriend(const std::string& friendId) {
    std::string url = buildUrl("/api/friends/" + friendId);
    HttpResponse response = httpClient->deleteRequest(url);
    
    if (response.statusCode == 200) {
        FriendResponse friendResponse(response.body);
        return friendResponse.success;
    }
    
    return false;
}

bool FriendManager::blockFriend(const std::string& friendId) {
    Json::Value root;
    root["action"] = "block";
    
    Json::StreamWriterBuilder builder;
    std::string jsonData = Json::writeString(builder, root);
    
    std::string url = buildUrl("/api/friends/" + friendId + "/block");
    HttpResponse response = httpClient->put(url, jsonData);
    
    if (response.statusCode == 200) {
        FriendResponse friendResponse(response.body);
        return friendResponse.success;
    }
    
    return false;
}

bool FriendManager::unblockFriend(const std::string& friendId) {
    Json::Value root;
    root["action"] = "unblock";
    
    Json::StreamWriterBuilder builder;
    std::string jsonData = Json::writeString(builder, root);
    
    std::string url = buildUrl("/api/friends/" + friendId + "/unblock");
    HttpResponse response = httpClient->put(url, jsonData);
    
    if (response.statusCode == 200) {
        FriendResponse friendResponse(response.body);
        return friendResponse.success;
    }
    
    return false;
}

// 异步方法实现
void FriendManager::getFriendListAsync(FriendListCallback callback) {
    std::thread([this, callback]() {
        FriendResponse response = getFriendList();
        if (callback) {
            callback(response);
        }
    }).detach();
}

void FriendManager::getFriendRequestsAsync(FriendRequestCallback callback) {
    std::thread([this, callback]() {
        FriendResponse response = getFriendRequests();
        if (callback) {
            callback(response);
        }
    }).detach();
}

void FriendManager::addFriendAsync(const std::string& username, const std::string& message, AddFriendCallback callback) {
    std::thread([this, username, message, callback]() {
        bool success = addFriend(username, message);
        if (callback) {
            callback(success, success ? "Friend request sent successfully" : "Failed to send friend request");
        }
    }).detach();
}

void FriendManager::respondToFriendRequestAsync(const std::string& requestId, bool accept, RespondFriendCallback callback) {
    std::thread([this, requestId, accept, callback]() {
        bool success = respondToFriendRequest(requestId, accept);
        std::string message = success ? 
            (accept ? "Friend request accepted" : "Friend request declined") : 
            "Failed to respond to friend request";
        if (callback) {
            callback(success, message);
        }
    }).detach();
}

void FriendManager::removeFriendAsync(const std::string& friendId, RemoveFriendCallback callback) {
    std::thread([this, friendId, callback]() {
        bool success = removeFriend(friendId);
        if (callback) {
            callback(success, success ? "Friend removed successfully" : "Failed to remove friend");
        }
    }).detach();
}

void FriendManager::blockFriendAsync(const std::string& friendId, RespondFriendCallback callback) {
    std::thread([this, friendId, callback]() {
        bool success = blockFriend(friendId);
        if (callback) {
            callback(success, success ? "Friend blocked successfully" : "Failed to block friend");
        }
    }).detach();
}

void FriendManager::unblockFriendAsync(const std::string& friendId, RespondFriendCallback callback) {
    std::thread([this, friendId, callback]() {
        bool success = unblockFriend(friendId);
        if (callback) {
            callback(success, success ? "Friend unblocked successfully" : "Failed to unblock friend");
        }
    }).detach();
}

// 好友状态管理
void FriendManager::updateOnlineStatus(OnlineStatus status) {
    currentOnlineStatus = status;
    
    Json::Value root;
    root["status"] = onlineStatusToString(status);
    
    Json::StreamWriterBuilder builder;
    std::string jsonData = Json::writeString(builder, root);
    
    std::string url = buildUrl("/api/user/status");
    httpClient->put(url, jsonData);
}

OnlineStatus FriendManager::getOnlineStatus() const {
    return currentOnlineStatus;
}

Friend* FriendManager::findFriendById(const std::string& friendId) {
    for (auto& friend_ : friends) {
        if (friend_.id == friendId) {
            return &friend_;
        }
    }
    return nullptr;
}

Friend* FriendManager::findFriendByUsername(const std::string& username) {
    for (auto& friend_ : friends) {
        if (friend_.username == username) {
            return &friend_;
        }
    }
    return nullptr;
}

// 配置方法
void FriendManager::setServerUrl(const std::string& url) {
    serverUrl = url;
}

void FriendManager::enableAutoRefresh(bool enable) {
    autoRefreshEnabled = enable;
    if (enable && isRunning) {
        startAutoRefresh();
    } else {
        stopAutoRefresh();
    }
}

void FriendManager::setRefreshInterval(int seconds) {
    refreshInterval = seconds;
}

// 事件回调设置
void FriendManager::setOnFriendListUpdated(FriendListCallback callback) {
    onFriendListUpdated = callback;
}

void FriendManager::setOnFriendRequestReceived(FriendRequestCallback callback) {
    onFriendRequestReceived = callback;
}

void FriendManager::setOnFriendStatusChanged(std::function<void(const Friend&)> callback) {
    onFriendStatusChanged = callback;
}

// 内部方法
std::string FriendManager::buildUrl(const std::string& endpoint) const {
    return serverUrl + endpoint;
}

HttpRequest FriendManager::createAuthenticatedRequest(const std::string& url, const std::string& method) const {
    HttpRequest request;
    request.url = url;
    request.method = method;
    
    // 添加认证头
    if (g_authManager && g_authManager->isLoggedIn()) {
        std::string token = g_authManager->getAccessToken();
        if (!token.empty()) {
            request.headers["Authorization"] = "Bearer " + token;
        }
    }
    
    request.headers["Content-Type"] = "application/json";
    return request;
}

FriendResponse FriendManager::parseResponse(const HttpResponse& response) const {
    if (response.statusCode == 200) {
        return FriendResponse(response.body);
    }
    
    FriendResponse errorResponse;
    errorResponse.success = false;
    errorResponse.message = "HTTP Error: " + std::to_string(response.statusCode);
    return errorResponse;
}

void FriendManager::startAutoRefresh() {
    if (refreshThread.joinable()) {
        return; // 已经在运行
    }
    
    refreshThread = std::thread([this]() {
        while (isRunning && autoRefreshEnabled) {
            refreshFriendData();
            std::this_thread::sleep_for(std::chrono::seconds(refreshInterval));
        }
    });
}

void FriendManager::stopAutoRefresh() {
    // 线程会在下次循环时自动停止
}

void FriendManager::refreshFriendData() {
    // 刷新好友列表
    FriendResponse friendListResponse = getFriendList();
    if (friendListResponse.success) {
        friends = friendListResponse.friends;
        if (onFriendListUpdated) {
            onFriendListUpdated(friendListResponse);
        }
    }
    
    // 刷新好友请求
    FriendResponse requestsResponse = getFriendRequests();
    if (requestsResponse.success) {
        pendingRequests = requestsResponse.requests;
        if (onFriendRequestReceived) {
            onFriendRequestReceived(requestsResponse);
        }
    }
}

// 辅助函数实现
std::string friendStatusToString(FriendStatus status) {
    switch (status) {
        case FriendStatus::PENDING: return "pending";
        case FriendStatus::ACCEPTED: return "accepted";
        case FriendStatus::BLOCKED: return "blocked";
        case FriendStatus::DECLINED: return "declined";
        default: return "unknown";
    }
}

FriendStatus stringToFriendStatus(const std::string& status) {
    if (status == "pending") return FriendStatus::PENDING;
    if (status == "accepted") return FriendStatus::ACCEPTED;
    if (status == "blocked") return FriendStatus::BLOCKED;
    if (status == "declined") return FriendStatus::DECLINED;
    return FriendStatus::PENDING;
}

std::string onlineStatusToString(OnlineStatus status) {
    switch (status) {
        case OnlineStatus::ONLINE: return "online";
        case OnlineStatus::OFFLINE: return "offline";
        case OnlineStatus::AWAY: return "away";
        case OnlineStatus::BUSY: return "busy";
        default: return "offline";
    }
}

OnlineStatus stringToOnlineStatus(const std::string& status) {
    if (status == "online") return OnlineStatus::ONLINE;
    if (status == "offline") return OnlineStatus::OFFLINE;
    if (status == "away") return OnlineStatus::AWAY;
    if (status == "busy") return OnlineStatus::BUSY;
    return OnlineStatus::OFFLINE;
}