import { test, expect } from '@playwright/test';

/**
 * Localization E2E Tests
 * Tests multi-language support, timezone handling, and cultural adaptations
 */
test.describe('Localization Features @localization @global', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
  });

  test('should display content in Chinese by default', async ({ page }) => {
    // Check if the page loads with Chinese content
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText('欢迎使用太上老君AI平台');
    await expect(page.locator('[data-testid="navigation-dashboard"]')).toContainText('仪表板');
  });

  test('should switch to English when language is changed', async ({ page }) => {
    // Open language selector
    await page.click('[data-testid="language-selector"]');
    
    // Select English
    await page.click('[data-testid="language-option-en"]');
    
    // Verify content is now in English
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText('Welcome to Taishang Laojun AI Platform');
    await expect(page.locator('[data-testid="navigation-dashboard"]')).toContainText('Dashboard');
  });

  test('should switch to German when language is changed', async ({ page }) => {
    // Open language selector
    await page.click('[data-testid="language-selector"]');
    
    // Select German
    await page.click('[data-testid="language-option-de"]');
    
    // Verify content is now in German
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText('Willkommen bei der Taishang Laojun AI-Plattform');
    await expect(page.locator('[data-testid="navigation-dashboard"]')).toContainText('Dashboard');
  });

  test('should persist language preference across sessions', async ({ page, context }) => {
    // Change language to English
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-en"]');
    
    // Verify language changed
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText('Welcome');
    
    // Create new page (simulating new session)
    const newPage = await context.newPage();
    await newPage.goto('/');
    
    // Verify language preference is persisted
    await expect(newPage.locator('[data-testid="welcome-message"]')).toContainText('Welcome');
    
    await newPage.close();
  });

  test('should display correct date and time format for different locales', async ({ page }) => {
    // Test Chinese locale (default)
    await expect(page.locator('[data-testid="current-date"]')).toMatch(/\d{4}年\d{1,2}月\d{1,2}日/);
    
    // Switch to US locale
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-en"]');
    
    // Verify US date format
    await expect(page.locator('[data-testid="current-date"]')).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/);
    
    // Switch to German locale
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-de"]');
    
    // Verify German date format
    await expect(page.locator('[data-testid="current-date"]')).toMatch(/\d{1,2}\.\d{1,2}\.\d{4}/);
  });

  test('should handle timezone conversion correctly', async ({ page }) => {
    // Navigate to timezone settings
    await page.goto('/settings/timezone');
    
    // Select New York timezone
    await page.selectOption('[data-testid="timezone-selector"]', 'America/New_York');
    
    // Verify timezone display
    await expect(page.locator('[data-testid="selected-timezone"]')).toContainText('America/New_York');
    
    // Check if time is displayed correctly for the selected timezone
    const timeElement = page.locator('[data-testid="current-time"]');
    await expect(timeElement).toBeVisible();
    
    // Switch to Berlin timezone
    await page.selectOption('[data-testid="timezone-selector"]', 'Europe/Berlin');
    await expect(page.locator('[data-testid="selected-timezone"]')).toContainText('Europe/Berlin');
  });

  test('should display currency format based on locale', async ({ page }) => {
    // Navigate to pricing page
    await page.goto('/pricing');
    
    // Default should show Chinese currency
    await expect(page.locator('[data-testid="price-display"]')).toContainText('¥');
    
    // Switch to US locale
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-en"]');
    
    // Should show USD
    await expect(page.locator('[data-testid="price-display"]')).toContainText('$');
    
    // Switch to German locale
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-de"]');
    
    // Should show EUR
    await expect(page.locator('[data-testid="price-display"]')).toContainText('€');
  });

  test('should handle RTL languages correctly', async ({ page }) => {
    // This test would be relevant if Arabic or Hebrew support is added
    // For now, we'll test the framework's capability
    
    // Check if RTL detection works
    const htmlElement = page.locator('html');
    await expect(htmlElement).toHaveAttribute('dir', 'ltr'); // Default LTR
    
    // If RTL language is selected, dir should change to 'rtl'
    // This is a placeholder for future RTL language support
  });

  test('should load appropriate fonts for different languages', async ({ page }) => {
    // Check if Chinese fonts are loaded
    const bodyElement = page.locator('body');
    const fontFamily = await bodyElement.evaluate(el => getComputedStyle(el).fontFamily);
    
    // Should include Chinese-friendly fonts
    expect(fontFamily).toMatch(/(PingFang|Hiragino|Microsoft YaHei|SimSun)/i);
    
    // Switch to English and verify font changes if needed
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-en"]');
    
    // Font family might change for better English readability
    const newFontFamily = await bodyElement.evaluate(el => getComputedStyle(el).fontFamily);
    expect(newFontFamily).toBeDefined();
  });

  test('should handle pluralization correctly in different languages', async ({ page }) => {
    // Navigate to a page with countable items
    await page.goto('/dashboard');
    
    // Test Chinese pluralization (no plural forms)
    await expect(page.locator('[data-testid="item-count"]')).toContainText('个项目');
    
    // Switch to English
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-en"]');
    
    // Test English pluralization
    const itemCount = await page.locator('[data-testid="item-count"]').textContent();
    expect(itemCount).toMatch(/(item|items)/);
    
    // Switch to German
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-de"]');
    
    // Test German pluralization
    const germanItemCount = await page.locator('[data-testid="item-count"]').textContent();
    expect(germanItemCount).toMatch(/(Element|Elemente)/);
  });

  test('should handle number formatting for different locales', async ({ page }) => {
    // Navigate to analytics page with numbers
    await page.goto('/analytics');
    
    // Chinese number formatting (default)
    await expect(page.locator('[data-testid="large-number"]')).toContainText('1,234,567');
    
    // Switch to German locale
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-de"]');
    
    // German number formatting uses dots for thousands
    await expect(page.locator('[data-testid="large-number"]')).toContainText('1.234.567');
  });
});