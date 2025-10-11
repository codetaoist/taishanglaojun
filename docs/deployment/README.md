# 太上老君AI平台 部署指南

## 概述

本文档详细介绍了太上老君AI平台的部署流程，包括本地开发环境搭建、生产环境部署、监控配置等内容。

## 系统要求

### 硬件要求

**最低配置**:
- CPU: 4核心
- 内存: 8GB RAM
- 存储: 100GB SSD
- 网络: 100Mbps

**推荐配置**:
- CPU: 8核心以上
- 内存: 16GB RAM以上
- 存储: 500GB SSD以上
- 网络: 1Gbps以上

### 软件要求

- **操作系统**: Ubuntu 20.04+ / CentOS 8+ / macOS 12+
- **容器运行时**: Docker 20.10+ / containerd 1.6+
- **编排工具**: Kubernetes 1.24+
- **数据库**: PostgreSQL 14+ / Redis 7+
- **负载均衡**: Nginx 1.20+ / HAProxy 2.4+

## 架构概览

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户设备      │    │   CDN/WAF       │    │   负载均衡      │
│                 │────│                 │────│                 │
│ Web/Mobile App  │    │   CloudFlare    │    │   Nginx/ALB     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                       ┌────────────────────────────────┼────────────────────────────────┐
                       │                                │                                │
                ┌─────────────────┐            ┌─────────────────┐            ┌─────────────────┐
                │   前端服务      │            │   API网关       │            │   后端服务      │
                │                 │            │                 │            │                 │
                │   React App     │            │   Kong/Envoy    │            │   Go Services   │
                └─────────────────┘            └─────────────────┘            └─────────────────┘
                                                        │
                       ┌────────────────────────────────┼────────────────────────────────┐
                       │                                │                                │
                ┌─────────────────┐            ┌─────────────────┐            ┌─────────────────┐
                │   数据库        │            │   缓存          │            │   存储          │
                │                 │            │                 │            │                 │
                │   PostgreSQL    │            │   Redis         │            │   S3/MinIO      │
                └─────────────────┘            └─────────────────┘            └─────────────────┘
```

## 本地开发环境

### 1. 环境准备

#### 安装必要工具

```bash
# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.15.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装 kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# 安装 Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# 安装 Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 安装 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

#### 克隆项目

```bash
git clone https://github.com/taishanglaojun/platform.git
cd platform
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env.local

# 编辑环境变量
vim .env.local
```

**环境变量示例**:
```bash
# 应用配置
ENVIRONMENT=development
PORT=8080
DOMAIN=localhost

# 数据库配置
DATABASE_URL=postgres://postgres:password@localhost:5432/taishanglaojun
REDIS_URL=redis://localhost:6379

# AI服务配置
OPENAI_API_KEY=sk-your-openai-api-key
OPENAI_BASE_URL=https://api.openai.com/v1

# 认证配置
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRES_IN=24h

# 存储配置
S3_BUCKET=taishanglaojun-assets
S3_REGION=us-west-2
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# 监控配置
SENTRY_DSN=https://your-sentry-dsn
PROMETHEUS_ENABLED=true

# 支付配置
STRIPE_SECRET_KEY=sk_test_your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=whsec_your-webhook-secret
```

### 3. 启动开发环境

#### 使用 Docker Compose

```bash
# 启动基础服务（数据库、缓存等）
docker-compose -f docker-compose.dev.yml up -d postgres redis minio

# 等待服务启动
sleep 30

# 运行数据库迁移
make migrate-up

# 启动后端服务
make dev-backend

# 启动前端服务（新终端）
make dev-frontend

# 启动移动端开发服务（可选）
make dev-mobile
```

#### 手动启动

```bash
# 启动数据库
docker run -d --name postgres \
  -e POSTGRES_DB=taishanglaojun \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:15

# 启动Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:7-alpine

# 启动后端服务
cd core-services
go mod download
go run cmd/api/main.go

# 启动前端服务
cd frontend/web-app
npm install
npm run dev

# 启动移动端开发服务
cd mobile-app
npm install
npx expo start
```

### 4. 验证安装

```bash
# 检查后端服务
curl http://localhost:8080/health

# 检查前端服务
curl http://localhost:3000

# 检查数据库连接
psql -h localhost -U postgres -d taishanglaojun -c "SELECT version();"

# 检查Redis连接
redis-cli ping
```

## 生产环境部署

### 1. 基础设施准备

#### AWS 基础设施

```bash
# 进入Terraform目录
cd infrastructure/terraform

# 初始化Terraform
terraform init

# 创建terraform.tfvars文件
cat > terraform.tfvars << EOF
environment = "production"
aws_region = "us-west-2"
domain_name = "taishanglaojun.ai"

# VPC配置
vpc_cidr = "10.0.0.0/16"
availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

# EKS配置
kubernetes_version = "1.28"
node_instance_types = ["t3.medium", "t3.large"]
node_desired_capacity = 3
node_max_capacity = 10
node_min_capacity = 1

# RDS配置
db_instance_class = "db.t3.micro"
db_allocated_storage = 20
db_max_allocated_storage = 100

# Redis配置
redis_node_type = "cache.t3.micro"
redis_num_cache_nodes = 1

# 应用密钥
openai_api_key = "sk-your-openai-api-key"
jwt_secret = "your-jwt-secret"
stripe_secret_key = "sk_live_your-stripe-secret"
EOF

# 规划部署
terraform plan

# 执行部署
terraform apply
```

#### 获取基础设施信息

```bash
# 获取EKS集群配置
aws eks update-kubeconfig --region us-west-2 --name taishanglaojun-production

# 验证集群连接
kubectl get nodes

# 获取数据库连接信息
terraform output database_connection_string

# 获取Redis连接信息
terraform output redis_connection_string
```

### 2. 应用部署

#### 使用 Helm 部署

```bash
# 添加必要的Helm仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add jetstack https://charts.jetstack.io
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# 创建命名空间
kubectl create namespace taishanglaojun-production

# 安装cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.13.2 \
  --set installCRDs=true

# 安装ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer

# 创建生产环境配置
cat > values-production.yaml << EOF
environment: production
domain: taishanglaojun.ai

image:
  registry: your-registry.com
  tag: "v1.0.0"

frontend:
  replicaCount: 3
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi

backend:
  replicaCount: 5
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 200m
      memory: 256Mi

externalServices:
  database:
    host: "your-rds-endpoint"
    port: 5432
    name: "taishanglaojun"
    username: "postgres"
    password: "your-db-password"
  
  redis:
    host: "your-redis-endpoint"
    port: 6379
    password: "your-redis-password"

secrets:
  data:
    database-url: "postgres://postgres:password@your-rds-endpoint:5432/taishanglaojun"
    redis-url: "redis://:password@your-redis-endpoint:6379"
    jwt-secret: "your-jwt-secret"
    openai-api-key: "sk-your-openai-api-key"
    stripe-api-key: "sk_live_your-stripe-key"

monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
EOF

# 部署应用
helm install taishanglaojun ./infrastructure/helm/taishang-laojun \
  --namespace taishanglaojun-production \
  --values values-production.yaml

# 等待部署完成
kubectl wait --for=condition=available --timeout=600s deployment --all -n taishanglaojun-production
```

#### 使用部署脚本

```bash
# 使用自动化部署脚本
chmod +x scripts/deploy.sh

# 部署到生产环境
./scripts/deploy.sh deploy production

# 检查部署状态
./scripts/deploy.sh health production
```

### 3. 域名和SSL配置

#### 配置DNS

```bash
# 获取LoadBalancer IP
kubectl get svc ingress-nginx-controller -n ingress-nginx

# 配置DNS记录（在你的DNS提供商处）
# A记录: taishanglaojun.ai -> LoadBalancer-IP
# A记录: api.taishanglaojun.ai -> LoadBalancer-IP
# A记录: *.taishanglaojun.ai -> LoadBalancer-IP
```

#### 配置SSL证书

```bash
# 创建ClusterIssuer
cat > cluster-issuer.yaml << EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@taishanglaojun.ai
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

kubectl apply -f cluster-issuer.yaml
```

### 4. 监控和日志

#### 部署监控栈

```bash
# 安装Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values monitoring/prometheus/values-production.yaml

# 安装Grafana仪表板
kubectl create configmap grafana-dashboards \
  --from-file=monitoring/grafana/dashboards/ \
  -n monitoring

# 配置Alertmanager
kubectl apply -f monitoring/alertmanager/config.yaml
```

#### 配置日志收集

```bash
# 安装Fluent Bit
helm install fluent-bit fluent/fluent-bit \
  --namespace logging \
  --create-namespace \
  --values logging/fluent-bit/values.yaml

# 安装Elasticsearch
helm install elasticsearch elastic/elasticsearch \
  --namespace logging \
  --values logging/elasticsearch/values.yaml

# 安装Kibana
helm install kibana elastic/kibana \
  --namespace logging \
  --values logging/kibana/values.yaml
```

## 运维管理

### 1. 健康检查

```bash
# 检查所有Pod状态
kubectl get pods -n taishanglaojun-production

# 检查服务状态
kubectl get svc -n taishanglaojun-production

# 检查Ingress状态
kubectl get ingress -n taishanglaojun-production

# 检查证书状态
kubectl get certificates -n taishanglaojun-production

# 应用健康检查
curl https://taishanglaojun.ai/health
curl https://api.taishanglaojun.ai/health
```

### 2. 扩容和缩容

```bash
# 手动扩容后端服务
kubectl scale deployment taishanglaojun-backend --replicas=10 -n taishanglaojun-production

# 手动扩容前端服务
kubectl scale deployment taishanglaojun-frontend --replicas=5 -n taishanglaojun-production

# 配置HPA自动扩容
kubectl apply -f k8s/production/hpa.yaml
```

### 3. 更新部署

```bash
# 滚动更新
helm upgrade taishanglaojun ./infrastructure/helm/taishang-laojun \
  --namespace taishanglaojun-production \
  --values values-production.yaml \
  --set image.tag=v1.1.0

# 回滚到上一个版本
helm rollback taishanglaojun 1 -n taishanglaojun-production

# 查看发布历史
helm history taishanglaojun -n taishanglaojun-production
```

### 4. 备份和恢复

#### 数据库备份

```bash
# 创建数据库备份
kubectl create job --from=cronjob/postgres-backup postgres-backup-manual -n taishanglaojun-production

# 手动备份
pg_dump -h your-rds-endpoint -U postgres -d taishanglaojun > backup-$(date +%Y%m%d).sql

# 恢复数据库
psql -h your-rds-endpoint -U postgres -d taishanglaojun < backup-20240101.sql
```

#### 应用配置备份

```bash
# 备份Kubernetes配置
kubectl get all -n taishanglaojun-production -o yaml > k8s-backup-$(date +%Y%m%d).yaml

# 备份Helm配置
helm get values taishanglaojun -n taishanglaojun-production > helm-values-backup-$(date +%Y%m%d).yaml
```

## 故障排除

### 1. 常见问题

#### Pod启动失败

```bash
# 查看Pod状态
kubectl describe pod <pod-name> -n taishanglaojun-production

# 查看Pod日志
kubectl logs <pod-name> -n taishanglaojun-production

# 查看事件
kubectl get events -n taishanglaojun-production --sort-by='.lastTimestamp'
```

#### 服务无法访问

```bash
# 检查Service
kubectl get svc -n taishanglaojun-production
kubectl describe svc <service-name> -n taishanglaojun-production

# 检查Ingress
kubectl describe ingress <ingress-name> -n taishanglaojun-production

# 检查DNS解析
nslookup taishanglaojun.ai
dig taishanglaojun.ai
```

#### 数据库连接问题

```bash
# 测试数据库连接
kubectl run postgres-client --rm -it --image=postgres:15 -- psql -h your-rds-endpoint -U postgres -d taishanglaojun

# 检查网络策略
kubectl get networkpolicies -n taishanglaojun-production

# 检查安全组配置
aws ec2 describe-security-groups --group-ids sg-xxxxxxxxx
```

### 2. 性能优化

#### 资源优化

```bash
# 查看资源使用情况
kubectl top nodes
kubectl top pods -n taishanglaojun-production

# 优化资源限制
kubectl patch deployment taishanglaojun-backend -n taishanglaojun-production -p '{"spec":{"template":{"spec":{"containers":[{"name":"backend","resources":{"limits":{"cpu":"2000m","memory":"2Gi"},"requests":{"cpu":"500m","memory":"512Mi"}}}]}}}}'
```

#### 数据库优化

```bash
# 查看数据库性能
psql -h your-rds-endpoint -U postgres -d taishanglaojun -c "SELECT * FROM pg_stat_activity;"

# 优化查询
psql -h your-rds-endpoint -U postgres -d taishanglaojun -c "EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user@example.com';"
```

## 安全配置

### 1. 网络安全

```bash
# 配置网络策略
kubectl apply -f k8s/security/network-policies.yaml

# 配置Pod安全策略
kubectl apply -f k8s/security/pod-security-policies.yaml

# 配置RBAC
kubectl apply -f k8s/security/rbac.yaml
```

### 2. 密钥管理

```bash
# 使用Kubernetes Secrets
kubectl create secret generic app-secrets \
  --from-literal=database-url="postgres://..." \
  --from-literal=jwt-secret="..." \
  -n taishanglaojun-production

# 使用AWS Secrets Manager
aws secretsmanager create-secret \
  --name taishanglaojun/production/app-secrets \
  --secret-string '{"database_url":"postgres://...","jwt_secret":"..."}'
```

### 3. 镜像安全

```bash
# 扫描镜像漏洞
trivy image your-registry.com/taishanglaojun/backend:v1.0.0

# 签名镜像
cosign sign your-registry.com/taishanglaojun/backend:v1.0.0

# 验证镜像签名
cosign verify your-registry.com/taishanglaojun/backend:v1.0.0
```

## 监控和告警

### 1. 关键指标

- **应用指标**: 请求量、响应时间、错误率
- **基础设施指标**: CPU、内存、磁盘、网络
- **业务指标**: 用户注册、对话数量、支付成功率

### 2. 告警规则

```yaml
# 高错误率告警
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High error rate detected"

# 高延迟告警
- alert: HighLatency
  expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High latency detected"
```

## 联系支持

如果在部署过程中遇到问题，请联系：

- **技术支持**: support@taishanglaojun.ai
- **文档**: https://docs.taishanglaojun.ai
- **GitHub Issues**: https://github.com/taishanglaojun/platform/issues
- **Slack社区**: https://taishanglaojun.slack.com