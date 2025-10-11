# 第三方集成开发指南

## 概述

太上老君AI平台提供了强大的第三方集成能力，让开发者可以轻松地将平台功能集成到自己的应用中，或者为平台开发插件和扩展。

## 快速开始

### 1. 获取API密钥

首先，您需要在太上老君AI平台中创建一个API密钥：

1. 登录太上老君AI平台
2. 进入"第三方集成" -> "API密钥管理"
3. 点击"创建API密钥"
4. 填写密钥信息并设置权限
5. 保存生成的API密钥

### 2. 安装SDK

#### JavaScript/Node.js
```bash
npm install @taishanglaojun/sdk
```

#### Python
```bash
pip install taishanglaojun-sdk
```

#### Go
```bash
go get github.com/taishanglaojun/go-sdk
```

### 3. 初始化客户端

#### JavaScript
```javascript
import { TaiShangLaoJunAPI } from '@taishanglaojun/sdk';

const api = new TaiShangLaoJunAPI({
  apiKey: process.env.TAISHANGLAOJUN_API_KEY,
  baseURL: 'https://api.taishanglaojun.com/v1'
});
```

#### Python
```python
import os
from taishanglaojun import TaiShangLaoJunAPI

api = TaiShangLaoJunAPI(
    api_key=os.getenv('TAISHANGLAOJUN_API_KEY'),
    base_url='https://api.taishanglaojun.com/v1'
)
```

#### Go
```go
package main

import (
    "os"
    "github.com/taishanglaojun/go-sdk"
)

func main() {
    client := taishanglaojun.NewClient(&taishanglaojun.Config{
        APIKey:  os.Getenv("TAISHANGLAOJUN_API_KEY"),
        BaseURL: "https://api.taishanglaojun.com/v1",
    })
}
```

## 核心功能集成

### 1. AI对话集成

#### 基础对话
```javascript
// 发送消息到AI
const response = await api.chat.send({
  message: "你好，请介绍一下太上老君AI平台",
  conversation_id: "conv_123", // 可选，用于维持对话上下文
  model: "taishanglaojun-v1" // 可选，指定模型
});

console.log(response.message); // AI的回复
console.log(response.conversation_id); // 对话ID
```

#### 流式对话
```javascript
const stream = api.chat.stream({
  message: "请写一首关于人工智能的诗",
  conversation_id: "conv_123"
});

for await (const chunk of stream) {
  process.stdout.write(chunk.content);
}
```

### 2. 知识库集成

#### 创建知识库
```javascript
const knowledgeBase = await api.knowledge.create({
  name: "产品文档",
  description: "公司产品相关文档",
  type: "document"
});
```

#### 上传文档
```javascript
const document = await api.knowledge.uploadDocument({
  knowledge_base_id: knowledgeBase.id,
  file: fs.createReadStream('./product-guide.pdf'),
  metadata: {
    title: "产品使用指南",
    category: "documentation"
  }
});
```

#### 知识检索
```javascript
const results = await api.knowledge.search({
  query: "如何使用AI对话功能",
  knowledge_base_id: knowledgeBase.id,
  limit: 5
});
```

### 3. 用户管理集成

#### 创建用户
```javascript
const user = await api.users.create({
  email: "user@example.com",
  name: "张三",
  role: "user",
  metadata: {
    department: "技术部",
    position: "工程师"
  }
});
```

#### 用户认证
```javascript
const authResult = await api.auth.authenticate({
  email: "user@example.com",
  password: "password123"
});

if (authResult.success) {
  console.log("认证成功", authResult.token);
}
```

## 插件开发

### 1. 插件结构

```
my-plugin/
├── plugin.json          # 插件配置文件
├── index.js             # 插件入口文件
├── package.json         # 依赖配置
├── README.md           # 插件说明
└── assets/             # 静态资源
    ├── icon.png
    └── screenshot.png
```

### 2. 插件配置文件 (plugin.json)

```json
{
  "id": "my-weather-plugin",
  "name": "天气插件",
  "version": "1.0.0",
  "description": "获取实时天气信息",
  "author": "Your Name",
  "homepage": "https://github.com/yourname/weather-plugin",
  "main": "index.js",
  "permissions": ["network", "storage"],
  "config_schema": {
    "type": "object",
    "properties": {
      "api_key": {
        "type": "string",
        "title": "API密钥",
        "description": "天气API的密钥"
      },
      "default_city": {
        "type": "string",
        "title": "默认城市",
        "default": "北京"
      }
    },
    "required": ["api_key"]
  },
  "commands": [
    {
      "name": "weather",
      "description": "获取天气信息",
      "parameters": {
        "city": {
          "type": "string",
          "description": "城市名称"
        }
      }
    }
  ]
}
```

### 3. 插件入口文件 (index.js)

```javascript
class WeatherPlugin {
  constructor(config, api) {
    this.config = config;
    this.api = api;
  }

  // 插件初始化
  async initialize() {
    console.log('天气插件初始化完成');
  }

  // 处理命令
  async handleCommand(command, parameters) {
    switch (command) {
      case 'weather':
        return await this.getWeather(parameters.city || this.config.default_city);
      default:
        throw new Error(`未知命令: ${command}`);
    }
  }

  // 获取天气信息
  async getWeather(city) {
    try {
      const response = await fetch(
        `https://api.weather.com/v1/current?key=${this.config.api_key}&city=${city}`
      );
      const data = await response.json();
      
      return {
        type: 'weather',
        data: {
          city: city,
          temperature: data.temperature,
          description: data.description,
          humidity: data.humidity
        }
      };
    } catch (error) {
      throw new Error(`获取天气信息失败: ${error.message}`);
    }
  }

  // 插件卸载
  async cleanup() {
    console.log('天气插件清理完成');
  }
}

module.exports = WeatherPlugin;
```

### 4. 插件测试

```javascript
// test/weather-plugin.test.js
const WeatherPlugin = require('../index');

describe('WeatherPlugin', () => {
  let plugin;

  beforeEach(() => {
    plugin = new WeatherPlugin({
      api_key: 'test_api_key',
      default_city: '北京'
    });
  });

  test('should get weather for default city', async () => {
    const result = await plugin.handleCommand('weather', {});
    expect(result.type).toBe('weather');
    expect(result.data.city).toBe('北京');
  });

  test('should get weather for specified city', async () => {
    const result = await plugin.handleCommand('weather', { city: '上海' });
    expect(result.data.city).toBe('上海');
  });
});
```

## Webhook集成

### 1. 设置Webhook端点

```javascript
// Express.js 示例
const express = require('express');
const crypto = require('crypto');
const app = express();

app.use(express.json());

// Webhook处理端点
app.post('/webhook', (req, res) => {
  const signature = req.headers['x-taishanglaojun-signature'];
  const payload = JSON.stringify(req.body);
  
  // 验证签名
  if (!verifySignature(payload, signature, process.env.WEBHOOK_SECRET)) {
    return res.status(401).send('Unauthorized');
  }

  // 处理事件
  handleWebhookEvent(req.body);
  
  res.status(200).send('OK');
});

function verifySignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature, 'hex'),
    Buffer.from(expectedSignature, 'hex')
  );
}

function handleWebhookEvent(event) {
  switch (event.event) {
    case 'user.created':
      console.log('新用户创建:', event.data);
      break;
    case 'chat.message':
      console.log('新消息:', event.data);
      break;
    default:
      console.log('未知事件:', event.event);
  }
}

app.listen(3000, () => {
  console.log('Webhook服务器运行在端口3000');
});
```

### 2. 注册Webhook

```javascript
const webhook = await api.webhooks.create({
  url: 'https://your-app.com/webhook',
  events: ['user.created', 'chat.message'],
  secret: 'your_webhook_secret'
});
```

## OAuth集成

### 1. 创建OAuth应用

```javascript
const oauthApp = await api.oauth.createApp({
  name: '我的应用',
  description: '第三方应用集成',
  redirect_uris: ['https://myapp.com/callback'],
  scopes: ['read', 'write']
});

console.log('Client ID:', oauthApp.client_id);
console.log('Client Secret:', oauthApp.client_secret);
```

### 2. 授权流程

#### 步骤1: 重定向到授权页面
```javascript
const authURL = `https://api.taishanglaojun.com/v1/oauth/authorize?` +
  `client_id=${client_id}&` +
  `redirect_uri=${encodeURIComponent(redirect_uri)}&` +
  `scope=${scopes.join(' ')}&` +
  `state=${state}`;

// 重定向用户到授权页面
window.location.href = authURL;
```

#### 步骤2: 处理回调
```javascript
// 在回调端点处理授权码
app.get('/callback', async (req, res) => {
  const { code, state } = req.query;
  
  // 验证state参数
  if (state !== expectedState) {
    return res.status(400).send('Invalid state');
  }

  // 交换授权码获取访问令牌
  const tokenResponse = await api.oauth.exchangeCode({
    code: code,
    client_id: client_id,
    client_secret: client_secret,
    redirect_uri: redirect_uri
  });

  // 保存访问令牌
  const { access_token, refresh_token } = tokenResponse;
  
  res.send('授权成功！');
});
```

#### 步骤3: 使用访问令牌
```javascript
const authenticatedAPI = new TaiShangLaoJunAPI({
  accessToken: access_token,
  baseURL: 'https://api.taishanglaojun.com/v1'
});

// 现在可以代表用户进行API调用
const userProfile = await authenticatedAPI.users.getProfile();
```

## 最佳实践

### 1. 错误处理

```javascript
try {
  const result = await api.chat.send({
    message: "Hello"
  });
} catch (error) {
  if (error.code === 'RATE_LIMITED') {
    // 处理速率限制
    console.log('请求过于频繁，请稍后重试');
  } else if (error.code === 'UNAUTHORIZED') {
    // 处理认证错误
    console.log('API密钥无效或已过期');
  } else {
    // 处理其他错误
    console.error('API调用失败:', error.message);
  }
}
```

### 2. 重试机制

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

// 使用示例
const result = await apiCallWithRetry(() => 
  api.chat.send({ message: "Hello" })
);
```

### 3. 缓存策略

```javascript
const cache = new Map();

async function getCachedKnowledgeBase(id) {
  const cacheKey = `kb_${id}`;
  
  if (cache.has(cacheKey)) {
    return cache.get(cacheKey);
  }

  const knowledgeBase = await api.knowledge.get(id);
  cache.set(cacheKey, knowledgeBase);
  
  // 设置缓存过期时间
  setTimeout(() => cache.delete(cacheKey), 5 * 60 * 1000); // 5分钟
  
  return knowledgeBase;
}
```

### 4. 安全考虑

- **API密钥安全**: 永远不要在客户端代码中暴露API密钥
- **HTTPS**: 始终使用HTTPS进行API通信
- **签名验证**: 验证Webhook签名以确保请求来源
- **权限最小化**: 只申请必要的API权限
- **令牌刷新**: 及时刷新过期的访问令牌

### 5. 性能优化

- **批量操作**: 尽可能使用批量API减少请求次数
- **分页处理**: 合理设置分页大小避免超时
- **并发控制**: 控制并发请求数量避免触发速率限制
- **连接复用**: 复用HTTP连接减少建立连接的开销

## 示例项目

### 1. 聊天机器人集成

```javascript
// chatbot.js
const { TaiShangLaoJunAPI } = require('@taishanglaojun/sdk');

class ChatBot {
  constructor(apiKey) {
    this.api = new TaiShangLaoJunAPI({ apiKey });
    this.conversations = new Map();
  }

  async handleMessage(userId, message) {
    let conversationId = this.conversations.get(userId);
    
    const response = await this.api.chat.send({
      message,
      conversation_id: conversationId,
      user_id: userId
    });

    this.conversations.set(userId, response.conversation_id);
    return response.message;
  }
}

module.exports = ChatBot;
```

### 2. 知识库搜索服务

```javascript
// knowledge-search.js
class KnowledgeSearchService {
  constructor(api) {
    this.api = api;
  }

  async search(query, options = {}) {
    const {
      knowledge_base_id,
      limit = 10,
      threshold = 0.7
    } = options;

    const results = await this.api.knowledge.search({
      query,
      knowledge_base_id,
      limit,
      threshold
    });

    return results.map(result => ({
      title: result.metadata.title,
      content: result.content,
      score: result.score,
      source: result.metadata.source
    }));
  }
}
```

## 故障排除

### 常见问题

1. **API密钥无效**
   - 检查API密钥是否正确
   - 确认API密钥未过期
   - 验证API密钥权限

2. **请求超时**
   - 检查网络连接
   - 增加请求超时时间
   - 使用重试机制

3. **速率限制**
   - 实现指数退避重试
   - 减少请求频率
   - 考虑升级API计划

4. **Webhook未收到**
   - 检查Webhook URL是否可访问
   - 验证SSL证书
   - 检查防火墙设置

## 社区和支持

- **官方文档**: https://docs.taishanglaojun.com
- **GitHub**: https://github.com/taishanglaojun
- **社区论坛**: https://community.taishanglaojun.com
- **技术支持**: support@taishanglaojun.com

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 支持基础API集成
- 提供JavaScript、Python、Go SDK
- 支持插件开发框架
- 支持Webhook和OAuth集成