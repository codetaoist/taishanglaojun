package com.taishanglaojun.wearos.utils;

import android.content.Context;
import android.content.pm.PackageManager;
import android.hardware.Sensor;
import android.hardware.SensorManager;
import android.os.Build;
import android.util.Log;
import android.provider.Settings;
import android.content.Intent;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.Locale;
import java.util.UUID;

/**
 * Wear OS 工具类
 * 提供各种实用工具方法
 */
public class WearOSUtils {
    
    private static final String TAG = "WearOSUtils";
    
    // 日期格式化器
    private static final SimpleDateFormat DATE_FORMAT = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault());
    private static final SimpleDateFormat TIME_FORMAT = new SimpleDateFormat("HH:mm:ss", Locale.getDefault());
    
    /**
     * 私有构造函数，防止实例化
     */
    private WearOSUtils() {
        throw new AssertionError("工具类不应被实例化");
    }
    
    // 设备信息相关方法
    
    /**
     * 获取设备唯一标识符
     */
    public static String getDeviceId(Context context) {
        try {
            String androidId = Settings.Secure.getString(context.getContentResolver(), Settings.Secure.ANDROID_ID);
            if (androidId != null && !androidId.isEmpty() && !"9774d56d682e549c".equals(androidId)) {
                return androidId;
            }
        } catch (Exception e) {
            Log.w(TAG, "无法获取Android ID: " + e.getMessage());
        }
        
        // 如果无法获取Android ID，生成一个随机UUID
        return UUID.randomUUID().toString();
    }
    
    /**
     * 获取设备型号
     */
    public static String getDeviceModel() {
        return Build.MODEL;
    }
    
    /**
     * 获取Android版本
     */
    public static String getAndroidVersion() {
        return Build.VERSION.RELEASE;
    }
    
    /**
     * 获取API级别
     */
    public static int getApiLevel() {
        return Build.VERSION.SDK_INT;
    }
    
    /**
     * 检查是否为Wear OS设备
     */
    public static boolean isWearOSDevice(Context context) {
        PackageManager pm = context.getPackageManager();
        return pm.hasSystemFeature(PackageManager.FEATURE_WATCH);
    }
    
    // 传感器相关方法
    
    /**
     * 检查是否支持心率传感器
     */
    public static boolean hasHeartRateSensor(Context context) {
        SensorManager sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        return sensorManager != null && sensorManager.getDefaultSensor(Sensor.TYPE_HEART_RATE) != null;
    }
    
    /**
     * 检查是否支持步数传感器
     */
    public static boolean hasStepCounterSensor(Context context) {
        SensorManager sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        return sensorManager != null && sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER) != null;
    }
    
    /**
     * 检查是否支持加速度传感器
     */
    public static boolean hasAccelerometerSensor(Context context) {
        SensorManager sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        return sensorManager != null && sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER) != null;
    }
    
    /**
     * 检查是否支持陀螺仪传感器
     */
    public static boolean hasGyroscopeSensor(Context context) {
        SensorManager sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        return sensorManager != null && sensorManager.getDefaultSensor(Sensor.TYPE_GYROSCOPE) != null;
    }
    
    /**
     * 检查是否支持环境光传感器
     */
    public static boolean hasAmbientLightSensor(Context context) {
        SensorManager sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        return sensorManager != null && sensorManager.getDefaultSensor(Sensor.TYPE_LIGHT) != null;
    }
    
    /**
     * 获取可用传感器列表
     */
    public static String getAvailableSensors(Context context) {
        SensorManager sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
        if (sensorManager == null) {
            return "传感器管理器不可用";
        }
        
        StringBuilder sensors = new StringBuilder();
        
        if (hasHeartRateSensor(context)) {
            sensors.append("心率传感器, ");
        }
        if (hasStepCounterSensor(context)) {
            sensors.append("步数传感器, ");
        }
        if (hasAccelerometerSensor(context)) {
            sensors.append("加速度传感器, ");
        }
        if (hasGyroscopeSensor(context)) {
            sensors.append("陀螺仪传感器, ");
        }
        if (hasAmbientLightSensor(context)) {
            sensors.append("环境光传感器, ");
        }
        
        String result = sensors.toString();
        return result.isEmpty() ? "无可用传感器" : result.substring(0, result.length() - 2);
    }
    
    // 权限相关方法
    
    /**
     * 检查是否有位置权限
     */
    public static boolean hasLocationPermission(Context context) {
        return context.checkSelfPermission(android.Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED ||
               context.checkSelfPermission(android.Manifest.permission.ACCESS_COARSE_LOCATION) == PackageManager.PERMISSION_GRANTED;
    }
    
    /**
     * 检查是否有身体传感器权限
     */
    public static boolean hasBodySensorsPermission(Context context) {
        return context.checkSelfPermission(android.Manifest.permission.BODY_SENSORS) == PackageManager.PERMISSION_GRANTED;
    }
    
    /**
     * 检查是否有活动识别权限
     */
    public static boolean hasActivityRecognitionPermission(Context context) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            return context.checkSelfPermission(android.Manifest.permission.ACTIVITY_RECOGNITION) == PackageManager.PERMISSION_GRANTED;
        }
        return true; // Android Q以下版本不需要此权限
    }
    
    // 网络相关方法
    
    /**
     * 检查网络连接状态
     */
    public static boolean isNetworkAvailable(Context context) {
        ConnectivityManager connectivityManager = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
        if (connectivityManager == null) {
            return false;
        }
        
        NetworkInfo activeNetworkInfo = connectivityManager.getActiveNetworkInfo();
        return activeNetworkInfo != null && activeNetworkInfo.isConnected();
    }
    
    /**
     * 检查是否连接到WiFi
     */
    public static boolean isWiFiConnected(Context context) {
        ConnectivityManager connectivityManager = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
        if (connectivityManager == null) {
            return false;
        }
        
        NetworkInfo wifiInfo = connectivityManager.getNetworkInfo(ConnectivityManager.TYPE_WIFI);
        return wifiInfo != null && wifiInfo.isConnected();
    }
    
    /**
     * 检查是否连接到移动网络
     */
    public static boolean isMobileConnected(Context context) {
        ConnectivityManager connectivityManager = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
        if (connectivityManager == null) {
            return false;
        }
        
        NetworkInfo mobileInfo = connectivityManager.getNetworkInfo(ConnectivityManager.TYPE_MOBILE);
        return mobileInfo != null && mobileInfo.isConnected();
    }
    
    // 时间和日期相关方法
    
    /**
     * 格式化时间戳为日期时间字符串
     */
    public static String formatDateTime(long timestamp) {
        return DATE_FORMAT.format(new Date(timestamp));
    }
    
    /**
     * 格式化时间戳为时间字符串
     */
    public static String formatTime(long timestamp) {
        return TIME_FORMAT.format(new Date(timestamp));
    }
    
    /**
     * 获取当前时间戳
     */
    public static long getCurrentTimestamp() {
        return System.currentTimeMillis();
    }
    
    /**
     * 获取当前日期时间字符串
     */
    public static String getCurrentDateTime() {
        return formatDateTime(getCurrentTimestamp());
    }
    
    /**
     * 获取当前时间字符串
     */
    public static String getCurrentTime() {
        return formatTime(getCurrentTimestamp());
    }
    
    // 数据处理相关方法
    
    /**
     * 生成唯一ID
     */
    public static String generateUniqueId() {
        return UUID.randomUUID().toString();
    }
    
    /**
     * 生成带前缀的唯一ID
     */
    public static String generateUniqueId(String prefix) {
        return prefix + "_" + UUID.randomUUID().toString().replace("-", "").substring(0, 8);
    }
    
    /**
     * 安全地转换字符串为整数
     */
    public static int safeParseInt(String value, int defaultValue) {
        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }
    
    /**
     * 安全地转换字符串为浮点数
     */
    public static float safeParseFloat(String value, float defaultValue) {
        try {
            return Float.parseFloat(value);
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }
    
    /**
     * 安全地转换字符串为双精度浮点数
     */
    public static double safeParseDouble(String value, double defaultValue) {
        try {
            return Double.parseDouble(value);
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }
    
    /**
     * 检查字符串是否为空或null
     */
    public static boolean isEmpty(String str) {
        return str == null || str.trim().isEmpty();
    }
    
    /**
     * 安全地获取字符串，如果为null则返回默认值
     */
    public static String safeString(String str, String defaultValue) {
        return isEmpty(str) ? defaultValue : str;
    }
    
    // 日志相关方法
    
    /**
     * 记录调试日志
     */
    public static void logDebug(String tag, String message) {
        Log.d(tag, message);
    }
    
    /**
     * 记录信息日志
     */
    public static void logInfo(String tag, String message) {
        Log.i(tag, message);
    }
    
    /**
     * 记录警告日志
     */
    public static void logWarning(String tag, String message) {
        Log.w(tag, message);
    }
    
    /**
     * 记录错误日志
     */
    public static void logError(String tag, String message, Throwable throwable) {
        Log.e(tag, message, throwable);
    }
    
    // 应用相关方法
    
    /**
     * 启动服务
     */
    public static void startService(Context context, Class<?> serviceClass) {
        try {
            Intent intent = new Intent(context, serviceClass);
            context.startService(intent);
            Log.d(TAG, "服务启动: " + serviceClass.getSimpleName());
        } catch (Exception e) {
            Log.e(TAG, "启动服务失败: " + serviceClass.getSimpleName(), e);
        }
    }
    
    /**
     * 停止服务
     */
    public static void stopService(Context context, Class<?> serviceClass) {
        try {
            Intent intent = new Intent(context, serviceClass);
            context.stopService(intent);
            Log.d(TAG, "服务停止: " + serviceClass.getSimpleName());
        } catch (Exception e) {
            Log.e(TAG, "停止服务失败: " + serviceClass.getSimpleName(), e);
        }
    }
}