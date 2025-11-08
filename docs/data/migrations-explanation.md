# 数据库迁移说明

本项目使用多个数据库迁移目录，各自有不同的用途和版本控制策略。

## 迁移目录说明

### 1. db/migrations/
- **用途**: 项目整体数据库结构定义
- **文件**: V1__init_lao.sql, V2__init_tai.sql
- **说明**: 
  - V1__init_lao.sql: 初始化老君基础域(lao_)表结构
  - V2__init_tai.sql: 初始化太上域(tai_)表结构
  - 使用Flyway命名约定，版本号前缀V__

### 2. services/api/migrations/
- **用途**: API服务的数据库迁移文件
- **文件**: 20230101000000_create_users_table.up/down.sql等
- **说明**: 
  - 使用时间戳命名约定，支持up/down迁移
  - 包含更详细的表结构变更历史
  - 与API服务紧密集成

### 3. migrations/ (根目录)
- **用途**: 临时迁移文件，建议整合到services/api/migrations/
- **文件**: 001_init.sql, 002_taishang_tables.sql等
- **说明**: 
  - 与services/api/migrations/功能重复
  - 建议迁移完成后删除此目录

## 迁移执行顺序

1. 首次初始化: 执行db/migrations/中的V1和V2文件
2. 后续更新: 使用services/api/migrations/中的时间戳文件

## 建议

- 统一使用services/api/migrations/作为主要迁移目录
- 将db/migrations/中的结构定义作为参考文档
- 考虑将根目录migrations/的内容合并到services/api/migrations/中