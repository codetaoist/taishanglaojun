package com.taishanglaojun.watch.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.taishanglaojun.watch.data.repository.TaskRepository
import com.taishanglaojun.watch.data.models.TaskStatistics
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.services.ConnectivityService
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MainViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val connectivityService: ConnectivityService
) : ViewModel() {

    // UI State
    private val _uiState = MutableStateFlow(MainUiState())
    val uiState: StateFlow<MainUiState> = _uiState.asStateFlow()

    // Connection state from service
    val connectionState = connectivityService.connectionState
    val lastSyncTime = connectivityService.lastSyncTime
    val isSyncing = connectivityService.isSyncing

    // Task statistics
    val taskStatistics = taskRepository.getTaskStatistics()
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = TaskStatistics()
        )

    // Recent tasks
    val recentTasks = taskRepository.getRecentTasks(limit = 3)
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )

    init {
        loadData()
        observeErrors()
    }

    private fun loadData() {
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isLoading = true, error = null) }
                
                // Initialize connectivity service if needed
                connectivityService.initialize()
                
                // Load cached tasks
                taskRepository.loadTasks()
                
                _uiState.update { it.copy(isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isLoading = false, 
                        error = "加载数据失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    private fun observeErrors() {
        viewModelScope.launch {
            // Observe connectivity errors
            connectivityService.connectionState.collect { state ->
                if (state.error != null) {
                    _uiState.update { 
                        it.copy(error = "连接错误: ${state.error}")
                    }
                }
            }
        }
    }

    fun forceSync() {
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(error = null) }
                connectivityService.forceSync()
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "同步失败: ${e.message}")
                }
            }
        }
    }

    fun refreshData() {
        loadData()
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }

    fun retryLastAction() {
        when {
            uiState.value.error?.contains("连接") == true -> forceSync()
            uiState.value.error?.contains("加载") == true -> refreshData()
            else -> refreshData()
        }
    }

    // Quick actions
    fun acceptTask(taskId: String) {
        viewModelScope.launch {
            try {
                connectivityService.acceptTask(taskId)
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "接受任务失败: ${e.message}")
                }
            }
        }
    }

    fun startTask(taskId: String) {
        viewModelScope.launch {
            try {
                taskRepository.updateTaskStatus(taskId, com.taishanglaojun.watch.data.models.TaskStatus.IN_PROGRESS)
                connectivityService.updateTaskProgress(taskId, 0, "任务已开始")
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "开始任务失败: ${e.message}")
                }
            }
        }
    }

    fun completeTask(taskId: String) {
        viewModelScope.launch {
            try {
                taskRepository.updateTaskStatus(taskId, com.taishanglaojun.watch.data.models.TaskStatus.COMPLETED)
                connectivityService.updateTaskProgress(taskId, 100, "任务已完成")
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "完成任务失败: ${e.message}")
                }
            }
        }
    }
}

data class MainUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val lastRefreshTime: Long = System.currentTimeMillis()
)