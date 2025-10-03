package com.taishanglaojun.watch.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.taishanglaojun.watch.data.repository.TaskRepository
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.data.models.TaskStatus
import com.taishanglaojun.watch.services.ConnectivityService
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskDetailViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val connectivityService: ConnectivityService
) : ViewModel() {

    // UI State
    private val _uiState = MutableStateFlow(TaskDetailUiState())
    val uiState: StateFlow<TaskDetailUiState> = _uiState.asStateFlow()

    // Current task
    private val _currentTaskId = MutableStateFlow<String?>(null)
    val currentTask = _currentTaskId
        .filterNotNull()
        .flatMapLatest { taskId ->
            taskRepository.getTaskById(taskId)
        }
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = null
        )

    // Connection state
    val connectionState = connectivityService.connectionState
    val isSyncing = connectivityService.isSyncing

    fun loadTask(taskId: String) {
        _currentTaskId.value = taskId
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isLoading = true, error = null) }
                // Task will be loaded automatically through the flow
                _uiState.update { it.copy(isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isLoading = false, 
                        error = "加载任务详情失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun acceptTask() {
        val taskId = _currentTaskId.value ?: return
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isProcessing = true, error = null) }
                connectivityService.acceptTask(taskId)
                _uiState.update { it.copy(isProcessing = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        error = "接受任务失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun startTask() {
        val taskId = _currentTaskId.value ?: return
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isProcessing = true, error = null) }
                taskRepository.updateTaskStatus(taskId, TaskStatus.IN_PROGRESS)
                connectivityService.updateTaskProgress(taskId, 0, "任务已开始")
                _uiState.update { it.copy(isProcessing = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        error = "开始任务失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun pauseTask() {
        val taskId = _currentTaskId.value ?: return
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isProcessing = true, error = null) }
                taskRepository.updateTaskStatus(taskId, TaskStatus.PENDING)
                connectivityService.updateTaskProgress(taskId, null, "任务已暂停")
                _uiState.update { it.copy(isProcessing = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        error = "暂停任务失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun completeTask() {
        val taskId = _currentTaskId.value ?: return
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isProcessing = true, error = null) }
                taskRepository.updateTaskStatus(taskId, TaskStatus.COMPLETED)
                connectivityService.updateTaskProgress(taskId, 100, "任务已完成")
                _uiState.update { it.copy(isProcessing = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        error = "完成任务失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun cancelTask() {
        val taskId = _currentTaskId.value ?: return
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isProcessing = true, error = null) }
                taskRepository.updateTaskStatus(taskId, TaskStatus.CANCELLED)
                connectivityService.updateTaskProgress(taskId, null, "任务已取消")
                _uiState.update { it.copy(isProcessing = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        error = "取消任务失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun updateProgress(progress: Int, note: String) {
        val taskId = _currentTaskId.value ?: return
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isProcessing = true, error = null) }
                taskRepository.updateTaskProgress(taskId, progress)
                connectivityService.updateTaskProgress(taskId, progress, note)
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        showProgressDialog = false
                    ) 
                }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isProcessing = false, 
                        error = "更新进度失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun showProgressDialog() {
        _uiState.update { it.copy(showProgressDialog = true) }
    }

    fun hideProgressDialog() {
        _uiState.update { it.copy(showProgressDialog = false) }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }

    fun refreshTask() {
        val taskId = _currentTaskId.value ?: return
        loadTask(taskId)
    }

    // Get available actions for current task
    fun getAvailableActions(task: WatchTask?): List<TaskAction> {
        if (task == null) return emptyList()
        
        return when (task.status) {
            TaskStatus.PENDING -> listOf(
                TaskAction.ACCEPT,
                TaskAction.START,
                TaskAction.CANCEL
            )
            TaskStatus.ACCEPTED -> listOf(
                TaskAction.START,
                TaskAction.CANCEL
            )
            TaskStatus.IN_PROGRESS -> listOf(
                TaskAction.UPDATE_PROGRESS,
                TaskAction.PAUSE,
                TaskAction.COMPLETE,
                TaskAction.CANCEL
            )
            TaskStatus.COMPLETED -> emptyList()
            TaskStatus.CANCELLED -> emptyList()
        }
    }
}

data class TaskDetailUiState(
    val isLoading: Boolean = false,
    val isProcessing: Boolean = false,
    val error: String? = null,
    val showProgressDialog: Boolean = false
)

enum class TaskAction {
    ACCEPT,
    START,
    PAUSE,
    COMPLETE,
    CANCEL,
    UPDATE_PROGRESS
}