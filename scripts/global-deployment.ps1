# 太上老君AI平台全球化部署脚本
# Global Deployment Script for Taishang Laojun AI Platform

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("dev", "staging", "prod")]
    [string]$Environment,
    
    [Parameter(Mandatory=$true)]
    [ValidateSet("ap-east-1", "eu-west-1", "us-east-1", "ap-southeast-1", "eu-central-1", "us-west-2")]
    [string]$Region,
    
    [Parameter(Mandatory=$false)]
    [string]$ConfigFile = "deployment-config.yaml",
    
    [Parameter(Mandatory=$false)]
    [switch]$DryRun,
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipValidation,
    
    [Parameter(Mandatory=$false)]
    [switch]$EnableMonitoring = $true,
    
    [Parameter(Mandatory=$false)]
    [switch]$EnableCompliance = $true
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 日志函数
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    Write-Host $logMessage
    Add-Content -Path "deployment-$Environment-$Region.log" -Value $logMessage
}

# 验证先决条件
function Test-Prerequisites {
    Write-Log "检查部署先决条件..." "INFO"
    
    # 检查必需的工具
    $requiredTools = @("kubectl", "helm", "docker", "terraform")
    foreach ($tool in $requiredTools) {
        if (!(Get-Command $tool -ErrorAction SilentlyContinue)) {
            Write-Log "缺少必需工具: $tool" "ERROR"
            throw "缺少必需工具: $tool"
        }
    }
    
    # 检查配置文件
    if (!(Test-Path $ConfigFile)) {
        Write-Log "配置文件不存在: $ConfigFile" "ERROR"
        throw "配置文件不存在: $ConfigFile"
    }
    
    # 检查环境变量
    $requiredEnvVars = @("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "KUBECONFIG")
    foreach ($envVar in $requiredEnvVars) {
        if (!(Get-ChildItem Env:$envVar -ErrorAction SilentlyContinue)) {
            Write-Log "缺少环境变量: $envVar" "ERROR"
            throw "缺少环境变量: $envVar"
        }
    }
    
    Write-Log "先决条件检查完成" "INFO"
}

# 加载配置
function Get-DeploymentConfig {
    Write-Log "加载部署配置..." "INFO"
    
    $config = Get-Content $ConfigFile | ConvertFrom-Yaml
    
    # 验证配置结构
    $requiredSections = @("regions", "services", "database", "monitoring", "compliance")
    foreach ($section in $requiredSections) {
        if (!$config.$section) {
            Write-Log "配置文件缺少必需部分: $section" "ERROR"
            throw "配置文件缺少必需部分: $section"
        }
    }
    
    return $config
}

# 设置基础设施
function Deploy-Infrastructure {
    param($config)
    
    Write-Log "部署基础设施..." "INFO"
    
    # 切换到基础设施目录
    Push-Location "infrastructure"
    
    try {
        # 初始化Terraform
        Write-Log "初始化Terraform..." "INFO"
        terraform init -backend-config="region=$Region"
        
        # 创建Terraform变量文件
        $tfVars = @"
environment = "$Environment"
region = "$Region"
project_name = "taishang-laojun"
enable_monitoring = $($EnableMonitoring.ToString().ToLower())
enable_compliance = $($EnableCompliance.ToString().ToLower())
"@
        $tfVars | Out-File -FilePath "terraform.tfvars" -Encoding UTF8
        
        if ($DryRun) {
            Write-Log "执行Terraform计划 (Dry Run)..." "INFO"
            terraform plan -var-file="terraform.tfvars"
        } else {
            Write-Log "应用Terraform配置..." "INFO"
            terraform apply -var-file="terraform.tfvars" -auto-approve
        }
        
        # 获取输出
        $outputs = terraform output -json | ConvertFrom-Json
        Write-Log "基础设施部署完成" "INFO"
        
        return $outputs
    }
    finally {
        Pop-Location
    }
}

# 部署数据库
function Deploy-Database {
    param($config, $outputs)
    
    Write-Log "部署数据库..." "INFO"
    
    # 部署PostgreSQL主库
    Write-Log "部署PostgreSQL主库..." "INFO"
    kubectl apply -f "infrastructure/database/postgresql-primary.yaml"
    
    # 部署PostgreSQL只读副本
    if ($config.database.replicas -and $config.database.replicas.Count -gt 0) {
        Write-Log "部署PostgreSQL只读副本..." "INFO"
        foreach ($replica in $config.database.replicas) {
            $replicaConfig = Get-Content "infrastructure/database/postgresql-replica.yaml" | 
                ForEach-Object { $_ -replace "{{REPLICA_REGION}}", $replica.region }
            $replicaConfig | kubectl apply -f -
        }
    }
    
    # 部署Redis集群
    Write-Log "部署Redis集群..." "INFO"
    kubectl apply -f "infrastructure/database/redis-cluster.yaml"
    
    # 等待数据库就绪
    Write-Log "等待数据库就绪..." "INFO"
    kubectl wait --for=condition=ready pod -l app=postgresql-primary --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis-cluster --timeout=300s
    
    Write-Log "数据库部署完成" "INFO"
}

# 部署核心服务
function Deploy-CoreServices {
    param($config)
    
    Write-Log "部署核心服务..." "INFO"
    
    # 构建和推送Docker镜像
    Write-Log "构建Docker镜像..." "INFO"
    Push-Location "core-services"
    
    try {
        $imageTag = "$Environment-$Region-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
        
        # 构建镜像
        docker build -t "taishang-laojun/core-services:$imageTag" .
        
        # 推送到镜像仓库
        if (!$DryRun) {
            docker push "taishang-laojun/core-services:$imageTag"
        }
        
        # 更新Kubernetes部署
        $deploymentConfig = Get-Content "k8s/deployment.yaml" | 
            ForEach-Object { $_ -replace "{{IMAGE_TAG}}", $imageTag }
        $deploymentConfig | kubectl apply -f -
        
        # 等待部署完成
        kubectl rollout status deployment/core-services --timeout=600s
        
        Write-Log "核心服务部署完成" "INFO"
    }
    finally {
        Pop-Location
    }
}

# 部署前端应用
function Deploy-Frontend {
    param($config)
    
    Write-Log "部署前端应用..." "INFO"
    
    Push-Location "frontend/web-app"
    
    try {
        # 安装依赖
        Write-Log "安装前端依赖..." "INFO"
        npm ci
        
        # 构建应用
        Write-Log "构建前端应用..." "INFO"
        $env:VITE_API_BASE_URL = $config.frontend.apiBaseUrl
        $env:VITE_REGION = $Region
        $env:VITE_ENVIRONMENT = $Environment
        npm run build
        
        # 部署到CDN
        if (!$DryRun) {
            Write-Log "部署到CDN..." "INFO"
            aws s3 sync dist/ "s3://$($config.frontend.s3Bucket)/" --delete
            aws cloudfront create-invalidation --distribution-id $config.frontend.cloudfrontId --paths "/*"
        }
        
        Write-Log "前端应用部署完成" "INFO"
    }
    finally {
        Pop-Location
    }
}

# 配置本地化
function Configure-Localization {
    param($config)
    
    Write-Log "配置本地化设置..." "INFO"
    
    # 部署本地化配置
    $localizationConfig = @{
        region = $Region
        supportedLanguages = $config.localization.supportedLanguages
        defaultLanguage = $config.localization.defaultLanguage
        timezones = $config.localization.timezones
        currencies = $config.localization.currencies
        culturalSettings = $config.localization.culturalSettings
    }
    
    $configJson = $localizationConfig | ConvertTo-Json -Depth 10
    kubectl create configmap localization-config --from-literal=config.json="$configJson" --dry-run=client -o yaml | kubectl apply -f -
    
    # 重启相关服务以应用配置
    kubectl rollout restart deployment/localization-service
    
    Write-Log "本地化配置完成" "INFO"
}

# 配置合规性
function Configure-Compliance {
    param($config)
    
    if (!$EnableCompliance) {
        Write-Log "跳过合规性配置" "INFO"
        return
    }
    
    Write-Log "配置合规性设置..." "INFO"
    
    # 根据区域配置合规性规则
    $complianceRules = switch ($Region) {
        { $_ -like "eu-*" } { $config.compliance.gdpr }
        { $_ -like "us-*" } { $config.compliance.ccpa }
        { $_ -like "ap-*" } { $config.compliance.pipl }
        default { $config.compliance.default }
    }
    
    $complianceConfig = @{
        region = $Region
        regulations = $complianceRules
        dataRetention = $config.compliance.dataRetention
        auditSettings = $config.compliance.auditSettings
    }
    
    $configJson = $complianceConfig | ConvertTo-Json -Depth 10
    kubectl create configmap compliance-config --from-literal=config.json="$configJson" --dry-run=client -o yaml | kubectl apply -f -
    
    # 重启合规性服务
    kubectl rollout restart deployment/compliance-service
    
    Write-Log "合规性配置完成" "INFO"
}

# 配置监控
function Configure-Monitoring {
    param($config)
    
    if (!$EnableMonitoring) {
        Write-Log "跳过监控配置" "INFO"
        return
    }
    
    Write-Log "配置监控系统..." "INFO"
    
    # 部署Prometheus
    Write-Log "部署Prometheus..." "INFO"
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    
    $prometheusValues = @"
server:
  global:
    external_labels:
      region: $Region
      environment: $Environment
  retention: "30d"
  
alertmanager:
  enabled: true
  config:
    global:
      smtp_smarthost: '$($config.monitoring.smtp.host):$($config.monitoring.smtp.port)'
      smtp_from: '$($config.monitoring.smtp.from)'
    
grafana:
  enabled: true
  adminPassword: '$($config.monitoring.grafana.adminPassword)'
"@
    
    $prometheusValues | Out-File -FilePath "prometheus-values.yaml" -Encoding UTF8
    
    if (!$DryRun) {
        helm upgrade --install prometheus prometheus-community/kube-prometheus-stack -f prometheus-values.yaml
    }
    
    # 部署自定义监控规则
    kubectl apply -f "infrastructure/monitoring/prometheus-rules.yaml"
    kubectl apply -f "infrastructure/monitoring/grafana-dashboards.yaml"
    
    Write-Log "监控系统配置完成" "INFO"
}

# 运行健康检查
function Test-Deployment {
    Write-Log "运行部署健康检查..." "INFO"
    
    # 检查服务状态
    $services = @("core-services", "localization-service", "compliance-service")
    foreach ($service in $services) {
        $status = kubectl get deployment $service -o jsonpath='{.status.readyReplicas}'
        $desired = kubectl get deployment $service -o jsonpath='{.spec.replicas}'
        
        if ($status -ne $desired) {
            Write-Log "服务 $service 未就绪: $status/$desired" "WARNING"
        } else {
            Write-Log "服务 $service 运行正常: $status/$desired" "INFO"
        }
    }
    
    # 检查数据库连接
    Write-Log "检查数据库连接..." "INFO"
    $dbPod = kubectl get pods -l app=postgresql-primary -o jsonpath='{.items[0].metadata.name}'
    $dbTest = kubectl exec $dbPod -- psql -U postgres -c "SELECT 1" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Log "数据库连接正常" "INFO"
    } else {
        Write-Log "数据库连接失败: $dbTest" "ERROR"
    }
    
    # 检查API端点
    Write-Log "检查API端点..." "INFO"
    $apiUrl = $config.api.baseUrl
    try {
        $response = Invoke-RestMethod -Uri "$apiUrl/health" -Method GET -TimeoutSec 30
        Write-Log "API健康检查通过: $($response.status)" "INFO"
    } catch {
        Write-Log "API健康检查失败: $($_.Exception.Message)" "ERROR"
    }
    
    Write-Log "健康检查完成" "INFO"
}

# 生成部署报告
function New-DeploymentReport {
    param($config, $startTime)
    
    $endTime = Get-Date
    $duration = $endTime - $startTime
    
    $report = @{
        deployment = @{
            environment = $Environment
            region = $Region
            startTime = $startTime.ToString("yyyy-MM-dd HH:mm:ss")
            endTime = $endTime.ToString("yyyy-MM-dd HH:mm:ss")
            duration = $duration.ToString("hh\:mm\:ss")
            status = "completed"
        }
        services = @{
            coreServices = (kubectl get deployment core-services -o jsonpath='{.status.readyReplicas}')
            localizationService = (kubectl get deployment localization-service -o jsonpath='{.status.readyReplicas}')
            complianceService = (kubectl get deployment compliance-service -o jsonpath='{.status.readyReplicas}')
        }
        infrastructure = @{
            region = $Region
            kubernetesVersion = (kubectl version --short --client)
            nodeCount = (kubectl get nodes --no-headers | Measure-Object).Count
        }
        configuration = @{
            localizationEnabled = $true
            complianceEnabled = $EnableCompliance
            monitoringEnabled = $EnableMonitoring
        }
    }
    
    $reportJson = $report | ConvertTo-Json -Depth 10
    $reportFile = "deployment-report-$Environment-$Region-$(Get-Date -Format 'yyyyMMdd-HHmmss').json"
    $reportJson | Out-File -FilePath $reportFile -Encoding UTF8
    
    Write-Log "部署报告已生成: $reportFile" "INFO"
    return $reportFile
}

# 主部署流程
function Start-GlobalDeployment {
    $startTime = Get-Date
    Write-Log "开始全球化部署: $Environment 环境, $Region 区域" "INFO"
    
    try {
        # 验证先决条件
        if (!$SkipValidation) {
            Test-Prerequisites
        }
        
        # 加载配置
        $config = Get-DeploymentConfig
        
        if ($DryRun) {
            Write-Log "执行干运行模式，不会进行实际部署" "INFO"
        }
        
        # 部署基础设施
        $outputs = Deploy-Infrastructure -config $config
        
        # 部署数据库
        Deploy-Database -config $config -outputs $outputs
        
        # 部署核心服务
        Deploy-CoreServices -config $config
        
        # 部署前端应用
        Deploy-Frontend -config $config
        
        # 配置本地化
        Configure-Localization -config $config
        
        # 配置合规性
        Configure-Compliance -config $config
        
        # 配置监控
        Configure-Monitoring -config $config
        
        # 运行健康检查
        if (!$DryRun) {
            Start-Sleep -Seconds 30  # 等待服务稳定
            Test-Deployment
        }
        
        # 生成部署报告
        $reportFile = New-DeploymentReport -config $config -startTime $startTime
        
        Write-Log "全球化部署完成！" "INFO"
        Write-Log "部署报告: $reportFile" "INFO"
        
    } catch {
        Write-Log "部署失败: $($_.Exception.Message)" "ERROR"
        Write-Log "错误详情: $($_.Exception.StackTrace)" "ERROR"
        throw
    }
}

# 执行部署
Start-GlobalDeployment