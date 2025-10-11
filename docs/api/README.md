# 太上老君AI平台 API 文档

## 概述

太上老君AI平台提供了一套完整的RESTful API，支持用户管理、AI对话、文件处理、支付等核心功能。本文档详细介绍了所有可用的API接口。

## 基础信息

- **基础URL**: `https://api.taishanglaojun.ai`
- **API版本**: v1
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 认证

所有API请求都需要在请求头中包含有效的JWT token：

```http
Authorization: Bearer <your-jwt-token>
```

### 获取Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "name": "用户名",
      "avatar": "https://example.com/avatar.jpg"
    }
  }
}
```

## 错误处理

API使用标准的HTTP状态码来表示请求结果：

- `200` - 成功
- `201` - 创建成功
- `400` - 请求参数错误
- `401` - 未授权
- `403` - 禁止访问
- `404` - 资源不存在
- `429` - 请求频率限制
- `500` - 服务器内部错误

**错误响应格式**:
```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求参数无效",
    "details": {
      "field": "email",
      "reason": "邮箱格式不正确"
    }
  }
}
```

## 分页

对于返回列表数据的API，支持分页查询：

**请求参数**:
- `page`: 页码（从1开始）
- `limit`: 每页数量（默认20，最大100）
- `sort`: 排序字段
- `order`: 排序方向（asc/desc）

**响应格式**:
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "pages": 5
    }
  }
}
```

## API 接口

### 1. 用户认证

#### 1.1 用户注册

```http
POST /api/v1/auth/register
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "用户名",
  "phone": "+86 138 0013 8000"
}
```

#### 1.2 用户登录

```http
POST /api/v1/auth/login
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

#### 1.3 刷新Token

```http
POST /api/v1/auth/refresh
```

#### 1.4 用户登出

```http
POST /api/v1/auth/logout
```

#### 1.5 忘记密码

```http
POST /api/v1/auth/forgot-password
```

**请求体**:
```json
{
  "email": "user@example.com"
}
```

#### 1.6 重置密码

```http
POST /api/v1/auth/reset-password
```

**请求体**:
```json
{
  "token": "reset-token",
  "password": "new-password"
}
```

### 2. 用户管理

#### 2.1 获取用户信息

```http
GET /api/v1/users/profile
```

#### 2.2 更新用户信息

```http
PUT /api/v1/users/profile
```

**请求体**:
```json
{
  "name": "新用户名",
  "avatar": "https://example.com/new-avatar.jpg",
  "bio": "个人简介"
}
```

#### 2.3 修改密码

```http
PUT /api/v1/users/password
```

**请求体**:
```json
{
  "current_password": "current-password",
  "new_password": "new-password"
}
```

#### 2.4 删除账户

```http
DELETE /api/v1/users/account
```

### 3. AI对话

#### 3.1 创建对话

```http
POST /api/v1/conversations
```

**请求体**:
```json
{
  "title": "对话标题",
  "model": "gpt-4",
  "system_prompt": "你是一个有用的AI助手"
}
```

#### 3.2 获取对话列表

```http
GET /api/v1/conversations?page=1&limit=20
```

#### 3.3 获取对话详情

```http
GET /api/v1/conversations/{conversation_id}
```

#### 3.4 发送消息

```http
POST /api/v1/conversations/{conversation_id}/messages
```

**请求体**:
```json
{
  "content": "你好，请帮我写一首诗",
  "type": "text",
  "attachments": [
    {
      "type": "image",
      "url": "https://example.com/image.jpg"
    }
  ]
}
```

#### 3.5 获取消息历史

```http
GET /api/v1/conversations/{conversation_id}/messages?page=1&limit=50
```

#### 3.6 删除对话

```http
DELETE /api/v1/conversations/{conversation_id}
```

### 4. 文件管理

#### 4.1 上传文件

```http
POST /api/v1/files/upload
Content-Type: multipart/form-data
```

**请求参数**:
- `file`: 文件数据
- `type`: 文件类型（image/document/audio/video）
- `folder`: 存储文件夹（可选）

#### 4.2 获取文件列表

```http
GET /api/v1/files?type=image&page=1&limit=20
```

#### 4.3 获取文件信息

```http
GET /api/v1/files/{file_id}
```

#### 4.4 删除文件

```http
DELETE /api/v1/files/{file_id}
```

#### 4.5 生成预签名URL

```http
POST /api/v1/files/presigned-url
```

**请求体**:
```json
{
  "filename": "example.jpg",
  "content_type": "image/jpeg",
  "expires_in": 3600
}
```

### 5. AI模型管理

#### 5.1 获取可用模型

```http
GET /api/v1/models
```

#### 5.2 获取模型详情

```http
GET /api/v1/models/{model_id}
```

#### 5.3 模型使用统计

```http
GET /api/v1/models/usage?start_date=2024-01-01&end_date=2024-01-31
```

### 6. 订阅和支付

#### 6.1 获取订阅计划

```http
GET /api/v1/subscriptions/plans
```

#### 6.2 创建订阅

```http
POST /api/v1/subscriptions
```

**请求体**:
```json
{
  "plan_id": "plan-premium",
  "payment_method": "stripe"
}
```

#### 6.3 获取当前订阅

```http
GET /api/v1/subscriptions/current
```

#### 6.4 取消订阅

```http
DELETE /api/v1/subscriptions/current
```

#### 6.5 获取账单历史

```http
GET /api/v1/billing/invoices?page=1&limit=20
```

### 7. 系统设置

#### 7.1 获取系统配置

```http
GET /api/v1/settings
```

#### 7.2 更新用户设置

```http
PUT /api/v1/settings/user
```

**请求体**:
```json
{
  "language": "zh-CN",
  "theme": "dark",
  "notifications": {
    "email": true,
    "push": false
  }
}
```

### 8. 统计和分析

#### 8.1 获取使用统计

```http
GET /api/v1/analytics/usage?period=30d
```

#### 8.2 获取对话统计

```http
GET /api/v1/analytics/conversations?start_date=2024-01-01&end_date=2024-01-31
```

## WebSocket API

### 实时对话

连接到WebSocket端点进行实时对话：

```
wss://api.taishanglaojun.ai/ws/conversations/{conversation_id}
```

**认证**: 在连接时通过查询参数传递token：
```
wss://api.taishanglaojun.ai/ws/conversations/{conversation_id}?token=your-jwt-token
```

**消息格式**:

发送消息：
```json
{
  "type": "message",
  "data": {
    "content": "你好",
    "attachments": []
  }
}
```

接收消息：
```json
{
  "type": "message",
  "data": {
    "id": "msg-123",
    "content": "你好！有什么我可以帮助你的吗？",
    "role": "assistant",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

## 速率限制

API实施了速率限制以确保服务稳定性：

- **认证接口**: 每分钟5次请求
- **消息发送**: 每分钟60次请求
- **文件上传**: 每分钟10次请求
- **其他接口**: 每分钟100次请求

当达到速率限制时，API将返回429状态码。

## SDK和示例

### JavaScript SDK

```bash
npm install @taishanglaojun/sdk
```

```javascript
import { TaishangLaojunClient } from '@taishanglaojun/sdk';

const client = new TaishangLaojunClient({
  apiKey: 'your-api-key',
  baseURL: 'https://api.taishanglaojun.ai'
});

// 发送消息
const response = await client.conversations.sendMessage('conv-123', {
  content: '你好，世界！'
});
```

### Python SDK

```bash
pip install taishanglaojun-sdk
```

```python
from taishanglaojun import Client

client = Client(api_key='your-api-key')

# 创建对话
conversation = client.conversations.create(
    title='新对话',
    model='gpt-4'
)

# 发送消息
response = client.conversations.send_message(
    conversation.id,
    content='你好，世界！'
)
```

## 更新日志

### v1.2.0 (2024-01-15)
- 新增文件批量上传接口
- 支持WebSocket实时对话
- 优化响应时间

### v1.1.0 (2024-01-01)
- 新增订阅和支付功能
- 支持多模态消息
- 增加使用统计接口

### v1.0.0 (2023-12-01)
- 初始版本发布
- 基础用户管理和AI对话功能

## 支持

如有问题或建议，请联系：

- 邮箱: api-support@taishanglaojun.ai
- 文档: https://docs.taishanglaojun.ai
- GitHub: https://github.com/taishanglaojun/platform