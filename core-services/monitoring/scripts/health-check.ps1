# 太上老君监控系统健康检查 PowerShell 脚本
# 用于检查应用和依赖服务的健康状态

param(
    [Parameter(Position = 0)]
    [ValidateSet("check", "app", "deps", "db", "redis", "metrics", "logs", "all", "help")]
    [string]$Command = "check",
    
    [Alias("e")]
    [ValidateSet("development", "staging", "production")]
    [string]$Environment = "",
    
    [Alias("h")]
    [string]$Host = $env:MONITORING_HOST,
    
    [Alias("p")]
    [int]$Port = [int]$env:MONITORING_PORT,
    
    [Alias("t")]
    [string]$Timeout = $env:HEALTH_CHECK_TIMEOUT,
    
    [Alias("r")]
    [int]$Retry = 3,
    
    [Alias("i")]
    [string]$Interval = "5s",
    
    [Alias("v")]
    [switch]$Verbose,
    
    [switch]$Json,
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

# 设置默认值
if (-not $Host) {
    $Host = "localhost"
}

if ($Port -eq 0) {
    $Port = 8080
}

if (-not $Timeout) {
    $Timeout = "30s"
}

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    if ($NoColor) {
        Write-Host $Message
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

# 显示帮助信息
function Show-Help {
    @"
太上老君监控系统健康检查 PowerShell 脚本

用法: .\health-check.ps1 [选项] [命令]

命令:
    check       执行健康检查
    app         检查应用健康状态
    deps        检查依赖服务
    db          检查数据库连接
    redis       检查 Redis 连接
    metrics     检查指标端点
    logs        检查日志输出
    all         执行所有检查
    help        显示此帮助信息

选项:
    -Environment, -e ENV    环境 (development|staging|production)
    -Host, -h HOST         主机地址 (默认: localhost)
    -Port, -p PORT         端口号 (默认: 8080)
    -Timeout, -t TIMEOUT   超时时间 (默认: 30s)
    -Retry, -r RETRY       重试次数 (默认: 3)
    -Interval, -i INTERVAL 重试间隔 (默认: 5s)
    -Verbose, -v           详细输出
    -Json                  JSON 格式输出
    -NoColor               禁用颜色输出
    -Help                  显示帮助信息

环境变量:
    MONITORING_HOST        监控服务主机
    MONITORING_PORT        监控服务端口
    DATABASE_URL           数据库连接地址
    REDIS_URL              Redis 连接地址
    HEALTH_CHECK_TIMEOUT   健康检查超时时间

示例:
    .\health-check.ps1 check
    .\health-check.ps1 app -Host localhost -Port 8080
    .\health-check.ps1 deps -Environment production
    .\health-check.ps1 all -Json

"@
}

# 解析超时时间
function ConvertTo-Seconds {
    param([string]$TimeString)
    
    # 如果已经是数字，直接返回
    if ($TimeString -match '^\d+$') {
        return [int]$TimeString
    }
    
    # 解析时间单位
    if ($TimeString -match '^(\d+)([a-zA-Z]+)$') {
        $value = [int]$Matches[1]
        $unit = $Matches[2].ToLower()
        
        switch ($unit) {
            { $_ -in @('s', 'sec', 'second', 'seconds') } { return $value }
            { $_ -in @('m', 'min', 'minute', 'minutes') } { return $value * 60 }
            { $_ -in @('h', 'hour', 'hours') } { return $value * 3600 }
            default { return 30 }  # 默认 30 秒
        }
    }
    
    return 30  # 默认 30 秒
}

# HTTP 请求函数
function Invoke-HttpRequest {
    param(
        [string]$Url,
        [int]$TimeoutSeconds,
        [string]$Method = "GET",
        [hashtable]$Headers = @{}
    )
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            TimeoutSec = $TimeoutSeconds
            UseBasicParsing = $true
        }
        
        if ($Headers.Count -gt 0) {
            $params.Headers = $Headers
        }
        
        $response = Invoke-WebRequest @params
        return @{
            Success = $true
            Content = $response.Content
            StatusCode = $response.StatusCode
        }
    } catch {
        return @{
            Success = $false
            Error = $_.Exception.Message
            StatusCode = 0
        }
    }
}

# 检查应用健康状态
function Test-AppHealth {
    param(
        [string]$Host,
        [int]$Port,
        [int]$TimeoutSeconds,
        [bool]$Verbose,
        [bool]$JsonOutput
    )
    
    Write-Info "检查应用健康状态..."
    
    $healthUrl = "http://${Host}:${Port}/health"
    $readyUrl = "http://${Host}:${Port}/ready"
    $liveUrl = "http://${Host}:${Port}/live"
    
    $healthStatus = "unknown"
    $readyStatus = "unknown"
    $liveStatus = "unknown"
    $responseTime = 0
    
    # 检查健康端点
    $startTime = Get-Date
    
    $healthResponse = Invoke-HttpRequest -Url $healthUrl -TimeoutSeconds $TimeoutSeconds
    if ($healthResponse.Success) {
        $healthStatus = "healthy"
        if ($Verbose) {
            Write-Info "健康检查响应: $($healthResponse.Content)"
        }
    } else {
        $healthStatus = "unhealthy"
        Write-Error "健康检查失败: $healthUrl"
    }
    
    $endTime = Get-Date
    $responseTime = ($endTime - $startTime).TotalSeconds
    
    # 检查就绪端点
    $readyResponse = Invoke-HttpRequest -Url $readyUrl -TimeoutSeconds $TimeoutSeconds
    if ($readyResponse.Success) {
        $readyStatus = "ready"
    } else {
        $readyStatus = "not_ready"
        Write-Warning "就绪检查失败: $readyUrl"
    }
    
    # 检查存活端点
    $liveResponse = Invoke-HttpRequest -Url $liveUrl -TimeoutSeconds $TimeoutSeconds
    if ($liveResponse.Success) {
        $liveStatus = "alive"
    } else {
        $liveStatus = "dead"
        Write-Error "存活检查失败: $liveUrl"
    }
    
    # 输出结果
    if ($JsonOutput) {
        $result = @{
            app = @{
                health = $healthStatus
                ready = $readyStatus
                live = $liveStatus
                response_time = $responseTime
                endpoints = @{
                    health = $healthUrl
                    ready = $readyUrl
                    live = $liveUrl
                }
            }
        }
        $result | ConvertTo-Json -Depth 10
    } else {
        Write-Host "应用健康状态:"
        Write-Host "  健康: $healthStatus"
        Write-Host "  就绪: $readyStatus"
        Write-Host "  存活: $liveStatus"
        Write-Host "  响应时间: ${responseTime}s"
    }
    
    # 返回状态
    return ($healthStatus -eq "healthy" -and $readyStatus -eq "ready" -and $liveStatus -eq "alive")
}

# 检查数据库连接
function Test-Database {
    param(
        [int]$TimeoutSeconds,
        [bool]$Verbose,
        [bool]$JsonOutput
    )
    
    Write-Info "检查数据库连接..."
    
    $dbStatus = "unknown"
    $dbVersion = ""
    $connectionTime = 0
    
    # 获取数据库连接信息
    $dbUrl = $env:DATABASE_URL
    
    if (-not $dbUrl) {
        $dbStatus = "not_configured"
        Write-Warning "数据库连接未配置"
    } else {
        $startTime = Get-Date
        
        # 尝试连接数据库
        try {
            if ($dbUrl -match "postgres") {
                # PostgreSQL
                if (Get-Command psql -ErrorAction SilentlyContinue) {
                    $dbVersion = & psql $dbUrl -t -c "SELECT version();" 2>$null | Select-Object -First 1
                    if ($dbVersion) {
                        $dbStatus = "connected"
                        if ($Verbose) {
                            Write-Info "数据库版本: $dbVersion"
                        }
                    } else {
                        $dbStatus = "connection_failed"
                        Write-Error "PostgreSQL 连接失败"
                    }
                } else {
                    $dbStatus = "client_not_found"
                    Write-Warning "PostgreSQL 客户端未安装"
                }
            } elseif ($dbUrl -match "mysql") {
                # MySQL
                if (Get-Command mysql -ErrorAction SilentlyContinue) {
                    $dbVersion = & mysql --version 2>$null | Select-Object -First 1
                    if ($dbVersion) {
                        $dbStatus = "connected"
                        if ($Verbose) {
                            Write-Info "数据库版本: $dbVersion"
                        }
                    } else {
                        $dbStatus = "connection_failed"
                        Write-Error "MySQL 连接失败"
                    }
                } else {
                    $dbStatus = "client_not_found"
                    Write-Warning "MySQL 客户端未安装"
                }
            } else {
                $dbStatus = "unsupported"
                Write-Warning "不支持的数据库类型"
            }
        } catch {
            $dbStatus = "error"
            Write-Error "数据库检查错误: $_"
        }
        
        $endTime = Get-Date
        $connectionTime = ($endTime - $startTime).TotalSeconds
    }
    
    # 输出结果
    if ($JsonOutput) {
        $result = @{
            database = @{
                status = $dbStatus
                version = $dbVersion
                connection_time = $connectionTime
            }
        }
        $result | ConvertTo-Json -Depth 10
    } else {
        Write-Host "数据库状态:"
        Write-Host "  状态: $dbStatus"
        Write-Host "  版本: $dbVersion"
        Write-Host "  连接时间: ${connectionTime}s"
    }
    
    # 返回状态
    return ($dbStatus -eq "connected")
}

# 检查 Redis 连接
function Test-Redis {
    param(
        [int]$TimeoutSeconds,
        [bool]$Verbose,
        [bool]$JsonOutput
    )
    
    Write-Info "检查 Redis 连接..."
    
    $redisStatus = "unknown"
    $redisVersion = ""
    $connectionTime = 0
    
    # 获取 Redis 连接信息
    $redisUrl = $env:REDIS_URL
    
    if (-not $redisUrl) {
        $redisStatus = "not_configured"
        Write-Warning "Redis 连接未配置"
    } else {
        $startTime = Get-Date
        
        # 尝试连接 Redis
        try {
            if (Get-Command redis-cli -ErrorAction SilentlyContinue) {
                $redisVersion = & redis-cli --version 2>$null | Select-Object -First 1
                if ($redisVersion) {
                    # 测试 PING 命令
                    $pingResult = & redis-cli -u $redisUrl ping 2>$null
                    if ($pingResult -eq "PONG") {
                        $redisStatus = "connected"
                        if ($Verbose) {
                            Write-Info "Redis 版本: $redisVersion"
                        }
                    } else {
                        $redisStatus = "connection_failed"
                        Write-Error "Redis PING 失败"
                    }
                } else {
                    $redisStatus = "client_error"
                    Write-Error "Redis 客户端错误"
                }
            } else {
                $redisStatus = "client_not_found"
                Write-Warning "Redis 客户端未安装"
            }
        } catch {
            $redisStatus = "error"
            Write-Error "Redis 检查错误: $_"
        }
        
        $endTime = Get-Date
        $connectionTime = ($endTime - $startTime).TotalSeconds
    }
    
    # 输出结果
    if ($JsonOutput) {
        $result = @{
            redis = @{
                status = $redisStatus
                version = $redisVersion
                connection_time = $connectionTime
            }
        }
        $result | ConvertTo-Json -Depth 10
    } else {
        Write-Host "Redis 状态:"
        Write-Host "  状态: $redisStatus"
        Write-Host "  版本: $redisVersion"
        Write-Host "  连接时间: ${connectionTime}s"
    }
    
    # 返回状态
    return ($redisStatus -eq "connected")
}

# 检查指标端点
function Test-Metrics {
    param(
        [string]$Host,
        [int]$Port,
        [int]$TimeoutSeconds,
        [bool]$Verbose,
        [bool]$JsonOutput
    )
    
    Write-Info "检查指标端点..."
    
    $metricsUrl = "http://${Host}:9090/metrics"
    $metricsStatus = "unknown"
    $metricsCount = 0
    $responseTime = 0
    
    $startTime = Get-Date
    
    $response = Invoke-HttpRequest -Url $metricsUrl -TimeoutSeconds $TimeoutSeconds
    if ($response.Success) {
        $metricsStatus = "available"
        $metricsCount = ($response.Content -split "`n" | Where-Object { $_ -match "^[a-zA-Z]" }).Count
        
        if ($Verbose) {
            Write-Info "指标数量: $metricsCount"
        }
    } else {
        $metricsStatus = "unavailable"
        Write-Error "指标端点不可用: $metricsUrl"
    }
    
    $endTime = Get-Date
    $responseTime = ($endTime - $startTime).TotalSeconds
    
    # 输出结果
    if ($JsonOutput) {
        $result = @{
            metrics = @{
                status = $metricsStatus
                count = $metricsCount
                response_time = $responseTime
                endpoint = $metricsUrl
            }
        }
        $result | ConvertTo-Json -Depth 10
    } else {
        Write-Host "指标状态:"
        Write-Host "  状态: $metricsStatus"
        Write-Host "  指标数量: $metricsCount"
        Write-Host "  响应时间: ${responseTime}s"
    }
    
    # 返回状态
    return ($metricsStatus -eq "available")
}

# 检查日志输出
function Test-Logs {
    param(
        [string]$Environment,
        [bool]$Verbose,
        [bool]$JsonOutput
    )
    
    Write-Info "检查日志输出..."
    
    $logStatus = "unknown"
    $logLines = 0
    $errorCount = 0
    $warningCount = 0
    
    # 根据环境确定日志路径
    $logPath = switch ($Environment) {
        "development" { "$env:TEMP\monitoring.log" }
        { $_ -in @("staging", "production") } { "C:\logs\monitoring\app.log" }
        default { ".\logs\app.log" }
    }
    
    if (Test-Path $logPath) {
        $logStatus = "available"
        $logContent = Get-Content $logPath -ErrorAction SilentlyContinue
        if ($logContent) {
            $logLines = $logContent.Count
            $errorCount = ($logContent | Where-Object { $_ -match "ERROR" }).Count
            $warningCount = ($logContent | Where-Object { $_ -match "WARNING|WARN" }).Count
        }
        
        if ($Verbose) {
            Write-Info "日志文件: $logPath"
            Write-Info "日志行数: $logLines"
            Write-Info "错误数量: $errorCount"
            Write-Info "警告数量: $warningCount"
        }
    } else {
        $logStatus = "not_found"
        Write-Warning "日志文件不存在: $logPath"
    }
    
    # 输出结果
    if ($JsonOutput) {
        $result = @{
            logs = @{
                status = $logStatus
                path = $logPath
                lines = $logLines
                errors = $errorCount
                warnings = $warningCount
            }
        }
        $result | ConvertTo-Json -Depth 10
    } else {
        Write-Host "日志状态:"
        Write-Host "  状态: $logStatus"
        Write-Host "  路径: $logPath"
        Write-Host "  行数: $logLines"
        Write-Host "  错误: $errorCount"
        Write-Host "  警告: $warningCount"
    }
    
    # 返回状态
    return ($logStatus -eq "available")
}

# 执行所有检查
function Test-All {
    param(
        [string]$Environment,
        [string]$Host,
        [int]$Port,
        [int]$TimeoutSeconds,
        [bool]$Verbose,
        [bool]$JsonOutput
    )
    
    Write-Info "执行所有健康检查..."
    
    $results = @{}
    $overallStatus = "healthy"
    
    # 检查应用
    if (Test-AppHealth $Host $Port $TimeoutSeconds $Verbose $false) {
        $results.app = "healthy"
    } else {
        $results.app = "unhealthy"
        $overallStatus = "unhealthy"
    }
    
    # 检查数据库
    if (Test-Database $TimeoutSeconds $Verbose $false) {
        $results.database = "healthy"
    } else {
        $results.database = "unhealthy"
        $overallStatus = "unhealthy"
    }
    
    # 检查 Redis
    if (Test-Redis $TimeoutSeconds $Verbose $false) {
        $results.redis = "healthy"
    } else {
        $results.redis = "unhealthy"
        $overallStatus = "unhealthy"
    }
    
    # 检查指标
    if (Test-Metrics $Host $Port $TimeoutSeconds $Verbose $false) {
        $results.metrics = "healthy"
    } else {
        $results.metrics = "unhealthy"
        $overallStatus = "unhealthy"
    }
    
    # 检查日志
    if (Test-Logs $Environment $Verbose $false) {
        $results.logs = "healthy"
    } else {
        $results.logs = "unhealthy"
        $overallStatus = "unhealthy"
    }
    
    # 输出结果
    if ($JsonOutput) {
        $result = @{
            overall_status = $overallStatus
            timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
            checks = $results
        }
        $result | ConvertTo-Json -Depth 10
    } else {
        Write-Host ""
        Write-Host "========================================="
        Write-Host "健康检查汇总"
        Write-Host "========================================="
        Write-Host "总体状态: $overallStatus"
        Write-Host "检查时间: $(Get-Date)"
        Write-Host ""
        
        foreach ($check in $results.GetEnumerator()) {
            Write-Host "$($check.Key): $($check.Value)"
        }
        
        Write-Host "========================================="
    }
    
    # 返回状态
    return ($overallStatus -eq "healthy")
}

# 主函数
function Main {
    # 如果启用详细输出
    if ($Verbose) {
        $VerbosePreference = "Continue"
    }
    
    # 解析超时时间
    $timeoutSeconds = ConvertTo-Seconds $Timeout
    $intervalSeconds = ConvertTo-Seconds $Interval
    
    # 执行命令
    $exitCode = 0
    $attempt = 1
    
    while ($attempt -le $Retry) {
        if ($attempt -gt 1) {
            Write-Info "重试 $attempt/$Retry..."
            Start-Sleep -Seconds $intervalSeconds
        }
        
        $success = $false
        
        switch ($Command) {
            { $_ -in @("check", "app") } {
                $success = Test-AppHealth $Host $Port $timeoutSeconds $Verbose $Json
            }
            "deps" {
                $dbHealthy = Test-Database $timeoutSeconds $Verbose $Json
                $redisHealthy = Test-Redis $timeoutSeconds $Verbose $Json
                $success = $dbHealthy -and $redisHealthy
            }
            "db" {
                $success = Test-Database $timeoutSeconds $Verbose $Json
            }
            "redis" {
                $success = Test-Redis $timeoutSeconds $Verbose $Json
            }
            "metrics" {
                $success = Test-Metrics $Host $Port $timeoutSeconds $Verbose $Json
            }
            "logs" {
                $success = Test-Logs $Environment $Verbose $Json
            }
            "all" {
                $success = Test-All $Environment $Host $Port $timeoutSeconds $Verbose $Json
            }
            "help" {
                Show-Help
                return
            }
            default {
                Write-Error "未知命令: $Command"
                Show-Help
                exit 1
            }
        }
        
        if ($success) {
            break
        }
        
        $exitCode = 1
        $attempt++
    }
    
    if ($exitCode -ne 0) {
        Write-Error "健康检查失败，已重试 $Retry 次"
    }
    
    exit $exitCode
}

# 执行主函数
Main