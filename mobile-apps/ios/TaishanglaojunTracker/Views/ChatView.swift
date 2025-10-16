import SwiftUI
import Combine

// MARK: - Chat View
struct ChatView: View {
    @StateObject private var viewModel = ChatViewModel()
    @State private var scrollProxy: ScrollViewReader?
    @FocusState private var isTextFieldFocused: Bool
    
    var body: some View {
        NavigationView {
            VStack(spacing: 0) {
                // Top Bar
                topBar
                
                // Messages Area
                messagesArea
                
                // Input Area
                inputArea
            }
            .navigationBarHidden(true)
            .background(Color(.systemGroupedBackground))
            .onAppear {
                viewModel.refresh()
            }
        }
        .sheet(isPresented: $viewModel.showConversationList) {
            ConversationListView(viewModel: viewModel)
        }
        .sheet(isPresented: $viewModel.showAIPersonalityPicker) {
            AIPersonalityPickerView(viewModel: viewModel)
        }
        .alert("新建对话", isPresented: $viewModel.showNewConversationDialog) {
            TextField("对话标题", text: $viewModel.newConversationTitle)
            Button("创建") {
                viewModel.createConversation()
            }
            Button("取消", role: .cancel) {
                viewModel.dismissNewConversationDialog()
            }
        }
        .alert("错误", isPresented: $viewModel.showError) {
            Button("确定") {
                viewModel.dismissError()
            }
        } message: {
            Text(viewModel.errorMessage ?? "未知错误")
        }
    }
    
    // MARK: - Top Bar
    private var topBar: some View {
        HStack {
            // Conversation List Button
            Button(action: {
                viewModel.showConversationListDialog()
            }) {
                Image(systemName: "list.bullet")
                    .font(.title2)
                    .foregroundColor(.primary)
            }
            
            Spacer()
            
            // Current Conversation Title
            VStack(spacing: 2) {
                Text(viewModel.currentConversation?.title ?? "AI对话")
                    .font(.headline)
                    .foregroundColor(.primary)
                
                Text(viewModel.getAIPersonalityDisplayName(viewModel.selectedAIPersonality))
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            Spacer()
            
            // Menu Button
            Menu {
                Button(action: {
                    viewModel.showAIPersonalityPickerDialog()
                }) {
                    Label("AI人格", systemImage: "person.crop.circle")
                }
                
                Button(action: {
                    viewModel.showNewConversationDialog()
                }) {
                    Label("新建对话", systemImage: "plus.message")
                }
                
                Button(action: {
                    viewModel.refresh()
                }) {
                    Label("刷新", systemImage: "arrow.clockwise")
                }
                
                if let conversation = viewModel.currentConversation {
                    Divider()
                    
                    Button(role: .destructive, action: {
                        viewModel.archiveConversation(conversation)
                    }) {
                        Label("归档对话", systemImage: "archivebox")
                    }
                }
            } label: {
                Image(systemName: "ellipsis.circle")
                    .font(.title2)
                    .foregroundColor(.primary)
            }
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
        .shadow(color: .black.opacity(0.1), radius: 1, x: 0, y: 1)
    }
    
    // MARK: - Messages Area
    private var messagesArea: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(spacing: 12) {
                    if viewModel.isLoading && viewModel.messages.isEmpty {
                        ProgressView("加载中...")
                            .frame(maxWidth: .infinity, maxHeight: .infinity)
                    } else if viewModel.messages.isEmpty {
                        emptyStateView
                    } else {
                        ForEach(viewModel.messages, id: \.id) { message in
                            MessageBubbleView(
                                message: message,
                                onRetry: {
                                    viewModel.retryMessage(message)
                                },
                                onDelete: {
                                    viewModel.deleteMessage(message)
                                }
                            )
                            .id(message.id)
                        }
                    }
                }
                .padding(.horizontal)
                .padding(.vertical, 8)
            }
            .onAppear {
                scrollProxy = proxy
            }
            .onChange(of: viewModel.messages.count) { _ in
                scrollToBottom()
            }
        }
    }
    
    // MARK: - Empty State
    private var emptyStateView: some View {
        VStack(spacing: 16) {
            Image(systemName: "message.circle")
                .font(.system(size: 60))
                .foregroundColor(.secondary)
            
            Text("开始新的对话")
                .font(.title2)
                .fontWeight(.medium)
                .foregroundColor(.primary)
            
            Text("选择一个AI人格，开始您的智慧之旅")
                .font(.body)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
            
            Button(action: {
                viewModel.showAIPersonalityPickerDialog()
            }) {
                Text("选择AI人格")
                    .font(.headline)
                    .foregroundColor(.white)
                    .padding(.horizontal, 24)
                    .padding(.vertical, 12)
                    .background(Color.accentColor)
                    .cornerRadius(25)
            }
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
    
    // MARK: - Input Area
    private var inputArea: some View {
        VStack(spacing: 8) {
            // Character Count
            if !viewModel.messageText.isEmpty {
                HStack {
                    Spacer()
                    Text(viewModel.messageCountText)
                        .font(.caption)
                        .foregroundColor(viewModel.isMessageTooLong ? .red : .secondary)
                }
                .padding(.horizontal)
            }
            
            // Input Field
            HStack(spacing: 12) {
                // Text Input
                TextField("输入消息...", text: $viewModel.messageText, axis: .vertical)
                    .textFieldStyle(.roundedBorder)
                    .focused($isTextFieldFocused)
                    .lineLimit(1...5)
                    .disabled(viewModel.isSending)
                
                // Send Button
                Button(action: {
                    viewModel.sendMessage()
                    isTextFieldFocused = false
                }) {
                    if viewModel.isSending {
                        ProgressView()
                            .scaleEffect(0.8)
                    } else {
                        Image(systemName: "arrow.up.circle.fill")
                            .font(.title2)
                    }
                }
                .disabled(!viewModel.canSendMessage)
                .foregroundColor(viewModel.canSendMessage ? .accentColor : .secondary)
            }
            .padding(.horizontal)
            .padding(.bottom, 8)
        }
        .background(Color(.systemBackground))
        .shadow(color: .black.opacity(0.1), radius: 1, x: 0, y: -1)
    }
    
    // MARK: - Helper Methods
    private func scrollToBottom() {
        guard let lastMessage = viewModel.messages.last else { return }
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            withAnimation(.easeInOut(duration: 0.3)) {
                scrollProxy?.scrollTo(lastMessage.id, anchor: .bottom)
            }
        }
    }
}

// MARK: - Message Bubble View
struct MessageBubbleView: View {
    let message: ChatMessage
    let onRetry: () -> Void
    let onDelete: () -> Void
    
    @State private var showActions = false
    
    var body: some View {
        HStack {
            if message.sender == .user {
                Spacer(minLength: 50)
            }
            
            VStack(alignment: message.sender == .user ? .trailing : .leading, spacing: 4) {
                // Message Content
                Text(message.content)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .background(messageBackgroundColor)
                    .foregroundColor(messageTextColor)
                    .cornerRadius(18)
                    .contextMenu {
                        Button(action: {
                            UIPasteboard.general.string = message.content
                        }) {
                            Label("复制", systemImage: "doc.on.doc")
                        }
                        
                        if message.sender == .user && message.status == .failed {
                            Button(action: onRetry) {
                                Label("重试", systemImage: "arrow.clockwise")
                            }
                        }
                        
                        Button(role: .destructive, action: onDelete) {
                            Label("删除", systemImage: "trash")
                        }
                    }
                
                // Message Info
                HStack(spacing: 4) {
                    Text(formatMessageTime(message.timestamp))
                        .font(.caption2)
                        .foregroundColor(.secondary)
                    
                    if message.sender == .user {
                        messageStatusIcon
                    }
                }
            }
            
            if message.sender == .ai {
                Spacer(minLength: 50)
            }
        }
    }
    
    private var messageBackgroundColor: Color {
        switch message.sender {
        case .user:
            return Color.accentColor
        case .ai:
            return Color(.systemGray5)
        }
    }
    
    private var messageTextColor: Color {
        switch message.sender {
        case .user:
            return .white
        case .ai:
            return .primary
        }
    }
    
    private var messageStatusIcon: some View {
        Group {
            switch message.status {
            case .sending:
                ProgressView()
                    .scaleEffect(0.6)
            case .sent:
                Image(systemName: "checkmark")
                    .font(.caption2)
                    .foregroundColor(.secondary)
            case .failed:
                Button(action: onRetry) {
                    Image(systemName: "exclamationmark.circle")
                        .font(.caption2)
                        .foregroundColor(.red)
                }
            case .received:
                Image(systemName: "checkmark.circle")
                    .font(.caption2)
                    .foregroundColor(.green)
            }
        }
    }
    
    private func formatMessageTime(_ date: Date) -> String {
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

// MARK: - Conversation List View
struct ConversationListView: View {
    @ObservedObject var viewModel: ChatViewModel
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        NavigationView {
            List {
                ForEach(viewModel.conversations, id: \.id) { conversation in
                    ConversationRowView(
                        conversation: conversation,
                        isSelected: conversation.id == viewModel.currentConversation?.id,
                        onSelect: {
                            viewModel.selectConversation(conversation)
                            dismiss()
                        },
                        onDelete: {
                            viewModel.deleteConversation(conversation)
                        },
                        onArchive: {
                            viewModel.archiveConversation(conversation)
                        }
                    )
                }
            }
            .navigationTitle("对话列表")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("关闭") {
                        dismiss()
                    }
                }
                
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: {
                        viewModel.showNewConversationDialog()
                        dismiss()
                    }) {
                        Image(systemName: "plus")
                    }
                }
            }
        }
    }
}

// MARK: - Conversation Row View
struct ConversationRowView: View {
    let conversation: Conversation
    let isSelected: Bool
    let onSelect: () -> Void
    let onDelete: () -> Void
    let onArchive: () -> Void
    
    var body: some View {
        Button(action: onSelect) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(conversation.title)
                        .font(.headline)
                        .foregroundColor(.primary)
                        .lineLimit(1)
                    
                    Text(getAIPersonalityDisplayName(conversation.aiPersonality))
                        .font(.caption)
                        .foregroundColor(.secondary)
                    
                    Text(formatDate(conversation.updatedAt))
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
                
                Spacer()
                
                if isSelected {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundColor(.accentColor)
                }
            }
            .padding(.vertical, 4)
        }
        .buttonStyle(.plain)
        .contextMenu {
            Button(action: onArchive) {
                Label("归档", systemImage: "archivebox")
            }
            
            Button(role: .destructive, action: onDelete) {
                Label("删除", systemImage: "trash")
            }
        }
    }
    
    private func getAIPersonalityDisplayName(_ personality: AIPersonality) -> String {
        switch personality {
        case .default: return "默认助手"
        case .wiseSage: return "智慧长者"
        case .friendlyGuide: return "友善向导"
        case .scholarly: return "学者专家"
        case .poetic: return "诗意文人"
        }
    }
    
    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        let calendar = Calendar.current
        
        if calendar.isToday(date) {
            formatter.dateFormat = "HH:mm"
            return "今天 " + formatter.string(from: date)
        } else if calendar.isYesterday(date) {
            formatter.dateFormat = "HH:mm"
            return "昨天 " + formatter.string(from: date)
        } else {
            formatter.dateFormat = "MM/dd"
            return formatter.string(from: date)
        }
    }
}

// MARK: - AI Personality Picker View
struct AIPersonalityPickerView: View {
    @ObservedObject var viewModel: ChatViewModel
    @Environment(\.dismiss) private var dismiss
    
    private let personalities: [AIPersonality] = [.default, .wiseSage, .friendlyGuide, .scholarly, .poetic]
    
    var body: some View {
        NavigationView {
            List {
                ForEach(personalities, id: \.self) { personality in
                    PersonalityRowView(
                        personality: personality,
                        isSelected: personality == viewModel.selectedAIPersonality,
                        onSelect: {
                            viewModel.changeAIPersonality(personality)
                            dismiss()
                        }
                    )
                }
            }
            .navigationTitle("选择AI人格")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("关闭") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MARK: - Personality Row View
struct PersonalityRowView: View {
    let personality: AIPersonality
    let isSelected: Bool
    let onSelect: () -> Void
    
    var body: some View {
        Button(action: onSelect) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(getDisplayName(personality))
                        .font(.headline)
                        .foregroundColor(.primary)
                    
                    Text(getDescription(personality))
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }
                
                Spacer()
                
                if isSelected {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundColor(.accentColor)
                }
            }
            .padding(.vertical, 4)
        }
        .buttonStyle(.plain)
    }
    
    private func getDisplayName(_ personality: AIPersonality) -> String {
        switch personality {
        case .default: return "默认助手"
        case .wiseSage: return "智慧长者"
        case .friendlyGuide: return "友善向导"
        case .scholarly: return "学者专家"
        case .poetic: return "诗意文人"
        }
    }
    
    private func getDescription(_ personality: AIPersonality) -> String {
        switch personality {
        case .default: return "通用AI助手，适合各种日常对话"
        case .wiseSage: return "融合古代智慧，以长者口吻提供人生指导"
        case .friendlyGuide: return "友善热情，善于解释和指导"
        case .scholarly: return "学术严谨，提供专业深入的分析"
        case .poetic: return "富有诗意，用优美的语言表达思想"
        }
    }
}

// MARK: - Preview
struct ChatView_Previews: PreviewProvider {
    static var previews: some View {
        ChatView()
    }
}