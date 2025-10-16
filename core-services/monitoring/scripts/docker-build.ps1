# 太上老君监控系统 Docker 构建脚本 (PowerShell)
# 支持多平台构建和推送到容器注册表

param(
    [Parameter(Position=0)]
    [ValidateSet("build", "push", "build-push", "clean", "help")]
    [string]$Command = "help",
    
    [Alias("t")]
    [string]$Tag = "latest",
    
    [Alias("p")]
    [string]$Platform = "linux/amd64,linux/arm64",
    
    [Alias("r")]
    [string]$Registry = $env:REGISTRY ?? "ghcr.io",
    
    [Alias("n")]
    [string]$Namespace = $env:NAMESPACE ?? "taishanglaojun",
    
    [switch]$NoCache,
    [switch]$Push,
    [switch]$Latest,
    
    [Alias("v")]
    [switch]$Verbose,
    
    [Alias("h")]
    [switch]$Help
)

# 脚本配置
$ServiceName = "monitoring"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ImageName = "$Registry/$Namespace/$ServiceName"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $colorMap = @{
        "Red" = [ConsoleColor]::Red
        "Green" = [ConsoleColor]::Green
        "Yellow" = [ConsoleColor]::Yellow
        "Blue" = [ConsoleColor]::Blue
        "White" = [ConsoleColor]::White
    }
    
    Write-Host $Message -ForegroundColor $colorMap[$Color]
}

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
太上老君监控系统 Docker 构建脚本 (PowerShell)

用法: .\docker-build.ps1 [命令] [选项]

命令:
    build       构建 Docker 镜像
    push        推送镜像到注册表
    build-push  构建并推送镜像
    clean       清理本地镜像
    help        显示此帮助信息

选项:
    -Tag, -t TAG           镜像标签 (默认: latest)
    -Platform, -p PLATFORM 目标平台 (默认: linux/amd64,linux/arm64)
    -Registry, -r REGISTRY 容器注册表 (默认: ghcr.io)
    -Namespace, -n NS      命名空间 (默认: taishanglaojun)
    -NoCache               不使用构建缓存
    -Push                  构建后自动推送
    -Latest                同时标记为 latest
    -Verbose, -v           详细输出
    -Help, -h              显示帮助信息

环境变量:
    REGISTRY               容器注册表地址
    NAMESPACE              镜像命名空间

示例:
    .\docker-build.ps1 build -Tag v1.0.0
    .\docker-build.ps1 build-push -Tag v1.0.0 -Latest
    .\docker-build.ps1 push -Tag v1.0.0
    .\docker-build.ps1 clean

"@
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker 未安装或不在 PATH 中"
        exit 1
    }
    
    try {
        docker buildx version | Out-Null
    }
    catch {
        Write-Error "Docker Buildx 未安装或不可用"
        exit 1
    }
    
    Write-Success "依赖检查完成"
}

# 获取版本信息
function Get-Version {
    $version = ""
    
    try {
        # 尝试从 git tag 获取版本
        $gitTag = git describe --tags --exact-match HEAD 2>$null
        if ($LASTEXITCODE -eq 0) {
            $version = $gitTag
        }
        else {
            $gitHash = git rev-parse --short HEAD 2>$null
            if ($LASTEXITCODE -eq 0) {
                $version = "dev-$gitHash"
            }
            else {
                $version = "dev-$(Get-Date -Format 'yyyyMMddHHmmss')"
            }
        }
    }
    catch {
        $version = "dev-$(Get-Date -Format 'yyyyMMddHHmmss')"
    }
    
    return $version
}

# 构建镜像
function Build-Image {
    param(
        [string]$ImageTag,
        [string]$Platforms,
        [bool]$UseNoCache,
        [bool]$ShouldPush,
        [bool]$TagLatest,
        [bool]$VerboseOutput
    )
    
    Write-Info "开始构建 Docker 镜像..."
    Write-Info "镜像名称: $ImageName"
    Write-Info "标签: $ImageTag"
    Write-Info "平台: $Platforms"
    
    Set-Location $ProjectRoot
    
    # 构建参数
    $buildArgs = @(
        "buildx", "build",
        "--platform", $Platforms,
        "--tag", "$ImageName`:$ImageTag",
        "--file", "Dockerfile"
    )
    
    # 添加构建时参数
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $vcsRef = try { git rev-parse --short HEAD 2>$null } catch { "unknown" }
    
    $buildArgs += @(
        "--build-arg", "VERSION=$ImageTag",
        "--build-arg", "BUILD_DATE=$buildDate",
        "--build-arg", "VCS_REF=$vcsRef"
    )
    
    if ($UseNoCache) {
        $buildArgs += "--no-cache"
    }
    
    if ($ShouldPush) {
        $buildArgs += "--push"
    }
    else {
        $buildArgs += "--load"
    }
    
    if ($TagLatest) {
        $buildArgs += @("--tag", "$ImageName`:latest")
    }
    
    if ($VerboseOutput) {
        $buildArgs += @("--progress", "plain")
    }
    
    $buildArgs += "."
    
    # 执行构建
    Write-Info "执行构建命令: docker $($buildArgs -join ' ')"
    
    $process = Start-Process -FilePath "docker" -ArgumentList $buildArgs -Wait -PassThru -NoNewWindow
    
    if ($process.ExitCode -eq 0) {
        Write-Success "镜像构建成功"
        
        if (-not $ShouldPush) {
            Write-Info "镜像已保存到本地"
            docker images "$ImageName`:$ImageTag"
        }
    }
    else {
        Write-Error "镜像构建失败"
        exit 1
    }
}

# 推送镜像
function Push-Image {
    param(
        [string]$ImageTag,
        [bool]$TagLatest
    )
    
    Write-Info "推送镜像到注册表..."
    
    $process = Start-Process -FilePath "docker" -ArgumentList @("push", "$ImageName`:$ImageTag") -Wait -PassThru -NoNewWindow
    
    if ($process.ExitCode -eq 0) {
        Write-Success "镜像 $ImageName`:$ImageTag 推送成功"
    }
    else {
        Write-Error "镜像推送失败"
        exit 1
    }
    
    if ($TagLatest) {
        $process = Start-Process -FilePath "docker" -ArgumentList @("push", "$ImageName`:latest") -Wait -PassThru -NoNewWindow
        
        if ($process.ExitCode -eq 0) {
            Write-Success "镜像 $ImageName`:latest 推送成功"
        }
        else {
            Write-Error "latest 标签推送失败"
            exit 1
        }
    }
}

# 清理镜像
function Remove-Images {
    Write-Info "清理本地镜像..."
    
    # 清理悬空镜像
    try {
        docker image prune -f | Out-Null
        Write-Success "悬空镜像清理完成"
    }
    catch {
        Write-Warning "悬空镜像清理失败"
    }
    
    # 清理项目相关镜像
    try {
        $images = docker images $ImageName -q
        if ($images) {
            Write-Info "删除项目镜像..."
            docker rmi $images 2>$null | Out-Null
            Write-Success "项目镜像清理完成"
        }
        else {
            Write-Info "没有找到项目相关镜像"
        }
    }
    catch {
        Write-Warning "项目镜像清理失败"
    }
}

# 主函数
function Main {
    # 如果指定了 Help 参数或命令是 help
    if ($Help -or $Command -eq "help") {
        Show-Help
        return
    }
    
    # 启用 BuildKit
    $env:DOCKER_BUILDKIT = "1"
    
    # 检查依赖
    Test-Dependencies
    
    # 如果标签是 auto，自动获取版本
    if ($Tag -eq "auto") {
        $Tag = Get-Version
        Write-Info "自动检测版本: $Tag"
    }
    
    # 执行命令
    switch ($Command) {
        "build" {
            Build-Image -ImageTag $Tag -Platforms $Platform -UseNoCache $NoCache -ShouldPush $Push -TagLatest $Latest -VerboseOutput $Verbose
        }
        "push" {
            Push-Image -ImageTag $Tag -TagLatest $Latest
        }
        "build-push" {
            Build-Image -ImageTag $Tag -Platforms $Platform -UseNoCache $NoCache -ShouldPush $true -TagLatest $Latest -VerboseOutput $Verbose
        }
        "clean" {
            Remove-Images
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