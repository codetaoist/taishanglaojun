# 太上老君AI平台 - Docker镜像构建脚本 (PowerShell版本)
# 版本: 1.0.0
# 创建时间: 2024-01-01

param(
    [switch]$Push
)

# 配置
$Registry = "taishanglaojun"
$Version = "latest"
$BuildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$GitCommit = try { (git rev-parse --short HEAD 2>$null) } catch { "unknown" }

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

# 检查Docker是否可用
function Test-Docker {
    try {
        $null = Get-Command docker -ErrorAction Stop
        $null = docker info 2>$null
        if ($LASTEXITCODE -ne 0) {
            throw "Docker daemon 未运行"
        }
        Write-LogSuccess "Docker 环境检查通过"
        return $true
    }
    catch {
        Write-LogError "Docker 未安装或daemon未运行: $($_.Exception.Message)"
        return $false
    }
}

# 构建核心服务镜像
function Build-CoreServices {
    Write-LogInfo "构建核心服务镜像..."
    
    Push-Location "core-services"
    
    try {
        # 创建优化的Dockerfile
        $DockerfileContent = @'
# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN adduser -D -s /bin/sh appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

# 创建必要的目录
RUN mkdir -p logs uploads && chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./main"]
'@
        
        $DockerfileContent | Out-File -FilePath "Dockerfile.optimized" -Encoding UTF8
        
        # 构建镜像
        docker build -f Dockerfile.optimized -t "${Registry}/core-services:${Version}" -t "${Registry}/core-services:${GitCommit}" --build-arg BUILD_DATE="$BuildDate" --build-arg GIT_COMMIT="$GitCommit" --build-arg VERSION="$Version" .
        
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "核心服务镜像构建完成"
        } else {
            throw "核心服务镜像构建失败"
        }
    }
    finally {
        Pop-Location
    }
}

# 构建前端镜像
function Build-Frontend {
    Write-LogInfo "构建前端镜像..."
    
    Push-Location "frontend\web-app"
    
    try {
        # 创建优化的Dockerfile
        $DockerfileContent = @'
# 多阶段构建 - 构建阶段
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制package文件
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 运行阶段
FROM nginx:alpine

# 复制自定义nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 从构建阶段复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 创建健康检查页面
RUN echo '<!DOCTYPE html><html><head><title>Health Check</title></head><body><h1>OK</h1></body></html>' > /usr/share/nginx/html/health

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
'@
        
        $DockerfileContent | Out-File -FilePath "Dockerfile.optimized" -Encoding UTF8
        
        # 创建nginx配置
        $NginxConfig = @'
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
    
    # 访问日志
    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;
    
    # 基本设置
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    
    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;
    
    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    
    server {
        listen 80;
        server_name localhost;
        root /usr/share/nginx/html;
        index index.html index.htm;
        
        # 静态文件缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
        
        # API代理
        location /api/ {
            proxy_pass http://core-services:8080/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        # SPA路由支持
        location / {
            try_files $uri $uri/ /index.html;
        }
        
        # 健康检查
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
'@
        
        $NginxConfig | Out-File -FilePath "nginx.conf" -Encoding UTF8
        
        # 构建镜像
        docker build -f Dockerfile.optimized -t "${Registry}/frontend:${Version}" -t "${Registry}/frontend:${GitCommit}" --build-arg BUILD_DATE="$BuildDate" --build-arg GIT_COMMIT="$GitCommit" --build-arg VERSION="$Version" .
        
        if ($LASTEXITCODE -eq 0) {
            Write-LogSuccess "前端镜像构建完成"
        } else {
            throw "前端镜像构建失败"
        }
    }
    finally {
        Pop-Location
    }
}

# 构建AI服务镜像
function Build-AIService {
    if (Test-Path "ai-service") {
        Write-LogInfo "构建AI服务镜像..."
        
        Push-Location "ai-service"
        
        try {
            # 创建优化的Dockerfile
            $DockerfileContent = @'
# 多阶段构建 - 构建阶段
FROM python:3.11-slim AS builder

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# 复制requirements文件
COPY requirements.txt .

# 安装Python依赖
RUN pip install --no-cache-dir --user -r requirements.txt

# 运行阶段
FROM python:3.11-slim

# 安装运行时依赖
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 创建非root用户
RUN useradd --create-home --shell /bin/bash appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制Python包
COPY --from=builder /root/.local /home/appuser/.local

# 复制应用代码
COPY . .

# 设置权限
RUN chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 设置PATH
ENV PATH=/home/appuser/.local/bin:$PATH

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# 启动应用
CMD ["python", "main.py"]
'@
            
            $DockerfileContent | Out-File -FilePath "Dockerfile.optimized" -Encoding UTF8
            
            # 构建镜像
            docker build -f Dockerfile.optimized -t "${Registry}/ai-service:${Version}" -t "${Registry}/ai-service:${GitCommit}" --build-arg BUILD_DATE="$BuildDate" --build-arg GIT_COMMIT="$GitCommit" --build-arg VERSION="$Version" .
            
            if ($LASTEXITCODE -eq 0) {
                Write-LogSuccess "AI服务镜像构建完成"
            } else {
                throw "AI服务镜像构建失败"
            }
        }
        finally {
            Pop-Location
        }
    } else {
        Write-LogWarning "ai-service目录不存在，跳过AI服务镜像构建"
    }
}

# 显示镜像信息
function Show-Images {
    Write-LogInfo "构建的镜像列表:"
    docker images | Select-String $Registry
    
    Write-Host ""
    Write-LogInfo "镜像大小统计:"
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | Select-String $Registry
}

# 推送镜像到仓库
function Push-Images {
    if ($Push) {
        Write-LogInfo "推送镜像到仓库..."
        
        docker push "${Registry}/core-services:${Version}"
        docker push "${Registry}/core-services:${GitCommit}"
        docker push "${Registry}/frontend:${Version}"
        docker push "${Registry}/frontend:${GitCommit}"
        
        if (Test-Path "ai-service") {
            docker push "${Registry}/ai-service:${Version}"
            docker push "${Registry}/ai-service:${GitCommit}"
        }
        
        Write-LogSuccess "镜像推送完成"
    }
}

# 清理构建文件
function Remove-BuildFiles {
    Write-LogInfo "清理构建文件..."
    Get-ChildItem -Recurse -Name "Dockerfile.optimized" | Remove-Item -Force
    Get-ChildItem -Recurse -Name "nginx.conf" | Remove-Item -Force -ErrorAction SilentlyContinue
    Write-LogSuccess "清理完成"
}

# 主函数
function Main {
    Write-Host "太上老君AI平台 - Docker镜像构建脚本" -ForegroundColor $Colors.Blue
    Write-Host "========================================" -ForegroundColor $Colors.Blue
    
    if (!(Test-Docker)) { exit 1 }
    
    try {
        # 构建所有镜像
        Build-CoreServices
        Build-Frontend
        Build-AIService
        
        Show-Images
        Push-Images
        
        Write-LogSuccess "所有镜像构建完成！"
    }
    catch {
        Write-LogError "构建过程中发生错误: $($_.Exception.Message)"
        exit 1
    }
    finally {
        Remove-BuildFiles
    }
}

# 执行主函数
Main