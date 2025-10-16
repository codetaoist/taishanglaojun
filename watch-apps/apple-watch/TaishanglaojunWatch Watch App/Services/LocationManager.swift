import Foundation
import CoreLocation
import Combine

@MainActor
class LocationManager: NSObject, ObservableObject {
    // MARK: - Published Properties
    @Published var currentLocation: CLLocation?
    @Published var isTracking = false
    @Published var authorizationStatus: CLAuthorizationStatus = .notDetermined
    @Published var nearbyTasks: [WatchTask] = []
    @Published var error: String?
    
    // MARK: - Private Properties
    private let locationManager = CLLocationManager()
    private var cancellables = Set<AnyCancellable>()
    private let proximityThreshold: CLLocationDistance = 100.0 // 100 meters
    
    // MARK: - Dependencies
    private let taskManager: WatchTaskManager
    private let settingsManager: WatchSettingsManager
    
    // MARK: - Initialization
    init(taskManager: WatchTaskManager, settingsManager: WatchSettingsManager) {
        self.taskManager = taskManager
        self.settingsManager = settingsManager
        
        super.init()
        
        setupLocationManager()
        setupObservers()
    }
    
    // MARK: - Setup
    private func setupLocationManager() {
        locationManager.delegate = self
        locationManager.desiredAccuracy = kCLLocationAccuracyBest
        locationManager.distanceFilter = 10.0 // Update every 10 meters
        
        authorizationStatus = locationManager.authorizationStatus
    }
    
    private func setupObservers() {
        // Observe task changes to update nearby tasks
        taskManager.$tasks
            .sink { [weak self] tasks in
                self?.updateNearbyTasks(tasks)
            }
            .store(in: &cancellables)
    }
    
    // MARK: - Public Methods
    func requestLocationPermission() {
        switch authorizationStatus {
        case .notDetermined:
            locationManager.requestWhenInUseAuthorization()
        case .denied, .restricted:
            error = "位置权限被拒绝，请在设置中启用"
        case .authorizedWhenInUse, .authorizedAlways:
            startLocationTracking()
        @unknown default:
            break
        }
    }
    
    func startLocationTracking() {
        guard authorizationStatus == .authorizedWhenInUse || authorizationStatus == .authorizedAlways else {
            requestLocationPermission()
            return
        }
        
        guard !isTracking else { return }
        
        locationManager.startUpdatingLocation()
        isTracking = true
        error = nil
    }
    
    func stopLocationTracking() {
        guard isTracking else { return }
        
        locationManager.stopUpdatingLocation()
        isTracking = false
    }
    
    func getDistanceToTask(_ task: WatchTask) -> CLLocationDistance? {
        guard let currentLocation = currentLocation,
              let taskCoordinate = task.coordinate else {
            return nil
        }
        
        let taskLocation = CLLocation(
            latitude: taskCoordinate.latitude,
            longitude: taskCoordinate.longitude
        )
        
        return currentLocation.distance(from: taskLocation)
    }
    
    func isNearTask(_ task: WatchTask) -> Bool {
        guard let distance = getDistanceToTask(task) else { return false }
        return distance <= proximityThreshold
    }
    
    // MARK: - Private Methods
    private func updateNearbyTasks(_ tasks: [WatchTask]) {
        guard let currentLocation = currentLocation else {
            nearbyTasks = []
            return
        }
        
        let nearby = tasks.filter { task in
            guard let taskCoordinate = task.coordinate else { return false }
            
            let taskLocation = CLLocation(
                latitude: taskCoordinate.latitude,
                longitude: taskCoordinate.longitude
            )
            
            let distance = currentLocation.distance(from: taskLocation)
            return distance <= proximityThreshold
        }
        
        // Check for new nearby tasks to notify
        let newNearbyTasks = nearby.filter { task in
            !nearbyTasks.contains { $0.id == task.id }
        }
        
        for task in newNearbyTasks {
            notifyTaskProximity(task)
        }
        
        nearbyTasks = nearby
    }
    
    private func notifyTaskProximity(_ task: WatchTask) {
        // Only notify for relevant task statuses
        guard task.status == .accepted || task.status == .inProgress else { return }
        
        // Trigger haptic feedback
        settingsManager.triggerHapticFeedback(.notification)
        
        // Create local notification if enabled
        if settingsManager.notificationsEnabled {
            scheduleProximityNotification(for: task)
        }
    }
    
    private func scheduleProximityNotification(for task: WatchTask) {
        // This would integrate with the notification system
        // For now, we'll just log it
        print("Arrived near task: \(task.title)")
    }
    
    // MARK: - Utility Methods
    func clearError() {
        error = nil
    }
    
    func formatDistance(_ distance: CLLocationDistance) -> String {
        let unit = settingsManager.distanceUnit
        
        switch unit {
        case .metric:
            if distance < 1000 {
                return String(format: "%.0f米", distance)
            } else {
                return String(format: "%.1f公里", distance / 1000)
            }
        case .imperial:
            let feet = distance * 3.28084
            if feet < 5280 {
                return String(format: "%.0f英尺", feet)
            } else {
                let miles = feet / 5280
                return String(format: "%.1f英里", miles)
            }
        }
    }
    
    func getLocationAccuracyDescription() -> String {
        guard let location = currentLocation else { return "未知" }
        
        let accuracy = location.horizontalAccuracy
        
        if accuracy < 0 {
            return "无效"
        } else if accuracy < 5 {
            return "极高"
        } else if accuracy < 10 {
            return "高"
        } else if accuracy < 50 {
            return "中等"
        } else {
            return "低"
        }
    }
}

// MARK: - CLLocationManagerDelegate
extension LocationManager: CLLocationManagerDelegate {
    func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {
        guard let location = locations.last else { return }
        
        currentLocation = location
        updateNearbyTasks(taskManager.tasks)
        
        // Clear any location-related errors
        if error?.contains("位置") == true {
            error = nil
        }
    }
    
    func locationManager(_ manager: CLLocationManager, didFailWithError error: Error) {
        self.error = "位置更新失败: \(error.localizedDescription)"
        
        // Stop tracking on persistent errors
        if let clError = error as? CLError {
            switch clError.code {
            case .denied, .locationUnknown:
                stopLocationTracking()
            default:
                break
            }
        }
    }
    
    func locationManager(_ manager: CLLocationManager, didChangeAuthorization status: CLAuthorizationStatus) {
        authorizationStatus = status
        
        switch status {
        case .authorizedWhenInUse, .authorizedAlways:
            if settingsManager.locationTrackingEnabled {
                startLocationTracking()
            }
        case .denied, .restricted:
            stopLocationTracking()
            error = "位置权限被拒绝"
        case .notDetermined:
            break
        @unknown default:
            break
        }
    }
}

// MARK: - Supporting Types
extension LocationManager {
    struct LocationInfo {
        let coordinate: CLLocationCoordinate2D
        let accuracy: CLLocationAccuracy
        let timestamp: Date
        let altitude: CLLocationDistance
        let speed: CLLocationSpeed
        
        init(from location: CLLocation) {
            self.coordinate = location.coordinate
            self.accuracy = location.horizontalAccuracy
            self.timestamp = location.timestamp
            self.altitude = location.altitude
            self.speed = location.speed
        }
    }
}