package com.taishanglaojun.wearos.utils;

import android.app.ActivityManager;
import android.content.Context;
import android.os.BatteryManager;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.List;
import java.util.ArrayList;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 性能优化工具类
 * 提供内存管理、电池优化、线程池管理等功能
 */
public class PerformanceOptimizer {
    
    private static final String TAG = "PerformanceOptimizer";
    
    // 单例实例
    private static volatile PerformanceOptimizer instance;
    
    // 线程池管理
    private ExecutorService backgroundExecutor;
    private ExecutorService networkExecutor;
    private Handler mainHandler;
    
    // 内存管理
    private static final long LOW_MEMORY_THRESHOLD = 50 * 1024 * 1024; // 50MB
    private static final long CRITICAL_MEMORY_THRESHOLD = 20 * 1024 * 1024; // 20MB
    
    // 电池优化
    private static final int LOW_BATTERY_THRESHOLD = 20; // 20%
    private static final int CRITICAL_BATTERY_THRESHOLD = 10; // 10%
    
    // 性能监控
    private final Map<String, Long> performanceMetrics = new ConcurrentHashMap<>();
    private final List<PerformanceListener> listeners = new ArrayList<>();
    
    /**
     * 私有构造函数
     */
    private PerformanceOptimizer() {
        initializeThreadPools();
        mainHandler = new Handler(Looper.getMainLooper());
    }
    
    /**
     * 获取单例实例
     */
    public static PerformanceOptimizer getInstance() {
        if (instance == null) {
            synchronized (PerformanceOptimizer.class) {
                if (instance == null) {
                    instance = new PerformanceOptimizer();
                }
            }
        }
        return instance;
    }
    
    /**
     * 初始化线程池
     */
    private void initializeThreadPools() {
        // 后台任务线程池 - 用于一般后台处理
        backgroundExecutor = Executors.newFixedThreadPool(2, r -> {
            Thread thread = new Thread(r, "WearOS-Background");
            thread.setPriority(Thread.NORM_PRIORITY - 1);
            return thread;
        });
        
        // 网络任务线程池 - 用于网络请求
        networkExecutor = Executors.newFixedThreadPool(1, r -> {
            Thread thread = new Thread(r, "WearOS-Network");
            thread.setPriority(Thread.NORM_PRIORITY);
            return thread;
        });
        
        Log.d(TAG, "线程池初始化完成");
    }
    
    // 内存管理相关方法
    
    /**
     * 检查内存状态
     */
    public MemoryStatus checkMemoryStatus(Context context) {
        ActivityManager activityManager = (ActivityManager) context.getSystemService(Context.ACTIVITY_SERVICE);
        ActivityManager.MemoryInfo memoryInfo = new ActivityManager.MemoryInfo();
        activityManager.getMemoryInfo(memoryInfo);
        
        long availableMemory = memoryInfo.availMem;
        long totalMemory = memoryInfo.totalMem;
        long usedMemory = totalMemory - availableMemory;
        
        MemoryStatus status = new MemoryStatus();
        status.availableMemory = availableMemory;
        status.totalMemory = totalMemory;
        status.usedMemory = usedMemory;
        status.memoryPercentage = (int) ((usedMemory * 100) / totalMemory);
        status.isLowMemory = memoryInfo.lowMemory;
        
        if (availableMemory < CRITICAL_MEMORY_THRESHOLD) {
            status.level = MemoryLevel.CRITICAL;
        } else if (availableMemory < LOW_MEMORY_THRESHOLD) {
            status.level = MemoryLevel.LOW;
        } else {
            status.level = MemoryLevel.NORMAL;
        }
        
        Log.d(TAG, String.format("内存状态: 可用=%dMB, 总计=%dMB, 使用率=%d%%, 级别=%s",
            availableMemory / (1024 * 1024), totalMemory / (1024 * 1024), 
            status.memoryPercentage, status.level));
        
        return status;
    }
    
    /**
     * 执行内存清理
     */
    public void performMemoryCleanup(Context context) {
        Log.d(TAG, "开始内存清理...");
        
        // 建议系统进行垃圾回收
        System.gc();
        
        // 清理缓存
        clearCaches(context);
        
        // 通知监听器
        notifyPerformanceEvent("memory_cleanup", System.currentTimeMillis());
        
        Log.d(TAG, "内存清理完成");
    }
    
    /**
     * 清理应用缓存
     */
    private void clearCaches(Context context) {
        try {
            // 清理内部缓存目录
            clearDirectory(context.getCacheDir());
            
            // 清理外部缓存目录
            if (context.getExternalCacheDir() != null) {
                clearDirectory(context.getExternalCacheDir());
            }
            
            Log.d(TAG, "缓存清理完成");
        } catch (Exception e) {
            Log.e(TAG, "缓存清理失败", e);
        }
    }
    
    /**
     * 清理目录
     */
    private void clearDirectory(java.io.File directory) {
        if (directory != null && directory.exists()) {
            java.io.File[] files = directory.listFiles();
            if (files != null) {
                for (java.io.File file : files) {
                    if (file.isDirectory()) {
                        clearDirectory(file);
                    }
                    file.delete();
                }
            }
        }
    }
    
    // 电池优化相关方法
    
    /**
     * 检查电池状态
     */
    public BatteryStatus checkBatteryStatus(Context context) {
        BatteryManager batteryManager = (BatteryManager) context.getSystemService(Context.BATTERY_SERVICE);
        
        BatteryStatus status = new BatteryStatus();
        
        if (batteryManager != null) {
            status.batteryLevel = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY);
            status.isCharging = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_STATUS) == BatteryManager.BATTERY_STATUS_CHARGING;
            status.temperature = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CURRENT_NOW);
            
            if (status.batteryLevel <= CRITICAL_BATTERY_THRESHOLD) {
                status.level = BatteryLevel.CRITICAL;
            } else if (status.batteryLevel <= LOW_BATTERY_THRESHOLD) {
                status.level = BatteryLevel.LOW;
            } else {
                status.level = BatteryLevel.NORMAL;
            }
        } else {
            status.batteryLevel = -1;
            status.level = BatteryLevel.UNKNOWN;
        }
        
        Log.d(TAG, String.format("电池状态: 电量=%d%%, 充电=%s, 级别=%s",
            status.batteryLevel, status.isCharging, status.level));
        
        return status;
    }
    
    /**
     * 应用电池优化策略
     */
    public void applyBatteryOptimization(Context context, BatteryStatus batteryStatus) {
        Log.d(TAG, "应用电池优化策略...");
        
        switch (batteryStatus.level) {
            case CRITICAL:
                applyCriticalBatteryMode();
                break;
            case LOW:
                applyLowBatteryMode();
                break;
            case NORMAL:
                applyNormalBatteryMode();
                break;
        }
        
        notifyPerformanceEvent("battery_optimization", System.currentTimeMillis());
    }
    
    /**
     * 应用严重低电量模式
     */
    private void applyCriticalBatteryMode() {
        Log.d(TAG, "启用严重低电量模式");
        
        // 减少后台任务频率
        reduceBackgroundTasks(0.2f); // 减少到20%
        
        // 降低传感器采样率
        reduceSensorSampling(0.1f); // 减少到10%
        
        // 暂停非关键服务
        pauseNonCriticalServices();
    }
    
    /**
     * 应用低电量模式
     */
    private void applyLowBatteryMode() {
        Log.d(TAG, "启用低电量模式");
        
        // 减少后台任务频率
        reduceBackgroundTasks(0.5f); // 减少到50%
        
        // 降低传感器采样率
        reduceSensorSampling(0.3f); // 减少到30%
        
        // 减少网络请求频率
        reduceNetworkRequests(0.5f); // 减少到50%
    }
    
    /**
     * 应用正常电量模式
     */
    private void applyNormalBatteryMode() {
        Log.d(TAG, "启用正常电量模式");
        
        // 恢复正常任务频率
        reduceBackgroundTasks(1.0f); // 100%
        
        // 恢复正常传感器采样率
        reduceSensorSampling(1.0f); // 100%
        
        // 恢复正常网络请求频率
        reduceNetworkRequests(1.0f); // 100%
    }
    
    // 任务调度优化
    
    /**
     * 执行后台任务
     */
    public void executeBackgroundTask(Runnable task) {
        if (backgroundExecutor != null && !backgroundExecutor.isShutdown()) {
            backgroundExecutor.execute(() -> {
                long startTime = System.currentTimeMillis();
                try {
                    task.run();
                } finally {
                    long duration = System.currentTimeMillis() - startTime;
                    recordPerformanceMetric("background_task_duration", duration);
                }
            });
        }
    }
    
    /**
     * 执行网络任务
     */
    public void executeNetworkTask(Runnable task) {
        if (networkExecutor != null && !networkExecutor.isShutdown()) {
            networkExecutor.execute(() -> {
                long startTime = System.currentTimeMillis();
                try {
                    task.run();
                } finally {
                    long duration = System.currentTimeMillis() - startTime;
                    recordPerformanceMetric("network_task_duration", duration);
                }
            });
        }
    }
    
    /**
     * 在主线程执行任务
     */
    public void executeOnMainThread(Runnable task) {
        if (mainHandler != null) {
            mainHandler.post(task);
        }
    }
    
    /**
     * 延迟在主线程执行任务
     */
    public void executeOnMainThreadDelayed(Runnable task, long delayMillis) {
        if (mainHandler != null) {
            mainHandler.postDelayed(task, delayMillis);
        }
    }
    
    // 性能监控
    
    /**
     * 记录性能指标
     */
    public void recordPerformanceMetric(String metricName, long value) {
        performanceMetrics.put(metricName, value);
        Log.d(TAG, String.format("性能指标: %s = %d", metricName, value));
    }
    
    /**
     * 获取性能指标
     */
    public Long getPerformanceMetric(String metricName) {
        return performanceMetrics.get(metricName);
    }
    
    /**
     * 获取所有性能指标
     */
    public Map<String, Long> getAllPerformanceMetrics() {
        return new ConcurrentHashMap<>(performanceMetrics);
    }
    
    /**
     * 清除性能指标
     */
    public void clearPerformanceMetrics() {
        performanceMetrics.clear();
        Log.d(TAG, "性能指标已清除");
    }
    
    // 私有辅助方法
    
    private void reduceBackgroundTasks(float factor) {
        // 实现后台任务频率调整逻辑
        Log.d(TAG, String.format("调整后台任务频率: %.1f%%", factor * 100));
    }
    
    private void reduceSensorSampling(float factor) {
        // 实现传感器采样率调整逻辑
        Log.d(TAG, String.format("调整传感器采样率: %.1f%%", factor * 100));
    }
    
    private void reduceNetworkRequests(float factor) {
        // 实现网络请求频率调整逻辑
        Log.d(TAG, String.format("调整网络请求频率: %.1f%%", factor * 100));
    }
    
    private void pauseNonCriticalServices() {
        // 实现暂停非关键服务逻辑
        Log.d(TAG, "暂停非关键服务");
    }
    
    private void notifyPerformanceEvent(String eventName, long timestamp) {
        for (PerformanceListener listener : listeners) {
            try {
                listener.onPerformanceEvent(eventName, timestamp);
            } catch (Exception e) {
                Log.e(TAG, "通知性能事件失败", e);
            }
        }
    }
    
    // 监听器管理
    
    /**
     * 添加性能监听器
     */
    public void addPerformanceListener(PerformanceListener listener) {
        if (listener != null && !listeners.contains(listener)) {
            listeners.add(listener);
        }
    }
    
    /**
     * 移除性能监听器
     */
    public void removePerformanceListener(PerformanceListener listener) {
        listeners.remove(listener);
    }
    
    // 清理资源
    
    /**
     * 清理资源
     */
    public void cleanup() {
        Log.d(TAG, "清理性能优化器资源...");
        
        if (backgroundExecutor != null && !backgroundExecutor.isShutdown()) {
            backgroundExecutor.shutdown();
            try {
                if (!backgroundExecutor.awaitTermination(5, TimeUnit.SECONDS)) {
                    backgroundExecutor.shutdownNow();
                }
            } catch (InterruptedException e) {
                backgroundExecutor.shutdownNow();
                Thread.currentThread().interrupt();
            }
        }
        
        if (networkExecutor != null && !networkExecutor.isShutdown()) {
            networkExecutor.shutdown();
            try {
                if (!networkExecutor.awaitTermination(5, TimeUnit.SECONDS)) {
                    networkExecutor.shutdownNow();
                }
            } catch (InterruptedException e) {
                networkExecutor.shutdownNow();
                Thread.currentThread().interrupt();
            }
        }
        
        listeners.clear();
        performanceMetrics.clear();
        
        Log.d(TAG, "性能优化器资源清理完成");
    }
    
    // 内部类和接口
    
    /**
     * 内存状态类
     */
    public static class MemoryStatus {
        public long availableMemory;
        public long totalMemory;
        public long usedMemory;
        public int memoryPercentage;
        public boolean isLowMemory;
        public MemoryLevel level;
    }
    
    /**
     * 电池状态类
     */
    public static class BatteryStatus {
        public int batteryLevel;
        public boolean isCharging;
        public int temperature;
        public BatteryLevel level;
    }
    
    /**
     * 内存级别枚举
     */
    public enum MemoryLevel {
        NORMAL, LOW, CRITICAL
    }
    
    /**
     * 电池级别枚举
     */
    public enum BatteryLevel {
        NORMAL, LOW, CRITICAL, UNKNOWN
    }
    
    /**
     * 性能监听器接口
     */
    public interface PerformanceListener {
        void onPerformanceEvent(String eventName, long timestamp);
    }
}