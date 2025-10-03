import Foundation
import Network
import CryptoKit
import Combine

// MARK: - Project Management Enums

public enum ProjectStatus: Int32, CaseIterable, Codable {
    case planning = 0
    case active = 1
    case onHold = 2
    case completed = 3
    case cancelled = 4
    case archived = 5
}

public enum ProjectPriority: Int32, CaseIterable, Codable {
    case low = 0
    case medium = 1
    case high = 2
    case critical = 3
}

public enum IssueType: Int32, CaseIterable, Codable {
    case bug = 0
    case feature = 1
    case task = 2
    case improvement = 3
    case epic = 4
}

public enum IssueStatus: Int32, CaseIterable, Codable {
    case open = 0
    case inProgress = 1
    case resolved = 2
    case closed = 3
    case reopened = 4
}

public enum IssuePriority: Int32, CaseIterable, Codable {
    case low = 0
    case medium = 1
    case high = 2
    case critical = 3
}

public enum MilestoneStatus: Int32, CaseIterable, Codable {
    case open = 0
    case closed = 1
}

public enum ProjectRole: Int32, CaseIterable, Codable {
    case viewer = 0
    case contributor = 1
    case maintainer = 2
    case admin = 3
    case owner = 4
}

public enum ProjectOperation: Int32, CaseIterable, Codable {
    case create = 0
    case update = 1
    case delete = 2
    case batch = 3
}

public enum ProjectError: Int32, CaseIterable, Codable {
    case none = 0
    case networkFailure = 1
    case authFailed = 2
    case protocolError = 3
    case dataCorruption = 4
    case storageError = 5
    case permissionDenied = 6
    case invalidData = 7
    case versionMismatch = 8
    case timeout = 9
}

public enum ProjectMessageType: Int32, CaseIterable, Codable {
    case handshake = 0
    case auth = 1
    case projectCreate = 2
    case projectUpdate = 3
    case projectDelete = 4
    case projectList = 5
    case issueCreate = 6
    case issueUpdate = 7
    case issueDelete = 8
    case issueList = 9
    case commentAdd = 10
    case commentUpdate = 11
    case commentDelete = 12
    case milestoneCreate = 13
    case milestoneUpdate = 14
    case milestoneDelete = 15
    case memberAdd = 16
    case memberRemove = 17
    case memberUpdate = 18
    case attachmentUpload = 19
    case attachmentDownload = 20
    case attachmentDelete = 21
    case sync = 22
    case notification = 23
    case heartbeat = 24
    case status = 25
    case error = 26
}

public enum NotificationType: Int32, CaseIterable, Codable {
    case projectCreated = 0
    case projectUpdated = 1
    case issueCreated = 2
    case issueUpdated = 3
    case issueAssigned = 4
    case commentAdded = 5
    case milestoneReached = 6
    case memberAdded = 7
    case memberRemoved = 8
    case deadlineApproaching = 9
    case syncCompleted = 10
}

// MARK: - Data Structures

public struct ProjectHeader: Codable {
    let magic: UInt32
    let version: UInt32
    let messageType: ProjectMessageType
    let messageId: UInt32
    let sessionId: UInt32
    let dataLength: UInt32
    let checksum: UInt32
    let timestamp: UInt64
    let reserved: [UInt8]
    
    init(messageType: ProjectMessageType, messageId: UInt32, sessionId: UInt32, dataLength: UInt32, checksum: UInt32) {
        self.magic = 0x50524A54 // "PRJT"
        self.version = 1
        self.messageType = messageType
        self.messageId = messageId
        self.sessionId = sessionId
        self.dataLength = dataLength
        self.checksum = checksum
        self.timestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.reserved = Array(repeating: 0, count: 16)
    }
}

public struct ProjectMember: Codable, Identifiable {
    public let id = UUID()
    let userId: String
    let username: String
    let email: String
    let role: ProjectRole
    let joinedTimestamp: UInt64
    let lastActiveTimestamp: UInt64
    let isActive: Bool
    
    public init(userId: String, username: String, email: String, role: ProjectRole) {
        self.userId = userId
        self.username = username
        self.email = email
        self.role = role
        self.joinedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.lastActiveTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.isActive = true
    }
}

public struct ProjectMilestone: Codable, Identifiable {
    public let id = UUID()
    let milestoneId: String
    let projectId: String
    let title: String
    let description: String
    let status: MilestoneStatus
    let dueDate: UInt64
    let createdTimestamp: UInt64
    let updatedTimestamp: UInt64
    let completedTimestamp: UInt64
    let totalIssues: UInt32
    let closedIssues: UInt32
    
    public init(milestoneId: String, projectId: String, title: String, description: String, dueDate: UInt64) {
        self.milestoneId = milestoneId
        self.projectId = projectId
        self.title = title
        self.description = description
        self.status = .open
        self.dueDate = dueDate
        self.createdTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.updatedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.completedTimestamp = 0
        self.totalIssues = 0
        self.closedIssues = 0
    }
}

public struct ProjectLabel: Codable, Identifiable {
    public let id = UUID()
    let labelId: String
    let name: String
    let color: String
    let description: String
    
    public init(labelId: String, name: String, color: String, description: String) {
        self.labelId = labelId
        self.name = name
        self.color = color
        self.description = description
    }
}

public struct IssueAttachment: Codable, Identifiable {
    public let id = UUID()
    let attachmentId: String
    let issueId: String
    let filename: String
    let originalFilename: String
    let mimeType: String
    let fileSize: UInt32
    let uploaderId: String
    let uploadedTimestamp: UInt64
    let downloadUrl: String
    
    public init(attachmentId: String, issueId: String, filename: String, originalFilename: String, mimeType: String, fileSize: UInt32, uploaderId: String, downloadUrl: String) {
        self.attachmentId = attachmentId
        self.issueId = issueId
        self.filename = filename
        self.originalFilename = originalFilename
        self.mimeType = mimeType
        self.fileSize = fileSize
        self.uploaderId = uploaderId
        self.uploadedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.downloadUrl = downloadUrl
    }
}

public struct IssueComment: Codable, Identifiable {
    public let id = UUID()
    let commentId: String
    let issueId: String
    let authorId: String
    let content: String
    let createdTimestamp: UInt64
    let updatedTimestamp: UInt64
    let isEdited: Bool
    let parentCommentId: String?
    
    public init(commentId: String, issueId: String, authorId: String, content: String, parentCommentId: String? = nil) {
        self.commentId = commentId
        self.issueId = issueId
        self.authorId = authorId
        self.content = content
        self.createdTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.updatedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.isEdited = false
        self.parentCommentId = parentCommentId
    }
}

public struct ProjectIssue: Codable, Identifiable {
    public let id = UUID()
    let issueId: String
    let projectId: String
    let title: String
    let description: String
    let type: IssueType
    var status: IssueStatus
    let priority: IssuePriority
    let reporterId: String
    let assigneeIds: [String]
    let labelIds: [String]
    let milestoneId: String?
    let createdTimestamp: UInt64
    var updatedTimestamp: UInt64
    let dueDate: UInt64
    let resolvedTimestamp: UInt64
    let estimatedHours: UInt32
    let loggedHours: UInt32
    var progressPercentage: Float
    
    public init(issueId: String, projectId: String, title: String, description: String, type: IssueType, priority: IssuePriority, reporterId: String, dueDate: UInt64 = 0) {
        self.issueId = issueId
        self.projectId = projectId
        self.title = title
        self.description = description
        self.type = type
        self.status = .open
        self.priority = priority
        self.reporterId = reporterId
        self.assigneeIds = []
        self.labelIds = []
        self.milestoneId = nil
        self.createdTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.updatedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.dueDate = dueDate
        self.resolvedTimestamp = 0
        self.estimatedHours = 0
        self.loggedHours = 0
        self.progressPercentage = 0.0
    }
}

public struct Project: Codable, Identifiable {
    public let id = UUID()
    let projectId: String
    let name: String
    let description: String
    let ownerId: String
    var status: ProjectStatus
    let priority: ProjectPriority
    let createdTimestamp: UInt64
    var updatedTimestamp: UInt64
    var lastActivityTimestamp: UInt64
    let startDate: UInt64
    let endDate: UInt64
    let isPublic: Bool
    let allowIssues: Bool
    let enableNotifications: Bool
    var totalIssues: UInt32
    var openIssues: UInt32
    var closedIssues: UInt32
    var totalMembers: UInt32
    
    public init(projectId: String, name: String, description: String, ownerId: String, priority: ProjectPriority, startDate: UInt64 = 0, endDate: UInt64 = 0, isPublic: Bool = false) {
        self.projectId = projectId
        self.name = name
        self.description = description
        self.ownerId = ownerId
        self.status = .planning
        self.priority = priority
        self.createdTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.updatedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.lastActivityTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.startDate = startDate
        self.endDate = endDate
        self.isPublic = isPublic
        self.allowIssues = true
        self.enableNotifications = true
        self.totalIssues = 0
        self.openIssues = 0
        self.closedIssues = 0
        self.totalMembers = 1
    }
}

public struct ProjectNotification: Codable, Identifiable {
    public let id = UUID()
    let notificationId: String
    let type: NotificationType
    let title: String
    let message: String
    let projectId: String?
    let issueId: String?
    let userId: String
    let timestamp: UInt64
    var isRead: Bool
    let data: [String: String]
    
    public init(notificationId: String, type: NotificationType, title: String, message: String, userId: String, projectId: String? = nil, issueId: String? = nil, data: [String: String] = [:]) {
        self.notificationId = notificationId
        self.type = type
        self.title = title
        self.message = message
        self.projectId = projectId
        self.issueId = issueId
        self.userId = userId
        self.timestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        self.isRead = false
        self.data = data
    }
}

// MARK: - Configuration

public struct ProjectManagerConfiguration {
    let serverUrl: String
    let serverPort: UInt16
    let userId: String
    let authToken: String
    let deviceId: String
    let connectionTimeout: UInt32
    let heartbeatInterval: UInt32
    let syncInterval: UInt32
    let maxRetries: UInt32
    let enableEncryption: Bool
    let enableCompression: Bool
    let enableNotifications: Bool
    let enableOfflineMode: Bool
    let autoSyncEnabled: Bool
    let localStoragePath: String
    let maxStorageSize: UInt64
    let cacheRetentionDays: UInt32
    let showCompletedIssues: Bool
    let groupByMilestone: Bool
    let itemsPerPage: UInt32
    
    public init(serverUrl: String = "localhost",
                serverPort: UInt16 = 8080,
                userId: String = "macos_user",
                authToken: String = "token",
                deviceId: String = "macos_device",
                connectionTimeout: UInt32 = 30000,
                heartbeatInterval: UInt32 = 30000,
                syncInterval: UInt32 = 300000,
                maxRetries: UInt32 = 3,
                enableEncryption: Bool = true,
                enableCompression: Bool = true,
                enableNotifications: Bool = true,
                enableOfflineMode: Bool = true,
                autoSyncEnabled: Bool = true,
                localStoragePath: String = "~/Library/Application Support/TaishangLaojun/ProjectData",
                maxStorageSize: UInt64 = 1024 * 1024 * 1024,
                cacheRetentionDays: UInt32 = 30,
                showCompletedIssues: Bool = false,
                groupByMilestone: Bool = true,
                itemsPerPage: UInt32 = 50) {
        self.serverUrl = serverUrl
        self.serverPort = serverPort
        self.userId = userId
        self.authToken = authToken
        self.deviceId = deviceId
        self.connectionTimeout = connectionTimeout
        self.heartbeatInterval = heartbeatInterval
        self.syncInterval = syncInterval
        self.maxRetries = maxRetries
        self.enableEncryption = enableEncryption
        self.enableCompression = enableCompression
        self.enableNotifications = enableNotifications
        self.enableOfflineMode = enableOfflineMode
        self.autoSyncEnabled = autoSyncEnabled
        self.localStoragePath = localStoragePath
        self.maxStorageSize = maxStorageSize
        self.cacheRetentionDays = cacheRetentionDays
        self.showCompletedIssues = showCompletedIssues
        self.groupByMilestone = groupByMilestone
        self.itemsPerPage = itemsPerPage
    }
}

// MARK: - Callback Types

public typealias ProjectStatusCallback = (ProjectStatus, Float) -> Void
public typealias ProjectDataCallback = (Project, ProjectOperation) -> Void
public typealias IssueDataCallback = (ProjectIssue, ProjectOperation) -> Void
public typealias NotificationCallback = (ProjectNotification) -> Void
public typealias ProjectErrorCallback = (ProjectError, String) -> Void
public typealias SyncCompleteCallback = (UInt32, UInt32, UInt32) -> Void

// MARK: - Storage Interface Types

public typealias StoreProjectCallback = (Project) -> Bool
public typealias RetrieveProjectCallback = (String) -> Project?
public typealias DeleteProjectCallback = (String) -> Bool
public typealias ListProjectsCallback = () -> [Project]

public typealias StoreIssueCallback = (ProjectIssue) -> Bool
public typealias RetrieveIssueCallback = (String) -> ProjectIssue?
public typealias DeleteIssueCallback = (String) -> Bool
public typealias ListIssuesCallback = (String) -> [ProjectIssue]

public typealias StoreCommentCallback = (IssueComment) -> Bool
public typealias RetrieveCommentsCallback = (String) -> [IssueComment]
public typealias DeleteCommentCallback = (String) -> Bool

public typealias StoreAttachmentCallback = (IssueAttachment, Data) -> Bool
public typealias RetrieveAttachmentCallback = (String) -> (IssueAttachment, Data)?
public typealias DeleteAttachmentCallback = (String) -> Bool

// MARK: - Protocol Messages

struct HandshakeRequest: Codable {
    let deviceId: String
    let deviceName: String
    let protocolVersion: UInt32
    let supportedFeatures: [String]
    let supportsEncryption: Bool
    let supportsCompression: Bool
    let supportsNotifications: Bool
}

struct HandshakeResponse: Codable {
    let handshakeAccepted: Bool
    let serverVersion: UInt32
    let supportedFeatures: [String]
    let encryptionEnabled: Bool
    let compressionEnabled: Bool
    let sessionTimeout: UInt32
}

struct AuthRequest: Codable {
    let userId: String
    let authToken: String
    let deviceSignature: String
    let timestamp: UInt64
}

struct AuthResponse: Codable {
    let authSuccess: Bool
    let sessionToken: String
    let userInfo: [String: String]
    let permissions: [String]
}

// MARK: - Project Manager

@available(macOS 10.15, *)
public class ProjectManager: ObservableObject {
    
    // MARK: - Properties
    
    private let configuration: ProjectManagerConfiguration
    @Published public private(set) var status: ProjectStatus = .planning
    @Published public private(set) var isRunning: Bool = false
    @Published public private(set) var isConnected: Bool = false
    @Published public private(set) var syncProgress: Float = 0.0
    
    private var sessionId: UInt32 = 0
    private var sessionToken: String = ""
    
    // Local data storage
    @Published public private(set) var projects: [String: Project] = [:]
    @Published public private(set) var projectIssues: [String: [ProjectIssue]] = [:]
    @Published public private(set) var issueComments: [String: [IssueComment]] = [:]
    @Published public private(set) var projectMilestones: [String: [ProjectMilestone]] = [:]
    @Published public private(set) var projectMembers: [String: [ProjectMember]] = [:]
    @Published public private(set) var notifications: [ProjectNotification] = []
    
    // Sync state
    private var lastSyncTimestamp: UInt64 = 0
    private var pendingSyncItems: UInt32 = 0
    private var syncedItems: UInt32 = 0
    private var failedItems: UInt32 = 0
    
    // Network
    private var connection: NWConnection?
    private var listener: NWListener?
    
    // Threading
    private let queue = DispatchQueue(label: "com.taishanglaojun.projectmanager", qos: .userInitiated)
    private var syncTimer: Timer?
    private var heartbeatTimer: Timer?
    private var notificationTimer: Timer?
    
    // Callbacks
    public var statusCallback: ProjectStatusCallback?
    public var projectCallback: ProjectDataCallback?
    public var issueCallback: IssueDataCallback?
    public var notificationCallback: NotificationCallback?
    public var errorCallback: ProjectErrorCallback?
    public var syncCompleteCallback: SyncCompleteCallback?
    
    // Storage interface
    public var storeProject: StoreProjectCallback?
    public var retrieveProject: RetrieveProjectCallback?
    public var deleteProject: DeleteProjectCallback?
    public var listProjects: ListProjectsCallback?
    
    public var storeIssue: StoreIssueCallback?
    public var retrieveIssue: RetrieveIssueCallback?
    public var deleteIssue: DeleteIssueCallback?
    public var listIssues: ListIssuesCallback?
    
    public var storeComment: StoreCommentCallback?
    public var retrieveComments: RetrieveCommentsCallback?
    public var deleteComment: DeleteCommentCallback?
    
    public var storeAttachment: StoreAttachmentCallback?
    public var retrieveAttachment: RetrieveAttachmentCallback?
    public var deleteAttachment: DeleteAttachmentCallback?
    
    // MARK: - Initialization
    
    public init(configuration: ProjectManagerConfiguration = ProjectManagerConfiguration()) {
        self.configuration = configuration
        
        // Create storage directory
        createStorageDirectory()
        
        print("macOS Project Manager initialized")
    }
    
    deinit {
        stop()
        print("macOS Project Manager deinitialized")
    }
    
    // MARK: - Lifecycle Management
    
    public func start() -> Bool {
        guard !isRunning else { return true }
        
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Load local data
            self.loadLocalData()
            
            // Start timers if auto-sync is enabled
            if self.configuration.autoSyncEnabled {
                DispatchQueue.main.async {
                    self.startSyncTimer()
                    self.startHeartbeatTimer()
                }
            }
            
            // Start notification timer if notifications are enabled
            if self.configuration.enableNotifications {
                DispatchQueue.main.async {
                    self.startNotificationTimer()
                }
            }
            
            DispatchQueue.main.async {
                self.isRunning = true
                self.status = .active
                self.notifyStatusChange()
            }
        }
        
        print("Project manager started")
        return true
    }
    
    public func stop() {
        guard isRunning else { return }
        
        // Stop timers
        syncTimer?.invalidate()
        syncTimer = nil
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
        notificationTimer?.invalidate()
        notificationTimer = nil
        
        // Disconnect if connected
        if isConnected {
            disconnect()
        }
        
        // Save local data
        queue.async { [weak self] in
            self?.saveLocalData()
        }
        
        isRunning = false
        status = .archived
        notifyStatusChange()
        
        print("Project manager stopped")
    }
    
    // MARK: - Connection Management
    
    public func connect() -> Bool {
        guard !isConnected else { return true }
        
        status = .planning // Connecting status
        notifyStatusChange()
        
        let host = NWEndpoint.Host(configuration.serverUrl)
        let port = NWEndpoint.Port(rawValue: configuration.serverPort) ?? NWEndpoint.Port(8080)
        let endpoint = NWEndpoint.hostPort(host: host, port: port)
        
        let parameters: NWParameters
        if configuration.enableEncryption {
            parameters = NWParameters(tls: .init(), tcp: .init())
        } else {
            parameters = NWParameters.tcp
        }
        
        connection = NWConnection(to: endpoint, using: parameters)
        
        connection?.stateUpdateHandler = { [weak self] state in
            switch state {
            case .ready:
                self?.handleConnectionReady()
            case .failed(let error):
                self?.handleConnectionError(error)
            case .cancelled:
                self?.handleConnectionCancelled()
            default:
                break
            }
        }
        
        connection?.start(queue: queue)
        
        return true
    }
    
    public func disconnect() {
        connection?.cancel()
        connection = nil
        
        isConnected = false
        sessionId = 0
        sessionToken = ""
        
        status = .onHold // Offline status
        notifyStatusChange()
        
        print("Disconnected from project server")
    }
    
    // MARK: - Project Operations
    
    public func createProject(_ project: Project) -> Bool {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Store locally
            if let storeProject = self.storeProject {
                _ = storeProject(project)
            }
            
            DispatchQueue.main.async {
                self.projects[project.projectId] = project
                
                // Notify callback
                self.projectCallback?(project, .create)
                
                // Sync if connected
                if self.isConnected && self.configuration.autoSyncEnabled {
                    self.syncProject(project.projectId)
                }
            }
        }
        
        return true
    }
    
    public func updateProject(_ project: Project) -> Bool {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Update locally
            if let storeProject = self.storeProject {
                _ = storeProject(project)
            }
            
            DispatchQueue.main.async {
                self.projects[project.projectId] = project
                
                // Notify callback
                self.projectCallback?(project, .update)
            }
        }
        
        return true
    }
    
    public func deleteProject(_ projectId: String) -> Bool {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Delete locally
            if let deleteProject = self.deleteProject {
                _ = deleteProject(projectId)
            }
            
            DispatchQueue.main.async {
                if let project = self.projects[projectId] {
                    // Notify callback before deletion
                    self.projectCallback?(project, .delete)
                    
                    self.projects.removeValue(forKey: projectId)
                    
                    // Also remove related data
                    self.projectIssues.removeValue(forKey: projectId)
                    self.projectMilestones.removeValue(forKey: projectId)
                    self.projectMembers.removeValue(forKey: projectId)
                }
            }
        }
        
        return true
    }
    
    public func getProject(_ projectId: String) -> Project? {
        // Try local cache first
        if let project = projects[projectId] {
            return project
        }
        
        // Try storage interface
        return retrieveProject?(projectId)
    }
    
    public func getAllProjects() -> [Project] {
        // Try storage interface first
        if let listProjects = listProjects {
            return listProjects()
        }
        
        // Use local cache
        return Array(projects.values)
    }
    
    // MARK: - Issue Operations
    
    public func createIssue(_ issue: ProjectIssue) -> Bool {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Store locally
            if let storeIssue = self.storeIssue {
                _ = storeIssue(issue)
            }
            
            DispatchQueue.main.async {
                if self.projectIssues[issue.projectId] == nil {
                    self.projectIssues[issue.projectId] = []
                }
                self.projectIssues[issue.projectId]?.append(issue)
                
                // Update project statistics
                if var project = self.projects[issue.projectId] {
                    project.totalIssues += 1
                    if issue.status == .open {
                        project.openIssues += 1
                    }
                    project.lastActivityTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
                    self.projects[issue.projectId] = project
                }
                
                // Notify callback
                self.issueCallback?(issue, .create)
            }
        }
        
        return true
    }
    
    public func updateIssue(_ issue: ProjectIssue) -> Bool {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Update locally
            if let storeIssue = self.storeIssue {
                _ = storeIssue(issue)
            }
            
            DispatchQueue.main.async {
                if let issues = self.projectIssues[issue.projectId] {
                    if let index = issues.firstIndex(where: { $0.issueId == issue.issueId }) {
                        self.projectIssues[issue.projectId]?[index] = issue
                    } else {
                        self.projectIssues[issue.projectId]?.append(issue)
                    }
                } else {
                    self.projectIssues[issue.projectId] = [issue]
                }
                
                // Notify callback
                self.issueCallback?(issue, .update)
            }
        }
        
        return true
    }
    
    public func deleteIssue(_ issueId: String) -> Bool {
        queue.async { [weak self] in
            guard let self = self else { return }
            
            // Delete locally
            if let deleteIssue = self.deleteIssue {
                _ = deleteIssue(issueId)
            }
            
            DispatchQueue.main.async {
                // Find and remove from local cache
                for (projectId, issues) in self.projectIssues {
                    if let index = issues.firstIndex(where: { $0.issueId == issueId }) {
                        let issue = issues[index]
                        
                        // Notify callback before deletion
                        self.issueCallback?(issue, .delete)
                        
                        self.projectIssues[projectId]?.remove(at: index)
                        
                        // Update project statistics
                        if var project = self.projects[projectId] {
                            project.totalIssues -= 1
                            project.lastActivityTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
                            self.projects[projectId] = project
                        }
                        
                        break
                    }
                }
            }
        }
        
        return true
    }
    
    public func getIssue(_ issueId: String) -> ProjectIssue? {
        // Try local cache first
        for issues in projectIssues.values {
            if let issue = issues.first(where: { $0.issueId == issueId }) {
                return issue
            }
        }
        
        // Try storage interface
        return retrieveIssue?(issueId)
    }
    
    public func getIssues(for projectId: String) -> [ProjectIssue] {
        // Try storage interface first
        if let listIssues = listIssues {
            return listIssues(projectId)
        }
        
        // Use local cache
        return projectIssues[projectId] ?? []
    }
    
    public func assignIssue(_ issueId: String, to assigneeId: String) -> Bool {
        guard var issue = getIssue(issueId) else { return false }
        
        // Add assignee if not already assigned
        if !issue.assigneeIds.contains(assigneeId) {
            var updatedIssue = issue
            updatedIssue.assigneeIds.append(assigneeId)
            updatedIssue.updatedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
            
            return updateIssue(updatedIssue)
        }
        
        return true
    }
    
    public func updateIssueStatus(_ issueId: String, status: IssueStatus) -> Bool {
        guard var issue = getIssue(issueId) else { return false }
        
        let oldStatus = issue.status
        issue.status = status
        issue.updatedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        
        if status == .resolved || status == .closed {
            issue.resolvedTimestamp = UInt64(Date().timeIntervalSince1970 * 1000)
        }
        
        let result = updateIssue(issue)
        
        // Update project statistics
        if result {
            DispatchQueue.main.async { [weak self] in
                guard let self = self else { return }
                
                if var project = self.projects[issue.projectId] {
                    if oldStatus == .open && status != .open {
                        project.openIssues -= 1
                        project.closedIssues += 1
                    } else if oldStatus != .open && status == .open {
                        project.openIssues += 1
                        project.closedIssues -= 1
                    }
                    self.projects[issue.projectId] = project
                }
            }
        }
        
        return result
    }
    
    // MARK: - Synchronization
    
    public func syncAll() -> Bool {
        guard isConnected else {
            if !connect() {
                return false
            }
        }
        
        status = .active // Syncing status
        notifyStatusChange()
        
        queue.async { [weak self] in
            guard let self = self else { return }
            
            var success = true
            
            // Sync projects
            for project in self.projects.values {
                // Implementation would send sync messages for each project
                self.syncedItems += 1
            }
            
            // Sync issues
            for issues in self.projectIssues.values {
                for issue in issues {
                    // Implementation would send sync messages for each issue
                    self.syncedItems += 1
                }
            }
            
            DispatchQueue.main.async {
                self.status = success ? .completed : .onHold
                self.notifyStatusChange()
                
                if let callback = self.syncCompleteCallback {
                    let totalProjects = UInt32(self.projects.count)
                    let totalIssues = UInt32(self.projectIssues.values.reduce(0) { $0 + $1.count })
                    callback(totalProjects, totalIssues, self.failedItems)
                }
            }
        }
        
        return true
    }
    
    public func syncProject(_ projectId: String) -> Bool {
        // Implementation would sync specific project
        return true
    }
    
    // MARK: - Statistics and Monitoring
    
    public func getStats() -> (totalProjects: UInt32, totalIssues: UInt32, pendingSync: UInt32) {
        let totalProjects = UInt32(projects.count)
        let totalIssues = UInt32(projectIssues.values.reduce(0) { $0 + $1.count })
        return (totalProjects, totalIssues, pendingSyncItems)
    }
    
    public func calculateProgress(for projectId: String) -> Float {
        let issues = getIssues(for: projectId)
        guard !issues.isEmpty else { return 0.0 }
        
        let totalProgress = issues.reduce(0.0) { $0 + $1.progressPercentage }
        return totalProgress / Float(issues.count)
    }
    
    // MARK: - Private Methods
    
    private func createStorageDirectory() {
        let fileManager = FileManager.default
        let expandedPath = NSString(string: configuration.localStoragePath).expandingTildeInPath
        
        if !fileManager.fileExists(atPath: expandedPath) {
            try? fileManager.createDirectory(atPath: expandedPath, withIntermediateDirectories: true, attributes: nil)
        }
    }
    
    private func handleConnectionReady() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            // Perform handshake and authentication
            self.queue.async {
                if self.performHandshake() && self.authenticate() {
                    DispatchQueue.main.async {
                        self.isConnected = true
                        self.status = .active
                        self.notifyStatusChange()
                        print("Connected to project server")
                    }
                } else {
                    self.disconnect()
                }
            }
        }
    }
    
    private func handleConnectionError(_ error: NWError) {
        DispatchQueue.main.async { [weak self] in
            self?.handleError(.networkFailure, message: "Connection failed: \(error.localizedDescription)")
        }
    }
    
    private func handleConnectionCancelled() {
        DispatchQueue.main.async { [weak self] in
            self?.isConnected = false
            self?.status = .onHold
            self?.notifyStatusChange()
        }
    }
    
    private func performHandshake() -> Bool {
        // Create handshake request
        let request = HandshakeRequest(
            deviceId: configuration.deviceId,
            deviceName: "macOS Desktop",
            protocolVersion: 1,
            supportedFeatures: ["projects", "issues", "comments", "milestones", "attachments"],
            supportsEncryption: configuration.enableEncryption,
            supportsCompression: configuration.enableCompression,
            supportsNotifications: configuration.enableNotifications
        )
        
        // Implementation would send handshake message and receive response
        // For now, return true to simulate successful handshake
        return true
    }
    
    private func authenticate() -> Bool {
        // Create auth request
        let request = AuthRequest(
            userId: configuration.userId,
            authToken: configuration.authToken,
            deviceSignature: generateDeviceSignature(),
            timestamp: UInt64(Date().timeIntervalSince1970 * 1000)
        )
        
        // Implementation would send auth message and receive response
        // For now, return true to simulate successful authentication
        sessionId = generateMessageId()
        sessionToken = "mock_session_token"
        return true
    }
    
    private func startSyncTimer() {
        syncTimer = Timer.scheduledTimer(withTimeInterval: TimeInterval(configuration.syncInterval / 1000), repeats: true) { [weak self] _ in
            guard let self = self else { return }
            if self.configuration.autoSyncEnabled && self.isConnected {
                _ = self.syncAll()
            }
        }
    }
    
    private func startHeartbeatTimer() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: TimeInterval(configuration.heartbeatInterval / 1000), repeats: true) { [weak self] _ in
            guard let self = self else { return }
            if self.isConnected {
                self.sendHeartbeat()
            }
        }
    }
    
    private func startNotificationTimer() {
        notificationTimer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { [weak self] _ in
            guard let self = self else { return }
            self.processNotifications()
        }
    }
    
    private func sendHeartbeat() {
        // Implementation would send heartbeat message
        print("Sending heartbeat")
    }
    
    private func processNotifications() {
        // Process pending notifications
        for notification in notifications where !notification.isRead {
            notificationCallback?(notification)
        }
    }
    
    private func loadLocalData() {
        let fileManager = FileManager.default
        let expandedPath = NSString(string: configuration.localStoragePath).expandingTildeInPath
        let projectsFile = "\(expandedPath)/projects.json"
        
        guard fileManager.fileExists(atPath: projectsFile),
              let data = fileManager.contents(atPath: projectsFile) else {
            return
        }
        
        do {
            let decoder = JSONDecoder()
            let projectData = try decoder.decode([String: Project].self, from: data)
            
            DispatchQueue.main.async { [weak self] in
                self?.projects = projectData
            }
        } catch {
            print("Failed to load local data: \(error)")
        }
    }
    
    private func saveLocalData() {
        let fileManager = FileManager.default
        let expandedPath = NSString(string: configuration.localStoragePath).expandingTildeInPath
        let projectsFile = "\(expandedPath)/projects.json"
        
        do {
            let encoder = JSONEncoder()
            encoder.outputFormatting = .prettyPrinted
            let data = try encoder.encode(projects)
            fileManager.createFile(atPath: projectsFile, contents: data, attributes: nil)
        } catch {
            print("Failed to save local data: \(error)")
        }
    }
    
    private func notifyStatusChange() {
        statusCallback?(status, syncProgress)
    }
    
    private func handleError(_ error: ProjectError, message: String) {
        status = .onHold // Error status
        errorCallback?(error, message)
        print("Project error: \(message)")
    }
    
    private func generateMessageId() -> UInt32 {
        return UInt32.random(in: 1...UInt32.max)
    }
    
    private func generateDeviceSignature() -> String {
        return "\(configuration.deviceId)_\(UInt64(Date().timeIntervalSince1970 * 1000))"
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
}

// MARK: - Extensions

extension ProjectError {
    public var description: String {
        switch self {
        case .none: return "No error"
        case .networkFailure: return "Network failure"
        case .authFailed: return "Authentication failed"
        case .protocolError: return "Protocol error"
        case .dataCorruption: return "Data corruption"
        case .storageError: return "Storage error"
        case .permissionDenied: return "Permission denied"
        case .invalidData: return "Invalid data"
        case .versionMismatch: return "Version mismatch"
        case .timeout: return "Timeout"
        }
    }
}

extension ProjectStatus {
    public var description: String {
        switch self {
        case .planning: return "Planning"
        case .active: return "Active"
        case .onHold: return "On Hold"
        case .completed: return "Completed"
        case .cancelled: return "Cancelled"
        case .archived: return "Archived"
        }
    }
}

extension IssueStatus {
    public var description: String {
        switch self {
        case .open: return "Open"
        case .inProgress: return "In Progress"
        case .resolved: return "Resolved"
        case .closed: return "Closed"
        case .reopened: return "Reopened"
        }
    }
}

// MARK: - Utility Functions

public func generateProjectId() -> String {
    return UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(16).lowercased()
}

public func getCurrentTimestamp() -> UInt64 {
    return UInt64(Date().timeIntervalSince1970 * 1000)
}

public func validateProjectData(_ project: Project) -> Bool {
    return !project.projectId.isEmpty && !project.name.isEmpty && !project.ownerId.isEmpty
}

public func validateIssueData(_ issue: ProjectIssue) -> Bool {
    return !issue.issueId.isEmpty && !issue.projectId.isEmpty && !issue.title.isEmpty && !issue.reporterId.isEmpty
}