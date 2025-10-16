import Cocoa
import Foundation

// MARK: - Menu Manager
class MenuManager: NSObject {
    
    // MARK: - Properties
    static let shared = MenuManager()
    
    private var statusItem: NSStatusItem?
    private var statusBarMenu: NSMenu?
    private var mainMenu: NSMenu?
    
    // MARK: - Initialization
    override init() {
        super.init()
        setupMainMenu()
        setupStatusBar()
    }
    
    // MARK: - Main Menu Setup
    private func setupMainMenu() {
        mainMenu = NSMenu()
        
        // 应用菜单
        let appMenuItem = NSMenuItem()
        let appMenu = NSMenu()
        
        // 关于
        let aboutItem = NSMenuItem(title: "关于太上老君AI", action: #selector(showAbout(_:)), keyEquivalent: "")
        aboutItem.target = self
        appMenu.addItem(aboutItem)
        
        appMenu.addItem(NSMenuItem.separator())
        
        // 偏好设置
        let preferencesItem = NSMenuItem(title: "偏好设置...", action: #selector(showPreferences(_:)), keyEquivalent: ",")
        preferencesItem.target = self
        appMenu.addItem(preferencesItem)
        
        appMenu.addItem(NSMenuItem.separator())
        
        // 服务
        let servicesItem = NSMenuItem(title: "服务", action: nil, keyEquivalent: "")
        let servicesMenu = NSMenu()
        servicesItem.submenu = servicesMenu
        NSApp.servicesMenu = servicesMenu
        appMenu.addItem(servicesItem)
        
        appMenu.addItem(NSMenuItem.separator())
        
        // 隐藏应用
        let hideItem = NSMenuItem(title: "隐藏太上老君AI", action: #selector(NSApplication.hide(_:)), keyEquivalent: "h")
        appMenu.addItem(hideItem)
        
        // 隐藏其他应用
        let hideOthersItem = NSMenuItem(title: "隐藏其他", action: #selector(NSApplication.hideOtherApplications(_:)), keyEquivalent: "h")
        hideOthersItem.keyEquivalentModifierMask = [.command, .option]
        appMenu.addItem(hideOthersItem)
        
        // 显示全部
        let showAllItem = NSMenuItem(title: "显示全部", action: #selector(NSApplication.unhideAllApplications(_:)), keyEquivalent: "")
        appMenu.addItem(showAllItem)
        
        appMenu.addItem(NSMenuItem.separator())
        
        // 退出
        let quitItem = NSMenuItem(title: "退出太上老君AI", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q")
        appMenu.addItem(quitItem)
        
        appMenuItem.submenu = appMenu
        mainMenu?.addItem(appMenuItem)
        
        // 文件菜单
        setupFileMenu()
        
        // 编辑菜单
        setupEditMenu()
        
        // 视图菜单
        setupViewMenu()
        
        // 工具菜单
        setupToolsMenu()
        
        // 窗口菜单
        setupWindowMenu()
        
        // 帮助菜单
        setupHelpMenu()
        
        NSApp.mainMenu = mainMenu
    }
    
    private func setupFileMenu() {
        let fileMenuItem = NSMenuItem(title: "文件", action: nil, keyEquivalent: "")
        let fileMenu = NSMenu(title: "文件")
        
        // 新建项目
        let newProjectItem = NSMenuItem(title: "新建项目...", action: #selector(newProject(_:)), keyEquivalent: "n")
        newProjectItem.target = self
        fileMenu.addItem(newProjectItem)
        
        // 打开项目
        let openProjectItem = NSMenuItem(title: "打开项目...", action: #selector(openProject(_:)), keyEquivalent: "o")
        openProjectItem.target = self
        fileMenu.addItem(openProjectItem)
        
        // 最近项目
        let recentItem = NSMenuItem(title: "最近项目", action: nil, keyEquivalent: "")
        let recentMenu = NSMenu(title: "最近项目")
        recentItem.submenu = recentMenu
        fileMenu.addItem(recentItem)
        
        fileMenu.addItem(NSMenuItem.separator())
        
        // 保存
        let saveItem = NSMenuItem(title: "保存", action: #selector(save(_:)), keyEquivalent: "s")
        saveItem.target = self
        fileMenu.addItem(saveItem)
        
        // 另存为
        let saveAsItem = NSMenuItem(title: "另存为...", action: #selector(saveAs(_:)), keyEquivalent: "S")
        saveAsItem.keyEquivalentModifierMask = [.command, .shift]
        saveAsItem.target = self
        fileMenu.addItem(saveAsItem)
        
        fileMenu.addItem(NSMenuItem.separator())
        
        // 导入
        let importItem = NSMenuItem(title: "导入...", action: #selector(importFile(_:)), keyEquivalent: "")
        importItem.target = self
        fileMenu.addItem(importItem)
        
        // 导出
        let exportItem = NSMenuItem(title: "导出...", action: #selector(exportFile(_:)), keyEquivalent: "")
        exportItem.target = self
        fileMenu.addItem(exportItem)
        
        fileMenu.addItem(NSMenuItem.separator())
        
        // 关闭窗口
        let closeItem = NSMenuItem(title: "关闭窗口", action: #selector(NSWindow.performClose(_:)), keyEquivalent: "w")
        fileMenu.addItem(closeItem)
        
        fileMenuItem.submenu = fileMenu
        mainMenu?.addItem(fileMenuItem)
    }
    
    private func setupEditMenu() {
        let editMenuItem = NSMenuItem(title: "编辑", action: nil, keyEquivalent: "")
        let editMenu = NSMenu(title: "编辑")
        
        // 撤销
        let undoItem = NSMenuItem(title: "撤销", action: #selector(UndoManager.undo), keyEquivalent: "z")
        editMenu.addItem(undoItem)
        
        // 重做
        let redoItem = NSMenuItem(title: "重做", action: #selector(UndoManager.redo), keyEquivalent: "Z")
        redoItem.keyEquivalentModifierMask = [.command, .shift]
        editMenu.addItem(redoItem)
        
        editMenu.addItem(NSMenuItem.separator())
        
        // 剪切
        let cutItem = NSMenuItem(title: "剪切", action: #selector(NSText.cut(_:)), keyEquivalent: "x")
        editMenu.addItem(cutItem)
        
        // 复制
        let copyItem = NSMenuItem(title: "复制", action: #selector(NSText.copy(_:)), keyEquivalent: "c")
        editMenu.addItem(copyItem)
        
        // 粘贴
        let pasteItem = NSMenuItem(title: "粘贴", action: #selector(NSText.paste(_:)), keyEquivalent: "v")
        editMenu.addItem(pasteItem)
        
        // 全选
        let selectAllItem = NSMenuItem(title: "全选", action: #selector(NSText.selectAll(_:)), keyEquivalent: "a")
        editMenu.addItem(selectAllItem)
        
        editMenu.addItem(NSMenuItem.separator())
        
        // 查找
        let findItem = NSMenuItem(title: "查找", action: nil, keyEquivalent: "")
        let findMenu = NSMenu(title: "查找")
        
        let findInProjectItem = NSMenuItem(title: "在项目中查找...", action: #selector(findInProject(_:)), keyEquivalent: "f")
        findInProjectItem.keyEquivalentModifierMask = [.command, .shift]
        findInProjectItem.target = self
        findMenu.addItem(findInProjectItem)
        
        let replaceItem = NSMenuItem(title: "查找和替换...", action: #selector(findAndReplace(_:)), keyEquivalent: "f")
        replaceItem.keyEquivalentModifierMask = [.command, .option]
        replaceItem.target = self
        findMenu.addItem(replaceItem)
        
        findItem.submenu = findMenu
        editMenu.addItem(findItem)
        
        editMenuItem.submenu = editMenu
        mainMenu?.addItem(editMenuItem)
    }
    
    private func setupViewMenu() {
        let viewMenuItem = NSMenuItem(title: "视图", action: nil, keyEquivalent: "")
        let viewMenu = NSMenu(title: "视图")
        
        // 显示/隐藏侧边栏
        let sidebarItem = NSMenuItem(title: "显示侧边栏", action: #selector(toggleSidebar(_:)), keyEquivalent: "s")
        sidebarItem.keyEquivalentModifierMask = [.command, .control]
        sidebarItem.target = self
        viewMenu.addItem(sidebarItem)
        
        // 显示/隐藏工具栏
        let toolbarItem = NSMenuItem(title: "显示工具栏", action: #selector(toggleToolbar(_:)), keyEquivalent: "t")
        toolbarItem.keyEquivalentModifierMask = [.command, .option]
        toolbarItem.target = self
        viewMenu.addItem(toolbarItem)
        
        // 显示/隐藏状态栏
        let statusBarItem = NSMenuItem(title: "显示状态栏", action: #selector(toggleStatusBar(_:)), keyEquivalent: "")
        statusBarItem.target = self
        viewMenu.addItem(statusBarItem)
        
        viewMenu.addItem(NSMenuItem.separator())
        
        // 全屏
        let fullScreenItem = NSMenuItem(title: "进入全屏", action: #selector(NSWindow.toggleFullScreen(_:)), keyEquivalent: "f")
        fullScreenItem.keyEquivalentModifierMask = [.command, .control]
        viewMenu.addItem(fullScreenItem)
        
        viewMenu.addItem(NSMenuItem.separator())
        
        // 缩放
        let zoomInItem = NSMenuItem(title: "放大", action: #selector(zoomIn(_:)), keyEquivalent: "+")
        zoomInItem.target = self
        viewMenu.addItem(zoomInItem)
        
        let zoomOutItem = NSMenuItem(title: "缩小", action: #selector(zoomOut(_:)), keyEquivalent: "-")
        zoomOutItem.target = self
        viewMenu.addItem(zoomOutItem)
        
        let actualSizeItem = NSMenuItem(title: "实际大小", action: #selector(actualSize(_:)), keyEquivalent: "0")
        actualSizeItem.target = self
        viewMenu.addItem(actualSizeItem)
        
        viewMenuItem.submenu = viewMenu
        mainMenu?.addItem(viewMenuItem)
    }
    
    private func setupToolsMenu() {
        let toolsMenuItem = NSMenuItem(title: "工具", action: nil, keyEquivalent: "")
        let toolsMenu = NSMenu(title: "工具")
        
        // 桌面宠物
        let petItem = NSMenuItem(title: "桌面宠物", action: nil, keyEquivalent: "")
        let petMenu = NSMenu(title: "桌面宠物")
        
        let showPetItem = NSMenuItem(title: "显示宠物", action: #selector(showDesktopPet(_:)), keyEquivalent: "")
        showPetItem.target = self
        petMenu.addItem(showPetItem)
        
        let hidePetItem = NSMenuItem(title: "隐藏宠物", action: #selector(hideDesktopPet(_:)), keyEquivalent: "")
        hidePetItem.target = self
        petMenu.addItem(hidePetItem)
        
        let petSettingsItem = NSMenuItem(title: "宠物设置...", action: #selector(showPetSettings(_:)), keyEquivalent: "")
        petSettingsItem.target = self
        petMenu.addItem(petSettingsItem)
        
        petItem.submenu = petMenu
        toolsMenu.addItem(petItem)
        
        // 文件传输
        let transferItem = NSMenuItem(title: "文件传输", action: #selector(showFileTransfer(_:)), keyEquivalent: "")
        transferItem.target = self
        toolsMenu.addItem(transferItem)
        
        // 聊天助手
        let chatItem = NSMenuItem(title: "聊天助手", action: #selector(showChat(_:)), keyEquivalent: "")
        chatItem.target = self
        toolsMenu.addItem(chatItem)
        
        toolsMenu.addItem(NSMenuItem.separator())
        
        // 数据同步
        let syncItem = NSMenuItem(title: "数据同步", action: #selector(syncData(_:)), keyEquivalent: "")
        syncItem.target = self
        toolsMenu.addItem(syncItem)
        
        // 备份与恢复
        let backupItem = NSMenuItem(title: "备份与恢复", action: nil, keyEquivalent: "")
        let backupMenu = NSMenu(title: "备份与恢复")
        
        let createBackupItem = NSMenuItem(title: "创建备份...", action: #selector(createBackup(_:)), keyEquivalent: "")
        createBackupItem.target = self
        backupMenu.addItem(createBackupItem)
        
        let restoreBackupItem = NSMenuItem(title: "恢复备份...", action: #selector(restoreBackup(_:)), keyEquivalent: "")
        restoreBackupItem.target = self
        backupMenu.addItem(restoreBackupItem)
        
        backupItem.submenu = backupMenu
        toolsMenu.addItem(backupItem)
        
        toolsMenu.addItem(NSMenuItem.separator())
        
        // 开发者工具
        let devItem = NSMenuItem(title: "开发者工具", action: nil, keyEquivalent: "")
        let devMenu = NSMenu(title: "开发者工具")
        
        let consoleItem = NSMenuItem(title: "控制台", action: #selector(showConsole(_:)), keyEquivalent: "")
        consoleItem.target = self
        devMenu.addItem(consoleItem)
        
        let logViewerItem = NSMenuItem(title: "日志查看器", action: #selector(showLogViewer(_:)), keyEquivalent: "")
        logViewerItem.target = self
        devMenu.addItem(logViewerItem)
        
        devItem.submenu = devMenu
        toolsMenu.addItem(devItem)
        
        toolsMenuItem.submenu = toolsMenu
        mainMenu?.addItem(toolsMenuItem)
    }
    
    private func setupWindowMenu() {
        let windowMenuItem = NSMenuItem(title: "窗口", action: nil, keyEquivalent: "")
        let windowMenu = NSMenu(title: "窗口")
        
        // 最小化
        let minimizeItem = NSMenuItem(title: "最小化", action: #selector(NSWindow.performMiniaturize(_:)), keyEquivalent: "m")
        windowMenu.addItem(minimizeItem)
        
        // 缩放
        let zoomItem = NSMenuItem(title: "缩放", action: #selector(NSWindow.performZoom(_:)), keyEquivalent: "")
        windowMenu.addItem(zoomItem)
        
        windowMenu.addItem(NSMenuItem.separator())
        
        // 前置全部窗口
        let bringAllToFrontItem = NSMenuItem(title: "前置全部窗口", action: #selector(NSApplication.arrangeInFront(_:)), keyEquivalent: "")
        windowMenu.addItem(bringAllToFrontItem)
        
        windowMenuItem.submenu = windowMenu
        mainMenu?.addItem(windowMenuItem)
        
        NSApp.windowsMenu = windowMenu
    }
    
    private func setupHelpMenu() {
        let helpMenuItem = NSMenuItem(title: "帮助", action: nil, keyEquivalent: "")
        let helpMenu = NSMenu(title: "帮助")
        
        // 用户指南
        let userGuideItem = NSMenuItem(title: "用户指南", action: #selector(showUserGuide(_:)), keyEquivalent: "")
        userGuideItem.target = self
        helpMenu.addItem(userGuideItem)
        
        // 快捷键
        let shortcutsItem = NSMenuItem(title: "快捷键", action: #selector(showShortcuts(_:)), keyEquivalent: "")
        shortcutsItem.target = self
        helpMenu.addItem(shortcutsItem)
        
        helpMenu.addItem(NSMenuItem.separator())
        
        // 反馈
        let feedbackItem = NSMenuItem(title: "发送反馈", action: #selector(sendFeedback(_:)), keyEquivalent: "")
        feedbackItem.target = self
        helpMenu.addItem(feedbackItem)
        
        // 报告问题
        let reportIssueItem = NSMenuItem(title: "报告问题", action: #selector(reportIssue(_:)), keyEquivalent: "")
        reportIssueItem.target = self
        helpMenu.addItem(reportIssueItem)
        
        helpMenu.addItem(NSMenuItem.separator())
        
        // 检查更新
        let checkUpdatesItem = NSMenuItem(title: "检查更新...", action: #selector(checkForUpdates(_:)), keyEquivalent: "")
        checkUpdatesItem.target = self
        helpMenu.addItem(checkUpdatesItem)
        
        // 关于
        let aboutHelpItem = NSMenuItem(title: "关于太上老君AI", action: #selector(showAbout(_:)), keyEquivalent: "")
        aboutHelpItem.target = self
        helpMenu.addItem(aboutHelpItem)
        
        helpMenuItem.submenu = helpMenu
        mainMenu?.addItem(helpMenuItem)
        
        NSApp.helpMenu = helpMenu
    }
    
    // MARK: - Status Bar Setup
    private func setupStatusBar() {
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        
        if let button = statusItem?.button {
            button.image = NSImage(systemSymbolName: "sparkles", accessibilityDescription: "太上老君AI")
            button.target = self
            button.action = #selector(statusBarClicked(_:))
            button.sendAction(on: [.leftMouseUp, .rightMouseUp])
        }
        
        setupStatusBarMenu()
    }
    
    private func setupStatusBarMenu() {
        statusBarMenu = NSMenu()
        
        // 显示主窗口
        let showMainItem = NSMenuItem(title: "显示主窗口", action: #selector(showMainWindow(_:)), keyEquivalent: "")
        showMainItem.target = self
        statusBarMenu?.addItem(showMainItem)
        
        statusBarMenu?.addItem(NSMenuItem.separator())
        
        // 桌面宠物
        let petStatusItem = NSMenuItem(title: "桌面宠物", action: #selector(toggleDesktopPet(_:)), keyEquivalent: "")
        petStatusItem.target = self
        statusBarMenu?.addItem(petStatusItem)
        
        // 快速聊天
        let quickChatItem = NSMenuItem(title: "快速聊天", action: #selector(showQuickChat(_:)), keyEquivalent: "")
        quickChatItem.target = self
        statusBarMenu?.addItem(quickChatItem)
        
        // 文件传输
        let transferStatusItem = NSMenuItem(title: "文件传输", action: #selector(showFileTransfer(_:)), keyEquivalent: "")
        transferStatusItem.target = self
        statusBarMenu?.addItem(transferStatusItem)
        
        statusBarMenu?.addItem(NSMenuItem.separator())
        
        // 同步状态
        let syncStatusItem = NSMenuItem(title: "同步状态: 已连接", action: nil, keyEquivalent: "")
        syncStatusItem.isEnabled = false
        statusBarMenu?.addItem(syncStatusItem)
        
        // 网络状态
        let networkStatusItem = NSMenuItem(title: "网络状态: 正常", action: nil, keyEquivalent: "")
        networkStatusItem.isEnabled = false
        statusBarMenu?.addItem(networkStatusItem)
        
        statusBarMenu?.addItem(NSMenuItem.separator())
        
        // 偏好设置
        let prefsStatusItem = NSMenuItem(title: "偏好设置...", action: #selector(showPreferences(_:)), keyEquivalent: "")
        prefsStatusItem.target = self
        statusBarMenu?.addItem(prefsStatusItem)
        
        statusBarMenu?.addItem(NSMenuItem.separator())
        
        // 退出
        let quitStatusItem = NSMenuItem(title: "退出", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "")
        statusBarMenu?.addItem(quitStatusItem)
    }
    
    // MARK: - Public Methods
    func updateStatusBarIcon(_ imageName: String) {
        statusItem?.button?.image = NSImage(systemSymbolName: imageName, accessibilityDescription: "太上老君AI")
    }
    
    func updateStatusBarTitle(_ title: String) {
        statusItem?.button?.title = title
    }
    
    func showStatusBarMenu() {
        guard let menu = statusBarMenu else { return }
        statusItem?.menu = menu
        statusItem?.button?.performClick(nil)
        statusItem?.menu = nil
    }
    
    func hideStatusBar() {
        statusItem = nil
    }
    
    func showStatusBar() {
        if statusItem == nil {
            setupStatusBar()
        }
    }
    
    // MARK: - Menu Actions
    @objc private func showAbout(_ sender: NSMenuItem) {
        let aboutPanel = NSAlert()
        aboutPanel.messageText = "太上老君AI"
        aboutPanel.informativeText = """
        版本 1.0.0
        
        一个智能的桌面助手应用，集成了AI聊天、文件管理、
        桌面宠物等功能。
        
        © 2024 太上老君AI团队
        """
        aboutPanel.alertStyle = .informational
        aboutPanel.runModal()
    }
    
    @objc private func showPreferences(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showSettings, object: nil)
    }
    
    @objc private func newProject(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .newProject, object: nil)
    }
    
    @objc private func openProject(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .openProject, object: nil)
    }
    
    @objc private func save(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .saveProject, object: nil)
    }
    
    @objc private func saveAs(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .saveProjectAs, object: nil)
    }
    
    @objc private func importFile(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .importFile, object: nil)
    }
    
    @objc private func exportFile(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .exportFile, object: nil)
    }
    
    @objc private func findInProject(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .findInProject, object: nil)
    }
    
    @objc private func findAndReplace(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .findAndReplace, object: nil)
    }
    
    @objc private func toggleSidebar(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .toggleSidebar, object: nil)
    }
    
    @objc private func toggleToolbar(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .toggleToolbar, object: nil)
    }
    
    @objc private func toggleStatusBar(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .toggleStatusBar, object: nil)
    }
    
    @objc private func zoomIn(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .zoomIn, object: nil)
    }
    
    @objc private func zoomOut(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .zoomOut, object: nil)
    }
    
    @objc private func actualSize(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .actualSize, object: nil)
    }
    
    @objc private func showDesktopPet(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showDesktopPet, object: nil)
    }
    
    @objc private func hideDesktopPet(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .hideDesktopPet, object: nil)
    }
    
    @objc private func toggleDesktopPet(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .toggleDesktopPet, object: nil)
    }
    
    @objc private func showPetSettings(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showPetSettings, object: nil)
    }
    
    @objc private func showFileTransfer(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showFileTransfer, object: nil)
    }
    
    @objc private func showChat(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showChat, object: nil)
    }
    
    @objc private func showQuickChat(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showQuickChat, object: nil)
    }
    
    @objc private func syncData(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .syncData, object: nil)
    }
    
    @objc private func createBackup(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .createBackup, object: nil)
    }
    
    @objc private func restoreBackup(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .restoreBackup, object: nil)
    }
    
    @objc private func showConsole(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showConsole, object: nil)
    }
    
    @objc private func showLogViewer(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showLogViewer, object: nil)
    }
    
    @objc private func showUserGuide(_ sender: NSMenuItem) {
        if let url = URL(string: "https://taishanglaojun.ai/guide") {
            NSWorkspace.shared.open(url)
        }
    }
    
    @objc private func showShortcuts(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showShortcuts, object: nil)
    }
    
    @objc private func sendFeedback(_ sender: NSMenuItem) {
        if let url = URL(string: "https://taishanglaojun.ai/feedback") {
            NSWorkspace.shared.open(url)
        }
    }
    
    @objc private func reportIssue(_ sender: NSMenuItem) {
        if let url = URL(string: "https://github.com/taishanglaojun/issues") {
            NSWorkspace.shared.open(url)
        }
    }
    
    @objc private func checkForUpdates(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .checkForUpdates, object: nil)
    }
    
    @objc private func statusBarClicked(_ sender: NSStatusBarButton) {
        let event = NSApp.currentEvent!
        
        if event.type == .rightMouseUp {
            showStatusBarMenu()
        } else {
            NotificationCenter.default.post(name: .showMainWindow, object: nil)
        }
    }
    
    @objc private func showMainWindow(_ sender: NSMenuItem) {
        NotificationCenter.default.post(name: .showMainWindow, object: nil)
    }
}

// MARK: - Notification Names
extension Notification.Name {
    static let showSettings = Notification.Name("showSettings")
    static let newProject = Notification.Name("newProject")
    static let openProject = Notification.Name("openProject")
    static let saveProject = Notification.Name("saveProject")
    static let saveProjectAs = Notification.Name("saveProjectAs")
    static let importFile = Notification.Name("importFile")
    static let exportFile = Notification.Name("exportFile")
    static let findInProject = Notification.Name("findInProject")
    static let findAndReplace = Notification.Name("findAndReplace")
    static let toggleSidebar = Notification.Name("toggleSidebar")
    static let toggleToolbar = Notification.Name("toggleToolbar")
    static let toggleStatusBar = Notification.Name("toggleStatusBar")
    static let zoomIn = Notification.Name("zoomIn")
    static let zoomOut = Notification.Name("zoomOut")
    static let actualSize = Notification.Name("actualSize")
    static let showDesktopPet = Notification.Name("showDesktopPet")
    static let hideDesktopPet = Notification.Name("hideDesktopPet")
    static let toggleDesktopPet = Notification.Name("toggleDesktopPet")
    static let showPetSettings = Notification.Name("showPetSettings")
    static let showFileTransfer = Notification.Name("showFileTransfer")
    static let showChat = Notification.Name("showChat")
    static let showQuickChat = Notification.Name("showQuickChat")
    static let syncData = Notification.Name("syncData")
    static let createBackup = Notification.Name("createBackup")
    static let restoreBackup = Notification.Name("restoreBackup")
    static let showConsole = Notification.Name("showConsole")
    static let showLogViewer = Notification.Name("showLogViewer")
    static let showShortcuts = Notification.Name("showShortcuts")
    static let checkForUpdates = Notification.Name("checkForUpdates")
    static let showMainWindow = Notification.Name("showMainWindow")
}