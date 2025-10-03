//
//  ChatIntegrationTests.swift
//  TaishanglaojunTrackerTests
//
//  Created by Taishanglaojun Team
//

import XCTest
import SwiftUI
import Combine
import CoreData
@testable import TaishanglaojunTracker

class ChatIntegrationTests: XCTestCase {
    
    var chatViewModel: ChatViewModel!
    var mockChatRepository: MockChatRepository!
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
        
        // 创建模拟repository
        mockChatRepository = MockChatRepository()
        
        // 初始化view model
        chatViewModel = ChatViewModel(chatRepository: mockChatRepository)
        
        cancellables = Set<AnyCancellable>()
    }
    
    override func tearDownWithError() throws {
        cancellables = nil
        chatViewModel = nil
        mockChatRepository = nil
        testContext = nil
        try super.tearDownWithError()
    }
    
    // MARK: - UI State Tests
    
    @MainActor
    func testInitialUIState() async {
        // Given - 初始状态
        
        // Then
        XCTAssertTrue(chatViewModel.conversations.isEmpty)
        XCTAssertNil(chatViewModel.currentConversation)
        XCTAssertTrue(chatViewModel.messages.isEmpty)
        XCTAssertTrue(chatViewModel.messageText.isEmpty)
        XCTAssertFalse(chatViewModel.isLoading)
        XCTAssertFalse(chatViewModel.isSending)
        XCTAssertFalse(chatViewModel.showError)
        XCTAssertFalse(chatViewModel.showConversationList)
        XCTAssertFalse(chatViewModel.showAIPersonalityPicker)
        XCTAssertFalse(chatViewModel.showNewConversationDialog)
    }
    
    @MainActor
    func testCreateNewConversationFlow() async throws {
        // Given
        let testConversation = createTestConversation(id: "new-conv", title: "New Conversation")
        mockChatRepository.createConversationResult = .success(testConversation)
        
        // When - 显示新对话对话框
        chatViewModel.showNewConversationDialog = true
        chatViewModel.newConversationTitle = "Test Conversation"
        chatViewModel.selectedAIPersonality = .creative
        
        // Then
        XCTAssertTrue(chatViewModel.showNewConversationDialog)
        XCTAssertEqual(chatViewModel.newConversationTitle, "Test Conversation")
        XCTAssertEqual(chatViewModel.selectedAIPersonality, .creative)
        
        // When - 创建对话
        chatViewModel.createConversation()
        
        // Wait for async operation
        try await Task.sleep(nanoseconds: 100_000_000) // 0.1 seconds
        
        // Then
        XCTAssertFalse(chatViewModel.conversations.isEmpty)
        XCTAssertEqual(chatViewModel.currentConversation?.id, "new-conv")
        XCTAssertFalse(chatViewModel.showNewConversationDialog)
        XCTAssertTrue(chatViewModel.newConversationTitle.isEmpty)
    }
    
    @MainActor
    func testSendMessageFlow() async throws {
        // Given
        let testConversation = createTestConversation(id: "test-conv", title: "Test")
        let testMessage = createTestMessage(id: "ai-msg", conversationId: "test-conv", content: "AI Response")
        
        mockChatRepository.getConversationsResult = .success([testConversation])
        mockChatRepository.sendMessageResult = .success(testMessage)
        mockChatRepository.getMessagesResult = .success([])
        
        // 加载对话
        chatViewModel.loadConversations()
        try await Task.sleep(nanoseconds: 50_000_000)
        
        // When - 输入消息
        chatViewModel.messageText = "Hello, AI!"
        
        // Then
        XCTAssertTrue(chatViewModel.canSendMessage)
        XCTAssertFalse(chatViewModel.isMessageTooLong)
        
        // When - 发送消息
        await chatViewModel.sendMessage()
        
        // Wait for async operation
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertTrue(chatViewModel.messageText.isEmpty)
        XCTAssertFalse(chatViewModel.isSending)
    }
    
    @MainActor
    func testMessageValidationFlow() async {
        // Given - 空消息
        chatViewModel.messageText = ""
        
        // Then
        XCTAssertFalse(chatViewModel.canSendMessage)
        
        // Given - 只有空格的消息
        chatViewModel.messageText = "   "
        
        // Then
        XCTAssertFalse(chatViewModel.canSendMessage)
        
        // Given - 有效消息
        chatViewModel.messageText = "Valid message"
        
        // Then
        XCTAssertTrue(chatViewModel.canSendMessage)
        XCTAssertEqual(chatViewModel.messageCountText, "13/2000")
        
        // Given - 超长消息
        chatViewModel.messageText = String(repeating: "a", count: 2001)
        
        // Then
        XCTAssertFalse(chatViewModel.canSendMessage)
        XCTAssertTrue(chatViewModel.isMessageTooLong)
        XCTAssertEqual(chatViewModel.messageCountText, "2001/2000")
    }
    
    @MainActor
    func testLoadingStatesFlow() async throws {
        // Given
        let testConversation = createTestConversation(id: "test-conv", title: "Test")
        mockChatRepository.getConversationsResult = .success([testConversation])
        
        // When - 开始加载
        XCTAssertFalse(chatViewModel.isLoading)
        
        chatViewModel.loadConversations()
        
        // Then - 加载状态应该被设置
        // 注意：由于异步操作，isLoading可能很快变为false
        
        // Wait for completion
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertFalse(chatViewModel.isLoading)
        XCTAssertFalse(chatViewModel.conversations.isEmpty)
    }
    
    @MainActor
    func testErrorHandlingFlow() async throws {
        // Given
        mockChatRepository.getConversationsResult = .failure(ChatRepositoryError.networkError)
        
        // When
        chatViewModel.loadConversations()
        
        // Wait for async operation
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertFalse(chatViewModel.isLoading)
        XCTAssertTrue(chatViewModel.showError)
        XCTAssertNotNil(chatViewModel.errorMessage)
    }
    
    @MainActor
    func testConversationListFlow() async throws {
        // Given
        let conversations = [
            createTestConversation(id: "conv1", title: "Conversation 1"),
            createTestConversation(id: "conv2", title: "Conversation 2")
        ]
        mockChatRepository.getConversationsResult = .success(conversations)
        
        // When - 加载对话列表
        chatViewModel.loadConversations()
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertEqual(chatViewModel.conversations.count, 2)
        XCTAssertEqual(chatViewModel.currentConversation?.id, "conv1") // 应该选择第一个
        
        // When - 显示对话列表
        chatViewModel.showConversationList = true
        
        // Then
        XCTAssertTrue(chatViewModel.showConversationList)
        
        // When - 选择另一个对话
        chatViewModel.selectConversation(conversations[1])
        
        // Then
        XCTAssertEqual(chatViewModel.currentConversation?.id, "conv2")
        XCTAssertFalse(chatViewModel.showConversationList)
    }
    
    @MainActor
    func testAIPersonalityPickerFlow() async {
        // Given
        chatViewModel.selectedAIPersonality = .default
        
        // When - 显示AI人格选择器
        chatViewModel.showAIPersonalityPicker = true
        
        // Then
        XCTAssertTrue(chatViewModel.showAIPersonalityPicker)
        
        // When - 选择新的AI人格
        chatViewModel.selectedAIPersonality = .creative
        chatViewModel.showAIPersonalityPicker = false
        
        // Then
        XCTAssertEqual(chatViewModel.selectedAIPersonality, .creative)
        XCTAssertFalse(chatViewModel.showAIPersonalityPicker)
    }
    
    @MainActor
    func testSearchFlow() async throws {
        // Given
        let testMessages = [
            createTestMessage(id: "msg1", conversationId: "conv1", content: "Hello world"),
            createTestMessage(id: "msg2", conversationId: "conv1", content: "Goodbye world")
        ]
        mockChatRepository.searchMessagesResult = .success(testMessages)
        
        // When - 输入搜索文本
        chatViewModel.searchText = "world"
        
        // Wait for debounce
        try await Task.sleep(nanoseconds: 600_000_000) // 0.6 seconds
        
        // Then
        XCTAssertEqual(chatViewModel.searchResults.count, 2)
        XCTAssertTrue(chatViewModel.searchResults.allSatisfy { $0.content.contains("world") })
        
        // When - 清空搜索
        chatViewModel.searchText = ""
        
        // Wait for debounce
        try await Task.sleep(nanoseconds: 600_000_000)
        
        // Then
        XCTAssertTrue(chatViewModel.searchResults.isEmpty)
    }
    
    @MainActor
    func testDeleteConversationFlow() async throws {
        // Given
        let conversations = [
            createTestConversation(id: "conv1", title: "Conversation 1"),
            createTestConversation(id: "conv2", title: "Conversation 2")
        ]
        mockChatRepository.getConversationsResult = .success(conversations)
        mockChatRepository.deleteConversationResult = .success(())
        
        // 加载对话
        chatViewModel.loadConversations()
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // When - 删除当前对话
        let conversationToDelete = chatViewModel.conversations.first!
        chatViewModel.deleteConversation(conversationToDelete)
        
        // Wait for async operation
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertEqual(chatViewModel.conversations.count, 1)
        XCTAssertNotEqual(chatViewModel.currentConversation?.id, conversationToDelete.id)
    }
    
    @MainActor
    func testArchiveConversationFlow() async throws {
        // Given
        let conversations = [
            createTestConversation(id: "conv1", title: "Conversation 1"),
            createTestConversation(id: "conv2", title: "Conversation 2")
        ]
        mockChatRepository.getConversationsResult = .success(conversations)
        mockChatRepository.archiveConversationResult = .success(())
        
        // 加载对话
        chatViewModel.loadConversations()
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // When - 归档当前对话
        let conversationToArchive = chatViewModel.conversations.first!
        chatViewModel.archiveConversation(conversationToArchive)
        
        // Wait for async operation
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertEqual(chatViewModel.conversations.count, 1)
        XCTAssertNotEqual(chatViewModel.currentConversation?.id, conversationToArchive.id)
    }
    
    @MainActor
    func testMessageStatisticsFlow() async throws {
        // Given
        let testConversation = createTestConversation(id: "test-conv", title: "Test")
        let testStatistics = MessageStatistics(
            totalMessages: 10,
            userMessages: 5,
            aiMessages: 5,
            averageResponseTime: 1.5,
            lastMessageTime: Date()
        )
        
        mockChatRepository.getConversationsResult = .success([testConversation])
        mockChatRepository.getMessagesResult = .success([])
        mockChatRepository.getMessageStatisticsResult = .success(testStatistics)
        
        // When - 选择对话（这会触发统计信息加载）
        chatViewModel.loadConversations()
        try await Task.sleep(nanoseconds: 100_000_000)
        
        // Then
        XCTAssertNotNil(chatViewModel.messageStatistics)
        XCTAssertEqual(chatViewModel.messageStatistics?.totalMessages, 10)
        XCTAssertEqual(chatViewModel.messageStatistics?.userMessages, 5)
        XCTAssertEqual(chatViewModel.messageStatistics?.aiMessages, 5)
    }
    
    // MARK: - Helper Methods
    
    private func createTestConversation(id: String, title: String) -> Conversation {
        return Conversation(
            id: id,
            title: title,
            createdAt: Date(),
            updatedAt: Date(),
            lastMessageId: nil,
            messageCount: 0,
            isArchived: false,
            aiPersonality: .default
        )
    }
    
    private func createTestMessage(id: String, conversationId: String, content: String) -> ChatMessage {
        return ChatMessage(
            id: id,
            conversationId: conversationId,
            content: content,
            messageType: .text,
            sender: .ai,
            timestamp: Date(),
            status: .sent
        )
    }
}

// MARK: - Mock Repository

class MockChatRepository: ChatRepositoryProtocol {
    var createConversationResult: Result<Conversation, Error> = .failure(ChatRepositoryError.networkError)
    var getConversationsResult: Result<[Conversation], Error> = .failure(ChatRepositoryError.networkError)
    var getConversationResult: Result<Conversation?, Error> = .failure(ChatRepositoryError.networkError)
    var updateConversationResult: Result<Conversation, Error> = .failure(ChatRepositoryError.networkError)
    var deleteConversationResult: Result<Void, Error> = .failure(ChatRepositoryError.networkError)
    var archiveConversationResult: Result<Void, Error> = .failure(ChatRepositoryError.networkError)
    var sendMessageResult: Result<ChatMessage, Error> = .failure(ChatRepositoryError.networkError)
    var getMessagesResult: Result<[ChatMessage], Error> = .failure(ChatRepositoryError.networkError)
    var retryMessageResult: Result<ChatMessage, Error> = .failure(ChatRepositoryError.networkError)
    var deleteMessageResult: Result<Void, Error> = .failure(ChatRepositoryError.networkError)
    var updateMessageStatusResult: Result<Void, Error> = .failure(ChatRepositoryError.networkError)
    var searchMessagesResult: Result<[ChatMessage], Error> = .failure(ChatRepositoryError.networkError)
    var getMessageStatisticsResult: Result<MessageStatistics, Error> = .failure(ChatRepositoryError.networkError)
    var syncConversationsResult: Result<Void, Error> = .failure(ChatRepositoryError.networkError)
    var syncMessagesResult: Result<Void, Error> = .failure(ChatRepositoryError.networkError)
    
    func createConversation(title: String, aiPersonality: AIPersonality) -> AnyPublisher<Conversation, Error> {
        return createConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func getConversations() -> AnyPublisher<[Conversation], Error> {
        return getConversationsResult.publisher.eraseToAnyPublisher()
    }
    
    func getConversation(id: String) -> AnyPublisher<Conversation?, Error> {
        return getConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func updateConversation(_ conversation: Conversation) -> AnyPublisher<Conversation, Error> {
        return updateConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func deleteConversation(id: String) -> AnyPublisher<Void, Error> {
        return deleteConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func archiveConversation(id: String) -> AnyPublisher<Void, Error> {
        return archiveConversationResult.publisher.eraseToAnyPublisher()
    }
    
    func sendMessage(conversationId: String, content: String, messageType: MessageType) -> AnyPublisher<ChatMessage, Error> {
        return sendMessageResult.publisher.eraseToAnyPublisher()
    }
    
    func getMessages(conversationId: String, limit: Int, before: Date?) -> AnyPublisher<[ChatMessage], Error> {
        return getMessagesResult.publisher.eraseToAnyPublisher()
    }
    
    func retryMessage(messageId: String) -> AnyPublisher<ChatMessage, Error> {
        return retryMessageResult.publisher.eraseToAnyPublisher()
    }
    
    func deleteMessage(messageId: String) -> AnyPublisher<Void, Error> {
        return deleteMessageResult.publisher.eraseToAnyPublisher()
    }
    
    func updateMessageStatus(messageId: String, status: MessageStatus) -> AnyPublisher<Void, Error> {
        return updateMessageStatusResult.publisher.eraseToAnyPublisher()
    }
    
    func searchMessages(query: String, conversationId: String?) -> AnyPublisher<[ChatMessage], Error> {
        return searchMessagesResult.publisher.eraseToAnyPublisher()
    }
    
    func getMessageStatistics(conversationId: String) -> AnyPublisher<MessageStatistics, Error> {
        return getMessageStatisticsResult.publisher.eraseToAnyPublisher()
    }
    
    func syncConversations() -> AnyPublisher<Void, Error> {
        return syncConversationsResult.publisher.eraseToAnyPublisher()
    }
    
    func syncMessages(conversationId: String) -> AnyPublisher<Void, Error> {
        return syncMessagesResult.publisher.eraseToAnyPublisher()
    }
}

enum ChatRepositoryError: Error {
    case networkError
    case conversationNotFound
    case messageNotFound
    case repositoryDeallocated
    case conversionError
    case noAIResponse
}