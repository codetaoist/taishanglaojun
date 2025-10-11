#ifndef TAISHANG_CACHE_H
#define TAISHANG_CACHE_H

#include <glib.h>
#include <json-c/json.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangCache TaishangCache;

// Cache statistics structure
typedef struct {
    gint64 hits;
    gint64 misses;
    gint64 evictions;
    gint64 current_size;
    gint64 max_size;
    gint64 entry_count;
    double hit_ratio;
} TaishangCacheStats;

// Initialization and cleanup
gboolean taishang_cache_init(gint64 max_size_mb, gint64 default_ttl_seconds);
void taishang_cache_cleanup(void);
TaishangCache *taishang_cache_get_instance(void);

// Cache operations
gboolean taishang_cache_set(const char *key, const char *data, size_t size, gint64 ttl);
char *taishang_cache_get(const char *key, size_t *size);
gboolean taishang_cache_exists(const char *key);
gboolean taishang_cache_delete(const char *key);
void taishang_cache_clear(void);

// Cache management
void taishang_cache_set_max_size(gint64 max_size_mb);
void taishang_cache_set_default_ttl(gint64 ttl_seconds);
gint64 taishang_cache_get_size(void);
gint64 taishang_cache_get_count(void);

// Statistics
TaishangCacheStats taishang_cache_get_stats(void);
void taishang_cache_reset_stats(void);

// Utility functions
gboolean taishang_cache_set_json(const char *key, json_object *json_obj, gint64 ttl);
json_object *taishang_cache_get_json(const char *key);

// Cache key helpers
#define TAISHANG_CACHE_KEY_USER_PREFIX "user:"
#define TAISHANG_CACHE_KEY_MESSAGE_PREFIX "message:"
#define TAISHANG_CACHE_KEY_PROJECT_PREFIX "project:"
#define TAISHANG_CACHE_KEY_FILE_PREFIX "file:"
#define TAISHANG_CACHE_KEY_FRIEND_PREFIX "friend:"

// Default TTL values (in seconds)
#define TAISHANG_CACHE_TTL_SHORT 300      // 5 minutes
#define TAISHANG_CACHE_TTL_MEDIUM 1800    // 30 minutes
#define TAISHANG_CACHE_TTL_LONG 3600      // 1 hour
#define TAISHANG_CACHE_TTL_VERY_LONG 86400 // 24 hours
#define TAISHANG_CACHE_TTL_NEVER 0        // Never expires

G_END_DECLS

#endif // TAISHANG_CACHE_H