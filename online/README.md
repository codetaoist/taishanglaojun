# 码道 (Code Taoist) 中心化平台服务

码道是一个融合道家哲学的智能编程助手平台，提供AI驱动的编程辅助、项目知识库管理和团队协作功能。

## 🏗️ 架构概览

本项目采用微服务架构，包含以下核心组件：

### 核心服务
- **AI智能体服务** (Python/FastAPI) - 处理自然语言编程指令
- **用户管理服务** (Go) - 用户和团队管理
- **记忆知识库服务** (Go) - 项目知识库和向量搜索
- **认证授权服务** (Keycloak) - 统一身份认证
- **API网关** (Nginx) - 请求路由和负载均衡

### 基础设施
- **PostgreSQL** - 主数据库
- **Redis** - 缓存和会话存储
- **Chroma** - 向量数据库
- **MinIO** - 对象存储
- **Prometheus + Grafana** - 监控和可视化

## 🚀 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- Git
- 至少 8GB 可用内存
- 至少 20GB 可用磁盘空间

### 一键部署

1. **克隆项目**
   ```bash
   git clone https://github.com/codetaoist/taishanglaojun.git
   cd taishanglaojun/online
   ```

2. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，设置必要的配置
   ```

3. **启动所有服务**
   ```bash
   docker-compose up -d
   ```

4. **等待服务启动**
   ```bash
   # 检查服务状态
   docker-compose ps
   
   # 查看日志
   docker-compose logs -f ai-service
   ```

5. **访问服务**
   - API网关: http://localhost
   - AI服务文档: http://localhost:8001/docs
   - Keycloak管理: http://localhost:8080
   - Grafana监控: http://localhost:3000
   - MinIO控制台: http://localhost:9001

## 📋 服务端口映射

| 服务 | 内部端口 | 外部端口 | 描述 |
|------|----------|----------|------|
| API网关 | 80/443 | 80/443 | 主入口 |
| AI智能体服务 | 8000 | 8001 | AI API |
| 用户管理服务 | 8080 | 8002 | 用户API |
| 记忆知识库服务 | 8080 | 8003 | 知识库API |
| PostgreSQL | 5432 | 5432 | 数据库 |
| Redis | 6379 | 6379 | 缓存 |
| Chroma | 8000 | 8000 | 向量数据库 |
| MinIO | 9000/9001 | 9000/9001 | 对象存储 |
| Keycloak | 8080 | 8080 | 认证服务 |
| Prometheus | 9090 | 9090 | 监控 |
| Grafana | 3000 | 3000 | 可视化 |

## 🔧 开发指南

### 本地开发环境

1. **启动基础设施**
   ```bash
   # 只启动数据库和缓存
   docker-compose up -d postgres redis chroma minio keycloak
   ```

2. **开发AI服务**
   ```bash
   cd ai-service
   python -m venv venv
   source venv/bin/activate  # Windows: venv\Scripts\activate
   pip install -r requirements.txt
   uvicorn main:app --reload --port 8001
   ```

3. **开发Go服务**
   ```bash
   cd user-service  # 或 memory-service
   go mod download
   go run main.go
   ```

### API 测试

```bash
# 健康检查
curl http://localhost:8001/health

# AI聊天接口 (需要认证)
curl -X POST http://localhost:8001/api/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "message": "用Go写一个HTTP服务器",
    "project_id": "proj_123"
  }'
```

## 🔐 认证配置

### Keycloak 初始设置

1. **访问管理控制台**
   - URL: http://localhost:8080
   - 用户名: admin
   - 密码: admin123 (可在.env中修改)

2. **创建Realm**
   - 创建名为 `codetaoist` 的realm
   - 配置客户端和用户

3. **配置客户端**
   - 客户端ID: `ai-service`
   - 客户端协议: `openid-connect`
   - 访问类型: `confidential`

## 📊 监控和日志

### Grafana 仪表盘

1. **访问Grafana**
   - URL: http://localhost:3000
   - 用户名: admin
   - 密码: admin123

2. **导入仪表盘**
   - 系统指标: `monitoring/grafana/dashboards/system.json`
   - 应用指标: `monitoring/grafana/dashboards/application.json`

### 日志查看

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f ai-service

# 查看错误日志
docker-compose logs --tail=100 ai-service | grep ERROR
```

## 🔧 故障排除

### 常见问题

1. **服务启动失败**
   ```bash
   # 检查端口占用
   netstat -tulpn | grep :8080
   
   # 重新构建镜像
   docker-compose build --no-cache ai-service
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec postgres pg_isready -U codetaoist
   
   # 重置数据库
   docker-compose down -v
   docker-compose up -d postgres
   ```

3. **内存不足**
   ```bash
   # 检查资源使用
   docker stats
   
   # 清理未使用的镜像
   docker system prune -a
   ```

### 性能优化

1. **调整资源限制**
   ```yaml
   # 在 docker-compose.yml 中添加
   deploy:
     resources:
       limits:
         memory: 2G
         cpus: '1.0'
   ```

2. **数据库优化**
   ```bash
   # 调整PostgreSQL配置
   # 编辑 config/postgresql.conf
   ```

## 🚀 生产部署

### Kubernetes 部署

```bash
# 使用 Helm Charts
helm install codetaoist ./charts/codetaoist \
  --namespace codetaoist \
  --create-namespace \
  -f values-production.yaml
```

### 环境变量配置

生产环境必须配置的环境变量：

```bash
# 安全配置
SECRET_KEY=your-production-secret-key
POSTGRES_PASSWORD=strong-database-password
KEYCLOAK_ADMIN_PASSWORD=strong-admin-password

# API 密钥
OPENAI_API_KEY=your-openai-api-key
DEEPSEEK_API_KEY=your-deepseek-api-key

# 域名配置
API_DOMAIN=api.yourdomain.com
ALLOWED_HOSTS=yourdomain.com,api.yourdomain.com
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 获取帮助

- 📖 [官方文档](https://docs.codetaoist.com)
- 🐛 [问题反馈](https://github.com/codetaoist/platform/issues)
- 💬 [讨论区](https://github.com/codetaoist/platform/discussions)
- 📧 [邮件支持](mailto:support@codetaoist.com)

---

**码道 (Code Taoist)** - 让编程回归本质，专注创造 🎯