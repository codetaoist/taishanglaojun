package com.taishanglaojun.wearos;

import com.taishanglaojun.wearos.models.WearOSData;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.junit.MockitoJUnitRunner;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.*;

/**
 * WearOSData 模型单元测试
 */
@RunWith(MockitoJUnitRunner.class)
public class WearOSDataTest {

    private WearOSData wearOSData;

    @Before
    public void setUp() {
        wearOSData = new WearOSData();
    }

    @Test
    public void testDefaultConstructor() {
        assertNotNull("数据映射不应为null", wearOSData.getData());
        assertNotNull("时间戳不应为null", wearOSData.getTimestamp());
        assertFalse("默认应未同步", wearOSData.isSynced());
        assertEquals("默认优先级应为0", 0, wearOSData.getPriority());
        assertEquals("默认数据大小应为0", 0, wearOSData.getDataSize());
    }

    @Test
    public void testConstructorWithParameters() {
        String id = "test_id";
        String type = "test_type";
        String subType = "test_subtype";
        
        WearOSData data = new WearOSData(id, type, subType);
        
        assertEquals("ID应正确设置", id, data.getId());
        assertEquals("类型应正确设置", type, data.getType());
        assertEquals("子类型应正确设置", subType, data.getSubType());
        assertNotNull("数据映射不应为null", data.getData());
        assertNotNull("时间戳不应为null", data.getTimestamp());
    }

    @Test
    public void testConstructorWithData() {
        String id = "test_id";
        String type = "test_type";
        String subType = "test_subtype";
        Map<String, Object> testData = new HashMap<>();
        testData.put("key1", "value1");
        testData.put("key2", 123);
        
        WearOSData data = new WearOSData(id, type, subType, testData);
        
        assertEquals("ID应正确设置", id, data.getId());
        assertEquals("类型应正确设置", type, data.getType());
        assertEquals("子类型应正确设置", subType, data.getSubType());
        assertEquals("数据大小应正确", 2, data.getDataSize());
        assertEquals("应包含测试数据", "value1", data.getData("key1"));
        assertEquals("应包含测试数据", 123, data.getData("key2"));
    }

    @Test
    public void testDataOperations() {
        // 添加数据
        wearOSData.addData("string_key", "string_value");
        wearOSData.addData("int_key", 42);
        wearOSData.addData("double_key", 3.14);
        wearOSData.addData("boolean_key", true);
        
        assertEquals("数据大小应为4", 4, wearOSData.getDataSize());
        
        // 获取数据
        assertEquals("字符串数据应正确", "string_value", wearOSData.getData("string_key"));
        assertEquals("整数数据应正确", 42, wearOSData.getData("int_key"));
        assertEquals("双精度数据应正确", 3.14, wearOSData.getData("double_key"));
        assertEquals("布尔数据应正确", true, wearOSData.getData("boolean_key"));
        
        // 检查键存在
        assertTrue("应包含字符串键", wearOSData.containsKey("string_key"));
        assertTrue("应包含整数键", wearOSData.containsKey("int_key"));
        assertFalse("不应包含不存在的键", wearOSData.containsKey("nonexistent_key"));
        
        // 移除数据
        Object removed = wearOSData.removeData("string_key");
        assertEquals("移除的数据应正确", "string_value", removed);
        assertEquals("数据大小应减少", 3, wearOSData.getDataSize());
        assertFalse("不应再包含已移除的键", wearOSData.containsKey("string_key"));
        
        // 清空数据
        wearOSData.clearData();
        assertEquals("清空后数据大小应为0", 0, wearOSData.getDataSize());
    }

    @Test
    public void testTypedDataGetters() {
        wearOSData.addData("string_key", "test_string");
        wearOSData.addData("int_key", 123);
        wearOSData.addData("double_key", 45.67);
        wearOSData.addData("boolean_key", false);
        wearOSData.addData("null_key", null);
        
        // 字符串数据获取
        assertEquals("字符串数据应正确", "test_string", wearOSData.getStringData("string_key"));
        assertEquals("数字转字符串应正确", "123", wearOSData.getStringData("int_key"));
        assertNull("null值应返回null", wearOSData.getStringData("null_key"));
        assertNull("不存在的键应返回null", wearOSData.getStringData("nonexistent"));
        
        // 数值数据获取
        assertEquals("整数转双精度应正确", Double.valueOf(123), wearOSData.getNumericData("int_key"));
        assertEquals("双精度数据应正确", Double.valueOf(45.67), wearOSData.getNumericData("double_key"));
        assertNull("字符串不应转为数值", wearOSData.getNumericData("string_key"));
        assertNull("null值应返回null", wearOSData.getNumericData("null_key"));
        
        // 布尔数据获取
        assertEquals("布尔数据应正确", Boolean.FALSE, wearOSData.getBooleanData("boolean_key"));
        assertNull("字符串不应转为布尔", wearOSData.getBooleanData("string_key"));
        assertNull("null值应返回null", wearOSData.getBooleanData("null_key"));
    }

    @Test
    public void testFactoryMethods() {
        // 位置数据
        WearOSData locationData = WearOSData.createLocationData("loc_001", 39.9042, 116.4074, 10.5f);
        assertTrue("应为位置数据", locationData.isLocationData());
        assertEquals("类型应正确", WearOSData.TYPE_LOCATION, locationData.getType());
        assertEquals("纬度应正确", Double.valueOf(39.9042), locationData.getNumericData("latitude"));
        assertEquals("经度应正确", Double.valueOf(116.4074), locationData.getNumericData("longitude"));
        assertEquals("精度应正确", Double.valueOf(10.5), locationData.getNumericData("accuracy"));
        
        // 心率数据
        WearOSData heartRateData = WearOSData.createHeartRateData("hr_001", 72.5f);
        assertTrue("应为健康数据", heartRateData.isHealthData());
        assertEquals("类型应正确", WearOSData.TYPE_HEALTH, heartRateData.getType());
        assertEquals("子类型应正确", WearOSData.HEALTH_HEART_RATE, heartRateData.getSubType());
        assertEquals("心率值应正确", Double.valueOf(72.5), heartRateData.getNumericData("heart_rate"));
        assertEquals("单位应正确", "BPM", heartRateData.getStringData("unit"));
        
        // 步数数据
        WearOSData stepData = WearOSData.createStepCountData("step_001", 8500);
        assertTrue("应为健康数据", stepData.isHealthData());
        assertEquals("子类型应正确", WearOSData.HEALTH_STEP_COUNT, stepData.getSubType());
        assertEquals("步数应正确", Double.valueOf(8500), stepData.getNumericData("step_count"));
        
        // 活动强度数据
        WearOSData activityData = WearOSData.createActivityLevelData("act_001", 2.5f);
        assertTrue("应为健康数据", activityData.isHealthData());
        assertEquals("子类型应正确", WearOSData.HEALTH_ACTIVITY_LEVEL, activityData.getSubType());
        assertEquals("活动强度应正确", Double.valueOf(2.5), activityData.getNumericData("activity_level"));
        
        // 加速度传感器数据
        WearOSData accelData = WearOSData.createAccelerometerData("acc_001", 1.2f, -0.8f, 9.8f);
        assertTrue("应为传感器数据", accelData.isSensorData());
        assertEquals("类型应正确", WearOSData.TYPE_SENSOR, accelData.getType());
        assertEquals("子类型应正确", WearOSData.SENSOR_ACCELEROMETER, accelData.getSubType());
        assertEquals("X轴数据应正确", Double.valueOf(1.2), accelData.getNumericData("x"));
        assertEquals("Y轴数据应正确", Double.valueOf(-0.8), accelData.getNumericData("y"));
        assertEquals("Z轴数据应正确", Double.valueOf(9.8), accelData.getNumericData("z"));
        
        // 通知数据
        WearOSData notificationData = WearOSData.createNotificationData("not_001", "测试标题", "测试内容", "com.test.app");
        assertTrue("应为通知数据", notificationData.isNotificationData());
        assertEquals("类型应正确", WearOSData.TYPE_NOTIFICATION, notificationData.getType());
        assertEquals("标题应正确", "测试标题", notificationData.getStringData("title"));
        assertEquals("内容应正确", "测试内容", notificationData.getStringData("content"));
        assertEquals("包名应正确", "com.test.app", notificationData.getStringData("package_name"));
        
        // 应用状态数据
        Map<String, Object> stateData = new HashMap<>();
        stateData.put("active", true);
        stateData.put("battery_level", 85);
        WearOSData appStateData = WearOSData.createAppStateData("app_001", "running", stateData);
        assertTrue("应为应用状态数据", appStateData.isAppStateData());
        assertEquals("类型应正确", WearOSData.TYPE_APP_STATE, appStateData.getType());
        assertEquals("子类型应正确", "running", appStateData.getSubType());
        assertEquals("状态数据应正确", true, appStateData.getData("active"));
        assertEquals("状态数据应正确", 85, appStateData.getData("battery_level"));
    }

    @Test
    public void testValidation() {
        // 无效数据
        WearOSData invalidData = new WearOSData();
        assertFalse("缺少必要字段应无效", invalidData.isValid());
        
        // 只有ID
        invalidData.setId("test_id");
        assertFalse("只有ID应无效", invalidData.isValid());
        
        // 有ID和类型
        invalidData.setType("test_type");
        assertTrue("有ID和类型应有效", invalidData.isValid());
        
        // 空ID
        invalidData.setId("");
        assertFalse("空ID应无效", invalidData.isValid());
        
        // 空类型
        invalidData.setId("test_id");
        invalidData.setType("");
        assertFalse("空类型应无效", invalidData.isValid());
        
        // null时间戳
        invalidData.setType("test_type");
        invalidData.setTimestamp(null);
        assertFalse("null时间戳应无效", invalidData.isValid());
    }

    @Test
    public void testTypeCheckers() {
        WearOSData data = new WearOSData();
        
        // 位置数据
        data.setType(WearOSData.TYPE_LOCATION);
        assertTrue("应识别为位置数据", data.isLocationData());
        assertFalse("不应识别为健康数据", data.isHealthData());
        assertFalse("不应识别为传感器数据", data.isSensorData());
        assertFalse("不应识别为通知数据", data.isNotificationData());
        assertFalse("不应识别为应用状态数据", data.isAppStateData());
        
        // 健康数据
        data.setType(WearOSData.TYPE_HEALTH);
        assertFalse("不应识别为位置数据", data.isLocationData());
        assertTrue("应识别为健康数据", data.isHealthData());
        assertFalse("不应识别为传感器数据", data.isSensorData());
        assertFalse("不应识别为通知数据", data.isNotificationData());
        assertFalse("不应识别为应用状态数据", data.isAppStateData());
        
        // 传感器数据
        data.setType(WearOSData.TYPE_SENSOR);
        assertFalse("不应识别为位置数据", data.isLocationData());
        assertFalse("不应识别为健康数据", data.isHealthData());
        assertTrue("应识别为传感器数据", data.isSensorData());
        assertFalse("不应识别为通知数据", data.isNotificationData());
        assertFalse("不应识别为应用状态数据", data.isAppStateData());
        
        // 通知数据
        data.setType(WearOSData.TYPE_NOTIFICATION);
        assertFalse("不应识别为位置数据", data.isLocationData());
        assertFalse("不应识别为健康数据", data.isHealthData());
        assertFalse("不应识别为传感器数据", data.isSensorData());
        assertTrue("应识别为通知数据", data.isNotificationData());
        assertFalse("不应识别为应用状态数据", data.isAppStateData());
        
        // 应用状态数据
        data.setType(WearOSData.TYPE_APP_STATE);
        assertFalse("不应识别为位置数据", data.isLocationData());
        assertFalse("不应识别为健康数据", data.isHealthData());
        assertFalse("不应识别为传感器数据", data.isSensorData());
        assertFalse("不应识别为通知数据", data.isNotificationData());
        assertTrue("应识别为应用状态数据", data.isAppStateData());
    }

    @Test
    public void testEqualsAndHashCode() {
        WearOSData data1 = new WearOSData("test_id", "test_type", "test_subtype");
        WearOSData data2 = new WearOSData("test_id", "different_type", "different_subtype");
        WearOSData data3 = new WearOSData("different_id", "test_type", "test_subtype");
        WearOSData data4 = new WearOSData();
        WearOSData data5 = new WearOSData();
        
        // 相同ID的对象应相等
        assertEquals("相同ID应相等", data1, data2);
        assertEquals("相同ID的hashCode应相等", data1.hashCode(), data2.hashCode());
        
        // 不同ID的对象不应相等
        assertNotEquals("不同ID不应相等", data1, data3);
        
        // null ID的对象
        assertEquals("都为null ID应相等", data4, data5);
        assertNotEquals("null ID与非null ID不应相等", data1, data4);
        
        // 与null比较
        assertNotEquals("与null不应相等", data1, null);
        
        // 与不同类型比较
        assertNotEquals("与字符串不应相等", data1, "test_string");
    }

    @Test
    public void testToString() {
        WearOSData data = new WearOSData("test_id", "test_type", "test_subtype");
        data.addData("key1", "value1");
        data.addData("key2", 123);
        data.setSynced(true);
        data.setPriority(5);
        
        String toString = data.toString();
        
        assertNotNull("toString不应为null", toString);
        assertTrue("应包含ID", toString.contains("test_id"));
        assertTrue("应包含类型", toString.contains("test_type"));
        assertTrue("应包含子类型", toString.contains("test_subtype"));
        assertTrue("应包含数据大小", toString.contains("dataSize=2"));
        assertTrue("应包含同步状态", toString.contains("synced=true"));
        assertTrue("应包含优先级", toString.contains("priority=5"));
    }

    @Test
    public void testSettersAndGetters() {
        Date testDate = new Date();
        
        wearOSData.setId("test_id");
        wearOSData.setType("test_type");
        wearOSData.setSubType("test_subtype");
        wearOSData.setDeviceId("device_123");
        wearOSData.setSynced(true);
        wearOSData.setPriority(10);
        wearOSData.setTimestamp(testDate);
        
        assertEquals("ID应正确设置", "test_id", wearOSData.getId());
        assertEquals("类型应正确设置", "test_type", wearOSData.getType());
        assertEquals("子类型应正确设置", "test_subtype", wearOSData.getSubType());
        assertEquals("设备ID应正确设置", "device_123", wearOSData.getDeviceId());
        assertTrue("同步状态应正确设置", wearOSData.isSynced());
        assertEquals("优先级应正确设置", 10, wearOSData.getPriority());
        assertEquals("时间戳应正确设置", testDate, wearOSData.getTimestamp());
        
        // 测试null数据设置
        wearOSData.setData(null);
        assertNotNull("设置null数据后应有空映射", wearOSData.getData());
        assertEquals("设置null数据后大小应为0", 0, wearOSData.getDataSize());
    }
}