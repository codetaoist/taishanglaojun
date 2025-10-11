#include "pch.h"
#include "application.h"
#include "desktop_pet.h"
#include "file_transfer.h"
#include "data_sync.h"
#include "auth_manager.h"
#include "chat_manager.h"
#include "project_management.h"
#include "friend_manager.h"

// 全局应用程序实例
static Application* g_pApplication = nullptr;

// 窗口类名
static const wchar_t* MAIN_WINDOW_CLASS = L"TaishanglaojunMainWindow";
static const wchar_t* TRAY_WINDOW_CLASS = L"TaishanglaojunTrayWindow";

Application::Application()
    : m_hInstance(nullptr)
    , m_hMainWnd(nullptr)
    , m_hTrayWnd(nullptr)
    , m_hTrayMenu(nullptr)
    , m_bInitialized(false)
    , m_bMainWindowVisible(false)
    , m_bShuttingDown(false)
{
    g_pApplication = this;
    ZeroMemory(&m_nid, sizeof(m_nid));
}

Application::~Application() {
    Shutdown();
    g_pApplication = nullptr;
}

bool Application::Initialize(HINSTANCE hInstance, int nCmdShow) {
    if (m_bInitialized) {
        return true;
    }

    m_hInstance = hInstance;

    // 获取应用程序数据路径
    wchar_t appDataPath[MAX_PATH];
    if (SUCCEEDED(SHGetFolderPath(nullptr, CSIDL_APPDATA, nullptr, 0, appDataPath))) {
        m_dataPath = appDataPath;
        m_dataPath += L"\\Taishanglaojun";
        CreateDirectory(m_dataPath.c_str(), nullptr);
        
        m_configPath = m_dataPath + L"\\config.json";
    }

    // 注册窗口类
    if (!RegisterWindowClasses()) {
        LOG_ERROR("Failed to register window classes");
        return false;
    }

    // 创建主窗口
    if (!CreateMainWindow()) {
        LOG_ERROR("Failed to create main window");
        return false;
    }

    // 创建托盘图标
    if (!CreateTrayIcon()) {
        LOG_ERROR("Failed to create tray icon");
        return false;
    }

    // 初始化组件
    if (!InitializeComponents()) {
        LOG_ERROR("Failed to initialize components");
        return false;
    }

    // 加载配置
    LoadConfiguration();

    // 显示主窗口（可选）
    if (nCmdShow != SW_HIDE) {
        ShowMainWindow();
    }

    m_bInitialized = true;
    LOG_INFO("Application initialized successfully");
    return true;
}

void Application::Shutdown() {
    if (m_bShuttingDown) {
        return;
    }

    m_bShuttingDown = true;
    LOG_INFO("Application shutting down...");

    // 保存配置
    SaveConfiguration();

    // 关闭组件
    if (m_pFriendManager) {
        m_pFriendManager.reset();
    }
    if (m_pProjectManager) {
        m_pProjectManager.reset();
    }
    if (m_pChatManager) {
        m_pChatManager.reset();
    }
    if (m_pAuthManager) {
        m_pAuthManager.reset();
    }
    if (m_pDataSync) {
        m_pDataSync.reset();
    }
    if (m_pFileTransfer) {
        m_pFileTransfer.reset();
    }
    if (m_pDesktopPet) {
        m_pDesktopPet.reset();
    }

    // 销毁托盘图标
    if (m_nid.cbSize > 0) {
        Shell_NotifyIcon(NIM_DELETE, &m_nid);
    }

    // 销毁菜单
    if (m_hTrayMenu) {
        DestroyMenu(m_hTrayMenu);
        m_hTrayMenu = nullptr;
    }

    // 销毁窗口
    if (m_hMainWnd) {
        DestroyWindow(m_hMainWnd);
        m_hMainWnd = nullptr;
    }
    if (m_hTrayWnd) {
        DestroyWindow(m_hTrayWnd);
        m_hTrayWnd = nullptr;
    }

    m_bInitialized = false;
    LOG_INFO("Application shutdown complete");
}

bool Application::PreTranslateMessage(MSG* pMsg) {
    // 处理快捷键等特殊消息
    if (pMsg->message == WM_KEYDOWN) {
        // 全局快捷键处理
        if (pMsg->wParam == VK_ESCAPE && GetAsyncKeyState(VK_CONTROL) < 0) {
            // Ctrl+Esc 显示/隐藏主窗口
            if (m_bMainWindowVisible) {
                HideMainWindow();
            } else {
                ShowMainWindow();
            }
            return true;
        }
    }
    return false;
}

void Application::OnIdle() {
    // 更新桌面宠物
    if (m_pDesktopPet) {
        // 桌面宠物有自己的更新线程，这里不需要手动更新
    }

    // 处理其他空闲时间任务
    static DWORD lastIdleTime = 0;
    DWORD currentTime = GetTickCount();
    if (currentTime - lastIdleTime > 1000) { // 每秒执行一次
        // 检查网络状态、更新统计信息等
        lastIdleTime = currentTime;
    }
}

void Application::OnTrayIconClick() {
    // 单击托盘图标 - 显示/隐藏主窗口
    if (m_bMainWindowVisible) {
        HideMainWindow();
    } else {
        ShowMainWindow();
    }
}

void Application::OnTrayIconDoubleClick() {
    // 双击托盘图标 - 显示主窗口
    ShowMainWindow();
}

void Application::OnTrayIconRightClick() {
    // 右键托盘图标 - 显示上下文菜单
    POINT pt;
    GetCursorPos(&pt);
    ShowContextMenu(pt);
}

void Application::ShowContextMenu(POINT pt) {
    if (!m_hTrayMenu) {
        return;
    }

    // 设置前台窗口以确保菜单正确显示
    SetForegroundWindow(m_hTrayWnd);

    // 显示上下文菜单
    TrackPopupMenu(
        GetSubMenu(m_hTrayMenu, 0),
        TPM_RIGHTBUTTON | TPM_BOTTOMALIGN | TPM_RIGHTALIGN,
        pt.x, pt.y,
        0,
        m_hTrayWnd,
        nullptr
    );

    // 发送空消息以关闭菜单
    PostMessage(m_hTrayWnd, WM_NULL, 0, 0);
}

void Application::ShowMainWindow() {
    if (m_hMainWnd && !m_bMainWindowVisible) {
        ShowWindow(m_hMainWnd, SW_SHOW);
        SetForegroundWindow(m_hMainWnd);
        m_bMainWindowVisible = true;
    }
}

void Application::HideMainWindow() {
    if (m_hMainWnd && m_bMainWindowVisible) {
        ShowWindow(m_hMainWnd, SW_HIDE);
        m_bMainWindowVisible = false;
    }
}

void Application::ExitApplication() {
    PostQuitMessage(0);
}

bool Application::RegisterWindowClasses() {
    WNDCLASSEX wc = {};

    // 注册主窗口类
    wc.cbSize = sizeof(WNDCLASSEX);
    wc.style = CS_HREDRAW | CS_VREDRAW;
    wc.lpfnWndProc = MainWndProc;
    wc.hInstance = m_hInstance;
    wc.hIcon = LoadIcon(m_hInstance, MAKEINTRESOURCE(101));
    wc.hCursor = LoadCursor(nullptr, IDC_ARROW);
    wc.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1);
    wc.lpszClassName = MAIN_WINDOW_CLASS;
    wc.hIconSm = LoadIcon(m_hInstance, MAKEINTRESOURCE(101));

    if (!RegisterClassEx(&wc)) {
        return false;
    }

    // 注册托盘窗口类
    wc.lpfnWndProc = TrayWndProc;
    wc.lpszClassName = TRAY_WINDOW_CLASS;
    wc.hbrBackground = nullptr;

    return RegisterClassEx(&wc) != 0;
}

bool Application::CreateMainWindow() {
    m_hMainWnd = CreateWindowEx(
        WS_EX_APPWINDOW,
        MAIN_WINDOW_CLASS,
        L"太上老君AI平台",
        WS_OVERLAPPEDWINDOW,
        CW_USEDEFAULT, CW_USEDEFAULT,
        1200, 800,
        nullptr,
        nullptr,
        m_hInstance,
        this
    );

    return m_hMainWnd != nullptr;
}

bool Application::CreateTrayIcon() {
    // 创建隐藏的托盘窗口
    m_hTrayWnd = CreateWindowEx(
        0,
        TRAY_WINDOW_CLASS,
        L"TrayWindow",
        0,
        0, 0, 0, 0,
        nullptr,
        nullptr,
        m_hInstance,
        this
    );

    if (!m_hTrayWnd) {
        return false;
    }

    // 创建托盘菜单
    m_hTrayMenu = CreatePopupMenu();
    if (m_hTrayMenu) {
        HMENU hSubMenu = CreatePopupMenu();
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_SHOW, L"显示主窗口");
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_HIDE, L"隐藏主窗口");
        AppendMenu(hSubMenu, MF_SEPARATOR, 0, nullptr);
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_DESKTOP_PET, L"桌面宠物");
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_FILE_TRANSFER, L"文件传输");
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_SYNC_DATA, L"数据同步");
        AppendMenu(hSubMenu, MF_SEPARATOR, 0, nullptr);
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_SETTINGS, L"设置");
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_ABOUT, L"关于");
        AppendMenu(hSubMenu, MF_SEPARATOR, 0, nullptr);
        AppendMenu(hSubMenu, MF_STRING, ID_TRAY_EXIT, L"退出");
        AppendMenu(m_hTrayMenu, MF_POPUP, (UINT_PTR)hSubMenu, L"太上老君AI平台");
    }

    // 设置托盘图标
    m_nid.cbSize = sizeof(NOTIFYICONDATA);
    m_nid.hWnd = m_hTrayWnd;
    m_nid.uID = 1;
    m_nid.uFlags = NIF_ICON | NIF_MESSAGE | NIF_TIP;
    m_nid.uCallbackMessage = WM_TRAYICON;
    m_nid.hIcon = LoadIcon(m_hInstance, MAKEINTRESOURCE(101));
    wcscpy_s(m_nid.szTip, L"太上老君AI平台");

    return Shell_NotifyIcon(NIM_ADD, &m_nid) != FALSE;
}

bool Application::InitializeComponents() {
    try {
        // 初始化桌面宠物管理器
        m_pDesktopPet = std::make_unique<DesktopPetManager>();
        if (!m_pDesktopPet->Initialize()) {
            LOG_WARN("Failed to initialize desktop pet manager");
        }

        // 初始化文件传输管理器
        m_pFileTransfer = std::make_unique<FileTransferManager>();
        if (!m_pFileTransfer->Initialize("Windows Desktop", DeviceType::DESKTOP_WINDOWS)) {
            LOG_WARN("Failed to initialize file transfer manager");
        }

        // 初始化其他组件...
        // 这些组件的具体实现将在后续创建

        return true;
    }
    catch (const std::exception& e) {
        LOG_ERROR("Exception during component initialization: %s", e.what());
        return false;
    }
}

bool Application::LoadConfiguration() {
    // TODO: 实现配置加载
    return true;
}

bool Application::SaveConfiguration() {
    // TODO: 实现配置保存
    return true;
}

LRESULT CALLBACK Application::MainWndProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    Application* pApp = nullptr;

    if (uMsg == WM_NCCREATE) {
        CREATESTRUCT* pcs = reinterpret_cast<CREATESTRUCT*>(lParam);
        pApp = static_cast<Application*>(pcs->lpCreateParams);
        SetWindowLongPtr(hwnd, GWLP_USERDATA, reinterpret_cast<LONG_PTR>(pApp));
    } else {
        pApp = reinterpret_cast<Application*>(GetWindowLongPtr(hwnd, GWLP_USERDATA));
    }

    if (pApp) {
        switch (uMsg) {
        case WM_CLOSE:
            pApp->HideMainWindow();
            return 0;

        case WM_DESTROY:
            if (hwnd == pApp->m_hMainWnd) {
                pApp->m_hMainWnd = nullptr;
            }
            break;

        case WM_SIZE:
            if (wParam == SIZE_MINIMIZED) {
                pApp->HideMainWindow();
            }
            break;
        }
    }

    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

LRESULT CALLBACK Application::TrayWndProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    Application* pApp = reinterpret_cast<Application*>(GetWindowLongPtr(hwnd, GWLP_USERDATA));

    if (uMsg == WM_CREATE) {
        CREATESTRUCT* pcs = reinterpret_cast<CREATESTRUCT*>(lParam);
        pApp = static_cast<Application*>(pcs->lpCreateParams);
        SetWindowLongPtr(hwnd, GWLP_USERDATA, reinterpret_cast<LONG_PTR>(pApp));
    }

    if (pApp) {
        switch (uMsg) {
        case WM_TRAYICON:
            switch (lParam) {
            case WM_LBUTTONUP:
                pApp->OnTrayIconClick();
                break;
            case WM_LBUTTONDBLCLK:
                pApp->OnTrayIconDoubleClick();
                break;
            case WM_RBUTTONUP:
                pApp->OnTrayIconRightClick();
                break;
            }
            break;

        case WM_COMMAND:
            switch (LOWORD(wParam)) {
            case ID_TRAY_SHOW:
                pApp->ShowMainWindow();
                break;
            case ID_TRAY_HIDE:
                pApp->HideMainWindow();
                break;
            case ID_TRAY_EXIT:
                pApp->ExitApplication();
                break;
            // TODO: 处理其他菜单项
            }
            break;

        case WM_DESTROY:
            if (hwnd == pApp->m_hTrayWnd) {
                pApp->m_hTrayWnd = nullptr;
            }
            break;
        }
    }

    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

// 全局函数实现
Application* GetApp() {
    return g_pApplication;
}