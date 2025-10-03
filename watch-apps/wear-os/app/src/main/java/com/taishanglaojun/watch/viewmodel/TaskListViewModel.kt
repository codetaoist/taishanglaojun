package com.taishanglaojun.watch.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.taishanglaojun.watch.data.repository.TaskRepository
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.data.models.TaskStatus
import com.taishanglaojun.watch.data.models.TaskPriority
import com.taishanglaojun.watch.services.ConnectivityService
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TaskListViewModel @Inject constructor(
    private val taskRepository: TaskRepository,
    private val connectivityService: ConnectivityService
) : ViewModel() {

    // UI State
    private val _uiState = MutableStateFlow(TaskListUiState())
    val uiState: StateFlow<TaskListUiState> = _uiState.asStateFlow()

    // Filter and sort state
    private val _filterState = MutableStateFlow(TaskFilter.ALL)
    val filterState: StateFlow<TaskFilter> = _filterState.asStateFlow()

    private val _sortState = MutableStateFlow(TaskSort.CREATED_DESC)
    val sortState: StateFlow<TaskSort> = _sortState.asStateFlow()

    private val _searchQuery = MutableStateFlow("")
    val searchQuery: StateFlow<String> = _searchQuery.asStateFlow()

    // Filtered and sorted tasks
    val tasks = combine(
        taskRepository.getAllTasks(),
        filterState,
        sortState,
        searchQuery
    ) { allTasks, filter, sort, query ->
        var filteredTasks = allTasks

        // Apply search filter
        if (query.isNotBlank()) {
            filteredTasks = filteredTasks.filter { task ->
                task.title.contains(query, ignoreCase = true) ||
                task.description.contains(query, ignoreCase = true)
            }
        }

        // Apply status filter
        filteredTasks = when (filter) {
            TaskFilter.ALL -> filteredTasks
            TaskFilter.PENDING -> filteredTasks.filter { it.status == TaskStatus.PENDING }
            TaskFilter.IN_PROGRESS -> filteredTasks.filter { it.status == TaskStatus.IN_PROGRESS }
            TaskFilter.COMPLETED -> filteredTasks.filter { it.status == TaskStatus.COMPLETED }
            TaskFilter.HIGH_PRIORITY -> filteredTasks.filter { it.priority == TaskPriority.HIGH }
            TaskFilter.QUICK_ACTIONS -> filteredTasks.filter { it.isQuickActionAvailable }
            TaskFilter.OFFLINE_AVAILABLE -> filteredTasks.filter { it.canCompleteOffline }
        }

        // Apply sorting
        when (sort) {
            TaskSort.CREATED_DESC -> filteredTasks.sortedByDescending { it.createdAt }
            TaskSort.CREATED_ASC -> filteredTasks.sortedBy { it.createdAt }
            TaskSort.PRIORITY_DESC -> filteredTasks.sortedWith(
                compareByDescending<WatchTask> { it.priority.ordinal }
                    .thenByDescending { it.createdAt }
            )
            TaskSort.PRIORITY_ASC -> filteredTasks.sortedWith(
                compareBy<WatchTask> { it.priority.ordinal }
                    .thenByDescending { it.createdAt }
            )
            TaskSort.STATUS -> filteredTasks.sortedWith(
                compareBy<WatchTask> { it.status.ordinal }
                    .thenByDescending { it.createdAt }
            )
            TaskSort.TITLE -> filteredTasks.sortedBy { it.title }
            TaskSort.DUE_DATE -> filteredTasks.sortedWith(
                compareBy<WatchTask> { it.dueDate ?: Long.MAX_VALUE }
                    .thenByDescending { it.createdAt }
            )
        }
    }.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = emptyList()
    )

    // Connection state
    val connectionState = connectivityService.connectionState
    val isSyncing = connectivityService.isSyncing

    init {
        loadTasks()
    }

    private fun loadTasks() {
        viewModelScope.launch {
            try {
                _uiState.update { it.copy(isLoading = true, error = null) }
                taskRepository.loadTasks()
                _uiState.update { it.copy(isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(
                        isLoading = false, 
                        error = "加载任务失败: ${e.message}"
                    ) 
                }
            }
        }
    }

    fun setFilter(filter: TaskFilter) {
        _filterState.value = filter
    }

    fun setSort(sort: TaskSort) {
        _sortState.value = sort
    }

    fun setSearchQuery(query: String) {
        _searchQuery.value = query
    }

    fun clearSearch() {
        _searchQuery.value = ""
    }

    fun refreshTasks() {
        loadTasks()
    }

    fun syncTasks() {
        viewModelScope.launch {
            try {
                connectivityService.forceSync()
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "同步失败: ${e.message}")
                }
            }
        }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }

    // Task actions
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
                taskRepository.updateTaskStatus(taskId, TaskStatus.IN_PROGRESS)
                connectivityService.updateTaskProgress(taskId, 0, "任务已开始")
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "开始任务失败: ${e.message}")
                }
            }
        }
    }

    fun pauseTask(taskId: String) {
        viewModelScope.launch {
            try {
                taskRepository.updateTaskStatus(taskId, TaskStatus.PENDING)
                connectivityService.updateTaskProgress(taskId, null, "任务已暂停")
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "暂停任务失败: ${e.message}")
                }
            }
        }
    }

    fun completeTask(taskId: String) {
        viewModelScope.launch {
            try {
                taskRepository.updateTaskStatus(taskId, TaskStatus.COMPLETED)
                connectivityService.updateTaskProgress(taskId, 100, "任务已完成")
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "完成任务失败: ${e.message}")
                }
            }
        }
    }

    fun cancelTask(taskId: String) {
        viewModelScope.launch {
            try {
                taskRepository.updateTaskStatus(taskId, TaskStatus.CANCELLED)
                connectivityService.updateTaskProgress(taskId, null, "任务已取消")
            } catch (e: Exception) {
                _uiState.update { 
                    it.copy(error = "取消任务失败: ${e.message}")
                }
            }
        }
    }
}

data class TaskListUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val showFilterDialog: Boolean = false,
    val showSortDialog: Boolean = false,
    val selectedTaskId: String? = null,
    val showTaskActionDialog: Boolean = false
)

enum class TaskFilter {
    ALL,
    PENDING,
    IN_PROGRESS,
    COMPLETED,
    HIGH_PRIORITY,
    QUICK_ACTIONS,
    OFFLINE_AVAILABLE
}

enum class TaskSort {
    CREATED_DESC,
    CREATED_ASC,
    PRIORITY_DESC,
    PRIORITY_ASC,
    STATUS,
    TITLE,
    DUE_DATE
}