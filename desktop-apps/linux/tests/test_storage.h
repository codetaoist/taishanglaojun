#ifndef TEST_STORAGE_H
#define TEST_STORAGE_H

#include <glib.h>

G_BEGIN_DECLS

// Test registration function
void register_storage_tests(void);

// Database tests
void test_database_init(void);
void test_database_cleanup(void);
void test_database_user_operations(void);
void test_database_message_operations(void);
void test_database_project_operations(void);
void test_database_file_operations(void);
void test_database_friend_operations(void);
void test_database_settings_operations(void);
void test_database_transactions(void);
void test_database_concurrent_access(void);
void test_database_migration(void);
void test_database_backup_restore(void);

// Cache tests
void test_cache_init(void);
void test_cache_cleanup(void);
void test_cache_set_get(void);
void test_cache_json_operations(void);
void test_cache_expiry(void);
void test_cache_lru_eviction(void);
void test_cache_size_limits(void);
void test_cache_statistics(void);
void test_cache_thread_safety(void);
void test_cache_persistence(void);

// File system tests
void test_file_operations(void);
void test_directory_operations(void);
void test_file_permissions(void);
void test_file_watching(void);
void test_file_compression(void);
void test_file_encryption(void);

// Performance tests
void test_database_performance(void);
void test_cache_performance(void);
void test_file_io_performance(void);

// Error handling tests
void test_database_error_handling(void);
void test_cache_error_handling(void);
void test_file_error_handling(void);

// Data integrity tests
void test_database_integrity(void);
void test_cache_integrity(void);
void test_file_integrity(void);

// Test data helpers
typedef struct {
    gchar *username;
    gchar *email;
    gchar *password_hash;
    gint status;
} TestUser;

typedef struct {
    gchar *content;
    gchar *sender;
    gchar *receiver;
    gint64 timestamp;
    gint type;
} TestMessage;

typedef struct {
    gchar *name;
    gchar *description;
    gchar *owner;
    gint64 created_at;
    gint status;
} TestProject;

TestUser *test_user_new(const char *username, const char *email);
void test_user_free(TestUser *user);

TestMessage *test_message_new(const char *content, const char *sender, const char *receiver);
void test_message_free(TestMessage *message);

TestProject *test_project_new(const char *name, const char *description, const char *owner);
void test_project_free(TestProject *project);

// Database test helpers
gboolean test_database_create_temp(void);
void test_database_cleanup_temp(void);
gboolean test_database_populate_sample_data(void);
void test_database_clear_all_data(void);

// Cache test helpers
void test_cache_fill_with_data(gint count);
void test_cache_verify_data(gint count);
void test_cache_stress_test(gint iterations);

G_END_DECLS

#endif // TEST_STORAGE_H