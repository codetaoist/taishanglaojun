# 太上老君监控系统 Windows 安装脚本
# PowerShell 5.0+ 支持

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("install", "uninstall", "start", "stop", "restart", "status")]
    [string]$Action = "install",
    
    [Parameter(Mandatory=$false)]
    [string]$InstallPath = "C:\Program Files\TaiShangLaoJun\Monitoring",
    
    [Parameter(Mandatory=$false)]
    [string]$ConfigPath = "C:\ProgramData\TaiShangLaoJun\Monitoring",
    
    [Parameter(Mandatory=$false)]
    [string]$ServiceName = "TaiShangLaoJunMonitoring",
    
    [Parameter(Mandatory=$false)]
    [string]$Version = "1.0.0"
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色定义
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

# 日志函数
function Write-Log {
    param(
        [string]$Message,
        [string]$Level = "INFO",
        [string]$Color = "White"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    
    Write-Host $logMessage -ForegroundColor $Color
}

function Write-Info {
    param([string]$Message)
    Write-Log -Message $Message -Level "INFO" -Color $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Log -Message $Message -Level "SUCCESS" -Color $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Log -Message $Message -Level "WARNING" -Color $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Log -Message $Message -Level "ERROR" -Color $Colors.Red
}

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 检查权限
function Test-Permissions {
    Write-Info "检查权限..."
    
    if (-not (Test-Administrator)) {
        Write-Error "此脚本需要管理员权限运行"
        Write-Info "请以管理员身份运行 PowerShell 后重试"
        exit 1
    }
    
    Write-Success "权限检查通过"
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    $dependencies = @("curl", "tar")
    $missingDeps = @()
    
    foreach ($dep in $dependencies) {
        if (-not (Get-Command $dep -ErrorAction SilentlyContinue)) {
            $missingDeps += $dep
        }
    }
    
    if ($missingDeps.Count -gt 0) {
        Write-Warning "缺少依赖: $($missingDeps -join ', ')"
        Write-Info "尝试安装缺少的依赖..."
        
        # 安装 curl（Windows 10 1803+ 内置）
        if ("curl" -in $missingDeps) {
            if (-not (Get-Command curl -ErrorAction SilentlyContinue)) {
                Write-Error "curl 不可用，请手动安装或升级 Windows"
                exit 1
            }
        }
        
        # 安装 tar（Windows 10 1903+ 内置）
        if ("tar" -in $missingDeps) {
            if (-not (Get-Command tar -ErrorAction SilentlyContinue)) {
                Write-Error "tar 不可用，请手动安装或升级 Windows"
                exit 1
            }
        }
    }
    
    Write-Success "依赖检查通过"
}

# 创建目录
function New-Directories {
    Write-Info "创建目录..."
    
    $directories = @(
        $InstallPath,
        $ConfigPath,
        "$ConfigPath\logs",
        "$ConfigPath\data"
    )
    
    foreach ($dir in $directories) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
            Write-Info "创建目录: $dir"
        }
    }
    
    Write-Success "目录创建完成"
}

# 下载二进制文件
function Get-Binary {
    Write-Info "下载二进制文件..."
    
    $binaryName = "taishanglaojun-monitoring-windows-amd64.exe"
    $downloadUrl = "https://github.com/taishanglaojun/core-services/releases/download/v$Version/$binaryName"
    $downloadPath = "$env:TEMP\$binaryName"
    $targetPath = "$InstallPath\taishanglaojun-monitoring.exe"
    
    try {
        Write-Info "从 $downloadUrl 下载..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $downloadPath -UseBasicParsing
        Write-Success "下载完成"
        
        # 移动文件
        Move-Item -Path $downloadPath -Destination $targetPath -Force
        Write-Success "二进制文件安装完成"
    }
    catch {
        Write-Error "下载失败: $($_.Exception.Message)"
        exit 1
    }
}

# 创建配置文件
function New-Config {
    Write-Info "创建配置文件..."
    
    $configContent = @"
# 太上老君监控系统配置文件

service:
  name: "taishanglaojun-monitoring"
  version: "$Version"
  environment: "production"
  host: "0.0.0.0"
  port: 8080
  log_level: "info"
  
tracing:
  enabled: true
  sampling_rate: 0.1
  batch_timeout: "5s"
  max_export_batch_size: 512
  max_queue_size: 2048
  exporters:
    console:
      enabled: false
    jaeger:
      enabled: true
      endpoint: "http://localhost:14268/api/traces"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "$($ConfigPath -replace '\\', '/')/logs/app.log"
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true

storage:
  prometheus:
    url: "http://localhost:9090"
    timeout: "30s"
  influxdb:
    url: "http://localhost:8086"
    token: "monitoring-token-123456789"
    org: "taishanglaojun"
    bucket: "monitoring"

alerting:
  enabled: true
  evaluation_interval: "30s"
  notification_channels:
    email:
      enabled: true
      smtp_host: "localhost"
      smtp_port: 587
      from: "monitoring@taishanglaojun.com"
      to: ["admin@taishanglaojun.com"]

dashboard:
  enabled: true
  refresh_interval: "30s"

performance:
  enabled: true
  collection_interval: "15s"
  
automation:
  enabled: true
  max_concurrent_workflows: 10
"@
    
    $configFile = "$ConfigPath\monitoring.yaml"
    $configContent | Out-File -FilePath $configFile -Encoding UTF8
    
    Write-Success "配置文件创建完成: $configFile"
}

# 创建 Windows 服务
function New-Service {
    Write-Info "创建 Windows 服务..."
    
    $servicePath = "$InstallPath\taishanglaojun-monitoring.exe"
    $serviceArgs = "-config `"$ConfigPath\monitoring.yaml`""
    
    # 检查服务是否已存在
    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Warning "服务 $ServiceName 已存在，正在删除..."
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        & sc.exe delete $ServiceName
        Start-Sleep -Seconds 2
    }
    
    # 创建服务
    & sc.exe create $ServiceName binPath= "`"$servicePath`" $serviceArgs" start= auto DisplayName= "太上老君监控系统"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Windows 服务创建完成"
        
        # 设置服务描述
        & sc.exe description $ServiceName "太上老君监控系统 - 提供分布式追踪、日志聚合、智能告警等功能"
        
        # 设置服务恢复选项
        & sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/30000
    }
    else {
        Write-Error "Windows 服务创建失败"
        exit 1
    }
}

# 启动服务
function Start-MonitoringService {
    Write-Info "启动服务..."
    
    try {
        Start-Service -Name $ServiceName
        Start-Sleep -Seconds 5
        
        $service = Get-Service -Name $ServiceName
        if ($service.Status -eq "Running") {
            Write-Success "服务启动成功"
            
            # 检查健康状态
            Write-Info "检查服务健康状态..."
            Start-Sleep -Seconds 10
            
            try {
                $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 10
                if ($response.StatusCode -eq 200) {
                    Write-Success "服务健康检查通过"
                }
                else {
                    Write-Warning "服务健康检查失败，状态码: $($response.StatusCode)"
                }
            }
            catch {
                Write-Warning "服务健康检查失败: $($_.Exception.Message)"
            }
        }
        else {
            Write-Error "服务启动失败，当前状态: $($service.Status)"
            exit 1
        }
    }
    catch {
        Write-Error "服务启动失败: $($_.Exception.Message)"
        exit 1
    }
}

# 停止服务
function Stop-MonitoringService {
    Write-Info "停止服务..."
    
    try {
        $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
        if ($service -and $service.Status -eq "Running") {
            Stop-Service -Name $ServiceName -Force
            Write-Success "服务已停止"
        }
        else {
            Write-Info "服务未运行"
        }
    }
    catch {
        Write-Error "停止服务失败: $($_.Exception.Message)"
    }
}

# 重启服务
function Restart-MonitoringService {
    Stop-MonitoringService
    Start-Sleep -Seconds 2
    Start-MonitoringService
}

# 查看服务状态
function Get-ServiceStatus {
    Write-Info "查看服务状态..."
    
    try {
        $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
        if ($service) {
            Write-Host "服务名称: $($service.Name)" -ForegroundColor White
            Write-Host "显示名称: $($service.DisplayName)" -ForegroundColor White
            Write-Host "服务状态: $($service.Status)" -ForegroundColor $(if ($service.Status -eq "Running") { $Colors.Green } else { $Colors.Red })
            Write-Host "启动类型: $($service.StartType)" -ForegroundColor White
            
            # 检查端口
            $port = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
            if ($port) {
                Write-Host "端口 8080: 监听中" -ForegroundColor $Colors.Green
            }
            else {
                Write-Host "端口 8080: 未监听" -ForegroundColor $Colors.Red
            }
        }
        else {
            Write-Warning "服务 $ServiceName 不存在"
        }
    }
    catch {
        Write-Error "获取服务状态失败: $($_.Exception.Message)"
    }
}

# 卸载
function Remove-Installation {
    Write-Info "开始卸载..."
    
    # 停止服务
    Stop-MonitoringService
    
    # 删除服务
    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($service) {
        & sc.exe delete $ServiceName
        if ($LASTEXITCODE -eq 0) {
            Write-Success "服务已删除"
        }
    }
    
    # 删除文件和目录
    if (Test-Path $InstallPath) {
        Remove-Item -Path $InstallPath -Recurse -Force
        Write-Success "安装目录已删除: $InstallPath"
    }
    
    if (Test-Path $ConfigPath) {
        $response = Read-Host "是否删除配置和数据目录? ($ConfigPath) [y/N]"
        if ($response -eq "y" -or $response -eq "Y") {
            Remove-Item -Path $ConfigPath -Recurse -Force
            Write-Success "配置目录已删除: $ConfigPath"
        }
    }
    
    Write-Success "卸载完成"
}

# 显示安装信息
function Show-InstallationInfo {
    Write-Success "安装完成！"
    Write-Host ""
    Write-Host "服务信息:" -ForegroundColor $Colors.White
    Write-Host "  名称: $ServiceName" -ForegroundColor $Colors.White
    Write-Host "  版本: $Version" -ForegroundColor $Colors.White
    Write-Host "  安装目录: $InstallPath" -ForegroundColor $Colors.White
    Write-Host "  配置目录: $ConfigPath" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "常用命令:" -ForegroundColor $Colors.White
    Write-Host "  启动服务: .\setup.ps1 -Action start" -ForegroundColor $Colors.Blue
    Write-Host "  停止服务: .\setup.ps1 -Action stop" -ForegroundColor $Colors.Blue
    Write-Host "  重启服务: .\setup.ps1 -Action restart" -ForegroundColor $Colors.Blue
    Write-Host "  查看状态: .\setup.ps1 -Action status" -ForegroundColor $Colors.Blue
    Write-Host "  卸载系统: .\setup.ps1 -Action uninstall" -ForegroundColor $Colors.Blue
    Write-Host ""
    Write-Host "Web 界面:" -ForegroundColor $Colors.White
    Write-Host "  监控面板: http://localhost:8080" -ForegroundColor $Colors.Green
    Write-Host "  健康检查: http://localhost:8080/health" -ForegroundColor $Colors.Green
    Write-Host "  指标接口: http://localhost:8080/metrics" -ForegroundColor $Colors.Green
    Write-Host ""
    Write-Host "配置文件: $ConfigPath\monitoring.yaml" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Info "请根据需要修改配置文件，然后重启服务"
}

# 主函数
function Main {
    Write-Host "太上老君监控系统 Windows 安装脚本 v$Version" -ForegroundColor $Colors.Blue
    Write-Host "================================================" -ForegroundColor $Colors.Blue
    Write-Host ""
    
    switch ($Action.ToLower()) {
        "install" {
            Test-Permissions
            Test-Dependencies
            New-Directories
            Get-Binary
            New-Config
            New-Service
            Start-MonitoringService
            Show-InstallationInfo
        }
        "uninstall" {
            Test-Permissions
            Remove-Installation
        }
        "start" {
            Test-Permissions
            Start-MonitoringService
        }
        "stop" {
            Test-Permissions
            Stop-MonitoringService
        }
        "restart" {
            Test-Permissions
            Restart-MonitoringService
        }
        "status" {
            Get-ServiceStatus
        }
        default {
            Write-Host "用法: .\setup.ps1 -Action [install|uninstall|start|stop|restart|status]" -ForegroundColor $Colors.Yellow
            Write-Host "  install   - 安装监控系统（默认）" -ForegroundColor $Colors.White
            Write-Host "  uninstall - 卸载监控系统" -ForegroundColor $Colors.White
            Write-Host "  start     - 启动服务" -ForegroundColor $Colors.White
            Write-Host "  stop      - 停止服务" -ForegroundColor $Colors.White
            Write-Host "  restart   - 重启服务" -ForegroundColor $Colors.White
            Write-Host "  status    - 查看状态" -ForegroundColor $Colors.White
            exit 1
        }
    }
}

# 执行主函数
try {
    Main
}
catch {
    Write-Error "脚本执行失败: $($_.Exception.Message)"
    exit 1
}