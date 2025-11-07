# API Gateway

这是一个基于Go和Gin框架构建的API网关服务，提供服务发现、负载均衡、请求代理、认证授权、限流等功能。

## 功能特性

- **服务发现**: 支持Consul和Mock两种服务发现机制
- **负载均衡**: 支持轮询和随机负载均衡策略
- **请求代理**: 将请求转发到后端服务
- **认证授权**: 支持JWT令牌验证
- **限流**: 支持基于服务的请求限流
- **CORS**: 支持跨域资源共享配置
- **健康检查**: 定期检查后端服务健康状态
- **指标收集**: 支持Prometheus指标收集
- **优雅关闭**: 支持服务优雅关闭

## 快速开始

### 前置条件

- Go 1.18+
- Consul (可选，用于服务发现)

### 安装依赖

```bash
go mod download
```

### 配置

复制并修改配置文件：

```bash
cp config.yaml.example config.yaml
```

配置文件说明：

```yaml
# 服务器配置
port: "8080"
environment: "development"  # development, staging, production
log_level: "info"          # debug, info, warn, error

# JWT配置
jwt:
  secret: "your-secret-key-change-in-production"
  expiration_time: 86400   # 24小时

# 服务发现配置
discovery:
  type: "consul"           # consul, etcd, mock
  address: "localhost:8500"
  datacenter: "dc1"
  token: ""

# 代理配置
proxy:
  timeout: "30s"
  health_check_interval: "10s"
  retry_attempts: 3

# CORS配置
cors:
  allowed_origins:
    - "*"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "*"
  allow_credentials: true
  max_age: 86400

# 指标配置
metrics:
  enabled: true
  path: "/metrics"
  port: "9090"

# 服务配置
services:
  api:
    path_prefix: "/api"
    auth_required: false
    rate_limit_enabled: false
    timeout: "30s"
    retry_attempts: 3
    load_balancer: "round_robin"
    
  auth:
    path_prefix: "/auth"
    auth_required: false
    rate_limit_enabled: true
    rate_limit:
      requests: 10
      window: "1m"
    timeout: "10s"
    retry_attempts: 2
    load_balancer: "round_robin"
```

### 运行

```bash
go run cmd/gateway/main.go
```

### 构建

```bash
go build -o bin/gateway cmd/gateway/main.go
```

## API文档

### 健康检查

```
GET /health
```

### 指标

```
GET /metrics
```

## 服务发现

### Consul模式

当使用Consul作为服务发现时，API网关会自动从Consul中获取服务实例信息。确保你的服务已正确注册到Consul。

### Mock模式

Mock模式用于开发和测试，它会使用预定义的服务实例列表。

## 负载均衡

API网关支持以下负载均衡策略：

- `round_robin`: 轮询
- `random`: 随机

## 认证授权

API网关支持JWT令牌验证。对于需要认证的服务，请求需要包含有效的JWT令牌。

### 生成JWT令牌

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'
```

### 使用JWT令牌

```bash
curl -X GET http://localhost:8080/api/protected \
  -H "Authorization: Bearer <your-jwt-token>"
```

## 限流

API网关支持基于服务的请求限流。限流配置在配置文件中的`services`部分。

## 环境变量

所有配置都可以通过环境变量覆盖，环境变量名格式为`GATEWAY_<SECTION>_<KEY>`，例如：

```bash
export GATEWAY_PORT=9090
export GATEWAY_DISCOVERY_ADDRESS=consul.example.com:8500
export GATEWAY_JWT_SECRET=my-secret-key
```

## 监控

API网关提供Prometheus格式的指标，可以通过`/metrics`端点获取。

## 开发

### 项目结构

```
gateway/
├── cmd/
│   └── gateway/
│       └── main.go         # 主程序入口
├── internal/
│   ├── config/             # 配置管理
│   ├── discovery/          # 服务发现
│   ├── middleware/         # 中间件
│   ├── proxy/              # 请求代理
│   └── router/             # 路由配置
├── config.yaml             # 配置文件
├── go.mod
└── README.md
```

### 添加新的服务发现

1. 在`internal/discovery`目录下创建新的实现文件
2. 实现`ServiceDiscovery`接口
3. 在主程序中添加对新实现的支持

### 添加新的中间件

1. 在`internal/middleware`目录下创建新的中间件文件
2. 在`internal/router/router.go`中注册新的中间件

## 许可证

MIT License