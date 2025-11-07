# 数据字典（lao_ 与 tai_ 前缀表）

本数据字典详列各表字段、类型、约束与索引，作为后端实现与数据治理参考。[0]

## lao_plugins
- id: VARCHAR(64), PK
- name: VARCHAR(128), UNIQUE
- description: TEXT
- status: VARCHAR(32), NOT NULL, DEFAULT 'inactive'
- created_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP
- updated_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`name` 唯一

## lao_plugin_versions
- id: SERIAL, PK
- plugin_id: VARCHAR(64), NOT NULL, FK → lao_plugins(id)
- version: VARCHAR(64), NOT NULL
- manifest: JSONB
- signature: TEXT
- created_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`plugin_id, version` 唯一

## lao_audit_logs
- id: SERIAL, PK
- actor: VARCHAR(64), NOT NULL
- action: VARCHAR(64), NOT NULL
- target: VARCHAR(64)
- payload: JSONB
- result: VARCHAR(32), NOT NULL
- created_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`actor, created_at` 复合；可分区或归档

## lao_configs
- key: VARCHAR(128), PK/UNIQUE
- value: TEXT/JSONB
- scope: VARCHAR(64)
- updated_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`scope` 普通索引

## tai_models
- id: VARCHAR(64), PK
- name: VARCHAR(128), NOT NULL
- provider: VARCHAR(64), NOT NULL
- version: VARCHAR(64)
- status: VARCHAR(32), NOT NULL, DEFAULT 'inactive'
- created_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`name, provider` 唯一

## tai_vector_collections
- id: SERIAL, PK
- name: VARCHAR(128), NOT NULL
- dims: INT, NOT NULL
- index_type: VARCHAR(64)
- created_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`name` 唯一

## tai_tasks
- id: SERIAL, PK
- type: VARCHAR(64), NOT NULL
- status: VARCHAR(32), NOT NULL
- payload: JSONB
- result: JSONB
- created_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP
- updated_at: TIMESTAMP, NOT NULL, DEFAULT CURRENT_TIMESTAMP

索引：`status, created_at` 复合；必要时分表

> 参考：[0] https://www.doubao.com/thread/w1745a17f59b91183