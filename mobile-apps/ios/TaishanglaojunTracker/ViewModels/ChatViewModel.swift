import Foundation
import Combine
import SwiftUI

// MARK: - Chat View Model
@MainActor
class ChatViewModel: ObservableObject {
    // MARK: - Published Properties
    @Published var conversations: [Conversation] = []
    @Published var currentConversation: Conversation?
    @Published var messages: [ChatMessage] = []
    @Published var messageText: String = ""
    @Published var isLoading: Bool = false
    @Published var isSending: Bool = false
    @Published var errorMessage: String?
    @Published var showError: Bool = false
    @Published var searchText: String = ""
    @Published var searchResults: [ChatMessage] = []
    @Published var isSearching: Bool = false
    @Published var showConversationList: Bool = false
    @Published var showAIPersonalityPicker: Bool = false
    @Published var selectedAIPersonality: AIPersonality = .default
    @Published var showNewConversationDialog: Bool = false
    @Published var newConversationTitle: String = ""
    @Published var messageStatistics: MessageStatistics?
    
    // MARK: - Private Properties
    private let chatRepository: ChatRepositoryProtocol
    private var cancellables = Set<AnyCancellable>()
    private let maxMessageLength = 2000
    private var lastMessageDate: Date?
    
    // MARK: - Computed Properties
    var canSendMessage: Bool {
        !messageText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
        !isSending &&
        messageText.count <= maxMessageLength
    }
    
    var messageCountText: String {
        "\(messageText.count)/\(maxMessageLength)"
    }
    
    var isMessageTooLong: Bool {
        messageText.count > maxMessageLength
    }
    
    var hasConversations: Bool {
        !conversations.isEmpty
    }
    
    var hasMessages: Bool {
        !messages.isEmpty
    }
    
    // MARK: - Initialization
    init(chatRepository: ChatRepositoryProtocol = ChatRepository.shared) {
        self.chatRepository = chatRepository
        setupBindings()
        loadConversations()
    }
    
    // MARK: - Setup
    private func setupBindings() {
        // 搜索文本变化时自动搜索
        $searchText
            .debounce(for: .milliseconds(500), scheduler: DispatchQueue.main)
            .removeDuplicates()
            .sink { [weak self] searchText in
                if !searchText.isEmpty {
                    self?.searchMessages(query: searchText)
                } else {
                    self?.searchResults = []
                }
            }
            .store(in: &cancellables)
        
        // 监听当前对话变化，加载消息
        $currentConversation
            .compactMap { $0 }
            .sink { [weak self] conversation in
                self?.loadMessages(for: conversation.id)
                self?.loadMessageStatistics(for: conversation.id)
            }
            .store(in: &cancellables)
    }
    
    // MARK: - Conversation Management
    func loadConversations() {
        isLoading = true
        
        chatRepository.getConversations()
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    self?.isLoading = false
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] conversations in
                    self?.conversations = conversations
                    
                    // 如果没有当前对话，选择第一个
                    if self?.currentConversation == nil && !conversations.isEmpty {
                        self?.currentConversation = conversations.first
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    func createConversation() {
        let title = newConversationTitle.isEmpty ? "新对话" : newConversationTitle
        isLoading = true
        
        chatRepository.createConversation(title: title, aiPersonality: selectedAIPersonality)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    self?.isLoading = false
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] conversation in
                    self?.conversations.insert(conversation, at: 0)
                    self?.currentConversation = conversation
                    self?.newConversationTitle = ""
                    self?.showNewConversationDialog = false
                }
            )
            .store(in: &cancellables)
    }
    
    func selectConversation(_ conversation: Conversation) {
        currentConversation = conversation
        showConversationList = false
    }
    
    func deleteConversation(_ conversation: Conversation) {
        chatRepository.deleteConversation(id: conversation.id)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] _ in
                    self?.conversations.removeAll { $0.id == conversation.id }
                    
                    // 如果删除的是当前对话，选择下一个
                    if self?.currentConversation?.id == conversation.id {
                        self?.currentConversation = self?.conversations.first
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    func archiveConversation(_ conversation: Conversation) {
        chatRepository.archiveConversation(id: conversation.id)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] _ in
                    self?.conversations.removeAll { $0.id == conversation.id }
                    
                    // 如果归档的是当前对话，选择下一个
                    if self?.currentConversation?.id == conversation.id {
                        self?.currentConversation = self?.conversations.first
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    func updateConversationTitle(_ conversation: Conversation, newTitle: String) {
        var updatedConversation = conversation
        updatedConversation.title = newTitle
        
        chatRepository.updateConversation(updatedConversation)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] conversation in
                    if let index = self?.conversations.firstIndex(where: { $0.id == conversation.id }) {
                        self?.conversations[index] = conversation
                    }
                    
                    if self?.currentConversation?.id == conversation.id {
                        self?.currentConversation = conversation
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - Message Management
    func loadMessages(for conversationId: String, loadMore: Bool = false) {
        let beforeDate = loadMore ? lastMessageDate : nil
        
        chatRepository.getMessages(conversationId: conversationId, limit: 50, before: beforeDate)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] newMessages in
                    if loadMore {
                        self?.messages = newMessages + (self?.messages ?? [])
                    } else {
                        self?.messages = newMessages
                    }
                    
                    if let firstMessage = newMessages.first {
                        self?.lastMessageDate = firstMessage.timestamp
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    func sendMessage() {
        guard let conversationId = currentConversation?.id,
              canSendMessage else { return }
        
        let content = messageText.trimmingCharacters(in: .whitespacesAndNewlines)
        messageText = ""
        isSending = true
        
        chatRepository.sendMessage(
            conversationId: conversationId,
            content: content,
            messageType: .text
        )
        .receive(on: DispatchQueue.main)
        .sink(
            receiveCompletion: { [weak self] completion in
                self?.isSending = false
                if case .failure(let error) = completion {
                    self?.handleError(error)
                }
            },
            receiveValue: { [weak self] message in
                // 消息已经通过Repository保存到本地，重新加载消息列表
                if let conversationId = self?.currentConversation?.id {
                    self?.loadMessages(for: conversationId)
                }
                
                // 更新对话的最后更新时间
                self?.updateConversationInList()
            }
        )
        .store(in: &cancellables)
    }
    
    func retryMessage(_ message: ChatMessage) {
        chatRepository.retryMessage(messageId: message.id)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] _ in
                    // 重新加载消息列表
                    if let conversationId = self?.currentConversation?.id {
                        self?.loadMessages(for: conversationId)
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    func deleteMessage(_ message: ChatMessage) {
        chatRepository.deleteMessage(messageId: message.id)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] _ in
                    self?.messages.removeAll { $0.id == message.id }
                }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - Search
    func searchMessages(query: String) {
        guard !query.isEmpty else {
            searchResults = []
            return
        }
        
        isSearching = true
        
        chatRepository.searchMessages(
            query: query,
            conversationId: currentConversation?.id
        )
        .receive(on: DispatchQueue.main)
        .sink(
            receiveCompletion: { [weak self] completion in
                self?.isSearching = false
                if case .failure(let error) = completion {
                    self?.handleError(error)
                }
            },
            receiveValue: { [weak self] results in
                self?.searchResults = results
            }
        )
        .store(in: &cancellables)
    }
    
    func clearSearch() {
        searchText = ""
        searchResults = []
    }
    
    // MARK: - Statistics
    func loadMessageStatistics(for conversationId: String) {
        chatRepository.getMessageStatistics(conversationId: conversationId)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        print("Failed to load statistics: \(error)")
                    }
                },
                receiveValue: { [weak self] statistics in
                    self?.messageStatistics = statistics
                }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - AI Personality
    func changeAIPersonality(_ personality: AIPersonality) {
        selectedAIPersonality = personality
        
        // 如果有当前对话，更新其AI人格
        if var conversation = currentConversation {
            conversation.aiPersonality = personality
            updateConversation(conversation)
        }
        
        showAIPersonalityPicker = false
    }
    
    private func updateConversation(_ conversation: Conversation) {
        chatRepository.updateConversation(conversation)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] updatedConversation in
                    self?.currentConversation = updatedConversation
                    
                    if let index = self?.conversations.firstIndex(where: { $0.id == updatedConversation.id }) {
                        self?.conversations[index] = updatedConversation
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - Refresh
    func refresh() {
        loadConversations()
        
        if let conversationId = currentConversation?.id {
            loadMessages(for: conversationId)
            loadMessageStatistics(for: conversationId)
        }
    }
    
    func syncData() {
        chatRepository.syncConversations()
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { [weak self] completion in
                    if case .failure(let error) = completion {
                        self?.handleError(error)
                    }
                },
                receiveValue: { [weak self] _ in
                    self?.loadConversations()
                }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - Helper Methods
    private func updateConversationInList() {
        guard let currentConversation = currentConversation else { return }
        
        // 将当前对话移到列表顶部
        conversations.removeAll { $0.id == currentConversation.id }
        conversations.insert(currentConversation, at: 0)
    }
    
    private func handleError(_ error: Error) {
        errorMessage = error.localizedDescription
        showError = true
    }
    
    func dismissError() {
        showError = false
        errorMessage = nil
    }
    
    // MARK: - Dialog Management
    func showNewConversationDialog() {
        newConversationTitle = ""
        showNewConversationDialog = true
    }
    
    func dismissNewConversationDialog() {
        showNewConversationDialog = false
        newConversationTitle = ""
    }
    
    func showConversationListDialog() {
        showConversationList = true
    }
    
    func dismissConversationListDialog() {
        showConversationList = false
    }
    
    func showAIPersonalityPickerDialog() {
        showAIPersonalityPicker = true
    }
    
    func dismissAIPersonalityPickerDialog() {
        showAIPersonalityPicker = false
    }
}

// MARK: - Extensions
extension ChatViewModel {
    // 获取AI人格的显示名称
    func getAIPersonalityDisplayName(_ personality: AIPersonality) -> String {
        switch personality {
        case .default:
            return "默认助手"
        case .wiseSage:
            return "智慧长者"
        case .friendlyGuide:
            return "友善向导"
        case .scholarly:
            return "学者专家"
        case .poetic:
            return "诗意文人"
        }
    }
    
    // 获取AI人格的描述
    func getAIPersonalityDescription(_ personality: AIPersonality) -> String {
        switch personality {
        case .default:
            return "通用AI助手，适合各种日常对话"
        case .wiseSage:
            return "融合古代智慧，以长者口吻提供人生指导"
        case .friendlyGuide:
            return "友善热情，善于解释和指导"
        case .scholarly:
            return "学术严谨，提供专业深入的分析"
        case .poetic:
            return "富有诗意，用优美的语言表达思想"
        }
    }
    
    // 获取消息状态的显示文本
    func getMessageStatusText(_ status: MessageStatus) -> String {
        switch status {
        case .sending:
            return "发送中..."
        case .sent:
            return "已发送"
        case .failed:
            return "发送失败"
        case .received:
            return "已接收"
        }
    }
    
    // 获取消息时间的显示格式
    func getMessageTimeText(_ date: Date) -> String {
        let formatter = DateFormatter()
        let calendar = Calendar.current
        
        if calendar.isToday(date) {
            formatter.dateFormat = "HH:mm"
        } else if calendar.isYesterday(date) {
            return "昨天 " + DateFormatter.localizedString(from: date, dateStyle: .none, timeStyle: .short)
        } else {
            formatter.dateFormat = "MM/dd HH:mm"
        }
        
        return formatter.string(from: date)
    }
}