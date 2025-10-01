#!/bin/bash

# API Gateway 部署脚本
# 支持开发、测试、生产环境部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
ENVIRONMENT="dev"
BUILD_IMAGE=false
PUSH_IMAGE=false
REGISTRY=""
TAG="latest"
CONFIG_FILE=""
COMPOSE_FILE="docker-compose.yml"

# 显示帮助信息
show_help() {
    echo "API Gateway 部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -e, --env ENVIRONMENT     部署环境 (dev|test|prod) [默认: dev]"
    echo "  -b, --build              构建Docker镜像"
    echo "  -p, --push               推送镜像到仓库"
    echo "  -r, --registry REGISTRY  Docker镜像仓库地址"
    echo "  -t, --tag TAG            镜像标签 [默认: latest]"
    echo "  -c, --config CONFIG      配置文件路径"
    echo "  -f, --file COMPOSE_FILE  Docker Compose文件 [默认: docker-compose.yml]"
    echo "  -h, --help               显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -e dev -b                    # 开发环境部署并构建镜像"
    echo "  $0 -e prod -b -p -r registry.com -t v1.0.0  # 生产环境部署"
    echo "  $0 -e test -c configs/test.yaml  # 测试环境部署"
}

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
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 验证环境
validate_environment() {
    case $ENVIRONMENT in
        dev|test|prod)
            log_info "部署环境: $ENVIRONMENT"
            ;;
        *)
            log_error "无效的环境: $ENVIRONMENT"
            exit 1
            ;;
    esac
}

# 设置环境变量
setup_environment() {
    log_info "设置环境变量..."
    
    export ENVIRONMENT=$ENVIRONMENT
    export TAG=$TAG
    
    if [ -n "$REGISTRY" ]; then
        export REGISTRY=$REGISTRY
    fi
    
    if [ -n "$CONFIG_FILE" ]; then
        export CONFIG_FILE=$CONFIG_FILE
    fi
    
    # 根据环境设置不同的配置
    case $ENVIRONMENT in
        dev)
            export GIN_MODE=debug
            export LOG_LEVEL=debug
            ;;
        test)
            export GIN_MODE=test
            export LOG_LEVEL=info
            ;;
        prod)
            export GIN_MODE=release
            export LOG_LEVEL=warn
            ;;
    esac
    
    log_success "环境变量设置完成"
}

# 构建镜像
build_image() {
    if [ "$BUILD_IMAGE" = true ]; then
        log_info "构建Docker镜像..."
        
        local image_name="api-gateway"
        if [ -n "$REGISTRY" ]; then
            image_name="$REGISTRY/api-gateway"
        fi
        
        docker build -t "$image_name:$TAG" .
        
        # 如果是latest标签，也打上环境标签
        if [ "$TAG" = "latest" ]; then
            docker tag "$image_name:$TAG" "$image_name:$ENVIRONMENT"
        fi
        
        log_success "镜像构建完成: $image_name:$TAG"
    fi
}

# 推送镜像
push_image() {
    if [ "$PUSH_IMAGE" = true ]; then
        if [ -z "$REGISTRY" ]; then
            log_error "推送镜像需要指定仓库地址 (-r)"
            exit 1
        fi
        
        log_info "推送镜像到仓库..."
        
        local image_name="$REGISTRY/api-gateway"
        docker push "$image_name:$TAG"
        
        if [ "$TAG" = "latest" ]; then
            docker push "$image_name:$ENVIRONMENT"
        fi
        
        log_success "镜像推送完成"
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p logs
    mkdir -p data
    mkdir -p configs
    
    # 根据环境创建不同的日志目录
    mkdir -p "logs/$ENVIRONMENT"
    
    log_success "目录创建完成"
}

# 验证配置文件
validate_config() {
    log_info "验证配置文件..."
    
    local config_path="configs/gateway.yaml"
    if [ -n "$CONFIG_FILE" ]; then
        config_path="$CONFIG_FILE"
    fi
    
    if [ ! -f "$config_path" ]; then
        log_error "配置文件不存在: $config_path"
        exit 1
    fi
    
    # 可以添加配置文件语法验证
    # go run cmd/main.go -config "$config_path" -validate
    
    log_success "配置文件验证完成"
}

# 部署服务
deploy_services() {
    log_info "部署服务..."
    
    # 停止现有服务
    docker-compose -f "$COMPOSE_FILE" down
    
    # 根据环境选择不同的compose配置
    case $ENVIRONMENT in
        dev)
            docker-compose -f "$COMPOSE_FILE" up -d redis prometheus grafana
            docker-compose -f "$COMPOSE_FILE" up -d api-gateway
            ;;
        test)
            docker-compose -f "$COMPOSE_FILE" up -d redis
            docker-compose -f "$COMPOSE_FILE" up -d api-gateway-test
            ;;
        prod)
            docker-compose -f "$COMPOSE_FILE" up -d redis prometheus grafana jaeger
            docker-compose -f "$COMPOSE_FILE" up -d api-gateway
            ;;
    esac
    
    log_success "服务部署完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            log_success "健康检查通过"
            return 0
        fi
        
        log_info "等待服务启动... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    log_error "健康检查失败"
    return 1
}

# 显示部署信息
show_deployment_info() {
    log_success "部署完成!"
    echo ""
    echo "服务信息:"
    echo "  API Gateway: http://localhost:8080"
    echo "  健康检查:   http://localhost:8080/health"
    echo "  就绪检查:   http://localhost:8080/ready"
    echo "  监控指标:   http://localhost:9090/metrics"
    
    if [ "$ENVIRONMENT" != "test" ]; then
        echo "  Prometheus: http://localhost:9091"
        echo "  Grafana:    http://localhost:3000 (admin/admin123)"
    fi
    
    if [ "$ENVIRONMENT" = "prod" ]; then
        echo "  Jaeger:     http://localhost:16686"
    fi
    
    echo ""
    echo "管理命令:"
    echo "  查看日志: docker-compose logs -f api-gateway"
    echo "  停止服务: docker-compose down"
    echo "  重启服务: docker-compose restart api-gateway"
}

# 清理函数
cleanup() {
    log_info "清理临时文件..."
    # 清理逻辑
}

# 信号处理
trap cleanup EXIT

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -b|--build)
            BUILD_IMAGE=true
            shift
            ;;
        -p|--push)
            PUSH_IMAGE=true
            shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -f|--file)
            COMPOSE_FILE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主流程
main() {
    log_info "开始部署 API Gateway..."
    
    check_dependencies
    validate_environment
    setup_environment
    create_directories
    validate_config
    build_image
    push_image
    deploy_services
    
    if health_check; then
        show_deployment_info
    else
        log_error "部署失败"
        exit 1
    fi
}

# 执行主流程
main