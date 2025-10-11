/**
 * Bundle Analyzer for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive bundle analysis including:
 * - Bundle size analysis and reporting
 * - Dependency tree analysis
 * - Code splitting recommendations
 * - Unused code detection
 * - Performance impact assessment
 * - Regional optimization suggestions
 * - Tree shaking effectiveness
 */

class BundleAnalyzer {
    constructor(config = {}) {
        this.config = {
            // Analysis settings
            enableRuntimeAnalysis: config.enableRuntimeAnalysis !== false,
            enableStaticAnalysis: config.enableStaticAnalysis !== false,
            enableDependencyAnalysis: config.enableDependencyAnalysis !== false,
            
            // Thresholds
            maxBundleSize: config.maxBundleSize || 250 * 1024, // 250KB
            maxChunkSize: config.maxChunkSize || 100 * 1024, // 100KB
            maxDependencySize: config.maxDependencySize || 50 * 1024, // 50KB
            
            // Performance budgets
            budgets: {
                initial: config.budgets?.initial || 200 * 1024, // 200KB
                async: config.budgets?.async || 100 * 1024, // 100KB
                total: config.budgets?.total || 500 * 1024 // 500KB
            },
            
            // Regional settings
            region: config.region || 'us-east-1',
            enableRegionalOptimization: config.enableRegionalOptimization !== false,
            
            // Reporting
            enableReporting: config.enableReporting !== false,
            reportInterval: config.reportInterval || 300000, // 5 minutes
            apiEndpoint: config.apiEndpoint || '/api/performance/bundle-analysis',
            
            ...config
        };

        this.bundleData = {
            chunks: new Map(),
            dependencies: new Map(),
            modules: new Map(),
            assets: new Map()
        };

        this.analysisResults = {
            totalSize: 0,
            gzippedSize: 0,
            chunkCount: 0,
            moduleCount: 0,
            duplicates: [],
            unusedCode: [],
            recommendations: [],
            performanceImpact: {}
        };

        this.performanceObserver = null;
        this.resourceObserver = null;
        
        this.init();
    }

    /**
     * Initialize bundle analyzer
     */
    async init() {
        try {
            // Initialize runtime analysis
            if (this.config.enableRuntimeAnalysis) {
                await this.initRuntimeAnalysis();
            }

            // Initialize static analysis
            if (this.config.enableStaticAnalysis) {
                await this.initStaticAnalysis();
            }

            // Initialize dependency analysis
            if (this.config.enableDependencyAnalysis) {
                await this.initDependencyAnalysis();
            }

            // Start performance monitoring
            this.startPerformanceMonitoring();

            // Start periodic reporting
            if (this.config.enableReporting) {
                this.startPeriodicReporting();
            }

            console.log('📊 Bundle analyzer initialized');

        } catch (error) {
            console.error('❌ Failed to initialize bundle analyzer:', error);
        }
    }

    /**
     * Initialize runtime analysis
     */
    async initRuntimeAnalysis() {
        // Analyze loaded scripts
        this.analyzeLoadedScripts();

        // Monitor resource loading
        this.monitorResourceLoading();

        // Analyze module dependencies
        this.analyzeModuleDependencies();

        // Check for code splitting opportunities
        this.analyzeCodeSplitting();
    }

    /**
     * Initialize static analysis
     */
    async initStaticAnalysis() {
        // This would typically be done at build time
        // Here we simulate with available runtime data
        
        try {
            // Fetch webpack stats if available
            const stats = await this.fetchWebpackStats();
            if (stats) {
                this.analyzeWebpackStats(stats);
            }

            // Analyze source maps if available
            await this.analyzeSourceMaps();

        } catch (error) {
            console.warn('Static analysis data not available:', error);
        }
    }

    /**
     * Initialize dependency analysis
     */
    async initDependencyAnalysis() {
        // Analyze package.json dependencies
        await this.analyzeDependencies();

        // Check for duplicate dependencies
        this.findDuplicateDependencies();

        // Analyze dependency sizes
        this.analyzeDependencySizes();

        // Check for unused dependencies
        this.findUnusedDependencies();
    }

    /**
     * Analyze loaded scripts
     */
    analyzeLoadedScripts() {
        const scripts = document.querySelectorAll('script[src]');
        
        scripts.forEach(script => {
            const src = script.src;
            const size = this.estimateScriptSize(script);
            
            this.bundleData.assets.set(src, {
                type: 'script',
                size,
                async: script.async,
                defer: script.defer,
                module: script.type === 'module',
                crossOrigin: script.crossOrigin,
                integrity: script.integrity
            });
        });

        // Analyze CSS files
        const stylesheets = document.querySelectorAll('link[rel="stylesheet"]');
        
        stylesheets.forEach(link => {
            const href = link.href;
            const size = this.estimateStylesheetSize(link);
            
            this.bundleData.assets.set(href, {
                type: 'stylesheet',
                size,
                media: link.media,
                crossOrigin: link.crossOrigin,
                integrity: link.integrity
            });
        });
    }

    /**
     * Monitor resource loading
     */
    monitorResourceLoading() {
        if ('PerformanceObserver' in window) {
            this.resourceObserver = new PerformanceObserver((list) => {
                list.getEntries().forEach(entry => {
                    if (entry.initiatorType === 'script' || entry.initiatorType === 'link') {
                        this.analyzeResourceTiming(entry);
                    }
                });
            });

            this.resourceObserver.observe({ type: 'resource', buffered: true });
        }
    }

    /**
     * Analyze module dependencies
     */
    analyzeModuleDependencies() {
        // Check if webpack is available
        if (typeof __webpack_require__ !== 'undefined') {
            this.analyzeWebpackModules();
        }

        // Analyze ES modules
        this.analyzeESModules();

        // Check for dynamic imports
        this.analyzeDynamicImports();
    }

    /**
     * Analyze code splitting opportunities
     */
    analyzeCodeSplitting() {
        const recommendations = [];

        // Check for large vendor bundles
        const vendorSize = this.calculateVendorBundleSize();
        if (vendorSize > this.config.maxBundleSize) {
            recommendations.push({
                type: 'code-splitting',
                priority: 'high',
                description: 'Large vendor bundle detected',
                suggestion: 'Split vendor dependencies into separate chunks',
                impact: 'Improved caching and parallel loading',
                estimatedSavings: vendorSize * 0.3
            });
        }

        // Check for route-based splitting opportunities
        const routeAnalysis = this.analyzeRoutes();
        if (routeAnalysis.canSplit) {
            recommendations.push({
                type: 'route-splitting',
                priority: 'medium',
                description: 'Route-based code splitting opportunity',
                suggestion: 'Implement lazy loading for route components',
                impact: 'Reduced initial bundle size',
                estimatedSavings: routeAnalysis.potentialSavings
            });
        }

        // Check for feature-based splitting
        const featureAnalysis = this.analyzeFeatures();
        featureAnalysis.forEach(feature => {
            if (feature.size > this.config.maxChunkSize) {
                recommendations.push({
                    type: 'feature-splitting',
                    priority: 'medium',
                    description: `Large feature module: ${feature.name}`,
                    suggestion: 'Split feature into separate chunk',
                    impact: 'Improved loading performance',
                    estimatedSavings: feature.size * 0.4
                });
            }
        });

        this.analysisResults.recommendations.push(...recommendations);
    }

    /**
     * Analyze webpack modules
     */
    analyzeWebpackModules() {
        if (typeof __webpack_require__ === 'undefined') return;

        try {
            // Get webpack module cache
            const moduleCache = __webpack_require__.cache;
            
            Object.keys(moduleCache).forEach(moduleId => {
                const module = moduleCache[moduleId];
                if (module && module.exports) {
                    this.bundleData.modules.set(moduleId, {
                        id: moduleId,
                        size: this.estimateModuleSize(module),
                        dependencies: this.getModuleDependencies(module),
                        exports: Object.keys(module.exports || {}),
                        loaded: module.loaded,
                        parent: module.parent
                    });
                }
            });

        } catch (error) {
            console.warn('Failed to analyze webpack modules:', error);
        }
    }

    /**
     * Analyze ES modules
     */
    analyzeESModules() {
        // This is limited in runtime analysis
        // Would typically be done at build time with AST parsing
        
        const scripts = document.querySelectorAll('script[type="module"]');
        scripts.forEach(script => {
            if (script.src) {
                this.bundleData.modules.set(script.src, {
                    type: 'es-module',
                    src: script.src,
                    async: true,
                    defer: true
                });
            }
        });
    }

    /**
     * Analyze dynamic imports
     */
    analyzeDynamicImports() {
        // Monitor dynamic imports through performance API
        if ('PerformanceObserver' in window) {
            const observer = new PerformanceObserver((list) => {
                list.getEntries().forEach(entry => {
                    if (entry.name.includes('chunk') || entry.name.includes('lazy')) {
                        this.bundleData.chunks.set(entry.name, {
                            name: entry.name,
                            size: entry.transferSize,
                            loadTime: entry.duration,
                            dynamic: true
                        });
                    }
                });
            });

            observer.observe({ type: 'resource', buffered: true });
        }
    }

    /**
     * Fetch webpack stats
     */
    async fetchWebpackStats() {
        try {
            const response = await fetch('/webpack-stats.json');
            if (response.ok) {
                return await response.json();
            }
        } catch (error) {
            // Stats not available
        }
        return null;
    }

    /**
     * Analyze webpack stats
     */
    analyzeWebpackStats(stats) {
        if (!stats.chunks) return;

        stats.chunks.forEach(chunk => {
            this.bundleData.chunks.set(chunk.id, {
                id: chunk.id,
                names: chunk.names,
                size: chunk.size,
                modules: chunk.modules?.length || 0,
                entry: chunk.entry,
                initial: chunk.initial,
                files: chunk.files
            });
        });

        // Analyze modules
        if (stats.modules) {
            stats.modules.forEach(module => {
                this.bundleData.modules.set(module.id, {
                    id: module.id,
                    name: module.name,
                    size: module.size,
                    chunks: module.chunks,
                    dependencies: module.dependencies?.length || 0,
                    source: module.source
                });
            });
        }

        // Calculate totals
        this.analysisResults.totalSize = stats.assets?.reduce((total, asset) => total + asset.size, 0) || 0;
        this.analysisResults.chunkCount = stats.chunks?.length || 0;
        this.analysisResults.moduleCount = stats.modules?.length || 0;
    }

    /**
     * Analyze source maps
     */
    async analyzeSourceMaps() {
        const scripts = document.querySelectorAll('script[src]');
        
        for (const script of scripts) {
            try {
                const response = await fetch(script.src);
                const content = await response.text();
                
                // Check for source map reference
                const sourceMapMatch = content.match(/\/\/# sourceMappingURL=(.+)/);
                if (sourceMapMatch) {
                    const sourceMapUrl = new URL(sourceMapMatch[1], script.src);
                    await this.analyzeSourceMap(sourceMapUrl.href, script.src);
                }
            } catch (error) {
                // Source map not available
            }
        }
    }

    /**
     * Analyze individual source map
     */
    async analyzeSourceMap(sourceMapUrl, scriptUrl) {
        try {
            const response = await fetch(sourceMapUrl);
            const sourceMap = await response.json();
            
            if (sourceMap.sources) {
                sourceMap.sources.forEach(source => {
                    if (!this.bundleData.modules.has(source)) {
                        this.bundleData.modules.set(source, {
                            source,
                            bundledIn: scriptUrl,
                            type: 'source-mapped'
                        });
                    }
                });
            }
        } catch (error) {
            console.warn('Failed to analyze source map:', error);
        }
    }

    /**
     * Analyze dependencies
     */
    async analyzeDependencies() {
        try {
            // This would typically read from package.json at build time
            // Here we simulate with known dependencies
            const dependencies = this.getKnownDependencies();
            
            dependencies.forEach(dep => {
                this.bundleData.dependencies.set(dep.name, {
                    name: dep.name,
                    version: dep.version,
                    size: dep.size,
                    type: dep.type, // 'production' or 'development'
                    used: this.isDependencyUsed(dep.name)
                });
            });

        } catch (error) {
            console.warn('Failed to analyze dependencies:', error);
        }
    }

    /**
     * Find duplicate dependencies
     */
    findDuplicateDependencies() {
        const duplicates = [];
        const dependencyVersions = new Map();

        this.bundleData.dependencies.forEach((dep, name) => {
            const baseName = name.split('@')[0];
            
            if (!dependencyVersions.has(baseName)) {
                dependencyVersions.set(baseName, []);
            }
            
            dependencyVersions.get(baseName).push(dep);
        });

        dependencyVersions.forEach((versions, name) => {
            if (versions.length > 1) {
                duplicates.push({
                    name,
                    versions: versions.map(v => v.version),
                    totalSize: versions.reduce((sum, v) => sum + v.size, 0),
                    potentialSavings: versions.slice(1).reduce((sum, v) => sum + v.size, 0)
                });
            }
        });

        this.analysisResults.duplicates = duplicates;
    }

    /**
     * Analyze dependency sizes
     */
    analyzeDependencySizes() {
        const largeDependencies = [];

        this.bundleData.dependencies.forEach(dep => {
            if (dep.size > this.config.maxDependencySize) {
                largeDependencies.push({
                    name: dep.name,
                    size: dep.size,
                    recommendation: this.getDependencyRecommendation(dep)
                });
            }
        });

        if (largeDependencies.length > 0) {
            this.analysisResults.recommendations.push({
                type: 'dependency-optimization',
                priority: 'medium',
                description: 'Large dependencies detected',
                dependencies: largeDependencies,
                suggestion: 'Consider alternatives or tree shaking'
            });
        }
    }

    /**
     * Find unused dependencies
     */
    findUnusedDependencies() {
        const unused = [];

        this.bundleData.dependencies.forEach(dep => {
            if (!dep.used && dep.type === 'production') {
                unused.push(dep.name);
            }
        });

        if (unused.length > 0) {
            this.analysisResults.unusedCode.push({
                type: 'unused-dependencies',
                items: unused,
                potentialSavings: unused.reduce((sum, name) => {
                    const dep = this.bundleData.dependencies.get(name);
                    return sum + (dep?.size || 0);
                }, 0)
            });
        }
    }

    /**
     * Analyze resource timing
     */
    analyzeResourceTiming(entry) {
        const asset = this.bundleData.assets.get(entry.name);
        if (asset) {
            asset.timing = {
                duration: entry.duration,
                transferSize: entry.transferSize,
                encodedBodySize: entry.encodedBodySize,
                decodedBodySize: entry.decodedBodySize,
                domainLookupTime: entry.domainLookupEnd - entry.domainLookupStart,
                connectTime: entry.connectEnd - entry.connectStart,
                requestTime: entry.responseStart - entry.requestStart,
                responseTime: entry.responseEnd - entry.responseStart
            };

            // Check for performance issues
            if (entry.duration > 1000) { // > 1 second
                this.analysisResults.recommendations.push({
                    type: 'slow-resource',
                    priority: 'high',
                    resource: entry.name,
                    duration: entry.duration,
                    suggestion: 'Optimize resource loading or consider CDN'
                });
            }
        }
    }

    /**
     * Calculate vendor bundle size
     */
    calculateVendorBundleSize() {
        let vendorSize = 0;
        
        this.bundleData.chunks.forEach(chunk => {
            if (chunk.names?.includes('vendor') || chunk.names?.includes('vendors')) {
                vendorSize += chunk.size;
            }
        });

        return vendorSize;
    }

    /**
     * Analyze routes for splitting opportunities
     */
    analyzeRoutes() {
        // This would typically analyze router configuration
        // Here we provide a simplified analysis
        
        const routeModules = Array.from(this.bundleData.modules.values())
            .filter(module => module.name?.includes('route') || module.name?.includes('page'));

        const totalRouteSize = routeModules.reduce((sum, module) => sum + module.size, 0);
        
        return {
            canSplit: routeModules.length > 3 && totalRouteSize > this.config.maxBundleSize,
            potentialSavings: totalRouteSize * 0.6, // Estimate 60% can be lazy loaded
            routeCount: routeModules.length
        };
    }

    /**
     * Analyze features for splitting opportunities
     */
    analyzeFeatures() {
        const features = [];
        
        // Group modules by feature (simplified heuristic)
        const featureGroups = new Map();
        
        this.bundleData.modules.forEach(module => {
            const featureName = this.extractFeatureName(module.name);
            if (featureName) {
                if (!featureGroups.has(featureName)) {
                    featureGroups.set(featureName, []);
                }
                featureGroups.get(featureName).push(module);
            }
        });

        featureGroups.forEach((modules, featureName) => {
            const totalSize = modules.reduce((sum, module) => sum + module.size, 0);
            features.push({
                name: featureName,
                size: totalSize,
                moduleCount: modules.length
            });
        });

        return features;
    }

    /**
     * Start performance monitoring
     */
    startPerformanceMonitoring() {
        if ('PerformanceObserver' in window) {
            this.performanceObserver = new PerformanceObserver((list) => {
                list.getEntries().forEach(entry => {
                    this.analyzePerformanceEntry(entry);
                });
            });

            this.performanceObserver.observe({ 
                type: 'navigation', 
                buffered: true 
            });
        }
    }

    /**
     * Analyze performance entry
     */
    analyzePerformanceEntry(entry) {
        if (entry.entryType === 'navigation') {
            this.analysisResults.performanceImpact = {
                domContentLoaded: entry.domContentLoadedEventEnd - entry.domContentLoadedEventStart,
                loadComplete: entry.loadEventEnd - entry.loadEventStart,
                totalLoadTime: entry.loadEventEnd - entry.fetchStart,
                resourceCount: performance.getEntriesByType('resource').length
            };
        }
    }

    /**
     * Start periodic reporting
     */
    startPeriodicReporting() {
        setInterval(() => {
            this.generateReport();
        }, this.config.reportInterval);

        // Report on page unload
        window.addEventListener('beforeunload', () => {
            this.generateReport(true);
        });
    }

    /**
     * Generate comprehensive analysis report
     */
    generateReport(isUnload = false) {
        const report = {
            timestamp: Date.now(),
            region: this.config.region,
            bundleAnalysis: {
                totalSize: this.analysisResults.totalSize,
                gzippedSize: this.analysisResults.gzippedSize,
                chunkCount: this.analysisResults.chunkCount,
                moduleCount: this.analysisResults.moduleCount,
                assetCount: this.bundleData.assets.size,
                dependencyCount: this.bundleData.dependencies.size
            },
            performanceImpact: this.analysisResults.performanceImpact,
            budgetStatus: this.checkBudgets(),
            recommendations: this.analysisResults.recommendations,
            duplicates: this.analysisResults.duplicates,
            unusedCode: this.analysisResults.unusedCode,
            optimizationOpportunities: this.getOptimizationOpportunities()
        };

        // Send report to server
        if (this.config.enableReporting) {
            this.sendReport(report, isUnload);
        }

        return report;
    }

    /**
     * Check budget compliance
     */
    checkBudgets() {
        const status = {
            initial: { budget: this.config.budgets.initial, actual: 0, compliant: true },
            async: { budget: this.config.budgets.async, actual: 0, compliant: true },
            total: { budget: this.config.budgets.total, actual: 0, compliant: true }
        };

        // Calculate actual sizes
        this.bundleData.chunks.forEach(chunk => {
            if (chunk.initial) {
                status.initial.actual += chunk.size;
            } else {
                status.async.actual += chunk.size;
            }
            status.total.actual += chunk.size;
        });

        // Check compliance
        status.initial.compliant = status.initial.actual <= status.initial.budget;
        status.async.compliant = status.async.actual <= status.async.budget;
        status.total.compliant = status.total.actual <= status.total.budget;

        return status;
    }

    /**
     * Get optimization opportunities
     */
    getOptimizationOpportunities() {
        const opportunities = [];

        // Tree shaking opportunities
        const unusedExports = this.findUnusedExports();
        if (unusedExports.length > 0) {
            opportunities.push({
                type: 'tree-shaking',
                description: 'Unused exports detected',
                impact: 'high',
                savings: unusedExports.reduce((sum, exp) => sum + exp.size, 0)
            });
        }

        // Compression opportunities
        const compressionSavings = this.calculateCompressionSavings();
        if (compressionSavings > 10 * 1024) { // > 10KB
            opportunities.push({
                type: 'compression',
                description: 'Additional compression possible',
                impact: 'medium',
                savings: compressionSavings
            });
        }

        // Minification opportunities
        const minificationSavings = this.calculateMinificationSavings();
        if (minificationSavings > 5 * 1024) { // > 5KB
            opportunities.push({
                type: 'minification',
                description: 'Additional minification possible',
                impact: 'low',
                savings: minificationSavings
            });
        }

        return opportunities;
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
            console.error('Failed to send bundle analysis report:', error);
        }
    }

    /**
     * Utility methods
     */
    estimateScriptSize(script) {
        // This is an estimation - actual size would come from network timing
        return script.textContent?.length || 0;
    }

    estimateStylesheetSize(link) {
        // This is an estimation - actual size would come from network timing
        return 0; // Would need to fetch to get actual size
    }

    estimateModuleSize(module) {
        // Rough estimation based on exports and source
        return JSON.stringify(module.exports || {}).length;
    }

    getModuleDependencies(module) {
        // Extract dependencies from webpack module
        return module.dependencies?.map(dep => dep.module) || [];
    }

    getKnownDependencies() {
        // This would typically come from package.json analysis
        return [
            { name: 'react', version: '18.0.0', size: 42 * 1024, type: 'production' },
            { name: 'react-dom', version: '18.0.0', size: 130 * 1024, type: 'production' },
            { name: 'lodash', version: '4.17.21', size: 70 * 1024, type: 'production' }
        ];
    }

    isDependencyUsed(name) {
        // Check if dependency is actually used in the bundle
        return Array.from(this.bundleData.modules.values())
            .some(module => module.name?.includes(name));
    }

    getDependencyRecommendation(dep) {
        const recommendations = {
            'lodash': 'Consider using lodash-es for better tree shaking',
            'moment': 'Consider using date-fns or dayjs for smaller bundle size',
            'axios': 'Consider using fetch API for simple requests'
        };
        
        return recommendations[dep.name] || 'Consider if this dependency is necessary';
    }

    extractFeatureName(moduleName) {
        if (!moduleName) return null;
        
        // Simple heuristic to extract feature name
        const parts = moduleName.split('/');
        return parts.find(part => 
            part.includes('feature') || 
            part.includes('component') || 
            part.includes('page')
        );
    }

    findUnusedExports() {
        // This would require static analysis of the codebase
        // Simplified implementation
        return [];
    }

    calculateCompressionSavings() {
        // Estimate additional compression savings
        return this.analysisResults.totalSize * 0.1; // 10% estimate
    }

    calculateMinificationSavings() {
        // Estimate additional minification savings
        return this.analysisResults.totalSize * 0.05; // 5% estimate
    }

    /**
     * Get analysis summary
     */
    getSummary() {
        return {
            totalSize: this.analysisResults.totalSize,
            chunkCount: this.analysisResults.chunkCount,
            moduleCount: this.analysisResults.moduleCount,
            recommendationCount: this.analysisResults.recommendations.length,
            duplicateCount: this.analysisResults.duplicates.length,
            unusedCodeCount: this.analysisResults.unusedCode.length,
            budgetCompliance: this.checkBudgets()
        };
    }

    /**
     * Destroy bundle analyzer
     */
    destroy() {
        if (this.performanceObserver) {
            this.performanceObserver.disconnect();
        }
        
        if (this.resourceObserver) {
            this.resourceObserver.disconnect();
        }

        this.bundleData.chunks.clear();
        this.bundleData.dependencies.clear();
        this.bundleData.modules.clear();
        this.bundleData.assets.clear();

        console.log('🛑 Bundle analyzer destroyed');
    }
}

// Export for use in applications
if (typeof module !== 'undefined' && module.exports) {
    module.exports = BundleAnalyzer;
} else if (typeof window !== 'undefined') {
    window.BundleAnalyzer = BundleAnalyzer;
}

// Auto-initialize if config is available
if (typeof window !== 'undefined' && window.BUNDLE_ANALYZER_CONFIG) {
    const bundleAnalyzer = new BundleAnalyzer(window.BUNDLE_ANALYZER_CONFIG);
    window.bundleAnalyzer = bundleAnalyzer;
}