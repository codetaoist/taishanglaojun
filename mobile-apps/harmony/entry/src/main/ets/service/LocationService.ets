import geoLocationManager from '@ohos.geoLocationManager';
import { BusinessError } from '@ohos.base';
import { LocationPoint } from '../model/LocationPoint';
import hilog from '@ohos.hilog';

/**
 * 位置服务状态
 */
export enum LocationServiceStatus {
  STOPPED = 'stopped',
  STARTING = 'starting',
  RUNNING = 'running',
  PAUSED = 'paused',
  ERROR = 'error'
}

/**
 * 位置服务错误类型
 */
export enum LocationError {
  PERMISSION_DENIED = 'permission_denied',
  LOCATION_DISABLED = 'location_disabled',
  TIMEOUT = 'timeout',
  UNAVAILABLE = 'unavailable',
  ACCURACY_TOO_LOW = 'accuracy_too_low',
  UNKNOWN = 'unknown'
}

/**
 * 位置更新回调接口
 */
export interface LocationUpdateCallback {
  onLocationUpdate: (location: LocationPoint) => void;
  onError: (error: LocationError, message: string) => void;
  onStatusChange: (status: LocationServiceStatus) => void;
}

/**
 * 位置服务配置
 */
export interface LocationServiceConfig {
  minUpdateInterval: number; // 最小更新间隔（毫秒）
  minDistance: number; // 最小距离变化（米）
  maxAccuracy: number; // 最大精度阈值（米）
  enableHighAccuracy: boolean; // 启用高精度模式
  timeout: number; // 超时时间（毫秒）
}

/**
 * 位置服务类
 */
export class LocationService {
  private static instance: LocationService;
  private status: LocationServiceStatus = LocationServiceStatus.STOPPED;
  private callback?: LocationUpdateCallback;
  private config: LocationServiceConfig;
  private requestId?: number;
  private lastLocation?: LocationPoint;
  private lastUpdateTime: number = 0;

  private constructor() {
    this.config = {
      minUpdateInterval: 1000, // 1秒
      minDistance: 5, // 5米
      maxAccuracy: 50, // 50米
      enableHighAccuracy: true,
      timeout: 30000 // 30秒
    };
  }

  /**
   * 获取单例实例
   */
  static getInstance(): LocationService {
    if (!LocationService.instance) {
      LocationService.instance = new LocationService();
    }
    return LocationService.instance;
  }

  /**
   * 检查位置权限
   */
  async checkLocationPermission(): Promise<boolean> {
    try {
      // 检查精确位置权限
      const precisePermission = await geoLocationManager.checkLocationPermission();
      
      // 检查后台位置权限
      const backgroundPermission = await geoLocationManager.checkLocationPermission();
      
      return precisePermission && backgroundPermission;
    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Check permission error: %{public}s', JSON.stringify(error));
      return false;
    }
  }

  /**
   * 请求位置权限
   */
  async requestLocationPermission(): Promise<boolean> {
    try {
      // 在实际应用中，需要使用abilityAccessCtrl来请求权限
      // 这里简化处理，假设权限已授予
      return await this.checkLocationPermission();
    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Request permission error: %{public}s', JSON.stringify(error));
      return false;
    }
  }

  /**
   * 检查位置服务是否可用
   */
  async isLocationEnabled(): Promise<boolean> {
    try {
      return await geoLocationManager.isLocationEnabled();
    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Check location enabled error: %{public}s', JSON.stringify(error));
      return false;
    }
  }

  /**
   * 获取当前状态
   */
  getStatus(): LocationServiceStatus {
    return this.status;
  }

  /**
   * 获取最后一次位置
   */
  getLastLocation(): LocationPoint | undefined {
    return this.lastLocation;
  }

  /**
   * 更新配置
   */
  updateConfig(config: Partial<LocationServiceConfig>): void {
    this.config = { ...this.config, ...config };
  }

  /**
   * 获取配置
   */
  getConfig(): LocationServiceConfig {
    return { ...this.config };
  }

  /**
   * 开始位置跟踪
   */
  async startTracking(callback: LocationUpdateCallback): Promise<boolean> {
    try {
      // 检查权限
      const hasPermission = await this.checkLocationPermission();
      if (!hasPermission) {
        callback.onError(LocationError.PERMISSION_DENIED, '位置权限未授予');
        return false;
      }

      // 检查位置服务是否启用
      const isEnabled = await this.isLocationEnabled();
      if (!isEnabled) {
        callback.onError(LocationError.LOCATION_DISABLED, '位置服务未启用');
        return false;
      }

      this.callback = callback;
      this.setStatus(LocationServiceStatus.STARTING);

      // 配置位置请求参数
      const requestInfo: geoLocationManager.LocationRequest = {
        priority: this.config.enableHighAccuracy ? 
          geoLocationManager.LocationRequestPriority.ACCURACY : 
          geoLocationManager.LocationRequestPriority.LOW_POWER,
        scenario: geoLocationManager.LocationRequestScenario.NAVIGATION,
        timeInterval: this.config.minUpdateInterval / 1000, // 转换为秒
        distanceInterval: this.config.minDistance,
        maxAccuracy: this.config.maxAccuracy
      };

      // 开始位置更新
      await geoLocationManager.on('locationChange', requestInfo, (location: geoLocationManager.Location) => {
        this.handleLocationUpdate(location);
      });

      this.setStatus(LocationServiceStatus.RUNNING);
      hilog.info(0x0000, 'LocationService', 'Location tracking started');
      return true;

    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Start tracking error: %{public}s', JSON.stringify(error));
      this.handleError(LocationError.UNKNOWN, `启动位置跟踪失败: ${error}`);
      return false;
    }
  }

  /**
   * 停止位置跟踪
   */
  async stopTracking(): Promise<void> {
    try {
      if (this.status === LocationServiceStatus.RUNNING || this.status === LocationServiceStatus.PAUSED) {
        await geoLocationManager.off('locationChange');
        this.setStatus(LocationServiceStatus.STOPPED);
        this.callback = undefined;
        this.lastLocation = undefined;
        this.lastUpdateTime = 0;
        hilog.info(0x0000, 'LocationService', 'Location tracking stopped');
      }
    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Stop tracking error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 暂停位置跟踪
   */
  pauseTracking(): void {
    if (this.status === LocationServiceStatus.RUNNING) {
      this.setStatus(LocationServiceStatus.PAUSED);
      hilog.info(0x0000, 'LocationService', 'Location tracking paused');
    }
  }

  /**
   * 恢复位置跟踪
   */
  resumeTracking(): void {
    if (this.status === LocationServiceStatus.PAUSED) {
      this.setStatus(LocationServiceStatus.RUNNING);
      hilog.info(0x0000, 'LocationService', 'Location tracking resumed');
    }
  }

  /**
   * 获取单次位置
   */
  async getCurrentLocation(): Promise<LocationPoint | null> {
    try {
      const hasPermission = await this.checkLocationPermission();
      if (!hasPermission) {
        throw new Error('位置权限未授予');
      }

      const requestInfo: geoLocationManager.CurrentLocationRequest = {
        priority: this.config.enableHighAccuracy ? 
          geoLocationManager.LocationRequestPriority.ACCURACY : 
          geoLocationManager.LocationRequestPriority.LOW_POWER,
        scenario: geoLocationManager.LocationRequestScenario.NAVIGATION,
        maxAccuracy: this.config.maxAccuracy,
        timeoutMs: this.config.timeout
      };

      const location = await geoLocationManager.getCurrentLocation(requestInfo);
      return this.convertToLocationPoint(location);

    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Get current location error: %{public}s', JSON.stringify(error));
      return null;
    }
  }

  /**
   * 处理位置更新
   */
  private handleLocationUpdate(location: geoLocationManager.Location): void {
    if (this.status !== LocationServiceStatus.RUNNING) {
      return;
    }

    const now = Date.now();
    
    // 检查更新间隔
    if (now - this.lastUpdateTime < this.config.minUpdateInterval) {
      return;
    }

    // 转换位置数据
    const locationPoint = this.convertToLocationPoint(location);
    
    // 验证位置精度
    if (locationPoint.accuracy && locationPoint.accuracy > this.config.maxAccuracy) {
      this.handleError(LocationError.ACCURACY_TOO_LOW, `位置精度过低: ${locationPoint.accuracy}m`);
      return;
    }

    // 检查距离变化
    if (this.lastLocation) {
      const distance = locationPoint.distanceTo(this.lastLocation);
      if (distance < this.config.minDistance) {
        return;
      }
    }

    this.lastLocation = locationPoint;
    this.lastUpdateTime = now;

    // 通知回调
    if (this.callback) {
      this.callback.onLocationUpdate(locationPoint);
    }
  }

  /**
   * 转换位置数据
   */
  private convertToLocationPoint(location: geoLocationManager.Location): LocationPoint {
    return new LocationPoint(
      location.latitude,
      location.longitude,
      location.timeStamp || Date.now(),
      {
        accuracy: location.accuracy,
        altitude: location.altitude,
        speed: location.speed,
        bearing: location.direction,
        provider: 'gps'
      }
    );
  }

  /**
   * 设置状态
   */
  private setStatus(status: LocationServiceStatus): void {
    if (this.status !== status) {
      this.status = status;
      if (this.callback) {
        this.callback.onStatusChange(status);
      }
    }
  }

  /**
   * 处理错误
   */
  private handleError(error: LocationError, message: string): void {
    this.setStatus(LocationServiceStatus.ERROR);
    if (this.callback) {
      this.callback.onError(error, message);
    }
    hilog.error(0x0000, 'LocationService', 'Location error: %{public}s - %{public}s', error, message);
  }

  /**
   * 获取位置提供者信息
   */
  async getLocationProviders(): Promise<string[]> {
    try {
      // 在实际应用中，可以获取可用的位置提供者
      return ['gps', 'network', 'passive'];
    } catch (error) {
      hilog.error(0x0000, 'LocationService', 'Get providers error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 清理资源
   */
  async cleanup(): Promise<void> {
    await this.stopTracking();
  }
}