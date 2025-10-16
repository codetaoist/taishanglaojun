# 太上老君AI - macOS桌面应用

一个功能丰富的macOS原生桌面应用，集成了AI聊天、项目管理、文件传输和桌面宠物等功能。

## 功能特性

### 🎯 核心功能
- **AI聊天助手**: 智能对话，支持语音输入和语音合成
- **项目管理**: 完整的项目生命周期管理
- **文件传输**: 高效的跨设备文件同步
- **桌面宠物**: 可爱的桌面伴侣，支持交互

### 🎨 界面特性
- **原生Cocoa界面**: 完全符合macOS设计规范
- **Metal渲染**: 高性能图形渲染和动画效果
- **响应式布局**: 支持各种屏幕尺寸和分辨率
- **深色模式**: 完整支持macOS深色模式

### 🔧 系统集成
- **菜单栏集成**: 完整的应用菜单和快捷键
- **状态栏支持**: 常驻状态栏，快速访问
- **通知中心**: 系统通知和提醒功能
- **多窗口管理**: 支持多窗口和全屏模式

## 技术架构

### 核心技术栈
- **Swift 5.9+**: 现代Swift语言特性
- **Cocoa/AppKit**: 原生macOS UI框架
- **Metal**: 高性能图形渲染
- **Core Animation**: 流畅的动画效果
- **UserNotifications**: 系统通知集成

### 依赖库
- **Alamofire**: HTTP网络请求
- **SwiftyJSON**: JSON数据处理
- **Starscream**: WebSocket连接
- **CryptoSwift**: 加密和安全
- **Swift-Log**: 日志记录

## 项目结构

```
Sources/
├── AppDelegate.swift              # 应用程序委托
├── MainWindowController.swift     # 主窗口控制器
├── MenuManager.swift             # 菜单管理器
├── DesktopPetWindow.swift        # 桌面宠物窗口
├── MetalRenderer.swift           # Metal渲染器
├── Shaders.metal                 # Metal着色器
├── NotificationManager.swift     # 通知管理器
├── ProjectViewController.swift   # 项目管理界面
├── ChatViewController.swift      # 聊天界面
├── FileTransferViewController.swift # 文件传输界面
├── SettingsViewController.swift  # 设置界面
├── AuthManager.swift            # 认证管理
├── ChatManager.swift            # 聊天管理
├── FileTransfer.swift           # 文件传输
├── FriendManager.swift          # 好友管理
├── HttpClient.swift             # HTTP客户端
└── Models/
    └── DesktopPet.swift         # 桌面宠物模型
```

## 构建和运行

### 系统要求
- macOS 13.0 或更高版本
- Xcode 15.0 或更高版本
- Swift 5.9 或更高版本

### 构建步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/taishanglaojun/desktop-apps.git
   cd desktop-apps/macos
   ```

2. **安装依赖**
   ```bash
   swift package resolve
   ```

3. **构建项目**
   ```bash
   swift build
   ```

4. **运行应用**
   ```bash
   swift run
   ```

### 使用Xcode开发

1. 在Xcode中打开`Package.swift`文件
2. 等待依赖解析完成
3. 选择目标设备并运行

## 功能说明

### 桌面宠物
- **显示/隐藏**: 通过菜单栏或快捷键控制
- **交互操作**: 点击、拖拽、右键菜单
- **状态系统**: 快乐、悲伤、睡觉、吃饭、玩耍
- **动画效果**: 流畅的Metal渲染动画

### AI聊天
- **智能对话**: 支持多轮对话和上下文理解
- **语音输入**: 系统语音识别集成
- **语音合成**: 自然的语音回复
- **消息历史**: 完整的聊天记录管理

### 项目管理
- **项目创建**: 支持多种项目类型
- **任务管理**: 完整的任务生命周期
- **团队协作**: 成员管理和权限控制
- **进度跟踪**: 可视化进度展示

### 文件传输
- **跨设备同步**: 支持多设备文件同步
- **断点续传**: 网络中断自动恢复
- **传输监控**: 实时传输状态和速度
- **文件管理**: 完整的文件操作功能

## 快捷键

| 功能 | 快捷键 |
|------|--------|
| 新建项目 | ⌘N |
| 打开项目 | ⌘O |
| 保存 | ⌘S |
| 查找 | ⌘⇧F |
| 偏好设置 | ⌘, |
| 显示/隐藏侧边栏 | ⌘⌃S |
| 全屏模式 | ⌘⌃F |
| 最小化 | ⌘M |
| 关闭窗口 | ⌘W |
| 退出应用 | ⌘Q |

## 配置说明

### 应用设置
- **外观主题**: 浅色/深色/自动
- **通知设置**: 自定义通知类型和频率
- **隐私设置**: 数据收集和使用控制
- **高级设置**: 开发者选项和调试功能

### 网络配置
- **服务器地址**: 配置后端服务地址
- **连接超时**: 网络请求超时设置
- **代理设置**: HTTP/SOCKS代理支持
- **SSL证书**: 自定义证书验证

## 故障排除

### 常见问题

1. **应用无法启动**
   - 检查macOS版本是否满足要求
   - 确认Xcode和Swift版本
   - 重新构建项目

2. **Metal渲染问题**
   - 检查显卡驱动是否最新
   - 确认Metal支持情况
   - 尝试重启应用

3. **网络连接失败**
   - 检查网络连接状态
   - 确认服务器地址配置
   - 检查防火墙设置

4. **通知不显示**
   - 检查系统通知权限
   - 确认通知设置
   - 重新授权通知权限

### 日志和调试
- 应用日志位置: `~/Library/Logs/TaishanglaojunDesktop/`
- 调试模式: 在Xcode中运行查看详细日志
- 性能分析: 使用Instruments进行性能分析

## 贡献指南

欢迎贡献代码和提出建议！

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 联系我们

- 官网: https://taishanglaojun.ai
- 邮箱: support@taishanglaojun.ai
- GitHub: https://github.com/taishanglaojun
- 反馈: https://github.com/taishanglaojun/desktop-apps/issues