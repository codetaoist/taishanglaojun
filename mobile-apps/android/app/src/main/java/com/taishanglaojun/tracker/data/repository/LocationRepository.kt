package com.taishanglaojun.tracker.data.repository

import com.taishanglaojun.tracker.data.local.dao.LocationPointDao
import com.taishanglaojun.tracker.data.local.dao.TrajectoryDao
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import com.taishanglaojun.tracker.data.remote.api.LocationApiService
import com.taishanglaojun.tracker.data.remote.dto.TrajectoryUploadRequest
import com.taishanglaojun.tracker.utils.CryptoUtils
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.datetime.LocalDate
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 位置数据仓库
 * 负责管理本地和远程位置数据
 */
@Singleton
class LocationRepository @Inject constructor(
    private val locationPointDao: LocationPointDao,
    private val trajectoryDao: TrajectoryDao,
    private val locationApiService: LocationApiService,
    private val cryptoUtils: CryptoUtils
) {

    private var currentTrajectory: Trajectory? = null

    /**
     * 开始新的轨迹记录
     */
    suspend fun startNewTrajectory(
        name: String = "",
        description: String = "",
        userId: String = ""
    ): Trajectory {
        val trajectory = Trajectory.createNew(name, description, userId)
        trajectoryDao.insertTrajectory(trajectory)
        currentTrajectory = trajectory
        return trajectory
    }

    /**
     * 完成当前轨迹记录
     */
    suspend fun finishCurrentTrajectory(): Trajectory? {
        return currentTrajectory?.let { trajectory ->
            val points = locationPointDao.getPointsByTrajectoryId(trajectory.id).first()
            val updatedTrajectory = trajectory.updateStats(points).finish()
            trajectoryDao.updateTrajectory(updatedTrajectory)
            currentTrajectory = null
            updatedTrajectory
        }
    }

    /**
     * 暂停当前轨迹记录
     */
    suspend fun pauseCurrentTrajectory(): Trajectory? {
        return currentTrajectory?.let { trajectory ->
            val pausedTrajectory = trajectory.pause()
            trajectoryDao.updateTrajectory(pausedTrajectory)
            currentTrajectory = pausedTrajectory
            pausedTrajectory
        }
    }

    /**
     * 恢复当前轨迹记录
     */
    suspend fun resumeCurrentTrajectory(): Trajectory? {
        return currentTrajectory?.let { trajectory ->
            val resumedTrajectory = trajectory.resume()
            trajectoryDao.updateTrajectory(resumedTrajectory)
            currentTrajectory = resumedTrajectory
            resumedTrajectory
        }
    }

    /**
     * 保存位置点
     */
    suspend fun saveLocationPoint(locationPoint: LocationPoint) {
        // 加密敏感位置数据（可选）
        val encryptedPoint = if (shouldEncryptLocation(locationPoint)) {
            encryptLocationPoint(locationPoint)
        } else {
            locationPoint
        }
        
        locationPointDao.insertLocationPoint(encryptedPoint)
        
        // 更新当前轨迹统计
        currentTrajectory?.let { trajectory ->
            val points = locationPointDao.getPointsByTrajectoryId(trajectory.id).first()
            val updatedTrajectory = trajectory.updateStats(points)
            trajectoryDao.updateTrajectory(updatedTrajectory)
            currentTrajectory = updatedTrajectory
        }
    }

    /**
     * 批量保存位置点
     */
    suspend fun saveLocationPoints(locationPoints: List<LocationPoint>) {
        val encryptedPoints = locationPoints.map { point ->
            if (shouldEncryptLocation(point)) {
                encryptLocationPoint(point)
            } else {
                point
            }
        }
        locationPointDao.insertLocationPoints(encryptedPoints)
    }

    /**
     * 获取轨迹列表
     */
    fun getTrajectories(): Flow<List<Trajectory>> {
        return trajectoryDao.getAllTrajectories()
    }

    /**
     * 根据日期获取轨迹
     */
    fun getTrajectoriesByDate(date: LocalDate): Flow<List<Trajectory>> {
        val startOfDay = date.atStartOfDay()
        val endOfDay = date.atTime(23, 59, 59)
        
        return trajectoryDao.getTrajectoriesByDateRange(
            startOfDay.toEpochMilliseconds(),
            endOfDay.toEpochMilliseconds()
        )
    }

    /**
     * 根据日期范围获取轨迹
     */
    fun getTrajectoriesByDateRange(startDate: LocalDate, endDate: LocalDate): Flow<List<Trajectory>> {
        val startTime = startDate.atStartOfDay().toEpochMilliseconds()
        val endTime = endDate.atTime(23, 59, 59).toEpochMilliseconds()
        
        return trajectoryDao.getTrajectoriesByDateRange(startTime, endTime)
    }

    /**
     * 根据ID获取轨迹
     */
    suspend fun getTrajectoryById(trajectoryId: String): Trajectory? {
        return trajectoryDao.getTrajectoryById(trajectoryId)
    }

    /**
     * 获取轨迹的位置点
     */
    fun getLocationPointsByTrajectoryId(trajectoryId: String): Flow<List<LocationPoint>> {
        return locationPointDao.getPointsByTrajectoryId(trajectoryId).map { points ->
            points.map { point ->
                if (isLocationEncrypted(point)) {
                    decryptLocationPoint(point)
                } else {
                    point
                }
            }
        }
    }

    /**
     * 删除轨迹
     */
    suspend fun deleteTrajectory(trajectoryId: String) {
        locationPointDao.deletePointsByTrajectoryId(trajectoryId)
        trajectoryDao.deleteTrajectory(trajectoryId)
    }

    /**
     * 清除所有数据
     */
    suspend fun clearAllData() {
        locationPointDao.deleteAllPoints()
        trajectoryDao.deleteAllTrajectories()
        currentTrajectory = null
    }

    /**
     * 获取统计信息
     */
    suspend fun getStatistics(): TrackingStatistics {
        val totalTrajectories = trajectoryDao.getTotalTrajectoryCount()
        val totalDistance = trajectoryDao.getTotalDistance()
        val totalDuration = trajectoryDao.getTotalDuration()
        val totalPoints = locationPointDao.getTotalPointCount()
        
        return TrackingStatistics(
            totalTrajectories = totalTrajectories,
            totalDistance = totalDistance,
            totalDuration = totalDuration,
            totalPoints = totalPoints
        )
    }

    /**
     * 获取未同步的轨迹
     */
    suspend fun getUnsyncedTrajectories(): List<Trajectory> {
        return trajectoryDao.getUnsyncedTrajectories()
    }

    /**
     * 上传轨迹到服务器
     */
    suspend fun uploadTrajectory(trajectory: Trajectory): Result<Unit> {
        return try {
            val points = locationPointDao.getPointsByTrajectoryId(trajectory.id).first()
            val decryptedPoints = points.map { point ->
                if (isLocationEncrypted(point)) {
                    decryptLocationPoint(point)
                } else {
                    point
                }
            }
            
            val request = TrajectoryUploadRequest(
                trajectory = trajectory,
                locationPoints = decryptedPoints
            )
            
            locationApiService.uploadTrajectory(request)
            
            // 标记为已同步
            val syncedTrajectory = trajectory.copy(synced = true)
            trajectoryDao.updateTrajectory(syncedTrajectory)
            
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * 批量上传未同步的轨迹
     */
    suspend fun syncUnsyncedTrajectories(): Result<Int> {
        return try {
            val unsyncedTrajectories = getUnsyncedTrajectories()
            var successCount = 0
            
            for (trajectory in unsyncedTrajectories) {
                uploadTrajectory(trajectory).onSuccess {
                    successCount++
                }
            }
            
            Result.success(successCount)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * 从服务器下载轨迹
     */
    suspend fun downloadTrajectories(userId: String): Result<List<Trajectory>> {
        return try {
            val trajectories = locationApiService.getUserTrajectories(userId)
            
            // 保存到本地数据库
            for (trajectory in trajectories) {
                trajectoryDao.insertTrajectory(trajectory.copy(synced = true))
            }
            
            Result.success(trajectories)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * 导出轨迹为GPX格式
     */
    suspend fun exportTrajectoryToGpx(trajectoryId: String): Result<String> {
        return try {
            val trajectory = getTrajectoryById(trajectoryId)
                ?: return Result.failure(Exception("轨迹不存在"))
            
            val points = getLocationPointsByTrajectoryId(trajectoryId).first()
            val gpxContent = generateGpxContent(trajectory, points)
            
            Result.success(gpxContent)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * 导出轨迹为JSON格式
     */
    suspend fun exportTrajectoryToJson(trajectoryId: String): Result<String> {
        return try {
            val trajectory = getTrajectoryById(trajectoryId)
                ?: return Result.failure(Exception("轨迹不存在"))
            
            val points = getLocationPointsByTrajectoryId(trajectoryId).first()
            val jsonContent = generateJsonContent(trajectory, points)
            
            Result.success(jsonContent)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * 判断是否需要加密位置数据
     */
    private fun shouldEncryptLocation(locationPoint: LocationPoint): Boolean {
        // 可以根据业务需求决定是否加密
        // 例如：敏感区域、高精度位置等
        return locationPoint.accuracy <= 10f // 高精度位置需要加密
    }

    /**
     * 加密位置点
     */
    private fun encryptLocationPoint(locationPoint: LocationPoint): LocationPoint {
        return try {
            val encryptedLat = cryptoUtils.encrypt(locationPoint.latitude.toString())
            val encryptedLng = cryptoUtils.encrypt(locationPoint.longitude.toString())
            
            locationPoint.copy(
                latitude = encryptedLat.hashCode().toDouble(), // 使用哈希值作为占位符
                longitude = encryptedLng.hashCode().toDouble()
                // 实际实现中应该有专门的字段存储加密数据
            )
        } catch (e: Exception) {
            locationPoint // 加密失败时返回原数据
        }
    }

    /**
     * 解密位置点
     */
    private fun decryptLocationPoint(locationPoint: LocationPoint): LocationPoint {
        return try {
            // 实际实现中应该从加密字段中解密数据
            locationPoint
        } catch (e: Exception) {
            locationPoint // 解密失败时返回原数据
        }
    }

    /**
     * 判断位置是否已加密
     */
    private fun isLocationEncrypted(locationPoint: LocationPoint): Boolean {
        // 实际实现中应该有标识字段
        return false
    }

    /**
     * 生成GPX内容
     */
    private fun generateGpxContent(trajectory: Trajectory, points: List<LocationPoint>): String {
        val gpxBuilder = StringBuilder()
        gpxBuilder.append("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
        gpxBuilder.append("<gpx version=\"1.1\" creator=\"TaishanglaojunTracker\">\n")
        gpxBuilder.append("  <trk>\n")
        gpxBuilder.append("    <name>${trajectory.name}</name>\n")
        gpxBuilder.append("    <desc>${trajectory.description}</desc>\n")
        gpxBuilder.append("    <trkseg>\n")
        
        for (point in points) {
            gpxBuilder.append("      <trkpt lat=\"${point.latitude}\" lon=\"${point.longitude}\">\n")
            gpxBuilder.append("        <ele>${point.altitude}</ele>\n")
            gpxBuilder.append("        <time>${point.getFormattedTimestamp()}</time>\n")
            gpxBuilder.append("      </trkpt>\n")
        }
        
        gpxBuilder.append("    </trkseg>\n")
        gpxBuilder.append("  </trk>\n")
        gpxBuilder.append("</gpx>")
        
        return gpxBuilder.toString()
    }

    /**
     * 生成JSON内容
     */
    private fun generateJsonContent(trajectory: Trajectory, points: List<LocationPoint>): String {
        // 使用Gson或其他JSON库生成JSON内容
        return "{\"trajectory\": $trajectory, \"points\": $points}"
    }
}

/**
 * 追踪统计数据
 */
data class TrackingStatistics(
    val totalTrajectories: Int,
    val totalDistance: Float,
    val totalDuration: Long,
    val totalPoints: Int
) {
    fun getFormattedDistance(): String {
        return when {
            totalDistance >= 1000 -> String.format("%.2f km", totalDistance / 1000)
            totalDistance >= 1 -> String.format("%.0f m", totalDistance)
            else -> "0 m"
        }
    }
    
    fun getFormattedDuration(): String {
        val hours = totalDuration / (1000 * 60 * 60)
        val minutes = (totalDuration % (1000 * 60 * 60)) / (1000 * 60)
        
        return when {
            hours > 0 -> String.format("%d小时%d分钟", hours, minutes)
            minutes > 0 -> String.format("%d分钟", minutes)
            else -> "0分钟"
        }
    }
}