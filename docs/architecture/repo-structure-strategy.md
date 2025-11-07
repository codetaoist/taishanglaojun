# 仓库结构策略与模块化建议

基于现状与目标，对比单模块、多模块与混合方案，给出落地建议与目录约定。

## 现状概览
- 后端：`services/api`（Go 模块），域内划分 `laojun/taishang`（文档侧清晰）
- 文档：`docs/*` 已分门别类，契约事实源为 `openapi/openapi.yaml`
- 迁移：`db/migrations` 已初始化 `lao_`/`tai_` 两域

## 方案对比
- 单模块（单 Go module + internal 分域）
  - 优点：简单、依赖管理轻、交付快
  - 风险：多语言 SDK/客户端与前端/原生同仓交付节奏受限
  - 适用：早期快速推进与小团队
- 多模块（Monorepo 多子模块/语言工作区）
  - 结构：按后端、SDK、客户端、前端、原生、部署脚本拆分独立模块
  - 优点：版本独立、并行 CI、边界清晰；适配跨语言生态
  - 风险：初期复杂度与维护成本高
  - 适用：插件生态与多语言长期演进
- 混合方案（推荐当前阶段）
  - 做法：后端维持单模块，拆出 SDK/客户端/部署到独立子模块
  - 益处：兼顾简单与演进空间；发布与版本管理更健康

## 推荐目录约定（混合）
- 后端：`services/api`（Go，内含 `internal/laojun`、`internal/taishang`）
- SDK：`sdks/plugin-sdk-go`（laojun 插件 SDK Go）、`sdks/ai-plugin-sdk-python`（taishang AI 插件 SDK Python）
- 客户端：`clients/laojun-go`、`clients/taishang-go`、`clients/taishang-python`
- 前端：`apps/admin-react`
- 原生：`apps/native-ios`、`apps/native-android`
- 部署：`deploy/compose.yaml`、`deploy/Dockerfile`、`deploy/backup/`、`deploy/scripts/`
- OpenAPI：`openapi/laojun.yaml`、`openapi/taishang.yaml`（分域事实源），生成产物各模块复用

## 版本与CI建议
- 语义化版本：后端、SDK、客户端各自维护版本；变更影响矩阵与发布说明
- CI 分层：
  - 合约：OpenAPI 变更触发受影响模块的契约测试与SDK再生成
  - 安全：秘密扫描与依赖风险扫描；插件清单签名校验
  - 测试：Unit/Contract/Integration/E2E 逐层门禁；仅关键路径强制

## 迁移与依赖管理
- 迁移：严格版本化；回滚脚本与审计验证；跨域变更需评审
- 依赖：后端与SDK独立管理；客户端按语言包管理最佳实践（Go modules/Python pip）

## 渐进计划
- 阶段1：维持 `services/api` 单模块 → 补齐接口契约与控制器骨架（代码阶段再执行）
- 阶段2：初始化 `sdks/*` 与 `clients/*` 空模块与README；CI流水线占位
- 阶段3：前端与原生模板入仓；部署脚本完善并与运维手册对齐

## 多模块服务化目录策略（初期开发）
- 目标：以服务矩阵驱动团队并行与独立发布，保持契约与依赖边界清晰，支持平滑扩展。
- 目录规划：
```
services/
  gateway/             # 统一鉴权/限流/路由，入口与跨域治理
  laojun-api/          # 插件治理（安装/启停/升级/卸载/审计/配置）
  taishang-api/        # 模型/向量/任务编排与成本治理
jobs/
  vector-indexer/      # 向量索引构建、数据入库与维护
  task-engine/         # 长任务调度执行（可与 taishang-api 同进程起步）
libs/
  common/              # 日志/配置/中间件/工具集
  contracts/           # OpenAPI 生成的 DTO/SDK（后端/前端共用）
apps/
  admin-react/         # 管理后台（登录/仪表盘/插件/审计/系统配置）
clients/
  ios/ android/ robot-qt/ watch/   # 原生端模板与契约集成
openapi/               # 契约事实源（laojun/taishang）
db/                    # 分域表（lao_/tai_）、迁移与回滚
scripts/               # 校验/签名/契约校验与 diff
```
- 边界与依赖：
  - 禁止跨域直接依赖实现：域间调用通过接口（HTTP/gRPC）或网关转发。
  - 路由前缀：`/api/laojun/*`、`/api/taishang/*`；统一响应包装与错误码映射。
  - 生成与版本：以 `openapi/*.yaml` 为事实源，生成 DTO 与 SDK；语义化版本治理。
- 工作空间与构建：
  - 根使用 `go.work` 管理多模块；各服务独立 `go.mod`；共享库在 `libs/*`。
  - CI：每服务独立 `build/test/scan` 作业；契约 `validate/diff/generate` 与 `schemathesis` 门禁。
- 数据与迁移：
  - 分域前缀 `lao_`/`tai_`；迁移脚本分模块维护；回滚与审计闭环。
- 演进路径：
  - M1：落地服务骨架与统一中间件，契约生成与最小联调。
  - M2：拆分 jobs 与队列，完善鉴权/审计/观测与发布策略。
  - M3：前端管理后台完成核心页面，插件样例与市场流程贯通。