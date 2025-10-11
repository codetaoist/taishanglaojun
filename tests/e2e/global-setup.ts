import { chromium, FullConfig } from '@playwright/test';
import axios from 'axios';

/**
 * Global setup for E2E tests
 * Prepares test environment and validates services
 */
async function globalSetup(config: FullConfig) {
  console.log('🚀 Starting global setup for Taishang Laojun E2E tests...');
  
  const baseURL = config.projects[0].use.baseURL || 'http://localhost:5173';
  const apiBaseURL = process.env.API_BASE_URL || 'http://localhost:8080';
  
  // Wait for services to be ready
  await waitForService(baseURL, 'Frontend Application');
  await waitForService(`${apiBaseURL}/health`, 'Core Services API');
  await waitForService(`${apiBaseURL}/api/localization/health`, 'Localization Service');
  await waitForService(`${apiBaseURL}/api/compliance/health`, 'Compliance Service');
  
  // Setup test data
  await setupTestData(apiBaseURL);
  
  // Create global browser context for authentication
  const browser = await chromium.launch();
  const context = await browser.newContext();
  
  // Perform global authentication if needed
  await performGlobalAuth(context, baseURL);
  
  // Save authentication state
  await context.storageState({ path: 'tests/auth/admin-auth.json' });
  await context.storageState({ path: 'tests/auth/user-auth.json' });
  
  await browser.close();
  
  console.log('✅ Global setup completed successfully');
}

/**
 * Wait for a service to be ready
 */
async function waitForService(url: string, serviceName: string, maxRetries = 30) {
  console.log(`⏳ Waiting for ${serviceName} at ${url}...`);
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await axios.get(url, { timeout: 5000 });
      if (response.status === 200) {
        console.log(`✅ ${serviceName} is ready`);
        return;
      }
    } catch (error) {
      console.log(`⏳ ${serviceName} not ready yet, retrying... (${i + 1}/${maxRetries})`);
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  }
  
  throw new Error(`❌ ${serviceName} failed to start after ${maxRetries} retries`);
}

/**
 * Setup test data
 */
async function setupTestData(apiBaseURL: string) {
  console.log('📊 Setting up test data...');
  
  try {
    // Create test users
    await axios.post(`${apiBaseURL}/api/test/users`, {
      users: [
        {
          id: 'test-admin',
          email: 'admin@test.com',
          role: 'admin',
          preferences: {
            language: 'zh-CN',
            timezone: 'Asia/Shanghai',
            region: 'CN'
          }
        },
        {
          id: 'test-user-us',
          email: 'user.us@test.com',
          role: 'user',
          preferences: {
            language: 'en-US',
            timezone: 'America/New_York',
            region: 'US'
          }
        },
        {
          id: 'test-user-eu',
          email: 'user.eu@test.com',
          role: 'user',
          preferences: {
            language: 'de-DE',
            timezone: 'Europe/Berlin',
            region: 'DE'
          }
        }
      ]
    });
    
    // Setup localization test data
    await axios.post(`${apiBaseURL}/api/test/localization`, {
      translations: {
        'en-US': {
          'welcome': 'Welcome to Taishang Laojun AI Platform',
          'dashboard': 'Dashboard',
          'settings': 'Settings'
        },
        'zh-CN': {
          'welcome': '欢迎使用太上老君AI平台',
          'dashboard': '仪表板',
          'settings': '设置'
        },
        'de-DE': {
          'welcome': 'Willkommen bei der Taishang Laojun AI-Plattform',
          'dashboard': 'Dashboard',
          'settings': 'Einstellungen'
        }
      }
    });
    
    // Setup compliance test data
    await axios.post(`${apiBaseURL}/api/test/compliance`, {
      regions: {
        'US': { regulation: 'CCPA', dataRetention: 365 },
        'EU': { regulation: 'GDPR', dataRetention: 730 },
        'CN': { regulation: 'PIPL', dataRetention: 365 }
      }
    });
    
    console.log('✅ Test data setup completed');
  } catch (error) {
    console.warn('⚠️ Test data setup failed, continuing with existing data:', error.message);
  }
}

/**
 * Perform global authentication
 */
async function performGlobalAuth(context: any, baseURL: string) {
  console.log('🔐 Setting up authentication...');
  
  const page = await context.newPage();
  
  try {
    // Navigate to login page
    await page.goto(`${baseURL}/login`);
    
    // Perform admin login
    await page.fill('[data-testid="email"]', 'admin@test.com');
    await page.fill('[data-testid="password"]', 'test123456');
    await page.click('[data-testid="login-button"]');
    
    // Wait for successful login
    await page.waitForURL(`${baseURL}/dashboard`, { timeout: 10000 });
    
    console.log('✅ Authentication setup completed');
  } catch (error) {
    console.warn('⚠️ Authentication setup failed, tests may need manual login:', error.message);
  } finally {
    await page.close();
  }
}

export default globalSetup;