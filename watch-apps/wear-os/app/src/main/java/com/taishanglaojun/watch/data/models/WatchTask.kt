package com.taishanglaojun.watch.data.models

import android.os.Parcelable
import androidx.room.ColumnInfo
import androidx.room.Embedded
import androidx.room.Entity
import androidx.room.PrimaryKey
import com.google.gson.annotations.SerializedName
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import kotlinx.parcelize.Parcelize
import kotlin.time.Duration.Companion.hours

/**
 * 手表任务数据模型
 */
@Entity(tableName = "watch_tasks")
@Parcelize
data class WatchTask(
    @PrimaryKey
    @SerializedName("id")
    val id: String,
    
    @ColumnInfo(name = "title")
    @SerializedName("title")
    val title: String,
    
    @ColumnInfo(name = "description")
    @SerializedName("description")
    val description: String,
    
    @ColumnInfo(name = "status")
    @SerializedName("status")
    val status: TaskStatus,
    
    @ColumnInfo(name = "priority")
    @SerializedName("priority")
    val priority: TaskPriority,
    
    @ColumnInfo(name = "difficulty")
    @SerializedName("difficulty")
    val difficulty: Int, // 1-5星难度
    
    @ColumnInfo(name = "reward")
    @SerializedName("reward")
    val reward: Double,
    
    @Embedded
    @SerializedName("coordinate")
    val coordinate: TaskCoordinate,
    
    @ColumnInfo(name = "progress")
    @SerializedName("progress")
    val progress: Double = 0.0,
    
    @ColumnInfo(name = "created_at")
    @SerializedName("created_at")
    val createdAt: Instant = Clock.System.now(),
    
    @ColumnInfo(name = "updated_at")
    @SerializedName("updated_at")
    val updatedAt: Instant = Clock.System.now(),
    
    @ColumnInfo(name = "due_date")
    @SerializedName("due_date")
    val dueDate: Instant? = null,
    
    @ColumnInfo(name = "estimated_duration")
    @SerializedName("estimated_duration")
    val estimatedDuration: Long? = null, // 预估时长（秒）
    
    // 手表特定属性
    @ColumnInfo(name = "is_quick_action_available")
    @SerializedName("is_quick_action_available")
    val isQuickActionAvailable: Boolean = false,
    
    @ColumnInfo(name = "requires_phone_connection")
    @SerializedName("requires_phone_connection")
    val requiresPhoneConnection: Boolean = true,
    
    @ColumnInfo(name = "can_complete_offline")
    @SerializedName("can_complete_offline")
    val canCompleteOffline: Boolean = false,
    
    @ColumnInfo(name = "location_required")
    @SerializedName("location_required")
    val locationRequired: Boolean = false,
    
    @ColumnInfo(name = "health_data_required")
    @SerializedName("health_data_required")
    val healthDataRequired: Boolean = false,
    
    // 缓存和同步相关
    @ColumnInfo(name = "is_cached")
    val isCached: Boolean = false,
    
    @ColumnInfo(name = "last_sync_time")
    val lastSyncTime: Instant? = null,
    
    @ColumnInfo(name = "sync_status")
    val syncStatus: SyncStatus = SyncStatus.PENDING
) : Parcelable {
    
    /**
     * 检查任务是否逾期
     */
    val isOverdue: Boolean
        get() = dueDate?.let { it < Clock.System.now() } ?: false
    
    /**
     * 获取剩余时间（秒）
     */
    val timeRemaining: Long?
        get() = dueDate?.let { 
            (it - Clock.System.now()).inWholeSeconds.coerceAtLeast(0)
        }
    
    /**
     * 格式化奖励显示
     */
    val formattedReward: String
        get() = when {
            reward >= 1000 -> String.format("%.1fK积分", reward / 1000)
            reward >= 100 -> String.format("%.0f积分", reward)
            else -> String.format("%.1f积分", reward)
        }
    
    /**
     * 获取难度星级字符串
     */
    val difficultyStars: String
        get() = "★".repeat(difficulty) + "☆".repeat(5 - difficulty)
    
    /**
     * 检查是否可以在手表上执行
     */
    val canExecuteOnWatch: Boolean
        get() = isQuickActionAvailable && (canCompleteOffline || !requiresPhoneConnection)
    
    /**
     * 获取进度百分比
     */
    val progressPercentage: Int
        get() = (progress * 100).toInt().coerceIn(0, 100)
    
    /**
     * 检查是否需要位置权限
     */
    val needsLocationPermission: Boolean
        get() = locationRequired
    
    /**
     * 检查是否需要健康数据权限
     */
    val needsHealthPermission: Boolean
        get() = healthDataRequired
}

/**
 * 任务状态枚举
 */
enum class TaskStatus(val displayName: String, val colorCode: String) {
    @SerializedName("available")
    AVAILABLE("可接受", "#2196F3"),
    
    @SerializedName("accepted")
    ACCEPTED("已接受", "#FF9800"),
    
    @SerializedName("in_progress")
    IN_PROGRESS("进行中", "#FFC107"),
    
    @SerializedName("completed")
    COMPLETED("已完成", "#4CAF50"),
    
    @SerializedName("cancelled")
    CANCELLED("已取消", "#F44336"),
    
    @SerializedName("paused")
    PAUSED("已暂停", "#9E9E9E");
    
    companion object {
        fun fromString(value: String): TaskStatus {
            return values().find { it.name.equals(value, ignoreCase = true) } ?: AVAILABLE
        }
    }
}

/**
 * 任务优先级枚举
 */
enum class TaskPriority(val level: Int, val displayName: String, val colorCode: String) {
    @SerializedName("low")
    LOW(1, "低", "#4CAF50"),
    
    @SerializedName("medium")
    MEDIUM(2, "中", "#FF9800"),
    
    @SerializedName("high")
    HIGH(3, "高", "#F44336"),
    
    @SerializedName("urgent")
    URGENT(4, "紧急", "#9C27B0");
    
    companion object {
        fun fromString(value: String): TaskPriority {
            return values().find { it.name.equals(value, ignoreCase = true) } ?: MEDIUM
        }
    }
}

/**
 * 任务坐标系统
 */
@Parcelize
data class TaskCoordinate(
    @ColumnInfo(name = "coordinate_s")
    @SerializedName("s")
    val s: Double, // 空间坐标
    
    @ColumnInfo(name = "coordinate_c")
    @SerializedName("c")
    val c: Double, // 认知坐标
    
    @ColumnInfo(name = "coordinate_t")
    @SerializedName("t")
    val t: Double  // 时间坐标
) : Parcelable {
    
    /**
     * 计算坐标向量的模长
     */
    val magnitude: Double
        get() = kotlin.math.sqrt(s * s + c * c + t * t)
    
    /**
     * 获取主导维度
     */
    val dominantDimension: String
        get() = when {
            s >= c && s >= t -> "S"
            c >= s && c >= t -> "C"
            else -> "T"
        }
    
    /**
     * 格式化坐标显示
     */
    fun formatCoordinate(): String {
        return "S:${String.format("%.1f", s)} C:${String.format("%.1f", c)} T:${String.format("%.1f", t)}"
    }
}

/**
 * 同步状态枚举
 */
enum class SyncStatus {
    PENDING,    // 待同步
    SYNCING,    // 同步中
    SYNCED,     // 已同步
    FAILED,     // 同步失败
    CONFLICT    // 同步冲突
}

/**
 * 通知类型枚举
 */
enum class NotificationType(val displayName: String) {
    NEW_TASK("新任务"),
    TASK_UPDATE("任务更新"),
    TASK_REMINDER("任务提醒"),
    SYNC_COMPLETE("同步完成"),
    ERROR("错误"),
    SUCCESS("成功");
    
    companion object {
        fun fromString(value: String): NotificationType {
            return values().find { it.name.equals(value, ignoreCase = true) } ?: NEW_TASK
        }
    }
}

/**
 * 任务统计数据
 */
@Parcelize
data class TaskStatistics(
    val total: Int = 0,
    val available: Int = 0,
    val accepted: Int = 0,
    val inProgress: Int = 0,
    val completed: Int = 0,
    val overdue: Int = 0,
    val totalReward: Double = 0.0,
    val completionRate: Double = 0.0
) : Parcelable {
    
    val completionPercentage: String
        get() = String.format("%.1f%%", completionRate * 100)
    
    val formattedTotalReward: String
        get() = when {
            totalReward >= 1000 -> String.format("%.1fK", totalReward / 1000)
            else -> String.format("%.0f", totalReward)
        }
}

/**
 * 任务过滤器
 */
data class TaskFilter(
    val status: TaskStatus? = null,
    val priority: TaskPriority? = null,
    val minDifficulty: Int? = null,
    val maxDifficulty: Int? = null,
    val requiresLocation: Boolean? = null,
    val requiresHealth: Boolean? = null,
    val canCompleteOffline: Boolean? = null,
    val isQuickAction: Boolean? = null,
    val searchQuery: String? = null
)

/**
 * 任务排序选项
 */
enum class TaskSortOption(val displayName: String) {
    CREATED_DATE_DESC("创建时间（新到旧）"),
    CREATED_DATE_ASC("创建时间（旧到新）"),
    PRIORITY_DESC("优先级（高到低）"),
    PRIORITY_ASC("优先级（低到高）"),
    DIFFICULTY_DESC("难度（高到低）"),
    DIFFICULTY_ASC("难度（低到高）"),
    REWARD_DESC("奖励（高到低）"),
    REWARD_ASC("奖励（低到高）"),
    DUE_DATE_ASC("截止时间（近到远）"),
    PROGRESS_DESC("进度（高到低）")
}

/**
 * 任务操作结果
 */
sealed class TaskOperationResult {
    object Success : TaskOperationResult()
    data class Error(val message: String, val throwable: Throwable? = null) : TaskOperationResult()
    object Loading : TaskOperationResult()
}

/**
 * 任务扩展函数
 */
fun WatchTask.canBeAccepted(): Boolean {
    return status == TaskStatus.AVAILABLE
}

fun WatchTask.canBeStarted(): Boolean {
    return status == TaskStatus.ACCEPTED
}

fun WatchTask.canBeCompleted(): Boolean {
    return status == TaskStatus.IN_PROGRESS && progress >= 0.95
}

fun WatchTask.canBePaused(): Boolean {
    return status == TaskStatus.IN_PROGRESS
}

fun WatchTask.canBeResumed(): Boolean {
    return status == TaskStatus.PAUSED
}

fun WatchTask.canBeCancelled(): Boolean {
    return status in listOf(TaskStatus.ACCEPTED, TaskStatus.IN_PROGRESS, TaskStatus.PAUSED)
}

/**
 * 任务列表扩展函数
 */
fun List<WatchTask>.filterByStatus(status: TaskStatus): List<WatchTask> {
    return filter { it.status == status }
}

fun List<WatchTask>.filterByPriority(priority: TaskPriority): List<WatchTask> {
    return filter { it.priority == priority }
}

fun List<WatchTask>.filterOverdue(): List<WatchTask> {
    return filter { it.isOverdue }
}

fun List<WatchTask>.filterQuickActions(): List<WatchTask> {
    return filter { it.isQuickActionAvailable }
}

fun List<WatchTask>.sortByPriority(): List<WatchTask> {
    return sortedByDescending { it.priority.level }
}

fun List<WatchTask>.sortByDueDate(): List<WatchTask> {
    return sortedBy { it.dueDate }
}

fun List<WatchTask>.calculateStatistics(): TaskStatistics {
    val total = size
    val available = count { it.status == TaskStatus.AVAILABLE }
    val accepted = count { it.status == TaskStatus.ACCEPTED }
    val inProgress = count { it.status == TaskStatus.IN_PROGRESS }
    val completed = count { it.status == TaskStatus.COMPLETED }
    val overdue = count { it.isOverdue }
    val totalReward = filter { it.status == TaskStatus.COMPLETED }.sumOf { it.reward }
    val completionRate = if (total > 0) completed.toDouble() / total else 0.0
    
    return TaskStatistics(
        total = total,
        available = available,
        accepted = accepted,
        inProgress = inProgress,
        completed = completed,
        overdue = overdue,
        totalReward = totalReward,
        completionRate = completionRate
    )
}