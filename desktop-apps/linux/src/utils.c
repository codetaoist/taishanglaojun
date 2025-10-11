/**
 * @file utils.c
 * @brief TaishangLaojun Desktop Application Utilities Implementation
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains utility function implementations for the
 * TaishangLaojun desktop application on Linux.
 */

#include "common.h"
#include "utils.h"
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/utsname.h>
#include <sys/sysinfo.h>
#include <dirent.h>
#include <unistd.h>
#include <pwd.h>
#include <grp.h>
#include <time.h>
#include <fcntl.h>
#include <errno.h>

/* Memory pool structure */
struct _TaishangMemoryPool {
    gsize block_size;
    gsize blocks_per_chunk;
    GSList *chunks;
    GSList *free_blocks;
    GMutex mutex;
    gsize total_allocated;
    gsize total_freed;
};

typedef struct {
    gpointer data;
    gsize size;
} MemoryChunk;

/* Static variables */
static GMutex log_mutex = G_MUTEX_INIT;
static GLogLevelFlags log_level = G_LOG_LEVEL_INFO;
static gchar *log_file_path = NULL;
static FILE *log_file = NULL;

/* Forward declarations */
static void taishang_utils_log_handler(const gchar *log_domain, GLogLevelFlags log_level_flags, const gchar *message, gpointer user_data);
static gchar *taishang_utils_get_timestamp_string(void);
static void memory_chunk_free(MemoryChunk *chunk);

/* String utilities implementation */

/**
 * taishang_utils_string_trim:
 * @str: String to trim
 * 
 * Trims whitespace from both ends of a string.
 * 
 * Returns: (transfer full): Trimmed string
 */
gchar *taishang_utils_string_trim(const gchar *str) {
    if (!str) {
        return NULL;
    }
    
    /* Skip leading whitespace */
    while (g_ascii_isspace(*str)) {
        str++;
    }
    
    if (*str == '\0') {
        return g_strdup("");
    }
    
    /* Find end of string */
    const gchar *end = str + strlen(str) - 1;
    
    /* Skip trailing whitespace */
    while (end > str && g_ascii_isspace(*end)) {
        end--;
    }
    
    /* Extract trimmed string */
    gsize len = end - str + 1;
    return g_strndup(str, len);
}

/**
 * taishang_utils_string_duplicate:
 * @str: String to duplicate
 * 
 * Safely duplicates a string.
 * 
 * Returns: (transfer full): Duplicated string
 */
gchar *taishang_utils_string_duplicate(const gchar *str) {
    return g_strdup(str);
}

/**
 * taishang_utils_string_to_lower:
 * @str: String to convert
 * 
 * Converts string to lowercase.
 * 
 * Returns: (transfer full): Lowercase string
 */
gchar *taishang_utils_string_to_lower(const gchar *str) {
    if (!str) {
        return NULL;
    }
    
    return g_ascii_strdown(str, -1);
}

/**
 * taishang_utils_string_to_upper:
 * @str: String to convert
 * 
 * Converts string to uppercase.
 * 
 * Returns: (transfer full): Uppercase string
 */
gchar *taishang_utils_string_to_upper(const gchar *str) {
    if (!str) {
        return NULL;
    }
    
    return g_ascii_strup(str, -1);
}

/**
 * taishang_utils_string_split:
 * @str: String to split
 * @delimiter: Delimiter character
 * 
 * Splits string by delimiter.
 * 
 * Returns: (transfer full) (array zero-terminated=1): Array of strings
 */
gchar **taishang_utils_string_split(const gchar *str, const gchar *delimiter) {
    if (!str || !delimiter) {
        return NULL;
    }
    
    return g_strsplit(str, delimiter, -1);
}

/* File utilities implementation */

/**
 * taishang_utils_file_write:
 * @file_path: Path to file
 * @content: Content to write
 * @error: Return location for error
 * 
 * Writes content to file.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_utils_file_write(const gchar *file_path, const gchar *content, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(file_path != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(content != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    return g_file_set_contents(file_path, content, -1, error);
}

/**
 * taishang_utils_file_exists:
 * @file_path: Path to file
 * 
 * Checks if file exists.
 * 
 * Returns: TRUE if file exists, FALSE otherwise
 */
gboolean taishang_utils_file_exists(const gchar *file_path) {
    if (!file_path) {
        return FALSE;
    }
    
    return g_file_test(file_path, G_FILE_TEST_EXISTS | G_FILE_TEST_IS_REGULAR);
}

/**
 * taishang_utils_file_read:
 * @file_path: Path to file
 * @error: Return location for error
 * 
 * Reads content from file.
 * 
 * Returns: (transfer full): File content
 */
gchar *taishang_utils_file_read(const gchar *file_path, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(file_path != NULL, NULL);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, NULL);
    
    gchar *content = NULL;
    if (!g_file_get_contents(file_path, &content, NULL, error)) {
        return NULL;
    }
    
    return content;
}

/**
 * taishang_utils_file_get_size:
 * @file_path: Path to file
 * 
 * Gets file size.
 * 
 * Returns: File size in bytes, -1 on error
 */
gint64 taishang_utils_file_get_size(const gchar *file_path) {
    if (!file_path) {
        return -1;
    }
    
    struct stat st;
    if (stat(file_path, &st) != 0) {
        return -1;
    }
    
    return st.st_size;
}

/**
 * taishang_utils_file_delete:
 * @file_path: Path to file
 * @error: Return location for error
 * 
 * Deletes a file.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_utils_file_delete(const gchar *file_path, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(file_path != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    if (g_unlink(file_path) != 0) {
        g_set_error(error, G_FILE_ERROR, g_file_error_from_errno(errno),
                   "Failed to delete file '%s': %s", file_path, g_strerror(errno));
        return FALSE;
    }
    
    return TRUE;
}

/* Directory utilities implementation */

/**
 * taishang_utils_create_directory:
 * @dir_path: Path to directory
 * @error: Return location for error
 * 
 * Creates a directory (and parent directories if needed).
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_utils_create_directory(const gchar *dir_path, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(dir_path != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    if (g_mkdir_with_parents(dir_path, 0755) != 0) {
        g_set_error(error, G_FILE_ERROR, g_file_error_from_errno(errno),
                   "Failed to create directory '%s': %s", dir_path, g_strerror(errno));
        return FALSE;
    }
    
    return TRUE;
}

/**
 * taishang_utils_directory_exists:
 * @dir_path: Path to directory
 * 
 * Checks if directory exists.
 * 
 * Returns: TRUE if directory exists, FALSE otherwise
 */
gboolean taishang_utils_directory_exists(const gchar *dir_path) {
    if (!dir_path) {
        return FALSE;
    }
    
    return g_file_test(dir_path, G_FILE_TEST_EXISTS | G_FILE_TEST_IS_DIR);
}

/**
 * taishang_utils_list_directory:
 * @dir_path: Path to directory
 * @error: Return location for error
 * 
 * Lists directory contents.
 * 
 * Returns: (transfer full) (array zero-terminated=1): Array of filenames
 */
gchar **taishang_utils_list_directory(const gchar *dir_path, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(dir_path != NULL, NULL);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, NULL);
    
    GDir *dir = g_dir_open(dir_path, 0, error);
    if (!dir) {
        return NULL;
    }
    
    GPtrArray *files = g_ptr_array_new();
    const gchar *name;
    
    while ((name = g_dir_read_name(dir)) != NULL) {
        g_ptr_array_add(files, g_strdup(name));
    }
    
    g_dir_close(dir);
    
    /* Null-terminate array */
    g_ptr_array_add(files, NULL);
    
    return (gchar **)g_ptr_array_free(files, FALSE);
}

/**
 * taishang_utils_remove_directory:
 * @dir_path: Path to directory
 * @error: Return location for error
 * 
 * Removes a directory (must be empty).
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_utils_remove_directory(const gchar *dir_path, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(dir_path != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    if (g_rmdir(dir_path) != 0) {
        g_set_error(error, G_FILE_ERROR, g_file_error_from_errno(errno),
                   "Failed to remove directory '%s': %s", dir_path, g_strerror(errno));
        return FALSE;
    }
    
    return TRUE;
}

/* Path utilities implementation */

/**
 * taishang_utils_path_join:
 * @first_element: First path element
 * @...: Additional path elements, terminated by NULL
 * 
 * Joins path elements.
 * 
 * Returns: (transfer full): Joined path
 */
gchar *taishang_utils_path_join(const gchar *first_element, ...) {
    if (!first_element) {
        return NULL;
    }
    
    va_list args;
    va_start(args, first_element);
    
    GPtrArray *elements = g_ptr_array_new();
    g_ptr_array_add(elements, (gpointer)first_element);
    
    const gchar *element;
    while ((element = va_arg(args, const gchar *)) != NULL) {
        g_ptr_array_add(elements, (gpointer)element);
    }
    
    va_end(args);
    
    /* Null-terminate array */
    g_ptr_array_add(elements, NULL);
    
    gchar *result = g_build_filenamev((gchar **)elements->pdata);
    g_ptr_array_free(elements, TRUE);
    
    return result;
}

/**
 * taishang_utils_path_normalize:
 * @path: Path to normalize
 * 
 * Normalizes a path (resolves . and .. components).
 * 
 * Returns: (transfer full): Normalized path
 */
gchar *taishang_utils_path_normalize(const gchar *path) {
    if (!path) {
        return NULL;
    }
    
    return g_canonicalize_filename(path, NULL);
}

/**
 * taishang_utils_path_get_basename:
 * @path: Path
 * 
 * Gets basename of path.
 * 
 * Returns: (transfer full): Basename
 */
gchar *taishang_utils_path_get_basename(const gchar *path) {
    if (!path) {
        return NULL;
    }
    
    return g_path_get_basename(path);
}

/**
 * taishang_utils_path_get_dirname:
 * @path: Path
 * 
 * Gets directory name of path.
 * 
 * Returns: (transfer full): Directory name
 */
gchar *taishang_utils_path_get_dirname(const gchar *path) {
    if (!path) {
        return NULL;
    }
    
    return g_path_get_dirname(path);
}

/**
 * taishang_utils_path_get_extension:
 * @path: Path
 * 
 * Gets file extension.
 * 
 * Returns: (transfer none): File extension (including dot)
 */
const gchar *taishang_utils_path_get_extension(const gchar *path) {
    if (!path) {
        return NULL;
    }
    
    const gchar *dot = strrchr(path, '.');
    const gchar *slash = strrchr(path, G_DIR_SEPARATOR);
    
    /* Make sure dot comes after last slash */
    if (dot && (!slash || dot > slash)) {
        return dot;
    }
    
    return NULL;
}

/* Time utilities implementation */

/**
 * taishang_utils_get_timestamp:
 * 
 * Gets current timestamp.
 * 
 * Returns: Current timestamp in seconds since epoch
 */
gint64 taishang_utils_get_timestamp(void) {
    return g_get_real_time() / G_USEC_PER_SEC;
}

/**
 * taishang_utils_format_time:
 * @timestamp: Timestamp to format
 * @format: Format string (strftime format)
 * 
 * Formats timestamp as string.
 * 
 * Returns: (transfer full): Formatted time string
 */
gchar *taishang_utils_format_time(gint64 timestamp, const gchar *format) {
    if (!format) {
        format = "%Y-%m-%d %H:%M:%S";
    }
    
    GDateTime *dt = g_date_time_new_from_unix_local(timestamp);
    if (!dt) {
        return NULL;
    }
    
    gchar *result = g_date_time_format(dt, format);
    g_date_time_unref(dt);
    
    return result;
}

/**
 * taishang_utils_parse_time:
 * @time_string: Time string to parse
 * @format: Format string (strftime format)
 * 
 * Parses time string to timestamp.
 * 
 * Returns: Timestamp, or -1 on error
 */
gint64 taishang_utils_parse_time(const gchar *time_string, const gchar *format) {
    if (!time_string || !format) {
        return -1;
    }
    
    struct tm tm = {0};
    if (!strptime(time_string, format, &tm)) {
        return -1;
    }
    
    return mktime(&tm);
}

/**
 * taishang_utils_get_elapsed_time:
 * @start_time: Start timestamp
 * 
 * Gets elapsed time since start.
 * 
 * Returns: Elapsed time in seconds
 */
gdouble taishang_utils_get_elapsed_time(gint64 start_time) {
    gint64 current_time = taishang_utils_get_timestamp();
    return (gdouble)(current_time - start_time);
}

/* Hash utilities implementation */

/**
 * taishang_utils_hash_md5:
 * @data: Data to hash
 * @length: Data length
 * 
 * Computes MD5 hash.
 * 
 * Returns: (transfer full): MD5 hash as hex string
 */
gchar *taishang_utils_hash_md5(const guchar *data, gsize length) {
    if (!data) {
        return NULL;
    }
    
    GChecksum *checksum = g_checksum_new(G_CHECKSUM_MD5);
    g_checksum_update(checksum, data, length);
    
    gchar *result = g_strdup(g_checksum_get_string(checksum));
    g_checksum_free(checksum);
    
    return result;
}

/**
 * taishang_utils_hash_sha256:
 * @data: Data to hash
 * @length: Data length
 * 
 * Computes SHA256 hash.
 * 
 * Returns: (transfer full): SHA256 hash as hex string
 */
gchar *taishang_utils_hash_sha256(const guchar *data, gsize length) {
    if (!data) {
        return NULL;
    }
    
    GChecksum *checksum = g_checksum_new(G_CHECKSUM_SHA256);
    g_checksum_update(checksum, data, length);
    
    gchar *result = g_strdup(g_checksum_get_string(checksum));
    g_checksum_free(checksum);
    
    return result;
}

/* Encoding utilities implementation */

/**
 * taishang_utils_base64_encode:
 * @data: Data to encode
 * @length: Data length
 * 
 * Encodes data as Base64.
 * 
 * Returns: (transfer full): Base64 encoded string
 */
gchar *taishang_utils_base64_encode(const guchar *data, gsize length) {
    if (!data) {
        return NULL;
    }
    
    return g_base64_encode(data, length);
}

/**
 * taishang_utils_base64_decode:
 * @text: Base64 text to decode
 * @out_length: Return location for decoded data length
 * 
 * Decodes Base64 text.
 * 
 * Returns: (transfer full): Decoded data
 */
guchar *taishang_utils_base64_decode(const gchar *text, gsize *out_length) {
    if (!text) {
        return NULL;
    }
    
    return g_base64_decode(text, out_length);
}

/**
 * taishang_utils_url_encode:
 * @text: Text to encode
 * 
 * URL encodes text.
 * 
 * Returns: (transfer full): URL encoded string
 */
gchar *taishang_utils_url_encode(const gchar *text) {
    if (!text) {
        return NULL;
    }
    
    return g_uri_escape_string(text, NULL, FALSE);
}

/**
 * taishang_utils_url_decode:
 * @text: URL encoded text
 * 
 * URL decodes text.
 * 
 * Returns: (transfer full): Decoded string
 */
gchar *taishang_utils_url_decode(const gchar *text) {
    if (!text) {
        return NULL;
    }
    
    return g_uri_unescape_string(text, NULL);
}

/* Random utilities implementation */

/**
 * taishang_utils_random_int:
 * @min: Minimum value
 * @max: Maximum value
 * 
 * Generates random integer.
 * 
 * Returns: Random integer between min and max (inclusive)
 */
gint taishang_utils_random_int(gint min, gint max) {
    if (min >= max) {
        return min;
    }
    
    return g_random_int_range(min, max + 1);
}

/**
 * taishang_utils_random_double:
 * 
 * Generates random double.
 * 
 * Returns: Random double between 0.0 and 1.0
 */
gdouble taishang_utils_random_double(void) {
    return g_random_double();
}

/**
 * taishang_utils_random_string:
 * @length: String length
 * 
 * Generates random string.
 * 
 * Returns: (transfer full): Random string
 */
gchar *taishang_utils_random_string(gsize length) {
    if (length == 0) {
        return g_strdup("");
    }
    
    static const gchar charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    gchar *result = g_malloc(length + 1);
    
    for (gsize i = 0; i < length; i++) {
        result[i] = charset[g_random_int_range(0, sizeof(charset) - 1)];
    }
    
    result[length] = '\0';
    return result;
}

/**
 * taishang_utils_generate_uuid:
 * 
 * Generates UUID.
 * 
 * Returns: (transfer full): UUID string
 */
gchar *taishang_utils_generate_uuid(void) {
    return g_uuid_string_random();
}

/* Memory utilities implementation */

/**
 * taishang_utils_safe_malloc:
 * @size: Size to allocate
 * 
 * Safe malloc with error checking.
 * 
 * Returns: (transfer full): Allocated memory
 */
gpointer taishang_utils_safe_malloc(gsize size) {
    if (size == 0) {
        return NULL;
    }
    
    gpointer ptr = g_malloc(size);
    if (!ptr) {
        g_error("Failed to allocate %zu bytes", size);
    }
    
    return ptr;
}

/**
 * taishang_utils_safe_realloc:
 * @ptr: Pointer to reallocate
 * @size: New size
 * 
 * Safe realloc with error checking.
 * 
 * Returns: (transfer full): Reallocated memory
 */
gpointer taishang_utils_safe_realloc(gpointer ptr, gsize size) {
    if (size == 0) {
        g_free(ptr);
        return NULL;
    }
    
    gpointer new_ptr = g_realloc(ptr, size);
    if (!new_ptr) {
        g_error("Failed to reallocate to %zu bytes", size);
    }
    
    return new_ptr;
}

/**
 * taishang_utils_safe_free:
 * @ptr: Pointer to free
 * 
 * Safe free that sets pointer to NULL.
 */
void taishang_utils_safe_free(gpointer *ptr) {
    if (ptr && *ptr) {
        g_free(*ptr);
        *ptr = NULL;
    }
}

/**
 * taishang_utils_memzero:
 * @ptr: Pointer to memory
 * @size: Size to zero
 * 
 * Securely zeros memory.
 */
void taishang_utils_memzero(gpointer ptr, gsize size) {
    if (ptr && size > 0) {
        volatile guchar *p = (volatile guchar *)ptr;
        for (gsize i = 0; i < size; i++) {
            p[i] = 0;
        }
    }
}

/* Memory pool implementation */

/**
 * taishang_utils_memory_pool_new:
 * @block_size: Size of each block
 * @blocks_per_chunk: Number of blocks per chunk
 * 
 * Creates a new memory pool.
 * 
 * Returns: (transfer full): New memory pool
 */
TaishangMemoryPool *taishang_utils_memory_pool_new(gsize block_size, gsize blocks_per_chunk) {
    if (block_size == 0 || blocks_per_chunk == 0) {
        return NULL;
    }
    
    TaishangMemoryPool *pool = g_new0(TaishangMemoryPool, 1);
    pool->block_size = block_size;
    pool->blocks_per_chunk = blocks_per_chunk;
    pool->chunks = NULL;
    pool->free_blocks = NULL;
    g_mutex_init(&pool->mutex);
    pool->total_allocated = 0;
    pool->total_freed = 0;
    
    return pool;
}

/**
 * taishang_utils_memory_pool_alloc:
 * @pool: Memory pool
 * 
 * Allocates a block from the pool.
 * 
 * Returns: (transfer full): Allocated block
 */
gpointer taishang_utils_memory_pool_alloc(TaishangMemoryPool *pool) {
    if (!pool) {
        return NULL;
    }
    
    g_mutex_lock(&pool->mutex);
    
    /* Check if we have free blocks */
    if (pool->free_blocks) {
        gpointer block = pool->free_blocks->data;
        pool->free_blocks = g_slist_delete_link(pool->free_blocks, pool->free_blocks);
        pool->total_allocated++;
        g_mutex_unlock(&pool->mutex);
        return block;
    }
    
    /* Allocate new chunk */
    gsize chunk_size = pool->block_size * pool->blocks_per_chunk;
    gpointer chunk_data = g_malloc(chunk_size);
    
    MemoryChunk *chunk = g_new(MemoryChunk, 1);
    chunk->data = chunk_data;
    chunk->size = chunk_size;
    pool->chunks = g_slist_prepend(pool->chunks, chunk);
    
    /* Add blocks to free list (except the first one) */
    for (gsize i = 1; i < pool->blocks_per_chunk; i++) {
        gpointer block = (guchar *)chunk_data + (i * pool->block_size);
        pool->free_blocks = g_slist_prepend(pool->free_blocks, block);
    }
    
    pool->total_allocated++;
    g_mutex_unlock(&pool->mutex);
    
    return chunk_data;
}

/**
 * taishang_utils_memory_pool_free:
 * @pool: Memory pool
 * @ptr: Pointer to free
 * 
 * Frees a block back to the pool.
 */
void taishang_utils_memory_pool_free(TaishangMemoryPool *pool, gpointer ptr) {
    if (!pool || !ptr) {
        return;
    }
    
    g_mutex_lock(&pool->mutex);
    pool->free_blocks = g_slist_prepend(pool->free_blocks, ptr);
    pool->total_freed++;
    g_mutex_unlock(&pool->mutex);
}

/**
 * taishang_utils_memory_pool_destroy:
 * @pool: Memory pool to destroy
 * 
 * Destroys a memory pool.
 */
void taishang_utils_memory_pool_destroy(TaishangMemoryPool *pool) {
    if (!pool) {
        return;
    }
    
    g_mutex_lock(&pool->mutex);
    
    /* Free all chunks */
    g_slist_free_full(pool->chunks, (GDestroyNotify)memory_chunk_free);
    g_slist_free(pool->free_blocks);
    
    g_mutex_unlock(&pool->mutex);
    g_mutex_clear(&pool->mutex);
    
    g_free(pool);
}

/* Logging utilities implementation */

/**
 * taishang_utils_init_logging:
 * @log_file: Log file path (NULL for stderr)
 * @level: Log level
 * 
 * Initializes logging system.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_utils_init_logging(const gchar *log_file, GLogLevelFlags level) {
    g_mutex_lock(&log_mutex);
    
    /* Close existing log file */
    if (log_file && log_file != stderr) {
        fclose(log_file);
        log_file = NULL;
    }
    
    /* Set log level */
    log_level = level;
    
    /* Open new log file */
    if (log_file) {
        TAISHANG_FREE(log_file_path);
        log_file_path = g_strdup(log_file);
        
        log_file = fopen(log_file_path, "a");
        if (!log_file) {
            g_mutex_unlock(&log_mutex);
            return FALSE;
        }
    }
    
    /* Set log handler */
    g_log_set_default_handler(taishang_utils_log_handler, NULL);
    
    g_mutex_unlock(&log_mutex);
    
    return TRUE;
}

/**
 * taishang_utils_log_message:
 * @level: Log level
 * @format: Format string
 * @...: Format arguments
 * 
 * Logs a message.
 */
void taishang_utils_log_message(GLogLevelFlags level, const gchar *format, ...) {
    if (!format || !(level & log_level)) {
        return;
    }
    
    va_list args;
    va_start(args, format);
    
    gchar *message = g_strdup_vprintf(format, args);
    g_log(NULL, level, "%s", message);
    
    g_free(message);
    va_end(args);
}

/* Process utilities implementation */

/**
 * taishang_utils_get_process_id:
 * 
 * Gets current process ID.
 * 
 * Returns: Process ID
 */
gint taishang_utils_get_process_id(void) {
    return getpid();
}

/**
 * taishang_utils_get_parent_process_id:
 * 
 * Gets parent process ID.
 * 
 * Returns: Parent process ID
 */
gint taishang_utils_get_parent_process_id(void) {
    return getppid();
}

/**
 * taishang_utils_get_user_name:
 * 
 * Gets current user name.
 * 
 * Returns: (transfer full): User name
 */
gchar *taishang_utils_get_user_name(void) {
    struct passwd *pw = getpwuid(getuid());
    if (!pw) {
        return NULL;
    }
    
    return g_strdup(pw->pw_name);
}

/**
 * taishang_utils_get_home_directory:
 * 
 * Gets user home directory.
 * 
 * Returns: (transfer none): Home directory path
 */
const gchar *taishang_utils_get_home_directory(void) {
    return g_get_home_dir();
}

/* System utilities implementation */

/**
 * taishang_utils_get_system_info:
 * 
 * Gets system information.
 * 
 * Returns: (transfer full): System info string
 */
gchar *taishang_utils_get_system_info(void) {
    struct utsname uts;
    if (uname(&uts) != 0) {
        return g_strdup("Unknown system");
    }
    
    return g_strdup_printf("%s %s %s %s", uts.sysname, uts.release, uts.version, uts.machine);
}

/**
 * taishang_utils_get_memory_usage:
 * 
 * Gets memory usage information.
 * 
 * Returns: Memory usage in bytes
 */
gint64 taishang_utils_get_memory_usage(void) {
    struct sysinfo si;
    if (sysinfo(&si) != 0) {
        return -1;
    }
    
    return (gint64)(si.totalram - si.freeram) * si.mem_unit;
}

/* Private helper functions */

static void taishang_utils_log_handler(const gchar *log_domain, GLogLevelFlags log_level_flags, const gchar *message, gpointer user_data) {
    g_mutex_lock(&log_mutex);
    
    FILE *output = log_file ? log_file : stderr;
    gchar *timestamp = taishang_utils_get_timestamp_string();
    
    const gchar *level_str;
    switch (log_level_flags & G_LOG_LEVEL_MASK) {
        case G_LOG_LEVEL_ERROR:
            level_str = "ERROR";
            break;
        case G_LOG_LEVEL_CRITICAL:
            level_str = "CRITICAL";
            break;
        case G_LOG_LEVEL_WARNING:
            level_str = "WARNING";
            break;
        case G_LOG_LEVEL_MESSAGE:
            level_str = "MESSAGE";
            break;
        case G_LOG_LEVEL_INFO:
            level_str = "INFO";
            break;
        case G_LOG_LEVEL_DEBUG:
            level_str = "DEBUG";
            break;
        default:
            level_str = "UNKNOWN";
            break;
    }
    
    fprintf(output, "[%s] %s: %s\n", timestamp, level_str, message);
    fflush(output);
    
    g_free(timestamp);
    g_mutex_unlock(&log_mutex);
}

static gchar *taishang_utils_get_timestamp_string(void) {
    GDateTime *now = g_date_time_new_now_local();
    gchar *timestamp = g_date_time_format(now, "%Y-%m-%d %H:%M:%S");
    g_date_time_unref(now);
    return timestamp;
}

static void memory_chunk_free(MemoryChunk *chunk) {
    if (chunk) {
        g_free(chunk->data);
        g_free(chunk);
    }
}