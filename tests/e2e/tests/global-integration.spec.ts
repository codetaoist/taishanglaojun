import { test, expect } from '@playwright/test';

/**
 * Global Integration E2E Tests
 * Tests end-to-end workflows across different regions and compliance requirements
 */
test.describe('Global Integration Tests @global @integration', () => {
  
  test('should handle complete user journey across regions', async ({ page, context }) => {
    // Simulate US user registration and usage
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'us-east-1',
      'Accept-Language': 'en-US,en;q=0.9'
    });
    
    // 1. Registration with CCPA compliance
    await page.goto('/register');
    
    // Should show CCPA notice
    await expect(page.locator('[data-testid="ccpa-notice"]')).toBeVisible();
    
    // Register new user
    await page.fill('[data-testid="email"]', 'global.user@test.com');
    await page.fill('[data-testid="password"]', 'GlobalTest123!');
    await page.fill('[data-testid="confirm-password"]', 'GlobalTest123!');
    
    // Accept terms and privacy policy
    await page.check('[data-testid="accept-terms"]');
    await page.check('[data-testid="accept-privacy"]');
    
    await page.click('[data-testid="register-button"]');
    
    // 2. Email verification
    await expect(page.locator('[data-testid="verification-sent"]')).toBeVisible();
    
    // Simulate email verification (in real test, would check email)
    await page.goto('/verify-email?token=test-verification-token');
    await expect(page.locator('[data-testid="email-verified"]')).toBeVisible();
    
    // 3. Initial setup with localization
    await page.goto('/onboarding');
    
    // Set preferences
    await page.selectOption('[data-testid="language-select"]', 'en-US');
    await page.selectOption('[data-testid="timezone-select"]', 'America/New_York');
    await page.selectOption('[data-testid="currency-select"]', 'USD');
    
    await page.click('[data-testid="save-preferences"]');
    
    // 4. Dashboard usage
    await expect(page).toHaveURL(/.*\/dashboard/);
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText('Welcome');
    
    // 5. Create content with AI assistance
    await page.click('[data-testid="create-content"]');
    await page.fill('[data-testid="content-prompt"]', 'Create a business plan for a tech startup');
    await page.click('[data-testid="generate-content"]');
    
    // Should show loading state
    await expect(page.locator('[data-testid="ai-generating"]')).toBeVisible();
    
    // Wait for content generation
    await expect(page.locator('[data-testid="generated-content"]')).toBeVisible({ timeout: 30000 });
    
    // 6. Save and manage content
    await page.click('[data-testid="save-content"]');
    await expect(page.locator('[data-testid="content-saved"]')).toBeVisible();
    
    // 7. Test data export (CCPA compliance)
    await page.goto('/settings/privacy');
    await page.click('[data-testid="export-data-button"]');
    await page.click('[data-testid="confirm-export"]');
    
    await expect(page.locator('[data-testid="export-initiated"]')).toBeVisible();
  });

  test('should handle region migration scenario', async ({ page, context }) => {
    // Start as EU user
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'eu-central-1',
      'Accept-Language': 'de-DE,de;q=0.9'
    });
    
    // Login as existing user
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'migration.user@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Verify EU compliance features
    await expect(page.locator('[data-testid="gdpr-compliant"]')).toBeVisible();
    
    // User moves to US - simulate region change
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'us-west-1',
      'Accept-Language': 'en-US,en;q=0.9'
    });
    
    await page.reload();
    
    // Should detect region change
    await expect(page.locator('[data-testid="region-change-notice"]')).toBeVisible();
    
    // Update compliance preferences
    await page.click('[data-testid="update-compliance"]');
    
    // Should now show CCPA compliance
    await expect(page.locator('[data-testid="ccpa-compliant"]')).toBeVisible();
    
    // Data should be migrated appropriately
    await page.goto('/settings/data-location');
    await expect(page.locator('[data-testid="data-location"]')).toContainText('United States');
  });

  test('should handle multi-language content creation and management', async ({ page }) => {
    // Login as admin
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to content management
    await page.goto('/admin/content');
    
    // Create content in multiple languages
    await page.click('[data-testid="create-multilingual-content"]');
    
    // English version
    await page.selectOption('[data-testid="content-language"]', 'en-US');
    await page.fill('[data-testid="content-title"]', 'Global AI Platform Features');
    await page.fill('[data-testid="content-body"]', 'Our platform offers advanced AI capabilities...');
    
    // Add Chinese version
    await page.click('[data-testid="add-language-version"]');
    await page.selectOption('[data-testid="content-language"]', 'zh-CN');
    await page.fill('[data-testid="content-title"]', '全球AI平台功能');
    await page.fill('[data-testid="content-body"]', '我们的平台提供先进的AI功能...');
    
    // Add German version
    await page.click('[data-testid="add-language-version"]');
    await page.selectOption('[data-testid="content-language"]', 'de-DE');
    await page.fill('[data-testid="content-title"]', 'Globale KI-Plattform-Funktionen');
    await page.fill('[data-testid="content-body"]', 'Unsere Plattform bietet fortschrittliche KI-Funktionen...');
    
    // Save multilingual content
    await page.click('[data-testid="save-multilingual-content"]');
    await expect(page.locator('[data-testid="content-saved-success"]')).toBeVisible();
    
    // Verify content appears correctly for different languages
    await page.goto('/');
    
    // Switch to Chinese
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-zh"]');
    await expect(page.locator('[data-testid="content-title"]')).toContainText('全球AI平台功能');
    
    // Switch to German
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-de"]');
    await expect(page.locator('[data-testid="content-title"]')).toContainText('Globale KI-Plattform-Funktionen');
  });

  test('should handle cross-border data compliance workflow', async ({ page, context }) => {
    // EU user accessing US-hosted service
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'eu-central-1',
      'X-Data-Location': 'us-east-1'
    });
    
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'eu.user@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Should show cross-border data transfer notice
    await expect(page.locator('[data-testid="cross-border-notice"]')).toBeVisible();
    
    // User can consent to cross-border transfer
    await page.click('[data-testid="consent-cross-border"]');
    
    // Should log consent in audit trail
    await page.goto('/settings/privacy/audit');
    await expect(page.locator('[data-testid="audit-entry"]')).toContainText('cross-border consent');
    
    // User can request data localization
    await page.goto('/settings/data-location');
    await page.click('[data-testid="request-data-localization"]');
    
    // Should show localization request confirmation
    await expect(page.locator('[data-testid="localization-request-sent"]')).toBeVisible();
  });

  test('should handle global search with localization', async ({ page }) => {
    await page.goto('/');
    
    // Test search in Chinese
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-zh"]');
    
    await page.fill('[data-testid="global-search"]', '人工智能');
    await page.click('[data-testid="search-button"]');
    
    // Should return Chinese results
    await expect(page.locator('[data-testid="search-results"]')).toBeVisible();
    const chineseResults = await page.locator('[data-testid="search-result-title"]').first().textContent();
    expect(chineseResults).toMatch(/[\u4e00-\u9fff]/); // Contains Chinese characters
    
    // Test search in English
    await page.click('[data-testid="language-selector"]');
    await page.click('[data-testid="language-option-en"]');
    
    await page.fill('[data-testid="global-search"]', 'artificial intelligence');
    await page.click('[data-testid="search-button"]');
    
    // Should return English results
    const englishResults = await page.locator('[data-testid="search-result-title"]').first().textContent();
    expect(englishResults).toMatch(/artificial intelligence/i);
    
    // Test search with translation
    await page.fill('[data-testid="global-search"]', 'künstliche intelligenz');
    await page.click('[data-testid="search-button"]');
    
    // Should handle German search and potentially translate
    await expect(page.locator('[data-testid="search-results"]')).toBeVisible();
  });

  test('should handle global payment processing', async ({ page, context }) => {
    // US user
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'us-east-1'
    });
    
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'payment.user@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to pricing
    await page.goto('/pricing');
    
    // Should show USD pricing
    await expect(page.locator('[data-testid="price-currency"]')).toContainText('$');
    
    // Select a plan
    await page.click('[data-testid="select-pro-plan"]');
    
    // Should show US payment methods
    await expect(page.locator('[data-testid="payment-method-card"]')).toBeVisible();
    await expect(page.locator('[data-testid="payment-method-paypal"]')).toBeVisible();
    
    // Fill payment information
    await page.fill('[data-testid="card-number"]', '4242424242424242');
    await page.fill('[data-testid="card-expiry"]', '12/25');
    await page.fill('[data-testid="card-cvc"]', '123');
    await page.fill('[data-testid="billing-zip"]', '10001');
    
    // Process payment
    await page.click('[data-testid="process-payment"]');
    
    // Should handle payment processing
    await expect(page.locator('[data-testid="payment-processing"]')).toBeVisible();
    
    // Should show success (in test environment)
    await expect(page.locator('[data-testid="payment-success"]')).toBeVisible({ timeout: 30000 });
    
    // Should update subscription status
    await page.goto('/dashboard');
    await expect(page.locator('[data-testid="subscription-status"]')).toContainText('Pro');
  });

  test('should handle global customer support workflow', async ({ page }) => {
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'support.user@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Access support
    await page.click('[data-testid="support-button"]');
    
    // Should show support options based on region/language
    await expect(page.locator('[data-testid="support-chat"]')).toBeVisible();
    await expect(page.locator('[data-testid="support-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="support-knowledge-base"]')).toBeVisible();
    
    // Create support ticket
    await page.click('[data-testid="create-ticket"]');
    
    await page.selectOption('[data-testid="ticket-category"]', 'technical');
    await page.selectOption('[data-testid="ticket-priority"]', 'medium');
    await page.fill('[data-testid="ticket-subject"]', 'Global deployment issue');
    await page.fill('[data-testid="ticket-description"]', 'Having issues with multi-region deployment...');
    
    // Should auto-detect user's region and language
    await expect(page.locator('[data-testid="detected-region"]')).toBeVisible();
    await expect(page.locator('[data-testid="detected-language"]')).toBeVisible();
    
    await page.click('[data-testid="submit-ticket"]');
    
    // Should show ticket created confirmation
    await expect(page.locator('[data-testid="ticket-created"]')).toBeVisible();
    
    // Should provide estimated response time based on region
    await expect(page.locator('[data-testid="response-time-estimate"]')).toBeVisible();
  });

  test('should handle global analytics and reporting', async ({ page }) => {
    // Login as admin
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to global analytics
    await page.goto('/admin/analytics/global');
    
    // Should show global metrics
    await expect(page.locator('[data-testid="global-users-metric"]')).toBeVisible();
    await expect(page.locator('[data-testid="regional-distribution"]')).toBeVisible();
    await expect(page.locator('[data-testid="language-usage"]')).toBeVisible();
    
    // Filter by region
    await page.selectOption('[data-testid="region-filter"]', 'us-east-1');
    
    // Should update metrics for selected region
    await expect(page.locator('[data-testid="region-specific-metrics"]')).toBeVisible();
    
    // Test compliance reporting
    await page.goto('/admin/compliance/reports');
    
    // Should show compliance metrics by region
    await expect(page.locator('[data-testid="gdpr-compliance-rate"]')).toBeVisible();
    await expect(page.locator('[data-testid="ccpa-compliance-rate"]')).toBeVisible();
    await expect(page.locator('[data-testid="pipl-compliance-rate"]')).toBeVisible();
    
    // Generate compliance report
    await page.click('[data-testid="generate-compliance-report"]');
    await page.selectOption('[data-testid="report-period"]', 'last-month');
    await page.click('[data-testid="download-report"]');
    
    // Should initiate report download
    await expect(page.locator('[data-testid="report-generating"]')).toBeVisible();
  });

  test('should handle disaster recovery scenario', async ({ page, context }) => {
    // Simulate primary region failure
    await context.route('**/api/**', route => {
      if (route.request().url().includes('us-east-1')) {
        route.fulfill({
          status: 503,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Service unavailable' })
        });
      } else {
        route.continue();
      }
    });
    
    await page.goto('/');
    
    // Should detect primary region failure and failover
    await expect(page.locator('[data-testid="failover-notice"]')).toBeVisible();
    
    // Should continue to function with backup region
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'disaster.user@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Should successfully login using backup infrastructure
    await expect(page).toHaveURL(/.*\/dashboard/);
    
    // Should show degraded service notice
    await expect(page.locator('[data-testid="degraded-service-notice"]')).toBeVisible();
    
    // Core functionality should still work
    await expect(page.locator('[data-testid="dashboard-content"]')).toBeVisible();
  });
});