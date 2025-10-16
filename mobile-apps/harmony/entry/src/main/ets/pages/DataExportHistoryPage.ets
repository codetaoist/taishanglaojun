import { DataExportService, ExportHistory } from '../service/DataExportService';
import { DataExportDialog } from '../components/DataExportDialog';

/**
 * 数据导出历史页面
 */
@Entry
@Component
struct DataExportHistoryPage {
  @State private exportHistory: ExportHistory[] = [];
  @State private isLoading: boolean = true;
  @State private selectedItems: Set<string> = new Set();
  @State private isSelectionMode: boolean = false;

  private exportService: DataExportService = DataExportService.getInstance();
  private exportDialog: DataExportDialog = new DataExportDialog();

  aboutToAppear() {
    this.loadExportHistory();
  }

  /**
   * 加载导出历史
   */
  private async loadExportHistory() {
    try {
      this.isLoading = true;
      this.exportHistory = await this.exportService.getExportHistory();
    } catch (error) {
      console.error('Failed to load export history:', error);
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * 开始新的导出
   */
  private startNewExport() {
    this.exportDialog.show({
      onComplete: (result) => {
        if (result.success) {
          this.loadExportHistory(); // 刷新历史记录
        }
      }
    });
  }

  /**
   * 删除导出记录
   */
  private async deleteExportRecord(id: string) {
    try {
      await this.exportService.deleteExportHistory(id);
      await this.loadExportHistory();
    } catch (error) {
      console.error('Failed to delete export record:', error);
    }
  }

  /**
   * 批量删除选中的记录
   */
  private async deleteSelectedRecords() {
    try {
      const deletePromises = Array.from(this.selectedItems).map(id => 
        this.exportService.deleteExportHistory(id)
      );
      await Promise.all(deletePromises);
      this.selectedItems.clear();
      this.isSelectionMode = false;
      await this.loadExportHistory();
    } catch (error) {
      console.error('Failed to delete selected records:', error);
    }
  }

  /**
   * 切换选择模式
   */
  private toggleSelectionMode() {
    this.isSelectionMode = !this.isSelectionMode;
    if (!this.isSelectionMode) {
      this.selectedItems.clear();
    }
  }

  /**
   * 切换项目选择状态
   */
  private toggleItemSelection(id: string) {
    if (this.selectedItems.has(id)) {
      this.selectedItems.delete(id);
    } else {
      this.selectedItems.add(id);
    }
  }

  /**
   * 全选/取消全选
   */
  private toggleSelectAll() {
    if (this.selectedItems.size === this.exportHistory.length) {
      this.selectedItems.clear();
    } else {
      this.selectedItems.clear();
      this.exportHistory.forEach(item => this.selectedItems.add(item.id));
    }
  }

  /**
   * 格式化文件大小
   */
  private formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * 格式化时间
   */
  private formatTime(timestamp: number): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - timestamp;
    
    if (diff < 60000) { // 1分钟内
      return '刚刚';
    } else if (diff < 3600000) { // 1小时内
      return `${Math.floor(diff / 60000)}分钟前`;
    } else if (diff < 86400000) { // 1天内
      return `${Math.floor(diff / 3600000)}小时前`;
    } else if (diff < 604800000) { // 1周内
      return `${Math.floor(diff / 86400000)}天前`;
    } else {
      return date.toLocaleDateString('zh-CN');
    }
  }

  /**
   * 获取导出类型显示名称
   */
  private getExportTypeDisplayName(type: string): string {
    switch (type) {
      case 'CHAT_MESSAGES':
        return '聊天消息';
      case 'CONVERSATIONS':
        return '对话列表';
      case 'TRAJECTORIES':
        return '轨迹数据';
      case 'LOCATION_POINTS':
        return '位置点';
      case 'ALL_DATA':
        return '所有数据';
      default:
        return '未知类型';
    }
  }

  /**
   * 获取导出格式显示名称
   */
  private getExportFormatDisplayName(format: string): string {
    switch (format) {
      case 'JSON':
        return 'JSON';
      case 'CSV':
        return 'CSV';
      case 'TXT':
        return 'TXT';
      default:
        return '未知';
    }
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
              // 返回上一页
            })

          Text('数据导出')
            .fontSize(18)
            .fontWeight(FontWeight.Medium)
            .fontColor(Color.Black)
        }

        Row({ space: 12 }) {
          if (this.isSelectionMode) {
            Text(`已选择 ${this.selectedItems.size}`)
              .fontSize(14)
              .fontColor('#007AFF')

            if (this.selectedItems.size > 0) {
              Image($r('app.media.ic_delete'))
                .width(24)
                .height(24)
                .fillColor(Color.Red)
                .onClick(() => this.deleteSelectedRecords())
            }

            Text('取消')
              .fontSize(14)
              .fontColor('#007AFF')
              .onClick(() => this.toggleSelectionMode())
          } else {
            Text('选择')
              .fontSize(14)
              .fontColor('#007AFF')
              .onClick(() => this.toggleSelectionMode())
              .visibility(this.exportHistory.length > 0 ? Visibility.Visible : Visibility.Hidden)

            Image($r('app.media.ic_add'))
              .width(24)
              .height(24)
              .fillColor('#007AFF')
              .onClick(() => this.startNewExport())
          }
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

      } else if (this.exportHistory.length === 0) {
        // 空状态
        Column({ space: 20 }) {
          Image($r('app.media.ic_export_empty'))
            .width(80)
            .height(80)
            .fillColor(Color.Gray)

          Text('暂无导出记录')
            .fontSize(16)
            .fontColor(Color.Gray)

          Text('点击右上角的 + 按钮开始导出数据')
            .fontSize(14)
            .fontColor('#999999')
            .textAlign(TextAlign.Center)

          Button('开始导出')
            .width(120)
            .height(40)
            .backgroundColor('#007AFF')
            .fontColor(Color.White)
            .borderRadius(20)
            .onClick(() => this.startNewExport())
        }
        .width('100%')
        .height('100%')
        .justifyContent(FlexAlign.Center)
        .alignItems(HorizontalAlign.Center)
        .padding(40)

      } else {
        // 导出历史列表
        Column() {
          // 批量操作栏
          if (this.isSelectionMode) {
            Row({ space: 12 }) {
              Checkbox()
                .select(this.selectedItems.size === this.exportHistory.length)
                .onChange(() => this.toggleSelectAll())

              Text('全选')
                .fontSize(14)
                .fontColor(Color.Black)
                .layoutWeight(1)

              if (this.selectedItems.size > 0) {
                Text(`删除选中的 ${this.selectedItems.size} 项`)
                  .fontSize(14)
                  .fontColor(Color.Red)
                  .onClick(() => this.deleteSelectedRecords())
              }
            }
            .width('100%')
            .height(48)
            .padding({ left: 16, right: 16 })
            .backgroundColor('#F8F8F8')
            .border({ width: { bottom: 0.5 }, color: '#E5E5E5' })
          }

          // 历史记录列表
          List({ space: 0 }) {
            ForEach(this.exportHistory, (item: ExportHistory) => {
              ListItem() {
                Row({ space: 12 }) {
                  if (this.isSelectionMode) {
                    Checkbox()
                      .select(this.selectedItems.has(item.id))
                      .onChange(() => this.toggleItemSelection(item.id))
                  }

                  // 导出类型图标
                  Column() {
                    Image($r('app.media.ic_export_file'))
                      .width(24)
                      .height(24)
                      .fillColor('#007AFF')
                  }
                  .width(40)
                  .height(40)
                  .backgroundColor('#F0F8FF')
                  .borderRadius(8)
                  .justifyContent(FlexAlign.Center)

                  // 导出信息
                  Column({ space: 4 }) {
                    Row({ space: 8 }) {
                      Text(this.getExportTypeDisplayName(item.type))
                        .fontSize(16)
                        .fontWeight(FontWeight.Medium)
                        .fontColor(Color.Black)
                        .layoutWeight(1)

                      Text(this.getExportFormatDisplayName(item.format))
                        .fontSize(12)
                        .fontColor(Color.White)
                        .backgroundColor('#007AFF')
                        .padding({ left: 6, right: 6, top: 2, bottom: 2 })
                        .borderRadius(4)
                    }
                    .width('100%')

                    if (item.fileName) {
                      Text(item.fileName)
                        .fontSize(14)
                        .fontColor(Color.Gray)
                        .maxLines(1)
                        .textOverflow({ overflow: TextOverflow.Ellipsis })
                    }

                    Row({ space: 12 }) {
                      Text(this.formatTime(item.timestamp))
                        .fontSize(12)
                        .fontColor('#999999')

                      if (item.fileSize) {
                        Text(this.formatFileSize(item.fileSize))
                          .fontSize(12)
                          .fontColor('#999999')
                      }

                      if (item.recordCount) {
                        Text(`${item.recordCount} 条记录`)
                          .fontSize(12)
                          .fontColor('#999999')
                      }
                    }
                  }
                  .layoutWeight(1)
                  .alignItems(HorizontalAlign.Start)

                  if (!this.isSelectionMode) {
                    // 删除按钮
                    Image($r('app.media.ic_delete'))
                      .width(20)
                      .height(20)
                      .fillColor(Color.Red)
                      .onClick(() => this.deleteExportRecord(item.id))
                  }
                }
                .width('100%')
                .padding({ left: 16, right: 16, top: 12, bottom: 12 })
                .onClick(() => {
                  if (this.isSelectionMode) {
                    this.toggleItemSelection(item.id);
                  }
                })
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

      // 数据导出对话框
      this.exportDialog
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F8F8F8')
  }
}