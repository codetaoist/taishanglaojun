# 老君（laojun）接口规范文档

以 `openapi/laojun.yaml` 为事实源，定义插件/审计/配置三大模块的接口、参数、响应与错误码。所有接口路径前缀：`/api/laojun`。

## 通用规范
- 认证：`Authorization: Bearer <JWT>`；角色与作用域见权限模型。
- 响应包装：`{ code, message, data, traceId }`；错误码见下文。
- 分页：查询参数 `page`, `pageSize`，响应 `{ total, page, pageSize, items }`。
- 排序与过滤：`sort=field:asc,created_at:desc`；常用过滤 `status`, `name`。
- 版本：路径或媒体类型版本（`Accept: application/vnd.codetaoist.v1+json`）。
- 头部：`X-Workspace-Id?`、`Idempotency-Key?`。

## 模块一：插件（Plugins）
- 列表插件
  - `GET /plugins/list`
  - Query：`status?`、`name?`、`page`、`pageSize`
  - 200 示例：
    ```json
    {"code":"OK","data":{"total":1,"page":1,"pageSize":20,"items":[{"id":"speech_tool","name":"语音工具","version":"1.0.0","status":"stopped","checksum":"sha256:..."}]}}
    ```
  - curl 示例：
    ```bash
    curl -H "Authorization: Bearer $JWT" \
      "$API/api/laojun/plugins/list?page=1&pageSize=20&status=stopped"
    ```
- 安装插件
  - `POST /plugins/install`
  - Body：`{ "pluginId": "string", "version": "string", "sourceUrl?": "string" }`
  - 200 示例：`{"code":"OK","data":{"installed":true,"pluginId":"speech_tool","version":"1.0.0"}}`
  - 幂等键：`pluginId+version`
  - curl 示例：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"pluginId":"speech_tool","version":"1.0.0","sourceUrl":"https://repo.example/..."}' \
      "$API/api/laojun/plugins/install"
    ```
- 启动插件
  - `POST /plugins/start`
  - Body：`{ "pluginId": "string" }`
  - 200 示例：`{"code":"OK","data":{"running":true,"pluginId":"speech_tool"}}`
  - curl 示例：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"pluginId":"speech_tool"}' "$API/api/laojun/plugins/start"
    ```
- 停止插件
  - `POST /plugins/stop`
  - Body：`{ "pluginId": "string" }`
  - 200 示例：`{"code":"OK","data":{"running":false,"pluginId":"speech_tool"}}`
  - curl 示例：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"pluginId":"speech_tool"}' "$API/api/laojun/plugins/stop"
    ```
- 升级插件
  - `POST /plugins/upgrade`
  - Body：`{ "pluginId": "string", "version": "string" }`
  - 200 示例：`{"code":"OK","data":{"upgraded":true,"pluginId":"speech_tool","version":"1.1.0"}}`
  - curl 示例：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"pluginId":"speech_tool","version":"1.1.0"}' "$API/api/laojun/plugins/upgrade"
    ```
- 卸载插件
  - `DELETE /plugins/uninstall`
  - Body：`{ "pluginId": "string" }`
  - 200 示例：`{"code":"OK","data":{"uninstalled":true,"pluginId":"speech_tool"}}`
  - curl 示例：
    ```bash
    curl -X DELETE -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"pluginId":"speech_tool"}' "$API/api/laojun/plugins/uninstall"
    ```
- 查询插件详情
  - `GET /plugins/list`（结合 `name` 过滤或通过 SDK 提供的详情接口）

## 模块二：审计（Audits）
- 查询审计日志
  - `GET /audits`
  - Query：`actor?`、`action?`、`resource?`、`from?`、`to?`、`page`、`pageSize`
  - 200 示例：
    ```json
    {"code":"OK","data":{"total":2,"page":1,"pageSize":20,"items":[{"id":1,"actor":"admin","action":"install","resource":"plugin:speech_tool","status":"SUCCESS","message":"installed","created_at":"2025-10-01T10:00:00Z"}]}}
    ```
- 获取审计详情
  - `GET /audits/{id}`
  - 200 示例：`{"code":"OK","data":{"id":1,"actor":"admin","action":"install","payload":{"pluginId":"speech_tool"}}}`

## 模块三：配置（Configs）
- 列表配置项
  - `GET /configs`
  - Query：`key?`、`page`、`pageSize`
  - 200 示例：`{"code":"OK","data":{"total":1,"items":[{"key":"jwt.secret","value":"***","scope":"global"}]}}`
- 获取配置详情
  - `GET /configs/{key}`
  - 200 示例：`{"code":"OK","data":{"key":"jwt.secret","value":"***","scope":"global"}}`
- 更新配置项
  - `PUT /configs/{key}`
  - Body：`{ "value": "string", "scope": "global|domain|plugin:<id>" }`
  - 200 示例：`{"code":"OK","data":{"updated":true,"key":"jwt.secret"}}`

## 错误码定义
- `OK`：成功
- `INVALID_ARGUMENT`：参数错误（字段缺失/类型不符）
- `UNAUTHENTICATED`：未认证（缺少/非法令牌）
- `PERMISSION_DENIED`：权限不足（角色/作用域不匹配）
- `NOT_FOUND`：资源不存在
- `CONFLICT`：冲突（重复安装/版本不兼容）
- `FAILED_PRECONDITION`：前置条件不满足（依赖未满足/插件状态不允许）
- `INTERNAL`：服务器错误
- `UNAVAILABLE`：服务不可用（限流/熔断/依赖宕机）

示例错误响应：
```json
{"code":"PERMISSION_DENIED","message":"role not allowed","traceId":"..."}
```

## 契约校验与示例约束
- 所有请求/响应结构以 `openapi/laojun.yaml` 为准；此文档作为人类可读补充。
- 新增/变更接口需评审并在 CI 中通过契约测试与覆盖率门禁。