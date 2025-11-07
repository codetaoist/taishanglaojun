# 开发规范

## 通用
- 统一以 OpenAPI 3.0 作为接口事实源，服务端与前端类型均由其派生。[0]
- 代码风格遵循各语言主流规范与本项目约定。
- 严格区分环境（dev/test/prod），配置走环境变量与集中配置。

## Go 规范
- 版本：Go ≥ 1.21。
- 模块地址：`github.com/codetaoist/taishanglaojun`。
- 框架：Gin。
- 命名：包小写、无下划线；导出类型使用驼峰；接口以 `er` 或能力名结尾；错误以 `ErrXxx` 表示。
- 目录：
  - `cmd/api/` 启动入口
  - `internal/` 业务逻辑（`app/`, `domain/`, `infra/`）
  - `pkg/` 通用工具
- 错误处理：显式错误返回，分层封装，日志中包含 traceId。

## Web 管理后台（React）规范
- 技术栈：React + Vite + TypeScript + Ant Design。
- 状态管理：Redux Toolkit 或 Zustand（按规模选型）；接口调用统一封装；类型从 OpenAPI/TS 类型生成。
- 组件命名：语义化命名；页面在 `pages/`，通用组件在 `components/`；路由在 `routes/`。
- 样式：Ant Design + CSS-in-JS 或 Less；主题统一管理；无内联样式除非必要。
- 交互：统一错误提示与重试；列表分页、搜索、过滤；支持批量操作与回滚提示。

## iOS 原生规范（Swift/SwiftUI）
- 架构：MVVM；Combine 或 async/await。
- 网络：统一 API 客户端；错误码与重试规则遵循接口规范。
- 存储：Keychain 管理机密；必要数据持久化使用 CoreData 或 SQLite。

## Android 原生规范（Kotlin/Jetpack Compose）
- 架构：MVVM + Coroutine + Flow。
- 网络：统一 API 客户端；错误码与重试规则遵循接口规范。
- 存储：EncryptedSharedPreferences 管理机密；Room 持久化。

## 机器人端原生（C++/Qt/QML）
- 架构：QML 界面 + C++ 业务；模块化组件；跨线程通信安全。
- 网络：统一 REST 客户端；错误处理与重试一致；日志与审计上报一致。

## 手表端原生
- Apple Watch：SwiftUI；轻交互、低耗电；通知优先。
- 微信小程序：原生框架；页面轻量、接口对接一致。

## 数据库与接口规范
- 数据库：统一建表 SQL 与迁移脚本；索引与约束必须在文档与 SQL 中明确。[0]
- 接口：遵循 REST；统一错误码结构；分页、过滤、排序规范统一。

## 插件规范
- 插件包结构、清单（manifest）、签名校验；版本语义化；生命周期回调。
- 安装/启动/升级/卸载流程统一由管理后台驱动，后端校验与执行；CI/CD 贯穿构建与部署。[0]

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1