# 太上老君AI平台 - Kubernetes部署指南

## 概述

本文档详细介绍如何在Kubernetes集群中部署太上老君AI平台，包括集群准备、资源配置、服务部署、扩缩容管理等内容。

## 集群架构

```mermaid
graph TB
    subgraph "Kubernetes集群"
        subgraph "Master节点"
            A[API Server]
            B[etcd]
            C[Controller Manager]
            D[Scheduler]
        end
        
        subgraph "Worker节点1"
            E[kubelet]
            F[kube-proxy]
            G[Container Runtime]
            H[Pod1]
            I[Pod2]
        end
        
        subgraph "Worker节点2"
            J[kubelet]
            K[kube-proxy]
            L[Container Runtime]
            M[Pod3]
            N[Pod4]
        end
        
        subgraph "Worker节点3"
            O[kubelet]
            P[kube-proxy]
            Q[Container Runtime]
            R[Pod5]
            S[Pod6]
        end
    end
    
    subgraph "外部组件"
        T[Load Balancer]
        U[Ingress Controller]
        V[DNS]
        W[监控系统]
    end
    
    T --> U
    U --> H
    U --> M
    U --> R
```

## 集群准备

### 1. 集群要求

#### 最小配置
- **Master节点**: 2 CPU, 4GB RAM, 20GB 存储
- **Worker节点**: 4 CPU, 8GB RAM, 50GB 存储
- **节点数量**: 1个Master + 3个Worker

#### 推荐配置
- **Master节点**: 4 CPU, 8GB RAM, 100GB SSD
- **Worker节点**: 8 CPU, 16GB RAM, 200GB SSD
- **节点数量**: 3个Master + 5个Worker

#### 生产配置
- **Master节点**: 8 CPU, 16GB RAM, 500GB SSD
- **Worker节点**: 16 CPU, 32GB RAM, 1TB SSD
- **节点数量**: 3个Master + 10个Worker

### 2. 集群初始化

#### 使用kubeadm初始化

```bash
# 初始化Master节点
sudo kubeadm init \
  --pod-network-cidr=10.244.0.0/16 \
  --service-cidr=10.96.0.0/12 \
  --kubernetes-version=v1.28.0 \
  --control-plane-endpoint=k8s-master.taishanglaojun.com:6443 \
  --upload-certs

# 配置kubectl
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# 安装网络插件（Flannel）
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml

# 加入Worker节点
kubeadm join k8s-master.taishanglaojun.com:6443 \
  --token <token> \
  --discovery-token-ca-cert-hash sha256:<hash>
```

#### 使用Terraform部署（AWS EKS）

```hcl
# terraform/eks.tf
provider "aws" {
  region = var.aws_region
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  
  name = "taishanglaojun-vpc"
  cidr = "10.0.0.0/16"
  
  azs             = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
  
  enable_nat_gateway = true
  enable_vpn_gateway = true
  
  tags = {
    Terraform = "true"
    Environment = var.environment
  }
}

module "eks" {
  source = "terraform-aws-modules/eks/aws"
  
  cluster_name    = "taishanglaojun-${var.environment}"
  cluster_version = "1.28"
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  
  node_groups = {
    main = {
      desired_capacity = 3
      max_capacity     = 10
      min_capacity     = 1
      
      instance_types = ["t3.large"]
      
      k8s_labels = {
        Environment = var.environment
        Application = "taishanglaojun"
      }
    }
  }
  
  tags = {
    Environment = var.environment
    Application = "taishanglaojun"
  }
}
```

### 3. 必要组件安装

#### Ingress Controller (Nginx)

```yaml
# k8s/ingress-nginx.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-nginx

---
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: ingress-nginx
  namespace: kube-system
spec:
  chart: ingress-nginx
  repo: https://kubernetes.github.io/ingress-nginx
  targetNamespace: ingress-nginx
  valuesContent: |-
    controller:
      service:
        type: LoadBalancer
        annotations:
          service.beta.kubernetes.io/aws-load-balancer-type: nlb
      metrics:
        enabled: true
      podAnnotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "10254"
```

#### Cert-Manager (SSL证书管理)

```yaml
# k8s/cert-manager.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager

---
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: cert-manager
  namespace: kube-system
spec:
  chart: cert-manager
  repo: https://charts.jetstack.io
  targetNamespace: cert-manager
  valuesContent: |-
    installCRDs: true
    global:
      leaderElection:
        namespace: cert-manager

---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@taishanglaojun.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

## 命名空间和资源管理

### 1. 命名空间配置

```yaml
# k8s/namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: taishanglaojun-prod
  labels:
    name: taishanglaojun-prod
    environment: production
    team: platform

---
apiVersion: v1
kind: Namespace
metadata:
  name: taishanglaojun-staging
  labels:
    name: taishanglaojun-staging
    environment: staging
    team: platform

---
apiVersion: v1
kind: Namespace
metadata:
  name: taishanglaojun-dev
  labels:
    name: taishanglaojun-dev
    environment: development
    team: platform

---
# 生产环境资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: taishanglaojun-prod
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    persistentvolumeclaims: "20"
    services: "20"
    secrets: "20"
    configmaps: "20"
    pods: "50"

---
# 资源限制范围
apiVersion: v1
kind: LimitRange
metadata:
  name: mem-limit-range
  namespace: taishanglaojun-prod
spec:
  limits:
  - default:
      memory: "1Gi"
      cpu: "500m"
    defaultRequest:
      memory: "512Mi"
      cpu: "100m"
    type: Container
  - max:
      memory: "4Gi"
      cpu: "2"
    min:
      memory: "128Mi"
      cpu: "50m"
    type: Container
```

### 2. 网络策略

```yaml
# k8s/network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: taishanglaojun-prod
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend-to-backend
  namespace: taishanglaojun-prod
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 8080

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-backend-to-database
  namespace: taishanglaojun-prod
spec:
  podSelector:
    matchLabels:
      app: postgres
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: backend
    - podSelector:
        matchLabels:
          app: ai-service
    ports:
    - protocol: TCP
      port: 5432
```

## 配置管理

### 1. ConfigMap配置

```yaml
# k8s/configmaps.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: taishanglaojun-prod
data:
  # 应用配置
  app.yaml: |
    server:
      port: 8080
      host: "0.0.0.0"
      read_timeout: 30s
      write_timeout: 30s
      idle_timeout: 60s
    
    cors:
      allowed_origins:
        - "https://taishanglaojun.com"
        - "https://www.taishanglaojun.com"
      allowed_methods:
        - GET
        - POST
        - PUT
        - DELETE
        - OPTIONS
      allowed_headers:
        - "*"
      max_age: 86400
  
  # 数据库配置
  database.yaml: |
    host: postgres-service
    port: 5432
    database: taishanglaojun
    sslmode: require
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600s
    
  # Redis配置
  redis.yaml: |
    host: redis-service
    port: 6379
    db: 0
    max_retries: 3
    pool_size: 100
    min_idle_conns: 10
    
  # AI服务配置
  ai.yaml: |
    providers:
      openai:
        base_url: "https://api.openai.com/v1"
        timeout: 30s
        max_retries: 3
      anthropic:
        base_url: "https://api.anthropic.com"
        timeout: 30s
        max_retries: 3
    
    models:
      default_chat: "gpt-3.5-turbo"
      default_embedding: "text-embedding-ada-002"
    
    limits:
      max_tokens: 4096
      max_requests_per_minute: 60

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config
  namespace: taishanglaojun-prod
data:
  nginx.conf: |
    user nginx;
    worker_processes auto;
    error_log /var/log/nginx/error.log warn;
    pid /var/run/nginx.pid;
    
    events {
        worker_connections 1024;
        use epoll;
        multi_accept on;
    }
    
    http {
        include /etc/nginx/mime.types;
        default_type application/octet-stream;
        
        log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                        '$status $body_bytes_sent "$http_referer" '
                        '"$http_user_agent" "$http_x_forwarded_for"';
        
        access_log /var/log/nginx/access.log main;
        
        sendfile on;
        tcp_nopush on;
        tcp_nodelay on;
        keepalive_timeout 65;
        types_hash_max_size 2048;
        
        gzip on;
        gzip_vary on;
        gzip_min_length 10240;
        gzip_proxied expired no-cache no-store private must-revalidate auth;
        gzip_types
            text/plain
            text/css
            text/xml
            text/javascript
            application/javascript
            application/xml+rss
            application/json;
        
        include /etc/nginx/conf.d/*.conf;
    }
```

### 2. Secret管理

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: taishanglaojun-prod
type: Opaque
data:
  # 数据库密码 (base64编码)
  database-password: <base64-encoded-password>
  
  # JWT密钥
  jwt-secret: <base64-encoded-jwt-secret>
  
  # AI服务API密钥
  openai-api-key: <base64-encoded-openai-key>
  anthropic-api-key: <base64-encoded-anthropic-key>
  
  # Redis密码
  redis-password: <base64-encoded-redis-password>
  
  # MinIO访问密钥
  minio-access-key: <base64-encoded-minio-access-key>
  minio-secret-key: <base64-encoded-minio-secret-key>

---
# 使用External Secrets Operator管理外部密钥
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: taishanglaojun-prod
spec:
  provider:
    vault:
      server: "https://vault.taishanglaojun.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "taishanglaojun-role"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets-external
  namespace: taishanglaojun-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: app-secrets-vault
    creationPolicy: Owner
  data:
  - secretKey: database-password
    remoteRef:
      key: taishanglaojun/database
      property: password
  - secretKey: jwt-secret
    remoteRef:
      key: taishanglaojun/auth
      property: jwt-secret
```

## 数据库部署

### 1. PostgreSQL StatefulSet

```yaml
# k8s/postgres.yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: taishanglaojun-prod
  labels:
    app: postgres
spec:
  ports:
  - port: 5432
    name: postgres
  clusterIP: None
  selector:
    app: postgres

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: taishanglaojun-prod
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      securityContext:
        fsGroup: 999
      containers:
      - name: postgres
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
          name: postgres
        env:
        - name: POSTGRES_DB
          value: taishanglaojun
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        - name: postgres-config
          mountPath: /etc/postgresql/postgresql.conf
          subPath: postgresql.conf
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - exec pg_isready -U postgres -h 127.0.0.1 -p 5432
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 6
        readinessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - exec pg_isready -U postgres -h 127.0.0.1 -p 5432
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3
      volumes:
      - name: postgres-config
        configMap:
          name: postgres-config
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 100Gi

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
  namespace: taishanglaojun-prod
data:
  postgresql.conf: |
    # 连接设置
    listen_addresses = '*'
    port = 5432
    max_connections = 200
    
    # 内存设置
    shared_buffers = 1GB
    effective_cache_size = 3GB
    work_mem = 16MB
    maintenance_work_mem = 256MB
    
    # WAL设置
    wal_level = replica
    max_wal_size = 2GB
    min_wal_size = 1GB
    checkpoint_completion_target = 0.9
    
    # 日志设置
    log_destination = 'stderr'
    logging_collector = on
    log_directory = 'log'
    log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
    log_statement = 'all'
    log_min_duration_statement = 1000
```

### 2. Redis集群部署

```yaml
# k8s/redis.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: taishanglaojun-prod
data:
  redis.conf: |
    bind 0.0.0.0
    port 6379
    protected-mode no
    
    # 持久化配置
    save 900 1
    save 300 10
    save 60 10000
    
    # AOF配置
    appendonly yes
    appendfsync everysec
    
    # 内存配置
    maxmemory 2gb
    maxmemory-policy allkeys-lru
    
    # 安全配置
    requirepass ${REDIS_PASSWORD}

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: taishanglaojun-prod
spec:
  serviceName: redis-service
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
          name: redis
        command:
        - redis-server
        - /etc/redis/redis.conf
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: redis-password
        volumeMounts:
        - name: redis-config
          mountPath: /etc/redis
        - name: redis-storage
          mountPath: /data
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: redis-config
        configMap:
          name: redis-config
  volumeClaimTemplates:
  - metadata:
      name: redis-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 20Gi

---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: taishanglaojun-prod
spec:
  ports:
  - port: 6379
    targetPort: 6379
  selector:
    app: redis
```

## 应用服务部署

### 1. 前端服务部署

```yaml
# k8s/frontend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: taishanglaojun-prod
  labels:
    app: frontend
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: frontend
      version: v1
  template:
    metadata:
      labels:
        app: frontend
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "80"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: frontend
        image: taishanglaojun/frontend:v1.0.0
        ports:
        - containerPort: 80
          name: http
        env:
        - name: NODE_ENV
          value: "production"
        - name: API_URL
          value: "https://api.taishanglaojun.com"
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx/nginx.conf
          subPath: nginx.conf
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 101
          capabilities:
            drop:
            - ALL
      volumes:
      - name: nginx-config
        configMap:
          name: nginx-config
      securityContext:
        fsGroup: 101

---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: taishanglaojun-prod
  labels:
    app: frontend
spec:
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: frontend
  type: ClusterIP

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: frontend-hpa
  namespace: taishanglaojun-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: frontend
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### 2. 后端API服务部署

```yaml
# k8s/backend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: taishanglaojun-prod
  labels:
    app: backend
    version: v1
spec:
  replicas: 5
  selector:
    matchLabels:
      app: backend
      version: v1
  template:
    metadata:
      labels:
        app: backend
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: backend-sa
      containers:
      - name: backend
        image: taishanglaojun/backend:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: GIN_MODE
          value: "release"
        - name: DATABASE_URL
          value: "postgres://postgres:$(DATABASE_PASSWORD)@postgres-service:5432/taishanglaojun?sslmode=require"
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-password
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: redis-password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: jwt-secret
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 65534
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config-volume
        configMap:
          name: app-config
      securityContext:
        fsGroup: 65534

---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: taishanglaojun-prod
  labels:
    app: backend
spec:
  ports:
  - port: 80
    targetPort: 8080
    name: http
  selector:
    app: backend
  type: ClusterIP

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: backend-sa
  namespace: taishanglaojun-prod

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: taishanglaojun-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 5
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### 3. AI服务部署

```yaml
# k8s/ai-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-service
  namespace: taishanglaojun-prod
  labels:
    app: ai-service
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ai-service
      version: v1
  template:
    metadata:
      labels:
        app: ai-service
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8000"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: ai-service
        image: taishanglaojun/ai-service:v1.0.0
        ports:
        - containerPort: 8000
          name: http
        env:
        - name: DATABASE_URL
          value: "postgres://postgres:$(DATABASE_PASSWORD)@postgres-service:5432/taishanglaojun?sslmode=require"
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-password
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: redis-password
        - name: QDRANT_URL
          value: "http://qdrant-service:6333"
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: openai-api-key
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: anthropic-api-key
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
        - name: model-cache
          mountPath: /app/models
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
            nvidia.com/gpu: 0
          limits:
            memory: "8Gi"
            cpu: "4000m"
            nvidia.com/gpu: 1
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 60
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 1000
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config-volume
        configMap:
          name: app-config
      - name: model-cache
        emptyDir:
          sizeLimit: 10Gi
      nodeSelector:
        node-type: gpu
      tolerations:
      - key: nvidia.com/gpu
        operator: Exists
        effect: NoSchedule

---
apiVersion: v1
kind: Service
metadata:
  name: ai-service
  namespace: taishanglaojun-prod
  labels:
    app: ai-service
spec:
  ports:
  - port: 80
    targetPort: 8000
    name: http
  selector:
    app: ai-service
  type: ClusterIP

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ai-service-hpa
  namespace: taishanglaojun-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ai-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Ingress配置

### 1. 主要Ingress配置

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: taishanglaojun-ingress
  namespace: taishanglaojun-prod
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "30"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "30"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "30"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://taishanglaojun.com,https://www.taishanglaojun.com"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-headers: "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization"
spec:
  tls:
  - hosts:
    - taishanglaojun.com
    - www.taishanglaojun.com
    - api.taishanglaojun.com
    secretName: taishanglaojun-tls
  rules:
  # 主站点
  - host: taishanglaojun.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
  # www重定向
  - host: www.taishanglaojun.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
  # API服务
  - host: api.taishanglaojun.com
    http:
      paths:
      - path: /api/v1
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 80
      - path: /ai/v1
        pathType: Prefix
        backend:
          service:
            name: ai-service
            port:
              number: 80

---
# WebSocket支持的Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: websocket-ingress
  namespace: taishanglaojun-prod
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    nginx.ingress.kubernetes.io/websocket-services: "backend-service"
spec:
  tls:
  - hosts:
    - ws.taishanglaojun.com
    secretName: websocket-tls
  rules:
  - host: ws.taishanglaojun.com
    http:
      paths:
      - path: /ws
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 80
```

## 存储配置

### 1. StorageClass定义

```yaml
# k8s/storage.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
  encrypted: "true"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
reclaimPolicy: Retain

---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard-hdd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp2
  encrypted: "true"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
reclaimPolicy: Delete

---
# 持久化卷声明模板
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-data-pvc
  namespace: taishanglaojun-prod
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 100Gi
```

### 2. 备份存储配置

```yaml
# k8s/backup-storage.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: backup-pvc
  namespace: taishanglaojun-prod
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: standard-hdd
  resources:
    requests:
      storage: 1Ti

---
# 备份任务
apiVersion: batch/v1
kind: CronJob
metadata:
  name: database-backup
  namespace: taishanglaojun-prod
spec:
  schedule: "0 2 * * *"  # 每天凌晨2点
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15-alpine
            command:
            - /bin/bash
            - -c
            - |
              BACKUP_FILE="/backup/postgres-$(date +%Y%m%d_%H%M%S).sql"
              pg_dump -h postgres-service -U postgres -d taishanglaojun > $BACKUP_FILE
              gzip $BACKUP_FILE
              # 删除7天前的备份
              find /backup -name "*.sql.gz" -mtime +7 -delete
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: database-password
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

## 监控和日志

### 1. Prometheus监控

```yaml
# k8s/monitoring.yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: taishanglaojun-monitor
  namespace: taishanglaojun-prod
  labels:
    app: taishanglaojun
spec:
  selector:
    matchLabels:
      app: backend
  endpoints:
  - port: http
    path: /metrics
    interval: 30s

---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: taishanglaojun-rules
  namespace: taishanglaojun-prod
spec:
  groups:
  - name: taishanglaojun.rules
    rules:
    - alert: HighCPUUsage
      expr: rate(container_cpu_usage_seconds_total[5m]) > 0.8
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High CPU usage detected"
        description: "CPU usage is above 80% for more than 5 minutes"
    
    - alert: HighMemoryUsage
      expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High memory usage detected"
        description: "Memory usage is above 90% for more than 5 minutes"
    
    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Pod is crash looping"
        description: "Pod {{ $labels.pod }} is restarting frequently"
```

### 2. 日志收集

```yaml
# k8s/logging.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: kube-system
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020

    [INPUT]
        Name              tail
        Path              /var/log/containers/*.log
        Parser            docker
        Tag               kube.*
        Refresh_Interval  5
        Mem_Buf_Limit     50MB
        Skip_Long_Lines   On

    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Merge_Log           On
        K8S-Logging.Parser  On
        K8S-Logging.Exclude Off

    [OUTPUT]
        Name  es
        Match *
        Host  elasticsearch.logging.svc.cluster.local
        Port  9200
        Index taishanglaojun-logs
        Type  _doc

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: fluent-bit
  template:
    metadata:
      labels:
        name: fluent-bit
    spec:
      serviceAccount: fluent-bit
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      containers:
      - name: fluent-bit
        image: fluent/fluent-bit:1.9
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: fluent-bit-config
          mountPath: /fluent-bit/etc/
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: fluent-bit-config
        configMap:
          name: fluent-bit-config
```

## 部署脚本和自动化

### 1. Helm部署脚本

```bash
#!/bin/bash
# scripts/k8s-deploy.sh

set -e

NAMESPACE=${1:-taishanglaojun-prod}
ENVIRONMENT=${2:-production}
VERSION=${3:-latest}

echo "🚀 开始部署到Kubernetes集群..."
echo "📋 部署参数:"
echo "  - 命名空间: $NAMESPACE"
echo "  - 环境: $ENVIRONMENT"
echo "  - 版本: $VERSION"

# 检查kubectl连接
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ 无法连接到Kubernetes集群"
    exit 1
fi

# 检查Helm
if ! command -v helm &> /dev/null; then
    echo "❌ Helm未安装"
    exit 1
fi

# 创建命名空间
echo "📁 创建命名空间..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 应用基础资源
echo "🔧 应用基础配置..."
kubectl apply -f k8s/namespaces.yaml
kubectl apply -f k8s/configmaps.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/storage.yaml

# 部署数据库
echo "🗄️ 部署数据库..."
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml

# 等待数据库就绪
echo "⏳ 等待数据库就绪..."
kubectl wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=300s

# 部署应用服务
echo "🚀 部署应用服务..."
kubectl apply -f k8s/frontend.yaml
kubectl apply -f k8s/backend.yaml
kubectl apply -f k8s/ai-service.yaml

# 等待应用就绪
echo "⏳ 等待应用就绪..."
kubectl wait --for=condition=available deployment/frontend -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=available deployment/backend -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=available deployment/ai-service -n $NAMESPACE --timeout=300s

# 应用Ingress
echo "🌐 配置Ingress..."
kubectl apply -f k8s/ingress.yaml

# 应用监控配置
echo "📊 配置监控..."
kubectl apply -f k8s/monitoring.yaml

# 检查部署状态
echo "🔍 检查部署状态..."
kubectl get pods -n $NAMESPACE
kubectl get services -n $NAMESPACE
kubectl get ingress -n $NAMESPACE

echo "✅ 部署完成！"
echo "🌐 访问地址: https://taishanglaojun.com"
```

### 2. 滚动更新脚本

```bash
#!/bin/bash
# scripts/rolling-update.sh

set -e

SERVICE=$1
VERSION=$2
NAMESPACE=${3:-taishanglaojun-prod}

if [ -z "$SERVICE" ] || [ -z "$VERSION" ]; then
    echo "用法: $0 <service> <version> [namespace]"
    echo "示例: $0 backend v1.2.0 taishanglaojun-prod"
    exit 1
fi

echo "🔄 开始滚动更新 $SERVICE 到版本 $VERSION..."

# 更新镜像
kubectl set image deployment/$SERVICE $SERVICE=taishanglaojun/$SERVICE:$VERSION -n $NAMESPACE

# 等待滚动更新完成
echo "⏳ 等待滚动更新完成..."
kubectl rollout status deployment/$SERVICE -n $NAMESPACE --timeout=600s

# 检查新版本状态
echo "🔍 检查新版本状态..."
kubectl get pods -l app=$SERVICE -n $NAMESPACE

# 验证服务健康状态
echo "🏥 验证服务健康状态..."
sleep 30

if kubectl get pods -l app=$SERVICE -n $NAMESPACE | grep -q "Running"; then
    echo "✅ 滚动更新成功完成！"
else
    echo "❌ 滚动更新失败，开始回滚..."
    kubectl rollout undo deployment/$SERVICE -n $NAMESPACE
    exit 1
fi
```

## 相关文档链接

- [部署概览](./deployment-overview.md)
- [Docker部署指南](./docker-deployment.md)
- [监控运维指南](./monitoring-operations.md)
- [安全配置指南](./security-configuration.md)
- [性能优化指南](./performance-optimization.md)
- [故障排除指南](./troubleshooting.md)