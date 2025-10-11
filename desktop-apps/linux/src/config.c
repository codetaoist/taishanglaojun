/**
 * @file config.c
 * @brief TaishangLaojun Desktop Application Configuration Implementation
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains the configuration management implementation for the
 * TaishangLaojun desktop application on Linux.
 */

#include "common.h"
#include "config.h"
#include "utils.h"

/* Private structure */
struct _TaishangConfigPrivate {
    gchar *config_file;
    json_object *root_object;
    GHashTable *watchers;
    
    gboolean loaded;
    gboolean modified;
    gboolean auto_save;
    
    GMutex config_mutex;
    GFileMonitor *file_monitor;
    
    /* Backup management */
    gchar *backup_dir;
    gint max_backups;
    
    /* Validation */
    GHashTable *validators;
    
    /* Default values */
    GHashTable *defaults;
    
    /* Migration */
    gint config_version;
    gint current_version;
};

/* Watcher structure */
typedef struct {
    TaishangConfigWatchFunc callback;
    gpointer user_data;
    GDestroyNotify destroy_func;
} ConfigWatcher;

/* Validator structure */
typedef struct {
    TaishangConfigValidateFunc callback;
    gpointer user_data;
    GDestroyNotify destroy_func;
} ConfigValidator;

/* GObject implementation */
G_DEFINE_TYPE_WITH_PRIVATE(TaishangConfig, taishang_config, G_TYPE_OBJECT)

/* Property IDs */
enum {
    PROP_0,
    PROP_CONFIG_FILE,
    PROP_LOADED,
    PROP_MODIFIED,
    PROP_AUTO_SAVE,
    N_PROPERTIES
};

/* Signal IDs */
enum {
    SIGNAL_CHANGED,
    SIGNAL_LOADED,
    SIGNAL_SAVED,
    SIGNAL_ERROR,
    N_SIGNALS
};

static GParamSpec *properties[N_PROPERTIES] = { NULL, };
static guint signals[N_SIGNALS] = { 0, };

/* Forward declarations */
static void taishang_config_finalize(GObject *object);
static void taishang_config_get_property(GObject *object, guint prop_id, GValue *value, GParamSpec *pspec);
static void taishang_config_set_property(GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec);

static json_object *taishang_config_get_group_object(TaishangConfig *config, const gchar *group, gboolean create);
static gboolean taishang_config_validate_key(TaishangConfig *config, const gchar *group, const gchar *key, json_object *value, GError **error);
static void taishang_config_emit_changed(TaishangConfig *config, const gchar *key);
static void taishang_config_create_backup(TaishangConfig *config);
static void taishang_config_cleanup_backups(TaishangConfig *config);
static gboolean taishang_config_migrate(TaishangConfig *config, GError **error);
static void on_file_changed(GFileMonitor *monitor, GFile *file, GFile *other_file, GFileMonitorEvent event_type, gpointer user_data);

/* Helper functions */
static ConfigWatcher *config_watcher_new(TaishangConfigWatchFunc callback, gpointer user_data, GDestroyNotify destroy_func);
static void config_watcher_free(ConfigWatcher *watcher);
static ConfigValidator *config_validator_new(TaishangConfigValidateFunc callback, gpointer user_data, GDestroyNotify destroy_func);
static void config_validator_free(ConfigValidator *validator);

/* Class initialization */
static void taishang_config_class_init(TaishangConfigClass *klass) {
    GObjectClass *object_class = G_OBJECT_CLASS(klass);
    
    object_class->finalize = taishang_config_finalize;
    object_class->get_property = taishang_config_get_property;
    object_class->set_property = taishang_config_set_property;
    
    /* Properties */
    properties[PROP_CONFIG_FILE] = g_param_spec_string(
        "config-file", "Config File", "Configuration file path",
        NULL,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_LOADED] = g_param_spec_boolean(
        "loaded", "Loaded", "Whether configuration is loaded",
        FALSE,
        TAISHANG_PARAM_READABLE);
    
    properties[PROP_MODIFIED] = g_param_spec_boolean(
        "modified", "Modified", "Whether configuration is modified",
        FALSE,
        TAISHANG_PARAM_READABLE);
    
    properties[PROP_AUTO_SAVE] = g_param_spec_boolean(
        "auto-save", "Auto Save", "Whether to auto-save configuration",
        TRUE,
        TAISHANG_PARAM_READWRITE);
    
    g_object_class_install_properties(object_class, N_PROPERTIES, properties);
    
    /* Signals */
    signals[SIGNAL_CHANGED] = g_signal_new(
        "changed",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangConfigClass, changed),
        NULL, NULL,
        g_cclosure_marshal_VOID__STRING,
        G_TYPE_NONE, 1, G_TYPE_STRING);
    
    signals[SIGNAL_LOADED] = g_signal_new(
        "loaded",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangConfigClass, loaded),
        NULL, NULL,
        g_cclosure_marshal_VOID__VOID,
        G_TYPE_NONE, 0);
    
    signals[SIGNAL_SAVED] = g_signal_new(
        "saved",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangConfigClass, saved),
        NULL, NULL,
        g_cclosure_marshal_VOID__VOID,
        G_TYPE_NONE, 0);
    
    signals[SIGNAL_ERROR] = g_signal_new(
        "error",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangConfigClass, error),
        NULL, NULL,
        g_cclosure_marshal_VOID__POINTER,
        G_TYPE_NONE, 1, G_TYPE_POINTER);
}

/* Instance initialization */
static void taishang_config_init(TaishangConfig *config) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    priv->config_file = NULL;
    priv->root_object = NULL;
    priv->watchers = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, (GDestroyNotify)config_watcher_free);
    
    priv->loaded = FALSE;
    priv->modified = FALSE;
    priv->auto_save = TRUE;
    
    g_mutex_init(&priv->config_mutex);
    priv->file_monitor = NULL;
    
    priv->backup_dir = NULL;
    priv->max_backups = 10;
    
    priv->validators = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, (GDestroyNotify)config_validator_free);
    priv->defaults = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, (GDestroyNotify)json_object_put);
    
    priv->config_version = 1;
    priv->current_version = 1;
}

/* Object finalization */
static void taishang_config_finalize(GObject *object) {
    TaishangConfig *config = TAISHANG_CONFIG(object);
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    /* Save if modified and auto-save is enabled */
    if (priv->modified && priv->auto_save) {
        GError *error = NULL;
        if (!taishang_config_save(config, &error)) {
            g_warning("Failed to auto-save configuration: %s", error->message);
            g_error_free(error);
        }
    }
    
    /* Cleanup file monitor */
    if (priv->file_monitor) {
        g_file_monitor_cancel(priv->file_monitor);
        g_object_unref(priv->file_monitor);
    }
    
    /* Cleanup JSON object */
    if (priv->root_object) {
        json_object_put(priv->root_object);
    }
    
    /* Cleanup hash tables */
    if (priv->watchers) {
        g_hash_table_destroy(priv->watchers);
    }
    
    if (priv->validators) {
        g_hash_table_destroy(priv->validators);
    }
    
    if (priv->defaults) {
        g_hash_table_destroy(priv->defaults);
    }
    
    /* Cleanup strings */
    TAISHANG_FREE(priv->config_file);
    TAISHANG_FREE(priv->backup_dir);
    
    g_mutex_clear(&priv->config_mutex);
    
    G_OBJECT_CLASS(taishang_config_parent_class)->finalize(object);
}

/* Property getters and setters */
static void taishang_config_get_property(GObject *object, guint prop_id, GValue *value, GParamSpec *pspec) {
    TaishangConfig *config = TAISHANG_CONFIG(object);
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    switch (prop_id) {
        case PROP_CONFIG_FILE:
            g_value_set_string(value, priv->config_file);
            break;
        case PROP_LOADED:
            g_value_set_boolean(value, priv->loaded);
            break;
        case PROP_MODIFIED:
            g_value_set_boolean(value, priv->modified);
            break;
        case PROP_AUTO_SAVE:
            g_value_set_boolean(value, priv->auto_save);
            break;
        default:
            G_OBJECT_WARN_INVALID_PROPERTY_ID(object, prop_id, pspec);
            break;
    }
}

static void taishang_config_set_property(GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec) {
    TaishangConfig *config = TAISHANG_CONFIG(object);
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    switch (prop_id) {
        case PROP_CONFIG_FILE:
            taishang_config_set_file(config, g_value_get_string(value));
            break;
        case PROP_AUTO_SAVE:
            priv->auto_save = g_value_get_boolean(value);
            break;
        default:
            G_OBJECT_WARN_INVALID_PROPERTY_ID(object, prop_id, pspec);
            break;
    }
}

/* Public API implementation */

/**
 * taishang_config_new:
 * 
 * Creates a new TaishangConfig instance.
 * 
 * Returns: (transfer full): A new TaishangConfig instance
 */
TaishangConfig *taishang_config_new(void) {
    return g_object_new(TAISHANG_TYPE_CONFIG, NULL);
}

/**
 * taishang_config_set_file:
 * @config: A TaishangConfig instance
 * @file_path: Configuration file path
 * 
 * Sets the configuration file path.
 */
void taishang_config_set_file(TaishangConfig *config, const gchar *file_path) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_CONFIG(config));
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    if (g_strcmp0(priv->config_file, file_path) != 0) {
        TAISHANG_FREE(priv->config_file);
        priv->config_file = g_strdup(file_path);
        
        /* Set up backup directory */
        if (file_path) {
            gchar *dir = g_path_get_dirname(file_path);
            priv->backup_dir = g_build_filename(dir, "backups", NULL);
            g_free(dir);
        }
        
        g_object_notify_by_pspec(G_OBJECT(config), properties[PROP_CONFIG_FILE]);
    }
    
    g_mutex_unlock(&priv->config_mutex);
}

/**
 * taishang_config_get_file:
 * @config: A TaishangConfig instance
 * 
 * Gets the configuration file path.
 * 
 * Returns: (transfer none): The configuration file path
 */
const gchar *taishang_config_get_file(TaishangConfig *config) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), NULL);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    return priv->config_file;
}

/**
 * taishang_config_load:
 * @config: A TaishangConfig instance
 * @error: Return location for error
 * 
 * Loads configuration from file.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_config_load(TaishangConfig *config, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    if (!priv->config_file) {
        g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_INVALID_ARGUMENT,
                   "No configuration file set");
        return FALSE;
    }
    
    g_mutex_lock(&priv->config_mutex);
    
    /* Cleanup existing configuration */
    if (priv->root_object) {
        json_object_put(priv->root_object);
        priv->root_object = NULL;
    }
    
    /* Check if file exists */
    if (!g_file_test(priv->config_file, G_FILE_TEST_EXISTS)) {
        /* Create default configuration */
        priv->root_object = json_object_new_object();
        priv->loaded = TRUE;
        priv->modified = TRUE;
        
        g_mutex_unlock(&priv->config_mutex);
        
        /* Save default configuration */
        if (!taishang_config_save(config, error)) {
            return FALSE;
        }
        
        g_signal_emit(config, signals[SIGNAL_LOADED], 0);
        return TRUE;
    }
    
    /* Load from file */
    gchar *contents = NULL;
    gsize length = 0;
    
    if (!g_file_get_contents(priv->config_file, &contents, &length, error)) {
        g_mutex_unlock(&priv->config_mutex);
        return FALSE;
    }
    
    /* Parse JSON */
    json_tokener *tokener = json_tokener_new();
    priv->root_object = json_tokener_parse_ex(tokener, contents, length);
    
    if (!priv->root_object) {
        enum json_tokener_error jerr = json_tokener_get_error(tokener);
        g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_INVALID_ARGUMENT,
                   "Failed to parse JSON: %s", json_tokener_error_desc(jerr));
        json_tokener_free(tokener);
        g_free(contents);
        g_mutex_unlock(&priv->config_mutex);
        return FALSE;
    }
    
    json_tokener_free(tokener);
    g_free(contents);
    
    /* Check version and migrate if needed */
    json_object *version_obj = NULL;
    if (json_object_object_get_ex(priv->root_object, "version", &version_obj)) {
        priv->config_version = json_object_get_int(version_obj);
    } else {
        priv->config_version = 1;
    }
    
    if (priv->config_version < priv->current_version) {
        if (!taishang_config_migrate(config, error)) {
            g_mutex_unlock(&priv->config_mutex);
            return FALSE;
        }
    }
    
    priv->loaded = TRUE;
    priv->modified = FALSE;
    
    g_mutex_unlock(&priv->config_mutex);
    
    /* Set up file monitoring */
    if (priv->config_file) {
        GFile *file = g_file_new_for_path(priv->config_file);
        priv->file_monitor = g_file_monitor_file(file, G_FILE_MONITOR_NONE, NULL, NULL);
        if (priv->file_monitor) {
            g_signal_connect(priv->file_monitor, "changed",
                           G_CALLBACK(on_file_changed), config);
        }
        g_object_unref(file);
    }
    
    g_signal_emit(config, signals[SIGNAL_LOADED], 0);
    
    g_info("Configuration loaded from %s", priv->config_file);
    
    return TRUE;
}

/**
 * taishang_config_save:
 * @config: A TaishangConfig instance
 * @error: Return location for error
 * 
 * Saves configuration to file.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_config_save(TaishangConfig *config, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    if (!priv->config_file) {
        g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_INVALID_ARGUMENT,
                   "No configuration file set");
        return FALSE;
    }
    
    if (!priv->root_object) {
        g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_INVALID_ARGUMENT,
                   "No configuration data to save");
        return FALSE;
    }
    
    g_mutex_lock(&priv->config_mutex);
    
    /* Create backup if file exists */
    if (g_file_test(priv->config_file, G_FILE_TEST_EXISTS)) {
        taishang_config_create_backup(config);
    }
    
    /* Set version */
    json_object *version_obj = json_object_new_int(priv->current_version);
    json_object_object_add(priv->root_object, "version", version_obj);
    
    /* Convert to string */
    const gchar *json_string = json_object_to_json_string_ext(priv->root_object, JSON_C_TO_STRING_PRETTY);
    
    /* Create directory if needed */
    gchar *dir = g_path_get_dirname(priv->config_file);
    if (!taishang_utils_create_directory(dir, error)) {
        g_free(dir);
        g_mutex_unlock(&priv->config_mutex);
        return FALSE;
    }
    g_free(dir);
    
    /* Write to file */
    if (!g_file_set_contents(priv->config_file, json_string, -1, error)) {
        g_mutex_unlock(&priv->config_mutex);
        return FALSE;
    }
    
    priv->modified = FALSE;
    
    g_mutex_unlock(&priv->config_mutex);
    
    /* Cleanup old backups */
    taishang_config_cleanup_backups(config);
    
    g_signal_emit(config, signals[SIGNAL_SAVED], 0);
    g_object_notify_by_pspec(G_OBJECT(config), properties[PROP_MODIFIED]);
    
    g_debug("Configuration saved to %s", priv->config_file);
    
    return TRUE;
}

/**
 * taishang_config_get_string:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * @default_value: Default value if key not found
 * 
 * Gets a string value from configuration.
 * 
 * Returns: (transfer full): The string value
 */
gchar *taishang_config_get_string(TaishangConfig *config, const gchar *group, const gchar *key, const gchar *default_value) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), g_strdup(default_value));
    TAISHANG_RETURN_VAL_IF_FAIL(group != NULL, g_strdup(default_value));
    TAISHANG_RETURN_VAL_IF_FAIL(key != NULL, g_strdup(default_value));
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, FALSE);
    if (!group_obj) {
        g_mutex_unlock(&priv->config_mutex);
        return g_strdup(default_value);
    }
    
    json_object *value_obj = NULL;
    if (!json_object_object_get_ex(group_obj, key, &value_obj)) {
        g_mutex_unlock(&priv->config_mutex);
        return g_strdup(default_value);
    }
    
    const gchar *value = json_object_get_string(value_obj);
    gchar *result = g_strdup(value ? value : default_value);
    
    g_mutex_unlock(&priv->config_mutex);
    
    return result;
}

/**
 * taishang_config_set_string:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * @value: Value to set
 * 
 * Sets a string value in configuration.
 */
void taishang_config_set_string(TaishangConfig *config, const gchar *group, const gchar *key, const gchar *value) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_CONFIG(config));
    TAISHANG_RETURN_IF_FAIL(group != NULL);
    TAISHANG_RETURN_IF_FAIL(key != NULL);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, TRUE);
    json_object *value_obj = json_object_new_string(value ? value : "");
    
    /* Validate value */
    GError *error = NULL;
    if (!taishang_config_validate_key(config, group, key, value_obj, &error)) {
        g_warning("Validation failed for %s.%s: %s", group, key, error->message);
        g_error_free(error);
        json_object_put(value_obj);
        g_mutex_unlock(&priv->config_mutex);
        return;
    }
    
    json_object_object_add(group_obj, key, value_obj);
    priv->modified = TRUE;
    
    g_mutex_unlock(&priv->config_mutex);
    
    gchar *full_key = g_strdup_printf("%s.%s", group, key);
    taishang_config_emit_changed(config, full_key);
    g_free(full_key);
    
    g_object_notify_by_pspec(G_OBJECT(config), properties[PROP_MODIFIED]);
}

/**
 * taishang_config_get_int:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * @default_value: Default value if key not found
 * 
 * Gets an integer value from configuration.
 * 
 * Returns: The integer value
 */
gint taishang_config_get_int(TaishangConfig *config, const gchar *group, const gchar *key, gint default_value) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), default_value);
    TAISHANG_RETURN_VAL_IF_FAIL(group != NULL, default_value);
    TAISHANG_RETURN_VAL_IF_FAIL(key != NULL, default_value);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, FALSE);
    if (!group_obj) {
        g_mutex_unlock(&priv->config_mutex);
        return default_value;
    }
    
    json_object *value_obj = NULL;
    if (!json_object_object_get_ex(group_obj, key, &value_obj)) {
        g_mutex_unlock(&priv->config_mutex);
        return default_value;
    }
    
    gint result = json_object_get_int(value_obj);
    
    g_mutex_unlock(&priv->config_mutex);
    
    return result;
}

/**
 * taishang_config_set_int:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * @value: Value to set
 * 
 * Sets an integer value in configuration.
 */
void taishang_config_set_int(TaishangConfig *config, const gchar *group, const gchar *key, gint value) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_CONFIG(config));
    TAISHANG_RETURN_IF_FAIL(group != NULL);
    TAISHANG_RETURN_IF_FAIL(key != NULL);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, TRUE);
    json_object *value_obj = json_object_new_int(value);
    
    /* Validate value */
    GError *error = NULL;
    if (!taishang_config_validate_key(config, group, key, value_obj, &error)) {
        g_warning("Validation failed for %s.%s: %s", group, key, error->message);
        g_error_free(error);
        json_object_put(value_obj);
        g_mutex_unlock(&priv->config_mutex);
        return;
    }
    
    json_object_object_add(group_obj, key, value_obj);
    priv->modified = TRUE;
    
    g_mutex_unlock(&priv->config_mutex);
    
    gchar *full_key = g_strdup_printf("%s.%s", group, key);
    taishang_config_emit_changed(config, full_key);
    g_free(full_key);
    
    g_object_notify_by_pspec(G_OBJECT(config), properties[PROP_MODIFIED]);
}

/**
 * taishang_config_get_boolean:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * @default_value: Default value if key not found
 * 
 * Gets a boolean value from configuration.
 * 
 * Returns: The boolean value
 */
gboolean taishang_config_get_boolean(TaishangConfig *config, const gchar *group, const gchar *key, gboolean default_value) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), default_value);
    TAISHANG_RETURN_VAL_IF_FAIL(group != NULL, default_value);
    TAISHANG_RETURN_VAL_IF_FAIL(key != NULL, default_value);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, FALSE);
    if (!group_obj) {
        g_mutex_unlock(&priv->config_mutex);
        return default_value;
    }
    
    json_object *value_obj = NULL;
    if (!json_object_object_get_ex(group_obj, key, &value_obj)) {
        g_mutex_unlock(&priv->config_mutex);
        return default_value;
    }
    
    gboolean result = json_object_get_boolean(value_obj);
    
    g_mutex_unlock(&priv->config_mutex);
    
    return result;
}

/**
 * taishang_config_set_boolean:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * @value: Value to set
 * 
 * Sets a boolean value in configuration.
 */
void taishang_config_set_boolean(TaishangConfig *config, const gchar *group, const gchar *key, gboolean value) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_CONFIG(config));
    TAISHANG_RETURN_IF_FAIL(group != NULL);
    TAISHANG_RETURN_IF_FAIL(key != NULL);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, TRUE);
    json_object *value_obj = json_object_new_boolean(value);
    
    /* Validate value */
    GError *error = NULL;
    if (!taishang_config_validate_key(config, group, key, value_obj, &error)) {
        g_warning("Validation failed for %s.%s: %s", group, key, error->message);
        g_error_free(error);
        json_object_put(value_obj);
        g_mutex_unlock(&priv->config_mutex);
        return;
    }
    
    json_object_object_add(group_obj, key, value_obj);
    priv->modified = TRUE;
    
    g_mutex_unlock(&priv->config_mutex);
    
    gchar *full_key = g_strdup_printf("%s.%s", group, key);
    taishang_config_emit_changed(config, full_key);
    g_free(full_key);
    
    g_object_notify_by_pspec(G_OBJECT(config), properties[PROP_MODIFIED]);
}

/**
 * taishang_config_has_key:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * 
 * Checks if a key exists in configuration.
 * 
 * Returns: TRUE if key exists, FALSE otherwise
 */
gboolean taishang_config_has_key(TaishangConfig *config, const gchar *group, const gchar *key) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(group != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(key != NULL, FALSE);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, FALSE);
    if (!group_obj) {
        g_mutex_unlock(&priv->config_mutex);
        return FALSE;
    }
    
    json_object *value_obj = NULL;
    gboolean result = json_object_object_get_ex(group_obj, key, &value_obj);
    
    g_mutex_unlock(&priv->config_mutex);
    
    return result;
}

/**
 * taishang_config_remove_key:
 * @config: A TaishangConfig instance
 * @group: Configuration group
 * @key: Configuration key
 * 
 * Removes a key from configuration.
 * 
 * Returns: TRUE if key was removed, FALSE if not found
 */
gboolean taishang_config_remove_key(TaishangConfig *config, const gchar *group, const gchar *key) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_CONFIG(config), FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(group != NULL, FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(key != NULL, FALSE);
    
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_mutex_lock(&priv->config_mutex);
    
    json_object *group_obj = taishang_config_get_group_object(config, group, FALSE);
    if (!group_obj) {
        g_mutex_unlock(&priv->config_mutex);
        return FALSE;
    }
    
    gboolean result = json_object_object_del(group_obj, key);
    if (result) {
        priv->modified = TRUE;
    }
    
    g_mutex_unlock(&priv->config_mutex);
    
    if (result) {
        gchar *full_key = g_strdup_printf("%s.%s", group, key);
        taishang_config_emit_changed(config, full_key);
        g_free(full_key);
        
        g_object_notify_by_pspec(G_OBJECT(config), properties[PROP_MODIFIED]);
    }
    
    return result;
}

/* Private helper functions */

static json_object *taishang_config_get_group_object(TaishangConfig *config, const gchar *group, gboolean create) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    if (!priv->root_object) {
        if (create) {
            priv->root_object = json_object_new_object();
        } else {
            return NULL;
        }
    }
    
    json_object *group_obj = NULL;
    if (!json_object_object_get_ex(priv->root_object, group, &group_obj)) {
        if (create) {
            group_obj = json_object_new_object();
            json_object_object_add(priv->root_object, group, group_obj);
        } else {
            return NULL;
        }
    }
    
    return group_obj;
}

static gboolean taishang_config_validate_key(TaishangConfig *config, const gchar *group, const gchar *key, json_object *value, GError **error) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    gchar *full_key = g_strdup_printf("%s.%s", group, key);
    ConfigValidator *validator = g_hash_table_lookup(priv->validators, full_key);
    
    if (validator && validator->callback) {
        gboolean result = validator->callback(config, group, key, value, validator->user_data, error);
        g_free(full_key);
        return result;
    }
    
    g_free(full_key);
    return TRUE;
}

static void taishang_config_emit_changed(TaishangConfig *config, const gchar *key) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    /* Notify watchers */
    ConfigWatcher *watcher = g_hash_table_lookup(priv->watchers, key);
    if (watcher && watcher->callback) {
        watcher->callback(config, key, watcher->user_data);
    }
    
    /* Emit signal */
    g_signal_emit(config, signals[SIGNAL_CHANGED], 0, key);
}

static void taishang_config_create_backup(TaishangConfig *config) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    if (!priv->backup_dir || !priv->config_file) {
        return;
    }
    
    /* Create backup directory */
    GError *error = NULL;
    if (!taishang_utils_create_directory(priv->backup_dir, &error)) {
        g_warning("Failed to create backup directory: %s", error->message);
        g_error_free(error);
        return;
    }
    
    /* Generate backup filename with timestamp */
    GDateTime *now = g_date_time_new_now_local();
    gchar *timestamp = g_date_time_format(now, "%Y%m%d_%H%M%S");
    gchar *basename = g_path_get_basename(priv->config_file);
    gchar *backup_name = g_strdup_printf("%s.%s.backup", basename, timestamp);
    gchar *backup_path = g_build_filename(priv->backup_dir, backup_name, NULL);
    
    /* Copy file */
    GFile *source = g_file_new_for_path(priv->config_file);
    GFile *dest = g_file_new_for_path(backup_path);
    
    if (!g_file_copy(source, dest, G_FILE_COPY_OVERWRITE, NULL, NULL, NULL, &error)) {
        g_warning("Failed to create backup: %s", error->message);
        g_error_free(error);
    } else {
        g_debug("Created backup: %s", backup_path);
    }
    
    g_object_unref(source);
    g_object_unref(dest);
    g_date_time_unref(now);
    g_free(timestamp);
    g_free(basename);
    g_free(backup_name);
    g_free(backup_path);
}

static void taishang_config_cleanup_backups(TaishangConfig *config) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    if (!priv->backup_dir) {
        return;
    }
    
    /* Get list of backup files */
    GDir *dir = g_dir_open(priv->backup_dir, 0, NULL);
    if (!dir) {
        return;
    }
    
    GPtrArray *backups = g_ptr_array_new_with_free_func(g_free);
    const gchar *name;
    
    while ((name = g_dir_read_name(dir)) != NULL) {
        if (g_str_has_suffix(name, ".backup")) {
            gchar *path = g_build_filename(priv->backup_dir, name, NULL);
            g_ptr_array_add(backups, path);
        }
    }
    
    g_dir_close(dir);
    
    /* Sort by modification time (newest first) */
    g_ptr_array_sort(backups, (GCompareFunc)strcmp);
    
    /* Remove old backups */
    if (backups->len > priv->max_backups) {
        for (guint i = priv->max_backups; i < backups->len; i++) {
            const gchar *path = g_ptr_array_index(backups, i);
            if (g_unlink(path) == 0) {
                g_debug("Removed old backup: %s", path);
            }
        }
    }
    
    g_ptr_array_unref(backups);
}

static gboolean taishang_config_migrate(TaishangConfig *config, GError **error) {
    TaishangConfigPrivate *priv = taishang_config_get_instance_private(config);
    
    g_info("Migrating configuration from version %d to %d", priv->config_version, priv->current_version);
    
    /* Add migration logic here as needed */
    
    priv->config_version = priv->current_version;
    priv->modified = TRUE;
    
    return TRUE;
}

static void on_file_changed(GFileMonitor *monitor, GFile *file, GFile *other_file, GFileMonitorEvent event_type, gpointer user_data) {
    TaishangConfig *config = TAISHANG_CONFIG(user_data);
    
    if (event_type == G_FILE_MONITOR_EVENT_CHANGED) {
        g_debug("Configuration file changed externally, reloading");
        
        GError *error = NULL;
        if (!taishang_config_load(config, &error)) {
            g_warning("Failed to reload configuration: %s", error->message);
            g_signal_emit(config, signals[SIGNAL_ERROR], 0, error);
            g_error_free(error);
        }
    }
}

/* Helper function implementations */

static ConfigWatcher *config_watcher_new(TaishangConfigWatchFunc callback, gpointer user_data, GDestroyNotify destroy_func) {
    ConfigWatcher *watcher = g_new0(ConfigWatcher, 1);
    watcher->callback = callback;
    watcher->user_data = user_data;
    watcher->destroy_func = destroy_func;
    return watcher;
}

static void config_watcher_free(ConfigWatcher *watcher) {
    if (watcher) {
        if (watcher->destroy_func && watcher->user_data) {
            watcher->destroy_func(watcher->user_data);
        }
        g_free(watcher);
    }
}

static ConfigValidator *config_validator_new(TaishangConfigValidateFunc callback, gpointer user_data, GDestroyNotify destroy_func) {
    ConfigValidator *validator = g_new0(ConfigValidator, 1);
    validator->callback = callback;
    validator->user_data = user_data;
    validator->destroy_func = destroy_func;
    return validator;
}

static void config_validator_free(ConfigValidator *validator) {
    if (validator) {
        if (validator->destroy_func && validator->user_data) {
            validator->destroy_func(validator->user_data);
        }
        g_free(validator);
    }
}