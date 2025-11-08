# 数据库与测试文件组织结构优化方案

## 现状分析

### 当前问题
1. **迁移文件分散**：存在三个迁移目录（db/migrations/, migrations/, services/api/migrations/），职责不清
2. **数据库结构定义重复**：schema.sql与迁移文件内容重复且可能不一致
3. **测试文件组织不统一**：集中式测试目录与服务内测试混合

### 冲突点
1. 根目录migrations/与services/api/migrations/功能重复
2. db/schema.sql与迁移文件定义不一致
3. 不同命名约定的迁移文件（时间戳 vs V1/V2）

## 优化方案

### 1. 数据库迁移文件统一

#### 目录结构
```
codetaoist/
├── db/
│   ├── schema.sql              # 保留作为完整数据库结构参考文档
│   └── migrations/             # 保留基础域初始化脚本
│       ├── V1__init_lao.sql    # 老君域基础表结构
│       └── V2__init_tai.sql    # 太上域基础表结构
├── services/
│   ├── api/
│   │   └── migrations/         # 统一迁移目录（时间戳命名）
│   │       ├── 20230101000000_create_users_table.up.sql
│   │       ├── 20230101000000_create_users_table.down.sql
│   │       └── ...
│   └── auth/
│       └── migrations/         # 认证服务专用迁移（如需要）
└── scripts/
    └── db/                     # 数据库管理脚本
        ├── init.sh             # 初始化脚本（执行V1/V2）
        └── migrate.sh          # 迁移脚本（执行时间戳迁移）
```

#### 迁移策略
1. **初始化阶段**：执行db/migrations/中的V1和V2脚本，创建基础域结构
2. **开发阶段**：使用services/api/migrations/中的时间戳迁移文件进行增量更新
3. **参考文档**：保留db/schema.sql作为完整数据库结构参考

### 2. 测试文件组织优化

#### 目录结构
```
codetaoist/
├── tests/                      # 集成测试与端到端测试
│   ├── integration/            # 服务间集成测试
│   ├── e2e/                    # 端到端测试
│   └── contracts/              # 契约测试
└── services/
    ├── api/
    │   ├── internal/
    │   │   └── **/             # 各模块单元测试（与源码同目录）
    │   │       ├── handler_test.go
    │   │       ├── service_test.go
    │   │       └── ...
    │   └── tests/              # API服务特定集成测试
    │       ├── api_test.go
    │       └── ...
    └── auth/
        ├── internal/
        │   └── **/             # 认证服务单元测试
        │       ├── handler_test.go
        │       └── ...
        └── tests/              # 认证服务特定集成测试
            └── auth_test.go
```

#### 测试策略
1. **单元测试**：与源码同目录，便于维护和执行
2. **服务集成测试**：各服务目录下的tests/子目录
3. **系统级测试**：根目录tests/目录下的集成测试和端到端测试

## 实施步骤

### 第一阶段：迁移文件整合
1. 将根目录migrations/中的有效内容合并到services/api/migrations/
2. 删除根目录migrations/目录
3. 更新相关文档和配置文件
4. 修正services/api/migrate.go中的硬编码路径

### 第二阶段：测试文件重组
1. 在各服务内部创建tests/目录
2. 将相关集成测试移动到对应服务目录
3. 为各模块添加单元测试文件

### 第三阶段：脚本与文档更新
1. 创建统一的数据库管理脚本
2. 更新README和相关文档
3. 调整CI/CD流程以适应新结构

## 预期收益

1. **清晰的职责划分**：初始化脚本与增量迁移分离
2. **一致的命名约定**：统一使用时间戳命名迁移文件
3. **更好的可维护性**：测试文件与源码关联更紧密
4. **减少重复**：消除冗余的迁移文件和数据库定义

## 风险与缓解措施

1. **迁移历史丢失风险**：保留所有迁移文件，仅重组目录结构
2. **现有脚本失效**：全面检查并更新所有引用路径的脚本
3. **团队适应成本**：提供详细的迁移指南和文档更新

## 实施进度

### 已完成工作

#### 第一阶段：迁移文件整合
- [x] 分析并合并根目录migrations到services/api/migrations
  - 确认services/api/migrations/20230101000005_create_taishang_tables.up.sql已包含根目录migrations/中的所有表定义
- [x] 删除根目录migrations目录
  - 已删除/Users/lida/Documents/work/codetaoist/migrations目录
- [x] 修正services/api/migrate.go中的硬编码路径
  - 将绝对路径改为相对路径，引用services/api/migrations/20230101000005_create_taishang_tables.up.sql

#### 第二阶段：测试文件重组
- [x] 在各服务内部创建tests/目录结构
  - 已为api、auth、gateway服务创建tests目录
  - 每个服务tests目录包含integration和unit子目录
  - 为每个服务创建README.md说明测试目录结构
- [x] 移动相关集成测试到对应服务目录
  - 已将tests/integration/中的API相关测试文件移动到services/api/tests/integration/
  - 移动的文件包括：api_test.go, conversation_api_test.go, embedding_api_test.go, text_generation_api_test.go

#### 第三阶段：脚本与文档更新
- [x] 创建统一的数据库管理脚本
  - 创建了scripts/db/init.sh用于执行基础域初始化脚本（V1和V2）
  - 创建了scripts/db/migrate.sh用于执行时间戳迁移文件
  - 为脚本添加了执行权限

### 待完成工作

- [ ] 更新README和相关文档，反映新的目录结构
- [ ] 调整CI/CD流程以适应新结构
- [ ] 为各模块添加单元测试文件

## 注意事项

1. 原根目录tests/integration/中的测试文件已移动到对应服务目录，如需保留一份副本，请及时处理
2. 数据库初始化和迁移流程已更新，请使用scripts/db/目录下的新脚本
3. 各服务的测试目录结构已标准化，新增测试请遵循新的目录结构