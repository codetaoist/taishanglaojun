# 高级AI功能部署指南

## 概述

本文档提供了太上老君项目中高级AI功能的完整部署指南，包括环境准备、依赖安装、配置设置、部署步骤和运维管理。

## 目录

1. [环境要求](#环境要求)
2. [依赖安装](#依赖安装)
3. [配置准备](#配置准备)
4. [本地开发部署](#本地开发部署)
5. [Docker部署](#docker部署)
6. [Kubernetes部署](#kubernetes部署)
7. [生产环境部署](#生产环境部署)
8. [监控和日志](#监控和日志)
9. [故障排除](#故障排除)
10. [性能优化](#性能优化)

## 环境要求

### 硬件要求

#### 最低配置
- **CPU**: 4核心 Intel/AMD x64处理器
- **内存**: 8GB RAM
- **存储**: 50GB可用空间
- **网络**: 100Mbps带宽

#### 推荐配置
- **CPU**: 8核心以上，支持AVX指令集
- **内存**: 32GB RAM或更多
- **GPU**: NVIDIA RTX 3080或更高（支持CUDA 11.0+）
- **存储**: 200GB SSD存储
- **网络**: 1Gbps带宽

#### 生产环境配置
- **CPU**: 16核心以上
- **内存**: 64GB RAM或更多
- **GPU**: 多张NVIDIA A100或V100
- **存储**: 1TB NVMe SSD + 网络存储
- **网络**: 10Gbps带宽

### 软件要求

#### 操作系统
- **Linux**: Ubuntu 20.04+ / CentOS 8+ / RHEL 8+
- **Windows**: Windows 10/11 Pro (仅开发环境)
- **macOS**: macOS 11+ (仅开发环境)

#### 基础软件
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **Kubernetes**: 1.24+
- **Git**: 2.30+
- **curl**: 7.68+

#### 编程环境
- **Go**: 1.19+
- **Python**: 3.9+
- **Node.js**: 16+
- **npm/yarn**: 最新版本

## 依赖安装

### 1. 系统依赖安装

#### Ubuntu/Debian
```bash
# 更新系统包
sudo apt update && sudo apt upgrade -y

# 安装基础依赖
sudo apt install -y \
    curl \
    wget \
    git \
    build-essential \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release

# 安装Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 安装Kubernetes工具
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt update
sudo apt install -y kubectl kubelet kubeadm
```

#### CentOS/RHEL
```bash
# 更新系统包
sudo yum update -y

# 安装基础依赖
sudo yum install -y \
    curl \
    wget \
    git \
    gcc \
    gcc-c++ \
    make \
    yum-utils \
    device-mapper-persistent-data \
    lvm2

# 安装Docker
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl start docker
sudo systemctl enable docker

# 安装Kubernetes工具
cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl
EOF
sudo yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
sudo systemctl enable kubelet
```

### 2. 编程环境安装

#### Go环境
```bash
# 下载并安装Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### Python环境
```bash
# 安装Python和pip
sudo apt install -y python3 python3-pip python3-venv

# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 升级pip
pip install --upgrade pip

# 安装AI相关依赖
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
pip install transformers datasets accelerate
pip install numpy pandas scikit-learn matplotlib seaborn
pip install fastapi uvicorn pydantic
pip install redis psycopg2-binary sqlalchemy
```

#### Node.js环境
```bash
# 安装Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 验证安装
node --version
npm --version

# 安装全局工具
sudo npm install -g yarn pm2
```

### 3. GPU支持安装

#### NVIDIA驱动和CUDA
```bash
# 安装NVIDIA驱动
sudo apt install -y nvidia-driver-525

# 下载并安装CUDA Toolkit
wget https://developer.download.nvidia.com/compute/cuda/11.8.0/local_installers/cuda_11.8.0_520.61.05_linux.run
sudo sh cuda_11.8.0_520.61.05_linux.run

# 配置环境变量
echo 'export PATH=/usr/local/cuda-11.8/bin:$PATH' >> ~/.bashrc
echo 'export LD_LIBRARY_PATH=/usr/local/cuda-11.8/lib64:$LD_LIBRARY_PATH' >> ~/.bashrc
source ~/.bashrc

# 验证安装
nvidia-smi
nvcc --version
```

#### Docker GPU支持
```bash
# 安装NVIDIA Container Toolkit
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list

sudo apt-get update
sudo apt-get install -y nvidia-docker2
sudo systemctl restart docker

# 测试GPU支持
docker run --rm --gpus all nvidia/cuda:11.8-base-ubuntu20.04 nvidia-smi
```

## 配置准备

### 1. 环境变量配置

创建环境配置文件：

```bash
# 创建配置目录
mkdir -p ~/taishanglaojun/config

# 创建环境变量文件
cat > ~/taishanglaojun/config/.env << EOF
# 应用配置
APP_ENV=development
APP_PORT=8080
APP_HOST=0.0.0.0
APP_DEBUG=true

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=taishanglaojun
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# AI服务配置
OPENAI_API_KEY=your_openai_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key
GOOGLE_API_KEY=your_google_api_key
AZURE_OPENAI_API_KEY=your_azure_api_key
HUGGINGFACE_API_KEY=your_huggingface_api_key

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY=24h

# 监控配置
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
GRAFANA_ADMIN_PASSWORD=admin123

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# 存储配置
STORAGE_TYPE=local
STORAGE_PATH=/data
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-west-2
AWS_BUCKET=taishanglaojun-storage
EOF
```

### 2. 数据库初始化

#### PostgreSQL设置
```bash
# 启动PostgreSQL容器
docker run -d \
  --name postgres \
  -e POSTGRES_DB=taishanglaojun \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=your_password \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  postgres:15

# 等待数据库启动
sleep 10

# 创建数据库表
docker exec -i postgres psql -U postgres -d taishanglaojun << EOF
-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建AI任务表
CREATE TABLE IF NOT EXISTS ai_tasks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    task_type VARCHAR(100) NOT NULL,
    input_data JSONB,
    output_data JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- 创建模型表
CREATE TABLE IF NOT EXISTS ai_models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    model_type VARCHAR(100) NOT NULL,
    config JSONB,
    performance_metrics JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, version)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ai_tasks_user_id ON ai_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_tasks(status);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_created_at ON ai_tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_models_name ON ai_models(name);
EOF
```

#### Redis设置
```bash
# 启动Redis容器
docker run -d \
  --name redis \
  -p 6379:6379 \
  -v redis_data:/data \
  redis:7-alpine redis-server --requirepass your_redis_password

# 测试连接
docker exec -it redis redis-cli -a your_redis_password ping
```

### 3. 配置文件准备

#### 应用配置文件
```yaml
# config/app.yaml
app:
  name: "TaiShangLaoJun Advanced AI"
  version: "1.0.0"
  environment: "development"
  host: "0.0.0.0"
  port: 8080
  debug: true

database:
  host: "localhost"
  port: 5432
  name: "taishanglaojun"
  user: "postgres"
  password: "your_password"
  ssl_mode: "disable"
  max_connections: 100
  max_idle_connections: 10
  connection_max_lifetime: "1h"

redis:
  host: "localhost"
  port: 6379
  password: "your_redis_password"
  db: 0
  pool_size: 10
  min_idle_connections: 5

ai_providers:
  openai:
    api_key: "your_openai_api_key"
    base_url: "https://api.openai.com/v1"
    model: "gpt-4"
    max_tokens: 4096
    temperature: 0.7
  
  anthropic:
    api_key: "your_anthropic_api_key"
    base_url: "https://api.anthropic.com"
    model: "claude-3-opus-20240229"
    max_tokens: 4096
  
  google:
    api_key: "your_google_api_key"
    project_id: "your_project_id"
    location: "us-central1"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "/var/log/taishanglaojun.log"
  max_size: 100
  max_backups: 5
  max_age: 30

monitoring:
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
  
  jaeger:
    enabled: true
    endpoint: "http://localhost:14268/api/traces"
  
  health_check:
    enabled: true
    path: "/health"
    interval: "30s"

security:
  jwt:
    secret: "your_jwt_secret_key"
    expiry: "24h"
    issuer: "taishanglaojun"
  
  cors:
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["*"]
    allow_credentials: true
  
  rate_limit:
    enabled: true
    requests_per_minute: 100
    burst: 200
```

## 本地开发部署

### 1. 克隆代码仓库

```bash
# 克隆项目
git clone https://github.com/your-org/taishanglaojun.git
cd taishanglaojun

# 切换到开发分支
git checkout develop

# 安装依赖
make install-deps
```

### 2. 启动开发环境

```bash
# 启动基础服务
docker-compose -f docker-compose.dev.yml up -d postgres redis

# 等待服务启动
sleep 10

# 运行数据库迁移
make migrate

# 启动后端服务
cd core-services/ai-integration
go mod tidy
go run main.go

# 启动前端服务（新终端）
cd frontend/web-app
npm install
npm run dev
```

### 3. 验证部署

```bash
# 检查后端服务
curl http://localhost:8080/health

# 检查前端服务
curl http://localhost:3000

# 检查AI功能
curl -X POST http://localhost:8080/api/v1/agi/reasoning \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the capital of France?", "context": ""}'
```

## Docker部署

### 1. 构建镜像

```bash
# 构建后端镜像
cd core-services/ai-integration
docker build -t taishanglaojun/ai-service:latest .

# 构建前端镜像
cd ../../frontend/web-app
docker build -t taishanglaojun/web-app:latest .

# 验证镜像
docker images | grep taishanglaojun
```

### 2. Docker Compose部署

创建生产环境的docker-compose文件：

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: taishanglaojun
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  ai-service:
    image: taishanglaojun/ai-service:latest
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ai_data:/app/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G

  web-app:
    image: taishanglaojun/web-app:latest
    environment:
      - VITE_API_BASE_URL=http://ai-service:8080/api/v1
    ports:
      - "3000:3000"
    depends_on:
      ai-service:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - ai-service
      - web-app
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources

volumes:
  postgres_data:
  redis_data:
  ai_data:
  prometheus_data:
  grafana_data:

networks:
  default:
    driver: bridge
```

### 3. 启动服务

```bash
# 设置环境变量
export DB_PASSWORD=your_secure_password
export REDIS_PASSWORD=your_redis_password
export OPENAI_API_KEY=your_openai_key
export JWT_SECRET=your_jwt_secret
export GRAFANA_ADMIN_PASSWORD=admin123

# 启动所有服务
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f ai-service
```

## Kubernetes部署

### 1. 准备Kubernetes集群

#### 使用kubeadm创建集群
```bash
# 初始化主节点
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

# 配置kubectl
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# 安装网络插件（Flannel）
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

# 验证集群状态
kubectl get nodes
kubectl get pods -A
```

#### 或使用托管Kubernetes服务
```bash
# AWS EKS
eksctl create cluster --name taishanglaojun --region us-west-2 --nodes 3

# Google GKE
gcloud container clusters create taishanglaojun --num-nodes=3 --zone=us-central1-a

# Azure AKS
az aks create --resource-group myResourceGroup --name taishanglaojun --node-count 3
```

### 2. 部署应用

```bash
# 进入Kubernetes配置目录
cd k8s

# 创建命名空间
kubectl apply -f namespace.yaml

# 部署存储配置
kubectl apply -f storage.yaml

# 部署密钥配置
kubectl apply -f secrets.yaml

# 部署RBAC配置
kubectl apply -f rbac.yaml

# 部署服务配置
kubectl apply -f service.yaml

# 部署应用
kubectl apply -f deployment.yaml

# 部署Ingress
kubectl apply -f ingress.yaml

# 部署HPA
kubectl apply -f hpa.yaml

# 部署监控
kubectl apply -f monitoring.yaml
```

### 3. 使用部署脚本

```bash
# 使用提供的部署脚本
chmod +x deploy-k8s.sh

# 完整部署
./deploy-k8s.sh deploy

# 验证部署
./deploy-k8s.sh verify

# 查看部署信息
./deploy-k8s.sh info

# 查看日志
./deploy-k8s.sh logs ai-service

# 更新部署
./deploy-k8s.sh update

# 清理部署
./deploy-k8s.sh cleanup
```

### 4. 验证Kubernetes部署

```bash
# 检查Pod状态
kubectl get pods -n taishanglaojun

# 检查服务状态
kubectl get svc -n taishanglaojun

# 检查Ingress状态
kubectl get ingress -n taishanglaojun

# 检查HPA状态
kubectl get hpa -n taishanglaojun

# 查看详细信息
kubectl describe deployment advanced-ai-service -n taishanglaojun

# 查看日志
kubectl logs -f deployment/advanced-ai-service -n taishanglaojun

# 端口转发测试
kubectl port-forward svc/advanced-ai-service 8080:80 -n taishanglaojun
```

## 生产环境部署

### 1. 生产环境准备

#### 基础设施要求
- **高可用Kubernetes集群**: 至少3个主节点
- **负载均衡器**: 云厂商提供的负载均衡服务
- **存储**: 高性能SSD存储，支持快照和备份
- **网络**: 专用网络，安全组配置
- **监控**: 完整的监控和告警体系

#### 安全配置
```bash
# 创建生产环境密钥
kubectl create secret generic prod-secrets \
  --from-literal=db-password=$(openssl rand -base64 32) \
  --from-literal=redis-password=$(openssl rand -base64 32) \
  --from-literal=jwt-secret=$(openssl rand -base64 64) \
  --from-literal=openai-api-key=your_production_key \
  -n taishanglaojun

# 创建TLS证书
kubectl create secret tls tls-secret \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n taishanglaojun

# 创建镜像拉取密钥
kubectl create secret docker-registry regcred \
  --docker-server=your-registry-server \
  --docker-username=your-username \
  --docker-password=your-password \
  --docker-email=your-email \
  -n taishanglaojun
```

### 2. 生产环境配置

#### 资源配置
```yaml
# production-values.yaml
replicaCount: 3

image:
  repository: your-registry/taishanglaojun/ai-service
  tag: "v1.0.0"
  pullPolicy: Always

resources:
  limits:
    cpu: 4000m
    memory: 8Gi
    nvidia.com/gpu: 1
  requests:
    cpu: 2000m
    memory: 4Gi
    nvidia.com/gpu: 1

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

persistence:
  enabled: true
  storageClass: "fast-ssd"
  size: 100Gi

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
  prometheusRule:
    enabled: true

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: ai.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: ai-tls
      hosts:
        - ai.yourdomain.com
```

### 3. CI/CD流水线

#### GitLab CI配置
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy-staging
  - deploy-production

variables:
  DOCKER_REGISTRY: your-registry.com
  IMAGE_NAME: taishanglaojun/ai-service
  KUBECONFIG_FILE: $KUBECONFIG_PROD

test:
  stage: test
  image: golang:1.21
  script:
    - cd core-services/ai-integration
    - go mod tidy
    - go test ./...
    - go vet ./...
  only:
    - merge_requests
    - develop
    - main

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $DOCKER_REGISTRY
    - docker build -t $DOCKER_REGISTRY/$IMAGE_NAME:$CI_COMMIT_SHA .
    - docker build -t $DOCKER_REGISTRY/$IMAGE_NAME:latest .
    - docker push $DOCKER_REGISTRY/$IMAGE_NAME:$CI_COMMIT_SHA
    - docker push $DOCKER_REGISTRY/$IMAGE_NAME:latest
  only:
    - develop
    - main

deploy-staging:
  stage: deploy-staging
  image: bitnami/kubectl:latest
  script:
    - echo $KUBECONFIG_STAGING | base64 -d > kubeconfig
    - export KUBECONFIG=kubeconfig
    - kubectl set image deployment/advanced-ai-service advanced-ai-service=$DOCKER_REGISTRY/$IMAGE_NAME:$CI_COMMIT_SHA -n taishanglaojun-staging
    - kubectl rollout status deployment/advanced-ai-service -n taishanglaojun-staging
  environment:
    name: staging
    url: https://staging-ai.yourdomain.com
  only:
    - develop

deploy-production:
  stage: deploy-production
  image: bitnami/kubectl:latest
  script:
    - echo $KUBECONFIG_PROD | base64 -d > kubeconfig
    - export KUBECONFIG=kubeconfig
    - kubectl set image deployment/advanced-ai-service advanced-ai-service=$DOCKER_REGISTRY/$IMAGE_NAME:$CI_COMMIT_SHA -n taishanglaojun
    - kubectl rollout status deployment/advanced-ai-service -n taishanglaojun
  environment:
    name: production
    url: https://ai.yourdomain.com
  when: manual
  only:
    - main
```

### 4. 数据库迁移

```bash
# 生产环境数据库迁移脚本
#!/bin/bash

set -e

# 配置变量
DB_HOST="your-prod-db-host"
DB_NAME="taishanglaojun"
DB_USER="postgres"
BACKUP_DIR="/backups"
MIGRATION_DIR="./migrations"

# 创建备份
echo "Creating database backup..."
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > $BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).sql

# 运行迁移
echo "Running database migrations..."
for migration in $MIGRATION_DIR/*.sql; do
    echo "Applying migration: $migration"
    psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $migration
done

echo "Database migration completed successfully!"
```

## 监控和日志

### 1. Prometheus监控配置

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'ai-service'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - taishanglaojun
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: advanced-ai-service-metrics

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx-exporter:9113']
```

### 2. Grafana仪表板

```json
{
  "dashboard": {
    "title": "TaiShangLaoJun AI Service Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "AI Inference Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(ai_inference_duration_seconds_bucket[5m]))",
            "legendFormat": "{{model_type}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])",
            "legendFormat": "Error Rate"
          }
        ]
      }
    ]
  }
}
```

### 3. 告警规则

```yaml
# monitoring/alert_rules.yml
groups:
  - name: ai-service-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"

      - alert: AIInferenceLatency
        expr: histogram_quantile(0.95, rate(ai_inference_duration_seconds_bucket[5m])) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High AI inference latency"
          description: "AI inference latency is {{ $value }}s"

      - alert: DatabaseConnectionFailure
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database connection failure"
          description: "Cannot connect to PostgreSQL database"
```

### 4. 日志聚合

#### ELK Stack配置
```yaml
# logging/elasticsearch.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elasticsearch
spec:
  replicas: 3
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
        env:
        - name: discovery.type
          value: single-node
        - name: ES_JAVA_OPTS
          value: "-Xms2g -Xmx2g"
        ports:
        - containerPort: 9200
        volumeMounts:
        - name: es-data
          mountPath: /usr/share/elasticsearch/data
      volumes:
      - name: es-data
        persistentVolumeClaim:
          claimName: elasticsearch-pvc

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: logstash
spec:
  replicas: 2
  selector:
    matchLabels:
      app: logstash
  template:
    metadata:
      labels:
        app: logstash
    spec:
      containers:
      - name: logstash
        image: docker.elastic.co/logstash/logstash:8.8.0
        ports:
        - containerPort: 5044
        volumeMounts:
        - name: logstash-config
          mountPath: /usr/share/logstash/pipeline
      volumes:
      - name: logstash-config
        configMap:
          name: logstash-config

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kibana
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kibana
  template:
    metadata:
      labels:
        app: kibana
    spec:
      containers:
      - name: kibana
        image: docker.elastic.co/kibana/kibana:8.8.0
        env:
        - name: ELASTICSEARCH_HOSTS
          value: "http://elasticsearch:9200"
        ports:
        - containerPort: 5601
```

## 故障排除

### 1. 常见问题诊断

#### 服务启动失败
```bash
# 检查Pod状态
kubectl get pods -n taishanglaojun

# 查看Pod详细信息
kubectl describe pod <pod-name> -n taishanglaojun

# 查看容器日志
kubectl logs <pod-name> -c <container-name> -n taishanglaojun

# 进入容器调试
kubectl exec -it <pod-name> -c <container-name> -n taishanglaojun -- /bin/bash
```

#### 数据库连接问题
```bash
# 测试数据库连接
kubectl run postgres-client --rm -it --image=postgres:15 -- psql -h postgres -U postgres -d taishanglaojun

# 检查数据库服务
kubectl get svc postgres -n taishanglaojun

# 查看数据库日志
kubectl logs deployment/postgres -n taishanglaojun
```

#### 网络连接问题
```bash
# 测试服务间连接
kubectl run network-test --rm -it --image=busybox -- nslookup advanced-ai-service.taishanglaojun.svc.cluster.local

# 检查网络策略
kubectl get networkpolicy -n taishanglaojun

# 测试端口连通性
kubectl run network-test --rm -it --image=busybox -- telnet advanced-ai-service 80
```

### 2. 性能问题排查

#### CPU和内存使用
```bash
# 查看资源使用情况
kubectl top pods -n taishanglaojun

# 查看节点资源使用
kubectl top nodes

# 查看详细资源指标
kubectl get --raw /apis/metrics.k8s.io/v1beta1/namespaces/taishanglaojun/pods
```

#### 存储问题
```bash
# 检查PVC状态
kubectl get pvc -n taishanglaojun

# 查看存储使用情况
kubectl exec -it <pod-name> -n taishanglaojun -- df -h

# 检查存储类
kubectl get storageclass
```

### 3. 日志分析

#### 应用日志分析
```bash
# 查看应用错误日志
kubectl logs -f deployment/advanced-ai-service -n taishanglaojun | grep ERROR

# 查看特定时间段的日志
kubectl logs --since=1h deployment/advanced-ai-service -n taishanglaojun

# 查看所有容器日志
kubectl logs -f deployment/advanced-ai-service --all-containers=true -n taishanglaojun
```

#### 系统事件分析
```bash
# 查看集群事件
kubectl get events --sort-by=.metadata.creationTimestamp -n taishanglaojun

# 查看特定资源的事件
kubectl describe deployment advanced-ai-service -n taishanglaojun
```

## 性能优化

### 1. 应用层优化

#### Go服务优化
```go
// 连接池优化
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(time.Hour)

// Redis连接池优化
rdb := redis.NewClient(&redis.Options{
    Addr:         "redis:6379",
    PoolSize:     10,
    MinIdleConns: 5,
    PoolTimeout:  30 * time.Second,
})

// HTTP客户端优化
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

#### AI模型优化
```python
# 模型量化
import torch
from torch.quantization import quantize_dynamic

# 动态量化
quantized_model = quantize_dynamic(
    model, {torch.nn.Linear}, dtype=torch.qint8
)

# 模型剪枝
import torch.nn.utils.prune as prune

prune.global_unstructured(
    parameters_to_prune,
    pruning_method=prune.L1Unstructured,
    amount=0.2,
)

# 模型缓存
from functools import lru_cache

@lru_cache(maxsize=1000)
def cached_inference(input_hash):
    return model.predict(input_data)
```

### 2. 基础设施优化

#### Kubernetes资源优化
```yaml
# 资源请求和限制优化
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi

# 节点亲和性配置
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node-type
          operator: In
          values:
          - gpu-node

# Pod反亲和性配置
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - advanced-ai-service
      topologyKey: kubernetes.io/hostname
```

#### 存储优化
```yaml
# 高性能存储类
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
  fsType: ext4
  encrypted: "true"
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
```

### 3. 网络优化

#### Ingress优化
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: advanced-ai-ingress
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "300"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-headers: "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization"
    nginx.ingress.kubernetes.io/use-gzip: "true"
    nginx.ingress.kubernetes.io/gzip-types: "text/plain,text/css,application/json,application/javascript,text/xml,application/xml,application/xml+rss,text/javascript"
```

#### 服务网格优化
```yaml
# Istio配置
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: ai-service-destination
spec:
  host: advanced-ai-service
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        maxRequestsPerConnection: 10
    loadBalancer:
      simple: LEAST_CONN
    outlierDetection:
      consecutiveErrors: 3
      interval: 30s
      baseEjectionTime: 30s
```

## 备份和恢复

### 1. 数据备份策略

#### 数据库备份
```bash
#!/bin/bash
# backup-database.sh

BACKUP_DIR="/backups/postgres"
DB_HOST="postgres"
DB_NAME="taishanglaojun"
DB_USER="postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 执行备份
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > $BACKUP_DIR/backup_${TIMESTAMP}.sql.gz

# 清理旧备份（保留30天）
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete

echo "Database backup completed: backup_${TIMESTAMP}.sql.gz"
```

#### 存储卷备份
```yaml
# 卷快照配置
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: ai-data-snapshot
spec:
  volumeSnapshotClassName: csi-aws-vsc
  source:
    persistentVolumeClaimName: ai-data-pvc
```

### 2. 恢复流程

#### 数据库恢复
```bash
#!/bin/bash
# restore-database.sh

BACKUP_FILE=$1
DB_HOST="postgres"
DB_NAME="taishanglaojun"
DB_USER="postgres"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# 停止应用服务
kubectl scale deployment advanced-ai-service --replicas=0 -n taishanglaojun

# 恢复数据库
gunzip -c $BACKUP_FILE | psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# 重启应用服务
kubectl scale deployment advanced-ai-service --replicas=3 -n taishanglaojun

echo "Database restore completed"
```

#### 配置备份和恢复
```bash
# 备份Kubernetes配置
kubectl get all,configmap,secret,pvc -n taishanglaojun -o yaml > backup-k8s-config.yaml

# 恢复Kubernetes配置
kubectl apply -f backup-k8s-config.yaml
```

## 安全最佳实践

### 1. 容器安全

#### 安全镜像构建
```dockerfile
# 使用最小化基础镜像
FROM alpine:3.18

# 创建非root用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# 安装必要的安全更新
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*

# 设置工作目录
WORKDIR /app

# 复制应用文件
COPY --chown=appuser:appgroup ./app /app/

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./app"]
```

#### Pod安全策略
```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    fsGroup: 1001
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: ai-service
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
    volumeMounts:
    - name: tmp
      mountPath: /tmp
    - name: var-run
      mountPath: /var/run
  volumes:
  - name: tmp
    emptyDir: {}
  - name: var-run
    emptyDir: {}
```

### 2. 网络安全

#### 网络策略
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: ai-service-network-policy
spec:
  podSelector:
    matchLabels:
      app: advanced-ai-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

### 3. 密钥管理

#### 使用外部密钥管理
```yaml
# 使用AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: ai-service-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: ai-service-secrets
    creationPolicy: Owner
  data:
  - secretKey: openai-api-key
    remoteRef:
      key: prod/ai-service/openai-api-key
  - secretKey: database-password
    remoteRef:
      key: prod/ai-service/database-password
```

## 总结

本部署指南提供了从开发环境到生产环境的完整部署流程，包括：

1. **环境准备**: 硬件要求、软件依赖、GPU支持
2. **本地开发**: 快速启动开发环境
3. **容器化部署**: Docker和Docker Compose部署
4. **Kubernetes部署**: 完整的K8s部署方案
5. **生产环境**: 高可用、安全的生产部署
6. **监控运维**: 完整的监控和日志体系
7. **故障排除**: 常见问题的诊断和解决
8. **性能优化**: 多层次的性能优化策略
9. **安全实践**: 全面的安全防护措施

通过遵循本指南，您可以成功部署和运维太上老君项目的高级AI功能，确保系统的稳定性、安全性和高性能。