# 启动包完整清单与交付检查（laojun+taishang 原生全栈）

面向 AI 大模型的开发启动包交付清单，确保文档、配置、模板和脚手架齐备，开箱即用。[0]

## 文档类（现有 + 新增）
- 索引：`docs/README.md`
- 架构：`docs/architecture.md`
- 模块设计：`docs/module-design.md`
- 开发规范：`docs/development-guidelines.md`
- 环境搭建：`docs/environment-setup.md`
- 前端规划：`docs/frontend-planning.md`
- 插件体系：`docs/plugin-system.md`
- 插件细化：`docs/plugin-system-details.md`
- 接口规范与错误码：`docs/interfaces-standard.md`
- 数据模型设计：`docs/data-model-design.md`
- 初始化 SQL：`db/schema.sql`（`lao_`/`tai_`）
- 测试策略：`docs/testing-strategy.md`
- 原生多端测试方案：`docs/testing-plan-native.md`
- 安全与合规：`docs/security-compliance.md`
- CI/CD 流水线：`docs/ci-cd-pipeline.md`
- 开发流程：`docs/development-process.md`
- 路线图与交付物：`docs/roadmap-deliverables.md`
- 多终端适配：`docs/multi-terminal-adaptation.md`
- 原生 UI/UX：`docs/native-ui-ux.md`
- 技术栈总览：`docs/tech-stack.md`

## 配置与契约类
- OpenAPI 单一事实源：`openapi/openapi.yaml`（分域路由、示例与错误码）
- Go 模块：`services/api/go.mod` → `github.com/codetaoist/taishanglaojun`
- 环境变量模板：`scripts/.env.example`（后续添加）
- 迁移脚本：`db/migrations/`（后续添加，版本化）

## 代码与模板类（建议后续交付）
- 后端：基础目录与骨架（健康检查、配置加载、日志、DI、仓储适配）
- 前端：`apps/admin-react` Vite 模板、路由与布局、鉴权与接口封装
- 原生端：iOS/Android/Qt/Watch 模板与集成契约（API、离线队列、权限与提示语）
- 插件：脚手架（模板、构建脚本、签名与清单生成、校验与上传）

## CI/CD 与门禁（建议落地）
- GitLab CI：阶段（setup/build/test/scan/package/deploy/e2e）、门禁（覆盖率阈值、秘密扫描、SAST/DAST）
- 发布策略：灰度/蓝绿、版本通道、自动回滚；构建产物签名与验签
- 报告：测试覆盖率、性能压测、变更审计与安全扫描

## 交付检查（Checklist）
- [ ] 文档索引完整、链接可用
- [ ] OpenAPI 路由与错误码一致、示例齐全
- [ ] DB 初始化脚本与前缀策略正确、索引完整
- [ ] Go 模块地址一致、编译通过（基础构建）
- [ ] React 管理后台模板能启动并拉取健康检查
- [ ] 原生端模板可编译启动（最简页面与API契约）
- [ ] 插件脚手架能生成清单与签名，并通过校验
- [ ] CI/CD 能跑通基础阶段、门禁生效

> 参考：[0] https://www.doubao.com/thread/w1745a17f59b91183