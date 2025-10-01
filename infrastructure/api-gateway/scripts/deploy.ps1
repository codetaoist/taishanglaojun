# API Gateway PowerShell 部署脚本
# 支持开发、测试、生产环境部署

param(
    [Parameter(HelpMessage="部署环境 (dev|test|prod)")]
    [ValidateSet("dev", "test", "prod")]
    [string]$Environment = "dev",
    
    [Parameter(HelpMessage="构建Docker镜像")]
    [switch]$Build,
    
    [Parameter(HelpMessage="推送镜像到仓库")]
    [switch]$Push,
    
    [Parameter(HelpMessage="Docker镜像仓库地址")]
    [string]$Registry = "",
    
    [Parameter(HelpMessage="镜像标签")]
    [string]$Tag = "latest",
    
    [Parameter(HelpMessage="配置文件路径")]
    [string]$ConfigFile = "",
    
    [Parameter(HelpMessage="Docker Compose文件")]
    [string]$ComposeFile = "docker-compose.yml",
    
    [Parameter(HelpMessage="显示帮助信息")]
    [switch]$Help
)

# 显示帮助信息
function Show-Help {
    Write-Host "API Gateway PowerShell 部署脚本" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "用法: .\deploy.ps1 [参数]" -ForegroundColor White
    Write-Host ""
    Write-Host "参数:" -ForegroundColor Yellow
    Write-Host "  -Environment <env>     部署环境 (dev|test|prod) [默认: dev]" -ForegroundColor White
    Write-Host "  -Build                 构建Docker镜像" -ForegroundColor White
    Write-Host "  -Push                  推送镜像到仓库" -ForegroundColor White
    Write-Host "  -Registry <registry>   Docker镜像仓库地址" -ForegroundColor White
    Write-Host "  -Tag <tag>             镜像标签 [默认: latest]" -ForegroundColor White
    Write-Host "  -ConfigFile <config>   配置文件路径" -ForegroundColor White
    Write-Host "  -ComposeFile <file>    Docker Compose文件 [默认: docker-compose.yml]" -ForegroundColor White
    Write-Host "  -Help                  显示帮助信息" -ForegroundColor White
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Yellow
    Write-Host "  .\deploy.ps1 -Environment dev -Build" -ForegroundColor White
    Write-Host "  .\deploy.ps1 -Environment prod -Build -Push -Registry registry.com -Tag v1.0.0" -ForegroundColor White
    Write-Host "  .\deploy.ps1 -Environment test -ConfigFile configs\test.yaml" -ForegroundColor White
}

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查依赖..."
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker 未安装或不在PATH中"
        exit 1
    }
    
    if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
        Write-Error "Docker Compose 未安装或不在PATH中"
        exit 1
    }
    
    Write-Success "依赖检查完成"
}

# 设置环境变量
function Set-EnvironmentVariables {
    Write-Info "设置环境变量..."
    
    $env:ENVIRONMENT = $Environment
    $env:TAG = $Tag
    
    if ($Registry) {
        $env:REGISTRY = $Registry
    }
    
    if ($ConfigFile) {
        $env:CONFIG_FILE = $ConfigFile
    }
    
    # 根据环境设置不同的配置
    switch ($Environment) {
        "dev" {
            $env:GIN_MODE = "debug"
            $env:LOG_LEVEL = "debug"
        }
        "test" {
            $env:GIN_MODE = "test"
            $env:LOG_LEVEL = "info"
        }
        "prod" {
            $env:GIN_MODE = "release"
            $env:LOG_LEVEL = "warn"
        }
    }
    
    Write-Success "环境变量设置完成"
}

# 构建镜像
function Build-Image {
    if ($Build) {
        Write-Info "构建Docker镜像..."
        
        $imageName = "api-gateway"
        if ($Registry) {
            $imageName = "$Registry/api-gateway"
        }
        
        docker build -t "${imageName}:$Tag" .
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "镜像构建失败"
            exit 1
        }
        
        # 如果是latest标签，也打上环境标签
        if ($Tag -eq "latest") {
            docker tag "${imageName}:$Tag" "${imageName}:$Environment"
        }
        
        Write-Success "镜像构建完成: ${imageName}:$Tag"
    }
}

# 推送镜像
function Push-Image {
    if ($Push) {
        if (-not $Registry) {
            Write-Error "推送镜像需要指定仓库地址 (-Registry)"
            exit 1
        }
        
        Write-Info "推送镜像到仓库..."
        
        $imageName = "$Registry/api-gateway"
        docker push "${imageName}:$Tag"
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "镜像推送失败"
            exit 1
        }
        
        if ($Tag -eq "latest") {
            docker push "${imageName}:$Environment"
        }
        
        Write-Success "镜像推送完成"
    }
}

# 创建必要的目录
function New-Directories {
    Write-Info "创建必要的目录..."
    
    $directories = @("logs", "data", "configs", "logs\$Environment")
    
    foreach ($dir in $directories) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }
    
    Write-Success "目录创建完成"
}

# 验证配置文件
function Test-Config {
    Write-Info "验证配置文件..."
    
    $configPath = "configs\gateway.yaml"
    if ($ConfigFile) {
        $configPath = $ConfigFile
    }
    
    if (-not (Test-Path $configPath)) {
        Write-Error "配置文件不存在: $configPath"
        exit 1
    }
    
    # 可以添加配置文件语法验证
    # go run cmd\main.go -config $configPath -validate
    
    Write-Success "配置文件验证完成"
}

# 部署服务
function Deploy-Services {
    Write-Info "部署服务..."
    
    # 停止现有服务
    docker-compose -f $ComposeFile down
    
    # 根据环境选择不同的compose配置
    switch ($Environment) {
        "dev" {
            docker-compose -f $ComposeFile up -d redis prometheus grafana
            docker-compose -f $ComposeFile up -d api-gateway
        }
        "test" {
            docker-compose -f $ComposeFile up -d redis
            docker-compose -f $ComposeFile up -d api-gateway-test
        }
        "prod" {
            docker-compose -f $ComposeFile up -d redis prometheus grafana jaeger
            docker-compose -f $ComposeFile up -d api-gateway
        }
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "服务部署失败"
        exit 1
    }
    
    Write-Success "服务部署完成"
}

# 健康检查
function Test-Health {
    Write-Info "执行健康检查..."
    
    $maxAttempts = 30
    $attempt = 1
    
    while ($attempt -le $maxAttempts) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                Write-Success "健康检查通过"
                return $true
            }
        }
        catch {
            # 忽略错误，继续尝试
        }
        
        Write-Info "等待服务启动... ($attempt/$maxAttempts)"
        Start-Sleep -Seconds 2
        $attempt++
    }
    
    Write-Error "健康检查失败"
    return $false
}

# 显示部署信息
function Show-DeploymentInfo {
    Write-Success "部署完成!"
    Write-Host ""
    Write-Host "服务信息:" -ForegroundColor Yellow
    Write-Host "  API Gateway: http://localhost:8080" -ForegroundColor White
    Write-Host "  健康检查:   http://localhost:8080/health" -ForegroundColor White
    Write-Host "  就绪检查:   http://localhost:8080/ready" -ForegroundColor White
    Write-Host "  监控指标:   http://localhost:9090/metrics" -ForegroundColor White
    
    if ($Environment -ne "test") {
        Write-Host "  Prometheus: http://localhost:9091" -ForegroundColor White
        Write-Host "  Grafana:    http://localhost:3000 (admin/admin123)" -ForegroundColor White
    }
    
    if ($Environment -eq "prod") {
        Write-Host "  Jaeger:     http://localhost:16686" -ForegroundColor White
    }
    
    Write-Host ""
    Write-Host "管理命令:" -ForegroundColor Yellow
    Write-Host "  查看日志: docker-compose logs -f api-gateway" -ForegroundColor White
    Write-Host "  停止服务: docker-compose down" -ForegroundColor White
    Write-Host "  重启服务: docker-compose restart api-gateway" -ForegroundColor White
}

# 主函数
function Main {
    if ($Help) {
        Show-Help
        return
    }
    
    Write-Info "开始部署 API Gateway..."
    Write-Info "部署环境: $Environment"
    
    try {
        Test-Dependencies
        Set-EnvironmentVariables
        New-Directories
        Test-Config
        Build-Image
        Push-Image
        Deploy-Services
        
        if (Test-Health) {
            Show-DeploymentInfo
        } else {
            Write-Error "部署失败"
            exit 1
        }
    }
    catch {
        Write-Error "部署过程中发生错误: $($_.Exception.Message)"
        exit 1
    }
}

# 执行主函数
Main