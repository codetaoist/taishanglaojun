#include <gtk/gtk.h>
#include <adwaita.h>
#include <glib.h>
#include <json-c/json.h>
#include "../../include/ui/gtk_helpers.h"

// Project data structure
typedef struct {
    char *name;
    char *path;
    char *description;
    char *language;
    char *last_modified;
    gboolean is_favorite;
} ProjectInfo;

// Project view structure
typedef struct {
    GtkWidget *main_box;
    GtkWidget *toolbar;
    GtkWidget *search_entry;
    GtkWidget *project_list;
    GtkWidget *project_details;
    GtkWidget *status_label;
    
    // Project management
    GList *projects;
    ProjectInfo *selected_project;
    
    // Callbacks
    void (*on_project_selected)(ProjectInfo *project, gpointer user_data);
    void (*on_project_opened)(ProjectInfo *project, gpointer user_data);
    gpointer user_data;
} TaishangProjectView;

// Forward declarations
static void project_info_free(ProjectInfo *project);
static ProjectInfo *project_info_new(const char *name, const char *path, const char *description);
static void project_view_init(TaishangProjectView *view);
static void project_view_setup_toolbar(TaishangProjectView *view);
static void project_view_setup_project_list(TaishangProjectView *view);
static void project_view_setup_project_details(TaishangProjectView *view);
static void project_view_refresh_list(TaishangProjectView *view);
static void project_view_update_details(TaishangProjectView *view, ProjectInfo *project);

// Callback functions
static void on_new_project_clicked(GtkButton *button, TaishangProjectView *view);
static void on_open_project_clicked(GtkButton *button, TaishangProjectView *view);
static void on_import_project_clicked(GtkButton *button, TaishangProjectView *view);
static void on_search_changed(GtkSearchEntry *entry, TaishangProjectView *view);
static void on_project_row_activated(GtkListBox *list_box, GtkListBoxRow *row, TaishangProjectView *view);
static void on_favorite_toggled(GtkToggleButton *button, ProjectInfo *project);

// Public functions
TaishangProjectView *taishang_project_view_new(void);
void taishang_project_view_free(TaishangProjectView *view);
GtkWidget *taishang_project_view_get_widget(TaishangProjectView *view);
void taishang_project_view_set_callbacks(TaishangProjectView *view,
                                          void (*on_project_selected)(ProjectInfo *project, gpointer user_data),
                                          void (*on_project_opened)(ProjectInfo *project, gpointer user_data),
                                          gpointer user_data);
void taishang_project_view_add_project(TaishangProjectView *view, const char *name, const char *path, const char *description);
void taishang_project_view_remove_project(TaishangProjectView *view, const char *path);
void taishang_project_view_refresh(TaishangProjectView *view);

// Implementation
static void project_info_free(ProjectInfo *project) {
    if (!project) return;
    
    g_free(project->name);
    g_free(project->path);
    g_free(project->description);
    g_free(project->language);
    g_free(project->last_modified);
    g_free(project);
}

static ProjectInfo *project_info_new(const char *name, const char *path, const char *description) {
    ProjectInfo *project = g_new0(ProjectInfo, 1);
    project->name = g_strdup(name ? name : "未命名项目");
    project->path = g_strdup(path ? path : "");
    project->description = g_strdup(description ? description : "");
    project->language = g_strdup("Unknown");
    project->last_modified = g_strdup("未知");
    project->is_favorite = FALSE;
    
    return project;
}

TaishangProjectView *taishang_project_view_new(void) {
    TaishangProjectView *view = g_new0(TaishangProjectView, 1);
    project_view_init(view);
    return view;
}

void taishang_project_view_free(TaishangProjectView *view) {
    if (!view) return;
    
    // Free project list
    g_list_free_full(view->projects, (GDestroyNotify)project_info_free);
    
    g_free(view);
}

GtkWidget *taishang_project_view_get_widget(TaishangProjectView *view) {
    return view ? view->main_box : NULL;
}

static void project_view_init(TaishangProjectView *view) {
    // Create main layout
    view->main_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
    
    // Setup components
    project_view_setup_toolbar(view);
    
    // Create horizontal paned layout
    GtkWidget *paned = gtk_paned_new(GTK_ORIENTATION_HORIZONTAL);
    gtk_widget_set_vexpand(paned, TRUE);
    gtk_widget_set_hexpand(paned, TRUE);
    
    // Setup project list and details
    project_view_setup_project_list(view);
    project_view_setup_project_details(view);
    
    // Add to paned layout
    gtk_paned_set_start_child(GTK_PANED(paned), view->project_list);
    gtk_paned_set_end_child(GTK_PANED(paned), view->project_details);
    gtk_paned_set_position(GTK_PANED(paned), 300);
    
    // Add to main box
    gtk_box_append(GTK_BOX(view->main_box), view->toolbar);
    gtk_box_append(GTK_BOX(view->main_box), paned);
    
    // Status label
    view->status_label = gtk_label_new("就绪");
    gtk_widget_set_halign(view->status_label, GTK_ALIGN_START);
    taishang_gtk_set_margins(view->status_label, 6);
    gtk_box_append(GTK_BOX(view->main_box), view->status_label);
    
    // Initialize with some sample projects
    taishang_project_view_add_project(view, "太上老君", "/home/user/taishanglaojun", "AI助手桌面应用");
    taishang_project_view_add_project(view, "示例项目", "/home/user/example", "示例项目描述");
}

static void project_view_setup_toolbar(TaishangProjectView *view) {
    view->toolbar = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    taishang_gtk_set_margins(view->toolbar, 12);
    
    // New project button
    GtkWidget *new_button = taishang_gtk_create_button_with_icon("document-new-symbolic", "新建项目");
    taishang_gtk_add_css_class(new_button, "suggested-action");
    g_signal_connect(new_button, "clicked", G_CALLBACK(on_new_project_clicked), view);
    gtk_box_append(GTK_BOX(view->toolbar), new_button);
    
    // Open project button
    GtkWidget *open_button = taishang_gtk_create_button_with_icon("document-open-symbolic", "打开项目");
    g_signal_connect(open_button, "clicked", G_CALLBACK(on_open_project_clicked), view);
    gtk_box_append(GTK_BOX(view->toolbar), open_button);
    
    // Import project button
    GtkWidget *import_button = taishang_gtk_create_button_with_icon("document-import-symbolic", "导入项目");
    g_signal_connect(import_button, "clicked", G_CALLBACK(on_import_project_clicked), view);
    gtk_box_append(GTK_BOX(view->toolbar), import_button);
    
    // Separator
    GtkWidget *separator = gtk_separator_new(GTK_ORIENTATION_VERTICAL);
    gtk_box_append(GTK_BOX(view->toolbar), separator);
    
    // Search entry
    view->search_entry = taishang_gtk_create_search_entry("搜索项目...");
    gtk_widget_set_hexpand(view->search_entry, TRUE);
    g_signal_connect(view->search_entry, "search-changed", G_CALLBACK(on_search_changed), view);
    gtk_box_append(GTK_BOX(view->toolbar), view->search_entry);
}

static void project_view_setup_project_list(TaishangProjectView *view) {
    // Create scrolled window for project list
    GtkWidget *scrolled = gtk_scrolled_window_new();
    gtk_scrolled_window_set_policy(GTK_SCROLLED_WINDOW(scrolled),
                                   GTK_POLICY_NEVER,
                                   GTK_POLICY_AUTOMATIC);
    gtk_widget_set_size_request(scrolled, 300, -1);
    
    // Create list box
    GtkWidget *list_box = gtk_list_box_new();
    gtk_list_box_set_selection_mode(GTK_LIST_BOX(list_box), GTK_SELECTION_SINGLE);
    g_signal_connect(list_box, "row-activated", G_CALLBACK(on_project_row_activated), view);
    
    gtk_scrolled_window_set_child(GTK_SCROLLED_WINDOW(scrolled), list_box);
    view->project_list = scrolled;
    
    // Store list box reference for later use
    g_object_set_data(G_OBJECT(view->project_list), "list_box", list_box);
}

static void project_view_setup_project_details(TaishangProjectView *view) {
    view->project_details = gtk_box_new(GTK_ORIENTATION_VERTICAL, 12);
    taishang_gtk_set_margins(view->project_details, 12);
    
    // Title
    GtkWidget *title_label = gtk_label_new("项目详情");
    taishang_gtk_add_css_class(title_label, "title-2");
    gtk_widget_set_halign(title_label, GTK_ALIGN_START);
    gtk_box_append(GTK_BOX(view->project_details), title_label);
    
    // Placeholder content
    GtkWidget *placeholder = gtk_label_new("选择一个项目查看详情");
    taishang_gtk_add_css_class(placeholder, "dim-label");
    gtk_widget_set_valign(placeholder, GTK_ALIGN_CENTER);
    gtk_widget_set_vexpand(placeholder, TRUE);
    gtk_box_append(GTK_BOX(view->project_details), placeholder);
    
    g_object_set_data(G_OBJECT(view->project_details), "placeholder", placeholder);
}

static void project_view_refresh_list(TaishangProjectView *view) {
    GtkWidget *list_box = g_object_get_data(G_OBJECT(view->project_list), "list_box");
    if (!list_box) return;
    
    // Clear existing items
    GtkWidget *child = gtk_widget_get_first_child(list_box);
    while (child) {
        GtkWidget *next = gtk_widget_get_next_sibling(child);
        gtk_list_box_remove(GTK_LIST_BOX(list_box), child);
        child = next;
    }
    
    // Add projects
    for (GList *l = view->projects; l != NULL; l = l->next) {
        ProjectInfo *project = (ProjectInfo *)l->data;
        
        // Create project row
        GtkWidget *row = gtk_list_box_row_new();
        GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 12);
        taishang_gtk_set_margins(box, 8);
        
        // Project icon
        GtkWidget *icon = gtk_image_new_from_icon_name("folder-symbolic");
        gtk_box_append(GTK_BOX(box), icon);
        
        // Project info
        GtkWidget *info_box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 4);
        gtk_widget_set_hexpand(info_box, TRUE);
        
        GtkWidget *name_label = gtk_label_new(project->name);
        taishang_gtk_add_css_class(name_label, "heading");
        gtk_widget_set_halign(name_label, GTK_ALIGN_START);
        gtk_box_append(GTK_BOX(info_box), name_label);
        
        GtkWidget *path_label = gtk_label_new(project->path);
        taishang_gtk_add_css_class(path_label, "caption");
        taishang_gtk_add_css_class(path_label, "dim-label");
        gtk_widget_set_halign(path_label, GTK_ALIGN_START);
        gtk_label_set_ellipsize(GTK_LABEL(path_label), PANGO_ELLIPSIZE_MIDDLE);
        gtk_box_append(GTK_BOX(info_box), path_label);
        
        gtk_box_append(GTK_BOX(box), info_box);
        
        // Favorite button
        GtkWidget *favorite_button = gtk_toggle_button_new();
        gtk_button_set_icon_name(GTK_BUTTON(favorite_button), 
                                 project->is_favorite ? "starred-symbolic" : "non-starred-symbolic");
        gtk_toggle_button_set_active(GTK_TOGGLE_BUTTON(favorite_button), project->is_favorite);
        g_signal_connect(favorite_button, "toggled", G_CALLBACK(on_favorite_toggled), project);
        gtk_box_append(GTK_BOX(box), favorite_button);
        
        gtk_list_box_row_set_child(GTK_LIST_BOX_ROW(row), box);
        g_object_set_data(G_OBJECT(row), "project", project);
        
        gtk_list_box_append(GTK_LIST_BOX(list_box), row);
    }
}

static void project_view_update_details(TaishangProjectView *view, ProjectInfo *project) {
    // Clear existing content
    GtkWidget *child = gtk_widget_get_first_child(view->project_details);
    while (child) {
        GtkWidget *next = gtk_widget_get_next_sibling(child);
        if (g_object_get_data(G_OBJECT(view->project_details), "placeholder") != child) {
            gtk_box_remove(GTK_BOX(view->project_details), child);
        }
        child = next;
    }
    
    if (!project) {
        // Show placeholder
        GtkWidget *placeholder = g_object_get_data(G_OBJECT(view->project_details), "placeholder");
        gtk_widget_set_visible(placeholder, TRUE);
        return;
    }
    
    // Hide placeholder
    GtkWidget *placeholder = g_object_get_data(G_OBJECT(view->project_details), "placeholder");
    gtk_widget_set_visible(placeholder, FALSE);
    
    // Add project details
    GtkWidget *name_label = gtk_label_new(project->name);
    taishang_gtk_add_css_class(name_label, "title-1");
    gtk_widget_set_halign(name_label, GTK_ALIGN_START);
    gtk_box_append(GTK_BOX(view->project_details), name_label);
    
    GtkWidget *path_label = gtk_label_new(project->path);
    taishang_gtk_add_css_class(path_label, "caption");
    gtk_widget_set_halign(path_label, GTK_ALIGN_START);
    gtk_label_set_selectable(GTK_LABEL(path_label), TRUE);
    gtk_box_append(GTK_BOX(view->project_details), path_label);
    
    if (project->description && strlen(project->description) > 0) {
        GtkWidget *desc_label = gtk_label_new(project->description);
        gtk_widget_set_halign(desc_label, GTK_ALIGN_START);
        gtk_label_set_wrap(GTK_LABEL(desc_label), TRUE);
        gtk_box_append(GTK_BOX(view->project_details), desc_label);
    }
    
    // Action buttons
    GtkWidget *button_box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 6);
    
    GtkWidget *open_button = gtk_button_new_with_label("打开项目");
    taishang_gtk_add_css_class(open_button, "suggested-action");
    gtk_box_append(GTK_BOX(button_box), open_button);
    
    GtkWidget *remove_button = gtk_button_new_with_label("移除");
    taishang_gtk_add_css_class(remove_button, "destructive-action");
    gtk_box_append(GTK_BOX(button_box), remove_button);
    
    gtk_box_append(GTK_BOX(view->project_details), button_box);
}

// Callback implementations
static void on_new_project_clicked(GtkButton *button, TaishangProjectView *view) {
    g_print("New project clicked\n");
    // TODO: Show new project dialog
}

static void on_open_project_clicked(GtkButton *button, TaishangProjectView *view) {
    g_print("Open project clicked\n");
    // TODO: Show file chooser dialog
}

static void on_import_project_clicked(GtkButton *button, TaishangProjectView *view) {
    g_print("Import project clicked\n");
    // TODO: Show import dialog
}

static void on_search_changed(GtkSearchEntry *entry, TaishangProjectView *view) {
    const char *search_text = gtk_editable_get_text(GTK_EDITABLE(entry));
    g_print("Search changed: %s\n", search_text);
    // TODO: Filter project list based on search text
}

static void on_project_row_activated(GtkListBox *list_box, GtkListBoxRow *row, TaishangProjectView *view) {
    ProjectInfo *project = g_object_get_data(G_OBJECT(row), "project");
    if (project) {
        view->selected_project = project;
        project_view_update_details(view, project);
        
        if (view->on_project_selected) {
            view->on_project_selected(project, view->user_data);
        }
        
        g_print("Project selected: %s\n", project->name);
    }
}

static void on_favorite_toggled(GtkToggleButton *button, ProjectInfo *project) {
    project->is_favorite = gtk_toggle_button_get_active(button);
    gtk_button_set_icon_name(GTK_BUTTON(button), 
                             project->is_favorite ? "starred-symbolic" : "non-starred-symbolic");
    g_print("Project %s favorite status: %s\n", project->name, project->is_favorite ? "true" : "false");
}

// Public function implementations
void taishang_project_view_set_callbacks(TaishangProjectView *view,
                                          void (*on_project_selected)(ProjectInfo *project, gpointer user_data),
                                          void (*on_project_opened)(ProjectInfo *project, gpointer user_data),
                                          gpointer user_data) {
    if (!view) return;
    
    view->on_project_selected = on_project_selected;
    view->on_project_opened = on_project_opened;
    view->user_data = user_data;
}

void taishang_project_view_add_project(TaishangProjectView *view, const char *name, const char *path, const char *description) {
    if (!view) return;
    
    ProjectInfo *project = project_info_new(name, path, description);
    view->projects = g_list_append(view->projects, project);
    
    project_view_refresh_list(view);
}

void taishang_project_view_remove_project(TaishangProjectView *view, const char *path) {
    if (!view || !path) return;
    
    for (GList *l = view->projects; l != NULL; l = l->next) {
        ProjectInfo *project = (ProjectInfo *)l->data;
        if (g_strcmp0(project->path, path) == 0) {
            view->projects = g_list_remove(view->projects, project);
            project_info_free(project);
            break;
        }
    }
    
    project_view_refresh_list(view);
}

void taishang_project_view_refresh(TaishangProjectView *view) {
    if (!view) return;
    
    project_view_refresh_list(view);
}