# 核心服务 (Core Services)

太上老君v2项目的核心服务层，提供AI集成和文化智慧服务。

## 功能特性

### 数据库支持
- **PostgreSQL** - 默认推荐数据库
- **MySQL** - 支持MySQL 5.7+
- **SQL Server** - 支持SQL Server 2017+

### 核心服务
- **AI集成服务** - 智能对话、内容生成、语义分析
- **文化智慧服务** - 内容管理、智能搜索、分类推荐

## 🏗️ 技术架构

```
├── ai-integration/          # AI集成服务
│   ├── providers/          # AI提供商接口
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑
│   └── handlers/          # HTTP处理器
├── cultural-wisdom/        # 文化智慧服务
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑
│   └── handlers/          # HTTP处理器
├── config/                # 配置文件
├── docs/                  # 文档
└── scripts/               # 脚本文件
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Elasticsearch 8.11+
- Docker & Docker Compose（可选）

### 本地开发

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd taishang-laojun-v2/02-core-services
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，配置数据库、Redis、AI服务等参数
   ```

4. **启动依赖服务**
   ```bash
   # 使用Docker Compose启动数据库和缓存
   docker-compose up -d postgres redis elasticsearch
   ```

5. **运行应用**
   ```bash
   # 开发模式
   make dev
   
   # 或直接运行
   go run cmd/server/main.go
   ```

### Docker部署

1. **构建镜像**
   ```bash
   make docker-build
   ```

2. **启动所有服务**
   ```bash
   make up
   ```

3. **查看服务状态**
   ```bash
   make ps
   ```

## 📖 API文档

API文档使用OpenAPI 3.0规范编写，详见 [api-docs.yaml](./api-docs.yaml)

### 主要接口

#### AI服务接口
- `POST /v1/ai/chat` - 智能对话
- `GET /v1/ai/sessions` - 获取对话列表
- `POST /v1/ai/generate/summary` - 生成摘要
- `POST /v1/ai/analyze/sentiment` - 情感分析

#### 文化智慧接口
- `GET /v1/wisdom` - 获取智慧内容列表
- `POST /v1/wisdom` - 创建智慧内容
- `GET /v1/wisdom/{id}` - 获取智慧内容详情
- `GET /v1/search` - 全文搜索
- `POST /v1/search/semantic` - 语义搜索

### 认证方式

使用JWT Bearer Token认证：
```bash
curl -H "Authorization: Bearer <your-token>" \
     https://api.taishanglaojun.com/v1/wisdom
```

## ⚙️ 配置说明

### 主配置文件

配置文件位于 `config/config.yaml`，支持环境变量替换：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "development"

database:
  primary:
    host: "${DB_HOST:localhost}"
    port: "${DB_PORT:5432}"
    database: "${DB_NAME:taishanglaojun}"
```

### 环境变量

主要环境变量说明：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `APP_MODE` | 运行模式 | `development` |
| `SERVER_PORT` | 服务端口 | `8080` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `REDIS_HOST` | Redis主机 | `localhost` |
| `OPENAI_API_KEY` | OpenAI API密钥 | - |
| `JWT_SECRET` | JWT密钥 | - |

## 🛠️ 开发工具

项目提供了丰富的Makefile命令：

```bash
# 开发相关
make dev              # 启动开发服务器
make watch            # 文件监控和自动重载
make dev-docker       # Docker开发环境

# 构建相关
make build            # 构建应用
make build-all        # 构建所有平台版本
make docker-build     # 构建Docker镜像

# 测试相关
make test             # 运行测试
make test-coverage    # 测试覆盖率
make benchmark        # 基准测试

# 代码质量
make fmt              # 格式化代码
make lint             # 代码检查
make check            # 运行所有检查

# Docker Compose
make up               # 启动所有服务
make down             # 停止所有服务
make logs             # 查看日志
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行竞态检测测试
make test-race

# 运行基准测试
make benchmark
```

### 测试结构

```
├── ai-integration/
│   ├── services/
│   │   └── chat_service_test.go
│   └── handlers/
│       └── chat_handler_test.go
└── cultural-wisdom/
    ├── services/
    │   └── wisdom_service_test.go
    └── handlers/
        └── wisdom_handler_test.go
```

## 📊 监控

### 健康检查

```bash
curl http://localhost:8080/health
```

### Prometheus指标

```bash
curl http://localhost:8080/metrics
```

### 可选监控组件

使用Docker Compose profiles启动监控组件：

```bash
# 启动监控组件
docker-compose --profile monitoring up -d

# 访问Grafana面板
open http://localhost:3000
# 用户名: admin, 密码: admin123

# 访问Prometheus
open http://localhost:9090
```

## 🚀 部署

### 生产环境部署

1. **构建发布包**
   ```bash
   make release
   ```

2. **配置生产环境变量**
   ```bash
   export APP_MODE=production
   export JWT_SECRET=your-production-jwt-secret
   export DB_HOST=your-production-db-host
   # ... 其他生产环境配置
   ```

3. **运行应用**
   ```bash
   ./build/taishang-laojun-core
   ```

### Kubernetes部署

```yaml
# 示例Kubernetes部署配置
apiVersion: apps/v1
kind: Deployment
metadata:
  name: taishang-laojun-core
spec:
  replicas: 3
  selector:
    matchLabels:
      app: taishang-laojun-core
  template:
    metadata:
      labels:
        app: taishang-laojun-core
    spec:
      containers:
      - name: app
        image: taishang-laojun-core:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_MODE
          value: "production"
        # ... 其他环境变量
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 代码规范

- 使用 `gofmt` 格式化代码
- 运行 `golangci-lint` 进行代码检查
- 编写单元测试，保持测试覆盖率 > 80%
- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 📞 联系方式

- 项目主页: https://github.com/taishanglaojun/taishang-laojun-v2
- 问题反馈: https://github.com/taishanglaojun/taishang-laojun-v2/issues
- 邮箱: dev@taishanglaojun.com

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

---

**太上老君团队** ❤️ **传承智慧，启迪未来**