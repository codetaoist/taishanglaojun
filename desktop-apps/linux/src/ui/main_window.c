#include <gtk/gtk.h>
#include <adwaita.h>
#include "../../include/application.h"

// Main window structure
typedef struct {
    GtkApplicationWindow parent;
    
    // Header bar
    AdwHeaderBar *header_bar;
    GtkButton *menu_button;
    GtkButton *settings_button;
    
    // Navigation
    AdwViewStack *view_stack;
    AdwViewSwitcher *view_switcher;
    
    // Pages
    GtkWidget *chat_page;
    GtkWidget *project_page;
    GtkWidget *file_transfer_page;
    GtkWidget *settings_page;
    
    // Status
    GtkLabel *status_label;
    GtkProgressBar *progress_bar;
    
    // Application reference
    TaishangApplication *app;
} TaishangMainWindow;

G_DEFINE_TYPE(TaishangMainWindow, taishang_main_window, GTK_TYPE_APPLICATION_WINDOW)

// Forward declarations
static void taishang_main_window_init(TaishangMainWindow *self);
static void taishang_main_window_class_init(TaishangMainWindowClass *klass);
static void taishang_main_window_setup_header_bar(TaishangMainWindow *self);
static void taishang_main_window_setup_navigation(TaishangMainWindow *self);
static void taishang_main_window_setup_pages(TaishangMainWindow *self);
static void taishang_main_window_setup_status_bar(TaishangMainWindow *self);

// Callback functions
static void on_menu_button_clicked(GtkButton *button, TaishangMainWindow *self);
static void on_settings_button_clicked(GtkButton *button, TaishangMainWindow *self);
static void on_view_stack_notify_visible_child(GObject *object, GParamSpec *pspec, TaishangMainWindow *self);
static gboolean on_window_close_request(GtkWindow *window, TaishangMainWindow *self);

// Public functions
GtkWidget *taishang_main_window_new(TaishangApplication *app);
void taishang_main_window_show_page(TaishangMainWindow *self, const char *page_name);
void taishang_main_window_set_status(TaishangMainWindow *self, const char *status);
void taishang_main_window_set_progress(TaishangMainWindow *self, double progress);
void taishang_main_window_add_notification(TaishangMainWindow *self, const char *message, const char *type);

// Class initialization
static void taishang_main_window_class_init(TaishangMainWindowClass *klass) {
    GtkWidgetClass *widget_class = GTK_WIDGET_CLASS(klass);
    
    // Set up CSS
    gtk_widget_class_set_css_name(widget_class, "taishang-main-window");
}

// Instance initialization
static void taishang_main_window_init(TaishangMainWindow *self) {
    // Set window properties
    gtk_window_set_title(GTK_WINDOW(self), "太上老君 - AI助手");
    gtk_window_set_default_size(GTK_WINDOW(self), 1200, 800);
    gtk_window_set_icon_name(GTK_WINDOW(self), "taishang-app");
    
    // Connect close signal
    g_signal_connect(self, "close-request", G_CALLBACK(on_window_close_request), self);
    
    // Setup UI components
    taishang_main_window_setup_header_bar(self);
    taishang_main_window_setup_navigation(self);
    taishang_main_window_setup_pages(self);
    taishang_main_window_setup_status_bar(self);
    
    // Create main layout
    GtkWidget *main_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
    gtk_window_set_child(GTK_WINDOW(self), main_box);
    
    // Add components to layout
    gtk_box_append(GTK_BOX(main_box), GTK_WIDGET(self->header_bar));
    gtk_box_append(GTK_BOX(main_box), GTK_WIDGET(self->view_switcher));
    gtk_box_append(GTK_BOX(main_box), GTK_WIDGET(self->view_stack));
    gtk_box_append(GTK_BOX(main_box), GTK_WIDGET(self->status_label));
    gtk_box_append(GTK_BOX(main_box), GTK_WIDGET(self->progress_bar));
    
    // Set expand properties
    gtk_widget_set_vexpand(GTK_WIDGET(self->view_stack), TRUE);
    gtk_widget_set_hexpand(GTK_WIDGET(self->view_stack), TRUE);
}

// Setup header bar
static void taishang_main_window_setup_header_bar(TaishangMainWindow *self) {
    self->header_bar = ADW_HEADER_BAR(adw_header_bar_new());
    adw_header_bar_set_title_widget(self->header_bar, gtk_label_new("太上老君"));
    
    // Menu button
    self->menu_button = GTK_BUTTON(gtk_button_new_from_icon_name("open-menu-symbolic"));
    gtk_widget_set_tooltip_text(GTK_WIDGET(self->menu_button), "主菜单");
    g_signal_connect(self->menu_button, "clicked", G_CALLBACK(on_menu_button_clicked), self);
    adw_header_bar_pack_start(self->header_bar, GTK_WIDGET(self->menu_button));
    
    // Settings button
    self->settings_button = GTK_BUTTON(gtk_button_new_from_icon_name("preferences-system-symbolic"));
    gtk_widget_set_tooltip_text(GTK_WIDGET(self->settings_button), "设置");
    g_signal_connect(self->settings_button, "clicked", G_CALLBACK(on_settings_button_clicked), self);
    adw_header_bar_pack_end(self->header_bar, GTK_WIDGET(self->settings_button));
}

// Setup navigation
static void taishang_main_window_setup_navigation(TaishangMainWindow *self) {
    self->view_stack = ADW_VIEW_STACK(adw_view_stack_new());
    self->view_switcher = ADW_VIEW_SWITCHER(adw_view_switcher_new());
    
    adw_view_switcher_set_stack(self->view_switcher, self->view_stack);
    g_signal_connect(self->view_stack, "notify::visible-child", 
                     G_CALLBACK(on_view_stack_notify_visible_child), self);
}

// Setup pages
static void taishang_main_window_setup_pages(TaishangMainWindow *self) {
    // Chat page
    self->chat_page = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    gtk_widget_set_margin_top(self->chat_page, 12);
    gtk_widget_set_margin_bottom(self->chat_page, 12);
    gtk_widget_set_margin_start(self->chat_page, 12);
    gtk_widget_set_margin_end(self->chat_page, 12);
    
    GtkWidget *chat_label = gtk_label_new("AI聊天助手");
    gtk_widget_add_css_class(chat_label, "title-1");
    gtk_box_append(GTK_BOX(self->chat_page), chat_label);
    
    // Chat input area
    GtkWidget *chat_input_box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    GtkWidget *chat_entry = gtk_entry_new();
    gtk_entry_set_placeholder_text(GTK_ENTRY(chat_entry), "输入您的问题...");
    gtk_widget_set_hexpand(chat_entry, TRUE);
    
    GtkWidget *send_button = gtk_button_new_with_label("发送");
    gtk_widget_add_css_class(send_button, "suggested-action");
    
    gtk_box_append(GTK_BOX(chat_input_box), chat_entry);
    gtk_box_append(GTK_BOX(chat_input_box), send_button);
    gtk_box_append(GTK_BOX(self->chat_page), chat_input_box);
    
    adw_view_stack_add_titled(self->view_stack, self->chat_page, "chat", "聊天");
    
    // Project page
    self->project_page = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    gtk_widget_set_margin_top(self->project_page, 12);
    gtk_widget_set_margin_bottom(self->project_page, 12);
    gtk_widget_set_margin_start(self->project_page, 12);
    gtk_widget_set_margin_end(self->project_page, 12);
    
    GtkWidget *project_label = gtk_label_new("项目管理");
    gtk_widget_add_css_class(project_label, "title-1");
    gtk_box_append(GTK_BOX(self->project_page), project_label);
    
    // Project toolbar
    GtkWidget *project_toolbar = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    GtkWidget *new_project_button = gtk_button_new_with_label("新建项目");
    GtkWidget *open_project_button = gtk_button_new_with_label("打开项目");
    gtk_widget_add_css_class(new_project_button, "suggested-action");
    
    gtk_box_append(GTK_BOX(project_toolbar), new_project_button);
    gtk_box_append(GTK_BOX(project_toolbar), open_project_button);
    gtk_box_append(GTK_BOX(self->project_page), project_toolbar);
    
    adw_view_stack_add_titled(self->view_stack, self->project_page, "project", "项目");
    
    // File transfer page
    self->file_transfer_page = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    gtk_widget_set_margin_top(self->file_transfer_page, 12);
    gtk_widget_set_margin_bottom(self->file_transfer_page, 12);
    gtk_widget_set_margin_start(self->file_transfer_page, 12);
    gtk_widget_set_margin_end(self->file_transfer_page, 12);
    
    GtkWidget *transfer_label = gtk_label_new("文件传输");
    gtk_widget_add_css_class(transfer_label, "title-1");
    gtk_box_append(GTK_BOX(self->file_transfer_page), transfer_label);
    
    // File transfer controls
    GtkWidget *transfer_toolbar = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    GtkWidget *upload_button = gtk_button_new_with_label("上传文件");
    GtkWidget *download_button = gtk_button_new_with_label("下载文件");
    gtk_widget_add_css_class(upload_button, "suggested-action");
    
    gtk_box_append(GTK_BOX(transfer_toolbar), upload_button);
    gtk_box_append(GTK_BOX(transfer_toolbar), download_button);
    gtk_box_append(GTK_BOX(self->file_transfer_page), transfer_toolbar);
    
    adw_view_stack_add_titled(self->view_stack, self->file_transfer_page, "transfer", "传输");
}

// Setup status bar
static void taishang_main_window_setup_status_bar(TaishangMainWindow *self) {
    self->status_label = GTK_LABEL(gtk_label_new("就绪"));
    gtk_widget_set_halign(GTK_WIDGET(self->status_label), GTK_ALIGN_START);
    gtk_widget_set_margin_start(GTK_WIDGET(self->status_label), 12);
    gtk_widget_set_margin_end(GTK_WIDGET(self->status_label), 12);
    
    self->progress_bar = GTK_PROGRESS_BAR(gtk_progress_bar_new());
    gtk_widget_set_visible(GTK_WIDGET(self->progress_bar), FALSE);
    gtk_widget_set_margin_start(GTK_WIDGET(self->progress_bar), 12);
    gtk_widget_set_margin_end(GTK_WIDGET(self->progress_bar), 12);
    gtk_widget_set_margin_bottom(GTK_WIDGET(self->progress_bar), 6);
}

// Callback implementations
static void on_menu_button_clicked(GtkButton *button, TaishangMainWindow *self) {
    g_print("Menu button clicked\n");
    // TODO: Show application menu
}

static void on_settings_button_clicked(GtkButton *button, TaishangMainWindow *self) {
    g_print("Settings button clicked\n");
    // TODO: Show settings dialog
}

static void on_view_stack_notify_visible_child(GObject *object, GParamSpec *pspec, TaishangMainWindow *self) {
    GtkWidget *visible_child = adw_view_stack_get_visible_child(self->view_stack);
    const char *child_name = adw_view_stack_get_visible_child_name(self->view_stack);
    
    g_print("Switched to page: %s\n", child_name ? child_name : "unknown");
    
    // Update status based on current page
    if (g_strcmp0(child_name, "chat") == 0) {
        taishang_main_window_set_status(self, "AI聊天助手已就绪");
    } else if (g_strcmp0(child_name, "project") == 0) {
        taishang_main_window_set_status(self, "项目管理");
    } else if (g_strcmp0(child_name, "transfer") == 0) {
        taishang_main_window_set_status(self, "文件传输");
    }
}

static gboolean on_window_close_request(GtkWindow *window, TaishangMainWindow *self) {
    // Hide to system tray instead of closing
    gtk_widget_set_visible(GTK_WIDGET(self), FALSE);
    return TRUE; // Prevent default close behavior
}

// Public function implementations
GtkWidget *taishang_main_window_new(TaishangApplication *app) {
    TaishangMainWindow *window = g_object_new(TAISHANG_TYPE_MAIN_WINDOW,
                                              "application", app,
                                              NULL);
    window->app = app;
    return GTK_WIDGET(window);
}

void taishang_main_window_show_page(TaishangMainWindow *self, const char *page_name) {
    g_return_if_fail(TAISHANG_IS_MAIN_WINDOW(self));
    g_return_if_fail(page_name != NULL);
    
    adw_view_stack_set_visible_child_name(self->view_stack, page_name);
}

void taishang_main_window_set_status(TaishangMainWindow *self, const char *status) {
    g_return_if_fail(TAISHANG_IS_MAIN_WINDOW(self));
    g_return_if_fail(status != NULL);
    
    gtk_label_set_text(self->status_label, status);
}

void taishang_main_window_set_progress(TaishangMainWindow *self, double progress) {
    g_return_if_fail(TAISHANG_IS_MAIN_WINDOW(self));
    g_return_if_fail(progress >= 0.0 && progress <= 1.0);
    
    if (progress > 0.0 && progress < 1.0) {
        gtk_widget_set_visible(GTK_WIDGET(self->progress_bar), TRUE);
        gtk_progress_bar_set_fraction(self->progress_bar, progress);
    } else {
        gtk_widget_set_visible(GTK_WIDGET(self->progress_bar), FALSE);
    }
}

void taishang_main_window_add_notification(TaishangMainWindow *self, const char *message, const char *type) {
    g_return_if_fail(TAISHANG_IS_MAIN_WINDOW(self));
    g_return_if_fail(message != NULL);
    
    // TODO: Implement toast notifications using AdwToast
    g_print("Notification [%s]: %s\n", type ? type : "info", message);
}