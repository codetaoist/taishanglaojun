# 太上老君AI平台 - AI开发指南

## 概述

太上老君AI平台的AI服务基于Python和现代AI框架构建，支持多种AI模型和任务类型。本指南涵盖AI服务架构、模型集成、训练部署和性能优化等方面。

## AI架构设计

```mermaid
graph TB
    subgraph "AI服务架构"
        A[API网关] --> B[模型路由器]
        B --> C[文本生成服务]
        B --> D[图像处理服务]
        B --> E[语音处理服务]
        B --> F[多模态服务]
        
        C --> G[OpenAI GPT]
        C --> H[本地LLM]
        D --> I[DALL-E]
        D --> J[Stable Diffusion]
        E --> K[Whisper]
        E --> L[TTS服务]
        F --> M[GPT-4V]
        F --> N[自定义多模态]
    end
    
    subgraph "模型管理"
        O[模型注册中心] --> P[版本控制]
        P --> Q[A/B测试]
        Q --> R[性能监控]
        R --> S[自动扩缩容]
    end
    
    subgraph "数据处理"
        T[数据预处理] --> U[特征提取]
        U --> V[数据增强]
        V --> W[批处理队列]
        W --> X[实时推理]
    end
```

## 技术栈

### 核心框架
- **Python 3.11+**: 主要开发语言
- **FastAPI**: 高性能API框架
- **PyTorch 2.0+**: 深度学习框架
- **Transformers**: Hugging Face模型库
- **LangChain**: LLM应用开发框架
- **Celery**: 异步任务队列

### AI模型库
- **OpenAI API**: GPT系列模型
- **Hugging Face**: 开源模型生态
- **Anthropic Claude**: 对话AI模型
- **Google PaLM**: 大语言模型
- **本地模型**: 自部署开源模型

### 数据存储
- **Vector Database**: Pinecone/Weaviate
- **Model Storage**: MinIO/S3
- **Cache**: Redis
- **Metadata**: PostgreSQL

## 项目结构

```python
# AI服务项目结构
ai_service/
├── app/
│   ├── __init__.py
│   ├── main.py                 # FastAPI应用入口
│   ├── config.py              # 配置管理
│   ├── dependencies.py        # 依赖注入
│   └── middleware/            # 中间件
│       ├── __init__.py
│       ├── auth.py           # 认证中间件
│       ├── cors.py           # CORS中间件
│       └── logging.py        # 日志中间件
├── core/
│   ├── __init__.py
│   ├── models/               # 数据模型
│   │   ├── __init__.py
│   │   ├── base.py          # 基础模型
│   │   ├── chat.py          # 对话模型
│   │   ├── generation.py    # 生成模型
│   │   └── multimodal.py    # 多模态模型
│   ├── services/            # 业务服务
│   │   ├── __init__.py
│   │   ├── chat_service.py  # 对话服务
│   │   ├── generation_service.py # 生成服务
│   │   ├── embedding_service.py  # 嵌入服务
│   │   └── multimodal_service.py # 多模态服务
│   └── providers/           # 模型提供商
│       ├── __init__.py
│       ├── openai_provider.py
│       ├── huggingface_provider.py
│       ├── anthropic_provider.py
│       └── local_provider.py
├── api/
│   ├── __init__.py
│   ├── v1/                  # API版本1
│   │   ├── __init__.py
│   │   ├── chat.py         # 对话API
│   │   ├── generation.py   # 生成API
│   │   ├── embedding.py    # 嵌入API
│   │   └── multimodal.py   # 多模态API
│   └── deps.py             # API依赖
├── utils/
│   ├── __init__.py
│   ├── logger.py           # 日志工具
│   ├── cache.py            # 缓存工具
│   ├── metrics.py          # 指标工具
│   └── validators.py       # 验证工具
├── workers/
│   ├── __init__.py
│   ├── celery_app.py       # Celery应用
│   ├── tasks.py            # 异步任务
│   └── schedulers.py       # 定时任务
├── tests/
│   ├── __init__.py
│   ├── conftest.py         # 测试配置
│   ├── test_api/           # API测试
│   ├── test_services/      # 服务测试
│   └── test_providers/     # 提供商测试
├── scripts/
│   ├── start_server.py     # 启动脚本
│   ├── model_download.py   # 模型下载
│   └── benchmark.py        # 性能测试
├── requirements.txt        # 依赖列表
├── requirements-dev.txt    # 开发依赖
├── Dockerfile             # Docker配置
└── docker-compose.yml     # 容器编排
```

## 核心配置

### 1. 应用配置

```python
# app/config.py
from pydantic import BaseSettings, Field
from typing import List, Optional
import os

class Settings(BaseSettings):
    # 应用配置
    app_name: str = "Taishang Laojun AI Service"
    app_version: str = "1.0.0"
    debug: bool = Field(default=False, env="DEBUG")
    
    # 服务器配置
    host: str = Field(default="0.0.0.0", env="HOST")
    port: int = Field(default=8000, env="PORT")
    workers: int = Field(default=4, env="WORKERS")
    
    # 数据库配置
    database_url: str = Field(env="DATABASE_URL")
    redis_url: str = Field(env="REDIS_URL")
    
    # AI模型配置
    openai_api_key: Optional[str] = Field(env="OPENAI_API_KEY")
    openai_base_url: str = Field(
        default="https://api.openai.com/v1",
        env="OPENAI_BASE_URL"
    )
    anthropic_api_key: Optional[str] = Field(env="ANTHROPIC_API_KEY")
    huggingface_token: Optional[str] = Field(env="HUGGINGFACE_TOKEN")
    
    # 模型存储配置
    model_cache_dir: str = Field(default="./models", env="MODEL_CACHE_DIR")
    model_storage_url: Optional[str] = Field(env="MODEL_STORAGE_URL")
    
    # 向量数据库配置
    pinecone_api_key: Optional[str] = Field(env="PINECONE_API_KEY")
    pinecone_environment: Optional[str] = Field(env="PINECONE_ENVIRONMENT")
    
    # 性能配置
    max_concurrent_requests: int = Field(default=100, env="MAX_CONCURRENT_REQUESTS")
    request_timeout: int = Field(default=300, env="REQUEST_TIMEOUT")
    model_timeout: int = Field(default=120, env="MODEL_TIMEOUT")
    
    # 缓存配置
    cache_ttl: int = Field(default=3600, env="CACHE_TTL")
    enable_cache: bool = Field(default=True, env="ENABLE_CACHE")
    
    # 日志配置
    log_level: str = Field(default="INFO", env="LOG_LEVEL")
    log_format: str = Field(default="json", env="LOG_FORMAT")
    
    # 安全配置
    allowed_hosts: List[str] = Field(default=["*"], env="ALLOWED_HOSTS")
    cors_origins: List[str] = Field(default=["*"], env="CORS_ORIGINS")
    
    # 监控配置
    enable_metrics: bool = Field(default=True, env="ENABLE_METRICS")
    metrics_port: int = Field(default=9090, env="METRICS_PORT")
    
    class Config:
        env_file = ".env"
        case_sensitive = False

# 全局设置实例
settings = Settings()
```

### 2. 模型提供商配置

```python
# core/providers/base.py
from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional, AsyncGenerator
from pydantic import BaseModel
import asyncio
import logging

logger = logging.getLogger(__name__)

class ModelConfig(BaseModel):
    """模型配置基类"""
    name: str
    provider: str
    model_type: str
    max_tokens: int = 4096
    temperature: float = 0.7
    top_p: float = 1.0
    frequency_penalty: float = 0.0
    presence_penalty: float = 0.0
    timeout: int = 120
    retry_attempts: int = 3
    retry_delay: float = 1.0

class ChatMessage(BaseModel):
    """聊天消息"""
    role: str  # system, user, assistant
    content: str
    metadata: Optional[Dict[str, Any]] = None

class ChatRequest(BaseModel):
    """聊天请求"""
    messages: List[ChatMessage]
    model: str
    max_tokens: Optional[int] = None
    temperature: Optional[float] = None
    stream: bool = False
    user_id: Optional[str] = None

class ChatResponse(BaseModel):
    """聊天响应"""
    id: str
    model: str
    choices: List[Dict[str, Any]]
    usage: Dict[str, int]
    created: int

class BaseProvider(ABC):
    """AI模型提供商基类"""
    
    def __init__(self, config: ModelConfig):
        self.config = config
        self.logger = logging.getLogger(f"{__name__}.{self.__class__.__name__}")
    
    @abstractmethod
    async def chat_completion(
        self, 
        request: ChatRequest
    ) -> ChatResponse:
        """聊天补全"""
        pass
    
    @abstractmethod
    async def chat_completion_stream(
        self, 
        request: ChatRequest
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """流式聊天补全"""
        pass
    
    @abstractmethod
    async def text_embedding(
        self, 
        texts: List[str]
    ) -> List[List[float]]:
        """文本嵌入"""
        pass
    
    @abstractmethod
    async def health_check(self) -> bool:
        """健康检查"""
        pass
    
    async def _retry_request(self, func, *args, **kwargs):
        """重试机制"""
        last_exception = None
        
        for attempt in range(self.config.retry_attempts):
            try:
                return await func(*args, **kwargs)
            except Exception as e:
                last_exception = e
                if attempt < self.config.retry_attempts - 1:
                    delay = self.config.retry_delay * (2 ** attempt)
                    self.logger.warning(
                        f"Request failed (attempt {attempt + 1}), retrying in {delay}s: {e}"
                    )
                    await asyncio.sleep(delay)
                else:
                    self.logger.error(f"Request failed after {self.config.retry_attempts} attempts: {e}")
        
        raise last_exception
```

### 3. OpenAI提供商实现

```python
# core/providers/openai_provider.py
import openai
import time
import uuid
from typing import List, Dict, Any, AsyncGenerator
from core.providers.base import BaseProvider, ChatRequest, ChatResponse, ChatMessage

class OpenAIProvider(BaseProvider):
    """OpenAI模型提供商"""
    
    def __init__(self, config: ModelConfig, api_key: str, base_url: str = None):
        super().__init__(config)
        self.client = openai.AsyncOpenAI(
            api_key=api_key,
            base_url=base_url
        )
    
    async def chat_completion(self, request: ChatRequest) -> ChatResponse:
        """聊天补全"""
        try:
            # 转换消息格式
            messages = [
                {"role": msg.role, "content": msg.content}
                for msg in request.messages
            ]
            
            # 调用OpenAI API
            response = await self._retry_request(
                self.client.chat.completions.create,
                model=request.model,
                messages=messages,
                max_tokens=request.max_tokens or self.config.max_tokens,
                temperature=request.temperature or self.config.temperature,
                top_p=self.config.top_p,
                frequency_penalty=self.config.frequency_penalty,
                presence_penalty=self.config.presence_penalty,
                user=request.user_id
            )
            
            # 转换响应格式
            return ChatResponse(
                id=response.id,
                model=response.model,
                choices=[
                    {
                        "index": choice.index,
                        "message": {
                            "role": choice.message.role,
                            "content": choice.message.content
                        },
                        "finish_reason": choice.finish_reason
                    }
                    for choice in response.choices
                ],
                usage={
                    "prompt_tokens": response.usage.prompt_tokens,
                    "completion_tokens": response.usage.completion_tokens,
                    "total_tokens": response.usage.total_tokens
                },
                created=response.created
            )
            
        except Exception as e:
            self.logger.error(f"OpenAI chat completion failed: {e}")
            raise
    
    async def chat_completion_stream(
        self, 
        request: ChatRequest
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """流式聊天补全"""
        try:
            messages = [
                {"role": msg.role, "content": msg.content}
                for msg in request.messages
            ]
            
            stream = await self.client.chat.completions.create(
                model=request.model,
                messages=messages,
                max_tokens=request.max_tokens or self.config.max_tokens,
                temperature=request.temperature or self.config.temperature,
                stream=True,
                user=request.user_id
            )
            
            async for chunk in stream:
                if chunk.choices:
                    choice = chunk.choices[0]
                    if choice.delta.content:
                        yield {
                            "id": chunk.id,
                            "model": chunk.model,
                            "choices": [{
                                "index": choice.index,
                                "delta": {
                                    "content": choice.delta.content
                                },
                                "finish_reason": choice.finish_reason
                            }],
                            "created": chunk.created
                        }
                        
        except Exception as e:
            self.logger.error(f"OpenAI stream completion failed: {e}")
            raise
    
    async def text_embedding(self, texts: List[str]) -> List[List[float]]:
        """文本嵌入"""
        try:
            response = await self._retry_request(
                self.client.embeddings.create,
                model="text-embedding-ada-002",
                input=texts
            )
            
            return [data.embedding for data in response.data]
            
        except Exception as e:
            self.logger.error(f"OpenAI embedding failed: {e}")
            raise
    
    async def health_check(self) -> bool:
        """健康检查"""
        try:
            await self.client.models.list()
            return True
        except Exception as e:
            self.logger.error(f"OpenAI health check failed: {e}")
            return False
```

## 服务层实现

### 1. 聊天服务

```python
# core/services/chat_service.py
from typing import List, Dict, Any, Optional, AsyncGenerator
import uuid
import time
from datetime import datetime

from core.providers.base import BaseProvider, ChatRequest, ChatResponse, ChatMessage
from core.models.chat import ChatSession, ChatHistory
from utils.cache import CacheManager
from utils.metrics import MetricsCollector

class ChatService:
    """聊天服务"""
    
    def __init__(
        self,
        providers: Dict[str, BaseProvider],
        cache_manager: CacheManager,
        metrics_collector: MetricsCollector
    ):
        self.providers = providers
        self.cache = cache_manager
        self.metrics = metrics_collector
        self.logger = logging.getLogger(__name__)
    
    async def create_session(
        self,
        user_id: str,
        model: str,
        system_prompt: Optional[str] = None
    ) -> ChatSession:
        """创建聊天会话"""
        session = ChatSession(
            id=str(uuid.uuid4()),
            user_id=user_id,
            model=model,
            system_prompt=system_prompt,
            created_at=datetime.utcnow(),
            updated_at=datetime.utcnow()
        )
        
        # 缓存会话信息
        await self.cache.set(
            f"chat_session:{session.id}",
            session.dict(),
            ttl=3600
        )
        
        self.logger.info(f"Created chat session {session.id} for user {user_id}")
        return session
    
    async def get_session(self, session_id: str) -> Optional[ChatSession]:
        """获取聊天会话"""
        session_data = await self.cache.get(f"chat_session:{session_id}")
        if session_data:
            return ChatSession(**session_data)
        return None
    
    async def chat(
        self,
        session_id: str,
        message: str,
        stream: bool = False
    ) -> ChatResponse:
        """发送聊天消息"""
        start_time = time.time()
        
        try:
            # 获取会话
            session = await self.get_session(session_id)
            if not session:
                raise ValueError(f"Session {session_id} not found")
            
            # 获取提供商
            provider = self.providers.get(session.model)
            if not provider:
                raise ValueError(f"Provider for model {session.model} not found")
            
            # 构建消息历史
            messages = []
            if session.system_prompt:
                messages.append(ChatMessage(
                    role="system",
                    content=session.system_prompt
                ))
            
            # 添加历史消息
            for history in session.history[-10:]:  # 只保留最近10条
                messages.append(ChatMessage(
                    role=history.role,
                    content=history.content
                ))
            
            # 添加当前消息
            messages.append(ChatMessage(
                role="user",
                content=message
            ))
            
            # 创建请求
            request = ChatRequest(
                messages=messages,
                model=session.model,
                stream=stream,
                user_id=session.user_id
            )
            
            # 调用模型
            if stream:
                return self._handle_stream_response(session, request, message)
            else:
                response = await provider.chat_completion(request)
                
                # 保存消息历史
                await self._save_message_history(session, message, response)
                
                # 记录指标
                duration = time.time() - start_time
                await self.metrics.record_chat_request(
                    model=session.model,
                    duration=duration,
                    tokens=response.usage["total_tokens"],
                    success=True
                )
                
                return response
                
        except Exception as e:
            # 记录错误指标
            duration = time.time() - start_time
            await self.metrics.record_chat_request(
                model=session.model if 'session' in locals() else "unknown",
                duration=duration,
                tokens=0,
                success=False,
                error=str(e)
            )
            raise
    
    async def _handle_stream_response(
        self,
        session: ChatSession,
        request: ChatRequest,
        user_message: str
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """处理流式响应"""
        provider = self.providers[session.model]
        assistant_message = ""
        
        async for chunk in provider.chat_completion_stream(request):
            if chunk.get("choices") and chunk["choices"][0].get("delta", {}).get("content"):
                content = chunk["choices"][0]["delta"]["content"]
                assistant_message += content
                yield chunk
        
        # 保存完整的消息历史
        if assistant_message:
            await self._save_message_history_manual(
                session, user_message, assistant_message
            )
    
    async def _save_message_history(
        self,
        session: ChatSession,
        user_message: str,
        response: ChatResponse
    ):
        """保存消息历史"""
        # 添加用户消息
        session.history.append(ChatHistory(
            role="user",
            content=user_message,
            timestamp=datetime.utcnow()
        ))
        
        # 添加助手回复
        if response.choices:
            assistant_message = response.choices[0]["message"]["content"]
            session.history.append(ChatHistory(
                role="assistant",
                content=assistant_message,
                timestamp=datetime.utcnow()
            ))
        
        # 更新会话
        session.updated_at = datetime.utcnow()
        await self.cache.set(
            f"chat_session:{session.id}",
            session.dict(),
            ttl=3600
        )
    
    async def _save_message_history_manual(
        self,
        session: ChatSession,
        user_message: str,
        assistant_message: str
    ):
        """手动保存消息历史（用于流式响应）"""
        session.history.append(ChatHistory(
            role="user",
            content=user_message,
            timestamp=datetime.utcnow()
        ))
        
        session.history.append(ChatHistory(
            role="assistant",
            content=assistant_message,
            timestamp=datetime.utcnow()
        ))
        
        session.updated_at = datetime.utcnow()
        await self.cache.set(
            f"chat_session:{session.id}",
            session.dict(),
            ttl=3600
        )
    
    async def get_chat_history(
        self,
        session_id: str,
        limit: int = 50
    ) -> List[ChatHistory]:
        """获取聊天历史"""
        session = await self.get_session(session_id)
        if not session:
            return []
        
        return session.history[-limit:]
    
    async def delete_session(self, session_id: str):
        """删除聊天会话"""
        await self.cache.delete(f"chat_session:{session_id}")
        self.logger.info(f"Deleted chat session {session_id}")
```

### 2. 嵌入服务

```python
# core/services/embedding_service.py
import hashlib
import numpy as np
from typing import List, Dict, Any, Optional
import time

from core.providers.base import BaseProvider
from utils.cache import CacheManager
from utils.metrics import MetricsCollector

class EmbeddingService:
    """嵌入服务"""
    
    def __init__(
        self,
        providers: Dict[str, BaseProvider],
        cache_manager: CacheManager,
        metrics_collector: MetricsCollector
    ):
        self.providers = providers
        self.cache = cache_manager
        self.metrics = metrics_collector
        self.logger = logging.getLogger(__name__)
    
    async def create_embeddings(
        self,
        texts: List[str],
        model: str = "text-embedding-ada-002"
    ) -> List[List[float]]:
        """创建文本嵌入"""
        start_time = time.time()
        
        try:
            # 检查缓存
            cached_embeddings = await self._get_cached_embeddings(texts, model)
            if cached_embeddings:
                self.logger.info(f"Retrieved {len(cached_embeddings)} embeddings from cache")
                return cached_embeddings
            
            # 获取提供商
            provider = self.providers.get(model)
            if not provider:
                raise ValueError(f"Provider for model {model} not found")
            
            # 生成嵌入
            embeddings = await provider.text_embedding(texts)
            
            # 缓存结果
            await self._cache_embeddings(texts, embeddings, model)
            
            # 记录指标
            duration = time.time() - start_time
            await self.metrics.record_embedding_request(
                model=model,
                duration=duration,
                text_count=len(texts),
                success=True
            )
            
            self.logger.info(f"Generated {len(embeddings)} embeddings for model {model}")
            return embeddings
            
        except Exception as e:
            duration = time.time() - start_time
            await self.metrics.record_embedding_request(
                model=model,
                duration=duration,
                text_count=len(texts),
                success=False,
                error=str(e)
            )
            raise
    
    async def similarity_search(
        self,
        query_embedding: List[float],
        candidate_embeddings: List[List[float]],
        top_k: int = 10
    ) -> List[Dict[str, Any]]:
        """相似度搜索"""
        try:
            # 计算余弦相似度
            query_norm = np.linalg.norm(query_embedding)
            similarities = []
            
            for i, candidate in enumerate(candidate_embeddings):
                candidate_norm = np.linalg.norm(candidate)
                if query_norm == 0 or candidate_norm == 0:
                    similarity = 0.0
                else:
                    similarity = np.dot(query_embedding, candidate) / (query_norm * candidate_norm)
                
                similarities.append({
                    "index": i,
                    "similarity": float(similarity)
                })
            
            # 排序并返回top_k
            similarities.sort(key=lambda x: x["similarity"], reverse=True)
            return similarities[:top_k]
            
        except Exception as e:
            self.logger.error(f"Similarity search failed: {e}")
            raise
    
    async def _get_cached_embeddings(
        self,
        texts: List[str],
        model: str
    ) -> Optional[List[List[float]]]:
        """从缓存获取嵌入"""
        cache_keys = [self._get_cache_key(text, model) for text in texts]
        cached_results = await self.cache.get_many(cache_keys)
        
        # 检查是否所有文本都有缓存
        if all(result is not None for result in cached_results):
            return cached_results
        
        return None
    
    async def _cache_embeddings(
        self,
        texts: List[str],
        embeddings: List[List[float]],
        model: str
    ):
        """缓存嵌入结果"""
        cache_data = {}
        for text, embedding in zip(texts, embeddings):
            cache_key = self._get_cache_key(text, model)
            cache_data[cache_key] = embedding
        
        await self.cache.set_many(cache_data, ttl=86400)  # 缓存24小时
    
    def _get_cache_key(self, text: str, model: str) -> str:
        """生成缓存键"""
        text_hash = hashlib.md5(text.encode()).hexdigest()
        return f"embedding:{model}:{text_hash}"
```

## API接口实现

### 1. 聊天API

```python
# api/v1/chat.py
from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
from fastapi.responses import StreamingResponse
from typing import List, Optional
import json

from core.services.chat_service import ChatService
from core.models.chat import ChatRequest, ChatResponse
from api.deps import get_chat_service, get_current_user
from utils.auth import User

router = APIRouter(prefix="/chat", tags=["chat"])

@router.post("/sessions", response_model=dict)
async def create_chat_session(
    model: str,
    system_prompt: Optional[str] = None,
    current_user: User = Depends(get_current_user),
    chat_service: ChatService = Depends(get_chat_service)
):
    """创建聊天会话"""
    try:
        session = await chat_service.create_session(
            user_id=current_user.id,
            model=model,
            system_prompt=system_prompt
        )
        return {"session_id": session.id}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@router.post("/sessions/{session_id}/messages")
async def send_message(
    session_id: str,
    message: str,
    stream: bool = False,
    current_user: User = Depends(get_current_user),
    chat_service: ChatService = Depends(get_chat_service)
):
    """发送聊天消息"""
    try:
        if stream:
            async def generate():
                async for chunk in chat_service.chat(session_id, message, stream=True):
                    yield f"data: {json.dumps(chunk)}\n\n"
                yield "data: [DONE]\n\n"
            
            return StreamingResponse(
                generate(),
                media_type="text/plain",
                headers={"Cache-Control": "no-cache"}
            )
        else:
            response = await chat_service.chat(session_id, message)
            return response.dict()
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/sessions/{session_id}/history")
async def get_chat_history(
    session_id: str,
    limit: int = 50,
    current_user: User = Depends(get_current_user),
    chat_service: ChatService = Depends(get_chat_service)
):
    """获取聊天历史"""
    try:
        history = await chat_service.get_chat_history(session_id, limit)
        return {"history": [h.dict() for h in history]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.delete("/sessions/{session_id}")
async def delete_chat_session(
    session_id: str,
    current_user: User = Depends(get_current_user),
    chat_service: ChatService = Depends(get_chat_service)
):
    """删除聊天会话"""
    try:
        await chat_service.delete_session(session_id)
        return {"message": "Session deleted successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
```

## 性能优化

### 1. 模型缓存策略

```python
# utils/model_cache.py
import asyncio
import time
from typing import Dict, Any, Optional
from dataclasses import dataclass
import logging

@dataclass
class CacheEntry:
    data: Any
    timestamp: float
    access_count: int
    ttl: float

class ModelCache:
    """模型缓存管理器"""
    
    def __init__(self, max_size: int = 1000, default_ttl: float = 3600):
        self.max_size = max_size
        self.default_ttl = default_ttl
        self.cache: Dict[str, CacheEntry] = {}
        self.access_times: Dict[str, float] = {}
        self.logger = logging.getLogger(__name__)
        
        # 启动清理任务
        asyncio.create_task(self._cleanup_task())
    
    async def get(self, key: str) -> Optional[Any]:
        """获取缓存项"""
        if key not in self.cache:
            return None
        
        entry = self.cache[key]
        current_time = time.time()
        
        # 检查是否过期
        if current_time - entry.timestamp > entry.ttl:
            del self.cache[key]
            if key in self.access_times:
                del self.access_times[key]
            return None
        
        # 更新访问信息
        entry.access_count += 1
        self.access_times[key] = current_time
        
        return entry.data
    
    async def set(self, key: str, value: Any, ttl: Optional[float] = None):
        """设置缓存项"""
        if ttl is None:
            ttl = self.default_ttl
        
        current_time = time.time()
        
        # 如果缓存已满，清理最少使用的项
        if len(self.cache) >= self.max_size and key not in self.cache:
            await self._evict_lru()
        
        self.cache[key] = CacheEntry(
            data=value,
            timestamp=current_time,
            access_count=1,
            ttl=ttl
        )
        self.access_times[key] = current_time
    
    async def delete(self, key: str):
        """删除缓存项"""
        if key in self.cache:
            del self.cache[key]
        if key in self.access_times:
            del self.access_times[key]
    
    async def clear(self):
        """清空缓存"""
        self.cache.clear()
        self.access_times.clear()
    
    async def _evict_lru(self):
        """清理最少使用的缓存项"""
        if not self.access_times:
            return
        
        # 找到最少使用的键
        lru_key = min(self.access_times.items(), key=lambda x: x[1])[0]
        await self.delete(lru_key)
        
        self.logger.debug(f"Evicted LRU cache entry: {lru_key}")
    
    async def _cleanup_task(self):
        """定期清理过期缓存"""
        while True:
            try:
                await asyncio.sleep(300)  # 每5分钟清理一次
                current_time = time.time()
                expired_keys = []
                
                for key, entry in self.cache.items():
                    if current_time - entry.timestamp > entry.ttl:
                        expired_keys.append(key)
                
                for key in expired_keys:
                    await self.delete(key)
                
                if expired_keys:
                    self.logger.info(f"Cleaned up {len(expired_keys)} expired cache entries")
                    
            except Exception as e:
                self.logger.error(f"Cache cleanup failed: {e}")
```

### 2. 批处理优化

```python
# utils/batch_processor.py
import asyncio
from typing import List, Callable, Any, Dict
from dataclasses import dataclass
import time
import logging

@dataclass
class BatchRequest:
    id: str
    data: Any
    future: asyncio.Future
    timestamp: float

class BatchProcessor:
    """批处理器"""
    
    def __init__(
        self,
        batch_size: int = 10,
        max_wait_time: float = 1.0,
        processor_func: Callable = None
    ):
        self.batch_size = batch_size
        self.max_wait_time = max_wait_time
        self.processor_func = processor_func
        self.pending_requests: List[BatchRequest] = []
        self.processing = False
        self.logger = logging.getLogger(__name__)
    
    async def add_request(self, request_id: str, data: Any) -> Any:
        """添加请求到批处理队列"""
        future = asyncio.Future()
        request = BatchRequest(
            id=request_id,
            data=data,
            future=future,
            timestamp=time.time()
        )
        
        self.pending_requests.append(request)
        
        # 如果达到批处理大小或者是第一个请求，立即处理
        if len(self.pending_requests) >= self.batch_size or len(self.pending_requests) == 1:
            asyncio.create_task(self._process_batch())
        
        return await future
    
    async def _process_batch(self):
        """处理批次"""
        if self.processing or not self.pending_requests:
            return
        
        self.processing = True
        
        try:
            # 等待更多请求或超时
            if len(self.pending_requests) < self.batch_size:
                await asyncio.sleep(self.max_wait_time)
            
            # 获取当前批次
            current_batch = self.pending_requests[:self.batch_size]
            self.pending_requests = self.pending_requests[self.batch_size:]
            
            if not current_batch:
                return
            
            # 处理批次
            batch_data = [req.data for req in current_batch]
            
            try:
                results = await self.processor_func(batch_data)
                
                # 返回结果
                for request, result in zip(current_batch, results):
                    if not request.future.done():
                        request.future.set_result(result)
                        
            except Exception as e:
                # 设置错误
                for request in current_batch:
                    if not request.future.done():
                        request.future.set_exception(e)
            
            self.logger.info(f"Processed batch of {len(current_batch)} requests")
            
        finally:
            self.processing = False
            
            # 如果还有待处理的请求，继续处理
            if self.pending_requests:
                asyncio.create_task(self._process_batch())
```

## 监控和指标

### 1. 指标收集

```python
# utils/metrics.py
from prometheus_client import Counter, Histogram, Gauge, start_http_server
import time
import logging

class MetricsCollector:
    """指标收集器"""
    
    def __init__(self):
        # 请求指标
        self.request_count = Counter(
            'ai_requests_total',
            'Total AI requests',
            ['model', 'endpoint', 'status']
        )
        
        self.request_duration = Histogram(
            'ai_request_duration_seconds',
            'AI request duration',
            ['model', 'endpoint']
        )
        
        self.token_usage = Counter(
            'ai_tokens_total',
            'Total tokens used',
            ['model', 'type']  # type: prompt, completion
        )
        
        # 模型指标
        self.model_health = Gauge(
            'ai_model_health',
            'Model health status',
            ['model', 'provider']
        )
        
        self.active_sessions = Gauge(
            'ai_active_sessions',
            'Number of active chat sessions'
        )
        
        # 缓存指标
        self.cache_hits = Counter(
            'ai_cache_hits_total',
            'Cache hits',
            ['cache_type']
        )
        
        self.cache_misses = Counter(
            'ai_cache_misses_total',
            'Cache misses',
            ['cache_type']
        )
        
        self.logger = logging.getLogger(__name__)
    
    async def record_chat_request(
        self,
        model: str,
        duration: float,
        tokens: int,
        success: bool,
        error: str = None
    ):
        """记录聊天请求指标"""
        status = "success" if success else "error"
        
        self.request_count.labels(
            model=model,
            endpoint="chat",
            status=status
        ).inc()
        
        if success:
            self.request_duration.labels(
                model=model,
                endpoint="chat"
            ).observe(duration)
            
            self.token_usage.labels(
                model=model,
                type="total"
            ).inc(tokens)
        
        if error:
            self.logger.error(f"Chat request failed: {error}")
    
    async def record_embedding_request(
        self,
        model: str,
        duration: float,
        text_count: int,
        success: bool,
        error: str = None
    ):
        """记录嵌入请求指标"""
        status = "success" if success else "error"
        
        self.request_count.labels(
            model=model,
            endpoint="embedding",
            status=status
        ).inc()
        
        if success:
            self.request_duration.labels(
                model=model,
                endpoint="embedding"
            ).observe(duration)
    
    async def update_model_health(self, model: str, provider: str, healthy: bool):
        """更新模型健康状态"""
        self.model_health.labels(
            model=model,
            provider=provider
        ).set(1 if healthy else 0)
    
    async def update_active_sessions(self, count: int):
        """更新活跃会话数"""
        self.active_sessions.set(count)
    
    async def record_cache_hit(self, cache_type: str):
        """记录缓存命中"""
        self.cache_hits.labels(cache_type=cache_type).inc()
    
    async def record_cache_miss(self, cache_type: str):
        """记录缓存未命中"""
        self.cache_misses.labels(cache_type=cache_type).inc()
    
    def start_metrics_server(self, port: int = 9090):
        """启动指标服务器"""
        start_http_server(port)
        self.logger.info(f"Metrics server started on port {port}")
```

## 部署配置

### 1. Docker配置

```dockerfile
# Dockerfile
FROM python:3.11-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 复制应用代码
COPY . .

# 创建非root用户
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# 暴露端口
EXPOSE 8000 9090

# 启动命令
CMD ["python", "scripts/start_server.py"]
```

### 2. Docker Compose配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  ai-service:
    build: .
    ports:
      - "8000:8000"
      - "9090:9090"
    environment:
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/taishanglaojun
      - REDIS_URL=redis://redis:6379/0
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - LOG_LEVEL=INFO
    depends_on:
      - postgres
      - redis
    volumes:
      - ./models:/app/models
    restart: unless-stopped
    
  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=taishanglaojun
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    
  celery-worker:
    build: .
    command: celery -A workers.celery_app worker --loglevel=info
    environment:
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/taishanglaojun
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - postgres
      - redis
    volumes:
      - ./models:/app/models
    restart: unless-stopped
    
  celery-beat:
    build: .
    command: celery -A workers.celery_app beat --loglevel=info
    environment:
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/taishanglaojun
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

## 相关文档链接

- [开发指南概览](./development-overview.md)
- [环境搭建指南](./environment-setup.md)
- [前端开发指南](./frontend-development.md)
- [后端开发指南](./backend-development.md)
- [测试指南](./testing-guide.md)
- [API文档](../06-API文档/ai-service-api.md)
- [部署运维指南](../08-部署运维/deployment-overview.md)