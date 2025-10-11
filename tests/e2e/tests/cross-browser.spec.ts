import { test, expect, devices } from '@playwright/test';
import { TestHelpers, AuthHelper } from '../utils/test-helpers';
import { testUsers, testBrowsers, testContent } from '../fixtures/test-data';

test.describe('Cross-Browser Compatibility Tests', () => {
  let helper: TestHelpers;
  let auth: AuthHelper;

  test.beforeEach(async ({ page }) => {
    helper = new TestHelpers(page);
    auth = new AuthHelper(page);
  });

  test.describe('Core Functionality Across Browsers', () => {
    testBrowsers.forEach(browserName => {
      test(`should work correctly in ${browserName}`, async ({ page, browserName: currentBrowser }) => {
        test.skip(currentBrowser !== browserName, `Skipping test for ${currentBrowser}, running only for ${browserName}`);
        
        await page.goto('/');

        // Test basic page load
        await expect(page).toHaveTitle(/Taishang Laojun/);
        
        // Test navigation
        const homeLink = await helper.waitForElement('[data-testid="nav-home"]');
        await expect(homeLink).toBeVisible();

        // Test JavaScript functionality
        await helper.clickElement('[data-testid="features-button"]');
        await helper.waitForNavigation('/features');
        await expect(page).toHaveURL(/\/features/);

        // Test form interactions
        await page.goto('/contact');
        await helper.fillField('[data-testid="name-input"]', 'Test User');
        await helper.fillField('[data-testid="email-input"]', 'test@example.com');
        await helper.fillField('[data-testid="message-input"]', 'Test message');
        
        await helper.clickElement('[data-testid="submit-button"]');
        const successMessage = await helper.waitForElement('[data-testid="success-message"]');
        await expect(successMessage).toBeVisible();

        console.log(`✓ Core functionality verified for ${browserName}`);
      });
    });
  });

  test.describe('CSS and Layout Consistency', () => {
    test('should have consistent layout across browsers', async ({ page }) => {
      await page.goto('/');

      // Check header layout
      const header = await helper.waitForElement('[data-testid="header"]');
      const headerBox = await header.boundingBox();
      expect(headerBox?.height).toBeGreaterThan(60);
      expect(headerBox?.width).toBeGreaterThan(0);

      // Check navigation layout
      const nav = page.locator('[data-testid="navigation"]');
      const navItems = page.locator('[data-testid="nav-item"]');
      const navCount = await navItems.count();
      expect(navCount).toBeGreaterThan(0);

      // Check responsive grid
      const grid = page.locator('[data-testid="feature-grid"]');
      const gridItems = page.locator('[data-testid="grid-item"]');
      const itemCount = await gridItems.count();
      
      for (let i = 0; i < itemCount; i++) {
        const item = gridItems.nth(i);
        const itemBox = await item.boundingBox();
        expect(itemBox?.width).toBeGreaterThan(200);
        expect(itemBox?.height).toBeGreaterThan(150);
      }

      // Check footer layout
      const footer = page.locator('[data-testid="footer"]');
      await helper.scrollToElement('[data-testid="footer"]');
      await expect(footer).toBeVisible();
    });

    test('should handle CSS Grid and Flexbox consistently', async ({ page }) => {
      await page.goto('/dashboard');

      // Test CSS Grid layout
      const dashboardGrid = page.locator('[data-testid="dashboard-grid"]');
      const gridStyle = await dashboardGrid.evaluate(el => 
        window.getComputedStyle(el).display
      );
      expect(gridStyle).toBe('grid');

      // Test Flexbox layout
      const toolbar = page.locator('[data-testid="toolbar"]');
      const flexStyle = await toolbar.evaluate(el => 
        window.getComputedStyle(el).display
      );
      expect(flexStyle).toBe('flex');

      // Check grid item positioning
      const gridItems = page.locator('[data-testid="dashboard-widget"]');
      const itemCount = await gridItems.count();
      
      for (let i = 0; i < itemCount; i++) {
        const item = gridItems.nth(i);
        const itemBox = await item.boundingBox();
        expect(itemBox?.x).toBeGreaterThanOrEqual(0);
        expect(itemBox?.y).toBeGreaterThanOrEqual(0);
      }
    });

    test('should render fonts consistently', async ({ page }) => {
      await page.goto('/');

      // Check primary font
      const heading = page.locator('h1').first();
      const headingFont = await heading.evaluate(el => 
        window.getComputedStyle(el).fontFamily
      );
      expect(headingFont).toContain('Inter'); // Assuming Inter is the primary font

      // Check body font
      const paragraph = page.locator('p').first();
      const bodyFont = await paragraph.evaluate(el => 
        window.getComputedStyle(el).fontFamily
      );
      expect(bodyFont).toContain('Inter');

      // Check font weights
      const boldText = page.locator('[data-testid="bold-text"]');
      if (await boldText.count() > 0) {
        const fontWeight = await boldText.evaluate(el => 
          window.getComputedStyle(el).fontWeight
        );
        expect(parseInt(fontWeight)).toBeGreaterThanOrEqual(600);
      }
    });
  });

  test.describe('JavaScript API Compatibility', () => {
    test('should handle modern JavaScript features', async ({ page }) => {
      await page.goto('/');

      // Test ES6+ features
      const modernJSSupport = await page.evaluate(() => {
        try {
          // Test arrow functions
          const arrow = () => 'arrow';
          
          // Test template literals
          const template = `template ${arrow()}`;
          
          // Test destructuring
          const { length } = [1, 2, 3];
          
          // Test async/await
          const asyncTest = async () => 'async';
          
          // Test Promises
          const promiseTest = Promise.resolve('promise');
          
          // Test Map/Set
          const map = new Map();
          const set = new Set();
          
          return {
            arrow: arrow() === 'arrow',
            template: template === 'template arrow',
            destructuring: length === 3,
            async: typeof asyncTest === 'function',
            promise: promiseTest instanceof Promise,
            map: map instanceof Map,
            set: set instanceof Set
          };
        } catch (error) {
          return { error: error.message };
        }
      });

      expect(modernJSSupport.arrow).toBe(true);
      expect(modernJSSupport.template).toBe(true);
      expect(modernJSSupport.destructuring).toBe(true);
      expect(modernJSSupport.async).toBe(true);
      expect(modernJSSupport.promise).toBe(true);
      expect(modernJSSupport.map).toBe(true);
      expect(modernJSSupport.set).toBe(true);
    });

    test('should handle Web APIs consistently', async ({ page }) => {
      await page.goto('/');

      // Test Fetch API
      const fetchSupport = await page.evaluate(() => {
        return typeof fetch === 'function';
      });
      expect(fetchSupport).toBe(true);

      // Test Local Storage
      const localStorageSupport = await page.evaluate(() => {
        try {
          localStorage.setItem('test', 'value');
          const value = localStorage.getItem('test');
          localStorage.removeItem('test');
          return value === 'value';
        } catch {
          return false;
        }
      });
      expect(localStorageSupport).toBe(true);

      // Test Session Storage
      const sessionStorageSupport = await page.evaluate(() => {
        try {
          sessionStorage.setItem('test', 'value');
          const value = sessionStorage.getItem('test');
          sessionStorage.removeItem('test');
          return value === 'value';
        } catch {
          return false;
        }
      });
      expect(sessionStorageSupport).toBe(true);

      // Test Geolocation API (if available)
      const geolocationSupport = await page.evaluate(() => {
        return 'geolocation' in navigator;
      });
      expect(geolocationSupport).toBe(true);

      // Test History API
      const historySupport = await page.evaluate(() => {
        return typeof history.pushState === 'function';
      });
      expect(historySupport).toBe(true);
    });

    test('should handle event listeners consistently', async ({ page }) => {
      await page.goto('/');

      // Test click events
      let clickHandled = false;
      await page.evaluate(() => {
        const button = document.querySelector('[data-testid="test-button"]');
        if (button) {
          button.addEventListener('click', () => {
            (window as any).clickHandled = true;
          });
        }
      });

      await helper.clickElement('[data-testid="test-button"]');
      clickHandled = await page.evaluate(() => (window as any).clickHandled);
      expect(clickHandled).toBe(true);

      // Test keyboard events
      await page.evaluate(() => {
        document.addEventListener('keydown', (e) => {
          if (e.key === 'Enter') {
            (window as any).enterPressed = true;
          }
        });
      });

      await helper.pressKeys('Enter');
      const enterPressed = await page.evaluate(() => (window as any).enterPressed);
      expect(enterPressed).toBe(true);

      // Test custom events
      await page.evaluate(() => {
        const customEvent = new CustomEvent('testEvent', { detail: 'test' });
        document.addEventListener('testEvent', (e: any) => {
          (window as any).customEventData = e.detail;
        });
        document.dispatchEvent(customEvent);
      });

      const customEventData = await page.evaluate(() => (window as any).customEventData);
      expect(customEventData).toBe('test');
    });
  });

  test.describe('Authentication Across Browsers', () => {
    test('should handle login/logout consistently', async ({ page }) => {
      await page.goto('/login');

      // Test login
      await auth.login(testUsers.user.email, testUsers.user.password);
      await expect(page).toHaveURL(/\/dashboard/);

      // Verify user session
      const userInfo = page.locator('[data-testid="user-info"]');
      await expect(userInfo).toBeVisible();

      // Test logout
      await auth.logout();
      await expect(page).toHaveURL('/');

      // Verify session cleared
      const loginButton = page.locator('[data-testid="login-button"]');
      await expect(loginButton).toBeVisible();
    });

    test('should handle session persistence', async ({ page }) => {
      // Login
      await auth.login(testUsers.user.email, testUsers.user.password);
      await expect(page).toHaveURL(/\/dashboard/);

      // Refresh page
      await page.reload();
      await expect(page).toHaveURL(/\/dashboard/);

      // Verify session persisted
      const userInfo = page.locator('[data-testid="user-info"]');
      await expect(userInfo).toBeVisible();

      // Test navigation
      await page.goto('/profile');
      await expect(page).toHaveURL(/\/profile/);

      // Go back to dashboard
      await page.goto('/dashboard');
      await expect(userInfo).toBeVisible();
    });
  });

  test.describe('Form Handling Across Browsers', () => {
    test('should handle form validation consistently', async ({ page }) => {
      await page.goto('/register');

      // Test HTML5 validation
      const emailInput = page.locator('[data-testid="email-input"]');
      await emailInput.fill('invalid-email');
      
      const validationMessage = await emailInput.evaluate((input: HTMLInputElement) => {
        return input.validationMessage;
      });
      expect(validationMessage).toBeTruthy();

      // Test custom validation
      await helper.fillField('[data-testid="password-input"]', '123');
      await helper.clickElement('[data-testid="register-button"]');

      const passwordError = await helper.waitForElement('[data-testid="password-error"]');
      await expect(passwordError).toBeVisible();
      await expect(passwordError).toContainText('at least 8 characters');
    });

    test('should handle file uploads consistently', async ({ page }) => {
      await page.goto('/upload');

      // Create test file
      const testFile = 'test-files/test-image.jpg';
      
      // Test file upload
      await helper.uploadFile('[data-testid="file-input"]', testFile);

      // Verify file selected
      const fileName = await helper.getElementText('[data-testid="file-name"]');
      expect(fileName).toContain('test-image.jpg');

      // Test upload progress
      await helper.clickElement('[data-testid="upload-button"]');
      
      const progressBar = page.locator('[data-testid="upload-progress"]');
      await expect(progressBar).toBeVisible();

      // Wait for upload completion
      const successMessage = await helper.waitForElement('[data-testid="upload-success"]');
      await expect(successMessage).toBeVisible();
    });
  });

  test.describe('Media and Graphics', () => {
    test('should handle images consistently', async ({ page }) => {
      await page.goto('/gallery');

      // Test image loading
      const images = page.locator('[data-testid="gallery-image"]');
      const imageCount = await images.count();
      expect(imageCount).toBeGreaterThan(0);

      // Check each image loads
      for (let i = 0; i < imageCount; i++) {
        const image = images.nth(i);
        await expect(image).toBeVisible();
        
        // Check image has loaded
        const naturalWidth = await image.evaluate((img: HTMLImageElement) => img.naturalWidth);
        expect(naturalWidth).toBeGreaterThan(0);
      }

      // Test image lazy loading
      const lazyImages = page.locator('[data-testid="lazy-image"]');
      if (await lazyImages.count() > 0) {
        const firstLazyImage = lazyImages.first();
        await helper.scrollToElement('[data-testid="lazy-image"]');
        await expect(firstLazyImage).toBeVisible();
      }
    });

    test('should handle CSS animations consistently', async ({ page }) => {
      await page.goto('/animations');

      // Test CSS transitions
      const animatedElement = page.locator('[data-testid="animated-element"]');
      await expect(animatedElement).toBeVisible();

      // Trigger animation
      await helper.hoverElement('[data-testid="animated-element"]');
      
      // Check animation properties
      const transform = await animatedElement.evaluate(el => 
        window.getComputedStyle(el).transform
      );
      expect(transform).not.toBe('none');

      // Test keyframe animations
      const keyframeElement = page.locator('[data-testid="keyframe-animation"]');
      if (await keyframeElement.count() > 0) {
        const animationName = await keyframeElement.evaluate(el => 
          window.getComputedStyle(el).animationName
        );
        expect(animationName).not.toBe('none');
      }
    });
  });

  test.describe('Performance Across Browsers', () => {
    test('should maintain performance standards', async ({ page }) => {
      const startTime = Date.now();
      await page.goto('/');
      const loadTime = Date.now() - startTime;

      // Check load time
      expect(loadTime).toBeLessThan(3000);

      // Check Core Web Vitals
      const vitals = await helper.checkCoreWebVitals();
      expect(vitals.fcp).toBeLessThan(1800);
      expect(vitals.lcp).toBeLessThan(2500);
      expect(vitals.cls).toBeLessThan(0.1);

      // Check JavaScript execution time
      const jsPerformance = await page.evaluate(() => {
        const start = performance.now();
        
        // Simulate some JavaScript work
        for (let i = 0; i < 10000; i++) {
          Math.random();
        }
        
        return performance.now() - start;
      });

      expect(jsPerformance).toBeLessThan(100); // Should complete in under 100ms
    });

    test('should handle memory usage efficiently', async ({ page }) => {
      await page.goto('/');

      // Get initial memory usage
      const initialMemory = await page.evaluate(() => {
        return (performance as any).memory?.usedJSHeapSize || 0;
      });

      // Navigate through several pages
      const pages = ['/features', '/pricing', '/about', '/contact', '/dashboard'];
      
      for (const pagePath of pages) {
        await page.goto(pagePath);
        await page.waitForLoadState('networkidle');
      }

      // Check final memory usage
      const finalMemory = await page.evaluate(() => {
        return (performance as any).memory?.usedJSHeapSize || 0;
      });

      // Memory increase should be reasonable
      if (initialMemory > 0 && finalMemory > 0) {
        const memoryIncrease = finalMemory - initialMemory;
        const memoryIncreasePercent = (memoryIncrease / initialMemory) * 100;
        expect(memoryIncreasePercent).toBeLessThan(200); // Less than 200% increase
      }
    });
  });

  test.describe('Error Handling Across Browsers', () => {
    test('should handle JavaScript errors gracefully', async ({ page }) => {
      const errors: string[] = [];
      
      page.on('pageerror', error => {
        errors.push(error.message);
      });

      page.on('console', msg => {
        if (msg.type() === 'error') {
          errors.push(msg.text());
        }
      });

      await page.goto('/');
      
      // Navigate through the app
      await page.goto('/features');
      await page.goto('/dashboard');
      
      // Should have no JavaScript errors
      expect(errors).toHaveLength(0);
    });

    test('should handle network errors gracefully', async ({ page }) => {
      await page.goto('/');

      // Mock network failure
      await page.route('**/api/**', route => {
        route.abort('failed');
      });

      // Try to perform an action that requires API
      await helper.clickElement('[data-testid="load-data-button"]');

      // Should show error message
      const errorMessage = await helper.waitForElement('[data-testid="error-message"]');
      await expect(errorMessage).toBeVisible();
      await expect(errorMessage).toContainText('network error');
    });
  });

  test.describe('Accessibility Across Browsers', () => {
    test('should maintain accessibility standards', async ({ page }) => {
      await page.goto('/');

      // Run accessibility audit
      await helper.checkAccessibility();

      // Check keyboard navigation
      await helper.pressKeys('Tab');
      const focusedElement = await page.evaluate(() => document.activeElement?.tagName);
      expect(focusedElement).toBeTruthy();

      // Check ARIA attributes
      const buttons = page.locator('button[aria-label], button[aria-labelledby]');
      const buttonCount = await buttons.count();
      expect(buttonCount).toBeGreaterThan(0);

      // Check heading structure
      const h1Count = await page.locator('h1').count();
      expect(h1Count).toBe(1);
    });
  });
});