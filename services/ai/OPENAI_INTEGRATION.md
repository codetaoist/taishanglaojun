# OpenAI集成指南

本指南介绍如何在Taishang Laojun AI服务中配置和使用OpenAI模型。

## 配置OpenAI API密钥

1. 复制环境变量示例文件：
```bash
cp .env.example .env
```

2. 编辑`.env`文件，添加您的OpenAI API密钥：
```
OPENAI_API_KEY=your-actual-openai-api-key-here
```

## 支持的OpenAI模型

### 文本生成模型
- `gpt-3.5-turbo` - OpenAI GPT-3.5 Turbo模型
- `gpt-4` - OpenAI GPT-4模型
- `gpt-4-turbo` - OpenAI GPT-4 Turbo模型

### 嵌入模型
- `text-embedding-ada-002` - OpenAI文本嵌入模型
- `text-embedding-3-small` - OpenAI小型文本嵌入模型
- `text-embedding-3-large` - OpenAI大型文本嵌入模型

## 注册和使用OpenAI模型

### 通过API注册模型

#### 注册文本生成模型
```bash
curl -X POST http://localhost:8083/api/v1/model/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-3.5-turbo",
    "provider": "openai",
    "model_path": "gpt-3.5-turbo",
    "model_type": "generation",
    "description": "OpenAI GPT-3.5 Turbo模型",
    "is_default": true
  }'
```

#### 注册嵌入模型
```bash
curl -X POST http://localhost:8083/api/v1/model/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "text-embedding-ada-002",
    "provider": "openai",
    "model_path": "text-embedding-ada-002",
    "model_type": "embedding",
    "description": "OpenAI文本嵌入模型",
    "is_default": true
  }'
```

### 使用模型

#### 文本生成
```bash
curl -X POST http://localhost:8083/api/v1/model/generate/text \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "请简单介绍一下人工智能。",
    "model_name": "gpt-3.5-turbo",
    "max_tokens": 100,
    "temperature": 0.7
  }'
```

#### 文本嵌入
```bash
curl -X POST http://localhost:8083/api/v1/model/generate/embedding \
  -H "Content-Type: application/json" \
  -d '{
    "text": "这是一个测试句子。",
    "model_name": "text-embedding-ada-002"
  }'
```

## 测试OpenAI集成

运行测试脚本以验证OpenAI集成是否正常工作：

```bash
python test_openai_integration.py
```

## 注意事项

1. 确保您的OpenAI API密钥有足够的配额
2. OpenAI API是付费服务，使用会产生费用
3. 请遵守OpenAI的使用条款和政策
4. 在生产环境中，请确保API密钥的安全存储

## 故障排除

### 常见问题

1. **API密钥错误**
   - 确保在`.env`文件中正确设置了`OPENAI_API_KEY`
   - 检查API密钥是否有效且未过期

2. **模型不可用**
   - 确保您选择的模型在OpenAI API中可用
   - 检查您的API密钥是否有访问该模型的权限

3. **配额不足**
   - 检查您的OpenAI账户余额
   - 考虑设置使用限制以避免意外费用

4. **网络连接问题**
   - 确保您的服务器可以访问OpenAI API
   - 检查防火墙和代理设置