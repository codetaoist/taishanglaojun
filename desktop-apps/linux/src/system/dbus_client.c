#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <glib.h>
#include <gio/gio.h>
#include "../../include/system/dbus_client.h"

// D-Bus service names and paths
#define NOTIFICATIONS_SERVICE "org.freedesktop.Notifications"
#define NOTIFICATIONS_PATH "/org/freedesktop/Notifications"
#define NOTIFICATIONS_INTERFACE "org.freedesktop.Notifications"

#define SCREENSAVER_SERVICE "org.freedesktop.ScreenSaver"
#define SCREENSAVER_PATH "/org/freedesktop/ScreenSaver"
#define SCREENSAVER_INTERFACE "org.freedesktop.ScreenSaver"

#define POWER_SERVICE "org.freedesktop.UPower"
#define POWER_PATH "/org/freedesktop/UPower"
#define POWER_INTERFACE "org.freedesktop.UPower"

#define NETWORK_MANAGER_SERVICE "org.freedesktop.NetworkManager"
#define NETWORK_MANAGER_PATH "/org/freedesktop/NetworkManager"
#define NETWORK_MANAGER_INTERFACE "org.freedesktop.NetworkManager"

// D-Bus client structure
typedef struct {
    GDBusConnection *connection;
    GDBusProxy *notifications_proxy;
    GDBusProxy *screensaver_proxy;
    GDBusProxy *power_proxy;
    GDBusProxy *network_proxy;
    
    gboolean initialized;
    GMutex mutex;
    
    // Callbacks
    TaishangDBusNotificationCallback notification_callback;
    TaishangDBusScreensaverCallback screensaver_callback;
    TaishangDBusPowerCallback power_callback;
    TaishangDBusNetworkCallback network_callback;
    void *user_data;
} TaishangDBusClient;

static TaishangDBusClient *dbus_client = NULL;

// Forward declarations
static void on_notification_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                                   GVariant *parameters, gpointer user_data);
static void on_screensaver_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                                  GVariant *parameters, gpointer user_data);
static void on_power_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                            GVariant *parameters, gpointer user_data);
static void on_network_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                              GVariant *parameters, gpointer user_data);

// Public functions
gboolean taishang_dbus_client_init(void);
void taishang_dbus_client_cleanup(void);
TaishangDBusClient *taishang_dbus_client_get_instance(void);

// Notification functions
guint32 taishang_dbus_send_notification(const char *app_name, const char *summary, 
                                        const char *body, const char *icon, 
                                        gint32 timeout, const char **actions);
gboolean taishang_dbus_close_notification(guint32 notification_id);
gboolean taishang_dbus_get_server_information(char **name, char **vendor, 
                                              char **version, char **spec_version);

// Screensaver functions
gboolean taishang_dbus_inhibit_screensaver(const char *app_name, const char *reason, guint32 *cookie);
gboolean taishang_dbus_uninhibit_screensaver(guint32 cookie);
gboolean taishang_dbus_get_screensaver_active(gboolean *active);

// Power management functions
gboolean taishang_dbus_get_power_info(TaishangPowerInfo *info);
gboolean taishang_dbus_suspend_system(void);
gboolean taishang_dbus_hibernate_system(void);

// Network functions
gboolean taishang_dbus_get_network_state(TaishangNetworkState *state);
GList *taishang_dbus_get_network_connections(void);

// Callback registration
void taishang_dbus_set_notification_callback(TaishangDBusNotificationCallback callback, void *user_data);
void taishang_dbus_set_screensaver_callback(TaishangDBusScreensaverCallback callback, void *user_data);
void taishang_dbus_set_power_callback(TaishangDBusPowerCallback callback, void *user_data);
void taishang_dbus_set_network_callback(TaishangDBusNetworkCallback callback, void *user_data);

// Implementation
gboolean taishang_dbus_client_init(void) {
    if (dbus_client != NULL) {
        g_warning("D-Bus client already initialized");
        return FALSE;
    }
    
    dbus_client = g_new0(TaishangDBusClient, 1);
    g_mutex_init(&dbus_client->mutex);
    
    GError *error = NULL;
    
    // Connect to session bus
    dbus_client->connection = g_bus_get_sync(G_BUS_TYPE_SESSION, NULL, &error);
    if (!dbus_client->connection) {
        g_error("Failed to connect to D-Bus: %s", error->message);
        g_error_free(error);
        g_free(dbus_client);
        dbus_client = NULL;
        return FALSE;
    }
    
    // Create notifications proxy
    dbus_client->notifications_proxy = g_dbus_proxy_new_sync(
        dbus_client->connection,
        G_DBUS_PROXY_FLAGS_NONE,
        NULL,
        NOTIFICATIONS_SERVICE,
        NOTIFICATIONS_PATH,
        NOTIFICATIONS_INTERFACE,
        NULL,
        &error
    );
    
    if (!dbus_client->notifications_proxy) {
        g_warning("Failed to create notifications proxy: %s", error->message);
        g_error_free(error);
    } else {
        g_signal_connect(dbus_client->notifications_proxy, "g-signal", 
                         G_CALLBACK(on_notification_signal), NULL);
    }
    
    // Create screensaver proxy
    dbus_client->screensaver_proxy = g_dbus_proxy_new_sync(
        dbus_client->connection,
        G_DBUS_PROXY_FLAGS_NONE,
        NULL,
        SCREENSAVER_SERVICE,
        SCREENSAVER_PATH,
        SCREENSAVER_INTERFACE,
        NULL,
        &error
    );
    
    if (!dbus_client->screensaver_proxy) {
        g_warning("Failed to create screensaver proxy: %s", error->message);
        g_error_free(error);
    } else {
        g_signal_connect(dbus_client->screensaver_proxy, "g-signal", 
                         G_CALLBACK(on_screensaver_signal), NULL);
    }
    
    // Create power proxy
    dbus_client->power_proxy = g_dbus_proxy_new_sync(
        dbus_client->connection,
        G_DBUS_PROXY_FLAGS_NONE,
        NULL,
        POWER_SERVICE,
        POWER_PATH,
        POWER_INTERFACE,
        NULL,
        &error
    );
    
    if (!dbus_client->power_proxy) {
        g_warning("Failed to create power proxy: %s", error->message);
        g_error_free(error);
    } else {
        g_signal_connect(dbus_client->power_proxy, "g-signal", 
                         G_CALLBACK(on_power_signal), NULL);
    }
    
    // Create network proxy
    dbus_client->network_proxy = g_dbus_proxy_new_sync(
        dbus_client->connection,
        G_DBUS_PROXY_FLAGS_NONE,
        NULL,
        NETWORK_MANAGER_SERVICE,
        NETWORK_MANAGER_PATH,
        NETWORK_MANAGER_INTERFACE,
        NULL,
        &error
    );
    
    if (!dbus_client->network_proxy) {
        g_warning("Failed to create network proxy: %s", error->message);
        g_error_free(error);
    } else {
        g_signal_connect(dbus_client->network_proxy, "g-signal", 
                         G_CALLBACK(on_network_signal), NULL);
    }
    
    dbus_client->initialized = TRUE;
    g_print("D-Bus client initialized\n");
    return TRUE;
}

void taishang_dbus_client_cleanup(void) {
    if (dbus_client == NULL) {
        return;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    if (dbus_client->notifications_proxy) {
        g_object_unref(dbus_client->notifications_proxy);
    }
    
    if (dbus_client->screensaver_proxy) {
        g_object_unref(dbus_client->screensaver_proxy);
    }
    
    if (dbus_client->power_proxy) {
        g_object_unref(dbus_client->power_proxy);
    }
    
    if (dbus_client->network_proxy) {
        g_object_unref(dbus_client->network_proxy);
    }
    
    if (dbus_client->connection) {
        g_object_unref(dbus_client->connection);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    g_mutex_clear(&dbus_client->mutex);
    
    g_free(dbus_client);
    dbus_client = NULL;
    
    g_print("D-Bus client cleaned up\n");
}

TaishangDBusClient *taishang_dbus_client_get_instance(void) {
    return dbus_client;
}

// Notification functions
guint32 taishang_dbus_send_notification(const char *app_name, const char *summary, 
                                        const char *body, const char *icon, 
                                        gint32 timeout, const char **actions) {
    if (!dbus_client || !dbus_client->notifications_proxy) {
        return 0;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    // Build actions array
    GVariantBuilder actions_builder;
    g_variant_builder_init(&actions_builder, G_VARIANT_TYPE("as"));
    
    if (actions) {
        for (int i = 0; actions[i] != NULL; i++) {
            g_variant_builder_add(&actions_builder, "s", actions[i]);
        }
    }
    
    // Build hints dictionary
    GVariantBuilder hints_builder;
    g_variant_builder_init(&hints_builder, G_VARIANT_TYPE("a{sv}"));
    
    GError *error = NULL;
    GVariant *result = g_dbus_proxy_call_sync(
        dbus_client->notifications_proxy,
        "Notify",
        g_variant_new("(susssasa{sv}i)",
                      app_name ? app_name : "TaishangApp",
                      0, // replaces_id
                      icon ? icon : "",
                      summary ? summary : "",
                      body ? body : "",
                      &actions_builder,
                      &hints_builder,
                      timeout),
        G_DBUS_CALL_FLAGS_NONE,
        -1,
        NULL,
        &error
    );
    
    guint32 notification_id = 0;
    if (result) {
        g_variant_get(result, "(u)", &notification_id);
        g_variant_unref(result);
    } else if (error) {
        g_warning("Failed to send notification: %s", error->message);
        g_error_free(error);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    
    g_print("Notification sent: %s -> ID %u\n", summary ? summary : "", notification_id);
    return notification_id;
}

gboolean taishang_dbus_close_notification(guint32 notification_id) {
    if (!dbus_client || !dbus_client->notifications_proxy) {
        return FALSE;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    GError *error = NULL;
    GVariant *result = g_dbus_proxy_call_sync(
        dbus_client->notifications_proxy,
        "CloseNotification",
        g_variant_new("(u)", notification_id),
        G_DBUS_CALL_FLAGS_NONE,
        -1,
        NULL,
        &error
    );
    
    gboolean success = (result != NULL);
    if (result) {
        g_variant_unref(result);
    } else if (error) {
        g_warning("Failed to close notification: %s", error->message);
        g_error_free(error);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    
    g_print("Notification closed: ID %u -> %s\n", notification_id, success ? "SUCCESS" : "FAILED");
    return success;
}

gboolean taishang_dbus_get_server_information(char **name, char **vendor, 
                                              char **version, char **spec_version) {
    if (!dbus_client || !dbus_client->notifications_proxy) {
        return FALSE;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    GError *error = NULL;
    GVariant *result = g_dbus_proxy_call_sync(
        dbus_client->notifications_proxy,
        "GetServerInformation",
        NULL,
        G_DBUS_CALL_FLAGS_NONE,
        -1,
        NULL,
        &error
    );
    
    gboolean success = FALSE;
    if (result) {
        const char *n, *v, *ver, *spec;
        g_variant_get(result, "(&s&s&s&s)", &n, &v, &ver, &spec);
        
        if (name) *name = g_strdup(n);
        if (vendor) *vendor = g_strdup(v);
        if (version) *version = g_strdup(ver);
        if (spec_version) *spec_version = g_strdup(spec);
        
        success = TRUE;
        g_variant_unref(result);
    } else if (error) {
        g_warning("Failed to get server information: %s", error->message);
        g_error_free(error);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    
    return success;
}

// Screensaver functions
gboolean taishang_dbus_inhibit_screensaver(const char *app_name, const char *reason, guint32 *cookie) {
    if (!dbus_client || !dbus_client->screensaver_proxy) {
        return FALSE;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    GError *error = NULL;
    GVariant *result = g_dbus_proxy_call_sync(
        dbus_client->screensaver_proxy,
        "Inhibit",
        g_variant_new("(ss)",
                      app_name ? app_name : "TaishangApp",
                      reason ? reason : "Application activity"),
        G_DBUS_CALL_FLAGS_NONE,
        -1,
        NULL,
        &error
    );
    
    gboolean success = FALSE;
    if (result) {
        guint32 c;
        g_variant_get(result, "(u)", &c);
        if (cookie) *cookie = c;
        success = TRUE;
        g_variant_unref(result);
    } else if (error) {
        g_warning("Failed to inhibit screensaver: %s", error->message);
        g_error_free(error);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    
    g_print("Screensaver inhibited: %s -> %s\n", reason ? reason : "", success ? "SUCCESS" : "FAILED");
    return success;
}

gboolean taishang_dbus_uninhibit_screensaver(guint32 cookie) {
    if (!dbus_client || !dbus_client->screensaver_proxy) {
        return FALSE;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    GError *error = NULL;
    GVariant *result = g_dbus_proxy_call_sync(
        dbus_client->screensaver_proxy,
        "UnInhibit",
        g_variant_new("(u)", cookie),
        G_DBUS_CALL_FLAGS_NONE,
        -1,
        NULL,
        &error
    );
    
    gboolean success = (result != NULL);
    if (result) {
        g_variant_unref(result);
    } else if (error) {
        g_warning("Failed to uninhibit screensaver: %s", error->message);
        g_error_free(error);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    
    g_print("Screensaver uninhibited: cookie %u -> %s\n", cookie, success ? "SUCCESS" : "FAILED");
    return success;
}

// Power management functions
gboolean taishang_dbus_get_power_info(TaishangPowerInfo *info) {
    if (!dbus_client || !dbus_client->power_proxy || !info) {
        return FALSE;
    }
    
    g_mutex_lock(&dbus_client->mutex);
    
    // Get power properties
    GError *error = NULL;
    GVariant *result = g_dbus_proxy_call_sync(
        dbus_client->power_proxy,
        "org.freedesktop.DBus.Properties.GetAll",
        g_variant_new("(s)", POWER_INTERFACE),
        G_DBUS_CALL_FLAGS_NONE,
        -1,
        NULL,
        &error
    );
    
    gboolean success = FALSE;
    if (result) {
        GVariantIter iter;
        const char *key;
        GVariant *value;
        
        g_variant_get(result, "(a{sv})", &iter);
        while (g_variant_iter_loop(&iter, "{&sv}", &key, &value)) {
            if (g_strcmp0(key, "OnBattery") == 0) {
                info->on_battery = g_variant_get_boolean(value);
            } else if (g_strcmp0(key, "LidIsClosed") == 0) {
                info->lid_closed = g_variant_get_boolean(value);
            } else if (g_strcmp0(key, "LidIsPresent") == 0) {
                info->lid_present = g_variant_get_boolean(value);
            }
        }
        
        success = TRUE;
        g_variant_unref(result);
    } else if (error) {
        g_warning("Failed to get power info: %s", error->message);
        g_error_free(error);
    }
    
    g_mutex_unlock(&dbus_client->mutex);
    
    return success;
}

// Callback registration
void taishang_dbus_set_notification_callback(TaishangDBusNotificationCallback callback, void *user_data) {
    if (!dbus_client) return;
    
    g_mutex_lock(&dbus_client->mutex);
    dbus_client->notification_callback = callback;
    dbus_client->user_data = user_data;
    g_mutex_unlock(&dbus_client->mutex);
}

void taishang_dbus_set_screensaver_callback(TaishangDBusScreensaverCallback callback, void *user_data) {
    if (!dbus_client) return;
    
    g_mutex_lock(&dbus_client->mutex);
    dbus_client->screensaver_callback = callback;
    dbus_client->user_data = user_data;
    g_mutex_unlock(&dbus_client->mutex);
}

void taishang_dbus_set_power_callback(TaishangDBusPowerCallback callback, void *user_data) {
    if (!dbus_client) return;
    
    g_mutex_lock(&dbus_client->mutex);
    dbus_client->power_callback = callback;
    dbus_client->user_data = user_data;
    g_mutex_unlock(&dbus_client->mutex);
}

void taishang_dbus_set_network_callback(TaishangDBusNetworkCallback callback, void *user_data) {
    if (!dbus_client) return;
    
    g_mutex_lock(&dbus_client->mutex);
    dbus_client->network_callback = callback;
    dbus_client->user_data = user_data;
    g_mutex_unlock(&dbus_client->mutex);
}

// Signal handlers
static void on_notification_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                                   GVariant *parameters, gpointer user_data) {
    if (!dbus_client || !dbus_client->notification_callback) {
        return;
    }
    
    if (g_strcmp0(signal_name, "NotificationClosed") == 0) {
        guint32 id;
        guint32 reason;
        g_variant_get(parameters, "(uu)", &id, &reason);
        
        dbus_client->notification_callback(TAISHANG_DBUS_NOTIFICATION_CLOSED, id, reason, dbus_client->user_data);
    } else if (g_strcmp0(signal_name, "ActionInvoked") == 0) {
        guint32 id;
        const char *action_key;
        g_variant_get(parameters, "(u&s)", &id, &action_key);
        
        dbus_client->notification_callback(TAISHANG_DBUS_NOTIFICATION_ACTION, id, 0, dbus_client->user_data);
    }
}

static void on_screensaver_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                                  GVariant *parameters, gpointer user_data) {
    if (!dbus_client || !dbus_client->screensaver_callback) {
        return;
    }
    
    if (g_strcmp0(signal_name, "ActiveChanged") == 0) {
        gboolean active;
        g_variant_get(parameters, "(b)", &active);
        
        dbus_client->screensaver_callback(active ? TAISHANG_DBUS_SCREENSAVER_ACTIVE : TAISHANG_DBUS_SCREENSAVER_INACTIVE, 
                                          dbus_client->user_data);
    }
}

static void on_power_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                            GVariant *parameters, gpointer user_data) {
    if (!dbus_client || !dbus_client->power_callback) {
        return;
    }
    
    if (g_strcmp0(signal_name, "Changed") == 0) {
        dbus_client->power_callback(TAISHANG_DBUS_POWER_CHANGED, dbus_client->user_data);
    }
}

static void on_network_signal(GDBusProxy *proxy, gchar *sender_name, gchar *signal_name, 
                              GVariant *parameters, gpointer user_data) {
    if (!dbus_client || !dbus_client->network_callback) {
        return;
    }
    
    if (g_strcmp0(signal_name, "StateChanged") == 0) {
        guint32 state;
        g_variant_get(parameters, "(u)", &state);
        
        dbus_client->network_callback(TAISHANG_DBUS_NETWORK_STATE_CHANGED, state, dbus_client->user_data);
    }
}