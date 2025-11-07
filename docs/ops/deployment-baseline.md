# 部署基线

## 环境划分
- 开发（dev）：本地与容器化环境，便于快速迭代。
- 测试（test）：与生产接近的预发布环境，含更严格的资源与安全策略。
- 生产（prod）：高可用，含监控告警、灰度发布与回滚。

## 硬件/软件配置（示例）[0]
- CPU：AMD EPYC 或 Intel Xeon，8~32 vCPU（按业务规模调整）。
- 内存：16GB~128GB。
- GPU：NVIDIA A10/A100（根据模型推理需要选择），显存 24GB~80GB。
- OS：Linux (Ubuntu 22.04 / CentOS Stream)。
- 容器：Docker / Kubernetes。
- 数据库：MySQL 8 / PostgreSQL 14（按团队熟悉度选择）。
- 缓存：Redis 6+。
- 向量库：Milvus/Faiss（按检索性能与部署便捷度选择）。

## 网络与安全
- 入口网关：统一鉴权与限流。
- 证书：HTTPS/TLS 强制；服务间 mTLS（可选）。
- 机密管理：Env + Secrets 管理；密钥轮转与最小权限原则。

## 部署策略
- CI/CD：GitLab CI 驱动构建/测试/部署；插件与后端服务流水线衔接。[0]
- 蓝绿/金丝雀发布：生产环境按需使用。
- 监控：Prometheus/Grafana；日志：ELK；告警：Alertmanager。

> 参考：[0] https://www.doubao.com/thread/wefa24b8b54e437a1