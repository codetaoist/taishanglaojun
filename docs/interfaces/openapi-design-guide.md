# OpenAPI 设计指南（分域路由与契约生成）

以 `openapi/laojun.yaml` 与 `openapi/taishang.yaml` 为事实源，统一分域路由、资源命名、分页/过滤/排序与错误码，驱动契约测试与前后端协作。[0]

## 路由与版本
- 分域：`/api/laojun/...`（插件/审计/配置），`/api/taishang/...`（模型/向量/任务）。
- 版本：路径或媒体类型版本（`Accept: application/vnd.codetaoist.v1+json`）。
- 命名：复数资源名（`/plugins`），动作型接口用动词路径（`/install`、`/upgrade`）。

## 资源与响应结构
- 统一响应包装：`{ code, message, data, traceId }`
- 分页：`{ total, page, pageSize, items }`，查询参数 `page`, `pageSize`
- 过滤：多条件查询参数，`status=active&name=foo`
- 排序：`sort=name:asc,created_at:desc`

## 错误码与幂等
- 错误码集合：`OK`, `INVALID_ARGUMENT`, `UNAUTHENTICATED`, `PERMISSION_DENIED`, `NOT_FOUND`, `CONFLICT`, `FAILED_PRECONDITION`, `INTERNAL`, `UNAVAILABLE`
- 幂等：安装/启停/升级/卸载需支持去重键；任务接口支持幂等提交与状态查询。

## 示例片段（laojun/plugins）
```yaml
paths:
  /api/laojun/plugins:
    get:
      summary: List plugins
      responses:
        '200':
          description: OK
  /api/laojun/plugins/install:
    post:
      summary: Install plugin
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [pluginId, version]
              properties:
                pluginId: { type: string }
                version: { type: string }
      responses:
        '200':
          description: OK
```

## 生成与校验
- 代码生成：使用 OpenAPI 工具生成后端 DTO、前端类型与 SDK。
- 契约测试：解析 `openapi/*.yaml`，生成请求/响应/错误码测试用例；CI 门禁。
- 版本变更：变更必须评审与版本标记；前后端同步升级。

> 参考：[0] https://www.doubao.com/thread/w1745a17f59b91183