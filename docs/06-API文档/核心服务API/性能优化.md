# 安全服务性能优化文档

## 概述

安全服务集成了全面的性能管理器，提供缓存管理、内存优化、并发控制和性能监控等功能，确保服务在高负载下的稳定运行。

## 性能管理器架构

### 核心组件

1. **缓存管理器 (CacheManager)**
   - 支持本地缓存、Redis缓存和混合缓存策略
   - 分片缓存设计，提高并发性能
   - 自动过期和清理机制

2. **并发管理器 (ConcurrencyManager)**
   - 工作池模式，控制并发数量
   - 任务队列管理
   - 超时和错误处理

3. **内存管理器 (MemoryManager)**
   - 垃圾回收优化
   - 内存使用监控
   - 自动内存清理

4. **性能监控器 (PerformanceMonitor)**
   - 实时性能指标收集
   - 内存、缓存、并发统计
   - 性能趋势分析

## 配置说明

### 基础配置

```yaml
performance:
  # 缓存配置
  cache:
    strategy: "hybrid"  # local, redis, hybrid
    default_ttl: "1h"
    
    redis:
      addr: "localhost:6379"
      password: ""
      db: 1
      pool_size: 100
      min_idle_conns: 10
      max_retries: 3
      dial_timeout: "5s"
      read_timeout: "3s"
      write_timeout: "3s"
      idle_timeout: "5m"
    
    local:
      max_size: 10000
      ttl: "30m"
      cleanup_time: "5m"
      shard_count: 256
  
  # 连接池配置
  connection_pool:
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: "1h"
    conn_max_idle_time: "10m"
  
  # 内存管理配置
  memory:
    gc_percent: 100
    max_memory_mb: 1024
    check_interval: "30s"
    force_gc_threshold: 0.8
  
  # 并发配置
  concurrency:
    max_workers: 100
    queue_size: 1000
    worker_timeout: "30s"
    shutdown_timeout: "10s"
  
  # 监控配置
  monitoring:
    metrics_interval: "10s"
    enable_profiling: true
    profiling_port: 6060
```

### 高级配置

```yaml
# 数据库优化配置
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "10m"
  query_timeout: "30s"
  slow_query_threshold: "1s"
  batch_size: 1000
  batch_timeout: "5s"

# HTTP服务器优化配置
http_server:
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  max_header_bytes: 1048576
  keep_alive: true
  keep_alive_timeout: "60s"

# 资源限制配置
resource_limits:
  cpu:
    max_usage: 80
    check_interval: "10s"
  memory:
    max_usage_mb: 2048
    check_interval: "10s"
  disk:
    max_usage: 90
    check_interval: "30s"
  network:
    max_bandwidth_mb: 100
    max_connections: 10000
```

## API 接口

### 性能监控接口

#### 获取性能指标

```http
GET /api/security/performance/metrics
```

**响应示例：**
```json
{
  "status": "success",
  "data": {
    "memory_usage": 134217728,
    "memory_allocated": 67108864,
    "gc_count": 15,
    "cache_stats": {
      "hits": 1500,
      "misses": 200,
      "sets": 800,
      "deletes": 50,
      "evictions": 10,
      "size": 750,
      "hit_rate": 0.88
    },
    "concurrency_stats": {
      "active_workers": 5,
      "queued_tasks": 12,
      "completed_tasks": 2500,
      "failed_tasks": 8
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

#### 获取缓存统计

```http
GET /api/security/performance/cache/stats
```

**响应示例：**
```json
{
  "status": "success",
  "data": {
    "hits": 1500,
    "misses": 200,
    "sets": 800,
    "deletes": 50,
    "evictions": 10,
    "size": 750,
    "hit_rate": 0.88
  }
}
```

#### 清理缓存

```http
POST /api/security/performance/cache/clear
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Cache cleared successfully"
}
```

#### 获取内存统计

```http
GET /api/security/performance/memory/stats
```

**响应示例：**
```json
{
  "status": "success",
  "data": {
    "alloc": 67108864,
    "total_alloc": 134217728,
    "sys": 201326592,
    "num_gc": 15,
    "gc_cpu_fraction": 0.001,
    "heap_alloc": 67108864,
    "heap_sys": 134217728,
    "heap_idle": 50331648,
    "heap_inuse": 83886080,
    "heap_released": 33554432,
    "heap_objects": 125000,
    "stack_inuse": 1048576,
    "stack_sys": 1048576,
    "next_gc": 134217728,
    "last_gc": 1642248600000000000,
    "pause_total_ns": 5000000,
    "num_forced_gc": 2
  }
}
```

#### 强制垃圾回收

```http
POST /api/security/performance/memory/gc
```

**响应示例：**
```json
{
  "status": "success",
  "message": "Garbage collection forced successfully"
}
```

#### 获取并发统计

```http
GET /api/security/performance/concurrency/stats
```

**响应示例：**
```json
{
  "status": "success",
  "data": {
    "active_workers": 5,
    "queued_tasks": 12,
    "completed_tasks": 2500,
    "failed_tasks": 8
  }
}
```

## 性能优化策略

### 1. 缓存优化

#### 缓存策略选择
- **本地缓存 (local)**: 适用于单机部署，响应速度最快
- **Redis缓存 (redis)**: 适用于分布式部署，支持数据共享
- **混合缓存 (hybrid)**: 本地缓存 + Redis缓存，兼顾性能和一致性

#### 缓存分片
- 使用FNV-1a哈希算法进行分片
- 默认256个分片，减少锁竞争
- 支持自定义分片数量

#### 缓存清理
- 自动过期机制
- 定期清理过期项
- LRU淘汰策略

### 2. 并发优化

#### 工作池模式
- 固定数量的工作协程
- 任务队列缓冲
- 超时控制和错误处理

#### 任务调度
- 优先级队列支持
- 批处理优化
- 背压控制

### 3. 内存优化

#### 垃圾回收优化
- 自定义GC百分比
- 内存阈值监控
- 强制GC触发

#### 内存池
- 对象复用
- 减少内存分配
- 内存泄漏检测

### 4. 数据库优化

#### 连接池管理
- 连接数控制
- 连接生命周期管理
- 空闲连接清理

#### 查询优化
- 慢查询监控
- 批量操作
- 索引优化建议

## 监控和告警

### 性能指标

1. **内存指标**
   - 内存使用量
   - 堆内存统计
   - GC频率和耗时

2. **缓存指标**
   - 命中率
   - 缓存大小
   - 操作延迟

3. **并发指标**
   - 活跃工作协程数
   - 队列长度
   - 任务完成率

4. **数据库指标**
   - 连接池状态
   - 查询延迟
   - 慢查询统计

### 告警规则

```yaml
alerts:
  - name: "high_memory_usage"
    condition: "memory_usage_percent > 80"
    duration: "5m"
    severity: "warning"
  
  - name: "low_cache_hit_rate"
    condition: "cache_hit_rate < 0.7"
    duration: "10m"
    severity: "warning"
  
  - name: "high_queue_length"
    condition: "queued_tasks > 500"
    duration: "2m"
    severity: "critical"
  
  - name: "database_slow_queries"
    condition: "slow_query_count > 10"
    duration: "5m"
    severity: "warning"
```

## 性能调优建议

### 1. 缓存调优

- 根据业务特点选择合适的缓存策略
- 调整缓存大小和TTL
- 监控缓存命中率，优化缓存键设计
- 使用缓存预热提高初始性能

### 2. 并发调优

- 根据CPU核心数调整工作协程数量
- 监控队列长度，避免任务积压
- 优化任务粒度，避免长时间阻塞
- 使用异步处理提高吞吐量

### 3. 内存调优

- 监控内存使用趋势
- 调整GC参数优化延迟
- 使用内存池减少分配开销
- 定期检查内存泄漏

### 4. 数据库调优

- 优化连接池配置
- 使用批量操作减少网络开销
- 添加适当的索引
- 监控慢查询并优化

## 故障排查

### 常见问题

1. **内存使用过高**
   - 检查内存泄漏
   - 调整GC参数
   - 优化数据结构

2. **缓存命中率低**
   - 分析缓存键分布
   - 调整缓存大小
   - 优化缓存策略

3. **任务队列积压**
   - 增加工作协程数量
   - 优化任务处理逻辑
   - 检查资源瓶颈

4. **数据库连接超时**
   - 调整连接池配置
   - 优化查询性能
   - 检查网络延迟

### 诊断工具

1. **性能分析**
   ```bash
   # 启用pprof
   go tool pprof http://localhost:6060/debug/pprof/profile
   
   # 内存分析
   go tool pprof http://localhost:6060/debug/pprof/heap
   
   # 协程分析
   go tool pprof http://localhost:6060/debug/pprof/goroutine
   ```

2. **指标监控**
   ```bash
   # 获取性能指标
   curl http://localhost:8080/api/security/performance/metrics
   
   # 获取内存统计
   curl http://localhost:8080/api/security/performance/memory/stats
   ```

## 最佳实践

1. **定期监控性能指标**
2. **设置合理的告警阈值**
3. **进行性能基准测试**
4. **优化热点代码路径**
5. **使用异步处理提高并发**
6. **合理配置资源限制**
7. **定期进行性能调优**
8. **建立性能回归测试**

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 实现基础性能管理功能
- 支持缓存、并发、内存管理
- 提供性能监控API

### 计划功能
- 分布式缓存一致性
- 自适应性能调优
- 机器学习性能预测
- 更丰富的监控指标