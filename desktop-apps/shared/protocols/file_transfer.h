#ifndef FILE_TRANSFER_H
#define FILE_TRANSFER_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

// 文件传输协议版本
#define FILE_TRANSFER_PROTOCOL_VERSION 1

// 文件传输端口
#define FILE_TRANSFER_DEFAULT_PORT 8888
#define FILE_TRANSFER_DISCOVERY_PORT 8889

// 文件传输限制
#define MAX_FILE_NAME_LENGTH 256
#define MAX_FILE_PATH_LENGTH 1024
#define MAX_DEVICE_NAME_LENGTH 64
#define MAX_DEVICE_ID_LENGTH 32
#define MAX_TRANSFER_SESSIONS 16
#define MAX_CHUNK_SIZE (1024 * 1024) // 1MB
#define MIN_CHUNK_SIZE (4 * 1024)    // 4KB
#define DEFAULT_CHUNK_SIZE (64 * 1024) // 64KB

// 文件传输消息类型
typedef enum {
    FT_MSG_DISCOVERY_REQUEST = 0x01,    // 设备发现请求
    FT_MSG_DISCOVERY_RESPONSE = 0x02,   // 设备发现响应
    FT_MSG_CONNECT_REQUEST = 0x03,      // 连接请求
    FT_MSG_CONNECT_RESPONSE = 0x04,     // 连接响应
    FT_MSG_AUTH_REQUEST = 0x05,         // 认证请求
    FT_MSG_AUTH_RESPONSE = 0x06,        // 认证响应
    FT_MSG_FILE_INFO = 0x10,            // 文件信息
    FT_MSG_FILE_REQUEST = 0x11,         // 文件请求
    FT_MSG_FILE_RESPONSE = 0x12,        // 文件响应
    FT_MSG_FILE_CHUNK = 0x13,           // 文件数据块
    FT_MSG_FILE_ACK = 0x14,             // 文件确认
    FT_MSG_TRANSFER_START = 0x15,       // 传输开始
    FT_MSG_TRANSFER_PAUSE = 0x16,       // 传输暂停
    FT_MSG_TRANSFER_RESUME = 0x17,      // 传输恢复
    FT_MSG_TRANSFER_CANCEL = 0x18,      // 传输取消
    FT_MSG_TRANSFER_COMPLETE = 0x19,    // 传输完成
    FT_MSG_ERROR = 0x20,                // 错误消息
    FT_MSG_HEARTBEAT = 0x30,            // 心跳消息
    FT_MSG_DISCONNECT = 0x31            // 断开连接
} FileTransferMessageType;

// 文件传输状态
typedef enum {
    FT_STATUS_IDLE = 0,
    FT_STATUS_DISCOVERING,
    FT_STATUS_CONNECTING,
    FT_STATUS_AUTHENTICATING,
    FT_STATUS_CONNECTED,
    FT_STATUS_TRANSFERRING,
    FT_STATUS_PAUSED,
    FT_STATUS_COMPLETED,
    FT_STATUS_CANCELLED,
    FT_STATUS_ERROR,
    FT_STATUS_DISCONNECTED
} FileTransferStatus;

// 设备类型
typedef enum {
    DEVICE_TYPE_UNKNOWN = 0,
    DEVICE_TYPE_DESKTOP_WINDOWS,
    DEVICE_TYPE_DESKTOP_MACOS,
    DEVICE_TYPE_DESKTOP_LINUX,
    DEVICE_TYPE_MOBILE_ANDROID,
    DEVICE_TYPE_MOBILE_IOS,
    DEVICE_TYPE_WEB_BROWSER
} DeviceType;

// 传输方向
typedef enum {
    TRANSFER_DIRECTION_SEND = 0,
    TRANSFER_DIRECTION_RECEIVE
} TransferDirection;

// 错误代码
typedef enum {
    FT_ERROR_NONE = 0,
    FT_ERROR_NETWORK_FAILURE = 1,
    FT_ERROR_CONNECTION_TIMEOUT = 2,
    FT_ERROR_AUTH_FAILED = 3,
    FT_ERROR_FILE_NOT_FOUND = 4,
    FT_ERROR_FILE_ACCESS_DENIED = 5,
    FT_ERROR_INSUFFICIENT_SPACE = 6,
    FT_ERROR_TRANSFER_CANCELLED = 7,
    FT_ERROR_PROTOCOL_ERROR = 8,
    FT_ERROR_CHECKSUM_MISMATCH = 9,
    FT_ERROR_DEVICE_NOT_FOUND = 10,
    FT_ERROR_INVALID_REQUEST = 11,
    FT_ERROR_UNSUPPORTED_VERSION = 12
} FileTransferError;

// 文件传输消息头
typedef struct {
    uint32_t magic;           // 魔数 0x46545250 ("FTRP")
    uint16_t version;         // 协议版本
    uint16_t message_type;    // 消息类型
    uint32_t message_id;      // 消息ID
    uint32_t session_id;      // 会话ID
    uint32_t data_length;     // 数据长度
    uint32_t checksum;        // 校验和
    uint64_t timestamp;       // 时间戳
} __attribute__((packed)) FileTransferHeader;

// 设备信息
typedef struct {
    char device_id[MAX_DEVICE_ID_LENGTH];
    char device_name[MAX_DEVICE_NAME_LENGTH];
    DeviceType device_type;
    uint32_t ip_address;
    uint16_t port;
    uint64_t last_seen;
    bool is_trusted;
    bool supports_encryption;
    uint32_t max_chunk_size;
} DeviceInfo;

// 文件信息
typedef struct {
    char file_name[MAX_FILE_NAME_LENGTH];
    char file_path[MAX_FILE_PATH_LENGTH];
    uint64_t file_size;
    uint64_t modified_time;
    uint32_t file_hash;
    char mime_type[64];
    bool is_directory;
    uint32_t permissions;
} FileInfo;

// 文件传输会话
typedef struct {
    uint32_t session_id;
    char session_token[64];
    DeviceInfo remote_device;
    FileInfo file_info;
    TransferDirection direction;
    FileTransferStatus status;
    uint64_t bytes_transferred;
    uint64_t total_bytes;
    uint32_t chunk_size;
    uint64_t start_time;
    uint64_t last_activity_time;
    float progress_percentage;
    float transfer_speed; // bytes per second
    uint32_t estimated_time_remaining; // seconds
    FileTransferError last_error;
    void* user_data;
} FileTransferSession;

// 发现请求消息
typedef struct {
    char device_id[MAX_DEVICE_ID_LENGTH];
    char device_name[MAX_DEVICE_NAME_LENGTH];
    DeviceType device_type;
    uint16_t listen_port;
    bool supports_encryption;
    uint32_t max_chunk_size;
} DiscoveryRequest;

// 发现响应消息
typedef struct {
    char device_id[MAX_DEVICE_ID_LENGTH];
    char device_name[MAX_DEVICE_NAME_LENGTH];
    DeviceType device_type;
    uint16_t listen_port;
    bool supports_encryption;
    uint32_t max_chunk_size;
    bool accepts_connections;
} DiscoveryResponse;

// 连接请求消息
typedef struct {
    char device_id[MAX_DEVICE_ID_LENGTH];
    char device_name[MAX_DEVICE_NAME_LENGTH];
    DeviceType device_type;
    uint32_t protocol_version;
    bool request_encryption;
} ConnectRequest;

// 连接响应消息
typedef struct {
    bool connection_accepted;
    uint32_t session_id;
    char session_token[64];
    bool encryption_enabled;
    uint32_t max_chunk_size;
    FileTransferError error_code;
} ConnectResponse;

// 认证请求消息
typedef struct {
    char device_id[MAX_DEVICE_ID_LENGTH];
    char auth_token[128];
    uint64_t timestamp;
    char signature[256];
} AuthRequest;

// 认证响应消息
typedef struct {
    bool auth_success;
    char session_token[64];
    uint64_t session_timeout;
    FileTransferError error_code;
} AuthResponse;

// 文件请求消息
typedef struct {
    FileInfo file_info;
    uint32_t chunk_size;
    bool resume_transfer;
    uint64_t resume_offset;
} FileRequest;

// 文件响应消息
typedef struct {
    bool request_accepted;
    uint32_t transfer_id;
    uint64_t file_size;
    uint32_t chunk_size;
    FileTransferError error_code;
} FileResponse;

// 文件数据块消息
typedef struct {
    uint32_t transfer_id;
    uint64_t chunk_offset;
    uint32_t chunk_size;
    uint32_t chunk_checksum;
    bool is_last_chunk;
    // 数据跟在结构体后面
} FileChunk;

// 文件确认消息
typedef struct {
    uint32_t transfer_id;
    uint64_t chunk_offset;
    bool chunk_received;
    FileTransferError error_code;
} FileAck;

// 传输控制消息
typedef struct {
    uint32_t transfer_id;
    FileTransferStatus new_status;
    uint64_t resume_offset;
    FileTransferError error_code;
} TransferControl;

// 错误消息
typedef struct {
    FileTransferError error_code;
    char error_message[256];
    uint32_t related_session_id;
    uint32_t related_transfer_id;
} ErrorMessage;

// 心跳消息
typedef struct {
    uint64_t timestamp;
    uint32_t active_transfers;
    uint64_t total_bytes_sent;
    uint64_t total_bytes_received;
} HeartbeatMessage;

// 回调函数类型定义
typedef void (*FileTransferProgressCallback)(uint32_t session_id, uint64_t bytes_transferred, uint64_t total_bytes, float speed, void* user_data);
typedef void (*FileTransferCompleteCallback)(uint32_t session_id, bool success, FileTransferError error, void* user_data);
typedef void (*FileTransferErrorCallback)(uint32_t session_id, FileTransferError error, const char* error_message, void* user_data);
typedef void (*DeviceDiscoveredCallback)(const DeviceInfo* device, void* user_data);
typedef void (*DeviceConnectedCallback)(const DeviceInfo* device, uint32_t session_id, void* user_data);
typedef void (*DeviceDisconnectedCallback)(const DeviceInfo* device, uint32_t session_id, void* user_data);
typedef bool (*FileReceiveRequestCallback)(const DeviceInfo* sender, const FileInfo* file_info, void* user_data);

// 文件传输管理器
typedef struct {
    char local_device_id[MAX_DEVICE_ID_LENGTH];
    char local_device_name[MAX_DEVICE_NAME_LENGTH];
    DeviceType local_device_type;
    uint16_t listen_port;
    bool is_running;
    bool discovery_enabled;
    bool encryption_enabled;
    uint32_t max_chunk_size;
    
    // 已发现的设备
    DeviceInfo discovered_devices[32];
    int discovered_device_count;
    
    // 活动会话
    FileTransferSession active_sessions[MAX_TRANSFER_SESSIONS];
    int active_session_count;
    
    // 网络相关
    int listen_socket;
    int discovery_socket;
    
    // 回调函数
    FileTransferProgressCallback progress_callback;
    FileTransferCompleteCallback complete_callback;
    FileTransferErrorCallback error_callback;
    DeviceDiscoveredCallback device_discovered_callback;
    DeviceConnectedCallback device_connected_callback;
    DeviceDisconnectedCallback device_disconnected_callback;
    FileReceiveRequestCallback file_receive_request_callback;
    void* callback_user_data;
    
    // 线程同步
    void* mutex;
    void* discovery_thread;
    void* server_thread;
    bool should_exit;
} FileTransferManager;

// 函数声明

// 文件传输管理器
FileTransferManager* file_transfer_manager_create(const char* device_name, DeviceType device_type);
void file_transfer_manager_destroy(FileTransferManager* manager);
bool file_transfer_manager_start(FileTransferManager* manager, uint16_t port);
void file_transfer_manager_stop(FileTransferManager* manager);
void file_transfer_manager_update(FileTransferManager* manager);

// 设备发现
bool file_transfer_start_discovery(FileTransferManager* manager);
void file_transfer_stop_discovery(FileTransferManager* manager);
int file_transfer_get_discovered_devices(FileTransferManager* manager, DeviceInfo* devices, int max_devices);
DeviceInfo* file_transfer_find_device_by_id(FileTransferManager* manager, const char* device_id);

// 连接管理
uint32_t file_transfer_connect_to_device(FileTransferManager* manager, const DeviceInfo* device);
bool file_transfer_disconnect_from_device(FileTransferManager* manager, uint32_t session_id);
bool file_transfer_is_connected_to_device(FileTransferManager* manager, const char* device_id);

// 文件传输
uint32_t file_transfer_send_file(FileTransferManager* manager, uint32_t session_id, const char* file_path);
uint32_t file_transfer_send_files(FileTransferManager* manager, uint32_t session_id, const char** file_paths, int file_count);
bool file_transfer_receive_file(FileTransferManager* manager, uint32_t session_id, uint32_t transfer_id, const char* save_path);
bool file_transfer_pause_transfer(FileTransferManager* manager, uint32_t transfer_id);
bool file_transfer_resume_transfer(FileTransferManager* manager, uint32_t transfer_id);
bool file_transfer_cancel_transfer(FileTransferManager* manager, uint32_t transfer_id);

// 会话管理
FileTransferSession* file_transfer_get_session(FileTransferManager* manager, uint32_t session_id);
int file_transfer_get_active_sessions(FileTransferManager* manager, FileTransferSession* sessions, int max_sessions);
bool file_transfer_close_session(FileTransferManager* manager, uint32_t session_id);

// 回调设置
void file_transfer_set_progress_callback(FileTransferManager* manager, FileTransferProgressCallback callback, void* user_data);
void file_transfer_set_complete_callback(FileTransferManager* manager, FileTransferCompleteCallback callback, void* user_data);
void file_transfer_set_error_callback(FileTransferManager* manager, FileTransferErrorCallback callback, void* user_data);
void file_transfer_set_device_discovered_callback(FileTransferManager* manager, DeviceDiscoveredCallback callback, void* user_data);
void file_transfer_set_device_connected_callback(FileTransferManager* manager, DeviceConnectedCallback callback, void* user_data);
void file_transfer_set_device_disconnected_callback(FileTransferManager* manager, DeviceDisconnectedCallback callback, void* user_data);
void file_transfer_set_file_receive_request_callback(FileTransferManager* manager, FileReceiveRequestCallback callback, void* user_data);

// 消息处理
bool file_transfer_send_message(int socket, const FileTransferHeader* header, const void* data);
bool file_transfer_receive_message(int socket, FileTransferHeader* header, void** data);
uint32_t file_transfer_calculate_checksum(const void* data, size_t size);
bool file_transfer_verify_checksum(const void* data, size_t size, uint32_t expected_checksum);

// 工具函数
const char* file_transfer_status_to_string(FileTransferStatus status);
const char* file_transfer_error_to_string(FileTransferError error);
const char* device_type_to_string(DeviceType type);
bool file_transfer_is_valid_device_id(const char* device_id);
void file_transfer_generate_device_id(char* device_id, size_t size);
void file_transfer_generate_session_token(char* token, size_t size);
uint64_t file_transfer_get_current_time_ms(void);
bool file_transfer_create_directory(const char* path);
bool file_transfer_file_exists(const char* path);
uint64_t file_transfer_get_file_size(const char* path);
uint32_t file_transfer_calculate_file_hash(const char* file_path);

// 加密相关（如果启用）
bool file_transfer_encrypt_data(const void* input, size_t input_size, void** output, size_t* output_size, const char* key);
bool file_transfer_decrypt_data(const void* input, size_t input_size, void** output, size_t* output_size, const char* key);
void file_transfer_generate_encryption_key(char* key, size_t size);

// 网络工具
bool file_transfer_create_socket(int* socket, bool is_udp);
bool file_transfer_bind_socket(int socket, uint16_t port);
bool file_transfer_listen_socket(int socket, int backlog);
bool file_transfer_connect_socket(int socket, uint32_t ip_address, uint16_t port);
bool file_transfer_send_data(int socket, const void* data, size_t size);
bool file_transfer_receive_data(int socket, void* buffer, size_t size, size_t* received);
void file_transfer_close_socket(int socket);
uint32_t file_transfer_get_local_ip_address(void);
bool file_transfer_broadcast_message(int socket, uint16_t port, const void* data, size_t size);

// 常量定义
#define FILE_TRANSFER_MAGIC 0x46545250 // "FTRP"
#define FILE_TRANSFER_DISCOVERY_INTERVAL_MS 5000
#define FILE_TRANSFER_HEARTBEAT_INTERVAL_MS 30000
#define FILE_TRANSFER_CONNECTION_TIMEOUT_MS 10000
#define FILE_TRANSFER_TRANSFER_TIMEOUT_MS 60000
#define FILE_TRANSFER_MAX_RETRY_COUNT 3

#ifdef __cplusplus
}
#endif

#endif // FILE_TRANSFER_H