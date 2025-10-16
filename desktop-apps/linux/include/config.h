/**
 * @file config.h
 * @brief TaishangLaojun Desktop Application Configuration Header
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains the configuration management structures, functions, and definitions
 * for the TaishangLaojun desktop application on Linux.
 */

#ifndef TAISHANG_CONFIG_H
#define TAISHANG_CONFIG_H

#include "common.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Configuration file constants */
#define TAISHANG_CONFIG_FILE_NAME           "config.json"
#define TAISHANG_CONFIG_BACKUP_SUFFIX       ".backup"
#define TAISHANG_CONFIG_TEMP_SUFFIX         ".tmp"
#define TAISHANG_CONFIG_VERSION             "1.0"

/* Configuration groups */
#define TAISHANG_CONFIG_GROUP_GENERAL       "general"
#define TAISHANG_CONFIG_GROUP_UI            "ui"
#define TAISHANG_CONFIG_GROUP_NETWORK       "network"
#define TAISHANG_CONFIG_GROUP_SECURITY      "security"
#define TAISHANG_CONFIG_GROUP_CHAT          "chat"
#define TAISHANG_CONFIG_GROUP_NOTIFICATIONS "notifications"
#define TAISHANG_CONFIG_GROUP_ADVANCED      "advanced"

/* Forward declarations */
typedef struct _TaishangConfig TaishangConfig;
typedef struct _TaishangConfigClass TaishangConfigClass;

/* Configuration value types */
typedef enum {
    TAISHANG_CONFIG_TYPE_BOOLEAN = 0,
    TAISHANG_CONFIG_TYPE_INTEGER,
    TAISHANG_CONFIG_TYPE_DOUBLE,
    TAISHANG_CONFIG_TYPE_STRING,
    TAISHANG_CONFIG_TYPE_STRING_LIST,
    TAISHANG_CONFIG_TYPE_OBJECT
} TaishangConfigType;

/* Configuration error codes */
typedef enum {
    TAISHANG_CONFIG_ERROR_NONE = 0,
    TAISHANG_CONFIG_ERROR_FILE_NOT_FOUND,
    TAISHANG_CONFIG_ERROR_PARSE_FAILED,
    TAISHANG_CONFIG_ERROR_WRITE_FAILED,
    TAISHANG_CONFIG_ERROR_INVALID_TYPE,
    TAISHANG_CONFIG_ERROR_INVALID_KEY,
    TAISHANG_CONFIG_ERROR_PERMISSION_DENIED,
    TAISHANG_CONFIG_ERROR_BACKUP_FAILED,
    TAISHANG_CONFIG_ERROR_VALIDATION_FAILED,
    TAISHANG_CONFIG_ERROR_UNKNOWN
} TaishangConfigError;

/* Configuration validation result */
typedef struct {
    gboolean valid;
    TaishangConfigError error;
    gchar *message;
    gchar *key;
} TaishangConfigValidation;

/* Configuration structure */
struct _TaishangConfig {
    GObject parent;
    
    /* Configuration data */
    JsonObject *root;
    gchar *config_file;
    gchar *config_dir;
    
    /* Configuration state */
    gboolean loaded;
    gboolean modified;
    gboolean auto_save;
    gint64 last_modified;
    
    /* Validation */
    JsonObject *schema;
    gboolean validate_on_load;
    gboolean validate_on_save;
    
    /* Backup settings */
    gboolean create_backups;
    gint max_backups;
    
    /* Error handling */
    TaishangConfigError last_error;
    gchar *last_error_message;
    
    /* Watchers */
    GHashTable *watchers;
    guint next_watcher_id;
    
    /* Private data */
    gpointer priv;
};

/* Configuration class structure */
struct _TaishangConfigClass {
    GObjectClass parent_class;
    
    /* Virtual methods */
    gboolean (*load)(TaishangConfig *config);
    gboolean (*save)(TaishangConfig *config);
    gboolean (*validate)(TaishangConfig *config, TaishangConfigValidation *result);
    void (*reset)(TaishangConfig *config);
    
    /* Signal handlers */
    void (*loaded)(TaishangConfig *config);
    void (*saved)(TaishangConfig *config);
    void (*changed)(TaishangConfig *config, const gchar *key, const gchar *group);
    void (*error_occurred)(TaishangConfig *config, TaishangConfigError error, const gchar *message);
    
    /* Reserved for future use */
    gpointer reserved[8];
};

/* Configuration watcher callback */
typedef void (*TaishangConfigWatcherFunc)(TaishangConfig *config, 
                                          const gchar *key, 
                                          const gchar *group,
                                          gpointer user_data);

/* GType macros */
#define TAISHANG_TYPE_CONFIG            (taishang_config_get_type())
#define TAISHANG_CONFIG(obj)            (G_TYPE_CHECK_INSTANCE_CAST((obj), TAISHANG_TYPE_CONFIG, TaishangConfig))
#define TAISHANG_CONFIG_CLASS(klass)    (G_TYPE_CHECK_CLASS_CAST((klass), TAISHANG_TYPE_CONFIG, TaishangConfigClass))
#define TAISHANG_IS_CONFIG(obj)         (G_TYPE_CHECK_INSTANCE_TYPE((obj), TAISHANG_TYPE_CONFIG))
#define TAISHANG_IS_CONFIG_CLASS(klass) (G_TYPE_CHECK_CLASS_TYPE((klass), TAISHANG_TYPE_CONFIG))
#define TAISHANG_CONFIG_GET_CLASS(obj)  (G_TYPE_INSTANCE_GET_CLASS((obj), TAISHANG_TYPE_CONFIG, TaishangConfigClass))

/* Configuration lifecycle functions */
GType taishang_config_get_type(void) G_GNUC_CONST;
TaishangConfig *taishang_config_new(void);
TaishangConfig *taishang_config_new_with_file(const gchar *config_file);

gboolean taishang_config_load(TaishangConfig *config);
gboolean taishang_config_save(TaishangConfig *config);
gboolean taishang_config_reload(TaishangConfig *config);
void taishang_config_reset(TaishangConfig *config);

/* Configuration file management */
const gchar *taishang_config_get_file(TaishangConfig *config);
void taishang_config_set_file(TaishangConfig *config, const gchar *config_file);

const gchar *taishang_config_get_dir(TaishangConfig *config);
void taishang_config_set_dir(TaishangConfig *config, const gchar *config_dir);

gboolean taishang_config_file_exists(TaishangConfig *config);
gboolean taishang_config_ensure_dir(TaishangConfig *config);

/* Configuration state */
gboolean taishang_config_is_loaded(TaishangConfig *config);
gboolean taishang_config_is_modified(TaishangConfig *config);
void taishang_config_set_modified(TaishangConfig *config, gboolean modified);

gboolean taishang_config_get_auto_save(TaishangConfig *config);
void taishang_config_set_auto_save(TaishangConfig *config, gboolean auto_save);

gint64 taishang_config_get_last_modified(TaishangConfig *config);

/* Value getters */
gboolean taishang_config_get_boolean(TaishangConfig *config, const gchar *group, const gchar *key, gboolean default_value);
gint taishang_config_get_integer(TaishangConfig *config, const gchar *group, const gchar *key, gint default_value);
gdouble taishang_config_get_double(TaishangConfig *config, const gchar *group, const gchar *key, gdouble default_value);
gchar *taishang_config_get_string(TaishangConfig *config, const gchar *group, const gchar *key, const gchar *default_value);
gchar **taishang_config_get_string_list(TaishangConfig *config, const gchar *group, const gchar *key, gchar **default_value);
JsonObject *taishang_config_get_object(TaishangConfig *config, const gchar *group, const gchar *key);

/* Value setters */
void taishang_config_set_boolean(TaishangConfig *config, const gchar *group, const gchar *key, gboolean value);
void taishang_config_set_integer(TaishangConfig *config, const gchar *group, const gchar *key, gint value);
void taishang_config_set_double(TaishangConfig *config, const gchar *group, const gchar *key, gdouble value);
void taishang_config_set_string(TaishangConfig *config, const gchar *group, const gchar *key, const gchar *value);
void taishang_config_set_string_list(TaishangConfig *config, const gchar *group, const gchar *key, gchar **value);
void taishang_config_set_object(TaishangConfig *config, const gchar *group, const gchar *key, JsonObject *value);

/* Key management */
gboolean taishang_config_has_key(TaishangConfig *config, const gchar *group, const gchar *key);
gboolean taishang_config_has_group(TaishangConfig *config, const gchar *group);
void taishang_config_remove_key(TaishangConfig *config, const gchar *group, const gchar *key);
void taishang_config_remove_group(TaishangConfig *config, const gchar *group);

gchar **taishang_config_get_groups(TaishangConfig *config);
gchar **taishang_config_get_keys(TaishangConfig *config, const gchar *group);

/* Type checking */
TaishangConfigType taishang_config_get_type_for_key(TaishangConfig *config, const gchar *group, const gchar *key);
gboolean taishang_config_is_type(TaishangConfig *config, const gchar *group, const gchar *key, TaishangConfigType type);

/* Validation */
gboolean taishang_config_validate(TaishangConfig *config, TaishangConfigValidation *result);
gboolean taishang_config_load_schema(TaishangConfig *config, const gchar *schema_file);
void taishang_config_set_validate_on_load(TaishangConfig *config, gboolean validate);
void taishang_config_set_validate_on_save(TaishangConfig *config, gboolean validate);

/* Backup management */
gboolean taishang_config_create_backup(TaishangConfig *config);
gboolean taishang_config_restore_backup(TaishangConfig *config);
void taishang_config_set_create_backups(TaishangConfig *config, gboolean create_backups);
void taishang_config_set_max_backups(TaishangConfig *config, gint max_backups);
gchar **taishang_config_list_backups(TaishangConfig *config);

/* Error handling */
TaishangConfigError taishang_config_get_last_error(TaishangConfig *config);
const gchar *taishang_config_get_last_error_message(TaishangConfig *config);
const gchar *taishang_config_error_to_string(TaishangConfigError error);

/* Watchers */
guint taishang_config_add_watcher(TaishangConfig *config, 
                                  const gchar *group, 
                                  const gchar *key,
                                  TaishangConfigWatcherFunc callback,
                                  gpointer user_data,
                                  GDestroyNotify destroy_notify);
void taishang_config_remove_watcher(TaishangConfig *config, guint watcher_id);

/* Default configuration */
void taishang_config_load_defaults(TaishangConfig *config);
void taishang_config_set_default_boolean(TaishangConfig *config, const gchar *group, const gchar *key, gboolean value);
void taishang_config_set_default_integer(TaishangConfig *config, const gchar *group, const gchar *key, gint value);
void taishang_config_set_default_double(TaishangConfig *config, const gchar *group, const gchar *key, gdouble value);
void taishang_config_set_default_string(TaishangConfig *config, const gchar *group, const gchar *key, const gchar *value);

/* Import/Export */
gboolean taishang_config_import_from_file(TaishangConfig *config, const gchar *file);
gboolean taishang_config_export_to_file(TaishangConfig *config, const gchar *file);
gboolean taishang_config_import_from_string(TaishangConfig *config, const gchar *json_string);
gchar *taishang_config_export_to_string(TaishangConfig *config);

/* Migration */
gboolean taishang_config_migrate(TaishangConfig *config, const gchar *from_version, const gchar *to_version);
const gchar *taishang_config_get_version(TaishangConfig *config);
void taishang_config_set_version(TaishangConfig *config, const gchar *version);

/* Signal emission helpers */
void taishang_config_emit_loaded(TaishangConfig *config);
void taishang_config_emit_saved(TaishangConfig *config);
void taishang_config_emit_changed(TaishangConfig *config, const gchar *key, const gchar *group);
void taishang_config_emit_error_occurred(TaishangConfig *config, TaishangConfigError error, const gchar *message);

/* Utility functions */
gchar *taishang_config_get_default_config_dir(void);
gchar *taishang_config_get_default_config_file(void);

/* Configuration validation result management */
TaishangConfigValidation *taishang_config_validation_new(void);
void taishang_config_validation_free(TaishangConfigValidation *validation);

/* Signal definitions */
#define TAISHANG_CONFIG_SIGNAL_LOADED         "loaded"
#define TAISHANG_CONFIG_SIGNAL_SAVED          "saved"
#define TAISHANG_CONFIG_SIGNAL_CHANGED        "changed"
#define TAISHANG_CONFIG_SIGNAL_ERROR_OCCURRED "error-occurred"

/* Default configuration values */
#define TAISHANG_CONFIG_DEFAULT_AUTO_SAVE           TRUE
#define TAISHANG_CONFIG_DEFAULT_CREATE_BACKUPS      TRUE
#define TAISHANG_CONFIG_DEFAULT_MAX_BACKUPS         5
#define TAISHANG_CONFIG_DEFAULT_VALIDATE_ON_LOAD    TRUE
#define TAISHANG_CONFIG_DEFAULT_VALIDATE_ON_SAVE    TRUE

/* Common configuration keys */
#define TAISHANG_CONFIG_KEY_FIRST_RUN           "first_run"
#define TAISHANG_CONFIG_KEY_LANGUAGE            "language"
#define TAISHANG_CONFIG_KEY_THEME               "theme"
#define TAISHANG_CONFIG_KEY_WINDOW_WIDTH        "window_width"
#define TAISHANG_CONFIG_KEY_WINDOW_HEIGHT       "window_height"
#define TAISHANG_CONFIG_KEY_WINDOW_X            "window_x"
#define TAISHANG_CONFIG_KEY_WINDOW_Y            "window_y"
#define TAISHANG_CONFIG_KEY_WINDOW_MAXIMIZED    "window_maximized"
#define TAISHANG_CONFIG_KEY_SHOW_TOOLBAR        "show_toolbar"
#define TAISHANG_CONFIG_KEY_SHOW_STATUS_BAR     "show_status_bar"
#define TAISHANG_CONFIG_KEY_SHOW_SIDEBAR        "show_sidebar"
#define TAISHANG_CONFIG_KEY_AUTO_START          "auto_start"
#define TAISHANG_CONFIG_KEY_MINIMIZE_TO_TRAY    "minimize_to_tray"
#define TAISHANG_CONFIG_KEY_NOTIFICATIONS       "notifications_enabled"
#define TAISHANG_CONFIG_KEY_SOUND_ENABLED       "sound_enabled"

#ifdef __cplusplus
}
#endif

#endif /* TAISHANG_CONFIG_H */