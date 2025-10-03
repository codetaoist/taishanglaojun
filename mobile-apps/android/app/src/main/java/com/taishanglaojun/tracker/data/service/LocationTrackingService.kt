package com.taishanglaojun.tracker.data.service

import android.Manifest
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.location.Location
import android.os.Binder
import android.os.Build
import android.os.IBinder
import android.os.Looper
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import com.google.android.gms.location.FusedLocationProviderClient
import com.google.android.gms.location.LocationCallback
import com.google.android.gms.location.LocationRequest
import com.google.android.gms.location.LocationResult
import com.google.android.gms.location.LocationServices
import com.google.android.gms.location.Priority
import com.taishanglaojun.tracker.R
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.repository.LocationRepository
import com.taishanglaojun.tracker.presentation.MainActivity
import com.taishanglaojun.tracker.utils.BatteryUtils
import com.taishanglaojun.tracker.utils.NetworkUtils
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * 位置追踪前台服务
 * 负责在后台持续获取用户位置信息
 */
@AndroidEntryPoint
class LocationTrackingService : Service() {

    companion object {
        private const val NOTIFICATION_ID = 1001
        private const val CHANNEL_ID = "location_tracking_channel"
        private const val CHANNEL_NAME = "位置追踪"
        
        // 位置更新间隔
        private const val LOCATION_UPDATE_INTERVAL = 5000L // 5秒
        private const val FASTEST_UPDATE_INTERVAL = 2000L // 2秒
        private const val MIN_DISPLACEMENT = 5f // 最小位移5米
        
        // Service Actions
        const val ACTION_START_TRACKING = "start_tracking"
        const val ACTION_STOP_TRACKING = "stop_tracking"
        const val ACTION_PAUSE_TRACKING = "pause_tracking"
        const val ACTION_RESUME_TRACKING = "resume_tracking"
    }

    @Inject
    lateinit var locationRepository: LocationRepository

    private lateinit var fusedLocationClient: FusedLocationProviderClient
    private lateinit var notificationManager: NotificationManager
    
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val binder = LocationServiceBinder()
    
    // 位置回调
    private val locationCallback = object : LocationCallback() {
        override fun onLocationResult(result: LocationResult) {
            super.onLocationResult(result)
            result.lastLocation?.let { location ->
                handleLocationUpdate(location)
            }
        }
    }
    
    // 服务状态
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
    
    private var currentTrajectoryId: String = ""
    private var lastLocation: Location? = null
    private var accumulatedDistance = 0f

    override fun onCreate() {
        super.onCreate()
        
        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)
        notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        
        createNotificationChannel()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START_TRACKING -> startTracking(intent.getStringExtra("trajectory_id") ?: "")
            ACTION_STOP_TRACKING -> stopTracking()
            ACTION_PAUSE_TRACKING -> pauseTracking()
            ACTION_RESUME_TRACKING -> resumeTracking()
        }
        
        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder = binder

    override fun onDestroy() {
        super.onDestroy()
        stopLocationUpdates()
        serviceScope.cancel()
    }

    /**
     * 开始位置追踪
     */
    private fun startTracking(trajectoryId: String) {
        if (_isTracking.value) return
        
        currentTrajectoryId = trajectoryId
        accumulatedDistance = 0f
        lastLocation = null
        
        _isTracking.value = true
        _isPaused.value = false
        _totalDistance.value = 0f
        _pointCount.value = 0
        
        startLocationUpdates()
        startForeground(NOTIFICATION_ID, createNotification())
        
        serviceScope.launch {
            locationRepository.startNewTrajectory(trajectoryId)
        }
    }

    /**
     * 停止位置追踪
     */
    private fun stopTracking() {
        if (!_isTracking.value) return
        
        _isTracking.value = false
        _isPaused.value = false
        
        stopLocationUpdates()
        stopForeground(true)
        
        serviceScope.launch {
            locationRepository.finishCurrentTrajectory()
        }
        
        stopSelf()
    }

    /**
     * 暂停位置追踪
     */
    private fun pauseTracking() {
        if (!_isTracking.value || _isPaused.value) return
        
        _isPaused.value = true
        stopLocationUpdates()
        
        // 更新通知显示暂停状态
        notificationManager.notify(NOTIFICATION_ID, createNotification())
    }

    /**
     * 恢复位置追踪
     */
    private fun resumeTracking() {
        if (!_isTracking.value || !_isPaused.value) return
        
        _isPaused.value = false
        startLocationUpdates()
        
        // 更新通知显示运行状态
        notificationManager.notify(NOTIFICATION_ID, createNotification())
    }

    /**
     * 开始位置更新
     */
    private fun startLocationUpdates() {
        if (!hasLocationPermission()) return
        
        val locationRequest = LocationRequest.Builder(Priority.PRIORITY_HIGH_ACCURACY, LOCATION_UPDATE_INTERVAL)
            .setMinUpdateIntervalMillis(FASTEST_UPDATE_INTERVAL)
            .setMinUpdateDistanceMeters(MIN_DISPLACEMENT)
            .build()
        
        try {
            fusedLocationClient.requestLocationUpdates(
                locationRequest,
                locationCallback,
                Looper.getMainLooper()
            )
        } catch (e: SecurityException) {
            // 权限被撤销
            stopTracking()
        }
    }

    /**
     * 停止位置更新
     */
    private fun stopLocationUpdates() {
        fusedLocationClient.removeLocationUpdates(locationCallback)
    }

    /**
     * 处理位置更新
     */
    private fun handleLocationUpdate(location: Location) {
        if (_isPaused.value) return
        
        val batteryLevel = BatteryUtils.getBatteryLevel(this)
        val networkType = NetworkUtils.getNetworkType(this)
        
        val locationPoint = LocationPoint.fromLocation(
            location = location,
            trajectoryId = currentTrajectoryId,
            batteryLevel = batteryLevel,
            networkType = networkType
        )
        
        // 验证位置点质量
        if (!isLocationValid(location)) return
        
        _currentLocation.value = locationPoint
        _pointCount.value = _pointCount.value + 1
        
        // 计算距离
        lastLocation?.let { last ->
            val distance = last.distanceTo(location)
            if (distance > MIN_DISPLACEMENT) {
                accumulatedDistance += distance
                _totalDistance.value = accumulatedDistance
            }
        }
        lastLocation = location
        
        // 保存位置点
        serviceScope.launch {
            locationRepository.saveLocationPoint(locationPoint)
        }
        
        // 更新通知
        notificationManager.notify(NOTIFICATION_ID, createNotification())
    }

    /**
     * 验证位置是否有效
     */
    private fun isLocationValid(location: Location): Boolean {
        // 检查精度
        if (location.accuracy > 100f) return false
        
        // 检查是否为模拟位置（可选）
        if (location.isFromMockProvider) return false
        
        // 检查时间戳是否合理
        val timeDiff = System.currentTimeMillis() - location.time
        if (timeDiff > 30000) return false // 超过30秒的位置数据
        
        return true
    }

    /**
     * 检查位置权限
     */
    private fun hasLocationPermission(): Boolean {
        return ActivityCompat.checkSelfPermission(
            this,
            Manifest.permission.ACCESS_FINE_LOCATION
        ) == PackageManager.PERMISSION_GRANTED
    }

    /**
     * 创建通知渠道
     */
    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                CHANNEL_NAME,
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "位置追踪服务通知"
                setShowBadge(false)
            }
            notificationManager.createNotificationChannel(channel)
        }
    }

    /**
     * 创建前台服务通知
     */
    private fun createNotification(): Notification {
        val intent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        val status = when {
            _isPaused.value -> "已暂停"
            _isTracking.value -> "追踪中"
            else -> "已停止"
        }
        
        val distance = _totalDistance.value
        val distanceText = if (distance >= 1000) {
            String.format("%.2f km", distance / 1000)
        } else {
            String.format("%.0f m", distance)
        }
        
        val contentText = "$status • $distanceText • ${_pointCount.value} 个点"
        
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("位置追踪")
            .setContentText(contentText)
            .setSmallIcon(R.drawable.ic_location)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setCategory(NotificationCompat.CATEGORY_SERVICE)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .addAction(createPauseResumeAction())
            .addAction(createStopAction())
            .build()
    }

    /**
     * 创建暂停/恢复操作
     */
    private fun createPauseResumeAction(): NotificationCompat.Action {
        val action = if (_isPaused.value) ACTION_RESUME_TRACKING else ACTION_PAUSE_TRACKING
        val title = if (_isPaused.value) "恢复" else "暂停"
        val icon = if (_isPaused.value) R.drawable.ic_play else R.drawable.ic_pause
        
        val intent = Intent(this, LocationTrackingService::class.java).apply {
            this.action = action
        }
        val pendingIntent = PendingIntent.getService(
            this, 1, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        return NotificationCompat.Action.Builder(icon, title, pendingIntent).build()
    }

    /**
     * 创建停止操作
     */
    private fun createStopAction(): NotificationCompat.Action {
        val intent = Intent(this, LocationTrackingService::class.java).apply {
            action = ACTION_STOP_TRACKING
        }
        val pendingIntent = PendingIntent.getService(
            this, 2, intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        return NotificationCompat.Action.Builder(
            R.drawable.ic_stop,
            "停止",
            pendingIntent
        ).build()
    }

    /**
     * Service Binder
     */
    inner class LocationServiceBinder : Binder() {
        fun getService(): LocationTrackingService = this@LocationTrackingService
    }
}