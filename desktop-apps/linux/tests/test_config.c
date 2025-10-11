/**
 * @file test_config.c
 * @brief Unit tests for configuration management
 * @author TaishangLaojun Development Team
 * @date 2024
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <unistd.h>
#include <sys/stat.h>
#include <errno.h>

#include "config.h"

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

// Test configuration file path
static const char *TEST_CONFIG_FILE = "/tmp/taishang_test_config.json";

// Test configuration initialization
static int test_config_init(void) {
    Config *config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    // Test default values
    TEST_ASSERT(strlen(config->app_name) > 0, "App name should be set");
    TEST_ASSERT(strlen(config->version) > 0, "Version should be set");
    TEST_ASSERT(config->window_width > 0, "Window width should be positive");
    TEST_ASSERT(config->window_height > 0, "Window height should be positive");
    
    config_cleanup(config);
    return 1;
}

// Test configuration loading from file
static int test_config_load_from_file(void) {
    // Create a test configuration file
    FILE *file = fopen(TEST_CONFIG_FILE, "w");
    TEST_ASSERT(file != NULL, "Test config file should be created");
    
    fprintf(file, "{\n");
    fprintf(file, "  \"app_name\": \"TaishangLaojun Test\",\n");
    fprintf(file, "  \"version\": \"1.0.0-test\",\n");
    fprintf(file, "  \"window_width\": 1024,\n");
    fprintf(file, "  \"window_height\": 768,\n");
    fprintf(file, "  \"theme\": \"dark\",\n");
    fprintf(file, "  \"auto_start\": true,\n");
    fprintf(file, "  \"notifications_enabled\": false\n");
    fprintf(file, "}\n");
    fclose(file);
    
    // Load configuration from file
    Config *config = config_load_from_file(TEST_CONFIG_FILE);
    TEST_ASSERT(config != NULL, "Configuration should be loaded from file");
    
    TEST_ASSERT(strcmp(config->app_name, "TaishangLaojun Test") == 0, "App name should match");
    TEST_ASSERT(strcmp(config->version, "1.0.0-test") == 0, "Version should match");
    TEST_ASSERT(config->window_width == 1024, "Window width should match");
    TEST_ASSERT(config->window_height == 768, "Window height should match");
    TEST_ASSERT(strcmp(config->theme, "dark") == 0, "Theme should match");
    TEST_ASSERT(config->auto_start == 1, "Auto start should be enabled");
    TEST_ASSERT(config->notifications_enabled == 0, "Notifications should be disabled");
    
    config_cleanup(config);
    unlink(TEST_CONFIG_FILE);
    return 1;
}

// Test configuration saving to file
static int test_config_save_to_file(void) {
    Config *config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    // Modify some values
    strncpy(config->app_name, "TaishangLaojun Save Test", sizeof(config->app_name) - 1);
    strncpy(config->theme, "light", sizeof(config->theme) - 1);
    config->window_width = 1200;
    config->window_height = 800;
    config->auto_start = 0;
    
    // Save to file
    int result = config_save_to_file(config, TEST_CONFIG_FILE);
    TEST_ASSERT(result == 0, "Configuration should be saved successfully");
    
    // Verify file exists
    struct stat st;
    TEST_ASSERT(stat(TEST_CONFIG_FILE, &st) == 0, "Config file should exist");
    
    // Load and verify
    Config *loaded_config = config_load_from_file(TEST_CONFIG_FILE);
    TEST_ASSERT(loaded_config != NULL, "Configuration should be loaded");
    
    TEST_ASSERT(strcmp(loaded_config->app_name, "TaishangLaojun Save Test") == 0, "App name should match");
    TEST_ASSERT(strcmp(loaded_config->theme, "light") == 0, "Theme should match");
    TEST_ASSERT(loaded_config->window_width == 1200, "Window width should match");
    TEST_ASSERT(loaded_config->window_height == 800, "Window height should match");
    TEST_ASSERT(loaded_config->auto_start == 0, "Auto start should be disabled");
    
    config_cleanup(config);
    config_cleanup(loaded_config);
    unlink(TEST_CONFIG_FILE);
    return 1;
}

// Test configuration validation
static int test_config_validation(void) {
    Config *config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    // Test valid configuration
    int result = config_validate(config);
    TEST_ASSERT(result == 0, "Default configuration should be valid");
    
    // Test invalid window dimensions
    config->window_width = -100;
    result = config_validate(config);
    TEST_ASSERT(result != 0, "Invalid window width should fail validation");
    
    config->window_width = 800;
    config->window_height = 0;
    result = config_validate(config);
    TEST_ASSERT(result != 0, "Invalid window height should fail validation");
    
    // Test invalid theme
    config->window_height = 600;
    strncpy(config->theme, "", sizeof(config->theme) - 1);
    result = config_validate(config);
    TEST_ASSERT(result != 0, "Empty theme should fail validation");
    
    config_cleanup(config);
    return 1;
}

// Test configuration defaults
static int test_config_defaults(void) {
    Config *config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    // Test default values
    TEST_ASSERT(config->window_width >= 800, "Default window width should be reasonable");
    TEST_ASSERT(config->window_height >= 600, "Default window height should be reasonable");
    TEST_ASSERT(strlen(config->theme) > 0, "Default theme should be set");
    TEST_ASSERT(config->notifications_enabled == 1, "Notifications should be enabled by default");
    TEST_ASSERT(config->auto_start == 0, "Auto start should be disabled by default");
    
    config_cleanup(config);
    return 1;
}

// Test configuration getters and setters
static int test_config_getters_setters(void) {
    Config *config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    // Test string setters/getters
    config_set_theme(config, "custom");
    const char *theme = config_get_theme(config);
    TEST_ASSERT(strcmp(theme, "custom") == 0, "Theme should be set correctly");
    
    // Test integer setters/getters
    config_set_window_size(config, 1024, 768);
    int width = config_get_window_width(config);
    int height = config_get_window_height(config);
    TEST_ASSERT(width == 1024, "Window width should be set correctly");
    TEST_ASSERT(height == 768, "Window height should be set correctly");
    
    // Test boolean setters/getters
    config_set_auto_start(config, 1);
    int auto_start = config_get_auto_start(config);
    TEST_ASSERT(auto_start == 1, "Auto start should be enabled");
    
    config_set_notifications_enabled(config, 0);
    int notifications = config_get_notifications_enabled(config);
    TEST_ASSERT(notifications == 0, "Notifications should be disabled");
    
    config_cleanup(config);
    return 1;
}

// Test configuration file error handling
static int test_config_file_error_handling(void) {
    // Test loading from non-existent file
    Config *config = config_load_from_file("/non/existent/path/config.json");
    TEST_ASSERT(config == NULL, "Loading from non-existent file should fail");
    
    // Test saving to invalid path
    config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    int result = config_save_to_file(config, "/invalid/path/config.json");
    TEST_ASSERT(result != 0, "Saving to invalid path should fail");
    
    config_cleanup(config);
    return 1;
}

// Test configuration memory management
static int test_config_memory_management(void) {
    // Test multiple init/cleanup cycles
    for (int i = 0; i < 10; i++) {
        Config *config = config_init();
        TEST_ASSERT(config != NULL, "Configuration should be initialized");
        config_cleanup(config);
    }
    
    // Test cleanup with NULL
    config_cleanup(NULL);  // Should not crash
    
    return 1;
}

// Test configuration JSON parsing
static int test_config_json_parsing(void) {
    // Create malformed JSON file
    FILE *file = fopen(TEST_CONFIG_FILE, "w");
    TEST_ASSERT(file != NULL, "Test config file should be created");
    
    fprintf(file, "{\n");
    fprintf(file, "  \"app_name\": \"Test\",\n");
    fprintf(file, "  \"invalid_json\": \n");  // Malformed JSON
    fprintf(file, "}\n");
    fclose(file);
    
    // Try to load malformed JSON
    Config *config = config_load_from_file(TEST_CONFIG_FILE);
    TEST_ASSERT(config == NULL, "Loading malformed JSON should fail");
    
    unlink(TEST_CONFIG_FILE);
    return 1;
}

// Test configuration backup and restore
static int test_config_backup_restore(void) {
    Config *config = config_init();
    TEST_ASSERT(config != NULL, "Configuration should be initialized");
    
    // Modify configuration
    strncpy(config->theme, "backup_test", sizeof(config->theme) - 1);
    config->window_width = 1337;
    
    // Create backup
    int result = config_create_backup(config, TEST_CONFIG_FILE);
    TEST_ASSERT(result == 0, "Configuration backup should be created");
    
    // Modify configuration again
    strncpy(config->theme, "modified", sizeof(config->theme) - 1);
    config->window_width = 999;
    
    // Restore from backup
    result = config_restore_from_backup(config, TEST_CONFIG_FILE);
    TEST_ASSERT(result == 0, "Configuration should be restored from backup");
    
    TEST_ASSERT(strcmp(config->theme, "backup_test") == 0, "Theme should be restored");
    TEST_ASSERT(config->window_width == 1337, "Window width should be restored");
    
    config_cleanup(config);
    unlink(TEST_CONFIG_FILE);
    return 1;
}

// Main test runner
int main(void) {
    printf("=== TaishangLaojun Configuration Tests ===\n\n");
    
    RUN_TEST(test_config_init);
    RUN_TEST(test_config_load_from_file);
    RUN_TEST(test_config_save_to_file);
    RUN_TEST(test_config_validation);
    RUN_TEST(test_config_defaults);
    RUN_TEST(test_config_getters_setters);
    RUN_TEST(test_config_file_error_handling);
    RUN_TEST(test_config_memory_management);
    RUN_TEST(test_config_json_parsing);
    RUN_TEST(test_config_backup_restore);
    
    printf("\n=== All Configuration Tests Passed! ===\n");
    return 0;
}