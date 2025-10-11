# E2E Test Runner Script for Taishang Laojun AI Platform
# PowerShell script to run various test suites

param(
    [string]$TestSuite = "all",
    [string]$Browser = "chromium",
    [string]$Environment = "local",
    [string]$Region = "us-east-1",
    [switch]$Headed = $false,
    [switch]$Debug = $false,
    [switch]$UpdateSnapshots = $false,
    [int]$Workers = 4,
    [int]$Retries = 2,
    [string]$Reporter = "html",
    [string]$OutputDir = "test-results",
    [switch]$CI = $false,
    [switch]$Coverage = $false,
    [string]$Grep = "",
    [string]$Project = "",
    [switch]$Help = $false
)

# Display help information
if ($Help) {
    Write-Host "E2E Test Runner for Taishang Laojun AI Platform" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage: .\run-tests.ps1 [OPTIONS]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Yellow
    Write-Host "  -TestSuite <suite>     Test suite to run (all, localization, compliance, performance, security, mobile, cross-browser, global-integration)" -ForegroundColor White
    Write-Host "  -Browser <browser>     Browser to use (chromium, firefox, webkit, chrome, edge)" -ForegroundColor White
    Write-Host "  -Environment <env>     Environment to test (local, staging, production)" -ForegroundColor White
    Write-Host "  -Region <region>       Region to test (us-east-1, eu-central-1, ap-east-1)" -ForegroundColor White
    Write-Host "  -Headed               Run tests in headed mode (visible browser)" -ForegroundColor White
    Write-Host "  -Debug                Enable debug mode" -ForegroundColor White
    Write-Host "  -UpdateSnapshots      Update visual snapshots" -ForegroundColor White
    Write-Host "  -Workers <number>     Number of parallel workers (default: 4)" -ForegroundColor White
    Write-Host "  -Retries <number>     Number of retries for failed tests (default: 2)" -ForegroundColor White
    Write-Host "  -Reporter <reporter>  Test reporter (html, json, junit, allure)" -ForegroundColor White
    Write-Host "  -OutputDir <dir>      Output directory for test results" -ForegroundColor White
    Write-Host "  -CI                   Run in CI mode" -ForegroundColor White
    Write-Host "  -Coverage             Enable code coverage" -ForegroundColor White
    Write-Host "  -Grep <pattern>       Run tests matching pattern" -ForegroundColor White
    Write-Host "  -Project <project>    Run specific project configuration" -ForegroundColor White
    Write-Host "  -Help                 Show this help message" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\run-tests.ps1 -TestSuite localization -Browser firefox" -ForegroundColor Gray
    Write-Host "  .\run-tests.ps1 -TestSuite performance -Environment staging -Headed" -ForegroundColor Gray
    Write-Host "  .\run-tests.ps1 -TestSuite all -CI -Reporter allure" -ForegroundColor Gray
    exit 0
}

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
$Green = "Green"
$Red = "Red"
$Yellow = "Yellow"
$Blue = "Blue"
$Cyan = "Cyan"

Write-Host "🚀 Starting E2E Tests for Taishang Laojun AI Platform" -ForegroundColor $Green
Write-Host "=================================================" -ForegroundColor $Green

# Validate parameters
$ValidTestSuites = @("all", "localization", "compliance", "performance", "security", "mobile", "cross-browser", "global-integration")
$ValidBrowsers = @("chromium", "firefox", "webkit", "chrome", "edge")
$ValidEnvironments = @("local", "staging", "production")
$ValidRegions = @("us-east-1", "eu-central-1", "ap-east-1")
$ValidReporters = @("html", "json", "junit", "allure")

if ($TestSuite -notin $ValidTestSuites) {
    Write-Host "❌ Invalid test suite: $TestSuite" -ForegroundColor $Red
    Write-Host "Valid options: $($ValidTestSuites -join ', ')" -ForegroundColor $Yellow
    exit 1
}

if ($Browser -notin $ValidBrowsers) {
    Write-Host "❌ Invalid browser: $Browser" -ForegroundColor $Red
    Write-Host "Valid options: $($ValidBrowsers -join ', ')" -ForegroundColor $Yellow
    exit 1
}

if ($Environment -notin $ValidEnvironments) {
    Write-Host "❌ Invalid environment: $Environment" -ForegroundColor $Red
    Write-Host "Valid options: $($ValidEnvironments -join ', ')" -ForegroundColor $Yellow
    exit 1
}

if ($Region -notin $ValidRegions) {
    Write-Host "❌ Invalid region: $Region" -ForegroundColor $Red
    Write-Host "Valid options: $($ValidRegions -join ', ')" -ForegroundColor $Yellow
    exit 1
}

if ($Reporter -notin $ValidReporters) {
    Write-Host "❌ Invalid reporter: $Reporter" -ForegroundColor $Red
    Write-Host "Valid options: $($ValidReporters -join ', ')" -ForegroundColor $Yellow
    exit 1
}

# Set environment variables
$env:NODE_ENV = "test"
$env:TEST_ENVIRONMENT = $Environment
$env:TEST_REGION = $Region
$env:TEST_BROWSER = $Browser

if ($CI) {
    $env:CI = "true"
    $env:GITHUB_ACTIONS = "true"
}

if ($Debug) {
    $env:DEBUG = "true"
    $env:PLAYWRIGHT_DEBUG = "1"
}

# Display configuration
Write-Host "📋 Test Configuration:" -ForegroundColor $Blue
Write-Host "  Test Suite: $TestSuite" -ForegroundColor $Cyan
Write-Host "  Browser: $Browser" -ForegroundColor $Cyan
Write-Host "  Environment: $Environment" -ForegroundColor $Cyan
Write-Host "  Region: $Region" -ForegroundColor $Cyan
Write-Host "  Workers: $Workers" -ForegroundColor $Cyan
Write-Host "  Retries: $Retries" -ForegroundColor $Cyan
Write-Host "  Reporter: $Reporter" -ForegroundColor $Cyan
Write-Host "  Output Directory: $OutputDir" -ForegroundColor $Cyan
Write-Host "  Headed Mode: $Headed" -ForegroundColor $Cyan
Write-Host "  Debug Mode: $Debug" -ForegroundColor $Cyan
Write-Host ""

# Check prerequisites
Write-Host "🔍 Checking Prerequisites..." -ForegroundColor $Blue

# Check Node.js
try {
    $nodeVersion = node --version
    Write-Host "  ✅ Node.js: $nodeVersion" -ForegroundColor $Green
} catch {
    Write-Host "  ❌ Node.js not found. Please install Node.js 18 or later." -ForegroundColor $Red
    exit 1
}

# Check npm
try {
    $npmVersion = npm --version
    Write-Host "  ✅ npm: $npmVersion" -ForegroundColor $Green
} catch {
    Write-Host "  ❌ npm not found." -ForegroundColor $Red
    exit 1
}

# Check if we're in the correct directory
if (-not (Test-Path "package.json")) {
    Write-Host "  ❌ package.json not found. Please run from the e2e test directory." -ForegroundColor $Red
    exit 1
}

Write-Host "  ✅ Prerequisites check passed" -ForegroundColor $Green
Write-Host ""

# Install dependencies if needed
if (-not (Test-Path "node_modules")) {
    Write-Host "📦 Installing Dependencies..." -ForegroundColor $Blue
    npm install
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to install dependencies" -ForegroundColor $Red
        exit 1
    }
    Write-Host "  ✅ Dependencies installed" -ForegroundColor $Green
    Write-Host ""
}

# Install Playwright browsers if needed
Write-Host "🌐 Checking Playwright Browsers..." -ForegroundColor $Blue
npx playwright install $Browser
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to install Playwright browsers" -ForegroundColor $Red
    exit 1
}
Write-Host "  ✅ Playwright browsers ready" -ForegroundColor $Green
Write-Host ""

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    Write-Host "📁 Created output directory: $OutputDir" -ForegroundColor $Blue
}

# Set up environment file
$envFile = ".env.test"
if (-not (Test-Path $envFile)) {
    Copy-Item ".env.example" $envFile
    Write-Host "📄 Created test environment file: $envFile" -ForegroundColor $Blue
}

# Build Playwright command
$playwrightArgs = @()

# Test files based on suite
switch ($TestSuite) {
    "localization" { $playwrightArgs += "tests/localization.spec.ts" }
    "compliance" { $playwrightArgs += "tests/compliance.spec.ts" }
    "performance" { $playwrightArgs += "tests/performance.spec.ts" }
    "security" { $playwrightArgs += "tests/security.spec.ts" }
    "mobile" { $playwrightArgs += "tests/mobile-responsive.spec.ts" }
    "cross-browser" { $playwrightArgs += "tests/cross-browser.spec.ts" }
    "global-integration" { $playwrightArgs += "tests/global-integration.spec.ts" }
    "all" { $playwrightArgs += "tests/" }
}

# Add browser project
if ($Project) {
    $playwrightArgs += "--project=$Project"
} else {
    $playwrightArgs += "--project=$Browser"
}

# Add workers
$playwrightArgs += "--workers=$Workers"

# Add retries
$playwrightArgs += "--retries=$Retries"

# Add reporter
switch ($Reporter) {
    "html" { $playwrightArgs += "--reporter=html" }
    "json" { $playwrightArgs += "--reporter=json" }
    "junit" { $playwrightArgs += "--reporter=junit" }
    "allure" { $playwrightArgs += "--reporter=allure-playwright" }
}

# Add output directory
$playwrightArgs += "--output-dir=$OutputDir"

# Add headed mode
if ($Headed) {
    $playwrightArgs += "--headed"
}

# Add debug mode
if ($Debug) {
    $playwrightArgs += "--debug"
}

# Add update snapshots
if ($UpdateSnapshots) {
    $playwrightArgs += "--update-snapshots"
}

# Add grep pattern
if ($Grep) {
    $playwrightArgs += "--grep=$Grep"
}

# Add CI mode
if ($CI) {
    $playwrightArgs += "--reporter=github"
}

# Start services if running locally
if ($Environment -eq "local") {
    Write-Host "🔧 Starting Local Services..." -ForegroundColor $Blue
    
    # Check if services are already running
    $frontendRunning = $false
    $backendRunning = $false
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:5173" -Method Head -TimeoutSec 5 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            $frontendRunning = $true
        }
    } catch {
        # Service not running
    }
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -Method Head -TimeoutSec 5 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            $backendRunning = $true
        }
    } catch {
        # Service not running
    }
    
    if (-not $frontendRunning) {
        Write-Host "  🚀 Starting frontend service..." -ForegroundColor $Yellow
        Start-Process -FilePath "npm" -ArgumentList "run", "dev" -WorkingDirectory "../../frontend/web-app" -WindowStyle Hidden
        Start-Sleep -Seconds 10
    } else {
        Write-Host "  ✅ Frontend service already running" -ForegroundColor $Green
    }
    
    if (-not $backendRunning) {
        Write-Host "  🚀 Starting backend services..." -ForegroundColor $Yellow
        # Start backend services here
        Start-Sleep -Seconds 15
    } else {
        Write-Host "  ✅ Backend services already running" -ForegroundColor $Green
    }
    
    Write-Host "  ✅ Local services ready" -ForegroundColor $Green
    Write-Host ""
}

# Run tests
Write-Host "🧪 Running E2E Tests..." -ForegroundColor $Blue
Write-Host "Command: npx playwright test $($playwrightArgs -join ' ')" -ForegroundColor $Gray
Write-Host ""

$testStartTime = Get-Date

try {
    & npx playwright test @playwrightArgs
    $testExitCode = $LASTEXITCODE
} catch {
    Write-Host "❌ Test execution failed: $($_.Exception.Message)" -ForegroundColor $Red
    exit 1
}

$testEndTime = Get-Date
$testDuration = $testEndTime - $testStartTime

Write-Host ""
Write-Host "📊 Test Results Summary:" -ForegroundColor $Blue
Write-Host "  Duration: $($testDuration.ToString('mm\:ss'))" -ForegroundColor $Cyan
Write-Host "  Exit Code: $testExitCode" -ForegroundColor $Cyan

if ($testExitCode -eq 0) {
    Write-Host "  Status: ✅ PASSED" -ForegroundColor $Green
} else {
    Write-Host "  Status: ❌ FAILED" -ForegroundColor $Red
}

# Generate reports
Write-Host ""
Write-Host "📈 Generating Reports..." -ForegroundColor $Blue

if ($Reporter -eq "html") {
    Write-Host "  📄 HTML Report: $OutputDir/playwright-report/index.html" -ForegroundColor $Cyan
}

if ($Reporter -eq "allure") {
    Write-Host "  📊 Generating Allure Report..." -ForegroundColor $Yellow
    try {
        & npx allure generate allure-results --clean -o allure-report
        Write-Host "  📄 Allure Report: allure-report/index.html" -ForegroundColor $Cyan
    } catch {
        Write-Host "  ⚠️  Failed to generate Allure report" -ForegroundColor $Yellow
    }
}

# Coverage report
if ($Coverage) {
    Write-Host "  📊 Generating Coverage Report..." -ForegroundColor $Yellow
    try {
        & npx nyc report --reporter=html --report-dir=coverage
        Write-Host "  📄 Coverage Report: coverage/index.html" -ForegroundColor $Cyan
    } catch {
        Write-Host "  ⚠️  Failed to generate coverage report" -ForegroundColor $Yellow
    }
}

# Archive results for CI
if ($CI) {
    Write-Host "  📦 Archiving test results..." -ForegroundColor $Yellow
    $archiveName = "test-results-$(Get-Date -Format 'yyyyMMdd-HHmmss').zip"
    Compress-Archive -Path $OutputDir -DestinationPath $archiveName -Force
    Write-Host "  📄 Test Archive: $archiveName" -ForegroundColor $Cyan
}

Write-Host ""

# Final status
if ($testExitCode -eq 0) {
    Write-Host "🎉 All tests completed successfully!" -ForegroundColor $Green
    Write-Host "=================================================" -ForegroundColor $Green
} else {
    Write-Host "💥 Some tests failed. Check the reports for details." -ForegroundColor $Red
    Write-Host "=================================================" -ForegroundColor $Red
}

# Open reports if not in CI mode
if (-not $CI -and $testExitCode -eq 0) {
    $openReport = Read-Host "Would you like to open the test report? (y/N)"
    if ($openReport -eq "y" -or $openReport -eq "Y") {
        if ($Reporter -eq "html" -and (Test-Path "$OutputDir/playwright-report/index.html")) {
            Start-Process "$OutputDir/playwright-report/index.html"
        } elseif ($Reporter -eq "allure" -and (Test-Path "allure-report/index.html")) {
            Start-Process "allure-report/index.html"
        }
    }
}

exit $testExitCode