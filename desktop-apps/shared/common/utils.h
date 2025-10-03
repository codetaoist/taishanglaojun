#ifndef UTILS_H
#define UTILS_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

// 平台检测宏
#ifdef _WIN32
    #define PLATFORM_WINDOWS
    #include <windows.h>
#elif defined(__APPLE__)
    #define PLATFORM_MACOS
    #include <mach/mach_time.h>
#elif defined(__linux__)
    #define PLATFORM_LINUX
    #include <unistd.h>
#endif

// 字符串工具函数
char* string_duplicate(const char* str);
bool string_equals(const char* str1, const char* str2);
bool string_starts_with(const char* str, const char* prefix);
bool string_ends_with(const char* str, const char* suffix);
char* string_trim(char* str);
void string_to_lower(char* str);
void string_to_upper(char* str);

// 文件路径工具
char* path_join(const char* path1, const char* path2);
char* path_get_directory(const char* filepath);
char* path_get_filename(const char* filepath);
char* path_get_extension(const char* filepath);
bool path_exists(const char* path);
bool path_is_directory(const char* path);
bool path_is_file(const char* path);
bool create_directory(const char* path);
bool create_directories(const char* path);

// 文件操作
bool file_exists(const char* filename);
uint64_t file_size(const char* filename);
bool file_copy(const char* src, const char* dst);
bool file_move(const char* src, const char* dst);
bool file_delete(const char* filename);
char* file_read_all_text(const char* filename);
bool file_write_all_text(const char* filename, const char* content);

// 时间工具
uint64_t get_current_timestamp_ms(void);
uint64_t get_current_timestamp_us(void);
char* format_timestamp(uint64_t timestamp, const char* format);
bool parse_timestamp(const char* time_str, const char* format, uint64_t* timestamp);

// 内存管理
void* safe_malloc(size_t size);
void* safe_calloc(size_t count, size_t size);
void* safe_realloc(void* ptr, size_t size);
void safe_free(void* ptr);

// 加密和哈希
bool calculate_md5(const void* data, size_t size, char* md5_str);
bool calculate_sha256(const void* data, size_t size, char* sha256_str);
bool calculate_file_md5(const char* filename, char* md5_str);
bool calculate_file_sha256(const char* filename, char* sha256_str);

// Base64编码/解码
char* base64_encode(const uint8_t* data, size_t size);
uint8_t* base64_decode(const char* encoded, size_t* decoded_size);

// JSON工具（简单的JSON解析和生成）
typedef struct json_object json_object_t;

json_object_t* json_parse(const char* json_str);
char* json_stringify(const json_object_t* obj);
void json_free(json_object_t* obj);

bool json_get_string(const json_object_t* obj, const char* key, char** value);
bool json_get_int(const json_object_t* obj, const char* key, int* value);
bool json_get_double(const json_object_t* obj, const char* key, double* value);
bool json_get_bool(const json_object_t* obj, const char* key, bool* value);

bool json_set_string(json_object_t* obj, const char* key, const char* value);
bool json_set_int(json_object_t* obj, const char* key, int value);
bool json_set_double(json_object_t* obj, const char* key, double value);
bool json_set_bool(json_object_t* obj, const char* key, bool value);

// 网络工具
typedef struct {
    char ip[16];
    int port;
} network_address_t;

bool resolve_hostname(const char* hostname, char* ip);
bool is_valid_ip(const char* ip);
bool is_valid_port(int port);
bool parse_address(const char* address, network_address_t* addr);

// 线程和同步（跨平台）
typedef struct thread_handle thread_handle_t;
typedef struct mutex_handle mutex_handle_t;
typedef struct condition_handle condition_handle_t;

typedef void* (*thread_func_t)(void* arg);

thread_handle_t* thread_create(thread_func_t func, void* arg);
bool thread_join(thread_handle_t* thread, void** result);
void thread_detach(thread_handle_t* thread);
void thread_sleep(uint32_t milliseconds);

mutex_handle_t* mutex_create(void);
void mutex_destroy(mutex_handle_t* mutex);
void mutex_lock(mutex_handle_t* mutex);
bool mutex_trylock(mutex_handle_t* mutex);
void mutex_unlock(mutex_handle_t* mutex);

condition_handle_t* condition_create(void);
void condition_destroy(condition_handle_t* cond);
void condition_wait(condition_handle_t* cond, mutex_handle_t* mutex);
bool condition_timedwait(condition_handle_t* cond, mutex_handle_t* mutex, uint32_t timeout_ms);
void condition_signal(condition_handle_t* cond);
void condition_broadcast(condition_handle_t* cond);

// 配置管理
typedef struct config_handle config_handle_t;

config_handle_t* config_load(const char* filename);
void config_save(config_handle_t* config, const char* filename);
void config_free(config_handle_t* config);

bool config_get_string(config_handle_t* config, const char* section, const char* key, char** value);
bool config_get_int(config_handle_t* config, const char* section, const char* key, int* value);
bool config_get_double(config_handle_t* config, const char* section, const char* key, double* value);
bool config_get_bool(config_handle_t* config, const char* section, const char* key, bool* value);

bool config_set_string(config_handle_t* config, const char* section, const char* key, const char* value);
bool config_set_int(config_handle_t* config, const char* section, const char* key, int value);
bool config_set_double(config_handle_t* config, const char* section, const char* key, double value);
bool config_set_bool(config_handle_t* config, const char* section, const char* key, bool value);

// 日志系统
typedef enum {
    LOG_LEVEL_DEBUG = 0,
    LOG_LEVEL_INFO = 1,
    LOG_LEVEL_WARN = 2,
    LOG_LEVEL_ERROR = 3,
    LOG_LEVEL_FATAL = 4
} log_level_t;

void log_init(const char* filename, log_level_t level);
void log_cleanup(void);
void log_write(log_level_t level, const char* format, ...);

#define LOG_DEBUG(fmt, ...) log_write(LOG_LEVEL_DEBUG, fmt, ##__VA_ARGS__)
#define LOG_INFO(fmt, ...)  log_write(LOG_LEVEL_INFO, fmt, ##__VA_ARGS__)
#define LOG_WARN(fmt, ...)  log_write(LOG_LEVEL_WARN, fmt, ##__VA_ARGS__)
#define LOG_ERROR(fmt, ...) log_write(LOG_LEVEL_ERROR, fmt, ##__VA_ARGS__)
#define LOG_FATAL(fmt, ...) log_write(LOG_LEVEL_FATAL, fmt, ##__VA_ARGS__)

// 错误处理
typedef struct {
    int code;
    char message[256];
    char context[256];
} error_info_t;

void set_last_error(int code, const char* message, const char* context);
error_info_t get_last_error(void);
void clear_last_error(void);

// 系统信息
typedef struct {
    char os_name[64];
    char os_version[64];
    char arch[32];
    uint64_t total_memory;
    uint64_t available_memory;
    int cpu_count;
} system_info_t;

bool get_system_info(system_info_t* info);
char* get_app_data_directory(void);
char* get_temp_directory(void);

// UUID生成
void generate_uuid(char* uuid_str);

// 数据压缩
uint8_t* compress_data(const uint8_t* data, size_t size, size_t* compressed_size);
uint8_t* decompress_data(const uint8_t* compressed_data, size_t compressed_size, size_t* decompressed_size);

#ifdef __cplusplus
}
#endif

#endif // UTILS_H