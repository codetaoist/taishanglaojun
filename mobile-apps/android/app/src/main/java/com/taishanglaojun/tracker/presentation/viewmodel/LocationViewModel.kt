package com.taishanglaojun.tracker.presentation.viewmodel

import android.app.Application
import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.ServiceConnection
import android.os.IBinder
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import com.taishanglaojun.tracker.data.repository.LocationRepository
import com.taishanglaojun.tracker.data.service.LocationTrackingService
import com.taishanglaojun.tracker.utils.PermissionUtils
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch
import kotlinx.datetime.LocalDate
import javax.inject.Inject

/**
 * 位置追踪ViewModel
 * 管理位置追踪相关的业务逻辑和UI状态
 */
@HiltViewModel
class LocationViewModel @Inject constructor(
    application: Application,
    private val locationRepository: LocationRepository
) : AndroidViewModel(application) {

    private val context = getApplication<Application>()
    
    // 服务连接
    private var locationService: LocationTrackingService? = null
    private var isServiceBound = false
    
    private val serviceConnection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, service: IBinder?) {
            val binder = service as LocationTrackingService.LocationServiceBinder
            locationService = binder.getService()
            isServiceBound = true
            
            // 监听服务状态
            viewModelScope.launch {
                locationService?.let { service ->
                    combine(
                        service.isTracking,
                        service.isPaused,
                        service.currentLocation,
                        service.totalDistance,
                        service.pointCount
                    ) { isTracking, isPaused, location, distance, count ->
                        TrackingState(isTracking, isPaused, location, distance, count)
                    }.collect { state ->
                        _isTracking.value = state.isTracking
                        _isPaused.value = state.isPaused
                        _currentLocation.value = state.currentLocation
                        _totalDistance.value = state.totalDistance
                        _pointCount.value = state.pointCount
                    }
                }
            }
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            locationService = null
            isServiceBound = false
        }
    }

    // UI状态
    private val _uiState = MutableStateFlow(LocationUiState())
    val uiState: StateFlow<LocationUiState> = _uiState.asStateFlow()

    // 追踪状态
    private val _isTracking = MutableStateFlow(false)
    val isTracking: StateFlow<Boolean> = _isTracking.asStateFlow()

    private val _isPaused = MutableStateFlow(false)
    val isPaused: StateFlow<Boolean> = _isPaused.asStateFlow()

    private val _currentLocation = MutableStateFlow<LocationPoint?>(null)
    val currentLocation: StateFlow<LocationPoint?> = _currentLocation.asStateFlow()

    private val _totalDistance = MutableStateFlow(0f)
    val totalDistance: StateFlow<Float> = _totalDistance.asStateFlow()

    private val _pointCount = MutableStateFlow(0)
    val pointCount: StateFlow<Int> = _pointCount.asStateFlow()

    // 权限状态
    private val _hasLocationPermission = MutableStateFlow(false)
    val hasLocationPermission: StateFlow<Boolean> = _hasLocationPermission.asStateFlow()

    private val _hasBackgroundLocationPermission = MutableStateFlow(false)
    val hasBackgroundLocationPermission: StateFlow<Boolean> = _hasBackgroundLocationPermission.asStateFlow()

    // 轨迹数据
    private val _trajectories = MutableStateFlow<List<Trajectory>>(emptyList())
    val trajectories: StateFlow<List<Trajectory>> = _trajectories.asStateFlow()

    private val _currentTrajectory = MutableStateFlow<Trajectory?>(null)
    val currentTrajectory: StateFlow<Trajectory?> = _currentTrajectory.asStateFlow()

    // 错误状态
    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage.asStateFlow()

    init {
        checkPermissions()
        loadTrajectories()
    }

    override fun onCleared() {
        super.onCleared()
        if (isServiceBound) {
            context.unbindService(serviceConnection)
        }
    }

    /**
     * 检查权限状态
     */
    fun checkPermissions() {
        _hasLocationPermission.value = PermissionUtils.hasLocationPermission(context)
        _hasBackgroundLocationPermission.value = PermissionUtils.hasBackgroundLocationPermission(context)
    }

    /**
     * 开始位置追踪
     */
    fun startTracking(trajectoryName: String = "") {
        if (!_hasLocationPermission.value) {
            _errorMessage.value = "需要位置权限才能开始追踪"
            return
        }

        viewModelScope.launch {
            try {
                val trajectory = locationRepository.startNewTrajectory(
                    name = trajectoryName.ifEmpty { "新轨迹" },
                    userId = getCurrentUserId()
                )
                _currentTrajectory.value = trajectory

                // 启动服务
                val intent = Intent(context, LocationTrackingService::class.java).apply {
                    action = LocationTrackingService.ACTION_START_TRACKING
                    putExtra("trajectory_id", trajectory.id)
                }
                context.startForegroundService(intent)

                // 绑定服务
                val bindIntent = Intent(context, LocationTrackingService::class.java)
                context.bindService(bindIntent, serviceConnection, Context.BIND_AUTO_CREATE)

                _uiState.value = _uiState.value.copy(isLoading = false)
            } catch (e: Exception) {
                _errorMessage.value = "启动追踪失败: ${e.message}"
                _uiState.value = _uiState.value.copy(isLoading = false)
            }
        }
    }

    /**
     * 停止位置追踪
     */
    fun stopTracking() {
        viewModelScope.launch {
            try {
                // 停止服务
                val intent = Intent(context, LocationTrackingService::class.java).apply {
                    action = LocationTrackingService.ACTION_STOP_TRACKING
                }
                context.startService(intent)

                // 解绑服务
                if (isServiceBound) {
                    context.unbindService(serviceConnection)
                    isServiceBound = false
                }

                // 完成轨迹记录
                locationRepository.finishCurrentTrajectory()
                _currentTrajectory.value = null

                // 重新加载轨迹列表
                loadTrajectories()

                _uiState.value = _uiState.value.copy(isLoading = false)
            } catch (e: Exception) {
                _errorMessage.value = "停止追踪失败: ${e.message}"
                _uiState.value = _uiState.value.copy(isLoading = false)
            }
        }
    }

    /**
     * 暂停位置追踪
     */
    fun pauseTracking() {
        val intent = Intent(context, LocationTrackingService::class.java).apply {
            action = LocationTrackingService.ACTION_PAUSE_TRACKING
        }
        context.startService(intent)

        viewModelScope.launch {
            locationRepository.pauseCurrentTrajectory()
        }
    }

    /**
     * 恢复位置追踪
     */
    fun resumeTracking() {
        val intent = Intent(context, LocationTrackingService::class.java).apply {
            action = LocationTrackingService.ACTION_RESUME_TRACKING
        }
        context.startService(intent)

        viewModelScope.launch {
            locationRepository.resumeCurrentTrajectory()
        }
    }

    /**
     * 加载轨迹列表
     */
    fun loadTrajectories() {
        viewModelScope.launch {
            try {
                locationRepository.getTrajectories().collect { trajectoryList ->
                    _trajectories.value = trajectoryList
                }
            } catch (e: Exception) {
                _errorMessage.value = "加载轨迹失败: ${e.message}"
            }
        }
    }

    /**
     * 根据日期加载轨迹
     */
    fun loadTrajectoriesByDate(date: LocalDate) {
        viewModelScope.launch {
            try {
                locationRepository.getTrajectoriesByDate(date).collect { trajectoryList ->
                    _trajectories.value = trajectoryList
                }
            } catch (e: Exception) {
                _errorMessage.value = "加载轨迹失败: ${e.message}"
            }
        }
    }

    /**
     * 删除轨迹
     */
    fun deleteTrajectory(trajectoryId: String) {
        viewModelScope.launch {
            try {
                locationRepository.deleteTrajectory(trajectoryId)
                loadTrajectories()
            } catch (e: Exception) {
                _errorMessage.value = "删除轨迹失败: ${e.message}"
            }
        }
    }

    /**
     * 导出轨迹
     */
    fun exportTrajectory(trajectoryId: String, format: ExportFormat) {
        viewModelScope.launch {
            try {
                _uiState.value = _uiState.value.copy(isLoading = true)
                
                val result = when (format) {
                    ExportFormat.GPX -> locationRepository.exportTrajectoryToGpx(trajectoryId)
                    ExportFormat.JSON -> locationRepository.exportTrajectoryToJson(trajectoryId)
                }
                
                result.onSuccess { content ->
                    // 保存文件或分享
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        exportContent = content
                    )
                }.onFailure { error ->
                    _errorMessage.value = "导出失败: ${error.message}"
                    _uiState.value = _uiState.value.copy(isLoading = false)
                }
            } catch (e: Exception) {
                _errorMessage.value = "导出失败: ${e.message}"
                _uiState.value = _uiState.value.copy(isLoading = false)
            }
        }
    }

    /**
     * 同步轨迹到服务器
     */
    fun syncTrajectories() {
        viewModelScope.launch {
            try {
                _uiState.value = _uiState.value.copy(isLoading = true)
                
                val result = locationRepository.syncUnsyncedTrajectories()
                result.onSuccess { count ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        syncMessage = "成功同步 $count 条轨迹"
                    )
                }.onFailure { error ->
                    _errorMessage.value = "同步失败: ${error.message}"
                    _uiState.value = _uiState.value.copy(isLoading = false)
                }
            } catch (e: Exception) {
                _errorMessage.value = "同步失败: ${e.message}"
                _uiState.value = _uiState.value.copy(isLoading = false)
            }
        }
    }

    /**
     * 清除错误消息
     */
    fun clearErrorMessage() {
        _errorMessage.value = null
    }

    /**
     * 清除同步消息
     */
    fun clearSyncMessage() {
        _uiState.value = _uiState.value.copy(syncMessage = null)
    }

    /**
     * 清除导出内容
     */
    fun clearExportContent() {
        _uiState.value = _uiState.value.copy(exportContent = null)
    }

    /**
     * 获取当前用户ID
     */
    private fun getCurrentUserId(): String {
        // 从SharedPreferences或其他地方获取用户ID
        return "current_user_id"
    }

    /**
     * 获取轨迹的位置点
     */
    fun getLocationPoints(trajectoryId: String) = locationRepository.getLocationPointsByTrajectoryId(trajectoryId)

    /**
     * 获取统计信息
     */
    suspend fun getStatistics() = locationRepository.getStatistics()
}

/**
 * 追踪状态数据类
 */
private data class TrackingState(
    val isTracking: Boolean,
    val isPaused: Boolean,
    val currentLocation: LocationPoint?,
    val totalDistance: Float,
    val pointCount: Int
)

/**
 * UI状态数据类
 */
data class LocationUiState(
    val isLoading: Boolean = false,
    val exportContent: String? = null,
    val syncMessage: String? = null
)

/**
 * 导出格式枚举
 */
enum class ExportFormat {
    GPX,
    JSON
}