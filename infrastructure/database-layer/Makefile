# 数据库层 Makefile

.PHONY: help build run test clean deps fmt lint docker-build docker-run migrate-up migrate-down migrate-create migrate-status

# 默认目标
help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 依赖管理
deps: ## 安装依赖
	go mod download
	go mod tidy

# 构建
build: ## 构建应用
	go build -o bin/database-layer ./cmd/main.go

# 运行
run: ## 运行应用
	go run ./cmd/main.go

# 开发模式运行（使用Air热重载）
dev: ## 开发模式运行
	air

# 测试
test: ## 运行测试
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码格式化
fmt: ## 格式化代码
	go fmt ./...
	goimports -w .

# 代码检查
lint: ## 运行代码检查
	golangci-lint run

# 清理
clean: ## 清理构建文件
	rm -rf bin/
	rm -f coverage.out coverage.html

# 安装开发工具
install-tools: ## 安装开发工具
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker 相关
docker-build: ## 构建Docker镜像
	docker build -t taishang/database-layer .

docker-run: ## 运行Docker容器
	docker run -p 8080:8080 --env-file .env taishang/database-layer

# 数据库迁移相关
migrate-up: ## 执行向上迁移
	migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/taishang?sslmode=disable" up

migrate-down: ## 执行向下迁移
	migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/taishang?sslmode=disable" down

migrate-create: ## 创建新的迁移文件 (使用: make migrate-create NAME=create_users_table)
	migrate create -ext sql -dir ./migrations -seq $(NAME)

migrate-status: ## 查看迁移状态
	migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/taishang?sslmode=disable" version

migrate-force: ## 强制设置迁移版本 (使用: make migrate-force VERSION=1)
	migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/taishang?sslmode=disable" force $(VERSION)

# 数据库操作
db-setup: ## 设置数据库
	createdb taishang || true
	createdb taishang_test || true
	createdb taishang_dev || true

db-drop: ## 删除数据库
	dropdb taishang || true
	dropdb taishang_test || true
	dropdb taishang_dev || true

db-reset: ## 重置数据库
	make db-drop
	make db-setup
	make migrate-up

# Redis 操作
redis-start: ## 启动Redis (使用Docker)
	docker run -d --name redis-taishang -p 6379:6379 redis:alpine

redis-stop: ## 停止Redis
	docker stop redis-taishang || true
	docker rm redis-taishang || true

redis-cli: ## 连接Redis CLI
	docker exec -it redis-taishang redis-cli

# PostgreSQL 操作
postgres-start: ## 启动PostgreSQL (使用Docker)
	docker run -d --name postgres-taishang \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=password \
		-e POSTGRES_DB=taishang \
		-p 5432:5432 \
		postgres:13-alpine

postgres-stop: ## 停止PostgreSQL
	docker stop postgres-taishang || true
	docker rm postgres-taishang || true

postgres-cli: ## 连接PostgreSQL CLI
	docker exec -it postgres-taishang psql -U postgres -d taishang

# 开发环境
dev-setup: ## 设置开发环境
	make postgres-start
	make redis-start
	sleep 5
	make db-setup
	make migrate-up
	make deps

dev-teardown: ## 清理开发环境
	make postgres-stop
	make redis-stop

# 生成模拟数据
mock-generate: ## 生成模拟接口
	mockgen -source=internal/repository/base.go -destination=mocks/repository_mock.go

# 基准测试
benchmark: ## 运行基准测试
	go test -bench=. -benchmem ./...

# 性能分析
profile: ## 运行性能分析
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	go tool pprof cpu.prof

# 安全检查
security: ## 运行安全检查
	gosec ./...

# 依赖检查
deps-check: ## 检查依赖更新
	go list -u -m all

# 模块验证
mod-verify: ## 验证模块
	go mod verify

# 完整检查
check: fmt lint test security ## 运行完整检查

# 发布准备
release-prepare: clean deps check build ## 准备发布