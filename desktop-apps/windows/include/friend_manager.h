#pragma once

#include "http_client.h"
#include "auth_manager.h"
#include <vector>
#include <string>
#include <functional>
#include <memory>

// 好友状态枚举
enum class FriendStatus {
    PENDING,    // 待确认
    ACCEPTED,   // 已接受
    BLOCKED,    // 已屏蔽
    DECLINED    // 已拒绝
};

// 在线状态枚举
enum class OnlineStatus {
    ONLINE,     // 在线
    OFFLINE,    // 离线
    AWAY,       // 离开
    BUSY        // 忙碌
};

// 好友信息结构
struct Friend {
    std::string id;
    std::string username;
    std::string email;
    std::string avatarUrl;
    FriendStatus status;
    OnlineStatus onlineStatus;
    std::string lastSeen;
    std::string createdAt;
    std::string updatedAt;
    
    Friend() = default;
    Friend(const std::string& json);
    std::string toJson() const;
};

// 好友请求结构
struct FriendRequest {
    std::string id;
    std::string fromUserId;
    std::string toUserId;
    std::string fromUsername;
    std::string toUsername;
    std::string message;
    FriendStatus status;
    std::string createdAt;
    std::string updatedAt;
    
    FriendRequest() = default;
    FriendRequest(const std::string& json);
    std::string toJson() const;
};

// 添加好友请求结构
struct AddFriendRequest {
    std::string username;
    std::string message;
    
    std::string toJson() const;
};

// 好友响应结构
struct FriendResponse {
    bool success;
    std::string message;
    std::vector<Friend> friends;
    std::vector<FriendRequest> requests;
    
    FriendResponse() = default;
    FriendResponse(const std::string& json);
};

// 回调函数类型
using FriendListCallback = std::function<void(const FriendResponse&)>;
using FriendRequestCallback = std::function<void(const FriendResponse&)>;
using AddFriendCallback = std::function<void(bool, const std::string&)>;
using RespondFriendCallback = std::function<void(bool, const std::string&)>;
using RemoveFriendCallback = std::function<void(bool, const std::string&)>;

// 好友管理器类
class FriendManager {
public:
    FriendManager();
    ~FriendManager();
    
    // 初始化和清理
    bool initialize();
    void cleanup();
    
    // 同步方法
    FriendResponse getFriendList();
    FriendResponse getFriendRequests();
    bool addFriend(const std::string& username, const std::string& message = "");
    bool respondToFriendRequest(const std::string& requestId, bool accept);
    bool removeFriend(const std::string& friendId);
    bool blockFriend(const std::string& friendId);
    bool unblockFriend(const std::string& friendId);
    
    // 异步方法
    void getFriendListAsync(FriendListCallback callback);
    void getFriendRequestsAsync(FriendRequestCallback callback);
    void addFriendAsync(const std::string& username, const std::string& message, AddFriendCallback callback);
    void respondToFriendRequestAsync(const std::string& requestId, bool accept, RespondFriendCallback callback);
    void removeFriendAsync(const std::string& friendId, RemoveFriendCallback callback);
    void blockFriendAsync(const std::string& friendId, RespondFriendCallback callback);
    void unblockFriendAsync(const std::string& friendId, RespondFriendCallback callback);
    
    // 好友状态管理
    void updateOnlineStatus(OnlineStatus status);
    OnlineStatus getOnlineStatus() const;
    Friend* findFriendById(const std::string& friendId);
    Friend* findFriendByUsername(const std::string& username);
    
    // 配置方法
    void setServerUrl(const std::string& url);
    void enableAutoRefresh(bool enable);
    void setRefreshInterval(int seconds);
    
    // 事件回调设置
    void setOnFriendListUpdated(FriendListCallback callback);
    void setOnFriendRequestReceived(FriendRequestCallback callback);
    void setOnFriendStatusChanged(std::function<void(const Friend&)> callback);
    
private:
    // 内部方法
    std::string buildUrl(const std::string& endpoint) const;
    HttpRequest createAuthenticatedRequest(const std::string& url, const std::string& method) const;
    FriendResponse parseResponse(const HttpResponse& response) const;
    void startAutoRefresh();
    void stopAutoRefresh();
    void refreshFriendData();
    
    // 成员变量
    std::unique_ptr<HttpClient> httpClient;
    std::string serverUrl;
    std::vector<Friend> friends;
    std::vector<FriendRequest> pendingRequests;
    OnlineStatus currentOnlineStatus;
    
    // 配置
    bool autoRefreshEnabled;
    int refreshInterval;
    
    // 回调函数
    FriendListCallback onFriendListUpdated;
    FriendRequestCallback onFriendRequestReceived;
    std::function<void(const Friend&)> onFriendStatusChanged;
    
    // 线程和定时器相关
    bool isRunning;
    std::thread refreshThread;
};

// 全局好友管理器实例
extern std::unique_ptr<FriendManager> g_friendManager;

// 全局初始化和清理函数
bool initFriendManager();
void cleanupFriendManager();

// 辅助函数
std::string friendStatusToString(FriendStatus status);
FriendStatus stringToFriendStatus(const std::string& status);
std::string onlineStatusToString(OnlineStatus status);
OnlineStatus stringToOnlineStatus(const std::string& status);