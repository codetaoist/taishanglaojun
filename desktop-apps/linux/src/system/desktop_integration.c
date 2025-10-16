#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <glib.h>
#include <gio/gio.h>
#include <gtk/gtk.h>
#include "../../include/system/desktop_integration.h"
#include "../../include/system/dbus_client.h"

// Desktop integration structure
typedef struct {
    gboolean initialized;
    gchar *app_name;
    gchar *app_id;
    gchar *app_version;
    gchar *app_description;
    gchar *app_icon;
    gchar *app_executable;
    
    // Desktop files
    gchar *desktop_file_path;
    gchar *autostart_file_path;
    
    // System tray
    GtkStatusIcon *status_icon;
    GtkWidget *tray_menu;
    gboolean tray_visible;
    
    // Desktop environment info
    TaishangDesktopEnvironment desktop_env;
    gchar *session_type;
    gchar *desktop_session;
    
    // Callbacks
    TaishangTrayCallback tray_callback;
    void *user_data;
    
    GMutex mutex;
} TaishangDesktopIntegration;

static TaishangDesktopIntegration *desktop_integration = NULL;

// Forward declarations
static gboolean detect_desktop_environment(void);
static gboolean create_desktop_file(void);
static gboolean create_autostart_file(void);
static void setup_system_tray(void);
static void on_tray_activate(GtkStatusIcon *status_icon, gpointer user_data);
static void on_tray_popup_menu(GtkStatusIcon *status_icon, guint button, guint activate_time, gpointer user_data);
static GtkWidget *create_tray_menu(void);

// Public functions
gboolean taishang_desktop_integration_init(const char *app_name, const char *app_id, 
                                           const char *app_version, const char *app_description);
void taishang_desktop_integration_cleanup(void);
TaishangDesktopIntegration *taishang_desktop_integration_get_instance(void);

// Desktop file functions
gboolean taishang_desktop_create_desktop_file(const char *name, const char *comment, 
                                              const char *exec, const char *icon, 
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

// Implementation
gboolean taishang_desktop_integration_init(const char *app_name, const char *app_id, 
                                           const char *app_version, const char *app_description) {
    if (desktop_integration != NULL) {
        g_warning("Desktop integration already initialized");
        return FALSE;
    }
    
    desktop_integration = g_new0(TaishangDesktopIntegration, 1);
    g_mutex_init(&desktop_integration->mutex);
    
    // Set application information
    desktop_integration->app_name = g_strdup(app_name ? app_name : "TaishangApp");
    desktop_integration->app_id = g_strdup(app_id ? app_id : "com.taishang.app");
    desktop_integration->app_version = g_strdup(app_version ? app_version : "1.0.0");
    desktop_integration->app_description = g_strdup(app_description ? app_description : "Taishang Desktop Application");
    
    // Get executable path
    gchar *exe_path = g_file_read_link("/proc/self/exe", NULL);
    desktop_integration->app_executable = exe_path ? exe_path : g_strdup("taishang-app");
    
    // Set default icon
    desktop_integration->app_icon = g_strdup("application-x-executable");
    
    // Set desktop file paths
    const gchar *home_dir = g_get_home_dir();
    desktop_integration->desktop_file_path = g_build_filename(
        home_dir, ".local", "share", "applications", 
        g_strdup_printf("%s.desktop", desktop_integration->app_id), NULL);
    
    desktop_integration->autostart_file_path = g_build_filename(
        home_dir, ".config", "autostart", 
        g_strdup_printf("%s.desktop", desktop_integration->app_id), NULL);
    
    // Detect desktop environment
    if (!detect_desktop_environment()) {
        g_warning("Failed to detect desktop environment");
    }
    
    // Setup system tray
    setup_system_tray();
    
    desktop_integration->initialized = TRUE;
    g_print("Desktop integration initialized for %s\n", desktop_integration->app_name);
    
    return TRUE;
}

void taishang_desktop_integration_cleanup(void) {
    if (desktop_integration == NULL) {
        return;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    
    // Hide tray icon
    if (desktop_integration->status_icon) {
        gtk_status_icon_set_visible(desktop_integration->status_icon, FALSE);
        g_object_unref(desktop_integration->status_icon);
    }
    
    if (desktop_integration->tray_menu) {
        gtk_widget_destroy(desktop_integration->tray_menu);
    }
    
    // Free strings
    g_free(desktop_integration->app_name);
    g_free(desktop_integration->app_id);
    g_free(desktop_integration->app_version);
    g_free(desktop_integration->app_description);
    g_free(desktop_integration->app_icon);
    g_free(desktop_integration->app_executable);
    g_free(desktop_integration->desktop_file_path);
    g_free(desktop_integration->autostart_file_path);
    g_free(desktop_integration->session_type);
    g_free(desktop_integration->desktop_session);
    
    g_mutex_unlock(&desktop_integration->mutex);
    g_mutex_clear(&desktop_integration->mutex);
    
    g_free(desktop_integration);
    desktop_integration = NULL;
    
    g_print("Desktop integration cleaned up\n");
}

TaishangDesktopIntegration *taishang_desktop_integration_get_instance(void) {
    return desktop_integration;
}

// Desktop file functions
gboolean taishang_desktop_create_desktop_file(const char *name, const char *comment, 
                                              const char *exec, const char *icon, 
                                              const char *categories) {
    if (!desktop_integration) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    
    // Create directories if they don't exist
    gchar *dir_path = g_path_get_dirname(desktop_integration->desktop_file_path);
    g_mkdir_with_parents(dir_path, 0755);
    g_free(dir_path);
    
    // Create desktop file content
    GKeyFile *key_file = g_key_file_new();
    
    g_key_file_set_string(key_file, "Desktop Entry", "Type", "Application");
    g_key_file_set_string(key_file, "Desktop Entry", "Version", "1.0");
    g_key_file_set_string(key_file, "Desktop Entry", "Name", name ? name : desktop_integration->app_name);
    g_key_file_set_string(key_file, "Desktop Entry", "Comment", comment ? comment : desktop_integration->app_description);
    g_key_file_set_string(key_file, "Desktop Entry", "Exec", exec ? exec : desktop_integration->app_executable);
    g_key_file_set_string(key_file, "Desktop Entry", "Icon", icon ? icon : desktop_integration->app_icon);
    g_key_file_set_string(key_file, "Desktop Entry", "Categories", categories ? categories : "Utility;Network;");
    g_key_file_set_boolean(key_file, "Desktop Entry", "Terminal", FALSE);
    g_key_file_set_boolean(key_file, "Desktop Entry", "StartupNotify", TRUE);
    
    // Save to file
    GError *error = NULL;
    gboolean success = g_key_file_save_to_file(key_file, desktop_integration->desktop_file_path, &error);
    
    if (!success) {
        g_warning("Failed to create desktop file: %s", error->message);
        g_error_free(error);
    } else {
        // Make file executable
        chmod(desktop_integration->desktop_file_path, 0755);
        g_print("Desktop file created: %s\n", desktop_integration->desktop_file_path);
    }
    
    g_key_file_free(key_file);
    g_mutex_unlock(&desktop_integration->mutex);
    
    return success;
}

gboolean taishang_desktop_remove_desktop_file(void) {
    if (!desktop_integration) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    
    gboolean success = TRUE;
    if (g_file_test(desktop_integration->desktop_file_path, G_FILE_TEST_EXISTS)) {
        if (g_unlink(desktop_integration->desktop_file_path) != 0) {
            g_warning("Failed to remove desktop file: %s", desktop_integration->desktop_file_path);
            success = FALSE;
        } else {
            g_print("Desktop file removed: %s\n", desktop_integration->desktop_file_path);
        }
    }
    
    g_mutex_unlock(&desktop_integration->mutex);
    
    return success;
}

// Autostart functions
gboolean taishang_desktop_enable_autostart(void) {
    if (!desktop_integration) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    
    // Create directories if they don't exist
    gchar *dir_path = g_path_get_dirname(desktop_integration->autostart_file_path);
    g_mkdir_with_parents(dir_path, 0755);
    g_free(dir_path);
    
    // Create autostart desktop file
    GKeyFile *key_file = g_key_file_new();
    
    g_key_file_set_string(key_file, "Desktop Entry", "Type", "Application");
    g_key_file_set_string(key_file, "Desktop Entry", "Version", "1.0");
    g_key_file_set_string(key_file, "Desktop Entry", "Name", desktop_integration->app_name);
    g_key_file_set_string(key_file, "Desktop Entry", "Comment", desktop_integration->app_description);
    g_key_file_set_string(key_file, "Desktop Entry", "Exec", desktop_integration->app_executable);
    g_key_file_set_string(key_file, "Desktop Entry", "Icon", desktop_integration->app_icon);
    g_key_file_set_boolean(key_file, "Desktop Entry", "Terminal", FALSE);
    g_key_file_set_boolean(key_file, "Desktop Entry", "Hidden", FALSE);
    g_key_file_set_boolean(key_file, "Desktop Entry", "X-GNOME-Autostart-enabled", TRUE);
    
    // Save to file
    GError *error = NULL;
    gboolean success = g_key_file_save_to_file(key_file, desktop_integration->autostart_file_path, &error);
    
    if (!success) {
        g_warning("Failed to create autostart file: %s", error->message);
        g_error_free(error);
    } else {
        g_print("Autostart enabled: %s\n", desktop_integration->autostart_file_path);
    }
    
    g_key_file_free(key_file);
    g_mutex_unlock(&desktop_integration->mutex);
    
    return success;
}

gboolean taishang_desktop_disable_autostart(void) {
    if (!desktop_integration) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    
    gboolean success = TRUE;
    if (g_file_test(desktop_integration->autostart_file_path, G_FILE_TEST_EXISTS)) {
        if (g_unlink(desktop_integration->autostart_file_path) != 0) {
            g_warning("Failed to remove autostart file: %s", desktop_integration->autostart_file_path);
            success = FALSE;
        } else {
            g_print("Autostart disabled: %s\n", desktop_integration->autostart_file_path);
        }
    }
    
    g_mutex_unlock(&desktop_integration->mutex);
    
    return success;
}

gboolean taishang_desktop_is_autostart_enabled(void) {
    if (!desktop_integration) {
        return FALSE;
    }
    
    return g_file_test(desktop_integration->autostart_file_path, G_FILE_TEST_EXISTS);
}

// System tray functions
gboolean taishang_desktop_show_tray_icon(void) {
    if (!desktop_integration || !desktop_integration->status_icon) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    gtk_status_icon_set_visible(desktop_integration->status_icon, TRUE);
    desktop_integration->tray_visible = TRUE;
    g_mutex_unlock(&desktop_integration->mutex);
    
    g_print("Tray icon shown\n");
    return TRUE;
}

gboolean taishang_desktop_hide_tray_icon(void) {
    if (!desktop_integration || !desktop_integration->status_icon) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    gtk_status_icon_set_visible(desktop_integration->status_icon, FALSE);
    desktop_integration->tray_visible = FALSE;
    g_mutex_unlock(&desktop_integration->mutex);
    
    g_print("Tray icon hidden\n");
    return TRUE;
}

gboolean taishang_desktop_set_tray_icon(const char *icon_name) {
    if (!desktop_integration || !desktop_integration->status_icon || !icon_name) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    
    if (g_path_is_absolute(icon_name)) {
        gtk_status_icon_set_from_file(desktop_integration->status_icon, icon_name);
    } else {
        gtk_status_icon_set_from_icon_name(desktop_integration->status_icon, icon_name);
    }
    
    g_free(desktop_integration->app_icon);
    desktop_integration->app_icon = g_strdup(icon_name);
    
    g_mutex_unlock(&desktop_integration->mutex);
    
    g_print("Tray icon set: %s\n", icon_name);
    return TRUE;
}

gboolean taishang_desktop_set_tray_tooltip(const char *tooltip) {
    if (!desktop_integration || !desktop_integration->status_icon || !tooltip) {
        return FALSE;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    gtk_status_icon_set_tooltip_text(desktop_integration->status_icon, tooltip);
    g_mutex_unlock(&desktop_integration->mutex);
    
    return TRUE;
}

void taishang_desktop_set_tray_callback(TaishangTrayCallback callback, void *user_data) {
    if (!desktop_integration) {
        return;
    }
    
    g_mutex_lock(&desktop_integration->mutex);
    desktop_integration->tray_callback = callback;
    desktop_integration->user_data = user_data;
    g_mutex_unlock(&desktop_integration->mutex);
}

// Desktop environment functions
TaishangDesktopEnvironment taishang_desktop_get_environment(void) {
    if (!desktop_integration) {
        return TAISHANG_DESKTOP_UNKNOWN;
    }
    
    return desktop_integration->desktop_env;
}

const char *taishang_desktop_get_session_type(void) {
    if (!desktop_integration) {
        return NULL;
    }
    
    return desktop_integration->session_type;
}

const char *taishang_desktop_get_desktop_session(void) {
    if (!desktop_integration) {
        return NULL;
    }
    
    return desktop_integration->desktop_session;
}

gboolean taishang_desktop_is_wayland(void) {
    if (!desktop_integration || !desktop_integration->session_type) {
        return FALSE;
    }
    
    return g_strcmp0(desktop_integration->session_type, "wayland") == 0;
}

gboolean taishang_desktop_is_x11(void) {
    if (!desktop_integration || !desktop_integration->session_type) {
        return FALSE;
    }
    
    return g_strcmp0(desktop_integration->session_type, "x11") == 0;
}

// Utility functions
gboolean taishang_desktop_open_file(const char *file_path) {
    if (!file_path) {
        return FALSE;
    }
    
    GError *error = NULL;
    gchar *uri = g_filename_to_uri(file_path, NULL, &error);
    
    if (!uri) {
        g_warning("Failed to convert file path to URI: %s", error->message);
        g_error_free(error);
        return FALSE;
    }
    
    gboolean success = gtk_show_uri_on_window(NULL, uri, GDK_CURRENT_TIME, &error);
    
    if (!success) {
        g_warning("Failed to open file: %s", error->message);
        g_error_free(error);
    }
    
    g_free(uri);
    return success;
}

gboolean taishang_desktop_open_url(const char *url) {
    if (!url) {
        return FALSE;
    }
    
    GError *error = NULL;
    gboolean success = gtk_show_uri_on_window(NULL, url, GDK_CURRENT_TIME, &error);
    
    if (!success) {
        g_warning("Failed to open URL: %s", error->message);
        g_error_free(error);
    }
    
    return success;
}

// Private functions
static gboolean detect_desktop_environment(void) {
    if (!desktop_integration) {
        return FALSE;
    }
    
    // Get session type
    const gchar *session_type = g_getenv("XDG_SESSION_TYPE");
    if (!session_type) {
        session_type = g_getenv("WAYLAND_DISPLAY") ? "wayland" : "x11";
    }
    desktop_integration->session_type = g_strdup(session_type);
    
    // Get desktop session
    const gchar *desktop_session = g_getenv("XDG_CURRENT_DESKTOP");
    if (!desktop_session) {
        desktop_session = g_getenv("DESKTOP_SESSION");
    }
    desktop_integration->desktop_session = g_strdup(desktop_session);
    
    // Detect desktop environment
    if (desktop_session) {
        if (g_strstr_len(desktop_session, -1, "GNOME")) {
            desktop_integration->desktop_env = TAISHANG_DESKTOP_GNOME;
        } else if (g_strstr_len(desktop_session, -1, "KDE")) {
            desktop_integration->desktop_env = TAISHANG_DESKTOP_KDE;
        } else if (g_strstr_len(desktop_session, -1, "XFCE")) {
            desktop_integration->desktop_env = TAISHANG_DESKTOP_XFCE;
        } else if (g_strstr_len(desktop_session, -1, "MATE")) {
            desktop_integration->desktop_env = TAISHANG_DESKTOP_MATE;
        } else if (g_strstr_len(desktop_session, -1, "Cinnamon")) {
            desktop_integration->desktop_env = TAISHANG_DESKTOP_CINNAMON;
        } else {
            desktop_integration->desktop_env = TAISHANG_DESKTOP_OTHER;
        }
    } else {
        desktop_integration->desktop_env = TAISHANG_DESKTOP_UNKNOWN;
    }
    
    g_print("Desktop environment detected: %s (%s)\n", 
            desktop_session ? desktop_session : "unknown", 
            session_type ? session_type : "unknown");
    
    return TRUE;
}

static void setup_system_tray(void) {
    if (!desktop_integration) {
        return;
    }
    
    // Create status icon
    desktop_integration->status_icon = gtk_status_icon_new_from_icon_name(desktop_integration->app_icon);
    
    if (desktop_integration->status_icon) {
        gtk_status_icon_set_tooltip_text(desktop_integration->status_icon, desktop_integration->app_name);
        gtk_status_icon_set_visible(desktop_integration->status_icon, FALSE);
        
        // Connect signals
        g_signal_connect(desktop_integration->status_icon, "activate", 
                         G_CALLBACK(on_tray_activate), NULL);
        g_signal_connect(desktop_integration->status_icon, "popup-menu", 
                         G_CALLBACK(on_tray_popup_menu), NULL);
        
        // Create tray menu
        desktop_integration->tray_menu = create_tray_menu();
        
        g_print("System tray setup completed\n");
    } else {
        g_warning("Failed to create system tray icon");
    }
}

static void on_tray_activate(GtkStatusIcon *status_icon, gpointer user_data) {
    if (desktop_integration && desktop_integration->tray_callback) {
        desktop_integration->tray_callback(TAISHANG_TRAY_ACTIVATE, desktop_integration->user_data);
    }
}

static void on_tray_popup_menu(GtkStatusIcon *status_icon, guint button, guint activate_time, gpointer user_data) {
    if (desktop_integration && desktop_integration->tray_menu) {
        gtk_menu_popup_at_pointer(GTK_MENU(desktop_integration->tray_menu), NULL);
    }
    
    if (desktop_integration && desktop_integration->tray_callback) {
        desktop_integration->tray_callback(TAISHANG_TRAY_POPUP_MENU, desktop_integration->user_data);
    }
}

static GtkWidget *create_tray_menu(void) {
    GtkWidget *menu = gtk_menu_new();
    
    // Show/Hide item
    GtkWidget *show_item = gtk_menu_item_new_with_label("显示/隐藏");
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), show_item);
    
    // Separator
    GtkWidget *separator = gtk_separator_menu_item_new();
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), separator);
    
    // Settings item
    GtkWidget *settings_item = gtk_menu_item_new_with_label("设置");
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), settings_item);
    
    // About item
    GtkWidget *about_item = gtk_menu_item_new_with_label("关于");
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), about_item);
    
    // Separator
    GtkWidget *separator2 = gtk_separator_menu_item_new();
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), separator2);
    
    // Quit item
    GtkWidget *quit_item = gtk_menu_item_new_with_label("退出");
    gtk_menu_shell_append(GTK_MENU_SHELL(menu), quit_item);
    
    gtk_widget_show_all(menu);
    
    return menu;
}