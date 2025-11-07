# 测试用例集（Unit / Integration / E2E）

覆盖老君（插件/审计/配置）与太上（模型/向量/任务）核心功能与边界场景。所有用例以 OpenAPI 契约为准。

## 一、单元测试（Unit）
- 鉴权中间件
  - 用例：缺失/非法令牌；不同角色的权限判定；作用域校验失败
  - 期望：返回 `UNAUTHENTICATED`/`PERMISSION_DENIED`；审计与日志一致
- 响应包装与错误码映射
  - 用例：控制器抛出异常 → 错误码统一映射；traceId 注入
- 插件服务（安装/启停/升级/卸载）
  - 用例：幂等键重复安装；版本不兼容；状态机非法转移
- 任务状态机（太上）
  - 用例：PENDING→RUNNING→SUCCEEDED/FAILED；重试与取消；幂等提交

## 二、集成测试（Integration）
- 插件全流程
  - 安装→启动→停止→升级→卸载；审计日志生成；数据库写入一致性
  - 边界：权限不足、签名不通过、依赖缺失
- 配置管理
  - 列表→查询→更新；作用域覆盖（global/domain/plugin）
  - 边界：非法 scope、值类型不匹配、并发写入冲突
- 模型与向量集合
  - 注册模型→禁用/启用；创建向量集合→查询参数→索引配置校验
  - 边界：重复名称、维度不匹配、索引类型非法

## 三、端到端（E2E）
- 管理后台（Web）
  - 插件管理页面：操作按钮权限控制；二次确认与结果提示；审计链接
  - 配置页面：表单校验与错误码提示；分页与过滤一致性
- 原生终端
  - 任务提交→状态追踪→通知/重试/取消；设备指纹绑定与安全校验
  - 边界：网络切换、低电量模式、限流/熔断触发

## 四、性能与资源（基准）
- API 延迟：P95 < 200ms；错误率 < 1%
- 向量检索：P95 < 500ms；并发与队列深度控制
- 插件运行：CPU/Mem 配额与限流；启动风暴抑制

## 五、报告与门禁
- 覆盖率门槛：Unit ≥ 70%；Contract ≥ 90%；E2E 关键路径 ≥ 90%
- 报告聚合：测试报告与审计对齐；CI 阶段门禁与失败重跑策略

## 六、接口级边界输入矩阵（示例）
- 插件安装（POST /api/plugins/install）：
  - 必填缺失：`pluginId=null` → `INVALID_ARGUMENT`
  - 版本格式：`version="abc"` 非语义化 → `INVALID_ARGUMENT`
  - 重复安装：同 `pluginId+version` 幂等键 → `CONFLICT`
  - 权限不足：普通用户安装需要 `plugin:install` → `PERMISSION_DENIED`
- 向量集合创建（POST /api/taishang/vectors/collections）：
  - 维度范围：`dim<=0` 或 `dim>4096` → `INVALID_ARGUMENT`
  - 索引枚举：`indexType="UNKNOWN"` → `INVALID_ARGUMENT`
  - 名称重复：`name` 已存在 → `CONFLICT`
- 任务提交（POST /api/taishang/tasks）：
  - 类型枚举：`type="unknown"` → `INVALID_ARGUMENT`
  - 负载大小：`payload` 超限 → `FAILED_PRECONDITION`
  - 未认证：缺失令牌 → `UNAUTHENTICATED`

## 七、并发场景表（关键接口）
- 插件启停（/api/plugins/start|stop）：
  - 并发启动同一插件：仅首个成功，后续返回 `FAILED_PRECONDITION`；状态机一致性
  - 启动与卸载竞争：卸载需阻塞 `running` 状态；并发请求返回 `FAILED_PRECONDITION`
- 向量写入（upsert/delete）：
  - 同集合并发 `upsert`：分区锁或乐观并发控制；写入顺序与最终一致性验证
  - 删除与查询竞争：删除采用软删除+版本号；查询基于快照一致性
- 任务状态机：
  - 并发 `retry/cancel`：以任务当前状态判定；重复操作返回幂等成功或 `FAILED_PRECONDITION`
  - 队列深度与限流：达到阈值触发 `UNAVAILABLE`；验证退避与重试策略