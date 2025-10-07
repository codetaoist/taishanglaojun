# 太上老君监控系统

一个全面的分布式监控系统，提供追踪、日志、告警、性能分析和自动化运维功能。

## 功能特性

### 🔍 分布式追踪 (Distributed Tracing)
- 支持 OpenTelemetry 标准
- 多种采样策略（概率采样、限流采样等）
- 多种导出器（Jaeger、Zipkin、OTLP、Console）
- 自动 span 关联和上下文传播
- 详细的性能统计和健康检查

### 📝 日志聚合 (Log Aggregation)
- 多源日志收集（文件、Syslog、HTTP、Docker、Kubernetes）
- 实时日志处理管道
- 灵活的日志处理器（过滤、解析、富化、转换）
- 多种输出目标（文件、Elasticsearch、Kafka、Redis）
- 高性能批处理和缓冲机制

### 🚨 智能告警 (Intelligent Alerting)
- 基于规则的告警引擎
- 多种告警严重级别
- 丰富的通知渠道（Webhook、邮件、Slack、钉钉）
- 告警聚合和去重
- 告警历史和统计分析

### 📊 可视化仪表板 (Dashboard)
- 实时监控仪表板
- 自定义图表和面板
- 响应式 Web 界面
- 数据缓存和性能优化
- RESTful API 支持

### ⚡ 性能分析 (Performance Analysis)
- 系统资源监控（CPU、内存、磁盘、网络）
- 异常检测和趋势分析
- 性能基线和容量规划
- 智能优化建议
- 健康评分系统

### 🤖 自动化运维 (Automation)
- 工作流编排引擎
- 多种执行器（Shell、Docker、Kubernetes、HTTP）
- 任务调度和依赖管理
- 自动故障恢复
- 运维脚本管理

## 核心功能

### 🔍 指标收集与监控
- **系统指标**: CPU、内存、磁盘、网络使用率
- **应用指标**: 请求量、响应时间、错误率、吞吐量
- **业务指标**: 用户活跃度、功能使用统计、业务流程监控
- **基础设施指标**: 数据库连接池、缓存命中率、消息队列状态

### 📊 实时数据处理
- **流式数据处理**: 基于时间窗口的实时指标计算
- **数据聚合**: 多维度数据聚合和统计分析
- **异常检测**: 基于机器学习的异常模式识别
- **趋势分析**: 历史数据趋势分析和预测

### 🚨 智能告警系统
- **多级告警**: 信息、警告、严重、紧急四级告警
- **告警规则**: 灵活的告警规则配置和管理
- **告警抑制**: 智能告警去重和抑制机制
- **通知渠道**: 邮件、短信、钉钉、企业微信等多渠道通知

### 📈 可视化仪表板
- **实时仪表板**: 实时数据展示和监控大屏
- **自定义图表**: 支持多种图表类型和自定义配置
- **交互式分析**: 支持数据钻取和交互式分析
- **移动端适配**: 响应式设计，支持移动端访问

### 🔗 分布式追踪
- **链路追踪**: 分布式系统调用链路追踪
- **性能分析**: 接口性能分析和瓶颈识别
- **依赖关系**: 服务依赖关系图谱
- **错误定位**: 快速错误定位和根因分析

### 🤖 自动化运维
- **自动扩缩容**: 基于指标的自动扩缩容
- **故障自愈**: 自动故障检测和恢复
- **容量规划**: 基于历史数据的容量规划建议
- **运维建议**: 智能运维建议和优化方案

## 技术架构

### 数据收集层
- **Agent**: 轻量级数据收集代理
- **SDK**: 应用程序集成SDK
- **Exporter**: 第三方系统数据导出器

### 数据处理层
- **时序数据库**: Prometheus + InfluxDB
- **流处理引擎**: Apache Kafka + Apache Flink
- **缓存层**: Redis集群

### 分析计算层
- **指标计算**: 实时指标计算和聚合
- **异常检测**: 基于统计学和机器学习的异常检测
- **预测分析**: 时间序列预测和趋势分析

### 展示服务层
- **API网关**: 统一API接口
- **Web界面**: 基于React的监控界面
- **移动端**: 移动端监控应用

## 目录结构

```
monitoring/
├── README.md                    # 项目说明文档
├── config/                      # 配置文件
│   ├── monitoring.yaml         # 主配置文件
│   ├── alerts.yaml             # 告警规则配置
│   └── dashboards/             # 仪表板配置
├── models/                      # 数据模型
│   ├── metrics.go              # 指标模型
│   ├── alerts.go               # 告警模型
│   ├── dashboards.go           # 仪表板模型
│   └── traces.go               # 追踪模型
├── collectors/                  # 数据收集器
│   ├── system_collector.go     # 系统指标收集器
│   ├── app_collector.go        # 应用指标收集器
│   ├── business_collector.go   # 业务指标收集器
│   └── custom_collector.go     # 自定义指标收集器
├── processors/                  # 数据处理器
│   ├── aggregator.go           # 数据聚合器
│   ├── anomaly_detector.go     # 异常检测器
│   ├── trend_analyzer.go       # 趋势分析器
│   └── predictor.go            # 预测分析器
├── alerting/                    # 告警系统
│   ├── rule_engine.go          # 告警规则引擎
│   ├── notification.go         # 通知服务
│   ├── suppression.go          # 告警抑制
│   └── escalation.go           # 告警升级
├── storage/                     # 数据存储
│   ├── timeseries.go           # 时序数据存储
│   ├── metadata.go             # 元数据存储
│   └── cache.go                # 缓存管理
├── api/                         # API接口
│   ├── metrics_api.go          # 指标API
│   ├── alerts_api.go           # 告警API
│   ├── dashboards_api.go       # 仪表板API
│   └── traces_api.go           # 追踪API
├── handlers/                    # HTTP处理器
│   ├── monitoring_handler.go   # 监控处理器
│   ├── dashboard_handler.go    # 仪表板处理器
│   └── alert_handler.go        # 告警处理器
├── services/                    # 业务服务
│   ├── monitoring_service.go   # 监控服务
│   ├── alert_service.go        # 告警服务
│   └── dashboard_service.go    # 仪表板服务
├── middleware/                  # 中间件
│   ├── metrics_middleware.go   # 指标收集中间件
│   └── tracing_middleware.go   # 追踪中间件
├── routes/                      # 路由配置
│   └── monitoring_routes.go    # 监控路由
├── migrations/                  # 数据库迁移
│   └── 001_create_monitoring_tables.sql
├── tests/                       # 测试文件
│   ├── monitoring_test.go      # 监控测试
│   ├── alert_test.go           # 告警测试
│   └── integration_test.go     # 集成测试
├── scripts/                     # 脚本文件
│   ├── setup.sh               # 安装脚本
│   └── deploy.sh              # 部署脚本
└── module.go                   # 模块入口
```

## 快速开始

### 1. 配置文件
```yaml
# config/monitoring.yaml
monitoring:
  collection:
    interval: 15s
    batch_size: 1000
  storage:
    retention: 30d
    compression: true
  alerting:
    evaluation_interval: 30s
    notification_timeout: 5m
```

### 2. 启动监控系统
```go
// 初始化监控模块
monitoringModule := monitoring.NewModule(db, redis, logger, config)
err := monitoringModule.Initialize()
if err != nil {
    log.Fatal("Failed to initialize monitoring module:", err)
}

// 启动监控服务
go monitoringModule.Start()
```

### 3. 集成应用监控
```go
// 添加监控中间件
router.Use(monitoring.MetricsMiddleware())
router.Use(monitoring.TracingMiddleware())

// 自定义指标
monitoring.RecordCustomMetric("user_login", 1, map[string]string{
    "method": "oauth",
    "provider": "google",
})
```

## 配置说明

### 指标收集配置
- `collection.interval`: 指标收集间隔
- `collection.batch_size`: 批量处理大小
- `collection.timeout`: 收集超时时间

### 存储配置
- `storage.retention`: 数据保留时间
- `storage.compression`: 是否启用压缩
- `storage.sharding`: 分片配置

### 告警配置
- `alerting.rules`: 告警规则文件路径
- `alerting.notification`: 通知配置
- `alerting.suppression`: 抑制规则

## API接口

### 指标查询
```http
GET /api/v1/metrics/query?query=cpu_usage&start=1h&step=1m
```

### 告警管理
```http
POST /api/v1/alerts/rules
GET /api/v1/alerts/active
PUT /api/v1/alerts/silence
```

### 仪表板管理
```http
GET /api/v1/dashboards
POST /api/v1/dashboards
PUT /api/v1/dashboards/{id}
DELETE /api/v1/dashboards/{id}
```

## 部署指南

### Docker部署
```bash
# 构建镜像
docker build -t taishang-monitoring .

# 运行容器
docker run -d \
  --name monitoring \
  -p 8080:8080 \
  -p 9090:9090 \
  -v ./config:/app/config \
  taishang-monitoring
```

### Kubernetes部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: monitoring
spec:
  replicas: 3
  selector:
    matchLabels:
      app: monitoring
  template:
    metadata:
      labels:
        app: monitoring
    spec:
      containers:
      - name: monitoring
        image: taishang-monitoring:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
```

## 监控指标

### 系统指标
- `system_cpu_usage`: CPU使用率
- `system_memory_usage`: 内存使用率
- `system_disk_usage`: 磁盘使用率
- `system_network_io`: 网络IO

### 应用指标
- `http_requests_total`: HTTP请求总数
- `http_request_duration`: 请求响应时间
- `http_request_size`: 请求大小
- `http_response_size`: 响应大小

### 业务指标
- `user_active_count`: 活跃用户数
- `feature_usage_count`: 功能使用次数
- `business_transaction_count`: 业务交易数

## 告警规则示例

```yaml
groups:
- name: system
  rules:
  - alert: HighCPUUsage
    expr: system_cpu_usage > 80
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High CPU usage detected"
      description: "CPU usage is above 80% for more than 5 minutes"

- name: application
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      description: "Error rate is above 10% for more than 2 minutes"
```

## 最佳实践

### 指标设计
1. 使用有意义的指标名称
2. 合理设置标签维度
3. 避免高基数标签
4. 定期清理无用指标

### 告警配置
1. 设置合理的告警阈值
2. 避免告警风暴
3. 配置告警抑制规则
4. 定期审查告警规则

### 性能优化
1. 合理设置采集间隔
2. 使用数据压缩
3. 配置数据分片
4. 定期清理历史数据

## 故障排查

### 常见问题
1. **指标收集失败**: 检查网络连接和权限配置
2. **告警不触发**: 检查告警规则和阈值设置
3. **仪表板加载慢**: 检查查询语句和数据量
4. **存储空间不足**: 调整数据保留策略

### 日志分析
```bash
# 查看监控服务日志
kubectl logs -f deployment/monitoring

# 查看指标收集日志
grep "collector" /var/log/monitoring/app.log

# 查看告警日志
grep "alert" /var/log/monitoring/app.log
```

## 扩展开发

### 自定义收集器
```go
type CustomCollector struct {
    // 实现Collector接口
}

func (c *CustomCollector) Collect() ([]Metric, error) {
    // 收集自定义指标
    return metrics, nil
}
```

### 自定义告警规则
```go
type CustomRule struct {
    // 实现Rule接口
}

func (r *CustomRule) Evaluate(metrics []Metric) ([]Alert, error) {
    // 评估告警条件
    return alerts, nil
}
```

## 版本历史

- v1.0.0: 基础监控功能
- v1.1.0: 添加告警系统
- v1.2.0: 添加仪表板功能
- v1.3.0: 添加分布式追踪
- v1.4.0: 添加自动化运维功能

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交代码变更
4. 创建Pull Request

## 许可证

MIT License