# API Gateway Makefile

# 变量定义
APP_NAME := api-gateway
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GO := go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOVERSION := $(shell go version | awk '{print $$3}')

# 构建相关变量
BUILD_DIR := build
BINARY := $(BUILD_DIR)/$(APP_NAME)
DOCKER_IMAGE := taishanglaojun/$(APP_NAME)
DOCKER_TAG := $(VERSION)

# 源码目录
SRC_DIRS := cmd internal
SRC_FILES := $(shell find $(SRC_DIRS) -name "*.go" 2>/dev/null)

# 测试相关
COVERAGE_DIR := coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

# 工具
GOLANGCI_LINT := golangci-lint
GOFUMPT := gofumpt
GOIMPORTS := goimports

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "API Gateway - 可用命令:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

# 构建相关
.PHONY: build
build: clean ## 构建应用程序
	@echo "构建 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" \
		-o $(BINARY) ./cmd/main.go
	@echo "构建完成: $(BINARY)"

.PHONY: build-linux
build-linux: clean ## 构建Linux版本
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" \
		-o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/main.go
	@echo "Linux 构建完成"

.PHONY: build-windows
build-windows: clean ## 构建Windows版本
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" \
		-o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/main.go
	@echo "Windows 构建完成"

.PHONY: build-all
build-all: build-linux build-windows ## 构建所有平台版本
	@echo "所有平台构建完成"

# 运行相关
.PHONY: run
run: ## 运行应用程序
	@echo "启动 $(APP_NAME)..."
	@$(GO) run ./cmd/main.go

.PHONY: run-dev
run-dev: ## 开发模式运行
	@echo "开发模式启动 $(APP_NAME)..."
	@$(GO) run ./cmd/main.go -config=configs/gateway.yaml -debug

.PHONY: run-prod
run-prod: build ## 生产模式运行
	@echo "生产模式启动 $(APP_NAME)..."
	@./$(BINARY) -config=configs/gateway.yaml

# 测试相关
.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	@$(GO) test -v ./...

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "运行竞态检测测试..."
	@$(GO) test -race -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "生成测试覆盖率报告..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -coverprofile=$(COVERAGE_FILE) ./...
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "覆盖率报告生成完成: $(COVERAGE_HTML)"

.PHONY: benchmark
benchmark: ## 运行基准测试
	@echo "运行基准测试..."
	@$(GO) test -bench=. -benchmem ./...

# 代码质量
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	@$(GOFUMPT) -w .
	@$(GOIMPORTS) -w .

.PHONY: lint
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@$(GOLANGCI_LINT) run

.PHONY: vet
vet: ## 运行go vet
	@echo "运行 go vet..."
	@$(GO) vet ./...

.PHONY: check
check: fmt vet lint test ## 运行所有检查

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "下载依赖..."
	@$(GO) mod download

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "更新依赖..."
	@$(GO) get -u ./...
	@$(GO) mod tidy

.PHONY: deps-verify
deps-verify: ## 验证依赖
	@echo "验证依赖..."
	@$(GO) mod verify

.PHONY: tidy
tidy: ## 整理依赖
	@echo "整理依赖..."
	@$(GO) mod tidy

# Docker相关
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "构建 Docker 镜像..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "Docker 镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行 Docker 容器..."
	@docker run --rm -p 8080:8080 -p 9090:9090 \
		-v $(PWD)/configs:/app/configs \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: docker-build ## 推送Docker镜像
	@echo "推送 Docker 镜像..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest

# 部署相关
.PHONY: deploy-dev
deploy-dev: ## 部署到开发环境
	@echo "部署到开发环境..."
	@kubectl apply -f deployments/dev/

.PHONY: deploy-prod
deploy-prod: ## 部署到生产环境
	@echo "部署到生产环境..."
	@kubectl apply -f deployments/prod/

# 工具安装
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install mvdan.cc/gofumpt@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/air-verse/air@latest
	@echo "工具安装完成"

# 热重载
.PHONY: dev
dev: ## 启动热重载开发服务器
	@echo "启动热重载开发服务器..."
	@air -c .air.toml

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -rf tmp/

.PHONY: clean-all
clean-all: clean ## 清理所有文件（包括依赖缓存）
	@echo "清理所有文件..."
	@$(GO) clean -modcache
	@docker system prune -f

# 生成相关
.PHONY: generate
generate: ## 运行go generate
	@echo "运行 go generate..."
	@$(GO) generate ./...

.PHONY: mock
mock: ## 生成mock文件
	@echo "生成 mock 文件..."
	@mockgen -source=internal/registry/registry.go -destination=internal/registry/mock_registry.go
	@mockgen -source=internal/proxy/proxy.go -destination=internal/proxy/mock_proxy.go

# 文档生成
.PHONY: docs
docs: ## 生成文档
	@echo "生成文档..."
	@godoc -http=:6060 &
	@echo "文档服务器启动在 http://localhost:6060"

# 性能分析
.PHONY: profile
profile: ## 运行性能分析
	@echo "运行性能分析..."
	@$(GO) test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "性能分析文件生成: cpu.prof, mem.prof"

# 安全检查
.PHONY: security
security: ## 运行安全检查
	@echo "运行安全检查..."
	@gosec ./...

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "Go版本: $(GOVERSION)"
	@echo "操作系统: $(GOOS)"
	@echo "架构: $(GOARCH)"

# 健康检查
.PHONY: health
health: ## 检查应用健康状态
	@echo "检查应用健康状态..."
	@curl -f http://localhost:8080/health || echo "应用未运行或不健康"

# 监控
.PHONY: metrics
metrics: ## 查看应用指标
	@echo "查看应用指标..."
	@curl -s http://localhost:9090/metrics

# 日志
.PHONY: logs
logs: ## 查看应用日志
	@echo "查看应用日志..."
	@tail -f /var/log/api-gateway/gateway.log

# 配置验证
.PHONY: config-validate
config-validate: ## 验证配置文件
	@echo "验证配置文件..."
	@$(GO) run ./cmd/main.go -config=configs/gateway.yaml -validate-config

# 数据库迁移（如果需要）
.PHONY: migrate
migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	@# 这里添加数据库迁移命令

# 备份
.PHONY: backup
backup: ## 备份配置和数据
	@echo "备份配置和数据..."
	@tar -czf backup-$(shell date +%Y%m%d-%H%M%S).tar.gz configs/ data/

# 恢复
.PHONY: restore
restore: ## 恢复配置和数据
	@echo "恢复配置和数据..."
	@# 这里添加恢复命令

# 集成测试
.PHONY: integration-test
integration-test: ## 运行集成测试
	@echo "运行集成测试..."
	@$(GO) test -tags=integration ./tests/integration/...

# 端到端测试
.PHONY: e2e-test
e2e-test: ## 运行端到端测试
	@echo "运行端到端测试..."
	@$(GO) test -tags=e2e ./tests/e2e/...

# 压力测试
.PHONY: stress-test
stress-test: ## 运行压力测试
	@echo "运行压力测试..."
	@hey -n 10000 -c 100 http://localhost:8080/health

# 全部任务
.PHONY: all
all: clean deps check build test ## 运行所有主要任务

# 快速构建
.PHONY: quick
quick: build ## 快速构建（跳过检查）

# 发布准备
.PHONY: release
release: clean deps check test build-all docker-build ## 准备发布版本