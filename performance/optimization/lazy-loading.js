/**
 * Advanced Lazy Loading System for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive lazy loading capabilities including:
 * - Component lazy loading with React.lazy and Suspense
 * - Route-based code splitting
 * - Image lazy loading with progressive enhancement
 * - Module lazy loading with dynamic imports
 * - Resource preloading and prefetching
 * - Intersection Observer optimizations
 * - Regional optimization for different markets
 */

import React, { Suspense, lazy, useState, useEffect, useRef, useCallback } from 'react';

/**
 * Enhanced lazy loading hook with error boundaries and retry logic
 */
export const useLazyComponent = (importFunc, options = {}) => {
    const {
        fallback = null,
        errorFallback = null,
        retryCount = 3,
        retryDelay = 1000,
        preload = false,
        region = 'us-east-1'
    } = options;

    const [Component, setComponent] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [retries, setRetries] = useState(0);
    const mountedRef = useRef(true);

    const loadComponent = useCallback(async () => {
        if (Component || loading) return Component;

        setLoading(true);
        setError(null);

        try {
            // Add regional optimization
            const startTime = performance.now();
            
            const module = await importFunc();
            const LazyComponent = lazy(() => Promise.resolve(module));

            if (mountedRef.current) {
                setComponent(() => LazyComponent);
                
                // Track loading performance
                const loadTime = performance.now() - startTime;
                trackLazyLoadPerformance({
                    type: 'component',
                    loadTime,
                    region,
                    success: true
                });
            }

            return LazyComponent;

        } catch (err) {
            console.error('Failed to load lazy component:', err);
            
            if (mountedRef.current) {
                setError(err);
                
                // Retry logic
                if (retries < retryCount) {
                    setTimeout(() => {
                        if (mountedRef.current) {
                            setRetries(prev => prev + 1);
                            loadComponent();
                        }
                    }, retryDelay * Math.pow(2, retries)); // Exponential backoff
                }

                // Track loading failure
                trackLazyLoadPerformance({
                    type: 'component',
                    region,
                    success: false,
                    error: err.message,
                    retries
                });
            }
        } finally {
            if (mountedRef.current) {
                setLoading(false);
            }
        }
    }, [importFunc, Component, loading, retries, retryCount, retryDelay, region]);

    useEffect(() => {
        if (preload) {
            loadComponent();
        }

        return () => {
            mountedRef.current = false;
        };
    }, [preload, loadComponent]);

    const LazyWrapper = useCallback(({ children, ...props }) => {
        if (error && errorFallback) {
            return errorFallback(error, () => {
                setError(null);
                setRetries(0);
                loadComponent();
            });
        }

        if (!Component) {
            loadComponent();
            return fallback;
        }

        return (
            <Suspense fallback={fallback}>
                <Component {...props}>
                    {children}
                </Component>
            </Suspense>
        );
    }, [Component, error, errorFallback, fallback, loadComponent]);

    return {
        Component: LazyWrapper,
        loading,
        error,
        retries,
        preload: loadComponent
    };
};

/**
 * Route-based lazy loading with preloading strategies
 */
export class RouteLazyLoader {
    constructor(options = {}) {
        this.options = {
            preloadStrategy: 'hover', // 'hover', 'viewport', 'idle', 'none'
            preloadDelay: 200,
            region: 'us-east-1',
            enablePrefetch: true,
            enablePreload: true,
            ...options
        };

        this.routeCache = new Map();
        this.preloadQueue = new Set();
        this.intersectionObserver = null;
        this.idleCallback = null;

        this.initializeObservers();
    }

    /**
     * Initialize intersection and idle observers
     */
    initializeObservers() {
        // Intersection Observer for viewport-based preloading
        if ('IntersectionObserver' in window) {
            this.intersectionObserver = new IntersectionObserver(
                (entries) => {
                    entries.forEach(entry => {
                        if (entry.isIntersecting) {
                            const route = entry.target.dataset.route;
                            if (route) {
                                this.preloadRoute(route);
                            }
                        }
                    });
                },
                {
                    rootMargin: '50px',
                    threshold: 0.1
                }
            );
        }

        // Idle callback for background preloading
        if ('requestIdleCallback' in window) {
            this.scheduleIdlePreloading();
        }
    }

    /**
     * Create lazy route component
     */
    createLazyRoute(routePath, importFunc, options = {}) {
        const {
            fallback = <div>Loading...</div>,
            errorBoundary = true,
            preload = false,
            priority = 'normal'
        } = options;

        // Cache the import function
        this.routeCache.set(routePath, {
            importFunc,
            loaded: false,
            loading: false,
            component: null,
            priority,
            preload
        });

        const LazyComponent = lazy(async () => {
            const routeData = this.routeCache.get(routePath);
            
            if (routeData.component) {
                return routeData.component;
            }

            routeData.loading = true;
            const startTime = performance.now();

            try {
                const module = await importFunc();
                const loadTime = performance.now() - startTime;

                routeData.component = module;
                routeData.loaded = true;
                routeData.loading = false;

                // Track performance
                trackLazyLoadPerformance({
                    type: 'route',
                    route: routePath,
                    loadTime,
                    region: this.options.region,
                    success: true,
                    priority
                });

                return module;

            } catch (error) {
                routeData.loading = false;
                
                trackLazyLoadPerformance({
                    type: 'route',
                    route: routePath,
                    region: this.options.region,
                    success: false,
                    error: error.message,
                    priority
                });

                throw error;
            }
        });

        // Preload if requested
        if (preload) {
            this.preloadRoute(routePath);
        }

        const WrappedComponent = (props) => {
            if (errorBoundary) {
                return (
                    <RouteErrorBoundary routePath={routePath}>
                        <Suspense fallback={fallback}>
                            <LazyComponent {...props} />
                        </Suspense>
                    </RouteErrorBoundary>
                );
            }

            return (
                <Suspense fallback={fallback}>
                    <LazyComponent {...props} />
                </Suspense>
            );
        };

        return WrappedComponent;
    }

    /**
     * Preload route component
     */
    async preloadRoute(routePath) {
        const routeData = this.routeCache.get(routePath);
        
        if (!routeData || routeData.loaded || routeData.loading) {
            return;
        }

        if (this.preloadQueue.has(routePath)) {
            return;
        }

        this.preloadQueue.add(routePath);

        try {
            await routeData.importFunc();
            routeData.loaded = true;
        } catch (error) {
            console.warn(`Failed to preload route ${routePath}:`, error);
        } finally {
            this.preloadQueue.delete(routePath);
        }
    }

    /**
     * Setup hover preloading for navigation links
     */
    setupHoverPreloading() {
        if (this.options.preloadStrategy !== 'hover') return;

        document.addEventListener('mouseover', (event) => {
            const link = event.target.closest('a[data-route]');
            if (link) {
                const route = link.dataset.route;
                setTimeout(() => {
                    this.preloadRoute(route);
                }, this.options.preloadDelay);
            }
        });
    }

    /**
     * Setup viewport preloading for navigation links
     */
    setupViewportPreloading() {
        if (this.options.preloadStrategy !== 'viewport' || !this.intersectionObserver) return;

        const links = document.querySelectorAll('a[data-route]');
        links.forEach(link => {
            this.intersectionObserver.observe(link);
        });
    }

    /**
     * Schedule idle preloading
     */
    scheduleIdlePreloading() {
        if (this.options.preloadStrategy !== 'idle') return;

        this.idleCallback = requestIdleCallback(() => {
            this.preloadHighPriorityRoutes();
        });
    }

    /**
     * Preload high priority routes during idle time
     */
    preloadHighPriorityRoutes() {
        const highPriorityRoutes = Array.from(this.routeCache.entries())
            .filter(([_, data]) => data.priority === 'high' && !data.loaded)
            .map(([route]) => route);

        highPriorityRoutes.forEach(route => {
            this.preloadRoute(route);
        });
    }

    /**
     * Get preloading statistics
     */
    getStats() {
        const total = this.routeCache.size;
        const loaded = Array.from(this.routeCache.values()).filter(data => data.loaded).length;
        const loading = Array.from(this.routeCache.values()).filter(data => data.loading).length;

        return {
            total,
            loaded,
            loading,
            pending: total - loaded - loading,
            preloadQueue: this.preloadQueue.size
        };
    }
}

/**
 * Advanced image lazy loading with progressive enhancement
 */
export class ImageLazyLoader {
    constructor(options = {}) {
        this.options = {
            rootMargin: '50px',
            threshold: 0.1,
            enableWebP: true,
            enableAVIF: true,
            enableBlurUp: true,
            enableProgressiveLoading: true,
            quality: 80,
            sizes: ['320w', '640w', '1024w', '1920w'],
            region: 'us-east-1',
            ...options
        };

        this.observer = null;
        this.loadedImages = new Set();
        this.formatSupport = {};

        this.init();
    }

    /**
     * Initialize image lazy loader
     */
    async init() {
        // Check format support
        await this.checkFormatSupport();

        // Initialize intersection observer
        this.initializeObserver();

        // Process existing images
        this.processExistingImages();
    }

    /**
     * Check browser support for modern image formats
     */
    async checkFormatSupport() {
        const formats = ['webp', 'avif'];
        
        for (const format of formats) {
            this.formatSupport[format] = await this.supportsFormat(format);
        }
    }

    /**
     * Check if browser supports specific image format
     */
    supportsFormat(format) {
        return new Promise((resolve) => {
            const img = new Image();
            img.onload = () => resolve(true);
            img.onerror = () => resolve(false);
            
            const testImages = {
                webp: 'data:image/webp;base64,UklGRjoAAABXRUJQVlA4IC4AAACyAgCdASoCAAIALmk0mk0iIiIiIgBoSygABc6WWgAA/veff/0PP8bA//LwYAAA',
                avif: 'data:image/avif;base64,AAAAIGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZk1BMUIAAADybWV0YQAAAAAAAAAoaGRscgAAAAAAAAAAcGljdAAAAAAAAAAAAAAAAGxpYmF2aWYAAAAADnBpdG0AAAAAAAEAAAAeaWxvYwAAAABEAAABAAEAAAABAAABGgAAAB0AAAAoaWluZgAAAAAAAQAAABppbmZlAgAAAAABAABhdjAxQ29sb3IAAAAAamlwcnAAAABLaXBjbwAAABRpc3BlAAAAAAAAAAIAAAACAAAAEHBpeGkAAAAAAwgICAAAAAxhdjFDgQ0MAAAAABNjb2xybmNseAACAAIAAYAAAAAXaXBtYQAAAAAAAAABAAEEAQKDBAAAACVtZGF0EgAKCBgABogQEAwgMg8f8D///8WfhwB8+ErK42A='
            };
            
            img.src = testImages[format];
        });
    }

    /**
     * Initialize intersection observer
     */
    initializeObserver() {
        if (!('IntersectionObserver' in window)) {
            // Fallback for browsers without IntersectionObserver
            this.loadAllImages();
            return;
        }

        this.observer = new IntersectionObserver(
            (entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        this.loadImage(entry.target);
                        this.observer.unobserve(entry.target);
                    }
                });
            },
            {
                rootMargin: this.options.rootMargin,
                threshold: this.options.threshold
            }
        );
    }

    /**
     * Process existing images on the page
     */
    processExistingImages() {
        const images = document.querySelectorAll('img[data-src], img[data-srcset]');
        images.forEach(img => this.observeImage(img));
    }

    /**
     * Observe image for lazy loading
     */
    observeImage(img) {
        if (this.observer) {
            this.observer.observe(img);
        } else {
            this.loadImage(img);
        }
    }

    /**
     * Load individual image with progressive enhancement
     */
    async loadImage(img) {
        if (this.loadedImages.has(img)) return;

        const startTime = performance.now();
        
        try {
            // Show blur-up placeholder if enabled
            if (this.options.enableBlurUp && img.dataset.placeholder) {
                this.showBlurUpPlaceholder(img);
            }

            // Determine best format and source
            const { src, srcset } = this.getBestImageSource(img);

            // Preload the image
            await this.preloadImage(src);

            // Apply the image
            if (srcset) {
                img.srcset = srcset;
            }
            img.src = src;

            // Remove data attributes
            delete img.dataset.src;
            delete img.dataset.srcset;

            // Add loaded class
            img.classList.add('lazy-loaded');

            // Remove blur-up effect
            if (this.options.enableBlurUp) {
                this.removeBlurUpEffect(img);
            }

            this.loadedImages.add(img);

            // Track performance
            const loadTime = performance.now() - startTime;
            trackLazyLoadPerformance({
                type: 'image',
                src,
                loadTime,
                region: this.options.region,
                success: true,
                format: this.getImageFormat(src)
            });

        } catch (error) {
            console.error('Failed to load lazy image:', error);
            
            // Fallback to original source
            if (img.dataset.fallback) {
                img.src = img.dataset.fallback;
            }

            trackLazyLoadPerformance({
                type: 'image',
                src: img.dataset.src,
                region: this.options.region,
                success: false,
                error: error.message
            });
        }
    }

    /**
     * Get best image source based on format support and device capabilities
     */
    getBestImageSource(img) {
        const dataSrc = img.dataset.src;
        const dataSrcset = img.dataset.srcset;

        if (!dataSrc) return { src: img.src };

        // Check for modern format alternatives
        let bestSrc = dataSrc;
        let bestSrcset = dataSrcset;

        // AVIF support
        if (this.options.enableAVIF && this.formatSupport.avif && img.dataset.avif) {
            bestSrc = img.dataset.avif;
            bestSrcset = img.dataset.avifSrcset;
        }
        // WebP support
        else if (this.options.enableWebP && this.formatSupport.webp && img.dataset.webp) {
            bestSrc = img.dataset.webp;
            bestSrcset = img.dataset.webpSrcset;
        }

        // Add quality and regional optimization
        bestSrc = this.optimizeImageUrl(bestSrc);
        
        if (bestSrcset) {
            bestSrcset = this.optimizeSrcset(bestSrcset);
        }

        return { src: bestSrc, srcset: bestSrcset };
    }

    /**
     * Optimize image URL with quality and regional parameters
     */
    optimizeImageUrl(url) {
        if (!url.includes('?')) {
            url += '?';
        } else {
            url += '&';
        }

        url += `q=${this.options.quality}&region=${this.options.region}`;

        // Add device pixel ratio optimization
        const dpr = window.devicePixelRatio || 1;
        if (dpr > 1) {
            url += `&dpr=${Math.min(dpr, 3)}`;
        }

        return url;
    }

    /**
     * Optimize srcset with quality parameters
     */
    optimizeSrcset(srcset) {
        return srcset.split(',').map(src => {
            const [url, descriptor] = src.trim().split(' ');
            return `${this.optimizeImageUrl(url)} ${descriptor}`;
        }).join(', ');
    }

    /**
     * Preload image
     */
    preloadImage(src) {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = resolve;
            img.onerror = reject;
            img.src = src;
        });
    }

    /**
     * Show blur-up placeholder
     */
    showBlurUpPlaceholder(img) {
        const placeholder = img.dataset.placeholder;
        if (placeholder) {
            img.style.backgroundImage = `url(${placeholder})`;
            img.style.backgroundSize = 'cover';
            img.style.backgroundPosition = 'center';
            img.style.filter = 'blur(5px)';
            img.style.transition = 'filter 0.3s ease';
        }
    }

    /**
     * Remove blur-up effect
     */
    removeBlurUpEffect(img) {
        setTimeout(() => {
            img.style.filter = 'none';
            img.style.backgroundImage = 'none';
        }, 100);
    }

    /**
     * Get image format from URL
     */
    getImageFormat(url) {
        const extension = url.split('.').pop().split('?')[0].toLowerCase();
        return extension;
    }

    /**
     * Load all images (fallback)
     */
    loadAllImages() {
        const images = document.querySelectorAll('img[data-src]');
        images.forEach(img => this.loadImage(img));
    }

    /**
     * Add new images to lazy loading
     */
    addImages(images) {
        images.forEach(img => this.observeImage(img));
    }

    /**
     * Get loading statistics
     */
    getStats() {
        const totalImages = document.querySelectorAll('img[data-src], img.lazy-loaded').length;
        const loadedImages = this.loadedImages.size;

        return {
            total: totalImages,
            loaded: loadedImages,
            pending: totalImages - loadedImages,
            formatSupport: this.formatSupport
        };
    }

    /**
     * Destroy lazy loader
     */
    destroy() {
        if (this.observer) {
            this.observer.disconnect();
        }
        this.loadedImages.clear();
    }
}

/**
 * Module lazy loader for dynamic imports
 */
export class ModuleLazyLoader {
    constructor(options = {}) {
        this.options = {
            retryCount: 3,
            retryDelay: 1000,
            enableCaching: true,
            enablePreloading: true,
            region: 'us-east-1',
            ...options
        };

        this.moduleCache = new Map();
        this.preloadQueue = new Set();
    }

    /**
     * Lazy load module with caching and retry logic
     */
    async loadModule(modulePath, options = {}) {
        const {
            retryCount = this.options.retryCount,
            retryDelay = this.options.retryDelay,
            enableCaching = this.options.enableCaching
        } = options;

        // Check cache first
        if (enableCaching && this.moduleCache.has(modulePath)) {
            const cached = this.moduleCache.get(modulePath);
            if (cached.loaded) {
                return cached.module;
            }
            if (cached.loading) {
                return cached.promise;
            }
        }

        const startTime = performance.now();
        let retries = 0;

        const loadPromise = async () => {
            while (retries <= retryCount) {
                try {
                    const module = await import(modulePath);
                    
                    if (enableCaching) {
                        this.moduleCache.set(modulePath, {
                            module,
                            loaded: true,
                            loading: false,
                            loadTime: performance.now() - startTime
                        });
                    }

                    // Track performance
                    trackLazyLoadPerformance({
                        type: 'module',
                        module: modulePath,
                        loadTime: performance.now() - startTime,
                        region: this.options.region,
                        success: true,
                        retries
                    });

                    return module;

                } catch (error) {
                    retries++;
                    
                    if (retries > retryCount) {
                        // Track failure
                        trackLazyLoadPerformance({
                            type: 'module',
                            module: modulePath,
                            region: this.options.region,
                            success: false,
                            error: error.message,
                            retries: retries - 1
                        });

                        throw error;
                    }

                    // Wait before retry with exponential backoff
                    await new Promise(resolve => 
                        setTimeout(resolve, retryDelay * Math.pow(2, retries - 1))
                    );
                }
            }
        };

        const promise = loadPromise();

        if (enableCaching) {
            this.moduleCache.set(modulePath, {
                module: null,
                loaded: false,
                loading: true,
                promise
            });
        }

        return promise;
    }

    /**
     * Preload module
     */
    async preloadModule(modulePath) {
        if (this.preloadQueue.has(modulePath)) return;
        
        this.preloadQueue.add(modulePath);
        
        try {
            await this.loadModule(modulePath);
        } catch (error) {
            console.warn(`Failed to preload module ${modulePath}:`, error);
        } finally {
            this.preloadQueue.delete(modulePath);
        }
    }

    /**
     * Preload multiple modules
     */
    async preloadModules(modulePaths) {
        const promises = modulePaths.map(path => this.preloadModule(path));
        await Promise.allSettled(promises);
    }

    /**
     * Get module cache statistics
     */
    getStats() {
        const cached = Array.from(this.moduleCache.values());
        const loaded = cached.filter(item => item.loaded).length;
        const loading = cached.filter(item => item.loading).length;

        return {
            total: this.moduleCache.size,
            loaded,
            loading,
            preloading: this.preloadQueue.size,
            averageLoadTime: cached
                .filter(item => item.loadTime)
                .reduce((sum, item, _, arr) => sum + item.loadTime / arr.length, 0)
        };
    }

    /**
     * Clear module cache
     */
    clearCache() {
        this.moduleCache.clear();
        this.preloadQueue.clear();
    }
}

/**
 * Route Error Boundary Component
 */
class RouteErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error };
    }

    componentDidCatch(error, errorInfo) {
        console.error(`Route error in ${this.props.routePath}:`, error, errorInfo);
        
        // Track error
        trackLazyLoadPerformance({
            type: 'route-error',
            route: this.props.routePath,
            success: false,
            error: error.message
        });
    }

    render() {
        if (this.state.hasError) {
            return (
                <div className="route-error">
                    <h2>Something went wrong loading this page.</h2>
                    <button onClick={() => window.location.reload()}>
                        Reload Page
                    </button>
                </div>
            );
        }

        return this.props.children;
    }
}

/**
 * Performance tracking utility
 */
function trackLazyLoadPerformance(data) {
    // Send performance data to analytics
    if (typeof window !== 'undefined' && window.analytics) {
        window.analytics.track('lazy_load_performance', data);
    }

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
        console.log('🚀 Lazy Load Performance:', data);
    }
}

/**
 * Global lazy loading manager
 */
export class LazyLoadingManager {
    constructor(options = {}) {
        this.routeLoader = new RouteLazyLoader(options.route);
        this.imageLoader = new ImageLazyLoader(options.image);
        this.moduleLoader = new ModuleLazyLoader(options.module);
        
        this.init();
    }

    async init() {
        // Initialize all loaders
        await Promise.all([
            this.imageLoader.init()
        ]);

        // Setup global event listeners
        this.setupEventListeners();

        console.log('🚀 Lazy loading manager initialized');
    }

    setupEventListeners() {
        // Handle dynamic content
        const observer = new MutationObserver((mutations) => {
            mutations.forEach(mutation => {
                mutation.addedNodes.forEach(node => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        // Check for new images
                        const images = node.querySelectorAll('img[data-src]');
                        if (images.length > 0) {
                            this.imageLoader.addImages(Array.from(images));
                        }
                    }
                });
            });
        });

        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
    }

    getStats() {
        return {
            routes: this.routeLoader.getStats(),
            images: this.imageLoader.getStats(),
            modules: this.moduleLoader.getStats()
        };
    }

    destroy() {
        this.imageLoader.destroy();
        this.moduleLoader.clearCache();
    }
}

// Export default instance
export default new LazyLoadingManager();