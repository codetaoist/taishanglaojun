/**
 * @file app.h
 * @brief TaishangLaojun Desktop Application Core Header
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains the core application structures, functions, and definitions
 * for the TaishangLaojun desktop application on Linux.
 */

#ifndef TAISHANG_APP_H
#define TAISHANG_APP_H

#include "common.h"
#include "config.h"
#include "ui.h"
#include "utils.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Application constants */
#define TAISHANG_APP_NAME           "TaishangLaojun"
#define TAISHANG_APP_ID             "com.taishanglaojun.desktop"
#define TAISHANG_APP_DESCRIPTION    "Secure Communication and Project Management Platform"
#define TAISHANG_APP_COPYRIGHT      "Copyright © 2024 TaishangLaojun Team"
#define TAISHANG_APP_LICENSE        "MIT License"
#define TAISHANG_APP_WEBSITE        "https://taishanglaojun.com"

/* Application state enumeration */
typedef enum {
    TAISHANG_APP_STATE_UNINITIALIZED = 0,
    TAISHANG_APP_STATE_INITIALIZING,
    TAISHANG_APP_STATE_RUNNING,
    TAISHANG_APP_STATE_PAUSED,
    TAISHANG_APP_STATE_STOPPING,
    TAISHANG_APP_STATE_STOPPED,
    TAISHANG_APP_STATE_ERROR
} TaishangAppState;

/* Application error codes */
typedef enum {
    TAISHANG_APP_ERROR_NONE = 0,
    TAISHANG_APP_ERROR_INIT_FAILED,
    TAISHANG_APP_ERROR_CONFIG_LOAD_FAILED,
    TAISHANG_APP_ERROR_UI_INIT_FAILED,
    TAISHANG_APP_ERROR_NETWORK_FAILED,
    TAISHANG_APP_ERROR_DATABASE_FAILED,
    TAISHANG_APP_ERROR_PERMISSION_DENIED,
    TAISHANG_APP_ERROR_RESOURCE_NOT_FOUND,
    TAISHANG_APP_ERROR_INVALID_ARGUMENT,
    TAISHANG_APP_ERROR_OUT_OF_MEMORY,
    TAISHANG_APP_ERROR_UNKNOWN
} TaishangAppError;

/* Forward declarations */
typedef struct _TaishangApp TaishangApp;
typedef struct _TaishangAppClass TaishangAppClass;

/* Application structure */
struct _TaishangApp {
    GtkApplication parent;
    
    /* Application state */
    TaishangAppState state;
    TaishangAppError last_error;
    
    /* Configuration */
    TaishangConfig *config;
    
    /* User interface */
    TaishangUI *ui;
    
    /* Application data */
    gchar *app_dir;
    gchar *config_dir;
    gchar *cache_dir;
    gchar *data_dir;
    
    /* Runtime information */
    gint64 start_time;
    gint64 last_activity;
    guint activity_timeout_id;
    
    /* Application flags */
    gboolean debug_mode;
    gboolean verbose_logging;
    gboolean auto_start;
    gboolean minimize_to_tray;
    
    /* Signal handlers */
    gulong activate_handler_id;
    gulong startup_handler_id;
    gulong shutdown_handler_id;
    
    /* Private data */
    gpointer priv;
};

/* Application class structure */
struct _TaishangAppClass {
    GtkApplicationClass parent_class;
    
    /* Virtual methods */
    void (*initialize)(TaishangApp *app);
    void (*finalize)(TaishangApp *app);
    void (*activate)(TaishangApp *app);
    void (*startup)(TaishangApp *app);
    void (*shutdown)(TaishangApp *app);
    
    /* Signal handlers */
    void (*state_changed)(TaishangApp *app, TaishangAppState old_state, TaishangAppState new_state);
    void (*error_occurred)(TaishangApp *app, TaishangAppError error, const gchar *message);
    void (*activity_detected)(TaishangApp *app);
    
    /* Reserved for future use */
    gpointer reserved[8];
};

/* GType macros */
#define TAISHANG_TYPE_APP            (taishang_app_get_type())
#define TAISHANG_APP(obj)            (G_TYPE_CHECK_INSTANCE_CAST((obj), TAISHANG_TYPE_APP, TaishangApp))
#define TAISHANG_APP_CLASS(klass)    (G_TYPE_CHECK_CLASS_CAST((klass), TAISHANG_TYPE_APP, TaishangAppClass))
#define TAISHANG_IS_APP(obj)         (G_TYPE_CHECK_INSTANCE_TYPE((obj), TAISHANG_TYPE_APP))
#define TAISHANG_IS_APP_CLASS(klass) (G_TYPE_CHECK_CLASS_TYPE((klass), TAISHANG_TYPE_APP))
#define TAISHANG_APP_GET_CLASS(obj)  (G_TYPE_INSTANCE_GET_CLASS((obj), TAISHANG_TYPE_APP, TaishangAppClass))

/* Application lifecycle functions */
GType taishang_app_get_type(void) G_GNUC_CONST;
TaishangApp *taishang_app_new(void);
TaishangApp *taishang_app_get_default(void);

gboolean taishang_app_initialize(TaishangApp *app, int argc, char **argv);
gboolean taishang_app_run(TaishangApp *app);
void taishang_app_quit(TaishangApp *app);
void taishang_app_shutdown(TaishangApp *app);

/* Application state management */
TaishangAppState taishang_app_get_state(TaishangApp *app);
void taishang_app_set_state(TaishangApp *app, TaishangAppState state);
const gchar *taishang_app_state_to_string(TaishangAppState state);

/* Error handling */
TaishangAppError taishang_app_get_last_error(TaishangApp *app);
void taishang_app_set_error(TaishangApp *app, TaishangAppError error, const gchar *message);
const gchar *taishang_app_error_to_string(TaishangAppError error);

/* Configuration management */
TaishangConfig *taishang_app_get_config(TaishangApp *app);
gboolean taishang_app_load_config(TaishangApp *app);
gboolean taishang_app_save_config(TaishangApp *app);
gboolean taishang_app_reset_config(TaishangApp *app);

/* Directory management */
const gchar *taishang_app_get_app_dir(TaishangApp *app);
const gchar *taishang_app_get_config_dir(TaishangApp *app);
const gchar *taishang_app_get_cache_dir(TaishangApp *app);
const gchar *taishang_app_get_data_dir(TaishangApp *app);

gboolean taishang_app_ensure_directories(TaishangApp *app);

/* Application information */
const gchar *taishang_app_get_name(void);
const gchar *taishang_app_get_version(void);
const gchar *taishang_app_get_description(void);
const gchar *taishang_app_get_copyright(void);
const gchar *taishang_app_get_license(void);
const gchar *taishang_app_get_website(void);

/* Runtime information */
gint64 taishang_app_get_start_time(TaishangApp *app);
gint64 taishang_app_get_uptime(TaishangApp *app);
gint64 taishang_app_get_last_activity(TaishangApp *app);
void taishang_app_update_activity(TaishangApp *app);

/* Application flags */
gboolean taishang_app_get_debug_mode(TaishangApp *app);
void taishang_app_set_debug_mode(TaishangApp *app, gboolean debug);

gboolean taishang_app_get_verbose_logging(TaishangApp *app);
void taishang_app_set_verbose_logging(TaishangApp *app, gboolean verbose);

gboolean taishang_app_get_auto_start(TaishangApp *app);
void taishang_app_set_auto_start(TaishangApp *app, gboolean auto_start);

gboolean taishang_app_get_minimize_to_tray(TaishangApp *app);
void taishang_app_set_minimize_to_tray(TaishangApp *app, gboolean minimize);

/* Signal emission helpers */
void taishang_app_emit_state_changed(TaishangApp *app, TaishangAppState old_state, TaishangAppState new_state);
void taishang_app_emit_error_occurred(TaishangApp *app, TaishangAppError error, const gchar *message);
void taishang_app_emit_activity_detected(TaishangApp *app);

/* Utility functions */
gboolean taishang_app_is_running(TaishangApp *app);
gboolean taishang_app_is_initialized(TaishangApp *app);
gboolean taishang_app_has_error(TaishangApp *app);

/* Command line handling */
gboolean taishang_app_parse_command_line(TaishangApp *app, int argc, char **argv);
void taishang_app_print_version(void);
void taishang_app_print_help(void);

/* Resource management */
gboolean taishang_app_load_resources(TaishangApp *app);
void taishang_app_unload_resources(TaishangApp *app);

/* Plugin system (future extension) */
gboolean taishang_app_load_plugins(TaishangApp *app);
void taishang_app_unload_plugins(TaishangApp *app);

/* Logging integration */
void taishang_app_log_message(TaishangApp *app, GLogLevelFlags level, const gchar *format, ...) G_GNUC_PRINTF(3, 4);
void taishang_app_log_debug(TaishangApp *app, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_app_log_info(TaishangApp *app, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_app_log_warning(TaishangApp *app, const gchar *format, ...) G_GNUC_PRINTF(2, 3);
void taishang_app_log_error(TaishangApp *app, const gchar *format, ...) G_GNUC_PRINTF(2, 3);

/* Signal definitions */
#define TAISHANG_APP_SIGNAL_STATE_CHANGED    "state-changed"
#define TAISHANG_APP_SIGNAL_ERROR_OCCURRED   "error-occurred"
#define TAISHANG_APP_SIGNAL_ACTIVITY_DETECTED "activity-detected"

#ifdef __cplusplus
}
#endif

#endif /* TAISHANG_APP_H */