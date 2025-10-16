import SwiftUI
import WatchKit
import UserNotifications

/**
 * 通知视图
 * 显示和管理手表端的通知
 */
struct NotificationView: View {
    
    @EnvironmentObject private var taskManager: WatchTaskManager
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    @EnvironmentObject private var connectivityManager: WatchConnectivityManager
    
    @State private var notifications: [WatchNotification] = []
    @State private var selectedTab: NotificationTab = .all
    @State private var isLoading = false
    @State private var showingClearAlert = false
    @State private var selectedNotification: WatchNotification?
    
    var body: some View {
        NavigationView {
            VStack(spacing: 0) {
                // 通知标签选择器
                NotificationTabSelector(
                    selectedTab: $selectedTab,
                    notificationCounts: getNotificationCounts()
                )
                
                // 通知列表
                if filteredNotifications.isEmpty {
                    EmptyNotificationsView(selectedTab: selectedTab)
                } else {
                    ScrollView {
                        LazyVStack(spacing: 8) {
                            ForEach(filteredNotifications) { notification in
                                NotificationRowView(
                                    notification: notification,
                                    onTap: { handleNotificationTap(notification) },
                                    onDismiss: { dismissNotification(notification) }
                                )
                            }
                        }
                        .padding(.horizontal, 8)
                    }
                }
                
                // 底部操作栏
                if !filteredNotifications.isEmpty {
                    NotificationActionBar(
                        onMarkAllRead: markAllAsRead,
                        onClearAll: { showingClearAlert = true }
                    )
                }
            }
            .navigationTitle("通知")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("设置") {
                        // 导航到通知设置
                    }
                    .font(.caption)
                }
            }
        }
        .onAppear {
            loadNotifications()
        }
        .alert("清空通知", isPresented: $showingClearAlert) {
            Button("取消", role: .cancel) { }
            Button("清空", role: .destructive) {
                clearAllNotifications()
            }
        } message: {
            Text("确定要清空所有通知吗？此操作无法撤销。")
        }
        .sheet(item: $selectedNotification) { notification in
            NotificationDetailView(notification: notification)
        }
    }
    
    // MARK: - Computed Properties
    
    private var filteredNotifications: [WatchNotification] {
        switch selectedTab {
        case .all:
            return notifications
        case .unread:
            return notifications.filter { !$0.isRead }
        case .tasks:
            return notifications.filter { $0.type == .taskUpdate || $0.type == .taskAssigned }
        case .system:
            return notifications.filter { $0.type == .systemAlert || $0.type == .syncComplete }
        }
    }
    
    // MARK: - Private Methods
    
    private func loadNotifications() {
        isLoading = true
        
        // 模拟加载通知数据
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
            notifications = generateSampleNotifications()
            isLoading = false
        }
    }
    
    private func getNotificationCounts() -> NotificationCounts {
        return NotificationCounts(
            all: notifications.count,
            unread: notifications.filter { !$0.isRead }.count,
            tasks: notifications.filter { $0.type == .taskUpdate || $0.type == .taskAssigned }.count,
            system: notifications.filter { $0.type == .systemAlert || $0.type == .syncComplete }.count
        )
    }
    
    private func handleNotificationTap(_ notification: WatchNotification) {
        // 标记为已读
        markAsRead(notification)
        
        // 根据通知类型执行相应操作
        switch notification.type {
        case .taskAssigned, .taskUpdate:
            if let taskId = notification.relatedTaskId {
                // 导航到任务详情
                selectedNotification = notification
            }
        case .syncComplete:
            // 刷新任务数据
            Task {
                await taskManager.refreshTasks()
            }
        case .systemAlert:
            selectedNotification = notification
        }
        
        // 触觉反馈
        settingsManager.triggerHapticFeedback(.click)
    }
    
    private func dismissNotification(_ notification: WatchNotification) {
        withAnimation(.easeInOut(duration: 0.3)) {
            notifications.removeAll { $0.id == notification.id }
        }
        
        settingsManager.triggerHapticFeedback(.success)
    }
    
    private func markAsRead(_ notification: WatchNotification) {
        if let index = notifications.firstIndex(where: { $0.id == notification.id }) {
            notifications[index].isRead = true
        }
    }
    
    private func markAllAsRead() {
        for index in notifications.indices {
            notifications[index].isRead = true
        }
        
        settingsManager.triggerHapticFeedback(.success)
    }
    
    private func clearAllNotifications() {
        withAnimation(.easeInOut(duration: 0.3)) {
            notifications.removeAll()
        }
        
        settingsManager.triggerHapticFeedback(.success)
    }
    
    private func generateSampleNotifications() -> [WatchNotification] {
        return [
            WatchNotification(
                id: "1",
                title: "新任务分配",
                message: "您有一个新的高优先级任务需要处理",
                type: .taskAssigned,
                timestamp: Date().addingTimeInterval(-300),
                isRead: false,
                relatedTaskId: "task_1"
            ),
            WatchNotification(
                id: "2",
                title: "任务进度更新",
                message: "任务"数据分析"已完成50%",
                type: .taskUpdate,
                timestamp: Date().addingTimeInterval(-600),
                isRead: false,
                relatedTaskId: "task_2"
            ),
            WatchNotification(
                id: "3",
                title: "同步完成",
                message: "数据同步已完成，共更新3个任务",
                type: .syncComplete,
                timestamp: Date().addingTimeInterval(-900),
                isRead: true
            ),
            WatchNotification(
                id: "4",
                title: "系统提醒",
                message: "手表电量较低，建议及时充电",
                type: .systemAlert,
                timestamp: Date().addingTimeInterval(-1200),
                isRead: false
            )
        ]
    }
}

/**
 * 通知标签选择器
 */
struct NotificationTabSelector: View {
    @Binding var selectedTab: NotificationTab
    let notificationCounts: NotificationCounts
    
    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 12) {
                ForEach(NotificationTab.allCases, id: \.self) { tab in
                    NotificationTabButton(
                        tab: tab,
                        count: getCount(for: tab),
                        isSelected: selectedTab == tab,
                        onTap: { selectedTab = tab }
                    )
                }
            }
            .padding(.horizontal, 16)
        }
        .padding(.vertical, 8)
    }
    
    private func getCount(for tab: NotificationTab) -> Int {
        switch tab {
        case .all: return notificationCounts.all
        case .unread: return notificationCounts.unread
        case .tasks: return notificationCounts.tasks
        case .system: return notificationCounts.system
        }
    }
}

/**
 * 通知标签按钮
 */
struct NotificationTabButton: View {
    let tab: NotificationTab
    let count: Int
    let isSelected: Bool
    let onTap: () -> Void
    
    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 4) {
                Text(tab.displayName)
                    .font(.caption2)
                    .fontWeight(isSelected ? .semibold : .regular)
                
                if count > 0 {
                    Text("\(count)")
                        .font(.caption2)
                        .fontWeight(.medium)
                        .foregroundColor(.white)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(
                            Capsule()
                                .fill(isSelected ? Color.white : Color.red)
                        )
                }
            }
            .foregroundColor(isSelected ? .white : .primary)
            .padding(.horizontal, 12)
            .padding(.vertical, 6)
            .background(
                Capsule()
                    .fill(isSelected ? Color.blue : Color.clear)
                    .stroke(Color.blue, lineWidth: isSelected ? 0 : 1)
            )
        }
        .buttonStyle(PlainButtonStyle())
    }
}

/**
 * 通知行视图
 */
struct NotificationRowView: View {
    let notification: WatchNotification
    let onTap: () -> Void
    let onDismiss: () -> Void
    
    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 12) {
                // 通知类型图标
                NotificationTypeIcon(type: notification.type)
                
                // 通知内容
                VStack(alignment: .leading, spacing: 4) {
                    HStack {
                        Text(notification.title)
                            .font(.caption)
                            .fontWeight(.medium)
                            .foregroundColor(.primary)
                            .lineLimit(1)
                        
                        Spacer()
                        
                        if !notification.isRead {
                            Circle()
                                .fill(Color.blue)
                                .frame(width: 6, height: 6)
                        }
                    }
                    
                    Text(notification.message)
                        .font(.caption2)
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    
                    Text(formatTimestamp(notification.timestamp))
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
                
                Spacer()
            }
            .padding(12)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(notification.isRead ? Color.clear : Color.blue.opacity(0.1))
                    .stroke(Color.gray.opacity(0.3), lineWidth: 0.5)
            )
        }
        .buttonStyle(PlainButtonStyle())
        .contextMenu {
            Button(action: onDismiss) {
                Label("删除", systemImage: "trash")
            }
            
            if !notification.isRead {
                Button("标记为已读") {
                    // 标记为已读的逻辑
                }
            }
        }
    }
    
    private func formatTimestamp(_ date: Date) -> String {
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: date, relativeTo: Date())
    }
}

/**
 * 通知类型图标
 */
struct NotificationTypeIcon: View {
    let type: NotificationType
    
    var body: some View {
        Image(systemName: type.iconName)
            .font(.caption)
            .foregroundColor(type.color)
            .frame(width: 20, height: 20)
            .background(
                Circle()
                    .fill(type.color.opacity(0.2))
            )
    }
}

/**
 * 空通知视图
 */
struct EmptyNotificationsView: View {
    let selectedTab: NotificationTab
    
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "bell.slash")
                .font(.title2)
                .foregroundColor(.secondary)
            
            Text(getEmptyMessage())
                .font(.caption)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
        }
        .padding(32)
    }
    
    private func getEmptyMessage() -> String {
        switch selectedTab {
        case .all:
            return "暂无通知"
        case .unread:
            return "没有未读通知"
        case .tasks:
            return "没有任务相关通知"
        case .system:
            return "没有系统通知"
        }
    }
}

/**
 * 通知操作栏
 */
struct NotificationActionBar: View {
    let onMarkAllRead: () -> Void
    let onClearAll: () -> Void
    
    var body: some View {
        HStack(spacing: 16) {
            Button("全部已读", action: onMarkAllRead)
                .font(.caption2)
                .foregroundColor(.blue)
            
            Spacer()
            
            Button("清空", action: onClearAll)
                .font(.caption2)
                .foregroundColor(.red)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
        .background(Color.gray.opacity(0.1))
    }
}

/**
 * 通知详情视图
 */
struct NotificationDetailView: View {
    let notification: WatchNotification
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(alignment: .leading, spacing: 16) {
                    // 通知头部
                    HStack {
                        NotificationTypeIcon(type: notification.type)
                        
                        VStack(alignment: .leading, spacing: 4) {
                            Text(notification.title)
                                .font(.caption)
                                .fontWeight(.medium)
                            
                            Text(formatFullTimestamp(notification.timestamp))
                                .font(.caption2)
                                .foregroundColor(.secondary)
                        }
                        
                        Spacer()
                    }
                    
                    Divider()
                    
                    // 通知内容
                    Text(notification.message)
                        .font(.caption)
                        .foregroundColor(.primary)
                    
                    // 相关操作
                    if notification.type == .taskAssigned || notification.type == .taskUpdate {
                        Button("查看任务") {
                            // 导航到任务详情
                            dismiss()
                        }
                        .font(.caption2)
                        .foregroundColor(.blue)
                    }
                }
                .padding(16)
            }
            .navigationTitle("通知详情")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("关闭") {
                        dismiss()
                    }
                    .font(.caption2)
                }
            }
        }
    }
    
    private func formatFullTimestamp(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

// MARK: - Supporting Types

/**
 * 通知标签枚举
 */
enum NotificationTab: CaseIterable {
    case all, unread, tasks, system
    
    var displayName: String {
        switch self {
        case .all: return "全部"
        case .unread: return "未读"
        case .tasks: return "任务"
        case .system: return "系统"
        }
    }
}

/**
 * 通知计数结构
 */
struct NotificationCounts {
    let all: Int
    let unread: Int
    let tasks: Int
    let system: Int
}

/**
 * 手表通知数据模型
 */
struct WatchNotification: Identifiable, Codable {
    let id: String
    let title: String
    let message: String
    let type: NotificationType
    let timestamp: Date
    var isRead: Bool
    let relatedTaskId: String?
    
    init(id: String, title: String, message: String, type: NotificationType, timestamp: Date, isRead: Bool = false, relatedTaskId: String? = nil) {
        self.id = id
        self.title = title
        self.message = message
        self.type = type
        self.timestamp = timestamp
        self.isRead = isRead
        self.relatedTaskId = relatedTaskId
    }
}

/**
 * 通知类型扩展
 */
extension NotificationType {
    var iconName: String {
        switch self {
        case .taskAssigned: return "plus.circle"
        case .taskUpdate: return "arrow.clockwise.circle"
        case .syncComplete: return "checkmark.circle"
        case .systemAlert: return "exclamationmark.triangle"
        }
    }
    
    var color: Color {
        switch self {
        case .taskAssigned: return .green
        case .taskUpdate: return .blue
        case .syncComplete: return .purple
        case .systemAlert: return .orange
        }
    }
}

#Preview {
    NotificationView()
        .environmentObject(WatchTaskManager())
        .environmentObject(WatchSettingsManager())
        .environmentObject(WatchConnectivityManager())
}