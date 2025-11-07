# Auth Service

认证服务提供用户认证、授权和会话管理功能。

## 功能特性

- 用户注册和登录
- JWT令牌认证
- 密码加密存储
- 会话管理
- 权限控制（基于角色的访问控制）
- 密码修改
- 令牌刷新

## API端点

### 公开端点（无需认证）

- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/refresh` - 刷新令牌

### 受保护端点（需要认证）

- `GET /api/v1/profile` - 获取用户资料
- `POST /api/v1/change-password` - 修改密码
- `POST /api/v1/logout` - 用户登出

### 管理员端点（需要管理员权限）

- `GET /api/v1/admin/users/:id` - 获取指定用户信息

## 运行服务

1. 安装依赖：
```bash
go mod tidy
```

2. 配置环境变量：
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/taishanglaojun?sslmode=disable"
export JWT_SECRET="your-secret-key-change-in-production"
```

3. 运行服务：
```bash
go run cmd/auth/main.go
```

## 配置

服务可以通过环境变量或配置文件进行配置。配置文件为`config.yaml`，包含以下选项：

- `port`: 服务端口（默认: 8081）
- `environment`: 运行环境（development/production）
- `log_level`: 日志级别（debug/info/warn/error）
- `database_url`: 数据库连接字符串
- `jwt_secret`: JWT签名密钥
- `jwt_expiration`: JWT令牌过期时间（秒）
- `allowed_origins`: CORS允许的源

## 数据库

服务使用PostgreSQL数据库，需要确保以下表存在：

- `lao_users`: 用户表
- `lao_sessions`: 会话表

表结构请参考迁移文件：`/Users/lida/Documents/work/codetaoist/services/api/migrations/20230101000004_rename_tables_with_prefixes.up.sql`

## 认证流程

1. 用户使用用户名和密码登录
2. 服务验证凭据并生成JWT令牌
3. 客户端在后续请求中使用`Authorization: Bearer <token>`头
4. 服务验证令牌并提取用户信息

## 错误响应

所有API端点返回标准化的错误响应格式：

```json
{
  "code": "ERROR_CODE",
  "message": "Human readable error message",
  "details": "Detailed error information"
}
```