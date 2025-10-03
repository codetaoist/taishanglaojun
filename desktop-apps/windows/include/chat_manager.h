#pragma once

#include <windows.h>
#include <winhttp.h>
#include <string>
#include <vector>
#include <functional>
#include <memory>
#include <mutex>
#include <thread>
#include <atomic>
#include <map>
#include "http_client.h"

// 消息类型枚举
enum class MessageType {
    TEXT,
    IMAGE,
    FILE,
    SYSTEM,
    EMOJI
};

// 聊天类型枚举
enum class ChatType {
    PRIVATE,
    GROUP
};

// 消息状态枚举
enum class MessageStatus {
    SENDING,
    SENT,
    DELIVERED,
    READ,
    FAILED
};

// 消息结构体
struct Message {
    std::string id;
    std::string chatId;
    std::string senderId;
    std::string senderUsername;
    std::string content;
    MessageType type;
    MessageStatus status;
    std::string timestamp;
    std::string createdAt;
    std::string updatedAt;
    
    // 文件消息相关
    std::string fileName;
    std::string fileUrl;
    size_t fileSize;
    
    // 回复消息相关
    std::string replyToMessageId;
    std::string replyToContent;
    
    Message();
    ~Message();
};

// 聊天会话结构体
struct Chat {
    std::string id;
    std::string name;
    ChatType type;
    std::string avatarUrl;
    std::string lastMessage;
    std::string lastMessageTime;
    int unreadCount;
    std::vector<std::string> participants;
    std::string createdAt;
    std::string updatedAt;
    
    Chat();
    ~Chat();
};

// 发送消息请求结构体
struct SendMessageRequest {
    std::string chatId;
    std::string content;
    MessageType type;
    std::string replyToMessageId;
    
    SendMessageRequest();
    ~SendMessageRequest();
};

// 创建聊天请求结构体
struct CreateChatRequest {
    ChatType type;
    std::string name;
    std::vector<std::string> participants;
    
    CreateChatRequest();
    ~CreateChatRequest();
};

// 聊天响应结构体
struct ChatResponse {
    bool success;
    std::string message;
    std::vector<Chat> chats;
    std::vector<Message> messages;
    Chat chat;
    Message messageData;
    
    ChatResponse();
    ~ChatResponse();
};

// WebSocket消息结构体
struct WebSocketMessage {
    std::string type;
    std::string chatId;
    std::string data;
    std::string timestamp;
    
    WebSocketMessage();
    ~WebSocketMessage();
};

// 事件回调函数类型
typedef std::function<void(const std::vector<Chat>&)> OnChatsUpdatedCallback;
typedef std::function<void(const std::vector<Message>&)> OnMessagesUpdatedCallback;
typedef std::function<void(const Message&)> OnNewMessageCallback;
typedef std::function<void(const Message&)> OnMessageStatusUpdatedCallback;
typedef std::function<void(const std::string&)> OnTypingStatusCallback;
typedef std::function<void(const std::string&)> OnErrorCallback;

// 聊天管理器类
class ChatManager {
public:
    ChatManager();
    ~ChatManager();
    
    // 初始化和清理
    bool Initialize();
    void Cleanup();
    
    // 聊天列表管理
    bool GetChatList();
    bool GetChatListAsync();
    
    // 消息管理
    bool GetMessages(const std::string& chatId, int page = 1, int limit = 50);
    bool GetMessagesAsync(const std::string& chatId, int page = 1, int limit = 50);
    bool SendMessage(const SendMessageRequest& request);
    bool SendMessageAsync(const SendMessageRequest& request);
    bool MarkMessageAsRead(const std::string& messageId);
    bool MarkChatAsRead(const std::string& chatId);
    
    // 聊天会话管理
    bool CreateChat(const CreateChatRequest& request);
    bool CreateChatAsync(const CreateChatRequest& request);
    bool DeleteChat(const std::string& chatId);
    bool LeaveChat(const std::string& chatId);
    bool AddParticipant(const std::string& chatId, const std::string& userId);
    bool RemoveParticipant(const std::string& chatId, const std::string& userId);
    
    // 实时功能
    bool ConnectWebSocket();
    void DisconnectWebSocket();
    bool IsWebSocketConnected() const;
    bool SendTypingStatus(const std::string& chatId, bool isTyping);
    
    // 文件传输
    bool SendFile(const std::string& chatId, const std::string& filePath);
    bool DownloadFile(const std::string& fileUrl, const std::string& savePath);
    
    // 搜索功能
    bool SearchMessages(const std::string& query, const std::string& chatId = "");
    bool SearchChats(const std::string& query);
    
    // 本地数据管理
    Chat* FindChatById(const std::string& chatId);
    Chat* FindChatByParticipant(const std::string& userId);
    Message* FindMessageById(const std::string& messageId);
    std::vector<Message> GetChatMessages(const std::string& chatId);
    
    // 事件回调设置
    void SetOnChatsUpdatedCallback(OnChatsUpdatedCallback callback);
    void SetOnMessagesUpdatedCallback(OnMessagesUpdatedCallback callback);
    void SetOnNewMessageCallback(OnNewMessageCallback callback);
    void SetOnMessageStatusUpdatedCallback(OnMessageStatusUpdatedCallback callback);
    void SetOnTypingStatusCallback(OnTypingStatusCallback callback);
    void SetOnErrorCallback(OnErrorCallback callback);
    
    // 配置方法
    void SetServerUrl(const std::string& url);
    void SetWebSocketUrl(const std::string& url);
    void EnableAutoReconnect(bool enable);
    void SetReconnectInterval(int seconds);
    
    // 状态查询
    bool IsInitialized() const;
    int GetUnreadMessageCount() const;
    int GetChatCount() const;
    
private:
    // 内部方法
    std::string BuildUrl(const std::string& endpoint);
    std::string BuildWebSocketUrl(const std::string& endpoint);
    HttpRequest CreateAuthenticatedRequest(const std::string& url, const std::string& method);
    ChatResponse* ParseChatResponse(const std::string& jsonStr);
    std::string CreateSendMessageJson(const SendMessageRequest& request);
    std::string CreateCreateChatJson(const CreateChatRequest& request);
    
    // WebSocket相关
    void WebSocketThreadFunc();
    void HandleWebSocketMessage(const WebSocketMessage& wsMessage);
    void ProcessIncomingMessage(const std::string& messageJson);
    void ProcessTypingStatus(const std::string& statusJson);
    void ProcessMessageStatus(const std::string& statusJson);
    
    // 数据同步
    void UpdateLocalChats(const std::vector<Chat>& chats);
    void UpdateLocalMessages(const std::string& chatId, const std::vector<Message>& messages);
    void AddNewMessage(const Message& message);
    void UpdateMessageStatus(const std::string& messageId, MessageStatus status);
    
    // 自动重连
    void StartAutoReconnect();
    void StopAutoReconnect();
    void ReconnectThreadFunc();
    
private:
    // 核心组件
    HttpClient* m_httpClient;
    
    // 配置
    std::string m_serverUrl;
    std::string m_webSocketUrl;
    bool m_autoReconnectEnabled;
    int m_reconnectInterval;
    
    // 数据存储
    std::vector<Chat> m_chats;
    std::map<std::string, std::vector<Message>> m_chatMessages;
    std::mutex m_dataMutex;
    
    // WebSocket相关
    HINTERNET m_webSocketHandle;
    std::thread m_webSocketThread;
    std::atomic<bool> m_webSocketConnected;
    std::atomic<bool> m_shouldStopWebSocket;
    
    // 自动重连
    std::thread m_reconnectThread;
    std::atomic<bool> m_shouldStopReconnect;
    
    // 事件回调
    OnChatsUpdatedCallback m_onChatsUpdated;
    OnMessagesUpdatedCallback m_onMessagesUpdated;
    OnNewMessageCallback m_onNewMessage;
    OnMessageStatusUpdatedCallback m_onMessageStatusUpdated;
    OnTypingStatusCallback m_onTypingStatus;
    OnErrorCallback m_onError;
    
    // 状态
    bool m_initialized;
};

// 全局实例管理
extern ChatManager* g_chatManager;

// 全局函数
bool InitChatManager();
void CleanupChatManager();
ChatManager* GetChatManager();

// 辅助函数
const char* MessageTypeToString(MessageType type);
MessageType StringToMessageType(const char* type);
const char* ChatTypeToString(ChatType type);
ChatType StringToChatType(const char* type);
const char* MessageStatusToString(MessageStatus status);
MessageStatus StringToMessageStatus(const char* status);

// 内存管理函数
void MessageFree(Message* message);
void ChatFree(Chat* chat);
void ChatResponseFree(ChatResponse* response);
void WebSocketMessageFree(WebSocketMessage* wsMessage);