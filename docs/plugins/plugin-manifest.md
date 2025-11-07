# 插件 manifest 与签名校验

## 清单结构（示例）
```json
{
  "id": "example-plugin",
  "name": "Example Plugin",
  "version": "1.2.3",
  "entry": {
    "type": "service",
    "command": "./bin/start",
    "args": ["--port=9000"]
  },
  "permissions": ["read:db", "write:cache"],
  "dependencies": {
    "runtime": ["redis>=6"],
    "plugins": ["other-plugin>=1.0.0"]
  },
  "checksum": "sha256:...",
  "signature": "base64-encoded-signature",
  "metadata": {
    "description": "...",
    "author": "...",
    "repo": "https://..."
  }
}
```

## 校验流程
1. 清单结构校验与必填字段确认。
2. 版本校验（语义化版本）。
3. 依赖检查（运行时与插件依赖）。
4. 校验和与签名验证（公钥验证）。
5. 审计记录（清单、版本、结果、traceId）。

## 安装与升级约束
- 不允许降级；升级需平滑迁移与回滚策略。
- 安装/升级必须通过管理后台触发，后端执行并记录审计。[0]

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1