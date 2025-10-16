# 太上老君AI平台 - Docker部署指南

## 概述

本文档详细介绍如何使用Docker容器化部署太上老君AI平台，包括镜像构建、容器编排、网络配置和存储管理等内容。

## Docker镜像构建

### 1. 前端镜像构建

#### 开发环境Dockerfile
```dockerfile
# frontend/Dockerfile.dev
FROM node:18-alpine

WORKDIR /app

# 安装依赖
COPY package*.json ./
RUN npm ci

# 复制源代码
COPY . .

# 暴露端口
EXPOSE 3000

# 启动开发服务器
CMD ["npm", "start"]
```

#### 生产环境Dockerfile
```dockerfile
# frontend/Dockerfile
FROM node:18-alpine AS builder

WORKDIR /app

# 复制package文件并安装依赖
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# 复制源代码并构建
COPY . .
RUN npm run build

# 生产阶段
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 创建非root用户
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# 设置权限
RUN chown -R nextjs:nodejs /usr/share/nginx/html && \
    chown -R nextjs:nodejs /var/cache/nginx && \
    chown -R nextjs:nodejs /var/log/nginx && \
    chown -R nextjs:nodejs /etc/nginx/conf.d

# 切换到非root用户
USER nextjs

EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

#### Nginx配置文件
```nginx
# frontend/nginx.conf
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    # 性能优化
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 10240;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/xml+rss
        application/json;

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    server {
        listen 80;
        server_name localhost;
        root /usr/share/nginx/html;
        index index.html index.htm;

        # 静态资源缓存
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # SPA路由支持
        location / {
            try_files $uri $uri/ /index.html;
        }

        # API代理
        location /api/ {
            proxy_pass http://backend:8080/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # 健康检查
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

### 2. 后端镜像构建

#### 开发环境Dockerfile
```dockerfile
# backend/Dockerfile.dev
FROM golang:1.21-alpine

WORKDIR /app

# 安装开发工具
RUN go install github.com/cosmtrek/air@latest

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 暴露端口
EXPOSE 8080

# 使用air进行热重载
CMD ["air", "-c", ".air.toml"]
```

#### 生产环境Dockerfile
```dockerfile
# backend/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# 生产阶段
FROM scratch

# 复制时区数据和CA证书
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 复制应用程序
COPY --from=builder /app/main /main

# 复制配置文件
COPY --from=builder /app/configs /configs

# 创建非root用户
USER 65534:65534

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/main", "health"]

ENTRYPOINT ["/main"]
```

#### Air配置文件
```toml
# backend/.air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
```

### 3. AI服务镜像构建

#### 开发环境Dockerfile
```dockerfile
# ai-service/Dockerfile.dev
FROM python:3.11-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# 复制requirements文件
COPY requirements-dev.txt ./
RUN pip install --no-cache-dir -r requirements-dev.txt

# 复制源代码
COPY . .

# 暴露端口
EXPOSE 8000

# 启动开发服务器
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--reload"]
```

#### 生产环境Dockerfile
```dockerfile
# ai-service/Dockerfile
FROM python:3.11-slim AS builder

WORKDIR /app

# 安装构建依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# 复制requirements文件并安装依赖
COPY requirements.txt ./
RUN pip install --no-cache-dir --user -r requirements.txt

# 生产阶段
FROM python:3.11-slim

WORKDIR /app

# 创建非root用户
RUN groupadd -r appuser && useradd -r -g appuser appuser

# 复制Python包
COPY --from=builder /root/.local /home/appuser/.local

# 复制应用代码
COPY . .

# 设置权限
RUN chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 设置PATH
ENV PATH=/home/appuser/.local/bin:$PATH

EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

# 启动应用
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "4"]
```

## Docker Compose配置

### 1. 开发环境配置

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  # 前端服务
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      - REACT_APP_API_URL=http://localhost:8080
      - REACT_APP_WS_URL=ws://localhost:8080
      - CHOKIDAR_USEPOLLING=true
    depends_on:
      - backend
    networks:
      - taishanglaojun-network

  # 后端API网关
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
    environment:
      - GIN_MODE=debug
      - DATABASE_URL=postgres://postgres:password@postgres:5432/taishanglaojun?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=dev-secret-key
      - LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - taishanglaojun-network

  # AI服务
  ai-service:
    build:
      context: ./ai-service
      dockerfile: Dockerfile.dev
    ports:
      - "8000:8000"
    volumes:
      - ./ai-service:/app
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/taishanglaojun?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - QDRANT_URL=http://qdrant:6333
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      qdrant:
        condition: service_started
    networks:
      - taishanglaojun-network

  # PostgreSQL数据库
  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=taishanglaojun
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_INITDB_ARGS=--encoding=UTF-8 --lc-collate=C --lc-ctype=C
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - taishanglaojun-network

  # Redis缓存
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes --requirepass ""
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - taishanglaojun-network

  # Elasticsearch搜索引擎
  elasticsearch:
    image: elasticsearch:8.11.0
    ports:
      - "9200:9200"
      - "9300:9300"
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms1g -Xmx1g"
      - bootstrap.memory_lock=true
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 5
    networks:
      - taishanglaojun-network

  # Qdrant向量数据库
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_data:/qdrant/storage
    environment:
      - QDRANT__SERVICE__HTTP_PORT=6333
      - QDRANT__SERVICE__GRPC_PORT=6334
    networks:
      - taishanglaojun-network

  # MinIO对象存储
  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin123
      - MINIO_BROWSER_REDIRECT_URL=http://localhost:9001
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
    networks:
      - taishanglaojun-network

  # Prometheus监控
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    networks:
      - taishanglaojun-network

  # Grafana可视化
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - taishanglaojun-network

  # Jaeger链路追踪
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - taishanglaojun-network

volumes:
  postgres_data:
  redis_data:
  elasticsearch_data:
  qdrant_data:
  minio_data:
  prometheus_data:
  grafana_data:

networks:
  taishanglaojun-network:
    driver: bridge
```

### 2. 生产环境配置

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  # Nginx反向代理
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./ssl:/etc/nginx/ssl
      - nginx_logs:/var/log/nginx
    depends_on:
      - frontend
      - backend
    networks:
      - taishanglaojun-network
    restart: unless-stopped

  # 前端服务
  frontend:
    image: taishanglaojun/frontend:${VERSION:-latest}
    expose:
      - "80"
    environment:
      - NODE_ENV=production
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M

  # 后端API网关
  backend:
    image: taishanglaojun/backend:${VERSION:-latest}
    expose:
      - "8080"
    environment:
      - GIN_MODE=release
      - DATABASE_URL=postgres://postgres:${POSTGRES_PASSWORD}@postgres:5432/taishanglaojun?sslmode=require
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - LOG_LEVEL=info
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

  # AI服务
  ai-service:
    image: taishanglaojun/ai-service:${VERSION:-latest}
    expose:
      - "8000"
    environment:
      - DATABASE_URL=postgres://postgres:${POSTGRES_PASSWORD}@postgres:5432/taishanglaojun?sslmode=require
      - REDIS_URL=redis://redis:6379
      - QDRANT_URL=http://qdrant:6333
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - LOG_LEVEL=info
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      qdrant:
        condition: service_started
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G

  # PostgreSQL主数据库
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=taishanglaojun
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_INITDB_ARGS=--encoding=UTF-8 --lc-collate=C --lc-ctype=C
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
      - ./backup:/backup
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G

  # Redis集群
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 1G

  # Elasticsearch集群
  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=true
      - ELASTIC_PASSWORD=${ELASTIC_PASSWORD}
      - "ES_JAVA_OPTS=-Xms2g -Xmx2g"
      - bootstrap.memory_lock=true
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    healthcheck:
      test: ["CMD-SHELL", "curl -u elastic:${ELASTIC_PASSWORD} -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 5
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G

  # Qdrant向量数据库
  qdrant:
    image: qdrant/qdrant:latest
    environment:
      - QDRANT__SERVICE__HTTP_PORT=6333
      - QDRANT__SERVICE__GRPC_PORT=6334
    volumes:
      - qdrant_data:/qdrant/storage
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 1G

  # MinIO对象存储
  minio:
    image: minio/minio:latest
    environment:
      - MINIO_ROOT_USER=${MINIO_ROOT_USER}
      - MINIO_ROOT_PASSWORD=${MINIO_ROOT_PASSWORD}
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
    networks:
      - taishanglaojun-network
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

volumes:
  postgres_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/postgres
  redis_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/redis
  elasticsearch_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/elasticsearch
  qdrant_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/qdrant
  minio_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/minio
  nginx_logs:
    driver: local

networks:
  taishanglaojun-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 环境变量配置

### 1. 开发环境变量

```bash
# .env.dev
# 数据库配置
POSTGRES_PASSWORD=password
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=dev-secret-key

# AI服务配置
OPENAI_API_KEY=sk-your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key

# 对象存储配置
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin123

# Elasticsearch配置
ELASTIC_PASSWORD=elastic123

# 版本标签
VERSION=dev
```

### 2. 生产环境变量

```bash
# .env.prod
# 数据库配置
POSTGRES_PASSWORD=your-secure-postgres-password
REDIS_PASSWORD=your-secure-redis-password

# JWT配置
JWT_SECRET=your-secure-jwt-secret

# AI服务配置
OPENAI_API_KEY=sk-your-production-openai-key
ANTHROPIC_API_KEY=your-production-anthropic-key

# 对象存储配置
MINIO_ROOT_USER=your-minio-user
MINIO_ROOT_PASSWORD=your-secure-minio-password

# Elasticsearch配置
ELASTIC_PASSWORD=your-secure-elastic-password

# 版本标签
VERSION=v1.0.0
```

## 部署脚本

### 1. 开发环境部署脚本

```bash
#!/bin/bash
# scripts/deploy-dev.sh

set -e

echo "🚀 开始部署开发环境..."

# 检查Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 加载环境变量
if [ -f .env.dev ]; then
    export $(cat .env.dev | grep -v '#' | xargs)
else
    echo "❌ .env.dev文件不存在"
    exit 1
fi

# 停止现有容器
echo "🛑 停止现有容器..."
docker-compose -f docker-compose.dev.yml down

# 清理旧镜像（可选）
read -p "是否清理旧镜像? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理旧镜像..."
    docker system prune -f
fi

# 构建镜像
echo "🔨 构建镜像..."
docker-compose -f docker-compose.dev.yml build

# 启动服务
echo "🚀 启动服务..."
docker-compose -f docker-compose.dev.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.dev.yml ps

# 运行数据库迁移
echo "📊 运行数据库迁移..."
docker-compose -f docker-compose.dev.yml exec backend ./main migrate

# 显示访问地址
echo "✅ 部署完成！"
echo "🌐 前端地址: http://localhost:3000"
echo "🔧 后端API: http://localhost:8080"
echo "🤖 AI服务: http://localhost:8000"
echo "📊 Grafana: http://localhost:3001 (admin/admin123)"
echo "📈 Prometheus: http://localhost:9090"
echo "🔍 Jaeger: http://localhost:16686"
```

### 2. 生产环境部署脚本

```bash
#!/bin/bash
# scripts/deploy-prod.sh

set -e

echo "🚀 开始部署生产环境..."

# 检查必要的工具
for cmd in docker docker-compose; do
    if ! command -v $cmd &> /dev/null; then
        echo "❌ $cmd未安装，请先安装"
        exit 1
    fi
done

# 加载环境变量
if [ -f .env.prod ]; then
    export $(cat .env.prod | grep -v '#' | xargs)
else
    echo "❌ .env.prod文件不存在"
    exit 1
fi

# 检查必要的环境变量
required_vars=("POSTGRES_PASSWORD" "JWT_SECRET" "OPENAI_API_KEY")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ 环境变量 $var 未设置"
        exit 1
    fi
done

# 创建数据目录
echo "📁 创建数据目录..."
sudo mkdir -p /data/{postgres,redis,elasticsearch,qdrant,minio}
sudo chown -R $USER:$USER /data

# 备份现有数据（如果存在）
if [ -d "/data/postgres" ] && [ "$(ls -A /data/postgres)" ]; then
    echo "💾 备份现有数据..."
    backup_dir="/backup/$(date +%Y%m%d_%H%M%S)"
    sudo mkdir -p $backup_dir
    sudo cp -r /data/* $backup_dir/
    echo "✅ 数据已备份到 $backup_dir"
fi

# 拉取最新镜像
echo "📥 拉取最新镜像..."
docker-compose -f docker-compose.prod.yml pull

# 滚动更新服务
echo "🔄 滚动更新服务..."
services=("frontend" "backend" "ai-service")
for service in "${services[@]}"; do
    echo "🔄 更新服务: $service"
    docker-compose -f docker-compose.prod.yml up -d --no-deps $service
    
    # 等待服务健康检查
    echo "⏳ 等待 $service 健康检查..."
    timeout 60 bash -c "until docker-compose -f docker-compose.prod.yml exec $service curl -f http://localhost/health; do sleep 5; done"
    
    if [ $? -ne 0 ]; then
        echo "❌ $service 健康检查失败，回滚..."
        docker-compose -f docker-compose.prod.yml rollback $service
        exit 1
    fi
    
    echo "✅ $service 更新成功"
done

# 清理旧镜像
echo "🧹 清理旧镜像..."
docker image prune -f

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.prod.yml ps

echo "✅ 生产环境部署完成！"
```

## 监控和日志

### 1. 容器监控脚本

```bash
#!/bin/bash
# scripts/monitor.sh

# 显示容器状态
echo "=== 容器状态 ==="
docker-compose ps

echo -e "\n=== 容器资源使用 ==="
docker stats --no-stream

echo -e "\n=== 磁盘使用 ==="
df -h

echo -e "\n=== 内存使用 ==="
free -h

echo -e "\n=== 网络连接 ==="
netstat -tuln | grep -E ':(80|443|3000|8080|8000|5432|6379|9200|6333)'

echo -e "\n=== 最近的错误日志 ==="
docker-compose logs --tail=10 | grep -i error
```

### 2. 日志收集配置

```yaml
# logging/docker-compose.logging.yml
version: '3.8'

services:
  # Filebeat日志收集
  filebeat:
    image: elastic/filebeat:8.11.0
    user: root
    volumes:
      - ./logging/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - ELASTICSEARCH_HOST=elasticsearch:9200
      - ELASTICSEARCH_USERNAME=elastic
      - ELASTICSEARCH_PASSWORD=${ELASTIC_PASSWORD}
    depends_on:
      - elasticsearch
    networks:
      - taishanglaojun-network

  # Logstash日志处理
  logstash:
    image: elastic/logstash:8.11.0
    volumes:
      - ./logging/logstash.conf:/usr/share/logstash/pipeline/logstash.conf:ro
    environment:
      - ELASTICSEARCH_HOST=elasticsearch:9200
      - ELASTICSEARCH_USERNAME=elastic
      - ELASTICSEARCH_PASSWORD=${ELASTIC_PASSWORD}
    depends_on:
      - elasticsearch
    networks:
      - taishanglaojun-network

networks:
  taishanglaojun-network:
    external: true
```

## 故障排除

### 1. 常见问题诊断脚本

```bash
#!/bin/bash
# scripts/diagnose.sh

echo "🔍 开始系统诊断..."

# 检查Docker服务
echo "=== Docker服务状态 ==="
systemctl status docker

# 检查容器状态
echo -e "\n=== 容器状态 ==="
docker-compose ps

# 检查容器日志
echo -e "\n=== 容器错误日志 ==="
for service in frontend backend ai-service postgres redis; do
    echo "--- $service 日志 ---"
    docker-compose logs --tail=20 $service | grep -i error || echo "无错误日志"
done

# 检查网络连接
echo -e "\n=== 网络连接测试 ==="
services=("frontend:80" "backend:8080" "ai-service:8000" "postgres:5432" "redis:6379")
for service in "${services[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if docker-compose exec $name nc -z localhost $port; then
        echo "✅ $name:$port 连接正常"
    else
        echo "❌ $name:$port 连接失败"
    fi
done

# 检查磁盘空间
echo -e "\n=== 磁盘空间 ==="
df -h

# 检查内存使用
echo -e "\n=== 内存使用 ==="
free -h

echo -e "\n🔍 诊断完成"
```

### 2. 自动恢复脚本

```bash
#!/bin/bash
# scripts/auto-recovery.sh

# 检查服务健康状态并自动恢复
check_and_recover() {
    local service=$1
    local health_url=$2
    
    echo "🔍 检查 $service 健康状态..."
    
    if ! curl -f $health_url > /dev/null 2>&1; then
        echo "❌ $service 不健康，尝试重启..."
        docker-compose restart $service
        
        # 等待服务恢复
        sleep 30
        
        if curl -f $health_url > /dev/null 2>&1; then
            echo "✅ $service 恢复正常"
        else
            echo "❌ $service 恢复失败，发送告警..."
            # 这里可以添加告警逻辑
        fi
    else
        echo "✅ $service 健康状态正常"
    fi
}

# 检查各个服务
check_and_recover "frontend" "http://localhost:3000/health"
check_and_recover "backend" "http://localhost:8080/health"
check_and_recover "ai-service" "http://localhost:8000/health"
```

## 相关文档链接

- [部署概览](./deployment-overview.md)
- [Kubernetes部署指南](./kubernetes-deployment.md)
- [监控运维指南](./monitoring-operations.md)
- [安全配置指南](./security-configuration.md)
- [性能优化指南](./performance-optimization.md)