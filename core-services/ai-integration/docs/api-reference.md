# 太上老君AI平台 - 高级AI功能API参考

## 概述

本文档详细描述了太上老君AI平台高级AI功能模块的所有API接口，包括请求格式、响应格式、错误码和使用示例。

**基础URL**: `http://localhost:8080/api/v1/advanced-ai`

**API版本**: v1.0.0

**认证方式**: JWT Bearer Token (可选)

**内容类型**: `application/json`

## 通用响应格式

### 成功响应

```json
{
  "request_id": "string",
  "success": true,
  "result": {},
  "confidence": 0.95,
  "process_time": "1.5s",
  "used_capabilities": ["agi", "reasoning"],
  "metadata": {},
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 错误响应

```json
{
  "error": "Error message",
  "details": "Detailed error description",
  "code": "ERROR_CODE",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req_123456789"
}
```

## 错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| `INVALID_REQUEST` | 400 | 请求格式无效 |
| `MISSING_PARAMETER` | 400 | 缺少必需参数 |
| `INVALID_PARAMETER` | 400 | 参数值无效 |
| `UNAUTHORIZED` | 401 | 未授权访问 |
| `FORBIDDEN` | 403 | 禁止访问 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `REQUEST_TOO_LARGE` | 413 | 请求体过大 |
| `RATE_LIMIT_EXCEEDED` | 429 | 请求频率超限 |
| `INTERNAL_ERROR` | 500 | 内部服务器错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 |
| `TIMEOUT` | 504 | 请求超时 |

## 核心API

### 1. 通用AI请求处理

处理各种类型的AI请求，支持多种能力模式。

**端点**: `POST /process`

**请求体**:
```json
{
  "id": "string (optional)",
  "type": "string (required)",
  "capability": "string (optional)",
  "input": "object (required)",
  "context": "object (optional)",
  "requirements": "object (optional)",
  "priority": "integer (optional, 1-5)",
  "timeout": "string (optional, duration format)"
}
```

**参数说明**:

| 参数 | 类型 | 必需 | 描述 | 示例 |
|------|------|------|------|------|
| `id` | string | 否 | 请求唯一标识符 | `"req_123456789"` |
| `type` | string | 是 | 请求类型 | `"reasoning"`, `"planning"`, `"generation"` |
| `capability` | string | 否 | 使用的能力模式 | `"agi"`, `"meta_learning"`, `"self_evolution"`, `"hybrid"` |
| `input` | object | 是 | 输入数据 | `{"text": "问题描述"}` |
| `context` | object | 否 | 上下文信息 | `{"domain": "science"}` |
| `requirements` | object | 否 | 特殊要求 | `{"max_length": 500}` |
| `priority` | integer | 否 | 优先级 (1-5) | `1` (最高) - `5` (最低) |
| `timeout` | string | 否 | 超时时间 | `"30s"`, `"5m"` |

**支持的请求类型**:

- `reasoning` - 推理任务
- `planning` - 规划任务
- `learning` - 学习任务
- `generation` - 生成任务
- `analysis` - 分析任务
- `classification` - 分类任务
- `question_answering` - 问答任务
- `summarization` - 摘要任务
- `translation` - 翻译任务
- `optimization` - 优化任务

**支持的能力模式**:

- `agi` - AGI能力集成
- `meta_learning` - 元学习引擎
- `self_evolution` - 自我进化系统
- `hybrid` - 混合模式 (默认)

**响应示例**:
```json
{
  "request_id": "req_123456789",
  "success": true,
  "result": {
    "answer": "基于深度学习的推理结果...",
    "reasoning_steps": [
      "步骤1: 分析问题",
      "步骤2: 搜索相关知识",
      "步骤3: 推理得出结论"
    ],
    "alternatives": [
      {
        "option": "方案A",
        "confidence": 0.85
      }
    ]
  },
  "confidence": 0.92,
  "process_time": "2.5s",
  "used_capabilities": ["agi", "reasoning"],
  "metadata": {
    "reasoning_depth": 5,
    "explored_alternatives": 3,
    "knowledge_sources": ["scientific_papers", "expert_knowledge"]
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 2. AGI任务处理

专门处理AGI相关的复杂任务。

**端点**: `POST /agi/task`

**请求体**:
```json
{
  "type": "string (required)",
  "input": "object (required)",
  "context": "object (optional)",
  "requirements": "object (optional)",
  "priority": "integer (optional)",
  "timeout": "integer (optional, seconds)"
}
```

**AGI任务类型**:

- `reasoning` - 复杂推理
- `planning` - 智能规划
- `creative_generation` - 创意生成
- `multimodal_analysis` - 多模态分析
- `metacognitive_reflection` - 元认知反思
- `adaptive_learning` - 自适应学习

**示例请求**:
```json
{
  "type": "planning",
  "input": {
    "goal": "设计一个可持续发展的城市交通系统",
    "constraints": {
      "budget": 5000000,
      "timeline": "2年",
      "population": 1000000,
      "environmental_impact": "minimal"
    },
    "current_state": {
      "existing_infrastructure": "传统公交系统",
      "traffic_congestion": "严重",
      "pollution_level": "高"
    }
  },
  "context": {
    "city_type": "大都市",
    "climate": "温带",
    "geography": "平原"
  },
  "requirements": {
    "detail_level": "high",
    "include_timeline": true,
    "include_risk_assessment": true,
    "include_alternatives": true
  },
  "priority": 1,
  "timeout": 300
}
```

**响应示例**:
```json
{
  "request_id": "agi_task_123456789",
  "success": true,
  "result": {
    "plan": {
      "phases": [
        {
          "phase": 1,
          "name": "基础设施评估",
          "duration": "3个月",
          "tasks": [
            "现有交通系统分析",
            "环境影响评估",
            "技术可行性研究"
          ],
          "resources": {
            "budget": 500000,
            "personnel": 20
          }
        }
      ],
      "total_cost": 4800000,
      "expected_outcomes": [
        "减少交通拥堵50%",
        "降低碳排放60%",
        "提高出行效率40%"
      ]
    },
    "risk_assessment": {
      "high_risks": [
        {
          "risk": "技术实施困难",
          "probability": 0.3,
          "impact": "high",
          "mitigation": "分阶段实施，技术验证"
        }
      ]
    },
    "alternatives": [
      {
        "name": "电动公交优先方案",
        "cost": 3000000,
        "timeline": "18个月",
        "pros": ["成本较低", "实施简单"],
        "cons": ["效果有限"]
      }
    ]
  },
  "confidence": 0.88,
  "process_time": "45.2s",
  "used_capabilities": ["agi", "planning", "reasoning"],
  "metadata": {
    "planning_depth": 8,
    "considered_factors": 25,
    "simulation_runs": 100
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 3. 元学习请求

触发元学习过程，快速适应新任务。

**端点**: `POST /meta-learning/learn`

**请求体**:
```json
{
  "task_type": "string (required)",
  "domain": "string (required)",
  "data": "array (required)",
  "strategy": "string (optional)",
  "parameters": "object (optional)"
}
```

**参数说明**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `task_type` | string | 是 | 学习任务类型 |
| `domain` | string | 是 | 应用领域 |
| `data` | array | 是 | 训练数据 |
| `strategy` | string | 否 | 学习策略 |
| `parameters` | object | 否 | 策略参数 |

**支持的任务类型**:

- `classification` - 分类任务
- `regression` - 回归任务
- `generation` - 生成任务
- `reinforcement_learning` - 强化学习
- `few_shot_learning` - 少样本学习
- `transfer_learning` - 迁移学习

**支持的学习策略**:

- `gradient_based` - 基于梯度的元学习
- `model_agnostic` - 模型无关元学习
- `memory_augmented` - 记忆增强学习
- `few_shot` - 少样本学习
- `transfer_learning` - 迁移学习
- `online_adaptation` - 在线适应

**示例请求**:
```json
{
  "task_type": "few_shot_learning",
  "domain": "medical_diagnosis",
  "data": [
    {
      "input": {
        "symptoms": ["发热", "咳嗽", "乏力"],
        "age": 35,
        "gender": "male",
        "medical_history": ["高血压"]
      },
      "label": "流感",
      "metadata": {
        "confidence": 0.9,
        "source": "expert_diagnosis"
      }
    }
  ],
  "strategy": "few_shot",
  "parameters": {
    "shots": 5,
    "ways": 3,
    "episodes": 1000,
    "learning_rate": 0.001,
    "adaptation_steps": 5
  }
}
```

**响应示例**:
```json
{
  "request_id": "meta_learn_123456789",
  "success": true,
  "result": {
    "model_id": "meta_model_abc123",
    "learning_performance": {
      "accuracy": 0.92,
      "loss": 0.08,
      "convergence_steps": 850
    },
    "adaptation_capability": {
      "few_shot_accuracy": 0.88,
      "adaptation_speed": "fast",
      "generalization_score": 0.85
    },
    "learned_features": [
      "症状模式识别",
      "年龄相关性分析",
      "病史关联性"
    ],
    "meta_knowledge": {
      "domain_patterns": ["症状聚类", "诊断规则"],
      "transfer_potential": 0.9
    }
  },
  "confidence": 0.91,
  "process_time": "120.5s",
  "used_capabilities": ["meta_learning", "few_shot"],
  "metadata": {
    "strategy_used": "few_shot",
    "data_samples": 100,
    "training_episodes": 1000
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 4. 自我进化优化

触发系统自我进化和优化过程。

**端点**: `POST /evolution/optimize`

**请求体**:
```json
{
  "optimization_targets": "array (required)",
  "strategy": "string (optional)",
  "parameters": "object (optional)"
}
```

**优化目标格式**:
```json
{
  "metric": "string (required)",
  "weight": "number (required, 0-1)",
  "target_value": "number (required)",
  "current_value": "number (optional)",
  "direction": "string (optional, 'maximize' or 'minimize')"
}
```

**支持的优化指标**:

- `accuracy` - 准确率
- `precision` - 精确率
- `recall` - 召回率
- `f1_score` - F1分数
- `efficiency` - 效率
- `speed` - 速度
- `memory_usage` - 内存使用
- `energy_consumption` - 能耗
- `robustness` - 鲁棒性
- `interpretability` - 可解释性

**支持的进化策略**:

- `genetic` - 遗传算法
- `neuro_evolution` - 神经进化
- `gradient_free` - 无梯度优化
- `hybrid` - 混合策略
- `reinforcement` - 强化学习
- `swarm_intelligence` - 群体智能

**示例请求**:
```json
{
  "optimization_targets": [
    {
      "metric": "accuracy",
      "weight": 0.4,
      "target_value": 0.95,
      "current_value": 0.88,
      "direction": "maximize"
    },
    {
      "metric": "inference_speed",
      "weight": 0.3,
      "target_value": 50,
      "current_value": 120,
      "direction": "minimize"
    },
    {
      "metric": "model_size",
      "weight": 0.2,
      "target_value": 10,
      "current_value": 25,
      "direction": "minimize"
    },
    {
      "metric": "robustness",
      "weight": 0.1,
      "target_value": 0.9,
      "current_value": 0.75,
      "direction": "maximize"
    }
  ],
  "strategy": "genetic",
  "parameters": {
    "population_size": 50,
    "generations": 100,
    "mutation_rate": 0.1,
    "crossover_rate": 0.8,
    "selection_pressure": 2.0,
    "elite_ratio": 0.1
  }
}
```

**响应示例**:
```json
{
  "request_id": "evolution_123456789",
  "success": true,
  "result": {
    "evolution_id": "evo_task_abc123",
    "status": "started",
    "estimated_duration": "2小时",
    "initial_population": {
      "size": 50,
      "diversity_score": 0.85
    },
    "optimization_progress": {
      "current_generation": 0,
      "best_fitness": 0.72,
      "average_fitness": 0.65,
      "improvement_rate": 0.0
    },
    "predicted_outcomes": {
      "accuracy_improvement": 0.06,
      "speed_improvement": 0.58,
      "size_reduction": 0.6,
      "robustness_improvement": 0.15
    }
  },
  "confidence": 0.87,
  "process_time": "5.2s",
  "used_capabilities": ["self_evolution", "genetic"],
  "metadata": {
    "strategy": "genetic",
    "population_initialized": true,
    "fitness_function_defined": true
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

## 监控API

### 1. 系统状态

获取系统整体状态和健康度。

**端点**: `GET /status`

**响应示例**:
```json
{
  "overall_health": 0.92,
  "active_requests": 15,
  "total_requests": 1250,
  "success_rate": 0.94,
  "avg_response_time": "1.8s",
  "uptime": "72h30m15s",
  "capabilities": {
    "agi": {
      "status": "active",
      "health": 0.95,
      "load": 0.6,
      "active_tasks": 8,
      "completed_tasks": 450
    },
    "meta_learning": {
      "status": "active",
      "health": 0.88,
      "load": 0.4,
      "active_learning_sessions": 3,
      "completed_sessions": 120
    },
    "self_evolution": {
      "status": "active",
      "health": 0.93,
      "load": 0.2,
      "active_evolutions": 1,
      "completed_evolutions": 25
    },
    "hybrid": {
      "status": "active",
      "health": 0.91,
      "load": 0.5,
      "hybrid_requests": 200
    }
  },
  "resource_usage": {
    "cpu": 0.65,
    "memory": 0.72,
    "gpu": 0.80,
    "disk": 0.45,
    "network": 0.30
  },
  "performance_metrics": {
    "requests_per_second": 25.5,
    "avg_latency": "1.2s",
    "p95_latency": "3.5s",
    "p99_latency": "8.2s"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 2. 性能指标

获取详细的性能指标历史数据。

**端点**: `GET /metrics`

**查询参数**:

| 参数 | 类型 | 必需 | 描述 | 默认值 |
|------|------|------|------|--------|
| `limit` | integer | 否 | 返回记录数量 | 100 |
| `capability` | string | 否 | 过滤特定能力 | 全部 |
| `time_range` | string | 否 | 时间范围 | 1h |
| `metric_type` | string | 否 | 指标类型 | 全部 |

**示例请求**:
```
GET /metrics?limit=50&capability=agi&time_range=24h&metric_type=performance
```

**响应示例**:
```json
{
  "metrics": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "capability": "agi",
      "metric_type": "performance",
      "values": {
        "response_time": 1.5,
        "success_rate": 0.95,
        "confidence": 0.88,
        "resource_usage": 0.65
      }
    }
  ],
  "count": 50,
  "time_range": "24h",
  "aggregated_stats": {
    "avg_response_time": 1.8,
    "avg_success_rate": 0.93,
    "avg_confidence": 0.87,
    "total_requests": 1250
  }
}
```

### 3. 健康检查

简单的健康检查端点。

**端点**: `GET /health`

**响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "72h30m15s",
  "overall_health": 0.92,
  "active_requests": 15,
  "success_rate": 0.94,
  "components": {
    "agi_service": "healthy",
    "meta_learning": "healthy",
    "self_evolution": "healthy",
    "database": "healthy",
    "cache": "healthy"
  }
}
```

## 管理API

### 1. 配置管理

#### 获取配置

**端点**: `GET /config`

**响应示例**:
```json
{
  "enable_agi": true,
  "enable_meta_learning": true,
  "enable_self_evolution": true,
  "enable_hybrid_mode": true,
  "max_concurrent_requests": 50,
  "default_timeout": "30s",
  "performance_monitoring": true,
  "auto_optimization": true,
  "log_level": "info",
  "security": {
    "enable_authentication": true,
    "enable_rate_limiting": true
  }
}
```

#### 更新配置

**端点**: `PUT /config`

**请求体**:
```json
{
  "max_concurrent_requests": 100,
  "default_timeout": "60s",
  "log_level": "debug",
  "auto_optimization": false
}
```

**响应示例**:
```json
{
  "message": "Configuration updated successfully",
  "updated_fields": [
    "max_concurrent_requests",
    "default_timeout",
    "log_level",
    "auto_optimization"
  ],
  "config": {
    "max_concurrent_requests": 100,
    "default_timeout": "60s",
    "log_level": "debug",
    "auto_optimization": false
  }
}
```

### 2. 能力管理

#### 获取可用能力

**端点**: `GET /capabilities`

**响应示例**:
```json
{
  "agi": {
    "enabled": true,
    "capabilities": [
      "reasoning",
      "planning",
      "learning",
      "creativity",
      "multimodal",
      "metacognition"
    ],
    "status": "active",
    "health": 0.95,
    "load": 0.6
  },
  "meta_learning": {
    "enabled": true,
    "strategies": [
      "gradient_based",
      "model_agnostic",
      "memory_augmented",
      "few_shot",
      "transfer_learning",
      "online_adaptation"
    ],
    "status": "active",
    "health": 0.88,
    "load": 0.4
  },
  "self_evolution": {
    "enabled": true,
    "strategies": [
      "genetic",
      "neuro_evolution",
      "gradient_free",
      "hybrid",
      "reinforcement",
      "swarm_intelligence"
    ],
    "status": "active",
    "health": 0.93,
    "load": 0.2
  },
  "hybrid": {
    "enabled": true,
    "status": "active",
    "health": 0.91,
    "load": 0.5
  }
}
```

#### 启用能力

**端点**: `POST /capabilities/{capability}/enable`

**路径参数**:
- `capability`: 能力名称 (`agi`, `meta_learning`, `self_evolution`, `hybrid`)

**响应示例**:
```json
{
  "message": "Capability 'agi' enabled successfully",
  "capability": "agi",
  "enabled": true,
  "status": "active"
}
```

#### 禁用能力

**端点**: `POST /capabilities/{capability}/disable`

**响应示例**:
```json
{
  "message": "Capability 'agi' disabled successfully",
  "capability": "agi",
  "enabled": false,
  "status": "inactive"
}
```

### 3. 历史记录

#### 获取请求历史

**端点**: `GET /history`

**查询参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `limit` | integer | 否 | 返回记录数量 |
| `capability` | string | 否 | 过滤特定能力 |
| `status` | string | 否 | 过滤状态 (`success`, `failed`) |
| `start_time` | string | 否 | 开始时间 (ISO 8601) |
| `end_time` | string | 否 | 结束时间 (ISO 8601) |

**示例请求**:
```
GET /history?limit=20&capability=agi&status=success&start_time=2024-01-15T00:00:00Z
```

**响应示例**:
```json
{
  "history": [
    {
      "request_id": "req_123456789",
      "capability": "agi",
      "type": "reasoning",
      "success": true,
      "confidence": 0.95,
      "process_time": "1.2s",
      "used_capabilities": ["agi", "reasoning"],
      "input_size": 1024,
      "output_size": 2048,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 20,
  "total": 1250,
  "filters": {
    "capability": "agi",
    "status": "success",
    "start_time": "2024-01-15T00:00:00Z"
  }
}
```

#### 获取统计信息

**端点**: `GET /statistics`

**查询参数**:

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `time_range` | string | 否 | 统计时间范围 |
| `group_by` | string | 否 | 分组方式 (`capability`, `type`, `hour`, `day`) |

**响应示例**:
```json
{
  "total_requests": 1250,
  "success_rate": 0.94,
  "avg_response_time": "1.8s",
  "active_requests": 15,
  "overall_health": 0.92,
  "capability_stats": {
    "agi": {
      "requests": 500,
      "success_rate": 0.96,
      "avg_confidence": 0.89,
      "avg_response_time": "2.1s"
    },
    "meta_learning": {
      "requests": 375,
      "success_rate": 0.92,
      "avg_confidence": 0.85,
      "avg_response_time": "5.2s"
    },
    "self_evolution": {
      "requests": 250,
      "success_rate": 0.95,
      "avg_confidence": 0.91,
      "avg_response_time": "15.8s"
    },
    "hybrid": {
      "requests": 125,
      "success_rate": 0.94,
      "avg_confidence": 0.93,
      "avg_response_time": "3.5s"
    }
  },
  "performance_trends": {
    "last_hour": {
      "requests": 120,
      "success_rate": 0.95,
      "avg_latency": "1.2s"
    },
    "last_day": {
      "requests": 2880,
      "success_rate": 0.93,
      "avg_latency": "1.8s"
    },
    "last_week": {
      "requests": 20160,
      "success_rate": 0.91,
      "avg_latency": "2.1s"
    }
  },
  "resource_usage": {
    "cpu": "65%",
    "memory": "72%",
    "gpu": "80%",
    "disk": "45%",
    "network": "30%"
  },
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## 系统管理API

### 1. 初始化服务

**端点**: `POST /initialize`

**请求体**:
```json
{
  "config": {
    "enable_agi": true,
    "enable_meta_learning": true,
    "enable_self_evolution": true
  },
  "force": false
}
```

**响应示例**:
```json
{
  "message": "Service initialized successfully",
  "initialized": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "components": {
    "agi_service": "initialized",
    "meta_learning": "initialized",
    "self_evolution": "initialized",
    "database": "connected",
    "cache": "ready"
  }
}
```

### 2. 关闭服务

**端点**: `POST /shutdown`

**请求体**:
```json
{
  "graceful": true,
  "timeout": 30
}
```

**响应示例**:
```json
{
  "message": "Service shutdown initiated",
  "graceful": true,
  "timeout": 30,
  "timestamp": "2024-01-15T10:30:00Z",
  "active_requests": 5,
  "estimated_completion": "2024-01-15T10:30:30Z"
}
```

### 3. 重置服务

**端点**: `POST /reset`

**请求体**:
```json
{
  "reset_data": false,
  "reset_config": false,
  "reset_metrics": true
}
```

**响应示例**:
```json
{
  "message": "Service reset successfully",
  "reset_data": false,
  "reset_config": false,
  "reset_metrics": true,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 认证和授权

### JWT认证

如果启用了认证，需要在请求头中包含JWT令牌：

```http
Authorization: Bearer <jwt_token>
```

### 权限要求

不同的API端点需要不同的权限：

| 端点类型 | 所需权限 |
|----------|----------|
| 核心API | `ai:read`, `ai:write` |
| 监控API | `ai:read` |
| 管理API | `ai:admin` |
| 系统管理API | `system:admin` |

## 限流

系统支持基于IP和用户的限流：

- **默认限制**: 每分钟100个请求
- **认证用户**: 每分钟500个请求
- **管理员**: 无限制

超出限制时返回HTTP 429状态码。

## 使用示例

### Python客户端示例

```python
import requests
import json

class AdvancedAIClient:
    def __init__(self, base_url, token=None):
        self.base_url = base_url
        self.headers = {"Content-Type": "application/json"}
        if token:
            self.headers["Authorization"] = f"Bearer {token}"
    
    def process_request(self, request_type, input_data, capability=None, **kwargs):
        """处理通用AI请求"""
        data = {
            "type": request_type,
            "input": input_data,
            **kwargs
        }
        if capability:
            data["capability"] = capability
        
        response = requests.post(
            f"{self.base_url}/process",
            headers=self.headers,
            data=json.dumps(data)
        )
        return response.json()
    
    def agi_task(self, task_type, input_data, **kwargs):
        """处理AGI任务"""
        data = {
            "type": task_type,
            "input": input_data,
            **kwargs
        }
        
        response = requests.post(
            f"{self.base_url}/agi/task",
            headers=self.headers,
            data=json.dumps(data)
        )
        return response.json()
    
    def meta_learning(self, task_type, domain, data, strategy=None, **kwargs):
        """触发元学习"""
        request_data = {
            "task_type": task_type,
            "domain": domain,
            "data": data,
            **kwargs
        }
        if strategy:
            request_data["strategy"] = strategy
        
        response = requests.post(
            f"{self.base_url}/meta-learning/learn",
            headers=self.headers,
            data=json.dumps(request_data)
        )
        return response.json()
    
    def trigger_evolution(self, optimization_targets, strategy=None, **kwargs):
        """触发自我进化"""
        data = {
            "optimization_targets": optimization_targets,
            **kwargs
        }
        if strategy:
            data["strategy"] = strategy
        
        response = requests.post(
            f"{self.base_url}/evolution/optimize",
            headers=self.headers,
            data=json.dumps(data)
        )
        return response.json()
    
    def get_status(self):
        """获取系统状态"""
        response = requests.get(
            f"{self.base_url}/status",
            headers=self.headers
        )
        return response.json()

# 使用示例
client = AdvancedAIClient("http://localhost:8080/api/v1/advanced-ai")

# 推理任务
result = client.process_request(
    request_type="reasoning",
    input_data={"problem": "如何解决气候变化问题？"},
    capability="agi",
    requirements={"detail_level": "high"}
)

# AGI规划任务
plan = client.agi_task(
    task_type="planning",
    input_data={
        "goal": "建设智慧城市",
        "constraints": {"budget": 1000000, "timeline": "5年"}
    },
    requirements={"include_risks": True}
)

# 元学习
learning_result = client.meta_learning(
    task_type="classification",
    domain="medical_diagnosis",
    data=[{"input": {...}, "label": "..."}],
    strategy="few_shot"
)

# 自我进化
evolution_result = client.trigger_evolution(
    optimization_targets=[
        {"metric": "accuracy", "weight": 0.6, "target_value": 0.95},
        {"metric": "speed", "weight": 0.4, "target_value": 100}
    ],
    strategy="genetic"
)

# 获取状态
status = client.get_status()
print(f"系统健康度: {status['overall_health']}")
```

### JavaScript客户端示例

```javascript
class AdvancedAIClient {
    constructor(baseUrl, token = null) {
        this.baseUrl = baseUrl;
        this.headers = {
            'Content-Type': 'application/json'
        };
        if (token) {
            this.headers['Authorization'] = `Bearer ${token}`;
        }
    }
    
    async processRequest(requestType, inputData, capability = null, options = {}) {
        const data = {
            type: requestType,
            input: inputData,
            ...options
        };
        if (capability) {
            data.capability = capability;
        }
        
        const response = await fetch(`${this.baseUrl}/process`, {
            method: 'POST',
            headers: this.headers,
            body: JSON.stringify(data)
        });
        
        return await response.json();
    }
    
    async agiTask(taskType, inputData, options = {}) {
        const data = {
            type: taskType,
            input: inputData,
            ...options
        };
        
        const response = await fetch(`${this.baseUrl}/agi/task`, {
            method: 'POST',
            headers: this.headers,
            body: JSON.stringify(data)
        });
        
        return await response.json();
    }
    
    async getStatus() {
        const response = await fetch(`${this.baseUrl}/status`, {
            headers: this.headers
        });
        
        return await response.json();
    }
}

// 使用示例
const client = new AdvancedAIClient('http://localhost:8080/api/v1/advanced-ai');

// 异步处理
(async () => {
    try {
        const result = await client.processRequest(
            'generation',
            { prompt: '写一首关于AI的诗' },
            'agi',
            { requirements: { style: 'modern', length: 'short' } }
        );
        
        console.log('生成结果:', result.result);
        console.log('置信度:', result.confidence);
    } catch (error) {
        console.error('请求失败:', error);
    }
})();
```

## 版本兼容性

当前API版本为v1.0.0，遵循语义化版本控制：

- **主版本号**: 不兼容的API变更
- **次版本号**: 向后兼容的功能性新增
- **修订号**: 向后兼容的问题修正

## 更新日志

### v1.0.0 (2024-01-15)
- 初始API版本发布
- 支持AGI能力集成
- 支持元学习引擎
- 支持自我进化系统
- 提供完整的监控和管理API

---

*太上老君AI平台 - 让AI更智能，让智能更人性*