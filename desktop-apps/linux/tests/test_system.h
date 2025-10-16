#ifndef TEST_SYSTEM_H
#define TEST_SYSTEM_H

#include <glib.h>

G_BEGIN_DECLS

// Test registration function
void register_system_tests(void);

// D-Bus client tests
void test_dbus_client_init(void);
void test_dbus_client_cleanup(void);
void test_dbus_notifications(void);
void test_dbus_screensaver(void);
void test_dbus_power_management(void);
void test_dbus_network_monitoring(void);
void test_dbus_signal_handling(void);
void test_dbus_error_handling(void);

// Desktop integration tests
void test_desktop_integration_init(void);
void test_desktop_integration_cleanup(void);
void test_desktop_file_creation(void);
void test_autostart_management(void);
void test_system_tray_operations(void);
void test_desktop_environment_detection(void);
void test_file_associations(void);
void test_url_handling(void);
void test_window_management(void);

// System information tests
void test_system_info_collection(void);
void test_hardware_detection(void);
void test_os_version_detection(void);
void test_display_information(void);
void test_audio_device_detection(void);
void test_network_interface_detection(void);

// Process management tests
void test_process_spawning(void);
void test_process_monitoring(void);
void test_process_termination(void);
void test_process_communication(void);

// File system monitoring tests
void test_file_system_watching(void);
void test_directory_monitoring(void);
void test_file_change_detection(void);
void test_mount_point_monitoring(void);

// Security tests
void test_permission_checking(void);
void test_sandbox_compliance(void);
void test_secure_storage(void);
void test_credential_management(void);

// Performance tests
void test_system_resource_usage(void);
void test_memory_management(void);
void test_cpu_usage_monitoring(void);
void test_disk_usage_monitoring(void);

// Error handling tests
void test_system_error_recovery(void);
void test_service_unavailable_handling(void);
void test_permission_denied_handling(void);
void test_resource_exhaustion_handling(void);

// Integration tests
void test_full_system_integration(void);
void test_multi_user_support(void);
void test_session_management(void);
void test_startup_shutdown_sequence(void);

// Mock system services
typedef struct _MockDBusService MockDBusService;
typedef struct _MockSystemService MockSystemService;

MockDBusService *mock_dbus_service_new(const char *service_name);
void mock_dbus_service_free(MockDBusService *service);
void mock_dbus_service_start(MockDBusService *service);
void mock_dbus_service_stop(MockDBusService *service);
void mock_dbus_service_add_method(MockDBusService *service, const char *method, const char *response);
void mock_dbus_service_emit_signal(MockDBusService *service, const char *signal, const char *data);

MockSystemService *mock_system_service_new(const char *service_name);
void mock_system_service_free(MockSystemService *service);
void mock_system_service_set_available(MockSystemService *service, gboolean available);
void mock_system_service_set_response_delay(MockSystemService *service, guint delay_ms);

// Test environment helpers
void test_system_setup_environment(void);
void test_system_cleanup_environment(void);
void test_system_create_temp_desktop_file(const char *name, const char *content);
void test_system_remove_temp_desktop_file(const char *name);
void test_system_simulate_user_session(void);
void test_system_simulate_system_shutdown(void);

// System state verification
gboolean test_system_verify_desktop_file_exists(const char *name);
gboolean test_system_verify_autostart_enabled(const char *name);
gboolean test_system_verify_system_tray_visible(void);
gboolean test_system_verify_notification_sent(const char *title);
gboolean test_system_verify_dbus_service_running(const char *service_name);

// Performance measurement
typedef struct {
    gdouble cpu_usage;
    gsize memory_usage;
    gsize disk_usage;
    gint file_descriptors;
    gint thread_count;
} SystemResourceUsage;

SystemResourceUsage test_system_measure_resource_usage(void);
void test_system_log_resource_usage(const char *test_name, SystemResourceUsage usage);

G_END_DECLS

#endif // TEST_SYSTEM_H