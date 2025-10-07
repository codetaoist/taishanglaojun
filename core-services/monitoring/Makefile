# 太上老君监控系统 Makefile
# 用于简化开发、测试、构建和部署流程

# ==================== 变量定义 ====================
PROJECT_NAME := taishanglaojun-monitoring
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT_SHA := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go 相关变量
GO_VERSION := 1.21
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

# Docker 相关变量
REGISTRY ?= ghcr.io
IMAGE_NAME ?= $(REGISTRY)/taishanglaojun/monitoring
PLATFORMS ?= linux/amd64,linux/arm64

# 目录变量
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage
DOCS_DIR := docs

# 测试相关变量
TEST_TIMEOUT := 10m
COVERAGE_THRESHOLD := 80

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitSHA=$(COMMIT_SHA)"

# 颜色定义
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

# ==================== 默认目标 ====================
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help:
	@echo "太上老君监控系统构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build          构建应用程序"
	@echo "  build-linux    构建 Linux 版本"
	@echo "  build-windows  构建 Windows 版本"
	@echo "  build-darwin   构建 macOS 版本"
	@echo "  build-all      构建所有平台版本"
	@echo "  test           运行测试"
	@echo "  test-coverage  运行测试并生成覆盖率报告"
	@echo "  bench          运行基准测试"
	@echo "  lint           代码检查"
	@echo "  fmt            格式化代码"
	@echo "  deps           下载依赖"
	@echo "  clean          清理构建文件"
	@echo "  docker-build   构建 Docker 镜像"
	@echo "  docker-run     运行 Docker 容器"
	@echo "  install        安装到系统"
	@echo "  uninstall      从系统卸载"
	@echo "  dev            开发模式运行"
	@echo "  release        创建发布包"

# 下载依赖
.PHONY: deps
deps:
	@echo "下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	$(GOFMT) ./...

# 代码检查
.PHONY: lint
lint:
	@echo "代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 运行基准测试
.PHONY: bench
bench:
	@echo "运行基准测试..."
	$(GOTEST) -bench=. -benchmem ./...

# 构建应用程序
.PHONY: build
build:
	@echo "构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) main.go

# 构建 Linux 版本
.PHONY: build-linux
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go

# 构建 Windows 版本
.PHONY: build-windows
build-windows:
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe main.go

# 构建 macOS 版本
.PHONY: build-darwin
build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 main.go

# 构建所有平台版本
.PHONY: build-all
build-all: build-linux build-windows build-darwin
	@echo "所有平台版本构建完成"

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

# 构建 Docker 镜像
.PHONY: docker-build
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# 运行 Docker 容器
.PHONY: docker-run
docker-run:
	@echo "运行 Docker 容器..."
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-v $(PWD)/$(CONFIG_DIR):/app/config \
		$(APP_NAME):latest

# 停止 Docker 容器
.PHONY: docker-stop
docker-stop:
	@echo "停止 Docker 容器..."
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

# 安装到系统
.PHONY: install
install: build
	@echo "安装到系统..."
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	sudo mkdir -p /etc/$(APP_NAME)
	sudo cp -r $(CONFIG_DIR)/* /etc/$(APP_NAME)/
	@echo "安装完成，可以使用 '$(APP_NAME)' 命令启动"

# 从系统卸载
.PHONY: uninstall
uninstall:
	@echo "从系统卸载..."
	sudo rm -f /usr/local/bin/$(APP_NAME)
	sudo rm -rf /etc/$(APP_NAME)
	@echo "卸载完成"

# 开发模式运行
.PHONY: dev
dev:
	@echo "开发模式运行..."
	$(GOCMD) run main.go -config $(CONFIG_DIR)/monitoring.yaml -log-level debug

# 创建发布包
.PHONY: release
release: clean build-all
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)
	
	# Linux 发布包
	@mkdir -p $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64
	cp $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64/$(APP_NAME)
	cp -r $(CONFIG_DIR) $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64/
	cp README.md $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64/
	cd $(DIST_DIR) && tar -czf $(APP_NAME)-$(VERSION)-linux-amd64.tar.gz $(APP_NAME)-$(VERSION)-linux-amd64
	
	# Windows 发布包
	@mkdir -p $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64
	cp $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64/$(APP_NAME).exe
	cp -r $(CONFIG_DIR) $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64/
	cp README.md $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64/
	cd $(DIST_DIR) && zip -r $(APP_NAME)-$(VERSION)-windows-amd64.zip $(APP_NAME)-$(VERSION)-windows-amd64
	
	# macOS 发布包
	@mkdir -p $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64
	cp $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64/$(APP_NAME)
	cp -r $(CONFIG_DIR) $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64/
	cp README.md $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64/
	cd $(DIST_DIR) && tar -czf $(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz $(APP_NAME)-$(VERSION)-darwin-amd64
	
	@echo "发布包创建完成:"
	@ls -la $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip

# 检查代码质量
.PHONY: quality
quality: fmt lint test
	@echo "代码质量检查完成"

# 性能分析
.PHONY: profile
profile:
	@echo "运行性能分析..."
	$(GOCMD) run main.go -cpuprofile=cpu.prof -memprofile=mem.prof &
	sleep 30
	pkill -f "go run main.go"
	$(GOCMD) tool pprof cpu.prof
	$(GOCMD) tool pprof mem.prof

# 生成文档
.PHONY: docs
docs:
	@echo "生成文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc 未安装，跳过文档生成"; \
		echo "安装命令: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 检查依赖更新
.PHONY: deps-check
deps-check:
	@echo "检查依赖更新..."
	$(GOCMD) list -u -m all

# 更新依赖
.PHONY: deps-update
deps-update:
	@echo "更新依赖..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# 安全检查
.PHONY: security
security:
	@echo "安全检查..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec 未安装，跳过安全检查"; \
		echo "安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 版本信息
.PHONY: version
version:
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git 提交: $(GIT_COMMIT)"
	@echo "Go 版本: $(GO_VERSION)"

# 初始化开发环境
.PHONY: init
init:
	@echo "初始化开发环境..."
	$(GOMOD) download
	@echo "安装开发工具..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) golang.org/x/tools/cmd/godoc@latest
	@echo "开发环境初始化完成"

# 监控运行状态
.PHONY: status
status:
	@echo "检查服务状态..."
	@if pgrep -f $(APP_NAME) > /dev/null; then \
		echo "服务正在运行"; \
		curl -s http://localhost:8080/health || echo "健康检查失败"; \
	else \
		echo "服务未运行"; \
	fi