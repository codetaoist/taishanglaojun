import SwiftUI
import WatchKit

struct QuickActionsView: View {
    @EnvironmentObject var taskManager: WatchTaskManager
    @EnvironmentObject var connectivityManager: WatchConnectivityManager
    @State private var showingVoiceReport = false
    @State private var showingTaskAccept = false
    @State private var selectedQuickTask: WatchTask?
    @State private var isPerformingAction = false
    @State private var actionFeedback: String?
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 16) {
                    // 连接状态卡片
                    ConnectionStatusCard()
                    
                    // 快速任务操作
                    QuickTaskSection()
                    
                    // 主要操作按钮
                    MainActionsSection()
                    
                    // 语音报告
                    VoiceReportSection()
                    
                    // 数据同步
                    SyncSection()
                }
                .padding(.horizontal, 8)
            }
            .navigationTitle("快速操作")
            .navigationBarTitleDisplayMode(.inline)
            .sheet(isPresented: $showingVoiceReport) {
                VoiceReportView()
            }
            .sheet(isPresented: $showingTaskAccept) {
                if let task = selectedQuickTask {
                    QuickTaskAcceptView(task: task)
                }
            }
            .alert("操作反馈", isPresented: .constant(actionFeedback != nil)) {
                Button("确定") {
                    actionFeedback = nil
                }
            } message: {
                if let feedback = actionFeedback {
                    Text(feedback)
                }
            }
        }
    }
}

// MARK: - 连接状态卡片
struct ConnectionStatusCard: View {
    @EnvironmentObject var connectivityManager: WatchConnectivityManager
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: connectionIcon)
                    .foregroundColor(connectionColor)
                
                Text("设备连接")
                    .font(.headline)
                    .fontWeight(.medium)
                
                Spacer()
                
                Text(connectivityManager.connectionStatus.displayName)
                    .font(.caption)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(connectionColor.opacity(0.2))
                    )
                    .foregroundColor(connectionColor)
            }
            
            if let lastSync = connectivityManager.lastSyncTime {
                Text("上次同步: \(formatTime(lastSync))")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            if !connectivityManager.isConnected {
                Text("请确保iPhone在附近并打开蓝牙")
                    .font(.caption)
                    .foregroundColor(.orange)
            }
        }
        .padding()
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color(.systemGray6))
        )
    }
    
    private var connectionIcon: String {
        switch connectivityManager.connectionStatus {
        case .connected:
            return "checkmark.circle.fill"
        case .paired:
            return "exclamationmark.circle.fill"
        case .disconnected:
            return "xmark.circle.fill"
        }
    }
    
    private var connectionColor: Color {
        switch connectivityManager.connectionStatus {
        case .connected:
            return .green
        case .paired:
            return .orange
        case .disconnected:
            return .red
        }
    }
    
    private func formatTime(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

// MARK: - 快速任务区域
struct QuickTaskSection: View {
    @EnvironmentObject var taskManager: WatchTaskManager
    @State private var selectedQuickTask: WatchTask?
    @State private var showingTaskAccept = false
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "bolt.fill")
                    .foregroundColor(.yellow)
                
                Text("快速任务")
                    .font(.headline)
                    .fontWeight(.medium)
            }
            
            let quickTasks = taskManager.getQuickActionTasks()
            
            if quickTasks.isEmpty {
                EmptyQuickTasksView()
            } else {
                LazyVGrid(columns: [
                    GridItem(.flexible()),
                    GridItem(.flexible())
                ], spacing: 8) {
                    ForEach(quickTasks.prefix(4), id: \.id) { task in
                        QuickTaskCard(task: task) {
                            selectedQuickTask = task
                            showingTaskAccept = true
                        }
                    }
                }
            }
        }
        .sheet(isPresented: $showingTaskAccept) {
            if let task = selectedQuickTask {
                QuickTaskAcceptView(task: task)
            }
        }
    }
}

struct QuickTaskCard: View {
    let task: WatchTask
    let onTap: () -> Void
    
    var body: some View {
        Button(action: onTap) {
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text(task.title)
                        .font(.caption)
                        .fontWeight(.medium)
                        .lineLimit(2)
                    
                    Spacer()
                    
                    Circle()
                        .fill(statusColor)
                        .frame(width: 6, height: 6)
                }
                
                Text(task.formattedReward)
                    .font(.caption2)
                    .foregroundColor(.green)
                    .fontWeight(.medium)
                
                HStack(spacing: 2) {
                    ForEach(0..<task.difficulty, id: \.self) { _ in
                        Image(systemName: "star.fill")
                            .font(.caption2)
                            .foregroundColor(.yellow)
                    }
                }
            }
            .padding(8)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color(.systemGray6))
            )
        }
        .buttonStyle(PlainButtonStyle())
    }
    
    private var statusColor: Color {
        switch task.status {
        case .available:
            return .blue
        case .accepted:
            return .orange
        case .inProgress:
            return .yellow
        default:
            return .gray
        }
    }
}

struct EmptyQuickTasksView: View {
    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: "clock")
                .font(.title2)
                .foregroundColor(.secondary)
            
            Text("暂无快速任务")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(.systemGray6))
        )
    }
}

// MARK: - 主要操作区域
struct MainActionsSection: View {
    @EnvironmentObject var taskManager: WatchTaskManager
    @State private var isPerformingAction = false
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "hand.tap.fill")
                    .foregroundColor(.blue)
                
                Text("主要操作")
                    .font(.headline)
                    .fontWeight(.medium)
            }
            
            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: 12) {
                ActionButton(
                    title: "刷新任务",
                    icon: "arrow.clockwise",
                    color: .blue,
                    isLoading: taskManager.isLoading
                ) {
                    taskManager.refreshTasks()
                }
                
                ActionButton(
                    title: "强制同步",
                    icon: "icloud.and.arrow.down",
                    color: .green,
                    isLoading: isPerformingAction
                ) {
                    performForceSync()
                }
                
                ActionButton(
                    title: "查看统计",
                    icon: "chart.bar.fill",
                    color: .purple
                ) {
                    // 导航到统计页面
                }
                
                ActionButton(
                    title: "设置",
                    icon: "gear",
                    color: .gray
                ) {
                    // 导航到设置页面
                }
            }
        }
    }
    
    private func performForceSync() {
        isPerformingAction = true
        taskManager.forceSync()
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
            isPerformingAction = false
        }
    }
}

struct ActionButton: View {
    let title: String
    let icon: String
    let color: Color
    var isLoading: Bool = false
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 8) {
                if isLoading {
                    ProgressView()
                        .progressViewStyle(CircularProgressViewStyle(tint: color))
                        .scaleEffect(0.8)
                } else {
                    Image(systemName: icon)
                        .font(.title2)
                        .foregroundColor(color)
                }
                
                Text(title)
                    .font(.caption)
                    .fontWeight(.medium)
                    .multilineTextAlignment(.center)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 60)
            .background(
                RoundedRectangle(cornerRadius: 12)
                    .fill(Color(.systemGray6))
            )
        }
        .buttonStyle(PlainButtonStyle())
        .disabled(isLoading)
    }
}

// MARK: - 语音报告区域
struct VoiceReportSection: View {
    @State private var showingVoiceReport = false
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "mic.fill")
                    .foregroundColor(.red)
                
                Text("语音报告")
                    .font(.headline)
                    .fontWeight(.medium)
            }
            
            Button(action: {
                showingVoiceReport = true
            }) {
                HStack {
                    Image(systemName: "waveform")
                        .font(.title2)
                        .foregroundColor(.red)
                    
                    VStack(alignment: .leading, spacing: 2) {
                        Text("开始语音报告")
                            .font(.headline)
                            .fontWeight(.medium)
                        
                        Text("快速记录任务进度和问题")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                    
                    Spacer()
                    
                    Image(systemName: "chevron.right")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
                .padding()
                .background(
                    RoundedRectangle(cornerRadius: 12)
                        .fill(Color(.systemGray6))
                )
            }
            .buttonStyle(PlainButtonStyle())
        }
        .sheet(isPresented: $showingVoiceReport) {
            VoiceReportView()
        }
    }
}

// MARK: - 数据同步区域
struct SyncSection: View {
    @EnvironmentObject var connectivityManager: WatchConnectivityManager
    @EnvironmentObject var taskManager: WatchTaskManager
    @State private var isSyncing = false
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "arrow.triangle.2.circlepath")
                    .foregroundColor(.orange)
                
                Text("数据同步")
                    .font(.headline)
                    .fontWeight(.medium)
            }
            
            VStack(spacing: 8) {
                SyncStatusRow(
                    title: "任务数据",
                    status: taskManager.lastUpdateTime != nil ? "已同步" : "未同步",
                    isSuccess: taskManager.lastUpdateTime != nil
                )
                
                SyncStatusRow(
                    title: "设备连接",
                    status: connectivityManager.connectionStatus.displayName,
                    isSuccess: connectivityManager.isConnected
                )
                
                Button(action: performFullSync) {
                    HStack {
                        if isSyncing {
                            ProgressView()
                                .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                .scaleEffect(0.8)
                        } else {
                            Image(systemName: "arrow.triangle.2.circlepath")
                        }
                        
                        Text(isSyncing ? "同步中..." : "立即同步")
                            .fontWeight(.medium)
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(
                        RoundedRectangle(cornerRadius: 12)
                            .fill(Color.accentColor)
                    )
                    .foregroundColor(.white)
                }
                .buttonStyle(PlainButtonStyle())
                .disabled(isSyncing || !connectivityManager.isConnected)
            }
        }
    }
    
    private func performFullSync() {
        isSyncing = true
        
        taskManager.forceSync()
        
        DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
            isSyncing = false
        }
    }
}

struct SyncStatusRow: View {
    let title: String
    let status: String
    let isSuccess: Bool
    
    var body: some View {
        HStack {
            Text(title)
                .font(.caption)
                .foregroundColor(.primary)
            
            Spacer()
            
            HStack(spacing: 4) {
                Circle()
                    .fill(isSuccess ? Color.green : Color.red)
                    .frame(width: 6, height: 6)
                
                Text(status)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(.systemGray6))
        )
    }
}

#Preview {
    QuickActionsView()
        .environmentObject(WatchTaskManager.shared)
        .environmentObject(WatchConnectivityManager.shared)
}