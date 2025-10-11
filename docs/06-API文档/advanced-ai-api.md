# 高级AI功能 API 文档

## 概述

本文档描述了太上老君项目中高级AI功能的API接口，包括AGI能力、元学习引擎和自我进化系统。

## 基础信息

- **基础URL**: `https://api.taishanglaojun.com/api/v1`
- **认证方式**: JWT Bearer Token
- **内容类型**: `application/json`
- **API版本**: v1.0.0

## 认证

所有API请求都需要在请求头中包含有效的JWT令牌：

```http
Authorization: Bearer <your-jwt-token>
```

## 通用响应格式

### 成功响应

```json
{
  "success": true,
  "data": {
    // 响应数据
  },
  "message": "操作成功",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req_123456789"
}
```

### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": "详细错误信息"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req_123456789"
}
```

## AGI 能力接口

### 1. 推理能力

#### POST /agi/reasoning

执行复杂推理任务。

**请求参数:**

```json
{
  "query": "需要推理的问题",
  "context": "相关上下文信息",
  "reasoning_type": "deductive|inductive|abductive",
  "max_steps": 10,
  "confidence_threshold": 0.8
}
```

**响应示例:**

```json
{
  "success": true,
  "data": {
    "reasoning_result": {
      "conclusion": "推理结论",
      "confidence": 0.92,
      "reasoning_steps": [
        {
          "step": 1,
          "description": "步骤描述",
          "evidence": ["证据1", "证据2"],
          "confidence": 0.85
        }
      ],
      "reasoning_type": "deductive",
      "execution_time": 1.23
    }
  }
}
```

### 2. 规划能力

#### POST /agi/planning

生成复杂任务的执行计划。

**请求参数:**

```json
{
  "goal": "目标描述",
  "constraints": ["约束条件1", "约束条件2"],
  "resources": {
    "time_limit": "2h",
    "budget": 1000,
    "tools": ["tool1", "tool2"]
  },
  "planning_algorithm": "hierarchical|sequential|parallel"
}
```

**响应示例:**

```json
{
  "success": true,
  "data": {
    "plan": {
      "goal": "目标描述",
      "total_steps": 5,
      "estimated_time": "1.5h",
      "estimated_cost": 800,
      "steps": [
        {
          "id": "step_1",
          "description": "步骤描述",
          "dependencies": [],
          "resources_required": {
            "time": "30m",
            "cost": 200
          },
          "success_criteria": "成功标准"
        }
      ],
      "risk_assessment": {
        "overall_risk": "medium",
        "risk_factors": ["风险因素1", "风险因素2"]
      }
    }
  }
}
```

### 3. 学习能力

#### POST /agi/learning

执行在线学习任务。

**请求参数:**

```json
{
  "learning_data": {
    "input": "学习输入数据",
    "expected_output": "期望输出",
    "feedback": "反馈信息"
  },
  "learning_type": "supervised|unsupervised|reinforcement",
  "adaptation_rate": 0.1
}
```

### 4. 创造能力

#### POST /agi/creativity

生成创新性内容或解决方案。

**请求参数:**

```json
{
  "prompt": "创作提示",
  "creativity_type": "text|image|music|solution",
  "style": "风格描述",
  "constraints": ["约束条件"],
  "novelty_threshold": 0.7
}
```

### 5. 多模态处理

#### POST /agi/multimodal

处理多模态输入数据。

**请求参数:**

```json
{
  "inputs": {
    "text": "文本输入",
    "image_url": "图片URL",
    "audio_url": "音频URL"
  },
  "task": "description|analysis|generation",
  "output_format": "text|json|structured"
}
```

## 元学习引擎接口

### 1. 策略学习

#### POST /meta-learning/strategy

学习和优化学习策略。

**请求参数:**

```json
{
  "task_type": "classification|regression|generation",
  "training_data": {
    "tasks": [
      {
        "id": "task_1",
        "data": "任务数据",
        "labels": "标签数据"
      }
    ]
  },
  "meta_algorithm": "maml|reptile|prototypical",
  "adaptation_steps": 5
}
```

**响应示例:**

```json
{
  "success": true,
  "data": {
    "strategy": {
      "id": "strategy_123",
      "algorithm": "maml",
      "performance_metrics": {
        "accuracy": 0.89,
        "adaptation_speed": 0.95,
        "generalization": 0.87
      },
      "learned_parameters": {
        "learning_rate": 0.001,
        "batch_size": 32,
        "architecture": "transformer"
      },
      "training_time": 45.6
    }
  }
}
```

### 2. 快速适应

#### POST /meta-learning/adapt

快速适应新任务。

**请求参数:**

```json
{
  "base_strategy_id": "strategy_123",
  "new_task": {
    "description": "新任务描述",
    "sample_data": "少量样本数据",
    "target_metric": "accuracy"
  },
  "adaptation_budget": {
    "max_iterations": 10,
    "max_time": "5m"
  }
}
```

### 3. 知识迁移

#### POST /meta-learning/transfer

在任务间迁移知识。

**请求参数:**

```json
{
  "source_task_id": "task_source",
  "target_task_id": "task_target",
  "transfer_method": "feature|parameter|gradient",
  "similarity_threshold": 0.6
}
```

## 自我进化系统接口

### 1. 性能评估

#### GET /self-evolution/performance

获取系统性能评估结果。

**响应示例:**

```json
{
  "success": true,
  "data": {
    "performance_metrics": {
      "overall_score": 0.87,
      "efficiency": 0.92,
      "accuracy": 0.85,
      "adaptability": 0.83,
      "robustness": 0.89
    },
    "component_scores": {
      "agi": 0.88,
      "meta_learning": 0.86,
      "nlp": 0.90
    },
    "trend_analysis": {
      "improvement_rate": 0.05,
      "stability": "high",
      "prediction": "继续改进"
    }
  }
}
```

### 2. 自动优化

#### POST /self-evolution/optimize

触发系统自动优化。

**请求参数:**

```json
{
  "optimization_target": "performance|efficiency|accuracy",
  "constraints": {
    "max_resource_usage": 0.8,
    "max_downtime": "5m",
    "preserve_accuracy": true
  },
  "optimization_algorithm": "genetic|gradient|bayesian"
}
```

**响应示例:**

```json
{
  "success": true,
  "data": {
    "optimization_id": "opt_789",
    "status": "running",
    "estimated_completion": "2024-01-15T11:00:00Z",
    "current_progress": 0.25,
    "improvements_found": [
      {
        "component": "agi_reasoning",
        "improvement": "算法参数优化",
        "expected_gain": 0.12
      }
    ]
  }
}
```

### 3. 进化历史

#### GET /self-evolution/history

获取系统进化历史。

**查询参数:**

- `start_date`: 开始日期
- `end_date`: 结束日期
- `component`: 组件名称（可选）
- `limit`: 返回数量限制

**响应示例:**

```json
{
  "success": true,
  "data": {
    "evolution_history": [
      {
        "timestamp": "2024-01-15T09:00:00Z",
        "version": "1.2.3",
        "changes": [
          {
            "component": "agi_reasoning",
            "type": "parameter_update",
            "description": "优化推理算法参数",
            "performance_impact": 0.08
          }
        ],
        "performance_before": 0.82,
        "performance_after": 0.87,
        "validation_results": {
          "test_accuracy": 0.89,
          "benchmark_score": 0.91
        }
      }
    ],
    "total_improvements": 15,
    "average_improvement": 0.06
  }
}
```

## 系统管理接口

### 1. 健康检查

#### GET /health

检查系统健康状态。

**响应示例:**

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "components": {
      "agi": "healthy",
      "meta_learning": "healthy",
      "self_evolution": "healthy",
      "database": "healthy",
      "cache": "healthy"
    },
    "metrics": {
      "uptime": "72h",
      "cpu_usage": 0.45,
      "memory_usage": 0.67,
      "response_time": 0.123
    }
  }
}
```

### 2. 系统配置

#### GET /config

获取系统配置信息。

#### PUT /config

更新系统配置。

**请求参数:**

```json
{
  "agi_config": {
    "reasoning_timeout": 30,
    "max_reasoning_steps": 20,
    "confidence_threshold": 0.8
  },
  "meta_learning_config": {
    "adaptation_rate": 0.1,
    "meta_batch_size": 16,
    "max_adaptation_steps": 10
  },
  "self_evolution_config": {
    "optimization_interval": "1h",
    "performance_threshold": 0.05,
    "auto_optimization": true
  }
}
```

### 3. 系统指标

#### GET /metrics

获取系统性能指标。

**响应示例:**

```json
{
  "success": true,
  "data": {
    "request_metrics": {
      "total_requests": 10000,
      "successful_requests": 9850,
      "error_rate": 0.015,
      "average_response_time": 0.234
    },
    "ai_metrics": {
      "agi_inferences": 5000,
      "meta_learning_adaptations": 150,
      "self_evolution_optimizations": 12
    },
    "resource_metrics": {
      "cpu_usage": 0.45,
      "memory_usage": 0.67,
      "disk_usage": 0.23,
      "network_io": 1024000
    }
  }
}
```

## WebSocket 接口

### 连接地址

```
wss://api.taishanglaojun.com/ws
```

### 认证

连接时需要在查询参数中提供JWT令牌：

```
wss://api.taishanglaojun.com/ws?token=<your-jwt-token>
```

### 消息格式

#### 客户端发送消息

```json
{
  "type": "request",
  "id": "req_123",
  "action": "agi_reasoning",
  "data": {
    "query": "推理问题",
    "context": "上下文"
  }
}
```

#### 服务端响应消息

```json
{
  "type": "response",
  "id": "req_123",
  "success": true,
  "data": {
    "result": "推理结果"
  }
}
```

#### 服务端推送消息

```json
{
  "type": "notification",
  "event": "optimization_completed",
  "data": {
    "optimization_id": "opt_789",
    "performance_improvement": 0.12
  }
}
```

## 错误代码

| 错误代码 | HTTP状态码 | 描述 |
|---------|-----------|------|
| AUTH_REQUIRED | 401 | 需要认证 |
| AUTH_INVALID | 401 | 认证无效 |
| PERMISSION_DENIED | 403 | 权限不足 |
| RESOURCE_NOT_FOUND | 404 | 资源不存在 |
| VALIDATION_ERROR | 400 | 请求参数验证失败 |
| RATE_LIMIT_EXCEEDED | 429 | 请求频率超限 |
| INTERNAL_ERROR | 500 | 内部服务器错误 |
| SERVICE_UNAVAILABLE | 503 | 服务不可用 |
| AGI_TIMEOUT | 408 | AGI推理超时 |
| META_LEARNING_FAILED | 422 | 元学习失败 |
| OPTIMIZATION_FAILED | 422 | 优化失败 |

## 使用示例

### Python 示例

```python
import requests
import json

# 配置
BASE_URL = "https://api.taishanglaojun.com/api/v1"
TOKEN = "your-jwt-token"

headers = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json"
}

# AGI推理示例
def agi_reasoning(query, context):
    url = f"{BASE_URL}/agi/reasoning"
    data = {
        "query": query,
        "context": context,
        "reasoning_type": "deductive",
        "max_steps": 10,
        "confidence_threshold": 0.8
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# 使用示例
result = agi_reasoning(
    "如果所有人都是凡人，苏格拉底是人，那么苏格拉底是什么？",
    "这是一个经典的三段论推理问题"
)
print(json.dumps(result, indent=2, ensure_ascii=False))
```

### JavaScript 示例

```javascript
const BASE_URL = "https://api.taishanglaojun.com/api/v1";
const TOKEN = "your-jwt-token";

const headers = {
    "Authorization": `Bearer ${TOKEN}`,
    "Content-Type": "application/json"
};

// 元学习适应示例
async function metaLearningAdapt(strategyId, newTask) {
    const url = `${BASE_URL}/meta-learning/adapt`;
    const data = {
        base_strategy_id: strategyId,
        new_task: newTask,
        adaptation_budget: {
            max_iterations: 10,
            max_time: "5m"
        }
    };
    
    const response = await fetch(url, {
        method: "POST",
        headers: headers,
        body: JSON.stringify(data)
    });
    
    return await response.json();
}

// 使用示例
metaLearningAdapt("strategy_123", {
    description: "图像分类任务",
    sample_data: "base64_encoded_images",
    target_metric: "accuracy"
}).then(result => {
    console.log(JSON.stringify(result, null, 2));
});
```

### WebSocket 示例

```javascript
const ws = new WebSocket(`wss://api.taishanglaojun.com/ws?token=${TOKEN}`);

ws.onopen = function(event) {
    console.log("WebSocket连接已建立");
    
    // 发送AGI推理请求
    ws.send(JSON.stringify({
        type: "request",
        id: "req_001",
        action: "agi_reasoning",
        data: {
            query: "分析这个商业案例的成功因素",
            context: "某科技公司的发展历程..."
        }
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log("收到消息:", message);
    
    if (message.type === "response") {
        console.log("推理结果:", message.data.result);
    } else if (message.type === "notification") {
        console.log("系统通知:", message.event, message.data);
    }
};

ws.onerror = function(error) {
    console.error("WebSocket错误:", error);
};

ws.onclose = function(event) {
    console.log("WebSocket连接已关闭");
};
```

## 最佳实践

### 1. 认证和安全

- 始终使用HTTPS进行API调用
- 定期轮换JWT令牌
- 不要在客户端代码中硬编码敏感信息
- 实施适当的速率限制

### 2. 错误处理

- 始终检查API响应的`success`字段
- 实施重试机制处理临时错误
- 记录错误信息用于调试
- 为用户提供友好的错误消息

### 3. 性能优化

- 使用适当的缓存策略
- 批量处理多个请求
- 使用WebSocket进行实时通信
- 监控API响应时间和错误率

### 4. 版本管理

- 在请求头中指定API版本
- 关注API变更通知
- 测试新版本兼容性
- 逐步迁移到新版本

## 支持和联系

- **技术支持**: api-support@taishanglaojun.com
- **文档更新**: docs@taishanglaojun.com
- **GitHub**: https://github.com/taishanglaojun/advanced-ai
- **开发者社区**: https://community.taishanglaojun.com

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 支持AGI基础能力
- 支持元学习引擎
- 支持自我进化系统
- 提供WebSocket实时通信

### v1.1.0 (计划中)
- 增强多模态处理能力
- 优化推理性能
- 新增批量处理接口
- 改进错误处理机制