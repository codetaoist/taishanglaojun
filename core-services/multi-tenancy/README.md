# 多租户架构模块

## 概述

多租户架构模块为太上老君系统提供企业级的多租户支持，允许多个组织或团队在同一个系统实例中独立运行，同时确保数据隔离和安全性。

## 核心功能

### 1. 租户管理
- 租户创建、更新、删除
- 租户配置管理
- 租户状态监控
- 租户资源配额管理

### 2. 数据隔离
- 基于租户ID的数据分离
- 数据库级别的隔离策略
- 缓存数据隔离
- 文件存储隔离

### 3. 权限控制
- 租户级别的权限管理
- 跨租户访问控制
- 角色权限继承
- 资源访问限制

### 4. 资源管理
- 计算资源配额
- 存储空间限制
- API调用频率限制
- 并发用户数限制

## 技术架构

### 数据隔离策略

1. **行级安全 (Row Level Security)**
   - 在数据库层面实现租户数据隔离
   - 自动添加租户ID过滤条件
   - 确保查询只返回当前租户的数据

2. **Schema隔离**
   - 为每个租户创建独立的数据库Schema
   - 完全的数据隔离
   - 适用于高安全要求的场景

3. **数据库隔离**
   - 为每个租户创建独立的数据库实例
   - 最高级别的数据隔离
   - 适用于超大型企业客户

### 中间件集成

- **租户识别中间件**: 从请求中提取租户信息
- **数据过滤中间件**: 自动添加租户过滤条件
- **权限验证中间件**: 验证跨租户访问权限
- **资源限制中间件**: 实施资源配额限制

## API 接口

### 租户管理

```
POST   /api/v1/tenants              # 创建租户
GET    /api/v1/tenants              # 获取租户列表
GET    /api/v1/tenants/:id          # 获取租户详情
PUT    /api/v1/tenants/:id          # 更新租户信息
DELETE /api/v1/tenants/:id          # 删除租户
```

### 租户配置

```
GET    /api/v1/tenants/:id/config   # 获取租户配置
PUT    /api/v1/tenants/:id/config   # 更新租户配置
```

### 资源管理

```
GET    /api/v1/tenants/:id/quota    # 获取资源配额
PUT    /api/v1/tenants/:id/quota    # 更新资源配额
GET    /api/v1/tenants/:id/usage    # 获取资源使用情况
```

### 用户管理

```
GET    /api/v1/tenants/:id/users    # 获取租户用户列表
POST   /api/v1/tenants/:id/users    # 添加用户到租户
DELETE /api/v1/tenants/:id/users/:user_id  # 从租户移除用户
```

## 配置说明

### 基础配置

```yaml
multi_tenancy:
  enabled: true
  isolation_strategy: "row_level"  # row_level, schema, database
  default_quota:
    max_users: 100
    max_storage_gb: 10
    max_api_calls_per_hour: 10000
    max_concurrent_sessions: 50
  
  # 租户识别策略
  tenant_resolution:
    strategy: "header"  # header, subdomain, path
    header_name: "X-Tenant-ID"
    subdomain_pattern: "{tenant}.example.com"
    path_prefix: "/tenant/{tenant}"
```

### 数据库配置

```yaml
database:
  multi_tenant:
    enable_rls: true  # 启用行级安全
    tenant_column: "tenant_id"
    auto_create_schema: true
    schema_prefix: "tenant_"
```

### 缓存配置

```yaml
cache:
  multi_tenant:
    key_prefix: "tenant:{tenant_id}:"
    isolation_enabled: true
```

## 部署指南

### 1. 数据库准备

```sql
-- 启用行级安全扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建租户表
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 为现有表添加租户ID列
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE posts ADD COLUMN tenant_id UUID REFERENCES tenants(id);
-- ... 其他表
```

### 2. 中间件配置

在主应用中注册多租户中间件：

```go
// 注册多租户中间件
router.Use(multitenancy.TenantResolutionMiddleware())
router.Use(multitenancy.DataIsolationMiddleware())
router.Use(multitenancy.ResourceLimitMiddleware())
```

### 3. 服务集成

```go
// 初始化多租户模块
tenantModule, err := multitenancy.NewModule(config, db, redis, logger)
if err != nil {
    log.Fatal("Failed to initialize multi-tenancy module:", err)
}

// 设置路由
tenantModule.SetupRoutes(router.Group("/api/v1"), jwtMiddleware)
```

## 监控和指标

### 关键指标

- 租户数量和增长趋势
- 每个租户的资源使用情况
- API调用频率和响应时间
- 数据存储使用量
- 并发用户数

### 告警规则

- 租户资源使用超过配额的80%
- API调用频率异常
- 数据库连接数过高
- 跨租户访问尝试

## 安全考虑

### 数据隔离

- 确保所有数据库查询都包含租户过滤条件
- 定期审计跨租户数据访问
- 实施数据备份和恢复的租户隔离

### 权限控制

- 严格的租户边界检查
- 超级管理员权限的谨慎使用
- 定期权限审计和清理

### 监控和审计

- 记录所有租户操作日志
- 监控异常的跨租户访问尝试
- 定期安全评估和渗透测试

## 最佳实践

1. **设计阶段**
   - 从一开始就考虑多租户架构
   - 选择合适的数据隔离策略
   - 设计灵活的权限模型

2. **开发阶段**
   - 使用中间件自动处理租户上下文
   - 在所有数据访问层添加租户过滤
   - 编写全面的单元测试和集成测试

3. **运维阶段**
   - 监控租户资源使用情况
   - 定期备份和测试恢复流程
   - 保持系统性能和可扩展性

## 故障排除

### 常见问题

1. **数据泄露**: 检查是否所有查询都包含租户过滤条件
2. **性能问题**: 优化租户相关的数据库索引
3. **权限错误**: 验证租户权限配置和继承关系
4. **资源超限**: 检查配额设置和使用情况监控

### 调试工具

- 租户数据隔离验证工具
- 权限检查调试接口
- 资源使用情况分析工具
- 性能监控和分析工具