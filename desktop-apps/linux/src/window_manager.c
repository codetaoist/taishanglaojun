#include <gtk/gtk.h>
#include <adwaita.h>
#include <glib.h>
#include "../include/application.h"
#include "../include/ui/main_window.h"

// Window manager structure
typedef struct {
    TaishangApplication *app;
    GHashTable *windows;
    TaishangMainWindow *main_window;
    GtkWidget *pet_window;
    GtkWidget *settings_dialog;
    GtkWidget *file_transfer_dialog;
    
    // Window state
    gboolean main_window_visible;
    gboolean pet_window_visible;
    
    // Layout settings
    gint window_width;
    gint window_height;
    gint window_x;
    gint window_y;
    gboolean window_maximized;
} TaishangWindowManager;

static TaishangWindowManager *window_manager = NULL;

// Forward declarations
static void window_manager_init(TaishangApplication *app);
static void window_manager_cleanup(void);
static void save_window_state(GtkWindow *window);
static void restore_window_state(GtkWindow *window);
static void on_window_state_changed(GtkWindow *window, GParamSpec *pspec, gpointer user_data);
static void on_window_size_changed(GtkWindow *window, gint width, gint height, gpointer user_data);

// Public functions
TaishangWindowManager *taishang_window_manager_get_instance(void);
gboolean taishang_window_manager_init(TaishangApplication *app);
void taishang_window_manager_cleanup(void);

// Window management
GtkWidget *taishang_window_manager_get_main_window(void);
GtkWidget *taishang_window_manager_get_pet_window(void);
GtkWidget *taishang_window_manager_show_settings_dialog(GtkWindow *parent);
GtkWidget *taishang_window_manager_show_file_transfer_dialog(GtkWindow *parent);

// Window state management
void taishang_window_manager_show_main_window(void);
void taishang_window_manager_hide_main_window(void);
void taishang_window_manager_toggle_main_window(void);
void taishang_window_manager_show_pet_window(void);
void taishang_window_manager_hide_pet_window(void);
void taishang_window_manager_toggle_pet_window(void);

// Layout management
void taishang_window_manager_save_layout(void);
void taishang_window_manager_restore_layout(void);
void taishang_window_manager_reset_layout(void);

// Implementation
TaishangWindowManager *taishang_window_manager_get_instance(void) {
    return window_manager;
}

gboolean taishang_window_manager_init(TaishangApplication *app) {
    if (window_manager != NULL) {
        g_warning("Window manager already initialized");
        return FALSE;
    }
    
    window_manager = g_new0(TaishangWindowManager, 1);
    window_manager->app = app;
    window_manager->windows = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, NULL);
    
    // Default window settings
    window_manager->window_width = 1200;
    window_manager->window_height = 800;
    window_manager->window_x = -1;
    window_manager->window_y = -1;
    window_manager->window_maximized = FALSE;
    
    // Create main window
    window_manager->main_window = TAISHANG_MAIN_WINDOW(taishang_main_window_new(app));
    g_hash_table_insert(window_manager->windows, g_strdup("main"), window_manager->main_window);
    
    // Connect window signals
    g_signal_connect(window_manager->main_window, "notify::default-width",
                     G_CALLBACK(on_window_state_changed), NULL);
    g_signal_connect(window_manager->main_window, "notify::default-height",
                     G_CALLBACK(on_window_state_changed), NULL);
    g_signal_connect(window_manager->main_window, "notify::maximized",
                     G_CALLBACK(on_window_state_changed), NULL);
    
    // Restore window state
    taishang_window_manager_restore_layout();
    
    g_print("Window manager initialized successfully\n");
    return TRUE;
}

void taishang_window_manager_cleanup(void) {
    if (window_manager == NULL) {
        return;
    }
    
    // Save current layout
    taishang_window_manager_save_layout();
    
    // Cleanup windows
    if (window_manager->windows) {
        g_hash_table_destroy(window_manager->windows);
    }
    
    g_free(window_manager);
    window_manager = NULL;
    
    g_print("Window manager cleaned up\n");
}

// Window management implementations
GtkWidget *taishang_window_manager_get_main_window(void) {
    if (window_manager == NULL) {
        return NULL;
    }
    return GTK_WIDGET(window_manager->main_window);
}

GtkWidget *taishang_window_manager_get_pet_window(void) {
    if (window_manager == NULL) {
        return NULL;
    }
    return window_manager->pet_window;
}

GtkWidget *taishang_window_manager_show_settings_dialog(GtkWindow *parent) {
    if (window_manager == NULL) {
        return NULL;
    }
    
    if (window_manager->settings_dialog == NULL) {
        // Create settings dialog
        window_manager->settings_dialog = gtk_dialog_new_with_buttons(
            "设置",
            parent,
            GTK_DIALOG_MODAL | GTK_DIALOG_DESTROY_WITH_PARENT,
            "取消", GTK_RESPONSE_CANCEL,
            "确定", GTK_RESPONSE_OK,
            NULL
        );
        
        gtk_window_set_default_size(GTK_WINDOW(window_manager->settings_dialog), 600, 400);
        
        // Add settings content
        GtkWidget *content_area = gtk_dialog_get_content_area(GTK_DIALOG(window_manager->settings_dialog));
        GtkWidget *settings_label = gtk_label_new("设置选项将在这里显示");
        gtk_widget_set_margin_top(settings_label, 20);
        gtk_widget_set_margin_bottom(settings_label, 20);
        gtk_widget_set_margin_start(settings_label, 20);
        gtk_widget_set_margin_end(settings_label, 20);
        gtk_box_append(GTK_BOX(content_area), settings_label);
    }
    
    gtk_widget_set_visible(window_manager->settings_dialog, TRUE);
    return window_manager->settings_dialog;
}

GtkWidget *taishang_window_manager_show_file_transfer_dialog(GtkWindow *parent) {
    if (window_manager == NULL) {
        return NULL;
    }
    
    if (window_manager->file_transfer_dialog == NULL) {
        // Create file transfer dialog
        window_manager->file_transfer_dialog = gtk_dialog_new_with_buttons(
            "文件传输",
            parent,
            GTK_DIALOG_MODAL | GTK_DIALOG_DESTROY_WITH_PARENT,
            "关闭", GTK_RESPONSE_CLOSE,
            NULL
        );
        
        gtk_window_set_default_size(GTK_WINDOW(window_manager->file_transfer_dialog), 800, 600);
        
        // Add file transfer content
        GtkWidget *content_area = gtk_dialog_get_content_area(GTK_DIALOG(window_manager->file_transfer_dialog));
        GtkWidget *transfer_label = gtk_label_new("文件传输界面将在这里显示");
        gtk_widget_set_margin_top(transfer_label, 20);
        gtk_widget_set_margin_bottom(transfer_label, 20);
        gtk_widget_set_margin_start(transfer_label, 20);
        gtk_widget_set_margin_end(transfer_label, 20);
        gtk_box_append(GTK_BOX(content_area), transfer_label);
    }
    
    gtk_widget_set_visible(window_manager->file_transfer_dialog, TRUE);
    return window_manager->file_transfer_dialog;
}

// Window state management implementations
void taishang_window_manager_show_main_window(void) {
    if (window_manager == NULL || window_manager->main_window == NULL) {
        return;
    }
    
    gtk_widget_set_visible(GTK_WIDGET(window_manager->main_window), TRUE);
    gtk_window_present(GTK_WINDOW(window_manager->main_window));
    window_manager->main_window_visible = TRUE;
}

void taishang_window_manager_hide_main_window(void) {
    if (window_manager == NULL || window_manager->main_window == NULL) {
        return;
    }
    
    gtk_widget_set_visible(GTK_WIDGET(window_manager->main_window), FALSE);
    window_manager->main_window_visible = FALSE;
}

void taishang_window_manager_toggle_main_window(void) {
    if (window_manager == NULL || window_manager->main_window == NULL) {
        return;
    }
    
    if (window_manager->main_window_visible) {
        taishang_window_manager_hide_main_window();
    } else {
        taishang_window_manager_show_main_window();
    }
}

void taishang_window_manager_show_pet_window(void) {
    if (window_manager == NULL) {
        return;
    }
    
    if (window_manager->pet_window == NULL) {
        // Create pet window if it doesn't exist
        window_manager->pet_window = gtk_window_new();
        gtk_window_set_title(GTK_WINDOW(window_manager->pet_window), "太上老君桌面宠物");
        gtk_window_set_default_size(GTK_WINDOW(window_manager->pet_window), 200, 200);
        gtk_window_set_decorated(GTK_WINDOW(window_manager->pet_window), FALSE);
        gtk_window_set_resizable(GTK_WINDOW(window_manager->pet_window), FALSE);
        
        // Add pet content
        GtkWidget *pet_label = gtk_label_new("🧙‍♂️");
        gtk_widget_set_halign(pet_label, GTK_ALIGN_CENTER);
        gtk_widget_set_valign(pet_label, GTK_ALIGN_CENTER);
        gtk_window_set_child(GTK_WINDOW(window_manager->pet_window), pet_label);
        
        g_hash_table_insert(window_manager->windows, g_strdup("pet"), window_manager->pet_window);
    }
    
    gtk_widget_set_visible(window_manager->pet_window, TRUE);
    window_manager->pet_window_visible = TRUE;
}

void taishang_window_manager_hide_pet_window(void) {
    if (window_manager == NULL || window_manager->pet_window == NULL) {
        return;
    }
    
    gtk_widget_set_visible(window_manager->pet_window, FALSE);
    window_manager->pet_window_visible = FALSE;
}

void taishang_window_manager_toggle_pet_window(void) {
    if (window_manager == NULL) {
        return;
    }
    
    if (window_manager->pet_window_visible) {
        taishang_window_manager_hide_pet_window();
    } else {
        taishang_window_manager_show_pet_window();
    }
}

// Layout management implementations
void taishang_window_manager_save_layout(void) {
    if (window_manager == NULL || window_manager->main_window == NULL) {
        return;
    }
    
    // Get current window state
    GtkWindow *window = GTK_WINDOW(window_manager->main_window);
    
    gtk_window_get_default_size(window, &window_manager->window_width, &window_manager->window_height);
    window_manager->window_maximized = gtk_window_is_maximized(window);
    
    // Save to configuration
    TaishangAppConfig *config = taishang_application_get_config(window_manager->app);
    if (config) {
        config->window_width = window_manager->window_width;
        config->window_height = window_manager->window_height;
        config->window_x = window_manager->window_x;
        config->window_y = window_manager->window_y;
        config->window_maximized = window_manager->window_maximized;
        
        taishang_app_config_save(config);
    }
    
    g_print("Window layout saved: %dx%d, maximized: %s\n",
            window_manager->window_width, window_manager->window_height,
            window_manager->window_maximized ? "yes" : "no");
}

void taishang_window_manager_restore_layout(void) {
    if (window_manager == NULL || window_manager->main_window == NULL) {
        return;
    }
    
    // Load from configuration
    TaishangAppConfig *config = taishang_application_get_config(window_manager->app);
    if (config) {
        window_manager->window_width = config->window_width;
        window_manager->window_height = config->window_height;
        window_manager->window_x = config->window_x;
        window_manager->window_y = config->window_y;
        window_manager->window_maximized = config->window_maximized;
    }
    
    // Apply to window
    GtkWindow *window = GTK_WINDOW(window_manager->main_window);
    gtk_window_set_default_size(window, window_manager->window_width, window_manager->window_height);
    
    if (window_manager->window_maximized) {
        gtk_window_maximize(window);
    }
    
    g_print("Window layout restored: %dx%d, maximized: %s\n",
            window_manager->window_width, window_manager->window_height,
            window_manager->window_maximized ? "yes" : "no");
}

void taishang_window_manager_reset_layout(void) {
    if (window_manager == NULL) {
        return;
    }
    
    // Reset to defaults
    window_manager->window_width = 1200;
    window_manager->window_height = 800;
    window_manager->window_x = -1;
    window_manager->window_y = -1;
    window_manager->window_maximized = FALSE;
    
    // Apply to main window if it exists
    if (window_manager->main_window) {
        GtkWindow *window = GTK_WINDOW(window_manager->main_window);
        gtk_window_set_default_size(window, window_manager->window_width, window_manager->window_height);
        gtk_window_unmaximize(window);
    }
    
    g_print("Window layout reset to defaults\n");
}

// Callback implementations
static void on_window_state_changed(GtkWindow *window, GParamSpec *pspec, gpointer user_data) {
    if (window_manager == NULL) {
        return;
    }
    
    // Auto-save window state changes
    taishang_window_manager_save_layout();
}

static void on_window_size_changed(GtkWindow *window, gint width, gint height, gpointer user_data) {
    if (window_manager == NULL) {
        return;
    }
    
    window_manager->window_width = width;
    window_manager->window_height = height;
    
    // Auto-save size changes
    taishang_window_manager_save_layout();
}