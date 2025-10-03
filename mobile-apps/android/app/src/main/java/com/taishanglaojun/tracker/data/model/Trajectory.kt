package com.taishanglaojun.tracker.data.model

import android.os.Parcelable
import androidx.room.Entity
import androidx.room.PrimaryKey
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize
import kotlinx.datetime.Instant
import kotlinx.datetime.LocalDate
import kotlinx.datetime.LocalDateTime
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime
import java.util.UUID
import kotlin.math.max
import kotlin.math.min

/**
 * 轨迹数据模型
 * 用于存储用户的移动轨迹信息
 */
@Entity(
    tableName = "trajectories",
    indices = [
        androidx.room.Index(value = ["startTime"]),
        androidx.room.Index(value = ["isRecording"])
    ]
)
@Parcelize
data class Trajectory(
    @PrimaryKey
    val id: String = UUID.randomUUID().toString(),
    
    @SerializedName("name")
    val name: String = "",
    
    @SerializedName("description")
    val description: String = "",
    
    @SerializedName("start_time")
    val startTime: Long = System.currentTimeMillis(),
    
    @SerializedName("end_time")
    val endTime: Long? = null,
    
    @SerializedName("is_recording")
    val isRecording: Boolean = true,
    
    @SerializedName("total_distance")
    val totalDistance: Float = 0f,
    
    @SerializedName("total_duration")
    val totalDuration: Long = 0L,
    
    @SerializedName("max_speed")
    val maxSpeed: Float = 0f,
    
    @SerializedName("avg_speed")
    val avgSpeed: Float = 0f,
    
    @SerializedName("point_count")
    val pointCount: Int = 0,
    
    @SerializedName("min_latitude")
    val minLatitude: Double = 0.0,
    
    @SerializedName("max_latitude")
    val maxLatitude: Double = 0.0,
    
    @SerializedName("min_longitude")
    val minLongitude: Double = 0.0,
    
    @SerializedName("max_longitude")
    val maxLongitude: Double = 0.0,
    
    @SerializedName("created_at")
    val createdAt: Long = System.currentTimeMillis(),
    
    @SerializedName("updated_at")
    val updatedAt: Long = System.currentTimeMillis(),
    
    @SerializedName("synced")
    val synced: Boolean = false,
    
    @SerializedName("user_id")
    val userId: String = "",
    
    @SerializedName("tags")
    val tags: String = "", // JSON字符串存储标签
    
    @SerializedName("color")
    val color: String = "#FF0000" // 轨迹显示颜色
) : Parcelable {

    companion object {
        /**
         * 创建新的轨迹记录
         */
        fun createNew(
            name: String = "",
            description: String = "",
            userId: String = ""
        ): Trajectory {
            val now = System.currentTimeMillis()
            return Trajectory(
                name = name.ifEmpty { "轨迹 ${getFormattedDate(now)}" },
                description = description,
                startTime = now,
                userId = userId,
                createdAt = now,
                updatedAt = now
            )
        }

        /**
         * 获取格式化日期
         */
        private fun getFormattedDate(timestamp: Long): String {
            val instant = Instant.fromEpochMilliseconds(timestamp)
            val localDateTime = instant.toLocalDateTime(TimeZone.currentSystemDefault())
            return "${localDateTime.monthNumber}月${localDateTime.dayOfMonth}日"
        }
    }

    /**
     * 获取轨迹持续时间（毫秒）
     */
    fun getDuration(): Long {
        return if (endTime != null) {
            endTime - startTime
        } else if (isRecording) {
            System.currentTimeMillis() - startTime
        } else {
            totalDuration
        }
    }

    /**
     * 获取格式化的持续时间
     */
    fun getFormattedDuration(): String {
        val duration = getDuration()
        val hours = duration / (1000 * 60 * 60)
        val minutes = (duration % (1000 * 60 * 60)) / (1000 * 60)
        val seconds = (duration % (1000 * 60)) / 1000

        return when {
            hours > 0 -> String.format("%d小时%d分钟", hours, minutes)
            minutes > 0 -> String.format("%d分钟%d秒", minutes, seconds)
            else -> String.format("%d秒", seconds)
        }
    }

    /**
     * 获取格式化的距离
     */
    fun getFormattedDistance(): String {
        return when {
            totalDistance >= 1000 -> String.format("%.2f km", totalDistance / 1000)
            totalDistance >= 1 -> String.format("%.0f m", totalDistance)
            else -> "0 m"
        }
    }

    /**
     * 获取格式化的平均速度
     */
    fun getFormattedAvgSpeed(): String {
        return if (avgSpeed > 0) {
            String.format("%.1f m/s", avgSpeed)
        } else {
            "0 m/s"
        }
    }

    /**
     * 获取格式化的最大速度
     */
    fun getFormattedMaxSpeed(): String {
        return if (maxSpeed > 0) {
            String.format("%.1f m/s", maxSpeed)
        } else {
            "0 m/s"
        }
    }

    /**
     * 获取开始时间的格式化字符串
     */
    fun getFormattedStartTime(): String {
        val instant = Instant.fromEpochMilliseconds(startTime)
        val localDateTime = instant.toLocalDateTime(TimeZone.currentSystemDefault())
        return "${localDateTime.date} ${localDateTime.hour}:${localDateTime.minute.toString().padStart(2, '0')}"
    }

    /**
     * 获取结束时间的格式化字符串
     */
    fun getFormattedEndTime(): String {
        return if (endTime != null) {
            val instant = Instant.fromEpochMilliseconds(endTime)
            val localDateTime = instant.toLocalDateTime(TimeZone.currentSystemDefault())
            "${localDateTime.date} ${localDateTime.hour}:${localDateTime.minute.toString().padStart(2, '0')}"
        } else {
            "进行中"
        }
    }

    /**
     * 获取轨迹日期
     */
    fun getDate(): LocalDate {
        val instant = Instant.fromEpochMilliseconds(startTime)
        return instant.toLocalDateTime(TimeZone.currentSystemDefault()).date
    }

    /**
     * 获取轨迹状态
     */
    fun getStatus(): TrajectoryStatus {
        return when {
            isRecording -> TrajectoryStatus.RECORDING
            endTime != null -> TrajectoryStatus.COMPLETED
            else -> TrajectoryStatus.PAUSED
        }
    }

    /**
     * 获取轨迹质量评级
     */
    fun getQualityRating(): TrajectoryQuality {
        return when {
            pointCount < 10 -> TrajectoryQuality.POOR
            totalDistance < 100 -> TrajectoryQuality.FAIR
            avgSpeed > 0 && pointCount > 50 -> TrajectoryQuality.EXCELLENT
            pointCount > 20 -> TrajectoryQuality.GOOD
            else -> TrajectoryQuality.FAIR
        }
    }

    /**
     * 检查轨迹是否有效
     */
    fun isValid(): Boolean {
        return pointCount > 0 && 
               startTime > 0 && 
               minLatitude != 0.0 && 
               maxLatitude != 0.0 &&
               minLongitude != 0.0 && 
               maxLongitude != 0.0
    }

    /**
     * 获取边界框中心点
     */
    fun getCenterPoint(): Pair<Double, Double> {
        val centerLat = (minLatitude + maxLatitude) / 2
        val centerLng = (minLongitude + maxLongitude) / 2
        return Pair(centerLat, centerLng)
    }

    /**
     * 获取边界框范围
     */
    fun getBoundingBox(): BoundingBox {
        return BoundingBox(
            minLatitude = minLatitude,
            maxLatitude = maxLatitude,
            minLongitude = minLongitude,
            maxLongitude = maxLongitude
        )
    }

    /**
     * 更新轨迹统计信息
     */
    fun updateStats(
        points: List<LocationPoint>
    ): Trajectory {
        if (points.isEmpty()) return this

        var distance = 0f
        var maxSpd = 0f
        var minLat = Double.MAX_VALUE
        var maxLat = Double.MIN_VALUE
        var minLng = Double.MAX_VALUE
        var maxLng = Double.MIN_VALUE

        // 计算统计信息
        for (i in points.indices) {
            val point = points[i]
            
            // 更新边界
            minLat = min(minLat, point.latitude)
            maxLat = max(maxLat, point.latitude)
            minLng = min(minLng, point.longitude)
            maxLng = max(maxLng, point.longitude)
            
            // 更新最大速度
            maxSpd = max(maxSpd, point.speed)
            
            // 计算距离
            if (i > 0) {
                distance += points[i - 1].distanceTo(point)
            }
        }

        val duration = getDuration()
        val avgSpd = if (duration > 0) distance / (duration / 1000f) else 0f

        return copy(
            totalDistance = distance,
            maxSpeed = maxSpd,
            avgSpeed = avgSpd,
            pointCount = points.size,
            minLatitude = minLat,
            maxLatitude = maxLat,
            minLongitude = minLng,
            maxLongitude = maxLng,
            updatedAt = System.currentTimeMillis()
        )
    }

    /**
     * 完成轨迹记录
     */
    fun finish(): Trajectory {
        return copy(
            isRecording = false,
            endTime = System.currentTimeMillis(),
            updatedAt = System.currentTimeMillis()
        )
    }

    /**
     * 暂停轨迹记录
     */
    fun pause(): Trajectory {
        return copy(
            isRecording = false,
            updatedAt = System.currentTimeMillis()
        )
    }

    /**
     * 恢复轨迹记录
     */
    fun resume(): Trajectory {
        return copy(
            isRecording = true,
            updatedAt = System.currentTimeMillis()
        )
    }
}

/**
 * 轨迹状态
 */
enum class TrajectoryStatus {
    RECORDING,   // 记录中
    PAUSED,      // 已暂停
    COMPLETED    // 已完成
}

/**
 * 轨迹质量等级
 */
enum class TrajectoryQuality {
    EXCELLENT,   // 优秀
    GOOD,        // 良好
    FAIR,        // 一般
    POOR         // 较差
}

/**
 * 边界框数据类
 */
@Parcelize
data class BoundingBox(
    val minLatitude: Double,
    val maxLatitude: Double,
    val minLongitude: Double,
    val maxLongitude: Double
) : Parcelable {
    
    /**
     * 获取中心点
     */
    fun getCenter(): Pair<Double, Double> {
        return Pair(
            (minLatitude + maxLatitude) / 2,
            (minLongitude + maxLongitude) / 2
        )
    }
    
    /**
     * 获取跨度
     */
    fun getSpan(): Pair<Double, Double> {
        return Pair(
            maxLatitude - minLatitude,
            maxLongitude - minLongitude
        )
    }
}