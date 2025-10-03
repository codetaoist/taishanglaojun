import Foundation
import Network

// MARK: - 枚举定义
public enum FriendStatus: String, CaseIterable, Codable {
    case pending = "pending"
    case accepted = "accepted"
    case blocked = "blocked"
    case declined = "declined"
}

public enum OnlineStatus: String, CaseIterable, Codable {
    case online = "online"
    case offline = "offline"
    case away = "away"
    case busy = "busy"
}

// MARK: - 数据模型
public struct Friend: Codable, Identifiable {
    public let id: String
    public let username: String
    public let email: String?
    public let avatarUrl: String?
    public let status: FriendStatus
    public let onlineStatus: OnlineStatus
    public let lastSeen: String?
    public let createdAt: String
    public let updatedAt: String
    
    enum CodingKeys: String, CodingKey {
        case id, username, email, status, createdAt, updatedAt
        case avatarUrl = "avatar_url"
        case onlineStatus = "online_status"
        case lastSeen = "last_seen"
    }
}

public struct FriendRequest: Codable, Identifiable {
    public let id: String
    public let fromUserId: String
    public let toUserId: String
    public let fromUsername: String
    public let toUsername: String
    public let message: String?
    public let status: FriendStatus
    public let createdAt: String
    public let updatedAt: String
    
    enum CodingKeys: String, CodingKey {
        case id, message, status, createdAt, updatedAt
        case fromUserId = "from_user_id"
        case toUserId = "to_user_id"
        case fromUsername = "from_username"
        case toUsername = "to_username"
    }
}

public struct AddFriendRequest: Codable {
    public let username: String
    public let message: String?
    
    public init(username: String, message: String? = nil) {
        self.username = username
        self.message = message
    }
}

public struct RespondToRequestData: Codable {
    public let action: String
    
    public init(accept: Bool) {
        self.action = accept ? "accept" : "decline"
    }
}

public struct UpdateStatusRequest: Codable {
    public let status: String
    
    public init(status: OnlineStatus) {
        self.status = status.rawValue
    }
}

public struct FriendResponse: Codable {
    public let success: Bool
    public let message: String?
    public let friends: [Friend]?
    public let requests: [FriendRequest]?
    
    public init(success: Bool, message: String? = nil, friends: [Friend]? = nil, requests: [FriendRequest]? = nil) {
        self.success = success
        self.message = message
        self.friends = friends
        self.requests = requests
    }
}

// MARK: - 错误定义
public enum FriendManagerError: Error, LocalizedError {
    case notAuthenticated
    case networkError(Error)
    case invalidResponse
    case serverError(String)
    case friendNotFound
    case requestNotFound
    
    public var errorDescription: String? {
        switch self {
        case .notAuthenticated:
            return "用户未认证"
        case .networkError(let error):
            return "网络错误: \(error.localizedDescription)"
        case .invalidResponse:
            return "无效的服务器响应"
        case .serverError(let message):
            return "服务器错误: \(message)"
        case .friendNotFound:
            return "好友未找到"
        case .requestNotFound:
            return "好友请求未找到"
        }
    }
}

// MARK: - 事件回调协议
public protocol FriendManagerDelegate: AnyObject {
    func friendManager(_ manager: FriendManager, didUpdateFriends friends: [Friend])
    func friendManager(_ manager: FriendManager, didUpdateRequests requests: [FriendRequest])
    func friendManager(_ manager: FriendManager, didReceiveNewRequest request: FriendRequest)
    func friendManager(_ manager: FriendManager, didUpdateFriendStatus friend: Friend)
    func friendManager(_ manager: FriendManager, didEncounterError error: FriendManagerError)
}

// MARK: - 好友管理器
@MainActor
public class FriendManager: ObservableObject {
    // MARK: - 属性
    private let httpClient: HttpClient
    private let authManager: AuthManager
    
    @Published public private(set) var friends: [Friend] = []
    @Published public private(set) var pendingRequests: [FriendRequest] = []
    @Published public private(set) var currentOnlineStatus: OnlineStatus = .offline
    @Published public private(set) var isLoading = false
    
    public weak var delegate: FriendManagerDelegate?
    
    private var serverUrl: String
    private var autoRefreshEnabled = true
    private var refreshInterval: TimeInterval = 30.0
    private var refreshTimer: Timer?
    private var networkMonitor: NWPathMonitor?
    
    // MARK: - 初始化
    public init(httpClient: HttpClient, authManager: AuthManager, serverUrl: String = "http://localhost:8081") {
        self.httpClient = httpClient
        self.authManager = authManager
        self.serverUrl = serverUrl
        
        setupNetworkMonitoring()
        startAutoRefresh()
    }
    
    deinit {
        stopAutoRefresh()
        networkMonitor?.cancel()
    }
    
    // MARK: - 网络监控
    private func setupNetworkMonitoring() {
        networkMonitor = NWPathMonitor()
        networkMonitor?.pathUpdateHandler = { [weak self] path in
            DispatchQueue.main.async {
                if path.status == .satisfied {
                    self?.refreshFriendData()
                }
            }
        }
        
        let queue = DispatchQueue(label: "NetworkMonitor")
        networkMonitor?.start(queue: queue)
    }
    
    // MARK: - 自动刷新
    private func startAutoRefresh() {
        guard autoRefreshEnabled else { return }
        
        refreshTimer = Timer.scheduledTimer(withTimeInterval: refreshInterval, repeats: true) { [weak self] _ in
            Task {
                await self?.refreshFriendData()
            }
        }
    }
    
    private func stopAutoRefresh() {
        refreshTimer?.invalidate()
        refreshTimer = nil
    }
    
    public func setAutoRefresh(enabled: Bool, interval: TimeInterval = 30.0) {
        autoRefreshEnabled = enabled
        refreshInterval = interval
        
        stopAutoRefresh()
        if enabled {
            startAutoRefresh()
        }
    }
    
    // MARK: - 好友列表管理
    public func getFriendList() async throws -> [Friend] {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let url = buildUrl(endpoint: "/api/friends")
            let request = createAuthenticatedRequest(url: url, method: "GET")
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 {
                let friendResponse = try JSONDecoder().decode(FriendResponse.self, from: response.data)
                
                if friendResponse.success {
                    let newFriends = friendResponse.friends ?? []
                    self.friends = newFriends
                    delegate?.friendManager(self, didUpdateFriends: newFriends)
                    return newFriends
                } else {
                    throw FriendManagerError.serverError(friendResponse.message ?? "获取好友列表失败")
                }
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    public func getFriendRequests() async throws -> [FriendRequest] {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let url = buildUrl(endpoint: "/api/friends/requests")
            let request = createAuthenticatedRequest(url: url, method: "GET")
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 {
                let friendResponse = try JSONDecoder().decode(FriendResponse.self, from: response.data)
                
                if friendResponse.success {
                    let newRequests = friendResponse.requests ?? []
                    self.pendingRequests = newRequests
                    delegate?.friendManager(self, didUpdateRequests: newRequests)
                    return newRequests
                } else {
                    throw FriendManagerError.serverError(friendResponse.message ?? "获取好友请求失败")
                }
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    // MARK: - 好友操作
    public func addFriend(username: String, message: String? = nil) async throws -> Bool {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let addRequest = AddFriendRequest(username: username, message: message)
            let requestData = try JSONEncoder().encode(addRequest)
            
            let url = buildUrl(endpoint: "/api/friends/add")
            let request = createAuthenticatedRequest(url: url, method: "POST", body: requestData)
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 || response.statusCode == 201 {
                let friendResponse = try JSONDecoder().decode(FriendResponse.self, from: response.data)
                
                if friendResponse.success {
                    // 刷新好友请求列表
                    try await getFriendRequests()
                    return true
                } else {
                    throw FriendManagerError.serverError(friendResponse.message ?? "添加好友失败")
                }
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    public func respondToFriendRequest(requestId: String, accept: Bool) async throws -> Bool {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let respondData = RespondToRequestData(accept: accept)
            let requestData = try JSONEncoder().encode(respondData)
            
            let url = buildUrl(endpoint: "/api/friends/requests/\(requestId)")
            let request = createAuthenticatedRequest(url: url, method: "PUT", body: requestData)
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 {
                let friendResponse = try JSONDecoder().decode(FriendResponse.self, from: response.data)
                
                if friendResponse.success {
                    // 刷新好友列表和请求列表
                    async let friendsTask = getFriendList()
                    async let requestsTask = getFriendRequests()
                    
                    _ = try await [friendsTask, requestsTask]
                    return true
                } else {
                    throw FriendManagerError.serverError(friendResponse.message ?? "响应好友请求失败")
                }
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    public func removeFriend(friendId: String) async throws -> Bool {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let url = buildUrl(endpoint: "/api/friends/\(friendId)")
            let request = createAuthenticatedRequest(url: url, method: "DELETE")
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 {
                let friendResponse = try JSONDecoder().decode(FriendResponse.self, from: response.data)
                
                if friendResponse.success {
                    // 从本地列表中移除好友
                    friends.removeAll { $0.id == friendId }
                    delegate?.friendManager(self, didUpdateFriends: friends)
                    return true
                } else {
                    throw FriendManagerError.serverError(friendResponse.message ?? "删除好友失败")
                }
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    public func blockFriend(friendId: String) async throws -> Bool {
        return try await updateFriendBlockStatus(friendId: friendId, block: true)
    }
    
    public func unblockFriend(friendId: String) async throws -> Bool {
        return try await updateFriendBlockStatus(friendId: friendId, block: false)
    }
    
    private func updateFriendBlockStatus(friendId: String, block: Bool) async throws -> Bool {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let action = block ? "block" : "unblock"
            let actionData = ["action": action]
            let requestData = try JSONSerialization.data(withJSONObject: actionData)
            
            let endpoint = block ? "/api/friends/\(friendId)/block" : "/api/friends/\(friendId)/unblock"
            let url = buildUrl(endpoint: endpoint)
            let request = createAuthenticatedRequest(url: url, method: "PUT", body: requestData)
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 {
                let friendResponse = try JSONDecoder().decode(FriendResponse.self, from: response.data)
                
                if friendResponse.success {
                    // 刷新好友列表
                    try await getFriendList()
                    return true
                } else {
                    let actionText = block ? "屏蔽" : "取消屏蔽"
                    throw FriendManagerError.serverError(friendResponse.message ?? "\(actionText)好友失败")
                }
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    // MARK: - 在线状态管理
    public func updateOnlineStatus(_ status: OnlineStatus) async throws {
        guard authManager.isLoggedIn else {
            throw FriendManagerError.notAuthenticated
        }
        
        do {
            let statusRequest = UpdateStatusRequest(status: status)
            let requestData = try JSONEncoder().encode(statusRequest)
            
            let url = buildUrl(endpoint: "/api/user/status")
            let request = createAuthenticatedRequest(url: url, method: "PUT", body: requestData)
            let response = try await httpClient.sendRequest(request)
            
            if response.statusCode == 200 {
                currentOnlineStatus = status
            } else {
                throw FriendManagerError.serverError("HTTP \(response.statusCode)")
            }
        } catch {
            let friendError = error as? FriendManagerError ?? FriendManagerError.networkError(error)
            delegate?.friendManager(self, didEncounterError: friendError)
            throw friendError
        }
    }
    
    // MARK: - 查找方法
    public func findFriend(byId friendId: String) -> Friend? {
        return friends.first { $0.id == friendId }
    }
    
    public func findFriend(byUsername username: String) -> Friend? {
        return friends.first { $0.username == username }
    }
    
    public func findRequest(byId requestId: String) -> FriendRequest? {
        return pendingRequests.first { $0.id == requestId }
    }
    
    // MARK: - 配置方法
    public func setServerUrl(_ url: String) {
        serverUrl = url
    }
    
    // MARK: - 数据刷新
    public func refreshFriendData() async {
        guard authManager.isLoggedIn else { return }
        
        do {
            async let friendsTask = getFriendList()
            async let requestsTask = getFriendRequests()
            
            _ = try await [friendsTask, requestsTask]
        } catch {
            // 错误已经在各自的方法中处理
        }
    }
    
    // MARK: - 辅助方法
    private func buildUrl(endpoint: String) -> String {
        return serverUrl + endpoint
    }
    
    private func createAuthenticatedRequest(url: String, method: String, body: Data? = nil) -> HttpRequest {
        var request = HttpRequest(
            url: url,
            method: method,
            headers: ["Content-Type": "application/json"],
            body: body
        )
        
        // 添加认证头
        if let token = authManager.accessToken {
            request.headers["Authorization"] = "Bearer \(token)"
        }
        
        return request
    }
}

// MARK: - 全局实例管理
public class FriendManagerFactory {
    public static let shared = FriendManagerFactory()
    private var _instance: FriendManager?
    
    private init() {}
    
    public func initialize(httpClient: HttpClient, authManager: AuthManager, serverUrl: String = "http://localhost:8081") {
        _instance = FriendManager(httpClient: httpClient, authManager: authManager, serverUrl: serverUrl)
    }
    
    public var instance: FriendManager? {
        return _instance
    }
    
    public func cleanup() {
        _instance = nil
    }
}