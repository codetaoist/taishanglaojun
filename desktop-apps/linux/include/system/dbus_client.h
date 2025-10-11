#ifndef TAISHANG_DBUS_CLIENT_H
#define TAISHANG_DBUS_CLIENT_H

#include <glib.h>
#include <gio/gio.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangDBusClient TaishangDBusClient;

// Notification event types
typedef enum {
    TAISHANG_DBUS_NOTIFICATION_CLOSED,
    TAISHANG_DBUS_NOTIFICATION_ACTION
} TaishangDBusNotificationEvent;

// Screensaver event types
typedef enum {
    TAISHANG_DBUS_SCREENSAVER_ACTIVE,
    TAISHANG_DBUS_SCREENSAVER_INACTIVE
} TaishangDBusScreensaverEvent;

// Power event types
typedef enum {
    TAISHANG_DBUS_POWER_CHANGED,
    TAISHANG_DBUS_POWER_SUSPEND,
    TAISHANG_DBUS_POWER_RESUME
} TaishangDBusPowerEvent;

// Network event types
typedef enum {
    TAISHANG_DBUS_NETWORK_STATE_CHANGED,
    TAISHANG_DBUS_NETWORK_CONNECTED,
    TAISHANG_DBUS_NETWORK_DISCONNECTED
} TaishangDBusNetworkEvent;

// Network state
typedef enum {
    TAISHANG_NETWORK_STATE_UNKNOWN = 0,
    TAISHANG_NETWORK_STATE_ASLEEP = 10,
    TAISHANG_NETWORK_STATE_DISCONNECTED = 20,
    TAISHANG_NETWORK_STATE_DISCONNECTING = 30,
    TAISHANG_NETWORK_STATE_CONNECTING = 40,
    TAISHANG_NETWORK_STATE_CONNECTED_LOCAL = 50,
    TAISHANG_NETWORK_STATE_CONNECTED_SITE = 60,
    TAISHANG_NETWORK_STATE_CONNECTED_GLOBAL = 70
} TaishangNetworkState;

// Power information structure
typedef struct {
    gboolean on_battery;
    gboolean lid_closed;
    gboolean lid_present;
    gdouble battery_level;
    gchar *battery_state;
} TaishangPowerInfo;

// Network connection information
typedef struct {
    gchar *id;
    gchar *name;
    gchar *type;
    gchar *device;
    gboolean active;
    gchar *state;
} TaishangNetworkConnection;

// Callback function types
typedef void (*TaishangDBusNotificationCallback)(TaishangDBusNotificationEvent event, 
                                                 guint32 notification_id, 
                                                 guint32 reason, 
                                                 gpointer user_data);

typedef void (*TaishangDBusScreensaverCallback)(TaishangDBusScreensaverEvent event, 
                                                gpointer user_data);

typedef void (*TaishangDBusPowerCallback)(TaishangDBusPowerEvent event, 
                                          gpointer user_data);

typedef void (*TaishangDBusNetworkCallback)(TaishangDBusNetworkEvent event, 
                                            guint32 state, 
                                            gpointer user_data);

// Initialization and cleanup
gboolean taishang_dbus_client_init(void);
void taishang_dbus_client_cleanup(void);
TaishangDBusClient *taishang_dbus_client_get_instance(void);

// Notification functions
guint32 taishang_dbus_send_notification(const char *app_name, 
                                        const char *summary, 
                                        const char *body, 
                                        const char *icon, 
                                        gint32 timeout, 
                                        const char **actions);

gboolean taishang_dbus_close_notification(guint32 notification_id);

gboolean taishang_dbus_get_server_information(char **name, 
                                              char **vendor, 
                                              char **version, 
                                              char **spec_version);

// Screensaver functions
gboolean taishang_dbus_inhibit_screensaver(const char *app_name, 
                                           const char *reason, 
                                           guint32 *cookie);

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
void taishang_dbus_set_notification_callback(TaishangDBusNotificationCallback callback, 
                                             void *user_data);

void taishang_dbus_set_screensaver_callback(TaishangDBusScreensaverCallback callback, 
                                            void *user_data);

void taishang_dbus_set_power_callback(TaishangDBusPowerCallback callback, 
                                      void *user_data);

void taishang_dbus_set_network_callback(TaishangDBusNetworkCallback callback, 
                                        void *user_data);

// Utility functions
void taishang_power_info_free(TaishangPowerInfo *info);
void taishang_network_connection_free(TaishangNetworkConnection *connection);
void taishang_network_connections_free(GList *connections);

// Convenience macros
#define TAISHANG_NOTIFICATION_TIMEOUT_DEFAULT -1
#define TAISHANG_NOTIFICATION_TIMEOUT_NEVER 0

// Common notification icons
#define TAISHANG_ICON_INFO "dialog-information"
#define TAISHANG_ICON_WARNING "dialog-warning"
#define TAISHANG_ICON_ERROR "dialog-error"
#define TAISHANG_ICON_QUESTION "dialog-question"
#define TAISHANG_ICON_MESSAGE "mail-message-new"
#define TAISHANG_ICON_NETWORK "network-wireless"
#define TAISHANG_ICON_BATTERY "battery"

G_END_DECLS

#endif // TAISHANG_DBUS_CLIENT_H