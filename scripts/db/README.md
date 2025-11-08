# 数据库迁移管理

本目录包含统一数据库迁移脚本 `migrate_unified.sh`，用于管理项目中的数据库迁移。

## 功能

- **统一迁移管理**：支持所有迁移文件的统一管理
- **多种操作模式**：支持应用迁移、查看状态、验证迁移等
- **错误修复**：自动修复常见的迁移问题
- **表结构分析**：分析数据库表结构和前缀使用情况
- **灵活配置**：支持通过环境变量自定义数据库连接

## 使用方法

### 基本用法

```bash
# 应用所有待执行的迁移
./scripts/db/migrate_unified.sh up

# 查看迁移状态
./scripts/db/migrate_unified.sh status

# 验证迁移文件
./scripts/db/migrate_unified.sh validate

# 修复迁移问题
./scripts/db/migrate_unified.sh fix

# 分析数据库表结构
./scripts/db/migrate_unified.sh analyze

# 显示帮助信息
./scripts/db/migrate_unified.sh help
```

### 高级选项

```bash
# 预览模式（不实际执行）
./scripts/db/migrate_unified.sh up --dry-run

# 强制执行（跳过确认）
./scripts/db/migrate_unified.sh clean --force
```

### 环境变量配置

可以通过以下环境变量自定义数据库连接：

```bash
export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=postgres
export DB_PASS=password
export DB_NAME=taishanglaojun
export CONTAINER_NAME=taishanglaojun-postgres
```

## 数据库表前缀

项目使用以下表前缀：

- `lao_`：老君系统相关表
- `tai_`：太上系统相关表

### 无前缀表

数据库中有两个无前缀表：

1. `schema_migrations`：迁移记录表，用于跟踪已应用的迁移
2. `users`：用户表，可能是历史遗留表

## 迁移文件

迁移文件位于 `db/migrations/` 目录下，使用以下命名约定：

- `V1__init_laojun.sql`：初始化老君系统
- `V2__init_taishang.sql`：初始化太上系统
- `V3__init_conversation.sql`：初始化对话系统
- `V4__fix_model_configs.sql`：修复模型配置

## 常见问题

### Vector扩展问题

如果遇到vector扩展问题，可以使用修复命令：

```bash
./scripts/db/migrate_unified.sh fix
```

该命令会：
1. 检查vector扩展是否安装
2. 如果未安装，尝试安装或创建不依赖vector的表结构
3. 修复缺少的tenant_id列
4. 标记迁移为已完成

### 迁移状态不一致

如果发现迁移状态不一致，可以使用验证命令：

```bash
./scripts/db/migrate_unified.sh validate
```

该命令会检查迁移文件的校验和，确保迁移文件未被修改。

## 脚本说明

本目录只包含一个主要脚本：

| 脚本 | 功能 | 使用方式 | 推荐场景 |
|------|------|----------|----------|
| `migrate_unified.sh` | 统一迁移管理 | 直接执行 | 所有迁移操作 |

## 最佳实践

1. 使用 `migrate_unified.sh` 作为主要的迁移工具
2. 在应用迁移前，使用 `--dry-run` 选项预览更改
3. 定期使用 `validate` 命令检查迁移状态
4. 在生产环境中，谨慎使用 `clean` 命令
5. 保持迁移文件的命名一致性，使用 `V` 前缀和双下划线分隔符

## 故障排除

### 连接问题

如果遇到数据库连接问题，请检查：
1. Docker容器是否运行
2. 数据库连接参数是否正确
3. 数据库是否已创建

### 迁移失败

如果迁移失败，请检查：
1. 迁移文件中的SQL语法是否正确
2. 是否有依赖的扩展未安装
3. 是否有权限问题

可以使用 `fix` 命令尝试自动修复常见问题。