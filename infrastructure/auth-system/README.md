# Auth System

JWT认证和用户管理系统 - 一个完整的Go语言认证服务

## 功能特性

### 🔐 认证功能
- 用户注册和登录
- JWT访问令牌和刷新令牌
- 密码重置和邮箱验证
- 会话管理和令牌撤销
- 多设备登录支持

### 👥 用户管理
- 用户资料管理
- 角色和权限系统
- 用户状态管理（激活/禁用/锁定）
- 批量用户操作

### 🛡️ 安全特性
- 密码加密存储（bcrypt）
- 登录失败锁定机制
- 速率限制
- CORS和安全头设置
- 令牌黑名单机制

### 📊 监控和日志
- 结构化日志记录
- 请求追踪
- 性能监控
- 健康检查端点

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL + Redis
- **ORM**: GORM
- **认证**: JWT (golang-jwt)
- **日志**: Zap
- **配置**: Viper + 环境变量
- **容器**: Docker + Docker Compose

## 项目结构

```
auth-system/
├── cmd/                    # 应用入口
│   └── main.go
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── handler/           # HTTP处理器
│   ├── jwt/               # JWT管理
│   ├── logger/            # 日志管理
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   ├── repository/        # 数据访问层
│   ├── routes/            # 路由配置
│   └── service/           # 业务逻辑层
├── configs/               # 配置文件
├── docs/                  # API文档
├── scripts/               # 脚本文件
├── tests/                 # 测试文件
├── .env.example           # 环境变量示例
├── docker-compose.yml     # Docker编排
├── Dockerfile             # Docker镜像
├── Makefile              # 构建脚本
└── README.md             # 项目文档
```

## 快速开始

### 1. 环境准备

确保已安装以下软件：
- Go 1.21+
- PostgreSQL 12+
- Redis 6+
- Docker & Docker Compose（可选）

### 2. 克隆项目

```bash
git clone <repository-url>
cd auth-system
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，设置数据库连接等配置
```

### 4. 安装依赖

```bash
make deps
```

### 5. 启动数据库（使用Docker）

```bash
make db-up
```

### 6. 运行应用

```bash
# 开发模式（热重载）
make dev

# 或者直接运行
make run
```

### 7. 测试API

访问 http://localhost:8080/api/v1/health 检查服务状态

## API文档

### 认证端点

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "first_name": "Test",
  "last_name": "User"
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

#### 刷新令牌
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}
```

#### 获取用户信息
```http
GET /api/v1/user/me
Authorization: Bearer your-access-token
```

### 管理员端点

#### 用户列表
```http
GET /api/v1/admin/users
Authorization: Bearer admin-access-token
```

#### 系统统计
```http
GET /api/v1/admin/stats/users
Authorization: Bearer admin-access-token
```

## 开发指南

### 代码规范

```bash
# 格式化代码
make fmt

# 代码检查
make lint

# 运行测试
make test

# 查看测试覆盖率
make coverage
```

### 构建和部署

```bash
# 构建应用
make build

# 构建所有平台版本
make build-all

# 构建Docker镜像
make docker-build

# 运行Docker容器
make docker-run
```

### 数据库管理

```bash
# 启动数据库
make db-up

# 停止数据库
make db-down

# 重置数据库
make db-reset
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `SERVER_HOST` | 服务器主机 | `localhost` |
| `SERVER_PORT` | 服务器端口 | `8080` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `JWT_SECRET_KEY` | JWT密钥 | 必须设置 |
| `JWT_ACCESS_TOKEN_TTL` | 访问令牌有效期 | `15m` |
| `JWT_REFRESH_TOKEN_TTL` | 刷新令牌有效期 | `168h` |

### 角色和权限

系统预定义了三种角色：

- **User**: 普通用户，只能访问自己的资源
- **Admin**: 管理员，可以管理用户和查看统计信息
- **SuperAdmin**: 超级管理员，拥有所有权限

## 监控和日志

### 健康检查

```http
GET /api/v1/health
```

### 日志配置

支持多种日志输出格式：
- JSON格式（生产环境推荐）
- 控制台格式（开发环境）
- 文件输出（支持日志轮转）

### 监控集成

支持集成以下监控工具：
- Prometheus（指标收集）
- Grafana（可视化仪表板）
- Jaeger（链路追踪）

## 安全最佳实践

1. **密码安全**
   - 使用bcrypt加密存储
   - 强制密码复杂度要求
   - 支持密码重置功能

2. **令牌安全**
   - JWT令牌短期有效
   - 支持令牌撤销和黑名单
   - 刷新令牌轮转机制

3. **访问控制**
   - 基于角色的权限控制
   - API端点权限验证
   - 请求速率限制

4. **数据保护**
   - 敏感数据加密存储
   - SQL注入防护
   - XSS和CSRF防护

## 测试

```bash
# 运行所有测试
make test

# 运行基准测试
make bench

# 安全检查
make security

# 完整CI流程
make ci
```

## 部署

### Docker部署

```bash
# 构建并启动所有服务
docker-compose up -d

# 仅启动应用和数据库
docker-compose up -d auth-system postgres redis

# 生产环境部署（包含Nginx）
docker-compose --profile production up -d
```

### 生产环境配置

1. 修改JWT密钥
2. 配置HTTPS证书
3. 设置数据库连接池
4. 配置日志轮转
5. 启用监控和告警

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数是否正确
   - 确认网络连通性

2. **JWT令牌验证失败**
   - 检查JWT密钥配置
   - 验证令牌是否过期
   - 确认令牌格式正确

3. **权限验证失败**
   - 检查用户角色设置
   - 验证权限配置
   - 确认中间件配置

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交代码变更
4. 运行测试确保通过
5. 提交Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交Issue或联系维护者。

## 📋 主要功能

### 1. 用户认证
- 用户注册/登录
- JWT Token 生成和验证
- 刷新Token机制
- 多设备登录管理

### 2. 权限管理
- L1-L9 分层权限系统
- 角色权限映射
- 资源访问控制
- 权限继承机制

### 3. 安全特性
- 密码加密存储
- 登录失败限制
- 会话管理
- 安全审计日志

## 🚀 开发优先级

**P0 - 立即开始**：
- [ ] JWT Token 生成和验证
- [ ] 用户注册/登录基础功能
- [ ] 密码加密和验证

**P1 - 第一周完成**：
- [ ] L1-L9 权限系统设计
- [ ] 角色权限管理
- [ ] 中间件集成

**P2 - 第二周完成**：
- [ ] 多设备登录管理
- [ ] 安全审计功能
- [ ] 权限缓存优化

## 🔧 技术栈

- **JWT库**：golang-jwt/jwt
- **密码加密**：bcrypt
- **权限模型**：Casbin
- **中间件**：Gin middleware
- **缓存**：Redis (权限缓存)

## 📁 目录结构

```
auth-system/
├── jwt/
│   ├── generator.go          # Token生成
│   ├── validator.go          # Token验证
│   └── refresh.go           # Token刷新
├── permissions/
│   ├── levels.go            # L1-L9权限定义
│   ├── rbac.go              # 角色权限管理
│   └── middleware.go        # 权限中间件
├── users/
│   ├── models.go            # 用户模型
│   ├── service.go           # 用户服务
│   └── handlers.go          # HTTP处理器
├── security/
│   ├── password.go          # 密码处理
│   ├── session.go           # 会话管理
│   └── audit.go             # 安全审计
├── config/
│   ├── jwt_config.go        # JWT配置
│   └── security_config.go   # 安全配置
└── tests/
    ├── auth_test.go         # 认证测试
    └── permission_test.go   # 权限测试
```

## 🎯 L1-L9 权限体系

```yaml
权限层级:
  L1: 游客访问 - 基础浏览权限
  L2: 注册用户 - 基本交互权限
  L3: 认证用户 - 内容创建权限
  L4: 进阶用户 - 高级功能权限
  L5: 专业用户 - 专业工具权限
  L6: 管理用户 - 用户管理权限
  L7: 系统管理 - 系统配置权限
  L8: 核心管理 - 核心功能权限
  L9: 至高权限 - 全系统控制权限
```

## 🎯 成功标准

- [ ] JWT认证系统正常工作
- [ ] L1-L9权限系统完整实现
- [ ] 用户注册登录流程无误
- [ ] 权限验证中间件正常工作
- [ ] 安全测试通过
- [ ] 性能测试达标（认证QPS > 500）