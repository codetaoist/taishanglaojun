package com.taishanglaojun.watch.service

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.taishanglaojun.watch.data.repository.TaskRepository
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import javax.inject.Inject

@AndroidEntryPoint
class TaskActionReceiver : BroadcastReceiver() {
    
    @Inject
    lateinit var taskRepository: TaskRepository
    
    @Inject
    lateinit var notificationService: NotificationService
    
    private val scope = CoroutineScope(Dispatchers.IO)
    
    override fun onReceive(context: Context?, intent: Intent?) {
        if (context == null || intent == null) return
        
        val taskId = intent.getStringExtra("task_id") ?: return
        val action = intent.action ?: return
        
        scope.launch {
            try {
                when (action) {
                    NotificationService.ACTION_ACCEPT_TASK -> {
                        taskRepository.acceptTask(taskId)
                        val task = taskRepository.getTaskById(taskId)
                        task?.let { notificationService.showTaskAcceptedNotification(it) }
                    }
                    
                    NotificationService.ACTION_START_TASK -> {
                        taskRepository.startTask(taskId)
                        val task = taskRepository.getTaskById(taskId)
                        task?.let { notificationService.showTaskStartedNotification(it) }
                    }
                    
                    NotificationService.ACTION_COMPLETE_TASK -> {
                        taskRepository.completeTask(taskId)
                        val task = taskRepository.getTaskById(taskId)
                        task?.let { notificationService.showTaskCompletedNotification(it) }
                    }
                }
            } catch (e: Exception) {
                // Handle error silently or log
            }
        }
    }
}