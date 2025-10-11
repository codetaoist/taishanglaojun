import { test, expect } from '@playwright/test';

/**
 * Security E2E Tests
 * Tests security features, authentication, authorization, and data protection
 */
test.describe('Security Features @security @global', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should enforce HTTPS and security headers', async ({ page }) => {
    // Check if page is served over HTTPS in production
    const url = page.url();
    if (process.env.NODE_ENV === 'production') {
      expect(url).toMatch(/^https:/);
    }
    
    // Check security headers
    const response = await page.goto('/');
    const headers = response?.headers() || {};
    
    // Should have security headers
    expect(headers['x-frame-options']).toBeDefined();
    expect(headers['x-content-type-options']).toBe('nosniff');
    expect(headers['x-xss-protection']).toBeDefined();
    expect(headers['strict-transport-security']).toBeDefined();
    expect(headers['content-security-policy']).toBeDefined();
  });

  test('should implement proper authentication flow', async ({ page }) => {
    // Should redirect to login when accessing protected route
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/.*\/login/);
    
    // Should show login form
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
    await expect(page.locator('[data-testid="email"]')).toBeVisible();
    await expect(page.locator('[data-testid="password"]')).toBeVisible();
    
    // Should validate email format
    await page.fill('[data-testid="email"]', 'invalid-email');
    await page.click('[data-testid="login-button"]');
    await expect(page.locator('[data-testid="email-error"]')).toBeVisible();
    
    // Should validate password requirements
    await page.fill('[data-testid="email"]', 'test@example.com');
    await page.fill('[data-testid="password"]', '123');
    await page.click('[data-testid="login-button"]');
    await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
  });

  test('should handle failed login attempts securely', async ({ page }) => {
    await page.goto('/login');
    
    // Attempt multiple failed logins
    for (let i = 0; i < 3; i++) {
      await page.fill('[data-testid="email"]', 'test@example.com');
      await page.fill('[data-testid="password"]', 'wrongpassword');
      await page.click('[data-testid="login-button"]');
      
      await expect(page.locator('[data-testid="login-error"]')).toBeVisible();
    }
    
    // Should implement rate limiting after multiple failures
    await page.fill('[data-testid="email"]', 'test@example.com');
    await page.fill('[data-testid="password"]', 'wrongpassword');
    await page.click('[data-testid="login-button"]');
    
    // Should show rate limit message or CAPTCHA
    const rateLimitMessage = page.locator('[data-testid="rate-limit-error"]');
    const captcha = page.locator('[data-testid="captcha"]');
    
    const hasRateLimit = await rateLimitMessage.count() > 0;
    const hasCaptcha = await captcha.count() > 0;
    
    expect(hasRateLimit || hasCaptcha).toBeTruthy();
  });

  test('should implement secure session management', async ({ page, context }) => {
    // Login successfully
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    await expect(page).toHaveURL(/.*\/dashboard/);
    
    // Check if session cookie is secure
    const cookies = await context.cookies();
    const sessionCookie = cookies.find(cookie => 
      cookie.name.includes('session') || cookie.name.includes('auth')
    );
    
    if (sessionCookie) {
      expect(sessionCookie.secure).toBeTruthy();
      expect(sessionCookie.httpOnly).toBeTruthy();
      expect(sessionCookie.sameSite).toBe('Strict');
    }
    
    // Test session timeout
    // Note: This would require configuring a short session timeout for testing
    // await page.waitForTimeout(60000); // Wait for session to expire
    // await page.reload();
    // await expect(page).toHaveURL(/.*\/login/);
  });

  test('should implement proper authorization controls', async ({ page }) => {
    // Login as regular user
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'user.us@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Should access user dashboard
    await expect(page).toHaveURL(/.*\/dashboard/);
    
    // Should not be able to access admin routes
    await page.goto('/admin');
    await expect(page.locator('[data-testid="access-denied"]')).toBeVisible();
    
    // Should not see admin navigation items
    await expect(page.locator('[data-testid="admin-nav"]')).not.toBeVisible();
  });

  test('should protect against XSS attacks', async ({ page }) => {
    // Test input sanitization
    await page.goto('/profile');
    
    const xssPayload = '<script>alert("XSS")</script>';
    
    // Try to inject XSS in profile name
    await page.fill('[data-testid="profile-name"]', xssPayload);
    await page.click('[data-testid="save-profile"]');
    
    // Should sanitize the input
    const profileName = await page.locator('[data-testid="profile-name"]').inputValue();
    expect(profileName).not.toContain('<script>');
    
    // Check if content is properly escaped in display
    await page.reload();
    const displayedName = await page.locator('[data-testid="displayed-name"]').textContent();
    expect(displayedName).not.toContain('<script>');
  });

  test('should protect against CSRF attacks', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Check if forms include CSRF tokens
    await page.goto('/settings');
    
    const form = page.locator('[data-testid="settings-form"]');
    const csrfToken = form.locator('input[name="_token"], input[name="csrf_token"]');
    
    // Should have CSRF token in forms
    await expect(csrfToken).toBeAttached();
    
    // Token should not be empty
    const tokenValue = await csrfToken.getAttribute('value');
    expect(tokenValue).toBeTruthy();
    expect(tokenValue?.length).toBeGreaterThan(10);
  });

  test('should implement secure password requirements', async ({ page }) => {
    await page.goto('/register');
    
    // Test weak password
    await page.fill('[data-testid="password"]', '123');
    await page.fill('[data-testid="confirm-password"]', '123');
    
    // Should show password strength indicator
    await expect(page.locator('[data-testid="password-strength"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-strength"]')).toContainText('Weak');
    
    // Test medium strength password
    await page.fill('[data-testid="password"]', 'password123');
    await expect(page.locator('[data-testid="password-strength"]')).toContainText('Medium');
    
    // Test strong password
    await page.fill('[data-testid="password"]', 'StrongP@ssw0rd123!');
    await expect(page.locator('[data-testid="password-strength"]')).toContainText('Strong');
    
    // Should enforce minimum requirements
    await page.fill('[data-testid="password"]', 'weak');
    await page.click('[data-testid="register-button"]');
    await expect(page.locator('[data-testid="password-requirements-error"]')).toBeVisible();
  });

  test('should implement secure file upload', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    await page.goto('/upload');
    
    // Should validate file types
    const fileInput = page.locator('[data-testid="file-input"]');
    
    // Check if file type restrictions are in place
    const acceptAttribute = await fileInput.getAttribute('accept');
    expect(acceptAttribute).toBeTruthy();
    
    // Should have file size limits
    await expect(page.locator('[data-testid="file-size-limit"]')).toBeVisible();
    
    // Should scan for malicious content
    // Note: This would require actual file upload testing with test files
  });

  test('should implement secure API communication', async ({ page }) => {
    // Monitor API requests
    const apiRequests = [];
    
    page.on('request', request => {
      if (request.url().includes('/api/')) {
        apiRequests.push({
          url: request.url(),
          method: request.method(),
          headers: request.headers()
        });
      }
    });
    
    // Login and make API calls
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Check API request security
    apiRequests.forEach(request => {
      // Should include authentication headers
      expect(request.headers.authorization || request.headers.cookie).toBeTruthy();
      
      // Should use HTTPS in production
      if (process.env.NODE_ENV === 'production') {
        expect(request.url).toMatch(/^https:/);
      }
    });
  });

  test('should handle sensitive data securely', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Navigate to sensitive data page
    await page.goto('/profile/sensitive');
    
    // Should mask sensitive information
    const maskedFields = page.locator('[data-testid*="masked"]');
    const maskedCount = await maskedFields.count();
    
    if (maskedCount > 0) {
      const maskedValue = await maskedFields.first().textContent();
      expect(maskedValue).toMatch(/\*+/); // Should contain asterisks
    }
    
    // Should require additional authentication for sensitive operations
    await page.click('[data-testid="view-sensitive-data"]');
    
    const authPrompt = page.locator('[data-testid="auth-prompt"]');
    const hasAuthPrompt = await authPrompt.count() > 0;
    
    if (hasAuthPrompt) {
      await expect(authPrompt).toBeVisible();
    }
  });

  test('should implement secure logout', async ({ page, context }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    await expect(page).toHaveURL(/.*\/dashboard/);
    
    // Logout
    await page.click('[data-testid="logout-button"]');
    
    // Should redirect to login page
    await expect(page).toHaveURL(/.*\/login/);
    
    // Should clear session cookies
    const cookies = await context.cookies();
    const sessionCookies = cookies.filter(cookie => 
      cookie.name.includes('session') || cookie.name.includes('auth')
    );
    
    // Session cookies should be cleared or expired
    sessionCookies.forEach(cookie => {
      expect(cookie.value).toBeFalsy();
    });
    
    // Should not be able to access protected routes
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/.*\/login/);
  });

  test('should protect against SQL injection', async ({ page }) => {
    // Test search functionality with SQL injection attempts
    await page.goto('/search');
    
    const sqlInjectionPayloads = [
      "'; DROP TABLE users; --",
      "' OR '1'='1",
      "' UNION SELECT * FROM users --"
    ];
    
    for (const payload of sqlInjectionPayloads) {
      await page.fill('[data-testid="search-input"]', payload);
      await page.click('[data-testid="search-button"]');
      
      // Should not return unexpected results or errors
      const errorMessage = page.locator('[data-testid="sql-error"]');
      const hasError = await errorMessage.count() > 0;
      
      // Should not expose SQL errors
      expect(hasError).toBeFalsy();
      
      // Should sanitize the search query
      const searchResults = page.locator('[data-testid="search-results"]');
      const resultsText = await searchResults.textContent();
      expect(resultsText).not.toContain('DROP TABLE');
      expect(resultsText).not.toContain('UNION SELECT');
    }
  });

  test('should implement proper error handling without information disclosure', async ({ page }) => {
    // Test 404 error
    await page.goto('/nonexistent-page');
    
    // Should show generic error message
    await expect(page.locator('[data-testid="404-error"]')).toBeVisible();
    
    // Should not expose system information
    const pageContent = await page.textContent('body');
    expect(pageContent).not.toMatch(/stack trace|debug|internal error/i);
    
    // Test API error handling
    await page.goto('/dashboard');
    
    // Simulate API error by intercepting requests
    await page.route('**/api/**', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' })
      });
    });
    
    await page.reload();
    
    // Should show user-friendly error message
    const errorMessage = page.locator('[data-testid="api-error"]');
    const hasError = await errorMessage.count() > 0;
    
    if (hasError) {
      const errorText = await errorMessage.textContent();
      expect(errorText).not.toMatch(/database|sql|server|stack/i);
    }
  });
});