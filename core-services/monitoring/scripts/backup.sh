#!/bin/bash

# 太上老君监控系统备份脚本
# 用于备份配置、数据和日志

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICE_NAME="monitoring"

# 默认配置
DEFAULT_BACKUP_DIR="/var/backups/monitoring"
DEFAULT_RETENTION_DAYS=30
DEFAULT_COMPRESSION="gzip"

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
太上老君监控系统备份脚本

用法: $0 [选项] [命令]

命令:
    backup      执行备份
    restore     恢复备份
    list        列出备份
    cleanup     清理旧备份
    verify      验证备份
    help        显示此帮助信息

选项:
    -t, --type TYPE         备份类型 (config|data|logs|all) [默认: all]
    -d, --dir DIR           备份目录 [默认: $DEFAULT_BACKUP_DIR]
    -e, --env ENV           环境 (development|staging|production)
    -n, --name NAME         备份名称 [默认: 自动生成]
    -c, --compression TYPE  压缩类型 (gzip|bzip2|xz|none) [默认: $DEFAULT_COMPRESSION]
    -r, --retention DAYS    保留天数 [默认: $DEFAULT_RETENTION_DAYS]
    -f, --force             强制执行
    -v, --verbose           详细输出
    -q, --quiet             静默模式
    --dry-run               试运行模式
    --exclude PATTERN       排除模式
    --include PATTERN       包含模式
    --encrypt               加密备份
    --password PASSWORD     加密密码

环境变量:
    BACKUP_DIR              备份目录
    BACKUP_RETENTION_DAYS   保留天数
    BACKUP_COMPRESSION      压缩类型
    BACKUP_ENCRYPT_KEY      加密密钥
    DATABASE_URL            数据库连接地址
    REDIS_URL               Redis 连接地址

示例:
    $0 backup
    $0 backup --type config --env production
    $0 restore --name monitoring-20231201-120000
    $0 list --env production
    $0 cleanup --retention 7

EOF
}

# 解析命令行参数
parse_args() {
    COMMAND=""
    BACKUP_TYPE="all"
    BACKUP_DIR="${BACKUP_DIR:-$DEFAULT_BACKUP_DIR}"
    ENVIRONMENT=""
    BACKUP_NAME=""
    COMPRESSION="${BACKUP_COMPRESSION:-$DEFAULT_COMPRESSION}"
    RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-$DEFAULT_RETENTION_DAYS}"
    FORCE=false
    VERBOSE=false
    QUIET=false
    DRY_RUN=false
    EXCLUDE_PATTERNS=()
    INCLUDE_PATTERNS=()
    ENCRYPT=false
    PASSWORD=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            backup|restore|list|cleanup|verify|help)
                COMMAND="$1"
                shift
                ;;
            -t|--type)
                BACKUP_TYPE="$2"
                shift 2
                ;;
            -d|--dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--name)
                BACKUP_NAME="$2"
                shift 2
                ;;
            -c|--compression)
                COMPRESSION="$2"
                shift 2
                ;;
            -r|--retention)
                RETENTION_DAYS="$2"
                shift 2
                ;;
            -f|--force)
                FORCE=true
                shift
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
            --exclude)
                EXCLUDE_PATTERNS+=("$2")
                shift 2
                ;;
            --include)
                INCLUDE_PATTERNS+=("$2")
                shift 2
                ;;
            --encrypt)
                ENCRYPT=true
                shift
                ;;
            --password)
                PASSWORD="$2"
                shift 2
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
        COMMAND="backup"
    fi

    # 验证参数
    case "$BACKUP_TYPE" in
        config|data|logs|all) ;;
        *)
            log_error "无效的备份类型: $BACKUP_TYPE"
            exit 1
            ;;
    esac

    case "$COMPRESSION" in
        gzip|bzip2|xz|none) ;;
        *)
            log_error "无效的压缩类型: $COMPRESSION"
            exit 1
            ;;
    esac

    if [[ ! "$RETENTION_DAYS" =~ ^[0-9]+$ ]]; then
        log_error "保留天数必须是数字: $RETENTION_DAYS"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    local deps=("tar" "find" "date")
    
    case "$COMPRESSION" in
        gzip) deps+=("gzip") ;;
        bzip2) deps+=("bzip2") ;;
        xz) deps+=("xz") ;;
    esac

    if [[ "$ENCRYPT" == "true" ]]; then
        deps+=("gpg")
    fi

    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "依赖未安装: $dep"
            exit 1
        fi
    done
}

# 创建备份目录
create_backup_dir() {
    local dir="$1"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 创建备份目录: $dir"
        return
    fi
    
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
        log_info "创建备份目录: $dir"
    fi
}

# 生成备份名称
generate_backup_name() {
    local type="$1"
    local env="$2"
    local timestamp=$(date +"%Y%m%d-%H%M%S")
    
    if [[ -n "$env" ]]; then
        echo "${SERVICE_NAME}-${env}-${type}-${timestamp}"
    else
        echo "${SERVICE_NAME}-${type}-${timestamp}"
    fi
}

# 获取压缩扩展名
get_compression_ext() {
    case "$1" in
        gzip) echo ".tar.gz" ;;
        bzip2) echo ".tar.bz2" ;;
        xz) echo ".tar.xz" ;;
        none) echo ".tar" ;;
    esac
}

# 获取压缩选项
get_compression_opts() {
    case "$1" in
        gzip) echo "z" ;;
        bzip2) echo "j" ;;
        xz) echo "J" ;;
        none) echo "" ;;
    esac
}

# 备份配置文件
backup_config() {
    local backup_dir="$1"
    local backup_name="$2"
    
    log_info "备份配置文件..."
    
    local config_dirs=(
        "$PROJECT_ROOT/config"
        "$PROJECT_ROOT/k8s"
        "$PROJECT_ROOT/docker"
        "$PROJECT_ROOT/scripts"
    )
    
    local config_files=(
        "$PROJECT_ROOT/Dockerfile"
        "$PROJECT_ROOT/docker-compose.yml"
        "$PROJECT_ROOT/package.json"
        "$PROJECT_ROOT/go.mod"
        "$PROJECT_ROOT/Makefile"
        "$PROJECT_ROOT/.env.example"
    )
    
    local temp_dir=$(mktemp -d)
    local config_backup_dir="$temp_dir/config"
    mkdir -p "$config_backup_dir"
    
    # 复制配置目录
    for dir in "${config_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            cp -r "$dir" "$config_backup_dir/"
            log_info "复制配置目录: $dir"
        fi
    done
    
    # 复制配置文件
    for file in "${config_files[@]}"; do
        if [[ -f "$file" ]]; then
            cp "$file" "$config_backup_dir/"
            log_info "复制配置文件: $file"
        fi
    done
    
    # 创建压缩包
    local ext=$(get_compression_ext "$COMPRESSION")
    local opts=$(get_compression_opts "$COMPRESSION")
    local archive_path="$backup_dir/${backup_name}-config${ext}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 创建配置备份: $archive_path"
    else
        tar -c${opts}f "$archive_path" -C "$temp_dir" config
        log_success "配置备份完成: $archive_path"
    fi
    
    # 清理临时目录
    rm -rf "$temp_dir"
}

# 备份数据
backup_data() {
    local backup_dir="$1"
    local backup_name="$2"
    
    log_info "备份数据..."
    
    local data_dirs=(
        "/var/lib/monitoring"
        "/opt/monitoring/data"
        "$PROJECT_ROOT/data"
    )
    
    local temp_dir=$(mktemp -d)
    local data_backup_dir="$temp_dir/data"
    mkdir -p "$data_backup_dir"
    
    # 备份文件数据
    for dir in "${data_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            cp -r "$dir" "$data_backup_dir/"
            log_info "复制数据目录: $dir"
        fi
    done
    
    # 备份数据库
    if [[ -n "${DATABASE_URL:-}" ]]; then
        backup_database "$data_backup_dir"
    fi
    
    # 备份 Redis
    if [[ -n "${REDIS_URL:-}" ]]; then
        backup_redis "$data_backup_dir"
    fi
    
    # 创建压缩包
    local ext=$(get_compression_ext "$COMPRESSION")
    local opts=$(get_compression_opts "$COMPRESSION")
    local archive_path="$backup_dir/${backup_name}-data${ext}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 创建数据备份: $archive_path"
    else
        tar -c${opts}f "$archive_path" -C "$temp_dir" data
        log_success "数据备份完成: $archive_path"
    fi
    
    # 清理临时目录
    rm -rf "$temp_dir"
}

# 备份数据库
backup_database() {
    local backup_dir="$1"
    
    log_info "备份数据库..."
    
    if [[ "$DATABASE_URL" =~ postgres ]]; then
        # PostgreSQL
        if command -v pg_dump &> /dev/null; then
            local db_backup="$backup_dir/database.sql"
            if [[ "$DRY_RUN" == "true" ]]; then
                log_info "[DRY RUN] 备份 PostgreSQL 数据库"
            else
                pg_dump "$DATABASE_URL" > "$db_backup"
                log_success "PostgreSQL 数据库备份完成"
            fi
        else
            log_warning "pg_dump 未安装，跳过数据库备份"
        fi
    elif [[ "$DATABASE_URL" =~ mysql ]]; then
        # MySQL
        if command -v mysqldump &> /dev/null; then
            local db_backup="$backup_dir/database.sql"
            if [[ "$DRY_RUN" == "true" ]]; then
                log_info "[DRY RUN] 备份 MySQL 数据库"
            else
                mysqldump --single-transaction --routines --triggers "$DATABASE_URL" > "$db_backup"
                log_success "MySQL 数据库备份完成"
            fi
        else
            log_warning "mysqldump 未安装，跳过数据库备份"
        fi
    else
        log_warning "不支持的数据库类型，跳过数据库备份"
    fi
}

# 备份 Redis
backup_redis() {
    local backup_dir="$1"
    
    log_info "备份 Redis..."
    
    if command -v redis-cli &> /dev/null; then
        local redis_backup="$backup_dir/redis.rdb"
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "[DRY RUN] 备份 Redis 数据"
        else
            redis-cli -u "$REDIS_URL" --rdb "$redis_backup"
            log_success "Redis 数据备份完成"
        fi
    else
        log_warning "redis-cli 未安装，跳过 Redis 备份"
    fi
}

# 备份日志
backup_logs() {
    local backup_dir="$1"
    local backup_name="$2"
    
    log_info "备份日志..."
    
    local log_dirs=(
        "/var/log/monitoring"
        "/opt/monitoring/logs"
        "$PROJECT_ROOT/logs"
    )
    
    local temp_dir=$(mktemp -d)
    local logs_backup_dir="$temp_dir/logs"
    mkdir -p "$logs_backup_dir"
    
    # 复制日志目录
    for dir in "${log_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            cp -r "$dir" "$logs_backup_dir/"
            log_info "复制日志目录: $dir"
        fi
    done
    
    # 创建压缩包
    local ext=$(get_compression_ext "$COMPRESSION")
    local opts=$(get_compression_opts "$COMPRESSION")
    local archive_path="$backup_dir/${backup_name}-logs${ext}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] 创建日志备份: $archive_path"
    else
        tar -c${opts}f "$archive_path" -C "$temp_dir" logs
        log_success "日志备份完成: $archive_path"
    fi
    
    # 清理临时目录
    rm -rf "$temp_dir"
}

# 执行备份
do_backup() {
    log_info "开始备份..."
    
    # 创建备份目录
    create_backup_dir "$BACKUP_DIR"
    
    # 生成备份名称
    if [[ -z "$BACKUP_NAME" ]]; then
        BACKUP_NAME=$(generate_backup_name "$BACKUP_TYPE" "$ENVIRONMENT")
    fi
    
    log_info "备份名称: $BACKUP_NAME"
    
    # 执行备份
    case "$BACKUP_TYPE" in
        config)
            backup_config "$BACKUP_DIR" "$BACKUP_NAME"
            ;;
        data)
            backup_data "$BACKUP_DIR" "$BACKUP_NAME"
            ;;
        logs)
            backup_logs "$BACKUP_DIR" "$BACKUP_NAME"
            ;;
        all)
            backup_config "$BACKUP_DIR" "$BACKUP_NAME"
            backup_data "$BACKUP_DIR" "$BACKUP_NAME"
            backup_logs "$BACKUP_DIR" "$BACKUP_NAME"
            ;;
    esac
    
    # 加密备份
    if [[ "$ENCRYPT" == "true" ]]; then
        encrypt_backups "$BACKUP_DIR" "$BACKUP_NAME"
    fi
    
    log_success "备份完成"
}

# 加密备份
encrypt_backups() {
    local backup_dir="$1"
    local backup_name="$2"
    
    log_info "加密备份文件..."
    
    local ext=$(get_compression_ext "$COMPRESSION")
    local files=("$backup_dir"/${backup_name}*${ext})
    
    for file in "${files[@]}"; do
        if [[ -f "$file" ]]; then
            if [[ "$DRY_RUN" == "true" ]]; then
                log_info "[DRY RUN] 加密文件: $file"
            else
                if [[ -n "$PASSWORD" ]]; then
                    gpg --batch --yes --passphrase "$PASSWORD" --symmetric --cipher-algo AES256 "$file"
                    rm "$file"
                    log_success "文件已加密: ${file}.gpg"
                else
                    gpg --symmetric --cipher-algo AES256 "$file"
                    rm "$file"
                    log_success "文件已加密: ${file}.gpg"
                fi
            fi
        fi
    done
}

# 列出备份
list_backups() {
    log_info "列出备份文件..."
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_warning "备份目录不存在: $BACKUP_DIR"
        return
    fi
    
    local pattern="${SERVICE_NAME}-"
    if [[ -n "$ENVIRONMENT" ]]; then
        pattern="${SERVICE_NAME}-${ENVIRONMENT}-"
    fi
    
    echo "备份目录: $BACKUP_DIR"
    echo "========================================="
    
    find "$BACKUP_DIR" -name "${pattern}*" -type f | sort | while read -r file; do
        local size=$(du -h "$file" | cut -f1)
        local date=$(stat -c %y "$file" 2>/dev/null || stat -f %Sm "$file" 2>/dev/null || echo "Unknown")
        echo "$(basename "$file") - $size - $date"
    done
}

# 清理旧备份
cleanup_backups() {
    log_info "清理旧备份..."
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_warning "备份目录不存在: $BACKUP_DIR"
        return
    fi
    
    local pattern="${SERVICE_NAME}-"
    if [[ -n "$ENVIRONMENT" ]]; then
        pattern="${SERVICE_NAME}-${ENVIRONMENT}-"
    fi
    
    local count=0
    find "$BACKUP_DIR" -name "${pattern}*" -type f -mtime +$RETENTION_DAYS | while read -r file; do
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "[DRY RUN] 删除旧备份: $file"
        else
            rm "$file"
            log_info "删除旧备份: $file"
        fi
        ((count++))
    done
    
    if [[ $count -eq 0 ]]; then
        log_info "没有需要清理的旧备份"
    else
        log_success "清理了 $count 个旧备份文件"
    fi
}

# 验证备份
verify_backup() {
    log_info "验证备份..."
    
    if [[ -z "$BACKUP_NAME" ]]; then
        log_error "请指定要验证的备份名称"
        exit 1
    fi
    
    local ext=$(get_compression_ext "$COMPRESSION")
    local files=("$BACKUP_DIR"/${BACKUP_NAME}*${ext})
    
    local verified=0
    local failed=0
    
    for file in "${files[@]}"; do
        if [[ -f "$file" ]]; then
            log_info "验证文件: $file"
            
            case "$COMPRESSION" in
                gzip)
                    if gzip -t "$file" 2>/dev/null; then
                        log_success "文件完整: $file"
                        ((verified++))
                    else
                        log_error "文件损坏: $file"
                        ((failed++))
                    fi
                    ;;
                bzip2)
                    if bzip2 -t "$file" 2>/dev/null; then
                        log_success "文件完整: $file"
                        ((verified++))
                    else
                        log_error "文件损坏: $file"
                        ((failed++))
                    fi
                    ;;
                xz)
                    if xz -t "$file" 2>/dev/null; then
                        log_success "文件完整: $file"
                        ((verified++))
                    else
                        log_error "文件损坏: $file"
                        ((failed++))
                    fi
                    ;;
                none)
                    if tar -tf "$file" >/dev/null 2>&1; then
                        log_success "文件完整: $file"
                        ((verified++))
                    else
                        log_error "文件损坏: $file"
                        ((failed++))
                    fi
                    ;;
            esac
        fi
    done
    
    log_info "验证完成: $verified 个文件完整, $failed 个文件损坏"
    
    if [[ $failed -gt 0 ]]; then
        exit 1
    fi
}

# 恢复备份
restore_backup() {
    log_info "恢复备份..."
    
    if [[ -z "$BACKUP_NAME" ]]; then
        log_error "请指定要恢复的备份名称"
        exit 1
    fi
    
    if [[ "$FORCE" != "true" ]]; then
        echo -n "确定要恢复备份 '$BACKUP_NAME' 吗? (y/N): "
        read -r confirm
        if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
            log_info "取消恢复操作"
            exit 0
        fi
    fi
    
    local ext=$(get_compression_ext "$COMPRESSION")
    local opts=$(get_compression_opts "$COMPRESSION")
    
    # 恢复配置
    local config_file="$BACKUP_DIR/${BACKUP_NAME}-config${ext}"
    if [[ -f "$config_file" ]]; then
        log_info "恢复配置文件..."
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "[DRY RUN] 恢复配置: $config_file"
        else
            tar -x${opts}f "$config_file" -C "$PROJECT_ROOT"
            log_success "配置恢复完成"
        fi
    fi
    
    # 恢复数据
    local data_file="$BACKUP_DIR/${BACKUP_NAME}-data${ext}"
    if [[ -f "$data_file" ]]; then
        log_info "恢复数据文件..."
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "[DRY RUN] 恢复数据: $data_file"
        else
            tar -x${opts}f "$data_file" -C "/"
            log_success "数据恢复完成"
        fi
    fi
    
    # 恢复日志
    local logs_file="$BACKUP_DIR/${BACKUP_NAME}-logs${ext}"
    if [[ -f "$logs_file" ]]; then
        log_info "恢复日志文件..."
        if [[ "$DRY_RUN" == "true" ]]; then
            log_info "[DRY RUN] 恢复日志: $logs_file"
        else
            tar -x${opts}f "$logs_file" -C "/"
            log_success "日志恢复完成"
        fi
    fi
    
    log_success "备份恢复完成"
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
        backup)
            do_backup
            ;;
        restore)
            restore_backup
            ;;
        list)
            list_backups
            ;;
        cleanup)
            cleanup_backups
            ;;
        verify)
            verify_backup
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