#!/bin/bash

# 码道 CLI 测试脚本

set -e

echo "🧪 开始运行码道 CLI 测试..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 运行测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "${BLUE}运行测试: $test_name${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command"; then
        echo -e "${GREEN}✅ $test_name 通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $test_name 失败${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo
}

# 检查Go环境
echo -e "${YELLOW}🔍 检查环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go 未安装或不在 PATH 中${NC}"
    exit 1
fi

echo "Go 版本: $(go version)"
echo

# 运行单元测试
echo -e "${YELLOW}🧪 运行单元测试...${NC}"
run_test "单元测试" "go test -v ./..."

# 运行竞态条件检测
echo -e "${YELLOW}🏃 运行竞态条件检测...${NC}"
run_test "竞态条件检测" "go test -race ./..."

# 代码覆盖率测试
echo -e "${YELLOW}📊 生成代码覆盖率报告...${NC}"
run_test "代码覆盖率" "go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"

# 代码格式检查
echo -e "${YELLOW}📝 检查代码格式...${NC}"
run_test "代码格式检查" "test -z \"$(gofmt -l .)\""

# 代码静态分析
if command -v golint &> /dev/null; then
    echo -e "${YELLOW}🔍 运行 golint 静态分析...${NC}"
    run_test "golint 检查" "golint ./..."
fi

if command -v go &> /dev/null; then
    echo -e "${YELLOW}🔍 运行 go vet 静态分析...${NC}"
    run_test "go vet 检查" "go vet ./..."
fi

# 依赖检查
echo -e "${YELLOW}📦 检查依赖...${NC}"
run_test "依赖整理" "go mod tidy && git diff --exit-code go.mod go.sum"

# 构建测试
echo -e "${YELLOW}🔨 测试构建...${NC}"
run_test "构建测试" "go build -o /tmp/lo-test ./cmd/lo && rm -f /tmp/lo-test"

# 交叉编译测试
echo -e "${YELLOW}🌍 测试交叉编译...${NC}"
run_test "Linux 交叉编译" "GOOS=linux GOARCH=amd64 go build -o /tmp/lo-linux ./cmd/lo && rm -f /tmp/lo-linux"
run_test "Windows 交叉编译" "GOOS=windows GOARCH=amd64 go build -o /tmp/lo-windows.exe ./cmd/lo && rm -f /tmp/lo-windows.exe"
run_test "macOS 交叉编译" "GOOS=darwin GOARCH=amd64 go build -o /tmp/lo-darwin ./cmd/lo && rm -f /tmp/lo-darwin"

# 安全检查
if command -v gosec &> /dev/null; then
    echo -e "${YELLOW}🔒 运行安全检查...${NC}"
    run_test "安全检查" "gosec ./..."
fi

# 性能基准测试
echo -e "${YELLOW}⚡ 运行性能基准测试...${NC}"
run_test "性能基准测试" "go test -bench=. -benchmem ./..."

# 清理临时文件
echo -e "${YELLOW}🧹 清理临时文件...${NC}"
rm -f coverage.out

# 测试结果汇总
echo -e "${BLUE}📋 测试结果汇总${NC}"
echo "=========================================="
echo -e "总测试数: ${TOTAL_TESTS}"
echo -e "${GREEN}通过: ${PASSED_TESTS}${NC}"
echo -e "${RED}失败: ${FAILED_TESTS}${NC}"
echo "=========================================="

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $FAILED_TESTS 个测试失败${NC}"
    exit 1
fi