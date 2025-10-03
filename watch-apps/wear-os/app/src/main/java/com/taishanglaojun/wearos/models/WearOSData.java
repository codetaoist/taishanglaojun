package com.taishanglaojun.wearos.models;

import java.util.Date;
import java.util.Map;
import java.util.HashMap;

/**
 * Wear OS 数据模型
 * 统一的数据结构，用于在应用内传递各种类型的数据
 */
public class WearOSData {
    
    // 数据类型常量
    public static final String TYPE_LOCATION = "location";
    public static final String TYPE_HEALTH = "health";
    public static final String TYPE_SENSOR = "sensor";
    public static final String TYPE_NOTIFICATION = "notification";
    public static final String TYPE_APP_STATE = "app_state";
    
    // 健康数据子类型
    public static final String HEALTH_HEART_RATE = "heart_rate";
    public static final String HEALTH_STEP_COUNT = "step_count";
    public static final String HEALTH_ACTIVITY_LEVEL = "activity_level";
    public static final String HEALTH_SLEEP_DATA = "sleep_data";
    
    // 传感器数据子类型
    public static final String SENSOR_ACCELEROMETER = "accelerometer";
    public static final String SENSOR_GYROSCOPE = "gyroscope";
    public static final String SENSOR_AMBIENT_LIGHT = "ambient_light";
    public static final String SENSOR_PRESSURE = "pressure";
    
    private String id;
    private String type;
    private String subType;
    private Map<String, Object> data;
    private Date timestamp;
    private String deviceId;
    private boolean synced;
    private int priority;
    
    /**
     * 默认构造函数
     */
    public WearOSData() {
        this.data = new HashMap<>();
        this.timestamp = new Date();
        this.synced = false;
        this.priority = 0;
    }
    
    /**
     * 构造函数
     */
    public WearOSData(String id, String type, String subType) {
        this();
        this.id = id;
        this.type = type;
        this.subType = subType;
    }
    
    /**
     * 构造函数 - 带数据
     */
    public WearOSData(String id, String type, String subType, Map<String, Object> data) {
        this(id, type, subType);
        if (data != null) {
            this.data.putAll(data);
        }
    }
    
    // Getter 和 Setter 方法
    
    public String getId() {
        return id;
    }
    
    public void setId(String id) {
        this.id = id;
    }
    
    public String getType() {
        return type;
    }
    
    public void setType(String type) {
        this.type = type;
    }
    
    public String getSubType() {
        return subType;
    }
    
    public void setSubType(String subType) {
        this.subType = subType;
    }
    
    public Map<String, Object> getData() {
        return data;
    }
    
    public void setData(Map<String, Object> data) {
        this.data = data != null ? data : new HashMap<>();
    }
    
    public Date getTimestamp() {
        return timestamp;
    }
    
    public void setTimestamp(Date timestamp) {
        this.timestamp = timestamp;
    }
    
    public String getDeviceId() {
        return deviceId;
    }
    
    public void setDeviceId(String deviceId) {
        this.deviceId = deviceId;
    }
    
    public boolean isSynced() {
        return synced;
    }
    
    public void setSynced(boolean synced) {
        this.synced = synced;
    }
    
    public int getPriority() {
        return priority;
    }
    
    public void setPriority(int priority) {
        this.priority = priority;
    }
    
    // 数据操作方法
    
    /**
     * 添加数据项
     */
    public void addData(String key, Object value) {
        this.data.put(key, value);
    }
    
    /**
     * 获取数据项
     */
    public Object getData(String key) {
        return this.data.get(key);
    }
    
    /**
     * 获取字符串数据项
     */
    public String getStringData(String key) {
        Object value = this.data.get(key);
        return value != null ? value.toString() : null;
    }
    
    /**
     * 获取数值数据项
     */
    public Double getNumericData(String key) {
        Object value = this.data.get(key);
        if (value instanceof Number) {
            return ((Number) value).doubleValue();
        }
        return null;
    }
    
    /**
     * 获取布尔数据项
     */
    public Boolean getBooleanData(String key) {
        Object value = this.data.get(key);
        if (value instanceof Boolean) {
            return (Boolean) value;
        }
        return null;
    }
    
    /**
     * 检查是否包含指定键
     */
    public boolean containsKey(String key) {
        return this.data.containsKey(key);
    }
    
    /**
     * 移除数据项
     */
    public Object removeData(String key) {
        return this.data.remove(key);
    }
    
    /**
     * 清空所有数据
     */
    public void clearData() {
        this.data.clear();
    }
    
    /**
     * 获取数据项数量
     */
    public int getDataSize() {
        return this.data.size();
    }
    
    // 工厂方法
    
    /**
     * 创建位置数据
     */
    public static WearOSData createLocationData(String id, double latitude, double longitude, float accuracy) {
        WearOSData data = new WearOSData(id, TYPE_LOCATION, null);
        data.addData("latitude", latitude);
        data.addData("longitude", longitude);
        data.addData("accuracy", accuracy);
        return data;
    }
    
    /**
     * 创建心率数据
     */
    public static WearOSData createHeartRateData(String id, float heartRate) {
        WearOSData data = new WearOSData(id, TYPE_HEALTH, HEALTH_HEART_RATE);
        data.addData("heart_rate", heartRate);
        data.addData("unit", "BPM");
        return data;
    }
    
    /**
     * 创建步数数据
     */
    public static WearOSData createStepCountData(String id, int stepCount) {
        WearOSData data = new WearOSData(id, TYPE_HEALTH, HEALTH_STEP_COUNT);
        data.addData("step_count", stepCount);
        return data;
    }
    
    /**
     * 创建活动强度数据
     */
    public static WearOSData createActivityLevelData(String id, float activityLevel) {
        WearOSData data = new WearOSData(id, TYPE_HEALTH, HEALTH_ACTIVITY_LEVEL);
        data.addData("activity_level", activityLevel);
        return data;
    }
    
    /**
     * 创建加速度传感器数据
     */
    public static WearOSData createAccelerometerData(String id, float x, float y, float z) {
        WearOSData data = new WearOSData(id, TYPE_SENSOR, SENSOR_ACCELEROMETER);
        data.addData("x", x);
        data.addData("y", y);
        data.addData("z", z);
        return data;
    }
    
    /**
     * 创建通知数据
     */
    public static WearOSData createNotificationData(String id, String title, String content, String packageName) {
        WearOSData data = new WearOSData(id, TYPE_NOTIFICATION, null);
        data.addData("title", title);
        data.addData("content", content);
        data.addData("package_name", packageName);
        return data;
    }
    
    /**
     * 创建应用状态数据
     */
    public static WearOSData createAppStateData(String id, String state, Map<String, Object> stateData) {
        WearOSData data = new WearOSData(id, TYPE_APP_STATE, state);
        if (stateData != null) {
            data.getData().putAll(stateData);
        }
        return data;
    }
    
    // 验证方法
    
    /**
     * 验证数据完整性
     */
    public boolean isValid() {
        return id != null && !id.isEmpty() && 
               type != null && !type.isEmpty() && 
               timestamp != null;
    }
    
    /**
     * 检查是否为位置数据
     */
    public boolean isLocationData() {
        return TYPE_LOCATION.equals(type);
    }
    
    /**
     * 检查是否为健康数据
     */
    public boolean isHealthData() {
        return TYPE_HEALTH.equals(type);
    }
    
    /**
     * 检查是否为传感器数据
     */
    public boolean isSensorData() {
        return TYPE_SENSOR.equals(type);
    }
    
    /**
     * 检查是否为通知数据
     */
    public boolean isNotificationData() {
        return TYPE_NOTIFICATION.equals(type);
    }
    
    /**
     * 检查是否为应用状态数据
     */
    public boolean isAppStateData() {
        return TYPE_APP_STATE.equals(type);
    }
    
    @Override
    public String toString() {
        return String.format("WearOSData{id='%s', type='%s', subType='%s', dataSize=%d, timestamp=%s, synced=%b, priority=%d}",
            id, type, subType, data.size(), timestamp, synced, priority);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (obj == null || getClass() != obj.getClass()) return false;
        
        WearOSData that = (WearOSData) obj;
        return id != null ? id.equals(that.id) : that.id == null;
    }
    
    @Override
    public int hashCode() {
        return id != null ? id.hashCode() : 0;
    }
}