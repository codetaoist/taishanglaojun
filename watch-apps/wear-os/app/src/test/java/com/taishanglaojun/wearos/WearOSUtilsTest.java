package com.taishanglaojun.wearos;

import com.taishanglaojun.wearos.utils.WearOSUtils;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.junit.MockitoJUnitRunner;

import static org.junit.Assert.*;

/**
 * WearOSUtils 单元测试
 */
@RunWith(MockitoJUnitRunner.class)
public class WearOSUtilsTest {

    @Test
    public void testGenerateUniqueId() {
        String id1 = WearOSUtils.generateUniqueId();
        String id2 = WearOSUtils.generateUniqueId();
        
        assertNotNull("生成的ID不应为null", id1);
        assertNotNull("生成的ID不应为null", id2);
        assertNotEquals("两次生成的ID应该不同", id1, id2);
        assertTrue("ID长度应大于0", id1.length() > 0);
    }

    @Test
    public void testGenerateUniqueIdWithPrefix() {
        String prefix = "test";
        String id = WearOSUtils.generateUniqueId(prefix);
        
        assertNotNull("生成的ID不应为null", id);
        assertTrue("ID应以指定前缀开头", id.startsWith(prefix + "_"));
        assertTrue("ID长度应大于前缀长度", id.length() > prefix.length() + 1);
    }

    @Test
    public void testSafeParseInt() {
        assertEquals("正常整数解析", 123, WearOSUtils.safeParseInt("123", 0));
        assertEquals("负数解析", -456, WearOSUtils.safeParseInt("-456", 0));
        assertEquals("无效字符串应返回默认值", 999, WearOSUtils.safeParseInt("abc", 999));
        assertEquals("null应返回默认值", 888, WearOSUtils.safeParseInt(null, 888));
        assertEquals("空字符串应返回默认值", 777, WearOSUtils.safeParseInt("", 777));
    }

    @Test
    public void testSafeParseFloat() {
        assertEquals("正常浮点数解析", 12.34f, WearOSUtils.safeParseFloat("12.34", 0.0f), 0.001f);
        assertEquals("负浮点数解析", -56.78f, WearOSUtils.safeParseFloat("-56.78", 0.0f), 0.001f);
        assertEquals("无效字符串应返回默认值", 99.9f, WearOSUtils.safeParseFloat("abc", 99.9f), 0.001f);
        assertEquals("null应返回默认值", 88.8f, WearOSUtils.safeParseFloat(null, 88.8f), 0.001f);
    }

    @Test
    public void testSafeParseDouble() {
        assertEquals("正常双精度解析", 12.345678, WearOSUtils.safeParseDouble("12.345678", 0.0), 0.000001);
        assertEquals("负双精度解析", -98.765432, WearOSUtils.safeParseDouble("-98.765432", 0.0), 0.000001);
        assertEquals("无效字符串应返回默认值", 99.99, WearOSUtils.safeParseDouble("invalid", 99.99), 0.000001);
    }

    @Test
    public void testIsEmpty() {
        assertTrue("null字符串应为空", WearOSUtils.isEmpty(null));
        assertTrue("空字符串应为空", WearOSUtils.isEmpty(""));
        assertTrue("只有空格的字符串应为空", WearOSUtils.isEmpty("   "));
        assertTrue("制表符和换行符应为空", WearOSUtils.isEmpty("\t\n"));
        
        assertFalse("正常字符串不应为空", WearOSUtils.isEmpty("hello"));
        assertFalse("包含空格的字符串不应为空", WearOSUtils.isEmpty(" hello "));
    }

    @Test
    public void testSafeString() {
        assertEquals("正常字符串应原样返回", "hello", WearOSUtils.safeString("hello", "default"));
        assertEquals("null应返回默认值", "default", WearOSUtils.safeString(null, "default"));
        assertEquals("空字符串应返回默认值", "default", WearOSUtils.safeString("", "default"));
        assertEquals("只有空格应返回默认值", "default", WearOSUtils.safeString("   ", "default"));
    }

    @Test
    public void testFormatDateTime() {
        long timestamp = 1640995200000L; // 2022-01-01 00:00:00 UTC
        String formatted = WearOSUtils.formatDateTime(timestamp);
        
        assertNotNull("格式化结果不应为null", formatted);
        assertTrue("应包含日期信息", formatted.contains("2022"));
        assertTrue("应包含时间分隔符", formatted.contains(":"));
    }

    @Test
    public void testFormatTime() {
        long timestamp = 1640995200000L; // 2022-01-01 00:00:00 UTC
        String formatted = WearOSUtils.formatTime(timestamp);
        
        assertNotNull("格式化结果不应为null", formatted);
        assertTrue("应包含时间分隔符", formatted.contains(":"));
        // 时间格式应该是 HH:mm:ss
        assertTrue("时间格式应正确", formatted.matches("\\d{2}:\\d{2}:\\d{2}"));
    }

    @Test
    public void testGetCurrentTimestamp() {
        long timestamp1 = WearOSUtils.getCurrentTimestamp();
        
        try {
            Thread.sleep(10); // 等待10毫秒
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        
        long timestamp2 = WearOSUtils.getCurrentTimestamp();
        
        assertTrue("第二个时间戳应大于第一个", timestamp2 > timestamp1);
        assertTrue("时间戳应为正数", timestamp1 > 0);
        assertTrue("时间戳应为正数", timestamp2 > 0);
    }

    @Test
    public void testGetCurrentDateTime() {
        String dateTime = WearOSUtils.getCurrentDateTime();
        
        assertNotNull("当前日期时间不应为null", dateTime);
        assertFalse("当前日期时间不应为空", dateTime.isEmpty());
        assertTrue("应包含日期信息", dateTime.length() > 10);
        assertTrue("应包含时间分隔符", dateTime.contains(":"));
    }

    @Test
    public void testGetCurrentTime() {
        String time = WearOSUtils.getCurrentTime();
        
        assertNotNull("当前时间不应为null", time);
        assertFalse("当前时间不应为空", time.isEmpty());
        assertTrue("时间格式应正确", time.matches("\\d{2}:\\d{2}:\\d{2}"));
    }

    @Test
    public void testGetDeviceModel() {
        String model = WearOSUtils.getDeviceModel();
        
        assertNotNull("设备型号不应为null", model);
        assertFalse("设备型号不应为空", model.isEmpty());
    }

    @Test
    public void testGetAndroidVersion() {
        String version = WearOSUtils.getAndroidVersion();
        
        assertNotNull("Android版本不应为null", version);
        assertFalse("Android版本不应为空", version.isEmpty());
    }

    @Test
    public void testGetApiLevel() {
        int apiLevel = WearOSUtils.getApiLevel();
        
        assertTrue("API级别应大于0", apiLevel > 0);
        assertTrue("API级别应在合理范围内", apiLevel >= 21 && apiLevel <= 50);
    }
}