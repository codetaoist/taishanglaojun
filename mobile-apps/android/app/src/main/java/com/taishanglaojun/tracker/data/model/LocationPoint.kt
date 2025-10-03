package com.taishanglaojun.tracker.data.model

import android.location.Location
import android.os.Parcelable
import androidx.room.Entity
import androidx.room.PrimaryKey
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize
import kotlinx.datetime.Instant
import kotlinx.datetime.LocalDateTime
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime
import java.util.UUID

/**
 * 位置点数据模型
 * 用于存储GPS位置信息
 */
@Entity(
    tableName = "location_points",
    foreignKeys = [
        androidx.room.ForeignKey(
            entity = Trajectory::class,
            parentColumns = ["id"],
            childColumns = ["trajectoryId"],
            onDelete = androidx.room.ForeignKey.CASCADE
        )
    ],
    indices = [
        androidx.room.Index(value = ["trajectoryId"]),
        androidx.room.Index(value = ["timestamp"])
    ]
)
@Parcelize
data class LocationPoint(
    @PrimaryKey
    val id: String = UUID.randomUUID().toString(),
    
    @SerializedName("latitude")
    val latitude: Double,
    
    @SerializedName("longitude")
    val longitude: Double,
    
    @SerializedName("timestamp")
    val timestamp: Long = System.currentTimeMillis(),
    
    @SerializedName("accuracy")
    val accuracy: Float = 0f,
    
    @SerializedName("altitude")
    val altitude: Double = 0.0,
    
    @SerializedName("speed")
    val speed: Float = 0f,
    
    @SerializedName("bearing")
    val bearing: Float = 0f,
    
    @SerializedName("trajectory_id")
    val trajectoryId: String = "",
    
    @SerializedName("provider")
    val provider: String = "gps",
    
    @SerializedName("is_mock")
    val isMock: Boolean = false,
    
    @SerializedName("battery_level")
    val batteryLevel: Int = -1,
    
    @SerializedName("network_type")
    val networkType: String = "unknown"
) : Parcelable {

    companion object {
        /**
         * 从Android Location对象创建LocationPoint
         */
        fun fromLocation(
            location: Location,
            trajectoryId: String = "",
            batteryLevel: Int = -1,
            networkType: String = "unknown"
        ): LocationPoint {
            return LocationPoint(
                latitude = location.latitude,
                longitude = location.longitude,
                timestamp = location.time,
                accuracy = location.accuracy,
                altitude = location.altitude,
                speed = if (location.hasSpeed()) location.speed else 0f,
                bearing = if (location.hasBearing()) location.bearing else 0f,
                trajectoryId = trajectoryId,
                provider = location.provider ?: "unknown",
                isMock = location.isFromMockProvider,
                batteryLevel = batteryLevel,
                networkType = networkType
            )
        }

        /**
         * 创建测试用的LocationPoint
         */
        fun createTest(
            latitude: Double = 39.9042,
            longitude: Double = 116.4074,
            trajectoryId: String = ""
        ): LocationPoint {
            return LocationPoint(
                latitude = latitude,
                longitude = longitude,
                trajectoryId = trajectoryId,
                accuracy = 5.0f,
                provider = "test"
            )
        }
    }

    /**
     * 转换为Android Location对象
     */
    fun toLocation(): Location {
        return Location(provider).apply {
            latitude = this@LocationPoint.latitude
            longitude = this@LocationPoint.longitude
            time = this@LocationPoint.timestamp
            accuracy = this@LocationPoint.accuracy
            altitude = this@LocationPoint.altitude
            if (this@LocationPoint.speed > 0) {
                speed = this@LocationPoint.speed
            }
            if (this@LocationPoint.bearing > 0) {
                bearing = this@LocationPoint.bearing
            }
        }
    }

    /**
     * 获取格式化的时间戳
     */
    fun getFormattedTimestamp(): String {
        val instant = Instant.fromEpochMilliseconds(timestamp)
        val localDateTime = instant.toLocalDateTime(TimeZone.currentSystemDefault())
        return "${localDateTime.date} ${localDateTime.time}"
    }

    /**
     * 获取本地日期时间
     */
    fun getLocalDateTime(): LocalDateTime {
        val instant = Instant.fromEpochMilliseconds(timestamp)
        return instant.toLocalDateTime(TimeZone.currentSystemDefault())
    }

    /**
     * 计算与另一个位置点的距离（米）
     */
    fun distanceTo(other: LocationPoint): Float {
        val results = FloatArray(1)
        Location.distanceBetween(
            latitude, longitude,
            other.latitude, other.longitude,
            results
        )
        return results[0]
    }

    /**
     * 计算与另一个位置点的方位角（度）
     */
    fun bearingTo(other: LocationPoint): Float {
        val results = FloatArray(2)
        Location.distanceBetween(
            latitude, longitude,
            other.latitude, other.longitude,
            results
        )
        return results[1]
    }

    /**
     * 验证位置点是否有效
     */
    fun isValid(): Boolean {
        return latitude in -90.0..90.0 &&
                longitude in -180.0..180.0 &&
                accuracy >= 0 &&
                timestamp > 0
    }

    /**
     * 检查位置精度是否足够好
     */
    fun hasGoodAccuracy(threshold: Float = 50f): Boolean {
        return accuracy > 0 && accuracy <= threshold
    }

    /**
     * 获取格式化的坐标字符串
     */
    fun getFormattedCoordinates(): String {
        return String.format("%.6f, %.6f", latitude, longitude)
    }

    /**
     * 获取格式化的速度字符串
     */
    fun getFormattedSpeed(): String {
        return if (speed > 0) {
            String.format("%.1f m/s", speed)
        } else {
            "静止"
        }
    }

    /**
     * 获取格式化的精度字符串
     */
    fun getFormattedAccuracy(): String {
        return String.format("±%.1f m", accuracy)
    }

    /**
     * 检查是否为室内位置（基于精度判断）
     */
    fun isIndoorLocation(): Boolean {
        return accuracy > 100f || provider == "network"
    }

    /**
     * 获取位置质量等级
     */
    fun getQualityLevel(): LocationQuality {
        return when {
            accuracy <= 5f -> LocationQuality.EXCELLENT
            accuracy <= 10f -> LocationQuality.GOOD
            accuracy <= 20f -> LocationQuality.FAIR
            accuracy <= 50f -> LocationQuality.POOR
            else -> LocationQuality.BAD
        }
    }
}

/**
 * 位置质量等级
 */
enum class LocationQuality {
    EXCELLENT,  // 优秀 (≤5m)
    GOOD,       // 良好 (≤10m)
    FAIR,       // 一般 (≤20m)
    POOR,       // 较差 (≤50m)
    BAD         // 很差 (>50m)
}