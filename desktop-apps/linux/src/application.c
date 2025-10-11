#include "application.h"
#include <glib/gi18n.h>
#include <json-c/json.h>

G_DEFINE_TYPE(TaishangApplication, taishang_application, ADW_TYPE_APPLICATION)

// 应用程序动作定义
static const GActionEntry app_entries[] = {
    { "new-project", taishang_application_action_new_project, NULL, NULL, NULL },
    { "open-project", taishang_application_action_open_project, NULL, NULL, NULL },
    { "preferences", taishang_application_action_preferences, NULL, NULL, NULL },
    { "about", taishang_application_action_about, NULL, NULL, NULL },
    { "quit", taishang_application_action_quit, NULL, NULL, NULL },
    { "toggle-pet", taishang_application_action_toggle_pet, NULL, NULL, NULL },
};

// 类初始化
static void taishang_application_class_init(TaishangApplicationClass *klass) {
    GApplicationClass *app_class = G_APPLICATION_CLASS(klass);
    
    app_class->activate = taishang_application_on_activate;
    app_class->startup = taishang_application_on_startup;
    app_class->shutdown = taishang_application_on_shutdown;
}

// 实例初始化
static void taishang_application_init(TaishangApplication *app) {
    app->state = TAISHANG_APP_STATE_INITIALIZING;
    app->config = taishang_app_config_new();
    app->main_window = NULL;
    app->pet_window = NULL;
    app->pet_visible = FALSE;
    app->status_icon = NULL;
    app->socket_service = NULL;
    app->connection = NULL;
    app->update_timer = 0;
    app->pet_animation_timer = 0;
}

// 创建新的应用程序实例
TaishangApplication *taishang_application_new(void) {
    return g_object_new(TAISHANG_TYPE_APPLICATION,
                       "application-id", "com.taishanglaojun.desktop",
                       "flags", G_APPLICATION_HANDLES_OPEN,
                       NULL);
}

// 应用程序启动
void taishang_application_on_startup(GApplication *app) {
    TaishangApplication *self = TAISHANG_APPLICATION(app);
    
    G_APPLICATION_CLASS(taishang_application_parent_class)->startup(app);
    
    // 加载配置
    taishang_app_config_load(self->config);
    
    // 添加动作
    g_action_map_add_action_entries(G_ACTION_MAP(app),
                                   app_entries,
                                   G_N_ELEMENTS(app_entries),
                                   app);
    
    // 设置键盘快捷键
    const gchar *quit_accels[2] = { "<Ctrl>Q", NULL };
    const gchar *new_project_accels[2] = { "<Ctrl>N", NULL };
    const gchar *open_project_accels[2] = { "<Ctrl>O", NULL };
    const gchar *preferences_accels[2] = { "<Ctrl>comma", NULL };
    
    gtk_application_set_accels_for_action(GTK_APPLICATION(app), "app.quit", quit_accels);
    gtk_application_set_accels_for_action(GTK_APPLICATION(app), "app.new-project", new_project_accels);
    gtk_application_set_accels_for_action(GTK_APPLICATION(app), "app.open-project", open_project_accels);
    gtk_application_set_accels_for_action(GTK_APPLICATION(app), "app.preferences", preferences_accels);
    
    // 设置应用程序菜单
    GMenu *app_menu = g_menu_new();
    g_menu_append(app_menu, _("Preferences"), "app.preferences");
    g_menu_append(app_menu, _("About"), "app.about");
    g_menu_append(app_menu, _("Quit"), "app.quit");
    gtk_application_set_app_menu(GTK_APPLICATION(app), G_MENU_MODEL(app_menu));
    g_object_unref(app_menu);
    
    // 设置系统托盘
    if (self->config->enable_system_tray) {
        taishang_application_setup_status_icon(self);
    }
    
    g_print("TaishangLaojun Desktop Application started\\n");
}

// 应用程序激活
void taishang_application_on_activate(GApplication *app) {
    TaishangApplication *self = TAISHANG_APPLICATION(app);
    
    if (self->main_window == NULL) {
        taishang_application_setup_main_window(self);
    }
    
    taishang_application_show_main_window(self);
    
    // 启动桌面宠物
    if (self->config->enable_desktop_pet) {
        taishang_application_setup_desktop_pet(self);
        taishang_application_toggle_desktop_pet(self);
    }
    
    // 启动定时器
    if (self->update_timer == 0) {
        self->update_timer = g_timeout_add_seconds(1, taishang_application_update_timer_callback, self);
    }
    
    self->state = TAISHANG_APP_STATE_RUNNING;
}

// 应用程序关闭
void taishang_application_on_shutdown(GApplication *app) {
    TaishangApplication *self = TAISHANG_APPLICATION(app);
    
    self->state = TAISHANG_APP_STATE_CLOSING;
    
    // 停止定时器
    if (self->update_timer > 0) {
        g_source_remove(self->update_timer);
        self->update_timer = 0;
    }
    
    if (self->pet_animation_timer > 0) {
        g_source_remove(self->pet_animation_timer);
        self->pet_animation_timer = 0;
    }
    
    // 保存配置
    taishang_app_config_save(self->config);
    
    // 清理资源
    if (self->socket_service) {
        g_socket_service_stop(self->socket_service);
        g_object_unref(self->socket_service);
    }
    
    if (self->status_icon) {
        g_object_unref(self->status_icon);
    }
    
    taishang_app_config_free(self->config);
    
    G_APPLICATION_CLASS(taishang_application_parent_class)->shutdown(app);
    
    g_print("TaishangLaojun Desktop Application shutdown\\n");
}

// 设置主窗口
void taishang_application_setup_main_window(TaishangApplication *app) {
    // 创建主窗口
    app->main_window = GTK_WINDOW(adw_application_window_new(GTK_APPLICATION(app)));
    gtk_window_set_title(app->main_window, _("TaishangLaojun Desktop"));
    gtk_window_set_default_size(app->main_window, 
                               app->config->window_width, 
                               app->config->window_height);
    
    if (app->config->window_maximized) {
        gtk_window_maximize(app->main_window);
    }
    
    // 创建头部栏
    app->header_bar = ADW_HEADER_BAR(adw_header_bar_new());
    adw_header_bar_set_title_widget(app->header_bar, 
                                   gtk_label_new(_("TaishangLaojun Desktop")));
    
    // 创建主布局
    app->main_box = GTK_BOX(gtk_box_new(GTK_ORIENTATION_VERTICAL, 0));
    gtk_box_append(app->main_box, GTK_WIDGET(app->header_bar));
    
    // 创建视图堆栈
    app->view_stack = ADW_VIEW_STACK(adw_view_stack_new());
    app->view_switcher = ADW_VIEW_SWITCHER(adw_view_switcher_new());
    adw_view_switcher_set_stack(app->view_switcher, app->view_stack);
    
    // 添加视图切换器到头部栏
    adw_header_bar_set_title_widget(app->header_bar, GTK_WIDGET(app->view_switcher));
    
    // 设置页面
    taishang_application_setup_project_page(app);
    taishang_application_setup_chat_page(app);
    taishang_application_setup_transfer_page(app);
    taishang_application_setup_settings_page(app);
    
    // 添加视图堆栈到主布局
    gtk_box_append(app->main_box, GTK_WIDGET(app->view_stack));
    
    // 设置窗口内容
    adw_application_window_set_content(ADW_APPLICATION_WINDOW(app->main_window), 
                                      GTK_WIDGET(app->main_box));
    
    // 连接窗口删除事件
    g_signal_connect(app->main_window, "close-request",
                    G_CALLBACK(taishang_application_on_window_delete), app);
}

// 设置项目页面
void taishang_application_setup_project_page(TaishangApplication *app) {
    // 创建项目页面
    GtkWidget *project_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    gtk_widget_set_margin_top(project_box, 24);
    gtk_widget_set_margin_bottom(project_box, 24);
    gtk_widget_set_margin_start(project_box, 24);
    gtk_widget_set_margin_end(project_box, 24);
    
    // 创建工具栏
    GtkWidget *toolbar = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    
    GtkWidget *new_btn = gtk_button_new_with_label(_("New Project"));
    gtk_widget_add_css_class(new_btn, "suggested-action");
    gtk_actionable_set_action_name(GTK_ACTIONABLE(new_btn), "app.new-project");
    
    GtkWidget *open_btn = gtk_button_new_with_label(_("Open Project"));
    gtk_actionable_set_action_name(GTK_ACTIONABLE(open_btn), "app.open-project");
    
    gtk_box_append(GTK_BOX(toolbar), new_btn);
    gtk_box_append(GTK_BOX(toolbar), open_btn);
    
    // 创建项目列表
    GtkWidget *scrolled = gtk_scrolled_window_new();
    gtk_widget_set_vexpand(scrolled, TRUE);
    gtk_widget_set_hexpand(scrolled, TRUE);
    
    GtkWidget *list_box = gtk_list_box_new();
    gtk_widget_add_css_class(list_box, "boxed-list");
    gtk_scrolled_window_set_child(GTK_SCROLLED_WINDOW(scrolled), list_box);
    
    // 添加示例项目
    for (int i = 0; i < 3; i++) {
        GtkWidget *row = adw_action_row_new();
        gchar *title = g_strdup_printf(_("Project %d"), i + 1);
        gchar *subtitle = g_strdup_printf(_("Description for project %d"), i + 1);
        
        adw_preferences_row_set_title(ADW_PREFERENCES_ROW(row), title);
        adw_action_row_set_subtitle(ADW_ACTION_ROW(row), subtitle);
        
        gtk_list_box_append(GTK_LIST_BOX(list_box), row);
        
        g_free(title);
        g_free(subtitle);
    }
    
    gtk_box_append(GTK_BOX(project_box), toolbar);
    gtk_box_append(GTK_BOX(project_box), scrolled);
    
    app->project_page = project_box;
    adw_view_stack_add_titled(app->view_stack, app->project_page, "projects", _("Projects"));
}

// 设置聊天页面
void taishang_application_setup_chat_page(TaishangApplication *app) {
    // 创建聊天页面
    GtkWidget *chat_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
    
    // 创建聊天区域
    GtkWidget *scrolled = gtk_scrolled_window_new();
    gtk_widget_set_vexpand(scrolled, TRUE);
    gtk_widget_set_hexpand(scrolled, TRUE);
    
    GtkWidget *chat_view = gtk_text_view_new();
    gtk_text_view_set_editable(GTK_TEXT_VIEW(chat_view), FALSE);
    gtk_text_view_set_wrap_mode(GTK_TEXT_VIEW(chat_view), GTK_WRAP_WORD);
    gtk_widget_set_margin_top(chat_view, 12);
    gtk_widget_set_margin_bottom(chat_view, 12);
    gtk_widget_set_margin_start(chat_view, 12);
    gtk_widget_set_margin_end(chat_view, 12);
    
    gtk_scrolled_window_set_child(GTK_SCROLLED_WINDOW(scrolled), chat_view);
    
    // 创建输入区域
    GtkWidget *input_box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    gtk_widget_set_margin_top(input_box, 12);
    gtk_widget_set_margin_bottom(input_box, 12);
    gtk_widget_set_margin_start(input_box, 12);
    gtk_widget_set_margin_end(input_box, 12);
    
    GtkWidget *input_entry = gtk_entry_new();
    gtk_widget_set_hexpand(input_entry, TRUE);
    gtk_entry_set_placeholder_text(GTK_ENTRY(input_entry), _("Type your message..."));
    
    GtkWidget *send_btn = gtk_button_new_with_label(_("Send"));
    gtk_widget_add_css_class(send_btn, "suggested-action");
    
    gtk_box_append(GTK_BOX(input_box), input_entry);
    gtk_box_append(GTK_BOX(input_box), send_btn);
    
    gtk_box_append(GTK_BOX(chat_box), scrolled);
    gtk_box_append(GTK_BOX(chat_box), input_box);
    
    app->chat_page = chat_box;
    adw_view_stack_add_titled(app->view_stack, app->chat_page, "chat", _("Chat"));
}

// 设置文件传输页面
void taishang_application_setup_transfer_page(TaishangApplication *app) {
    // 创建传输页面
    GtkWidget *transfer_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    gtk_widget_set_margin_top(transfer_box, 24);
    gtk_widget_set_margin_bottom(transfer_box, 24);
    gtk_widget_set_margin_start(transfer_box, 24);
    gtk_widget_set_margin_end(transfer_box, 24);
    
    // 创建状态卡片
    GtkWidget *status_group = adw_preferences_group_new();
    adw_preferences_group_set_title(ADW_PREFERENCES_GROUP(status_group), _("Transfer Status"));
    
    GtkWidget *status_row = adw_action_row_new();
    adw_preferences_row_set_title(ADW_PREFERENCES_ROW(status_row), _("Connection Status"));
    adw_action_row_set_subtitle(ADW_ACTION_ROW(status_row), _("Connected"));
    
    GtkWidget *speed_row = adw_action_row_new();
    adw_preferences_row_set_title(ADW_PREFERENCES_ROW(status_row), _("Transfer Speed"));
    adw_action_row_set_subtitle(ADW_ACTION_ROW(speed_row), _("0 KB/s"));
    
    adw_preferences_group_add(ADW_PREFERENCES_GROUP(status_group), status_row);
    adw_preferences_group_add(ADW_PREFERENCES_GROUP(status_group), speed_row);
    
    // 创建传输列表
    GtkWidget *transfer_group = adw_preferences_group_new();
    adw_preferences_group_set_title(ADW_PREFERENCES_GROUP(transfer_group), _("Active Transfers"));
    
    gtk_box_append(GTK_BOX(transfer_box), status_group);
    gtk_box_append(GTK_BOX(transfer_box), transfer_group);
    
    app->transfer_page = transfer_box;
    adw_view_stack_add_titled(app->view_stack, app->transfer_page, "transfer", _("Transfer"));
}

// 设置设置页面
void taishang_application_setup_settings_page(TaishangApplication *app) {
    // 创建设置页面
    GtkWidget *settings_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    gtk_widget_set_margin_top(settings_box, 24);
    gtk_widget_set_margin_bottom(settings_box, 24);
    gtk_widget_set_margin_start(settings_box, 24);
    gtk_widget_set_margin_end(settings_box, 24);
    
    // 创建常规设置组
    GtkWidget *general_group = adw_preferences_group_new();
    adw_preferences_group_set_title(ADW_PREFERENCES_GROUP(general_group), _("General"));
    
    // 桌面宠物开关
    GtkWidget *pet_row = adw_switch_row_new();
    adw_preferences_row_set_title(ADW_PREFERENCES_ROW(pet_row), _("Desktop Pet"));
    adw_action_row_set_subtitle(ADW_ACTION_ROW(pet_row), _("Show desktop pet companion"));
    adw_switch_row_set_active(ADW_SWITCH_ROW(pet_row), app->config->enable_desktop_pet);
    
    // 通知开关
    GtkWidget *notify_row = adw_switch_row_new();
    adw_preferences_row_set_title(ADW_PREFERENCES_ROW(notify_row), _("Notifications"));
    adw_action_row_set_subtitle(ADW_ACTION_ROW(notify_row), _("Show system notifications"));
    adw_switch_row_set_active(ADW_SWITCH_ROW(notify_row), app->config->enable_notifications);
    
    // 系统托盘开关
    GtkWidget *tray_row = adw_switch_row_new();
    adw_preferences_row_set_title(ADW_PREFERENCES_ROW(tray_row), _("System Tray"));
    adw_action_row_set_subtitle(ADW_ACTION_ROW(tray_row), _("Show icon in system tray"));
    adw_switch_row_set_active(ADW_SWITCH_ROW(tray_row), app->config->enable_system_tray);
    
    adw_preferences_group_add(ADW_PREFERENCES_GROUP(general_group), pet_row);
    adw_preferences_group_add(ADW_PREFERENCES_GROUP(general_group), notify_row);
    adw_preferences_group_add(ADW_PREFERENCES_GROUP(general_group), tray_row);
    
    gtk_box_append(GTK_BOX(settings_box), general_group);
    
    app->settings_page = settings_box;
    adw_view_stack_add_titled(app->view_stack, app->settings_page, "settings", _("Settings"));
}

// 显示主窗口
void taishang_application_show_main_window(TaishangApplication *app) {
    if (app->main_window) {
        gtk_window_present(app->main_window);
    }
}

// 隐藏主窗口
void taishang_application_hide_main_window(TaishangApplication *app) {
    if (app->main_window) {
        gtk_widget_set_visible(GTK_WIDGET(app->main_window), FALSE);
    }
}

// 窗口删除事件处理
gboolean taishang_application_on_window_delete(GtkWidget *widget, GdkEvent *event, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    if (app->config->enable_system_tray) {
        // 如果启用了系统托盘，隐藏到托盘而不是退出
        taishang_application_hide_main_window(app);
        return TRUE; // 阻止窗口关闭
    }
    
    return FALSE; // 允许窗口关闭
}

// 动作处理函数
void taishang_application_action_new_project(GSimpleAction *action, GVariant *parameter, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    taishang_application_show_info_dialog(app, _("New Project"), _("Create new project functionality will be implemented here."));
}

void taishang_application_action_open_project(GSimpleAction *action, GVariant *parameter, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    taishang_application_show_info_dialog(app, _("Open Project"), _("Open project functionality will be implemented here."));
}

void taishang_application_action_preferences(GSimpleAction *action, GVariant *parameter, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    // 切换到设置页面
    adw_view_stack_set_visible_child_name(app->view_stack, "settings");
    taishang_application_show_main_window(app);
}

void taishang_application_action_about(GSimpleAction *action, GVariant *parameter, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    GtkWidget *about_dialog = adw_about_window_new();
    adw_about_window_set_application_name(ADW_ABOUT_WINDOW(about_dialog), _("TaishangLaojun Desktop"));
    adw_about_window_set_version(ADW_ABOUT_WINDOW(about_dialog), "1.0.0");
    adw_about_window_set_developer_name(ADW_ABOUT_WINDOW(about_dialog), _("TaishangLaojun Team"));
    adw_about_window_set_license_type(ADW_ABOUT_WINDOW(about_dialog), GTK_LICENSE_MIT_X11);
    adw_about_window_set_website(ADW_ABOUT_WINDOW(about_dialog), "https://taishanglaojun.ai");
    adw_about_window_set_issue_url(ADW_ABOUT_WINDOW(about_dialog), "https://github.com/taishanglaojun/desktop-apps/issues");
    
    const gchar *developers[] = { "TaishangLaojun Team", NULL };
    adw_about_window_set_developers(ADW_ABOUT_WINDOW(about_dialog), developers);
    
    gtk_window_set_transient_for(GTK_WINDOW(about_dialog), app->main_window);
    gtk_window_present(GTK_WINDOW(about_dialog));
}

void taishang_application_action_quit(GSimpleAction *action, GVariant *parameter, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    taishang_application_quit(app);
}

void taishang_application_action_toggle_pet(GSimpleAction *action, GVariant *parameter, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    taishang_application_toggle_desktop_pet(app);
}

// 退出应用程序
void taishang_application_quit(TaishangApplication *app) {
    g_application_quit(G_APPLICATION(app));
}

// 定时器回调
gboolean taishang_application_update_timer_callback(gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    // 更新应用程序状态
    // 这里可以添加定期更新的逻辑
    
    return G_SOURCE_CONTINUE;
}

// 显示错误对话框
void taishang_application_show_error_dialog(TaishangApplication *app, const gchar *title, const gchar *message) {
    GtkWidget *dialog = adw_message_dialog_new(app->main_window, title, message);
    adw_message_dialog_add_response(ADW_MESSAGE_DIALOG(dialog), "ok", _("OK"));
    adw_message_dialog_set_default_response(ADW_MESSAGE_DIALOG(dialog), "ok");
    gtk_window_present(GTK_WINDOW(dialog));
}

// 显示信息对话框
void taishang_application_show_info_dialog(TaishangApplication *app, const gchar *title, const gchar *message) {
    GtkWidget *dialog = adw_message_dialog_new(app->main_window, title, message);
    adw_message_dialog_add_response(ADW_MESSAGE_DIALOG(dialog), "ok", _("OK"));
    adw_message_dialog_set_default_response(ADW_MESSAGE_DIALOG(dialog), "ok");
    gtk_window_present(GTK_WINDOW(dialog));
}

// 配置管理实现
TaishangAppConfig *taishang_app_config_new(void) {
    TaishangAppConfig *config = g_malloc0(sizeof(TaishangAppConfig));
    
    // 设置默认值
    config->enable_desktop_pet = TRUE;
    config->enable_notifications = TRUE;
    config->enable_system_tray = TRUE;
    config->auto_start = FALSE;
    config->theme_name = g_strdup("default");
    config->window_width = 1200;
    config->window_height = 800;
    config->window_maximized = FALSE;
    
    return config;
}

void taishang_app_config_free(TaishangAppConfig *config) {
    if (config) {
        g_free(config->theme_name);
        g_free(config);
    }
}

gboolean taishang_app_config_load(TaishangAppConfig *config) {
    gchar *config_file = g_build_filename(taishang_application_get_config_dir(), "config.json", NULL);
    
    if (!g_file_test(config_file, G_FILE_TEST_EXISTS)) {
        g_free(config_file);
        return TRUE; // 使用默认配置
    }
    
    gchar *content;
    gsize length;
    GError *error = NULL;
    
    if (!g_file_get_contents(config_file, &content, &length, &error)) {
        g_warning("Failed to load config file: %s", error->message);
        g_error_free(error);
        g_free(config_file);
        return FALSE;
    }
    
    json_object *root = json_tokener_parse(content);
    if (!root) {
        g_warning("Failed to parse config file");
        g_free(content);
        g_free(config_file);
        return FALSE;
    }
    
    // 解析配置项
    json_object *obj;
    
    if (json_object_object_get_ex(root, "enable_desktop_pet", &obj)) {
        config->enable_desktop_pet = json_object_get_boolean(obj);
    }
    
    if (json_object_object_get_ex(root, "enable_notifications", &obj)) {
        config->enable_notifications = json_object_get_boolean(obj);
    }
    
    if (json_object_object_get_ex(root, "enable_system_tray", &obj)) {
        config->enable_system_tray = json_object_get_boolean(obj);
    }
    
    if (json_object_object_get_ex(root, "auto_start", &obj)) {
        config->auto_start = json_object_get_boolean(obj);
    }
    
    if (json_object_object_get_ex(root, "theme_name", &obj)) {
        g_free(config->theme_name);
        config->theme_name = g_strdup(json_object_get_string(obj));
    }
    
    if (json_object_object_get_ex(root, "window_width", &obj)) {
        config->window_width = json_object_get_int(obj);
    }
    
    if (json_object_object_get_ex(root, "window_height", &obj)) {
        config->window_height = json_object_get_int(obj);
    }
    
    if (json_object_object_get_ex(root, "window_maximized", &obj)) {
        config->window_maximized = json_object_get_boolean(obj);
    }
    
    json_object_put(root);
    g_free(content);
    g_free(config_file);
    
    return TRUE;
}

gboolean taishang_app_config_save(TaishangAppConfig *config) {
    gchar *config_dir = taishang_application_get_config_dir();
    
    // 确保配置目录存在
    if (!g_file_test(config_dir, G_FILE_TEST_IS_DIR)) {
        g_mkdir_with_parents(config_dir, 0755);
    }
    
    gchar *config_file = g_build_filename(config_dir, "config.json", NULL);
    
    // 创建JSON对象
    json_object *root = json_object_new_object();
    
    json_object_object_add(root, "enable_desktop_pet", json_object_new_boolean(config->enable_desktop_pet));
    json_object_object_add(root, "enable_notifications", json_object_new_boolean(config->enable_notifications));
    json_object_object_add(root, "enable_system_tray", json_object_new_boolean(config->enable_system_tray));
    json_object_object_add(root, "auto_start", json_object_new_boolean(config->auto_start));
    json_object_object_add(root, "theme_name", json_object_new_string(config->theme_name));
    json_object_object_add(root, "window_width", json_object_new_int(config->window_width));
    json_object_object_add(root, "window_height", json_object_new_int(config->window_height));
    json_object_object_add(root, "window_maximized", json_object_new_boolean(config->window_maximized));
    
    // 写入文件
    const gchar *json_string = json_object_to_json_string_ext(root, JSON_C_TO_STRING_PRETTY);
    GError *error = NULL;
    
    gboolean success = g_file_set_contents(config_file, json_string, -1, &error);
    
    if (!success) {
        g_warning("Failed to save config file: %s", error->message);
        g_error_free(error);
    }
    
    json_object_put(root);
    g_free(config_file);
    g_free(config_dir);
    
    return success;
}

// 桌面宠物实现
void taishang_application_setup_desktop_pet(TaishangApplication *app) {
    if (app->pet_window) {
        return; // 已经创建
    }
    
    // 创建宠物窗口
    app->pet_window = GTK_WINDOW(gtk_window_new());
    gtk_window_set_title(app->pet_window, "Desktop Pet");
    gtk_window_set_default_size(app->pet_window, 200, 200);
    gtk_window_set_decorated(app->pet_window, FALSE);
    gtk_window_set_resizable(app->pet_window, FALSE);
    
    // 设置窗口属性
    gtk_widget_set_can_focus(GTK_WIDGET(app->pet_window), FALSE);
    
    // 创建绘图区域
    GtkWidget *drawing_area = gtk_drawing_area_new();
    gtk_widget_set_size_request(drawing_area, 200, 200);
    
    // 设置绘制回调
    gtk_drawing_area_set_draw_func(GTK_DRAWING_AREA(drawing_area),
                                  (GtkDrawingAreaDrawFunc)taishang_application_pet_draw_callback,
                                  app, NULL);
    
    gtk_window_set_child(app->pet_window, drawing_area);
    
    // 启动动画定时器
    if (app->pet_animation_timer == 0) {
        app->pet_animation_timer = g_timeout_add(100, taishang_application_pet_animation_callback, app);
    }
}

void taishang_application_toggle_desktop_pet(TaishangApplication *app) {
    if (!app->pet_window) {
        taishang_application_setup_desktop_pet(app);
    }
    
    if (app->pet_visible) {
        gtk_widget_set_visible(GTK_WIDGET(app->pet_window), FALSE);
        app->pet_visible = FALSE;
    } else {
        gtk_widget_set_visible(GTK_WIDGET(app->pet_window), TRUE);
        gtk_window_present(app->pet_window);
        app->pet_visible = TRUE;
    }
}

// 宠物绘制回调
void taishang_application_pet_draw_callback(GtkDrawingArea *area, cairo_t *cr, int width, int height, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    // 清除背景
    cairo_set_source_rgba(cr, 0, 0, 0, 0);
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);
    
    // 绘制宠物身体
    cairo_set_source_rgb(cr, 1.0, 0.8, 0.2); // 金色
    cairo_arc(cr, width/2, height/2, 60, 0, 2 * G_PI);
    cairo_fill(cr);
    
    // 绘制眼睛
    cairo_set_source_rgb(cr, 0, 0, 0);
    cairo_arc(cr, width/2 - 20, height/2 - 15, 8, 0, 2 * G_PI);
    cairo_fill(cr);
    cairo_arc(cr, width/2 + 20, height/2 - 15, 8, 0, 2 * G_PI);
    cairo_fill(cr);
    
    // 绘制嘴巴
    cairo_arc(cr, width/2, height/2 + 10, 15, 0, G_PI);
    cairo_stroke(cr);
    
    // 绘制光环效果
    static double angle = 0;
    angle += 0.1;
    
    cairo_set_source_rgba(cr, 1.0, 1.0, 0.0, 0.5);
    cairo_set_line_width(cr, 3);
    
    for (int i = 0; i < 8; i++) {
        double a = angle + i * G_PI / 4;
        double x = width/2 + 80 * cos(a);
        double y = height/2 + 80 * sin(a);
        cairo_arc(cr, x, y, 5, 0, 2 * G_PI);
        cairo_fill(cr);
    }
}

gboolean taishang_application_pet_animation_callback(gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    if (app->pet_window && app->pet_visible) {
        gtk_widget_queue_draw(GTK_WIDGET(app->pet_window));
    }
    
    return G_SOURCE_CONTINUE;
}

// 系统托盘实现
void taishang_application_setup_status_icon(TaishangApplication *app) {
    // 注意：GTK4中GtkStatusIcon已被弃用，这里使用传统方式作为示例
    // 在实际应用中应该使用libappindicator或其他现代方案
    
    app->status_icon = gtk_status_icon_new_from_icon_name("applications-system");
    gtk_status_icon_set_tooltip_text(app->status_icon, _("TaishangLaojun Desktop"));
    gtk_status_icon_set_visible(app->status_icon, TRUE);
    
    // 连接信号
    g_signal_connect(app->status_icon, "activate",
                    G_CALLBACK(taishang_application_status_icon_activate), app);
    g_signal_connect(app->status_icon, "popup-menu",
                    G_CALLBACK(taishang_application_status_icon_popup_menu), app);
    
    // 创建右键菜单
    app->status_menu = GTK_MENU(gtk_menu_new());
    
    GtkWidget *show_item = gtk_menu_item_new_with_label(_("Show Window"));
    GtkWidget *pet_item = gtk_menu_item_new_with_label(_("Toggle Pet"));
    GtkWidget *sep_item = gtk_separator_menu_item_new();
    GtkWidget *quit_item = gtk_menu_item_new_with_label(_("Quit"));
    
    gtk_menu_shell_append(GTK_MENU_SHELL(app->status_menu), show_item);
    gtk_menu_shell_append(GTK_MENU_SHELL(app->status_menu), pet_item);
    gtk_menu_shell_append(GTK_MENU_SHELL(app->status_menu), sep_item);
    gtk_menu_shell_append(GTK_MENU_SHELL(app->status_menu), quit_item);
    
    // 连接菜单项信号
    g_signal_connect_swapped(show_item, "activate",
                           G_CALLBACK(taishang_application_show_main_window), app);
    g_signal_connect_swapped(pet_item, "activate",
                           G_CALLBACK(taishang_application_toggle_desktop_pet), app);
    g_signal_connect_swapped(quit_item, "activate",
                           G_CALLBACK(taishang_application_quit), app);
    
    gtk_widget_show_all(GTK_WIDGET(app->status_menu));
}

void taishang_application_status_icon_activate(GtkStatusIcon *status_icon, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    if (gtk_widget_get_visible(GTK_WIDGET(app->main_window))) {
        taishang_application_hide_main_window(app);
    } else {
        taishang_application_show_main_window(app);
    }
}

void taishang_application_status_icon_popup_menu(GtkStatusIcon *status_icon, guint button, guint activate_time, gpointer user_data) {
    TaishangApplication *app = TAISHANG_APPLICATION(user_data);
    
    gtk_menu_popup_at_pointer(app->status_menu, NULL);
}

// 工具函数实现
gchar *taishang_application_get_config_dir(void) {
    return g_build_filename(g_get_user_config_dir(), "taishanglaojun-desktop", NULL);
}

gchar *taishang_application_get_data_dir(void) {
    return g_build_filename(g_get_user_data_dir(), "taishanglaojun-desktop", NULL);
}