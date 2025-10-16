package com.taishanglaojun.wearos.services;

import android.app.Service;
import android.content.Intent;
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.os.IBinder;
import android.util.Log;
import android.content.Context;

/**
 * 健康监测服务
 * 负责监测心率、步数、活动等健康数据
 */
public class HealthService extends Service implements SensorEventListener {
    
    private static final String TAG = "HealthService";
    
    private SensorManager sensorManager;
    private Sensor heartRateSensor;
    private Sensor stepCounterSensor;
    private Sensor accelerometerSensor;
    
    private boolean isHealthServiceRunning = false;
    private int stepCount = 0;
    private float heartRate = 0.0f;
    
    @Override
    public void onCreate() {
        super.onCreate();
        Log.d(TAG, "健康监测服务创建");
        
        sensorManager = (SensorManager) getSystemService(Context.SENSOR_SERVICE);
        initializeSensors();
    }
    
    /**
     * 初始化传感器
     */
    private void initializeSensors() {
        // 心率传感器
        heartRateSensor = sensorManager.getDefaultSensor(Sensor.TYPE_HEART_RATE);
        if (heartRateSensor == null) {
            Log.w(TAG, "心率传感器不可用");
        }
        
        // 步数传感器
        stepCounterSensor = sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER);
        if (stepCounterSensor == null) {
            Log.w(TAG, "步数传感器不可用");
        }
        
        // 加速度传感器
        accelerometerSensor = sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER);
        if (accelerometerSensor == null) {
            Log.w(TAG, "加速度传感器不可用");
        }
    }
    
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.d(TAG, "健康监测服务启动");
        
        startHealthMonitoring();
        
        return START_STICKY; // 服务被杀死后自动重启
    }
    
    /**
     * 开始健康监测
     */
    private void startHealthMonitoring() {
        if (isHealthServiceRunning) {
            Log.d(TAG, "健康监测服务已在运行");
            return;
        }
        
        // 注册心率传感器监听器
        if (heartRateSensor != null) {
            sensorManager.registerListener(this, heartRateSensor, SensorManager.SENSOR_DELAY_NORMAL);
            Log.d(TAG, "心率监测已启动");
        }
        
        // 注册步数传感器监听器
        if (stepCounterSensor != null) {
            sensorManager.registerListener(this, stepCounterSensor, SensorManager.SENSOR_DELAY_NORMAL);
            Log.d(TAG, "步数监测已启动");
        }
        
        // 注册加速度传感器监听器
        if (accelerometerSensor != null) {
            sensorManager.registerListener(this, accelerometerSensor, SensorManager.SENSOR_DELAY_NORMAL);
            Log.d(TAG, "活动监测已启动");
        }
        
        isHealthServiceRunning = true;
    }
    
    /**
     * 停止健康监测
     */
    private void stopHealthMonitoring() {
        if (sensorManager != null && isHealthServiceRunning) {
            sensorManager.unregisterListener(this);
            isHealthServiceRunning = false;
            Log.d(TAG, "健康监测已停止");
        }
    }
    
    @Override
    public void onSensorChanged(SensorEvent event) {
        switch (event.sensor.getType()) {
            case Sensor.TYPE_HEART_RATE:
                handleHeartRateData(event);
                break;
            case Sensor.TYPE_STEP_COUNTER:
                handleStepCountData(event);
                break;
            case Sensor.TYPE_ACCELEROMETER:
                handleAccelerometerData(event);
                break;
        }
    }
    
    /**
     * 处理心率数据
     */
    private void handleHeartRateData(SensorEvent event) {
        heartRate = event.values[0];
        Log.d(TAG, "心率数据: " + heartRate + " BPM");
        
        // 创建健康数据对象
        HealthData healthData = new HealthData(
            HealthData.TYPE_HEART_RATE,
            heartRate,
            System.currentTimeMillis()
        );
        
        // 上传健康数据
        uploadHealthData(healthData);
    }
    
    /**
     * 处理步数数据
     */
    private void handleStepCountData(SensorEvent event) {
        stepCount = (int) event.values[0];
        Log.d(TAG, "步数数据: " + stepCount);
        
        // 创建健康数据对象
        HealthData healthData = new HealthData(
            HealthData.TYPE_STEP_COUNT,
            stepCount,
            System.currentTimeMillis()
        );
        
        // 上传健康数据
        uploadHealthData(healthData);
    }
    
    /**
     * 处理加速度数据
     */
    private void handleAccelerometerData(SensorEvent event) {
        float x = event.values[0];
        float y = event.values[1];
        float z = event.values[2];
        
        // 计算活动强度
        float activityLevel = (float) Math.sqrt(x * x + y * y + z * z);
        
        Log.d(TAG, "活动强度: " + activityLevel);
        
        // 创建健康数据对象
        HealthData healthData = new HealthData(
            HealthData.TYPE_ACTIVITY_LEVEL,
            activityLevel,
            System.currentTimeMillis()
        );
        
        // 上传健康数据
        uploadHealthData(healthData);
    }
    
    /**
     * 上传健康数据到服务器
     */
    private void uploadHealthData(HealthData healthData) {
        // TODO: 实现数据上传逻辑
        Log.d(TAG, "健康数据准备上传: " + healthData.toString());
    }
    
    @Override
    public void onAccuracyChanged(Sensor sensor, int accuracy) {
        Log.d(TAG, "传感器精度变化: " + sensor.getName() + ", 精度: " + accuracy);
    }
    
    @Override
    public void onDestroy() {
        super.onDestroy();
        stopHealthMonitoring();
        Log.d(TAG, "健康监测服务销毁");
    }
    
    @Override
    public IBinder onBind(Intent intent) {
        return null; // 不支持绑定
    }
    
    /**
     * 健康数据内部类
     */
    private static class HealthData {
        public static final int TYPE_HEART_RATE = 1;
        public static final int TYPE_STEP_COUNT = 2;
        public static final int TYPE_ACTIVITY_LEVEL = 3;
        
        public final int type;
        public final float value;
        public final long timestamp;
        
        public HealthData(int type, float value, long timestamp) {
            this.type = type;
            this.value = value;
            this.timestamp = timestamp;
        }
        
        @Override
        public String toString() {
            String typeName = "";
            switch (type) {
                case TYPE_HEART_RATE:
                    typeName = "心率";
                    break;
                case TYPE_STEP_COUNT:
                    typeName = "步数";
                    break;
                case TYPE_ACTIVITY_LEVEL:
                    typeName = "活动强度";
                    break;
            }
            return String.format("HealthData{type=%s, value=%.2f, time=%d}",
                typeName, value, timestamp);
        }
    }
}