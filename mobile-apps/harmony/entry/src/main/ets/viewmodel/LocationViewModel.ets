import { LocationService, LocationServiceStatus, LocationError, LocationUpdateCallback } from '../service/LocationService';
import { DataService } from '../service/DataService';
import { LocationPoint } from '../model/LocationPoint';
import { Trajectory, TrajectoryStatus } from '../model/Trajectory';
import hilog from '@ohos.hilog';
import common from '@ohos.app.ability.common';

/**
 * 位置视图模型
 */
export class LocationViewModel implements LocationUpdateCallback {
  private locationService: LocationService;
  private dataService: DataService;
  private context?: common.UIAbilityContext;

  // 状态属性
  private _isTracking: boolean = false;
  private _currentLocation?: LocationPoint;
  private _currentTrajectory?: Trajectory;
  private _trajectories: Trajectory[] = [];
  private _locationServiceStatus: LocationServiceStatus = LocationServiceStatus.STOPPED;
  private _hasLocationPermission: boolean = false;
  private _errorMessage: string = '';
  private _isLoading: boolean = false;

  // 状态变化回调
  private stateChangeCallbacks: (() => void)[] = [];

  constructor() {
    this.locationService = LocationService.getInstance();
    this.dataService = DataService.getInstance();
  }

  /**
   * 初始化
   */
  async initialize(context: common.UIAbilityContext): Promise<void> {
    try {
      this.context = context;
      this._isLoading = true;
      this.notifyStateChange();

      // 初始化数据服务
      await this.dataService.initialize(context);

      // 检查位置权限
      await this.checkLocationPermission();

      // 加载轨迹数据
      await this.loadTrajectories();

      this._isLoading = false;
      this.notifyStateChange();

      hilog.info(0x0000, 'LocationViewModel', 'ViewModel initialized');
    } catch (error) {
      this._isLoading = false;
      this._errorMessage = `初始化失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Initialize error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 添加状态变化监听器
   */
  addStateChangeListener(callback: () => void): void {
    this.stateChangeCallbacks.push(callback);
  }

  /**
   * 移除状态变化监听器
   */
  removeStateChangeListener(callback: () => void): void {
    const index = this.stateChangeCallbacks.indexOf(callback);
    if (index > -1) {
      this.stateChangeCallbacks.splice(index, 1);
    }
  }

  /**
   * 通知状态变化
   */
  private notifyStateChange(): void {
    this.stateChangeCallbacks.forEach(callback => callback());
  }

  // Getter 方法
  get isTracking(): boolean {
    return this._isTracking;
  }

  get currentLocation(): LocationPoint | undefined {
    return this._currentLocation;
  }

  get currentTrajectory(): Trajectory | undefined {
    return this._currentTrajectory;
  }

  get trajectories(): Trajectory[] {
    return [...this._trajectories];
  }

  get locationServiceStatus(): LocationServiceStatus {
    return this._locationServiceStatus;
  }

  get hasLocationPermission(): boolean {
    return this._hasLocationPermission;
  }

  get errorMessage(): string {
    return this._errorMessage;
  }

  get isLoading(): boolean {
    return this._isLoading;
  }

  /**
   * 检查位置权限
   */
  async checkLocationPermission(): Promise<void> {
    try {
      this._hasLocationPermission = await this.locationService.checkLocationPermission();
      this.notifyStateChange();
    } catch (error) {
      this._hasLocationPermission = false;
      this._errorMessage = `权限检查失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Check permission error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 请求位置权限
   */
  async requestLocationPermission(): Promise<boolean> {
    try {
      const granted = await this.locationService.requestLocationPermission();
      this._hasLocationPermission = granted;
      
      if (!granted) {
        this._errorMessage = '位置权限被拒绝';
      } else {
        this._errorMessage = '';
      }
      
      this.notifyStateChange();
      return granted;
    } catch (error) {
      this._hasLocationPermission = false;
      this._errorMessage = `权限请求失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Request permission error: %{public}s', JSON.stringify(error));
      return false;
    }
  }

  /**
   * 开始位置跟踪
   */
  async startTracking(trajectoryName?: string): Promise<boolean> {
    try {
      if (this._isTracking) {
        return true;
      }

      // 检查权限
      if (!this._hasLocationPermission) {
        const granted = await this.requestLocationPermission();
        if (!granted) {
          return false;
        }
      }

      // 创建新轨迹
      const name = trajectoryName || `轨迹 ${new Date().toLocaleDateString()}`;
      this._currentTrajectory = Trajectory.createNew(name);
      
      // 保存轨迹
      await this.dataService.saveTrajectory(this._currentTrajectory);

      // 开始位置服务
      const success = await this.locationService.startTracking(this);
      
      if (success) {
        this._isTracking = true;
        this._errorMessage = '';
        await this.loadTrajectories(); // 刷新轨迹列表
      }

      this.notifyStateChange();
      return success;

    } catch (error) {
      this._errorMessage = `开始跟踪失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Start tracking error: %{public}s', JSON.stringify(error));
      return false;
    }
  }

  /**
   * 停止位置跟踪
   */
  async stopTracking(): Promise<void> {
    try {
      if (!this._isTracking) {
        return;
      }

      // 停止位置服务
      await this.locationService.stopTracking();

      // 完成当前轨迹
      if (this._currentTrajectory) {
        this._currentTrajectory.finish();
        await this.dataService.updateTrajectory(this._currentTrajectory);
      }

      this._isTracking = false;
      this._currentTrajectory = undefined;
      this._errorMessage = '';

      await this.loadTrajectories(); // 刷新轨迹列表
      this.notifyStateChange();

      hilog.info(0x0000, 'LocationViewModel', 'Tracking stopped');
    } catch (error) {
      this._errorMessage = `停止跟踪失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Stop tracking error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 暂停位置跟踪
   */
  pauseTracking(): void {
    if (this._isTracking && this._currentTrajectory) {
      this.locationService.pauseTracking();
      this._currentTrajectory.pause();
      this.notifyStateChange();
      hilog.info(0x0000, 'LocationViewModel', 'Tracking paused');
    }
  }

  /**
   * 恢复位置跟踪
   */
  resumeTracking(): void {
    if (this._isTracking && this._currentTrajectory) {
      this.locationService.resumeTracking();
      this._currentTrajectory.resume();
      this.notifyStateChange();
      hilog.info(0x0000, 'LocationViewModel', 'Tracking resumed');
    }
  }

  /**
   * 获取当前位置
   */
  async getCurrentLocation(): Promise<LocationPoint | null> {
    try {
      if (!this._hasLocationPermission) {
        const granted = await this.requestLocationPermission();
        if (!granted) {
          return null;
        }
      }

      const location = await this.locationService.getCurrentLocation();
      if (location) {
        this._currentLocation = location;
        this.notifyStateChange();
      }
      
      return location;
    } catch (error) {
      this._errorMessage = `获取位置失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Get current location error: %{public}s', JSON.stringify(error));
      return null;
    }
  }

  /**
   * 加载轨迹列表
   */
  async loadTrajectories(): Promise<void> {
    try {
      this._trajectories = await this.dataService.getAllTrajectories();
      this.notifyStateChange();
    } catch (error) {
      this._errorMessage = `加载轨迹失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Load trajectories error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 删除轨迹
   */
  async deleteTrajectory(trajectoryId: string): Promise<void> {
    try {
      await this.dataService.deleteTrajectory(trajectoryId);
      await this.loadTrajectories(); // 刷新列表
      this._errorMessage = '';
      this.notifyStateChange();
    } catch (error) {
      this._errorMessage = `删除轨迹失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Delete trajectory error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 获取轨迹的位置点
   */
  async getTrajectoryPoints(trajectoryId: string): Promise<LocationPoint[]> {
    try {
      return await this.dataService.getLocationPoints(trajectoryId);
    } catch (error) {
      this._errorMessage = `获取轨迹点失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Get trajectory points error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 获取统计信息
   */
  async getStatistics(): Promise<{
    totalTrajectories: number;
    totalDistance: number;
    totalDuration: number;
    totalPoints: number;
  }> {
    try {
      return await this.dataService.getStatistics();
    } catch (error) {
      hilog.error(0x0000, 'LocationViewModel', 'Get statistics error: %{public}s', JSON.stringify(error));
      return {
        totalTrajectories: 0,
        totalDistance: 0,
        totalDuration: 0,
        totalPoints: 0
      };
    }
  }

  /**
   * 清除所有数据
   */
  async clearAllData(): Promise<void> {
    try {
      // 先停止跟踪
      if (this._isTracking) {
        await this.stopTracking();
      }

      await this.dataService.clearAllData();
      this._trajectories = [];
      this._currentTrajectory = undefined;
      this._currentLocation = undefined;
      this._errorMessage = '';
      this.notifyStateChange();

      hilog.info(0x0000, 'LocationViewModel', 'All data cleared');
    } catch (error) {
      this._errorMessage = `清除数据失败: ${error}`;
      this.notifyStateChange();
      hilog.error(0x0000, 'LocationViewModel', 'Clear all data error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 清除错误消息
   */
  clearErrorMessage(): void {
    this._errorMessage = '';
    this.notifyStateChange();
  }

  // LocationUpdateCallback 实现
  async onLocationUpdate(location: LocationPoint): Promise<void> {
    try {
      this._currentLocation = location;

      if (this._currentTrajectory && this._isTracking) {
        // 设置轨迹ID
        location.trajectoryId = this._currentTrajectory.id;

        // 保存位置点
        await this.dataService.saveLocationPoint(location);

        // 更新轨迹统计
        const points = await this.dataService.getLocationPoints(this._currentTrajectory.id);
        this._currentTrajectory.updateStats(points);
        await this.dataService.updateTrajectory(this._currentTrajectory);
      }

      this.notifyStateChange();
    } catch (error) {
      hilog.error(0x0000, 'LocationViewModel', 'Location update error: %{public}s', JSON.stringify(error));
    }
  }

  onError(error: LocationError, message: string): void {
    this._errorMessage = message;
    this.notifyStateChange();
    hilog.error(0x0000, 'LocationViewModel', 'Location error: %{public}s - %{public}s', error, message);
  }

  onStatusChange(status: LocationServiceStatus): void {
    this._locationServiceStatus = status;
    this.notifyStateChange();
  }

  /**
   * 清理资源
   */
  async cleanup(): Promise<void> {
    try {
      if (this._isTracking) {
        await this.stopTracking();
      }
      
      await this.locationService.cleanup();
      await this.dataService.close();
      
      this.stateChangeCallbacks = [];
      hilog.info(0x0000, 'LocationViewModel', 'ViewModel cleanup completed');
    } catch (error) {
      hilog.error(0x0000, 'LocationViewModel', 'Cleanup error: %{public}s', JSON.stringify(error));
    }
  }
}