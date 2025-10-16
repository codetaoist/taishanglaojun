import { ChatMessage, MessageType, MessageSender, MessageStatus } from '../model/ChatMessage';
import { MediaFileInfo } from '../service/MediaService';

/**
 * 多模态消息项组件
 */
@Component
export struct MultimodalMessageItem {
  @Prop message: ChatMessage;
  private onImageClick?: (imageUrl: string) => void;
  private onAudioPlay?: (audioUrl: string) => void;
  private onFileOpen?: (fileUrl: string, fileName: string) => void;

  /**
   * 获取消息气泡背景色
   */
  private getMessageBubbleColor(): ResourceColor {
    return this.message.sender === MessageSender.USER ? '#007AFF' : '#F0F0F0';
  }

  /**
   * 获取消息文字颜色
   */
  private getMessageTextColor(): ResourceColor {
    return this.message.sender === MessageSender.USER ? Color.White : Color.Black;
  }

  /**
   * 获取消息对齐方式
   */
  private getMessageAlignment(): HorizontalAlign {
    return this.message.sender === MessageSender.USER ? HorizontalAlign.End : HorizontalAlign.Start;
  }

  /**
   * 格式化时间
   */
  private formatTime(timestamp: number): string {
    const date = new Date(timestamp);
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    return `${hours}:${minutes}`;
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
   * 格式化音频时长
   */
  private formatDuration(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  }

  /**
   * 渲染文本消息
   */
  @Builder
  private buildTextMessage() {
    Text(this.message.content)
      .fontSize(16)
      .fontColor(this.getMessageTextColor())
      .maxLines(10)
      .textOverflow({ overflow: TextOverflow.Ellipsis })
  }

  /**
   * 渲染图片消息
   */
  @Builder
  private buildImageMessage() {
    Column({ space: 8 }) {
      if (this.message.metadata?.imageUrl) {
        Image(this.message.metadata.imageUrl)
          .width(200)
          .height(150)
          .objectFit(ImageFit.Cover)
          .borderRadius(8)
          .onClick(() => {
            this.onImageClick?.(this.message.metadata.imageUrl);
          })
      } else {
        // 占位符
        Column() {
          Image($r('app.media.ic_image'))
            .width(48)
            .height(48)
            .fillColor(Color.Gray)
          
          Text('图片加载中...')
            .fontSize(12)
            .fontColor(Color.Gray)
            .margin({ top: 8 })
        }
        .width(200)
        .height(150)
        .justifyContent(FlexAlign.Center)
        .backgroundColor('#F5F5F5')
        .borderRadius(8)
      }

      if (this.message.content && this.message.content !== '[图片]') {
        Text(this.message.content)
          .fontSize(14)
          .fontColor(this.getMessageTextColor())
          .maxLines(3)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
      }
    }
  }

  /**
   * 渲染语音消息
   */
  @Builder
  private buildVoiceMessage() {
    Row({ space: 12 }) {
      // 播放按钮
      Button() {
        Image($r('app.media.ic_play'))
          .width(20)
          .height(20)
          .fillColor(this.getMessageTextColor())
      }
      .width(40)
      .height(40)
      .backgroundColor(Color.Transparent)
      .onClick(() => {
        if (this.message.metadata?.audioUrl) {
          this.onAudioPlay?.(this.message.metadata.audioUrl);
        }
      })

      Column({ space: 4 }) {
        Text('语音消息')
          .fontSize(14)
          .fontColor(this.getMessageTextColor())
        
        if (this.message.metadata?.duration) {
          Text(this.formatDuration(this.message.metadata.duration))
            .fontSize(12)
            .fontColor(this.getMessageTextColor())
            .opacity(0.7)
        }
      }
      .alignItems(HorizontalAlign.Start)
      .layoutWeight(1)

      // 音频波形图标
      Image($r('app.media.ic_waveform'))
        .width(60)
        .height(20)
        .fillColor(this.getMessageTextColor())
        .opacity(0.5)
    }
    .width(180)
    .padding(12)
  }

  /**
   * 渲染文件消息
   */
  @Builder
  private buildFileMessage() {
    Row({ space: 12 }) {
      // 文件图标
      Image($r('app.media.ic_file'))
        .width(32)
        .height(32)
        .fillColor(this.getMessageTextColor())

      Column({ space: 4 }) {
        Text(this.message.metadata?.mediaInfo?.fileName || '未知文件')
          .fontSize(14)
          .fontColor(this.getMessageTextColor())
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
        
        if (this.message.metadata?.mediaInfo?.fileSize) {
          Text(this.formatFileSize(this.message.metadata.mediaInfo.fileSize))
            .fontSize(12)
            .fontColor(this.getMessageTextColor())
            .opacity(0.7)
        }
      }
      .alignItems(HorizontalAlign.Start)
      .layoutWeight(1)
    }
    .width('100%')
    .padding(12)
    .onClick(() => {
      if (this.message.metadata?.mediaInfo?.filePath) {
        this.onFileOpen?.(
          this.message.metadata.mediaInfo.filePath,
          this.message.metadata.mediaInfo.fileName || '文件'
        );
      }
    })
  }

  /**
   * 渲染消息状态
   */
  @Builder
  private buildMessageStatus() {
    if (this.message.sender === MessageSender.USER) {
      Row({ space: 4 }) {
        Text(this.formatTime(this.message.timestamp))
          .fontSize(10)
          .fontColor(Color.Gray)

        if (this.message.status === MessageStatus.SENDING) {
          LoadingProgress()
            .width(12)
            .height(12)
            .color(Color.Gray)
        } else if (this.message.status === MessageStatus.SENT) {
          Image($r('app.media.ic_check'))
            .width(12)
            .height(12)
            .fillColor(Color.Green)
        } else if (this.message.status === MessageStatus.FAILED) {
          Image($r('app.media.ic_error'))
            .width(12)
            .height(12)
            .fillColor(Color.Red)
        }
      }
    } else {
      Text(this.formatTime(this.message.timestamp))
        .fontSize(10)
        .fontColor(Color.Gray)
    }
  }

  build() {
    Column({ space: 8 }) {
      Row() {
        Column({ space: 8 }) {
          // 消息气泡
          Column() {
            if (this.message.messageType === MessageType.TEXT) {
              this.buildTextMessage()
            } else if (this.message.messageType === MessageType.IMAGE) {
              this.buildImageMessage()
            } else if (this.message.messageType === MessageType.VOICE) {
              this.buildVoiceMessage()
            } else if (this.message.messageType === MessageType.FILE) {
              this.buildFileMessage()
            }
          }
          .padding(12)
          .backgroundColor(this.getMessageBubbleColor())
          .borderRadius(12)
          .constraintSize({ maxWidth: '70%' })

          // 消息状态
          this.buildMessageStatus()
        }
        .alignItems(this.message.sender === MessageSender.USER ? HorizontalAlign.End : HorizontalAlign.Start)
      }
      .width('100%')
      .justifyContent(this.message.sender === MessageSender.USER ? FlexAlign.End : FlexAlign.Start)
    }
    .width('100%')
    .padding({ left: 16, right: 16, top: 4, bottom: 4 })
  }
}