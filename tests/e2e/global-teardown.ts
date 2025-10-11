import axios from 'axios';

/**
 * Global teardown for E2E tests
 * Cleans up test environment and data
 */
async function globalTeardown() {
  console.log('🧹 Starting global teardown for Taishang Laojun E2E tests...');
  
  const apiBaseURL = process.env.API_BASE_URL || 'http://localhost:8080';
  
  try {
    // Clean up test data
    await cleanupTestData(apiBaseURL);
    
    // Generate test reports
    await generateTestReports();
    
    console.log('✅ Global teardown completed successfully');
  } catch (error) {
    console.warn('⚠️ Global teardown encountered issues:', error.message);
  }
}

/**
 * Clean up test data
 */
async function cleanupTestData(apiBaseURL: string) {
  console.log('🗑️ Cleaning up test data...');
  
  try {
    // Remove test users
    await axios.delete(`${apiBaseURL}/api/test/users`);
    
    // Clean up test sessions
    await axios.delete(`${apiBaseURL}/api/test/sessions`);
    
    // Clean up test files
    await axios.delete(`${apiBaseURL}/api/test/files`);
    
    console.log('✅ Test data cleanup completed');
  } catch (error) {
    console.warn('⚠️ Test data cleanup failed:', error.message);
  }
}

/**
 * Generate test reports
 */
async function generateTestReports() {
  console.log('📊 Generating test reports...');
  
  try {
    const fs = require('fs');
    const path = require('path');
    
    // Create reports directory if it doesn't exist
    const reportsDir = path.join(__dirname, 'reports');
    if (!fs.existsSync(reportsDir)) {
      fs.mkdirSync(reportsDir, { recursive: true });
    }
    
    // Generate summary report
    const summary = {
      timestamp: new Date().toISOString(),
      testRun: process.env.TEST_RUN_ID || 'local',
      environment: process.env.NODE_ENV || 'test',
      reports: {
        html: 'playwright-report/index.html',
        json: 'test-results.json',
        junit: 'test-results.xml',
        allure: 'allure-report/index.html'
      }
    };
    
    fs.writeFileSync(
      path.join(reportsDir, 'test-summary.json'),
      JSON.stringify(summary, null, 2)
    );
    
    console.log('✅ Test reports generated');
  } catch (error) {
    console.warn('⚠️ Test report generation failed:', error.message);
  }
}

export default globalTeardown;