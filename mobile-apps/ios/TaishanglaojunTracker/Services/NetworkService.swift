//
//  NetworkService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import Combine
import Network

/// 网络服务管理器
class NetworkService: ObservableObject {
    
    // MARK: - Published Properties
    @Published var isConnected = true
    @Published var connectionType: NWInterface.InterfaceType?
    
    // MARK: - Private Properties
    private let session: URLSession
    private let monitor = NWPathMonitor()
    private let monitorQueue = DispatchQueue(label: "NetworkMonitor")
    private var cancellables = Set<AnyCancellable>()
    
    // MARK: - Configuration
    private let baseURL = "https://api.taishanglaojun.com"
    private let apiVersion = "v1"
    
    // MARK: - Singleton
    static let shared = NetworkService()
    
    private init() {
        // 配置URLSession
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 60
        config.waitsForConnectivity = true
        
        self.session = URLSession(configuration: config)
        
        setupNetworkMonitoring()
    }
    
    // MARK: - Network Monitoring
    
    private func setupNetworkMonitoring() {
        monitor.pathUpdateHandler = { [weak self] path in
            DispatchQueue.main.async {
                self?.isConnected = path.status == .satisfied
                self?.connectionType = path.availableInterfaces.first?.type
                
                if path.status == .satisfied {
                    print("🌐 网络连接正常")
                } else {
                    print("❌ 网络连接断开")
                }
            }
        }
        
        monitor.start(queue: monitorQueue)
    }
    
    // MARK: - API Endpoints
    
    private func buildURL(endpoint: String) -> URL? {
        return URL(string: "\(baseURL)/\(apiVersion)/\(endpoint)")
    }
    
    // MARK: - Authentication
    
    private func createAuthenticatedRequest(url: URL, method: HTTPMethod = .GET) -> URLRequest {
        var request = URLRequest(url: url)
        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        
        // 添加认证头
        if let token = getAuthToken() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        // 添加设备信息
        request.setValue(UIDevice.current.identifierForVendor?.uuidString, forHTTPHeaderField: "X-Device-ID")
        request.setValue("iOS", forHTTPHeaderField: "X-Platform")
        request.setValue(UIDevice.current.systemVersion, forHTTPHeaderField: "X-OS-Version")
        
        return request
    }
    
    private func getAuthToken() -> String? {
        // 从Keychain获取认证token
        return KeychainService.shared.getToken()
    }
    
    // MARK: - Trajectory API
    
    /// 上传轨迹数据
    func uploadTrajectory(_ trajectory: Trajectory) -> AnyPublisher<TrajectoryResponse, Error> {
        guard let url = buildURL(endpoint: "trajectories") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = createAuthenticatedRequest(url: url, method: .POST)
        
        do {
            let trajectoryData = TrajectoryUploadData(from: trajectory)
            request.httpBody = try JSONEncoder().encode(trajectoryData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: TrajectoryResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    /// 获取轨迹列表
    func fetchTrajectories(page: Int = 1, limit: Int = 20) -> AnyPublisher<TrajectoryListResponse, Error> {
        guard let url = buildURL(endpoint: "trajectories?page=\(page)&limit=\(limit)") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        let request = createAuthenticatedRequest(url: url)
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: TrajectoryListResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    /// 获取轨迹详情
    func fetchTrajectory(id: UUID) -> AnyPublisher<TrajectoryResponse, Error> {
        guard let url = buildURL(endpoint: "trajectories/\(id.uuidString)") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        let request = createAuthenticatedRequest(url: url)
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: TrajectoryResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    /// 删除轨迹
    func deleteTrajectory(_ id: UUID) -> AnyPublisher<Bool, Error> {
        guard let url = buildURL(endpoint: "trajectories/\(id.uuidString)") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        let request = createAuthenticatedRequest(url: url, method: .DELETE)
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
                return true
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    // MARK: - Location API
    
    /// 实时上传位置点
    func uploadLocationPoint(_ point: LocationPoint) -> AnyPublisher<Bool, Error> {
        guard let url = buildURL(endpoint: "locations") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = createAuthenticatedRequest(url: url, method: .POST)
        
        do {
            let locationData = LocationUploadData(from: point)
            request.httpBody = try JSONEncoder().encode(locationData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
                return true
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    /// 批量上传位置点
    func uploadLocationPoints(_ points: [LocationPoint]) -> AnyPublisher<Bool, Error> {
        guard let url = buildURL(endpoint: "locations/batch") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = createAuthenticatedRequest(url: url, method: .POST)
        
        do {
            let locationsData = LocationBatchUploadData(points: points.map { LocationUploadData(from: $0) })
            request.httpBody = try JSONEncoder().encode(locationsData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
                return true
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    // MARK: - Chat API
    
    /// 上传对话
    func uploadConversation(_ conversation: Conversation) -> AnyPublisher<APIResponse, Error> {
        guard let url = buildURL(endpoint: "conversations") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = createAuthenticatedRequest(url: url, method: .POST)
        
        do {
            let conversationData = ConversationUploadData(from: conversation)
            request.httpBody = try JSONEncoder().encode(conversationData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: APIResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    /// 上传消息
    func uploadMessage(_ message: ChatMessage) -> AnyPublisher<APIResponse, Error> {
        guard let url = buildURL(endpoint: "messages") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = createAuthenticatedRequest(url: url, method: .POST)
        
        do {
            let messageData = MessageUploadData(from: message)
            request.httpBody = try JSONEncoder().encode(messageData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: APIResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    // MARK: - Authentication
    
    /// 用户登录
    func login(username: String, password: String) -> AnyPublisher<LoginResponse, Error> {
        guard let url = buildURL(endpoint: "auth/login") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.POST.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let loginData = LoginRequest(username: username, password: password)
        
        do {
            request.httpBody = try JSONEncoder().encode(loginData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: LoginResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    /// 用户注册
    func register(username: String, email: String, password: String) -> AnyPublisher<RegisterResponse, Error> {
        guard let url = buildURL(endpoint: "auth/register") else {
            return Fail(error: NetworkError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.POST.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let registerData = RegisterRequest(username: username, email: email, password: password)
        
        do {
            request.httpBody = try JSONEncoder().encode(registerData)
        } catch {
            return Fail(error: NetworkError.encodingError(error))
                .eraseToAnyPublisher()
        }
        
        return session.dataTaskPublisher(for: request)
            .tryMap { data, response in
                try self.handleResponse(data: data, response: response)
            }
            .decode(type: RegisterResponse.self, decoder: JSONDecoder())
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    // MARK: - Response Handling
    
    private func handleResponse(data: Data, response: URLResponse) throws -> Data {
        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.invalidResponse
        }
        
        switch httpResponse.statusCode {
        case 200...299:
            return data
        case 401:
            throw NetworkError.unauthorized
        case 403:
            throw NetworkError.forbidden
        case 404:
            throw NetworkError.notFound
        case 429:
            throw NetworkError.rateLimited
        case 500...599:
            throw NetworkError.serverError(httpResponse.statusCode)
        default:
            throw NetworkError.httpError(httpResponse.statusCode)
        }
    }
}

// MARK: - HTTP Method
enum HTTPMethod: String {
    case GET = "GET"
    case POST = "POST"
    case PUT = "PUT"
    case DELETE = "DELETE"
    case PATCH = "PATCH"
}

// MARK: - Network Errors
enum NetworkError: LocalizedError {
    case invalidURL
    case invalidResponse
    case unauthorized
    case forbidden
    case notFound
    case rateLimited
    case serverError(Int)
    case httpError(Int)
    case encodingError(Error)
    case decodingError(Error)
    case networkUnavailable
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "无效的URL"
        case .invalidResponse:
            return "无效的响应"
        case .unauthorized:
            return "未授权访问"
        case .forbidden:
            return "访问被禁止"
        case .notFound:
            return "资源未找到"
        case .rateLimited:
            return "请求频率过高"
        case .serverError(let code):
            return "服务器错误 (\(code))"
        case .httpError(let code):
            return "HTTP错误 (\(code))"
        case .encodingError(let error):
            return "编码错误: \(error.localizedDescription)"
        case .decodingError(let error):
            return "解码错误: \(error.localizedDescription)"
        case .networkUnavailable:
            return "网络不可用"
        }
    }
}

// MARK: - Data Transfer Objects

struct TrajectoryUploadData: Codable {
    let id: String
    let name: String
    let startTime: Date
    let endTime: Date?
    let points: [LocationUploadData]
    
    init(from trajectory: Trajectory) {
        self.id = trajectory.id.uuidString
        self.name = trajectory.name
        self.startTime = trajectory.startTime
        self.endTime = trajectory.endTime
        self.points = trajectory.points.map { LocationUploadData(from: $0) }
    }
}

struct LocationUploadData: Codable {
    let latitude: Double
    let longitude: Double
    let timestamp: Date
    let accuracy: Double
    let speed: Double?
    let course: Double?
    let altitude: Double?
    let trajectoryId: String?
    
    init(from point: LocationPoint) {
        self.latitude = point.latitude
        self.longitude = point.longitude
        self.timestamp = point.timestamp
        self.accuracy = point.accuracy
        self.speed = point.speed
        self.course = point.course
        self.altitude = point.altitude
        self.trajectoryId = point.trajectoryId?.uuidString
    }
}

struct LocationBatchUploadData: Codable {
    let points: [LocationUploadData]
}

struct LoginRequest: Codable {
    let username: String
    let password: String
}

struct RegisterRequest: Codable {
    let username: String
    let email: String
    let password: String
}

// MARK: - Chat Data Models

struct ConversationUploadData: Codable {
    let id: String
    let title: String
    let aiPersonality: String
    let createdAt: Date
    let updatedAt: Date
    let isArchived: Bool
    let messageCount: Int
    
    init(from conversation: Conversation) {
        self.id = conversation.id
        self.title = conversation.title
        self.aiPersonality = conversation.aiPersonality.rawValue
        self.createdAt = conversation.createdAt
        self.updatedAt = conversation.updatedAt
        self.isArchived = conversation.isArchived
        self.messageCount = conversation.messageCount
    }
}

struct MessageUploadData: Codable {
    let id: String
    let conversationId: String
    let content: String
    let messageType: String
    let sender: String
    let timestamp: Date
    let status: String
    let metadata: String?
    
    init(from message: ChatMessage) {
        self.id = message.id
        self.conversationId = message.conversationId
        self.content = message.content
        self.messageType = message.messageType.rawValue
        self.sender = message.sender.rawValue
        self.timestamp = message.timestamp
        self.status = message.status.rawValue
        self.metadata = message.metadata
    }
}

struct APIResponse: Codable {
    let success: Bool
    let message: String?
    let data: [String: String]?
}

// MARK: - Response Models

struct TrajectoryResponse: Codable {
    let id: String
    let name: String
    let startTime: Date
    let endTime: Date?
    let points: [LocationPointResponse]
    let createdAt: Date
    let updatedAt: Date
}

struct TrajectoryListResponse: Codable {
    let trajectories: [TrajectoryResponse]
    let totalCount: Int
    let page: Int
    let limit: Int
}

struct LocationPointResponse: Codable {
    let latitude: Double
    let longitude: Double
    let timestamp: Date
    let accuracy: Double
    let speed: Double?
    let course: Double?
    let altitude: Double?
}

struct LoginResponse: Codable {
    let token: String
    let refreshToken: String
    let user: UserResponse
    let expiresAt: Date
}

struct RegisterResponse: Codable {
    let user: UserResponse
    let message: String
}

struct UserResponse: Codable {
    let id: String
    let username: String
    let email: String
    let createdAt: Date
}