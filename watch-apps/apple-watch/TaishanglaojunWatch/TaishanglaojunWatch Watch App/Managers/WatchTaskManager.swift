import Foundation
import Combine
import WatchKit

// MARK: - 手表任务管理器
class WatchTaskManager: ObservableObject {
    static let shared = WatchTaskManager()
    
    // MARK: - Published Properties
    @Published var tasks: [WatchTask] = []
    @Published var activeTasks: [WatchTask] = []
    @Published var completedTasks: [WatchTask] = []
    @Published var isLoading: Bool = false
    @Published var lastUpdateTime: Date?
    @Published var errorMessage: String?
    
    // MARK: - Private Properties
    private var connectivityManager: WatchConnectivityManager
    private var cancellables = Set<AnyCancellable>()
    private var autoRefreshTimer: Timer?
    private var taskProgressTimer: Timer?
    
    // MARK: - Configuration
    private let autoRefreshInterval: TimeInterval = 300 // 5分钟
    private let progressUpdateInterval: TimeInterval = 30 // 30秒
    private let maxCachedTasks: Int = 50
    
    init(connectivityManager: WatchConnectivityManager = .shared) {
        self.connectivityManager = connectivityManager
        setupObservers()
        setupAutoRefresh()
    }
    
    deinit {
        stopAutoRefresh()
        stopProgressTimer()
    }
    
    // MARK: - Public Methods
    
    func loadTasks() {
        guard !isLoading else { return }
        
        isLoading = true
        errorMessage = nil
        
        connectivityManager.requestTaskData { [weak self] result in
            DispatchQueue.main.async {
                self?.isLoading = false
                
                switch result {
                case .success(let tasks):
                    self?.updateTasks(tasks)
                    self?.lastUpdateTime = Date()
                    self?.saveTasksToCache(tasks)
                    
                case .failure(let error):
                    self?.errorMessage = error.localizedDescription
                    // 如果网络请求失败，尝试加载缓存的任务
                    self?.loadTasksFromCache()
                }
            }
        }
    }
    
    func acceptTask(_ task: WatchTask) {
        guard task.status == .available else { return }
        
        connectivityManager.acceptTask(task.id) { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success(let success):
                    if success {
                        self?.updateTaskStatus(task.id, status: .accepted)
                        self?.showSuccessNotification("任务已接受")
                        WKInterfaceDevice.current().play(.success)
                    } else {
                        self?.showErrorNotification("任务接受失败")
                    }
                    
                case .failure(let error):
                    self?.showErrorNotification("接受任务时出错: \(error.localizedDescription)")
                }
            }
        }
    }
    
    func startTask(_ task: WatchTask) {
        guard task.status == .accepted else { return }
        
        updateTaskStatus(task.id, status: .inProgress)
        startProgressTracking(for: task)
        showSuccessNotification("任务已开始")
        WKInterfaceDevice.current().play(.start)
    }
    
    func completeTask(_ task: WatchTask) {
        guard task.status == .inProgress else { return }
        
        updateTaskStatus(task.id, status: .completed)
        updateTaskProgress(task.id, progress: 1.0)
        stopProgressTracking(for: task)
        showSuccessNotification("任务已完成")
        WKInterfaceDevice.current().play(.success)
    }
    
    func updateTaskProgress(_ taskId: String, progress: Double) {
        if let index = tasks.firstIndex(where: { $0.id == taskId }) {
            tasks[index].progress = min(max(progress, 0.0), 1.0)
            
            // 同步到手机端
            connectivityManager.updateTaskProgress(taskId, progress: progress) { result in
                if case .failure(let error) = result {
                    print("Progress update failed: \(error)")
                }
            }
        }
    }
    
    func refreshTasks() {
        loadTasks()
    }
    
    func forceSync() {
        connectivityManager.forceSync { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success:
                    self?.loadTasks()
                    self?.showSuccessNotification("同步完成")
                    
                case .failure(let error):
                    self?.showErrorNotification("同步失败: \(error.localizedDescription)")
                }
            }
        }
    }
    
    // MARK: - Task Filtering and Sorting
    
    func getTasksByStatus(_ status: TaskStatus) -> [WatchTask] {
        return tasks.filter { $0.status == status }
    }
    
    func getTasksByPriority(_ priority: TaskPriority) -> [WatchTask] {
        return tasks.filter { $0.priority == priority }
    }
    
    func getQuickActionTasks() -> [WatchTask] {
        return tasks.filter { $0.isQuickActionAvailable && $0.status != .completed }
    }
    
    func getOverdueTasks() -> [WatchTask] {
        return tasks.filter { $0.isOverdue }
    }
    
    func sortTasksByPriority() -> [WatchTask] {
        return tasks.sorted { task1, task2 in
            if task1.priority != task2.priority {
                return task1.priority.rawValue > task2.priority.rawValue
            }
            return task1.createdAt > task2.createdAt
        }
    }
    
    // MARK: - Private Methods
    
    private func setupObservers() {
        // 监听连接状态变化
        connectivityManager.$isConnected
            .sink { [weak self] isConnected in
                if isConnected {
                    self?.loadTasks()
                }
            }
            .store(in: &cancellables)
        
        // 监听新任务通知
        NotificationCenter.default.publisher(for: .newTaskReceived)
            .sink { [weak self] notification in
                self?.handleNewTaskNotification(notification)
            }
            .store(in: &cancellables)
        
        // 监听任务更新通知
        NotificationCenter.default.publisher(for: .taskUpdated)
            .sink { [weak self] notification in
                self?.handleTaskUpdateNotification(notification)
            }
            .store(in: &cancellables)
    }
    
    private func updateTasks(_ newTasks: [WatchTask]) {
        tasks = newTasks
        updateTaskCategories()
    }
    
    private func updateTaskCategories() {
        activeTasks = tasks.filter { $0.status == .accepted || $0.status == .inProgress }
        completedTasks = tasks.filter { $0.status == .completed }
    }
    
    private func updateTaskStatus(_ taskId: String, status: TaskStatus) {
        if let index = tasks.firstIndex(where: { $0.id == taskId }) {
            tasks[index].status = status
            tasks[index].updatedAt = Date()
            updateTaskCategories()
        }
    }
    
    private func setupAutoRefresh() {
        autoRefreshTimer = Timer.scheduledTimer(withTimeInterval: autoRefreshInterval, repeats: true) { [weak self] _ in
            if self?.connectivityManager.isConnected == true {
                self?.loadTasks()
            }
        }
    }
    
    private func stopAutoRefresh() {
        autoRefreshTimer?.invalidate()
        autoRefreshTimer = nil
    }
    
    private func startProgressTracking(for task: WatchTask) {
        // 为进行中的任务启动进度跟踪
        taskProgressTimer = Timer.scheduledTimer(withTimeInterval: progressUpdateInterval, repeats: true) { [weak self] _ in
            self?.trackTaskProgress(task)
        }
    }
    
    private func stopProgressTracking(for task: WatchTask) {
        taskProgressTimer?.invalidate()
        taskProgressTimer = nil
    }
    
    private func trackTaskProgress(_ task: WatchTask) {
        // 根据任务类型自动更新进度
        // 这里可以集成位置跟踪、时间跟踪等功能
        
        guard let index = tasks.firstIndex(where: { $0.id == task.id }),
              tasks[index].status == .inProgress else {
            return
        }
        
        // 示例：基于时间的进度更新
        let elapsed = Date().timeIntervalSince(tasks[index].updatedAt)
        let estimatedDuration = TimeInterval(task.difficulty * 3600) // 难度 * 1小时
        let progress = min(elapsed / estimatedDuration, 0.95) // 最多95%，需要手动完成
        
        updateTaskProgress(task.id, progress: progress)
    }
    
    private func handleNewTaskNotification(_ notification: Notification) {
        guard let message = notification.object as? [String: Any],
              let taskData = message["task"] as? Data else {
            return
        }
        
        do {
            let newTask = try JSONDecoder().decode(WatchTask.self, from: taskData)
            
            DispatchQueue.main.async {
                self.tasks.append(newTask)
                self.updateTaskCategories()
                self.showNewTaskNotification(newTask)
            }
        } catch {
            print("Failed to decode new task: \(error)")
        }
    }
    
    private func handleTaskUpdateNotification(_ notification: Notification) {
        guard let message = notification.object as? [String: Any],
              let taskId = message["taskId"] as? String,
              let statusString = message["status"] as? String,
              let status = TaskStatus(rawValue: statusString) else {
            return
        }
        
        DispatchQueue.main.async {
            self.updateTaskStatus(taskId, status: status)
        }
    }
    
    private func showNewTaskNotification(_ task: WatchTask) {
        let notification = UNMutableNotificationContent()
        notification.title = "新任务"
        notification.body = task.title
        notification.sound = .default
        
        let request = UNNotificationRequest(
            identifier: "newTask_\(task.id)",
            content: notification,
            trigger: nil
        )
        
        UNUserNotificationCenter.current().add(request)
    }
    
    private func showSuccessNotification(_ message: String) {
        // 显示成功消息
        errorMessage = nil
        // 这里可以添加更复杂的通知逻辑
        print("Success: \(message)")
    }
    
    private func showErrorNotification(_ message: String) {
        errorMessage = message
        WKInterfaceDevice.current().play(.failure)
    }
    
    // MARK: - Cache Management
    
    private func saveTasksToCache(_ tasks: [WatchTask]) {
        do {
            let data = try JSONEncoder().encode(tasks)
            UserDefaults.standard.set(data, forKey: "cachedTasks")
        } catch {
            print("Failed to cache tasks: \(error)")
        }
    }
    
    private func loadTasksFromCache() {
        guard let data = UserDefaults.standard.data(forKey: "cachedTasks") else {
            return
        }
        
        do {
            let cachedTasks = try JSONDecoder().decode([WatchTask].self, from: data)
            updateTasks(cachedTasks)
        } catch {
            print("Failed to load cached tasks: \(error)")
        }
    }
    
    private func clearCache() {
        UserDefaults.standard.removeObject(forKey: "cachedTasks")
    }
}

// MARK: - Task Statistics
extension WatchTaskManager {
    var taskStatistics: TaskStatistics {
        let total = tasks.count
        let completed = completedTasks.count
        let active = activeTasks.count
        let overdue = getOverdueTasks().count
        
        return TaskStatistics(
            total: total,
            completed: completed,
            active: active,
            overdue: overdue,
            completionRate: total > 0 ? Double(completed) / Double(total) : 0.0
        )
    }
}

struct TaskStatistics {
    let total: Int
    let completed: Int
    let active: Int
    let overdue: Int
    let completionRate: Double
    
    var completionPercentage: String {
        return String(format: "%.1f%%", completionRate * 100)
    }
}