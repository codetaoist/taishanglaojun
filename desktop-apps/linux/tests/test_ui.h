#ifndef TEST_UI_H
#define TEST_UI_H

#include <glib.h>
#include <gtk/gtk.h>

G_BEGIN_DECLS

// Test registration function
void register_ui_tests(void);

// Main window tests
void test_ui_main_window_creation(void);
void test_ui_main_window_initialization(void);
void test_ui_main_window_destruction(void);
void test_ui_main_window_show_hide(void);
void test_ui_main_window_resize(void);
void test_ui_main_window_minimize_restore(void);

// Chat interface tests
void test_ui_chat_interface_creation(void);
void test_ui_chat_message_display(void);
void test_ui_chat_message_input(void);
void test_ui_chat_message_sending(void);
void test_ui_chat_history_loading(void);
void test_ui_chat_scrolling(void);
void test_ui_chat_emoji_support(void);
void test_ui_chat_file_attachments(void);

// Project management UI tests
void test_ui_project_list_display(void);
void test_ui_project_creation_dialog(void);
void test_ui_project_editing_dialog(void);
void test_ui_project_deletion_confirmation(void);
void test_ui_project_file_browser(void);
void test_ui_project_search_functionality(void);

// Settings dialog tests
void test_ui_settings_dialog_creation(void);
void test_ui_settings_general_tab(void);
void test_ui_settings_appearance_tab(void);
void test_ui_settings_network_tab(void);
void test_ui_settings_audio_tab(void);
void test_ui_settings_advanced_tab(void);
void test_ui_settings_save_apply(void);

// Friend management UI tests
void test_ui_friend_list_display(void);
void test_ui_friend_add_dialog(void);
void test_ui_friend_profile_dialog(void);
void test_ui_friend_remove_confirmation(void);
void test_ui_friend_status_updates(void);

// Notification system tests
void test_ui_notification_display(void);
void test_ui_notification_positioning(void);
void test_ui_notification_timeout(void);
void test_ui_notification_interaction(void);
void test_ui_notification_queue_management(void);

// Theme and styling tests
void test_ui_theme_loading(void);
void test_ui_theme_switching(void);
void test_ui_dark_mode_toggle(void);
void test_ui_custom_css_loading(void);
void test_ui_font_scaling(void);

// Widget tests
void test_ui_custom_widgets(void);
void test_ui_widget_styling(void);
void test_ui_widget_events(void);
void test_ui_widget_accessibility(void);

// Layout tests
void test_ui_responsive_layout(void);
void test_ui_window_state_persistence(void);
void test_ui_panel_resizing(void);
void test_ui_toolbar_customization(void);

// Input handling tests
void test_ui_keyboard_shortcuts(void);
void test_ui_mouse_interactions(void);
void test_ui_touch_support(void);
void test_ui_drag_and_drop(void);

// Menu system tests
void test_ui_main_menu_creation(void);
void test_ui_context_menus(void);
void test_ui_menu_actions(void);
void test_ui_menu_accelerators(void);

// Dialog tests
void test_ui_modal_dialogs(void);
void test_ui_file_chooser_dialogs(void);
void test_ui_message_dialogs(void);
void test_ui_progress_dialogs(void);

// Accessibility tests
void test_ui_screen_reader_support(void);
void test_ui_keyboard_navigation(void);
void test_ui_high_contrast_support(void);
void test_ui_font_size_scaling(void);

// Performance tests
void test_ui_rendering_performance(void);
void test_ui_memory_usage(void);
void test_ui_startup_time(void);
void test_ui_large_data_handling(void);

// Integration tests
void test_ui_network_integration(void);
void test_ui_storage_integration(void);
void test_ui_audio_integration(void);
void test_ui_system_integration(void);

// Error handling tests
void test_ui_error_dialogs(void);
void test_ui_network_error_handling(void);
void test_ui_file_error_handling(void);
void test_ui_graceful_degradation(void);

// Internationalization tests
void test_ui_locale_support(void);
void test_ui_text_direction(void);
void test_ui_date_time_formatting(void);
void test_ui_number_formatting(void);

// Mock UI components
typedef struct _MockUIComponent MockUIComponent;

MockUIComponent *mock_ui_component_new(const char *name, GType widget_type);
void mock_ui_component_free(MockUIComponent *component);
void mock_ui_component_set_property(MockUIComponent *component, const char *property, const GValue *value);
void mock_ui_component_emit_signal(MockUIComponent *component, const char *signal_name);
GtkWidget *mock_ui_component_get_widget(MockUIComponent *component);

// Test helpers
void test_ui_setup_environment(void);
void test_ui_cleanup_environment(void);
GtkWidget *test_ui_create_test_window(void);
void test_ui_destroy_test_window(GtkWidget *window);
gboolean test_ui_wait_for_events(guint timeout_ms);
void test_ui_simulate_key_press(GtkWidget *widget, guint keyval, GdkModifierType modifiers);
void test_ui_simulate_mouse_click(GtkWidget *widget, gdouble x, gdouble y, guint button);

// UI state verification
gboolean test_ui_verify_widget_visible(GtkWidget *widget);
gboolean test_ui_verify_widget_sensitive(GtkWidget *widget);
gboolean test_ui_verify_text_content(GtkWidget *widget, const char *expected_text);
gboolean test_ui_verify_widget_style(GtkWidget *widget, const char *css_class);

// Event simulation
void test_ui_simulate_window_resize(GtkWidget *window, gint width, gint height);
void test_ui_simulate_window_close(GtkWidget *window);
void test_ui_simulate_text_input(GtkWidget *entry, const char *text);
void test_ui_simulate_button_click(GtkWidget *button);
void test_ui_simulate_menu_activation(GtkWidget *menu_item);

// Performance measurement
typedef struct {
    gdouble render_time;
    gdouble event_processing_time;
    gsize memory_usage;
    guint32 widget_count;
    guint32 signal_emissions;
} UIPerformanceMetrics;

UIPerformanceMetrics test_ui_measure_performance(void (*test_func)(void));
void test_ui_log_performance(const char *test_name, UIPerformanceMetrics metrics);

// Screenshot and visual testing
gboolean test_ui_capture_screenshot(GtkWidget *widget, const char *filename);
gboolean test_ui_compare_screenshots(const char *expected_file, const char *actual_file, gdouble tolerance);
void test_ui_generate_reference_screenshots(void);

// Stress testing
void test_ui_stress_widget_creation(gint widget_count);
void test_ui_stress_event_processing(gint event_count);
void test_ui_stress_memory_allocation(gint allocation_count);

G_END_DECLS

#endif // TEST_UI_H