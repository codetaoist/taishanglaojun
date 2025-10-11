/**
 * @file app.c
 * @brief TaishangLaojun Desktop Application Main Implementation
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains the main application implementation for the
 * TaishangLaojun desktop application on Linux.
 */

#include "common.h"
#include "app.h"
#include "ui.h"
#include "config.h"
#include "utils.h"

/* Private structure */
struct _TaishangAppPrivate {
    TaishangAppState state;
    TaishangConfig *config;
    TaishangUI *ui;
    GApplication *app;
    
    gchar *config_dir;
    gchar *data_dir;
    gchar *cache_dir;
    gchar *log_file;
    
    gboolean debug_mode;
    gboolean verbose_mode;
    gboolean headless_mode;
    
    GMainLoop *main_loop;
    guint auto_save_timeout_id;
    
    GHashTable *plugins;
    GList *signal_handlers;
    
    GMutex state_mutex;
    GCond state_cond;
    
    gint exit_code;
    gboolean shutdown_requested;
};

/* GObject implementation */
G_DEFINE_TYPE_WITH_PRIVATE(TaishangApp, taishang_app, G_TYPE_OBJECT)

/* Property IDs */
enum {
    PROP_0,
    PROP_STATE,
    PROP_CONFIG,
    PROP_UI,
    PROP_DEBUG_MODE,
    PROP_VERBOSE_MODE,
    PROP_HEADLESS_MODE,
    N_PROPERTIES
};

/* Signal IDs */
enum {
    SIGNAL_STATE_CHANGED,
    SIGNAL_STARTUP,
    SIGNAL_SHUTDOWN,
    SIGNAL_ERROR,
    SIGNAL_CONFIG_CHANGED,
    N_SIGNALS
};

static GParamSpec *properties[N_PROPERTIES] = { NULL, };
static guint signals[N_SIGNALS] = { 0, };

/* Static variables */
static TaishangApp *app_instance = NULL;
static gboolean app_initialized = FALSE;

/* Forward declarations */
static void taishang_app_finalize(GObject *object);
static void taishang_app_get_property(GObject *object, guint prop_id, GValue *value, GParamSpec *pspec);
static void taishang_app_set_property(GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec);

static gboolean taishang_app_setup_directories(TaishangApp *app, GError **error);
static gboolean taishang_app_setup_logging(TaishangApp *app, GError **error);
static gboolean taishang_app_load_configuration(TaishangApp *app, GError **error);
static gboolean taishang_app_setup_ui(TaishangApp *app, GError **error);
static gboolean taishang_app_setup_plugins(TaishangApp *app, GError **error);
static void taishang_app_setup_signal_handlers(TaishangApp *app);
static gboolean taishang_app_auto_save_callback(gpointer user_data);
static void taishang_app_cleanup_resources(TaishangApp *app);

/* Signal handlers */
static void on_unix_signal(int signum);
static void on_config_changed(TaishangConfig *config, const gchar *key, gpointer user_data);
static void on_ui_close_request(TaishangUI *ui, gpointer user_data);

/* Class initialization */
static void taishang_app_class_init(TaishangAppClass *klass) {
    GObjectClass *object_class = G_OBJECT_CLASS(klass);
    
    object_class->finalize = taishang_app_finalize;
    object_class->get_property = taishang_app_get_property;
    object_class->set_property = taishang_app_set_property;
    
    /* Properties */
    properties[PROP_STATE] = g_param_spec_enum(
        "state", "State", "Application state",
        TAISHANG_TYPE_APP_STATE, TAISHANG_APP_STATE_UNINITIALIZED,
        TAISHANG_PARAM_READABLE);
    
    properties[PROP_CONFIG] = g_param_spec_object(
        "config", "Config", "Configuration object",
        TAISHANG_TYPE_CONFIG,
        TAISHANG_PARAM_READABLE);
    
    properties[PROP_UI] = g_param_spec_object(
        "ui", "UI", "User interface object",
        TAISHANG_TYPE_UI,
        TAISHANG_PARAM_READABLE);
    
    properties[PROP_DEBUG_MODE] = g_param_spec_boolean(
        "debug-mode", "Debug Mode", "Enable debug mode",
        FALSE,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_VERBOSE_MODE] = g_param_spec_boolean(
        "verbose-mode", "Verbose Mode", "Enable verbose logging",
        FALSE,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_HEADLESS_MODE] = g_param_spec_boolean(
        "headless-mode", "Headless Mode", "Run without UI",
        FALSE,
        TAISHANG_PARAM_READWRITE);
    
    g_object_class_install_properties(object_class, N_PROPERTIES, properties);
    
    /* Signals */
    signals[SIGNAL_STATE_CHANGED] = g_signal_new(
        "state-changed",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangAppClass, state_changed),
        NULL, NULL,
        g_cclosure_marshal_VOID__ENUM,
        G_TYPE_NONE, 1, TAISHANG_TYPE_APP_STATE);
    
    signals[SIGNAL_STARTUP] = g_signal_new(
        "startup",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangAppClass, startup),
        NULL, NULL,
        g_cclosure_marshal_VOID__VOID,
        G_TYPE_NONE, 0);
    
    signals[SIGNAL_SHUTDOWN] = g_signal_new(
        "shutdown",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangAppClass, shutdown),
        NULL, NULL,
        g_cclosure_marshal_VOID__VOID,
        G_TYPE_NONE, 0);
    
    signals[SIGNAL_ERROR] = g_signal_new(
        "error",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangAppClass, error),
        NULL, NULL,
        g_cclosure_marshal_VOID__POINTER,
        G_TYPE_NONE, 1, G_TYPE_POINTER);
    
    signals[SIGNAL_CONFIG_CHANGED] = g_signal_new(
        "config-changed",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangAppClass, config_changed),
        NULL, NULL,
        g_cclosure_marshal_VOID__STRING,
        G_TYPE_NONE, 1, G_TYPE_STRING);
}

/* Instance initialization */
static void taishang_app_init(TaishangApp *app) {
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    priv->state = TAISHANG_APP_STATE_UNINITIALIZED;
    priv->config = NULL;
    priv->ui = NULL;
    priv->app = NULL;
    
    priv->config_dir = NULL;
    priv->data_dir = NULL;
    priv->cache_dir = NULL;
    priv->log_file = NULL;
    
    priv->debug_mode = FALSE;
    priv->verbose_mode = FALSE;
    priv->headless_mode = FALSE;
    
    priv->main_loop = NULL;
    priv->auto_save_timeout_id = 0;
    
    priv->plugins = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, g_object_unref);
    priv->signal_handlers = NULL;
    
    g_mutex_init(&priv->state_mutex);
    g_cond_init(&priv->state_cond);
    
    priv->exit_code = 0;
    priv->shutdown_requested = FALSE;
}

/* Object finalization */
static void taishang_app_finalize(GObject *object) {
    TaishangApp *app = TAISHANG_APP(object);
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    taishang_app_cleanup_resources(app);
    
    if (priv->auto_save_timeout_id > 0) {
        g_source_remove(priv->auto_save_timeout_id);
    }
    
    if (priv->main_loop) {
        g_main_loop_unref(priv->main_loop);
    }
    
    if (priv->plugins) {
        g_hash_table_destroy(priv->plugins);
    }
    
    if (priv->signal_handlers) {
        g_list_free(priv->signal_handlers);
    }
    
    TAISHANG_FREE(priv->config_dir);
    TAISHANG_FREE(priv->data_dir);
    TAISHANG_FREE(priv->cache_dir);
    TAISHANG_FREE(priv->log_file);
    
    TAISHANG_UNREF(priv->config);
    TAISHANG_UNREF(priv->ui);
    TAISHANG_UNREF(priv->app);
    
    g_mutex_clear(&priv->state_mutex);
    g_cond_clear(&priv->state_cond);
    
    G_OBJECT_CLASS(taishang_app_parent_class)->finalize(object);
}

/* Property getters and setters */
static void taishang_app_get_property(GObject *object, guint prop_id, GValue *value, GParamSpec *pspec) {
    TaishangApp *app = TAISHANG_APP(object);
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    switch (prop_id) {
        case PROP_STATE:
            g_value_set_enum(value, priv->state);
            break;
        case PROP_CONFIG:
            g_value_set_object(value, priv->config);
            break;
        case PROP_UI:
            g_value_set_object(value, priv->ui);
            break;
        case PROP_DEBUG_MODE:
            g_value_set_boolean(value, priv->debug_mode);
            break;
        case PROP_VERBOSE_MODE:
            g_value_set_boolean(value, priv->verbose_mode);
            break;
        case PROP_HEADLESS_MODE:
            g_value_set_boolean(value, priv->headless_mode);
            break;
        default:
            G_OBJECT_WARN_INVALID_PROPERTY_ID(object, prop_id, pspec);
            break;
    }
}

static void taishang_app_set_property(GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec) {
    TaishangApp *app = TAISHANG_APP(object);
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    switch (prop_id) {
        case PROP_DEBUG_MODE:
            priv->debug_mode = g_value_get_boolean(value);
            break;
        case PROP_VERBOSE_MODE:
            priv->verbose_mode = g_value_get_boolean(value);
            break;
        case PROP_HEADLESS_MODE:
            priv->headless_mode = g_value_get_boolean(value);
            break;
        default:
            G_OBJECT_WARN_INVALID_PROPERTY_ID(object, prop_id, pspec);
            break;
    }
}

/* Public API implementation */

/**
 * taishang_app_new:
 * 
 * Creates a new TaishangApp instance.
 * 
 * Returns: (transfer full): A new TaishangApp instance
 */
TaishangApp *taishang_app_new(void) {
    return g_object_new(TAISHANG_TYPE_APP, NULL);
}

/**
 * taishang_app_get_default:
 * 
 * Gets the default application instance.
 * 
 * Returns: (transfer none): The default application instance
 */
TaishangApp *taishang_app_get_default(void) {
    if (!app_instance) {
        app_instance = taishang_app_new();
    }
    return app_instance;
}

/**
 * taishang_app_initialize:
 * @app: A TaishangApp instance
 * @argc: Command line argument count
 * @argv: Command line arguments
 * @error: Return location for error
 * 
 * Initializes the application.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_app_initialize(TaishangApp *app, int argc, char **argv, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_APP(app), FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    if (app_initialized) {
        g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_INVALID_ARGUMENT,
                   "Application already initialized");
        return FALSE;
    }
    
    /* Parse command line arguments */
    GOptionContext *context = g_option_context_new("- TaishangLaojun Desktop Application");
    
    GOptionEntry entries[] = {
        { "debug", 'd', 0, G_OPTION_ARG_NONE, &priv->debug_mode,
          "Enable debug mode", NULL },
        { "verbose", 'v', 0, G_OPTION_ARG_NONE, &priv->verbose_mode,
          "Enable verbose logging", NULL },
        { "headless", 'h', 0, G_OPTION_ARG_NONE, &priv->headless_mode,
          "Run without UI", NULL },
        { NULL }
    };
    
    g_option_context_add_main_entries(context, entries, NULL);
    g_option_context_add_group(context, gtk_get_option_group(TRUE));
    
    if (!g_option_context_parse(context, &argc, &argv, error)) {
        g_option_context_free(context);
        return FALSE;
    }
    
    g_option_context_free(context);
    
    /* Initialize GTK if not in headless mode */
    if (!priv->headless_mode) {
        if (!gtk_init_check(&argc, &argv)) {
            g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_NOT_IMPLEMENTED,
                       "Failed to initialize GTK");
            return FALSE;
        }
    }
    
    /* Initialize internationalization */
    taishang_init_i18n();
    
    /* Set up directories */
    if (!taishang_app_setup_directories(app, error)) {
        return FALSE;
    }
    
    /* Set up logging */
    if (!taishang_app_setup_logging(app, error)) {
        return FALSE;
    }
    
    /* Load configuration */
    if (!taishang_app_load_configuration(app, error)) {
        return FALSE;
    }
    
    /* Set up UI if not in headless mode */
    if (!priv->headless_mode) {
        if (!taishang_app_setup_ui(app, error)) {
            return FALSE;
        }
    }
    
    /* Set up plugins */
    if (!taishang_app_setup_plugins(app, error)) {
        return FALSE;
    }
    
    /* Set up signal handlers */
    taishang_app_setup_signal_handlers(app);
    
    /* Set up auto-save */
    priv->auto_save_timeout_id = g_timeout_add_seconds(300, /* 5 minutes */
                                                      taishang_app_auto_save_callback,
                                                      app);
    
    /* Create main loop */
    priv->main_loop = g_main_loop_new(NULL, FALSE);
    
    /* Update state */
    taishang_app_set_state(app, TAISHANG_APP_STATE_INITIALIZED);
    
    app_initialized = TRUE;
    
    g_signal_emit(app, signals[SIGNAL_STARTUP], 0);
    
    g_info("TaishangLaojun application initialized successfully");
    
    return TRUE;
}

/**
 * taishang_app_run:
 * @app: A TaishangApp instance
 * 
 * Runs the application main loop.
 * 
 * Returns: Exit code
 */
gint taishang_app_run(TaishangApp *app) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_APP(app), 1);
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    if (priv->state != TAISHANG_APP_STATE_INITIALIZED) {
        g_warning("Application not properly initialized");
        return 1;
    }
    
    taishang_app_set_state(app, TAISHANG_APP_STATE_RUNNING);
    
    g_info("Starting TaishangLaojun application");
    
    /* Show UI if not in headless mode */
    if (!priv->headless_mode && priv->ui) {
        taishang_ui_show(priv->ui);
    }
    
    /* Run main loop */
    g_main_loop_run(priv->main_loop);
    
    g_info("TaishangLaojun application finished with exit code %d", priv->exit_code);
    
    return priv->exit_code;
}

/**
 * taishang_app_shutdown:
 * @app: A TaishangApp instance
 * @exit_code: Exit code
 * 
 * Shuts down the application.
 */
void taishang_app_shutdown(TaishangApp *app, gint exit_code) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_APP(app));
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    if (priv->shutdown_requested) {
        return;
    }
    
    priv->shutdown_requested = TRUE;
    priv->exit_code = exit_code;
    
    g_signal_emit(app, signals[SIGNAL_SHUTDOWN], 0);
    
    taishang_app_set_state(app, TAISHANG_APP_STATE_SHUTTING_DOWN);
    
    g_info("Shutting down TaishangLaojun application");
    
    /* Save configuration */
    if (priv->config) {
        GError *error = NULL;
        if (!taishang_config_save(priv->config, &error)) {
            g_warning("Failed to save configuration: %s", error->message);
            g_error_free(error);
        }
    }
    
    /* Hide UI */
    if (priv->ui) {
        taishang_ui_hide(priv->ui);
    }
    
    /* Stop main loop */
    if (priv->main_loop && g_main_loop_is_running(priv->main_loop)) {
        g_main_loop_quit(priv->main_loop);
    }
    
    taishang_app_set_state(app, TAISHANG_APP_STATE_SHUTDOWN);
}

/**
 * taishang_app_get_state:
 * @app: A TaishangApp instance
 * 
 * Gets the current application state.
 * 
 * Returns: The current state
 */
TaishangAppState taishang_app_get_state(TaishangApp *app) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_APP(app), TAISHANG_APP_STATE_UNINITIALIZED);
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    g_mutex_lock(&priv->state_mutex);
    TaishangAppState state = priv->state;
    g_mutex_unlock(&priv->state_mutex);
    
    return state;
}

/**
 * taishang_app_set_state:
 * @app: A TaishangApp instance
 * @state: New state
 * 
 * Sets the application state.
 */
void taishang_app_set_state(TaishangApp *app, TaishangAppState state) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_APP(app));
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    g_mutex_lock(&priv->state_mutex);
    
    if (priv->state != state) {
        TaishangAppState old_state = priv->state;
        priv->state = state;
        
        g_mutex_unlock(&priv->state_mutex);
        
        g_signal_emit(app, signals[SIGNAL_STATE_CHANGED], 0, state);
        g_object_notify_by_pspec(G_OBJECT(app), properties[PROP_STATE]);
        
        g_debug("Application state changed from %d to %d", old_state, state);
        
        g_cond_broadcast(&priv->state_cond);
    } else {
        g_mutex_unlock(&priv->state_mutex);
    }
}

/**
 * taishang_app_get_config:
 * @app: A TaishangApp instance
 * 
 * Gets the application configuration.
 * 
 * Returns: (transfer none): The configuration object
 */
TaishangConfig *taishang_app_get_config(TaishangApp *app) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_APP(app), NULL);
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    return priv->config;
}

/**
 * taishang_app_get_ui:
 * @app: A TaishangApp instance
 * 
 * Gets the application UI.
 * 
 * Returns: (transfer none): The UI object
 */
TaishangUI *taishang_app_get_ui(TaishangApp *app) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_APP(app), NULL);
    
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    return priv->ui;
}

/* Private helper functions */

static gboolean taishang_app_setup_directories(TaishangApp *app, GError **error) {
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    /* Get user directories */
    const gchar *home_dir = g_get_home_dir();
    const gchar *config_home = g_get_user_config_dir();
    const gchar *data_home = g_get_user_data_dir();
    const gchar *cache_home = g_get_user_cache_dir();
    
    /* Set up application directories */
    priv->config_dir = g_build_filename(config_home, "taishang-laojun", NULL);
    priv->data_dir = g_build_filename(data_home, "taishang-laojun", NULL);
    priv->cache_dir = g_build_filename(cache_home, "taishang-laojun", NULL);
    
    /* Create directories if they don't exist */
    if (!taishang_utils_create_directory(priv->config_dir, error) ||
        !taishang_utils_create_directory(priv->data_dir, error) ||
        !taishang_utils_create_directory(priv->cache_dir, error)) {
        return FALSE;
    }
    
    g_debug("Application directories set up successfully");
    g_debug("Config dir: %s", priv->config_dir);
    g_debug("Data dir: %s", priv->data_dir);
    g_debug("Cache dir: %s", priv->cache_dir);
    
    return TRUE;
}

static gboolean taishang_app_setup_logging(TaishangApp *app, GError **error) {
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    /* Set up log file */
    priv->log_file = g_build_filename(priv->cache_dir, "taishang-laojun.log", NULL);
    
    /* Initialize logging */
    taishang_init_logging();
    
    /* Set log level based on debug/verbose mode */
    if (priv->debug_mode) {
        g_setenv("G_MESSAGES_DEBUG", "all", TRUE);
    } else if (priv->verbose_mode) {
        g_setenv("G_MESSAGES_DEBUG", "taishang-laojun", TRUE);
    }
    
    g_info("Logging initialized, log file: %s", priv->log_file);
    
    return TRUE;
}

static gboolean taishang_app_load_configuration(TaishangApp *app, GError **error) {
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    /* Create configuration object */
    priv->config = taishang_config_new();
    
    /* Set configuration file path */
    gchar *config_file = g_build_filename(priv->config_dir, "config.json", NULL);
    taishang_config_set_file(priv->config, config_file);
    g_free(config_file);
    
    /* Load configuration */
    if (!taishang_config_load(priv->config, error)) {
        return FALSE;
    }
    
    /* Connect to configuration change signal */
    g_signal_connect(priv->config, "changed",
                    G_CALLBACK(on_config_changed), app);
    
    g_info("Configuration loaded successfully");
    
    return TRUE;
}

static gboolean taishang_app_setup_ui(TaishangApp *app, GError **error) {
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    /* Create UI object */
    priv->ui = taishang_ui_new();
    
    /* Set configuration */
    taishang_ui_set_config(priv->ui, priv->config);
    
    /* Initialize UI */
    if (!taishang_ui_initialize(priv->ui, error)) {
        return FALSE;
    }
    
    /* Connect to UI signals */
    g_signal_connect(priv->ui, "close-request",
                    G_CALLBACK(on_ui_close_request), app);
    
    g_info("UI initialized successfully");
    
    return TRUE;
}

static gboolean taishang_app_setup_plugins(TaishangApp *app, GError **error) {
    /* Plugin system not implemented yet */
    g_debug("Plugin system not implemented");
    return TRUE;
}

static void taishang_app_setup_signal_handlers(TaishangApp *app) {
    /* Set up Unix signal handlers */
    signal(SIGINT, on_unix_signal);
    signal(SIGTERM, on_unix_signal);
    signal(SIGHUP, on_unix_signal);
    
    g_debug("Signal handlers set up");
}

static gboolean taishang_app_auto_save_callback(gpointer user_data) {
    TaishangApp *app = TAISHANG_APP(user_data);
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    if (priv->config) {
        GError *error = NULL;
        if (!taishang_config_save(priv->config, &error)) {
            g_warning("Auto-save failed: %s", error->message);
            g_error_free(error);
        } else {
            g_debug("Configuration auto-saved");
        }
    }
    
    return G_SOURCE_CONTINUE;
}

static void taishang_app_cleanup_resources(TaishangApp *app) {
    TaishangAppPrivate *priv = taishang_app_get_instance_private(app);
    
    /* Cleanup will be done in finalize */
    g_debug("Cleaning up application resources");
}

/* Signal handlers */

static void on_unix_signal(int signum) {
    g_info("Received signal %d, shutting down", signum);
    
    if (app_instance) {
        taishang_app_shutdown(app_instance, 0);
    }
}

static void on_config_changed(TaishangConfig *config, const gchar *key, gpointer user_data) {
    TaishangApp *app = TAISHANG_APP(user_data);
    
    g_signal_emit(app, signals[SIGNAL_CONFIG_CHANGED], 0, key);
    
    g_debug("Configuration changed: %s", key);
}

static void on_ui_close_request(TaishangUI *ui, gpointer user_data) {
    TaishangApp *app = TAISHANG_APP(user_data);
    
    taishang_app_shutdown(app, 0);
}

/* Error handling */

GQuark taishang_error_quark(void) {
    return g_quark_from_static_string("taishang-error-quark");
}

const gchar *taishang_error_to_string(TaishangError error) {
    switch (error) {
        case TAISHANG_ERROR_NONE:
            return "No error";
        case TAISHANG_ERROR_INVALID_ARGUMENT:
            return "Invalid argument";
        case TAISHANG_ERROR_FILE_NOT_FOUND:
            return "File not found";
        case TAISHANG_ERROR_PERMISSION_DENIED:
            return "Permission denied";
        case TAISHANG_ERROR_OUT_OF_MEMORY:
            return "Out of memory";
        case TAISHANG_ERROR_NETWORK_ERROR:
            return "Network error";
        case TAISHANG_ERROR_TIMEOUT:
            return "Timeout";
        case TAISHANG_ERROR_CANCELLED:
            return "Cancelled";
        case TAISHANG_ERROR_NOT_IMPLEMENTED:
            return "Not implemented";
        case TAISHANG_ERROR_UNKNOWN:
        default:
            return "Unknown error";
    }
}

/* Version information */

const gchar *taishang_get_version(void) {
    return TAISHANG_VERSION;
}

gint taishang_get_major_version(void) {
    return 1;
}

gint taishang_get_minor_version(void) {
    return 0;
}

gint taishang_get_micro_version(void) {
    return 0;
}

const gchar *taishang_get_build_date(void) {
    return __DATE__;
}

const gchar *taishang_get_build_time(void) {
    return __TIME__;
}

const gchar *taishang_get_compiler_info(void) {
#ifdef __GNUC__
    return "GCC " __VERSION__;
#elif defined(__clang__)
    return "Clang " __clang_version__;
#else
    return "Unknown compiler";
#endif
}

gboolean taishang_check_version(gint required_major, gint required_minor, gint required_micro) {
    gint major = taishang_get_major_version();
    gint minor = taishang_get_minor_version();
    gint micro = taishang_get_micro_version();
    
    if (major > required_major) return TRUE;
    if (major < required_major) return FALSE;
    
    if (minor > required_minor) return TRUE;
    if (minor < required_minor) return FALSE;
    
    return micro >= required_micro;
}

/* Internationalization */

void taishang_init_i18n(void) {
    setlocale(LC_ALL, "");
    bindtextdomain("taishang-laojun", TAISHANG_LOCALEDIR);
    textdomain("taishang-laojun");
    
    g_debug("Internationalization initialized");
}

const gchar *taishang_get_locale(void) {
    return setlocale(LC_ALL, NULL);
}

void taishang_set_locale(const gchar *locale) {
    setlocale(LC_ALL, locale);
    g_debug("Locale set to: %s", locale);
}

/* Logging */

void taishang_init_logging(void) {
    /* Set up custom log handler if needed */
    g_debug("Logging system initialized");
}

void taishang_cleanup_logging(void) {
    g_debug("Logging system cleaned up");
}

/* Main function */

int main(int argc, char **argv) {
    TaishangApp *app;
    GError *error = NULL;
    gint exit_code = 0;
    
    /* Create application */
    app = taishang_app_get_default();
    
    /* Initialize application */
    if (!taishang_app_initialize(app, argc, argv, &error)) {
        g_printerr("Failed to initialize application: %s\n", error->message);
        g_error_free(error);
        exit_code = 1;
        goto cleanup;
    }
    
    /* Run application */
    exit_code = taishang_app_run(app);
    
cleanup:
    /* Cleanup */
    if (app) {
        taishang_app_shutdown(app, exit_code);
        g_object_unref(app);
    }
    
    return exit_code;
}