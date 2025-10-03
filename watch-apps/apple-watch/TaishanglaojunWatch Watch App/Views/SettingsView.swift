import SwiftUI
import WatchKit

/**
 * 设置视图
 * 管理手表端的各种设置和偏好
 */
struct SettingsView: View {
    
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    @EnvironmentObject private var connectivityManager: WatchConnectivityManager
    @EnvironmentObject private var taskManager: WatchTaskManager
    
    @State private var showingResetAlert = false
    @State private var showingExportSheet = false
    @State private var showingAboutSheet = false
    
    var body: some View {
        NavigationView {
            List {
                // 通知设置
                Section("通知") {
                    SettingsToggleRow(
                        title: "启用通知",
                        icon: "bell",
                        isOn: $settingsManager.notificationsEnabled
                    )
                    
                    SettingsToggleRow(
                        title: "任务提醒",
                        icon: "alarm",
                        isOn: $settingsManager.taskRemindersEnabled
                    )
                    
                    SettingsToggleRow(
                        title: "进度通知",
                        icon: "chart.bar",
                        isOn: $settingsManager.progressNotificationsEnabled
                    )
                }
                
                // 交互设置
                Section("交互") {
                    SettingsToggleRow(
                        title: "触觉反馈",
                        icon: "hand.tap",
                        isOn: $settingsManager.hapticFeedbackEnabled
                    )
                    
                    SettingsToggleRow(
                        title: "语音报告",
                        icon: "mic",
                        isOn: $settingsManager.voiceReportsEnabled
                    )
                    
                    NavigationLink(destination: DisplaySettingsView()) {
                        SettingsRow(
                            title: "显示设置",
                            icon: "display",
                            value: settingsManager.taskDisplayMode.displayName
                        )
                    }
                }
                
                // 同步设置
                Section("同步") {
                    SettingsToggleRow(
                        title: "自动同步",
                        icon: "arrow.clockwise",
                        isOn: $settingsManager.autoSyncEnabled
                    )
                    
                    NavigationLink(destination: SyncSettingsView()) {
                        SettingsRow(
                            title: "同步间隔",
                            icon: "timer",
                            value: settingsManager.describeSyncInterval()
                        )
                    }
                    
                    Button(action: forceSyncData) {
                        HStack {
                            Image(systemName: "icloud.and.arrow.down")
                                .foregroundColor(.blue)
                            Text("立即同步")
                                .foregroundColor(.blue)
                            Spacer()
                            if connectivityManager.isSyncing {
                                ProgressView()
                                    .scaleEffect(0.8)
                            }
                        }
                    }
                    .disabled(connectivityManager.isSyncing)
                }
                
                // 快捷操作设置
                Section("快捷操作") {
                    NavigationLink(destination: QuickActionsSettingsView()) {
                        SettingsRow(
                            title: "快捷按钮",
                            icon: "square.grid.2x2",
                            value: "\(settingsManager.enabledQuickActions.count)个已启用"
                        )
                    }
                    
                    NavigationLink(destination: ComplicationSettingsView()) {
                        SettingsRow(
                            title: "复杂功能",
                            icon: "app.badge",
                            value: settingsManager.complicationStyle.displayName
                        )
                    }
                }
                
                // 系统设置
                Section("系统") {
                    NavigationLink(destination: LanguageSettingsView()) {
                        SettingsRow(
                            title: "语言",
                            icon: "globe",
                            value: settingsManager.language.displayName
                        )
                    }
                    
                    NavigationLink(destination: UnitsSettingsView()) {
                        SettingsRow(
                            title: "单位",
                            icon: "ruler",
                            value: "\(settingsManager.timeFormat.displayName), \(settingsManager.distanceUnit.displayName)"
                        )
                    }
                    
                    SettingsSliderRow(
                        title: "屏幕亮度",
                        icon: "sun.max",
                        value: $settingsManager.displayBrightness,
                        range: 0.1...1.0
                    )
                }
                
                // 数据管理
                Section("数据") {
                    Button(action: { showingExportSheet = true }) {
                        SettingsActionRow(
                            title: "导出设置",
                            icon: "square.and.arrow.up",
                            color: .blue
                        )
                    }
                    
                    Button(action: clearCache) {
                        SettingsActionRow(
                            title: "清除缓存",
                            icon: "trash",
                            color: .orange
                        )
                    }
                    
                    Button(action: { showingResetAlert = true }) {
                        SettingsActionRow(
                            title: "重置设置",
                            icon: "arrow.counterclockwise",
                            color: .red
                        )
                    }
                }
                
                // 关于
                Section("关于") {
                    Button(action: { showingAboutSheet = true }) {
                        SettingsRow(
                            title: "应用信息",
                            icon: "info.circle",
                            value: settingsManager.appInfo.version
                        )
                    }
                    
                    if settingsManager.debugModeEnabled {
                        NavigationLink(destination: DebugSettingsView()) {
                            SettingsRow(
                                title: "调试模式",
                                icon: "ladybug",
                                value: "已启用"
                            )
                        }
                    }
                }
            }
            .navigationTitle("设置")
            .navigationBarTitleDisplayMode(.inline)
        }
        .alert("重置设置", isPresented: $showingResetAlert) {
            Button("取消", role: .cancel) { }
            Button("重置", role: .destructive) {
                resetAllSettings()
            }
        } message: {
            Text("确定要重置所有设置吗？此操作无法撤销。")
        }
        .sheet(isPresented: $showingExportSheet) {
            ExportSettingsView()
        }
        .sheet(isPresented: $showingAboutSheet) {
            AboutView()
        }
    }
    
    // MARK: - Private Methods
    
    private func forceSyncData() {
        Task {
            await connectivityManager.forceSync()
            settingsManager.triggerHapticFeedback(.success)
        }
    }
    
    private func clearCache() {
        Task {
            await taskManager.clearCache()
            settingsManager.triggerHapticFeedback(.success)
        }
    }
    
    private func resetAllSettings() {
        settingsManager.resetToDefaults()
        settingsManager.triggerHapticFeedback(.success)
    }
}

// MARK: - Settings Components

/**
 * 设置行视图
 */
struct SettingsRow: View {
    let title: String
    let icon: String
    let value: String?
    
    init(title: String, icon: String, value: String? = nil) {
        self.title = title
        self.icon = icon
        self.value = value
    }
    
    var body: some View {
        HStack {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 20)
            
            Text(title)
                .font(.caption)
            
            Spacer()
            
            if let value = value {
                Text(value)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            
            Image(systemName: "chevron.right")
                .font(.caption2)
                .foregroundColor(.secondary)
        }
    }
}

/**
 * 设置开关行视图
 */
struct SettingsToggleRow: View {
    let title: String
    let icon: String
    @Binding var isOn: Bool
    
    var body: some View {
        HStack {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 20)
            
            Text(title)
                .font(.caption)
            
            Spacer()
            
            Toggle("", isOn: $isOn)
                .labelsHidden()
        }
    }
}

/**
 * 设置滑块行视图
 */
struct SettingsSliderRow: View {
    let title: String
    let icon: String
    @Binding var value: Double
    let range: ClosedRange<Double>
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: icon)
                    .foregroundColor(.blue)
                    .frame(width: 20)
                
                Text(title)
                    .font(.caption)
                
                Spacer()
                
                Text("\(Int(value * 100))%")
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            
            Slider(value: $value, in: range)
                .accentColor(.blue)
        }
    }
}

/**
 * 设置操作行视图
 */
struct SettingsActionRow: View {
    let title: String
    let icon: String
    let color: Color
    
    var body: some View {
        HStack {
            Image(systemName: icon)
                .foregroundColor(color)
                .frame(width: 20)
            
            Text(title)
                .font(.caption)
                .foregroundColor(color)
            
            Spacer()
        }
    }
}

// MARK: - Sub Settings Views

/**
 * 显示设置视图
 */
struct DisplaySettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    
    var body: some View {
        List {
            Section("任务显示") {
                ForEach(TaskDisplayMode.allCases, id: \.self) { mode in
                    Button(action: { settingsManager.taskDisplayMode = mode }) {
                        HStack {
                            Text(mode.displayName)
                                .font(.caption)
                                .foregroundColor(.primary)
                            
                            Spacer()
                            
                            if settingsManager.taskDisplayMode == mode {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
            
            Section("亮度") {
                SettingsSliderRow(
                    title: "屏幕亮度",
                    icon: "sun.max",
                    value: $settingsManager.displayBrightness,
                    range: 0.1...1.0
                )
            }
        }
        .navigationTitle("显示设置")
        .navigationBarTitleDisplayMode(.inline)
    }
}

/**
 * 同步设置视图
 */
struct SyncSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    
    private let syncIntervals: [TimeInterval] = [30, 60, 300, 600, 1800, 3600]
    
    var body: some View {
        List {
            Section("自动同步间隔") {
                ForEach(syncIntervals, id: \.self) { interval in
                    Button(action: { settingsManager.autoSyncInterval = interval }) {
                        HStack {
                            Text(describeSyncInterval(interval))
                                .font(.caption)
                                .foregroundColor(.primary)
                            
                            Spacer()
                            
                            if settingsManager.autoSyncInterval == interval {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("同步设置")
        .navigationBarTitleDisplayMode(.inline)
    }
    
    private func describeSyncInterval(_ interval: TimeInterval) -> String {
        if interval < 60 {
            return "\(Int(interval))秒"
        } else if interval < 3600 {
            return "\(Int(interval / 60))分钟"
        } else {
            return "\(Int(interval / 3600))小时"
        }
    }
}

/**
 * 快捷操作设置视图
 */
struct QuickActionsSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    
    var body: some View {
        List {
            Section("可用快捷操作") {
                ForEach(QuickActionType.allCases, id: \.self) { action in
                    SettingsToggleRow(
                        title: action.displayName,
                        icon: action.iconName,
                        isOn: Binding(
                            get: { settingsManager.enabledQuickActions.contains(action) },
                            set: { enabled in
                                if enabled {
                                    settingsManager.enabledQuickActions.insert(action)
                                } else {
                                    settingsManager.enabledQuickActions.remove(action)
                                }
                            }
                        )
                    )
                }
            }
        }
        .navigationTitle("快捷操作")
        .navigationBarTitleDisplayMode(.inline)
    }
}

/**
 * 复杂功能设置视图
 */
struct ComplicationSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    
    var body: some View {
        List {
            Section("复杂功能样式") {
                ForEach(ComplicationStyle.allCases, id: \.self) { style in
                    Button(action: { settingsManager.complicationStyle = style }) {
                        HStack {
                            Text(style.displayName)
                                .font(.caption)
                                .foregroundColor(.primary)
                            
                            Spacer()
                            
                            if settingsManager.complicationStyle == style {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("复杂功能")
        .navigationBarTitleDisplayMode(.inline)
    }
}

/**
 * 语言设置视图
 */
struct LanguageSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    
    var body: some View {
        List {
            Section("应用语言") {
                ForEach(AppLanguage.allCases, id: \.self) { language in
                    Button(action: { settingsManager.language = language }) {
                        HStack {
                            Text(language.displayName)
                                .font(.caption)
                                .foregroundColor(.primary)
                            
                            Spacer()
                            
                            if settingsManager.language == language {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("语言")
        .navigationBarTitleDisplayMode(.inline)
    }
}

/**
 * 单位设置视图
 */
struct UnitsSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    
    var body: some View {
        List {
            Section("时间格式") {
                ForEach(TimeFormat.allCases, id: \.self) { format in
                    Button(action: { settingsManager.timeFormat = format }) {
                        HStack {
                            Text(format.displayName)
                                .font(.caption)
                                .foregroundColor(.primary)
                            
                            Spacer()
                            
                            if settingsManager.timeFormat == format {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
            
            Section("距离单位") {
                ForEach(DistanceUnit.allCases, id: \.self) { unit in
                    Button(action: { settingsManager.distanceUnit = unit }) {
                        HStack {
                            Text(unit.displayName)
                                .font(.caption)
                                .foregroundColor(.primary)
                            
                            Spacer()
                            
                            if settingsManager.distanceUnit == unit {
                                Image(systemName: "checkmark")
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("单位")
        .navigationBarTitleDisplayMode(.inline)
    }
}

/**
 * 调试设置视图
 */
struct DebugSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    @EnvironmentObject private var connectivityManager: WatchConnectivityManager
    @EnvironmentObject private var taskManager: WatchTaskManager
    
    var body: some View {
        List {
            Section("连接状态") {
                HStack {
                    Text("连接状态")
                        .font(.caption)
                    Spacer()
                    Text(connectivityManager.connectionStatus.rawValue)
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
                
                HStack {
                    Text("最后同步")
                        .font(.caption)
                    Spacer()
                    Text(formatLastSync())
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
            }
            
            Section("任务统计") {
                HStack {
                    Text("缓存任务数")
                        .font(.caption)
                    Spacer()
                    Text("\(taskManager.tasks.count)")
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
                
                HStack {
                    Text("活跃任务数")
                        .font(.caption)
                    Spacer()
                    Text("\(taskManager.activeTasks.count)")
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
            }
            
            Section("应用信息") {
                HStack {
                    Text("版本")
                        .font(.caption)
                    Spacer()
                    Text(settingsManager.appInfo.version)
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
                
                HStack {
                    Text("构建号")
                        .font(.caption)
                    Spacer()
                    Text(settingsManager.appInfo.buildNumber)
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
            }
        }
        .navigationTitle("调试信息")
        .navigationBarTitleDisplayMode(.inline)
    }
    
    private func formatLastSync() -> String {
        guard let lastSync = connectivityManager.lastSyncTime else {
            return "从未同步"
        }
        
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: lastSync, relativeTo: Date())
    }
}

/**
 * 导出设置视图
 */
struct ExportSettingsView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    @Environment(\.dismiss) private var dismiss
    
    @State private var isExporting = false
    @State private var exportSuccess = false
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                Image(systemName: "square.and.arrow.up")
                    .font(.largeTitle)
                    .foregroundColor(.blue)
                
                Text("导出设置")
                    .font(.headline)
                
                Text("将当前设置导出为配置文件，可在其他设备上导入使用。")
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                
                if exportSuccess {
                    Text("设置已成功导出")
                        .font(.caption)
                        .foregroundColor(.green)
                }
                
                Button(action: exportSettings) {
                    HStack {
                        if isExporting {
                            ProgressView()
                                .scaleEffect(0.8)
                        } else {
                            Image(systemName: "square.and.arrow.up")
                        }
                        Text("导出设置")
                    }
                    .font(.caption)
                    .foregroundColor(.white)
                    .padding()
                    .background(Color.blue)
                    .cornerRadius(8)
                }
                .disabled(isExporting)
                
                Spacer()
            }
            .padding()
            .navigationTitle("导出设置")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("关闭") {
                        dismiss()
                    }
                    .font(.caption2)
                }
            }
        }
    }
    
    private func exportSettings() {
        isExporting = true
        
        // 模拟导出过程
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
            settingsManager.exportSettings()
            isExporting = false
            exportSuccess = true
            
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
                dismiss()
            }
        }
    }
}

/**
 * 关于视图
 */
struct AboutView: View {
    @EnvironmentObject private var settingsManager: WatchSettingsManager
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 20) {
                    // 应用图标和名称
                    Image(systemName: "applewatch")
                        .font(.system(size: 60))
                        .foregroundColor(.blue)
                    
                    Text("太上老君手表端")
                        .font(.headline)
                        .fontWeight(.bold)
                    
                    Text("版本 \(settingsManager.appInfo.version)")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    
                    Divider()
                    
                    // 应用信息
                    VStack(alignment: .leading, spacing: 12) {
                        InfoRow(title: "开发者", value: "太上老君团队")
                        InfoRow(title: "版本", value: settingsManager.appInfo.version)
                        InfoRow(title: "构建号", value: settingsManager.appInfo.buildNumber)
                        InfoRow(title: "发布日期", value: settingsManager.appInfo.releaseDate)
                    }
                    
                    Divider()
                    
                    // 功能介绍
                    VStack(alignment: .leading, spacing: 8) {
                        Text("主要功能")
                            .font(.caption)
                            .fontWeight(.semibold)
                        
                        Text("• 任务管理和跟踪")
                        Text("• 实时数据同步")
                        Text("• 语音交互支持")
                        Text("• 智能通知提醒")
                        Text("• 快捷操作面板")
                    }
                    .font(.caption2)
                    .foregroundColor(.secondary)
                    
                    Spacer()
                }
                .padding()
            }
            .navigationTitle("关于")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("关闭") {
                        dismiss()
                    }
                    .font(.caption2)
                }
            }
        }
    }
}

/**
 * 信息行视图
 */
struct InfoRow: View {
    let title: String
    let value: String
    
    var body: some View {
        HStack {
            Text(title)
                .font(.caption)
                .foregroundColor(.secondary)
            
            Spacer()
            
            Text(value)
                .font(.caption)
                .foregroundColor(.primary)
        }
    }
}

#Preview {
    SettingsView()
        .environmentObject(WatchSettingsManager())
        .environmentObject(WatchConnectivityManager())
        .environmentObject(WatchTaskManager())
}