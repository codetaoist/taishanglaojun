# 第三方集成API文档

## 概述

第三方集成API提供了完整的开发者平台功能，包括API密钥管理、插件系统、服务集成、Webhook支持和OAuth认证等功能。

## 基础信息

- **基础URL**: `https://api.taishanglaojun.com/v1`
- **认证方式**: Bearer Token 或 API Key
- **数据格式**: JSON
- **字符编码**: UTF-8

## 认证

### API密钥认证
```http
Authorization: Bearer YOUR_API_KEY
```

### OAuth认证
```http
Authorization: Bearer YOUR_ACCESS_TOKEN
```

## API端点

### 1. API密钥管理

#### 1.1 创建API密钥
```http
POST /api-keys
```

**请求体**:
```json
{
  "name": "我的API密钥",
  "description": "用于第三方应用集成",
  "permissions": ["read", "write"],
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**响应**:
```json
{
  "id": "ak_1234567890",
  "name": "我的API密钥",
  "key": "sk_live_1234567890abcdef",
  "prefix": "sk_live_1234",
  "permissions": ["read", "write"],
  "created_at": "2024-01-01T00:00:00Z",
  "expires_at": "2024-12-31T23:59:59Z",
  "last_used_at": null,
  "usage_count": 0
}
```

#### 1.2 获取API密钥列表
```http
GET /api-keys?page=1&limit=20
```

**响应**:
```json
{
  "data": [
    {
      "id": "ak_1234567890",
      "name": "我的API密钥",
      "prefix": "sk_live_1234",
      "permissions": ["read", "write"],
      "created_at": "2024-01-01T00:00:00Z",
      "expires_at": "2024-12-31T23:59:59Z",
      "last_used_at": "2024-01-15T10:30:00Z",
      "usage_count": 150
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

#### 1.3 获取API密钥详情
```http
GET /api-keys/{id}
```

#### 1.4 更新API密钥
```http
PUT /api-keys/{id}
```

#### 1.5 删除API密钥
```http
DELETE /api-keys/{id}
```

#### 1.6 获取使用统计
```http
GET /api-keys/{id}/usage
```

### 2. 插件管理

#### 2.1 安装插件
```http
POST /plugins/install
```

**请求体**:
```json
{
  "plugin_id": "plugin_weather_v1.0.0",
  "source": "marketplace",
  "config": {
    "api_key": "your_weather_api_key",
    "default_city": "Beijing"
  }
}
```

#### 2.2 获取插件列表
```http
GET /plugins?status=active&page=1&limit=20
```

**响应**:
```json
{
  "data": [
    {
      "id": "plugin_weather_v1.0.0",
      "name": "天气插件",
      "description": "获取实时天气信息",
      "version": "1.0.0",
      "author": "WeatherCorp",
      "status": "active",
      "installed_at": "2024-01-01T00:00:00Z",
      "config": {
        "api_key": "***",
        "default_city": "Beijing"
      }
    }
  ]
}
```

#### 2.3 启用/禁用插件
```http
PATCH /plugins/{id}/status
```

**请求体**:
```json
{
  "status": "active" // 或 "inactive"
}
```

#### 2.4 卸载插件
```http
DELETE /plugins/{id}
```

#### 2.5 更新插件配置
```http
PUT /plugins/{id}/config
```

### 3. 服务集成

#### 3.1 创建集成
```http
POST /integrations
```

**请求体**:
```json
{
  "name": "Slack集成",
  "service_type": "slack",
  "config": {
    "webhook_url": "https://hooks.slack.com/services/...",
    "channel": "#general",
    "username": "TaiShangLaoJun"
  },
  "settings": {
    "auto_sync": true,
    "sync_interval": 300
  }
}
```

#### 3.2 获取集成列表
```http
GET /integrations?service_type=slack&status=active
```

#### 3.3 测试集成连接
```http
POST /integrations/{id}/test
```

#### 3.4 同步集成数据
```http
POST /integrations/{id}/sync
```

#### 3.5 更新集成配置
```http
PUT /integrations/{id}
```

#### 3.6 删除集成
```http
DELETE /integrations/{id}
```

### 4. Webhook管理

#### 4.1 创建Webhook
```http
POST /webhooks
```

**请求体**:
```json
{
  "url": "https://your-app.com/webhook",
  "events": ["user.created", "user.updated"],
  "secret": "your_webhook_secret",
  "headers": {
    "X-Custom-Header": "value"
  },
  "active": true
}
```

#### 4.2 获取Webhook列表
```http
GET /webhooks
```

#### 4.3 测试Webhook
```http
POST /webhooks/{id}/test
```

#### 4.4 获取Webhook日志
```http
GET /webhooks/{id}/logs?page=1&limit=50
```

**响应**:
```json
{
  "data": [
    {
      "id": "log_1234567890",
      "webhook_id": "wh_1234567890",
      "event": "user.created",
      "status": "success",
      "response_code": 200,
      "response_time": 150,
      "created_at": "2024-01-15T10:30:00Z",
      "payload": {
        "event": "user.created",
        "data": {
          "user_id": "user_123",
          "email": "user@example.com"
        }
      }
    }
  ]
}
```

#### 4.5 重试失败的Webhook
```http
POST /webhooks/{id}/retry
```

### 5. OAuth应用管理

#### 5.1 创建OAuth应用
```http
POST /oauth/apps
```

**请求体**:
```json
{
  "name": "我的应用",
  "description": "第三方应用集成",
  "redirect_uris": ["https://myapp.com/callback"],
  "scopes": ["read", "write"]
}
```

**响应**:
```json
{
  "id": "app_1234567890",
  "name": "我的应用",
  "client_id": "client_1234567890",
  "client_secret": "secret_abcdef1234567890",
  "redirect_uris": ["https://myapp.com/callback"],
  "scopes": ["read", "write"],
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### 5.2 获取授权URL
```http
GET /oauth/authorize?client_id={client_id}&redirect_uri={redirect_uri}&scope={scope}&state={state}
```

#### 5.3 交换授权码获取令牌
```http
POST /oauth/token
```

**请求体**:
```json
{
  "grant_type": "authorization_code",
  "client_id": "client_1234567890",
  "client_secret": "secret_abcdef1234567890",
  "code": "auth_code_123",
  "redirect_uri": "https://myapp.com/callback"
}
```

**响应**:
```json
{
  "access_token": "access_token_123",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_456",
  "scope": "read write"
}
```

#### 5.4 刷新令牌
```http
POST /oauth/token
```

**请求体**:
```json
{
  "grant_type": "refresh_token",
  "client_id": "client_1234567890",
  "client_secret": "secret_abcdef1234567890",
  "refresh_token": "refresh_token_456"
}
```

#### 5.5 撤销令牌
```http
POST /oauth/revoke
```

## 错误处理

### 错误响应格式
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求参数无效",
    "details": {
      "field": "email",
      "reason": "格式不正确"
    }
  }
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| INVALID_REQUEST | 400 | 请求参数无效 |
| UNAUTHORIZED | 401 | 未授权访问 |
| FORBIDDEN | 403 | 权限不足 |
| NOT_FOUND | 404 | 资源不存在 |
| RATE_LIMITED | 429 | 请求频率超限 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |

## 速率限制

- **API密钥**: 每分钟1000次请求
- **OAuth令牌**: 每分钟500次请求
- **Webhook**: 每分钟100次请求

## Webhook事件

### 支持的事件类型

| 事件 | 描述 |
|------|------|
| user.created | 用户创建 |
| user.updated | 用户更新 |
| user.deleted | 用户删除 |
| api_key.created | API密钥创建 |
| api_key.deleted | API密钥删除 |
| plugin.installed | 插件安装 |
| plugin.uninstalled | 插件卸载 |
| integration.created | 集成创建 |
| integration.updated | 集成更新 |

### Webhook负载格式
```json
{
  "event": "user.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "user_id": "user_123",
    "email": "user@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

## SDK和示例

### JavaScript SDK
```javascript
import { TaiShangLaoJunAPI } from '@taishanglaojun/sdk';

const api = new TaiShangLaoJunAPI({
  apiKey: 'your_api_key',
  baseURL: 'https://api.taishanglaojun.com/v1'
});

// 创建API密钥
const apiKey = await api.apiKeys.create({
  name: '我的API密钥',
  permissions: ['read', 'write']
});

// 安装插件
const plugin = await api.plugins.install({
  plugin_id: 'plugin_weather_v1.0.0',
  config: { api_key: 'weather_api_key' }
});
```

### Python SDK
```python
from taishanglaojun import TaiShangLaoJunAPI

api = TaiShangLaoJunAPI(
    api_key='your_api_key',
    base_url='https://api.taishanglaojun.com/v1'
)

# 创建集成
integration = api.integrations.create(
    name='Slack集成',
    service_type='slack',
    config={
        'webhook_url': 'https://hooks.slack.com/services/...',
        'channel': '#general'
    }
)
```

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 支持API密钥管理
- 支持插件系统
- 支持服务集成
- 支持Webhook
- 支持OAuth认证