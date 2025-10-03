package com.taishanglaojun.tracker.ui

import android.os.Bundle
import android.view.MenuItem
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.google.android.material.bottomnavigation.BottomNavigationView
import com.taishanglaojun.tracker.R
import com.taishanglaojun.tracker.data.model.Trajectory
import com.taishanglaojun.tracker.databinding.ActivityTrajectoryVisualizationBinding
import com.taishanglaojun.tracker.ui.fragment.MapFragment
import com.taishanglaojun.tracker.ui.fragment.TimelineFragment
import com.taishanglaojun.tracker.viewmodel.LocationViewModel
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.launch

/**
 * 轨迹可视化活动
 * 提供地图视图和时间轴视图的切换
 */
@AndroidEntryPoint
class TrajectoryVisualizationActivity : AppCompatActivity() {
    
    private lateinit var binding: ActivityTrajectoryVisualizationBinding
    private val locationViewModel: LocationViewModel by viewModels()
    
    private var mapFragment: MapFragment? = null
    private var timelineFragment: TimelineFragment? = null
    private var selectedTrajectory: Trajectory? = null
    
    companion object {
        const val EXTRA_TRAJECTORY_ID = "trajectory_id"
    }
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityTrajectoryVisualizationBinding.inflate(layoutInflater)
        setContentView(binding.root)
        
        setupToolbar()
        setupBottomNavigation()
        loadTrajectory()
        
        // 默认显示地图视图
        if (savedInstanceState == null) {
            showMapFragment()
        }
    }
    
    /**
     * 设置工具栏
     */
    private fun setupToolbar() {
        setSupportActionBar(binding.toolbar)
        supportActionBar?.apply {
            setDisplayHomeAsUpEnabled(true)
            setDisplayShowHomeEnabled(true)
            title = "轨迹可视化"
        }
    }
    
    /**
     * 设置底部导航
     */
    private fun setupBottomNavigation() {
        binding.bottomNavigation.setOnItemSelectedListener { item ->
            when (item.itemId) {
                R.id.nav_map -> {
                    showMapFragment()
                    true
                }
                R.id.nav_timeline -> {
                    showTimelineFragment()
                    true
                }
                else -> false
            }
        }
    }
    
    /**
     * 加载轨迹数据
     */
    private fun loadTrajectory() {
        val trajectoryId = intent.getStringExtra(EXTRA_TRAJECTORY_ID)
        if (trajectoryId != null) {
            lifecycleScope.launch {
                try {
                    val trajectories = locationViewModel.getTrajectories()
                    selectedTrajectory = trajectories.find { it.id == trajectoryId }
                    selectedTrajectory?.let { trajectory ->
                        supportActionBar?.title = trajectory.name ?: "轨迹可视化"
                        updateFragments(trajectory)
                    }
                } catch (e: Exception) {
                    // 处理错误
                    finish()
                }
            }
        }
    }
    
    /**
     * 显示地图片段
     */
    private fun showMapFragment() {
        if (mapFragment == null) {
            mapFragment = MapFragment.newInstance()
        }
        
        mapFragment?.let { fragment ->
            selectedTrajectory?.let { trajectory ->
                fragment.setTrajectory(trajectory)
            }
            replaceFragment(fragment)
        }
    }
    
    /**
     * 显示时间轴片段
     */
    private fun showTimelineFragment() {
        if (timelineFragment == null) {
            timelineFragment = TimelineFragment.newInstance()
        }
        
        timelineFragment?.let { fragment ->
            selectedTrajectory?.let { trajectory ->
                fragment.setTrajectory(trajectory)
            }
            replaceFragment(fragment)
        }
    }
    
    /**
     * 替换片段
     */
    private fun replaceFragment(fragment: Fragment) {
        supportFragmentManager.beginTransaction()
            .replace(R.id.fragment_container, fragment)
            .commit()
    }
    
    /**
     * 更新片段数据
     */
    private fun updateFragments(trajectory: Trajectory) {
        mapFragment?.setTrajectory(trajectory)
        timelineFragment?.setTrajectory(trajectory)
    }
    
    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
            android.R.id.home -> {
                onBackPressed()
                true
            }
            else -> super.onOptionsItemSelected(item)
        }
    }
}