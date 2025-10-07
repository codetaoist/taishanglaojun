#!/bin/bash

# 太上老君监控系统部署脚本
# 用于自动化部署到不同环境

set -euo pipefail

# ==================== 变量定义 ====================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_NAME="taishanglaojun-monitoring"

# 默认配置
DEFAULT_ENVIRONMENT="development"
DEFAULT_NAMESPACE="monitoring"
DEFAULT_REGISTRY="ghcr.io"
DEFAULT_IMAGE_NAME="taishanglaojun/monitoring"

# 环境变量
ENVIRONMENT="${ENVIRONMENT:-$DEFAULT_ENVIRONMENT}"
NAMESPACE="${NAMESPACE:-$DEFAULT_NAMESPACE}"
REGISTRY="${REGISTRY:-$DEFAULT_REGISTRY}"
IMAGE_NAME="${IMAGE_NAME:-$DEFAULT_IMAGE_NAME}"
VERSION="${VERSION:-latest}"
KUBECONFIG="${KUBECONFIG:-}"
HELM_CHART_PATH="${HELM_CHART_PATH:-${PROJECT_ROOT}/helm/monitoring}"

# 部署配置
TIMEOUT="${TIMEOUT:-600s}"
WAIT="${WAIT:-true}"
DRY_RUN="${DRY_RUN:-false}"
FORCE="${FORCE:-false}"
SKIP_TESTS="${SKIP_TESTS:-false}"

# 颜色定义
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
BLUE='\033[34m'
MAGENTA='\033[35m'
CYAN='\033[36m'
WHITE='\033[37m'
RESET='\033[0m'

# ==================== 工具函数 ====================
log_info() {
    echo -e "${BLUE}[INFO]${RESET} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${RESET} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${RESET} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${RESET} $*"
}

log_step() {
    echo -e "${CYAN}[STEP]${RESET} $*"
}

check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "命令 '$1' 未找到，请先安装"
        exit 1
    fi
}

confirm_action() {
    local message="$1"
    local default="${2:-n}"
    
    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi
    
    echo -e "${YELLOW}$message${RESET}"
    read -p "是否继续? [y/N]: " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        return 0
    else
        log_info "操作已取消"
        exit 0
    fi
}

wait_for_rollout() {
    local deployment="$1"
    local namespace="$2"
    local timeout="${3:-600s}"
    
    log_step "等待部署完成: $deployment"
    
    if kubectl rollout status deployment/"$deployment" -n "$namespace" --timeout="$timeout"; then
        log_success "部署完成: $deployment"
    else
        log_error "部署超时或失败: $deployment"
        return 1
    fi
}

check_health() {
    local service_url="$1"
    local max_attempts="${2:-30}"
    local sleep_time="${3:-10}"
    
    log_step "检查服务健康状态: $service_url"
    
    for ((i=1; i<=max_attempts; i++)); do
        if curl -f -s "$service_url/health" > /dev/null 2>&1; then
            log_success "服务健康检查通过"
            return 0
        fi
        
        log_info "健康检查失败 ($i/$max_attempts)，等待 $sleep_time 秒后重试..."
        sleep "$sleep_time"
    done
    
    log_error "服务健康检查失败"
    return 1
}

# ==================== 环境检查 ====================
check_environment() {
    log_step "检查部署环境..."
    
    # 检查必要的命令
    check_command "kubectl"
    check_command "helm"
    check_command "docker"
    
    # 检查 Kubernetes 连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群，请检查 kubeconfig"
        exit 1
    fi
    
    # 检查 Helm Chart
    if [[ ! -d "$HELM_CHART_PATH" ]]; then
        log_error "Helm Chart 目录不存在: $HELM_CHART_PATH"
        exit 1
    fi
    
    # 显示当前上下文
    local current_context
    current_context=$(kubectl config current-context)
    log_info "当前 Kubernetes 上下文: $current_context"
    
    # 检查命名空间
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "命名空间 '$NAMESPACE' 不存在，将自动创建"
    fi
    
    log_success "环境检查通过"
}

# ==================== 镜像管理 ====================
check_image() {
    local image="$1"
    
    log_step "检查镜像: $image"
    
    if docker manifest inspect "$image" &> /dev/null; then
        log_success "镜像存在: $image"
        return 0
    else
        log_error "镜像不存在: $image"
        return 1
    fi
}

pull_image() {
    local image="$1"
    
    log_step "拉取镜像: $image"
    
    if docker pull "$image"; then
        log_success "镜像拉取成功: $image"
    else
        log_error "镜像拉取失败: $image"
        exit 1
    fi
}

# ==================== 配置管理 ====================
generate_values_file() {
    local environment="$1"
    local values_file="${PROJECT_ROOT}/helm/values-${environment}.yaml"
    
    log_step "生成 Helm values 文件: $values_file"
    
    cat > "$values_file" << EOF
# 太上老君监控系统 - $environment 环境配置
environment: $environment

image:
  repository: $REGISTRY/$IMAGE_NAME
  tag: $VERSION
  pullPolicy: Always

replicaCount: $(get_replica_count "$environment")

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: $(get_ingress_enabled "$environment")
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: $(get_service_host "$environment")
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: monitoring-tls
      hosts:
        - $(get_service_host "$environment")

resources:
  limits:
    cpu: $(get_cpu_limit "$environment")
    memory: $(get_memory_limit "$environment")
  requests:
    cpu: $(get_cpu_request "$environment")
    memory: $(get_memory_request "$environment")

autoscaling:
  enabled: $(get_autoscaling_enabled "$environment")
  minReplicas: $(get_min_replicas "$environment")
  maxReplicas: $(get_max_replicas "$environment")
  targetCPUUtilizationPercentage: 70

nodeSelector: {}
tolerations: []
affinity: {}

config:
  database:
    host: postgres-service
    port: 5432
    name: taishanglaojun_monitoring
    user: monitoring_user
    sslMode: $(get_db_ssl_mode "$environment")
  
  redis:
    host: redis-service
    port: 6379
    db: 0
  
  monitoring:
    logLevel: $(get_log_level "$environment")
    debug: $(get_debug_enabled "$environment")
    metricsEnabled: true
    tracingEnabled: true

secrets:
  database:
    password: ""  # 将从 Secret 中读取
  redis:
    password: ""  # 将从 Secret 中读取
  jwt:
    secret: ""    # 将从 Secret 中读取
EOF
    
    log_success "Values 文件生成完成: $values_file"
    echo "$values_file"
}

# 环境特定配置函数
get_replica_count() {
    case "$1" in
        development) echo "1" ;;
        staging) echo "2" ;;
        production) echo "3" ;;
        *) echo "1" ;;
    esac
}

get_ingress_enabled() {
    case "$1" in
        development) echo "true" ;;
        staging) echo "true" ;;
        production) echo "true" ;;
        *) echo "false" ;;
    esac
}

get_service_host() {
    case "$1" in
        development) echo "monitoring-dev.taishanglaojun.com" ;;
        staging) echo "monitoring-staging.taishanglaojun.com" ;;
        production) echo "monitoring.taishanglaojun.com" ;;
        *) echo "monitoring-local.taishanglaojun.com" ;;
    esac
}

get_cpu_limit() {
    case "$1" in
        development) echo "500m" ;;
        staging) echo "1000m" ;;
        production) echo "2000m" ;;
        *) echo "500m" ;;
    esac
}

get_memory_limit() {
    case "$1" in
        development) echo "512Mi" ;;
        staging) echo "1Gi" ;;
        production) echo "2Gi" ;;
        *) echo "512Mi" ;;
    esac
}

get_cpu_request() {
    case "$1" in
        development) echo "100m" ;;
        staging) echo "250m" ;;
        production) echo "500m" ;;
        *) echo "100m" ;;
    esac
}

get_memory_request() {
    case "$1" in
        development) echo "256Mi" ;;
        staging) echo "512Mi" ;;
        production) echo "1Gi" ;;
        *) echo "256Mi" ;;
    esac
}

get_autoscaling_enabled() {
    case "$1" in
        development) echo "false" ;;
        staging) echo "true" ;;
        production) echo "true" ;;
        *) echo "false" ;;
    esac
}

get_min_replicas() {
    case "$1" in
        development) echo "1" ;;
        staging) echo "2" ;;
        production) echo "3" ;;
        *) echo "1" ;;
    esac
}

get_max_replicas() {
    case "$1" in
        development) echo "2" ;;
        staging) echo "5" ;;
        production) echo "10" ;;
        *) echo "2" ;;
    esac
}

get_db_ssl_mode() {
    case "$1" in
        development) echo "disable" ;;
        staging) echo "require" ;;
        production) echo "require" ;;
        *) echo "disable" ;;
    esac
}

get_log_level() {
    case "$1" in
        development) echo "debug" ;;
        staging) echo "info" ;;
        production) echo "info" ;;
        *) echo "debug" ;;
    esac
}

get_debug_enabled() {
    case "$1" in
        development) echo "true" ;;
        staging) echo "false" ;;
        production) echo "false" ;;
        *) echo "true" ;;
    esac
}

# ==================== 密钥管理 ====================
create_secrets() {
    local namespace="$1"
    local environment="$2"
    
    log_step "创建 Kubernetes Secrets..."
    
    # 创建数据库密钥
    kubectl create secret generic monitoring-db-secret \
        --from-literal=password="${DB_PASSWORD:-$(generate_password)}" \
        --namespace="$namespace" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # 创建 Redis 密钥
    kubectl create secret generic monitoring-redis-secret \
        --from-literal=password="${REDIS_PASSWORD:-$(generate_password)}" \
        --namespace="$namespace" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # 创建 JWT 密钥
    kubectl create secret generic monitoring-jwt-secret \
        --from-literal=secret="${JWT_SECRET:-$(generate_jwt_secret)}" \
        --namespace="$namespace" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Secrets 创建完成"
}

generate_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-25
}

generate_jwt_secret() {
    openssl rand -base64 64 | tr -d "=+/" | cut -c1-64
}

# ==================== 部署函数 ====================
deploy_dependencies() {
    local namespace="$1"
    local environment="$2"
    
    log_step "部署依赖服务..."
    
    # 部署 PostgreSQL
    if [[ "$environment" != "production" ]]; then
        helm upgrade --install postgres-$environment \
            bitnami/postgresql \
            --namespace="$namespace" \
            --create-namespace \
            --set auth.postgresPassword="${DB_PASSWORD:-$(generate_password)}" \
            --set auth.database="taishanglaojun_monitoring" \
            --set auth.username="monitoring_user" \
            --set primary.persistence.size="10Gi" \
            --wait --timeout="$TIMEOUT"
    fi
    
    # 部署 Redis
    if [[ "$environment" != "production" ]]; then
        helm upgrade --install redis-$environment \
            bitnami/redis \
            --namespace="$namespace" \
            --create-namespace \
            --set auth.password="${REDIS_PASSWORD:-$(generate_password)}" \
            --set master.persistence.size="5Gi" \
            --wait --timeout="$TIMEOUT"
    fi
    
    log_success "依赖服务部署完成"
}

deploy_monitoring() {
    local namespace="$1"
    local environment="$2"
    local values_file="$3"
    
    log_step "部署监控系统..."
    
    local release_name="monitoring-$environment"
    local image="$REGISTRY/$IMAGE_NAME:$VERSION"
    
    # 检查镜像
    if ! check_image "$image"; then
        if [[ "$FORCE" != "true" ]]; then
            log_error "镜像不存在，部署终止"
            exit 1
        else
            log_warning "镜像不存在，但强制部署"
        fi
    fi
    
    # 构建 Helm 命令
    local helm_cmd=(
        helm upgrade --install "$release_name"
        "$HELM_CHART_PATH"
        --namespace="$namespace"
        --create-namespace
        --values="$values_file"
        --set image.repository="$REGISTRY/$IMAGE_NAME"
        --set image.tag="$VERSION"
        --set environment="$environment"
    )
    
    if [[ "$WAIT" == "true" ]]; then
        helm_cmd+=(--wait --timeout="$TIMEOUT")
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd+=(--dry-run)
        log_info "执行 Dry Run 模式"
    fi
    
    # 执行部署
    if "${helm_cmd[@]}"; then
        log_success "监控系统部署完成"
    else
        log_error "监控系统部署失败"
        exit 1
    fi
    
    # 等待部署完成
    if [[ "$DRY_RUN" != "true" && "$WAIT" == "true" ]]; then
        wait_for_rollout "$release_name" "$namespace"
    fi
}

# ==================== 测试函数 ====================
run_deployment_tests() {
    local namespace="$1"
    local environment="$2"
    
    if [[ "$SKIP_TESTS" == "true" ]]; then
        log_warning "跳过部署测试"
        return 0
    fi
    
    log_step "运行部署测试..."
    
    # 检查 Pod 状态
    log_info "检查 Pod 状态..."
    kubectl get pods -n "$namespace" -l app=monitoring
    
    # 检查服务状态
    log_info "检查服务状态..."
    kubectl get services -n "$namespace"
    
    # 端口转发进行健康检查
    log_info "进行健康检查..."
    kubectl port-forward -n "$namespace" service/monitoring-$environment 8080:8080 &
    local port_forward_pid=$!
    
    sleep 10
    
    if check_health "http://localhost:8080" 5 5; then
        log_success "健康检查通过"
    else
        log_error "健康检查失败"
        kill $port_forward_pid 2>/dev/null || true
        exit 1
    fi
    
    kill $port_forward_pid 2>/dev/null || true
    
    log_success "部署测试完成"
}

# ==================== 回滚函数 ====================
rollback_deployment() {
    local namespace="$1"
    local environment="$2"
    local revision="${3:-}"
    
    log_step "回滚部署..."
    
    local release_name="monitoring-$environment"
    
    if [[ -n "$revision" ]]; then
        helm rollback "$release_name" "$revision" --namespace="$namespace"
    else
        helm rollback "$release_name" --namespace="$namespace"
    fi
    
    wait_for_rollout "$release_name" "$namespace"
    
    log_success "部署回滚完成"
}

# ==================== 清理函数 ====================
cleanup_deployment() {
    local namespace="$1"
    local environment="$2"
    
    confirm_action "这将删除 $environment 环境的所有资源"
    
    log_step "清理部署..."
    
    local release_name="monitoring-$environment"
    
    # 删除 Helm 发布
    if helm list -n "$namespace" | grep -q "$release_name"; then
        helm uninstall "$release_name" --namespace="$namespace"
    fi
    
    # 删除 Secrets
    kubectl delete secret monitoring-db-secret monitoring-redis-secret monitoring-jwt-secret \
        --namespace="$namespace" --ignore-not-found=true
    
    # 删除命名空间（如果为空）
    if [[ "$(kubectl get all -n "$namespace" --no-headers | wc -l)" -eq 0 ]]; then
        kubectl delete namespace "$namespace" --ignore-not-found=true
    fi
    
    log_success "清理完成"
}

# ==================== 状态检查 ====================
show_status() {
    local namespace="$1"
    local environment="$2"
    
    log_step "显示部署状态..."
    
    echo -e "${CYAN}=== Helm 发布状态 ===${RESET}"
    helm list -n "$namespace"
    
    echo -e "${CYAN}=== Pod 状态 ===${RESET}"
    kubectl get pods -n "$namespace" -l app=monitoring
    
    echo -e "${CYAN}=== 服务状态 ===${RESET}"
    kubectl get services -n "$namespace"
    
    echo -e "${CYAN}=== Ingress 状态 ===${RESET}"
    kubectl get ingress -n "$namespace"
    
    echo -e "${CYAN}=== 事件 ===${RESET}"
    kubectl get events -n "$namespace" --sort-by='.lastTimestamp' | tail -10
}

# ==================== 日志查看 ====================
show_logs() {
    local namespace="$1"
    local environment="$2"
    local follow="${3:-false}"
    
    log_step "显示应用日志..."
    
    local pod_selector="app=monitoring"
    
    if [[ "$follow" == "true" ]]; then
        kubectl logs -f -n "$namespace" -l "$pod_selector" --tail=100
    else
        kubectl logs -n "$namespace" -l "$pod_selector" --tail=100
    fi
}

# ==================== 帮助信息 ====================
show_help() {
    cat << EOF
太上老君监控系统部署脚本

用法: $0 [命令] [选项]

命令:
  deploy          部署到指定环境
  rollback        回滚部署
  status          显示部署状态
  logs            显示应用日志
  test            运行部署测试
  cleanup         清理部署
  help            显示帮助信息

选项:
  -e, --environment   目标环境 (development|staging|production)
  -n, --namespace     Kubernetes 命名空间
  -v, --version       镜像版本
  -f, --force         强制执行，跳过确认
  -d, --dry-run       Dry run 模式
  -s, --skip-tests    跳过测试
  -w, --no-wait       不等待部署完成
  -t, --timeout       超时时间 (默认: 600s)

环境变量:
  ENVIRONMENT         目标环境
  NAMESPACE           Kubernetes 命名空间
  VERSION             镜像版本
  REGISTRY            Docker 注册表
  IMAGE_NAME          镜像名称
  KUBECONFIG          Kubernetes 配置文件
  HELM_CHART_PATH     Helm Chart 路径
  FORCE               强制执行
  DRY_RUN             Dry run 模式
  SKIP_TESTS          跳过测试
  WAIT                等待部署完成
  TIMEOUT             超时时间

示例:
  $0 deploy -e development           # 部署到开发环境
  $0 deploy -e production -v v1.0.0  # 部署指定版本到生产环境
  $0 rollback -e staging             # 回滚预发布环境
  $0 status -e production            # 查看生产环境状态
  $0 logs -e development -f          # 查看开发环境日志
  $0 cleanup -e development          # 清理开发环境

EOF
}

# ==================== 参数解析 ====================
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -f|--force)
                FORCE="true"
                shift
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            -s|--skip-tests)
                SKIP_TESTS="true"
                shift
                ;;
            -w|--no-wait)
                WAIT="false"
                shift
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
}

# ==================== 主函数 ====================
main() {
    local command="${1:-help}"
    shift || true
    
    # 解析参数
    parse_args "$@"
    
    # 验证环境
    if [[ ! "$ENVIRONMENT" =~ ^(development|staging|production)$ ]]; then
        log_error "无效的环境: $ENVIRONMENT"
        log_info "支持的环境: development, staging, production"
        exit 1
    fi
    
    # 设置命名空间
    NAMESPACE="${NAMESPACE}-${ENVIRONMENT}"
    
    case "$command" in
        deploy)
            check_environment
            
            log_info "部署配置:"
            log_info "  环境: $ENVIRONMENT"
            log_info "  命名空间: $NAMESPACE"
            log_info "  镜像: $REGISTRY/$IMAGE_NAME:$VERSION"
            log_info "  Dry Run: $DRY_RUN"
            
            if [[ "$ENVIRONMENT" == "production" ]]; then
                confirm_action "即将部署到生产环境"
            fi
            
            # 生成配置文件
            values_file=$(generate_values_file "$ENVIRONMENT")
            
            # 创建密钥
            create_secrets "$NAMESPACE" "$ENVIRONMENT"
            
            # 部署依赖
            deploy_dependencies "$NAMESPACE" "$ENVIRONMENT"
            
            # 部署监控系统
            deploy_monitoring "$NAMESPACE" "$ENVIRONMENT" "$values_file"
            
            # 运行测试
            run_deployment_tests "$NAMESPACE" "$ENVIRONMENT"
            
            log_success "部署完成"
            ;;
        rollback)
            check_environment
            rollback_deployment "$NAMESPACE" "$ENVIRONMENT" "${1:-}"
            ;;
        status)
            check_environment
            show_status "$NAMESPACE" "$ENVIRONMENT"
            ;;
        logs)
            check_environment
            show_logs "$NAMESPACE" "$ENVIRONMENT" "${1:-false}"
            ;;
        test)
            check_environment
            run_deployment_tests "$NAMESPACE" "$ENVIRONMENT"
            ;;
        cleanup)
            check_environment
            cleanup_deployment "$NAMESPACE" "$ENVIRONMENT"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"