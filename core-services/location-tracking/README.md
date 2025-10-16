# 位置跟踪模块 (Location Tracking Module)

## 概述

位置跟踪模块负责处理移动应用的位置数据接收、存储、查询和管理功能。该模块提供安全的API接口，支持实时位置上传、轨迹管理和数据同步。

## 功能特性

### 核心功能
- **位置数据接收**: 接收来自移动端的GPS位置数据
- **轨迹管理**: 创建、更新、删除和查询用户轨迹
- **数据存储**: 安全存储位置点和轨迹数据
- **数据同步**: 支持多设备间的数据同步
- **权限控制**: 基于用户身份的数据访问控制

### 安全特性
- **数据加密**: 支持传输和存储加密
- **身份验证**: JWT令牌验证
- **权限管理**: 细粒度的数据访问权限
- **数据完整性**: 校验和验证机制

## API接口

### 位置点管理
- `POST /api/v1/locations/points` - 上传位置点
- `GET /api/v1/locations/points` - 查询位置点
- `DELETE /api/v1/locations/points/:id` - 删除位置点

### 轨迹管理
- `POST /api/v1/locations/trajectories` - 创建轨迹
- `GET /api/v1/locations/trajectories` - 获取轨迹列表
- `GET /api/v1/locations/trajectories/:id` - 获取轨迹详情
- `PUT /api/v1/locations/trajectories/:id` - 更新轨迹
- `DELETE /api/v1/locations/trajectories/:id` - 删除轨迹

### 数据同步
- `POST /api/v1/locations/sync` - 数据同步
- `GET /api/v1/locations/sync/status` - 同步状态查询

## 数据模型

### LocationPoint (位置点)
```go
type LocationPoint struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    UserID      string    `json:"user_id" gorm:"index"`
    TrajectoryID string   `json:"trajectory_id" gorm:"index"`
    Latitude    float64   `json:"latitude"`
    Longitude   float64   `json:"longitude"`
    Altitude    *float64  `json:"altitude,omitempty"`
    Accuracy    *float64  `json:"accuracy,omitempty"`
    Speed       *float64  `json:"speed,omitempty"`
    Bearing     *float64  `json:"bearing,omitempty"`
    Timestamp   int64     `json:"timestamp"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Trajectory (轨迹)
```go
type Trajectory struct {
    ID           string    `json:"id" gorm:"primaryKey"`
    UserID       string    `json:"user_id" gorm:"index"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    StartTime    int64     `json:"start_time"`
    EndTime      *int64    `json:"end_time,omitempty"`
    Distance     float64   `json:"distance"`
    Duration     int64     `json:"duration"`
    MaxSpeed     float64   `json:"max_speed"`
    AvgSpeed     float64   `json:"avg_speed"`
    PointCount   int       `json:"point_count"`
    MinLatitude  float64   `json:"min_latitude"`
    MaxLatitude  float64   `json:"max_latitude"`
    MinLongitude float64   `json:"min_longitude"`
    MaxLongitude float64   `json:"max_longitude"`
    IsActive     bool      `json:"is_active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

## 使用示例

### 上传位置点
```bash
curl -X POST "http://localhost:8080/api/v1/locations/points" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "trajectory_id": "traj_123",
    "latitude": 39.9042,
    "longitude": 116.4074,
    "accuracy": 5.0,
    "speed": 1.2,
    "timestamp": 1640995200000
  }'
```

### 获取轨迹列表
```bash
curl -X GET "http://localhost:8080/api/v1/locations/trajectories" \
  -H "Authorization: Bearer <token>"
```

## 配置说明

### 环境变量
- `LOCATION_DB_HOST`: 数据库主机地址
- `LOCATION_DB_PORT`: 数据库端口
- `LOCATION_DB_NAME`: 数据库名称
- `LOCATION_ENCRYPTION_KEY`: 数据加密密钥
- `LOCATION_MAX_POINTS_PER_REQUEST`: 单次请求最大位置点数量

### 配置文件
```yaml
location_tracking:
  max_points_per_request: 1000
  encryption:
    enabled: true
    algorithm: "AES-256-GCM"
  storage:
    retention_days: 365
    cleanup_interval: "24h"
```

## 部署说明

### Docker部署
```bash
docker build -t location-tracking .
docker run -p 8080:8080 location-tracking
```

### 数据库迁移
```bash
go run cmd/migrate.go
```

## 开发指南

### 本地开发
1. 安装依赖: `go mod download`
2. 配置环境变量
3. 运行数据库迁移
4. 启动服务: `go run cmd/main.go`

### 测试
```bash
go test ./...
```

## 监控和日志

### 健康检查
- `GET /health` - 服务健康状态
- `GET /metrics` - Prometheus指标

### 日志格式
```json
{
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "info",
  "service": "location-tracking",
  "user_id": "user_123",
  "action": "upload_points",
  "count": 10,
  "duration_ms": 150
}
```