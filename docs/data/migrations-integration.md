# 数据库迁移整合说明

## 概述

本文档说明了根目录migrations/与services/api/migrations/的整合过程，以及整合后的迁移文件结构。

## 整合前状态

### 根目录migrations/文件
- 001_init.sql: 创建lao_users和lao_sessions表
- 002_taishang_tables.sql: 创建tai_models、tai_vector_collections和tai_tasks表
- 003_update_vector_collections.sql: 更新tai_vector_collections表结构
- 004_add_conversation_tables.sql: 创建tai_conversations和tai_messages表
- 005_add_user_id_to_conversations.sql: 为tai_conversations和tai_messages添加user_id字段

### services/api/migrations/文件
- 20230101000000_create_users_table: 创建users表
- 20230101000001_create_taishang_domains_table: 创建taishang_domains表
- 20230101000002_create_laojun_domains_table: 创建laojun_domains表
- 20230101000003_create_sessions_table: 创建sessions表
- 20230101000004_rename_tables_with_prefixes: 重命名表添加前缀
- 20251108130026_create_model_configs_table: 创建model_configs表

## 整合过程

1. **分析重复内容**
   - 根目录001_init.sql与services/api中的users和sessions表创建重复
   - 根目录002-005.sql与services/api中缺少的太上域表结构对应

2. **创建整合迁移文件**
   - 创建20230101000005_create_taishang_tables.up.sql
   - 整合根目录002-005.sql的所有太上域表结构
   - 包含完整的索引定义
   - 创建对应的down.sql回滚脚本

3. **保持一致性**
   - 使用与services/api/migrations/相同的命名约定
   - 保持时间戳顺序
   - 包含up和down迁移文件

## 整合后建议

1. **删除根目录migrations/**
   - 迁移内容已整合到services/api/migrations/
   - 保留db/migrations/作为参考文档
   - 统一使用services/api/migrations/进行数据库版本管理

2. **更新文档**
   - 更新README.md中的目录结构说明
   - 更新数据库迁移文档
   - 添加迁移整合说明文档

3. **执行顺序**
   - 首次初始化: 执行db/migrations/中的V1和V2文件
   - 后续更新: 使用services/api/migrations/中的时间戳文件

## 迁移文件结构

```
services/api/migrations/
├── 20230101000000_create_users_table.up.sql
├── 20230101000000_create_users_table.down.sql
├── 20230101000001_create_taishang_domains_table.up.sql
├── 20230101000001_create_taishang_domains_table.down.sql
├── 20230101000002_create_laojun_domains_table.up.sql
├── 20230101000002_create_laojun_domains_table.down.sql
├── 20230101000003_create_sessions_table.up.sql
├── 20230101000003_create_sessions_table.down.sql
├── 20230101000004_rename_tables_with_prefixes.up.sql
├── 20230101000004_rename_tables_with_prefixes.down.sql
├── 20230101000005_create_taishang_tables.up.sql  # 新增
├── 20230101000005_create_taishang_tables.down.sql  # 新增
├── 20251108130026_create_model_configs_table.up.sql
└── 20251108130026_create_model_configs_table.down.sql
```

## 总结

通过整合根目录migrations/到services/api/migrations/，我们实现了：
1. 统一的迁移管理
2. 清晰的版本控制
3. 完整的回滚支持
4. 简化的项目结构

建议在完成整合后删除根目录migrations/目录，以避免混淆和维护负担。