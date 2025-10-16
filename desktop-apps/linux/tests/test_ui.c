/**
 * @file test_ui.c
 * @brief Unit tests for user interface functionality
 * @author TaishangLaojun Development Team
 * @date 2024
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <gtk/gtk.h>

#include "ui.h"
#include "app.h"

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

// Global UI context for testing
static UIContext *test_ui_ctx = NULL;

// Setup function for UI tests
static void setup_ui_test(void) {
    // Initialize GTK for testing (without display)
    gtk_init_check(NULL, NULL);
    
    AppContext *app_ctx = app_init();
    test_ui_ctx = ui_init(app_ctx);
}

// Cleanup function for UI tests
static void cleanup_ui_test(void) {
    if (test_ui_ctx) {
        ui_cleanup(test_ui_ctx);
        test_ui_ctx = NULL;
    }
}

// Test UI initialization
static int test_ui_init(void) {
    setup_ui_test();
    
    TEST_ASSERT(test_ui_ctx != NULL, "UI context should be initialized");
    TEST_ASSERT(test_ui_ctx->main_window != NULL, "Main window should be created");
    TEST_ASSERT(test_ui_ctx->app_context != NULL, "App context should be set");
    
    cleanup_ui_test();
    return 1;
}

// Test main window creation
static int test_main_window_creation(void) {
    setup_ui_test();
    
    GtkWidget *window = test_ui_ctx->main_window;
    TEST_ASSERT(GTK_IS_WINDOW(window), "Main window should be a GTK window");
    
    const char *title = gtk_window_get_title(GTK_WINDOW(window));
    TEST_ASSERT(title != NULL, "Window title should be set");
    TEST_ASSERT(strlen(title) > 0, "Window title should not be empty");
    
    cleanup_ui_test();
    return 1;
}

// Test window properties
static int test_window_properties(void) {
    setup_ui_test();
    
    GtkWidget *window = test_ui_ctx->main_window;
    
    // Test window size
    int width, height;
    gtk_window_get_default_size(GTK_WINDOW(window), &width, &height);
    TEST_ASSERT(width > 0, "Window width should be positive");
    TEST_ASSERT(height > 0, "Window height should be positive");
    
    // Test window position
    GtkWindowPosition position = gtk_window_get_position(GTK_WINDOW(window));
    TEST_ASSERT(position == GTK_WIN_POS_CENTER || position == GTK_WIN_POS_CENTER_ALWAYS,
                "Window should be centered");
    
    cleanup_ui_test();
    return 1;
}

// Test menu bar creation
static int test_menu_bar_creation(void) {
    setup_ui_test();
    
    GtkWidget *menubar = ui_create_menu_bar(test_ui_ctx);
    TEST_ASSERT(menubar != NULL, "Menu bar should be created");
    TEST_ASSERT(GTK_IS_MENU_BAR(menubar), "Should be a GTK menu bar");
    
    cleanup_ui_test();
    return 1;
}

// Test toolbar creation
static int test_toolbar_creation(void) {
    setup_ui_test();
    
    GtkWidget *toolbar = ui_create_toolbar(test_ui_ctx);
    TEST_ASSERT(toolbar != NULL, "Toolbar should be created");
    TEST_ASSERT(GTK_IS_TOOLBAR(toolbar), "Should be a GTK toolbar");
    
    cleanup_ui_test();
    return 1;
}

// Test status bar creation
static int test_status_bar_creation(void) {
    setup_ui_test();
    
    GtkWidget *statusbar = ui_create_status_bar(test_ui_ctx);
    TEST_ASSERT(statusbar != NULL, "Status bar should be created");
    TEST_ASSERT(GTK_IS_STATUSBAR(statusbar), "Should be a GTK status bar");
    
    cleanup_ui_test();
    return 1;
}

// Test chat area creation
static int test_chat_area_creation(void) {
    setup_ui_test();
    
    GtkWidget *chat_area = ui_create_chat_area(test_ui_ctx);
    TEST_ASSERT(chat_area != NULL, "Chat area should be created");
    TEST_ASSERT(GTK_IS_WIDGET(chat_area), "Should be a GTK widget");
    
    cleanup_ui_test();
    return 1;
}

// Test sidebar creation
static int test_sidebar_creation(void) {
    setup_ui_test();
    
    GtkWidget *sidebar = ui_create_sidebar(test_ui_ctx);
    TEST_ASSERT(sidebar != NULL, "Sidebar should be created");
    TEST_ASSERT(GTK_IS_WIDGET(sidebar), "Should be a GTK widget");
    
    cleanup_ui_test();
    return 1;
}

// Test dialog creation
static int test_dialog_creation(void) {
    setup_ui_test();
    
    // Test preferences dialog
    GtkWidget *prefs_dialog = ui_create_preferences_dialog(test_ui_ctx);
    TEST_ASSERT(prefs_dialog != NULL, "Preferences dialog should be created");
    TEST_ASSERT(GTK_IS_DIALOG(prefs_dialog), "Should be a GTK dialog");
    
    // Test about dialog
    GtkWidget *about_dialog = ui_create_about_dialog(test_ui_ctx);
    TEST_ASSERT(about_dialog != NULL, "About dialog should be created");
    TEST_ASSERT(GTK_IS_ABOUT_DIALOG(about_dialog), "Should be a GTK about dialog");
    
    cleanup_ui_test();
    return 1;
}

// Test theme management
static int test_theme_management(void) {
    setup_ui_test();
    
    // Test theme loading
    int result = ui_load_theme(test_ui_ctx, "default");
    TEST_ASSERT(result == 0, "Default theme should load successfully");
    
    // Test theme switching
    result = ui_set_theme(test_ui_ctx, "dark");
    TEST_ASSERT(result == 0, "Theme switching should succeed");
    
    const char *current_theme = ui_get_current_theme(test_ui_ctx);
    TEST_ASSERT(current_theme != NULL, "Current theme should be retrievable");
    
    cleanup_ui_test();
    return 1;
}

// Test UI event handling
static int test_ui_event_handling(void) {
    setup_ui_test();
    
    // Test signal connection
    int result = ui_connect_signals(test_ui_ctx);
    TEST_ASSERT(result == 0, "Signal connection should succeed");
    
    // Test event handler registration
    result = ui_register_event_handlers(test_ui_ctx);
    TEST_ASSERT(result == 0, "Event handler registration should succeed");
    
    cleanup_ui_test();
    return 1;
}

// Test UI state management
static int test_ui_state_management(void) {
    setup_ui_test();
    
    // Test initial state
    UIState state = ui_get_state(test_ui_ctx);
    TEST_ASSERT(state == UI_STATE_INITIALIZED, "Initial state should be INITIALIZED");
    
    // Test state transitions
    ui_set_state(test_ui_ctx, UI_STATE_READY);
    state = ui_get_state(test_ui_ctx);
    TEST_ASSERT(state == UI_STATE_READY, "State should be READY");
    
    cleanup_ui_test();
    return 1;
}

// Test UI responsiveness
static int test_ui_responsiveness(void) {
    setup_ui_test();
    
    // Test window resizing
    gtk_window_resize(GTK_WINDOW(test_ui_ctx->main_window), 800, 600);
    
    int width, height;
    gtk_window_get_size(GTK_WINDOW(test_ui_ctx->main_window), &width, &height);
    TEST_ASSERT(width == 800, "Window width should be updated");
    TEST_ASSERT(height == 600, "Window height should be updated");
    
    cleanup_ui_test();
    return 1;
}

// Test UI accessibility
static int test_ui_accessibility(void) {
    setup_ui_test();
    
    // Test accessibility features
    int result = ui_setup_accessibility(test_ui_ctx);
    TEST_ASSERT(result == 0, "Accessibility setup should succeed");
    
    // Test keyboard navigation
    result = ui_enable_keyboard_navigation(test_ui_ctx);
    TEST_ASSERT(result == 0, "Keyboard navigation should be enabled");
    
    cleanup_ui_test();
    return 1;
}

// Test UI cleanup
static int test_ui_cleanup(void) {
    setup_ui_test();
    
    // Test proper cleanup
    ui_cleanup(test_ui_ctx);
    test_ui_ctx = NULL;
    
    // Verify cleanup was successful
    TEST_ASSERT(test_ui_ctx == NULL, "UI context should be cleaned up");
    
    return 1;
}

// Main test runner
int main(int argc, char *argv[]) {
    printf("=== TaishangLaojun UI Tests ===\n\n");
    
    // Initialize GTK for testing
    gtk_init(&argc, &argv);
    
    RUN_TEST(test_ui_init);
    RUN_TEST(test_main_window_creation);
    RUN_TEST(test_window_properties);
    RUN_TEST(test_menu_bar_creation);
    RUN_TEST(test_toolbar_creation);
    RUN_TEST(test_status_bar_creation);
    RUN_TEST(test_chat_area_creation);
    RUN_TEST(test_sidebar_creation);
    RUN_TEST(test_dialog_creation);
    RUN_TEST(test_theme_management);
    RUN_TEST(test_ui_event_handling);
    RUN_TEST(test_ui_state_management);
    RUN_TEST(test_ui_responsiveness);
    RUN_TEST(test_ui_accessibility);
    RUN_TEST(test_ui_cleanup);
    
    printf("\n=== All UI Tests Passed! ===\n");
    return 0;
}