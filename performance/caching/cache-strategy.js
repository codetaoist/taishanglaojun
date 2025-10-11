/**
 * Advanced Caching Strategy for Taishang Laojun AI Platform
 * 
 * This module implements a multi-layered caching strategy including:
 * - Browser cache optimization
 * - Service Worker caching
 * - Memory cache for API responses
 * - Regional cache distribution
 * - Cache invalidation strategies
 * - Performance-aware cache policies
 */

class CacheStrategy {
    constructor(config = {}) {
        this.config = {
            region: config.region || 'us-east-1',
            enableServiceWorker: config.enableServiceWorker !== false,
            enableMemoryCache: config.enableMemoryCache !== false,
            enableIndexedDB: config.enableIndexedDB !== false,
            maxMemoryCacheSize: config.maxMemoryCacheSize || 50 * 1024 * 1024, // 50MB
            maxIndexedDBSize: config.maxIndexedDBSize || 100 * 1024 * 1024, // 100MB
            defaultTTL: config.defaultTTL || 3600000, // 1 hour
            apiCacheTTL: config.apiCacheTTL || 300000, // 5 minutes
            staticCacheTTL: config.staticCacheTTL || 86400000, // 24 hours
            imageCacheTTL: config.imageCacheTTL || 604800000, // 7 days
            ...config
        };

        this.memoryCache = new Map();
        this.memoryCacheSize = 0;
        this.cacheStats = {
            hits: 0,
            misses: 0,
            evictions: 0,
            errors: 0
        };

        this.init();
    }

    /**
     * Initialize caching system
     */
    async init() {
        try {
            // Initialize Service Worker
            if (this.config.enableServiceWorker && 'serviceWorker' in navigator) {
                await this.initServiceWorker();
            }

            // Initialize IndexedDB
            if (this.config.enableIndexedDB) {
                await this.initIndexedDB();
            }

            // Set up cache headers optimization
            this.setupCacheHeaders();

            // Set up cache warming
            this.setupCacheWarming();

            // Set up cache cleanup
            this.setupCacheCleanup();

            console.log('🗄️ Cache strategy initialized');

        } catch (error) {
            console.error('❌ Failed to initialize cache strategy:', error);
        }
    }

    /**
     * Initialize Service Worker for advanced caching
     */
    async initServiceWorker() {
        try {
            const registration = await navigator.serviceWorker.register('/sw.js', {
                scope: '/'
            });

            // Send cache configuration to service worker
            if (registration.active) {
                registration.active.postMessage({
                    type: 'CACHE_CONFIG',
                    config: this.config
                });
            }

            // Listen for cache events from service worker
            navigator.serviceWorker.addEventListener('message', (event) => {
                if (event.data.type === 'CACHE_STATS') {
                    this.updateCacheStats(event.data.stats);
                }
            });

            console.log('🔧 Service Worker registered for caching');

        } catch (error) {
            console.error('Failed to register Service Worker:', error);
        }
    }

    /**
     * Initialize IndexedDB for large data caching
     */
    async initIndexedDB() {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open('TaishangLaojunCache', 1);

            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
                this.db = request.result;
                resolve();
            };

            request.onupgradeneeded = (event) => {
                const db = event.target.result;

                // Create object stores
                if (!db.objectStoreNames.contains('api_cache')) {
                    const apiStore = db.createObjectStore('api_cache', { keyPath: 'key' });
                    apiStore.createIndex('timestamp', 'timestamp');
                    apiStore.createIndex('region', 'region');
                }

                if (!db.objectStoreNames.contains('static_cache')) {
                    const staticStore = db.createObjectStore('static_cache', { keyPath: 'key' });
                    staticStore.createIndex('timestamp', 'timestamp');
                    staticStore.createIndex('type', 'type');
                }

                if (!db.objectStoreNames.contains('user_cache')) {
                    const userStore = db.createObjectStore('user_cache', { keyPath: 'key' });
                    userStore.createIndex('userId', 'userId');
                    userStore.createIndex('timestamp', 'timestamp');
                }
            };
        });
    }

    /**
     * Set up optimal cache headers
     */
    setupCacheHeaders() {
        // Intercept fetch requests to add cache headers
        const originalFetch = window.fetch;
        
        window.fetch = async (input, init = {}) => {
            const url = typeof input === 'string' ? input : input.url;
            const headers = new Headers(init.headers);

            // Add cache-control headers based on resource type
            if (this.isStaticResource(url)) {
                headers.set('Cache-Control', `public, max-age=${this.config.staticCacheTTL / 1000}`);
            } else if (this.isImageResource(url)) {
                headers.set('Cache-Control', `public, max-age=${this.config.imageCacheTTL / 1000}`);
            } else if (this.isAPIRequest(url)) {
                headers.set('Cache-Control', `private, max-age=${this.config.apiCacheTTL / 1000}`);
            }

            // Add ETag support
            headers.set('If-None-Match', this.getETag(url));

            return originalFetch(input, { ...init, headers });
        };
    }

    /**
     * Set up cache warming for critical resources
     */
    setupCacheWarming() {
        // Warm cache with critical resources
        const criticalResources = [
            '/api/user/profile',
            '/api/localization/strings',
            '/api/compliance/settings',
            '/static/css/critical.css',
            '/static/js/vendor.js'
        ];

        // Warm cache on idle
        if ('requestIdleCallback' in window) {
            requestIdleCallback(() => {
                this.warmCache(criticalResources);
            });
        } else {
            setTimeout(() => {
                this.warmCache(criticalResources);
            }, 1000);
        }
    }

    /**
     * Set up periodic cache cleanup
     */
    setupCacheCleanup() {
        // Clean up expired cache entries every 5 minutes
        setInterval(() => {
            this.cleanupExpiredCache();
        }, 300000);

        // Clean up on memory pressure
        if ('memory' in performance) {
            setInterval(() => {
                const memory = performance.memory;
                const memoryUsage = memory.usedJSHeapSize / memory.jsHeapSizeLimit;
                
                if (memoryUsage > 0.8) { // 80% memory usage
                    this.aggressiveCleanup();
                }
            }, 60000);
        }
    }

    /**
     * Get data from cache with fallback strategy
     */
    async get(key, options = {}) {
        const {
            fallback,
            ttl = this.config.defaultTTL,
            region = this.config.region,
            useMemory = true,
            useIndexedDB = true,
            useServiceWorker = true
        } = options;

        try {
            // Try memory cache first (fastest)
            if (useMemory && this.config.enableMemoryCache) {
                const memoryResult = this.getFromMemoryCache(key);
                if (memoryResult && !this.isExpired(memoryResult, ttl)) {
                    this.cacheStats.hits++;
                    return memoryResult.data;
                }
            }

            // Try IndexedDB (medium speed)
            if (useIndexedDB && this.config.enableIndexedDB && this.db) {
                const dbResult = await this.getFromIndexedDB(key);
                if (dbResult && !this.isExpired(dbResult, ttl)) {
                    // Promote to memory cache
                    this.setInMemoryCache(key, dbResult.data, dbResult.timestamp);
                    this.cacheStats.hits++;
                    return dbResult.data;
                }
            }

            // Try Service Worker cache (slower but persistent)
            if (useServiceWorker && 'serviceWorker' in navigator) {
                const swResult = await this.getFromServiceWorker(key);
                if (swResult && !this.isExpired(swResult, ttl)) {
                    // Promote to higher-level caches
                    this.setInMemoryCache(key, swResult.data, swResult.timestamp);
                    if (this.db) {
                        await this.setInIndexedDB(key, swResult.data, swResult.timestamp);
                    }
                    this.cacheStats.hits++;
                    return swResult.data;
                }
            }

            // Cache miss - use fallback if provided
            this.cacheStats.misses++;
            
            if (fallback && typeof fallback === 'function') {
                const data = await fallback();
                await this.set(key, data, { ttl, region });
                return data;
            }

            return null;

        } catch (error) {
            console.error('Cache get error:', error);
            this.cacheStats.errors++;
            
            if (fallback && typeof fallback === 'function') {
                return await fallback();
            }
            
            return null;
        }
    }

    /**
     * Set data in cache with multi-layer strategy
     */
    async set(key, data, options = {}) {
        const {
            ttl = this.config.defaultTTL,
            region = this.config.region,
            priority = 'normal',
            useMemory = true,
            useIndexedDB = true,
            useServiceWorker = true
        } = options;

        const timestamp = Date.now();
        const cacheEntry = {
            key,
            data,
            timestamp,
            ttl,
            region,
            priority
        };

        try {
            // Set in memory cache (fastest access)
            if (useMemory && this.config.enableMemoryCache) {
                this.setInMemoryCache(key, data, timestamp, { ttl, priority });
            }

            // Set in IndexedDB (persistent, larger capacity)
            if (useIndexedDB && this.config.enableIndexedDB && this.db) {
                await this.setInIndexedDB(key, data, timestamp, { ttl, region });
            }

            // Set in Service Worker cache (network-aware)
            if (useServiceWorker && 'serviceWorker' in navigator) {
                await this.setInServiceWorker(key, data, timestamp, { ttl, region });
            }

        } catch (error) {
            console.error('Cache set error:', error);
            this.cacheStats.errors++;
        }
    }

    /**
     * Memory cache operations
     */
    getFromMemoryCache(key) {
        return this.memoryCache.get(key);
    }

    setInMemoryCache(key, data, timestamp, options = {}) {
        const entry = {
            data,
            timestamp,
            size: this.calculateSize(data),
            priority: options.priority || 'normal'
        };

        // Check memory limits
        if (this.memoryCacheSize + entry.size > this.config.maxMemoryCacheSize) {
            this.evictFromMemoryCache(entry.size);
        }

        this.memoryCache.set(key, entry);
        this.memoryCacheSize += entry.size;
    }

    evictFromMemoryCache(requiredSize) {
        // LRU eviction with priority consideration
        const entries = Array.from(this.memoryCache.entries())
            .sort((a, b) => {
                // Sort by priority (low first) then by timestamp (old first)
                const priorityOrder = { low: 0, normal: 1, high: 2 };
                const priorityDiff = priorityOrder[a[1].priority] - priorityOrder[b[1].priority];
                
                if (priorityDiff !== 0) return priorityDiff;
                return a[1].timestamp - b[1].timestamp;
            });

        let freedSize = 0;
        for (const [key, entry] of entries) {
            this.memoryCache.delete(key);
            this.memoryCacheSize -= entry.size;
            freedSize += entry.size;
            this.cacheStats.evictions++;

            if (freedSize >= requiredSize) break;
        }
    }

    /**
     * IndexedDB cache operations
     */
    async getFromIndexedDB(key) {
        if (!this.db) return null;

        return new Promise((resolve, reject) => {
            const transaction = this.db.transaction(['api_cache'], 'readonly');
            const store = transaction.objectStore('api_cache');
            const request = store.get(key);

            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async setInIndexedDB(key, data, timestamp, options = {}) {
        if (!this.db) return;

        return new Promise((resolve, reject) => {
            const transaction = this.db.transaction(['api_cache'], 'readwrite');
            const store = transaction.objectStore('api_cache');
            
            const entry = {
                key,
                data,
                timestamp,
                ttl: options.ttl,
                region: options.region,
                size: this.calculateSize(data)
            };

            const request = store.put(entry);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    /**
     * Service Worker cache operations
     */
    async getFromServiceWorker(key) {
        if (!('serviceWorker' in navigator) || !navigator.serviceWorker.controller) {
            return null;
        }

        return new Promise((resolve) => {
            const messageChannel = new MessageChannel();
            
            messageChannel.port1.onmessage = (event) => {
                resolve(event.data.result);
            };

            navigator.serviceWorker.controller.postMessage({
                type: 'CACHE_GET',
                key
            }, [messageChannel.port2]);
        });
    }

    async setInServiceWorker(key, data, timestamp, options = {}) {
        if (!('serviceWorker' in navigator) || !navigator.serviceWorker.controller) {
            return;
        }

        navigator.serviceWorker.controller.postMessage({
            type: 'CACHE_SET',
            key,
            data,
            timestamp,
            options
        });
    }

    /**
     * Cache warming
     */
    async warmCache(resources) {
        console.log('🔥 Warming cache with critical resources');
        
        const promises = resources.map(async (resource) => {
            try {
                if (resource.startsWith('/api/')) {
                    // Warm API cache
                    const response = await fetch(resource);
                    const data = await response.json();
                    await this.set(resource, data, { priority: 'high' });
                } else {
                    // Warm static resource cache
                    await fetch(resource);
                }
            } catch (error) {
                console.warn(`Failed to warm cache for ${resource}:`, error);
            }
        });

        await Promise.allSettled(promises);
    }

    /**
     * Cache cleanup operations
     */
    async cleanupExpiredCache() {
        const now = Date.now();

        // Clean memory cache
        for (const [key, entry] of this.memoryCache.entries()) {
            if (this.isExpired(entry, entry.ttl || this.config.defaultTTL)) {
                this.memoryCache.delete(key);
                this.memoryCacheSize -= entry.size;
            }
        }

        // Clean IndexedDB cache
        if (this.db) {
            await this.cleanupIndexedDBExpired(now);
        }
    }

    async cleanupIndexedDBExpired(now) {
        return new Promise((resolve, reject) => {
            const transaction = this.db.transaction(['api_cache'], 'readwrite');
            const store = transaction.objectStore('api_cache');
            const index = store.index('timestamp');
            
            const request = index.openCursor();
            request.onsuccess = (event) => {
                const cursor = event.target.result;
                if (cursor) {
                    const entry = cursor.value;
                    if (this.isExpired(entry, entry.ttl || this.config.defaultTTL)) {
                        cursor.delete();
                    }
                    cursor.continue();
                } else {
                    resolve();
                }
            };
            
            request.onerror = () => reject(request.error);
        });
    }

    async aggressiveCleanup() {
        console.log('🧹 Performing aggressive cache cleanup due to memory pressure');
        
        // Clear low priority items from memory cache
        for (const [key, entry] of this.memoryCache.entries()) {
            if (entry.priority === 'low') {
                this.memoryCache.delete(key);
                this.memoryCacheSize -= entry.size;
                this.cacheStats.evictions++;
            }
        }

        // Clear old entries even if not expired
        const cutoffTime = Date.now() - (this.config.defaultTTL / 2);
        for (const [key, entry] of this.memoryCache.entries()) {
            if (entry.timestamp < cutoffTime && entry.priority !== 'high') {
                this.memoryCache.delete(key);
                this.memoryCacheSize -= entry.size;
                this.cacheStats.evictions++;
            }
        }
    }

    /**
     * Utility methods
     */
    isExpired(entry, ttl) {
        return Date.now() - entry.timestamp > ttl;
    }

    isStaticResource(url) {
        return /\.(js|css|woff|woff2|ttf|otf)$/i.test(url);
    }

    isImageResource(url) {
        return /\.(png|jpg|jpeg|gif|svg|webp|avif)$/i.test(url);
    }

    isAPIRequest(url) {
        return url.includes('/api/');
    }

    calculateSize(data) {
        return new Blob([JSON.stringify(data)]).size;
    }

    getETag(url) {
        // Simple ETag generation based on URL and timestamp
        return btoa(url + Date.now()).substring(0, 16);
    }

    updateCacheStats(stats) {
        Object.assign(this.cacheStats, stats);
    }

    /**
     * Get cache statistics
     */
    getStats() {
        const hitRate = this.cacheStats.hits / (this.cacheStats.hits + this.cacheStats.misses) * 100;
        
        return {
            ...this.cacheStats,
            hitRate: hitRate.toFixed(2) + '%',
            memoryCacheSize: this.memoryCacheSize,
            memoryCacheEntries: this.memoryCache.size,
            region: this.config.region
        };
    }

    /**
     * Clear all caches
     */
    async clearAll() {
        // Clear memory cache
        this.memoryCache.clear();
        this.memoryCacheSize = 0;

        // Clear IndexedDB
        if (this.db) {
            const transaction = this.db.transaction(['api_cache', 'static_cache', 'user_cache'], 'readwrite');
            await Promise.all([
                transaction.objectStore('api_cache').clear(),
                transaction.objectStore('static_cache').clear(),
                transaction.objectStore('user_cache').clear()
            ]);
        }

        // Clear Service Worker cache
        if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
            navigator.serviceWorker.controller.postMessage({
                type: 'CACHE_CLEAR_ALL'
            });
        }

        console.log('🗑️ All caches cleared');
    }
}

// Export for use in applications
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CacheStrategy;
} else if (typeof window !== 'undefined') {
    window.CacheStrategy = CacheStrategy;
}

// Auto-initialize if config is available
if (typeof window !== 'undefined' && window.CACHE_CONFIG) {
    const cacheStrategy = new CacheStrategy(window.CACHE_CONFIG);
    window.cacheStrategy = cacheStrategy;
}