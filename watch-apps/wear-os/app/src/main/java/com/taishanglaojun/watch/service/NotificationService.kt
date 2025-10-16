package com.taishanglaojun.watch.service

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import com.taishanglaojun.watch.R
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.data.models.TaskStatus
import com.taishanglaojun.watch.data.models.TaskPriority
import com.taishanglaojun.watch.ui.MainActivity
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class NotificationService @Inject constructor(
    @ApplicationContext private val context: Context
) {
    
    private val notificationManager = NotificationManagerCompat.from(context)
    
    init {
        createNotificationChannels()
    }
    
    private fun createNotificationChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channels = listOf(
                NotificationChannel(
                    CHANNEL_TASK_UPDATES,
                    "任务更新",
                    NotificationManager.IMPORTANCE_DEFAULT
                ).apply {
                    description = "任务状态更新和提醒"
                    enableVibration(true)
                },
                
                NotificationChannel(
                    CHANNEL_LOCATION_ALERTS,
                    "位置提醒",
                    NotificationManager.IMPORTANCE_HIGH
                ).apply {
                    description = "到达任务位置时的提醒"
                    enableVibration(true)
                },
                
                NotificationChannel(
                    CHANNEL_SYNC_STATUS,
                    "同步状态",
                    NotificationManager.IMPORTANCE_LOW
                ).apply {
                    description = "数据同步状态通知"
                    enableVibration(false)
                },
                
                NotificationChannel(
                    CHANNEL_SYSTEM,
                    "系统通知",
                    NotificationManager.IMPORTANCE_DEFAULT
                ).apply {
                    description = "系统相关通知"
                    enableVibration(true)
                }
            )
            
            channels.forEach { channel ->
                notificationManager.createNotificationChannel(channel)
            }
        }
    }
    
    // Task-related notifications
    fun showTaskAcceptedNotification(task: WatchTask) {
        val notification = createTaskNotification(
            task = task,
            title = "任务已接受",
            content = "任务「${task.title}」已成功接受",
            channelId = CHANNEL_TASK_UPDATES,
            notificationId = NOTIFICATION_TASK_ACCEPTED + task.id.hashCode()
        )
        
        notificationManager.notify(NOTIFICATION_TASK_ACCEPTED + task.id.hashCode(), notification)
    }
    
    fun showTaskStartedNotification(task: WatchTask) {
        val notification = createTaskNotification(
            task = task,
            title = "任务已开始",
            content = "任务「${task.title}」已开始执行",
            channelId = CHANNEL_TASK_UPDATES,
            notificationId = NOTIFICATION_TASK_STARTED + task.id.hashCode()
        )
        
        notificationManager.notify(NOTIFICATION_TASK_STARTED + task.id.hashCode(), notification)
    }
    
    fun showTaskCompletedNotification(task: WatchTask) {
        val notification = createTaskNotification(
            task = task,
            title = "任务已完成",
            content = "恭喜！任务「${task.title}」已完成",
            channelId = CHANNEL_TASK_UPDATES,
            notificationId = NOTIFICATION_TASK_COMPLETED + task.id.hashCode()
        )
        
        notificationManager.notify(NOTIFICATION_TASK_COMPLETED + task.id.hashCode(), notification)
    }
    
    fun showTaskReminderNotification(task: WatchTask) {
        val notification = createTaskNotification(
            task = task,
            title = "任务提醒",
            content = "任务「${task.title}」需要您的关注",
            channelId = CHANNEL_TASK_UPDATES,
            notificationId = NOTIFICATION_TASK_REMINDER + task.id.hashCode()
        )
        
        notificationManager.notify(NOTIFICATION_TASK_REMINDER + task.id.hashCode(), notification)
    }
    
    // Location-based notifications
    fun showLocationArrivalNotification(task: WatchTask, distance: Double) {
        val distanceText = if (distance < 1000) {
            "${distance.toInt()}米"
        } else {
            "${(distance / 1000).toInt()}公里"
        }
        
        val notification = createTaskNotification(
            task = task,
            title = "已到达任务位置",
            content = "您距离任务「${task.title}」仅${distanceText}",
            channelId = CHANNEL_LOCATION_ALERTS,
            notificationId = NOTIFICATION_LOCATION_ARRIVAL + task.id.hashCode()
        )
        
        notificationManager.notify(NOTIFICATION_LOCATION_ARRIVAL + task.id.hashCode(), notification)
    }
    
    // Sync notifications
    fun showSyncStartedNotification() {
        val notification = NotificationCompat.Builder(context, CHANNEL_SYNC_STATUS)
            .setSmallIcon(R.drawable.ic_sync)
            .setContentTitle("正在同步")
            .setContentText("正在与手机同步任务数据...")
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setOngoing(true)
            .build()
        
        notificationManager.notify(NOTIFICATION_SYNC_PROGRESS, notification)
    }
    
    fun showSyncCompletedNotification(taskCount: Int) {
        val notification = NotificationCompat.Builder(context, CHANNEL_SYNC_STATUS)
            .setSmallIcon(R.drawable.ic_sync_done)
            .setContentTitle("同步完成")
            .setContentText("已同步 $taskCount 个任务")
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setAutoCancel(true)
            .build()
        
        notificationManager.notify(NOTIFICATION_SYNC_COMPLETED, notification)
    }
    
    fun showSyncErrorNotification(error: String) {
        val notification = NotificationCompat.Builder(context, CHANNEL_SYNC_STATUS)
            .setSmallIcon(R.drawable.ic_sync_error)
            .setContentTitle("同步失败")
            .setContentText("同步失败: $error")
            .setPriority(NotificationCompat.PRIORITY_DEFAULT)
            .setAutoCancel(true)
            .build()
        
        notificationManager.notify(NOTIFICATION_SYNC_ERROR, notification)
    }
    
    // System notifications
    fun showConnectionLostNotification() {
        val notification = NotificationCompat.Builder(context, CHANNEL_SYSTEM)
            .setSmallIcon(R.drawable.ic_connection_lost)
            .setContentTitle("连接断开")
            .setContentText("与手机的连接已断开")
            .setPriority(NotificationCompat.PRIORITY_DEFAULT)
            .setAutoCancel(true)
            .build()
        
        notificationManager.notify(NOTIFICATION_CONNECTION_LOST, notification)
    }
    
    fun showConnectionRestoredNotification() {
        val notification = NotificationCompat.Builder(context, CHANNEL_SYSTEM)
            .setSmallIcon(R.drawable.ic_connection_restored)
            .setContentTitle("连接已恢复")
            .setContentText("与手机的连接已恢复")
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setAutoCancel(true)
            .build()
        
        notificationManager.notify(NOTIFICATION_CONNECTION_RESTORED, notification)
    }
    
    // Utility methods
    private fun createTaskNotification(
        task: WatchTask,
        title: String,
        content: String,
        channelId: String,
        notificationId: Int
    ): Notification {
        val intent = Intent(context, MainActivity::class.java).apply {
            putExtra("task_id", task.id)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        val pendingIntent = PendingIntent.getActivity(
            context,
            notificationId,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        val priority = when (task.priority) {
            TaskPriority.HIGH -> NotificationCompat.PRIORITY_HIGH
            TaskPriority.MEDIUM -> NotificationCompat.PRIORITY_DEFAULT
            TaskPriority.LOW -> NotificationCompat.PRIORITY_LOW
        }
        
        return NotificationCompat.Builder(context, channelId)
            .setSmallIcon(getTaskStatusIcon(task.status))
            .setContentTitle(title)
            .setContentText(content)
            .setPriority(priority)
            .setContentIntent(pendingIntent)
            .setAutoCancel(true)
            .addTaskActions(task)
            .build()
    }
    
    private fun NotificationCompat.Builder.addTaskActions(task: WatchTask): NotificationCompat.Builder {
        when (task.status) {
            TaskStatus.PENDING -> {
                // Add accept action
                val acceptIntent = createTaskActionIntent(task.id, ACTION_ACCEPT_TASK)
                val acceptPendingIntent = PendingIntent.getBroadcast(
                    context,
                    task.id.hashCode(),
                    acceptIntent,
                    PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
                )
                addAction(R.drawable.ic_check, "接受", acceptPendingIntent)
            }
            
            TaskStatus.ACCEPTED -> {
                // Add start action
                val startIntent = createTaskActionIntent(task.id, ACTION_START_TASK)
                val startPendingIntent = PendingIntent.getBroadcast(
                    context,
                    task.id.hashCode(),
                    startIntent,
                    PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
                )
                addAction(R.drawable.ic_play, "开始", startPendingIntent)
            }
            
            TaskStatus.IN_PROGRESS -> {
                // Add complete action
                val completeIntent = createTaskActionIntent(task.id, ACTION_COMPLETE_TASK)
                val completePendingIntent = PendingIntent.getBroadcast(
                    context,
                    task.id.hashCode(),
                    completeIntent,
                    PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
                )
                addAction(R.drawable.ic_done, "完成", completePendingIntent)
            }
            
            else -> {
                // No actions for completed or cancelled tasks
            }
        }
        
        return this
    }
    
    private fun createTaskActionIntent(taskId: String, action: String): Intent {
        return Intent(context, TaskActionReceiver::class.java).apply {
            this.action = action
            putExtra("task_id", taskId)
        }
    }
    
    private fun getTaskStatusIcon(status: TaskStatus): Int {
        return when (status) {
            TaskStatus.PENDING -> R.drawable.ic_pending
            TaskStatus.ACCEPTED -> R.drawable.ic_accepted
            TaskStatus.IN_PROGRESS -> R.drawable.ic_in_progress
            TaskStatus.COMPLETED -> R.drawable.ic_completed
            TaskStatus.CANCELLED -> R.drawable.ic_cancelled
        }
    }
    
    // Clear notifications
    fun clearTaskNotifications(taskId: String) {
        val notificationIds = listOf(
            NOTIFICATION_TASK_ACCEPTED + taskId.hashCode(),
            NOTIFICATION_TASK_STARTED + taskId.hashCode(),
            NOTIFICATION_TASK_COMPLETED + taskId.hashCode(),
            NOTIFICATION_TASK_REMINDER + taskId.hashCode(),
            NOTIFICATION_LOCATION_ARRIVAL + taskId.hashCode()
        )
        
        notificationIds.forEach { id ->
            notificationManager.cancel(id)
        }
    }
    
    fun clearSyncNotifications() {
        notificationManager.cancel(NOTIFICATION_SYNC_PROGRESS)
        notificationManager.cancel(NOTIFICATION_SYNC_COMPLETED)
        notificationManager.cancel(NOTIFICATION_SYNC_ERROR)
    }
    
    fun clearAllNotifications() {
        notificationManager.cancelAll()
    }
    
    companion object {
        // Notification channels
        private const val CHANNEL_TASK_UPDATES = "task_updates"
        private const val CHANNEL_LOCATION_ALERTS = "location_alerts"
        private const val CHANNEL_SYNC_STATUS = "sync_status"
        private const val CHANNEL_SYSTEM = "system"
        
        // Notification IDs
        private const val NOTIFICATION_TASK_ACCEPTED = 1000
        private const val NOTIFICATION_TASK_STARTED = 2000
        private const val NOTIFICATION_TASK_COMPLETED = 3000
        private const val NOTIFICATION_TASK_REMINDER = 4000
        private const val NOTIFICATION_LOCATION_ARRIVAL = 5000
        private const val NOTIFICATION_SYNC_PROGRESS = 6000
        private const val NOTIFICATION_SYNC_COMPLETED = 6001
        private const val NOTIFICATION_SYNC_ERROR = 6002
        private const val NOTIFICATION_CONNECTION_LOST = 7000
        private const val NOTIFICATION_CONNECTION_RESTORED = 7001
        
        // Actions
        const val ACTION_ACCEPT_TASK = "accept_task"
        const val ACTION_START_TASK = "start_task"
        const val ACTION_COMPLETE_TASK = "complete_task"
    }
}