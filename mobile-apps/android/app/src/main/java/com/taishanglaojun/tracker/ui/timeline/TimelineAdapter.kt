package com.taishanglaojun.tracker.ui.timeline

import android.graphics.Color
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.taishanglaojun.tracker.R
import com.taishanglaojun.tracker.data.model.LocationPoint
import java.text.SimpleDateFormat
import java.util.*
import kotlin.math.*

class TimelineAdapter : RecyclerView.Adapter<TimelineAdapter.TimelineViewHolder>() {

    private var locationPoints: List<LocationPoint> = emptyList()
    private val timeFormat = SimpleDateFormat("HH:mm:ss", Locale.getDefault())
    private val dateFormat = SimpleDateFormat("MM-dd HH:mm:ss", Locale.getDefault())

    fun setLocationPoints(points: List<LocationPoint>) {
        locationPoints = points.sortedBy { it.timestamp }
        notifyDataSetChanged()
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): TimelineViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_timeline, parent, false)
        return TimelineViewHolder(view)
    }

    override fun onBindViewHolder(holder: TimelineViewHolder, position: Int) {
        val point = locationPoints[position]
        val previousPoint = if (position > 0) locationPoints[position - 1] else null
        
        holder.bind(point, previousPoint, position == 0, position == locationPoints.size - 1)
    }

    override fun getItemCount(): Int = locationPoints.size

    class TimelineViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val textTime: TextView = itemView.findViewById(R.id.text_time)
        private val textStatus: TextView = itemView.findViewById(R.id.text_status)
        private val textCoordinates: TextView = itemView.findViewById(R.id.text_coordinates)
        private val textSpeed: TextView = itemView.findViewById(R.id.text_speed)
        private val textAccuracy: TextView = itemView.findViewById(R.id.text_accuracy)
        private val textDistanceFromPrevious: TextView = itemView.findViewById(R.id.text_distance_from_previous)
        private val textTimeFromPrevious: TextView = itemView.findViewById(R.id.text_time_from_previous)
        private val layoutDistanceTime: View = itemView.findViewById(R.id.layout_distance_time)
        private val lineTop: View = itemView.findViewById(R.id.line_top)
        private val lineBottom: View = itemView.findViewById(R.id.line_bottom)
        private val timelineDot: View = itemView.findViewById(R.id.timeline_dot)

        private val timeFormat = SimpleDateFormat("HH:mm:ss", Locale.getDefault())

        fun bind(point: LocationPoint, previousPoint: LocationPoint?, isFirst: Boolean, isLast: Boolean) {
            // 设置时间
            textTime.text = timeFormat.format(Date(point.timestamp))

            // 设置状态
            val status = getPointStatus(point)
            textStatus.text = status.first
            textStatus.setBackgroundColor(status.second)

            // 设置坐标
            textCoordinates.text = "纬度: ${String.format("%.6f", point.latitude)}, 经度: ${String.format("%.6f", point.longitude)}"

            // 设置速度
            val speedKmh = (point.speed ?: 0f) * 3.6f // m/s to km/h
            textSpeed.text = "速度: ${String.format("%.1f", speedKmh)} km/h"

            // 设置精度
            textAccuracy.text = "精度: ±${point.accuracy?.toInt() ?: 0}m"

            // 设置距离和时间差信息
            if (previousPoint != null) {
                val distance = calculateDistance(
                    previousPoint.latitude, previousPoint.longitude,
                    point.latitude, point.longitude
                )
                val timeDiff = (point.timestamp - previousPoint.timestamp) / 1000 // seconds

                textDistanceFromPrevious.text = "距离: +${String.format("%.0f", distance)}m"
                textTimeFromPrevious.text = "间隔: ${formatTimeDifference(timeDiff)}"
                layoutDistanceTime.visibility = View.VISIBLE
            } else {
                layoutDistanceTime.visibility = View.GONE
            }

            // 设置时间轴线条
            lineTop.visibility = if (isFirst) View.INVISIBLE else View.VISIBLE
            lineBottom.visibility = if (isLast) View.INVISIBLE else View.VISIBLE

            // 设置时间轴点颜色
            timelineDot.isSelected = isFirst || isLast
        }

        private fun getPointStatus(point: LocationPoint): Pair<String, Int> {
            val speed = point.speed ?: 0f
            return when {
                speed < 0.5f -> "静止" to Color.parseColor("#FF9800")
                speed < 2.0f -> "步行" to Color.parseColor("#4CAF50")
                speed < 10.0f -> "慢跑" to Color.parseColor("#2196F3")
                speed < 25.0f -> "骑行" to Color.parseColor("#9C27B0")
                else -> "驾车" to Color.parseColor("#F44336")
            }
        }

        private fun calculateDistance(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Double {
            val earthRadius = 6371000.0 // meters
            val dLat = Math.toRadians(lat2 - lat1)
            val dLon = Math.toRadians(lon2 - lon1)
            val a = sin(dLat / 2) * sin(dLat / 2) +
                    cos(Math.toRadians(lat1)) * cos(Math.toRadians(lat2)) *
                    sin(dLon / 2) * sin(dLon / 2)
            val c = 2 * atan2(sqrt(a), sqrt(1 - a))
            return earthRadius * c
        }

        private fun formatTimeDifference(seconds: Long): String {
            return when {
                seconds < 60 -> "${seconds}秒"
                seconds < 3600 -> "${seconds / 60}分${seconds % 60}秒"
                else -> "${seconds / 3600}时${(seconds % 3600) / 60}分"
            }
        }
    }
}