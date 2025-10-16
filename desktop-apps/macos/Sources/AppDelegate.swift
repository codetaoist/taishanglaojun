import Cocoa
import Foundation
import Network
import UserNotifications

@main
class AppDelegate: NSObject, NSApplicationDelegate {
    
    // MARK: - Properties
    var mainWindowController: MainWindowController?
    var desktopPetController: DesktopPetController?
    var fileTransferManager: FileTransferManager?
    var dataSyncManager: DataSyncManager?
    var projectManager: ProjectManager?
    var networkManager: NetworkManager?
    var statusBarController: StatusBarController?
    var menuManager: MenuManager?
    var desktopPetManager: DesktopPetManager?
    
    private var isFirstLaunch = true
    private var applicationSupportURL: URL?
    
    // MARK: - Application Lifecycle
    
    func applicationDidFinishLaunching(_ aNotification: Notification) {
        setupApplicationEnvironment()
        setupManagers()
        setupUI()
        setupNotifications()
        
        // 检查是否是首次启动
        if isFirstLaunch {
            showWelcomeScreen()
        } else {
            showMainWindow()
        }
        
        // 启动桌面宠物
        startDesktopPet()
        
        // 开始数据同步
        startDataSync()
        
        // 设置通知观察者
        setupNotificationObservers()
        
        // 启动桌面宠物（可选）
        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
            self.desktopPetManager?.showPet()
        }
        
        NSLog("太上老君AI平台桌面版启动完成")
    }
    
    func applicationWillTerminate(_ aNotification: Notification) {
        // 保存应用状态
        saveApplicationState()
        
        // 停止所有服务
        stopAllServices()
        
        // 清理通知观察者
        NotificationCenter.default.removeObserver(self)
        
        NSLog("太上老君AI平台桌面版正在退出")
    }
    
    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        // 不要在最后一个窗口关闭时退出应用，保持在状态栏运行
        return false
    }
    
    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        // 当用户点击Dock图标时显示主窗口
        if !flag {
            showMainWindow()
        }
        return true
    }
    
    // MARK: - Setup Methods
    
    private func setupApplicationEnvironment() {
        // 设置应用支持目录
        let fileManager = FileManager.default
        let appSupportDir = fileManager.urls(for: .applicationSupportDirectory, 
                                           in: .userDomainMask).first!
        applicationSupportURL = appSupportDir.appendingPathComponent("TaishangLaojun")
        
        // 创建必要的目录
        try? fileManager.createDirectory(at: applicationSupportURL!, 
                                       withIntermediateDirectories: true)
        
        // 设置日志系统
        setupLogging()
        
        // 检查系统权限
        requestPermissions()
    }
    
    private func setupManagers() {
        // 菜单管理器
        menuManager = MenuManager.shared
        
        // 桌面宠物管理器
        desktopPetManager = DesktopPetManager.shared
        
        // 网络管理器
        networkManager = NetworkManager()
        
        // 文件传输管理器
        fileTransferManager = FileTransferManager(networkManager: networkManager!)
        
        // 数据同步管理器
        dataSyncManager = DataSyncManager(networkManager: networkManager!)
        
        // 项目管理器
        projectManager = ProjectManager(dataSyncManager: dataSyncManager!)
        
        // 桌面宠物控制器
        desktopPetController = DesktopPetController()
        
        // 状态栏控制器
        statusBarController = StatusBarController()
        statusBarController?.delegate = self
    }
    
    private func setupUI() {
        // 设置应用图标
        if let appIcon = NSImage(named: "AppIcon") {
            NSApplication.shared.applicationIconImage = appIcon
        }
        
        // 设置菜单栏
        setupMenuBar()
        
        // 创建主窗口控制器
        mainWindowController = MainWindowController()
        mainWindowController?.fileTransferManager = fileTransferManager
        mainWindowController?.dataSyncManager = dataSyncManager
        mainWindowController?.projectManager = projectManager
    }
    
    private func setupMenuBar() {
        let mainMenu = NSMenu()
        
        // 应用菜单
        let appMenuItem = NSMenuItem()
        let appMenu = NSMenu()
        
        appMenu.addItem(withTitle: "关于太上老君AI平台", action: #selector(showAbout), keyEquivalent: "")
        appMenu.addItem(NSMenuItem.separator())
        appMenu.addItem(withTitle: "偏好设置...", action: #selector(showPreferences), keyEquivalent: ",")
        appMenu.addItem(NSMenuItem.separator())
        appMenu.addItem(withTitle: "隐藏太上老君AI平台", action: #selector(NSApplication.hide(_:)), keyEquivalent: "h")
        appMenu.addItem(withTitle: "隐藏其他", action: #selector(NSApplication.hideOtherApplications(_:)), keyEquivalent: "h")
        appMenu.addItem(withTitle: "显示全部", action: #selector(NSApplication.unhideAllApplications(_:)), keyEquivalent: "")
        appMenu.addItem(NSMenuItem.separator())
        appMenu.addItem(withTitle: "退出太上老君AI平台", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q")
        
        appMenuItem.submenu = appMenu
        mainMenu.addItem(appMenuItem)
        
        // 文件菜单
        let fileMenuItem = NSMenuItem(title: "文件", action: nil, keyEquivalent: "")
        let fileMenu = NSMenu(title: "文件")
        
        fileMenu.addItem(withTitle: "新建项目", action: #selector(newProject), keyEquivalent: "n")
        fileMenu.addItem(withTitle: "打开项目", action: #selector(openProject), keyEquivalent: "o")
        fileMenu.addItem(NSMenuItem.separator())
        fileMenu.addItem(withTitle: "导入文件", action: #selector(importFile), keyEquivalent: "i")
        fileMenu.addItem(withTitle: "导出项目", action: #selector(exportProject), keyEquivalent: "e")
        fileMenu.addItem(NSMenuItem.separator())
        fileMenu.addItem(withTitle: "关闭窗口", action: #selector(NSWindow.performClose(_:)), keyEquivalent: "w")
        
        fileMenuItem.submenu = fileMenu
        mainMenu.addItem(fileMenuItem)
        
        // 编辑菜单
        let editMenuItem = NSMenuItem(title: "编辑", action: nil, keyEquivalent: "")
        let editMenu = NSMenu(title: "编辑")
        
        editMenu.addItem(withTitle: "撤销", action: #selector(NSResponder.undo(_:)), keyEquivalent: "z")
        editMenu.addItem(withTitle: "重做", action: #selector(NSResponder.redo(_:)), keyEquivalent: "Z")
        editMenu.addItem(NSMenuItem.separator())
        editMenu.addItem(withTitle: "剪切", action: #selector(NSText.cut(_:)), keyEquivalent: "x")
        editMenu.addItem(withTitle: "复制", action: #selector(NSText.copy(_:)), keyEquivalent: "c")
        editMenu.addItem(withTitle: "粘贴", action: #selector(NSText.paste(_:)), keyEquivalent: "v")
        editMenu.addItem(withTitle: "全选", action: #selector(NSResponder.selectAll(_:)), keyEquivalent: "a")
        
        editMenuItem.submenu = editMenu
        mainMenu.addItem(editMenuItem)
        
        // 视图菜单
        let viewMenuItem = NSMenuItem(title: "视图", action: nil, keyEquivalent: "")
        let viewMenu = NSMenu(title: "视图")
        
        viewMenu.addItem(withTitle: "显示桌面宠物", action: #selector(toggleDesktopPet), keyEquivalent: "p")
        viewMenu.addItem(withTitle: "显示文件传输", action: #selector(showFileTransfer), keyEquivalent: "t")
        viewMenu.addItem(withTitle: "显示项目管理", action: #selector(showProjectManager), keyEquivalent: "m")
        viewMenu.addItem(NSMenuItem.separator())
        viewMenu.addItem(withTitle: "进入全屏", action: #selector(NSWindow.toggleFullScreen(_:)), keyEquivalent: "f")
        
        viewMenuItem.submenu = viewMenu
        mainMenu.addItem(viewMenuItem)
        
        // 窗口菜单
        let windowMenuItem = NSMenuItem(title: "窗口", action: nil, keyEquivalent: "")
        let windowMenu = NSMenu(title: "窗口")
        
        windowMenu.addItem(withTitle: "最小化", action: #selector(NSWindow.miniaturize(_:)), keyEquivalent: "m")
        windowMenu.addItem(withTitle: "缩放", action: #selector(NSWindow.performZoom(_:)), keyEquivalent: "")
        windowMenu.addItem(NSMenuItem.separator())
        windowMenu.addItem(withTitle: "前置全部窗口", action: #selector(NSApplication.arrangeInFront(_:)), keyEquivalent: "")
        
        windowMenuItem.submenu = windowMenu
        mainMenu.addItem(windowMenuItem)
        
        // 帮助菜单
        let helpMenuItem = NSMenuItem(title: "帮助", action: nil, keyEquivalent: "")
        let helpMenu = NSMenu(title: "帮助")
        
        helpMenu.addItem(withTitle: "太上老君AI平台帮助", action: #selector(showHelp), keyEquivalent: "?")
        helpMenu.addItem(withTitle: "用户手册", action: #selector(showUserManual), keyEquivalent: "")
        helpMenu.addItem(withTitle: "快捷键", action: #selector(showShortcuts), keyEquivalent: "")
        helpMenu.addItem(NSMenuItem.separator())
        helpMenu.addItem(withTitle: "反馈问题", action: #selector(reportIssue), keyEquivalent: "")
        helpMenu.addItem(withTitle: "检查更新", action: #selector(checkForUpdates), keyEquivalent: "")
        
        helpMenuItem.submenu = helpMenu
        mainMenu.addItem(helpMenuItem)
        
        NSApplication.shared.mainMenu = mainMenu
    }
    
    private func setupNotifications() {
        // 请求通知权限
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .badge, .sound]) { granted, error in
            if granted {
                NSLog("通知权限已授予")
            } else if let error = error {
                NSLog("通知权限请求失败: \(error)")
            }
        }
        
        // 设置通知代理
        UNUserNotificationCenter.current().delegate = self
    }
    
    private func setupLogging() {
        // 设置日志文件路径
        let logURL = applicationSupportURL!.appendingPathComponent("app.log")
        
        // 这里可以集成第三方日志库，如CocoaLumberjack
        NSLog("日志系统已初始化，日志文件: \(logURL.path)")
    }
    
    private func requestPermissions() {
        // 请求必要的系统权限
        // 这里可以添加权限请求逻辑
    }
    
    // MARK: - UI Methods
    
    private func showWelcomeScreen() {
        // 显示欢迎界面
        let welcomeController = WelcomeViewController()
        welcomeController.delegate = self
        
        let window = NSWindow(contentViewController: welcomeController)
        window.title = "欢迎使用太上老君AI平台"
        window.styleMask = [.titled, .closable]
        window.center()
        window.makeKeyAndOrderFront(nil)
    }
    
    private func showMainWindow() {
        mainWindowController?.showWindow(nil)
        mainWindowController?.window?.makeKeyAndOrderFront(nil)
        NSApplication.shared.activate(ignoringOtherApps: true)
    }
    
    private func startDesktopPet() {
        desktopPetController?.startPet()
    }
    
    private func startDataSync() {
        dataSyncManager?.startSync()
    }
    
    // MARK: - Menu Actions
    
    @objc private func showAbout() {
        NSApplication.shared.orderFrontStandardAboutPanel(nil)
    }
    
    @objc private func showPreferences() {
        // 显示偏好设置窗口
        let preferencesController = PreferencesWindowController()
        preferencesController.showWindow(nil)
    }
    
    @objc private func newProject() {
        projectManager?.createNewProject()
    }
    
    @objc private func openProject() {
        projectManager?.openProject()
    }
    
    @objc private func importFile() {
        fileTransferManager?.importFile()
    }
    
    @objc private func exportProject() {
        projectManager?.exportCurrentProject()
    }
    
    @objc private func toggleDesktopPet() {
        desktopPetController?.toggleVisibility()
    }
    
    @objc private func showFileTransfer() {
        let fileTransferController = FileTransferViewController()
        fileTransferController.fileTransferManager = fileTransferManager
        
        let window = NSWindow(contentViewController: fileTransferController)
        window.title = "文件传输"
        window.styleMask = [.titled, .closable, .resizable]
        window.setContentSize(NSSize(width: 600, height: 400))
        window.center()
        window.makeKeyAndOrderFront(nil)
    }
    
    @objc private func showProjectManager() {
        let projectController = ProjectViewController()
        projectController.projectManager = projectManager
        
        let window = NSWindow(contentViewController: projectController)
        window.title = "项目管理"
        window.styleMask = [.titled, .closable, .resizable]
        window.setContentSize(NSSize(width: 800, height: 600))
        window.center()
        window.makeKeyAndOrderFront(nil)
    }
    
    @objc private func showHelp() {
        if let helpURL = URL(string: "https://help.taishanglaojun.com") {
            NSWorkspace.shared.open(helpURL)
        }
    }
    
    @objc private func showUserManual() {
        if let manualURL = URL(string: "https://docs.taishanglaojun.com") {
            NSWorkspace.shared.open(manualURL)
        }
    }
    
    @objc private func showShortcuts() {
        // 显示快捷键窗口
        let shortcutsController = ShortcutsViewController()
        
        let window = NSWindow(contentViewController: shortcutsController)
        window.title = "快捷键"
        window.styleMask = [.titled, .closable]
        window.setContentSize(NSSize(width: 400, height: 300))
        window.center()
        window.makeKeyAndOrderFront(nil)
    }
    
    @objc private func reportIssue() {
        if let issueURL = URL(string: "https://github.com/taishanglaojun/desktop/issues") {
            NSWorkspace.shared.open(issueURL)
        }
    }
    
    @objc private func checkForUpdates() {
        // 检查应用更新
        let updateChecker = UpdateChecker()
        updateChecker.checkForUpdates()
    }
    
    // MARK: - Application State
    
    private func saveApplicationState() {
        let userDefaults = UserDefaults.standard
        
        // 保存窗口状态
        if let mainWindow = mainWindowController?.window {
            userDefaults.set(NSStringFromRect(mainWindow.frame), forKey: "MainWindowFrame")
        }
        
        // 保存其他应用状态
        userDefaults.set(desktopPetController?.isVisible ?? false, forKey: "DesktopPetVisible")
        
        userDefaults.synchronize()
    }
    
    private func stopAllServices() {
        dataSyncManager?.stopSync()
        fileTransferManager?.cancelAllTransfers()
        desktopPetController?.stopPet()
        networkManager?.disconnect()
    }
    
    // MARK: - Notification Observers
    private func setupNotificationObservers() {
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(showMainWindowFromNotification),
            name: .showMainWindow,
            object: nil
        )
        
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(showSettingsFromNotification),
            name: .showSettings,
            object: nil
        )
        
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(showChatFromNotification),
            name: .showChat,
            object: nil
        )
        
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(showFileTransferFromNotification),
            name: .showFileTransfer,
            object: nil
        )
    }
    
    @objc private func showMainWindowFromNotification() {
        showMainWindow()
        NSApplication.shared.activate(ignoringOtherApps: true)
    }
    
    @objc private func showSettingsFromNotification() {
        showPreferences()
    }
    
    @objc private func showChatFromNotification() {
        showMainWindow()
        // 可以在这里添加切换到聊天界面的逻辑
    }
    
    @objc private func showFileTransferFromNotification() {
        showFileTransfer()
    }
}

// MARK: - StatusBarControllerDelegate

extension AppDelegate: StatusBarControllerDelegate {
    func statusBarControllerDidRequestShowMainWindow(_ controller: StatusBarController) {
        showMainWindow()
    }
    
    func statusBarControllerDidRequestQuit(_ controller: StatusBarController) {
        NSApplication.shared.terminate(nil)
    }
}

// MARK: - WelcomeViewControllerDelegate

extension AppDelegate: WelcomeViewControllerDelegate {
    func welcomeViewControllerDidFinish(_ controller: WelcomeViewController) {
        controller.view.window?.close()
        showMainWindow()
        isFirstLaunch = false
    }
}

// MARK: - UNUserNotificationCenterDelegate

extension AppDelegate: UNUserNotificationCenterDelegate {
    func userNotificationCenter(_ center: UNUserNotificationCenter, 
                              willPresent notification: UNNotification, 
                              withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void) {
        completionHandler([.alert, .badge, .sound])
    }
    
    func userNotificationCenter(_ center: UNUserNotificationCenter, 
                              didReceive response: UNNotificationResponse, 
                              withCompletionHandler completionHandler: @escaping () -> Void) {
        // 处理通知点击
        let userInfo = response.notification.request.content.userInfo
        
        if let actionType = userInfo["actionType"] as? String {
            switch actionType {
            case "showMainWindow":
                showMainWindow()
            case "showFileTransfer":
                showFileTransfer()
            case "showProject":
                showProjectManager()
            default:
                break
            }
        }
        
        completionHandler()
    }
}