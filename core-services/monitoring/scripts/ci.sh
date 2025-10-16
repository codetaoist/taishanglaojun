#!/bin/bash

# 太上老君监控系统 CI/CD 脚本
# 支持构建、测试、部署的完整流水线

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICE_NAME="monitoring"

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
太上老君监控系统 CI/CD 脚本

用法: $0 [选项] [命令]

命令:
    build       构建应用
    test        运行测试
    package     打包应用
    deploy      部署应用
    pipeline    运行完整流水线
    release     发布版本
    rollback    回滚部署
    status      查看状态
    clean       清理环境
    help        显示此帮助信息

选项:
    -e, --environment ENV   目标环境 (development|staging|production)
    -v, --version VERSION   版本号
    -b, --branch BRANCH     Git 分支
    -t, --tag TAG          Git 标签
    -r, --registry REG     Docker 镜像仓库
    -n, --namespace NS     Kubernetes 命名空间
    --skip-tests           跳过测试
    --skip-build           跳过构建
    --skip-push            跳过推送
    --force                强制执行
    --dry-run              仅显示将要执行的操作
    --verbose              详细输出
    -h, --help             显示帮助信息

环境变量:
    CI                     CI 环境标识
    GITHUB_TOKEN           GitHub Token
    DOCKER_REGISTRY        Docker 镜像仓库
    DOCKER_USERNAME        Docker 用户名
    DOCKER_PASSWORD        Docker 密码
    KUBECONFIG             Kubernetes 配置文件
    SLACK_WEBHOOK_URL      Slack 通知 Webhook

示例:
    $0 build -v v1.0.0
    $0 test --verbose
    $0 deploy -e production -v v1.0.0
    $0 pipeline -e staging -b develop
    $0 release -v v1.0.0 -t v1.0.0

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    local missing_deps=()
    
    # 检查必需工具
    for tool in git go docker kubectl helm; do
        if ! command -v "$tool" &> /dev/null; then
            missing_deps+=("$tool")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "缺少依赖: ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "依赖检查完成"
}

# 获取版本信息
get_version_info() {
    local version="$1"
    local branch="$2"
    local tag="$3"
    
    # 如果没有指定版本，尝试从 Git 获取
    if [[ -z "$version" ]]; then
        if [[ -n "$tag" ]]; then
            version="$tag"
        elif git describe --tags --exact-match HEAD &> /dev/null; then
            version=$(git describe --tags --exact-match HEAD)
        else
            local commit_hash
            commit_hash=$(git rev-parse --short HEAD)
            version="${branch:-main}-${commit_hash}"
        fi
    fi
    
    echo "$version"
}

# 设置环境变量
setup_environment() {
    local environment="$1"
    local version="$2"
    
    log_info "设置环境变量..."
    
    # 基础环境变量
    export CI_ENVIRONMENT="$environment"
    export CI_VERSION="$version"
    export CI_BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    export CI_COMMIT_SHA=$(git rev-parse HEAD)
    export CI_COMMIT_SHORT_SHA=$(git rev-parse --short HEAD)
    export CI_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    
    # Docker 相关
    export DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io}"
    export DOCKER_NAMESPACE="${DOCKER_NAMESPACE:-taishanglaojun}"
    export DOCKER_IMAGE_NAME="${DOCKER_NAMESPACE}/${SERVICE_NAME}"
    export DOCKER_IMAGE_TAG="$version"
    export DOCKER_IMAGE_FULL="${DOCKER_REGISTRY}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
    
    # Kubernetes 相关
    case $environment in
        development)
            export K8S_NAMESPACE="taishanglaojun-monitoring-dev"
            ;;
        staging)
            export K8S_NAMESPACE="taishanglaojun-monitoring-staging"
            ;;
        production)
            export K8S_NAMESPACE="taishanglaojun-monitoring"
            ;;
    esac
    
    log_info "环境: $CI_ENVIRONMENT"
    log_info "版本: $CI_VERSION"
    log_info "镜像: $DOCKER_IMAGE_FULL"
    log_info "命名空间: $K8S_NAMESPACE"
}

# 运行测试
run_tests() {
    local skip_tests="$1"
    local verbose="$2"
    
    if [[ "$skip_tests" == "true" ]]; then
        log_warning "跳过测试"
        return 0
    fi
    
    log_info "运行测试..."
    
    cd "$PROJECT_ROOT"
    
    # 运行测试脚本
    local test_args=("all")
    
    if [[ "$verbose" == "true" ]]; then
        test_args+=("-v")
    fi
    
    if [[ -f "$SCRIPT_DIR/test.sh" ]]; then
        "$SCRIPT_DIR/test.sh" "${test_args[@]}"
    else
        log_error "测试脚本不存在: $SCRIPT_DIR/test.sh"
        exit 1
    fi
    
    log_success "测试完成"
}

# 构建应用
build_application() {
    local skip_build="$1"
    local version="$2"
    local verbose="$3"
    
    if [[ "$skip_build" == "true" ]]; then
        log_warning "跳过构建"
        return 0
    fi
    
    log_info "构建应用..."
    
    cd "$PROJECT_ROOT"
    
    # 运行构建脚本
    local build_args=("build")
    build_args+=("-t" "$version")
    
    if [[ "$verbose" == "true" ]]; then
        build_args+=("-v")
    fi
    
    if [[ -f "$SCRIPT_DIR/docker-build.sh" ]]; then
        "$SCRIPT_DIR/docker-build.sh" "${build_args[@]}"
    else
        log_error "构建脚本不存在: $SCRIPT_DIR/docker-build.sh"
        exit 1
    fi
    
    log_success "构建完成"
}

# 推送镜像
push_image() {
    local skip_push="$1"
    local registry="$2"
    local verbose="$3"
    
    if [[ "$skip_push" == "true" ]]; then
        log_warning "跳过推送"
        return 0
    fi
    
    log_info "推送镜像..."
    
    # Docker 登录
    if [[ -n "${DOCKER_USERNAME:-}" && -n "${DOCKER_PASSWORD:-}" ]]; then
        echo "$DOCKER_PASSWORD" | docker login "$registry" -u "$DOCKER_USERNAME" --password-stdin
    elif [[ -n "${GITHUB_TOKEN:-}" && "$registry" == "ghcr.io" ]]; then
        echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_ACTOR" --password-stdin
    fi
    
    # 推送镜像
    docker push "$DOCKER_IMAGE_FULL"
    
    log_success "镜像推送完成: $DOCKER_IMAGE_FULL"
}

# 部署应用
deploy_application() {
    local environment="$1"
    local version="$2"
    local namespace="$3"
    local force="$4"
    local dry_run="$5"
    
    log_info "部署应用到 $environment 环境..."
    
    cd "$PROJECT_ROOT"
    
    # 运行部署脚本
    local deploy_args=("deploy")
    deploy_args+=("-e" "$environment")
    deploy_args+=("-t" "$version")
    deploy_args+=("-n" "$namespace")
    deploy_args+=("--wait")
    
    if [[ "$force" == "true" ]]; then
        deploy_args+=("--force")
    fi
    
    if [[ "$dry_run" == "true" ]]; then
        deploy_args+=("--dry-run")
    fi
    
    if [[ -f "$SCRIPT_DIR/deploy.sh" ]]; then
        "$SCRIPT_DIR/deploy.sh" "${deploy_args[@]}"
    else
        log_error "部署脚本不存在: $SCRIPT_DIR/deploy.sh"
        exit 1
    fi
    
    log_success "部署完成"
}

# 发送通知
send_notification() {
    local status="$1"
    local environment="$2"
    local version="$3"
    local message="$4"
    
    if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
        return 0
    fi
    
    local color
    case $status in
        success) color="good" ;;
        failure) color="danger" ;;
        warning) color="warning" ;;
        *) color="#439FE0" ;;
    esac
    
    local payload
    payload=$(cat << EOF
{
    "attachments": [
        {
            "color": "$color",
            "title": "太上老君监控系统 - $status",
            "fields": [
                {
                    "title": "环境",
                    "value": "$environment",
                    "short": true
                },
                {
                    "title": "版本",
                    "value": "$version",
                    "short": true
                },
                {
                    "title": "分支",
                    "value": "$CI_BRANCH",
                    "short": true
                },
                {
                    "title": "提交",
                    "value": "$CI_COMMIT_SHORT_SHA",
                    "short": true
                }
            ],
            "text": "$message",
            "ts": $(date +%s)
        }
    ]
}
EOF
)
    
    curl -X POST -H 'Content-type: application/json' \
        --data "$payload" \
        "$SLACK_WEBHOOK_URL" || true
}

# 运行完整流水线
run_pipeline() {
    local environment="$1"
    local version="$2"
    local skip_tests="$3"
    local skip_build="$4"
    local skip_push="$5"
    local force="$6"
    local dry_run="$7"
    local verbose="$8"
    
    log_info "开始 CI/CD 流水线..."
    
    local start_time
    start_time=$(date +%s)
    
    # 发送开始通知
    send_notification "info" "$environment" "$version" "开始部署流水线"
    
    # 运行测试
    if ! run_tests "$skip_tests" "$verbose"; then
        send_notification "failure" "$environment" "$version" "测试失败"
        exit 1
    fi
    
    # 构建应用
    if ! build_application "$skip_build" "$version" "$verbose"; then
        send_notification "failure" "$environment" "$version" "构建失败"
        exit 1
    fi
    
    # 推送镜像
    if ! push_image "$skip_push" "$DOCKER_REGISTRY" "$verbose"; then
        send_notification "failure" "$environment" "$version" "镜像推送失败"
        exit 1
    fi
    
    # 部署应用
    if ! deploy_application "$environment" "$version" "$K8S_NAMESPACE" "$force" "$dry_run"; then
        send_notification "failure" "$environment" "$version" "部署失败"
        exit 1
    fi
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_success "CI/CD 流水线完成，耗时: ${duration}s"
    send_notification "success" "$environment" "$version" "部署成功，耗时: ${duration}s"
}

# 发布版本
release_version() {
    local version="$1"
    local tag="$2"
    local force="$3"
    
    log_info "发布版本: $version"
    
    # 检查是否在主分支
    local current_branch
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    
    if [[ "$current_branch" != "main" && "$current_branch" != "master" && "$force" != "true" ]]; then
        log_error "发布必须在主分支进行，当前分支: $current_branch"
        exit 1
    fi
    
    # 检查工作目录是否干净
    if [[ -n "$(git status --porcelain)" && "$force" != "true" ]]; then
        log_error "工作目录不干净，请先提交或暂存更改"
        exit 1
    fi
    
    # 创建标签
    if [[ -n "$tag" ]]; then
        log_info "创建标签: $tag"
        git tag -a "$tag" -m "Release $version"
        git push origin "$tag"
    fi
    
    # 运行生产环境流水线
    run_pipeline "production" "$version" "false" "false" "false" "$force" "false" "true"
    
    log_success "版本发布完成: $version"
}

# 回滚部署
rollback_deployment() {
    local environment="$1"
    local revision="$2"
    local force="$3"
    
    log_info "回滚 $environment 环境部署..."
    
    cd "$PROJECT_ROOT"
    
    # 运行回滚脚本
    local rollback_args=("rollback")
    rollback_args+=("-e" "$environment")
    
    if [[ -n "$revision" ]]; then
        rollback_args+=("--revision" "$revision")
    fi
    
    if [[ "$force" == "true" ]]; then
        rollback_args+=("--force")
    fi
    
    if [[ -f "$SCRIPT_DIR/deploy.sh" ]]; then
        "$SCRIPT_DIR/deploy.sh" "${rollback_args[@]}"
    else
        log_error "部署脚本不存在: $SCRIPT_DIR/deploy.sh"
        exit 1
    fi
    
    log_success "回滚完成"
    send_notification "warning" "$environment" "rollback" "部署已回滚"
}

# 查看状态
show_status() {
    local environment="$1"
    
    log_info "查看 $environment 环境状态..."
    
    cd "$PROJECT_ROOT"
    
    # 运行状态查看脚本
    local status_args=("status")
    status_args+=("-e" "$environment")
    
    if [[ -f "$SCRIPT_DIR/deploy.sh" ]]; then
        "$SCRIPT_DIR/deploy.sh" "${status_args[@]}"
    else
        log_error "部署脚本不存在: $SCRIPT_DIR/deploy.sh"
        exit 1
    fi
}

# 清理环境
clean_environment() {
    log_info "清理环境..."
    
    # 清理 Docker 镜像
    docker system prune -f || true
    
    # 清理测试结果
    rm -rf "$PROJECT_ROOT/test-results" || true
    
    log_success "环境清理完成"
}

# 主函数
main() {
    # 默认参数
    local command=""
    local environment=""
    local version=""
    local branch=""
    local tag=""
    local registry="${DOCKER_REGISTRY:-ghcr.io}"
    local namespace=""
    local skip_tests="false"
    local skip_build="false"
    local skip_push="false"
    local force="false"
    local dry_run="false"
    local verbose="false"
    local revision=""
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            build|test|package|deploy|pipeline|release|rollback|status|clean|help)
                command="$1"
                shift
                ;;
            -e|--environment)
                environment="$2"
                shift 2
                ;;
            -v|--version)
                version="$2"
                shift 2
                ;;
            -b|--branch)
                branch="$2"
                shift 2
                ;;
            -t|--tag)
                tag="$2"
                shift 2
                ;;
            -r|--registry)
                registry="$2"
                shift 2
                ;;
            -n|--namespace)
                namespace="$2"
                shift 2
                ;;
            --skip-tests)
                skip_tests="true"
                shift
                ;;
            --skip-build)
                skip_build="true"
                shift
                ;;
            --skip-push)
                skip_push="true"
                shift
                ;;
            --force)
                force="true"
                shift
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            --verbose)
                verbose="true"
                set -x
                shift
                ;;
            --revision)
                revision="$2"
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
    
    # 如果没有指定命令，显示帮助
    if [[ -z "$command" ]]; then
        show_help
        exit 1
    fi
    
    # 检查依赖
    check_dependencies
    
    # 获取版本信息
    version=$(get_version_info "$version" "$branch" "$tag")
    
    # 设置环境变量
    if [[ -n "$environment" ]]; then
        setup_environment "$environment" "$version"
        
        # 如果没有指定命名空间，使用环境默认值
        if [[ -z "$namespace" ]]; then
            namespace="$K8S_NAMESPACE"
        fi
    fi
    
    # 执行命令
    case $command in
        build)
            build_application "$skip_build" "$version" "$verbose"
            ;;
        test)
            run_tests "$skip_tests" "$verbose"
            ;;
        package)
            build_application "false" "$version" "$verbose"
            push_image "$skip_push" "$registry" "$verbose"
            ;;
        deploy)
            if [[ -z "$environment" ]]; then
                log_error "必须指定环境 (-e|--environment)"
                exit 1
            fi
            deploy_application "$environment" "$version" "$namespace" "$force" "$dry_run"
            ;;
        pipeline)
            if [[ -z "$environment" ]]; then
                log_error "必须指定环境 (-e|--environment)"
                exit 1
            fi
            run_pipeline "$environment" "$version" "$skip_tests" "$skip_build" "$skip_push" "$force" "$dry_run" "$verbose"
            ;;
        release)
            release_version "$version" "$tag" "$force"
            ;;
        rollback)
            if [[ -z "$environment" ]]; then
                log_error "必须指定环境 (-e|--environment)"
                exit 1
            fi
            rollback_deployment "$environment" "$revision" "$force"
            ;;
        status)
            if [[ -z "$environment" ]]; then
                log_error "必须指定环境 (-e|--environment)"
                exit 1
            fi
            show_status "$environment"
            ;;
        clean)
            clean_environment
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