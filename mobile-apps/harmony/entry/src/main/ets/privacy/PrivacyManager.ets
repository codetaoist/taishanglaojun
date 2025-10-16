import abilityAccessCtrl, { Permissions } from '@ohos.abilityAccessCtrl';
import bundleManager from '@ohos.bundle.bundleManager';
import preferences from '@ohos.data.preferences';
import { LocationPoint } from '../model/LocationPoint';
import { Trajectory } from '../model/Trajectory';
import hilog from '@ohos.hilog';
import cryptoFramework from '@ohos.security.cryptoFramework';

/**
 * 隐私管理器
 * 负责处理权限申请、用户同意和数据脱敏
 */
export class PrivacyManager {
  private static instance: PrivacyManager;
  private context: Context;
  private preferences: preferences.Preferences | null = null;
  private atManager: abilityAccessCtrl.AtManager;
  
  // SharedPreferences键
  private static readonly PREFS_NAME = 'privacy_prefs';
  private static readonly KEY_PRIVACY_POLICY_ACCEPTED = 'privacy_policy_accepted';
  private static readonly KEY_LOCATION_PERMISSION_GRANTED = 'location_permission_granted';
  private static readonly KEY_BACKGROUND_LOCATION_GRANTED = 'background_location_granted';
  private static readonly KEY_DATA_COLLECTION_CONSENT = 'data_collection_consent';
  private static readonly KEY_DATA_SHARING_CONSENT = 'data_sharing_consent';
  private static readonly KEY_ANALYTICS_CONSENT = 'analytics_consent';
  private static readonly KEY_FIRST_LAUNCH = 'first_launch';
  
  private constructor(context: Context) {
    this.context = context;
    this.atManager = abilityAccessCtrl.createAtManager();
    this.initPreferences();
  }
  
  /**
   * 获取单例实例
   */
  public static getInstance(context: Context): PrivacyManager {
    if (!PrivacyManager.instance) {
      PrivacyManager.instance = new PrivacyManager(context);
    }
    return PrivacyManager.instance;
  }
  
  /**
   * 初始化首选项
   */
  private async initPreferences(): Promise<void> {
    try {
      this.preferences = await preferences.getPreferences(this.context, PrivacyManager.PREFS_NAME);
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Init preferences error: %{public}s', JSON.stringify(error));
    }
  }
  
  /**
   * 隐私同意状态
   */
  export interface PrivacyConsent {
    privacyPolicyAccepted: boolean;
    dataCollectionConsent: boolean;
    dataSharingConsent: boolean;
    analyticsConsent: boolean;
  }
  
  /**
   * 权限状态
   */
  export interface PermissionStatus {
    locationPermission: boolean;
    backgroundLocationPermission: boolean;
    preciseLocationPermission: boolean;
  }
  
  /**
   * 数据脱敏级别
   */
  export enum DataSensitivityLevel {
    NONE = 'none',       // 不脱敏
    LOW = 'low',         // 低级脱敏（保留小数点后4位）
    MEDIUM = 'medium',   // 中级脱敏（保留小数点后3位）
    HIGH = 'high'        // 高级脱敏（保留小数点后2位）
  }
  
  /**
   * 检查是否首次启动
   */
  public async isFirstLaunch(): Promise<boolean> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      return await this.preferences!.get(PrivacyManager.KEY_FIRST_LAUNCH, true) as boolean;
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Get first launch error: %{public}s', JSON.stringify(error));
      return true;
    }
  }
  
  /**
   * 标记首次启动完成
   */
  public async markFirstLaunchCompleted(): Promise<void> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      await this.preferences!.put(PrivacyManager.KEY_FIRST_LAUNCH, false);
      await this.preferences!.flush();
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Mark first launch completed error: %{public}s', JSON.stringify(error));
    }
  }
  
  /**
   * 检查隐私政策是否已接受
   */
  public async isPrivacyPolicyAccepted(): Promise<boolean> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      return await this.preferences!.get(PrivacyManager.KEY_PRIVACY_POLICY_ACCEPTED, false) as boolean;
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Get privacy policy accepted error: %{public}s', JSON.stringify(error));
      return false;
    }
  }
  
  /**
   * 设置隐私政策接受状态
   */
  public async setPrivacyPolicyAccepted(accepted: boolean): Promise<void> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      await this.preferences!.put(PrivacyManager.KEY_PRIVACY_POLICY_ACCEPTED, accepted);
      await this.preferences!.flush();
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Set privacy policy accepted error: %{public}s', JSON.stringify(error));
    }
  }
  
  /**
   * 获取隐私同意状态
   */
  public async getPrivacyConsent(): Promise<PrivacyConsent> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      return {
        privacyPolicyAccepted: await this.preferences!.get(PrivacyManager.KEY_PRIVACY_POLICY_ACCEPTED, false) as boolean,
        dataCollectionConsent: await this.preferences!.get(PrivacyManager.KEY_DATA_COLLECTION_CONSENT, false) as boolean,
        dataSharingConsent: await this.preferences!.get(PrivacyManager.KEY_DATA_SHARING_CONSENT, false) as boolean,
        analyticsConsent: await this.preferences!.get(PrivacyManager.KEY_ANALYTICS_CONSENT, false) as boolean
      };
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Get privacy consent error: %{public}s', JSON.stringify(error));
      return {
        privacyPolicyAccepted: false,
        dataCollectionConsent: false,
        dataSharingConsent: false,
        analyticsConsent: false
      };
    }
  }
  
  /**
   * 设置隐私同意状态
   */
  public async setPrivacyConsent(consent: PrivacyConsent): Promise<void> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      await this.preferences!.put(PrivacyManager.KEY_PRIVACY_POLICY_ACCEPTED, consent.privacyPolicyAccepted);
      await this.preferences!.put(PrivacyManager.KEY_DATA_COLLECTION_CONSENT, consent.dataCollectionConsent);
      await this.preferences!.put(PrivacyManager.KEY_DATA_SHARING_CONSENT, consent.dataSharingConsent);
      await this.preferences!.put(PrivacyManager.KEY_ANALYTICS_CONSENT, consent.analyticsConsent);
      await this.preferences!.flush();
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Set privacy consent error: %{public}s', JSON.stringify(error));
    }
  }
  
  /**
   * 检查位置权限状态
   */
  public async checkLocationPermissions(): Promise<PermissionStatus> {
    try {
      const bundleInfo = await bundleManager.getBundleInfoForSelf(bundleManager.BundleFlag.GET_BUNDLE_INFO_WITH_APPLICATION);
      const tokenId = bundleInfo.appInfo.accessTokenId;
      
      const preciseLocationPermission = await this.atManager.checkAccessToken(
        tokenId, 'ohos.permission.LOCATION'
      ) === abilityAccessCtrl.GrantStatus.PERMISSION_GRANTED;
      
      const approximateLocationPermission = await this.atManager.checkAccessToken(
        tokenId, 'ohos.permission.APPROXIMATELY_LOCATION'
      ) === abilityAccessCtrl.GrantStatus.PERMISSION_GRANTED;
      
      const backgroundLocationPermission = await this.atManager.checkAccessToken(
        tokenId, 'ohos.permission.LOCATION_IN_BACKGROUND'
      ) === abilityAccessCtrl.GrantStatus.PERMISSION_GRANTED;
      
      return {
        locationPermission: preciseLocationPermission || approximateLocationPermission,
        backgroundLocationPermission: backgroundLocationPermission,
        preciseLocationPermission: preciseLocationPermission
      };
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Check location permissions error: %{public}s', JSON.stringify(error));
      return {
        locationPermission: false,
        backgroundLocationPermission: false,
        preciseLocationPermission: false
      };
    }
  }
  
  /**
   * 请求位置权限
   */
  public async requestLocationPermissions(): Promise<boolean> {
    try {
      const permissions: Array<Permissions> = [
        'ohos.permission.APPROXIMATELY_LOCATION',
        'ohos.permission.LOCATION'
      ];
      
      const result = await this.atManager.requestPermissionsFromUser(this.context, permissions);
      
      return result.authResults.some(result => result === abilityAccessCtrl.GrantStatus.PERMISSION_GRANTED);
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Request location permissions error: %{public}s', JSON.stringify(error));
      return false;
    }
  }
  
  /**
   * 请求后台位置权限
   */
  public async requestBackgroundLocationPermission(): Promise<boolean> {
    try {
      const permissions: Array<Permissions> = ['ohos.permission.LOCATION_IN_BACKGROUND'];
      const result = await this.atManager.requestPermissionsFromUser(this.context, permissions);
      
      return result.authResults[0] === abilityAccessCtrl.GrantStatus.PERMISSION_GRANTED;
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Request background location permission error: %{public}s', JSON.stringify(error));
      return false;
    }
  }
  
  /**
   * 对位置点进行数据脱敏
   */
  public sanitizeLocationPoint(
    locationPoint: LocationPoint,
    level: DataSensitivityLevel = DataSensitivityLevel.MEDIUM
  ): LocationPoint {
    if (level === DataSensitivityLevel.NONE) {
      return locationPoint;
    }
    
    let precision: number;
    switch (level) {
      case DataSensitivityLevel.LOW:
        precision = 4;
        break;
      case DataSensitivityLevel.MEDIUM:
        precision = 3;
        break;
      case DataSensitivityLevel.HIGH:
        precision = 2;
        break;
      default:
        precision = 6;
    }
    
    const multiplier = Math.pow(10, precision);
    
    return {
      ...locationPoint,
      latitude: Math.round(locationPoint.latitude * multiplier) / multiplier,
      longitude: Math.round(locationPoint.longitude * multiplier) / multiplier,
      altitude: locationPoint.altitude ? Math.round(locationPoint.altitude * 10) / 10 : locationPoint.altitude,
      speed: locationPoint.speed ? Math.round(locationPoint.speed * 10) / 10 : locationPoint.speed,
      bearing: locationPoint.bearing ? Math.round(locationPoint.bearing) : locationPoint.bearing
    };
  }
  
  /**
   * 对轨迹进行数据脱敏
   */
  public sanitizeTrajectory(
    trajectory: Trajectory,
    level: DataSensitivityLevel = DataSensitivityLevel.MEDIUM
  ): Trajectory {
    if (level === DataSensitivityLevel.NONE) {
      return trajectory;
    }
    
    let precision: number;
    switch (level) {
      case DataSensitivityLevel.LOW:
        precision = 4;
        break;
      case DataSensitivityLevel.MEDIUM:
        precision = 3;
        break;
      case DataSensitivityLevel.HIGH:
        precision = 2;
        break;
      default:
        precision = 6;
    }
    
    const multiplier = Math.pow(10, precision);
    
    return {
      ...trajectory,
      name: level === DataSensitivityLevel.HIGH ? `轨迹${trajectory.id.toString().slice(-4)}` : trajectory.name,
      description: level === DataSensitivityLevel.HIGH ? '' : trajectory.description,
      minLatitude: trajectory.minLatitude ? Math.round(trajectory.minLatitude * multiplier) / multiplier : trajectory.minLatitude,
      maxLatitude: trajectory.maxLatitude ? Math.round(trajectory.maxLatitude * multiplier) / multiplier : trajectory.maxLatitude,
      minLongitude: trajectory.minLongitude ? Math.round(trajectory.minLongitude * multiplier) / multiplier : trajectory.minLongitude,
      maxLongitude: trajectory.maxLongitude ? Math.round(trajectory.maxLongitude * multiplier) / multiplier : trajectory.maxLongitude,
      totalDistance: Math.round(trajectory.totalDistance * 10) / 10,
      averageSpeed: trajectory.averageSpeed ? Math.round(trajectory.averageSpeed * 10) / 10 : trajectory.averageSpeed,
      maxSpeed: trajectory.maxSpeed ? Math.round(trajectory.maxSpeed * 10) / 10 : trajectory.maxSpeed
    };
  }
  
  /**
   * 生成匿名用户ID
   */
  public async generateAnonymousUserId(): Promise<string> {
    try {
      const deviceId = 'harmony_device_' + Date.now().toString();
      return await this.hashString(deviceId + Date.now().toString());
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Generate anonymous user ID error: %{public}s', JSON.stringify(error));
      return 'anonymous_' + Date.now().toString().slice(-8);
    }
  }
  
  /**
   * 对字符串进行哈希处理
   */
  private async hashString(input: string): Promise<string> {
    try {
      const md = cryptoFramework.createMd('SHA256');
      await md.update({ data: new Uint8Array(Buffer.from(input, 'utf-8')) });
      const result = await md.digest();
      
      // 转换为十六进制字符串
      const bytes = new Uint8Array(result.data);
      let hex = '';
      for (let i = 0; i < Math.min(bytes.length, 8); i++) {
        hex += bytes[i].toString(16).padStart(2, '0');
      }
      return hex;
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Hash string error: %{public}s', JSON.stringify(error));
      return input.slice(0, 16);
    }
  }
  
  /**
   * 检查数据收集是否被允许
   */
  public async isDataCollectionAllowed(): Promise<boolean> {
    const consent = await this.getPrivacyConsent();
    return consent.dataCollectionConsent;
  }
  
  /**
   * 检查数据分享是否被允许
   */
  public async isDataSharingAllowed(): Promise<boolean> {
    const consent = await this.getPrivacyConsent();
    return consent.dataSharingConsent;
  }
  
  /**
   * 检查分析数据收集是否被允许
   */
  public async isAnalyticsAllowed(): Promise<boolean> {
    const consent = await this.getPrivacyConsent();
    return consent.analyticsConsent;
  }
  
  /**
   * 获取隐私政策文本
   */
  public getPrivacyPolicyText(): string {
    const currentDate = new Date().toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
    
    return `隐私政策

1. 数据收集
我们收集您的位置信息以提供轨迹跟踪服务。

2. 数据使用
您的位置数据仅用于：
- 记录和显示您的移动轨迹
- 计算距离、速度等统计信息
- 提供轨迹分析功能

3. 数据存储
您的数据存储在本地设备上，我们不会将其上传到服务器，除非您明确同意。

4. 数据分享
我们不会与第三方分享您的个人位置数据，除非获得您的明确同意。

5. 数据安全
我们采用加密技术保护您的数据安全。

6. 您的权利
您可以随时：
- 查看您的数据
- 删除您的数据
- 撤回同意
- 导出您的数据

最后更新：${currentDate}`;
  }
  
  /**
   * 重置所有隐私设置
   */
  public async resetPrivacySettings(): Promise<void> {
    if (!this.preferences) {
      await this.initPreferences();
    }
    try {
      await this.preferences!.clear();
      await this.preferences!.flush();
    } catch (error) {
      hilog.error(0x0000, 'PrivacyManager', 'Reset privacy settings error: %{public}s', JSON.stringify(error));
    }
  }
}