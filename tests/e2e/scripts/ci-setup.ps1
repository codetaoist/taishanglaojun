# CI/CD Setup Script for E2E Tests
# This script sets up the environment for running E2E tests in CI/CD pipelines

param(
    [string]$Environment = "ci",
    [string]$Region = "us-east-1",
    [switch]$InstallDependencies = $false,
    [switch]$SetupServices = $false,
    [switch]$Verbose = $false,
    [switch]$Help = $false
)

# Show help
if ($Help) {
    Write-Host @"
CI/CD Setup Script for E2E Tests

USAGE:
    .\ci-setup.ps1 [OPTIONS]

OPTIONS:
    -Environment <env>      Target environment (ci, staging, production) [default: ci]
    -Region <region>        AWS region for testing [default: us-east-1]
    -InstallDependencies    Install Node.js dependencies and Playwright browsers
    -SetupServices         Start required services for testing
    -Verbose               Enable verbose output
    -Help                  Show this help message

EXAMPLES:
    .\ci-setup.ps1 -InstallDependencies -SetupServices
    .\ci-setup.ps1 -Environment staging -Region eu-central-1
    .\ci-setup.ps1 -Verbose

"@
    exit 0
}

# Enable verbose output
if ($Verbose) {
    $VerbosePreference = "Continue"
}

Write-Host "🚀 Setting up CI/CD environment for E2E tests..." -ForegroundColor Green

# Function to check if command exists
function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Function to wait for service
function Wait-ForService {
    param(
        [string]$Url,
        [int]$TimeoutSeconds = 60,
        [string]$ServiceName = "Service"
    )
    
    Write-Host "⏳ Waiting for $ServiceName to be ready at $Url..." -ForegroundColor Yellow
    
    $timeout = (Get-Date).AddSeconds($TimeoutSeconds)
    
    while ((Get-Date) -lt $timeout) {
        try {
            $response = Invoke-WebRequest -Uri $Url -Method GET -TimeoutSec 5 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                Write-Host "✅ $ServiceName is ready!" -ForegroundColor Green
                return $true
            }
        }
        catch {
            Write-Verbose "Service not ready yet, retrying..."
        }
        
        Start-Sleep -Seconds 2
    }
    
    Write-Host "❌ $ServiceName failed to start within $TimeoutSeconds seconds" -ForegroundColor Red
    return $false
}

# Check prerequisites
Write-Host "🔍 Checking prerequisites..." -ForegroundColor Blue

$prerequisites = @()

if (-not (Test-Command "node")) {
    $prerequisites += "Node.js"
}

if (-not (Test-Command "npm")) {
    $prerequisites += "npm"
}

if (-not (Test-Command "git")) {
    $prerequisites += "Git"
}

if ($prerequisites.Count -gt 0) {
    Write-Host "❌ Missing prerequisites: $($prerequisites -join ', ')" -ForegroundColor Red
    Write-Host "Please install the missing prerequisites and try again." -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ All prerequisites are installed" -ForegroundColor Green

# Set environment variables
Write-Host "🔧 Setting up environment variables..." -ForegroundColor Blue

$env:NODE_ENV = $Environment
$env:CI = "true"
$env:HEADLESS = "true"
$env:TIMEOUT = "60000"
$env:RETRIES = "2"
$env:WORKERS = "2"
$env:DEFAULT_REGION = $Region

# Environment-specific configurations
switch ($Environment) {
    "ci" {
        $env:BASE_URL = "http://localhost:5173"
        $env:API_BASE_URL = "http://localhost:8080"
        $env:LOCALIZATION_SERVICE_URL = "http://localhost:8081"
        $env:COMPLIANCE_SERVICE_URL = "http://localhost:8082"
    }
    "staging" {
        $env:BASE_URL = "https://staging.taishanglaojun.ai"
        $env:API_BASE_URL = "https://api-staging.taishanglaojun.ai"
        $env:LOCALIZATION_SERVICE_URL = "https://localization-staging.taishanglaojun.ai"
        $env:COMPLIANCE_SERVICE_URL = "https://compliance-staging.taishanglaojun.ai"
    }
    "production" {
        $env:BASE_URL = "https://taishanglaojun.ai"
        $env:API_BASE_URL = "https://api.taishanglaojun.ai"
        $env:LOCALIZATION_SERVICE_URL = "https://localization.taishanglaojun.ai"
        $env:COMPLIANCE_SERVICE_URL = "https://compliance.taishanglaojun.ai"
    }
}

# Regional configurations
switch ($Region) {
    "us-east-1" {
        $env:TEST_TIMEZONE = "America/New_York"
        $env:TEST_LOCALE = "en-US"
        $env:TEST_CURRENCY = "USD"
    }
    "eu-central-1" {
        $env:TEST_TIMEZONE = "Europe/Berlin"
        $env:TEST_LOCALE = "de-DE"
        $env:TEST_CURRENCY = "EUR"
        $env:GDPR_ENABLED = "true"
    }
    "ap-east-1" {
        $env:TEST_TIMEZONE = "Asia/Hong_Kong"
        $env:TEST_LOCALE = "zh-CN"
        $env:TEST_CURRENCY = "CNY"
        $env:PIPL_ENABLED = "true"
    }
}

Write-Host "✅ Environment variables configured for $Environment in $Region" -ForegroundColor Green

# Install dependencies
if ($InstallDependencies) {
    Write-Host "📦 Installing dependencies..." -ForegroundColor Blue
    
    # Check if package.json exists
    if (-not (Test-Path "package.json")) {
        Write-Host "❌ package.json not found. Please run this script from the e2e tests directory." -ForegroundColor Red
        exit 1
    }
    
    # Install npm dependencies
    Write-Host "Installing npm dependencies..." -ForegroundColor Yellow
    npm ci
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to install npm dependencies" -ForegroundColor Red
        exit 1
    }
    
    # Install Playwright browsers
    Write-Host "Installing Playwright browsers..." -ForegroundColor Yellow
    npx playwright install --with-deps
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to install Playwright browsers" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ Dependencies installed successfully" -ForegroundColor Green
}

# Setup services
if ($SetupServices -and $Environment -eq "ci") {
    Write-Host "🚀 Starting services for CI environment..." -ForegroundColor Blue
    
    # Start frontend service
    Write-Host "Starting frontend service..." -ForegroundColor Yellow
    Start-Process -FilePath "npm" -ArgumentList "run", "dev" -WorkingDirectory "../../frontend/web-app" -WindowStyle Hidden
    
    # Start core service
    Write-Host "Starting core service..." -ForegroundColor Yellow
    Start-Process -FilePath "npm" -ArgumentList "start" -WorkingDirectory "../../core-services" -WindowStyle Hidden
    
    # Start localization service
    Write-Host "Starting localization service..." -ForegroundColor Yellow
    Start-Process -FilePath "npm" -ArgumentList "start" -WorkingDirectory "../../localization-service" -WindowStyle Hidden
    
    # Start compliance service
    Write-Host "Starting compliance service..." -ForegroundColor Yellow
    Start-Process -FilePath "npm" -ArgumentList "start" -WorkingDirectory "../../compliance-service" -WindowStyle Hidden
    
    # Wait for services to be ready
    $services = @(
        @{ Url = $env:BASE_URL; Name = "Frontend" },
        @{ Url = "$($env:API_BASE_URL)/health"; Name = "Core API" },
        @{ Url = "$($env:LOCALIZATION_SERVICE_URL)/health"; Name = "Localization Service" },
        @{ Url = "$($env:COMPLIANCE_SERVICE_URL)/health"; Name = "Compliance Service" }
    )
    
    $allServicesReady = $true
    foreach ($service in $services) {
        if (-not (Wait-ForService -Url $service.Url -ServiceName $service.Name -TimeoutSeconds 120)) {
            $allServicesReady = $false
        }
    }
    
    if (-not $allServicesReady) {
        Write-Host "❌ Some services failed to start" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ All services are ready" -ForegroundColor Green
}

# Create test directories
Write-Host "📁 Creating test directories..." -ForegroundColor Blue

$testDirs = @(
    "test-results",
    "test-results/screenshots",
    "test-results/videos",
    "test-results/traces",
    "test-results/reports",
    "test-results/reports/html",
    "test-results/reports/json",
    "test-results/reports/junit",
    "test-results/reports/allure"
)

foreach ($dir in $testDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Verbose "Created directory: $dir"
    }
}

Write-Host "✅ Test directories created" -ForegroundColor Green

# Setup test data
Write-Host "🗃️ Setting up test data..." -ForegroundColor Blue

# Create test environment file
$envContent = @"
# CI/CD Environment Configuration
NODE_ENV=$Environment
CI=true
HEADLESS=true
TIMEOUT=60000
RETRIES=2
WORKERS=2

# Application URLs
BASE_URL=$($env:BASE_URL)
API_BASE_URL=$($env:API_BASE_URL)
LOCALIZATION_SERVICE_URL=$($env:LOCALIZATION_SERVICE_URL)
COMPLIANCE_SERVICE_URL=$($env:COMPLIANCE_SERVICE_URL)

# Regional Configuration
DEFAULT_REGION=$Region
TEST_TIMEZONE=$($env:TEST_TIMEZONE)
TEST_LOCALE=$($env:TEST_LOCALE)
TEST_CURRENCY=$($env:TEST_CURRENCY)

# Compliance Configuration
GDPR_ENABLED=$($env:GDPR_ENABLED)
CCPA_ENABLED=$($env:CCPA_ENABLED)
PIPL_ENABLED=$($env:PIPL_ENABLED)

# Performance Budgets
PERFORMANCE_BUDGET_FCP=1800
PERFORMANCE_BUDGET_LCP=2500
PERFORMANCE_BUDGET_LOAD_TIME=3000

# Test Configuration
PARALLEL_TESTS=true
RETRY_FAILED_TESTS=true
GENERATE_REPORTS=true
UPLOAD_ARTIFACTS=true

# Debug Configuration
DEBUG_MODE=false
TRACE_ON_FAILURE=true
SCREENSHOT_ON_FAILURE=true
VIDEO_ON_FAILURE=true
"@

$envContent | Out-File -FilePath ".env.ci" -Encoding UTF8
Write-Host "✅ Test environment configuration created" -ForegroundColor Green

# Validate configuration
Write-Host "🔍 Validating configuration..." -ForegroundColor Blue

$configErrors = @()

# Check required environment variables
$requiredVars = @("BASE_URL", "API_BASE_URL", "NODE_ENV", "DEFAULT_REGION")
foreach ($var in $requiredVars) {
    if (-not (Get-Item "env:$var" -ErrorAction SilentlyContinue)) {
        $configErrors += "Missing environment variable: $var"
    }
}

# Check URLs are accessible (for non-CI environments)
if ($Environment -ne "ci") {
    try {
        $response = Invoke-WebRequest -Uri $env:BASE_URL -Method HEAD -TimeoutSec 10 -UseBasicParsing
        if ($response.StatusCode -ne 200) {
            $configErrors += "Frontend URL not accessible: $($env:BASE_URL)"
        }
    }
    catch {
        $configErrors += "Frontend URL not accessible: $($env:BASE_URL)"
    }
}

if ($configErrors.Count -gt 0) {
    Write-Host "❌ Configuration validation failed:" -ForegroundColor Red
    foreach ($error in $configErrors) {
        Write-Host "  - $error" -ForegroundColor Red
    }
    exit 1
}

Write-Host "✅ Configuration validation passed" -ForegroundColor Green

# Generate CI configuration files
Write-Host "📄 Generating CI configuration files..." -ForegroundColor Blue

# GitHub Actions workflow
$githubWorkflow = @"
name: E2E Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        browser: [chromium, firefox, webkit]
        region: [us-east-1, eu-central-1, ap-east-1]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Install dependencies
        run: |
          cd tests/e2e
          npm ci
      
      - name: Install Playwright browsers
        run: |
          cd tests/e2e
          npx playwright install --with-deps
      
      - name: Setup CI environment
        run: |
          cd tests/e2e
          ./scripts/ci-setup.ps1 -Environment ci -Region \${{ matrix.region }} -SetupServices
        shell: pwsh
      
      - name: Run E2E tests
        run: |
          cd tests/e2e
          npm run test:ci -- --project=\${{ matrix.browser }}
        env:
          DEFAULT_REGION: \${{ matrix.region }}
      
      - name: Upload test results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: test-results-\${{ matrix.browser }}-\${{ matrix.region }}
          path: tests/e2e/test-results/
          retention-days: 7
      
      - name: Upload Allure results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: allure-results-\${{ matrix.browser }}-\${{ matrix.region }}
          path: tests/e2e/allure-results/
          retention-days: 7

  publish-report:
    needs: e2e-tests
    runs-on: ubuntu-latest
    if: always()
    
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4
      
      - name: Publish Allure Report
        uses: simple-elf/allure-report-action@master
        if: always()
        with:
          allure_results: allure-results-*
          allure_report: allure-report
          gh_pages: gh-pages
"@

if (-not (Test-Path "../../.github/workflows")) {
    New-Item -ItemType Directory -Path "../../.github/workflows" -Force | Out-Null
}
$githubWorkflow | Out-File -FilePath "../../.github/workflows/e2e-tests.yml" -Encoding UTF8

Write-Host "✅ CI configuration files generated" -ForegroundColor Green

# Summary
Write-Host "`n🎉 CI/CD setup completed successfully!" -ForegroundColor Green
Write-Host @"

Configuration Summary:
  Environment: $Environment
  Region: $Region
  Base URL: $($env:BASE_URL)
  API URL: $($env:API_BASE_URL)
  Timezone: $($env:TEST_TIMEZONE)
  Locale: $($env:TEST_LOCALE)

Next Steps:
  1. Run tests: npm run test:ci
  2. View reports: npm run test:show-report
  3. Check logs: npm run test:debug

Files Created:
  - .env.ci (Environment configuration)
  - ../../.github/workflows/e2e-tests.yml (GitHub Actions workflow)

"@ -ForegroundColor Cyan

Write-Host "Happy testing! 🧪✨" -ForegroundColor Green