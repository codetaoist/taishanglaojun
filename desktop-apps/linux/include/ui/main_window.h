#ifndef TAISHANG_MAIN_WINDOW_H
#define TAISHANG_MAIN_WINDOW_H

#include <gtk/gtk.h>
#include <adwaita.h>
#include "../application.h"

G_BEGIN_DECLS

#define TAISHANG_TYPE_MAIN_WINDOW (taishang_main_window_get_type())
G_DECLARE_FINAL_TYPE(TaishangMainWindow, taishang_main_window, TAISHANG, MAIN_WINDOW, GtkApplicationWindow)

// Constructor
GtkWidget *taishang_main_window_new(TaishangApplication *app);

// Page management
void taishang_main_window_show_page(TaishangMainWindow *self, const char *page_name);

// Status management
void taishang_main_window_set_status(TaishangMainWindow *self, const char *status);
void taishang_main_window_set_progress(TaishangMainWindow *self, double progress);

// Notifications
void taishang_main_window_add_notification(TaishangMainWindow *self, const char *message, const char *type);

G_END_DECLS

#endif // TAISHANG_MAIN_WINDOW_H