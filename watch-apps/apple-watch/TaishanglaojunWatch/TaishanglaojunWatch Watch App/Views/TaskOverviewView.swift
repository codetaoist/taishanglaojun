import SwiftUI
import WatchKit

struct TaskOverviewView: View {
    @EnvironmentObject var taskManager: WatchTaskManager
    @EnvironmentObject var connectivityManager: WatchConnectivityManager
    @State private var selectedTab: TaskTab = .active
    @State private var showingTaskDetail = false
    @State private var selectedTask: WatchTask?
    @State private var isRefreshing = false
    
    var body: some View {
        NavigationView {
            VStack(spacing: 0) {
                // 连接状态指示器
                ConnectionStatusBar()
                
                // 任务统计
                TaskStatisticsView()
                
                // 标签选择器
                TaskTabSelector(selectedTab: $selectedTab)
                
                // 任务列表
                TaskListView(
                    tasks: filteredTasks,
                    selectedTask: $selectedTask,
                    showingTaskDetail: $showingTaskDetail
                )
            }
            .navigationTitle("任务")
            .navigationBarTitleDisplayMode(.inline)
            .refreshable {
                await refreshTasks()
            }
            .sheet(isPresented: $showingTaskDetail) {
                if let task = selectedTask {
                    TaskDetailView(task: task)
                }
            }
            .onAppear {
                loadTasksIfNeeded()
            }
        }
    }
    
    private var filteredTasks: [WatchTask] {
        switch selectedTab {
        case .active:
            return taskManager.activeTasks
        case .available:
            return taskManager.getTasksByStatus(.available)
        case .completed:
            return taskManager.completedTasks
        case .overdue:
            return taskManager.getOverdueTasks()
        }
    }
    
    private func loadTasksIfNeeded() {
        if taskManager.tasks.isEmpty || shouldRefresh() {
            taskManager.loadTasks()
        }
    }
    
    private func shouldRefresh() -> Bool {
        guard let lastUpdate = taskManager.lastUpdateTime else { return true }
        return Date().timeIntervalSince(lastUpdate) > 300 // 5分钟
    }
    
    @MainActor
    private func refreshTasks() async {
        isRefreshing = true
        taskManager.refreshTasks()
        
        // 等待刷新完成
        try? await Task.sleep(nanoseconds: 1_000_000_000) // 1秒
        isRefreshing = false
    }
}

// MARK: - 连接状态栏
struct ConnectionStatusBar: View {
    @EnvironmentObject var connectivityManager: WatchConnectivityManager
    
    var body: some View {
        HStack {
            Circle()
                .fill(statusColor)
                .frame(width: 8, height: 8)
            
            Text(connectivityManager.connectionStatus.displayName)
                .font(.caption2)
                .foregroundColor(.secondary)
            
            Spacer()
            
            if let lastSync = connectivityManager.lastSyncTime {
                Text(formatSyncTime(lastSync))
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(Color(.systemGray6))
    }
    
    private var statusColor: Color {
        switch connectivityManager.connectionStatus {
        case .connected:
            return .green
        case .paired:
            return .orange
        case .disconnected:
            return .red
        }
    }
    
    private func formatSyncTime(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

// MARK: - 任务统计视图
struct TaskStatisticsView: View {
    @EnvironmentObject var taskManager: WatchTaskManager
    
    var body: some View {
        let stats = taskManager.taskStatistics
        
        HStack(spacing: 12) {
            StatisticItem(
                title: "总计",
                value: "\(stats.total)",
                color: .blue
            )
            
            StatisticItem(
                title: "进行中",
                value: "\(stats.active)",
                color: .orange
            )
            
            StatisticItem(
                title: "已完成",
                value: "\(stats.completed)",
                color: .green
            )
            
            if stats.overdue > 0 {
                StatisticItem(
                    title: "逾期",
                    value: "\(stats.overdue)",
                    color: .red
                )
            }
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 6)
    }
}

struct StatisticItem: View {
    let title: String
    let value: String
    let color: Color
    
    var body: some View {
        VStack(spacing: 2) {
            Text(value)
                .font(.headline)
                .fontWeight(.semibold)
                .foregroundColor(color)
            
            Text(title)
                .font(.caption2)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}

// MARK: - 标签选择器
struct TaskTabSelector: View {
    @Binding var selectedTab: TaskTab
    
    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(TaskTab.allCases, id: \.self) { tab in
                    TaskTabButton(
                        tab: tab,
                        isSelected: selectedTab == tab
                    ) {
                        selectedTab = tab
                    }
                }
            }
            .padding(.horizontal, 8)
        }
        .padding(.vertical, 4)
    }
}

struct TaskTabButton: View {
    let tab: TaskTab
    let isSelected: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Text(tab.displayName)
                .font(.caption)
                .fontWeight(isSelected ? .semibold : .regular)
                .foregroundColor(isSelected ? .white : .primary)
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(
                    RoundedRectangle(cornerRadius: 16)
                        .fill(isSelected ? Color.accentColor : Color(.systemGray5))
                )
        }
        .buttonStyle(PlainButtonStyle())
    }
}

// MARK: - 任务列表视图
struct TaskListView: View {
    let tasks: [WatchTask]
    @Binding var selectedTask: WatchTask?
    @Binding var showingTaskDetail: Bool
    @EnvironmentObject var taskManager: WatchTaskManager
    
    var body: some View {
        if tasks.isEmpty {
            EmptyTasksView()
        } else {
            List(tasks, id: \.id) { task in
                TaskRowView(task: task) {
                    selectedTask = task
                    showingTaskDetail = true
                }
            }
            .listStyle(PlainListStyle())
        }
    }
}

// MARK: - 任务行视图
struct TaskRowView: View {
    let task: WatchTask
    let onTap: () -> Void
    @EnvironmentObject var taskManager: WatchTaskManager
    
    var body: some View {
        Button(action: onTap) {
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    // 任务标题
                    Text(task.title)
                        .font(.headline)
                        .fontWeight(.medium)
                        .lineLimit(1)
                    
                    Spacer()
                    
                    // 任务状态指示器
                    TaskStatusIndicator(status: task.status)
                }
                
                // 任务描述
                if !task.description.isEmpty {
                    Text(task.description)
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }
                
                HStack {
                    // 优先级和难度
                    HStack(spacing: 4) {
                        PriorityIndicator(priority: task.priority)
                        DifficultyStars(difficulty: task.difficulty)
                    }
                    
                    Spacer()
                    
                    // 奖励
                    if task.reward > 0 {
                        Text(task.formattedReward)
                            .font(.caption)
                            .fontWeight(.medium)
                            .foregroundColor(.green)
                    }
                }
                
                // 进度条（仅对进行中的任务显示）
                if task.status == .inProgress {
                    ProgressView(value: task.progress)
                        .progressViewStyle(LinearProgressViewStyle())
                        .scaleEffect(y: 0.5)
                }
                
                // 坐标指示器
                CoordinateIndicator(coordinate: task.coordinate)
            }
            .padding(.vertical, 4)
        }
        .buttonStyle(PlainButtonStyle())
        .contextMenu {
            TaskContextMenu(task: task)
        }
    }
}

// MARK: - 任务状态指示器
struct TaskStatusIndicator: View {
    let status: TaskStatus
    
    var body: some View {
        Circle()
            .fill(statusColor)
            .frame(width: 8, height: 8)
            .overlay(
                Circle()
                    .stroke(Color.white, lineWidth: 1)
            )
    }
    
    private var statusColor: Color {
        switch status {
        case .available:
            return .blue
        case .accepted:
            return .orange
        case .inProgress:
            return .yellow
        case .completed:
            return .green
        case .cancelled:
            return .red
        }
    }
}

// MARK: - 优先级指示器
struct PriorityIndicator: View {
    let priority: TaskPriority
    
    var body: some View {
        Text(priority.symbol)
            .font(.caption2)
            .fontWeight(.bold)
            .foregroundColor(priorityColor)
    }
    
    private var priorityColor: Color {
        switch priority {
        case .low:
            return .green
        case .medium:
            return .orange
        case .high:
            return .red
        case .urgent:
            return .purple
        }
    }
}

// MARK: - 难度星级
struct DifficultyStars: View {
    let difficulty: Int
    
    var body: some View {
        HStack(spacing: 1) {
            ForEach(0..<5, id: \.self) { index in
                Image(systemName: index < difficulty ? "star.fill" : "star")
                    .font(.caption2)
                    .foregroundColor(.yellow)
            }
        }
    }
}

// MARK: - 坐标指示器
struct CoordinateIndicator: View {
    let coordinate: TaskCoordinate
    
    var body: some View {
        HStack(spacing: 8) {
            CoordinateAxis(label: "S", value: coordinate.s, color: .red)
            CoordinateAxis(label: "C", value: coordinate.c, color: .green)
            CoordinateAxis(label: "T", value: coordinate.t, color: .blue)
        }
    }
}

struct CoordinateAxis: View {
    let label: String
    let value: Double
    let color: Color
    
    var body: some View {
        VStack(spacing: 1) {
            Text(label)
                .font(.caption2)
                .fontWeight(.bold)
                .foregroundColor(color)
            
            Text(String(format: "%.1f", value))
                .font(.caption2)
                .foregroundColor(.secondary)
        }
    }
}

// MARK: - 任务上下文菜单
struct TaskContextMenu: View {
    let task: WatchTask
    @EnvironmentObject var taskManager: WatchTaskManager
    
    var body: some View {
        Group {
            if task.status == .available {
                Button("接受任务") {
                    taskManager.acceptTask(task)
                }
            }
            
            if task.status == .accepted {
                Button("开始任务") {
                    taskManager.startTask(task)
                }
            }
            
            if task.status == .inProgress {
                Button("完成任务") {
                    taskManager.completeTask(task)
                }
            }
            
            Button("查看详情") {
                // 触发详情视图
            }
        }
    }
}

// MARK: - 空任务视图
struct EmptyTasksView: View {
    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: "checkmark.circle")
                .font(.largeTitle)
                .foregroundColor(.secondary)
            
            Text("暂无任务")
                .font(.headline)
                .foregroundColor(.secondary)
            
            Text("下拉刷新获取最新任务")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - 支持枚举
enum TaskTab: CaseIterable {
    case active
    case available
    case completed
    case overdue
    
    var displayName: String {
        switch self {
        case .active:
            return "进行中"
        case .available:
            return "可接受"
        case .completed:
            return "已完成"
        case .overdue:
            return "逾期"
        }
    }
}

// MARK: - 优先级扩展
extension TaskPriority {
    var symbol: String {
        switch self {
        case .low:
            return "↓"
        case .medium:
            return "→"
        case .high:
            return "↑"
        case .urgent:
            return "‼️"
        }
    }
}

#Preview {
    TaskOverviewView()
        .environmentObject(WatchTaskManager.shared)
        .environmentObject(WatchConnectivityManager.shared)
}