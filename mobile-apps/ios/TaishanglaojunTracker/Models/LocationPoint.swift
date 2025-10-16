//
//  LocationPoint.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import CoreLocation

/// 位置点数据模型
struct LocationPoint: Codable, Identifiable {
    let id: UUID
    let latitude: Double
    let longitude: Double
    let timestamp: Date
    let accuracy: Double
    let speed: Double?
    let course: Double?
    let altitude: Double?
    let trajectoryId: UUID
    
    init(from location: CLLocation, trajectoryId: UUID) {
        self.id = UUID()
        self.latitude = location.coordinate.latitude
        self.longitude = location.coordinate.longitude
        self.timestamp = location.timestamp
        self.accuracy = location.horizontalAccuracy
        self.speed = location.speed >= 0 ? location.speed : nil
        self.course = location.course >= 0 ? location.course : nil
        self.altitude = location.altitude
        self.trajectoryId = trajectoryId
    }
    
    init(id: UUID = UUID(), 
         latitude: Double, 
         longitude: Double, 
         timestamp: Date = Date(), 
         accuracy: Double, 
         speed: Double? = nil, 
         course: Double? = nil, 
         altitude: Double? = nil, 
         trajectoryId: UUID) {
        self.id = id
        self.latitude = latitude
        self.longitude = longitude
        self.timestamp = timestamp
        self.accuracy = accuracy
        self.speed = speed
        self.course = course
        self.altitude = altitude
        self.trajectoryId = trajectoryId
    }
}

// MARK: - Core Location Extensions
extension LocationPoint {
    /// 转换为CLLocationCoordinate2D
    var coordinate: CLLocationCoordinate2D {
        CLLocationCoordinate2D(latitude: latitude, longitude: longitude)
    }
    
    /// 转换为CLLocation
    var clLocation: CLLocation {
        CLLocation(
            coordinate: coordinate,
            altitude: altitude ?? 0,
            horizontalAccuracy: accuracy,
            verticalAccuracy: -1,
            course: course ?? -1,
            speed: speed ?? -1,
            timestamp: timestamp
        )
    }
    
    /// 计算与另一个位置点的距离（米）
    func distance(to other: LocationPoint) -> Double {
        let location1 = CLLocation(latitude: latitude, longitude: longitude)
        let location2 = CLLocation(latitude: other.latitude, longitude: other.longitude)
        return location1.distance(from: location2)
    }
}

// MARK: - Validation
extension LocationPoint {
    /// 验证位置点数据是否有效
    var isValid: Bool {
        // 检查纬度范围
        guard latitude >= -90 && latitude <= 90 else { return false }
        
        // 检查经度范围
        guard longitude >= -180 && longitude <= 180 else { return false }
        
        // 检查精度（负值表示无效）
        guard accuracy >= 0 else { return false }
        
        // 检查时间戳不能是未来时间
        guard timestamp <= Date() else { return false }
        
        return true
    }
    
    /// 检查位置精度是否足够好
    var hasGoodAccuracy: Bool {
        accuracy <= 50 // 精度在50米以内认为是好的
    }
}

// MARK: - Formatting
extension LocationPoint {
    /// 格式化坐标显示
    var formattedCoordinate: String {
        String(format: "%.6f, %.6f", latitude, longitude)
    }
    
    /// 格式化时间显示
    var formattedTime: String {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        formatter.timeStyle = .medium
        return formatter.string(from: timestamp)
    }
    
    /// 格式化速度显示
    var formattedSpeed: String {
        guard let speed = speed, speed >= 0 else { return "未知" }
        return String(format: "%.1f m/s", speed)
    }
    
    /// 格式化精度显示
    var formattedAccuracy: String {
        String(format: "±%.1f m", accuracy)
    }
}