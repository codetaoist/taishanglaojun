# 太上老君AI平台API参考文档

## 概述

太上老君AI平台提供完整的RESTful API，支持开发者集成平台的各项AI功能。本文档详细介绍了所有可用的API接口、参数说明、响应格式和使用示例。

## API基础信息

### 基础URL

```
生产环境: https://api.taishanglaojun.ai/v1
测试环境: https://api-dev.taishanglaojun.ai/v1
```

### 认证方式

API使用Bearer Token认证，需要在请求头中包含有效的API密钥：

```http
Authorization: Bearer YOUR_API_KEY
```

### 请求格式

- **Content-Type**: `application/json`
- **字符编码**: UTF-8
- **请求方法**: GET, POST, PUT, DELETE

### 响应格式

所有API响应都采用统一的JSON格式：

```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "code": 200,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求参数无效",
    "details": "参数 'content' 不能为空"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 认证与授权

### 1. 获取API密钥

#### 创建API密钥

```http
POST /auth/api-keys
```

**请求头**
```http
Authorization: Bearer USER_TOKEN
Content-Type: application/json
```

**请求体**
```json
{
  "name": "我的API密钥",
  "description": "用于集成聊天功能",
  "permissions": ["chat", "image_generation"],
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "ak_1234567890",
    "name": "我的API密钥",
    "key": "sk-1234567890abcdef...",
    "permissions": ["chat", "image_generation"],
    "created_at": "2024-01-15T10:30:00Z",
    "expires_at": "2024-12-31T23:59:59Z"
  }
}
```

#### 列出API密钥

```http
GET /auth/api-keys
```

**响应**
```json
{
  "success": true,
  "data": {
    "keys": [
      {
        "id": "ak_1234567890",
        "name": "我的API密钥",
        "permissions": ["chat", "image_generation"],
        "created_at": "2024-01-15T10:30:00Z",
        "last_used": "2024-01-20T15:45:00Z",
        "status": "active"
      }
    ],
    "total": 1
  }
}
```

#### 删除API密钥

```http
DELETE /auth/api-keys/{key_id}
```

**响应**
```json
{
  "success": true,
  "message": "API密钥已删除"
}
```

### 2. 权限验证

#### 检查API权限

```http
GET /auth/permissions
```

**响应**
```json
{
  "success": true,
  "data": {
    "permissions": ["chat", "image_generation", "document_analysis"],
    "rate_limits": {
      "requests_per_minute": 60,
      "requests_per_day": 1000
    },
    "usage": {
      "requests_today": 150,
      "requests_this_month": 4500
    }
  }
}
```

## 智能对话API

### 1. 创建对话

#### 发送消息

```http
POST /chat/messages
```

**请求体**
```json
{
  "content": "你好，请介绍一下人工智能的发展历史",
  "conversation_id": "conv_1234567890",
  "model": "gpt-4",
  "temperature": 0.7,
  "max_tokens": 2000,
  "stream": false
}
```

**参数说明**
- `content` (string, 必需): 用户消息内容
- `conversation_id` (string, 可选): 对话ID，不提供则创建新对话
- `model` (string, 可选): 使用的AI模型，默认为 `gpt-4`
- `temperature` (float, 可选): 创造性参数，范围0-1，默认0.7
- `max_tokens` (integer, 可选): 最大响应长度，默认2000
- `stream` (boolean, 可选): 是否流式响应，默认false

**响应**
```json
{
  "success": true,
  "data": {
    "id": "msg_1234567890",
    "conversation_id": "conv_1234567890",
    "role": "assistant",
    "content": "人工智能的发展历史可以追溯到20世纪50年代...",
    "model": "gpt-4",
    "tokens_used": 156,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 流式对话

```http
POST /chat/messages
```

**请求体**
```json
{
  "content": "请写一首关于春天的诗",
  "stream": true
}
```

**响应 (Server-Sent Events)**
```
data: {"type": "start", "conversation_id": "conv_1234567890"}

data: {"type": "content", "content": "春"}

data: {"type": "content", "content": "风"}

data: {"type": "content", "content": "拂"}

data: {"type": "end", "message_id": "msg_1234567890", "tokens_used": 45}
```

### 2. 对话管理

#### 获取对话列表

```http
GET /chat/conversations
```

**查询参数**
- `page` (integer, 可选): 页码，默认1
- `limit` (integer, 可选): 每页数量，默认20
- `sort` (string, 可选): 排序方式，`created_at` 或 `updated_at`

**响应**
```json
{
  "success": true,
  "data": {
    "conversations": [
      {
        "id": "conv_1234567890",
        "title": "关于人工智能的讨论",
        "message_count": 8,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T11:45:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1,
      "pages": 1
    }
  }
}
```

#### 获取对话详情

```http
GET /chat/conversations/{conversation_id}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "conv_1234567890",
    "title": "关于人工智能的讨论",
    "messages": [
      {
        "id": "msg_1234567890",
        "role": "user",
        "content": "你好，请介绍一下人工智能",
        "created_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": "msg_1234567891",
        "role": "assistant",
        "content": "人工智能是计算机科学的一个分支...",
        "created_at": "2024-01-15T10:30:15Z"
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:45:00Z"
  }
}
```

#### 删除对话

```http
DELETE /chat/conversations/{conversation_id}
```

**响应**
```json
{
  "success": true,
  "message": "对话已删除"
}
```

### 3. 高级功能

#### 代码生成

```http
POST /chat/code-generation
```

**请求体**
```json
{
  "prompt": "创建一个Python函数来计算斐波那契数列",
  "language": "python",
  "style": "clean",
  "include_comments": true
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "code": "def fibonacci(n):\n    \"\"\"\n    计算斐波那契数列的第n项\n    \"\"\"\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)",
    "language": "python",
    "explanation": "这是一个递归实现的斐波那契函数...",
    "complexity": {
      "time": "O(2^n)",
      "space": "O(n)"
    }
  }
}
```

#### 数据分析

```http
POST /chat/data-analysis
```

**请求体**
```json
{
  "data": [
    {"month": "1月", "sales": 10000},
    {"month": "2月", "sales": 12000},
    {"month": "3月", "sales": 15000}
  ],
  "analysis_type": "trend",
  "questions": ["销售趋势如何？", "预测下个月销售额"]
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "analysis": "根据数据显示，销售额呈现上升趋势...",
    "insights": [
      "月度增长率约为20%",
      "预测4月销售额约为18000"
    ],
    "charts": [
      {
        "type": "line",
        "data": {...},
        "config": {...}
      }
    ]
  }
}
```

## 图像生成API

### 1. 基础图像生成

#### 文本生成图像

```http
POST /images/generate
```

**请求体**
```json
{
  "prompt": "一只可爱的小猫坐在花园里，阳光明媚，油画风格",
  "negative_prompt": "模糊，低质量，变形",
  "size": "1024x1024",
  "quality": "standard",
  "style": "natural",
  "n": 1
}
```

**参数说明**
- `prompt` (string, 必需): 图像描述提示词
- `negative_prompt` (string, 可选): 负面提示词
- `size` (string, 可选): 图像尺寸，支持 `256x256`, `512x512`, `1024x1024`
- `quality` (string, 可选): 图像质量，`standard` 或 `hd`
- `style` (string, 可选): 图像风格，`natural`, `vivid`
- `n` (integer, 可选): 生成图像数量，1-4

**响应**
```json
{
  "success": true,
  "data": {
    "id": "img_1234567890",
    "images": [
      {
        "url": "https://cdn.taishanglaojun.ai/images/img_1234567890_1.png",
        "width": 1024,
        "height": 1024,
        "format": "png"
      }
    ],
    "prompt": "一只可爱的小猫坐在花园里，阳光明媚，油画风格",
    "model": "dall-e-3",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. 图像编辑

#### 图像修复

```http
POST /images/edit
```

**请求体 (multipart/form-data)**
```
image: [图像文件]
mask: [遮罩文件]
prompt: "将猫咪的颜色改为橙色"
n: 1
size: "1024x1024"
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "img_edit_1234567890",
    "images": [
      {
        "url": "https://cdn.taishanglaojun.ai/images/img_edit_1234567890_1.png",
        "width": 1024,
        "height": 1024
      }
    ],
    "original_image": "https://cdn.taishanglaojun.ai/images/original_1234567890.png",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 图像变体

```http
POST /images/variations
```

**请求体 (multipart/form-data)**
```
image: [图像文件]
n: 2
size: "1024x1024"
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "img_var_1234567890",
    "images": [
      {
        "url": "https://cdn.taishanglaojun.ai/images/img_var_1234567890_1.png",
        "width": 1024,
        "height": 1024
      },
      {
        "url": "https://cdn.taishanglaojun.ai/images/img_var_1234567890_2.png",
        "width": 1024,
        "height": 1024
      }
    ],
    "original_image": "https://cdn.taishanglaojun.ai/images/original_1234567890.png",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 3. 图像管理

#### 获取图像列表

```http
GET /images
```

**查询参数**
- `page` (integer, 可选): 页码，默认1
- `limit` (integer, 可选): 每页数量，默认20
- `type` (string, 可选): 图像类型，`generated`, `edited`, `variation`

**响应**
```json
{
  "success": true,
  "data": {
    "images": [
      {
        "id": "img_1234567890",
        "url": "https://cdn.taishanglaojun.ai/images/img_1234567890_1.png",
        "prompt": "一只可爱的小猫坐在花园里",
        "size": "1024x1024",
        "type": "generated",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1,
      "pages": 1
    }
  }
}
```

#### 删除图像

```http
DELETE /images/{image_id}
```

**响应**
```json
{
  "success": true,
  "message": "图像已删除"
}
```

## 文档分析API

### 1. 文档上传

#### 上传文档

```http
POST /documents/upload
```

**请求体 (multipart/form-data)**
```
file: [文档文件]
name: "我的文档.pdf"
description: "这是一份重要的技术文档"
```

**支持的文件格式**
- PDF (.pdf)
- Word (.doc, .docx)
- PowerPoint (.ppt, .pptx)
- Excel (.xls, .xlsx)
- 文本文件 (.txt, .md)

**响应**
```json
{
  "success": true,
  "data": {
    "id": "doc_1234567890",
    "name": "我的文档.pdf",
    "size": 2048576,
    "type": "pdf",
    "pages": 10,
    "status": "processing",
    "upload_url": "https://cdn.taishanglaojun.ai/documents/doc_1234567890.pdf",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. 文档分析

#### 文档摘要

```http
POST /documents/{document_id}/summarize
```

**请求体**
```json
{
  "length": "medium",
  "language": "zh",
  "focus": ["主要观点", "结论", "建议"]
}
```

**参数说明**
- `length` (string, 可选): 摘要长度，`short`, `medium`, `long`
- `language` (string, 可选): 摘要语言，`zh`, `en`
- `focus` (array, 可选): 关注重点

**响应**
```json
{
  "success": true,
  "data": {
    "summary": "本文档主要讨论了人工智能在医疗领域的应用...",
    "key_points": [
      "AI在诊断方面的突破",
      "机器学习在药物发现中的作用",
      "未来发展趋势和挑战"
    ],
    "word_count": 156,
    "reading_time": "2分钟",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 关键词提取

```http
POST /documents/{document_id}/keywords
```

**请求体**
```json
{
  "count": 10,
  "min_frequency": 2,
  "include_phrases": true
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "keywords": [
      {
        "word": "人工智能",
        "frequency": 15,
        "relevance": 0.95
      },
      {
        "word": "机器学习",
        "frequency": 12,
        "relevance": 0.88
      }
    ],
    "phrases": [
      {
        "phrase": "深度学习算法",
        "frequency": 8,
        "relevance": 0.92
      }
    ],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 文档问答

```http
POST /documents/{document_id}/qa
```

**请求体**
```json
{
  "question": "文档中提到了哪些AI应用场景？",
  "context_length": 500
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "answer": "文档中提到了以下AI应用场景：1. 医疗诊断 2. 自动驾驶 3. 智能客服...",
    "confidence": 0.92,
    "sources": [
      {
        "page": 3,
        "content": "在医疗诊断领域，AI技术已经..."
      },
      {
        "page": 7,
        "content": "自动驾驶技术的发展..."
      }
    ],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 3. 文档管理

#### 获取文档列表

```http
GET /documents
```

**查询参数**
- `page` (integer, 可选): 页码，默认1
- `limit` (integer, 可选): 每页数量，默认20
- `type` (string, 可选): 文档类型过滤
- `status` (string, 可选): 处理状态过滤

**响应**
```json
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "doc_1234567890",
        "name": "我的文档.pdf",
        "type": "pdf",
        "size": 2048576,
        "pages": 10,
        "status": "completed",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:35:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1,
      "pages": 1
    }
  }
}
```

#### 获取文档详情

```http
GET /documents/{document_id}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "doc_1234567890",
    "name": "我的文档.pdf",
    "description": "这是一份重要的技术文档",
    "type": "pdf",
    "size": 2048576,
    "pages": 10,
    "status": "completed",
    "url": "https://cdn.taishanglaojun.ai/documents/doc_1234567890.pdf",
    "metadata": {
      "author": "张三",
      "created_date": "2024-01-10",
      "language": "zh"
    },
    "analysis": {
      "summary_available": true,
      "keywords_available": true,
      "qa_available": true
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

#### 删除文档

```http
DELETE /documents/{document_id}
```

**响应**
```json
{
  "success": true,
  "message": "文档已删除"
}
```

## 用户管理API

### 1. 用户信息

#### 获取当前用户信息

```http
GET /users/me
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "user_1234567890",
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "avatar": "https://cdn.taishanglaojun.ai/avatars/user_1234567890.jpg",
    "plan": "pro",
    "created_at": "2024-01-01T00:00:00Z",
    "last_login": "2024-01-15T10:30:00Z",
    "preferences": {
      "language": "zh",
      "theme": "light",
      "notifications": true
    }
  }
}
```

#### 更新用户信息

```http
PUT /users/me
```

**请求体**
```json
{
  "username": "zhangsan_new",
  "avatar": "https://example.com/new-avatar.jpg",
  "preferences": {
    "language": "en",
    "theme": "dark",
    "notifications": false
  }
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "user_1234567890",
    "username": "zhangsan_new",
    "email": "zhangsan@example.com",
    "avatar": "https://example.com/new-avatar.jpg",
    "preferences": {
      "language": "en",
      "theme": "dark",
      "notifications": false
    },
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. 使用统计

#### 获取使用统计

```http
GET /users/me/usage
```

**查询参数**
- `period` (string, 可选): 统计周期，`day`, `week`, `month`

**响应**
```json
{
  "success": true,
  "data": {
    "period": "month",
    "chat": {
      "messages_sent": 150,
      "tokens_used": 45000,
      "conversations": 12
    },
    "images": {
      "generated": 25,
      "edited": 5,
      "variations": 8
    },
    "documents": {
      "uploaded": 3,
      "analyzed": 3,
      "storage_used": 15728640
    },
    "api_calls": {
      "total": 200,
      "remaining": 800
    },
    "limits": {
      "messages_per_month": 1000,
      "images_per_month": 100,
      "storage_limit": 1073741824,
      "api_calls_per_month": 1000
    }
  }
}
```

## 错误代码

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 429 | 请求频率超限 |
| 500 | 服务器内部错误 |

### 业务错误代码

| 错误代码 | 说明 |
|----------|------|
| INVALID_API_KEY | API密钥无效 |
| INSUFFICIENT_PERMISSIONS | 权限不足 |
| RATE_LIMIT_EXCEEDED | 请求频率超限 |
| QUOTA_EXCEEDED | 配额已用完 |
| INVALID_REQUEST | 请求参数无效 |
| RESOURCE_NOT_FOUND | 资源不存在 |
| PROCESSING_ERROR | 处理错误 |
| UNSUPPORTED_FORMAT | 不支持的格式 |
| FILE_TOO_LARGE | 文件过大 |
| CONTENT_FILTERED | 内容被过滤 |

## 速率限制

### 限制规则

| 端点类型 | 限制 |
|----------|------|
| 认证相关 | 10 请求/分钟 |
| 聊天消息 | 60 请求/分钟 |
| 图像生成 | 20 请求/分钟 |
| 文档上传 | 10 请求/分钟 |
| 其他API | 100 请求/分钟 |

### 限制响应头

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1642248000
```

### 超限响应

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求频率超限，请稍后重试",
    "retry_after": 60
  }
}
```

## SDK和示例

### JavaScript SDK

#### 安装

```bash
npm install @taishanglaojun/api-client
```

#### 使用示例

```javascript
import { TaishanglaojunClient } from '@taishanglaojun/api-client';

const client = new TaishanglaojunClient({
  apiKey: 'your-api-key',
  baseURL: 'https://api.taishanglaojun.ai/v1'
});

// 发送聊天消息
const response = await client.chat.sendMessage({
  content: '你好，世界！',
  model: 'gpt-4'
});

console.log(response.data.content);

// 生成图像
const imageResponse = await client.images.generate({
  prompt: '一只可爱的小猫',
  size: '1024x1024'
});

console.log(imageResponse.data.images[0].url);
```

### Python SDK

#### 安装

```bash
pip install taishanglaojun-api
```

#### 使用示例

```python
from taishanglaojun import TaishanglaojunClient

client = TaishanglaojunClient(
    api_key="your-api-key",
    base_url="https://api.taishanglaojun.ai/v1"
)

# 发送聊天消息
response = client.chat.send_message(
    content="你好，世界！",
    model="gpt-4"
)

print(response.data.content)

# 生成图像
image_response = client.images.generate(
    prompt="一只可爱的小猫",
    size="1024x1024"
)

print(image_response.data.images[0].url)
```

### cURL示例

#### 发送聊天消息

```bash
curl -X POST https://api.taishanglaojun.ai/v1/chat/messages \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你好，世界！",
    "model": "gpt-4"
  }'
```

#### 生成图像

```bash
curl -X POST https://api.taishanglaojun.ai/v1/images/generate \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫",
    "size": "1024x1024"
  }'
```

#### 上传文档

```bash
curl -X POST https://api.taishanglaojun.ai/v1/documents/upload \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@document.pdf" \
  -F "name=我的文档.pdf"
```

## Webhook

### 配置Webhook

```http
POST /webhooks
```

**请求体**
```json
{
  "url": "https://your-app.com/webhook",
  "events": ["chat.message.completed", "image.generation.completed"],
  "secret": "your-webhook-secret"
}
```

### Webhook事件

#### 聊天消息完成

```json
{
  "event": "chat.message.completed",
  "data": {
    "message_id": "msg_1234567890",
    "conversation_id": "conv_1234567890",
    "content": "这是AI的回复",
    "tokens_used": 156
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### 图像生成完成

```json
{
  "event": "image.generation.completed",
  "data": {
    "image_id": "img_1234567890",
    "url": "https://cdn.taishanglaojun.ai/images/img_1234567890.png",
    "prompt": "一只可爱的小猫"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 验证Webhook

```javascript
const crypto = require('crypto');

function verifyWebhook(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return signature === `sha256=${expectedSignature}`;
}
```

## 更新日志

### v1.2.0 (2024-01-15)

- 新增文档问答功能
- 优化图像生成质量
- 增加Webhook支持
- 修复已知问题

### v1.1.0 (2024-01-01)

- 新增流式聊天响应
- 增加图像编辑功能
- 优化API响应速度
- 增加更多错误代码

### v1.0.0 (2023-12-01)

- 初始版本发布
- 基础聊天功能
- 图像生成功能
- 文档分析功能

---

© 2024 太上老君AI平台. 保留所有权利。