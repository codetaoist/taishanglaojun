# 技术栈总览（混合架构：Go + Python 微服务）

根据项目发展需求，采用混合架构：Laojun域使用Go语言，Taishang域核心API使用Go语言，AI计算功能使用Python微服务。模块地址为 `github.com/codetaoist/taishanglaojun`。[0]

## 终端与栈
- Web 管理后台：React + Vite + TypeScript + Ant Design。
- iOS：Swift/SwiftUI（Xcode 15+）。
- Android：Kotlin/Jetpack Compose（Android Studio Giraffe+）。
- 机器人端：C++/Qt（含 QML）。
- 手表端：Apple Watch（SwiftUI）或 微信小程序原生框架。

## 后端与数据（混合架构）
- Laojun域（基础功能）：Go（Gin）
- Taishang域（AI能力）：
  - 核心API：Go（Gin）
  - AI计算微服务：Python（FastAPI）
- 服务间通信：gRPC + Protocol Buffers
- 数据存储：MySQL/PostgreSQL；Redis；向量库 Milvus/Faiss
- 模块地址：`github.com/codetaoist/taishanglaojun`

## Python技术栈详情（Taishang AI微服务）
- Web框架：FastAPI
  - 高性能异步框架，自动API文档生成
  - 支持类型提示和Pydantic数据验证
  - 原生支持异步处理，适合AI计算场景
- AI/ML框架：
  - Transformers（Hugging Face）：预训练模型集成
  - LangChain：LLM应用开发框架
  - NumPy/Pandas：数据处理与分析
  - Scikit-learn：机器学习算法库
  - PyTorch/TensorFlow：深度学习框架
- 向量处理：
  - Sentence-Transformers：文本嵌入模型
  - FAISS：向量相似度搜索库
  - Milvus SDK：向量数据库客户端
- 异步处理：
  - Celery：分布式任务队列
  - asyncio：Python异步编程支持
  - aiohttp：异步HTTP客户端/服务器
- 容器化与部署：
  - Docker：容器化部署
  - Uvicorn：ASGI服务器
  - Gunicorn：WSGI服务器（用于生产环境）

## 跨语言通信设计
- gRPC协议：
  - Protocol Buffers定义服务接口
  - 支持双向流通信
  - 高性能二进制传输
- 服务发现：
  - Consul或Kubernetes Service
  - 健康检查和负载均衡
- 错误处理：
  - 统一错误码和错误消息格式
  - 跨语言异常传播机制
- 监控与追踪：
  - OpenTelemetry：分布式追踪
  - Prometheus：指标收集
  - 结构化日志：跨服务日志关联

## 插件与CI/CD
- 管理后台（B/S）主导插件操作；C/S 为本地开发辅助；GitLab CI 负责构建/测试/部署与回滚。

## 选型理由简述
- 原生移动端：性能与平台能力最佳，满足通知、离线与权限场景。
- 机器人端原生：低延迟与设备级能力，适合控制与状态场景。
- React 管理后台：前端生态成熟、类型体系完整、企业后台组件丰富。
- Go 后端（Laojun域 + Taishang核心API）：并发与性能优势，生态成熟，便于模块化与部署。
- Python 微服务（Taishang AI计算）：AI/ML生态丰富，便于集成各种AI框架和模型。
- 混合架构优势：结合Go的高并发性能和Python的AI生态优势，实现最佳技术组合。

## 混合架构设计
- 短期：Laojun域使用Go，Taishang域核心API使用Go，AI计算引入Python微服务
- 中期：扩展Python应用范围，建立跨语言通信标准
- 长期：评估整体架构，可能将Taishang域完全迁移至Python

## 性能考虑
- Go服务：处理高并发请求，提供稳定的API服务
- Python服务：专注于AI计算，通过异步处理提高吞吐量
- 跨语言调用：优化gRPC连接池，减少序列化/反序列化开销
- 缓存策略：在Go服务层实现缓存，减少对Python服务的调用

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1