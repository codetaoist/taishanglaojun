import Foundation
import UserNotifications
import WatchKit

class WatchNotificationManager: NSObject, ObservableObject {
    static let shared = WatchNotificationManager()
    
    @Published var notificationPermissionGranted = false
    @Published var notifications: [WatchNotification] = []
    @Published var unreadCount = 0
    
    private let userNotificationCenter = UNUserNotificationCenter.current()
    
    override init() {
        super.init()
        userNotificationCenter.delegate = self
        requestNotificationPermission()
        loadStoredNotifications()
    }
    
    // MARK: - Permission Management
    
    private func requestNotificationPermission() {
        userNotificationCenter.requestAuthorization(options: [.alert, .sound, .badge]) { [weak self] granted, error in
            DispatchQueue.main.async {
                self?.notificationPermissionGranted = granted
            }
            
            if let error = error {
                print("Notification permission error: \(error)")
            }
        }
    }
    
    // MARK: - Task Notifications
    
    func scheduleTaskAcceptedNotification(for task: WatchTask) {
        let content = UNMutableNotificationContent()
        content.title = "任务已接受"
        content.body = "任务「\(task.title)」已成功接受"
        content.sound = .default
        content.userInfo = [
            "task_id": task.id,
            "notification_type": "task_accepted"
        ]
        
        let request = UNNotificationRequest(
            identifier: "task_accepted_\(task.id)",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule task accepted notification: \(error)")
            }
        }
        
        addNotification(
            id: "task_accepted_\(task.id)",
            type: .task,
            title: content.title,
            message: content.body,
            taskId: task.id
        )
    }
    
    func scheduleTaskStartedNotification(for task: WatchTask) {
        let content = UNMutableNotificationContent()
        content.title = "任务已开始"
        content.body = "任务「\(task.title)」已开始执行"
        content.sound = .default
        content.userInfo = [
            "task_id": task.id,
            "notification_type": "task_started"
        ]
        
        let request = UNNotificationRequest(
            identifier: "task_started_\(task.id)",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule task started notification: \(error)")
            }
        }
        
        addNotification(
            id: "task_started_\(task.id)",
            type: .task,
            title: content.title,
            message: content.body,
            taskId: task.id
        )
    }
    
    func scheduleTaskCompletedNotification(for task: WatchTask) {
        let content = UNMutableNotificationContent()
        content.title = "任务已完成"
        content.body = "恭喜！任务「\(task.title)」已完成"
        content.sound = .default
        content.userInfo = [
            "task_id": task.id,
            "notification_type": "task_completed"
        ]
        
        let request = UNNotificationRequest(
            identifier: "task_completed_\(task.id)",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule task completed notification: \(error)")
            }
        }
        
        addNotification(
            id: "task_completed_\(task.id)",
            type: .task,
            title: content.title,
            message: content.body,
            taskId: task.id
        )
        
        // Trigger haptic feedback for completion
        WKInterfaceDevice.current().play(.success)
    }
    
    func scheduleTaskReminderNotification(for task: WatchTask, delay: TimeInterval = 0) {
        let content = UNMutableNotificationContent()
        content.title = "任务提醒"
        content.body = "任务「\(task.title)」需要您的关注"
        content.sound = .default
        content.userInfo = [
            "task_id": task.id,
            "notification_type": "task_reminder"
        ]
        
        let trigger = delay > 0 ? UNTimeIntervalNotificationTrigger(timeInterval: delay, repeats: false) : nil
        
        let request = UNNotificationRequest(
            identifier: "task_reminder_\(task.id)",
            content: content,
            trigger: trigger
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule task reminder notification: \(error)")
            }
        }
        
        if delay == 0 {
            addNotification(
                id: "task_reminder_\(task.id)",
                type: .task,
                title: content.title,
                message: content.body,
                taskId: task.id
            )
        }
    }
    
    // MARK: - Location Notifications
    
    func scheduleLocationArrivalNotification(for task: WatchTask, distance: Double) {
        let distanceText = distance < 1000 ? "\(Int(distance))米" : "\(Int(distance / 1000))公里"
        
        let content = UNMutableNotificationContent()
        content.title = "已到达任务位置"
        content.body = "您距离任务「\(task.title)」仅\(distanceText)"
        content.sound = .default
        content.userInfo = [
            "task_id": task.id,
            "notification_type": "location_arrival",
            "distance": distance
        ]
        
        let request = UNNotificationRequest(
            identifier: "location_arrival_\(task.id)",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule location arrival notification: \(error)")
            }
        }
        
        addNotification(
            id: "location_arrival_\(task.id)",
            type: .task,
            title: content.title,
            message: content.body,
            taskId: task.id
        )
        
        // Trigger haptic feedback for location arrival
        WKInterfaceDevice.current().play(.notification)
    }
    
    // MARK: - System Notifications
    
    func scheduleConnectionLostNotification() {
        let content = UNMutableNotificationContent()
        content.title = "连接断开"
        content.body = "与iPhone的连接已断开"
        content.sound = .default
        content.userInfo = ["notification_type": "connection_lost"]
        
        let request = UNNotificationRequest(
            identifier: "connection_lost",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule connection lost notification: \(error)")
            }
        }
        
        addNotification(
            id: "connection_lost",
            type: .system,
            title: content.title,
            message: content.body
        )
    }
    
    func scheduleConnectionRestoredNotification() {
        let content = UNMutableNotificationContent()
        content.title = "连接已恢复"
        content.body = "与iPhone的连接已恢复"
        content.sound = .default
        content.userInfo = ["notification_type": "connection_restored"]
        
        let request = UNNotificationRequest(
            identifier: "connection_restored",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule connection restored notification: \(error)")
            }
        }
        
        addNotification(
            id: "connection_restored",
            type: .system,
            title: content.title,
            message: content.body
        )
        
        // Remove connection lost notification
        removeNotification(id: "connection_lost")
    }
    
    func scheduleSyncCompletedNotification(taskCount: Int) {
        let content = UNMutableNotificationContent()
        content.title = "同步完成"
        content.body = "已同步 \(taskCount) 个任务"
        content.sound = .default
        content.userInfo = [
            "notification_type": "sync_completed",
            "task_count": taskCount
        ]
        
        let request = UNNotificationRequest(
            identifier: "sync_completed",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule sync completed notification: \(error)")
            }
        }
        
        addNotification(
            id: "sync_completed",
            type: .system,
            title: content.title,
            message: content.body
        )
    }
    
    func scheduleSyncErrorNotification(error: String) {
        let content = UNMutableNotificationContent()
        content.title = "同步失败"
        content.body = "同步失败: \(error)"
        content.sound = .default
        content.userInfo = [
            "notification_type": "sync_error",
            "error": error
        ]
        
        let request = UNNotificationRequest(
            identifier: "sync_error",
            content: content,
            trigger: nil
        )
        
        userNotificationCenter.add(request) { error in
            if let error = error {
                print("Failed to schedule sync error notification: \(error)")
            }
        }
        
        addNotification(
            id: "sync_error",
            type: .system,
            title: content.title,
            message: content.body
        )
    }
    
    // MARK: - Notification Management
    
    private func addNotification(
        id: String,
        type: NotificationType,
        title: String,
        message: String,
        taskId: String? = nil
    ) {
        let notification = WatchNotification(
            id: id,
            type: type,
            title: title,
            message: message,
            timestamp: Date(),
            isRead: false,
            taskId: taskId
        )
        
        DispatchQueue.main.async {
            self.notifications.insert(notification, at: 0)
            self.updateUnreadCount()
            self.saveNotifications()
        }
    }
    
    func markAsRead(_ notification: WatchNotification) {
        if let index = notifications.firstIndex(where: { $0.id == notification.id }) {
            notifications[index].isRead = true
            updateUnreadCount()
            saveNotifications()
        }
    }
    
    func markAllAsRead() {
        for index in notifications.indices {
            notifications[index].isRead = true
        }
        updateUnreadCount()
        saveNotifications()
    }
    
    func removeNotification(id: String) {
        notifications.removeAll { $0.id == id }
        userNotificationCenter.removePendingNotificationRequests(withIdentifiers: [id])
        updateUnreadCount()
        saveNotifications()
    }
    
    func clearAllNotifications() {
        notifications.removeAll()
        userNotificationCenter.removeAllPendingNotificationRequests()
        updateUnreadCount()
        saveNotifications()
    }
    
    func clearTaskNotifications(taskId: String) {
        let taskNotificationIds = notifications
            .filter { $0.taskId == taskId }
            .map { $0.id }
        
        notifications.removeAll { $0.taskId == taskId }
        userNotificationCenter.removePendingNotificationRequests(withIdentifiers: taskNotificationIds)
        updateUnreadCount()
        saveNotifications()
    }
    
    private func updateUnreadCount() {
        unreadCount = notifications.filter { !$0.isRead }.count
    }
    
    // MARK: - Persistence
    
    private func saveNotifications() {
        if let data = try? JSONEncoder().encode(notifications) {
            UserDefaults.standard.set(data, forKey: "watch_notifications")
        }
    }
    
    private func loadStoredNotifications() {
        guard let data = UserDefaults.standard.data(forKey: "watch_notifications"),
              let storedNotifications = try? JSONDecoder().decode([WatchNotification].self, from: data) else {
            return
        }
        
        // Keep only recent notifications (last 50)
        notifications = Array(storedNotifications.prefix(50))
        updateUnreadCount()
    }
    
    // MARK: - Filtering
    
    func getNotifications(for tab: NotificationTab) -> [WatchNotification] {
        switch tab {
        case .all:
            return notifications
        case .unread:
            return notifications.filter { !$0.isRead }
        case .tasks:
            return notifications.filter { $0.type == .task }
        case .system:
            return notifications.filter { $0.type == .system }
        }
    }
    
    func getNotificationCounts() -> NotificationCounts {
        let unreadCount = notifications.filter { !$0.isRead }.count
        let taskCount = notifications.filter { $0.type == .task }.count
        let systemCount = notifications.filter { $0.type == .system }.count
        
        return NotificationCounts(
            all: notifications.count,
            unread: unreadCount,
            tasks: taskCount,
            system: systemCount
        )
    }
}

// MARK: - UNUserNotificationCenterDelegate

extension WatchNotificationManager: UNUserNotificationCenterDelegate {
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        // Show notification even when app is in foreground
        completionHandler([.banner, .sound])
    }
    
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        
        // Handle notification tap
        if let taskId = userInfo["task_id"] as? String {
            // Navigate to task detail or perform action
            NotificationCenter.default.post(
                name: .taskNotificationTapped,
                object: nil,
                userInfo: ["task_id": taskId]
            )
        }
        
        completionHandler()
    }
}

// MARK: - Notification Names

extension Notification.Name {
    static let taskNotificationTapped = Notification.Name("taskNotificationTapped")
}