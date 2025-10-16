#ifndef TAISHANG_DATABASE_H
#define TAISHANG_DATABASE_H

#include <glib.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangDatabase TaishangDatabase;

// Data structures
typedef struct {
    int id;
    char *username;
    char *email;
    char *display_name;
    char *avatar_url;
    char *status;
    gint64 last_seen;
    gint64 created_at;
    gint64 updated_at;
} TaishangUser;

typedef struct {
    int id;
    int sender_id;
    int recipient_id;
    char *content;
    char *message_type;
    gint64 timestamp;
    gboolean read_status;
} TaishangMessage;

typedef struct {
    int id;
    char *name;
    char *description;
    char *project_type;
    int owner_id;
    char *status;
    gint64 created_at;
    gint64 updated_at;
} TaishangProject;

typedef struct {
    int id;
    char *filename;
    char *file_path;
    gint64 file_size;
    char *mime_type;
    int owner_id;
    int project_id;
    gint64 upload_date;
} TaishangFile;

typedef struct {
    int id;
    int user_id;
    int friend_id;
    char *status;
    gint64 created_at;
} TaishangFriend;

// Initialization and cleanup
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

G_END_DECLS

#endif // TAISHANG_DATABASE_H