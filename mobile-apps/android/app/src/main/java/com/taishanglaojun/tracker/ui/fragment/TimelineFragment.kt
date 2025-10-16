package com.taishanglaojun.tracker.ui.fragment

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import com.taishanglaojun.tracker.databinding.FragmentTimelineBinding
import com.taishanglaojun.tracker.ui.adapter.TimelineAdapter
import com.taishanglaojun.tracker.viewmodel.LocationViewModel
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*
import kotlin.math.*

/**
 * 时间轴片段
 * 显示轨迹的时间序列和详细信息
 */
@AndroidEntryPoint
class TimelineFragment : Fragment() {
    
    private var _binding: FragmentTimelineBinding? = null
    private val binding get() = _binding!!
    
    private val locationViewModel: LocationViewModel by viewModels()
    private lateinit var timelineAdapter: TimelineAdapter
    
    private var trajectory: Trajectory? = null
    private var locationPoints: List<LocationPoint> = emptyList()
    
    companion object {
        fun newInstance(): TimelineFragment {
            return TimelineFragment()
        }
    }
    
    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentTimelineBinding.inflate(inflater, container, false)
        return binding.root
    }
    
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        setupRecyclerView()
        setupUI()
    }
    
    /**
     * 设置RecyclerView
     */
    private fun setupRecyclerView() {
        timelineAdapter = TimelineAdapter { point, position ->
            // 处理时间轴项目点击
            onTimelineItemClick(point, position)
        }
        
        binding.recyclerViewTimeline.apply {
            layoutManager = LinearLayoutManager(requireContext())
            adapter = timelineAdapter
        }
    }
    
    /**
     * 设置UI
     */
    private fun setupUI() {
        trajectory?.let { updateTrajectoryInfo(it) }
    }
    
    /**
     * 设置轨迹
     */
    fun setTrajectory(trajectory: Trajectory) {
        this.trajectory = trajectory
        updateTrajectoryInfo(trajectory)
        loadTrajectoryPoints(trajectory.id)
    }
    
    /**
     * 更新轨迹信息
     */
    private fun updateTrajectoryInfo(trajectory: Trajectory) {
        binding.apply {
            textTrajectoryName.text = trajectory.name ?: "未命名轨迹"
            textStartTime.text = "开始: ${formatDateTime(trajectory.startTime)}"
            
            if (trajectory.endTime != null) {
                textEndTime.text = "结束: ${formatTime(trajectory.endTime)}"
                textEndTime.visibility = View.VISIBLE
                
                val duration = (trajectory.endTime - trajectory.startTime) / 60000 // 分钟
                textDuration.text = "总时长: ${duration}分钟"
                textDuration.visibility = View.VISIBLE
                
                val distance = trajectory.totalDistance / 1000.0 // 公里
                textDistance.text = "总距离: ${"%.2f".format(distance)}km"
                textDistance.visibility = View.VISIBLE
            } else {
                textEndTime.visibility = View.GONE
                textDuration.visibility = View.GONE
                textDistance.visibility = View.GONE
            }
        }
    }
    
    /**
     * 加载轨迹点
     */
    private fun loadTrajectoryPoints(trajectoryId: String) {
        lifecycleScope.launch {
            try {
                binding.progressBar.visibility = View.VISIBLE
                binding.recyclerViewTimeline.visibility = View.GONE
                binding.textEmptyState.visibility = View.GONE
                
                locationPoints = locationViewModel.getTrajectoryPoints(trajectoryId)
                
                if (locationPoints.isEmpty()) {
                    binding.textEmptyState.visibility = View.VISIBLE
                } else {
                    // 计算额外信息
                    val enhancedPoints = enhanceLocationPoints(locationPoints)
                    timelineAdapter.submitList(enhancedPoints)
                    binding.recyclerViewTimeline.visibility = View.VISIBLE
                }
            } catch (e: Exception) {
                binding.textEmptyState.text = "加载失败: ${e.message}"
                binding.textEmptyState.visibility = View.VISIBLE
            } finally {
                binding.progressBar.visibility = View.GONE
            }
        }
    }
    
    /**
     * 增强位置点信息
     */
    private fun enhanceLocationPoints(points: List<LocationPoint>): List<TimelineItem> {
        return points.mapIndexed { index, point ->
            val prevPoint = if (index > 0) points[index - 1] else null
            val distance = prevPoint?.let { calculateDistance(it, point) } ?: 0.0
            val timeDiff = prevPoint?.let { (point.timestamp - it.timestamp) / 1000 } ?: 0L // 秒
            
            TimelineItem(
                point = point,
                index = index,
                distanceFromPrevious = distance,
                timeFromPrevious = timeDiff,
                isFirst = index == 0,
                isLast = index == points.size - 1
            )
        }
    }
    
    /**
     * 计算两点间距离（米）
     */
    private fun calculateDistance(point1: LocationPoint, point2: LocationPoint): Double {
        val R = 6371000.0 // 地球半径（米）
        val lat1Rad = Math.toRadians(point1.latitude)
        val lat2Rad = Math.toRadians(point2.latitude)
        val deltaLatRad = Math.toRadians(point2.latitude - point1.latitude)
        val deltaLngRad = Math.toRadians(point2.longitude - point1.longitude)
        
        val a = sin(deltaLatRad / 2).pow(2) +
                cos(lat1Rad) * cos(lat2Rad) *
                sin(deltaLngRad / 2).pow(2)
        val c = 2 * atan2(sqrt(a), sqrt(1 - a))
        
        return R * c
    }
    
    /**
     * 时间轴项目点击处理
     */
    private fun onTimelineItemClick(point: LocationPoint, position: Int) {
        // 可以在这里添加点击处理逻辑，比如显示详细信息对话框
    }
    
    /**
     * 格式化日期时间
     */
    private fun formatDateTime(timestamp: Long): String {
        val sdf = SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault())
        return sdf.format(Date(timestamp))
    }
    
    /**
     * 格式化时间
     */
    private fun formatTime(timestamp: Long): String {
        val sdf = SimpleDateFormat("HH:mm:ss", Locale.getDefault())
        return sdf.format(Date(timestamp))
    }
    
    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
    
    /**
     * 时间轴项目数据类
     */
    data class TimelineItem(
        val point: LocationPoint,
        val index: Int,
        val distanceFromPrevious: Double,
        val timeFromPrevious: Long,
        val isFirst: Boolean,
        val isLast: Boolean
    )
}