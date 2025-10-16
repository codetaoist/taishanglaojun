# 太上老君 iOS 应用

## 项目概述

基于Swift开发的iOS原生应用，实现GPS位置追踪和轨迹记录功能。

## 技术栈

- **语言**: Swift 5.9+
- **最低版本**: iOS 14.0
- **框架**: 
  - Core Location (位置服务)
  - MapKit (地图显示)
  - Core Data (本地存储)
  - CryptoKit (数据加密)
  - Combine (响应式编程)

## 核心功能

### 🗺️ 位置服务
- 实时GPS定位
- 后台位置更新
- 位置精度控制
- 电池优化

### 📊 轨迹管理
- 轨迹实时记录
- 历史轨迹查询
- 轨迹可视化
- 数据导出

### 🔒 隐私安全
- 位置权限管理
- 数据加密存储
- 安全网络传输
- 用户隐私控制

## 项目结构

```
TaishanglaojunTracker/
├── TaishanglaojunTracker/
│   ├── App/
│   │   ├── AppDelegate.swift
│   │   ├── SceneDelegate.swift
│   │   └── Info.plist
│   ├── Models/
│   │   ├── LocationPoint.swift
│   │   ├── Trajectory.swift
│   │   └── User.swift
│   ├── Services/
│   │   ├── LocationService.swift
│   │   ├── DataService.swift
│   │   ├── NetworkService.swift
│   │   └── CryptoService.swift
│   ├── ViewModels/
│   │   ├── LocationViewModel.swift
│   │   ├── TrajectoryViewModel.swift
│   │   └── SettingsViewModel.swift
│   ├── Views/
│   │   ├── ContentView.swift
│   │   ├── MapView.swift
│   │   ├── TrajectoryListView.swift
│   │   ├── TrajectoryDetailView.swift
│   │   └── SettingsView.swift
│   └── Resources/
│       ├── Assets.xcassets
│       └── Localizable.strings
├── TaishanglaojunTrackerTests/
└── TaishanglaojunTrackerUITests/
```

## 开发环境

### 要求
- macOS 13.0+
- Xcode 15.0+
- iOS 14.0+ 设备或模拟器

### 安装步骤
1. 打开 `TaishanglaojunTracker.xcodeproj`
2. 选择开发团队和Bundle ID
3. 配置位置权限描述
4. 运行项目

## 权限配置

在 `Info.plist` 中添加以下权限：

```xml
<key>NSLocationAlwaysAndWhenInUseUsageDescription</key>
<string>应用需要访问您的位置来记录移动轨迹</string>
<key>NSLocationWhenInUseUsageDescription</key>
<string>应用需要访问您的位置来记录移动轨迹</string>
<key>UIBackgroundModes</key>
<array>
    <string>location</string>
    <string>background-processing</string>
</array>
```

## API集成

### 后端接口
- `POST /api/locations` - 上传位置数据
- `GET /api/trajectories` - 获取轨迹列表
- `GET /api/trajectories/{id}` - 获取轨迹详情
- `DELETE /api/trajectories/{id}` - 删除轨迹

### 数据格式
```json
{
  "latitude": 39.9042,
  "longitude": 116.4074,
  "timestamp": "2024-01-01T12:00:00Z",
  "accuracy": 5.0,
  "speed": 1.2,
  "heading": 45.0
}
```

## 隐私合规

- 遵循Apple隐私指南
- 支持位置权限精确控制
- 提供数据删除功能
- 透明的隐私政策

## 测试

### 单元测试
```bash
# 运行单元测试
xcodebuild test -scheme TaishanglaojunTracker -destination 'platform=iOS Simulator,name=iPhone 15'
```

### UI测试
```bash
# 运行UI测试
xcodebuild test -scheme TaishanglaojunTrackerUITests -destination 'platform=iOS Simulator,name=iPhone 15'
```

## 发布

### App Store发布
1. 配置发布证书
2. 更新版本号
3. 创建Archive
4. 上传到App Store Connect
5. 提交审核

### 企业发布
1. 配置企业证书
2. 创建Ad Hoc Archive
3. 分发给测试用户