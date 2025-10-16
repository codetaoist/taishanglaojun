import Foundation
import Network
import CryptoKit
import Combine

// MARK: - Data Sync Enums

public enum SyncDataType: UInt32, CaseIterable {
    case aiConversation = 0
    case bookmark = 1
    case project = 2
    case userPreference = 3
    case custom = 4
}

public enum SyncOperation: UInt32 {
    case create = 0
    case update = 1
    case delete = 2
    case batch = 3
}

public enum SyncStatus: UInt32 {
    case idle = 0
    case connecting = 1
    case authenticating = 2
    case syncing = 3
    case conflictResolution = 4
    case completed = 5
    case error = 6
    case offline = 7
}

public enum SyncConflictResolution: UInt32 {
    case manual = 0
    case localWins = 1
    case remoteWins = 2
    case merge = 3
    case latestTimestamp = 4
}

public enum SyncError: UInt32 {
    case none = 0
    case networkFailure = 1
    case authFailed = 2
    case protocolError = 3
    case dataCorruption = 4
    case conflictUnresolved = 5
    case storageFull = 6
    case permissionDenied = 7
    case invalidData = 8
    case versionMismatch = 9
    case timeout = 10
}

public enum SyncMessageType: UInt32 {
    case handshake = 0
    case auth = 1
    case data = 2
    case ack = 3
    case conflict = 4
    case resolution = 5
    case heartbeat = 6
    case status = 7
    case error = 8
    case complete = 9
}

// MARK: - Data Structures

public struct SyncConfiguration {
    var serverUrl: String = ""
    var serverPort: UInt16 = 8443
    var deviceId: String = ""
    var userId: String = ""
    var authToken: String = ""
    var enableEncryption: Bool = true
    var enableCompression: Bool = true
    var autoSyncEnabled: Bool = true
    var syncInterval: UInt32 = 30000 // milliseconds
    var connectionTimeout: UInt32 = 10000
    var maxBatchSize: UInt32 = 100
    var maxRetries: UInt32 = 3
    var conflictResolution: SyncConflictResolution = .latestTimestamp
    var localStoragePath: String = ""
    
    public init() {}
}

public struct SyncItem {
    var syncId: String = ""
    var dataType: SyncDataType = .custom
    var operation: SyncOperation = .create
    var timestamp: UInt64 = 0
    var version: UInt64 = 0
    var dataLength: UInt32 = 0
    var metadataLength: UInt32 = 0
    var checksum: UInt32 = 0
    var deviceId: String = ""
    var userId: String = ""
    
    public init() {}
}

public struct SyncData {
    var item: SyncItem = SyncItem()
    var data: Data = Data()
    var metadata: Data = Data()
    
    public init() {}
}

public struct SyncConflict {
    var conflictId: String = ""
    var syncId: String = ""
    var localItem: SyncItem = SyncItem()
    var remoteItem: SyncItem = SyncItem()
    var conflictType: UInt32 = 0
    var timestamp: UInt64 = 0
    var resolution: SyncConflictResolution = .manual
    var isResolved: Bool = false
    
    public init() {}
}

public struct SyncCollection {
    var collectionId: String = ""
    var dataType: SyncDataType = .custom
    var itemCount: UInt32 = 0
    var lastSyncTimestamp: UInt64 = 0
    var version: UInt64 = 0
    var isDirty: Bool = false
    
    public init() {}
}

// MARK: - Protocol Messages

struct SyncHeader {
    let magic: UInt32 = 0x53594E43 // "SYNC"
    let version: UInt32 = 1
    let messageType: SyncMessageType
    let messageId: UInt32
    let sessionId: UInt32
    let dataLength: UInt32
    let checksum: UInt32
    let timestamp: UInt64
}

struct SyncHandshakeRequest {
    let deviceId: String
    let deviceName: String
    let protocolVersion: UInt32
    let supportedDataTypes: UInt32
    let supportsEncryption: Bool
    let supportsCompression: Bool
    let maxBatchSize: UInt32
}

struct SyncHandshakeResponse {
    let handshakeAccepted: Bool
    let sessionId: UInt32
    let serverVersion: UInt32
    let supportedDataTypes: UInt32
    let maxBatchSize: UInt32
    let errorCode: SyncError
}

struct SyncAuthRequest {
    let userId: String
    let authToken: String
    let deviceSignature: String
    let timestamp: UInt64
}

struct SyncAuthResponse {
    let authSuccess: Bool
    let sessionToken: String
    let userPermissions: UInt32
    let errorCode: SyncError
}

// MARK: - Callback Type Aliases

public typealias SyncStatusCallback = (SyncStatus, Float) -> Void
public typealias SyncDataCallback = (SyncData, SyncOperation) -> Void
public typealias SyncConflictCallback = (SyncConflict) -> Void
public typealias SyncErrorCallback = (SyncError, String) -> Void
public typealias SyncCompleteCallback = (UInt32, UInt32) -> Void

public typealias StoreItemCallback = (SyncData) -> Bool
public typealias RetrieveItemCallback = (String) -> SyncData?
public typealias DeleteItemCallback = (String) -> Bool
public typealias ListItemsCallback = (SyncDataType) -> [SyncItem]
public typealias UpdateCollectionCallback = (SyncCollection) -> Bool

// MARK: - Data Sync Manager

@available(macOS 10.15, *)
public class DataSyncManager: ObservableObject {
    
    // MARK: - Properties
    
    private var configuration: SyncConfiguration
    @Published public private(set) var status: SyncStatus = .idle
    @Published public private(set) var isConnected: Bool = false
    @Published public private(set) var progress: Float = 0.0
    
    private var isRunning: Bool = false
    private var sessionId: UInt32 = 0
    private var sessionToken: String = ""
    
    // Collections and sync state
    private var collections: [SyncCollection] = []
    private var lastSyncTimestamp: UInt64 = 0
    private var pendingItems: UInt32 = 0
    private var syncedItems: UInt32 = 0
    private var failedItems: UInt32 = 0
    
    // Conflicts
    private var activeConflicts: [SyncConflict] = []
    
    // Network
    private var connection: NWConnection?
    private var listener: NWListener?
    private let queue = DispatchQueue(label: "com.taishanglaojun.datasync", qos: .utility)
    
    // Threading
    private var syncTimer: Timer?
    private var heartbeatTimer: Timer?
    
    // Callbacks
    private var statusCallback: SyncStatusCallback?
    private var dataCallback: SyncDataCallback?
    private var conflictCallback: SyncConflictCallback?
    private var errorCallback: SyncErrorCallback?
    private var completeCallback: SyncCompleteCallback?
    
    // Storage interface
    private var storeItemCallback: StoreItemCallback?
    private var retrieveItemCallback: RetrieveItemCallback?
    private var deleteItemCallback: DeleteItemCallback?
    private var listItemsCallback: ListItemsCallback?
    private var updateCollectionCallback: UpdateCollectionCallback?
    
    // Local storage
    private var localCache: [String: SyncData] = [:]
    private var storagePath: String
    
    // MARK: - Initialization
    
    public init(configuration: SyncConfiguration) {
        self.configuration = configuration
        
        // Setup storage path
        if configuration.localStoragePath.isEmpty {
            let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, 
                                                    in: .userDomainMask).first!
            self.storagePath = appSupport.appendingPathComponent("TaiShangLaoJun/DataSync").path
        } else {
            self.storagePath = configuration.localStoragePath
        }
        
        // Create storage directory
        try? FileManager.default.createDirectory(atPath: storagePath, 
                                               withIntermediateDirectories: true, 
                                               attributes: nil)
        
        print("macOS Data Sync Manager created")
    }
    
    deinit {
        stop()
        print("macOS Data Sync Manager destroyed")
    }
    
    // MARK: - Manager Lifecycle
    
    public func start() -> Bool {
        guard !isRunning else { return true }
        
        // Load local collections
        loadCollections()
        
        // Start sync timer if auto-sync is enabled
        if configuration.autoSyncEnabled {
            startSyncTimer()
        }
        
        isRunning = true
        status = .idle
        
        print("Data sync manager started")
        return true
    }
    
    public func stop() {
        guard isRunning else { return }
        
        // Stop timers
        syncTimer?.invalidate()
        syncTimer = nil
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
        
        // Disconnect if connected
        if isConnected {
            disconnect()
        }
        
        isRunning = false
        
        print("Data sync manager stopped")
    }
    
    // MARK: - Connection Management
    
    public func connect() -> Bool {
        guard !isConnected else { return true }
        
        status = .connecting
        notifyStatusChange()
        
        // Create connection
        let host = NWEndpoint.Host(configuration.serverUrl)
        let port = NWEndpoint.Port(rawValue: configuration.serverPort)!
        let endpoint = NWEndpoint.hostPort(host: host, port: port)
        
        let parameters = NWParameters.tls
        parameters.requiredInterfaceType = .wifi
        
        connection = NWConnection(to: endpoint, using: parameters)
        
        connection?.stateUpdateHandler = { [weak self] state in
            DispatchQueue.main.async {
                self?.handleConnectionState(state)
            }
        }
        
        connection?.start(queue: queue)
        
        // Wait for connection with timeout
        let semaphore = DispatchSemaphore(value: 0)
        var connectionResult = false
        
        DispatchQueue.global().asyncAfter(deadline: .now() + .milliseconds(Int(configuration.connectionTimeout))) {
            if self.connection?.state == .ready {
                connectionResult = true
            }
            semaphore.signal()
        }
        
        semaphore.wait()
        
        if connectionResult {
            // Perform handshake
            if performHandshake() && authenticate() {
                isConnected = true
                status = .idle
                notifyStatusChange()
                startHeartbeat()
                print("Connected to sync server")
                return true
            }
        }
        
        disconnect()
        return false
    }
    
    public func disconnect() {
        connection?.cancel()
        connection = nil
        
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
        
        isConnected = false
        sessionId = 0
        sessionToken = ""
        
        status = .offline
        notifyStatusChange()
        
        print("Disconnected from sync server")
    }
    
    // MARK: - Sync Operations
    
    public func syncAll() -> Bool {
        guard isConnected || connect() else { return false }
        
        status = .syncing
        notifyStatusChange()
        
        var success = true
        
        // Sync each collection
        for collection in collections {
            if !syncCollection(collection.dataType) {
                success = false
            }
        }
        
        status = success ? .completed : .error
        notifyStatusChange()
        
        completeCallback?(syncedItems, failedItems)
        
        return success
    }
    
    public func syncCollection(_ type: SyncDataType) -> Bool {
        guard let items = listItemsCallback?(type) else { return false }
        
        // Send items in batches
        let batchSize = Int(configuration.maxBatchSize)
        let totalBatches = (items.count + batchSize - 1) / batchSize
        
        for batch in 0..<totalBatches {
            let startIdx = batch * batchSize
            let endIdx = min(startIdx + batchSize, items.count)
            let batchItems = Array(items[startIdx..<endIdx])
            
            if !sendBatch(type: type, items: batchItems, batchNum: UInt32(batch), totalBatches: UInt32(totalBatches)) {
                return false
            }
        }
        
        return true
    }
    
    // MARK: - Data Operations
    
    public func addItem(_ data: SyncData) -> Bool {
        // Store locally
        if let storeCallback = storeItemCallback, !storeCallback(data) {
            return false
        }
        
        // Add to local cache
        localCache[data.item.syncId] = data
        
        // Mark collection as dirty
        markCollectionDirty(data.item.dataType)
        
        // Sync immediately if auto-sync is enabled and connected
        if configuration.autoSyncEnabled && isConnected {
            DispatchQueue.global(qos: .utility).async {
                _ = self.syncCollection(data.item.dataType)
            }
        }
        
        return true
    }
    
    public func updateItem(_ data: SyncData) -> Bool {
        // Update locally
        if let storeCallback = storeItemCallback, !storeCallback(data) {
            return false
        }
        
        // Update local cache
        localCache[data.item.syncId] = data
        
        // Mark collection as dirty
        markCollectionDirty(data.item.dataType)
        
        // Sync immediately if auto-sync is enabled and connected
        if configuration.autoSyncEnabled && isConnected {
            DispatchQueue.global(qos: .utility).async {
                _ = self.syncCollection(data.item.dataType)
            }
        }
        
        return true
    }
    
    public func deleteItem(_ syncId: String) -> Bool {
        // Delete locally
        if let deleteCallback = deleteItemCallback, !deleteCallback(syncId) {
            return false
        }
        
        // Remove from local cache
        if let data = localCache.removeValue(forKey: syncId) {
            // Mark collection as dirty
            markCollectionDirty(data.item.dataType)
        }
        
        return true
    }
    
    public func getItem(_ syncId: String) -> SyncData? {
        // Check local cache first
        if let data = localCache[syncId] {
            return data
        }
        
        // Try storage interface
        return retrieveItemCallback?(syncId)
    }
    
    // MARK: - Status and Monitoring
    
    public func getStats() -> (synced: UInt32, pending: UInt32, failed: UInt32) {
        return (syncedItems, pendingItems, failedItems)
    }
    
    // MARK: - Callback Setters
    
    public func setStatusCallback(_ callback: @escaping SyncStatusCallback) {
        statusCallback = callback
    }
    
    public func setDataCallback(_ callback: @escaping SyncDataCallback) {
        dataCallback = callback
    }
    
    public func setConflictCallback(_ callback: @escaping SyncConflictCallback) {
        conflictCallback = callback
    }
    
    public func setErrorCallback(_ callback: @escaping SyncErrorCallback) {
        errorCallback = callback
    }
    
    public func setCompleteCallback(_ callback: @escaping SyncCompleteCallback) {
        completeCallback = callback
    }
    
    // MARK: - Storage Interface Setters
    
    public func setStorageInterface(
        storeItem: @escaping StoreItemCallback,
        retrieveItem: @escaping RetrieveItemCallback,
        deleteItem: @escaping DeleteItemCallback,
        listItems: @escaping ListItemsCallback,
        updateCollection: @escaping UpdateCollectionCallback
    ) {
        storeItemCallback = storeItem
        retrieveItemCallback = retrieveItem
        deleteItemCallback = deleteItem
        listItemsCallback = listItems
        updateCollectionCallback = updateCollection
    }
    
    // MARK: - Private Methods
    
    private func handleConnectionState(_ state: NWConnection.State) {
        switch state {
        case .ready:
            print("Connection ready")
        case .failed(let error):
            print("Connection failed: \(error)")
            handleError(.networkFailure, "Connection failed: \(error.localizedDescription)")
        case .cancelled:
            print("Connection cancelled")
        default:
            break
        }
    }
    
    private func performHandshake() -> Bool {
        let request = SyncHandshakeRequest(
            deviceId: configuration.deviceId,
            deviceName: "macOS Desktop",
            protocolVersion: 1,
            supportedDataTypes: 0xFFFFFFFF,
            supportsEncryption: configuration.enableEncryption,
            supportsCompression: configuration.enableCompression,
            maxBatchSize: configuration.maxBatchSize
        )
        
        let header = SyncHeader(
            messageType: .handshake,
            messageId: generateMessageId(),
            sessionId: 0,
            dataLength: 0, // Will be calculated
            checksum: 0,   // Will be calculated
            timestamp: getCurrentTimestamp()
        )
        
        guard sendMessage(header: header, data: request) else {
            handleError(.protocolError, "Failed to send handshake request")
            return false
        }
        
        // Receive response
        guard let (responseHeader, responseData): (SyncHeader, SyncHandshakeResponse) = receiveMessage() else {
            handleError(.protocolError, "Failed to receive handshake response")
            return false
        }
        
        guard responseHeader.messageType == .handshake else {
            handleError(.protocolError, "Invalid handshake response")
            return false
        }
        
        guard responseData.handshakeAccepted else {
            handleError(responseData.errorCode, "Handshake rejected")
            return false
        }
        
        // Update configuration based on server capabilities
        configuration.maxBatchSize = min(configuration.maxBatchSize, responseData.maxBatchSize)
        
        return true
    }
    
    private func authenticate() -> Bool {
        status = .authenticating
        notifyStatusChange()
        
        let request = SyncAuthRequest(
            userId: configuration.userId,
            authToken: configuration.authToken,
            deviceSignature: generateDeviceSignature(),
            timestamp: getCurrentTimestamp()
        )
        
        let header = SyncHeader(
            messageType: .auth,
            messageId: generateMessageId(),
            sessionId: 0,
            dataLength: 0,
            checksum: 0,
            timestamp: getCurrentTimestamp()
        )
        
        guard sendMessage(header: header, data: request) else {
            handleError(.authFailed, "Failed to send auth request")
            return false
        }
        
        // Receive response
        guard let (responseHeader, responseData): (SyncHeader, SyncAuthResponse) = receiveMessage() else {
            handleError(.authFailed, "Failed to receive auth response")
            return false
        }
        
        guard responseHeader.messageType == .auth else {
            handleError(.protocolError, "Invalid auth response")
            return false
        }
        
        guard responseData.authSuccess else {
            handleError(responseData.errorCode, "Authentication failed")
            return false
        }
        
        // Store session info
        sessionId = responseHeader.sessionId
        sessionToken = responseData.sessionToken
        
        return true
    }
    
    private func sendBatch(type: SyncDataType, items: [SyncItem], batchNum: UInt32, totalBatches: UInt32) -> Bool {
        // Create batch data
        var batchData = Data()
        
        // Add batch header
        let batchHeader = [
            "batch_id": generateMessageId(),
            "item_count": items.count,
            "total_batches": totalBatches,
            "current_batch": batchNum,
            "data_type": type.rawValue,
            "is_final_batch": batchNum == totalBatches - 1
        ]
        
        if let headerData = try? JSONSerialization.data(withJSONObject: batchHeader) {
            batchData.append(headerData)
        }
        
        // Add items data
        for item in items {
            if let itemData = getItem(item.syncId) {
                if let jsonData = try? JSONEncoder().encode(itemData) {
                    batchData.append(jsonData)
                }
            }
        }
        
        let header = SyncHeader(
            messageType: .data,
            messageId: generateMessageId(),
            sessionId: sessionId,
            dataLength: UInt32(batchData.count),
            checksum: calculateChecksum(batchData),
            timestamp: getCurrentTimestamp()
        )
        
        guard sendMessage(header: header, data: batchData) else {
            return false
        }
        
        // Wait for acknowledgment
        guard let (ackHeader, ackData): (SyncHeader, [String: Any]) = receiveMessage() else {
            return false
        }
        
        guard ackHeader.messageType == .ack else {
            return false
        }
        
        if let processedItems = ackData["processed_items"] as? UInt32,
           let failedItemsCount = ackData["failed_items"] as? UInt32 {
            syncedItems += processedItems
            failedItems += failedItemsCount
        }
        
        return ackData["batch_complete"] as? Bool ?? false
    }
    
    private func sendMessage<T: Codable>(header: SyncHeader, data: T) -> Bool {
        guard let connection = connection else { return false }
        
        do {
            // Encode data
            let jsonData = try JSONEncoder().encode(data)
            
            // Create header with correct data length and checksum
            var mutableHeader = header
            mutableHeader = SyncHeader(
                messageType: header.messageType,
                messageId: header.messageId,
                sessionId: header.sessionId,
                dataLength: UInt32(jsonData.count),
                checksum: calculateChecksum(jsonData),
                timestamp: header.timestamp
            )
            
            // Send header
            let headerData = withUnsafeBytes(of: mutableHeader) { Data($0) }
            
            let semaphore = DispatchSemaphore(value: 0)
            var success = false
            
            connection.send(content: headerData, completion: .contentProcessed { error in
                if error == nil {
                    // Send data
                    connection.send(content: jsonData, completion: .contentProcessed { error in
                        success = (error == nil)
                        semaphore.signal()
                    })
                } else {
                    semaphore.signal()
                }
            })
            
            semaphore.wait()
            return success
            
        } catch {
            return false
        }
    }
    
    private func receiveMessage<T: Codable>() -> (SyncHeader, T)? {
        guard let connection = connection else { return nil }
        
        let semaphore = DispatchSemaphore(value: 0)
        var result: (SyncHeader, T)? = nil
        
        // Receive header
        connection.receive(minimumIncompleteLength: MemoryLayout<SyncHeader>.size,
                          maximumLength: MemoryLayout<SyncHeader>.size) { data, _, _, error in
            guard let headerData = data, error == nil else {
                semaphore.signal()
                return
            }
            
            let header = headerData.withUnsafeBytes { $0.load(as: SyncHeader.self) }
            
            // Validate header
            guard header.magic == 0x53594E43, header.version == 1 else {
                semaphore.signal()
                return
            }
            
            // Receive data if present
            if header.dataLength > 0 {
                connection.receive(minimumIncompleteLength: Int(header.dataLength),
                                  maximumLength: Int(header.dataLength)) { data, _, _, error in
                    guard let messageData = data, error == nil else {
                        semaphore.signal()
                        return
                    }
                    
                    // Verify checksum
                    let calculatedChecksum = self.calculateChecksum(messageData)
                    guard calculatedChecksum == header.checksum else {
                        semaphore.signal()
                        return
                    }
                    
                    // Decode data
                    do {
                        let decodedData = try JSONDecoder().decode(T.self, from: messageData)
                        result = (header, decodedData)
                    } catch {
                        print("Failed to decode message data: \(error)")
                    }
                    
                    semaphore.signal()
                }
            } else {
                semaphore.signal()
            }
        }
        
        semaphore.wait()
        return result
    }
    
    private func startSyncTimer() {
        syncTimer = Timer.scheduledTimer(withTimeInterval: TimeInterval(configuration.syncInterval) / 1000.0, repeats: true) { _ in
            if self.isConnected {
                DispatchQueue.global(qos: .utility).async {
                    _ = self.syncAll()
                }
            }
        }
    }
    
    private func startHeartbeat() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: 30.0, repeats: true) { _ in
            self.sendHeartbeat()
        }
    }
    
    private func sendHeartbeat() {
        let header = SyncHeader(
            messageType: .heartbeat,
            messageId: generateMessageId(),
            sessionId: sessionId,
            dataLength: 0,
            checksum: 0,
            timestamp: getCurrentTimestamp()
        )
        
        if !sendMessage(header: header, data: EmptyData()) {
            // Heartbeat failed, disconnect
            disconnect()
        }
    }
    
    private func loadCollections() {
        let collectionsFile = URL(fileURLWithPath: storagePath).appendingPathComponent("collections.json")
        
        guard let data = try? Data(contentsOf: collectionsFile),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let collectionsArray = json["collections"] as? [[String: Any]] else {
            return
        }
        
        collections.removeAll()
        
        for item in collectionsArray {
            var collection = SyncCollection()
            collection.collectionId = item["id"] as? String ?? ""
            collection.dataType = SyncDataType(rawValue: item["type"] as? UInt32 ?? 0) ?? .custom
            collection.itemCount = item["count"] as? UInt32 ?? 0
            collection.lastSyncTimestamp = item["last_sync"] as? UInt64 ?? 0
            collection.version = item["version"] as? UInt64 ?? 0
            collection.isDirty = item["dirty"] as? Bool ?? false
            
            collections.append(collection)
        }
    }
    
    private func saveCollections() {
        let collectionsArray = collections.map { collection in
            return [
                "id": collection.collectionId,
                "type": collection.dataType.rawValue,
                "count": collection.itemCount,
                "last_sync": collection.lastSyncTimestamp,
                "version": collection.version,
                "dirty": collection.isDirty
            ]
        }
        
        let json = ["collections": collectionsArray]
        
        guard let data = try? JSONSerialization.data(withJSONObject: json, options: .prettyPrinted) else {
            return
        }
        
        let collectionsFile = URL(fileURLWithPath: storagePath).appendingPathComponent("collections.json")
        try? data.write(to: collectionsFile)
    }
    
    private func markCollectionDirty(_ type: SyncDataType) {
        for i in 0..<collections.count {
            if collections[i].dataType == type {
                collections[i].isDirty = true
                break
            }
        }
        saveCollections()
    }
    
    private func notifyStatusChange() {
        DispatchQueue.main.async {
            self.statusCallback?(self.status, self.progress)
        }
    }
    
    private func handleError(_ error: SyncError, _ message: String) {
        status = .error
        
        DispatchQueue.main.async {
            self.errorCallback?(error, message)
        }
        
        print("Sync error: \(message)")
    }
    
    private func generateMessageId() -> UInt32 {
        return UInt32.random(in: 1...UInt32.max)
    }
    
    private func getCurrentTimestamp() -> UInt64 {
        return UInt64(Date().timeIntervalSince1970 * 1000)
    }
    
    private func calculateChecksum(_ data: Data) -> UInt32 {
        var checksum: UInt32 = 0
        
        data.withUnsafeBytes { bytes in
            for byte in bytes {
                checksum = (checksum << 1) ^ UInt32(byte)
            }
        }
        
        return checksum
    }
    
    private func generateDeviceSignature() -> String {
        let signatureData = configuration.deviceId + String(getCurrentTimestamp())
        return String(signatureData.hash)
    }
}

// MARK: - Helper Structures

private struct EmptyData: Codable {}

// MARK: - Utility Extensions

extension SyncError {
    public var description: String {
        switch self {
        case .none: return "No error"
        case .networkFailure: return "Network failure"
        case .authFailed: return "Authentication failed"
        case .protocolError: return "Protocol error"
        case .dataCorruption: return "Data corruption"
        case .conflictUnresolved: return "Conflict unresolved"
        case .storageFull: return "Storage full"
        case .permissionDenied: return "Permission denied"
        case .invalidData: return "Invalid data"
        case .versionMismatch: return "Version mismatch"
        case .timeout: return "Timeout"
        }
    }
}

extension SyncStatus {
    public var description: String {
        switch self {
        case .idle: return "Idle"
        case .connecting: return "Connecting"
        case .authenticating: return "Authenticating"
        case .syncing: return "Syncing"
        case .conflictResolution: return "Resolving conflicts"
        case .completed: return "Completed"
        case .error: return "Error"
        case .offline: return "Offline"
        }
    }
}

// MARK: - Codable Conformance

extension SyncData: Codable {}
extension SyncItem: Codable {}
extension SyncConflict: Codable {}
extension SyncCollection: Codable {}