import { Trajectory } from '../model/Trajectory';
import { LocationPoint } from '../model/LocationPoint';
import { DataService } from '../service/DataService';
import { TrajectoryManagementService, ExportFormat } from '../service/TrajectoryManagementService';
import { LocationUtils } from '../utils/LocationUtils';
import router from '@ohos.router';
import map from '@ohos.geoLocationManager';

/**
 * 轨迹详情页面
 */
@Entry
@Component
struct TrajectoryDetailPage {
  @State private trajectory: Trajectory | null = null;
  @State private locationPoints: LocationPoint[] = [];
  @State private isLoading: boolean = true;
  @State private selectedTab: number = 0;
  @State private showDeleteDialog: boolean = false;
  @State private showExportDialog: boolean = false;
  @State private showShareDialog: boolean = false;
  @State private exportFormat: ExportFormat = ExportFormat.GPX;
  @State private isExporting: boolean = false;
  
  // 地图相关状态
  @State private mapController: any = null;
  @State private showMap: boolean = false;
  @State private mapCenter: { latitude: number, longitude: number } = { latitude: 39.9042, longitude: 116.4074 };
  @State private mapZoom: number = 15;
  
  // 轨迹回放相关状态
  @State private isPlaying: boolean = false;
  @State private playbackIndex: number = 0;
  @State private playbackSpeed: number = 1; // 1x, 2x, 4x, 8x
  @State private playbackTimer: number = -1;
  @State private showPlaybackControls: boolean = false;

  private trajectoryId: string = '';
  private dataService: DataService = DataService.getInstance();
  private managementService: TrajectoryManagementService = TrajectoryManagementService.getInstance();

  aboutToAppear() {
    // 从路由参数获取轨迹ID
    const params = router.getParams() as Record<string, Object>;
    this.trajectoryId = params?.trajectoryId as string || 'test_trajectory_id';
    this.loadTrajectoryDetail();
  }

  /**
   * 加载轨迹详情
   */
  private async loadTrajectoryDetail() {
    try {
      this.isLoading = true;
      
      // 加载轨迹基本信息
      this.trajectory = await this.dataService.getTrajectory(this.trajectoryId);
      
      // 加载轨迹点数据
      this.locationPoints = await this.dataService.getLocationPointsByTrajectory(this.trajectoryId);
      
      // 初始化地图中心点
      if (this.locationPoints.length > 0) {
        const firstPoint = this.locationPoints[0];
        this.mapCenter = {
          latitude: firstPoint.latitude,
          longitude: firstPoint.longitude
        };
      }
      
    } catch (error) {
      console.error('Failed to load trajectory detail:', error);
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * 初始化地图
   */
  private initializeMap() {
    if (this.locationPoints.length === 0) return;

    // 计算地图边界
    let minLat = this.locationPoints[0].latitude;
    let maxLat = this.locationPoints[0].latitude;
    let minLng = this.locationPoints[0].longitude;
    let maxLng = this.locationPoints[0].longitude;

    this.locationPoints.forEach(point => {
      minLat = Math.min(minLat, point.latitude);
      maxLat = Math.max(maxLat, point.latitude);
      minLng = Math.min(minLng, point.longitude);
      maxLng = Math.max(maxLng, point.longitude);
    });

    // 设置地图中心和缩放级别
    this.mapCenter = {
      latitude: (minLat + maxLat) / 2,
      longitude: (minLng + maxLng) / 2
    };

    // 根据轨迹范围计算合适的缩放级别
    const latDiff = maxLat - minLat;
    const lngDiff = maxLng - minLng;
    const maxDiff = Math.max(latDiff, lngDiff);
    
    if (maxDiff > 0.1) {
      this.mapZoom = 10;
    } else if (maxDiff > 0.01) {
      this.mapZoom = 13;
    } else {
      this.mapZoom = 15;
    }
  }

  /**
   * 开始轨迹回放
   */
  private startPlayback() {
    if (this.locationPoints.length === 0) return;

    this.isPlaying = true;
    this.playbackIndex = 0;
    this.showPlaybackControls = true;

    const interval = 1000 / this.playbackSpeed; // 根据速度调整间隔
    this.playbackTimer = setInterval(() => {
      if (this.playbackIndex < this.locationPoints.length - 1) {
        this.playbackIndex++;
        
        // 更新地图中心到当前回放位置
        const currentPoint = this.locationPoints[this.playbackIndex];
        this.mapCenter = {
          latitude: currentPoint.latitude,
          longitude: currentPoint.longitude
        };
      } else {
        this.stopPlayback();
      }
    }, interval);
  }

  /**
   * 停止轨迹回放
   */
  private stopPlayback() {
    this.isPlaying = false;
    if (this.playbackTimer !== -1) {
      clearInterval(this.playbackTimer);
      this.playbackTimer = -1;
    }
  }

  /**
   * 暂停/恢复轨迹回放
   */
  private togglePlayback() {
    if (this.isPlaying) {
      this.stopPlayback();
    } else {
      this.startPlayback();
    }
  }

  /**
   * 重置轨迹回放
   */
  private resetPlayback() {
    this.stopPlayback();
    this.playbackIndex = 0;
    if (this.locationPoints.length > 0) {
      const firstPoint = this.locationPoints[0];
      this.mapCenter = {
        latitude: firstPoint.latitude,
        longitude: firstPoint.longitude
      };
    }
  }

  /**
   * 设置回放速度
   */
  private setPlaybackSpeed(speed: number) {
    this.playbackSpeed = speed;
    if (this.isPlaying) {
      this.stopPlayback();
      this.startPlayback();
    }
  }

  /**
   * 跳转到指定回放位置
   */
  private seekToPosition(index: number) {
    if (index >= 0 && index < this.locationPoints.length) {
      this.playbackIndex = index;
      const point = this.locationPoints[index];
      this.mapCenter = {
        latitude: point.latitude,
        longitude: point.longitude
      };
    }
  }

  /**
   * 删除轨迹
   */
  private async deleteTrajectory() {
    try {
      if (this.trajectory) {
        await this.dataService.deleteTrajectory(this.trajectory.id);
        // 返回上一页
        // router.back();
      }
    } catch (error) {
      console.error('Failed to delete trajectory:', error);
    }
  }

  /**
   * 导出轨迹
   */
  private async exportTrajectory() {
    if (!this.trajectory) return;

    this.isExporting = true;
    try {
      const result = await this.managementService.exportTrajectory(this.trajectory.id, this.exportFormat);
      if (result.success) {
        // 显示成功提示
        console.log('轨迹导出成功:', result.filePath);
      } else {
        console.error('轨迹导出失败:', result.error);
      }
    } catch (error) {
      console.error('导出轨迹时发生错误:', error);
    } finally {
      this.isExporting = false;
      this.showExportDialog = false;
    }
  }

  /**
   * 分享轨迹
   */
  private async shareTrajectory() {
    if (!this.trajectory) return;

    this.isExporting = true;
    try {
      const shareOptions = {
        title: `轨迹分享 - ${this.trajectory.name}`,
        description: `距离: ${this.formatDistance(this.trajectory.distance)}, 时长: ${this.formatDuration(this.trajectory.duration)}`
      };
      
      const result = await this.managementService.shareTrajectory(this.trajectory.id, this.exportFormat, shareOptions);
      if (result.success) {
        console.log('轨迹分享成功');
      } else {
        console.error('轨迹分享失败:', result.error);
      }
    } catch (error) {
      console.error('分享轨迹时发生错误:', error);
    } finally {
      this.isExporting = false;
      this.showShareDialog = false;
    }
  }

  /**
   * 格式化时间
   */
  private formatTime(timestamp: number): string {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN');
  }

  /**
   * 格式化持续时间
   */
  private formatDuration(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}小时${minutes % 60}分钟`;
    } else if (minutes > 0) {
      return `${minutes}分钟${seconds % 60}秒`;
    } else {
      return `${seconds}秒`;
    }
  }

  /**
   * 格式化距离
   */
  private formatDistance(meters: number): string {
    if (meters >= 1000) {
      return `${(meters / 1000).toFixed(2)} 公里`;
    } else {
      return `${meters.toFixed(0)} 米`;
    }
  }

  /**
   * 格式化速度
   */
  private formatSpeed(metersPerSecond: number): string {
    const kmPerHour = metersPerSecond * 3.6;
    return `${kmPerHour.toFixed(1)} km/h`;
  }

  /**
   * 获取活动类型文本
   */
  private getActivityTypeText(activityType: string): string {
    switch (activityType) {
      case 'walking': return '步行';
      case 'running': return '跑步';
      case 'cycling': return '骑行';
      case 'driving': return '驾车';
      default: return activityType;
    }
  }

  /**
   * 获取轨迹类型文本
   */
  private getTrajectoryTypeText(trajectoryType: string): string {
    switch (trajectoryType) {
      case 'exercise': return '运动';
      case 'commute': return '通勤';
      case 'travel': return '旅行';
      case 'leisure': return '休闲';
      default: return trajectoryType;
    }
  }

  /**
   * 计算轨迹统计信息
   */
  private getTrajectoryStats() {
    if (!this.trajectory || this.locationPoints.length === 0) {
      return {
        totalPoints: 0,
        averageAccuracy: 0,
        elevationGain: 0,
        elevationLoss: 0,
        movingTime: 0,
        stoppedTime: 0
      };
    }

    const points = this.locationPoints;
    let totalAccuracy = 0;
    let elevationGain = 0;
    let elevationLoss = 0;
    let movingTime = 0;
    let stoppedTime = 0;
    let lastElevation = points[0].altitude || 0;
    let lastTimestamp = points[0].timestamp;

    for (let i = 1; i < points.length; i++) {
      const point = points[i];
      const prevPoint = points[i - 1];
      
      // 累计精度
      totalAccuracy += point.accuracy || 0;
      
      // 计算海拔变化
      if (point.altitude && lastElevation) {
        const elevationDiff = point.altitude - lastElevation;
        if (elevationDiff > 0) {
          elevationGain += elevationDiff;
        } else {
          elevationLoss += Math.abs(elevationDiff);
        }
        lastElevation = point.altitude;
      }
      
      // 计算移动时间和停止时间
      const timeDiff = point.timestamp - lastTimestamp;
      const distance = LocationUtils.calculateDistance(
        prevPoint.latitude, prevPoint.longitude,
        point.latitude, point.longitude
      );
      
      // 如果距离很小且时间间隔较长，认为是停止状态
      if (distance < 5 && timeDiff > 30000) { // 5米内且超过30秒
        stoppedTime += timeDiff;
      } else {
        movingTime += timeDiff;
      }
      
      lastTimestamp = point.timestamp;
    }

    return {
      totalPoints: points.length,
      averageAccuracy: totalAccuracy / points.length,
      elevationGain,
      elevationLoss,
      movingTime,
      stoppedTime
    };
  }

  build() {
    Column() {
      // 顶部导航栏
      Row() {
        Row({ space: 12 }) {
          Image($r('app.media.ic_arrow_back'))
            .width(24)
            .height(24)
            .fillColor(Color.Black)
            .onClick(() => {
              // router.back();
            })

          Text('轨迹详情')
            .fontSize(18)
            .fontWeight(FontWeight.Medium)
            .fontColor(Color.Black)
        }

        Row({ space: 12 }) {
          Image($r('app.media.ic_export'))
            .width(24)
            .height(24)
            .fillColor('#34C759')
            .onClick(() => {
              this.showExportDialog = true;
            })

          Image($r('app.media.ic_share'))
            .width(24)
            .height(24)
            .fillColor('#007AFF')
            .onClick(() => {
              this.showShareDialog = true;
            })

          Image($r('app.media.ic_delete'))
            .width(24)
            .height(24)
            .fillColor(Color.Red)
            .onClick(() => {
              this.showDeleteDialog = true;
            })
        }
      }
      .width('100%')
      .height(56)
      .padding({ left: 16, right: 16 })
      .justifyContent(FlexAlign.SpaceBetween)
      .backgroundColor(Color.White)
      .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })

      if (this.isLoading) {
        // 加载状态
        Column({ space: 16 }) {
          LoadingProgress()
            .width(40)
            .height(40)
            .color('#007AFF')

          Text('加载中...')
            .fontSize(14)
            .fontColor(Color.Gray)
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)

      } else if (!this.trajectory) {
        // 错误状态
        Column({ space: 16 }) {
          Image($r('app.media.ic_error'))
            .width(60)
            .height(60)
            .fillColor(Color.Gray)

          Text('轨迹不存在')
            .fontSize(16)
            .fontColor(Color.Gray)

          Button('返回')
            .width(100)
            .height(40)
            .backgroundColor('#007AFF')
            .fontColor(Color.White)
            .onClick(() => {
              // router.back();
            })
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)

      } else {
        // 主要内容
        Column() {
          // 标签页导航
          Row() {
            ForEach(['概览', '地图', '统计', '详细'], (tab: string, index: number) => {
              Text(tab)
                .fontSize(16)
                .fontWeight(this.selectedTab === index ? FontWeight.Medium : FontWeight.Normal)
                .fontColor(this.selectedTab === index ? '#007AFF' : '#666666')
                .padding({ top: 12, bottom: 12 })
                .layoutWeight(1)
                .textAlign(TextAlign.Center)
                .onClick(() => {
                  this.selectedTab = index;
                  if (index === 1) { // 地图标签页
                    this.showMap = true;
                    this.initializeMap();
                  }
                })
                .border({
                  width: { bottom: this.selectedTab === index ? 2 : 0 },
                  color: '#007AFF'
                })
            })
          }
          .width('100%')
          .backgroundColor(Color.White)
          .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })

          // 标签页内容
          if (this.selectedTab === 0) {
            this.buildOverviewTab()
          } else if (this.selectedTab === 1) {
            this.buildMapTab()
          } else if (this.selectedTab === 2) {
            this.buildStatisticsTab()
          } else {
            this.buildDetailTab()
          }
        }
        .layoutWeight(1)
      }

      // 删除确认对话框
      if (this.showDeleteDialog) {
        this.buildDeleteDialog()
      }
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F8F8F8')
  }

  /**
   * 构建概览标签页
   */
  @Builder
  buildOverviewTab() {
    Scroll() {
      Column({ space: 16 }) {
        // 基本信息卡片
        Column({ space: 12 }) {
          Text('基本信息')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .alignSelf(ItemAlign.Start)

          Column({ space: 8 }) {
            Row() {
              Text('开始时间')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(this.formatTime(this.trajectory!.startTime))
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            Row() {
              Text('结束时间')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(this.formatTime(this.trajectory!.endTime))
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            Row() {
              Text('持续时间')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(this.formatDuration(this.trajectory!.endTime - this.trajectory!.startTime))
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            Row() {
              Text('总距离')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(this.formatDistance(this.trajectory!.totalDistance))
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            if (this.trajectory!.averageSpeed) {
              Row() {
                Text('平均速度')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(this.formatSpeed(this.trajectory!.averageSpeed))
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')
            }

            if (this.trajectory!.maxSpeed) {
              Row() {
                Text('最高速度')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(this.formatSpeed(this.trajectory!.maxSpeed))
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')
            }

            if (this.trajectory!.activityType) {
              Row() {
                Text('活动类型')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(this.getActivityTypeText(this.trajectory!.activityType))
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')
            }

            if (this.trajectory!.trajectoryType) {
              Row() {
                Text('轨迹类型')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(this.getTrajectoryTypeText(this.trajectory!.trajectoryType))
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')
            }
          }
        }
        .width('100%')
        .padding(16)
        .backgroundColor(Color.White)
        .borderRadius(8)

        // 地图预览（占位）
        Column({ space: 12 }) {
          Text('轨迹地图')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .alignSelf(ItemAlign.Start)

          Column() {
            Text('地图预览')
              .fontSize(14)
              .fontColor('#999999')
          }
          .width('100%')
          .height(200)
          .backgroundColor('#F0F0F0')
          .borderRadius(8)
          .justifyContent(FlexAlign.Center)
          .alignItems(HorizontalAlign.Center)
        }
        .width('100%')
        .padding(16)
        .backgroundColor(Color.White)
        .borderRadius(8)
      }
      .width('100%')
      .padding(16)
    }
    .width('100%')
    .layoutWeight(1)
  }

  /**
   * 构建统计标签页
   */
  @Builder
  buildStatisticsTab() {
    Scroll() {
      Column({ space: 16 }) {
        const stats = this.getTrajectoryStats();

        // 轨迹点统计
        Column({ space: 12 }) {
          Text('轨迹点统计')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .alignSelf(ItemAlign.Start)

          Column({ space: 8 }) {
            Row() {
              Text('总点数')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(`${stats.totalPoints} 个`)
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            Row() {
              Text('平均精度')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(`${stats.averageAccuracy.toFixed(1)} 米`)
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')
          }
        }
        .width('100%')
        .padding(16)
        .backgroundColor(Color.White)
        .borderRadius(8)

        // 海拔统计
        if (stats.elevationGain > 0 || stats.elevationLoss > 0) {
          Column({ space: 12 }) {
            Text('海拔统计')
              .fontSize(16)
              .fontWeight(FontWeight.Medium)
              .fontColor('#333333')
              .alignSelf(ItemAlign.Start)

            Column({ space: 8 }) {
              Row() {
                Text('累计爬升')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(`${stats.elevationGain.toFixed(1)} 米`)
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')

              Row() {
                Text('累计下降')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(`${stats.elevationLoss.toFixed(1)} 米`)
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')
            }
          }
          .width('100%')
          .padding(16)
          .backgroundColor(Color.White)
          .borderRadius(8)
        }

        // 时间统计
        Column({ space: 12 }) {
          Text('时间统计')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .alignSelf(ItemAlign.Start)

          Column({ space: 8 }) {
            Row() {
              Text('移动时间')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(this.formatDuration(stats.movingTime))
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            Row() {
              Text('停止时间')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              
              Text(this.formatDuration(stats.stoppedTime))
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')
          }
        }
        .width('100%')
        .padding(16)
        .backgroundColor(Color.White)
        .borderRadius(8)

        // 范围统计
        if (this.trajectory!.minLatitude && this.trajectory!.maxLatitude) {
          Column({ space: 12 }) {
            Text('范围统计')
              .fontSize(16)
              .fontWeight(FontWeight.Medium)
              .fontColor('#333333')
              .alignSelf(ItemAlign.Start)

            Column({ space: 8 }) {
              Row() {
                Text('纬度范围')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(`${this.trajectory!.minLatitude!.toFixed(6)} ~ ${this.trajectory!.maxLatitude!.toFixed(6)}`)
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')

              Row() {
                Text('经度范围')
                  .fontSize(14)
                  .fontColor('#666666')
                  .layoutWeight(1)
                
                Text(`${this.trajectory!.minLongitude!.toFixed(6)} ~ ${this.trajectory!.maxLongitude!.toFixed(6)}`)
                  .fontSize(14)
                  .fontColor('#333333')
              }
              .width('100%')
            }
          }
          .width('100%')
          .padding(16)
          .backgroundColor(Color.White)
          .borderRadius(8)
        }
      }
      .width('100%')
      .padding(16)
    }
    .width('100%')
    .layoutWeight(1)
  }

  /**
   * 构建详细标签页
   */
  @Builder
  buildDetailTab() {
    Column() {
      // 轨迹点列表头部
      Row() {
        Text(`轨迹点列表 (${this.locationPoints.length})`)
          .fontSize(16)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
          .layoutWeight(1)

        Text('导出')
          .fontSize(14)
          .fontColor('#007AFF')
          .onClick(() => {
            // 导出轨迹点数据
          })
      }
      .width('100%')
      .height(48)
      .padding({ left: 16, right: 16 })
      .backgroundColor(Color.White)
      .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })

      // 轨迹点列表
      List({ space: 0 }) {
        ForEach(this.locationPoints, (point: LocationPoint, index: number) => {
          ListItem() {
            Row({ space: 12 }) {
              // 序号
              Text(`${index + 1}`)
                .fontSize(12)
                .fontColor('#999999')
                .width(30)
                .textAlign(TextAlign.Center)

              // 位置信息
              Column({ space: 4 }) {
                Text(`${point.latitude.toFixed(6)}, ${point.longitude.toFixed(6)}`)
                  .fontSize(14)
                  .fontColor('#333333')
                  .alignSelf(ItemAlign.Start)

                Row({ space: 12 }) {
                  Text(this.formatTime(point.timestamp))
                    .fontSize(12)
                    .fontColor('#666666')

                  if (point.accuracy) {
                    Text(`精度: ${point.accuracy.toFixed(1)}m`)
                      .fontSize(12)
                      .fontColor('#666666')
                  }

                  if (point.altitude) {
                    Text(`海拔: ${point.altitude.toFixed(1)}m`)
                      .fontSize(12)
                      .fontColor('#666666')
                  }
                }
              }
              .layoutWeight(1)
              .alignItems(HorizontalAlign.Start)
            }
            .width('100%')
            .padding({ left: 16, right: 16, top: 12, bottom: 12 })
          }
          .backgroundColor(Color.White)
          .border({ width: { bottom: 0.5 }, color: '#F0F0F0' })
        })
      }
      .width('100%')
      .layoutWeight(1)
      .backgroundColor(Color.White)
    }
    .width('100%')
    .layoutWeight(1)
  }

  /**
   * 构建地图标签页
   */
  @Builder
  buildMapTab() {
    Column() {
      // 地图容器
      Stack() {
        // 地图组件 (使用模拟地图)
        Column() {
          Text('轨迹地图')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .margin({ bottom: 16 })

          // 模拟地图显示
          Column() {
            Row() {
              Text('中心位置:')
                .fontSize(12)
                .fontColor('#666666')
              Text(`${this.mapCenter.latitude.toFixed(6)}, ${this.mapCenter.longitude.toFixed(6)}`)
                .fontSize(12)
                .fontColor('#333333')
                .margin({ left: 8 })
            }
            .width('100%')
            .margin({ bottom: 8 })

            Row() {
              Text('缩放级别:')
                .fontSize(12)
                .fontColor('#666666')
              Text(`${this.mapZoom}`)
                .fontSize(12)
                .fontColor('#333333')
                .margin({ left: 8 })
            }
            .width('100%')
            .margin({ bottom: 8 })

            Row() {
              Text('轨迹点数:')
                .fontSize(12)
                .fontColor('#666666')
              Text(`${this.locationPoints.length}`)
                .fontSize(12)
                .fontColor('#333333')
                .margin({ left: 8 })
            }
            .width('100%')
            .margin({ bottom: 16 })

            // 轨迹路径可视化 (简化版)
            if (this.locationPoints.length > 0) {
              Column() {
                Text('轨迹路径')
                  .fontSize(14)
                  .fontWeight(FontWeight.Medium)
                  .fontColor('#333333')
                  .margin({ bottom: 8 })

                // 显示当前回放位置
                if (this.showPlaybackControls) {
                  Row() {
                    Text('当前位置:')
                      .fontSize(12)
                      .fontColor('#666666')
                    Text(`${this.playbackIndex + 1}/${this.locationPoints.length}`)
                      .fontSize(12)
                      .fontColor('#007AFF')
                      .margin({ left: 8 })
                  }
                  .width('100%')
                  .margin({ bottom: 8 })

                  // 当前位置信息
                  if (this.playbackIndex < this.locationPoints.length) {
                    const currentPoint = this.locationPoints[this.playbackIndex];
                    Column({ space: 4 }) {
                      Text(`纬度: ${currentPoint.latitude.toFixed(6)}`)
                        .fontSize(12)
                        .fontColor('#333333')
                      Text(`经度: ${currentPoint.longitude.toFixed(6)}`)
                        .fontSize(12)
                        .fontColor('#333333')
                      if (currentPoint.altitude) {
                        Text(`海拔: ${currentPoint.altitude.toFixed(1)}m`)
                          .fontSize(12)
                          .fontColor('#333333')
                      }
                      Text(`时间: ${this.formatTime(currentPoint.timestamp)}`)
                        .fontSize(12)
                        .fontColor('#333333')
                    }
                    .alignItems(HorizontalAlign.Start)
                    .width('100%')
                    .padding(8)
                    .backgroundColor('#F0F8FF')
                    .borderRadius(4)
                  }
                }
              }
              .width('100%')
              .padding(12)
              .backgroundColor('#F8F8F8')
              .borderRadius(8)
            }
          }
          .width('100%')
          .height(300)
          .padding(16)
          .backgroundColor('#E8F4FD')
          .borderRadius(8)
          .border({ width: 1, color: '#D0E8FF' })
        }
        .width('100%')
        .layoutWeight(1)

        // 地图控制按钮
        Column() {
          Button() {
            Image($r('app.media.ic_my_location'))
              .width(24)
              .height(24)
              .fillColor(Color.White)
          }
          .width(48)
          .height(48)
          .backgroundColor('#007AFF')
          .borderRadius(24)
          .onClick(() => {
            // 回到轨迹起点
            if (this.locationPoints.length > 0) {
              const firstPoint = this.locationPoints[0];
              this.mapCenter = {
                latitude: firstPoint.latitude,
                longitude: firstPoint.longitude
              };
            }
          })

          Button() {
            Image($r('app.media.ic_zoom_in'))
              .width(24)
              .height(24)
              .fillColor(Color.White)
          }
          .width(48)
          .height(48)
          .backgroundColor('#007AFF')
          .borderRadius(24)
          .margin({ top: 8 })
          .onClick(() => {
            this.mapZoom = Math.min(this.mapZoom + 1, 20);
          })

          Button() {
            Image($r('app.media.ic_zoom_out'))
              .width(24)
              .height(24)
              .fillColor(Color.White)
          }
          .width(48)
          .height(48)
          .backgroundColor('#007AFF')
          .borderRadius(24)
          .margin({ top: 8 })
          .onClick(() => {
            this.mapZoom = Math.max(this.mapZoom - 1, 1);
          })
        }
        .position({ x: '90%', y: 20 })
        .translate({ x: '-50%' })
      }
      .width('100%')
      .layoutWeight(1)

      // 轨迹回放控制面板
      Column({ space: 16 }) {
        Text('轨迹回放')
          .fontSize(16)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')

        // 回放进度条
        if (this.locationPoints.length > 0) {
          Column({ space: 8 }) {
            Row() {
              Text('进度')
                .fontSize(14)
                .fontColor('#666666')
                .layoutWeight(1)
              Text(`${this.playbackIndex + 1}/${this.locationPoints.length}`)
                .fontSize(14)
                .fontColor('#333333')
            }
            .width('100%')

            Slider({
              value: this.playbackIndex,
              min: 0,
              max: this.locationPoints.length - 1,
              step: 1
            })
              .width('100%')
              .trackColor('#E0E0E0')
              .selectedColor('#007AFF')
              .blockColor('#007AFF')
              .onChange((value: number) => {
                this.seekToPosition(Math.round(value));
              })
          }
        }

        // 回放控制按钮
        Row({ space: 16 }) {
          Button() {
            Image($r('app.media.ic_skip_previous'))
              .width(20)
              .height(20)
              .fillColor(Color.White)
          }
          .width(44)
          .height(44)
          .backgroundColor('#666666')
          .borderRadius(22)
          .onClick(() => {
            this.resetPlayback();
          })

          Button() {
            Image(this.isPlaying ? $r('app.media.ic_pause') : $r('app.media.ic_play'))
              .width(24)
              .height(24)
              .fillColor(Color.White)
          }
          .width(56)
          .height(56)
          .backgroundColor('#007AFF')
          .borderRadius(28)
          .onClick(() => {
            this.togglePlayback();
          })

          Button() {
            Image($r('app.media.ic_stop'))
              .width(20)
              .height(20)
              .fillColor(Color.White)
          }
          .width(44)
          .height(44)
          .backgroundColor('#666666')
          .borderRadius(22)
          .onClick(() => {
            this.stopPlayback();
            this.showPlaybackControls = false;
          })
        }
        .width('100%')
        .justifyContent(FlexAlign.Center)

        // 回放速度控制
        Row({ space: 8 }) {
          Text('速度:')
            .fontSize(14)
            .fontColor('#666666')

          ForEach([1, 2, 4, 8], (speed: number) => {
            Button(`${speed}x`)
              .fontSize(12)
              .width(40)
              .height(32)
              .backgroundColor(this.playbackSpeed === speed ? '#007AFF' : '#F0F0F0')
              .fontColor(this.playbackSpeed === speed ? Color.White : '#333333')
              .onClick(() => {
                this.setPlaybackSpeed(speed);
              })
          })
        }
        .width('100%')
        .justifyContent(FlexAlign.Center)
      }
      .width('100%')
      .padding(16)
      .backgroundColor(Color.White)
      .borderRadius(8)
      .margin({ top: 16 })
    }
    .width('100%')
    .height('100%')
    .padding(16)
    .backgroundColor('#F8F8F8')
  }

  /**
   * 构建删除确认对话框
   */
  @Builder
  buildDeleteDialog() {
    Stack({ alignContent: Alignment.Center }) {
      // 背景遮罩
      Column()
        .width('100%')
        .height('100%')
        .backgroundColor(Color.Black)
        .opacity(0.5)
        .onClick(() => {
          this.showDeleteDialog = false;
        })

      // 对话框内容
      Column({ space: 20 }) {
        Text('删除轨迹')
          .fontSize(18)
          .fontWeight(FontWeight.Medium)
          .fontColor(Color.Black)

        Text('确定要删除这条轨迹吗？删除后无法恢复。')
          .fontSize(14)
          .fontColor('#666666')
          .textAlign(TextAlign.Center)

        Row({ space: 12 }) {
          Button('取消')
            .width(80)
            .height(40)
            .backgroundColor(Color.Gray)
            .fontColor(Color.White)
            .onClick(() => {
              this.showDeleteDialog = false;
            })

          Button('删除')
            .width(80)
            .height(40)
            .backgroundColor(Color.Red)
            .fontColor(Color.White)
            .onClick(() => {
              this.showDeleteDialog = false;
              this.deleteTrajectory();
            })
        }
        .width('100%')
        .justifyContent(FlexAlign.Center)
      }
      .width('80%')
      .maxWidth(300)
      .padding(24)
      .backgroundColor(Color.White)
      .borderRadius(12)
      .shadow({
        radius: 20,
        color: Color.Black,
        offsetX: 0,
        offsetY: 4
      })
    }
    .width('100%')
    .height('100%')
    .position({ x: 0, y: 0 })
    .zIndex(1000)

    // 导出对话框
    if (this.showExportDialog) {
      this.buildExportDialog()
    }

    // 分享对话框
    if (this.showShareDialog) {
      this.buildShareDialog()
    }
  }

  /**
   * 构建导出对话框
   */
  @Builder
  buildExportDialog() {
    Column({ space: 0 }) {
      // 标题
      Text('导出轨迹')
        .fontSize(18)
        .fontWeight(FontWeight.Medium)
        .margin({ bottom: 20 })

      // 格式选择
      Text('选择导出格式:')
        .fontSize(14)
        .fontColor('#666666')
        .alignSelf(ItemAlign.Start)
        .margin({ bottom: 12 })

      Column({ space: 8 }) {
        ForEach([
          { format: ExportFormat.GPX, name: 'GPX格式', desc: '通用GPS数据格式' },
          { format: ExportFormat.KML, name: 'KML格式', desc: 'Google Earth格式' },
          { format: ExportFormat.JSON, name: 'JSON格式', desc: '结构化数据格式' },
          { format: ExportFormat.CSV, name: 'CSV格式', desc: '表格数据格式' }
        ], (item) => {
          Row({ space: 12 }) {
            Radio({ value: item.format.toString(), group: 'exportFormat' })
              .checked(this.exportFormat === item.format)
              .onChange((isChecked: boolean) => {
                if (isChecked) {
                  this.exportFormat = item.format;
                }
              })

            Column({ space: 2 }) {
              Text(item.name)
                .fontSize(16)
                .fontColor(Color.Black)
              Text(item.desc)
                .fontSize(12)
                .fontColor('#999999')
            }
            .alignItems(HorizontalAlign.Start)
          }
          .width('100%')
          .padding({ vertical: 8 })
        })
      }
      .margin({ bottom: 24 })

      // 按钮
      Row({ space: 12 }) {
        Button('取消')
          .width(80)
          .height(40)
          .backgroundColor(Color.Gray)
          .fontColor(Color.White)
          .onClick(() => {
            this.showExportDialog = false;
          })

        Button(this.isExporting ? '导出中...' : '导出')
          .width(80)
          .height(40)
          .backgroundColor('#007AFF')
          .fontColor(Color.White)
          .enabled(!this.isExporting)
          .onClick(() => {
            this.exportTrajectory();
          })
      }
      .width('100%')
      .justifyContent(FlexAlign.Center)
    }
    .width('80%')
    .maxWidth(300)
    .padding(24)
    .backgroundColor(Color.White)
    .borderRadius(12)
    .shadow({
      radius: 20,
      color: Color.Black,
      offsetX: 0,
      offsetY: 4
    })
    .position({ x: '50%', y: '50%' })
    .translate({ x: '-50%', y: '-50%' })
    .zIndex(1001)
  }

  /**
   * 构建分享对话框
   */
  @Builder
  buildShareDialog() {
    Column({ space: 0 }) {
      // 标题
      Text('分享轨迹')
        .fontSize(18)
        .fontWeight(FontWeight.Medium)
        .margin({ bottom: 20 })

      // 格式选择
      Text('选择分享格式:')
        .fontSize(14)
        .fontColor('#666666')
        .alignSelf(ItemAlign.Start)
        .margin({ bottom: 12 })

      Column({ space: 8 }) {
        ForEach([
          { format: ExportFormat.GPX, name: 'GPX格式', desc: '通用GPS数据格式' },
          { format: ExportFormat.KML, name: 'KML格式', desc: 'Google Earth格式' },
          { format: ExportFormat.JSON, name: 'JSON格式', desc: '结构化数据格式' }
        ], (item) => {
          Row({ space: 12 }) {
            Radio({ value: item.format.toString(), group: 'shareFormat' })
              .checked(this.exportFormat === item.format)
              .onChange((isChecked: boolean) => {
                if (isChecked) {
                  this.exportFormat = item.format;
                }
              })

            Column({ space: 2 }) {
              Text(item.name)
                .fontSize(16)
                .fontColor(Color.Black)
              Text(item.desc)
                .fontSize(12)
                .fontColor('#999999')
            }
            .alignItems(HorizontalAlign.Start)
          }
          .width('100%')
          .padding({ vertical: 8 })
        })
      }
      .margin({ bottom: 24 })

      // 按钮
      Row({ space: 12 }) {
        Button('取消')
          .width(80)
          .height(40)
          .backgroundColor(Color.Gray)
          .fontColor(Color.White)
          .onClick(() => {
            this.showShareDialog = false;
          })

        Button(this.isExporting ? '分享中...' : '分享')
          .width(80)
          .height(40)
          .backgroundColor('#34C759')
          .fontColor(Color.White)
          .enabled(!this.isExporting)
          .onClick(() => {
            this.shareTrajectory();
          })
      }
      .width('100%')
      .justifyContent(FlexAlign.Center)
    }
    .width('80%')
    .maxWidth(300)
    .padding(24)
    .backgroundColor(Color.White)
    .borderRadius(12)
    .shadow({
      radius: 20,
      color: Color.Black,
      offsetX: 0,
      offsetY: 4
    })
    .position({ x: '50%', y: '50%' })
    .translate({ x: '-50%', y: '-50%' })
    .zIndex(1001)
  }
}