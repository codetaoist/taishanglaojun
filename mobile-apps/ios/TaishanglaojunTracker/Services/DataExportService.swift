//
//  DataExportService.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import Combine
import CoreData
import ZipArchive

/// 数据导出服务
class DataExportService: ObservableObject {
    
    // MARK: - Singleton
    static let shared = DataExportService()
    
    // MARK: - Published Properties
    @Published var isExporting = false
    @Published var exportProgress: Double = 0.0
    @Published var exportError: String?
    @Published var lastExportURL: URL?
    
    // MARK: - Private Properties
    private let dataService: DataService
    private let context: NSManagedObjectContext
    private let documentsDirectory = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
    private var cancellables = Set<AnyCancellable>()
    
    // MARK: - Initialization
    private init(dataService: DataService = DataService.shared) {
        self.dataService = dataService
        self.context = dataService.persistentContainer.viewContext
    }
    
    // MARK: - Export Methods
    
    /// 导出所有数据
    func exportAllData() -> AnyPublisher<URL, Error> {
        return Publishers.CombineLatest3(
            exportChatData(),
            exportTrajectoryData(),
            exportUserSettings()
        )
        .flatMap { [weak self] chatURL, trajectoryURL, settingsURL -> AnyPublisher<URL, Error> in
            guard let self = self else {
                return Fail(error: DataExportError.serviceUnavailable)
                    .eraseToAnyPublisher()
            }
            
            return self.createZipArchive(files: [
                ("chat_data.json", chatURL),
                ("trajectory_data.json", trajectoryURL),
                ("user_settings.json", settingsURL)
            ])
        }
        .handleEvents(
            receiveSubscription: { [weak self] _ in
                self?.isExporting = true
                self?.exportProgress = 0.0
            },
            receiveCompletion: { [weak self] completion in
                self?.isExporting = false
                if case .failure(let error) = completion {
                    self?.exportError = error.localizedDescription
                }
            },
            receiveCancel: { [weak self] in
                self?.isExporting = false
            }
        )
        .eraseToAnyPublisher()
    }
    
    /// 导出聊天数据
    func exportChatData() -> AnyPublisher<URL, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            self.exportProgress = 0.1
            
            do {
                // 获取所有对话
                let conversationRequest: NSFetchRequest<ConversationEntity> = ConversationEntity.fetchRequest()
                let conversations = try self.context.fetch(conversationRequest)
                
                self.exportProgress = 0.3
                
                // 获取所有消息
                let messageRequest: NSFetchRequest<ChatMessageEntity> = ChatMessageEntity.fetchRequest()
                let messages = try self.context.fetch(messageRequest)
                
                self.exportProgress = 0.5
                
                // 构建导出数据
                let exportData = ChatExportData(
                    conversations: conversations.compactMap { $0.toConversation() },
                    messages: messages.compactMap { $0.toChatMessage() },
                    exportDate: Date(),
                    version: "1.0"
                )
                
                self.exportProgress = 0.7
                
                // 序列化为JSON
                let jsonData = try JSONEncoder().encode(exportData)
                
                // 保存到文件
                let fileName = "chat_export_\(self.dateFormatter.string(from: Date())).json"
                let fileURL = self.documentsDirectory.appendingPathComponent(fileName)
                try jsonData.write(to: fileURL)
                
                self.exportProgress = 1.0
                self.lastExportURL = fileURL
                
                promise(.success(fileURL))
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    /// 导出轨迹数据
    func exportTrajectoryData() -> AnyPublisher<URL, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            do {
                // 获取所有轨迹数据
                let trajectoryRequest: NSFetchRequest<TrajectoryEntity> = TrajectoryEntity.fetchRequest()
                let trajectories = try self.context.fetch(trajectoryRequest)
                
                // 获取所有位置点
                let locationRequest: NSFetchRequest<LocationEntity> = LocationEntity.fetchRequest()
                let locations = try self.context.fetch(locationRequest)
                
                // 构建导出数据
                let exportData = TrajectoryExportData(
                    trajectories: trajectories.compactMap { $0.toTrajectory() },
                    locations: locations.compactMap { $0.toLocation() },
                    exportDate: Date(),
                    version: "1.0"
                )
                
                // 序列化为JSON
                let jsonData = try JSONEncoder().encode(exportData)
                
                // 保存到文件
                let fileName = "trajectory_export_\(self.dateFormatter.string(from: Date())).json"
                let fileURL = self.documentsDirectory.appendingPathComponent(fileName)
                try jsonData.write(to: fileURL)
                
                promise(.success(fileURL))
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    /// 导出用户设置
    func exportUserSettings() -> AnyPublisher<URL, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            do {
                // 获取用户设置
                let settings = UserDefaults.standard.dictionaryRepresentation()
                
                // 过滤敏感信息
                let filteredSettings = settings.filter { key, _ in
                    !key.contains("password") &&
                    !key.contains("token") &&
                    !key.contains("secret") &&
                    !key.contains("key")
                }
                
                let exportData = UserSettingsExportData(
                    settings: filteredSettings,
                    exportDate: Date(),
                    version: "1.0"
                )
                
                // 序列化为JSON
                let jsonData = try JSONEncoder().encode(exportData)
                
                // 保存到文件
                let fileName = "settings_export_\(self.dateFormatter.string(from: Date())).json"
                let fileURL = self.documentsDirectory.appendingPathComponent(fileName)
                try jsonData.write(to: fileURL)
                
                promise(.success(fileURL))
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    /// 创建ZIP压缩包
    private func createZipArchive(files: [(String, URL)]) -> AnyPublisher<URL, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            do {
                let zipFileName = "taishanglaojun_export_\(self.dateFormatter.string(from: Date())).zip"
                let zipURL = self.documentsDirectory.appendingPathComponent(zipFileName)
                
                // 删除已存在的文件
                if FileManager.default.fileExists(atPath: zipURL.path) {
                    try FileManager.default.removeItem(at: zipURL)
                }
                
                // 创建ZIP文件
                let success = SSZipArchive.createZipFile(atPath: zipURL.path, withFilesAtPaths: files.map { $0.1.path })
                
                if success {
                    // 清理临时文件
                    for (_, fileURL) in files {
                        try? FileManager.default.removeItem(at: fileURL)
                    }
                    
                    self.lastExportURL = zipURL
                    promise(.success(zipURL))
                } else {
                    promise(.failure(DataExportError.zipCreationFailed))
                }
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    // MARK: - Import Methods
    
    /// 导入数据
    func importData(from url: URL) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            do {
                let data = try Data(contentsOf: url)
                
                // 尝试解析为聊天数据
                if let chatData = try? JSONDecoder().decode(ChatExportData.self, from: data) {
                    self.importChatData(chatData)
                        .sink(
                            receiveCompletion: { completion in
                                if case .failure(let error) = completion {
                                    promise(.failure(error))
                                } else {
                                    promise(.success(()))
                                }
                            },
                            receiveValue: { _ in }
                        )
                        .store(in: &self.cancellables)
                    return
                }
                
                // 尝试解析为轨迹数据
                if let trajectoryData = try? JSONDecoder().decode(TrajectoryExportData.self, from: data) {
                    self.importTrajectoryData(trajectoryData)
                        .sink(
                            receiveCompletion: { completion in
                                if case .failure(let error) = completion {
                                    promise(.failure(error))
                                } else {
                                    promise(.success(()))
                                }
                            },
                            receiveValue: { _ in }
                        )
                        .store(in: &self.cancellables)
                    return
                }
                
                promise(.failure(DataExportError.unsupportedFormat))
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    /// 导入聊天数据
    private func importChatData(_ data: ChatExportData) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            do {
                // 导入对话
                for conversation in data.conversations {
                    let entity = ConversationEntity(context: self.context)
                    entity.id = conversation.id
                    entity.title = conversation.title
                    entity.aiPersonality = conversation.aiPersonality.rawValue
                    entity.createdAt = conversation.createdAt
                    entity.updatedAt = conversation.updatedAt
                    entity.isArchived = conversation.isArchived
                }
                
                // 导入消息
                for message in data.messages {
                    let entity = ChatMessageEntity(context: self.context)
                    entity.id = message.id
                    entity.conversationId = message.conversationId
                    entity.content = message.content
                    entity.messageType = message.messageType.rawValue
                    entity.sender = message.sender.rawValue
                    entity.timestamp = message.timestamp
                    entity.status = message.status.rawValue
                    // metadata处理
                    if let metadata = message.metadata {
                        entity.metadata = try? JSONSerialization.data(withJSONObject: metadata)
                    }
                }
                
                try self.context.save()
                promise(.success(()))
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    /// 导入轨迹数据
    private func importTrajectoryData(_ data: TrajectoryExportData) -> AnyPublisher<Void, Error> {
        return Future { [weak self] promise in
            guard let self = self else {
                promise(.failure(DataExportError.serviceUnavailable))
                return
            }
            
            do {
                // 导入轨迹
                for trajectory in data.trajectories {
                    let entity = TrajectoryEntity(context: self.context)
                    entity.id = trajectory.id
                    entity.name = trajectory.name
                    entity.startTime = trajectory.startTime
                    entity.endTime = trajectory.endTime
                    entity.distance = trajectory.distance
                    entity.duration = trajectory.duration
                }
                
                // 导入位置点
                for location in data.locations {
                    let entity = LocationEntity(context: self.context)
                    entity.id = location.id
                    entity.latitude = location.latitude
                    entity.longitude = location.longitude
                    entity.altitude = location.altitude
                    entity.timestamp = location.timestamp
                    entity.accuracy = location.accuracy
                    entity.speed = location.speed
                    entity.course = location.course
                }
                
                try self.context.save()
                promise(.success(()))
                
            } catch {
                promise(.failure(error))
            }
        }
        .eraseToAnyPublisher()
    }
    
    // MARK: - Utility Methods
    
    /// 清除导出错误
    func clearError() {
        exportError = nil
    }
    
    /// 获取导出文件大小
    func getExportFileSize(_ url: URL) -> Int64 {
        do {
            let attributes = try FileManager.default.attributesOfItem(atPath: url.path)
            return attributes[.size] as? Int64 ?? 0
        } catch {
            return 0
        }
    }
    
    /// 删除导出文件
    func deleteExportFile(_ url: URL) {
        do {
            try FileManager.default.removeItem(at: url)
            if lastExportURL == url {
                lastExportURL = nil
            }
        } catch {
            print("❌ 删除导出文件失败: \(error)")
        }
    }
    
    // MARK: - Private Properties
    private lazy var dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyyMMdd_HHmmss"
        return formatter
    }()
}

// MARK: - Export Data Models

/// 聊天数据导出模型
struct ChatExportData: Codable {
    let conversations: [Conversation]
    let messages: [ChatMessage]
    let exportDate: Date
    let version: String
}

/// 轨迹数据导出模型
struct TrajectoryExportData: Codable {
    let trajectories: [Trajectory]
    let locations: [Location]
    let exportDate: Date
    let version: String
}

/// 用户设置导出模型
struct UserSettingsExportData: Codable {
    let settings: [String: Any]
    let exportDate: Date
    let version: String
    
    enum CodingKeys: String, CodingKey {
        case settings, exportDate, version
    }
    
    init(settings: [String: Any], exportDate: Date, version: String) {
        self.settings = settings
        self.exportDate = exportDate
        self.version = version
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        self.settings = try container.decode([String: AnyCodable].self, forKey: .settings).mapValues { $0.value }
        self.exportDate = try container.decode(Date.self, forKey: .exportDate)
        self.version = try container.decode(String.self, forKey: .version)
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(settings.mapValues { AnyCodable($0) }, forKey: .settings)
        try container.encode(exportDate, forKey: .exportDate)
        try container.encode(version, forKey: .version)
    }
}

// MARK: - Data Export Errors
enum DataExportError: LocalizedError {
    case serviceUnavailable
    case exportFailed
    case importFailed
    case unsupportedFormat
    case zipCreationFailed
    case fileNotFound
    
    var errorDescription: String? {
        switch self {
        case .serviceUnavailable:
            return "数据导出服务不可用"
        case .exportFailed:
            return "数据导出失败"
        case .importFailed:
            return "数据导入失败"
        case .unsupportedFormat:
            return "不支持的文件格式"
        case .zipCreationFailed:
            return "ZIP文件创建失败"
        case .fileNotFound:
            return "文件未找到"
        }
    }
}

// MARK: - AnyCodable Helper
struct AnyCodable: Codable {
    let value: Any
    
    init(_ value: Any) {
        self.value = value
    }
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if let bool = try? container.decode(Bool.self) {
            value = bool
        } else if let int = try? container.decode(Int.self) {
            value = int
        } else if let double = try? container.decode(Double.self) {
            value = double
        } else if let string = try? container.decode(String.self) {
            value = string
        } else if let array = try? container.decode([AnyCodable].self) {
            value = array.map { $0.value }
        } else if let dictionary = try? container.decode([String: AnyCodable].self) {
            value = dictionary.mapValues { $0.value }
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "无法解码值")
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        
        switch value {
        case let bool as Bool:
            try container.encode(bool)
        case let int as Int:
            try container.encode(int)
        case let double as Double:
            try container.encode(double)
        case let string as String:
            try container.encode(string)
        case let array as [Any]:
            try container.encode(array.map { AnyCodable($0) })
        case let dictionary as [String: Any]:
            try container.encode(dictionary.mapValues { AnyCodable($0) })
        default:
            throw EncodingError.invalidValue(value, EncodingError.Context(codingPath: container.codingPath, debugDescription: "无法编码值"))
        }
    }
}