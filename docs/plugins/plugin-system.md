# 插件体系文档

## 架构与原则
- 架构：B/S 为主（Web 管理后台操作插件），C/S 为辅（本地开发工具），CI/CD 贯穿插件全生命周期。[0]
- 原则：
  - 清单（manifest）与签名校验，来源可信。
  - 版本语义化，升级路径可控。
  - 安装/启动/升级/卸载流程标准化，统一审计与日志。

## 生命周期
- 开发：本地 C/S 工具创建模板 → 编码 → 本地构建与签名。
- 提交：推送仓库 → 触发 CI/CD 构建与测试。[0]
- 部署：CI/CD 部署至目标环境 → 后端注册与启用。
- 运行：监控与日志 → 升级或卸载。

## 数据落库与表前缀
- 归属域：laojun（老君基础域）
- 核心表：
  - `lao_plugins` 插件主表（唯一索引：`name`）
  - `lao_plugin_versions` 版本表（唯一索引：`plugin_id, version`）
  - `lao_audit_logs` 审计日志（索引：`actor, created_at`）
- 审计策略：安装/启停/升级/卸载均记录审计，含操作人、目标、结果与payload。

## 使用流程（示例）
1. Web 后台调用“安装插件”接口：`POST /api/laojun/plugins/install` （pluginId, version）。
2. 后端校验清单与签名，触发 CI/CD 构建/部署。
3. CI/CD 返回结果，后端记录审计与日志，反馈至后台。[0]
4. 后台可执行启停、升级、卸载操作；后端保证一致性与回滚。

## 接口示例
- 安装：`POST /api/laojun/plugins/install`
- 启动：`POST /api/laojun/plugins/start`
- 停止：`POST /api/laojun/plugins/stop`
- 升级：`POST /api/laojun/plugins/upgrade`
- 卸载：`DELETE /api/laojun/plugins/uninstall`

## CI/CD 步骤（GitLab CI）[0]
- 开发者提交代码 → 触发流水线 → 构建（含插件打包）→ 测试 → 部署 → 通知。
- 插件流水线与后端服务流水线衔接，保证部署一致性与回滚机制一致。

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1