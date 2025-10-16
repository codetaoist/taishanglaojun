#ifndef TAISHANG_WINDOW_MANAGER_H
#define TAISHANG_WINDOW_MANAGER_H

#include <gtk/gtk.h>
#include "application.h"

G_BEGIN_DECLS

// Window manager type
typedef struct _TaishangWindowManager TaishangWindowManager;

// Initialization and cleanup
gboolean taishang_window_manager_init(TaishangApplication *app);
void taishang_window_manager_cleanup(void);
TaishangWindowManager *taishang_window_manager_get_instance(void);

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

G_END_DECLS

#endif // TAISHANG_WINDOW_MANAGER_H