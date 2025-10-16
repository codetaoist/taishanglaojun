import SwiftUI
import WatchKit

struct ContentView: View {
    @EnvironmentObject var taskManager: WatchTaskManager
    @EnvironmentObject var connectivityManager: WatchConnectivityManager
    @State private var selectedTab = 0
    
    var body: some View {
        TabView(selection: $selectedTab) {
            // 任务概览
            TaskOverviewView()
                .tabItem {
                    Image(systemName: "list.bullet")
                    Text("任务")
                }
                .tag(0)
            
            // 快速操作
            QuickActionsView()
                .tabItem {
                    Image(systemName: "bolt.fill")
                    Text("快捷")
                }
                .tag(1)
            
            // 通知中心
            NotificationView()
                .tabItem {
                    Image(systemName: "bell.fill")
                    Text("通知")
                }
                .tag(2)
            
            // 设置
            SettingsView()
                .tabItem {
                    Image(systemName: "gear")
                    Text("设置")
                }
                .tag(3)
        }
        .tabViewStyle(PageTabViewStyle())
        .onAppear {
            setupInitialView()
        }
    }
    
    private func setupInitialView() {
        // 检查连接状态
        if !connectivityManager.isConnected {
            // 显示连接提示
            showConnectionAlert()
        }
        
        // 加载任务数据
        taskManager.loadTasks { success in
            if !success {
                // 显示加载失败提示
                showLoadingErrorAlert()
            }
        }
    }
    
    private func showConnectionAlert() {
        // 显示连接状态提示
        WKInterfaceDevice.current().play(.failure)
    }
    
    private func showLoadingErrorAlert() {
        // 显示加载错误提示
        WKInterfaceDevice.current().play(.failure)
    }
}

#Preview {
    ContentView()
        .environmentObject(WatchTaskManager())
        .environmentObject(WatchConnectivityManager())
}