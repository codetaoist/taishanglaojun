# 太上老君手表应用

太上老君任务管理系统的手表端应用，支持 Apple Watch 和 Wear OS 平台。

## 项目结构

```
watch-apps/
├── apple-watch/                    # Apple Watch 应用
│   ├── TaishanglaojunWatch/
│   │   ├── TaishanglaojunWatchApp.swift
│   │   ├── ContentView.swift
│   │   ├── Models/                 # 数据模型
│   │   ├── Views/                  # UI 视图
│   │   ├── Services/               # 核心服务
│   │   │   ├── WatchTaskManager.swift
│   │   │   ├── WatchConnectivityManager.swift
│   │   │   ├── LocationManager.swift
│   │   │   ├── WatchNotificationManager.swift
│   │   │   └── WatchHealthManager.swift
│   │   └── Assets.xcassets
│   └── TaishanglaojunWatch.xcodeproj
├── wear-os/                        # Wear OS 应用
│   ├── app/
│   │   ├── src/main/
│   │   │   ├── java/com/taishanglaojun/watch/
│   │   │   │   ├── MainActivity.kt
│   │   │   │   ├── data/           # 数据层
│   │   │   │   ├── ui/             # UI 层
│   │   │   │   ├── services/       # 服务层
│   │   │   │   └── utils/          # 工具类
│   │   │   ├── res/                # 资源文件
│   │   │   └── AndroidManifest.xml
│   │   └── build.gradle
│   ├── build.gradle
│   └── settings.gradle
├── build-scripts/                  # 构建脚本
│   ├── build-watch.sh
│   └── deploy-watch.ps1
├── test-scripts/                   # 测试脚本
│   └── test-watch.ps1
└── README.md
```

## 功能特性

### 核心功能
- 📱 **任务管理**: 查看、接受、开始、完成任务
- 📍 **位置服务**: 基于位置的任务提醒和导航
- 💓 **健康集成**: 集成健康数据，提供健康奖励
- 🔔 **智能通知**: 任务提醒、位置到达通知
- 🔄 **实时同步**: 与手机应用实时同步数据

### 平台特性

#### Apple Watch
- 使用 SwiftUI 构建现代化界面
- 集成 HealthKit 进行健康数据监控
- 支持 WatchConnectivity 与 iPhone 通信
- 利用 CoreLocation 进行精确定位
- 支持 Haptic Feedback 和本地通知

#### Wear OS
- 使用 Jetpack Compose 构建响应式 UI
- 集成 Health Services 进行健康监控
- 支持 Wear OS 连接 API 与手机通信
- 使用 FusedLocationProvider 进行定位
- 支持振动反馈和系统通知

## 开发环境要求

### Apple Watch 开发
- macOS 12.0 或更高版本
- Xcode 14.0 或更高版本
- iOS 16.0 SDK
- watchOS 9.0 SDK
- Apple Developer 账户（用于设备测试）

### Wear OS 开发
- Android Studio Arctic Fox 或更高版本
- Android SDK API 30 或更高版本
- Wear OS SDK
- Java 11 或更高版本
- Kotlin 1.8.0 或更高版本

## 快速开始

### 1. 克隆项目
```bash
git clone <repository-url>
cd taishanglaojun/watch-apps
```

### 2. 环境配置

#### Apple Watch
```bash
# 在 macOS 上
cd apple-watch
open TaishanglaojunWatch.xcodeproj
```

#### Wear OS
```bash
# 在任何平台上
cd wear-os
# 使用 Android Studio 打开项目
```

### 3. 依赖安装

#### Apple Watch
依赖通过 Xcode 自动管理，无需额外安装。

#### Wear OS
```bash
cd wear-os
./gradlew build
```

## 构建和部署

### 使用构建脚本

#### 构建所有平台（Linux/macOS）
```bash
chmod +x build-scripts/build-watch.sh
./build-scripts/build-watch.sh debug all
```

#### 构建特定平台
```bash
# 仅构建 Wear OS
./build-scripts/build-watch.sh release wear-os

# 仅构建 Apple Watch（需要 macOS）
./build-scripts/build-watch.sh debug apple-watch
```

### 使用部署脚本（Windows）

#### 部署到设备
```powershell
# 部署 Wear OS 应用
.\build-scripts\deploy-watch.ps1 -Platform wear-os -Install -Launch

# 查看帮助
Get-Help .\build-scripts\deploy-watch.ps1 -Full
```

### 手动构建

#### Apple Watch
```bash
cd apple-watch
xcodebuild -project TaishanglaojunWatch.xcodeproj \
           -scheme TaishanglaojunWatch \
           -configuration Debug \
           -destination 'generic/platform=watchOS' \
           build
```

#### Wear OS
```bash
cd wear-os
./gradlew assembleDebug
```

## 测试

### 运行测试脚本
```powershell
# 运行所有测试
.\test-scripts\test-watch.ps1 -TestType all -Platform all -Coverage

# 运行单元测试
.\test-scripts\test-watch.ps1 -TestType unit -Platform wear-os

# 详细输出
.\test-scripts\test-watch.ps1 -Verbose
```

### 手动测试

#### Apple Watch
```bash
cd apple-watch
xcodebuild test -project TaishanglaojunWatch.xcodeproj \
               -scheme TaishanglaojunWatch \
               -destination 'platform=watchOS Simulator,name=Apple Watch Series 9 (45mm)'
```

#### Wear OS
```bash
cd wear-os
./gradlew testDebugUnitTest
./gradlew connectedAndroidTest  # 需要连接设备
```

## 配置说明

### API 端点配置
在各自的配置文件中设置 API 端点：

#### Apple Watch
```swift
// WatchTaskManager.swift
private let baseURL = "https://api.taishanglaojun.com"
```

#### Wear OS
```kotlin
// ApiConfig.kt
const val BASE_URL = "https://api.taishanglaojun.com"
```

### 权限配置

#### Apple Watch (Info.plist)
```xml
<key>NSLocationWhenInUseUsageDescription</key>
<string>太上老君需要访问您的位置来提供基于位置的任务提醒和导航功能。</string>

<key>NSHealthShareUsageDescription</key>
<string>太上老君需要访问您的健康数据来提供个性化的任务建议和健康奖励。</string>
```

#### Wear OS (AndroidManifest.xml)
```xml
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.BODY_SENSORS" />
<uses-permission android:name="android.permission.ACTIVITY_RECOGNITION" />
```

## 开发指南

### 代码规范
- **Swift**: 遵循 Swift API Design Guidelines
- **Kotlin**: 遵循 Kotlin Coding Conventions
- 使用有意义的变量和函数命名
- 添加适当的注释和文档

### 架构模式
- **Apple Watch**: MVVM + Combine
- **Wear OS**: MVVM + Jetpack Compose + Coroutines

### 数据同步
两个平台使用相同的 API 端点和数据格式，确保数据一致性：

```json
{
  "id": "task-123",
  "title": "完成项目文档",
  "description": "编写项目的技术文档",
  "status": "pending",
  "priority": "high",
  "location": {
    "latitude": 39.9042,
    "longitude": 116.4074
  },
  "dueDate": "2024-01-15T10:00:00Z"
}
```

## 故障排除

### 常见问题

#### Apple Watch
1. **构建失败**: 检查 Xcode 版本和 SDK 版本
2. **设备连接问题**: 确保 Apple Watch 已配对并信任开发者
3. **权限问题**: 检查 Info.plist 中的权限描述

#### Wear OS
1. **Gradle 同步失败**: 检查网络连接和 Gradle 版本
2. **设备连接问题**: 启用开发者选项和 ADB 调试
3. **权限问题**: 检查 AndroidManifest.xml 中的权限声明

### 调试技巧

#### Apple Watch
```bash
# 查看设备日志
xcrun devicectl list devices
xcrun devicectl device install app --device <device-id> <app-path>
```

#### Wear OS
```bash
# 查看连接的设备
adb devices

# 查看应用日志
adb logcat | grep taishanglaojun

# 安装应用
adb install app-debug.apk
```

## 性能优化

### 电池优化
- 合理使用位置服务，避免持续定位
- 优化网络请求频率
- 使用后台任务限制

### 内存优化
- 及时释放不需要的资源
- 使用图片缓存机制
- 避免内存泄漏

## 发布流程

### Apple Watch
1. 在 Xcode 中配置签名和证书
2. 选择 Release 配置
3. 构建并上传到 App Store Connect
4. 提交审核

### Wear OS
1. 生成签名的 APK
2. 在 Google Play Console 中创建应用
3. 上传 APK 并填写应用信息
4. 提交审核

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目维护者: [维护者姓名]
- 邮箱: [邮箱地址]
- 项目链接: [项目链接]

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 支持基本的任务管理功能
- 集成位置服务和健康数据
- 支持 Apple Watch 和 Wear OS 平台

## 核心功能

### 🎯 任务管理系统
- **任务概览**：快速查看当前任务状态
- **快速操作**：一键接受推荐任务
- **进度跟踪**：实时更新任务进度
- **语音报告**：语音汇报工作进度

### 📱 设备连接
- **手机同步**：与手机端应用实时数据同步
- **离线模式**：支持离线查看和操作
- **推送通知**：及时接收任务更新通知
- **触觉反馈**：丰富的触觉交互体验

### 🔄 三轴协同
- **S轴（硅基）**：技术能力评估和匹配
- **C轴（碳基）**：人文智慧和文化传承
- **T轴（时间）**：时间管理和效率优化

### 💡 智能特性
- **健康集成**：心率监测与工作状态关联
- **活动提醒**：智能工作休息提醒
- **复杂功能**：表盘显示任务信息
- **快捷操作**：Digital Crown快速导航

## 项目结构

```
watch-apps/
├── README.md
├── apple-watch/                    # Apple Watch应用
│   └── TaishanglaojunWatch/       # Xcode项目
│       ├── TaishanglaojunWatch Watch App/
│       │   ├── Views/             # SwiftUI视图
│       │   ├── Models/            # 数据模型
│       │   ├── Services/          # 业务服务
│       │   ├── Managers/          # 管理器类
│       │   ├── Utils/             # 工具类
│       │   └── Resources/         # 资源文件
│       └── TaishanglaojunWatch.xcodeproj
├── wear-os/                       # Wear OS应用
│   ├── app/
│   │   └── src/main/
│   │       ├── java/com/taishanglaojun/watch/
│   │       │   ├── ui/            # Compose UI
│   │       │   ├── data/          # 数据层
│   │       │   ├── services/      # 服务层
│   │       │   └── utils/         # 工具类
│   │       └── res/               # 资源文件
│   ├── build.gradle.kts
│   └── settings.gradle.kts
└── shared/                        # 共享资源
    ├── models/                    # 共享数据模型
    ├── protocols/                 # 通信协议
    └── assets/                    # 共享资源文件
```

## 开发环境要求

### Apple Watch开发
- **Xcode**: 15.0+
- **watchOS**: 9.0+
- **Swift**: 5.9+
- **iOS设备**: 用于配对和测试

### Wear OS开发
- **Android Studio**: 2023.1+
- **Wear OS**: 3.0+ (API Level 30+)
- **Kotlin**: 1.9+
- **Android设备**: 用于配对和测试

## 快速开始

### Apple Watch应用

1. 打开Xcode
2. 选择 `apple-watch/TaishanglaojunWatch.xcodeproj`
3. 连接iPhone和Apple Watch
4. 选择Watch目标并运行

### Wear OS应用

1. 打开Android Studio
2. 导入 `wear-os` 项目
3. 连接Android手机和Wear OS设备
4. 选择Wear模块并运行

## 技术架构

### 数据同步
- **实时同步**：WebSocket连接保持数据实时性
- **离线缓存**：本地数据库存储关键信息
- **冲突解决**：智能合并策略处理数据冲突
- **增量更新**：只同步变更数据，节省电量

### 用户界面
- **原生设计**：遵循各平台设计规范
- **响应式布局**：适配不同尺寸屏幕
- **无障碍支持**：完整的辅助功能支持
- **暗黑模式**：支持系统主题切换

### 性能优化
- **电量管理**：智能后台任务调度
- **内存优化**：高效的数据结构和缓存策略
- **网络优化**：请求合并和智能重试
- **渲染优化**：流畅的动画和交互

## API集成

手表端应用将与后端API进行集成：
- **任务数据同步**：获取和更新任务信息
- **用户认证**：安全的身份验证机制
- **实时通知**：推送服务集成
- **数据分析**：使用行为数据收集

## 隐私和安全

- **数据加密**：本地存储和网络传输加密
- **权限管理**：最小化权限申请
- **隐私合规**：符合各地区隐私法规
- **安全通信**：HTTPS和证书验证

## 测试策略

- **单元测试**：核心业务逻辑测试
- **UI测试**：用户界面自动化测试
- **集成测试**：与手机端应用联调测试
- **性能测试**：电量消耗和响应时间测试

## 部署和发布

- **Apple Watch**：通过App Store Connect发布
- **Wear OS**：通过Google Play Console发布
- **企业分发**：支持企业内部分发
- **测试版本**：TestFlight和内部测试轨道

## 贡献指南

1. Fork项目仓库
2. 创建功能分支
3. 提交代码变更
4. 创建Pull Request
5. 代码审查和合并

## 许可证

Apache License 2.0

## 联系方式

- 项目维护者：太上老君开发团队
- 技术支持：watch-support@taishanglaojun.com
- 问题反馈：GitHub Issues