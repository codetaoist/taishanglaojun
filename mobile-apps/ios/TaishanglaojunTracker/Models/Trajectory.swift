//
//  Trajectory.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import CoreLocation

/// 轨迹数据模型
struct Trajectory: Codable, Identifiable {
    let id: UUID
    var name: String
    let startTime: Date
    var endTime: Date?
    var points: [LocationPoint]
    
    init(id: UUID = UUID(), name: String? = nil) {
        self.id = id
        self.name = name ?? "轨迹 \(DateFormatter.shortDate.string(from: Date()))"
        self.startTime = Date()
        self.endTime = nil
        self.points = []
    }
}

// MARK: - Computed Properties
extension Trajectory {
    /// 轨迹是否正在记录中
    var isRecording: Bool {
        endTime == nil
    }
    
    /// 轨迹总距离（米）
    var totalDistance: Double {
        guard points.count > 1 else { return 0 }
        
        var distance: Double = 0
        for i in 1..<points.count {
            distance += points[i-1].distance(to: points[i])
        }
        return distance
    }
    
    /// 轨迹持续时间（秒）
    var duration: TimeInterval {
        let end = endTime ?? Date()
        return end.timeIntervalSince(startTime)
    }
    
    /// 平均速度（m/s）
    var averageSpeed: Double {
        guard duration > 0 else { return 0 }
        return totalDistance / duration
    }
    
    /// 最高速度（m/s）
    var maxSpeed: Double {
        points.compactMap { $0.speed }.max() ?? 0
    }
    
    /// 轨迹边界框
    var boundingBox: (minLat: Double, maxLat: Double, minLon: Double, maxLon: Double)? {
        guard !points.isEmpty else { return nil }
        
        let latitudes = points.map { $0.latitude }
        let longitudes = points.map { $0.longitude }
        
        return (
            minLat: latitudes.min()!,
            maxLat: latitudes.max()!,
            minLon: longitudes.min()!,
            maxLon: longitudes.max()!
        )
    }
    
    /// 轨迹中心点
    var centerCoordinate: CLLocationCoordinate2D? {
        guard let bounds = boundingBox else { return nil }
        
        return CLLocationCoordinate2D(
            latitude: (bounds.minLat + bounds.maxLat) / 2,
            longitude: (bounds.minLon + bounds.maxLon) / 2
        )
    }
    
    /// 起始位置
    var startLocation: LocationPoint? {
        points.first
    }
    
    /// 结束位置
    var endLocation: LocationPoint? {
        points.last
    }
}

// MARK: - Trajectory Management
extension Trajectory {
    /// 添加位置点
    mutating func addPoint(_ point: LocationPoint) {
        // 验证位置点是否有效
        guard point.isValid else { return }
        
        // 如果是第一个点，直接添加
        if points.isEmpty {
            points.append(point)
            return
        }
        
        // 检查与上一个点的距离和时间间隔
        let lastPoint = points.last!
        let distance = lastPoint.distance(to: point)
        let timeInterval = point.timestamp.timeIntervalSince(lastPoint.timestamp)
        
        // 过滤掉距离太近或时间间隔太短的点
        guard distance > 5 || timeInterval > 10 else { return }
        
        points.append(point)
    }
    
    /// 结束轨迹记录
    mutating func finishRecording() {
        endTime = Date()
    }
    
    /// 清空轨迹点
    mutating func clearPoints() {
        points.removeAll()
    }
    
    /// 删除指定范围的点
    mutating func removePoints(in range: Range<Int>) {
        guard range.lowerBound >= 0 && range.upperBound <= points.count else { return }
        points.removeSubrange(range)
    }
}

// MARK: - Statistics
extension Trajectory {
    /// 轨迹统计信息
    struct Statistics {
        let totalDistance: Double
        let duration: TimeInterval
        let averageSpeed: Double
        let maxSpeed: Double
        let pointCount: Int
        let startTime: Date
        let endTime: Date?
    }
    
    /// 获取轨迹统计信息
    var statistics: Statistics {
        Statistics(
            totalDistance: totalDistance,
            duration: duration,
            averageSpeed: averageSpeed,
            maxSpeed: maxSpeed,
            pointCount: points.count,
            startTime: startTime,
            endTime: endTime
        )
    }
}

// MARK: - Formatting
extension Trajectory {
    /// 格式化距离显示
    var formattedDistance: String {
        let distance = totalDistance
        if distance < 1000 {
            return String(format: "%.0f m", distance)
        } else {
            return String(format: "%.2f km", distance / 1000)
        }
    }
    
    /// 格式化持续时间显示
    var formattedDuration: String {
        let duration = self.duration
        let hours = Int(duration) / 3600
        let minutes = Int(duration) % 3600 / 60
        let seconds = Int(duration) % 60
        
        if hours > 0 {
            return String(format: "%d:%02d:%02d", hours, minutes, seconds)
        } else {
            return String(format: "%02d:%02d", minutes, seconds)
        }
    }
    
    /// 格式化平均速度显示
    var formattedAverageSpeed: String {
        let speed = averageSpeed * 3.6 // 转换为 km/h
        return String(format: "%.1f km/h", speed)
    }
    
    /// 格式化开始时间显示
    var formattedStartTime: String {
        DateFormatter.fullDateTime.string(from: startTime)
    }
}

// MARK: - Export
extension Trajectory {
    /// 导出为GPX格式
    func exportToGPX() -> String {
        var gpx = """
        <?xml version="1.0" encoding="UTF-8"?>
        <gpx version="1.1" creator="TaishanglaojunTracker">
        <trk>
        <name>\(name)</name>
        <trkseg>
        """
        
        for point in points {
            gpx += """
            <trkpt lat="\(point.latitude)" lon="\(point.longitude)">
            <time>\(ISO8601DateFormatter().string(from: point.timestamp))</time>
            """
            
            if let altitude = point.altitude {
                gpx += "<ele>\(altitude)</ele>"
            }
            
            gpx += "</trkpt>"
        }
        
        gpx += """
        </trkseg>
        </trk>
        </gpx>
        """
        
        return gpx
    }
    
    /// 导出为JSON格式
    func exportToJSON() -> Data? {
        try? JSONEncoder().encode(self)
    }
}

// MARK: - DateFormatter Extensions
private extension DateFormatter {
    static let shortDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        return formatter
    }()
    
    static let fullDateTime: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        return formatter
    }()
}