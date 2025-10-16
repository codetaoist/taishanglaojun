import { DailyStats, HourlyStats, MonthlyStats } from './TrajectoryAnalytics';

/**
 * 图表数据点接口
 */
export interface ChartDataPoint {
  label: string;
  value: number;
  color?: string;
}

/**
 * 轨迹图表组件
 */
@Component
export struct TrajectoryChart {
  @Prop chartType: string = 'line'; // line, bar, pie
  @Prop dataType: string = 'distance'; // distance, speed, duration, count
  @Prop dailyStats: DailyStats[] = [];
  @Prop hourlyStats: HourlyStats[] = [];
  @Prop monthlyStats: MonthlyStats[] = [];
  @Prop selectedPeriod: string = '7d';
  
  @State private chartData: ChartDataPoint[] = [];
  @State private maxValue: number = 0;
  @State private minValue: number = 0;

  aboutToAppear() {
    this.updateChartData();
  }

  aboutToUpdate() {
    this.updateChartData();
  }

  /**
   * 更新图表数据
   */
  private updateChartData() {
    let data: ChartDataPoint[] = [];
    
    switch (this.selectedPeriod) {
      case '7d':
      case '30d':
      case '90d':
        data = this.getDailyChartData();
        break;
      case '1y':
        data = this.getMonthlyChartData();
        break;
      case 'all':
        data = this.getMonthlyChartData();
        break;
      default:
        data = this.getDailyChartData();
        break;
    }
    
    this.chartData = data;
    this.calculateMinMax();
  }

  /**
   * 获取每日图表数据
   */
  private getDailyChartData(): ChartDataPoint[] {
    return this.dailyStats.map(stat => {
      let value = 0;
      let color = '#007AFF';
      
      switch (this.dataType) {
        case 'distance':
          value = stat.totalDistance / 1000; // 转换为公里
          color = '#34C759';
          break;
        case 'speed':
          value = stat.averageSpeed * 3.6; // 转换为 km/h
          color = '#FF9500';
          break;
        case 'duration':
          value = stat.totalDuration / (1000 * 60 * 60); // 转换为小时
          color = '#AF52DE';
          break;
        case 'count':
          value = stat.trajectoryCount;
          color = '#007AFF';
          break;
      }
      
      return {
        label: this.formatDateLabel(stat.date),
        value,
        color
      };
    });
  }

  /**
   * 获取每月图表数据
   */
  private getMonthlyChartData(): ChartDataPoint[] {
    return this.monthlyStats.map(stat => {
      let value = 0;
      let color = '#007AFF';
      
      switch (this.dataType) {
        case 'distance':
          value = stat.totalDistance / 1000; // 转换为公里
          color = '#34C759';
          break;
        case 'speed':
          value = stat.averageSpeed * 3.6; // 转换为 km/h
          color = '#FF9500';
          break;
        case 'duration':
          value = stat.totalDuration / (1000 * 60 * 60); // 转换为小时
          color = '#AF52DE';
          break;
        case 'count':
          value = stat.trajectoryCount;
          color = '#007AFF';
          break;
      }
      
      return {
        label: this.formatMonthLabel(stat.month),
        value,
        color
      };
    });
  }

  /**
   * 计算最大最小值
   */
  private calculateMinMax() {
    if (this.chartData.length === 0) {
      this.maxValue = 0;
      this.minValue = 0;
      return;
    }
    
    const values = this.chartData.map(d => d.value);
    this.maxValue = Math.max(...values);
    this.minValue = Math.min(...values);
    
    // 为了更好的显示效果，给最大值增加一些余量
    this.maxValue = this.maxValue * 1.1;
  }

  /**
   * 格式化日期标签
   */
  private formatDateLabel(date: string): string {
    const d = new Date(date);
    return `${d.getMonth() + 1}/${d.getDate()}`;
  }

  /**
   * 格式化月份标签
   */
  private formatMonthLabel(month: string): string {
    const [year, monthNum] = month.split('-');
    return `${year}/${monthNum}`;
  }

  /**
   * 获取数据单位
   */
  private getDataUnit(): string {
    switch (this.dataType) {
      case 'distance':
        return 'km';
      case 'speed':
        return 'km/h';
      case 'duration':
        return 'h';
      case 'count':
        return '次';
      default:
        return '';
    }
  }

  /**
   * 格式化数值显示
   */
  private formatValue(value: number): string {
    if (this.dataType === 'duration') {
      return value.toFixed(1);
    } else if (this.dataType === 'count') {
      return value.toString();
    } else {
      return value.toFixed(2);
    }
  }

  build() {
    Column({ space: 12 }) {
      // 图表标题
      Row() {
        Text(this.getChartTitle())
          .fontSize(16)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
        
        Blank()
        
        Text(this.getDataUnit())
          .fontSize(12)
          .fontColor('#999999')
      }
      .width('100%')

      if (this.chartData.length === 0) {
        // 无数据状态
        Column({ space: 8 }) {
          Text('暂无数据')
            .fontSize(14)
            .fontColor('#999999')
          
          Text('选择其他时间段查看')
            .fontSize(12)
            .fontColor('#CCCCCC')
        }
        .width('100%')
        .height(200)
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
        .backgroundColor('#F8F8F8')
        .borderRadius(8)
        
      } else {
        // 图表内容
        if (this.chartType === 'line') {
          this.buildLineChart();
        } else if (this.chartType === 'bar') {
          this.buildBarChart();
        } else {
          this.buildBarChart(); // 默认使用柱状图
        }
      }
    }
    .width('100%')
    .padding(16)
    .backgroundColor(Color.White)
    .borderRadius(8)
  }

  /**
   * 获取图表标题
   */
  private getChartTitle(): string {
    const typeMap = {
      'distance': '距离',
      'speed': '速度',
      'duration': '时长',
      'count': '次数'
    };
    
    const periodMap = {
      '7d': '7天',
      '30d': '30天',
      '90d': '90天',
      '1y': '1年',
      'all': '全部'
    };
    
    return `${typeMap[this.dataType] || '数据'}趋势 (${periodMap[this.selectedPeriod] || ''})`;
  }

  /**
   * 构建折线图
   */
  @Builder
  buildLineChart() {
    Column({ space: 8 }) {
      // Y轴标签和图表区域
      Row() {
        // Y轴
        Column({ space: 4 }) {
          Text(this.formatValue(this.maxValue))
            .fontSize(10)
            .fontColor('#999999')
          
          Blank()
          
          Text(this.formatValue(this.maxValue / 2))
            .fontSize(10)
            .fontColor('#999999')
          
          Blank()
          
          Text('0')
            .fontSize(10)
            .fontColor('#999999')
        }
        .width(40)
        .height(160)
        .alignItems(HorizontalAlign.End)
        
        // 图表区域
        Stack() {
          // 网格线
          Column() {
            ForEach([0, 1, 2, 3, 4], (index: number) => {
              Divider()
                .color('#F0F0F0')
                .strokeWidth(0.5)
            })
          }
          .width('100%')
          .height('100%')
          .justifyContent(FlexAlign.SpaceBetween)
          
          // 折线图
          this.buildLineChartPath()
        }
        .layoutWeight(1)
        .height(160)
        .margin({ left: 8 })
      }
      .width('100%')
      
      // X轴标签
      Row() {
        Blank().width(48)
        
        Row() {
          ForEach(this.chartData, (data: ChartDataPoint, index: number) => {
            Text(data.label)
              .fontSize(10)
              .fontColor('#999999')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .layoutWeight(1)
      }
      .width('100%')
    }
  }

  /**
   * 构建折线图路径
   */
  @Builder
  buildLineChartPath() {
    Row() {
      ForEach(this.chartData, (data: ChartDataPoint, index: number) => {
        Stack() {
          // 数据点
          Circle({ width: 6, height: 6 })
            .fill(data.color || '#007AFF')
            .position({
              x: '50%',
              y: `${100 - (data.value / this.maxValue) * 100}%`
            })
            .translate({ x: '-50%', y: '-50%' })
          
          // 连接线（简化实现）
          if (index < this.chartData.length - 1) {
            Line()
              .startPoint([0, (1 - data.value / this.maxValue) * 160])
              .endPoint([100, (1 - this.chartData[index + 1].value / this.maxValue) * 160])
              .stroke(data.color || '#007AFF')
              .strokeWidth(2)
              .width('100%')
              .height('100%')
          }
        }
        .layoutWeight(1)
        .height('100%')
      })
    }
    .width('100%')
    .height('100%')
  }

  /**
   * 构建柱状图
   */
  @Builder
  buildBarChart() {
    Column({ space: 8 }) {
      // Y轴标签和图表区域
      Row() {
        // Y轴
        Column({ space: 4 }) {
          Text(this.formatValue(this.maxValue))
            .fontSize(10)
            .fontColor('#999999')
          
          Blank()
          
          Text(this.formatValue(this.maxValue / 2))
            .fontSize(10)
            .fontColor('#999999')
          
          Blank()
          
          Text('0')
            .fontSize(10)
            .fontColor('#999999')
        }
        .width(40)
        .height(160)
        .alignItems(HorizontalAlign.End)
        
        // 图表区域
        Stack() {
          // 网格线
          Column() {
            ForEach([0, 1, 2, 3, 4], (index: number) => {
              Divider()
                .color('#F0F0F0')
                .strokeWidth(0.5)
            })
          }
          .width('100%')
          .height('100%')
          .justifyContent(FlexAlign.SpaceBetween)
          
          // 柱状图
          Row({ space: 2 }) {
            ForEach(this.chartData, (data: ChartDataPoint, index: number) => {
              Column() {
                Blank()
                
                Column()
                  .width('100%')
                  .height(`${(data.value / this.maxValue) * 100}%`)
                  .backgroundColor(data.color || '#007AFF')
                  .borderRadius({ topLeft: 2, topRight: 2 })
              }
              .layoutWeight(1)
              .height('100%')
              .justifyContent(FlexAlign.End)
            })
          }
          .width('100%')
          .height('100%')
          .padding({ left: 4, right: 4 })
        }
        .layoutWeight(1)
        .height(160)
        .margin({ left: 8 })
      }
      .width('100%')
      
      // X轴标签
      Row() {
        Blank().width(48)
        
        Row({ space: 2 }) {
          ForEach(this.chartData, (data: ChartDataPoint, index: number) => {
            Text(data.label)
              .fontSize(10)
              .fontColor('#999999')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .layoutWeight(1)
        .padding({ left: 4, right: 4 })
      }
      .width('100%')
    }
  }
}