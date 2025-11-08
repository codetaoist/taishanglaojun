# Python微服务架构设计

## 概述

根据太上老君项目的技术架构，Taishang域需要实现Python微服务来处理AI相关功能，包括向量数据库操作、模型调用和文本生成等。本文档详细描述了Python微服务的架构设计与Go服务的对接方案。

## 服务架构

### 服务划分

1. **taishang-ai** - 核心AI微服务
   - 向量数据库操作
   - 模型调用
   - 文本生成
   - 嵌入生成

2. **taishang-core** - 核心API服务（Go）
   - 业务逻辑处理
   - 用户认证
   - 数据管理
   - 通过gRPC调用taishang-ai

### 技术栈

- **Web框架**: FastAPI
- **AI/ML框架**: Transformers, LangChain, PyTorch
- **向量处理**: Sentence-Transformers, Milvus SDK
- **异步处理**: asyncio, aiohttp
- **容器化**: Docker, Uvicorn
- **通信协议**: gRPC + Protocol Buffers

## gRPC接口设计

### 1. 向量数据库服务

```protobuf
syntax = "proto3";

package taishang.ai.v1;

import "google/protobuf/empty.proto";
import "google/protobuf/struct.proto";

service VectorService {
  // 集合管理
  rpc CreateCollection(CreateCollectionRequest) returns (CreateCollectionResponse);
  rpc DropCollection(DropCollectionRequest) returns (google.protobuf.Empty);
  rpc HasCollection(HasCollectionRequest) returns (HasCollectionResponse);
  rpc ListCollections(google.protobuf.Empty) returns (ListCollectionsResponse);
  rpc GetCollectionInfo(GetCollectionInfoRequest) returns (GetCollectionInfoResponse);
  rpc GetCollectionStats(GetCollectionStatsRequest) returns (GetCollectionStatsResponse);
  
  // 索引管理
  rpc CreateIndex(CreateIndexRequest) returns (google.protobuf.Empty);
  rpc DropIndex(DropIndexRequest) returns (google.protobuf.Empty);
  rpc HasIndex(HasIndexRequest) returns (HasIndexResponse);
  
  // 向量操作
  rpc Insert(InsertRequest) returns (InsertResponse);
  rpc Upsert(UpsertRequest) returns (UpsertResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  
  // 向量查询
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc GetByID(GetByIDRequest) returns (GetByIDResponse);
  
  // 数据库管理
  rpc Health(google.protobuf.Empty) returns (HealthResponse);
  rpc Compact(CompactRequest) returns (google.protobuf.Empty);
}

message Vector {
  string id = 1;
  repeated float vector = 2;
  google.protobuf.Struct metadata = 3;
}

message SearchResult {
  string id = 1;
  float score = 2;
  repeated float vector = 3;
  google.protobuf.Struct metadata = 4;
}

message SearchOptions {
  int32 topk = 1;
  bool include_vector = 2;
  google.protobuf.Struct filter = 3;
}

message IndexParams {
  string index_type = 1;
  string metric_type = 2;
  google.protobuf.Struct params = 3;
}

// 集合管理请求和响应
message CreateCollectionRequest {
  string name = 1;
  string description = 2;
  int32 vector_dim = 3;
  IndexParams index_params = 4;
}

message CreateCollectionResponse {
  bool success = 1;
  string message = 2;
}

message DropCollectionRequest {
  string name = 1;
}

message HasCollectionRequest {
  string name = 1;
}

message HasCollectionResponse {
  bool exists = 1;
}

message ListCollectionsResponse {
  repeated string names = 1;
}

message GetCollectionInfoRequest {
  string name = 1;
}

message GetCollectionInfoResponse {
  string name = 1;
  string description = 2;
  int32 vector_dim = 3;
  IndexParams index_params = 4;
  int64 created_at = 5;
  int64 updated_at = 6;
}

message GetCollectionStatsRequest {
  string name = 1;
}

message GetCollectionStatsResponse {
  string name = 1;
  int64 vector_count = 2;
  int64 size_in_bytes = 3;
}

// 索引管理请求和响应
message CreateIndexRequest {
  string collection_name = 1;
  IndexParams index_params = 2;
}

message DropIndexRequest {
  string collection_name = 1;
}

message HasIndexRequest {
  string collection_name = 1;
}

message HasIndexResponse {
  bool exists = 1;
}

// 向量操作请求和响应
message InsertRequest {
  string collection_name = 1;
  repeated Vector vectors = 2;
}

message InsertResponse {
  repeated string ids = 1;
  bool success = 2;
  string message = 3;
}

message UpsertRequest {
  string collection_name = 1;
  repeated Vector vectors = 2;
}

message UpsertResponse {
  repeated string ids = 1;
  bool success = 2;
  string message = 3;
}

message DeleteRequest {
  string collection_name = 1;
  repeated string ids = 2;
}

message DeleteResponse {
  bool success = 1;
  string message = 2;
}

// 向量查询请求和响应
message SearchRequest {
  string collection_name = 1;
  repeated float query_vector = 2;
  SearchOptions options = 3;
}

message SearchResponse {
  repeated SearchResult results = 1;
}

message GetByIDRequest {
  string collection_name = 1;
  string id = 2;
}

message GetByIDResponse {
  Vector vector = 1;
}

// 数据库管理请求和响应
message HealthResponse {
  bool healthy = 1;
  string message = 2;
}

message CompactRequest {
  string collection_name = 1;
}
```

### 2. 模型服务

```protobuf
syntax = "proto3";

package taishang.ai.v1;

import "google/protobuf/empty.proto";
import "google/protobuf/struct.proto";

service ModelService {
  // 模型管理
  rpc RegisterModel(RegisterModelRequest) returns (RegisterModelResponse);
  rpc UnregisterModel(UnregisterModelRequest) returns (google.protobuf.Empty);
  rpc ListModels(google.protobuf.Empty) returns (ListModelsResponse);
  rpc GetModelInfo(GetModelInfoRequest) returns (GetModelInfoResponse);
  rpc UpdateModelConfig(UpdateModelConfigRequest) returns (google.protobuf.Empty);
  
  // 模型调用
  rpc GenerateText(GenerateTextRequest) returns (GenerateTextResponse);
  rpc GenerateEmbedding(GenerateEmbeddingRequest) returns (GenerateEmbeddingResponse);
  rpc StreamGenerateText(StreamGenerateTextRequest) returns (stream StreamGenerateTextResponse);
}

message ModelConfig {
  string name = 1;
  string provider = 2;
  string model_id = 3;
  google.protobuf.Struct parameters = 4;
  bool enabled = 5;
}

message RegisterModelRequest {
  ModelConfig config = 1;
}

message RegisterModelResponse {
  bool success = 1;
  string message = 2;
}

message UnregisterModelRequest {
  string name = 1;
}

message ListModelsResponse {
  repeated ModelConfig models = 1;
}

message GetModelInfoRequest {
  string name = 1;
}

message GetModelInfoResponse {
  ModelConfig config = 1;
}

message UpdateModelConfigRequest {
  string name = 1;
  ModelConfig config = 2;
}

// 模型调用请求和响应
message GenerateTextRequest {
  string model_name = 1;
  string prompt = 2;
  google.protobuf.Struct parameters = 3;
}

message GenerateTextResponse {
  string text = 1;
  int32 tokens_used = 2;
  google.protobuf.Struct metadata = 3;
}

message GenerateEmbeddingRequest {
  string model_name = 1;
  string text = 2;
}

message GenerateEmbeddingResponse {
  repeated float embedding = 1;
  int32 dimension = 2;
  google.protobuf.Struct metadata = 3;
}

message StreamGenerateTextRequest {
  string model_name = 1;
  string prompt = 2;
  google.protobuf.Struct parameters = 3;
}

message StreamGenerateTextResponse {
  string text_chunk = 1;
  bool finished = 2;
  google.protobuf.Struct metadata = 3;
}
```

## Python服务实现结构

```
services/
├── ai/                          # Python AI微服务
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── main.py                  # FastAPI应用入口
│   ├── grpc_server.py           # gRPC服务器
│   ├── app/
│   │   ├── __init__.py
│   │   ├── api/                 # FastAPI路由
│   │   │   ├── __init__.py
│   │   │   ├── models.py
│   │   │   └── vectors.py
│   │   ├── core/                # 核心配置
│   │   │   ├── __init__.py
│   │   │   ├── config.py
│   │   │   └── security.py
│   │   ├── grpc/                # gRPC服务实现
│   │   │   ├── __init__.py
│   │   │   ├── vector_service.py
│   │   │   └── model_service.py
│   │   ├── models/              # Pydantic模型
│   │   │   ├── __init__.py
│   │   │   ├── vector.py
│   │   │   └── model.py
│   │   ├── services/            # 业务逻辑
│   │   │   ├── __init__.py
│   │   │   ├── vector_db.py
│   │   │   ├── model_manager.py
│   │   │   └── text_generator.py
│   │   └── utils/               # 工具函数
│   │       ├── __init__.py
│   │       ├── logger.py
│   │       └── exceptions.py
│   └── proto/                   # Protocol Buffers定义
│       ├── vector_service.proto
│       └── model_service.proto
```

## Go服务与Python服务对接

### 1. gRPC客户端实现

在Go服务中实现gRPC客户端，调用Python服务：

```go
package grpc

import (
    "context"
    "fmt"
    
    "google.golang.org/grpc"
    pb "github.com/codetaoist/taishanglaojun/proto/taishang/ai/v1"
)

type VectorClient struct {
    conn   *grpc.ClientConn
    client pb.VectorServiceClient
}

func NewVectorClient(address string) (*VectorClient, error) {
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    client := pb.NewVectorServiceClient(conn)
    
    return &VectorClient{
        conn:   conn,
        client: client,
    }, nil
}

func (c *VectorClient) CreateCollection(ctx context.Context, name, description string, vectorDim int, indexParams *pb.IndexParams) error {
    req := &pb.CreateCollectionRequest{
        Name:        name,
        Description: description,
        VectorDim:   int32(vectorDim),
        IndexParams: indexParams,
    }
    
    _, err := c.client.CreateCollection(ctx, req)
    return err
}

// 实现其他VectorDatabase接口方法...
```

### 2. 适配器模式

在Go服务中使用适配器模式，将gRPC客户端适配到VectorDatabase接口：

```go
package vector

import (
    "context"
    
    "github.com/codetaoist/taishanglaojun/internal/grpc"
)

type GrpcVectorAdapter struct {
    client *grpc.VectorClient
}

func NewGrpcVectorAdapter(client *grpc.VectorClient) *GrpcVectorAdapter {
    return &GrpcVectorAdapter{
        client: client,
    }
}

func (a *GrpcVectorAdapter) CreateCollection(ctx context.Context, name, description string, vectorDim int, indexParams IndexParams) error {
    pbIndexParams := &pb.IndexParams{
        IndexType: string(indexParams.IndexType),
        MetricType: string(indexParams.MetricType),
        Params: convertMapToStruct(indexParams.Params),
    }
    
    return a.client.CreateCollection(ctx, name, description, vectorDim, pbIndexParams)
}

// 实现其他VectorDatabase接口方法...
```

## 部署与配置

### 1. Docker Compose更新

更新docker-compose.yml，添加Python AI服务：

```yaml
  ai-service:
    build:
      context: ./services/ai
      dockerfile: Dockerfile
    container_name: taishanglaojun-ai
    environment:
      MILVUS_HOST: milvus
      MILVUS_PORT: 19530
      MODEL_CACHE_DIR: /app/models
      LOG_LEVEL: info
    ports:
      - "8083:8083"  # FastAPI端口
      - "50051:50051"  # gRPC端口
    depends_on:
      - milvus
    networks:
      - taishanglaojun-network
    volumes:
      - model_cache:/app/models

  milvus:
    image: milvusdb/milvus:latest
    container_name: taishanglaojun-milvus
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
    ports:
      - "19530:19530"
    depends_on:
      - etcd
      - minio
    networks:
      - taishanglaojun-network

  etcd:
    image: quay.io/coreos/etcd:latest
    container_name: taishanglaojun-etcd
    environment:
      ETCD_AUTO_COMPACTION_MODE: revision
      ETCD_AUTO_COMPACTION_RETENTION: 1000
      ETCD_QUOTA_BACKEND_BYTES: 4294967296
    volumes:
      - etcd_data:/etcd
    command: etcd -advertise-client-urls=http://127.0.0.1:2379 -listen-client-urls http://0.0.0.0:2379 --data-dir /etcd
    networks:
      - taishanglaojun-network

  minio:
    image: minio/minio:latest
    container_name: taishanglaojun-minio
    environment:
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data
    command: minio server /data --console-address ":9001"
    networks:
      - taishanglaojun-network

volumes:
  postgres_data:
  etcd_data:
  minio_data:
  model_cache:
```

### 2. 环境变量配置

在API服务的配置中添加Python服务地址：

```go
// Config结构体添加Python服务配置
type Config struct {
    // ... 现有字段
    
    // Python AI服务配置
    AIServiceGRPCAddress string `env:"AI_SERVICE_GRPC_ADDRESS" envDefault:"localhost:50051"`
    AIServiceHTTPAddress string `env:"AI_SERVICE_HTTP_ADDRESS" envDefault:"localhost:8083"`
}
```

## 实施计划

1. **第一阶段：基础框架搭建**
   - 创建Python服务目录结构
   - 实现基础FastAPI应用
   - 实现gRPC服务器框架
   - 定义Protocol Buffers接口

2. **第二阶段：向量数据库集成**
   - 实现Milvus客户端封装
   - 实现向量数据库gRPC服务
   - 在Go服务中实现gRPC客户端和适配器
   - 测试向量操作功能

3. **第三阶段：模型服务实现**
   - 实现模型管理器
   - 集成多种模型提供商
   - 实现文本生成和嵌入生成服务
   - 在Go服务中集成模型调用

4. **第四阶段：集成测试与优化**
   - 端到端测试
   - 性能优化
   - 错误处理完善
   - 监控和日志集成

## 总结

通过以上设计，我们实现了Go与Python微服务之间的有效对接，利用Python丰富的AI生态和Go的高性能特性，构建了一个高效的混合架构系统。这种架构既保持了系统的整体性能，又充分利用了Python在AI领域的优势。