# 太上老君AI平台 - 部署运维

## 📋 部署运维概览

本目录包含太上老君AI平台的完整部署和运维指南，为运维团队提供生产环境部署、监控、维护和故障处理的详细说明。

## 📚 运维文档

### 🎯 部署概览
- **[部署概览](./deployment-overview.md)** - 部署架构和流程概述

### 🐳 容器化部署
- **[Docker部署](./docker-deployment.md)** - Docker容器化部署指南
- **[Kubernetes部署](./kubernetes-deployment.md)** - K8s集群部署和管理

### 📊 监控运维
- **[监控运维](./monitoring-operations.md)** - 系统监控和告警配置
- **[性能调优](./performance-tuning.md)** - 系统性能优化指南

### 🔒 安全配置
- **[安全配置](./security-configuration.md)** - 生产环境安全加固

## 🚀 快速部署

### 1. 环境要求
```bash
# 最小系统要求
- CPU: 4核心
- 内存: 8GB
- 存储: 100GB SSD
- 网络: 100Mbps

# 推荐系统要求
- CPU: 8核心
- 内存: 16GB
- 存储: 500GB SSD
- 网络: 1Gbps
```

### 2. Docker快速部署
```bash
# 克隆项目
git clone https://github.com/taishanglaojun/taishanglaojun.git
cd taishanglaojun

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置数据库等信息

# 启动服务
docker-compose up -d

# 检查服务状态
docker-compose ps
```

### 3. Kubernetes部署
```bash
# 创建命名空间
kubectl create namespace taishanglaojun

# 部署配置
kubectl apply -f k8s/

# 检查部署状态
kubectl get pods -n taishanglaojun
```

## 📊 服务状态

| 服务 | 状态 | 版本 | 端口 | 健康检查 |
|------|------|------|------|----------|
| API网关 | ✅ 运行中 | v1.0 | 8080 | /health |
| AI服务 | ✅ 运行中 | v1.2 | 8081 | /health |
| 用户服务 | ✅ 运行中 | v1.1 | 8082 | /health |
| 数据服务 | ✅ 运行中 | v1.0 | 8083 | /health |
| Web前端 | ✅ 运行中 | v1.0 | 3000 | / |

## 🔧 运维工具

### 监控工具
- **Prometheus**: 指标收集
- **Grafana**: 可视化监控
- **AlertManager**: 告警管理
- **Jaeger**: 链路追踪

### 日志工具
- **ELK Stack**: 日志收集和分析
- **Fluentd**: 日志转发
- **Kibana**: 日志可视化

### 部署工具
- **Docker**: 容器化
- **Kubernetes**: 容器编排
- **Helm**: K8s包管理
- **ArgoCD**: GitOps部署

## 📈 监控指标

### 系统指标
- **CPU使用率**: < 70%
- **内存使用率**: < 80%
- **磁盘使用率**: < 85%
- **网络延迟**: < 100ms

### 应用指标
- **API响应时间**: < 500ms
- **错误率**: < 1%
- **QPS**: 监控请求量
- **可用性**: > 99.9%

## 🚨 故障处理

### 常见问题
1. **服务无响应**: 检查容器状态和日志
2. **数据库连接失败**: 检查数据库服务和网络
3. **内存不足**: 检查内存使用和配置
4. **磁盘空间不足**: 清理日志和临时文件

### 应急联系
- **运维值班**: ops-oncall@taishanglaojun.com
- **技术支持**: tech-support@taishanglaojun.com
- **紧急电话**: +86-xxx-xxxx-xxxx

## 📖 相关文档

- **[项目概览](../00-项目概览/README.md)** - 平台整体介绍
- **[架构设计](../02-架构设计/README.md)** - 技术架构详解
- **[核心服务](../03-核心服务/README.md)** - 后端服务文档
- **[开发指南](../07-开发指南/README.md)** - 开发规范和指南
- **[基础设施](../05-基础设施/README.md)** - 基础设施文档

## 📋 运维检查清单

### 日常检查
- [ ] 服务健康状态
- [ ] 系统资源使用率
- [ ] 错误日志检查
- [ ] 备份状态确认

### 周期检查
- [ ] 安全补丁更新
- [ ] 性能指标分析
- [ ] 容量规划评估
- [ ] 灾备演练

---

**最后更新**: 2024年12月  
**文档版本**: v1.0  
**维护团队**: 运维团队