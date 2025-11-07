# 仓库结构策略（双域模块化：laojun / taishang）

本策略采用“Domain-Driven Monorepo + Platform Governance”的模式，统一平台层的认证、鉴权、网关与配置，域层按业务边界独立演进；以 OpenAPI 合同为事实源保障跨域耦合最小化与演进可控。

## 顶层目录布局
- `openapi/` 平台与双域的合同源（`laojun.yaml`, `taishang.yaml`）。
- `docs/` 设计与规范文档（前端规划、权限模型、接口标准、CI/CD）。
- `apps/` 前端/终端应用：
  - `laojun-admin-web/` 基础域管理台
  - `laojun-market-web/` 插件市场
  - `taishang-studio-web/` 数据与任务工作室
  - `libs/ui/`, `libs/utils/` 共享组件与工具
- `services/` 后端服务：
  - `platform/` 统一治理：`iam-api`, `gateway`, `config-center`, `discovery`
  - `laojun/` 插件与运行时：`plugin-registry`, `plugin-runtime`, `audit-api`
  - `taishang/` 模型与数据：`model-api`, `vector-api`, `task-orchestrator`
- `plugins/` 第三方/自研插件源与示例（清单、签名、运行脚本）。
- `ops/` 运维与 CI/CD：流水线、环境、监控与告警。
- `libs/` 通用库：`observability`, `contract-client`, `storage-adapters`。
- `tests/` 合同测试与端到端脚本。

## 双域模块化边界
- laojun（基础域）：
  - 职责：插件生命周期（安装/启停/升级/卸载）、清单与签名校验、权限与配额、审计与观测、UI 容器与嵌入。
  - 典型模块：`plugin-manifest`, `plugin-signer`, `runtime-executor`, `audit-log`, `quota-manager`。
- taishang（高级域）：
  - 职责：模型注册与版本、向量集合与索引、任务编排与成本控制、数据流水线与质检、长任务状态与进度。
  - 典型模块：`model-registry`, `embedding-service`, `vector-index`, `task-engine`, `pipeline-etl`, `cost-governor`。

## 平台治理与统一能力
- 认证鉴权：`iam-api` 提供用户、角色、权限、菜单与按钮级授权；域侧只做资源级授权（作用域：全局/工作空间/实例）。
- 网关与路由：`gateway` 统一入口，服务发现与熔断；前端路由前缀：`/laojun/*`、`/taishang/*`、`/platform/*`。
- 配置中心：`config-center` 提供环境与特性开关（Feature Flags）；支持 A/B 与灰度。
- 观测与审计：统一 traceId、指标、日志与审计事件；敏感字段脱敏。

## 合同驱动与版本策略
- 事实源：`openapi/*` 是代码与文档的权威来源；生成客户端与服务端骨架。
- 版本：`semver`；破坏性变更必须在次版本或大版本，并提供兼容层或迁移指引。
- 依赖边界：域内模块依赖走接口层（go: interfaces / ts: libs/api-client），禁止跨域直接调用实现。

## 构建与工作流
- 前端：`pnpm + turborepo` 管理 apps 与 libs；统一 lint、测试与打包。
- 后端：Go 工作空间或多模块；按服务独立编译与部署；共享库放入 `libs/`。
- CI 门禁（详见 `docs/ops/ci-cd-pipeline.md`）：
  - 合同一致性：OpenAPI 校验与客户端更新检查。
  - 原生质量：lint/test/安全扫描/性能基线。
  - 插件生态：清单字段校验、签名验证、运行 sandbox 检查。

## 迁移与演进
- 阶段化路线：
  - Phase 0 基础域 MVP：插件清单/签名/安装与审计闭环；管理台基础路由与菜单。
  - Phase 1 高级域 MVP：模型注册/向量集合/任务编排；Studio 基础数据流。
  - Phase 2 协同增强：工作空间、多终端一致性、跨域流程编排与治理指标。
- 风险控制：
  - 合同漂移：以合同测试与兼容层防护；发布前 diff 审查。
  - 状态重复：避免两域重复持有权威状态；以事件总线或接口同步为准。
  - 配额与成本：平台统一治理，域侧只执行策略。

## 代码约束与命名
- 前端：路由前缀与菜单 key 统一：`laojun.*`、`taishang.*`、`platform.*`。
- 后端：资源名使用复数；错误码与统一包装遵循 `docs/interfaces/standard.md`。
- 插件：`manifest.json` 字段与签名流程遵循 `docs/plugins/development-manual.md`。

## 与文档的映射
- 前端规划：`docs/frontend/planning.md`（应用/路由/菜单权限）。
- 权限模型：`docs/security/permission-model.md`（角色/资源/菜单按钮）。
- 接口标准：`docs/interfaces/standard.md`（统一包装/错误码/安全）。
- CI/CD：`docs/ops/ci-cd-pipeline.md`（门禁与报告）。

> 本文档作为仓库结构与模块边界的事实依据，配合合同与 CI 门禁，保障双域在持续迭代中的一致性与可演进性。