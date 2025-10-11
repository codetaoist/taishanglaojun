#ifndef TAISHANG_GTK_HELPERS_H
#define TAISHANG_GTK_HELPERS_H

#include <gtk/gtk.h>
#include <adwaita.h>

G_BEGIN_DECLS

// Initialization and cleanup
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

G_END_DECLS

#endif // TAISHANG_GTK_HELPERS_H