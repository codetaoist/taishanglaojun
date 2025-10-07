# 太上老君监控系统管理 PowerShell 脚本
# 用于管理监控系统的启动、停止、状态检查等操作

param(
    [Parameter(Position = 0)]
    [ValidateSet("start", "stop", "restart", "status", "logs", "health", "metrics", "config", "install", "uninstall", "update", "backup", "restore", "help")]
    [string]$Command = "status",
    
    [Alias("e")]
    [ValidateSet("development", "staging", "production")]
    [string]$Environment = "development",
    
    [Alias("p")]
    [int]$Port = 8080,
    
    [Alias("m")]
    [int]$MetricsPort = 9090,
    
    [Alias("l")]
    [ValidateSet("debug", "info", "warn", "error")]
    [string]$LogLevel = "info",
    
    [Alias("c")]
    [string]$ConfigFile = "",
    
    [Alias("d")]
    [switch]$Daemon,
    
    [Alias("f")]
    [switch]$Follow,
    
    [Alias("n")]
    [int]$Lines = 50,
    
    [Alias("t")]
    [int]$Timeout = 30,
    
    [Alias("v")]
    [switch]$Verbose,
    
    [Alias("q")]
    [switch]$Quiet,
    
    [switch]$DryRun,
    [switch]$Force,
    [switch]$NoColor,
    [switch]$Help
)

# 脚本配置
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ServiceName = "monitoring"

# 如果指定了帮助参数，显示帮助
if ($Help) {
    $Command = "help"
}

# 设置环境变量默认值
if ($env:MONITORING_ENV) {
    $Environment = $env:MONITORING_ENV
}
if ($env:MONITORING_PORT) {
    $Port = [int]$env:MONITORING_PORT
}
if ($env:MONITORING_METRICS_PORT) {
    $MetricsPort = [int]$env:MONITORING_METRICS_PORT
}
if ($env:MONITORING_LOG_LEVEL) {
    $LogLevel = $env:MONITORING_LOG_LEVEL
}
if ($env:MONITORING_CONFIG) {
    $ConfigFile = $env:MONITORING_CONFIG
}

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    if ($NoColor -or $Quiet) {
        if (-not $Quiet) {
            Write-Host $Message
        }
        return
    }
    
    $colorMap = @{
        "Red" = "Red"
        "Green" = "Green"
        "Yellow" = "Yellow"
        "Blue" = "Blue"
        "Cyan" = "Cyan"
        "Magenta" = "Magenta"
        "White" = "White"
    }
    
    Write-Host $Message -ForegroundColor $colorMap[$Color]
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

function Write-Debug {
    param([string]$Message)
    if ($Verbose) {
        Write-ColorOutput "[DEBUG] $Message" "Cyan"
    }
}

# 显示帮助信息
function Show-Help {
    @"
太上老君监控系统管理 PowerShell 脚本

用法: .\monitoring.ps1 [选项] [命令]

命令:
    start       启动监控服务
    stop        停止监控服务
    restart     重启监控服务
    status      查看服务状态
    logs        查看服务日志
    health      健康检查
    metrics     查看指标
    config      配置管理
    install     安装服务
    uninstall   卸载服务
    update      更新服务
    backup      备份数据
    restore     恢复数据
    help        显示此帮助信息

选项:
    -Environment, -e ENV    环境 (development|staging|production) [默认: development]
    -Port, -p PORT          HTTP 端口 [默认: 8080]
    -MetricsPort, -m PORT   指标端口 [默认: 9090]
    -LogLevel, -l LEVEL     日志级别 (debug|info|warn|error) [默认: info]
    -ConfigFile, -c FILE    配置文件路径
    -Daemon, -d             后台运行
    -Follow, -f             跟踪日志输出
    -Lines, -n NUM          日志行数 [默认: 50]
    -Timeout, -t SECONDS    超时时间 [默认: 30]
    -Verbose, -v            详细输出
    -Quiet, -q              静默模式
    -DryRun                 试运行模式
    -Force                  强制执行
    -NoColor                禁用颜色输出

环境变量:
    MONITORING_ENV          监控环境
    MONITORING_PORT         HTTP 端口
    MONITORING_METRICS_PORT 指标端口
    MONITORING_LOG_LEVEL    日志级别
    MONITORING_CONFIG       配置文件路径
    DATABASE_URL            数据库连接地址
    REDIS_URL               Redis 连接地址

示例:
    .\monitoring.ps1 start
    .\monitoring.ps1 start -Environment production -Port 8080
    .\monitoring.ps1 stop -Force
    .\monitoring.ps1 restart -Daemon
    .\monitoring.ps1 status
    .\monitoring.ps1 logs -Follow -Lines 100
    .\monitoring.ps1 health
    .\monitoring.ps1 metrics

"@
}

# 检查依赖
function Test-Dependencies {
    $deps = @("curl")
    
    foreach ($dep in $deps) {
        if (-not (Get-Command $dep -ErrorAction SilentlyContinue)) {
            Write-Warning "依赖未安装: $dep"
        }
    }
}

# 获取 PID 文件路径
function Get-PidFile {
    return "$env:TEMP\${ServiceName}.pid"
}

# 获取日志文件路径
function Get-LogFile {
    switch ($Environment) {
        "development" {
            return "$ProjectRoot\logs\app.log"
        }
        { $_ -in @("staging", "production") } {
            return "C:\Logs\${ServiceName}\app.log"
        }
    }
}

# 获取配置文件路径
function Get-ConfigFile {
    if ($ConfigFile) {
        return $ConfigFile
    } else {
        return "$ProjectRoot\config\${Environment}.yml"
    }
}

# 检查服务是否运行
function Test-ServiceRunning {
    $pidFile = Get-PidFile
    
    if (Test-Path $pidFile) {
        $pid = Get-Content $pidFile -ErrorAction SilentlyContinue
        if ($pid -and (Get-Process -Id $pid -ErrorAction SilentlyContinue)) {
            return $true
        } else {
            # PID 文件存在但进程不存在，清理 PID 文件
            Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
            return $false
        }
    } else {
        return $false
    }
}

# 获取服务 PID
function Get-ServicePid {
    $pidFile = Get-PidFile
    
    if (Test-Path $pidFile) {
        return Get-Content $pidFile -ErrorAction SilentlyContinue
    } else {
        return $null
    }
}

# 等待服务启动
function Wait-ForService {
    param([int]$TimeoutSeconds)
    
    Write-Info "等待服务启动..."
    
    $startTime = Get-Date
    
    while ($true) {
        if (Test-ServiceRunning) {
            # 检查 HTTP 端点
            try {
                $response = Invoke-WebRequest -Uri "http://localhost:${Port}/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction SilentlyContinue
                if ($response.StatusCode -eq 200) {
                    Write-Success "服务启动成功"
                    return $true
                }
            } catch {
                # 继续等待
            }
        }
        
        $elapsed = (Get-Date) - $startTime
        if ($elapsed.TotalSeconds -ge $TimeoutSeconds) {
            Write-Error "服务启动超时"
            return $false
        }
        
        Start-Sleep -Seconds 1
    }
}

# 等待服务停止
function Wait-ForServiceStop {
    param([int]$TimeoutSeconds)
    
    Write-Info "等待服务停止..."
    
    $startTime = Get-Date
    
    while ($true) {
        if (-not (Test-ServiceRunning)) {
            Write-Success "服务停止成功"
            return $true
        }
        
        $elapsed = (Get-Date) - $startTime
        if ($elapsed.TotalSeconds -ge $TimeoutSeconds) {
            Write-Error "服务停止超时"
            return $false
        }
        
        Start-Sleep -Seconds 1
    }
}

# 启动服务
function Start-Service {
    Write-Info "启动监控服务..."
    
    if (Test-ServiceRunning) {
        Write-Warning "服务已在运行中"
        return
    }
    
    # 检查配置文件
    $configFile = Get-ConfigFile
    if (-not (Test-Path $configFile)) {
        Write-Error "配置文件不存在: $configFile"
        exit 1
    }
    
    # 创建日志目录
    $logFile = Get-LogFile
    $logDir = Split-Path $logFile -Parent
    if (-not (Test-Path $logDir)) {
        New-Item -ItemType Directory -Path $logDir -Force | Out-Null
    }
    
    # 设置环境变量
    $env:MONITORING_ENV = $Environment
    $env:MONITORING_PORT = $Port
    $env:MONITORING_METRICS_PORT = $MetricsPort
    $env:MONITORING_LOG_LEVEL = $LogLevel
    $env:MONITORING_CONFIG = $configFile
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 启动服务"
        Write-Info "环境: $Environment"
        Write-Info "端口: $Port"
        Write-Info "指标端口: $MetricsPort"
        Write-Info "日志级别: $LogLevel"
        Write-Info "配置文件: $configFile"
        return
    }
    
    # 启动服务
    $exePath = "$ProjectRoot\bin\monitoring.exe"
    if (-not (Test-Path $exePath)) {
        Write-Error "可执行文件不存在: $exePath"
        exit 1
    }
    
    if ($Daemon) {
        # 后台运行
        $process = Start-Process -FilePath $exePath -WindowStyle Hidden -PassThru -RedirectStandardOutput $logFile -RedirectStandardError $logFile
        $process.Id | Out-File -FilePath (Get-PidFile) -Encoding ASCII
        Write-Info "服务已在后台启动，PID: $($process.Id)"
        
        # 等待服务启动
        if (-not (Wait-ForService $Timeout)) {
            Stop-Service
            exit 1
        }
    } else {
        # 前台运行
        & $exePath
    }
}

# 停止服务
function Stop-Service {
    Write-Info "停止监控服务..."
    
    if (-not (Test-ServiceRunning)) {
        Write-Warning "服务未运行"
        return
    }
    
    $pid = Get-ServicePid
    $pidFile = Get-PidFile
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 停止服务，PID: $pid"
        return
    }
    
    # 停止进程
    try {
        Write-Info "停止进程 $pid"
        Stop-Process -Id $pid -Force:$Force -ErrorAction Stop
        
        # 等待服务停止
        if (Wait-ForServiceStop $Timeout) {
            Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
            return
        }
    } catch {
        Write-Error "停止服务失败: $_"
    }
    
    # 强制停止
    if ($Force) {
        Write-Warning "强制停止服务"
        try {
            Stop-Process -Id $pid -Force -ErrorAction Stop
            Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
            Write-Success "服务已强制停止"
        } catch {
            Write-Error "强制停止失败: $_"
            exit 1
        }
    } else {
        Write-Error "服务停止失败，使用 -Force 强制停止"
        exit 1
    }
}

# 重启服务
function Restart-Service {
    Write-Info "重启监控服务..."
    
    if (Test-ServiceRunning) {
        Stop-Service
    }
    
    Start-Service
}

# 查看服务状态
function Show-Status {
    Write-Info "查看服务状态..."
    
    if (Test-ServiceRunning) {
        $pid = Get-ServicePid
        Write-Success "服务正在运行，PID: $pid"
        
        # 显示详细信息
        if ($Verbose) {
            Write-Host ""
            Write-Host "详细信息:"
            Write-Host "  环境: $Environment"
            Write-Host "  端口: $Port"
            Write-Host "  指标端口: $MetricsPort"
            Write-Host "  日志级别: $LogLevel"
            Write-Host "  配置文件: $(Get-ConfigFile)"
            Write-Host "  日志文件: $(Get-LogFile)"
            Write-Host "  PID 文件: $(Get-PidFile)"
            
            # 显示进程信息
            try {
                $process = Get-Process -Id $pid -ErrorAction Stop
                Write-Host ""
                Write-Host "进程信息:"
                Write-Host "  进程名: $($process.ProcessName)"
                Write-Host "  启动时间: $($process.StartTime)"
                Write-Host "  CPU 时间: $($process.TotalProcessorTime)"
                Write-Host "  内存使用: $([math]::Round($process.WorkingSet64 / 1MB, 2))MB"
            } catch {
                Write-Warning "无法获取进程信息"
            }
            
            # 显示端口监听
            try {
                $connections = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
                if ($connections) {
                    Write-Host ""
                    Write-Host "端口监听:"
                    Write-Host "  HTTP 端口 $Port : 监听中"
                }
                
                $metricsConnections = Get-NetTCPConnection -LocalPort $MetricsPort -ErrorAction SilentlyContinue
                if ($metricsConnections) {
                    Write-Host "  指标端口 $MetricsPort : 监听中"
                }
            } catch {
                Write-Debug "无法获取端口信息"
            }
        }
    } else {
        Write-Error "服务未运行"
        exit 1
    }
}

# 查看日志
function Show-Logs {
    $logFile = Get-LogFile
    
    if (-not (Test-Path $logFile)) {
        Write-Error "日志文件不存在: $logFile"
        exit 1
    }
    
    Write-Info "查看服务日志: $logFile"
    
    if ($Follow) {
        Get-Content $logFile -Tail $Lines -Wait
    } else {
        Get-Content $logFile -Tail $Lines
    }
}

# 健康检查
function Test-Health {
    Write-Info "执行健康检查..."
    
    if (-not (Test-ServiceRunning)) {
        Write-Error "服务未运行"
        exit 1
    }
    
    # 检查 HTTP 端点
    $healthUrl = "http://localhost:${Port}/health"
    $readyUrl = "http://localhost:${Port}/ready"
    $liveUrl = "http://localhost:${Port}/live"
    
    Write-Info "检查健康端点: $healthUrl"
    try {
        $response = Invoke-WebRequest -Uri $healthUrl -UseBasicParsing -TimeoutSec $Timeout
        if ($response.StatusCode -eq 200) {
            Write-Success "健康检查通过"
        } else {
            Write-Error "健康检查失败，状态码: $($response.StatusCode)"
            exit 1
        }
    } catch {
        Write-Error "健康检查失败: $_"
        exit 1
    }
    
    Write-Info "检查就绪端点: $readyUrl"
    try {
        $response = Invoke-WebRequest -Uri $readyUrl -UseBasicParsing -TimeoutSec $Timeout
        if ($response.StatusCode -eq 200) {
            Write-Success "就绪检查通过"
        } else {
            Write-Error "就绪检查失败，状态码: $($response.StatusCode)"
            exit 1
        }
    } catch {
        Write-Error "就绪检查失败: $_"
        exit 1
    }
    
    Write-Info "检查存活端点: $liveUrl"
    try {
        $response = Invoke-WebRequest -Uri $liveUrl -UseBasicParsing -TimeoutSec $Timeout
        if ($response.StatusCode -eq 200) {
            Write-Success "存活检查通过"
        } else {
            Write-Error "存活检查失败，状态码: $($response.StatusCode)"
            exit 1
        }
    } catch {
        Write-Error "存活检查失败: $_"
        exit 1
    }
    
    # 检查指标端点
    $metricsUrl = "http://localhost:${MetricsPort}/metrics"
    Write-Info "检查指标端点: $metricsUrl"
    try {
        $response = Invoke-WebRequest -Uri $metricsUrl -UseBasicParsing -TimeoutSec $Timeout
        if ($response.StatusCode -eq 200) {
            Write-Success "指标端点正常"
        } else {
            Write-Error "指标端点异常，状态码: $($response.StatusCode)"
            exit 1
        }
    } catch {
        Write-Error "指标端点异常: $_"
        exit 1
    }
    
    Write-Success "所有健康检查通过"
}

# 查看指标
function Show-Metrics {
    Write-Info "查看服务指标..."
    
    if (-not (Test-ServiceRunning)) {
        Write-Error "服务未运行"
        exit 1
    }
    
    $metricsUrl = "http://localhost:${MetricsPort}/metrics"
    
    try {
        $response = Invoke-WebRequest -Uri $metricsUrl -UseBasicParsing -TimeoutSec $Timeout
        Write-Host $response.Content
    } catch {
        Write-Error "获取指标失败: $_"
        exit 1
    }
}

# 配置管理
function Manage-Config {
    Write-Info "配置管理..."
    
    $configFile = Get-ConfigFile
    
    Write-Host "配置文件: $configFile"
    
    if (Test-Path $configFile) {
        Write-Host "配置内容:"
        Get-Content $configFile
    } else {
        Write-Error "配置文件不存在"
        exit 1
    }
}

# 安装服务
function Install-Service {
    Write-Info "安装监控服务..."
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 安装服务"
        return
    }
    
    # 创建目录
    $dirs = @(
        "C:\Program Files\monitoring",
        "C:\Logs\monitoring",
        "C:\ProgramData\monitoring",
        "C:\monitoring\config"
    )
    
    foreach ($dir in $dirs) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
            Write-Info "创建目录: $dir"
        }
    }
    
    # 复制文件
    Copy-Item "$ProjectRoot\bin\monitoring.exe" "C:\Program Files\monitoring\" -Force
    Copy-Item "$ProjectRoot\config\*.yml" "C:\monitoring\config\" -Force
    
    # 创建 Windows 服务
    try {
        $servicePath = "C:\Program Files\monitoring\monitoring.exe"
        $serviceParams = @{
            Name = $ServiceName
            BinaryPathName = $servicePath
            DisplayName = "Taishanglaojun Monitoring Service"
            Description = "太上老君监控系统服务"
            StartupType = "Automatic"
        }
        
        New-Service @serviceParams -ErrorAction Stop
        Write-Success "Windows 服务创建成功"
    } catch {
        Write-Error "创建 Windows 服务失败: $_"
        exit 1
    }
    
    Write-Success "服务安装完成"
}

# 卸载服务
function Uninstall-Service {
    Write-Info "卸载监控服务..."
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 卸载服务"
        return
    }
    
    # 停止并删除 Windows 服务
    try {
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Remove-Service -Name $ServiceName -ErrorAction Stop
        Write-Info "Windows 服务删除成功"
    } catch {
        Write-Warning "删除 Windows 服务失败: $_"
    }
    
    # 删除文件和目录
    Remove-Item "C:\Program Files\monitoring" -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item "C:\monitoring\config" -Recurse -Force -ErrorAction SilentlyContinue
    
    if ($Force) {
        Remove-Item "C:\Logs\monitoring" -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item "C:\ProgramData\monitoring" -Recurse -Force -ErrorAction SilentlyContinue
    }
    
    Write-Success "服务卸载完成"
}

# 更新服务
function Update-Service {
    Write-Info "更新监控服务..."
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 更新服务"
        return
    }
    
    # 停止服务
    if (Test-ServiceRunning) {
        Stop-Service
    }
    
    # 备份当前版本
    $currentExe = "C:\Program Files\monitoring\monitoring.exe"
    if (Test-Path $currentExe) {
        $backupName = "monitoring.backup.$(Get-Date -Format 'yyyyMMdd-HHmmss').exe"
        Copy-Item $currentExe "C:\Program Files\monitoring\$backupName" -Force
        Write-Info "备份当前版本: $backupName"
    }
    
    # 更新二进制文件
    Copy-Item "$ProjectRoot\bin\monitoring.exe" "C:\Program Files\monitoring\" -Force
    
    # 更新配置文件
    Copy-Item "$ProjectRoot\config\*.yml" "C:\monitoring\config\" -Force
    
    # 重启服务
    Start-Service
    
    Write-Success "服务更新完成"
}

# 备份数据
function Backup-Data {
    Write-Info "备份监控数据..."
    
    & "$ScriptDir\backup.ps1" backup -Type all -Environment $Environment
}

# 恢复数据
function Restore-Data {
    Write-Info "恢复监控数据..."
    
    & "$ScriptDir\backup.ps1" restore -Environment $Environment
}

# 主函数
function Main {
    # 如果启用详细输出
    if ($Verbose) {
        $VerbosePreference = "Continue"
    }
    
    # 检查依赖
    Test-Dependencies
    
    # 执行命令
    switch ($Command) {
        "start" {
            Start-Service
        }
        "stop" {
            Stop-Service
        }
        "restart" {
            Restart-Service
        }
        "status" {
            Show-Status
        }
        "logs" {
            Show-Logs
        }
        "health" {
            Test-Health
        }
        "metrics" {
            Show-Metrics
        }
        "config" {
            Manage-Config
        }
        "install" {
            Install-Service
        }
        "uninstall" {
            Uninstall-Service
        }
        "update" {
            Update-Service
        }
        "backup" {
            Backup-Data
        }
        "restore" {
            Restore-Data
        }
        "help" {
            Show-Help
        }
        default {
            Write-Error "未知命令: $Command"
            Show-Help
            exit 1
        }
    }
}

# 执行主函数
Main