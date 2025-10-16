# 太上老君AI平台 - 环境搭建指南

## 概述

本指南将帮助开发者快速搭建太上老君AI平台的完整开发环境，包括前端、后端、AI服务和相关基础设施的配置。

## 系统要求

### 硬件要求

```yaml
minimum_requirements:
  cpu: 4核心
  memory: 8GB RAM
  storage: 50GB 可用空间
  network: 稳定的互联网连接

recommended_requirements:
  cpu: 8核心
  memory: 16GB RAM
  storage: 100GB SSD
  gpu: NVIDIA GPU (AI开发推荐)
  network: 高速互联网连接
```

### 软件要求

```yaml
operating_systems:
  - Windows 10/11 (推荐WSL2)
  - macOS 12+
  - Ubuntu 20.04+
  - CentOS 8+

required_software:
  - Docker Desktop 4.0+
  - Docker Compose 2.0+
  - Git 2.30+
  - Node.js 18+
  - Go 1.21+
  - Python 3.9+
```

## 基础环境安装

### 1. Docker环境配置

#### Windows (WSL2)

```powershell
# 启用WSL2
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# 重启后设置WSL2为默认版本
wsl --set-default-version 2

# 安装Ubuntu
wsl --install -d Ubuntu-20.04

# 安装Docker Desktop
# 下载并安装 Docker Desktop for Windows
# 确保启用WSL2集成
```

#### macOS

```bash
# 使用Homebrew安装
brew install --cask docker
brew install docker-compose
brew install git
brew install node
brew install go
brew install python@3.9

# 启动Docker Desktop
open /Applications/Docker.app
```

#### Ubuntu/CentOS

```bash
# Ubuntu
sudo apt update
sudo apt install -y docker.io docker-compose git curl

# 添加用户到docker组
sudo usermod -aG docker $USER

# 安装Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 安装Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# 安装Python
sudo apt install -y python3.9 python3.9-pip python3.9-venv
```

### 2. 开发工具安装

#### Visual Studio Code配置

```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-typescript-next",
    "bradlc.vscode-tailwindcss",
    "ms-python.python",
    "ms-vscode.vscode-docker",
    "ms-kubernetes-tools.vscode-kubernetes-tools",
    "humao.rest-client",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint"
  ]
}
```

```json
// .vscode/settings.json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true,
    "source.organizeImports": true
  },
  "python.defaultInterpreterPath": "./venv/bin/python",
  "python.linting.enabled": true,
  "python.linting.pylintEnabled": true,
  "docker.showStartPage": false
}
```

#### Git配置

```bash
# 全局配置
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
git config --global init.defaultBranch main
git config --global pull.rebase false

# 配置SSH密钥
ssh-keygen -t ed25519 -C "your.email@example.com"
cat ~/.ssh/id_ed25519.pub
# 将公钥添加到GitLab/GitHub

# 配置Git别名
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.lg "log --oneline --graph --decorate --all"
```

## 项目环境搭建

### 1. 克隆项目

```bash
# 克隆主项目
git clone git@gitlab.com:taishanglaojun/platform.git
cd platform

# 初始化子模块
git submodule update --init --recursive

# 查看项目结构
tree -L 2
```

### 2. 环境变量配置

```bash
# 复制环境变量模板
cp .env.example .env.local

# 编辑环境变量
nano .env.local
```

```bash
# .env.local 示例配置
# 数据库配置
DATABASE_URL=postgresql://dev_user:dev_password@localhost:5432/taishanglaojun_dev
REDIS_URL=redis://localhost:6379/0

# JWT配置
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRES_IN=24h

# AI服务配置
OPENAI_API_KEY=your-openai-api-key
OPENAI_BASE_URL=https://api.openai.com/v1

# 对象存储配置
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# 邮件服务配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password

# 监控配置
PROMETHEUS_ENDPOINT=http://localhost:9090
GRAFANA_ENDPOINT=http://localhost:3000
```

### 3. Docker开发环境

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  # PostgreSQL数据库
  postgres:
    image: postgres:15-alpine
    container_name: taishang-postgres
    environment:
      POSTGRES_DB: taishanglaojun_dev
      POSTGRES_USER: dev_user
      POSTGRES_PASSWORD: dev_password
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dev_user -d taishanglaojun_dev"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis缓存
  redis:
    image: redis:7-alpine
    container_name: taishang-redis
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes --requirepass dev_password
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  # Elasticsearch搜索引擎
  elasticsearch:
    image: elasticsearch:8.8.0
    container_name: taishang-elasticsearch
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
      - "9300:9300"
    volumes:
      - es_data:/usr/share/elasticsearch/data
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 5

  # MinIO对象存储
  minio:
    image: minio/minio:latest
    container_name: taishang-minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

  # RabbitMQ消息队列
  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: taishang-rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: dev_user
      RABBITMQ_DEFAULT_PASS: dev_password
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Prometheus监控
  prometheus:
    image: prom/prometheus:latest
    container_name: taishang-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  # Grafana可视化
  grafana:
    image: grafana/grafana:latest
    container_name: taishang-grafana
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin123
    volumes:
      - grafana_data:/var/lib/grafana
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./config/grafana/datasources:/etc/grafana/provisioning/datasources

  # Jaeger链路追踪
  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: taishang-jaeger
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      COLLECTOR_OTLP_ENABLED: true

volumes:
  postgres_data:
  redis_data:
  es_data:
  minio_data:
  rabbitmq_data:
  prometheus_data:
  grafana_data:

networks:
  default:
    name: taishang-network
```

### 4. 启动开发环境

```bash
# 启动所有基础服务
docker-compose -f docker-compose.dev.yml up -d

# 检查服务状态
docker-compose -f docker-compose.dev.yml ps

# 查看服务日志
docker-compose -f docker-compose.dev.yml logs -f postgres

# 等待所有服务健康检查通过
./scripts/wait-for-services.sh
```

## 前端环境配置

### 1. Node.js项目初始化

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 或使用yarn
yarn install

# 或使用pnpm (推荐)
pnpm install
```

### 2. 前端配置文件

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@types': path.resolve(__dirname, './src/types'),
      '@services': path.resolve(__dirname, './src/services'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          antd: ['antd'],
          utils: ['lodash', 'dayjs'],
        },
      },
    },
  },
});
```

```json
// package.json scripts
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint src --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "lint:fix": "eslint src --ext ts,tsx --fix",
    "type-check": "tsc --noEmit",
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage"
  }
}
```

### 3. 启动前端开发服务器

```bash
# 开发模式启动
npm run dev

# 或
yarn dev

# 或
pnpm dev

# 访问 http://localhost:3000
```

## 后端环境配置

### 1. Go项目初始化

```bash
# 进入后端目录
cd backend

# 初始化Go模块
go mod init github.com/taishanglaojun/platform

# 下载依赖
go mod download

# 安装开发工具
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golang/mock/mockgen@latest
```

### 2. Go项目配置

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/taishanglaojun/platform/internal/config"
    "github.com/taishanglaojun/platform/internal/server"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 创建服务器
    srv, err := server.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // 启动服务器
    httpServer := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: srv.Handler(),
    }

    go func() {
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    log.Printf("Server started on port %s", cfg.Server.Port)

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := httpServer.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}
```

```yaml
# .air.toml - 热重载配置
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
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

[screen]
  clear_on_rebuild = false
```

### 3. 启动后端开发服务器

```bash
# 使用Air热重载启动
air

# 或直接运行
go run cmd/server/main.go

# 访问 http://localhost:8080
```

## AI服务环境配置

### 1. Python虚拟环境

```bash
# 进入AI服务目录
cd ai-service

# 创建虚拟环境
python3.9 -m venv venv

# 激活虚拟环境
# Linux/macOS
source venv/bin/activate

# Windows
venv\Scripts\activate

# 升级pip
pip install --upgrade pip

# 安装依赖
pip install -r requirements.txt
```

### 2. Python依赖配置

```txt
# requirements.txt
# Web框架
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0

# AI/ML库
torch==2.1.0
transformers==4.35.0
langchain==0.0.335
openai==1.3.0
sentence-transformers==2.2.2

# 数据处理
numpy==1.24.3
pandas==2.0.3
scikit-learn==1.3.0

# 图像处理
Pillow==10.0.1
opencv-python==4.8.1.78

# 数据库
asyncpg==0.29.0
redis==5.0.1

# 监控和日志
prometheus-client==0.19.0
structlog==23.2.0

# 开发工具
pytest==7.4.3
black==23.11.0
isort==5.12.0
mypy==1.7.0
```

```txt
# requirements-dev.txt
-r requirements.txt

# 开发工具
pytest-asyncio==0.21.1
pytest-cov==4.1.0
httpx==0.25.2
pytest-mock==3.12.0

# 代码质量
flake8==6.1.0
bandit==1.7.5
safety==2.3.5

# 文档
mkdocs==1.5.3
mkdocs-material==9.4.8
```

### 3. AI服务配置

```python
# config/settings.py
from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # 服务配置
    app_name: str = "Taishang Laojun AI Service"
    debug: bool = False
    host: str = "0.0.0.0"
    port: int = 8000
    
    # 数据库配置
    database_url: str
    redis_url: str
    
    # AI模型配置
    openai_api_key: Optional[str] = None
    openai_base_url: str = "https://api.openai.com/v1"
    model_cache_dir: str = "./models"
    
    # 向量数据库配置
    vector_db_url: Optional[str] = None
    embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    
    # 监控配置
    enable_metrics: bool = True
    log_level: str = "INFO"
    
    class Config:
        env_file = ".env"
        case_sensitive = False

settings = Settings()
```

### 4. 启动AI服务

```bash
# 激活虚拟环境
source venv/bin/activate

# 启动开发服务器
uvicorn main:app --reload --host 0.0.0.0 --port 8000

# 访问 http://localhost:8000/docs
```

## 数据库初始化

### 1. 数据库迁移脚本

```sql
-- scripts/init-db.sql
-- 创建数据库和用户
CREATE DATABASE taishanglaojun_dev;
CREATE USER dev_user WITH PASSWORD 'dev_password';
GRANT ALL PRIVILEGES ON DATABASE taishanglaojun_dev TO dev_user;

-- 切换到应用数据库
\c taishanglaojun_dev;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- 创建基础表结构
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 2. 数据库迁移工具

```go
// internal/database/migrate.go
package database

import (
    "database/sql"
    "fmt"
    
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB, migrationsPath string) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("could not create postgres driver: %w", err)
    }
    
    m, err := migrate.NewWithDatabaseInstance(
        fmt.Sprintf("file://%s", migrationsPath),
        "postgres",
        driver,
    )
    if err != nil {
        return fmt.Errorf("could not create migrate instance: %w", err)
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("could not run migrations: %w", err)
    }
    
    return nil
}
```

### 3. 执行数据库初始化

```bash
# 等待数据库启动
./scripts/wait-for-postgres.sh

# 运行迁移
go run cmd/migrate/main.go up

# 或使用migrate工具
migrate -path migrations -database "postgresql://dev_user:dev_password@localhost:5432/taishanglaojun_dev?sslmode=disable" up
```

## 开发工作流

### 1. 日常开发流程

```bash
# 1. 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 2. 启动前端开发服务器
cd frontend && npm run dev

# 3. 启动后端开发服务器
cd backend && air

# 4. 启动AI服务
cd ai-service && source venv/bin/activate && uvicorn main:app --reload

# 5. 运行测试
npm run test        # 前端测试
go test ./...       # 后端测试
pytest             # AI服务测试
```

### 2. 代码提交流程

```bash
# 1. 创建功能分支
git checkout -b feature/user-authentication

# 2. 开发和测试
# ... 编写代码 ...

# 3. 运行代码检查
npm run lint        # 前端代码检查
golangci-lint run   # 后端代码检查
black . && isort .  # Python代码格式化

# 4. 运行测试
npm run test:coverage
go test -race -coverprofile=coverage.out ./...
pytest --cov=.

# 5. 提交代码
git add .
git commit -m "feat: implement user authentication"

# 6. 推送分支
git push origin feature/user-authentication

# 7. 创建合并请求
# 在GitLab/GitHub上创建Pull Request/Merge Request
```

## 故障排除

### 1. 常见问题解决

```bash
# Docker服务无法启动
docker system prune -a  # 清理Docker缓存
docker-compose down -v  # 停止并删除卷
docker-compose up -d    # 重新启动

# 端口冲突
lsof -i :3000          # 查看端口占用
kill -9 <PID>          # 终止进程

# 数据库连接失败
docker-compose logs postgres  # 查看数据库日志
docker exec -it taishang-postgres psql -U dev_user -d taishanglaojun_dev

# Node.js依赖问题
rm -rf node_modules package-lock.json
npm install

# Go模块问题
go clean -modcache
go mod download

# Python依赖问题
pip install --upgrade pip
pip install -r requirements.txt --force-reinstall
```

### 2. 性能调优

```yaml
# Docker性能优化
docker_optimization:
  memory_limit: 4g
  cpu_limit: 2
  restart_policy: unless-stopped
  
  postgres:
    shared_buffers: 256MB
    effective_cache_size: 1GB
    work_mem: 4MB
    
  redis:
    maxmemory: 512mb
    maxmemory_policy: allkeys-lru
```

## 相关文档链接

- [开发指南概览](./development-overview.md)
- [前端开发指南](./frontend-development.md)
- [后端开发指南](./backend-development.md)
- [AI开发指南](./ai-development.md)
- [测试指南](./testing-guide.md)
- [API文档](../06-API文档/api-overview.md)
- [部署指南](../08-部署运维/deployment-guide.md)