import Foundation
import UserNotifications
import Cocoa

class NotificationManager: NSObject {
    
    // MARK: - Singleton
    static let shared = NotificationManager()
    
    // MARK: - Properties
    private let notificationCenter = UNUserNotificationCenter.current()
    private var isAuthorized = false
    
    // 通知类型
    enum NotificationType: String, CaseIterable {
        case petAction = "pet_action"
        case fileTransfer = "file_transfer"
        case dataSync = "data_sync"
        case aiResponse = "ai_response"
        case systemAlert = "system_alert"
        case reminder = "reminder"
        
        var categoryIdentifier: String {
            return "category_" + self.rawValue
        }
    }
    
    // 通知优先级
    enum NotificationPriority {
        case low
        case normal
        case high
        case critical
        
        var interruptionLevel: UNNotificationInterruptionLevel {
            switch self {
            case .low:
                return .passive
            case .normal:
                return .active
            case .high:
                return .timeSensitive
            case .critical:
                return .critical
            }
        }
    }
    
    // MARK: - Initialization
    
    private override init() {
        super.init()
        setupNotificationCenter()
    }
    
    // MARK: - Setup
    
    private func setupNotificationCenter() {
        notificationCenter.delegate = self
        requestAuthorization()
        setupNotificationCategories()
    }
    
    private func requestAuthorization() {
        let options: UNAuthorizationOptions = [.alert, .sound, .badge, .provisional]
        
        notificationCenter.requestAuthorization(options: options) { [weak self] granted, error in
            DispatchQueue.main.async {
                self?.isAuthorized = granted
                
                if let error = error {
                    NSLog("通知授权失败: \\(error.localizedDescription)")
                } else if granted {
                    NSLog("通知授权成功")
                    self?.registerForRemoteNotifications()
                } else {
                    NSLog("用户拒绝了通知授权")
                }
            }
        }
    }
    
    private func registerForRemoteNotifications() {
        DispatchQueue.main.async {
            NSApplication.shared.registerForRemoteNotifications()
        }
    }
    
    private func setupNotificationCategories() {
        var categories: Set<UNNotificationCategory> = []
        
        // 桌面宠物通知类别
        let petActions = [
            UNNotificationAction(identifier: "pet_interact", title: "与宠物互动", options: []),
            UNNotificationAction(identifier: "pet_feed", title: "喂食", options: []),
            UNNotificationAction(identifier: "pet_dismiss", title: "忽略", options: [.destructive])
        ]
        let petCategory = UNNotificationCategory(
            identifier: NotificationType.petAction.categoryIdentifier,
            actions: petActions,
            intentIdentifiers: [],
            options: []
        )
        categories.insert(petCategory)
        
        // 文件传输通知类别
        let transferActions = [
            UNNotificationAction(identifier: "transfer_open", title: "打开文件", options: [.foreground]),
            UNNotificationAction(identifier: "transfer_folder", title: "打开文件夹", options: [.foreground]),
            UNNotificationAction(identifier: "transfer_dismiss", title: "关闭", options: [])
        ]
        let transferCategory = UNNotificationCategory(
            identifier: NotificationType.fileTransfer.categoryIdentifier,
            actions: transferActions,
            intentIdentifiers: [],
            options: []
        )
        categories.insert(transferCategory)
        
        // 数据同步通知类别
        let syncActions = [
            UNNotificationAction(identifier: "sync_view", title: "查看详情", options: [.foreground]),
            UNNotificationAction(identifier: "sync_retry", title: "重试", options: []),
            UNNotificationAction(identifier: "sync_dismiss", title: "关闭", options: [])
        ]
        let syncCategory = UNNotificationCategory(
            identifier: NotificationType.dataSync.categoryIdentifier,
            actions: syncActions,
            intentIdentifiers: [],
            options: []
        )
        categories.insert(syncCategory)
        
        // AI响应通知类别
        let aiActions = [
            UNNotificationAction(identifier: "ai_reply", title: "回复", options: [.foreground]),
            UNNotificationAction(identifier: "ai_copy", title: "复制", options: []),
            UNNotificationAction(identifier: "ai_dismiss", title: "关闭", options: [])
        ]
        let aiCategory = UNNotificationCategory(
            identifier: NotificationType.aiResponse.categoryIdentifier,
            actions: aiActions,
            intentIdentifiers: [],
            options: []
        )
        categories.insert(aiCategory)
        
        notificationCenter.setNotificationCategories(categories)
    }
    
    // MARK: - Public Methods
    
    func sendNotification(
        title: String,
        body: String,
        type: NotificationType = .systemAlert,
        priority: NotificationPriority = .normal,
        delay: TimeInterval = 0,
        userInfo: [String: Any] = [:],
        completion: ((Error?) -> Void)? = nil
    ) {
        guard isAuthorized else {
            NSLog("通知未授权，无法发送通知")
            completion?(NSError(domain: "NotificationManager", code: 1, userInfo: [NSLocalizedDescriptionKey: "通知未授权"]))
            return
        }
        
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default
        content.categoryIdentifier = type.categoryIdentifier
        content.userInfo = userInfo
        
        // 设置中断级别
        if #available(macOS 12.0, *) {
            content.interruptionLevel = priority.interruptionLevel
        }
        
        // 设置图标
        if let iconURL = Bundle.main.url(forResource: "notification_icon", withExtension: "png") {
            do {
                let attachment = try UNNotificationAttachment(identifier: "icon", url: iconURL, options: nil)
                content.attachments = [attachment]
            } catch {
                NSLog("添加通知图标失败: \\(error)")
            }
        }
        
        // 创建触发器
        let trigger: UNNotificationTrigger?
        if delay > 0 {
            trigger = UNTimeIntervalNotificationTrigger(timeInterval: delay, repeats: false)
        } else {
            trigger = nil
        }
        
        // 创建请求
        let identifier = UUID().uuidString
        let request = UNNotificationRequest(identifier: identifier, content: content, trigger: trigger)
        
        // 发送通知
        notificationCenter.add(request) { error in
            DispatchQueue.main.async {
                if let error = error {
                    NSLog("发送通知失败: \\(error.localizedDescription)")
                } else {
                    NSLog("通知发送成功: \\(title)")
                }
                completion?(error)
            }
        }
    }
    
    func sendPetNotification(petName: String, action: String, completion: ((Error?) -> Void)? = nil) {
        sendNotification(
            title: "桌面宠物 - \\(petName)",
            body: action,
            type: .petAction,
            priority: .normal,
            userInfo: ["pet_name": petName, "action": action],
            completion: completion
        )
    }
    
    func sendFileTransferNotification(fileName: String, isComplete: Bool, completion: ((Error?) -> Void)? = nil) {
        let title = isComplete ? "文件传输完成" : "文件传输进行中"
        let body = isComplete ? "文件 \\(fileName) 传输完成" : "正在传输文件 \\(fileName)"
        
        sendNotification(
            title: title,
            body: body,
            type: .fileTransfer,
            priority: isComplete ? .normal : .low,
            userInfo: ["file_name": fileName, "is_complete": isComplete],
            completion: completion
        )
    }
    
    func sendDataSyncNotification(isSuccess: Bool, details: String = "", completion: ((Error?) -> Void)? = nil) {
        let title = isSuccess ? "数据同步成功" : "数据同步失败"
        let body = details.isEmpty ? title : details
        
        sendNotification(
            title: title,
            body: body,
            type: .dataSync,
            priority: isSuccess ? .normal : .high,
            userInfo: ["is_success": isSuccess, "details": details],
            completion: completion
        )
    }
    
    func sendAIResponseNotification(response: String, completion: ((Error?) -> Void)? = nil) {
        sendNotification(
            title: "AI助手回复",
            body: response,
            type: .aiResponse,
            priority: .normal,
            userInfo: ["ai_response": response],
            completion: completion
        )
    }
    
    func sendSystemAlert(message: String, priority: NotificationPriority = .high, completion: ((Error?) -> Void)? = nil) {
        sendNotification(
            title: "系统提醒",
            body: message,
            type: .systemAlert,
            priority: priority,
            userInfo: ["alert_message": message],
            completion: completion
        )
    }
    
    func scheduleReminder(title: String, body: String, date: Date, completion: ((Error?) -> Void)? = nil) {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = .default
        content.categoryIdentifier = NotificationType.reminder.categoryIdentifier
        
        let calendar = Calendar.current
        let components = calendar.dateComponents([.year, .month, .day, .hour, .minute], from: date)
        let trigger = UNCalendarNotificationTrigger(dateMatching: components, repeats: false)
        
        let identifier = "reminder_\\(UUID().uuidString)"
        let request = UNNotificationRequest(identifier: identifier, content: content, trigger: trigger)
        
        notificationCenter.add(request) { error in
            DispatchQueue.main.async {
                completion?(error)
            }
        }
    }
    
    func removeAllNotifications() {
        notificationCenter.removeAllPendingNotificationRequests()
        notificationCenter.removeAllDeliveredNotifications()
    }
    
    func removeNotifications(withIdentifiers identifiers: [String]) {
        notificationCenter.removePendingNotificationRequests(withIdentifiers: identifiers)
        notificationCenter.removeDeliveredNotifications(withIdentifiers: identifiers)
    }
    
    // MARK: - Badge Management
    
    func updateBadgeCount(_ count: Int) {
        DispatchQueue.main.async {
            NSApplication.shared.dockTile.badgeLabel = count > 0 ? "\\(count)" : nil
        }
    }
    
    func clearBadge() {
        updateBadgeCount(0)
    }
}

// MARK: - UNUserNotificationCenterDelegate

extension NotificationManager: UNUserNotificationCenterDelegate {
    
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        // 应用在前台时也显示通知
        completionHandler([.banner, .sound, .badge])
    }
    
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        let actionIdentifier = response.actionIdentifier
        
        // 处理不同的通知动作
        switch actionIdentifier {
        case "pet_interact":
            handlePetInteraction(userInfo: userInfo)
        case "pet_feed":
            handlePetFeeding(userInfo: userInfo)
        case "transfer_open":
            handleFileOpen(userInfo: userInfo)
        case "transfer_folder":
            handleFolderOpen(userInfo: userInfo)
        case "sync_view":
            handleSyncView(userInfo: userInfo)
        case "sync_retry":
            handleSyncRetry(userInfo: userInfo)
        case "ai_reply":
            handleAIReply(userInfo: userInfo)
        case "ai_copy":
            handleAICopy(userInfo: userInfo)
        case UNNotificationDefaultActionIdentifier:
            handleDefaultAction(userInfo: userInfo)
        default:
            break
        }
        
        completionHandler()
    }
    
    // MARK: - Action Handlers
    
    private func handlePetInteraction(userInfo: [AnyHashable: Any]) {
        if let petName = userInfo["pet_name"] as? String {
            NSLog("用户选择与宠物 \\(petName) 互动")
            // 触发宠物互动逻辑
            NotificationCenter.default.post(name: .petInteractionRequested, object: petName)
        }
    }
    
    private func handlePetFeeding(userInfo: [AnyHashable: Any]) {
        if let petName = userInfo["pet_name"] as? String {
            NSLog("用户选择喂食宠物 \\(petName)")
            // 触发宠物喂食逻辑
            NotificationCenter.default.post(name: .petFeedingRequested, object: petName)
        }
    }
    
    private func handleFileOpen(userInfo: [AnyHashable: Any]) {
        if let fileName = userInfo["file_name"] as? String {
            NSLog("用户选择打开文件 \\(fileName)")
            // 打开文件
            NotificationCenter.default.post(name: .fileOpenRequested, object: fileName)
        }
    }
    
    private func handleFolderOpen(userInfo: [AnyHashable: Any]) {
        if let fileName = userInfo["file_name"] as? String {
            NSLog("用户选择打开文件夹")
            // 打开包含文件的文件夹
            NotificationCenter.default.post(name: .folderOpenRequested, object: fileName)
        }
    }
    
    private func handleSyncView(userInfo: [AnyHashable: Any]) {
        NSLog("用户选择查看同步详情")
        // 显示同步详情界面
        NotificationCenter.default.post(name: .syncViewRequested, object: userInfo)
    }
    
    private func handleSyncRetry(userInfo: [AnyHashable: Any]) {
        NSLog("用户选择重试同步")
        // 重试数据同步
        NotificationCenter.default.post(name: .syncRetryRequested, object: userInfo)
    }
    
    private func handleAIReply(userInfo: [AnyHashable: Any]) {
        NSLog("用户选择回复AI")
        // 打开AI聊天界面
        NotificationCenter.default.post(name: .aiReplyRequested, object: userInfo)
    }
    
    private func handleAICopy(userInfo: [AnyHashable: Any]) {
        if let response = userInfo["ai_response"] as? String {
            NSLog("用户选择复制AI回复")
            // 复制到剪贴板
            let pasteboard = NSPasteboard.general
            pasteboard.clearContents()
            pasteboard.setString(response, forType: .string)
        }
    }
    
    private func handleDefaultAction(userInfo: [AnyHashable: Any]) {
        NSLog("用户点击了通知")
        // 显示主窗口
        NotificationCenter.default.post(name: .showMainWindowRequested, object: userInfo)
    }
}

// MARK: - Notification Names

extension Notification.Name {
    static let petInteractionRequested = Notification.Name("petInteractionRequested")
    static let petFeedingRequested = Notification.Name("petFeedingRequested")
    static let fileOpenRequested = Notification.Name("fileOpenRequested")
    static let folderOpenRequested = Notification.Name("folderOpenRequested")
    static let syncViewRequested = Notification.Name("syncViewRequested")
    static let syncRetryRequested = Notification.Name("syncRetryRequested")
    static let aiReplyRequested = Notification.Name("aiReplyRequested")
    static let showMainWindowRequested = Notification.Name("showMainWindowRequested")
}