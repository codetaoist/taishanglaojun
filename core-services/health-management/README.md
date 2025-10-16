# 健康管理服务 (Health Management Service)

太上老君AI平台的健康管理核心服务，提供全面的健康数据管理、分析和监控功能。

## 🎯 功能特性

### 核心功能
- **健康数据管理**: 支持多种健康指标的记录、查询和分析
- **健康档案管理**: 完整的用户健康档案信息管理
- **智能健康分析**: 基于AI的健康数据分析和趋势预测
- **异常检测**: 实时健康数据异常监测和预警
- **健康报告**: 自动生成个性化健康报告
- **多设备支持**: 支持智能手表、医疗设备等多种数据源

### 技术特性
- **微服务架构**: 基于DDD设计的清洁架构
- **高性能**: 支持大规模健康数据处理
- **实时监控**: 完整的监控和告警体系
- **数据安全**: 符合医疗数据安全标准
- **可扩展性**: 支持水平扩展和负载均衡

## 🏗️ 架构设计

### 领域驱动设计 (DDD)
```
health-management/
├── cmd/                    # 应用入口
│   └── server/
├── internal/              # 内部模块
│   ├── domain/           # 领域层
│   │   ├── health_data.go      # 健康数据聚合根
│   │   ├── health_profile.go   # 健康档案聚合根
│   │   ├── events.go           # 领域事件
│   │   └── repository.go       # 仓储接口
│   ├── application/      # 应用层
│   │   ├── health_data_service.go
│   │   └── health_profile_service.go
│   ├── infrastructure/   # 基础设施层
│   │   └── repository/
│   └── interfaces/       # 接口层
│       └── http/
├── configs/              # 配置文件
├── scripts/              # 脚本文件
└── docs/                 # 文档
```

### 数据模型

#### 健康数据类型
- `heart_rate`: 心率 (bpm)
- `blood_pressure`: 血压 (mmHg)
- `blood_sugar`: 血糖 (mmol/L)
- `body_temperature`: 体温 (°C)
- `weight`: 体重 (kg)
- `height`: 身高 (cm)
- `bmi`: 身体质量指数
- `steps`: 步数
- `sleep_duration`: 睡眠时长 (hours)
- `stress_level`: 压力水平 (1-10)
- `oxygen_saturation`: 血氧饱和度 (%)
- `respiratory_rate`: 呼吸频率 (次/分钟)

#### 数据来源
- `manual_input`: 手动输入
- `smart_watch`: 智能手表
- `fitness_tracker`: 健身追踪器
- `medical_device`: 医疗设备
- `mobile_app`: 移动应用
- `iot_sensor`: IoT传感器
- `hospital_system`: 医院系统

## 🚀 快速开始

### 环境要求
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd health-management
```

2. **安装依赖**
```bash
go mod download
```

3. **启动依赖服务**
```bash
docker-compose up -d postgres redis
```

4. **运行数据库迁移**
```bash
psql -h localhost -U postgres -d health_management -f scripts/init.sql
```

5. **启动服务**
```bash
go run cmd/server/main.go
```

### Docker 部署

1. **构建并启动所有服务**
```bash
docker-compose up -d
```

2. **查看服务状态**
```bash
docker-compose ps
```

3. **查看日志**
```bash
docker-compose logs -f health-management
```

## 📚 API 文档

### 健康数据 API

#### 创建健康数据
```http
POST /api/v1/health-data
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "data_type": "heart_rate",
  "value": 72.5,
  "unit": "bpm",
  "source": "smart_watch",
  "device_id": "apple_watch_001",
  "recorded_at": "2024-01-15T10:30:00Z"
}
```

#### 获取用户健康数据
```http
GET /api/v1/health-data/user/{user_id}?data_type=heart_rate&start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T23:59:59Z
```

#### 获取健康数据统计
```http
GET /api/v1/health-data/user/{user_id}/statistics?data_type=heart_rate&period=week
```

### 健康档案 API

#### 创建健康档案
```http
POST /api/v1/health-profiles
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "gender": "male",
  "date_of_birth": "1990-01-15",
  "height": 175.5,
  "blood_type": "A+",
  "emergency_contact": {
    "name": "张三",
    "phone": "13800138000",
    "relationship": "配偶"
  }
}
```

#### 更新健康档案
```http
PUT /api/v1/health-profiles/{id}
Content-Type: application/json

{
  "height": 176.0,
  "medical_history": ["高血压", "糖尿病"],
  "allergies": ["青霉素"],
  "medications": ["降压药"]
}
```

## 🔧 配置说明

### 环境变量
```bash
# 服务配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=health_management

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# JWT配置
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=24h
```

### 配置文件
详细配置请参考 `configs/config.yaml`

## 📊 监控和运维

### 健康检查
```bash
curl http://localhost:8080/health
```

### Prometheus 指标
```bash
curl http://localhost:8080/metrics
```

### 日志查看
```bash
# Docker 环境
docker-compose logs -f health-management

# 本地环境
tail -f /var/log/health-management/app.log
```

## 🧪 测试

### 运行单元测试
```bash
go test ./...
```

### 运行集成测试
```bash
go test -tags=integration ./...
```

### 性能测试
```bash
go test -bench=. ./...
```

## 🔒 安全考虑

### 数据加密
- 敏感健康数据采用AES-256加密存储
- 传输过程使用TLS 1.3加密
- 数据库连接使用SSL

### 访问控制
- 基于JWT的身份认证
- 细粒度的权限控制
- API限流和防护

### 合规性
- 符合GDPR数据保护要求
- 遵循医疗数据安全标准
- 支持数据删除和导出

## 🚀 部署指南

### Kubernetes 部署
```yaml
# 参考 k8s/ 目录下的配置文件
kubectl apply -f k8s/
```

### 生产环境配置
- 启用TLS证书
- 配置数据库主从复制
- 设置Redis集群
- 配置监控告警

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 联系我们

- 项目维护者: 太上老君AI团队
- 邮箱: health@taishanglaojun.ai
- 文档: https://docs.taishanglaojun.ai/health-management

## 🗺️ 路线图

### v1.0 (当前版本)
- [x] 基础健康数据管理
- [x] 健康档案管理
- [x] RESTful API
- [x] Docker 支持

### v1.1 (计划中)
- [ ] 实时数据流处理
- [ ] 机器学习健康预测
- [ ] 移动端SDK
- [ ] 第三方设备集成

### v2.0 (未来版本)
- [ ] 区块链健康数据存储
- [ ] 联邦学习支持
- [ ] 多租户架构
- [ ] 国际化支持

---

**太上老君AI平台 - 让健康管理更智能** 🏥✨