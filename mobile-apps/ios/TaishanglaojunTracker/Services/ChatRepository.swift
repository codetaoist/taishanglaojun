import Foundation
import Combine
import CoreData

// MARK: - Chat Repository Protocol
protocol ChatRepositoryProtocol {
    // Conversation Management
    func createConversation(title: String, aiPersonality: AIPersonality) -> AnyPublisher<Conversation, Error>
    func getConversations() -> AnyPublisher<[Conversation], Error>
    func getConversation(id: String) -> AnyPublisher<Conversation?, Error>
    func updateConversation(_ conversation: Conversation) -> AnyPublisher<Conversation, Error>
    func deleteConversation(id: String) -> AnyPublisher<Void, Error>
    func archiveConversation(id: String) -> AnyPublisher<Void, Error>
    
    // Message Management
    func sendMessage(conversationId: String, content: String, messageType: MessageType) -> AnyPublisher<ChatMessage, Error>
    func getMessages(conversationId: String, limit: Int, before: Date?) -> AnyPublisher<[ChatMessage], Error>
    func retryMessage(messageId: String) -> AnyPublisher<ChatMessage, Error>
    func deleteMessage(messageId: String) -> AnyPublisher<Void, Error>
    func updateMessageStatus(messageId: String, status: MessageStatus) -> AnyPublisher<Void, Error>
    
    // Search and Statistics
    func searchMessages(query: String, conversationId: String?) -> AnyPublisher<[ChatMessage], Error>
    func getMessageStatistics(conversationId: String) -> AnyPublisher<MessageStatistics, Error>
    
    // Sync
    func syncConversations() -> AnyPublisher<Void, Error>
    func syncMessages(conversationId: String) -> AnyPublisher<Void, Error>
}

// MARK: - Message Statistics
struct MessageStatistics {
    let totalMessages: Int
    let userMessages: Int
    let aiMessages: Int
    let averageResponseTime: TimeInterval
    let lastMessageTime: Date?
}

// MARK: - Chat Repository Implementation
class ChatRepository: ChatRepositoryProtocol {
    static let shared = ChatRepository()
    
    private let aiService: AIServiceProtocol
    private let dataService: DataService
    private let context: NSManagedObjectContext
    private var cancellables = Set<AnyCancellable>()
    
    init(aiService: AIServiceProtocol = AIService.shared, dataService: DataService = DataService.shared) {
        self.aiService = aiService
        self.dataService = dataService
        self.context = dataService.persistentContainer.viewContext
    }
    
    // MARK: - Conversation Management
    func createConversation(title: String, aiPersonality: AIPersonality) -> AnyPublisher<Conversation, Error> {
        let request = CreateConversationRequest(
            title: title,
            aiPersonality: aiPersonality
        )
        
        return aiService.createConversation(request)
            .flatMap { [weak self] response -> AnyPublisher<Conversation, Error> in
                guard let self = self else {
                    return Fail(error: ChatRepositoryError.repositoryDeallocated)
                        .eraseToAnyPublisher()
                }
                
                return self.saveConversationLocally(
                    id: response.conversationId,
                    title: title,
                    aiPersonality: aiPersonality,
                    createdAt: Date(timeIntervalSince1970: response.createdAt)
                )
            }
            .eraseToAnyPublisher()
    }
    
    func getConversations() -> AnyPublisher<[Conversation], Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
            request.sortDescriptors = [NSSortDescriptor(key: "updatedAt", ascending: false)]
            request.predicate = NSPredicate(format: "isArchived == NO")
            
            do {
                let entities = try self.context.fetch(request)
                let conversations = entities.compactMap { $0.toConversation() }
                promise(.success(conversations))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func getConversation(id: String) -> AnyPublisher<Conversation?, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", id)
            request.fetchLimit = 1
            
            do {
                let entities = try self.context.fetch(request)
                let conversation = entities.first?.toConversation()
                promise(.success(conversation))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func updateConversation(_ conversation: Conversation) -> AnyPublisher<Conversation, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", conversation.id)
            request.fetchLimit = 1
            
            do {
                if let entity = try self.context.fetch(request).first {
                    entity.title = conversation.title
                    entity.aiPersonality = conversation.aiPersonality.rawValue
                    entity.updatedAt = Date()
                    
                    try self.context.save()
                    
                    if let updatedConversation = entity.toConversation() {
                        promise(.success(updatedConversation))
                    } else {
                        promise(.failure(ChatRepositoryError.conversionError))
                    }
                } else {
                    promise(.failure(ChatRepositoryError.conversationNotFound))
                }
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func deleteConversation(id: String) -> AnyPublisher<Void, Error> {
        return aiService.deleteConversation(conversationId: id)
            .flatMap { [weak self] _ -> AnyPublisher<Void, Error> in
                guard let self = self else {
                    return Fail(error: ChatRepositoryError.repositoryDeallocated)
                        .eraseToAnyPublisher()
                }
                
                return self.deleteConversationLocally(id: id)
            }
            .eraseToAnyPublisher()
    }
    
    func archiveConversation(id: String) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", id)
            request.fetchLimit = 1
            
            do {
                if let entity = try self.context.fetch(request).first {
                    entity.isArchived = true
                    entity.updatedAt = Date()
                    try self.context.save()
                    promise(.success(()))
                } else {
                    promise(.failure(ChatRepositoryError.conversationNotFound))
                }
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    // MARK: - Message Management
    func sendMessage(conversationId: String, content: String, messageType: MessageType) -> AnyPublisher<ChatMessage, Error> {
        // 首先保存用户消息到本地
        return saveMessageLocally(
            conversationId: conversationId,
            content: content,
            messageType: messageType,
            sender: .user,
            status: .sending
        )
        .flatMap { [weak self] userMessage -> AnyPublisher<ChatMessage, Error> in
            guard let self = self else {
                return Fail(error: ChatRepositoryError.repositoryDeallocated)
                    .eraseToAnyPublisher()
            }
            
            // 获取对话信息以确定AI人格
            return self.getConversation(id: conversationId)
                .flatMap { conversation -> AnyPublisher<ChatMessage, Error> in
                    guard let conversation = conversation else {
                        return Fail(error: ChatRepositoryError.conversationNotFound)
                            .eraseToAnyPublisher()
                    }
                    
                    let request = SendMessageRequest(
                        conversationId: conversationId,
                        message: content,
                        messageType: messageType,
                        aiPersonality: conversation.aiPersonality
                    )
                    
                    // 发送消息到AI服务
                    return self.aiService.sendMessage(request)
                        .flatMap { response -> AnyPublisher<ChatMessage, Error> in
                            // 更新用户消息状态为已发送
                            return self.updateMessageStatusLocally(messageId: userMessage.id, status: .sent)
                                .flatMap { _ -> AnyPublisher<ChatMessage, Error> in
                                    // 保存AI回复
                                    guard let aiMessage = response.message else {
                                        return Fail(error: ChatRepositoryError.noAIResponse)
                                            .eraseToAnyPublisher()
                                    }
                                    
                                    return self.saveMessageLocally(
                                        conversationId: conversationId,
                                        content: aiMessage.content,
                                        messageType: aiMessage.messageType,
                                        sender: .ai,
                                        status: .sent,
                                        metadata: aiMessage.metadata
                                    )
                                }
                        }
                        .catch { error -> AnyPublisher<ChatMessage, Error> in
                            // 如果发送失败，更新用户消息状态
                            return self.updateMessageStatusLocally(messageId: userMessage.id, status: .failed)
                                .flatMap { _ in
                                    Fail<ChatMessage, Error>(error: error)
                                }
                                .eraseToAnyPublisher()
                        }
                }
        }
        .eraseToAnyPublisher()
    }
    
    func getMessages(conversationId: String, limit: Int = 50, before: Date? = nil) -> AnyPublisher<[ChatMessage], Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
            var predicates = [NSPredicate(format: "conversationId == %@", conversationId)]
            
            if let before = before {
                predicates.append(NSPredicate(format: "timestamp < %@", before as NSDate))
            }
            
            request.predicate = NSCompoundPredicate(andPredicateWithSubpredicates: predicates)
            request.sortDescriptors = [NSSortDescriptor(key: "timestamp", ascending: false)]
            request.fetchLimit = limit
            
            do {
                let entities = try self.context.fetch(request)
                let messages = entities.compactMap { $0.toChatMessage() }.reversed()
                promise(.success(Array(messages)))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func retryMessage(messageId: String) -> AnyPublisher<ChatMessage, Error> {
        return getMessageLocally(id: messageId)
            .flatMap { [weak self] message -> AnyPublisher<ChatMessage, Error> in
                guard let self = self, let message = message else {
                    return Fail(error: ChatRepositoryError.messageNotFound)
                        .eraseToAnyPublisher()
                }
                
                return self.sendMessage(
                    conversationId: message.conversationId,
                    content: message.content,
                    messageType: message.messageType
                )
            }
            .eraseToAnyPublisher()
    }
    
    func deleteMessage(messageId: String) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", messageId)
            request.fetchLimit = 1
            
            do {
                if let entity = try self.context.fetch(request).first {
                    self.context.delete(entity)
                    try self.context.save()
                    promise(.success(()))
                } else {
                    promise(.failure(ChatRepositoryError.messageNotFound))
                }
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func updateMessageStatus(messageId: String, status: MessageStatus) -> AnyPublisher<Void, Error> {
        return updateMessageStatusLocally(messageId: messageId, status: status)
    }
    
    // MARK: - Search and Statistics
    func searchMessages(query: String, conversationId: String? = nil) -> AnyPublisher<[ChatMessage], Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
            var predicates = [NSPredicate(format: "content CONTAINS[cd] %@", query)]
            
            if let conversationId = conversationId {
                predicates.append(NSPredicate(format: "conversationId == %@", conversationId))
            }
            
            request.predicate = NSCompoundPredicate(andPredicateWithSubpredicates: predicates)
            request.sortDescriptors = [NSSortDescriptor(key: "timestamp", ascending: false)]
            request.fetchLimit = 100
            
            do {
                let entities = try self.context.fetch(request)
                let messages = entities.compactMap { $0.toChatMessage() }
                promise(.success(messages))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func getMessageStatistics(conversationId: String) -> AnyPublisher<MessageStatistics, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
            request.predicate = NSPredicate(format: "conversationId == %@", conversationId)
            
            do {
                let entities = try self.context.fetch(request)
                let messages = entities.compactMap { $0.toChatMessage() }
                
                let totalMessages = messages.count
                let userMessages = messages.filter { $0.sender == .user }.count
                let aiMessages = messages.filter { $0.sender == .ai }.count
                let lastMessageTime = messages.max(by: { $0.timestamp < $1.timestamp })?.timestamp
                
                // 计算平均响应时间（简化版本）
                let averageResponseTime: TimeInterval = 1.5
                
                let statistics = MessageStatistics(
                    totalMessages: totalMessages,
                    userMessages: userMessages,
                    aiMessages: aiMessages,
                    averageResponseTime: averageResponseTime,
                    lastMessageTime: lastMessageTime
                )
                
                promise(.success(statistics))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    // MARK: - Sync
    func syncConversations() -> AnyPublisher<Void, Error> {
        return aiService.getConversations()
            .flatMap { [weak self] remoteConversations -> AnyPublisher<Void, Error> in
                guard let self = self else {
                    return Fail(error: ChatRepositoryError.repositoryDeallocated)
                        .eraseToAnyPublisher()
                }
                
                // 这里可以实现更复杂的同步逻辑
                return Just(())
                    .setFailureType(to: Error.self)
                    .eraseToAnyPublisher()
            }
            .eraseToAnyPublisher()
    }
    
    func syncMessages(conversationId: String) -> AnyPublisher<Void, Error> {
        return aiService.getMessages(conversationId: conversationId, page: 1, limit: 100)
            .flatMap { [weak self] messagesData -> AnyPublisher<Void, Error> in
                guard let self = self else {
                    return Fail(error: ChatRepositoryError.repositoryDeallocated)
                        .eraseToAnyPublisher()
                }
                
                // 这里可以实现更复杂的同步逻辑
                return Just(())
                    .setFailureType(to: Error.self)
                    .eraseToAnyPublisher()
            }
            .eraseToAnyPublisher()
    }
}

// MARK: - Private Helper Methods
private extension ChatRepository {
    func saveConversationLocally(id: String, title: String, aiPersonality: AIPersonality, createdAt: Date) -> AnyPublisher<Conversation, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let entity = ConversationEntity(context: self.context)
            entity.id = id
            entity.title = title
            entity.aiPersonality = aiPersonality.rawValue
            entity.createdAt = createdAt
            entity.updatedAt = createdAt
            entity.isArchived = false
            
            do {
                try self.context.save()
                if let conversation = entity.toConversation() {
                    promise(.success(conversation))
                } else {
                    promise(.failure(ChatRepositoryError.conversionError))
                }
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func deleteConversationLocally(id: String) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", id)
            
            do {
                let entities = try self.context.fetch(request)
                entities.forEach { self.context.delete($0) }
                try self.context.save()
                promise(.success(()))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func saveMessageLocally(
        conversationId: String,
        content: String,
        messageType: MessageType,
        sender: MessageSender,
        status: MessageStatus,
        metadata: [String: AnyCodable]? = nil
    ) -> AnyPublisher<ChatMessage, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let entity = ChatMessageEntity(context: self.context)
            entity.id = UUID().uuidString
            entity.conversationId = conversationId
            entity.content = content
            entity.messageType = messageType.rawValue
            entity.sender = sender.rawValue
            entity.status = status.rawValue
            entity.timestamp = Date()
            
            if let metadata = metadata {
                do {
                    let data = try JSONEncoder().encode(metadata)
                    entity.metadata = data
                } catch {
                    print("Failed to encode metadata: \(error)")
                }
            }
            
            do {
                try self.context.save()
                if let message = entity.toChatMessage() {
                    promise(.success(message))
                } else {
                    promise(.failure(ChatRepositoryError.conversionError))
                }
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func getMessageLocally(id: String) -> AnyPublisher<ChatMessage?, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", id)
            request.fetchLimit = 1
            
            do {
                let entities = try self.context.fetch(request)
                let message = entities.first?.toChatMessage()
                promise(.success(message))
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    func updateMessageStatusLocally(messageId: String, status: MessageStatus) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(ChatRepositoryError.repositoryDeallocated))
                return
            }
            
            let request: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
            request.predicate = NSPredicate(format: "id == %@", messageId)
            request.fetchLimit = 1
            
            do {
                if let entity = try self.context.fetch(request).first {
                    entity.status = status.rawValue
                    try self.context.save()
                    promise(.success(()))
                } else {
                    promise(.failure(ChatRepositoryError.messageNotFound))
                }
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
}

// MARK: - Chat Repository Error
enum ChatRepositoryError: LocalizedError {
    case repositoryDeallocated
    case conversationNotFound
    case messageNotFound
    case conversionError
    case noAIResponse
    
    var errorDescription: String? {
        switch self {
        case .repositoryDeallocated:
            return "Repository已被释放"
        case .conversationNotFound:
            return "未找到对话"
        case .messageNotFound:
            return "未找到消息"
        case .conversionError:
            return "数据转换错误"
        case .noAIResponse:
            return "没有AI回复"
        }
    }
}