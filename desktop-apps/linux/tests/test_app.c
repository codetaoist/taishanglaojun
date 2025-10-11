/**
 * @file test_app.c
 * @brief Unit tests for application core functionality
 * @author TaishangLaojun Development Team
 * @date 2024
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <unistd.h>
#include <sys/stat.h>

#include "app.h"
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

// Test application initialization
static int test_app_init(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    TEST_ASSERT(ctx->is_running == 0, "Application should not be running initially");
    TEST_ASSERT(ctx->config != NULL, "Configuration should be loaded");
    
    app_cleanup(ctx);
    return 1;
}

// Test application startup
static int test_app_startup(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    int result = app_startup(ctx);
    TEST_ASSERT(result == 0, "Application startup should succeed");
    TEST_ASSERT(ctx->is_running == 1, "Application should be running after startup");
    
    app_shutdown(ctx);
    app_cleanup(ctx);
    return 1;
}

// Test application shutdown
static int test_app_shutdown(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    app_startup(ctx);
    TEST_ASSERT(ctx->is_running == 1, "Application should be running");
    
    app_shutdown(ctx);
    TEST_ASSERT(ctx->is_running == 0, "Application should not be running after shutdown");
    
    app_cleanup(ctx);
    return 1;
}

// Test application configuration loading
static int test_app_config_loading(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    Config *config = ctx->config;
    TEST_ASSERT(config != NULL, "Configuration should be loaded");
    TEST_ASSERT(strlen(config->app_name) > 0, "Application name should be set");
    TEST_ASSERT(strlen(config->version) > 0, "Version should be set");
    
    app_cleanup(ctx);
    return 1;
}

// Test application state management
static int test_app_state_management(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    // Test initial state
    AppState state = app_get_state(ctx);
    TEST_ASSERT(state == APP_STATE_INITIALIZED, "Initial state should be INITIALIZED");
    
    // Test state transitions
    app_set_state(ctx, APP_STATE_RUNNING);
    state = app_get_state(ctx);
    TEST_ASSERT(state == APP_STATE_RUNNING, "State should be RUNNING");
    
    app_set_state(ctx, APP_STATE_PAUSED);
    state = app_get_state(ctx);
    TEST_ASSERT(state == APP_STATE_PAUSED, "State should be PAUSED");
    
    app_cleanup(ctx);
    return 1;
}

// Test application error handling
static int test_app_error_handling(void) {
    // Test with NULL context
    int result = app_startup(NULL);
    TEST_ASSERT(result != 0, "Startup with NULL context should fail");
    
    result = app_shutdown(NULL);
    TEST_ASSERT(result != 0, "Shutdown with NULL context should fail");
    
    AppState state = app_get_state(NULL);
    TEST_ASSERT(state == APP_STATE_ERROR, "Get state with NULL context should return ERROR");
    
    return 1;
}

// Test application resource management
static int test_app_resource_management(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    // Test resource allocation
    int result = app_allocate_resources(ctx);
    TEST_ASSERT(result == 0, "Resource allocation should succeed");
    
    // Test resource cleanup
    app_free_resources(ctx);
    
    app_cleanup(ctx);
    return 1;
}

// Test application signal handling
static int test_app_signal_handling(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    // Test signal handler registration
    int result = app_setup_signal_handlers(ctx);
    TEST_ASSERT(result == 0, "Signal handler setup should succeed");
    
    app_cleanup(ctx);
    return 1;
}

// Test application logging
static int test_app_logging(void) {
    AppContext *ctx = app_init();
    TEST_ASSERT(ctx != NULL, "Application context should be initialized");
    
    // Test log initialization
    int result = app_init_logging(ctx);
    TEST_ASSERT(result == 0, "Logging initialization should succeed");
    
    // Test logging functions
    app_log_info(ctx, "Test info message");
    app_log_warning(ctx, "Test warning message");
    app_log_error(ctx, "Test error message");
    
    app_cleanup_logging(ctx);
    app_cleanup(ctx);
    return 1;
}

// Test application version information
static int test_app_version_info(void) {
    const char *version = app_get_version();
    TEST_ASSERT(version != NULL, "Version string should not be NULL");
    TEST_ASSERT(strlen(version) > 0, "Version string should not be empty");
    
    const char *build_info = app_get_build_info();
    TEST_ASSERT(build_info != NULL, "Build info should not be NULL");
    
    return 1;
}

// Main test runner
int main(void) {
    printf("=== TaishangLaojun Application Tests ===\n\n");
    
    RUN_TEST(test_app_init);
    RUN_TEST(test_app_startup);
    RUN_TEST(test_app_shutdown);
    RUN_TEST(test_app_config_loading);
    RUN_TEST(test_app_state_management);
    RUN_TEST(test_app_error_handling);
    RUN_TEST(test_app_resource_management);
    RUN_TEST(test_app_signal_handling);
    RUN_TEST(test_app_logging);
    RUN_TEST(test_app_version_info);
    
    printf("\n=== All Application Tests Passed! ===\n");
    return 0;
}