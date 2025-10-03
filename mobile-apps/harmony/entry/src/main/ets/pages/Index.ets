import { LocationViewModel } from '../viewmodel/LocationViewModel';
import { LocationServiceStatus } from '../service/LocationService';
import { TrajectoryStatus } from '../model/Trajectory';
import { LocationPoint } from '../data/LocationPoint';
import { Trajectory } from '../data/Trajectory';
import { MapPage } from './MapPage';
import { TimelineView } from './TimelineView';
import hilog from '@ohos.hilog';

@Entry
@Component
struct Index {
  @State private locationViewModel: LocationViewModel = new LocationViewModel();
  @State private isInitialized: boolean = false;
  @State private selectedTab: number = 0;
  @State private selectedTrajectory: Trajectory | null = null;

  aboutToAppear() {
    this.initializeViewModel();
  }

  aboutToDisappear() {
    this.locationViewModel.cleanup();
  }

  /**
   * 初始化视图模型
   */
  private async initializeViewModel(): Promise<void> {
    try {
      const context = getContext(this) as Context;
      await this.locationViewModel.initialize(context);
      
      // 添加状态变化监听器
      this.locationViewModel.addStateChangeListener(() => {
        // 触发UI更新
        this.forceUpdate();
      });
      
      this.isInitialized = true;
    } catch (error) {
      hilog.error(0x0000, 'Index', 'Initialize error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 强制更新UI
   */
  private forceUpdate(): void {
    // 通过修改状态来触发UI更新
    this.selectedTab = this.selectedTab;
  }

  build() {
    Column() {
      if (!this.isInitialized) {
        this.buildLoadingView();
      } else {
        this.buildMainContent();
      }
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F5F5F5')
  }

  @Builder
  buildLoadingView() {
    Column() {
      LoadingProgress()
        .width(50)
        .height(50)
        .color('#007AFF')
      
      Text('初始化中...')
        .fontSize(16)
        .fontColor('#666666')
        .margin({ top: 16 })
    }
    .width('100%')
    .height('100%')
    .justifyContent(FlexAlign.Center)
    .alignItems(HorizontalAlign.Center)
  }

  @Builder
  buildMainContent() {
    Column() {
      // 顶部导航栏
      this.buildTopBar();
      
      // 标签页内容
      Tabs({ barPosition: BarPosition.End }) {
        TabContent() {
          this.buildTrackingTab();
        }
        .tabBar(this.buildTabBar('跟踪', $r('sys.symbol.location'), 0))

        TabContent() {
          this.buildHistoryTab();
        }
        .tabBar(this.buildTabBar('历史', $r('sys.symbol.clock'), 1))

        TabContent() {
          this.buildStatisticsTab();
        }
        .tabBar(this.buildTabBar('统计', $r('sys.symbol.chart_bar'), 2))

        TabContent() {
          this.buildSettingsTab();
        }
        .tabBar(this.buildTabBar('设置', $r('sys.symbol.gear'), 3))

        TabContent() {
          this.buildMapTab();
        }
        .tabBar(this.buildTabBar('地图', $r('sys.symbol.map'), 4))

        TabContent() {
          this.buildTimelineTab();
        }
        .tabBar(this.buildTabBar('时间轴', $r('sys.symbol.timeline'), 5))
      }
      .onChange((index: number) => {
          this.selectedTab = index;
        })
      .layoutWeight(1)
    }
    .width('100%')
    .height('100%')
  }

  @Builder
  buildTopBar() {
    Row() {
      Text('太上老君轨迹跟踪')
        .fontSize(20)
        .fontWeight(FontWeight.Bold)
        .fontColor('#333333')
      
      Blank()
      
      // 位置服务状态指示器
      this.buildStatusIndicator();
    }
    .width('100%')
    .height(56)
    .padding({ left: 16, right: 16 })
    .backgroundColor('#FFFFFF')
    .shadow({ radius: 2, color: '#00000010', offsetY: 1 })
  }

  @Builder
  buildStatusIndicator() {
    Row() {
      Circle()
        .width(8)
        .height(8)
        .fill(this.getStatusColor())
      
      Text(this.getStatusText())
        .fontSize(12)
        .fontColor('#666666')
        .margin({ left: 4 })
    }
  }

  @Builder
  buildTabBar(title: string, icon: Resource, index: number) {
    Column() {
      Image(icon)
          .width(20)
          .height(20)
          .fillColor(this.selectedTab === index ? '#007AFF' : '#999999')
        
        Text(title)
          .fontSize(10)
          .fontColor(this.selectedTab === index ? '#007AFF' : '#999999')
        .margin({ top: 4 })
    }
  }

  @Builder
  buildTrackingTab() {
    Column() {
      // 错误消息显示
      if (this.locationViewModel.errorMessage) {
        this.buildErrorMessage();
      }
      
      // 当前位置信息
      this.buildLocationInfo();
      
      // 当前轨迹信息
      if (this.locationViewModel.currentTrajectory) {
        this.buildCurrentTrajectoryInfo();
      }
      
      // 控制按钮
      this.buildControlButtons();
      
      Blank()
    }
    .width('100%')
    .height('100%')
    .padding(16)
  }

  @Builder
  buildErrorMessage() {
    Row() {
      Text(this.locationViewModel.errorMessage)
        .fontSize(14)
        .fontColor('#FF3B30')
        .layoutWeight(1)
      
      Button('关闭')
        .fontSize(12)
        .backgroundColor('#FF3B30')
        .borderRadius(4)
        .padding({ left: 8, right: 8, top: 4, bottom: 4 })
        .onClick(() => {
          this.locationViewModel.clearErrorMessage();
        })
    }
    .width('100%')
    .padding(12)
    .backgroundColor('#FFEBEE')
    .borderRadius(8)
    .margin({ bottom: 16 })
  }

  @Builder
  buildLocationInfo() {
    Column() {
      Row() {
        Text('当前位置')
          .fontSize(16)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
        
        Blank()
        
        Button('刷新')
          .fontSize(12)
          .backgroundColor('#007AFF')
          .borderRadius(4)
          .padding({ left: 8, right: 8, top: 4, bottom: 4 })
          .onClick(() => {
            this.locationViewModel.getCurrentLocation();
          })
      }
      .width('100%')
      .margin({ bottom: 12 })
      
      if (this.locationViewModel.currentLocation) {
        Column() {
          Text(`纬度: ${this.locationViewModel.currentLocation.latitude.toFixed(6)}`)
            .fontSize(14)
            .fontColor('#666666')
            .width('100%')
          
          Text(`经度: ${this.locationViewModel.currentLocation.longitude.toFixed(6)}`)
            .fontSize(14)
            .fontColor('#666666')
            .width('100%')
            .margin({ top: 4 })
          
          if (this.locationViewModel.currentLocation.accuracy) {
            Text(`精度: ${this.locationViewModel.currentLocation.accuracy.toFixed(1)}m`)
              .fontSize(14)
              .fontColor('#666666')
              .width('100%')
              .margin({ top: 4 })
          }
          
          Text(`时间: ${this.locationViewModel.currentLocation.getFormattedTime()}`)
            .fontSize(12)
            .fontColor('#999999')
            .width('100%')
            .margin({ top: 4 })
        }
        .alignItems(HorizontalAlign.Start)
      } else {
        Text('暂无位置信息')
          .fontSize(14)
          .fontColor('#999999')
          .width('100%')
      }
    }
    .width('100%')
    .padding(16)
    .backgroundColor('#FFFFFF')
    .borderRadius(8)
    .margin({ bottom: 16 })
  }

  @Builder
  buildCurrentTrajectoryInfo() {
    Column() {
      Text('当前轨迹')
        .fontSize(16)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')
        .width('100%')
        .margin({ bottom: 12 })
      
      Column() {
        Text(this.locationViewModel.currentTrajectory!.name)
          .fontSize(14)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
          .width('100%')
        
        Text(`状态: ${this.locationViewModel.currentTrajectory!.getStatusText()}`)
          .fontSize(14)
          .fontColor('#666666')
          .width('100%')
          .margin({ top: 4 })
        
        Text(`距离: ${this.locationViewModel.currentTrajectory!.getFormattedDistance()}`)
          .fontSize(14)
          .fontColor('#666666')
          .width('100%')
          .margin({ top: 4 })
        
        Text(`时长: ${this.locationViewModel.currentTrajectory!.getFormattedDuration()}`)
          .fontSize(14)
          .fontColor('#666666')
          .width('100%')
          .margin({ top: 4 })
        
        Text(`点数: ${this.locationViewModel.currentTrajectory!.pointCount}`)
          .fontSize(14)
          .fontColor('#666666')
          .width('100%')
          .margin({ top: 4 })
      }
      .alignItems(HorizontalAlign.Start)
    }
    .width('100%')
    .padding(16)
    .backgroundColor('#FFFFFF')
    .borderRadius(8)
    .margin({ bottom: 16 })
  }

  @Builder
  buildControlButtons() {
    Column() {
      if (!this.locationViewModel.hasLocationPermission) {
        Button('请求位置权限')
          .width('100%')
          .height(48)
          .fontSize(16)
          .backgroundColor('#007AFF')
          .borderRadius(8)
          .onClick(() => {
            this.locationViewModel.requestLocationPermission();
          })
      } else if (!this.locationViewModel.isTracking) {
        Button('开始跟踪')
          .width('100%')
          .height(48)
          .fontSize(16)
          .backgroundColor('#34C759')
          .borderRadius(8)
          .onClick(() => {
            this.locationViewModel.startTracking();
          })
      } else {
        Row() {
          if (this.locationViewModel.currentTrajectory?.status === TrajectoryStatus.RECORDING) {
            Button('暂停')
              .layoutWeight(1)
              .height(48)
              .fontSize(16)
              .backgroundColor('#FF9500')
              .borderRadius(8)
              .margin({ right: 8 })
              .onClick(() => {
                this.locationViewModel.pauseTracking();
              })
          } else {
            Button('继续')
              .layoutWeight(1)
              .height(48)
              .fontSize(16)
              .backgroundColor('#34C759')
              .borderRadius(8)
              .margin({ right: 8 })
              .onClick(() => {
                this.locationViewModel.resumeTracking();
              })
          }
          
          Button('停止')
            .layoutWeight(1)
            .height(48)
            .fontSize(16)
            .backgroundColor('#FF3B30')
            .borderRadius(8)
            .onClick(() => {
              this.locationViewModel.stopTracking();
            })
        }
        .width('100%')
      }
    }
  }

  @Builder
  buildHistoryTab() {
    Column() {
      Text('轨迹历史')
        .fontSize(20)
        .fontWeight(FontWeight.Bold)
        .fontColor('#333333')
        .width('100%')
        .margin({ bottom: 16 })
      
      if (this.locationViewModel.trajectories.length === 0) {
        Column() {
          Text('暂无轨迹记录')
            .fontSize(16)
            .fontColor('#999999')
        }
        .width('100%')
        .height(200)
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
      } else {
        List() {
          ForEach(this.locationViewModel.trajectories, (trajectory) => {
            ListItem() {
              this.buildTrajectoryItem(trajectory);
            }
          })
        }
        .layoutWeight(1)
      }
    }
    .width('100%')
    .height('100%')
    .padding(16)
  }

  @Builder
  buildTrajectoryItem(trajectory) {
    Row() {
      Column() {
        Text(trajectory.name)
          .fontSize(16)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
          .width('100%')
        
        Text(`${trajectory.getFormattedDistance()} | ${trajectory.getFormattedDuration()}`)
          .fontSize(14)
          .fontColor('#666666')
          .width('100%')
          .margin({ top: 4 })
        
        Text(trajectory.getFormattedStartTime())
          .fontSize(12)
          .fontColor('#999999')
          .width('100%')
          .margin({ top: 4 })
      }
      .layoutWeight(1)
      .alignItems(HorizontalAlign.Start)
      
      Row() {
        Button('查看时间轴')
          .fontSize(12)
          .backgroundColor('#4CAF50')
          .borderRadius(4)
          .padding({ left: 8, right: 8, top: 4, bottom: 4 })
          .margin({ right: 8 })
          .onClick(() => {
            this.selectTrajectoryForTimeline(trajectory);
          })
        
        Button('删除')
          .fontSize(12)
          .backgroundColor('#FF3B30')
          .borderRadius(4)
          .padding({ left: 8, right: 8, top: 4, bottom: 4 })
          .onClick(() => {
            this.locationViewModel.deleteTrajectory(trajectory.id);
          })
      }
    }
    .width('100%')
    .padding(16)
    .backgroundColor('#FFFFFF')
    .borderRadius(8)
    .margin({ bottom: 8 })
  }

  /**
    * 选择轨迹并显示时间轴
    */
   private selectTrajectoryForTimeline(trajectory: Trajectory): void {
     this.selectedTrajectory = trajectory;
     this.selectedTab = 5; // 切换到时间轴标签页
   }

  @Builder
   buildMapTab() {
     Column() {
       Text('地图视图')
         .fontSize(20)
         .fontWeight(FontWeight.Bold)
         .fontColor('#333333')
         .width('100%')
         .margin({ bottom: 16 })
       
       // 地图页面组件
       MapPage()
         .layoutWeight(1)
     }
     .width('100%')
     .height('100%')
     .padding(16)
   }

  @Builder
   buildTimelineTab() {
     Column() {
       Text('时间轴视图')
         .fontSize(20)
         .fontWeight(FontWeight.Bold)
         .fontColor('#333333')
         .width('100%')
         .margin({ bottom: 16 })
       
       if (this.selectedTrajectory) {
         TimelineView({ trajectory: this.selectedTrajectory })
           .layoutWeight(1)
       } else {
         Column() {
           Text('请先从历史记录中选择一个轨迹')
             .fontSize(16)
             .fontColor('#999999')
             .margin({ bottom: 16 })
           
           Button('返回历史记录')
             .type(ButtonType.Normal)
             .fontSize(14)
             .backgroundColor('#2196F3')
             .fontColor(Color.White)
             .borderRadius(8)
             .padding({ left: 16, right: 16, top: 8, bottom: 8 })
             .onClick(() => {
               this.selectedTab = 1; // 切换到历史标签页
             })
         }
         .width('100%')
         .height('100%')
         .justifyContent(FlexAlign.Center)
         .alignItems(HorizontalAlign.Center)
       }
     }
     .width('100%')
     .height('100%')
     .padding(16)
   }

  @Builder
  buildStatisticsTab() {
    Column() {
      Text('统计信息')
        .fontSize(20)
        .fontWeight(FontWeight.Bold)
        .fontColor('#333333')
        .width('100%')
        .margin({ bottom: 16 })
      
      // 这里可以添加统计信息的显示
      Text('功能开发中...')
        .fontSize(16)
        .fontColor('#999999')
    }
    .width('100%')
    .height('100%')
    .padding(16)
  }

  @Builder
  buildSettingsTab() {
    Column() {
      Text('设置')
        .fontSize(20)
        .fontWeight(FontWeight.Bold)
        .fontColor('#333333')
        .width('100%')
        .margin({ bottom: 16 })
      
      Button('清除所有数据')
        .width('100%')
        .height(48)
        .fontSize(16)
        .backgroundColor('#FF3B30')
        .borderRadius(8)
        .onClick(() => {
          this.locationViewModel.clearAllData();
        })
    }
    .width('100%')
    .height('100%')
    .padding(16)
  }

  /**
   * 获取状态颜色
   */
  private getStatusColor(): string {
    switch (this.locationViewModel.locationServiceStatus) {
      case LocationServiceStatus.RUNNING:
        return '#34C759';
      case LocationServiceStatus.PAUSED:
        return '#FF9500';
      case LocationServiceStatus.ERROR:
        return '#FF3B30';
      default:
        return '#999999';
    }
  }

  /**
   * 获取状态文本
   */
  private getStatusText(): string {
    switch (this.locationViewModel.locationServiceStatus) {
      case LocationServiceStatus.RUNNING:
        return '运行中';
      case LocationServiceStatus.PAUSED:
        return '已暂停';
      case LocationServiceStatus.STARTING:
        return '启动中';
      case LocationServiceStatus.ERROR:
        return '错误';
      default:
        return '已停止';
    }
  }
}