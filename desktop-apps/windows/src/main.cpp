#include "pch.h"
#include "application.h"

// 全局变量
HINSTANCE g_hInstance = nullptr;
Application* g_pApp = nullptr;

// 前向声明
LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
bool InitializeApplication(HINSTANCE hInstance);
void CleanupApplication();
bool CheckSingleInstance();
void ShowErrorMessage(const wchar_t* message);

// 程序入口点
int WINAPI wWinMain(
    _In_ HINSTANCE hInstance,
    _In_opt_ HINSTANCE hPrevInstance,
    _In_ LPWSTR lpCmdLine,
    _In_ int nCmdShow)
{
    UNREFERENCED_PARAMETER(hPrevInstance);
    UNREFERENCED_PARAMETER(lpCmdLine);

    // 设置DPI感知
    SetProcessDPIAware();

    // 初始化COM
    HRESULT hr = CoInitializeEx(nullptr, COINIT_APARTMENTTHREADED | COINIT_DISABLE_OLE1DDE);
    if (FAILED(hr)) {
        ShowErrorMessage(L"Failed to initialize COM library");
        return -1;
    }

    // 检查单实例运行
    if (!CheckSingleInstance()) {
        ShowErrorMessage(L"Application is already running");
        CoUninitialize();
        return -1;
    }

    // 初始化应用程序
    if (!InitializeApplication(hInstance)) {
        ShowErrorMessage(L"Failed to initialize application");
        CleanupApplication();
        CoUninitialize();
        return -1;
    }

    // 创建应用程序实例
    g_pApp = new Application();
    if (!g_pApp->Initialize(hInstance, nCmdShow)) {
        ShowErrorMessage(L"Failed to initialize application instance");
        delete g_pApp;
        g_pApp = nullptr;
        CleanupApplication();
        CoUninitialize();
        return -1;
    }

    // 消息循环
    MSG msg = {};
    int exitCode = 0;

    while (true) {
        BOOL bRet = GetMessage(&msg, nullptr, 0, 0);
        
        if (bRet == 0) {
            // WM_QUIT消息
            exitCode = static_cast<int>(msg.wParam);
            break;
        }
        else if (bRet == -1) {
            // 错误
            LOG_ERROR("GetMessage failed: %d", GetLastError());
            exitCode = -1;
            break;
        }
        else {
            // 处理消息
            if (!g_pApp->PreTranslateMessage(&msg)) {
                TranslateMessage(&msg);
                DispatchMessage(&msg);
            }
        }

        // 处理应用程序空闲时间
        g_pApp->OnIdle();
    }

    // 清理
    if (g_pApp) {
        g_pApp->Shutdown();
        delete g_pApp;
        g_pApp = nullptr;
    }

    CleanupApplication();
    CoUninitialize();

    return exitCode;
}

// 初始化应用程序
bool InitializeApplication(HINSTANCE hInstance)
{
    g_hInstance = hInstance;

    // 初始化通用控件
    INITCOMMONCONTROLSEX icex = {};
    icex.dwSize = sizeof(INITCOMMONCONTROLSEX);
    icex.dwICC = ICC_WIN95_CLASSES | ICC_COOL_CLASSES | ICC_USEREX_CLASSES;
    
    if (!InitCommonControlsEx(&icex)) {
        LOG_ERROR("Failed to initialize common controls");
        return false;
    }

    // 初始化Winsock
    WSADATA wsaData;
    int result = WSAStartup(MAKEWORD(2, 2), &wsaData);
    if (result != 0) {
        LOG_ERROR("WSAStartup failed: %d", result);
        return false;
    }

    // 注册窗口类
    WNDCLASSEXW wcex = {};
    wcex.cbSize = sizeof(WNDCLASSEXW);
    wcex.style = CS_HREDRAW | CS_VREDRAW | CS_DBLCLKS;
    wcex.lpfnWndProc = WindowProc;
    wcex.cbClsExtra = 0;
    wcex.cbWndExtra = 0;
    wcex.hInstance = hInstance;
    wcex.hIcon = LoadIcon(hInstance, MAKEINTRESOURCE(IDI_APPLICATION));
    wcex.hCursor = LoadCursor(nullptr, IDC_ARROW);
    wcex.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1);
    wcex.lpszMenuName = nullptr;
    wcex.lpszClassName = APP_CLASS_NAME;
    wcex.hIconSm = LoadIcon(hInstance, MAKEINTRESOURCE(IDI_APPLICATION));

    if (!RegisterClassExW(&wcex)) {
        LOG_ERROR("Failed to register window class: %d", GetLastError());
        return false;
    }

    return true;
}

// 清理应用程序
void CleanupApplication()
{
    // 清理Winsock
    WSACleanup();

    // 注销窗口类
    if (g_hInstance) {
        UnregisterClassW(APP_CLASS_NAME, g_hInstance);
    }
}

// 检查单实例运行
bool CheckSingleInstance()
{
    // 创建互斥体
    HANDLE hMutex = CreateMutexW(nullptr, TRUE, L"TaishangLaojunDesktopApp_Mutex");
    if (hMutex == nullptr) {
        return false;
    }

    // 检查是否已经存在
    if (GetLastError() == ERROR_ALREADY_EXISTS) {
        CloseHandle(hMutex);
        
        // 尝试找到已运行的实例并激活它
        HWND hWnd = FindWindowW(APP_CLASS_NAME, nullptr);
        if (hWnd) {
            // 如果窗口最小化，恢复它
            if (IsIconic(hWnd)) {
                ShowWindow(hWnd, SW_RESTORE);
            }
            
            // 将窗口置于前台
            SetForegroundWindow(hWnd);
        }
        
        return false;
    }

    return true;
}

// 显示错误消息
void ShowErrorMessage(const wchar_t* message)
{
    MessageBoxW(nullptr, message, APP_NAME, MB_OK | MB_ICONERROR);
}

// 窗口过程
LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam)
{
    // 如果应用程序实例存在，让它处理消息
    if (g_pApp) {
        LRESULT result = 0;
        if (g_pApp->HandleMessage(hwnd, uMsg, wParam, lParam, &result)) {
            return result;
        }
    }

    // 默认处理
    switch (uMsg) {
    case WM_CREATE:
        LOG_INFO("Window created: HWND=0x%p", hwnd);
        break;

    case WM_DESTROY:
        LOG_INFO("Window destroyed: HWND=0x%p", hwnd);
        PostQuitMessage(0);
        break;

    case WM_CLOSE:
        // 询问用户是否真的要退出
        if (MessageBoxW(hwnd, L"确定要退出太上老君AI平台吗？", APP_NAME, 
                       MB_YESNO | MB_ICONQUESTION) == IDYES) {
            DestroyWindow(hwnd);
        }
        return 0;

    case WM_SIZE:
        if (g_pApp) {
            g_pApp->OnWindowResize(hwnd, LOWORD(lParam), HIWORD(lParam));
        }
        break;

    case WM_PAINT:
        {
            PAINTSTRUCT ps;
            HDC hdc = BeginPaint(hwnd, &ps);
            
            if (g_pApp) {
                g_pApp->OnPaint(hwnd, hdc, &ps.rcPaint);
            }
            
            EndPaint(hwnd, &ps);
        }
        return 0;

    case WM_COMMAND:
        if (g_pApp) {
            g_pApp->OnCommand(hwnd, LOWORD(wParam), HIWORD(wParam), (HWND)lParam);
        }
        break;

    case WM_NOTIFY:
        if (g_pApp) {
            return g_pApp->OnNotify(hwnd, (int)wParam, (LPNMHDR)lParam);
        }
        break;

    case WM_TRAY_ICON:
        if (g_pApp) {
            g_pApp->OnTrayIcon(hwnd, wParam, lParam);
        }
        break;

    case WM_PET_UPDATE:
        if (g_pApp) {
            g_pApp->OnPetUpdate(wParam, lParam);
        }
        break;

    case WM_FILE_TRANSFER:
        if (g_pApp) {
            g_pApp->OnFileTransfer(wParam, lParam);
        }
        break;

    case WM_DATA_SYNC:
        if (g_pApp) {
            g_pApp->OnDataSync(wParam, lParam);
        }
        break;

    case WM_NOTIFICATION:
        if (g_pApp) {
            g_pApp->OnNotification(wParam, lParam);
        }
        break;

    case WM_QUERYENDSESSION:
        // 系统关机或注销
        if (g_pApp) {
            return g_pApp->OnQueryEndSession();
        }
        return TRUE;

    case WM_ENDSESSION:
        // 系统正在关机或注销
        if (wParam && g_pApp) {
            g_pApp->OnEndSession();
        }
        break;

    case WM_POWERBROADCAST:
        // 电源管理消息
        if (g_pApp) {
            g_pApp->OnPowerBroadcast(wParam, lParam);
        }
        break;

    case WM_DEVICECHANGE:
        // 设备变化消息
        if (g_pApp) {
            g_pApp->OnDeviceChange(wParam, lParam);
        }
        break;

    default:
        break;
    }

    return DefWindowProcW(hwnd, uMsg, wParam, lParam);
}