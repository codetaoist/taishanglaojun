# API 服务

这是 TaiShangLaoJun 项目的 API 服务。

## 环境要求

- Go 1.19+
- PostgreSQL 数据库

## 设置步骤

1. 创建数据库：
   ```bash
   psql -U postgres -c "CREATE DATABASE codetaoist;"
   ```

2. 设置环境变量（可选，有默认值）：
   ```bash
   export ENV=dev
   export LOG_LEVEL=info
   export LAOJUN_API_PORT=8081
   export DATABASE_URL=postgres://postgres:password@localhost/codetaoist?sslmode=disable
   export JWT_SECRET=your-secret-key
   export ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
   ```

3. 运行服务：
   ```bash
   go run cmd/api/main.go
   ```

## API 文档

API 文档可以通过以下端点访问：
- 健康检查: GET /health
- Laojun 域 API: /api/v1/laojun/*
- Taishang 域 API: /api/v1/taishang/*

## 开发

在开发模式下，可以通过设置 `DEV_SKIP_SIGNATURE=true` 来跳过 JWT 签名验证。

## 故障排除

如果遇到 "database does not exist" 错误，请确保已创建数据库：
```bash
psql -U postgres -f scripts/init_db.sql
```