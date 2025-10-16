import Foundation
import Combine
import WatchKit

@MainActor
class WatchTaskManager: ObservableObject {
    // MARK: - Published Properties
    @Published var tasks: [WatchTask] = []
    @Published var recentTasks: [WatchTask] = []
    @Published var taskStatistics = TaskStatistics()
    @Published var isLoading = false
    @Published var error: String?
    @Published var lastUpdateTime: Date?
    
    // MARK: - Private Properties
    private var cancellables = Set<AnyCancellable>()
    private let userDefaults = UserDefaults.standard
    private let tasksKey = "cached_tasks"
    private let lastUpdateKey = "last_task_update"
    
    // MARK: - Dependencies
    private let connectivityManager: WatchConnectivityManager
    private let settingsManager: WatchSettingsManager
    
    // MARK: - Initialization
    init(connectivityManager: WatchConnectivityManager, settingsManager: WatchSettingsManager) {
        self.connectivityManager = connectivityManager
        self.settingsManager = settingsManager
        
        setupObservers()
        loadCachedTasks()
    }
    
    // MARK: - Setup
    private func setupObservers() {
        // Observe connectivity changes
        connectivityManager.$isConnected
            .sink { [weak self] isConnected in
                if isConnected {
                    Task {
                        await self?.syncTasks()
                    }
                }
            }
            .store(in: &cancellables)
        
        // Observe received tasks from phone
        connectivityManager.$receivedTasks
            .compactMap { $0 }
            .sink { [weak self] tasks in
                self?.updateTasks(tasks)
            }
            .store(in: &cancellables)
    }
    
    // MARK: - Task Management
    func loadTasks() async {
        isLoading = true
        error = nil
        
        do {
            // Try to sync with phone first
            if connectivityManager.isConnected {
                await syncTasks()
            } else {
                // Load from cache
                loadCachedTasks()
            }
        } catch {
            self.error = "加载任务失败: \(error.localizedDescription)"
        }
        
        isLoading = false
    }
    
    func syncTasks() async {
        guard connectivityManager.isConnected else {
            error = "设备未连接，无法同步"
            return
        }
        
        isLoading = true
        error = nil
        
        do {
            await connectivityManager.requestTasks()
            lastUpdateTime = Date()
            userDefaults.set(lastUpdateTime, forKey: lastUpdateKey)
        } catch {
            self.error = "同步失败: \(error.localizedDescription)"
        }
        
        isLoading = false
    }
    
    func acceptTask(_ task: WatchTask) async {
        guard connectivityManager.isConnected else {
            error = "设备未连接，无法接受任务"
            return
        }
        
        do {
            await connectivityManager.acceptTask(task.id)
            
            // Update local task status
            if let index = tasks.firstIndex(where: { $0.id == task.id }) {
                tasks[index].status = .accepted
                cacheTasks()
                updateStatistics()
            }
        } catch {
            self.error = "接受任务失败: \(error.localizedDescription)"
        }
    }
    
    func startTask(_ task: WatchTask) async {
        do {
            // Update local status first
            if let index = tasks.firstIndex(where: { $0.id == task.id }) {
                tasks[index].status = .inProgress
                tasks[index].startedAt = Date()
                cacheTasks()
                updateStatistics()
            }
            
            // Notify phone if connected
            if connectivityManager.isConnected {
                await connectivityManager.updateTaskProgress(task.id, progress: 0, note: "任务已开始")
            }
        } catch {
            self.error = "开始任务失败: \(error.localizedDescription)"
        }
    }
    
    func completeTask(_ task: WatchTask) async {
        do {
            // Update local status
            if let index = tasks.firstIndex(where: { $0.id == task.id }) {
                tasks[index].status = .completed
                tasks[index].completedAt = Date()
                tasks[index].progress = 100
                cacheTasks()
                updateStatistics()
                
                // Update settings
                settingsManager.incrementCompletedTasks()
            }
            
            // Notify phone if connected
            if connectivityManager.isConnected {
                await connectivityManager.updateTaskProgress(task.id, progress: 100, note: "任务已完成")
            }
            
            // Trigger haptic feedback
            settingsManager.triggerHapticFeedback(.success)
            
        } catch {
            self.error = "完成任务失败: \(error.localizedDescription)"
        }
    }
    
    func updateTaskProgress(_ task: WatchTask, progress: Int, note: String? = nil) async {
        do {
            // Update local progress
            if let index = tasks.firstIndex(where: { $0.id == task.id }) {
                tasks[index].progress = progress
                tasks[index].lastUpdated = Date()
                cacheTasks()
            }
            
            // Notify phone if connected
            if connectivityManager.isConnected {
                await connectivityManager.updateTaskProgress(task.id, progress: progress, note: note ?? "进度更新")
            }
        } catch {
            self.error = "更新进度失败: \(error.localizedDescription)"
        }
    }
    
    // MARK: - Task Queries
    func getTask(by id: String) -> WatchTask? {
        return tasks.first { $0.id == id }
    }
    
    func getTasks(by status: TaskStatus) -> [WatchTask] {
        return tasks.filter { $0.status == status }
    }
    
    func getTasks(by priority: TaskPriority) -> [WatchTask] {
        return tasks.filter { $0.priority == priority }
    }
    
    func getQuickActionTasks() -> [WatchTask] {
        return tasks.filter { $0.isQuickActionAvailable }
    }
    
    func searchTasks(query: String) -> [WatchTask] {
        guard !query.isEmpty else { return tasks }
        
        return tasks.filter { task in
            task.title.localizedCaseInsensitiveContains(query) ||
            task.description.localizedCaseInsensitiveContains(query)
        }
    }
    
    // MARK: - Private Methods
    private func updateTasks(_ newTasks: [WatchTask]) {
        tasks = newTasks
        updateRecentTasks()
        updateStatistics()
        cacheTasks()
        lastUpdateTime = Date()
        userDefaults.set(lastUpdateTime, forKey: lastUpdateKey)
    }
    
    private func updateRecentTasks() {
        recentTasks = Array(tasks
            .sorted { $0.lastUpdated > $1.lastUpdated }
            .prefix(5))
    }
    
    private func updateStatistics() {
        let total = tasks.count
        let pending = tasks.filter { $0.status == .pending }.count
        let inProgress = tasks.filter { $0.status == .inProgress }.count
        let completed = tasks.filter { $0.status == .completed }.count
        let overdue = tasks.filter { $0.isOverdue }.count
        
        taskStatistics = TaskStatistics(
            total: total,
            pending: pending,
            inProgress: inProgress,
            completed: completed,
            overdue: overdue
        )
    }
    
    private func cacheTasks() {
        do {
            let data = try JSONEncoder().encode(tasks)
            userDefaults.set(data, forKey: tasksKey)
        } catch {
            print("Failed to cache tasks: \(error)")
        }
    }
    
    private func loadCachedTasks() {
        guard let data = userDefaults.data(forKey: tasksKey) else { return }
        
        do {
            tasks = try JSONDecoder().decode([WatchTask].self, from: data)
            updateRecentTasks()
            updateStatistics()
            lastUpdateTime = userDefaults.object(forKey: lastUpdateKey) as? Date
        } catch {
            print("Failed to load cached tasks: \(error)")
        }
    }
    
    // MARK: - Public Utilities
    func clearCache() {
        tasks.removeAll()
        recentTasks.removeAll()
        taskStatistics = TaskStatistics()
        userDefaults.removeObject(forKey: tasksKey)
        userDefaults.removeObject(forKey: lastUpdateKey)
        lastUpdateTime = nil
    }
    
    func refreshTasks() async {
        await loadTasks()
    }
    
    func clearError() {
        error = nil
    }
}

// MARK: - Supporting Types
struct TaskStatistics: Codable {
    let total: Int
    let pending: Int
    let inProgress: Int
    let completed: Int
    let overdue: Int
    
    init(total: Int = 0, pending: Int = 0, inProgress: Int = 0, completed: Int = 0, overdue: Int = 0) {
        self.total = total
        self.pending = pending
        self.inProgress = inProgress
        self.completed = completed
        self.overdue = overdue
    }
    
    var completionRate: Double {
        guard total > 0 else { return 0 }
        return Double(completed) / Double(total)
    }
    
    var activeTasksCount: Int {
        return pending + inProgress
    }
}