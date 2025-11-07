# 前端与终端规划

## 应用划分
- 管理端（Laojun Admin）：插件治理（清单/安装/启停/升级/卸载）、权限与资源、审计与回滚。
- 市场端（Laojun Marketplace）：插件发现、来源与签名、评分与评论、安装向导。
- 数据工作台（Taishang Studio）：模型注册与配置、向量集合与检索、任务编排与监控、数据管道。

## 路由与导航
- 顶级导航：`管理端（Laojun）`、`市场（Laojun）`、`工作台（Taishang）`、`安全与设置`、`观测与告警`。
- 路由前缀：`/admin/laojun/*`、`/market/laojun/*`、`/studio/taishang/*`。
- 统一认证：登录后拉取“用户-角色-权限-菜单”树，跨域共享会话与审计。

## 菜单与权限渲染
- 后端返回：菜单树（绑定路由与动作）、用户可用 `enabledActions` 列表。
- 页面与按钮权限：按角色/资源域授权渲染可见性与操作可用性；敏感动作二次确认。
- 资源域：
  - Laojun：`plugin.install|start|stop|upgrade|uninstall|view`。
  - Taishang：`model.create|update|delete|view`、`vector.collection.create|upsert|search|delete`、`task.create|cancel|view`。

## 页面清单（示例）
- Laojun Admin：
  - 插件列表（筛选、状态、版本、签名/来源）。
  - 安装向导（清单校验、验签、权限与资源评审、可回滚提示）。
  - 插件详情（生命周期操作、日志/指标、UI 组件注册说明）。
  - 审计与回滚（失败阈值告警、traceId、版本回滚）。
- Laojun Marketplace：
  - 市场首页（分类、搜索、评分）。
  - 插件详情（来源可信、清单预览、权限声明）。
  - 安装引导（操作前权限最小化确认）。
- Taishang Studio：
  - 模型管理（注册/配置/上架/下架/更新/回滚）。
  - 向量集合（创建/分片/度量、upsert/search、TTL/索引构建）。
  - 任务编排（创建/队列/运行/完成/失败/取消、优先级/配额/重试）。
  - 数据管道（来源/清洗/切分、嵌入器选择、监控与告警）。

## 状态管理与契约对齐
- 契约客户端：对接 `openapi/laojun.yaml` 与 `openapi/taishang.yaml` 生成客户端；统一 `code/data/message` 包装与错误码提示。
- 长任务：提供进度流与可取消；失败统一错误码与回滚提示。
- 观察性：页面集成指标面板（吞吐/延迟/召回/精度）；traceId 贯通链路。

## 原生终端（补充）
- 端侧路由与容器：插件 UI 组件在端侧容器注册；权限提示与主题隔离。
- 多端一致性：iOS/Android/Harmony/桌面统一交互约束；性能阈值不退化（参见 `docs/testing/strategy.md`）。

## Web 管理后台（React + Ant Design）
- 页面结构：
  - 仪表盘（系统状态、插件概览）
  - 插件管理（列表、安装、启停、升级、卸载）
  - 审计与日志（操作日志、构建/部署日志）
  - 系统配置（环境变量与开关）
- 交互逻辑：
  - 所有操作走确认与回滚提示；接口统一错误提示与重试。
  - 列表分页、搜索、过滤；批量操作。
- 接口对接：OpenAPI 3.0 生成类型；统一 `apiClient`；状态管理 Redux Toolkit/Zustand。

## iOS 原生（Swift/SwiftUI）
- 页面：核心功能轻量化展示与通知；插件状态查看与基础操作（按权限）。
- 对接：统一 `APIClient`；错误与重试一致。

## Android 原生（Kotlin/Jetpack Compose）
- 页面：核心功能轻量化展示与通知；插件状态查看与基础操作（按权限）。
- 对接：统一 `APIClient`；错误与重试一致。

## 机器人端原生（C++/Qt/QML）
- 页面：设备控制与状态；紧急操作入口最小化延迟；离线提示。

## 手表端原生
- Apple Watch：状态与快捷功能；Notifications。
- 微信小程序：极简状态与快捷功能。

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1