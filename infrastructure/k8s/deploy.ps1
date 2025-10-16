# 太上老君AI平台 - Kubernetes部署脚本 (PowerShell版本)
# 版本: 1.0.0
# 创建时间: 2024-01-01

param(
    [Parameter(Position=0)]
    [ValidateSet("deploy", "cleanup", "verify", "info")]
    [string]$Action = "deploy"
)

# 颜色定义
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

# 日志函数
function Write-LogInfo {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Blue
}

function Write-LogSuccess {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Green
}

function Write-LogWarning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-LogError {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
}

# 检查kubectl是否可用
function Test-Kubectl {
    try {
        $null = Get-Command kubectl -ErrorAction Stop
        $null = kubectl cluster-info 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "无法连接到Kubernetes集群"
        }
        Write-LogSuccess "Kubernetes集群连接正常"
        return $true
    }
    catch {
        Write-LogError "kubectl 未安装或无法连接到集群: $($_.Exception.Message)"
        return $false
    }
}

# 创建命名空间
function New-Namespace {
    Write-LogInfo "创建命名空间..."
    kubectl apply -f namespace.yaml
    if ($LASTEXITCODE -eq 0) {
        Write-LogSuccess "命名空间创建完成"
    } else {
        Write-LogError "命名空间创建失败"
        exit 1
    }
}

# 创建ConfigMaps和Secrets
function New-Configs {
    Write-LogInfo "创建配置文件..."
    kubectl apply -f configmap.yaml
    kubectl apply -f secrets.yaml
    if ($LASTEXITCODE -eq 0) {
        Write-LogSuccess "配置文件创建完成"
    } else {
        Write-LogError "配置文件创建失败"
        exit 1
    }
}

# 部署数据库
function Deploy-Database {
    Write-LogInfo "部署PostgreSQL数据库..."
    kubectl apply -f postgres.yaml
    
    if ($LASTEXITCODE -eq 0) {
        Write-LogInfo "等待PostgreSQL启动..."
        kubectl wait --for=condition=ready pod -l app=postgres -n taishanglaojun --timeout=300s
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "PostgreSQL部署完成"
        } else {
            Write-LogError "PostgreSQL启动超时"
            exit 1
        }
    } else {
        Write-LogError "PostgreSQL部署失败"
        exit 1
    }
}

# 部署Redis
function Deploy-Redis {
    Write-LogInfo "部署Redis缓存..."
    kubectl apply -f redis.yaml
    
    if ($LASTEXITCODE -eq 0) {
        Write-LogInfo "等待Redis启动..."
        kubectl wait --for=condition=ready pod -l app=redis -n taishanglaojun --timeout=300s
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "Redis部署完成"
        } else {
            Write-LogError "Redis启动超时"
            exit 1
        }
    } else {
        Write-LogError "Redis部署失败"
        exit 1
    }
}

# 部署核心服务
function Deploy-CoreServices {
    Write-LogInfo "部署核心服务..."
    kubectl apply -f core-services.yaml
    
    if ($LASTEXITCODE -eq 0) {
        Write-LogInfo "等待核心服务启动..."
        kubectl wait --for=condition=ready pod -l app=core-services -n taishanglaojun --timeout=300s
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "核心服务部署完成"
        } else {
            Write-LogError "核心服务启动超时"
            exit 1
        }
    } else {
        Write-LogError "核心服务部署失败"
        exit 1
    }
}

# 部署前端
function Deploy-Frontend {
    Write-LogInfo "部署前端应用..."
    kubectl apply -f frontend.yaml
    
    if ($LASTEXITCODE -eq 0) {
        Write-LogInfo "等待前端应用启动..."
        kubectl wait --for=condition=ready pod -l app=frontend -n taishanglaojun --timeout=300s
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "前端应用部署完成"
        } else {
            Write-LogError "前端应用启动超时"
            exit 1
        }
    } else {
        Write-LogError "前端应用部署失败"
        exit 1
    }
}

# 部署监控
function Deploy-Monitoring {
    Write-LogInfo "部署监控系统..."
    kubectl apply -f monitoring.yaml
    
    if ($LASTEXITCODE -eq 0) {
        Write-LogInfo "等待监控系统启动..."
        kubectl wait --for=condition=ready pod -l app=prometheus -n taishanglaojun --timeout=300s
        kubectl wait --for=condition=ready pod -l app=grafana -n taishanglaojun --timeout=300s
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "监控系统部署完成"
        } else {
            Write-LogError "监控系统启动超时"
            exit 1
        }
    } else {
        Write-LogError "监控系统部署失败"
        exit 1
    }
}

# 验证部署
function Test-Deployment {
    Write-LogInfo "验证部署状态..."
    
    Write-Host ""
    Write-LogInfo "Pod状态:"
    kubectl get pods -n taishanglaojun
    
    Write-Host ""
    Write-LogInfo "Service状态:"
    kubectl get services -n taishanglaojun
    
    Write-Host ""
    Write-LogInfo "Ingress状态:"
    kubectl get ingress -n taishanglaojun
    
    Write-Host ""
    Write-LogInfo "PVC状态:"
    kubectl get pvc -n taishanglaojun
    
    # 检查所有Pod是否运行正常
    $failedPods = kubectl get pods -n taishanglaojun --field-selector=status.phase!=Running --no-headers 2>$null
    if ([string]::IsNullOrEmpty($failedPods)) {
        Write-LogSuccess "所有Pod运行正常"
    } else {
        $failedCount = ($failedPods | Measure-Object -Line).Lines
        Write-LogWarning "有 $failedCount 个Pod未正常运行"
        kubectl get pods -n taishanglaojun --field-selector=status.phase!=Running
    }
}

# 显示访问信息
function Show-AccessInfo {
    Write-Host ""
    Write-LogInfo "=== 访问信息 ==="
    
    # 获取Ingress信息
    $ingressIp = kubectl get ingress frontend-ingress -n taishanglaojun -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>$null
    
    if (![string]::IsNullOrEmpty($ingressIp) -and $ingressIp -ne "pending") {
        Write-Host "前端应用: https://taishanglaojun.com" -ForegroundColor $Colors.Green
        Write-Host "API接口: https://taishanglaojun.com/api" -ForegroundColor $Colors.Green
    } else {
        Write-LogInfo "使用端口转发访问应用:"
        Write-Host "前端应用: kubectl port-forward -n taishanglaojun svc/frontend-service 8080:80" -ForegroundColor $Colors.Yellow
        Write-Host "核心服务: kubectl port-forward -n taishanglaojun svc/core-services 8081:8080" -ForegroundColor $Colors.Yellow
        Write-Host "Grafana: kubectl port-forward -n taishanglaojun svc/grafana-service 3000:3000" -ForegroundColor $Colors.Yellow
        Write-Host "Prometheus: kubectl port-forward -n taishanglaojun svc/prometheus-service 9090:9090" -ForegroundColor $Colors.Yellow
    }
    
    Write-Host ""
    Write-LogInfo "默认登录信息:"
    Write-Host "Grafana - 用户名: admin, 密码: 查看secret获取" -ForegroundColor $Colors.White
    Write-Host "数据库 - 用户名: postgres, 密码: 查看secret获取" -ForegroundColor $Colors.White
    
    Write-Host ""
    Write-LogInfo "获取密码命令:"
    Write-Host "kubectl get secret taishanglaojun-secrets -n taishanglaojun -o jsonpath='{.data.grafana-admin-password}' | base64 -d" -ForegroundColor $Colors.Yellow
}

# 清理部署
function Remove-Deployment {
    Write-LogWarning "清理所有部署资源..."
    kubectl delete namespace taishanglaojun --ignore-not-found=true
    if ($LASTEXITCODE -eq 0) {
        Write-LogSuccess "清理完成"
    } else {
        Write-LogError "清理失败"
    }
}

# 主函数
function Main {
    Write-Host "太上老君AI平台 - Kubernetes部署脚本" -ForegroundColor $Colors.Blue
    Write-Host "========================================" -ForegroundColor $Colors.Blue
    
    switch ($Action) {
        "deploy" {
            if (!(Test-Kubectl)) { exit 1 }
            New-Namespace
            New-Configs
            Deploy-Database
            Deploy-Redis
            Deploy-CoreServices
            Deploy-Frontend
            Deploy-Monitoring
            Test-Deployment
            Show-AccessInfo
            Write-LogSuccess "部署完成！"
        }
        "cleanup" {
            Remove-Deployment
        }
        "verify" {
            Test-Deployment
        }
        "info" {
            Show-AccessInfo
        }
        default {
            Write-Host "用法: .\deploy.ps1 [deploy|cleanup|verify|info]" -ForegroundColor $Colors.White
            Write-Host "  deploy  - 部署所有服务 (默认)" -ForegroundColor $Colors.White
            Write-Host "  cleanup - 清理所有资源" -ForegroundColor $Colors.White
            Write-Host "  verify  - 验证部署状态" -ForegroundColor $Colors.White
            Write-Host "  info    - 显示访问信息" -ForegroundColor $Colors.White
            exit 1
        }
    }
}

# 执行主函数
Main