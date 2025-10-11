import { test, expect } from '@playwright/test';

/**
 * Compliance E2E Tests
 * Tests GDPR, CCPA, PIPL compliance features and data protection
 */
test.describe('Compliance Features @compliance @global', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display GDPR compliance banner for EU users', async ({ page, context }) => {
    // Simulate EU user by setting geolocation and headers
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'eu-central-1',
      'Accept-Language': 'de-DE,de;q=0.9'
    });
    
    await page.goto('/');
    
    // Should display GDPR compliance banner
    await expect(page.locator('[data-testid="gdpr-banner"]')).toBeVisible();
    await expect(page.locator('[data-testid="gdpr-banner"]')).toContainText('GDPR');
    
    // Should have cookie consent options
    await expect(page.locator('[data-testid="cookie-accept-all"]')).toBeVisible();
    await expect(page.locator('[data-testid="cookie-reject-all"]')).toBeVisible();
    await expect(page.locator('[data-testid="cookie-customize"]')).toBeVisible();
  });

  test('should display CCPA compliance notice for California users', async ({ page, context }) => {
    // Simulate California user
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'us-west-1',
      'X-Test-State': 'CA',
      'Accept-Language': 'en-US,en;q=0.9'
    });
    
    await page.goto('/');
    
    // Should display CCPA compliance notice
    await expect(page.locator('[data-testid="ccpa-notice"]')).toBeVisible();
    await expect(page.locator('[data-testid="ccpa-notice"]')).toContainText('Do Not Sell My Personal Information');
  });

  test('should display PIPL compliance for Chinese users', async ({ page, context }) => {
    // Simulate Chinese user
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'ap-east-1',
      'Accept-Language': 'zh-CN,zh;q=0.9'
    });
    
    await page.goto('/');
    
    // Should display PIPL compliance notice
    await expect(page.locator('[data-testid="pipl-notice"]')).toBeVisible();
    await expect(page.locator('[data-testid="pipl-notice"]')).toContainText('个人信息保护');
  });

  test('should allow users to manage cookie preferences', async ({ page }) => {
    // Navigate to cookie preferences
    await page.goto('/');
    
    // Open cookie settings
    await page.click('[data-testid="cookie-customize"]');
    
    // Should display cookie categories
    await expect(page.locator('[data-testid="cookie-essential"]')).toBeVisible();
    await expect(page.locator('[data-testid="cookie-analytics"]')).toBeVisible();
    await expect(page.locator('[data-testid="cookie-marketing"]')).toBeVisible();
    
    // Essential cookies should be disabled (required)
    await expect(page.locator('[data-testid="cookie-essential"] input')).toBeDisabled();
    await expect(page.locator('[data-testid="cookie-essential"] input')).toBeChecked();
    
    // Analytics cookies should be toggleable
    await page.click('[data-testid="cookie-analytics"] input');
    await expect(page.locator('[data-testid="cookie-analytics"] input')).not.toBeChecked();
    
    // Save preferences
    await page.click('[data-testid="save-cookie-preferences"]');
    
    // Verify preferences are saved
    await expect(page.locator('[data-testid="cookie-preferences-saved"]')).toBeVisible();
  });

  test('should provide data export functionality for GDPR compliance', async ({ page }) => {
    // Login as test user
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'user.eu@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to privacy settings
    await page.goto('/settings/privacy');
    
    // Should have data export option
    await expect(page.locator('[data-testid="export-data-button"]')).toBeVisible();
    
    // Click export data
    await page.click('[data-testid="export-data-button"]');
    
    // Should show export confirmation
    await expect(page.locator('[data-testid="export-confirmation"]')).toBeVisible();
    
    // Confirm export
    await page.click('[data-testid="confirm-export"]');
    
    // Should show export in progress
    await expect(page.locator('[data-testid="export-in-progress"]')).toBeVisible();
  });

  test('should provide data deletion functionality for GDPR compliance', async ({ page }) => {
    // Login as test user
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'user.eu@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to privacy settings
    await page.goto('/settings/privacy');
    
    // Should have data deletion option
    await expect(page.locator('[data-testid="delete-account-button"]')).toBeVisible();
    
    // Click delete account
    await page.click('[data-testid="delete-account-button"]');
    
    // Should show deletion warning
    await expect(page.locator('[data-testid="deletion-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="deletion-warning"]')).toContainText('permanent');
    
    // Should require confirmation
    await expect(page.locator('[data-testid="deletion-confirmation-input"]')).toBeVisible();
    await page.fill('[data-testid="deletion-confirmation-input"]', 'DELETE');
    
    // Confirm deletion button should be enabled
    await expect(page.locator('[data-testid="confirm-deletion"]')).toBeEnabled();
  });

  test('should handle consent withdrawal for marketing communications', async ({ page }) => {
    // Login as test user
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'user.eu@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to communication preferences
    await page.goto('/settings/communications');
    
    // Should show current consent status
    await expect(page.locator('[data-testid="marketing-consent"]')).toBeVisible();
    
    // Withdraw consent
    await page.click('[data-testid="marketing-consent"] input');
    await expect(page.locator('[data-testid="marketing-consent"] input')).not.toBeChecked();
    
    // Save changes
    await page.click('[data-testid="save-communication-preferences"]');
    
    // Should show confirmation
    await expect(page.locator('[data-testid="preferences-saved"]')).toBeVisible();
  });

  test('should display appropriate data retention policies', async ({ page, context }) => {
    // Test for EU region (GDPR)
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'eu-central-1'
    });
    
    await page.goto('/privacy-policy');
    
    // Should display GDPR data retention period
    await expect(page.locator('[data-testid="data-retention-period"]')).toContainText('2 years');
    
    // Test for US region (CCPA)
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'us-west-1'
    });
    
    await page.reload();
    
    // Should display CCPA data retention period
    await expect(page.locator('[data-testid="data-retention-period"]')).toContainText('1 year');
  });

  test('should handle age verification for different regions', async ({ page, context }) => {
    // Test for EU region (GDPR - 16 years)
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'eu-central-1'
    });
    
    await page.goto('/register');
    
    // Should show age verification with 16 as minimum
    await expect(page.locator('[data-testid="age-verification"]')).toBeVisible();
    await expect(page.locator('[data-testid="minimum-age"]')).toContainText('16');
    
    // Test for US region (COPPA - 13 years)
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'us-east-1'
    });
    
    await page.reload();
    
    // Should show age verification with 13 as minimum
    await expect(page.locator('[data-testid="minimum-age"]')).toContainText('13');
  });

  test('should provide audit trail for compliance actions', async ({ page }) => {
    // Login as admin
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to compliance audit
    await page.goto('/admin/compliance/audit');
    
    // Should display audit trail
    await expect(page.locator('[data-testid="audit-trail"]')).toBeVisible();
    
    // Should have filterable events
    await expect(page.locator('[data-testid="audit-filter-consent"]')).toBeVisible();
    await expect(page.locator('[data-testid="audit-filter-data-export"]')).toBeVisible();
    await expect(page.locator('[data-testid="audit-filter-data-deletion"]')).toBeVisible();
    
    // Filter by consent events
    await page.click('[data-testid="audit-filter-consent"]');
    
    // Should show only consent-related events
    const auditEntries = page.locator('[data-testid="audit-entry"]');
    await expect(auditEntries.first()).toContainText('consent');
  });

  test('should handle cross-border data transfer notifications', async ({ page, context }) => {
    // Simulate EU user accessing US-hosted data
    await context.setExtraHTTPHeaders({
      'X-Test-Region': 'eu-central-1',
      'X-Data-Location': 'us-east-1'
    });
    
    await page.goto('/dashboard');
    
    // Should display cross-border transfer notification
    await expect(page.locator('[data-testid="cross-border-notice"]')).toBeVisible();
    await expect(page.locator('[data-testid="cross-border-notice"]')).toContainText('adequacy decision');
  });

  test('should enforce data minimization principles', async ({ page }) => {
    // Navigate to registration form
    await page.goto('/register');
    
    // Should only request necessary information
    await expect(page.locator('[data-testid="email-field"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-field"]')).toBeVisible();
    
    // Optional fields should be clearly marked
    const optionalFields = page.locator('[data-testid*="optional"]');
    const optionalCount = await optionalFields.count();
    
    // Should have clear indication of optional vs required fields
    if (optionalCount > 0) {
      await expect(optionalFields.first()).toContainText('optional');
    }
    
    // Should explain why each piece of data is collected
    await expect(page.locator('[data-testid="data-collection-purpose"]')).toBeVisible();
  });
});