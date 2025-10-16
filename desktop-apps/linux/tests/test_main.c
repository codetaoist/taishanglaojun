#include <glib.h>
#include <gtk/gtk.h>
#include <stdio.h>
#include <stdlib.h>

// Test includes
#include "test_network.h"
#include "test_storage.h"
#include "test_system.h"
#include "test_graphics.h"
#include "test_audio.h"
#include "test_ui.h"

// Test configuration
typedef struct {
    gboolean verbose;
    gboolean quick;
    gchar *test_filter;
    gchar *output_file;
} TestConfig;

// Global test configuration
static TestConfig g_test_config = {0};

// Test statistics
typedef struct {
    gint total_tests;
    gint passed_tests;
    gint failed_tests;
    gint skipped_tests;
    gdouble total_time;
} TestStats;

static TestStats g_test_stats = {0};

// Test result logging
static FILE *g_log_file = NULL;

// Helper functions
static void print_test_header(void);
static void print_test_summary(void);
static void setup_test_environment(void);
static void cleanup_test_environment(void);
static gboolean parse_command_line(int argc, char **argv);
static void print_usage(const char *program_name);

// Test logging functions
static void test_log(const char *format, ...) G_GNUC_PRINTF(1, 2);
static void test_log_result(const char *test_name, gboolean passed, gdouble time);

int main(int argc, char **argv) {
    // Initialize GLib test framework
    g_test_init(&argc, &argv, NULL);
    
    // Parse command line arguments
    if (!parse_command_line(argc, argv)) {
        return EXIT_FAILURE;
    }
    
    // Initialize GTK for UI tests
    gtk_init(&argc, &argv);
    
    // Setup test environment
    setup_test_environment();
    
    // Print test header
    print_test_header();
    
    // Register test suites
    test_log("Registering test suites...\n");
    
    // Network tests
    if (!g_test_config.test_filter || g_str_has_prefix("network", g_test_config.test_filter)) {
        test_log("- Network tests\n");
        register_network_tests();
    }
    
    // Storage tests
    if (!g_test_config.test_filter || g_str_has_prefix("storage", g_test_config.test_filter)) {
        test_log("- Storage tests\n");
        register_storage_tests();
    }
    
    // System integration tests
    if (!g_test_config.test_filter || g_str_has_prefix("system", g_test_config.test_filter)) {
        test_log("- System integration tests\n");
        register_system_tests();
    }
    
    // Graphics tests
    if (!g_test_config.test_filter || g_str_has_prefix("graphics", g_test_config.test_filter)) {
        test_log("- Graphics tests\n");
        register_graphics_tests();
    }
    
    // Audio tests
    if (!g_test_config.test_filter || g_str_has_prefix("audio", g_test_config.test_filter)) {
        test_log("- Audio tests\n");
        register_audio_tests();
    }
    
    // UI tests
    if (!g_test_config.test_filter || g_str_has_prefix("ui", g_test_config.test_filter)) {
        test_log("- UI tests\n");
        register_ui_tests();
    }
    
    test_log("\nRunning tests...\n");
    test_log("================\n\n");
    
    // Record start time
    GTimer *timer = g_timer_new();
    
    // Run tests
    gint result = g_test_run();
    
    // Record end time
    g_test_stats.total_time = g_timer_elapsed(timer, NULL);
    g_timer_destroy(timer);
    
    // Print test summary
    print_test_summary();
    
    // Cleanup test environment
    cleanup_test_environment();
    
    return result;
}

static void print_test_header(void) {
    test_log("=================================================\n");
    test_log("         Taishang Desktop App Test Suite        \n");
    test_log("=================================================\n");
    test_log("Build: %s %s\n", __DATE__, __TIME__);
    test_log("GLib version: %d.%d.%d\n", 
             GLIB_MAJOR_VERSION, GLIB_MINOR_VERSION, GLIB_MICRO_VERSION);
    test_log("GTK version: %d.%d.%d\n", 
             GTK_MAJOR_VERSION, GTK_MINOR_VERSION, GTK_MICRO_VERSION);
    
    if (g_test_config.test_filter) {
        test_log("Filter: %s\n", g_test_config.test_filter);
    }
    
    if (g_test_config.quick) {
        test_log("Mode: Quick tests only\n");
    }
    
    test_log("=================================================\n\n");
}

static void print_test_summary(void) {
    test_log("\n=================================================\n");
    test_log("                 Test Summary                    \n");
    test_log("=================================================\n");
    test_log("Total tests:   %d\n", g_test_stats.total_tests);
    test_log("Passed:        %d\n", g_test_stats.passed_tests);
    test_log("Failed:        %d\n", g_test_stats.failed_tests);
    test_log("Skipped:       %d\n", g_test_stats.skipped_tests);
    test_log("Total time:    %.2f seconds\n", g_test_stats.total_time);
    
    if (g_test_stats.total_tests > 0) {
        gdouble pass_rate = (gdouble)g_test_stats.passed_tests / g_test_stats.total_tests * 100.0;
        test_log("Pass rate:     %.1f%%\n", pass_rate);
    }
    
    test_log("=================================================\n");
    
    if (g_test_stats.failed_tests > 0) {
        test_log("\n❌ Some tests failed. Check the output above for details.\n");
    } else {
        test_log("\n✅ All tests passed!\n");
    }
}

static void setup_test_environment(void) {
    // Create test directories
    g_mkdir_with_parents("test_data", 0755);
    g_mkdir_with_parents("test_output", 0755);
    g_mkdir_with_parents("test_cache", 0755);
    
    // Set environment variables for testing
    g_setenv("TAISHANG_TEST_MODE", "1", TRUE);
    g_setenv("TAISHANG_DATA_DIR", "test_data", TRUE);
    g_setenv("TAISHANG_CACHE_DIR", "test_cache", TRUE);
    g_setenv("TAISHANG_CONFIG_DIR", "test_data", TRUE);
    
    // Open log file if specified
    if (g_test_config.output_file) {
        g_log_file = fopen(g_test_config.output_file, "w");
        if (!g_log_file) {
            g_warning("Failed to open log file: %s", g_test_config.output_file);
        }
    }
    
    // Set up test logging
    g_log_set_default_handler((GLogFunc)g_log_default_handler, NULL);
    
    test_log("Test environment initialized\n");
}

static void cleanup_test_environment(void) {
    // Remove test directories
    g_rmdir("test_cache");
    g_rmdir("test_output");
    g_rmdir("test_data");
    
    // Close log file
    if (g_log_file) {
        fclose(g_log_file);
        g_log_file = NULL;
    }
    
    // Unset environment variables
    g_unsetenv("TAISHANG_TEST_MODE");
    g_unsetenv("TAISHANG_DATA_DIR");
    g_unsetenv("TAISHANG_CACHE_DIR");
    g_unsetenv("TAISHANG_CONFIG_DIR");
    
    test_log("Test environment cleaned up\n");
}

static gboolean parse_command_line(int argc, char **argv) {
    GOptionContext *context;
    GError *error = NULL;
    
    GOptionEntry entries[] = {
        { "verbose", 'v', 0, G_OPTION_ARG_NONE, &g_test_config.verbose,
          "Enable verbose output", NULL },
        { "quick", 'q', 0, G_OPTION_ARG_NONE, &g_test_config.quick,
          "Run only quick tests", NULL },
        { "filter", 'f', 0, G_OPTION_ARG_STRING, &g_test_config.test_filter,
          "Filter tests by name pattern", "PATTERN" },
        { "output", 'o', 0, G_OPTION_ARG_FILENAME, &g_test_config.output_file,
          "Write output to file", "FILE" },
        { NULL }
    };
    
    context = g_option_context_new("- Taishang Desktop App Test Suite");
    g_option_context_add_main_entries(context, entries, NULL);
    g_option_context_add_group(context, gtk_get_option_group(TRUE));
    
    if (!g_option_context_parse(context, &argc, &argv, &error)) {
        g_print("Option parsing failed: %s\n", error->message);
        g_error_free(error);
        g_option_context_free(context);
        return FALSE;
    }
    
    g_option_context_free(context);
    return TRUE;
}

static void print_usage(const char *program_name) {
    g_print("Usage: %s [OPTIONS]\n", program_name);
    g_print("\nOptions:\n");
    g_print("  -v, --verbose     Enable verbose output\n");
    g_print("  -q, --quick       Run only quick tests\n");
    g_print("  -f, --filter=PATTERN  Filter tests by name pattern\n");
    g_print("  -o, --output=FILE Write output to file\n");
    g_print("  -h, --help        Show this help message\n");
    g_print("\nExamples:\n");
    g_print("  %s                    # Run all tests\n", program_name);
    g_print("  %s --quick            # Run only quick tests\n", program_name);
    g_print("  %s --filter=network   # Run only network tests\n", program_name);
    g_print("  %s --output=test.log  # Write output to file\n", program_name);
}

static void test_log(const char *format, ...) {
    va_list args;
    
    // Print to stdout
    if (g_test_config.verbose || !g_test_quiet()) {
        va_start(args, format);
        vprintf(format, args);
        va_end(args);
        fflush(stdout);
    }
    
    // Write to log file
    if (g_log_file) {
        va_start(args, format);
        vfprintf(g_log_file, format, args);
        va_end(args);
        fflush(g_log_file);
    }
}

static void test_log_result(const char *test_name, gboolean passed, gdouble time) {
    g_test_stats.total_tests++;
    
    if (passed) {
        g_test_stats.passed_tests++;
        test_log("✅ %s (%.3fs)\n", test_name, time);
    } else {
        g_test_stats.failed_tests++;
        test_log("❌ %s (%.3fs)\n", test_name, time);
    }
}

// Test fixture helpers
void test_fixture_setup(void) {
    // Common setup for all tests
    g_test_bug_base("https://github.com/taishang/desktop-app/issues/");
}

void test_fixture_teardown(void) {
    // Common teardown for all tests
}

// Memory leak detection helpers
#ifdef ENABLE_MEMORY_TESTING
static gsize initial_memory = 0;

void test_memory_start(void) {
    initial_memory = g_mem_profile_get_current_memory();
}

void test_memory_end(const char *test_name) {
    gsize final_memory = g_mem_profile_get_current_memory();
    gsize leaked = final_memory - initial_memory;
    
    if (leaked > 0) {
        g_test_message("Memory leak detected in %s: %zu bytes", test_name, leaked);
    }
}
#else
void test_memory_start(void) {}
void test_memory_end(const char *test_name) { (void)test_name; }
#endif

// Performance testing helpers
typedef struct {
    GTimer *timer;
    const char *operation;
} PerformanceTest;

PerformanceTest *performance_test_start(const char *operation) {
    PerformanceTest *perf = g_new(PerformanceTest, 1);
    perf->timer = g_timer_new();
    perf->operation = operation;
    return perf;
}

void performance_test_end(PerformanceTest *perf, gdouble expected_max_time) {
    gdouble elapsed = g_timer_elapsed(perf->timer, NULL);
    
    if (elapsed > expected_max_time) {
        g_test_message("Performance warning: %s took %.3fs (expected < %.3fs)",
                       perf->operation, elapsed, expected_max_time);
    }
    
    g_timer_destroy(perf->timer);
    g_free(perf);
}

// Async test helpers
typedef struct {
    GMainLoop *loop;
    gboolean completed;
    gpointer result;
} AsyncTestData;

AsyncTestData *async_test_data_new(void) {
    AsyncTestData *data = g_new0(AsyncTestData, 1);
    data->loop = g_main_loop_new(NULL, FALSE);
    return data;
}

void async_test_data_free(AsyncTestData *data) {
    if (data->loop) {
        g_main_loop_unref(data->loop);
    }
    g_free(data);
}

void async_test_complete(AsyncTestData *data, gpointer result) {
    data->completed = TRUE;
    data->result = result;
    g_main_loop_quit(data->loop);
}

gboolean async_test_wait(AsyncTestData *data, guint timeout_ms) {
    if (timeout_ms > 0) {
        g_timeout_add(timeout_ms, (GSourceFunc)g_main_loop_quit, data->loop);
    }
    
    g_main_loop_run(data->loop);
    return data->completed;
}

// Test data generators
gchar *test_create_temp_file(const char *content) {
    gchar *filename = g_build_filename("test_data", "temp_XXXXXX", NULL);
    gint fd = g_mkstemp(filename);
    
    if (fd != -1) {
        if (content) {
            write(fd, content, strlen(content));
        }
        close(fd);
        return filename;
    }
    
    g_free(filename);
    return NULL;
}

void test_remove_temp_file(const char *filename) {
    if (filename) {
        g_unlink(filename);
    }
}

// Mock object helpers
typedef struct {
    GHashTable *method_calls;
    GHashTable *return_values;
} MockObject;

MockObject *mock_object_new(void) {
    MockObject *mock = g_new0(MockObject, 1);
    mock->method_calls = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, NULL);
    mock->return_values = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, g_free);
    return mock;
}

void mock_object_free(MockObject *mock) {
    g_hash_table_destroy(mock->method_calls);
    g_hash_table_destroy(mock->return_values);
    g_free(mock);
}

void mock_object_set_return_value(MockObject *mock, const char *method, gpointer value) {
    g_hash_table_insert(mock->return_values, g_strdup(method), value);
}

gpointer mock_object_get_return_value(MockObject *mock, const char *method) {
    return g_hash_table_lookup(mock->return_values, method);
}

void mock_object_record_call(MockObject *mock, const char *method) {
    gint count = GPOINTER_TO_INT(g_hash_table_lookup(mock->method_calls, method));
    g_hash_table_insert(mock->method_calls, g_strdup(method), GINT_TO_POINTER(count + 1));
}

gint mock_object_get_call_count(MockObject *mock, const char *method) {
    return GPOINTER_TO_INT(g_hash_table_lookup(mock->method_calls, method));
}