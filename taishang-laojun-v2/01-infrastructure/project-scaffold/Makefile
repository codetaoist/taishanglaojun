# Makefile for Taishang Project Scaffold

# 变量定义
APP_NAME := taishang-service
VERSION := 1.0.0
BUILD_DIR := build
BINARY_NAME := $(APP_NAME)

# Go相关变量
GO_VERSION := 1.21
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# 默认目标
.PHONY: all
all: clean deps test build

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 构建
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

# 运行
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run ./cmd/main.go

# 开发模式运行（使用Air热重载）
.PHONY: dev
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing..."; \
		$(GOGET) -u github.com/cosmtrek/air; \
		air; \
	fi

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 测试覆盖率
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please install it first."; \
	fi

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# 安装工具
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/cosmtrek/air
	@echo "Please install golangci-lint manually: https://golangci-lint.run/usage/install/"

# Docker构建
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Docker运行
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

# 生成模拟数据
.PHONY: mock
mock:
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		mockgen -source=internal/service/interface.go -destination=internal/service/mock.go; \
	else \
		echo "mockgen not found. Installing..."; \
		$(GOGET) github.com/golang/mock/mockgen; \
		mockgen -source=internal/service/interface.go -destination=internal/service/mock.go; \
	fi

# 数据库迁移
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	# TODO: 添加数据库迁移命令

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	# TODO: 添加数据库回滚命令

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Clean, install deps, test, and build"
	@echo "  deps          - Install dependencies"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode with hot reload"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  mock          - Generate mocks"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  help          - Show this help message"