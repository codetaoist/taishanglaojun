# 太上老君AI平台 - 部署运维概览

## 概述

太上老君AI平台采用云原生架构，支持多种部署方式，包括本地开发、测试环境、预生产环境和生产环境。本文档提供部署运维的全面指南。

## 部署架构

```mermaid
graph TB
    subgraph "负载均衡层"
        A[Nginx/HAProxy] --> B[SSL终端]
        B --> C[负载均衡器]
    end
    
    subgraph "应用层"
        C --> D[API网关集群]
        D --> E[用户服务集群]
        D --> F[AI服务集群]
        D --> G[数据服务集群]
        D --> H[系统服务集群]
    end
    
    subgraph "数据层"
        E --> I[(PostgreSQL主从)]
        F --> I
        G --> I
        H --> I
        
        E --> J[(Redis集群)]
        F --> J
        G --> J
        
        F --> K[(Elasticsearch集群)]
        G --> K
        
        F --> L[(MinIO集群)]
        G --> L
        
        F --> M[(Qdrant集群)]
    end
    
    subgraph "监控层"
        N[Prometheus] --> O[Grafana]
        P[Jaeger] --> O
        Q[ELK Stack] --> O
        R[AlertManager] --> S[通知系统]
    end
    
    subgraph "基础设施"
        T[Kubernetes集群]
        U[Docker Registry]
        V[CI/CD Pipeline]
        W[配置管理]
    end
```

## 部署策略

### 1. 环境分层

#### 开发环境 (Development)
- **目的**: 本地开发和调试
- **特点**: 单机部署，快速启动
- **工具**: Docker Compose
- **数据**: 测试数据，可重置

#### 测试环境 (Testing)
- **目的**: 功能测试和集成测试
- **特点**: 模拟生产环境，自动化测试
- **工具**: Kubernetes + Helm
- **数据**: 测试数据集，定期刷新

#### 预生产环境 (Staging)
- **目的**: 生产前验证
- **特点**: 与生产环境一致的配置
- **工具**: Kubernetes + Helm
- **数据**: 生产数据副本（脱敏）

#### 生产环境 (Production)
- **目的**: 正式服务用户
- **特点**: 高可用、高性能、安全
- **工具**: Kubernetes + Helm + GitOps
- **数据**: 真实业务数据

### 2. 部署模式

#### 蓝绿部署 (Blue-Green Deployment)
```yaml
# 蓝绿部署配置示例
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: taishanglaojun-api
spec:
  replicas: 5
  strategy:
    blueGreen:
      activeService: api-active
      previewService: api-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: api-preview
      postPromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: api-active
```

#### 金丝雀部署 (Canary Deployment)
```yaml
# 金丝雀部署配置
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: taishanglaojun-ai
spec:
  replicas: 10
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 10m}
      - setWeight: 20
      - pause: {duration: 10m}
      - setWeight: 50
      - pause: {duration: 10m}
      - setWeight: 100
      canaryService: ai-canary
      stableService: ai-stable
      trafficRouting:
        istio:
          virtualService:
            name: ai-vsvc
```

#### 滚动更新 (Rolling Update)
```yaml
# 滚动更新配置
apiVersion: apps/v1
kind: Deployment
metadata:
  name: taishanglaojun-user
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  template:
    spec:
      containers:
      - name: user-service
        image: taishanglaojun/user-service:v1.2.0
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
```

## 容器化部署

### 1. Docker镜像构建

#### 多阶段构建 - 前端
```dockerfile
# frontend/Dockerfile
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### 多阶段构建 - 后端
```dockerfile
# backend/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]
```

#### 多阶段构建 - AI服务
```dockerfile
# ai-service/Dockerfile
FROM python:3.11-slim AS builder

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

FROM python:3.11-slim
WORKDIR /app

COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=builder /app .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### 2. Docker Compose 开发环境

```yaml
# docker-compose.yml
version: '3.8'

services:
  # 前端服务
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      - REACT_APP_API_URL=http://localhost:8080
    depends_on:
      - backend

  # 后端API网关
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/taishanglaojun?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=dev-secret-key
    depends_on:
      - postgres
      - redis
      - elasticsearch

  # AI服务
  ai-service:
    build:
      context: ./ai-service
      dockerfile: Dockerfile.dev
    ports:
      - "8000:8000"
    volumes:
      - ./ai-service:/app
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/taishanglaojun?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      - postgres
      - redis
      - qdrant

  # PostgreSQL数据库
  postgres:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=taishanglaojun
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql

  # Redis缓存
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # Elasticsearch搜索引擎
  elasticsearch:
    image: elasticsearch:8.11.0
    ports:
      - "9200:9200"
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  # Qdrant向量数据库
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
    volumes:
      - qdrant_data:/qdrant/storage

  # MinIO对象存储
  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"

  # Prometheus监控
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus

  # Grafana可视化
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
  elasticsearch_data:
  qdrant_data:
  minio_data:
  prometheus_data:
  grafana_data:
```

## Kubernetes部署

### 1. 命名空间和资源配额

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: taishanglaojun
  labels:
    name: taishanglaojun
    environment: production

---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: taishanglaojun
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    persistentvolumeclaims: "10"
    services: "10"
    secrets: "10"
    configmaps: "10"

---
apiVersion: v1
kind: LimitRange
metadata:
  name: mem-limit-range
  namespace: taishanglaojun
spec:
  limits:
  - default:
      memory: "512Mi"
      cpu: "500m"
    defaultRequest:
      memory: "256Mi"
      cpu: "100m"
    type: Container
```

### 2. 配置管理

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: taishanglaojun
data:
  database.yaml: |
    host: postgres-service
    port: 5432
    database: taishanglaojun
    sslmode: require
    max_connections: 100
    
  redis.yaml: |
    host: redis-service
    port: 6379
    db: 0
    max_connections: 100
    
  ai.yaml: |
    providers:
      openai:
        base_url: https://api.openai.com/v1
        timeout: 30s
      anthropic:
        base_url: https://api.anthropic.com
        timeout: 30s

---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: taishanglaojun
type: Opaque
data:
  database-password: cGFzc3dvcmQ=  # base64 encoded
  jwt-secret: c2VjcmV0LWtleQ==
  openai-api-key: c2stZXhhbXBsZQ==
  anthropic-api-key: YW50aHJvcGljLWtleQ==
```

### 3. 数据库部署

```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: taishanglaojun
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
      containers:
      - name: postgres
        image: postgres:15
        ports:
        - containerPort: 5432
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
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
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
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: taishanglaojun
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
  type: ClusterIP
```

### 4. 应用服务部署

```yaml
# k8s/api-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: taishanglaojun
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: taishanglaojun/api-gateway:v1.0.0
        ports:
        - containerPort: 8080
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
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
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
      volumes:
      - name: config-volume
        configMap:
          name: app-config

---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-service
  namespace: taishanglaojun
spec:
  selector:
    app: api-gateway
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway-ingress
  namespace: taishanglaojun
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - api.taishanglaojun.com
    secretName: api-tls
  rules:
  - host: api.taishanglaojun.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway-service
            port:
              number: 80
```

## Helm Charts

### 1. Chart结构

```
charts/taishanglaojun/
├── Chart.yaml
├── values.yaml
├── values-dev.yaml
├── values-staging.yaml
├── values-prod.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── hpa.yaml
│   └── pdb.yaml
└── charts/
    ├── postgresql/
    ├── redis/
    └── elasticsearch/
```

### 2. Chart.yaml

```yaml
# charts/taishanglaojun/Chart.yaml
apiVersion: v2
name: taishanglaojun
description: 太上老君AI平台 Helm Chart
type: application
version: 1.0.0
appVersion: "1.0.0"

dependencies:
- name: postgresql
  version: 12.1.9
  repository: https://charts.bitnami.com/bitnami
  condition: postgresql.enabled

- name: redis
  version: 17.3.7
  repository: https://charts.bitnami.com/bitnami
  condition: redis.enabled

- name: elasticsearch
  version: 19.5.0
  repository: https://charts.bitnami.com/bitnami
  condition: elasticsearch.enabled

maintainers:
- name: TaiShangLaoJun Team
  email: team@taishanglaojun.com
```

### 3. values.yaml

```yaml
# charts/taishanglaojun/values.yaml
global:
  imageRegistry: ""
  imagePullSecrets: []
  storageClass: ""

replicaCount: 3

image:
  registry: docker.io
  repository: taishanglaojun/api-gateway
  tag: "v1.0.0"
  pullPolicy: IfNotPresent

nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations: {}

podSecurityContext:
  fsGroup: 2000

securityContext:
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: api.taishanglaojun.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: api-tls
      hosts:
        - api.taishanglaojun.com

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}

# 数据库配置
postgresql:
  enabled: true
  auth:
    postgresPassword: "password"
    database: "taishanglaojun"
  primary:
    persistence:
      enabled: true
      size: 20Gi

# Redis配置
redis:
  enabled: true
  auth:
    enabled: false
  master:
    persistence:
      enabled: true
      size: 8Gi

# Elasticsearch配置
elasticsearch:
  enabled: true
  master:
    replicaCount: 1
  data:
    replicaCount: 1
  coordinating:
    replicaCount: 1
```

### 4. 部署模板

```yaml
# charts/taishanglaojun/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "taishanglaojun.fullname" . }}
  labels:
    {{- include "taishanglaojun.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "taishanglaojun.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "taishanglaojun.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "taishanglaojun.serviceAccountName" . }}
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
              value: "postgres://postgres:{{ .Values.postgresql.auth.postgresPassword }}@{{ include "taishanglaojun.fullname" . }}-postgresql:5432/{{ .Values.postgresql.auth.database }}?sslmode=require"
            - name: REDIS_URL
              value: "redis://{{ include "taishanglaojun.fullname" . }}-redis-master:6379"
            - name: ELASTICSEARCH_URL
              value: "http://{{ include "taishanglaojun.fullname" . }}-elasticsearch:9200"
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

## CI/CD流水线

### 1. GitLab CI配置

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy-dev
  - deploy-staging
  - deploy-prod

variables:
  DOCKER_REGISTRY: registry.gitlab.com/taishanglaojun
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

# 测试阶段
test:frontend:
  stage: test
  image: node:18-alpine
  script:
    - cd frontend
    - npm ci
    - npm run lint
    - npm run test:coverage
  coverage: '/Lines\s*:\s*(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: frontend/coverage/cobertura-coverage.xml

test:backend:
  stage: test
  image: golang:1.21-alpine
  script:
    - cd backend
    - go mod download
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  coverage: '/total:\s*\(statements\)\s*(\d+\.\d+)%/'

test:ai-service:
  stage: test
  image: python:3.11-slim
  script:
    - cd ai-service
    - pip install -r requirements-dev.txt
    - pytest --cov=. --cov-report=xml
  coverage: '/TOTAL.*\s+(\d+%)$/'

# 构建阶段
build:frontend:
  stage: build
  image: docker:20.10.16
  services:
    - docker:20.10.16-dind
  script:
    - cd frontend
    - docker build -t $DOCKER_REGISTRY/frontend:$CI_COMMIT_SHA .
    - docker push $DOCKER_REGISTRY/frontend:$CI_COMMIT_SHA
  only:
    - main
    - develop

build:backend:
  stage: build
  image: docker:20.10.16
  services:
    - docker:20.10.16-dind
  script:
    - cd backend
    - docker build -t $DOCKER_REGISTRY/backend:$CI_COMMIT_SHA .
    - docker push $DOCKER_REGISTRY/backend:$CI_COMMIT_SHA
  only:
    - main
    - develop

build:ai-service:
  stage: build
  image: docker:20.10.16
  services:
    - docker:20.10.16-dind
  script:
    - cd ai-service
    - docker build -t $DOCKER_REGISTRY/ai-service:$CI_COMMIT_SHA .
    - docker push $DOCKER_REGISTRY/ai-service:$CI_COMMIT_SHA
  only:
    - main
    - develop

# 开发环境部署
deploy:dev:
  stage: deploy-dev
  image: alpine/helm:latest
  script:
    - helm upgrade --install taishanglaojun-dev ./charts/taishanglaojun 
      --namespace taishanglaojun-dev 
      --create-namespace
      --values ./charts/taishanglaojun/values-dev.yaml
      --set image.tag=$CI_COMMIT_SHA
  environment:
    name: development
    url: https://dev.taishanglaojun.com
  only:
    - develop

# 预生产环境部署
deploy:staging:
  stage: deploy-staging
  image: alpine/helm:latest
  script:
    - helm upgrade --install taishanglaojun-staging ./charts/taishanglaojun 
      --namespace taishanglaojun-staging 
      --create-namespace
      --values ./charts/taishanglaojun/values-staging.yaml
      --set image.tag=$CI_COMMIT_SHA
  environment:
    name: staging
    url: https://staging.taishanglaojun.com
  when: manual
  only:
    - main

# 生产环境部署
deploy:prod:
  stage: deploy-prod
  image: alpine/helm:latest
  script:
    - helm upgrade --install taishanglaojun ./charts/taishanglaojun 
      --namespace taishanglaojun 
      --create-namespace
      --values ./charts/taishanglaojun/values-prod.yaml
      --set image.tag=$CI_COMMIT_SHA
  environment:
    name: production
    url: https://taishanglaojun.com
  when: manual
  only:
    - main
```

### 2. GitHub Actions配置

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [frontend, backend, ai-service]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Test Frontend
      if: matrix.service == 'frontend'
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    - run: |
        cd frontend
        npm ci
        npm run lint
        npm run test:coverage
    
    - name: Test Backend
      if: matrix.service == 'backend'
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    - run: |
        cd backend
        go mod download
        go test -v -race -coverprofile=coverage.out ./...
    
    - name: Test AI Service
      if: matrix.service == 'ai-service'
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    - run: |
        cd ai-service
        pip install -r requirements-dev.txt
        pytest --cov=. --cov-report=xml

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/develop'
    
    strategy:
      matrix:
        service: [frontend, backend, ai-service]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/${{ matrix.service }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: ./${{ matrix.service }}
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}

  deploy-dev:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: development
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to Development
      run: |
        # 这里添加部署到开发环境的脚本
        echo "Deploying to development environment"

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: staging
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to Staging
      run: |
        # 这里添加部署到预生产环境的脚本
        echo "Deploying to staging environment"

  deploy-prod:
    needs: deploy-staging
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to Production
      run: |
        # 这里添加部署到生产环境的脚本
        echo "Deploying to production environment"
```

## 相关文档链接

- [Docker部署指南](./docker-deployment.md)
- [Kubernetes部署指南](./kubernetes-deployment.md)
- [监控运维指南](./monitoring-operations.md)
- [安全配置指南](./security-configuration.md)
- [性能优化指南](./performance-optimization.md)
- [故障排除指南](./troubleshooting.md)