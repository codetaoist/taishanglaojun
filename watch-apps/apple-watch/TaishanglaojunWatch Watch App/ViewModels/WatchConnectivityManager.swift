import Foundation
import WatchConnectivity
import Combine

@MainActor
class WatchConnectivityManager: NSObject, ObservableObject {
    // MARK: - Published Properties
    @Published var isConnected = false
    @Published var isReachable = false
    @Published var lastSyncTime: Date?
    @Published var isSyncing = false
    @Published var error: String?
    @Published var receivedTasks: [WatchTask]?
    @Published var connectionStatus: ConnectionStatus = .disconnected
    
    // MARK: - Private Properties
    private var session: WCSession?
    private let userDefaults = UserDefaults.standard
    private let lastSyncKey = "last_sync_time"
    
    // MARK: - Message Keys
    private enum MessageKey {
        static let requestTasks = "request_tasks"
        static let acceptTask = "accept_task"
        static let updateProgress = "update_progress"
        static let forceSync = "force_sync"
        static let heartbeat = "heartbeat"
        static let taskData = "task_data"
        static let taskId = "task_id"
        static let progress = "progress"
        static let note = "note"
        static let timestamp = "timestamp"
    }
    
    // MARK: - Initialization
    override init() {
        super.init()
        setupWatchConnectivity()
        loadLastSyncTime()
    }
    
    // MARK: - Setup
    private func setupWatchConnectivity() {
        guard WCSession.isSupported() else {
            error = "Watch Connectivity 不支持"
            return
        }
        
        session = WCSession.default
        session?.delegate = self
        session?.activate()
    }
    
    private func loadLastSyncTime() {
        lastSyncTime = userDefaults.object(forKey: lastSyncKey) as? Date
    }
    
    // MARK: - Public Methods
    func requestTasks() async {
        guard let session = session, session.isReachable else {
            error = "iPhone 不可达"
            return
        }
        
        isSyncing = true
        error = nil
        
        let message = [
            MessageKey.requestTasks: true,
            MessageKey.timestamp: Date().timeIntervalSince1970
        ] as [String: Any]
        
        do {
            _ = try await session.sendMessage(message)
        } catch {
            self.error = "请求任务失败: \(error.localizedDescription)"
        }
        
        isSyncing = false
    }
    
    func acceptTask(_ taskId: String) async {
        guard let session = session, session.isReachable else {
            error = "iPhone 不可达"
            return
        }
        
        let message = [
            MessageKey.acceptTask: true,
            MessageKey.taskId: taskId,
            MessageKey.timestamp: Date().timeIntervalSince1970
        ] as [String: Any]
        
        do {
            _ = try await session.sendMessage(message)
        } catch {
            self.error = "接受任务失败: \(error.localizedDescription)"
        }
    }
    
    func updateTaskProgress(_ taskId: String, progress: Int, note: String) async {
        guard let session = session else { return }
        
        let message = [
            MessageKey.updateProgress: true,
            MessageKey.taskId: taskId,
            MessageKey.progress: progress,
            MessageKey.note: note,
            MessageKey.timestamp: Date().timeIntervalSince1970
        ] as [String: Any]
        
        do {
            if session.isReachable {
                _ = try await session.sendMessage(message)
            } else {
                // Store for later transmission
                try session.updateApplicationContext(message)
            }
        } catch {
            self.error = "更新进度失败: \(error.localizedDescription)"
        }
    }
    
    func forceSync() async {
        guard let session = session, session.isReachable else {
            error = "iPhone 不可达"
            return
        }
        
        isSyncing = true
        error = nil
        
        let message = [
            MessageKey.forceSync: true,
            MessageKey.timestamp: Date().timeIntervalSince1970
        ] as [String: Any]
        
        do {
            _ = try await session.sendMessage(message)
            lastSyncTime = Date()
            userDefaults.set(lastSyncTime, forKey: lastSyncKey)
        } catch {
            self.error = "强制同步失败: \(error.localizedDescription)"
        }
        
        isSyncing = false
    }
    
    func sendHeartbeat() async {
        guard let session = session, session.isReachable else { return }
        
        let message = [
            MessageKey.heartbeat: true,
            MessageKey.timestamp: Date().timeIntervalSince1970
        ] as [String: Any]
        
        do {
            _ = try await session.sendMessage(message)
        } catch {
            // Heartbeat failures are not critical
            print("Heartbeat failed: \(error)")
        }
    }
    
    // MARK: - Utility Methods
    func clearError() {
        error = nil
    }
    
    private func updateConnectionStatus() {
        guard let session = session else {
            connectionStatus = .disconnected
            isConnected = false
            isReachable = false
            return
        }
        
        isConnected = session.isPaired && session.isWatchAppInstalled
        isReachable = session.isReachable
        
        if !session.isPaired {
            connectionStatus = .disconnected
        } else if !session.isWatchAppInstalled {
            connectionStatus = .appNotInstalled
        } else if !session.isReachable {
            connectionStatus = .notReachable
        } else {
            connectionStatus = .connected
        }
    }
    
    private func handleReceivedTasks(_ data: Data) {
        do {
            let tasks = try JSONDecoder().decode([WatchTask].self, from: data)
            receivedTasks = tasks
            lastSyncTime = Date()
            userDefaults.set(lastSyncTime, forKey: lastSyncKey)
        } catch {
            self.error = "解析任务数据失败: \(error.localizedDescription)"
        }
    }
}

// MARK: - WCSessionDelegate
extension WatchConnectivityManager: WCSessionDelegate {
    func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        DispatchQueue.main.async {
            if let error = error {
                self.error = "连接激活失败: \(error.localizedDescription)"
            }
            self.updateConnectionStatus()
        }
    }
    
    func sessionReachabilityDidChange(_ session: WCSession) {
        DispatchQueue.main.async {
            self.updateConnectionStatus()
        }
    }
    
    func session(_ session: WCSession, didReceiveMessage message: [String : Any]) {
        DispatchQueue.main.async {
            self.handleReceivedMessage(message)
        }
    }
    
    func session(_ session: WCSession, didReceiveMessage message: [String : Any], replyHandler: @escaping ([String : Any]) -> Void) {
        DispatchQueue.main.async {
            self.handleReceivedMessage(message)
            replyHandler(["status": "received"])
        }
    }
    
    func session(_ session: WCSession, didReceiveApplicationContext applicationContext: [String : Any]) {
        DispatchQueue.main.async {
            self.handleReceivedMessage(applicationContext)
        }
    }
    
    func session(_ session: WCSession, didReceiveUserInfo userInfo: [String : Any] = [:]) {
        DispatchQueue.main.async {
            self.handleReceivedMessage(userInfo)
        }
    }
    
    func session(_ session: WCSession, didReceive file: WCSessionFile) {
        DispatchQueue.main.async {
            do {
                let data = try Data(contentsOf: file.fileURL)
                self.handleReceivedTasks(data)
            } catch {
                self.error = "接收文件失败: \(error.localizedDescription)"
            }
        }
    }
    
    private func handleReceivedMessage(_ message: [String: Any]) {
        if let taskData = message[MessageKey.taskData] as? Data {
            handleReceivedTasks(taskData)
        }
        
        if message[MessageKey.heartbeat] as? Bool == true {
            // Heartbeat received, connection is healthy
            lastSyncTime = Date()
        }
    }
}

// MARK: - Supporting Types
enum ConnectionStatus {
    case connected
    case disconnected
    case notReachable
    case appNotInstalled
    
    var description: String {
        switch self {
        case .connected:
            return "已连接"
        case .disconnected:
            return "未连接"
        case .notReachable:
            return "不可达"
        case .appNotInstalled:
            return "应用未安装"
        }
    }
    
    var isHealthy: Bool {
        return self == .connected
    }
}