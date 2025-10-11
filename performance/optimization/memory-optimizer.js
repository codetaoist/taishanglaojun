/**
 * Memory Optimizer for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive memory optimization including:
 * - Memory leak detection and prevention
 * - Garbage collection optimization
 * - Memory usage monitoring
 * - Component cleanup management
 * - Event listener cleanup
 * - Cache management with LRU eviction
 * - Memory pressure handling
 * - Regional memory optimization
 */

class MemoryOptimizer {
    constructor(options = {}) {
        this.config = {
            // Memory thresholds (in MB)
            warningThreshold: options.warningThreshold || 100,
            criticalThreshold: options.criticalThreshold || 200,
            maxHeapSize: options.maxHeapSize || 500,
            
            // Monitoring settings
            monitoringInterval: options.monitoringInterval || 30000, // 30 seconds
            enableAutoCleanup: options.enableAutoCleanup !== false,
            enableGCOptimization: options.enableGCOptimization !== false,
            
            // Cache settings
            maxCacheSize: options.maxCacheSize || 50 * 1024 * 1024, // 50MB
            cacheEvictionRatio: options.cacheEvictionRatio || 0.2, // 20%
            
            // Regional settings
            region: options.region || 'us-east-1',
            enableRegionalOptimization: options.enableRegionalOptimization !== false,
            
            // Reporting
            enableReporting: options.enableReporting !== false,
            reportInterval: options.reportInterval || 300000, // 5 minutes
            apiEndpoint: options.apiEndpoint || '/api/performance/memory',
            
            ...options
        };

        // Memory tracking
        this.memoryStats = {
            peak: 0,
            current: 0,
            baseline: 0,
            leaks: [],
            cleanups: 0,
            gcCount: 0
        };

        // Component tracking
        this.componentRegistry = new Map();
        this.eventListenerRegistry = new WeakMap();
        this.timerRegistry = new Set();
        this.observerRegistry = new Set();

        // Cache management
        this.caches = new Map();
        this.cacheStats = {
            hits: 0,
            misses: 0,
            evictions: 0,
            totalSize: 0
        };

        // Memory pressure detection
        this.memoryPressure = {
            level: 'normal', // 'normal', 'moderate', 'critical'
            lastCheck: Date.now(),
            history: []
        };

        this.monitoringTimer = null;
        this.reportingTimer = null;
        this.gcTimer = null;

        this.init();
    }

    /**
     * Initialize memory optimizer
     */
    async init() {
        try {
            // Get baseline memory usage
            await this.measureBaseline();

            // Start memory monitoring
            this.startMemoryMonitoring();

            // Setup automatic cleanup
            if (this.config.enableAutoCleanup) {
                this.setupAutoCleanup();
            }

            // Setup GC optimization
            if (this.config.enableGCOptimization) {
                this.setupGCOptimization();
            }

            // Start periodic reporting
            if (this.config.enableReporting) {
                this.startPeriodicReporting();
            }

            // Setup page lifecycle handlers
            this.setupPageLifecycleHandlers();

            console.log('🧠 Memory optimizer initialized');

        } catch (error) {
            console.error('❌ Failed to initialize memory optimizer:', error);
        }
    }

    /**
     * Measure baseline memory usage
     */
    async measureBaseline() {
        const memory = await this.getMemoryUsage();
        this.memoryStats.baseline = memory.usedJSHeapSize;
        this.memoryStats.current = memory.usedJSHeapSize;
        
        console.log(`📊 Baseline memory usage: ${this.formatBytes(this.memoryStats.baseline)}`);
    }

    /**
     * Get current memory usage
     */
    async getMemoryUsage() {
        if ('memory' in performance) {
            return performance.memory;
        }

        // Fallback estimation
        return {
            usedJSHeapSize: this.estimateMemoryUsage(),
            totalJSHeapSize: this.estimateMemoryUsage() * 1.5,
            jsHeapSizeLimit: this.config.maxHeapSize * 1024 * 1024
        };
    }

    /**
     * Estimate memory usage (fallback)
     */
    estimateMemoryUsage() {
        // Rough estimation based on DOM nodes and registered components
        const domNodes = document.querySelectorAll('*').length;
        const components = this.componentRegistry.size;
        const cacheSize = this.getTotalCacheSize();
        
        return (domNodes * 1000) + (components * 50000) + cacheSize;
    }

    /**
     * Start memory monitoring
     */
    startMemoryMonitoring() {
        this.monitoringTimer = setInterval(async () => {
            await this.checkMemoryUsage();
        }, this.config.monitoringInterval);
    }

    /**
     * Check current memory usage and detect issues
     */
    async checkMemoryUsage() {
        try {
            const memory = await this.getMemoryUsage();
            const currentUsage = memory.usedJSHeapSize;
            const usageMB = currentUsage / (1024 * 1024);

            // Update stats
            this.memoryStats.current = currentUsage;
            if (currentUsage > this.memoryStats.peak) {
                this.memoryStats.peak = currentUsage;
            }

            // Update memory pressure
            this.updateMemoryPressure(usageMB);

            // Check thresholds
            if (usageMB > this.config.criticalThreshold) {
                await this.handleCriticalMemoryUsage(usageMB);
            } else if (usageMB > this.config.warningThreshold) {
                await this.handleWarningMemoryUsage(usageMB);
            }

            // Detect memory leaks
            this.detectMemoryLeaks(currentUsage);

            // Log memory stats in development
            if (process.env.NODE_ENV === 'development') {
                console.log(`🧠 Memory: ${this.formatBytes(currentUsage)} (${this.memoryPressure.level})`);
            }

        } catch (error) {
            console.error('Failed to check memory usage:', error);
        }
    }

    /**
     * Update memory pressure level
     */
    updateMemoryPressure(usageMB) {
        let level = 'normal';
        
        if (usageMB > this.config.criticalThreshold) {
            level = 'critical';
        } else if (usageMB > this.config.warningThreshold) {
            level = 'moderate';
        }

        this.memoryPressure.level = level;
        this.memoryPressure.lastCheck = Date.now();
        this.memoryPressure.history.push({
            timestamp: Date.now(),
            usage: usageMB,
            level
        });

        // Keep only last 100 entries
        if (this.memoryPressure.history.length > 100) {
            this.memoryPressure.history = this.memoryPressure.history.slice(-100);
        }
    }

    /**
     * Handle critical memory usage
     */
    async handleCriticalMemoryUsage(usageMB) {
        console.warn(`🚨 Critical memory usage: ${usageMB.toFixed(2)}MB`);

        // Aggressive cleanup
        await this.performAggressiveCleanup();

        // Force garbage collection if available
        if (window.gc) {
            window.gc();
            this.memoryStats.gcCount++;
        }

        // Clear non-essential caches
        this.clearNonEssentialCaches();

        // Notify application
        this.notifyMemoryPressure('critical', usageMB);
    }

    /**
     * Handle warning memory usage
     */
    async handleWarningMemoryUsage(usageMB) {
        console.warn(`⚠️ High memory usage: ${usageMB.toFixed(2)}MB`);

        // Moderate cleanup
        await this.performModerateCleanup();

        // Evict old cache entries
        this.evictOldCacheEntries();

        // Notify application
        this.notifyMemoryPressure('moderate', usageMB);
    }

    /**
     * Detect memory leaks
     */
    detectMemoryLeaks(currentUsage) {
        const growthRate = (currentUsage - this.memoryStats.baseline) / this.memoryStats.baseline;
        const timeSinceBaseline = Date.now() - this.memoryStats.baseline;
        const expectedGrowth = 0.1; // 10% growth is acceptable

        if (growthRate > expectedGrowth && timeSinceBaseline > 300000) { // 5 minutes
            const leak = {
                timestamp: Date.now(),
                usage: currentUsage,
                growthRate,
                components: this.componentRegistry.size,
                eventListeners: this.getEventListenerCount(),
                timers: this.timerRegistry.size
            };

            this.memoryStats.leaks.push(leak);
            console.warn('🔍 Potential memory leak detected:', leak);

            // Keep only last 10 leak detections
            if (this.memoryStats.leaks.length > 10) {
                this.memoryStats.leaks = this.memoryStats.leaks.slice(-10);
            }
        }
    }

    /**
     * Perform aggressive cleanup
     */
    async performAggressiveCleanup() {
        // Clean up unused components
        this.cleanupUnusedComponents();

        // Clear all non-critical caches
        this.clearNonCriticalCaches();

        // Clean up event listeners
        this.cleanupEventListeners();

        // Clear timers and intervals
        this.cleanupTimers();

        // Clean up observers
        this.cleanupObservers();

        // Force DOM cleanup
        this.performDOMCleanup();

        this.memoryStats.cleanups++;
    }

    /**
     * Perform moderate cleanup
     */
    async performModerateCleanup() {
        // Clean up old cache entries
        this.evictOldCacheEntries();

        // Clean up unused components (less aggressive)
        this.cleanupUnusedComponents(false);

        // Clean up some event listeners
        this.cleanupEventListeners(false);

        this.memoryStats.cleanups++;
    }

    /**
     * Register component for memory tracking
     */
    registerComponent(component, metadata = {}) {
        const id = this.generateComponentId(component);
        
        this.componentRegistry.set(id, {
            component,
            metadata,
            registeredAt: Date.now(),
            lastAccessed: Date.now(),
            accessCount: 0,
            ...metadata
        });

        return id;
    }

    /**
     * Unregister component
     */
    unregisterComponent(componentId) {
        const componentData = this.componentRegistry.get(componentId);
        
        if (componentData) {
            // Clean up component resources
            this.cleanupComponentResources(componentData);
            this.componentRegistry.delete(componentId);
        }
    }

    /**
     * Track component access
     */
    trackComponentAccess(componentId) {
        const componentData = this.componentRegistry.get(componentId);
        
        if (componentData) {
            componentData.lastAccessed = Date.now();
            componentData.accessCount++;
        }
    }

    /**
     * Register event listener for tracking
     */
    registerEventListener(element, event, listener, options = {}) {
        if (!this.eventListenerRegistry.has(element)) {
            this.eventListenerRegistry.set(element, new Map());
        }

        const listeners = this.eventListenerRegistry.get(element);
        const key = `${event}_${listener.toString()}`;
        
        listeners.set(key, {
            event,
            listener,
            options,
            registeredAt: Date.now()
        });

        // Add the actual event listener
        element.addEventListener(event, listener, options);
    }

    /**
     * Unregister event listener
     */
    unregisterEventListener(element, event, listener) {
        const listeners = this.eventListenerRegistry.get(element);
        
        if (listeners) {
            const key = `${event}_${listener.toString()}`;
            const listenerData = listeners.get(key);
            
            if (listenerData) {
                element.removeEventListener(event, listener, listenerData.options);
                listeners.delete(key);
                
                if (listeners.size === 0) {
                    this.eventListenerRegistry.delete(element);
                }
            }
        }
    }

    /**
     * Register timer for tracking
     */
    registerTimer(timerId, type = 'timeout') {
        this.timerRegistry.add({
            id: timerId,
            type,
            registeredAt: Date.now()
        });
    }

    /**
     * Unregister timer
     */
    unregisterTimer(timerId) {
        for (const timer of this.timerRegistry) {
            if (timer.id === timerId) {
                this.timerRegistry.delete(timer);
                break;
            }
        }
    }

    /**
     * Register observer for tracking
     */
    registerObserver(observer, type = 'unknown') {
        this.observerRegistry.add({
            observer,
            type,
            registeredAt: Date.now()
        });
    }

    /**
     * Unregister observer
     */
    unregisterObserver(observer) {
        for (const observerData of this.observerRegistry) {
            if (observerData.observer === observer) {
                this.observerRegistry.delete(observerData);
                break;
            }
        }
    }

    /**
     * Create memory-optimized cache
     */
    createCache(name, options = {}) {
        const cache = new LRUCache({
            maxSize: options.maxSize || 10 * 1024 * 1024, // 10MB
            maxAge: options.maxAge || 3600000, // 1 hour
            onEvict: (key, value) => {
                this.cacheStats.evictions++;
                this.cacheStats.totalSize -= this.estimateObjectSize(value);
            }
        });

        this.caches.set(name, cache);
        return cache;
    }

    /**
     * Get cache by name
     */
    getCache(name) {
        return this.caches.get(name);
    }

    /**
     * Clear cache
     */
    clearCache(name) {
        const cache = this.caches.get(name);
        if (cache) {
            cache.clear();
        }
    }

    /**
     * Clear all caches
     */
    clearAllCaches() {
        this.caches.forEach(cache => cache.clear());
        this.cacheStats.totalSize = 0;
    }

    /**
     * Clear non-essential caches
     */
    clearNonEssentialCaches() {
        const nonEssential = ['images', 'api-cache', 'ui-cache'];
        
        nonEssential.forEach(name => {
            this.clearCache(name);
        });
    }

    /**
     * Clear non-critical caches
     */
    clearNonCriticalCaches() {
        // Keep only authentication and core caches
        const critical = ['auth', 'core', 'security'];
        
        this.caches.forEach((cache, name) => {
            if (!critical.includes(name)) {
                cache.clear();
            }
        });
    }

    /**
     * Evict old cache entries
     */
    evictOldCacheEntries() {
        this.caches.forEach(cache => {
            if (cache.evictOld) {
                cache.evictOld();
            }
        });
    }

    /**
     * Clean up unused components
     */
    cleanupUnusedComponents(aggressive = true) {
        const now = Date.now();
        const threshold = aggressive ? 300000 : 600000; // 5 or 10 minutes
        
        for (const [id, componentData] of this.componentRegistry) {
            const timeSinceAccess = now - componentData.lastAccessed;
            
            if (timeSinceAccess > threshold && componentData.accessCount < 5) {
                this.unregisterComponent(id);
            }
        }
    }

    /**
     * Clean up event listeners
     */
    cleanupEventListeners(aggressive = true) {
        const now = Date.now();
        const threshold = aggressive ? 300000 : 600000; // 5 or 10 minutes
        
        for (const [element, listeners] of this.eventListenerRegistry) {
            // Check if element is still in DOM
            if (!document.contains(element)) {
                // Remove all listeners for this element
                for (const [key, listenerData] of listeners) {
                    element.removeEventListener(
                        listenerData.event, 
                        listenerData.listener, 
                        listenerData.options
                    );
                }
                this.eventListenerRegistry.delete(element);
                continue;
            }

            // Remove old listeners
            for (const [key, listenerData] of listeners) {
                const age = now - listenerData.registeredAt;
                if (age > threshold) {
                    element.removeEventListener(
                        listenerData.event, 
                        listenerData.listener, 
                        listenerData.options
                    );
                    listeners.delete(key);
                }
            }

            if (listeners.size === 0) {
                this.eventListenerRegistry.delete(element);
            }
        }
    }

    /**
     * Clean up timers
     */
    cleanupTimers() {
        for (const timer of this.timerRegistry) {
            if (timer.type === 'timeout') {
                clearTimeout(timer.id);
            } else if (timer.type === 'interval') {
                clearInterval(timer.id);
            }
        }
        this.timerRegistry.clear();
    }

    /**
     * Clean up observers
     */
    cleanupObservers() {
        for (const observerData of this.observerRegistry) {
            if (observerData.observer && observerData.observer.disconnect) {
                observerData.observer.disconnect();
            }
        }
        this.observerRegistry.clear();
    }

    /**
     * Perform DOM cleanup
     */
    performDOMCleanup() {
        // Remove detached DOM nodes
        const walker = document.createTreeWalker(
            document.body,
            NodeFilter.SHOW_ELEMENT,
            {
                acceptNode: (node) => {
                    // Check if node is detached or has memory leaks
                    return node.parentNode ? NodeFilter.FILTER_SKIP : NodeFilter.FILTER_ACCEPT;
                }
            }
        );

        const detachedNodes = [];
        let node;
        
        while (node = walker.nextNode()) {
            detachedNodes.push(node);
        }

        // Remove detached nodes
        detachedNodes.forEach(node => {
            if (node.parentNode) {
                node.parentNode.removeChild(node);
            }
        });
    }

    /**
     * Setup automatic cleanup
     */
    setupAutoCleanup() {
        // Cleanup on page visibility change
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.performModerateCleanup();
            }
        });

        // Cleanup on low memory (if supported)
        if ('memory' in navigator) {
            // This is a proposed API, may not be available
            try {
                navigator.memory.addEventListener('memorywarning', () => {
                    this.handleCriticalMemoryUsage(this.memoryStats.current / (1024 * 1024));
                });
            } catch (error) {
                // API not supported
            }
        }
    }

    /**
     * Setup GC optimization
     */
    setupGCOptimization() {
        // Schedule periodic GC hints
        this.gcTimer = setInterval(() => {
            if (this.memoryPressure.level !== 'normal') {
                this.suggestGarbageCollection();
            }
        }, 60000); // Every minute
    }

    /**
     * Suggest garbage collection
     */
    suggestGarbageCollection() {
        // Force GC if available (development only)
        if (window.gc && process.env.NODE_ENV === 'development') {
            window.gc();
            this.memoryStats.gcCount++;
        }

        // Create temporary objects to trigger GC
        const temp = new Array(1000).fill(null).map(() => ({}));
        temp.length = 0; // Clear reference
    }

    /**
     * Setup page lifecycle handlers
     */
    setupPageLifecycleHandlers() {
        // Cleanup on page unload
        window.addEventListener('beforeunload', () => {
            this.performAggressiveCleanup();
            this.generateReport(true);
        });

        // Cleanup on page freeze (if supported)
        if ('onfreeze' in document) {
            document.addEventListener('freeze', () => {
                this.performModerateCleanup();
            });
        }
    }

    /**
     * Start periodic reporting
     */
    startPeriodicReporting() {
        this.reportingTimer = setInterval(() => {
            this.generateReport();
        }, this.config.reportInterval);
    }

    /**
     * Generate memory report
     */
    generateReport(isUnload = false) {
        const report = {
            timestamp: Date.now(),
            region: this.config.region,
            memoryStats: {
                ...this.memoryStats,
                current: this.memoryStats.current / (1024 * 1024), // Convert to MB
                peak: this.memoryStats.peak / (1024 * 1024),
                baseline: this.memoryStats.baseline / (1024 * 1024)
            },
            memoryPressure: this.memoryPressure,
            componentStats: {
                registered: this.componentRegistry.size,
                eventListeners: this.getEventListenerCount(),
                timers: this.timerRegistry.size,
                observers: this.observerRegistry.size
            },
            cacheStats: {
                ...this.cacheStats,
                totalSize: this.cacheStats.totalSize / (1024 * 1024), // Convert to MB
                cacheCount: this.caches.size
            },
            recommendations: this.getOptimizationRecommendations()
        };

        // Send report to server
        if (this.config.enableReporting) {
            this.sendReport(report, isUnload);
        }

        return report;
    }

    /**
     * Get optimization recommendations
     */
    getOptimizationRecommendations() {
        const recommendations = [];

        // Memory usage recommendations
        if (this.memoryStats.current > this.config.warningThreshold * 1024 * 1024) {
            recommendations.push({
                type: 'memory-usage',
                priority: 'high',
                description: 'High memory usage detected',
                suggestion: 'Consider implementing more aggressive cleanup strategies'
            });
        }

        // Component cleanup recommendations
        if (this.componentRegistry.size > 100) {
            recommendations.push({
                type: 'component-cleanup',
                priority: 'medium',
                description: 'Large number of registered components',
                suggestion: 'Implement component lifecycle management'
            });
        }

        // Cache optimization recommendations
        if (this.cacheStats.totalSize > this.config.maxCacheSize) {
            recommendations.push({
                type: 'cache-optimization',
                priority: 'medium',
                description: 'Cache size exceeds limit',
                suggestion: 'Implement more aggressive cache eviction'
            });
        }

        // Memory leak recommendations
        if (this.memoryStats.leaks.length > 3) {
            recommendations.push({
                type: 'memory-leaks',
                priority: 'high',
                description: 'Multiple memory leaks detected',
                suggestion: 'Review component cleanup and event listener management'
            });
        }

        return recommendations;
    }

    /**
     * Send report to server
     */
    async sendReport(report, isUnload = false) {
        try {
            const method = isUnload && 'sendBeacon' in navigator ? 'beacon' : 'fetch';
            
            if (method === 'beacon') {
                navigator.sendBeacon(
                    this.config.apiEndpoint,
                    JSON.stringify(report)
                );
            } else {
                await fetch(this.config.apiEndpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(report)
                });
            }

        } catch (error) {
            console.error('Failed to send memory report:', error);
        }
    }

    /**
     * Notify application of memory pressure
     */
    notifyMemoryPressure(level, usage) {
        const event = new CustomEvent('memoryPressure', {
            detail: { level, usage, timestamp: Date.now() }
        });
        
        window.dispatchEvent(event);
    }

    /**
     * Utility methods
     */
    generateComponentId(component) {
        return `component_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    cleanupComponentResources(componentData) {
        // Override in specific implementations
        if (componentData.cleanup && typeof componentData.cleanup === 'function') {
            componentData.cleanup();
        }
    }

    getEventListenerCount() {
        let count = 0;
        for (const listeners of this.eventListenerRegistry.values()) {
            count += listeners.size;
        }
        return count;
    }

    getTotalCacheSize() {
        return this.cacheStats.totalSize;
    }

    estimateObjectSize(obj) {
        // Rough estimation of object size in bytes
        return JSON.stringify(obj).length * 2; // Approximate UTF-16 encoding
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    /**
     * Get memory statistics
     */
    getStats() {
        return {
            memory: {
                current: this.formatBytes(this.memoryStats.current),
                peak: this.formatBytes(this.memoryStats.peak),
                baseline: this.formatBytes(this.memoryStats.baseline),
                pressure: this.memoryPressure.level
            },
            components: {
                registered: this.componentRegistry.size,
                eventListeners: this.getEventListenerCount(),
                timers: this.timerRegistry.size,
                observers: this.observerRegistry.size
            },
            cache: {
                totalSize: this.formatBytes(this.cacheStats.totalSize),
                cacheCount: this.caches.size,
                hitRate: this.cacheStats.hits / (this.cacheStats.hits + this.cacheStats.misses) || 0
            },
            cleanup: {
                cleanups: this.memoryStats.cleanups,
                gcCount: this.memoryStats.gcCount,
                leaks: this.memoryStats.leaks.length
            }
        };
    }

    /**
     * Destroy memory optimizer
     */
    destroy() {
        // Clear all timers
        if (this.monitoringTimer) {
            clearInterval(this.monitoringTimer);
        }
        
        if (this.reportingTimer) {
            clearInterval(this.reportingTimer);
        }
        
        if (this.gcTimer) {
            clearInterval(this.gcTimer);
        }

        // Perform final cleanup
        this.performAggressiveCleanup();

        // Clear all registries
        this.componentRegistry.clear();
        this.timerRegistry.clear();
        this.observerRegistry.clear();
        this.caches.clear();

        console.log('🛑 Memory optimizer destroyed');
    }
}

/**
 * LRU Cache implementation for memory optimization
 */
class LRUCache {
    constructor(options = {}) {
        this.maxSize = options.maxSize || 10 * 1024 * 1024; // 10MB
        this.maxAge = options.maxAge || 3600000; // 1 hour
        this.onEvict = options.onEvict || (() => {});
        
        this.cache = new Map();
        this.sizes = new Map();
        this.accessTimes = new Map();
        this.currentSize = 0;
    }

    get(key) {
        if (!this.cache.has(key)) {
            return undefined;
        }

        // Check if expired
        const accessTime = this.accessTimes.get(key);
        if (Date.now() - accessTime > this.maxAge) {
            this.delete(key);
            return undefined;
        }

        // Update access time
        this.accessTimes.set(key, Date.now());
        
        // Move to end (most recently used)
        const value = this.cache.get(key);
        this.cache.delete(key);
        this.cache.set(key, value);
        
        return value;
    }

    set(key, value) {
        const size = this.estimateSize(value);
        
        // Remove existing entry if present
        if (this.cache.has(key)) {
            this.delete(key);
        }

        // Check if single item exceeds max size
        if (size > this.maxSize) {
            console.warn('Item too large for cache:', key);
            return false;
        }

        // Evict items if necessary
        while (this.currentSize + size > this.maxSize && this.cache.size > 0) {
            this.evictLRU();
        }

        // Add new item
        this.cache.set(key, value);
        this.sizes.set(key, size);
        this.accessTimes.set(key, Date.now());
        this.currentSize += size;

        return true;
    }

    delete(key) {
        if (!this.cache.has(key)) {
            return false;
        }

        const value = this.cache.get(key);
        const size = this.sizes.get(key);

        this.cache.delete(key);
        this.sizes.delete(key);
        this.accessTimes.delete(key);
        this.currentSize -= size;

        this.onEvict(key, value);
        return true;
    }

    clear() {
        for (const [key, value] of this.cache) {
            this.onEvict(key, value);
        }
        
        this.cache.clear();
        this.sizes.clear();
        this.accessTimes.clear();
        this.currentSize = 0;
    }

    evictLRU() {
        const firstKey = this.cache.keys().next().value;
        if (firstKey !== undefined) {
            this.delete(firstKey);
        }
    }

    evictOld() {
        const now = Date.now();
        const keysToEvict = [];

        for (const [key, accessTime] of this.accessTimes) {
            if (now - accessTime > this.maxAge) {
                keysToEvict.push(key);
            }
        }

        keysToEvict.forEach(key => this.delete(key));
    }

    estimateSize(value) {
        if (typeof value === 'string') {
            return value.length * 2; // UTF-16
        }
        
        if (value instanceof ArrayBuffer) {
            return value.byteLength;
        }
        
        if (value instanceof Blob) {
            return value.size;
        }
        
        // Fallback: JSON stringify
        try {
            return JSON.stringify(value).length * 2;
        } catch {
            return 1000; // Default estimate
        }
    }

    get size() {
        return this.cache.size;
    }

    get totalSize() {
        return this.currentSize;
    }
}

// Export for use in applications
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { MemoryOptimizer, LRUCache };
} else if (typeof window !== 'undefined') {
    window.MemoryOptimizer = MemoryOptimizer;
    window.LRUCache = LRUCache;
}

// Auto-initialize if config is available
if (typeof window !== 'undefined' && window.MEMORY_OPTIMIZER_CONFIG) {
    const memoryOptimizer = new MemoryOptimizer(window.MEMORY_OPTIMIZER_CONFIG);
    window.memoryOptimizer = memoryOptimizer;
}