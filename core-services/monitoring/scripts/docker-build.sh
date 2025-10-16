#!/bin/bash

# 太上老君监控系统 Docker 构建脚本
# 支持多平台构建和推送到容器注册表

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICE_NAME="monitoring"
REGISTRY="${REGISTRY:-ghcr.io}"
NAMESPACE="${NAMESPACE:-taishanglaojun}"
IMAGE_NAME="${REGISTRY}/${NAMESPACE}/${SERVICE_NAME}"

# 颜色输出
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

# 显示帮助信息
show_help() {
    cat << EOF
太上老君监控系统 Docker 构建脚本

用法: $0 [选项] [命令]

命令:
    build       构建 Docker 镜像
    push        推送镜像到注册表
    build-push  构建并推送镜像
    clean       清理本地镜像
    help        显示此帮助信息

选项:
    -t, --tag TAG           镜像标签 (默认: latest)
    -p, --platform PLATFORM 目标平台 (默认: linux/amd64,linux/arm64)
    -r, --registry REGISTRY 容器注册表 (默认: ghcr.io)
    -n, --namespace NS      命名空间 (默认: taishanglaojun)
    --no-cache             不使用构建缓存
    --push                 构建后自动推送
    --latest               同时标记为 latest
    -v, --verbose          详细输出
    -h, --help             显示帮助信息

环境变量:
    REGISTRY               容器注册表地址
    NAMESPACE              镜像命名空间
    DOCKER_BUILDKIT        启用 BuildKit (默认: 1)
    BUILDX_PLATFORMS       构建平台列表

示例:
    $0 build -t v1.0.0
    $0 build-push -t v1.0.0 --latest
    $0 push -t v1.0.0
    $0 clean

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi
    
    if ! docker buildx version &> /dev/null; then
        log_error "Docker Buildx 未安装或不可用"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 获取版本信息
get_version() {
    local version=""
    
    # 尝试从 git tag 获取版本
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        version=$(git describe --tags --exact-match HEAD)
    elif git rev-parse --short HEAD 2>/dev/null; then
        version="dev-$(git rev-parse --short HEAD)"
    else
        version="dev-$(date +%Y%m%d%H%M%S)"
    fi
    
    echo "$version"
}

# 构建镜像
build_image() {
    local tag="$1"
    local platforms="$2"
    local no_cache="$3"
    local push="$4"
    local latest="$5"
    local verbose="$6"
    
    log_info "开始构建 Docker 镜像..."
    log_info "镜像名称: ${IMAGE_NAME}"
    log_info "标签: ${tag}"
    log_info "平台: ${platforms}"
    
    cd "$PROJECT_ROOT"
    
    # 构建参数
    local build_args=(
        "buildx" "build"
        "--platform" "$platforms"
        "--tag" "${IMAGE_NAME}:${tag}"
        "--file" "Dockerfile"
    )
    
    # 添加构建时参数
    build_args+=(
        "--build-arg" "VERSION=${tag}"
        "--build-arg" "BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
        "--build-arg" "VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
    )
    
    if [[ "$no_cache" == "true" ]]; then
        build_args+=("--no-cache")
    fi
    
    if [[ "$push" == "true" ]]; then
        build_args+=("--push")
    else
        build_args+=("--load")
    fi
    
    if [[ "$latest" == "true" ]]; then
        build_args+=("--tag" "${IMAGE_NAME}:latest")
    fi
    
    if [[ "$verbose" == "true" ]]; then
        build_args+=("--progress" "plain")
    fi
    
    build_args+=(".")
    
    # 执行构建
    log_info "执行构建命令: docker ${build_args[*]}"
    
    if docker "${build_args[@]}"; then
        log_success "镜像构建成功"
        
        if [[ "$push" != "true" ]]; then
            log_info "镜像已保存到本地"
            docker images "${IMAGE_NAME}:${tag}"
        fi
    else
        log_error "镜像构建失败"
        exit 1
    fi
}

# 推送镜像
push_image() {
    local tag="$1"
    local latest="$2"
    
    log_info "推送镜像到注册表..."
    
    if docker push "${IMAGE_NAME}:${tag}"; then
        log_success "镜像 ${IMAGE_NAME}:${tag} 推送成功"
    else
        log_error "镜像推送失败"
        exit 1
    fi
    
    if [[ "$latest" == "true" ]]; then
        if docker push "${IMAGE_NAME}:latest"; then
            log_success "镜像 ${IMAGE_NAME}:latest 推送成功"
        else
            log_error "latest 标签推送失败"
            exit 1
        fi
    fi
}

# 清理镜像
clean_images() {
    log_info "清理本地镜像..."
    
    # 清理悬空镜像
    if docker image prune -f; then
        log_success "悬空镜像清理完成"
    fi
    
    # 清理项目相关镜像
    local images
    images=$(docker images "${IMAGE_NAME}" -q)
    if [[ -n "$images" ]]; then
        log_info "删除项目镜像..."
        docker rmi $images || true
        log_success "项目镜像清理完成"
    else
        log_info "没有找到项目相关镜像"
    fi
}

# 主函数
main() {
    # 默认参数
    local command=""
    local tag="latest"
    local platforms="linux/amd64,linux/arm64"
    local no_cache="false"
    local push="false"
    local latest="false"
    local verbose="false"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            build|push|build-push|clean|help)
                command="$1"
                shift
                ;;
            -t|--tag)
                tag="$2"
                shift 2
                ;;
            -p|--platform)
                platforms="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                IMAGE_NAME="${REGISTRY}/${NAMESPACE}/${SERVICE_NAME}"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                IMAGE_NAME="${REGISTRY}/${NAMESPACE}/${SERVICE_NAME}"
                shift 2
                ;;
            --no-cache)
                no_cache="true"
                shift
                ;;
            --push)
                push="true"
                shift
                ;;
            --latest)
                latest="true"
                shift
                ;;
            -v|--verbose)
                verbose="true"
                shift
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
    
    # 如果没有指定命令，显示帮助
    if [[ -z "$command" ]]; then
        show_help
        exit 1
    fi
    
    # 启用 BuildKit
    export DOCKER_BUILDKIT=1
    
    # 检查依赖
    check_dependencies
    
    # 如果标签是 auto，自动获取版本
    if [[ "$tag" == "auto" ]]; then
        tag=$(get_version)
        log_info "自动检测版本: $tag"
    fi
    
    # 执行命令
    case $command in
        build)
            build_image "$tag" "$platforms" "$no_cache" "$push" "$latest" "$verbose"
            ;;
        push)
            push_image "$tag" "$latest"
            ;;
        build-push)
            build_image "$tag" "$platforms" "$no_cache" "true" "$latest" "$verbose"
            ;;
        clean)
            clean_images
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"