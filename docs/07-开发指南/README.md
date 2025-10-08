# 太上老君AI平台 - 开发指南

## 📋 开发指南概览

本目录包含太上老君AI平台的完整开发指南，为开发团队提供标准化的开发流程、代码规范和最佳实践。

## 📚 指南文档

### 🎯 开发概览
- **[开发概览](./development-overview.md)** - 开发流程和团队协作概述

### 🛠️ 环境配置
- **[环境搭建](./environment-setup.md)** - 开发环境配置指南

### 💻 技术开发指南
- **[后端开发指南](./backend-development.md)** - Go微服务开发规范
- **[前端开发指南](./frontend-development.md)** - React/TypeScript开发规范
- **[AI开发指南](./ai-development.md)** - AI模型集成和开发规范

### 🗄️ 数据库设计
- **[数据库设计指南](./database-design.md)** - 数据库架构和设计规范

### 🧪 测试指南
- **[测试指南](./testing-guide.md)** - 单元测试、集成测试和E2E测试

## 🚀 快速开始

### 1. 环境准备
```bash
# 克隆项目
git clone https://github.com/taishanglaojun/taishanglaojun.git
cd taishanglaojun

# 安装依赖
make install

# 启动开发环境
make dev
```

### 2. 开发流程
1. **创建功能分支**: `git checkout -b feature/your-feature`
2. **编写代码**: 遵循代码规范
3. **编写测试**: 确保测试覆盖率
4. **提交代码**: 使用规范的提交信息
5. **创建PR**: 通过代码审查

### 3. 代码规范
- **Go代码**: 遵循Go官方规范和项目约定
- **TypeScript**: 使用ESLint和Prettier
- **提交信息**: 使用Conventional Commits规范

## 📊 开发状态

| 模块 | 状态 | 负责人 | 最后更新 |
|------|------|--------|----------|
| 后端服务 | ✅ 开发中 | 后端团队 | 2024-12 |
| Web前端 | ✅ 开发中 | 前端团队 | 2024-12 |
| 移动端 | 🚧 开发中 | 移动端团队 | 2024-12 |
| 桌面端 | 📋 计划中 | 桌面端团队 | 2024-12 |

## 🔧 开发工具

### 必需工具
- **Go 1.21+**: 后端开发
- **Node.js 18+**: 前端开发
- **Docker**: 容器化开发
- **PostgreSQL**: 主数据库
- **Redis**: 缓存数据库

### 推荐工具
- **VS Code**: 代码编辑器
- **Postman**: API测试
- **DBeaver**: 数据库管理
- **Git**: 版本控制

## 📖 相关文档

- **[项目概览](../00-项目概览/README.md)** - 平台整体介绍
- **[架构设计](../02-架构设计/README.md)** - 技术架构详解
- **[核心服务](../03-核心服务/README.md)** - 后端服务文档
- **[API文档](../06-API文档/README.md)** - API接口文档
- **[部署运维](../08-部署运维/README.md)** - 部署和运维指南

## 🤝 贡献指南

### 代码贡献
1. Fork项目到个人仓库
2. 创建功能分支
3. 提交代码变更
4. 创建Pull Request
5. 通过代码审查

### 文档贡献
1. 发现文档问题或改进点
2. 编辑相关文档
3. 提交文档更新
4. 通过文档审查

## 📞 技术支持

- **开发问题**: dev-support@taishanglaojun.com
- **技术讨论**: 开发者微信群
- **GitHub Issues**: [提交问题](https://github.com/taishanglaojun/issues)

---

**最后更新**: 2024年12月  
**文档版本**: v1.0  
**维护团队**: 开发团队