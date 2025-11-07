# 权限模型（平台级 IAM 与域级资源）

## 目标
- 提供平台级统一的用户/角色/权限/菜单治理；域内仅定义资源动作，由平台判定授权。
- 支撑 Laojun（插件治理）与 Taishang（模型/向量/任务编排）两域的细粒度控制。

## 角色建议
- SuperAdmin：全局策略、租户与密钥管理、全域审计。
- OpsAdmin：平台运维与监控、限流与回滚。
- PluginAdmin：插件安装/启停/升级/卸载、清单与签名校验。
- ModelAdmin：模型注册/配置/上架/下架/更新/回滚、路由策略。
- DataEngineer：向量集合创建/维护、upsert/search、索引构建策略、任务编排。
- Analyst：数据检索与任务查看、分析报表；只读权限。

## 权限域与资源动作
- 平台级（跨域通用）：`user.*`、`role.*`、`permission.*`、`menu.*`、`workspace.*`、`tenant.*`、`secrets.*`、`policy.*`。
- Laojun（插件域）：`plugin.install|start|stop|upgrade|uninstall|view`、`plugin.audit.view`、`plugin.ui.register`。
- Taishang（治理域）：
  - 模型：`model.create|update|delete|view`、`model.route.manage`。
  - 向量：`vector.collection.create|upsert|search|delete|view`、`vector.index.manage`。
  - 任务：`task.create|cancel|view`、`task.policy.manage`。

## 资源作用域
- 全局：平台与运维策略；密钥与合规。
- 工作空间：项目维度的数据/模型/向量/任务。
- 插件实例：实例级操作与审计轨迹。

## 菜单与动作映射（示例）
- 菜单项绑定路由与动作集合；后端返回用户可见菜单与 `enabledActions`，前端按按钮级控制：
- 示例：
  - `菜单项：插件列表` → 路由 `/admin/laojun/plugins` → 动作 `plugin.view`。
  - `菜单项：插件安装` → 路由 `/admin/laojun/install` → 动作 `plugin.install`。
  - `菜单项：模型管理` → 路由 `/studio/taishang/models` → 动作 `model.create|update|delete|view`。
  - `菜单项：向量检索` → 路由 `/studio/taishang/vectors/search` → 动作 `vector.search`。
  - `菜单项：任务编排` → 路由 `/studio/taishang/tasks` → 动作 `task.create|cancel|view`。

## 授权判定（建议）
- RBAC 为主，ABAC 可选：在域内动作判定时附加上下文属性（租户/工作空间/标签）。
- 最小化授权：默认只读，敏感动作需提升与二次确认；所有敏感动作写审计。
- 审计与回滚：traceId 与动作记录落库；失败阈值触发告警与回滚策略。

## 接口与契约（与前端衔接）
- 平台 IAM：`/api/auth/*` 登录与会话；`/api/admin/users|roles|permissions|menus` 管理接口。
- 前端获取：登录后拉取菜单树与 `enabledActions`；会话中包含 `tenant_id`/`workspace_id` 注入。
- 契约安全：所有受保护端点在 OpenAPI 中声明 `securitySchemes` 与 `security`；统一 `code/data/message` 响应包装与错误码映射。

> 参见：`docs/frontend/planning.md`、`openapi/laojun.yaml`、`openapi/taishang.yaml`、`docs/interfaces/standard.md`