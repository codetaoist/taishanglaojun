//
//  DataService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import CoreData
import Combine
import CryptoKit

/// 数据服务管理器
class DataService: ObservableObject {
    
    // MARK: - Published Properties
    @Published var trajectories: [Trajectory] = []
    @Published var isLoading = false
    @Published var syncStatus: SyncStatus = .idle
    
    // MARK: - Private Properties
    private let networkService = NetworkService.shared
    private let cryptoService = CryptoService.shared
    private var cancellables = Set<AnyCancellable>()
    
    // MARK: - Core Data Stack
    lazy var persistentContainer: NSPersistentContainer = {
        let container = NSPersistentContainer(name: "TaishanglaojunTracker")
        container.loadPersistentStores { _, error in
            if let error = error {
                fatalError("Core Data error: \(error)")
            }
        }
        return container
    }()
    
    private var context: NSManagedObjectContext {
        persistentContainer.viewContext
    }
    
    // MARK: - Singleton
    static let shared = DataService()
    
    private init() {
        loadTrajectories()
        setupAutoSync()
    }
    
    // MARK: - Trajectory Management
    
    /// 保存轨迹
    func saveTrajectory(_ trajectory: Trajectory) {
        // 加密敏感数据
        let encryptedTrajectory = cryptoService.encryptTrajectory(trajectory)
        
        // 保存到本地数据库
        saveToLocalDatabase(encryptedTrajectory)
        
        // 添加到内存数组
        if let index = trajectories.firstIndex(where: { $0.id == trajectory.id }) {
            trajectories[index] = trajectory
        } else {
            trajectories.append(trajectory)
        }
        
        // 异步上传到服务器
        uploadTrajectoryAsync(trajectory)
        
        print("💾 轨迹已保存: \(trajectory.name)")
    }
    
    /// 删除轨迹
    func deleteTrajectory(_ trajectory: Trajectory) {
        // 从本地数据库删除
        deleteFromLocalDatabase(trajectory.id)
        
        // 从内存数组删除
        trajectories.removeAll { $0.id == trajectory.id }
        
        // 从服务器删除
        deleteTrajectoryFromServer(trajectory.id)
        
        print("🗑️ 轨迹已删除: \(trajectory.name)")
    }
    
    /// 更新轨迹
    func updateTrajectory(_ trajectory: Trajectory) {
        saveTrajectory(trajectory)
    }
    
    /// 获取轨迹详情
    func getTrajectory(by id: UUID) -> Trajectory? {
        return trajectories.first { $0.id == id }
    }
    
    // MARK: - Local Database Operations
    
    private func saveToLocalDatabase(_ trajectory: Trajectory) {
        let entity = NSEntityDescription.entity(forEntityName: "TrajectoryEntity", in: context)!
        let trajectoryEntity = NSManagedObject(entity: entity, insertInto: context)
        
        trajectoryEntity.setValue(trajectory.id.uuidString, forKey: "id")
        trajectoryEntity.setValue(trajectory.name, forKey: "name")
        trajectoryEntity.setValue(trajectory.startTime, forKey: "startTime")
        trajectoryEntity.setValue(trajectory.endTime, forKey: "endTime")
        
        // 序列化位置点数据
        if let pointsData = try? JSONEncoder().encode(trajectory.points) {
            let encryptedData = cryptoService.encrypt(pointsData)
            trajectoryEntity.setValue(encryptedData, forKey: "pointsData")
        }
        
        saveContext()
    }
    
    private func loadTrajectories() {
        isLoading = true
        
        let request: NSFetchRequest<NSManagedObject> = NSFetchRequest(entityName: "TrajectoryEntity")
        request.sortDescriptors = [NSSortDescriptor(key: "startTime", ascending: false)]
        
        do {
            let results = try context.fetch(request)
            var loadedTrajectories: [Trajectory] = []
            
            for result in results {
                if let trajectory = parseTrajectoryEntity(result) {
                    loadedTrajectories.append(trajectory)
                }
            }
            
            DispatchQueue.main.async {
                self.trajectories = loadedTrajectories
                self.isLoading = false
            }
            
            print("📚 已加载 \(loadedTrajectories.count) 条轨迹")
            
        } catch {
            print("❌ 加载轨迹失败: \(error)")
            isLoading = false
        }
    }
    
    private func parseTrajectoryEntity(_ entity: NSManagedObject) -> Trajectory? {
        guard let idString = entity.value(forKey: "id") as? String,
              let id = UUID(uuidString: idString),
              let name = entity.value(forKey: "name") as? String,
              let startTime = entity.value(forKey: "startTime") as? Date else {
            return nil
        }
        
        var trajectory = Trajectory(id: id, name: name)
        trajectory.startTime = startTime
        trajectory.endTime = entity.value(forKey: "endTime") as? Date
        
        // 解密并解析位置点数据
        if let encryptedData = entity.value(forKey: "pointsData") as? Data,
           let decryptedData = cryptoService.decrypt(encryptedData),
           let points = try? JSONDecoder().decode([LocationPoint].self, from: decryptedData) {
            trajectory.points = points
        }
        
        return trajectory
    }
    
    private func deleteFromLocalDatabase(_ id: UUID) {
        let request: NSFetchRequest<NSManagedObject> = NSFetchRequest(entityName: "TrajectoryEntity")
        request.predicate = NSPredicate(format: "id == %@", id.uuidString)
        
        do {
            let results = try context.fetch(request)
            for result in results {
                context.delete(result)
            }
            saveContext()
        } catch {
            print("❌ 删除轨迹失败: \(error)")
        }
    }
    
    private func saveContext() {
        if context.hasChanges {
            do {
                try context.save()
            } catch {
                print("❌ 保存上下文失败: \(error)")
            }
        }
    }
    
    // MARK: - Network Sync
    
    private func setupAutoSync() {
        // 每5分钟自动同步一次
        Timer.publish(every: 300, on: .main, in: .common)
            .autoconnect()
            .sink { _ in
                self.syncWithServer()
            }
            .store(in: &cancellables)
    }
    
    /// 与服务器同步
    func syncWithServer() {
        guard syncStatus != .syncing else { return }
        
        syncStatus = .syncing
        
        // 上传本地未同步的轨迹
        let unsyncedTrajectories = trajectories.filter { !$0.isSynced }
        
        let uploadTasks = unsyncedTrajectories.map { trajectory in
            uploadTrajectory(trajectory)
        }
        
        Publishers.MergeMany(uploadTasks)
            .collect()
            .sink(
                receiveCompletion: { completion in
                    DispatchQueue.main.async {
                        switch completion {
                        case .finished:
                            self.syncStatus = .success
                        case .failure(let error):
                            self.syncStatus = .failed(error)
                        }
                    }
                },
                receiveValue: { _ in
                    print("✅ 同步完成")
                }
            )
            .store(in: &cancellables)
    }
    
    private func uploadTrajectoryAsync(_ trajectory: Trajectory) {
        uploadTrajectory(trajectory)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        print("❌ 上传轨迹失败: \(error)")
                    }
                },
                receiveValue: { success in
                    if success {
                        print("✅ 轨迹上传成功: \(trajectory.name)")
                    }
                }
            )
            .store(in: &cancellables)
    }
    
    private func uploadTrajectory(_ trajectory: Trajectory) -> AnyPublisher<Bool, Error> {
        return networkService.uploadTrajectory(trajectory)
            .map { _ in true }
            .eraseToAnyPublisher()
    }
    
    private func deleteTrajectoryFromServer(_ id: UUID) {
        networkService.deleteTrajectory(id)
            .sink(
                receiveCompletion: { completion in
                    if case .failure(let error) = completion {
                        print("❌ 删除服务器轨迹失败: \(error)")
                    }
                },
                receiveValue: { success in
                    if success {
                        print("✅ 服务器轨迹删除成功")
                    }
                }
            )
            .store(in: &cancellables)
    }
}

// MARK: - Export Functions
extension DataService {
    
    /// 导出轨迹为GPX文件
    func exportTrajectoryAsGPX(_ trajectory: Trajectory) -> URL? {
        let gpxContent = trajectory.exportToGPX()
        
        let documentsPath = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
        let fileName = "\(trajectory.name)_\(DateFormatter.fileNameDate.string(from: trajectory.startTime)).gpx"
        let fileURL = documentsPath.appendingPathComponent(fileName)
        
        do {
            try gpxContent.write(to: fileURL, atomically: true, encoding: .utf8)
            return fileURL
        } catch {
            print("❌ 导出GPX失败: \(error)")
            return nil
        }
    }
    
    /// 导出轨迹为JSON文件
    func exportTrajectoryAsJSON(_ trajectory: Trajectory) -> URL? {
        guard let jsonData = trajectory.exportToJSON() else { return nil }
        
        let documentsPath = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
        let fileName = "\(trajectory.name)_\(DateFormatter.fileNameDate.string(from: trajectory.startTime)).json"
        let fileURL = documentsPath.appendingPathComponent(fileName)
        
        do {
            try jsonData.write(to: fileURL)
            return fileURL
        } catch {
            print("❌ 导出JSON失败: \(error)")
            return nil
        }
    }
}

// MARK: - Statistics
extension DataService {
    
    /// 获取统计信息
    var statistics: TrackingStatistics {
        let totalDistance = trajectories.reduce(0) { $0 + $1.totalDistance }
        let totalDuration = trajectories.reduce(0) { $0 + $1.duration }
        let totalTrajectories = trajectories.count
        
        return TrackingStatistics(
            totalDistance: totalDistance,
            totalDuration: totalDuration,
            totalTrajectories: totalTrajectories,
            averageDistance: totalTrajectories > 0 ? totalDistance / Double(totalTrajectories) : 0,
            averageDuration: totalTrajectories > 0 ? totalDuration / Double(totalTrajectories) : 0
        )
    }
}

// MARK: - Supporting Types
enum SyncStatus {
    case idle
    case syncing
    case success
    case failed(Error)
}

struct TrackingStatistics {
    let totalDistance: Double
    let totalDuration: TimeInterval
    let totalTrajectories: Int
    let averageDistance: Double
    let averageDuration: TimeInterval
}

// MARK: - Trajectory Extension
extension Trajectory {
    var isSynced: Bool {
        // 这里可以添加同步状态的逻辑
        // 暂时返回false，表示需要同步
        return false
    }
}

// MARK: - DateFormatter Extension
private extension DateFormatter {
    static let fileNameDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd_HH-mm-ss"
        return formatter
    }()
}