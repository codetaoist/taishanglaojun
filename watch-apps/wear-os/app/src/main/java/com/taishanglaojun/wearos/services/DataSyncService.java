package com.taishanglaojun.wearos.services;

import android.app.Service;
import android.content.Intent;
import android.os.IBinder;
import android.util.Log;
import android.os.Handler;
import android.os.Looper;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.List;
import java.util.ArrayList;
import java.util.concurrent.TimeUnit;

/**
 * 数据同步服务
 * 负责与手机应用和云端服务器同步数据
 */
public class DataSyncService extends Service {
    
    private static final String TAG = "DataSyncService";
    private static final long SYNC_INTERVAL = TimeUnit.MINUTES.toMillis(5); // 5分钟同步一次
    
    private ExecutorService executorService;
    private Handler mainHandler;
    private Runnable syncRunnable;
    private boolean isSyncServiceRunning = false;
    
    // 模拟数据存储
    private List<SyncData> pendingSyncData;
    
    @Override
    public void onCreate() {
        super.onCreate();
        Log.d(TAG, "数据同步服务创建");
        
        executorService = Executors.newSingleThreadExecutor();
        mainHandler = new Handler(Looper.getMainLooper());
        pendingSyncData = new ArrayList<>();
        
        initializeSyncRunnable();
    }
    
    /**
     * 初始化同步任务
     */
    private void initializeSyncRunnable() {
        syncRunnable = new Runnable() {
            @Override
            public void run() {
                if (isSyncServiceRunning) {
                    performDataSync();
                    // 安排下次同步
                    mainHandler.postDelayed(this, SYNC_INTERVAL);
                }
            }
        };
    }
    
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.d(TAG, "数据同步服务启动");
        
        startDataSync();
        
        return START_STICKY; // 服务被杀死后自动重启
    }
    
    /**
     * 开始数据同步
     */
    private void startDataSync() {
        if (isSyncServiceRunning) {
            Log.d(TAG, "数据同步服务已在运行");
            return;
        }
        
        isSyncServiceRunning = true;
        
        // 立即执行一次同步
        mainHandler.post(syncRunnable);
        
        Log.d(TAG, "数据同步已启动，同步间隔: " + SYNC_INTERVAL / 1000 + " 秒");
    }
    
    /**
     * 停止数据同步
     */
    private void stopDataSync() {
        if (isSyncServiceRunning) {
            isSyncServiceRunning = false;
            mainHandler.removeCallbacks(syncRunnable);
            Log.d(TAG, "数据同步已停止");
        }
    }
    
    /**
     * 执行数据同步
     */
    private void performDataSync() {
        executorService.execute(new Runnable() {
            @Override
            public void run() {
                try {
                    Log.d(TAG, "开始数据同步...");
                    
                    // 同步位置数据
                    syncLocationData();
                    
                    // 同步健康数据
                    syncHealthData();
                    
                    // 同步应用数据
                    syncAppData();
                    
                    // 清理已同步的数据
                    cleanupSyncedData();
                    
                    Log.d(TAG, "数据同步完成");
                    
                } catch (Exception e) {
                    Log.e(TAG, "数据同步失败: " + e.getMessage(), e);
                }
            }
        });
    }
    
    /**
     * 同步位置数据
     */
    private void syncLocationData() {
        Log.d(TAG, "同步位置数据...");
        
        // 模拟位置数据同步
        List<SyncData> locationData = getLocationDataForSync();
        
        for (SyncData data : locationData) {
            if (uploadDataToServer(data)) {
                data.setSynced(true);
                Log.d(TAG, "位置数据同步成功: " + data.getId());
            } else {
                Log.w(TAG, "位置数据同步失败: " + data.getId());
            }
        }
    }
    
    /**
     * 同步健康数据
     */
    private void syncHealthData() {
        Log.d(TAG, "同步健康数据...");
        
        // 模拟健康数据同步
        List<SyncData> healthData = getHealthDataForSync();
        
        for (SyncData data : healthData) {
            if (uploadDataToServer(data)) {
                data.setSynced(true);
                Log.d(TAG, "健康数据同步成功: " + data.getId());
            } else {
                Log.w(TAG, "健康数据同步失败: " + data.getId());
            }
        }
    }
    
    /**
     * 同步应用数据
     */
    private void syncAppData() {
        Log.d(TAG, "同步应用数据...");
        
        // 模拟应用数据同步
        List<SyncData> appData = getAppDataForSync();
        
        for (SyncData data : appData) {
            if (uploadDataToServer(data)) {
                data.setSynced(true);
                Log.d(TAG, "应用数据同步成功: " + data.getId());
            } else {
                Log.w(TAG, "应用数据同步失败: " + data.getId());
            }
        }
    }
    
    /**
     * 获取待同步的位置数据
     */
    private List<SyncData> getLocationDataForSync() {
        List<SyncData> locationData = new ArrayList<>();
        
        // 模拟位置数据
        locationData.add(new SyncData("loc_001", "location", "纬度: 39.9042, 经度: 116.4074", System.currentTimeMillis()));
        locationData.add(new SyncData("loc_002", "location", "纬度: 39.9043, 经度: 116.4075", System.currentTimeMillis()));
        
        return locationData;
    }
    
    /**
     * 获取待同步的健康数据
     */
    private List<SyncData> getHealthDataForSync() {
        List<SyncData> healthData = new ArrayList<>();
        
        // 模拟健康数据
        healthData.add(new SyncData("health_001", "heart_rate", "心率: 72 BPM", System.currentTimeMillis()));
        healthData.add(new SyncData("health_002", "step_count", "步数: 8500", System.currentTimeMillis()));
        
        return healthData;
    }
    
    /**
     * 获取待同步的应用数据
     */
    private List<SyncData> getAppDataForSync() {
        List<SyncData> appData = new ArrayList<>();
        
        // 模拟应用数据
        appData.add(new SyncData("app_001", "settings", "设置更新", System.currentTimeMillis()));
        appData.add(new SyncData("app_002", "preferences", "用户偏好", System.currentTimeMillis()));
        
        return appData;
    }
    
    /**
     * 上传数据到服务器
     */
    private boolean uploadDataToServer(SyncData data) {
        try {
            // 模拟网络请求延迟
            Thread.sleep(100);
            
            // 模拟上传成功率 (90%)
            return Math.random() > 0.1;
            
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return false;
        }
    }
    
    /**
     * 清理已同步的数据
     */
    private void cleanupSyncedData() {
        int cleanedCount = 0;
        
        for (int i = pendingSyncData.size() - 1; i >= 0; i--) {
            SyncData data = pendingSyncData.get(i);
            if (data.isSynced()) {
                pendingSyncData.remove(i);
                cleanedCount++;
            }
        }
        
        if (cleanedCount > 0) {
            Log.d(TAG, "清理已同步数据: " + cleanedCount + " 条");
        }
    }
    
    /**
     * 添加待同步数据
     */
    public void addSyncData(SyncData data) {
        pendingSyncData.add(data);
        Log.d(TAG, "添加待同步数据: " + data.getId());
    }
    
    @Override
    public void onDestroy() {
        super.onDestroy();
        stopDataSync();
        
        if (executorService != null && !executorService.isShutdown()) {
            executorService.shutdown();
        }
        
        Log.d(TAG, "数据同步服务销毁");
    }
    
    @Override
    public IBinder onBind(Intent intent) {
        return null; // 不支持绑定
    }
    
    /**
     * 同步数据内部类
     */
    public static class SyncData {
        private final String id;
        private final String type;
        private final String data;
        private final long timestamp;
        private boolean synced;
        
        public SyncData(String id, String type, String data, long timestamp) {
            this.id = id;
            this.type = type;
            this.data = data;
            this.timestamp = timestamp;
            this.synced = false;
        }
        
        public String getId() {
            return id;
        }
        
        public String getType() {
            return type;
        }
        
        public String getData() {
            return data;
        }
        
        public long getTimestamp() {
            return timestamp;
        }
        
        public boolean isSynced() {
            return synced;
        }
        
        public void setSynced(boolean synced) {
            this.synced = synced;
        }
        
        @Override
        public String toString() {
            return String.format("SyncData{id='%s', type='%s', data='%s', timestamp=%d, synced=%b}",
                id, type, data, timestamp, synced);
        }
    }
}