# 插件体系细节（扩充版）

本设计细化插件从元数据到工具链，再到后端接口与前端集成的全流程，覆盖 laojun 基础域的落库策略，保证来源可信与生命周期可控。[0]

## 元数据与清单（manifest）
- 字段建议：
  - `id`, `name`, `version`, `entry`, `permissions`, `dependencies`, `checksum`, `signature`, `metadata`, `compatibility`（后端版本/接口版本）。
- JSON Schema（简版）：
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["id", "name", "version", "entry", "checksum"],
  "properties": {
    "id": {"type": "string", "maxLength": 64},
    "name": {"type": "string", "maxLength": 128},
    "version": {"type": "string"},
    "entry": {"type": "string"},
    "permissions": {"type": "array", "items": {"type": "string"}},
    "dependencies": {"type": "array", "items": {"type": "string"}},
    "checksum": {"type": "string"},
    "signature": {"type": "string"},
    "metadata": {"type": "object"},
    "compatibility": {"type": "object", "properties": {
      "api": {"type": "string"},
      "backend": {"type": "string"}
    }}
  }
}
```

## 签名与校验流程
- 校验顺序：结构 → 版本与兼容性 → 依赖 → checksum → 签名。
- 签名来源：开发者私钥签名，平台公钥验证；CI/CD 内置重签名与验签。
- 审计记录：所有安装/启停/升级/卸载操作，落库 `lao_audit_logs`，包含操作人、目标、结果、payload。

## 后端接口（laojun 域）
- 插件注册/安装：`POST /api/laojun/plugins/install`（`pluginId`, `version`）
- 启停：`POST /api/laojun/plugins/start|stop`（`pluginId`）
- 升级：`POST /api/laojun/plugins/upgrade`（`pluginId`, `version`）
- 卸载：`DELETE /api/laojun/plugins/uninstall`（`pluginId`）
- 查询与审计：`GET /api/laojun/plugins`，`GET /api/laojun/audits`
- 数据落库：`lao_plugins`, `lao_plugin_versions`, `lao_audit_logs`

## 本地开发工具链（C/S）
- 能力：脚手架/模板生成、构建打包、签名校验、诊断日志采集与上传。
- 流程：初始化模板 → 编码 → 本地构建 → 签名与清单生成 → 推送仓库触发 CI/CD。
- 产物：源码包、构建产物（归档）、manifest、签名文件；可选 SourceMap/符号表。

## 前端集成（React 管理后台 + 原生端）
- React 管理后台：
  - 远程组件协议：插件提供组件入口（`entry`），后台通过动态加载（remote import）或 iframe 安全沙箱渲染（按安全级别决定）。
  - UI 集成：定义统一事件总线（`postMessage`/`EventEmitter`），插件仅可调用白名单接口；权限映射到后端鉴权。
  - 数据隔离：每插件独立命名空间与存储，前端缓存与后端资源隔离。
- 原生端：
  - 原生移动/机器人/手表不直接嵌入前端组件，提供轻量能力（任务触发、状态展示、日志），防止复杂 UI 插件破坏原生体验。
  - 与后端交互走统一接口与安全策略，支持离线队列与幂等处理。

## 版本策略与兼容性
- 语义化版本：插件版本与后端接口版本绑定；兼容性在 `compatibility` 字段声明并做校验。
- 兼容通道：灰度/蓝绿发布；版本通道标记 `stable`/`beta`/`dev`；回滚策略完善。

## 诊断与监控
- 指标：安装与启停成功率、升级失败率、平均耗时。
- 日志：插件运行日志与后端接口日志关联；故障采样与追踪。
- 报警：失败阈值触发告警与自动回滚。

## 元数据与清单
- 清单字段：`id, name, version, description, permissions, resources, endpoints, ui_components, checksum`
- 签名：SHA256 + `RSA-2048`/`Ed25519`，平台侧校验流程与公钥管理详见开发手册。
- 兼容策略：版本语义化；破坏性变更需发布说明与回滚方案。

## 接口定义（端点）
- 生命周期：`Init()`, `Start()`, `Stop()`, `Health()`
- 业务处理：`Handle(event, payload)` → 返回统一响应包装与错误码映射。
- 资源约束：显式声明 CPU/Mem/GPU；平台侧限流与熔断策略。

## 调试工具与流程
- 本地调试：模拟平台回调与事件；Mock 权限与资源限制。
- 远端调试：沙箱隔离环境；灰度发布与回滚按钮。
- 观测与审计：日志采集、指标上报与审计轨迹（traceId）。

## 前端 UI 组件集成（原生）
- 组件注册：通过 `ui_components` 声明插入点与权限要求。
- iOS：Swift/Objective-C 组件桥接；大小写与资源加载处理；
- Android：View/Fragment 集成；主题与资源命名空间隔离；
- Harmony/桌面：组件容器与权限、文件系统隔离。

## 安全与合规
- 权限最小化：仅授予必要资源与动作；敏感能力需二次确认。
- 秘密管理：不在清单中明文存储；平台侧下发临时令牌。

> 参考：[0] https://www.doubao.com/thread/w30e30a4dcadbb935