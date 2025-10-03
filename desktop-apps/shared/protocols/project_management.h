#ifndef PROJECT_MANAGEMENT_H
#define PROJECT_MANAGEMENT_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// MARK: - Constants

#define PROJECT_MANAGEMENT_PROTOCOL_VERSION 1
#define PROJECT_MANAGEMENT_MAGIC 0x504D4754  // "PMGT"

#define MAX_PROJECT_ID_LENGTH 64
#define MAX_PROJECT_NAME_LENGTH 256
#define MAX_PROJECT_DESCRIPTION_LENGTH 2048
#define MAX_ISSUE_ID_LENGTH 64
#define MAX_ISSUE_TITLE_LENGTH 512
#define MAX_ISSUE_DESCRIPTION_LENGTH 4096
#define MAX_COMMENT_LENGTH 2048
#define MAX_TAG_LENGTH 64
#define MAX_USER_ID_LENGTH 64
#define MAX_USER_NAME_LENGTH 128
#define MAX_FILE_PATH_LENGTH 512
#define MAX_MILESTONE_NAME_LENGTH 256
#define MAX_LABEL_NAME_LENGTH 64
#define MAX_ATTACHMENT_NAME_LENGTH 256

#define MAX_TAGS_PER_ISSUE 10
#define MAX_ASSIGNEES_PER_ISSUE 5
#define MAX_LABELS_PER_ISSUE 10
#define MAX_ATTACHMENTS_PER_ISSUE 20
#define MAX_COMMENTS_PER_ISSUE 1000
#define MAX_ISSUES_PER_PROJECT 10000
#define MAX_MILESTONES_PER_PROJECT 50
#define MAX_MEMBERS_PER_PROJECT 100

#define DEFAULT_PROJECT_PORT 8444
#define CONNECTION_TIMEOUT_MS 15000
#define HEARTBEAT_INTERVAL_MS 30000
#define SYNC_INTERVAL_MS 60000

// MARK: - Enums

typedef enum {
    PROJECT_STATUS_PLANNING = 0,
    PROJECT_STATUS_ACTIVE = 1,
    PROJECT_STATUS_ON_HOLD = 2,
    PROJECT_STATUS_COMPLETED = 3,
    PROJECT_STATUS_CANCELLED = 4,
    PROJECT_STATUS_ARCHIVED = 5
} ProjectStatus;

typedef enum {
    PROJECT_PRIORITY_LOW = 0,
    PROJECT_PRIORITY_MEDIUM = 1,
    PROJECT_PRIORITY_HIGH = 2,
    PROJECT_PRIORITY_CRITICAL = 3
} ProjectPriority;

typedef enum {
    ISSUE_TYPE_BUG = 0,
    ISSUE_TYPE_FEATURE = 1,
    ISSUE_TYPE_TASK = 2,
    ISSUE_TYPE_IMPROVEMENT = 3,
    ISSUE_TYPE_EPIC = 4,
    ISSUE_TYPE_STORY = 5,
    ISSUE_TYPE_SUBTASK = 6
} IssueType;

typedef enum {
    ISSUE_STATUS_OPEN = 0,
    ISSUE_STATUS_IN_PROGRESS = 1,
    ISSUE_STATUS_IN_REVIEW = 2,
    ISSUE_STATUS_TESTING = 3,
    ISSUE_STATUS_RESOLVED = 4,
    ISSUE_STATUS_CLOSED = 5,
    ISSUE_STATUS_REOPENED = 6,
    ISSUE_STATUS_BLOCKED = 7
} IssueStatus;

typedef enum {
    ISSUE_PRIORITY_LOWEST = 0,
    ISSUE_PRIORITY_LOW = 1,
    ISSUE_PRIORITY_MEDIUM = 2,
    ISSUE_PRIORITY_HIGH = 3,
    ISSUE_PRIORITY_HIGHEST = 4,
    ISSUE_PRIORITY_BLOCKER = 5
} IssuePriority;

typedef enum {
    MILESTONE_STATUS_OPEN = 0,
    MILESTONE_STATUS_CLOSED = 1
} MilestoneStatus;

typedef enum {
    PROJECT_ROLE_OWNER = 0,
    PROJECT_ROLE_ADMIN = 1,
    PROJECT_ROLE_DEVELOPER = 2,
    PROJECT_ROLE_TESTER = 3,
    PROJECT_ROLE_VIEWER = 4,
    PROJECT_ROLE_GUEST = 5
} ProjectRole;

typedef enum {
    PROJECT_OPERATION_CREATE = 0,
    PROJECT_OPERATION_UPDATE = 1,
    PROJECT_OPERATION_DELETE = 2,
    PROJECT_OPERATION_ARCHIVE = 3,
    PROJECT_OPERATION_RESTORE = 4
} ProjectOperation;

typedef enum {
    PROJECT_ERROR_NONE = 0,
    PROJECT_ERROR_NETWORK_FAILURE = 1,
    PROJECT_ERROR_AUTH_FAILED = 2,
    PROJECT_ERROR_PERMISSION_DENIED = 3,
    PROJECT_ERROR_PROJECT_NOT_FOUND = 4,
    PROJECT_ERROR_ISSUE_NOT_FOUND = 5,
    PROJECT_ERROR_INVALID_DATA = 6,
    PROJECT_ERROR_STORAGE_FULL = 7,
    PROJECT_ERROR_PROTOCOL_ERROR = 8,
    PROJECT_ERROR_VERSION_MISMATCH = 9,
    PROJECT_ERROR_TIMEOUT = 10,
    PROJECT_ERROR_CONFLICT = 11,
    PROJECT_ERROR_QUOTA_EXCEEDED = 12
} ProjectError;

typedef enum {
    PROJECT_MSG_TYPE_HANDSHAKE = 0,
    PROJECT_MSG_TYPE_AUTH = 1,
    PROJECT_MSG_TYPE_PROJECT_LIST = 2,
    PROJECT_MSG_TYPE_PROJECT_CREATE = 3,
    PROJECT_MSG_TYPE_PROJECT_UPDATE = 4,
    PROJECT_MSG_TYPE_PROJECT_DELETE = 5,
    PROJECT_MSG_TYPE_ISSUE_LIST = 6,
    PROJECT_MSG_TYPE_ISSUE_CREATE = 7,
    PROJECT_MSG_TYPE_ISSUE_UPDATE = 8,
    PROJECT_MSG_TYPE_ISSUE_DELETE = 9,
    PROJECT_MSG_TYPE_COMMENT_ADD = 10,
    PROJECT_MSG_TYPE_COMMENT_UPDATE = 11,
    PROJECT_MSG_TYPE_COMMENT_DELETE = 12,
    PROJECT_MSG_TYPE_MILESTONE_CREATE = 13,
    PROJECT_MSG_TYPE_MILESTONE_UPDATE = 14,
    PROJECT_MSG_TYPE_MILESTONE_DELETE = 15,
    PROJECT_MSG_TYPE_MEMBER_ADD = 16,
    PROJECT_MSG_TYPE_MEMBER_REMOVE = 17,
    PROJECT_MSG_TYPE_MEMBER_UPDATE = 18,
    PROJECT_MSG_TYPE_ATTACHMENT_UPLOAD = 19,
    PROJECT_MSG_TYPE_ATTACHMENT_DELETE = 20,
    PROJECT_MSG_TYPE_SYNC_REQUEST = 21,
    PROJECT_MSG_TYPE_SYNC_RESPONSE = 22,
    PROJECT_MSG_TYPE_NOTIFICATION = 23,
    PROJECT_MSG_TYPE_HEARTBEAT = 24,
    PROJECT_MSG_TYPE_ERROR = 25,
    PROJECT_MSG_TYPE_ACK = 26
} ProjectMessageType;

typedef enum {
    NOTIFICATION_TYPE_ISSUE_CREATED = 0,
    NOTIFICATION_TYPE_ISSUE_UPDATED = 1,
    NOTIFICATION_TYPE_ISSUE_ASSIGNED = 2,
    NOTIFICATION_TYPE_ISSUE_COMMENTED = 3,
    NOTIFICATION_TYPE_ISSUE_STATUS_CHANGED = 4,
    NOTIFICATION_TYPE_PROJECT_UPDATED = 5,
    NOTIFICATION_TYPE_MILESTONE_REACHED = 6,
    NOTIFICATION_TYPE_DEADLINE_APPROACHING = 7,
    NOTIFICATION_TYPE_MEMBER_ADDED = 8,
    NOTIFICATION_TYPE_MEMBER_REMOVED = 9
} NotificationType;

// MARK: - Core Data Structures

typedef struct {
    uint32_t magic;
    uint16_t version;
    uint16_t message_type;
    uint32_t message_id;
    uint32_t session_id;
    uint32_t data_length;
    uint32_t checksum;
    uint64_t timestamp;
    char reserved[8];
} ProjectHeader;

typedef struct {
    char user_id[MAX_USER_ID_LENGTH];
    char name[MAX_USER_NAME_LENGTH];
    char email[MAX_USER_NAME_LENGTH];
    char avatar_url[MAX_FILE_PATH_LENGTH];
    ProjectRole role;
    uint64_t joined_timestamp;
    bool is_active;
} ProjectMember;

typedef struct {
    char milestone_id[MAX_PROJECT_ID_LENGTH];
    char name[MAX_MILESTONE_NAME_LENGTH];
    char description[MAX_PROJECT_DESCRIPTION_LENGTH];
    uint64_t due_date;
    uint64_t created_timestamp;
    uint64_t updated_timestamp;
    MilestoneStatus status;
    uint32_t total_issues;
    uint32_t completed_issues;
    float progress_percentage;
} ProjectMilestone;

typedef struct {
    char label_id[MAX_PROJECT_ID_LENGTH];
    char name[MAX_LABEL_NAME_LENGTH];
    char color[16]; // Hex color code
    char description[MAX_PROJECT_DESCRIPTION_LENGTH];
} ProjectLabel;

typedef struct {
    char attachment_id[MAX_PROJECT_ID_LENGTH];
    char filename[MAX_ATTACHMENT_NAME_LENGTH];
    char file_path[MAX_FILE_PATH_LENGTH];
    char mime_type[64];
    uint64_t file_size;
    uint64_t uploaded_timestamp;
    char uploaded_by[MAX_USER_ID_LENGTH];
    uint32_t download_count;
} IssueAttachment;

typedef struct {
    char comment_id[MAX_PROJECT_ID_LENGTH];
    char issue_id[MAX_ISSUE_ID_LENGTH];
    char author_id[MAX_USER_ID_LENGTH];
    char content[MAX_COMMENT_LENGTH];
    uint64_t created_timestamp;
    uint64_t updated_timestamp;
    bool is_edited;
    char parent_comment_id[MAX_PROJECT_ID_LENGTH]; // For threaded comments
} IssueComment;

typedef struct {
    char issue_id[MAX_ISSUE_ID_LENGTH];
    char project_id[MAX_PROJECT_ID_LENGTH];
    char title[MAX_ISSUE_TITLE_LENGTH];
    char description[MAX_ISSUE_DESCRIPTION_LENGTH];
    IssueType type;
    IssueStatus status;
    IssuePriority priority;
    
    // Relationships
    char reporter_id[MAX_USER_ID_LENGTH];
    char assignee_ids[MAX_ASSIGNEES_PER_ISSUE][MAX_USER_ID_LENGTH];
    uint32_t assignee_count;
    char milestone_id[MAX_PROJECT_ID_LENGTH];
    char parent_issue_id[MAX_ISSUE_ID_LENGTH]; // For subtasks
    
    // Labels and tags
    char labels[MAX_LABELS_PER_ISSUE][MAX_LABEL_NAME_LENGTH];
    uint32_t label_count;
    char tags[MAX_TAGS_PER_ISSUE][MAX_TAG_LENGTH];
    uint32_t tag_count;
    
    // Timestamps
    uint64_t created_timestamp;
    uint64_t updated_timestamp;
    uint64_t due_date;
    uint64_t resolved_timestamp;
    
    // Progress tracking
    uint32_t estimated_hours;
    uint32_t logged_hours;
    float progress_percentage;
    
    // Metrics
    uint32_t comment_count;
    uint32_t attachment_count;
    uint32_t view_count;
    uint32_t vote_count;
    
    // Flags
    bool is_locked;
    bool is_pinned;
    bool is_archived;
    bool has_subtasks;
} ProjectIssue;

typedef struct {
    char project_id[MAX_PROJECT_ID_LENGTH];
    char name[MAX_PROJECT_NAME_LENGTH];
    char description[MAX_PROJECT_DESCRIPTION_LENGTH];
    char owner_id[MAX_USER_ID_LENGTH];
    ProjectStatus status;
    ProjectPriority priority;
    
    // Timestamps
    uint64_t created_timestamp;
    uint64_t updated_timestamp;
    uint64_t start_date;
    uint64_t end_date;
    uint64_t last_activity_timestamp;
    
    // Statistics
    uint32_t total_issues;
    uint32_t open_issues;
    uint32_t closed_issues;
    uint32_t member_count;
    uint32_t milestone_count;
    
    // Progress tracking
    float completion_percentage;
    uint32_t total_estimated_hours;
    uint32_t total_logged_hours;
    
    // Settings
    bool is_public;
    bool allow_issues;
    bool allow_wiki;
    bool enable_notifications;
    
    // Repository info (if applicable)
    char repository_url[MAX_FILE_PATH_LENGTH];
    char default_branch[64];
    
    // Tags
    char tags[MAX_TAGS_PER_ISSUE][MAX_TAG_LENGTH];
    uint32_t tag_count;
} Project;

typedef struct {
    char notification_id[MAX_PROJECT_ID_LENGTH];
    NotificationType type;
    char project_id[MAX_PROJECT_ID_LENGTH];
    char issue_id[MAX_ISSUE_ID_LENGTH];
    char user_id[MAX_USER_ID_LENGTH];
    char title[MAX_ISSUE_TITLE_LENGTH];
    char message[MAX_COMMENT_LENGTH];
    uint64_t timestamp;
    bool is_read;
    bool is_important;
} ProjectNotification;

// MARK: - Configuration Structures

typedef struct {
    char server_url[256];
    uint16_t server_port;
    char user_id[MAX_USER_ID_LENGTH];
    char auth_token[512];
    char device_id[MAX_USER_ID_LENGTH];
    
    // Connection settings
    uint32_t connection_timeout;
    uint32_t heartbeat_interval;
    uint32_t sync_interval;
    uint32_t max_retries;
    
    // Feature flags
    bool enable_encryption;
    bool enable_compression;
    bool enable_notifications;
    bool enable_offline_mode;
    bool auto_sync_enabled;
    
    // Local storage
    char local_storage_path[MAX_FILE_PATH_LENGTH];
    uint64_t max_storage_size;
    uint32_t cache_retention_days;
    
    // UI preferences
    bool show_completed_issues;
    bool group_by_milestone;
    uint32_t items_per_page;
} ProjectManagerConfiguration;

// MARK: - Manager Structure

typedef struct ProjectManager ProjectManager;

// MARK: - Callback Function Types

typedef void (*ProjectStatusCallback)(ProjectStatus status, float progress);
typedef void (*ProjectDataCallback)(const Project* project, ProjectOperation operation);
typedef void (*IssueDataCallback)(const ProjectIssue* issue, ProjectOperation operation);
typedef void (*NotificationCallback)(const ProjectNotification* notification);
typedef void (*ProjectErrorCallback)(ProjectError error, const char* message);
typedef void (*SyncCompleteCallback)(uint32_t synced_projects, uint32_t synced_issues, uint32_t failed_items);

// Storage interface callbacks
typedef bool (*StoreProjectCallback)(const Project* project);
typedef bool (*RetrieveProjectCallback)(const char* project_id, Project* project);
typedef bool (*DeleteProjectCallback)(const char* project_id);
typedef bool (*ListProjectsCallback)(Project** projects, uint32_t* count);

typedef bool (*StoreIssueCallback)(const ProjectIssue* issue);
typedef bool (*RetrieveIssueCallback)(const char* issue_id, ProjectIssue* issue);
typedef bool (*DeleteIssueCallback)(const char* issue_id);
typedef bool (*ListIssuesCallback)(const char* project_id, ProjectIssue** issues, uint32_t* count);

typedef bool (*StoreCommentCallback)(const IssueComment* comment);
typedef bool (*RetrieveCommentsCallback)(const char* issue_id, IssueComment** comments, uint32_t* count);
typedef bool (*DeleteCommentCallback)(const char* comment_id);

typedef bool (*StoreAttachmentCallback)(const IssueAttachment* attachment, const void* data, uint32_t data_size);
typedef bool (*RetrieveAttachmentCallback)(const char* attachment_id, IssueAttachment* attachment, void** data, uint32_t* data_size);
typedef bool (*DeleteAttachmentCallback)(const char* attachment_id);

// MARK: - Public API Functions

// Manager lifecycle
ProjectManager* project_manager_create(const ProjectManagerConfiguration* config);
void project_manager_destroy(ProjectManager* manager);
bool project_manager_start(ProjectManager* manager);
void project_manager_stop(ProjectManager* manager);

// Connection management
bool project_manager_connect(ProjectManager* manager);
void project_manager_disconnect(ProjectManager* manager);
bool project_manager_is_connected(const ProjectManager* manager);

// Project operations
bool project_manager_create_project(ProjectManager* manager, const Project* project);
bool project_manager_update_project(ProjectManager* manager, const Project* project);
bool project_manager_delete_project(ProjectManager* manager, const char* project_id);
bool project_manager_get_project(ProjectManager* manager, const char* project_id, Project* project);
bool project_manager_list_projects(ProjectManager* manager, Project** projects, uint32_t* count);

// Issue operations
bool project_manager_create_issue(ProjectManager* manager, const ProjectIssue* issue);
bool project_manager_update_issue(ProjectManager* manager, const ProjectIssue* issue);
bool project_manager_delete_issue(ProjectManager* manager, const char* issue_id);
bool project_manager_get_issue(ProjectManager* manager, const char* issue_id, ProjectIssue* issue);
bool project_manager_list_issues(ProjectManager* manager, const char* project_id, ProjectIssue** issues, uint32_t* count);
bool project_manager_assign_issue(ProjectManager* manager, const char* issue_id, const char* assignee_id);
bool project_manager_update_issue_status(ProjectManager* manager, const char* issue_id, IssueStatus status);

// Comment operations
bool project_manager_add_comment(ProjectManager* manager, const IssueComment* comment);
bool project_manager_update_comment(ProjectManager* manager, const IssueComment* comment);
bool project_manager_delete_comment(ProjectManager* manager, const char* comment_id);
bool project_manager_get_comments(ProjectManager* manager, const char* issue_id, IssueComment** comments, uint32_t* count);

// Milestone operations
bool project_manager_create_milestone(ProjectManager* manager, const ProjectMilestone* milestone);
bool project_manager_update_milestone(ProjectManager* manager, const ProjectMilestone* milestone);
bool project_manager_delete_milestone(ProjectManager* manager, const char* milestone_id);
bool project_manager_list_milestones(ProjectManager* manager, const char* project_id, ProjectMilestone** milestones, uint32_t* count);

// Member operations
bool project_manager_add_member(ProjectManager* manager, const char* project_id, const ProjectMember* member);
bool project_manager_remove_member(ProjectManager* manager, const char* project_id, const char* user_id);
bool project_manager_update_member_role(ProjectManager* manager, const char* project_id, const char* user_id, ProjectRole role);
bool project_manager_list_members(ProjectManager* manager, const char* project_id, ProjectMember** members, uint32_t* count);

// Attachment operations
bool project_manager_upload_attachment(ProjectManager* manager, const char* issue_id, const IssueAttachment* attachment, const void* data, uint32_t data_size);
bool project_manager_download_attachment(ProjectManager* manager, const char* attachment_id, IssueAttachment* attachment, void** data, uint32_t* data_size);
bool project_manager_delete_attachment(ProjectManager* manager, const char* attachment_id);

// Synchronization
bool project_manager_sync_all(ProjectManager* manager);
bool project_manager_sync_project(ProjectManager* manager, const char* project_id);

// Status and monitoring
ProjectStatus project_manager_get_status(const ProjectManager* manager);
float project_manager_get_sync_progress(const ProjectManager* manager);
void project_manager_get_stats(const ProjectManager* manager, uint32_t* total_projects, uint32_t* total_issues, uint32_t* pending_sync);

// Notifications
bool project_manager_get_notifications(ProjectManager* manager, ProjectNotification** notifications, uint32_t* count);
bool project_manager_mark_notification_read(ProjectManager* manager, const char* notification_id);
bool project_manager_clear_notifications(ProjectManager* manager);

// Search and filtering
bool project_manager_search_issues(ProjectManager* manager, const char* project_id, const char* query, ProjectIssue** issues, uint32_t* count);
bool project_manager_filter_issues(ProjectManager* manager, const char* project_id, IssueStatus status, IssuePriority priority, const char* assignee_id, ProjectIssue** issues, uint32_t* count);

// Callback setters
void project_manager_set_status_callback(ProjectManager* manager, ProjectStatusCallback callback);
void project_manager_set_project_callback(ProjectManager* manager, ProjectDataCallback callback);
void project_manager_set_issue_callback(ProjectManager* manager, IssueDataCallback callback);
void project_manager_set_notification_callback(ProjectManager* manager, NotificationCallback callback);
void project_manager_set_error_callback(ProjectManager* manager, ProjectErrorCallback callback);
void project_manager_set_sync_complete_callback(ProjectManager* manager, SyncCompleteCallback callback);

// Storage interface setters
void project_manager_set_project_storage(ProjectManager* manager,
                                        StoreProjectCallback store_project,
                                        RetrieveProjectCallback retrieve_project,
                                        DeleteProjectCallback delete_project,
                                        ListProjectsCallback list_projects);

void project_manager_set_issue_storage(ProjectManager* manager,
                                      StoreIssueCallback store_issue,
                                      RetrieveIssueCallback retrieve_issue,
                                      DeleteIssueCallback delete_issue,
                                      ListIssuesCallback list_issues);

void project_manager_set_comment_storage(ProjectManager* manager,
                                        StoreCommentCallback store_comment,
                                        RetrieveCommentsCallback retrieve_comments,
                                        DeleteCommentCallback delete_comment);

void project_manager_set_attachment_storage(ProjectManager* manager,
                                           StoreAttachmentCallback store_attachment,
                                           RetrieveAttachmentCallback retrieve_attachment,
                                           DeleteAttachmentCallback delete_attachment);

// MARK: - Utility Functions

char* generate_project_id(void);
char* generate_issue_id(void);
char* generate_comment_id(void);
char* generate_milestone_id(void);
char* generate_attachment_id(void);
uint64_t get_current_timestamp_pm(void);
uint32_t calculate_project_checksum(const void* data, uint32_t length);
bool validate_project_data(const Project* project);
bool validate_issue_data(const ProjectIssue* issue);
float calculate_project_progress(const Project* project);
float calculate_milestone_progress(const ProjectMilestone* milestone);

// String conversion functions
const char* project_status_to_string(ProjectStatus status);
const char* issue_status_to_string(IssueStatus status);
const char* issue_type_to_string(IssueType type);
const char* issue_priority_to_string(IssuePriority priority);
const char* project_priority_to_string(ProjectPriority priority);
const char* project_role_to_string(ProjectRole role);
const char* project_error_to_string(ProjectError error);
const char* notification_type_to_string(NotificationType type);

ProjectStatus string_to_project_status(const char* status);
IssueStatus string_to_issue_status(const char* status);
IssueType string_to_issue_type(const char* type);
IssuePriority string_to_issue_priority(const char* priority);
ProjectPriority string_to_project_priority(const char* priority);
ProjectRole string_to_project_role(const char* role);

#ifdef __cplusplus
}
#endif

#endif // PROJECT_MANAGEMENT_H