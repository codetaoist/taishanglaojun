# Makefile for Auth System

# 变量定义
APP_NAME := auth-system
VERSION := 1.0.0
BUILD_DIR := build
BINARY_NAME := $(APP_NAME)
MAIN_PATH := ./cmd/main.go

# Go相关变量
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOMOD := $(GO) mod

# 构建标志
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
BUILD_FLAGS := -v $(LDFLAGS)

# 默认目标
.PHONY: all
all: clean deps fmt lint test build

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# 代码检查
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# 查看测试覆盖率
.PHONY: coverage
coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 构建应用
.PHONY: build
build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 构建Linux版本
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)

# 构建Windows版本
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_PATH)

# 构建macOS版本
.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_PATH)

# 构建所有平台版本
.PHONY: build-all
build-all: build-linux build-windows build-darwin

# 运行应用
.PHONY: run
run:
	@echo "Running application..."
	$(GO) run $(MAIN_PATH)

# 开发模式运行（带热重载）
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing..."; \
		$(GO) install github.com/cosmtrek/air@latest; \
		air; \
	fi

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# 安装开发工具
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/cosmtrek/air@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest

# 生成API文档
.PHONY: docs
docs:
	@echo "Generating API documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/main.go -o docs; \
	else \
		echo "Swag not found. Please run 'make install-tools' first"; \
	fi

# 数据库相关命令
.PHONY: db-up
db-up:
	@echo "Starting database containers..."
	docker-compose up -d postgres redis

.PHONY: db-down
db-down:
	@echo "Stopping database containers..."
	docker-compose down

.PHONY: db-reset
db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres redis

# Docker相关命令
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest

# 部署相关命令
.PHONY: deploy-staging
deploy-staging:
	@echo "Deploying to staging..."
	# 添加部署到测试环境的命令

.PHONY: deploy-prod
deploy-prod:
	@echo "Deploying to production..."
	# 添加部署到生产环境的命令

# 基准测试
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# 性能分析
.PHONY: profile
profile:
	@echo "Running performance profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...

# 安全检查
.PHONY: security
security:
	@echo "Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Installing..."; \
		$(GO) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# 依赖检查
.PHONY: deps-check
deps-check:
	@echo "Checking for dependency updates..."
	$(GO) list -u -m all

# 模块验证
.PHONY: verify
verify:
	@echo "Verifying modules..."
	$(GOMOD) verify

# 完整的CI流程
.PHONY: ci
ci: deps fmt lint test security build

# 帮助信息
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  all          - Run clean, deps, fmt, lint, test, build"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  test         - Run tests"
	@echo "  coverage     - Generate test coverage report"
	@echo "  build        - Build application"
	@echo "  build-all    - Build for all platforms"
	@echo "  run          - Run application"
	@echo "  dev          - Run in development mode with hot reload"
	@echo "  clean        - Clean build files"
	@echo "  install-tools- Install development tools"
	@echo "  docs         - Generate API documentation"
	@echo "  db-up        - Start database containers"
	@echo "  db-down      - Stop database containers"
	@echo "  db-reset     - Reset database"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  bench        - Run benchmarks"
	@echo "  security     - Run security checks"
	@echo "  ci           - Run CI pipeline"
	@echo "  help         - Show this help message"