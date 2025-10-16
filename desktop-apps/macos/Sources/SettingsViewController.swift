import Cocoa
import Foundation

// MARK: - Settings Models
struct AppSettings {
    var general: GeneralSettings
    var appearance: AppearanceSettings
    var notifications: NotificationSettings
    var privacy: PrivacySettings
    var advanced: AdvancedSettings
    
    static let shared = AppSettings(
        general: GeneralSettings(),
        appearance: AppearanceSettings(),
        notifications: NotificationSettings(),
        privacy: PrivacySettings(),
        advanced: AdvancedSettings()
    )
}

struct GeneralSettings {
    var autoStart: Bool = false
    var minimizeToTray: Bool = true
    var closeToTray: Bool = true
    var language: String = "zh-CN"
    var checkUpdates: Bool = true
    var updateChannel: UpdateChannel = .stable
}

struct AppearanceSettings {
    var theme: AppTheme = .system
    var accentColor: AccentColor = .blue
    var fontSize: FontSize = .medium
    var showMenuBar: Bool = true
    var showStatusBar: Bool = true
    var windowOpacity: Double = 1.0
}

struct NotificationSettings {
    var enableNotifications: Bool = true
    var soundEnabled: Bool = true
    var badgeEnabled: Bool = true
    var petNotifications: Bool = true
    var fileTransferNotifications: Bool = true
    var chatNotifications: Bool = true
    var systemNotifications: Bool = true
}

struct PrivacySettings {
    var analyticsEnabled: Bool = false
    var crashReportsEnabled: Bool = true
    var locationEnabled: Bool = false
    var cameraEnabled: Bool = false
    var microphoneEnabled: Bool = false
    var dataCollection: DataCollectionLevel = .minimal
}

struct AdvancedSettings {
    var debugMode: Bool = false
    var logLevel: LogLevel = .info
    var maxLogFiles: Int = 10
    var networkTimeout: Int = 30
    var maxConcurrentTransfers: Int = 3
    var cacheSize: Int = 100
}

enum AppTheme: String, CaseIterable {
    case light = "浅色"
    case dark = "深色"
    case system = "跟随系统"
}

enum AccentColor: String, CaseIterable {
    case blue = "蓝色"
    case purple = "紫色"
    case pink = "粉色"
    case red = "红色"
    case orange = "橙色"
    case yellow = "黄色"
    case green = "绿色"
    case graphite = "石墨色"
    
    var color: NSColor {
        switch self {
        case .blue: return .systemBlue
        case .purple: return .systemPurple
        case .pink: return .systemPink
        case .red: return .systemRed
        case .orange: return .systemOrange
        case .yellow: return .systemYellow
        case .green: return .systemGreen
        case .graphite: return .systemGray
        }
    }
}

enum FontSize: String, CaseIterable {
    case small = "小"
    case medium = "中"
    case large = "大"
    
    var size: CGFloat {
        switch self {
        case .small: return 12
        case .medium: return 14
        case .large: return 16
        }
    }
}

enum UpdateChannel: String, CaseIterable {
    case stable = "稳定版"
    case beta = "测试版"
    case nightly = "每夜版"
}

enum DataCollectionLevel: String, CaseIterable {
    case none = "无"
    case minimal = "最少"
    case standard = "标准"
    case full = "完整"
}

enum LogLevel: String, CaseIterable {
    case error = "错误"
    case warning = "警告"
    case info = "信息"
    case debug = "调试"
    case verbose = "详细"
}

// MARK: - Settings View Controller
class SettingsViewController: NSViewController {
    
    // MARK: - Properties
    private var settings = AppSettings.shared
    private var currentCategory: SettingsCategory = .general
    
    // MARK: - UI Components
    private lazy var splitView: NSSplitView = {
        let splitView = NSSplitView()
        splitView.isVertical = true
        splitView.dividerStyle = .thin
        return splitView
    }()
    
    private lazy var sidebar: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var contentArea: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        return view
    }()
    
    private lazy var categoryTableView: NSTableView = {
        let tableView = NSTableView()
        tableView.delegate = self
        tableView.dataSource = self
        tableView.target = self
        tableView.action = #selector(selectCategory(_:))
        
        let column = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("category"))
        column.title = "设置分类"
        tableView.addTableColumn(column)
        tableView.headerView = nil
        
        return tableView
    }()
    
    private lazy var categoryScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.documentView = categoryTableView
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        return scrollView
    }()
    
    private lazy var contentScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        return scrollView
    }()
    
    private lazy var contentView: NSView = {
        let view = NSView()
        return view
    }()
    
    private lazy var titleLabel: NSTextField = {
        let label = NSTextField(labelWithString: "通用")
        label.font = NSFont.boldSystemFont(ofSize: 18)
        return label
    }()
    
    private lazy var resetButton: NSButton = {
        let button = NSButton()
        button.title = "重置为默认"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(resetSettings(_:))
        return button
    }()
    
    private lazy var exportButton: NSButton = {
        let button = NSButton()
        button.title = "导出设置"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(exportSettings(_:))
        return button
    }()
    
    private lazy var importButton: NSButton = {
        let button = NSButton()
        button.title = "导入设置"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(importSettings(_:))
        return button
    }()
    
    // MARK: - Lifecycle
    override func loadView() {
        view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.windowBackgroundColor.cgColor
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        loadSettings()
        selectDefaultCategory()
    }
    
    // MARK: - UI Setup
    private func setupUI() {
        setupSplitView()
        setupSidebar()
        setupContentArea()
        setupConstraints()
    }
    
    private func setupSplitView() {
        view.addSubview(splitView)
        splitView.addArrangedSubview(sidebar)
        splitView.addArrangedSubview(contentArea)
        
        // 设置分割比例
        splitView.setPosition(200, ofDividerAt: 0)
    }
    
    private func setupSidebar() {
        sidebar.addSubview(categoryScrollView)
    }
    
    private func setupContentArea() {
        contentArea.addSubview(titleLabel)
        contentArea.addSubview(contentScrollView)
        contentArea.addSubview(resetButton)
        contentArea.addSubview(exportButton)
        contentArea.addSubview(importButton)
        
        contentScrollView.documentView = contentView
    }
    
    private func setupConstraints() {
        // 禁用自动调整大小
        [splitView, categoryScrollView, titleLabel, contentScrollView, resetButton, exportButton, importButton].forEach {
            $0.translatesAutoresizingMaskIntoConstraints = false
        }
        
        NSLayoutConstraint.activate([
            // Split View
            splitView.topAnchor.constraint(equalTo: view.topAnchor),
            splitView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            splitView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            splitView.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            
            // Sidebar
            categoryScrollView.topAnchor.constraint(equalTo: sidebar.topAnchor, constant: 20),
            categoryScrollView.leadingAnchor.constraint(equalTo: sidebar.leadingAnchor, constant: 10),
            categoryScrollView.trailingAnchor.constraint(equalTo: sidebar.trailingAnchor, constant: -10),
            categoryScrollView.bottomAnchor.constraint(equalTo: sidebar.bottomAnchor, constant: -20),
            
            // Content Area
            titleLabel.topAnchor.constraint(equalTo: contentArea.topAnchor, constant: 20),
            titleLabel.leadingAnchor.constraint(equalTo: contentArea.leadingAnchor, constant: 20),
            
            contentScrollView.topAnchor.constraint(equalTo: titleLabel.bottomAnchor, constant: 20),
            contentScrollView.leadingAnchor.constraint(equalTo: contentArea.leadingAnchor, constant: 20),
            contentScrollView.trailingAnchor.constraint(equalTo: contentArea.trailingAnchor, constant: -20),
            contentScrollView.bottomAnchor.constraint(equalTo: resetButton.topAnchor, constant: -20),
            
            // Bottom buttons
            resetButton.leadingAnchor.constraint(equalTo: contentArea.leadingAnchor, constant: 20),
            resetButton.bottomAnchor.constraint(equalTo: contentArea.bottomAnchor, constant: -20),
            
            exportButton.leadingAnchor.constraint(equalTo: resetButton.trailingAnchor, constant: 10),
            exportButton.bottomAnchor.constraint(equalTo: contentArea.bottomAnchor, constant: -20),
            
            importButton.leadingAnchor.constraint(equalTo: exportButton.trailingAnchor, constant: 10),
            importButton.bottomAnchor.constraint(equalTo: contentArea.bottomAnchor, constant: -20)
        ])
    }
    
    // MARK: - Data Management
    private func loadSettings() {
        // 从用户偏好设置加载配置
        categoryTableView.reloadData()
    }
    
    private func selectDefaultCategory() {
        categoryTableView.selectRowIndexes(IndexSet(integer: 0), byExtendingSelection: false)
        showCategoryContent(.general)
    }
    
    private func showCategoryContent(_ category: SettingsCategory) {
        currentCategory = category
        titleLabel.stringValue = category.title
        
        // 清除现有内容
        contentView.subviews.forEach { $0.removeFromSuperview() }
        
        // 根据分类显示相应内容
        switch category {
        case .general:
            setupGeneralSettings()
        case .appearance:
            setupAppearanceSettings()
        case .notifications:
            setupNotificationSettings()
        case .privacy:
            setupPrivacySettings()
        case .advanced:
            setupAdvancedSettings()
        }
        
        // 更新内容视图大小
        updateContentViewSize()
    }
    
    private func updateContentViewSize() {
        let height = max(contentView.subviews.map { $0.frame.maxY }.max() ?? 0, 400)
        contentView.frame = NSRect(x: 0, y: 0, width: contentScrollView.frame.width, height: height)
    }
    
    // MARK: - Settings Panels
    private func setupGeneralSettings() {
        var yPosition: CGFloat = 20
        
        // 启动设置
        let startupSection = createSectionLabel("启动设置", y: yPosition)
        contentView.addSubview(startupSection)
        yPosition += 40
        
        let autoStartCheckbox = createCheckbox("开机自动启动", value: settings.general.autoStart, y: yPosition) { [weak self] value in
            self?.settings.general.autoStart = value
        }
        contentView.addSubview(autoStartCheckbox)
        yPosition += 30
        
        let minimizeCheckbox = createCheckbox("最小化到系统托盘", value: settings.general.minimizeToTray, y: yPosition) { [weak self] value in
            self?.settings.general.minimizeToTray = value
        }
        contentView.addSubview(minimizeCheckbox)
        yPosition += 30
        
        let closeCheckbox = createCheckbox("关闭时最小化到托盘", value: settings.general.closeToTray, y: yPosition) { [weak self] value in
            self?.settings.general.closeToTray = value
        }
        contentView.addSubview(closeCheckbox)
        yPosition += 50
        
        // 语言设置
        let languageSection = createSectionLabel("语言设置", y: yPosition)
        contentView.addSubview(languageSection)
        yPosition += 40
        
        let languagePopup = createPopupButton(["简体中文", "English", "日本語"], selectedIndex: 0, y: yPosition) { [weak self] index in
            let languages = ["zh-CN", "en-US", "ja-JP"]
            self?.settings.general.language = languages[index]
        }
        contentView.addSubview(languagePopup)
        yPosition += 50
        
        // 更新设置
        let updateSection = createSectionLabel("更新设置", y: yPosition)
        contentView.addSubview(updateSection)
        yPosition += 40
        
        let updateCheckbox = createCheckbox("自动检查更新", value: settings.general.checkUpdates, y: yPosition) { [weak self] value in
            self?.settings.general.checkUpdates = value
        }
        contentView.addSubview(updateCheckbox)
        yPosition += 30
        
        let channelPopup = createPopupButton(UpdateChannel.allCases.map { $0.rawValue }, selectedIndex: 0, y: yPosition) { [weak self] index in
            self?.settings.general.updateChannel = UpdateChannel.allCases[index]
        }
        contentView.addSubview(channelPopup)
    }
    
    private func setupAppearanceSettings() {
        var yPosition: CGFloat = 20
        
        // 主题设置
        let themeSection = createSectionLabel("主题设置", y: yPosition)
        contentView.addSubview(themeSection)
        yPosition += 40
        
        let themePopup = createPopupButton(AppTheme.allCases.map { $0.rawValue }, selectedIndex: 2, y: yPosition) { [weak self] index in
            self?.settings.appearance.theme = AppTheme.allCases[index]
        }
        contentView.addSubview(themePopup)
        yPosition += 50
        
        // 强调色设置
        let accentSection = createSectionLabel("强调色", y: yPosition)
        contentView.addSubview(accentSection)
        yPosition += 40
        
        let accentPopup = createPopupButton(AccentColor.allCases.map { $0.rawValue }, selectedIndex: 0, y: yPosition) { [weak self] index in
            self?.settings.appearance.accentColor = AccentColor.allCases[index]
        }
        contentView.addSubview(accentPopup)
        yPosition += 50
        
        // 字体大小
        let fontSection = createSectionLabel("字体大小", y: yPosition)
        contentView.addSubview(fontSection)
        yPosition += 40
        
        let fontPopup = createPopupButton(FontSize.allCases.map { $0.rawValue }, selectedIndex: 1, y: yPosition) { [weak self] index in
            self?.settings.appearance.fontSize = FontSize.allCases[index]
        }
        contentView.addSubview(fontPopup)
        yPosition += 50
        
        // 界面选项
        let interfaceSection = createSectionLabel("界面选项", y: yPosition)
        contentView.addSubview(interfaceSection)
        yPosition += 40
        
        let menuBarCheckbox = createCheckbox("显示菜单栏", value: settings.appearance.showMenuBar, y: yPosition) { [weak self] value in
            self?.settings.appearance.showMenuBar = value
        }
        contentView.addSubview(menuBarCheckbox)
        yPosition += 30
        
        let statusBarCheckbox = createCheckbox("显示状态栏", value: settings.appearance.showStatusBar, y: yPosition) { [weak self] value in
            self?.settings.appearance.showStatusBar = value
        }
        contentView.addSubview(statusBarCheckbox)
        yPosition += 50
        
        // 窗口透明度
        let opacitySection = createSectionLabel("窗口透明度", y: yPosition)
        contentView.addSubview(opacitySection)
        yPosition += 40
        
        let opacitySlider = createSlider(value: settings.appearance.windowOpacity, y: yPosition) { [weak self] value in
            self?.settings.appearance.windowOpacity = value
        }
        contentView.addSubview(opacitySlider)
    }
    
    private func setupNotificationSettings() {
        var yPosition: CGFloat = 20
        
        // 通知总开关
        let generalSection = createSectionLabel("通知设置", y: yPosition)
        contentView.addSubview(generalSection)
        yPosition += 40
        
        let enableCheckbox = createCheckbox("启用通知", value: settings.notifications.enableNotifications, y: yPosition) { [weak self] value in
            self?.settings.notifications.enableNotifications = value
        }
        contentView.addSubview(enableCheckbox)
        yPosition += 30
        
        let soundCheckbox = createCheckbox("通知声音", value: settings.notifications.soundEnabled, y: yPosition) { [weak self] value in
            self?.settings.notifications.soundEnabled = value
        }
        contentView.addSubview(soundCheckbox)
        yPosition += 30
        
        let badgeCheckbox = createCheckbox("应用图标徽章", value: settings.notifications.badgeEnabled, y: yPosition) { [weak self] value in
            self?.settings.notifications.badgeEnabled = value
        }
        contentView.addSubview(badgeCheckbox)
        yPosition += 50
        
        // 分类通知
        let categorySection = createSectionLabel("分类通知", y: yPosition)
        contentView.addSubview(categorySection)
        yPosition += 40
        
        let petCheckbox = createCheckbox("桌面宠物通知", value: settings.notifications.petNotifications, y: yPosition) { [weak self] value in
            self?.settings.notifications.petNotifications = value
        }
        contentView.addSubview(petCheckbox)
        yPosition += 30
        
        let transferCheckbox = createCheckbox("文件传输通知", value: settings.notifications.fileTransferNotifications, y: yPosition) { [weak self] value in
            self?.settings.notifications.fileTransferNotifications = value
        }
        contentView.addSubview(transferCheckbox)
        yPosition += 30
        
        let chatCheckbox = createCheckbox("聊天消息通知", value: settings.notifications.chatNotifications, y: yPosition) { [weak self] value in
            self?.settings.notifications.chatNotifications = value
        }
        contentView.addSubview(chatCheckbox)
        yPosition += 30
        
        let systemCheckbox = createCheckbox("系统通知", value: settings.notifications.systemNotifications, y: yPosition) { [weak self] value in
            self?.settings.notifications.systemNotifications = value
        }
        contentView.addSubview(systemCheckbox)
    }
    
    private func setupPrivacySettings() {
        var yPosition: CGFloat = 20
        
        // 数据收集
        let dataSection = createSectionLabel("数据收集", y: yPosition)
        contentView.addSubview(dataSection)
        yPosition += 40
        
        let analyticsCheckbox = createCheckbox("发送使用统计", value: settings.privacy.analyticsEnabled, y: yPosition) { [weak self] value in
            self?.settings.privacy.analyticsEnabled = value
        }
        contentView.addSubview(analyticsCheckbox)
        yPosition += 30
        
        let crashCheckbox = createCheckbox("发送崩溃报告", value: settings.privacy.crashReportsEnabled, y: yPosition) { [weak self] value in
            self?.settings.privacy.crashReportsEnabled = value
        }
        contentView.addSubview(crashCheckbox)
        yPosition += 50
        
        // 权限设置
        let permissionSection = createSectionLabel("权限设置", y: yPosition)
        contentView.addSubview(permissionSection)
        yPosition += 40
        
        let locationCheckbox = createCheckbox("位置访问", value: settings.privacy.locationEnabled, y: yPosition) { [weak self] value in
            self?.settings.privacy.locationEnabled = value
        }
        contentView.addSubview(locationCheckbox)
        yPosition += 30
        
        let cameraCheckbox = createCheckbox("摄像头访问", value: settings.privacy.cameraEnabled, y: yPosition) { [weak self] value in
            self?.settings.privacy.cameraEnabled = value
        }
        contentView.addSubview(cameraCheckbox)
        yPosition += 30
        
        let micCheckbox = createCheckbox("麦克风访问", value: settings.privacy.microphoneEnabled, y: yPosition) { [weak self] value in
            self?.settings.privacy.microphoneEnabled = value
        }
        contentView.addSubview(micCheckbox)
        yPosition += 50
        
        // 数据收集级别
        let levelSection = createSectionLabel("数据收集级别", y: yPosition)
        contentView.addSubview(levelSection)
        yPosition += 40
        
        let levelPopup = createPopupButton(DataCollectionLevel.allCases.map { $0.rawValue }, selectedIndex: 1, y: yPosition) { [weak self] index in
            self?.settings.privacy.dataCollection = DataCollectionLevel.allCases[index]
        }
        contentView.addSubview(levelPopup)
    }
    
    private func setupAdvancedSettings() {
        var yPosition: CGFloat = 20
        
        // 调试设置
        let debugSection = createSectionLabel("调试设置", y: yPosition)
        contentView.addSubview(debugSection)
        yPosition += 40
        
        let debugCheckbox = createCheckbox("启用调试模式", value: settings.advanced.debugMode, y: yPosition) { [weak self] value in
            self?.settings.advanced.debugMode = value
        }
        contentView.addSubview(debugCheckbox)
        yPosition += 30
        
        let logPopup = createPopupButton(LogLevel.allCases.map { $0.rawValue }, selectedIndex: 2, y: yPosition) { [weak self] index in
            self?.settings.advanced.logLevel = LogLevel.allCases[index]
        }
        contentView.addSubview(logPopup)
        yPosition += 50
        
        // 性能设置
        let performanceSection = createSectionLabel("性能设置", y: yPosition)
        contentView.addSubview(performanceSection)
        yPosition += 40
        
        let timeoutField = createNumberField(value: Double(settings.advanced.networkTimeout), label: "网络超时 (秒)", y: yPosition) { [weak self] value in
            self?.settings.advanced.networkTimeout = Int(value)
        }
        contentView.addSubview(timeoutField)
        yPosition += 40
        
        let transferField = createNumberField(value: Double(settings.advanced.maxConcurrentTransfers), label: "最大并发传输", y: yPosition) { [weak self] value in
            self?.settings.advanced.maxConcurrentTransfers = Int(value)
        }
        contentView.addSubview(transferField)
        yPosition += 40
        
        let cacheField = createNumberField(value: Double(settings.advanced.cacheSize), label: "缓存大小 (MB)", y: yPosition) { [weak self] value in
            self?.settings.advanced.cacheSize = Int(value)
        }
        contentView.addSubview(cacheField)
    }
    
    // MARK: - UI Helpers
    private func createSectionLabel(_ title: String, y: CGFloat) -> NSTextField {
        let label = NSTextField(labelWithString: title)
        label.font = NSFont.boldSystemFont(ofSize: 14)
        label.frame = NSRect(x: 0, y: y, width: 200, height: 20)
        return label
    }
    
    private func createCheckbox(_ title: String, value: Bool, y: CGFloat, action: @escaping (Bool) -> Void) -> NSButton {
        let checkbox = NSButton(checkboxWithTitle: title, target: nil, action: nil)
        checkbox.state = value ? .on : .off
        checkbox.frame = NSRect(x: 20, y: y, width: 300, height: 20)
        
        checkbox.target = self
        checkbox.action = #selector(checkboxChanged(_:))
        
        // 存储回调
        objc_setAssociatedObject(checkbox, "action", action, .OBJC_ASSOCIATION_RETAIN_NONATOMIC)
        
        return checkbox
    }
    
    private func createPopupButton(_ items: [String], selectedIndex: Int, y: CGFloat, action: @escaping (Int) -> Void) -> NSPopUpButton {
        let popup = NSPopUpButton()
        popup.frame = NSRect(x: 20, y: y, width: 200, height: 25)
        
        for item in items {
            popup.addItem(withTitle: item)
        }
        popup.selectItem(at: selectedIndex)
        
        popup.target = self
        popup.action = #selector(popupChanged(_:))
        
        // 存储回调
        objc_setAssociatedObject(popup, "action", action, .OBJC_ASSOCIATION_RETAIN_NONATOMIC)
        
        return popup
    }
    
    private func createSlider(value: Double, y: CGFloat, action: @escaping (Double) -> Void) -> NSSlider {
        let slider = NSSlider()
        slider.frame = NSRect(x: 20, y: y, width: 200, height: 20)
        slider.minValue = 0.5
        slider.maxValue = 1.0
        slider.doubleValue = value
        
        slider.target = self
        slider.action = #selector(sliderChanged(_:))
        
        // 存储回调
        objc_setAssociatedObject(slider, "action", action, .OBJC_ASSOCIATION_RETAIN_NONATOMIC)
        
        return slider
    }
    
    private func createNumberField(value: Double, label: String, y: CGFloat, action: @escaping (Double) -> Void) -> NSView {
        let container = NSView()
        container.frame = NSRect(x: 0, y: y, width: 300, height: 25)
        
        let labelField = NSTextField(labelWithString: label)
        labelField.frame = NSRect(x: 20, y: 0, width: 150, height: 20)
        container.addSubview(labelField)
        
        let numberField = NSTextField()
        numberField.frame = NSRect(x: 180, y: 0, width: 80, height: 20)
        numberField.doubleValue = value
        numberField.target = self
        numberField.action = #selector(numberFieldChanged(_:))
        container.addSubview(numberField)
        
        // 存储回调
        objc_setAssociatedObject(numberField, "action", action, .OBJC_ASSOCIATION_RETAIN_NONATOMIC)
        
        return container
    }
    
    // MARK: - Actions
    @objc private func selectCategory(_ sender: NSTableView) {
        let selectedRow = sender.selectedRow
        guard selectedRow >= 0 && selectedRow < SettingsCategory.allCases.count else { return }
        
        let category = SettingsCategory.allCases[selectedRow]
        showCategoryContent(category)
    }
    
    @objc private func checkboxChanged(_ sender: NSButton) {
        if let action = objc_getAssociatedObject(sender, "action") as? (Bool) -> Void {
            action(sender.state == .on)
        }
    }
    
    @objc private func popupChanged(_ sender: NSPopUpButton) {
        if let action = objc_getAssociatedObject(sender, "action") as? (Int) -> Void {
            action(sender.indexOfSelectedItem)
        }
    }
    
    @objc private func sliderChanged(_ sender: NSSlider) {
        if let action = objc_getAssociatedObject(sender, "action") as? (Double) -> Void {
            action(sender.doubleValue)
        }
    }
    
    @objc private func numberFieldChanged(_ sender: NSTextField) {
        if let action = objc_getAssociatedObject(sender, "action") as? (Double) -> Void {
            action(sender.doubleValue)
        }
    }
    
    @objc private func resetSettings(_ sender: NSButton) {
        let alert = NSAlert()
        alert.messageText = "重置设置"
        alert.informativeText = "确定要将所有设置重置为默认值吗？此操作无法撤销。"
        alert.addButton(withTitle: "重置")
        alert.addButton(withTitle: "取消")
        alert.alertStyle = .warning
        
        if alert.runModal() == .alertFirstButtonReturn {
            settings = AppSettings.shared
            showCategoryContent(currentCategory)
        }
    }
    
    @objc private func exportSettings(_ sender: NSButton) {
        let savePanel = NSSavePanel()
        savePanel.allowedContentTypes = [.json]
        savePanel.nameFieldStringValue = "settings.json"
        
        if savePanel.runModal() == .OK {
            // 导出设置到JSON文件
            exportSettingsToFile(savePanel.url!)
        }
    }
    
    @objc private func importSettings(_ sender: NSButton) {
        let openPanel = NSOpenPanel()
        openPanel.allowedContentTypes = [.json]
        openPanel.allowsMultipleSelection = false
        
        if openPanel.runModal() == .OK {
            // 从JSON文件导入设置
            importSettingsFromFile(openPanel.url!)
        }
    }
    
    private func exportSettingsToFile(_ url: URL) {
        do {
            let encoder = JSONEncoder()
            encoder.outputFormatting = .prettyPrinted
            let data = try encoder.encode(settings)
            try data.write(to: url)
            
            let alert = NSAlert()
            alert.messageText = "导出成功"
            alert.informativeText = "设置已成功导出到文件。"
            alert.runModal()
        } catch {
            let alert = NSAlert()
            alert.messageText = "导出失败"
            alert.informativeText = "无法导出设置：\(error.localizedDescription)"
            alert.alertStyle = .critical
            alert.runModal()
        }
    }
    
    private func importSettingsFromFile(_ url: URL) {
        do {
            let data = try Data(contentsOf: url)
            let decoder = JSONDecoder()
            settings = try decoder.decode(AppSettings.self, from: data)
            
            showCategoryContent(currentCategory)
            
            let alert = NSAlert()
            alert.messageText = "导入成功"
            alert.informativeText = "设置已成功从文件导入。"
            alert.runModal()
        } catch {
            let alert = NSAlert()
            alert.messageText = "导入失败"
            alert.informativeText = "无法导入设置：\(error.localizedDescription)"
            alert.alertStyle = .critical
            alert.runModal()
        }
    }
}

// MARK: - Settings Category
enum SettingsCategory: String, CaseIterable {
    case general = "通用"
    case appearance = "外观"
    case notifications = "通知"
    case privacy = "隐私"
    case advanced = "高级"
    
    var title: String {
        return self.rawValue
    }
    
    var icon: String {
        switch self {
        case .general: return "⚙️"
        case .appearance: return "🎨"
        case .notifications: return "🔔"
        case .privacy: return "🔒"
        case .advanced: return "🔧"
        }
    }
}

// MARK: - Table View Data Source
extension SettingsViewController: NSTableViewDataSource {
    func numberOfRows(in tableView: NSTableView) -> Int {
        return SettingsCategory.allCases.count
    }
}

// MARK: - Table View Delegate
extension SettingsViewController: NSTableViewDelegate {
    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        guard row < SettingsCategory.allCases.count else { return nil }
        
        let category = SettingsCategory.allCases[row]
        
        let cellView = NSTableCellView()
        let textField = NSTextField()
        textField.isBordered = false
        textField.isEditable = false
        textField.backgroundColor = NSColor.clear
        textField.stringValue = "\(category.icon) \(category.title)"
        textField.font = NSFont.systemFont(ofSize: 13)
        
        cellView.addSubview(textField)
        cellView.textField = textField
        
        textField.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 10),
            textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -10),
            textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
        ])
        
        return cellView
    }
}

// MARK: - Codable Extensions
extension AppSettings: Codable {}
extension GeneralSettings: Codable {}
extension AppearanceSettings: Codable {}
extension NotificationSettings: Codable {}
extension PrivacySettings: Codable {}
extension AdvancedSettings: Codable {}