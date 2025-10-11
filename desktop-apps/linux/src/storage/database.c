#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sqlite3.h>
#include <glib.h>
#include <json-c/json.h>
#include "../../include/storage/database.h"

// Database structure
typedef struct {
    sqlite3 *db;
    char *db_path;
    gboolean initialized;
    GMutex mutex;
} TaishangDatabase;

static TaishangDatabase *database = NULL;

// SQL statements for table creation
static const char *CREATE_USERS_TABLE = 
    "CREATE TABLE IF NOT EXISTS users ("
    "id INTEGER PRIMARY KEY AUTOINCREMENT,"
    "username TEXT UNIQUE NOT NULL,"
    "email TEXT UNIQUE NOT NULL,"
    "display_name TEXT,"
    "avatar_url TEXT,"
    "status TEXT DEFAULT 'offline',"
    "last_seen INTEGER,"
    "created_at INTEGER DEFAULT (strftime('%s', 'now')),"
    "updated_at INTEGER DEFAULT (strftime('%s', 'now'))"
    ");";

static const char *CREATE_MESSAGES_TABLE = 
    "CREATE TABLE IF NOT EXISTS messages ("
    "id INTEGER PRIMARY KEY AUTOINCREMENT,"
    "sender_id INTEGER NOT NULL,"
    "recipient_id INTEGER NOT NULL,"
    "content TEXT NOT NULL,"
    "message_type TEXT DEFAULT 'text',"
    "timestamp INTEGER DEFAULT (strftime('%s', 'now')),"
    "read_status INTEGER DEFAULT 0,"
    "FOREIGN KEY (sender_id) REFERENCES users (id),"
    "FOREIGN KEY (recipient_id) REFERENCES users (id)"
    ");";

static const char *CREATE_PROJECTS_TABLE = 
    "CREATE TABLE IF NOT EXISTS projects ("
    "id INTEGER PRIMARY KEY AUTOINCREMENT,"
    "name TEXT NOT NULL,"
    "description TEXT,"
    "project_type TEXT DEFAULT 'general',"
    "owner_id INTEGER NOT NULL,"
    "status TEXT DEFAULT 'active',"
    "created_at INTEGER DEFAULT (strftime('%s', 'now')),"
    "updated_at INTEGER DEFAULT (strftime('%s', 'now')),"
    "FOREIGN KEY (owner_id) REFERENCES users (id)"
    ");";

static const char *CREATE_FILES_TABLE = 
    "CREATE TABLE IF NOT EXISTS files ("
    "id INTEGER PRIMARY KEY AUTOINCREMENT,"
    "filename TEXT NOT NULL,"
    "file_path TEXT NOT NULL,"
    "file_size INTEGER,"
    "mime_type TEXT,"
    "owner_id INTEGER NOT NULL,"
    "project_id INTEGER,"
    "upload_date INTEGER DEFAULT (strftime('%s', 'now')),"
    "FOREIGN KEY (owner_id) REFERENCES users (id),"
    "FOREIGN KEY (project_id) REFERENCES projects (id)"
    ");";

static const char *CREATE_FRIENDS_TABLE = 
    "CREATE TABLE IF NOT EXISTS friends ("
    "id INTEGER PRIMARY KEY AUTOINCREMENT,"
    "user_id INTEGER NOT NULL,"
    "friend_id INTEGER NOT NULL,"
    "status TEXT DEFAULT 'pending',"
    "created_at INTEGER DEFAULT (strftime('%s', 'now')),"
    "FOREIGN KEY (user_id) REFERENCES users (id),"
    "FOREIGN KEY (friend_id) REFERENCES users (id),"
    "UNIQUE(user_id, friend_id)"
    ");";

static const char *CREATE_SETTINGS_TABLE = 
    "CREATE TABLE IF NOT EXISTS settings ("
    "id INTEGER PRIMARY KEY AUTOINCREMENT,"
    "key TEXT UNIQUE NOT NULL,"
    "value TEXT,"
    "updated_at INTEGER DEFAULT (strftime('%s', 'now'))"
    ");";

// Forward declarations
static gboolean create_tables(void);
static int callback_count(void *data, int argc, char **argv, char **azColName);
static int callback_single_row(void *data, int argc, char **argv, char **azColName);
static int callback_multiple_rows(void *data, int argc, char **argv, char **azColName);

// Public functions
gboolean taishang_database_init(const char *db_path);
void taishang_database_cleanup(void);
TaishangDatabase *taishang_database_get_instance(void);

// User functions
gboolean taishang_database_save_user(const TaishangUser *user);
TaishangUser *taishang_database_get_user(int user_id);
TaishangUser *taishang_database_get_user_by_username(const char *username);
GList *taishang_database_get_all_users(void);
gboolean taishang_database_update_user_status(int user_id, const char *status);
gboolean taishang_database_delete_user(int user_id);

// Message functions
gboolean taishang_database_save_message(const TaishangMessage *message);
GList *taishang_database_get_messages(int user1_id, int user2_id, int limit, int offset);
GList *taishang_database_get_recent_conversations(int user_id);
gboolean taishang_database_mark_message_read(int message_id);
gboolean taishang_database_delete_message(int message_id);

// Project functions
gboolean taishang_database_save_project(const TaishangProject *project);
TaishangProject *taishang_database_get_project(int project_id);
GList *taishang_database_get_user_projects(int user_id);
gboolean taishang_database_update_project(const TaishangProject *project);
gboolean taishang_database_delete_project(int project_id);

// File functions
gboolean taishang_database_save_file(const TaishangFile *file);
TaishangFile *taishang_database_get_file(int file_id);
GList *taishang_database_get_user_files(int user_id);
GList *taishang_database_get_project_files(int project_id);
gboolean taishang_database_delete_file(int file_id);

// Friend functions
gboolean taishang_database_add_friend(int user_id, int friend_id);
GList *taishang_database_get_friends(int user_id);
GList *taishang_database_get_friend_requests(int user_id);
gboolean taishang_database_accept_friend_request(int user_id, int friend_id);
gboolean taishang_database_remove_friend(int user_id, int friend_id);

// Settings functions
gboolean taishang_database_set_setting(const char *key, const char *value);
char *taishang_database_get_setting(const char *key);
gboolean taishang_database_delete_setting(const char *key);

// Utility functions
void taishang_user_free(TaishangUser *user);
void taishang_message_free(TaishangMessage *message);
void taishang_project_free(TaishangProject *project);
void taishang_file_free(TaishangFile *file);

// Implementation
gboolean taishang_database_init(const char *db_path) {
    if (database != NULL) {
        g_warning("Database already initialized");
        return FALSE;
    }
    
    database = g_new0(TaishangDatabase, 1);
    g_mutex_init(&database->mutex);
    
    // Set database path
    if (db_path) {
        database->db_path = g_strdup(db_path);
    } else {
        // Use default path in user data directory
        char *data_dir = g_build_filename(g_get_user_data_dir(), "taishang", NULL);
        g_mkdir_with_parents(data_dir, 0755);
        database->db_path = g_build_filename(data_dir, "taishang.db", NULL);
        g_free(data_dir);
    }
    
    // Open database
    int rc = sqlite3_open(database->db_path, &database->db);
    if (rc != SQLITE_OK) {
        g_error("Cannot open database: %s", sqlite3_errmsg(database->db));
        g_free(database->db_path);
        g_free(database);
        database = NULL;
        return FALSE;
    }
    
    // Enable foreign keys
    sqlite3_exec(database->db, "PRAGMA foreign_keys = ON;", NULL, NULL, NULL);
    
    // Create tables
    if (!create_tables()) {
        g_error("Failed to create database tables");
        sqlite3_close(database->db);
        g_free(database->db_path);
        g_free(database);
        database = NULL;
        return FALSE;
    }
    
    database->initialized = TRUE;
    g_print("Database initialized: %s\n", database->db_path);
    return TRUE;
}

void taishang_database_cleanup(void) {
    if (database == NULL) {
        return;
    }
    
    g_mutex_lock(&database->mutex);
    
    if (database->db) {
        sqlite3_close(database->db);
    }
    
    g_free(database->db_path);
    
    g_mutex_unlock(&database->mutex);
    g_mutex_clear(&database->mutex);
    
    g_free(database);
    database = NULL;
    
    g_print("Database cleaned up\n");
}

TaishangDatabase *taishang_database_get_instance(void) {
    return database;
}

static gboolean create_tables(void) {
    char *err_msg = NULL;
    
    // Create tables
    const char *tables[] = {
        CREATE_USERS_TABLE,
        CREATE_MESSAGES_TABLE,
        CREATE_PROJECTS_TABLE,
        CREATE_FILES_TABLE,
        CREATE_FRIENDS_TABLE,
        CREATE_SETTINGS_TABLE,
        NULL
    };
    
    for (int i = 0; tables[i] != NULL; i++) {
        int rc = sqlite3_exec(database->db, tables[i], NULL, NULL, &err_msg);
        if (rc != SQLITE_OK) {
            g_error("SQL error: %s", err_msg);
            sqlite3_free(err_msg);
            return FALSE;
        }
    }
    
    return TRUE;
}

// User functions implementation
gboolean taishang_database_save_user(const TaishangUser *user) {
    if (!database || !user) return FALSE;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "INSERT OR REPLACE INTO users (id, username, email, display_name, avatar_url, status, last_seen) "
                      "VALUES (?, ?, ?, ?, ?, ?, ?);";
    
    sqlite3_stmt *stmt;
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_int(stmt, 1, user->id);
        sqlite3_bind_text(stmt, 2, user->username, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 3, user->email, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 4, user->display_name, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 5, user->avatar_url, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 6, user->status, -1, SQLITE_STATIC);
        sqlite3_bind_int64(stmt, 7, user->last_seen);
        
        rc = sqlite3_step(stmt);
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return (rc == SQLITE_DONE);
}

TaishangUser *taishang_database_get_user(int user_id) {
    if (!database) return NULL;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "SELECT id, username, email, display_name, avatar_url, status, last_seen, created_at "
                      "FROM users WHERE id = ?;";
    
    sqlite3_stmt *stmt;
    TaishangUser *user = NULL;
    
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_int(stmt, 1, user_id);
        
        if (sqlite3_step(stmt) == SQLITE_ROW) {
            user = g_new0(TaishangUser, 1);
            user->id = sqlite3_column_int(stmt, 0);
            user->username = g_strdup((char*)sqlite3_column_text(stmt, 1));
            user->email = g_strdup((char*)sqlite3_column_text(stmt, 2));
            user->display_name = g_strdup((char*)sqlite3_column_text(stmt, 3));
            user->avatar_url = g_strdup((char*)sqlite3_column_text(stmt, 4));
            user->status = g_strdup((char*)sqlite3_column_text(stmt, 5));
            user->last_seen = sqlite3_column_int64(stmt, 6);
            user->created_at = sqlite3_column_int64(stmt, 7);
        }
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return user;
}

TaishangUser *taishang_database_get_user_by_username(const char *username) {
    if (!database || !username) return NULL;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "SELECT id, username, email, display_name, avatar_url, status, last_seen, created_at "
                      "FROM users WHERE username = ?;";
    
    sqlite3_stmt *stmt;
    TaishangUser *user = NULL;
    
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_text(stmt, 1, username, -1, SQLITE_STATIC);
        
        if (sqlite3_step(stmt) == SQLITE_ROW) {
            user = g_new0(TaishangUser, 1);
            user->id = sqlite3_column_int(stmt, 0);
            user->username = g_strdup((char*)sqlite3_column_text(stmt, 1));
            user->email = g_strdup((char*)sqlite3_column_text(stmt, 2));
            user->display_name = g_strdup((char*)sqlite3_column_text(stmt, 3));
            user->avatar_url = g_strdup((char*)sqlite3_column_text(stmt, 4));
            user->status = g_strdup((char*)sqlite3_column_text(stmt, 5));
            user->last_seen = sqlite3_column_int64(stmt, 6);
            user->created_at = sqlite3_column_int64(stmt, 7);
        }
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return user;
}

// Message functions implementation
gboolean taishang_database_save_message(const TaishangMessage *message) {
    if (!database || !message) return FALSE;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "INSERT INTO messages (sender_id, recipient_id, content, message_type, timestamp, read_status) "
                      "VALUES (?, ?, ?, ?, ?, ?);";
    
    sqlite3_stmt *stmt;
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_int(stmt, 1, message->sender_id);
        sqlite3_bind_int(stmt, 2, message->recipient_id);
        sqlite3_bind_text(stmt, 3, message->content, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 4, message->message_type, -1, SQLITE_STATIC);
        sqlite3_bind_int64(stmt, 5, message->timestamp);
        sqlite3_bind_int(stmt, 6, message->read_status ? 1 : 0);
        
        rc = sqlite3_step(stmt);
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return (rc == SQLITE_DONE);
}

GList *taishang_database_get_messages(int user1_id, int user2_id, int limit, int offset) {
    if (!database) return NULL;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "SELECT id, sender_id, recipient_id, content, message_type, timestamp, read_status "
                      "FROM messages "
                      "WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?) "
                      "ORDER BY timestamp DESC "
                      "LIMIT ? OFFSET ?;";
    
    sqlite3_stmt *stmt;
    GList *messages = NULL;
    
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_int(stmt, 1, user1_id);
        sqlite3_bind_int(stmt, 2, user2_id);
        sqlite3_bind_int(stmt, 3, user2_id);
        sqlite3_bind_int(stmt, 4, user1_id);
        sqlite3_bind_int(stmt, 5, limit > 0 ? limit : 50);
        sqlite3_bind_int(stmt, 6, offset);
        
        while (sqlite3_step(stmt) == SQLITE_ROW) {
            TaishangMessage *message = g_new0(TaishangMessage, 1);
            message->id = sqlite3_column_int(stmt, 0);
            message->sender_id = sqlite3_column_int(stmt, 1);
            message->recipient_id = sqlite3_column_int(stmt, 2);
            message->content = g_strdup((char*)sqlite3_column_text(stmt, 3));
            message->message_type = g_strdup((char*)sqlite3_column_text(stmt, 4));
            message->timestamp = sqlite3_column_int64(stmt, 5);
            message->read_status = sqlite3_column_int(stmt, 6) == 1;
            
            messages = g_list_prepend(messages, message);
        }
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return g_list_reverse(messages);
}

// Project functions implementation
gboolean taishang_database_save_project(const TaishangProject *project) {
    if (!database || !project) return FALSE;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "INSERT OR REPLACE INTO projects (id, name, description, project_type, owner_id, status) "
                      "VALUES (?, ?, ?, ?, ?, ?);";
    
    sqlite3_stmt *stmt;
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_int(stmt, 1, project->id);
        sqlite3_bind_text(stmt, 2, project->name, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 3, project->description, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 4, project->project_type, -1, SQLITE_STATIC);
        sqlite3_bind_int(stmt, 5, project->owner_id);
        sqlite3_bind_text(stmt, 6, project->status, -1, SQLITE_STATIC);
        
        rc = sqlite3_step(stmt);
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return (rc == SQLITE_DONE);
}

// Settings functions implementation
gboolean taishang_database_set_setting(const char *key, const char *value) {
    if (!database || !key) return FALSE;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?);";
    
    sqlite3_stmt *stmt;
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_text(stmt, 1, key, -1, SQLITE_STATIC);
        sqlite3_bind_text(stmt, 2, value, -1, SQLITE_STATIC);
        
        rc = sqlite3_step(stmt);
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return (rc == SQLITE_DONE);
}

char *taishang_database_get_setting(const char *key) {
    if (!database || !key) return NULL;
    
    g_mutex_lock(&database->mutex);
    
    const char *sql = "SELECT value FROM settings WHERE key = ?;";
    
    sqlite3_stmt *stmt;
    char *value = NULL;
    
    int rc = sqlite3_prepare_v2(database->db, sql, -1, &stmt, NULL);
    
    if (rc == SQLITE_OK) {
        sqlite3_bind_text(stmt, 1, key, -1, SQLITE_STATIC);
        
        if (sqlite3_step(stmt) == SQLITE_ROW) {
            const char *result = (char*)sqlite3_column_text(stmt, 0);
            if (result) {
                value = g_strdup(result);
            }
        }
    }
    
    sqlite3_finalize(stmt);
    g_mutex_unlock(&database->mutex);
    
    return value;
}

// Utility functions
void taishang_user_free(TaishangUser *user) {
    if (!user) return;
    
    g_free(user->username);
    g_free(user->email);
    g_free(user->display_name);
    g_free(user->avatar_url);
    g_free(user->status);
    g_free(user);
}

void taishang_message_free(TaishangMessage *message) {
    if (!message) return;
    
    g_free(message->content);
    g_free(message->message_type);
    g_free(message);
}

void taishang_project_free(TaishangProject *project) {
    if (!project) return;
    
    g_free(project->name);
    g_free(project->description);
    g_free(project->project_type);
    g_free(project->status);
    g_free(project);
}

void taishang_file_free(TaishangFile *file) {
    if (!file) return;
    
    g_free(file->filename);
    g_free(file->file_path);
    g_free(file->mime_type);
    g_free(file);
}