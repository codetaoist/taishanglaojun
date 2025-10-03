import SwiftUI
import WatchKit
import WatchConnectivity

@main
struct TaishanglaojunWatchApp: App {
    @StateObject private var connectivityManager = WatchConnectivityManager()
    @StateObject private var taskManager = WatchTaskManager()
    @StateObject private var settingsManager = WatchSettingsManager()
    
    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(connectivityManager)
                .environmentObject(taskManager)
                .environmentObject(settingsManager)
                .onAppear {
                    setupApp()
                }
        }
    }
    
    private func setupApp() {
        // 初始化应用设置
        WatchFeatures.setupHapticFeedback()
        WatchFeatures.setupComplications()
        
        // 启动连接管理器
        connectivityManager.activate()
        
        // 加载初始数据
        taskManager.loadInitialData()
    }
}

// MARK: - 应用生命周期管理
extension TaishanglaojunWatchApp {
    func scenePhase(_ phase: ScenePhase) {
        switch phase {
        case .active:
            // 应用激活时刷新数据
            taskManager.refreshData()
        case .inactive:
            // 应用非激活状态
            break
        case .background:
            // 应用进入后台
            taskManager.saveCurrentState()
        @unknown default:
            break
        }
    }
}