/**
 * @file ui.h
 * @brief TaishangLaojun Desktop Application UI Header
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains the user interface structures, functions, and definitions
 * for the TaishangLaojun desktop application on Linux.
 */

#ifndef TAISHANG_UI_H
#define TAISHANG_UI_H

#include "common.h"
#include "config.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Forward declarations */
typedef struct _TaishangUI TaishangUI;
typedef struct _TaishangUIClass TaishangUIClass;
typedef struct _TaishangMainWindow TaishangMainWindow;
typedef struct _TaishangChatView TaishangChatView;
typedef struct _TaishangSidebar TaishangSidebar;

/* UI theme enumeration */
typedef enum {
    TAISHANG_UI_THEME_SYSTEM = 0,
    TAISHANG_UI_THEME_LIGHT,
    TAISHANG_UI_THEME_DARK,
    TAISHANG_UI_THEME_HIGH_CONTRAST
} TaishangUITheme;

/* UI state enumeration */
typedef enum {
    TAISHANG_UI_STATE_UNINITIALIZED = 0,
    TAISHANG_UI_STATE_INITIALIZING,
    TAISHANG_UI_STATE_READY,
    TAISHANG_UI_STATE_BUSY,
    TAISHANG_UI_STATE_ERROR
} TaishangUIState;

/* Window state enumeration */
typedef enum {
    TAISHANG_WINDOW_STATE_NORMAL = 0,
    TAISHANG_WINDOW_STATE_MINIMIZED,
    TAISHANG_WINDOW_STATE_MAXIMIZED,
    TAISHANG_WINDOW_STATE_FULLSCREEN,
    TAISHANG_WINDOW_STATE_HIDDEN
} TaishangWindowState;

/* UI structure */
struct _TaishangUI {
    GObject parent;
    
    /* UI state */
    TaishangUIState state;
    TaishangUITheme theme;
    
    /* Main components */
    TaishangMainWindow *main_window;
    TaishangChatView *chat_view;
    TaishangSidebar *sidebar;
    
    /* GTK widgets */
    GtkApplication *app;
    GtkWindow *window;
    GtkHeaderBar *header_bar;
    GtkMenuBar *menu_bar;
    GtkToolbar *toolbar;
    GtkStatusbar *status_bar;
    GtkPaned *main_paned;
    GtkStack *main_stack;
    GtkStackSwitcher *stack_switcher;
    
    /* Dialogs */
    GtkDialog *preferences_dialog;
    GtkDialog *about_dialog;
    GtkDialog *connect_dialog;
    GtkFileChooserDialog *file_chooser;
    
    /* UI settings */
    gint window_width;
    gint window_height;
    gint window_x;
    gint window_y;
    gboolean window_maximized;
    gboolean show_toolbar;
    gboolean show_status_bar;
    gboolean show_sidebar;
    
    /* Theme and styling */
    GtkCssProvider *css_provider;
    gchar *theme_name;
    gchar *icon_theme_name;
    
    /* Configuration */
    TaishangConfig *config;
    
    /* Private data */
    gpointer priv;
};

/* UI class structure */
struct _TaishangUIClass {
    GObjectClass parent_class;
    
    /* Virtual methods */
    gboolean (*initialize)(TaishangUI *ui, GtkApplication *app);
    void (*finalize)(TaishangUI *ui);
    void (*show)(TaishangUI *ui);
    void (*hide)(TaishangUI *ui);
    void (*update)(TaishangUI *ui);
    
    /* Signal handlers */
    void (*theme_changed)(TaishangUI *ui, TaishangUITheme theme);
    void (*state_changed)(TaishangUI *ui, TaishangUIState old_state, TaishangUIState new_state);
    void (*window_state_changed)(TaishangUI *ui, TaishangWindowState state);
    
    /* Reserved for future use */
    gpointer reserved[8];
};

/* GType macros */
#define TAISHANG_TYPE_UI            (taishang_ui_get_type())
#define TAISHANG_UI(obj)            (G_TYPE_CHECK_INSTANCE_CAST((obj), TAISHANG_TYPE_UI, TaishangUI))
#define TAISHANG_UI_CLASS(klass)    (G_TYPE_CHECK_CLASS_CAST((klass), TAISHANG_TYPE_UI, TaishangUIClass))
#define TAISHANG_IS_UI(obj)         (G_TYPE_CHECK_INSTANCE_TYPE((obj), TAISHANG_TYPE_UI))
#define TAISHANG_IS_UI_CLASS(klass) (G_TYPE_CHECK_CLASS_TYPE((klass), TAISHANG_TYPE_UI))
#define TAISHANG_UI_GET_CLASS(obj)  (G_TYPE_INSTANCE_GET_CLASS((obj), TAISHANG_TYPE_UI, TaishangUIClass))

/* UI lifecycle functions */
GType taishang_ui_get_type(void) G_GNUC_CONST;
TaishangUI *taishang_ui_new(void);
gboolean taishang_ui_initialize(TaishangUI *ui, GtkApplication *app);
void taishang_ui_finalize(TaishangUI *ui);

/* UI display functions */
void taishang_ui_show(TaishangUI *ui);
void taishang_ui_hide(TaishangUI *ui);
void taishang_ui_present(TaishangUI *ui);
void taishang_ui_minimize(TaishangUI *ui);
void taishang_ui_maximize(TaishangUI *ui);
void taishang_ui_fullscreen(TaishangUI *ui);
void taishang_ui_unfullscreen(TaishangUI *ui);

/* UI state management */
TaishangUIState taishang_ui_get_state(TaishangUI *ui);
void taishang_ui_set_state(TaishangUI *ui, TaishangUIState state);
const gchar *taishang_ui_state_to_string(TaishangUIState state);

/* Window management */
GtkWindow *taishang_ui_get_main_window(TaishangUI *ui);
TaishangWindowState taishang_ui_get_window_state(TaishangUI *ui);
void taishang_ui_set_window_state(TaishangUI *ui, TaishangWindowState state);

gboolean taishang_ui_get_window_geometry(TaishangUI *ui, gint *x, gint *y, gint *width, gint *height);
void taishang_ui_set_window_geometry(TaishangUI *ui, gint x, gint y, gint width, gint height);

gboolean taishang_ui_is_window_maximized(TaishangUI *ui);
void taishang_ui_set_window_maximized(TaishangUI *ui, gboolean maximized);

/* Theme management */
TaishangUITheme taishang_ui_get_theme(TaishangUI *ui);
void taishang_ui_set_theme(TaishangUI *ui, TaishangUITheme theme);
const gchar *taishang_ui_theme_to_string(TaishangUITheme theme);
TaishangUITheme taishang_ui_theme_from_string(const gchar *theme_name);

gboolean taishang_ui_load_css(TaishangUI *ui, const gchar *css_file);
void taishang_ui_apply_theme(TaishangUI *ui);

/* Component access */
TaishangMainWindow *taishang_ui_get_main_window_component(TaishangUI *ui);
TaishangChatView *taishang_ui_get_chat_view(TaishangUI *ui);
TaishangSidebar *taishang_ui_get_sidebar(TaishangUI *ui);

GtkHeaderBar *taishang_ui_get_header_bar(TaishangUI *ui);
GtkMenuBar *taishang_ui_get_menu_bar(TaishangUI *ui);
GtkToolbar *taishang_ui_get_toolbar(TaishangUI *ui);
GtkStatusbar *taishang_ui_get_status_bar(TaishangUI *ui);

/* UI element visibility */
gboolean taishang_ui_get_toolbar_visible(TaishangUI *ui);
void taishang_ui_set_toolbar_visible(TaishangUI *ui, gboolean visible);

gboolean taishang_ui_get_status_bar_visible(TaishangUI *ui);
void taishang_ui_set_status_bar_visible(TaishangUI *ui, gboolean visible);

gboolean taishang_ui_get_sidebar_visible(TaishangUI *ui);
void taishang_ui_set_sidebar_visible(TaishangUI *ui, gboolean visible);

/* Dialog management */
void taishang_ui_show_preferences_dialog(TaishangUI *ui);
void taishang_ui_show_about_dialog(TaishangUI *ui);
void taishang_ui_show_connect_dialog(TaishangUI *ui);

GtkFileChooserDialog *taishang_ui_create_file_chooser(TaishangUI *ui, 
                                                      const gchar *title,
                                                      GtkFileChooserAction action);

/* Status bar functions */
guint taishang_ui_status_bar_push(TaishangUI *ui, const gchar *message);
void taishang_ui_status_bar_pop(TaishangUI *ui, guint context_id);
void taishang_ui_status_bar_clear(TaishangUI *ui);

/* Progress indication */
void taishang_ui_show_progress(TaishangUI *ui, const gchar *message);
void taishang_ui_update_progress(TaishangUI *ui, gdouble fraction, const gchar *message);
void taishang_ui_hide_progress(TaishangUI *ui);

/* Notification functions */
void taishang_ui_show_notification(TaishangUI *ui, const gchar *title, const gchar *message);
void taishang_ui_show_error_message(TaishangUI *ui, const gchar *title, const gchar *message);
void taishang_ui_show_warning_message(TaishangUI *ui, const gchar *title, const gchar *message);
void taishang_ui_show_info_message(TaishangUI *ui, const gchar *title, const gchar *message);

gboolean taishang_ui_show_question_dialog(TaishangUI *ui, const gchar *title, const gchar *message);

/* Menu and toolbar functions */
void taishang_ui_update_menu_sensitivity(TaishangUI *ui);
void taishang_ui_update_toolbar_sensitivity(TaishangUI *ui);

/* Configuration integration */
void taishang_ui_load_settings(TaishangUI *ui);
void taishang_ui_save_settings(TaishangUI *ui);
void taishang_ui_apply_config(TaishangUI *ui, TaishangConfig *config);

/* Accessibility support */
void taishang_ui_setup_accessibility(TaishangUI *ui);
void taishang_ui_update_accessibility(TaishangUI *ui);

/* Keyboard shortcuts */
void taishang_ui_setup_accelerators(TaishangUI *ui);
void taishang_ui_add_accelerator(TaishangUI *ui, const gchar *accel, const gchar *action);

/* Drag and drop support */
void taishang_ui_setup_drag_and_drop(TaishangUI *ui);

/* Signal emission helpers */
void taishang_ui_emit_theme_changed(TaishangUI *ui, TaishangUITheme theme);
void taishang_ui_emit_state_changed(TaishangUI *ui, TaishangUIState old_state, TaishangUIState new_state);
void taishang_ui_emit_window_state_changed(TaishangUI *ui, TaishangWindowState state);

/* Utility functions */
gboolean taishang_ui_is_initialized(TaishangUI *ui);
gboolean taishang_ui_is_visible(TaishangUI *ui);
gboolean taishang_ui_is_ready(TaishangUI *ui);

/* Resource management */
gboolean taishang_ui_load_ui_file(TaishangUI *ui, const gchar *ui_file);
GdkPixbuf *taishang_ui_load_icon(TaishangUI *ui, const gchar *icon_name, gint size);

/* Signal definitions */
#define TAISHANG_UI_SIGNAL_THEME_CHANGED        "theme-changed"
#define TAISHANG_UI_SIGNAL_STATE_CHANGED        "state-changed"
#define TAISHANG_UI_SIGNAL_WINDOW_STATE_CHANGED "window-state-changed"

/* CSS class names */
#define TAISHANG_UI_CSS_CLASS_MAIN_WINDOW       "taishang-main-window"
#define TAISHANG_UI_CSS_CLASS_HEADER_BAR        "taishang-header-bar"
#define TAISHANG_UI_CSS_CLASS_SIDEBAR           "taishang-sidebar"
#define TAISHANG_UI_CSS_CLASS_CHAT_VIEW         "taishang-chat-view"
#define TAISHANG_UI_CSS_CLASS_STATUS_BAR        "taishang-status-bar"

/* Default UI settings */
#define TAISHANG_UI_DEFAULT_WINDOW_WIDTH        1200
#define TAISHANG_UI_DEFAULT_WINDOW_HEIGHT       800
#define TAISHANG_UI_MIN_WINDOW_WIDTH            800
#define TAISHANG_UI_MIN_WINDOW_HEIGHT           600

#ifdef __cplusplus
}
#endif

#endif /* TAISHANG_UI_H */