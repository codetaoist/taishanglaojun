#ifndef COMMUNICATION_H
#define COMMUNICATION_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// 消息类型定义
typedef enum {
    MSG_TYPE_HEARTBEAT = 0x01,
    MSG_TYPE_AUTH = 0x02,
    MSG_TYPE_CHAT = 0x03,
    MSG_TYPE_FILE_TRANSFER = 0x04,
    MSG_TYPE_SYNC_REQUEST = 0x05,
    MSG_TYPE_SYNC_RESPONSE = 0x06,
    MSG_TYPE_PROJECT_UPDATE = 0x07,
    MSG_TYPE_NOTIFICATION = 0x08,
    MSG_TYPE_ERROR = 0xFF
} MessageType;

// 文件传输状态
typedef enum {
    FILE_TRANSFER_INIT = 0,
    FILE_TRANSFER_PROGRESS = 1,
    FILE_TRANSFER_COMPLETE = 2,
    FILE_TRANSFER_ERROR = 3,
    FILE_TRANSFER_CANCELLED = 4
} FileTransferStatus;

// 数据同步类型
typedef enum {
    SYNC_TYPE_CHAT_HISTORY = 1,
    SYNC_TYPE_FAVORITES = 2,
    SYNC_TYPE_PROJECT_DATA = 3,
    SYNC_TYPE_USER_SETTINGS = 4
} SyncType;

// 消息头结构
typedef struct {
    uint32_t magic;           // 魔数标识: 0x544C4A41 ("TLJA")
    uint16_t version;         // 协议版本
    uint16_t message_type;    // 消息类型
    uint32_t message_id;      // 消息ID
    uint32_t payload_size;    // 负载大小
    uint32_t checksum;        // 校验和
    uint64_t timestamp;       // 时间戳
} MessageHeader;

// 认证消息
typedef struct {
    char user_id[64];
    char token[256];
    char device_id[128];
    char platform[32];
} AuthMessage;

// 聊天消息
typedef struct {
    char conversation_id[64];
    char message_id[64];
    char user_id[64];
    char content[4096];
    uint64_t timestamp;
    uint8_t message_type;     // 0: text, 1: image, 2: file
} ChatMessage;

// 文件传输消息
typedef struct {
    char file_id[64];
    char filename[256];
    uint64_t file_size;
    uint64_t transferred_size;
    uint8_t status;           // FileTransferStatus
    char checksum[64];
    uint8_t chunk_data[8192]; // 文件块数据
} FileTransferMessage;

// 数据同步请求
typedef struct {
    uint8_t sync_type;        // SyncType
    uint64_t last_sync_time;
    char device_id[128];
    uint32_t batch_size;
} SyncRequest;

// 数据同步响应
typedef struct {
    uint8_t sync_type;
    uint32_t record_count;
    uint64_t sync_time;
    bool has_more;
    char data[16384];         // JSON格式的数据
} SyncResponse;

// 项目更新消息
typedef struct {
    char project_id[64];
    char update_type[32];     // "create", "update", "delete"
    char data[8192];          // JSON格式的项目数据
    uint64_t timestamp;
} ProjectUpdateMessage;

// 通知消息
typedef struct {
    char notification_id[64];
    char title[256];
    char content[1024];
    uint8_t priority;         // 0: low, 1: normal, 2: high, 3: urgent
    uint64_t timestamp;
    char action_url[512];
} NotificationMessage;

// 错误消息
typedef struct {
    uint32_t error_code;
    char error_message[512];
    char context[256];
} ErrorMessage;

// 通用消息结构
typedef struct {
    MessageHeader header;
    union {
        AuthMessage auth;
        ChatMessage chat;
        FileTransferMessage file_transfer;
        SyncRequest sync_request;
        SyncResponse sync_response;
        ProjectUpdateMessage project_update;
        NotificationMessage notification;
        ErrorMessage error;
    } payload;
} Message;

// 函数声明
bool validate_message(const Message* msg);
uint32_t calculate_checksum(const void* data, size_t size);
bool serialize_message(const Message* msg, uint8_t* buffer, size_t buffer_size, size_t* serialized_size);
bool deserialize_message(const uint8_t* buffer, size_t buffer_size, Message* msg);
void init_message_header(MessageHeader* header, MessageType type, uint32_t payload_size);

// 错误代码定义
#define ERROR_SUCCESS           0x0000
#define ERROR_INVALID_MESSAGE   0x0001
#define ERROR_AUTH_FAILED       0x0002
#define ERROR_FILE_NOT_FOUND    0x0003
#define ERROR_TRANSFER_FAILED   0x0004
#define ERROR_SYNC_FAILED       0x0005
#define ERROR_NETWORK_ERROR     0x0006
#define ERROR_INSUFFICIENT_SPACE 0x0007
#define ERROR_PERMISSION_DENIED 0x0008

// 常量定义
#define PROTOCOL_MAGIC          0x544C4A41  // "TLJA"
#define PROTOCOL_VERSION        0x0001
#define MAX_MESSAGE_SIZE        65536
#define HEARTBEAT_INTERVAL      30          // 秒
#define CONNECTION_TIMEOUT      60          // 秒
#define FILE_CHUNK_SIZE         8192        // 字节
#define MAX_RETRY_COUNT         3

#ifdef __cplusplus
}
#endif

#endif // COMMUNICATION_H