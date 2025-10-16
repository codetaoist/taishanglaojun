#!/bin/bash

# 太上老君监控系统构建脚本
# 用于自动化构建、测试和部署流程

set -euo pipefail

# ==================== 变量定义 ====================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_NAME="taishanglaojun-monitoring"

# 版本信息
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
COMMIT_SHA="${COMMIT_SHA:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

# 构建配置
GO_VERSION="${GO_VERSION:-1.21}"
CGO_ENABLED="${CGO_ENABLED:-0}"
BUILD_DIR="${BUILD_DIR:-${PROJECT_ROOT}/build}"
DIST_DIR="${DIST_DIR:-${PROJECT_ROOT}/dist}"

# Docker 配置
REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_NAME="${IMAGE_NAME:-${REGISTRY}/taishanglaojun/monitoring}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

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

check_go_version() {
    local current_version
    current_version=$(go version | awk '{print $3}' | sed 's/go//')
    local required_version="${GO_VERSION}"
    
    if ! printf '%s\n%s\n' "$required_version" "$current_version" | sort -V -C; then
        log_warning "Go 版本 $current_version 可能不兼容，推荐版本 $required_version"
    fi
}

create_directories() {
    log_step "创建构建目录..."
    mkdir -p "${BUILD_DIR}" "${DIST_DIR}"
    log_success "目录创建完成"
}

# ==================== 环境检查 ====================
check_environment() {
    log_step "检查构建环境..."
    
    # 检查必要的命令
    check_command "go"
    check_command "git"
    
    # 检查 Go 版本
    check_go_version
    
    # 检查项目根目录
    if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
        log_error "未找到 go.mod 文件，请确保在正确的项目目录中运行"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# ==================== 依赖管理 ====================
download_dependencies() {
    log_step "下载依赖..."
    cd "${PROJECT_ROOT}"
    
    go mod download
    go mod tidy
    go mod verify
    
    log_success "依赖下载完成"
}

# ==================== 代码质量检查 ====================
format_code() {
    log_step "格式化代码..."
    cd "${PROJECT_ROOT}"
    
    gofmt -s -w .
    
    if command -v goimports &> /dev/null; then
        goimports -w .
    fi
    
    log_success "代码格式化完成"
}

lint_code() {
    log_step "运行代码检查..."
    cd "${PROJECT_ROOT}"
    
    # 运行 go vet
    go vet ./...
    
    # 运行 golangci-lint（如果可用）
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run --timeout=10m
    else
        log_warning "golangci-lint 未安装，跳过高级代码检查"
    fi
    
    log_success "代码检查通过"
}

# ==================== 测试 ====================
run_tests() {
    log_step "运行测试..."
    cd "${PROJECT_ROOT}"
    
    local coverage_dir="${PROJECT_ROOT}/coverage"
    mkdir -p "${coverage_dir}"
    
    # 运行单元测试
    go test -v -race -timeout=10m -coverprofile="${coverage_dir}/coverage.out" ./...
    
    # 生成覆盖率报告
    if [[ -f "${coverage_dir}/coverage.out" ]]; then
        go tool cover -html="${coverage_dir}/coverage.out" -o "${coverage_dir}/coverage.html"
        
        # 检查覆盖率
        local coverage
        coverage=$(go tool cover -func="${coverage_dir}/coverage.out" | tail -1 | awk '{print $3}' | sed 's/%//')
        
        log_info "测试覆盖率: ${coverage}%"
        
        if (( $(echo "$coverage < 80" | bc -l) )); then
            log_warning "测试覆盖率低于 80%"
        fi
    fi
    
    log_success "测试完成"
}

run_integration_tests() {
    log_step "运行集成测试..."
    cd "${PROJECT_ROOT}"
    
    if [[ -d "tests/integration" ]]; then
        go test -v -tags=integration -timeout=10m ./tests/integration/...
        log_success "集成测试完成"
    else
        log_warning "未找到集成测试目录，跳过集成测试"
    fi
}

# ==================== 安全检查 ====================
security_scan() {
    log_step "运行安全扫描..."
    cd "${PROJECT_ROOT}"
    
    # 运行 gosec（如果可用）
    if command -v gosec &> /dev/null; then
        gosec ./...
    else
        log_warning "gosec 未安装，跳过安全扫描"
    fi
    
    # 运行漏洞检查（如果可用）
    if command -v govulncheck &> /dev/null; then
        govulncheck ./...
    else
        log_warning "govulncheck 未安装，跳过漏洞检查"
    fi
    
    log_success "安全扫描完成"
}

# ==================== 构建 ====================
build_binary() {
    local os="${1:-$(go env GOOS)}"
    local arch="${2:-$(go env GOARCH)}"
    local output_name="${PROJECT_NAME}"
    
    if [[ "$os" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi
    
    local output_path="${DIST_DIR}/${PROJECT_NAME}-${os}-${arch}"
    if [[ "$os" == "windows" ]]; then
        output_path="${output_path}.exe"
    fi
    
    log_step "构建 ${os}/${arch} 版本..."
    
    cd "${PROJECT_ROOT}"
    
    CGO_ENABLED="${CGO_ENABLED}" GOOS="$os" GOARCH="$arch" go build \
        -ldflags="-X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA} -X main.BuildTime=${BUILD_TIME}" \
        -o "${output_path}" \
        ./cmd/monitoring
    
    log_success "构建完成: ${output_path}"
}

build_all_platforms() {
    log_step "构建所有平台版本..."
    
    create_directories
    
    # 构建不同平台的二进制文件
    local platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )
    
    for platform in "${platforms[@]}"; do
        local os="${platform%/*}"
        local arch="${platform#*/}"
        build_binary "$os" "$arch"
    done
    
    log_success "所有平台构建完成"
}

build_current_platform() {
    log_step "构建当前平台版本..."
    
    create_directories
    
    local current_os
    local current_arch
    current_os=$(go env GOOS)
    current_arch=$(go env GOARCH)
    
    build_binary "$current_os" "$current_arch"
    
    # 创建符号链接到 build 目录
    local output_name="${PROJECT_NAME}"
    if [[ "$current_os" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi
    
    ln -sf "${DIST_DIR}/${PROJECT_NAME}-${current_os}-${current_arch}" "${BUILD_DIR}/${output_name}" 2>/dev/null || \
    cp "${DIST_DIR}/${PROJECT_NAME}-${current_os}-${current_arch}" "${BUILD_DIR}/${output_name}"
    
    log_success "当前平台构建完成"
}

# ==================== Docker 构建 ====================
build_docker_image() {
    log_step "构建 Docker 镜像..."
    
    check_command "docker"
    
    cd "${PROJECT_ROOT}"
    
    docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg BUILD_TIME="${BUILD_TIME}" \
        --build-arg COMMIT_SHA="${COMMIT_SHA}" \
        -t "${IMAGE_NAME}:${VERSION}" \
        -t "${IMAGE_NAME}:latest" \
        .
    
    log_success "Docker 镜像构建完成"
}

build_docker_multiplatform() {
    log_step "构建多平台 Docker 镜像..."
    
    check_command "docker"
    
    # 检查 buildx 是否可用
    if ! docker buildx version &> /dev/null; then
        log_error "Docker buildx 不可用，无法构建多平台镜像"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
    
    docker buildx build \
        --platform "${PLATFORMS}" \
        --build-arg VERSION="${VERSION}" \
        --build-arg BUILD_TIME="${BUILD_TIME}" \
        --build-arg COMMIT_SHA="${COMMIT_SHA}" \
        -t "${IMAGE_NAME}:${VERSION}" \
        -t "${IMAGE_NAME}:latest" \
        --push \
        .
    
    log_success "多平台 Docker 镜像构建完成"
}

# ==================== 清理 ====================
clean_build() {
    log_step "清理构建文件..."
    
    rm -rf "${BUILD_DIR}" "${DIST_DIR}" "${PROJECT_ROOT}/coverage"
    go clean -cache -testcache -modcache
    
    log_success "清理完成"
}

# ==================== 版本信息 ====================
show_version() {
    echo -e "${CYAN}太上老君监控系统构建信息${RESET}"
    echo "版本: ${VERSION}"
    echo "提交: ${COMMIT_SHA}"
    echo "构建时间: ${BUILD_TIME}"
    echo "Go 版本: $(go version | awk '{print $3}')"
    echo "平台: $(go env GOOS)/$(go env GOARCH)"
}

# ==================== 帮助信息 ====================
show_help() {
    cat << EOF
太上老君监控系统构建脚本

用法: $0 [命令] [选项]

命令:
  check           检查构建环境
  deps            下载依赖
  format          格式化代码
  lint            代码检查
  test            运行测试
  test-integration 运行集成测试
  security        安全扫描
  build           构建当前平台版本
  build-all       构建所有平台版本
  docker          构建 Docker 镜像
  docker-multi    构建多平台 Docker 镜像
  clean           清理构建文件
  version         显示版本信息
  ci              运行 CI 流程 (check + deps + format + lint + test + build)
  cd              运行 CD 流程 (docker + push)
  all             运行完整流程 (ci + docker)
  help            显示帮助信息

环境变量:
  VERSION         版本号 (默认: git describe)
  COMMIT_SHA      提交哈希 (默认: git rev-parse)
  BUILD_TIME      构建时间 (默认: 当前时间)
  REGISTRY        Docker 注册表 (默认: ghcr.io)
  IMAGE_NAME      镜像名称 (默认: ghcr.io/taishanglaojun/monitoring)
  PLATFORMS       构建平台 (默认: linux/amd64,linux/arm64)

示例:
  $0 build                    # 构建当前平台版本
  $0 build-all                # 构建所有平台版本
  $0 docker                   # 构建 Docker 镜像
  VERSION=v1.0.0 $0 build     # 指定版本构建
  $0 ci                       # 运行 CI 流程

EOF
}

# ==================== 主函数 ====================
main() {
    local command="${1:-help}"
    
    case "$command" in
        check)
            check_environment
            ;;
        deps)
            check_environment
            download_dependencies
            ;;
        format)
            check_environment
            format_code
            ;;
        lint)
            check_environment
            lint_code
            ;;
        test)
            check_environment
            run_tests
            ;;
        test-integration)
            check_environment
            run_integration_tests
            ;;
        security)
            check_environment
            security_scan
            ;;
        build)
            check_environment
            download_dependencies
            build_current_platform
            ;;
        build-all)
            check_environment
            download_dependencies
            build_all_platforms
            ;;
        docker)
            check_environment
            build_docker_image
            ;;
        docker-multi)
            check_environment
            build_docker_multiplatform
            ;;
        clean)
            clean_build
            ;;
        version)
            show_version
            ;;
        ci)
            check_environment
            download_dependencies
            format_code
            lint_code
            run_tests
            build_current_platform
            ;;
        cd)
            check_environment
            build_docker_image
            ;;
        all)
            check_environment
            download_dependencies
            format_code
            lint_code
            run_tests
            security_scan
            build_current_platform
            build_docker_image
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