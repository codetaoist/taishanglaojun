# 原生终端架构与跨端适配

覆盖 iOS/Android/机器人/手表终端，统一插件交互、任务编排与测试适配。

## 架构概览
- 核心模块：Auth、Contracts、TaskQueue、PluginBridge、Analytics。
- 通信：HTTPS + JWT；WebSocket 订阅任务状态；设备指纹绑定安全策略。
- 插件桥接：原生与插件通信协议（JSON/Protobuf）；权限约束与沙箱。

## 终端适配
- iOS：Swift + Combine；后台任务与通知；Keychain 管理密钥。
- Android：Kotlin + Coroutines；WorkManager 任务；Keystore 安全。
- Robot：ROS/自定义SDK；任务执行与传感器数据上报；安全通道。
- Watch：精简 UI + 任务通知；低功耗与限流策略。

## 任务编排
- 状态机：`PENDING -> RUNNING -> SUCCEEDED/FAILED -> RETRY/CANCELLED`
- 幂等与重试：任务ID防重；指数退避；失败审计与上报。
- 资源管理：并发上限、网络策略与低电量模式保护。

## 测试适配
- 单元：平台原生测试框架；覆盖核心模块。
- 合同：与 OpenAPI 契约对齐；Mock 后端与边界测试。
- 端到端：真实设备/模拟器；任务流与通知验证；覆盖关键路径。

## 安全与审计
- 设备绑定与安全存储；敏感日志脱敏；异常审计与上报。
- 插件签名与权限验证；越权阻断与用户提示。