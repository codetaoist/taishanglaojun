#!/bin/bash

# 太上老君AI平台 - 高级AI功能部署脚本
# 版本: 1.0.0
# 作者: 太上老君AI团队

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
PROJECT_NAME="taishanglaojun-advanced-ai"
SERVICE_NAME="advanced-ai-service"
DOCKER_IMAGE="taishanglaojun/advanced-ai"
DOCKER_TAG="latest"
NAMESPACE="taishanglaojun"
CONFIG_DIR="./config"
SCRIPTS_DIR="./scripts"
DOCS_DIR="./docs"

# 环境变量
ENVIRONMENT=${ENVIRONMENT:-"production"}
DEPLOY_MODE=${DEPLOY_MODE:-"docker"}  # docker, kubernetes, standalone
BUILD_MODE=${BUILD_MODE:-"release"}   # debug, release
ENABLE_MONITORING=${ENABLE_MONITORING:-"true"}
ENABLE_LOGGING=${ENABLE_LOGGING:-"true"}

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 错误处理
error_exit() {
    log_error "$1"
    exit 1
}

# 检查依赖
check_dependencies() {
    log_info "检查部署依赖..."
    
    # 检查必需的命令
    local required_commands=("go" "docker")
    
    if [[ "$DEPLOY_MODE" == "kubernetes" ]]; then
        required_commands+=("kubectl" "helm")
    fi
    
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            error_exit "缺少必需的命令: $cmd"
        fi
    done
    
    # 检查Go版本
    local go_version=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
    local min_go_version="1.19"
    
    if [[ "$(printf '%s\n' "$min_go_version" "$go_version" | sort -V | head -n1)" != "$min_go_version" ]]; then
        error_exit "Go版本过低，需要 >= $min_go_version，当前版本: $go_version"
    fi
    
    # 检查Docker版本
    if ! docker --version &> /dev/null; then
        error_exit "Docker未安装或无法访问"
    fi
    
    log_success "依赖检查完成"
}

# 构建应用
build_application() {
    log_info "构建应用程序..."
    
    # 设置构建环境
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    # 构建标志
    local build_flags="-trimpath"
    local ldflags="-s -w"
    
    if [[ "$BUILD_MODE" == "debug" ]]; then
        build_flags=""
        ldflags=""
    fi
    
    # 构建主服务
    log_info "构建主服务..."
    go build $build_flags -ldflags "$ldflags" -o bin/advanced-ai-service ./cmd/advanced-ai/
    
    # 构建工具
    log_info "构建管理工具..."
    if [[ -d "./cmd/tools" ]]; then
        go build $build_flags -ldflags "$ldflags" -o bin/ai-admin ./cmd/tools/admin/
        go build $build_flags -ldflags "$ldflags" -o bin/ai-monitor ./cmd/tools/monitor/
    fi
    
    log_success "应用构建完成"
}

# 运行测试
run_tests() {
    log_info "运行测试套件..."
    
    # 单元测试
    log_info "运行单元测试..."
    go test -v -race -coverprofile=coverage.out ./...
    
    # 集成测试
    if [[ -f "./tests/integration_test.go" ]]; then
        log_info "运行集成测试..."
        go test -v -tags=integration ./tests/
    fi
    
    # 性能测试
    if [[ "$BUILD_MODE" == "release" ]]; then
        log_info "运行性能测试..."
        go test -v -bench=. -benchmem ./tests/
    fi
    
    # 生成测试报告
    if command -v go-junit-report &> /dev/null; then
        go test -v ./... 2>&1 | go-junit-report > test-report.xml
    fi
    
    log_success "测试完成"
}

# 构建Docker镜像
build_docker_image() {
    log_info "构建Docker镜像..."
    
    # 创建Dockerfile
    cat > Dockerfile << EOF
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o advanced-ai-service ./cmd/advanced-ai/

# 运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# 设置工作目录
WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/advanced-ai-service .
COPY --from=builder /app/config ./config
COPY --from=builder /app/docs ./docs

# 设置权限
RUN chown -R appuser:appgroup /app
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/advanced-ai/health || exit 1

# 启动命令
CMD ["./advanced-ai-service"]
EOF
    
    # 构建镜像
    docker build -t "${DOCKER_IMAGE}:${DOCKER_TAG}" .
    
    # 标记镜像
    if [[ "$ENVIRONMENT" != "development" ]]; then
        docker tag "${DOCKER_IMAGE}:${DOCKER_TAG}" "${DOCKER_IMAGE}:${ENVIRONMENT}"
    fi
    
    log_success "Docker镜像构建完成"
}

# Docker部署
deploy_docker() {
    log_info "使用Docker部署..."
    
    # 停止现有容器
    if docker ps -q -f name="$SERVICE_NAME" | grep -q .; then
        log_info "停止现有容器..."
        docker stop "$SERVICE_NAME" || true
        docker rm "$SERVICE_NAME" || true
    fi
    
    # 创建网络
    if ! docker network ls | grep -q taishanglaojun-network; then
        docker network create taishanglaojun-network
    fi
    
    # 启动依赖服务
    deploy_dependencies_docker
    
    # 启动主服务
    log_info "启动主服务容器..."
    docker run -d \
        --name "$SERVICE_NAME" \
        --network taishanglaojun-network \
        -p 8080:8080 \
        -e ENVIRONMENT="$ENVIRONMENT" \
        -e ENABLE_MONITORING="$ENABLE_MONITORING" \
        -e ENABLE_LOGGING="$ENABLE_LOGGING" \
        -v "$(pwd)/config:/app/config:ro" \
        -v "$(pwd)/logs:/app/logs" \
        --restart unless-stopped \
        --memory 4g \
        --cpus 2 \
        "${DOCKER_IMAGE}:${DOCKER_TAG}"
    
    # 等待服务启动
    wait_for_service "http://localhost:8080/api/v1/advanced-ai/health"
    
    log_success "Docker部署完成"
}

# 部署依赖服务
deploy_dependencies_docker() {
    log_info "部署依赖服务..."
    
    # Redis
    if ! docker ps -q -f name="redis-taishanglaojun" | grep -q .; then
        log_info "启动Redis..."
        docker run -d \
            --name redis-taishanglaojun \
            --network taishanglaojun-network \
            -p 6379:6379 \
            -v redis-data:/data \
            --restart unless-stopped \
            redis:7-alpine redis-server --appendonly yes
    fi
    
    # PostgreSQL
    if ! docker ps -q -f name="postgres-taishanglaojun" | grep -q .; then
        log_info "启动PostgreSQL..."
        docker run -d \
            --name postgres-taishanglaojun \
            --network taishanglaojun-network \
            -p 5432:5432 \
            -e POSTGRES_DB=taishanglaojun \
            -e POSTGRES_USER=taishanglaojun \
            -e POSTGRES_PASSWORD=taishanglaojun123 \
            -v postgres-data:/var/lib/postgresql/data \
            --restart unless-stopped \
            postgres:15-alpine
    fi
    
    # Prometheus (如果启用监控)
    if [[ "$ENABLE_MONITORING" == "true" ]] && ! docker ps -q -f name="prometheus-taishanglaojun" | grep -q .; then
        log_info "启动Prometheus..."
        
        # 创建Prometheus配置
        mkdir -p monitoring/prometheus
        cat > monitoring/prometheus/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'advanced-ai-service'
    static_configs:
      - targets: ['${SERVICE_NAME}:8080']
    metrics_path: '/api/v1/advanced-ai/metrics'
    scrape_interval: 10s
EOF
        
        docker run -d \
            --name prometheus-taishanglaojun \
            --network taishanglaojun-network \
            -p 9090:9090 \
            -v "$(pwd)/monitoring/prometheus:/etc/prometheus:ro" \
            --restart unless-stopped \
            prom/prometheus:latest
    fi
    
    # Grafana (如果启用监控)
    if [[ "$ENABLE_MONITORING" == "true" ]] && ! docker ps -q -f name="grafana-taishanglaojun" | grep -q .; then
        log_info "启动Grafana..."
        docker run -d \
            --name grafana-taishanglaojun \
            --network taishanglaojun-network \
            -p 3000:3000 \
            -e GF_SECURITY_ADMIN_PASSWORD=admin123 \
            -v grafana-data:/var/lib/grafana \
            --restart unless-stopped \
            grafana/grafana:latest
    fi
    
    log_success "依赖服务部署完成"
}

# Kubernetes部署
deploy_kubernetes() {
    log_info "使用Kubernetes部署..."
    
    # 创建命名空间
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # 创建ConfigMap
    create_kubernetes_configmap
    
    # 创建Secret
    create_kubernetes_secret
    
    # 部署依赖服务
    deploy_dependencies_kubernetes
    
    # 部署主服务
    deploy_main_service_kubernetes
    
    # 创建Service和Ingress
    create_kubernetes_service
    
    # 等待部署完成
    kubectl rollout status deployment/"$SERVICE_NAME" -n "$NAMESPACE" --timeout=300s
    
    log_success "Kubernetes部署完成"
}

# 创建Kubernetes ConfigMap
create_kubernetes_configmap() {
    log_info "创建ConfigMap..."
    
    kubectl create configmap advanced-ai-config \
        --from-file="$CONFIG_DIR" \
        --namespace="$NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
}

# 创建Kubernetes Secret
create_kubernetes_secret() {
    log_info "创建Secret..."
    
    kubectl create secret generic advanced-ai-secret \
        --from-literal=jwt-secret="$(openssl rand -base64 32)" \
        --from-literal=db-password="taishanglaojun123" \
        --from-literal=redis-password="" \
        --namespace="$NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
}

# 部署Kubernetes依赖服务
deploy_dependencies_kubernetes() {
    log_info "部署Kubernetes依赖服务..."
    
    # Redis部署
    cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: $NAMESPACE
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: $NAMESPACE
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
EOF
    
    # PostgreSQL部署
    cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: $NAMESPACE
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: taishanglaojun
        - name: POSTGRES_USER
          value: taishanglaojun
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: advanced-ai-secret
              key: db-password
        ports:
        - containerPort: 5432
        resources:
          requests:
            memory: "512Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: $NAMESPACE
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: $NAMESPACE
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
EOF
}

# 部署Kubernetes主服务
deploy_main_service_kubernetes() {
    log_info "部署主服务..."
    
    cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $SERVICE_NAME
  namespace: $NAMESPACE
  labels:
    app: $SERVICE_NAME
    version: $DOCKER_TAG
spec:
  replicas: 3
  selector:
    matchLabels:
      app: $SERVICE_NAME
  template:
    metadata:
      labels:
        app: $SERVICE_NAME
        version: $DOCKER_TAG
    spec:
      containers:
      - name: $SERVICE_NAME
        image: ${DOCKER_IMAGE}:${DOCKER_TAG}
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: ENVIRONMENT
          value: "$ENVIRONMENT"
        - name: ENABLE_MONITORING
          value: "$ENABLE_MONITORING"
        - name: ENABLE_LOGGING
          value: "$ENABLE_LOGGING"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: advanced-ai-secret
              key: jwt-secret
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: advanced-ai-secret
              key: db-password
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "2"
        livenessProbe:
          httpGet:
            path: /api/v1/advanced-ai/health
            port: 8080
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /api/v1/advanced-ai/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        volumeMounts:
        - name: config-volume
          mountPath: /app/config
          readOnly: true
        - name: logs-volume
          mountPath: /app/logs
      volumes:
      - name: config-volume
        configMap:
          name: advanced-ai-config
      - name: logs-volume
        emptyDir: {}
      imagePullSecrets:
      - name: docker-registry-secret
EOF
}

# 创建Kubernetes Service
create_kubernetes_service() {
    log_info "创建Service和Ingress..."
    
    # Service
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: $SERVICE_NAME
  namespace: $NAMESPACE
  labels:
    app: $SERVICE_NAME
spec:
  selector:
    app: $SERVICE_NAME
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
EOF
    
    # Ingress
    cat << EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: $SERVICE_NAME
  namespace: $NAMESPACE
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
spec:
  tls:
  - hosts:
    - ai.taishanglaojun.com
    secretName: tls-secret
  rules:
  - host: ai.taishanglaojun.com
    http:
      paths:
      - path: /api/v1/advanced-ai
        pathType: Prefix
        backend:
          service:
            name: $SERVICE_NAME
            port:
              number: 80
EOF
}

# 独立部署
deploy_standalone() {
    log_info "独立部署模式..."
    
    # 创建目录结构
    mkdir -p /opt/taishanglaojun/{bin,config,logs,data}
    
    # 复制文件
    cp bin/advanced-ai-service /opt/taishanglaojun/bin/
    cp -r config/* /opt/taishanglaojun/config/
    cp -r docs /opt/taishanglaojun/
    
    # 创建systemd服务
    create_systemd_service
    
    # 启动服务
    sudo systemctl daemon-reload
    sudo systemctl enable advanced-ai-service
    sudo systemctl start advanced-ai-service
    
    # 等待服务启动
    wait_for_service "http://localhost:8080/api/v1/advanced-ai/health"
    
    log_success "独立部署完成"
}

# 创建systemd服务
create_systemd_service() {
    log_info "创建systemd服务..."
    
    sudo tee /etc/systemd/system/advanced-ai-service.service > /dev/null << EOF
[Unit]
Description=太上老君AI平台 - 高级AI功能服务
After=network.target
Wants=network.target

[Service]
Type=simple
User=taishanglaojun
Group=taishanglaojun
WorkingDirectory=/opt/taishanglaojun
ExecStart=/opt/taishanglaojun/bin/advanced-ai-service
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
Environment=ENVIRONMENT=$ENVIRONMENT
Environment=ENABLE_MONITORING=$ENABLE_MONITORING
Environment=ENABLE_LOGGING=$ENABLE_LOGGING

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/taishanglaojun/logs /opt/taishanglaojun/data

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
    
    # 创建用户
    if ! id "taishanglaojun" &>/dev/null; then
        sudo useradd -r -s /bin/false taishanglaojun
    fi
    
    # 设置权限
    sudo chown -R taishanglaojun:taishanglaojun /opt/taishanglaojun
    sudo chmod +x /opt/taishanglaojun/bin/advanced-ai-service
}

# 等待服务启动
wait_for_service() {
    local url="$1"
    local max_attempts=30
    local attempt=1
    
    log_info "等待服务启动..."
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -f -s "$url" > /dev/null 2>&1; then
            log_success "服务已启动"
            return 0
        fi
        
        log_info "等待服务启动... ($attempt/$max_attempts)"
        sleep 10
        ((attempt++))
    done
    
    error_exit "服务启动超时"
}

# 验证部署
verify_deployment() {
    log_info "验证部署..."
    
    local base_url
    case "$DEPLOY_MODE" in
        "docker")
            base_url="http://localhost:8080/api/v1/advanced-ai"
            ;;
        "kubernetes")
            base_url="http://ai.taishanglaojun.com/api/v1/advanced-ai"
            ;;
        "standalone")
            base_url="http://localhost:8080/api/v1/advanced-ai"
            ;;
    esac
    
    # 健康检查
    log_info "检查服务健康状态..."
    local health_response=$(curl -s "$base_url/health")
    if [[ $(echo "$health_response" | jq -r '.status') != "healthy" ]]; then
        error_exit "服务健康检查失败"
    fi
    
    # 功能测试
    log_info "测试基本功能..."
    local test_request='{"type":"reasoning","input":{"problem":"测试问题"}}'
    local test_response=$(curl -s -X POST -H "Content-Type: application/json" -d "$test_request" "$base_url/process")
    if [[ $(echo "$test_response" | jq -r '.success') != "true" ]]; then
        error_exit "功能测试失败"
    fi
    
    # 性能测试
    log_info "性能测试..."
    local response_time=$(curl -w "%{time_total}" -s -o /dev/null "$base_url/status")
    if (( $(echo "$response_time > 5.0" | bc -l) )); then
        log_warning "响应时间较慢: ${response_time}s"
    fi
    
    log_success "部署验证完成"
}

# 清理资源
cleanup() {
    log_info "清理临时资源..."
    
    # 清理Docker资源
    if [[ "$DEPLOY_MODE" == "docker" ]]; then
        docker system prune -f
    fi
    
    # 清理构建文件
    rm -f Dockerfile coverage.out test-report.xml
    rm -rf bin/
    
    log_success "清理完成"
}

# 显示部署信息
show_deployment_info() {
    log_success "部署完成！"
    echo
    echo "=== 部署信息 ==="
    echo "项目名称: $PROJECT_NAME"
    echo "服务名称: $SERVICE_NAME"
    echo "部署模式: $DEPLOY_MODE"
    echo "环境: $ENVIRONMENT"
    echo "版本: $DOCKER_TAG"
    echo
    
    case "$DEPLOY_MODE" in
        "docker")
            echo "=== 访问信息 ==="
            echo "服务地址: http://localhost:8080"
            echo "API文档: http://localhost:8080/api/v1/advanced-ai/docs"
            echo "健康检查: http://localhost:8080/api/v1/advanced-ai/health"
            echo "监控面板: http://localhost:3000 (admin/admin123)"
            echo
            echo "=== 管理命令 ==="
            echo "查看日志: docker logs $SERVICE_NAME"
            echo "重启服务: docker restart $SERVICE_NAME"
            echo "停止服务: docker stop $SERVICE_NAME"
            ;;
        "kubernetes")
            echo "=== 访问信息 ==="
            echo "服务地址: http://ai.taishanglaojun.com"
            echo "API文档: http://ai.taishanglaojun.com/api/v1/advanced-ai/docs"
            echo
            echo "=== 管理命令 ==="
            echo "查看Pod: kubectl get pods -n $NAMESPACE"
            echo "查看日志: kubectl logs -f deployment/$SERVICE_NAME -n $NAMESPACE"
            echo "扩容: kubectl scale deployment $SERVICE_NAME --replicas=5 -n $NAMESPACE"
            ;;
        "standalone")
            echo "=== 访问信息 ==="
            echo "服务地址: http://localhost:8080"
            echo "API文档: http://localhost:8080/api/v1/advanced-ai/docs"
            echo
            echo "=== 管理命令 ==="
            echo "查看状态: sudo systemctl status advanced-ai-service"
            echo "查看日志: sudo journalctl -u advanced-ai-service -f"
            echo "重启服务: sudo systemctl restart advanced-ai-service"
            ;;
    esac
    
    echo
    echo "=== 下一步 ==="
    echo "1. 访问API文档了解接口使用方法"
    echo "2. 配置监控和告警"
    echo "3. 设置备份策略"
    echo "4. 进行性能调优"
}

# 主函数
main() {
    log_info "开始部署太上老君AI平台 - 高级AI功能"
    echo "部署模式: $DEPLOY_MODE"
    echo "环境: $ENVIRONMENT"
    echo "构建模式: $BUILD_MODE"
    echo
    
    # 检查依赖
    check_dependencies
    
    # 构建应用
    build_application
    
    # 运行测试
    if [[ "$BUILD_MODE" == "release" ]]; then
        run_tests
    fi
    
    # 根据部署模式执行部署
    case "$DEPLOY_MODE" in
        "docker")
            build_docker_image
            deploy_docker
            ;;
        "kubernetes")
            build_docker_image
            deploy_kubernetes
            ;;
        "standalone")
            deploy_standalone
            ;;
        *)
            error_exit "不支持的部署模式: $DEPLOY_MODE"
            ;;
    esac
    
    # 验证部署
    verify_deployment
    
    # 清理资源
    cleanup
    
    # 显示部署信息
    show_deployment_info
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            DEPLOY_MODE="$2"
            shift 2
            ;;
        --env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --build)
            BUILD_MODE="$2"
            shift 2
            ;;
        --tag)
            DOCKER_TAG="$2"
            shift 2
            ;;
        --no-monitoring)
            ENABLE_MONITORING="false"
            shift
            ;;
        --no-logging)
            ENABLE_LOGGING="false"
            shift
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo "选项:"
            echo "  --mode MODE        部署模式 (docker|kubernetes|standalone)"
            echo "  --env ENV          环境 (development|staging|production)"
            echo "  --build BUILD      构建模式 (debug|release)"
            echo "  --tag TAG          Docker标签"
            echo "  --no-monitoring    禁用监控"
            echo "  --no-logging       禁用日志"
            echo "  --help             显示帮助信息"
            exit 0
            ;;
        *)
            error_exit "未知参数: $1"
            ;;
    esac
done

# 执行主函数
main