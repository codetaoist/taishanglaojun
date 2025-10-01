# 太上老君 - 项目脚手架

## 概述

项目脚手架是太上老君系统的基础设施模块，提供了标准化的Go服务开发框架，包含配置管理、日志系统、HTTP服务器、中间件等核心功能。

## 功能特性

### 核心功能
- **配置管理**: 基于Viper的多源配置支持（文件、环境变量）
- **日志系统**: 基于Zap的结构化日志，支持多种输出格式
- **HTTP服务器**: 基于Gin的高性能Web服务器
- **中间件**: 日志、恢复、CORS、限流、认证等中间件
- **优雅关闭**: 支持信号处理和优雅关闭

### 开发工具
- **热重载**: 使用Air实现开发时热重载
- **代码质量**: 集成格式化、静态检查工具
- **测试覆盖**: 完整的测试框架和覆盖率报告
- **构建工具**: Makefile自动化构建和部署

## 项目结构

```
project-scaffold/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── logger/           # 日志系统
│   ├── server/           # HTTP服务器
│   └── middleware/       # 中间件
├── configs/              # 配置文件
│   └── config.yaml      # 默认配置
├── build/               # 构建输出
├── tmp/                # 临时文件（Air使用）
├── go.mod              # Go模块定义
├── Makefile           # 构建脚本
├── .air.toml         # Air配置
└── README.md         # 项目文档
```

## 快速开始

### 环境要求
- Go 1.21+
- Make（可选）

### 安装依赖
```bash
# 使用Make
make deps

# 或直接使用Go
go mod download
go mod tidy
```

### 开发模式运行
```bash
# 使用热重载（推荐）
make dev

# 或直接运行
make run
```

### 构建
```bash
# 构建可执行文件
make build

# 运行构建的文件
./build/taishang-service
```

## 配置说明

### 配置文件
配置文件位于 `configs/config.yaml`，支持以下配置项：

```yaml
app:
  name: "taishang-service"      # 应用名称
  version: "1.0.0"              # 版本号
  environment: "development"     # 环境（development/production）

server:
  host: "0.0.0.0"              # 监听地址
  port: 8080                    # 监听端口
  read_timeout: 30              # 读取超时（秒）
  write_timeout: 30             # 写入超时（秒）

database:
  host: "localhost"             # 数据库地址
  port: 5432                    # 数据库端口
  username: "postgres"          # 用户名
  password: "password"          # 密码
  database: "taishang"          # 数据库名
  ssl_mode: "disable"           # SSL模式

redis:
  host: "localhost"             # Redis地址
  port: 6379                    # Redis端口
  password: ""                  # Redis密码
  database: 0                   # Redis数据库

log:
  level: "info"                 # 日志级别
  format: "json"                # 日志格式（json/console）

jwt:
  secret: "your-secret-key"     # JWT密钥
  expire_time: 3600             # 过期时间（秒）
```

### 环境变量
所有配置项都可以通过环境变量覆盖，格式为 `TAISHANG_<SECTION>_<KEY>`：

```bash
export TAISHANG_SERVER_PORT=9090
export TAISHANG_DATABASE_HOST=db.example.com
export TAISHANG_LOG_LEVEL=debug
```

## API接口

### 健康检查
```
GET /health
```

响应：
```json
{
  "status": "ok",
  "timestamp": 1640995200,
  "service": "taishang-service",
  "version": "1.0.0"
}
```

### API版本
```
GET /api/v1/version
```

响应：
```json
{
  "name": "taishang-service",
  "version": "1.0.0",
  "environment": "development"
}
```

### Ping测试
```
GET /api/v1/ping
```

响应：
```json
{
  "message": "pong"
}
```

## 开发指南

### 添加新的路由
在 `internal/server/server.go` 的 `setupRoutes` 方法中添加：

```go
func (s *Server) setupRoutes() {
    // 现有路由...
    
    // 添加新路由
    v1 := s.router.Group("/api/v1")
    {
        v1.GET("/users", s.getUsers)
        v1.POST("/users", s.createUser)
    }
}
```

### 添加新的中间件
在 `internal/middleware/middleware.go` 中添加：

```go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 中间件逻辑
        c.Next()
    }
}
```

### 添加新的配置项
1. 在 `internal/config/config.go` 中添加结构体字段
2. 在 `setDefaults` 函数中设置默认值
3. 在 `configs/config.yaml` 中添加配置项

## 测试

### 运行测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 代码质量检查
```bash
# 格式化代码
make fmt

# 静态检查
make lint
```

## 部署

### Docker部署
```bash
# 构建Docker镜像
make docker-build

# 运行Docker容器
make docker-run
```

### 生产环境配置
1. 设置环境变量 `TAISHANG_APP_ENVIRONMENT=production`
2. 配置生产环境的数据库和Redis连接
3. 设置安全的JWT密钥
4. 配置适当的日志级别

## 开发工具

### Make命令
- `make deps` - 安装依赖
- `make build` - 构建应用
- `make run` - 运行应用
- `make dev` - 开发模式（热重载）
- `make test` - 运行测试
- `make clean` - 清理构建文件
- `make help` - 显示帮助信息

### 推荐工具
- [Air](https://github.com/cosmtrek/air) - 热重载工具
- [golangci-lint](https://golangci-lint.run/) - 代码静态检查
- [mockgen](https://github.com/golang/mock) - Mock生成工具

## 贡献指南

1. 遵循Go代码规范
2. 添加适当的测试用例
3. 更新相关文档
4. 确保所有测试通过

## 许可证

MIT License