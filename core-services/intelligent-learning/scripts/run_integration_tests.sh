#!/bin/bash

# 智能学习服务集成测试运行脚本
# 作者: 太上老君开发团队
# 版本: 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_CONFIG="${PROJECT_ROOT}/test/test_config.yaml"
TEST_REPORTS_DIR="${PROJECT_ROOT}/test_reports"
DOCKER_COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.test.yml"
LOG_FILE="${PROJECT_ROOT}/logs/integration_test.log"

# 创建必要的目录
mkdir -p "${TEST_REPORTS_DIR}"
mkdir -p "$(dirname "${LOG_FILE}")"

# 日志函数
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "${LOG_FILE}"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✓${NC} $1" | tee -a "${LOG_FILE}"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ✗${NC} $1" | tee -a "${LOG_FILE}"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠${NC} $1" | tee -a "${LOG_FILE}"
}

# 清理函数
cleanup() {
    log "开始清理测试环境..."
    
    # 停止测试服务
    if [ -f "${DOCKER_COMPOSE_FILE}" ]; then
        docker-compose -f "${DOCKER_COMPOSE_FILE}" down -v 2>/dev/null || true
    fi
    
    # 清理测试数据
    if [ "$PRESERVE_TEST_DATA" != "true" ]; then
        log "清理测试数据..."
        # 这里可以添加清理测试数据库的命令
    fi
    
    log_success "测试环境清理完成"
}

# 错误处理
handle_error() {
    log_error "测试执行失败，错误码: $1"
    cleanup
    exit $1
}

# 设置错误处理
trap 'handle_error $?' ERR

# 检查依赖
check_dependencies() {
    log "检查测试依赖..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        log_error "Go环境未安装"
        exit 1
    fi
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 启动测试环境
start_test_environment() {
    log "启动测试环境..."
    
    # 创建测试用的docker-compose文件
    cat > "${DOCKER_COMPOSE_FILE}" << EOF
version: '3.8'

services:
  postgres-test:
    image: postgres:13
    environment:
      POSTGRES_DB: intelligent_learning_test
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_password
    ports:
      - "5433:5432"
    volumes:
      - postgres_test_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test_user -d intelligent_learning_test"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis-test:
    image: redis:6-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  elasticsearch-test:
    image: elasticsearch:7.17.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9201:9200"
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 5

  neo4j-test:
    image: neo4j:4.4
    environment:
      NEO4J_AUTH: neo4j/test_password
      NEO4J_dbms_default__database: test
    ports:
      - "7688:7687"
      - "7475:7474"
    healthcheck:
      test: ["CMD", "cypher-shell", "-u", "neo4j", "-p", "test_password", "RETURN 1"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  postgres_test_data:

networks:
  default:
    name: intelligent_learning_test_network
EOF

    # 启动测试服务
    docker-compose -f "${DOCKER_COMPOSE_FILE}" up -d
    
    # 等待服务启动
    log "等待测试服务启动..."
    sleep 30
    
    # 检查服务健康状态
    for service in postgres-test redis-test elasticsearch-test neo4j-test; do
        log "检查 ${service} 健康状态..."
        timeout 60 docker-compose -f "${DOCKER_COMPOSE_FILE}" exec -T "${service}" echo "Service is ready" || {
            log_error "${service} 启动失败"
            exit 1
        }
    done
    
    log_success "测试环境启动完成"
}

# 初始化测试数据
init_test_data() {
    log "初始化测试数据..."
    
    # 运行数据库迁移
    cd "${PROJECT_ROOT}"
    
    # 设置测试环境变量
    export TEST_MODE=true
    export DB_HOST=localhost
    export DB_PORT=5433
    export DB_NAME=intelligent_learning_test
    export DB_USER=test_user
    export DB_PASSWORD=test_password
    export REDIS_HOST=localhost
    export REDIS_PORT=6380
    export ELASTICSEARCH_HOST=localhost:9201
    export NEO4J_URI=bolt://localhost:7688
    export NEO4J_USERNAME=neo4j
    export NEO4J_PASSWORD=test_password
    
    # 运行数据库初始化脚本
    if [ -f "${PROJECT_ROOT}/scripts/init-test-db.sql" ]; then
        PGPASSWORD=test_password psql -h localhost -p 5433 -U test_user -d intelligent_learning_test -f "${PROJECT_ROOT}/scripts/init-test-db.sql"
    fi
    
    log_success "测试数据初始化完成"
}

# 运行单元测试
run_unit_tests() {
    log "运行单元测试..."
    
    cd "${PROJECT_ROOT}"
    
    # 运行单元测试并生成覆盖率报告
    go test -v -race -coverprofile="${TEST_REPORTS_DIR}/coverage.out" -covermode=atomic ./... | tee "${TEST_REPORTS_DIR}/unit_test.log"
    
    # 生成HTML覆盖率报告
    go tool cover -html="${TEST_REPORTS_DIR}/coverage.out" -o "${TEST_REPORTS_DIR}/coverage.html"
    
    # 检查覆盖率
    COVERAGE=$(go tool cover -func="${TEST_REPORTS_DIR}/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')
    COVERAGE_THRESHOLD=80
    
    if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
        log_warning "代码覆盖率 ${COVERAGE}% 低于阈值 ${COVERAGE_THRESHOLD}%"
    else
        log_success "代码覆盖率 ${COVERAGE}% 达到要求"
    fi
}

# 运行集成测试
run_integration_tests() {
    log "运行集成测试..."
    
    cd "${PROJECT_ROOT}"
    
    # 运行集成测试
    go test -v -tags=integration ./test/... | tee "${TEST_REPORTS_DIR}/integration_test.log"
    
    log_success "集成测试完成"
}

# 运行性能测试
run_performance_tests() {
    log "运行性能测试..."
    
    cd "${PROJECT_ROOT}"
    
    # 运行基准测试
    go test -bench=. -benchmem -cpuprofile="${TEST_REPORTS_DIR}/cpu.prof" -memprofile="${TEST_REPORTS_DIR}/mem.prof" ./test/... | tee "${TEST_REPORTS_DIR}/benchmark.log"
    
    log_success "性能测试完成"
}

# 运行API测试
run_api_tests() {
    log "运行API测试..."
    
    # 启动测试服务器
    cd "${PROJECT_ROOT}"
    go run cmd/main.go &
    SERVER_PID=$!
    
    # 等待服务器启动
    sleep 10
    
    # 检查服务器是否启动成功
    if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_error "测试服务器启动失败"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
    
    # 运行API测试
    log "执行API端点测试..."
    
    # 测试健康检查
    curl -f http://localhost:8080/health || log_error "健康检查失败"
    
    # 测试状态端点
    curl -f http://localhost:8080/status || log_error "状态检查失败"
    
    # 测试指标端点
    curl -f http://localhost:8080/metrics || log_error "指标检查失败"
    
    # 停止测试服务器
    kill $SERVER_PID 2>/dev/null || true
    
    log_success "API测试完成"
}

# 生成测试报告
generate_test_report() {
    log "生成测试报告..."
    
    # 创建HTML测试报告
    cat > "${TEST_REPORTS_DIR}/test_report.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>智能学习服务集成测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { color: green; }
        .error { color: red; }
        .warning { color: orange; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="header">
        <h1>智能学习服务集成测试报告</h1>
        <p class="timestamp">生成时间: $(date)</p>
    </div>
    
    <div class="section">
        <h2>测试概览</h2>
        <p>项目: 智能学习服务</p>
        <p>版本: 1.0.0</p>
        <p>测试环境: 集成测试</p>
    </div>
    
    <div class="section">
        <h2>测试结果</h2>
        <p>详细的测试结果请查看相应的日志文件:</p>
        <ul>
            <li><a href="unit_test.log">单元测试日志</a></li>
            <li><a href="integration_test.log">集成测试日志</a></li>
            <li><a href="benchmark.log">性能测试日志</a></li>
            <li><a href="coverage.html">代码覆盖率报告</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>测试文件</h2>
        <ul>
            <li>覆盖率文件: coverage.out</li>
            <li>CPU性能文件: cpu.prof</li>
            <li>内存性能文件: mem.prof</li>
        </ul>
    </div>
</body>
</html>
EOF

    log_success "测试报告生成完成: ${TEST_REPORTS_DIR}/test_report.html"
}

# 主函数
main() {
    log "开始智能学习服务集成测试..."
    
    # 解析命令行参数
    SKIP_UNIT_TESTS=false
    SKIP_INTEGRATION_TESTS=false
    SKIP_PERFORMANCE_TESTS=false
    SKIP_API_TESTS=false
    PRESERVE_TEST_DATA=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-unit)
                SKIP_UNIT_TESTS=true
                shift
                ;;
            --skip-integration)
                SKIP_INTEGRATION_TESTS=true
                shift
                ;;
            --skip-performance)
                SKIP_PERFORMANCE_TESTS=true
                shift
                ;;
            --skip-api)
                SKIP_API_TESTS=true
                shift
                ;;
            --preserve-data)
                PRESERVE_TEST_DATA=true
                shift
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --skip-unit         跳过单元测试"
                echo "  --skip-integration  跳过集成测试"
                echo "  --skip-performance  跳过性能测试"
                echo "  --skip-api          跳过API测试"
                echo "  --preserve-data     保留测试数据"
                echo "  --help              显示帮助信息"
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    # 执行测试流程
    check_dependencies
    start_test_environment
    init_test_data
    
    if [ "$SKIP_UNIT_TESTS" != "true" ]; then
        run_unit_tests
    fi
    
    if [ "$SKIP_INTEGRATION_TESTS" != "true" ]; then
        run_integration_tests
    fi
    
    if [ "$SKIP_PERFORMANCE_TESTS" != "true" ]; then
        run_performance_tests
    fi
    
    if [ "$SKIP_API_TESTS" != "true" ]; then
        run_api_tests
    fi
    
    generate_test_report
    cleanup
    
    log_success "所有测试完成！测试报告位于: ${TEST_REPORTS_DIR}"
}

# 运行主函数
main "$@"