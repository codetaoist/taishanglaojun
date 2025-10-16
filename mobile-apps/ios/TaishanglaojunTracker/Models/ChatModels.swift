import Foundation
import CoreData

// MARK: - Chat Message Model
struct ChatMessage: Codable, Identifiable, Equatable {
    let id: String
    let conversationId: String
    let content: String
    let messageType: MessageType
    let sender: MessageSender
    let timestamp: TimeInterval
    var status: MessageStatus
    let metadata: [String: AnyCodable]
    
    init(
        id: String = UUID().uuidString,
        conversationId: String,
        content: String,
        messageType: MessageType = .text,
        sender: MessageSender,
        timestamp: TimeInterval = Date().timeIntervalSince1970,
        status: MessageStatus = .sending,
        metadata: [String: AnyCodable] = [:]
    ) {
        self.id = id
        self.conversationId = conversationId
        self.content = content
        self.messageType = messageType
        self.sender = sender
        self.timestamp = timestamp
        self.status = status
        self.metadata = metadata
    }
}

// MARK: - Message Type
enum MessageType: String, Codable, CaseIterable {
    case text = "TEXT"
    case image = "IMAGE"
    case audio = "AUDIO"
    case system = "SYSTEM"
    
    var displayName: String {
        switch self {
        case .text: return "文本"
        case .image: return "图片"
        case .audio: return "语音"
        case .system: return "系统"
        }
    }
}

// MARK: - Message Sender
enum MessageSender: String, Codable, CaseIterable {
    case user = "USER"
    case ai = "AI"
    case system = "SYSTEM"
    
    var displayName: String {
        switch self {
        case .user: return "用户"
        case .ai: return "AI助手"
        case .system: return "系统"
        }
    }
}

// MARK: - Message Status
enum MessageStatus: String, Codable, CaseIterable {
    case sending = "SENDING"
    case sent = "SENT"
    case delivered = "DELIVERED"
    case read = "READ"
    case failed = "FAILED"
    
    var displayName: String {
        switch self {
        case .sending: return "发送中"
        case .sent: return "已发送"
        case .delivered: return "已送达"
        case .read: return "已读"
        case .failed: return "发送失败"
        }
    }
}

// MARK: - Conversation Model
struct Conversation: Codable, Identifiable, Equatable {
    let id: String
    var title: String
    let createdAt: TimeInterval
    var updatedAt: TimeInterval
    var lastMessage: ChatMessage?
    var messageCount: Int
    var isArchived: Bool
    var aiPersonality: AIPersonality
    
    init(
        id: String = UUID().uuidString,
        title: String,
        createdAt: TimeInterval = Date().timeIntervalSince1970,
        updatedAt: TimeInterval = Date().timeIntervalSince1970,
        lastMessage: ChatMessage? = nil,
        messageCount: Int = 0,
        isArchived: Bool = false,
        aiPersonality: AIPersonality = .default
    ) {
        self.id = id
        self.title = title
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.lastMessage = lastMessage
        self.messageCount = messageCount
        self.isArchived = isArchived
        self.aiPersonality = aiPersonality
    }
}

// MARK: - AI Personality
enum AIPersonality: String, Codable, CaseIterable {
    case `default` = "DEFAULT"
    case wiseSage = "WISE_SAGE"
    case friendlyGuide = "FRIENDLY_GUIDE"
    case scholarly = "SCHOLARLY"
    case poetic = "POETIC"
    
    var displayName: String {
        switch self {
        case .default: return "默认"
        case .wiseSage: return "智慧长者"
        case .friendlyGuide: return "友善向导"
        case .scholarly: return "学者风格"
        case .poetic: return "诗意风格"
        }
    }
    
    var description: String {
        switch self {
        case .default: return "平衡的AI助手，适合日常对话"
        case .wiseSage: return "以古代圣贤的智慧回答问题"
        case .friendlyGuide: return "温和友善，循循善诱的引导者"
        case .scholarly: return "严谨学术，深入分析问题"
        case .poetic: return "富有诗意，文雅的表达方式"
        }
    }
}

// MARK: - AI Response Models
struct AIResponse: Codable {
    let success: Bool
    let message: AIMessage?
    let suggestions: [String]
    let error: AIError?
    
    init(success: Bool, message: AIMessage? = nil, suggestions: [String] = [], error: AIError? = nil) {
        self.success = success
        self.message = message
        self.suggestions = suggestions
        self.error = error
    }
}

struct AIMessage: Codable {
    let content: String
    let messageType: MessageType
    let timestamp: TimeInterval
    let metadata: [String: AnyCodable]
    
    init(content: String, messageType: MessageType = .text, timestamp: TimeInterval = Date().timeIntervalSince1970, metadata: [String: AnyCodable] = [:]) {
        self.content = content
        self.messageType = messageType
        self.timestamp = timestamp
        self.metadata = metadata
    }
}

struct AIError: Codable {
    let code: String
    let message: String
    let details: [String: AnyCodable]
    
    init(code: String, message: String, details: [String: AnyCodable] = [:]) {
        self.code = code
        self.message = message
        self.details = details
    }
}

// MARK: - Request Models
struct SendMessageRequest: Codable {
    let conversationId: String
    let message: String
    let messageType: MessageType
    let aiPersonality: AIPersonality
    
    init(conversationId: String, message: String, messageType: MessageType = .text, aiPersonality: AIPersonality = .default) {
        self.conversationId = conversationId
        self.message = message
        self.messageType = messageType
        self.aiPersonality = aiPersonality
    }
}

struct CreateConversationRequest: Codable {
    let title: String
    let aiPersonality: AIPersonality
    
    init(title: String, aiPersonality: AIPersonality = .default) {
        self.title = title
        self.aiPersonality = aiPersonality
    }
}

// MARK: - Response Models
struct MessagesResponse: Codable {
    let success: Bool
    let data: MessagesData
    let error: AIError?
}

struct MessagesData: Codable {
    let messages: [ChatMessage]
    let hasMore: Bool
    let total: Int
}

struct ConversationResponse: Codable {
    let success: Bool
    let data: ConversationData
    let error: AIError?
}

struct ConversationData: Codable {
    let conversationId: String
    let title: String
    let createdAt: TimeInterval
}

struct ConversationsResponse: Codable {
    let success: Bool
    let data: [Conversation]
    let error: AIError?
}

// MARK: - AnyCodable Helper
struct AnyCodable: Codable {
    let value: Any
    
    init<T>(_ value: T?) {
        self.value = value ?? ()
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if container.decodeNil() {
            self.init(())
        } else if let bool = try? container.decode(Bool.self) {
            self.init(bool)
        } else if let int = try? container.decode(Int.self) {
            self.init(int)
        } else if let double = try? container.decode(Double.self) {
            self.init(double)
        } else if let string = try? container.decode(String.self) {
            self.init(string)
        } else if let array = try? container.decode([AnyCodable].self) {
            self.init(array.map { $0.value })
        } else if let dictionary = try? container.decode([String: AnyCodable].self) {
            self.init(dictionary.mapValues { $0.value })
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "AnyCodable value cannot be decoded")
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        switch value {
        case is Void:
            try container.encodeNil()
        case let bool as Bool:
            try container.encode(bool)
        case let int as Int:
            try container.encode(int)
        case let double as Double:
            try container.encode(double)
        case let string as String:
            try container.encode(string)
        case let array as [Any]:
            try container.encode(array.map { AnyCodable($0) })
        case let dictionary as [String: Any]:
            try container.encode(dictionary.mapValues { AnyCodable($0) })
        default:
            let context = EncodingError.Context(codingPath: container.codingPath, debugDescription: "AnyCodable value cannot be encoded")
            throw EncodingError.invalidValue(value, context)
        }
    }
}

// MARK: - Core Data Extensions
extension ChatMessage {
    func toCoreDataEntity(context: NSManagedObjectContext) -> ChatMessageEntity {
        let entity = ChatMessageEntity(context: context)
        entity.id = self.id
        entity.conversationId = self.conversationId
        entity.content = self.content
        entity.messageType = self.messageType.rawValue
        entity.sender = self.sender.rawValue
        entity.timestamp = self.timestamp
        entity.status = self.status.rawValue
        
        // Convert metadata to JSON string
        if let data = try? JSONEncoder().encode(self.metadata),
           let jsonString = String(data: data, encoding: .utf8) {
            entity.metadata = jsonString
        }
        
        return entity
    }
    
    static func fromCoreDataEntity(_ entity: ChatMessageEntity) -> ChatMessage? {
        guard let id = entity.id,
              let conversationId = entity.conversationId,
              let content = entity.content,
              let messageTypeRaw = entity.messageType,
              let messageType = MessageType(rawValue: messageTypeRaw),
              let senderRaw = entity.sender,
              let sender = MessageSender(rawValue: senderRaw),
              let statusRaw = entity.status,
              let status = MessageStatus(rawValue: statusRaw) else {
            return nil
        }
        
        var metadata: [String: AnyCodable] = [:]
        if let metadataString = entity.metadata,
           let data = metadataString.data(using: .utf8) {
            metadata = (try? JSONDecoder().decode([String: AnyCodable].self, from: data)) ?? [:]
        }
        
        return ChatMessage(
            id: id,
            conversationId: conversationId,
            content: content,
            messageType: messageType,
            sender: sender,
            timestamp: entity.timestamp,
            status: status,
            metadata: metadata
        )
    }
}

extension Conversation {
    func toCoreDataEntity(context: NSManagedObjectContext) -> ConversationEntity {
        let entity = ConversationEntity(context: context)
        entity.id = self.id
        entity.title = self.title
        entity.createdAt = self.createdAt
        entity.updatedAt = self.updatedAt
        entity.lastMessageId = self.lastMessage?.id
        entity.messageCount = Int32(self.messageCount)
        entity.isArchived = self.isArchived
        entity.aiPersonality = self.aiPersonality.rawValue
        return entity
    }
    
    static func fromCoreDataEntity(_ entity: ConversationEntity) -> Conversation? {
        guard let id = entity.id,
              let title = entity.title,
              let personalityRaw = entity.aiPersonality,
              let personality = AIPersonality(rawValue: personalityRaw) else {
            return nil
        }
        
        return Conversation(
            id: id,
            title: title,
            createdAt: entity.createdAt,
            updatedAt: entity.updatedAt,
            messageCount: Int(entity.messageCount),
            isArchived: entity.isArchived,
            aiPersonality: personality
        )
    }
}