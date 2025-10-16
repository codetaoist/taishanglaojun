import { DataExportService, ExportType, ExportFormat, ExportOptions, ExportResult } from '../service/DataExportService';

/**
 * 数据导出对话框组件
 */
@Component
export struct DataExportDialog {
  @State private isVisible: boolean = false;
  @State private selectedType: ExportType = ExportType.ALL_DATA;
  @State private selectedFormat: ExportFormat = ExportFormat.JSON;
  @State private startDate: string = '';
  @State private endDate: string = '';
  @State private includeMetadata: boolean = true;
  @State private isExporting: boolean = false;
  @State private exportProgress: number = 0;
  @State private exportMessage: string = '';
  @State private exportResult: ExportResult | null = null;

  private exportService: DataExportService = DataExportService.getInstance();
  private onComplete?: (result: ExportResult) => void;
  private onCancel?: () => void;

  /**
   * 显示对话框
   */
  show(callbacks?: {
    onComplete?: (result: ExportResult) => void;
    onCancel?: () => void;
  }) {
    this.onComplete = callbacks?.onComplete;
    this.onCancel = callbacks?.onCancel;
    this.isVisible = true;
    this.resetForm();
  }

  /**
   * 隐藏对话框
   */
  hide() {
    this.isVisible = false;
    this.resetForm();
  }

  /**
   * 重置表单
   */
  private resetForm() {
    this.selectedType = ExportType.ALL_DATA;
    this.selectedFormat = ExportFormat.JSON;
    this.startDate = '';
    this.endDate = '';
    this.includeMetadata = true;
    this.isExporting = false;
    this.exportProgress = 0;
    this.exportMessage = '';
    this.exportResult = null;
  }

  /**
   * 开始导出
   */
  private async startExport() {
    try {
      this.isExporting = true;
      this.exportProgress = 0;
      this.exportMessage = '准备导出...';
      this.exportResult = null;

      const options: ExportOptions = {
        type: this.selectedType,
        format: this.selectedFormat,
        includeMetadata: this.includeMetadata
      };

      // 添加日期范围
      if (this.startDate && this.endDate) {
        options.dateRange = {
          startDate: new Date(this.startDate).getTime(),
          endDate: new Date(this.endDate).getTime()
        };
      }

      const result = await this.exportService.exportData(options, {
        onProgress: (progress: number, message: string) => {
          this.exportProgress = progress;
          this.exportMessage = message;
        },
        onComplete: (result: ExportResult) => {
          this.exportResult = result;
          this.isExporting = false;
          this.onComplete?.(result);
        },
        onError: (error: string) => {
          this.exportResult = {
            success: false,
            error: error
          };
          this.isExporting = false;
        }
      });

    } catch (error) {
      console.error('Export failed:', error);
      this.exportResult = {
        success: false,
        error: error instanceof Error ? error.message : '导出失败'
      };
      this.isExporting = false;
    }
  }

  /**
   * 取消导出
   */
  private cancel() {
    this.onCancel?.();
    this.hide();
  }

  /**
   * 获取导出类型显示名称
   */
  private getExportTypeDisplayName(type: ExportType): string {
    switch (type) {
      case ExportType.CHAT_MESSAGES:
        return '聊天消息';
      case ExportType.CONVERSATIONS:
        return '对话列表';
      case ExportType.TRAJECTORIES:
        return '轨迹数据';
      case ExportType.LOCATION_POINTS:
        return '位置点';
      case ExportType.ALL_DATA:
        return '所有数据';
      default:
        return '未知类型';
    }
  }

  /**
   * 获取导出格式显示名称
   */
  private getExportFormatDisplayName(format: ExportFormat): string {
    switch (format) {
      case ExportFormat.JSON:
        return 'JSON格式';
      case ExportFormat.CSV:
        return 'CSV格式';
      case ExportFormat.TXT:
        return '文本格式';
      default:
        return '未知格式';
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

  build() {
    if (!this.isVisible) {
      return;
    }

    Stack({ alignContent: Alignment.Center }) {
      // 背景遮罩
      Column()
        .width('100%')
        .height('100%')
        .backgroundColor(Color.Black)
        .opacity(0.5)
        .onClick(() => {
          if (!this.isExporting) {
            this.cancel();
          }
        })

      // 对话框内容
      Column({ space: 20 }) {
        // 标题
        Text('数据导出')
          .fontSize(20)
          .fontWeight(FontWeight.Bold)
          .fontColor(Color.Black)

        if (!this.isExporting && !this.exportResult) {
          // 导出配置表单
          Column({ space: 16 }) {
            // 导出类型选择
            Column({ space: 8 }) {
              Text('导出类型')
                .fontSize(16)
                .fontWeight(FontWeight.Medium)
                .fontColor(Color.Black)
                .alignSelf(ItemAlign.Start)

              Column({ space: 8 }) {
                ForEach([
                  ExportType.ALL_DATA,
                  ExportType.CHAT_MESSAGES,
                  ExportType.CONVERSATIONS,
                  ExportType.TRAJECTORIES,
                  ExportType.LOCATION_POINTS
                ], (type: ExportType) => {
                  Row({ space: 8 }) {
                    Radio({ value: type, group: 'exportType' })
                      .checked(this.selectedType === type)
                      .onChange((isChecked: boolean) => {
                        if (isChecked) {
                          this.selectedType = type;
                        }
                      })

                    Text(this.getExportTypeDisplayName(type))
                      .fontSize(14)
                      .fontColor(Color.Black)
                      .layoutWeight(1)
                  }
                  .width('100%')
                  .onClick(() => {
                    this.selectedType = type;
                  })
                })
              }
            }

            // 导出格式选择
            Column({ space: 8 }) {
              Text('导出格式')
                .fontSize(16)
                .fontWeight(FontWeight.Medium)
                .fontColor(Color.Black)
                .alignSelf(ItemAlign.Start)

              Row({ space: 16 }) {
                ForEach([
                  ExportFormat.JSON,
                  ExportFormat.CSV,
                  ExportFormat.TXT
                ], (format: ExportFormat) => {
                  Row({ space: 8 }) {
                    Radio({ value: format, group: 'exportFormat' })
                      .checked(this.selectedFormat === format)
                      .onChange((isChecked: boolean) => {
                        if (isChecked) {
                          this.selectedFormat = format;
                        }
                      })

                    Text(this.getExportFormatDisplayName(format))
                      .fontSize(14)
                      .fontColor(Color.Black)
                  }
                  .onClick(() => {
                    this.selectedFormat = format;
                  })
                })
              }
              .width('100%')
            }

            // 日期范围选择
            Column({ space: 8 }) {
              Text('日期范围（可选）')
                .fontSize(16)
                .fontWeight(FontWeight.Medium)
                .fontColor(Color.Black)
                .alignSelf(ItemAlign.Start)

              Row({ space: 12 }) {
                Column({ space: 4 }) {
                  Text('开始日期')
                    .fontSize(12)
                    .fontColor(Color.Gray)
                  
                  TextInput({ placeholder: 'YYYY-MM-DD' })
                    .width(120)
                    .height(40)
                    .fontSize(14)
                    .onChange((value: string) => {
                      this.startDate = value;
                    })
                }

                Text('至')
                  .fontSize(14)
                  .fontColor(Color.Gray)

                Column({ space: 4 }) {
                  Text('结束日期')
                    .fontSize(12)
                    .fontColor(Color.Gray)
                  
                  TextInput({ placeholder: 'YYYY-MM-DD' })
                    .width(120)
                    .height(40)
                    .fontSize(14)
                    .onChange((value: string) => {
                      this.endDate = value;
                    })
                }
              }
              .width('100%')
              .justifyContent(FlexAlign.SpaceBetween)
            }

            // 其他选项
            Column({ space: 8 }) {
              Row({ space: 8 }) {
                Checkbox()
                  .select(this.includeMetadata)
                  .onChange((value: boolean) => {
                    this.includeMetadata = value;
                  })

                Text('包含元数据')
                  .fontSize(14)
                  .fontColor(Color.Black)
                  .layoutWeight(1)
              }
              .width('100%')
              .onClick(() => {
                this.includeMetadata = !this.includeMetadata;
              })
            }
          }

          // 操作按钮
          Row({ space: 12 }) {
            Button('取消')
              .width(80)
              .height(40)
              .backgroundColor(Color.Gray)
              .fontColor(Color.White)
              .onClick(() => this.cancel())

            Button('开始导出')
              .width(100)
              .height(40)
              .backgroundColor('#007AFF')
              .fontColor(Color.White)
              .onClick(() => this.startExport())
          }
          .width('100%')
          .justifyContent(FlexAlign.End)

        } else if (this.isExporting) {
          // 导出进度
          Column({ space: 16 }) {
            Progress({ value: this.exportProgress, total: 100, type: ProgressType.Linear })
              .width('100%')
              .height(8)
              .color('#007AFF')

            Text(`${this.exportProgress.toFixed(0)}%`)
              .fontSize(16)
              .fontWeight(FontWeight.Medium)
              .fontColor(Color.Black)

            Text(this.exportMessage)
              .fontSize(14)
              .fontColor(Color.Gray)
              .maxLines(2)
              .textAlign(TextAlign.Center)
          }
          .width('100%')

        } else if (this.exportResult) {
          // 导出结果
          Column({ space: 16 }) {
            if (this.exportResult.success) {
              Column({ space: 12 }) {
                Image($r('app.media.ic_check_circle'))
                  .width(48)
                  .height(48)
                  .fillColor(Color.Green)

                Text('导出成功！')
                  .fontSize(18)
                  .fontWeight(FontWeight.Medium)
                  .fontColor(Color.Green)

                if (this.exportResult.fileName) {
                  Text(`文件名：${this.exportResult.fileName}`)
                    .fontSize(14)
                    .fontColor(Color.Black)
                    .maxLines(2)
                    .textAlign(TextAlign.Center)
                }

                if (this.exportResult.fileSize) {
                  Text(`文件大小：${this.formatFileSize(this.exportResult.fileSize)}`)
                    .fontSize(14)
                    .fontColor(Color.Gray)
                }

                if (this.exportResult.recordCount) {
                  Text(`记录数量：${this.exportResult.recordCount}`)
                    .fontSize(14)
                    .fontColor(Color.Gray)
                }
              }
            } else {
              Column({ space: 12 }) {
                Image($r('app.media.ic_error_circle'))
                  .width(48)
                  .height(48)
                  .fillColor(Color.Red)

                Text('导出失败')
                  .fontSize(18)
                  .fontWeight(FontWeight.Medium)
                  .fontColor(Color.Red)

                if (this.exportResult.error) {
                  Text(this.exportResult.error)
                    .fontSize(14)
                    .fontColor(Color.Red)
                    .maxLines(3)
                    .textAlign(TextAlign.Center)
                }
              }
            }

            Button('确定')
              .width(100)
              .height(40)
              .backgroundColor('#007AFF')
              .fontColor(Color.White)
              .onClick(() => this.hide())
          }
          .width('100%')
        }
      }
      .width('90%')
      .maxWidth(400)
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
  }
}