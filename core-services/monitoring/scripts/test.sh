#!/bin/bash

# 太上老君监控系统测试脚本
# 支持单元测试、集成测试、端到端测试和性能测试

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
太上老君监控系统测试脚本

用法: $0 [选项] [命令]

命令:
    unit        运行单元测试
    integration 运行集成测试
    e2e         运行端到端测试
    performance 运行性能测试
    coverage    生成测试覆盖率报告
    lint        运行代码检查
    security    运行安全扫描
    all         运行所有测试
    clean       清理测试环境
    help        显示此帮助信息

选项:
    -e, --environment ENV   测试环境 (local|docker|k8s) [默认: local]
    -p, --package PKG      指定测试包路径
    -t, --test TEST        指定测试函数名
    -c, --coverage         生成覆盖率报告
    -v, --verbose          详细输出
    -r, --race             启用竞态检测
    -b, --bench            运行基准测试
    -s, --short            运行短测试
    -f, --fail-fast        遇到失败立即停止
    --timeout TIMEOUT      测试超时时间 (默认: 10m)
    --parallel N           并行测试数量
    --output-dir DIR       输出目录 (默认: ./test-results)
    --docker-image IMAGE   Docker 镜像名称
    --k8s-namespace NS     Kubernetes 命名空间
    -h, --help             显示帮助信息

环境变量:
    TEST_DATABASE_URL      测试数据库连接字符串
    TEST_REDIS_URL         测试 Redis 连接字符串
    TEST_LOG_LEVEL         测试日志级别
    COVERAGE_THRESHOLD     覆盖率阈值 (默认: 80)

示例:
    $0 unit -v
    $0 integration -e docker
    $0 e2e -e k8s --k8s-namespace test
    $0 performance -b
    $0 coverage --output-dir ./coverage
    $0 all -c -v

EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 Go 版本
    local go_version
    go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+')
    log_info "Go 版本: $go_version"
    
    log_success "依赖检查完成"
}

# 检查 Docker 依赖
check_docker_dependencies() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装或不在 PATH 中"
        exit 1
    fi
}

# 检查 Kubernetes 依赖
check_k8s_dependencies() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查 kubectl 连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
}

# 设置测试环境
setup_test_environment() {
    local environment="$1"
    local output_dir="$2"
    
    log_info "设置测试环境: $environment"
    
    # 创建输出目录
    mkdir -p "$output_dir"
    
    # 设置环境变量
    export GO_ENV=test
    export LOG_LEVEL=debug
    
    case $environment in
        local)
            setup_local_environment
            ;;
        docker)
            setup_docker_environment
            ;;
        k8s)
            setup_k8s_environment
            ;;
        *)
            log_error "未知测试环境: $environment"
            exit 1
            ;;
    esac
}

# 设置本地测试环境
setup_local_environment() {
    log_info "设置本地测试环境..."
    
    # 设置测试数据库
    export TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgres://test:test@localhost:5432/monitoring_test?sslmode=disable}"
    export TEST_REDIS_URL="${TEST_REDIS_URL:-redis://localhost:6379/1}"
    
    # 检查测试数据库连接
    if command -v psql &> /dev/null; then
        if ! psql "$TEST_DATABASE_URL" -c "SELECT 1;" &> /dev/null; then
            log_warning "无法连接到测试数据库，某些测试可能会失败"
        fi
    fi
    
    # 检查测试 Redis 连接
    if command -v redis-cli &> /dev/null; then
        local redis_host redis_port
        redis_host=$(echo "$TEST_REDIS_URL" | sed -n 's|redis://\([^:]*\):.*|\1|p')
        redis_port=$(echo "$TEST_REDIS_URL" | sed -n 's|redis://[^:]*:\([0-9]*\)/.*|\1|p')
        
        if ! redis-cli -h "$redis_host" -p "$redis_port" ping &> /dev/null; then
            log_warning "无法连接到测试 Redis，某些测试可能会失败"
        fi
    fi
}

# 设置 Docker 测试环境
setup_docker_environment() {
    log_info "设置 Docker 测试环境..."
    
    check_docker_dependencies
    
    # 启动测试服务
    cd "$PROJECT_ROOT"
    
    if [[ -f "docker-compose.test.yml" ]]; then
        log_info "启动测试服务..."
        docker-compose -f docker-compose.test.yml up -d
        
        # 等待服务启动
        log_info "等待服务启动..."
        sleep 10
        
        # 设置环境变量
        export TEST_DATABASE_URL="postgres://test:test@localhost:5433/monitoring_test?sslmode=disable"
        export TEST_REDIS_URL="redis://localhost:6380/1"
    else
        log_warning "docker-compose.test.yml 不存在，使用默认配置"
        setup_local_environment
    fi
}

# 设置 Kubernetes 测试环境
setup_k8s_environment() {
    log_info "设置 Kubernetes 测试环境..."
    
    check_k8s_dependencies
    
    # 这里可以添加 Kubernetes 测试环境的设置逻辑
    # 例如：部署测试数据库、Redis 等
    
    log_warning "Kubernetes 测试环境设置尚未实现"
}

# 清理测试环境
cleanup_test_environment() {
    local environment="$1"
    
    log_info "清理测试环境: $environment"
    
    case $environment in
        docker)
            cd "$PROJECT_ROOT"
            if [[ -f "docker-compose.test.yml" ]]; then
                docker-compose -f docker-compose.test.yml down -v
            fi
            ;;
        k8s)
            # 清理 Kubernetes 测试资源
            log_info "清理 Kubernetes 测试资源..."
            ;;
    esac
}

# 运行单元测试
run_unit_tests() {
    local package="$1"
    local test_name="$2"
    local verbose="$3"
    local race="$4"
    local coverage="$5"
    local timeout="$6"
    local parallel="$7"
    local output_dir="$8"
    local short="$9"
    local fail_fast="${10}"
    
    log_info "运行单元测试..."
    
    cd "$PROJECT_ROOT"
    
    # 构建测试命令
    local test_args=("test")
    
    if [[ -n "$package" ]]; then
        test_args+=("$package")
    else
        test_args+=("./...")
    fi
    
    if [[ -n "$test_name" ]]; then
        test_args+=("-run" "$test_name")
    fi
    
    if [[ "$verbose" == "true" ]]; then
        test_args+=("-v")
    fi
    
    if [[ "$race" == "true" ]]; then
        test_args+=("-race")
    fi
    
    if [[ "$coverage" == "true" ]]; then
        test_args+=("-coverprofile=$output_dir/coverage.out")
        test_args+=("-covermode=atomic")
    fi
    
    if [[ "$short" == "true" ]]; then
        test_args+=("-short")
    fi
    
    if [[ "$fail_fast" == "true" ]]; then
        test_args+=("-failfast")
    fi
    
    test_args+=("-timeout" "$timeout")
    
    if [[ -n "$parallel" ]]; then
        test_args+=("-parallel" "$parallel")
    fi
    
    # 执行测试
    log_info "执行命令: go ${test_args[*]}"
    
    if go "${test_args[@]}"; then
        log_success "单元测试通过"
    else
        log_error "单元测试失败"
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    local package="$1"
    local test_name="$2"
    local verbose="$3"
    local timeout="$4"
    local output_dir="$5"
    local fail_fast="$6"
    
    log_info "运行集成测试..."
    
    cd "$PROJECT_ROOT"
    
    # 构建测试命令
    local test_args=("test")
    
    if [[ -n "$package" ]]; then
        test_args+=("$package")
    else
        test_args+=("./...")
    fi
    
    # 集成测试标签
    test_args+=("-tags=integration")
    
    if [[ -n "$test_name" ]]; then
        test_args+=("-run" "$test_name")
    fi
    
    if [[ "$verbose" == "true" ]]; then
        test_args+=("-v")
    fi
    
    if [[ "$fail_fast" == "true" ]]; then
        test_args+=("-failfast")
    fi
    
    test_args+=("-timeout" "$timeout")
    
    # 执行测试
    log_info "执行命令: go ${test_args[*]}"
    
    if go "${test_args[@]}"; then
        log_success "集成测试通过"
    else
        log_error "集成测试失败"
        return 1
    fi
}

# 运行端到端测试
run_e2e_tests() {
    local environment="$1"
    local k8s_namespace="$2"
    local verbose="$3"
    local timeout="$4"
    local output_dir="$5"
    
    log_info "运行端到端测试..."
    
    cd "$PROJECT_ROOT"
    
    # 设置环境变量
    if [[ "$environment" == "k8s" && -n "$k8s_namespace" ]]; then
        export E2E_NAMESPACE="$k8s_namespace"
    fi
    
    # 构建测试命令
    local test_args=("test" "./tests/e2e/...")
    test_args+=("-tags=e2e")
    
    if [[ "$verbose" == "true" ]]; then
        test_args+=("-v")
    fi
    
    test_args+=("-timeout" "$timeout")
    
    # 执行测试
    log_info "执行命令: go ${test_args[*]}"
    
    if go "${test_args[@]}"; then
        log_success "端到端测试通过"
    else
        log_error "端到端测试失败"
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    local package="$1"
    local bench="$2"
    local verbose="$3"
    local timeout="$4"
    local output_dir="$5"
    
    log_info "运行性能测试..."
    
    cd "$PROJECT_ROOT"
    
    # 构建测试命令
    local test_args=("test")
    
    if [[ -n "$package" ]]; then
        test_args+=("$package")
    else
        test_args+=("./...")
    fi
    
    if [[ "$bench" == "true" ]]; then
        test_args+=("-bench=.")
        test_args+=("-benchmem")
    fi
    
    if [[ "$verbose" == "true" ]]; then
        test_args+=("-v")
    fi
    
    test_args+=("-timeout" "$timeout")
    
    # 输出到文件
    local bench_output="$output_dir/benchmark.txt"
    
    # 执行测试
    log_info "执行命令: go ${test_args[*]}"
    
    if go "${test_args[@]}" | tee "$bench_output"; then
        log_success "性能测试完成，结果保存到: $bench_output"
    else
        log_error "性能测试失败"
        return 1
    fi
}

# 生成覆盖率报告
generate_coverage_report() {
    local output_dir="$1"
    local threshold="${2:-80}"
    
    log_info "生成覆盖率报告..."
    
    cd "$PROJECT_ROOT"
    
    local coverage_file="$output_dir/coverage.out"
    local coverage_html="$output_dir/coverage.html"
    
    if [[ ! -f "$coverage_file" ]]; then
        log_error "覆盖率文件不存在: $coverage_file"
        return 1
    fi
    
    # 生成 HTML 报告
    go tool cover -html="$coverage_file" -o "$coverage_html"
    log_success "HTML 覆盖率报告生成: $coverage_html"
    
    # 显示覆盖率统计
    local coverage_percent
    coverage_percent=$(go tool cover -func="$coverage_file" | tail -1 | awk '{print $3}' | sed 's/%//')
    
    log_info "总覆盖率: ${coverage_percent}%"
    
    # 检查覆盖率阈值
    if (( $(echo "$coverage_percent >= $threshold" | bc -l) )); then
        log_success "覆盖率达到阈值 (${threshold}%)"
    else
        log_warning "覆盖率未达到阈值 (${threshold}%)，当前: ${coverage_percent}%"
        return 1
    fi
}

# 运行代码检查
run_lint() {
    log_info "运行代码检查..."
    
    cd "$PROJECT_ROOT"
    
    # 检查 golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint 未安装，跳过代码检查"
        return 0
    fi
    
    # 运行 golangci-lint
    if golangci-lint run ./...; then
        log_success "代码检查通过"
    else
        log_error "代码检查失败"
        return 1
    fi
}

# 运行安全扫描
run_security_scan() {
    log_info "运行安全扫描..."
    
    cd "$PROJECT_ROOT"
    
    # 检查 gosec
    if ! command -v gosec &> /dev/null; then
        log_warning "gosec 未安装，跳过安全扫描"
        return 0
    fi
    
    # 运行 gosec
    if gosec ./...; then
        log_success "安全扫描通过"
    else
        log_error "安全扫描发现问题"
        return 1
    fi
}

# 主函数
main() {
    # 默认参数
    local command=""
    local environment="local"
    local package=""
    local test_name=""
    local coverage="false"
    local verbose="false"
    local race="false"
    local bench="false"
    local short="false"
    local fail_fast="false"
    local timeout="10m"
    local parallel=""
    local output_dir="./test-results"
    local docker_image=""
    local k8s_namespace=""
    local coverage_threshold="${COVERAGE_THRESHOLD:-80}"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            unit|integration|e2e|performance|coverage|lint|security|all|clean|help)
                command="$1"
                shift
                ;;
            -e|--environment)
                environment="$2"
                shift 2
                ;;
            -p|--package)
                package="$2"
                shift 2
                ;;
            -t|--test)
                test_name="$2"
                shift 2
                ;;
            -c|--coverage)
                coverage="true"
                shift
                ;;
            -v|--verbose)
                verbose="true"
                shift
                ;;
            -r|--race)
                race="true"
                shift
                ;;
            -b|--bench)
                bench="true"
                shift
                ;;
            -s|--short)
                short="true"
                shift
                ;;
            -f|--fail-fast)
                fail_fast="true"
                shift
                ;;
            --timeout)
                timeout="$2"
                shift 2
                ;;
            --parallel)
                parallel="$2"
                shift 2
                ;;
            --output-dir)
                output_dir="$2"
                shift 2
                ;;
            --docker-image)
                docker_image="$2"
                shift 2
                ;;
            --k8s-namespace)
                k8s_namespace="$2"
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
    
    # 设置测试环境
    setup_test_environment "$environment" "$output_dir"
    
    # 设置清理函数
    trap 'cleanup_test_environment "$environment"' EXIT
    
    # 执行命令
    case $command in
        unit)
            run_unit_tests "$package" "$test_name" "$verbose" "$race" "$coverage" "$timeout" "$parallel" "$output_dir" "$short" "$fail_fast"
            ;;
        integration)
            run_integration_tests "$package" "$test_name" "$verbose" "$timeout" "$output_dir" "$fail_fast"
            ;;
        e2e)
            run_e2e_tests "$environment" "$k8s_namespace" "$verbose" "$timeout" "$output_dir"
            ;;
        performance)
            run_performance_tests "$package" "$bench" "$verbose" "$timeout" "$output_dir"
            ;;
        coverage)
            # 先运行单元测试生成覆盖率
            run_unit_tests "$package" "$test_name" "$verbose" "$race" "true" "$timeout" "$parallel" "$output_dir" "$short" "$fail_fast"
            generate_coverage_report "$output_dir" "$coverage_threshold"
            ;;
        lint)
            run_lint
            ;;
        security)
            run_security_scan
            ;;
        all)
            log_info "运行所有测试..."
            
            # 运行代码检查
            run_lint || exit 1
            
            # 运行安全扫描
            run_security_scan || exit 1
            
            # 运行单元测试
            run_unit_tests "$package" "$test_name" "$verbose" "$race" "true" "$timeout" "$parallel" "$output_dir" "$short" "$fail_fast" || exit 1
            
            # 运行集成测试
            run_integration_tests "$package" "$test_name" "$verbose" "$timeout" "$output_dir" "$fail_fast" || exit 1
            
            # 生成覆盖率报告
            generate_coverage_report "$output_dir" "$coverage_threshold" || exit 1
            
            log_success "所有测试完成"
            ;;
        clean)
            log_info "清理测试环境..."
            cleanup_test_environment "$environment"
            rm -rf "$output_dir"
            log_success "清理完成"
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