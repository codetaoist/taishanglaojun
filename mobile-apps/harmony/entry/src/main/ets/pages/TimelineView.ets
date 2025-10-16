import { LocationPoint } from '../data/LocationPoint';
import { Trajectory } from '../data/Trajectory';
import { LocationViewModel } from '../viewmodel/LocationViewModel';

/**
 * 时间轴视图
 * 显示轨迹的时间序列和详细信息
 */
@Component
export struct TimelineView {
  @Prop trajectory: Trajectory | null = null;
  @State private locationViewModel: LocationViewModel = new LocationViewModel();
  @State private trajectoryPoints: LocationPoint[] = [];
  @State private isLoading: boolean = false;
  @State private expandedItems: Set<string> = new Set();
  
  aboutToAppear() {
    this.loadTrajectoryPoints();
  }
  
  /**
   * 监听轨迹变化
   */
  aboutToUpdate() {
    this.loadTrajectoryPoints();
  }
  
  /**
   * 加载轨迹点数据
   */
  private async loadTrajectoryPoints(): Promise<void> {
    if (!this.trajectory) return;
    
    this.isLoading = true;
    try {
      this.trajectoryPoints = await this.locationViewModel.getTrajectoryPoints(this.trajectory.id);
    } catch (error) {
      console.error('Failed to load trajectory points:', error);
    } finally {
      this.isLoading = false;
    }
  }
  
  /**
   * 切换项目展开状态
   */
  private toggleExpanded(pointId: string): void {
    if (this.expandedItems.has(pointId)) {
      this.expandedItems.delete(pointId);
    } else {
      this.expandedItems.add(pointId);
    }
  }
  
  /**
   * 格式化时间
   */
  private formatTime(timestamp: number): string {
    const date = new Date(timestamp);
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}:${date.getSeconds().toString().padStart(2, '0')}`;
  }
  
  /**
   * 格式化日期
   */
  private formatDate(timestamp: number): string {
    const date = new Date(timestamp);
    return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')}`;
  }
  
  /**
   * 格式化坐标
   */
  private formatCoordinate(value: number): string {
    return value.toFixed(6);
  }
  
  /**
   * 格式化速度
   */
  private formatSpeed(speed?: number): string {
    if (speed === undefined || speed === null) return '未知';
    return `${(speed * 3.6).toFixed(1)} km/h`; // 转换为 km/h
  }
  
  /**
   * 格式化精度
   */
  private formatAccuracy(accuracy?: number): string {
    if (accuracy === undefined || accuracy === null) return '未知';
    return `±${accuracy.toFixed(1)}m`;
  }
  
  /**
   * 计算两点间距离
   */
  private calculateDistance(point1: LocationPoint, point2: LocationPoint): number {
    const R = 6371000; // 地球半径（米）
    const lat1Rad = point1.latitude * Math.PI / 180;
    const lat2Rad = point2.latitude * Math.PI / 180;
    const deltaLatRad = (point2.latitude - point1.latitude) * Math.PI / 180;
    const deltaLngRad = (point2.longitude - point1.longitude) * Math.PI / 180;
    
    const a = Math.sin(deltaLatRad / 2) * Math.sin(deltaLatRad / 2) +
              Math.cos(lat1Rad) * Math.cos(lat2Rad) *
              Math.sin(deltaLngRad / 2) * Math.sin(deltaLngRad / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    
    return R * c;
  }
  
  /**
   * 获取点的状态颜色
   */
  private getPointStatusColor(point: LocationPoint): Color {
    if (point.accuracy && point.accuracy > 50) {
      return Color.Orange; // 精度较低
    }
    if (point.speed && point.speed > 0) {
      return Color.Green; // 移动中
    }
    return Color.Blue; // 静止
  }
  
  build() {
    Column() {
      if (!this.trajectory) {
        // 空状态
        Column() {
          Image($r('app.media.ic_timeline'))
            .width(64)
            .height(64)
            .fillColor(Color.Gray)
            .margin({ bottom: 16 })
          
          Text('选择一个轨迹查看时间轴')
            .fontSize(16)
            .fontColor(Color.Gray)
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
      } else {
        // 轨迹信息头部
        Column() {
          Text(this.trajectory.name || '未命名轨迹')
            .fontSize(18)
            .fontWeight(FontWeight.Bold)
            .margin({ bottom: 8 })
          
          Row() {
            Text(`开始: ${this.formatDate(this.trajectory.startTime)} ${this.formatTime(this.trajectory.startTime)}`)
              .fontSize(12)
              .fontColor(Color.Gray)
              .layoutWeight(1)
            
            if (this.trajectory.endTime) {
              Text(`结束: ${this.formatTime(this.trajectory.endTime)}`)
                .fontSize(12)
                .fontColor(Color.Gray)
            }
          }
          .width('100%')
          
          if (this.trajectory.endTime) {
            Row() {
              Text(`总时长: ${Math.floor((this.trajectory.endTime - this.trajectory.startTime) / 60000)}分钟`)
                .fontSize(12)
                .fontColor(Color.Gray)
                .layoutWeight(1)
              
              Text(`总距离: ${(this.trajectory.totalDistance / 1000).toFixed(2)}km`)
                .fontSize(12)
                .fontColor(Color.Gray)
            }
            .width('100%')
            .margin({ top: 4 })
          }
        }
        .width('100%')
        .padding(16)
        .backgroundColor(Color.White)
        .borderRadius(8)
        .margin({ bottom: 8 })
        
        // 时间轴列表
        if (this.isLoading) {
          Column() {
            LoadingProgress()
              .width(32)
              .height(32)
              .margin({ bottom: 16 })
            
            Text('加载轨迹点...')
              .fontSize(14)
              .fontColor(Color.Gray)
          }
          .width('100%')
          .height(200)
          .justifyContent(FlexAlign.Center)
          .alignItems(HorizontalAlign.Center)
        } else {
          List() {
            ForEach(this.trajectoryPoints, (point: LocationPoint, index: number) => {
              ListItem() {
                this.TimelineItemBuilder(point, index)
              }
            })
          }
          .width('100%')
          .layoutWeight(1)
          .divider({
            strokeWidth: 1,
            color: '#F0F0F0',
            startMargin: 60,
            endMargin: 16
          })
        }
      }
    }
    .width('100%')
    .height('100%')
    .padding(16)
    .backgroundColor('#F5F5F5')
  }
  
  /**
   * 时间轴项目构建器
   */
  @Builder
  private TimelineItemBuilder(point: LocationPoint, index: number) {
    Row() {
      // 时间轴线条和圆点
      Column() {
        if (index > 0) {
          Line()
            .width(2)
            .height(20)
            .stroke(Color.Gray)
        }
        
        Circle({ width: 12, height: 12 })
          .fill(this.getPointStatusColor(point))
          .stroke(Color.White)
          .strokeWidth(2)
        
        if (index < this.trajectoryPoints.length - 1) {
          Line()
            .width(2)
            .height(20)
            .stroke(Color.Gray)
        }
      }
      .width(44)
      .alignItems(HorizontalAlign.Center)
      
      // 内容区域
      Column() {
        Row() {
          Column() {
            Text(this.formatTime(point.timestamp))
              .fontSize(14)
              .fontWeight(FontWeight.Medium)
              .margin({ bottom: 2 })
            
            Text(`${this.formatCoordinate(point.latitude)}, ${this.formatCoordinate(point.longitude)}`)
              .fontSize(12)
              .fontColor(Color.Gray)
          }
          .alignItems(HorizontalAlign.Start)
          .layoutWeight(1)
          
          // 展开/收起按钮
          Button() {
            Image($r('app.media.ic_expand_more'))
              .width(16)
              .height(16)
              .fillColor(Color.Gray)
              .rotate({
                angle: this.expandedItems.has(point.id) ? 180 : 0
              })
              .animation({
                duration: 200,
                curve: Curve.EaseInOut
              })
          }
          .type(ButtonType.Circle)
          .width(32)
          .height(32)
          .backgroundColor(Color.Transparent)
          .onClick(() => this.toggleExpanded(point.id))
        }
        .width('100%')
        
        // 展开的详细信息
        if (this.expandedItems.has(point.id)) {
          Column() {
            Divider()
              .margin({ top: 8, bottom: 8 })
            
            Row() {
              Text('速度:')
                .fontSize(12)
                .fontColor(Color.Gray)
                .width(60)
              
              Text(this.formatSpeed(point.speed))
                .fontSize(12)
                .fontColor(Color.Black)
                .layoutWeight(1)
            }
            .width('100%')
            .margin({ bottom: 4 })
            
            Row() {
              Text('精度:')
                .fontSize(12)
                .fontColor(Color.Gray)
                .width(60)
              
              Text(this.formatAccuracy(point.accuracy))
                .fontSize(12)
                .fontColor(Color.Black)
                .layoutWeight(1)
            }
            .width('100%')
            .margin({ bottom: 4 })
            
            if (point.altitude !== undefined && point.altitude !== null) {
              Row() {
                Text('海拔:')
                  .fontSize(12)
                  .fontColor(Color.Gray)
                  .width(60)
                
                Text(`${point.altitude.toFixed(1)}m`)
                  .fontSize(12)
                  .fontColor(Color.Black)
                  .layoutWeight(1)
              }
              .width('100%')
              .margin({ bottom: 4 })
            }
            
            if (index > 0) {
              const prevPoint = this.trajectoryPoints[index - 1];
              const distance = this.calculateDistance(prevPoint, point);
              const timeDiff = (point.timestamp - prevPoint.timestamp) / 1000; // 秒
              
              Row() {
                Text('距离:')
                  .fontSize(12)
                  .fontColor(Color.Gray)
                  .width(60)
                
                Text(`${distance.toFixed(1)}m`)
                  .fontSize(12)
                  .fontColor(Color.Black)
                  .layoutWeight(1)
              }
              .width('100%')
              .margin({ bottom: 4 })
              
              Row() {
                Text('间隔:')
                  .fontSize(12)
                  .fontColor(Color.Gray)
                  .width(60)
                
                Text(`${timeDiff.toFixed(0)}秒`)
                  .fontSize(12)
                  .fontColor(Color.Black)
                  .layoutWeight(1)
              }
              .width('100%')
            }
          }
          .width('100%')
          .animation({
            duration: 200,
            curve: Curve.EaseInOut
          })
        }
      }
      .layoutWeight(1)
      .padding({ left: 12, right: 16, top: 8, bottom: 8 })
      .backgroundColor(Color.White)
      .borderRadius(8)
    }
    .width('100%')
    .alignItems(VerticalAlign.Top)
  }
}