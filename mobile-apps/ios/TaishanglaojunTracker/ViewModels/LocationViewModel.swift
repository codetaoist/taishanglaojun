//
//  LocationViewModel.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import Combine
import CoreLocation
import SwiftUI

/// 位置追踪视图模型
class LocationViewModel: ObservableObject {
    
    // MARK: - Published Properties
    @Published var currentLocation: CLLocation?
    @Published var currentTrajectory: Trajectory?
    @Published var isTracking = false
    @Published var authorizationStatus: CLAuthorizationStatus = .notDetermined
    @Published var locationError: LocationError?
    @Published var trackingStatistics = TrackingStatistics(
        totalDistance: 0,
        totalDuration: 0,
        totalTrajectories: 0,
        averageDistance: 0,
        averageDuration: 0
    )
    
    // MARK: - Services
    private let locationService = LocationService.shared
    private let dataService = DataService.shared
    private let networkService = NetworkService.shared
    private let cryptoService = CryptoService.shared
    
    // MARK: - Private Properties
    private var cancellables = Set<AnyCancellable>()
    private var trackingStartTime: Date?
    private var lastLocationUploadTime: Date?
    private let locationUploadInterval: TimeInterval = 30 // 30秒上传一次
    
    // MARK: - Initialization
    
    init() {
        setupBindings()
        loadStatistics()
    }
    
    // MARK: - Setup
    
    private func setupBindings() {
        // 监听位置服务状态
        locationService.$authorizationStatus
            .receive(on: DispatchQueue.main)
            .assign(to: \.authorizationStatus, on: self)
            .store(in: &cancellables)
        
        locationService.$currentLocation
            .receive(on: DispatchQueue.main)
            .sink { [weak self] location in
                self?.handleLocationUpdate(location)
            }
            .store(in: &cancellables)
        
        locationService.$isTracking
            .receive(on: DispatchQueue.main)
            .assign(to: \.isTracking, on: self)
            .store(in: &cancellables)
        
        locationService.$locationError
            .receive(on: DispatchQueue.main)
            .assign(to: \.locationError, on: self)
            .store(in: &cancellables)
        
        // 监听数据服务统计信息
        dataService.$trajectories
            .receive(on: DispatchQueue.main)
            .sink { [weak self] _ in
                self?.updateStatistics()
            }
            .store(in: &cancellables)
    }
    
    // MARK: - Location Tracking
    
    /// 请求位置权限
    func requestLocationPermission() {
        locationService.requestLocationPermission()
    }
    
    /// 开始追踪
    func startTracking() {
        guard authorizationStatus == .authorizedAlways || authorizationStatus == .authorizedWhenInUse else {
            locationError = .permissionDenied
            return
        }
        
        // 创建新轨迹
        let trajectoryName = generateTrajectoryName()
        currentTrajectory = Trajectory(name: trajectoryName)
        trackingStartTime = Date()
        
        // 开始位置服务
        locationService.startTracking()
        
        print("🚀 开始追踪轨迹: \(trajectoryName)")
    }
    
    /// 停止追踪
    func stopTracking() {
        locationService.stopTracking()
        
        // 完成当前轨迹
        if var trajectory = currentTrajectory {
            trajectory.finishRecording()
            
            // 保存轨迹
            dataService.saveTrajectory(trajectory)
            
            print("⏹️ 停止追踪，轨迹已保存: \(trajectory.name)")
            
            // 重置状态
            currentTrajectory = nil
            trackingStartTime = nil
        }
    }
    
    /// 暂停追踪
    func pauseTracking() {
        locationService.stopTracking()
        print("⏸️ 暂停追踪")
    }
    
    /// 恢复追踪
    func resumeTracking() {
        guard currentTrajectory != nil else {
            startTracking()
            return
        }
        
        locationService.startTracking()
        print("▶️ 恢复追踪")
    }
    
    // MARK: - Location Handling
    
    private func handleLocationUpdate(_ location: CLLocation?) {
        guard let location = location,
              var trajectory = currentTrajectory else {
            return
        }
        
        currentLocation = location
        
        // 创建位置点
        let locationPoint = LocationPoint(from: location, trajectoryId: trajectory.id)
        
        // 添加到轨迹
        trajectory.addPoint(locationPoint)
        currentTrajectory = trajectory
        
        // 实时上传位置（如果需要）
        uploadLocationIfNeeded(locationPoint)
        
        print("📍 位置更新: \(location.coordinate.latitude), \(location.coordinate.longitude)")
    }
    
    private func uploadLocationIfNeeded(_ point: LocationPoint) {
        let now = Date()
        
        // 检查是否需要上传
        if let lastUpload = lastLocationUploadTime,
           now.timeIntervalSince(lastUpload) < locationUploadInterval {
            return
        }
        
        // 上传位置点
        networkService.uploadLocationPoint(point)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        print("❌ 位置上传失败: \(error)")
                    }
                },
                receiveValue: { success in
                    if success {
                        print("✅ 位置上传成功")
                    }
                }
            )
            .store(in: &cancellables)
        
        lastLocationUploadTime = now
    }
    
    // MARK: - Trajectory Management
    
    private func generateTrajectoryName() -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm"
        return "轨迹 \(formatter.string(from: Date()))"
    }
    
    /// 获取当前轨迹信息
    var currentTrajectoryInfo: TrajectoryInfo? {
        guard let trajectory = currentTrajectory,
              let startTime = trackingStartTime else {
            return nil
        }
        
        let duration = Date().timeIntervalSince(startTime)
        
        return TrajectoryInfo(
            name: trajectory.name,
            pointCount: trajectory.points.count,
            distance: trajectory.totalDistance,
            duration: duration,
            averageSpeed: trajectory.averageSpeed,
            maxSpeed: trajectory.maxSpeed
        )
    }
    
    // MARK: - Statistics
    
    private func loadStatistics() {
        updateStatistics()
    }
    
    private func updateStatistics() {
        trackingStatistics = dataService.statistics
    }
    
    // MARK: - Export Functions
    
    /// 导出当前轨迹为GPX
    func exportCurrentTrajectoryAsGPX() -> URL? {
        guard let trajectory = currentTrajectory else { return nil }
        return dataService.exportTrajectoryAsGPX(trajectory)
    }
    
    /// 导出当前轨迹为JSON
    func exportCurrentTrajectoryAsJSON() -> URL? {
        guard let trajectory = currentTrajectory else { return nil }
        return dataService.exportTrajectoryAsJSON(trajectory)
    }
    
    // MARK: - Privacy Functions
    
    /// 匿名化当前轨迹
    func anonymizeCurrentTrajectory() {
        guard var trajectory = currentTrajectory else { return }
        
        trajectory.points = trajectory.points.map { point in
            cryptoService.anonymizeLocation(point)
        }
        
        currentTrajectory = trajectory
        print("🔒 当前轨迹已匿名化")
    }
    
    /// 清除敏感数据
    func clearSensitiveData() {
        currentTrajectory = nil
        currentLocation = nil
        trackingStartTime = nil
        lastLocationUploadTime = nil
        
        print("🧹 敏感数据已清除")
    }
}

// MARK: - Supporting Types

struct TrajectoryInfo {
    let name: String
    let pointCount: Int
    let distance: Double
    let duration: TimeInterval
    let averageSpeed: Double
    let maxSpeed: Double
    
    var formattedDistance: String {
        if distance < 1000 {
            return String(format: "%.0f m", distance)
        } else {
            return String(format: "%.2f km", distance / 1000)
        }
    }
    
    var formattedDuration: String {
        let hours = Int(duration) / 3600
        let minutes = Int(duration) % 3600 / 60
        let seconds = Int(duration) % 60
        
        if hours > 0 {
            return String(format: "%d:%02d:%02d", hours, minutes, seconds)
        } else {
            return String(format: "%02d:%02d", minutes, seconds)
        }
    }
    
    var formattedAverageSpeed: String {
        return String(format: "%.1f km/h", averageSpeed * 3.6)
    }
    
    var formattedMaxSpeed: String {
        return String(format: "%.1f km/h", maxSpeed * 3.6)
    }
}

// MARK: - Location Error Extension
extension LocationError {
    var localizedDescription: String {
        switch self {
        case .permissionDenied:
            return "位置权限被拒绝"
        case .locationUnavailable:
            return "位置服务不可用"
        case .networkError:
            return "网络错误"
        case .unknown:
            return "未知错误"
        }
    }
}