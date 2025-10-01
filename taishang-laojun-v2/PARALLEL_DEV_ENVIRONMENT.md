# 并行开发环境配置指南

## 🎯 环境目标

建立支持多窗口并行开发的完整工作环境，确保各模块能够独立开发、测试和部署。

## 🏗️ 环境架构设计

### 1. 开发环境分层

```yaml
环境分层架构:
  本地开发环境:
    - 各窗口独立的开发环境
    - 热重载和快速反馈
    - 本地数据库和服务

  集成测试环境:
    - 模块间集成测试
    - 接口兼容性验证
    - 端到端功能测试

  预生产环境:
    - 生产环境模拟
    - 性能和压力测试
    - 部署流程验证

  生产环境:
    - 容器化部署
    - 监控和告警
    - 自动化运维
```

### 2. 容器化环境设计

#### Docker Compose 配置
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  # 数据库服务
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: taishang_dev
      POSTGRES_USER: dev_user
      POSTGRES_PASSWORD: dev_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  mongodb:
    image: mongo:6
    environment:
      MONGO_INITDB_ROOT_USERNAME: dev_user
      MONGO_INITDB_ROOT_PASSWORD: dev_pass
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
    volumes:
      - qdrant_data:/qdrant/storage

  # 后端服务
  auth-service:
    build:
      context: ./01-infrastructure/auth-system
      dockerfile: Dockerfile.dev
    ports:
      - "8081:8081"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - ./01-infrastructure/auth-system:/app
    command: air -c .air.toml

  api-gateway:
    build:
      context: ./01-infrastructure/api-gateway
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8081
      - CULTURAL_SERVICE_URL=http://cultural-service:8082
    depends_on:
      - auth-service
    volumes:
      - ./01-infrastructure/api-gateway:/app
    command: air -c .air.toml

  cultural-service:
    build:
      context: ./02-core-services/cultural-wisdom
      dockerfile: Dockerfile.dev
    ports:
      - "8082:8082"
    environment:
      - DB_HOST=postgres
      - MONGO_HOST=mongodb
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - mongodb
      - redis
    volumes:
      - ./02-core-services/cultural-wisdom:/app
    command: air -c .air.toml

  ai-service:
    build:
      context: ./02-core-services/ai-integration
      dockerfile: Dockerfile.dev
    ports:
      - "8083:8083"
    environment:
      - QDRANT_HOST=qdrant
      - REDIS_HOST=redis
    depends_on:
      - qdrant
      - redis
    volumes:
      - ./02-core-services/ai-integration:/app
    command: air -c .air.toml

  # 前端应用
  web-app:
    build:
      context: ./03-frontend/web-app
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    environment:
      - REACT_APP_API_BASE_URL=http://localhost:8080
    volumes:
      - ./03-frontend/web-app:/app
      - /app/node_modules
    command: npm run dev

volumes:
  postgres_data:
  redis_data:
  mongodb_data:
  qdrant_data:
```

## 🔧 开发工具配置

### 1. Go开发环境

#### Air配置（热重载）
```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

#### Makefile配置
```makefile
# Makefile
.PHONY: dev test build clean docker-up docker-down

# 开发环境
dev:
	air -c .air.toml

# 测试
test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 构建
build:
	go build -o bin/app .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/app-linux .

# 清理
clean:
	rm -rf tmp/ bin/ coverage.out coverage.html

# Docker
docker-up:
	docker-compose -f docker-compose.dev.yml up -d

docker-down:
	docker-compose -f docker-compose.dev.yml down

docker-logs:
	docker-compose -f docker-compose.dev.yml logs -f

# 数据库
db-migrate:
	migrate -path ./migrations -database "postgres://dev_user:dev_pass@localhost:5432/taishang_dev?sslmode=disable" up

db-rollback:
	migrate -path ./migrations -database "postgres://dev_user:dev_pass@localhost:5432/taishang_dev?sslmode=disable" down 1

# 代码质量
lint:
	golangci-lint run

format:
	gofmt -s -w .
	goimports -w .
```

### 2. TypeScript/React开发环境

#### Vite配置
```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@services': path.resolve(__dirname, './src/services'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@types': path.resolve(__dirname, './src/types'),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router-dom'],
          ui: ['antd'],
        },
      },
    },
  },
})
```

#### Package.json脚本
```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "lint": "eslint src --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "lint:fix": "eslint src --ext ts,tsx --fix",
    "type-check": "tsc --noEmit",
    "format": "prettier --write \"src/**/*.{ts,tsx,js,jsx,json,css,md}\"",
    "format:check": "prettier --check \"src/**/*.{ts,tsx,js,jsx,json,css,md}\""
  }
}
```

## 🚀 快速启动脚本

### 1. 全环境启动脚本
```bash
#!/bin/bash
# start-dev-env.sh

echo "🚀 启动太上老君开发环境..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 启动基础服务
echo "📦 启动基础服务（数据库、缓存等）..."
docker-compose -f docker-compose.dev.yml up -d postgres redis mongodb qdrant

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 数据库迁移
echo "🗄️ 执行数据库迁移..."
cd 01-infrastructure/database-layer
make db-migrate
cd ../..

# 启动后端服务
echo "🔧 启动后端服务..."
docker-compose -f docker-compose.dev.yml up -d auth-service api-gateway cultural-service ai-service

# 启动前端应用
echo "🎨 启动前端应用..."
docker-compose -f docker-compose.dev.yml up -d web-app

echo "✅ 开发环境启动完成！"
echo "🌐 前端应用: http://localhost:3000"
echo "🔌 API网关: http://localhost:8080"
echo "📊 服务状态: docker-compose -f docker-compose.dev.yml ps"
```

### 2. 窗口专用启动脚本

#### 窗口1 - 基础设施启动
```bash
#!/bin/bash
# start-infrastructure.sh

echo "🏗️ 启动基础设施开发环境..."

# 启动数据库服务
docker-compose -f docker-compose.dev.yml up -d postgres redis

# 进入认证系统目录
cd 01-infrastructure/auth-system
echo "🔐 启动认证系统开发模式..."
make dev &

# 进入API网关目录
cd ../api-gateway
echo "🚪 启动API网关开发模式..."
make dev &

echo "✅ 基础设施开发环境就绪！"
```

#### 窗口2 - 核心服务启动
```bash
#!/bin/bash
# start-core-services.sh

echo "⚙️ 启动核心服务开发环境..."

# 启动依赖服务
docker-compose -f docker-compose.dev.yml up -d mongodb qdrant

# 进入文化智慧服务目录
cd 02-core-services/cultural-wisdom
echo "📚 启动文化智慧服务开发模式..."
make dev &

# 进入AI集成服务目录
cd ../ai-integration
echo "🤖 启动AI集成服务开发模式..."
make dev &

echo "✅ 核心服务开发环境就绪！"
```

#### 窗口3 - 前端应用启动
```bash
#!/bin/bash
# start-frontend.sh

echo "🎨 启动前端开发环境..."

# 进入前端应用目录
cd 03-frontend/web-app

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install
fi

# 启动开发服务器
echo "🚀 启动前端开发服务器..."
npm run dev

echo "✅ 前端开发环境就绪！"
echo "🌐 访问地址: http://localhost:3000"
```

## 🔍 监控和调试工具

### 1. 服务监控配置

#### Prometheus配置
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'auth-service'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'

  - job_name: 'api-gateway'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'

  - job_name: 'cultural-service'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: '/metrics'

  - job_name: 'ai-service'
    static_configs:
      - targets: ['localhost:8083']
    metrics_path: '/metrics'
```

#### 日志聚合配置
```yaml
# docker-compose.logging.yml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports:
      - "9200:9200"

  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

  logstash:
    image: docker.elastic.co/logstash/logstash:8.8.0
    volumes:
      - ./config/logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    ports:
      - "5044:5044"
    depends_on:
      - elasticsearch
```

### 2. 开发工具集成

#### VS Code配置（如果需要）
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.gopath": "",
  "go.goroot": "",
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "typescript.preferences.importModuleSpecifier": "relative",
  "eslint.workingDirectories": ["03-frontend/web-app"],
  "prettier.configPath": "03-frontend/web-app/.prettierrc"
}
```

## 📊 环境健康检查

### 1. 健康检查脚本
```bash
#!/bin/bash
# health-check.sh

echo "🏥 执行环境健康检查..."

# 检查Docker服务
echo "📦 检查Docker服务..."
docker-compose -f docker-compose.dev.yml ps

# 检查数据库连接
echo "🗄️ 检查数据库连接..."
pg_isready -h localhost -p 5432 -U dev_user

# 检查Redis连接
echo "🔴 检查Redis连接..."
redis-cli -h localhost -p 6379 ping

# 检查MongoDB连接
echo "🍃 检查MongoDB连接..."
mongosh --host localhost:27017 --eval "db.adminCommand('ping')"

# 检查Qdrant连接
echo "🔍 检查Qdrant连接..."
curl -s http://localhost:6333/health

# 检查后端服务
echo "🔧 检查后端服务..."
curl -s http://localhost:8081/health || echo "认证服务未响应"
curl -s http://localhost:8080/health || echo "API网关未响应"
curl -s http://localhost:8082/health || echo "文化智慧服务未响应"
curl -s http://localhost:8083/health || echo "AI服务未响应"

# 检查前端应用
echo "🎨 检查前端应用..."
curl -s http://localhost:3000 > /dev/null && echo "前端应用正常" || echo "前端应用未响应"

echo "✅ 健康检查完成！"
```

## 🎯 环境配置检查清单

- [ ] Docker和Docker Compose已安装
- [ ] 所有必要的端口未被占用
- [ ] 数据库初始化脚本准备就绪
- [ ] 各服务的Dockerfile.dev已创建
- [ ] 热重载工具配置完成
- [ ] 环境变量配置正确
- [ ] 健康检查脚本可执行
- [ ] 监控和日志系统配置
- [ ] 各窗口启动脚本准备
- [ ] 开发工具集成配置完成

## 🚀 快速开始

1. **克隆项目并进入目录**
   ```bash
   cd d:\work\taishanglaojun\taishang-laojun-v2
   ```

2. **启动开发环境**
   ```bash
   chmod +x start-dev-env.sh
   ./start-dev-env.sh
   ```

3. **验证环境**
   ```bash
   chmod +x health-check.sh
   ./health-check.sh
   ```

4. **开始并行开发**
   - 窗口1: 运行 `./start-infrastructure.sh`
   - 窗口2: 运行 `./start-core-services.sh`
   - 窗口3: 运行 `./start-frontend.sh`

现在您已经拥有了一个完整的并行开发环境！🎉