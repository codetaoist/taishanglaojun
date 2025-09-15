# 码道 (Code Taoist) CLI 工具

> 🧘‍♂️ **"道法自然，无为而治"** - 让AI成为你的编程伙伴

码道是一个融合道家哲学的智能编程助手工具，通过自然语言交互帮助开发者高效完成编程任务。

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

## 🚀 快速开始

### 安装要求

- Go 1.22 或更高版本
- 网络连接（用于与码道平台通信）

### 安装步骤

1. **克隆代码库**
   ```bash
   git clone https://github.com/codetaoist/taishanglaojun.git
   cd taishanglaojun/offline
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **构建程序**
   ```bash
   # 基本构建
   go build -o lo ./cmd/lo
   
   # Windows下如需.exe扩展名
   go build -o lo.exe ./cmd/lo
   
   # 使用构建脚本（推荐）
   # Windows
   .\scripts\build.ps1
   
   # Linux/macOS
   ./scripts/build.sh
   ```

4. **验证构建**
   ```bash
   # Windows
   .\lo --help
   
   # Linux/macOS
   ./lo --help
   ```

5. **安装到系统路径**（可选）
   ```bash
   # Linux/macOS
   sudo cp lo /usr/local/bin/
   
   # Windows (管理员权限)
   copy lo.exe C:\Windows\System32\
   # 或添加当前目录到PATH环境变量
   ```

## 🎯 快速体验

构建完成后，立即体验码道的核心功能：

```bash
# Windows
.\lo --help          # 查看帮助
.\lo version         # 查看版本信息
.\lo login           # 登录平台（模拟）
.\lo project list    # 查看项目列表
.\lo ai "Hello World" # AI智能交互

# Linux/macOS
./lo --help
./lo version
./lo login
./lo project list
./lo ai "Hello World"
```

## 📖 使用指南

### 基础命令

#### 1. 登录
```bash
lo login
```
通过设备码流程登录到码道平台。

#### 2. 项目管理
```bash
# 创建项目
lo project create "我的项目"

# 列出项目
lo project list

# 绑定项目到当前目录
lo project link proj_1234567890

# 查看项目信息
lo project info
```

#### 3. AI 智能编程
```bash
# 自然语言编程指令
lo ai "用Go写一个HTTP服务器"
lo ai "解释这个函数的作用"
lo ai "优化当前代码的性能"

# 查询项目知识库
lo ask "项目的数据库配置在哪里？"
lo ask "如何运行测试？"
lo ask "API接口文档"
```

#### 4. 配置管理
```bash
# 查看所有配置
lo config get

# 设置配置
lo config set api-endpoint https://api.codetaoist.com
lo config set model deepseek
lo config set verbose true

# 删除配置
lo config unset verbose
```

#### 5. 其他实用命令
```bash
# 查看状态
lo status

# 查看历史
lo history

# 查看版本
lo version

# 退出登录
lo logout
```

### 高级用法

#### 自然语言编程示例

1. **创建 HTTP 服务器**
   ```bash
   lo ai "创建一个Go语言的HTTP服务器，包含健康检查接口"
   ```

2. **代码分析**
   ```bash
   lo ai "分析当前目录下的Go代码，找出潜在的性能问题"
   ```

3. **生成测试代码**
   ```bash
   lo ai "为main.go中的handleRequest函数生成单元测试"
   ```

4. **文档生成**
   ```bash
   lo ai "根据当前项目的API代码生成OpenAPI文档"
   ```

#### 项目知识库查询

```bash
# 查询配置信息
lo ask "数据库连接配置"

# 查询API文档
lo ask "用户登录接口的参数"

# 查询部署信息
lo ask "如何部署到生产环境"
```

## ⚙️ 配置说明

### 配置文件位置
- Linux/macOS: `~/.lo.yaml`
- Windows: `%USERPROFILE%\.lo.yaml`

### 主要配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `api-endpoint` | API服务地址 | `https://api.codetaoist.com` |
| `verbose` | 详细输出模式 | `false` |
| `model` | 默认AI模型 | `gpt-4` |
| `timeout` | 请求超时时间 | `30s` |

### 环境变量

可以通过环境变量覆盖配置：

```bash
export LO_API_ENDPOINT=https://api.codetaoist.com
export LO_VERBOSE=true
export LO_MODEL=deepseek
```

## 🔧 开发说明

### 项目结构

```
.
├── cmd/
│   └── lo/              # 主程序入口
│       └── main.go
├── internal/            # 内部包（不对外暴露）
│   ├── api/            # API客户端
│   ├── auth/           # 认证管理
│   ├── commands/       # 命令实现
│   ├── config/         # 配置管理
│   ├── executor/       # 命令执行器
│   └── utils/          # 工具函数
├── pkg/                # 公共包（可复用）
│   ├── codegen/        # 代码生成器
│   ├── project/        # 项目管理器
│   └── security/       # 安全验证器
├── scripts/            # 构建脚本
│   ├── build.sh
│   ├── build.ps1
│   └── test.sh
├── go.mod              # Go模块定义
└── README.md           # 说明文档
```

### 依赖库

- `github.com/spf13/cobra` - 命令行框架
- `github.com/fatih/color` - 彩色输出
- `github.com/olekukonko/tablewriter` - 表格输出
- `github.com/briandowns/spinner` - 加载动画
- `github.com/manifoldco/promptui` - 交互式提示
- `github.com/go-resty/resty/v2` - HTTP客户端
- `github.com/spf13/viper` - 配置管理

### 构建和测试

```bash
# 运行测试
go test ./...

# 构建
go build -o lo ./cmd/lo

# 使用构建脚本（推荐）
# Linux/macOS
./scripts/build.sh

# Windows
.\scripts\build.ps1

# 运行测试套件
./scripts/test.sh

# 交叉编译
# Linux
GOOS=linux GOARCH=amd64 go build -o lo-linux ./cmd/lo

# Windows
GOOS=windows GOARCH=amd64 go build -o lo.exe ./cmd/lo

# macOS
GOOS=darwin GOARCH=amd64 go build -o lo-macos ./cmd/lo
```

## 🔧 故障排除

### 常见问题

**1. Go命令未找到**
```bash
bash: go: command not found
```
**解决方案**: 安装Go 1.22+，参考 [Go官网](https://golang.org/dl/)

**2. 依赖下载失败**
```bash
go: module download failed
```
**解决方案**: 配置Go代理
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

**3. 构建缺少依赖**
```bash
missing go.sum entry
```
**解决方案**: 整理模块依赖
```bash
go mod tidy
go build -o lo ./cmd/lo
```

**4. Windows下权限问题**
```bash
Access denied
```
**解决方案**: 以管理员身份运行PowerShell

**5. 配置文件问题**
- **位置**: `~/.lo.yaml` (Linux/macOS) 或 `%USERPROFILE%\.lo.yaml` (Windows)
- **重置**: 删除配置文件重新开始

### 性能优化

**构建优化**:
```bash
# 启用Go模块缓存
go env -w GOPROXY=https://goproxy.cn,direct

# 并行构建
go build -race -o lo ./cmd/lo
```

**运行时优化**:
```bash
# 启用详细日志
lo config set verbose true

# 调整超时时间
lo config set timeout 60s
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 开发环境

```bash
# 克隆开发版本
git clone https://github.com/codetaoist/taishanglaojun.git
cd taishanglaojun/offline

# 安装开发依赖
go mod download

# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 静态检查
go vet ./...
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 获取帮助

- 📖 [官方文档](https://docs.codetaoist.com)
- 🐛 [问题反馈](https://github.com/codetaoist/taishanglaojun/issues)
- 💬 [讨论区](https://github.com/codetaoist/taishanglaojun/discussions)
- 📧 [邮件支持](mailto:support@codetaoist.com)

---

**码道 (Code Taoist)** - 让编程回归本质，专注创造 🎯