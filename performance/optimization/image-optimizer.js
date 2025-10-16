/**
 * Advanced Image Optimization for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive image optimization including:
 * - Lazy loading with intersection observer
 * - Responsive image loading
 * - WebP/AVIF format detection and serving
 * - Image compression and resizing
 * - Progressive loading
 * - Blur-up technique
 * - Critical image prioritization
 * - Regional CDN optimization
 */

class ImageOptimizer {
    constructor(config = {}) {
        this.config = {
            // Lazy loading settings
            rootMargin: config.rootMargin || '50px',
            threshold: config.threshold || 0.1,
            enableLazyLoading: config.enableLazyLoading !== false,
            
            // Format support
            enableWebP: config.enableWebP !== false,
            enableAVIF: config.enableAVIF !== false,
            fallbackFormat: config.fallbackFormat || 'jpg',
            
            // Quality settings
            defaultQuality: config.defaultQuality || 80,
            retinaQuality: config.retinaQuality || 70,
            thumbnailQuality: config.thumbnailQuality || 60,
            
            // Size settings
            maxWidth: config.maxWidth || 1920,
            maxHeight: config.maxHeight || 1080,
            thumbnailSize: config.thumbnailSize || 300,
            
            // Progressive loading
            enableProgressiveLoading: config.enableProgressiveLoading !== false,
            enableBlurUp: config.enableBlurUp !== false,
            blurRadius: config.blurRadius || 10,
            
            // CDN settings
            cdnBaseUrl: config.cdnBaseUrl || '',
            region: config.region || 'us-east-1',
            enableCDN: config.enableCDN !== false,
            
            // Performance
            enableCriticalImages: config.enableCriticalImages !== false,
            criticalImageCount: config.criticalImageCount || 3,
            preloadCritical: config.preloadCritical !== false,
            
            ...config
        };

        this.observer = null;
        this.formatSupport = {};
        this.loadedImages = new Set();
        this.criticalImages = new Set();
        this.imageCache = new Map();
        
        this.init();
    }

    /**
     * Initialize image optimizer
     */
    async init() {
        try {
            // Detect format support
            await this.detectFormatSupport();
            
            // Initialize lazy loading
            if (this.config.enableLazyLoading) {
                this.initLazyLoading();
            }
            
            // Initialize critical image handling
            if (this.config.enableCriticalImages) {
                this.initCriticalImages();
            }
            
            // Set up responsive image handling
            this.initResponsiveImages();
            
            // Set up progressive loading
            if (this.config.enableProgressiveLoading) {
                this.initProgressiveLoading();
            }
            
            console.log('🖼️ Image optimizer initialized');
            
        } catch (error) {
            console.error('❌ Failed to initialize image optimizer:', error);
        }
    }

    /**
     * Detect browser format support
     */
    async detectFormatSupport() {
        const formats = ['webp', 'avif'];
        
        for (const format of formats) {
            this.formatSupport[format] = await this.canUseFormat(format);
        }
        
        console.log('📋 Format support detected:', this.formatSupport);
    }

    /**
     * Check if browser supports image format
     */
    canUseFormat(format) {
        return new Promise((resolve) => {
            const img = new Image();
            
            img.onload = () => resolve(true);
            img.onerror = () => resolve(false);
            
            // Test images for format detection
            const testImages = {
                webp: 'data:image/webp;base64,UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEADsD+JaQAA3AAAAAA',
                avif: 'data:image/avif;base64,AAAAIGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZk1BMUIAAADybWV0YQAAAAAAAAAoaGRscgAAAAAAAAAAcGljdAAAAAAAAAAAAAAAAGxpYmF2aWYAAAAADnBpdG0AAAAAAAEAAAAeaWxvYwAAAABEAAABAAEAAAABAAABGgAAAB0AAAAoaWluZgAAAAAAAQAAABppbmZlAgAAAAABAABhdjAxQ29sb3IAAAAAamlwcnAAAABLaXBjbwAAABRpc3BlAAAAAAAAAAIAAAACAAAAEHBpeGkAAAAAAwgICAAAAAxhdjFDgQ0MAAAAABNjb2xybmNseAACAAIAAYAAAAAXaXBtYQAAAAAAAAABAAEEAQKDBAAAACVtZGF0EgAKCBgABogQEAwgMg8f8D///8WfhwB8+ErK42A='
            };
            
            img.src = testImages[format];
        });
    }

    /**
     * Initialize lazy loading with Intersection Observer
     */
    initLazyLoading() {
        if (!('IntersectionObserver' in window)) {
            // Fallback for browsers without Intersection Observer
            this.loadAllImages();
            return;
        }

        this.observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    this.loadImage(entry.target);
                    this.observer.unobserve(entry.target);
                }
            });
        }, {
            rootMargin: this.config.rootMargin,
            threshold: this.config.threshold
        });

        // Observe all lazy images
        this.observeLazyImages();
    }

    /**
     * Initialize critical image handling
     */
    initCriticalImages() {
        // Find critical images (above the fold)
        const images = document.querySelectorAll('img[data-src], img[src]');
        let criticalCount = 0;

        images.forEach((img, index) => {
            if (criticalCount < this.config.criticalImageCount) {
                const rect = img.getBoundingClientRect();
                const isAboveFold = rect.top < window.innerHeight;
                
                if (isAboveFold || index < this.config.criticalImageCount) {
                    this.criticalImages.add(img);
                    img.setAttribute('data-critical', 'true');
                    criticalCount++;
                    
                    // Load critical images immediately
                    if (img.hasAttribute('data-src')) {
                        this.loadImage(img, true);
                    }
                }
            }
        });

        // Preload critical images
        if (this.config.preloadCritical) {
            this.preloadCriticalImages();
        }
    }

    /**
     * Initialize responsive image handling
     */
    initResponsiveImages() {
        // Handle window resize for responsive images
        let resizeTimeout;
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                this.updateResponsiveImages();
            }, 250);
        });

        // Handle device pixel ratio changes
        if ('matchMedia' in window) {
            const mediaQuery = window.matchMedia('(min-resolution: 2dppx)');
            mediaQuery.addListener(() => {
                this.updateResponsiveImages();
            });
        }
    }

    /**
     * Initialize progressive loading
     */
    initProgressiveLoading() {
        // Add CSS for progressive loading effects
        if (!document.getElementById('image-optimizer-styles')) {
            const style = document.createElement('style');
            style.id = 'image-optimizer-styles';
            style.textContent = `
                .img-loading {
                    filter: blur(${this.config.blurRadius}px);
                    transition: filter 0.3s ease;
                }
                
                .img-loaded {
                    filter: blur(0);
                }
                
                .img-placeholder {
                    background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
                    background-size: 200% 100%;
                    animation: loading 1.5s infinite;
                }
                
                @keyframes loading {
                    0% { background-position: 200% 0; }
                    100% { background-position: -200% 0; }
                }
                
                .img-error {
                    background: #f5f5f5;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: #999;
                    font-size: 14px;
                }
            `;
            document.head.appendChild(style);
        }
    }

    /**
     * Observe lazy images
     */
    observeLazyImages() {
        const lazyImages = document.querySelectorAll('img[data-src]:not([data-critical="true"])');
        
        lazyImages.forEach(img => {
            // Add placeholder while loading
            this.addPlaceholder(img);
            
            if (this.observer) {
                this.observer.observe(img);
            }
        });
    }

    /**
     * Load image with optimization
     */
    async loadImage(img, isCritical = false) {
        if (this.loadedImages.has(img)) {
            return;
        }

        const originalSrc = img.getAttribute('data-src') || img.src;
        if (!originalSrc) return;

        try {
            // Generate optimized image URL
            const optimizedSrc = this.getOptimizedImageUrl(originalSrc, img);
            
            // Create new image for preloading
            const newImg = new Image();
            
            // Set up loading states
            if (this.config.enableProgressiveLoading) {
                img.classList.add('img-loading');
            }
            
            // Handle successful load
            newImg.onload = () => {
                img.src = optimizedSrc;
                img.removeAttribute('data-src');
                
                if (this.config.enableProgressiveLoading) {
                    img.classList.remove('img-loading');
                    img.classList.add('img-loaded');
                }
                
                this.loadedImages.add(img);
                this.removePlaceholder(img);
                
                // Trigger custom event
                img.dispatchEvent(new CustomEvent('imageLoaded', {
                    detail: { src: optimizedSrc, isCritical }
                }));
            };
            
            // Handle load error
            newImg.onerror = () => {
                this.handleImageError(img, originalSrc);
            };
            
            // Start loading
            newImg.src = optimizedSrc;
            
            // Set loading priority
            if (isCritical) {
                newImg.loading = 'eager';
                newImg.fetchPriority = 'high';
            } else {
                newImg.loading = 'lazy';
                newImg.fetchPriority = 'low';
            }
            
        } catch (error) {
            console.error('Failed to load image:', error);
            this.handleImageError(img, originalSrc);
        }
    }

    /**
     * Generate optimized image URL
     */
    getOptimizedImageUrl(originalSrc, img) {
        // Check cache first
        const cacheKey = this.getCacheKey(originalSrc, img);
        if (this.imageCache.has(cacheKey)) {
            return this.imageCache.get(cacheKey);
        }

        let optimizedUrl = originalSrc;

        // Use CDN if enabled
        if (this.config.enableCDN && this.config.cdnBaseUrl) {
            optimizedUrl = this.buildCDNUrl(originalSrc, img);
        } else {
            optimizedUrl = this.buildOptimizedUrl(originalSrc, img);
        }

        // Cache the result
        this.imageCache.set(cacheKey, optimizedUrl);
        
        return optimizedUrl;
    }

    /**
     * Build CDN URL with optimizations
     */
    buildCDNUrl(src, img) {
        const url = new URL(this.config.cdnBaseUrl);
        
        // Get image dimensions
        const { width, height } = this.getImageDimensions(img);
        const quality = this.getImageQuality(img);
        const format = this.getBestFormat();
        
        // Build CDN parameters
        const params = new URLSearchParams();
        
        if (width) params.set('w', width);
        if (height) params.set('h', height);
        params.set('q', quality);
        params.set('f', format);
        params.set('fit', 'cover');
        params.set('auto', 'compress');
        
        // Add regional optimization
        if (this.config.region) {
            params.set('region', this.config.region);
        }
        
        // Add source URL
        params.set('url', encodeURIComponent(src));
        
        url.search = params.toString();
        return url.toString();
    }

    /**
     * Build optimized URL for local processing
     */
    buildOptimizedUrl(src, img) {
        const url = new URL(src, window.location.origin);
        const params = new URLSearchParams(url.search);
        
        // Get image dimensions
        const { width, height } = this.getImageDimensions(img);
        const quality = this.getImageQuality(img);
        const format = this.getBestFormat();
        
        // Add optimization parameters
        if (width) params.set('w', width);
        if (height) params.set('h', height);
        params.set('q', quality);
        params.set('format', format);
        
        url.search = params.toString();
        return url.toString();
    }

    /**
     * Get optimal image dimensions
     */
    getImageDimensions(img) {
        const rect = img.getBoundingClientRect();
        const dpr = window.devicePixelRatio || 1;
        
        // Get container dimensions
        let width = rect.width || img.getAttribute('width') || img.naturalWidth;
        let height = rect.height || img.getAttribute('height') || img.naturalHeight;
        
        // Apply device pixel ratio for retina displays
        width = Math.ceil(width * dpr);
        height = Math.ceil(height * dpr);
        
        // Respect maximum dimensions
        if (width > this.config.maxWidth) {
            height = (height * this.config.maxWidth) / width;
            width = this.config.maxWidth;
        }
        
        if (height > this.config.maxHeight) {
            width = (width * this.config.maxHeight) / height;
            height = this.config.maxHeight;
        }
        
        return { width: Math.round(width), height: Math.round(height) };
    }

    /**
     * Get optimal image quality
     */
    getImageQuality(img) {
        const dpr = window.devicePixelRatio || 1;
        const isCritical = img.hasAttribute('data-critical');
        const isThumbnail = img.classList.contains('thumbnail');
        
        if (isThumbnail) {
            return this.config.thumbnailQuality;
        }
        
        if (dpr > 1) {
            return this.config.retinaQuality;
        }
        
        return this.config.defaultQuality;
    }

    /**
     * Get best supported format
     */
    getBestFormat() {
        if (this.config.enableAVIF && this.formatSupport.avif) {
            return 'avif';
        }
        
        if (this.config.enableWebP && this.formatSupport.webp) {
            return 'webp';
        }
        
        return this.config.fallbackFormat;
    }

    /**
     * Generate cache key
     */
    getCacheKey(src, img) {
        const { width, height } = this.getImageDimensions(img);
        const quality = this.getImageQuality(img);
        const format = this.getBestFormat();
        
        return `${src}-${width}x${height}-q${quality}-${format}`;
    }

    /**
     * Add placeholder while loading
     */
    addPlaceholder(img) {
        if (!this.config.enableBlurUp) return;
        
        const placeholder = img.getAttribute('data-placeholder');
        if (placeholder) {
            img.src = placeholder;
            img.classList.add('img-placeholder');
        } else {
            // Generate a simple placeholder
            const { width, height } = this.getImageDimensions(img);
            const canvas = document.createElement('canvas');
            canvas.width = Math.min(width, 40);
            canvas.height = Math.min(height, 40);
            
            const ctx = canvas.getContext('2d');
            ctx.fillStyle = '#f0f0f0';
            ctx.fillRect(0, 0, canvas.width, canvas.height);
            
            img.src = canvas.toDataURL();
            img.classList.add('img-placeholder');
        }
    }

    /**
     * Remove placeholder after loading
     */
    removePlaceholder(img) {
        img.classList.remove('img-placeholder');
    }

    /**
     * Handle image loading errors
     */
    handleImageError(img, originalSrc) {
        console.warn('Failed to load image:', originalSrc);
        
        img.classList.remove('img-loading', 'img-placeholder');
        img.classList.add('img-error');
        
        // Try fallback format
        if (!img.hasAttribute('data-fallback-tried')) {
            img.setAttribute('data-fallback-tried', 'true');
            
            const fallbackSrc = originalSrc.replace(/\.(webp|avif)$/i, '.jpg');
            if (fallbackSrc !== originalSrc) {
                img.src = fallbackSrc;
                return;
            }
        }
        
        // Show error placeholder
        img.alt = img.alt || 'Image failed to load';
        img.title = 'Failed to load image';
        
        // Trigger error event
        img.dispatchEvent(new CustomEvent('imageError', {
            detail: { src: originalSrc }
        }));
    }

    /**
     * Preload critical images
     */
    preloadCriticalImages() {
        this.criticalImages.forEach(img => {
            const src = img.getAttribute('data-src') || img.src;
            if (src) {
                const link = document.createElement('link');
                link.rel = 'preload';
                link.as = 'image';
                link.href = this.getOptimizedImageUrl(src, img);
                link.fetchPriority = 'high';
                document.head.appendChild(link);
            }
        });
    }

    /**
     * Update responsive images on resize
     */
    updateResponsiveImages() {
        const images = document.querySelectorAll('img[src], img[data-src]');
        
        images.forEach(img => {
            if (this.loadedImages.has(img)) {
                const currentSrc = img.src;
                const newSrc = this.getOptimizedImageUrl(currentSrc, img);
                
                if (newSrc !== currentSrc) {
                    img.src = newSrc;
                }
            }
        });
    }

    /**
     * Load all images (fallback for no Intersection Observer)
     */
    loadAllImages() {
        const lazyImages = document.querySelectorAll('img[data-src]');
        lazyImages.forEach(img => this.loadImage(img));
    }

    /**
     * Get optimization statistics
     */
    getStats() {
        return {
            loadedImages: this.loadedImages.size,
            criticalImages: this.criticalImages.size,
            formatSupport: this.formatSupport,
            cacheSize: this.imageCache.size,
            region: this.config.region
        };
    }

    /**
     * Destroy image optimizer
     */
    destroy() {
        if (this.observer) {
            this.observer.disconnect();
        }
        
        this.loadedImages.clear();
        this.criticalImages.clear();
        this.imageCache.clear();
        
        console.log('🛑 Image optimizer destroyed');
    }
}

// Export for use in applications
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ImageOptimizer;
} else if (typeof window !== 'undefined') {
    window.ImageOptimizer = ImageOptimizer;
}

// Auto-initialize if config is available
if (typeof window !== 'undefined' && window.IMAGE_OPTIMIZER_CONFIG) {
    const imageOptimizer = new ImageOptimizer(window.IMAGE_OPTIMIZER_CONFIG);
    window.imageOptimizer = imageOptimizer;
}