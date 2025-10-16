/**
 * @file test_utils.c
 * @brief Unit tests for utility functions
 * @author TaishangLaojun Development Team
 * @date 2024
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <unistd.h>
#include <sys/stat.h>
#include <time.h>
#include <errno.h>

#include "utils.h"

// Test framework macros
#define TEST_ASSERT(condition, message) \
    do { \
        if (!(condition)) { \
            fprintf(stderr, "FAIL: %s - %s\n", __func__, message); \
            return 0; \
        } \
        printf("PASS: %s - %s\n", __func__, message); \
    } while(0)

#define RUN_TEST(test_func) \
    do { \
        printf("Running %s...\n", #test_func); \
        if (!test_func()) { \
            fprintf(stderr, "Test %s failed!\n", #test_func); \
            return 1; \
        } \
    } while(0)

// Test string utilities
static int test_string_utils(void) {
    // Test string trimming
    char test_str1[] = "  hello world  ";
    char *trimmed = utils_string_trim(test_str1);
    TEST_ASSERT(strcmp(trimmed, "hello world") == 0, "String should be trimmed");
    free(trimmed);
    
    // Test string duplication
    const char *original = "test string";
    char *duplicated = utils_string_duplicate(original);
    TEST_ASSERT(duplicated != NULL, "String should be duplicated");
    TEST_ASSERT(strcmp(duplicated, original) == 0, "Duplicated string should match original");
    TEST_ASSERT(duplicated != original, "Duplicated string should be different pointer");
    free(duplicated);
    
    // Test string case conversion
    char test_str2[] = "Hello World";
    char *lowercase = utils_string_to_lowercase(test_str2);
    TEST_ASSERT(strcmp(lowercase, "hello world") == 0, "String should be converted to lowercase");
    free(lowercase);
    
    char *uppercase = utils_string_to_uppercase(test_str2);
    TEST_ASSERT(strcmp(uppercase, "HELLO WORLD") == 0, "String should be converted to uppercase");
    free(uppercase);
    
    // Test string splitting
    const char *csv = "apple,banana,cherry";
    char **tokens = utils_string_split(csv, ",");
    TEST_ASSERT(tokens != NULL, "String should be split");
    TEST_ASSERT(strcmp(tokens[0], "apple") == 0, "First token should be 'apple'");
    TEST_ASSERT(strcmp(tokens[1], "banana") == 0, "Second token should be 'banana'");
    TEST_ASSERT(strcmp(tokens[2], "cherry") == 0, "Third token should be 'cherry'");
    TEST_ASSERT(tokens[3] == NULL, "Token array should be NULL-terminated");
    utils_string_array_free(tokens);
    
    return 1;
}

// Test file utilities
static int test_file_utils(void) {
    const char *test_file = "/tmp/taishang_test_file.txt";
    const char *test_content = "This is a test file content.";
    
    // Test file writing
    int result = utils_file_write(test_file, test_content);
    TEST_ASSERT(result == 0, "File should be written successfully");
    
    // Test file existence
    TEST_ASSERT(utils_file_exists(test_file), "File should exist");
    
    // Test file reading
    char *content = utils_file_read(test_file);
    TEST_ASSERT(content != NULL, "File content should be read");
    TEST_ASSERT(strcmp(content, test_content) == 0, "File content should match");
    free(content);
    
    // Test file size
    size_t size = utils_file_get_size(test_file);
    TEST_ASSERT(size == strlen(test_content), "File size should match content length");
    
    // Test file deletion
    result = utils_file_delete(test_file);
    TEST_ASSERT(result == 0, "File should be deleted successfully");
    TEST_ASSERT(!utils_file_exists(test_file), "File should not exist after deletion");
    
    return 1;
}

// Test directory utilities
static int test_directory_utils(void) {
    const char *test_dir = "/tmp/taishang_test_dir";
    
    // Test directory creation
    int result = utils_directory_create(test_dir);
    TEST_ASSERT(result == 0, "Directory should be created successfully");
    TEST_ASSERT(utils_directory_exists(test_dir), "Directory should exist");
    
    // Test directory listing
    char **files = utils_directory_list(test_dir);
    TEST_ASSERT(files != NULL, "Directory listing should succeed");
    utils_string_array_free(files);
    
    // Test directory removal
    result = utils_directory_remove(test_dir);
    TEST_ASSERT(result == 0, "Directory should be removed successfully");
    TEST_ASSERT(!utils_directory_exists(test_dir), "Directory should not exist after removal");
    
    return 1;
}

// Test path utilities
static int test_path_utils(void) {
    // Test path joining
    char *joined = utils_path_join("/home/user", "documents/file.txt");
    TEST_ASSERT(joined != NULL, "Path should be joined");
    TEST_ASSERT(strcmp(joined, "/home/user/documents/file.txt") == 0, "Joined path should be correct");
    free(joined);
    
    // Test path normalization
    char *normalized = utils_path_normalize("/home/user/../user/./documents");
    TEST_ASSERT(normalized != NULL, "Path should be normalized");
    TEST_ASSERT(strcmp(normalized, "/home/user/documents") == 0, "Normalized path should be correct");
    free(normalized);
    
    // Test basename extraction
    char *basename = utils_path_get_basename("/home/user/documents/file.txt");
    TEST_ASSERT(basename != NULL, "Basename should be extracted");
    TEST_ASSERT(strcmp(basename, "file.txt") == 0, "Basename should be correct");
    free(basename);
    
    // Test dirname extraction
    char *dirname = utils_path_get_dirname("/home/user/documents/file.txt");
    TEST_ASSERT(dirname != NULL, "Dirname should be extracted");
    TEST_ASSERT(strcmp(dirname, "/home/user/documents") == 0, "Dirname should be correct");
    free(dirname);
    
    // Test extension extraction
    char *extension = utils_path_get_extension("file.txt");
    TEST_ASSERT(extension != NULL, "Extension should be extracted");
    TEST_ASSERT(strcmp(extension, "txt") == 0, "Extension should be correct");
    free(extension);
    
    return 1;
}

// Test time utilities
static int test_time_utils(void) {
    // Test current timestamp
    time_t timestamp = utils_time_get_current_timestamp();
    TEST_ASSERT(timestamp > 0, "Current timestamp should be positive");
    
    // Test time formatting
    char *formatted = utils_time_format_timestamp(timestamp, "%Y-%m-%d %H:%M:%S");
    TEST_ASSERT(formatted != NULL, "Timestamp should be formatted");
    TEST_ASSERT(strlen(formatted) > 0, "Formatted time should not be empty");
    free(formatted);
    
    // Test time parsing
    time_t parsed = utils_time_parse_iso8601("2024-01-01T12:00:00Z");
    TEST_ASSERT(parsed > 0, "ISO8601 time should be parsed");
    
    // Test elapsed time
    time_t start = utils_time_get_current_timestamp();
    usleep(10000);  // Sleep for 10ms
    double elapsed = utils_time_get_elapsed_seconds(start);
    TEST_ASSERT(elapsed > 0.0, "Elapsed time should be positive");
    TEST_ASSERT(elapsed < 1.0, "Elapsed time should be less than 1 second");
    
    return 1;
}

// Test hash utilities
static int test_hash_utils(void) {
    const char *test_data = "Hello, World!";
    
    // Test MD5 hash
    char *md5_hash = utils_hash_md5(test_data);
    TEST_ASSERT(md5_hash != NULL, "MD5 hash should be generated");
    TEST_ASSERT(strlen(md5_hash) == 32, "MD5 hash should be 32 characters");
    free(md5_hash);
    
    // Test SHA256 hash
    char *sha256_hash = utils_hash_sha256(test_data);
    TEST_ASSERT(sha256_hash != NULL, "SHA256 hash should be generated");
    TEST_ASSERT(strlen(sha256_hash) == 64, "SHA256 hash should be 64 characters");
    free(sha256_hash);
    
    // Test hash consistency
    char *hash1 = utils_hash_sha256(test_data);
    char *hash2 = utils_hash_sha256(test_data);
    TEST_ASSERT(strcmp(hash1, hash2) == 0, "Same input should produce same hash");
    free(hash1);
    free(hash2);
    
    return 1;
}

// Test encoding utilities
static int test_encoding_utils(void) {
    const char *test_data = "Hello, World!";
    
    // Test Base64 encoding
    char *encoded = utils_base64_encode(test_data, strlen(test_data));
    TEST_ASSERT(encoded != NULL, "Base64 encoding should succeed");
    TEST_ASSERT(strlen(encoded) > 0, "Encoded string should not be empty");
    
    // Test Base64 decoding
    size_t decoded_len;
    char *decoded = utils_base64_decode(encoded, &decoded_len);
    TEST_ASSERT(decoded != NULL, "Base64 decoding should succeed");
    TEST_ASSERT(decoded_len == strlen(test_data), "Decoded length should match original");
    TEST_ASSERT(memcmp(decoded, test_data, decoded_len) == 0, "Decoded data should match original");
    
    free(encoded);
    free(decoded);
    
    // Test URL encoding
    const char *url_test = "hello world & special chars!";
    char *url_encoded = utils_url_encode(url_test);
    TEST_ASSERT(url_encoded != NULL, "URL encoding should succeed");
    
    char *url_decoded = utils_url_decode(url_encoded);
    TEST_ASSERT(url_decoded != NULL, "URL decoding should succeed");
    TEST_ASSERT(strcmp(url_decoded, url_test) == 0, "URL decoded data should match original");
    
    free(url_encoded);
    free(url_decoded);
    
    return 1;
}

// Test random utilities
static int test_random_utils(void) {
    // Test random number generation
    utils_random_seed();
    
    int random_int = utils_random_int(1, 100);
    TEST_ASSERT(random_int >= 1 && random_int <= 100, "Random int should be in range");
    
    double random_double = utils_random_double();
    TEST_ASSERT(random_double >= 0.0 && random_double < 1.0, "Random double should be in range [0, 1)");
    
    // Test random string generation
    char *random_string = utils_random_string(16);
    TEST_ASSERT(random_string != NULL, "Random string should be generated");
    TEST_ASSERT(strlen(random_string) == 16, "Random string should have correct length");
    free(random_string);
    
    // Test UUID generation
    char *uuid = utils_generate_uuid();
    TEST_ASSERT(uuid != NULL, "UUID should be generated");
    TEST_ASSERT(strlen(uuid) == 36, "UUID should be 36 characters long");
    free(uuid);
    
    return 1;
}

// Test memory utilities
static int test_memory_utils(void) {
    // Test safe memory allocation
    void *ptr = utils_malloc(1024);
    TEST_ASSERT(ptr != NULL, "Memory allocation should succeed");
    
    // Test memory zeroing
    utils_memzero(ptr, 1024);
    
    // Test safe memory reallocation
    ptr = utils_realloc(ptr, 2048);
    TEST_ASSERT(ptr != NULL, "Memory reallocation should succeed");
    
    // Test safe memory freeing
    utils_free(ptr);
    
    // Test memory pool
    MemoryPool *pool = utils_memory_pool_create(1024, 16);
    TEST_ASSERT(pool != NULL, "Memory pool should be created");
    
    void *pool_ptr = utils_memory_pool_alloc(pool);
    TEST_ASSERT(pool_ptr != NULL, "Pool allocation should succeed");
    
    utils_memory_pool_free(pool, pool_ptr);
    utils_memory_pool_destroy(pool);
    
    return 1;
}

// Test logging utilities
static int test_logging_utils(void) {
    const char *log_file = "/tmp/taishang_test.log";
    
    // Test log initialization
    int result = utils_log_init(log_file, LOG_LEVEL_DEBUG);
    TEST_ASSERT(result == 0, "Log initialization should succeed");
    
    // Test logging functions
    utils_log_debug("Debug message");
    utils_log_info("Info message");
    utils_log_warning("Warning message");
    utils_log_error("Error message");
    
    // Test log file existence
    TEST_ASSERT(utils_file_exists(log_file), "Log file should exist");
    
    // Test log cleanup
    utils_log_cleanup();
    
    unlink(log_file);
    return 1;
}

// Main test runner
int main(void) {
    printf("=== TaishangLaojun Utilities Tests ===\n\n");
    
    RUN_TEST(test_string_utils);
    RUN_TEST(test_file_utils);
    RUN_TEST(test_directory_utils);
    RUN_TEST(test_path_utils);
    RUN_TEST(test_time_utils);
    RUN_TEST(test_hash_utils);
    RUN_TEST(test_encoding_utils);
    RUN_TEST(test_random_utils);
    RUN_TEST(test_memory_utils);
    RUN_TEST(test_logging_utils);
    
    printf("\n=== All Utilities Tests Passed! ===\n");
    return 0;
}