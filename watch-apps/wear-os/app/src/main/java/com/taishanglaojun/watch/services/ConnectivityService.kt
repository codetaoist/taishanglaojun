package com.taishanglaojun.watch.services

import android.content.Context
import android.util.Log
import com.google.android.gms.tasks.Task
import com.google.android.gms.wearable.*
import com.taishanglaojun.watch.AppConfig
import com.taishanglaojun.watch.LogTags
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.data.repository.TaskRepository
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.coroutines.resume
import kotlin.coroutines.suspendCoroutine

/**
 * Wear OS 连接服务
 * 负责与配对的手机应用进行通信和数据同步
 */
@Singleton
class ConnectivityService @Inject constructor(
    @ApplicationContext private val context: Context,
    private val taskRepository: TaskRepository
) : WearableListenerService() {
    
    companion object {
        private const val TAG = LogTags.CONNECTIVITY
        
        // 消息路径
        private const val PATH_TASK_REQUEST = "/task/request"
        private const val PATH_TASK_ACCEPT = "/task/accept"
        private const val PATH_TASK_UPDATE = "/task/update"
        private const val PATH_SYNC_REQUEST = "/sync/request"
        private const val PATH_HEARTBEAT = "/heartbeat"
        
        // 数据键
        private const val KEY_TASKS = "tasks"
        private const val KEY_TASK_ID = "task_id"
        private const val KEY_PROGRESS = "progress"
        private const val KEY_STATUS = "status"
        private const val KEY_TIMESTAMP = "timestamp"
        private const val KEY_SUCCESS = "success"
        private const val KEY_ERROR = "error"
    }
    
    // Wearable API 客户端
    private lateinit var dataClient: DataClient
    private lateinit var messageClient: MessageClient
    private lateinit var capabilityClient: CapabilityClient
    private lateinit var nodeClient: NodeClient
    
    // 协程作用域
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    
    // 连接状态
    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState.asStateFlow()
    
    private val _lastSyncTime = MutableStateFlow<Instant?>(null)
    val lastSyncTime: StateFlow<Instant?> = _lastSyncTime.asStateFlow()
    
    private val _isDataSyncing = MutableStateFlow(false)
    val isDataSyncing: StateFlow<Boolean> = _isDataSyncing.asStateFlow()
    
    // 连接的节点
    private var connectedNodes: Set<Node> = emptySet()
    
    // 心跳定时器
    private var heartbeatJob: Job? = null
    
    /**
     * 初始化连接服务
     */
    fun initialize() {
        try {
            dataClient = Wearable.getDataClient(context)
            messageClient = Wearable.getMessageClient(context)
            capabilityClient = Wearable.getCapabilityClient(context)
            nodeClient = Wearable.getNodeClient(context)
            
            // 添加监听器
            dataClient.addListener(this)
            messageClient.addListener(this)
            capabilityClient.addListener(this)
            
            // 检查连接状态
            checkConnectedNodes()
            
            // 启动心跳
            startHeartbeat()
            
            Log.d(TAG, "ConnectivityService initialized successfully")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize ConnectivityService", e)
        }
    }
    
    /**
     * 启动数据同步
     */
    fun startDataSync() {
        serviceScope.launch {
            try {
                _isDataSyncing.value = true
                
                // 请求任务数据
                val success = requestTaskData()
                
                if (success) {
                    _lastSyncTime.value = Clock.System.now()
                    Log.d(TAG, "Data sync completed successfully")
                } else {
                    Log.w(TAG, "Data sync failed")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error during data sync", e)
            } finally {
                _isDataSyncing.value = false
            }
        }
    }
    
    /**
     * 请求任务数据
     */
    suspend fun requestTaskData(): Boolean = suspendCoroutine { continuation ->
        val targetNode = getTargetNode()
        if (targetNode == null) {
            Log.w(TAG, "No target node available for task request")
            continuation.resume(false)
            return@suspendCoroutine
        }
        
        val requestData = mapOf(
            KEY_TIMESTAMP to Clock.System.now().epochSeconds
        ).toByteArray()
        
        messageClient.sendMessage(targetNode.id, PATH_TASK_REQUEST, requestData)
            .addOnSuccessListener {
                Log.d(TAG, "Task request sent successfully")
                continuation.resume(true)
            }
            .addOnFailureListener { e ->
                Log.e(TAG, "Failed to send task request", e)
                continuation.resume(false)
            }
    }
    
    /**
     * 接受任务
     */
    suspend fun acceptTask(taskId: String): Boolean = suspendCoroutine { continuation ->
        val targetNode = getTargetNode()
        if (targetNode == null) {
            continuation.resume(false)
            return@suspendCoroutine
        }
        
        val acceptData = mapOf(
            KEY_TASK_ID to taskId,
            KEY_TIMESTAMP to Clock.System.now().epochSeconds
        ).toByteArray()
        
        messageClient.sendMessage(targetNode.id, PATH_TASK_ACCEPT, acceptData)
            .addOnSuccessListener {
                Log.d(TAG, "Task accept request sent: $taskId")
                continuation.resume(true)
            }
            .addOnFailureListener { e ->
                Log.e(TAG, "Failed to send task accept request", e)
                continuation.resume(false)
            }
    }
    
    /**
     * 更新任务进度
     */
    suspend fun updateTaskProgress(taskId: String, progress: Double): Boolean = suspendCoroutine { continuation ->
        val targetNode = getTargetNode()
        if (targetNode == null) {
            continuation.resume(false)
            return@suspendCoroutine
        }
        
        val updateData = mapOf(
            KEY_TASK_ID to taskId,
            KEY_PROGRESS to progress,
            KEY_TIMESTAMP to Clock.System.now().epochSeconds
        ).toByteArray()
        
        messageClient.sendMessage(targetNode.id, PATH_TASK_UPDATE, updateData)
            .addOnSuccessListener {
                Log.d(TAG, "Task progress update sent: $taskId, progress: $progress")
                continuation.resume(true)
            }
            .addOnFailureListener { e ->
                Log.e(TAG, "Failed to send task progress update", e)
                continuation.resume(false)
            }
    }
    
    /**
     * 强制同步
     */
    suspend fun forceSync(): Boolean = suspendCoroutine { continuation ->
        val targetNode = getTargetNode()
        if (targetNode == null) {
            continuation.resume(false)
            return@suspendCoroutine
        }
        
        val syncData = mapOf(
            KEY_TIMESTAMP to Clock.System.now().epochSeconds
        ).toByteArray()
        
        messageClient.sendMessage(targetNode.id, PATH_SYNC_REQUEST, syncData)
            .addOnSuccessListener {
                Log.d(TAG, "Force sync request sent")
                continuation.resume(true)
            }
            .addOnFailureListener { e ->
                Log.e(TAG, "Failed to send force sync request", e)
                continuation.resume(false)
            }
    }
    
    // MARK: - WearableListenerService 回调
    
    override fun onDataChanged(dataEvents: DataEventBuffer) {
        super.onDataChanged(dataEvents)
        
        for (event in dataEvents) {
            if (event.type == DataEvent.TYPE_CHANGED) {
                val dataItem = event.dataItem
                handleDataItemChanged(dataItem)
            }
        }
    }
    
    override fun onMessageReceived(messageEvent: MessageEvent) {
        super.onMessageReceived(messageEvent)
        
        Log.d(TAG, "Message received: ${messageEvent.path}")
        
        when (messageEvent.path) {
            PATH_TASK_REQUEST -> handleTaskResponse(messageEvent.data)
            PATH_SYNC_REQUEST -> handleSyncResponse(messageEvent.data)
            PATH_HEARTBEAT -> handleHeartbeatResponse(messageEvent.data)
            else -> Log.w(TAG, "Unknown message path: ${messageEvent.path}")
        }
    }
    
    override fun onCapabilityChanged(capabilityInfo: CapabilityInfo) {
        super.onCapabilityChanged(capabilityInfo)
        
        connectedNodes = capabilityInfo.nodes
        updateConnectionState()
        
        Log.d(TAG, "Capability changed, connected nodes: ${connectedNodes.size}")
    }
    
    override fun onPeerConnected(peer: Node) {
        super.onPeerConnected(peer)
        
        Log.d(TAG, "Peer connected: ${peer.displayName}")
        checkConnectedNodes()
    }
    
    override fun onPeerDisconnected(peer: Node) {
        super.onPeerDisconnected(peer)
        
        Log.d(TAG, "Peer disconnected: ${peer.displayName}")
        checkConnectedNodes()
    }
    
    // MARK: - 私有方法
    
    private fun handleDataItemChanged(dataItem: DataItem) {
        when (dataItem.uri.path) {
            "/tasks" -> {
                val dataMap = DataMapItem.fromDataItem(dataItem).dataMap
                handleTasksDataUpdate(dataMap)
            }
            "/sync_status" -> {
                val dataMap = DataMapItem.fromDataItem(dataItem).dataMap
                handleSyncStatusUpdate(dataMap)
            }
        }
    }
    
    private fun handleTasksDataUpdate(dataMap: DataMap) {
        serviceScope.launch {
            try {
                val tasksData = dataMap.getByteArray(KEY_TASKS)
                if (tasksData != null) {
                    val tasks = tasksData.fromByteArray<List<WatchTask>>()
                    taskRepository.updateTasks(tasks)
                    
                    _lastSyncTime.value = Clock.System.now()
                    Log.d(TAG, "Tasks updated: ${tasks.size} tasks")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error handling tasks data update", e)
            }
        }
    }
    
    private fun handleSyncStatusUpdate(dataMap: DataMap) {
        val timestamp = dataMap.getLong(KEY_TIMESTAMP, 0L)
        if (timestamp > 0) {
            _lastSyncTime.value = Instant.fromEpochSeconds(timestamp)
        }
    }
    
    private fun handleTaskResponse(data: ByteArray?) {
        if (data == null) return
        
        serviceScope.launch {
            try {
                val responseMap = data.fromByteArray<Map<String, Any>>()
                val success = responseMap[KEY_SUCCESS] as? Boolean ?: false
                
                if (success) {
                    val tasksData = responseMap[KEY_TASKS] as? ByteArray
                    if (tasksData != null) {
                        val tasks = tasksData.fromByteArray<List<WatchTask>>()
                        taskRepository.updateTasks(tasks)
                        _lastSyncTime.value = Clock.System.now()
                    }
                } else {
                    val error = responseMap[KEY_ERROR] as? String
                    Log.e(TAG, "Task request failed: $error")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error handling task response", e)
            }
        }
    }
    
    private fun handleSyncResponse(data: ByteArray?) {
        if (data == null) return
        
        try {
            val responseMap = data.fromByteArray<Map<String, Any>>()
            val success = responseMap[KEY_SUCCESS] as? Boolean ?: false
            
            if (success) {
                _lastSyncTime.value = Clock.System.now()
                Log.d(TAG, "Sync completed successfully")
            } else {
                val error = responseMap[KEY_ERROR] as? String
                Log.e(TAG, "Sync failed: $error")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error handling sync response", e)
        }
    }
    
    private fun handleHeartbeatResponse(data: ByteArray?) {
        // 更新连接状态
        _connectionState.value = ConnectionState.CONNECTED
        Log.d(TAG, "Heartbeat response received")
    }
    
    private fun checkConnectedNodes() {
        nodeClient.connectedNodes.addOnSuccessListener { nodes ->
            connectedNodes = nodes.toSet()
            updateConnectionState()
            
            Log.d(TAG, "Connected nodes updated: ${nodes.size} nodes")
        }.addOnFailureListener { e ->
            Log.e(TAG, "Failed to get connected nodes", e)
            _connectionState.value = ConnectionState.DISCONNECTED
        }
    }
    
    private fun updateConnectionState() {
        _connectionState.value = when {
            connectedNodes.isEmpty() -> ConnectionState.DISCONNECTED
            connectedNodes.any { it.isNearby } -> ConnectionState.CONNECTED
            else -> ConnectionState.PAIRED
        }
    }
    
    private fun getTargetNode(): Node? {
        return connectedNodes.firstOrNull { it.isNearby }
            ?: connectedNodes.firstOrNull()
    }
    
    private fun startHeartbeat() {
        heartbeatJob?.cancel()
        heartbeatJob = serviceScope.launch {
            while (isActive) {
                try {
                    sendHeartbeat()
                    delay(AppConfig.HEARTBEAT_INTERVAL_MS)
                } catch (e: Exception) {
                    Log.e(TAG, "Error in heartbeat", e)
                    delay(5000) // 错误时等待5秒再重试
                }
            }
        }
    }
    
    private suspend fun sendHeartbeat() {
        val targetNode = getTargetNode()
        if (targetNode != null) {
            val heartbeatData = mapOf(
                KEY_TIMESTAMP to Clock.System.now().epochSeconds
            ).toByteArray()
            
            try {
                messageClient.sendMessage(targetNode.id, PATH_HEARTBEAT, heartbeatData).await()
            } catch (e: Exception) {
                Log.w(TAG, "Heartbeat failed", e)
                _connectionState.value = ConnectionState.DISCONNECTED
            }
        }
    }
    
    /**
     * 清理资源
     */
    fun cleanup() {
        heartbeatJob?.cancel()
        serviceScope.cancel()
        
        try {
            dataClient.removeListener(this)
            messageClient.removeListener(this)
            capabilityClient.removeListener(this)
        } catch (e: Exception) {
            Log.e(TAG, "Error during cleanup", e)
        }
    }
}

/**
 * 连接状态枚举
 */
enum class ConnectionState(val displayName: String) {
    DISCONNECTED("未连接"),
    PAIRED("已配对"),
    CONNECTED("已连接")
}

/**
 * 扩展函数：将对象转换为字节数组
 */
private fun Map<String, Any>.toByteArray(): ByteArray {
    return toString().toByteArray()
}

/**
 * 扩展函数：从字节数组解析对象
 */
private inline fun <reified T> ByteArray.fromByteArray(): T {
    // 这里应该使用实际的序列化库，如 Gson 或 Kotlinx.serialization
    // 为了简化，这里使用 toString() 方法
    throw NotImplementedError("需要实现实际的序列化/反序列化逻辑")
}

/**
 * 扩展函数：等待 Task 完成
 */
private suspend fun <T> Task<T>.await(): T = suspendCoroutine { continuation ->
    addOnSuccessListener { result ->
        continuation.resume(result)
    }
    addOnFailureListener { exception ->
        throw exception
    }
}