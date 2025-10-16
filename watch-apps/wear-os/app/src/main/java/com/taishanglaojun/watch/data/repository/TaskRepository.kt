package com.taishanglaojun.watch.data.repository

import android.content.Context
import android.util.Log
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import androidx.datastore.preferences.preferencesDataStore
import com.taishanglaojun.watch.LogTags
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.data.models.TaskStatus
import com.taishanglaojun.watch.data.models.TaskPriority
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.*
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 任务数据仓库
 * 负责管理任务数据的存储、缓存和同步
 */
@Singleton
class TaskRepository @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private const val TAG = LogTags.TASK_REPOSITORY
        private const val DATASTORE_NAME = "task_preferences"
        
        // DataStore 键
        private val TASKS_KEY = stringPreferencesKey("tasks")
        private val LAST_SYNC_KEY = longPreferencesKey("last_sync")
        private val CACHE_VERSION_KEY = intPreferencesKey("cache_version")
        private val OFFLINE_TASKS_KEY = stringPreferencesKey("offline_tasks")
    }
    
    // DataStore 实例
    private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = DATASTORE_NAME)
    
    // 内存缓存
    private val _tasks = MutableStateFlow<List<WatchTask>>(emptyList())
    val tasks: StateFlow<List<WatchTask>> = _tasks.asStateFlow()
    
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    private val _lastSyncTime = MutableStateFlow<Instant?>(null)
    val lastSyncTime: StateFlow<Instant?> = _lastSyncTime.asStateFlow()
    
    // 离线任务队列
    private val _offlineTasks = MutableStateFlow<List<WatchTask>>(emptyList())
    val offlineTasks: StateFlow<List<WatchTask>> = _offlineTasks.asStateFlow()
    
    init {
        // 初始化时加载缓存数据
        loadCachedTasks()
    }
    
    /**
     * 加载缓存的任务数据
     */
    private fun loadCachedTasks() {
        context.dataStore.data
            .catch { exception ->
                Log.e(TAG, "Error reading cached tasks", exception)
                emit(emptyPreferences())
            }
            .onEach { preferences ->
                try {
                    // 加载任务列表
                    val tasksJson = preferences[TASKS_KEY]
                    if (!tasksJson.isNullOrEmpty()) {
                        val tasks = parseTasksFromJson(tasksJson)
                        _tasks.value = tasks
                    }
                    
                    // 加载最后同步时间
                    val lastSync = preferences[LAST_SYNC_KEY]
                    if (lastSync != null && lastSync > 0) {
                        _lastSyncTime.value = Instant.fromEpochSeconds(lastSync)
                    }
                    
                    // 加载离线任务
                    val offlineTasksJson = preferences[OFFLINE_TASKS_KEY]
                    if (!offlineTasksJson.isNullOrEmpty()) {
                        val offlineTasks = parseTasksFromJson(offlineTasksJson)
                        _offlineTasks.value = offlineTasks
                    }
                    
                    Log.d(TAG, "Cached tasks loaded: ${_tasks.value.size} tasks, ${_offlineTasks.value.size} offline tasks")
                } catch (e: Exception) {
                    Log.e(TAG, "Error parsing cached tasks", e)
                }
            }
            .launchIn(kotlinx.coroutines.GlobalScope)
    }
    
    /**
     * 更新任务列表
     */
    suspend fun updateTasks(newTasks: List<WatchTask>) {
        try {
            _isLoading.value = true
            
            // 更新内存缓存
            _tasks.value = newTasks
            
            // 保存到 DataStore
            context.dataStore.edit { preferences ->
                preferences[TASKS_KEY] = serializeTasksToJson(newTasks)
                preferences[LAST_SYNC_KEY] = Clock.System.now().epochSeconds
                preferences[CACHE_VERSION_KEY] = (preferences[CACHE_VERSION_KEY] ?: 0) + 1
            }
            
            _lastSyncTime.value = Clock.System.now()
            
            Log.d(TAG, "Tasks updated: ${newTasks.size} tasks")
        } catch (e: Exception) {
            Log.e(TAG, "Error updating tasks", e)
        } finally {
            _isLoading.value = false
        }
    }
    
    /**
     * 获取指定ID的任务
     */
    fun getTaskById(taskId: String): WatchTask? {
        return _tasks.value.find { it.id == taskId }
    }
    
    /**
     * 获取按状态筛选的任务
     */
    fun getTasksByStatus(status: TaskStatus): Flow<List<WatchTask>> {
        return tasks.map { taskList ->
            taskList.filter { it.status == status }
        }
    }
    
    /**
     * 获取按优先级筛选的任务
     */
    fun getTasksByPriority(priority: TaskPriority): Flow<List<WatchTask>> {
        return tasks.map { taskList ->
            taskList.filter { it.priority == priority }
        }
    }
    
    /**
     * 获取可快速操作的任务
     */
    fun getQuickActionTasks(): Flow<List<WatchTask>> {
        return tasks.map { taskList ->
            taskList.filter { it.isQuickActionAvailable }
        }
    }
    
    /**
     * 获取逾期任务
     */
    fun getOverdueTasks(): Flow<List<WatchTask>> {
        return tasks.map { taskList ->
            taskList.filter { it.isOverdue() }
        }
    }
    
    /**
     * 搜索任务
     */
    fun searchTasks(query: String): Flow<List<WatchTask>> {
        return tasks.map { taskList ->
            if (query.isBlank()) {
                taskList
            } else {
                taskList.filter { task ->
                    task.title.contains(query, ignoreCase = true) ||
                    task.description.contains(query, ignoreCase = true)
                }
            }
        }
    }
    
    /**
     * 更新任务状态
     */
    suspend fun updateTaskStatus(taskId: String, newStatus: TaskStatus) {
        try {
            val updatedTasks = _tasks.value.map { task ->
                if (task.id == taskId) {
                    task.copy(
                        status = newStatus,
                        lastUpdated = Clock.System.now()
                    )
                } else {
                    task
                }
            }
            
            updateTasks(updatedTasks)
            
            Log.d(TAG, "Task status updated: $taskId -> $newStatus")
        } catch (e: Exception) {
            Log.e(TAG, "Error updating task status", e)
        }
    }
    
    /**
     * 更新任务进度
     */
    suspend fun updateTaskProgress(taskId: String, progress: Double) {
        try {
            val updatedTasks = _tasks.value.map { task ->
                if (task.id == taskId) {
                    task.copy(
                        progress = progress.coerceIn(0.0, 1.0),
                        lastUpdated = Clock.System.now()
                    )
                } else {
                    task
                }
            }
            
            updateTasks(updatedTasks)
            
            Log.d(TAG, "Task progress updated: $taskId -> $progress")
        } catch (e: Exception) {
            Log.e(TAG, "Error updating task progress", e)
        }
    }
    
    /**
     * 添加离线任务
     */
    suspend fun addOfflineTask(task: WatchTask) {
        try {
            val updatedOfflineTasks = _offlineTasks.value + task
            _offlineTasks.value = updatedOfflineTasks
            
            // 保存到 DataStore
            context.dataStore.edit { preferences ->
                preferences[OFFLINE_TASKS_KEY] = serializeTasksToJson(updatedOfflineTasks)
            }
            
            Log.d(TAG, "Offline task added: ${task.id}")
        } catch (e: Exception) {
            Log.e(TAG, "Error adding offline task", e)
        }
    }
    
    /**
     * 移除离线任务
     */
    suspend fun removeOfflineTask(taskId: String) {
        try {
            val updatedOfflineTasks = _offlineTasks.value.filter { it.id != taskId }
            _offlineTasks.value = updatedOfflineTasks
            
            // 保存到 DataStore
            context.dataStore.edit { preferences ->
                preferences[OFFLINE_TASKS_KEY] = serializeTasksToJson(updatedOfflineTasks)
            }
            
            Log.d(TAG, "Offline task removed: $taskId")
        } catch (e: Exception) {
            Log.e(TAG, "Error removing offline task", e)
        }
    }
    
    /**
     * 清空离线任务队列
     */
    suspend fun clearOfflineTasks() {
        try {
            _offlineTasks.value = emptyList()
            
            context.dataStore.edit { preferences ->
                preferences.remove(OFFLINE_TASKS_KEY)
            }
            
            Log.d(TAG, "Offline tasks cleared")
        } catch (e: Exception) {
            Log.e(TAG, "Error clearing offline tasks", e)
        }
    }
    
    /**
     * 获取任务统计信息
     */
    fun getTaskStatistics(): Flow<TaskStatistics> {
        return tasks.map { taskList ->
            TaskStatistics(
                total = taskList.size,
                active = taskList.count { it.status == TaskStatus.IN_PROGRESS },
                completed = taskList.count { it.status == TaskStatus.COMPLETED },
                pending = taskList.count { it.status == TaskStatus.PENDING },
                overdue = taskList.count { it.isOverdue() },
                highPriority = taskList.count { it.priority == TaskPriority.HIGH },
                quickActions = taskList.count { it.isQuickActionAvailable }
            )
        }
    }
    
    /**
     * 清空所有缓存
     */
    suspend fun clearCache() {
        try {
            _tasks.value = emptyList()
            _offlineTasks.value = emptyList()
            _lastSyncTime.value = null
            
            context.dataStore.edit { preferences ->
                preferences.clear()
            }
            
            Log.d(TAG, "Cache cleared")
        } catch (e: Exception) {
            Log.e(TAG, "Error clearing cache", e)
        }
    }
    
    /**
     * 检查缓存是否过期
     */
    fun isCacheExpired(): Boolean {
        val lastSync = _lastSyncTime.value ?: return true
        val now = Clock.System.now()
        val cacheAge = now.minus(lastSync)
        
        // 缓存超过5分钟视为过期
        return cacheAge.inWholeMinutes > 5
    }
    
    // MARK: - 私有方法
    
    /**
     * 将任务列表序列化为JSON字符串
     */
    private fun serializeTasksToJson(tasks: List<WatchTask>): String {
        // 这里应该使用实际的JSON序列化库，如 Gson 或 Kotlinx.serialization
        // 为了简化，这里使用简单的字符串拼接
        return tasks.joinToString(separator = "|||") { task ->
            "${task.id}::${task.title}::${task.description}::${task.status}::${task.priority}::${task.difficulty}::${task.reward}::${task.progress}::${task.isQuickActionAvailable}::${task.requiresPhoneConnection}::${task.canCompleteOffline}::${task.lastUpdated.epochSeconds}"
        }
    }
    
    /**
     * 从JSON字符串解析任务列表
     */
    private fun parseTasksFromJson(json: String): List<WatchTask> {
        return try {
            if (json.isBlank()) return emptyList()
            
            json.split("|||").mapNotNull { taskString ->
                val parts = taskString.split("::")
                if (parts.size >= 12) {
                    WatchTask(
                        id = parts[0],
                        title = parts[1],
                        description = parts[2],
                        status = TaskStatus.valueOf(parts[3]),
                        priority = TaskPriority.valueOf(parts[4]),
                        difficulty = parts[5].toIntOrNull() ?: 1,
                        reward = parts[6].toDoubleOrNull() ?: 0.0,
                        progress = parts[7].toDoubleOrNull() ?: 0.0,
                        isQuickActionAvailable = parts[8].toBoolean(),
                        requiresPhoneConnection = parts[9].toBoolean(),
                        canCompleteOffline = parts[10].toBoolean(),
                        lastUpdated = Instant.fromEpochSeconds(parts[11].toLongOrNull() ?: 0L)
                    )
                } else {
                    null
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error parsing tasks from JSON", e)
            emptyList()
        }
    }
}

/**
 * 任务统计信息数据类
 */
data class TaskStatistics(
    val total: Int = 0,
    val active: Int = 0,
    val completed: Int = 0,
    val pending: Int = 0,
    val overdue: Int = 0,
    val highPriority: Int = 0,
    val quickActions: Int = 0
) {
    val completionRate: Double
        get() = if (total > 0) completed.toDouble() / total else 0.0
    
    val activeRate: Double
        get() = if (total > 0) active.toDouble() / total else 0.0
    
    val overdueRate: Double
        get() = if (total > 0) overdue.toDouble() / total else 0.0
}