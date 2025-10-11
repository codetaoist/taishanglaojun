import Cocoa
import Metal
import MetalKit

class MainWindowController: NSWindowController {
    
    // MARK: - Properties
    private var mainViewController: MainViewController?
    private var metalView: MTKView?
    private var metalRenderer: MetalRenderer?
    
    // MARK: - View Controllers
    private var projectViewController: ProjectViewController?
    private var chatViewController: ChatViewController?
    private var fileTransferViewController: FileTransferViewController?
    private var settingsViewController: SettingsViewController?
    
    // MARK: - Initialization
    
    convenience init() {
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 1200, height: 800),
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        
        self.init(window: window)
        setupWindow()
        setupMetalView()
        setupViewController()
    }
    
    // MARK: - Window Setup
    
    private func setupWindow() {
        guard let window = window else { return }
        
        window.title = "太上老君AI平台"
        window.minSize = NSSize(width: 800, height: 600)
        window.center()
        window.setFrameAutosaveName("MainWindow")
        
        // 设置窗口外观
        window.titlebarAppearsTransparent = false
        window.titleVisibility = .visible
        
        // 设置窗口级别
        window.level = .normal
        
        // 设置窗口代理
        window.delegate = self
    }
    
    private func setupMetalView() {
        guard let window = window else { return }
        
        // 创建Metal设备
        guard let device = MTLCreateSystemDefaultDevice() else {
            NSLog("Metal不可用，使用软件渲染")
            return
        }
        
        // 创建Metal视图
        metalView = MTKView(frame: window.contentView?.bounds ?? .zero, device: device)
        metalView?.autoresizingMask = [.width, .height]
        metalView?.colorPixelFormat = .bgra8Unorm
        metalView?.clearColor = MTLClearColor(red: 0.95, green: 0.95, blue: 0.97, alpha: 1.0)
        
        // 创建Metal渲染器
        if let metalView = metalView {
            metalRenderer = MetalRenderer(metalView: metalView)
            metalView.delegate = metalRenderer
        }
        
        // 添加到窗口
        if let metalView = metalView {
            window.contentView?.addSubview(metalView)
        }
    }
    
    private func setupViewController() {
        mainViewController = MainViewController()
        
        if let mainViewController = mainViewController {
            window?.contentViewController = mainViewController
            
            // 设置Metal视图到主视图控制器
            if let metalView = metalView {
                mainViewController.setMetalView(metalView)
            }
        }
    }
    
    // MARK: - Public Methods
    
    func showWindow() {
        window?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    func hideWindow() {
        window?.orderOut(nil)
    }
    
    func toggleWindow() {
        if window?.isVisible == true {
            hideWindow()
        } else {
            showWindow()
        }
    }
    
    // MARK: - View Management
    func showProjects() {
        if projectViewController == nil {
            projectViewController = ProjectViewController()
        }
        mainViewController?.showContent(projectViewController!)
    }
    
    func showChat() {
        if chatViewController == nil {
            chatViewController = ChatViewController()
        }
        mainViewController?.showContent(chatViewController!)
    }
    
    func showFileTransfer() {
        if fileTransferViewController == nil {
            fileTransferViewController = FileTransferViewController()
        }
        mainViewController?.showContent(fileTransferViewController!)
    }
    
    func showSettings() {
        if settingsViewController == nil {
            settingsViewController = SettingsViewController()
        }
        mainViewController?.showContent(settingsViewController!)
    }
    
    func updateMetalRenderer() {
        metalRenderer?.updateFrame()
    }
}

// MARK: - NSWindowDelegate

extension MainWindowController: NSWindowDelegate {
    
    func windowShouldClose(_ sender: NSWindow) -> Bool {
        // 不直接关闭窗口，而是隐藏到系统托盘
        hideWindow()
        return false
    }
    
    func windowDidResize(_ notification: Notification) {
        // 更新Metal视图大小
        metalRenderer?.updateViewport()
    }
    
    func windowDidMiniaturize(_ notification: Notification) {
        // 窗口最小化时的处理
        NSLog("主窗口已最小化")
    }
    
    func windowDidDeminiaturize(_ notification: Notification) {
        // 窗口恢复时的处理
        NSLog("主窗口已恢复")
    }
    
    func windowDidBecomeKey(_ notification: Notification) {
        // 窗口获得焦点时的处理
        mainViewController?.windowDidBecomeActive()
    }
    
    func windowDidResignKey(_ notification: Notification) {
        // 窗口失去焦点时的处理
        mainViewController?.windowDidResignActive()
    }
}

// MARK: - MainViewController

class MainViewController: NSViewController {
    
    // MARK: - Properties
    private var metalView: MTKView?
    private var toolbarView: NSView?
    private var sidebarView: NSView?
    private var contentView: NSView?
    private var statusView: NSView?
    
    // UI组件
    private var projectListView: NSOutlineView?
    private var chatView: NSView?
    private var fileTransferView: NSView?
    private var settingsView: NSView?
    
    // 当前显示的视图
    private var currentContentView: NSView?
    
    // MARK: - Lifecycle
    
    override func loadView() {
        view = NSView(frame: NSRect(x: 0, y: 0, width: 1200, height: 800))
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        setupConstraints()
        setupDefaultContent()
    }
    
    // MARK: - UI Setup
    
    private func setupUI() {
        setupToolbar()
        setupSidebar()
        setupContentArea()
        setupStatusBar()
    }
    
    private func setupToolbar() {
        toolbarView = NSView()
        toolbarView?.wantsLayer = true
        toolbarView?.layer?.backgroundColor = NSColor.controlColor.cgColor
        
        // 创建工具栏按钮
        let projectButton = createToolbarButton(title: "项目", action: #selector(showProjectView))
        let chatButton = createToolbarButton(title: "聊天", action: #selector(showChatView))
        let transferButton = createToolbarButton(title: "传输", action: #selector(showTransferView))
        let settingsButton = createToolbarButton(title: "设置", action: #selector(showSettingsView))
        
        if let toolbarView = toolbarView {
            view.addSubview(toolbarView)
            toolbarView.addSubview(projectButton)
            toolbarView.addSubview(chatButton)
            toolbarView.addSubview(transferButton)
            toolbarView.addSubview(settingsButton)
        }
    }
    
    private func setupSidebar() {
        sidebarView = NSView()
        sidebarView?.wantsLayer = true
        sidebarView?.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        
        // 创建项目列表
        projectListView = NSOutlineView()
        let scrollView = NSScrollView()
        scrollView.documentView = projectListView
        scrollView.hasVerticalScroller = true
        
        if let sidebarView = sidebarView {
            view.addSubview(sidebarView)
            sidebarView.addSubview(scrollView)
        }
    }
    
    private func setupContentArea() {
        contentView = NSView()
        contentView?.wantsLayer = true
        contentView?.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        
        if let contentView = contentView {
            view.addSubview(contentView)
        }
    }
    
    private func setupStatusBar() {
        statusView = NSView()
        statusView?.wantsLayer = true
        statusView?.layer?.backgroundColor = NSColor.controlColor.cgColor
        
        // 创建状态标签
        let statusLabel = NSTextField(labelWithString: "就绪")
        statusLabel.font = NSFont.systemFont(ofSize: 12)
        statusLabel.textColor = NSColor.secondaryLabelColor
        
        if let statusView = statusView {
            view.addSubview(statusView)
            statusView.addSubview(statusLabel)
        }
    }
    
    private func setupConstraints() {
        // 使用Auto Layout设置约束
        guard let toolbarView = toolbarView,
              let sidebarView = sidebarView,
              let contentView = contentView,
              let statusView = statusView else { return }
        
        toolbarView.translatesAutoresizingMaskIntoConstraints = false
        sidebarView.translatesAutoresizingMaskIntoConstraints = false
        contentView.translatesAutoresizingMaskIntoConstraints = false
        statusView.translatesAutoresizingMaskIntoConstraints = false
        
        NSLayoutConstraint.activate([
            // 工具栏约束
            toolbarView.topAnchor.constraint(equalTo: view.topAnchor),
            toolbarView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            toolbarView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            toolbarView.heightAnchor.constraint(equalToConstant: 44),
            
            // 侧边栏约束
            sidebarView.topAnchor.constraint(equalTo: toolbarView.bottomAnchor),
            sidebarView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            sidebarView.widthAnchor.constraint(equalToConstant: 250),
            sidebarView.bottomAnchor.constraint(equalTo: statusView.topAnchor),
            
            // 内容区域约束
            contentView.topAnchor.constraint(equalTo: toolbarView.bottomAnchor),
            contentView.leadingAnchor.constraint(equalTo: sidebarView.trailingAnchor),
            contentView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            contentView.bottomAnchor.constraint(equalTo: statusView.topAnchor),
            
            // 状态栏约束
            statusView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            statusView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            statusView.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            statusView.heightAnchor.constraint(equalToConstant: 24)
        ])
    }
    
    private func setupDefaultContent() {
        showProjectView()
    }
    
    // MARK: - Helper Methods
    
    private func createToolbarButton(title: String, action: Selector) -> NSButton {
        let button = NSButton()
        button.title = title
        button.target = self
        button.action = action
        button.bezelStyle = .rounded
        button.font = NSFont.systemFont(ofSize: 13)
        return button
    }
    
    // MARK: - Content Management
    
    private func showContentView(_ newView: NSView) {
        // 移除当前内容视图
        currentContentView?.removeFromSuperview()
        
        // 添加新的内容视图
        contentView?.addSubview(newView)
        newView.translatesAutoresizingMaskIntoConstraints = false
        
        if let contentView = contentView {
            NSLayoutConstraint.activate([
                newView.topAnchor.constraint(equalTo: contentView.topAnchor),
                newView.leadingAnchor.constraint(equalTo: contentView.leadingAnchor),
                newView.trailingAnchor.constraint(equalTo: contentView.trailingAnchor),
                newView.bottomAnchor.constraint(equalTo: contentView.bottomAnchor)
            ])
        }
        
        currentContentView = newView
    }
    
    func showContent(_ viewController: NSViewController) {
        // 移除当前内容视图控制器
        if let currentChild = children.first {
            currentChild.removeFromParent()
            currentChild.view.removeFromSuperview()
        }
        
        // 添加新的视图控制器
        addChild(viewController)
        
        if let contentView = contentView {
            contentView.addSubview(viewController.view)
            
            // 设置约束
            viewController.view.translatesAutoresizingMaskIntoConstraints = false
            NSLayoutConstraint.activate([
                viewController.view.topAnchor.constraint(equalTo: contentView.topAnchor),
                viewController.view.leadingAnchor.constraint(equalTo: contentView.leadingAnchor),
                viewController.view.trailingAnchor.constraint(equalTo: contentView.trailingAnchor),
                viewController.view.bottomAnchor.constraint(equalTo: contentView.bottomAnchor)
            ])
        }
    }
    
    // MARK: - Action Methods
    
    @objc private func showProjectView() {
        let projectView = createProjectView()
        showContentView(projectView)
    }
    
    @objc private func showChatView() {
        let chatView = createChatView()
        showContentView(chatView)
    }
    
    @objc private func showTransferView() {
        let transferView = createTransferView()
        showContentView(transferView)
    }
    
    @objc private func showSettingsView() {
        let settingsView = createSettingsView()
        showContentView(settingsView)
    }
    
    // MARK: - Content View Creation
    
    private func createProjectView() -> NSView {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        
        let label = NSTextField(labelWithString: "项目管理")
        label.font = NSFont.boldSystemFont(ofSize: 18)
        label.alignment = .center
        
        view.addSubview(label)
        label.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            label.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            label.centerYAnchor.constraint(equalTo: view.centerYAnchor)
        ])
        
        return view
    }
    
    private func createChatView() -> NSView {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        
        let label = NSTextField(labelWithString: "AI聊天")
        label.font = NSFont.boldSystemFont(ofSize: 18)
        label.alignment = .center
        
        view.addSubview(label)
        label.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            label.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            label.centerYAnchor.constraint(equalTo: view.centerYAnchor)
        ])
        
        return view
    }
    
    private func createTransferView() -> NSView {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        
        let label = NSTextField(labelWithString: "文件传输")
        label.font = NSFont.boldSystemFont(ofSize: 18)
        label.alignment = .center
        
        view.addSubview(label)
        label.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            label.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            label.centerYAnchor.constraint(equalTo: view.centerYAnchor)
        ])
        
        return view
    }
    
    private func createSettingsView() -> NSView {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        
        let label = NSTextField(labelWithString: "设置")
        label.font = NSFont.boldSystemFont(ofSize: 18)
        label.alignment = .center
        
        view.addSubview(label)
        label.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            label.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            label.centerYAnchor.constraint(equalTo: view.centerYAnchor)
        ])
        
        return view
    }
    
    // MARK: - Metal View Management
    
    func setMetalView(_ metalView: MTKView) {
        self.metalView = metalView
    }
    
    // MARK: - Window State Management
    
    func windowDidBecomeActive() {
        // 窗口激活时的处理
    }
    
    func windowDidResignActive() {
        // 窗口失去激活状态时的处理
    }
}