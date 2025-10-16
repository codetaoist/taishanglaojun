# 太上老君AI平台 - 容器化服务

## 概述

太上老君AI平台采用云原生容器化架构，基于S×C×T三轴理论设计，提供高可用、可扩展、易维护的微服务架构。

## 容器化架构

```mermaid
graph TB
    subgraph "容器编排层"
        K8S[Kubernetes集群]
        Helm[Helm包管理]
        Operator[自定义Operator]
    end
    
    subgraph "应用容器层 (S1-S3)"
        subgraph "前端服务"
            WebApp[Web应用容器]
            MobileAPI[移动端API容器]
            AdminPanel[管理面板容器]
        end
        
        subgraph "核心服务"
            AIService[AI服务容器]
            UserService[用户服务容器]
            DataService[数据服务容器]
            AuthService[认证服务容器]
        end
        
        subgraph "智能服务 (S3-S5)"
            MLPipeline[机器学习流水线]
            KnowledgeGraph[知识图谱服务]
            RecommendEngine[推荐引擎]
        end
    end
    
    subgraph "基础设施容器层 (S0-S2)"
        subgraph "数据存储"
            PostgreSQL[PostgreSQL容器]
            Redis[Redis容器]
            MongoDB[MongoDB容器]
            Neo4j[Neo4j容器]
        end
        
        subgraph "消息队列"
            Kafka[Kafka容器]
            RabbitMQ[RabbitMQ容器]
        end
        
        subgraph "监控日志"
            Prometheus[Prometheus容器]
            Grafana[Grafana容器]
            ELK[ELK Stack容器]
        end
    end
    
    subgraph "网络层"
        Ingress[Ingress Controller]
        ServiceMesh[Istio Service Mesh]
        LoadBalancer[负载均衡器]
    end
```

## 核心服务容器

### 1. AI服务容器

```dockerfile
# Dockerfile.ai-service
FROM python:3.11-slim as builder

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    make \
    libffi-dev \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# 生产阶段
FROM python:3.11-slim

WORKDIR /app

# 创建非root用户
RUN groupadd -r aiuser && useradd -r -g aiuser aiuser

# 复制Python依赖
COPY --from=builder /root/.local /home/aiuser/.local

# 复制应用代码
COPY --chown=aiuser:aiuser . .

# 设置环境变量
ENV PATH=/home/aiuser/.local/bin:$PATH
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8080/health')"

# 切换到非root用户
USER aiuser

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
```

### 2. 用户服务容器

```dockerfile
# Dockerfile.user-service
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/user-service

# 生产阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 创建非root用户
RUN addgroup -g 1001 appuser && \
    adduser -D -s /bin/sh -u 1001 -G appuser appuser

# 切换到非root用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./main"]
```

### 3. 数据服务容器

```dockerfile
# Dockerfile.data-service
FROM node:18-alpine AS builder

WORKDIR /app

# 复制package文件
COPY package*.json ./
RUN npm ci --only=production

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 生产阶段
FROM node:18-alpine

WORKDIR /app

# 安装dumb-init
RUN apk add --no-cache dumb-init

# 创建非root用户
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# 复制构建产物
COPY --from=builder --chown=nextjs:nodejs /app/dist ./dist
COPY --from=builder --chown=nextjs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nextjs:nodejs /app/package.json ./package.json

# 切换到非root用户
USER nextjs

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD node healthcheck.js

# 暴露端口
EXPOSE 3000

# 启动命令
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "dist/index.js"]
```

## Kubernetes部署配置

### 1. AI服务部署

```yaml
# k8s/ai-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-service
  namespace: taishang-ai
  labels:
    app: ai-service
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ai-service
  template:
    metadata:
      labels:
        app: ai-service
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: ai-service
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: ai-service
        image: taishang/ai-service:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: ai-service-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: ai-service-secrets
              key: redis-url
        - name: LOG_LEVEL
          value: "info"
        - name: ENVIRONMENT
          value: "production"
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
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
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: models
          mountPath: /app/models
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: ai-service-config
      - name: models
        persistentVolumeClaim:
          claimName: ai-models-pvc
      nodeSelector:
        node-type: compute-optimized
      tolerations:
      - key: "ai-workload"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"

---
apiVersion: v1
kind: Service
metadata:
  name: ai-service
  namespace: taishang-ai
  labels:
    app: ai-service
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: ai-service

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ai-service
  namespace: taishang-ai
  labels:
    app: ai-service
```

### 2. 用户服务部署

```yaml
# k8s/user-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: taishang-core
  labels:
    app: user-service
    version: v1.0.0
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: user-service
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: user-service
        image: taishang/user-service:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        - containerPort: 9090
          name: grpc
          protocol: TCP
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: user-service-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: user-service-secrets
              key: jwt-secret
        - name: REDIS_URL
          valueFrom:
            configMapKeyRef:
              name: user-service-config
              key: redis-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "250m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /root/configs
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: user-service-config
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - user-service
              topologyKey: kubernetes.io/hostname

---
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: taishang-core
  labels:
    app: user-service
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  - port: 9090
    targetPort: 9090
    protocol: TCP
    name: grpc
  selector:
    app: user-service
```

## Docker Compose配置

### 1. 开发环境配置

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  # 前端服务
  web-app:
    build:
      context: ./frontend/web-app
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    volumes:
      - ./frontend/web-app:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
      - REACT_APP_API_URL=http://localhost:8080
    depends_on:
      - api-gateway

  # API网关
  api-gateway:
    build:
      context: ./core-services
      dockerfile: Dockerfile.gateway
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=development
      - LOG_LEVEL=debug
    volumes:
      - ./core-services/configs:/app/configs
    depends_on:
      - user-service
      - ai-service
      - data-service

  # 用户服务
  user-service:
    build:
      context: ./core-services
      dockerfile: Dockerfile.user-service
    ports:
      - "8081:8080"
    environment:
      - DATABASE_URL=postgres://user:password@postgres:5432/taishang_users
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=dev-secret-key
    depends_on:
      - postgres
      - redis

  # AI服务
  ai-service:
    build:
      context: ./core-services
      dockerfile: Dockerfile.ai-service
    ports:
      - "8082:8080"
    environment:
      - DATABASE_URL=postgres://user:password@postgres:5432/taishang_ai
      - REDIS_URL=redis://redis:6379
      - MODEL_PATH=/app/models
    volumes:
      - ./models:/app/models
    depends_on:
      - postgres
      - redis
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

  # 数据服务
  data-service:
    build:
      context: ./core-services
      dockerfile: Dockerfile.data-service
    ports:
      - "8083:3000"
    environment:
      - DATABASE_URL=postgres://user:password@postgres:5432/taishang_data
      - MONGODB_URL=mongodb://mongo:27017/taishang
      - NEO4J_URL=bolt://neo4j:7687
    depends_on:
      - postgres
      - mongo
      - neo4j

  # 数据库服务
  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=taishang
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

  mongo:
    image: mongo:6
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
    volumes:
      - mongo_data:/data/db

  neo4j:
    image: neo4j:5
    ports:
      - "7474:7474"
      - "7687:7687"
    environment:
      - NEO4J_AUTH=neo4j/password
      - NEO4J_PLUGINS=["apoc"]
    volumes:
      - neo4j_data:/data

  # 监控服务
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources

volumes:
  postgres_data:
  redis_data:
  mongo_data:
  neo4j_data:
  prometheus_data:
  grafana_data:

networks:
  default:
    name: taishang-dev
```

### 2. 生产环境配置

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  # 前端服务
  web-app:
    image: taishang/web-app:${VERSION:-latest}
    restart: unless-stopped
    environment:
      - NODE_ENV=production
      - REACT_APP_API_URL=https://api.taishang.ai
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webapp.rule=Host(`app.taishang.ai`)"
      - "traefik.http.routers.webapp.tls=true"
      - "traefik.http.routers.webapp.tls.certresolver=letsencrypt"
    networks:
      - frontend
      - backend

  # API网关
  api-gateway:
    image: taishang/api-gateway:${VERSION:-latest}
    restart: unless-stopped
    environment:
      - ENVIRONMENT=production
      - LOG_LEVEL=info
      - DATABASE_URL_FILE=/run/secrets/database_url
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
    secrets:
      - database_url
      - jwt_secret
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`api.taishang.ai`)"
      - "traefik.http.routers.api.tls=true"
      - "traefik.http.routers.api.tls.certresolver=letsencrypt"
    networks:
      - frontend
      - backend
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3

  # 用户服务
  user-service:
    image: taishang/user-service:${VERSION:-latest}
    restart: unless-stopped
    environment:
      - ENVIRONMENT=production
      - DATABASE_URL_FILE=/run/secrets/database_url
      - REDIS_URL_FILE=/run/secrets/redis_url
    secrets:
      - database_url
      - redis_url
    networks:
      - backend
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M

  # AI服务
  ai-service:
    image: taishang/ai-service:${VERSION:-latest}
    restart: unless-stopped
    environment:
      - ENVIRONMENT=production
      - MODEL_PATH=/app/models
    volumes:
      - ai_models:/app/models:ro
    networks:
      - backend
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

  # 负载均衡器
  traefik:
    image: traefik:v3.0
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    command:
      - --api.dashboard=true
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.email=admin@taishang.ai
      - --certificatesresolvers.letsencrypt.acme.storage=/acme.json
      - --certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_acme:/acme.json
    networks:
      - frontend
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.dashboard.rule=Host(`traefik.taishang.ai`)"
      - "traefik.http.routers.dashboard.tls=true"

secrets:
  database_url:
    external: true
  redis_url:
    external: true
  jwt_secret:
    external: true

volumes:
  ai_models:
    external: true
  traefik_acme:

networks:
  frontend:
    external: true
  backend:
    external: true
```

## Helm Charts配置

### 1. Chart.yaml

```yaml
# helm/taishang-platform/Chart.yaml
apiVersion: v2
name: taishang-platform
description: 太上老君AI平台Helm Chart
type: application
version: 1.0.0
appVersion: "1.0.0"
keywords:
  - ai
  - platform
  - microservices
home: https://taishang.ai
sources:
  - https://github.com/taishang-ai/platform
maintainers:
  - name: Taishang Team
    email: dev@taishang.ai
dependencies:
  - name: postgresql
    version: "12.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
  - name: redis
    version: "18.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
  - name: mongodb
    version: "13.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: mongodb.enabled
```

### 2. Values.yaml

```yaml
# helm/taishang-platform/values.yaml
global:
  imageRegistry: ""
  imagePullSecrets: []
  storageClass: ""

replicaCount: 3

image:
  registry: docker.io
  repository: taishang/platform
  tag: "1.0.0"
  pullPolicy: IfNotPresent

nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"

podSecurityContext:
  fsGroup: 1001
  runAsNonRoot: true
  runAsUser: 1001

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1001

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: api.taishang.ai
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: taishang-tls
      hosts:
        - api.taishang.ai

resources:
  limits:
    cpu: 500m
    memory: 1Gi
  requests:
    cpu: 250m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - taishang-platform
        topologyKey: kubernetes.io/hostname

# 数据库配置
postgresql:
  enabled: true
  auth:
    postgresPassword: "changeme"
    username: "taishang"
    password: "changeme"
    database: "taishang"
  primary:
    persistence:
      enabled: true
      size: 20Gi

redis:
  enabled: true
  auth:
    enabled: true
    password: "changeme"
  master:
    persistence:
      enabled: true
      size: 8Gi

mongodb:
  enabled: true
  auth:
    enabled: true
    rootPassword: "changeme"
    username: "taishang"
    password: "changeme"
    database: "taishang"
  persistence:
    enabled: true
    size: 20Gi

# 监控配置
monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
    adminPassword: "changeme"

# AI服务配置
aiService:
  enabled: true
  replicaCount: 2
  image:
    repository: taishang/ai-service
    tag: "1.0.0"
  resources:
    limits:
      cpu: 1000m
      memory: 2Gi
      nvidia.com/gpu: 1
    requests:
      cpu: 500m
      memory: 1Gi
  persistence:
    enabled: true
    size: 50Gi
    storageClass: "fast-ssd"

# 用户服务配置
userService:
  enabled: true
  replicaCount: 2
  image:
    repository: taishang/user-service
    tag: "1.0.0"
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 256Mi
```

### 3. 部署模板

```yaml
# helm/taishang-platform/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "taishang-platform.fullname" . }}
  labels:
    {{- include "taishang-platform.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "taishang-platform.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "taishang-platform.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.global.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "taishang-platform.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.registry }}/{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.targetPort }}
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: {{ include "taishang-platform.fullname" . }}-secrets
                  key: database-url
            - name: REDIS_URL
              valueFrom:
                secretKeyRef:
                  name: {{ include "taishang-platform.fullname" . }}-secrets
                  key: redis-url
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

## 容器安全配置

### 1. 安全策略

```yaml
# security/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: taishang-restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: taishang-network-policy
  namespace: taishang-production
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: taishang-system
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: taishang-database
    ports:
    - protocol: TCP
      port: 5432
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

### 2. 镜像扫描配置

```yaml
# security/image-scan-policy.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: image-scan-config
  namespace: taishang-security
data:
  policy.yaml: |
    policies:
      - name: "vulnerability-scan"
        rules:
          - selector:
              matchLabels:
                scan: "required"
            requirements:
              - key: "vulnerability-scan"
                operator: "Exists"
                values: ["passed"]
          - selector:
              matchLabels:
                environment: "production"
            requirements:
              - key: "security-scan"
                operator: "In"
                values: ["high", "critical"]
                
      - name: "base-image-policy"
        rules:
          - selector: {}
            requirements:
              - key: "base-image"
                operator: "In"
                values: ["alpine", "distroless", "scratch"]
                
    scanners:
      - name: "trivy"
        endpoint: "http://trivy-server:4954"
        timeout: "5m"
      - name: "clair"
        endpoint: "http://clair:6060"
        timeout: "3m"
```

## 性能优化

### 1. 资源限制和请求

```yaml
# performance/resource-optimization.yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: taishang-limits
  namespace: taishang-production
spec:
  limits:
  - default:
      cpu: "500m"
      memory: "1Gi"
    defaultRequest:
      cpu: "100m"
      memory: "256Mi"
    max:
      cpu: "2"
      memory: "4Gi"
    min:
      cpu: "50m"
      memory: "128Mi"
    type: Container
  - max:
      storage: "100Gi"
    min:
      storage: "1Gi"
    type: PersistentVolumeClaim

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: taishang-hpa
  namespace: taishang-production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: taishang-platform
  minReplicas: 3
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

### 2. 缓存优化

```yaml
# performance/cache-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cache-config
  namespace: taishang-production
data:
  redis.conf: |
    # Redis缓存优化配置
    maxmemory 2gb
    maxmemory-policy allkeys-lru
    save 900 1
    save 300 10
    save 60 10000
    
    # 网络优化
    tcp-keepalive 300
    timeout 0
    
    # 性能优化
    hash-max-ziplist-entries 512
    hash-max-ziplist-value 64
    list-max-ziplist-size -2
    set-max-intset-entries 512
    zset-max-ziplist-entries 128
    zset-max-ziplist-value 64
    
  nginx.conf: |
    # Nginx缓存配置
    proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m max_size=1g inactive=60m use_temp_path=off;
    
    server {
        listen 80;
        server_name api.taishang.ai;
        
        location /api/v1/static/ {
            proxy_cache api_cache;
            proxy_cache_valid 200 1h;
            proxy_cache_valid 404 1m;
            proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
            proxy_cache_lock on;
            
            add_header X-Cache-Status $upstream_cache_status;
            proxy_pass http://backend;
        }
        
        location / {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
```

## 相关文档

- [云基础设施](./cloud-infrastructure.md)
- [数据库管理](./database-management.md)
- [监控告警](../08-部署运维/monitoring-alerting.md)
- [安全配置](../08-部署运维/security-config.md)