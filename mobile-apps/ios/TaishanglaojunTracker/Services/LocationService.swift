//
//  LocationService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import CoreLocation
import Combine

/// 位置服务管理器
class LocationService: NSObject, ObservableObject {
    
    // MARK: - Published Properties
    @Published var authorizationStatus: CLAuthorizationStatus = .notDetermined
    @Published var currentLocation: CLLocation?
    @Published var isTracking = false
    @Published var locationError: LocationError?
    
    // MARK: - Private Properties
    private let locationManager = CLLocationManager()
    private var currentTrajectory: Trajectory?
    private let dataService = DataService.shared
    
    // MARK: - Configuration
    private struct LocationConfig {
        static let desiredAccuracy = kCLLocationAccuracyBest
        static let distanceFilter: CLLocationDistance = 5.0 // 5米
        static let minimumTimeInterval: TimeInterval = 10.0 // 10秒
        static let maximumAccuracy: CLLocationAccuracy = 50.0 // 50米
    }
    
    // MARK: - Singleton
    static let shared = LocationService()
    
    override init() {
        super.init()
        setupLocationManager()
    }
    
    // MARK: - Setup
    private func setupLocationManager() {
        locationManager.delegate = self
        locationManager.desiredAccuracy = LocationConfig.desiredAccuracy
        locationManager.distanceFilter = LocationConfig.distanceFilter
        
        // 允许后台位置更新
        locationManager.allowsBackgroundLocationUpdates = true
        locationManager.pausesLocationUpdatesAutomatically = false
        
        // 更新授权状态
        authorizationStatus = locationManager.authorizationStatus
    }
    
    // MARK: - Public Methods
    
    /// 请求位置权限
    func requestLocationPermission() {
        switch authorizationStatus {
        case .notDetermined:
            locationManager.requestWhenInUseAuthorization()
        case .denied, .restricted:
            locationError = .permissionDenied
        case .authorizedWhenInUse:
            locationManager.requestAlwaysAuthorization()
        case .authorizedAlways:
            break
        @unknown default:
            break
        }
    }
    
    /// 开始位置追踪
    func startTracking() {
        guard canStartTracking() else { return }
        
        // 创建新轨迹
        currentTrajectory = Trajectory()
        
        // 开始位置更新
        locationManager.startUpdatingLocation()
        locationManager.startMonitoringSignificantLocationChanges()
        
        isTracking = true
        locationError = nil
        
        print("📍 开始位置追踪")
    }
    
    /// 停止位置追踪
    func stopTracking() {
        locationManager.stopUpdatingLocation()
        locationManager.stopMonitoringSignificantLocationChanges()
        
        // 保存当前轨迹
        if var trajectory = currentTrajectory {
            trajectory.finishRecording()
            dataService.saveTrajectory(trajectory)
        }
        
        currentTrajectory = nil
        isTracking = false
        
        print("🛑 停止位置追踪")
    }
    
    /// 获取当前位置（一次性）
    func getCurrentLocation() {
        guard authorizationStatus == .authorizedWhenInUse || authorizationStatus == .authorizedAlways else {
            locationError = .permissionDenied
            return
        }
        
        locationManager.requestLocation()
    }
    
    // MARK: - Private Methods
    
    private func canStartTracking() -> Bool {
        // 检查权限
        guard authorizationStatus == .authorizedAlways else {
            locationError = .permissionDenied
            return false
        }
        
        // 检查位置服务是否可用
        guard CLLocationManager.locationServicesEnabled() else {
            locationError = .locationServicesDisabled
            return false
        }
        
        // 检查是否已在追踪
        guard !isTracking else {
            return false
        }
        
        return true
    }
    
    private func processLocationUpdate(_ location: CLLocation) {
        // 更新当前位置
        currentLocation = location
        
        // 验证位置精度
        guard location.horizontalAccuracy <= LocationConfig.maximumAccuracy else {
            print("⚠️ 位置精度不够: \(location.horizontalAccuracy)m")
            return
        }
        
        // 验证位置时间
        guard abs(location.timestamp.timeIntervalSinceNow) < 30 else {
            print("⚠️ 位置数据过旧")
            return
        }
        
        // 添加到当前轨迹
        if var trajectory = currentTrajectory {
            let locationPoint = LocationPoint(from: location, trajectoryId: trajectory.id)
            trajectory.addPoint(locationPoint)
            currentTrajectory = trajectory
            
            print("📍 记录位置: \(locationPoint.formattedCoordinate), 精度: \(locationPoint.formattedAccuracy)")
        }
    }
}

// MARK: - CLLocationManagerDelegate
extension LocationService: CLLocationManagerDelegate {
    
    func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {
        guard let location = locations.last else { return }
        processLocationUpdate(location)
    }
    
    func locationManager(_ manager: CLLocationManager, didFailWithError error: Error) {
        print("❌ 位置更新失败: \(error.localizedDescription)")
        
        if let clError = error as? CLError {
            switch clError.code {
            case .denied:
                locationError = .permissionDenied
            case .locationUnknown:
                locationError = .locationUnavailable
            case .network:
                locationError = .networkError
            default:
                locationError = .unknown(error.localizedDescription)
            }
        } else {
            locationError = .unknown(error.localizedDescription)
        }
    }
    
    func locationManager(_ manager: CLLocationManager, didChangeAuthorization status: CLAuthorizationStatus) {
        DispatchQueue.main.async {
            self.authorizationStatus = status
        }
        
        switch status {
        case .notDetermined:
            print("📍 位置权限: 未确定")
        case .denied, .restricted:
            print("❌ 位置权限: 被拒绝")
            locationError = .permissionDenied
            if isTracking {
                stopTracking()
            }
        case .authorizedWhenInUse:
            print("✅ 位置权限: 使用时允许")
        case .authorizedAlways:
            print("✅ 位置权限: 始终允许")
        @unknown default:
            break
        }
    }
    
    func locationManagerDidPauseLocationUpdates(_ manager: CLLocationManager) {
        print("⏸️ 位置更新已暂停")
    }
    
    func locationManagerDidResumeLocationUpdates(_ manager: CLLocationManager) {
        print("▶️ 位置更新已恢复")
    }
}

// MARK: - LocationError
enum LocationError: LocalizedError {
    case permissionDenied
    case locationServicesDisabled
    case locationUnavailable
    case networkError
    case unknown(String)
    
    var errorDescription: String? {
        switch self {
        case .permissionDenied:
            return "位置权限被拒绝，请在设置中允许位置访问"
        case .locationServicesDisabled:
            return "位置服务已关闭，请在设置中开启位置服务"
        case .locationUnavailable:
            return "无法获取位置信息，请检查GPS信号"
        case .networkError:
            return "网络错误，无法获取位置信息"
        case .unknown(let message):
            return "未知错误: \(message)"
        }
    }
    
    var recoverySuggestion: String? {
        switch self {
        case .permissionDenied:
            return "前往设置 > 隐私与安全性 > 定位服务，允许应用访问位置"
        case .locationServicesDisabled:
            return "前往设置 > 隐私与安全性 > 定位服务，开启定位服务"
        case .locationUnavailable:
            return "请移动到空旷区域，确保GPS信号良好"
        case .networkError:
            return "请检查网络连接"
        case .unknown:
            return "请重启应用或联系技术支持"
        }
    }
}

// MARK: - Background Location Updates
extension LocationService {
    
    /// 配置后台位置更新
    func configureBackgroundLocationUpdates() {
        guard authorizationStatus == .authorizedAlways else { return }
        
        // 设置后台位置更新
        locationManager.allowsBackgroundLocationUpdates = true
        locationManager.pausesLocationUpdatesAutomatically = false
        
        // 设置后台任务标识符
        // 需要在Info.plist中配置UIBackgroundModes
    }
    
    /// 处理应用进入后台
    func handleAppDidEnterBackground() {
        if isTracking {
            print("📱 应用进入后台，继续位置追踪")
            // 可以在这里调整位置更新频率以节省电量
            locationManager.desiredAccuracy = kCLLocationAccuracyHundredMeters
        }
    }
    
    /// 处理应用进入前台
    func handleAppWillEnterForeground() {
        if isTracking {
            print("📱 应用进入前台，恢复高精度位置追踪")
            locationManager.desiredAccuracy = LocationConfig.desiredAccuracy
        }
    }
}