# Kubernetes部署指南 - 高级AI功能

本目录包含了将高级AI功能部署到Kubernetes集群的完整配置文件和脚本。

## 📁 文件结构

```
k8s/
├── namespace.yaml          # 命名空间和资源配额
├── configmap.yaml         # 配置映射
├── secrets.yaml           # 密钥配置
├── storage.yaml           # 存储配置
├── rbac.yaml             # RBAC和安全配置
├── deployment.yaml        # 部署配置
├── service.yaml          # 服务配置
├── ingress.yaml          # Ingress配置
├── hpa.yaml              # 水平Pod自动扩缩
├── monitoring.yaml       # 监控和告警配置
├── deploy-k8s.sh         # 部署脚本
└── README.md             # 本文档
```

## 🚀 快速开始

### 前置条件

1. **Kubernetes集群** (v1.20+)
2. **kubectl** 已配置并连接到集群
3. **Docker** 用于构建镜像
4. **Helm** (可选，用于某些组件)

### 基本部署

```bash
# 1. 克隆项目
git clone <repository-url>
cd core-services/ai-integration/k8s

# 2. 配置环境变量
export IMAGE_TAG=v1.0.0
export ENVIRONMENT=production

# 3. 执行部署
chmod +x deploy-k8s.sh
./deploy-k8s.sh deploy
```

### 自定义部署

```bash
# 开发环境部署
ENVIRONMENT=development ./deploy-k8s.sh deploy

# 指定镜像版本
IMAGE_TAG=v2.0.0 ./deploy-k8s.sh deploy

# 更新现有部署
IMAGE_TAG=v2.1.0 ./deploy-k8s.sh update
```

## 📋 详细配置

### 1. 命名空间配置 (namespace.yaml)

- 创建 `taishanglaojun` 命名空间
- 设置资源配额限制
- 配置默认资源限制

### 2. 密钥管理 (secrets.yaml)

⚠️ **重要**: 生产环境部署前必须更新以下密钥:

```bash
# JWT密钥
kubectl create secret generic advanced-ai-secret \
  --from-literal=jwt-secret="your-super-secret-jwt-key" \
  -n taishanglaojun

# 数据库密码
kubectl create secret generic advanced-ai-secret \
  --from-literal=db-password="your-database-password" \
  -n taishanglaojun

# AI提供商API密钥
kubectl create secret generic ai-provider-secrets \
  --from-literal=openai-api-key="your-openai-key" \
  --from-literal=anthropic-api-key="your-anthropic-key" \
  -n taishanglaojun
```

### 3. 存储配置 (storage.yaml)

- **StorageClass**: 使用SSD存储
- **PersistentVolumeClaim**: 为各服务分配存储
- **VolumeSnapshot**: 数据库备份支持

### 4. 网络配置 (service.yaml, ingress.yaml)

- **服务发现**: ClusterIP服务用于内部通信
- **负载均衡**: LoadBalancer服务用于外部访问
- **SSL终止**: Ingress处理HTTPS
- **域名路由**: 支持多域名访问

### 5. 自动扩缩 (hpa.yaml)

- **水平扩缩**: 基于CPU、内存和自定义指标
- **垂直扩缩**: 自动调整资源请求和限制
- **集群扩缩**: 节点自动扩缩配置

## 🔧 运维操作

### 查看部署状态

```bash
# 查看所有资源
kubectl get all -n taishanglaojun

# 查看Pod状态
kubectl get pods -n taishanglaojun -o wide

# 查看服务状态
kubectl get services -n taishanglaojun

# 查看Ingress状态
kubectl get ingress -n taishanglaojun
```

### 查看日志

```bash
# 查看应用日志
./deploy-k8s.sh logs <pod-name>

# 实时查看日志
kubectl logs -f deployment/advanced-ai-service -n taishanglaojun

# 查看所有容器日志
kubectl logs -f deployment/advanced-ai-service -n taishanglaojun --all-containers
```

### 进入容器

```bash
# 进入应用容器
./deploy-k8s.sh exec <pod-name>

# 或直接使用kubectl
kubectl exec -it deployment/advanced-ai-service -n taishanglaojun -- /bin/bash
```

### 更新部署

```bash
# 更新镜像
kubectl set image deployment/advanced-ai-service \
  advanced-ai-service=your-registry.com/advanced-ai-service:v2.0.0 \
  -n taishanglaojun

# 查看滚动更新状态
kubectl rollout status deployment/advanced-ai-service -n taishanglaojun

# 回滚到上一版本
kubectl rollout undo deployment/advanced-ai-service -n taishanglaojun
```

### 扩缩容操作

```bash
# 手动扩容
kubectl scale deployment advanced-ai-service --replicas=5 -n taishanglaojun

# 查看HPA状态
kubectl get hpa -n taishanglaojun

# 查看资源使用情况
kubectl top pods -n taishanglaojun
kubectl top nodes
```

## 📊 监控和告警

### Prometheus指标

访问 `https://prometheus.taishanglaojun.com` 查看指标:

- **应用指标**: HTTP请求、响应时间、错误率
- **AI指标**: AGI推理延迟、元学习准确率、自我进化分数
- **基础设施指标**: CPU、内存、磁盘、网络

### Grafana仪表板

访问 `https://grafana.taishanglaojun.com` 查看仪表板:

- **应用性能仪表板**: 请求量、响应时间、错误率
- **AI功能仪表板**: AGI、元学习、自我进化指标
- **基础设施仪表板**: 集群资源使用情况

### 告警规则

配置的告警包括:

- **服务可用性**: 服务下线告警
- **性能指标**: 高CPU、内存使用率
- **错误率**: HTTP 5xx错误率过高
- **AI功能**: 推理延迟、训练失败

## 🔒 安全配置

### RBAC权限

- **ServiceAccount**: 为每个服务创建专用账户
- **Role/ClusterRole**: 最小权限原则
- **NetworkPolicy**: 网络隔离和访问控制

### Pod安全策略

- **非特权容器**: 禁止特权访问
- **只读根文件系统**: 增强安全性
- **资源限制**: 防止资源耗尽攻击

### 网络安全

- **TLS加密**: 所有外部通信使用HTTPS
- **内部通信**: 服务间通信加密
- **访问控制**: IP白名单和认证

## 🚨 故障排除

### 常见问题

1. **Pod启动失败**
   ```bash
   kubectl describe pod <pod-name> -n taishanglaojun
   kubectl logs <pod-name> -n taishanglaojun
   ```

2. **服务无法访问**
   ```bash
   kubectl get endpoints -n taishanglaojun
   kubectl describe service <service-name> -n taishanglaojun
   ```

3. **Ingress配置问题**
   ```bash
   kubectl describe ingress -n taishanglaojun
   kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
   ```

4. **存储问题**
   ```bash
   kubectl get pv,pvc -n taishanglaojun
   kubectl describe pvc <pvc-name> -n taishanglaojun
   ```

### 调试命令

```bash
# 检查集群状态
kubectl cluster-info
kubectl get nodes

# 检查资源配额
kubectl describe quota -n taishanglaojun

# 检查网络策略
kubectl get networkpolicy -n taishanglaojun

# 检查证书
kubectl get certificates -n taishanglaojun
```

## 📈 性能优化

### 资源调优

1. **CPU和内存**: 根据实际使用情况调整requests和limits
2. **存储**: 使用SSD存储提高I/O性能
3. **网络**: 配置适当的网络策略和负载均衡

### 缓存优化

1. **Redis配置**: 调整内存和持久化策略
2. **应用缓存**: 配置多级缓存策略
3. **CDN**: 静态资源使用CDN加速

### 数据库优化

1. **连接池**: 调整数据库连接池大小
2. **索引**: 优化数据库查询索引
3. **分片**: 考虑数据库分片策略

## 🔄 备份和恢复

### 数据备份

```bash
# 备份配置
./deploy-k8s.sh backup

# 数据库备份
kubectl exec -n taishanglaojun deployment/postgres -- \
  pg_dump -U postgres taishanglaojun > backup.sql

# 创建卷快照
kubectl create -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: postgres-snapshot-$(date +%Y%m%d)
  namespace: taishanglaojun
spec:
  source:
    persistentVolumeClaimName: postgres-pvc
  volumeSnapshotClassName: csi-aws-vsc
EOF
```

### 数据恢复

```bash
# 恢复配置
./deploy-k8s.sh restore backup.yaml

# 恢复数据库
kubectl exec -i -n taishanglaojun deployment/postgres -- \
  psql -U postgres taishanglaojun < backup.sql
```

## 🌐 多环境部署

### 开发环境

```bash
ENVIRONMENT=development ./deploy-k8s.sh deploy
```

### 测试环境

```bash
ENVIRONMENT=staging ./deploy-k8s.sh deploy
```

### 生产环境

```bash
ENVIRONMENT=production ./deploy-k8s.sh deploy
```

## 📞 支持和联系

如有问题或需要支持，请联系:

- **技术支持**: tech-support@taishanglaojun.com
- **文档**: https://docs.taishanglaojun.com
- **GitHub**: https://github.com/taishanglaojun/advanced-ai

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../../../LICENSE) 文件。