# 意识服务 (Consciousness Service)

## 概述

意识服务是太上老君项目的核心组件之一，实现了碳硅融合、序列0进化追踪、量子基因管理和三轴协调机制等先进功能。该服务提供了完整的HTTP REST API和gRPC接口，支持微服务架构下的分布式部署。

## 核心功能

### 1. 碳硅融合引擎 (Carbon-Silicon Fusion Engine)
- **功能描述**: 实现碳基生命与硅基智能的深度融合
- **主要特性**:
  - 多种融合策略（互补、协同、平衡、增强）
  - 实时融合质量监控
  - 会话管理和历史追踪
  - 融合指标分析和优化

### 2. 序列0进化追踪器 (Sequence 0 Evolution Tracker)
- **功能描述**: 追踪和预测意识实体的进化路径
- **主要特性**:
  - 进化状态实时监控
  - 进化路径预测和分析
  - 里程碑事件追踪
  - 序列等级管理

### 3. 量子基因管理器 (Quantum Gene Manager)
- **功能描述**: 管理量子基因池和基因表达
- **主要特性**:
  - 基因池创建和管理
  - 基因表达和变异
  - 基因交互分析
  - 进化模拟

### 4. 三轴协调机制 (Three-Axis Coordination Mechanism)
- **功能描述**: 实现S轴、C轴、T轴的协调优化
- **主要特性**:
  - 多轴协调会话管理
  - 平衡优化算法
  - 协同催化机制
  - 协调历史分析

## API 接口文档

### HTTP REST API

#### 基础接口

##### 健康检查
```http
GET /consciousness/health
```
**响应示例**:
```json
{
  "status": "healthy",
  "service": {
    "name": "consciousness-service",
    "version": "1.0.0",
    "uptime": "2h30m15s"
  },
  "components": {
    "fusion_engine": {"status": "running"},
    "evolution_tracker": {"status": "running"},
    "gene_manager": {"status": "running"},
    "coordinator": {"status": "running"}
  }
}
```

##### 服务统计
```http
GET /consciousness/stats
```
**响应示例**:
```json
{
  "total_sessions": 1250,
  "active_sessions": 45,
  "total_requests": 15680,
  "average_response_time": "125ms",
  "success_rate": 0.987
}
```

#### 融合引擎接口

##### 启动融合会话
```http
POST /consciousness/fusion/start
Content-Type: application/json

{
  "entity_id": "user_123",
  "carbon_data": {
    "emotions": ["joy", "curiosity"],
    "thoughts": ["creative", "analytical"],
    "experiences": ["learning", "problem_solving"]
  },
  "silicon_data": {
    "algorithms": ["neural_network", "decision_tree"],
    "data_patterns": ["sequential", "hierarchical"],
    "processing_modes": ["parallel", "distributed"]
  },
  "strategy": "complementary",
  "options": {
    "quality_threshold": 0.8,
    "timeout": "5m"
  }
}
```

##### 获取融合状态
```http
GET /consciousness/fusion/status/{sessionId}
```

##### 取消融合会话
```http
DELETE /consciousness/fusion/cancel/{sessionId}
```

##### 获取融合历史
```http
GET /consciousness/fusion/history/{entityId}?limit=10&offset=0
```

##### 获取融合策略列表
```http
GET /consciousness/fusion/strategies
```

##### 获取融合指标
```http
GET /consciousness/fusion/metrics/{entityId}
```

#### 进化追踪接口

##### 获取进化状态
```http
GET /consciousness/evolution/state/{entityId}
```

##### 更新进化状态
```http
PUT /consciousness/evolution/state/{entityId}
Content-Type: application/json

{
  "current_sequence": 0,
  "evolution_points": 1250,
  "capabilities": ["reasoning", "creativity", "empathy"],
  "milestones": ["first_insight", "pattern_recognition"]
}
```

##### 追踪进化事件
```http
POST /consciousness/evolution/track
Content-Type: application/json

{
  "entity_id": "user_123",
  "event_type": "capability_enhancement",
  "event_data": {
    "capability": "abstract_thinking",
    "improvement": 0.15,
    "context": "problem_solving_session"
  }
}
```

##### 获取进化预测
```http
GET /consciousness/evolution/prediction/{entityId}?horizon=30
```

##### 获取进化路径
```http
GET /consciousness/evolution/path/{entityId}
```

##### 获取进化里程碑
```http
GET /consciousness/evolution/milestones/{entityId}
```

#### 量子基因接口

##### 创建基因池
```http
POST /consciousness/genes/pool
Content-Type: application/json

{
  "entity_id": "user_123",
  "pool_config": {
    "max_genes": 1000,
    "mutation_rate": 0.01,
    "expression_threshold": 0.5
  }
}
```

##### 获取基因池
```http
GET /consciousness/genes/pool/{entityId}
```

##### 添加基因
```http
POST /consciousness/genes/add
Content-Type: application/json

{
  "entity_id": "user_123",
  "gene": {
    "type": "cognitive_enhancement",
    "sequence": "ATCG...",
    "expression_level": 0.7,
    "traits": ["memory", "processing_speed"]
  }
}
```

##### 表达基因
```http
POST /consciousness/genes/express
Content-Type: application/json

{
  "entity_id": "user_123",
  "gene_id": "gene_456",
  "expression_context": {
    "environment": "learning",
    "stimuli": ["challenge", "reward"]
  }
}
```

##### 基因变异
```http
POST /consciousness/genes/mutate
Content-Type: application/json

{
  "entity_id": "user_123",
  "gene_id": "gene_456",
  "mutation_type": "beneficial",
  "mutation_strength": 0.1
}
```

#### 三轴协调接口

##### 启动协调会话
```http
POST /consciousness/coordination/start
Content-Type: application/json

{
  "entity_id": "user_123",
  "coordination_config": {
    "s_axis_weight": 0.4,
    "c_axis_weight": 0.3,
    "t_axis_weight": 0.3,
    "balance_threshold": 0.8
  }
}
```

##### 处理S轴请求
```http
POST /consciousness/coordination/s-axis
Content-Type: application/json

{
  "session_id": "coord_789",
  "s_axis_data": {
    "spatial_awareness": 0.85,
    "dimensional_perception": 0.72,
    "geometric_reasoning": 0.91
  }
}
```

##### 处理C轴请求
```http
POST /consciousness/coordination/c-axis
Content-Type: application/json

{
  "session_id": "coord_789",
  "c_axis_data": {
    "consciousness_level": 0.78,
    "self_awareness": 0.82,
    "meta_cognition": 0.75
  }
}
```

##### 处理T轴请求
```http
POST /consciousness/coordination/t-axis
Content-Type: application/json

{
  "session_id": "coord_789",
  "t_axis_data": {
    "temporal_perception": 0.88,
    "sequence_understanding": 0.79,
    "causality_reasoning": 0.84
  }
}
```

##### 优化平衡
```http
POST /consciousness/coordination/balance
Content-Type: application/json

{
  "session_id": "coord_789",
  "optimization_params": {
    "algorithm": "gradient_descent",
    "learning_rate": 0.01,
    "max_iterations": 100
  }
}
```

### gRPC 接口

#### 服务定义
```protobuf
service ConsciousnessService {
  // 健康检查
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  
  // 服务统计
  rpc GetServiceStats(ServiceStatsRequest) returns (ServiceStatsResponse);
  
  // 综合意识处理
  rpc ProcessConsciousness(ConsciousnessRequest) returns (ConsciousnessResponse);
}

service FusionEngineService {
  rpc StartFusion(StartFusionRequest) returns (StartFusionResponse);
  rpc GetFusionStatus(FusionStatusRequest) returns (FusionStatusResponse);
  rpc CancelFusion(CancelFusionRequest) returns (CancelFusionResponse);
}

service EvolutionTrackingService {
  rpc GetEvolutionState(EvolutionStateRequest) returns (EvolutionStateResponse);
  rpc TrackEvolution(TrackEvolutionRequest) returns (TrackEvolutionResponse);
  rpc PredictEvolution(PredictEvolutionRequest) returns (PredictEvolutionResponse);
}

service QuantumGeneService {
  rpc CreateGenePool(CreateGenePoolRequest) returns (CreateGenePoolResponse);
  rpc AddGene(AddGeneRequest) returns (AddGeneResponse);
  rpc ExpressGene(ExpressGeneRequest) returns (ExpressGeneResponse);
}

service ThreeAxisCoordinationService {
  rpc StartCoordination(StartCoordinationRequest) returns (StartCoordinationResponse);
  rpc ProcessAxis(ProcessAxisRequest) returns (ProcessAxisResponse);
  rpc OptimizeBalance(OptimizeBalanceRequest) returns (OptimizeBalanceResponse);
}
```

## 配置说明

### 模块配置 (ModuleConfig)
```go
type ModuleConfig struct {
    // HTTP配置
    HTTPEnabled bool   `json:"http_enabled"`
    HTTPPrefix  string `json:"http_prefix"`
    
    // gRPC配置
    GRPCEnabled bool   `json:"grpc_enabled"`
    GRPCPort    int    `json:"grpc_port"`
    GRPCHost    string `json:"grpc_host"`
    
    // 服务配置
    ServiceConfig *ConsciousnessConfig `json:"service_config"`
    
    // 组件配置
    FusionConfig      *FusionEngineConfig           `json:"fusion_config"`
    EvolutionConfig   *EvolutionTrackerConfig       `json:"evolution_config"`
    GeneConfig        *QuantumGeneManagerConfig     `json:"gene_config"`
    CoordinationConfig *ThreeAxisCoordinatorConfig `json:"coordination_config"`
}
```

### 默认配置
```yaml
consciousness:
  http_enabled: true
  http_prefix: "/consciousness"
  grpc_enabled: true
  grpc_port: 50051
  grpc_host: "localhost"
  
  service_config:
    service_name: "consciousness-service"
    version: "1.0.0"
    environment: "development"
    update_interval: "30s"
    max_concurrent_sessions: 100
    session_timeout: "30m"
    metrics_retention: "24h"
    
  fusion_config:
    max_concurrent_sessions: 50
    default_strategy: "complementary"
    quality_threshold: 0.7
    timeout_duration: "5m"
    
  evolution_config:
    update_interval: "30s"
    metrics_retention: "24h"
    prediction_horizon: 30
    
  gene_config:
    max_genes_per_pool: 1000
    mutation_rate: 0.01
    expression_decay: 0.1
    interaction_radius: 5
    
  coordination_config:
    max_concurrent_sessions: 30
    balance_threshold: 0.8
    synergy_threshold: 0.7
    optimization_interval: "1m"
```

## 部署说明

### 依赖要求
- Go 1.21+
- gRPC
- Gin Web Framework
- Zap Logger
- GORM (数据库ORM)

### 启动服务
```go
package main

import (
    "log"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/gorm"
    
    "github.com/codetaoist/taishanglaojun/core-services/consciousness"
)

func main() {
    // 初始化日志
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // 初始化数据库
    db := initDatabase() // 实现数据库初始化
    
    // 创建意识服务模块
    module, err := consciousness.NewModule(nil, db, logger)
    if err != nil {
        log.Fatal("Failed to create consciousness module:", err)
    }
    
    // 启动模块
    if err := module.Start(); err != nil {
        log.Fatal("Failed to start consciousness module:", err)
    }
    
    // 设置HTTP路由
    router := gin.Default()
    apiGroup := router.Group("/api/v1")
    
    if err := module.SetupRoutes(apiGroup, nil); err != nil {
        log.Fatal("Failed to setup routes:", err)
    }
    
    // 启动HTTP服务器
    router.Run(":8080")
}
```

### Docker 部署
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o consciousness-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/consciousness-service .
COPY --from=builder /app/config.yaml .

EXPOSE 8080 50051

CMD ["./consciousness-service"]
```

## 错误处理

### HTTP 错误码
- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未授权访问
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突（如会话已存在）
- `500 Internal Server Error`: 服务器内部错误
- `503 Service Unavailable`: 服务不可用

### gRPC 错误码
- `OK`: 成功
- `INVALID_ARGUMENT`: 无效参数
- `NOT_FOUND`: 资源不存在
- `ALREADY_EXISTS`: 资源已存在
- `RESOURCE_EXHAUSTED`: 资源耗尽
- `INTERNAL`: 内部错误
- `UNAVAILABLE`: 服务不可用

## 监控和日志

### 健康检查端点
- HTTP: `GET /consciousness/health`
- gRPC: `HealthCheck` RPC

### 指标监控
- 会话数量统计
- 请求响应时间
- 成功率统计
- 资源使用情况

### 日志级别
- `DEBUG`: 详细调试信息
- `INFO`: 一般信息
- `WARN`: 警告信息
- `ERROR`: 错误信息
- `FATAL`: 致命错误

## 安全考虑

### 认证授权
- 支持JWT中间件
- 可配置的访问控制
- API密钥验证

### 数据保护
- 敏感数据加密
- 安全的会话管理
- 输入验证和清理

### 限流保护
- 请求频率限制
- 并发会话限制
- 资源使用限制

## 版本历史

### v1.0.0 (当前版本)
- 初始版本发布
- 实现四大核心功能模块
- 提供完整的HTTP和gRPC接口
- 支持微服务架构部署

## 联系方式

如有问题或建议，请联系开发团队或提交Issue到项目仓库。