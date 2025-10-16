#ifndef DATA_SYNC_H
#define DATA_SYNC_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

// MARK: - Constants

#define DATA_SYNC_PROTOCOL_VERSION 1
#define DATA_SYNC_MAGIC 0x44535950 // "DSYP"

#define MAX_SYNC_ID_LENGTH 64
#define MAX_SYNC_DATA_LENGTH 1048576 // 1MB
#define MAX_SYNC_METADATA_LENGTH 4096
#define MAX_SYNC_COLLECTIONS 100
#define MAX_SYNC_ITEMS_PER_BATCH 50
#define MAX_SYNC_CONFLICTS 10

#define DEFAULT_SYNC_PORT 8890
#define SYNC_HEARTBEAT_INTERVAL 30000 // 30 seconds
#define SYNC_RETRY_INTERVAL 5000      // 5 seconds
#define SYNC_TIMEOUT 60000            // 60 seconds

// MARK: - Enums

typedef enum {
    SYNC_TYPE_AI_CONVERSATION = 1,
    SYNC_TYPE_BOOKMARK = 2,
    SYNC_TYPE_PROJECT = 3,
    SYNC_TYPE_USER_PREFERENCE = 4,
    SYNC_TYPE_CUSTOM = 100
} SyncDataType;

typedef enum {
    SYNC_OPERATION_CREATE = 1,
    SYNC_OPERATION_UPDATE = 2,
    SYNC_OPERATION_DELETE = 3,
    SYNC_OPERATION_BATCH = 4
} SyncOperation;

typedef enum {
    SYNC_STATUS_IDLE = 0,
    SYNC_STATUS_CONNECTING = 1,
    SYNC_STATUS_AUTHENTICATING = 2,
    SYNC_STATUS_SYNCING = 3,
    SYNC_STATUS_CONFLICT_RESOLUTION = 4,
    SYNC_STATUS_COMPLETED = 5,
    SYNC_STATUS_ERROR = 6,
    SYNC_STATUS_OFFLINE = 7
} SyncStatus;

typedef enum {
    SYNC_CONFLICT_RESOLUTION_MANUAL = 0,
    SYNC_CONFLICT_RESOLUTION_LOCAL_WINS = 1,
    SYNC_CONFLICT_RESOLUTION_REMOTE_WINS = 2,
    SYNC_CONFLICT_RESOLUTION_MERGE = 3,
    SYNC_CONFLICT_RESOLUTION_LATEST_TIMESTAMP = 4
} SyncConflictResolution;

typedef enum {
    SYNC_ERROR_NONE = 0,
    SYNC_ERROR_NETWORK_FAILURE = 1,
    SYNC_ERROR_AUTH_FAILED = 2,
    SYNC_ERROR_PROTOCOL_ERROR = 3,
    SYNC_ERROR_DATA_CORRUPTION = 4,
    SYNC_ERROR_CONFLICT_UNRESOLVED = 5,
    SYNC_ERROR_STORAGE_FULL = 6,
    SYNC_ERROR_PERMISSION_DENIED = 7,
    SYNC_ERROR_INVALID_DATA = 8,
    SYNC_ERROR_VERSION_MISMATCH = 9,
    SYNC_ERROR_TIMEOUT = 10
} SyncError;

typedef enum {
    MSG_TYPE_SYNC_HANDSHAKE = 0x01,
    MSG_TYPE_SYNC_AUTH = 0x02,
    MSG_TYPE_SYNC_DATA = 0x03,
    MSG_TYPE_SYNC_ACK = 0x04,
    MSG_TYPE_SYNC_CONFLICT = 0x05,
    MSG_TYPE_SYNC_RESOLUTION = 0x06,
    MSG_TYPE_SYNC_HEARTBEAT = 0x07,
    MSG_TYPE_SYNC_STATUS = 0x08,
    MSG_TYPE_SYNC_ERROR = 0x09,
    MSG_TYPE_SYNC_COMPLETE = 0x0A
} SyncMessageType;

// MARK: - Core Data Structures

typedef struct {
    uint32_t magic;
    uint16_t version;
    SyncMessageType message_type;
    uint32_t message_id;
    uint32_t session_id;
    uint32_t data_length;
    uint32_t checksum;
    uint64_t timestamp;
    uint8_t reserved[8];
} SyncHeader;

typedef struct {
    char sync_id[MAX_SYNC_ID_LENGTH];
    SyncDataType data_type;
    SyncOperation operation;
    uint64_t timestamp;
    uint64_t version;
    uint32_t data_length;
    uint32_t metadata_length;
    uint32_t checksum;
    bool is_deleted;
    char device_id[MAX_SYNC_ID_LENGTH];
    char user_id[MAX_SYNC_ID_LENGTH];
} SyncItem;

typedef struct {
    SyncItem item;
    void* data;
    void* metadata;
} SyncData;

typedef struct {
    char conflict_id[MAX_SYNC_ID_LENGTH];
    SyncItem local_item;
    SyncItem remote_item;
    SyncConflictResolution resolution_strategy;
    uint64_t detected_timestamp;
    bool is_resolved;
} SyncConflict;

typedef struct {
    char collection_id[MAX_SYNC_ID_LENGTH];
    SyncDataType data_type;
    uint32_t item_count;
    uint64_t last_sync_timestamp;
    uint64_t version;
    bool is_dirty;
} SyncCollection;

// MARK: - Protocol Messages

typedef struct {
    char device_id[MAX_SYNC_ID_LENGTH];
    char device_name[MAX_SYNC_ID_LENGTH];
    uint16_t protocol_version;
    uint32_t supported_data_types;
    bool supports_encryption;
    bool supports_compression;
    uint32_t max_batch_size;
} SyncHandshakeRequest;

typedef struct {
    bool handshake_accepted;
    char session_id[MAX_SYNC_ID_LENGTH];
    uint16_t protocol_version;
    uint32_t supported_data_types;
    bool encryption_enabled;
    bool compression_enabled;
    uint32_t max_batch_size;
    SyncError error_code;
} SyncHandshakeResponse;

typedef struct {
    char user_id[MAX_SYNC_ID_LENGTH];
    char auth_token[256];
    char device_signature[256];
    uint64_t timestamp;
} SyncAuthRequest;

typedef struct {
    bool auth_success;
    char session_token[256];
    uint64_t token_expires;
    uint32_t permissions;
    SyncError error_code;
} SyncAuthResponse;

typedef struct {
    uint32_t batch_id;
    uint32_t item_count;
    uint32_t total_batches;
    uint32_t current_batch;
    SyncDataType data_type;
    bool is_final_batch;
} SyncBatchHeader;

typedef struct {
    uint32_t batch_id;
    uint32_t processed_items;
    uint32_t failed_items;
    uint32_t conflict_count;
    SyncError error_code;
    bool batch_complete;
} SyncBatchAck;

typedef struct {
    uint32_t conflict_count;
    SyncConflict conflicts[MAX_SYNC_CONFLICTS];
} SyncConflictMessage;

typedef struct {
    uint32_t resolution_count;
    struct {
        char conflict_id[MAX_SYNC_ID_LENGTH];
        SyncConflictResolution resolution;
        SyncItem resolved_item;
    } resolutions[MAX_SYNC_CONFLICTS];
} SyncResolutionMessage;

typedef struct {
    SyncStatus status;
    uint64_t timestamp;
    uint32_t items_synced;
    uint32_t items_pending;
    uint32_t conflicts_pending;
    float progress_percentage;
} SyncStatusMessage;

typedef struct {
    SyncError error_code;
    char error_message[256];
    char context[512];
    uint64_t timestamp;
    bool is_recoverable;
} SyncErrorMessage;

// MARK: - Manager Structures

typedef struct DataSyncManager DataSyncManager;

typedef struct {
    // Connection settings
    char server_url[256];
    uint16_t server_port;
    uint32_t connection_timeout;
    uint32_t sync_timeout;
    
    // Authentication
    char user_id[MAX_SYNC_ID_LENGTH];
    char auth_token[256];
    char device_id[MAX_SYNC_ID_LENGTH];
    
    // Sync settings
    bool auto_sync_enabled;
    uint32_t sync_interval;
    uint32_t max_batch_size;
    bool enable_encryption;
    bool enable_compression;
    
    // Conflict resolution
    SyncConflictResolution default_conflict_resolution;
    bool auto_resolve_conflicts;
    
    // Storage settings
    char local_storage_path[512];
    uint64_t max_storage_size;
    uint32_t max_history_entries;
} SyncConfiguration;

struct DataSyncManager {
    // Configuration
    SyncConfiguration config;
    
    // State
    SyncStatus status;
    bool is_running;
    bool is_connected;
    uint32_t session_id;
    char session_token[256];
    
    // Collections
    SyncCollection collections[MAX_SYNC_COLLECTIONS];
    uint32_t collection_count;
    
    // Sync state
    uint64_t last_sync_timestamp;
    uint32_t pending_items;
    uint32_t synced_items;
    uint32_t failed_items;
    
    // Conflicts
    SyncConflict active_conflicts[MAX_SYNC_CONFLICTS];
    uint32_t conflict_count;
    
    // Network
    int socket_fd;
    void* ssl_context;
    
    // Threading
    void* sync_thread;
    void* heartbeat_thread;
    bool shutdown_requested;
    void* mutex;
    void* condition;
    
    // Callbacks
    void (*status_callback)(SyncStatus status, float progress);
    void (*data_callback)(const SyncData* data, SyncOperation operation);
    void (*conflict_callback)(const SyncConflict* conflict);
    void (*error_callback)(SyncError error, const char* message);
    void (*complete_callback)(uint32_t synced_items, uint32_t failed_items);
    
    // Storage interface
    bool (*store_item)(const SyncData* data);
    bool (*retrieve_item)(const char* sync_id, SyncData* data);
    bool (*delete_item)(const char* sync_id);
    bool (*list_items)(SyncDataType type, SyncItem** items, uint32_t* count);
    bool (*update_collection)(const SyncCollection* collection);
};

// MARK: - Function Declarations

// Manager lifecycle
DataSyncManager* data_sync_manager_create(const SyncConfiguration* config);
void data_sync_manager_destroy(DataSyncManager* manager);
bool data_sync_manager_start(DataSyncManager* manager);
void data_sync_manager_stop(DataSyncManager* manager);

// Connection management
bool data_sync_manager_connect(DataSyncManager* manager);
void data_sync_manager_disconnect(DataSyncManager* manager);
bool data_sync_manager_is_connected(const DataSyncManager* manager);

// Sync operations
bool data_sync_manager_sync_all(DataSyncManager* manager);
bool data_sync_manager_sync_collection(DataSyncManager* manager, SyncDataType type);
bool data_sync_manager_sync_item(DataSyncManager* manager, const SyncData* data);

// Data operations
bool data_sync_manager_add_item(DataSyncManager* manager, const SyncData* data);
bool data_sync_manager_update_item(DataSyncManager* manager, const SyncData* data);
bool data_sync_manager_delete_item(DataSyncManager* manager, const char* sync_id);
bool data_sync_manager_get_item(DataSyncManager* manager, const char* sync_id, SyncData* data);

// Conflict resolution
bool data_sync_manager_resolve_conflict(DataSyncManager* manager, const char* conflict_id, 
                                       SyncConflictResolution resolution, const SyncItem* resolved_item);
uint32_t data_sync_manager_get_conflicts(const DataSyncManager* manager, SyncConflict* conflicts, uint32_t max_count);

// Configuration
bool data_sync_manager_update_config(DataSyncManager* manager, const SyncConfiguration* config);
void data_sync_manager_get_config(const DataSyncManager* manager, SyncConfiguration* config);

// Status and monitoring
SyncStatus data_sync_manager_get_status(const DataSyncManager* manager);
float data_sync_manager_get_progress(const DataSyncManager* manager);
void data_sync_manager_get_stats(const DataSyncManager* manager, uint32_t* synced, uint32_t* pending, uint32_t* failed);

// Callbacks
void data_sync_manager_set_status_callback(DataSyncManager* manager, 
                                          void (*callback)(SyncStatus status, float progress));
void data_sync_manager_set_data_callback(DataSyncManager* manager, 
                                        void (*callback)(const SyncData* data, SyncOperation operation));
void data_sync_manager_set_conflict_callback(DataSyncManager* manager, 
                                            void (*callback)(const SyncConflict* conflict));
void data_sync_manager_set_error_callback(DataSyncManager* manager, 
                                         void (*callback)(SyncError error, const char* message));
void data_sync_manager_set_complete_callback(DataSyncManager* manager, 
                                            void (*callback)(uint32_t synced_items, uint32_t failed_items));

// Storage interface
void data_sync_manager_set_storage_interface(DataSyncManager* manager,
                                            bool (*store_item)(const SyncData* data),
                                            bool (*retrieve_item)(const char* sync_id, SyncData* data),
                                            bool (*delete_item)(const char* sync_id),
                                            bool (*list_items)(SyncDataType type, SyncItem** items, uint32_t* count),
                                            bool (*update_collection)(const SyncCollection* collection));

// Utility functions
char* generate_sync_id(void);
uint64_t get_current_timestamp(void);
uint32_t calculate_data_checksum(const void* data, uint32_t length);
bool validate_sync_item(const SyncItem* item);
bool is_sync_item_newer(const SyncItem* item1, const SyncItem* item2);

// Message handling
bool send_sync_message(int socket_fd, const SyncHeader* header, const void* data);
bool receive_sync_message(int socket_fd, SyncHeader* header, void** data);
void free_sync_message_data(void* data);

// Encryption/Compression (if enabled)
bool encrypt_sync_data(const void* input, uint32_t input_length, void** output, uint32_t* output_length);
bool decrypt_sync_data(const void* input, uint32_t input_length, void** output, uint32_t* output_length);
bool compress_sync_data(const void* input, uint32_t input_length, void** output, uint32_t* output_length);
bool decompress_sync_data(const void* input, uint32_t input_length, void** output, uint32_t* output_length);

// Error handling
const char* sync_error_to_string(SyncError error);
const char* sync_status_to_string(SyncStatus status);
const char* sync_operation_to_string(SyncOperation operation);
const char* sync_data_type_to_string(SyncDataType type);

#ifdef __cplusplus
}
#endif

#endif // DATA_SYNC_H