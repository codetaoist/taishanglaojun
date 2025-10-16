# 变量定义
APP_NAME := task-management
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# 默认目标
.PHONY: all
all: clean test build

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf dist/
	@go clean

# 安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

# 测试覆盖率
.PHONY: coverage
coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html

# 构建
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin/
	@go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server

# 交叉编译
.PHONY: build-all
build-all:
	@echo "Cross-compiling..."
	@mkdir -p dist/
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-amd64 ./cmd/server
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-windows-amd64.exe ./cmd/server
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-amd64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-arm64 ./cmd/server

# 运行
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	@./bin/$(APP_NAME)

# 开发模式运行
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@go run ./cmd/server

# Docker构建
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Docker运行
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(APP_NAME):latest

# Docker Compose启动
.PHONY: docker-up
docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

# Docker Compose停止
.PHONY: docker-down
docker-down:
	@echo "Stopping services with Docker Compose..."
	@docker-compose down

# Docker Compose重启
.PHONY: docker-restart
docker-restart: docker-down docker-up

# 查看日志
.PHONY: logs
logs:
	@docker-compose logs -f task-management

# 数据库迁移
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	@migrate -path ./migrations -database "postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable" up

# 数据库回滚
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path ./migrations -database "postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable" down

# 生成API文档
.PHONY: docs
docs:
	@echo "Generating API documentation..."
	@swag init -g ./cmd/server/main.go

# 安全扫描
.PHONY: security
security:
	@echo "Running security scan..."
	@gosec ./...

# 性能测试
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# 代码生成
.PHONY: generate
generate:
	@echo "Running code generation..."
	@go generate ./...

# 完整检查
.PHONY: check
check: fmt lint test security

# 发布准备
.PHONY: release
release: clean check build-all docker-build

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, test, and build"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  test         - Run tests"
	@echo "  coverage     - Generate test coverage report"
	@echo "  build        - Build the application"
	@echo "  build-all    - Cross-compile for all platforms"
	@echo "  run          - Build and run the application"
	@echo "  dev          - Run in development mode"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  docker-up    - Start services with Docker Compose"
	@echo "  docker-down  - Stop services with Docker Compose"
	@echo "  docker-restart - Restart Docker Compose services"
	@echo "  logs         - View application logs"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  docs         - Generate API documentation"
	@echo "  security     - Run security scan"
	@echo "  bench        - Run benchmarks"
	@echo "  generate     - Run code generation"
	@echo "  check        - Run all checks (fmt, lint, test, security)"
	@echo "  release      - Prepare for release"
	@echo "  help         - Show this help message"