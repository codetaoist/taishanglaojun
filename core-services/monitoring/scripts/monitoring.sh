#!/bin/bash

# 太上老君监控系统管理脚本
# 用于管理监控系统的启动、停止、状态检查等操作

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICE_NAME="monitoring"

# 默认配置
DEFAULT_ENVIRONMENT="development"
DEFAULT_PORT=8080
DEFAULT_METRICS_PORT=9090
DEFAULT_LOG_LEVEL="info"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

log_debug() {
    if [[ "${VERBOSE:-false}" == "true" ]]; then
        echo -e "${CYAN}[DEBUG]${NC} $1"
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
太上老君监控系统管理脚本

用法: $0 [选项] [命令]

命令:
    start       启动监控服务
    stop        停止监控服务
    restart     重启监控服务
    status      查看服务状态
    logs        查看服务日志
    health      健康检查
    metrics     查看指标
    config      配置管理
    install     安装服务
    uninstall   卸载服务
    update      更新服务
    backup      备份数据
    restore     恢复数据
    help        显示此帮助信息

选项:
    -e, --env ENV           环境 (development|staging|production) [默认: $DEFAULT_ENVIRONMENT]
    -p, --port PORT         HTTP 端口 [默认: $DEFAULT_PORT]
    -m, --metrics-port PORT 指标端口 [默认: $DEFAULT_METRICS_PORT]
    -l, --log-level LEVEL   日志级别 (debug|info|warn|error) [默认: $DEFAULT_LOG_LEVEL]
    -c, --config FILE       配置文件路径
    -d, --daemon            后台运行
    -f, --follow            跟踪日志输出
    -n, --lines NUM         日志行数
    -t, --timeout SECONDS   超时时间
    -v, --verbose           详细输出
    -q, --quiet             静默模式
    --dry-run               试运行模式
    --force                 强制执行
    --no-color              禁用颜色输出

环境变量:
    MONITORING_ENV          监控环境
    MONITORING_PORT         HTTP 端口
    MONITORING_METRICS_PORT 指标端口
    MONITORING_LOG_LEVEL    日志级别
    MONITORING_CONFIG       配置文件路径
    DATABASE_URL            数据库连接地址
    REDIS_URL               Redis 连接地址

示例:
    $0 start
    $0 start --env production --port 8080
    $0 stop --force
    $0 restart --daemon
    $0 status
    $0 logs --follow --lines 100
    $0 health
    $0 metrics

EOF
}

# 解析命令行参数
parse_args() {
    COMMAND=""
    ENVIRONMENT="${MONITORING_ENV:-$DEFAULT_ENVIRONMENT}"
    PORT="${MONITORING_PORT:-$DEFAULT_PORT}"
    METRICS_PORT="${MONITORING_METRICS_PORT:-$DEFAULT_METRICS_PORT}"
    LOG_LEVEL="${MONITORING_LOG_LEVEL:-$DEFAULT_LOG_LEVEL}"
    CONFIG_FILE="${MONITORING_CONFIG:-}"
    DAEMON=false
    FOLLOW=false
    LINES=50
    TIMEOUT=30
    VERBOSE=false
    QUIET=false
    DRY_RUN=false
    FORCE=false
    NO_COLOR=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            start|stop|restart|status|logs|health|metrics|config|install|uninstall|update|backup|restore|help)
                COMMAND="$1"
                shift
                ;;
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            -m|--metrics-port)
                METRICS_PORT="$2"
                shift 2
                ;;
            -l|--log-level)
                LOG_LEVEL="$2"
                shift 2
                ;;
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -d|--daemon)
                DAEMON=true
                shift
                ;;
            -f|--follow)
                FOLLOW=true
                shift
                ;;
            -n|--lines)
                LINES="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -q|--quiet)
                QUIET=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --no-color)
                NO_COLOR=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 默认命令
    if [[ -z "$COMMAND" ]]; then
        COMMAND="status"
    fi

    # 禁用颜色
    if [[ "$NO_COLOR" == "true" ]]; then
        RED=""
        GREEN=""
        YELLOW=""
        BLUE=""
        CYAN=""
        NC=""
    fi

    # 静默模式
    if [[ "$QUIET" == "true" ]]; then
        VERBOSE=false
    fi

    # 验证参数
    case "$ENVIRONMENT" in
        development|staging|production) ;;
        *)
            log_error "无效的环境: $ENVIRONMENT"
            exit 1
            ;;
    esac

    case "$LOG_LEVEL" in
        debug|info|warn|error) ;;
        *)
            log_error "无效的日志级别: $LOG_LEVEL"
            exit 1
            ;;
    esac

    if [[ ! "$PORT" =~ ^[0-9]+$ ]] || [[ "$PORT" -lt 1 ]] || [[ "$PORT" -gt 65535 ]]; then
        log_error "无效的端口号: $PORT"
        exit 1
    fi

    if [[ ! "$METRICS_PORT" =~ ^[0-9]+$ ]] || [[ "$METRICS_PORT" -lt 1 ]] || [[ "$METRICS_PORT" -gt 65535 ]]; then
        log_error "无效的指标端口号: $METRICS_PORT"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    local deps=("curl" "jq")
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_warning "依赖未安装: $dep"
        fi
    done
}

# 获取 PID 文件路径
get_pid_file() {
    echo "/var/run/${SERVICE_NAME}.pid"
}

# 获取日志文件路径
get_log_file() {
    case "$ENVIRONMENT" in
        development)
            echo "$PROJECT_ROOT/logs/app.log"
            ;;
        staging|production)
            echo "/var/log/${SERVICE_NAME}/app.log"
            ;;
    esac
}

# 获取配置文件路径
get_config_file() {
    if [[ -n "$CONFIG_FILE" ]]; then
        echo "$CONFIG_FILE"
    else
        echo "$PROJECT_ROOT/config/${ENVIRONMENT}.yml"
    fi
}

# 检查服务是否运行
is_service_running() {
    local pid_file=$(get_pid_file)
    
    if [[ -f "$pid_file" ]]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            return 0
        else
            # PID 文件存在但进程不存在，清理 PID 文件
            rm -f "$pid_file"
            return 1
        fi
    else
        return 1
    fi
}

# 获取服务 PID
get_service_pid() {
    local pid_file=$(get_pid_file)
    
    if [[ -f "$pid_file" ]]; then
        cat "$pid_file"
    else
        echo ""
    fi
}

# 等待服务启动
wait_for_service() {
    local timeout="$1"
    local start_time=$(date +%s)
    
    log_info "等待服务启动..."
    
    while true; do
        if is_service_running; then
            # 检查 HTTP 端点
            if curl -s "http://localhost:${PORT}/health" >/dev/null 2>&1; then
                log_success "服务启动成功"
                return 0
            fi
        fi
        
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        
        if [[ $elapsed -ge $timeout ]]; then
            log_error "服务启动超时"
            return 1
        fi
        
        sleep 1
    done
}

# 等待服务停止
wait_for_service_stop() {
    local timeout="$1"
    local start_time=$(date +%s)
    
    log_info "等待服务停止..."
    
    while true; do
        if ! is_service_running; then
            log_success "服务停止成功"
            return 0
        fi
        
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        
        if [[ $elapsed -ge $timeout ]]; then
            log_error "服务停止超时"
            return 1
        fi
        
        sleep 1
    done
}

# 启动服务
start_service() {
    log_info "启动监控服务..."
    
    if is_service_running; then
        log_warning "服务已在运行中"
        return 0
    fi
    
    # 检查配置文件
    local config_file=$(get_config_file)
    if [[ ! -f "$config_file" ]]; then
        log_error "配置文件不存在: $config_file"
        exit 1
    fi
    
    # 创建日志目录
    local log_file=$(get_log_file)
    local log_dir=$(dirname "$log_file")
    if [[ ! -d "$log_dir" ]]; then
        mkdir -p "$log_dir"
    fi
    
    # 设置环境变量
    export MONITORING_ENV="$ENVIRONMENT"
    export MONITORING_PORT="$PORT"
    export MONITORING_METRICS_PORT="$METRICS_PORT"
    export MONITORING_LOG_LEVEL="$LOG_LEVEL"
    export MONITORING_CONFIG="$config_file"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 启动服务"
        log_info "环境: $ENVIRONMENT"
        log_info "端口: $PORT"
        log_info "指标端口: $METRICS_PORT"
        log_info "日志级别: $LOG_LEVEL"
        log_info "配置文件: $config_file"
        return 0
    fi
    
    # 启动服务
    if [[ "$DAEMON" == "true" ]]; then
        # 后台运行
        nohup "$PROJECT_ROOT/bin/monitoring" > "$log_file" 2>&1 &
        local pid=$!
        echo "$pid" > "$(get_pid_file)"
        log_info "服务已在后台启动，PID: $pid"
        
        # 等待服务启动
        if ! wait_for_service "$TIMEOUT"; then
            stop_service
            exit 1
        fi
    else
        # 前台运行
        exec "$PROJECT_ROOT/bin/monitoring"
    fi
}

# 停止服务
stop_service() {
    log_info "停止监控服务..."
    
    if ! is_service_running; then
        log_warning "服务未运行"
        return 0
    fi
    
    local pid=$(get_service_pid)
    local pid_file=$(get_pid_file)
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 停止服务，PID: $pid"
        return 0
    fi
    
    # 发送 TERM 信号
    log_info "发送停止信号到进程 $pid"
    kill -TERM "$pid" 2>/dev/null || true
    
    # 等待服务停止
    if wait_for_service_stop "$TIMEOUT"; then
        rm -f "$pid_file"
        return 0
    fi
    
    # 强制停止
    if [[ "$FORCE" == "true" ]]; then
        log_warning "强制停止服务"
        kill -KILL "$pid" 2>/dev/null || true
        rm -f "$pid_file"
        log_success "服务已强制停止"
    else
        log_error "服务停止失败，使用 --force 强制停止"
        exit 1
    fi
}

# 重启服务
restart_service() {
    log_info "重启监控服务..."
    
    if is_service_running; then
        stop_service
    fi
    
    start_service
}

# 查看服务状态
show_status() {
    log_info "查看服务状态..."
    
    if is_service_running; then
        local pid=$(get_service_pid)
        log_success "服务正在运行，PID: $pid"
        
        # 显示详细信息
        if [[ "$VERBOSE" == "true" ]]; then
            echo ""
            echo "详细信息:"
            echo "  环境: $ENVIRONMENT"
            echo "  端口: $PORT"
            echo "  指标端口: $METRICS_PORT"
            echo "  日志级别: $LOG_LEVEL"
            echo "  配置文件: $(get_config_file)"
            echo "  日志文件: $(get_log_file)"
            echo "  PID 文件: $(get_pid_file)"
            
            # 显示进程信息
            if command -v ps &> /dev/null; then
                echo ""
                echo "进程信息:"
                ps -p "$pid" -o pid,ppid,user,start,time,command 2>/dev/null || true
            fi
            
            # 显示端口监听
            if command -v netstat &> /dev/null; then
                echo ""
                echo "端口监听:"
                netstat -tlnp 2>/dev/null | grep ":$PORT " || true
                netstat -tlnp 2>/dev/null | grep ":$METRICS_PORT " || true
            elif command -v ss &> /dev/null; then
                echo ""
                echo "端口监听:"
                ss -tlnp | grep ":$PORT " || true
                ss -tlnp | grep ":$METRICS_PORT " || true
            fi
        fi
    else
        log_error "服务未运行"
        exit 1
    fi
}

# 查看日志
show_logs() {
    local log_file=$(get_log_file)
    
    if [[ ! -f "$log_file" ]]; then
        log_error "日志文件不存在: $log_file"
        exit 1
    fi
    
    log_info "查看服务日志: $log_file"
    
    if [[ "$FOLLOW" == "true" ]]; then
        tail -f -n "$LINES" "$log_file"
    else
        tail -n "$LINES" "$log_file"
    fi
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    if ! is_service_running; then
        log_error "服务未运行"
        exit 1
    fi
    
    # 检查 HTTP 端点
    local health_url="http://localhost:${PORT}/health"
    local ready_url="http://localhost:${PORT}/ready"
    local live_url="http://localhost:${PORT}/live"
    
    log_info "检查健康端点: $health_url"
    if curl -s -f "$health_url" >/dev/null; then
        log_success "健康检查通过"
    else
        log_error "健康检查失败"
        exit 1
    fi
    
    log_info "检查就绪端点: $ready_url"
    if curl -s -f "$ready_url" >/dev/null; then
        log_success "就绪检查通过"
    else
        log_error "就绪检查失败"
        exit 1
    fi
    
    log_info "检查存活端点: $live_url"
    if curl -s -f "$live_url" >/dev/null; then
        log_success "存活检查通过"
    else
        log_error "存活检查失败"
        exit 1
    fi
    
    # 检查指标端点
    local metrics_url="http://localhost:${METRICS_PORT}/metrics"
    log_info "检查指标端点: $metrics_url"
    if curl -s -f "$metrics_url" >/dev/null; then
        log_success "指标端点正常"
    else
        log_error "指标端点异常"
        exit 1
    fi
    
    log_success "所有健康检查通过"
}

# 查看指标
show_metrics() {
    log_info "查看服务指标..."
    
    if ! is_service_running; then
        log_error "服务未运行"
        exit 1
    fi
    
    local metrics_url="http://localhost:${METRICS_PORT}/metrics"
    
    if command -v curl &> /dev/null; then
        curl -s "$metrics_url"
    else
        log_error "curl 未安装"
        exit 1
    fi
}

# 配置管理
manage_config() {
    log_info "配置管理..."
    
    local config_file=$(get_config_file)
    
    echo "配置文件: $config_file"
    
    if [[ -f "$config_file" ]]; then
        echo "配置内容:"
        cat "$config_file"
    else
        log_error "配置文件不存在"
        exit 1
    fi
}

# 安装服务
install_service() {
    log_info "安装监控服务..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 安装服务"
        return 0
    fi
    
    # 创建服务用户
    if ! id monitoring &>/dev/null; then
        useradd -r -s /bin/false monitoring
        log_info "创建服务用户: monitoring"
    fi
    
    # 创建目录
    local dirs=(
        "/opt/monitoring"
        "/var/log/monitoring"
        "/var/lib/monitoring"
        "/etc/monitoring"
    )
    
    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            chown monitoring:monitoring "$dir"
            log_info "创建目录: $dir"
        fi
    done
    
    # 复制文件
    cp "$PROJECT_ROOT/bin/monitoring" "/opt/monitoring/"
    cp "$PROJECT_ROOT/config/"*.yml "/etc/monitoring/"
    chown -R monitoring:monitoring "/opt/monitoring" "/etc/monitoring"
    
    # 创建 systemd 服务文件
    cat > "/etc/systemd/system/${SERVICE_NAME}.service" << EOF
[Unit]
Description=Taishanglaojun Monitoring Service
After=network.target

[Service]
Type=simple
User=monitoring
Group=monitoring
ExecStart=/opt/monitoring/monitoring
Restart=always
RestartSec=5
Environment=MONITORING_ENV=production
Environment=MONITORING_CONFIG=/etc/monitoring/production.yml

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    
    log_success "服务安装完成"
}

# 卸载服务
uninstall_service() {
    log_info "卸载监控服务..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 卸载服务"
        return 0
    fi
    
    # 停止并禁用服务
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    
    # 删除服务文件
    rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
    systemctl daemon-reload
    
    # 删除文件和目录
    rm -rf "/opt/monitoring"
    rm -rf "/etc/monitoring"
    
    if [[ "$FORCE" == "true" ]]; then
        rm -rf "/var/log/monitoring"
        rm -rf "/var/lib/monitoring"
    fi
    
    # 删除用户
    if id monitoring &>/dev/null; then
        userdel monitoring 2>/dev/null || true
        log_info "删除服务用户: monitoring"
    fi
    
    log_success "服务卸载完成"
}

# 更新服务
update_service() {
    log_info "更新监控服务..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 更新服务"
        return 0
    fi
    
    # 停止服务
    if is_service_running; then
        stop_service
    fi
    
    # 备份当前版本
    if [[ -f "/opt/monitoring/monitoring" ]]; then
        cp "/opt/monitoring/monitoring" "/opt/monitoring/monitoring.backup.$(date +%Y%m%d-%H%M%S)"
        log_info "备份当前版本"
    fi
    
    # 更新二进制文件
    cp "$PROJECT_ROOT/bin/monitoring" "/opt/monitoring/"
    chown monitoring:monitoring "/opt/monitoring/monitoring"
    
    # 更新配置文件
    cp "$PROJECT_ROOT/config/"*.yml "/etc/monitoring/"
    chown -R monitoring:monitoring "/etc/monitoring"
    
    # 重启服务
    start_service
    
    log_success "服务更新完成"
}

# 备份数据
backup_data() {
    log_info "备份监控数据..."
    
    "$SCRIPT_DIR/backup.sh" backup --type all --env "$ENVIRONMENT"
}

# 恢复数据
restore_data() {
    log_info "恢复监控数据..."
    
    "$SCRIPT_DIR/backup.sh" restore --env "$ENVIRONMENT"
}

# 主函数
main() {
    parse_args "$@"
    
    # 静默模式
    if [[ "$QUIET" == "true" ]]; then
        exec >/dev/null 2>&1
    fi
    
    # 检查依赖
    check_dependencies
    
    # 执行命令
    case "$COMMAND" in
        start)
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        health)
            health_check
            ;;
        metrics)
            show_metrics
            ;;
        config)
            manage_config
            ;;
        install)
            install_service
            ;;
        uninstall)
            uninstall_service
            ;;
        update)
            update_service
            ;;
        backup)
            backup_data
            ;;
        restore)
            restore_data
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"