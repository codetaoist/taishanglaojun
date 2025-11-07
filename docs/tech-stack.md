# 技术栈总览（原生 + React）

根据页面最终要求，各端使用原生语言，Web 前端采用 React；后端 Go 模块地址为 `github.com/codetaoist/taishanglaojun`。[0]

## 终端与栈
- Web 管理后台：React + Vite + TypeScript + Ant Design。
- iOS：Swift/SwiftUI（Xcode 15+）。
- Android：Kotlin/Jetpack Compose（Android Studio Giraffe+）。
- 机器人端：C++/Qt（含 QML）。
- 手表端：Apple Watch（SwiftUI）或 微信小程序原生框架。

## 后端与数据
- Go（Gin）；OpenAPI 3.0；Redis；MySQL/PostgreSQL；向量库 Milvus/Faiss。
- 模块地址：`github.com/codetaoist/taishanglaojun`。

## 插件与CI/CD
- 管理后台（B/S）主导插件操作；C/S 为本地开发辅助；GitLab CI 负责构建/测试/部署与回滚。

## 选型理由简述
- 原生移动端：性能与平台能力最佳，满足通知、离线与权限场景。
- 机器人端原生：低延迟与设备级能力，适合控制与状态场景。
- React 管理后台：前端生态成熟、类型体系完整、企业后台组件丰富。
- Go 后端：并发与性能优势，生态成熟，便于模块化与部署。

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1