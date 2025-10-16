import ServiceExtensionAbility from '@ohos.app.ability.ServiceExtensionAbility';
import Want from '@ohos.app.ability.Want';
import hilog from '@ohos.hilog';
import notification from '@ohos.notificationManager';
import { LocationService, LocationServiceStatus, LocationError, LocationUpdateCallback } from './LocationService';
import { LocationPoint } from '../model/LocationPoint';
import { Trajectory, TrajectoryStatus } from '../model/Trajectory';
import { DataService } from './DataService';

/**
 * 后台位置跟踪服务
 */
export default class LocationTrackingService extends ServiceExtensionAbility implements LocationUpdateCallback {
  private locationService: LocationService;
  private dataService: DataService;
  private currentTrajectory?: Trajectory;
  private isTracking: boolean = false;
  private notificationId: number = 1001;

  onCreate(want: Want): void {
    hilog.info(0x0000, 'LocationTrackingService', 'Service onCreate');
    
    this.locationService = LocationService.getInstance();
    this.dataService = DataService.getInstance();
    
    // 创建通知渠道
    this.createNotificationChannel();
  }

  onRequest(want: Want, startId: number): void {
    hilog.info(0x0000, 'LocationTrackingService', 'Service onRequest, startId: %{public}d', startId);
    
    const action = want.parameters?.['action'] as string;
    
    switch (action) {
      case 'start_tracking':
        this.startTracking(want);
        break;
      case 'stop_tracking':
        this.stopTracking();
        break;
      case 'pause_tracking':
        this.pauseTracking();
        break;
      case 'resume_tracking':
        this.resumeTracking();
        break;
      default:
        hilog.warn(0x0000, 'LocationTrackingService', 'Unknown action: %{public}s', action);
    }
  }

  onDestroy(): void {
    hilog.info(0x0000, 'LocationTrackingService', 'Service onDestroy');
    this.cleanup();
  }

  /**
   * 开始位置跟踪
   */
  private async startTracking(want: Want): Promise<void> {
    try {
      if (this.isTracking) {
        hilog.warn(0x0000, 'LocationTrackingService', 'Tracking already started');
        return;
      }

      // 创建新轨迹
      const trajectoryName = want.parameters?.['trajectoryName'] as string || '新轨迹';
      this.currentTrajectory = Trajectory.createNew(trajectoryName);
      
      // 保存轨迹到数据库
      await this.dataService.saveTrajectory(this.currentTrajectory);

      // 开始位置服务
      const success = await this.locationService.startTracking(this);
      
      if (success) {
        this.isTracking = true;
        this.showTrackingNotification();
        hilog.info(0x0000, 'LocationTrackingService', 'Tracking started successfully');
      } else {
        hilog.error(0x0000, 'LocationTrackingService', 'Failed to start tracking');
      }

    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Start tracking error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 停止位置跟踪
   */
  private async stopTracking(): Promise<void> {
    try {
      if (!this.isTracking) {
        return;
      }

      // 停止位置服务
      await this.locationService.stopTracking();

      // 完成当前轨迹
      if (this.currentTrajectory) {
        this.currentTrajectory.finish();
        await this.dataService.updateTrajectory(this.currentTrajectory);
      }

      this.isTracking = false;
      this.currentTrajectory = undefined;
      
      // 取消通知
      this.cancelNotification();
      
      hilog.info(0x0000, 'LocationTrackingService', 'Tracking stopped');

    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Stop tracking error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 暂停位置跟踪
   */
  private pauseTracking(): void {
    if (this.isTracking && this.currentTrajectory) {
      this.locationService.pauseTracking();
      this.currentTrajectory.pause();
      this.updateNotification('轨迹记录已暂停');
      hilog.info(0x0000, 'LocationTrackingService', 'Tracking paused');
    }
  }

  /**
   * 恢复位置跟踪
   */
  private resumeTracking(): void {
    if (this.isTracking && this.currentTrajectory) {
      this.locationService.resumeTracking();
      this.currentTrajectory.resume();
      this.updateNotification('正在记录轨迹...');
      hilog.info(0x0000, 'LocationTrackingService', 'Tracking resumed');
    }
  }

  /**
   * 位置更新回调
   */
  async onLocationUpdate(location: LocationPoint): Promise<void> {
    try {
      if (!this.currentTrajectory || !this.isTracking) {
        return;
      }

      // 设置轨迹ID
      location.trajectoryId = this.currentTrajectory.id;

      // 保存位置点
      await this.dataService.saveLocationPoint(location);

      // 更新轨迹统计
      const points = await this.dataService.getLocationPoints(this.currentTrajectory.id);
      this.currentTrajectory.updateStats(points);
      await this.dataService.updateTrajectory(this.currentTrajectory);

      // 更新通知
      this.updateTrackingNotification();

      hilog.debug(0x0000, 'LocationTrackingService', 'Location updated: %{public}s', location.getFormattedCoordinates());

    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Location update error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 错误回调
   */
  onError(error: LocationError, message: string): void {
    hilog.error(0x0000, 'LocationTrackingService', 'Location error: %{public}s - %{public}s', error, message);
    
    // 显示错误通知
    this.showErrorNotification(message);
    
    // 根据错误类型决定是否停止跟踪
    if (error === LocationError.PERMISSION_DENIED || error === LocationError.LOCATION_DISABLED) {
      this.stopTracking();
    }
  }

  /**
   * 状态变化回调
   */
  onStatusChange(status: LocationServiceStatus): void {
    hilog.info(0x0000, 'LocationTrackingService', 'Location status changed: %{public}s', status);
    
    switch (status) {
      case LocationServiceStatus.RUNNING:
        this.updateNotification('正在记录轨迹...');
        break;
      case LocationServiceStatus.PAUSED:
        this.updateNotification('轨迹记录已暂停');
        break;
      case LocationServiceStatus.ERROR:
        this.updateNotification('位置服务出错');
        break;
    }
  }

  /**
   * 创建通知渠道
   */
  private async createNotificationChannel(): Promise<void> {
    try {
      const channelRequest: notification.NotificationSlot = {
        type: notification.SlotType.SERVICE_INFORMATION,
        slotId: 'location_tracking',
        name: '位置跟踪',
        description: '显示位置跟踪状态',
        level: notification.SlotLevel.LEVEL_DEFAULT,
        badgeFlag: true,
        bypassDnd: false,
        lockscreenVisibility: notification.VisibilityType.PUBLIC,
        vibrationEnabled: false,
        sound: '',
        lightEnabled: false,
        lightColor: 0
      };

      await notification.addSlot(channelRequest);
      hilog.info(0x0000, 'LocationTrackingService', 'Notification channel created');

    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Create notification channel error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 显示跟踪通知
   */
  private async showTrackingNotification(): Promise<void> {
    const notificationRequest: notification.NotificationRequest = {
      id: this.notificationId,
      slotType: notification.SlotType.SERVICE_INFORMATION,
      content: {
        contentType: notification.ContentType.NOTIFICATION_CONTENT_BASIC_TEXT,
        normal: {
          title: '太上老君轨迹跟踪',
          text: '正在记录轨迹...',
          additionalText: ''
        }
      },
      actionButtons: [
        {
          title: '暂停',
          wantAgent: undefined // 实际应用中需要创建WantAgent
        },
        {
          title: '停止',
          wantAgent: undefined // 实际应用中需要创建WantAgent
        }
      ],
      isOngoing: true,
      isUnremovable: true,
      deliveryTime: Date.now(),
      showDeliveryTime: true
    };

    try {
      await notification.publish(notificationRequest);
      hilog.info(0x0000, 'LocationTrackingService', 'Tracking notification shown');
    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Show notification error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 更新跟踪通知
   */
  private async updateTrackingNotification(): Promise<void> {
    if (!this.currentTrajectory) {
      return;
    }

    const distance = this.currentTrajectory.getFormattedDistance();
    const duration = this.currentTrajectory.getFormattedDuration();
    const text = `距离: ${distance} | 时长: ${duration}`;

    await this.updateNotification(text);
  }

  /**
   * 更新通知内容
   */
  private async updateNotification(text: string): Promise<void> {
    const notificationRequest: notification.NotificationRequest = {
      id: this.notificationId,
      slotType: notification.SlotType.SERVICE_INFORMATION,
      content: {
        contentType: notification.ContentType.NOTIFICATION_CONTENT_BASIC_TEXT,
        normal: {
          title: '太上老君轨迹跟踪',
          text: text,
          additionalText: ''
        }
      },
      isOngoing: true,
      isUnremovable: true,
      deliveryTime: Date.now(),
      showDeliveryTime: true
    };

    try {
      await notification.publish(notificationRequest);
    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Update notification error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 显示错误通知
   */
  private async showErrorNotification(message: string): Promise<void> {
    const notificationRequest: notification.NotificationRequest = {
      id: this.notificationId + 1,
      slotType: notification.SlotType.SERVICE_INFORMATION,
      content: {
        contentType: notification.ContentType.NOTIFICATION_CONTENT_BASIC_TEXT,
        normal: {
          title: '位置跟踪错误',
          text: message,
          additionalText: ''
        }
      },
      isOngoing: false,
      deliveryTime: Date.now(),
      showDeliveryTime: true
    };

    try {
      await notification.publish(notificationRequest);
    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Show error notification error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 取消通知
   */
  private async cancelNotification(): Promise<void> {
    try {
      await notification.cancel(this.notificationId);
      hilog.info(0x0000, 'LocationTrackingService', 'Notification cancelled');
    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Cancel notification error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 清理资源
   */
  private async cleanup(): Promise<void> {
    try {
      await this.stopTracking();
      await this.locationService.cleanup();
      hilog.info(0x0000, 'LocationTrackingService', 'Service cleanup completed');
    } catch (error) {
      hilog.error(0x0000, 'LocationTrackingService', 'Cleanup error: %{public}s', JSON.stringify(error));
    }
  }
}