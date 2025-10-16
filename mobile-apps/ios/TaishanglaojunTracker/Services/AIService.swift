import Foundation
import Combine
import Network

// MARK: - AI Service Protocol
protocol AIServiceProtocol {
    func sendMessage(_ request: SendMessageRequest) -> AnyPublisher<AIResponse, Error>
    func getMessages(conversationId: String, page: Int, limit: Int, before: TimeInterval?) -> AnyPublisher<MessagesData, Error>
    func createConversation(_ request: CreateConversationRequest) -> AnyPublisher<ConversationData, Error>
    func getConversations() -> AnyPublisher<[Conversation], Error>
    func deleteConversation(conversationId: String) -> AnyPublisher<Void, Error>
    func sendMessageStream(_ request: SendMessageRequest) -> AnyPublisher<String, Error>
}

// MARK: - AI Service Implementation
class AIService: AIServiceProtocol {
    static let shared = AIService()
    
    private let baseURL = "https://api.taishanglaojun.com"
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    private let monitor = NWPathMonitor()
    private let monitorQueue = DispatchQueue(label: "NetworkMonitor")
    
    @Published private(set) var isConnected = true
    
    private init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 60
        self.session = URLSession(configuration: config)
        
        self.decoder = JSONDecoder()
        self.encoder = JSONEncoder()
        
        setupNetworkMonitoring()
    }
    
    private func setupNetworkMonitoring() {
        monitor.pathUpdateHandler = { [weak self] path in
            DispatchQueue.main.async {
                self?.isConnected = path.status == .satisfied
            }
        }
        monitor.start(queue: monitorQueue)
    }
    
    // MARK: - API Methods
    func sendMessage(_ request: SendMessageRequest) -> AnyPublisher<AIResponse, Error> {
        guard let url = URL(string: "\(baseURL)/api/ai/chat/send") else {
            return Fail(error: AIServiceError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        return createRequest(url: url, method: "POST", body: request)
            .flatMap { request in
                self.session.dataTaskPublisher(for: request)
                    .map(\.data)
                    .decode(type: AIResponse.self, decoder: self.decoder)
                    .eraseToAnyPublisher()
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    func getMessages(conversationId: String, page: Int = 1, limit: Int = 50, before: TimeInterval? = nil) -> AnyPublisher<MessagesData, Error> {
        var components = URLComponents(string: "\(baseURL)/api/ai/chat/conversations/\(conversationId)/messages")
        components?.queryItems = [
            URLQueryItem(name: "page", value: "\(page)"),
            URLQueryItem(name: "limit", value: "\(limit)")
        ]
        
        if let before = before {
            components?.queryItems?.append(URLQueryItem(name: "before", value: "\(Int(before))"))
        }
        
        guard let url = components?.url else {
            return Fail(error: AIServiceError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        return createRequest(url: url, method: "GET")
            .flatMap { request in
                self.session.dataTaskPublisher(for: request)
                    .map(\.data)
                    .decode(type: MessagesResponse.self, decoder: self.decoder)
                    .map { response in
                        if response.success {
                            return response.data
                        } else {
                            throw AIServiceError.serverError(response.error?.message ?? "Unknown error")
                        }
                    }
                    .eraseToAnyPublisher()
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    func createConversation(_ request: CreateConversationRequest) -> AnyPublisher<ConversationData, Error> {
        guard let url = URL(string: "\(baseURL)/api/ai/chat/conversations") else {
            return Fail(error: AIServiceError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        return createRequest(url: url, method: "POST", body: request)
            .flatMap { request in
                self.session.dataTaskPublisher(for: request)
                    .map(\.data)
                    .decode(type: ConversationResponse.self, decoder: self.decoder)
                    .map { response in
                        if response.success {
                            return response.data
                        } else {
                            throw AIServiceError.serverError(response.error?.message ?? "Unknown error")
                        }
                    }
                    .eraseToAnyPublisher()
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    func getConversations() -> AnyPublisher<[Conversation], Error> {
        guard let url = URL(string: "\(baseURL)/api/ai/chat/conversations") else {
            return Fail(error: AIServiceError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        return createRequest(url: url, method: "GET")
            .flatMap { request in
                self.session.dataTaskPublisher(for: request)
                    .map(\.data)
                    .decode(type: ConversationsResponse.self, decoder: self.decoder)
                    .map { response in
                        if response.success {
                            return response.data
                        } else {
                            throw AIServiceError.serverError(response.error?.message ?? "Unknown error")
                        }
                    }
                    .eraseToAnyPublisher()
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    func deleteConversation(conversationId: String) -> AnyPublisher<Void, Error> {
        guard let url = URL(string: "\(baseURL)/api/ai/chat/conversations/\(conversationId)") else {
            return Fail(error: AIServiceError.invalidURL)
                .eraseToAnyPublisher()
        }
        
        return createRequest(url: url, method: "DELETE")
            .flatMap { request in
                self.session.dataTaskPublisher(for: request)
                    .map { _ in () }
                    .eraseToAnyPublisher()
            }
            .receive(on: DispatchQueue.main)
            .eraseToAnyPublisher()
    }
    
    func sendMessageStream(_ request: SendMessageRequest) -> AnyPublisher<String, Error> {
        // 模拟流式响应，实际项目中可能需要WebSocket或SSE
        return sendMessage(request)
            .flatMap { response -> AnyPublisher<String, Error> in
                guard let content = response.message?.content else {
                    return Fail(error: AIServiceError.noContent)
                        .eraseToAnyPublisher()
                }
                
                let words = content.components(separatedBy: " ")
                return Publishers.Sequence(sequence: words)
                    .delay(for: .milliseconds(100), scheduler: DispatchQueue.main)
                    .eraseToAnyPublisher()
            }
            .eraseToAnyPublisher()
    }
    
    // MARK: - Helper Methods
    private func createRequest<T: Codable>(url: URL, method: String, body: T? = nil) -> AnyPublisher<URLRequest, Error> {
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        // Add authorization header if available
        if let token = getAuthToken() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        if let body = body {
            do {
                request.httpBody = try encoder.encode(body)
            } catch {
                return Fail(error: error)
                    .eraseToAnyPublisher()
            }
        }
        
        return Just(request)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    private func getAuthToken() -> String? {
        // 从Keychain或UserDefaults获取认证令牌
        return UserDefaults.standard.string(forKey: "auth_token")
    }
}

// MARK: - Mock AI Service for Testing
class MockAIService: AIServiceProtocol {
    private let delay: TimeInterval
    
    init(delay: TimeInterval = 1.0) {
        self.delay = delay
    }
    
    func sendMessage(_ request: SendMessageRequest) -> AnyPublisher<AIResponse, Error> {
        let mockResponse = AIResponse(
            success: true,
            message: AIMessage(
                content: generateMockResponse(for: request.message, personality: request.aiPersonality),
                messageType: .text,
                timestamp: Date().timeIntervalSince1970,
                metadata: [
                    "model": AnyCodable("taishang-v1"),
                    "tokens_used": AnyCodable(Int.random(in: 20...100)),
                    "response_time": AnyCodable(Int.random(in: 800...2000))
                ]
            ),
            suggestions: generateSuggestions(for: request.aiPersonality)
        )
        
        return Just(mockResponse)
            .delay(for: .seconds(delay), scheduler: DispatchQueue.main)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    func getMessages(conversationId: String, page: Int, limit: Int, before: TimeInterval?) -> AnyPublisher<MessagesData, Error> {
        let mockData = MessagesData(
            messages: [],
            hasMore: false,
            total: 0
        )
        
        return Just(mockData)
            .delay(for: .seconds(delay), scheduler: DispatchQueue.main)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    func createConversation(_ request: CreateConversationRequest) -> AnyPublisher<ConversationData, Error> {
        let mockData = ConversationData(
            conversationId: UUID().uuidString,
            title: request.title,
            createdAt: Date().timeIntervalSince1970
        )
        
        return Just(mockData)
            .delay(for: .seconds(delay), scheduler: DispatchQueue.main)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    func getConversations() -> AnyPublisher<[Conversation], Error> {
        return Just([])
            .delay(for: .seconds(delay), scheduler: DispatchQueue.main)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    func deleteConversation(conversationId: String) -> AnyPublisher<Void, Error> {
        return Just(())
            .delay(for: .seconds(delay), scheduler: DispatchQueue.main)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    func sendMessageStream(_ request: SendMessageRequest) -> AnyPublisher<String, Error> {
        let response = generateMockResponse(for: request.message, personality: request.aiPersonality)
        let words = response.components(separatedBy: " ")
        
        return Publishers.Sequence(sequence: words)
            .delay(for: .milliseconds(100), scheduler: DispatchQueue.main)
            .setFailureType(to: Error.self)
            .eraseToAnyPublisher()
    }
    
    private func generateMockResponse(for message: String, personality: AIPersonality) -> String {
        switch personality {
        case .default:
            return "感谢您的提问。关于「\(message)」，我认为这是一个很有意思的话题。"
        case .wiseSage:
            return "古人云：「\(message)」，此言深有道理。老夫以为，万事皆有其道，需细细参悟。"
        case .friendlyGuide:
            return "您好！关于「\(message)」这个问题，让我来为您详细解答一下。"
        case .scholarly:
            return "根据相关研究和理论分析，「\(message)」这一概念具有重要的学术价值和实践意义。"
        case .poetic:
            return "「\(message)」如春风化雨，润物无声。正所谓：山重水复疑无路，柳暗花明又一村。"
        }
    }
    
    private func generateSuggestions(for personality: AIPersonality) -> [String] {
        switch personality {
        case .default:
            return ["请问您还有其他问题吗？", "我可以为您详细解释", "需要更多信息吗？"]
        case .wiseSage:
            return ["可否请教更深层的道理？", "古籍中还有何见解？", "此理可有实践之法？"]
        case .friendlyGuide:
            return ["还有什么我可以帮助您的？", "您想了解更多细节吗？", "有其他相关问题吗？"]
        case .scholarly:
            return ["是否需要相关文献资料？", "可以进一步分析吗？", "有实证研究支持吗？"]
        case .poetic:
            return ["可否再吟一首？", "此情此景，还有何感？", "诗意人生，何处寻觅？"]
        }
    }
}

// MARK: - AI Service Error
enum AIServiceError: LocalizedError {
    case invalidURL
    case noContent
    case serverError(String)
    case networkError
    case decodingError
    
    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "无效的URL"
        case .noContent:
            return "没有内容"
        case .serverError(let message):
            return "服务器错误: \(message)"
        case .networkError:
            return "网络连接错误"
        case .decodingError:
            return "数据解析错误"
        }
    }
}