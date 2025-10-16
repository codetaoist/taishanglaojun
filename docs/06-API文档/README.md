# 太上老君AI平台 - API文档

## 📋 文档概览

本目录包含太上老君AI平台的完整API文档，为开发者提供详细的接口说明、参数定义和使用示例。

## 📚 文档结构

### 📖 主要文档
- **[API概览](./api-overview.md)** - API架构和基础信息
- **[API参考手册](./API参考手册.md)** - 完整的API接口参考
- **[认证授权](./认证授权.md)** - 认证和授权机制
- **[错误码定义](./错误码定义.md)** - 错误代码和处理方式
- **[API变更记录](./API变更记录.md)** - API版本变更历史

### 🔧 核心服务API
- **[AI服务API](./核心服务API/ai-service-api.md)** - AI智能服务接口
- **[数据服务API](./核心服务API/data-service-api.md)** - 数据管理服务接口
- **[学习服务API](./核心服务API/learning-service-api.md)** - 智能学习服务接口
- **[系统服务API](./核心服务API/system-service-api.md)** - 系统管理服务接口
- **[用户服务API](./核心服务API/user-service-api.md)** - 用户管理服务接口

## 🚀 快速开始

### 1. 获取API密钥
```bash
# 注册账户并获取API密钥
curl -X POST "https://api.taishanglaojun.com/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email": "your@email.com", "password": "your_password"}'
```

### 2. 基础API调用
```bash
# 智能对话示例
curl -X POST "https://api.taishanglaojun.com/v1/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

### 3. SDK使用
```javascript
import { TaishanglaojunAI } from '@taishanglaojun/sdk';

const client = new TaishanglaojunAI({
  apiKey: 'YOUR_API_KEY'
});

const response = await client.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: '你好' }]
});
```

## 📊 API状态

| 服务 | 状态 | 版本 | 最后更新 |
|------|------|------|----------|
| AI服务 | ✅ 稳定 | v1.2 | 2024-12 |
| 数据服务 | ✅ 稳定 | v1.1 | 2024-12 |
| 学习服务 | ✅ 稳定 | v1.0 | 2024-12 |
| 系统服务 | ✅ 稳定 | v1.0 | 2024-12 |
| 用户服务 | ✅ 稳定 | v1.1 | 2024-12 |

## 🔗 相关链接

- **[项目概览](../00-项目概览/README.md)** - 平台整体介绍
- **[架构设计](../02-架构设计/README.md)** - 技术架构详解
- **[核心服务](../03-核心服务/README.md)** - 后端服务文档
- **[开发指南](../07-开发指南/README.md)** - 开发规范和指南
- **[部署运维](../08-部署运维/README.md)** - 部署和运维指南

## 📞 技术支持

- **API问题**: api-support@taishanglaojun.com
- **技术文档**: docs@taishanglaojun.com
- **GitHub Issues**: [提交问题](https://github.com/taishanglaojun/issues)

---

**最后更新**: 2024年12月  
**文档版本**: v1.0  
**维护团队**: API开发组