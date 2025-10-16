import Foundation
import Network
import Combine

// MARK: - Enums
enum MessageType: String, CaseIterable, Codable {
    case text = "text"
    case image = "image"
    case file = "file"
    case audio = "audio"
    case video = "video"
    case system = "system"
}

enum ChatType: String, CaseIterable, Codable {
    case private = "private"
    case group = "group"
}

enum MessageStatus: String, CaseIterable, Codable {
    case sending = "sending"
    case sent = "sent"
    case delivered = "delivered"
    case read = "read"
    case failed = "failed"
}

// MARK: - Data Models
struct Message: Identifiable, Codable, Equatable {
    let id: String
    let chatId: String
    let senderId: String
    let senderUsername: String
    let content: String
    let type: MessageType
    let status: MessageStatus
    let timestamp: String
    let createdAt: String
    let updatedAt: String
    let fileName: String?
    let fileUrl: String?
    let fileSize: Int64?
    let replyToMessageId: String?
    let replyToContent: String?
    
    enum CodingKeys: String, CodingKey {
        case id, content, type, status, timestamp
        case chatId = "chat_id"
        case senderId = "sender_id"
        case senderUsername = "sender_username"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case fileName = "file_name"
        case fileUrl = "file_url"
        case fileSize = "file_size"
        case replyToMessageId = "reply_to_message_id"
        case replyToContent = "reply_to_content"
    }
}

struct Chat: Identifiable, Codable, Equatable {
    let id: String
    let name: String?
    let type: ChatType
    let avatarUrl: String?
    let lastMessage: String?
    let lastMessageTime: String?
    let unreadCount: Int
    let participants: [String]
    let createdAt: String
    let updatedAt: String
    
    enum CodingKeys: String, CodingKey {
        case id, name, type, participants
        case avatarUrl = "avatar_url"
        case lastMessage = "last_message"
        case lastMessageTime = "last_message_time"
        case unreadCount = "unread_count"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct SendMessageRequest: Codable {
    let chatId: String
    let content: String
    let type: MessageType
    let replyToMessageId: String?
    
    enum CodingKeys: String, CodingKey {
        case content, type
        case chatId = "chat_id"
        case replyToMessageId = "reply_to_message_id"
    }
}

struct CreateChatRequest: Codable {
    let type: ChatType
    let name: String?
    let participants: [String]
}

struct ChatResponse<T: Codable>: Codable {
    let success: Bool
    let message: String?
    let data: T?
    let error: String?
}

struct WebSocketMessage: Codable {
    let type: String
    let chatId: String?
    let data: String
    
    enum CodingKeys: String, CodingKey {
        case type, data
        case chatId = "chat_id"
    }
}

struct TypingStatusData: Codable {
    let userId: String
    let isTyping: Bool
    
    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case isTyping = "is_typing"
    }
}

// MARK: - Error Types
enum ChatManagerError: Error, LocalizedError {
    case notInitialized
    case notAuthenticated
    case networkError(String)
    case invalidResponse
    case websocketError(String)
    case fileTransferError(String)
    
    var errorDescription: String? {
        switch self {
        case .notInitialized:
            return "Chat manager not initialized"
        case .notAuthenticated:
            return "User not authenticated"
        case .networkError(let message):
            return "Network error: \(message)"
        case .invalidResponse:
            return "Invalid response from server"
        case .websocketError(let message):
            return "WebSocket error: \(message)"
        case .fileTransferError(let message):
            return "File transfer error: \(message)"
        }
    }
}

// MARK: - Delegate Protocol
protocol ChatManagerDelegate: AnyObject {
    func chatManager(_ manager: ChatManager, didReceiveNewMessage message: Message)
    func chatManager(_ manager: ChatManager, didUpdateMessageStatus message: Message)
    func chatManager(_ manager: ChatManager, didReceiveTypingStatus chatId: String, userId: String, isTyping: Bool)
    func chatManager(_ manager: ChatManager, didUpdateChatList chats: [Chat])
    func chatManager(_ manager: ChatManager, didConnectWebSocket: Void)
    func chatManager(_ manager: ChatManager, didDisconnectWebSocket error: Error?)
    func chatManager(_ manager: ChatManager, didEncounterError error: Error)
}

// MARK: - Chat Manager
class ChatManager: ObservableObject {
    // MARK: - Properties
    @Published var chats: [Chat] = []
    @Published var messages: [String: [Message]] = [:]
    @Published var isLoading = false
    @Published var isWebSocketConnected = false
    @Published var typingUsers: [String: Set<String>] = [:]
    
    weak var delegate: ChatManagerDelegate?
    
    private let httpClient: HttpClient
    private let authManager: AuthManager
    private var webSocketTask: URLSessionWebSocketTask?
    private var networkMonitor: NWPathMonitor?
    private var monitorQueue = DispatchQueue(label: "NetworkMonitor")
    private var reconnectTimer: Timer?
    private var heartbeatTimer: Timer?
    
    // Configuration
    private var baseUrl: String
    private var websocketUrl: String
    private var autoReconnectEnabled = true
    private var reconnectInterval: TimeInterval = 5.0
    private var heartbeatInterval: TimeInterval = 30.0
    private var maxReconnectAttempts = 10
    private var currentReconnectAttempts = 0
    
    // MARK: - Initialization
    init(httpClient: HttpClient, authManager: AuthManager, baseUrl: String = "http://localhost:8080") {
        self.httpClient = httpClient
        self.authManager = authManager
        self.baseUrl = baseUrl
        self.websocketUrl = baseUrl.replacingOccurrences(of: "http", with: "ws") + "/ws/chat"
        
        setupNetworkMonitoring()
        connectWebSocket()
    }
    
    deinit {
        disconnectWebSocket()
        networkMonitor?.cancel()
        reconnectTimer?.invalidate()
        heartbeatTimer?.invalidate()
    }
    
    // MARK: - Network Monitoring
    private func setupNetworkMonitoring() {
        networkMonitor = NWPathMonitor()
        networkMonitor?.pathUpdateHandler = { [weak self] path in
            DispatchQueue.main.async {
                if path.status == .satisfied && !self?.isWebSocketConnected ?? false {
                    self?.connectWebSocket()
                }
            }
        }
        networkMonitor?.start(queue: monitorQueue)
    }
    
    // MARK: - Chat List Management
    func getChatList() async throws -> [Chat] {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let request = HttpRequest(
                url: "\(baseUrl)/api/chats",
                method: .GET,
                headers: await getAuthHeaders()
            )
            
            let response: ChatResponse<[Chat]> = try await httpClient.request(request)
            
            if response.success, let chats = response.data {
                await MainActor.run {
                    self.chats = chats
                }
                delegate?.chatManager(self, didUpdateChatList: chats)
                return chats
            } else {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    func createChat(_ request: CreateChatRequest) async throws -> Chat {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let httpRequest = HttpRequest(
                url: "\(baseUrl)/api/chats",
                method: .POST,
                headers: await getAuthHeaders(),
                body: try JSONEncoder().encode(request)
            )
            
            let response: ChatResponse<Chat> = try await httpClient.request(httpRequest)
            
            if response.success, let chat = response.data {
                await MainActor.run {
                    self.chats.append(chat)
                }
                return chat
            } else {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    func deleteChat(_ chatId: String) async throws {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let request = HttpRequest(
                url: "\(baseUrl)/api/chats/\(chatId)",
                method: .DELETE,
                headers: await getAuthHeaders()
            )
            
            let response: ChatResponse<String> = try await httpClient.request(request)
            
            if response.success {
                await MainActor.run {
                    self.chats.removeAll { $0.id == chatId }
                    self.messages.removeValue(forKey: chatId)
                }
            } else {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    func leaveChat(_ chatId: String) async throws {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let request = HttpRequest(
                url: "\(baseUrl)/api/chats/\(chatId)/leave",
                method: .POST,
                headers: await getAuthHeaders()
            )
            
            let response: ChatResponse<String> = try await httpClient.request(request)
            
            if response.success {
                await MainActor.run {
                    self.chats.removeAll { $0.id == chatId }
                    self.messages.removeValue(forKey: chatId)
                }
            } else {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    // MARK: - Message Management
    func getMessages(for chatId: String, page: Int = 1, limit: Int = 50) async throws -> [Message] {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let request = HttpRequest(
                url: "\(baseUrl)/api/chats/\(chatId)/messages?page=\(page)&limit=\(limit)",
                method: .GET,
                headers: await getAuthHeaders()
            )
            
            let response: ChatResponse<[Message]> = try await httpClient.request(request)
            
            if response.success, let messages = response.data {
                await MainActor.run {
                    if page == 1 {
                        self.messages[chatId] = messages
                    } else {
                        var existingMessages = self.messages[chatId] ?? []
                        existingMessages.append(contentsOf: messages)
                        self.messages[chatId] = existingMessages
                    }
                }
                return messages
            } else {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    func sendMessage(_ request: SendMessageRequest) async throws -> Message {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        // Create temporary message for immediate UI update
        let tempMessage = Message(
            id: UUID().uuidString,
            chatId: request.chatId,
            senderId: authManager.currentUser?.id ?? "",
            senderUsername: authManager.currentUser?.username ?? "",
            content: request.content,
            type: request.type,
            status: .sending,
            timestamp: ISO8601DateFormatter().string(from: Date()),
            createdAt: ISO8601DateFormatter().string(from: Date()),
            updatedAt: ISO8601DateFormatter().string(from: Date()),
            fileName: nil,
            fileUrl: nil,
            fileSize: nil,
            replyToMessageId: request.replyToMessageId,
            replyToContent: nil
        )
        
        await MainActor.run {
            var chatMessages = self.messages[request.chatId] ?? []
            chatMessages.append(tempMessage)
            self.messages[request.chatId] = chatMessages
        }
        
        do {
            let httpRequest = HttpRequest(
                url: "\(baseUrl)/api/chats/\(request.chatId)/messages",
                method: .POST,
                headers: await getAuthHeaders(),
                body: try JSONEncoder().encode(request)
            )
            
            let response: ChatResponse<Message> = try await httpClient.request(httpRequest)
            
            if response.success, let message = response.data {
                await MainActor.run {
                    // Replace temporary message with actual message
                    if var chatMessages = self.messages[request.chatId] {
                        if let index = chatMessages.firstIndex(where: { $0.id == tempMessage.id }) {
                            chatMessages[index] = message
                            self.messages[request.chatId] = chatMessages
                        }
                    }
                }
                return message
            } else {
                // Mark message as failed
                await MainActor.run {
                    if var chatMessages = self.messages[request.chatId] {
                        if let index = chatMessages.firstIndex(where: { $0.id == tempMessage.id }) {
                            var failedMessage = chatMessages[index]
                            chatMessages[index] = Message(
                                id: failedMessage.id,
                                chatId: failedMessage.chatId,
                                senderId: failedMessage.senderId,
                                senderUsername: failedMessage.senderUsername,
                                content: failedMessage.content,
                                type: failedMessage.type,
                                status: .failed,
                                timestamp: failedMessage.timestamp,
                                createdAt: failedMessage.createdAt,
                                updatedAt: failedMessage.updatedAt,
                                fileName: failedMessage.fileName,
                                fileUrl: failedMessage.fileUrl,
                                fileSize: failedMessage.fileSize,
                                replyToMessageId: failedMessage.replyToMessageId,
                                replyToContent: failedMessage.replyToContent
                            )
                            self.messages[request.chatId] = chatMessages
                        }
                    }
                }
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    func markMessagesAsRead(chatId: String, messageIds: [String]) async throws {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        do {
            let requestBody = ["message_ids": messageIds]
            let httpRequest = HttpRequest(
                url: "\(baseUrl)/api/chats/\(chatId)/messages/read",
                method: .POST,
                headers: await getAuthHeaders(),
                body: try JSONEncoder().encode(requestBody)
            )
            
            let response: ChatResponse<String> = try await httpClient.request(httpRequest)
            
            if !response.success {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    // MARK: - Chat Session Management
    func addParticipant(chatId: String, userId: String) async throws {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        do {
            let requestBody = ["user_id": userId]
            let httpRequest = HttpRequest(
                url: "\(baseUrl)/api/chats/\(chatId)/participants",
                method: .POST,
                headers: await getAuthHeaders(),
                body: try JSONEncoder().encode(requestBody)
            )
            
            let response: ChatResponse<String> = try await httpClient.request(httpRequest)
            
            if !response.success {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    func removeParticipant(chatId: String, userId: String) async throws {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        do {
            let httpRequest = HttpRequest(
                url: "\(baseUrl)/api/chats/\(chatId)/participants/\(userId)",
                method: .DELETE,
                headers: await getAuthHeaders()
            )
            
            let response: ChatResponse<String> = try await httpClient.request(httpRequest)
            
            if !response.success {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    // MARK: - Real-time Features (WebSocket)
    func connectWebSocket() {
        guard authManager.isAuthenticated else { return }
        
        disconnectWebSocket()
        
        guard let url = URL(string: websocketUrl) else { return }
        
        var request = URLRequest(url: url)
        if let token = authManager.accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        webSocketTask = URLSession.shared.webSocketTask(with: request)
        webSocketTask?.resume()
        
        isWebSocketConnected = true
        currentReconnectAttempts = 0
        delegate?.chatManager(self, didConnectWebSocket: ())
        
        startHeartbeat()
        receiveWebSocketMessage()
    }
    
    func disconnectWebSocket() {
        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        isWebSocketConnected = false
        stopHeartbeat()
        reconnectTimer?.invalidate()
    }
    
    private func receiveWebSocketMessage() {
        webSocketTask?.receive { [weak self] result in
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    self?.handleWebSocketMessage(text)
                case .data(let data):
                    if let text = String(data: data, encoding: .utf8) {
                        self?.handleWebSocketMessage(text)
                    }
                @unknown default:
                    break
                }
                self?.receiveWebSocketMessage()
                
            case .failure(let error):
                DispatchQueue.main.async {
                    self?.isWebSocketConnected = false
                    self?.delegate?.chatManager(self!, didDisconnectWebSocket: error)
                    
                    if self?.autoReconnectEnabled == true {
                        self?.scheduleReconnect()
                    }
                }
            }
        }
    }
    
    private func handleWebSocketMessage(_ message: String) {
        guard let data = message.data(using: .utf8),
              let wsMessage = try? JSONDecoder().decode(WebSocketMessage.self, from: data) else {
            return
        }
        
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            switch wsMessage.type {
            case "new_message":
                if let messageData = wsMessage.data.data(using: .utf8),
                   let newMessage = try? JSONDecoder().decode(Message.self, from: messageData) {
                    
                    var chatMessages = self.messages[newMessage.chatId] ?? []
                    chatMessages.append(newMessage)
                    self.messages[newMessage.chatId] = chatMessages
                    
                    self.delegate?.chatManager(self, didReceiveNewMessage: newMessage)
                }
                
            case "message_status_updated":
                if let messageData = wsMessage.data.data(using: .utf8),
                   let updatedMessage = try? JSONDecoder().decode(Message.self, from: messageData) {
                    
                    if var chatMessages = self.messages[updatedMessage.chatId] {
                        if let index = chatMessages.firstIndex(where: { $0.id == updatedMessage.id }) {
                            chatMessages[index] = updatedMessage
                            self.messages[updatedMessage.chatId] = chatMessages
                        }
                    }
                    
                    self.delegate?.chatManager(self, didUpdateMessageStatus: updatedMessage)
                }
                
            case "typing_status":
                if let chatId = wsMessage.chatId,
                   let typingData = wsMessage.data.data(using: .utf8),
                   let typingStatus = try? JSONDecoder().decode(TypingStatusData.self, from: typingData) {
                    
                    var chatTypingUsers = self.typingUsers[chatId] ?? Set<String>()
                    
                    if typingStatus.isTyping {
                        chatTypingUsers.insert(typingStatus.userId)
                    } else {
                        chatTypingUsers.remove(typingStatus.userId)
                    }
                    
                    self.typingUsers[chatId] = chatTypingUsers
                    
                    self.delegate?.chatManager(self, didReceiveTypingStatus: chatId, userId: typingStatus.userId, isTyping: typingStatus.isTyping)
                }
                
            default:
                break
            }
        }
    }
    
    func sendTypingStatus(chatId: String, isTyping: Bool) {
        guard isWebSocketConnected else { return }
        
        let typingData = TypingStatusData(userId: authManager.currentUser?.id ?? "", isTyping: isTyping)
        let wsMessage = WebSocketMessage(type: "typing_status", chatId: chatId, data: "")
        
        if let messageData = try? JSONEncoder().encode(wsMessage),
           let messageString = String(data: messageData, encoding: .utf8) {
            webSocketTask?.send(.string(messageString)) { _ in }
        }
    }
    
    private func startHeartbeat() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: heartbeatInterval, repeats: true) { [weak self] _ in
            self?.sendHeartbeat()
        }
    }
    
    private func stopHeartbeat() {
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
    }
    
    private func sendHeartbeat() {
        guard isWebSocketConnected else { return }
        
        webSocketTask?.sendPing { [weak self] error in
            if let error = error {
                DispatchQueue.main.async {
                    self?.isWebSocketConnected = false
                    self?.delegate?.chatManager(self!, didDisconnectWebSocket: error)
                    
                    if self?.autoReconnectEnabled == true {
                        self?.scheduleReconnect()
                    }
                }
            }
        }
    }
    
    private func scheduleReconnect() {
        guard currentReconnectAttempts < maxReconnectAttempts else { return }
        
        currentReconnectAttempts += 1
        
        reconnectTimer = Timer.scheduledTimer(withTimeInterval: reconnectInterval, repeats: false) { [weak self] _ in
            self?.connectWebSocket()
        }
    }
    
    // MARK: - File Transfer
    func uploadFile(chatId: String, fileUrl: URL, fileName: String) async throws -> Message {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        // TODO: Implement file upload functionality
        // This would involve multipart form data upload
        throw ChatManagerError.fileTransferError("File upload not implemented yet")
    }
    
    func downloadFile(fileUrl: String, destinationUrl: URL) async throws {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        // TODO: Implement file download functionality
        throw ChatManagerError.fileTransferError("File download not implemented yet")
    }
    
    // MARK: - Search
    func searchMessages(query: String, chatId: String? = nil) async throws -> [Message] {
        guard authManager.isAuthenticated else {
            throw ChatManagerError.notAuthenticated
        }
        
        var urlString = "\(baseUrl)/api/messages/search?q=\(query.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? "")"
        if let chatId = chatId {
            urlString += "&chat_id=\(chatId)"
        }
        
        do {
            let request = HttpRequest(
                url: urlString,
                method: .GET,
                headers: await getAuthHeaders()
            )
            
            let response: ChatResponse<[Message]> = try await httpClient.request(request)
            
            if response.success, let messages = response.data {
                return messages
            } else {
                throw ChatManagerError.networkError(response.error ?? "Unknown error")
            }
        } catch {
            delegate?.chatManager(self, didEncounterError: error)
            throw error
        }
    }
    
    // MARK: - Configuration
    func setAutoReconnect(_ enabled: Bool) {
        autoReconnectEnabled = enabled
    }
    
    func setReconnectInterval(_ interval: TimeInterval) {
        reconnectInterval = interval
    }
    
    func setHeartbeatInterval(_ interval: TimeInterval) {
        heartbeatInterval = interval
        
        if isWebSocketConnected {
            stopHeartbeat()
            startHeartbeat()
        }
    }
    
    // MARK: - Status Queries
    func isConnected() -> Bool {
        return isWebSocketConnected
    }
    
    func getConnectionStatus() -> String {
        return isWebSocketConnected ? "Connected" : "Disconnected"
    }
    
    // MARK: - Helper Methods
    private func getAuthHeaders() async -> [String: String] {
        var headers = [
            "Content-Type": "application/json",
            "Accept": "application/json"
        ]
        
        if let token = authManager.accessToken {
            headers["Authorization"] = "Bearer \(token)"
        }
        
        return headers
    }
}

// MARK: - Factory
class ChatManagerFactory {
    static let shared = ChatManagerFactory()
    private var chatManager: ChatManager?
    
    private init() {}
    
    func getChatManager() -> ChatManager? {
        return chatManager
    }
    
    func createChatManager(httpClient: HttpClient, authManager: AuthManager, baseUrl: String = "http://localhost:8080") -> ChatManager {
        let manager = ChatManager(httpClient: httpClient, authManager: authManager, baseUrl: baseUrl)
        self.chatManager = manager
        return manager
    }
    
    func destroyChatManager() {
        chatManager = nil
    }
}