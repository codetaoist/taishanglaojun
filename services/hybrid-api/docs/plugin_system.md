# 插件系统文档

## 概述

本文档描述了为太上老君系统实现的插件系统架构、功能和使用方法。插件系统允许用户动态加载、运行和管理插件，扩展系统的功能，同时提供安全隔离和生命周期管理。

## 系统架构

### 核心组件

1. **PluginSystemService** (`internal/service/plugin_system.go`)
   - 插件系统的核心服务，负责管理插件的生命周期
   - 提供插件的安装、启动、停止、升级和卸载功能
   - 维护插件注册表和运行状态

2. **Plugin** (`internal/plugin/plugin.go`)
   - 插件接口定义，所有插件必须实现此接口
   - 定义了插件的基本方法和元数据

3. **PluginRegistry** (`internal/plugin/registry.go`)
   - 插件注册表，维护所有已安装插件的信息
   - 提供插件查找和过滤功能

4. **PluginSandbox** (`internal/plugin/sandbox.go`)
   - 插件沙箱环境，提供安全隔离
   - 限制插件的资源访问和系统调用

5. **PluginManager** (`internal/plugin/manager.go`)
   - 插件管理器，负责插件的加载、卸载和执行
   - 处理插件间的通信和依赖关系

### 数据模型

1. **Plugin** (`internal/models/plugin.go`)
   - 插件数据模型，存储在数据库中
   - 包含插件的基本信息、状态和配置

2. **PluginLog** (`internal/models/plugin_log.go`)
   - 插件日志模型，记录插件的运行日志
   - 用于调试和监控插件运行状态

## API 接口

插件系统提供了以下 REST API 接口：

### 1. 获取插件列表
```
GET /api/v1/laojun/plugins/list
```

### 2. 安装插件
```
POST /api/v1/laojun/plugins/install
Content-Type: application/json

{
  "pluginId": "plugin-name",
  "version": "1.0.0",
  "sourceUrl": "https://example.com/plugin.zip"
}
```

### 3. 启动插件
```
POST /api/v1/laojun/plugins/start
Content-Type: application/json

{
  "pluginId": "plugin-name"
}
```

### 4. 停止插件
```
POST /api/v1/laojun/plugins/stop
Content-Type: application/json

{
  "pluginId": "plugin-name"
}
```

### 5. 升级插件
```
POST /api/v1/laojun/plugins/upgrade
Content-Type: application/json

{
  "pluginId": "plugin-name",
  "version": "2.0.0"
}
```

### 6. 卸载插件
```
DELETE /api/v1/laojun/plugins/uninstall
Content-Type: application/json

{
  "pluginId": "plugin-name"
}
```

## 插件开发指南

### 插件接口

所有插件必须实现 `Plugin` 接口：

```go
type Plugin interface {
    // 获取插件ID
    ID() string
    
    // 获取插件名称
    Name() string
    
    // 获取插件版本
    Version() string
    
    // 获取插件描述
    Description() string
    
    // 初始化插件
    Initialize(config map[string]interface{}) error
    
    // 启动插件
    Start() error
    
    // 停止插件
    Stop() error
    
    // 清理资源
    Cleanup() error
    
    // 处理请求
    Handle(ctx context.Context, req Request) (Response, error)
}
```

### 插件示例

```go
package main

import (
    "context"
    "github.com/codetaoist/taishanglaojun/api/internal/plugin"
)

type MyPlugin struct {
    id          string
    name        string
    version     string
    description string
    config      map[string]interface{}
}

func (p *MyPlugin) ID() string {
    return p.id
}

func (p *MyPlugin) Name() string {
    return p.name
}

func (p *MyPlugin) Version() string {
    return p.version
}

func (p *MyPlugin) Description() string {
    return p.description
}

func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *MyPlugin) Start() error {
    // 插件启动逻辑
    return nil
}

func (p *MyPlugin) Stop() error {
    // 插件停止逻辑
    return nil
}

func (p *MyPlugin) Cleanup() error {
    // 清理资源
    return nil
}

func (p *MyPlugin) Handle(ctx context.Context, req plugin.Request) (plugin.Response, error) {
    // 处理请求
    return plugin.Response{
        Data: map[string]interface{}{
            "message": "Hello from MyPlugin",
        },
    }, nil
}

// 插件入口点
func NewPlugin() plugin.Plugin {
    return &MyPlugin{
        id:          "my-plugin",
        name:        "My Plugin",
        version:     "1.0.0",
        description: "A sample plugin",
    }
}
```

## 安全机制

### 沙箱隔离

插件在沙箱环境中运行，限制以下操作：

1. 文件系统访问：只能访问指定的目录
2. 网络访问：只能访问指定的网络资源
3. 系统调用：限制危险的系统调用
4. 资源使用：限制CPU、内存和磁盘使用

### 权限控制

插件系统实现了基于角色的访问控制：

1. 插件安装和卸载需要管理员权限
2. 插件启动和停止需要插件所有者权限
3. 插件配置需要适当的权限验证

## 生命周期管理

插件的生命周期包括以下阶段：

1. **安装**：从源URL下载插件包，验证签名，解压到插件目录
2. **注册**：将插件信息注册到插件注册表
3. **初始化**：调用插件的Initialize方法，传递配置参数
4. **启动**：调用插件的Start方法，启动插件服务
5. **运行**：插件处理请求，执行业务逻辑
6. **停止**：调用插件的Stop方法，停止插件服务
7. **卸载**：清理插件资源，从注册表中移除插件

## 配置

插件系统的配置通过环境变量和配置文件进行：

```go
// 插件目录
PluginDir = getString("PLUGIN_DIR", "./plugins")

// 插件最大数量
MaxPlugins = getInt("MAX_PLUGINS", 10)

// 插件超时时间（秒）
PluginTimeout = getInt("PLUGIN_TIMEOUT", 30)

// 是否启用插件沙箱
EnableSandbox = getBool("ENABLE_SANDBOX", true)

// 插件资源限制
PluginMemoryLimit = getInt64("PLUGIN_MEMORY_LIMIT", 100*1024*1024) // 100MB
PluginCPULimit = getInt("PLUGIN_CPU_LIMIT", 50) // 50%
```

## 监控和日志

插件系统提供了全面的监控和日志功能：

1. **插件状态监控**：实时监控插件的运行状态
2. **性能监控**：监控插件的资源使用情况
3. **错误日志**：记录插件的错误和异常
4. **审计日志**：记录插件的管理操作

## 故障排除

### 常见问题

1. **插件安装失败**
   - 检查网络连接
   - 验证插件包的完整性
   - 确认插件签名

2. **插件启动失败**
   - 检查插件配置
   - 查看插件日志
   - 确认依赖项

3. **插件运行异常**
   - 查看错误日志
   - 检查资源使用情况
   - 验证插件权限

### 调试技巧

1. 使用插件日志进行调试
2. 启用详细日志记录
3. 使用插件测试工具
4. 检查沙箱配置

## 未来计划

1. 插件市场：提供插件发现和分享平台
2. 插件模板：提供插件开发模板和工具
3. 插件热更新：支持插件的热更新
4. 插件依赖管理：支持插件间的依赖关系管理
5. 插件版本控制：支持插件的版本控制和回滚