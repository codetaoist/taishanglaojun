# Makefile for Intelligent Learning System

# 变量定义
APP_NAME := intelligent-learning
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 目录定义
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Docker 相关变量
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_LATEST := $(APP_NAME):latest

# 默认目标
.PHONY: all
all: clean deps fmt lint test build

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  run            - Run the application"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-push    - Push Docker image"
	@echo "  compose-up     - Start services with docker-compose"
	@echo "  compose-down   - Stop services with docker-compose"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  install        - Install the application"
	@echo "  uninstall      - Uninstall the application"

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f $(APP_NAME)

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
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
	$(GOTEST) -v -race ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# 构建应用程序
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe cmd/server/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 cmd/server/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 cmd/server/main.go

# 构建当前平台的二进制文件
.PHONY: build-local
build-local:
	@echo "Building $(APP_NAME) for local platform..."
	$(GOBUILD) $(LDFLAGS) -o $(APP_NAME) cmd/server/main.go

# 运行应用程序
.PHONY: run
run: build-local
	@echo "Running $(APP_NAME)..."
	./$(APP_NAME)

# 开发模式运行
.PHONY: dev
dev:
	@echo "Running in development mode..."
	$(GOCMD) run cmd/server/main.go

# 生成 Swagger 文档
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o docs/swagger

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	docker tag $(DOCKER_IMAGE) $(DOCKER_LATEST)

# Docker 运行
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --name $(APP_NAME) $(DOCKER_LATEST)

# Docker 推送
.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE)
	docker push $(DOCKER_LATEST)

# Docker Compose 启动
.PHONY: compose-up
compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

# Docker Compose 停止
.PHONY: compose-down
compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

# Docker Compose 重启
.PHONY: compose-restart
compose-restart: compose-down compose-up

# 查看 Docker Compose 日志
.PHONY: compose-logs
compose-logs:
	docker-compose logs -f

# 安装应用程序
.PHONY: install
install: build-local
	@echo "Installing $(APP_NAME)..."
	sudo cp $(APP_NAME) /usr/local/bin/

# 卸载应用程序
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	sudo rm -f /usr/local/bin/$(APP_NAME)

# 创建发布包
.PHONY: release
release: clean deps fmt lint test build
	@echo "Creating release package..."
	mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-amd64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-amd64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-arm64
	zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe

# 数据库迁移
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	# 这里可以添加数据库迁移命令

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	# 这里可以添加数据库回滚命令

# 健康检查
.PHONY: health-check
health-check:
	@echo "Checking application health..."
	curl -f http://localhost:8080/health || exit 1

# 性能测试
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# 安全扫描
.PHONY: security-scan
security-scan:
	@echo "Running security scan..."
	gosec ./...

# 依赖检查
.PHONY: deps-check
deps-check:
	@echo "Checking for outdated dependencies..."
	$(GOCMD) list -u -m all

# 更新依赖
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# 显示版本信息
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"