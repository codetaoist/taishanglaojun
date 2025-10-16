import router from '@ohos.router';
import { TrajectoryAnalytics } from '../components/TrajectoryAnalytics';
import { TrajectoryChart } from '../components/TrajectoryChart';
import { DataService } from '../service/DataService';
import { Trajectory } from '../model/Trajectory';

/**
 * 轨迹统计页面
 */
@Entry
@Component
struct TrajectoryStatsPage {
  @State private selectedTab: number = 0; // 0: 总览, 1: 图表, 2: 排行, 3: 对比
  @State private isLoading: boolean = false;
  @State private trajectories: Trajectory[] = [];
  @State private compareTrajectories: Trajectory[] = [];
  @State private showCompareDialog: boolean = false;

  private dataService: DataService = DataService.getInstance();

  aboutToAppear() {
    this.loadTrajectories();
  }

  /**
   * 加载轨迹数据
   */
  private async loadTrajectories() {
    try {
      this.isLoading = true;
      this.trajectories = await this.dataService.getTrajectories({
        limit: 100,
        orderBy: 'startTime',
        order: 'desc'
      });
    } catch (error) {
      console.error('Failed to load trajectories:', error);
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * 添加到对比列表
   */
  private addToCompare(trajectory: Trajectory) {
    if (this.compareTrajectories.length < 3 && 
        !this.compareTrajectories.find(t => t.id === trajectory.id)) {
      this.compareTrajectories.push(trajectory);
    }
  }

  /**
   * 从对比列表移除
   */
  private removeFromCompare(trajectoryId: string) {
    this.compareTrajectories = this.compareTrajectories.filter(t => t.id !== trajectoryId);
  }

  /**
   * 格式化距离
   */
  private formatDistance(meters: number): string {
    if (meters >= 1000) {
      return `${(meters / 1000).toFixed(2)} km`;
    } else {
      return `${meters.toFixed(0)} m`;
    }
  }

  /**
   * 格式化持续时间
   */
  private formatDuration(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    } else {
      return `${seconds}s`;
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
   * 格式化日期
   */
  private formatDate(timestamp: number): string {
    const date = new Date(timestamp);
    return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  }

  build() {
    Column() {
      // 顶部导航栏
      Row() {
        Image($r('app.media.ic_back'))
          .width(24)
          .height(24)
          .fillColor('#333333')
          .onClick(() => {
            router.back();
          })

        Text('轨迹统计')
          .fontSize(18)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
          .layoutWeight(1)
          .textAlign(TextAlign.Center)

        Image($r('app.media.ic_more'))
          .width(24)
          .height(24)
          .fillColor('#333333')
          .onClick(() => {
            // 更多选项
          })
      }
      .width('100%')
      .height(56)
      .padding({ left: 16, right: 16 })
      .backgroundColor(Color.White)
      .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })

      // 标签页导航
      Row() {
        ForEach([
          { index: 0, title: '总览', icon: $r('app.media.ic_overview') },
          { index: 1, title: '图表', icon: $r('app.media.ic_chart') },
          { index: 2, title: '排行', icon: $r('app.media.ic_ranking') },
          { index: 3, title: '对比', icon: $r('app.media.ic_compare') }
        ], (tab: any) => {
          Column({ space: 4 }) {
            Image(tab.icon)
              .width(20)
              .height(20)
              .fillColor(this.selectedTab === tab.index ? '#007AFF' : '#999999')

            Text(tab.title)
              .fontSize(12)
              .fontColor(this.selectedTab === tab.index ? '#007AFF' : '#999999')
          }
          .layoutWeight(1)
          .padding({ top: 8, bottom: 8 })
          .onClick(() => {
            this.selectedTab = tab.index;
          })
        })
      }
      .width('100%')
      .backgroundColor(Color.White)
      .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })

      // 内容区域
      if (this.selectedTab === 0) {
        // 总览标签页
        this.buildOverviewTab();
      } else if (this.selectedTab === 1) {
        // 图表标签页
        this.buildChartTab();
      } else if (this.selectedTab === 2) {
        // 排行标签页
        this.buildRankingTab();
      } else if (this.selectedTab === 3) {
        // 对比标签页
        this.buildCompareTab();
      }

      // 对比选择对话框
      if (this.showCompareDialog) {
        this.buildCompareDialog();
      }
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F8F8F8')
  }

  /**
   * 构建总览标签页
   */
  @Builder
  buildOverviewTab() {
    TrajectoryAnalytics()
      .width('100%')
      .layoutWeight(1)
  }

  /**
   * 构建图表标签页
   */
  @Builder
  buildChartTab() {
    Scroll() {
      Column({ space: 16 }) {
        // 快速统计卡片
        Row({ space: 12 }) {
          Column({ space: 4 }) {
            Text(`${this.trajectories.length}`)
              .fontSize(20)
              .fontWeight(FontWeight.Bold)
              .fontColor('#007AFF')
            
            Text('总轨迹')
              .fontSize(12)
              .fontColor('#666666')
          }
          .layoutWeight(1)
          .padding(12)
          .backgroundColor(Color.White)
          .borderRadius(8)
          .alignItems(HorizontalAlign.Center)

          Column({ space: 4 }) {
            Text(this.formatDistance(this.trajectories.reduce((sum, t) => sum + t.totalDistance, 0)))
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
          .padding(12)
          .backgroundColor(Color.White)
          .borderRadius(8)
          .alignItems(HorizontalAlign.Center)

          Column({ space: 4 }) {
            Text(this.formatDuration(this.trajectories.reduce((sum, t) => sum + (t.endTime - t.startTime), 0)))
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
          .padding(12)
          .backgroundColor(Color.White)
          .borderRadius(8)
          .alignItems(HorizontalAlign.Center)
        }
        .width('100%')

        // 详细图表分析
        TrajectoryAnalytics()
      }
      .width('100%')
      .padding(16)
    }
    .width('100%')
    .layoutWeight(1)
  }

  /**
   * 构建排行标签页
   */
  @Builder
  buildRankingTab() {
    Scroll() {
      Column({ space: 16 }) {
        // 距离排行
        this.buildRankingSection('距离排行', 'distance');
        
        // 速度排行
        this.buildRankingSection('速度排行', 'speed');
        
        // 时长排行
        this.buildRankingSection('时长排行', 'duration');
      }
      .width('100%')
      .padding(16)
    }
    .width('100%')
    .layoutWeight(1)
  }

  /**
   * 构建排行榜区块
   */
  @Builder
  buildRankingSection(title: string, type: string) {
    Column({ space: 12 }) {
      Text(title)
        .fontSize(16)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')
        .alignSelf(ItemAlign.Start)

      Column({ space: 8 }) {
        ForEach(this.getSortedTrajectories(type).slice(0, 5), (trajectory: Trajectory, index: number) => {
          Row({ space: 12 }) {
            // 排名
            Text(`${index + 1}`)
              .fontSize(16)
              .fontWeight(FontWeight.Bold)
              .fontColor(index === 0 ? '#FFD700' : index === 1 ? '#C0C0C0' : index === 2 ? '#CD7F32' : '#666666')
              .width(24)
              .textAlign(TextAlign.Center)

            // 轨迹信息
            Column({ space: 4 }) {
              Text(this.formatDate(trajectory.startTime))
                .fontSize(14)
                .fontColor('#333333')
                .alignSelf(ItemAlign.Start)
              
              Text(this.getTrajectoryValue(trajectory, type))
                .fontSize(12)
                .fontColor('#666666')
                .alignSelf(ItemAlign.Start)
            }
            .layoutWeight(1)
            .alignItems(HorizontalAlign.Start)

            // 查看按钮
            Text('查看')
              .fontSize(12)
              .fontColor('#007AFF')
              .onClick(() => {
                router.pushUrl({
                  url: 'pages/TrajectoryDetailPage',
                  params: { trajectoryId: trajectory.id }
                });
              })
          }
          .width('100%')
          .padding(12)
          .backgroundColor(Color.White)
          .borderRadius(8)
        })
      }
    }
    .width('100%')
    .padding(16)
    .backgroundColor(Color.White)
    .borderRadius(8)
  }

  /**
   * 获取排序后的轨迹列表
   */
  private getSortedTrajectories(type: string): Trajectory[] {
    return [...this.trajectories].sort((a, b) => {
      switch (type) {
        case 'distance':
          return b.totalDistance - a.totalDistance;
        case 'speed':
          return (b.maxSpeed || 0) - (a.maxSpeed || 0);
        case 'duration':
          return (b.endTime - b.startTime) - (a.endTime - a.startTime);
        default:
          return 0;
      }
    });
  }

  /**
   * 获取轨迹指定类型的值
   */
  private getTrajectoryValue(trajectory: Trajectory, type: string): string {
    switch (type) {
      case 'distance':
        return this.formatDistance(trajectory.totalDistance);
      case 'speed':
        return this.formatSpeed(trajectory.maxSpeed || 0);
      case 'duration':
        return this.formatDuration(trajectory.endTime - trajectory.startTime);
      default:
        return '';
    }
  }

  /**
   * 构建对比标签页
   */
  @Builder
  buildCompareTab() {
    Column({ space: 16 }) {
      // 对比列表
      if (this.compareTrajectories.length > 0) {
        Column({ space: 12 }) {
          Text('对比轨迹')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .alignSelf(ItemAlign.Start)

          ForEach(this.compareTrajectories, (trajectory: Trajectory, index: number) => {
            Row({ space: 12 }) {
              Text(`${index + 1}`)
                .fontSize(14)
                .fontWeight(FontWeight.Bold)
                .fontColor('#007AFF')
                .width(24)
                .textAlign(TextAlign.Center)

              Column({ space: 4 }) {
                Text(this.formatDate(trajectory.startTime))
                  .fontSize(14)
                  .fontColor('#333333')
                  .alignSelf(ItemAlign.Start)
                
                Row({ space: 16 }) {
                  Text(this.formatDistance(trajectory.totalDistance))
                    .fontSize(12)
                    .fontColor('#666666')
                  
                  Text(this.formatDuration(trajectory.endTime - trajectory.startTime))
                    .fontSize(12)
                    .fontColor('#666666')
                  
                  Text(this.formatSpeed(trajectory.averageSpeed || 0))
                    .fontSize(12)
                    .fontColor('#666666')
                }
              }
              .layoutWeight(1)
              .alignItems(HorizontalAlign.Start)

              Image($r('app.media.ic_delete'))
                .width(20)
                .height(20)
                .fillColor('#FF3B30')
                .onClick(() => {
                  this.removeFromCompare(trajectory.id);
                })
            }
            .width('100%')
            .padding(12)
            .backgroundColor(Color.White)
            .borderRadius(8)
          })

          // 对比图表
          if (this.compareTrajectories.length >= 2) {
            this.buildCompareChart();
          }
        }
        .width('100%')
        .padding(16)

      } else {
        // 空状态
        Column({ space: 16 }) {
          Image($r('app.media.ic_compare_empty'))
            .width(80)
            .height(80)
            .fillColor('#CCCCCC')

          Text('选择轨迹进行对比')
            .fontSize(16)
            .fontColor('#666666')

          Text('最多可选择3条轨迹')
            .fontSize(14)
            .fontColor('#999999')
        }
        .width('100%')
        .layoutWeight(1)
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
      }

      // 添加轨迹按钮
      Button('选择轨迹')
        .width('100%')
        .height(44)
        .backgroundColor('#007AFF')
        .fontColor(Color.White)
        .margin(16)
        .onClick(() => {
          this.showCompareDialog = true;
        })
    }
    .width('100%')
    .layoutWeight(1)
  }

  /**
   * 构建对比图表
   */
  @Builder
  buildCompareChart() {
    Column({ space: 12 }) {
      Text('对比分析')
        .fontSize(16)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')
        .alignSelf(ItemAlign.Start)

      // 对比表格
      Column({ space: 8 }) {
        // 表头
        Row() {
          Text('指标')
            .fontSize(12)
            .fontColor('#666666')
            .width(60)
          
          ForEach(this.compareTrajectories, (trajectory: Trajectory, index: number) => {
            Text(`轨迹${index + 1}`)
              .fontSize(12)
              .fontColor('#666666')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .width('100%')
        .padding(8)
        .backgroundColor('#F8F8F8')

        // 距离对比
        Row() {
          Text('距离')
            .fontSize(12)
            .fontColor('#333333')
            .width(60)
          
          ForEach(this.compareTrajectories, (trajectory: Trajectory) => {
            Text(this.formatDistance(trajectory.totalDistance))
              .fontSize(12)
              .fontColor('#333333')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .width('100%')
        .padding(8)

        // 时长对比
        Row() {
          Text('时长')
            .fontSize(12)
            .fontColor('#333333')
            .width(60)
          
          ForEach(this.compareTrajectories, (trajectory: Trajectory) => {
            Text(this.formatDuration(trajectory.endTime - trajectory.startTime))
              .fontSize(12)
              .fontColor('#333333')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .width('100%')
        .padding(8)

        // 平均速度对比
        Row() {
          Text('平均速度')
            .fontSize(12)
            .fontColor('#333333')
            .width(60)
          
          ForEach(this.compareTrajectories, (trajectory: Trajectory) => {
            Text(this.formatSpeed(trajectory.averageSpeed || 0))
              .fontSize(12)
              .fontColor('#333333')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .width('100%')
        .padding(8)

        // 最高速度对比
        Row() {
          Text('最高速度')
            .fontSize(12)
            .fontColor('#333333')
            .width(60)
          
          ForEach(this.compareTrajectories, (trajectory: Trajectory) => {
            Text(this.formatSpeed(trajectory.maxSpeed || 0))
              .fontSize(12)
              .fontColor('#333333')
              .layoutWeight(1)
              .textAlign(TextAlign.Center)
          })
        }
        .width('100%')
        .padding(8)
      }
      .width('100%')
      .backgroundColor(Color.White)
      .borderRadius(8)
    }
    .width('100%')
    .padding(16)
    .backgroundColor(Color.White)
    .borderRadius(8)
  }

  /**
   * 构建对比选择对话框
   */
  @Builder
  buildCompareDialog() {
    Stack() {
      // 背景遮罩
      Column()
        .width('100%')
        .height('100%')
        .backgroundColor('rgba(0, 0, 0, 0.5)')
        .onClick(() => {
          this.showCompareDialog = false;
        })

      // 对话框内容
      Column({ space: 16 }) {
        Text('选择对比轨迹')
          .fontSize(18)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')

        Scroll() {
          Column({ space: 8 }) {
            ForEach(this.trajectories.slice(0, 20), (trajectory: Trajectory) => {
              Row({ space: 12 }) {
                Checkbox()
                  .select(this.compareTrajectories.find(t => t.id === trajectory.id) !== undefined)
                  .onChange((value: boolean) => {
                    if (value) {
                      this.addToCompare(trajectory);
                    } else {
                      this.removeFromCompare(trajectory.id);
                    }
                  })

                Column({ space: 4 }) {
                  Text(this.formatDate(trajectory.startTime))
                    .fontSize(14)
                    .fontColor('#333333')
                    .alignSelf(ItemAlign.Start)
                  
                  Row({ space: 12 }) {
                    Text(this.formatDistance(trajectory.totalDistance))
                      .fontSize(12)
                      .fontColor('#666666')
                    
                    Text(this.formatDuration(trajectory.endTime - trajectory.startTime))
                      .fontSize(12)
                      .fontColor('#666666')
                  }
                }
                .layoutWeight(1)
                .alignItems(HorizontalAlign.Start)
              }
              .width('100%')
              .padding(12)
              .backgroundColor('#F8F8F8')
              .borderRadius(8)
            })
          }
        }
        .width('100%')
        .height(300)

        Row({ space: 12 }) {
          Button('取消')
            .layoutWeight(1)
            .height(44)
            .backgroundColor('#F8F8F8')
            .fontColor('#333333')
            .onClick(() => {
              this.showCompareDialog = false;
            })

          Button('确定')
            .layoutWeight(1)
            .height(44)
            .backgroundColor('#007AFF')
            .fontColor(Color.White)
            .onClick(() => {
              this.showCompareDialog = false;
            })
        }
        .width('100%')
      }
      .width('90%')
      .padding(20)
      .backgroundColor(Color.White)
      .borderRadius(12)
    }
    .width('100%')
    .height('100%')
    .position({ x: 0, y: 0 })
  }
}