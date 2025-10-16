#pragma once

#include "pch.h"
#include <memory>
#include <vector>
#include <string>
#include <functional>

// 前向声明
class DesktopPet;
class FileTransferManager;
class DataSyncManager;
class AuthManager;
class ChatManager;
class ProjectManager;
class FriendManager;

// 应用程序主类
class Application {
public:
    Application();
    ~Application();

    // 初始化和清理
    bool Initialize(HINSTANCE hInstance, int nCmdShow);
    void Shutdown();

    // 消息处理
    bool PreTranslateMessage(MSG* pMsg);
    void OnIdle();

    // 窗口管理
    HWND GetMainWindow() const { return m_hMainWnd; }
    HWND GetTrayWindow() const { return m_hTrayWnd; }

    // 组件访问
    DesktopPet* GetDesktopPet() const { return m_pDesktopPet.get(); }
    FileTransferManager* GetFileTransfer() const { return m_pFileTransfer.get(); }
    DataSyncManager* GetDataSync() const { return m_pDataSync.get(); }
    AuthManager* GetAuthManager() const { return m_pAuthManager.get(); }
    ChatManager* GetChatManager() const { return m_pChatManager.get(); }
    ProjectManager* GetProjectManager() const { return m_pProjectManager.get(); }
    FriendManager* GetFriendManager() const { return m_pFriendManager.get(); }

    // 事件处理
    void OnTrayIconClick();
    void OnTrayIconDoubleClick();
    void OnTrayIconRightClick();
    void ShowContextMenu(POINT pt);

    // 应用程序控制
    void ShowMainWindow();
    void HideMainWindow();
    void ExitApplication();

    // 配置管理
    bool LoadConfiguration();
    bool SaveConfiguration();

private:
    // 窗口创建和初始化
    bool CreateMainWindow();
    bool CreateTrayIcon();
    bool InitializeComponents();

    // 窗口过程
    static LRESULT CALLBACK MainWndProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
    static LRESULT CALLBACK TrayWndProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);

    // 成员变量
    HINSTANCE m_hInstance;
    HWND m_hMainWnd;
    HWND m_hTrayWnd;
    NOTIFYICONDATA m_nid;
    HMENU m_hTrayMenu;

    // 组件管理器
    std::unique_ptr<DesktopPet> m_pDesktopPet;
    std::unique_ptr<FileTransferManager> m_pFileTransfer;
    std::unique_ptr<DataSyncManager> m_pDataSync;
    std::unique_ptr<AuthManager> m_pAuthManager;
    std::unique_ptr<ChatManager> m_pChatManager;
    std::unique_ptr<ProjectManager> m_pProjectManager;
    std::unique_ptr<FriendManager> m_pFriendManager;

    // 状态标志
    bool m_bInitialized;
    bool m_bMainWindowVisible;
    bool m_bShuttingDown;

    // 配置数据
    std::wstring m_configPath;
    std::wstring m_dataPath;
};

// 全局应用程序实例访问
extern Application* GetApp();

// 常量定义
const UINT WM_TRAYICON = WM_USER + 1;
const UINT WM_SHOW_MAIN_WINDOW = WM_USER + 2;
const UINT WM_HIDE_MAIN_WINDOW = WM_USER + 3;
const UINT WM_EXIT_APPLICATION = WM_USER + 4;

// 托盘菜单ID
const UINT ID_TRAY_SHOW = 1001;
const UINT ID_TRAY_HIDE = 1002;
const UINT ID_TRAY_SETTINGS = 1003;
const UINT ID_TRAY_ABOUT = 1004;
const UINT ID_TRAY_EXIT = 1005;
const UINT ID_TRAY_DESKTOP_PET = 1006;
const UINT ID_TRAY_FILE_TRANSFER = 1007;
const UINT ID_TRAY_SYNC_DATA = 1008;