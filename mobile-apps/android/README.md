# 太上老君 Android 应用

## 项目概述

基于Kotlin开发的Android原生应用，实现GPS位置追踪和轨迹记录功能。

## 技术栈

- **语言**: Kotlin 1.9+
- **最低版本**: Android 7.0 (API 24)
- **目标版本**: Android 14 (API 34)
- **架构**: MVVM + Repository Pattern
- **框架**:
  - Jetpack Compose (UI)
  - Room (本地数据库)
  - Retrofit (网络请求)
  - Hilt (依赖注入)
  - Coroutines (异步处理)
  - Location Services (位置服务)

## 核心功能

### 🗺️ 位置服务
- 融合位置提供器 (Fused Location Provider)
- 前台服务保持后台运行
- 位置精度和频率控制
- 电池优化策略

### 📊 轨迹管理
- 实时轨迹记录
- 轨迹数据可视化
- 历史记录查询
- 轨迹数据导出

### 🔒 安全特性
- 位置权限动态申请
- 数据加密存储
- HTTPS安全传输
- 用户隐私控制

## 项目结构

```
app/
├── src/
│   ├── main/
│   │   ├── java/com/taishanglaojun/tracker/
│   │   │   ├── data/
│   │   │   │   ├── local/
│   │   │   │   │   ├── dao/
│   │   │   │   │   ├── database/
│   │   │   │   │   └── entities/
│   │   │   │   ├── remote/
│   │   │   │   │   ├── api/
│   │   │   │   │   └── dto/
│   │   │   │   └── repository/
│   │   │   ├── domain/
│   │   │   │   ├── model/
│   │   │   │   ├── repository/
│   │   │   │   └── usecase/
│   │   │   ├── presentation/
│   │   │   │   ├── ui/
│   │   │   │   │   ├── components/
│   │   │   │   │   ├── screens/
│   │   │   │   │   └── theme/
│   │   │   │   └── viewmodel/
│   │   │   ├── service/
│   │   │   │   ├── LocationService.kt
│   │   │   │   └── LocationForegroundService.kt
│   │   │   ├── utils/
│   │   │   │   ├── Constants.kt
│   │   │   │   ├── Extensions.kt
│   │   │   │   └── PermissionUtils.kt
│   │   │   └── di/
│   │   │       ├── DatabaseModule.kt
│   │   │       ├── NetworkModule.kt
│   │   │       └── ServiceModule.kt
│   │   ├── res/
│   │   │   ├── layout/
│   │   │   ├── values/
│   │   │   ├── drawable/
│   │   │   └── xml/
│   │   └── AndroidManifest.xml
│   ├── test/
│   └── androidTest/
├── build.gradle.kts
└── proguard-rules.pro
```

## 开发环境

### 要求
- Android Studio 2023.1+
- JDK 17+
- Android SDK 34
- Gradle 8.0+

### 依赖配置

```kotlin
dependencies {
    // Jetpack Compose
    implementation "androidx.compose.ui:ui:$compose_version"
    implementation "androidx.compose.material3:material3:$material3_version"
    
    // Location Services
    implementation "com.google.android.gms:play-services-location:21.0.1"
    implementation "com.google.android.gms:play-services-maps:18.2.0"
    
    // Room Database
    implementation "androidx.room:room-runtime:$room_version"
    implementation "androidx.room:room-ktx:$room_version"
    kapt "androidx.room:room-compiler:$room_version"
    
    // Network
    implementation "com.squareup.retrofit2:retrofit:$retrofit_version"
    implementation "com.squareup.okhttp3:logging-interceptor:$okhttp_version"
    
    // Dependency Injection
    implementation "com.google.dagger:hilt-android:$hilt_version"
    kapt "com.google.dagger:hilt-compiler:$hilt_version"
    
    // Coroutines
    implementation "org.jetbrains.kotlinx:kotlinx-coroutines-android:$coroutines_version"
}
```

## 权限配置

在 `AndroidManifest.xml` 中添加：

```xml
<!-- 位置权限 -->
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_BACKGROUND_LOCATION" />

<!-- 前台服务权限 -->
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
<uses-permission android:name="android.permission.FOREGROUND_SERVICE_LOCATION" />

<!-- 网络权限 -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

<!-- 存储权限 -->
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />

<!-- 前台服务声明 -->
<service
    android:name=".service.LocationForegroundService"
    android:foregroundServiceType="location"
    android:exported="false" />
```

## 核心实现

### 位置服务
```kotlin
@AndroidEntryPoint
class LocationForegroundService : Service() {
    
    @Inject
    lateinit var locationRepository: LocationRepository
    
    private val fusedLocationClient by lazy {
        LocationServices.getFusedLocationProviderClient(this)
    }
    
    private val locationCallback = object : LocationCallback() {
        override fun onLocationResult(result: LocationResult) {
            result.locations.forEach { location ->
                // 处理位置更新
                handleLocationUpdate(location)
            }
        }
    }
    
    private fun startLocationUpdates() {
        val locationRequest = LocationRequest.Builder(
            Priority.PRIORITY_HIGH_ACCURACY,
            10000L // 10秒更新一次
        ).build()
        
        fusedLocationClient.requestLocationUpdates(
            locationRequest,
            locationCallback,
            Looper.getMainLooper()
        )
    }
}
```

### 数据模型
```kotlin
@Entity(tableName = "location_points")
data class LocationPoint(
    @PrimaryKey val id: String = UUID.randomUUID().toString(),
    val latitude: Double,
    val longitude: Double,
    val timestamp: Long,
    val accuracy: Float,
    val speed: Float?,
    val bearing: Float?,
    val trajectoryId: String
)

@Entity(tableName = "trajectories")
data class Trajectory(
    @PrimaryKey val id: String = UUID.randomUUID().toString(),
    val name: String,
    val startTime: Long,
    val endTime: Long?,
    val distance: Double = 0.0,
    val duration: Long = 0L
)
```

## API集成

### 网络服务
```kotlin
interface LocationApi {
    @POST("api/locations")
    suspend fun uploadLocations(@Body locations: List<LocationDto>): Response<Unit>
    
    @GET("api/trajectories")
    suspend fun getTrajectories(): Response<List<TrajectoryDto>>
    
    @GET("api/trajectories/{id}")
    suspend fun getTrajectory(@Path("id") id: String): Response<TrajectoryDto>
}
```

## 隐私合规

### 权限申请
```kotlin
class PermissionManager {
    fun requestLocationPermissions(activity: Activity) {
        when {
            hasLocationPermission() -> {
                // 已有权限，继续操作
            }
            shouldShowRequestPermissionRationale() -> {
                // 显示权限说明
                showPermissionRationale()
            }
            else -> {
                // 请求权限
                requestPermissions()
            }
        }
    }
}
```

## 测试

### 单元测试
```bash
./gradlew test
```

### 集成测试
```bash
./gradlew connectedAndroidTest
```

## 构建和发布

### Debug构建
```bash
./gradlew assembleDebug
```

### Release构建
```bash
./gradlew assembleRelease
```

### 签名配置
在 `build.gradle.kts` 中配置：
```kotlin
android {
    signingConfigs {
        create("release") {
            storeFile = file("keystore/release.keystore")
            storePassword = "your_store_password"
            keyAlias = "your_key_alias"
            keyPassword = "your_key_password"
        }
    }
}
```