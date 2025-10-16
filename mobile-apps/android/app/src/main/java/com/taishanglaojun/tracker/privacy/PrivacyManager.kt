package com.taishanglaojun.tracker.privacy

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.SharedPreferences
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import java.security.MessageDigest
import java.text.SimpleDateFormat
import java.util.*
import kotlin.math.round

/**
 * 隐私管理器
 * 负责处理权限申请、用户同意和数据脱敏
 */
class PrivacyManager private constructor(private val context: Context) {
    
    companion object {
        @Volatile
        private var INSTANCE: PrivacyManager? = null
        
        // 权限请求码
        const val REQUEST_LOCATION_PERMISSION = 1001
        const val REQUEST_BACKGROUND_LOCATION_PERMISSION = 1002
        
        // SharedPreferences键
        private const val PREFS_NAME = "privacy_prefs"
        private const val KEY_PRIVACY_POLICY_ACCEPTED = "privacy_policy_accepted"
        private const val KEY_LOCATION_PERMISSION_GRANTED = "location_permission_granted"
        private const val KEY_BACKGROUND_LOCATION_GRANTED = "background_location_granted"
        private const val KEY_DATA_COLLECTION_CONSENT = "data_collection_consent"
        private const val KEY_DATA_SHARING_CONSENT = "data_sharing_consent"
        private const val KEY_ANALYTICS_CONSENT = "analytics_consent"
        private const val KEY_FIRST_LAUNCH = "first_launch"
        
        fun getInstance(context: Context): PrivacyManager {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: PrivacyManager(context.applicationContext).also { INSTANCE = it }
            }
        }
    }
    
    private val sharedPreferences: SharedPreferences = 
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    
    /**
     * 隐私同意状态
     */
    data class PrivacyConsent(
        val privacyPolicyAccepted: Boolean = false,
        val dataCollectionConsent: Boolean = false,
        val dataSharingConsent: Boolean = false,
        val analyticsConsent: Boolean = false
    )
    
    /**
     * 权限状态
     */
    data class PermissionStatus(
        val locationPermission: Boolean = false,
        val backgroundLocationPermission: Boolean = false,
        val preciseLocationPermission: Boolean = false
    )
    
    /**
     * 数据脱敏级别
     */
    enum class DataSensitivityLevel {
        NONE,       // 不脱敏
        LOW,        // 低级脱敏（保留小数点后4位）
        MEDIUM,     // 中级脱敏（保留小数点后3位）
        HIGH        // 高级脱敏（保留小数点后2位）
    }
    
    /**
     * 检查是否首次启动
     */
    fun isFirstLaunch(): Boolean {
        return sharedPreferences.getBoolean(KEY_FIRST_LAUNCH, true)
    }
    
    /**
     * 标记首次启动完成
     */
    fun markFirstLaunchCompleted() {
        sharedPreferences.edit()
            .putBoolean(KEY_FIRST_LAUNCH, false)
            .apply()
    }
    
    /**
     * 检查隐私政策是否已接受
     */
    fun isPrivacyPolicyAccepted(): Boolean {
        return sharedPreferences.getBoolean(KEY_PRIVACY_POLICY_ACCEPTED, false)
    }
    
    /**
     * 设置隐私政策接受状态
     */
    fun setPrivacyPolicyAccepted(accepted: Boolean) {
        sharedPreferences.edit()
            .putBoolean(KEY_PRIVACY_POLICY_ACCEPTED, accepted)
            .apply()
    }
    
    /**
     * 获取隐私同意状态
     */
    fun getPrivacyConsent(): PrivacyConsent {
        return PrivacyConsent(
            privacyPolicyAccepted = sharedPreferences.getBoolean(KEY_PRIVACY_POLICY_ACCEPTED, false),
            dataCollectionConsent = sharedPreferences.getBoolean(KEY_DATA_COLLECTION_CONSENT, false),
            dataSharingConsent = sharedPreferences.getBoolean(KEY_DATA_SHARING_CONSENT, false),
            analyticsConsent = sharedPreferences.getBoolean(KEY_ANALYTICS_CONSENT, false)
        )
    }
    
    /**
     * 设置隐私同意状态
     */
    fun setPrivacyConsent(consent: PrivacyConsent) {
        sharedPreferences.edit()
            .putBoolean(KEY_PRIVACY_POLICY_ACCEPTED, consent.privacyPolicyAccepted)
            .putBoolean(KEY_DATA_COLLECTION_CONSENT, consent.dataCollectionConsent)
            .putBoolean(KEY_DATA_SHARING_CONSENT, consent.dataSharingConsent)
            .putBoolean(KEY_ANALYTICS_CONSENT, consent.analyticsConsent)
            .apply()
    }
    
    /**
     * 检查位置权限状态
     */
    fun checkLocationPermissions(): PermissionStatus {
        val locationPermission = ContextCompat.checkSelfPermission(
            context, Manifest.permission.ACCESS_FINE_LOCATION
        ) == PackageManager.PERMISSION_GRANTED
        
        val coarseLocationPermission = ContextCompat.checkSelfPermission(
            context, Manifest.permission.ACCESS_COARSE_LOCATION
        ) == PackageManager.PERMISSION_GRANTED
        
        val backgroundLocationPermission = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            ContextCompat.checkSelfPermission(
                context, Manifest.permission.ACCESS_BACKGROUND_LOCATION
            ) == PackageManager.PERMISSION_GRANTED
        } else {
            locationPermission || coarseLocationPermission
        }
        
        return PermissionStatus(
            locationPermission = locationPermission || coarseLocationPermission,
            backgroundLocationPermission = backgroundLocationPermission,
            preciseLocationPermission = locationPermission
        )
    }
    
    /**
     * 请求位置权限
     */
    fun requestLocationPermissions(activity: Activity) {
        val permissions = mutableListOf<String>()
        
        if (ContextCompat.checkSelfPermission(
                context, Manifest.permission.ACCESS_FINE_LOCATION
            ) != PackageManager.PERMISSION_GRANTED
        ) {
            permissions.add(Manifest.permission.ACCESS_FINE_LOCATION)
        }
        
        if (ContextCompat.checkSelfPermission(
                context, Manifest.permission.ACCESS_COARSE_LOCATION
            ) != PackageManager.PERMISSION_GRANTED
        ) {
            permissions.add(Manifest.permission.ACCESS_COARSE_LOCATION)
        }
        
        if (permissions.isNotEmpty()) {
            ActivityCompat.requestPermissions(
                activity,
                permissions.toTypedArray(),
                REQUEST_LOCATION_PERMISSION
            )
        }
    }
    
    /**
     * 请求后台位置权限
     */
    fun requestBackgroundLocationPermission(activity: Activity) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            if (ContextCompat.checkSelfPermission(
                    context, Manifest.permission.ACCESS_BACKGROUND_LOCATION
                ) != PackageManager.PERMISSION_GRANTED
            ) {
                ActivityCompat.requestPermissions(
                    activity,
                    arrayOf(Manifest.permission.ACCESS_BACKGROUND_LOCATION),
                    REQUEST_BACKGROUND_LOCATION_PERMISSION
                )
            }
        }
    }
    
    /**
     * 检查是否应该显示权限说明
     */
    fun shouldShowLocationPermissionRationale(activity: Activity): Boolean {
        return ActivityCompat.shouldShowRequestPermissionRationale(
            activity, Manifest.permission.ACCESS_FINE_LOCATION
        ) || ActivityCompat.shouldShowRequestPermissionRationale(
            activity, Manifest.permission.ACCESS_COARSE_LOCATION
        )
    }
    
    /**
     * 检查是否应该显示后台位置权限说明
     */
    fun shouldShowBackgroundLocationPermissionRationale(activity: Activity): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            ActivityCompat.shouldShowRequestPermissionRationale(
                activity, Manifest.permission.ACCESS_BACKGROUND_LOCATION
            )
        } else {
            false
        }
    }
    
    /**
     * 对位置点进行数据脱敏
     */
    fun sanitizeLocationPoint(
        locationPoint: LocationPoint,
        level: DataSensitivityLevel = DataSensitivityLevel.MEDIUM
    ): LocationPoint {
        if (level == DataSensitivityLevel.NONE) {
            return locationPoint
        }
        
        val precision = when (level) {
            DataSensitivityLevel.LOW -> 4
            DataSensitivityLevel.MEDIUM -> 3
            DataSensitivityLevel.HIGH -> 2
            else -> 6
        }
        
        val multiplier = Math.pow(10.0, precision.toDouble())
        
        return locationPoint.copy(
            latitude = round(locationPoint.latitude * multiplier) / multiplier,
            longitude = round(locationPoint.longitude * multiplier) / multiplier,
            altitude = locationPoint.altitude?.let { round(it * 10) / 10 }, // 保留1位小数
            speed = locationPoint.speed?.let { round(it * 10) / 10 }, // 保留1位小数
            bearing = locationPoint.bearing?.let { round(it).toFloat() } // 取整
        )
    }
    
    /**
     * 对轨迹进行数据脱敏
     */
    fun sanitizeTrajectory(
        trajectory: Trajectory,
        level: DataSensitivityLevel = DataSensitivityLevel.MEDIUM
    ): Trajectory {
        if (level == DataSensitivityLevel.NONE) {
            return trajectory
        }
        
        val precision = when (level) {
            DataSensitivityLevel.LOW -> 4
            DataSensitivityLevel.MEDIUM -> 3
            DataSensitivityLevel.HIGH -> 2
            else -> 6
        }
        
        val multiplier = Math.pow(10.0, precision.toDouble())
        
        return trajectory.copy(
            name = if (level == DataSensitivityLevel.HIGH) "轨迹${trajectory.id.hashCode().toString().takeLast(4)}" else trajectory.name,
            description = if (level == DataSensitivityLevel.HIGH) "" else trajectory.description,
            minLatitude = trajectory.minLatitude?.let { round(it * multiplier) / multiplier },
            maxLatitude = trajectory.maxLatitude?.let { round(it * multiplier) / multiplier },
            minLongitude = trajectory.minLongitude?.let { round(it * multiplier) / multiplier },
            maxLongitude = trajectory.maxLongitude?.let { round(it * multiplier) / multiplier },
            totalDistance = round(trajectory.totalDistance * 10) / 10, // 保留1位小数
            averageSpeed = trajectory.averageSpeed?.let { round(it * 10) / 10 }, // 保留1位小数
            maxSpeed = trajectory.maxSpeed?.let { round(it * 10) / 10 } // 保留1位小数
        )
    }
    
    /**
     * 生成匿名用户ID
     */
    fun generateAnonymousUserId(): String {
        val deviceId = android.provider.Settings.Secure.getString(
            context.contentResolver,
            android.provider.Settings.Secure.ANDROID_ID
        )
        
        return hashString("$deviceId${System.currentTimeMillis()}")
    }
    
    /**
     * 对字符串进行哈希处理
     */
    private fun hashString(input: String): String {
        val bytes = MessageDigest.getInstance("SHA-256").digest(input.toByteArray())
        return bytes.joinToString("") { "%02x".format(it) }.take(16)
    }
    
    /**
     * 检查数据收集是否被允许
     */
    fun isDataCollectionAllowed(): Boolean {
        return getPrivacyConsent().dataCollectionConsent
    }
    
    /**
     * 检查数据分享是否被允许
     */
    fun isDataSharingAllowed(): Boolean {
        return getPrivacyConsent().dataSharingConsent
    }
    
    /**
     * 检查分析数据收集是否被允许
     */
    fun isAnalyticsAllowed(): Boolean {
        return getPrivacyConsent().analyticsConsent
    }
    
    /**
     * 获取隐私政策文本
     */
    fun getPrivacyPolicyText(): String {
        return """
            隐私政策
            
            1. 数据收集
            我们收集您的位置信息以提供轨迹跟踪服务。
            
            2. 数据使用
            您的位置数据仅用于：
            - 记录和显示您的移动轨迹
            - 计算距离、速度等统计信息
            - 提供轨迹分析功能
            
            3. 数据存储
            您的数据存储在本地设备上，我们不会将其上传到服务器，除非您明确同意。
            
            4. 数据分享
            我们不会与第三方分享您的个人位置数据，除非获得您的明确同意。
            
            5. 数据安全
            我们采用加密技术保护您的数据安全。
            
            6. 您的权利
            您可以随时：
            - 查看您的数据
            - 删除您的数据
            - 撤回同意
            - 导出您的数据
            
            最后更新：${SimpleDateFormat("yyyy年MM月dd日", Locale.getDefault()).format(Date())}
        """.trimIndent()
    }
    
    /**
     * 重置所有隐私设置
     */
    fun resetPrivacySettings() {
        sharedPreferences.edit().clear().apply()
    }
}