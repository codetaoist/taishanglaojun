import Cocoa
import Foundation
import AVFoundation

// MARK: - Chat Models
struct ChatMessage: Codable, Identifiable {
    let id: UUID
    var content: String
    var type: MessageType
    var sender: MessageSender
    var timestamp: Date
    var attachments: [MessageAttachment]
    var isRead: Bool
    var reactions: [MessageReaction]
    
    init(content: String, type: MessageType, sender: MessageSender) {
        self.id = UUID()
        self.content = content
        self.type = type
        self.sender = sender
        self.timestamp = Date()
        self.attachments = []
        self.isRead = false
        self.reactions = []
    }
}

enum MessageType: String, Codable {
    case text = "文本"
    case image = "图片"
    case file = "文件"
    case voice = "语音"
    case video = "视频"
    case code = "代码"
    case system = "系统"
}

enum MessageSender: String, Codable {
    case user = "用户"
    case ai = "AI助手"
    case system = "系统"
}

struct MessageAttachment: Codable, Identifiable {
    let id: UUID
    var name: String
    var type: String
    var size: Int64
    var url: String
    var thumbnail: String?
}

struct MessageReaction: Codable, Identifiable {
    let id: UUID
    var emoji: String
    var count: Int
    var users: [String]
}

struct ChatSession: Codable, Identifiable {
    let id: UUID
    var title: String
    var messages: [ChatMessage]
    var createdAt: Date
    var updatedAt: Date
    var isActive: Bool
    
    init(title: String) {
        self.id = UUID()
        self.title = title
        self.messages = []
        self.createdAt = Date()
        self.updatedAt = Date()
        self.isActive = true
    }
}

// MARK: - Chat View Controller
class ChatViewController: NSViewController {
    
    // MARK: - Properties
    private var chatSessions: [ChatSession] = []
    private var currentSession: ChatSession?
    private var isTyping = false
    private var speechSynthesizer = AVSpeechSynthesizer()
    
    // MARK: - UI Components
    private lazy var splitView: NSSplitView = {
        let splitView = NSSplitView()
        splitView.isVertical = true
        splitView.dividerStyle = .thin
        return splitView
    }()
    
    private lazy var sessionSidebar: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var chatArea: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        return view
    }()
    
    private lazy var newChatButton: NSButton = {
        let button = NSButton()
        button.title = "新建对话"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(newChat(_:))
        return button
    }()
    
    private lazy var sessionTableView: NSTableView = {
        let tableView = NSTableView()
        tableView.delegate = self
        tableView.dataSource = self
        tableView.target = self
        tableView.action = #selector(selectSession(_:))
        
        let column = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("session"))
        column.title = "对话列表"
        tableView.addTableColumn(column)
        tableView.headerView = nil
        
        return tableView
    }()
    
    private lazy var sessionScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.documentView = sessionTableView
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        return scrollView
    }()
    
    private lazy var messagesScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        scrollView.documentView = messagesContainerView
        return scrollView
    }()
    
    private lazy var messagesContainerView: NSView = {
        let view = NSView()
        return view
    }()
    
    private lazy var messagesStackView: NSStackView = {
        let stackView = NSStackView()
        stackView.orientation = .vertical
        stackView.spacing = 10
        stackView.alignment = .leading
        return stackView
    }()
    
    private lazy var inputContainer: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var inputTextView: NSTextView = {
        let textView = NSTextView()
        textView.isRichText = false
        textView.font = NSFont.systemFont(ofSize: 14)
        textView.textColor = NSColor.labelColor
        textView.backgroundColor = NSColor.textBackgroundColor
        textView.isVerticallyResizable = true
        textView.isHorizontallyResizable = false
        textView.textContainer?.widthTracksTextView = true
        textView.textContainer?.containerSize = NSSize(width: 0, height: CGFloat.greatestFiniteMagnitude)
        textView.delegate = self
        return textView
    }()
    
    private lazy var inputScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.documentView = inputTextView
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        scrollView.borderType = .lineBorder
        return scrollView
    }()
    
    private lazy var sendButton: NSButton = {
        let button = NSButton()
        button.title = "发送"
        button.bezelStyle = .rounded
        button.keyEquivalent = "\r"
        button.target = self
        button.action = #selector(sendMessage(_:))
        return button
    }()
    
    private lazy var attachButton: NSButton = {
        let button = NSButton()
        button.title = "📎"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(attachFile(_:))
        return button
    }()
    
    private lazy var voiceButton: NSButton = {
        let button = NSButton()
        button.title = "🎤"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(toggleVoiceInput(_:))
        return button
    }()
    
    private lazy var typingIndicator: NSTextField = {
        let label = NSTextField(labelWithString: "AI正在思考...")
        label.font = NSFont.systemFont(ofSize: 12)
        label.textColor = NSColor.secondaryLabelColor
        label.isHidden = true
        return label
    }()
    
    // MARK: - Lifecycle
    override func loadView() {
        view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.windowBackgroundColor.cgColor
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        loadChatSessions()
        createDefaultSession()
    }
    
    // MARK: - UI Setup
    private func setupUI() {
        setupSplitView()
        setupSessionSidebar()
        setupChatArea()
        setupConstraints()
    }
    
    private func setupSplitView() {
        view.addSubview(splitView)
        splitView.addArrangedSubview(sessionSidebar)
        splitView.addArrangedSubview(chatArea)
        
        // 设置分割比例
        splitView.setPosition(250, ofDividerAt: 0)
    }
    
    private func setupSessionSidebar() {
        sessionSidebar.addSubview(newChatButton)
        sessionSidebar.addSubview(sessionScrollView)
    }
    
    private func setupChatArea() {
        chatArea.addSubview(messagesScrollView)
        chatArea.addSubview(inputContainer)
        chatArea.addSubview(typingIndicator)
        
        inputContainer.addSubview(inputScrollView)
        inputContainer.addSubview(sendButton)
        inputContainer.addSubview(attachButton)
        inputContainer.addSubview(voiceButton)
        
        messagesContainerView.addSubview(messagesStackView)
    }
    
    private func setupConstraints() {
        // 禁用自动调整大小
        [splitView, newChatButton, sessionScrollView, messagesScrollView, inputContainer,
         inputScrollView, sendButton, attachButton, voiceButton, typingIndicator,
         messagesContainerView, messagesStackView].forEach {
            $0.translatesAutoresizingMaskIntoConstraints = false
        }
        
        NSLayoutConstraint.activate([
            // Split View
            splitView.topAnchor.constraint(equalTo: view.topAnchor),
            splitView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            splitView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            splitView.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            
            // Session Sidebar
            newChatButton.topAnchor.constraint(equalTo: sessionSidebar.topAnchor, constant: 20),
            newChatButton.leadingAnchor.constraint(equalTo: sessionSidebar.leadingAnchor, constant: 20),
            newChatButton.trailingAnchor.constraint(equalTo: sessionSidebar.trailingAnchor, constant: -20),
            
            sessionScrollView.topAnchor.constraint(equalTo: newChatButton.bottomAnchor, constant: 20),
            sessionScrollView.leadingAnchor.constraint(equalTo: sessionSidebar.leadingAnchor, constant: 20),
            sessionScrollView.trailingAnchor.constraint(equalTo: sessionSidebar.trailingAnchor, constant: -20),
            sessionScrollView.bottomAnchor.constraint(equalTo: sessionSidebar.bottomAnchor, constant: -20),
            
            // Chat Area
            messagesScrollView.topAnchor.constraint(equalTo: chatArea.topAnchor, constant: 20),
            messagesScrollView.leadingAnchor.constraint(equalTo: chatArea.leadingAnchor, constant: 20),
            messagesScrollView.trailingAnchor.constraint(equalTo: chatArea.trailingAnchor, constant: -20),
            messagesScrollView.bottomAnchor.constraint(equalTo: inputContainer.topAnchor, constant: -20),
            
            // Messages Container
            messagesContainerView.topAnchor.constraint(equalTo: messagesScrollView.topAnchor),
            messagesContainerView.leadingAnchor.constraint(equalTo: messagesScrollView.leadingAnchor),
            messagesContainerView.trailingAnchor.constraint(equalTo: messagesScrollView.trailingAnchor),
            messagesContainerView.bottomAnchor.constraint(greaterThanOrEqualTo: messagesScrollView.bottomAnchor),
            messagesContainerView.widthAnchor.constraint(equalTo: messagesScrollView.widthAnchor),
            
            // Messages Stack View
            messagesStackView.topAnchor.constraint(equalTo: messagesContainerView.topAnchor, constant: 20),
            messagesStackView.leadingAnchor.constraint(equalTo: messagesContainerView.leadingAnchor, constant: 20),
            messagesStackView.trailingAnchor.constraint(equalTo: messagesContainerView.trailingAnchor, constant: -20),
            messagesStackView.bottomAnchor.constraint(equalTo: messagesContainerView.bottomAnchor, constant: -20),
            
            // Input Container
            inputContainer.leadingAnchor.constraint(equalTo: chatArea.leadingAnchor, constant: 20),
            inputContainer.trailingAnchor.constraint(equalTo: chatArea.trailingAnchor, constant: -20),
            inputContainer.bottomAnchor.constraint(equalTo: chatArea.bottomAnchor, constant: -20),
            inputContainer.heightAnchor.constraint(equalToConstant: 80),
            
            // Input Components
            attachButton.leadingAnchor.constraint(equalTo: inputContainer.leadingAnchor, constant: 10),
            attachButton.centerYAnchor.constraint(equalTo: inputContainer.centerYAnchor),
            attachButton.widthAnchor.constraint(equalToConstant: 40),
            
            voiceButton.leadingAnchor.constraint(equalTo: attachButton.trailingAnchor, constant: 5),
            voiceButton.centerYAnchor.constraint(equalTo: inputContainer.centerYAnchor),
            voiceButton.widthAnchor.constraint(equalToConstant: 40),
            
            inputScrollView.leadingAnchor.constraint(equalTo: voiceButton.trailingAnchor, constant: 10),
            inputScrollView.topAnchor.constraint(equalTo: inputContainer.topAnchor, constant: 10),
            inputScrollView.bottomAnchor.constraint(equalTo: inputContainer.bottomAnchor, constant: -10),
            inputScrollView.trailingAnchor.constraint(equalTo: sendButton.leadingAnchor, constant: -10),
            
            sendButton.trailingAnchor.constraint(equalTo: inputContainer.trailingAnchor, constant: -10),
            sendButton.centerYAnchor.constraint(equalTo: inputContainer.centerYAnchor),
            sendButton.widthAnchor.constraint(equalToConstant: 60),
            
            // Typing Indicator
            typingIndicator.leadingAnchor.constraint(equalTo: chatArea.leadingAnchor, constant: 40),
            typingIndicator.bottomAnchor.constraint(equalTo: inputContainer.topAnchor, constant: -5)
        ])
    }
    
    // MARK: - Data Management
    private func loadChatSessions() {
        // 模拟加载聊天会话
        chatSessions = []
        sessionTableView.reloadData()
    }
    
    private func createDefaultSession() {
        let session = ChatSession(title: "新对话")
        chatSessions.append(session)
        currentSession = session
        sessionTableView.reloadData()
        sessionTableView.selectRowIndexes(IndexSet(integer: 0), byExtendingSelection: false)
        updateMessagesView()
    }
    
    private func updateMessagesView() {
        // 清除现有消息视图
        messagesStackView.arrangedSubviews.forEach { $0.removeFromSuperview() }
        
        guard let session = currentSession else { return }
        
        // 添加消息视图
        for message in session.messages {
            let messageView = createMessageView(message: message)
            messagesStackView.addArrangedSubview(messageView)
        }
        
        // 滚动到底部
        DispatchQueue.main.async {
            self.scrollToBottom()
        }
    }
    
    private func createMessageView(message: ChatMessage) -> NSView {
        let containerView = NSView()
        containerView.wantsLayer = true
        
        let bubbleView = NSView()
        bubbleView.wantsLayer = true
        bubbleView.layer?.cornerRadius = 12
        
        let messageLabel = NSTextField(wrappingLabelWithString: message.content)
        messageLabel.font = NSFont.systemFont(ofSize: 14)
        messageLabel.maximumNumberOfLines = 0
        
        let timeLabel = NSTextField(labelWithString: formatTime(message.timestamp))
        timeLabel.font = NSFont.systemFont(ofSize: 11)
        timeLabel.textColor = NSColor.secondaryLabelColor
        
        // 根据发送者设置样式
        if message.sender == .user {
            bubbleView.layer?.backgroundColor = NSColor.controlAccentColor.cgColor
            messageLabel.textColor = NSColor.white
            
            containerView.addSubview(bubbleView)
            bubbleView.addSubview(messageLabel)
            containerView.addSubview(timeLabel)
            
            bubbleView.translatesAutoresizingMaskIntoConstraints = false
            messageLabel.translatesAutoresizingMaskIntoConstraints = false
            timeLabel.translatesAutoresizingMaskIntoConstraints = false
            
            NSLayoutConstraint.activate([
                bubbleView.trailingAnchor.constraint(equalTo: containerView.trailingAnchor),
                bubbleView.topAnchor.constraint(equalTo: containerView.topAnchor),
                bubbleView.widthAnchor.constraint(lessThanOrEqualToConstant: 300),
                
                messageLabel.topAnchor.constraint(equalTo: bubbleView.topAnchor, constant: 8),
                messageLabel.leadingAnchor.constraint(equalTo: bubbleView.leadingAnchor, constant: 12),
                messageLabel.trailingAnchor.constraint(equalTo: bubbleView.trailingAnchor, constant: -12),
                messageLabel.bottomAnchor.constraint(equalTo: bubbleView.bottomAnchor, constant: -8),
                
                timeLabel.topAnchor.constraint(equalTo: bubbleView.bottomAnchor, constant: 4),
                timeLabel.trailingAnchor.constraint(equalTo: bubbleView.trailingAnchor),
                timeLabel.bottomAnchor.constraint(equalTo: containerView.bottomAnchor)
            ])
        } else {
            bubbleView.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
            messageLabel.textColor = NSColor.labelColor
            
            containerView.addSubview(bubbleView)
            bubbleView.addSubview(messageLabel)
            containerView.addSubview(timeLabel)
            
            bubbleView.translatesAutoresizingMaskIntoConstraints = false
            messageLabel.translatesAutoresizingMaskIntoConstraints = false
            timeLabel.translatesAutoresizingMaskIntoConstraints = false
            
            NSLayoutConstraint.activate([
                bubbleView.leadingAnchor.constraint(equalTo: containerView.leadingAnchor),
                bubbleView.topAnchor.constraint(equalTo: containerView.topAnchor),
                bubbleView.widthAnchor.constraint(lessThanOrEqualToConstant: 300),
                
                messageLabel.topAnchor.constraint(equalTo: bubbleView.topAnchor, constant: 8),
                messageLabel.leadingAnchor.constraint(equalTo: bubbleView.leadingAnchor, constant: 12),
                messageLabel.trailingAnchor.constraint(equalTo: bubbleView.trailingAnchor, constant: -12),
                messageLabel.bottomAnchor.constraint(equalTo: bubbleView.bottomAnchor, constant: -8),
                
                timeLabel.topAnchor.constraint(equalTo: bubbleView.bottomAnchor, constant: 4),
                timeLabel.leadingAnchor.constraint(equalTo: bubbleView.leadingAnchor),
                timeLabel.bottomAnchor.constraint(equalTo: containerView.bottomAnchor)
            ])
        }
        
        return containerView
    }
    
    private func formatTime(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
    
    private func scrollToBottom() {
        let contentHeight = messagesStackView.fittingSize.height
        let scrollViewHeight = messagesScrollView.bounds.height
        
        if contentHeight > scrollViewHeight {
            let point = NSPoint(x: 0, y: contentHeight - scrollViewHeight)
            messagesScrollView.documentView?.scroll(point)
        }
    }
    
    // MARK: - Actions
    @objc private func newChat(_ sender: NSButton) {
        let session = ChatSession(title: "新对话 \(chatSessions.count + 1)")
        chatSessions.append(session)
        currentSession = session
        sessionTableView.reloadData()
        sessionTableView.selectRowIndexes(IndexSet(integer: chatSessions.count - 1), byExtendingSelection: false)
        updateMessagesView()
    }
    
    @objc private func selectSession(_ sender: NSTableView) {
        let selectedRow = sender.selectedRow
        guard selectedRow >= 0 && selectedRow < chatSessions.count else { return }
        
        currentSession = chatSessions[selectedRow]
        updateMessagesView()
    }
    
    @objc private func sendMessage(_ sender: NSButton) {
        guard let session = currentSession else { return }
        
        let messageText = inputTextView.string.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !messageText.isEmpty else { return }
        
        // 添加用户消息
        let userMessage = ChatMessage(content: messageText, type: .text, sender: .user)
        session.messages.append(userMessage)
        
        // 清空输入框
        inputTextView.string = ""
        
        // 更新会话标题
        if session.messages.count == 1 {
            let title = String(messageText.prefix(20))
            if let index = chatSessions.firstIndex(where: { $0.id == session.id }) {
                chatSessions[index].title = title
                sessionTableView.reloadData()
            }
        }
        
        // 更新消息视图
        updateMessagesView()
        
        // 模拟AI回复
        simulateAIResponse(for: session)
    }
    
    @objc private func attachFile(_ sender: NSButton) {
        let openPanel = NSOpenPanel()
        openPanel.allowsMultipleSelection = true
        openPanel.canChooseDirectories = false
        openPanel.canChooseFiles = true
        
        if openPanel.runModal() == .OK {
            for url in openPanel.urls {
                // 处理文件附件
                print("附加文件: \(url.lastPathComponent)")
            }
        }
    }
    
    @objc private func toggleVoiceInput(_ sender: NSButton) {
        // 实现语音输入功能
        print("切换语音输入")
    }
    
    private func simulateAIResponse(for session: ChatSession) {
        showTypingIndicator()
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
            self.hideTypingIndicator()
            
            let responses = [
                "我理解您的问题，让我为您详细解答。",
                "这是一个很好的问题！根据我的分析...",
                "我可以帮您解决这个问题。首先，我们需要...",
                "基于您提供的信息，我建议...",
                "让我为您提供一个全面的解决方案。"
            ]
            
            let randomResponse = responses.randomElement() ?? "感谢您的提问！"
            let aiMessage = ChatMessage(content: randomResponse, type: .text, sender: .ai)
            session.messages.append(aiMessage)
            
            self.updateMessagesView()
            
            // 语音播报（可选）
            self.speakMessage(randomResponse)
        }
    }
    
    private func showTypingIndicator() {
        typingIndicator.isHidden = false
        isTyping = true
    }
    
    private func hideTypingIndicator() {
        typingIndicator.isHidden = true
        isTyping = false
    }
    
    private func speakMessage(_ text: String) {
        let utterance = AVSpeechUtterance(string: text)
        utterance.voice = AVSpeechSynthesisVoice(language: "zh-CN")
        utterance.rate = 0.5
        speechSynthesizer.speak(utterance)
    }
}

// MARK: - Table View Data Source
extension ChatViewController: NSTableViewDataSource {
    func numberOfRows(in tableView: NSTableView) -> Int {
        return chatSessions.count
    }
}

// MARK: - Table View Delegate
extension ChatViewController: NSTableViewDelegate {
    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        guard row < chatSessions.count else { return nil }
        
        let session = chatSessions[row]
        
        let cellView = NSTableCellView()
        let textField = NSTextField()
        textField.isBordered = false
        textField.isEditable = false
        textField.backgroundColor = NSColor.clear
        textField.stringValue = session.title
        textField.font = NSFont.systemFont(ofSize: 13)
        
        cellView.addSubview(textField)
        cellView.textField = textField
        
        textField.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
            textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
            textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
        ])
        
        return cellView
    }
}

// MARK: - Text View Delegate
extension ChatViewController: NSTextViewDelegate {
    func textView(_ textView: NSTextView, doCommandBy commandSelector: Selector) -> Bool {
        if commandSelector == #selector(NSResponder.insertNewline(_:)) {
            // Shift+Enter 换行，Enter 发送
            if NSEvent.modifierFlags.contains(.shift) {
                return false // 允许换行
            } else {
                sendMessage(sendButton)
                return true // 阻止默认换行行为
            }
        }
        return false
    }
}