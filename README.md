# 码道 (Code Taoist)

> 🧘‍♂️ **"道法自然，无为而治"** - 让AI成为你的编程伙伴

码道是一个融合道家哲学的智能编程助手生态系统，通过自然语言交互帮助开发者高效完成编程任务。项目采用"本地CLI + 云端服务"的混合架构，既保证了本地操作的快速响应，又提供了强大的AI能力和团队协作功能。

## ✨ 核心特性

- 🤖 **AI智能编程**: 自然语言描述需求，AI自动生成代码
- 📚 **项目知识库**: 智能理解项目上下文，提供精准建议
- 🔐 **安全执行**: 内置安全验证，支持DryRun模式
- 🎨 **代码生成**: 模板化代码生成，支持多种编程语言
- 🌐 **跨平台**: 支持Windows、macOS、Linux
- 🔧 **可扩展**: 模块化设计，易于扩展和定制

## 🎯 使用场景

- **快速原型**: "创建一个RESTful API服务"
- **代码解释**: "解释这个算法的时间复杂度"
- **性能优化**: "优化这段代码的内存使用"
- **测试生成**: "为这个函数生成单元测试"
- **文档生成**: "根据代码生成API文档"
- **代码重构**: "将这个类重构为更好的设计模式"

## 🏗️ 项目架构

码道采用**混合架构设计**，将本地高效工具与云端AI服务完美结合：

- **本地CLI** (`offline/`) - Go语言开发的跨平台命令行工具，提供快速响应和离线能力
- **云端平台** (`online/`) - 基于微服务的AI智能体平台，提供强大的AI能力和团队协作

### 📁 目录结构

```
codetaoist/
├── offline/                    # CLI工具 (Go)
│   ├── cmd/lo/                # 主程序入口
│   ├── internal/              # 内部模块
│   │   ├── api/              # API客户端
│   │   ├── auth/             # 认证管理
│   │   ├── commands/         # 命令实现
│   │   ├── config/           # 配置管理
│   │   ├── executor/         # 命令执行器
│   │   └── utils/            # 工具函数
│   ├── pkg/                  # 公共包
│   │   ├── codegen/          # 代码生成器
│   │   ├── project/          # 项目管理器
│   │   └── security/         # 安全验证器
│   ├── scripts/              # 构建脚本
│   └── README.md             # CLI工具文档
├── online/                     # 微服务平台 (Python/Go)
│   ├── ai-service/           # AI智能体服务
│   ├── api-gateway/          # API网关
│   ├── auth-service/         # 认证服务
│   ├── memory-service/       # 记忆与知识库服务
│   ├── user-service/         # 用户管理服务
│   ├── docker-compose.yml   # 容器编排
│   └── README.md             # 平台服务文档
└── README.md                   # 项目总览
```

### 🔧 技术栈

| 组件 | 技术选型 | 特点 |
|------|----------|------|
| **CLI工具** | Go 1.22+ + Cobra + Viper | 跨平台、单二进制、快速启动 |
| **AI智能体** | Python 3.11+ + FastAPI + LangChain | 丰富的AI生态、高性能异步 |
| **基础服务** | Go 1.22+ + 高性能框架 | 高并发、低延迟、易部署 |
| **数据存储** | PostgreSQL + Chroma + MinIO | 结构化+向量化+对象存储 |
| **缓存层** | Redis | 高性能缓存和会话管理 |
| **认证授权** | Keycloak | 企业级身份认证和权限管理 |
| **容器化** | Docker + Kubernetes | 云原生部署和扩缩容 |

## 🚀 快速开始

### 🎯 选择你的使用方式

**🖥️ 仅使用CLI工具** - 适合个人开发者，快速体验AI编程助手
```bash
# 1. 克隆仓库
git clone https://github.com/codetaoist/taishanglaojun.git
cd taishanglaojun/offline

# 2. 构建CLI工具
go build -o lo ./cmd/lo

# 3. 开始使用
./lo --help
./lo ai "用Go写一个HTTP服务器"
```

**🌐 部署完整平台** - 适合团队使用，获得完整的AI协作体验
```bash
# 1. 进入在线服务目录
cd taishanglaojun/online

# 2. 配置环境
cp .env.example .env

# 3. 一键启动所有服务
docker-compose up -d

# 4. 访问服务
# - API文档: http://localhost:8001/docs
# - 管理后台: http://localhost:8080
```

### 💡 使用示例

**本地CLI使用**
```bash
# 基础命令
lo --help                           # 查看帮助
lo --version                        # 查看版本

# AI智能编程
lo ai "用Go写一个HTTP服务器"         # 代码生成
lo ai "解释这个函数的作用"           # 代码解释
lo ai "优化当前代码的性能"           # 代码优化

# 项目管理（需要平台服务）
lo login                            # 登录平台
lo project create "我的项目"        # 创建项目
lo ask "如何运行测试？"             # 知识库查询
```

**平台服务使用**
```bash
# 健康检查
curl http://localhost:8001/health

# AI聊天API
curl -X POST http://localhost:8001/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "用Python写一个爬虫", "project_id": "proj_123"}'
```



## 📚 详细文档

| 文档类型 | 链接 | 描述 |
|----------|------|------|
| 🖥️ **CLI工具** | [offline/README.md](./offline/README.md) | 本地命令行工具的安装、配置和使用指南 |
| 🌐 **平台服务** | [online/README.md](./online/README.md) | 微服务平台的部署、配置和API文档 |
| 📖 **API文档** | [docs.codetaoist.com/api](https://docs.codetaoist.com/api) | 完整的REST API接口文档 |
| 👨‍💻 **开发指南** | [docs.codetaoist.com/dev](https://docs.codetaoist.com/dev) | 贡献代码和扩展开发指南 |

## 🤝 参与贡献

我们欢迎所有形式的贡献！无论是代码、文档、问题反馈还是功能建议。

### 🛠️ 开发环境搭建

```bash
# 1. Fork并克隆仓库
git clone https://github.com/your-username/taishanglaojun.git
cd taishanglaojun

# 2. 设置CLI开发环境
cd offline
go mod download
go test ./...  # 运行测试

# 3. 设置平台开发环境
cd ../online
pip install -r ai-service/requirements.txt
pytest  # 运行Python测试

# 4. 启动开发环境
docker-compose -f docker-compose.dev.yml up -d
```

### 📝 贡献流程

1. **🍴 Fork项目** → 创建你的分支
2. **🔧 开发功能** → 编写代码和测试
3. **✅ 质量检查** → 运行测试和代码检查
4. **📤 提交PR** → 详细描述你的更改
5. **🔄 代码审查** → 响应反馈并完善

### 🎯 贡献方向

- **🐛 Bug修复** - 修复已知问题
- **✨ 新功能** - 添加有用的功能
- **📚 文档改进** - 完善使用文档
- **🧪 测试覆盖** - 增加测试用例
- **🎨 UI/UX** - 改进用户体验
- **🌍 国际化** - 支持多语言

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

## 🆘 获取帮助

- 📖 [官方文档](https://docs.codetaoist.com)
- 🐛 [问题反馈](https://github.com/codetaoist/taishanglaojun/issues)
- 💬 [讨论区](https://github.com/codetaoist/taishanglaojun/discussions)
- 📧 [邮件支持](mailto:support@codetaoist.com)

## 🌟 致谢

感谢所有贡献者和支持者！

---

**© 2025 码道 (Code Taoist) | 道法自然，码由心生**