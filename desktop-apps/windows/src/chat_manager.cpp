#include "chat_manager.h"
#include "auth_manager.h"
#include <json/json.h>
#include <sstream>
#include <algorithm>
#include <fstream>

// 全局实例
ChatManager* g_chatManager = nullptr;

// 构造函数和析构函数实现
Message::Message() : type(MessageType::TEXT), status(MessageStatus::SENDING), fileSize(0) {}
Message::~Message() {}

Chat::Chat() : type(ChatType::PRIVATE), unreadCount(0) {}
Chat::~Chat() {}

SendMessageRequest::SendMessageRequest() : type(MessageType::TEXT) {}
SendMessageRequest::~SendMessageRequest() {}

CreateChatRequest::CreateChatRequest() : type(ChatType::PRIVATE) {}
CreateChatRequest::~CreateChatRequest() {}

ChatResponse::ChatResponse() : success(false) {}
ChatResponse::~ChatResponse() {}

WebSocketMessage::WebSocketMessage() {}
WebSocketMessage::~WebSocketMessage() {}

// ChatManager实现
ChatManager::ChatManager() 
    : m_httpClient(nullptr)
    , m_serverUrl("http://localhost:8081")
    , m_webSocketUrl("ws://localhost:8081")
    , m_autoReconnectEnabled(true)
    , m_reconnectInterval(5)
    , m_webSocketHandle(nullptr)
    , m_webSocketConnected(false)
    , m_shouldStopWebSocket(false)
    , m_shouldStopReconnect(false)
    , m_initialized(false) {
}

ChatManager::~ChatManager() {
    Cleanup();
}

bool ChatManager::Initialize() {
    if (m_initialized) {
        return true;
    }
    
    // 获取HTTP客户端实例
    m_httpClient = GetHttpClient();
    if (!m_httpClient) {
        return false;
    }
    
    m_initialized = true;
    
    // 连接WebSocket
    ConnectWebSocket();
    
    return true;
}

void ChatManager::Cleanup() {
    if (!m_initialized) {
        return;
    }
    
    // 断开WebSocket连接
    DisconnectWebSocket();
    
    // 停止自动重连
    StopAutoReconnect();
    
    // 清理数据
    std::lock_guard<std::mutex> lock(m_dataMutex);
    m_chats.clear();
    m_chatMessages.clear();
    
    m_initialized = false;
}

// 聊天列表管理
bool ChatManager::GetChatList() {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/chats");
    HttpRequest request = CreateAuthenticatedRequest(url, "GET");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        ChatResponse* chatResponse = ParseChatResponse(response.body);
        if (chatResponse && chatResponse->success) {
            UpdateLocalChats(chatResponse->chats);
            
            if (m_onChatsUpdated) {
                m_onChatsUpdated(m_chats);
            }
            
            ChatResponseFree(chatResponse);
            return true;
        }
        ChatResponseFree(chatResponse);
    }
    
    return false;
}

bool ChatManager::GetChatListAsync() {
    std::thread([this]() {
        GetChatList();
    }).detach();
    
    return true;
}

// 消息管理
bool ChatManager::GetMessages(const std::string& chatId, int page, int limit) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty()) {
        return false;
    }
    
    std::ostringstream oss;
    oss << "/api/chats/" << chatId << "/messages?page=" << page << "&limit=" << limit;
    std::string url = BuildUrl(oss.str());
    
    HttpRequest request = CreateAuthenticatedRequest(url, "GET");
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        ChatResponse* chatResponse = ParseChatResponse(response.body);
        if (chatResponse && chatResponse->success) {
            UpdateLocalMessages(chatId, chatResponse->messages);
            
            if (m_onMessagesUpdated) {
                m_onMessagesUpdated(chatResponse->messages);
            }
            
            ChatResponseFree(chatResponse);
            return true;
        }
        ChatResponseFree(chatResponse);
    }
    
    return false;
}

bool ChatManager::GetMessagesAsync(const std::string& chatId, int page, int limit) {
    std::thread([this, chatId, page, limit]() {
        GetMessages(chatId, page, limit);
    }).detach();
    
    return true;
}

bool ChatManager::SendMessage(const SendMessageRequest& request) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || request.chatId.empty()) {
        return false;
    }
    
    std::string jsonStr = CreateSendMessageJson(request);
    std::string url = BuildUrl("/api/messages");
    
    HttpRequest httpRequest = CreateAuthenticatedRequest(url, "POST");
    httpRequest.body = jsonStr;
    
    HttpResponse response = m_httpClient->SendRequest(httpRequest);
    
    if (response.statusCode == 200 || response.statusCode == 201) {
        ChatResponse* chatResponse = ParseChatResponse(response.body);
        if (chatResponse && chatResponse->success) {
            AddNewMessage(chatResponse->messageData);
            
            if (m_onNewMessage) {
                m_onNewMessage(chatResponse->messageData);
            }
            
            ChatResponseFree(chatResponse);
            return true;
        }
        ChatResponseFree(chatResponse);
    }
    
    return false;
}

bool ChatManager::SendMessageAsync(const SendMessageRequest& request) {
    std::thread([this, request]() {
        SendMessage(request);
    }).detach();
    
    return true;
}

bool ChatManager::MarkMessageAsRead(const std::string& messageId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || messageId.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/messages/" + messageId + "/read");
    HttpRequest request = CreateAuthenticatedRequest(url, "PUT");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        UpdateMessageStatus(messageId, MessageStatus::READ);
        return true;
    }
    
    return false;
}

bool ChatManager::MarkChatAsRead(const std::string& chatId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/chats/" + chatId + "/read");
    HttpRequest request = CreateAuthenticatedRequest(url, "PUT");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        // 更新本地聊天的未读计数
        std::lock_guard<std::mutex> lock(m_dataMutex);
        for (auto& chat : m_chats) {
            if (chat.id == chatId) {
                chat.unreadCount = 0;
                break;
            }
        }
        
        if (m_onChatsUpdated) {
            m_onChatsUpdated(m_chats);
        }
        
        return true;
    }
    
    return false;
}

// 聊天会话管理
bool ChatManager::CreateChat(const CreateChatRequest& request) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn()) {
        return false;
    }
    
    std::string jsonStr = CreateCreateChatJson(request);
    std::string url = BuildUrl("/api/chats");
    
    HttpRequest httpRequest = CreateAuthenticatedRequest(url, "POST");
    httpRequest.body = jsonStr;
    
    HttpResponse response = m_httpClient->SendRequest(httpRequest);
    
    if (response.statusCode == 200 || response.statusCode == 201) {
        ChatResponse* chatResponse = ParseChatResponse(response.body);
        if (chatResponse && chatResponse->success) {
            // 添加新聊天到本地列表
            std::lock_guard<std::mutex> lock(m_dataMutex);
            m_chats.push_back(chatResponse->chat);
            
            if (m_onChatsUpdated) {
                m_onChatsUpdated(m_chats);
            }
            
            ChatResponseFree(chatResponse);
            return true;
        }
        ChatResponseFree(chatResponse);
    }
    
    return false;
}

bool ChatManager::CreateChatAsync(const CreateChatRequest& request) {
    std::thread([this, request]() {
        CreateChat(request);
    }).detach();
    
    return true;
}

bool ChatManager::DeleteChat(const std::string& chatId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/chats/" + chatId);
    HttpRequest request = CreateAuthenticatedRequest(url, "DELETE");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        // 从本地列表中移除聊天
        std::lock_guard<std::mutex> lock(m_dataMutex);
        m_chats.erase(std::remove_if(m_chats.begin(), m_chats.end(),
            [&chatId](const Chat& chat) { return chat.id == chatId; }), m_chats.end());
        
        // 移除聊天消息
        m_chatMessages.erase(chatId);
        
        if (m_onChatsUpdated) {
            m_onChatsUpdated(m_chats);
        }
        
        return true;
    }
    
    return false;
}

bool ChatManager::LeaveChat(const std::string& chatId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/chats/" + chatId + "/leave");
    HttpRequest request = CreateAuthenticatedRequest(url, "POST");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    return response.statusCode == 200;
}

bool ChatManager::AddParticipant(const std::string& chatId, const std::string& userId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty() || userId.empty()) {
        return false;
    }
    
    Json::Value json;
    json["user_id"] = userId;
    
    Json::StreamWriterBuilder builder;
    std::string jsonStr = Json::writeString(builder, json);
    
    std::string url = BuildUrl("/api/chats/" + chatId + "/participants");
    HttpRequest request = CreateAuthenticatedRequest(url, "POST");
    request.body = jsonStr;
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    return response.statusCode == 200;
}

bool ChatManager::RemoveParticipant(const std::string& chatId, const std::string& userId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty() || userId.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/chats/" + chatId + "/participants/" + userId);
    HttpRequest request = CreateAuthenticatedRequest(url, "DELETE");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    return response.statusCode == 200;
}

// WebSocket相关实现
bool ChatManager::ConnectWebSocket() {
    if (m_webSocketConnected) {
        return true;
    }
    
    m_shouldStopWebSocket = false;
    m_webSocketThread = std::thread(&ChatManager::WebSocketThreadFunc, this);
    
    return true;
}

void ChatManager::DisconnectWebSocket() {
    m_shouldStopWebSocket = true;
    m_webSocketConnected = false;
    
    if (m_webSocketHandle) {
        WinHttpCloseHandle(m_webSocketHandle);
        m_webSocketHandle = nullptr;
    }
    
    if (m_webSocketThread.joinable()) {
        m_webSocketThread.join();
    }
}

bool ChatManager::IsWebSocketConnected() const {
    return m_webSocketConnected;
}

bool ChatManager::SendTypingStatus(const std::string& chatId, bool isTyping) {
    if (!m_webSocketConnected || chatId.empty()) {
        return false;
    }
    
    Json::Value json;
    json["type"] = "typing";
    json["chat_id"] = chatId;
    json["is_typing"] = isTyping;
    
    Json::StreamWriterBuilder builder;
    std::string jsonStr = Json::writeString(builder, json);
    
    // TODO: 实现WebSocket发送逻辑
    return true;
}

// 文件传输
bool ChatManager::SendFile(const std::string& chatId, const std::string& filePath) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || chatId.empty() || filePath.empty()) {
        return false;
    }
    
    // TODO: 实现文件上传逻辑
    return false;
}

bool ChatManager::DownloadFile(const std::string& fileUrl, const std::string& savePath) {
    if (fileUrl.empty() || savePath.empty()) {
        return false;
    }
    
    // TODO: 实现文件下载逻辑
    return false;
}

// 搜索功能
bool ChatManager::SearchMessages(const std::string& query, const std::string& chatId) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || query.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/messages/search?q=" + query);
    if (!chatId.empty()) {
        url += "&chat_id=" + chatId;
    }
    
    HttpRequest request = CreateAuthenticatedRequest(url, "GET");
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        ChatResponse* chatResponse = ParseChatResponse(response.body);
        if (chatResponse && chatResponse->success) {
            if (m_onMessagesUpdated) {
                m_onMessagesUpdated(chatResponse->messages);
            }
            
            ChatResponseFree(chatResponse);
            return true;
        }
        ChatResponseFree(chatResponse);
    }
    
    return false;
}

bool ChatManager::SearchChats(const std::string& query) {
    if (!m_initialized || !g_authManager || !g_authManager->IsLoggedIn() || query.empty()) {
        return false;
    }
    
    std::string url = BuildUrl("/api/chats/search?q=" + query);
    HttpRequest request = CreateAuthenticatedRequest(url, "GET");
    
    HttpResponse response = m_httpClient->SendRequest(request);
    
    if (response.statusCode == 200) {
        ChatResponse* chatResponse = ParseChatResponse(response.body);
        if (chatResponse && chatResponse->success) {
            if (m_onChatsUpdated) {
                m_onChatsUpdated(chatResponse->chats);
            }
            
            ChatResponseFree(chatResponse);
            return true;
        }
        ChatResponseFree(chatResponse);
    }
    
    return false;
}

// 本地数据管理
Chat* ChatManager::FindChatById(const std::string& chatId) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    for (auto& chat : m_chats) {
        if (chat.id == chatId) {
            return &chat;
        }
    }
    return nullptr;
}

Chat* ChatManager::FindChatByParticipant(const std::string& userId) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    for (auto& chat : m_chats) {
        if (chat.type == ChatType::PRIVATE) {
            for (const auto& participant : chat.participants) {
                if (participant == userId) {
                    return &chat;
                }
            }
        }
    }
    return nullptr;
}

Message* ChatManager::FindMessageById(const std::string& messageId) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    for (auto& chatPair : m_chatMessages) {
        for (auto& message : chatPair.second) {
            if (message.id == messageId) {
                return &message;
            }
        }
    }
    return nullptr;
}

std::vector<Message> ChatManager::GetChatMessages(const std::string& chatId) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    auto it = m_chatMessages.find(chatId);
    if (it != m_chatMessages.end()) {
        return it->second;
    }
    return std::vector<Message>();
}

// 事件回调设置
void ChatManager::SetOnChatsUpdatedCallback(OnChatsUpdatedCallback callback) {
    m_onChatsUpdated = callback;
}

void ChatManager::SetOnMessagesUpdatedCallback(OnMessagesUpdatedCallback callback) {
    m_onMessagesUpdated = callback;
}

void ChatManager::SetOnNewMessageCallback(OnNewMessageCallback callback) {
    m_onNewMessage = callback;
}

void ChatManager::SetOnMessageStatusUpdatedCallback(OnMessageStatusUpdatedCallback callback) {
    m_onMessageStatusUpdated = callback;
}

void ChatManager::SetOnTypingStatusCallback(OnTypingStatusCallback callback) {
    m_onTypingStatus = callback;
}

void ChatManager::SetOnErrorCallback(OnErrorCallback callback) {
    m_onError = callback;
}

// 配置方法
void ChatManager::SetServerUrl(const std::string& url) {
    m_serverUrl = url;
}

void ChatManager::SetWebSocketUrl(const std::string& url) {
    m_webSocketUrl = url;
}

void ChatManager::EnableAutoReconnect(bool enable) {
    m_autoReconnectEnabled = enable;
    if (enable) {
        StartAutoReconnect();
    } else {
        StopAutoReconnect();
    }
}

void ChatManager::SetReconnectInterval(int seconds) {
    m_reconnectInterval = seconds;
}

// 状态查询
bool ChatManager::IsInitialized() const {
    return m_initialized;
}

int ChatManager::GetUnreadMessageCount() const {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    int totalUnread = 0;
    for (const auto& chat : m_chats) {
        totalUnread += chat.unreadCount;
    }
    return totalUnread;
}

int ChatManager::GetChatCount() const {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    return static_cast<int>(m_chats.size());
}

// 内部辅助方法实现
std::string ChatManager::BuildUrl(const std::string& endpoint) {
    return m_serverUrl + endpoint;
}

std::string ChatManager::BuildWebSocketUrl(const std::string& endpoint) {
    return m_webSocketUrl + endpoint;
}

HttpRequest ChatManager::CreateAuthenticatedRequest(const std::string& url, const std::string& method) {
    HttpRequest request;
    request.url = url;
    request.method = method;
    request.headers["Content-Type"] = "application/json";
    
    if (g_authManager && g_authManager->IsLoggedIn()) {
        std::string token = g_authManager->GetAccessToken();
        if (!token.empty()) {
            request.headers["Authorization"] = "Bearer " + token;
        }
    }
    
    return request;
}

ChatResponse* ChatManager::ParseChatResponse(const std::string& jsonStr) {
    // TODO: 实现JSON解析逻辑
    return nullptr;
}

std::string ChatManager::CreateSendMessageJson(const SendMessageRequest& request) {
    Json::Value json;
    json["chat_id"] = request.chatId;
    json["content"] = request.content;
    json["type"] = MessageTypeToString(request.type);
    
    if (!request.replyToMessageId.empty()) {
        json["reply_to_message_id"] = request.replyToMessageId;
    }
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, json);
}

std::string ChatManager::CreateCreateChatJson(const CreateChatRequest& request) {
    Json::Value json;
    json["type"] = ChatTypeToString(request.type);
    json["name"] = request.name;
    
    Json::Value participants(Json::arrayValue);
    for (const auto& participant : request.participants) {
        participants.append(participant);
    }
    json["participants"] = participants;
    
    Json::StreamWriterBuilder builder;
    return Json::writeString(builder, json);
}

// WebSocket线程函数
void ChatManager::WebSocketThreadFunc() {
    // TODO: 实现WebSocket连接和消息处理逻辑
}

void ChatManager::HandleWebSocketMessage(const WebSocketMessage& wsMessage) {
    // TODO: 实现WebSocket消息处理逻辑
}

// 数据同步方法
void ChatManager::UpdateLocalChats(const std::vector<Chat>& chats) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    m_chats = chats;
}

void ChatManager::UpdateLocalMessages(const std::string& chatId, const std::vector<Message>& messages) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    m_chatMessages[chatId] = messages;
}

void ChatManager::AddNewMessage(const Message& message) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    m_chatMessages[message.chatId].push_back(message);
    
    // 更新聊天的最后消息信息
    for (auto& chat : m_chats) {
        if (chat.id == message.chatId) {
            chat.lastMessage = message.content;
            chat.lastMessageTime = message.timestamp;
            chat.unreadCount++;
            break;
        }
    }
}

void ChatManager::UpdateMessageStatus(const std::string& messageId, MessageStatus status) {
    std::lock_guard<std::mutex> lock(m_dataMutex);
    for (auto& chatPair : m_chatMessages) {
        for (auto& message : chatPair.second) {
            if (message.id == messageId) {
                message.status = status;
                
                if (m_onMessageStatusUpdated) {
                    m_onMessageStatusUpdated(message);
                }
                return;
            }
        }
    }
}

// 自动重连实现
void ChatManager::StartAutoReconnect() {
    if (!m_autoReconnectEnabled) return;
    
    m_shouldStopReconnect = false;
    m_reconnectThread = std::thread(&ChatManager::ReconnectThreadFunc, this);
}

void ChatManager::StopAutoReconnect() {
    m_shouldStopReconnect = true;
    if (m_reconnectThread.joinable()) {
        m_reconnectThread.join();
    }
}

void ChatManager::ReconnectThreadFunc() {
    while (!m_shouldStopReconnect) {
        if (!m_webSocketConnected) {
            ConnectWebSocket();
        }
        
        std::this_thread::sleep_for(std::chrono::seconds(m_reconnectInterval));
    }
}

// 全局函数实现
bool InitChatManager() {
    if (g_chatManager != nullptr) {
        return true;
    }
    
    g_chatManager = new ChatManager();
    return g_chatManager->Initialize();
}

void CleanupChatManager() {
    if (g_chatManager != nullptr) {
        g_chatManager->Cleanup();
        delete g_chatManager;
        g_chatManager = nullptr;
    }
}

ChatManager* GetChatManager() {
    return g_chatManager;
}

// 辅助函数实现
const char* MessageTypeToString(MessageType type) {
    switch (type) {
        case MessageType::TEXT: return "text";
        case MessageType::IMAGE: return "image";
        case MessageType::FILE: return "file";
        case MessageType::SYSTEM: return "system";
        case MessageType::EMOJI: return "emoji";
        default: return "text";
    }
}

MessageType StringToMessageType(const char* type) {
    if (!type) return MessageType::TEXT;
    
    std::string typeStr(type);
    if (typeStr == "text") return MessageType::TEXT;
    if (typeStr == "image") return MessageType::IMAGE;
    if (typeStr == "file") return MessageType::FILE;
    if (typeStr == "system") return MessageType::SYSTEM;
    if (typeStr == "emoji") return MessageType::EMOJI;
    
    return MessageType::TEXT;
}

const char* ChatTypeToString(ChatType type) {
    switch (type) {
        case ChatType::PRIVATE: return "private";
        case ChatType::GROUP: return "group";
        default: return "private";
    }
}

ChatType StringToChatType(const char* type) {
    if (!type) return ChatType::PRIVATE;
    
    std::string typeStr(type);
    if (typeStr == "private") return ChatType::PRIVATE;
    if (typeStr == "group") return ChatType::GROUP;
    
    return ChatType::PRIVATE;
}

const char* MessageStatusToString(MessageStatus status) {
    switch (status) {
        case MessageStatus::SENDING: return "sending";
        case MessageStatus::SENT: return "sent";
        case MessageStatus::DELIVERED: return "delivered";
        case MessageStatus::READ: return "read";
        case MessageStatus::FAILED: return "failed";
        default: return "sending";
    }
}

MessageStatus StringToMessageStatus(const char* status) {
    if (!status) return MessageStatus::SENDING;
    
    std::string statusStr(status);
    if (statusStr == "sending") return MessageStatus::SENDING;
    if (statusStr == "sent") return MessageStatus::SENT;
    if (statusStr == "delivered") return MessageStatus::DELIVERED;
    if (statusStr == "read") return MessageStatus::READ;
    if (statusStr == "failed") return MessageStatus::FAILED;
    
    return MessageStatus::SENDING;
}

// 内存管理函数实现
void MessageFree(Message* message) {
    if (message) {
        // C++对象会自动清理std::string成员
    }
}

void ChatFree(Chat* chat) {
    if (chat) {
        // C++对象会自动清理std::string和std::vector成员
    }
}

void ChatResponseFree(ChatResponse* response) {
    if (response) {
        delete response;
    }
}

void WebSocketMessageFree(WebSocketMessage* wsMessage) {
    if (wsMessage) {
        delete wsMessage;
    }
}