#!/bin/bash

# 太上老君监控系统健康检查脚本
# 用于检查应用和依赖服务的健康状态

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
太上老君监控系统健康检查脚本

用法: $0 [选项] [命令]

命令:
    check       执行健康检查
    app         检查应用健康状态
    deps        检查依赖服务
    db          检查数据库连接
    redis       检查 Redis 连接
    metrics     检查指标端点
    logs        检查日志输出
    all         执行所有检查
    help        显示此帮助信息

选项:
    -e, --environment ENV   环境 (development|staging|production)
    -h, --host HOST        主机地址 (默认: localhost)
    -p, --port PORT        端口号 (默认: 8080)
    -t, --timeout TIMEOUT  超时时间 (默认: 30s)
    -r, --retry RETRY      重试次数 (默认: 3)
    -i, --interval INTERVAL 重试间隔 (默认: 5s)
    -v, --verbose          详细输出
    --json                 JSON 格式输出
    --no-color             禁用颜色输出
    --help                 显示帮助信息

环境变量:
    MONITORING_HOST        监控服务主机
    MONITORING_PORT        监控服务端口
    DATABASE_URL           数据库连接地址
    REDIS_URL              Redis 连接地址
    HEALTH_CHECK_TIMEOUT   健康检查超时时间

示例:
    $0 check
    $0 app -h localhost -p 8080
    $0 deps -e production
    $0 all --json

EOF
}

# 解析超时时间
parse_timeout() {
    local timeout="$1"
    
    # 如果已经是数字，直接返回
    if [[ "$timeout" =~ ^[0-9]+$ ]]; then
        echo "$timeout"
        return
    fi
    
    # 解析时间单位
    local value="${timeout%[a-zA-Z]*}"
    local unit="${timeout#$value}"
    
    case "$unit" in
        s|sec|second|seconds)
            echo "$value"
            ;;
        m|min|minute|minutes)
            echo $((value * 60))
            ;;
        h|hour|hours)
            echo $((value * 3600))
            ;;
        *)
            echo "30"  # 默认 30 秒
            ;;
    esac
}

# HTTP 请求函数
http_request() {
    local url="$1"
    local timeout="$2"
    local method="${3:-GET}"
    local headers="${4:-}"
    
    local curl_cmd=("curl" "-s" "-f" "--max-time" "$timeout" "-X" "$method")
    
    if [[ -n "$headers" ]]; then
        IFS=',' read -ra header_array <<< "$headers"
        for header in "${header_array[@]}"; do
            curl_cmd+=("-H" "$header")
        done
    fi
    
    curl_cmd+=("$url")
    
    "${curl_cmd[@]}"
}

# 检查应用健康状态
check_app_health() {
    local host="$1"
    local port="$2"
    local timeout="$3"
    local verbose="$4"
    local json_output="$5"
    
    log_info "检查应用健康状态..."
    
    local health_url="http://${host}:${port}/health"
    local ready_url="http://${host}:${port}/ready"
    local live_url="http://${host}:${port}/live"
    
    local health_status="unknown"
    local ready_status="unknown"
    local live_status="unknown"
    local response_time=0
    
    # 检查健康端点
    local start_time
    start_time=$(date +%s.%N)
    
    if response=$(http_request "$health_url" "$timeout" "GET" 2>/dev/null); then
        health_status="healthy"
        if [[ "$verbose" == "true" ]]; then
            log_info "健康检查响应: $response"
        fi
    else
        health_status="unhealthy"
        log_error "健康检查失败: $health_url"
    fi
    
    local end_time
    end_time=$(date +%s.%N)
    response_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    
    # 检查就绪端点
    if http_request "$ready_url" "$timeout" "GET" >/dev/null 2>&1; then
        ready_status="ready"
    else
        ready_status="not_ready"
        log_warning "就绪检查失败: $ready_url"
    fi
    
    # 检查存活端点
    if http_request "$live_url" "$timeout" "GET" >/dev/null 2>&1; then
        live_status="alive"
    else
        live_status="dead"
        log_error "存活检查失败: $live_url"
    fi
    
    # 输出结果
    if [[ "$json_output" == "true" ]]; then
        cat << EOF
{
    "app": {
        "health": "$health_status",
        "ready": "$ready_status",
        "live": "$live_status",
        "response_time": $response_time,
        "endpoints": {
            "health": "$health_url",
            "ready": "$ready_url",
            "live": "$live_url"
        }
    }
}
EOF
    else
        echo "应用健康状态:"
        echo "  健康: $health_status"
        echo "  就绪: $ready_status"
        echo "  存活: $live_status"
        echo "  响应时间: ${response_time}s"
    fi
    
    # 返回状态
    if [[ "$health_status" == "healthy" && "$ready_status" == "ready" && "$live_status" == "alive" ]]; then
        return 0
    else
        return 1
    fi
}

# 检查数据库连接
check_database() {
    local timeout="$1"
    local verbose="$2"
    local json_output="$3"
    
    log_info "检查数据库连接..."
    
    local db_status="unknown"
    local db_version=""
    local connection_time=0
    
    # 获取数据库连接信息
    local db_url="${DATABASE_URL:-}"
    
    if [[ -z "$db_url" ]]; then
        db_status="not_configured"
        log_warning "数据库连接未配置"
    else
        local start_time
        start_time=$(date +%s.%N)
        
        # 尝试连接数据库
        if command -v psql &> /dev/null && [[ "$db_url" =~ postgres ]]; then
            # PostgreSQL
            if db_version=$(psql "$db_url" -t -c "SELECT version();" 2>/dev/null | head -1 | xargs); then
                db_status="connected"
                if [[ "$verbose" == "true" ]]; then
                    log_info "数据库版本: $db_version"
                fi
            else
                db_status="connection_failed"
                log_error "PostgreSQL 连接失败"
            fi
        elif command -v mysql &> /dev/null && [[ "$db_url" =~ mysql ]]; then
            # MySQL
            if db_version=$(mysql --version 2>/dev/null | head -1); then
                db_status="connected"
                if [[ "$verbose" == "true" ]]; then
                    log_info "数据库版本: $db_version"
                fi
            else
                db_status="connection_failed"
                log_error "MySQL 连接失败"
            fi
        else
            db_status="unsupported"
            log_warning "不支持的数据库类型或客户端未安装"
        fi
        
        local end_time
        end_time=$(date +%s.%N)
        connection_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    fi
    
    # 输出结果
    if [[ "$json_output" == "true" ]]; then
        cat << EOF
{
    "database": {
        "status": "$db_status",
        "version": "$db_version",
        "connection_time": $connection_time
    }
}
EOF
    else
        echo "数据库状态:"
        echo "  状态: $db_status"
        echo "  版本: $db_version"
        echo "  连接时间: ${connection_time}s"
    fi
    
    # 返回状态
    if [[ "$db_status" == "connected" ]]; then
        return 0
    else
        return 1
    fi
}

# 检查 Redis 连接
check_redis() {
    local timeout="$1"
    local verbose="$2"
    local json_output="$3"
    
    log_info "检查 Redis 连接..."
    
    local redis_status="unknown"
    local redis_version=""
    local connection_time=0
    
    # 获取 Redis 连接信息
    local redis_url="${REDIS_URL:-}"
    
    if [[ -z "$redis_url" ]]; then
        redis_status="not_configured"
        log_warning "Redis 连接未配置"
    else
        local start_time
        start_time=$(date +%s.%N)
        
        # 尝试连接 Redis
        if command -v redis-cli &> /dev/null; then
            if redis_version=$(redis-cli --version 2>/dev/null | head -1); then
                # 测试 PING 命令
                if redis-cli -u "$redis_url" ping >/dev/null 2>&1; then
                    redis_status="connected"
                    if [[ "$verbose" == "true" ]]; then
                        log_info "Redis 版本: $redis_version"
                    fi
                else
                    redis_status="connection_failed"
                    log_error "Redis PING 失败"
                fi
            else
                redis_status="client_error"
                log_error "Redis 客户端错误"
            fi
        else
            redis_status="client_not_found"
            log_warning "Redis 客户端未安装"
        fi
        
        local end_time
        end_time=$(date +%s.%N)
        connection_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    fi
    
    # 输出结果
    if [[ "$json_output" == "true" ]]; then
        cat << EOF
{
    "redis": {
        "status": "$redis_status",
        "version": "$redis_version",
        "connection_time": $connection_time
    }
}
EOF
    else
        echo "Redis 状态:"
        echo "  状态: $redis_status"
        echo "  版本: $redis_version"
        echo "  连接时间: ${connection_time}s"
    fi
    
    # 返回状态
    if [[ "$redis_status" == "connected" ]]; then
        return 0
    else
        return 1
    fi
}

# 检查指标端点
check_metrics() {
    local host="$1"
    local port="$2"
    local timeout="$3"
    local verbose="$4"
    local json_output="$5"
    
    log_info "检查指标端点..."
    
    local metrics_url="http://${host}:9090/metrics"
    local metrics_status="unknown"
    local metrics_count=0
    local response_time=0
    
    local start_time
    start_time=$(date +%s.%N)
    
    if response=$(http_request "$metrics_url" "$timeout" "GET" 2>/dev/null); then
        metrics_status="available"
        metrics_count=$(echo "$response" | grep -c "^[a-zA-Z]" || echo "0")
        
        if [[ "$verbose" == "true" ]]; then
            log_info "指标数量: $metrics_count"
        fi
    else
        metrics_status="unavailable"
        log_error "指标端点不可用: $metrics_url"
    fi
    
    local end_time
    end_time=$(date +%s.%N)
    response_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    
    # 输出结果
    if [[ "$json_output" == "true" ]]; then
        cat << EOF
{
    "metrics": {
        "status": "$metrics_status",
        "count": $metrics_count,
        "response_time": $response_time,
        "endpoint": "$metrics_url"
    }
}
EOF
    else
        echo "指标状态:"
        echo "  状态: $metrics_status"
        echo "  指标数量: $metrics_count"
        echo "  响应时间: ${response_time}s"
    fi
    
    # 返回状态
    if [[ "$metrics_status" == "available" ]]; then
        return 0
    else
        return 1
    fi
}

# 检查日志输出
check_logs() {
    local environment="$1"
    local verbose="$2"
    local json_output="$3"
    
    log_info "检查日志输出..."
    
    local log_status="unknown"
    local log_lines=0
    local error_count=0
    local warning_count=0
    
    # 根据环境确定日志路径
    local log_path=""
    case "$environment" in
        development)
            log_path="/tmp/monitoring.log"
            ;;
        staging|production)
            log_path="/var/log/monitoring/app.log"
            ;;
        *)
            log_path="./logs/app.log"
            ;;
    esac
    
    if [[ -f "$log_path" ]]; then
        log_status="available"
        log_lines=$(wc -l < "$log_path" 2>/dev/null || echo "0")
        error_count=$(grep -c "ERROR" "$log_path" 2>/dev/null || echo "0")
        warning_count=$(grep -c "WARNING\|WARN" "$log_path" 2>/dev/null || echo "0")
        
        if [[ "$verbose" == "true" ]]; then
            log_info "日志文件: $log_path"
            log_info "日志行数: $log_lines"
            log_info "错误数量: $error_count"
            log_info "警告数量: $warning_count"
        fi
    else
        log_status="not_found"
        log_warning "日志文件不存在: $log_path"
    fi
    
    # 输出结果
    if [[ "$json_output" == "true" ]]; then
        cat << EOF
{
    "logs": {
        "status": "$log_status",
        "path": "$log_path",
        "lines": $log_lines,
        "errors": $error_count,
        "warnings": $warning_count
    }
}
EOF
    else
        echo "日志状态:"
        echo "  状态: $log_status"
        echo "  路径: $log_path"
        echo "  行数: $log_lines"
        echo "  错误: $error_count"
        echo "  警告: $warning_count"
    fi
    
    # 返回状态
    if [[ "$log_status" == "available" ]]; then
        return 0
    else
        return 1
    fi
}

# 执行所有检查
check_all() {
    local environment="$1"
    local host="$2"
    local port="$3"
    local timeout="$4"
    local verbose="$5"
    local json_output="$6"
    
    log_info "执行所有健康检查..."
    
    local results=()
    local overall_status="healthy"
    
    # 检查应用
    if check_app_health "$host" "$port" "$timeout" "$verbose" "false"; then
        results+=("app:healthy")
    else
        results+=("app:unhealthy")
        overall_status="unhealthy"
    fi
    
    # 检查数据库
    if check_database "$timeout" "$verbose" "false"; then
        results+=("database:healthy")
    else
        results+=("database:unhealthy")
        overall_status="unhealthy"
    fi
    
    # 检查 Redis
    if check_redis "$timeout" "$verbose" "false"; then
        results+=("redis:healthy")
    else
        results+=("redis:unhealthy")
        overall_status="unhealthy"
    fi
    
    # 检查指标
    if check_metrics "$host" "$port" "$timeout" "$verbose" "false"; then
        results+=("metrics:healthy")
    else
        results+=("metrics:unhealthy")
        overall_status="unhealthy"
    fi
    
    # 检查日志
    if check_logs "$environment" "$verbose" "false"; then
        results+=("logs:healthy")
    else
        results+=("logs:unhealthy")
        overall_status="unhealthy"
    fi
    
    # 输出结果
    if [[ "$json_output" == "true" ]]; then
        echo "{"
        echo "  \"overall_status\": \"$overall_status\","
        echo "  \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\","
        echo "  \"checks\": {"
        
        local first=true
        for result in "${results[@]}"; do
            local check_name="${result%:*}"
            local check_status="${result#*:}"
            
            if [[ "$first" == "true" ]]; then
                first=false
            else
                echo ","
            fi
            
            echo -n "    \"$check_name\": \"$check_status\""
        done
        
        echo ""
        echo "  }"
        echo "}"
    else
        echo ""
        echo "========================================="
        echo "健康检查汇总"
        echo "========================================="
        echo "总体状态: $overall_status"
        echo "检查时间: $(date)"
        echo ""
        
        for result in "${results[@]}"; do
            local check_name="${result%:*}"
            local check_status="${result#*:}"
            echo "$check_name: $check_status"
        done
        
        echo "========================================="
    fi
    
    # 返回状态
    if [[ "$overall_status" == "healthy" ]]; then
        return 0
    else
        return 1
    fi
}

# 主函数
main() {
    # 默认参数
    local command=""
    local environment=""
    local host="${MONITORING_HOST:-localhost}"
    local port="${MONITORING_PORT:-8080}"
    local timeout="${HEALTH_CHECK_TIMEOUT:-30s}"
    local retry="3"
    local interval="5s"
    local verbose="false"
    local json_output="false"
    local no_color="false"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            check|app|deps|db|redis|metrics|logs|all|help)
                command="$1"
                shift
                ;;
            -e|--environment)
                environment="$2"
                shift 2
                ;;
            -h|--host)
                host="$2"
                shift 2
                ;;
            -p|--port)
                port="$2"
                shift 2
                ;;
            -t|--timeout)
                timeout="$2"
                shift 2
                ;;
            -r|--retry)
                retry="$2"
                shift 2
                ;;
            -i|--interval)
                interval="$2"
                shift 2
                ;;
            -v|--verbose)
                verbose="true"
                shift
                ;;
            --json)
                json_output="true"
                shift
                ;;
            --no-color)
                no_color="true"
                shift
                ;;
            --help)
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
    
    # 如果没有指定命令，默认为 check
    if [[ -z "$command" ]]; then
        command="check"
    fi
    
    # 禁用颜色输出
    if [[ "$no_color" == "true" ]]; then
        RED=""
        GREEN=""
        YELLOW=""
        BLUE=""
        NC=""
    fi
    
    # 解析超时时间
    timeout=$(parse_timeout "$timeout")
    interval=$(parse_timeout "$interval")
    
    # 执行命令
    local exit_code=0
    local attempt=1
    
    while [[ $attempt -le $retry ]]; do
        if [[ $attempt -gt 1 ]]; then
            log_info "重试 $attempt/$retry..."
            sleep "$interval"
        fi
        
        case $command in
            check|app)
                if check_app_health "$host" "$port" "$timeout" "$verbose" "$json_output"; then
                    break
                fi
                ;;
            deps)
                local deps_healthy=true
                if ! check_database "$timeout" "$verbose" "$json_output"; then
                    deps_healthy=false
                fi
                if ! check_redis "$timeout" "$verbose" "$json_output"; then
                    deps_healthy=false
                fi
                if [[ "$deps_healthy" == "true" ]]; then
                    break
                fi
                ;;
            db)
                if check_database "$timeout" "$verbose" "$json_output"; then
                    break
                fi
                ;;
            redis)
                if check_redis "$timeout" "$verbose" "$json_output"; then
                    break
                fi
                ;;
            metrics)
                if check_metrics "$host" "$port" "$timeout" "$verbose" "$json_output"; then
                    break
                fi
                ;;
            logs)
                if check_logs "$environment" "$verbose" "$json_output"; then
                    break
                fi
                ;;
            all)
                if check_all "$environment" "$host" "$port" "$timeout" "$verbose" "$json_output"; then
                    break
                fi
                ;;
            help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知命令: $command"
                show_help
                exit 1
                ;;
        esac
        
        exit_code=1
        ((attempt++))
    done
    
    if [[ $exit_code -ne 0 ]]; then
        log_error "健康检查失败，已重试 $retry 次"
    fi
    
    exit $exit_code
}

# 执行主函数
main "$@"