# AI服务

## 接口概览
| 方法 | 路径 | 名称 | 来源 |
|------|------|------|------|
| POST | /api/v1/ai/text/generate | 文本生成 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/text/generate/stream | 文本生成流式 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/chat/sessions | 创建会话 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/chat/sessions/{session_id}/messages | 发送消息 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| GET  | /api/v1/ai/chat/sessions/{session_id}/messages | 会话消息列表 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/image/generate | 图像生成 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/image/analyze | 图像分析 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/speech/synthesize | 语音合成 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| POST | /api/v1/ai/moderation/text | 文本审核 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| GET  | /api/v1/ai/models | 模型列表 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |
| GET  | /api/v1/ai/models/{model_id}/metrics | 模型指标 | [docs/06-API文档/核心服务API/AI服务API.md](../../06-API文档/核心服务API/AI服务API.md) |