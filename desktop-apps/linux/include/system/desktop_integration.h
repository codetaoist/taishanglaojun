#ifndef TAISHANG_DESKTOP_INTEGRATION_H
#define TAISHANG_DESKTOP_INTEGRATION_H

#include <glib.h>
#include <gtk/gtk.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangDesktopIntegration TaishangDesktopIntegration;

// Desktop environment types
typedef enum {
    TAISHANG_DESKTOP_UNKNOWN,
    TAISHANG_DESKTOP_GNOME,
    TAISHANG_DESKTOP_KDE,
    TAISHANG_DESKTOP_XFCE,
    TAISHANG_DESKTOP_MATE,
    TAISHANG_DESKTOP_CINNAMON,
    TAISHANG_DESKTOP_LXDE,
    TAISHANG_DESKTOP_LXQT,
    TAISHANG_DESKTOP_BUDGIE,
    TAISHANG_DESKTOP_PANTHEON,
    TAISHANG_DESKTOP_UNITY,
    TAISHANG_DESKTOP_I3,
    TAISHANG_DESKTOP_SWAY,
    TAISHANG_DESKTOP_OTHER
} TaishangDesktopEnvironment;

// System tray event types
typedef enum {
    TAISHANG_TRAY_ACTIVATE,
    TAISHANG_TRAY_POPUP_MENU,
    TAISHANG_TRAY_SCROLL_UP,
    TAISHANG_TRAY_SCROLL_DOWN,
    TAISHANG_TRAY_MIDDLE_CLICK,
    TAISHANG_TRAY_RIGHT_CLICK
} TaishangTrayEvent;

// Callback function types
typedef void (*TaishangTrayCallback)(TaishangTrayEvent event, gpointer user_data);

// Initialization and cleanup
gboolean taishang_desktop_integration_init(const char *app_name, 
                                           const char *app_id, 
                                           const char *app_version, 
                                           const char *app_description);

void taishang_desktop_integration_cleanup(void);

TaishangDesktopIntegration *taishang_desktop_integration_get_instance(void);

// Desktop file functions
gboolean taishang_desktop_create_desktop_file(const char *name, 
                                              const char *comment, 
                                              const char *exec, 
                                              const char *icon, 
                                              const char *categories);

gboolean taishang_desktop_remove_desktop_file(void);

gboolean taishang_desktop_update_desktop_file(const char *key, const char *value);

// Autostart functions
gboolean taishang_desktop_enable_autostart(void);
gboolean taishang_desktop_disable_autostart(void);
gboolean taishang_desktop_is_autostart_enabled(void);

// System tray functions
gboolean taishang_desktop_show_tray_icon(void);
gboolean taishang_desktop_hide_tray_icon(void);
gboolean taishang_desktop_set_tray_icon(const char *icon_name);
gboolean taishang_desktop_set_tray_tooltip(const char *tooltip);
void taishang_desktop_set_tray_callback(TaishangTrayCallback callback, void *user_data);

// Desktop environment functions
TaishangDesktopEnvironment taishang_desktop_get_environment(void);
const char *taishang_desktop_get_session_type(void);
const char *taishang_desktop_get_desktop_session(void);
gboolean taishang_desktop_is_wayland(void);
gboolean taishang_desktop_is_x11(void);

// Utility functions
gboolean taishang_desktop_open_file(const char *file_path);
gboolean taishang_desktop_open_url(const char *url);
gboolean taishang_desktop_show_in_file_manager(const char *file_path);

// Desktop environment detection helpers
const char *taishang_desktop_environment_to_string(TaishangDesktopEnvironment env);
gboolean taishang_desktop_supports_system_tray(void);
gboolean taishang_desktop_supports_notifications(void);
gboolean taishang_desktop_supports_global_menu(void);

// Application registration
gboolean taishang_desktop_register_mime_type(const char *mime_type, 
                                             const char *description, 
                                             const char *icon, 
                                             const char **extensions);

gboolean taishang_desktop_unregister_mime_type(const char *mime_type);

gboolean taishang_desktop_set_default_application(const char *mime_type);

// Window management helpers
gboolean taishang_desktop_set_window_skip_taskbar(GtkWindow *window, gboolean skip);
gboolean taishang_desktop_set_window_keep_above(GtkWindow *window, gboolean keep_above);
gboolean taishang_desktop_set_window_sticky(GtkWindow *window, gboolean sticky);

// Session management
gboolean taishang_desktop_register_session_client(void);
gboolean taishang_desktop_unregister_session_client(void);

// Convenience macros
#define TAISHANG_DESKTOP_FILE_CATEGORIES_UTILITY "Utility;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_NETWORK "Network;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_OFFICE "Office;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_GRAPHICS "Graphics;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_MULTIMEDIA "AudioVideo;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_DEVELOPMENT "Development;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_GAME "Game;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_EDUCATION "Education;"
#define TAISHANG_DESKTOP_FILE_CATEGORIES_SYSTEM "System;"

// Common desktop file keys
#define TAISHANG_DESKTOP_KEY_NAME "Name"
#define TAISHANG_DESKTOP_KEY_COMMENT "Comment"
#define TAISHANG_DESKTOP_KEY_EXEC "Exec"
#define TAISHANG_DESKTOP_KEY_ICON "Icon"
#define TAISHANG_DESKTOP_KEY_CATEGORIES "Categories"
#define TAISHANG_DESKTOP_KEY_TERMINAL "Terminal"
#define TAISHANG_DESKTOP_KEY_STARTUP_NOTIFY "StartupNotify"
#define TAISHANG_DESKTOP_KEY_HIDDEN "Hidden"

G_END_DECLS

#endif // TAISHANG_DESKTOP_INTEGRATION_H