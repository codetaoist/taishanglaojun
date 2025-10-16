#!/bin/bash

# Kubernetes部署脚本 - 高级AI功能
# 作者: TaiShangLaoJun Team
# 版本: 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
NAMESPACE="taishanglaojun"
APP_NAME="advanced-ai-service"
DOCKER_REGISTRY="your-registry.com"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ENVIRONMENT="${ENVIRONMENT:-production}"
KUBECTL_TIMEOUT="300s"

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

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl未安装，请先安装kubectl"
        exit 1
    fi
    
    # 检查helm (可选)
    if ! command -v helm &> /dev/null; then
        log_warning "helm未安装，某些功能可能不可用"
    fi
    
    # 检查docker
    if ! command -v docker &> /dev/null; then
        log_error "docker未安装，请先安装docker"
        exit 1
    fi
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 创建命名空间
create_namespace() {
    log_info "创建命名空间..."
    
    if kubectl get namespace $NAMESPACE &> /dev/null; then
        log_warning "命名空间 $NAMESPACE 已存在"
    else
        kubectl apply -f namespace.yaml
        log_success "命名空间 $NAMESPACE 创建成功"
    fi
}

# 应用密钥
apply_secrets() {
    log_info "应用密钥配置..."
    
    # 检查是否需要更新密钥
    if [[ "$ENVIRONMENT" == "production" ]]; then
        log_warning "生产环境部署，请确保已正确配置所有密钥"
        read -p "是否继续? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_error "部署已取消"
            exit 1
        fi
    fi
    
    kubectl apply -f secrets.yaml
    log_success "密钥配置应用成功"
}

# 应用配置映射
apply_configmaps() {
    log_info "应用配置映射..."
    kubectl apply -f configmap.yaml
    log_success "配置映射应用成功"
}

# 应用存储配置
apply_storage() {
    log_info "应用存储配置..."
    kubectl apply -f storage.yaml
    
    # 等待PVC绑定
    log_info "等待PVC绑定..."
    kubectl wait --for=condition=Bound pvc --all -n $NAMESPACE --timeout=$KUBECTL_TIMEOUT
    log_success "存储配置应用成功"
}

# 应用RBAC配置
apply_rbac() {
    log_info "应用RBAC配置..."
    kubectl apply -f rbac.yaml
    log_success "RBAC配置应用成功"
}

# 应用服务配置
apply_services() {
    log_info "应用服务配置..."
    kubectl apply -f service.yaml
    log_success "服务配置应用成功"
}

# 应用部署配置
apply_deployments() {
    log_info "应用部署配置..."
    
    # 替换镜像标签
    if [[ "$IMAGE_TAG" != "latest" ]]; then
        sed -i.bak "s|image: $DOCKER_REGISTRY/$APP_NAME:latest|image: $DOCKER_REGISTRY/$APP_NAME:$IMAGE_TAG|g" deployment.yaml
    fi
    
    kubectl apply -f deployment.yaml
    
    # 等待部署完成
    log_info "等待部署完成..."
    kubectl rollout status deployment/$APP_NAME -n $NAMESPACE --timeout=$KUBECTL_TIMEOUT
    kubectl rollout status deployment/postgres -n $NAMESPACE --timeout=$KUBECTL_TIMEOUT
    kubectl rollout status deployment/redis -n $NAMESPACE --timeout=$KUBECTL_TIMEOUT
    kubectl rollout status deployment/nginx -n $NAMESPACE --timeout=$KUBECTL_TIMEOUT
    
    log_success "部署配置应用成功"
}

# 应用HPA配置
apply_hpa() {
    log_info "应用HPA配置..."
    kubectl apply -f hpa.yaml
    log_success "HPA配置应用成功"
}

# 应用Ingress配置
apply_ingress() {
    log_info "应用Ingress配置..."
    kubectl apply -f ingress.yaml
    log_success "Ingress配置应用成功"
}

# 应用监控配置
apply_monitoring() {
    log_info "应用监控配置..."
    kubectl apply -f monitoring.yaml
    log_success "监控配置应用成功"
}

# 验证部署
verify_deployment() {
    log_info "验证部署..."
    
    # 检查Pod状态
    log_info "检查Pod状态..."
    kubectl get pods -n $NAMESPACE
    
    # 检查服务状态
    log_info "检查服务状态..."
    kubectl get services -n $NAMESPACE
    
    # 检查Ingress状态
    log_info "检查Ingress状态..."
    kubectl get ingress -n $NAMESPACE
    
    # 健康检查
    log_info "执行健康检查..."
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if kubectl exec -n $NAMESPACE deployment/$APP_NAME -- curl -f http://localhost:8080/health &> /dev/null; then
            log_success "健康检查通过"
            break
        fi
        
        log_info "健康检查失败，重试 ($attempt/$max_attempts)..."
        sleep 10
        ((attempt++))
    done
    
    if [[ $attempt -gt $max_attempts ]]; then
        log_error "健康检查失败"
        return 1
    fi
    
    log_success "部署验证完成"
}

# 显示部署信息
show_deployment_info() {
    log_info "部署信息:"
    echo "=================================="
    echo "命名空间: $NAMESPACE"
    echo "应用名称: $APP_NAME"
    echo "镜像标签: $IMAGE_TAG"
    echo "环境: $ENVIRONMENT"
    echo "=================================="
    
    # 获取外部访问地址
    local ingress_ip=$(kubectl get ingress advanced-ai-ingress -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")
    local ingress_hostname=$(kubectl get ingress advanced-ai-ingress -n $NAMESPACE -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "N/A")
    
    echo "访问地址:"
    echo "  主域名: https://$ingress_hostname"
    echo "  IP地址: $ingress_ip"
    echo "  健康检查: https://$ingress_hostname/health"
    echo "  API文档: https://$ingress_hostname/api/v1/docs"
    echo "  监控面板: https://grafana.$ingress_hostname"
    echo "  Prometheus: https://prometheus.$ingress_hostname"
    echo "=================================="
    
    # 显示Pod状态
    echo "Pod状态:"
    kubectl get pods -n $NAMESPACE -o wide
    echo "=================================="
    
    # 显示资源使用情况
    echo "资源使用情况:"
    kubectl top pods -n $NAMESPACE 2>/dev/null || echo "Metrics server未安装"
    echo "=================================="
}

# 清理部署
cleanup_deployment() {
    log_warning "清理部署..."
    read -p "确定要删除所有资源吗? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete namespace $NAMESPACE
        log_success "部署清理完成"
    else
        log_info "清理已取消"
    fi
}

# 更新部署
update_deployment() {
    log_info "更新部署..."
    
    # 更新镜像
    kubectl set image deployment/$APP_NAME $APP_NAME=$DOCKER_REGISTRY/$APP_NAME:$IMAGE_TAG -n $NAMESPACE
    
    # 等待滚动更新完成
    kubectl rollout status deployment/$APP_NAME -n $NAMESPACE --timeout=$KUBECTL_TIMEOUT
    
    log_success "部署更新完成"
}

# 备份配置
backup_config() {
    log_info "备份配置..."
    local backup_dir="backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p $backup_dir
    
    # 备份所有配置
    kubectl get all,configmap,secret,pvc,ingress -n $NAMESPACE -o yaml > $backup_dir/backup.yaml
    
    log_success "配置已备份到 $backup_dir/"
}

# 恢复配置
restore_config() {
    local backup_file="$1"
    if [[ -z "$backup_file" ]]; then
        log_error "请指定备份文件"
        exit 1
    fi
    
    log_info "恢复配置..."
    kubectl apply -f $backup_file
    log_success "配置恢复完成"
}

# 查看日志
view_logs() {
    local pod_name="$1"
    if [[ -z "$pod_name" ]]; then
        # 显示所有Pod
        kubectl get pods -n $NAMESPACE
        read -p "请输入Pod名称: " pod_name
    fi
    
    kubectl logs -f $pod_name -n $NAMESPACE
}

# 进入Pod
exec_pod() {
    local pod_name="$1"
    if [[ -z "$pod_name" ]]; then
        # 显示所有Pod
        kubectl get pods -n $NAMESPACE
        read -p "请输入Pod名称: " pod_name
    fi
    
    kubectl exec -it $pod_name -n $NAMESPACE -- /bin/bash
}

# 主函数
main() {
    case "${1:-deploy}" in
        "deploy")
            log_info "开始部署高级AI功能到Kubernetes..."
            check_dependencies
            create_namespace
            apply_secrets
            apply_configmaps
            apply_storage
            apply_rbac
            apply_services
            apply_deployments
            apply_hpa
            apply_ingress
            apply_monitoring
            verify_deployment
            show_deployment_info
            log_success "部署完成!"
            ;;
        "update")
            log_info "更新部署..."
            check_dependencies
            update_deployment
            verify_deployment
            ;;
        "cleanup")
            cleanup_deployment
            ;;
        "verify")
            verify_deployment
            ;;
        "info")
            show_deployment_info
            ;;
        "backup")
            backup_config
            ;;
        "restore")
            restore_config "$2"
            ;;
        "logs")
            view_logs "$2"
            ;;
        "exec")
            exec_pod "$2"
            ;;
        "help"|"-h"|"--help")
            echo "用法: $0 [命令] [参数]"
            echo ""
            echo "命令:"
            echo "  deploy    - 部署应用 (默认)"
            echo "  update    - 更新部署"
            echo "  cleanup   - 清理部署"
            echo "  verify    - 验证部署"
            echo "  info      - 显示部署信息"
            echo "  backup    - 备份配置"
            echo "  restore   - 恢复配置"
            echo "  logs      - 查看日志"
            echo "  exec      - 进入Pod"
            echo "  help      - 显示帮助"
            echo ""
            echo "环境变量:"
            echo "  IMAGE_TAG     - 镜像标签 (默认: latest)"
            echo "  ENVIRONMENT   - 环境 (默认: production)"
            echo ""
            echo "示例:"
            echo "  $0 deploy"
            echo "  IMAGE_TAG=v1.0.0 $0 update"
            echo "  $0 logs advanced-ai-service-xxx"
            echo "  $0 restore backup.yaml"
            ;;
        *)
            log_error "未知命令: $1"
            echo "使用 '$0 help' 查看帮助"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"