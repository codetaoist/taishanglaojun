import { Page, expect, Locator } from '@playwright/test';

/**
 * Test helper utilities for E2E tests
 */

export class TestHelpers {
  constructor(private page: Page) {}

  /**
   * Wait for element to be visible and stable
   */
  async waitForElement(selector: string, timeout = 10000): Promise<Locator> {
    const element = this.page.locator(selector);
    await element.waitFor({ state: 'visible', timeout });
    await element.waitFor({ state: 'attached', timeout });
    return element;
  }

  /**
   * Fill form field with validation
   */
  async fillField(selector: string, value: string, validate = true): Promise<void> {
    const field = await this.waitForElement(selector);
    await field.clear();
    await field.fill(value);
    
    if (validate) {
      await expect(field).toHaveValue(value);
    }
  }

  /**
   * Click element with retry mechanism
   */
  async clickElement(selector: string, retries = 3): Promise<void> {
    for (let i = 0; i < retries; i++) {
      try {
        const element = await this.waitForElement(selector);
        await element.click();
        return;
      } catch (error) {
        if (i === retries - 1) throw error;
        await this.page.waitForTimeout(1000);
      }
    }
  }

  /**
   * Wait for navigation to complete
   */
  async waitForNavigation(url?: string, timeout = 10000): Promise<void> {
    if (url) {
      await this.page.waitForURL(url, { timeout });
    } else {
      await this.page.waitForLoadState('networkidle', { timeout });
    }
  }

  /**
   * Take screenshot with timestamp
   */
  async takeScreenshot(name: string): Promise<void> {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    await this.page.screenshot({
      path: `test-results/screenshots/${name}-${timestamp}.png`,
      fullPage: true
    });
  }

  /**
   * Check if element exists without throwing
   */
  async elementExists(selector: string): Promise<boolean> {
    try {
      await this.page.locator(selector).waitFor({ state: 'attached', timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get element text content
   */
  async getElementText(selector: string): Promise<string> {
    const element = await this.waitForElement(selector);
    return await element.textContent() || '';
  }

  /**
   * Scroll element into view
   */
  async scrollToElement(selector: string): Promise<void> {
    const element = await this.waitForElement(selector);
    await element.scrollIntoViewIfNeeded();
  }

  /**
   * Wait for API response
   */
  async waitForApiResponse(urlPattern: string | RegExp, timeout = 10000): Promise<any> {
    const response = await this.page.waitForResponse(urlPattern, { timeout });
    return await response.json();
  }

  /**
   * Mock API response
   */
  async mockApiResponse(urlPattern: string | RegExp, responseData: any): Promise<void> {
    await this.page.route(urlPattern, route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(responseData)
      });
    });
  }

  /**
   * Set geolocation for testing
   */
  async setGeolocation(latitude: number, longitude: number): Promise<void> {
    await this.page.context().setGeolocation({ latitude, longitude });
  }

  /**
   * Set timezone for testing
   */
  async setTimezone(timezone: string): Promise<void> {
    await this.page.context().addInitScript(`
      Object.defineProperty(Intl, 'DateTimeFormat', {
        value: class extends Intl.DateTimeFormat {
          constructor(...args) {
            super(...args);
            this.resolvedOptions = () => ({ ...super.resolvedOptions(), timeZone: '${timezone}' });
          }
        }
      });
    `);
  }

  /**
   * Set locale for testing
   */
  async setLocale(locale: string): Promise<void> {
    await this.page.context().addInitScript(`
      Object.defineProperty(navigator, 'language', {
        get: function() { return '${locale}'; }
      });
      Object.defineProperty(navigator, 'languages', {
        get: function() { return ['${locale}']; }
      });
    `);
  }

  /**
   * Simulate slow network
   */
  async simulateSlowNetwork(): Promise<void> {
    await this.page.context().route('**/*', route => {
      setTimeout(() => route.continue(), 1000);
    });
  }

  /**
   * Check accessibility violations
   */
  async checkAccessibility(): Promise<void> {
    await this.page.addScriptTag({
      url: 'https://unpkg.com/axe-core@4.7.0/axe.min.js'
    });

    const violations = await this.page.evaluate(() => {
      return new Promise((resolve) => {
        // @ts-ignore
        axe.run(document, (err, results) => {
          resolve(results.violations);
        });
      });
    });

    expect(violations).toHaveLength(0);
  }

  /**
   * Measure page performance
   */
  async measurePerformance(): Promise<any> {
    return await this.page.evaluate(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      const paint = performance.getEntriesByType('paint');
      
      return {
        loadTime: navigation.loadEventEnd - navigation.loadEventStart,
        domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
        firstPaint: paint.find(p => p.name === 'first-paint')?.startTime || 0,
        firstContentfulPaint: paint.find(p => p.name === 'first-contentful-paint')?.startTime || 0,
        ttfb: navigation.responseStart - navigation.requestStart
      };
    });
  }

  /**
   * Check Core Web Vitals
   */
  async checkCoreWebVitals(): Promise<any> {
    return await this.page.evaluate(() => {
      return new Promise((resolve) => {
        const vitals = {
          fcp: 0,
          lcp: 0,
          fid: 0,
          cls: 0
        };

        // First Contentful Paint
        new PerformanceObserver((list) => {
          const entries = list.getEntries();
          vitals.fcp = entries[0]?.startTime || 0;
        }).observe({ entryTypes: ['paint'] });

        // Largest Contentful Paint
        new PerformanceObserver((list) => {
          const entries = list.getEntries();
          vitals.lcp = entries[entries.length - 1]?.startTime || 0;
        }).observe({ entryTypes: ['largest-contentful-paint'] });

        // First Input Delay
        new PerformanceObserver((list) => {
          const entries = list.getEntries();
          vitals.fid = entries[0]?.processingStart - entries[0]?.startTime || 0;
        }).observe({ entryTypes: ['first-input'] });

        // Cumulative Layout Shift
        new PerformanceObserver((list) => {
          const entries = list.getEntries();
          vitals.cls = entries.reduce((sum, entry) => sum + entry.value, 0);
        }).observe({ entryTypes: ['layout-shift'] });

        setTimeout(() => resolve(vitals), 3000);
      });
    });
  }

  /**
   * Upload file for testing
   */
  async uploadFile(selector: string, filePath: string): Promise<void> {
    const fileInput = await this.waitForElement(selector);
    await fileInput.setInputFiles(filePath);
  }

  /**
   * Download file and verify
   */
  async downloadFile(triggerSelector: string): Promise<string> {
    const downloadPromise = this.page.waitForEvent('download');
    await this.clickElement(triggerSelector);
    const download = await downloadPromise;
    const path = await download.path();
    return path || '';
  }

  /**
   * Clear browser storage
   */
  async clearStorage(): Promise<void> {
    await this.page.evaluate(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
    await this.page.context().clearCookies();
  }

  /**
   * Set cookie for testing
   */
  async setCookie(name: string, value: string, domain?: string): Promise<void> {
    await this.page.context().addCookies([{
      name,
      value,
      domain: domain || new URL(this.page.url()).hostname,
      path: '/'
    }]);
  }

  /**
   * Get cookie value
   */
  async getCookie(name: string): Promise<string | undefined> {
    const cookies = await this.page.context().cookies();
    return cookies.find(cookie => cookie.name === name)?.value;
  }

  /**
   * Wait for element to disappear
   */
  async waitForElementToDisappear(selector: string, timeout = 10000): Promise<void> {
    await this.page.locator(selector).waitFor({ state: 'detached', timeout });
  }

  /**
   * Hover over element
   */
  async hoverElement(selector: string): Promise<void> {
    const element = await this.waitForElement(selector);
    await element.hover();
  }

  /**
   * Double click element
   */
  async doubleClickElement(selector: string): Promise<void> {
    const element = await this.waitForElement(selector);
    await element.dblclick();
  }

  /**
   * Right click element
   */
  async rightClickElement(selector: string): Promise<void> {
    const element = await this.waitForElement(selector);
    await element.click({ button: 'right' });
  }

  /**
   * Press key combination
   */
  async pressKeys(keys: string): Promise<void> {
    await this.page.keyboard.press(keys);
  }

  /**
   * Type text with delay
   */
  async typeText(text: string, delay = 100): Promise<void> {
    await this.page.keyboard.type(text, { delay });
  }

  /**
   * Drag and drop
   */
  async dragAndDrop(sourceSelector: string, targetSelector: string): Promise<void> {
    const source = await this.waitForElement(sourceSelector);
    const target = await this.waitForElement(targetSelector);
    await source.dragTo(target);
  }

  /**
   * Check if page has error
   */
  async checkForErrors(): Promise<string[]> {
    const errors: string[] = [];
    
    this.page.on('pageerror', error => {
      errors.push(error.message);
    });

    this.page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    return errors;
  }

  /**
   * Wait for specific text to appear
   */
  async waitForText(text: string, timeout = 10000): Promise<void> {
    await this.page.waitForFunction(
      text => document.body.textContent?.includes(text),
      text,
      { timeout }
    );
  }

  /**
   * Get page title
   */
  async getPageTitle(): Promise<string> {
    return await this.page.title();
  }

  /**
   * Get current URL
   */
  async getCurrentUrl(): Promise<string> {
    return this.page.url();
  }

  /**
   * Reload page
   */
  async reloadPage(): Promise<void> {
    await this.page.reload({ waitUntil: 'networkidle' });
  }

  /**
   * Go back in browser history
   */
  async goBack(): Promise<void> {
    await this.page.goBack({ waitUntil: 'networkidle' });
  }

  /**
   * Go forward in browser history
   */
  async goForward(): Promise<void> {
    await this.page.goForward({ waitUntil: 'networkidle' });
  }
}

/**
 * Authentication helper
 */
export class AuthHelper {
  constructor(private page: Page) {}

  async login(email: string, password: string): Promise<void> {
    const helper = new TestHelpers(this.page);
    
    await this.page.goto('/login');
    await helper.fillField('[data-testid="email-input"]', email);
    await helper.fillField('[data-testid="password-input"]', password);
    await helper.clickElement('[data-testid="login-button"]');
    await helper.waitForNavigation('/dashboard');
  }

  async logout(): Promise<void> {
    const helper = new TestHelpers(this.page);
    
    await helper.clickElement('[data-testid="user-menu"]');
    await helper.clickElement('[data-testid="logout-button"]');
    await helper.waitForNavigation('/');
  }

  async register(userData: {
    email: string;
    password: string;
    firstName: string;
    lastName: string;
    region: string;
  }): Promise<void> {
    const helper = new TestHelpers(this.page);
    
    await this.page.goto('/register');
    await helper.fillField('[data-testid="email-input"]', userData.email);
    await helper.fillField('[data-testid="password-input"]', userData.password);
    await helper.fillField('[data-testid="first-name-input"]', userData.firstName);
    await helper.fillField('[data-testid="last-name-input"]', userData.lastName);
    await helper.clickElement(`[data-testid="region-${userData.region}"]`);
    await helper.clickElement('[data-testid="register-button"]');
  }
}

/**
 * API helper for test data management
 */
export class ApiHelper {
  constructor(private baseUrl: string) {}

  async createTestUser(userData: any): Promise<any> {
    const response = await fetch(`${this.baseUrl}/api/test/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });
    return await response.json();
  }

  async deleteTestUser(userId: string): Promise<void> {
    await fetch(`${this.baseUrl}/api/test/users/${userId}`, {
      method: 'DELETE'
    });
  }

  async createTestData(type: string, data: any): Promise<any> {
    const response = await fetch(`${this.baseUrl}/api/test/${type}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    return await response.json();
  }

  async cleanupTestData(): Promise<void> {
    await fetch(`${this.baseUrl}/api/test/cleanup`, {
      method: 'POST'
    });
  }
}

/**
 * Database helper for test data
 */
export class DatabaseHelper {
  constructor(private connectionString: string) {}

  async executeQuery(query: string, params: any[] = []): Promise<any> {
    // Implementation would depend on your database client
    // This is a placeholder for the actual implementation
    console.log('Executing query:', query, 'with params:', params);
  }

  async seedTestData(): Promise<void> {
    // Seed test data for E2E tests
    console.log('Seeding test data...');
  }

  async cleanupTestData(): Promise<void> {
    // Clean up test data after tests
    console.log('Cleaning up test data...');
  }
}

export { TestHelpers as default };