# 太上老君序列0 - Taishang Laojun Sequence Zero

## 项目概述

太上老君序列0是一个基于现代微服务架构的智能系统，融合了传统文化智慧与现代技术，实现了分层权限管理、意识状态分析和文化智慧服务。

## 核心特性

### 🔐 分层权限系统 (L1-L9)
- **L1-L3**: 基础访问权限，用户注册、登录、基本数据操作
- **L4-L6**: 进阶功能权限，意识分析、文化智慧访问
- **L7-L9**: 高级管理权限，系统管理、核心控制、至高权限

### 🧠 意识融合核心
- 用户意识状态监测与分析
- 多维度认知数据处理
- 智能行为模式识别

### 📚 文化智慧服务
- 传统文化知识库集成
- 智能文化内容分析
- 个性化智慧推荐

### 🔍 全面监控体系
- Prometheus + Grafana 监控
- 实时性能指标收集
- 智能告警系统

## 技术架构

### 后端服务
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **图数据库**: Neo4j 5+
- **认证**: JWT + bcrypt

### 基础设施
- **容器化**: Docker + Docker Compose
- **监控**: Prometheus + Grafana
- **部署**: Kubernetes (生产环境)

## 项目结构

```
taishang-sequence-zero/
├── backend/                    # 后端服务
│   ├── cmd/                   # 服务入口
│   │   ├── auth-service/      # 认证服务
│   │   └── permission-service/ # 权限服务
│   ├── internal/              # 内部业务逻辑
│   │   ├── auth/             # 认证模块
│   │   └── permission/       # 权限模块
│   └── pkg/                  # 共享包
│       └── database/         # 数据库连接
├── frontend/                  # 前端应用 (待开发)
├── ai-services/              # AI服务 (待开发)
├── deployments/              # 部署配置
│   ├── kubernetes/           # K8s配置
│   └── monitoring/           # 监控配置
├── scripts/                  # 脚本文件
└── docs/                     # 文档
```

## 快速开始

### 前置要求
- Docker 和 Docker Compose
- Go 1.21+ (用于本地开发)
- PostgreSQL 15+ (或使用Docker)

### 启动开发环境

1. **启动基础设施服务**
   ```powershell
   # 启动数据库和中间件
   docker-compose up -d postgres redis neo4j prometheus grafana
   ```

2. **初始化数据库**
   ```powershell
   # 执行数据库初始化脚本
   psql -h localhost -U taishang -d taishang_sequence_zero -f scripts/init-db.sql
   ```

3. **启动认证服务**
   ```powershell
   cd backend
   go run cmd/auth-service/main.go
   ```

4. **启动权限服务**
   ```powershell
   cd backend
   go run cmd/permission-service/main.go
   ```

### 服务端口
- **认证服务**: http://localhost:8080
- **权限服务**: http://localhost:8081
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **Neo4j**: http://localhost:7474
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

## API 文档

### 认证服务 API

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

#### 令牌验证
```http
GET /api/v1/auth/verify
Authorization: Bearer <access_token>
```

### 权限服务 API

#### 权限检查
```http
POST /api/v1/check
Content-Type: application/json

{
  "user_id": 1,
  "permission_name": "数据查看"
}
```

#### 获取用户权限等级
```http
GET /api/v1/users/1/level
```

#### 更新用户权限等级
```http
PUT /api/v1/users/1/level
Content-Type: application/json

{
  "level": 5,
  "updated_by": 1
}
```

## 权限等级说明

| 等级 | 名称 | 描述 | 可访问功能 |
|------|------|------|------------|
| L1 | 基础访问 | 新用户默认等级 | 注册、登录、基本信息查看 |
| L2 | 数据查看 | 可查看个人数据 | L1功能 + 个人数据查看 |
| L3 | 数据修改 | 可修改个人数据 | L2功能 + 个人数据修改 |
| L4 | 意识分析 | 意识状态分析 | L3功能 + 意识状态分析 |
| L5 | 文化智慧 | 文化智慧访问 | L4功能 + 文化智慧服务 |
| L6 | 高级功能 | 高级功能访问 | L5功能 + 高级分析工具 |
| L7 | 系统管理 | 系统管理权限 | L6功能 + 用户管理 |
| L8 | 核心控制 | 核心系统控制 | L7功能 + 系统配置 |
| L9 | 至高权限 | 最高级别权限 | 所有功能 + 系统核心控制 |

## 数据库设计

### 核心表结构

- **users**: 用户基本信息和权限等级
- **permissions**: 权限定义和等级要求
- **user_sessions**: 用户会话管理
- **consciousness_states**: 意识状态数据
- **cultural_analyses**: 文化分析结果
- **audit_logs**: 系统审计日志

## 监控和运维

### Prometheus 指标
- HTTP请求计数和延迟
- 数据库连接池状态
- 认证成功/失败率
- 权限检查统计

### Grafana 仪表板
- 系统概览仪表板
- 服务性能监控
- 用户行为分析
- 安全事件监控

## 开发指南

### 代码规范
- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 编写单元测试
- 添加适当的注释

### 提交规范
- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- test: 测试相关
- chore: 构建过程或辅助工具的变动

## 安全考虑

- 密码使用bcrypt加密存储
- JWT令牌有效期控制
- API访问频率限制
- 敏感操作审计日志
- 数据库连接加密
- 输入参数验证和过滤

## 部署说明

### 开发环境
使用Docker Compose进行本地开发环境部署

### 生产环境
- 使用Kubernetes进行容器编排
- 配置负载均衡和自动扩缩容
- 设置数据备份和恢复策略
- 配置SSL/TLS证书
- 设置监控告警

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系方式

- 项目维护者: Taishang Team
- 邮箱: contact@taishang.dev
- 项目地址: https://github.com/taishang/sequence-zero

---

**太上老君序列0** - 融合传统智慧与现代技术的智能系统