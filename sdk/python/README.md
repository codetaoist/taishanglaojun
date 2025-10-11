# 太上老君AI平台 Python SDK

[![PyPI version](https://badge.fury.io/py/taishanglaojun-sdk.svg)](https://badge.fury.io/py/taishanglaojun-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/python-3.7+-blue.svg)](https://www.python.org/downloads/)

太上老君AI平台的官方Python SDK，提供简单易用的API接口，让您轻松集成AI功能到您的Python应用中。

## 特性

- 🚀 **简单易用** - 直观的API设计，快速上手
- 🐍 **纯Python** - 原生Python实现，无需额外依赖
- 🔄 **异步支持** - 支持asyncio异步编程
- 📝 **类型提示** - 完整的类型注解，更好的IDE支持
- 🔄 **流式响应** - 支持实时流式AI对话
- 🔐 **安全认证** - 多种认证方式，保障API安全
- 📦 **轻量级** - 最小化依赖，快速安装
- 🌐 **跨平台** - 支持Windows、macOS、Linux
- 🔧 **可配置** - 灵活的配置选项

## 安装

### pip
```bash
pip install taishanglaojun-sdk
```

### conda
```bash
conda install -c conda-forge taishanglaojun-sdk
```

### 从源码安装
```bash
git clone https://github.com/taishanglaojun/python-sdk.git
cd python-sdk
pip install -e .
```

## 快速开始

### 1. 获取API密钥

首先，您需要在[太上老君AI平台](https://taishanglaojun.com)注册账号并获取API密钥。

### 2. 初始化客户端

```python
from taishanglaojun import TaiShangLaoJunAPI

api = TaiShangLaoJunAPI(
    api_key='your-api-key-here',
    base_url='https://api.taishanglaojun.com/v1'  # 可选，默认值
)
```

### 3. 发送第一条消息

```python
response = api.chat.send(
    message='你好，太上老君！',
    model='taishanglaojun-v1'
)

print(response.message)
```

## 基础用法

### 聊天对话

#### 基础对话
```python
response = api.chat.send(
    message='请介绍一下人工智能',
    model='taishanglaojun-v1',
    temperature=0.7,
    max_tokens=1000
)

print(response.message)
print(f"使用token数: {response.usage.total_tokens}")
```

#### 流式对话
```python
stream = api.chat.stream(
    message='请写一首关于AI的诗',
    model='taishanglaojun-v1'
)

for chunk in stream:
    print(chunk.content, end='', flush=True)
```

#### 维持对话上下文
```python
conversation_id = None

# 第一条消息
response1 = api.chat.send(
    message='我的名字是张三',
    model='taishanglaojun-v1'
)
conversation_id = response1.conversation_id

# 第二条消息，AI会记住之前的对话
response2 = api.chat.send(
    message='我的名字是什么？',
    conversation_id=conversation_id,
    model='taishanglaojun-v1'
)

print(response2.message)  # AI会回答"您的名字是张三"
```

### 异步编程

```python
import asyncio
from taishanglaojun import TaiShangLaoJunAPI

async def main():
    api = TaiShangLaoJunAPI(
        api_key='your-api-key',
        async_mode=True  # 启用异步模式
    )
    
    # 异步聊天
    response = await api.chat.send(
        message='你好！',
        model='taishanglaojun-v1'
    )
    print(response.message)
    
    # 异步流式聊天
    async for chunk in api.chat.stream(message='写一首诗'):
        print(chunk.content, end='', flush=True)

asyncio.run(main())
```

### 知识库管理

#### 创建知识库
```python
knowledge_base = api.knowledge.create(
    name='产品文档',
    description='公司产品相关文档',
    type='document'
)
print(f"知识库ID: {knowledge_base.id}")
```

#### 上传文档
```python
with open('./document.pdf', 'rb') as f:
    document = api.knowledge.upload_document(
        knowledge_base_id=knowledge_base.id,
        file=f,
        metadata={
            'title': '产品使用指南',
            'category': 'documentation'
        }
    )
```

#### 搜索知识库
```python
results = api.knowledge.search(
    query='如何使用AI功能',
    knowledge_base_id=knowledge_base.id,
    limit=5,
    threshold=0.7
)

for result in results:
    print(f"标题: {result.metadata['title']}")
    print(f"内容: {result.content}")
    print(f"相似度: {result.score}")
    print("---")
```

### 用户管理

#### 创建用户
```python
user = api.users.create(
    email='user@example.com',
    name='张三',
    role='user',
    metadata={
        'department': '技术部'
    }
)
print(f"用户ID: {user.id}")
```

#### 获取用户信息
```python
user = api.users.get('user-id')
print(f"用户名: {user.name}")
print(f"邮箱: {user.email}")
```

#### 更新用户信息
```python
updated_user = api.users.update(
    'user-id',
    name='李四',
    metadata={
        'department': '产品部'
    }
)
```

### API密钥管理

#### 创建API密钥
```python
from datetime import datetime, timedelta

api_key = api.api_keys.create(
    name='生产环境密钥',
    description='用于生产环境的API密钥',
    permissions=['chat:read', 'chat:write'],
    expires_at=datetime.now() + timedelta(days=365)
)
print(f"API密钥: {api_key.key}")
```

#### 列出API密钥
```python
api_keys = api.api_keys.list(page=1, limit=10)
for key in api_keys.items:
    print(f"名称: {key.name}, 状态: {key.status}")
```

## 高级功能

### 错误处理

```python
from taishanglaojun import TaiShangLaoJunError

try:
    response = api.chat.send(message='Hello')
except TaiShangLaoJunError as e:
    if e.code == 'UNAUTHORIZED':
        print('API密钥无效')
    elif e.code == 'RATE_LIMITED':
        print('请求过于频繁')
    elif e.code == 'VALIDATION_ERROR':
        print('请求参数错误')
    else:
        print(f'未知错误: {e.message}')
except Exception as e:
    print(f'网络错误: {e}')
```

### 重试机制

```python
import time
from taishanglaojun import TaiShangLaoJunError

def api_call_with_retry(api_call, max_retries=3):
    for i in range(max_retries):
        try:
            return api_call()
        except TaiShangLaoJunError as e:
            if e.code == 'RATE_LIMITED' and i < max_retries - 1:
                delay = 2 ** i  # 指数退避
                time.sleep(delay)
                continue
            raise e

result = api_call_with_retry(
    lambda: api.chat.send(message='Hello')
)
```

### 批量操作

```python
messages = [
    {'message': '什么是AI？'},
    {'message': '什么是机器学习？'},
    {'message': '什么是深度学习？'}
]

responses = api.chat.batch(
    messages=messages,
    model='taishanglaojun-v1'
)

for i, (msg, resp) in enumerate(zip(messages, responses)):
    print(f"问题{i+1}: {msg['message']}")
    print(f"回答{i+1}: {resp.message}")
    print("---")
```

### 自定义配置

```python
api = TaiShangLaoJunAPI(
    api_key='your-api-key',
    base_url='https://api.taishanglaojun.com/v1',
    timeout=30,  # 30秒超时
    retries=3,   # 自动重试3次
    retry_delay=1,  # 重试延迟1秒
    headers={
        'User-Agent': 'MyApp/1.0.0'
    }
)
```

## 环境变量

您可以使用环境变量来配置SDK：

```bash
# .env 文件
TAISHANGLAOJUN_API_KEY=your-api-key-here
TAISHANGLAOJUN_BASE_URL=https://api.taishanglaojun.com/v1
TAISHANGLAOJUN_TIMEOUT=30
```

```python
import os
from taishanglaojun import TaiShangLaoJunAPI

# 自动从环境变量读取配置
api = TaiShangLaoJunAPI(
    api_key=os.getenv('TAISHANGLAOJUN_API_KEY')
)
```

## 示例项目

查看 `examples/` 目录中的完整示例：

- [基础用法](./examples/basic_usage.py) - 展示基本API调用
- [聊天机器人](./examples/chatbot.py) - 构建简单的聊天机器人
- [知识库搜索](./examples/knowledge_search.py) - 知识库搜索功能
- [Flask Web应用](./examples/flask_app.py) - Flask集成示例
- [FastAPI应用](./examples/fastapi_app.py) - FastAPI集成示例
- [Streamlit应用](./examples/streamlit_app.py) - Streamlit集成示例

## 框架集成

### Flask 集成

```python
from flask import Flask, request, jsonify
from taishanglaojun import TaiShangLaoJunAPI

app = Flask(__name__)
api = TaiShangLaoJunAPI(api_key='your-api-key')

@app.route('/chat', methods=['POST'])
def chat():
    data = request.json
    response = api.chat.send(
        message=data['message'],
        model='taishanglaojun-v1'
    )
    return jsonify({'response': response.message})

if __name__ == '__main__':
    app.run(debug=True)
```

### FastAPI 集成

```python
from fastapi import FastAPI
from pydantic import BaseModel
from taishanglaojun import TaiShangLaoJunAPI

app = FastAPI()
api = TaiShangLaoJunAPI(api_key='your-api-key', async_mode=True)

class ChatRequest(BaseModel):
    message: str

@app.post('/chat')
async def chat(request: ChatRequest):
    response = await api.chat.send(
        message=request.message,
        model='taishanglaojun-v1'
    )
    return {'response': response.message}
```

### Django 集成

```python
# views.py
from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
import json
from taishanglaojun import TaiShangLaoJunAPI

api = TaiShangLaoJunAPI(api_key='your-api-key')

@csrf_exempt
@require_http_methods(["POST"])
def chat_view(request):
    data = json.loads(request.body)
    response = api.chat.send(
        message=data['message'],
        model='taishanglaojun-v1'
    )
    return JsonResponse({'response': response.message})
```

## API 参考

### 聊天 API

#### `api.chat.send(**kwargs)`

发送聊天消息。

**参数:**
- `message` (str, 必需) - 用户消息
- `model` (str, 可选) - 使用的模型，默认 'taishanglaojun-v1'
- `conversation_id` (str, 可选) - 对话ID，用于维持上下文
- `user_id` (str, 可选) - 用户ID
- `temperature` (float, 可选) - 温度参数，0-1之间
- `max_tokens` (int, 可选) - 最大token数
- `stream` (bool, 可选) - 是否流式响应

**返回:**
```python
ChatResponse(
    message=str,
    conversation_id=str,
    model=str,
    usage=Usage(
        prompt_tokens=int,
        completion_tokens=int,
        total_tokens=int
    )
)
```

#### `api.chat.stream(**kwargs)`

流式聊天响应。

**参数:** 同 `api.chat.send()`

**返回:** Iterator[ChatChunk]

### 知识库 API

#### `api.knowledge.create(**kwargs)`

创建知识库。

**参数:**
- `name` (str, 必需) - 知识库名称
- `description` (str, 可选) - 描述
- `type` (str, 必需) - 类型: 'document' | 'qa' | 'code'

#### `api.knowledge.search(**kwargs)`

搜索知识库。

**参数:**
- `query` (str, 必需) - 搜索查询
- `knowledge_base_id` (str, 必需) - 知识库ID
- `limit` (int, 可选) - 结果数量限制
- `threshold` (float, 可选) - 相似度阈值

## 错误代码

| 错误代码 | 描述 | 处理建议 |
|---------|------|----------|
| `UNAUTHORIZED` | API密钥无效或过期 | 检查API密钥 |
| `RATE_LIMITED` | 请求频率超限 | 实现重试机制 |
| `VALIDATION_ERROR` | 请求参数错误 | 检查参数格式 |
| `NOT_FOUND` | 资源不存在 | 检查资源ID |
| `INTERNAL_ERROR` | 服务器内部错误 | 联系技术支持 |

## 开发

### 设置开发环境

```bash
git clone https://github.com/taishanglaojun/python-sdk.git
cd python-sdk

# 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 安装开发依赖
pip install -e ".[dev]"

# 安装pre-commit钩子
pre-commit install
```

### 运行测试

```bash
# 运行所有测试
pytest

# 运行测试并生成覆盖率报告
pytest --cov=taishanglaojun --cov-report=html

# 运行特定测试
pytest tests/test_chat.py
```

### 代码格式化

```bash
# 格式化代码
black taishanglaojun tests examples

# 排序导入
isort taishanglaojun tests examples

# 检查代码风格
flake8 taishanglaojun tests examples

# 类型检查
mypy taishanglaojun
```

## 贡献

我们欢迎社区贡献！请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解如何参与开发。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](./LICENSE) 文件。

## 支持

- 📖 [官方文档](https://docs.taishanglaojun.com)
- 💬 [社区论坛](https://community.taishanglaojun.com)
- 📧 [技术支持](mailto:support@taishanglaojun.com)
- 🐛 [问题反馈](https://github.com/taishanglaojun/python-sdk/issues)

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md) 了解版本更新历史。