import { test, expect } from '@playwright/test';

/**
 * Performance E2E Tests
 * Tests application performance across different regions and network conditions
 */
test.describe('Performance Tests @performance @global', () => {
  
  test.beforeEach(async ({ page }) => {
    // Set up performance monitoring
    await page.addInitScript(() => {
      window.performanceMetrics = {
        navigationStart: performance.timing.navigationStart,
        loadEventEnd: performance.timing.loadEventEnd,
        domContentLoaded: performance.timing.domContentLoadedEventEnd
      };
    });
  });

  test('should load homepage within acceptable time limits', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/');
    
    // Wait for page to be fully loaded
    await page.waitForLoadState('networkidle');
    
    const loadTime = Date.now() - startTime;
    
    // Homepage should load within 3 seconds
    expect(loadTime).toBeLessThan(3000);
    
    // Check Core Web Vitals
    const metrics = await page.evaluate(() => {
      return new Promise((resolve) => {
        new PerformanceObserver((list) => {
          const entries = list.getEntries();
          const vitals = {};
          
          entries.forEach((entry) => {
            if (entry.name === 'first-contentful-paint') {
              vitals.fcp = entry.startTime;
            }
            if (entry.name === 'largest-contentful-paint') {
              vitals.lcp = entry.startTime;
            }
          });
          
          resolve(vitals);
        }).observe({ entryTypes: ['paint', 'largest-contentful-paint'] });
        
        // Fallback timeout
        setTimeout(() => resolve({}), 5000);
      });
    });
    
    // First Contentful Paint should be under 1.8s
    if (metrics.fcp) {
      expect(metrics.fcp).toBeLessThan(1800);
    }
    
    // Largest Contentful Paint should be under 2.5s
    if (metrics.lcp) {
      expect(metrics.lcp).toBeLessThan(2500);
    }
  });

  test('should handle slow network conditions gracefully', async ({ page, context }) => {
    // Simulate slow 3G network
    await context.route('**/*', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 100)); // Add 100ms delay
      await route.continue();
    });
    
    const startTime = Date.now();
    await page.goto('/');
    
    // Should show loading indicators
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible();
    
    // Wait for content to load
    await page.waitForLoadState('networkidle');
    
    const loadTime = Date.now() - startTime;
    
    // Should still load within reasonable time even on slow network
    expect(loadTime).toBeLessThan(10000);
    
    // Loading indicator should disappear
    await expect(page.locator('[data-testid="loading-indicator"]')).not.toBeVisible();
  });

  test('should optimize image loading and display', async ({ page }) => {
    await page.goto('/gallery');
    
    // Check if images are lazy loaded
    const images = page.locator('img[loading="lazy"]');
    const imageCount = await images.count();
    
    // Should have lazy loading enabled for non-critical images
    expect(imageCount).toBeGreaterThan(0);
    
    // Check if images have proper alt text
    const imagesWithAlt = page.locator('img[alt]');
    const altCount = await imagesWithAlt.count();
    
    expect(altCount).toEqual(imageCount);
    
    // Check if WebP format is used when supported
    const webpImages = page.locator('img[src*=".webp"], source[srcset*=".webp"]');
    const webpCount = await webpImages.count();
    
    // Should use modern image formats
    expect(webpCount).toBeGreaterThan(0);
  });

  test('should implement efficient caching strategies', async ({ page }) => {
    // First visit
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check if static assets are cached
    const responses = [];
    page.on('response', response => {
      if (response.url().includes('.js') || response.url().includes('.css')) {
        responses.push({
          url: response.url(),
          status: response.status(),
          headers: response.headers()
        });
      }
    });
    
    // Second visit (should use cache)
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    // Check cache headers
    const cachedResponses = responses.filter(r => 
      r.headers['cache-control'] && 
      (r.headers['cache-control'].includes('max-age') || r.status === 304)
    );
    
    expect(cachedResponses.length).toBeGreaterThan(0);
  });

  test('should handle large datasets efficiently', async ({ page }) => {
    // Navigate to page with large dataset
    await page.goto('/analytics/large-dataset');
    
    // Should implement virtual scrolling or pagination
    const virtualScroll = page.locator('[data-testid="virtual-scroll"]');
    const pagination = page.locator('[data-testid="pagination"]');
    
    const hasVirtualScroll = await virtualScroll.count() > 0;
    const hasPagination = await pagination.count() > 0;
    
    // Should use either virtual scrolling or pagination for large datasets
    expect(hasVirtualScroll || hasPagination).toBeTruthy();
    
    // Check if only visible items are rendered
    if (hasVirtualScroll) {
      const renderedItems = page.locator('[data-testid="data-item"]');
      const itemCount = await renderedItems.count();
      
      // Should not render all items at once
      expect(itemCount).toBeLessThan(100);
    }
  });

  test('should optimize bundle size and code splitting', async ({ page }) => {
    // Monitor network requests
    const jsRequests = [];
    
    page.on('response', response => {
      if (response.url().endsWith('.js')) {
        jsRequests.push({
          url: response.url(),
          size: parseInt(response.headers()['content-length'] || '0')
        });
      }
    });
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check initial bundle size
    const mainBundle = jsRequests.find(req => req.url.includes('main') || req.url.includes('index'));
    
    if (mainBundle) {
      // Main bundle should be under 500KB
      expect(mainBundle.size).toBeLessThan(500 * 1024);
    }
    
    // Navigate to different route to test code splitting
    await page.click('[data-testid="navigation-analytics"]');
    await page.waitForLoadState('networkidle');
    
    // Should load additional chunks for new route
    const totalRequests = jsRequests.length;
    expect(totalRequests).toBeGreaterThan(1);
  });

  test('should handle concurrent user interactions efficiently', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Simulate multiple rapid interactions
    const startTime = Date.now();
    
    // Click multiple buttons rapidly
    await Promise.all([
      page.click('[data-testid="refresh-button"]'),
      page.click('[data-testid="filter-button"]'),
      page.click('[data-testid="sort-button"]')
    ]);
    
    // Wait for all interactions to complete
    await page.waitForLoadState('networkidle');
    
    const responseTime = Date.now() - startTime;
    
    // Should handle concurrent interactions within 2 seconds
    expect(responseTime).toBeLessThan(2000);
    
    // UI should remain responsive
    await expect(page.locator('[data-testid="dashboard-content"]')).toBeVisible();
  });

  test('should optimize API response times', async ({ page }) => {
    const apiResponses = [];
    
    page.on('response', response => {
      if (response.url().includes('/api/')) {
        apiResponses.push({
          url: response.url(),
          status: response.status(),
          timing: response.timing()
        });
      }
    });
    
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Check API response times
    apiResponses.forEach(response => {
      // API responses should be under 1 second
      expect(response.timing.responseEnd - response.timing.requestStart).toBeLessThan(1000);
    });
    
    // Should have successful responses
    const successfulResponses = apiResponses.filter(r => r.status >= 200 && r.status < 300);
    expect(successfulResponses.length).toBeGreaterThan(0);
  });

  test('should implement efficient search functionality', async ({ page }) => {
    await page.goto('/search');
    
    // Test search performance
    const searchInput = page.locator('[data-testid="search-input"]');
    
    // Should implement debounced search
    await searchInput.fill('test query');
    
    // Wait for debounce
    await page.waitForTimeout(300);
    
    // Should show search results
    await expect(page.locator('[data-testid="search-results"]')).toBeVisible();
    
    // Test search result loading time
    const startTime = Date.now();
    
    await searchInput.fill('another query');
    await page.waitForSelector('[data-testid="search-results"] [data-testid="result-item"]');
    
    const searchTime = Date.now() - startTime;
    
    // Search should return results within 1 second
    expect(searchTime).toBeLessThan(1000);
  });

  test('should handle memory usage efficiently', async ({ page }) => {
    await page.goto('/');
    
    // Get initial memory usage
    const initialMemory = await page.evaluate(() => {
      return (performance as any).memory ? (performance as any).memory.usedJSHeapSize : 0;
    });
    
    // Navigate through multiple pages
    for (let i = 0; i < 5; i++) {
      await page.goto(`/page-${i}`);
      await page.waitForLoadState('networkidle');
    }
    
    // Force garbage collection if available
    await page.evaluate(() => {
      if ((window as any).gc) {
        (window as any).gc();
      }
    });
    
    // Get final memory usage
    const finalMemory = await page.evaluate(() => {
      return (performance as any).memory ? (performance as any).memory.usedJSHeapSize : 0;
    });
    
    if (initialMemory > 0 && finalMemory > 0) {
      // Memory usage should not increase dramatically
      const memoryIncrease = finalMemory - initialMemory;
      const memoryIncreasePercentage = (memoryIncrease / initialMemory) * 100;
      
      // Memory increase should be less than 50%
      expect(memoryIncreasePercentage).toBeLessThan(50);
    }
  });

  test('should optimize for different device capabilities', async ({ page, browserName }) => {
    // Test performance on different browsers
    await page.goto('/');
    
    // Measure performance metrics
    const performanceMetrics = await page.evaluate(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      
      return {
        domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
        loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
        firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime || 0,
        firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime || 0
      };
    });
    
    // Performance thresholds may vary by browser
    const thresholds = {
      chromium: { domContentLoaded: 1000, loadComplete: 2000 },
      firefox: { domContentLoaded: 1200, loadComplete: 2500 },
      webkit: { domContentLoaded: 1500, loadComplete: 3000 }
    };
    
    const threshold = thresholds[browserName] || thresholds.chromium;
    
    expect(performanceMetrics.domContentLoaded).toBeLessThan(threshold.domContentLoaded);
    expect(performanceMetrics.loadComplete).toBeLessThan(threshold.loadComplete);
  });
});