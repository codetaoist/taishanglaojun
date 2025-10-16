import Foundation
import WatchConnectivity
import Combine

// MARK: - 手表连接管理器
class WatchConnectivityManager: NSObject, ObservableObject, WCSessionDelegate {
    static let shared = WatchConnectivityManager()
    
    // MARK: - Published Properties
    @Published var isConnected: Bool = false
    @Published var isReachable: Bool = false
    @Published var connectionStatus: ConnectionStatus = .disconnected
    @Published var lastSyncTime: Date?
    
    // MARK: - Private Properties
    private var session: WCSession?
    private var messageQueue: [WatchMessage] = []
    private var isProcessingQueue: Bool = false
    private var heartbeatTimer: Timer?
    private var reconnectTimer: Timer?
    
    // MARK: - Configuration
    private let heartbeatInterval: TimeInterval = 30.0
    private let reconnectInterval: TimeInterval = 5.0
    private let maxRetryAttempts: Int = 3
    
    override init() {
        super.init()
        setupWatchConnectivity()
    }
    
    // MARK: - Public Methods
    
    func activate() {
        guard WCSession.isSupported() else {
            print("WatchConnectivity not supported")
            return
        }
        
        session?.activate()
        startHeartbeat()
    }
    
    func deactivate() {
        stopHeartbeat()
        stopReconnectTimer()
        session = nil
    }
    
    // MARK: - 数据请求方法
    
    func requestTaskData(completion: @escaping (Result<[WatchTask], WatchError>) -> Void) {
        sendMessage(
            WatchMessage(
                action: .getTasks,
                data: nil
            ),
            completion: { result in
                switch result {
                case .success(let response):
                    if let tasksData = response["tasks"] as? Data {
                        do {
                            let tasks = try JSONDecoder().decode([WatchTask].self, from: tasksData)
                            completion(.success(tasks))
                        } catch {
                            completion(.failure(.decodingError(error)))
                        }
                    } else {
                        completion(.failure(.invalidResponse))
                    }
                case .failure(let error):
                    completion(.failure(error))
                }
            }
        )
    }
    
    func acceptTask(_ taskId: String, completion: @escaping (Result<Bool, WatchError>) -> Void) {
        let taskData = ["taskId": taskId]
        
        sendMessage(
            WatchMessage(
                action: .acceptTask,
                data: taskData
            ),
            completion: { result in
                switch result {
                case .success(let response):
                    let success = response["success"] as? Bool ?? false
                    completion(.success(success))
                case .failure(let error):
                    completion(.failure(error))
                }
            }
        )
    }
    
    func updateTaskProgress(_ taskId: String, progress: Double, completion: @escaping (Result<Bool, WatchError>) -> Void) {
        let progressData = [
            "taskId": taskId,
            "progress": progress
        ] as [String : Any]
        
        sendMessage(
            WatchMessage(
                action: .updateProgress,
                data: progressData
            ),
            completion: { result in
                switch result {
                case .success(let response):
                    let success = response["success"] as? Bool ?? false
                    completion(.success(success))
                case .failure(let error):
                    completion(.failure(error))
                }
            }
        )
    }
    
    func forceSync(completion: @escaping (Result<Bool, WatchError>) -> Void) {
        sendMessage(
            WatchMessage(
                action: .forceSync,
                data: nil
            ),
            completion: { result in
                switch result {
                case .success:
                    self.lastSyncTime = Date()
                    completion(.success(true))
                case .failure(let error):
                    completion(.failure(error))
                }
            }
        )
    }
    
    // MARK: - Private Methods
    
    private func setupWatchConnectivity() {
        guard WCSession.isSupported() else { return }
        
        session = WCSession.default
        session?.delegate = self
        
        updateConnectionStatus()
    }
    
    private func sendMessage(_ message: WatchMessage, completion: @escaping (Result<[String: Any], WatchError>) -> Void) {
        guard let session = session else {
            completion(.failure(.sessionNotAvailable))
            return
        }
        
        guard session.isReachable else {
            // 添加到队列等待连接恢复
            messageQueue.append(message)
            completion(.failure(.phoneNotReachable))
            return
        }
        
        do {
            let messageData = try JSONEncoder().encode(message)
            let messageDict = try JSONSerialization.jsonObject(with: messageData) as? [String: Any] ?? [:]
            
            session.sendMessage(messageDict, replyHandler: { response in
                completion(.success(response))
            }, errorHandler: { error in
                completion(.failure(.communicationError(error)))
            })
        } catch {
            completion(.failure(.encodingError(error)))
        }
    }
    
    private func processMessageQueue() {
        guard !isProcessingQueue && !messageQueue.isEmpty else { return }
        
        isProcessingQueue = true
        
        let message = messageQueue.removeFirst()
        sendMessage(message) { [weak self] result in
            self?.isProcessingQueue = false
            
            // 继续处理队列中的其他消息
            if !self?.messageQueue.isEmpty ?? true {
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                    self?.processMessageQueue()
                }
            }
        }
    }
    
    private func startHeartbeat() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: heartbeatInterval, repeats: true) { [weak self] _ in
            self?.sendHeartbeat()
        }
    }
    
    private func stopHeartbeat() {
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
    }
    
    private func sendHeartbeat() {
        sendMessage(
            WatchMessage(action: .heartbeat, data: nil)
        ) { [weak self] result in
            switch result {
            case .success:
                self?.updateConnectionStatus()
            case .failure:
                self?.handleConnectionLoss()
            }
        }
    }
    
    private func handleConnectionLoss() {
        connectionStatus = .disconnected
        isReachable = false
        startReconnectTimer()
    }
    
    private func startReconnectTimer() {
        stopReconnectTimer()
        
        reconnectTimer = Timer.scheduledTimer(withTimeInterval: reconnectInterval, repeats: true) { [weak self] _ in
            self?.attemptReconnection()
        }
    }
    
    private func stopReconnectTimer() {
        reconnectTimer?.invalidate()
        reconnectTimer = nil
    }
    
    private func attemptReconnection() {
        guard let session = session else { return }
        
        if session.isReachable {
            stopReconnectTimer()
            updateConnectionStatus()
            processMessageQueue()
        }
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
        
        if isConnected && isReachable {
            connectionStatus = .connected
        } else if isConnected {
            connectionStatus = .paired
        } else {
            connectionStatus = .disconnected
        }
    }
    
    // MARK: - WCSessionDelegate
    
    func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        DispatchQueue.main.async {
            self.updateConnectionStatus()
            
            if activationState == .activated {
                self.processMessageQueue()
            }
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
    
    private func handleReceivedMessage(_ message: [String: Any]) {
        guard let actionString = message["action"] as? String,
              let action = WatchMessageAction(rawValue: actionString) else {
            return
        }
        
        switch action {
        case .newTask:
            handleNewTaskNotification(message)
        case .taskUpdate:
            handleTaskUpdate(message)
        case .syncComplete:
            handleSyncComplete(message)
        case .notification:
            handleNotification(message)
        default:
            break
        }
    }
    
    private func handleNewTaskNotification(_ message: [String: Any]) {
        // 显示新任务通知
        NotificationCenter.default.post(name: .newTaskReceived, object: message)
        
        // 触觉反馈
        WKInterfaceDevice.current().play(.notification)
    }
    
    private func handleTaskUpdate(_ message: [String: Any]) {
        // 更新任务状态
        NotificationCenter.default.post(name: .taskUpdated, object: message)
    }
    
    private func handleSyncComplete(_ message: [String: Any]) {
        lastSyncTime = Date()
        NotificationCenter.default.post(name: .syncCompleted, object: message)
    }
    
    private func handleNotification(_ message: [String: Any]) {
        NotificationCenter.default.post(name: .watchNotificationReceived, object: message)
        
        // 根据通知类型播放不同的触觉反馈
        if let type = message["type"] as? String {
            switch type {
            case "success":
                WKInterfaceDevice.current().play(.success)
            case "error":
                WKInterfaceDevice.current().play(.failure)
            default:
                WKInterfaceDevice.current().play(.notification)
            }
        }
    }
}

// MARK: - 支持类型和枚举

enum ConnectionStatus {
    case disconnected
    case paired
    case connected
    
    var displayName: String {
        switch self {
        case .disconnected: return "未连接"
        case .paired: return "已配对"
        case .connected: return "已连接"
        }
    }
}

struct WatchMessage: Codable {
    let id: String
    let action: WatchMessageAction
    let data: [String: Any]?
    let timestamp: Date
    
    init(action: WatchMessageAction, data: [String: Any]?) {
        self.id = UUID().uuidString
        self.action = action
        self.data = data
        self.timestamp = Date()
    }
    
    enum CodingKeys: String, CodingKey {
        case id, action, data, timestamp
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        action = try container.decode(WatchMessageAction.self, forKey: .action)
        timestamp = try container.decode(Date.self, forKey: .timestamp)
        
        // 处理 data 字段的解码
        if let dataDict = try container.decodeIfPresent([String: AnyCodable].self, forKey: .data) {
            data = dataDict.mapValues { $0.value }
        } else {
            data = nil
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(id, forKey: .id)
        try container.encode(action, forKey: .action)
        try container.encode(timestamp, forKey: .timestamp)
        
        if let data = data {
            let codableData = data.mapValues { AnyCodable($0) }
            try container.encode(codableData, forKey: .data)
        }
    }
}

enum WatchMessageAction: String, Codable {
    case getTasks = "getTasks"
    case acceptTask = "acceptTask"
    case updateProgress = "updateProgress"
    case forceSync = "forceSync"
    case heartbeat = "heartbeat"
    case newTask = "newTask"
    case taskUpdate = "taskUpdate"
    case syncComplete = "syncComplete"
    case notification = "notification"
}

enum WatchError: Error, LocalizedError {
    case sessionNotAvailable
    case phoneNotReachable
    case communicationError(Error)
    case encodingError(Error)
    case decodingError(Error)
    case invalidResponse
    case syncFailed
    case taskNotFound
    
    var errorDescription: String? {
        switch self {
        case .sessionNotAvailable:
            return "Watch连接会话不可用"
        case .phoneNotReachable:
            return "无法连接到iPhone"
        case .communicationError(let error):
            return "通信错误: \(error.localizedDescription)"
        case .encodingError(let error):
            return "数据编码错误: \(error.localizedDescription)"
        case .decodingError(let error):
            return "数据解码错误: \(error.localizedDescription)"
        case .invalidResponse:
            return "无效的响应数据"
        case .syncFailed:
            return "数据同步失败"
        case .taskNotFound:
            return "任务未找到"
        }
    }
}

// MARK: - 辅助类型
struct AnyCodable: Codable {
    let value: Any
    
    init(_ value: Any) {
        self.value = value
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if let intValue = try? container.decode(Int.self) {
            value = intValue
        } else if let doubleValue = try? container.decode(Double.self) {
            value = doubleValue
        } else if let stringValue = try? container.decode(String.self) {
            value = stringValue
        } else if let boolValue = try? container.decode(Bool.self) {
            value = boolValue
        } else {
            throw DecodingError.typeMismatch(AnyCodable.self, DecodingError.Context(codingPath: decoder.codingPath, debugDescription: "Unsupported type"))
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        if let intValue = value as? Int {
            try container.encode(intValue)
        } else if let doubleValue = value as? Double {
            try container.encode(doubleValue)
        } else if let stringValue = value as? String {
            try container.encode(stringValue)
        } else if let boolValue = value as? Bool {
            try container.encode(boolValue)
        } else {
            throw EncodingError.invalidValue(value, EncodingError.Context(codingPath: encoder.codingPath, debugDescription: "Unsupported type"))
        }
    }
}

// MARK: - Notification Names
extension Notification.Name {
    static let newTaskReceived = Notification.Name("newTaskReceived")
    static let taskUpdated = Notification.Name("taskUpdated")
    static let syncCompleted = Notification.Name("syncCompleted")
    static let watchNotificationReceived = Notification.Name("watchNotificationReceived")
}