# 开发环境搭建指南

## 前置要求
- macOS/Linux 开发机。
- Go ≥ 1.21（后端，模块地址：`github.com/codetaoist/taishanglaojun`）。
- Node.js ≥ 18 与 pnpm（Web 管理后台 React）。
- iOS：Xcode 15+（Swift/SwiftUI）。
- Android：Android Studio Giraffe+（Kotlin/Jetpack Compose）。
- 机器人端：Qt 6 开发环境（含 QML）。
- 手表端：Xcode（Apple Watch）或 微信开发者工具（原生小程序）。
- Git 与 GitLab 访问权限（CI/CD）。
- 数据库与缓存：本地 MySQL/PostgreSQL、Redis；（可选）Milvus/Faiss。

## 克隆与基础目录
- 目录结构参考项目根 `README.md` 与 `docs/README.md`。

## 环境变量
- 后端：`PORT`, `DB_DSN`, `REDIS_URL`, `JWT_SECRET`, `CI_GATEWAY_URL`。
- Web：`VITE_API_BASE`。
- 移动/机器人/手表端：统一 `API_BASE` 与认证配置。

## 本地数据库初始化
- 执行 `db/schema.sql`，初始化基础表与审计日志表。

## 文档与接口
- 以 `openapi/openapi.yaml` 为契约；推荐使用 `Swagger UI` 或 `Redoc` 本地预览。

## 常见问题
- 端口占用：调整 `PORT` 或关闭占用进程。
- 数据库连接失败：确认 `DB_DSN`，检查数据库服务与权限。
- 原生端 SDK 版本不匹配：统一升级到文档指定版本。

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1