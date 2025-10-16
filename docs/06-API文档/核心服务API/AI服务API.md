# AI服务API文档

## 1. 服务概述

AI服务是太上老君AI平台的核心智能引擎，基于S×C×T三轴理论设计，集成多种AI模型和算法，提供文本生成、图像处理、语音识别、知识问答、个性化推荐等智能服务。

### 1.1 服务架构

```mermaid
graph TB
    subgraph "AI服务架构"
        subgraph "API网关层"
            REST[REST API]
            GraphQL[GraphQL API]
            WS[WebSocket API]
            STREAM[流式API]
        end
        
        subgraph "AI引擎层"
            LLM[大语言模型]
            CV[计算机视觉]
            ASR[语音识别]
            TTS[语音合成]
            NLP[自然语言处理]
            ML[机器学习]
        end
        
        subgraph "模型管理层"
            MODEL_STORE[模型仓库]
            MODEL_LOADER[模型加载器]
            MODEL_CACHE[模型缓存]
            VERSION_CTRL[版本控制]
        end
        
        subgraph "推理服务层"
            INFERENCE[推理引擎]
            BATCH[批处理]
            QUEUE[任务队列]
            SCHEDULER[调度器]
        end
        
        subgraph "数据层"
            VECTOR_DB[(向量数据库)]
            KNOWLEDGE_BASE[(知识库)]
            TRAINING_DATA[(训练数据)]
            CACHE[(缓存)]
        end
    end
    
    REST --> LLM
    GraphQL --> CV
    WS --> ASR
    STREAM --> TTS
    
    LLM --> MODEL_STORE
    CV --> MODEL_LOADER
    ASR --> MODEL_CACHE
    TTS --> VERSION_CTRL
    
    MODEL_STORE --> INFERENCE
    MODEL_LOADER --> BATCH
    MODEL_CACHE --> QUEUE
    VERSION_CTRL --> SCHEDULER
    
    INFERENCE --> VECTOR_DB
    BATCH --> KNOWLEDGE_BASE
    QUEUE --> TRAINING_DATA
    SCHEDULER --> CACHE
```

### 1.2 核心功能

- **文本生成**：基于大语言模型的智能文本生成
- **对话系统**：多轮对话和上下文理解
- **知识问答**：基于知识库的智能问答
- **图像处理**：图像生成、识别、分析和编辑
- **语音服务**：语音识别、合成和处理
- **个性化推荐**：基于用户行为的智能推荐
- **内容审核**：智能内容安全检测
- **数据分析**：智能数据洞察和预测

## 2. 文本生成服务

### 2.1 通用文本生成

#### 2.1.1 单次文本生成

```http
POST /api/v1/ai/text/generate
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "prompt": "请写一篇关于人工智能发展趋势的文章",
  "model": "gpt-4-turbo",
  "parameters": {
    "max_tokens": 2000,
    "temperature": 0.7,
    "top_p": 0.9,
    "frequency_penalty": 0.1,
    "presence_penalty": 0.1,
    "stop_sequences": ["[END]"]
  },
  "context": {
    "user_id": "usr_1234567890abcdef",
    "session_id": "sess_abcdef1234567890",
    "language": "zh-CN",
    "domain": "technology"
  },
  "options": {
    "stream": false,
    "include_usage": true,
    "safety_check": true,
    "format": "markdown"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "generation_id": "gen_1234567890abcdef",
    "text": "# 人工智能发展趋势\n\n人工智能（AI）作为21世纪最具革命性的技术之一...",
    "model": "gpt-4-turbo",
    "usage": {
      "prompt_tokens": 25,
      "completion_tokens": 1500,
      "total_tokens": 1525
    },
    "safety": {
      "flagged": false,
      "categories": {
        "hate": false,
        "harassment": false,
        "self_harm": false,
        "sexual": false,
        "violence": false
      }
    },
    "metadata": {
      "generation_time": 3.2,
      "model_version": "gpt-4-turbo-2024-01-01",
      "finish_reason": "stop"
    }
  },
  "meta": {
    "api_version": "v1",
    "request_id": "req_1234567890abcdef",
    "execution_time": 3200
  }
}
```

#### 2.1.2 流式文本生成

```http
POST /api/v1/ai/text/generate/stream
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "prompt": "请解释量子计算的基本原理",
  "model": "gpt-4-turbo",
  "parameters": {
    "max_tokens": 1000,
    "temperature": 0.5,
    "stream": true
  }
}
```

**流式响应示例：**

```
data: {"type": "start", "generation_id": "gen_1234567890abcdef"}

data: {"type": "token", "text": "量子", "index": 0}

data: {"type": "token", "text": "计算", "index": 1}

data: {"type": "token", "text": "是", "index": 2}

data: {"type": "token", "text": "一种", "index": 3}

...

data: {"type": "finish", "finish_reason": "stop", "usage": {"total_tokens": 856}}

data: [DONE]
```

### 2.2 专业文本生成

#### 2.2.1 学术写作

```http
POST /api/v1/ai/text/academic
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "type": "research_paper",
  "topic": "深度学习在医疗诊断中的应用",
  "requirements": {
    "sections": ["abstract", "introduction", "methodology", "results", "conclusion"],
    "citation_style": "APA",
    "word_count": 5000,
    "academic_level": "graduate"
  },
  "references": [
    {
      "title": "Deep Learning for Medical Image Analysis",
      "authors": ["Smith, J.", "Doe, A."],
      "year": 2023,
      "journal": "Nature Medicine"
    }
  ]
}
```

#### 2.2.2 创意写作

```http
POST /api/v1/ai/text/creative
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "type": "story",
  "genre": "science_fiction",
  "theme": "人工智能觉醒",
  "style": "现代小说",
  "length": "short_story",
  "parameters": {
    "character_count": 3,
    "setting": "2050年的上海",
    "tone": "suspenseful",
    "perspective": "third_person"
  }
}
```

#### 2.2.3 商业文档

```http
POST /api/v1/ai/text/business
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "type": "business_plan",
  "industry": "artificial_intelligence",
  "company_info": {
    "name": "智慧科技有限公司",
    "stage": "startup",
    "target_market": "企业级AI解决方案"
  },
  "sections": [
    "executive_summary",
    "market_analysis",
    "product_description",
    "marketing_strategy",
    "financial_projections"
  ]
}
```

## 3. 对话系统

### 3.1 创建对话会话

```http
POST /api/v1/ai/chat/sessions
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "AI助手对话",
  "model": "gpt-4-turbo",
  "system_prompt": "你是一个专业的AI助手，擅长回答各种问题并提供有用的建议。",
  "parameters": {
    "temperature": 0.7,
    "max_tokens": 2000,
    "context_window": 8000
  },
  "persona": {
    "name": "小智",
    "personality": "友善、专业、耐心",
    "expertise": ["人工智能", "技术咨询", "学习指导"]
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "session_id": "chat_1234567890abcdef",
    "name": "AI助手对话",
    "model": "gpt-4-turbo",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "message_count": 0,
    "token_usage": 0
  }
}
```

### 3.2 发送消息

```http
POST /api/v1/ai/chat/sessions/{session_id}/messages
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "message": "你好，我想了解一下机器学习的基本概念",
  "type": "text",
  "attachments": [
    {
      "type": "image",
      "url": "https://example.com/image.jpg",
      "description": "相关图表"
    }
  ],
  "options": {
    "stream": true,
    "include_sources": true,
    "temperature": 0.7
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "message_id": "msg_1234567890abcdef",
    "session_id": "chat_1234567890abcdef",
    "role": "assistant",
    "content": "你好！我很高兴为你介绍机器学习的基本概念。\n\n机器学习是人工智能的一个重要分支...",
    "type": "text",
    "sources": [
      {
        "title": "机器学习基础",
        "url": "https://knowledge.taishanglaojun.com/ml-basics",
        "relevance": 0.95
      }
    ],
    "usage": {
      "prompt_tokens": 150,
      "completion_tokens": 800,
      "total_tokens": 950
    },
    "created_at": "2024-01-15T10:31:00Z"
  }
}
```

### 3.3 获取对话历史

```http
GET /api/v1/ai/chat/sessions/{session_id}/messages?limit=50&offset=0
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 3.4 多模态对话

```http
POST /api/v1/ai/chat/sessions/{session_id}/messages/multimodal
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

message: "请分析这张图片中的内容"
image: [binary file data]
audio: [binary file data]
options: {"include_ocr": true, "include_objects": true}
```

## 4. 知识问答服务

### 4.1 知识库问答

```http
POST /api/v1/ai/knowledge/query
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "question": "太上老君在道教中的地位是什么？",
  "knowledge_bases": ["chinese_culture", "taoism", "philosophy"],
  "parameters": {
    "max_results": 5,
    "min_relevance": 0.7,
    "include_sources": true,
    "language": "zh-CN"
  },
  "context": {
    "user_background": "对中国传统文化感兴趣的学习者",
    "detail_level": "intermediate"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "answer": "太上老君在道教中具有至高无上的地位，被尊为道教三清之一...",
    "confidence": 0.92,
    "sources": [
      {
        "id": "kb_taoism_001",
        "title": "道教三清尊神",
        "content": "太上老君，又称道德天尊...",
        "relevance": 0.95,
        "source_type": "knowledge_base",
        "url": "https://knowledge.taishanglaojun.com/taoism/sanqing"
      }
    ],
    "related_questions": [
      "道教三清分别是哪三位？",
      "太上老君与老子的关系是什么？",
      "道教的基本教义有哪些？"
    ],
    "metadata": {
      "query_time": 0.8,
      "knowledge_bases_searched": 3,
      "total_documents": 1250
    }
  }
}
```

### 4.2 文档问答

```http
POST /api/v1/ai/knowledge/document-qa
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "question": "这份报告的主要结论是什么？",
  "document": {
    "type": "pdf",
    "url": "https://example.com/report.pdf",
    "content": "文档文本内容...",
    "metadata": {
      "title": "2024年AI发展报告",
      "author": "研究院",
      "pages": 50
    }
  },
  "options": {
    "extract_key_points": true,
    "summarize": true,
    "language": "zh-CN"
  }
}
```

### 4.3 实时知识检索

```http
POST /api/v1/ai/knowledge/search
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "query": "人工智能在教育领域的应用",
  "filters": {
    "date_range": {
      "start": "2023-01-01",
      "end": "2024-01-01"
    },
    "content_type": ["article", "research_paper", "case_study"],
    "language": "zh-CN",
    "domain": "education"
  },
  "search_options": {
    "semantic_search": true,
    "hybrid_search": true,
    "rerank": true,
    "max_results": 20
  }
}
```

## 5. 图像处理服务

### 5.1 图像生成

#### 5.1.1 文本到图像

```http
POST /api/v1/ai/image/generate
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "prompt": "一位古代中国哲学家在山顶冥想，水墨画风格",
  "negative_prompt": "现代建筑，汽车，手机",
  "model": "stable-diffusion-xl",
  "parameters": {
    "width": 1024,
    "height": 1024,
    "steps": 50,
    "guidance_scale": 7.5,
    "seed": 42,
    "sampler": "DPM++ 2M Karras"
  },
  "style": {
    "art_style": "chinese_ink_painting",
    "color_palette": "monochrome",
    "mood": "serene"
  },
  "options": {
    "num_images": 4,
    "safety_check": true,
    "upscale": false
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "generation_id": "img_gen_1234567890abcdef",
    "images": [
      {
        "url": "https://cdn.taishanglaojun.com/generated/img_001.jpg",
        "thumbnail_url": "https://cdn.taishanglaojun.com/generated/thumb_img_001.jpg",
        "width": 1024,
        "height": 1024,
        "seed": 42,
        "safety_score": 0.95
      }
    ],
    "parameters_used": {
      "model": "stable-diffusion-xl",
      "steps": 50,
      "guidance_scale": 7.5
    },
    "metadata": {
      "generation_time": 15.2,
      "gpu_time": 12.8,
      "model_version": "sdxl-1.0"
    }
  }
}
```

#### 5.1.2 图像到图像

```http
POST /api/v1/ai/image/img2img
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

prompt: "将这幅画转换为油画风格"
image: [binary file data]
strength: 0.7
guidance_scale: 7.5
steps: 30
```

#### 5.1.3 图像修复

```http
POST /api/v1/ai/image/inpaint
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

prompt: "移除图片中的现代元素"
image: [binary file data]
mask: [binary file data]
strength: 0.8
```

### 5.2 图像分析

#### 5.2.1 图像识别

```http
POST /api/v1/ai/image/analyze
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

image: [binary file data]
analysis_types: ["objects", "faces", "text", "scenes", "emotions"]
options: {"include_confidence": true, "detailed_description": true}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "analysis_id": "img_analysis_1234567890abcdef",
    "objects": [
      {
        "name": "person",
        "confidence": 0.95,
        "bounding_box": {
          "x": 100,
          "y": 150,
          "width": 200,
          "height": 300
        }
      }
    ],
    "faces": [
      {
        "confidence": 0.98,
        "age_range": "30-40",
        "gender": "male",
        "emotions": {
          "peaceful": 0.8,
          "contemplative": 0.7
        },
        "bounding_box": {
          "x": 150,
          "y": 180,
          "width": 100,
          "height": 120
        }
      }
    ],
    "text": [
      {
        "text": "道可道，非常道",
        "confidence": 0.92,
        "language": "zh-CN",
        "bounding_box": {
          "x": 50,
          "y": 50,
          "width": 300,
          "height": 40
        }
      }
    ],
    "scene": {
      "description": "一位古代哲学家在山顶冥想的场景",
      "tags": ["mountain", "meditation", "philosophy", "ancient"],
      "confidence": 0.88
    },
    "overall_description": "这是一幅描绘古代中国哲学家在山顶进行冥想的艺术作品..."
  }
}
```

#### 5.2.2 图像相似度比较

```http
POST /api/v1/ai/image/similarity
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

image1: [binary file data]
image2: [binary file data]
comparison_type: "semantic"  # semantic, visual, structural
```

### 5.3 图像编辑

#### 5.3.1 智能裁剪

```http
POST /api/v1/ai/image/smart-crop
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

image: [binary file data]
target_width: 800
target_height: 600
focus_areas: ["faces", "text", "main_objects"]
```

#### 5.3.2 风格转换

```http
POST /api/v1/ai/image/style-transfer
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

content_image: [binary file data]
style_image: [binary file data]
strength: 0.8
preserve_content: true
```

## 6. 语音服务

### 6.1 语音识别 (ASR)

#### 6.1.1 实时语音识别

```http
POST /api/v1/ai/speech/recognize/stream
Content-Type: audio/wav
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

[binary audio data]
```

**WebSocket连接示例：**

```javascript
const ws = new WebSocket('wss://api.taishanglaojun.com/v1/ai/speech/recognize/ws');

ws.onopen = function() {
  // 发送配置
  ws.send(JSON.stringify({
    type: 'config',
    language: 'zh-CN',
    sample_rate: 16000,
    encoding: 'LINEAR16',
    enable_punctuation: true,
    enable_word_timestamps: true
  }));
};

ws.onmessage = function(event) {
  const result = JSON.parse(event.data);
  console.log('识别结果:', result);
};
```

#### 6.1.2 批量语音识别

```http
POST /api/v1/ai/speech/recognize/batch
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

audio: [binary file data]
language: "zh-CN"
model: "whisper-large-v3"
options: {
  "enable_diarization": true,
  "enable_punctuation": true,
  "enable_word_timestamps": true,
  "filter_profanity": false
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "transcription_id": "asr_1234567890abcdef",
    "text": "太上老君是道教的重要神祇，被尊为道德天尊。",
    "language": "zh-CN",
    "confidence": 0.95,
    "duration": 8.5,
    "segments": [
      {
        "text": "太上老君是道教的重要神祇",
        "start_time": 0.0,
        "end_time": 3.2,
        "confidence": 0.96
      },
      {
        "text": "被尊为道德天尊",
        "start_time": 3.5,
        "end_time": 6.8,
        "confidence": 0.94
      }
    ],
    "words": [
      {
        "word": "太上老君",
        "start_time": 0.0,
        "end_time": 1.2,
        "confidence": 0.98
      }
    ],
    "speakers": [
      {
        "speaker_id": "speaker_1",
        "segments": [0, 1]
      }
    ]
  }
}
```

### 6.2 语音合成 (TTS)

#### 6.2.1 文本转语音

```http
POST /api/v1/ai/speech/synthesize
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "text": "太上老君，道德天尊，万物之始，无为而治。",
  "voice": {
    "name": "zh-CN-XiaoxiaoNeural",
    "gender": "female",
    "age": "adult",
    "style": "calm",
    "speed": 1.0,
    "pitch": 0.0,
    "volume": 0.8
  },
  "output_format": {
    "format": "mp3",
    "sample_rate": 22050,
    "bit_rate": 128
  },
  "options": {
    "add_pauses": true,
    "emphasize_keywords": ["太上老君", "道德天尊"],
    "emotion": "peaceful"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "synthesis_id": "tts_1234567890abcdef",
    "audio_url": "https://cdn.taishanglaojun.com/audio/tts_1234567890abcdef.mp3",
    "duration": 12.5,
    "format": "mp3",
    "sample_rate": 22050,
    "file_size": 156789,
    "metadata": {
      "voice_used": "zh-CN-XiaoxiaoNeural",
      "synthesis_time": 2.1,
      "character_count": 20
    }
  }
}
```

#### 6.2.2 SSML语音合成

```http
POST /api/v1/ai/speech/synthesize/ssml
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "ssml": "<speak><prosody rate='slow' pitch='low'>太上老君</prosody>，<break time='500ms'/>道德天尊。</speak>",
  "voice": "zh-CN-XiaoxiaoNeural",
  "output_format": "wav"
}
```

### 6.3 语音克隆

```http
POST /api/v1/ai/speech/clone
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

reference_audio: [binary file data]
target_text: "这是使用克隆语音合成的文本"
similarity_threshold: 0.8
```

## 7. 个性化推荐服务

### 7.1 内容推荐

```http
POST /api/v1/ai/recommendation/content
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "user_id": "usr_1234567890abcdef",
  "content_types": ["article", "video", "course", "book"],
  "categories": ["philosophy", "health", "learning", "culture"],
  "parameters": {
    "max_results": 20,
    "diversity_factor": 0.3,
    "novelty_factor": 0.2,
    "recency_weight": 0.1
  },
  "context": {
    "current_activity": "browsing_philosophy",
    "time_of_day": "evening",
    "device_type": "mobile"
  },
  "filters": {
    "difficulty_level": ["beginner", "intermediate"],
    "duration_range": {
      "min": 5,
      "max": 30
    },
    "language": "zh-CN"
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "recommendations": [
      {
        "item_id": "content_1234567890abcdef",
        "type": "article",
        "title": "道德经的现代解读",
        "description": "从现代视角重新理解老子的智慧",
        "category": "philosophy",
        "score": 0.95,
        "reasons": [
          "基于您对道教哲学的兴趣",
          "与您最近的阅读历史相关",
          "适合您的知识水平"
        ],
        "metadata": {
          "author": "张三",
          "duration": 15,
          "difficulty": "intermediate",
          "tags": ["道德经", "哲学", "传统文化"]
        }
      }
    ],
    "explanation": {
      "algorithm": "collaborative_filtering + content_based",
      "factors": {
        "user_preferences": 0.4,
        "item_popularity": 0.2,
        "contextual_relevance": 0.3,
        "diversity": 0.1
      }
    },
    "metadata": {
      "total_candidates": 10000,
      "filtered_candidates": 500,
      "recommendation_time": 0.15
    }
  }
}
```

### 7.2 学习路径推荐

```http
POST /api/v1/ai/recommendation/learning-path
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "user_id": "usr_1234567890abcdef",
  "learning_goals": [
    "understand_chinese_philosophy",
    "improve_meditation_practice",
    "learn_traditional_medicine"
  ],
  "current_level": {
    "philosophy": "beginner",
    "meditation": "intermediate",
    "medicine": "novice"
  },
  "time_constraints": {
    "daily_time": 60,
    "weekly_time": 420,
    "target_completion": "3_months"
  },
  "preferences": {
    "learning_style": "visual_auditory",
    "content_types": ["video", "interactive", "reading"],
    "difficulty_progression": "gradual"
  }
}
```

### 7.3 社交推荐

```http
POST /api/v1/ai/recommendation/social
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "user_id": "usr_1234567890abcdef",
  "recommendation_types": ["friends", "communities", "mentors", "study_groups"],
  "criteria": {
    "shared_interests": ["philosophy", "meditation", "health"],
    "skill_complementarity": true,
    "geographic_proximity": false,
    "activity_level": "active"
  },
  "max_results": 10
}
```

## 8. 内容审核服务

### 8.1 文本内容审核

```http
POST /api/v1/ai/moderation/text
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "text": "需要审核的文本内容",
  "context": {
    "content_type": "user_comment",
    "platform": "community",
    "user_age_group": "adult"
  },
  "checks": [
    "hate_speech",
    "harassment",
    "violence",
    "sexual_content",
    "spam",
    "misinformation",
    "personal_info"
  ],
  "options": {
    "severity_threshold": "medium",
    "include_explanation": true,
    "suggest_modifications": true
  }
}
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "moderation_id": "mod_1234567890abcdef",
    "overall_result": "approved",
    "confidence": 0.92,
    "flags": [
      {
        "category": "spam",
        "severity": "low",
        "confidence": 0.15,
        "flagged": false
      }
    ],
    "suggestions": [
      {
        "type": "improvement",
        "description": "建议使用更正式的语言表达",
        "modified_text": "修改后的文本建议..."
      }
    ],
    "metadata": {
      "processing_time": 0.3,
      "model_version": "moderation-v2.1",
      "language_detected": "zh-CN"
    }
  }
}
```

### 8.2 图像内容审核

```http
POST /api/v1/ai/moderation/image
Content-Type: multipart/form-data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

image: [binary file data]
checks: ["adult_content", "violence", "hate_symbols", "inappropriate_text"]
sensitivity: "high"
```

### 8.3 批量内容审核

```http
POST /api/v1/ai/moderation/batch
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "items": [
    {
      "id": "item_001",
      "type": "text",
      "content": "文本内容1"
    },
    {
      "id": "item_002",
      "type": "image",
      "url": "https://example.com/image.jpg"
    }
  ],
  "options": {
    "priority": "normal",
    "callback_url": "https://your-app.com/moderation-callback"
  }
}
```

## 9. 数据分析服务

### 9.1 用户行为分析

```http
POST /api/v1/ai/analytics/user-behavior
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "user_id": "usr_1234567890abcdef",
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "analysis_types": [
    "learning_patterns",
    "content_preferences",
    "engagement_trends",
    "skill_development"
  ],
  "options": {
    "include_predictions": true,
    "include_recommendations": true,
    "detail_level": "comprehensive"
  }
}
```

### 9.2 内容性能分析

```http
POST /api/v1/ai/analytics/content-performance
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "content_ids": ["content_001", "content_002", "content_003"],
  "metrics": [
    "engagement_rate",
    "completion_rate",
    "user_satisfaction",
    "learning_effectiveness"
  ],
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "segmentation": {
    "by_user_type": true,
    "by_device": true,
    "by_time_of_day": true
  }
}
```

### 9.3 预测分析

```http
POST /api/v1/ai/analytics/predictions
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "prediction_type": "user_churn",
  "target_users": ["usr_001", "usr_002"],
  "prediction_horizon": "30_days",
  "features": [
    "engagement_frequency",
    "session_duration",
    "content_completion_rate",
    "social_activity"
  ],
  "options": {
    "include_confidence_intervals": true,
    "include_feature_importance": true,
    "include_recommendations": true
  }
}
```

## 10. 模型管理

### 10.1 获取可用模型

```http
GET /api/v1/ai/models?category=text_generation&status=active
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应示例：**

```json
{
  "success": true,
  "data": {
    "models": [
      {
        "model_id": "gpt-4-turbo",
        "name": "GPT-4 Turbo",
        "category": "text_generation",
        "description": "最新的大语言模型，支持多模态输入",
        "capabilities": [
          "text_generation",
          "conversation",
          "code_generation",
          "analysis"
        ],
        "parameters": {
          "max_tokens": 128000,
          "context_window": 128000,
          "supports_streaming": true,
          "supports_function_calling": true
        },
        "pricing": {
          "input_tokens": 0.01,
          "output_tokens": 0.03,
          "currency": "USD",
          "unit": "per_1k_tokens"
        },
        "status": "active",
        "version": "2024-01-01",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 10.2 模型性能监控

```http
GET /api/v1/ai/models/{model_id}/metrics?time_range=24h
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 10.3 自定义模型部署

```http
POST /api/v1/ai/models/deploy
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "model_name": "custom-philosophy-model",
  "model_type": "text_generation",
  "model_file_url": "https://storage.example.com/models/philosophy-model.bin",
  "config": {
    "max_tokens": 4096,
    "temperature_range": [0.1, 2.0],
    "batch_size": 8
  },
  "deployment_options": {
    "auto_scaling": true,
    "min_instances": 1,
    "max_instances": 10,
    "gpu_type": "A100"
  }
}
```

## 11. 错误处理

### 11.1 AI服务特定错误

```typescript
enum AIServiceErrorCode {
  // 模型错误
  MODEL_NOT_AVAILABLE = 'AI1001',
  MODEL_OVERLOADED = 'AI1002',
  MODEL_TIMEOUT = 'AI1003',
  INVALID_MODEL_PARAMETERS = 'AI1004',
  
  // 内容错误
  CONTENT_TOO_LONG = 'AI2001',
  CONTENT_FILTERED = 'AI2002',
  INVALID_CONTENT_FORMAT = 'AI2003',
  CONTENT_GENERATION_FAILED = 'AI2004',
  
  // 配额错误
  QUOTA_EXCEEDED = 'AI3001',
  RATE_LIMIT_EXCEEDED = 'AI3002',
  INSUFFICIENT_CREDITS = 'AI3003',
  
  // 处理错误
  PROCESSING_FAILED = 'AI4001',
  INVALID_INPUT_FORMAT = 'AI4002',
  UNSUPPORTED_OPERATION = 'AI4003',
  
  // 资源错误
  INSUFFICIENT_RESOURCES = 'AI5001',
  STORAGE_FULL = 'AI5002',
  NETWORK_ERROR = 'AI5003',
}
```

## 12. 性能优化

### 12.1 缓存策略

```yaml
# AI服务缓存配置
cache_strategy:
  model_outputs:
    ttl: 3600
    key_pattern: "ai:output:{model}:{hash}"
    conditions:
      - deterministic_parameters
      - non_streaming_request
      
  embeddings:
    ttl: 86400
    key_pattern: "ai:embedding:{model}:{text_hash}"
    
  knowledge_base:
    ttl: 7200
    key_pattern: "ai:kb:{query_hash}"
    
  user_preferences:
    ttl: 1800
    key_pattern: "ai:prefs:{user_id}"
```

### 12.2 批处理优化

```http
POST /api/v1/ai/batch/text-generation
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "requests": [
    {
      "id": "req_001",
      "prompt": "第一个请求的提示词",
      "parameters": {"max_tokens": 100}
    },
    {
      "id": "req_002", 
      "prompt": "第二个请求的提示词",
      "parameters": {"max_tokens": 200}
    }
  ],
  "batch_options": {
    "priority": "normal",
    "callback_url": "https://your-app.com/batch-callback",
    "max_wait_time": 300
  }
}
```

## 13. 监控与指标

### 13.1 关键指标

```yaml
# AI服务监控指标
metrics:
  performance:
    - name: "model_response_time"
      description: "模型响应时间"
      target: "< 2s"
      
    - name: "generation_success_rate"
      description: "生成成功率"
      target: "> 99%"
      
    - name: "content_quality_score"
      description: "内容质量评分"
      target: "> 4.0/5.0"
      
  usage:
    - name: "tokens_per_second"
      description: "每秒处理token数"
      
    - name: "concurrent_requests"
      description: "并发请求数"
      
    - name: "model_utilization"
      description: "模型利用率"
      target: "60-80%"
      
  quality:
    - name: "content_safety_score"
      description: "内容安全评分"
      target: "> 95%"
      
    - name: "user_satisfaction"
      description: "用户满意度"
      target: "> 4.5/5.0"
```

## 14. 相关文档

- [API概览文档](../api-overview.md)
- [用户服务API](./user-service-api.md)
- [学习服务API](./learning-service-api.md)
- [模型训练指南](../model-training-guide.md)
- [AI安全最佳实践](../ai-security-guide.md)