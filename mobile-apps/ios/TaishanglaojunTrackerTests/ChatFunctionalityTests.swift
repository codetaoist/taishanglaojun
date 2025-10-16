//
//  ChatFunctionalityTests.swift
//  TaishanglaojunTrackerTests
//
//  Created by Taishanglaojun Team
//

import XCTest
import Combine
import CoreData
@testable import TaishanglaojunTracker

class ChatFunctionalityTests: XCTestCase {
    
    var chatRepository: ChatRepository!
    var chatViewModel: ChatViewModel!
    var mockAIService: MockAIService!
    var testContext: NSManagedObjectContext!
    var cancellables: Set<AnyCancellable>!
    
    override func setUpWithError() throws {
        try super.setUpWithError()
        
        // 创建内存中的Core Data栈用于测试
        let persistentContainer = NSPersistentContainer(name: "TaishanglaojunTracker")
        let description = NSPersistentStoreDescription()
        description.type = NSInMemoryStoreType
        persistentContainer.persistentStoreDescriptions = [description]
        
        persistentContainer.loadPersistentStores { _, error in
            if let error = error {
                fatalError("Failed to load test store: \(error)")
            }
        }
        
        testContext = persistentContainer.viewContext
        
        // 创建模拟AI服务
        mockAIService = MockAIService()
        
        // 创建测试用的数据服务
        let testDataService = MockDataService(context: testContext)
        
        // 初始化repository和view model
        chatRepository = ChatRepository(aiService: mockAIService, dataService: testDataService)
        chatViewModel = ChatViewModel(chatRepository: chatRepository)
        
        cancellables = Set<AnyCancellable>()
    }
    
    override func tearDownWithError() throws {
        cancellables = nil
        chatViewModel = nil
        chatRepository = nil
        mockAIService = nil
        testContext = nil
        try super.tearDownWithError()
    }
    
    // MARK: - Conversation Tests
    
    func testCreateConversation() throws {
        // Given
        let expectation = XCTestExpectation(description: "Create conversation")
        let title = "Test Conversation"
        let personality = AIPersonality.creative
        
        mockAIService.createConversationResult = .success(
            CreateConversationResponse(
                conversationId: "test-id",
                createdAt: Date().timeIntervalSince1970
            )
        )
        
        // When
        chatRepository.createConversation(title: title, aiPersonality: personality)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        XCTFail("Expected success, got error: \(error)")
                    }
                },
                receiveValue: { conversation in
                    // Then
                    XCTAssertEqual(conversation.title, title)
                    XCTAssertEqual(conversation.aiPersonality, personality)
                    XCTAssertFalse(conversation.isArchived)
                    expectation.fulfill()
                }
            )
            .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
    }
    
    func testGetConversations() throws {
        // Given
        let expectation = XCTestExpectation(description: "Get conversations")
        
        // 先创建一些测试对话
        createTestConversation(id: "conv1", title: "Conversation 1")
        createTestConversation(id: "conv2", title: "Conversation 2")
        
        // When
        chatRepository.getConversations()
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        XCTFail("Expected success, got error: \(error)")
                    }
                },
                receiveValue: { conversations in
                    // Then
                    XCTAssertEqual(conversations.count, 2)
                    XCTAssertTrue(conversations.contains { $0.title == "Conversation 1" })
                    XCTAssertTrue(conversations.contains { $0.title == "Conversation 2" })
                    expectation.fulfill()
                }
            )
            .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
    }
    
    func testDeleteConversation() throws {
        // Given
        let expectation = XCTestExpectation(description: "Delete conversation")
        let conversationId = "test-conv-id"
        
        createTestConversation(id: conversationId, title: "Test Conversation")
        mockAIService.deleteConversationResult = .success(())
        
        // When
        chatRepository.deleteConversation(id: conversationId)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        XCTFail("Expected success, got error: \(error)")
                    } else {
                        expectation.fulfill()
                    }
                },
                receiveValue: { _ in }
            )
            .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
        
        // Then - 验证对话已被删除
        let fetchRequest: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
        fetchRequest.predicate = NSPredicate(format: "id == %@", conversationId)
        let conversations = try testContext.fetch(fetchRequest)
        XCTAssertTrue(conversations.isEmpty)
    }
    
    func testArchiveConversation() throws {
        // Given
        let expectation = XCTestExpectation(description: "Archive conversation")
        let conversationId = "test-conv-id"
        
        createTestConversation(id: conversationId, title: "Test Conversation")
        
        // When
        chatRepository.archiveConversation(id: conversationId)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        XCTFail("Expected success, got error: \(error)")
                    } else {
                        expectation.fulfill()
                    }
                },
                receiveValue: { _ in }
            )
            .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
        
        // Then - 验证对话已被归档
        let fetchRequest: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
        fetchRequest.predicate = NSPredicate(format: "id == %@", conversationId)
        let conversations = try testContext.fetch(fetchRequest)
        XCTAssertEqual(conversations.count, 1)
        XCTAssertTrue(conversations.first?.isArchived ?? false)
    }
    
    // MARK: - Message Tests
    
    func testSendMessage() throws {
        // Given
        let expectation = XCTestExpectation(description: "Send message")
        let conversationId = "test-conv-id"
        let messageContent = "Hello, AI!"
        
        createTestConversation(id: conversationId, title: "Test Conversation")
        
        mockAIService.sendMessageResult = .success(
            SendMessageResponse(
                message: ChatMessage(
                    id: "ai-msg-id",
                    conversationId: conversationId,
                    content: "Hello! How can I help you?",
                    messageType: .text,
                    sender: .ai,
                    timestamp: Date(),
                    status: .sent
                )
            )
        )
        
        // When
        chatRepository.sendMessage(
            conversationId: conversationId,
            content: messageContent,
            messageType: .text
        )
        .sink(
            receiveCompletion: { completion in
                if case .failure(let error) = completion {
                    XCTFail("Expected success, got error: \(error)")
                }
            },
            receiveValue: { message in
                // Then
                XCTAssertEqual(message.sender, .ai)
                XCTAssertEqual(message.conversationId, conversationId)
                XCTAssertEqual(message.status, .sent)
                expectation.fulfill()
            }
        )
        .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
    }
    
    func testSendMessageFailure() throws {
        // Given
        let expectation = XCTestExpectation(description: "Send message failure")
        let conversationId = "test-conv-id"
        let messageContent = "Hello, AI!"
        
        createTestConversation(id: conversationId, title: "Test Conversation")
        mockAIService.sendMessageResult = .failure(AIServiceError.networkError)
        
        // When
        chatRepository.sendMessage(
            conversationId: conversationId,
            content: messageContent,
            messageType: .text
        )
        .sink(
            receiveCompletion: { completion in
                if case .failure = completion {
                    expectation.fulfill()
                } else {
                    XCTFail("Expected failure")
                }
            },
            receiveValue: { _ in
                XCTFail("Expected failure, got success")
            }
        )
        .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
    }
    
    func testGetMessages() throws {
        // Given
        let expectation = XCTestExpectation(description: "Get messages")
        let conversationId = "test-conv-id"
        
        createTestConversation(id: conversationId, title: "Test Conversation")
        createTestMessage(id: "msg1", conversationId: conversationId, content: "Message 1")
        createTestMessage(id: "msg2", conversationId: conversationId, content: "Message 2")
        
        // When
        chatRepository.getMessages(conversationId: conversationId)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        XCTFail("Expected success, got error: \(error)")
                    }
                },
                receiveValue: { messages in
                    // Then
                    XCTAssertEqual(messages.count, 2)
                    XCTAssertTrue(messages.contains { $0.content == "Message 1" })
                    XCTAssertTrue(messages.contains { $0.content == "Message 2" })
                    expectation.fulfill()
                }
            )
            .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
    }
    
    func testSearchMessages() throws {
        // Given
        let expectation = XCTestExpectation(description: "Search messages")
        let conversationId = "test-conv-id"
        
        createTestConversation(id: conversationId, title: "Test Conversation")
        createTestMessage(id: "msg1", conversationId: conversationId, content: "Hello world")
        createTestMessage(id: "msg2", conversationId: conversationId, content: "Goodbye world")
        createTestMessage(id: "msg3", conversationId: conversationId, content: "Test message")
        
        // When
        chatRepository.searchMessages(query: "world", conversationId: conversationId)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        XCTFail("Expected success, got error: \(error)")
                    }
                },
                receiveValue: { messages in
                    // Then
                    XCTAssertEqual(messages.count, 2)
                    XCTAssertTrue(messages.allSatisfy { $0.content.contains("world") })
                    expectation.fulfill()
                }
            )
            .store(in: &cancellables)
        
        wait(for: [expectation], timeout: 5.0)
    }
    
    // MARK: - ViewModel Tests
    
    @MainActor
    func testViewModelCreateConversation() async throws {
        // Given
        mockAIService.createConversationResult = .success(
            CreateConversationResponse(
                conversationId: "new-conv-id",
                createdAt: Date().timeIntervalSince1970
            )
        )
        
        chatViewModel.newConversationTitle = "New Test Conversation"
        chatViewModel.selectedAIPersonality = .creative
        
        // When
        chatViewModel.createConversation()
        
        // Wait for async operation
        try await Task.sleep(nanoseconds: 100_000_000) // 0.1 seconds
        
        // Then
        XCTAssertFalse(chatViewModel.conversations.isEmpty)
        XCTAssertEqual(chatViewModel.currentConversation?.title, "New Test Conversation")
        XCTAssertEqual(chatViewModel.currentConversation?.aiPersonality, .creative)
        XCTAssertFalse(chatViewModel.showNewConversationDialog)
        XCTAssertTrue(chatViewModel.newConversationTitle.isEmpty)
    }
    
    @MainActor
    func testViewModelMessageValidation() {
        // Given
        chatViewModel.messageText = ""
        
        // Then
        XCTAssertFalse(chatViewModel.canSendMessage)
        
        // Given
        chatViewModel.messageText = "Valid message"
        
        // Then
        XCTAssertTrue(chatViewModel.canSendMessage)
        
        // Given - 超长消息
        chatViewModel.messageText = String(repeating: "a", count: 2001)
        
        // Then
        XCTAssertFalse(chatViewModel.canSendMessage)
        XCTAssertTrue(chatViewModel.isMessageTooLong)
    }
    
    // MARK: - Helper Methods
    
    private func createTestConversation(id: String, title: String) {
        let entity = ConversationEntity(context: testContext)
        entity.id = id
        entity.title = title
        entity.createdAt = Date().timeIntervalSince1970
        entity.updatedAt = Date().timeIntervalSince1970
        entity.aiPersonality = AIPersonality.default.rawValue
        entity.isArchived = false
        entity.messageCount = 0
        
        try! testContext.save()
    }
    
    private func createTestMessage(id: String, conversationId: String, content: String) {
        let entity = ChatMessageEntity(context: testContext)
        entity.id = id
        entity.conversationId = conversationId
        entity.content = content
        entity.messageType = MessageType.text.rawValue
        entity.sender = MessageSender.user.rawValue
        entity.timestamp = Date().timeIntervalSince1970
        entity.status = MessageStatus.sent.rawValue
        
        try! testContext.save()
    }
}

// MARK: - Mock Classes

class MockAIService: AIServiceProtocol {
    var createConversationResult: Result<CreateConversationResponse, Error> = .failure(AIServiceError.networkError)
    var sendMessageResult: Result<SendMessageResponse, Error> = .failure(AIServiceError.networkError)
    var deleteConversationResult: Result<Void, Error> = .failure(AIServiceError.networkError)
    
    func createConversation(_ request: CreateConversationRequest) -> AnyPublisher<CreateConversationResponse, Error> {
        return createConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func sendMessage(_ request: SendMessageRequest) -> AnyPublisher<SendMessageResponse, Error> {
        return sendMessageResult.publisher.eraseToAnyPublisher()
    }
    
    func deleteConversation(conversationId: String) -> AnyPublisher<Void, Error> {
        return deleteConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func getChatHistory(conversationId: String, limit: Int, before: Date?) -> AnyPublisher<ChatHistoryResponse, Error> {
        return Fail(error: AIServiceError.networkError).eraseToAnyPublisher()
    }
}

class MockDataService: DataService {
    private let context: NSManagedObjectContext
    
    init(context: NSManagedObjectContext) {
        self.context = context
        super.init()
    }
    
    override var persistentContainer: NSPersistentContainer {
        let container = NSPersistentContainer(name: "TaishanglaojunTracker")
        container.viewContext = context
        return container
    }
}

enum AIServiceError: Error {
    case networkError
    case invalidResponse
    case unauthorized
}

extension Result {
    var publisher: AnyPublisher<Success, Failure> {
        switch self {
        case .success(let value):
            return Just(value)
                .setFailureType(to: Failure.self)
                .eraseToAnyPublisher()
        case .failure(let error):
            return Fail(error: error)
                .eraseToAnyPublisher()
        }
    }
}