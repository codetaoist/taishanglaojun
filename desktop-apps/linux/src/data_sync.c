#include "data_sync.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <openssl/ssl.h>
#include <openssl/err.h>
#include <openssl/sha.h>
#include <json-c/json.h>
#include <time.h>
#include <errno.h>
#include <signal.h>
#include <sys/stat.h>
#include <dirent.h>

// MARK: - Internal Structures

typedef struct {
    int socket_fd;
    SSL* ssl;
    struct sockaddr_in address;
    bool is_connected;
    pthread_t thread;
} ConnectionContext;

typedef struct {
    DataSyncManager* manager;
    SyncDataType data_type;
    SyncItem* items;
    uint32_t item_count;
    uint32_t batch_num;
    uint32_t total_batches;
} BatchContext;

// MARK: - Global Variables

static volatile bool g_shutdown_requested = false;

// MARK: - Signal Handlers

static void signal_handler(int sig) {
    g_shutdown_requested = true;
}

// MARK: - Private Function Declarations

static void* sync_thread_func(void* arg);
static void* heartbeat_thread_func(void* arg);
static bool initialize_ssl(DataSyncManager* manager);
static void cleanup_ssl(DataSyncManager* manager);
static bool perform_handshake(DataSyncManager* manager);
static bool authenticate(DataSyncManager* manager);
static bool send_batch(DataSyncManager* manager, SyncDataType type, SyncItem* items, 
                      uint32_t count, uint32_t batch_num, uint32_t total_batches);
static bool send_message(DataSyncManager* manager, const SyncHeader* header, const void* data);
static bool receive_message(DataSyncManager* manager, SyncHeader* header, void** data);
static void send_heartbeat(DataSyncManager* manager);
static void load_collections(DataSyncManager* manager);
static void save_collections(DataSyncManager* manager);
static void mark_collection_dirty(DataSyncManager* manager, SyncDataType type);
static void notify_status_change(DataSyncManager* manager);
static void handle_error(DataSyncManager* manager, SyncError error, const char* message);
static uint32_t generate_message_id(void);
static uint64_t get_current_timestamp_internal(void);
static uint32_t calculate_checksum_internal(const void* data, uint32_t length);
static char* generate_device_signature(const DataSyncManager* manager);

// MARK: - Data Sync Manager Implementation

struct DataSyncManager {
    SyncConfiguration config;
    SyncStatus status;
    bool is_running;
    bool is_connected;
    uint32_t session_id;
    char session_token[256];
    
    // Collections and sync state
    SyncCollection* collections;
    uint32_t collection_count;
    uint64_t last_sync_timestamp;
    uint32_t pending_items;
    uint32_t synced_items;
    uint32_t failed_items;
    
    // Conflicts
    SyncConflict* active_conflicts;
    uint32_t conflict_count;
    
    // Network
    int socket_fd;
    SSL_CTX* ssl_ctx;
    SSL* ssl;
    
    // Threading
    pthread_t sync_thread;
    pthread_t heartbeat_thread;
    pthread_mutex_t mutex;
    pthread_cond_t condition;
    bool shutdown_requested;
    
    // Callbacks
    SyncStatusCallback status_callback;
    SyncDataCallback data_callback;
    SyncConflictCallback conflict_callback;
    SyncErrorCallback error_callback;
    SyncCompleteCallback complete_callback;
    
    // Storage interface
    StoreItemCallback store_item;
    RetrieveItemCallback retrieve_item;
    DeleteItemCallback delete_item;
    ListItemsCallback list_items;
    UpdateCollectionCallback update_collection;
    
    // Local storage
    char storage_path[512];
};

// MARK: - Public API Implementation

DataSyncManager* data_sync_manager_create(const SyncConfiguration* config) {
    DataSyncManager* manager = calloc(1, sizeof(DataSyncManager));
    if (!manager) {
        return NULL;
    }
    
    // Copy configuration
    if (config) {
        manager->config = *config;
    } else {
        // Set default configuration
        strcpy(manager->config.server_url, "localhost");
        manager->config.server_port = 8443;
        strcpy(manager->config.device_id, "linux_device");
        strcpy(manager->config.user_id, "user");
        strcpy(manager->config.auth_token, "token");
        manager->config.enable_encryption = true;
        manager->config.enable_compression = true;
        manager->config.auto_sync_enabled = true;
        manager->config.sync_interval = 30000;
        manager->config.connection_timeout = 10000;
        manager->config.max_batch_size = 100;
        manager->config.max_retries = 3;
        manager->config.conflict_resolution = SYNC_CONFLICT_LATEST_TIMESTAMP;
    }
    
    // Setup storage path
    if (strlen(manager->config.local_storage_path) == 0) {
        const char* home = getenv("HOME");
        if (home) {
            snprintf(manager->storage_path, sizeof(manager->storage_path), 
                    "%s/.taishanglaojun/datasync", home);
        } else {
            strcpy(manager->storage_path, "./datasync");
        }
    } else {
        strcpy(manager->storage_path, manager->config.local_storage_path);
    }
    
    // Create storage directory
    char mkdir_cmd[1024];
    snprintf(mkdir_cmd, sizeof(mkdir_cmd), "mkdir -p %s", manager->storage_path);
    system(mkdir_cmd);
    
    // Initialize mutex and condition
    pthread_mutex_init(&manager->mutex, NULL);
    pthread_cond_init(&manager->condition, NULL);
    
    // Initialize OpenSSL
    SSL_library_init();
    SSL_load_error_strings();
    OpenSSL_add_all_algorithms();
    
    // Set up signal handlers
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    manager->status = SYNC_STATUS_IDLE;
    manager->socket_fd = -1;
    
    printf("Linux Data Sync Manager created\n");
    return manager;
}

void data_sync_manager_destroy(DataSyncManager* manager) {
    if (!manager) return;
    
    data_sync_manager_stop(manager);
    
    // Cleanup SSL
    cleanup_ssl(manager);
    
    // Cleanup collections
    if (manager->collections) {
        free(manager->collections);
    }
    
    // Cleanup conflicts
    if (manager->active_conflicts) {
        free(manager->active_conflicts);
    }
    
    // Cleanup mutex and condition
    pthread_mutex_destroy(&manager->mutex);
    pthread_cond_destroy(&manager->condition);
    
    free(manager);
    printf("Linux Data Sync Manager destroyed\n");
}

bool data_sync_manager_start(DataSyncManager* manager) {
    if (!manager) return false;
    
    pthread_mutex_lock(&manager->mutex);
    
    if (manager->is_running) {
        pthread_mutex_unlock(&manager->mutex);
        return true;
    }
    
    manager->shutdown_requested = false;
    g_shutdown_requested = false;
    
    // Initialize SSL context if encryption is enabled
    if (manager->config.enable_encryption) {
        if (!initialize_ssl(manager)) {
            pthread_mutex_unlock(&manager->mutex);
            return false;
        }
    }
    
    // Load local collections
    load_collections(manager);
    
    // Start sync thread
    if (pthread_create(&manager->sync_thread, NULL, sync_thread_func, manager) != 0) {
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Start heartbeat thread if auto-sync is enabled
    if (manager->config.auto_sync_enabled) {
        if (pthread_create(&manager->heartbeat_thread, NULL, heartbeat_thread_func, manager) != 0) {
            manager->shutdown_requested = true;
            pthread_join(manager->sync_thread, NULL);
            pthread_mutex_unlock(&manager->mutex);
            return false;
        }
    }
    
    manager->is_running = true;
    manager->status = SYNC_STATUS_IDLE;
    
    pthread_mutex_unlock(&manager->mutex);
    
    printf("Data sync manager started\n");
    return true;
}

void data_sync_manager_stop(DataSyncManager* manager) {
    if (!manager) return;
    
    pthread_mutex_lock(&manager->mutex);
    
    if (!manager->is_running) {
        pthread_mutex_unlock(&manager->mutex);
        return;
    }
    
    manager->shutdown_requested = true;
    g_shutdown_requested = true;
    pthread_cond_broadcast(&manager->condition);
    
    // Disconnect if connected
    if (manager->is_connected) {
        data_sync_manager_disconnect(manager);
    }
    
    manager->is_running = false;
    
    pthread_mutex_unlock(&manager->mutex);
    
    // Wait for threads to finish
    if (manager->sync_thread) {
        pthread_join(manager->sync_thread, NULL);
    }
    if (manager->heartbeat_thread) {
        pthread_join(manager->heartbeat_thread, NULL);
    }
    
    printf("Data sync manager stopped\n");
}

bool data_sync_manager_connect(DataSyncManager* manager) {
    if (!manager) return false;
    
    pthread_mutex_lock(&manager->mutex);
    
    if (manager->is_connected) {
        pthread_mutex_unlock(&manager->mutex);
        return true;
    }
    
    manager->status = SYNC_STATUS_CONNECTING;
    notify_status_change(manager);
    
    // Create socket
    manager->socket_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (manager->socket_fd < 0) {
        handle_error(manager, SYNC_ERROR_NETWORK_FAILURE, "Failed to create socket");
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Set socket timeout
    struct timeval timeout;
    timeout.tv_sec = manager->config.connection_timeout / 1000;
    timeout.tv_usec = (manager->config.connection_timeout % 1000) * 1000;
    setsockopt(manager->socket_fd, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout));
    setsockopt(manager->socket_fd, SOL_SOCKET, SO_SNDTIMEO, &timeout, sizeof(timeout));
    
    // Connect to server
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(manager->config.server_port);
    
    if (inet_pton(AF_INET, manager->config.server_url, &server_addr.sin_addr) <= 0) {
        // Try to resolve hostname
        struct hostent* host = gethostbyname(manager->config.server_url);
        if (!host) {
            close(manager->socket_fd);
            manager->socket_fd = -1;
            handle_error(manager, SYNC_ERROR_NETWORK_FAILURE, "Failed to resolve server address");
            pthread_mutex_unlock(&manager->mutex);
            return false;
        }
        
        memcpy(&server_addr.sin_addr, host->h_addr_list[0], host->h_length);
    }
    
    if (connect(manager->socket_fd, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        close(manager->socket_fd);
        manager->socket_fd = -1;
        handle_error(manager, SYNC_ERROR_NETWORK_FAILURE, "Failed to connect to server");
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Setup SSL if enabled
    if (manager->config.enable_encryption) {
        manager->ssl = SSL_new(manager->ssl_ctx);
        SSL_set_fd(manager->ssl, manager->socket_fd);
        
        if (SSL_connect(manager->ssl) <= 0) {
            SSL_free(manager->ssl);
            manager->ssl = NULL;
            close(manager->socket_fd);
            manager->socket_fd = -1;
            handle_error(manager, SYNC_ERROR_NETWORK_FAILURE, "SSL connection failed");
            pthread_mutex_unlock(&manager->mutex);
            return false;
        }
    }
    
    // Perform handshake
    if (!perform_handshake(manager)) {
        data_sync_manager_disconnect(manager);
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Authenticate
    if (!authenticate(manager)) {
        data_sync_manager_disconnect(manager);
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    manager->is_connected = true;
    manager->status = SYNC_STATUS_IDLE;
    notify_status_change(manager);
    
    pthread_mutex_unlock(&manager->mutex);
    
    printf("Connected to sync server\n");
    return true;
}

void data_sync_manager_disconnect(DataSyncManager* manager) {
    if (!manager) return;
    
    if (manager->ssl) {
        SSL_shutdown(manager->ssl);
        SSL_free(manager->ssl);
        manager->ssl = NULL;
    }
    
    if (manager->socket_fd >= 0) {
        close(manager->socket_fd);
        manager->socket_fd = -1;
    }
    
    manager->is_connected = false;
    manager->session_id = 0;
    memset(manager->session_token, 0, sizeof(manager->session_token));
    
    manager->status = SYNC_STATUS_OFFLINE;
    notify_status_change(manager);
    
    printf("Disconnected from sync server\n");
}

bool data_sync_manager_is_connected(const DataSyncManager* manager) {
    return manager ? manager->is_connected : false;
}

bool data_sync_manager_sync_all(DataSyncManager* manager) {
    if (!manager) return false;
    
    if (!manager->is_connected) {
        if (!data_sync_manager_connect(manager)) {
            return false;
        }
    }
    
    pthread_mutex_lock(&manager->mutex);
    
    manager->status = SYNC_STATUS_SYNCING;
    notify_status_change(manager);
    
    bool success = true;
    
    // Sync each collection
    for (uint32_t i = 0; i < manager->collection_count; i++) {
        if (!data_sync_manager_sync_collection(manager, manager->collections[i].data_type)) {
            success = false;
        }
    }
    
    manager->status = success ? SYNC_STATUS_COMPLETED : SYNC_STATUS_ERROR;
    notify_status_change(manager);
    
    if (manager->complete_callback) {
        manager->complete_callback(manager->synced_items, manager->failed_items);
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return success;
}

bool data_sync_manager_sync_collection(DataSyncManager* manager, SyncDataType type) {
    if (!manager) return false;
    
    // Get local items for this collection
    SyncItem* items = NULL;
    uint32_t count = 0;
    
    if (manager->list_items && !manager->list_items(type, &items, &count)) {
        return false;
    }
    
    // Send items in batches
    uint32_t batch_size = manager->config.max_batch_size;
    uint32_t total_batches = (count + batch_size - 1) / batch_size;
    
    bool success = true;
    
    for (uint32_t batch = 0; batch < total_batches; batch++) {
        uint32_t start_idx = batch * batch_size;
        uint32_t end_idx = (start_idx + batch_size < count) ? start_idx + batch_size : count;
        
        if (!send_batch(manager, type, &items[start_idx], end_idx - start_idx, batch, total_batches)) {
            success = false;
            break;
        }
    }
    
    if (items) {
        free(items);
    }
    
    return success;
}

bool data_sync_manager_add_item(DataSyncManager* manager, const SyncData* data) {
    if (!manager || !data) return false;
    
    pthread_mutex_lock(&manager->mutex);
    
    // Store locally
    if (manager->store_item && !manager->store_item(data)) {
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Mark collection as dirty
    mark_collection_dirty(manager, data->item.data_type);
    
    // Sync immediately if auto-sync is enabled and connected
    if (manager->config.auto_sync_enabled && manager->is_connected) {
        pthread_cond_signal(&manager->condition);
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return true;
}

bool data_sync_manager_update_item(DataSyncManager* manager, const SyncData* data) {
    if (!manager || !data) return false;
    
    pthread_mutex_lock(&manager->mutex);
    
    // Update locally
    if (manager->store_item && !manager->store_item(data)) {
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Mark collection as dirty
    mark_collection_dirty(manager, data->item.data_type);
    
    // Sync immediately if auto-sync is enabled and connected
    if (manager->config.auto_sync_enabled && manager->is_connected) {
        pthread_cond_signal(&manager->condition);
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return true;
}

bool data_sync_manager_delete_item(DataSyncManager* manager, const char* sync_id) {
    if (!manager || !sync_id) return false;
    
    pthread_mutex_lock(&manager->mutex);
    
    // Delete locally
    if (manager->delete_item && !manager->delete_item(sync_id)) {
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Sync immediately if auto-sync is enabled and connected
    if (manager->config.auto_sync_enabled && manager->is_connected) {
        pthread_cond_signal(&manager->condition);
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return true;
}

bool data_sync_manager_get_item(DataSyncManager* manager, const char* sync_id, SyncData* data) {
    if (!manager || !sync_id || !data) return false;
    
    pthread_mutex_lock(&manager->mutex);
    
    // Try storage interface
    bool result = false;
    if (manager->retrieve_item) {
        result = manager->retrieve_item(sync_id, data);
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return result;
}

SyncStatus data_sync_manager_get_status(const DataSyncManager* manager) {
    return manager ? manager->status : SYNC_STATUS_ERROR;
}

float data_sync_manager_get_progress(const DataSyncManager* manager) {
    if (!manager || manager->pending_items == 0) return 1.0f;
    return (float)manager->synced_items / (manager->synced_items + manager->pending_items);
}

void data_sync_manager_get_stats(const DataSyncManager* manager, uint32_t* synced, uint32_t* pending, uint32_t* failed) {
    if (!manager) return;
    
    if (synced) *synced = manager->synced_items;
    if (pending) *pending = manager->pending_items;
    if (failed) *failed = manager->failed_items;
}

// MARK: - Callback Setters

void data_sync_set_status_callback(DataSyncManager* manager, SyncStatusCallback callback) {
    if (manager) {
        manager->status_callback = callback;
    }
}

void data_sync_set_data_callback(DataSyncManager* manager, SyncDataCallback callback) {
    if (manager) {
        manager->data_callback = callback;
    }
}

void data_sync_set_conflict_callback(DataSyncManager* manager, SyncConflictCallback callback) {
    if (manager) {
        manager->conflict_callback = callback;
    }
}

void data_sync_set_error_callback(DataSyncManager* manager, SyncErrorCallback callback) {
    if (manager) {
        manager->error_callback = callback;
    }
}

void data_sync_set_complete_callback(DataSyncManager* manager, SyncCompleteCallback callback) {
    if (manager) {
        manager->complete_callback = callback;
    }
}

// MARK: - Storage Interface Setters

void data_sync_set_storage_interface(DataSyncManager* manager,
                                   StoreItemCallback store_item,
                                   RetrieveItemCallback retrieve_item,
                                   DeleteItemCallback delete_item,
                                   ListItemsCallback list_items,
                                   UpdateCollectionCallback update_collection) {
    if (!manager) return;
    
    manager->store_item = store_item;
    manager->retrieve_item = retrieve_item;
    manager->delete_item = delete_item;
    manager->list_items = list_items;
    manager->update_collection = update_collection;
}

// MARK: - Private Function Implementations

static void* sync_thread_func(void* arg) {
    DataSyncManager* manager = (DataSyncManager*)arg;
    
    while (!manager->shutdown_requested && !g_shutdown_requested) {
        pthread_mutex_lock(&manager->mutex);
        
        // Wait for sync trigger or timeout
        struct timespec timeout;
        clock_gettime(CLOCK_REALTIME, &timeout);
        timeout.tv_sec += manager->config.sync_interval / 1000;
        timeout.tv_nsec += (manager->config.sync_interval % 1000) * 1000000;
        
        int result = pthread_cond_timedwait(&manager->condition, &manager->mutex, &timeout);
        
        if (manager->shutdown_requested || g_shutdown_requested) {
            pthread_mutex_unlock(&manager->mutex);
            break;
        }
        
        if (manager->config.auto_sync_enabled && manager->is_connected) {
            pthread_mutex_unlock(&manager->mutex);
            data_sync_manager_sync_all(manager);
        } else {
            pthread_mutex_unlock(&manager->mutex);
        }
    }
    
    return NULL;
}

static void* heartbeat_thread_func(void* arg) {
    DataSyncManager* manager = (DataSyncManager*)arg;
    
    while (!manager->shutdown_requested && !g_shutdown_requested) {
        sleep(30); // 30 seconds heartbeat interval
        
        if (manager->shutdown_requested || g_shutdown_requested) {
            break;
        }
        
        if (manager->is_connected) {
            send_heartbeat(manager);
        }
    }
    
    return NULL;
}

static bool initialize_ssl(DataSyncManager* manager) {
    manager->ssl_ctx = SSL_CTX_new(TLS_client_method());
    if (!manager->ssl_ctx) {
        return false;
    }
    
    // Set verification mode
    SSL_CTX_set_verify(manager->ssl_ctx, SSL_VERIFY_PEER, NULL);
    
    // Load default CA certificates
    SSL_CTX_set_default_verify_paths(manager->ssl_ctx);
    
    return true;
}

static void cleanup_ssl(DataSyncManager* manager) {
    if (manager->ssl) {
        SSL_free(manager->ssl);
        manager->ssl = NULL;
    }
    if (manager->ssl_ctx) {
        SSL_CTX_free(manager->ssl_ctx);
        manager->ssl_ctx = NULL;
    }
}

static bool perform_handshake(DataSyncManager* manager) {
    // Create handshake request
    json_object* request = json_object_new_object();
    json_object_object_add(request, "device_id", json_object_new_string(manager->config.device_id));
    json_object_object_add(request, "device_name", json_object_new_string("Linux Desktop"));
    json_object_object_add(request, "protocol_version", json_object_new_int(DATA_SYNC_PROTOCOL_VERSION));
    json_object_object_add(request, "supported_data_types", json_object_new_int(0xFFFFFFFF));
    json_object_object_add(request, "supports_encryption", json_object_new_boolean(manager->config.enable_encryption));
    json_object_object_add(request, "supports_compression", json_object_new_boolean(manager->config.enable_compression));
    json_object_object_add(request, "max_batch_size", json_object_new_int(manager->config.max_batch_size));
    
    const char* request_str = json_object_to_json_string(request);
    
    SyncHeader header = {
        .magic = DATA_SYNC_MAGIC,
        .version = DATA_SYNC_PROTOCOL_VERSION,
        .message_type = MSG_TYPE_SYNC_HANDSHAKE,
        .message_id = generate_message_id(),
        .session_id = 0,
        .data_length = strlen(request_str),
        .timestamp = get_current_timestamp_internal()
    };
    header.checksum = calculate_checksum_internal(request_str, header.data_length);
    
    if (!send_message(manager, &header, request_str)) {
        json_object_put(request);
        handle_error(manager, SYNC_ERROR_PROTOCOL_ERROR, "Failed to send handshake request");
        return false;
    }
    
    json_object_put(request);
    
    // Receive response
    SyncHeader response_header;
    void* response_data = NULL;
    
    if (!receive_message(manager, &response_header, &response_data)) {
        handle_error(manager, SYNC_ERROR_PROTOCOL_ERROR, "Failed to receive handshake response");
        return false;
    }
    
    if (response_header.message_type != MSG_TYPE_SYNC_HANDSHAKE) {
        free(response_data);
        handle_error(manager, SYNC_ERROR_PROTOCOL_ERROR, "Invalid handshake response");
        return false;
    }
    
    // Parse response
    json_object* response = json_tokener_parse((char*)response_data);
    json_object* accepted_obj;
    
    if (!json_object_object_get_ex(response, "handshake_accepted", &accepted_obj) ||
        !json_object_get_boolean(accepted_obj)) {
        json_object_put(response);
        free(response_data);
        handle_error(manager, SYNC_ERROR_PROTOCOL_ERROR, "Handshake rejected");
        return false;
    }
    
    // Update configuration based on server capabilities
    json_object* max_batch_obj;
    if (json_object_object_get_ex(response, "max_batch_size", &max_batch_obj)) {
        uint32_t server_max_batch = json_object_get_int(max_batch_obj);
        if (server_max_batch < manager->config.max_batch_size) {
            manager->config.max_batch_size = server_max_batch;
        }
    }
    
    json_object_put(response);
    free(response_data);
    return true;
}

static bool authenticate(DataSyncManager* manager) {
    manager->status = SYNC_STATUS_AUTHENTICATING;
    notify_status_change(manager);
    
    // Create auth request
    json_object* request = json_object_new_object();
    json_object_object_add(request, "user_id", json_object_new_string(manager->config.user_id));
    json_object_object_add(request, "auth_token", json_object_new_string(manager->config.auth_token));
    
    char* device_signature = generate_device_signature(manager);
    json_object_object_add(request, "device_signature", json_object_new_string(device_signature));
    json_object_object_add(request, "timestamp", json_object_new_int64(get_current_timestamp_internal()));
    
    const char* request_str = json_object_to_json_string(request);
    
    SyncHeader header = {
        .magic = DATA_SYNC_MAGIC,
        .version = DATA_SYNC_PROTOCOL_VERSION,
        .message_type = MSG_TYPE_SYNC_AUTH,
        .message_id = generate_message_id(),
        .session_id = 0,
        .data_length = strlen(request_str),
        .timestamp = get_current_timestamp_internal()
    };
    header.checksum = calculate_checksum_internal(request_str, header.data_length);
    
    if (!send_message(manager, &header, request_str)) {
        json_object_put(request);
        free(device_signature);
        handle_error(manager, SYNC_ERROR_AUTH_FAILED, "Failed to send auth request");
        return false;
    }
    
    json_object_put(request);
    free(device_signature);
    
    // Receive response
    SyncHeader response_header;
    void* response_data = NULL;
    
    if (!receive_message(manager, &response_header, &response_data)) {
        handle_error(manager, SYNC_ERROR_AUTH_FAILED, "Failed to receive auth response");
        return false;
    }
    
    if (response_header.message_type != MSG_TYPE_SYNC_AUTH) {
        free(response_data);
        handle_error(manager, SYNC_ERROR_PROTOCOL_ERROR, "Invalid auth response");
        return false;
    }
    
    // Parse response
    json_object* response = json_tokener_parse((char*)response_data);
    json_object* success_obj;
    
    if (!json_object_object_get_ex(response, "auth_success", &success_obj) ||
        !json_object_get_boolean(success_obj)) {
        json_object_put(response);
        free(response_data);
        handle_error(manager, SYNC_ERROR_AUTH_FAILED, "Authentication failed");
        return false;
    }
    
    // Store session info
    manager->session_id = response_header.session_id;
    
    json_object* token_obj;
    if (json_object_object_get_ex(response, "session_token", &token_obj)) {
        strncpy(manager->session_token, json_object_get_string(token_obj), 
                sizeof(manager->session_token) - 1);
    }
    
    json_object_put(response);
    free(response_data);
    return true;
}

static bool send_batch(DataSyncManager* manager, SyncDataType type, SyncItem* items, 
                      uint32_t count, uint32_t batch_num, uint32_t total_batches) {
    // Create batch JSON
    json_object* batch = json_object_new_object();
    json_object_object_add(batch, "batch_id", json_object_new_int(generate_message_id()));
    json_object_object_add(batch, "item_count", json_object_new_int(count));
    json_object_object_add(batch, "total_batches", json_object_new_int(total_batches));
    json_object_object_add(batch, "current_batch", json_object_new_int(batch_num));
    json_object_object_add(batch, "data_type", json_object_new_int(type));
    json_object_object_add(batch, "is_final_batch", json_object_new_boolean(batch_num == total_batches - 1));
    
    json_object* items_array = json_object_new_array();
    
    for (uint32_t i = 0; i < count; i++) {
        json_object* item_obj = json_object_new_object();
        json_object_object_add(item_obj, "sync_id", json_object_new_string(items[i].sync_id));
        json_object_object_add(item_obj, "data_type", json_object_new_int(items[i].data_type));
        json_object_object_add(item_obj, "operation", json_object_new_int(items[i].operation));
        json_object_object_add(item_obj, "timestamp", json_object_new_int64(items[i].timestamp));
        json_object_object_add(item_obj, "version", json_object_new_int64(items[i].version));
        json_object_object_add(item_obj, "checksum", json_object_new_int(items[i].checksum));
        json_object_object_add(item_obj, "device_id", json_object_new_string(items[i].device_id));
        json_object_object_add(item_obj, "user_id", json_object_new_string(items[i].user_id));
        
        // Get item data
        SyncData sync_data;
        if (manager->retrieve_item && manager->retrieve_item(items[i].sync_id, &sync_data)) {
            if (sync_data.data && sync_data.item.data_length > 0) {
                // Encode data as base64 or hex (simplified as string for now)
                json_object_object_add(item_obj, "data", json_object_new_string("data_placeholder"));
            }
            if (sync_data.metadata && sync_data.item.metadata_length > 0) {
                json_object_object_add(item_obj, "metadata", json_object_new_string("metadata_placeholder"));
            }
        }
        
        json_object_array_add(items_array, item_obj);
    }
    
    json_object_object_add(batch, "items", items_array);
    
    const char* batch_str = json_object_to_json_string(batch);
    
    SyncHeader header = {
        .magic = DATA_SYNC_MAGIC,
        .version = DATA_SYNC_PROTOCOL_VERSION,
        .message_type = MSG_TYPE_SYNC_DATA,
        .message_id = generate_message_id(),
        .session_id = manager->session_id,
        .data_length = strlen(batch_str),
        .timestamp = get_current_timestamp_internal()
    };
    header.checksum = calculate_checksum_internal(batch_str, header.data_length);
    
    if (!send_message(manager, &header, batch_str)) {
        json_object_put(batch);
        return false;
    }
    
    json_object_put(batch);
    
    // Wait for acknowledgment
    SyncHeader ack_header;
    void* ack_data = NULL;
    
    if (!receive_message(manager, &ack_header, &ack_data)) {
        return false;
    }
    
    if (ack_header.message_type != MSG_TYPE_SYNC_ACK) {
        free(ack_data);
        return false;
    }
    
    // Parse acknowledgment
    json_object* ack = json_tokener_parse((char*)ack_data);
    json_object* processed_obj, *failed_obj, *complete_obj;
    
    if (json_object_object_get_ex(ack, "processed_items", &processed_obj)) {
        manager->synced_items += json_object_get_int(processed_obj);
    }
    
    if (json_object_object_get_ex(ack, "failed_items", &failed_obj)) {
        manager->failed_items += json_object_get_int(failed_obj);
    }
    
    bool batch_complete = false;
    if (json_object_object_get_ex(ack, "batch_complete", &complete_obj)) {
        batch_complete = json_object_get_boolean(complete_obj);
    }
    
    json_object_put(ack);
    free(ack_data);
    
    return batch_complete;
}

static bool send_message(DataSyncManager* manager, const SyncHeader* header, const void* data) {
    // Send header
    ssize_t bytes_sent;
    if (manager->ssl) {
        bytes_sent = SSL_write(manager->ssl, header, sizeof(SyncHeader));
    } else {
        bytes_sent = send(manager->socket_fd, header, sizeof(SyncHeader), 0);
    }
    
    if (bytes_sent != sizeof(SyncHeader)) {
        return false;
    }
    
    // Send data if present
    if (data && header->data_length > 0) {
        if (manager->ssl) {
            bytes_sent = SSL_write(manager->ssl, data, header->data_length);
        } else {
            bytes_sent = send(manager->socket_fd, data, header->data_length, 0);
        }
        
        if (bytes_sent != header->data_length) {
            return false;
        }
    }
    
    return true;
}

static bool receive_message(DataSyncManager* manager, SyncHeader* header, void** data) {
    // Receive header
    ssize_t bytes_received;
    if (manager->ssl) {
        bytes_received = SSL_read(manager->ssl, header, sizeof(SyncHeader));
    } else {
        bytes_received = recv(manager->socket_fd, header, sizeof(SyncHeader), 0);
    }
    
    if (bytes_received != sizeof(SyncHeader)) {
        return false;
    }
    
    // Validate header
    if (header->magic != DATA_SYNC_MAGIC || header->version != DATA_SYNC_PROTOCOL_VERSION) {
        return false;
    }
    
    // Receive data if present
    *data = NULL;
    if (header->data_length > 0) {
        *data = malloc(header->data_length + 1); // +1 for null terminator
        if (!*data) {
            return false;
        }
        
        if (manager->ssl) {
            bytes_received = SSL_read(manager->ssl, *data, header->data_length);
        } else {
            bytes_received = recv(manager->socket_fd, *data, header->data_length, 0);
        }
        
        if (bytes_received != header->data_length) {
            free(*data);
            *data = NULL;
            return false;
        }
        
        // Add null terminator for string data
        ((char*)*data)[header->data_length] = '\0';
        
        // Verify checksum
        uint32_t calculated_checksum = calculate_checksum_internal(*data, header->data_length);
        if (calculated_checksum != header->checksum) {
            free(*data);
            *data = NULL;
            return false;
        }
    }
    
    return true;
}

static void send_heartbeat(DataSyncManager* manager) {
    SyncHeader header = {
        .magic = DATA_SYNC_MAGIC,
        .version = DATA_SYNC_PROTOCOL_VERSION,
        .message_type = MSG_TYPE_SYNC_HEARTBEAT,
        .message_id = generate_message_id(),
        .session_id = manager->session_id,
        .data_length = 0,
        .checksum = 0,
        .timestamp = get_current_timestamp_internal()
    };
    
    if (!send_message(manager, &header, NULL)) {
        // Heartbeat failed, disconnect
        data_sync_manager_disconnect(manager);
    }
}

static void load_collections(DataSyncManager* manager) {
    char collections_file[1024];
    snprintf(collections_file, sizeof(collections_file), "%s/collections.json", manager->storage_path);
    
    FILE* file = fopen(collections_file, "r");
    if (!file) {
        return;
    }
    
    fseek(file, 0, SEEK_END);
    long file_size = ftell(file);
    fseek(file, 0, SEEK_SET);
    
    char* file_content = malloc(file_size + 1);
    if (!file_content) {
        fclose(file);
        return;
    }
    
    fread(file_content, 1, file_size, file);
    file_content[file_size] = '\0';
    fclose(file);
    
    json_object* root = json_tokener_parse(file_content);
    free(file_content);
    
    if (!root) {
        return;
    }
    
    json_object* collections_array;
    if (json_object_object_get_ex(root, "collections", &collections_array)) {
        int array_len = json_object_array_length(collections_array);
        
        if (manager->collections) {
            free(manager->collections);
        }
        
        manager->collections = calloc(array_len, sizeof(SyncCollection));
        manager->collection_count = array_len;
        
        for (int i = 0; i < array_len; i++) {
            json_object* item = json_object_array_get_idx(collections_array, i);
            SyncCollection* collection = &manager->collections[i];
            
            json_object* id_obj, *type_obj, *count_obj, *sync_obj, *version_obj, *dirty_obj;
            
            if (json_object_object_get_ex(item, "id", &id_obj)) {
                strncpy(collection->collection_id, json_object_get_string(id_obj), MAX_SYNC_ID_LENGTH - 1);
            }
            if (json_object_object_get_ex(item, "type", &type_obj)) {
                collection->data_type = json_object_get_int(type_obj);
            }
            if (json_object_object_get_ex(item, "count", &count_obj)) {
                collection->item_count = json_object_get_int(count_obj);
            }
            if (json_object_object_get_ex(item, "last_sync", &sync_obj)) {
                collection->last_sync_timestamp = json_object_get_int64(sync_obj);
            }
            if (json_object_object_get_ex(item, "version", &version_obj)) {
                collection->version = json_object_get_int64(version_obj);
            }
            if (json_object_object_get_ex(item, "dirty", &dirty_obj)) {
                collection->is_dirty = json_object_get_boolean(dirty_obj);
            }
        }
    }
    
    json_object_put(root);
}

static void save_collections(DataSyncManager* manager) {
    json_object* root = json_object_new_object();
    json_object* collections_array = json_object_new_array();
    
    for (uint32_t i = 0; i < manager->collection_count; i++) {
        SyncCollection* collection = &manager->collections[i];
        
        json_object* item = json_object_new_object();
        json_object_object_add(item, "id", json_object_new_string(collection->collection_id));
        json_object_object_add(item, "type", json_object_new_int(collection->data_type));
        json_object_object_add(item, "count", json_object_new_int(collection->item_count));
        json_object_object_add(item, "last_sync", json_object_new_int64(collection->last_sync_timestamp));
        json_object_object_add(item, "version", json_object_new_int64(collection->version));
        json_object_object_add(item, "dirty", json_object_new_boolean(collection->is_dirty));
        
        json_object_array_add(collections_array, item);
    }
    
    json_object_object_add(root, "collections", collections_array);
    
    char collections_file[1024];
    snprintf(collections_file, sizeof(collections_file), "%s/collections.json", manager->storage_path);
    
    FILE* file = fopen(collections_file, "w");
    if (file) {
        const char* json_string = json_object_to_json_string_ext(root, JSON_C_TO_STRING_PRETTY);
        fprintf(file, "%s", json_string);
        fclose(file);
    }
    
    json_object_put(root);
}

static void mark_collection_dirty(DataSyncManager* manager, SyncDataType type) {
    for (uint32_t i = 0; i < manager->collection_count; i++) {
        if (manager->collections[i].data_type == type) {
            manager->collections[i].is_dirty = true;
            break;
        }
    }
    save_collections(manager);
}

static void notify_status_change(DataSyncManager* manager) {
    if (manager->status_callback) {
        float progress = data_sync_manager_get_progress(manager);
        manager->status_callback(manager->status, progress);
    }
}

static void handle_error(DataSyncManager* manager, SyncError error, const char* message) {
    manager->status = SYNC_STATUS_ERROR;
    
    if (manager->error_callback) {
        manager->error_callback(error, message);
    }
    
    printf("Sync error: %s\n", message);
}

static uint32_t generate_message_id(void) {
    static uint32_t counter = 0;
    return ++counter;
}

static uint64_t get_current_timestamp_internal(void) {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (uint64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

static uint32_t calculate_checksum_internal(const void* data, uint32_t length) {
    const uint8_t* bytes = (const uint8_t*)data;
    uint32_t checksum = 0;
    
    for (uint32_t i = 0; i < length; i++) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
}

static char* generate_device_signature(const DataSyncManager* manager) {
    char* signature = malloc(256);
    if (signature) {
        snprintf(signature, 256, "%s_%lu", manager->config.device_id, get_current_timestamp_internal());
    }
    return signature;
}

// MARK: - Utility Functions

char* generate_sync_id(void) {
    static uint32_t counter = 0;
    char* id = malloc(MAX_SYNC_ID_LENGTH);
    if (id) {
        snprintf(id, MAX_SYNC_ID_LENGTH, "SYNC_%08X_%08X", 
                (uint32_t)time(NULL), ++counter);
    }
    return id;
}

uint64_t get_current_timestamp(void) {
    return get_current_timestamp_internal();
}

uint32_t calculate_data_checksum(const void* data, uint32_t length) {
    return calculate_checksum_internal(data, length);
}

const char* sync_error_to_string(SyncError error) {
    switch (error) {
        case SYNC_ERROR_NONE: return "No error";
        case SYNC_ERROR_NETWORK_FAILURE: return "Network failure";
        case SYNC_ERROR_AUTH_FAILED: return "Authentication failed";
        case SYNC_ERROR_PROTOCOL_ERROR: return "Protocol error";
        case SYNC_ERROR_DATA_CORRUPTION: return "Data corruption";
        case SYNC_ERROR_CONFLICT_UNRESOLVED: return "Conflict unresolved";
        case SYNC_ERROR_STORAGE_FULL: return "Storage full";
        case SYNC_ERROR_PERMISSION_DENIED: return "Permission denied";
        case SYNC_ERROR_INVALID_DATA: return "Invalid data";
        case SYNC_ERROR_VERSION_MISMATCH: return "Version mismatch";
        case SYNC_ERROR_TIMEOUT: return "Timeout";
        default: return "Unknown error";
    }
}

const char* sync_status_to_string(SyncStatus status) {
    switch (status) {
        case SYNC_STATUS_IDLE: return "Idle";
        case SYNC_STATUS_CONNECTING: return "Connecting";
        case SYNC_STATUS_AUTHENTICATING: return "Authenticating";
        case SYNC_STATUS_SYNCING: return "Syncing";
        case SYNC_STATUS_CONFLICT_RESOLUTION: return "Resolving conflicts";
        case SYNC_STATUS_COMPLETED: return "Completed";
        case SYNC_STATUS_ERROR: return "Error";
        case SYNC_STATUS_OFFLINE: return "Offline";
        default: return "Unknown";
    }
}