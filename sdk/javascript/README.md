# 太上老君AI平台 JavaScript SDK

[![npm version](https://badge.fury.io/js/%40taishanglaojun%2Fsdk.svg)](https://badge.fury.io/js/%40taishanglaojun%2Fsdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/%3C%2F%3E-TypeScript-%230074c1.svg)](http://www.typescriptlang.org/)

太上老君AI平台的官方JavaScript/Node.js SDK，提供简单易用的API接口，让您轻松集成AI功能到您的应用中。

## 特性

- 🚀 **简单易用** - 直观的API设计，快速上手
- 📝 **TypeScript支持** - 完整的类型定义，更好的开发体验
- 🔄 **流式响应** - 支持实时流式AI对话
- 🔐 **安全认证** - 多种认证方式，保障API安全
- 📦 **轻量级** - 最小化依赖，减少包体积
- 🌐 **跨平台** - 支持Node.js和现代浏览器
- ⚡ **异步支持** - 基于Promise的异步API
- 🔧 **可配置** - 灵活的配置选项

## 安装

### npm
```bash
npm install @taishanglaojun/sdk
```

### yarn
```bash
yarn add @taishanglaojun/sdk
```

### pnpm
```bash
pnpm add @taishanglaojun/sdk
```

## 快速开始

### 1. 获取API密钥

首先，您需要在[太上老君AI平台](https://taishanglaojun.com)注册账号并获取API密钥。

### 2. 初始化客户端

```javascript
import { TaiShangLaoJunAPI } from '@taishanglaojun/sdk';

const api = new TaiShangLaoJunAPI({
  apiKey: 'your-api-key-here',
  baseURL: 'https://api.taishanglaojun.com/v1' // 可选，默认值
});
```

### 3. 发送第一条消息

```javascript
async function chat() {
  try {
    const response = await api.chat.send({
      message: '你好，太上老君！',
      model: 'taishanglaojun-v1'
    });
    
    console.log(response.message);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

chat();
```

## 基础用法

### 聊天对话

#### 基础对话
```javascript
const response = await api.chat.send({
  message: '请介绍一下人工智能',
  model: 'taishanglaojun-v1',
  temperature: 0.7,
  max_tokens: 1000
});

console.log(response.message);
```

#### 流式对话
```javascript
const stream = api.chat.stream({
  message: '请写一首关于AI的诗',
  model: 'taishanglaojun-v1'
});

for await (const chunk of stream) {
  process.stdout.write(chunk.content);
}
```

#### 维持对话上下文
```javascript
let conversationId = null;

// 第一条消息
const response1 = await api.chat.send({
  message: '我的名字是张三',
  model: 'taishanglaojun-v1'
});
conversationId = response1.conversation_id;

// 第二条消息，AI会记住之前的对话
const response2 = await api.chat.send({
  message: '我的名字是什么？',
  conversation_id: conversationId,
  model: 'taishanglaojun-v1'
});

console.log(response2.message); // AI会回答"您的名字是张三"
```

### 知识库管理

#### 创建知识库
```javascript
const knowledgeBase = await api.knowledge.create({
  name: '产品文档',
  description: '公司产品相关文档',
  type: 'document'
});
```

#### 上传文档
```javascript
import fs from 'fs';

const document = await api.knowledge.uploadDocument({
  knowledge_base_id: knowledgeBase.id,
  file: fs.createReadStream('./document.pdf'),
  metadata: {
    title: '产品使用指南',
    category: 'documentation'
  }
});
```

#### 搜索知识库
```javascript
const results = await api.knowledge.search({
  query: '如何使用AI功能',
  knowledge_base_id: knowledgeBase.id,
  limit: 5,
  threshold: 0.7
});

results.forEach(result => {
  console.log(`标题: ${result.metadata.title}`);
  console.log(`内容: ${result.content}`);
  console.log(`相似度: ${result.score}`);
});
```

### 用户管理

#### 创建用户
```javascript
const user = await api.users.create({
  email: 'user@example.com',
  name: '张三',
  role: 'user',
  metadata: {
    department: '技术部'
  }
});
```

#### 获取用户信息
```javascript
const user = await api.users.get('user-id');
console.log(user);
```

#### 更新用户信息
```javascript
const updatedUser = await api.users.update('user-id', {
  name: '李四',
  metadata: {
    department: '产品部'
  }
});
```

### API密钥管理

#### 创建API密钥
```javascript
const apiKey = await api.apiKeys.create({
  name: '生产环境密钥',
  description: '用于生产环境的API密钥',
  permissions: ['chat:read', 'chat:write'],
  expires_at: new Date('2024-12-31')
});
```

#### 列出API密钥
```javascript
const apiKeys = await api.apiKeys.list({
  page: 1,
  limit: 10
});
```

## 高级功能

### 错误处理

```javascript
import { TaiShangLaoJunError } from '@taishanglaojun/sdk';

try {
  const response = await api.chat.send({
    message: 'Hello'
  });
} catch (error) {
  if (error instanceof TaiShangLaoJunError) {
    switch (error.code) {
      case 'UNAUTHORIZED':
        console.log('API密钥无效');
        break;
      case 'RATE_LIMITED':
        console.log('请求过于频繁');
        break;
      case 'VALIDATION_ERROR':
        console.log('请求参数错误');
        break;
      default:
        console.log('未知错误:', error.message);
    }
  } else {
    console.log('网络错误:', error.message);
  }
}
```

### 重试机制

```javascript
async function apiCallWithRetry(apiCall, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await apiCall();
    } catch (error) {
      if (error.code === 'RATE_LIMITED' && i < maxRetries - 1) {
        const delay = Math.pow(2, i) * 1000; // 指数退避
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }
      throw error;
    }
  }
}

const result = await apiCallWithRetry(() => 
  api.chat.send({ message: 'Hello' })
);
```

### 批量操作

```javascript
const messages = [
  { message: '什么是AI？' },
  { message: '什么是机器学习？' },
  { message: '什么是深度学习？' }
];

const responses = await api.chat.batch({
  messages,
  model: 'taishanglaojun-v1'
});

responses.forEach((response, index) => {
  console.log(`问题${index + 1}: ${messages[index].message}`);
  console.log(`回答${index + 1}: ${response.message}`);
});
```

### 自定义配置

```javascript
const api = new TaiShangLaoJunAPI({
  apiKey: 'your-api-key',
  baseURL: 'https://api.taishanglaojun.com/v1',
  timeout: 30000, // 30秒超时
  retries: 3, // 自动重试3次
  retryDelay: 1000, // 重试延迟1秒
  headers: {
    'User-Agent': 'MyApp/1.0.0'
  }
});
```

## TypeScript 支持

SDK完全支持TypeScript，提供完整的类型定义：

```typescript
import { TaiShangLaoJunAPI, ChatResponse, KnowledgeBase } from '@taishanglaojun/sdk';

const api = new TaiShangLaoJunAPI({
  apiKey: process.env.TAISHANGLAOJUN_API_KEY!
});

// 类型安全的API调用
const response: ChatResponse = await api.chat.send({
  message: 'Hello',
  model: 'taishanglaojun-v1'
});

const knowledgeBase: KnowledgeBase = await api.knowledge.create({
  name: 'My KB',
  type: 'document'
});
```

## 环境变量

您可以使用环境变量来配置SDK：

```bash
# .env 文件
TAISHANGLAOJUN_API_KEY=your-api-key-here
TAISHANGLAOJUN_BASE_URL=https://api.taishanglaojun.com/v1
TAISHANGLAOJUN_TIMEOUT=30000
```

```javascript
// 自动从环境变量读取配置
const api = new TaiShangLaoJunAPI();
```

## 示例项目

查看 `examples/` 目录中的完整示例：

- [基础用法](./examples/basic-usage.js) - 展示基本API调用
- [聊天机器人](./examples/chatbot.js) - 构建简单的聊天机器人
- [知识库搜索](./examples/knowledge-search.js) - 知识库搜索功能
- [Webhook服务器](./examples/webhook-server.js) - 处理Webhook事件

## API 参考

### 聊天 API

#### `api.chat.send(options)`

发送聊天消息。

**参数:**
- `message` (string, 必需) - 用户消息
- `model` (string, 可选) - 使用的模型，默认 'taishanglaojun-v1'
- `conversation_id` (string, 可选) - 对话ID，用于维持上下文
- `user_id` (string, 可选) - 用户ID
- `temperature` (number, 可选) - 温度参数，0-1之间
- `max_tokens` (number, 可选) - 最大token数
- `stream` (boolean, 可选) - 是否流式响应

**返回:**
```typescript
{
  message: string;
  conversation_id: string;
  model: string;
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}
```

#### `api.chat.stream(options)`

流式聊天响应。

**参数:** 同 `api.chat.send()`

**返回:** AsyncIterable<ChatChunk>

### 知识库 API

#### `api.knowledge.create(options)`

创建知识库。

**参数:**
- `name` (string, 必需) - 知识库名称
- `description` (string, 可选) - 描述
- `type` (string, 必需) - 类型: 'document' | 'qa' | 'code'

#### `api.knowledge.search(options)`

搜索知识库。

**参数:**
- `query` (string, 必需) - 搜索查询
- `knowledge_base_id` (string, 必需) - 知识库ID
- `limit` (number, 可选) - 结果数量限制
- `threshold` (number, 可选) - 相似度阈值

## 错误代码

| 错误代码 | 描述 | 处理建议 |
|---------|------|----------|
| `UNAUTHORIZED` | API密钥无效或过期 | 检查API密钥 |
| `RATE_LIMITED` | 请求频率超限 | 实现重试机制 |
| `VALIDATION_ERROR` | 请求参数错误 | 检查参数格式 |
| `NOT_FOUND` | 资源不存在 | 检查资源ID |
| `INTERNAL_ERROR` | 服务器内部错误 | 联系技术支持 |

## 贡献

我们欢迎社区贡献！请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解如何参与开发。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](./LICENSE) 文件。

## 支持

- 📖 [官方文档](https://docs.taishanglaojun.com)
- 💬 [社区论坛](https://community.taishanglaojun.com)
- 📧 [技术支持](mailto:support@taishanglaojun.com)
- 🐛 [问题反馈](https://github.com/taishanglaojun/javascript-sdk/issues)

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md) 了解版本更新历史。