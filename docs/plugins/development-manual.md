# 插件开发手册（Go Plugin / Docker 插件 / AI 插件）

面向三类插件的统一开发规范：清单、权限、SDK 接口、开发示例与打包流程。安装/启停/升级/卸载均由老君域接口驱动。

## 通用要求
- 清单与签名：`manifest.json` 包含 `id,name,version,entry,permissions,checksum,signature`；详见 `plugin-manifest.md`。
- 权限声明：`permissions` 映射后端权限位；安装时校验与审计。
- 版本与兼容：语义化版本；声明兼容范围与依赖。
- 测试与发布：契约/单测/集成/E2E 均需达标；CI 门禁。

## Go Plugin（动态插件/进程内或外部进程）
- SDK 接口（示意）
  - 初始化：`Init(ctx, cfg)`
  - 启动：`Start(ctx)`；停止：`Stop(ctx)`
  - 能力接口：按插件类型暴露如 `Handle(input) -> output`
- 开发示例要点
  - 读取配置与权限；校验清单与签名；暴露健康探针。
  - 安全隔离：尽量进程外（gRPC/HTTP）以降风险；进程内需严格审计。
- 打包流程
  - 编译与产物：`plugin.bin` 或容器化（推荐）
  - 生成清单与签名：`manifest.json` + `signature`（私钥签名、公钥验签）
  - 发布：上传产物与清单至插件仓库或对象存储；登记版本。

## Docker 插件（容器化交付）
- SDK 接口
  - HTTP/gRPC 协议暴露；插件自身实现 `/healthz`、`/readyz`。
  - 与平台通信：通过平台提供的 `PluginBridge`（JWT + 作用域）。
- 开发示例要点
  - 显式资源配额（CPU/Mem）与限流；日志与指标上报。
  - 网络与安全：仅开放必要端口；最小权限运行用户。
- 打包流程
  - `Dockerfile` 构建镜像；打标签 `plugin:<id>-<version>`。
  - 推送至镜像仓库；生成 `manifest.json` 记录镜像引用与校验和。

## AI 插件（模型/推理/工具类）
- SDK 接口（Python/Go 均可）
  - 注册能力：`register_tool(name, schema)` 或 `RegisterModel(name, version)`
  - 推理入口：`infer(inputs, params) -> outputs`
  - 资源管理：缓存与显存调度（需配置化）；任务队列与优先级。
- 开发示例要点
  - 兼容契约：请求/响应严格遵循平台定义；数据脱敏与审计。
  - 性能建议：量化、批处理与并发；错误重试与降级。
- 打包流程
  - Python 包或容器镜像；产物含模型文件与校验；`manifest.json` 记录依赖与版本。

## 本地开发与调试
- 本地工具链：启动平台的 `dev` 模式；提供插件热重载与日志聚合。
- 契约测试：基于 OpenAPI 生成插件侧的接口测试；联调前必须通过。
- 诊断与监控：插件级指标、日志与审计可在管理后台可视化。

## 发布与回滚
- 发布门禁：签名校验、权限审查、测试覆盖率、秘密扫描与依赖风险扫描。
- 回滚策略：上一版本快速回退；禁用插件与清理副作用；审计记录全链路。

## 签名工具与验签流程
- 清单签名格式：`manifest.json` 以 `SHA256` 计算 `checksum`，使用 `RSA-2048` 或 `Ed25519` 私钥对 `checksum` 进行签名，生成 `signature`。
- 参考工具：
  - Go：`crypto/ed25519` 或 `crypto/rsa`；示例命令：`go run tools/sign.go -manifest manifest.json -key private.key`。
  - Python：`cryptography`；示例：`python tools/sign.py --manifest manifest.json --key private.key`。
- 验签流程（平台侧）：
  1) 校验 `manifest.json` 结构与必填项；
  2) 重新计算 `checksum` 与清单中的值一致；
  3) 使用注册的公钥（插件发布者或组织）验证 `signature`；
  4) 比对权限声明与后端权限位，进行审计与告警；
  5) 通过后进入安装流程；失败则拒绝并记录审计。
- 公钥管理：平台维护公钥白名单（按组织/发布者），定期轮转；撤销列表用于阻断被泄露密钥的插件发布。

## 沙箱隔离建议
- 进程隔离优先：Docker/容器化交付，禁用特权，设置 `CPU/Mem` 配额与 `ulimit`；
- 网络最小化：仅开放必需端口；出站白名单；禁止访问宿主敏感路径；
- 文件系统限制：只读根文件系统；挂载仅限必要数据目录；
- 权限与令牌：插件运行用户非 `root`；作用域化 JWT（最小权限原则）；
- 资源防护：限流与熔断；启动风暴抑制；任务队列优先级隔离；
- 观测与审计：插件级日志/指标/追踪；关键操作（安装/启停/升级/卸载）全量审计；
- 灰度与回滚：新版本分批放量；异常自动回滚并冻结问题版本；

## 最小可运行脚本示例（校验/签名/合同）
- 本地依赖安装：`pip install -r scripts/requirements.txt`
- 清单校验（laojun）：`python scripts/manifest_validate.py --manifest path/to/manifest.json`
- 清单签名（示例 HMAC）：`python scripts/manifest_sign.py --manifest path/to/manifest.json --secret <secret> --out path/to/manifest.signed.json`
- OpenAPI 校验：`python scripts/openapi_validate.py --spec openapi/laojun.skeleton.yaml --full`
- 合同差异对比：`python scripts/openapi_contract_diff.py --base openapi/laojun.yaml --candidate openapi/laojun.skeleton.yaml`

### 校验项明细（Laojun 侧）
- 必填字段：`id,name,version,entry(type),permissions`；可选：`dependencies, checksum, signature`。
- 版本格式：语义化版本（如 `1.2.3`）。
- 权限声明：需映射后端权限位（菜单/按钮级授权），未匹配则提示并拒绝安装。
- 签名一致性：`checksum` 与 `signature` 均应存在且能验证；最小示例使用 HMAC，生产使用 RSA/Ed25519。
- 审计记录：校验失败与成功均记录（含 `traceId`）。

### 签名与验签（与平台侧集成）
- 开发侧：使用脚本生成 `checksum` 与 `signature`；提交产物与清单。
- 平台侧：使用白名单公钥验证签名；权限与依赖二次审查；通过后进入安装任务。
- 风险提示：密钥泄露、恶意权限、依赖风险需纳入门禁与告警。

### CI 门禁集成
- 合同门禁：OpenAPI 校验与 diff 检查（禁止未审查的破坏性变更）。
- 原生质量：清单校验必须通过；测试覆盖率与安全扫描达标。
- 插件生态：签名验证、依赖安全与资源配额检查，阻断违规发布。

## 域职责分离与边界
- laojun（基础域）：专注插件生态（清单/签名/安装/启停/升级/卸载）与审计观测，提供 UI 容器与权限治理。
- taishang（高级域）：专注模型/向量/任务编排与成本治理；不处理插件清单与签名（由 laojun 完成）。

> 参见：`docs/interfaces/standard.md`、`docs/security/permission-model.md`、`docs/ops/ci-cd-pipeline.md`、`openapi/*.skeleton.yaml`