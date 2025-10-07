# 智能学习系统 (Intelligent Learning System)

## 概述

智能学习系统是太上老君平台的核心服务之一，提供个性化学习路径推荐和知识图谱管理功能。系统基于现代微服务架构设计，支持大规模用户的个性化学习需求。

## 核心功能

### 🎯 学习者管理
- 学习者档案管理
- 学习目标设置与跟踪
- 技能评估与进度监控
- 学习活动记录与分析
- 个性化推荐系统

### 📚 学习内容管理
- 多媒体学习资源管理
- 内容标签与分类
- 全文搜索与智能推荐
- 内容发布与版本控制
- 学习效果分析

### 🕸️ 知识图谱
- 知识节点与关系管理
- 学习路径智能生成
- 概念地图可视化
- 知识依赖分析
- 图谱统计与验证

## 新增智能学习服务

### 🤖 跨模态服务 (Cross-Modal Service)
- 支持文本、图像、音频、视频等多种模态的内容处理
- 提供统一的多模态内容分析接口
- 支持模态间的特征融合和转换

### 🧠 智能关系推理引擎 (Intelligent Relation Inference Engine)
- 基于语义相似度、图神经网络、Transformer等算法
- 自动推理实体间的关系
- 支持多层次的关系推理

### 📈 自适应学习引擎 (Adaptive Learning Engine)
- 个性化学习路径规划
- 动态难度调整
- 基于学习者行为的内容推荐

### ⚡ 实时学习分析服务 (Real-time Learning Analytics Service)
- 实时监控学习行为和进度
- 智能预警和干预
- 多维度学习数据分析

### 🔗 自动化知识图谱服务 (Automated Knowledge Graph Service)
- 自动构建和维护知识图谱
- 实体和关系抽取
- 知识图谱优化和验证

### 📊 学习分析报告服务 (Learning Analytics Reporting Service)
- 生成多种类型的学习分析报告
- 支持多种导出格式
- 可视化图表生成

### 💡 智能内容推荐服务 (Intelligent Content Recommendation Service)
- 基于协同过滤、内容推荐、混合推荐算法
- 个性化内容推荐
- 学习者画像构建

## 技术架构

### 架构模式
- **领域驱动设计 (DDD)**: 清晰的业务边界和领域模型
- **六边形架构**: 松耦合的端口适配器模式
- **微服务架构**: 独立部署和扩展

### 技术栈
- **后端**: Go 1.21+, Gin Web Framework
- **数据库**: PostgreSQL (主数据库)
- **缓存**: Redis
- **搜索**: Elasticsearch
- **图数据库**: Neo4j
- **认证**: JWT
- **文档**: Swagger/OpenAPI
- **监控**: 结构化日志 + 健康检查

## 项目结构

```
intelligent-learning/
├── cmd/
│   └── server/           # 应用程序入口
├── internal/
│   ├── domain/          # 领域层
│   │   ├── entities/    # 实体
│   │   ├── repositories/ # 仓储接口
│   │   └── services/    # 领域服务
│   ├── application/     # 应用层
│   │   └── services/    # 应用服务
│   ├── infrastructure/  # 基础设施层
│   │   ├── config/      # 配置管理
│   │   └── persistence/ # 数据持久化
│   └── interfaces/      # 接口层
│       └── http/        # HTTP接口
├── configs/             # 配置文件
├── docs/               # 文档
└── scripts/            # 脚本文件
```

## 快速开始

### 环境要求
- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- Elasticsearch 7+
- Neo4j 4+

### 安装依赖
```bash
go mod download
```

### 配置环境
1. 复制配置文件模板：
```bash
cp configs/config.yaml.example configs/config.yaml
```

2. 修改配置文件中的数据库连接信息

### 运行服务
```bash
# 开发模式
go run cmd/server/main.go

# 或使用 Makefile
make run
```

### API 文档
启动服务后访问：http://localhost:8080/swagger/index.html

## API 接口

### 学习者管理
- `POST /api/v1/learners` - 创建学习者
- `GET /api/v1/learners/{id}` - 获取学习者信息
- `PUT /api/v1/learners/{id}` - 更新学习者信息
- `DELETE /api/v1/learners/{id}` - 删除学习者
- `POST /api/v1/learners/{id}/goals` - 添加学习目标
- `POST /api/v1/learners/{id}/activities` - 记录学习活动
- `GET /api/v1/learners/{id}/analytics` - 获取学习分析报告

### 学习内容管理
- `POST /api/v1/content` - 创建学习内容
- `GET /api/v1/content/{id}` - 获取内容详情
- `PUT /api/v1/content/{id}` - 更新内容
- `DELETE /api/v1/content/{id}` - 删除内容
- `GET /api/v1/content/search` - 搜索内容
- `GET /api/v1/content/personalized` - 个性化推荐

### 知识图谱
- `POST /api/v1/knowledge-graph/nodes` - 创建知识节点
- `POST /api/v1/knowledge-graph/relations` - 创建知识关系
- `GET /api/v1/knowledge-graph/nodes/search` - 搜索节点
- `GET /api/v1/knowledge-graph/learning-path` - 生成学习路径
- `GET /api/v1/knowledge-graph/concept-map` - 生成概念地图

## 开发指南

### 代码规范
- 遵循 Go 官方代码规范
- 使用 gofmt 格式化代码
- 编写单元测试和集成测试
- 添加适当的注释和文档

### 测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 构建
```bash
# 构建二进制文件
make build

# 构建 Docker 镜像
make docker-build
```

## 部署

### Docker 部署
```bash
# 构建镜像
docker build -t intelligent-learning:latest .

# 运行容器
docker run -p 8080:8080 intelligent-learning:latest
```

### Docker Compose 部署
```bash
docker-compose up -d
```

## 监控与运维

### 健康检查
- `GET /health` - 基础健康检查
- `GET /health/ready` - 就绪状态检查
- `GET /health/live` - 存活状态检查

### 日志
系统使用结构化日志，支持不同级别的日志输出：
- DEBUG: 调试信息
- INFO: 一般信息
- WARN: 警告信息
- ERROR: 错误信息

### 性能监控
- 请求响应时间监控
- 数据库连接池监控
- 缓存命中率监控
- API 调用频率限制

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目维护者: 太上老君开发团队
- 邮箱: dev@taishanglaojun.com
- 项目地址: https://github.com/taishanglaojun/core-services

## 更新日志

### v1.0.0 (2024-01-XX)
- 初始版本发布
- 实现基础的学习者管理功能
- 实现学习内容管理功能
- 实现知识图谱管理功能
- 支持个性化学习路径推荐