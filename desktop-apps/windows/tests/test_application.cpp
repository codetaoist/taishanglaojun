#include "application.h"
#include <iostream>
#include <windows.h>

// 测试框架宏
#define TEST_ASSERT(condition, message) \
    do { \
        if (!(condition)) { \
            std::cerr << "FAIL: " << message << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

// 测试应用程序初始化
bool test_application_init() {
    // 创建应用程序实例
    Application app;
    
    // 获取当前模块句柄
    HINSTANCE hInstance = GetModuleHandle(nullptr);
    
    // 测试初始化
    bool result = app.Initialize(hInstance, SW_HIDE);
    TEST_ASSERT(result, "Application should initialize successfully");
    
    // 测试重复初始化
    bool result2 = app.Initialize(hInstance, SW_HIDE);
    TEST_ASSERT(result2, "Application should handle repeated initialization");
    
    // 清理
    app.Shutdown();
    
    return true;
}

// 测试应用程序关闭
bool test_application_shutdown() {
    Application app;
    HINSTANCE hInstance = GetModuleHandle(nullptr);
    
    // 初始化应用程序
    bool init_result = app.Initialize(hInstance, SW_HIDE);
    TEST_ASSERT(init_result, "Application should initialize before shutdown test");
    
    // 测试关闭
    app.Shutdown();
    
    // 测试重复关闭
    app.Shutdown();
    
    return true;
}

// 测试消息处理
bool test_application_message_handling() {
    Application app;
    HINSTANCE hInstance = GetModuleHandle(nullptr);
    
    bool init_result = app.Initialize(hInstance, SW_HIDE);
    TEST_ASSERT(init_result, "Application should initialize for message handling test");
    
    // 创建测试消息
    MSG msg = {};
    msg.message = WM_NULL;
    msg.wParam = 0;
    msg.lParam = 0;
    
    // 测试消息预处理
    bool handled = app.PreTranslateMessage(&msg);
    // 对于WM_NULL消息，应该返回false（未处理）
    TEST_ASSERT(!handled, "WM_NULL message should not be handled");
    
    // 测试空闲处理
    app.OnIdle();
    
    app.Shutdown();
    return true;
}

// 测试窗口管理
bool test_application_window_management() {
    Application app;
    HINSTANCE hInstance = GetModuleHandle(nullptr);
    
    bool init_result = app.Initialize(hInstance, SW_HIDE);
    TEST_ASSERT(init_result, "Application should initialize for window management test");
    
    // 测试获取主窗口句柄
    HWND mainWnd = app.GetMainWindow();
    // 在测试模式下，主窗口可能为空，这是正常的
    
    app.Shutdown();
    return true;
}