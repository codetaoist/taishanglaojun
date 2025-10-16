/**
 * @file utils.h
 * @brief TaishangLaojun Desktop Application Utilities Header
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains utility functions, macros, and definitions
 * for the TaishangLaojun desktop application on Linux.
 */

#ifndef TAISHANG_UTILS_H
#define TAISHANG_UTILS_H

#include "common.h"

#ifdef __cplusplus
extern "C" {
#endif

/* String utilities */
gchar *taishang_utils_string_trim(const gchar *str);
gchar *taishang_utils_string_trim_whitespace(const gchar *str);
gchar *taishang_utils_string_duplicate(const gchar *str);
gchar *taishang_utils_string_to_lower(const gchar *str);
gchar *taishang_utils_string_to_upper(const gchar *str);
gchar *taishang_utils_string_capitalize(const gchar *str);

gboolean taishang_utils_string_is_empty(const gchar *str);
gboolean taishang_utils_string_is_whitespace(const gchar *str);
gboolean taishang_utils_string_starts_with(const gchar *str, const gchar *prefix);
gboolean taishang_utils_string_ends_with(const gchar *str, const gchar *suffix);
gboolean taishang_utils_string_contains(const gchar *str, const gchar *substring);

gchar **taishang_utils_string_split(const gchar *str, const gchar *delimiter);
gchar *taishang_utils_string_join(gchar **strv, const gchar *separator);
gchar *taishang_utils_string_replace(const gchar *str, const gchar *old_str, const gchar *new_str);

gint taishang_utils_string_compare_case_insensitive(const gchar *str1, const gchar *str2);
gboolean taishang_utils_string_equal_case_insensitive(const gchar *str1, const gchar *str2);

/* File utilities */
gboolean taishang_utils_file_exists(const gchar *file_path);
gboolean taishang_utils_file_is_readable(const gchar *file_path);
gboolean taishang_utils_file_is_writable(const gchar *file_path);
gboolean taishang_utils_file_is_executable(const gchar *file_path);

gint64 taishang_utils_file_get_size(const gchar *file_path);
gint64 taishang_utils_file_get_modified_time(const gchar *file_path);

gchar *taishang_utils_file_read_contents(const gchar *file_path, GError **error);
gboolean taishang_utils_file_write_contents(const gchar *file_path, const gchar *contents, GError **error);

gboolean taishang_utils_file_copy(const gchar *src_path, const gchar *dest_path, GError **error);
gboolean taishang_utils_file_move(const gchar *src_path, const gchar *dest_path, GError **error);
gboolean taishang_utils_file_delete(const gchar *file_path, GError **error);

gchar *taishang_utils_file_get_mime_type(const gchar *file_path);
gchar *taishang_utils_file_get_extension(const gchar *file_path);
gchar *taishang_utils_file_get_basename(const gchar *file_path);
gchar *taishang_utils_file_get_dirname(const gchar *file_path);

/* Directory utilities */
gboolean taishang_utils_dir_exists(const gchar *dir_path);
gboolean taishang_utils_dir_create(const gchar *dir_path, gint mode, GError **error);
gboolean taishang_utils_dir_create_recursive(const gchar *dir_path, gint mode, GError **error);
gboolean taishang_utils_dir_remove(const gchar *dir_path, GError **error);
gboolean taishang_utils_dir_remove_recursive(const gchar *dir_path, GError **error);

gchar **taishang_utils_dir_list_files(const gchar *dir_path, GError **error);
gchar **taishang_utils_dir_list_directories(const gchar *dir_path, GError **error);
gchar **taishang_utils_dir_list_all(const gchar *dir_path, GError **error);

gboolean taishang_utils_dir_is_empty(const gchar *dir_path);
gint64 taishang_utils_dir_get_size(const gchar *dir_path);

/* Path utilities */
gchar *taishang_utils_path_join(const gchar *first_element, ...);
gchar *taishang_utils_path_join_array(gchar **path_elements);
gchar *taishang_utils_path_normalize(const gchar *path);
gchar *taishang_utils_path_get_absolute(const gchar *path);
gchar *taishang_utils_path_get_relative(const gchar *path, const gchar *base);

gboolean taishang_utils_path_is_absolute(const gchar *path);
gboolean taishang_utils_path_is_relative(const gchar *path);

gchar *taishang_utils_path_get_home_dir(void);
gchar *taishang_utils_path_get_config_dir(void);
gchar *taishang_utils_path_get_cache_dir(void);
gchar *taishang_utils_path_get_data_dir(void);
gchar *taishang_utils_path_get_temp_dir(void);

/* Time utilities */
gint64 taishang_utils_time_get_timestamp(void);
gint64 taishang_utils_time_get_timestamp_ms(void);
gchar *taishang_utils_time_format_timestamp(gint64 timestamp, const gchar *format);
gchar *taishang_utils_time_format_iso8601(gint64 timestamp);
gchar *taishang_utils_time_format_human_readable(gint64 timestamp);

gint64 taishang_utils_time_parse_iso8601(const gchar *time_str);
gint64 taishang_utils_time_parse_format(const gchar *time_str, const gchar *format);

gint64 taishang_utils_time_elapsed_since(gint64 start_time);
gchar *taishang_utils_time_elapsed_string(gint64 elapsed_time);

/* Hash utilities */
gchar *taishang_utils_hash_md5(const gchar *data, gsize length);
gchar *taishang_utils_hash_sha1(const gchar *data, gsize length);
gchar *taishang_utils_hash_sha256(const gchar *data, gsize length);
gchar *taishang_utils_hash_sha512(const gchar *data, gsize length);

gchar *taishang_utils_hash_file_md5(const gchar *file_path, GError **error);
gchar *taishang_utils_hash_file_sha256(const gchar *file_path, GError **error);

/* Encoding utilities */
gchar *taishang_utils_base64_encode(const guchar *data, gsize length);
guchar *taishang_utils_base64_decode(const gchar *encoded, gsize *out_length);

gchar *taishang_utils_url_encode(const gchar *str);
gchar *taishang_utils_url_decode(const gchar *str);

gchar *taishang_utils_html_escape(const gchar *str);
gchar *taishang_utils_html_unescape(const gchar *str);

/* Random utilities */
gint taishang_utils_random_int(gint min, gint max);
gdouble taishang_utils_random_double(gdouble min, gdouble max);
gchar *taishang_utils_random_string(gint length, const gchar *charset);
gchar *taishang_utils_random_uuid(void);

void taishang_utils_random_seed(guint32 seed);
void taishang_utils_random_seed_from_time(void);

/* Memory utilities */
gpointer taishang_utils_malloc_safe(gsize size);
gpointer taishang_utils_realloc_safe(gpointer ptr, gsize size);
void taishang_utils_free_safe(gpointer ptr);
void taishang_utils_memzero(gpointer ptr, gsize size);

gchar *taishang_utils_strdup_safe(const gchar *str);
gchar **taishang_utils_strdupv_safe(gchar **strv);

/* Memory pool for efficient allocation */
typedef struct _TaishangMemoryPool TaishangMemoryPool;

TaishangMemoryPool *taishang_utils_memory_pool_new(gsize block_size);
void taishang_utils_memory_pool_free(TaishangMemoryPool *pool);
gpointer taishang_utils_memory_pool_alloc(TaishangMemoryPool *pool, gsize size);
void taishang_utils_memory_pool_clear(TaishangMemoryPool *pool);

/* Logging utilities */
typedef enum {
    TAISHANG_LOG_LEVEL_ERROR = 0,
    TAISHANG_LOG_LEVEL_WARNING,
    TAISHANG_LOG_LEVEL_INFO,
    TAISHANG_LOG_LEVEL_DEBUG,
    TAISHANG_LOG_LEVEL_TRACE
} TaishangLogLevel;

void taishang_utils_log_init(const gchar *log_file, TaishangLogLevel level);
void taishang_utils_log_cleanup(void);
void taishang_utils_log_set_level(TaishangLogLevel level);
TaishangLogLevel taishang_utils_log_get_level(void);

void taishang_utils_log_message(TaishangLogLevel level, const gchar *domain, const gchar *format, ...) G_GNUC_PRINTF(3, 4);
void taishang_utils_log_error(const gchar *domain, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_utils_log_warning(const gchar *domain, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_utils_log_info(const gchar *domain, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_utils_log_debug(const gchar *domain, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_utils_log_trace(const gchar *domain, const gchar *format, ...) G_GNUC_PRINTF(2, 3);

/* Process utilities */
gint taishang_utils_process_get_pid(void);
gchar *taishang_utils_process_get_name(void);
gchar *taishang_utils_process_get_executable_path(void);

gboolean taishang_utils_process_is_running(gint pid);
gboolean taishang_utils_process_kill(gint pid, gint signal);

gint taishang_utils_process_execute(const gchar *command, gchar **argv, gchar **envp, 
                                   gchar **stdout_output, gchar **stderr_output, GError **error);
gboolean taishang_utils_process_execute_async(const gchar *command, gchar **argv, gchar **envp, 
                                             GPid *child_pid, GError **error);

/* System utilities */
gchar *taishang_utils_system_get_hostname(void);
gchar *taishang_utils_system_get_username(void);
gchar *taishang_utils_system_get_os_name(void);
gchar *taishang_utils_system_get_os_version(void);
gchar *taishang_utils_system_get_architecture(void);

gint64 taishang_utils_system_get_memory_total(void);
gint64 taishang_utils_system_get_memory_available(void);
gint64 taishang_utils_system_get_disk_space_total(const gchar *path);
gint64 taishang_utils_system_get_disk_space_free(const gchar *path);

gint taishang_utils_system_get_cpu_count(void);
gdouble taishang_utils_system_get_cpu_usage(void);

/* Network utilities */
gboolean taishang_utils_network_is_online(void);
gchar *taishang_utils_network_get_local_ip(void);
gchar *taishang_utils_network_get_public_ip(void);

gboolean taishang_utils_network_is_port_open(const gchar *host, gint port);
gboolean taishang_utils_network_download_file(const gchar *url, const gchar *dest_path, GError **error);

/* Validation utilities */
gboolean taishang_utils_validate_email(const gchar *email);
gboolean taishang_utils_validate_url(const gchar *url);
gboolean taishang_utils_validate_ip_address(const gchar *ip);
gboolean taishang_utils_validate_domain_name(const gchar *domain);
gboolean taishang_utils_validate_uuid(const gchar *uuid);

/* Configuration utilities */
gchar *taishang_utils_config_get_user_config_dir(const gchar *app_name);
gchar *taishang_utils_config_get_user_cache_dir(const gchar *app_name);
gchar *taishang_utils_config_get_user_data_dir(const gchar *app_name);

/* Desktop integration utilities */
gboolean taishang_utils_desktop_create_shortcut(const gchar *name, const gchar *exec_path, 
                                               const gchar *icon_path, const gchar *comment);
gboolean taishang_utils_desktop_remove_shortcut(const gchar *name);
gboolean taishang_utils_desktop_set_autostart(const gchar *name, const gchar *exec_path, gboolean enable);

/* Notification utilities */
gboolean taishang_utils_notification_show(const gchar *title, const gchar *message, 
                                         const gchar *icon, gint timeout);
gboolean taishang_utils_notification_is_supported(void);

/* Clipboard utilities */
gchar *taishang_utils_clipboard_get_text(void);
gboolean taishang_utils_clipboard_set_text(const gchar *text);
gboolean taishang_utils_clipboard_has_text(void);

/* Error handling utilities */
#define TAISHANG_UTILS_ERROR (taishang_utils_error_quark())
GQuark taishang_utils_error_quark(void);

typedef enum {
    TAISHANG_UTILS_ERROR_INVALID_ARGUMENT,
    TAISHANG_UTILS_ERROR_FILE_NOT_FOUND,
    TAISHANG_UTILS_ERROR_PERMISSION_DENIED,
    TAISHANG_UTILS_ERROR_OUT_OF_MEMORY,
    TAISHANG_UTILS_ERROR_NETWORK_ERROR,
    TAISHANG_UTILS_ERROR_TIMEOUT,
    TAISHANG_UTILS_ERROR_UNKNOWN
} TaishangUtilsError;

/* Macros for common operations */
#define TAISHANG_UTILS_SAFE_FREE(ptr) do { \
    if ((ptr) != NULL) { \
        g_free(ptr); \
        (ptr) = NULL; \
    } \
} while (0)

#define TAISHANG_UTILS_SAFE_UNREF(obj) do { \
    if (G_IS_OBJECT(obj)) { \
        g_object_unref(obj); \
        (obj) = NULL; \
    } \
} while (0)

#define TAISHANG_UTILS_RETURN_IF_FAIL(expr) do { \
    if (G_UNLIKELY(!(expr))) { \
        g_return_if_fail(expr); \
        return; \
    } \
} while (0)

#define TAISHANG_UTILS_RETURN_VAL_IF_FAIL(expr, val) do { \
    if (G_UNLIKELY(!(expr))) { \
        g_return_val_if_fail(expr, val); \
        return (val); \
    } \
} while (0)

/* Debug macros */
#ifdef DEBUG
#define TAISHANG_UTILS_DEBUG(format, ...) \
    taishang_utils_log_debug(G_LOG_DOMAIN, format, ##__VA_ARGS__)
#define TAISHANG_UTILS_TRACE(format, ...) \
    taishang_utils_log_trace(G_LOG_DOMAIN, format, ##__VA_ARGS__)
#else
#define TAISHANG_UTILS_DEBUG(format, ...) G_STMT_START { } G_STMT_END
#define TAISHANG_UTILS_TRACE(format, ...) G_STMT_START { } G_STMT_END
#endif

/* Performance measurement */
typedef struct _TaishangStopwatch TaishangStopwatch;

TaishangStopwatch *taishang_utils_stopwatch_new(void);
void taishang_utils_stopwatch_free(TaishangStopwatch *stopwatch);
void taishang_utils_stopwatch_start(TaishangStopwatch *stopwatch);
void taishang_utils_stopwatch_stop(TaishangStopwatch *stopwatch);
void taishang_utils_stopwatch_reset(TaishangStopwatch *stopwatch);
gdouble taishang_utils_stopwatch_elapsed(TaishangStopwatch *stopwatch);

#ifdef __cplusplus
}
#endif

#endif /* TAISHANG_UTILS_H */