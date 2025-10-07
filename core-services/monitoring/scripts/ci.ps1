# 太上老君监控系统 CI/CD PowerShell 脚本
# 支持构建、测试、部署的完整流水线

param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "test", "package", "deploy", "pipeline", "release", "rollback", "status", "clean", "help")]
    [string]$Command = "help",
    
    [Alias("e")]
    [ValidateSet("development", "staging", "production")]
    [string]$Environment = "",
    
    [Alias("v")]
    [string]$Version = "",
    
    [Alias("b")]
    [string]$Branch = "",
    
    [Alias("t")]
    [string]$Tag = "",
    
    [Alias("r")]
    [string]$Registry = $env:DOCKER_REGISTRY,
    
    [Alias("n")]
    [string]$Namespace = "",
    
    [switch]$SkipTests,
    [switch]$SkipBuild,
    [switch]$SkipPush,
    [switch]$Force,
    [switch]$DryRun,
    [switch]$Verbose,
    [string]$Revision = "",
    
    [Alias("h")]
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

# 如果没有指定注册表，使用默认值
if (-not $Registry) {
    $Registry = "ghcr.io"
}

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
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
太上老君监控系统 CI/CD PowerShell 脚本

用法: .\ci.ps1 [选项] [命令]

命令:
    build       构建应用
    test        运行测试
    package     打包应用
    deploy      部署应用
    pipeline    运行完整流水线
    release     发布版本
    rollback    回滚部署
    status      查看状态
    clean       清理环境
    help        显示此帮助信息

选项:
    -Environment, -e ENV    目标环境 (development|staging|production)
    -Version, -v VERSION    版本号
    -Branch, -b BRANCH      Git 分支
    -Tag, -t TAG           Git 标签
    -Registry, -r REG      Docker 镜像仓库
    -Namespace, -n NS      Kubernetes 命名空间
    -SkipTests             跳过测试
    -SkipBuild             跳过构建
    -SkipPush              跳过推送
    -Force                 强制执行
    -DryRun                仅显示将要执行的操作
    -Verbose               详细输出
    -Revision REV          回滚版本
    -Help, -h              显示帮助信息

环境变量:
    CI                     CI 环境标识
    GITHUB_TOKEN           GitHub Token
    DOCKER_REGISTRY        Docker 镜像仓库
    DOCKER_USERNAME        Docker 用户名
    DOCKER_PASSWORD        Docker 密码
    KUBECONFIG             Kubernetes 配置文件
    SLACK_WEBHOOK_URL      Slack 通知 Webhook

示例:
    .\ci.ps1 build -Version v1.0.0
    .\ci.ps1 test -Verbose
    .\ci.ps1 deploy -Environment production -Version v1.0.0
    .\ci.ps1 pipeline -Environment staging -Branch develop
    .\ci.ps1 release -Version v1.0.0 -Tag v1.0.0

"@
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    $missingDeps = @()
    
    # 检查必需工具
    $tools = @("git", "go", "docker", "kubectl", "helm")
    
    foreach ($tool in $tools) {
        if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
            $missingDeps += $tool
        }
    }
    
    if ($missingDeps.Count -gt 0) {
        Write-Error "缺少依赖: $($missingDeps -join ', ')"
        exit 1
    }
    
    Write-Success "依赖检查完成"
}

# 获取版本信息
function Get-VersionInfo {
    param(
        [string]$Version,
        [string]$Branch,
        [string]$Tag
    )
    
    # 如果没有指定版本，尝试从 Git 获取
    if (-not $Version) {
        if ($Tag) {
            $Version = $Tag
        } elseif (git describe --tags --exact-match HEAD 2>$null) {
            $Version = git describe --tags --exact-match HEAD
        } else {
            $commitHash = git rev-parse --short HEAD
            $currentBranch = if ($Branch) { $Branch } else { "main" }
            $Version = "$currentBranch-$commitHash"
        }
    }
    
    return $Version
}

# 设置环境变量
function Set-Environment {
    param(
        [string]$Environment,
        [string]$Version
    )
    
    Write-Info "设置环境变量..."
    
    # 基础环境变量
    $env:CI_ENVIRONMENT = $Environment
    $env:CI_VERSION = $Version
    $env:CI_BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $env:CI_COMMIT_SHA = git rev-parse HEAD
    $env:CI_COMMIT_SHORT_SHA = git rev-parse --short HEAD
    $env:CI_BRANCH = git rev-parse --abbrev-ref HEAD
    
    # Docker 相关
    if (-not $env:DOCKER_REGISTRY) { $env:DOCKER_REGISTRY = $Registry }
    if (-not $env:DOCKER_NAMESPACE) { $env:DOCKER_NAMESPACE = "taishanglaojun" }
    $env:DOCKER_IMAGE_NAME = "$($env:DOCKER_NAMESPACE)/$ServiceName"
    $env:DOCKER_IMAGE_TAG = $Version
    $env:DOCKER_IMAGE_FULL = "$($env:DOCKER_REGISTRY)/$($env:DOCKER_IMAGE_NAME):$($env:DOCKER_IMAGE_TAG)"
    
    # Kubernetes 相关
    switch ($Environment) {
        "development" { $env:K8S_NAMESPACE = "taishanglaojun-monitoring-dev" }
        "staging" { $env:K8S_NAMESPACE = "taishanglaojun-monitoring-staging" }
        "production" { $env:K8S_NAMESPACE = "taishanglaojun-monitoring" }
    }
    
    Write-Info "环境: $($env:CI_ENVIRONMENT)"
    Write-Info "版本: $($env:CI_VERSION)"
    Write-Info "镜像: $($env:DOCKER_IMAGE_FULL)"
    Write-Info "命名空间: $($env:K8S_NAMESPACE)"
}

# 运行测试
function Invoke-Tests {
    param(
        [bool]$SkipTests,
        [bool]$Verbose
    )
    
    if ($SkipTests) {
        Write-Warning "跳过测试"
        return $true
    }
    
    Write-Info "运行测试..."
    
    Push-Location $ProjectRoot
    
    try {
        # 运行测试脚本
        $testArgs = @("all")
        
        if ($Verbose) {
            $testArgs += "-Verbose"
        }
        
        $testScript = Join-Path $ScriptDir "test.ps1"
        if (Test-Path $testScript) {
            & $testScript @testArgs
            if ($LASTEXITCODE -ne 0) {
                throw "测试失败"
            }
        } else {
            throw "测试脚本不存在: $testScript"
        }
        
        Write-Success "测试完成"
        return $true
    } catch {
        Write-Error "测试失败: $_"
        return $false
    } finally {
        Pop-Location
    }
}

# 构建应用
function Build-Application {
    param(
        [bool]$SkipBuild,
        [string]$Version,
        [bool]$Verbose
    )
    
    if ($SkipBuild) {
        Write-Warning "跳过构建"
        return $true
    }
    
    Write-Info "构建应用..."
    
    Push-Location $ProjectRoot
    
    try {
        # 运行构建脚本
        $buildArgs = @("build", "-Tag", $Version)
        
        if ($Verbose) {
            $buildArgs += "-Verbose"
        }
        
        $buildScript = Join-Path $ScriptDir "docker-build.ps1"
        if (Test-Path $buildScript) {
            & $buildScript @buildArgs
            if ($LASTEXITCODE -ne 0) {
                throw "构建失败"
            }
        } else {
            throw "构建脚本不存在: $buildScript"
        }
        
        Write-Success "构建完成"
        return $true
    } catch {
        Write-Error "构建失败: $_"
        return $false
    } finally {
        Pop-Location
    }
}

# 推送镜像
function Push-Image {
    param(
        [bool]$SkipPush,
        [string]$Registry,
        [bool]$Verbose
    )
    
    if ($SkipPush) {
        Write-Warning "跳过推送"
        return $true
    }
    
    Write-Info "推送镜像..."
    
    try {
        # Docker 登录
        if ($env:DOCKER_USERNAME -and $env:DOCKER_PASSWORD) {
            $env:DOCKER_PASSWORD | docker login $Registry -u $env:DOCKER_USERNAME --password-stdin
        } elseif ($env:GITHUB_TOKEN -and $Registry -eq "ghcr.io") {
            $env:GITHUB_TOKEN | docker login ghcr.io -u $env:GITHUB_ACTOR --password-stdin
        }
        
        # 推送镜像
        docker push $env:DOCKER_IMAGE_FULL
        if ($LASTEXITCODE -ne 0) {
            throw "镜像推送失败"
        }
        
        Write-Success "镜像推送完成: $($env:DOCKER_IMAGE_FULL)"
        return $true
    } catch {
        Write-Error "推送失败: $_"
        return $false
    }
}

# 部署应用
function Deploy-Application {
    param(
        [string]$Environment,
        [string]$Version,
        [string]$Namespace,
        [bool]$Force,
        [bool]$DryRun
    )
    
    Write-Info "部署应用到 $Environment 环境..."
    
    Push-Location $ProjectRoot
    
    try {
        # 运行部署脚本
        $deployArgs = @("deploy", "-Environment", $Environment, "-Tag", $Version, "-Namespace", $Namespace, "-Wait")
        
        if ($Force) {
            $deployArgs += "-Force"
        }
        
        if ($DryRun) {
            $deployArgs += "-DryRun"
        }
        
        $deployScript = Join-Path $ScriptDir "deploy.ps1"
        if (Test-Path $deployScript) {
            & $deployScript @deployArgs
            if ($LASTEXITCODE -ne 0) {
                throw "部署失败"
            }
        } else {
            throw "部署脚本不存在: $deployScript"
        }
        
        Write-Success "部署完成"
        return $true
    } catch {
        Write-Error "部署失败: $_"
        return $false
    } finally {
        Pop-Location
    }
}

# 发送通知
function Send-Notification {
    param(
        [string]$Status,
        [string]$Environment,
        [string]$Version,
        [string]$Message
    )
    
    if (-not $env:SLACK_WEBHOOK_URL) {
        return
    }
    
    $color = switch ($Status) {
        "success" { "good" }
        "failure" { "danger" }
        "warning" { "warning" }
        default { "#439FE0" }
    }
    
    $payload = @{
        attachments = @(
            @{
                color = $color
                title = "太上老君监控系统 - $Status"
                fields = @(
                    @{
                        title = "环境"
                        value = $Environment
                        short = $true
                    },
                    @{
                        title = "版本"
                        value = $Version
                        short = $true
                    },
                    @{
                        title = "分支"
                        value = $env:CI_BRANCH
                        short = $true
                    },
                    @{
                        title = "提交"
                        value = $env:CI_COMMIT_SHORT_SHA
                        short = $true
                    }
                )
                text = $Message
                ts = [int][double]::Parse((Get-Date -UFormat %s))
            }
        )
    } | ConvertTo-Json -Depth 10
    
    try {
        Invoke-RestMethod -Uri $env:SLACK_WEBHOOK_URL -Method Post -Body $payload -ContentType "application/json"
    } catch {
        # 忽略通知失败
    }
}

# 运行完整流水线
function Invoke-Pipeline {
    param(
        [string]$Environment,
        [string]$Version,
        [bool]$SkipTests,
        [bool]$SkipBuild,
        [bool]$SkipPush,
        [bool]$Force,
        [bool]$DryRun,
        [bool]$Verbose
    )
    
    Write-Info "开始 CI/CD 流水线..."
    
    $startTime = Get-Date
    
    # 发送开始通知
    Send-Notification "info" $Environment $Version "开始部署流水线"
    
    # 运行测试
    if (-not (Invoke-Tests $SkipTests $Verbose)) {
        Send-Notification "failure" $Environment $Version "测试失败"
        exit 1
    }
    
    # 构建应用
    if (-not (Build-Application $SkipBuild $Version $Verbose)) {
        Send-Notification "failure" $Environment $Version "构建失败"
        exit 1
    }
    
    # 推送镜像
    if (-not (Push-Image $SkipPush $Registry $Verbose)) {
        Send-Notification "failure" $Environment $Version "镜像推送失败"
        exit 1
    }
    
    # 部署应用
    if (-not (Deploy-Application $Environment $Version $env:K8S_NAMESPACE $Force $DryRun)) {
        Send-Notification "failure" $Environment $Version "部署失败"
        exit 1
    }
    
    $endTime = Get-Date
    $duration = [int]($endTime - $startTime).TotalSeconds
    
    Write-Success "CI/CD 流水线完成，耗时: ${duration}s"
    Send-Notification "success" $Environment $Version "部署成功，耗时: ${duration}s"
}

# 发布版本
function Publish-Release {
    param(
        [string]$Version,
        [string]$Tag,
        [bool]$Force
    )
    
    Write-Info "发布版本: $Version"
    
    # 检查是否在主分支
    $currentBranch = git rev-parse --abbrev-ref HEAD
    
    if ($currentBranch -notin @("main", "master") -and -not $Force) {
        Write-Error "发布必须在主分支进行，当前分支: $currentBranch"
        exit 1
    }
    
    # 检查工作目录是否干净
    $status = git status --porcelain
    if ($status -and -not $Force) {
        Write-Error "工作目录不干净，请先提交或暂存更改"
        exit 1
    }
    
    # 创建标签
    if ($Tag) {
        Write-Info "创建标签: $Tag"
        git tag -a $Tag -m "Release $Version"
        git push origin $Tag
    }
    
    # 运行生产环境流水线
    Invoke-Pipeline "production" $Version $false $false $false $Force $false $true
    
    Write-Success "版本发布完成: $Version"
}

# 回滚部署
function Invoke-Rollback {
    param(
        [string]$Environment,
        [string]$Revision,
        [bool]$Force
    )
    
    Write-Info "回滚 $Environment 环境部署..."
    
    Push-Location $ProjectRoot
    
    try {
        # 运行回滚脚本
        $rollbackArgs = @("rollback", "-Environment", $Environment)
        
        if ($Revision) {
            $rollbackArgs += "-Revision", $Revision
        }
        
        if ($Force) {
            $rollbackArgs += "-Force"
        }
        
        $deployScript = Join-Path $ScriptDir "deploy.ps1"
        if (Test-Path $deployScript) {
            & $deployScript @rollbackArgs
            if ($LASTEXITCODE -ne 0) {
                throw "回滚失败"
            }
        } else {
            throw "部署脚本不存在: $deployScript"
        }
        
        Write-Success "回滚完成"
        Send-Notification "warning" $Environment "rollback" "部署已回滚"
    } catch {
        Write-Error "回滚失败: $_"
        exit 1
    } finally {
        Pop-Location
    }
}

# 查看状态
function Show-Status {
    param([string]$Environment)
    
    Write-Info "查看 $Environment 环境状态..."
    
    Push-Location $ProjectRoot
    
    try {
        # 运行状态查看脚本
        $statusArgs = @("status", "-Environment", $Environment)
        
        $deployScript = Join-Path $ScriptDir "deploy.ps1"
        if (Test-Path $deployScript) {
            & $deployScript @statusArgs
        } else {
            throw "部署脚本不存在: $deployScript"
        }
    } catch {
        Write-Error "状态查看失败: $_"
        exit 1
    } finally {
        Pop-Location
    }
}

# 清理环境
function Clear-Environment {
    Write-Info "清理环境..."
    
    # 清理 Docker 镜像
    try {
        docker system prune -f
    } catch {
        # 忽略清理失败
    }
    
    # 清理测试结果
    $testResults = Join-Path $ProjectRoot "test-results"
    if (Test-Path $testResults) {
        Remove-Item $testResults -Recurse -Force
    }
    
    Write-Success "环境清理完成"
}

# 主函数
function Main {
    # 如果启用详细输出
    if ($Verbose) {
        $VerbosePreference = "Continue"
    }
    
    # 检查依赖
    Test-Dependencies
    
    # 获取版本信息
    $Version = Get-VersionInfo $Version $Branch $Tag
    
    # 设置环境变量
    if ($Environment) {
        Set-Environment $Environment $Version
        
        # 如果没有指定命名空间，使用环境默认值
        if (-not $Namespace) {
            $Namespace = $env:K8S_NAMESPACE
        }
    }
    
    # 执行命令
    switch ($Command) {
        "build" {
            Build-Application $SkipBuild $Version $Verbose
        }
        "test" {
            Invoke-Tests $SkipTests $Verbose
        }
        "package" {
            Build-Application $false $Version $Verbose
            Push-Image $SkipPush $Registry $Verbose
        }
        "deploy" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            Deploy-Application $Environment $Version $Namespace $Force $DryRun
        }
        "pipeline" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            Invoke-Pipeline $Environment $Version $SkipTests $SkipBuild $SkipPush $Force $DryRun $Verbose
        }
        "release" {
            Publish-Release $Version $Tag $Force
        }
        "rollback" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            Invoke-Rollback $Environment $Revision $Force
        }
        "status" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            Show-Status $Environment
        }
        "clean" {
            Clear-Environment
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