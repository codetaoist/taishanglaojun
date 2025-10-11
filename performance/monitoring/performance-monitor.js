/**
 * Performance Monitoring System for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive performance monitoring including:
 * - Core Web Vitals tracking
 * - Real User Monitoring (RUM)
 * - Synthetic monitoring
 * - Custom performance metrics
 * - Regional performance analysis
 * - Performance alerting
 */

class PerformanceMonitor {
    constructor(config = {}) {
        this.config = {
            apiEndpoint: config.apiEndpoint || '/api/performance',
            region: config.region || 'us-east-1',
            enableRUM: config.enableRUM !== false,
            enableSynthetic: config.enableSynthetic !== false,
            enableWebVitals: config.enableWebVitals !== false,
            enableCustomMetrics: config.enableCustomMetrics !== false,
            sampleRate: config.sampleRate || 0.1, // 10% sampling
            bufferSize: config.bufferSize || 100,
            flushInterval: config.flushInterval || 30000, // 30 seconds
            ...config
        };
        
        this.metrics = [];
        this.observers = new Map();
        this.isInitialized = false;
        
        // Performance budgets from config
        this.budgets = {
            FCP: 1800, // First Contentful Paint
            LCP: 2500, // Largest Contentful Paint
            FID: 100,  // First Input Delay
            CLS: 0.1,  // Cumulative Layout Shift
            TTFB: 600, // Time to First Byte
            TTI: 3800, // Time to Interactive
            SI: 3000   // Speed Index
        };
        
        this.init();
    }

    /**
     * Initialize performance monitoring
     */
    init() {
        if (this.isInitialized || typeof window === 'undefined') {
            return;
        }

        try {
            // Initialize Web Vitals monitoring
            if (this.config.enableWebVitals) {
                this.initWebVitals();
            }

            // Initialize Real User Monitoring
            if (this.config.enableRUM) {
                this.initRUM();
            }

            // Initialize custom metrics
            if (this.config.enableCustomMetrics) {
                this.initCustomMetrics();
            }

            // Initialize navigation timing
            this.initNavigationTiming();

            // Initialize resource timing
            this.initResourceTiming();

            // Initialize error tracking
            this.initErrorTracking();

            // Start periodic flushing
            this.startPeriodicFlush();

            this.isInitialized = true;
            console.log('🚀 Performance monitoring initialized');

        } catch (error) {
            console.error('❌ Failed to initialize performance monitoring:', error);
        }
    }

    /**
     * Initialize Core Web Vitals monitoring
     */
    initWebVitals() {
        // First Contentful Paint (FCP)
        this.observePerformanceEntry('paint', (entries) => {
            entries.forEach(entry => {
                if (entry.name === 'first-contentful-paint') {
                    this.recordMetric('FCP', entry.startTime, {
                        budget: this.budgets.FCP,
                        critical: true
                    });
                }
            });
        });

        // Largest Contentful Paint (LCP)
        this.observePerformanceEntry('largest-contentful-paint', (entries) => {
            const lastEntry = entries[entries.length - 1];
            if (lastEntry) {
                this.recordMetric('LCP', lastEntry.startTime, {
                    budget: this.budgets.LCP,
                    critical: true,
                    element: lastEntry.element?.tagName
                });
            }
        });

        // First Input Delay (FID)
        this.observePerformanceEntry('first-input', (entries) => {
            entries.forEach(entry => {
                this.recordMetric('FID', entry.processingStart - entry.startTime, {
                    budget: this.budgets.FID,
                    critical: true,
                    eventType: entry.name
                });
            });
        });

        // Cumulative Layout Shift (CLS)
        let clsValue = 0;
        let clsEntries = [];
        this.observePerformanceEntry('layout-shift', (entries) => {
            entries.forEach(entry => {
                if (!entry.hadRecentInput) {
                    clsValue += entry.value;
                    clsEntries.push(entry);
                }
            });
            
            this.recordMetric('CLS', clsValue, {
                budget: this.budgets.CLS,
                critical: true,
                entries: clsEntries.length
            });
        });

        // Time to First Byte (TTFB)
        if (window.performance && window.performance.timing) {
            const ttfb = window.performance.timing.responseStart - window.performance.timing.requestStart;
            this.recordMetric('TTFB', ttfb, {
                budget: this.budgets.TTFB,
                critical: true
            });
        }
    }

    /**
     * Initialize Real User Monitoring
     */
    initRUM() {
        // Page load timing
        window.addEventListener('load', () => {
            setTimeout(() => {
                const timing = window.performance.timing;
                const loadTime = timing.loadEventEnd - timing.navigationStart;
                
                this.recordMetric('PageLoad', loadTime, {
                    type: 'navigation',
                    critical: true
                });

                // DNS lookup time
                const dnsTime = timing.domainLookupEnd - timing.domainLookupStart;
                this.recordMetric('DNSLookup', dnsTime, { type: 'network' });

                // TCP connection time
                const tcpTime = timing.connectEnd - timing.connectStart;
                this.recordMetric('TCPConnection', tcpTime, { type: 'network' });

                // SSL handshake time
                if (timing.secureConnectionStart > 0) {
                    const sslTime = timing.connectEnd - timing.secureConnectionStart;
                    this.recordMetric('SSLHandshake', sslTime, { type: 'network' });
                }

                // DOM processing time
                const domTime = timing.domComplete - timing.domLoading;
                this.recordMetric('DOMProcessing', domTime, { type: 'rendering' });

            }, 0);
        });

        // User interactions
        ['click', 'keydown', 'scroll'].forEach(eventType => {
            document.addEventListener(eventType, (event) => {
                const startTime = performance.now();
                
                requestAnimationFrame(() => {
                    const duration = performance.now() - startTime;
                    this.recordMetric('UserInteraction', duration, {
                        type: eventType,
                        target: event.target?.tagName
                    });
                });
            }, { passive: true });
        });

        // Viewport changes
        if ('ResizeObserver' in window) {
            const resizeObserver = new ResizeObserver((entries) => {
                entries.forEach(entry => {
                    this.recordMetric('ViewportChange', performance.now(), {
                        width: entry.contentRect.width,
                        height: entry.contentRect.height
                    });
                });
            });
            resizeObserver.observe(document.documentElement);
        }
    }

    /**
     * Initialize custom metrics for AI platform
     */
    initCustomMetrics() {
        // AI response time tracking
        window.addEventListener('ai-request-start', (event) => {
            const requestId = event.detail.requestId;
            this.startTimer(`ai-request-${requestId}`);
        });

        window.addEventListener('ai-request-end', (event) => {
            const requestId = event.detail.requestId;
            const duration = this.endTimer(`ai-request-${requestId}`);
            
            this.recordMetric('AIResponseTime', duration, {
                requestId,
                type: event.detail.type,
                model: event.detail.model,
                critical: true
            });
        });

        // Localization loading time
        window.addEventListener('localization-start', (event) => {
            this.startTimer(`localization-${event.detail.locale}`);
        });

        window.addEventListener('localization-end', (event) => {
            const duration = this.endTimer(`localization-${event.detail.locale}`);
            
            this.recordMetric('LocalizationLoadTime', duration, {
                locale: event.detail.locale,
                stringsCount: event.detail.stringsCount
            });
        });

        // Compliance check time
        window.addEventListener('compliance-check-start', (event) => {
            this.startTimer(`compliance-${event.detail.type}`);
        });

        window.addEventListener('compliance-check-end', (event) => {
            const duration = this.endTimer(`compliance-${event.detail.type}`);
            
            this.recordMetric('ComplianceCheckTime', duration, {
                type: event.detail.type,
                region: event.detail.region
            });
        });

        // Memory usage monitoring
        if ('memory' in performance) {
            setInterval(() => {
                const memory = performance.memory;
                this.recordMetric('MemoryUsage', memory.usedJSHeapSize, {
                    total: memory.totalJSHeapSize,
                    limit: memory.jsHeapSizeLimit,
                    percentage: (memory.usedJSHeapSize / memory.totalJSHeapSize) * 100
                });
            }, 60000); // Every minute
        }
    }

    /**
     * Initialize navigation timing monitoring
     */
    initNavigationTiming() {
        this.observePerformanceEntry('navigation', (entries) => {
            entries.forEach(entry => {
                this.recordMetric('NavigationTiming', entry.duration, {
                    type: entry.type,
                    redirectCount: entry.redirectCount,
                    transferSize: entry.transferSize,
                    encodedBodySize: entry.encodedBodySize,
                    decodedBodySize: entry.decodedBodySize
                });
            });
        });
    }

    /**
     * Initialize resource timing monitoring
     */
    initResourceTiming() {
        this.observePerformanceEntry('resource', (entries) => {
            entries.forEach(entry => {
                // Only track significant resources
                if (entry.transferSize > 1024) { // > 1KB
                    this.recordMetric('ResourceTiming', entry.duration, {
                        name: entry.name,
                        type: this.getResourceType(entry.name),
                        size: entry.transferSize,
                        cached: entry.transferSize === 0
                    });
                }
            });
        });
    }

    /**
     * Initialize error tracking
     */
    initErrorTracking() {
        // JavaScript errors
        window.addEventListener('error', (event) => {
            this.recordMetric('JavaScriptError', 1, {
                message: event.message,
                filename: event.filename,
                lineno: event.lineno,
                colno: event.colno,
                stack: event.error?.stack
            });
        });

        // Unhandled promise rejections
        window.addEventListener('unhandledrejection', (event) => {
            this.recordMetric('UnhandledRejection', 1, {
                reason: event.reason?.toString(),
                stack: event.reason?.stack
            });
        });

        // Resource loading errors
        document.addEventListener('error', (event) => {
            if (event.target !== window) {
                this.recordMetric('ResourceError', 1, {
                    type: event.target.tagName,
                    source: event.target.src || event.target.href,
                    message: 'Failed to load resource'
                });
            }
        }, true);
    }

    /**
     * Observe performance entries
     */
    observePerformanceEntry(type, callback) {
        if ('PerformanceObserver' in window) {
            try {
                const observer = new PerformanceObserver((list) => {
                    callback(list.getEntries());
                });
                
                observer.observe({ type, buffered: true });
                this.observers.set(type, observer);
                
            } catch (error) {
                console.warn(`Failed to observe ${type}:`, error);
            }
        }
    }

    /**
     * Record a performance metric
     */
    recordMetric(name, value, metadata = {}) {
        // Apply sampling
        if (Math.random() > this.config.sampleRate) {
            return;
        }

        const metric = {
            name,
            value,
            timestamp: Date.now(),
            region: this.config.region,
            url: window.location.href,
            userAgent: navigator.userAgent,
            connection: this.getConnectionInfo(),
            ...metadata
        };

        // Check against budget
        if (metadata.budget && metadata.critical) {
            const exceedsBudget = value > metadata.budget;
            metric.exceedsBudget = exceedsBudget;
            
            if (exceedsBudget) {
                console.warn(`⚠️ Performance budget exceeded: ${name} = ${value}ms (budget: ${metadata.budget}ms)`);
                this.triggerAlert(metric);
            }
        }

        this.metrics.push(metric);

        // Flush if buffer is full
        if (this.metrics.length >= this.config.bufferSize) {
            this.flush();
        }
    }

    /**
     * Start a timer
     */
    startTimer(id) {
        this.timers = this.timers || new Map();
        this.timers.set(id, performance.now());
    }

    /**
     * End a timer and return duration
     */
    endTimer(id) {
        this.timers = this.timers || new Map();
        const startTime = this.timers.get(id);
        
        if (startTime) {
            this.timers.delete(id);
            return performance.now() - startTime;
        }
        
        return 0;
    }

    /**
     * Get connection information
     */
    getConnectionInfo() {
        if ('connection' in navigator) {
            const conn = navigator.connection;
            return {
                effectiveType: conn.effectiveType,
                downlink: conn.downlink,
                rtt: conn.rtt,
                saveData: conn.saveData
            };
        }
        return {};
    }

    /**
     * Get resource type from URL
     */
    getResourceType(url) {
        const extension = url.split('.').pop()?.toLowerCase();
        
        if (['js', 'mjs'].includes(extension)) return 'script';
        if (['css'].includes(extension)) return 'stylesheet';
        if (['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'avif'].includes(extension)) return 'image';
        if (['woff', 'woff2', 'ttf', 'otf'].includes(extension)) return 'font';
        if (['mp4', 'webm', 'ogg'].includes(extension)) return 'video';
        if (['mp3', 'wav', 'ogg'].includes(extension)) return 'audio';
        
        return 'other';
    }

    /**
     * Trigger performance alert
     */
    triggerAlert(metric) {
        // Send immediate alert for critical performance issues
        fetch(this.config.apiEndpoint + '/alert', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                type: 'performance_budget_exceeded',
                metric,
                severity: 'warning',
                timestamp: Date.now()
            })
        }).catch(error => {
            console.error('Failed to send performance alert:', error);
        });
    }

    /**
     * Flush metrics to server
     */
    async flush() {
        if (this.metrics.length === 0) {
            return;
        }

        const metricsToSend = [...this.metrics];
        this.metrics = [];

        try {
            await fetch(this.config.apiEndpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    metrics: metricsToSend,
                    region: this.config.region,
                    timestamp: Date.now()
                })
            });

            console.log(`📊 Sent ${metricsToSend.length} performance metrics`);

        } catch (error) {
            console.error('Failed to send performance metrics:', error);
            // Re-add metrics to buffer for retry
            this.metrics.unshift(...metricsToSend);
        }
    }

    /**
     * Start periodic flushing
     */
    startPeriodicFlush() {
        setInterval(() => {
            this.flush();
        }, this.config.flushInterval);

        // Flush on page unload
        window.addEventListener('beforeunload', () => {
            if (this.metrics.length > 0) {
                // Use sendBeacon for reliable delivery
                if ('sendBeacon' in navigator) {
                    navigator.sendBeacon(
                        this.config.apiEndpoint,
                        JSON.stringify({
                            metrics: this.metrics,
                            region: this.config.region,
                            timestamp: Date.now()
                        })
                    );
                }
            }
        });
    }

    /**
     * Get current performance summary
     */
    getPerformanceSummary() {
        const summary = {
            region: this.config.region,
            timestamp: Date.now(),
            metrics: {},
            budgetViolations: 0
        };

        // Group metrics by name
        const groupedMetrics = this.metrics.reduce((acc, metric) => {
            if (!acc[metric.name]) {
                acc[metric.name] = [];
            }
            acc[metric.name].push(metric);
            return acc;
        }, {});

        // Calculate statistics for each metric
        Object.entries(groupedMetrics).forEach(([name, metrics]) => {
            const values = metrics.map(m => m.value);
            const budgetViolations = metrics.filter(m => m.exceedsBudget).length;
            
            summary.metrics[name] = {
                count: values.length,
                min: Math.min(...values),
                max: Math.max(...values),
                avg: values.reduce((a, b) => a + b, 0) / values.length,
                p95: this.percentile(values, 95),
                budgetViolations
            };
            
            summary.budgetViolations += budgetViolations;
        });

        return summary;
    }

    /**
     * Calculate percentile
     */
    percentile(values, p) {
        const sorted = values.sort((a, b) => a - b);
        const index = Math.ceil((p / 100) * sorted.length) - 1;
        return sorted[index];
    }

    /**
     * Destroy performance monitor
     */
    destroy() {
        // Disconnect all observers
        this.observers.forEach(observer => {
            observer.disconnect();
        });
        this.observers.clear();

        // Final flush
        this.flush();

        this.isInitialized = false;
        console.log('🛑 Performance monitoring destroyed');
    }
}

// Export for use in applications
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PerformanceMonitor;
} else if (typeof window !== 'undefined') {
    window.PerformanceMonitor = PerformanceMonitor;
}

// Auto-initialize if config is available
if (typeof window !== 'undefined' && window.PERFORMANCE_CONFIG) {
    const monitor = new PerformanceMonitor(window.PERFORMANCE_CONFIG);
    window.performanceMonitor = monitor;
}