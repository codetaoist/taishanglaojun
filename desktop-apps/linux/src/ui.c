/**
 * @file ui.c
 * @brief TaishangLaojun Desktop Application UI Implementation
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains the user interface implementation for the
 * TaishangLaojun desktop application on Linux using GTK.
 */

#include "common.h"
#include "ui.h"
#include "config.h"
#include "utils.h"

/* Private structure */
struct _TaishangUIPrivate {
    /* Main window and layout */
    GtkWidget *main_window;
    GtkWidget *header_bar;
    GtkWidget *main_box;
    GtkWidget *content_paned;
    
    /* Menu and toolbar */
    GtkWidget *menu_bar;
    GtkWidget *toolbar;
    GtkWidget *hamburger_menu;
    
    /* Main content areas */
    GtkWidget *sidebar;
    GtkWidget *chat_area;
    GtkWidget *input_area;
    
    /* Chat components */
    GtkWidget *chat_scrolled;
    GtkWidget *chat_listbox;
    GtkWidget *message_entry;
    GtkWidget *send_button;
    
    /* Sidebar components */
    GtkWidget *sidebar_stack;
    GtkWidget *conversation_list;
    GtkWidget *settings_panel;
    
    /* Status and progress */
    GtkWidget *status_bar;
    GtkWidget *progress_bar;
    GtkWidget *connection_indicator;
    
    /* Dialogs */
    GtkWidget *preferences_dialog;
    GtkWidget *about_dialog;
    GtkWidget *file_chooser;
    
    /* Configuration and state */
    TaishangConfig *config;
    TaishangUITheme theme;
    TaishangUIState state;
    
    /* Window state */
    gint window_width;
    gint window_height;
    gint window_x;
    gint window_y;
    gboolean window_maximized;
    gboolean window_fullscreen;
    
    /* UI settings */
    gboolean sidebar_visible;
    gboolean toolbar_visible;
    gboolean status_bar_visible;
    gint sidebar_width;
    
    /* CSS provider */
    GtkCssProvider *css_provider;
    
    /* Keyboard shortcuts */
    GtkAccelGroup *accel_group;
    
    /* Notifications */
    gboolean notifications_enabled;
    
    /* Accessibility */
    gboolean high_contrast;
    gboolean large_text;
    
    /* Animation and effects */
    gboolean animations_enabled;
    gboolean transparency_enabled;
};

/* GObject implementation */
G_DEFINE_TYPE_WITH_PRIVATE(TaishangUI, taishang_ui, G_TYPE_OBJECT)

/* Property IDs */
enum {
    PROP_0,
    PROP_CONFIG,
    PROP_THEME,
    PROP_STATE,
    PROP_SIDEBAR_VISIBLE,
    PROP_TOOLBAR_VISIBLE,
    PROP_STATUS_BAR_VISIBLE,
    N_PROPERTIES
};

/* Signal IDs */
enum {
    SIGNAL_CLOSE_REQUEST,
    SIGNAL_THEME_CHANGED,
    SIGNAL_STATE_CHANGED,
    SIGNAL_MESSAGE_SENT,
    SIGNAL_FILE_SELECTED,
    N_SIGNALS
};

static GParamSpec *properties[N_PROPERTIES] = { NULL, };
static guint signals[N_SIGNALS] = { 0, };

/* Forward declarations */
static void taishang_ui_finalize(GObject *object);
static void taishang_ui_get_property(GObject *object, guint prop_id, GValue *value, GParamSpec *pspec);
static void taishang_ui_set_property(GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec);

static void taishang_ui_create_main_window(TaishangUI *ui);
static void taishang_ui_create_header_bar(TaishangUI *ui);
static void taishang_ui_create_menu_bar(TaishangUI *ui);
static void taishang_ui_create_toolbar(TaishangUI *ui);
static void taishang_ui_create_sidebar(TaishangUI *ui);
static void taishang_ui_create_chat_area(TaishangUI *ui);
static void taishang_ui_create_input_area(TaishangUI *ui);
static void taishang_ui_create_status_bar(TaishangUI *ui);
static void taishang_ui_create_dialogs(TaishangUI *ui);
static void taishang_ui_setup_css(TaishangUI *ui);
static void taishang_ui_setup_shortcuts(TaishangUI *ui);
static void taishang_ui_load_settings(TaishangUI *ui);
static void taishang_ui_save_settings(TaishangUI *ui);

/* Event handlers */
static gboolean on_window_delete_event(GtkWidget *widget, GdkEvent *event, gpointer user_data);
static void on_window_size_allocate(GtkWidget *widget, GtkAllocation *allocation, gpointer user_data);
static void on_window_state_event(GtkWidget *widget, GdkEventWindowState *event, gpointer user_data);
static void on_send_button_clicked(GtkButton *button, gpointer user_data);
static void on_message_entry_activate(GtkEntry *entry, gpointer user_data);
static void on_sidebar_toggle(GtkToggleButton *button, gpointer user_data);
static void on_theme_changed(GtkComboBox *combo, gpointer user_data);
static void on_preferences_activate(GtkMenuItem *item, gpointer user_data);
static void on_about_activate(GtkMenuItem *item, gpointer user_data);
static void on_quit_activate(GtkMenuItem *item, gpointer user_data);

/* Class initialization */
static void taishang_ui_class_init(TaishangUIClass *klass) {
    GObjectClass *object_class = G_OBJECT_CLASS(klass);
    
    object_class->finalize = taishang_ui_finalize;
    object_class->get_property = taishang_ui_get_property;
    object_class->set_property = taishang_ui_set_property;
    
    /* Properties */
    properties[PROP_CONFIG] = g_param_spec_object(
        "config", "Config", "Configuration object",
        TAISHANG_TYPE_CONFIG,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_THEME] = g_param_spec_enum(
        "theme", "Theme", "UI theme",
        TAISHANG_TYPE_UI_THEME, TAISHANG_UI_THEME_SYSTEM,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_STATE] = g_param_spec_enum(
        "state", "State", "UI state",
        TAISHANG_TYPE_UI_STATE, TAISHANG_UI_STATE_HIDDEN,
        TAISHANG_PARAM_READABLE);
    
    properties[PROP_SIDEBAR_VISIBLE] = g_param_spec_boolean(
        "sidebar-visible", "Sidebar Visible", "Whether sidebar is visible",
        TRUE,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_TOOLBAR_VISIBLE] = g_param_spec_boolean(
        "toolbar-visible", "Toolbar Visible", "Whether toolbar is visible",
        TRUE,
        TAISHANG_PARAM_READWRITE);
    
    properties[PROP_STATUS_BAR_VISIBLE] = g_param_spec_boolean(
        "status-bar-visible", "Status Bar Visible", "Whether status bar is visible",
        TRUE,
        TAISHANG_PARAM_READWRITE);
    
    g_object_class_install_properties(object_class, N_PROPERTIES, properties);
    
    /* Signals */
    signals[SIGNAL_CLOSE_REQUEST] = g_signal_new(
        "close-request",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangUIClass, close_request),
        NULL, NULL,
        g_cclosure_marshal_VOID__VOID,
        G_TYPE_NONE, 0);
    
    signals[SIGNAL_THEME_CHANGED] = g_signal_new(
        "theme-changed",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangUIClass, theme_changed),
        NULL, NULL,
        g_cclosure_marshal_VOID__ENUM,
        G_TYPE_NONE, 1, TAISHANG_TYPE_UI_THEME);
    
    signals[SIGNAL_STATE_CHANGED] = g_signal_new(
        "state-changed",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangUIClass, state_changed),
        NULL, NULL,
        g_cclosure_marshal_VOID__ENUM,
        G_TYPE_NONE, 1, TAISHANG_TYPE_UI_STATE);
    
    signals[SIGNAL_MESSAGE_SENT] = g_signal_new(
        "message-sent",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangUIClass, message_sent),
        NULL, NULL,
        g_cclosure_marshal_VOID__STRING,
        G_TYPE_NONE, 1, G_TYPE_STRING);
    
    signals[SIGNAL_FILE_SELECTED] = g_signal_new(
        "file-selected",
        G_TYPE_FROM_CLASS(klass),
        G_SIGNAL_RUN_LAST,
        G_STRUCT_OFFSET(TaishangUIClass, file_selected),
        NULL, NULL,
        g_cclosure_marshal_VOID__STRING,
        G_TYPE_NONE, 1, G_TYPE_STRING);
}

/* Instance initialization */
static void taishang_ui_init(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Initialize all pointers to NULL */
    memset(priv, 0, sizeof(TaishangUIPrivate));
    
    /* Set default values */
    priv->theme = TAISHANG_UI_THEME_SYSTEM;
    priv->state = TAISHANG_UI_STATE_HIDDEN;
    
    priv->window_width = 1200;
    priv->window_height = 800;
    priv->window_x = -1;
    priv->window_y = -1;
    priv->window_maximized = FALSE;
    priv->window_fullscreen = FALSE;
    
    priv->sidebar_visible = TRUE;
    priv->toolbar_visible = TRUE;
    priv->status_bar_visible = TRUE;
    priv->sidebar_width = 300;
    
    priv->notifications_enabled = TRUE;
    priv->high_contrast = FALSE;
    priv->large_text = FALSE;
    priv->animations_enabled = TRUE;
    priv->transparency_enabled = TRUE;
}

/* Object finalization */
static void taishang_ui_finalize(GObject *object) {
    TaishangUI *ui = TAISHANG_UI(object);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Save settings before cleanup */
    taishang_ui_save_settings(ui);
    
    /* Cleanup CSS provider */
    if (priv->css_provider) {
        g_object_unref(priv->css_provider);
    }
    
    /* Cleanup configuration */
    if (priv->config) {
        g_object_unref(priv->config);
    }
    
    /* Cleanup main window (this will cleanup all child widgets) */
    if (priv->main_window) {
        gtk_widget_destroy(priv->main_window);
    }
    
    G_OBJECT_CLASS(taishang_ui_parent_class)->finalize(object);
}

/* Property getters and setters */
static void taishang_ui_get_property(GObject *object, guint prop_id, GValue *value, GParamSpec *pspec) {
    TaishangUI *ui = TAISHANG_UI(object);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    switch (prop_id) {
        case PROP_CONFIG:
            g_value_set_object(value, priv->config);
            break;
        case PROP_THEME:
            g_value_set_enum(value, priv->theme);
            break;
        case PROP_STATE:
            g_value_set_enum(value, priv->state);
            break;
        case PROP_SIDEBAR_VISIBLE:
            g_value_set_boolean(value, priv->sidebar_visible);
            break;
        case PROP_TOOLBAR_VISIBLE:
            g_value_set_boolean(value, priv->toolbar_visible);
            break;
        case PROP_STATUS_BAR_VISIBLE:
            g_value_set_boolean(value, priv->status_bar_visible);
            break;
        default:
            G_OBJECT_WARN_INVALID_PROPERTY_ID(object, prop_id, pspec);
            break;
    }
}

static void taishang_ui_set_property(GObject *object, guint prop_id, const GValue *value, GParamSpec *pspec) {
    TaishangUI *ui = TAISHANG_UI(object);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    switch (prop_id) {
        case PROP_CONFIG:
            if (priv->config) {
                g_object_unref(priv->config);
            }
            priv->config = g_value_dup_object(value);
            break;
        case PROP_THEME:
            taishang_ui_set_theme(ui, g_value_get_enum(value));
            break;
        case PROP_SIDEBAR_VISIBLE:
            taishang_ui_set_sidebar_visible(ui, g_value_get_boolean(value));
            break;
        case PROP_TOOLBAR_VISIBLE:
            taishang_ui_set_toolbar_visible(ui, g_value_get_boolean(value));
            break;
        case PROP_STATUS_BAR_VISIBLE:
            taishang_ui_set_status_bar_visible(ui, g_value_get_boolean(value));
            break;
        default:
            G_OBJECT_WARN_INVALID_PROPERTY_ID(object, prop_id, pspec);
            break;
    }
}

/* Public API implementation */

/**
 * taishang_ui_new:
 * 
 * Creates a new TaishangUI instance.
 * 
 * Returns: (transfer full): A new TaishangUI instance
 */
TaishangUI *taishang_ui_new(void) {
    return g_object_new(TAISHANG_TYPE_UI, NULL);
}

/**
 * taishang_ui_initialize:
 * @ui: A TaishangUI instance
 * @error: Return location for error
 * 
 * Initializes the UI.
 * 
 * Returns: TRUE on success, FALSE on error
 */
gboolean taishang_ui_initialize(TaishangUI *ui, GError **error) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_UI(ui), FALSE);
    TAISHANG_RETURN_VAL_IF_FAIL(error == NULL || *error == NULL, FALSE);
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->state != TAISHANG_UI_STATE_HIDDEN) {
        g_set_error(error, TAISHANG_ERROR, TAISHANG_ERROR_INVALID_ARGUMENT,
                   "UI already initialized");
        return FALSE;
    }
    
    /* Create main window and components */
    taishang_ui_create_main_window(ui);
    taishang_ui_create_header_bar(ui);
    taishang_ui_create_menu_bar(ui);
    taishang_ui_create_toolbar(ui);
    taishang_ui_create_sidebar(ui);
    taishang_ui_create_chat_area(ui);
    taishang_ui_create_input_area(ui);
    taishang_ui_create_status_bar(ui);
    taishang_ui_create_dialogs(ui);
    
    /* Set up CSS and shortcuts */
    taishang_ui_setup_css(ui);
    taishang_ui_setup_shortcuts(ui);
    
    /* Load settings */
    taishang_ui_load_settings(ui);
    
    /* Update state */
    priv->state = TAISHANG_UI_STATE_READY;
    g_signal_emit(ui, signals[SIGNAL_STATE_CHANGED], 0, priv->state);
    
    g_info("UI initialized successfully");
    
    return TRUE;
}

/**
 * taishang_ui_show:
 * @ui: A TaishangUI instance
 * 
 * Shows the UI.
 */
void taishang_ui_show(TaishangUI *ui) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->main_window) {
        gtk_widget_show_all(priv->main_window);
        gtk_window_present(GTK_WINDOW(priv->main_window));
        
        priv->state = TAISHANG_UI_STATE_VISIBLE;
        g_signal_emit(ui, signals[SIGNAL_STATE_CHANGED], 0, priv->state);
    }
}

/**
 * taishang_ui_hide:
 * @ui: A TaishangUI instance
 * 
 * Hides the UI.
 */
void taishang_ui_hide(TaishangUI *ui) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->main_window) {
        gtk_widget_hide(priv->main_window);
        
        priv->state = TAISHANG_UI_STATE_HIDDEN;
        g_signal_emit(ui, signals[SIGNAL_STATE_CHANGED], 0, priv->state);
    }
}

/**
 * taishang_ui_set_config:
 * @ui: A TaishangUI instance
 * @config: Configuration object
 * 
 * Sets the configuration object.
 */
void taishang_ui_set_config(TaishangUI *ui, TaishangConfig *config) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    g_object_set(ui, "config", config, NULL);
}

/**
 * taishang_ui_get_config:
 * @ui: A TaishangUI instance
 * 
 * Gets the configuration object.
 * 
 * Returns: (transfer none): The configuration object
 */
TaishangConfig *taishang_ui_get_config(TaishangUI *ui) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_UI(ui), NULL);
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    return priv->config;
}

/**
 * taishang_ui_set_theme:
 * @ui: A TaishangUI instance
 * @theme: Theme to set
 * 
 * Sets the UI theme.
 */
void taishang_ui_set_theme(TaishangUI *ui, TaishangUITheme theme) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->theme != theme) {
        priv->theme = theme;
        
        /* Apply theme changes */
        taishang_ui_setup_css(ui);
        
        g_signal_emit(ui, signals[SIGNAL_THEME_CHANGED], 0, theme);
        g_object_notify_by_pspec(G_OBJECT(ui), properties[PROP_THEME]);
    }
}

/**
 * taishang_ui_get_theme:
 * @ui: A TaishangUI instance
 * 
 * Gets the current UI theme.
 * 
 * Returns: The current theme
 */
TaishangUITheme taishang_ui_get_theme(TaishangUI *ui) {
    TAISHANG_RETURN_VAL_IF_FAIL(TAISHANG_IS_UI(ui), TAISHANG_UI_THEME_SYSTEM);
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    return priv->theme;
}

/**
 * taishang_ui_set_sidebar_visible:
 * @ui: A TaishangUI instance
 * @visible: Whether to show sidebar
 * 
 * Sets sidebar visibility.
 */
void taishang_ui_set_sidebar_visible(TaishangUI *ui, gboolean visible) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->sidebar_visible != visible) {
        priv->sidebar_visible = visible;
        
        if (priv->sidebar) {
            gtk_widget_set_visible(priv->sidebar, visible);
        }
        
        g_object_notify_by_pspec(G_OBJECT(ui), properties[PROP_SIDEBAR_VISIBLE]);
    }
}

/**
 * taishang_ui_set_toolbar_visible:
 * @ui: A TaishangUI instance
 * @visible: Whether to show toolbar
 * 
 * Sets toolbar visibility.
 */
void taishang_ui_set_toolbar_visible(TaishangUI *ui, gboolean visible) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->toolbar_visible != visible) {
        priv->toolbar_visible = visible;
        
        if (priv->toolbar) {
            gtk_widget_set_visible(priv->toolbar, visible);
        }
        
        g_object_notify_by_pspec(G_OBJECT(ui), properties[PROP_TOOLBAR_VISIBLE]);
    }
}

/**
 * taishang_ui_set_status_bar_visible:
 * @ui: A TaishangUI instance
 * @visible: Whether to show status bar
 * 
 * Sets status bar visibility.
 */
void taishang_ui_set_status_bar_visible(TaishangUI *ui, gboolean visible) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->status_bar_visible != visible) {
        priv->status_bar_visible = visible;
        
        if (priv->status_bar) {
            gtk_widget_set_visible(priv->status_bar, visible);
        }
        
        g_object_notify_by_pspec(G_OBJECT(ui), properties[PROP_STATUS_BAR_VISIBLE]);
    }
}

/**
 * taishang_ui_add_message:
 * @ui: A TaishangUI instance
 * @message: Message text
 * @is_user: Whether message is from user
 * 
 * Adds a message to the chat area.
 */
void taishang_ui_add_message(TaishangUI *ui, const gchar *message, gboolean is_user) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    TAISHANG_RETURN_IF_FAIL(message != NULL);
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->chat_listbox) {
        GtkWidget *row = gtk_list_box_row_new();
        GtkWidget *label = gtk_label_new(message);
        
        gtk_label_set_line_wrap(GTK_LABEL(label), TRUE);
        gtk_label_set_line_wrap_mode(GTK_LABEL(label), PANGO_WRAP_WORD_CHAR);
        gtk_label_set_selectable(GTK_LABEL(label), TRUE);
        
        if (is_user) {
            gtk_widget_set_halign(label, GTK_ALIGN_END);
            gtk_style_context_add_class(gtk_widget_get_style_context(label), "user-message");
        } else {
            gtk_widget_set_halign(label, GTK_ALIGN_START);
            gtk_style_context_add_class(gtk_widget_get_style_context(label), "assistant-message");
        }
        
        gtk_container_add(GTK_CONTAINER(row), label);
        gtk_list_box_insert(GTK_LIST_BOX(priv->chat_listbox), row, -1);
        gtk_widget_show_all(row);
        
        /* Scroll to bottom */
        GtkAdjustment *adj = gtk_scrolled_window_get_vadjustment(GTK_SCROLLED_WINDOW(priv->chat_scrolled));
        gtk_adjustment_set_value(adj, gtk_adjustment_get_upper(adj));
    }
}

/**
 * taishang_ui_clear_messages:
 * @ui: A TaishangUI instance
 * 
 * Clears all messages from the chat area.
 */
void taishang_ui_clear_messages(TaishangUI *ui) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->chat_listbox) {
        GList *children = gtk_container_get_children(GTK_CONTAINER(priv->chat_listbox));
        for (GList *l = children; l != NULL; l = l->next) {
            gtk_widget_destroy(GTK_WIDGET(l->data));
        }
        g_list_free(children);
    }
}

/**
 * taishang_ui_set_status:
 * @ui: A TaishangUI instance
 * @status: Status message
 * 
 * Sets the status bar message.
 */
void taishang_ui_set_status(TaishangUI *ui, const gchar *status) {
    TAISHANG_RETURN_IF_FAIL(TAISHANG_IS_UI(ui));
    
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->status_bar) {
        gtk_statusbar_pop(GTK_STATUSBAR(priv->status_bar), 0);
        if (status && *status) {
            gtk_statusbar_push(GTK_STATUSBAR(priv->status_bar), 0, status);
        }
    }
}

/* Private helper functions */

static void taishang_ui_create_main_window(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create main window */
    priv->main_window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_title(GTK_WINDOW(priv->main_window), "TaishangLaojun");
    gtk_window_set_default_size(GTK_WINDOW(priv->main_window), priv->window_width, priv->window_height);
    gtk_window_set_icon_name(GTK_WINDOW(priv->main_window), "taishang-laojun");
    
    /* Connect signals */
    g_signal_connect(priv->main_window, "delete-event",
                    G_CALLBACK(on_window_delete_event), ui);
    g_signal_connect(priv->main_window, "size-allocate",
                    G_CALLBACK(on_window_size_allocate), ui);
    g_signal_connect(priv->main_window, "window-state-event",
                    G_CALLBACK(on_window_state_event), ui);
    
    /* Create main layout */
    priv->main_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
    gtk_container_add(GTK_CONTAINER(priv->main_window), priv->main_box);
}

static void taishang_ui_create_header_bar(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create header bar */
    priv->header_bar = gtk_header_bar_new();
    gtk_header_bar_set_title(GTK_HEADER_BAR(priv->header_bar), "TaishangLaojun");
    gtk_header_bar_set_subtitle(GTK_HEADER_BAR(priv->header_bar), "AI Assistant");
    gtk_header_bar_set_show_close_button(GTK_HEADER_BAR(priv->header_bar), TRUE);
    
    gtk_window_set_titlebar(GTK_WINDOW(priv->main_window), priv->header_bar);
    
    /* Add hamburger menu button */
    priv->hamburger_menu = gtk_menu_button_new();
    gtk_button_set_image(GTK_BUTTON(priv->hamburger_menu),
                        gtk_image_new_from_icon_name("open-menu-symbolic", GTK_ICON_SIZE_BUTTON));
    gtk_header_bar_pack_end(GTK_HEADER_BAR(priv->header_bar), priv->hamburger_menu);
}

static void taishang_ui_create_menu_bar(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create menu bar */
    priv->menu_bar = gtk_menu_bar_new();
    
    /* File menu */
    GtkWidget *file_menu = gtk_menu_new();
    GtkWidget *file_item = gtk_menu_item_new_with_label("File");
    gtk_menu_item_set_submenu(GTK_MENU_ITEM(file_item), file_menu);
    gtk_menu_shell_append(GTK_MENU_SHELL(priv->menu_bar), file_item);
    
    GtkWidget *quit_item = gtk_menu_item_new_with_label("Quit");
    g_signal_connect(quit_item, "activate", G_CALLBACK(on_quit_activate), ui);
    gtk_menu_shell_append(GTK_MENU_SHELL(file_menu), quit_item);
    
    /* Edit menu */
    GtkWidget *edit_menu = gtk_menu_new();
    GtkWidget *edit_item = gtk_menu_item_new_with_label("Edit");
    gtk_menu_item_set_submenu(GTK_MENU_ITEM(edit_item), edit_menu);
    gtk_menu_shell_append(GTK_MENU_SHELL(priv->menu_bar), edit_item);
    
    GtkWidget *preferences_item = gtk_menu_item_new_with_label("Preferences");
    g_signal_connect(preferences_item, "activate", G_CALLBACK(on_preferences_activate), ui);
    gtk_menu_shell_append(GTK_MENU_SHELL(edit_menu), preferences_item);
    
    /* Help menu */
    GtkWidget *help_menu = gtk_menu_new();
    GtkWidget *help_item = gtk_menu_item_new_with_label("Help");
    gtk_menu_item_set_submenu(GTK_MENU_ITEM(help_item), help_menu);
    gtk_menu_shell_append(GTK_MENU_SHELL(priv->menu_bar), help_item);
    
    GtkWidget *about_item = gtk_menu_item_new_with_label("About");
    g_signal_connect(about_item, "activate", G_CALLBACK(on_about_activate), ui);
    gtk_menu_shell_append(GTK_MENU_SHELL(help_menu), about_item);
    
    gtk_box_pack_start(GTK_BOX(priv->main_box), priv->menu_bar, FALSE, FALSE, 0);
}

static void taishang_ui_create_toolbar(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create toolbar */
    priv->toolbar = gtk_toolbar_new();
    gtk_toolbar_set_style(GTK_TOOLBAR(priv->toolbar), GTK_TOOLBAR_BOTH_HORIZ);
    
    /* Add toolbar buttons */
    GtkToolItem *sidebar_toggle = gtk_toggle_tool_button_new();
    gtk_tool_button_set_icon_name(GTK_TOOL_BUTTON(sidebar_toggle), "view-sidebar-symbolic");
    gtk_tool_item_set_tooltip_text(sidebar_toggle, "Toggle Sidebar");
    g_signal_connect(sidebar_toggle, "toggled", G_CALLBACK(on_sidebar_toggle), ui);
    gtk_toolbar_insert(GTK_TOOLBAR(priv->toolbar), sidebar_toggle, -1);
    
    gtk_box_pack_start(GTK_BOX(priv->main_box), priv->toolbar, FALSE, FALSE, 0);
}

static void taishang_ui_create_sidebar(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create sidebar */
    priv->sidebar = gtk_box_new(GTK_ORIENTATION_VERTICAL, 6);
    gtk_widget_set_size_request(priv->sidebar, priv->sidebar_width, -1);
    
    /* Create sidebar stack */
    priv->sidebar_stack = gtk_stack_new();
    gtk_box_pack_start(GTK_BOX(priv->sidebar), priv->sidebar_stack, TRUE, TRUE, 0);
    
    /* Conversation list */
    priv->conversation_list = gtk_list_box_new();
    gtk_stack_add_titled(GTK_STACK(priv->sidebar_stack), priv->conversation_list,
                        "conversations", "Conversations");
    
    /* Settings panel */
    priv->settings_panel = gtk_box_new(GTK_ORIENTATION_VERTICAL, 6);
    gtk_stack_add_titled(GTK_STACK(priv->sidebar_stack), priv->settings_panel,
                        "settings", "Settings");
    
    /* Add stack switcher */
    GtkWidget *stack_switcher = gtk_stack_switcher_new();
    gtk_stack_switcher_set_stack(GTK_STACK_SWITCHER(stack_switcher), GTK_STACK(priv->sidebar_stack));
    gtk_box_pack_start(GTK_BOX(priv->sidebar), stack_switcher, FALSE, FALSE, 0);
}

static void taishang_ui_create_chat_area(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create chat area */
    priv->chat_area = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
    
    /* Create scrolled window for messages */
    priv->chat_scrolled = gtk_scrolled_window_new(NULL, NULL);
    gtk_scrolled_window_set_policy(GTK_SCROLLED_WINDOW(priv->chat_scrolled),
                                  GTK_POLICY_NEVER, GTK_POLICY_AUTOMATIC);
    
    /* Create message list */
    priv->chat_listbox = gtk_list_box_new();
    gtk_list_box_set_selection_mode(GTK_LIST_BOX(priv->chat_listbox), GTK_SELECTION_NONE);
    gtk_container_add(GTK_CONTAINER(priv->chat_scrolled), priv->chat_listbox);
    
    gtk_box_pack_start(GTK_BOX(priv->chat_area), priv->chat_scrolled, TRUE, TRUE, 0);
}

static void taishang_ui_create_input_area(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create input area */
    priv->input_area = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    gtk_container_set_border_width(GTK_CONTAINER(priv->input_area), 6);
    
    /* Create message entry */
    priv->message_entry = gtk_entry_new();
    gtk_entry_set_placeholder_text(GTK_ENTRY(priv->message_entry), "Type your message...");
    g_signal_connect(priv->message_entry, "activate", G_CALLBACK(on_message_entry_activate), ui);
    gtk_box_pack_start(GTK_BOX(priv->input_area), priv->message_entry, TRUE, TRUE, 0);
    
    /* Create send button */
    priv->send_button = gtk_button_new_with_label("Send");
    gtk_style_context_add_class(gtk_widget_get_style_context(priv->send_button), "suggested-action");
    g_signal_connect(priv->send_button, "clicked", G_CALLBACK(on_send_button_clicked), ui);
    gtk_box_pack_start(GTK_BOX(priv->input_area), priv->send_button, FALSE, FALSE, 0);
    
    gtk_box_pack_start(GTK_BOX(priv->chat_area), priv->input_area, FALSE, FALSE, 0);
}

static void taishang_ui_create_status_bar(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create status bar */
    priv->status_bar = gtk_statusbar_new();
    
    /* Create progress bar */
    priv->progress_bar = gtk_progress_bar_new();
    gtk_widget_set_size_request(priv->progress_bar, 200, -1);
    gtk_widget_set_no_show_all(priv->progress_bar, TRUE);
    
    /* Pack into status bar */
    gtk_box_pack_end(GTK_BOX(priv->status_bar), priv->progress_bar, FALSE, FALSE, 0);
    
    gtk_box_pack_start(GTK_BOX(priv->main_box), priv->status_bar, FALSE, FALSE, 0);
}

static void taishang_ui_create_dialogs(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create preferences dialog */
    priv->preferences_dialog = gtk_dialog_new_with_buttons(
        "Preferences",
        GTK_WINDOW(priv->main_window),
        GTK_DIALOG_MODAL | GTK_DIALOG_DESTROY_WITH_PARENT,
        "Close", GTK_RESPONSE_CLOSE,
        NULL);
    
    /* Create about dialog */
    priv->about_dialog = gtk_about_dialog_new();
    gtk_about_dialog_set_program_name(GTK_ABOUT_DIALOG(priv->about_dialog), "TaishangLaojun");
    gtk_about_dialog_set_version(GTK_ABOUT_DIALOG(priv->about_dialog), TAISHANG_VERSION);
    gtk_about_dialog_set_comments(GTK_ABOUT_DIALOG(priv->about_dialog), "AI Assistant Desktop Application");
    gtk_window_set_transient_for(GTK_WINDOW(priv->about_dialog), GTK_WINDOW(priv->main_window));
}

static void taishang_ui_setup_css(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create CSS provider */
    if (!priv->css_provider) {
        priv->css_provider = gtk_css_provider_new();
    }
    
    /* Load CSS based on theme */
    const gchar *css_data = 
        ".user-message { "
        "  background-color: #2196F3; "
        "  color: white; "
        "  border-radius: 12px; "
        "  padding: 8px 12px; "
        "  margin: 4px; "
        "} "
        ".assistant-message { "
        "  background-color: #f5f5f5; "
        "  color: black; "
        "  border-radius: 12px; "
        "  padding: 8px 12px; "
        "  margin: 4px; "
        "}";
    
    gtk_css_provider_load_from_data(priv->css_provider, css_data, -1, NULL);
    
    /* Apply CSS */
    GdkScreen *screen = gdk_screen_get_default();
    gtk_style_context_add_provider_for_screen(screen,
                                             GTK_STYLE_PROVIDER(priv->css_provider),
                                             GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
}

static void taishang_ui_setup_shortcuts(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create accelerator group */
    priv->accel_group = gtk_accel_group_new();
    gtk_window_add_accel_group(GTK_WINDOW(priv->main_window), priv->accel_group);
    
    /* Add keyboard shortcuts */
    /* Ctrl+Q for quit */
    gtk_accel_group_connect(priv->accel_group, GDK_KEY_q, GDK_CONTROL_MASK, 0,
                           g_cclosure_new_swap(G_CALLBACK(on_quit_activate), ui, NULL));
    
    /* Ctrl+, for preferences */
    gtk_accel_group_connect(priv->accel_group, GDK_KEY_comma, GDK_CONTROL_MASK, 0,
                           g_cclosure_new_swap(G_CALLBACK(on_preferences_activate), ui, NULL));
}

static void taishang_ui_load_settings(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->config) {
        /* Load window settings */
        priv->window_width = taishang_config_get_int(priv->config, "ui", "window-width", 1200);
        priv->window_height = taishang_config_get_int(priv->config, "ui", "window-height", 800);
        priv->window_maximized = taishang_config_get_boolean(priv->config, "ui", "window-maximized", FALSE);
        
        /* Load UI settings */
        priv->sidebar_visible = taishang_config_get_boolean(priv->config, "ui", "sidebar-visible", TRUE);
        priv->toolbar_visible = taishang_config_get_boolean(priv->config, "ui", "toolbar-visible", TRUE);
        priv->status_bar_visible = taishang_config_get_boolean(priv->config, "ui", "status-bar-visible", TRUE);
        
        /* Apply settings */
        if (priv->main_window) {
            gtk_window_resize(GTK_WINDOW(priv->main_window), priv->window_width, priv->window_height);
            if (priv->window_maximized) {
                gtk_window_maximize(GTK_WINDOW(priv->main_window));
            }
        }
        
        taishang_ui_set_sidebar_visible(ui, priv->sidebar_visible);
        taishang_ui_set_toolbar_visible(ui, priv->toolbar_visible);
        taishang_ui_set_status_bar_visible(ui, priv->status_bar_visible);
    }
}

static void taishang_ui_save_settings(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->config && priv->main_window) {
        /* Save window settings */
        taishang_config_set_int(priv->config, "ui", "window-width", priv->window_width);
        taishang_config_set_int(priv->config, "ui", "window-height", priv->window_height);
        taishang_config_set_boolean(priv->config, "ui", "window-maximized", priv->window_maximized);
        
        /* Save UI settings */
        taishang_config_set_boolean(priv->config, "ui", "sidebar-visible", priv->sidebar_visible);
        taishang_config_set_boolean(priv->config, "ui", "toolbar-visible", priv->toolbar_visible);
        taishang_config_set_boolean(priv->config, "ui", "status-bar-visible", priv->status_bar_visible);
    }
}

/* Event handlers */

static gboolean on_window_delete_event(GtkWidget *widget, GdkEvent *event, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    
    g_signal_emit(ui, signals[SIGNAL_CLOSE_REQUEST], 0);
    
    return TRUE; /* Prevent default handler */
}

static void on_window_size_allocate(GtkWidget *widget, GtkAllocation *allocation, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (!priv->window_maximized && !priv->window_fullscreen) {
        priv->window_width = allocation->width;
        priv->window_height = allocation->height;
    }
}

static void on_window_state_event(GtkWidget *widget, GdkEventWindowState *event, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    priv->window_maximized = (event->new_window_state & GDK_WINDOW_STATE_MAXIMIZED) != 0;
    priv->window_fullscreen = (event->new_window_state & GDK_WINDOW_STATE_FULLSCREEN) != 0;
}

static void on_send_button_clicked(GtkButton *button, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    const gchar *text = gtk_entry_get_text(GTK_ENTRY(priv->message_entry));
    if (text && *text) {
        g_signal_emit(ui, signals[SIGNAL_MESSAGE_SENT], 0, text);
        gtk_entry_set_text(GTK_ENTRY(priv->message_entry), "");
    }
}

static void on_message_entry_activate(GtkEntry *entry, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    on_send_button_clicked(GTK_BUTTON(priv->send_button), user_data);
}

static void on_sidebar_toggle(GtkToggleButton *button, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    
    gboolean active = gtk_toggle_button_get_active(button);
    taishang_ui_set_sidebar_visible(ui, active);
}

static void on_theme_changed(GtkComboBox *combo, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    
    gint active = gtk_combo_box_get_active(combo);
    taishang_ui_set_theme(ui, (TaishangUITheme)active);
}

static void on_preferences_activate(GtkMenuItem *item, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->preferences_dialog) {
        gtk_dialog_run(GTK_DIALOG(priv->preferences_dialog));
        gtk_widget_hide(priv->preferences_dialog);
    }
}

static void on_about_activate(GtkMenuItem *item, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    if (priv->about_dialog) {
        gtk_dialog_run(GTK_DIALOG(priv->about_dialog));
        gtk_widget_hide(priv->about_dialog);
    }
}

static void on_quit_activate(GtkMenuItem *item, gpointer user_data) {
    TaishangUI *ui = TAISHANG_UI(user_data);
    
    g_signal_emit(ui, signals[SIGNAL_CLOSE_REQUEST], 0);
}

/* Layout management */

static void taishang_ui_setup_layout(TaishangUI *ui) {
    TaishangUIPrivate *priv = taishang_ui_get_instance_private(ui);
    
    /* Create main content paned */
    priv->content_paned = gtk_paned_new(GTK_ORIENTATION_HORIZONTAL);
    gtk_box_pack_start(GTK_BOX(priv->main_box), priv->content_paned, TRUE, TRUE, 0);
    
    /* Add sidebar to left pane */
    gtk_paned_pack1(GTK_PANED(priv->content_paned), priv->sidebar, FALSE, FALSE);
    
    /* Add chat area to right pane */
    gtk_paned_pack2(GTK_PANED(priv->content_paned), priv->chat_area, TRUE, FALSE);
    
    /* Set paned position */
    gtk_paned_set_position(GTK_PANED(priv->content_paned), priv->sidebar_width);
}