#include "project_management.h"
#include <windows.h>
#include <winsock2.h>
#include <ws2tcpip.h>
#include <openssl/ssl.h>
#include <openssl/err.h>
#include <openssl/sha.h>
#include <json/json.h>
#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <atomic>
#include <memory>
#include <fstream>
#include <sstream>
#include <chrono>
#include <queue>
#include <algorithm>

#pragma comment(lib, "ws2_32.lib")
#pragma comment(lib, "libssl.lib")
#pragma comment(lib, "libcrypto.lib")

// MARK: - Internal Structures

struct ConnectionContext {
    SOCKET socket_fd;
    SSL* ssl;
    sockaddr_in address;
    bool is_connected;
    std::thread thread;
};

struct TaskContext {
    ProjectManager* manager;
    std::function<void()> task;
    uint32_t priority;
    uint64_t timestamp;
};

// MARK: - Windows Project Manager Implementation

class WindowsProjectManager {
public:
    WindowsProjectManager(const ProjectManagerConfiguration* config);
    ~WindowsProjectManager();
    
    // Lifecycle
    bool start();
    void stop();
    
    // Connection management
    bool connect();
    void disconnect();
    bool isConnected() const { return is_connected_; }
    
    // Project operations
    bool createProject(const Project* project);
    bool updateProject(const Project* project);
    bool deleteProject(const char* project_id);
    bool getProject(const char* project_id, Project* project);
    bool listProjects(Project** projects, uint32_t* count);
    
    // Issue operations
    bool createIssue(const ProjectIssue* issue);
    bool updateIssue(const ProjectIssue* issue);
    bool deleteIssue(const char* issue_id);
    bool getIssue(const char* issue_id, ProjectIssue* issue);
    bool listIssues(const char* project_id, ProjectIssue** issues, uint32_t* count);
    bool assignIssue(const char* issue_id, const char* assignee_id);
    bool updateIssueStatus(const char* issue_id, IssueStatus status);
    
    // Comment operations
    bool addComment(const IssueComment* comment);
    bool updateComment(const IssueComment* comment);
    bool deleteComment(const char* comment_id);
    bool getComments(const char* issue_id, IssueComment** comments, uint32_t* count);
    
    // Milestone operations
    bool createMilestone(const ProjectMilestone* milestone);
    bool updateMilestone(const ProjectMilestone* milestone);
    bool deleteMilestone(const char* milestone_id);
    bool listMilestones(const char* project_id, ProjectMilestone** milestones, uint32_t* count);
    
    // Member operations
    bool addMember(const char* project_id, const ProjectMember* member);
    bool removeMember(const char* project_id, const char* user_id);
    bool updateMemberRole(const char* project_id, const char* user_id, ProjectRole role);
    bool listMembers(const char* project_id, ProjectMember** members, uint32_t* count);
    
    // Attachment operations
    bool uploadAttachment(const char* issue_id, const IssueAttachment* attachment, const void* data, uint32_t data_size);
    bool downloadAttachment(const char* attachment_id, IssueAttachment* attachment, void** data, uint32_t* data_size);
    bool deleteAttachment(const char* attachment_id);
    
    // Synchronization
    bool syncAll();
    bool syncProject(const char* project_id);
    
    // Status and monitoring
    ProjectStatus getStatus() const { return status_; }
    float getSyncProgress() const;
    void getStats(uint32_t* total_projects, uint32_t* total_issues, uint32_t* pending_sync);
    
    // Notifications
    bool getNotifications(ProjectNotification** notifications, uint32_t* count);
    bool markNotificationRead(const char* notification_id);
    bool clearNotifications();
    
    // Search and filtering
    bool searchIssues(const char* project_id, const char* query, ProjectIssue** issues, uint32_t* count);
    bool filterIssues(const char* project_id, IssueStatus status, IssuePriority priority, const char* assignee_id, ProjectIssue** issues, uint32_t* count);
    
    // Callback setters
    void setStatusCallback(ProjectStatusCallback callback) { status_callback_ = callback; }
    void setProjectCallback(ProjectDataCallback callback) { project_callback_ = callback; }
    void setIssueCallback(IssueDataCallback callback) { issue_callback_ = callback; }
    void setNotificationCallback(NotificationCallback callback) { notification_callback_ = callback; }
    void setErrorCallback(ProjectErrorCallback callback) { error_callback_ = callback; }
    void setSyncCompleteCallback(SyncCompleteCallback callback) { sync_complete_callback_ = callback; }
    
    // Storage interface setters
    void setProjectStorage(StoreProjectCallback store_project,
                          RetrieveProjectCallback retrieve_project,
                          DeleteProjectCallback delete_project,
                          ListProjectsCallback list_projects);
    
    void setIssueStorage(StoreIssueCallback store_issue,
                        RetrieveIssueCallback retrieve_issue,
                        DeleteIssueCallback delete_issue,
                        ListIssuesCallback list_issues);
    
    void setCommentStorage(StoreCommentCallback store_comment,
                          RetrieveCommentsCallback retrieve_comments,
                          DeleteCommentCallback delete_comment);
    
    void setAttachmentStorage(StoreAttachmentCallback store_attachment,
                             RetrieveAttachmentCallback retrieve_attachment,
                             DeleteAttachmentCallback delete_attachment);

private:
    // Configuration and state
    ProjectManagerConfiguration config_;
    ProjectStatus status_;
    bool is_running_;
    bool is_connected_;
    uint32_t session_id_;
    std::string session_token_;
    
    // Local data
    std::map<std::string, Project> projects_;
    std::map<std::string, std::vector<ProjectIssue>> project_issues_;
    std::map<std::string, std::vector<IssueComment>> issue_comments_;
    std::map<std::string, std::vector<ProjectMilestone>> project_milestones_;
    std::map<std::string, std::vector<ProjectMember>> project_members_;
    std::queue<ProjectNotification> notifications_;
    
    // Sync state
    uint64_t last_sync_timestamp_;
    uint32_t pending_sync_items_;
    uint32_t synced_items_;
    uint32_t failed_items_;
    
    // Network
    SOCKET socket_fd_;
    SSL_CTX* ssl_ctx_;
    SSL* ssl_;
    
    // Threading
    std::thread sync_thread_;
    std::thread heartbeat_thread_;
    std::thread notification_thread_;
    std::mutex mutex_;
    std::condition_variable condition_;
    std::atomic<bool> shutdown_requested_;
    
    // Task queue
    std::queue<TaskContext> task_queue_;
    std::mutex task_mutex_;
    std::condition_variable task_condition_;
    std::thread task_thread_;
    
    // Callbacks
    ProjectStatusCallback status_callback_;
    ProjectDataCallback project_callback_;
    IssueDataCallback issue_callback_;
    NotificationCallback notification_callback_;
    ProjectErrorCallback error_callback_;
    SyncCompleteCallback sync_complete_callback_;
    
    // Storage interface
    StoreProjectCallback store_project_;
    RetrieveProjectCallback retrieve_project_;
    DeleteProjectCallback delete_project_;
    ListProjectsCallback list_projects_;
    
    StoreIssueCallback store_issue_;
    RetrieveIssueCallback retrieve_issue_;
    DeleteIssueCallback delete_issue_;
    ListIssuesCallback list_issues_;
    
    StoreCommentCallback store_comment_;
    RetrieveCommentsCallback retrieve_comments_;
    DeleteCommentCallback delete_comment_;
    
    StoreAttachmentCallback store_attachment_;
    RetrieveAttachmentCallback retrieve_attachment_;
    DeleteAttachmentCallback delete_attachment_;
    
    // Private methods
    bool initializeWinsock();
    void cleanupWinsock();
    bool initializeSSL();
    void cleanupSSL();
    bool performHandshake();
    bool authenticate();
    bool sendMessage(const ProjectHeader* header, const void* data);
    bool receiveMessage(ProjectHeader* header, void** data);
    void sendHeartbeat();
    void processTaskQueue();
    void syncThreadFunc();
    void heartbeatThreadFunc();
    void notificationThreadFunc();
    void loadLocalData();
    void saveLocalData();
    void notifyStatusChange();
    void handleError(ProjectError error, const std::string& message);
    uint32_t generateMessageId();
    uint64_t getCurrentTimestamp();
    uint32_t calculateChecksum(const void* data, uint32_t length);
    std::string generateDeviceSignature();
    Json::Value projectToJson(const Project* project);
    Json::Value issueToJson(const ProjectIssue* issue);
    Json::Value commentToJson(const IssueComment* comment);
    Json::Value milestoneToJson(const ProjectMilestone* milestone);
    Json::Value memberToJson(const ProjectMember* member);
    bool jsonToProject(const Json::Value& json, Project* project);
    bool jsonToIssue(const Json::Value& json, ProjectIssue* issue);
    bool jsonToComment(const Json::Value& json, IssueComment* comment);
    bool jsonToMilestone(const Json::Value& json, ProjectMilestone* milestone);
    bool jsonToMember(const Json::Value& json, ProjectMember* member);
};

// MARK: - WindowsProjectManager Implementation

WindowsProjectManager::WindowsProjectManager(const ProjectManagerConfiguration* config)
    : status_(PROJECT_STATUS_PLANNING)
    , is_running_(false)
    , is_connected_(false)
    , session_id_(0)
    , last_sync_timestamp_(0)
    , pending_sync_items_(0)
    , synced_items_(0)
    , failed_items_(0)
    , socket_fd_(INVALID_SOCKET)
    , ssl_ctx_(nullptr)
    , ssl_(nullptr)
    , shutdown_requested_(false)
    , status_callback_(nullptr)
    , project_callback_(nullptr)
    , issue_callback_(nullptr)
    , notification_callback_(nullptr)
    , error_callback_(nullptr)
    , sync_complete_callback_(nullptr)
    , store_project_(nullptr)
    , retrieve_project_(nullptr)
    , delete_project_(nullptr)
    , list_projects_(nullptr)
    , store_issue_(nullptr)
    , retrieve_issue_(nullptr)
    , delete_issue_(nullptr)
    , list_issues_(nullptr)
    , store_comment_(nullptr)
    , retrieve_comments_(nullptr)
    , delete_comment_(nullptr)
    , store_attachment_(nullptr)
    , retrieve_attachment_(nullptr)
    , delete_attachment_(nullptr) {
    
    if (config) {
        config_ = *config;
    } else {
        // Set default configuration
        strcpy_s(config_.server_url, "localhost");
        config_.server_port = DEFAULT_PROJECT_PORT;
        strcpy_s(config_.user_id, "windows_user");
        strcpy_s(config_.auth_token, "token");
        strcpy_s(config_.device_id, "windows_device");
        config_.connection_timeout = CONNECTION_TIMEOUT_MS;
        config_.heartbeat_interval = HEARTBEAT_INTERVAL_MS;
        config_.sync_interval = SYNC_INTERVAL_MS;
        config_.max_retries = 3;
        config_.enable_encryption = true;
        config_.enable_compression = true;
        config_.enable_notifications = true;
        config_.enable_offline_mode = true;
        config_.auto_sync_enabled = true;
        strcpy_s(config_.local_storage_path, "./project_data");
        config_.max_storage_size = 1024 * 1024 * 1024; // 1GB
        config_.cache_retention_days = 30;
        config_.show_completed_issues = false;
        config_.group_by_milestone = true;
        config_.items_per_page = 50;
    }
    
    // Create storage directory
    CreateDirectoryA(config_.local_storage_path, NULL);
    
    std::cout << "Windows Project Manager created" << std::endl;
}

WindowsProjectManager::~WindowsProjectManager() {
    stop();
    cleanupSSL();
    cleanupWinsock();
    std::cout << "Windows Project Manager destroyed" << std::endl;
}

bool WindowsProjectManager::start() {
    std::lock_guard<std::mutex> lock(mutex_);
    
    if (is_running_) {
        return true;
    }
    
    shutdown_requested_ = false;
    
    // Initialize Winsock
    if (!initializeWinsock()) {
        return false;
    }
    
    // Initialize SSL if encryption is enabled
    if (config_.enable_encryption) {
        if (!initializeSSL()) {
            cleanupWinsock();
            return false;
        }
    }
    
    // Load local data
    loadLocalData();
    
    // Start worker threads
    try {
        task_thread_ = std::thread(&WindowsProjectManager::processTaskQueue, this);
        
        if (config_.auto_sync_enabled) {
            sync_thread_ = std::thread(&WindowsProjectManager::syncThreadFunc, this);
            heartbeat_thread_ = std::thread(&WindowsProjectManager::heartbeatThreadFunc, this);
        }
        
        if (config_.enable_notifications) {
            notification_thread_ = std::thread(&WindowsProjectManager::notificationThreadFunc, this);
        }
    } catch (const std::exception& e) {
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Failed to start worker threads: " + std::string(e.what()));
        return false;
    }
    
    is_running_ = true;
    status_ = PROJECT_STATUS_ACTIVE;
    notifyStatusChange();
    
    std::cout << "Project manager started" << std::endl;
    return true;
}

void WindowsProjectManager::stop() {
    std::lock_guard<std::mutex> lock(mutex_);
    
    if (!is_running_) {
        return;
    }
    
    shutdown_requested_ = true;
    condition_.notify_all();
    task_condition_.notify_all();
    
    // Disconnect if connected
    if (is_connected_) {
        disconnect();
    }
    
    is_running_ = false;
    
    // Wait for threads to finish
    if (task_thread_.joinable()) {
        task_thread_.join();
    }
    if (sync_thread_.joinable()) {
        sync_thread_.join();
    }
    if (heartbeat_thread_.joinable()) {
        heartbeat_thread_.join();
    }
    if (notification_thread_.joinable()) {
        notification_thread_.join();
    }
    
    // Save local data
    saveLocalData();
    
    status_ = PROJECT_STATUS_ARCHIVED;
    notifyStatusChange();
    
    std::cout << "Project manager stopped" << std::endl;
}

bool WindowsProjectManager::connect() {
    std::lock_guard<std::mutex> lock(mutex_);
    
    if (is_connected_) {
        return true;
    }
    
    status_ = PROJECT_STATUS_PLANNING; // Connecting status
    notifyStatusChange();
    
    // Create socket
    socket_fd_ = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (socket_fd_ == INVALID_SOCKET) {
        handleError(PROJECT_ERROR_NETWORK_FAILURE, "Failed to create socket");
        return false;
    }
    
    // Set socket timeout
    DWORD timeout = config_.connection_timeout;
    setsockopt(socket_fd_, SOL_SOCKET, SO_RCVTIMEO, (char*)&timeout, sizeof(timeout));
    setsockopt(socket_fd_, SOL_SOCKET, SO_SNDTIMEO, (char*)&timeout, sizeof(timeout));
    
    // Connect to server
    sockaddr_in server_addr;
    ZeroMemory(&server_addr, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(config_.server_port);
    
    if (inet_pton(AF_INET, config_.server_url, &server_addr.sin_addr) <= 0) {
        // Try to resolve hostname
        struct addrinfo hints, *result;
        ZeroMemory(&hints, sizeof(hints));
        hints.ai_family = AF_INET;
        hints.ai_socktype = SOCK_STREAM;
        
        if (getaddrinfo(config_.server_url, nullptr, &hints, &result) != 0) {
            closesocket(socket_fd_);
            socket_fd_ = INVALID_SOCKET;
            handleError(PROJECT_ERROR_NETWORK_FAILURE, "Failed to resolve server address");
            return false;
        }
        
        server_addr.sin_addr = ((sockaddr_in*)result->ai_addr)->sin_addr;
        freeaddrinfo(result);
    }
    
    if (::connect(socket_fd_, (sockaddr*)&server_addr, sizeof(server_addr)) == SOCKET_ERROR) {
        closesocket(socket_fd_);
        socket_fd_ = INVALID_SOCKET;
        handleError(PROJECT_ERROR_NETWORK_FAILURE, "Failed to connect to server");
        return false;
    }
    
    // Setup SSL if enabled
    if (config_.enable_encryption) {
        ssl_ = SSL_new(ssl_ctx_);
        SSL_set_fd(ssl_, static_cast<int>(socket_fd_));
        
        if (SSL_connect(ssl_) <= 0) {
            SSL_free(ssl_);
            ssl_ = nullptr;
            closesocket(socket_fd_);
            socket_fd_ = INVALID_SOCKET;
            handleError(PROJECT_ERROR_NETWORK_FAILURE, "SSL connection failed");
            return false;
        }
    }
    
    // Perform handshake
    if (!performHandshake()) {
        disconnect();
        return false;
    }
    
    // Authenticate
    if (!authenticate()) {
        disconnect();
        return false;
    }
    
    is_connected_ = true;
    status_ = PROJECT_STATUS_ACTIVE;
    notifyStatusChange();
    
    std::cout << "Connected to project server" << std::endl;
    return true;
}

void WindowsProjectManager::disconnect() {
    if (ssl_) {
        SSL_shutdown(ssl_);
        SSL_free(ssl_);
        ssl_ = nullptr;
    }
    
    if (socket_fd_ != INVALID_SOCKET) {
        closesocket(socket_fd_);
        socket_fd_ = INVALID_SOCKET;
    }
    
    is_connected_ = false;
    session_id_ = 0;
    session_token_.clear();
    
    status_ = PROJECT_STATUS_ON_HOLD; // Offline status
    notifyStatusChange();
    
    std::cout << "Disconnected from project server" << std::endl;
}

bool WindowsProjectManager::createProject(const Project* project) {
    if (!project) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Store locally
    if (store_project_ && !store_project_(project)) {
        return false;
    }
    
    projects_[project->project_id] = *project;
    
    // Notify callback
    if (project_callback_) {
        project_callback_(project, PROJECT_OPERATION_CREATE);
    }
    
    // Sync if connected
    if (is_connected_ && config_.auto_sync_enabled) {
        // Add to task queue for async sync
        TaskContext task;
        task.manager = reinterpret_cast<ProjectManager*>(this);
        task.task = [this, proj = *project]() {
            // Send create project message
            Json::Value request = projectToJson(&proj);
            std::string request_str = request.toStyledString();
            
            ProjectHeader header = {};
            header.magic = PROJECT_MANAGEMENT_MAGIC;
            header.version = PROJECT_MANAGEMENT_PROTOCOL_VERSION;
            header.message_type = PROJECT_MSG_TYPE_PROJECT_CREATE;
            header.message_id = generateMessageId();
            header.session_id = session_id_;
            header.data_length = static_cast<uint32_t>(request_str.length());
            header.timestamp = getCurrentTimestamp();
            header.checksum = calculateChecksum(request_str.c_str(), header.data_length);
            
            sendMessage(&header, request_str.c_str());
        };
        task.priority = 1;
        task.timestamp = getCurrentTimestamp();
        
        std::lock_guard<std::mutex> task_lock(task_mutex_);
        task_queue_.push(task);
        task_condition_.notify_one();
    }
    
    return true;
}

bool WindowsProjectManager::updateProject(const Project* project) {
    if (!project) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Update locally
    if (store_project_ && !store_project_(project)) {
        return false;
    }
    
    projects_[project->project_id] = *project;
    
    // Notify callback
    if (project_callback_) {
        project_callback_(project, PROJECT_OPERATION_UPDATE);
    }
    
    return true;
}

bool WindowsProjectManager::deleteProject(const char* project_id) {
    if (!project_id) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Delete locally
    if (delete_project_ && !delete_project_(project_id)) {
        return false;
    }
    
    auto it = projects_.find(project_id);
    if (it != projects_.end()) {
        // Notify callback before deletion
        if (project_callback_) {
            project_callback_(&it->second, PROJECT_OPERATION_DELETE);
        }
        
        projects_.erase(it);
        
        // Also remove related data
        project_issues_.erase(project_id);
        project_milestones_.erase(project_id);
        project_members_.erase(project_id);
    }
    
    return true;
}

bool WindowsProjectManager::getProject(const char* project_id, Project* project) {
    if (!project_id || !project) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Try local cache first
    auto it = projects_.find(project_id);
    if (it != projects_.end()) {
        *project = it->second;
        return true;
    }
    
    // Try storage interface
    if (retrieve_project_) {
        return retrieve_project_(project_id, project);
    }
    
    return false;
}

bool WindowsProjectManager::listProjects(Project** projects, uint32_t* count) {
    if (!projects || !count) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Try storage interface first
    if (list_projects_) {
        return list_projects_(projects, count);
    }
    
    // Use local cache
    *count = static_cast<uint32_t>(projects_.size());
    if (*count == 0) {
        *projects = nullptr;
        return true;
    }
    
    *projects = static_cast<Project*>(malloc(*count * sizeof(Project)));
    if (!*projects) {
        *count = 0;
        return false;
    }
    
    uint32_t index = 0;
    for (const auto& pair : projects_) {
        (*projects)[index++] = pair.second;
    }
    
    return true;
}

bool WindowsProjectManager::createIssue(const ProjectIssue* issue) {
    if (!issue) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Store locally
    if (store_issue_ && !store_issue_(issue)) {
        return false;
    }
    
    project_issues_[issue->project_id].push_back(*issue);
    
    // Update project statistics
    auto proj_it = projects_.find(issue->project_id);
    if (proj_it != projects_.end()) {
        proj_it->second.total_issues++;
        if (issue->status == ISSUE_STATUS_OPEN) {
            proj_it->second.open_issues++;
        }
        proj_it->second.last_activity_timestamp = getCurrentTimestamp();
    }
    
    // Notify callback
    if (issue_callback_) {
        issue_callback_(issue, PROJECT_OPERATION_CREATE);
    }
    
    return true;
}

bool WindowsProjectManager::updateIssue(const ProjectIssue* issue) {
    if (!issue) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Update locally
    if (store_issue_ && !store_issue_(issue)) {
        return false;
    }
    
    auto& issues = project_issues_[issue->project_id];
    auto it = std::find_if(issues.begin(), issues.end(),
        [issue](const ProjectIssue& existing) {
            return strcmp(existing.issue_id, issue->issue_id) == 0;
        });
    
    if (it != issues.end()) {
        *it = *issue;
    } else {
        issues.push_back(*issue);
    }
    
    // Notify callback
    if (issue_callback_) {
        issue_callback_(issue, PROJECT_OPERATION_UPDATE);
    }
    
    return true;
}

bool WindowsProjectManager::deleteIssue(const char* issue_id) {
    if (!issue_id) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Delete locally
    if (delete_issue_ && !delete_issue_(issue_id)) {
        return false;
    }
    
    // Find and remove from local cache
    for (auto& pair : project_issues_) {
        auto& issues = pair.second;
        auto it = std::find_if(issues.begin(), issues.end(),
            [issue_id](const ProjectIssue& issue) {
                return strcmp(issue.issue_id, issue_id) == 0;
            });
        
        if (it != issues.end()) {
            // Notify callback before deletion
            if (issue_callback_) {
                issue_callback_(&*it, PROJECT_OPERATION_DELETE);
            }
            
            issues.erase(it);
            
            // Update project statistics
            auto proj_it = projects_.find(pair.first);
            if (proj_it != projects_.end()) {
                proj_it->second.total_issues--;
                proj_it->second.last_activity_timestamp = getCurrentTimestamp();
            }
            
            break;
        }
    }
    
    return true;
}

bool WindowsProjectManager::getIssue(const char* issue_id, ProjectIssue* issue) {
    if (!issue_id || !issue) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Try local cache first
    for (const auto& pair : project_issues_) {
        const auto& issues = pair.second;
        auto it = std::find_if(issues.begin(), issues.end(),
            [issue_id](const ProjectIssue& existing) {
                return strcmp(existing.issue_id, issue_id) == 0;
            });
        
        if (it != issues.end()) {
            *issue = *it;
            return true;
        }
    }
    
    // Try storage interface
    if (retrieve_issue_) {
        return retrieve_issue_(issue_id, issue);
    }
    
    return false;
}

bool WindowsProjectManager::listIssues(const char* project_id, ProjectIssue** issues, uint32_t* count) {
    if (!project_id || !issues || !count) return false;
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    // Try storage interface first
    if (list_issues_) {
        return list_issues_(project_id, issues, count);
    }
    
    // Use local cache
    auto it = project_issues_.find(project_id);
    if (it == project_issues_.end()) {
        *count = 0;
        *issues = nullptr;
        return true;
    }
    
    const auto& issue_list = it->second;
    *count = static_cast<uint32_t>(issue_list.size());
    
    if (*count == 0) {
        *issues = nullptr;
        return true;
    }
    
    *issues = static_cast<ProjectIssue*>(malloc(*count * sizeof(ProjectIssue)));
    if (!*issues) {
        *count = 0;
        return false;
    }
    
    for (uint32_t i = 0; i < *count; i++) {
        (*issues)[i] = issue_list[i];
    }
    
    return true;
}

bool WindowsProjectManager::assignIssue(const char* issue_id, const char* assignee_id) {
    if (!issue_id || !assignee_id) return false;
    
    ProjectIssue issue;
    if (!getIssue(issue_id, &issue)) {
        return false;
    }
    
    // Add assignee if not already assigned
    for (uint32_t i = 0; i < issue.assignee_count; i++) {
        if (strcmp(issue.assignee_ids[i], assignee_id) == 0) {
            return true; // Already assigned
        }
    }
    
    if (issue.assignee_count < MAX_ASSIGNEES_PER_ISSUE) {
        strcpy_s(issue.assignee_ids[issue.assignee_count], assignee_id);
        issue.assignee_count++;
        issue.updated_timestamp = getCurrentTimestamp();
        
        return updateIssue(&issue);
    }
    
    return false;
}

bool WindowsProjectManager::updateIssueStatus(const char* issue_id, IssueStatus status) {
    if (!issue_id) return false;
    
    ProjectIssue issue;
    if (!getIssue(issue_id, &issue)) {
        return false;
    }
    
    IssueStatus old_status = issue.status;
    issue.status = status;
    issue.updated_timestamp = getCurrentTimestamp();
    
    if (status == ISSUE_STATUS_RESOLVED || status == ISSUE_STATUS_CLOSED) {
        issue.resolved_timestamp = getCurrentTimestamp();
    }
    
    bool result = updateIssue(&issue);
    
    // Update project statistics
    if (result) {
        std::lock_guard<std::mutex> lock(mutex_);
        auto proj_it = projects_.find(issue.project_id);
        if (proj_it != projects_.end()) {
            if (old_status == ISSUE_STATUS_OPEN && status != ISSUE_STATUS_OPEN) {
                proj_it->second.open_issues--;
                proj_it->second.closed_issues++;
            } else if (old_status != ISSUE_STATUS_OPEN && status == ISSUE_STATUS_OPEN) {
                proj_it->second.open_issues++;
                proj_it->second.closed_issues--;
            }
        }
    }
    
    return result;
}

// MARK: - Private Method Implementations

bool WindowsProjectManager::initializeWinsock() {
    WSADATA wsaData;
    int result = WSAStartup(MAKEWORD(2, 2), &wsaData);
    if (result != 0) {
        handleError(PROJECT_ERROR_NETWORK_FAILURE, "WSAStartup failed");
        return false;
    }
    return true;
}

void WindowsProjectManager::cleanupWinsock() {
    WSACleanup();
}

bool WindowsProjectManager::initializeSSL() {
    SSL_library_init();
    SSL_load_error_strings();
    OpenSSL_add_all_algorithms();
    
    ssl_ctx_ = SSL_CTX_new(TLS_client_method());
    if (!ssl_ctx_) {
        return false;
    }
    
    // Set verification mode
    SSL_CTX_set_verify(ssl_ctx_, SSL_VERIFY_PEER, nullptr);
    
    // Load default CA certificates
    SSL_CTX_set_default_verify_paths(ssl_ctx_);
    
    return true;
}

void WindowsProjectManager::cleanupSSL() {
    if (ssl_) {
        SSL_free(ssl_);
        ssl_ = nullptr;
    }
    if (ssl_ctx_) {
        SSL_CTX_free(ssl_ctx_);
        ssl_ctx_ = nullptr;
    }
}

bool WindowsProjectManager::performHandshake() {
    // Create handshake request
    Json::Value request;
    request["device_id"] = config_.device_id;
    request["device_name"] = "Windows Desktop";
    request["protocol_version"] = PROJECT_MANAGEMENT_PROTOCOL_VERSION;
    request["supported_features"] = Json::Value(Json::arrayValue);
    request["supported_features"].append("projects");
    request["supported_features"].append("issues");
    request["supported_features"].append("comments");
    request["supported_features"].append("milestones");
    request["supported_features"].append("attachments");
    request["supports_encryption"] = config_.enable_encryption;
    request["supports_compression"] = config_.enable_compression;
    request["supports_notifications"] = config_.enable_notifications;
    
    std::string request_str = request.toStyledString();
    
    ProjectHeader header = {};
    header.magic = PROJECT_MANAGEMENT_MAGIC;
    header.version = PROJECT_MANAGEMENT_PROTOCOL_VERSION;
    header.message_type = PROJECT_MSG_TYPE_HANDSHAKE;
    header.message_id = generateMessageId();
    header.session_id = 0;
    header.data_length = static_cast<uint32_t>(request_str.length());
    header.timestamp = getCurrentTimestamp();
    header.checksum = calculateChecksum(request_str.c_str(), header.data_length);
    
    if (!sendMessage(&header, request_str.c_str())) {
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Failed to send handshake request");
        return false;
    }
    
    // Receive response
    ProjectHeader response_header;
    void* response_data = nullptr;
    
    if (!receiveMessage(&response_header, &response_data)) {
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Failed to receive handshake response");
        return false;
    }
    
    if (response_header.message_type != PROJECT_MSG_TYPE_HANDSHAKE) {
        free(response_data);
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Invalid handshake response");
        return false;
    }
    
    // Parse response
    Json::Value response;
    Json::Reader reader;
    if (!reader.parse(static_cast<char*>(response_data), response)) {
        free(response_data);
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Failed to parse handshake response");
        return false;
    }
    
    if (!response["handshake_accepted"].asBool()) {
        free(response_data);
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Handshake rejected");
        return false;
    }
    
    free(response_data);
    return true;
}

bool WindowsProjectManager::authenticate() {
    // Create auth request
    Json::Value request;
    request["user_id"] = config_.user_id;
    request["auth_token"] = config_.auth_token;
    request["device_signature"] = generateDeviceSignature();
    request["timestamp"] = static_cast<Json::Int64>(getCurrentTimestamp());
    
    std::string request_str = request.toStyledString();
    
    ProjectHeader header = {};
    header.magic = PROJECT_MANAGEMENT_MAGIC;
    header.version = PROJECT_MANAGEMENT_PROTOCOL_VERSION;
    header.message_type = PROJECT_MSG_TYPE_AUTH;
    header.message_id = generateMessageId();
    header.session_id = 0;
    header.data_length = static_cast<uint32_t>(request_str.length());
    header.timestamp = getCurrentTimestamp();
    header.checksum = calculateChecksum(request_str.c_str(), header.data_length);
    
    if (!sendMessage(&header, request_str.c_str())) {
        handleError(PROJECT_ERROR_AUTH_FAILED, "Failed to send auth request");
        return false;
    }
    
    // Receive response
    ProjectHeader response_header;
    void* response_data = nullptr;
    
    if (!receiveMessage(&response_header, &response_data)) {
        handleError(PROJECT_ERROR_AUTH_FAILED, "Failed to receive auth response");
        return false;
    }
    
    if (response_header.message_type != PROJECT_MSG_TYPE_AUTH) {
        free(response_data);
        handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Invalid auth response");
        return false;
    }
    
    // Parse response
    Json::Value response;
    Json::Reader reader;
    if (!reader.parse(static_cast<char*>(response_data), response)) {
        free(response_data);
        handleError(PROJECT_ERROR_AUTH_FAILED, "Failed to parse auth response");
        return false;
    }
    
    if (!response["auth_success"].asBool()) {
        free(response_data);
        handleError(PROJECT_ERROR_AUTH_FAILED, "Authentication failed");
        return false;
    }
    
    // Store session info
    session_id_ = response_header.session_id;
    session_token_ = response["session_token"].asString();
    
    free(response_data);
    return true;
}

bool WindowsProjectManager::sendMessage(const ProjectHeader* header, const void* data) {
    // Send header
    int bytes_sent;
    if (ssl_) {
        bytes_sent = SSL_write(ssl_, header, sizeof(ProjectHeader));
    } else {
        bytes_sent = send(socket_fd_, reinterpret_cast<const char*>(header), sizeof(ProjectHeader), 0);
    }
    
    if (bytes_sent != sizeof(ProjectHeader)) {
        return false;
    }
    
    // Send data if present
    if (data && header->data_length > 0) {
        if (ssl_) {
            bytes_sent = SSL_write(ssl_, data, header->data_length);
        } else {
            bytes_sent = send(socket_fd_, static_cast<const char*>(data), header->data_length, 0);
        }
        
        if (bytes_sent != static_cast<int>(header->data_length)) {
            return false;
        }
    }
    
    return true;
}

bool WindowsProjectManager::receiveMessage(ProjectHeader* header, void** data) {
    // Receive header
    int bytes_received;
    if (ssl_) {
        bytes_received = SSL_read(ssl_, header, sizeof(ProjectHeader));
    } else {
        bytes_received = recv(socket_fd_, reinterpret_cast<char*>(header), sizeof(ProjectHeader), 0);
    }
    
    if (bytes_received != sizeof(ProjectHeader)) {
        return false;
    }
    
    // Validate header
    if (header->magic != PROJECT_MANAGEMENT_MAGIC || header->version != PROJECT_MANAGEMENT_PROTOCOL_VERSION) {
        return false;
    }
    
    // Receive data if present
    *data = nullptr;
    if (header->data_length > 0) {
        *data = malloc(header->data_length + 1); // +1 for null terminator
        if (!*data) {
            return false;
        }
        
        if (ssl_) {
            bytes_received = SSL_read(ssl_, *data, header->data_length);
        } else {
            bytes_received = recv(socket_fd_, static_cast<char*>(*data), header->data_length, 0);
        }
        
        if (bytes_received != static_cast<int>(header->data_length)) {
            free(*data);
            *data = nullptr;
            return false;
        }
        
        // Add null terminator for string data
        static_cast<char*>(*data)[header->data_length] = '\0';
        
        // Verify checksum
        uint32_t calculated_checksum = calculateChecksum(*data, header->data_length);
        if (calculated_checksum != header->checksum) {
            free(*data);
            *data = nullptr;
            return false;
        }
    }
    
    return true;
}

void WindowsProjectManager::sendHeartbeat() {
    ProjectHeader header = {};
    header.magic = PROJECT_MANAGEMENT_MAGIC;
    header.version = PROJECT_MANAGEMENT_PROTOCOL_VERSION;
    header.message_type = PROJECT_MSG_TYPE_HEARTBEAT;
    header.message_id = generateMessageId();
    header.session_id = session_id_;
    header.data_length = 0;
    header.checksum = 0;
    header.timestamp = getCurrentTimestamp();
    
    if (!sendMessage(&header, nullptr)) {
        // Heartbeat failed, disconnect
        disconnect();
    }
}

void WindowsProjectManager::processTaskQueue() {
    while (!shutdown_requested_) {
        std::unique_lock<std::mutex> lock(task_mutex_);
        task_condition_.wait(lock, [this] { return !task_queue_.empty() || shutdown_requested_; });
        
        if (shutdown_requested_) {
            break;
        }
        
        if (!task_queue_.empty()) {
            TaskContext task = task_queue_.front();
            task_queue_.pop();
            lock.unlock();
            
            try {
                task.task();
            } catch (const std::exception& e) {
                handleError(PROJECT_ERROR_PROTOCOL_ERROR, "Task execution failed: " + std::string(e.what()));
            }
        }
    }
}

void WindowsProjectManager::syncThreadFunc() {
    while (!shutdown_requested_) {
        std::unique_lock<std::mutex> lock(mutex_);
        condition_.wait_for(lock, std::chrono::milliseconds(config_.sync_interval),
            [this] { return shutdown_requested_; });
        
        if (shutdown_requested_) {
            break;
        }
        
        if (config_.auto_sync_enabled && is_connected_) {
            lock.unlock();
            syncAll();
        }
    }
}

void WindowsProjectManager::heartbeatThreadFunc() {
    while (!shutdown_requested_) {
        std::this_thread::sleep_for(std::chrono::milliseconds(config_.heartbeat_interval));
        
        if (shutdown_requested_) {
            break;
        }
        
        if (is_connected_) {
            sendHeartbeat();
        }
    }
}

void WindowsProjectManager::notificationThreadFunc() {
    while (!shutdown_requested_) {
        std::this_thread::sleep_for(std::chrono::milliseconds(5000)); // Check every 5 seconds
        
        if (shutdown_requested_) {
            break;
        }
        
        // Process pending notifications
        std::lock_guard<std::mutex> lock(mutex_);
        while (!notifications_.empty() && notification_callback_) {
            ProjectNotification notification = notifications_.front();
            notifications_.pop();
            notification_callback_(&notification);
        }
    }
}

void WindowsProjectManager::loadLocalData() {
    std::string projects_file = std::string(config_.local_storage_path) + "\\projects.json";
    std::ifstream file(projects_file);
    
    if (!file.is_open()) {
        return;
    }
    
    Json::Value root;
    Json::Reader reader;
    
    if (!reader.parse(file, root)) {
        file.close();
        return;
    }
    
    file.close();
    
    // Load projects
    if (root.isMember("projects") && root["projects"].isArray()) {
        for (const auto& project_json : root["projects"]) {
            Project project;
            if (jsonToProject(project_json, &project)) {
                projects_[project.project_id] = project;
            }
        }
    }
    
    // Load issues
    if (root.isMember("issues") && root["issues"].isObject()) {
        for (const auto& project_id : root["issues"].getMemberNames()) {
            const auto& issues_json = root["issues"][project_id];
            if (issues_json.isArray()) {
                std::vector<ProjectIssue> issues;
                for (const auto& issue_json : issues_json) {
                    ProjectIssue issue;
                    if (jsonToIssue(issue_json, &issue)) {
                        issues.push_back(issue);
                    }
                }
                project_issues_[project_id] = issues;
            }
        }
    }
}

void WindowsProjectManager::saveLocalData() {
    Json::Value root;
    
    // Save projects
    Json::Value projects_json(Json::arrayValue);
    for (const auto& pair : projects_) {
        projects_json.append(projectToJson(&pair.second));
    }
    root["projects"] = projects_json;
    
    // Save issues
    Json::Value issues_json(Json::objectValue);
    for (const auto& pair : project_issues_) {
        Json::Value project_issues_json(Json::arrayValue);
        for (const auto& issue : pair.second) {
            project_issues_json.append(issueToJson(&issue));
        }
        issues_json[pair.first] = project_issues_json;
    }
    root["issues"] = issues_json;
    
    // Write to file
    std::string projects_file = std::string(config_.local_storage_path) + "\\projects.json";
    std::ofstream file(projects_file);
    
    if (file.is_open()) {
        Json::StreamWriterBuilder builder;
        builder["indentation"] = "  ";
        std::unique_ptr<Json::StreamWriter> writer(builder.newStreamWriter());
        writer->write(root, &file);
        file.close();
    }
}

void WindowsProjectManager::notifyStatusChange() {
    if (status_callback_) {
        float progress = getSyncProgress();
        status_callback_(status_, progress);
    }
}

void WindowsProjectManager::handleError(ProjectError error, const std::string& message) {
    status_ = PROJECT_STATUS_ON_HOLD; // Error status
    
    if (error_callback_) {
        error_callback_(error, message.c_str());
    }
    
    std::cout << "Project error: " << message << std::endl;
}

uint32_t WindowsProjectManager::generateMessageId() {
    static uint32_t counter = 0;
    return ++counter;
}

uint64_t WindowsProjectManager::getCurrentTimestamp() {
    auto now = std::chrono::system_clock::now();
    auto duration = now.time_since_epoch();
    return std::chrono::duration_cast<std::chrono::milliseconds>(duration).count();
}

uint32_t WindowsProjectManager::calculateChecksum(const void* data, uint32_t length) {
    const uint8_t* bytes = static_cast<const uint8_t*>(data);
    uint32_t checksum = 0;
    
    for (uint32_t i = 0; i < length; i++) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
}

std::string WindowsProjectManager::generateDeviceSignature() {
    return std::string(config_.device_id) + "_" + std::to_string(getCurrentTimestamp());
}

Json::Value WindowsProjectManager::projectToJson(const Project* project) {
    Json::Value json;
    json["project_id"] = project->project_id;
    json["name"] = project->name;
    json["description"] = project->description;
    json["owner_id"] = project->owner_id;
    json["status"] = static_cast<int>(project->status);
    json["priority"] = static_cast<int>(project->priority);
    json["created_timestamp"] = static_cast<Json::Int64>(project->created_timestamp);
    json["updated_timestamp"] = static_cast<Json::Int64>(project->updated_timestamp);
    json["start_date"] = static_cast<Json::Int64>(project->start_date);
    json["end_date"] = static_cast<Json::Int64>(project->end_date);
    json["is_public"] = project->is_public;
    json["allow_issues"] = project->allow_issues;
    json["enable_notifications"] = project->enable_notifications;
    return json;
}

Json::Value WindowsProjectManager::issueToJson(const ProjectIssue* issue) {
    Json::Value json;
    json["issue_id"] = issue->issue_id;
    json["project_id"] = issue->project_id;
    json["title"] = issue->title;
    json["description"] = issue->description;
    json["type"] = static_cast<int>(issue->type);
    json["status"] = static_cast<int>(issue->status);
    json["priority"] = static_cast<int>(issue->priority);
    json["reporter_id"] = issue->reporter_id;
    json["created_timestamp"] = static_cast<Json::Int64>(issue->created_timestamp);
    json["updated_timestamp"] = static_cast<Json::Int64>(issue->updated_timestamp);
    json["due_date"] = static_cast<Json::Int64>(issue->due_date);
    json["estimated_hours"] = issue->estimated_hours;
    json["logged_hours"] = issue->logged_hours;
    json["progress_percentage"] = issue->progress_percentage;
    return json;
}

bool WindowsProjectManager::jsonToProject(const Json::Value& json, Project* project) {
    if (!json.isObject()) return false;
    
    memset(project, 0, sizeof(Project));
    
    if (json.isMember("project_id")) strcpy_s(project->project_id, json["project_id"].asString().c_str());
    if (json.isMember("name")) strcpy_s(project->name, json["name"].asString().c_str());
    if (json.isMember("description")) strcpy_s(project->description, json["description"].asString().c_str());
    if (json.isMember("owner_id")) strcpy_s(project->owner_id, json["owner_id"].asString().c_str());
    if (json.isMember("status")) project->status = static_cast<ProjectStatus>(json["status"].asInt());
    if (json.isMember("priority")) project->priority = static_cast<ProjectPriority>(json["priority"].asInt());
    if (json.isMember("created_timestamp")) project->created_timestamp = json["created_timestamp"].asUInt64();
    if (json.isMember("updated_timestamp")) project->updated_timestamp = json["updated_timestamp"].asUInt64();
    if (json.isMember("start_date")) project->start_date = json["start_date"].asUInt64();
    if (json.isMember("end_date")) project->end_date = json["end_date"].asUInt64();
    if (json.isMember("is_public")) project->is_public = json["is_public"].asBool();
    if (json.isMember("allow_issues")) project->allow_issues = json["allow_issues"].asBool();
    if (json.isMember("enable_notifications")) project->enable_notifications = json["enable_notifications"].asBool();
    
    return true;
}

bool WindowsProjectManager::jsonToIssue(const Json::Value& json, ProjectIssue* issue) {
    if (!json.isObject()) return false;
    
    memset(issue, 0, sizeof(ProjectIssue));
    
    if (json.isMember("issue_id")) strcpy_s(issue->issue_id, json["issue_id"].asString().c_str());
    if (json.isMember("project_id")) strcpy_s(issue->project_id, json["project_id"].asString().c_str());
    if (json.isMember("title")) strcpy_s(issue->title, json["title"].asString().c_str());
    if (json.isMember("description")) strcpy_s(issue->description, json["description"].asString().c_str());
    if (json.isMember("type")) issue->type = static_cast<IssueType>(json["type"].asInt());
    if (json.isMember("status")) issue->status = static_cast<IssueStatus>(json["status"].asInt());
    if (json.isMember("priority")) issue->priority = static_cast<IssuePriority>(json["priority"].asInt());
    if (json.isMember("reporter_id")) strcpy_s(issue->reporter_id, json["reporter_id"].asString().c_str());
    if (json.isMember("created_timestamp")) issue->created_timestamp = json["created_timestamp"].asUInt64();
    if (json.isMember("updated_timestamp")) issue->updated_timestamp = json["updated_timestamp"].asUInt64();
    if (json.isMember("due_date")) issue->due_date = json["due_date"].asUInt64();
    if (json.isMember("estimated_hours")) issue->estimated_hours = json["estimated_hours"].asUInt();
    if (json.isMember("logged_hours")) issue->logged_hours = json["logged_hours"].asUInt();
    if (json.isMember("progress_percentage")) issue->progress_percentage = json["progress_percentage"].asFloat();
    
    return true;
}

float WindowsProjectManager::getSyncProgress() const {
    if (pending_sync_items_ == 0) return 1.0f;
    return static_cast<float>(synced_items_) / (synced_items_ + pending_sync_items_);
}

void WindowsProjectManager::getStats(uint32_t* total_projects, uint32_t* total_issues, uint32_t* pending_sync) {
    if (total_projects) *total_projects = static_cast<uint32_t>(projects_.size());
    
    if (total_issues) {
        uint32_t count = 0;
        for (const auto& pair : project_issues_) {
            count += static_cast<uint32_t>(pair.second.size());
        }
        *total_issues = count;
    }
    
    if (pending_sync) *pending_sync = pending_sync_items_;
}

// Additional method implementations would continue here...
// For brevity, I'm showing the key methods. The full implementation would include
// all remaining methods like syncAll, addComment, createMilestone, etc.

bool WindowsProjectManager::syncAll() {
    if (!is_connected_) {
        if (!connect()) {
            return false;
        }
    }
    
    std::lock_guard<std::mutex> lock(mutex_);
    
    status_ = PROJECT_STATUS_ACTIVE; // Syncing status
    notifyStatusChange();
    
    bool success = true;
    
    // Sync projects
    for (const auto& pair : projects_) {
        // Implementation would send sync messages for each project
        // This is a simplified version
        synced_items_++;
    }
    
    // Sync issues
    for (const auto& pair : project_issues_) {
        for (const auto& issue : pair.second) {
            // Implementation would send sync messages for each issue
            synced_items_++;
        }
    }
    
    status_ = success ? PROJECT_STATUS_COMPLETED : PROJECT_STATUS_ON_HOLD;
    notifyStatusChange();
    
    if (sync_complete_callback_) {
        uint32_t total_projects = static_cast<uint32_t>(projects_.size());
        uint32_t total_issues = 0;
        for (const auto& pair : project_issues_) {
            total_issues += static_cast<uint32_t>(pair.second.size());
        }
        sync_complete_callback_(total_projects, total_issues, failed_items_);
    }
    
    return success;
}

// MARK: - C API Implementation

extern "C" {

ProjectManager* project_manager_create(const ProjectManagerConfiguration* config) {
    try {
        return reinterpret_cast<ProjectManager*>(new WindowsProjectManager(config));
    } catch (const std::exception& e) {
        std::cout << "Failed to create project manager: " << e.what() << std::endl;
        return nullptr;
    }
}

void project_manager_destroy(ProjectManager* manager) {
    if (manager) {
        delete reinterpret_cast<WindowsProjectManager*>(manager);
    }
}

bool project_manager_start(ProjectManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->start();
}

void project_manager_stop(ProjectManager* manager) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->stop();
    }
}

bool project_manager_connect(ProjectManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->connect();
}

void project_manager_disconnect(ProjectManager* manager) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->disconnect();
    }
}

bool project_manager_is_connected(const ProjectManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<const WindowsProjectManager*>(manager)->isConnected();
}

bool project_manager_create_project(ProjectManager* manager, const Project* project) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->createProject(project);
}

bool project_manager_update_project(ProjectManager* manager, const Project* project) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->updateProject(project);
}

bool project_manager_delete_project(ProjectManager* manager, const char* project_id) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->deleteProject(project_id);
}

bool project_manager_get_project(ProjectManager* manager, const char* project_id, Project* project) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->getProject(project_id, project);
}

bool project_manager_list_projects(ProjectManager* manager, Project** projects, uint32_t* count) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->listProjects(projects, count);
}

bool project_manager_create_issue(ProjectManager* manager, const ProjectIssue* issue) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->createIssue(issue);
}

bool project_manager_update_issue(ProjectManager* manager, const ProjectIssue* issue) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->updateIssue(issue);
}

bool project_manager_delete_issue(ProjectManager* manager, const char* issue_id) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->deleteIssue(issue_id);
}

bool project_manager_get_issue(ProjectManager* manager, const char* issue_id, ProjectIssue* issue) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->getIssue(issue_id, issue);
}

bool project_manager_list_issues(ProjectManager* manager, const char* project_id, ProjectIssue** issues, uint32_t* count) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->listIssues(project_id, issues, count);
}

bool project_manager_assign_issue(ProjectManager* manager, const char* issue_id, const char* assignee_id) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->assignIssue(issue_id, assignee_id);
}

bool project_manager_update_issue_status(ProjectManager* manager, const char* issue_id, IssueStatus status) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->updateIssueStatus(issue_id, status);
}

ProjectStatus project_manager_get_status(const ProjectManager* manager) {
    if (!manager) return PROJECT_STATUS_PLANNING;
    return reinterpret_cast<const WindowsProjectManager*>(manager)->getStatus();
}

float project_manager_get_sync_progress(const ProjectManager* manager) {
    if (!manager) return 0.0f;
    return reinterpret_cast<const WindowsProjectManager*>(manager)->getSyncProgress();
}

void project_manager_get_stats(const ProjectManager* manager, uint32_t* total_projects, uint32_t* total_issues, uint32_t* pending_sync) {
    if (manager) {
        reinterpret_cast<const WindowsProjectManager*>(manager)->getStats(total_projects, total_issues, pending_sync);
    }
}

bool project_manager_sync_all(ProjectManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->syncAll();
}

bool project_manager_sync_project(ProjectManager* manager, const char* project_id) {
    if (!manager) return false;
    return reinterpret_cast<WindowsProjectManager*>(manager)->syncProject(project_id);
}

void project_manager_set_status_callback(ProjectManager* manager, ProjectStatusCallback callback) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->setStatusCallback(callback);
    }
}

void project_manager_set_project_callback(ProjectManager* manager, ProjectDataCallback callback) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->setProjectCallback(callback);
    }
}

void project_manager_set_issue_callback(ProjectManager* manager, IssueDataCallback callback) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->setIssueCallback(callback);
    }
}

void project_manager_set_notification_callback(ProjectManager* manager, NotificationCallback callback) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->setNotificationCallback(callback);
    }
}

void project_manager_set_error_callback(ProjectManager* manager, ProjectErrorCallback callback) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->setErrorCallback(callback);
    }
}

void project_manager_set_sync_complete_callback(ProjectManager* manager, SyncCompleteCallback callback) {
    if (manager) {
        reinterpret_cast<WindowsProjectManager*>(manager)->setSyncCompleteCallback(callback);
    }
}

// Utility functions
void project_generate_id(char* id, uint32_t length) {
    if (!id || length < PROJECT_ID_LENGTH) return;
    
    static const char chars[] = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    static std::random_device rd;
    static std::mt19937 gen(rd());
    static std::uniform_int_distribution<> dis(0, sizeof(chars) - 2);
    
    for (uint32_t i = 0; i < length - 1; i++) {
        id[i] = chars[dis(gen)];
    }
    id[length - 1] = '\0';
}

uint64_t project_get_current_timestamp() {
    auto now = std::chrono::system_clock::now();
    auto duration = now.time_since_epoch();
    return std::chrono::duration_cast<std::chrono::milliseconds>(duration).count();
}

uint32_t project_calculate_checksum(const void* data, uint32_t length) {
    if (!data || length == 0) return 0;
    
    const uint8_t* bytes = static_cast<const uint8_t*>(data);
    uint32_t checksum = 0;
    
    for (uint32_t i = 0; i < length; i++) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
}

bool project_validate_project_data(const Project* project) {
    if (!project) return false;
    
    // Check required fields
    if (strlen(project->project_id) == 0 || strlen(project->name) == 0 || strlen(project->owner_id) == 0) {
        return false;
    }
    
    // Check field lengths
    if (strlen(project->project_id) >= PROJECT_ID_LENGTH ||
        strlen(project->name) >= MAX_PROJECT_NAME_LENGTH ||
        strlen(project->description) >= MAX_PROJECT_DESCRIPTION_LENGTH ||
        strlen(project->owner_id) >= PROJECT_ID_LENGTH) {
        return false;
    }
    
    // Check enum values
    if (project->status < PROJECT_STATUS_PLANNING || project->status > PROJECT_STATUS_ARCHIVED) {
        return false;
    }
    
    if (project->priority < PROJECT_PRIORITY_LOW || project->priority > PROJECT_PRIORITY_CRITICAL) {
        return false;
    }
    
    return true;
}

bool project_validate_issue_data(const ProjectIssue* issue) {
    if (!issue) return false;
    
    // Check required fields
    if (strlen(issue->issue_id) == 0 || strlen(issue->project_id) == 0 || 
        strlen(issue->title) == 0 || strlen(issue->reporter_id) == 0) {
        return false;
    }
    
    // Check field lengths
    if (strlen(issue->issue_id) >= PROJECT_ID_LENGTH ||
        strlen(issue->project_id) >= PROJECT_ID_LENGTH ||
        strlen(issue->title) >= MAX_ISSUE_TITLE_LENGTH ||
        strlen(issue->description) >= MAX_ISSUE_DESCRIPTION_LENGTH ||
        strlen(issue->reporter_id) >= PROJECT_ID_LENGTH) {
        return false;
    }
    
    // Check enum values
    if (issue->type < ISSUE_TYPE_BUG || issue->type > ISSUE_TYPE_EPIC) {
        return false;
    }
    
    if (issue->status < ISSUE_STATUS_OPEN || issue->status > ISSUE_STATUS_CLOSED) {
        return false;
    }
    
    if (issue->priority < ISSUE_PRIORITY_LOW || issue->priority > ISSUE_PRIORITY_CRITICAL) {
        return false;
    }
    
    // Check progress percentage
    if (issue->progress_percentage < 0.0f || issue->progress_percentage > 100.0f) {
        return false;
    }
    
    return true;
}

float project_calculate_progress(const char* project_id, ProjectManager* manager) {
    if (!project_id || !manager) return 0.0f;
    
    ProjectIssue* issues = nullptr;
    uint32_t count = 0;
    
    if (!project_manager_list_issues(manager, project_id, &issues, &count) || count == 0) {
        return 0.0f;
    }
    
    float total_progress = 0.0f;
    for (uint32_t i = 0; i < count; i++) {
        total_progress += issues[i].progress_percentage;
    }
    
    free(issues);
    return total_progress / count;
}

const char* project_error_to_string(ProjectError error) {
    switch (error) {
        case PROJECT_ERROR_NONE: return "No error";
        case PROJECT_ERROR_NETWORK_FAILURE: return "Network failure";
        case PROJECT_ERROR_AUTH_FAILED: return "Authentication failed";
        case PROJECT_ERROR_PROTOCOL_ERROR: return "Protocol error";
        case PROJECT_ERROR_DATA_CORRUPTION: return "Data corruption";
        case PROJECT_ERROR_STORAGE_ERROR: return "Storage error";
        case PROJECT_ERROR_PERMISSION_DENIED: return "Permission denied";
        case PROJECT_ERROR_INVALID_DATA: return "Invalid data";
        case PROJECT_ERROR_VERSION_MISMATCH: return "Version mismatch";
        case PROJECT_ERROR_TIMEOUT: return "Timeout";
        default: return "Unknown error";
    }
}

const char* project_status_to_string(ProjectStatus status) {
    switch (status) {
        case PROJECT_STATUS_PLANNING: return "Planning";
        case PROJECT_STATUS_ACTIVE: return "Active";
        case PROJECT_STATUS_ON_HOLD: return "On Hold";
        case PROJECT_STATUS_COMPLETED: return "Completed";
        case PROJECT_STATUS_CANCELLED: return "Cancelled";
        case PROJECT_STATUS_ARCHIVED: return "Archived";
        default: return "Unknown";
    }
}

const char* issue_status_to_string(IssueStatus status) {
    switch (status) {
        case ISSUE_STATUS_OPEN: return "Open";
        case ISSUE_STATUS_IN_PROGRESS: return "In Progress";
        case ISSUE_STATUS_RESOLVED: return "Resolved";
        case ISSUE_STATUS_CLOSED: return "Closed";
        case ISSUE_STATUS_REOPENED: return "Reopened";
        default: return "Unknown";
    }
}

} // extern "C"