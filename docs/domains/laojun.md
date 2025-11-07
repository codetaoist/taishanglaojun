# 老君基础域（laojun）深入设计

定位：平台的基础能力域，负责插件全生命周期、审计日志、系统配置，路径 `/api/laojun/...`，表前缀 `lao_`。[0]

## 业务职责
- 插件生命周期：安装/启动/升级/停止/卸载，版本语义与兼容校验。
- 审计与合规：所有操作记录、查询与导出，支持追踪与告警。
- 系统配置：全局/模块配置，作用域与优先级，变更审计。

## 核心实体
- Plugin（插件）
  - 字段：`id`, `name`, `description`, `status`, `created_at`, `updated_at`
  - 约束：`name` 唯一；状态机 `inactive/active/installed`。
- PluginVersion（版本）
  - 字段：`id`, `plugin_id`, `version`, `manifest`, `signature`, `created_at`
  - 约束：`plugin_id+version` 唯一；FK 指向 `Plugin`。
- AuditLog（审计）
  - 字段：`id`, `actor`, `action`, `target`, `payload`, `result`, `created_at`
  - 索引：`actor, created_at`；可分区或归档。
- Config（配置）
  - 字段：`key`, `value`, `scope`, `updated_at`

## 数据落库
- 表：`lao_plugins`, `lao_plugin_versions`, `lao_audit_logs`, `lao_configs`
- 索引：`name`唯一、`plugin_id+version`唯一、`actor+created_at`复合、`scope`普通索引。

## 接口与路由
- 插件：`GET /api/laojun/plugins`, `POST /api/laojun/plugins/install`, `POST /api/laojun/plugins/start|stop`, `POST /api/laojun/plugins/upgrade`, `DELETE /api/laojun/plugins/uninstall`
- 审计：`GET /api/laojun/audits`
- 配置：`GET/PUT /api/laojun/configs`

## 关键流程
- 安装：校验清单→签名→依赖→触发CI→部署→审计。
- 升级：版本校验→兼容性检查→灰度→回滚预案→审计。
- 启停：状态机切换→资源检查→审计。

## 策略与门禁
- 幂等：操作带去重键；失败可重试；审计不重复。
- 权限：管理员/运维/只读；操作级权限粒度到动作。
- 安全：manifest与签名校验；来源可信；接口鉴权与速率限制。

## 规划总览
- 目标：构建可信、可管控的插件运行与集成域，覆盖清单/签名/权限、生命周期、UI 组件集成与审计回滚。
- 契约：`/api/laojun/plugins/*`（install/list/start/stop/upgrade/uninstall），统一响应包装与错误码映射。

## 业务流程（高层）
- 安装（Install）：校验清单 → 验签 → 存储注册 → 资源配额与权限授予 → 可回滚。
- 启停（Start/Stop）：资源检查 → 沙箱化加载/卸载 → 健康检查与观测 → 容错与熔断。
- 升级（Upgrade）：版本兼容评审 → 双轨部署（旧版并存）→ 切换与回滚。
- 卸载（Uninstall）：依赖检查 → 清理注册与持久化记录 → 留存审计轨迹。

## 功能分解
- 清单管理：字段完整性、`checksum` 校验、兼容策略（SemVer）。
- 签名与来源：`RSA-2048`/`Ed25519` 验签，公钥管理与撤销列表。
- 权限与资源：最小授权、CPU/Mem/GPU 配额、文件系统与网络隔离。
- 生命周期 API：`Init/Start/Stop/Health`，异常回滚与幂等处理。
- UI 组件集成：组件注册、权限提示、原生端插入点与主题隔离。
- 观测与审计：日志/指标/traceId，失败阈值告警与自动回滚策略。

## 细节逻辑与约束
- 交易性保障：安装/升级操作需事务化，失败自动回滚并记录。
- 并发与限流：并发启动队列与回压；接口与资源限流统一配置。
- 错误码对齐：平台侧统一错误码，插件侧映射并在响应包装返回。
- 数据持久化：插件注册表、权限授予记录、审计日志；参见 `db/migrations/V1__init_lao.sql`。
- 安全与合规：隐私与数据收集说明，供应链扫描（依赖来源与签名）。

## 接口契约映射
- 安装：`POST /api/laojun/plugins/install` → 返回 `code/data/message` 包装。
- 列表：`GET /api/laojun/plugins/list`
- 启动：`POST /api/laojun/plugins/start`
- 停止：`POST /api/laojun/plugins/stop`
- 升级：`POST /api/laojun/plugins/upgrade`
- 卸载：`POST /api/laojun/plugins/uninstall`

> 参考：`openapi/laojun.yaml`，`docs/plugins/plugin-system-details.md`，`docs/security/permission-model.md`
- 安全：manifest与签名校验；来源可信；接口鉴权与速率限制。

## 能力列表（What we do）
- 插件生态治理：开发→测试→部署→安装→运行→卸载全闭环，版本兼容与回滚策略。
- 平台审计与合规：操作事件审计、告警与导出；接口防护与来源可信。
- 系统与设备基础能力：用户与权限、设备管理、配置与适配（多协议）。

## 实现路径（How we do it）
- 契约与脚本：使用 `openapi/laojun.yaml` 定义接口契约；在本地与 CI 使用 `scripts/manifest_validate.py`、`scripts/manifest_sign.py` 与 `scripts/openapi_contract_diff.py` 做清单校验、签名与合同 diff。
- 生命周期与沙箱：通过 `Init/Start/Stop/Health` API 驱动插件生命周期，在容器/沙箱内隔离运行并施加配额与权限。
- 权限与门禁：RBAC 与作用域授权（global/workspace/instance），`enabledActions` 驱动前端可见；门禁校验覆盖契约/安全/测试与性能基线。
- 审计与观测：统一日志/指标/traceId，失败阈值与自动回滚；插件与平台的审计事件闭环。

> 参阅：`docs/plugins/development-manual.md`、`scripts/README.md`、`docs/interfaces/laojun-api-spec.md`
- 参考：[0] https://www.doubao.com/thread/w1745a17f59b91183