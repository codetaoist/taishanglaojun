#ifndef APPLICATION_H
#define APPLICATION_H

#include <gtk/gtk.h>
#include <adwaita.h>
#include <glib.h>
#include <gio/gio.h>

G_BEGIN_DECLS

#define TAISHANG_TYPE_APPLICATION (taishang_application_get_type())
G_DECLARE_FINAL_TYPE(TaishangApplication, taishang_application, TAISHANG, APPLICATION, AdwApplication)

// 应用程序状态
typedef enum {
    TAISHANG_APP_STATE_INITIALIZING,
    TAISHANG_APP_STATE_RUNNING,
    TAISHANG_APP_STATE_MINIMIZED,
    TAISHANG_APP_STATE_CLOSING
} TaishangAppState;

// 应用程序配置
typedef struct {
    gboolean enable_desktop_pet;
    gboolean enable_notifications;
    gboolean enable_system_tray;
    gboolean auto_start;
    gchar *theme_name;
    gint window_width;
    gint window_height;
    gboolean window_maximized;
} TaishangAppConfig;

// 应用程序结构
struct _TaishangApplication {
    AdwApplication parent_instance;
    
    // 应用程序状态
    TaishangAppState state;
    TaishangAppConfig *config;
    
    // 主窗口
    GtkWindow *main_window;
    
    // UI组件
    AdwHeaderBar *header_bar;
    AdwViewStack *view_stack;
    AdwViewSwitcher *view_switcher;
    GtkBox *main_box;
    
    // 视图页面
    GtkWidget *project_page;
    GtkWidget *chat_page;
    GtkWidget *transfer_page;
    GtkWidget *settings_page;
    
    // 桌面宠物
    GtkWindow *pet_window;
    gboolean pet_visible;
    
    // 系统托盘
    GtkStatusIcon *status_icon;
    GtkMenu *status_menu;
    
    // 网络管理
    GSocketService *socket_service;
    GSocketConnection *connection;
    
    // 定时器
    guint update_timer;
    guint pet_animation_timer;
};

// 公共函数
TaishangApplication *taishang_application_new(void);
void taishang_application_show_main_window(TaishangApplication *app);
void taishang_application_hide_main_window(TaishangApplication *app);
void taishang_application_toggle_desktop_pet(TaishangApplication *app);
void taishang_application_show_preferences(TaishangApplication *app);
void taishang_application_quit(TaishangApplication *app);

// 配置管理
TaishangAppConfig *taishang_app_config_new(void);
void taishang_app_config_free(TaishangAppConfig *config);
gboolean taishang_app_config_load(TaishangAppConfig *config);
gboolean taishang_app_config_save(TaishangAppConfig *config);

// 窗口管理
void taishang_application_setup_main_window(TaishangApplication *app);
void taishang_application_setup_desktop_pet(TaishangApplication *app);
void taishang_application_setup_status_icon(TaishangApplication *app);

// 页面管理
void taishang_application_setup_project_page(TaishangApplication *app);
void taishang_application_setup_chat_page(TaishangApplication *app);
void taishang_application_setup_transfer_page(TaishangApplication *app);
void taishang_application_setup_settings_page(TaishangApplication *app);

// 事件处理
void taishang_application_on_activate(GApplication *app);
void taishang_application_on_startup(GApplication *app);
void taishang_application_on_shutdown(GApplication *app);
gboolean taishang_application_on_window_delete(GtkWidget *widget, GdkEvent *event, gpointer user_data);

// 动作处理
void taishang_application_action_new_project(GSimpleAction *action, GVariant *parameter, gpointer user_data);
void taishang_application_action_open_project(GSimpleAction *action, GVariant *parameter, gpointer user_data);
void taishang_application_action_preferences(GSimpleAction *action, GVariant *parameter, gpointer user_data);
void taishang_application_action_about(GSimpleAction *action, GVariant *parameter, gpointer user_data);
void taishang_application_action_quit(GSimpleAction *action, GVariant *parameter, gpointer user_data);
void taishang_application_action_toggle_pet(GSimpleAction *action, GVariant *parameter, gpointer user_data);

// 状态图标处理
void taishang_application_status_icon_activate(GtkStatusIcon *status_icon, gpointer user_data);
void taishang_application_status_icon_popup_menu(GtkStatusIcon *status_icon, guint button, guint activate_time, gpointer user_data);

// 定时器回调
gboolean taishang_application_update_timer_callback(gpointer user_data);
gboolean taishang_application_pet_animation_callback(gpointer user_data);

// 工具函数
void taishang_application_show_error_dialog(TaishangApplication *app, const gchar *title, const gchar *message);
void taishang_application_show_info_dialog(TaishangApplication *app, const gchar *title, const gchar *message);
gchar *taishang_application_get_config_dir(void);
gchar *taishang_application_get_data_dir(void);

G_END_DECLS

#endif // APPLICATION_H