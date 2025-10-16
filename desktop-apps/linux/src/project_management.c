#include "project_management.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <openssl/ssl.h>
#include <openssl/err.h>
#include <json-c/json.h>
#include <signal.h>
#include <errno.h>
#include <time.h>
#include <sys/stat.h>
#include <dirent.h>

// Internal structures
typedef struct {
    int socket_fd;
    SSL *ssl;
    SSL_CTX *ssl_ctx;
    struct sockaddr_in server_addr;
    int is_connected;
    uint32_t session_id;
    char session_token[256];
} ConnectionContext;

typedef struct {
    ProjectOperation operation;
    char project_id[PROJECT_ID_MAX_LENGTH];
    char issue_id[ISSUE_ID_MAX_LENGTH];
    void *data;
    size_t data_size;
} BatchContext;

// Global state
static ProjectManager *g_manager = NULL;
static volatile int g_shutdown_requested = 0;

// Signal handler
static void signal_handler(int sig) {
    if (sig == SIGINT || sig == SIGTERM) {
        g_shutdown_requested = 1;
    }
}

// Private function declarations
static void *sync_thread(void *arg);
static void *heartbeat_thread(void *arg);
static int init_ssl(ProjectManager *manager);
static void cleanup_ssl(ProjectManager *manager);
static int perform_handshake(ProjectManager *manager);
static int authenticate(ProjectManager *manager);
static int send_message(ProjectManager *manager, ProjectMessageType type, const void *data, size_t data_size);
static int receive_message(ProjectManager *manager, ProjectMessageType *type, void **data, size_t *data_size);
static void send_heartbeat(ProjectManager *manager);
static void load_local_data(ProjectManager *manager);
static void save_local_data(ProjectManager *manager);
static void mark_project_dirty(ProjectManager *manager, const char *project_id);
static void notify_status_change(ProjectManager *manager);
static void handle_error(ProjectManager *manager, ProjectError error, const char *message);
static uint32_t generate_message_id(void);
static uint64_t get_current_timestamp(void);
static uint32_t calculate_checksum(const void *data, size_t size);
static void generate_device_signature(char *signature, size_t size);
static json_object *project_to_json(const Project *project);
static Project *json_to_project(json_object *json);
static json_object *issue_to_json(const ProjectIssue *issue);
static ProjectIssue *json_to_issue(json_object *json);

// Project Manager structure
struct ProjectManager {
    ProjectManagerConfiguration config;
    ProjectStatus status;
    int is_running;
    int is_connected;
    float sync_progress;
    
    // Session info
    uint32_t session_id;
    char session_token[256];
    
    // Local data storage
    Project **projects;
    size_t project_count;
    size_t project_capacity;
    
    ProjectIssue **issues;
    size_t issue_count;
    size_t issue_capacity;
    
    IssueComment **comments;
    size_t comment_count;
    size_t comment_capacity;
    
    ProjectMilestone **milestones;
    size_t milestone_count;
    size_t milestone_capacity;
    
    ProjectMember **members;
    size_t member_count;
    size_t member_capacity;
    
    ProjectNotification **notifications;
    size_t notification_count;
    size_t notification_capacity;
    
    // Sync state
    uint64_t last_sync_timestamp;
    uint32_t pending_sync_items;
    uint32_t synced_items;
    uint32_t failed_items;
    
    // Network
    ConnectionContext connection;
    
    // Threading
    pthread_t sync_thread_id;
    pthread_t heartbeat_thread_id;
    pthread_mutex_t data_mutex;
    pthread_cond_t sync_cond;
    int sync_thread_running;
    int heartbeat_thread_running;
    
    // Callbacks
    ProjectStatusCallback status_callback;
    ProjectDataCallback project_callback;
    IssueDataCallback issue_callback;
    NotificationCallback notification_callback;
    ProjectErrorCallback error_callback;
    SyncCompleteCallback sync_complete_callback;
    
    // Storage interface
    StoreProjectCallback store_project;
    RetrieveProjectCallback retrieve_project;
    DeleteProjectCallback delete_project;
    ListProjectsCallback list_projects;
    
    StoreIssueCallback store_issue;
    RetrieveIssueCallback retrieve_issue;
    DeleteIssueCallback delete_issue;
    ListIssuesCallback list_issues;
    
    StoreCommentCallback store_comment;
    RetrieveCommentsCallback retrieve_comments;
    DeleteCommentCallback delete_comment;
    
    StoreAttachmentCallback store_attachment;
    RetrieveAttachmentCallback retrieve_attachment;
    DeleteAttachmentCallback delete_attachment;
};

// Public API Implementation

ProjectManager *project_manager_create(const ProjectManagerConfiguration *config) {
    ProjectManager *manager = calloc(1, sizeof(ProjectManager));
    if (!manager) {
        return NULL;
    }
    
    // Copy configuration
    if (config) {
        manager->config = *config;
    } else {
        // Set default configuration
        strcpy(manager->config.server_url, "localhost");
        manager->config.server_port = 8080;
        strcpy(manager->config.user_id, "linux_user");
        strcpy(manager->config.auth_token, "token");
        strcpy(manager->config.device_id, "linux_device");
        manager->config.connection_timeout = 30000;
        manager->config.heartbeat_interval = 30000;
        manager->config.sync_interval = 300000;
        manager->config.max_retries = 3;
        manager->config.enable_encryption = 1;
        manager->config.enable_compression = 1;
        manager->config.enable_notifications = 1;
        manager->config.enable_offline_mode = 1;
        manager->config.auto_sync_enabled = 1;
        strcpy(manager->config.local_storage_path, "~/.taishanglaojun/project_data");
        manager->config.max_storage_size = 1024 * 1024 * 1024; // 1GB
        manager->config.cache_retention_days = 30;
        manager->config.show_completed_issues = 0;
        manager->config.group_by_milestone = 1;
        manager->config.items_per_page = 50;
    }
    
    // Initialize data arrays
    manager->project_capacity = 100;
    manager->projects = calloc(manager->project_capacity, sizeof(Project*));
    
    manager->issue_capacity = 1000;
    manager->issues = calloc(manager->issue_capacity, sizeof(ProjectIssue*));
    
    manager->comment_capacity = 5000;
    manager->comments = calloc(manager->comment_capacity, sizeof(IssueComment*));
    
    manager->milestone_capacity = 100;
    manager->milestones = calloc(manager->milestone_capacity, sizeof(ProjectMilestone*));
    
    manager->member_capacity = 1000;
    manager->members = calloc(manager->member_capacity, sizeof(ProjectMember*));
    
    manager->notification_capacity = 1000;
    manager->notifications = calloc(manager->notification_capacity, sizeof(ProjectNotification*));
    
    // Initialize mutex and condition variable
    pthread_mutex_init(&manager->data_mutex, NULL);
    pthread_cond_init(&manager->sync_cond, NULL);
    
    // Initialize connection
    manager->connection.socket_fd = -1;
    manager->connection.is_connected = 0;
    
    // Set initial status
    manager->status = PROJECT_STATUS_PLANNING;
    manager->is_running = 0;
    manager->is_connected = 0;
    manager->sync_progress = 0.0f;
    
    // Set up signal handlers
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    g_manager = manager;
    
    printf("Linux Project Manager created\n");
    return manager;
}

void project_manager_destroy(ProjectManager *manager) {
    if (!manager) return;
    
    // Stop if running
    if (manager->is_running) {
        project_manager_stop(manager);
    }
    
    // Free data arrays
    for (size_t i = 0; i < manager->project_count; i++) {
        free(manager->projects[i]);
    }
    free(manager->projects);
    
    for (size_t i = 0; i < manager->issue_count; i++) {
        free(manager->issues[i]);
    }
    free(manager->issues);
    
    for (size_t i = 0; i < manager->comment_count; i++) {
        free(manager->comments[i]);
    }
    free(manager->comments);
    
    for (size_t i = 0; i < manager->milestone_count; i++) {
        free(manager->milestones[i]);
    }
    free(manager->milestones);
    
    for (size_t i = 0; i < manager->member_count; i++) {
        free(manager->members[i]);
    }
    free(manager->members);
    
    for (size_t i = 0; i < manager->notification_count; i++) {
        free(manager->notifications[i]);
    }
    free(manager->notifications);
    
    // Cleanup threading
    pthread_mutex_destroy(&manager->data_mutex);
    pthread_cond_destroy(&manager->sync_cond);
    
    free(manager);
    g_manager = NULL;
    
    printf("Linux Project Manager destroyed\n");
}

int project_manager_start(ProjectManager *manager) {
    if (!manager || manager->is_running) {
        return 0;
    }
    
    // Create storage directory
    char expanded_path[512];
    if (manager->config.local_storage_path[0] == '~') {
        snprintf(expanded_path, sizeof(expanded_path), "%s%s", 
                getenv("HOME"), manager->config.local_storage_path + 1);
    } else {
        strcpy(expanded_path, manager->config.local_storage_path);
    }
    
    struct stat st = {0};
    if (stat(expanded_path, &st) == -1) {
        mkdir(expanded_path, 0755);
    }
    
    // Load local data
    load_local_data(manager);
    
    // Initialize SSL if encryption is enabled
    if (manager->config.enable_encryption) {
        if (!init_ssl(manager)) {
            handle_error(manager, PROJECT_ERROR_NETWORK_FAILURE, "Failed to initialize SSL");
            return 0;
        }
    }
    
    // Start sync thread if auto-sync is enabled
    if (manager->config.auto_sync_enabled) {
        manager->sync_thread_running = 1;
        if (pthread_create(&manager->sync_thread_id, NULL, sync_thread, manager) != 0) {
            handle_error(manager, PROJECT_ERROR_STORAGE_ERROR, "Failed to create sync thread");
            return 0;
        }
    }
    
    // Start heartbeat thread
    manager->heartbeat_thread_running = 1;
    if (pthread_create(&manager->heartbeat_thread_id, NULL, heartbeat_thread, manager) != 0) {
        handle_error(manager, PROJECT_ERROR_STORAGE_ERROR, "Failed to create heartbeat thread");
        return 0;
    }
    
    manager->is_running = 1;
    manager->status = PROJECT_STATUS_ACTIVE;
    notify_status_change(manager);
    
    printf("Project manager started\n");
    return 1;
}

int project_manager_stop(ProjectManager *manager) {
    if (!manager || !manager->is_running) {
        return 0;
    }
    
    // Signal threads to stop
    manager->sync_thread_running = 0;
    manager->heartbeat_thread_running = 0;
    
    // Wake up sync thread
    pthread_cond_signal(&manager->sync_cond);
    
    // Wait for threads to finish
    if (manager->config.auto_sync_enabled) {
        pthread_join(manager->sync_thread_id, NULL);
    }
    pthread_join(manager->heartbeat_thread_id, NULL);
    
    // Disconnect if connected
    if (manager->is_connected) {
        project_manager_disconnect(manager);
    }
    
    // Save local data
    save_local_data(manager);
    
    // Cleanup SSL
    if (manager->config.enable_encryption) {
        cleanup_ssl(manager);
    }
    
    manager->is_running = 0;
    manager->status = PROJECT_STATUS_ARCHIVED;
    notify_status_change(manager);
    
    printf("Project manager stopped\n");
    return 1;
}

int project_manager_connect(ProjectManager *manager) {
    if (!manager || manager->is_connected) {
        return manager ? manager->is_connected : 0;
    }
    
    manager->status = PROJECT_STATUS_PLANNING; // Connecting status
    notify_status_change(manager);
    
    // Create socket
    manager->connection.socket_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (manager->connection.socket_fd < 0) {
        handle_error(manager, PROJECT_ERROR_NETWORK_FAILURE, "Failed to create socket");
        return 0;
    }
    
    // Set up server address
    memset(&manager->connection.server_addr, 0, sizeof(manager->connection.server_addr));
    manager->connection.server_addr.sin_family = AF_INET;
    manager->connection.server_addr.sin_port = htons(manager->config.server_port);
    
    if (inet_pton(AF_INET, manager->config.server_url, &manager->connection.server_addr.sin_addr) <= 0) {
        handle_error(manager, PROJECT_ERROR_NETWORK_FAILURE, "Invalid server address");
        close(manager->connection.socket_fd);
        manager->connection.socket_fd = -1;
        return 0;
    }
    
    // Connect to server
    if (connect(manager->connection.socket_fd, 
                (struct sockaddr*)&manager->connection.server_addr, 
                sizeof(manager->connection.server_addr)) < 0) {
        handle_error(manager, PROJECT_ERROR_NETWORK_FAILURE, "Failed to connect to server");
        close(manager->connection.socket_fd);
        manager->connection.socket_fd = -1;
        return 0;
    }
    
    // Set up SSL if encryption is enabled
    if (manager->config.enable_encryption) {
        manager->connection.ssl = SSL_new(manager->connection.ssl_ctx);
        if (!manager->connection.ssl) {
            handle_error(manager, PROJECT_ERROR_NETWORK_FAILURE, "Failed to create SSL connection");
            close(manager->connection.socket_fd);
            manager->connection.socket_fd = -1;
            return 0;
        }
        
        SSL_set_fd(manager->connection.ssl, manager->connection.socket_fd);
        
        if (SSL_connect(manager->connection.ssl) <= 0) {
            handle_error(manager, PROJECT_ERROR_NETWORK_FAILURE, "SSL handshake failed");
            SSL_free(manager->connection.ssl);
            manager->connection.ssl = NULL;
            close(manager->connection.socket_fd);
            manager->connection.socket_fd = -1;
            return 0;
        }
    }
    
    // Perform handshake and authentication
    if (!perform_handshake(manager) || !authenticate(manager)) {
        project_manager_disconnect(manager);
        return 0;
    }
    
    manager->connection.is_connected = 1;
    manager->is_connected = 1;
    manager->status = PROJECT_STATUS_ACTIVE;
    notify_status_change(manager);
    
    printf("Connected to project server\n");
    return 1;
}

int project_manager_disconnect(ProjectManager *manager) {
    if (!manager) return 0;
    
    if (manager->connection.ssl) {
        SSL_shutdown(manager->connection.ssl);
        SSL_free(manager->connection.ssl);
        manager->connection.ssl = NULL;
    }
    
    if (manager->connection.socket_fd >= 0) {
        close(manager->connection.socket_fd);
        manager->connection.socket_fd = -1;
    }
    
    manager->connection.is_connected = 0;
    manager->is_connected = 0;
    manager->session_id = 0;
    memset(manager->session_token, 0, sizeof(manager->session_token));
    
    manager->status = PROJECT_STATUS_ON_HOLD; // Offline status
    notify_status_change(manager);
    
    printf("Disconnected from project server\n");
    return 1;
}

// Project operations
int project_manager_create_project(ProjectManager *manager, const Project *project) {
    if (!manager || !project) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Check capacity
    if (manager->project_count >= manager->project_capacity) {
        manager->project_capacity *= 2;
        manager->projects = realloc(manager->projects, 
                                  manager->project_capacity * sizeof(Project*));
    }
    
    // Create copy
    Project *new_project = malloc(sizeof(Project));
    *new_project = *project;
    new_project->created_timestamp = get_current_timestamp();
    new_project->updated_timestamp = new_project->created_timestamp;
    new_project->last_activity_timestamp = new_project->created_timestamp;
    
    manager->projects[manager->project_count++] = new_project;
    
    // Store using storage interface if available
    if (manager->store_project) {
        manager->store_project(new_project);
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    // Notify callback
    if (manager->project_callback) {
        manager->project_callback(new_project, PROJECT_OPERATION_CREATE);
    }
    
    // Mark for sync if connected
    if (manager->is_connected && manager->config.auto_sync_enabled) {
        mark_project_dirty(manager, project->project_id);
    }
    
    return 1;
}

int project_manager_update_project(ProjectManager *manager, const Project *project) {
    if (!manager || !project) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Find existing project
    for (size_t i = 0; i < manager->project_count; i++) {
        if (strcmp(manager->projects[i]->project_id, project->project_id) == 0) {
            // Update project
            Project *existing = manager->projects[i];
            *existing = *project;
            existing->updated_timestamp = get_current_timestamp();
            
            // Store using storage interface if available
            if (manager->store_project) {
                manager->store_project(existing);
            }
            
            pthread_mutex_unlock(&manager->data_mutex);
            
            // Notify callback
            if (manager->project_callback) {
                manager->project_callback(existing, PROJECT_OPERATION_UPDATE);
            }
            
            return 1;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return 0;
}

int project_manager_delete_project(ProjectManager *manager, const char *project_id) {
    if (!manager || !project_id) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Find and remove project
    for (size_t i = 0; i < manager->project_count; i++) {
        if (strcmp(manager->projects[i]->project_id, project_id) == 0) {
            Project *project = manager->projects[i];
            
            // Notify callback before deletion
            if (manager->project_callback) {
                manager->project_callback(project, PROJECT_OPERATION_DELETE);
            }
            
            // Delete using storage interface if available
            if (manager->delete_project) {
                manager->delete_project(project_id);
            }
            
            // Remove from array
            free(project);
            for (size_t j = i; j < manager->project_count - 1; j++) {
                manager->projects[j] = manager->projects[j + 1];
            }
            manager->project_count--;
            
            pthread_mutex_unlock(&manager->data_mutex);
            return 1;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return 0;
}

const Project *project_manager_get_project(ProjectManager *manager, const char *project_id) {
    if (!manager || !project_id) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Search in local cache
    for (size_t i = 0; i < manager->project_count; i++) {
        if (strcmp(manager->projects[i]->project_id, project_id) == 0) {
            Project *project = manager->projects[i];
            pthread_mutex_unlock(&manager->data_mutex);
            return project;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    // Try storage interface
    if (manager->retrieve_project) {
        return manager->retrieve_project(project_id);
    }
    
    return NULL;
}

const Project **project_manager_list_projects(ProjectManager *manager, size_t *count) {
    if (!manager || !count) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Try storage interface first
    if (manager->list_projects) {
        pthread_mutex_unlock(&manager->data_mutex);
        return manager->list_projects(count);
    }
    
    // Use local cache
    *count = manager->project_count;
    const Project **result = malloc(manager->project_count * sizeof(Project*));
    for (size_t i = 0; i < manager->project_count; i++) {
        result[i] = manager->projects[i];
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return result;
}

// Issue operations
int project_manager_create_issue(ProjectManager *manager, const ProjectIssue *issue) {
    if (!manager || !issue) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Check capacity
    if (manager->issue_count >= manager->issue_capacity) {
        manager->issue_capacity *= 2;
        manager->issues = realloc(manager->issues, 
                                manager->issue_capacity * sizeof(ProjectIssue*));
    }
    
    // Create copy
    ProjectIssue *new_issue = malloc(sizeof(ProjectIssue));
    *new_issue = *issue;
    new_issue->created_timestamp = get_current_timestamp();
    new_issue->updated_timestamp = new_issue->created_timestamp;
    
    manager->issues[manager->issue_count++] = new_issue;
    
    // Update project statistics
    for (size_t i = 0; i < manager->project_count; i++) {
        if (strcmp(manager->projects[i]->project_id, issue->project_id) == 0) {
            manager->projects[i]->total_issues++;
            if (issue->status == ISSUE_STATUS_OPEN) {
                manager->projects[i]->open_issues++;
            }
            manager->projects[i]->last_activity_timestamp = get_current_timestamp();
            break;
        }
    }
    
    // Store using storage interface if available
    if (manager->store_issue) {
        manager->store_issue(new_issue);
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    // Notify callback
    if (manager->issue_callback) {
        manager->issue_callback(new_issue, PROJECT_OPERATION_CREATE);
    }
    
    return 1;
}

int project_manager_update_issue(ProjectManager *manager, const ProjectIssue *issue) {
    if (!manager || !issue) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Find existing issue
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->issue_id, issue->issue_id) == 0) {
            // Update issue
            IssueStatus old_status = manager->issues[i]->status;
            ProjectIssue *existing = manager->issues[i];
            *existing = *issue;
            existing->updated_timestamp = get_current_timestamp();
            
            // Update project statistics if status changed
            if (old_status != issue->status) {
                for (size_t j = 0; j < manager->project_count; j++) {
                    if (strcmp(manager->projects[j]->project_id, issue->project_id) == 0) {
                        if (old_status == ISSUE_STATUS_OPEN && issue->status != ISSUE_STATUS_OPEN) {
                            manager->projects[j]->open_issues--;
                            manager->projects[j]->closed_issues++;
                        } else if (old_status != ISSUE_STATUS_OPEN && issue->status == ISSUE_STATUS_OPEN) {
                            manager->projects[j]->open_issues++;
                            manager->projects[j]->closed_issues--;
                        }
                        manager->projects[j]->last_activity_timestamp = get_current_timestamp();
                        break;
                    }
                }
            }
            
            // Store using storage interface if available
            if (manager->store_issue) {
                manager->store_issue(existing);
            }
            
            pthread_mutex_unlock(&manager->data_mutex);
            
            // Notify callback
            if (manager->issue_callback) {
                manager->issue_callback(existing, PROJECT_OPERATION_UPDATE);
            }
            
            return 1;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return 0;
}

int project_manager_delete_issue(ProjectManager *manager, const char *issue_id) {
    if (!manager || !issue_id) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Find and remove issue
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->issue_id, issue_id) == 0) {
            ProjectIssue *issue = manager->issues[i];
            
            // Update project statistics
            for (size_t j = 0; j < manager->project_count; j++) {
                if (strcmp(manager->projects[j]->project_id, issue->project_id) == 0) {
                    manager->projects[j]->total_issues--;
                    manager->projects[j]->last_activity_timestamp = get_current_timestamp();
                    break;
                }
            }
            
            // Notify callback before deletion
            if (manager->issue_callback) {
                manager->issue_callback(issue, PROJECT_OPERATION_DELETE);
            }
            
            // Delete using storage interface if available
            if (manager->delete_issue) {
                manager->delete_issue(issue_id);
            }
            
            // Remove from array
            free(issue);
            for (size_t j = i; j < manager->issue_count - 1; j++) {
                manager->issues[j] = manager->issues[j + 1];
            }
            manager->issue_count--;
            
            pthread_mutex_unlock(&manager->data_mutex);
            return 1;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return 0;
}

const ProjectIssue *project_manager_get_issue(ProjectManager *manager, const char *issue_id) {
    if (!manager || !issue_id) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Search in local cache
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->issue_id, issue_id) == 0) {
            ProjectIssue *issue = manager->issues[i];
            pthread_mutex_unlock(&manager->data_mutex);
            return issue;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    // Try storage interface
    if (manager->retrieve_issue) {
        return manager->retrieve_issue(issue_id);
    }
    
    return NULL;
}

const ProjectIssue **project_manager_list_issues(ProjectManager *manager, const char *project_id, size_t *count) {
    if (!manager || !project_id || !count) return NULL;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Try storage interface first
    if (manager->list_issues) {
        pthread_mutex_unlock(&manager->data_mutex);
        return manager->list_issues(project_id, count);
    }
    
    // Count matching issues
    size_t matching_count = 0;
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->project_id, project_id) == 0) {
            matching_count++;
        }
    }
    
    if (matching_count == 0) {
        *count = 0;
        pthread_mutex_unlock(&manager->data_mutex);
        return NULL;
    }
    
    // Collect matching issues
    const ProjectIssue **result = malloc(matching_count * sizeof(ProjectIssue*));
    size_t result_index = 0;
    
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->project_id, project_id) == 0) {
            result[result_index++] = manager->issues[i];
        }
    }
    
    *count = matching_count;
    pthread_mutex_unlock(&manager->data_mutex);
    return result;
}

int project_manager_assign_issue(ProjectManager *manager, const char *issue_id, const char *assignee_id) {
    if (!manager || !issue_id || !assignee_id) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Find issue
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->issue_id, issue_id) == 0) {
            ProjectIssue *issue = manager->issues[i];
            
            // Check if already assigned
            for (int j = 0; j < issue->assignee_count; j++) {
                if (strcmp(issue->assignee_ids[j], assignee_id) == 0) {
                    pthread_mutex_unlock(&manager->data_mutex);
                    return 1; // Already assigned
                }
            }
            
            // Add assignee if there's space
            if (issue->assignee_count < PROJECT_MAX_ASSIGNEES) {
                strcpy(issue->assignee_ids[issue->assignee_count], assignee_id);
                issue->assignee_count++;
                issue->updated_timestamp = get_current_timestamp();
                
                // Store using storage interface if available
                if (manager->store_issue) {
                    manager->store_issue(issue);
                }
                
                pthread_mutex_unlock(&manager->data_mutex);
                
                // Notify callback
                if (manager->issue_callback) {
                    manager->issue_callback(issue, PROJECT_OPERATION_UPDATE);
                }
                
                return 1;
            }
            
            break;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return 0;
}

int project_manager_update_issue_status(ProjectManager *manager, const char *issue_id, IssueStatus status) {
    if (!manager || !issue_id) return 0;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    // Find issue
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->issue_id, issue_id) == 0) {
            ProjectIssue *issue = manager->issues[i];
            IssueStatus old_status = issue->status;
            
            issue->status = status;
            issue->updated_timestamp = get_current_timestamp();
            
            if (status == ISSUE_STATUS_RESOLVED || status == ISSUE_STATUS_CLOSED) {
                issue->resolved_timestamp = get_current_timestamp();
            }
            
            // Update project statistics
            for (size_t j = 0; j < manager->project_count; j++) {
                if (strcmp(manager->projects[j]->project_id, issue->project_id) == 0) {
                    if (old_status == ISSUE_STATUS_OPEN && status != ISSUE_STATUS_OPEN) {
                        manager->projects[j]->open_issues--;
                        manager->projects[j]->closed_issues++;
                    } else if (old_status != ISSUE_STATUS_OPEN && status == ISSUE_STATUS_OPEN) {
                        manager->projects[j]->open_issues++;
                        manager->projects[j]->closed_issues--;
                    }
                    manager->projects[j]->last_activity_timestamp = get_current_timestamp();
                    break;
                }
            }
            
            // Store using storage interface if available
            if (manager->store_issue) {
                manager->store_issue(issue);
            }
            
            pthread_mutex_unlock(&manager->data_mutex);
            
            // Notify callback
            if (manager->issue_callback) {
                manager->issue_callback(issue, PROJECT_OPERATION_UPDATE);
            }
            
            return 1;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    return 0;
}

// Synchronization
int project_manager_sync_all(ProjectManager *manager) {
    if (!manager) return 0;
    
    if (!manager->is_connected) {
        if (!project_manager_connect(manager)) {
            return 0;
        }
    }
    
    manager->status = PROJECT_STATUS_ACTIVE; // Syncing status
    notify_status_change(manager);
    
    pthread_mutex_lock(&manager->data_mutex);
    
    int success = 1;
    manager->synced_items = 0;
    manager->failed_items = 0;
    
    // Sync projects
    for (size_t i = 0; i < manager->project_count; i++) {
        // Implementation would send sync messages for each project
        manager->synced_items++;
    }
    
    // Sync issues
    for (size_t i = 0; i < manager->issue_count; i++) {
        // Implementation would send sync messages for each issue
        manager->synced_items++;
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    manager->status = success ? PROJECT_STATUS_COMPLETED : PROJECT_STATUS_ON_HOLD;
    notify_status_change(manager);
    
    if (manager->sync_complete_callback) {
        uint32_t total_projects = (uint32_t)manager->project_count;
        uint32_t total_issues = (uint32_t)manager->issue_count;
        manager->sync_complete_callback(total_projects, total_issues, manager->failed_items);
    }
    
    return success;
}

int project_manager_sync_project(ProjectManager *manager, const char *project_id) {
    if (!manager || !project_id) return 0;
    
    // Implementation would sync specific project
    return 1;
}

// Status and monitoring
ProjectStatus project_manager_get_status(ProjectManager *manager) {
    return manager ? manager->status : PROJECT_STATUS_PLANNING;
}

float project_manager_get_sync_progress(ProjectManager *manager) {
    return manager ? manager->sync_progress : 0.0f;
}

void project_manager_get_stats(ProjectManager *manager, uint32_t *total_projects, uint32_t *total_issues, uint32_t *pending_sync) {
    if (!manager) return;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    if (total_projects) *total_projects = (uint32_t)manager->project_count;
    if (total_issues) *total_issues = (uint32_t)manager->issue_count;
    if (pending_sync) *pending_sync = manager->pending_sync_items;
    
    pthread_mutex_unlock(&manager->data_mutex);
}

float project_manager_calculate_progress(ProjectManager *manager, const char *project_id) {
    if (!manager || !project_id) return 0.0f;
    
    pthread_mutex_lock(&manager->data_mutex);
    
    float total_progress = 0.0f;
    int issue_count = 0;
    
    for (size_t i = 0; i < manager->issue_count; i++) {
        if (strcmp(manager->issues[i]->project_id, project_id) == 0) {
            total_progress += manager->issues[i]->progress_percentage;
            issue_count++;
        }
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    return issue_count > 0 ? total_progress / issue_count : 0.0f;
}

// Callback setters
void project_manager_set_status_callback(ProjectManager *manager, ProjectStatusCallback callback) {
    if (manager) manager->status_callback = callback;
}

void project_manager_set_project_callback(ProjectManager *manager, ProjectDataCallback callback) {
    if (manager) manager->project_callback = callback;
}

void project_manager_set_issue_callback(ProjectManager *manager, IssueDataCallback callback) {
    if (manager) manager->issue_callback = callback;
}

void project_manager_set_notification_callback(ProjectManager *manager, NotificationCallback callback) {
    if (manager) manager->notification_callback = callback;
}

void project_manager_set_error_callback(ProjectManager *manager, ProjectErrorCallback callback) {
    if (manager) manager->error_callback = callback;
}

void project_manager_set_sync_complete_callback(ProjectManager *manager, SyncCompleteCallback callback) {
    if (manager) manager->sync_complete_callback = callback;
}

// Storage interface setters
void project_manager_set_store_project(ProjectManager *manager, StoreProjectCallback callback) {
    if (manager) manager->store_project = callback;
}

void project_manager_set_retrieve_project(ProjectManager *manager, RetrieveProjectCallback callback) {
    if (manager) manager->retrieve_project = callback;
}

void project_manager_set_delete_project(ProjectManager *manager, DeleteProjectCallback callback) {
    if (manager) manager->delete_project = callback;
}

void project_manager_set_list_projects(ProjectManager *manager, ListProjectsCallback callback) {
    if (manager) manager->list_projects = callback;
}

void project_manager_set_store_issue(ProjectManager *manager, StoreIssueCallback callback) {
    if (manager) manager->store_issue = callback;
}

void project_manager_set_retrieve_issue(ProjectManager *manager, RetrieveIssueCallback callback) {
    if (manager) manager->retrieve_issue = callback;
}

void project_manager_set_delete_issue(ProjectManager *manager, DeleteIssueCallback callback) {
    if (manager) manager->delete_issue = callback;
}

void project_manager_set_list_issues(ProjectManager *manager, ListIssuesCallback callback) {
    if (manager) manager->list_issues = callback;
}

void project_manager_set_store_comment(ProjectManager *manager, StoreCommentCallback callback) {
    if (manager) manager->store_comment = callback;
}

void project_manager_set_retrieve_comments(ProjectManager *manager, RetrieveCommentsCallback callback) {
    if (manager) manager->retrieve_comments = callback;
}

void project_manager_set_delete_comment(ProjectManager *manager, DeleteCommentCallback callback) {
    if (manager) manager->delete_comment = callback;
}

void project_manager_set_store_attachment(ProjectManager *manager, StoreAttachmentCallback callback) {
    if (manager) manager->store_attachment = callback;
}

void project_manager_set_retrieve_attachment(ProjectManager *manager, RetrieveAttachmentCallback callback) {
    if (manager) manager->retrieve_attachment = callback;
}

void project_manager_set_delete_attachment(ProjectManager *manager, DeleteAttachmentCallback callback) {
    if (manager) manager->delete_attachment = callback;
}

// Utility functions
char *generate_project_id(void) {
    static char id[17];
    const char chars[] = "0123456789abcdef";
    
    for (int i = 0; i < 16; i++) {
        id[i] = chars[rand() % 16];
    }
    id[16] = '\0';
    
    return id;
}

uint64_t get_current_timestamp(void) {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (uint64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

uint32_t calculate_checksum(const void *data, size_t size) {
    uint32_t checksum = 0;
    const uint8_t *bytes = (const uint8_t*)data;
    
    for (size_t i = 0; i < size; i++) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
}

int validate_project_data(const Project *project) {
    if (!project) return 0;
    
    return strlen(project->project_id) > 0 && 
           strlen(project->name) > 0 && 
           strlen(project->owner_id) > 0;
}

int validate_issue_data(const ProjectIssue *issue) {
    if (!issue) return 0;
    
    return strlen(issue->issue_id) > 0 && 
           strlen(issue->project_id) > 0 && 
           strlen(issue->title) > 0 && 
           strlen(issue->reporter_id) > 0;
}

const char *project_error_to_string(ProjectError error) {
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

const char *project_status_to_string(ProjectStatus status) {
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

const char *issue_status_to_string(IssueStatus status) {
    switch (status) {
        case ISSUE_STATUS_OPEN: return "Open";
        case ISSUE_STATUS_IN_PROGRESS: return "In Progress";
        case ISSUE_STATUS_RESOLVED: return "Resolved";
        case ISSUE_STATUS_CLOSED: return "Closed";
        case ISSUE_STATUS_REOPENED: return "Reopened";
        default: return "Unknown";
    }
}

// Private function implementations

static void *sync_thread(void *arg) {
    ProjectManager *manager = (ProjectManager*)arg;
    
    while (manager->sync_thread_running && !g_shutdown_requested) {
        pthread_mutex_lock(&manager->data_mutex);
        
        // Wait for sync interval or signal
        struct timespec timeout;
        clock_gettime(CLOCK_REALTIME, &timeout);
        timeout.tv_sec += manager->config.sync_interval / 1000;
        
        int result = pthread_cond_timedwait(&manager->sync_cond, &manager->data_mutex, &timeout);
        
        pthread_mutex_unlock(&manager->data_mutex);
        
        if (result == ETIMEDOUT && manager->config.auto_sync_enabled && manager->is_connected) {
            project_manager_sync_all(manager);
        }
    }
    
    return NULL;
}

static void *heartbeat_thread(void *arg) {
    ProjectManager *manager = (ProjectManager*)arg;
    
    while (manager->heartbeat_thread_running && !g_shutdown_requested) {
        if (manager->is_connected) {
            send_heartbeat(manager);
        }
        
        usleep(manager->config.heartbeat_interval * 1000);
    }
    
    return NULL;
}

static int init_ssl(ProjectManager *manager) {
    SSL_library_init();
    SSL_load_error_strings();
    OpenSSL_add_all_algorithms();
    
    manager->connection.ssl_ctx = SSL_CTX_new(TLS_client_method());
    if (!manager->connection.ssl_ctx) {
        return 0;
    }
    
    return 1;
}

static void cleanup_ssl(ProjectManager *manager) {
    if (manager->connection.ssl_ctx) {
        SSL_CTX_free(manager->connection.ssl_ctx);
        manager->connection.ssl_ctx = NULL;
    }
    
    EVP_cleanup();
}

static int perform_handshake(ProjectManager *manager) {
    // Create handshake message
    json_object *handshake = json_object_new_object();
    json_object *device_id = json_object_new_string(manager->config.device_id);
    json_object *device_name = json_object_new_string("Linux Desktop");
    json_object *protocol_version = json_object_new_int(1);
    
    json_object_object_add(handshake, "device_id", device_id);
    json_object_object_add(handshake, "device_name", device_name);
    json_object_object_add(handshake, "protocol_version", protocol_version);
    
    const char *handshake_str = json_object_to_json_string(handshake);
    
    // Send handshake
    int result = send_message(manager, PROJECT_MESSAGE_HANDSHAKE, 
                             handshake_str, strlen(handshake_str));
    
    json_object_put(handshake);
    
    if (!result) {
        return 0;
    }
    
    // Receive handshake response
    ProjectMessageType response_type;
    void *response_data;
    size_t response_size;
    
    if (!receive_message(manager, &response_type, &response_data, &response_size)) {
        return 0;
    }
    
    if (response_type != PROJECT_MESSAGE_HANDSHAKE) {
        free(response_data);
        return 0;
    }
    
    // Parse response
    json_object *response = json_tokener_parse((char*)response_data);
    json_object *accepted;
    
    int handshake_accepted = 0;
    if (json_object_object_get_ex(response, "handshake_accepted", &accepted)) {
        handshake_accepted = json_object_get_boolean(accepted);
    }
    
    json_object_put(response);
    free(response_data);
    
    return handshake_accepted;
}

static int authenticate(ProjectManager *manager) {
    // Create auth message
    json_object *auth = json_object_new_object();
    json_object *user_id = json_object_new_string(manager->config.user_id);
    json_object *auth_token = json_object_new_string(manager->config.auth_token);
    json_object *timestamp = json_object_new_int64(get_current_timestamp());
    
    char device_signature[256];
    generate_device_signature(device_signature, sizeof(device_signature));
    json_object *signature = json_object_new_string(device_signature);
    
    json_object_object_add(auth, "user_id", user_id);
    json_object_object_add(auth, "auth_token", auth_token);
    json_object_object_add(auth, "device_signature", signature);
    json_object_object_add(auth, "timestamp", timestamp);
    
    const char *auth_str = json_object_to_json_string(auth);
    
    // Send auth
    int result = send_message(manager, PROJECT_MESSAGE_AUTH, 
                             auth_str, strlen(auth_str));
    
    json_object_put(auth);
    
    if (!result) {
        return 0;
    }
    
    // Receive auth response
    ProjectMessageType response_type;
    void *response_data;
    size_t response_size;
    
    if (!receive_message(manager, &response_type, &response_data, &response_size)) {
        return 0;
    }
    
    if (response_type != PROJECT_MESSAGE_AUTH) {
        free(response_data);
        return 0;
    }
    
    // Parse response
    json_object *response = json_tokener_parse((char*)response_data);
    json_object *success, *session_token;
    
    int auth_success = 0;
    if (json_object_object_get_ex(response, "auth_success", &success)) {
        auth_success = json_object_get_boolean(success);
    }
    
    if (auth_success && json_object_object_get_ex(response, "session_token", &session_token)) {
        strcpy(manager->session_token, json_object_get_string(session_token));
        manager->session_id = generate_message_id();
    }
    
    json_object_put(response);
    free(response_data);
    
    return auth_success;
}

static int send_message(ProjectManager *manager, ProjectMessageType type, const void *data, size_t data_size) {
    if (!manager || !data || data_size == 0) return 0;
    
    // Create header
    ProjectHeader header;
    header.magic = PROJECT_MAGIC_NUMBER;
    header.version = PROJECT_PROTOCOL_VERSION;
    header.message_type = type;
    header.message_id = generate_message_id();
    header.session_id = manager->session_id;
    header.data_length = (uint32_t)data_size;
    header.checksum = calculate_checksum(data, data_size);
    header.timestamp = get_current_timestamp();
    memset(header.reserved, 0, sizeof(header.reserved));
    
    // Send header
    ssize_t sent;
    if (manager->connection.ssl) {
        sent = SSL_write(manager->connection.ssl, &header, sizeof(header));
    } else {
        sent = send(manager->connection.socket_fd, &header, sizeof(header), 0);
    }
    
    if (sent != sizeof(header)) {
        return 0;
    }
    
    // Send data
    if (manager->connection.ssl) {
        sent = SSL_write(manager->connection.ssl, data, data_size);
    } else {
        sent = send(manager->connection.socket_fd, data, data_size, 0);
    }
    
    return sent == (ssize_t)data_size;
}

static int receive_message(ProjectManager *manager, ProjectMessageType *type, void **data, size_t *data_size) {
    if (!manager || !type || !data || !data_size) return 0;
    
    // Receive header
    ProjectHeader header;
    ssize_t received;
    
    if (manager->connection.ssl) {
        received = SSL_read(manager->connection.ssl, &header, sizeof(header));
    } else {
        received = recv(manager->connection.socket_fd, &header, sizeof(header), 0);
    }
    
    if (received != sizeof(header) || header.magic != PROJECT_MAGIC_NUMBER) {
        return 0;
    }
    
    *type = header.message_type;
    *data_size = header.data_length;
    
    if (*data_size == 0) {
        *data = NULL;
        return 1;
    }
    
    // Receive data
    *data = malloc(*data_size);
    if (!*data) {
        return 0;
    }
    
    if (manager->connection.ssl) {
        received = SSL_read(manager->connection.ssl, *data, *data_size);
    } else {
        received = recv(manager->connection.socket_fd, *data, *data_size, 0);
    }
    
    if (received != (ssize_t)*data_size) {
        free(*data);
        *data = NULL;
        return 0;
    }
    
    // Verify checksum
    uint32_t calculated_checksum = calculate_checksum(*data, *data_size);
    if (calculated_checksum != header.checksum) {
        free(*data);
        *data = NULL;
        return 0;
    }
    
    return 1;
}

static void send_heartbeat(ProjectManager *manager) {
    if (!manager || !manager->is_connected) return;
    
    json_object *heartbeat = json_object_new_object();
    json_object *timestamp = json_object_new_int64(get_current_timestamp());
    json_object_object_add(heartbeat, "timestamp", timestamp);
    
    const char *heartbeat_str = json_object_to_json_string(heartbeat);
    send_message(manager, PROJECT_MESSAGE_HEARTBEAT, heartbeat_str, strlen(heartbeat_str));
    
    json_object_put(heartbeat);
}

static void load_local_data(ProjectManager *manager) {
    char expanded_path[512];
    if (manager->config.local_storage_path[0] == '~') {
        snprintf(expanded_path, sizeof(expanded_path), "%s%s", 
                getenv("HOME"), manager->config.local_storage_path + 1);
    } else {
        strcpy(expanded_path, manager->config.local_storage_path);
    }
    
    char projects_file[600];
    snprintf(projects_file, sizeof(projects_file), "%s/projects.json", expanded_path);
    
    FILE *file = fopen(projects_file, "r");
    if (!file) return;
    
    fseek(file, 0, SEEK_END);
    long file_size = ftell(file);
    fseek(file, 0, SEEK_SET);
    
    char *json_str = malloc(file_size + 1);
    fread(json_str, 1, file_size, file);
    json_str[file_size] = '\0';
    fclose(file);
    
    json_object *projects_json = json_tokener_parse(json_str);
    if (projects_json) {
        // Parse projects from JSON
        // Implementation would deserialize projects
        json_object_put(projects_json);
    }
    
    free(json_str);
}

static void save_local_data(ProjectManager *manager) {
    char expanded_path[512];
    if (manager->config.local_storage_path[0] == '~') {
        snprintf(expanded_path, sizeof(expanded_path), "%s%s", 
                getenv("HOME"), manager->config.local_storage_path + 1);
    } else {
        strcpy(expanded_path, manager->config.local_storage_path);
    }
    
    char projects_file[600];
    snprintf(projects_file, sizeof(projects_file), "%s/projects.json", expanded_path);
    
    json_object *projects_json = json_object_new_array();
    
    pthread_mutex_lock(&manager->data_mutex);
    
    for (size_t i = 0; i < manager->project_count; i++) {
        json_object *project_json = project_to_json(manager->projects[i]);
        json_object_array_add(projects_json, project_json);
    }
    
    pthread_mutex_unlock(&manager->data_mutex);
    
    const char *json_str = json_object_to_json_string_ext(projects_json, JSON_C_TO_STRING_PRETTY);
    
    FILE *file = fopen(projects_file, "w");
    if (file) {
        fprintf(file, "%s", json_str);
        fclose(file);
    }
    
    json_object_put(projects_json);
}

static void mark_project_dirty(ProjectManager *manager, const char *project_id) {
    // Mark project for sync
    manager->pending_sync_items++;
}

static void notify_status_change(ProjectManager *manager) {
    if (manager && manager->status_callback) {
        manager->status_callback(manager->status, manager->sync_progress);
    }
}

static void handle_error(ProjectManager *manager, ProjectError error, const char *message) {
    if (!manager) return;
    
    manager->status = PROJECT_STATUS_ON_HOLD; // Error status
    
    if (manager->error_callback) {
        manager->error_callback(error, message);
    }
    
    printf("Project error: %s\n", message);
}

static uint32_t generate_message_id(void) {
    return (uint32_t)rand();
}

static void generate_device_signature(char *signature, size_t size) {
    snprintf(signature, size, "%s_%lu", 
             g_manager ? g_manager->config.device_id : "unknown",
             (unsigned long)get_current_timestamp());
}

static json_object *project_to_json(const Project *project) {
    if (!project) return NULL;
    
    json_object *json = json_object_new_object();
    
    json_object_object_add(json, "project_id", json_object_new_string(project->project_id));
    json_object_object_add(json, "name", json_object_new_string(project->name));
    json_object_object_add(json, "description", json_object_new_string(project->description));
    json_object_object_add(json, "owner_id", json_object_new_string(project->owner_id));
    json_object_object_add(json, "status", json_object_new_int(project->status));
    json_object_object_add(json, "priority", json_object_new_int(project->priority));
    json_object_object_add(json, "created_timestamp", json_object_new_int64(project->created_timestamp));
    json_object_object_add(json, "updated_timestamp", json_object_new_int64(project->updated_timestamp));
    
    return json;
}

static Project *json_to_project(json_object *json) {
    if (!json) return NULL;
    
    Project *project = malloc(sizeof(Project));
    memset(project, 0, sizeof(Project));
    
    json_object *field;
    
    if (json_object_object_get_ex(json, "project_id", &field)) {
        strcpy(project->project_id, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "name", &field)) {
        strcpy(project->name, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "description", &field)) {
        strcpy(project->description, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "owner_id", &field)) {
        strcpy(project->owner_id, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "status", &field)) {
        project->status = json_object_get_int(field);
    }
    if (json_object_object_get_ex(json, "priority", &field)) {
        project->priority = json_object_get_int(field);
    }
    if (json_object_object_get_ex(json, "created_timestamp", &field)) {
        project->created_timestamp = json_object_get_int64(field);
    }
    if (json_object_object_get_ex(json, "updated_timestamp", &field)) {
        project->updated_timestamp = json_object_get_int64(field);
    }
    
    return project;
}

static json_object *issue_to_json(const ProjectIssue *issue) {
    if (!issue) return NULL;
    
    json_object *json = json_object_new_object();
    
    json_object_object_add(json, "issue_id", json_object_new_string(issue->issue_id));
    json_object_object_add(json, "project_id", json_object_new_string(issue->project_id));
    json_object_object_add(json, "title", json_object_new_string(issue->title));
    json_object_object_add(json, "description", json_object_new_string(issue->description));
    json_object_object_add(json, "reporter_id", json_object_new_string(issue->reporter_id));
    json_object_object_add(json, "type", json_object_new_int(issue->type));
    json_object_object_add(json, "status", json_object_new_int(issue->status));
    json_object_object_add(json, "priority", json_object_new_int(issue->priority));
    json_object_object_add(json, "progress_percentage", json_object_new_double(issue->progress_percentage));
    json_object_object_add(json, "created_timestamp", json_object_new_int64(issue->created_timestamp));
    json_object_object_add(json, "updated_timestamp", json_object_new_int64(issue->updated_timestamp));
    
    // Add assignees array
    json_object *assignees = json_object_new_array();
    for (int i = 0; i < issue->assignee_count; i++) {
        json_object_array_add(assignees, json_object_new_string(issue->assignee_ids[i]));
    }
    json_object_object_add(json, "assignees", assignees);
    
    return json;
}

static ProjectIssue *json_to_issue(json_object *json) {
    if (!json) return NULL;
    
    ProjectIssue *issue = malloc(sizeof(ProjectIssue));
    memset(issue, 0, sizeof(ProjectIssue));
    
    json_object *field;
    
    if (json_object_object_get_ex(json, "issue_id", &field)) {
        strcpy(issue->issue_id, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "project_id", &field)) {
        strcpy(issue->project_id, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "title", &field)) {
        strcpy(issue->title, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "description", &field)) {
        strcpy(issue->description, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "reporter_id", &field)) {
        strcpy(issue->reporter_id, json_object_get_string(field));
    }
    if (json_object_object_get_ex(json, "type", &field)) {
        issue->type = json_object_get_int(field);
    }
    if (json_object_object_get_ex(json, "status", &field)) {
        issue->status = json_object_get_int(field);
    }
    if (json_object_object_get_ex(json, "priority", &field)) {
        issue->priority = json_object_get_int(field);
    }
    if (json_object_object_get_ex(json, "progress_percentage", &field)) {
        issue->progress_percentage = json_object_get_double(field);
    }
    if (json_object_object_get_ex(json, "created_timestamp", &field)) {
        issue->created_timestamp = json_object_get_int64(field);
    }
    if (json_object_object_get_ex(json, "updated_timestamp", &field)) {
        issue->updated_timestamp = json_object_get_int64(field);
    }
    
    // Parse assignees array
    if (json_object_object_get_ex(json, "assignees", &field)) {
        int array_len = json_object_array_length(field);
        issue->assignee_count = array_len > PROJECT_MAX_ASSIGNEES ? PROJECT_MAX_ASSIGNEES : array_len;
        
        for (int i = 0; i < issue->assignee_count; i++) {
            json_object *assignee = json_object_array_get_idx(field, i);
            strcpy(issue->assignee_ids[i], json_object_get_string(assignee));
        }
    }
    
    return issue;
}