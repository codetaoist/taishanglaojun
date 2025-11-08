# 数据库迁移文件重构指南

## 概述

本文档描述了如何从旧的迁移文件结构迁移到新的统一结构。

## 迁移文件结构对比

### 旧结构
```
db/
├── schema.sql
├── migrations/
│   ├── V1__init_lao.sql
│   └── V2__init_tai.sql
services/api/migrations/
├── 20230101000000_create_users_table.up.sql
├── 20230101000000_create_users_table.down.sql
├── 20230101000001_create_taishang_domains_table.up.sql
├── 20230101000001_create_taishang_domains_table.down.sql
├── ...
└── 20251108130026_create_model_configs_table.up.sql
services/auth/migrations/
└── create_blacklist_table.sql
```

### 新结构
```
db/
├── schema.sql (更新为schema_new.sql)
├── migrations/
│   ├── V1__init_laojun.sql
│   ├── V2__init_taishang.sql
│   ├── V3__init_conversation.sql
│   └── V4__fix_model_configs.sql
```

## 迁移步骤

### 1. 备份现有数据库
```bash
# 创建数据库备份
pg_dump codetaoist > codetaoist_backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. 应用新的迁移文件
```bash
# 使用Flyway或类似工具应用新迁移
flyway -url=jdbc:postgresql://localhost:5432/codetaoist \
       -user=your_username \
       -password=your_password \
       -locations=filesystem:db/migrations_new \
       migrate
```

### 3. 更新应用程序配置
更新所有应用程序中的数据库迁移路径，指向新的统一迁移目录。

### 4. 验证数据完整性
```sql
-- 检查所有表是否正确创建
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
AND (table_name LIKE 'lao_%' OR table_name LIKE 'tai_%');

-- 检查数据是否正确迁移
SELECT COUNT(*) FROM lao_users;
SELECT COUNT(*) FROM tai_models;
-- ...其他表
```

## 表名前缀规则

新结构采用以下表名前缀规则：

- `lao_` 前缀：老君基础域表，包括用户管理、认证、审计日志等基础功能
  - lao_users
  - lao_sessions
  - lao_token_blacklist
  - lao_configs
  - lao_plugins
  - lao_plugin_versions
  - lao_audit_logs
  - lao_domains

- `tai_` 前缀：太上域表，包括模型管理、向量集合、任务编排等业务功能
  - tai_domains
  - tai_models
  - tai_model_configs
  - tai_vector_collections
  - tai_vectors
  - tai_tasks
  - tai_conversations
  - tai_messages

## 注意事项

1. **多租户支持**：所有表都包含 `tenant_id` 字段，支持多租户隔离
2. **行级安全策略**：所有表都启用了RLS策略，确保数据安全
3. **分区表**：大表（如审计日志、向量数据、任务表）采用分区设计
4. **索引优化**：为常用查询字段创建了索引
5. **触发器**：为自动更新时间戳字段创建了触发器

## 迁移验证清单

- [ ] 所有表已正确创建
- [ ] 所有索引已正确创建
- [ ] 所有RLS策略已正确配置
- [ ] 所有触发器已正确创建
- [ ] 数据已正确迁移
- [ ] 应用程序可以正常连接和操作数据库
- [ ] 备份已成功创建并可恢复

## 回滚计划

如果迁移过程中出现问题，可以按以下步骤回滚：

1. 停止应用程序
2. 从备份恢复数据库
3. 恢复旧的迁移文件结构
4. 重启应用程序

## 联系方式

如有任何问题或疑问，请联系数据库管理员或开发团队。