/**
 * @file common.h
 * @brief TaishangLaojun Desktop Application Common Header
 * @author TaishangLaojun Development Team
 * @version 1.0.0
 * @date 2024
 * 
 * This file contains common includes, macros, and definitions
 * used throughout the TaishangLaojun desktop application on Linux.
 */

#ifndef TAISHANG_COMMON_H
#define TAISHANG_COMMON_H

/* Standard C library includes */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>
#include <stdarg.h>
#include <errno.h>
#include <time.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <signal.h>
#include <locale.h>

/* GLib and GObject includes */
#include <glib.h>
#include <glib/gi18n.h>
#include <glib/gstdio.h>
#include <gio/gio.h>
#include <gobject/gobject.h>

/* GTK includes */
#include <gtk/gtk.h>
#include <gdk/gdk.h>
#include <gdk/gdkkeysyms.h>

/* Additional libraries */
#include <json-c/json.h>
#include <openssl/ssl.h>
#include <openssl/crypto.h>
#include <sqlite3.h>
#include <curl/curl.h>
#include <libnotify/notify.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Application information */
#ifndef TAISHANG_VERSION
#define TAISHANG_VERSION "1.0.0"
#endif

#ifndef TAISHANG_DATADIR
#define TAISHANG_DATADIR "/usr/share"
#endif

#ifndef TAISHANG_LOCALEDIR
#define TAISHANG_LOCALEDIR "/usr/share/locale"
#endif

/* Compiler attributes */
#ifdef __GNUC__
#define TAISHANG_UNUSED __attribute__((unused))
#define TAISHANG_DEPRECATED __attribute__((deprecated))
#define TAISHANG_PRINTF(format_idx, arg_idx) __attribute__((format(printf, format_idx, arg_idx)))
#define TAISHANG_LIKELY(x) __builtin_expect(!!(x), 1)
#define TAISHANG_UNLIKELY(x) __builtin_expect(!!(x), 0)
#else
#define TAISHANG_UNUSED
#define TAISHANG_DEPRECATED
#define TAISHANG_PRINTF(format_idx, arg_idx)
#define TAISHANG_LIKELY(x) (x)
#define TAISHANG_UNLIKELY(x) (x)
#endif

/* Memory management macros */
#define TAISHANG_NEW(type) ((type*)g_malloc(sizeof(type)))
#define TAISHANG_NEW0(type) ((type*)g_malloc0(sizeof(type)))
#define TAISHANG_NEWV(type, count) ((type*)g_malloc(sizeof(type) * (count)))
#define TAISHANG_NEW0V(type, count) ((type*)g_malloc0(sizeof(type) * (count)))

#define TAISHANG_FREE(ptr) do { \
    if ((ptr) != NULL) { \
        g_free(ptr); \
        (ptr) = NULL; \
    } \
} while (0)

#define TAISHANG_UNREF(obj) do { \
    if (G_IS_OBJECT(obj)) { \
        g_object_unref(obj); \
        (obj) = NULL; \
    } \
} while (0)

/* String macros */
#define TAISHANG_STR_EMPTY(str) ((str) == NULL || *(str) == '\0')
#define TAISHANG_STR_NOT_EMPTY(str) ((str) != NULL && *(str) != '\0')
#define TAISHANG_STR_EQUAL(str1, str2) (g_strcmp0((str1), (str2)) == 0)
#define TAISHANG_STR_NOT_EQUAL(str1, str2) (g_strcmp0((str1), (str2)) != 0)

/* Array macros */
#define TAISHANG_ARRAY_SIZE(arr) (sizeof(arr) / sizeof((arr)[0]))
#define TAISHANG_ARRAY_LAST_INDEX(arr) (TAISHANG_ARRAY_SIZE(arr) - 1)

/* Math macros */
#define TAISHANG_MIN(a, b) ((a) < (b) ? (a) : (b))
#define TAISHANG_MAX(a, b) ((a) > (b) ? (a) : (b))
#define TAISHANG_CLAMP(x, min, max) (TAISHANG_MIN(TAISHANG_MAX((x), (min)), (max)))
#define TAISHANG_ABS(x) ((x) < 0 ? -(x) : (x))

/* Bit manipulation macros */
#define TAISHANG_BIT_SET(var, bit) ((var) |= (1 << (bit)))
#define TAISHANG_BIT_CLEAR(var, bit) ((var) &= ~(1 << (bit)))
#define TAISHANG_BIT_TOGGLE(var, bit) ((var) ^= (1 << (bit)))
#define TAISHANG_BIT_CHECK(var, bit) (((var) >> (bit)) & 1)

/* Debug macros */
#ifdef DEBUG
#define TAISHANG_DEBUG_PRINT(format, ...) \
    g_print("[DEBUG] %s:%d: " format "\n", __FILE__, __LINE__, ##__VA_ARGS__)
#define TAISHANG_DEBUG_ENTER() \
    g_print("[DEBUG] Entering %s\n", __FUNCTION__)
#define TAISHANG_DEBUG_LEAVE() \
    g_print("[DEBUG] Leaving %s\n", __FUNCTION__)
#else
#define TAISHANG_DEBUG_PRINT(format, ...) G_STMT_START { } G_STMT_END
#define TAISHANG_DEBUG_ENTER() G_STMT_START { } G_STMT_END
#define TAISHANG_DEBUG_LEAVE() G_STMT_START { } G_STMT_END
#endif

/* Error handling macros */
#define TAISHANG_RETURN_IF_FAIL(expr) do { \
    if (TAISHANG_UNLIKELY(!(expr))) { \
        g_return_if_fail(expr); \
        return; \
    } \
} while (0)

#define TAISHANG_RETURN_VAL_IF_FAIL(expr, val) do { \
    if (TAISHANG_UNLIKELY(!(expr))) { \
        g_return_val_if_fail(expr, val); \
        return (val); \
    } \
} while (0)

#define TAISHANG_WARN_IF_FAIL(expr) do { \
    if (TAISHANG_UNLIKELY(!(expr))) { \
        g_warning("Expression '%s' failed at %s:%d", #expr, __FILE__, __LINE__); \
    } \
} while (0)

/* Signal connection macros */
#define TAISHANG_CONNECT(instance, signal, callback, data) \
    g_signal_connect((instance), (signal), G_CALLBACK(callback), (data))

#define TAISHANG_CONNECT_SWAPPED(instance, signal, callback, data) \
    g_signal_connect_swapped((instance), (signal), G_CALLBACK(callback), (data))

#define TAISHANG_CONNECT_AFTER(instance, signal, callback, data) \
    g_signal_connect_after((instance), (signal), G_CALLBACK(callback), (data))

/* GObject property macros */
#define TAISHANG_PARAM_READABLE (G_PARAM_READABLE | G_PARAM_STATIC_STRINGS)
#define TAISHANG_PARAM_WRITABLE (G_PARAM_WRITABLE | G_PARAM_STATIC_STRINGS)
#define TAISHANG_PARAM_READWRITE (G_PARAM_READWRITE | G_PARAM_STATIC_STRINGS)
#define TAISHANG_PARAM_CONSTRUCT (G_PARAM_READWRITE | G_PARAM_CONSTRUCT | G_PARAM_STATIC_STRINGS)
#define TAISHANG_PARAM_CONSTRUCT_ONLY (G_PARAM_WRITABLE | G_PARAM_CONSTRUCT_ONLY | G_PARAM_STATIC_STRINGS)

/* Common constants */
#define TAISHANG_BUFFER_SIZE_SMALL      256
#define TAISHANG_BUFFER_SIZE_MEDIUM     1024
#define TAISHANG_BUFFER_SIZE_LARGE      4096
#define TAISHANG_BUFFER_SIZE_HUGE       16384

#define TAISHANG_TIMEOUT_SHORT          1000    /* 1 second */
#define TAISHANG_TIMEOUT_MEDIUM         5000    /* 5 seconds */
#define TAISHANG_TIMEOUT_LONG           30000   /* 30 seconds */

/* File permissions */
#define TAISHANG_FILE_MODE_READ         0644
#define TAISHANG_FILE_MODE_WRITE        0644
#define TAISHANG_FILE_MODE_EXECUTE      0755
#define TAISHANG_DIR_MODE_DEFAULT       0755

/* Network constants */
#define TAISHANG_DEFAULT_PORT           8080
#define TAISHANG_MAX_CONNECTIONS        100
#define TAISHANG_NETWORK_TIMEOUT        30

/* UI constants */
#define TAISHANG_UI_SPACING_SMALL       6
#define TAISHANG_UI_SPACING_MEDIUM      12
#define TAISHANG_UI_SPACING_LARGE       18
#define TAISHANG_UI_BORDER_WIDTH        1
#define TAISHANG_UI_MARGIN_DEFAULT      6

/* Color constants (as strings for CSS) */
#define TAISHANG_COLOR_PRIMARY          "#2196F3"
#define TAISHANG_COLOR_SECONDARY        "#FFC107"
#define TAISHANG_COLOR_SUCCESS          "#4CAF50"
#define TAISHANG_COLOR_WARNING          "#FF9800"
#define TAISHANG_COLOR_ERROR            "#F44336"
#define TAISHANG_COLOR_INFO             "#2196F3"

/* Common error domains */
#define TAISHANG_ERROR (taishang_error_quark())
GQuark taishang_error_quark(void);

typedef enum {
    TAISHANG_ERROR_NONE = 0,
    TAISHANG_ERROR_INVALID_ARGUMENT,
    TAISHANG_ERROR_FILE_NOT_FOUND,
    TAISHANG_ERROR_PERMISSION_DENIED,
    TAISHANG_ERROR_OUT_OF_MEMORY,
    TAISHANG_ERROR_NETWORK_ERROR,
    TAISHANG_ERROR_TIMEOUT,
    TAISHANG_ERROR_CANCELLED,
    TAISHANG_ERROR_NOT_IMPLEMENTED,
    TAISHANG_ERROR_UNKNOWN
} TaishangError;

/* Common callback types */
typedef void (*TaishangCallback)(gpointer user_data);
typedef gboolean (*TaishangBooleanCallback)(gpointer user_data);
typedef void (*TaishangErrorCallback)(GError *error, gpointer user_data);
typedef void (*TaishangProgressCallback)(gdouble progress, const gchar *message, gpointer user_data);

/* Utility functions */
const gchar *taishang_error_to_string(TaishangError error);
void taishang_init_logging(void);
void taishang_cleanup_logging(void);

/* Version information */
const gchar *taishang_get_version(void);
gint taishang_get_major_version(void);
gint taishang_get_minor_version(void);
gint taishang_get_micro_version(void);

/* Build information */
const gchar *taishang_get_build_date(void);
const gchar *taishang_get_build_time(void);
const gchar *taishang_get_compiler_info(void);

/* Runtime checks */
gboolean taishang_check_version(gint required_major, gint required_minor, gint required_micro);

/* Internationalization support */
void taishang_init_i18n(void);
const gchar *taishang_get_locale(void);
void taishang_set_locale(const gchar *locale);

/* Platform detection */
#ifdef __linux__
#define TAISHANG_PLATFORM_LINUX 1
#define TAISHANG_PLATFORM_NAME "Linux"
#elif defined(__APPLE__)
#define TAISHANG_PLATFORM_MACOS 1
#define TAISHANG_PLATFORM_NAME "macOS"
#elif defined(_WIN32)
#define TAISHANG_PLATFORM_WINDOWS 1
#define TAISHANG_PLATFORM_NAME "Windows"
#else
#define TAISHANG_PLATFORM_UNKNOWN 1
#define TAISHANG_PLATFORM_NAME "Unknown"
#endif

/* Architecture detection */
#ifdef __x86_64__
#define TAISHANG_ARCH_X86_64 1
#define TAISHANG_ARCH_NAME "x86_64"
#elif defined(__i386__)
#define TAISHANG_ARCH_I386 1
#define TAISHANG_ARCH_NAME "i386"
#elif defined(__aarch64__)
#define TAISHANG_ARCH_ARM64 1
#define TAISHANG_ARCH_NAME "arm64"
#elif defined(__arm__)
#define TAISHANG_ARCH_ARM 1
#define TAISHANG_ARCH_NAME "arm"
#else
#define TAISHANG_ARCH_UNKNOWN 1
#define TAISHANG_ARCH_NAME "unknown"
#endif

/* Feature detection */
#ifdef HAVE_LIBNOTIFY
#define TAISHANG_FEATURE_NOTIFICATIONS 1
#endif

#ifdef HAVE_OPENSSL
#define TAISHANG_FEATURE_SSL 1
#endif

#ifdef HAVE_SQLITE3
#define TAISHANG_FEATURE_DATABASE 1
#endif

#ifdef HAVE_CURL
#define TAISHANG_FEATURE_HTTP 1
#endif

/* Cleanup macros for auto-cleanup */
#define TAISHANG_AUTO_FREE __attribute__((cleanup(taishang_auto_free)))
#define TAISHANG_AUTO_UNREF __attribute__((cleanup(taishang_auto_unref)))

static inline void taishang_auto_free(void *p) {
    void **pp = (void**)p;
    if (*pp) {
        g_free(*pp);
    }
}

static inline void taishang_auto_unref(void *p) {
    GObject **pp = (GObject**)p;
    if (*pp && G_IS_OBJECT(*pp)) {
        g_object_unref(*pp);
    }
}

#ifdef __cplusplus
}
#endif

#endif /* TAISHANG_COMMON_H */