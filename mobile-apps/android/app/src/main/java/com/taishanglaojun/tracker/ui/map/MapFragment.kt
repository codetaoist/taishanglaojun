package com.taishanglaojun.tracker.ui.map

import android.Manifest
import android.content.pm.PackageManager
import android.graphics.Color
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.core.app.ActivityCompat
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import com.google.android.gms.maps.CameraUpdateFactory
import com.google.android.gms.maps.GoogleMap
import com.google.android.gms.maps.OnMapReadyCallback
import com.google.android.gms.maps.SupportMapFragment
import com.google.android.gms.maps.model.*
import com.taishanglaojun.tracker.R
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.data.model.Trajectory
import com.taishanglaojun.tracker.databinding.FragmentMapBinding
import com.taishanglaojun.tracker.viewmodel.LocationViewModel
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.launch
import kotlin.math.*

/**
 * 地图界面Fragment
 * 负责显示轨迹和位置点的可视化
 */
@AndroidEntryPoint
class MapFragment : Fragment(), OnMapReadyCallback {
    
    private var _binding: FragmentMapBinding? = null
    private val binding get() = _binding!!
    
    private val locationViewModel: LocationViewModel by viewModels()
    private var googleMap: GoogleMap? = null
    
    // 地图相关
    private val trajectoryPolylines = mutableMapOf<String, Polyline>()
    private val locationMarkers = mutableMapOf<String, Marker>()
    private var currentLocationMarker: Marker? = null
    
    // 可视化配置
    private val trajectoryColors = listOf(
        Color.BLUE, Color.RED, Color.GREEN, Color.MAGENTA,
        Color.CYAN, Color.YELLOW, Color.parseColor("#FF9800"),
        Color.parseColor("#9C27B0"), Color.parseColor("#4CAF50")
    )
    
    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentMapBinding.inflate(inflater, container, false)
        return binding.root
    }
    
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        // 初始化地图
        val mapFragment = childFragmentManager.findFragmentById(R.id.map) as SupportMapFragment?
        mapFragment?.getMapAsync(this)
        
        // 设置UI事件监听
        setupUIListeners()
        
        // 观察数据变化
        observeViewModel()
    }
    
    override fun onMapReady(map: GoogleMap) {
        googleMap = map
        
        // 配置地图
        configureMap()
        
        // 加载轨迹数据
        loadTrajectories()
    }
    
    /**
     * 配置地图设置
     */
    private fun configureMap() {
        googleMap?.let { map ->
            // 启用我的位置
            if (ActivityCompat.checkSelfPermission(
                    requireContext(),
                    Manifest.permission.ACCESS_FINE_LOCATION
                ) == PackageManager.PERMISSION_GRANTED
            ) {
                map.isMyLocationEnabled = true
                map.uiSettings.isMyLocationButtonEnabled = true
            }
            
            // 配置地图UI
            map.uiSettings.apply {
                isZoomControlsEnabled = true
                isCompassEnabled = true
                isMapToolbarEnabled = true
                isRotateGesturesEnabled = true
                isScrollGesturesEnabled = true
                isTiltGesturesEnabled = true
                isZoomGesturesEnabled = true
            }
            
            // 设置地图类型
            map.mapType = GoogleMap.MAP_TYPE_NORMAL
            
            // 设置地图点击监听
            map.setOnMapClickListener { latLng ->
                onMapClick(latLng)
            }
            
            map.setOnMarkerClickListener { marker ->
                onMarkerClick(marker)
            }
            
            map.setOnPolylineClickListener { polyline ->
                onPolylineClick(polyline)
            }
        }
    }
    
    /**
     * 设置UI监听器
     */
    private fun setupUIListeners() {
        binding.apply {
            // 地图类型切换
            btnMapType.setOnClickListener {
                toggleMapType()
            }
            
            // 显示/隐藏轨迹
            btnToggleTrajectories.setOnClickListener {
                toggleTrajectoriesVisibility()
            }
            
            // 居中到当前位置
            btnCenterLocation.setOnClickListener {
                centerToCurrentLocation()
            }
            
            // 适应所有轨迹
            btnFitAllTrajectories.setOnClickListener {
                fitAllTrajectories()
            }
            
            // 清除地图
            btnClearMap.setOnClickListener {
                clearMap()
            }
        }
    }
    
    /**
     * 观察ViewModel数据变化
     */
    private fun observeViewModel() {
        // 观察当前位置
        locationViewModel.currentLocation.observe(viewLifecycleOwner) { location ->
            location?.let {
                updateCurrentLocationMarker(it)
            }
        }
        
        // 观察当前轨迹
        locationViewModel.currentTrajectory.observe(viewLifecycleOwner) { trajectory ->
            trajectory?.let {
                updateCurrentTrajectory(it)
            }
        }
        
        // 观察轨迹列表
        locationViewModel.trajectories.observe(viewLifecycleOwner) { trajectories ->
            updateTrajectories(trajectories)
        }
        
        // 观察跟踪状态
        locationViewModel.isTracking.observe(viewLifecycleOwner) { isTracking ->
            updateTrackingUI(isTracking)
        }
    }
    
    /**
     * 加载轨迹数据
     */
    private fun loadTrajectories() {
        lifecycleScope.launch {
            locationViewModel.loadTrajectories()
        }
    }
    
    /**
     * 更新当前位置标记
     */
    private fun updateCurrentLocationMarker(location: LocationPoint) {
        googleMap?.let { map ->
            val latLng = LatLng(location.latitude, location.longitude)
            
            if (currentLocationMarker == null) {
                val markerOptions = MarkerOptions()
                    .position(latLng)
                    .title("当前位置")
                    .snippet("精度: ${location.accuracy}m")
                    .icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_BLUE))
                
                currentLocationMarker = map.addMarker(markerOptions)
            } else {
                currentLocationMarker?.position = latLng
                currentLocationMarker?.snippet = "精度: ${location.accuracy}m"
            }
            
            // 如果正在跟踪，移动相机到当前位置
            if (locationViewModel.isTracking.value == true) {
                map.animateCamera(CameraUpdateFactory.newLatLngZoom(latLng, 16f))
            }
        }
    }
    
    /**
     * 更新当前轨迹
     */
    private fun updateCurrentTrajectory(trajectory: Trajectory) {
        lifecycleScope.launch {
            val points = locationViewModel.getTrajectoryPoints(trajectory.id)
            if (points.isNotEmpty()) {
                drawTrajectory(trajectory, points, true)
            }
        }
    }
    
    /**
     * 更新轨迹列表
     */
    private fun updateTrajectories(trajectories: List<Trajectory>) {
        lifecycleScope.launch {
            // 清除现有轨迹
            clearTrajectories()
            
            // 绘制所有轨迹
            trajectories.forEachIndexed { index, trajectory ->
                val points = locationViewModel.getTrajectoryPoints(trajectory.id)
                if (points.isNotEmpty()) {
                    val color = trajectoryColors[index % trajectoryColors.size]
                    drawTrajectory(trajectory, points, false, color)
                }
            }
        }
    }
    
    /**
     * 绘制轨迹
     */
    private fun drawTrajectory(
        trajectory: Trajectory,
        points: List<LocationPoint>,
        isCurrentTrajectory: Boolean,
        color: Int = Color.BLUE
    ) {
        googleMap?.let { map ->
            // 移除现有的轨迹线
            trajectoryPolylines[trajectory.id]?.remove()
            
            // 创建轨迹点列表
            val latLngPoints = points.map { LatLng(it.latitude, it.longitude) }
            
            // 创建轨迹线
            val polylineOptions = PolylineOptions()
                .addAll(latLngPoints)
                .color(color)
                .width(if (isCurrentTrajectory) 8f else 5f)
                .pattern(if (isCurrentTrajectory) null else listOf(Dash(10f), Gap(5f)))
                .clickable(true)
            
            val polyline = map.addPolyline(polylineOptions)
            polyline.tag = trajectory
            trajectoryPolylines[trajectory.id] = polyline
            
            // 添加起点和终点标记
            if (latLngPoints.isNotEmpty()) {
                // 起点标记
                val startMarker = MarkerOptions()
                    .position(latLngPoints.first())
                    .title("${trajectory.name} - 起点")
                    .snippet(formatTime(trajectory.startTime))
                    .icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_GREEN))
                
                val startMarkerInstance = map.addMarker(startMarker)
                startMarkerInstance?.tag = "start_${trajectory.id}"
                locationMarkers["start_${trajectory.id}"] = startMarkerInstance!!
                
                // 终点标记（如果轨迹已结束）
                if (trajectory.endTime != null && latLngPoints.size > 1) {
                    val endMarker = MarkerOptions()
                        .position(latLngPoints.last())
                        .title("${trajectory.name} - 终点")
                        .snippet(formatTime(trajectory.endTime))
                        .icon(BitmapDescriptorFactory.defaultMarker(BitmapDescriptorFactory.HUE_RED))
                    
                    val endMarkerInstance = map.addMarker(endMarker)
                    endMarkerInstance?.tag = "end_${trajectory.id}"
                    locationMarkers["end_${trajectory.id}"] = endMarkerInstance!!
                }
            }
        }
    }
    
    /**
     * 切换地图类型
     */
    private fun toggleMapType() {
        googleMap?.let { map ->
            map.mapType = when (map.mapType) {
                GoogleMap.MAP_TYPE_NORMAL -> GoogleMap.MAP_TYPE_SATELLITE
                GoogleMap.MAP_TYPE_SATELLITE -> GoogleMap.MAP_TYPE_HYBRID
                GoogleMap.MAP_TYPE_HYBRID -> GoogleMap.MAP_TYPE_TERRAIN
                else -> GoogleMap.MAP_TYPE_NORMAL
            }
        }
    }
    
    /**
     * 切换轨迹可见性
     */
    private fun toggleTrajectoriesVisibility() {
        val isVisible = trajectoryPolylines.values.firstOrNull()?.isVisible ?: true
        
        trajectoryPolylines.values.forEach { polyline ->
            polyline.isVisible = !isVisible
        }
        
        locationMarkers.values.forEach { marker ->
            marker.isVisible = !isVisible
        }
        
        binding.btnToggleTrajectories.text = if (isVisible) "显示轨迹" else "隐藏轨迹"
    }
    
    /**
     * 居中到当前位置
     */
    private fun centerToCurrentLocation() {
        locationViewModel.currentLocation.value?.let { location ->
            val latLng = LatLng(location.latitude, location.longitude)
            googleMap?.animateCamera(CameraUpdateFactory.newLatLngZoom(latLng, 16f))
        }
    }
    
    /**
     * 适应所有轨迹
     */
    private fun fitAllTrajectories() {
        val allPoints = mutableListOf<LatLng>()
        
        trajectoryPolylines.values.forEach { polyline ->
            allPoints.addAll(polyline.points)
        }
        
        if (allPoints.isNotEmpty()) {
            val boundsBuilder = LatLngBounds.Builder()
            allPoints.forEach { boundsBuilder.include(it) }
            
            val bounds = boundsBuilder.build()
            val padding = 100 // 边距像素
            
            googleMap?.animateCamera(CameraUpdateFactory.newLatLngBounds(bounds, padding))
        }
    }
    
    /**
     * 清除地图
     */
    private fun clearMap() {
        clearTrajectories()
        currentLocationMarker?.remove()
        currentLocationMarker = null
    }
    
    /**
     * 清除轨迹
     */
    private fun clearTrajectories() {
        trajectoryPolylines.values.forEach { it.remove() }
        trajectoryPolylines.clear()
        
        locationMarkers.values.forEach { it.remove() }
        locationMarkers.clear()
    }
    
    /**
     * 更新跟踪状态UI
     */
    private fun updateTrackingUI(isTracking: Boolean) {
        binding.apply {
            if (isTracking) {
                trackingStatusIndicator.setBackgroundColor(Color.GREEN)
                trackingStatusText.text = "正在跟踪"
            } else {
                trackingStatusIndicator.setBackgroundColor(Color.GRAY)
                trackingStatusText.text = "未跟踪"
            }
        }
    }
    
    /**
     * 地图点击事件
     */
    private fun onMapClick(latLng: LatLng) {
        // 可以在这里添加地图点击的处理逻辑
        // 例如：显示坐标信息、添加标记等
    }
    
    /**
     * 标记点击事件
     */
    private fun onMarkerClick(marker: Marker): Boolean {
        // 显示标记信息
        marker.showInfoWindow()
        
        // 移动相机到标记位置
        googleMap?.animateCamera(CameraUpdateFactory.newLatLngZoom(marker.position, 16f))
        
        return true
    }
    
    /**
     * 轨迹线点击事件
     */
    private fun onPolylineClick(polyline: Polyline) {
        val trajectory = polyline.tag as? Trajectory
        trajectory?.let {
            // 显示轨迹详情
            showTrajectoryDetails(it)
        }
    }
    
    /**
     * 显示轨迹详情
     */
    private fun showTrajectoryDetails(trajectory: Trajectory) {
        // 这里可以显示轨迹详情对话框或跳转到详情页面
        // 暂时使用简单的信息显示
        val details = """
            轨迹名称: ${trajectory.name}
            开始时间: ${formatTime(trajectory.startTime)}
            结束时间: ${formatTime(trajectory.endTime)}
            总距离: ${trajectory.formatDistance()}
            持续时间: ${trajectory.formatDuration()}
            平均速度: ${trajectory.formatAverageSpeed()}
        """.trimIndent()
        
        // 可以使用AlertDialog或BottomSheet显示详情
        // 这里简化处理
    }
    
    /**
     * 格式化时间
     */
    private fun formatTime(timestamp: Long?): String {
        return if (timestamp != null) {
            java.text.SimpleDateFormat("yyyy-MM-dd HH:mm:ss", java.util.Locale.getDefault())
                .format(java.util.Date(timestamp))
        } else {
            "未知"
        }
    }
    
    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}