# 太上老君轨迹跟踪应用 - HarmonyOS版本

## 项目简介

这是太上老君轨迹跟踪应用的HarmonyOS版本，提供GPS位置跟踪、轨迹记录和数据管理功能。

## 功能特性

- **实时位置跟踪**: 使用GPS获取精确的位置信息
- **轨迹记录**: 记录和保存用户的移动轨迹
- **数据管理**: 本地存储轨迹数据，支持查看、删除等操作
- **后台服务**: 支持后台持续跟踪位置
- **权限管理**: 合规的位置权限申请和管理
- **用户界面**: 直观的Material Design风格界面

## 技术架构

### 核心组件

1. **数据模型**
   - `LocationPoint.ets`: 位置点数据模型
   - `Trajectory.ets`: 轨迹数据模型

2. **服务层**
   - `LocationService.ets`: 位置服务，处理GPS定位
   - `LocationTrackingService.ets`: 后台跟踪服务
   - `DataService.ets`: 数据存储服务

3. **视图模型**
   - `LocationViewModel.ets`: 位置跟踪业务逻辑管理

4. **用户界面**
   - `Index.ets`: 主页面，包含跟踪、历史、统计、设置等功能

### 权限要求

- `ohos.permission.APPROXIMATELY_LOCATION`: 大致位置权限
- `ohos.permission.LOCATION`: 精确位置权限
- `ohos.permission.LOCATION_IN_BACKGROUND`: 后台位置权限
- `ohos.permission.INTERNET`: 网络访问权限
- `ohos.permission.GET_NETWORK_INFO`: 网络信息权限
- `ohos.permission.KEEP_BACKGROUND_RUNNING`: 后台运行权限
- `ohos.permission.NOTIFICATION_CONTROLLER`: 通知权限

## 开发环境

- HarmonyOS SDK API Level 10+
- DevEco Studio 4.0+
- ArkTS/TypeScript

## 构建和运行

1. 使用DevEco Studio打开项目
2. 配置签名证书
3. 连接HarmonyOS设备或启动模拟器
4. 点击运行按钮构建和安装应用

## 项目结构

```
harmony/
├── AppScope/                 # 应用级配置
│   └── app.json5            # 应用配置文件
├── entry/                   # 主模块
│   ├── src/main/
│   │   ├── ets/
│   │   │   ├── entryability/
│   │   │   │   └── EntryAbility.ets
│   │   │   ├── model/       # 数据模型
│   │   │   │   ├── LocationPoint.ets
│   │   │   │   └── Trajectory.ets
│   │   │   ├── service/     # 服务层
│   │   │   │   ├── LocationService.ets
│   │   │   │   ├── LocationTrackingService.ets
│   │   │   │   └── DataService.ets
│   │   │   ├── viewmodel/   # 视图模型
│   │   │   │   └── LocationViewModel.ets
│   │   │   └── pages/       # 页面
│   │   │       └── Index.ets
│   │   └── module.json5     # 模块配置
│   ├── build-profile.json5  # 构建配置
│   ├── hvigorfile.ts       # 构建脚本
│   └── oh-package.json5    # 包配置
├── build-profile.json5      # 项目构建配置
├── hvigorfile.ts           # 项目构建脚本
└── oh-package.json5        # 项目包配置
```

## 注意事项

1. 确保设备支持GPS功能
2. 首次运行需要授予位置权限
3. 后台跟踪需要用户明确同意
4. 建议在真机上测试GPS功能

## 许可证

Apache License 2.0
- **目标版本**: HarmonyOS API 11
- **架构**: MVVM + 分层架构
- **框架**:
  - ArkUI (声明式UI)
  - 位置服务 (@ohos.geoLocationManager)
  - 关系型数据库 (@ohos.data.relationalStore)
  - 网络请求 (@ohos.net.http)
  - 后台任务 (@ohos.backgroundTaskManager)

## 核心功能

### 🗺️ 位置服务
- 高精度GPS定位
- 连续定位服务
- 后台任务管理
- 位置缓存策略

### 📊 轨迹管理
- 实时轨迹绘制
- 轨迹数据存储
- 历史轨迹查询
- 轨迹统计分析

### 🔒 安全特性
- 位置权限管理
- 数据加密存储
- 安全网络通信
- 隐私保护机制

## 项目结构

```
entry/
├── src/
│   ├── main/
│   │   ├── ets/
│   │   │   ├── common/
│   │   │   │   ├── constants/
│   │   │   │   ├── utils/
│   │   │   │   └── types/
│   │   │   ├── data/
│   │   │   │   ├── local/
│   │   │   │   │   ├── database/
│   │   │   │   │   └── preferences/
│   │   │   │   ├── remote/
│   │   │   │   │   ├── api/
│   │   │   │   │   └── dto/
│   │   │   │   └── repository/
│   │   │   ├── domain/
│   │   │   │   ├── model/
│   │   │   │   ├── repository/
│   │   │   │   └── usecase/
│   │   │   ├── presentation/
│   │   │   │   ├── pages/
│   │   │   │   ├── components/
│   │   │   │   └── viewmodel/
│   │   │   ├── service/
│   │   │   │   ├── LocationService.ets
│   │   │   │   └── BackgroundService.ets
│   │   │   └── entryability/
│   │   │       └── EntryAbility.ets
│   │   ├── resources/
│   │   │   ├── base/
│   │   │   │   ├── element/
│   │   │   │   ├── media/
│   │   │   │   └── profile/
│   │   │   └── rawfile/
│   │   └── module.json5
│   └── ohosTest/
├── build-profile.json5
├── hvigorfile.ts
└── oh-package.json5
```

## 开发环境

### 要求
- DevEco Studio 4.0+
- HarmonyOS SDK API 9+
- Node.js 16+

### 项目配置

在 `oh-package.json5` 中配置依赖：
```json
{
  "dependencies": {
    "@ohos/hypium": "1.0.6"
  },
  "devDependencies": {
    "@ohos/hvigor-ohos-plugin": "4.0.2",
    "@ohos/hvigor": "4.0.2"
  }
}
```

## 权限配置

在 `module.json5` 中添加权限：

```json
{
  "module": {
    "requestPermissions": [
      {
        "name": "ohos.permission.APPROXIMATELY_LOCATION",
        "reason": "$string:location_permission_reason",
        "usedScene": {
          "abilities": ["EntryAbility"],
          "when": "inuse"
        }
      },
      {
        "name": "ohos.permission.LOCATION",
        "reason": "$string:location_permission_reason",
        "usedScene": {
          "abilities": ["EntryAbility"],
          "when": "always"
        }
      },
      {
        "name": "ohos.permission.LOCATION_IN_BACKGROUND",
        "reason": "$string:background_location_reason",
        "usedScene": {
          "abilities": ["EntryAbility"],
          "when": "always"
        }
      },
      {
        "name": "ohos.permission.INTERNET",
        "reason": "$string:internet_permission_reason"
      },
      {
        "name": "ohos.permission.KEEP_BACKGROUND_RUNNING",
        "reason": "$string:background_running_reason"
      }
    ]
  }
}
```

## 核心实现

### 位置服务
```typescript
import geoLocationManager from '@ohos.geoLocationManager';
import backgroundTaskManager from '@ohos.backgroundTaskManager';

export class LocationService {
  private locationRequest: geoLocationManager.LocationRequest = {
    priority: geoLocationManager.LocationRequestPriority.FIRST_FIX,
    scenario: geoLocationManager.LocationRequestScenario.TRAJECTORY_TRACKING,
    timeInterval: 10, // 10秒更新一次
    distanceInterval: 10, // 10米更新一次
    maxAccuracy: 0
  };

  async startLocationTracking(): Promise<void> {
    try {
      // 申请后台任务
      const bgTaskId = await backgroundTaskManager.requestSuspendDelay(
        backgroundTaskManager.DelaySuspendInfo.NORMAL,
        () => {
          console.info('Background task expired');
        }
      );

      // 开始位置更新
      geoLocationManager.on('locationChange', this.locationRequest, (location) => {
        this.handleLocationUpdate(location);
      });

    } catch (error) {
      console.error('Failed to start location tracking:', error);
    }
  }

  private handleLocationUpdate(location: geoLocationManager.Location): void {
    const locationPoint = {
      latitude: location.latitude,
      longitude: location.longitude,
      timestamp: Date.now(),
      accuracy: location.accuracy,
      speed: location.speed,
      bearing: location.direction
    };
    
    // 保存位置数据
    this.saveLocationPoint(locationPoint);
  }
}
```

### 数据模型
```typescript
export interface LocationPoint {
  id: string;
  latitude: number;
  longitude: number;
  timestamp: number;
  accuracy: number;
  speed?: number;
  bearing?: number;
  trajectoryId: string;
}

export interface Trajectory {
  id: string;
  name: string;
  startTime: number;
  endTime?: number;
  distance: number;
  duration: number;
  points: LocationPoint[];
}
```

### 数据库操作
```typescript
import relationalStore from '@ohos.data.relationalStore';

export class DatabaseManager {
  private rdbStore: relationalStore.RdbStore | null = null;

  async initDatabase(): Promise<void> {
    const config: relationalStore.StoreConfig = {
      name: 'TaishanglaojunTracker.db',
      securityLevel: relationalStore.SecurityLevel.S1
    };

    this.rdbStore = await relationalStore.getRdbStore(getContext(), config);
    await this.createTables();
  }

  private async createTables(): Promise<void> {
    const createLocationTable = `
      CREATE TABLE IF NOT EXISTS location_points (
        id TEXT PRIMARY KEY,
        latitude REAL NOT NULL,
        longitude REAL NOT NULL,
        timestamp INTEGER NOT NULL,
        accuracy REAL NOT NULL,
        speed REAL,
        bearing REAL,
        trajectory_id TEXT NOT NULL
      )
    `;

    const createTrajectoryTable = `
      CREATE TABLE IF NOT EXISTS trajectories (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        start_time INTEGER NOT NULL,
        end_time INTEGER,
        distance REAL DEFAULT 0,
        duration INTEGER DEFAULT 0
      )
    `;

    await this.rdbStore?.executeSql(createLocationTable);
    await this.rdbStore?.executeSql(createTrajectoryTable);
  }
}
```

### UI组件
```typescript
@Component
export struct TrajectoryMapView {
  @State trajectoryPoints: LocationPoint[] = [];
  @State mapController: MapController = new MapController();

  build() {
    Column() {
      // 地图组件
      Map(this.mapController) {
        // 绘制轨迹线
        ForEach(this.trajectoryPoints, (point: LocationPoint, index: number) => {
          if (index > 0) {
            Polyline({
              points: [
                { latitude: this.trajectoryPoints[index - 1].latitude, longitude: this.trajectoryPoints[index - 1].longitude },
                { latitude: point.latitude, longitude: point.longitude }
              ]
            })
              .stroke(Color.Blue)
              .strokeWidth(3)
          }
        })

        // 标记起点和终点
        if (this.trajectoryPoints.length > 0) {
          Marker({
            position: {
              latitude: this.trajectoryPoints[0].latitude,
              longitude: this.trajectoryPoints[0].longitude
            }
          })
            .markIcon('/common/images/start_marker.png')
        }
      }
      .width('100%')
      .height('70%')

      // 轨迹信息
      this.TrajectoryInfo()
    }
  }

  @Builder
  TrajectoryInfo() {
    Row() {
      Text(`距离: ${this.calculateDistance()}km`)
        .fontSize(16)
        .margin({ right: 20 })
      
      Text(`时长: ${this.formatDuration()}`)
        .fontSize(16)
    }
    .width('100%')
    .justifyContent(FlexAlign.SpaceAround)
    .padding(16)
  }
}
```

## API集成

### 网络服务
```typescript
import http from '@ohos.net.http';

export class ApiService {
  private baseUrl = 'https://api.taishanglaojun.com';

  async uploadLocations(locations: LocationPoint[]): Promise<boolean> {
    try {
      const httpRequest = http.createHttp();
      const response = await httpRequest.request(`${this.baseUrl}/api/locations`, {
        method: http.RequestMethod.POST,
        header: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${await this.getAuthToken()}`
        },
        extraData: JSON.stringify(locations)
      });

      return response.responseCode === 200;
    } catch (error) {
      console.error('Failed to upload locations:', error);
      return false;
    }
  }
}
```

## 隐私合规

### 权限申请
```typescript
import abilityAccessCtrl from '@ohos.abilityAccessCtrl';

export class PermissionManager {
  async requestLocationPermissions(): Promise<boolean> {
    try {
      const atManager = abilityAccessCtrl.createAtManager();
      const permissions = [
        'ohos.permission.LOCATION',
        'ohos.permission.APPROXIMATELY_LOCATION',
        'ohos.permission.LOCATION_IN_BACKGROUND'
      ];

      const grantStatus = await atManager.requestPermissionsFromUser(
        getContext(),
        permissions
      );

      return grantStatus.authResults.every(result => 
        result === abilityAccessCtrl.GrantStatus.PERMISSION_GRANTED
      );
    } catch (error) {
      console.error('Permission request failed:', error);
      return false;
    }
  }
}
```

## 构建和发布

### 开发构建
```bash
hvigorw assembleHap --mode module -p product=default
```

### 发布构建
```bash
hvigorw assembleHap --mode module -p product=default --release
```

### 应用签名
在 `build-profile.json5` 中配置签名：
```json
{
  "app": {
    "signingConfigs": [
      {
        "name": "default",
        "type": "HarmonyOS",
        "material": {
          "certpath": "cert/release.p7b",
          "storePassword": "your_store_password",
          "keyAlias": "your_key_alias",
          "keyPassword": "your_key_password",
          "profile": "cert/release.p7b",
          "signAlg": "SHA256withECDSA",
          "storeFile": "cert/release.p12"
        }
      }
    ]
  }
}
```

## 测试

### 单元测试
```bash
hvigorw clean test --mode module -p product=default
```

### 集成测试
```bash
hvigorw connectedTest --mode module -p product=default
```