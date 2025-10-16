# 权限管理API使用指南

## 概述

太上老君核心服务权限管理模块提供了完整的RBAC（基于角色的访问控制）API接口。本指南详细介绍了如何使用这些API来管理权限、角色和用户权限。

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **Content-Type**: `application/json`

## 认证

### 获取访问令牌

```bash
POST /auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "user_id",
      "username": "username",
      "role": "ADMIN"
    }
  }
}
```

### 使用令牌

在所有API请求的Header中包含：
```
Authorization: Bearer <your_jwt_token>
```

## 权限管理API

### 1. 创建权限

```bash
POST /permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "权限名称",
  "code": "permission_code",
  "description": "权限描述",
  "resource": "资源名称",
  "action": "操作类型"
}
```

**参数说明**:
- `name`: 权限显示名称
- `code`: 权限唯一标识码
- `description`: 权限描述（可选）
- `resource`: 权限控制的资源
- `action`: 允许的操作（如：read, write, delete等）

**响应示例**:
```json
{
  "id": "60610a24-c514-488e-936a-2f4e6850cd35",
  "name": "测试权限",
  "code": "test_permission",
  "description": "这是一个测试权限",
  "resource": "test",
  "action": "read",
  "created_at": "2025-10-14T19:35:26.908+08:00",
  "updated_at": "2025-10-14T19:35:26.908+08:00"
}
```

### 2. 获取权限列表

```bash
GET /permissions?page=1&limit=20
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认20）

**响应示例**:
```json
{
  "permissions": [
    {
      "id": "permission_id",
      "name": "权限名称",
      "code": "permission_code",
      "description": "权限描述",
      "resource": "资源名称",
      "action": "操作类型",
      "created_at": "2025-10-14T19:35:26.908+08:00",
      "updated_at": "2025-10-14T19:35:26.908+08:00"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "pages": 1
}
```

### 3. 获取单个权限

```bash
GET /permissions/{permission_id}
Authorization: Bearer <token>
```

### 4. 更新权限

```bash
PUT /permissions/{permission_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "更新后的权限名称",
  "description": "更新后的描述"
}
```

### 5. 删除权限

```bash
DELETE /permissions/{permission_id}
Authorization: Bearer <token>
```

## 角色管理API

### 1. 创建角色

```bash
POST /roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "角色名称",
  "code": "role_code",
  "description": "角色描述",
  "level": 2
}
```

**参数说明**:
- `name`: 角色显示名称
- `code`: 角色唯一标识码
- `description`: 角色描述（可选）
- `level`: 角色级别（数字，用于权限层级）

**响应示例**:
```json
{
  "id": "62767ff0-e356-4d85-88e7-657f239ab5fc",
  "name": "测试角色",
  "code": "test_role",
  "description": "这是一个测试角色",
  "level": 2,
  "status": "active",
  "permissions": [],
  "created_at": "2025-10-14T19:35:57.706+08:00",
  "updated_at": "2025-10-14T19:35:57.706+08:00"
}
```

### 2. 获取角色列表

```bash
GET /roles?page=1&limit=20
Authorization: Bearer <token>
```

### 3. 获取单个角色

```bash
GET /roles/{role_id}
Authorization: Bearer <token>
```

### 4. 更新角色

```bash
PUT /roles/{role_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "更新后的角色名称",
  "description": "更新后的描述"
}
```

### 5. 删除角色

```bash
DELETE /roles/{role_id}
Authorization: Bearer <token>
```

## 角色权限管理API

### 1. 为角色分配权限

```bash
POST /roles/{role_id}/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "permission_ids": ["permission_id_1", "permission_id_2"]
}
```

**响应示例**:
```json
{
  "message": "Permissions assigned successfully",
  "role_id": "62767ff0-e356-4d85-88e7-657f239ab5fc",
  "permissions": [
    {
      "id": "permission_id",
      "name": "权限名称",
      "code": "permission_code",
      "description": "权限描述",
      "resource": "资源名称",
      "action": "操作类型"
    }
  ]
}
```

### 2. 获取角色权限

```bash
GET /roles/{role_id}/permissions
Authorization: Bearer <token>
```

### 3. 移除角色权限

```bash
DELETE /roles/{role_id}/permissions/{permission_id}
Authorization: Bearer <token>
```

## 用户角色管理API

### 1. 为用户分配角色

```bash
POST /user-roles/{user_id}/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "role_ids": ["role_id_1", "role_id_2"]
}
```

**响应示例**:
```json
{
  "message": "Roles assigned successfully",
  "user_id": "1caed520-8bb5-4d86-b508-64fa0aebacfd",
  "role_ids": ["62767ff0-e356-4d85-88e7-657f239ab5fc"]
}
```

### 2. 获取用户角色

```bash
GET /user-roles/{user_id}/roles
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "user_id": "user_id",
  "roles": [
    {
      "id": "role_id",
      "name": "角色名称",
      "code": "role_code",
      "description": "角色描述",
      "level": 2,
      "status": "active",
      "permissions": [
        {
          "id": "permission_id",
          "name": "权限名称",
          "code": "permission_code",
          "resource": "资源名称",
          "action": "操作类型"
        }
      ]
    }
  ]
}
```

### 3. 移除用户角色

```bash
DELETE /user-roles/{user_id}/roles/{role_id}
Authorization: Bearer <token>
```

## 权限检查API

### 1. 检查单个权限

```bash
POST /permissions/check
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "user_id",
  "resource": "资源名称",
  "action": "操作类型"
}
```

**响应示例**:
```json
{
  "has_permission": true,
  "user_id": "1caed520-8bb5-4d86-b508-64fa0aebacfd",
  "resource": "test",
  "action": "read"
}
```

### 2. 批量检查权限

```bash
POST /permissions/check/batch
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "user_id",
  "permissions": [
    {
      "resource": "资源1",
      "action": "read"
    },
    {
      "resource": "资源2",
      "action": "write"
    }
  ]
}
```

## 错误处理

### 标准错误响应

```json
{
  "error": "错误类型",
  "message": "详细错误信息"
}
```

### 常见错误码

- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未提供有效的认证令牌
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突（如重复的code）
- `500 Internal Server Error`: 服务器内部错误

## 使用示例

### 完整的权限管理流程

```bash
# 1. 登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 2. 创建权限
curl -X POST http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "用户管理",
    "code": "user_management",
    "description": "用户管理权限",
    "resource": "user",
    "action": "manage"
  }'

# 3. 创建角色
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "管理员",
    "code": "admin",
    "description": "系统管理员角色",
    "level": 5
  }'

# 4. 为角色分配权限
curl -X POST http://localhost:8080/api/v1/roles/{role_id}/permissions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": ["{permission_id}"]
  }'

# 5. 为用户分配角色
curl -X POST http://localhost:8080/api/v1/user-roles/{user_id}/roles \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": ["{role_id}"]
  }'

# 6. 检查用户权限
curl -X POST http://localhost:8080/api/v1/permissions/check \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "{user_id}",
    "resource": "user",
    "action": "manage"
  }'
```

## 最佳实践

### 1. 权限设计原则
- 使用清晰的命名约定
- 权限粒度要适中，既不过于细化也不过于粗糙
- 资源和操作要有明确的语义

### 2. 角色设计原则
- 基于业务职能设计角色
- 避免角色权限重叠
- 使用角色级别控制权限层次

### 3. 安全建议
- 定期审查用户权限
- 实施最小权限原则
- 记录权限变更日志
- 使用强密码和定期更换token

### 4. 性能优化
- 对权限检查结果进行缓存
- 批量操作减少API调用次数
- 合理使用分页避免大量数据传输

---

**文档版本**: 1.0  
**最后更新**: 2025年10月14日