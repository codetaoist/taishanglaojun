#!/bin/bash

# 统一数据库迁移脚本
# 支持多种迁移方式和表前缀管理

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DB_DIR="$PROJECT_ROOT/db"
MIGRATIONS_DIR="$DB_DIR/migrations"

# 数据库连接配置
DB_HOST=${DB_HOST:-127.0.0.1}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASS:-password}
DB_NAME=${DB_NAME:-taishanglaojun}
CONTAINER_NAME=${CONTAINER_NAME:-taishanglaojun-postgres}

# 显示帮助信息
show_help() {
    echo "统一数据库迁移脚本"
    echo ""
    echo "用法: $0 [选项] [参数]"
    echo ""
    echo "选项:"
    echo "  up              应用所有待执行的迁移"
    echo "  status          显示迁移状态"
    echo "  validate        验证迁移文件"
    echo "  fix             修复迁移问题"
    echo "  clean           清除数据库中的所有对象（危险操作）"
    echo "  analyze         分析数据库表结构"
    echo "  help            显示此帮助信息"
    echo ""
    echo "参数:"
    echo "  --force         强制执行，跳过确认"
    echo "  --dry-run       预览模式，不实际执行"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST         数据库主机 (默认: 127.0.0.1)"
    echo "  DB_PORT         数据库端口 (默认: 5432)"
    echo "  DB_USER         数据库用户 (默认: postgres)"
    echo "  DB_PASS         数据库密码 (默认: password)"
    echo "  DB_NAME         数据库名称 (默认: taishanglaojun)"
    echo "  CONTAINER_NAME  PostgreSQL容器名称 (默认: taishanglaojun-postgres)"
}

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查Docker是否运行
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker未运行，请先启动Docker"
        exit 1
    fi
}

# 检查数据库连接
check_connection() {
    log_info "检查数据库连接..."
    if ! docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1;" > /dev/null 2>&1; then
        log_error "无法连接到数据库，请检查连接配置"
        exit 1
    fi
    log_info "数据库连接成功"
}

# 创建迁移记录表
create_migration_table() {
    log_info "创建迁移记录表..."
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            checksum VARCHAR(255),
            installed_by VARCHAR(255),
            installed_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    " > /dev/null 2>&1
}

# 检查迁移是否已执行
is_migration_applied() {
    local version=$1
    local result=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version = '$version';" | tr -d ' ')
    [ "$result" = "1" ]
}

# 标记迁移为已执行
mark_migration_applied() {
    local version=$1
    local checksum=$2
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "INSERT INTO schema_migrations (version, checksum, installed_by) VALUES ('$version', '$checksum', 'migrate.sh');" > /dev/null 2>&1
}

# 计算文件校验和
calculate_checksum() {
    local file=$1
    if command -v md5 > /dev/null 2>&1; then
        md5 -q "$file"
    elif command -v md5sum > /dev/null 2>&1; then
        md5sum "$file" | cut -d' ' -f1
    else
        echo ""
    fi
}

# 执行迁移文件
execute_migration() {
    local file=$1
    local version=$(basename "$file" | cut -d'_' -f1)
    local checksum=$(calculate_checksum "$file")
    
    if is_migration_applied "$version"; then
        local stored_checksum=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT checksum FROM schema_migrations WHERE version = '$version';" | tr -d ' ')
        
        if [ "$checksum" != "$stored_checksum" ]; then
            log_warn "迁移 $(basename "$file") 已应用但校验和不匹配，可能已被修改"
        else
            log_debug "跳过已应用的迁移: $(basename "$file")"
            return 0
        fi
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "[预览] 应用迁移: $(basename "$file")"
        return 0
    fi
    
    log_info "应用迁移: $(basename "$file")"
    if docker exec -i ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} < "$file"; then
        mark_migration_applied "$version" "$checksum"
        log_info "迁移 $(basename "$file") 应用成功"
    else
        log_error "迁移 $(basename "$file") 应用失败"
        exit 1
    fi
}

# 应用迁移
migrate_up() {
    log_info "开始应用数据库迁移..."
    check_docker
    check_connection
    create_migration_table
    
    # 按文件名排序执行迁移
    for file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$file" ]; then
            execute_migration "$file"
        fi
    done
    
    log_info "所有迁移已成功应用！"
}

# 显示迁移状态
migrate_status() {
    log_info "显示迁移状态..."
    check_docker
    check_connection
    create_migration_table
    
    echo "已应用的迁移:"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "SELECT version, installed_on FROM schema_migrations ORDER BY version;" 2>/dev/null || echo "无迁移记录"
    
    echo ""
    echo "可用的迁移文件:"
    for file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$file" ]; then
            version=$(basename "$file" | cut -d'_' -f1)
            if is_migration_applied "$version"; then
                echo "  [已应用] $(basename "$file")"
            else
                echo "  [未应用] $(basename "$file")"
            fi
        fi
    done
}

# 验证迁移文件
validate_migrations() {
    log_info "验证迁移文件..."
    
    for file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$file" ]; then
            local version=$(basename "$file" | cut -d'_' -f1)
            local checksum=$(calculate_checksum "$file")
            
            if is_migration_applied "$version"; then
                local stored_checksum=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT checksum FROM schema_migrations WHERE version = '$version';" | tr -d ' ')
                
                if [ "$checksum" != "$stored_checksum" ]; then
                    log_warn "迁移 $(basename "$file") 已应用但校验和不匹配，可能已被修改"
                else
                    log_info "迁移 $(basename "$file") 验证通过"
                fi
            else
                log_info "迁移 $(basename "$file") 尚未应用"
            fi
        fi
    done
}

# 修复迁移问题
fix_migrations() {
    log_info "开始修复迁移问题..."
    check_docker
    check_connection
    
    # 1. 修复缺少的vector扩展问题
    log_info "检查vector扩展..."
    if ! docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1 FROM pg_extension WHERE extname = 'vector';" | grep -q "1"; then
        log_warn "vector扩展未安装，尝试安装..."
        if docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "CREATE EXTENSION IF NOT EXISTS vector;" 2>/dev/null; then
            log_info "vector扩展安装成功"
        else
            log_warn "vector扩展安装失败，将创建不依赖vector的表结构"
            
            # 创建不依赖vector的tai_vectors表
            docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
                CREATE TABLE IF NOT EXISTS tai_vectors (
                    id BIGSERIAL PRIMARY KEY,
                    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
                    collection_id INT NOT NULL,
                    external_id VARCHAR(256),
                    metadata JSONB,
                    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    CONSTRAINT fk_tai_vectors_collection
                        FOREIGN KEY(collection_id) REFERENCES tai_vector_collections(id) ON DELETE CASCADE
                );
                
                CREATE INDEX IF NOT EXISTS idx_tai_vectors_tenant_collection ON tai_vectors(tenant_id, collection_id);
                CREATE INDEX IF NOT EXISTS idx_tai_vectors_collection_external ON tai_vectors(collection_id, external_id);
            " || log_warn "创建tai_vectors表失败"
        fi
    else
        log_info "vector扩展已安装"
    fi
    
    # 2. 修复缺少的tenant_id列
    log_info "检查并修复缺少的tenant_id列..."
    
    # 检查lao_domains表
    if docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\d lao_domains" | grep -q "lao_domains"; then
        if ! docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\d lao_domains" | grep -q "tenant_id"; then
            log_info "为lao_domains表添加tenant_id列..."
            docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "ALTER TABLE lao_domains ADD COLUMN tenant_id VARCHAR(64) NOT NULL DEFAULT 'default';" || log_warn "添加lao_domains.tenant_id失败"
        fi
    fi
    
    # 检查tai_model_configs表
    if docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\d tai_model_configs" | grep -q "tai_model_configs"; then
        if ! docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\d tai_model_configs" | grep -q "tenant_id"; then
            log_info "为tai_model_configs表添加tenant_id列..."
            docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "ALTER TABLE tai_model_configs ADD COLUMN tenant_id VARCHAR(64) NOT NULL DEFAULT 'default';" || log_warn "添加tai_model_configs.tenant_id失败"
        fi
    fi
    
    # 3. 标记迁移为已完成
    log_info "标记迁移为已完成..."
    create_migration_table
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "INSERT INTO schema_migrations (version) VALUES ('V1__init_laojun.sql') ON CONFLICT (version) DO NOTHING;"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "INSERT INTO schema_migrations (version) VALUES ('V2__init_taishang.sql') ON CONFLICT (version) DO NOTHING;"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "INSERT INTO schema_migrations (version) VALUES ('V3__init_conversation.sql') ON CONFLICT (version) DO NOTHING;"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "INSERT INTO schema_migrations (version) VALUES ('V4__fix_model_configs.sql') ON CONFLICT (version) DO NOTHING;"
    
    log_info "迁移修复完成！"
}

# 清除数据库
migrate_clean() {
    log_warn "警告: 此操作将删除数据库中的所有对象！"
    if [ "$FORCE" != "true" ]; then
        read -p "确定要继续吗？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "操作已取消"
            exit 0
        fi
    fi
    
    check_docker
    check_connection
    
    log_info "清除数据库中的所有对象..."
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" || log_error "清除数据库失败"
    
    log_info "数据库已成功清除！"
}

# 分析数据库表结构
analyze_database() {
    log_info "分析数据库表结构..."
    check_docker
    check_connection
    
    echo "数据库中的所有表:"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "\dt" || log_error "获取表列表失败"
    
    echo ""
    echo "不带前缀的表:"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name NOT LIKE 'lao_%' AND table_name NOT LIKE 'tai_%' ORDER BY table_name;" || log_error "获取不带前缀的表失败"
    
    echo ""
    echo "表前缀统计:"
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "
        SELECT 
            CASE 
                WHEN table_name LIKE 'lao_%' THEN 'lao_'
                WHEN table_name LIKE 'tai_%' THEN 'tai_'
                ELSE '无前缀'
            END AS prefix,
            COUNT(*) AS count
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        GROUP BY 
            CASE 
                WHEN table_name LIKE 'lao_%' THEN 'lao_'
                WHEN table_name LIKE 'tai_%' THEN 'tai_'
                ELSE '无前缀'
            END
        ORDER BY prefix;
    " || log_error "获取表前缀统计失败"
}

# 解析命令行参数
FORCE=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        up|status|validate|fix|clean|analyze|help)
            COMMAND=$1
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 处理命令
case "${COMMAND:-help}" in
    up)
        migrate_up
        ;;
    status)
        migrate_status
        ;;
    validate)
        validate_migrations
        ;;
    fix)
        fix_migrations
        ;;
    clean)
        migrate_clean
        ;;
    analyze)
        analyze_database
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "未知命令: ${COMMAND:-}"
        echo ""
        show_help
        exit 1
        ;;
esac