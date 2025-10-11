/**
 * Service Worker for Taishang Laojun AI Platform
 * 
 * Provides advanced caching strategies including:
 * - Network-first for API calls
 * - Cache-first for static assets
 * - Stale-while-revalidate for dynamic content
 * - Background sync for offline functionality
 * - Push notifications
 * - Regional cache optimization
 */

const CACHE_VERSION = 'v1.2.0';
const CACHE_PREFIX = 'taishang-laojun';

// Cache names
const CACHES = {
    STATIC: `${CACHE_PREFIX}-static-${CACHE_VERSION}`,
    DYNAMIC: `${CACHE_PREFIX}-dynamic-${CACHE_VERSION}`,
    API: `${CACHE_PREFIX}-api-${CACHE_VERSION}`,
    IMAGES: `${CACHE_PREFIX}-images-${CACHE_VERSION}`,
    FONTS: `${CACHE_PREFIX}-fonts-${CACHE_VERSION}`
};

// Cache strategies
const CACHE_STRATEGIES = {
    NETWORK_FIRST: 'network-first',
    CACHE_FIRST: 'cache-first',
    STALE_WHILE_REVALIDATE: 'stale-while-revalidate',
    NETWORK_ONLY: 'network-only',
    CACHE_ONLY: 'cache-only'
};

// Cache configuration
let cacheConfig = {
    region: 'us-east-1',
    staticCacheTTL: 86400000, // 24 hours
    apiCacheTTL: 300000, // 5 minutes
    imageCacheTTL: 604800000, // 7 days
    fontCacheTTL: 2592000000, // 30 days
    maxCacheSize: 100 * 1024 * 1024, // 100MB
    enableBackgroundSync: true,
    enablePushNotifications: true
};

// Cache statistics
let cacheStats = {
    hits: 0,
    misses: 0,
    networkRequests: 0,
    cacheSize: 0,
    lastCleanup: Date.now()
};

// Resource patterns and their cache strategies
const CACHE_PATTERNS = [
    {
        pattern: /\/api\/auth\//,
        strategy: CACHE_STRATEGIES.NETWORK_ONLY,
        cache: null
    },
    {
        pattern: /\/api\/user\/profile/,
        strategy: CACHE_STRATEGIES.STALE_WHILE_REVALIDATE,
        cache: CACHES.API,
        ttl: 300000 // 5 minutes
    },
    {
        pattern: /\/api\/localization\//,
        strategy: CACHE_STRATEGIES.STALE_WHILE_REVALIDATE,
        cache: CACHES.API,
        ttl: 3600000 // 1 hour
    },
    {
        pattern: /\/api\/compliance\//,
        strategy: CACHE_STRATEGIES.NETWORK_FIRST,
        cache: CACHES.API,
        ttl: 600000 // 10 minutes
    },
    {
        pattern: /\/api\//,
        strategy: CACHE_STRATEGIES.NETWORK_FIRST,
        cache: CACHES.API,
        ttl: 300000 // 5 minutes
    },
    {
        pattern: /\.(js|css)$/,
        strategy: CACHE_STRATEGIES.STALE_WHILE_REVALIDATE,
        cache: CACHES.STATIC,
        ttl: 86400000 // 24 hours
    },
    {
        pattern: /\.(png|jpg|jpeg|gif|svg|webp|avif)$/,
        strategy: CACHE_STRATEGIES.CACHE_FIRST,
        cache: CACHES.IMAGES,
        ttl: 604800000 // 7 days
    },
    {
        pattern: /\.(woff|woff2|ttf|otf)$/,
        strategy: CACHE_STRATEGIES.CACHE_FIRST,
        cache: CACHES.FONTS,
        ttl: 2592000000 // 30 days
    },
    {
        pattern: /\.(html|htm)$/,
        strategy: CACHE_STRATEGIES.NETWORK_FIRST,
        cache: CACHES.DYNAMIC,
        ttl: 3600000 // 1 hour
    }
];

// Critical resources to cache on install
const CRITICAL_RESOURCES = [
    '/',
    '/static/css/critical.css',
    '/static/js/vendor.js',
    '/static/js/app.js',
    '/static/fonts/inter-var.woff2',
    '/manifest.json'
];

/**
 * Service Worker Installation
 */
self.addEventListener('install', (event) => {
    console.log('🔧 Service Worker installing...');
    
    event.waitUntil(
        (async () => {
            try {
                // Cache critical resources
                const staticCache = await caches.open(CACHES.STATIC);
                await staticCache.addAll(CRITICAL_RESOURCES);
                
                console.log('✅ Critical resources cached');
                
                // Skip waiting to activate immediately
                self.skipWaiting();
                
            } catch (error) {
                console.error('❌ Failed to cache critical resources:', error);
            }
        })()
    );
});

/**
 * Service Worker Activation
 */
self.addEventListener('activate', (event) => {
    console.log('🚀 Service Worker activating...');
    
    event.waitUntil(
        (async () => {
            try {
                // Clean up old caches
                const cacheNames = await caches.keys();
                const oldCaches = cacheNames.filter(name => 
                    name.startsWith(CACHE_PREFIX) && 
                    !Object.values(CACHES).includes(name)
                );
                
                await Promise.all(
                    oldCaches.map(cacheName => caches.delete(cacheName))
                );
                
                if (oldCaches.length > 0) {
                    console.log(`🗑️ Cleaned up ${oldCaches.length} old caches`);
                }
                
                // Claim all clients
                await self.clients.claim();
                
                // Initialize cache size tracking
                await updateCacheSize();
                
                console.log('✅ Service Worker activated');
                
            } catch (error) {
                console.error('❌ Failed to activate Service Worker:', error);
            }
        })()
    );
});

/**
 * Fetch Event Handler
 */
self.addEventListener('fetch', (event) => {
    // Skip non-GET requests
    if (event.request.method !== 'GET') {
        return;
    }
    
    // Skip chrome-extension and other non-http requests
    if (!event.request.url.startsWith('http')) {
        return;
    }
    
    event.respondWith(handleRequest(event.request));
});

/**
 * Handle fetch requests with appropriate cache strategy
 */
async function handleRequest(request) {
    const url = new URL(request.url);
    const pattern = findMatchingPattern(url.pathname + url.search);
    
    if (!pattern) {
        // Default to network-first for unmatched requests
        return networkFirst(request, CACHES.DYNAMIC, cacheConfig.apiCacheTTL);
    }
    
    switch (pattern.strategy) {
        case CACHE_STRATEGIES.NETWORK_FIRST:
            return networkFirst(request, pattern.cache, pattern.ttl);
            
        case CACHE_STRATEGIES.CACHE_FIRST:
            return cacheFirst(request, pattern.cache, pattern.ttl);
            
        case CACHE_STRATEGIES.STALE_WHILE_REVALIDATE:
            return staleWhileRevalidate(request, pattern.cache, pattern.ttl);
            
        case CACHE_STRATEGIES.NETWORK_ONLY:
            return networkOnly(request);
            
        case CACHE_STRATEGIES.CACHE_ONLY:
            return cacheOnly(request, pattern.cache);
            
        default:
            return networkFirst(request, pattern.cache, pattern.ttl);
    }
}

/**
 * Network-first strategy
 */
async function networkFirst(request, cacheName, ttl) {
    try {
        const networkResponse = await fetch(request);
        cacheStats.networkRequests++;
        
        if (networkResponse.ok && cacheName) {
            // Clone response for caching
            const responseClone = networkResponse.clone();
            await cacheResponse(request, responseClone, cacheName, ttl);
        }
        
        return networkResponse;
        
    } catch (error) {
        // Network failed, try cache
        if (cacheName) {
            const cachedResponse = await getCachedResponse(request, cacheName);
            if (cachedResponse) {
                cacheStats.hits++;
                return cachedResponse;
            }
        }
        
        cacheStats.misses++;
        throw error;
    }
}

/**
 * Cache-first strategy
 */
async function cacheFirst(request, cacheName, ttl) {
    const cachedResponse = await getCachedResponse(request, cacheName);
    
    if (cachedResponse && !isExpired(cachedResponse, ttl)) {
        cacheStats.hits++;
        return cachedResponse;
    }
    
    try {
        const networkResponse = await fetch(request);
        cacheStats.networkRequests++;
        
        if (networkResponse.ok) {
            const responseClone = networkResponse.clone();
            await cacheResponse(request, responseClone, cacheName, ttl);
        }
        
        return networkResponse;
        
    } catch (error) {
        if (cachedResponse) {
            cacheStats.hits++;
            return cachedResponse;
        }
        
        cacheStats.misses++;
        throw error;
    }
}

/**
 * Stale-while-revalidate strategy
 */
async function staleWhileRevalidate(request, cacheName, ttl) {
    const cachedResponse = await getCachedResponse(request, cacheName);
    
    // Always try to update cache in background
    const networkPromise = fetch(request).then(async (networkResponse) => {
        if (networkResponse.ok) {
            const responseClone = networkResponse.clone();
            await cacheResponse(request, responseClone, cacheName, ttl);
        }
        return networkResponse;
    }).catch(() => {
        // Ignore network errors in background update
    });
    
    if (cachedResponse) {
        cacheStats.hits++;
        
        // If cache is fresh, return it immediately
        if (!isExpired(cachedResponse, ttl)) {
            return cachedResponse;
        }
        
        // If cache is stale, try to return fresh response or fallback to stale
        try {
            const networkResponse = await Promise.race([
                networkPromise,
                new Promise((_, reject) => setTimeout(() => reject(new Error('Timeout')), 1000))
            ]);
            
            if (networkResponse && networkResponse.ok) {
                cacheStats.networkRequests++;
                return networkResponse;
            }
        } catch (error) {
            // Return stale cache on network error or timeout
        }
        
        return cachedResponse;
    }
    
    // No cache, wait for network
    try {
        const networkResponse = await networkPromise;
        cacheStats.networkRequests++;
        return networkResponse;
    } catch (error) {
        cacheStats.misses++;
        throw error;
    }
}

/**
 * Network-only strategy
 */
async function networkOnly(request) {
    cacheStats.networkRequests++;
    return fetch(request);
}

/**
 * Cache-only strategy
 */
async function cacheOnly(request, cacheName) {
    const cachedResponse = await getCachedResponse(request, cacheName);
    
    if (cachedResponse) {
        cacheStats.hits++;
        return cachedResponse;
    }
    
    cacheStats.misses++;
    throw new Error('No cached response available');
}

/**
 * Cache a response with metadata
 */
async function cacheResponse(request, response, cacheName, ttl) {
    try {
        const cache = await caches.open(cacheName);
        
        // Add cache metadata headers
        const headers = new Headers(response.headers);
        headers.set('sw-cache-timestamp', Date.now().toString());
        headers.set('sw-cache-ttl', ttl.toString());
        headers.set('sw-cache-region', cacheConfig.region);
        
        const responseWithMetadata = new Response(response.body, {
            status: response.status,
            statusText: response.statusText,
            headers
        });
        
        await cache.put(request, responseWithMetadata);
        await updateCacheSize();
        
    } catch (error) {
        console.error('Failed to cache response:', error);
    }
}

/**
 * Get cached response
 */
async function getCachedResponse(request, cacheName) {
    try {
        const cache = await caches.open(cacheName);
        return await cache.match(request);
    } catch (error) {
        console.error('Failed to get cached response:', error);
        return null;
    }
}

/**
 * Check if cached response is expired
 */
function isExpired(response, ttl) {
    const cacheTimestamp = response.headers.get('sw-cache-timestamp');
    if (!cacheTimestamp) return true;
    
    const age = Date.now() - parseInt(cacheTimestamp);
    return age > ttl;
}

/**
 * Find matching cache pattern for URL
 */
function findMatchingPattern(url) {
    return CACHE_PATTERNS.find(pattern => pattern.pattern.test(url));
}

/**
 * Update cache size statistics
 */
async function updateCacheSize() {
    try {
        let totalSize = 0;
        
        for (const cacheName of Object.values(CACHES)) {
            const cache = await caches.open(cacheName);
            const requests = await cache.keys();
            
            for (const request of requests) {
                const response = await cache.match(request);
                if (response) {
                    const blob = await response.blob();
                    totalSize += blob.size;
                }
            }
        }
        
        cacheStats.cacheSize = totalSize;
        
        // Clean up if cache is too large
        if (totalSize > cacheConfig.maxCacheSize) {
            await cleanupOldCache();
        }
        
    } catch (error) {
        console.error('Failed to update cache size:', error);
    }
}

/**
 * Clean up old cache entries
 */
async function cleanupOldCache() {
    console.log('🧹 Cleaning up old cache entries...');
    
    try {
        const now = Date.now();
        let cleanedSize = 0;
        
        for (const cacheName of Object.values(CACHES)) {
            const cache = await caches.open(cacheName);
            const requests = await cache.keys();
            
            for (const request of requests) {
                const response = await cache.match(request);
                if (response) {
                    const cacheTimestamp = response.headers.get('sw-cache-timestamp');
                    const cacheTTL = response.headers.get('sw-cache-ttl');
                    
                    if (cacheTimestamp && cacheTTL) {
                        const age = now - parseInt(cacheTimestamp);
                        const ttl = parseInt(cacheTTL);
                        
                        if (age > ttl) {
                            const blob = await response.blob();
                            cleanedSize += blob.size;
                            await cache.delete(request);
                        }
                    }
                }
            }
        }
        
        cacheStats.lastCleanup = now;
        await updateCacheSize();
        
        console.log(`✅ Cleaned up ${cleanedSize} bytes from cache`);
        
    } catch (error) {
        console.error('Failed to cleanup cache:', error);
    }
}

/**
 * Message handler for communication with main thread
 */
self.addEventListener('message', async (event) => {
    const { type, data } = event.data;
    
    switch (type) {
        case 'CACHE_CONFIG':
            cacheConfig = { ...cacheConfig, ...data.config };
            break;
            
        case 'CACHE_GET':
            const cachedResponse = await getCachedResponse(
                new Request(data.key), 
                CACHES.API
            );
            event.ports[0].postMessage({
                result: cachedResponse ? {
                    data: await cachedResponse.json(),
                    timestamp: parseInt(cachedResponse.headers.get('sw-cache-timestamp'))
                } : null
            });
            break;
            
        case 'CACHE_SET':
            const response = new Response(JSON.stringify(data.data), {
                headers: {
                    'Content-Type': 'application/json',
                    'sw-cache-timestamp': data.timestamp.toString(),
                    'sw-cache-ttl': (data.options.ttl || cacheConfig.apiCacheTTL).toString()
                }
            });
            await cacheResponse(
                new Request(data.key),
                response,
                CACHES.API,
                data.options.ttl || cacheConfig.apiCacheTTL
            );
            break;
            
        case 'CACHE_CLEAR_ALL':
            for (const cacheName of Object.values(CACHES)) {
                await caches.delete(cacheName);
            }
            cacheStats = {
                hits: 0,
                misses: 0,
                networkRequests: 0,
                cacheSize: 0,
                lastCleanup: Date.now()
            };
            break;
            
        case 'GET_CACHE_STATS':
            event.ports[0].postMessage({ stats: cacheStats });
            break;
            
        case 'CLEANUP_CACHE':
            await cleanupOldCache();
            break;
    }
});

/**
 * Background sync for offline functionality
 */
self.addEventListener('sync', (event) => {
    if (event.tag === 'background-sync') {
        event.waitUntil(handleBackgroundSync());
    }
});

/**
 * Handle background sync
 */
async function handleBackgroundSync() {
    try {
        // Sync pending API requests
        const pendingRequests = await getPendingRequests();
        
        for (const request of pendingRequests) {
            try {
                await fetch(request.url, request.options);
                await removePendingRequest(request.id);
            } catch (error) {
                console.error('Failed to sync request:', error);
            }
        }
        
    } catch (error) {
        console.error('Background sync failed:', error);
    }
}

/**
 * Push notification handler
 */
self.addEventListener('push', (event) => {
    if (!event.data) return;
    
    const data = event.data.json();
    
    const options = {
        body: data.body,
        icon: '/static/icons/icon-192x192.png',
        badge: '/static/icons/badge-72x72.png',
        tag: data.tag || 'default',
        data: data.data || {},
        actions: data.actions || [],
        requireInteraction: data.requireInteraction || false
    };
    
    event.waitUntil(
        self.registration.showNotification(data.title, options)
    );
});

/**
 * Notification click handler
 */
self.addEventListener('notificationclick', (event) => {
    event.notification.close();
    
    const action = event.action;
    const data = event.notification.data;
    
    event.waitUntil(
        clients.openWindow(data.url || '/')
    );
});

/**
 * Periodic cache cleanup
 */
setInterval(async () => {
    const now = Date.now();
    const lastCleanup = cacheStats.lastCleanup;
    
    // Clean up every hour
    if (now - lastCleanup > 3600000) {
        await cleanupOldCache();
    }
}, 300000); // Check every 5 minutes

/**
 * Utility functions for background sync
 */
async function getPendingRequests() {
    // Implementation would depend on IndexedDB storage
    return [];
}

async function removePendingRequest(id) {
    // Implementation would depend on IndexedDB storage
}

console.log('🔧 Service Worker loaded and ready');