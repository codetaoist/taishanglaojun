#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <glib.h>
#include <json-c/json.h>
#include "../../include/storage/cache.h"

// Cache entry structure
typedef struct {
    char *key;
    char *data;
    size_t size;
    gint64 timestamp;
    gint64 expiry;
    gint64 access_count;
    gint64 last_access;
} CacheEntry;

// Cache structure
typedef struct {
    GHashTable *entries;
    GMutex mutex;
    gint64 max_size;
    gint64 current_size;
    gint64 default_ttl;
    gboolean initialized;
    
    // Statistics
    gint64 hits;
    gint64 misses;
    gint64 evictions;
} TaishangCache;

static TaishangCache *cache = NULL;

// Forward declarations
static void cache_entry_free(CacheEntry *entry);
static gboolean cache_entry_is_expired(CacheEntry *entry);
static void cache_cleanup_expired(void);
static void cache_evict_lru(void);
static gboolean cache_cleanup_timer(gpointer user_data);

// Public functions
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

// Implementation
gboolean taishang_cache_init(gint64 max_size_mb, gint64 default_ttl_seconds) {
    if (cache != NULL) {
        g_warning("Cache already initialized");
        return FALSE;
    }
    
    cache = g_new0(TaishangCache, 1);
    cache->entries = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, (GDestroyNotify)cache_entry_free);
    g_mutex_init(&cache->mutex);
    
    cache->max_size = max_size_mb * 1024 * 1024; // Convert MB to bytes
    cache->current_size = 0;
    cache->default_ttl = default_ttl_seconds;
    cache->initialized = TRUE;
    
    // Reset statistics
    cache->hits = 0;
    cache->misses = 0;
    cache->evictions = 0;
    
    // Start cleanup timer (every 5 minutes)
    g_timeout_add_seconds(300, cache_cleanup_timer, NULL);
    
    g_print("Cache initialized: max_size=%ld MB, default_ttl=%ld seconds\n", 
            max_size_mb, default_ttl_seconds);
    return TRUE;
}

void taishang_cache_cleanup(void) {
    if (cache == NULL) {
        return;
    }
    
    g_mutex_lock(&cache->mutex);
    
    if (cache->entries) {
        g_hash_table_destroy(cache->entries);
    }
    
    g_mutex_unlock(&cache->mutex);
    g_mutex_clear(&cache->mutex);
    
    g_free(cache);
    cache = NULL;
    
    g_print("Cache cleaned up\n");
}

TaishangCache *taishang_cache_get_instance(void) {
    return cache;
}

gboolean taishang_cache_set(const char *key, const char *data, size_t size, gint64 ttl) {
    if (!cache || !key || !data) {
        return FALSE;
    }
    
    g_mutex_lock(&cache->mutex);
    
    // Check if we need to evict entries to make space
    while (cache->current_size + size > cache->max_size && g_hash_table_size(cache->entries) > 0) {
        cache_evict_lru();
    }
    
    // If still not enough space, fail
    if (cache->current_size + size > cache->max_size) {
        g_mutex_unlock(&cache->mutex);
        g_warning("Cache entry too large: %zu bytes", size);
        return FALSE;
    }
    
    // Remove existing entry if it exists
    CacheEntry *existing = g_hash_table_lookup(cache->entries, key);
    if (existing) {
        cache->current_size -= existing->size;
    }
    
    // Create new entry
    CacheEntry *entry = g_new0(CacheEntry, 1);
    entry->key = g_strdup(key);
    entry->data = g_memdup2(data, size);
    entry->size = size;
    entry->timestamp = g_get_real_time();
    entry->last_access = entry->timestamp;
    entry->access_count = 0;
    
    // Set expiry
    if (ttl > 0) {
        entry->expiry = entry->timestamp + (ttl * G_USEC_PER_SEC);
    } else if (cache->default_ttl > 0) {
        entry->expiry = entry->timestamp + (cache->default_ttl * G_USEC_PER_SEC);
    } else {
        entry->expiry = 0; // Never expires
    }
    
    // Insert into cache
    g_hash_table_insert(cache->entries, g_strdup(key), entry);
    cache->current_size += size;
    
    g_mutex_unlock(&cache->mutex);
    
    g_print("Cache set: %s (%zu bytes, ttl=%ld)\n", key, size, ttl);
    return TRUE;
}

char *taishang_cache_get(const char *key, size_t *size) {
    if (!cache || !key) {
        return NULL;
    }
    
    g_mutex_lock(&cache->mutex);
    
    CacheEntry *entry = g_hash_table_lookup(cache->entries, key);
    
    if (!entry) {
        cache->misses++;
        g_mutex_unlock(&cache->mutex);
        return NULL;
    }
    
    // Check if expired
    if (cache_entry_is_expired(entry)) {
        cache->misses++;
        g_hash_table_remove(cache->entries, key);
        g_mutex_unlock(&cache->mutex);
        return NULL;
    }
    
    // Update access statistics
    entry->access_count++;
    entry->last_access = g_get_real_time();
    cache->hits++;
    
    // Copy data
    char *data = g_memdup2(entry->data, entry->size);
    if (size) {
        *size = entry->size;
    }
    
    g_mutex_unlock(&cache->mutex);
    
    g_print("Cache hit: %s (%zu bytes)\n", key, entry->size);
    return data;
}

gboolean taishang_cache_exists(const char *key) {
    if (!cache || !key) {
        return FALSE;
    }
    
    g_mutex_lock(&cache->mutex);
    
    CacheEntry *entry = g_hash_table_lookup(cache->entries, key);
    gboolean exists = (entry != NULL && !cache_entry_is_expired(entry));
    
    g_mutex_unlock(&cache->mutex);
    
    return exists;
}

gboolean taishang_cache_delete(const char *key) {
    if (!cache || !key) {
        return FALSE;
    }
    
    g_mutex_lock(&cache->mutex);
    
    CacheEntry *entry = g_hash_table_lookup(cache->entries, key);
    if (entry) {
        cache->current_size -= entry->size;
        g_hash_table_remove(cache->entries, key);
        g_mutex_unlock(&cache->mutex);
        g_print("Cache delete: %s\n", key);
        return TRUE;
    }
    
    g_mutex_unlock(&cache->mutex);
    return FALSE;
}

void taishang_cache_clear(void) {
    if (!cache) {
        return;
    }
    
    g_mutex_lock(&cache->mutex);
    
    g_hash_table_remove_all(cache->entries);
    cache->current_size = 0;
    
    g_mutex_unlock(&cache->mutex);
    
    g_print("Cache cleared\n");
}

void taishang_cache_set_max_size(gint64 max_size_mb) {
    if (!cache) {
        return;
    }
    
    g_mutex_lock(&cache->mutex);
    cache->max_size = max_size_mb * 1024 * 1024;
    g_mutex_unlock(&cache->mutex);
    
    g_print("Cache max size set to: %ld MB\n", max_size_mb);
}

void taishang_cache_set_default_ttl(gint64 ttl_seconds) {
    if (!cache) {
        return;
    }
    
    g_mutex_lock(&cache->mutex);
    cache->default_ttl = ttl_seconds;
    g_mutex_unlock(&cache->mutex);
    
    g_print("Cache default TTL set to: %ld seconds\n", ttl_seconds);
}

gint64 taishang_cache_get_size(void) {
    if (!cache) {
        return 0;
    }
    
    g_mutex_lock(&cache->mutex);
    gint64 size = cache->current_size;
    g_mutex_unlock(&cache->mutex);
    
    return size;
}

gint64 taishang_cache_get_count(void) {
    if (!cache) {
        return 0;
    }
    
    g_mutex_lock(&cache->mutex);
    gint64 count = g_hash_table_size(cache->entries);
    g_mutex_unlock(&cache->mutex);
    
    return count;
}

TaishangCacheStats taishang_cache_get_stats(void) {
    TaishangCacheStats stats = {0};
    
    if (!cache) {
        return stats;
    }
    
    g_mutex_lock(&cache->mutex);
    
    stats.hits = cache->hits;
    stats.misses = cache->misses;
    stats.evictions = cache->evictions;
    stats.current_size = cache->current_size;
    stats.max_size = cache->max_size;
    stats.entry_count = g_hash_table_size(cache->entries);
    
    if (stats.hits + stats.misses > 0) {
        stats.hit_ratio = (double)stats.hits / (stats.hits + stats.misses);
    }
    
    g_mutex_unlock(&cache->mutex);
    
    return stats;
}

void taishang_cache_reset_stats(void) {
    if (!cache) {
        return;
    }
    
    g_mutex_lock(&cache->mutex);
    cache->hits = 0;
    cache->misses = 0;
    cache->evictions = 0;
    g_mutex_unlock(&cache->mutex);
    
    g_print("Cache statistics reset\n");
}

gboolean taishang_cache_set_json(const char *key, json_object *json_obj, gint64 ttl) {
    if (!json_obj) {
        return FALSE;
    }
    
    const char *json_string = json_object_to_json_string(json_obj);
    if (!json_string) {
        return FALSE;
    }
    
    return taishang_cache_set(key, json_string, strlen(json_string), ttl);
}

json_object *taishang_cache_get_json(const char *key) {
    size_t size;
    char *data = taishang_cache_get(key, &size);
    
    if (!data) {
        return NULL;
    }
    
    json_object *json_obj = json_tokener_parse(data);
    g_free(data);
    
    return json_obj;
}

// Private functions
static void cache_entry_free(CacheEntry *entry) {
    if (!entry) return;
    
    g_free(entry->key);
    g_free(entry->data);
    g_free(entry);
}

static gboolean cache_entry_is_expired(CacheEntry *entry) {
    if (!entry || entry->expiry == 0) {
        return FALSE;
    }
    
    return g_get_real_time() > entry->expiry;
}

static void cache_cleanup_expired(void) {
    if (!cache) return;
    
    GHashTableIter iter;
    gpointer key, value;
    GList *expired_keys = NULL;
    
    g_hash_table_iter_init(&iter, cache->entries);
    while (g_hash_table_iter_next(&iter, &key, &value)) {
        CacheEntry *entry = (CacheEntry *)value;
        if (cache_entry_is_expired(entry)) {
            expired_keys = g_list_prepend(expired_keys, g_strdup((char *)key));
        }
    }
    
    // Remove expired entries
    for (GList *l = expired_keys; l != NULL; l = l->next) {
        char *expired_key = (char *)l->data;
        CacheEntry *entry = g_hash_table_lookup(cache->entries, expired_key);
        if (entry) {
            cache->current_size -= entry->size;
        }
        g_hash_table_remove(cache->entries, expired_key);
        g_free(expired_key);
    }
    
    g_list_free(expired_keys);
}

static void cache_evict_lru(void) {
    if (!cache || g_hash_table_size(cache->entries) == 0) {
        return;
    }
    
    GHashTableIter iter;
    gpointer key, value;
    CacheEntry *lru_entry = NULL;
    char *lru_key = NULL;
    gint64 oldest_access = G_MAXINT64;
    
    g_hash_table_iter_init(&iter, cache->entries);
    while (g_hash_table_iter_next(&iter, &key, &value)) {
        CacheEntry *entry = (CacheEntry *)value;
        if (entry->last_access < oldest_access) {
            oldest_access = entry->last_access;
            lru_entry = entry;
            lru_key = (char *)key;
        }
    }
    
    if (lru_entry) {
        cache->current_size -= lru_entry->size;
        cache->evictions++;
        g_hash_table_remove(cache->entries, lru_key);
        g_print("Cache evicted LRU entry: %s\n", lru_key);
    }
}

static gboolean cache_cleanup_timer(gpointer user_data) {
    if (!cache) {
        return FALSE; // Stop timer
    }
    
    g_mutex_lock(&cache->mutex);
    cache_cleanup_expired();
    g_mutex_unlock(&cache->mutex);
    
    return TRUE; // Continue timer
}