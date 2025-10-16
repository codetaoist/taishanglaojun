import { Trajectory } from '../model/Trajectory';
import { LocationPoint } from '../model/LocationPoint';
import { DataService } from '../service/DataService';
import { LocationUtils } from '../utils/LocationUtils';
import { TrajectoryChart } from './TrajectoryChart';
import { TrajectoryHeatmap } from './TrajectoryHeatmap';

/**
 * 轨迹统计数据接口
 */
export interface TrajectoryStats {
  totalTrajectories: number;
  totalDistance: number;
  totalDuration: number;
  averageSpeed: number;
  maxSpeed: number;
  totalPoints: number;
  averageAccuracy: number;
  elevationGain: number;
  elevationLoss: number;
  mostActiveDay: string;
  mostActiveHour: number;
  longestTrajectory: Trajectory | null;
  fastestTrajectory: Trajectory | null;
  dailyStats: DailyStats[];
  hourlyStats: HourlyStats[];
  monthlyStats: MonthlyStats[];
}

/**
 * 每日统计数据
 */
export interface DailyStats {
  date: string;
  trajectoryCount: number;
  totalDistance: number;
  totalDuration: number;
  averageSpeed: number;
}

/**
 * 每小时统计数据
 */
export interface HourlyStats {
  hour: number;
  trajectoryCount: number;
  totalDistance: number;
  averageSpeed: number;
}

/**
 * 每月统计数据
 */
export interface MonthlyStats {
  month: string;
  trajectoryCount: number;
  totalDistance: number;
  totalDuration: number;
  averageSpeed: number;
}

/**
 * 轨迹统计分析组件
 */
@Component
export struct TrajectoryAnalytics {
  @State private stats: TrajectoryStats | null = null;
  @State private isLoading: boolean = true;
  @State private selectedPeriod: string = '7d'; // 7d, 30d, 90d, 1y, all
  @State private selectedChart: string = 'distance'; // distance, speed, duration, count
  @State private chartType: string = 'bar'; // line, bar

  private dataService: DataService = DataService.getInstance();

  aboutToAppear() {
    this.loadStatistics();
  }

  /**
   * 加载统计数据
   */
  private async loadStatistics() {
    try {
      this.isLoading = true;
      
      // 计算时间范围
      const endTime = Date.now();
      let startTime = 0;
      
      switch (this.selectedPeriod) {
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
        limit: 1000
      });
      
      // 计算统计数据
      this.stats = await this.calculateStatistics(trajectories);
      
    } catch (error) {
      console.error('Failed to load statistics:', error);
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * 计算统计数据
   */
  private async calculateStatistics(trajectories: Trajectory[]): Promise<TrajectoryStats> {
    let totalDistance = 0;
    let totalDuration = 0;
    let totalPoints = 0;
    let totalAccuracy = 0;
    let maxSpeed = 0;
    let elevationGain = 0;
    let elevationLoss = 0;
    let longestTrajectory: Trajectory | null = null;
    let fastestTrajectory: Trajectory | null = null;
    
    const dailyStatsMap = new Map<string, DailyStats>();
    const hourlyStatsMap = new Map<number, HourlyStats>();
    const monthlyStatsMap = new Map<string, MonthlyStats>();
    
    for (const trajectory of trajectories) {
      // 累计基本统计
      totalDistance += trajectory.totalDistance;
      totalDuration += trajectory.endTime - trajectory.startTime;
      
      if (trajectory.maxSpeed && trajectory.maxSpeed > maxSpeed) {
        maxSpeed = trajectory.maxSpeed;
        fastestTrajectory = trajectory;
      }
      
      if (!longestTrajectory || trajectory.totalDistance > longestTrajectory.totalDistance) {
        longestTrajectory = trajectory;
      }
      
      // 获取轨迹点进行详细分析
      try {
        const points = await this.dataService.getLocationPointsByTrajectory(trajectory.id);
        totalPoints += points.length;
        
        // 计算精度和海拔统计
        let lastElevation: number | null = null;
        for (const point of points) {
          if (point.accuracy) {
            totalAccuracy += point.accuracy;
          }
          
          if (point.altitude && lastElevation !== null) {
            const elevationDiff = point.altitude - lastElevation;
            if (elevationDiff > 0) {
              elevationGain += elevationDiff;
            } else {
              elevationLoss += Math.abs(elevationDiff);
            }
          }
          lastElevation = point.altitude || lastElevation;
        }
      } catch (error) {
        console.error('Failed to load points for trajectory:', trajectory.id, error);
      }
      
      // 按日期统计
      const date = new Date(trajectory.startTime).toISOString().split('T')[0];
      const dailyStat = dailyStatsMap.get(date) || {
        date,
        trajectoryCount: 0,
        totalDistance: 0,
        totalDuration: 0,
        averageSpeed: 0
      };
      dailyStat.trajectoryCount++;
      dailyStat.totalDistance += trajectory.totalDistance;
      dailyStat.totalDuration += trajectory.endTime - trajectory.startTime;
      dailyStat.averageSpeed = dailyStat.totalDistance / (dailyStat.totalDuration / 1000);
      dailyStatsMap.set(date, dailyStat);
      
      // 按小时统计
      const hour = new Date(trajectory.startTime).getHours();
      const hourlyStat = hourlyStatsMap.get(hour) || {
        hour,
        trajectoryCount: 0,
        totalDistance: 0,
        averageSpeed: 0
      };
      hourlyStat.trajectoryCount++;
      hourlyStat.totalDistance += trajectory.totalDistance;
      hourlyStat.averageSpeed = trajectory.averageSpeed || 0;
      hourlyStatsMap.set(hour, hourlyStat);
      
      // 按月统计
      const month = new Date(trajectory.startTime).toISOString().substring(0, 7);
      const monthlyStat = monthlyStatsMap.get(month) || {
        month,
        trajectoryCount: 0,
        totalDistance: 0,
        totalDuration: 0,
        averageSpeed: 0
      };
      monthlyStat.trajectoryCount++;
      monthlyStat.totalDistance += trajectory.totalDistance;
      monthlyStat.totalDuration += trajectory.endTime - trajectory.startTime;
      monthlyStat.averageSpeed = monthlyStat.totalDistance / (monthlyStat.totalDuration / 1000);
      monthlyStatsMap.set(month, monthlyStat);
    }
    
    // 找出最活跃的日期和小时
    let mostActiveDay = '';
    let maxDailyCount = 0;
    for (const [date, stat] of dailyStatsMap) {
      if (stat.trajectoryCount > maxDailyCount) {
        maxDailyCount = stat.trajectoryCount;
        mostActiveDay = date;
      }
    }
    
    let mostActiveHour = 0;
    let maxHourlyCount = 0;
    for (const [hour, stat] of hourlyStatsMap) {
      if (stat.trajectoryCount > maxHourlyCount) {
        maxHourlyCount = stat.trajectoryCount;
        mostActiveHour = hour;
      }
    }
    
    return {
      totalTrajectories: trajectories.length,
      totalDistance,
      totalDuration,
      averageSpeed: totalDistance / (totalDuration / 1000),
      maxSpeed,
      totalPoints,
      averageAccuracy: totalPoints > 0 ? totalAccuracy / totalPoints : 0,
      elevationGain,
      elevationLoss,
      mostActiveDay,
      mostActiveHour,
      longestTrajectory,
      fastestTrajectory,
      dailyStats: Array.from(dailyStatsMap.values()).sort((a, b) => a.date.localeCompare(b.date)),
      hourlyStats: Array.from(hourlyStatsMap.values()).sort((a, b) => a.hour - b.hour),
      monthlyStats: Array.from(monthlyStatsMap.values()).sort((a, b) => a.month.localeCompare(b.month))
    };
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
   * 格式化持续时间
   */
  private formatDuration(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) {
      return `${days}天${hours % 24}小时`;
    } else if (hours > 0) {
      return `${hours}小时${minutes % 60}分钟`;
    } else if (minutes > 0) {
      return `${minutes}分钟`;
    } else {
      return `${seconds}秒`;
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
   * 获取时间段显示名称
   */
  private getPeriodDisplayName(period: string): string {
    switch (period) {
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

  /**
   * 获取图表类型显示名称
   */
  private getChartDisplayName(chart: string): string {
    switch (chart) {
      case 'distance':
        return '距离';
      case 'speed':
        return '速度';
      case 'duration':
        return '时长';
      case 'count':
        return '次数';
      default:
        return '未知';
    }
  }

  build() {
    Column() {
      // 时间段选择
      Row({ space: 8 }) {
        ForEach(['7d', '30d', '90d', '1y', 'all'], (period: string) => {
          Text(this.getPeriodDisplayName(period))
            .fontSize(14)
            .fontColor(this.selectedPeriod === period ? Color.White : '#007AFF')
            .backgroundColor(this.selectedPeriod === period ? '#007AFF' : Color.Transparent)
            .padding({ left: 12, right: 12, top: 6, bottom: 6 })
            .borderRadius(16)
            .border({
              width: 1,
              color: '#007AFF'
            })
            .onClick(() => {
              this.selectedPeriod = period;
              this.loadStatistics();
            })
        })
      }
      .width('100%')
      .padding({ left: 16, right: 16, top: 12, bottom: 12 })
      .backgroundColor(Color.White)
      .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })

      if (this.isLoading) {
        // 加载状态
        Column({ space: 16 }) {
          LoadingProgress()
            .width(40)
            .height(40)
            .color('#007AFF')

          Text('分析中...')
            .fontSize(14)
            .fontColor(Color.Gray)
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)

      } else if (!this.stats) {
        // 错误状态
        Column({ space: 16 }) {
          Image($r('app.media.ic_error'))
            .width(60)
            .height(60)
            .fillColor(Color.Gray)

          Text('加载失败')
            .fontSize(16)
            .fontColor(Color.Gray)

          Button('重试')
            .width(100)
            .height(40)
            .backgroundColor('#007AFF')
            .fontColor(Color.White)
            .onClick(() => this.loadStatistics())
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)

      } else if (this.stats.totalTrajectories === 0) {
        // 空状态
        Column({ space: 16 }) {
          Image($r('app.media.ic_trajectory_empty'))
            .width(80)
            .height(80)
            .fillColor(Color.Gray)

          Text('暂无轨迹数据')
            .fontSize(16)
            .fontColor(Color.Gray)

          Text('开始记录您的第一条轨迹吧')
            .fontSize(14)
            .fontColor('#999999')
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)

      } else {
        // 统计内容
        Scroll() {
          Column({ space: 16 }) {
            // 总览统计卡片
            this.buildOverviewCards()
            
            // 图表选择
            this.buildChartSelector()
            
            // 图表显示
            this.buildChart()
            
            // 热力图显示
            this.buildHeatmap()
            
            // 详细统计
            this.buildDetailedStats()
            
            // 记录排行
            this.buildRecords()
          }
          .width('100%')
          .padding(16)
        }
        .width('100%')
        .layoutWeight(1)
      }
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F8F8F8')
  }

  /**
   * 构建总览统计卡片
   */
  @Builder
  buildOverviewCards() {
    Row({ space: 12 }) {
      // 总轨迹数
      Column({ space: 8 }) {
        Text(`${this.stats!.totalTrajectories}`)
          .fontSize(24)
          .fontWeight(FontWeight.Bold)
          .fontColor('#007AFF')
        
        Text('总轨迹数')
          .fontSize(12)
          .fontColor('#666666')
      }
      .layoutWeight(1)
      .padding(16)
      .backgroundColor(Color.White)
      .borderRadius(8)
      .alignItems(HorizontalAlign.Center)

      // 总距离
      Column({ space: 8 }) {
        Text(this.formatDistance(this.stats!.totalDistance))
          .fontSize(16)
          .fontWeight(FontWeight.Bold)
          .fontColor('#34C759')
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
        
        Text('总距离')
          .fontSize(12)
          .fontColor('#666666')
      }
      .layoutWeight(1)
      .padding(16)
      .backgroundColor(Color.White)
      .borderRadius(8)
      .alignItems(HorizontalAlign.Center)

      // 总时长
      Column({ space: 8 }) {
        Text(this.formatDuration(this.stats!.totalDuration))
          .fontSize(16)
          .fontWeight(FontWeight.Bold)
          .fontColor('#FF9500')
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
        
        Text('总时长')
          .fontSize(12)
          .fontColor('#666666')
      }
      .layoutWeight(1)
      .padding(16)
      .backgroundColor(Color.White)
      .borderRadius(8)
      .alignItems(HorizontalAlign.Center)
    }
    .width('100%')
  }

  /**
   * 构建图表选择器
   */
  @Builder
  buildChartSelector() {
    Column({ space: 12 }) {
      // 数据类型选择
      Row({ space: 8 }) {
        ForEach(['distance', 'speed', 'duration', 'count'], (chart: string) => {
          Text(this.getChartDisplayName(chart))
            .fontSize(14)
            .fontColor(this.selectedChart === chart ? Color.White : '#007AFF')
            .backgroundColor(this.selectedChart === chart ? '#007AFF' : Color.Transparent)
            .padding({ left: 12, right: 12, top: 6, bottom: 6 })
            .borderRadius(16)
            .border({
              width: 1,
              color: '#007AFF'
            })
            .onClick(() => {
              this.selectedChart = chart;
            })
        })
      }
      .width('100%')
      
      // 图表类型选择
      Row({ space: 8 }) {
        Text('图表类型:')
          .fontSize(14)
          .fontColor('#666666')
        
        ForEach(['bar', 'line'], (type: string) => {
          Text(type === 'bar' ? '柱状图' : '折线图')
            .fontSize(12)
            .fontColor(this.chartType === type ? Color.White : '#007AFF')
            .backgroundColor(this.chartType === type ? '#007AFF' : Color.Transparent)
            .padding({ left: 8, right: 8, top: 4, bottom: 4 })
            .borderRadius(12)
            .border({
              width: 1,
              color: '#007AFF'
            })
            .onClick(() => {
              this.chartType = type;
            })
        })
      }
      .width('100%')
    }
    .width('100%')
    .padding(16)
    .backgroundColor(Color.White)
    .borderRadius(8)
  }

  /**
   * 构建图表
   */
  @Builder
  buildChart() {
    if (this.stats) {
      TrajectoryChart({
        chartType: this.chartType,
        dataType: this.selectedChart,
        dailyStats: this.stats.dailyStats,
        hourlyStats: this.stats.hourlyStats,
        monthlyStats: this.stats.monthlyStats,
        selectedPeriod: this.selectedPeriod
      })
    }
  }

  /**
   * 构建热力图
   */
  @Builder
  buildHeatmap() {
    TrajectoryHeatmap({
      timeRange: this.selectedPeriod
    })
  }

  /**
   * 构建详细统计
   */
  @Builder
  buildDetailedStats() {
    Column({ space: 12 }) {
      Text('详细统计')
        .fontSize(16)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')
        .alignSelf(ItemAlign.Start)

      Column({ space: 8 }) {
        Row() {
          Text('平均速度')
            .fontSize(14)
            .fontColor('#666666')
            .layoutWeight(1)
          
          Text(this.formatSpeed(this.stats!.averageSpeed))
            .fontSize(14)
            .fontColor('#333333')
        }
        .width('100%')

        Row() {
          Text('最高速度')
            .fontSize(14)
            .fontColor('#666666')
            .layoutWeight(1)
          
          Text(this.formatSpeed(this.stats!.maxSpeed))
            .fontSize(14)
            .fontColor('#333333')
        }
        .width('100%')

        Row() {
          Text('总轨迹点')
            .fontSize(14)
            .fontColor('#666666')
            .layoutWeight(1)
          
          Text(`${this.stats!.totalPoints} 个`)
            .fontSize(14)
            .fontColor('#333333')
        }
        .width('100%')

        Row() {
          Text('平均精度')
            .fontSize(14)
            .fontColor('#666666')
            .layoutWeight(1)
          
          Text(`${this.stats!.averageAccuracy.toFixed(1)} 米`)
            .fontSize(14)
            .fontColor('#333333')
        }
        .width('100%')

        if (this.stats!.elevationGain > 0) {
          Row() {
            Text('累计爬升')
              .fontSize(14)
              .fontColor('#666666')
              .layoutWeight(1)
            
            Text(`${this.stats!.elevationGain.toFixed(1)} 米`)
              .fontSize(14)
              .fontColor('#333333')
          }
          .width('100%')
        }

        if (this.stats!.elevationLoss > 0) {
          Row() {
            Text('累计下降')
              .fontSize(14)
              .fontColor('#666666')
              .layoutWeight(1)
            
            Text(`${this.stats!.elevationLoss.toFixed(1)} 米`)
              .fontSize(14)
              .fontColor('#333333')
          }
          .width('100%')
        }

        if (this.stats!.mostActiveDay) {
          Row() {
            Text('最活跃日期')
              .fontSize(14)
              .fontColor('#666666')
              .layoutWeight(1)
            
            Text(this.stats!.mostActiveDay)
              .fontSize(14)
              .fontColor('#333333')
          }
          .width('100%')
        }

        Row() {
          Text('最活跃时段')
            .fontSize(14)
            .fontColor('#666666')
            .layoutWeight(1)
          
          Text(`${this.stats!.mostActiveHour}:00 - ${this.stats!.mostActiveHour + 1}:00`)
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

  /**
   * 构建记录排行
   */
  @Builder
  buildRecords() {
    Column({ space: 12 }) {
      Text('记录排行')
        .fontSize(16)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')
        .alignSelf(ItemAlign.Start)

      if (this.stats!.longestTrajectory) {
        Row({ space: 12 }) {
          Image($r('app.media.ic_trophy'))
            .width(24)
            .height(24)
            .fillColor('#FFD700')

          Column({ space: 4 }) {
            Text('最长距离')
              .fontSize(14)
              .fontWeight(FontWeight.Medium)
              .fontColor('#333333')
              .alignSelf(ItemAlign.Start)
            
            Text(this.formatDistance(this.stats!.longestTrajectory.totalDistance))
              .fontSize(12)
              .fontColor('#666666')
              .alignSelf(ItemAlign.Start)
          }
          .layoutWeight(1)
          .alignItems(HorizontalAlign.Start)

          Text('查看')
            .fontSize(12)
            .fontColor('#007AFF')
            .onClick(() => {
              // 跳转到轨迹详情
            })
        }
        .width('100%')
        .padding(12)
        .backgroundColor('#FFF9E6')
        .borderRadius(8)
      }

      if (this.stats!.fastestTrajectory) {
        Row({ space: 12 }) {
          Image($r('app.media.ic_speed'))
            .width(24)
            .height(24)
            .fillColor('#FF6B6B')

          Column({ space: 4 }) {
            Text('最高速度')
              .fontSize(14)
              .fontWeight(FontWeight.Medium)
              .fontColor('#333333')
              .alignSelf(ItemAlign.Start)
            
            Text(this.formatSpeed(this.stats!.fastestTrajectory.maxSpeed || 0))
              .fontSize(12)
              .fontColor('#666666')
              .alignSelf(ItemAlign.Start)
          }
          .layoutWeight(1)
          .alignItems(HorizontalAlign.Start)

          Text('查看')
            .fontSize(12)
            .fontColor('#007AFF')
            .onClick(() => {
              // 跳转到轨迹详情
            })
        }
        .width('100%')
        .padding(12)
        .backgroundColor('#FFE6E6')
        .borderRadius(8)
      }
    }
    .width('100%')
    .padding(16)
    .backgroundColor(Color.White)
    .borderRadius(8)
  }
}