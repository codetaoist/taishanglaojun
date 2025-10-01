# 太上老君 - 核心服务 Makefile
# 提供常用的开发、测试、构建和部署命令

# 变量定义
APP_NAME := taishang-laojun-core
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION := $(shell go version | awk '{print $$3}')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Go相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# 目录定义
BUILD_DIR := ./build
DIST_DIR := ./dist
CMD_DIR := ./cmd/server
MAIN_FILE := $(CMD_DIR)/main.go

# Docker相关
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_LATEST := $(APP_NAME):latest
DOCKER_REGISTRY := registry.example.com

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "太上老君 - 核心服务构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关命令
.PHONY: dev
dev: ## 启动开发服务器
	@echo "启动开发服务器..."
	@$(GOCMD) run $(CMD_DIR)/main.go

.PHONY: dev-docker
dev-docker: ## 使用Docker Compose启动开发环境
	@echo "启动Docker开发环境..."
	@docker-compose up -d postgres redis elasticsearch
	@echo "等待服务启动..."
	@sleep 10
	@make dev

.PHONY: watch
watch: ## 启动文件监控和自动重载
	@echo "启动文件监控..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "请先安装 air: go install github.com/cosmtrek/air@latest"; \
	fi

# 构建相关命令
.PHONY: build
build: clean ## 构建应用程序
	@echo "构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

.PHONY: build-linux
build-linux: clean ## 构建Linux版本
	@echo "构建Linux版本..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)-linux"

.PHONY: build-windows
build-windows: clean ## 构建Windows版本
	@echo "构建Windows版本..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows.exe $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)-windows.exe"

.PHONY: build-darwin
build-darwin: clean ## 构建macOS版本
	@echo "构建macOS版本..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)-darwin"

.PHONY: build-all
build-all: build-linux build-windows build-darwin ## 构建所有平台版本
	@echo "所有平台构建完成"

# 测试相关命令
.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	@$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试覆盖率..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "运行竞态检测测试..."
	@$(GOTEST) -race -v ./...

.PHONY: benchmark
benchmark: ## 运行基准测试
	@echo "运行基准测试..."
	@$(GOTEST) -bench=. -benchmem ./...

# 代码质量相关命令
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	@$(GOFMT) -s -w .

.PHONY: lint
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run; \
	else \
		echo "请先安装 golangci-lint: https://golangci-lint.run/usage/install/"; \
	fi

.PHONY: vet
vet: ## 运行go vet检查
	@echo "运行go vet检查..."
	@$(GOCMD) vet ./...

.PHONY: check
check: fmt vet lint test ## 运行所有代码检查

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "下载依赖..."
	@$(GOMOD) download

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "更新依赖..."
	@$(GOMOD) tidy
	@$(GOGET) -u ./...

.PHONY: deps-vendor
deps-vendor: ## 创建vendor目录
	@echo "创建vendor目录..."
	@$(GOMOD) vendor

# Docker相关命令
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	@docker build -t $(DOCKER_IMAGE) -t $(DOCKER_LATEST) .
	@echo "Docker镜像构建完成: $(DOCKER_IMAGE)"

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker run -d --name $(APP_NAME) -p 8080:8080 $(DOCKER_LATEST)

.PHONY: docker-stop
docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true

.PHONY: docker-push
docker-push: docker-build ## 推送Docker镜像到仓库
	@echo "推送Docker镜像..."
	@docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

# Docker Compose相关命令
.PHONY: up
up: ## 启动所有服务
	@echo "启动所有服务..."
	@docker-compose up -d

.PHONY: down
down: ## 停止所有服务
	@echo "停止所有服务..."
	@docker-compose down

.PHONY: logs
logs: ## 查看服务日志
	@docker-compose logs -f

.PHONY: ps
ps: ## 查看服务状态
	@docker-compose ps

# 数据库相关命令
.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	@$(GOCMD) run $(CMD_DIR)/migrate.go up

.PHONY: db-rollback
db-rollback: ## 回滚数据库迁移
	@echo "回滚数据库迁移..."
	@$(GOCMD) run $(CMD_DIR)/migrate.go down

.PHONY: db-reset
db-reset: ## 重置数据库
	@echo "重置数据库..."
	@$(GOCMD) run $(CMD_DIR)/migrate.go reset

# 清理相关命令
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out coverage.html

.PHONY: clean-docker
clean-docker: ## 清理Docker资源
	@echo "清理Docker资源..."
	@docker system prune -f
	@docker volume prune -f

# 发布相关命令
.PHONY: release
release: clean build-all ## 创建发布包
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)
	@cp -r $(BUILD_DIR)/* $(DIST_DIR)/
	@cp config/config.yaml $(DIST_DIR)/
	@cp .env.example $(DIST_DIR)/
	@tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION).tar.gz -C $(DIST_DIR) .
	@echo "发布包创建完成: $(DIST_DIR)/$(APP_NAME)-$(VERSION).tar.gz"

# 安装相关命令
.PHONY: install
install: build ## 安装到系统
	@echo "安装到系统..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "安装完成"

.PHONY: uninstall
uninstall: ## 从系统卸载
	@echo "从系统卸载..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "卸载完成"

# 工具安装
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	@$(GOGET) github.com/cosmtrek/air@latest
	@$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发工具安装完成"

# 信息显示
.PHONY: info
info: ## 显示构建信息
	@echo "应用信息:"
	@echo "  名称: $(APP_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Go版本: $(GO_VERSION)"