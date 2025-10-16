# PNPM 迁移文档

## 概述

本项目已从 npm 迁移到 pnpm 作为包管理器。这个变更提供了更好的性能、更严格的依赖管理和更高效的磁盘空间使用。

## 变更详情

### 包管理器变更
- **之前**: npm
- **现在**: pnpm

### 主要优势
1. **更快的安装速度**: pnpm 使用符号链接和硬链接来避免重复下载相同的包
2. **更严格的依赖管理**: 防止幽灵依赖（phantom dependencies）
3. **节省磁盘空间**: 全局存储避免重复包
4. **更好的 monorepo 支持**: 原生支持工作空间

## 命令对照表

| 操作 | npm 命令 | pnpm 命令 |
|------|----------|-----------|
| 安装依赖 | `npm install` | `pnpm install` |
| 添加依赖 | `npm install <package>` | `pnpm add <package>` |
| 添加开发依赖 | `npm install -D <package>` | `pnpm add -D <package>` |
| 移除依赖 | `npm uninstall <package>` | `pnpm remove <package>` |
| 运行脚本 | `npm run <script>` | `pnpm run <script>` 或 `pnpm <script>` |
| 启动开发服务器 | `npm run dev` | `pnpm dev` |
| 构建项目 | `npm run build` | `pnpm build` |
| 启动 Tauri 开发 | `npm run tauri dev` | `pnpm tauri dev` |

## 项目特定命令

### Tauri 开发
```bash
# 启动 Tauri 开发服务器
pnpm tauri dev

# 构建 Tauri 应用
pnpm tauri build

# 添加 Tauri 插件
pnpm add @tauri-apps/plugin-<plugin-name>
```

### 依赖管理
```bash
# 安装所有依赖
pnpm install

# 添加新的依赖
pnpm add <package-name>

# 添加开发依赖
pnpm add -D <package-name>

# 更新依赖
pnpm update

# 查看过时的依赖
pnpm outdated
```

## 已安装的 Tauri 相关依赖

以下 Tauri 相关依赖已通过 pnpm 安装：

- `@tauri-apps/api`: Tauri 核心 API
- `@tauri-apps/plugin-os`: 操作系统相关功能
- `@tauri-apps/plugin-dialog`: 对话框功能
- `@tauri-apps/plugin-shell`: Shell 命令执行

## 迁移过程中的问题解决

### 导入路径更新
在迁移过程中，我们更新了 Tauri v2 的导入路径：

```typescript
// 旧的导入方式
import { invoke } from '@tauri-apps/api/tauri';
import { platform } from '@tauri-apps/api/os';

// 新的导入方式 (Tauri v2)
import { invoke } from '@tauri-apps/api/core';
import { platform } from '@tauri-apps/plugin-os';
```

### 依赖安装
确保安装了必要的 Tauri 插件：
```bash
pnpm add @tauri-apps/api @tauri-apps/plugin-os
```

## 开发环境设置

1. 确保已安装 pnpm:
   ```bash
   npm install -g pnpm
   ```

2. 安装项目依赖:
   ```bash
   pnpm install
   ```

3. 启动开发服务器:
   ```bash
   pnpm tauri dev
   ```

## 注意事项

- 所有团队成员都应该使用 pnpm 而不是 npm
- 不要提交 `package-lock.json` 文件，使用 `pnpm-lock.yaml`
- 在 CI/CD 流水线中也需要更新为使用 pnpm

## 相关链接

- [pnpm 官方文档](https://pnpm.io/)
- [Tauri v2 文档](https://v2.tauri.app/)
- [pnpm vs npm 性能对比](https://pnpm.io/benchmarks)

---

**更新日期**: 2024年12月
**更新人**: AI Assistant
**状态**: 已完成迁移