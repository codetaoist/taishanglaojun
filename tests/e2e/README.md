# E2E Tests for Taishang Laojun AI Platform

This directory contains comprehensive end-to-end tests for the Taishang Laojun AI Platform, focusing on global features, localization, compliance, performance, and security.

## 🚀 Quick Start

### Prerequisites

- Node.js 18 or later
- npm or yarn
- PowerShell (for Windows scripts)

### Installation

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install
```

### Running Tests

```bash
# Run all tests
npm test

# Run specific test suite
npm run test:localization
npm run test:compliance
npm run test:performance
npm run test:security
npm run test:mobile
npm run test:cross-browser
npm run test:global-integration

# Run tests with specific browser
npm run test:chrome
npm run test:firefox
npm run test:safari

# Run tests in headed mode (visible browser)
npm run test:headed

# Run tests with debug mode
npm run test:debug
```

### Using PowerShell Script

```powershell
# Run all tests
.\scripts\run-tests.ps1

# Run specific test suite
.\scripts\run-tests.ps1 -TestSuite localization

# Run with specific browser
.\scripts\run-tests.ps1 -Browser firefox -Headed

# Run in CI mode
.\scripts\run-tests.ps1 -CI -Reporter allure

# Get help
.\scripts\run-tests.ps1 -Help
```

## 📁 Project Structure

```
tests/e2e/
├── tests/                          # Test files
│   ├── localization.spec.ts        # Localization tests
│   ├── compliance.spec.ts          # GDPR/CCPA/PIPL compliance tests
│   ├── performance.spec.ts         # Performance and Core Web Vitals tests
│   ├── security.spec.ts            # Security and authentication tests
│   ├── mobile-responsive.spec.ts   # Mobile and responsive design tests
│   ├── cross-browser.spec.ts       # Cross-browser compatibility tests
│   └── global-integration.spec.ts  # End-to-end global integration tests
├── utils/                          # Test utilities
│   └── test-helpers.ts             # Common test helper functions
├── fixtures/                       # Test data and fixtures
│   └── test-data.ts                # Test data constants and fixtures
├── scripts/                        # Test scripts
│   └── run-tests.ps1               # PowerShell test runner script
├── playwright.config.ts            # Playwright configuration
├── global-setup.ts                 # Global test setup
├── global-teardown.ts              # Global test teardown
├── package.json                    # Dependencies and scripts
├── .env.example                    # Environment variables example
└── README.md                       # This file
```

## 🧪 Test Suites

### 1. Localization Tests (`localization.spec.ts`)

Tests the internationalization and localization features:

- **Language Switching**: Verify content changes when switching languages
- **Date/Time Formatting**: Check locale-specific date and time formats
- **Currency Display**: Validate currency formatting for different regions
- **Number Formatting**: Test number formatting conventions
- **RTL Language Support**: Right-to-left language layout support
- **Font Loading**: Ensure appropriate fonts load for different languages
- **Pluralization**: Test plural forms in different languages

**Key Features Tested:**
- Chinese (Simplified/Traditional), English, German, Japanese, Korean
- Timezone handling across regions
- Currency conversion and display
- Cultural preferences

### 2. Compliance Tests (`compliance.spec.ts`)

Tests data privacy and regulatory compliance:

- **GDPR Compliance** (EU):
  - Cookie consent banners
  - Data export functionality
  - Right to be forgotten (data deletion)
  - Consent withdrawal
  - Data retention policies
  - Age verification
  - Audit trails

- **CCPA Compliance** (California):
  - Privacy notices
  - Opt-out mechanisms
  - Data deletion requests

- **PIPL Compliance** (China):
  - Data processing consent
  - Cross-border data transfer notices

**Key Features Tested:**
- Consent management
- Data subject rights
- Privacy policy display
- Compliance workflow automation

### 3. Performance Tests (`performance.spec.ts`)

Tests application performance and Core Web Vitals:

- **Core Web Vitals**:
  - First Contentful Paint (FCP)
  - Largest Contentful Paint (LCP)
  - First Input Delay (FID)
  - Cumulative Layout Shift (CLS)

- **Load Performance**:
  - Page load times
  - Resource loading optimization
  - Bundle size analysis
  - Code splitting effectiveness

- **Runtime Performance**:
  - Memory usage
  - CPU utilization
  - Network efficiency
  - Caching strategies

**Key Features Tested:**
- Performance budgets
- Image optimization
- Lazy loading
- Virtual scrolling
- API response times

### 4. Security Tests (`security.spec.ts`)

Tests security features and vulnerabilities:

- **Authentication & Authorization**:
  - Login/logout flows
  - Session management
  - Role-based access control
  - Multi-factor authentication

- **Security Headers**:
  - HTTPS enforcement
  - Content Security Policy (CSP)
  - X-Frame-Options
  - X-Content-Type-Options

- **Input Validation**:
  - XSS protection
  - CSRF protection
  - SQL injection prevention
  - File upload security

**Key Features Tested:**
- Secure session handling
- Password policies
- API security
- Data encryption
- Error handling

### 5. Mobile Responsive Tests (`mobile-responsive.spec.ts`)

Tests mobile device compatibility and responsive design:

- **Device Compatibility**:
  - iPhone, Android, tablet support
  - Various screen sizes and resolutions
  - Touch interactions and gestures

- **Responsive Design**:
  - Layout adaptation
  - Navigation patterns
  - Form interactions
  - Image optimization

- **Mobile Performance**:
  - Load times on mobile networks
  - Touch target sizes
  - Accessibility on mobile

**Key Features Tested:**
- Mobile navigation
- Touch gestures
- Orientation changes
- Mobile-specific features
- Cross-device consistency

### 6. Cross-Browser Tests (`cross-browser.spec.ts`)

Tests compatibility across different browsers:

- **Browser Support**:
  - Chrome, Firefox, Safari, Edge
  - JavaScript API compatibility
  - CSS feature support

- **Functionality Consistency**:
  - Core features across browsers
  - Form handling
  - Media playback
  - Animation support

**Key Features Tested:**
- JavaScript compatibility
- CSS rendering consistency
- Performance across browsers
- Error handling
- Accessibility standards

### 7. Global Integration Tests (`global-integration.spec.ts`)

Tests complete user journeys across global features:

- **End-to-End Workflows**:
  - User registration with regional compliance
  - Multi-language content creation
  - Cross-border data handling
  - Global search functionality

- **Regional Migration**:
  - Data transfer between regions
  - Compliance updates during migration
  - Service continuity

**Key Features Tested:**
- Complete user journeys
- Global feature integration
- Disaster recovery scenarios
- Multi-region workflows

## 🔧 Configuration

### Environment Variables

Copy `.env.example` to `.env.test` and configure:

```bash
# Application URLs
BASE_URL=http://localhost:5173
API_BASE_URL=http://localhost:8080

# Test Configuration
NODE_ENV=test
HEADLESS=true
TIMEOUT=30000

# Regional Testing
DEFAULT_REGION=us-east-1
TEST_REGIONS=us-east-1,eu-central-1,ap-east-1

# Compliance Testing
GDPR_ENABLED=true
CCPA_ENABLED=true
PIPL_ENABLED=true

# Performance Budgets
PERFORMANCE_BUDGET_FCP=1800
PERFORMANCE_BUDGET_LCP=2500
PERFORMANCE_BUDGET_LOAD_TIME=3000
```

### Playwright Configuration

The `playwright.config.ts` file includes:

- **Multiple Browser Support**: Chromium, Firefox, WebKit, Chrome, Edge
- **Device Testing**: Mobile and tablet configurations
- **Global Testing**: Region-specific configurations
- **Parallel Execution**: Optimized for CI/CD
- **Reporting**: HTML, JSON, JUnit, Allure reports

### Test Data Management

Test data is managed in `fixtures/test-data.ts`:

- **User Accounts**: Different user types for various scenarios
- **Regional Data**: Region-specific configurations
- **Localization Data**: Language and locale information
- **Compliance Data**: Privacy and consent configurations

## 📊 Reporting

### HTML Reports

```bash
# Generate HTML report
npm run test:report

# Open report
npm run test:show-report
```

### Allure Reports

```bash
# Generate Allure report
npm run test:allure

# Serve Allure report
npm run test:allure:serve
```

### CI/CD Integration

```bash
# Run tests in CI mode
npm run test:ci

# Generate CI-friendly reports
npm run test:ci:report
```

## 🚀 CI/CD Integration

### GitHub Actions

```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - name: Install dependencies
        run: npm ci
      - name: Install Playwright
        run: npx playwright install --with-deps
      - name: Run E2E tests
        run: npm run test:ci
      - name: Upload test results
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-results
          path: test-results/
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Install') {
            steps {
                sh 'npm ci'
                sh 'npx playwright install --with-deps'
            }
        }
        stage('Test') {
            steps {
                sh 'npm run test:ci'
            }
            post {
                always {
                    publishHTML([
                        allowMissing: false,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'test-results',
                        reportFiles: 'index.html',
                        reportName: 'E2E Test Report'
                    ])
                }
            }
        }
    }
}
```

## 🔍 Debugging

### Debug Mode

```bash
# Run tests in debug mode
npm run test:debug

# Debug specific test
npx playwright test tests/localization.spec.ts --debug
```

### Visual Debugging

```bash
# Run tests in headed mode
npm run test:headed

# Update visual snapshots
npm run test:update-snapshots
```

### Trace Viewer

```bash
# Generate traces
npm run test:trace

# View traces
npx playwright show-trace trace.zip
```

## 📈 Performance Monitoring

### Core Web Vitals Budgets

- **FCP (First Contentful Paint)**: < 1.8s
- **LCP (Largest Contentful Paint)**: < 2.5s
- **FID (First Input Delay)**: < 100ms
- **CLS (Cumulative Layout Shift)**: < 0.1

### Performance Testing

```bash
# Run performance tests
npm run test:performance

# Run with network throttling
npm run test:performance:slow-3g
```

## 🛡️ Security Testing

### Security Checklist

- ✅ HTTPS enforcement
- ✅ Security headers validation
- ✅ Authentication flow testing
- ✅ Authorization checks
- ✅ Input validation
- ✅ XSS protection
- ✅ CSRF protection
- ✅ File upload security

### Security Testing

```bash
# Run security tests
npm run test:security

# Run with security audit
npm run test:security:audit
```

## 🌍 Global Testing

### Multi-Region Testing

```bash
# Test US region
npm run test:region:us

# Test EU region
npm run test:region:eu

# Test Asia region
npm run test:region:asia
```

### Localization Testing

```bash
# Test Chinese localization
npm run test:locale:zh-CN

# Test English localization
npm run test:locale:en-US

# Test German localization
npm run test:locale:de-DE
```

## 📱 Mobile Testing

### Device Testing

```bash
# Test on mobile devices
npm run test:mobile

# Test on tablets
npm run test:tablet

# Test responsive design
npm run test:responsive
```

## 🤝 Contributing

### Adding New Tests

1. Create test file in `tests/` directory
2. Follow naming convention: `feature-name.spec.ts`
3. Use test helpers from `utils/test-helpers.ts`
4. Add test data to `fixtures/test-data.ts`
5. Update this README with test description

### Test Guidelines

- **Descriptive Names**: Use clear, descriptive test names
- **Independent Tests**: Each test should be independent
- **Data Cleanup**: Clean up test data after tests
- **Error Handling**: Handle errors gracefully
- **Performance**: Keep tests fast and efficient

### Code Style

- Use TypeScript for type safety
- Follow ESLint configuration
- Use Prettier for code formatting
- Add JSDoc comments for complex functions

## 📞 Support

For questions or issues with E2E tests:

1. Check existing test documentation
2. Review test logs and reports
3. Create issue with test details
4. Contact the QA team

## 📚 Resources

- [Playwright Documentation](https://playwright.dev/)
- [Testing Best Practices](https://playwright.dev/docs/best-practices)
- [Accessibility Testing](https://playwright.dev/docs/accessibility-testing)
- [Performance Testing](https://playwright.dev/docs/performance)
- [Visual Testing](https://playwright.dev/docs/test-snapshots)

---

**Happy Testing! 🧪✨**