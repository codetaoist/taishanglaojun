# Gin 路由与中间件细节

聚焦鉴权、审计与契约校验，实现安全与一致性。

## 鉴权中间件（示例）
```go
func Auth() gin.HandlerFunc {
  return func(c *gin.Context) {
    token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
    claims, err := jwt.Parse(token)
    if err != nil { c.AbortWithStatusJSON(401, resp.Err("UNAUTHENTICATED")); return }
    c.Set("role", claims.Role)
    c.Next()
  }
}
```

## 审计中间件（示例）
```go
func Audit() gin.HandlerFunc {
  return func(c *gin.Context) {
    start := time.Now()
    c.Next()
    if c.Request.Method != http.MethodGet {
      go writeAudit(c, start)
    }
  }
}
```

## 契约校验（示例）
```go
func Contracts(schema *openapi.Schema) gin.HandlerFunc {
  return func(c *gin.Context) {
    if err := schema.ValidateRequest(c.Request); err != nil {
      c.AbortWithStatusJSON(400, resp.Err("INVALID_ARGUMENT")); return
    }
    c.Next()
  }
}
```

## 统一响应
- 成功：`{ code: "OK", message: "", data }`
- 失败：`{ code: "<ERROR>", message, traceId }`

## 路由组合建议
- `/api/laojun`：插件生命周期操作默认开启审计与限流。
- `/api/taishang`：任务提交与模型管理开启契约校验与优先级队列。