package com.taishanglaojun.tracker.data.service

import android.app.Service
import android.content.Intent
import android.os.IBinder
import android.util.Log
import com.taishanglaojun.tracker.data.repository.LocationRepository
import com.taishanglaojun.tracker.data.repository.ChatRepository
import com.taishanglaojun.tracker.network.SecureApiClient
import com.taishanglaojun.tracker.utils.NetworkUtils
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject

/**
 * 数据同步服务
 * 负责在后台同步位置数据、轨迹数据和聊天数据
 */
@AndroidEntryPoint
class DataSyncService : Service() {

    companion object {
        private const val TAG = "DataSyncService"
        private const val SYNC_INTERVAL = 30000L // 30秒同步一次
        private const val RETRY_DELAY = 60000L // 重试延迟60秒
        private const val MAX_RETRY_COUNT = 3
        
        const val ACTION_START_SYNC = "start_sync"
        const val ACTION_STOP_SYNC = "stop_sync"
        const val ACTION_FORCE_SYNC = "force_sync"
    }

    @Inject
    lateinit var locationRepository: LocationRepository
    
    @Inject
    lateinit var chatRepository: ChatRepository
    
    @Inject
    lateinit var apiClient: SecureApiClient
    
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var syncJob: Job? = null
    
    // 同步状态
    private val _syncState = MutableStateFlow(SyncState.IDLE)
    val syncState: StateFlow<SyncState> = _syncState.asStateFlow()
    
    // 同步统计
    private val _syncStats = MutableStateFlow(SyncStats())
    val syncStats: StateFlow<SyncStats> = _syncStats.asStateFlow()

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START_SYNC -> startPeriodicSync()
            ACTION_STOP_SYNC -> stopPeriodicSync()
            ACTION_FORCE_SYNC -> performForceSync()
        }
        return START_STICKY
    }

    override fun onDestroy() {
        super.onDestroy()
        stopPeriodicSync()
        serviceScope.cancel()
    }

    /**
     * 开始周期性同步
     */
    private fun startPeriodicSync() {
        if (syncJob?.isActive == true) {
            Log.d(TAG, "同步任务已在运行")
            return
        }

        Log.d(TAG, "开始周期性数据同步")
        _syncState.value = SyncState.RUNNING

        syncJob = serviceScope.launch {
            while (isActive) {
                try {
                    if (NetworkUtils.isNetworkAvailable(this@DataSyncService)) {
                        performSync()
                    } else {
                        Log.w(TAG, "网络不可用，跳过同步")
                        _syncState.value = SyncState.WAITING_NETWORK
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "同步过程中发生错误", e)
                    _syncState.value = SyncState.ERROR
                }
                
                delay(SYNC_INTERVAL)
            }
        }
    }

    /**
     * 停止周期性同步
     */
    private fun stopPeriodicSync() {
        Log.d(TAG, "停止周期性数据同步")
        syncJob?.cancel()
        _syncState.value = SyncState.IDLE
    }

    /**
     * 执行强制同步
     */
    private fun performForceSync() {
        serviceScope.launch {
            try {
                _syncState.value = SyncState.RUNNING
                performSync()
            } catch (e: Exception) {
                Log.e(TAG, "强制同步失败", e)
                _syncState.value = SyncState.ERROR
            }
        }
    }

    /**
     * 执行数据同步
     */
    private suspend fun performSync() {
        Log.d(TAG, "开始执行数据同步")
        val startTime = System.currentTimeMillis()
        
        var locationSyncCount = 0
        var trajectorySyncCount = 0
        var chatSyncCount = 0
        var errorCount = 0

        try {
            // 同步位置数据
            locationSyncCount = syncLocationData()
            
            // 同步轨迹数据
            trajectorySyncCount = syncTrajectoryData()
            
            // 同步聊天数据
            chatSyncCount = syncChatData()
            
            _syncState.value = SyncState.SUCCESS
            
        } catch (e: Exception) {
            Log.e(TAG, "数据同步失败", e)
            errorCount++
            _syncState.value = SyncState.ERROR
        }

        // 更新同步统计
        val duration = System.currentTimeMillis() - startTime
        _syncStats.value = _syncStats.value.copy(
            lastSyncTime = System.currentTimeMillis(),
            totalSyncCount = _syncStats.value.totalSyncCount + 1,
            locationSyncCount = _syncStats.value.locationSyncCount + locationSyncCount,
            trajectorySyncCount = _syncStats.value.trajectorySyncCount + trajectorySyncCount,
            chatSyncCount = _syncStats.value.chatSyncCount + chatSyncCount,
            errorCount = _syncStats.value.errorCount + errorCount,
            lastSyncDuration = duration
        )

        Log.d(TAG, "数据同步完成 - 位置: $locationSyncCount, 轨迹: $trajectorySyncCount, 聊天: $chatSyncCount, 耗时: ${duration}ms")
    }

    /**
     * 同步位置数据
     */
    private suspend fun syncLocationData(): Int {
        return try {
            val unsyncedPoints = locationRepository.getUnsyncedLocationPoints()
            if (unsyncedPoints.isEmpty()) {
                Log.d(TAG, "没有未同步的位置数据")
                return 0
            }

            Log.d(TAG, "开始同步 ${unsyncedPoints.size} 个位置点")
            val response = apiClient.uploadLocationPoints(unsyncedPoints)
            
            if (response.success) {
                // 标记为已同步
                locationRepository.markLocationPointsAsSynced(unsyncedPoints.map { it.id })
                Log.d(TAG, "位置数据同步成功")
                unsyncedPoints.size
            } else {
                Log.e(TAG, "位置数据同步失败: ${response.message}")
                0
            }
        } catch (e: Exception) {
            Log.e(TAG, "位置数据同步异常", e)
            0
        }
    }

    /**
     * 同步轨迹数据
     */
    private suspend fun syncTrajectoryData(): Int {
        return try {
            val result = locationRepository.syncUnsyncedTrajectories()
            result.getOrElse { 
                Log.e(TAG, "轨迹数据同步失败")
                0 
            }
        } catch (e: Exception) {
            Log.e(TAG, "轨迹数据同步异常", e)
            0
        }
    }

    /**
     * 同步聊天数据
     */
    private suspend fun syncChatData(): Int {
        return try {
            var syncCount = 0
            
            // 同步对话
            val conversationResult = chatRepository.syncConversations()
            if (conversationResult.isSuccess) {
                syncCount++
                Log.d(TAG, "对话数据同步成功")
            } else {
                Log.e(TAG, "对话数据同步失败")
            }
            
            // 同步消息（这里可以根据需要实现具体逻辑）
            // val messageResult = chatRepository.syncMessages()
            
            syncCount
        } catch (e: Exception) {
            Log.e(TAG, "聊天数据同步异常", e)
            0
        }
    }

    /**
     * 同步状态枚举
     */
    enum class SyncState {
        IDLE,           // 空闲
        RUNNING,        // 运行中
        SUCCESS,        // 成功
        ERROR,          // 错误
        WAITING_NETWORK // 等待网络
    }

    /**
     * 同步统计数据
     */
    data class SyncStats(
        val lastSyncTime: Long = 0,
        val totalSyncCount: Int = 0,
        val locationSyncCount: Int = 0,
        val trajectorySyncCount: Int = 0,
        val chatSyncCount: Int = 0,
        val errorCount: Int = 0,
        val lastSyncDuration: Long = 0
    )
}