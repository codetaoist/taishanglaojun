# 太上老君AI平台 Makefile
# 提供常用的开发、测试、构建和部署命令

.PHONY: help dev build test clean docker-build docker-up docker-down lint format setup deps-update security-scan docs

# 默认目标
.DEFAULT_GOAL := help

# 项目信息
PROJECT_NAME := taishanglaojun-ai-platform
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# 颜色定义
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

# 帮助信息
help: ## 显示帮助信息
	@echo "$(CYAN)太上老君AI平台 - 开发工具$(RESET)"
	@echo "$(YELLOW)版本: $(VERSION)$(RESET)"
	@echo ""
	@echo "$(GREEN)可用命令:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==================== 开发命令 ====================

dev: ## 启动开发环境
	@echo "$(GREEN)🚀 启动开发环境...$(RESET)"
	@npm run dev

dev-frontend: ## 启动前端开发服务器
	@echo "$(GREEN)🌐 启动前端开发服务器...$(RESET)"
	@npm run dev:frontend

dev-backend: ## 启动后端开发服务器
	@echo "$(GREEN)⚙️ 启动后端开发服务器...$(RESET)"
	@npm run dev:backend

dev-ai: ## 启动AI服务开发服务器
	@echo "$(GREEN)🤖 启动AI服务开发服务器...$(RESET)"
	@npm run dev:ai

# ==================== 构建命令 ====================

build: ## 构建所有服务
	@echo "$(GREEN)🔨 构建所有服务...$(RESET)"
	@npm run build

build-frontend: ## 构建前端应用
	@echo "$(GREEN)🌐 构建前端应用...$(RESET)"
	@npm run build:frontend

build-backend: ## 构建后端服务
	@echo "$(GREEN)⚙️ 构建后端服务...$(RESET)"
	@npm run build:backend

# ==================== 测试命令 ====================

test: ## 运行所有测试
	@echo "$(GREEN)🧪 运行所有测试...$(RESET)"
	@npm run test

test-frontend: ## 运行前端测试
	@echo "$(GREEN)🌐 运行前端测试...$(RESET)"
	@npm run test:frontend

test-backend: ## 运行后端测试
	@echo "$(GREEN)⚙️ 运行后端测试...$(RESET)"
	@npm run test:backend

test-coverage: ## 生成测试覆盖率报告
	@echo "$(GREEN)📊 生成测试覆盖率报告...$(RESET)"
	@cd frontend/web-app && npm run test:coverage
	@cd core-services && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# ==================== 代码质量命令 ====================

lint: ## 运行代码检查
	@echo "$(GREEN)🔍 运行代码检查...$(RESET)"
	@npm run lint

lint-fix: ## 自动修复代码问题
	@echo "$(GREEN)🔧 自动修复代码问题...$(RESET)"
	@npm run lint:frontend -- --fix
	@cd core-services && golangci-lint run --fix

format: ## 格式化代码
	@echo "$(GREEN)✨ 格式化代码...$(RESET)"
	@npm run format

security-scan: ## 运行安全扫描
	@echo "$(GREEN)🔒 运行安全扫描...$(RESET)"
	@npm run security:scan

# ==================== Docker命令 ====================

docker-build: ## 构建Docker镜像
	@echo "$(GREEN)🐳 构建Docker镜像...$(RESET)"
	@docker-compose build

docker-up: ## 启动Docker服务
	@echo "$(GREEN)🚀 启动Docker服务...$(RESET)"
	@docker-compose up -d

docker-down: ## 停止Docker服务
	@echo "$(YELLOW)⏹️ 停止Docker服务...$(RESET)"
	@docker-compose down

docker-logs: ## 查看Docker日志
	@echo "$(GREEN)📋 查看Docker日志...$(RESET)"
	@docker-compose logs -f

docker-ps: ## 查看Docker容器状态
	@echo "$(GREEN)📊 查看Docker容器状态...$(RESET)"
	@docker-compose ps

docker-clean: ## 清理Docker资源
	@echo "$(YELLOW)🧹 清理Docker资源...$(RESET)"
	@docker-compose down -v
	@docker system prune -f

# ==================== Kubernetes命令 ====================

k8s-deploy: ## 部署到Kubernetes
	@echo "$(GREEN)☸️ 部署到Kubernetes...$(RESET)"
	@kubectl apply -f k8s/

k8s-delete: ## 从Kubernetes删除
	@echo "$(YELLOW)🗑️ 从Kubernetes删除...$(RESET)"
	@kubectl delete -f k8s/

k8s-status: ## 查看Kubernetes状态
	@echo "$(GREEN)📊 查看Kubernetes状态...$(RESET)"
	@kubectl get pods -l app=taishanglaojun

k8s-logs: ## 查看Kubernetes日志
	@echo "$(GREEN)📋 查看Kubernetes日志...$(RESET)"
	@kubectl logs -l app=taishanglaojun -f

# ==================== 环境管理命令 ====================

setup: ## 初始化开发环境
	@echo "$(GREEN)⚙️ 初始化开发环境...$(RESET)"
	@npm run setup
	@cp .env.example .env
	@echo "$(YELLOW)请编辑 .env 文件配置环境变量$(RESET)"

clean: ## 清理构建文件
	@echo "$(YELLOW)🧹 清理构建文件...$(RESET)"
	@npm run clean
	@rm -rf dist build coverage.out coverage.html

deps-install: ## 安装依赖
	@echo "$(GREEN)📦 安装依赖...$(RESET)"
	@npm install
	@cd frontend/web-app && npm install
	@cd core-services && go mod download

deps-update: ## 更新依赖
	@echo "$(GREEN)🔄 更新依赖...$(RESET)"
	@npm run deps:update

# ==================== 数据库命令 ====================

db-migrate: ## 运行数据库迁移
	@echo "$(GREEN)🗄️ 运行数据库迁移...$(RESET)"
	@cd core-services && go run cmd/migrate/main.go up

db-rollback: ## 回滚数据库迁移
	@echo "$(YELLOW)↩️ 回滚数据库迁移...$(RESET)"
	@cd core-services && go run cmd/migrate/main.go down

db-seed: ## 填充测试数据
	@echo "$(GREEN)🌱 填充测试数据...$(RESET)"
	@cd core-services && go run cmd/seed/main.go

db-reset: ## 重置数据库
	@echo "$(RED)🔄 重置数据库...$(RESET)"
	@cd core-services && go run cmd/migrate/main.go reset

# ==================== 文档命令 ====================

docs-serve: ## 启动文档服务器
	@echo "$(GREEN)📚 启动文档服务器...$(RESET)"
	@npm run docs:serve

docs-build: ## 构建文档
	@echo "$(GREEN)📖 构建文档...$(RESET)"
	@npm run docs:build

# ==================== 发布命令 ====================

release: ## 创建发布版本
	@echo "$(GREEN)🚀 创建发布版本 $(VERSION)...$(RESET)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

release-notes: ## 生成发布说明
	@echo "$(GREEN)📝 生成发布说明...$(RESET)"
	@git log --pretty=format:"- %s" $(shell git describe --tags --abbrev=0)..HEAD

# ==================== 监控命令 ====================

monitor: ## 查看系统监控
	@echo "$(GREEN)📊 系统监控信息:$(RESET)"
	@echo "$(CYAN)Prometheus:$(RESET) http://localhost:9090"
	@echo "$(CYAN)Grafana:$(RESET) http://localhost:3001 (admin/admin123)"
	@echo "$(CYAN)应用状态:$(RESET)"
	@curl -s http://localhost:8080/health | jq . || echo "后端服务未运行"

logs: ## 查看应用日志
	@echo "$(GREEN)📋 查看应用日志...$(RESET)"
	@docker-compose logs -f --tail=100

# ==================== 工具命令 ====================

generate-api: ## 生成API文档
	@echo "$(GREEN)📄 生成API文档...$(RESET)"
	@cd core-services && swag init -g cmd/server/main.go

generate-proto: ## 生成protobuf文件
	@echo "$(GREEN)🔧 生成protobuf文件...$(RESET)"
	@cd core-services && protoc --go_out=. --go-grpc_out=. proto/*.proto

benchmark: ## 运行性能测试
	@echo "$(GREEN)⚡ 运行性能测试...$(RESET)"
	@cd core-services && go test -bench=. -benchmem ./...

profile: ## 生成性能分析报告
	@echo "$(GREEN)📊 生成性能分析报告...$(RESET)"
	@cd core-services && go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.

# ==================== 信息命令 ====================

info: ## 显示项目信息
	@echo "$(CYAN)项目信息:$(RESET)"
	@echo "$(BLUE)名称:$(RESET) $(PROJECT_NAME)"
	@echo "$(BLUE)版本:$(RESET) $(VERSION)"
	@echo "$(BLUE)构建时间:$(RESET) $(BUILD_TIME)"
	@echo "$(BLUE)Git提交:$(RESET) $(GIT_COMMIT)"
	@echo "$(BLUE)Go版本:$(RESET) $(shell go version 2>/dev/null || echo "未安装")"
	@echo "$(BLUE)Node版本:$(RESET) $(shell node --version 2>/dev/null || echo "未安装")"
	@echo "$(BLUE)Docker版本:$(RESET) $(shell docker --version 2>/dev/null || echo "未安装")"

status: ## 检查服务状态
	@echo "$(GREEN)🔍 检查服务状态...$(RESET)"
	@echo "$(CYAN)前端服务:$(RESET)"
	@curl -s http://localhost:3000 > /dev/null && echo "✅ 运行中" || echo "❌ 未运行"
	@echo "$(CYAN)后端服务:$(RESET)"
	@curl -s http://localhost:8080/health > /dev/null && echo "✅ 运行中" || echo "❌ 未运行"
	@echo "$(CYAN)AI服务:$(RESET)"
	@curl -s http://localhost:8000/health > /dev/null && echo "✅ 运行中" || echo "❌ 未运行"
	@echo "$(CYAN)数据库:$(RESET)"
	@docker-compose ps postgres | grep -q "Up" && echo "✅ 运行中" || echo "❌ 未运行"
	@echo "$(CYAN)Redis:$(RESET)"
	@docker-compose ps redis | grep -q "Up" && echo "✅ 运行中" || echo "❌ 未运行"