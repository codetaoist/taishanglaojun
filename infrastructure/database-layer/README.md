# 数据库连接层 (Database Layer)

太上老君v2项目的数据库连接层模块，提供统一的数据库访问接口和基础操作。

## 概述

数据库层是整个系统的数据访问基础设施，提供以下核心功能：

- **多数据库支持**: PostgreSQL 和 Redis
- **连接管理**: 连接池、健康检查、自动重连
- **基础仓储**: 通用CRUD操作、分页、搜索、过滤
- **数据迁移**: 数据库结构版本管理
- **缓存服务**: Redis缓存和内存缓存备选方案
- **事务支持**: 数据库事务管理
- **日志集成**: 完整的操作日志记录

## 核心特性

### 🗄️ 数据库管理
- PostgreSQL连接池管理
- Redis连接池管理
- 统一的健康检查接口
- 优雅的连接关闭

### 📊 仓储模式
- 泛型基础仓储实现
- 支持CRUD操作
- 分页查询支持
- 动态过滤和搜索
- 批量操作支持

### 🔄 数据迁移
- 基于golang-migrate的迁移管理
- 支持向上/向下迁移
- 版本控制和状态管理
- 自动迁移支持

### 💾 缓存服务
- Redis缓存实现
- 内存缓存备选方案
- 统一的缓存接口
- 多种数据结构支持

## 项目结构

```
database-layer/
├── cmd/
│   └── main.go                 # 主程序入口
├── internal/
│   ├── database/              # 数据库连接管理
│   │   ├── postgres.go        # PostgreSQL连接
│   │   ├── redis.go          # Redis连接
│   │   └── manager.go        # 数据库管理器
│   ├── models/               # 数据模型
│   │   └── base.go          # 基础模型定义
│   ├── repository/           # 仓储层
│   │   └── base.go          # 基础仓储实现
│   └── migrations/           # 数据迁移
│       └── migrate.go       # 迁移管理器
├── configs/
│   └── config.yaml          # 配置文件
├── migrations/              # SQL迁移文件目录
├── go.mod                   # Go模块定义
├── Makefile                # 构建和管理命令
└── README.md               # 项目文档
```

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 设置开发环境

```bash
# 启动数据库服务（使用Docker）
make dev-setup
```

### 3. 运行测试程序

```bash
make run
```

### 4. 开发模式运行

```bash
make dev
```

## 配置说明

### 环境变量配置

```bash
# PostgreSQL配置
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=taishang

# Redis配置
REDIS_HOST=localhost
REDIS_PASSWORD=
```

### 配置文件

配置文件位于 `configs/config.yaml`，支持多环境配置：

```yaml
# PostgreSQL配置
postgres:
  host: localhost
  port: 5432
  username: postgres
  password: password
  database: taishang
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5
  max_lifetime: 300s

# Redis配置
redis:
  host: localhost
  port: 6379
  password: ""
  database: 0
  pool_size: 10
```

## 使用示例

### 数据库管理器

```go
// 创建数据库配置
config := &database.Config{
    Postgres: &database.PostgresConfig{
        Host:     "localhost",
        Port:     5432,
        Username: "postgres",
        Password: "password",
        Database: "taishang",
        SSLMode:  "disable",
    },
    Redis: &database.RedisConfig{
        Host: "localhost",
        Port: 6379,
    },
}

// 初始化管理器
manager, err := database.NewManager(config, logger)
if err != nil {
    log.Fatal(err)
}
defer manager.Close()
```

### 基础仓储使用

```go
// 定义模型
type User struct {
    models.BaseModel
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 创建仓储
repo := repository.NewBaseRepository[User](db, logger)

// 创建用户
user := &User{Name: "张三", Email: "zhangsan@example.com"}
err := repo.Create(ctx, user)

// 查询用户
user, err := repo.GetByID(ctx, 1)

// 分页查询
opts := &models.QueryOptions{
    Pagination: &models.PaginationQuery{
        Page:     1,
        PageSize: 10,
    },
}
result, err := repo.Paginate(ctx, opts)
```

### 缓存服务使用

```go
// 获取缓存服务
cache := manager.GetCacheService()

// 设置缓存
err := cache.Set(ctx, "key", "value", 5*time.Minute)

// 获取缓存
value, err := cache.Get(ctx, "key")

// Redis特定操作
redis := manager.GetRedis()
err := redis.HSet(ctx, "hash", "field", "value")
```

### 数据迁移

```bash
# 创建迁移文件
make migrate-create NAME=create_users_table

# 执行迁移
make migrate-up

# 回滚迁移
make migrate-down

# 查看迁移状态
make migrate-status
```

## API接口

### 数据库管理器

- `NewManager(config, logger)` - 创建数据库管理器
- `GetPostgres()` - 获取PostgreSQL实例
- `GetRedis()` - 获取Redis实例
- `GetCacheService()` - 获取缓存服务
- `Health()` - 健康检查
- `Close()` - 关闭所有连接

### 基础仓储

- `Create(ctx, entity)` - 创建实体
- `GetByID(ctx, id)` - 根据ID获取
- `Update(ctx, entity)` - 更新实体
- `Delete(ctx, id)` - 删除实体
- `List(ctx, opts)` - 列表查询
- `Paginate(ctx, opts)` - 分页查询
- `BatchCreate(ctx, entities)` - 批量创建

### 查询选项

```go
type QueryOptions struct {
    Pagination *PaginationQuery  // 分页参数
    Filters    []FilterQuery     // 过滤条件
    Search     *SearchQuery      // 搜索条件
    Preload    []string          // 预加载关联
    Select     []string          // 选择字段
    Omit       []string          // 忽略字段
}
```

## 开发指南

### 添加新模型

1. 在 `internal/models/` 目录下创建模型文件
2. 继承 `BaseModel` 或实现相关接口
3. 定义表名和验证规则

### 扩展仓储功能

1. 创建特定的仓储接口
2. 继承 `BaseRepository` 并添加自定义方法
3. 实现业务特定的查询逻辑

### 添加数据迁移

```bash
# 创建迁移文件
make migrate-create NAME=your_migration_name

# 编辑生成的SQL文件
# migrations/000001_your_migration_name.up.sql
# migrations/000001_your_migration_name.down.sql
```

## 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行基准测试
make benchmark
```

### 测试数据库

测试使用独立的测试数据库，配置在 `config.yaml` 的 `test` 部分。

## 部署

### Docker部署

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

### 生产环境配置

1. 设置环境变量或配置文件
2. 启用SSL连接
3. 调整连接池参数
4. 配置日志级别

## 性能优化

### 连接池调优

```yaml
postgres:
  max_open_conns: 50    # 最大连接数
  max_idle_conns: 10    # 最大空闲连接数
  max_lifetime: 300s    # 连接最大生存时间

redis:
  pool_size: 20         # 连接池大小
  min_idle_conns: 5     # 最小空闲连接数
```

### 查询优化

- 使用适当的索引
- 避免N+1查询问题
- 使用预加载优化关联查询
- 合理使用分页和过滤

## 监控和日志

### 健康检查

```go
health := manager.Health()
// 返回各数据库的健康状态和统计信息
```

### 日志记录

所有数据库操作都会记录详细的日志，包括：
- 连接状态变化
- 查询执行时间
- 错误信息
- 性能统计

## 故障排除

### 常见问题

1. **连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数
   - 检查网络连接

2. **迁移失败**
   - 检查迁移文件语法
   - 验证数据库权限
   - 查看迁移状态

3. **性能问题**
   - 调整连接池参数
   - 优化查询语句
   - 添加适当索引

### 调试模式

```bash
# 启用调试日志
export LOG_LEVEL=debug
make run
```

## 开发工具

项目提供了丰富的Make命令来简化开发流程：

```bash
make help              # 查看所有可用命令
make dev-setup         # 设置开发环境
make dev               # 开发模式运行
make test              # 运行测试
make lint              # 代码检查
make fmt               # 代码格式化
make clean             # 清理构建文件
```

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 运行测试
5. 创建Pull Request

## 许可证

本项目采用MIT许可证。详见LICENSE文件。

## 📋 主要功能

### 1. 多数据库支持
- **PostgreSQL**：关系型数据存储（用户、权限、配置）
- **MongoDB**：文档型数据存储（文化智慧内容）
- **Redis**：缓存和会话管理
- **Qdrant**：向量数据库（AI语义搜索）

### 2. 数据访问层
- Repository 模式实现
- 数据库连接池管理
- 事务管理
- 数据迁移工具

### 3. 缓存策略
- 多级缓存架构
- 缓存失效策略
- 分布式缓存同步

## 🚀 开发优先级

**P0 - 立即开始**：
- [ ] PostgreSQL 连接和基础CRUD
- [ ] Redis 缓存基础功能
- [ ] 数据库配置管理

**P1 - 第一周完成**：
- [ ] MongoDB 文档操作
- [ ] Repository 模式实现
- [ ] 数据迁移工具

**P2 - 第二周完成**：
- [ ] Qdrant 向量数据库集成
- [ ] 缓存策略优化
- [ ] 性能监控和优化

## 🔧 技术栈

- **Go ORM**：GORM (PostgreSQL)
- **MongoDB Driver**：官方Go驱动
- **Redis Client**：go-redis
- **Vector DB**：Qdrant Go客户端
- **连接池**：pgxpool, MongoDB连接池
- **迁移工具**：golang-migrate

## 📁 目录结构

```
database-layer/
├── postgres/
│   ├── models/               # 数据模型
│   ├── repositories/         # 仓储实现
│   ├── migrations/           # 数据库迁移
│   └── config.go            # 配置管理
├── mongodb/
│   ├── collections/          # 集合定义
│   ├── repositories/         # 文档操作
│   └── indexes/             # 索引定义
├── redis/
│   ├── cache/               # 缓存操作
│   ├── session/             # 会话管理
│   └── pubsub/              # 发布订阅
├── qdrant/
│   ├── vectors/             # 向量操作
│   ├── collections/         # 集合管理
│   └── search/              # 语义搜索
├── shared/
│   ├── interfaces/          # 通用接口
│   ├── types/               # 数据类型
│   └── utils/               # 工具函数
└── tests/
    ├── integration/         # 集成测试
    └── benchmarks/          # 性能测试
```

## 🎯 成功标准

- [ ] 支持所有四种数据库的基础操作
- [ ] Repository模式统一数据访问接口
- [ ] 缓存命中率达到80%以上
- [ ] 数据库连接池稳定运行
- [ ] 支持数据库迁移和版本管理
- [ ] 性能测试通过（QPS > 1000）