#!/bin/bash

# 太上老君AI平台 - Kubernetes部署脚本
# 版本: 1.0.0
# 创建时间: 2024-01-01

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查kubectl是否可用
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在PATH中"
        exit 1
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    log_success "Kubernetes集群连接正常"
}

# 创建命名空间
create_namespace() {
    log_info "创建命名空间..."
    kubectl apply -f namespace.yaml
    log_success "命名空间创建完成"
}

# 创建ConfigMaps和Secrets
create_configs() {
    log_info "创建配置文件..."
    kubectl apply -f configmap.yaml
    kubectl apply -f secrets.yaml
    log_success "配置文件创建完成"
}

# 部署数据库
deploy_database() {
    log_info "部署PostgreSQL数据库..."
    kubectl apply -f postgres.yaml
    
    log_info "等待PostgreSQL启动..."
    kubectl wait --for=condition=ready pod -l app=postgres -n taishanglaojun --timeout=300s
    log_success "PostgreSQL部署完成"
}

# 部署Redis
deploy_redis() {
    log_info "部署Redis缓存..."
    kubectl apply -f redis.yaml
    
    log_info "等待Redis启动..."
    kubectl wait --for=condition=ready pod -l app=redis -n taishanglaojun --timeout=300s
    log_success "Redis部署完成"
}

# 部署核心服务
deploy_core_services() {
    log_info "部署核心服务..."
    kubectl apply -f core-services.yaml
    
    log_info "等待核心服务启动..."
    kubectl wait --for=condition=ready pod -l app=core-services -n taishanglaojun --timeout=300s
    log_success "核心服务部署完成"
}

# 部署前端
deploy_frontend() {
    log_info "部署前端应用..."
    kubectl apply -f frontend.yaml
    
    log_info "等待前端应用启动..."
    kubectl wait --for=condition=ready pod -l app=frontend -n taishanglaojun --timeout=300s
    log_success "前端应用部署完成"
}

# 部署监控
deploy_monitoring() {
    log_info "部署监控系统..."
    kubectl apply -f monitoring.yaml
    
    log_info "等待监控系统启动..."
    kubectl wait --for=condition=ready pod -l app=prometheus -n taishanglaojun --timeout=300s
    kubectl wait --for=condition=ready pod -l app=grafana -n taishanglaojun --timeout=300s
    log_success "监控系统部署完成"
}

# 验证部署
verify_deployment() {
    log_info "验证部署状态..."
    
    echo ""
    log_info "Pod状态:"
    kubectl get pods -n taishanglaojun
    
    echo ""
    log_info "Service状态:"
    kubectl get services -n taishanglaojun
    
    echo ""
    log_info "Ingress状态:"
    kubectl get ingress -n taishanglaojun
    
    echo ""
    log_info "PVC状态:"
    kubectl get pvc -n taishanglaojun
    
    # 检查所有Pod是否运行正常
    failed_pods=$(kubectl get pods -n taishanglaojun --field-selector=status.phase!=Running --no-headers 2>/dev/null | wc -l)
    if [ "$failed_pods" -eq 0 ]; then
        log_success "所有Pod运行正常"
    else
        log_warning "有 $failed_pods 个Pod未正常运行"
        kubectl get pods -n taishanglaojun --field-selector=status.phase!=Running
    fi
}

# 显示访问信息
show_access_info() {
    echo ""
    log_info "=== 访问信息 ==="
    
    # 获取Ingress信息
    ingress_ip=$(kubectl get ingress frontend-ingress -n taishanglaojun -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
    
    if [ "$ingress_ip" != "pending" ] && [ -n "$ingress_ip" ]; then
        echo "前端应用: https://taishanglaojun.com"
        echo "API接口: https://taishanglaojun.com/api"
    else
        # 如果没有Ingress，显示NodePort或端口转发信息
        log_info "使用端口转发访问应用:"
        echo "前端应用: kubectl port-forward -n taishanglaojun svc/frontend-service 8080:80"
        echo "核心服务: kubectl port-forward -n taishanglaojun svc/core-services 8081:8080"
        echo "Grafana: kubectl port-forward -n taishanglaojun svc/grafana-service 3000:3000"
        echo "Prometheus: kubectl port-forward -n taishanglaojun svc/prometheus-service 9090:9090"
    fi
    
    echo ""
    log_info "默认登录信息:"
    echo "Grafana - 用户名: admin, 密码: 查看secret获取"
    echo "数据库 - 用户名: postgres, 密码: 查看secret获取"
    
    echo ""
    log_info "获取密码命令:"
    echo "kubectl get secret taishanglaojun-secrets -n taishanglaojun -o jsonpath='{.data.grafana-admin-password}' | base64 -d"
}

# 清理部署
cleanup() {
    log_warning "清理所有部署资源..."
    kubectl delete namespace taishanglaojun --ignore-not-found=true
    log_success "清理完成"
}

# 主函数
main() {
    echo "太上老君AI平台 - Kubernetes部署脚本"
    echo "========================================"
    
    case "${1:-deploy}" in
        "deploy")
            check_kubectl
            create_namespace
            create_configs
            deploy_database
            deploy_redis
            deploy_core_services
            deploy_frontend
            deploy_monitoring
            verify_deployment
            show_access_info
            log_success "部署完成！"
            ;;
        "cleanup")
            cleanup
            ;;
        "verify")
            verify_deployment
            ;;
        "info")
            show_access_info
            ;;
        *)
            echo "用法: $0 [deploy|cleanup|verify|info]"
            echo "  deploy  - 部署所有服务 (默认)"
            echo "  cleanup - 清理所有资源"
            echo "  verify  - 验证部署状态"
            echo "  info    - 显示访问信息"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"