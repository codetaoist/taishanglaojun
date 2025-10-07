//
//  DataSyncService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import Combine
import Network

/// 数据同步服务
class DataSyncService: ObservableObject {
    
    // MARK: - Singleton
    static let shared = DataSyncService()
    
    // MARK: - Published Properties
    @Published var syncState: SyncState = .idle
    @Published var syncStats = SyncStats()
    @Published var lastSyncError: String?
    
    // MARK: - Private Properties
    private let networkService = NetworkService.shared
    private let dataService = DataService.shared
    private let cryptoService = CryptoService.shared
    
    private var syncTimer: Timer?
    private var cancellables = Set<AnyCancellable>()
    private let networkMonitor = NWPathMonitor()
    private let monitorQueue = DispatchQueue(label: "NetworkMonitor")
    
    // 同步配置
    private let syncInterval: TimeInterval = 30 // 30秒
    private let retryDelay: TimeInterval = 60 // 重试延迟60秒
    private let maxRetryCount = 3
    
    private init() {
        setupNetworkMonitoring()
    }
    
    deinit {
        stopPeriodicSync()
        networkMonitor.cancel()
    }
    
    // MARK: - Network Monitoring
    
    private func setupNetworkMonitoring() {
        networkMonitor.pathUpdateHandler = { [weak self] path in
            DispatchQueue.main.async {
                if path.status == .satisfied && self?.syncState == .waitingNetwork {
                    self?.performSync()
                }
            }
        }
        networkMonitor.start(queue: monitorQueue)
    }
    
    // MARK: - Public Methods
    
    /// 开始周期性同步
    func startPeriodicSync() {
        guard syncTimer == nil else {
            print("🔄 同步任务已在运行")
            return
        }
        
        print("🚀 开始周期性数据同步")
        syncState = .running
        
        // 立即执行一次同步
        performSync()
        
        // 设置定时器
        syncTimer = Timer.scheduledTimer(withTimeInterval: syncInterval, repeats: true) { [weak self] _ in
            self?.performSync()
        }
    }
    
    /// 停止周期性同步
    func stopPeriodicSync() {
        print("🛑 停止周期性数据同步")
        syncTimer?.invalidate()
        syncTimer = nil
        syncState = .idle
    }
    
    /// 执行强制同步
    func forceSync() {
        print("⚡ 执行强制同步")
        performSync()
    }
    
    // MARK: - Private Methods
    
    /// 执行数据同步
    private func performSync() {
        guard networkMonitor.currentPath.status == .satisfied else {
            print("⚠️ 网络不可用，等待网络连接")
            syncState = .waitingNetwork
            return
        }
        
        syncState = .running
        lastSyncError = nil
        
        let startTime = Date()
        
        Task {
            do {
                let locationSyncCount = await syncLocationData()
                let trajectorySyncCount = await syncTrajectoryData()
                let chatSyncCount = await syncChatData()
                
                await MainActor.run {
                    let duration = Date().timeIntervalSince(startTime)
                    
                    syncStats = SyncStats(
                        lastSyncTime: Date(),
                        totalSyncCount: syncStats.totalSyncCount + 1,
                        locationSyncCount: syncStats.locationSyncCount + locationSyncCount,
                        trajectorySyncCount: syncStats.trajectorySyncCount + trajectorySyncCount,
                        chatSyncCount: syncStats.chatSyncCount + chatSyncCount,
                        errorCount: syncStats.errorCount,
                        lastSyncDuration: duration
                    )
                    
                    syncState = .success
                    
                    print("✅ 数据同步完成 - 位置: \(locationSyncCount), 轨迹: \(trajectorySyncCount), 聊天: \(chatSyncCount), 耗时: \(String(format: "%.2f", duration))s")
                }
                
            } catch {
                await MainActor.run {
                    syncState = .error
                    lastSyncError = error.localizedDescription
                    syncStats = SyncStats(
                        lastSyncTime: syncStats.lastSyncTime,
                        totalSyncCount: syncStats.totalSyncCount,
                        locationSyncCount: syncStats.locationSyncCount,
                        trajectorySyncCount: syncStats.trajectorySyncCount,
                        chatSyncCount: syncStats.chatSyncCount,
                        errorCount: syncStats.errorCount + 1,
                        lastSyncDuration: syncStats.lastSyncDuration
                    )
                    
                    print("❌ 数据同步失败: \(error.localizedDescription)")
                }
            }
        }
    }
    
    /// 同步位置数据
    private func syncLocationData() async -> Int {
        do {
            let unsyncedPoints = dataService.getUnsyncedLocationPoints()
            
            guard !unsyncedPoints.isEmpty else {
                print("📍 没有未同步的位置数据")
                return 0
            }
            
            print("📍 开始同步 \(unsyncedPoints.count) 个位置点")
            
            let response = try await networkService.uploadLocationPoints(unsyncedPoints)
            
            if response.success {
                // 标记为已同步
                dataService.markLocationPointsAsSynced(unsyncedPoints.map { $0.id })
                print("✅ 位置数据同步成功")
                return unsyncedPoints.count
            } else {
                print("❌ 位置数据同步失败: \(response.message ?? "未知错误")")
                return 0
            }
            
        } catch {
            print("❌ 位置数据同步异常: \(error.localizedDescription)")
            return 0
        }
    }
    
    /// 同步轨迹数据
    private func syncTrajectoryData() async -> Int {
        do {
            let unsyncedTrajectories = dataService.getUnsyncedTrajectories()
            
            guard !unsyncedTrajectories.isEmpty else {
                print("🛤️ 没有未同步的轨迹数据")
                return 0
            }
            
            print("🛤️ 开始同步 \(unsyncedTrajectories.count) 条轨迹")
            
            var syncCount = 0
            
            for trajectory in unsyncedTrajectories {
                let response = try await networkService.uploadTrajectory(trajectory)
                
                if response.success {
                    // 标记为已同步
                    var syncedTrajectory = trajectory
                    syncedTrajectory.synced = true
                    dataService.saveTrajectory(syncedTrajectory)
                    syncCount += 1
                } else {
                    print("❌ 轨迹同步失败: \(response.message ?? "未知错误")")
                }
            }
            
            print("✅ 轨迹数据同步完成，成功: \(syncCount)/\(unsyncedTrajectories.count)")
            return syncCount
            
        } catch {
            print("❌ 轨迹数据同步异常: \(error.localizedDescription)")
            return 0
        }
    }
    
    /// 同步聊天数据
    private func syncChatData() async -> Int {
        // 这里可以实现聊天数据的同步逻辑
        // 目前返回0，表示没有聊天数据需要同步
        print("💬 聊天数据同步（暂未实现）")
        return 0
    }
}

// MARK: - Supporting Types

extension DataSyncService {
    
    /// 同步状态
    enum SyncState {
        case idle           // 空闲
        case running        // 运行中
        case success        // 成功
        case error          // 错误
        case waitingNetwork // 等待网络
        
        var description: String {
            switch self {
            case .idle: return "空闲"
            case .running: return "同步中"
            case .success: return "同步成功"
            case .error: return "同步失败"
            case .waitingNetwork: return "等待网络"
            }
        }
    }
    
    /// 同步统计
    struct SyncStats {
        let lastSyncTime: Date
        let totalSyncCount: Int
        let locationSyncCount: Int
        let trajectorySyncCount: Int
        let chatSyncCount: Int
        let errorCount: Int
        let lastSyncDuration: TimeInterval
        
        init(
            lastSyncTime: Date = Date(),
            totalSyncCount: Int = 0,
            locationSyncCount: Int = 0,
            trajectorySyncCount: Int = 0,
            chatSyncCount: Int = 0,
            errorCount: Int = 0,
            lastSyncDuration: TimeInterval = 0
        ) {
            self.lastSyncTime = lastSyncTime
            self.totalSyncCount = totalSyncCount
            self.locationSyncCount = locationSyncCount
            self.trajectorySyncCount = trajectorySyncCount
            self.chatSyncCount = chatSyncCount
            self.errorCount = errorCount
            self.lastSyncDuration = lastSyncDuration
        }
        
        var formattedLastSyncTime: String {
            let formatter = DateFormatter()
            formatter.dateStyle = .short
            formatter.timeStyle = .medium
            return formatter.string(from: lastSyncTime)
        }
        
        var formattedDuration: String {
            return String(format: "%.2fs", lastSyncDuration)
        }
    }
}

// MARK: - Extensions

extension DataService {
    
    /// 获取未同步的位置点
    func getUnsyncedLocationPoints() -> [LocationPoint] {
        // 这里应该从本地数据库获取未同步的位置点
        // 目前返回空数组
        return []
    }
    
    /// 标记位置点为已同步
    func markLocationPointsAsSynced(_ ids: [UUID]) {
        // 这里应该更新本地数据库中位置点的同步状态
        print("📍 标记 \(ids.count) 个位置点为已同步")
    }
    
    /// 获取未同步的轨迹
    func getUnsyncedTrajectories() -> [Trajectory] {
        return trajectories.filter { !$0.synced }
    }
}