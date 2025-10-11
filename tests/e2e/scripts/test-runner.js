#!/usr/bin/env node

/**
 * Advanced Test Runner for Taishang Laojun AI Platform E2E Tests
 * 
 * This script provides advanced test execution capabilities including:
 * - Parallel test execution across multiple browsers and regions
 * - Test result aggregation and reporting
 * - Performance monitoring and alerting
 * - Failure analysis and retry logic
 * - Integration with monitoring systems
 */

const { spawn, exec } = require('child_process');
const fs = require('fs').promises;
const path = require('path');
const os = require('os');

class TestRunner {
    constructor(options = {}) {
        this.options = {
            parallel: options.parallel || 4,
            retries: options.retries || 2,
            timeout: options.timeout || 300000, // 5 minutes
            browsers: options.browsers || ['chromium', 'firefox', 'webkit'],
            regions: options.regions || ['us-east-1', 'eu-central-1', 'ap-east-1'],
            testSuites: options.testSuites || ['all'],
            environment: options.environment || 'ci',
            reportFormats: options.reportFormats || ['html', 'json', 'junit'],
            verbose: options.verbose || false,
            ...options
        };
        
        this.results = {
            total: 0,
            passed: 0,
            failed: 0,
            skipped: 0,
            duration: 0,
            suites: [],
            errors: []
        };
        
        this.startTime = Date.now();
    }

    /**
     * Main test execution method
     */
    async run() {
        try {
            console.log('🚀 Starting Taishang Laojun AI Platform E2E Tests...');
            console.log(`📊 Configuration: ${JSON.stringify(this.options, null, 2)}`);
            
            await this.validateEnvironment();
            await this.setupTestEnvironment();
            
            const testMatrix = this.generateTestMatrix();
            console.log(`🧪 Generated ${testMatrix.length} test configurations`);
            
            const results = await this.executeTestMatrix(testMatrix);
            await this.aggregateResults(results);
            await this.generateReports();
            await this.analyzeResults();
            
            this.printSummary();
            
            // Exit with appropriate code
            process.exit(this.results.failed > 0 ? 1 : 0);
            
        } catch (error) {
            console.error('❌ Test runner failed:', error);
            process.exit(1);
        }
    }

    /**
     * Validate test environment
     */
    async validateEnvironment() {
        console.log('🔍 Validating test environment...');
        
        // Check Node.js version
        const nodeVersion = process.version;
        const majorVersion = parseInt(nodeVersion.slice(1).split('.')[0]);
        if (majorVersion < 16) {
            throw new Error(`Node.js 16+ required, found ${nodeVersion}`);
        }
        
        // Check if Playwright is installed
        try {
            await this.execCommand('npx playwright --version');
        } catch (error) {
            throw new Error('Playwright not found. Run: npm install && npx playwright install');
        }
        
        // Check required environment variables
        const requiredVars = ['BASE_URL', 'API_BASE_URL'];
        for (const varName of requiredVars) {
            if (!process.env[varName]) {
                throw new Error(`Required environment variable ${varName} not set`);
            }
        }
        
        console.log('✅ Environment validation passed');
    }

    /**
     * Setup test environment
     */
    async setupTestEnvironment() {
        console.log('🔧 Setting up test environment...');
        
        // Create test directories
        const dirs = [
            'test-results',
            'test-results/screenshots',
            'test-results/videos',
            'test-results/traces',
            'test-results/reports'
        ];
        
        for (const dir of dirs) {
            await fs.mkdir(dir, { recursive: true });
        }
        
        // Clean previous test results
        try {
            await fs.rm('test-results', { recursive: true, force: true });
            await fs.mkdir('test-results', { recursive: true });
        } catch (error) {
            // Ignore if directory doesn't exist
        }
        
        console.log('✅ Test environment setup completed');
    }

    /**
     * Generate test matrix based on configuration
     */
    generateTestMatrix() {
        const matrix = [];
        
        for (const browser of this.options.browsers) {
            for (const region of this.options.regions) {
                for (const suite of this.options.testSuites) {
                    matrix.push({
                        browser,
                        region,
                        suite,
                        id: `${browser}-${region}-${suite}`,
                        retries: 0
                    });
                }
            }
        }
        
        return matrix;
    }

    /**
     * Execute test matrix with parallel execution
     */
    async executeTestMatrix(testMatrix) {
        console.log(`🔄 Executing ${testMatrix.length} test configurations...`);
        
        const results = [];
        const chunks = this.chunkArray(testMatrix, this.options.parallel);
        
        for (const chunk of chunks) {
            const chunkPromises = chunk.map(config => this.executeTestConfig(config));
            const chunkResults = await Promise.allSettled(chunkPromises);
            results.push(...chunkResults);
        }
        
        return results;
    }

    /**
     * Execute single test configuration
     */
    async executeTestConfig(config) {
        const startTime = Date.now();
        console.log(`🧪 Running tests: ${config.id}`);
        
        try {
            // Set environment variables for this test run
            const env = {
                ...process.env,
                DEFAULT_REGION: config.region,
                TEST_BROWSER: config.browser,
                TEST_SUITE: config.suite,
                PWTEST_OUTPUT_DIR: `test-results/${config.id}`
            };
            
            // Build Playwright command
            const args = [
                'playwright', 'test',
                '--project', config.browser,
                '--output-dir', `test-results/${config.id}`,
                '--reporter', 'json'
            ];
            
            // Add test suite filter
            if (config.suite !== 'all') {
                args.push(`tests/${config.suite}.spec.ts`);
            }
            
            // Execute test
            const result = await this.execCommand(`npx ${args.join(' ')}`, { env });
            
            const duration = Date.now() - startTime;
            
            return {
                config,
                status: 'fulfilled',
                result: {
                    success: true,
                    output: result.stdout,
                    error: result.stderr,
                    duration,
                    exitCode: 0
                }
            };
            
        } catch (error) {
            const duration = Date.now() - startTime;
            
            // Retry logic
            if (config.retries < this.options.retries) {
                console.log(`🔄 Retrying ${config.id} (attempt ${config.retries + 1}/${this.options.retries})`);
                config.retries++;
                return this.executeTestConfig(config);
            }
            
            return {
                config,
                status: 'rejected',
                result: {
                    success: false,
                    output: error.stdout || '',
                    error: error.stderr || error.message,
                    duration,
                    exitCode: error.code || 1
                }
            };
        }
    }

    /**
     * Aggregate test results from all configurations
     */
    async aggregateResults(results) {
        console.log('📊 Aggregating test results...');
        
        for (const result of results) {
            const { config, status, result: testResult } = result;
            
            if (status === 'fulfilled' && testResult.success) {
                // Parse Playwright JSON output
                try {
                    const outputDir = `test-results/${config.id}`;
                    const reportPath = path.join(outputDir, 'results.json');
                    
                    if (await this.fileExists(reportPath)) {
                        const reportData = JSON.parse(await fs.readFile(reportPath, 'utf8'));
                        
                        this.results.total += reportData.stats?.total || 0;
                        this.results.passed += reportData.stats?.passed || 0;
                        this.results.failed += reportData.stats?.failed || 0;
                        this.results.skipped += reportData.stats?.skipped || 0;
                        
                        this.results.suites.push({
                            config,
                            stats: reportData.stats,
                            duration: testResult.duration,
                            tests: reportData.tests || []
                        });
                    }
                } catch (error) {
                    console.warn(`⚠️ Failed to parse results for ${config.id}:`, error.message);
                }
            } else {
                this.results.failed++;
                this.results.errors.push({
                    config,
                    error: testResult.error,
                    duration: testResult.duration
                });
            }
        }
        
        this.results.duration = Date.now() - this.startTime;
    }

    /**
     * Generate test reports in multiple formats
     */
    async generateReports() {
        console.log('📄 Generating test reports...');
        
        const reportDir = 'test-results/reports';
        await fs.mkdir(reportDir, { recursive: true });
        
        // Generate HTML report
        if (this.options.reportFormats.includes('html')) {
            await this.generateHtmlReport(reportDir);
        }
        
        // Generate JSON report
        if (this.options.reportFormats.includes('json')) {
            await this.generateJsonReport(reportDir);
        }
        
        // Generate JUnit report
        if (this.options.reportFormats.includes('junit')) {
            await this.generateJunitReport(reportDir);
        }
        
        // Generate Allure report
        if (this.options.reportFormats.includes('allure')) {
            await this.generateAllureReport(reportDir);
        }
        
        console.log('✅ Test reports generated');
    }

    /**
     * Generate HTML report
     */
    async generateHtmlReport(reportDir) {
        const htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Taishang Laojun AI Platform - E2E Test Report</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 8px 8px 0 0; }
        .header h1 { margin: 0; font-size: 2.5em; }
        .header .subtitle { opacity: 0.9; margin-top: 10px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; padding: 30px; }
        .stat-card { background: #f8f9fa; padding: 20px; border-radius: 8px; text-align: center; border-left: 4px solid #667eea; }
        .stat-card.passed { border-left-color: #28a745; }
        .stat-card.failed { border-left-color: #dc3545; }
        .stat-card.skipped { border-left-color: #ffc107; }
        .stat-number { font-size: 2.5em; font-weight: bold; margin-bottom: 5px; }
        .stat-label { color: #6c757d; text-transform: uppercase; font-size: 0.9em; letter-spacing: 1px; }
        .suites { padding: 0 30px 30px; }
        .suite { margin-bottom: 20px; border: 1px solid #e9ecef; border-radius: 8px; overflow: hidden; }
        .suite-header { background: #f8f9fa; padding: 15px; font-weight: bold; cursor: pointer; }
        .suite-content { padding: 15px; display: none; }
        .suite-content.active { display: block; }
        .test-item { padding: 10px; border-bottom: 1px solid #e9ecef; display: flex; justify-content: space-between; align-items: center; }
        .test-item:last-child { border-bottom: none; }
        .test-status { padding: 4px 8px; border-radius: 4px; font-size: 0.8em; font-weight: bold; }
        .test-status.passed { background: #d4edda; color: #155724; }
        .test-status.failed { background: #f8d7da; color: #721c24; }
        .test-status.skipped { background: #fff3cd; color: #856404; }
        .errors { padding: 0 30px 30px; }
        .error-item { background: #f8d7da; border: 1px solid #f5c6cb; border-radius: 8px; padding: 15px; margin-bottom: 15px; }
        .error-title { font-weight: bold; color: #721c24; margin-bottom: 10px; }
        .error-message { font-family: monospace; background: #fff; padding: 10px; border-radius: 4px; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 E2E Test Report</h1>
            <div class="subtitle">Taishang Laojun AI Platform - ${new Date().toLocaleString()}</div>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">${this.results.total}</div>
                <div class="stat-label">Total Tests</div>
            </div>
            <div class="stat-card passed">
                <div class="stat-number">${this.results.passed}</div>
                <div class="stat-label">Passed</div>
            </div>
            <div class="stat-card failed">
                <div class="stat-number">${this.results.failed}</div>
                <div class="stat-label">Failed</div>
            </div>
            <div class="stat-card skipped">
                <div class="stat-number">${this.results.skipped}</div>
                <div class="stat-label">Skipped</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">${Math.round(this.results.duration / 1000)}s</div>
                <div class="stat-label">Duration</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">${Math.round((this.results.passed / this.results.total) * 100) || 0}%</div>
                <div class="stat-label">Success Rate</div>
            </div>
        </div>
        
        <div class="suites">
            <h2>Test Suites</h2>
            ${this.results.suites.map(suite => `
                <div class="suite">
                    <div class="suite-header" onclick="toggleSuite(this)">
                        ${suite.config.browser} - ${suite.config.region} - ${suite.config.suite}
                        <span style="float: right;">${suite.stats?.passed || 0}/${suite.stats?.total || 0} passed</span>
                    </div>
                    <div class="suite-content">
                        ${(suite.tests || []).map(test => `
                            <div class="test-item">
                                <span>${test.title || test.name || 'Unknown test'}</span>
                                <span class="test-status ${test.status || 'unknown'}">${test.status || 'unknown'}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `).join('')}
        </div>
        
        ${this.results.errors.length > 0 ? `
        <div class="errors">
            <h2>Errors</h2>
            ${this.results.errors.map(error => `
                <div class="error-item">
                    <div class="error-title">${error.config.browser} - ${error.config.region} - ${error.config.suite}</div>
                    <div class="error-message">${error.error}</div>
                </div>
            `).join('')}
        </div>
        ` : ''}
    </div>
    
    <script>
        function toggleSuite(header) {
            const content = header.nextElementSibling;
            content.classList.toggle('active');
        }
    </script>
</body>
</html>`;
        
        await fs.writeFile(path.join(reportDir, 'index.html'), htmlContent);
    }

    /**
     * Generate JSON report
     */
    async generateJsonReport(reportDir) {
        const jsonReport = {
            summary: {
                total: this.results.total,
                passed: this.results.passed,
                failed: this.results.failed,
                skipped: this.results.skipped,
                duration: this.results.duration,
                successRate: Math.round((this.results.passed / this.results.total) * 100) || 0,
                timestamp: new Date().toISOString()
            },
            suites: this.results.suites,
            errors: this.results.errors,
            configuration: this.options
        };
        
        await fs.writeFile(
            path.join(reportDir, 'results.json'),
            JSON.stringify(jsonReport, null, 2)
        );
    }

    /**
     * Generate JUnit XML report
     */
    async generateJunitReport(reportDir) {
        const xml = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="Taishang Laojun E2E Tests" tests="${this.results.total}" failures="${this.results.failed}" time="${this.results.duration / 1000}">
${this.results.suites.map(suite => `
    <testsuite name="${suite.config.browser}-${suite.config.region}-${suite.config.suite}" tests="${suite.stats?.total || 0}" failures="${suite.stats?.failed || 0}" time="${suite.duration / 1000}">
    ${(suite.tests || []).map(test => `
        <testcase name="${test.title || test.name || 'Unknown'}" time="${test.duration / 1000 || 0}">
        ${test.status === 'failed' ? `<failure message="${test.error || 'Test failed'}">${test.error || 'Test failed'}</failure>` : ''}
        </testcase>
    `).join('')}
    </testsuite>
`).join('')}
</testsuites>`;
        
        await fs.writeFile(path.join(reportDir, 'junit.xml'), xml);
    }

    /**
     * Generate Allure report
     */
    async generateAllureReport(reportDir) {
        try {
            await this.execCommand('npx allure generate allure-results --clean -o ' + path.join(reportDir, 'allure'));
        } catch (error) {
            console.warn('⚠️ Failed to generate Allure report:', error.message);
        }
    }

    /**
     * Analyze test results and provide insights
     */
    async analyzeResults() {
        console.log('🔍 Analyzing test results...');
        
        const analysis = {
            performance: this.analyzePerformance(),
            reliability: this.analyzeReliability(),
            coverage: this.analyzeCoverage(),
            recommendations: this.generateRecommendations()
        };
        
        // Save analysis
        await fs.writeFile(
            'test-results/reports/analysis.json',
            JSON.stringify(analysis, null, 2)
        );
        
        console.log('✅ Test analysis completed');
    }

    /**
     * Analyze performance metrics
     */
    analyzePerformance() {
        const avgDuration = this.results.suites.reduce((sum, suite) => sum + suite.duration, 0) / this.results.suites.length;
        const slowestSuite = this.results.suites.reduce((slowest, suite) => 
            suite.duration > slowest.duration ? suite : slowest, { duration: 0 });
        
        return {
            averageDuration: avgDuration,
            slowestSuite: slowestSuite.config,
            slowestDuration: slowestSuite.duration,
            performanceGrade: avgDuration < 60000 ? 'A' : avgDuration < 120000 ? 'B' : 'C'
        };
    }

    /**
     * Analyze test reliability
     */
    analyzeReliability() {
        const successRate = (this.results.passed / this.results.total) * 100;
        const flakyTests = this.results.suites.filter(suite => 
            suite.config.retries > 0).length;
        
        return {
            successRate,
            flakyTestCount: flakyTests,
            reliabilityGrade: successRate >= 95 ? 'A' : successRate >= 85 ? 'B' : 'C'
        };
    }

    /**
     * Analyze test coverage
     */
    analyzeCoverage() {
        const browserCoverage = [...new Set(this.results.suites.map(s => s.config.browser))];
        const regionCoverage = [...new Set(this.results.suites.map(s => s.config.region))];
        const suiteCoverage = [...new Set(this.results.suites.map(s => s.config.suite))];
        
        return {
            browsers: browserCoverage,
            regions: regionCoverage,
            suites: suiteCoverage,
            coverageScore: (browserCoverage.length + regionCoverage.length + suiteCoverage.length) / 9 * 100
        };
    }

    /**
     * Generate recommendations based on analysis
     */
    generateRecommendations() {
        const recommendations = [];
        
        if (this.results.failed > 0) {
            recommendations.push('Investigate and fix failing tests to improve reliability');
        }
        
        if (this.results.duration > 300000) { // 5 minutes
            recommendations.push('Consider optimizing test execution time or increasing parallelization');
        }
        
        if (this.results.errors.length > 0) {
            recommendations.push('Review error logs and improve test stability');
        }
        
        return recommendations;
    }

    /**
     * Print test summary
     */
    printSummary() {
        console.log('\n🎉 Test Execution Summary');
        console.log('========================');
        console.log(`Total Tests: ${this.results.total}`);
        console.log(`✅ Passed: ${this.results.passed}`);
        console.log(`❌ Failed: ${this.results.failed}`);
        console.log(`⏭️ Skipped: ${this.results.skipped}`);
        console.log(`⏱️ Duration: ${Math.round(this.results.duration / 1000)}s`);
        console.log(`📊 Success Rate: ${Math.round((this.results.passed / this.results.total) * 100) || 0}%`);
        
        if (this.results.errors.length > 0) {
            console.log('\n❌ Errors:');
            this.results.errors.forEach(error => {
                console.log(`  - ${error.config.id}: ${error.error.split('\n')[0]}`);
            });
        }
        
        console.log(`\n📄 Reports available in: test-results/reports/`);
    }

    /**
     * Utility methods
     */
    async execCommand(command, options = {}) {
        return new Promise((resolve, reject) => {
            exec(command, { timeout: this.options.timeout, ...options }, (error, stdout, stderr) => {
                if (error) {
                    error.stdout = stdout;
                    error.stderr = stderr;
                    reject(error);
                } else {
                    resolve({ stdout, stderr });
                }
            });
        });
    }

    async fileExists(filePath) {
        try {
            await fs.access(filePath);
            return true;
        } catch {
            return false;
        }
    }

    chunkArray(array, size) {
        const chunks = [];
        for (let i = 0; i < array.length; i += size) {
            chunks.push(array.slice(i, i + size));
        }
        return chunks;
    }
}

// CLI interface
if (require.main === module) {
    const args = process.argv.slice(2);
    const options = {};
    
    // Parse command line arguments
    for (let i = 0; i < args.length; i += 2) {
        const key = args[i].replace(/^--/, '');
        const value = args[i + 1];
        
        switch (key) {
            case 'browsers':
                options.browsers = value.split(',');
                break;
            case 'regions':
                options.regions = value.split(',');
                break;
            case 'suites':
                options.testSuites = value.split(',');
                break;
            case 'parallel':
                options.parallel = parseInt(value);
                break;
            case 'retries':
                options.retries = parseInt(value);
                break;
            case 'timeout':
                options.timeout = parseInt(value);
                break;
            case 'environment':
                options.environment = value;
                break;
            case 'reports':
                options.reportFormats = value.split(',');
                break;
            case 'verbose':
                options.verbose = true;
                i--; // No value for this flag
                break;
            case 'help':
                console.log(`
Taishang Laojun AI Platform E2E Test Runner

Usage: node test-runner.js [options]

Options:
  --browsers <list>     Comma-separated list of browsers (chromium,firefox,webkit)
  --regions <list>      Comma-separated list of regions (us-east-1,eu-central-1,ap-east-1)
  --suites <list>       Comma-separated list of test suites (all,localization,compliance,etc.)
  --parallel <number>   Number of parallel test executions (default: 4)
  --retries <number>    Number of retries for failed tests (default: 2)
  --timeout <number>    Test timeout in milliseconds (default: 300000)
  --environment <env>   Test environment (ci,staging,production)
  --reports <list>      Report formats (html,json,junit,allure)
  --verbose             Enable verbose output
  --help                Show this help message

Examples:
  node test-runner.js --browsers chromium,firefox --regions us-east-1 --suites localization
  node test-runner.js --parallel 2 --retries 1 --environment staging
  node test-runner.js --reports html,json,allure --verbose
                `);
                process.exit(0);
        }
    }
    
    const runner = new TestRunner(options);
    runner.run().catch(error => {
        console.error('Test runner failed:', error);
        process.exit(1);
    });
}

module.exports = TestRunner;