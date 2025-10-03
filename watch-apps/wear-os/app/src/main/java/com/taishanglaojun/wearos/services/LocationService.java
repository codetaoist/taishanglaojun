package com.taishanglaojun.wearos.services;

import android.app.Service;
import android.content.Intent;
import android.location.Location;
import android.location.LocationListener;
import android.location.LocationManager;
import android.os.IBinder;
import android.util.Log;
import android.content.Context;
import android.os.Bundle;

/**
 * 位置追踪服务
 * 负责获取用户位置信息并上传到服务器
 */
public class LocationService extends Service implements LocationListener {
    
    private static final String TAG = "LocationService";
    private static final long MIN_TIME_BETWEEN_UPDATES = 10000; // 10秒
    private static final float MIN_DISTANCE_CHANGE_FOR_UPDATES = 10; // 10米
    
    private LocationManager locationManager;
    private boolean isLocationServiceRunning = false;
    
    @Override
    public void onCreate() {
        super.onCreate();
        Log.d(TAG, "位置服务创建");
        
        locationManager = (LocationManager) getSystemService(Context.LOCATION_SERVICE);
    }
    
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.d(TAG, "位置服务启动");
        
        startLocationTracking();
        
        return START_STICKY; // 服务被杀死后自动重启
    }
    
    /**
     * 开始位置追踪
     */
    private void startLocationTracking() {
        if (isLocationServiceRunning) {
            Log.d(TAG, "位置服务已在运行");
            return;
        }
        
        try {
            // 检查权限
            if (checkSelfPermission(android.Manifest.permission.ACCESS_FINE_LOCATION) 
                != android.content.pm.PackageManager.PERMISSION_GRANTED) {
                Log.e(TAG, "缺少位置权限");
                return;
            }
            
            // 请求位置更新
            if (locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER)) {
                locationManager.requestLocationUpdates(
                    LocationManager.GPS_PROVIDER,
                    MIN_TIME_BETWEEN_UPDATES,
                    MIN_DISTANCE_CHANGE_FOR_UPDATES,
                    this
                );
                isLocationServiceRunning = true;
                Log.d(TAG, "GPS位置追踪已启动");
            }
            
            if (locationManager.isProviderEnabled(LocationManager.NETWORK_PROVIDER)) {
                locationManager.requestLocationUpdates(
                    LocationManager.NETWORK_PROVIDER,
                    MIN_TIME_BETWEEN_UPDATES,
                    MIN_DISTANCE_CHANGE_FOR_UPDATES,
                    this
                );
                Log.d(TAG, "网络位置追踪已启动");
            }
            
        } catch (SecurityException e) {
            Log.e(TAG, "位置权限异常: " + e.getMessage());
        }
    }
    
    /**
     * 停止位置追踪
     */
    private void stopLocationTracking() {
        if (locationManager != null && isLocationServiceRunning) {
            try {
                locationManager.removeUpdates(this);
                isLocationServiceRunning = false;
                Log.d(TAG, "位置追踪已停止");
            } catch (SecurityException e) {
                Log.e(TAG, "停止位置追踪异常: " + e.getMessage());
            }
        }
    }
    
    @Override
    public void onLocationChanged(Location location) {
        Log.d(TAG, "位置更新: " + location.getLatitude() + ", " + location.getLongitude());
        
        // 处理位置数据
        processLocationData(location);
    }
    
    /**
     * 处理位置数据
     */
    private void processLocationData(Location location) {
        // 创建位置数据对象
        LocationData locationData = new LocationData(
            location.getLatitude(),
            location.getLongitude(),
            location.getAltitude(),
            location.getAccuracy(),
            System.currentTimeMillis()
        );
        
        // 上传到服务器
        uploadLocationData(locationData);
    }
    
    /**
     * 上传位置数据到服务器
     */
    private void uploadLocationData(LocationData locationData) {
        // TODO: 实现数据上传逻辑
        Log.d(TAG, "位置数据准备上传: " + locationData.toString());
    }
    
    @Override
    public void onStatusChanged(String provider, int status, Bundle extras) {
        Log.d(TAG, "位置提供者状态变化: " + provider + ", 状态: " + status);
    }
    
    @Override
    public void onProviderEnabled(String provider) {
        Log.d(TAG, "位置提供者启用: " + provider);
    }
    
    @Override
    public void onProviderDisabled(String provider) {
        Log.d(TAG, "位置提供者禁用: " + provider);
    }
    
    @Override
    public void onDestroy() {
        super.onDestroy();
        stopLocationTracking();
        Log.d(TAG, "位置服务销毁");
    }
    
    @Override
    public IBinder onBind(Intent intent) {
        return null; // 不支持绑定
    }
    
    /**
     * 位置数据内部类
     */
    private static class LocationData {
        public final double latitude;
        public final double longitude;
        public final double altitude;
        public final float accuracy;
        public final long timestamp;
        
        public LocationData(double lat, double lng, double alt, float acc, long time) {
            this.latitude = lat;
            this.longitude = lng;
            this.altitude = alt;
            this.accuracy = acc;
            this.timestamp = time;
        }
        
        @Override
        public String toString() {
            return String.format("LocationData{lat=%.6f, lng=%.6f, alt=%.2f, acc=%.2f, time=%d}",
                latitude, longitude, altitude, accuracy, timestamp);
        }
    }
}