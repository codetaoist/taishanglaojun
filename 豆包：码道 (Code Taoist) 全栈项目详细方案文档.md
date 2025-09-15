# "码道" (Code Taoist) 全栈项目详细方案文档

## 1. 项目概述

### 1.1 项目名称

码道 (Code Taoist)

### 1.2 核心标识

- **核心命令**：`ct`（取自 "Code Taoist" 首字母缩写，简洁易记，符合命令行工具的使用习惯）
- **核心域名**：`https://api.codetaoist.com`（作为中心化平台服务的统一入口，支持全球开发者访问）

### 1.3 核心理念

融合道家 "无为而治"（减少人工干预，让工具自主高效运行）与 "道法自然"（顺应开发者习惯，通过自然语言交互降低使用门槛）的哲学思想，打造**智能编程助手生态**。

通过本地轻量 CLI 与云端强大 AI 服务的协同，实现 "理解意图 - 自主规划 - 安全执行 - 记忆迭代" 的闭环，最终助力开发者达到 "天人合一" 的编码状态 —— 工具与开发者无缝协作，专注创意与逻辑而非机械操作。

### 1.4 项目定位

- **本质**：连接开发者与 AI 能力的 "智能中间层"，既是本地高效工具，也是团队协作平台。
- **形态**：以 Go 编写的跨平台 CLI（`ct`）为前端入口，以 Python 构建的中心化服务为后端大脑，兼顾本地响应速度与云端协作能力。
- **兼容性**：支持终端（Terminal、iTerm2、CMD、PowerShell）及操作系统（macOS、Windows、Linux），实现 "一处安装，多端可用"。

## 2. 核心目标

### 2.1 开发者体验目标

- **自然交互**：通过自然语言（中文 / 英文）直接下达开发指令（如 "生成一个 Go 语言的 HTTP 服务器框架"），无需记忆复杂命令。
- **本地高效**：CLI 启动速度 < 0.5s，本地轻量操作（文件读写、Git 提交）无需网络依赖，响应延迟 < 100ms。
- **自主执行**：AI 智能体可自动拆解复杂需求（如 "从数据库读取用户信息并生成 API 文档"），规划执行步骤（调用数据库工具、生成 Markdown），并在用户确认后自动执行。

### 2.2 团队协作目标

- **知识沉淀**：自动收集项目代码、文档、对话历史，形成团队专属 "知识库"，新成员可通过`ct ask "项目登录流程是怎样的"`快速获取上下文。
- **规范统一**：支持团队级代码规范（如命名风格、注释要求）、提示词模板（如 "生成符合 RESTful 规范的接口"），确保 AI 输出符合团队标准。
- **权限隔离**：基于项目维度的资源隔离，团队成员仅能访问授权项目的代码、记忆与工具，保障数据安全。

### 2.3 技术平台目标

- **可扩展性**：支持接入多类 LLM（OpenAI、DeepSeek、私有化模型），通过 "模型网关" 动态切换，满足不同场景（成本、精度）需求。
- **安全可控**：所有操作可审计，敏感动作（如修改生产代码）需二次确认，支持私有化部署（满足企业数据不出境需求）。
- **轻量化部署**：个人开发者可通过 Docker Compose 一键启动核心服务，企业级部署支持 Kubernetes 扩缩容，适配从单人到千级团队的规模。

## 3. 架构设计

采用 "**本地 CLI（Go）+ 中心化平台服务（Go/Python）+ 云原生基础设施**" 的混合架构，平衡灵活性与管理能力。

### 3.1 架构总览

代码

flowchart TD    subgraph 终端环境[用户终端环境]        A1[ct CLI<br>(Go)] -->|本地进程通信| A2[本地API代理<br>(轻量HTTP服务)]        A1 -->|富文本展示| A3[终端UI<br>(rich库)]        A1 -->|本地操作| A4[系统工具<br>(Git/Shell/文件系统)]    end     subgraph 网络层        B1[HTTPS加密传输] --> B2[API网关<br>(Nginx, 基于api.codetaoist.com)]    end     subgraph 中心化服务层        C1[认证授权服务<br>(Keycloak + Go)] --> C2[用户/项目管理服务<br>(Go)]        C3[AI智能体服务<br>(Python/FastAPI)] -->|调用工具| C4[工具服务<br>(Go, 如Git/数据库操作)]        C3 -->|检索记忆| C5[记忆与知识库服务<br>(Go + Chroma)]        C2 -->|数据存储| C6[元数据服务<br>(Go + PostgreSQL)]    end     subgraph 基础设施层        D1[PostgreSQL<br>(用户/项目元数据)]        D2[Chroma<br>(向量数据库, 代码/文档嵌入)]        D3[MinIO<br>(大文件/快照存储)]        D4[Redis<br>(缓存/会话)]        D5[Docker Compose<br>(单机部署)]        D6[Kubernetes<br>(集群部署)]    end     A1 <--> B1    B2 <--> C1    B2 <--> C3    C2 <--> D1    C5 <--> D2    C5 <--> D3    C1 <--> D4    C6 <--> D1

### 3.2 分层详细说明

#### 3.2.1 本地 CLI 层（前端）

- **核心组件**：`ct` 二进制文件（Go 编译，无依赖）。
- **功能职责**：
  - 接收用户输入（自然语言指令 / 命令），通过`typer`或`click`解析参数。
  - 本地安全沙箱内执行低风险操作（如创建文件、Git status），高风险操作（如 Git push、删除目录）需用户确认。
  - 通过`rich`库实现富文本终端展示（彩色日志、进度条、表格化结果）。
  - 作为本地代理，与中心化服务建立 HTTPS 连接，同步上下文与执行结果。
- **关键设计**：
  - 离线模式：支持缓存常用命令模板、本地历史，无网络时可执行基础操作（如`ct history`）。
  - 轻量启动：通过静态编译减少依赖，单文件大小 < 10MB，启动时间 < 300ms。

#### 3.2.2 接入与安全层

- **API 网关**：基于 Nginx/Traefik，部署在`api.codetaoist.com`域名下，负责：
  - 路由转发（将`ct`请求分发至对应服务，如`/auth`→认证服务，`/ai`→智能体服务）。
  - 安全控制（SSL 终止、请求限流、IP 黑名单过滤）。
  - 监控统计（记录请求延迟、成功率，为告警提供数据）。
- **认证授权**：基于 Keycloak 实现：
  - 支持 OAuth 2.1/OIDC 协议，`ct login`通过设备码流程（Device Flow）完成认证（无需在终端输入密码）。
  - 基于 RBAC 模型管理权限（如 "开发者" 可调用 AI 生成代码，"管理员" 可配置团队模型权限）。

#### 3.2.3 核心业务层

- **用户 / 项目管理服务（Go）**：
  - 管理实体：用户（账号、偏好设置）、团队（成员、角色）、项目（名称、关联代码仓库、权限配置）。
  - 核心接口：`/projects`（创建 / 查询项目）、`/teams`（添加成员）、`/users/preferences`（更新模型偏好）。
- **AI 智能体服务（Python/FastAPI）**：
  - 核心框架：基于 LangChain 实现意图识别、任务分解、工具调用链。
  - 能力：
    - 解析自然语言需求（如 "修复这个函数的空指针异常"），生成可执行步骤（"1. 定位异常位置；2. 添加非空判断；3. 生成测试用例"）。
    - 集成 OpenAI/DeepSeek SDK，根据项目上下文（代码片段、历史对话）调用合适的 LLM。
    - 通过工具服务调用外部能力（如 GitLab API 拉取代码、数据库查询）。
- **记忆与知识库服务（Go）**：
  - 数据处理：将项目代码、文档、对话历史转换为向量（通过 Embedding 模型），存储至 Chroma。
  - 检索能力：支持语义检索（如 "找到与用户登录相关的代码"），返回最相关的上下文给 AI 智能体。
  - 隔离机制：通过 Chroma 的命名空间（Namespace）实现项目级数据隔离（每个项目一个命名空间）。

#### 3.2.4 数据与基础设施层

- **元数据库（PostgreSQL）**：存储结构化数据：
  - 用户信息（账号、邮箱、角色）、团队关系（团队 - 成员映射）。
  - 项目元数据（ID、名称、关联仓库地址、创建时间）。
  - 操作审计日志（谁在何时执行了什么命令，结果如何）。
- **向量数据库（Chroma）**：存储非结构化数据的向量表示：
  - 代码片段（如函数、类定义）、文档（API 手册、需求说明）。
  - 对话历史（用户与 AI 的交互记录，用于上下文延续）。
- **对象存储（MinIO）**：存储大文件：
  - AI 生成的大型代码包、项目快照（用于版本回溯）。
  - 日志归档（超过 30 天的操作日志）。
- **缓存（Redis）**：
  - 会话缓存（用户登录状态，有效期 24 小时）。
  - 热点数据（常用项目的元数据、高频调用的提示词模板）。

## 4. 技术栈选型说明

| 层级         | 技术 / 工具              | 选型理由                                                     |
| ------------ | ------------------------ | ------------------------------------------------------------ |
| **本地 CLI** | Go                       | 编译为单二进制文件，跨平台（macOS/Windows/Linux），启动快、性能高，适合 CLI 工具。 |
|              | typer/click              | Go 生态中成熟的命令行参数解析库，支持自动生成帮助文档，降低开发成本。 |
|              | rich（Go 版）            | 提供终端富文本渲染（彩色、表格、进度条），提升用户交互体验。 |
| **后端服务** | Python 3.11+             | AI 生态丰富（LangChain、OpenAI SDK），适合快速开发复杂的智能体逻辑。 |
|              | FastAPI                  | 高性能异步 API 框架，自动生成 Swagger 文档，便于服务调试与集成。 |
|              | Go 1.22+                 | 用于用户 / 项目管理等基础服务，高并发支持好，部署简单（单二进制）。 |
| **数据库**   | PostgreSQL 16+           | 支持 JSONB 类型（存储半结构化数据如用户偏好），事务可靠，适合企业级元数据存储。 |
|              | Chroma                   | 轻量级向量数据库，API 简洁，支持命名空间隔离，适合项目级向量数据管理。 |
|              | MinIO                    | 兼容 S3 协议，可私有化部署，适合存储大文件且需权限控制的场景。 |
|              | Redis 7+                 | 高性能缓存，支持多种数据结构（字符串、哈希），适合会话与热点数据存储。 |
| **认证授权** | Keycloak                 | 开源 OIDC 实现，支持 SSO、RBAC，减少自研认证系统的安全风险。 |
| **部署工具** | Docker Compose           | 适合单机 / 小规模部署，通过 YAML 定义服务依赖（PostgreSQL、Chroma 等），一键启动。 |
|              | Kubernetes               | 适合大规模集群部署，支持自动扩缩容、滚动更新，保障服务高可用。 |
|              | Terraform                | 基础设施即代码（IaC），自动化创建云资源（VM、网络、K8s 集群），环境一致性强。 |
| **CI/CD**    | GitLab CI/GitHub Actions | 自动化代码检查、测试、镜像构建，支持多环境（开发 / 预发 / 生产）部署流水线。 |
| **监控日志** | Prometheus + Grafana     | 监控服务指标（API 延迟、错误率），通过仪表盘可视化，支持自定义告警。 |
|              | Loki                     | 轻量级日志聚合工具，适合存储结构化日志（JSON），查询效率高。 |

## 5. 详细功能规格

### 5.1 本地 CLI（`ct`）核心功能

#### 5.1.1 用户认证与配置

- `ct login`：通过设备码流程登录（终端显示二维码 / 链接，用户在浏览器确认后完成认证）。
- `ct logout`：清除本地认证信息，断开与平台的连接。
- `ct config set <key> <value>`：配置本地参数（如`ct config set model deepseek`设置默认模型为 DeepSeek）。
- `ct config get <key>`：查询配置（如`ct config get api_endpoint`查看当前 API 地址）。

#### 5.1.2 项目管理

- `ct project create <name>`：在平台创建新项目（自动生成项目 ID）。
- `ct project link <project-id>`：将当前本地目录与平台项目绑定（后续操作默认关联该项目）。
- `ct project list`：列出当前用户有权访问的项目（通过 rich 表格展示名称、ID、最后活动时间）。
- `ct project info`：查看当前绑定项目的详情（成员、关联仓库、模型配置）。

#### 5.1.3 自然语言交互（核心功能）

- 基础用法：`ct "自然语言指令"`，例如：
  - `ct "用Go写一个读取JSON文件的函数，处理可能的错误"`
  - `ct "解释当前目录下main.go中handleRequest函数的逻辑"`
  - `ct "基于项目中的数据库模型，生成一个用户注册的API接口文档"`
- 交互流程：
  1. CLI 将指令、本地上下文（当前目录文件列表、绑定项目 ID）发送至 AI 智能体服务。
  2. AI 返回执行计划（如 "1. 分析需求；2. 生成代码；3. 检查格式"），终端通过 rich 展示步骤。
  3. 用户确认后（输入`y`），AI 执行计划（生成代码→CLI 写入文件 / 展示结果）。

#### 5.1.4 本地工具集成

- **Git 操作**：AI 可自动调用（需用户授权），如`ct "提交当前修改，备注为'修复登录bug'"`→自动执行`git add .`→`git commit -m "..."`。
- **文件操作**：生成 / 修改文件（如`ct "在utils/目录下创建一个md5.go文件，实现字符串MD5加密"`），CLI 自动创建文件并写入内容。
- **代码检查**：集成`golint`/`pylint`等工具，如`ct "检查当前目录下的Python代码是否符合PEP8规范"`→返回检查结果。

#### 5.1.5 记忆与查询

- `ct ask "问题"`：查询项目知识库，如`ct ask "项目中数据库连接池的配置参数是什么"`→AI 从项目文档 / 代码中检索答案。
- `ct history`：查看本地与 AI 的对话历史（支持按时间 / 关键词筛选）。
- `ct memory sync`：手动同步本地文件至平台知识库（自动忽略`.gitignore`中的文件）。

### 5.2 中心化平台功能

#### 5.2.1 用户与团队管理

- **用户管理**：支持注册（邮箱验证）、资料修改（昵称、头像）、密码重置（通过 Keycloak）。
- **团队管理**：
  - 创建团队（`ct team create <name>`），添加成员（`ct team add <user-email> --role developer`）。
  - 角色定义：
    - 管理员（Admin）：管理团队成员、配置模型权限、删除项目。
    - 开发者（Developer）：创建项目、调用 AI 服务、修改项目内容。
    - 访客（Guest）：仅查看项目内容，无修改权限。

#### 5.2.2 AI 智能体配置

- **模型管理**：平台管理员可配置可用模型（如 OpenAI GPT-4、DeepSeek Code），设置每个模型的调用额度（如 "团队 A 每月可用 100 万 Token"）。
- **提示词模板**：团队可创建共享模板（如`ct prompt save "go-controller" "你是一个Go后端开发者，生成的Controller需包含参数校验和错误处理..."`），AI 调用时自动应用。
- **工具注册**：管理员可注册团队专属工具（如 "调用内部 Jenkins 构建"），定义工具参数与权限（如仅管理员可触发）。

#### 5.2.3 记忆与知识库

- **自动上下文抽取**：当项目绑定 Git 仓库后，平台定期拉取代码，自动抽取函数、类、注释等信息，转换为向量存储至 Chroma。
- **多模态检索**：支持按代码片段（如`ct ask "找到使用了redis.Pipeline的代码"`）、文档关键词（如`ct ask "搜索'支付流程'相关的文档"`）检索。
- **项目隔离**：通过 Chroma 命名空间和 PostgreSQL 行级权限，确保团队 A 无法访问团队 B 的项目记忆。

#### 5.2.4 审计与成本分析

- **操作日志**：记录所有关键操作（`ct`命令、AI 调用、文件修改），包含用户 ID、时间、操作内容、结果（成功 / 失败）。
- **成本统计**：按项目 / 用户统计 LLM Token 消耗（如`ct cost project <project-id> --month 2024-05`），生成成本趋势图（通过 Web 后台展示）。

### 5.3 安全设计

- **数据传输**：所有通信（CLI 与 API 网关、服务间调用）采用 TLS 1.3 加密，防止中间人攻击。
- **操作安全**：
  - 高风险操作（删除项目、推送代码至主分支）需二次确认（终端输入`confirm`）。
  - 工具调用权限粒度控制（如 "允许 AI 读取文件，但禁止删除文件"）。
- **敏感信息管理**：
  - 数据库密码、API Key 等通过 Kubernetes Secrets/Vault 存储，服务启动时动态加载，严禁硬编码。
  - 日志中自动脱敏敏感信息（如邮箱、Token）。

## 6. 部署、运维与监控方案

### 6.1 部署架构

#### 6.1.1 单机 / 小规模部署（适合团队内部试用）

- **工具**：Docker Compose
- **部署步骤**：
  1. 克隆代码仓库：`git clone https://github.com/xxx/codetaoist.git && cd codetaoist`。
  2. 配置环境变量：复制`docker-compose/.env.example`为`.env`，填写域名（如`API_DOMAIN=api.codetaoist.com`）、数据库密码等。
  3. 启动服务：`docker-compose up -d`（自动启动 PostgreSQL、Chroma、MinIO、Keycloak、API 服务）。
  4. 初始化数据：`docker exec -it codetaoist-api-1 python scripts/init_db.py`（创建默认管理员账号）。

#### 6.1.2 大规模 / 生产部署（适合企业级使用）

- **工具**：Kubernetes + Terraform
- **部署流程**：
  1. **基础设施创建**：通过 Terraform 在云厂商（AWS/Aliyun）创建：
     - VPC、子网、安全组（开放 80/443 端口，限制内部服务端口访问）。
     - Kubernetes 集群（至少 3 节点，2 核 4G 以上）。
     - 持久化存储（如 AWS EBS、阿里云云盘，用于数据库 / 存储数据）。
  2. **服务部署**：
     - 通过 Helm Charts 部署各组件（`helm install codetaoist ./charts`），包含 API 服务、数据库、缓存等。
     - 配置 Ingress（绑定`api.codetaoist.com`域名，配置 SSL 证书）。
  3. **监控部署**：部署 Prometheus、Grafana、Loki（`helm install monitoring ./charts/monitoring`）。

### 6.2 CI/CD 流水线（基于 GitLab CI）

yaml











```yaml
# .gitlab-ci.yml 示例
stages:
  - test
  - build
  - scan
  - deploy

# 测试阶段：代码检查、单元测试
test:
  image: golang:1.22-python3.11
  script:
    - cd cli && go test ./...  # 测试Go CLI
    - cd api && pip install -r requirements.txt && pytest  # 测试Python API

# 构建阶段：编译CLI、构建Docker镜像
build:
  image: docker:24
  script:
    - cd cli && GOOS=linux go build -o ct-linux  # 编译Linux版本CLI
    - cd api && docker build -t harbor.codetaoist.com/codetaoist/api:${CI_COMMIT_SHA} .  # 构建API镜像
    - docker push harbor.codetaoist.com/codetaoist/api:${CI_COMMIT_SHA}  # 推送到Harbor

# 安全扫描：镜像漏洞检查
scan:
  image: aquasec/trivy
  script:
    - trivy image harbor.codetaoist.com/codetaoist/api:${CI_COMMIT_SHA} --severity HIGH,CRITICAL

# 部署阶段：推送到K8s集群（预发环境）
deploy-staging:
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/api api=harbor.codetaoist.com/codetaoist/api:${CI_COMMIT_SHA} -n staging
  only:
    - develop  # 开发分支部署到预发

# 生产环境部署（需人工审批）
deploy-production:
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/api api=harbor.codetaoist.com/codetaoist/api:${CI_COMMIT_SHA} -n production
  only:
    - main  # 主分支部署到生产
  when: manual  # 手动触发审批
```

### 6.3 监控与告警

#### 6.3.1 关键监控指标

| 指标类型 | 具体指标                     | 阈值                 |
| -------- | ---------------------------- | -------------------- |
| API 性能 | P99 延迟                     | >3s 告警             |
|          | 错误率（5xx/4xx）            | >1% 告警             |
| AI 服务  | LLM 调用成功率               | <95% 告警            |
|          | Token 消耗（日累计）         | 超过配额 80% 预警    |
| 基础设施 | PostgreSQL 连接数            | >80% 最大连接数 预警 |
|          | 磁盘使用率（MinIO / 数据库） | >85% 告警            |

#### 6.3.2 日志与可视化

- **日志收集**：所有服务输出 JSON 格式日志（包含`timestamp`、`level`、`service`、`message`字段），通过 FluentBit 收集至 Loki。
- **监控面板**：Grafana 创建多维度仪表盘：
  - 服务健康度：各服务实例状态、API 延迟趋势。
  - 成本分析：按项目 / 模型的 Token 消耗饼图、趋势图。
  - 用户活跃：`ct`命令调用次数、活跃用户数。

#### 6.3.3 灾备策略

- **数据备份**：
  - PostgreSQL：每日全量备份 + 实时增量备份（基于 WAL），备份文件存储至 MinIO 并同步至异地。
  - Chroma 向量数据：每小时快照，保留 7 天历史版本。
- **故障转移**：
  - 核心服务（API、AI 智能体）在 K8s 中配置 2 个以上副本，分布在不同节点，单节点故障时自动切换。
  - 数据库主从架构：主库故障时，从库自动提升为主库（基于 Patroni）。

## 7. 团队开发与操作手册

### 7.1 开发环境搭建（开发者）

#### 7.1.1 本地依赖安装

bash











```bash
# 安装Go（1.22+）
brew install go  # macOS
# 或 https://golang.org/dl/ 下载安装包

# 安装Python（3.11+）
brew install python@3.11  # macOS
# 或 https://www.python.org/downloads/

# 安装Docker与Docker Compose
brew install docker docker-compose  # macOS
# 或参考官方文档安装

# 安装Git
brew install git  # macOS
```

#### 7.1.2 代码拉取与初始化

bash











```bash
# 克隆代码
git clone https://github.com/xxx/codetaoist.git
cd codetaoist

# 初始化Go CLI
cd cli
go mod download  # 下载依赖
go run main.go --help  # 验证是否可运行

# 初始化Python API服务
cd ../api
python -m venv venv
source venv/bin/activate  # Linux/macOS
# 或 venv\Scripts\activate  # Windows
pip install -r requirements.txt  # 安装依赖
uvicorn app.main:app --reload  # 启动开发服务器（热重载）
```

#### 7.1.3 启动基础设施

bash











```bash
# 启动本地数据库、缓存等服务
cd docker-compose
cp .env.example .env  # 配置本地环境变量（默认值可直接使用）
docker-compose up -d postgresql chroma minio redis keycloak

# 初始化数据库表结构
cd ../api
python scripts/init_db.py  # 创建表和默认数据
```

### 7.2 生产环境部署流程（DevOps）

1. **准备 K8s 集群**：
   - 通过 Terraform 创建集群：`cd terraform && terraform apply -var-file=prod.tfvars`。
   - 配置 kubectl：`gcloud container clusters get-credentials codetaoist-prod --zone us-central1-a`（根据云厂商调整）。
2. **配置镜像仓库**：
   - 在 Harbor 中创建项目`codetaoist`，配置机器人账号（用于 CI 推送镜像）。
3. **部署应用**：
   - 配置 Helm 参数：`cp charts/values.yaml charts/prod-values.yaml`，修改生产环境配置（如副本数、资源限制）。
   - 部署：`helm install codetaoist ./charts -f charts/prod-values.yaml -n production --create-namespace`。
4. **验证部署**：
   - 检查 Pod 状态：`kubectl get pods -n production`。
   - 测试 API：`curl https://api.codetaoist.com/health`，返回`{"status": "healthy"}`即为成功。

### 7.3 管理员操作手册

#### 7.3.1 用户与权限管理

- **添加用户**：登录 Keycloak 管理控制台（`https://auth.codetaoist.com/admin`），在 "Users" 页面点击 "Add user"，填写邮箱并设置初始密码。
- **分配团队角色**：在 Web 后台（`https://admin.codetaoist.com`）进入团队详情，点击 "添加成员"，输入用户邮箱并选择角色（Admin/Developer/Guest）。

#### 7.3.2 项目审批与管理

- **审批项目**：新创建的项目需管理员审批（防止垃圾项目），在 Web 后台 "待审批项目" 列表中点击 "通过" 或 "拒绝"。
- **清理过期项目**：执行`ct admin project clean --days 90`（删除 90 天无活动的项目），需二次确认。

#### 7.3.3 问题排查

- **查询用户操作日志**：在 Grafana Loki 中执行查询：`{service="api", username="user@example.com"} |= "error"`，筛选该用户的错误操作。
- **检查 AI 服务状态**：查看 AI 服务 Pod 日志：`kubectl logs -n production <ai-pod-name> -f`，排查模型调用失败原因。
- **恢复数据**：从备份恢复 PostgreSQL：`pg_restore -d codetaoist /backups/postgres/2024-05-01.dump`。

## 8. 总结与迭代规划

"码道" 项目通过融合道家哲学与现代 AI 技术，构建了一套从本地开发到云端协作的完整工具链，核心价值在于**降低开发心智负担，提升团队协作效率**。

### 迭代路线图

- **Phase 1（1-2 个月）**：实现 MVP，包含基础 CLI（`ct`）、AI 代码生成、单用户记忆功能。
- **Phase 2（3-4 个月）**：支持团队协作、项目隔离、多模型集成，完善 Web 管理后台。
- **Phase 3（5-6 个月）**：优化监控与成本控制，支持私有化部署，集成企业内部工具（如 Jira、Jenkins）。

通过循序渐进的迭代，最终实现 "让开发者专注创造，工具自然协作" 的 "无为而治" 境界。