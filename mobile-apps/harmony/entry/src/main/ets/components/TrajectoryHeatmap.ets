import { LocationPoint } from '../model/LocationPoint';
import { DataService } from '../service/DataService';

/**
 * 热力图数据点
 */
export interface HeatmapPoint {
  latitude: number;
  longitude: number;
  intensity: number; // 强度值 0-1
}

/**
 * 网格单元
 */
interface GridCell {
  lat: number;
  lng: number;
  count: number;
  intensity: number;
}

/**
 * 轨迹热力图组件
 */
@Component
export struct TrajectoryHeatmap {
  @Prop timeRange: string = '30d'; // 7d, 30d, 90d, 1y, all
  @State private heatmapData: HeatmapPoint[] = [];
  @State private isLoading: boolean = true;
  @State private mapCenter: { lat: number, lng: number } = { lat: 39.9042, lng: 116.4074 };
  @State private mapZoom: number = 10;
  @State private gridSize: number = 0.01; // 网格大小（度）
  @State private maxIntensity: number = 0;

  private dataService: DataService = DataService.getInstance();

  aboutToAppear() {
    this.loadHeatmapData();
  }

  aboutToUpdate() {
    this.loadHeatmapData();
  }

  /**
   * 加载热力图数据
   */
  private async loadHeatmapData() {
    try {
      this.isLoading = true;
      
      // 计算时间范围
      const endTime = Date.now();
      let startTime = 0;
      
      switch (this.timeRange) {
        case '7d':
          startTime = endTime - 7 * 24 * 60 * 60 * 1000;
          break;
        case '30d':
          startTime = endTime - 30 * 24 * 60 * 60 * 1000;
          break;
        case '90d':
          startTime = endTime - 90 * 24 * 60 * 60 * 1000;
          break;
        case '1y':
          startTime = endTime - 365 * 24 * 60 * 60 * 1000;
          break;
        case 'all':
        default:
          startTime = 0;
          break;
      }
      
      // 获取轨迹数据
      const trajectories = await this.dataService.getTrajectories({
        startTime,
        endTime,
        limit: 100
      });
      
      // 收集所有位置点
      const allPoints: LocationPoint[] = [];
      for (const trajectory of trajectories) {
        try {
          const points = await this.dataService.getLocationPointsByTrajectory(trajectory.id);
          allPoints.push(...points);
        } catch (error) {
          console.error('Failed to load points for trajectory:', trajectory.id, error);
        }
      }
      
      // 生成热力图数据
      this.heatmapData = this.generateHeatmapData(allPoints);
      this.calculateMapBounds(allPoints);
      
    } catch (error) {
      console.error('Failed to load heatmap data:', error);
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * 生成热力图数据
   */
  private generateHeatmapData(points: LocationPoint[]): HeatmapPoint[] {
    if (points.length === 0) {
      return [];
    }
    
    // 创建网格
    const gridMap = new Map<string, GridCell>();
    
    // 将点分配到网格中
    for (const point of points) {
      const gridLat = Math.floor(point.latitude / this.gridSize) * this.gridSize;
      const gridLng = Math.floor(point.longitude / this.gridSize) * this.gridSize;
      const key = `${gridLat},${gridLng}`;
      
      if (gridMap.has(key)) {
        gridMap.get(key)!.count++;
      } else {
        gridMap.set(key, {
          lat: gridLat + this.gridSize / 2,
          lng: gridLng + this.gridSize / 2,
          count: 1,
          intensity: 0
        });
      }
    }
    
    // 计算最大计数
    const maxCount = Math.max(...Array.from(gridMap.values()).map(cell => cell.count));
    this.maxIntensity = maxCount;
    
    // 计算强度并转换为热力图点
    const heatmapPoints: HeatmapPoint[] = [];
    for (const cell of gridMap.values()) {
      cell.intensity = cell.count / maxCount;
      heatmapPoints.push({
        latitude: cell.lat,
        longitude: cell.lng,
        intensity: cell.intensity
      });
    }
    
    return heatmapPoints;
  }

  /**
   * 计算地图边界
   */
  private calculateMapBounds(points: LocationPoint[]) {
    if (points.length === 0) {
      return;
    }
    
    let minLat = points[0].latitude;
    let maxLat = points[0].latitude;
    let minLng = points[0].longitude;
    let maxLng = points[0].longitude;
    
    for (const point of points) {
      minLat = Math.min(minLat, point.latitude);
      maxLat = Math.max(maxLat, point.latitude);
      minLng = Math.min(minLng, point.longitude);
      maxLng = Math.max(maxLng, point.longitude);
    }
    
    // 计算中心点
    this.mapCenter = {
      lat: (minLat + maxLat) / 2,
      lng: (minLng + maxLng) / 2
    };
    
    // 计算缩放级别
    const latDiff = maxLat - minLat;
    const lngDiff = maxLng - minLng;
    const maxDiff = Math.max(latDiff, lngDiff);
    
    if (maxDiff > 1) {
      this.mapZoom = 8;
    } else if (maxDiff > 0.1) {
      this.mapZoom = 10;
    } else if (maxDiff > 0.01) {
      this.mapZoom = 12;
    } else {
      this.mapZoom = 14;
    }
  }

  /**
   * 获取热力图颜色
   */
  private getHeatmapColor(intensity: number): string {
    if (intensity <= 0.2) {
      return `rgba(0, 255, 0, ${intensity * 0.6})`;
    } else if (intensity <= 0.4) {
      return `rgba(255, 255, 0, ${intensity * 0.7})`;
    } else if (intensity <= 0.6) {
      return `rgba(255, 165, 0, ${intensity * 0.8})`;
    } else if (intensity <= 0.8) {
      return `rgba(255, 69, 0, ${intensity * 0.9})`;
    } else {
      return `rgba(255, 0, 0, ${intensity})`;
    }
  }

  /**
   * 获取时间范围显示名称
   */
  private getTimeRangeDisplayName(): string {
    switch (this.timeRange) {
      case '7d':
        return '最近7天';
      case '30d':
        return '最近30天';
      case '90d':
        return '最近90天';
      case '1y':
        return '最近1年';
      case 'all':
        return '全部时间';
      default:
        return '未知';
    }
  }

  build() {
    Column({ space: 12 }) {
      // 标题和说明
      Row() {
        Column({ space: 4 }) {
          Text('活动热力图')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .alignSelf(ItemAlign.Start)
          
          Text(`${this.getTimeRangeDisplayName()} · ${this.heatmapData.length} 个活动区域`)
            .fontSize(12)
            .fontColor('#666666')
            .alignSelf(ItemAlign.Start)
        }
        .layoutWeight(1)
        .alignItems(HorizontalAlign.Start)

        // 图例
        Row({ space: 8 }) {
          Text('活动强度:')
            .fontSize(10)
            .fontColor('#666666')
          
          Row({ space: 2 }) {
            ForEach([0.2, 0.4, 0.6, 0.8, 1.0], (intensity: number) => {
              Circle({ width: 8, height: 8 })
                .fill(this.getHeatmapColor(intensity))
            })
          }
          
          Text('高')
            .fontSize(10)
            .fontColor('#666666')
        }
      }
      .width('100%')

      if (this.isLoading) {
        // 加载状态
        Column({ space: 12 }) {
          LoadingProgress()
            .width(40)
            .height(40)
            .color('#007AFF')

          Text('生成热力图中...')
            .fontSize(14)
            .fontColor('#666666')
        }
        .width('100%')
        .height(300)
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
        .backgroundColor('#F8F8F8')
        .borderRadius(8)

      } else if (this.heatmapData.length === 0) {
        // 无数据状态
        Column({ space: 12 }) {
          Image($r('app.media.ic_map_empty'))
            .width(60)
            .height(60)
            .fillColor('#CCCCCC')

          Text('暂无轨迹数据')
            .fontSize(14)
            .fontColor('#666666')

          Text('开始记录轨迹以查看活动热力图')
            .fontSize(12)
            .fontColor('#999999')
        }
        .width('100%')
        .height(300)
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
        .backgroundColor('#F8F8F8')
        .borderRadius(8)

      } else {
        // 热力图显示
        Stack() {
          // 地图背景（模拟）
          Column() {
            Text('热力图显示区域')
              .fontSize(14)
              .fontColor('#999999')
            
            Text(`中心: ${this.mapCenter.lat.toFixed(4)}, ${this.mapCenter.lng.toFixed(4)}`)
              .fontSize(10)
              .fontColor('#CCCCCC')
              .margin({ top: 4 })
            
            Text(`缩放: ${this.mapZoom} · 网格: ${this.gridSize}°`)
              .fontSize(10)
              .fontColor('#CCCCCC')
              .margin({ top: 2 })
          }
          .width('100%')
          .height('100%')
          .justifyContent(FlexAlign.Center)
          .alignItems(HorizontalAlign.Center)
          .backgroundColor('#E8F4FD')
          .border({
            width: 1,
            color: '#B3D9F2',
            style: BorderStyle.Dashed
          })

          // 热力图点（简化显示）
          Grid() {
            ForEach(this.heatmapData.slice(0, 50), (point: HeatmapPoint, index: number) => {
              GridItem() {
                Circle({ width: 12, height: 12 })
                  .fill(this.getHeatmapColor(point.intensity))
                  .opacity(0.8)
              }
              .width(20)
              .height(20)
            })
          }
          .columnsTemplate('1fr 1fr 1fr 1fr 1fr 1fr 1fr 1fr 1fr 1fr')
          .rowsGap(8)
          .columnsGap(8)
          .width('90%')
          .height('80%')

          // 统计信息
          Column({ space: 4 }) {
            Text(`最高强度: ${this.maxIntensity}`)
              .fontSize(10)
              .fontColor('#333333')
              .padding({ left: 6, right: 6, top: 2, bottom: 2 })
              .backgroundColor('rgba(255, 255, 255, 0.9)')
              .borderRadius(4)
            
            Text(`活动点: ${this.heatmapData.length}`)
              .fontSize(10)
              .fontColor('#333333')
              .padding({ left: 6, right: 6, top: 2, bottom: 2 })
              .backgroundColor('rgba(255, 255, 255, 0.9)')
              .borderRadius(4)
          }
          .position({ x: 12, y: 12 })
          .alignItems(HorizontalAlign.Start)
        }
        .width('100%')
        .height(300)
        .borderRadius(8)
        .clip(true)
      }

      // 控制选项
      Row({ space: 12 }) {
        Button('调整网格')
          .layoutWeight(1)
          .height(36)
          .fontSize(12)
          .backgroundColor('#F8F8F8')
          .fontColor('#333333')
          .onClick(() => {
            // 调整网格大小
            this.gridSize = this.gridSize === 0.01 ? 0.005 : 0.01;
            this.loadHeatmapData();
          })

        Button('重新生成')
          .layoutWeight(1)
          .height(36)
          .fontSize(12)
          .backgroundColor('#007AFF')
          .fontColor(Color.White)
          .onClick(() => {
            this.loadHeatmapData();
          })
      }
      .width('100%')
    }
    .width('100%')
    .padding(16)
    .backgroundColor(Color.White)
    .borderRadius(8)
  }
}