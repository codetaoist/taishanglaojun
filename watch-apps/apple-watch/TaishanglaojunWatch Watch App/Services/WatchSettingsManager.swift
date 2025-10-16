import Foundation
import WatchKit
import Combine

/**
 * 手表设置管理器
 * 负责管理手表应用的各种设置和偏好
 */
@MainActor
class WatchSettingsManager: ObservableObject {
    
    // MARK: - Published Properties
    
    @Published var isNotificationsEnabled = true
    @Published var isHapticFeedbackEnabled = true
    @Published var isVoiceReportEnabled = true
    @Published var isAutoSyncEnabled = true
    @Published var syncInterval: TimeInterval = 300 // 5分钟
    @Published var displayBrightness: Double = 0.8
    @Published var taskDisplayMode: TaskDisplayMode = .list
    @Published var quickActionButtons: [QuickActionType] = [.acceptTask, .syncData, .voiceReport]
    @Published var complicationStyle: ComplicationStyle = .circular
    @Published var language: AppLanguage = .chinese
    @Published var timeFormat: TimeFormat = .twentyFourHour
    @Published var distanceUnit: DistanceUnit = .metric
    @Published var isDebugModeEnabled = false
    
    // MARK: - Private Properties
    
    private let userDefaults = UserDefaults.standard
    private var cancellables = Set<AnyCancellable>()
    
    // MARK: - Settings Keys
    
    private enum SettingsKeys {
        static let notificationsEnabled = "notifications_enabled"
        static let hapticFeedbackEnabled = "haptic_feedback_enabled"
        static let voiceReportEnabled = "voice_report_enabled"
        static let autoSyncEnabled = "auto_sync_enabled"
        static let syncInterval = "sync_interval"
        static let displayBrightness = "display_brightness"
        static let taskDisplayMode = "task_display_mode"
        static let quickActionButtons = "quick_action_buttons"
        static let complicationStyle = "complication_style"
        static let language = "app_language"
        static let timeFormat = "time_format"
        static let distanceUnit = "distance_unit"
        static let debugModeEnabled = "debug_mode_enabled"
        static let firstLaunch = "first_launch"
        static let lastSyncTime = "last_sync_time"
        static let totalTasksCompleted = "total_tasks_completed"
        static let appVersion = "app_version"
    }
    
    // MARK: - Initialization
    
    init() {
        loadSettings()
        setupSettingsObservers()
        
        // 检查是否首次启动
        if isFirstLaunch {
            setupDefaultSettings()
        }
    }
    
    // MARK: - Public Methods
    
    /**
     * 重置所有设置为默认值
     */
    func resetToDefaults() {
        isNotificationsEnabled = true
        isHapticFeedbackEnabled = true
        isVoiceReportEnabled = true
        isAutoSyncEnabled = true
        syncInterval = 300
        displayBrightness = 0.8
        taskDisplayMode = .list
        quickActionButtons = [.acceptTask, .syncData, .voiceReport]
        complicationStyle = .circular
        language = .chinese
        timeFormat = .twentyFourHour
        distanceUnit = .metric
        isDebugModeEnabled = false
        
        saveAllSettings()
        
        // 触觉反馈
        if isHapticFeedbackEnabled {
            WKInterfaceDevice.current().play(.success)
        }
    }
    
    /**
     * 导出设置
     */
    func exportSettings() -> [String: Any] {
        return [
            SettingsKeys.notificationsEnabled: isNotificationsEnabled,
            SettingsKeys.hapticFeedbackEnabled: isHapticFeedbackEnabled,
            SettingsKeys.voiceReportEnabled: isVoiceReportEnabled,
            SettingsKeys.autoSyncEnabled: isAutoSyncEnabled,
            SettingsKeys.syncInterval: syncInterval,
            SettingsKeys.displayBrightness: displayBrightness,
            SettingsKeys.taskDisplayMode: taskDisplayMode.rawValue,
            SettingsKeys.quickActionButtons: quickActionButtons.map { $0.rawValue },
            SettingsKeys.complicationStyle: complicationStyle.rawValue,
            SettingsKeys.language: language.rawValue,
            SettingsKeys.timeFormat: timeFormat.rawValue,
            SettingsKeys.distanceUnit: distanceUnit.rawValue,
            SettingsKeys.debugModeEnabled: isDebugModeEnabled
        ]
    }
    
    /**
     * 导入设置
     */
    func importSettings(_ settings: [String: Any]) {
        if let value = settings[SettingsKeys.notificationsEnabled] as? Bool {
            isNotificationsEnabled = value
        }
        if let value = settings[SettingsKeys.hapticFeedbackEnabled] as? Bool {
            isHapticFeedbackEnabled = value
        }
        if let value = settings[SettingsKeys.voiceReportEnabled] as? Bool {
            isVoiceReportEnabled = value
        }
        if let value = settings[SettingsKeys.autoSyncEnabled] as? Bool {
            isAutoSyncEnabled = value
        }
        if let value = settings[SettingsKeys.syncInterval] as? TimeInterval {
            syncInterval = value
        }
        if let value = settings[SettingsKeys.displayBrightness] as? Double {
            displayBrightness = value
        }
        if let value = settings[SettingsKeys.taskDisplayMode] as? String {
            taskDisplayMode = TaskDisplayMode(rawValue: value) ?? .list
        }
        if let values = settings[SettingsKeys.quickActionButtons] as? [String] {
            quickActionButtons = values.compactMap { QuickActionType(rawValue: $0) }
        }
        if let value = settings[SettingsKeys.complicationStyle] as? String {
            complicationStyle = ComplicationStyle(rawValue: value) ?? .circular
        }
        if let value = settings[SettingsKeys.language] as? String {
            language = AppLanguage(rawValue: value) ?? .chinese
        }
        if let value = settings[SettingsKeys.timeFormat] as? String {
            timeFormat = TimeFormat(rawValue: value) ?? .twentyFourHour
        }
        if let value = settings[SettingsKeys.distanceUnit] as? String {
            distanceUnit = DistanceUnit(rawValue: value) ?? .metric
        }
        if let value = settings[SettingsKeys.debugModeEnabled] as? Bool {
            isDebugModeEnabled = value
        }
        
        saveAllSettings()
    }
    
    /**
     * 获取应用信息
     */
    func getAppInfo() -> AppInfo {
        let bundle = Bundle.main
        let version = bundle.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
        let build = bundle.infoDictionary?["CFBundleVersion"] as? String ?? "1"
        
        return AppInfo(
            version: version,
            build: build,
            lastSyncTime: userDefaults.object(forKey: SettingsKeys.lastSyncTime) as? Date,
            totalTasksCompleted: userDefaults.integer(forKey: SettingsKeys.totalTasksCompleted),
            isFirstLaunch: isFirstLaunch
        )
    }
    
    /**
     * 更新最后同步时间
     */
    func updateLastSyncTime(_ date: Date = Date()) {
        userDefaults.set(date, forKey: SettingsKeys.lastSyncTime)
    }
    
    /**
     * 增加完成任务计数
     */
    func incrementCompletedTasks() {
        let current = userDefaults.integer(forKey: SettingsKeys.totalTasksCompleted)
        userDefaults.set(current + 1, forKey: SettingsKeys.totalTasksCompleted)
    }
    
    /**
     * 触发触觉反馈
     */
    func triggerHapticFeedback(_ type: WKHapticType) {
        if isHapticFeedbackEnabled {
            WKInterfaceDevice.current().play(type)
        }
    }
    
    /**
     * 检查功能是否可用
     */
    func isFeatureEnabled(_ feature: AppFeature) -> Bool {
        switch feature {
        case .notifications:
            return isNotificationsEnabled
        case .hapticFeedback:
            return isHapticFeedbackEnabled
        case .voiceReport:
            return isVoiceReportEnabled
        case .autoSync:
            return isAutoSyncEnabled
        case .debugMode:
            return isDebugModeEnabled
        }
    }
    
    /**
     * 获取同步间隔描述
     */
    func getSyncIntervalDescription() -> String {
        let minutes = Int(syncInterval / 60)
        switch minutes {
        case 1:
            return "1分钟"
        case 5:
            return "5分钟"
        case 10:
            return "10分钟"
        case 30:
            return "30分钟"
        case 60:
            return "1小时"
        default:
            return "\(minutes)分钟"
        }
    }
    
    // MARK: - Private Methods
    
    private var isFirstLaunch: Bool {
        return !userDefaults.bool(forKey: SettingsKeys.firstLaunch)
    }
    
    private func setupDefaultSettings() {
        // 标记已完成首次启动
        userDefaults.set(true, forKey: SettingsKeys.firstLaunch)
        
        // 设置默认值
        saveAllSettings()
        
        // 记录应用版本
        let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
        userDefaults.set(version, forKey: SettingsKeys.appVersion)
    }
    
    private func loadSettings() {
        isNotificationsEnabled = userDefaults.object(forKey: SettingsKeys.notificationsEnabled) as? Bool ?? true
        isHapticFeedbackEnabled = userDefaults.object(forKey: SettingsKeys.hapticFeedbackEnabled) as? Bool ?? true
        isVoiceReportEnabled = userDefaults.object(forKey: SettingsKeys.voiceReportEnabled) as? Bool ?? true
        isAutoSyncEnabled = userDefaults.object(forKey: SettingsKeys.autoSyncEnabled) as? Bool ?? true
        syncInterval = userDefaults.object(forKey: SettingsKeys.syncInterval) as? TimeInterval ?? 300
        displayBrightness = userDefaults.object(forKey: SettingsKeys.displayBrightness) as? Double ?? 0.8
        
        if let modeString = userDefaults.string(forKey: SettingsKeys.taskDisplayMode) {
            taskDisplayMode = TaskDisplayMode(rawValue: modeString) ?? .list
        }
        
        if let buttonStrings = userDefaults.array(forKey: SettingsKeys.quickActionButtons) as? [String] {
            quickActionButtons = buttonStrings.compactMap { QuickActionType(rawValue: $0) }
        }
        
        if let styleString = userDefaults.string(forKey: SettingsKeys.complicationStyle) {
            complicationStyle = ComplicationStyle(rawValue: styleString) ?? .circular
        }
        
        if let languageString = userDefaults.string(forKey: SettingsKeys.language) {
            language = AppLanguage(rawValue: languageString) ?? .chinese
        }
        
        if let formatString = userDefaults.string(forKey: SettingsKeys.timeFormat) {
            timeFormat = TimeFormat(rawValue: formatString) ?? .twentyFourHour
        }
        
        if let unitString = userDefaults.string(forKey: SettingsKeys.distanceUnit) {
            distanceUnit = DistanceUnit(rawValue: unitString) ?? .metric
        }
        
        isDebugModeEnabled = userDefaults.bool(forKey: SettingsKeys.debugModeEnabled)
    }
    
    private func saveAllSettings() {
        userDefaults.set(isNotificationsEnabled, forKey: SettingsKeys.notificationsEnabled)
        userDefaults.set(isHapticFeedbackEnabled, forKey: SettingsKeys.hapticFeedbackEnabled)
        userDefaults.set(isVoiceReportEnabled, forKey: SettingsKeys.voiceReportEnabled)
        userDefaults.set(isAutoSyncEnabled, forKey: SettingsKeys.autoSyncEnabled)
        userDefaults.set(syncInterval, forKey: SettingsKeys.syncInterval)
        userDefaults.set(displayBrightness, forKey: SettingsKeys.displayBrightness)
        userDefaults.set(taskDisplayMode.rawValue, forKey: SettingsKeys.taskDisplayMode)
        userDefaults.set(quickActionButtons.map { $0.rawValue }, forKey: SettingsKeys.quickActionButtons)
        userDefaults.set(complicationStyle.rawValue, forKey: SettingsKeys.complicationStyle)
        userDefaults.set(language.rawValue, forKey: SettingsKeys.language)
        userDefaults.set(timeFormat.rawValue, forKey: SettingsKeys.timeFormat)
        userDefaults.set(distanceUnit.rawValue, forKey: SettingsKeys.distanceUnit)
        userDefaults.set(isDebugModeEnabled, forKey: SettingsKeys.debugModeEnabled)
    }
    
    private func setupSettingsObservers() {
        // 监听设置变化并自动保存
        $isNotificationsEnabled
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.notificationsEnabled)
            }
            .store(in: &cancellables)
        
        $isHapticFeedbackEnabled
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.hapticFeedbackEnabled)
            }
            .store(in: &cancellables)
        
        $isVoiceReportEnabled
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.voiceReportEnabled)
            }
            .store(in: &cancellables)
        
        $isAutoSyncEnabled
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.autoSyncEnabled)
            }
            .store(in: &cancellables)
        
        $syncInterval
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.syncInterval)
            }
            .store(in: &cancellables)
        
        $displayBrightness
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.displayBrightness)
            }
            .store(in: &cancellables)
        
        $taskDisplayMode
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value.rawValue, forKey: SettingsKeys.taskDisplayMode)
            }
            .store(in: &cancellables)
        
        $quickActionButtons
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value.map { $0.rawValue }, forKey: SettingsKeys.quickActionButtons)
            }
            .store(in: &cancellables)
        
        $complicationStyle
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value.rawValue, forKey: SettingsKeys.complicationStyle)
            }
            .store(in: &cancellables)
        
        $language
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value.rawValue, forKey: SettingsKeys.language)
            }
            .store(in: &cancellables)
        
        $timeFormat
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value.rawValue, forKey: SettingsKeys.timeFormat)
            }
            .store(in: &cancellables)
        
        $distanceUnit
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value.rawValue, forKey: SettingsKeys.distanceUnit)
            }
            .store(in: &cancellables)
        
        $isDebugModeEnabled
            .dropFirst()
            .sink { [weak self] value in
                self?.userDefaults.set(value, forKey: SettingsKeys.debugModeEnabled)
            }
            .store(in: &cancellables)
    }
}

// MARK: - Supporting Types

/**
 * 任务显示模式
 */
enum TaskDisplayMode: String, CaseIterable {
    case list = "list"
    case grid = "grid"
    case compact = "compact"
    
    var displayName: String {
        switch self {
        case .list: return "列表"
        case .grid: return "网格"
        case .compact: return "紧凑"
        }
    }
}

/**
 * 快速操作类型
 */
enum QuickActionType: String, CaseIterable {
    case acceptTask = "accept_task"
    case completeTask = "complete_task"
    case syncData = "sync_data"
    case voiceReport = "voice_report"
    case showStats = "show_stats"
    case settings = "settings"
    
    var displayName: String {
        switch self {
        case .acceptTask: return "接受任务"
        case .completeTask: return "完成任务"
        case .syncData: return "同步数据"
        case .voiceReport: return "语音报告"
        case .showStats: return "显示统计"
        case .settings: return "设置"
        }
    }
    
    var systemImage: String {
        switch self {
        case .acceptTask: return "checkmark.circle"
        case .completeTask: return "checkmark.circle.fill"
        case .syncData: return "arrow.clockwise"
        case .voiceReport: return "mic.circle"
        case .showStats: return "chart.bar"
        case .settings: return "gear"
        }
    }
}

/**
 * 复杂功能样式
 */
enum ComplicationStyle: String, CaseIterable {
    case circular = "circular"
    case rectangular = "rectangular"
    case corner = "corner"
    case inline = "inline"
    
    var displayName: String {
        switch self {
        case .circular: return "圆形"
        case .rectangular: return "矩形"
        case .corner: return "角落"
        case .inline: return "内联"
        }
    }
}

/**
 * 应用语言
 */
enum AppLanguage: String, CaseIterable {
    case chinese = "zh-CN"
    case english = "en-US"
    
    var displayName: String {
        switch self {
        case .chinese: return "中文"
        case .english: return "English"
        }
    }
}

/**
 * 时间格式
 */
enum TimeFormat: String, CaseIterable {
    case twelveHour = "12h"
    case twentyFourHour = "24h"
    
    var displayName: String {
        switch self {
        case .twelveHour: return "12小时制"
        case .twentyFourHour: return "24小时制"
        }
    }
}

/**
 * 距离单位
 */
enum DistanceUnit: String, CaseIterable {
    case metric = "metric"
    case imperial = "imperial"
    
    var displayName: String {
        switch self {
        case .metric: return "公制"
        case .imperial: return "英制"
        }
    }
}

/**
 * 应用功能
 */
enum AppFeature {
    case notifications
    case hapticFeedback
    case voiceReport
    case autoSync
    case debugMode
}

/**
 * 应用信息
 */
struct AppInfo {
    let version: String
    let build: String
    let lastSyncTime: Date?
    let totalTasksCompleted: Int
    let isFirstLaunch: Bool
}