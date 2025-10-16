import Foundation
import SwiftUI

// MARK: - 手表任务模型
struct WatchTask: Identifiable, Codable, Hashable {
    let id: String
    let title: String
    let description: String
    let status: TaskStatus
    let priority: TaskPriority
    let difficulty: Int // 1-5
    let reward: Double
    let coordinate: TaskCoordinate
    let createdAt: Date
    let updatedAt: Date
    let dueDate: Date?
    let estimatedDuration: TimeInterval
    let tags: [String]
    
    // 手表端特有属性
    let isQuickActionAvailable: Bool
    let requiresPhoneConnection: Bool
    let canCompleteOffline: Bool
    
    init(id: String = UUID().uuidString,
         title: String,
         description: String,
         status: TaskStatus = .pending,
         priority: TaskPriority = .medium,
         difficulty: Int = 1,
         reward: Double = 0.0,
         coordinate: TaskCoordinate = TaskCoordinate(),
         createdAt: Date = Date(),
         updatedAt: Date = Date(),
         dueDate: Date? = nil,
         estimatedDuration: TimeInterval = 3600,
         tags: [String] = [],
         isQuickActionAvailable: Bool = true,
         requiresPhoneConnection: Bool = false,
         canCompleteOffline: Bool = true) {
        
        self.id = id
        self.title = title
        self.description = description
        self.status = status
        self.priority = priority
        self.difficulty = difficulty
        self.reward = reward
        self.coordinate = coordinate
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.dueDate = dueDate
        self.estimatedDuration = estimatedDuration
        self.tags = tags
        self.isQuickActionAvailable = isQuickActionAvailable
        self.requiresPhoneConnection = requiresPhoneConnection
        self.canCompleteOffline = canCompleteOffline
    }
}

// MARK: - 任务状态
enum TaskStatus: String, Codable, CaseIterable {
    case pending = "pending"
    case inProgress = "in_progress"
    case completed = "completed"
    case cancelled = "cancelled"
    case paused = "paused"
    
    var displayName: String {
        switch self {
        case .pending: return "待处理"
        case .inProgress: return "进行中"
        case .completed: return "已完成"
        case .cancelled: return "已取消"
        case .paused: return "已暂停"
        }
    }
    
    var color: Color {
        switch self {
        case .pending: return .orange
        case .inProgress: return .blue
        case .completed: return .green
        case .cancelled: return .red
        case .paused: return .yellow
        }
    }
    
    var systemImage: String {
        switch self {
        case .pending: return "clock"
        case .inProgress: return "play.circle"
        case .completed: return "checkmark.circle"
        case .cancelled: return "xmark.circle"
        case .paused: return "pause.circle"
        }
    }
}

// MARK: - 任务优先级
enum TaskPriority: String, Codable, CaseIterable {
    case low = "low"
    case medium = "medium"
    case high = "high"
    case urgent = "urgent"
    
    var displayName: String {
        switch self {
        case .low: return "低"
        case .medium: return "中"
        case .high: return "高"
        case .urgent: return "紧急"
        }
    }
    
    var color: Color {
        switch self {
        case .low: return .gray
        case .medium: return .blue
        case .high: return .orange
        case .urgent: return .red
        }
    }
    
    var sortOrder: Int {
        switch self {
        case .low: return 1
        case .medium: return 2
        case .high: return 3
        case .urgent: return 4
        }
    }
}

// MARK: - 三轴坐标
struct TaskCoordinate: Codable, Hashable {
    let s: Int // 硅基轴 (1-5)
    let c: Int // 碳基轴 (1-5)
    let t: Int // 时间轴 (1-5)
    
    init(s: Int = 1, c: Int = 1, t: Int = 1) {
        self.s = max(1, min(5, s))
        self.c = max(1, min(5, c))
        self.t = max(1, min(5, t))
    }
    
    var cLayerOpacity: Double {
        return Double(c) / 5.0
    }
    
    var displayString: String {
        return "S\(s)C\(c)T\(t)"
    }
}

// MARK: - 手表通知
struct WatchNotification: Identifiable, Codable {
    let id: String
    let title: String
    let message: String
    let type: NotificationType
    let createdAt: Date
    let isRead: Bool
    let actionRequired: Bool
    let relatedTaskId: String?
    
    init(id: String = UUID().uuidString,
         title: String,
         message: String,
         type: NotificationType = .info,
         createdAt: Date = Date(),
         isRead: Bool = false,
         actionRequired: Bool = false,
         relatedTaskId: String? = nil) {
        
        self.id = id
        self.title = title
        self.message = message
        self.type = type
        self.createdAt = createdAt
        self.isRead = isRead
        self.actionRequired = actionRequired
        self.relatedTaskId = relatedTaskId
    }
}

// MARK: - 通知类型
enum NotificationType: String, Codable, CaseIterable {
    case info = "info"
    case warning = "warning"
    case error = "error"
    case success = "success"
    case taskUpdate = "task_update"
    case reminder = "reminder"
    
    var color: Color {
        switch self {
        case .info: return .blue
        case .warning: return .orange
        case .error: return .red
        case .success: return .green
        case .taskUpdate: return .purple
        case .reminder: return .yellow
        }
    }
    
    var systemImage: String {
        switch self {
        case .info: return "info.circle"
        case .warning: return "exclamationmark.triangle"
        case .error: return "xmark.circle"
        case .success: return "checkmark.circle"
        case .taskUpdate: return "arrow.clockwise"
        case .reminder: return "bell"
        }
    }
}

// MARK: - 扩展方法
extension WatchTask {
    var isOverdue: Bool {
        guard let dueDate = dueDate else { return false }
        return Date() > dueDate && status != .completed
    }
    
    var timeRemaining: TimeInterval? {
        guard let dueDate = dueDate else { return nil }
        let remaining = dueDate.timeIntervalSince(Date())
        return remaining > 0 ? remaining : 0
    }
    
    var formattedReward: String {
        return String(format: "¥%.2f", reward)
    }
    
    var difficultyStars: String {
        return String(repeating: "★", count: difficulty) + String(repeating: "☆", count: 5 - difficulty)
    }
}