# 太上老君监控系统测试脚本 (PowerShell)
# 支持单元测试、集成测试、端到端测试和性能测试

param(
    [Parameter(Position = 0)]
    [ValidateSet("unit", "integration", "e2e", "performance", "coverage", "lint", "security", "all", "clean", "help")]
    [string]$Command = "help",
    
    [Parameter()]
    [Alias("e")]
    [ValidateSet("local", "docker", "k8s")]
    [string]$Environment = "local",
    
    [Parameter()]
    [Alias("p")]
    [string]$Package,
    
    [Parameter()]
    [Alias("t")]
    [string]$Test,
    
    [Parameter()]
    [Alias("c")]
    [switch]$Coverage,
    
    [Parameter()]
    [Alias("v")]
    [switch]$Verbose,
    
    [Parameter()]
    [Alias("r")]
    [switch]$Race,
    
    [Parameter()]
    [Alias("b")]
    [switch]$Bench,
    
    [Parameter()]
    [Alias("s")]
    [switch]$Short,
    
    [Parameter()]
    [Alias("f")]
    [switch]$FailFast,
    
    [Parameter()]
    [string]$Timeout = "10m",
    
    [Parameter()]
    [int]$Parallel,
    
    [Parameter()]
    [string]$OutputDir = ".\test-results",
    
    [Parameter()]
    [string]$DockerImage,
    
    [Parameter()]
    [string]$K8sNamespace,
    
    [Parameter()]
    [int]$CoverageThreshold = 80,
    
    [Parameter()]
    [Alias("h")]
    [switch]$Help
)

# 脚本配置
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ServiceName = "monitoring"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $ColorMap = @{
        "Red" = [ConsoleColor]::Red
        "Green" = [ConsoleColor]::Green
        "Yellow" = [ConsoleColor]::Yellow
        "Blue" = [ConsoleColor]::Blue
        "White" = [ConsoleColor]::White
        "Cyan" = [ConsoleColor]::Cyan
        "Magenta" = [ConsoleColor]::Magenta
    }
    
    Write-Host $Message -ForegroundColor $ColorMap[$Color]
}

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 显示帮助信息
function Show-Help {
    @"
太上老君监控系统测试脚本 (PowerShell)

用法: .\test.ps1 [命令] [选项]

命令:
    unit        运行单元测试
    integration 运行集成测试
    e2e         运行端到端测试
    performance 运行性能测试
    coverage    生成测试覆盖率报告
    lint        运行代码检查
    security    运行安全扫描
    all         运行所有测试
    clean       清理测试环境
    help        显示此帮助信息

选项:
    -Environment, -e ENV    测试环境 (local|docker|k8s) [默认: local]
    -Package, -p PKG       指定测试包路径
    -Test, -t TEST         指定测试函数名
    -Coverage, -c          生成覆盖率报告
    -Verbose, -v           详细输出
    -Race, -r              启用竞态检测
    -Bench, -b             运行基准测试
    -Short, -s             运行短测试
    -FailFast, -f          遇到失败立即停止
    -Timeout TIMEOUT       测试超时时间 (默认: 10m)
    -Parallel N            并行测试数量
    -OutputDir DIR         输出目录 (默认: .\test-results)
    -DockerImage IMAGE     Docker 镜像名称
    -K8sNamespace NS       Kubernetes 命名空间
    -CoverageThreshold N   覆盖率阈值 (默认: 80)
    -Help, -h              显示帮助信息

环境变量:
    TEST_DATABASE_URL      测试数据库连接字符串
    TEST_REDIS_URL         测试 Redis 连接字符串
    TEST_LOG_LEVEL         测试日志级别
    COVERAGE_THRESHOLD     覆盖率阈值

示例:
    .\test.ps1 unit -Verbose
    .\test.ps1 integration -Environment docker
    .\test.ps1 e2e -Environment k8s -K8sNamespace test
    .\test.ps1 performance -Bench
    .\test.ps1 coverage -OutputDir .\coverage
    .\test.ps1 all -Coverage -Verbose

"@
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    # 检查 Go
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "Go 未安装或不在 PATH 中"
        exit 1
    }
    
    # 检查 Go 版本
    $goVersion = go version
    Write-Info "Go 版本: $goVersion"
    
    Write-Success "依赖检查完成"
}

# 检查 Docker 依赖
function Test-DockerDependencies {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker 未安装或不在 PATH 中"
        exit 1
    }
    
    if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
        Write-Error "Docker Compose 未安装或不在 PATH 中"
        exit 1
    }
}

# 检查 Kubernetes 依赖
function Test-K8sDependencies {
    if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
        Write-Error "kubectl 未安装或不在 PATH 中"
        exit 1
    }
    
    # 检查 kubectl 连接
    try {
        kubectl cluster-info | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "kubectl cluster-info failed"
        }
    }
    catch {
        Write-Error "无法连接到 Kubernetes 集群"
        exit 1
    }
}

# 设置测试环境
function Set-TestEnvironment {
    param(
        [string]$Environment,
        [string]$OutputDir
    )
    
    Write-Info "设置测试环境: $Environment"
    
    # 创建输出目录
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }
    
    # 设置环境变量
    $env:GO_ENV = "test"
    $env:LOG_LEVEL = "debug"
    
    switch ($Environment) {
        "local" { Set-LocalEnvironment }
        "docker" { Set-DockerEnvironment }
        "k8s" { Set-K8sEnvironment }
        default {
            Write-Error "未知测试环境: $Environment"
            exit 1
        }
    }
}

# 设置本地测试环境
function Set-LocalEnvironment {
    Write-Info "设置本地测试环境..."
    
    # 设置测试数据库
    if (-not $env:TEST_DATABASE_URL) {
        $env:TEST_DATABASE_URL = "postgres://test:test@localhost:5432/monitoring_test?sslmode=disable"
    }
    
    if (-not $env:TEST_REDIS_URL) {
        $env:TEST_REDIS_URL = "redis://localhost:6379/1"
    }
    
    # 检查测试数据库连接
    if (Get-Command psql -ErrorAction SilentlyContinue) {
        try {
            psql $env:TEST_DATABASE_URL -c "SELECT 1;" | Out-Null
        }
        catch {
            Write-Warning "无法连接到测试数据库，某些测试可能会失败"
        }
    }
    
    # 检查测试 Redis 连接
    if (Get-Command redis-cli -ErrorAction SilentlyContinue) {
        $redisUrl = [System.Uri]$env:TEST_REDIS_URL
        try {
            redis-cli -h $redisUrl.Host -p $redisUrl.Port ping | Out-Null
        }
        catch {
            Write-Warning "无法连接到测试 Redis，某些测试可能会失败"
        }
    }
}

# 设置 Docker 测试环境
function Set-DockerEnvironment {
    Write-Info "设置 Docker 测试环境..."
    
    Test-DockerDependencies
    
    # 启动测试服务
    Push-Location $ProjectRoot
    try {
        if (Test-Path "docker-compose.test.yml") {
            Write-Info "启动测试服务..."
            docker-compose -f docker-compose.test.yml up -d
            
            # 等待服务启动
            Write-Info "等待服务启动..."
            Start-Sleep -Seconds 10
            
            # 设置环境变量
            $env:TEST_DATABASE_URL = "postgres://test:test@localhost:5433/monitoring_test?sslmode=disable"
            $env:TEST_REDIS_URL = "redis://localhost:6380/1"
        }
        else {
            Write-Warning "docker-compose.test.yml 不存在，使用默认配置"
            Set-LocalEnvironment
        }
    }
    finally {
        Pop-Location
    }
}

# 设置 Kubernetes 测试环境
function Set-K8sEnvironment {
    Write-Info "设置 Kubernetes 测试环境..."
    
    Test-K8sDependencies
    
    # 这里可以添加 Kubernetes 测试环境的设置逻辑
    # 例如：部署测试数据库、Redis 等
    
    Write-Warning "Kubernetes 测试环境设置尚未实现"
}

# 清理测试环境
function Clear-TestEnvironment {
    param([string]$Environment)
    
    Write-Info "清理测试环境: $Environment"
    
    switch ($Environment) {
        "docker" {
            Push-Location $ProjectRoot
            try {
                if (Test-Path "docker-compose.test.yml") {
                    docker-compose -f docker-compose.test.yml down -v
                }
            }
            finally {
                Pop-Location
            }
        }
        "k8s" {
            # 清理 Kubernetes 测试资源
            Write-Info "清理 Kubernetes 测试资源..."
        }
    }
}

# 运行单元测试
function Invoke-UnitTests {
    param(
        [string]$Package,
        [string]$TestName,
        [bool]$Verbose,
        [bool]$Race,
        [bool]$Coverage,
        [string]$Timeout,
        [int]$Parallel,
        [string]$OutputDir,
        [bool]$Short,
        [bool]$FailFast
    )
    
    Write-Info "运行单元测试..."
    
    Push-Location $ProjectRoot
    try {
        # 构建测试命令参数
        $testArgs = @("test")
        
        if ($Package) {
            $testArgs += $Package
        }
        else {
            $testArgs += "./..."
        }
        
        if ($TestName) {
            $testArgs += @("-run", $TestName)
        }
        
        if ($Verbose) {
            $testArgs += "-v"
        }
        
        if ($Race) {
            $testArgs += "-race"
        }
        
        if ($Coverage) {
            $testArgs += @("-coverprofile=$OutputDir\coverage.out", "-covermode=atomic")
        }
        
        if ($Short) {
            $testArgs += "-short"
        }
        
        if ($FailFast) {
            $testArgs += "-failfast"
        }
        
        $testArgs += @("-timeout", $Timeout)
        
        if ($Parallel -gt 0) {
            $testArgs += @("-parallel", $Parallel.ToString())
        }
        
        # 执行测试
        Write-Info "执行命令: go $($testArgs -join ' ')"
        
        & go @testArgs
        if ($LASTEXITCODE -eq 0) {
            Write-Success "单元测试通过"
        }
        else {
            Write-Error "单元测试失败"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 运行集成测试
function Invoke-IntegrationTests {
    param(
        [string]$Package,
        [string]$TestName,
        [bool]$Verbose,
        [string]$Timeout,
        [string]$OutputDir,
        [bool]$FailFast
    )
    
    Write-Info "运行集成测试..."
    
    Push-Location $ProjectRoot
    try {
        # 构建测试命令参数
        $testArgs = @("test")
        
        if ($Package) {
            $testArgs += $Package
        }
        else {
            $testArgs += "./..."
        }
        
        # 集成测试标签
        $testArgs += @("-tags=integration")
        
        if ($TestName) {
            $testArgs += @("-run", $TestName)
        }
        
        if ($Verbose) {
            $testArgs += "-v"
        }
        
        if ($FailFast) {
            $testArgs += "-failfast"
        }
        
        $testArgs += @("-timeout", $Timeout)
        
        # 执行测试
        Write-Info "执行命令: go $($testArgs -join ' ')"
        
        & go @testArgs
        if ($LASTEXITCODE -eq 0) {
            Write-Success "集成测试通过"
        }
        else {
            Write-Error "集成测试失败"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 运行端到端测试
function Invoke-E2ETests {
    param(
        [string]$Environment,
        [string]$K8sNamespace,
        [bool]$Verbose,
        [string]$Timeout,
        [string]$OutputDir
    )
    
    Write-Info "运行端到端测试..."
    
    Push-Location $ProjectRoot
    try {
        # 设置环境变量
        if ($Environment -eq "k8s" -and $K8sNamespace) {
            $env:E2E_NAMESPACE = $K8sNamespace
        }
        
        # 构建测试命令参数
        $testArgs = @("test", "./tests/e2e/...")
        $testArgs += @("-tags=e2e")
        
        if ($Verbose) {
            $testArgs += "-v"
        }
        
        $testArgs += @("-timeout", $Timeout)
        
        # 执行测试
        Write-Info "执行命令: go $($testArgs -join ' ')"
        
        & go @testArgs
        if ($LASTEXITCODE -eq 0) {
            Write-Success "端到端测试通过"
        }
        else {
            Write-Error "端到端测试失败"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 运行性能测试
function Invoke-PerformanceTests {
    param(
        [string]$Package,
        [bool]$Bench,
        [bool]$Verbose,
        [string]$Timeout,
        [string]$OutputDir
    )
    
    Write-Info "运行性能测试..."
    
    Push-Location $ProjectRoot
    try {
        # 构建测试命令参数
        $testArgs = @("test")
        
        if ($Package) {
            $testArgs += $Package
        }
        else {
            $testArgs += "./..."
        }
        
        if ($Bench) {
            $testArgs += @("-bench=.", "-benchmem")
        }
        
        if ($Verbose) {
            $testArgs += "-v"
        }
        
        $testArgs += @("-timeout", $Timeout)
        
        # 输出到文件
        $benchOutput = Join-Path $OutputDir "benchmark.txt"
        
        # 执行测试
        Write-Info "执行命令: go $($testArgs -join ' ')"
        
        & go @testArgs | Tee-Object -FilePath $benchOutput
        if ($LASTEXITCODE -eq 0) {
            Write-Success "性能测试完成，结果保存到: $benchOutput"
        }
        else {
            Write-Error "性能测试失败"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 生成覆盖率报告
function New-CoverageReport {
    param(
        [string]$OutputDir,
        [int]$Threshold = 80
    )
    
    Write-Info "生成覆盖率报告..."
    
    Push-Location $ProjectRoot
    try {
        $coverageFile = Join-Path $OutputDir "coverage.out"
        $coverageHtml = Join-Path $OutputDir "coverage.html"
        
        if (-not (Test-Path $coverageFile)) {
            Write-Error "覆盖率文件不存在: $coverageFile"
            return $false
        }
        
        # 生成 HTML 报告
        go tool cover -html=$coverageFile -o $coverageHtml
        Write-Success "HTML 覆盖率报告生成: $coverageHtml"
        
        # 显示覆盖率统计
        $coverageOutput = go tool cover -func=$coverageFile
        $coveragePercent = ($coverageOutput | Select-Object -Last 1) -replace '.*\s(\d+\.\d+)%.*', '$1'
        
        Write-Info "总覆盖率: ${coveragePercent}%"
        
        # 检查覆盖率阈值
        if ([double]$coveragePercent -ge $Threshold) {
            Write-Success "覆盖率达到阈值 (${Threshold}%)"
        }
        else {
            Write-Warning "覆盖率未达到阈值 (${Threshold}%)，当前: ${coveragePercent}%"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 运行代码检查
function Invoke-Lint {
    Write-Info "运行代码检查..."
    
    Push-Location $ProjectRoot
    try {
        # 检查 golangci-lint
        if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
            Write-Warning "golangci-lint 未安装，跳过代码检查"
            return $true
        }
        
        # 运行 golangci-lint
        golangci-lint run ./...
        if ($LASTEXITCODE -eq 0) {
            Write-Success "代码检查通过"
        }
        else {
            Write-Error "代码检查失败"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 运行安全扫描
function Invoke-SecurityScan {
    Write-Info "运行安全扫描..."
    
    Push-Location $ProjectRoot
    try {
        # 检查 gosec
        if (-not (Get-Command gosec -ErrorAction SilentlyContinue)) {
            Write-Warning "gosec 未安装，跳过安全扫描"
            return $true
        }
        
        # 运行 gosec
        gosec ./...
        if ($LASTEXITCODE -eq 0) {
            Write-Success "安全扫描通过"
        }
        else {
            Write-Error "安全扫描发现问题"
            return $false
        }
    }
    finally {
        Pop-Location
    }
    
    return $true
}

# 主函数
function Main {
    # 如果指定了帮助参数，显示帮助
    if ($Help -or $Command -eq "help") {
        Show-Help
        return
    }
    
    # 启用详细输出
    if ($Verbose) {
        $VerbosePreference = "Continue"
    }
    
    # 检查依赖
    Test-Dependencies
    
    # 设置测试环境
    Set-TestEnvironment $Environment $OutputDir
    
    # 设置清理函数
    try {
        # 执行命令
        $success = $true
        
        switch ($Command) {
            "unit" {
                $success = Invoke-UnitTests $Package $Test $Verbose.IsPresent $Race.IsPresent $Coverage.IsPresent $Timeout $Parallel $OutputDir $Short.IsPresent $FailFast.IsPresent
            }
            "integration" {
                $success = Invoke-IntegrationTests $Package $Test $Verbose.IsPresent $Timeout $OutputDir $FailFast.IsPresent
            }
            "e2e" {
                $success = Invoke-E2ETests $Environment $K8sNamespace $Verbose.IsPresent $Timeout $OutputDir
            }
            "performance" {
                $success = Invoke-PerformanceTests $Package $Bench.IsPresent $Verbose.IsPresent $Timeout $OutputDir
            }
            "coverage" {
                # 先运行单元测试生成覆盖率
                $success = Invoke-UnitTests $Package $Test $Verbose.IsPresent $Race.IsPresent $true $Timeout $Parallel $OutputDir $Short.IsPresent $FailFast.IsPresent
                if ($success) {
                    $success = New-CoverageReport $OutputDir $CoverageThreshold
                }
            }
            "lint" {
                $success = Invoke-Lint
            }
            "security" {
                $success = Invoke-SecurityScan
            }
            "all" {
                Write-Info "运行所有测试..."
                
                # 运行代码检查
                $success = Invoke-Lint
                if (-not $success) { throw "代码检查失败" }
                
                # 运行安全扫描
                $success = Invoke-SecurityScan
                if (-not $success) { throw "安全扫描失败" }
                
                # 运行单元测试
                $success = Invoke-UnitTests $Package $Test $Verbose.IsPresent $Race.IsPresent $true $Timeout $Parallel $OutputDir $Short.IsPresent $FailFast.IsPresent
                if (-not $success) { throw "单元测试失败" }
                
                # 运行集成测试
                $success = Invoke-IntegrationTests $Package $Test $Verbose.IsPresent $Timeout $OutputDir $FailFast.IsPresent
                if (-not $success) { throw "集成测试失败" }
                
                # 生成覆盖率报告
                $success = New-CoverageReport $OutputDir $CoverageThreshold
                if (-not $success) { throw "覆盖率检查失败" }
                
                Write-Success "所有测试完成"
            }
            "clean" {
                Write-Info "清理测试环境..."
                Clear-TestEnvironment $Environment
                if (Test-Path $OutputDir) {
                    Remove-Item -Path $OutputDir -Recurse -Force
                }
                Write-Success "清理完成"
            }
            default {
                Write-Error "未知命令: $Command"
                Show-Help
                exit 1
            }
        }
        
        if (-not $success) {
            exit 1
        }
    }
    finally {
        # 清理测试环境
        Clear-TestEnvironment $Environment
    }
}

# 执行主函数
Main