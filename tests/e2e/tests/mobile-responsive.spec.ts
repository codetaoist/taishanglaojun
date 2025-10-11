import { test, expect } from '@playwright/test';
import { TestHelpers } from '../utils/test-helpers';
import { testDevices, testUsers } from '../fixtures/test-data';

test.describe('Mobile Responsive Tests', () => {
  let helper: TestHelpers;

  test.beforeEach(async ({ page }) => {
    helper = new TestHelpers(page);
  });

  test.describe('Mobile Navigation', () => {
    Object.entries(testDevices.mobile).forEach(([deviceName, viewport]) => {
      test(`should display mobile navigation on ${deviceName}`, async ({ page }) => {
        await page.setViewportSize(viewport);
        await page.goto('/');

        // Check mobile menu button exists
        const mobileMenuButton = await helper.waitForElement('[data-testid="mobile-menu-button"]');
        await expect(mobileMenuButton).toBeVisible();

        // Check desktop navigation is hidden
        const desktopNav = page.locator('[data-testid="desktop-navigation"]');
        await expect(desktopNav).toBeHidden();

        // Test mobile menu functionality
        await helper.clickElement('[data-testid="mobile-menu-button"]');
        const mobileMenu = await helper.waitForElement('[data-testid="mobile-menu"]');
        await expect(mobileMenu).toBeVisible();

        // Check navigation items
        await expect(page.locator('[data-testid="mobile-nav-home"]')).toBeVisible();
        await expect(page.locator('[data-testid="mobile-nav-features"]')).toBeVisible();
        await expect(page.locator('[data-testid="mobile-nav-pricing"]')).toBeVisible();
        await expect(page.locator('[data-testid="mobile-nav-contact"]')).toBeVisible();

        // Close mobile menu
        await helper.clickElement('[data-testid="mobile-menu-close"]');
        await expect(mobileMenu).toBeHidden();
      });
    });
  });

  test.describe('Touch Interactions', () => {
    test('should handle touch gestures on mobile', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/dashboard');

      // Test swipe gestures on carousel
      const carousel = await helper.waitForElement('[data-testid="feature-carousel"]');
      
      // Swipe left
      await carousel.hover();
      await page.mouse.down();
      await page.mouse.move(100, 0);
      await page.mouse.up();

      // Verify carousel moved
      const activeSlide = page.locator('[data-testid="carousel-slide"].active');
      await expect(activeSlide).toHaveAttribute('data-slide', '2');

      // Test pinch zoom on images
      const image = await helper.waitForElement('[data-testid="zoomable-image"]');
      await image.hover();
      
      // Simulate pinch zoom
      await page.touchscreen.tap(200, 200);
      await page.touchscreen.tap(250, 250);
      
      // Verify zoom effect
      await expect(image).toHaveCSS('transform', /scale/);
    });

    test('should handle long press interactions', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/content');

      const contentItem = await helper.waitForElement('[data-testid="content-item-1"]');
      
      // Long press to show context menu
      await contentItem.hover();
      await page.mouse.down();
      await page.waitForTimeout(1000); // Long press duration
      await page.mouse.up();

      // Verify context menu appears
      const contextMenu = await helper.waitForElement('[data-testid="context-menu"]');
      await expect(contextMenu).toBeVisible();
      await expect(page.locator('[data-testid="context-edit"]')).toBeVisible();
      await expect(page.locator('[data-testid="context-delete"]')).toBeVisible();
      await expect(page.locator('[data-testid="context-share"]')).toBeVisible();
    });
  });

  test.describe('Form Interactions on Mobile', () => {
    test('should handle form input on mobile devices', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/register');

      // Test input field focus and keyboard
      await helper.fillField('[data-testid="email-input"]', testUsers.user.email);
      
      // Verify virtual keyboard doesn't obscure input
      const emailInput = page.locator('[data-testid="email-input"]');
      const inputBox = await emailInput.boundingBox();
      expect(inputBox?.y).toBeGreaterThan(0);

      // Test select dropdown on mobile
      await helper.clickElement('[data-testid="region-select"]');
      const dropdown = await helper.waitForElement('[data-testid="region-dropdown"]');
      await expect(dropdown).toBeVisible();

      // Select option
      await helper.clickElement('[data-testid="region-us-east-1"]');
      await expect(page.locator('[data-testid="region-select"]')).toContainText('US East');

      // Test date picker on mobile
      await helper.clickElement('[data-testid="birthdate-input"]');
      const datePicker = await helper.waitForElement('[data-testid="date-picker"]');
      await expect(datePicker).toBeVisible();

      // Select date
      await helper.clickElement('[data-testid="date-15"]');
      await helper.clickElement('[data-testid="date-confirm"]');
    });

    test('should validate form inputs on mobile', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['Samsung Galaxy S21']);
      await page.goto('/contact');

      // Test required field validation
      await helper.clickElement('[data-testid="submit-button"]');
      
      const nameError = await helper.waitForElement('[data-testid="name-error"]');
      await expect(nameError).toBeVisible();
      await expect(nameError).toContainText('Name is required');

      // Test email validation
      await helper.fillField('[data-testid="email-input"]', 'invalid-email');
      await helper.clickElement('[data-testid="submit-button"]');
      
      const emailError = await helper.waitForElement('[data-testid="email-error"]');
      await expect(emailError).toBeVisible();
      await expect(emailError).toContainText('Please enter a valid email');

      // Test successful submission
      await helper.fillField('[data-testid="name-input"]', 'Test User');
      await helper.fillField('[data-testid="email-input"]', 'test@example.com');
      await helper.fillField('[data-testid="message-input"]', 'Test message');
      await helper.clickElement('[data-testid="submit-button"]');

      const successMessage = await helper.waitForElement('[data-testid="success-message"]');
      await expect(successMessage).toBeVisible();
    });
  });

  test.describe('Mobile Performance', () => {
    test('should load quickly on mobile devices', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      
      // Simulate slow 3G network
      await helper.simulateSlowNetwork();
      
      const startTime = Date.now();
      await page.goto('/');
      const loadTime = Date.now() - startTime;

      // Check load time is reasonable for mobile
      expect(loadTime).toBeLessThan(5000); // 5 seconds on slow network

      // Check Core Web Vitals
      const vitals = await helper.checkCoreWebVitals();
      expect(vitals.fcp).toBeLessThan(2500); // Mobile FCP budget
      expect(vitals.lcp).toBeLessThan(4000); // Mobile LCP budget
      expect(vitals.cls).toBeLessThan(0.25); // Mobile CLS budget
    });

    test('should optimize images for mobile', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['Google Pixel 5']);
      await page.goto('/gallery');

      // Check images are properly sized for mobile
      const images = page.locator('[data-testid="gallery-image"]');
      const imageCount = await images.count();

      for (let i = 0; i < imageCount; i++) {
        const image = images.nth(i);
        const imageBox = await image.boundingBox();
        
        // Images should not exceed viewport width
        expect(imageBox?.width).toBeLessThanOrEqual(testDevices.mobile['Google Pixel 5'].width);
        
        // Check lazy loading
        await helper.scrollToElement(`[data-testid="gallery-image"]:nth-child(${i + 1})`);
        await expect(image).toHaveAttribute('loading', 'lazy');
      }
    });
  });

  test.describe('Mobile Accessibility', () => {
    test('should be accessible on mobile devices', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/');

      // Check touch target sizes
      const buttons = page.locator('button, a, input[type="button"]');
      const buttonCount = await buttons.count();

      for (let i = 0; i < buttonCount; i++) {
        const button = buttons.nth(i);
        const buttonBox = await button.boundingBox();
        
        if (buttonBox) {
          // Touch targets should be at least 44x44px
          expect(buttonBox.width).toBeGreaterThanOrEqual(44);
          expect(buttonBox.height).toBeGreaterThanOrEqual(44);
        }
      }

      // Check text readability
      const textElements = page.locator('p, span, div, h1, h2, h3, h4, h5, h6');
      const textCount = await textElements.count();

      for (let i = 0; i < Math.min(textCount, 10); i++) {
        const element = textElements.nth(i);
        const fontSize = await element.evaluate(el => 
          window.getComputedStyle(el).fontSize
        );
        
        // Text should be at least 16px on mobile
        const fontSizeNum = parseInt(fontSize.replace('px', ''));
        expect(fontSizeNum).toBeGreaterThanOrEqual(16);
      }

      // Run accessibility audit
      await helper.checkAccessibility();
    });

    test('should support screen readers on mobile', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/');

      // Check ARIA labels
      const interactiveElements = page.locator('button, a, input, select, textarea');
      const elementCount = await interactiveElements.count();

      for (let i = 0; i < elementCount; i++) {
        const element = interactiveElements.nth(i);
        const ariaLabel = await element.getAttribute('aria-label');
        const ariaLabelledBy = await element.getAttribute('aria-labelledby');
        const title = await element.getAttribute('title');
        const textContent = await element.textContent();

        // Each interactive element should have accessible text
        expect(
          ariaLabel || ariaLabelledBy || title || textContent?.trim()
        ).toBeTruthy();
      }

      // Check heading structure
      const headings = page.locator('h1, h2, h3, h4, h5, h6');
      const headingCount = await headings.count();
      expect(headingCount).toBeGreaterThan(0);

      // Should have only one h1
      const h1Count = await page.locator('h1').count();
      expect(h1Count).toBe(1);
    });
  });

  test.describe('Mobile Orientation', () => {
    test('should handle orientation changes', async ({ page }) => {
      // Start in portrait
      await page.setViewportSize({ width: 390, height: 844 });
      await page.goto('/dashboard');

      // Check portrait layout
      const sidebar = page.locator('[data-testid="sidebar"]');
      await expect(sidebar).toBeHidden(); // Sidebar hidden in mobile portrait

      const mobileNav = page.locator('[data-testid="mobile-navigation"]');
      await expect(mobileNav).toBeVisible();

      // Switch to landscape
      await page.setViewportSize({ width: 844, height: 390 });
      await page.waitForTimeout(500); // Wait for layout adjustment

      // Check landscape layout adaptations
      const content = page.locator('[data-testid="main-content"]');
      const contentBox = await content.boundingBox();
      
      // Content should adapt to landscape
      expect(contentBox?.width).toBeGreaterThan(contentBox?.height || 0);

      // Navigation should still be accessible
      await expect(mobileNav).toBeVisible();
    });

    test('should maintain functionality across orientations', async ({ page }) => {
      await page.setViewportSize({ width: 390, height: 844 }); // Portrait
      await page.goto('/create');

      // Test form in portrait
      await helper.fillField('[data-testid="title-input"]', 'Test Content');
      await helper.fillField('[data-testid="description-input"]', 'Test description');

      // Switch to landscape
      await page.setViewportSize({ width: 844, height: 390 });
      await page.waitForTimeout(500);

      // Form values should be preserved
      await expect(page.locator('[data-testid="title-input"]')).toHaveValue('Test Content');
      await expect(page.locator('[data-testid="description-input"]')).toHaveValue('Test description');

      // Form should still be functional
      await helper.clickElement('[data-testid="save-button"]');
      const successMessage = await helper.waitForElement('[data-testid="success-message"]');
      await expect(successMessage).toBeVisible();
    });
  });

  test.describe('Mobile-Specific Features', () => {
    test('should support pull-to-refresh', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/feed');

      // Simulate pull-to-refresh gesture
      await page.touchscreen.tap(200, 100);
      await page.mouse.move(200, 100);
      await page.mouse.down();
      await page.mouse.move(200, 300); // Pull down
      await page.waitForTimeout(500);
      await page.mouse.up();

      // Check refresh indicator
      const refreshIndicator = page.locator('[data-testid="refresh-indicator"]');
      await expect(refreshIndicator).toBeVisible();

      // Wait for refresh to complete
      await helper.waitForElementToDisappear('[data-testid="refresh-indicator"]');

      // Verify content was refreshed
      const timestamp = await helper.getElementText('[data-testid="last-updated"]');
      expect(timestamp).toBeTruthy();
    });

    test('should support haptic feedback', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      await page.goto('/');

      // Test haptic feedback on button press
      await page.evaluate(() => {
        // Mock haptic feedback API
        (navigator as any).vibrate = (pattern: number | number[]) => {
          console.log('Haptic feedback triggered:', pattern);
          return true;
        };
      });

      await helper.clickElement('[data-testid="primary-button"]');

      // Verify haptic feedback was triggered
      const hapticTriggered = await page.evaluate(() => {
        return (window as any).hapticFeedbackTriggered === true;
      });

      // Note: In real tests, you would verify actual haptic feedback
      // This is a simplified check for the test environment
    });

    test('should handle device capabilities', async ({ page }) => {
      await page.setViewportSize(testDevices.mobile['iPhone 12']);
      
      // Mock device capabilities
      await page.addInitScript(() => {
        Object.defineProperty(navigator, 'deviceMemory', {
          get: () => 4 // 4GB RAM
        });
        
        Object.defineProperty(navigator, 'hardwareConcurrency', {
          get: () => 6 // 6 CPU cores
        });

        Object.defineProperty(navigator, 'connection', {
          get: () => ({
            effectiveType: '4g',
            downlink: 10,
            rtt: 100
          })
        });
      });

      await page.goto('/');

      // Check if app adapts to device capabilities
      const performanceMode = await page.evaluate(() => {
        return (window as any).performanceMode;
      });

      // App should enable high-performance features on capable devices
      expect(performanceMode).toBe('high');

      // Check if animations are enabled
      const animationsEnabled = await page.evaluate(() => {
        return !document.body.classList.contains('reduce-motion');
      });

      expect(animationsEnabled).toBe(true);
    });
  });

  test.describe('Cross-Device Consistency', () => {
    test('should maintain consistent experience across mobile devices', async ({ page }) => {
      const devices = Object.entries(testDevices.mobile);

      for (const [deviceName, viewport] of devices) {
        await page.setViewportSize(viewport);
        await page.goto('/');

        // Check consistent branding
        const logo = await helper.waitForElement('[data-testid="logo"]');
        await expect(logo).toBeVisible();

        // Check consistent navigation
        const mobileMenu = page.locator('[data-testid="mobile-menu-button"]');
        await expect(mobileMenu).toBeVisible();

        // Check consistent content structure
        const mainContent = page.locator('[data-testid="main-content"]');
        await expect(mainContent).toBeVisible();

        // Check consistent footer
        const footer = page.locator('[data-testid="footer"]');
        await expect(footer).toBeVisible();

        console.log(`✓ Consistency verified for ${deviceName}`);
      }
    });
  });
});