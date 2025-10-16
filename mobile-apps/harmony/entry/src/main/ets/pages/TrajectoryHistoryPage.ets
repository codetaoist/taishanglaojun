import { Trajectory } from '../model/Trajectory'
import { DataService } from '../service/DataService'
import { LocationService } from '../service/LocationService'
import { TrajectoryManagementService, ExportFormat } from '../service/TrajectoryManagementService'
import router from '@ohos.router'

// 搜索过滤选项
interface SearchFilters {
  keyword: string
  startDate: Date | null
  endDate: Date | null
  minDistance: number
  maxDistance: number
  minDuration: number
  maxDuration: number
  activityType: string | null // 活动类型：步行、跑步、骑行、驾车等
  trajectoryType: string | null // 轨迹类型：运动、通勤、旅行等
  sortBy: 'date' | 'distance' | 'duration' | 'speed'
  sortOrder: 'asc' | 'desc'
}

// 分页信息
interface PaginationInfo {
  currentPage: number
  pageSize: number
  totalCount: number
  hasMore: boolean
}

@Entry
@Component
struct TrajectoryHistoryPage {
  @State trajectories: Trajectory[] = []
  @State filteredTrajectories: Trajectory[] = []
  @State isLoading: boolean = false
  @State isRefreshing: boolean = false
  @State showSearchDialog: boolean = false
  @State selectedTrajectories: Set<string> = new Set()
  @State isSelectionMode: boolean = false
  @State showExportDialog: boolean = false
  @State showShareDialog: boolean = false
  @State exportFormat: ExportFormat = ExportFormat.GPX
  @State isExporting: boolean = false

  // 搜索和过滤状态
  @State searchFilters: SearchFilters = {
    keyword: '',
    startDate: null,
    endDate: null,
    minDistance: 0,
    maxDistance: 100000, // 100km
    minDuration: 0,
    maxDuration: 86400000, // 24小时（毫秒）
    activityType: null,
    trajectoryType: null,
    sortBy: 'date',
    sortOrder: 'desc'
  }

  // 分页状态
  @State pagination: PaginationInfo = {
    currentPage: 1,
    pageSize: 20,
    totalCount: 0,
    hasMore: true
  }

  private dataService: DataService = new DataService()
  private locationService: LocationService = new LocationService()
  private managementService: TrajectoryManagementService = TrajectoryManagementService.getInstance()

  async aboutToAppear() {
    await this.loadTrajectories()
  }

  // 加载轨迹数据
  async loadTrajectories(refresh: boolean = false) {
    if (this.isLoading) return

    this.isLoading = true
    if (refresh) {
      this.isRefreshing = true
      this.pagination.currentPage = 1
      this.trajectories = []
    }

    try {
      const result = await this.dataService.getTrajectories({
        limit: this.pagination.pageSize,
        offset: (this.pagination.currentPage - 1) * this.pagination.pageSize,
        startDate: this.searchFilters.startDate,
        endDate: this.searchFilters.endDate
      })

      if (refresh) {
        this.trajectories = result.trajectories
      } else {
        this.trajectories = [...this.trajectories, ...result.trajectories]
      }

      this.pagination.totalCount = result.total
      this.pagination.hasMore = this.trajectories.length < result.total

      this.applyFilters()
    } catch (error) {
      console.error('Failed to load trajectories:', error)
    } finally {
      this.isLoading = false
      this.isRefreshing = false
    }
  }

  // 应用搜索和过滤
  applyFilters() {
    let filtered = [...this.trajectories]

    // 关键词搜索
    if (this.searchFilters.keyword) {
      const keyword = this.searchFilters.keyword.toLowerCase()
      filtered = filtered.filter(trajectory => 
        trajectory.name?.toLowerCase().includes(keyword) ||
        trajectory.description?.toLowerCase().includes(keyword)
      )
    }

    // 距离过滤
    filtered = filtered.filter(trajectory => 
      trajectory.totalDistance >= this.searchFilters.minDistance &&
      trajectory.totalDistance <= this.searchFilters.maxDistance
    )

    // 时长过滤
    filtered = filtered.filter(trajectory => 
      trajectory.duration >= this.searchFilters.minDuration &&
      trajectory.duration <= this.searchFilters.maxDuration
    )

    // 活动类型过滤
    if (this.searchFilters.activityType) {
      filtered = filtered.filter(trajectory => 
        trajectory.activityType === this.searchFilters.activityType
      )
    }

    // 轨迹类型过滤
    if (this.searchFilters.trajectoryType) {
      filtered = filtered.filter(trajectory => 
        trajectory.trajectoryType === this.searchFilters.trajectoryType
      )
    }

    // 排序
    filtered.sort((a, b) => {
      let comparison = 0
      switch (this.searchFilters.sortBy) {
        case 'date':
          comparison = new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
          break
        case 'distance':
          comparison = a.totalDistance - b.totalDistance
          break
        case 'duration':
          comparison = a.duration - b.duration
          break
        case 'speed':
          const speedA = a.totalDistance / (a.duration / 1000) // m/s
          const speedB = b.totalDistance / (b.duration / 1000) // m/s
          comparison = speedA - speedB
          break
      }
      return this.searchFilters.sortOrder === 'desc' ? -comparison : comparison
    })

    this.filteredTrajectories = filtered
  }

  // 加载更多数据
  async loadMore() {
    if (!this.pagination.hasMore || this.isLoading) return

    this.pagination.currentPage++
    await this.loadTrajectories()
  }

  // 刷新数据
  async onRefresh() {
    await this.loadTrajectories(true)
  }

  // 搜索轨迹
  onSearch() {
    this.showSearchDialog = true
  }

  // 应用搜索过滤器
  async applySearchFilters() {
    this.showSearchDialog = false
    await this.loadTrajectories(true)
  }

  // 重置搜索过滤器
  resetFilters() {
    this.searchFilters = {
      keyword: '',
      startDate: null,
      endDate: null,
      minDistance: 0,
      maxDistance: 100000,
      minDuration: 0,
      maxDuration: 86400000,
      activityType: null,
      trajectoryType: null,
      sortBy: 'date',
      sortOrder: 'desc'
    }
    this.applyFilters()
  }

  // 切换选择模式
  toggleSelectionMode() {
    this.isSelectionMode = !this.isSelectionMode
    if (!this.isSelectionMode) {
      this.selectedTrajectories.clear()
    }
  }

  // 选择/取消选择轨迹
  toggleTrajectorySelection(trajectoryId: string) {
    if (this.selectedTrajectories.has(trajectoryId)) {
      this.selectedTrajectories.delete(trajectoryId)
    } else {
      this.selectedTrajectories.add(trajectoryId)
    }
  }

  // 全选/取消全选
  toggleSelectAll() {
    if (this.selectedTrajectories.size === this.filteredTrajectories.length) {
      this.selectedTrajectories.clear()
    } else {
      this.selectedTrajectories.clear()
      this.filteredTrajectories.forEach(trajectory => {
        this.selectedTrajectories.add(trajectory.id)
      })
    }
  }

  // 删除选中的轨迹
  async deleteSelectedTrajectories() {
    if (this.selectedTrajectories.size === 0) return

    try {
      const trajectoryIds = Array.from(this.selectedTrajectories)
      const result = await this.managementService.deleteTrajectories(trajectoryIds)
      
      if (result.success > 0) {
        console.log(`Successfully deleted ${result.success} trajectories`)
      }
      if (result.failed > 0) {
        console.error(`Failed to delete ${result.failed} trajectories`)
      }
      
      this.selectedTrajectories.clear()
      this.isSelectionMode = false
      await this.loadTrajectories(true)
    } catch (error) {
      console.error('Failed to delete trajectories:', error)
    }
  }

  // 导出选中的轨迹
  async exportSelectedTrajectories() {
    if (this.selectedTrajectories.size === 0) return

    this.showExportDialog = true
  }

  // 执行导出
  async performExport() {
    if (this.selectedTrajectories.size === 0) return

    this.isExporting = true
    this.showExportDialog = false

    try {
      const trajectoryIds = Array.from(this.selectedTrajectories)
      const trajectories = this.filteredTrajectories.filter(t => trajectoryIds.includes(t.id))
      
      if (trajectories.length === 1) {
        // 单个轨迹导出
        const result = await this.managementService.exportTrajectory(trajectories[0], this.exportFormat)
        if (result.success) {
          console.log(`Exported trajectory to ${result.filePath}`)
        } else {
          console.error(`Export failed: ${result.error}`)
        }
      } else {
        // 批量导出
        const result = await this.managementService.exportTrajectories(trajectories, this.exportFormat)
        if (result.success) {
          console.log(`Exported ${trajectories.length} trajectories to ${result.filePath}`)
        } else {
          console.error(`Export failed: ${result.error}`)
        }
      }
    } catch (error) {
      console.error('Failed to export trajectories:', error)
    } finally {
      this.isExporting = false
    }
  }

  // 分享选中的轨迹
  async shareSelectedTrajectories() {
    if (this.selectedTrajectories.size === 0) return

    this.showShareDialog = true
  }

  // 执行分享
  async performShare() {
    if (this.selectedTrajectories.size === 0) return

    this.showShareDialog = false

    try {
      const trajectoryIds = Array.from(this.selectedTrajectories)
      const trajectories = this.filteredTrajectories.filter(t => trajectoryIds.includes(t.id))
      
      for (const trajectory of trajectories) {
        await this.managementService.shareTrajectory(trajectory, {
          title: `分享轨迹: ${trajectory.name}`,
          text: `我刚刚完成了一段${this.formatDistance(trajectory.totalDistance)}的行程`,
          format: this.exportFormat
        })
      }
    } catch (error) {
      console.error('Failed to share trajectories:', error)
    }
  }

  // 导航到轨迹详情
  navigateToDetail(trajectory: Trajectory) {
    if (this.isSelectionMode) {
      this.toggleTrajectorySelection(trajectory.id)
    } else {
      router.pushUrl({
        url: 'pages/TrajectoryDetailPage',
        params: { trajectoryId: trajectory.id }
      })
    }
  }

  // 格式化时间
  formatTime(timestamp: number): string {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) {
      return `今天 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
    } else if (diffDays === 1) {
      return `昨天 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
    } else if (diffDays < 7) {
      return `${diffDays}天前`
    } else {
      return date.toLocaleDateString('zh-CN')
    }
  }

  // 格式化距离
  formatDistance(distance: number): string {
    if (distance < 1000) {
      return `${distance.toFixed(0)}m`
    } else {
      return `${(distance / 1000).toFixed(1)}km`
    }
  }

  // 格式化持续时间
  formatDuration(duration: number): string {
    const hours = Math.floor(duration / 3600)
    const minutes = Math.floor((duration % 3600) / 60)
    
    if (hours > 0) {
      return `${hours}小时${minutes}分钟`
    } else {
      return `${minutes}分钟`
    }
  }

  // 获取活动类型文本
  getActivityTypeText(activityType: string): string {
    switch (activityType) {
      case 'walking': return '步行';
      case 'running': return '跑步';
      case 'cycling': return '骑行';
      case 'driving': return '驾车';
      default: return activityType;
    }
  }

  // 获取轨迹类型文本
  getTrajectoryTypeText(trajectoryType: string): string {
    switch (trajectoryType) {
      case 'exercise': return '运动';
      case 'commute': return '通勤';
      case 'travel': return '旅行';
      case 'leisure': return '休闲';
      default: return trajectoryType;
    }
  }

  build() {
    Column() {
      // 顶部导航栏
      Row() {
        Button() {
          Image($r('app.media.ic_back'))
            .width(24)
            .height(24)
            .fillColor(Color.Black)
        }
        .backgroundColor(Color.Transparent)
        .onClick(() => router.back())

        Text('轨迹历史')
          .fontSize(18)
          .fontWeight(FontWeight.Medium)
          .layoutWeight(1)
          .textAlign(TextAlign.Center)

        Row() {
          Button() {
            Image($r('app.media.ic_search'))
              .width(24)
              .height(24)
              .fillColor(Color.Black)
          }
          .backgroundColor(Color.Transparent)
          .onClick(() => this.onSearch())

          Button() {
            Image(this.isSelectionMode ? $r('app.media.ic_close') : $r('app.media.ic_select'))
              .width(24)
              .height(24)
              .fillColor(Color.Black)
          }
          .backgroundColor(Color.Transparent)
          .onClick(() => this.toggleSelectionMode())
        }
      }
      .width('100%')
      .height(56)
      .padding({ left: 16, right: 16 })
      .justifyContent(FlexAlign.SpaceBetween)

      // 选择模式工具栏
      if (this.isSelectionMode) {
        Row() {
          Button(`全选 (${this.selectedTrajectories.size}/${this.filteredTrajectories.length})`)
            .fontSize(14)
            .backgroundColor(Color.Transparent)
            .fontColor('#007AFF')
            .onClick(() => this.toggleSelectAll())

          Blank()

          Row({ space: 8 }) {
            Button('导出')
              .fontSize(12)
              .backgroundColor('#34C759')
              .fontColor(Color.White)
              .padding({ left: 12, right: 12, top: 6, bottom: 6 })
              .enabled(this.selectedTrajectories.size > 0)
              .onClick(() => this.exportSelectedTrajectories())

            Button('分享')
              .fontSize(12)
              .backgroundColor('#007AFF')
              .fontColor(Color.White)
              .padding({ left: 12, right: 12, top: 6, bottom: 6 })
              .enabled(this.selectedTrajectories.size > 0)
              .onClick(() => this.shareSelectedTrajectories())

            Button('删除')
              .fontSize(12)
              .backgroundColor('#FF3B30')
              .fontColor(Color.White)
              .padding({ left: 12, right: 12, top: 6, bottom: 6 })
              .enabled(this.selectedTrajectories.size > 0)
              .onClick(() => this.deleteSelectedTrajectories())
          }
        }
        .width('100%')
        .height(44)
        .padding({ left: 16, right: 16 })
        .backgroundColor('#F8F8F8')
      }

      // 轨迹列表
      if (this.filteredTrajectories.length === 0 && !this.isLoading) {
        Column() {
          Image($r('app.media.ic_empty_trajectory'))
            .width(120)
            .height(120)
            .opacity(0.3)

          Text('暂无轨迹记录')
            .fontSize(16)
            .fontColor('#999999')
            .margin({ top: 16 })

          Text('开始记录您的行程轨迹')
            .fontSize(14)
            .fontColor('#CCCCCC')
            .margin({ top: 8 })
        }
        .width('100%')
        .layoutWeight(1)
        .justifyContent(FlexAlign.Center)
      } else {
        Refresh({ refreshing: this.isRefreshing, offset: 64, friction: 100 }) {
          List() {
            ForEach(this.filteredTrajectories, (trajectory: Trajectory) => {
              ListItem() {
                this.buildTrajectoryItem(trajectory)
              }
              .onClick(() => this.navigateToDetail(trajectory))
            })

            // 加载更多指示器
            if (this.pagination.hasMore) {
              ListItem() {
                Row() {
                  LoadingProgress()
                    .width(20)
                    .height(20)
                    .margin({ right: 8 })

                  Text('加载更多...')
                    .fontSize(14)
                    .fontColor('#999999')
                }
                .width('100%')
                .height(60)
                .justifyContent(FlexAlign.Center)
              }
              .onClick(() => this.loadMore())
            }
          }
          .width('100%')
          .layoutWeight(1)
          .divider({ strokeWidth: 1, color: '#F0F0F0' })
        }
        .onRefreshing(() => this.onRefresh())
      }

      // 搜索对话框
      if (this.showSearchDialog) {
        this.buildSearchDialog()
      }

      // 导出对话框
      if (this.showExportDialog) {
        this.buildExportDialog()
      }

      // 分享对话框
      if (this.showShareDialog) {
        this.buildShareDialog()
      }
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#FFFFFF')
  }

  @Builder
  buildTrajectoryItem(trajectory: Trajectory) {
    Row() {
      // 选择框
      if (this.isSelectionMode) {
        Checkbox({ name: trajectory.id, group: 'trajectories' })
          .select(this.selectedTrajectories.has(trajectory.id))
          .onChange((value: boolean) => {
            if (value) {
              this.selectedTrajectories.add(trajectory.id)
            } else {
              this.selectedTrajectories.delete(trajectory.id)
            }
          })
          .margin({ right: 12 })
      }

      // 轨迹图标
      Column() {
        Image($r('app.media.ic_trajectory'))
          .width(40)
          .height(40)
          .fillColor('#007AFF')
      }
      .width(56)
      .height(56)
      .backgroundColor('#F0F8FF')
      .borderRadius(8)
      .justifyContent(FlexAlign.Center)

      // 轨迹信息
      Column() {
        Row() {
          Text(trajectory.name || '未命名轨迹')
            .fontSize(16)
            .fontWeight(FontWeight.Medium)
            .maxLines(1)
            .textOverflow({ overflow: TextOverflow.Ellipsis })
            .layoutWeight(1)

          Text(this.formatTime(trajectory.startTime))
            .fontSize(12)
            .fontColor('#999999')
        }
        .width('100%')
        .margin({ bottom: 4 })

        Row() {
          Text(`${this.formatDistance(trajectory.totalDistance)} · ${this.formatDuration(trajectory.duration)}`)
            .fontSize(14)
            .fontColor('#666666')
            .layoutWeight(1)

          if (trajectory.isCompleted) {
            Text('已完成')
              .fontSize(12)
              .fontColor('#34C759')
              .backgroundColor('#E8F5E8')
              .padding({ left: 8, right: 8, top: 2, bottom: 2 })
              .borderRadius(4)
          } else {
            Text('进行中')
              .fontSize(12)
              .fontColor('#FF9500')
              .backgroundColor('#FFF4E6')
              .padding({ left: 8, right: 8, top: 2, bottom: 2 })
              .borderRadius(4)
          }
        }
        .width('100%')

        if (trajectory.description) {
          Text(trajectory.description)
            .fontSize(12)
            .fontColor('#999999')
            .maxLines(2)
            .textOverflow({ overflow: TextOverflow.Ellipsis })
            .margin({ top: 4 })
        }

        // 类型标签
        if (trajectory.activityType || trajectory.trajectoryType) {
          Row({ space: 6 }) {
            if (trajectory.activityType) {
              Text(this.getActivityTypeText(trajectory.activityType))
                .fontSize(10)
                .fontColor('#007AFF')
                .backgroundColor('#E8F4FD')
                .padding({ left: 6, right: 6, top: 2, bottom: 2 })
                .borderRadius(8)
            }
            if (trajectory.trajectoryType) {
              Text(this.getTrajectoryTypeText(trajectory.trajectoryType))
                .fontSize(10)
                .fontColor('#34C759')
                .backgroundColor('#E8F5E8')
                .padding({ left: 6, right: 6, top: 2, bottom: 2 })
                .borderRadius(8)
            }
          }
          .margin({ top: 4 })
        }
      }
      .layoutWeight(1)
      .margin({ left: 12, right: 8 })
      .alignItems(HorizontalAlign.Start)

      // 箭头图标
      if (!this.isSelectionMode) {
        Image($r('app.media.ic_arrow_right'))
          .width(16)
          .height(16)
          .fillColor('#CCCCCC')
      }
    }
    .width('100%')
    .padding(16)
    .alignItems(VerticalAlign.Center)
  }

  @Builder
  buildSearchDialog() {
    Column() {
      // 对话框背景遮罩
      Column()
        .width('100%')
        .height('100%')
        .backgroundColor('rgba(0, 0, 0, 0.5)')
        .onClick(() => this.showSearchDialog = false)

      // 搜索对话框内容
      Column() {
        Text('搜索和过滤')
          .fontSize(18)
          .fontWeight(FontWeight.Medium)
          .margin({ bottom: 20 })

        // 关键词搜索
        TextInput({ placeholder: '搜索轨迹名称或描述' })
          .width('100%')
          .height(44)
          .text(this.searchFilters.keyword)
          .onChange((value: string) => {
            this.searchFilters.keyword = value
          })
          .margin({ bottom: 16 })

        // 距离范围
        Row() {
          Text('距离范围:')
            .fontSize(14)
            .fontColor('#333333')
            .width(80)

          TextInput({ placeholder: '最小' })
            .width(80)
            .height(36)
            .text(this.searchFilters.minDistance.toString())
            .type(InputType.Number)
            .onChange((value: string) => {
              this.searchFilters.minDistance = parseInt(value) || 0
            })

          Text(' - ')
            .margin({ left: 8, right: 8 })

          TextInput({ placeholder: '最大' })
            .width(80)
            .height(36)
            .text(this.searchFilters.maxDistance.toString())
            .type(InputType.Number)
            .onChange((value: string) => {
              this.searchFilters.maxDistance = parseInt(value) || 100000
            })

          Text('米')
            .fontSize(14)
            .fontColor('#666666')
            .margin({ left: 8 })
        }
        .width('100%')
        .margin({ bottom: 16 })
        .alignItems(VerticalAlign.Center)

        // 时长范围
        Row() {
          Text('时长范围:')
            .fontSize(14)
            .fontColor('#333333')
            .width(80)

          TextInput({ placeholder: '最小' })
            .width(80)
            .height(36)
            .text((this.searchFilters.minDuration / 60000).toString()) // 转换为分钟
            .type(InputType.Number)
            .onChange((value: string) => {
              this.searchFilters.minDuration = (parseInt(value) || 0) * 60000 // 转换为毫秒
            })

          Text(' - ')
            .margin({ left: 8, right: 8 })

          TextInput({ placeholder: '最大' })
            .width(80)
            .height(36)
            .text((this.searchFilters.maxDuration / 60000).toString()) // 转换为分钟
            .type(InputType.Number)
            .onChange((value: string) => {
              this.searchFilters.maxDuration = (parseInt(value) || 1440) * 60000 // 转换为毫秒
            })

          Text('分钟')
            .fontSize(14)
            .fontColor('#666666')
            .margin({ left: 8 })
        }
        .width('100%')
        .margin({ bottom: 16 })
        .alignItems(VerticalAlign.Center)

        // 活动类型
        Row() {
          Text('活动类型:')
            .fontSize(14)
            .fontColor('#333333')
            .width(80)

          Select([
            { value: '', text: '全部' },
            { value: 'walking', text: '步行' },
            { value: 'running', text: '跑步' },
            { value: 'cycling', text: '骑行' },
            { value: 'driving', text: '驾车' }
          ])
            .selected(this.searchFilters.activityType === null ? 0 :
                     this.searchFilters.activityType === 'walking' ? 1 :
                     this.searchFilters.activityType === 'running' ? 2 :
                     this.searchFilters.activityType === 'cycling' ? 3 : 4)
            .value(this.searchFilters.activityType === null ? '全部' :
                   this.searchFilters.activityType === 'walking' ? '步行' :
                   this.searchFilters.activityType === 'running' ? '跑步' :
                   this.searchFilters.activityType === 'cycling' ? '骑行' : '驾车')
            .onSelect((index: number) => {
              const types = [null, 'walking', 'running', 'cycling', 'driving']
              this.searchFilters.activityType = types[index]
            })
            .width(120)
        }
        .width('100%')
        .margin({ bottom: 16 })
        .alignItems(VerticalAlign.Center)

        // 轨迹类型
        Row() {
          Text('轨迹类型:')
            .fontSize(14)
            .fontColor('#333333')
            .width(80)

          Select([
            { value: '', text: '全部' },
            { value: 'exercise', text: '运动' },
            { value: 'commute', text: '通勤' },
            { value: 'travel', text: '旅行' },
            { value: 'leisure', text: '休闲' }
          ])
            .selected(this.searchFilters.trajectoryType === null ? 0 :
                     this.searchFilters.trajectoryType === 'exercise' ? 1 :
                     this.searchFilters.trajectoryType === 'commute' ? 2 :
                     this.searchFilters.trajectoryType === 'travel' ? 3 : 4)
            .value(this.searchFilters.trajectoryType === null ? '全部' :
                   this.searchFilters.trajectoryType === 'exercise' ? '运动' :
                   this.searchFilters.trajectoryType === 'commute' ? '通勤' :
                   this.searchFilters.trajectoryType === 'travel' ? '旅行' : '休闲')
            .onSelect((index: number) => {
              const types = [null, 'exercise', 'commute', 'travel', 'leisure']
              this.searchFilters.trajectoryType = types[index]
            })
            .width(120)
        }
        .width('100%')
        .margin({ bottom: 16 })
        .alignItems(VerticalAlign.Center)

        // 排序方式
        Row() {
          Text('排序:')
            .fontSize(14)
            .fontColor('#333333')
            .width(80)

          Select([
            { value: 'date', text: '按时间' },
            { value: 'distance', text: '按距离' },
            { value: 'duration', text: '按时长' },
            { value: 'speed', text: '按速度' }
          ])
            .selected(this.searchFilters.sortBy === 'date' ? 0 : 
                     this.searchFilters.sortBy === 'distance' ? 1 : 
                     this.searchFilters.sortBy === 'duration' ? 2 : 3)
            .value(this.searchFilters.sortBy === 'date' ? '按时间' :
                   this.searchFilters.sortBy === 'distance' ? '按距离' : 
                   this.searchFilters.sortBy === 'duration' ? '按时长' : '按速度')
            .onSelect((index: number) => {
              this.searchFilters.sortBy = ['date', 'distance', 'duration', 'speed'][index] as 'date' | 'distance' | 'duration' | 'speed'
            })
            .width(100)
            .margin({ right: 8 })

          Select([
            { value: 'desc', text: '降序' },
            { value: 'asc', text: '升序' }
          ])
            .selected(this.searchFilters.sortOrder === 'desc' ? 0 : 1)
            .value(this.searchFilters.sortOrder === 'desc' ? '降序' : '升序')
            .onSelect((index: number) => {
              this.searchFilters.sortOrder = index === 0 ? 'desc' : 'asc'
            })
            .width(80)
        }
        .width('100%')
        .margin({ bottom: 20 })
        .alignItems(VerticalAlign.Center)

        // 按钮组
        Row() {
          Button('重置')
            .fontSize(16)
            .backgroundColor('#F0F0F0')
            .fontColor('#333333')
            .layoutWeight(1)
            .onClick(() => this.resetFilters())

          Button('应用')
            .fontSize(16)
            .backgroundColor('#007AFF')
            .fontColor(Color.White)
            .layoutWeight(1)
            .margin({ left: 12 })
            .onClick(() => this.applySearchFilters())
        }
        .width('100%')
      }
      .width('90%')
      .backgroundColor(Color.White)
      .borderRadius(12)
      .padding(20)
      .position({ x: '5%', y: '20%' })
    }
    .width('100%')
    .height('100%')
    .position({ x: 0, y: 0 })
  }

  /**
   * 构建导出对话框
   */
  @Builder
  buildExportDialog() {
    // 背景遮罩
    Column()
      .width('100%')
      .height('100%')
      .backgroundColor('rgba(0, 0, 0, 0.5)')
      .onClick(() => {
        this.showExportDialog = false
      })
      .position({ x: 0, y: 0 })

    // 对话框内容
    Column({ space: 20 }) {
      Text('导出轨迹')
        .fontSize(18)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')

      Text(`选择导出格式 (${this.selectedTrajectories.size} 条轨迹)`)
        .fontSize(14)
        .fontColor('#666666')

      // 格式选择
      Column({ space: 12 }) {
        ForEach([
          { format: ExportFormat.GPX, name: 'GPX', desc: '通用GPS格式，兼容性好' },
          { format: ExportFormat.KML, name: 'KML', desc: '谷歌地球格式，支持地图显示' },
          { format: ExportFormat.JSON, name: 'JSON', desc: '数据格式，便于程序处理' },
          { format: ExportFormat.CSV, name: 'CSV', desc: '表格格式，便于数据分析' }
        ], (item: any) => {
          Row() {
            Radio({ value: item.format, group: 'exportFormat' })
              .checked(this.exportFormat === item.format)
              .onChange((isChecked: boolean) => {
                if (isChecked) {
                  this.exportFormat = item.format
                }
              })

            Column({ space: 4 }) {
              Text(item.name)
                .fontSize(16)
                .fontColor('#333333')
                .alignSelf(ItemAlign.Start)

              Text(item.desc)
                .fontSize(12)
                .fontColor('#999999')
                .alignSelf(ItemAlign.Start)
            }
            .layoutWeight(1)
            .alignItems(HorizontalAlign.Start)
            .margin({ left: 12 })
          }
          .width('100%')
          .padding(8)
          .onClick(() => {
            this.exportFormat = item.format
          })
        })
      }

      // 按钮组
      Row() {
        Button('取消')
          .fontSize(16)
          .backgroundColor('#F0F0F0')
          .fontColor('#333333')
          .layoutWeight(1)
          .onClick(() => {
            this.showExportDialog = false
          })

        Button(this.isExporting ? '导出中...' : '导出')
          .fontSize(16)
          .backgroundColor('#34C759')
          .fontColor(Color.White)
          .layoutWeight(1)
          .margin({ left: 12 })
          .enabled(!this.isExporting)
          .onClick(() => this.performExport())
      }
      .width('100%')
    }
    .width('85%')
    .backgroundColor(Color.White)
    .borderRadius(12)
    .padding(20)
    .position({ x: '7.5%', y: '25%' })
  }

  /**
   * 构建分享对话框
   */
  @Builder
  buildShareDialog() {
    // 背景遮罩
    Column()
      .width('100%')
      .height('100%')
      .backgroundColor('rgba(0, 0, 0, 0.5)')
      .onClick(() => {
        this.showShareDialog = false
      })
      .position({ x: 0, y: 0 })

    // 对话框内容
    Column({ space: 20 }) {
      Text('分享轨迹')
        .fontSize(18)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')

      Text(`选择分享格式 (${this.selectedTrajectories.size} 条轨迹)`)
        .fontSize(14)
        .fontColor('#666666')

      // 格式选择
      Column({ space: 12 }) {
        ForEach([
          { format: ExportFormat.GPX, name: 'GPX', desc: '通用GPS格式' },
          { format: ExportFormat.KML, name: 'KML', desc: '谷歌地球格式' },
          { format: ExportFormat.JSON, name: 'JSON', desc: '数据格式' }
        ], (item: any) => {
          Row() {
            Radio({ value: item.format, group: 'shareFormat' })
              .checked(this.exportFormat === item.format)
              .onChange((isChecked: boolean) => {
                if (isChecked) {
                  this.exportFormat = item.format
                }
              })

            Column({ space: 4 }) {
              Text(item.name)
                .fontSize(16)
                .fontColor('#333333')
                .alignSelf(ItemAlign.Start)

              Text(item.desc)
                .fontSize(12)
                .fontColor('#999999')
                .alignSelf(ItemAlign.Start)
            }
            .layoutWeight(1)
            .alignItems(HorizontalAlign.Start)
            .margin({ left: 12 })
          }
          .width('100%')
          .padding(8)
          .onClick(() => {
            this.exportFormat = item.format
          })
        })
      }

      // 按钮组
      Row() {
        Button('取消')
          .fontSize(16)
          .backgroundColor('#F0F0F0')
          .fontColor('#333333')
          .layoutWeight(1)
          .onClick(() => {
            this.showShareDialog = false
          })

        Button('分享')
          .fontSize(16)
          .backgroundColor('#007AFF')
          .fontColor(Color.White)
          .layoutWeight(1)
          .margin({ left: 12 })
          .onClick(() => this.performShare())
      }
      .width('100%')
    }
    .width('85%')
    .backgroundColor(Color.White)
    .borderRadius(12)
    .padding(20)
    .position({ x: '7.5%', y: '25%' })
  }
}