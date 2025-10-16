#include "data_sync.h"
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
#include <mutex>
#include <thread>
#include <condition_variable>
#include <chrono>
#include <fstream>
#include <sstream>

#pragma comment(lib, "ws2_32.lib")
#pragma comment(lib, "libssl.lib")
#pragma comment(lib, "libcrypto.lib")

// MARK: - Windows Data Sync Manager Implementation

class WindowsDataSyncManager {
private:
    SyncConfiguration config_;
    SyncStatus status_;
    bool is_running_;
    bool is_connected_;
    uint32_t session_id_;
    std::string session_token_;
    
    // Collections and sync state
    std::vector<SyncCollection> collections_;
    uint64_t last_sync_timestamp_;
    uint32_t pending_items_;
    uint32_t synced_items_;
    uint32_t failed_items_;
    
    // Conflicts
    std::vector<SyncConflict> active_conflicts_;
    
    // Network
    SOCKET socket_fd_;
    SSL_CTX* ssl_ctx_;
    SSL* ssl_;
    
    // Threading
    std::thread sync_thread_;
    std::thread heartbeat_thread_;
    bool shutdown_requested_;
    std::mutex mutex_;
    std::condition_variable condition_;
    
    // Callbacks
    std::function<void(SyncStatus, float)> status_callback_;
    std::function<void(const SyncData*, SyncOperation)> data_callback_;
    std::function<void(const SyncConflict*)> conflict_callback_;
    std::function<void(SyncError, const char*)> error_callback_;
    std::function<void(uint32_t, uint32_t)> complete_callback_;
    
    // Storage interface
    std::function<bool(const SyncData*)> store_item_;
    std::function<bool(const char*, SyncData*)> retrieve_item_;
    std::function<bool(const char*)> delete_item_;
    std::function<bool(SyncDataType, SyncItem**, uint32_t*)> list_items_;
    std::function<bool(const SyncCollection*)> update_collection_;
    
    // Local storage
    std::map<std::string, SyncData> local_cache_;
    std::string storage_path_;

public:
    WindowsDataSyncManager(const SyncConfiguration* config) 
        : status_(SYNC_STATUS_IDLE)
        , is_running_(false)
        , is_connected_(false)
        , session_id_(0)
        , last_sync_timestamp_(0)
        , pending_items_(0)
        , synced_items_(0)
        , failed_items_(0)
        , socket_fd_(INVALID_SOCKET)
        , ssl_ctx_(nullptr)
        , ssl_(nullptr)
        , shutdown_requested_(false) {
        
        if (config) {
            config_ = *config;
        }
        
        // Initialize storage path
        storage_path_ = config_.local_storage_path;
        if (storage_path_.empty()) {
            char app_data[MAX_PATH];
            if (SHGetFolderPathA(NULL, CSIDL_APPDATA, NULL, 0, app_data) == S_OK) {
                storage_path_ = std::string(app_data) + "\\TaiShangLaoJun\\DataSync";
            } else {
                storage_path_ = ".\\DataSync";
            }
        }
        
        // Create storage directory
        CreateDirectoryA(storage_path_.c_str(), NULL);
        
        // Initialize Winsock
        WSADATA wsaData;
        WSAStartup(MAKEWORD(2, 2), &wsaData);
        
        // Initialize OpenSSL
        SSL_library_init();
        SSL_load_error_strings();
        OpenSSL_add_all_algorithms();
        
        std::cout << "Windows Data Sync Manager created" << std::endl;
    }
    
    ~WindowsDataSyncManager() {
        stop();
        
        // Cleanup SSL
        if (ssl_) {
            SSL_free(ssl_);
        }
        if (ssl_ctx_) {
            SSL_CTX_free(ssl_ctx_);
        }
        
        // Cleanup Winsock
        WSACleanup();
        
        std::cout << "Windows Data Sync Manager destroyed" << std::endl;
    }
    
    bool start() {
        std::lock_guard<std::mutex> lock(mutex_);
        
        if (is_running_) {
            return true;
        }
        
        shutdown_requested_ = false;
        
        // Initialize SSL context if encryption is enabled
        if (config_.enable_encryption) {
            if (!initializeSSL()) {
                return false;
            }
        }
        
        // Load local collections
        loadCollections();
        
        // Start sync thread
        sync_thread_ = std::thread(&WindowsDataSyncManager::syncThreadFunc, this);
        
        // Start heartbeat thread if connected
        if (config_.auto_sync_enabled) {
            heartbeat_thread_ = std::thread(&WindowsDataSyncManager::heartbeatThreadFunc, this);
        }
        
        is_running_ = true;
        status_ = SYNC_STATUS_IDLE;
        
        std::cout << "Data sync manager started" << std::endl;
        return true;
    }
    
    void stop() {
        std::lock_guard<std::mutex> lock(mutex_);
        
        if (!is_running_) {
            return;
        }
        
        shutdown_requested_ = true;
        condition_.notify_all();
        
        // Disconnect if connected
        if (is_connected_) {
            disconnect();
        }
        
        is_running_ = false;
        
        // Wait for threads to finish
        if (sync_thread_.joinable()) {
            sync_thread_.join();
        }
        if (heartbeat_thread_.joinable()) {
            heartbeat_thread_.join();
        }
        
        std::cout << "Data sync manager stopped" << std::endl;
    }
    
    bool connect() {
        std::lock_guard<std::mutex> lock(mutex_);
        
        if (is_connected_) {
            return true;
        }
        
        status_ = SYNC_STATUS_CONNECTING;
        notifyStatusChange();
        
        // Create socket
        socket_fd_ = socket(AF_INET, SOCK_STREAM, 0);
        if (socket_fd_ == INVALID_SOCKET) {
            handleError(SYNC_ERROR_NETWORK_FAILURE, "Failed to create socket");
            return false;
        }
        
        // Set socket timeout
        DWORD timeout = config_.connection_timeout;
        setsockopt(socket_fd_, SOL_SOCKET, SO_RCVTIMEO, (char*)&timeout, sizeof(timeout));
        setsockopt(socket_fd_, SOL_SOCKET, SO_SNDTIMEO, (char*)&timeout, sizeof(timeout));
        
        // Connect to server
        sockaddr_in server_addr;
        memset(&server_addr, 0, sizeof(server_addr));
        server_addr.sin_family = AF_INET;
        server_addr.sin_port = htons(config_.server_port);
        
        if (inet_pton(AF_INET, config_.server_url, &server_addr.sin_addr) <= 0) {
            // Try to resolve hostname
            struct addrinfo hints, *result;
            memset(&hints, 0, sizeof(hints));
            hints.ai_family = AF_INET;
            hints.ai_socktype = SOCK_STREAM;
            
            if (getaddrinfo(config_.server_url, std::to_string(config_.server_port).c_str(), &hints, &result) != 0) {
                closesocket(socket_fd_);
                socket_fd_ = INVALID_SOCKET;
                handleError(SYNC_ERROR_NETWORK_FAILURE, "Failed to resolve server address");
                return false;
            }
            
            server_addr = *(sockaddr_in*)result->ai_addr;
            freeaddrinfo(result);
        }
        
        if (::connect(socket_fd_, (sockaddr*)&server_addr, sizeof(server_addr)) == SOCKET_ERROR) {
            closesocket(socket_fd_);
            socket_fd_ = INVALID_SOCKET;
            handleError(SYNC_ERROR_NETWORK_FAILURE, "Failed to connect to server");
            return false;
        }
        
        // Setup SSL if enabled
        if (config_.enable_encryption) {
            ssl_ = SSL_new(ssl_ctx_);
            SSL_set_fd(ssl_, socket_fd_);
            
            if (SSL_connect(ssl_) <= 0) {
                SSL_free(ssl_);
                ssl_ = nullptr;
                closesocket(socket_fd_);
                socket_fd_ = INVALID_SOCKET;
                handleError(SYNC_ERROR_NETWORK_FAILURE, "SSL connection failed");
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
        status_ = SYNC_STATUS_IDLE;
        notifyStatusChange();
        
        std::cout << "Connected to sync server" << std::endl;
        return true;
    }
    
    void disconnect() {
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
        
        status_ = SYNC_STATUS_OFFLINE;
        notifyStatusChange();
        
        std::cout << "Disconnected from sync server" << std::endl;
    }
    
    bool syncAll() {
        if (!is_connected_) {
            if (!connect()) {
                return false;
            }
        }
        
        std::lock_guard<std::mutex> lock(mutex_);
        
        status_ = SYNC_STATUS_SYNCING;
        notifyStatusChange();
        
        bool success = true;
        
        // Sync each collection
        for (auto& collection : collections_) {
            if (!syncCollection(collection.data_type)) {
                success = false;
            }
        }
        
        status_ = success ? SYNC_STATUS_COMPLETED : SYNC_STATUS_ERROR;
        notifyStatusChange();
        
        if (complete_callback_) {
            complete_callback_(synced_items_, failed_items_);
        }
        
        return success;
    }
    
    bool syncCollection(SyncDataType type) {
        // Get local items for this collection
        SyncItem* items = nullptr;
        uint32_t count = 0;
        
        if (list_items_ && !list_items_(type, &items, &count)) {
            return false;
        }
        
        // Send items in batches
        uint32_t batch_size = config_.max_batch_size;
        uint32_t total_batches = (count + batch_size - 1) / batch_size;
        
        for (uint32_t batch = 0; batch < total_batches; batch++) {
            uint32_t start_idx = batch * batch_size;
            uint32_t end_idx = std::min(start_idx + batch_size, count);
            
            if (!sendBatch(type, &items[start_idx], end_idx - start_idx, batch, total_batches)) {
                if (items) free(items);
                return false;
            }
        }
        
        if (items) {
            free(items);
        }
        
        return true;
    }
    
    bool addItem(const SyncData* data) {
        if (!data) return false;
        
        std::lock_guard<std::mutex> lock(mutex_);
        
        // Store locally
        if (store_item_ && !store_item_(data)) {
            return false;
        }
        
        // Add to local cache
        local_cache_[data->item.sync_id] = *data;
        
        // Mark collection as dirty
        markCollectionDirty(data->item.data_type);
        
        // Sync immediately if auto-sync is enabled and connected
        if (config_.auto_sync_enabled && is_connected_) {
            condition_.notify_one();
        }
        
        return true;
    }
    
    bool updateItem(const SyncData* data) {
        if (!data) return false;
        
        std::lock_guard<std::mutex> lock(mutex_);
        
        // Update locally
        if (store_item_ && !store_item_(data)) {
            return false;
        }
        
        // Update local cache
        local_cache_[data->item.sync_id] = *data;
        
        // Mark collection as dirty
        markCollectionDirty(data->item.data_type);
        
        // Sync immediately if auto-sync is enabled and connected
        if (config_.auto_sync_enabled && is_connected_) {
            condition_.notify_one();
        }
        
        return true;
    }
    
    bool deleteItem(const char* sync_id) {
        if (!sync_id) return false;
        
        std::lock_guard<std::mutex> lock(mutex_);
        
        // Delete locally
        if (delete_item_ && !delete_item_(sync_id)) {
            return false;
        }
        
        // Remove from local cache
        auto it = local_cache_.find(sync_id);
        if (it != local_cache_.end()) {
            SyncDataType type = it->second.item.data_type;
            local_cache_.erase(it);
            
            // Mark collection as dirty
            markCollectionDirty(type);
        }
        
        // Sync immediately if auto-sync is enabled and connected
        if (config_.auto_sync_enabled && is_connected_) {
            condition_.notify_one();
        }
        
        return true;
    }
    
    bool getItem(const char* sync_id, SyncData* data) {
        if (!sync_id || !data) return false;
        
        std::lock_guard<std::mutex> lock(mutex_);
        
        // Check local cache first
        auto it = local_cache_.find(sync_id);
        if (it != local_cache_.end()) {
            *data = it->second;
            return true;
        }
        
        // Try storage interface
        if (retrieve_item_) {
            return retrieve_item_(sync_id, data);
        }
        
        return false;
    }
    
    // Getters
    SyncStatus getStatus() const { return status_; }
    bool isConnected() const { return is_connected_; }
    float getProgress() const {
        if (pending_items_ == 0) return 1.0f;
        return (float)synced_items_ / (synced_items_ + pending_items_);
    }
    
    void getStats(uint32_t* synced, uint32_t* pending, uint32_t* failed) const {
        if (synced) *synced = synced_items_;
        if (pending) *pending = pending_items_;
        if (failed) *failed = failed_items_;
    }
    
    // Callback setters
    void setStatusCallback(std::function<void(SyncStatus, float)> callback) {
        status_callback_ = callback;
    }
    
    void setDataCallback(std::function<void(const SyncData*, SyncOperation)> callback) {
        data_callback_ = callback;
    }
    
    void setConflictCallback(std::function<void(const SyncConflict*)> callback) {
        conflict_callback_ = callback;
    }
    
    void setErrorCallback(std::function<void(SyncError, const char*)> callback) {
        error_callback_ = callback;
    }
    
    void setCompleteCallback(std::function<void(uint32_t, uint32_t)> callback) {
        complete_callback_ = callback;
    }
    
    // Storage interface setters
    void setStorageInterface(
        std::function<bool(const SyncData*)> store_item,
        std::function<bool(const char*, SyncData*)> retrieve_item,
        std::function<bool(const char*)> delete_item,
        std::function<bool(SyncDataType, SyncItem**, uint32_t*)> list_items,
        std::function<bool(const SyncCollection*)> update_collection) {
        
        store_item_ = store_item;
        retrieve_item_ = retrieve_item;
        delete_item_ = delete_item;
        list_items_ = list_items;
        update_collection_ = update_collection;
    }

private:
    bool initializeSSL() {
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
    
    bool performHandshake() {
        SyncHandshakeRequest request;
        memset(&request, 0, sizeof(request));
        
        strncpy_s(request.device_id, config_.device_id, MAX_SYNC_ID_LENGTH - 1);
        strncpy_s(request.device_name, "Windows Desktop", MAX_SYNC_ID_LENGTH - 1);
        request.protocol_version = DATA_SYNC_PROTOCOL_VERSION;
        request.supported_data_types = 0xFFFFFFFF; // Support all types
        request.supports_encryption = config_.enable_encryption;
        request.supports_compression = config_.enable_compression;
        request.max_batch_size = config_.max_batch_size;
        
        SyncHeader header;
        memset(&header, 0, sizeof(header));
        header.magic = DATA_SYNC_MAGIC;
        header.version = DATA_SYNC_PROTOCOL_VERSION;
        header.message_type = MSG_TYPE_SYNC_HANDSHAKE;
        header.message_id = generateMessageId();
        header.session_id = 0;
        header.data_length = sizeof(request);
        header.timestamp = getCurrentTimestamp();
        
        if (!sendMessage(&header, &request)) {
            handleError(SYNC_ERROR_PROTOCOL_ERROR, "Failed to send handshake request");
            return false;
        }
        
        // Receive response
        SyncHeader response_header;
        void* response_data = nullptr;
        
        if (!receiveMessage(&response_header, &response_data)) {
            handleError(SYNC_ERROR_PROTOCOL_ERROR, "Failed to receive handshake response");
            return false;
        }
        
        if (response_header.message_type != MSG_TYPE_SYNC_HANDSHAKE) {
            free(response_data);
            handleError(SYNC_ERROR_PROTOCOL_ERROR, "Invalid handshake response");
            return false;
        }
        
        SyncHandshakeResponse* response = (SyncHandshakeResponse*)response_data;
        
        if (!response->handshake_accepted) {
            free(response_data);
            handleError(response->error_code, "Handshake rejected");
            return false;
        }
        
        // Update configuration based on server capabilities
        config_.max_batch_size = std::min(config_.max_batch_size, response->max_batch_size);
        
        free(response_data);
        return true;
    }
    
    bool authenticate() {
        status_ = SYNC_STATUS_AUTHENTICATING;
        notifyStatusChange();
        
        SyncAuthRequest request;
        memset(&request, 0, sizeof(request));
        
        strncpy_s(request.user_id, config_.user_id, MAX_SYNC_ID_LENGTH - 1);
        strncpy_s(request.auth_token, config_.auth_token, sizeof(request.auth_token) - 1);
        request.timestamp = getCurrentTimestamp();
        
        // Generate device signature (simplified)
        std::string signature_data = std::string(config_.device_id) + std::to_string(request.timestamp);
        strncpy_s(request.device_signature, signature_data.c_str(), sizeof(request.device_signature) - 1);
        
        SyncHeader header;
        memset(&header, 0, sizeof(header));
        header.magic = DATA_SYNC_MAGIC;
        header.version = DATA_SYNC_PROTOCOL_VERSION;
        header.message_type = MSG_TYPE_SYNC_AUTH;
        header.message_id = generateMessageId();
        header.session_id = 0;
        header.data_length = sizeof(request);
        header.timestamp = getCurrentTimestamp();
        
        if (!sendMessage(&header, &request)) {
            handleError(SYNC_ERROR_AUTH_FAILED, "Failed to send auth request");
            return false;
        }
        
        // Receive response
        SyncHeader response_header;
        void* response_data = nullptr;
        
        if (!receiveMessage(&response_header, &response_data)) {
            handleError(SYNC_ERROR_AUTH_FAILED, "Failed to receive auth response");
            return false;
        }
        
        if (response_header.message_type != MSG_TYPE_SYNC_AUTH) {
            free(response_data);
            handleError(SYNC_ERROR_PROTOCOL_ERROR, "Invalid auth response");
            return false;
        }
        
        SyncAuthResponse* response = (SyncAuthResponse*)response_data;
        
        if (!response->auth_success) {
            free(response_data);
            handleError(response->error_code, "Authentication failed");
            return false;
        }
        
        // Store session info
        session_id_ = response_header.session_id;
        session_token_ = response->session_token;
        
        free(response_data);
        return true;
    }
    
    bool sendBatch(SyncDataType type, SyncItem* items, uint32_t count, uint32_t batch_num, uint32_t total_batches) {
        SyncBatchHeader batch_header;
        memset(&batch_header, 0, sizeof(batch_header));
        batch_header.batch_id = generateMessageId();
        batch_header.item_count = count;
        batch_header.total_batches = total_batches;
        batch_header.current_batch = batch_num;
        batch_header.data_type = type;
        batch_header.is_final_batch = (batch_num == total_batches - 1);
        
        // Calculate total data size
        uint32_t total_size = sizeof(batch_header) + count * sizeof(SyncItem);
        for (uint32_t i = 0; i < count; i++) {
            total_size += items[i].data_length + items[i].metadata_length;
        }
        
        // Allocate buffer
        std::vector<uint8_t> buffer(total_size);
        uint8_t* ptr = buffer.data();
        
        // Copy batch header
        memcpy(ptr, &batch_header, sizeof(batch_header));
        ptr += sizeof(batch_header);
        
        // Copy items and data
        for (uint32_t i = 0; i < count; i++) {
            memcpy(ptr, &items[i], sizeof(SyncItem));
            ptr += sizeof(SyncItem);
            
            // Copy item data
            if (items[i].data_length > 0) {
                SyncData sync_data;
                if (getItem(items[i].sync_id, &sync_data)) {
                    memcpy(ptr, sync_data.data, items[i].data_length);
                    ptr += items[i].data_length;
                    
                    if (items[i].metadata_length > 0) {
                        memcpy(ptr, sync_data.metadata, items[i].metadata_length);
                        ptr += items[i].metadata_length;
                    }
                }
            }
        }
        
        // Send batch
        SyncHeader header;
        memset(&header, 0, sizeof(header));
        header.magic = DATA_SYNC_MAGIC;
        header.version = DATA_SYNC_PROTOCOL_VERSION;
        header.message_type = MSG_TYPE_SYNC_DATA;
        header.message_id = generateMessageId();
        header.session_id = session_id_;
        header.data_length = total_size;
        header.timestamp = getCurrentTimestamp();
        
        if (!sendMessage(&header, buffer.data())) {
            return false;
        }
        
        // Wait for acknowledgment
        SyncHeader ack_header;
        void* ack_data = nullptr;
        
        if (!receiveMessage(&ack_header, &ack_data)) {
            return false;
        }
        
        if (ack_header.message_type != MSG_TYPE_SYNC_ACK) {
            free(ack_data);
            return false;
        }
        
        SyncBatchAck* ack = (SyncBatchAck*)ack_data;
        
        synced_items_ += ack->processed_items;
        failed_items_ += ack->failed_items;
        
        if (ack->conflict_count > 0) {
            // Handle conflicts
            // TODO: Implement conflict handling
        }
        
        free(ack_data);
        return ack->batch_complete;
    }
    
    bool sendMessage(const SyncHeader* header, const void* data) {
        // Calculate checksum
        SyncHeader mutable_header = *header;
        if (data && header->data_length > 0) {
            mutable_header.checksum = calculateChecksum(data, header->data_length);
        }
        
        // Send header
        int bytes_sent;
        if (ssl_) {
            bytes_sent = SSL_write(ssl_, &mutable_header, sizeof(SyncHeader));
        } else {
            bytes_sent = send(socket_fd_, (char*)&mutable_header, sizeof(SyncHeader), 0);
        }
        
        if (bytes_sent != sizeof(SyncHeader)) {
            return false;
        }
        
        // Send data if present
        if (data && header->data_length > 0) {
            if (ssl_) {
                bytes_sent = SSL_write(ssl_, data, header->data_length);
            } else {
                bytes_sent = send(socket_fd_, (char*)data, header->data_length, 0);
            }
            
            if (bytes_sent != header->data_length) {
                return false;
            }
        }
        
        return true;
    }
    
    bool receiveMessage(SyncHeader* header, void** data) {
        // Receive header
        int bytes_received;
        if (ssl_) {
            bytes_received = SSL_read(ssl_, header, sizeof(SyncHeader));
        } else {
            bytes_received = recv(socket_fd_, (char*)header, sizeof(SyncHeader), 0);
        }
        
        if (bytes_received != sizeof(SyncHeader)) {
            return false;
        }
        
        // Validate header
        if (header->magic != DATA_SYNC_MAGIC || header->version != DATA_SYNC_PROTOCOL_VERSION) {
            return false;
        }
        
        // Receive data if present
        *data = nullptr;
        if (header->data_length > 0) {
            *data = malloc(header->data_length);
            if (!*data) {
                return false;
            }
            
            if (ssl_) {
                bytes_received = SSL_read(ssl_, *data, header->data_length);
            } else {
                bytes_received = recv(socket_fd_, (char*)*data, header->data_length, 0);
            }
            
            if (bytes_received != header->data_length) {
                free(*data);
                *data = nullptr;
                return false;
            }
            
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
    
    void syncThreadFunc() {
        while (!shutdown_requested_) {
            std::unique_lock<std::mutex> lock(mutex_);
            
            // Wait for sync trigger or timeout
            condition_.wait_for(lock, std::chrono::milliseconds(config_.sync_interval), 
                              [this] { return shutdown_requested_ || (config_.auto_sync_enabled && is_connected_); });
            
            if (shutdown_requested_) {
                break;
            }
            
            if (config_.auto_sync_enabled && is_connected_) {
                lock.unlock();
                syncAll();
                lock.lock();
            }
        }
    }
    
    void heartbeatThreadFunc() {
        while (!shutdown_requested_) {
            std::this_thread::sleep_for(std::chrono::milliseconds(SYNC_HEARTBEAT_INTERVAL));
            
            if (shutdown_requested_) {
                break;
            }
            
            if (is_connected_) {
                sendHeartbeat();
            }
        }
    }
    
    void sendHeartbeat() {
        SyncHeader header;
        memset(&header, 0, sizeof(header));
        header.magic = DATA_SYNC_MAGIC;
        header.version = DATA_SYNC_PROTOCOL_VERSION;
        header.message_type = MSG_TYPE_SYNC_HEARTBEAT;
        header.message_id = generateMessageId();
        header.session_id = session_id_;
        header.data_length = 0;
        header.timestamp = getCurrentTimestamp();
        
        if (!sendMessage(&header, nullptr)) {
            // Heartbeat failed, disconnect
            disconnect();
        }
    }
    
    void loadCollections() {
        // Load collections from storage
        std::string collections_file = storage_path_ + "\\collections.json";
        std::ifstream file(collections_file);
        
        if (file.is_open()) {
            Json::Value root;
            file >> root;
            
            collections_.clear();
            for (const auto& item : root["collections"]) {
                SyncCollection collection;
                memset(&collection, 0, sizeof(collection));
                
                strncpy_s(collection.collection_id, item["id"].asString().c_str(), MAX_SYNC_ID_LENGTH - 1);
                collection.data_type = (SyncDataType)item["type"].asInt();
                collection.item_count = item["count"].asUInt();
                collection.last_sync_timestamp = item["last_sync"].asUInt64();
                collection.version = item["version"].asUInt64();
                collection.is_dirty = item["dirty"].asBool();
                
                collections_.push_back(collection);
            }
        }
    }
    
    void saveCollections() {
        Json::Value root;
        Json::Value collections_array(Json::arrayValue);
        
        for (const auto& collection : collections_) {
            Json::Value item;
            item["id"] = collection.collection_id;
            item["type"] = collection.data_type;
            item["count"] = collection.item_count;
            item["last_sync"] = collection.last_sync_timestamp;
            item["version"] = collection.version;
            item["dirty"] = collection.is_dirty;
            
            collections_array.append(item);
        }
        
        root["collections"] = collections_array;
        
        std::string collections_file = storage_path_ + "\\collections.json";
        std::ofstream file(collections_file);
        file << root;
    }
    
    void markCollectionDirty(SyncDataType type) {
        for (auto& collection : collections_) {
            if (collection.data_type == type) {
                collection.is_dirty = true;
                break;
            }
        }
        saveCollections();
    }
    
    void notifyStatusChange() {
        if (status_callback_) {
            status_callback_(status_, getProgress());
        }
    }
    
    void handleError(SyncError error, const char* message) {
        status_ = SYNC_STATUS_ERROR;
        
        if (error_callback_) {
            error_callback_(error, message);
        }
        
        std::cout << "Sync error: " << message << std::endl;
    }
    
    uint32_t generateMessageId() {
        static uint32_t counter = 0;
        return ++counter;
    }
    
    uint64_t getCurrentTimestamp() {
        return std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()).count();
    }
    
    uint32_t calculateChecksum(const void* data, uint32_t length) {
        const uint8_t* bytes = (const uint8_t*)data;
        uint32_t checksum = 0;
        
        for (uint32_t i = 0; i < length; i++) {
            checksum = (checksum << 1) ^ bytes[i];
        }
        
        return checksum;
    }
};

// MARK: - C API Implementation

extern "C" {

DataSyncManager* data_sync_manager_create(const SyncConfiguration* config) {
    try {
        return reinterpret_cast<DataSyncManager*>(new WindowsDataSyncManager(config));
    } catch (...) {
        return nullptr;
    }
}

void data_sync_manager_destroy(DataSyncManager* manager) {
    if (manager) {
        delete reinterpret_cast<WindowsDataSyncManager*>(manager);
    }
}

bool data_sync_manager_start(DataSyncManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->start();
}

void data_sync_manager_stop(DataSyncManager* manager) {
    if (manager) {
        reinterpret_cast<WindowsDataSyncManager*>(manager)->stop();
    }
}

bool data_sync_manager_connect(DataSyncManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->connect();
}

void data_sync_manager_disconnect(DataSyncManager* manager) {
    if (manager) {
        reinterpret_cast<WindowsDataSyncManager*>(manager)->disconnect();
    }
}

bool data_sync_manager_is_connected(const DataSyncManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<const WindowsDataSyncManager*>(manager)->isConnected();
}

bool data_sync_manager_sync_all(DataSyncManager* manager) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->syncAll();
}

bool data_sync_manager_sync_collection(DataSyncManager* manager, SyncDataType type) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->syncCollection(type);
}

bool data_sync_manager_add_item(DataSyncManager* manager, const SyncData* data) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->addItem(data);
}

bool data_sync_manager_update_item(DataSyncManager* manager, const SyncData* data) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->updateItem(data);
}

bool data_sync_manager_delete_item(DataSyncManager* manager, const char* sync_id) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->deleteItem(sync_id);
}

bool data_sync_manager_get_item(DataSyncManager* manager, const char* sync_id, SyncData* data) {
    if (!manager) return false;
    return reinterpret_cast<WindowsDataSyncManager*>(manager)->getItem(sync_id, data);
}

SyncStatus data_sync_manager_get_status(const DataSyncManager* manager) {
    if (!manager) return SYNC_STATUS_ERROR;
    return reinterpret_cast<const WindowsDataSyncManager*>(manager)->getStatus();
}

float data_sync_manager_get_progress(const DataSyncManager* manager) {
    if (!manager) return 0.0f;
    return reinterpret_cast<const WindowsDataSyncManager*>(manager)->getProgress();
}

void data_sync_manager_get_stats(const DataSyncManager* manager, uint32_t* synced, uint32_t* pending, uint32_t* failed) {
    if (manager) {
        reinterpret_cast<const WindowsDataSyncManager*>(manager)->getStats(synced, pending, failed);
    }
}

// Utility functions
char* generate_sync_id(void) {
    static uint32_t counter = 0;
    char* id = (char*)malloc(MAX_SYNC_ID_LENGTH);
    if (id) {
        snprintf(id, MAX_SYNC_ID_LENGTH, "SYNC_%08X_%08X", 
                (uint32_t)time(nullptr), ++counter);
    }
    return id;
}

uint64_t get_current_timestamp(void) {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
        std::chrono::system_clock::now().time_since_epoch()).count();
}

uint32_t calculate_data_checksum(const void* data, uint32_t length) {
    const uint8_t* bytes = (const uint8_t*)data;
    uint32_t checksum = 0;
    
    for (uint32_t i = 0; i < length; i++) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
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

} // extern "C"