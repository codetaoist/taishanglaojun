# 任务管理系统 (Task Management System)

一个基于Go语言开发的企业级任务管理系统，采用领域驱动设计(DDD)架构，提供完整的任务分配、调度、性能分析和通知功能。

## 功能特性

### 核心功能
- **任务管理**: 创建、更新、删除、查询任务
- **项目管理**: 项目生命周期管理，成员管理，里程碑跟踪
- **团队管理**: 团队组建，成员管理，技能管理
- **智能分配**: 基于技能、工作量、优先级的自动任务分配
- **调度优化**: 任务调度，冲突检测，关键路径分析
- **性能分析**: 用户、团队、项目性能分析和趋势预测
- **通知系统**: 多渠道通知（邮件、短信、推送）

### 技术特性
- **领域驱动设计**: 清晰的业务逻辑分层
- **RESTful API**: 标准化的HTTP接口
- **中间件支持**: 日志、认证、限流、CORS等
- **容器化部署**: Docker和Docker Compose支持
- **监控集成**: Prometheus和Grafana监控
- **优雅关闭**: 支持优雅的服务关闭

## 架构设计

```
├── cmd/                    # 应用程序入口
│   └── server/            # HTTP服务器
├── internal/              # 内部代码
│   ├── domain/           # 领域层
│   │   ├── entities/     # 实体
│   │   ├── repositories/ # 仓储接口
│   │   └── services/     # 领域服务接口
│   ├── application/      # 应用层
│   │   └── services/     # 应用服务
│   ├── infrastructure/   # 基础设施层
│   │   ├── persistence/  # 数据持久化
│   │   └── services/     # 基础设施服务
│   └── interfaces/       # 接口层
│       └── http/         # HTTP接口
├── scripts/              # 脚本文件
├── monitoring/           # 监控配置
└── nginx/               # Nginx配置
```

## 快速开始

### 环境要求
- Go 1.21+
- Docker & Docker Compose
- Make (可选)

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd task-management
```

2. **安装依赖**
```bash
make deps
# 或者
go mod download
```

3. **运行开发服务器**
```bash
make dev
# 或者
go run ./cmd/server
```

4. **运行测试**
```bash
make test
# 或者
go test ./...
```

### Docker部署

1. **构建镜像**
```bash
make docker-build
# 或者
docker build -t task-management .
```

2. **启动服务**
```bash
make docker-up
# 或者
docker-compose up -d
```

3. **查看日志**
```bash
make logs
# 或者
docker-compose logs -f task-management
```

## API文档

### 任务管理

#### 创建任务
```http
POST /api/v1/tasks
Content-Type: application/json

{
  "title": "任务标题",
  "description": "任务描述",
  "priority": "high",
  "estimated_hours": 8,
  "due_date": "2024-12-31T23:59:59Z",
  "project_id": "project-uuid",
  "tags": ["backend", "api"]
}
```

#### 获取任务
```http
GET /api/v1/tasks/{id}
```

#### 更新任务
```http
PUT /api/v1/tasks/{id}
Content-Type: application/json

{
  "title": "更新的标题",
  "status": "in_progress"
}
```

#### 分配任务
```http
POST /api/v1/tasks/{id}/assign
Content-Type: application/json

{
  "assignee_id": "user-uuid"
}
```

#### 自动分配任务
```http
POST /api/v1/tasks/auto-assign
Content-Type: application/json

{
  "task_ids": ["task-uuid-1", "task-uuid-2"],
  "strategy": "balanced"
}
```

### 项目管理

#### 创建项目
```http
POST /api/v1/projects
Content-Type: application/json

{
  "name": "项目名称",
  "description": "项目描述",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-12-31T23:59:59Z",
  "budget": 100000
}
```

#### 添加项目成员
```http
POST /api/v1/projects/{id}/members
Content-Type: application/json

{
  "user_id": "user-uuid",
  "role": "developer"
}
```

### 团队管理

#### 创建团队
```http
POST /api/v1/teams
Content-Type: application/json

{
  "name": "开发团队",
  "description": "后端开发团队",
  "department": "技术部"
}
```

#### 添加团队成员
```http
POST /api/v1/teams/{id}/members
Content-Type: application/json

{
  "user_id": "user-uuid",
  "role": "developer",
  "skills": ["Go", "PostgreSQL", "Docker"]
}
```

## 配置

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| PORT | 服务端口 | 8080 |
| READ_TIMEOUT | 读取超时 | 15s |
| WRITE_TIMEOUT | 写入超时 | 15s |
| IDLE_TIMEOUT | 空闲超时 | 60s |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户 | taskuser |
| DB_PASSWORD | 数据库密码 | taskpass |
| DB_NAME | 数据库名称 | taskdb |
| REDIS_HOST | Redis主机 | localhost |
| REDIS_PORT | Redis端口 | 6379 |

## 监控

系统集成了Prometheus和Grafana进行监控：

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### 监控指标

- HTTP请求数量和延迟
- 数据库连接池状态
- 内存和CPU使用率
- 业务指标（任务数量、完成率等）

## 开发指南

### 代码规范

1. **遵循Go语言规范**
2. **使用gofmt格式化代码**
3. **编写单元测试**
4. **添加适当的注释**

### 提交规范

```
type(scope): description

[optional body]

[optional footer]
```

类型：
- feat: 新功能
- fix: 修复
- docs: 文档
- style: 格式
- refactor: 重构
- test: 测试
- chore: 构建

### 测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/domain/entities

# 运行基准测试
make bench

# 生成覆盖率报告
make coverage
```

## 部署

### 生产环境部署

1. **构建生产镜像**
```bash
make release
```

2. **部署到Kubernetes**
```bash
kubectl apply -f k8s/
```

3. **配置负载均衡器**
4. **设置监控和日志收集**

### 性能优化

- 使用连接池
- 启用HTTP/2
- 配置适当的缓存策略
- 优化数据库查询
- 使用CDN加速静态资源

## 故障排除

### 常见问题

1. **服务启动失败**
   - 检查端口是否被占用
   - 验证环境变量配置
   - 查看日志文件

2. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接参数
   - 检查网络连通性

3. **性能问题**
   - 查看监控指标
   - 分析慢查询日志
   - 检查资源使用情况

### 日志级别

- ERROR: 错误信息
- WARN: 警告信息
- INFO: 一般信息
- DEBUG: 调试信息

## 贡献

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License

## 联系方式

- 项目维护者: [维护者姓名]
- 邮箱: [邮箱地址]
- 问题反馈: [GitHub Issues链接]