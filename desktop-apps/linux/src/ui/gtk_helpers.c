#include <gtk/gtk.h>
#include <adwaita.h>
#include <glib.h>

// CSS styling
static const char *app_css = 
"window.taishang-main-window {\n"
"  background-color: @window_bg_color;\n"
"}\n"
"\n"
".taishang-chat-bubble {\n"
"  background-color: @card_bg_color;\n"
"  border-radius: 12px;\n"
"  padding: 12px;\n"
"  margin: 6px;\n"
"  box-shadow: 0 1px 3px rgba(0,0,0,0.1);\n"
"}\n"
"\n"
".taishang-chat-bubble.user {\n"
"  background-color: @accent_bg_color;\n"
"  color: @accent_fg_color;\n"
"}\n"
"\n"
".taishang-chat-bubble.assistant {\n"
"  background-color: @card_bg_color;\n"
"  color: @card_fg_color;\n"
"}\n"
"\n"
".taishang-project-card {\n"
"  background-color: @card_bg_color;\n"
"  border-radius: 8px;\n"
"  padding: 16px;\n"
"  margin: 8px;\n"
"  border: 1px solid @borders;\n"
"  transition: all 200ms ease;\n"
"}\n"
"\n"
".taishang-project-card:hover {\n"
"  background-color: @view_hover_bg_color;\n"
"  transform: translateY(-2px);\n"
"  box-shadow: 0 4px 12px rgba(0,0,0,0.15);\n"
"}\n"
"\n"
".taishang-status-bar {\n"
"  background-color: @headerbar_bg_color;\n"
"  border-top: 1px solid @borders;\n"
"  padding: 6px 12px;\n"
"}\n"
"\n"
".taishang-pet-window {\n"
"  background-color: transparent;\n"
"}\n"
"\n"
".taishang-notification {\n"
"  background-color: @accent_bg_color;\n"
"  color: @accent_fg_color;\n"
"  border-radius: 6px;\n"
"  padding: 8px 12px;\n"
"  margin: 4px;\n"
"}\n"
"\n"
".taishang-notification.error {\n"
"  background-color: @error_bg_color;\n"
"  color: @error_fg_color;\n"
"}\n"
"\n"
".taishang-notification.warning {\n"
"  background-color: @warning_bg_color;\n"
"  color: @warning_fg_color;\n"
"}\n"
"\n"
".taishang-notification.success {\n"
"  background-color: @success_bg_color;\n"
"  color: @success_fg_color;\n"
"}\n";

// Forward declarations
static void load_css_from_resource(void);
static void apply_custom_css(void);

// Public functions
void taishang_gtk_helpers_init(void);
void taishang_gtk_helpers_cleanup(void);

// Widget creation helpers
GtkWidget *taishang_gtk_create_header_bar(const char *title);
GtkWidget *taishang_gtk_create_button_with_icon(const char *icon_name, const char *label);
GtkWidget *taishang_gtk_create_menu_button(GMenuModel *menu_model);
GtkWidget *taishang_gtk_create_search_entry(const char *placeholder);
GtkWidget *taishang_gtk_create_info_bar(const char *message, GtkMessageType type);

// Layout helpers
GtkWidget *taishang_gtk_create_scrolled_window(GtkWidget *child);
GtkWidget *taishang_gtk_create_paned_window(GtkWidget *child1, GtkWidget *child2, GtkOrientation orientation);
void taishang_gtk_set_margins(GtkWidget *widget, int margin);
void taishang_gtk_set_spacing(GtkWidget *box, int spacing);

// Styling helpers
void taishang_gtk_add_css_class(GtkWidget *widget, const char *css_class);
void taishang_gtk_remove_css_class(GtkWidget *widget, const char *css_class);
void taishang_gtk_set_widget_style(GtkWidget *widget, const char *css);

// Dialog helpers
GtkWidget *taishang_gtk_create_message_dialog(GtkWindow *parent, const char *title, const char *message, GtkMessageType type);
GtkWidget *taishang_gtk_create_file_chooser_dialog(GtkWindow *parent, const char *title, GtkFileChooserAction action);
gboolean taishang_gtk_show_confirmation_dialog(GtkWindow *parent, const char *title, const char *message);

// Animation helpers
void taishang_gtk_fade_in_widget(GtkWidget *widget, guint duration_ms);
void taishang_gtk_fade_out_widget(GtkWidget *widget, guint duration_ms);
void taishang_gtk_slide_in_widget(GtkWidget *widget, GtkOrientation direction, guint duration_ms);

// Utility functions
void taishang_gtk_show_toast(GtkWidget *parent, const char *message);
void taishang_gtk_copy_to_clipboard(const char *text);
char *taishang_gtk_get_clipboard_text(void);

// Implementation
void taishang_gtk_helpers_init(void) {
    apply_custom_css();
    g_print("GTK helpers initialized\n");
}

void taishang_gtk_helpers_cleanup(void) {
    g_print("GTK helpers cleaned up\n");
}

static void apply_custom_css(void) {
    GtkCssProvider *provider = gtk_css_provider_new();
    gtk_css_provider_load_from_data(provider, app_css, -1);
    
    gtk_style_context_add_provider_for_display(
        gdk_display_get_default(),
        GTK_STYLE_PROVIDER(provider),
        GTK_STYLE_PROVIDER_PRIORITY_APPLICATION
    );
    
    g_object_unref(provider);
}

// Widget creation helpers implementation
GtkWidget *taishang_gtk_create_header_bar(const char *title) {
    GtkWidget *header_bar = adw_header_bar_new();
    
    if (title) {
        GtkWidget *title_label = gtk_label_new(title);
        gtk_widget_add_css_class(title_label, "title");
        adw_header_bar_set_title_widget(ADW_HEADER_BAR(header_bar), title_label);
    }
    
    return header_bar;
}

GtkWidget *taishang_gtk_create_button_with_icon(const char *icon_name, const char *label) {
    GtkWidget *button;
    
    if (label) {
        button = gtk_button_new();
        GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
        
        if (icon_name) {
            GtkWidget *icon = gtk_image_new_from_icon_name(icon_name);
            gtk_box_append(GTK_BOX(box), icon);
        }
        
        GtkWidget *label_widget = gtk_label_new(label);
        gtk_box_append(GTK_BOX(box), label_widget);
        
        gtk_button_set_child(GTK_BUTTON(button), box);
    } else if (icon_name) {
        button = gtk_button_new_from_icon_name(icon_name);
    } else {
        button = gtk_button_new();
    }
    
    return button;
}

GtkWidget *taishang_gtk_create_menu_button(GMenuModel *menu_model) {
    GtkWidget *menu_button = gtk_menu_button_new();
    
    if (menu_model) {
        gtk_menu_button_set_menu_model(GTK_MENU_BUTTON(menu_button), menu_model);
    }
    
    gtk_menu_button_set_icon_name(GTK_MENU_BUTTON(menu_button), "open-menu-symbolic");
    
    return menu_button;
}

GtkWidget *taishang_gtk_create_search_entry(const char *placeholder) {
    GtkWidget *search_entry = gtk_search_entry_new();
    
    if (placeholder) {
        gtk_entry_set_placeholder_text(GTK_ENTRY(search_entry), placeholder);
    }
    
    return search_entry;
}

GtkWidget *taishang_gtk_create_info_bar(const char *message, GtkMessageType type) {
    GtkWidget *info_bar = gtk_info_bar_new();
    gtk_info_bar_set_message_type(GTK_INFO_BAR(info_bar), type);
    
    if (message) {
        GtkWidget *label = gtk_label_new(message);
        gtk_info_bar_add_child(GTK_INFO_BAR(info_bar), label);
    }
    
    return info_bar;
}

// Layout helpers implementation
GtkWidget *taishang_gtk_create_scrolled_window(GtkWidget *child) {
    GtkWidget *scrolled = gtk_scrolled_window_new();
    gtk_scrolled_window_set_policy(GTK_SCROLLED_WINDOW(scrolled),
                                   GTK_POLICY_AUTOMATIC,
                                   GTK_POLICY_AUTOMATIC);
    
    if (child) {
        gtk_scrolled_window_set_child(GTK_SCROLLED_WINDOW(scrolled), child);
    }
    
    return scrolled;
}

GtkWidget *taishang_gtk_create_paned_window(GtkWidget *child1, GtkWidget *child2, GtkOrientation orientation) {
    GtkWidget *paned = gtk_paned_new(orientation);
    
    if (child1) {
        gtk_paned_set_start_child(GTK_PANED(paned), child1);
    }
    
    if (child2) {
        gtk_paned_set_end_child(GTK_PANED(paned), child2);
    }
    
    return paned;
}

void taishang_gtk_set_margins(GtkWidget *widget, int margin) {
    gtk_widget_set_margin_top(widget, margin);
    gtk_widget_set_margin_bottom(widget, margin);
    gtk_widget_set_margin_start(widget, margin);
    gtk_widget_set_margin_end(widget, margin);
}

void taishang_gtk_set_spacing(GtkWidget *box, int spacing) {
    if (GTK_IS_BOX(box)) {
        gtk_box_set_spacing(GTK_BOX(box), spacing);
    }
}

// Styling helpers implementation
void taishang_gtk_add_css_class(GtkWidget *widget, const char *css_class) {
    if (widget && css_class) {
        gtk_widget_add_css_class(widget, css_class);
    }
}

void taishang_gtk_remove_css_class(GtkWidget *widget, const char *css_class) {
    if (widget && css_class) {
        gtk_widget_remove_css_class(widget, css_class);
    }
}

void taishang_gtk_set_widget_style(GtkWidget *widget, const char *css) {
    if (!widget || !css) {
        return;
    }
    
    GtkCssProvider *provider = gtk_css_provider_new();
    gtk_css_provider_load_from_data(provider, css, -1);
    
    GtkStyleContext *context = gtk_widget_get_style_context(widget);
    gtk_style_context_add_provider(context, GTK_STYLE_PROVIDER(provider),
                                   GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
    
    g_object_unref(provider);
}

// Dialog helpers implementation
GtkWidget *taishang_gtk_create_message_dialog(GtkWindow *parent, const char *title, const char *message, GtkMessageType type) {
    GtkWidget *dialog = gtk_message_dialog_new(parent,
                                               GTK_DIALOG_MODAL | GTK_DIALOG_DESTROY_WITH_PARENT,
                                               type,
                                               GTK_BUTTONS_OK,
                                               "%s", message ? message : "");
    
    if (title) {
        gtk_window_set_title(GTK_WINDOW(dialog), title);
    }
    
    return dialog;
}

GtkWidget *taishang_gtk_create_file_chooser_dialog(GtkWindow *parent, const char *title, GtkFileChooserAction action) {
    const char *button_text = (action == GTK_FILE_CHOOSER_ACTION_SAVE) ? "保存" : "打开";
    
    GtkWidget *dialog = gtk_file_chooser_dialog_new(title,
                                                    parent,
                                                    action,
                                                    "取消", GTK_RESPONSE_CANCEL,
                                                    button_text, GTK_RESPONSE_ACCEPT,
                                                    NULL);
    
    return dialog;
}

gboolean taishang_gtk_show_confirmation_dialog(GtkWindow *parent, const char *title, const char *message) {
    GtkWidget *dialog = gtk_message_dialog_new(parent,
                                               GTK_DIALOG_MODAL | GTK_DIALOG_DESTROY_WITH_PARENT,
                                               GTK_MESSAGE_QUESTION,
                                               GTK_BUTTONS_YES_NO,
                                               "%s", message ? message : "");
    
    if (title) {
        gtk_window_set_title(GTK_WINDOW(dialog), title);
    }
    
    int response = gtk_dialog_run(GTK_DIALOG(dialog));
    gtk_window_destroy(GTK_WINDOW(dialog));
    
    return (response == GTK_RESPONSE_YES);
}

// Animation helpers implementation
void taishang_gtk_fade_in_widget(GtkWidget *widget, guint duration_ms) {
    if (!widget) return;
    
    gtk_widget_set_opacity(widget, 0.0);
    gtk_widget_set_visible(widget, TRUE);
    
    // TODO: Implement proper animation using GtkAnimation or similar
    // For now, just set full opacity
    gtk_widget_set_opacity(widget, 1.0);
}

void taishang_gtk_fade_out_widget(GtkWidget *widget, guint duration_ms) {
    if (!widget) return;
    
    // TODO: Implement proper animation
    // For now, just hide the widget
    gtk_widget_set_visible(widget, FALSE);
}

void taishang_gtk_slide_in_widget(GtkWidget *widget, GtkOrientation direction, guint duration_ms) {
    if (!widget) return;
    
    // TODO: Implement slide animation
    gtk_widget_set_visible(widget, TRUE);
}

// Utility functions implementation
void taishang_gtk_show_toast(GtkWidget *parent, const char *message) {
    if (!message) return;
    
    // TODO: Implement toast notification using AdwToast
    g_print("Toast: %s\n", message);
}

void taishang_gtk_copy_to_clipboard(const char *text) {
    if (!text) return;
    
    GdkClipboard *clipboard = gdk_display_get_clipboard(gdk_display_get_default());
    gdk_clipboard_set_text(clipboard, text);
}

char *taishang_gtk_get_clipboard_text(void) {
    GdkClipboard *clipboard = gdk_display_get_clipboard(gdk_display_get_default());
    
    // TODO: Implement async clipboard reading
    // For now, return NULL
    return NULL;
}