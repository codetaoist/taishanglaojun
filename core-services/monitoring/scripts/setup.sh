#!/bin/bash

# 太上老君监控系统安装脚本
# 支持 Linux 和 macOS

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 版本信息
VERSION="1.0.0"
APP_NAME="taishanglaojun-monitoring"

# 默认配置
INSTALL_DIR="/opt/${APP_NAME}"
CONFIG_DIR="/etc/${APP_NAME}"
LOG_DIR="/var/log/${APP_NAME}"
DATA_DIR="/var/lib/${APP_NAME}"
SERVICE_USER="${APP_NAME}"
DOWNLOAD_URL="https://github.com/taishanglaojun/core-services/releases/download/v${VERSION}"

# 函数定义
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

# 检查操作系统
check_os() {
    log_info "检查操作系统..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            aarch64) ARCH="arm64" ;;
            *) log_error "不支持的架构: $ARCH"; exit 1 ;;
        esac
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            arm64) ARCH="arm64" ;;
            *) log_error "不支持的架构: $ARCH"; exit 1 ;;
        esac
    else
        log_error "不支持的操作系统: $OSTYPE"
        exit 1
    fi
    
    log_success "操作系统: $OS, 架构: $ARCH"
}

# 检查权限
check_permissions() {
    log_info "检查权限..."
    
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要 root 权限运行"
        log_info "请使用: sudo $0"
        exit 1
    fi
    
    log_success "权限检查通过"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    local deps=("curl" "tar" "systemctl")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "缺少依赖: ${missing_deps[*]}"
        log_info "请安装缺少的依赖后重试"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 创建用户
create_user() {
    log_info "创建服务用户..."
    
    if id "$SERVICE_USER" &>/dev/null; then
        log_warning "用户 $SERVICE_USER 已存在"
    else
        useradd --system --shell /bin/false --home-dir "$DATA_DIR" --create-home "$SERVICE_USER"
        log_success "用户 $SERVICE_USER 创建成功"
    fi
}

# 创建目录
create_directories() {
    log_info "创建目录..."
    
    local dirs=("$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR" "$DATA_DIR")
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
        chown "$SERVICE_USER:$SERVICE_USER" "$dir"
        chmod 755 "$dir"
    done
    
    log_success "目录创建完成"
}

# 下载二进制文件
download_binary() {
    log_info "下载二进制文件..."
    
    local binary_name="${APP_NAME}-${OS}-${ARCH}"
    if [[ "$OS" == "windows" ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    local download_file="${binary_name}.tar.gz"
    local download_path="/tmp/${download_file}"
    
    log_info "从 ${DOWNLOAD_URL}/${download_file} 下载..."
    
    if curl -L -o "$download_path" "${DOWNLOAD_URL}/${download_file}"; then
        log_success "下载完成"
    else
        log_error "下载失败"
        exit 1
    fi
    
    # 解压
    log_info "解压文件..."
    tar -xzf "$download_path" -C /tmp/
    
    # 移动二进制文件
    mv "/tmp/${binary_name}" "${INSTALL_DIR}/${APP_NAME}"
    chmod +x "${INSTALL_DIR}/${APP_NAME}"
    chown "$SERVICE_USER:$SERVICE_USER" "${INSTALL_DIR}/${APP_NAME}"
    
    # 清理临时文件
    rm -f "$download_path"
    rm -rf "/tmp/${binary_name%.*}"
    
    log_success "二进制文件安装完成"
}

# 创建配置文件
create_config() {
    log_info "创建配置文件..."
    
    cat > "${CONFIG_DIR}/monitoring.yaml" << 'EOF'
# 太上老君监控系统配置文件

service:
  name: "taishanglaojun-monitoring"
  version: "1.0.0"
  environment: "production"
  host: "0.0.0.0"
  port: 8080
  log_level: "info"
  
tracing:
  enabled: true
  sampling_rate: 0.1
  batch_timeout: "5s"
  max_export_batch_size: 512
  max_queue_size: 2048
  exporters:
    console:
      enabled: false
    jaeger:
      enabled: true
      endpoint: "http://localhost:14268/api/traces"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/taishanglaojun-monitoring/app.log"
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true

storage:
  prometheus:
    url: "http://localhost:9090"
    timeout: "30s"
  influxdb:
    url: "http://localhost:8086"
    token: "monitoring-token-123456789"
    org: "taishanglaojun"
    bucket: "monitoring"

alerting:
  enabled: true
  evaluation_interval: "30s"
  notification_channels:
    email:
      enabled: true
      smtp_host: "localhost"
      smtp_port: 587
      from: "monitoring@taishanglaojun.com"
      to: ["admin@taishanglaojun.com"]

dashboard:
  enabled: true
  refresh_interval: "30s"

performance:
  enabled: true
  collection_interval: "15s"
  
automation:
  enabled: true
  max_concurrent_workflows: 10
EOF
    
    chown "$SERVICE_USER:$SERVICE_USER" "${CONFIG_DIR}/monitoring.yaml"
    chmod 644 "${CONFIG_DIR}/monitoring.yaml"
    
    log_success "配置文件创建完成"
}

# 创建 systemd 服务
create_service() {
    log_info "创建 systemd 服务..."
    
    cat > "/etc/systemd/system/${APP_NAME}.service" << EOF
[Unit]
Description=太上老君监控系统
Documentation=https://github.com/taishanglaojun/core-services/tree/main/monitoring
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=${INSTALL_DIR}/${APP_NAME} -config ${CONFIG_DIR}/monitoring.yaml
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=always
RestartSec=5
StartLimitInterval=0

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${LOG_DIR} ${DATA_DIR}
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

# 环境变量
Environment=MONITORING_CONFIG_FILE=${CONFIG_DIR}/monitoring.yaml
Environment=MONITORING_LOG_DIR=${LOG_DIR}
Environment=MONITORING_DATA_DIR=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable "${APP_NAME}.service"
    
    log_success "systemd 服务创建完成"
}

# 创建日志轮转配置
create_logrotate() {
    log_info "创建日志轮转配置..."
    
    cat > "/etc/logrotate.d/${APP_NAME}" << EOF
${LOG_DIR}/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $SERVICE_USER $SERVICE_USER
    postrotate
        systemctl reload ${APP_NAME} > /dev/null 2>&1 || true
    endscript
}
EOF
    
    log_success "日志轮转配置创建完成"
}

# 启动服务
start_service() {
    log_info "启动服务..."
    
    systemctl start "${APP_NAME}.service"
    
    # 等待服务启动
    sleep 5
    
    if systemctl is-active --quiet "${APP_NAME}.service"; then
        log_success "服务启动成功"
        
        # 显示服务状态
        systemctl status "${APP_NAME}.service" --no-pager -l
        
        # 检查健康状态
        log_info "检查服务健康状态..."
        if curl -f http://localhost:8080/health &>/dev/null; then
            log_success "服务健康检查通过"
        else
            log_warning "服务健康检查失败，请检查配置"
        fi
    else
        log_error "服务启动失败"
        systemctl status "${APP_NAME}.service" --no-pager -l
        exit 1
    fi
}

# 显示安装信息
show_info() {
    log_success "安装完成！"
    echo
    echo "服务信息:"
    echo "  名称: ${APP_NAME}"
    echo "  版本: ${VERSION}"
    echo "  安装目录: ${INSTALL_DIR}"
    echo "  配置目录: ${CONFIG_DIR}"
    echo "  日志目录: ${LOG_DIR}"
    echo "  数据目录: ${DATA_DIR}"
    echo
    echo "常用命令:"
    echo "  启动服务: systemctl start ${APP_NAME}"
    echo "  停止服务: systemctl stop ${APP_NAME}"
    echo "  重启服务: systemctl restart ${APP_NAME}"
    echo "  查看状态: systemctl status ${APP_NAME}"
    echo "  查看日志: journalctl -u ${APP_NAME} -f"
    echo
    echo "Web 界面:"
    echo "  监控面板: http://localhost:8080"
    echo "  健康检查: http://localhost:8080/health"
    echo "  指标接口: http://localhost:8080/metrics"
    echo
    echo "配置文件: ${CONFIG_DIR}/monitoring.yaml"
    echo
    log_info "请根据需要修改配置文件，然后重启服务"
}

# 卸载函数
uninstall() {
    log_info "开始卸载..."
    
    # 停止服务
    if systemctl is-active --quiet "${APP_NAME}.service"; then
        systemctl stop "${APP_NAME}.service"
    fi
    
    # 禁用服务
    if systemctl is-enabled --quiet "${APP_NAME}.service"; then
        systemctl disable "${APP_NAME}.service"
    fi
    
    # 删除服务文件
    rm -f "/etc/systemd/system/${APP_NAME}.service"
    systemctl daemon-reload
    
    # 删除日志轮转配置
    rm -f "/etc/logrotate.d/${APP_NAME}"
    
    # 删除文件和目录
    rm -rf "$INSTALL_DIR"
    rm -rf "$CONFIG_DIR"
    rm -rf "$LOG_DIR"
    rm -rf "$DATA_DIR"
    
    # 删除用户
    if id "$SERVICE_USER" &>/dev/null; then
        userdel "$SERVICE_USER"
    fi
    
    log_success "卸载完成"
}

# 主函数
main() {
    echo "太上老君监控系统安装脚本 v${VERSION}"
    echo "========================================"
    
    case "${1:-install}" in
        install)
            check_os
            check_permissions
            check_dependencies
            create_user
            create_directories
            download_binary
            create_config
            create_service
            create_logrotate
            start_service
            show_info
            ;;
        uninstall)
            check_permissions
            uninstall
            ;;
        *)
            echo "用法: $0 [install|uninstall]"
            echo "  install   - 安装监控系统（默认）"
            echo "  uninstall - 卸载监控系统"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"