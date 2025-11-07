# CodeTaoist 项目启动包（Monorepo）

## 文档元信息
- 项目中文名称：太上老君
- 作者：Lida
- 域名：codetaoist.com
- 时间：2025.11
- Github： https://github.com/codetaoist/taishanglaojun

# 文档索引与导航

该项目文档已按主题分目录组织，覆盖宏观到微观的全链路设计、流程、数据、接口、安全、后端、前端、原生终端与运维交付。

## 顶层文档
- `docs/development-process.md`：分阶段开发流程与门禁
- `docs/architecture.md`：架构总览（项目级）
- `docs/tech-stack.md`：技术栈总览
- `docs/startup-package-checklist.md`：启动包清单与交付检查
- `openapi/laojun.yaml`、`openapi/taishang.yaml`：OpenAPI 3.0 接口事实源

## 架构与仓库
- `docs/architecture/repo-structure-strategy.md`：仓库结构策略与模块化建议（新增）
- `docs/architecture/repo-structure.md`：仓库结构策略（事实依据，已落地）

## 分域与数据
- `docs/domains/laojun.md`：老君域（插件/审计/配置）
- `docs/domains/taishang.md`：太上域（模型/向量/任务）
- `docs/data/data-model-design.md`：数据模型设计
- `docs/data/data-dictionary.md`：数据字典（`lao_`/`tai_` 表）
- `docs/data/migrations-plan.md`：迁移版本化与回滚策略
- 迁移文件：
  - `db/migrations/V1__init_lao.sql`
  - `db/migrations/V2__init_tai.sql`

## 接口契约
- `docs/interfaces/standard.md`：接口规范与错误码
- `docs/interfaces/openapi-design-guide.md`：OpenAPI 设计指南（分域路由/版本/生成）
- `docs/interfaces/laojun-api-spec.md`：老君接口规范文档
- `docs/interfaces/taishang-api-spec.md`：太上接口规范文档
- `openapi/laojun.skeleton.yaml`、`openapi/taishang.skeleton.yaml`：分域骨架示例（统一包装/安全/幂等/分页）

## 安全与权限
- `docs/security/permission-model.md`：角色/资源/动作的权限模型与审计
- `docs/security/threat-model.md`：STRIDE 威胁分析与防护策略
- `docs/security/compliance.md`：安全与合规策略

## 后端（Go + Gin）
- `docs/backend/architecture.md`：分层架构、目录约定、事务一致性
- `docs/backend/gin-router-middleware.md`：路由与中间件（鉴权/审计/契约）细节
- `docs/backend/module-design.md`：模块设计说明书

## 前端（React Admin）
- `docs/frontend/planning.md`：前端规划
- `docs/frontend/ia-and-contracts.md`：信息架构、页面契约与权限可见性

## 原生终端
- `docs/native/architecture-and-adaptation.md`：iOS/Android/机器人/手表架构与跨端适配
- `docs/native/multi-terminal-adaptation.md`：多终端适配细节
- `docs/native/ui-ux.md`：原生 UI/UX 设计规范

## 插件生态
- `docs/plugins/README.md`：插件文档索引
- `docs/plugins/plugin-system.md`：插件体系文档
- `docs/plugins/plugin-system-details.md`：插件体系细化设计
- `docs/plugins/plugin-manifest.md`：插件 manifest 与签名校验
- `docs/plugins/development-manual.md`：插件开发手册（Go/Docker/AI）

## 测试工程
- `docs/testing/README.md`：测试文档索引
- `docs/testing/strategy.md`：测试策略
- `docs/testing/plan-native.md`：原生多端测试方案
- `docs/testing/test-cases.md`：测试用例集

## 运维交付
- `docs/ops/ci-cd-pipeline.md`：CI/CD 分阶段与示例配置
- `docs/ops/deployment-baseline.md`：部署基线
- `docs/ops/environment-setup.md`：环境搭建
- `docs/ops/deployment-observability-performance.md`：部署策略、可观测性与性能规划
- `docs/ops/deployment-operations-manual.md`：部署运维手册

## AI 助手
- `docs/ai/llm-development-guide.md`：AI 大模型辅助开发指引（新增）

## 导航建议
- 按“分域 → 接口 → 数据 → 安全 → 端侧/前端 → 插件 → 测试 → 运维”的阅读顺序，配合 `openapi/laojun.yaml` 与 `openapi/taishang.yaml` 作为契约事实源。
- 变更需同步更新对应目录文档并在 CI 通过契约/安全/覆盖率门禁。

## 文档索引（关键主题）
- 接口契约与规范：`openapi/laojun.yaml`，`openapi/taishang.yaml`，`docs/interfaces/openapi-design-guide.md`，`docs/interfaces/laojun-api-spec.md`，`docs/interfaces/taishang-api-spec.md`
- 开发流程与路线：`docs/development-process.md`，`docs/development-guidelines.md`，`docs/roadmap-deliverables.md`
- 插件体系：`docs/plugins/plugin-system-details.md`，`docs/plugins/plugin-manifest.md`，`docs/plugins/development-manual.md`
- 原生适配与 UI/UX：`docs/native/multi-terminal-adaptation.md`，`docs/native/ui-ux.md`，`docs/native/architecture-and-adaptation.md`
- 测试策略与用例：`docs/testing/strategy.md`，`docs/testing/plan-native.md`，`docs/testing/test-cases.md`
- 启动包与交付项：`docs/release/startup-package.md`，`docs/startup-package-checklist.md`
- 安全与合规：`docs/security/permission-model.md`，`docs/security/compliance.md`，`docs/security/threat-model.md`
- 架构与模块设计：`docs/architecture.md`，`docs/backend/architecture.md`，`docs/backend/gin-router-middleware.md`，`docs/backend/module-design.md`，`docs/architecture/repo-structure-strategy.md`，`docs/frontend/ia-and-contracts.md`
- 运维与流水线：`docs/ops/ci-cd-pipeline.md`，`docs/ops/deployment-baseline.md`，`docs/ops/environment-setup.md`，`docs/ops/deployment-observability-performance.md`，`docs/ops/deployment-operations-manual.md`
- 数据模型与迁移：`docs/data/data-model-design.md`，`docs/data/data-dictionary.md`，`docs/data/migrations-plan.md`
- AI 开发指引：`docs/ai/llm-development-guide.md`

## 工具与脚本
- `scripts/README.md`：脚本使用说明（manifest 校验/签名、OpenAPI 校验与合同 diff）
- `scripts/requirements.txt`：Python 依赖
- `scripts/manifest_validate.py`：清单校验（laojun）
- `scripts/manifest_sign.py`：最小签名示例（HMAC-SHA256）
- `scripts/openapi_validate.py`：OpenAPI 校验（支持完整规范）
- `scripts/openapi_contract_diff.py`：OpenAPI 合同差异对比


## 文档写作规范与模板
- 统一结构：
  - 标题（中文/英文）
  - 概述（Purpose/Scope/Audience/Status）
  - 相关文档（Links）
  - 约定（响应包装/错误码/鉴权/分页/版本）
  - 正文（分章节：背景→设计→流程→接口→数据→安全→测试→运维）
  - 变更记录（日期/变更内容/作者）
- 接口示例：统一使用 `curl`，包含请求与响应示例；字段与枚举引用 OpenAPI。
- 错误码：引用 `docs/interfaces/standard.md` 与 `openapi/*.yaml` 的枚举；在各接口文档中附示例错误响应。
- 版本与审查：新增/变更文档需评审；与 CI 门禁（契约/安全/覆盖率）联动。

示例模板（复制到新文档开头）：
```
# 文档标题（Module/Topic）

概述
- Purpose：一句话说明目的
- Scope：边界与包含
- Audience：阅读对象
- Status：草案/稳定/已发布

相关文档
- docs/interfaces/standard.md
- openapi/laojun.yaml

约定
- 响应：{ code, message, data, traceId }
- 错误码：OK | INVALID_ARGUMENT | ...
- 鉴权：bearerAuth（JWT）
- 分页：page/pageSize；排序：sort=name:asc

（正文章节）
```

## 多模块服务化规划（初期）
- 服务矩阵：
  - 后端服务：`gateway`（统一鉴权/限流/路由）、`laojun-api`（插件治理）、`taishang-api`（模型/向量/任务）
  - 作业与任务：`jobs/vector-indexer`（索引构建）、`jobs/task-engine`（编排执行，可与 `taishang-api` 同进程起步）
  - 前端与原生：`apps/admin-react`、`clients/ios`、`clients/android`、`clients/robot-qt`、`clients/watch`
  - 共用库：`libs/common`（日志/配置/中间件）、`libs/contracts`（OpenAPI 生成 DTO/SDK）
  - 契约与测试：`openapi/*`（事实源）、`tests/contracts/*`（契约测试）
- 目录与边界：
  - 使用 `go.work` 管理多模块，服务独立 `go.mod`；禁止跨域直接依赖实现，统一经网关或接口调用。
  - 路由前缀分域：`/api/laojun/*` 与 `/api/taishang/*`；统一响应包装 `{ code, message, data, traceId }` 与错误码映射。
  - 数据分域：`lao_` 与 `tai_` 表前缀；迁移版本化与回滚（`db/migrations`）。
- CI/CD：
  - 每服务 `build/test/scan` 独立作业；契约 `validate/diff/generate` 与 `schemathesis` 合同测试门禁。
  - 产物与版本：语义化版本、构建产物签名与验签；报告纳入审查（覆盖率/安全/契约）。
- 部署：
  - 本地 `docker-compose`（DB/Redis/向量库/各服务）；生产 K8s（网关/服务/作业与队列）；灰度与观测（指标/日志/审计）。
- 里程碑：
  - M1：落地服务骨架、契约生成、最小端到端联调与契约测试。
  - M2：拆出 `gateway` 与 `jobs`，完善鉴权与审计、中间件与错误码映射。
  - M3：管理后台页面、插件样例与门户、合同测试门禁与发布策略。