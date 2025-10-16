import Foundation
import Network
import CryptoKit
import SystemConfiguration

// MARK: - File Transfer Manager for macOS

@available(macOS 10.15, *)
class FileTransferManager: ObservableObject {
    // MARK: - Properties
    
    private let localDeviceId: String
    private let localDeviceName: String
    private let localDeviceType: DeviceType = .desktopMacOS
    
    @Published var isRunning: Bool = false
    @Published var discoveryEnabled: Bool = false
    @Published var encryptionEnabled: Bool = true
    @Published var discoveredDevices: [DeviceInfo] = []
    @Published var activeSessions: [FileTransferSession] = []
    
    private var listenPort: UInt16 = 8888
    private var maxChunkSize: UInt32 = 65536 // 64KB
    
    // Network components
    private var listener: NWListener?
    private var discoveryListener: NWListener?
    private var discoveryConnection: NWConnection?
    private var connections: [String: NWConnection] = [:]
    
    // Queues
    private let networkQueue = DispatchQueue(label: "com.taishanglaojun.filetransfer.network", qos: .userInitiated)
    private let discoveryQueue = DispatchQueue(label: "com.taishanglaojun.filetransfer.discovery", qos: .utility)
    private let transferQueue = DispatchQueue(label: "com.taishanglaojun.filetransfer.transfer", qos: .userInitiated)
    
    // Timers
    private var discoveryTimer: Timer?
    private var heartbeatTimer: Timer?
    
    // Callbacks
    var progressCallback: FileTransferProgressCallback?
    var completeCallback: FileTransferCompleteCallback?
    var errorCallback: FileTransferErrorCallback?
    var deviceDiscoveredCallback: DeviceDiscoveredCallback?
    var deviceConnectedCallback: DeviceConnectedCallback?
    var deviceDisconnectedCallback: DeviceDisconnectedCallback?
    var fileReceiveRequestCallback: FileReceiveRequestCallback?
    
    // MARK: - Initialization
    
    init(deviceName: String? = nil) {
        self.localDeviceId = Self.generateDeviceId()
        self.localDeviceName = deviceName ?? Self.getDefaultDeviceName()
        
        setupNetworkMonitoring()
    }
    
    deinit {
        stop()
    }
    
    // MARK: - Public Methods
    
    func start(port: UInt16 = 8888) -> Bool {
        guard !isRunning else { return true }
        
        listenPort = port
        
        do {
            // Start main listener
            try startMainListener()
            
            // Start discovery listener
            try startDiscoveryListener()
            
            isRunning = true
            
            // Start heartbeat timer
            startHeartbeatTimer()
            
            print("File transfer manager started on port \(port)")
            return true
            
        } catch {
            print("Failed to start file transfer manager: \(error)")
            stop()
            return false
        }
    }
    
    func stop() {
        guard isRunning else { return }
        
        isRunning = false
        discoveryEnabled = false
        
        // Stop timers
        discoveryTimer?.invalidate()
        discoveryTimer = nil
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
        
        // Close all connections
        connections.values.forEach { $0.cancel() }
        connections.removeAll()
        
        // Stop listeners
        listener?.cancel()
        listener = nil
        discoveryListener?.cancel()
        discoveryListener = nil
        discoveryConnection?.cancel()
        discoveryConnection = nil
        
        // Clear sessions
        activeSessions.removeAll()
        
        print("File transfer manager stopped")
    }
    
    func startDiscovery() -> Bool {
        guard isRunning else { return false }
        
        discoveryEnabled = true
        
        // Start discovery timer
        discoveryTimer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { [weak self] _ in
            self?.sendDiscoveryBroadcast()
        }
        
        // Send initial discovery broadcast
        sendDiscoveryBroadcast()
        
        print("Device discovery started")
        return true
    }
    
    func stopDiscovery() {
        discoveryEnabled = false
        discoveryTimer?.invalidate()
        discoveryTimer = nil
        
        print("Device discovery stopped")
    }
    
    func connectToDevice(_ device: DeviceInfo) -> UInt32? {
        guard isRunning else { return nil }
        
        let sessionId = generateSessionId()
        
        // Create connection
        let endpoint = NWEndpoint.hostPort(host: NWEndpoint.Host(device.ipAddress), port: NWEndpoint.Port(integerLiteral: device.port))
        let connection = NWConnection(to: endpoint, using: .tcp)
        
        connection.stateUpdateHandler = { [weak self] state in
            self?.handleConnectionStateChange(sessionId: sessionId, device: device, state: state)
        }
        
        connection.start(queue: networkQueue)
        connections[device.deviceId] = connection
        
        // Send connection request
        sendConnectionRequest(to: connection, sessionId: sessionId, device: device)
        
        return sessionId
    }
    
    func sendFile(sessionId: UInt32, filePath: String) -> UInt32? {
        guard let session = activeSessions.first(where: { $0.sessionId == sessionId }),
              session.status == .connected else {
            return nil
        }
        
        let transferId = generateTransferId()
        
        transferQueue.async { [weak self] in
            self?.performFileSend(sessionId: sessionId, transferId: transferId, filePath: filePath)
        }
        
        return transferId
    }
    
    func sendFiles(sessionId: UInt32, filePaths: [String]) -> [UInt32] {
        var transferIds: [UInt32] = []
        
        for filePath in filePaths {
            if let transferId = sendFile(sessionId: sessionId, filePath: filePath) {
                transferIds.append(transferId)
            }
        }
        
        return transferIds
    }
    
    func pauseTransfer(transferId: UInt32) -> Bool {
        // TODO: Implement transfer pause
        return false
    }
    
    func resumeTransfer(transferId: UInt32) -> Bool {
        // TODO: Implement transfer resume
        return false
    }
    
    func cancelTransfer(transferId: UInt32) -> Bool {
        // TODO: Implement transfer cancel
        return false
    }
    
    // MARK: - Private Methods
    
    private func startMainListener() throws {
        let parameters = NWParameters.tcp
        parameters.allowLocalEndpointReuse = true
        
        listener = try NWListener(using: parameters, on: NWEndpoint.Port(integerLiteral: listenPort))
        
        listener?.newConnectionHandler = { [weak self] connection in
            self?.handleNewConnection(connection)
        }
        
        listener?.stateUpdateHandler = { state in
            switch state {
            case .ready:
                print("Main listener ready on port \(self.listenPort)")
            case .failed(let error):
                print("Main listener failed: \(error)")
            default:
                break
            }
        }
        
        listener?.start(queue: networkQueue)
    }
    
    private func startDiscoveryListener() throws {
        let parameters = NWParameters.udp
        parameters.allowLocalEndpointReuse = true
        
        discoveryListener = try NWListener(using: parameters, on: NWEndpoint.Port(integerLiteral: 8889))
        
        discoveryListener?.newConnectionHandler = { [weak self] connection in
            self?.handleDiscoveryConnection(connection)
        }
        
        discoveryListener?.stateUpdateHandler = { state in
            switch state {
            case .ready:
                print("Discovery listener ready on port 8889")
            case .failed(let error):
                print("Discovery listener failed: \(error)")
            default:
                break
            }
        }
        
        discoveryListener?.start(queue: discoveryQueue)
    }
    
    private func handleNewConnection(_ connection: NWConnection) {
        connection.stateUpdateHandler = { [weak self] state in
            switch state {
            case .ready:
                self?.receiveMessage(from: connection)
            case .failed(let error):
                print("Connection failed: \(error)")
                connection.cancel()
            case .cancelled:
                print("Connection cancelled")
            default:
                break
            }
        }
        
        connection.start(queue: networkQueue)
    }
    
    private func handleDiscoveryConnection(_ connection: NWConnection) {
        connection.stateUpdateHandler = { [weak self] state in
            switch state {
            case .ready:
                self?.receiveDiscoveryMessage(from: connection)
            case .failed(let error):
                print("Discovery connection failed: \(error)")
                connection.cancel()
            default:
                break
            }
        }
        
        connection.start(queue: discoveryQueue)
    }
    
    private func sendDiscoveryBroadcast() {
        guard discoveryEnabled else { return }
        
        let request = DiscoveryRequest(
            deviceId: localDeviceId,
            deviceName: localDeviceName,
            deviceType: localDeviceType,
            listenPort: listenPort,
            supportsEncryption: encryptionEnabled,
            maxChunkSize: maxChunkSize
        )
        
        let header = FileTransferHeader(
            magic: FileTransferProtocol.magic,
            version: FileTransferProtocol.version,
            messageType: .discoveryRequest,
            messageId: generateMessageId(),
            sessionId: 0,
            dataLength: UInt32(MemoryLayout<DiscoveryRequest>.size),
            checksum: 0,
            timestamp: UInt64(Date().timeIntervalSince1970 * 1000)
        )
        
        // Create broadcast connection
        let endpoint = NWEndpoint.hostPort(host: "255.255.255.255", port: 8889)
        let parameters = NWParameters.udp
        parameters.allowLocalEndpointReuse = true
        
        let broadcastConnection = NWConnection(to: endpoint, using: parameters)
        
        broadcastConnection.stateUpdateHandler = { state in
            switch state {
            case .ready:
                self.sendMessage(header: header, data: request, to: broadcastConnection) { success in
                    broadcastConnection.cancel()
                }
            case .failed(let error):
                print("Broadcast connection failed: \(error)")
                broadcastConnection.cancel()
            default:
                break
            }
        }
        
        broadcastConnection.start(queue: discoveryQueue)
    }
    
    private func receiveDiscoveryMessage(from connection: NWConnection) {
        connection.receive(minimumIncompleteLength: MemoryLayout<FileTransferHeader>.size,
                          maximumLength: 1024) { [weak self] data, _, isComplete, error in
            
            if let error = error {
                print("Discovery receive error: \(error)")
                return
            }
            
            guard let data = data, !data.isEmpty else {
                if !isComplete {
                    self?.receiveDiscoveryMessage(from: connection)
                }
                return
            }
            
            self?.processDiscoveryMessage(data: data, from: connection)
            
            if !isComplete {
                self?.receiveDiscoveryMessage(from: connection)
            }
        }
    }
    
    private func processDiscoveryMessage(data: Data, from connection: NWConnection) {
        guard data.count >= MemoryLayout<FileTransferHeader>.size else { return }
        
        let header = data.withUnsafeBytes { $0.load(as: FileTransferHeader.self) }
        
        guard header.magic == FileTransferProtocol.magic,
              header.version == FileTransferProtocol.version else {
            return
        }
        
        let messageData = data.dropFirst(MemoryLayout<FileTransferHeader>.size)
        
        switch header.messageType {
        case .discoveryRequest:
            handleDiscoveryRequest(data: messageData, from: connection)
        case .discoveryResponse:
            handleDiscoveryResponse(data: messageData, from: connection)
        default:
            break
        }
    }
    
    private func handleDiscoveryRequest(data: Data, from connection: NWConnection) {
        guard data.count >= MemoryLayout<DiscoveryRequest>.size else { return }
        
        let request = data.withUnsafeBytes { $0.load(as: DiscoveryRequest.self) }
        
        // Don't respond to our own requests
        if String(cString: request.deviceId) == localDeviceId {
            return
        }
        
        // Send discovery response
        let response = DiscoveryResponse(
            deviceId: localDeviceId,
            deviceName: localDeviceName,
            deviceType: localDeviceType,
            listenPort: listenPort,
            supportsEncryption: encryptionEnabled,
            maxChunkSize: maxChunkSize,
            acceptsConnections: true
        )
        
        let header = FileTransferHeader(
            magic: FileTransferProtocol.magic,
            version: FileTransferProtocol.version,
            messageType: .discoveryResponse,
            messageId: generateMessageId(),
            sessionId: 0,
            dataLength: UInt32(MemoryLayout<DiscoveryResponse>.size),
            checksum: 0,
            timestamp: UInt64(Date().timeIntervalSince1970 * 1000)
        )
        
        sendMessage(header: header, data: response, to: connection) { _ in }
    }
    
    private func handleDiscoveryResponse(data: Data, from connection: NWConnection) {
        guard data.count >= MemoryLayout<DiscoveryResponse>.size else { return }
        
        let response = data.withUnsafeBytes { $0.load(as: DiscoveryResponse.self) }
        
        // Don't add our own device
        let deviceId = String(cString: response.deviceId)
        if deviceId == localDeviceId {
            return
        }
        
        // Extract IP address from connection
        guard let endpoint = connection.endpoint,
              case .hostPort(let host, _) = endpoint,
              case .ipv4(let ipv4Address) = host else {
            return
        }
        
        let device = DeviceInfo(
            deviceId: deviceId,
            deviceName: String(cString: response.deviceName),
            deviceType: response.deviceType,
            ipAddress: ipv4Address.rawValue.withUnsafeBytes { $0.load(as: UInt32.self) },
            port: response.listenPort,
            lastSeen: UInt64(Date().timeIntervalSince1970 * 1000),
            isTrusted: false,
            supportsEncryption: response.supportsEncryption,
            maxChunkSize: response.maxChunkSize
        )
        
        DispatchQueue.main.async {
            // Update or add device
            if let index = self.discoveredDevices.firstIndex(where: { $0.deviceId == deviceId }) {
                self.discoveredDevices[index] = device
            } else {
                self.discoveredDevices.append(device)
            }
            
            // Notify callback
            self.deviceDiscoveredCallback?(device)
        }
    }
    
    private func sendConnectionRequest(to connection: NWConnection, sessionId: UInt32, device: DeviceInfo) {
        let request = ConnectRequest(
            deviceId: localDeviceId,
            deviceName: localDeviceName,
            deviceType: localDeviceType,
            protocolVersion: FileTransferProtocol.version,
            requestEncryption: encryptionEnabled
        )
        
        let header = FileTransferHeader(
            magic: FileTransferProtocol.magic,
            version: FileTransferProtocol.version,
            messageType: .connectRequest,
            messageId: generateMessageId(),
            sessionId: sessionId,
            dataLength: UInt32(MemoryLayout<ConnectRequest>.size),
            checksum: 0,
            timestamp: UInt64(Date().timeIntervalSince1970 * 1000)
        )
        
        sendMessage(header: header, data: request, to: connection) { [weak self] success in
            if success {
                self?.receiveMessage(from: connection)
            } else {
                print("Failed to send connection request")
                connection.cancel()
            }
        }
    }
    
    private func receiveMessage(from connection: NWConnection) {
        connection.receive(minimumIncompleteLength: MemoryLayout<FileTransferHeader>.size,
                          maximumLength: MemoryLayout<FileTransferHeader>.size) { [weak self] data, _, isComplete, error in
            
            if let error = error {
                print("Receive error: \(error)")
                return
            }
            
            guard let data = data, data.count == MemoryLayout<FileTransferHeader>.size else {
                if !isComplete {
                    self?.receiveMessage(from: connection)
                }
                return
            }
            
            let header = data.withUnsafeBytes { $0.load(as: FileTransferHeader.self) }
            
            // Receive message data if present
            if header.dataLength > 0 {
                self?.receiveMessageData(from: connection, header: header)
            } else {
                self?.processMessage(header: header, data: Data(), from: connection)
            }
        }
    }
    
    private func receiveMessageData(from connection: NWConnection, header: FileTransferHeader) {
        connection.receive(minimumIncompleteLength: Int(header.dataLength),
                          maximumLength: Int(header.dataLength)) { [weak self] data, _, isComplete, error in
            
            if let error = error {
                print("Receive data error: \(error)")
                return
            }
            
            guard let data = data, data.count == header.dataLength else {
                if !isComplete {
                    self?.receiveMessageData(from: connection, header: header)
                }
                return
            }
            
            self?.processMessage(header: header, data: data, from: connection)
            
            if !isComplete {
                self?.receiveMessage(from: connection)
            }
        }
    }
    
    private func processMessage(header: FileTransferHeader, data: Data, from connection: NWConnection) {
        switch header.messageType {
        case .connectResponse:
            handleConnectResponse(header: header, data: data, from: connection)
        case .fileRequest:
            handleFileRequest(header: header, data: data, from: connection)
        case .fileResponse:
            handleFileResponse(header: header, data: data, from: connection)
        case .fileChunk:
            handleFileChunk(header: header, data: data, from: connection)
        case .fileAck:
            handleFileAck(header: header, data: data, from: connection)
        default:
            print("Unhandled message type: \(header.messageType)")
        }
    }
    
    private func handleConnectResponse(header: FileTransferHeader, data: Data, from connection: NWConnection) {
        guard data.count >= MemoryLayout<ConnectResponse>.size else { return }
        
        let response = data.withUnsafeBytes { $0.load(as: ConnectResponse.self) }
        
        if response.connectionAccepted {
            // Create session
            let session = FileTransferSession(
                sessionId: header.sessionId,
                sessionToken: String(cString: response.sessionToken),
                remoteDevice: DeviceInfo(), // TODO: Fill with actual device info
                fileInfo: FileInfo(),
                direction: .send,
                status: .connected,
                bytesTransferred: 0,
                totalBytes: 0,
                chunkSize: min(response.maxChunkSize, maxChunkSize),
                startTime: UInt64(Date().timeIntervalSince1970 * 1000),
                lastActivityTime: UInt64(Date().timeIntervalSince1970 * 1000),
                progressPercentage: 0.0,
                transferSpeed: 0.0,
                estimatedTimeRemaining: 0,
                lastError: .none
            )
            
            DispatchQueue.main.async {
                self.activeSessions.append(session)
                // TODO: Notify connection callback
            }
        } else {
            print("Connection rejected: \(response.errorCode)")
            connection.cancel()
        }
    }
    
    private func handleFileRequest(header: FileTransferHeader, data: Data, from connection: NWConnection) {
        // TODO: Implement file request handling
    }
    
    private func handleFileResponse(header: FileTransferHeader, data: Data, from connection: NWConnection) {
        // TODO: Implement file response handling
    }
    
    private func handleFileChunk(header: FileTransferHeader, data: Data, from connection: NWConnection) {
        // TODO: Implement file chunk handling
    }
    
    private func handleFileAck(header: FileTransferHeader, data: Data, from connection: NWConnection) {
        // TODO: Implement file acknowledgment handling
    }
    
    private func performFileSend(sessionId: UInt32, transferId: UInt32, filePath: String) {
        // TODO: Implement file sending logic
        print("Sending file: \(filePath) (Session: \(sessionId), Transfer: \(transferId))")
    }
    
    private func sendMessage<T>(header: FileTransferHeader, data: T, to connection: NWConnection, completion: @escaping (Bool) -> Void) {
        var mutableHeader = header
        
        // Calculate checksum
        let dataBytes = withUnsafeBytes(of: data) { Data($0) }
        mutableHeader.checksum = calculateChecksum(data: dataBytes)
        
        // Send header
        let headerData = withUnsafeBytes(of: mutableHeader) { Data($0) }
        
        connection.send(content: headerData, completion: .contentProcessed { error in
            if let error = error {
                print("Failed to send header: \(error)")
                completion(false)
                return
            }
            
            // Send data
            connection.send(content: dataBytes, completion: .contentProcessed { error in
                if let error = error {
                    print("Failed to send data: \(error)")
                    completion(false)
                } else {
                    completion(true)
                }
            })
        })
    }
    
    private func handleConnectionStateChange(sessionId: UInt32, device: DeviceInfo, state: NWConnection.State) {
        switch state {
        case .ready:
            print("Connected to device: \(device.deviceName)")
        case .failed(let error):
            print("Connection to \(device.deviceName) failed: \(error)")
            connections.removeValue(forKey: device.deviceId)
        case .cancelled:
            print("Connection to \(device.deviceName) cancelled")
            connections.removeValue(forKey: device.deviceId)
        default:
            break
        }
    }
    
    private func startHeartbeatTimer() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: 30.0, repeats: true) { [weak self] _ in
            self?.sendHeartbeats()
        }
    }
    
    private func sendHeartbeats() {
        // TODO: Implement heartbeat sending
    }
    
    private func setupNetworkMonitoring() {
        // TODO: Implement network change monitoring
    }
    
    // MARK: - Utility Methods
    
    private static func generateDeviceId() -> String {
        // Use system UUID or hardware identifier
        if let uuid = getSystemUUID() {
            return "MAC_\(uuid)"
        } else {
            return "MAC_\(UUID().uuidString)"
        }
    }
    
    private static func getSystemUUID() -> String? {
        let platformExpert = IOServiceGetMatchingService(kIOMasterPortDefault, IOServiceMatching("IOPlatformExpertDevice"))
        
        guard platformExpert > 0 else {
            return nil
        }
        
        defer {
            IOObjectRelease(platformExpert)
        }
        
        guard let serialNumberAsCFString = IORegistryEntryCreateCFProperty(platformExpert, kIOPlatformUUIDKey, kCFAllocatorDefault, 0) else {
            return nil
        }
        
        return serialNumberAsCFString.takeUnretainedValue() as? String
    }
    
    private static func getDefaultDeviceName() -> String {
        return Host.current().localizedName ?? "Mac Desktop"
    }
    
    private func generateSessionId() -> UInt32 {
        return arc4random()
    }
    
    private func generateTransferId() -> UInt32 {
        return arc4random()
    }
    
    private func generateMessageId() -> UInt32 {
        return arc4random()
    }
    
    private func calculateChecksum(data: Data) -> UInt32 {
        var checksum: UInt32 = 0
        
        for byte in data {
            checksum = (checksum << 1) ^ UInt32(byte)
        }
        
        return checksum
    }
}

// MARK: - Supporting Types

enum DeviceType: UInt32, CaseIterable {
    case unknown = 0
    case desktopWindows = 1
    case desktopMacOS = 2
    case desktopLinux = 3
    case mobileAndroid = 4
    case mobileiOS = 5
    case webBrowser = 6
}

enum FileTransferStatus: UInt32 {
    case idle = 0
    case discovering = 1
    case connecting = 2
    case authenticating = 3
    case connected = 4
    case transferring = 5
    case paused = 6
    case completed = 7
    case cancelled = 8
    case error = 9
    case disconnected = 10
}

enum TransferDirection: UInt32 {
    case send = 0
    case receive = 1
}

enum FileTransferError: UInt32 {
    case none = 0
    case networkFailure = 1
    case connectionTimeout = 2
    case authFailed = 3
    case fileNotFound = 4
    case fileAccessDenied = 5
    case insufficientSpace = 6
    case transferCancelled = 7
    case protocolError = 8
    case checksumMismatch = 9
    case deviceNotFound = 10
    case invalidRequest = 11
    case unsupportedVersion = 12
}

enum MessageType: UInt16 {
    case discoveryRequest = 0x01
    case discoveryResponse = 0x02
    case connectRequest = 0x03
    case connectResponse = 0x04
    case authRequest = 0x05
    case authResponse = 0x06
    case fileInfo = 0x10
    case fileRequest = 0x11
    case fileResponse = 0x12
    case fileChunk = 0x13
    case fileAck = 0x14
    case transferStart = 0x15
    case transferPause = 0x16
    case transferResume = 0x17
    case transferCancel = 0x18
    case transferComplete = 0x19
    case error = 0x20
    case heartbeat = 0x30
    case disconnect = 0x31
}

struct FileTransferProtocol {
    static let magic: UInt32 = 0x46545250 // "FTRP"
    static let version: UInt16 = 1
}

// MARK: - Protocol Structures

struct FileTransferHeader {
    let magic: UInt32
    let version: UInt16
    let messageType: MessageType
    let messageId: UInt32
    let sessionId: UInt32
    let dataLength: UInt32
    var checksum: UInt32
    let timestamp: UInt64
    
    init(magic: UInt32, version: UInt16, messageType: MessageType, messageId: UInt32, sessionId: UInt32, dataLength: UInt32, checksum: UInt32, timestamp: UInt64) {
        self.magic = magic
        self.version = version
        self.messageType = messageType
        self.messageId = messageId
        self.sessionId = sessionId
        self.dataLength = dataLength
        self.checksum = checksum
        self.timestamp = timestamp
    }
}

struct DeviceInfo {
    let deviceId: String
    let deviceName: String
    let deviceType: DeviceType
    let ipAddress: UInt32
    let port: UInt16
    let lastSeen: UInt64
    let isTrusted: Bool
    let supportsEncryption: Bool
    let maxChunkSize: UInt32
    
    init(deviceId: String = "", deviceName: String = "", deviceType: DeviceType = .unknown, ipAddress: UInt32 = 0, port: UInt16 = 0, lastSeen: UInt64 = 0, isTrusted: Bool = false, supportsEncryption: Bool = false, maxChunkSize: UInt32 = 0) {
        self.deviceId = deviceId
        self.deviceName = deviceName
        self.deviceType = deviceType
        self.ipAddress = ipAddress
        self.port = port
        self.lastSeen = lastSeen
        self.isTrusted = isTrusted
        self.supportsEncryption = supportsEncryption
        self.maxChunkSize = maxChunkSize
    }
}

struct FileInfo {
    let fileName: String
    let filePath: String
    let fileSize: UInt64
    let modifiedTime: UInt64
    let fileHash: UInt32
    let mimeType: String
    let isDirectory: Bool
    let permissions: UInt32
    
    init(fileName: String = "", filePath: String = "", fileSize: UInt64 = 0, modifiedTime: UInt64 = 0, fileHash: UInt32 = 0, mimeType: String = "", isDirectory: Bool = false, permissions: UInt32 = 0) {
        self.fileName = fileName
        self.filePath = filePath
        self.fileSize = fileSize
        self.modifiedTime = modifiedTime
        self.fileHash = fileHash
        self.mimeType = mimeType
        self.isDirectory = isDirectory
        self.permissions = permissions
    }
}

struct FileTransferSession {
    let sessionId: UInt32
    let sessionToken: String
    let remoteDevice: DeviceInfo
    let fileInfo: FileInfo
    let direction: TransferDirection
    var status: FileTransferStatus
    var bytesTransferred: UInt64
    let totalBytes: UInt64
    let chunkSize: UInt32
    let startTime: UInt64
    var lastActivityTime: UInt64
    var progressPercentage: Float
    var transferSpeed: Float
    var estimatedTimeRemaining: UInt32
    var lastError: FileTransferError
}

struct DiscoveryRequest {
    let deviceId: String
    let deviceName: String
    let deviceType: DeviceType
    let listenPort: UInt16
    let supportsEncryption: Bool
    let maxChunkSize: UInt32
    
    init(deviceId: String, deviceName: String, deviceType: DeviceType, listenPort: UInt16, supportsEncryption: Bool, maxChunkSize: UInt32) {
        self.deviceId = deviceId
        self.deviceName = deviceName
        self.deviceType = deviceType
        self.listenPort = listenPort
        self.supportsEncryption = supportsEncryption
        self.maxChunkSize = maxChunkSize
    }
}

struct DiscoveryResponse {
    let deviceId: String
    let deviceName: String
    let deviceType: DeviceType
    let listenPort: UInt16
    let supportsEncryption: Bool
    let maxChunkSize: UInt32
    let acceptsConnections: Bool
    
    init(deviceId: String, deviceName: String, deviceType: DeviceType, listenPort: UInt16, supportsEncryption: Bool, maxChunkSize: UInt32, acceptsConnections: Bool) {
        self.deviceId = deviceId
        self.deviceName = deviceName
        self.deviceType = deviceType
        self.listenPort = listenPort
        self.supportsEncryption = supportsEncryption
        self.maxChunkSize = maxChunkSize
        self.acceptsConnections = acceptsConnections
    }
}

struct ConnectRequest {
    let deviceId: String
    let deviceName: String
    let deviceType: DeviceType
    let protocolVersion: UInt16
    let requestEncryption: Bool
    
    init(deviceId: String, deviceName: String, deviceType: DeviceType, protocolVersion: UInt16, requestEncryption: Bool) {
        self.deviceId = deviceId
        self.deviceName = deviceName
        self.deviceType = deviceType
        self.protocolVersion = protocolVersion
        self.requestEncryption = requestEncryption
    }
}

struct ConnectResponse {
    let connectionAccepted: Bool
    let sessionId: UInt32
    let sessionToken: String
    let encryptionEnabled: Bool
    let maxChunkSize: UInt32
    let errorCode: FileTransferError
    
    init(connectionAccepted: Bool, sessionId: UInt32, sessionToken: String, encryptionEnabled: Bool, maxChunkSize: UInt32, errorCode: FileTransferError) {
        self.connectionAccepted = connectionAccepted
        self.sessionId = sessionId
        self.sessionToken = sessionToken
        self.encryptionEnabled = encryptionEnabled
        self.maxChunkSize = maxChunkSize
        self.errorCode = errorCode
    }
}

// MARK: - Callback Type Aliases

typealias FileTransferProgressCallback = (UInt32, UInt64, UInt64, Float) -> Void
typealias FileTransferCompleteCallback = (UInt32, Bool, FileTransferError) -> Void
typealias FileTransferErrorCallback = (UInt32, FileTransferError, String) -> Void
typealias DeviceDiscoveredCallback = (DeviceInfo) -> Void
typealias DeviceConnectedCallback = (DeviceInfo, UInt32) -> Void
typealias DeviceDisconnectedCallback = (DeviceInfo, UInt32) -> Void
typealias FileReceiveRequestCallback = (DeviceInfo, FileInfo) -> Bool