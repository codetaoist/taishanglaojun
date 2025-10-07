# 太上老君监控系统部署脚本 (PowerShell)
# 支持 Helm 和 Kustomize 两种部署方式

param(
    [Parameter(Position = 0)]
    [ValidateSet("deploy", "upgrade", "rollback", "status", "logs", "delete", "help")]
    [string]$Command = "help",
    
    [Parameter()]
    [Alias("e")]
    [ValidateSet("development", "staging", "production")]
    [string]$Environment,
    
    [Parameter()]
    [Alias("m")]
    [ValidateSet("helm", "kustomize")]
    [string]$Method = "helm",
    
    [Parameter()]
    [Alias("n")]
    [string]$Namespace,
    
    [Parameter()]
    [Alias("t")]
    [string]$Tag,
    
    [Parameter()]
    [Alias("f")]
    [string]$ValuesFile,
    
    [Parameter()]
    [Alias("w")]
    [switch]$Wait,
    
    [Parameter()]
    [string]$Timeout = "600s",
    
    [Parameter()]
    [switch]$DryRun,
    
    [Parameter()]
    [switch]$Force,
    
    [Parameter()]
    [switch]$Follow,
    
    [Parameter()]
    [string]$Revision,
    
    [Parameter()]
    [Alias("v")]
    [switch]$Verbose,
    
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
太上老君监控系统部署脚本 (PowerShell)

用法: .\deploy.ps1 [命令] [选项]

命令:
    deploy      部署应用
    upgrade     升级应用
    rollback    回滚应用
    status      查看部署状态
    logs        查看应用日志
    delete      删除部署
    help        显示此帮助信息

选项:
    -Environment, -e ENV    部署环境 (development|staging|production)
    -Method, -m METHOD      部署方法 (helm|kustomize) [默认: helm]
    -Namespace, -n NS       Kubernetes 命名空间
    -Tag, -t TAG           镜像标签
    -ValuesFile, -f FILE   Helm values 文件路径
    -Wait, -w              等待部署完成
    -Timeout TIMEOUT       部署超时时间 (默认: 600s)
    -DryRun                仅显示将要执行的操作
    -Force                 强制部署
    -Follow                跟踪日志输出
    -Revision REV          回滚到指定版本
    -Verbose, -v           详细输出
    -Help, -h              显示帮助信息

环境变量:
    KUBECONFIG             Kubernetes 配置文件路径
    HELM_NAMESPACE         Helm 默认命名空间
    KUSTOMIZE_NAMESPACE    Kustomize 默认命名空间

示例:
    .\deploy.ps1 deploy -Environment development
    .\deploy.ps1 deploy -Environment production -Tag v1.0.0 -Wait
    .\deploy.ps1 upgrade -Environment staging -Method kustomize
    .\deploy.ps1 rollback -Environment production -Force
    .\deploy.ps1 status -Environment development
    .\deploy.ps1 logs -Environment production -Follow

"@
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    # 检查 kubectl
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
    
    Write-Success "依赖检查完成"
}

# 检查 Helm
function Test-Helm {
    if (-not (Get-Command helm -ErrorAction SilentlyContinue)) {
        Write-Error "Helm 未安装或不在 PATH 中"
        exit 1
    }
    
    # 检查 Helm 版本
    $helmVersion = helm version --short --client
    Write-Info "Helm 版本: $helmVersion"
}

# 检查 Kustomize
function Test-Kustomize {
    if (-not (Get-Command kustomize -ErrorAction SilentlyContinue)) {
        Write-Error "Kustomize 未安装或不在 PATH 中"
        exit 1
    }
    
    # 检查 Kustomize 版本
    $kustomizeVersion = kustomize version --short
    Write-Info "Kustomize 版本: $kustomizeVersion"
}

# 获取命名空间
function Get-Namespace {
    param([string]$Environment)
    
    switch ($Environment) {
        "development" { return "taishanglaojun-monitoring-dev" }
        "staging" { return "taishanglaojun-monitoring-staging" }
        "production" { return "taishanglaojun-monitoring" }
        default {
            Write-Error "未知环境: $Environment"
            exit 1
        }
    }
}

# 创建命名空间
function New-Namespace {
    param([string]$Namespace)
    
    $existingNamespace = kubectl get namespace $Namespace -o name 2>$null
    if (-not $existingNamespace) {
        Write-Info "创建命名空间: $Namespace"
        kubectl create namespace $Namespace
        
        # 添加标签
        kubectl label namespace $Namespace `
            name=$Namespace `
            app.kubernetes.io/name=monitoring `
            app.kubernetes.io/part-of=taishanglaojun `
            --overwrite
    }
    else {
        Write-Info "命名空间已存在: $Namespace"
    }
}

# Helm 部署
function Invoke-HelmDeploy {
    param(
        [string]$Environment,
        [string]$Namespace,
        [string]$Tag,
        [string]$ValuesFile,
        [bool]$Wait,
        [string]$Timeout,
        [bool]$DryRun,
        [bool]$Force
    )
    
    Test-Helm
    
    $releaseName = "monitoring-$Environment"
    $chartPath = Join-Path $ProjectRoot "helm\monitoring"
    
    # 默认 values 文件
    if (-not $ValuesFile) {
        $ValuesFile = Join-Path $ProjectRoot "helm\values-$Environment.yaml"
    }
    
    # 检查 values 文件
    if (-not (Test-Path $ValuesFile)) {
        Write-Error "Values 文件不存在: $ValuesFile"
        exit 1
    }
    
    Write-Info "使用 Helm 部署..."
    Write-Info "Release: $releaseName"
    Write-Info "Chart: $chartPath"
    Write-Info "Values: $ValuesFile"
    Write-Info "Namespace: $Namespace"
    
    # 构建 Helm 命令参数
    $helmArgs = @(
        "upgrade", "--install",
        $releaseName,
        $chartPath,
        "--namespace", $Namespace,
        "--values", $ValuesFile
    )
    
    if ($Tag) {
        $helmArgs += @("--set", "image.tag=$Tag")
    }
    
    if ($Wait) {
        $helmArgs += @("--wait", "--timeout", $Timeout)
    }
    
    if ($DryRun) {
        $helmArgs += @("--dry-run")
    }
    
    if ($Force) {
        $helmArgs += @("--force")
    }
    
    # 执行部署
    Write-Info "执行命令: helm $($helmArgs -join ' ')"
    
    & helm @helmArgs
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Helm 部署成功"
    }
    else {
        Write-Error "Helm 部署失败"
        exit 1
    }
}

# Kustomize 部署
function Invoke-KustomizeDeploy {
    param(
        [string]$Environment,
        [string]$Namespace,
        [string]$Tag,
        [bool]$DryRun
    )
    
    Test-Kustomize
    
    $overlayPath = Join-Path $ProjectRoot "k8s\overlays\$Environment"
    
    # 检查 overlay 目录
    if (-not (Test-Path $overlayPath)) {
        Write-Error "Overlay 目录不存在: $overlayPath"
        exit 1
    }
    
    Write-Info "使用 Kustomize 部署..."
    Write-Info "Overlay: $overlayPath"
    Write-Info "Namespace: $Namespace"
    
    # 临时修改镜像标签
    if ($Tag) {
        Write-Info "设置镜像标签: $Tag"
        Push-Location $overlayPath
        try {
            kustomize edit set image "ghcr.io/taishanglaojun/monitoring:$Tag"
        }
        finally {
            Pop-Location
        }
    }
    
    # 构建和应用
    if ($DryRun) {
        Write-Info "执行 dry-run..."
        kustomize build $overlayPath
    }
    else {
        Write-Info "应用配置..."
        kustomize build $overlayPath | kubectl apply -f -
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Kustomize 部署成功"
        }
        else {
            Write-Error "Kustomize 部署失败"
            exit 1
        }
    }
}

# 查看部署状态
function Show-Status {
    param(
        [string]$Environment,
        [string]$Namespace,
        [string]$Method
    )
    
    Write-Info "查看部署状态..."
    
    if ($Method -eq "helm") {
        $releaseName = "monitoring-$Environment"
        
        Write-Info "Helm Release 状态:"
        helm status $releaseName --namespace $Namespace
        
        Write-Info "Helm Release 历史:"
        helm history $releaseName --namespace $Namespace
    }
    
    Write-Info "Pod 状态:"
    kubectl get pods -n $Namespace -l app.kubernetes.io/name=monitoring
    
    Write-Info "Service 状态:"
    kubectl get services -n $Namespace -l app.kubernetes.io/name=monitoring
    
    Write-Info "Ingress 状态:"
    kubectl get ingress -n $Namespace -l app.kubernetes.io/name=monitoring
    
    Write-Info "HPA 状态:"
    kubectl get hpa -n $Namespace -l app.kubernetes.io/name=monitoring 2>$null
}

# 查看日志
function Show-Logs {
    param(
        [string]$Namespace,
        [bool]$Follow
    )
    
    Write-Info "查看应用日志..."
    
    $kubectlArgs = @(
        "logs",
        "-n", $Namespace,
        "-l", "app.kubernetes.io/name=monitoring",
        "--tail=100"
    )
    
    if ($Follow) {
        $kubectlArgs += @("-f")
    }
    
    & kubectl @kubectlArgs
}

# 回滚部署
function Invoke-Rollback {
    param(
        [string]$Environment,
        [string]$Namespace,
        [string]$Method,
        [string]$Revision,
        [bool]$Force
    )
    
    if ($Method -eq "helm") {
        Test-Helm
        
        $releaseName = "monitoring-$Environment"
        
        Write-Info "回滚 Helm Release: $releaseName"
        
        $helmArgs = @(
            "rollback",
            $releaseName,
            "--namespace", $Namespace
        )
        
        if ($Revision) {
            $helmArgs += $Revision
        }
        
        if ($Force) {
            $helmArgs += @("--force")
        }
        
        & helm @helmArgs
        if ($LASTEXITCODE -eq 0) {
            Write-Success "回滚成功"
        }
        else {
            Write-Error "回滚失败"
            exit 1
        }
    }
    else {
        Write-Error "Kustomize 不支持自动回滚，请手动操作"
        exit 1
    }
}

# 删除部署
function Remove-Deployment {
    param(
        [string]$Environment,
        [string]$Namespace,
        [string]$Method,
        [bool]$Force
    )
    
    if ($Method -eq "helm") {
        Test-Helm
        
        $releaseName = "monitoring-$Environment"
        
        Write-Warning "删除 Helm Release: $releaseName"
        
        if ($Force -or (Read-Host "确认删除? (y/N)") -eq "y") {
            helm uninstall $releaseName --namespace $Namespace
            if ($LASTEXITCODE -eq 0) {
                Write-Success "删除成功"
            }
            else {
                Write-Error "删除失败"
                exit 1
            }
        }
        else {
            Write-Info "取消删除"
        }
    }
    else {
        Write-Warning "删除 Kustomize 部署"
        
        $overlayPath = Join-Path $ProjectRoot "k8s\overlays\$Environment"
        
        if ($Force -or (Read-Host "确认删除? (y/N)") -eq "y") {
            kustomize build $overlayPath | kubectl delete -f -
            if ($LASTEXITCODE -eq 0) {
                Write-Success "删除成功"
            }
            else {
                Write-Error "删除失败"
                exit 1
            }
        }
        else {
            Write-Info "取消删除"
        }
    }
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
    
    # 如果没有指定命名空间，根据环境自动获取
    if (-not $Namespace -and $Environment) {
        $Namespace = Get-Namespace $Environment
    }
    
    # 执行命令
    switch ($Command) {
        { $_ -in @("deploy", "upgrade") } {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            
            New-Namespace $Namespace
            
            if ($Method -eq "helm") {
                Invoke-HelmDeploy $Environment $Namespace $Tag $ValuesFile $Wait.IsPresent $Timeout $DryRun.IsPresent $Force.IsPresent
            }
            elseif ($Method -eq "kustomize") {
                Invoke-KustomizeDeploy $Environment $Namespace $Tag $DryRun.IsPresent
            }
            else {
                Write-Error "未知部署方法: $Method"
                exit 1
            }
        }
        "rollback" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            
            Invoke-Rollback $Environment $Namespace $Method $Revision $Force.IsPresent
        }
        "status" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            
            Show-Status $Environment $Namespace $Method
        }
        "logs" {
            if (-not $Namespace) {
                Write-Error "必须指定命名空间 (-Namespace)"
                exit 1
            }
            
            Show-Logs $Namespace $Follow.IsPresent
        }
        "delete" {
            if (-not $Environment) {
                Write-Error "必须指定环境 (-Environment)"
                exit 1
            }
            
            Remove-Deployment $Environment $Namespace $Method $Force.IsPresent
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