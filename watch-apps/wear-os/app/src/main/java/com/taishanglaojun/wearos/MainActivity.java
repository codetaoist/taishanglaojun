package com.taishanglaojun.wearos;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.TextView;
import androidx.wear.widget.BoxInsetLayout;

import com.taishanglaojun.wearos.services.LocationService;
import com.taishanglaojun.wearos.services.HealthService;
import com.taishanglaojun.wearos.services.DataSyncService;
import com.taishanglaojun.wearos.utils.PerformanceOptimizer;

/**
 * Wear OS 主活动
 * 提供用户界面和服务控制功能
 */
public class MainActivity extends Activity implements PerformanceOptimizer.PerformanceListener {
    
    private static final String TAG = "MainActivity";
    
    // UI 组件
    private TextView statusText;
    private Button locationButton;
    private Button healthButton;
    private Button syncButton;
    
    // 服务状态
    private boolean isLocationServiceRunning = false;
    private boolean isHealthServiceRunning = false;
    private boolean isSyncServiceRunning = false;
    
    // 性能优化器
    private PerformanceOptimizer performanceOptimizer;
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        
        Log.d(TAG, "MainActivity onCreate");
        
        // 初始化性能优化器
        performanceOptimizer = PerformanceOptimizer.getInstance();
        performanceOptimizer.addPerformanceListener(this);
        
        // 检查内存和电池状态
        checkSystemStatus();
        
        // 初始化UI组件
        initializeUI();
        
        // 设置按钮监听器
        setupButtonListeners();
        
        // 更新状态显示
        updateStatusDisplay();
    }
    
    /**
     * 检查系统状态
     */
    private void checkSystemStatus() {
        performanceOptimizer.executeBackgroundTask(() -> {
            // 检查内存状态
            PerformanceOptimizer.MemoryStatus memoryStatus = performanceOptimizer.checkMemoryStatus(this);
            if (memoryStatus.level == PerformanceOptimizer.MemoryLevel.LOW || 
                memoryStatus.level == PerformanceOptimizer.MemoryLevel.CRITICAL) {
                performanceOptimizer.performMemoryCleanup(this);
            }
            
            // 检查电池状态
            PerformanceOptimizer.BatteryStatus batteryStatus = performanceOptimizer.checkBatteryStatus(this);
            performanceOptimizer.applyBatteryOptimization(this, batteryStatus);
            
            // 在主线程更新UI
            performanceOptimizer.executeOnMainThread(() -> {
                updateSystemStatusDisplay(memoryStatus, batteryStatus);
            });
        });
    }
    
    /**
     * 更新系统状态显示
     */
    private void updateSystemStatusDisplay(PerformanceOptimizer.MemoryStatus memoryStatus, 
                                         PerformanceOptimizer.BatteryStatus batteryStatus) {
        String statusMessage = String.format("内存: %d%% | 电池: %d%%", 
            memoryStatus.memoryPercentage, batteryStatus.batteryLevel);
        
        if (statusText != null) {
            statusText.setText(statusMessage);
        }
        
        Log.d(TAG, "系统状态更新: " + statusMessage);
    }
    
    /**
     * 初始化界面组件
     */
    private void initializeUI() {
        statusText = findViewById(R.id.status_text);
        locationButton = findViewById(R.id.location_button);
        healthButton = findViewById(R.id.health_button);
        syncButton = findViewById(R.id.sync_button);
        
        statusText.setText("Wear OS 应用已启动");
    }
    
    /**
     * 设置按钮点击监听器
     */
    private void setupButtonListeners() {
        locationButton.setOnClickListener(v -> toggleLocationService());
        healthButton.setOnClickListener(v -> toggleHealthService());
        syncButton.setOnClickListener(v -> startSyncService());
    }
    
    /**
     * 更新状态显示
     */
    private void updateStatusDisplay() {
        String status = String.format("位置: %s | 健康: %s | 同步: %s",
            isLocationServiceRunning ? "运行中" : "已停止",
            isHealthServiceRunning ? "运行中" : "已停止", 
            isSyncServiceRunning ? "运行中" : "已停止");
        
        if (statusText != null) {
            statusText.setText(status);
        }
    }
    
    /**
     * 切换位置服务状态
     */
    private void toggleLocationService() {
        performanceOptimizer.executeBackgroundTask(() -> {
            Intent locationIntent = new Intent(this, LocationService.class);
            
            if (!isLocationServiceRunning) {
                startService(locationIntent);
                isLocationServiceRunning = true;
                Log.d(TAG, "位置服务已启动");
                
                performanceOptimizer.executeOnMainThread(() -> {
                    updateStatusDisplay();
                });
            } else {
                stopService(locationIntent);
                isLocationServiceRunning = false;
                Log.d(TAG, "位置服务已停止");
                
                performanceOptimizer.executeOnMainThread(() -> {
                    updateStatusDisplay();
                });
            }
        });
    }
    
    /**
     * 切换健康监测服务状态
     */
    private void toggleHealthService() {
        performanceOptimizer.executeBackgroundTask(() -> {
            Intent healthIntent = new Intent(this, HealthService.class);
            
            if (!isHealthServiceRunning) {
                startService(healthIntent);
                isHealthServiceRunning = true;
                Log.d(TAG, "健康监测服务已启动");
                
                performanceOptimizer.executeOnMainThread(() -> {
                    updateStatusDisplay();
                });
            } else {
                stopService(healthIntent);
                isHealthServiceRunning = false;
                Log.d(TAG, "健康监测服务已停止");
                
                performanceOptimizer.executeOnMainThread(() -> {
                    updateStatusDisplay();
                });
            }
        });
    }
    
    /**
     * 启动数据同步服务
     */
    private void startSyncService() {
        performanceOptimizer.executeNetworkTask(() -> {
            Intent syncIntent = new Intent(this, DataSyncService.class);
            startService(syncIntent);
            isSyncServiceRunning = true;
            Log.d(TAG, "数据同步服务已启动");
            
            performanceOptimizer.executeOnMainThread(() -> {
                updateStatusDisplay();
            });
        });
    }
    
    @Override
    protected void onResume() {
        super.onResume();
        Log.d(TAG, "应用恢复运行");
        
        // 重新检查系统状态
        checkSystemStatus();
        
        // 更新状态显示
        updateStatusDisplay();
    }
    
    @Override
    protected void onPause() {
        super.onPause();
        Log.d(TAG, "应用暂停");
        
        // 执行内存清理
        if (performanceOptimizer != null) {
            performanceOptimizer.performMemoryCleanup(this);
        }
    }
    
    @Override
    protected void onDestroy() {
        super.onDestroy();
        Log.d(TAG, "应用销毁");
        
        // 清理性能优化器资源
        if (performanceOptimizer != null) {
            performanceOptimizer.removePerformanceListener(this);
            performanceOptimizer.cleanup();
        }
    }
    
    // 实现PerformanceListener接口
    @Override
    public void onPerformanceEvent(String eventName, long timestamp) {
        Log.d(TAG, String.format("性能事件: %s 在 %d", eventName, timestamp));
        
        // 根据性能事件更新UI或执行相应操作
        performanceOptimizer.executeOnMainThread(() -> {
            switch (eventName) {
                case "memory_cleanup":
                    // 内存清理完成，可以更新状态
                    checkSystemStatus();
                    break;
                case "battery_optimization":
                    // 电池优化完成，可以更新状态
                    checkSystemStatus();
                    break;
                default:
                    // 其他性能事件
                    break;
            }
        });
    }
}