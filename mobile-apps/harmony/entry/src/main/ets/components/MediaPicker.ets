import { MediaType } from '../service/MediaService';

/**
 * 多媒体选择器组件
 */
@Component
export struct MediaPicker {
  @State private isVisible: boolean = false;
  private onImageSelected?: (imageData: ArrayBuffer) => void;
  private onAudioSelected?: (audioData: ArrayBuffer, duration?: number) => void;
  private onFileSelected?: (fileData: ArrayBuffer, fileName: string, mimeType: string) => void;
  private onCancel?: () => void;

  /**
   * 显示选择器
   */
  show(callbacks: {
    onImageSelected?: (imageData: ArrayBuffer) => void;
    onAudioSelected?: (audioData: ArrayBuffer, duration?: number) => void;
    onFileSelected?: (fileData: ArrayBuffer, fileName: string, mimeType: string) => void;
    onCancel?: () => void;
  }) {
    this.onImageSelected = callbacks.onImageSelected;
    this.onAudioSelected = callbacks.onAudioSelected;
    this.onFileSelected = callbacks.onFileSelected;
    this.onCancel = callbacks.onCancel;
    this.isVisible = true;
  }

  /**
   * 隐藏选择器
   */
  hide() {
    this.isVisible = false;
    this.onImageSelected = undefined;
    this.onAudioSelected = undefined;
    this.onFileSelected = undefined;
    this.onCancel = undefined;
  }

  /**
   * 选择图片
   */
  private async selectImage() {
    try {
      // 这里应该调用系统的图片选择器
      // 暂时使用模拟数据
      const mockImageData = new ArrayBuffer(1024);
      this.onImageSelected?.(mockImageData);
      this.hide();
    } catch (error) {
      console.error('Failed to select image:', error);
    }
  }

  /**
   * 选择音频
   */
  private async selectAudio() {
    try {
      // 这里应该调用系统的音频选择器或录音功能
      // 暂时使用模拟数据
      const mockAudioData = new ArrayBuffer(2048);
      const mockDuration = 30000; // 30秒
      this.onAudioSelected?.(mockAudioData, mockDuration);
      this.hide();
    } catch (error) {
      console.error('Failed to select audio:', error);
    }
  }

  /**
   * 选择文件
   */
  private async selectFile() {
    try {
      // 这里应该调用系统的文件选择器
      // 暂时使用模拟数据
      const mockFileData = new ArrayBuffer(4096);
      const mockFileName = 'document.pdf';
      const mockMimeType = 'application/pdf';
      this.onFileSelected?.(mockFileData, mockFileName, mockMimeType);
      this.hide();
    } catch (error) {
      console.error('Failed to select file:', error);
    }
  }

  /**
   * 取消选择
   */
  private cancel() {
    this.onCancel?.();
    this.hide();
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
        .onClick(() => this.cancel())

      // 选择器面板
      Column({ space: 16 }) {
        Text('选择媒体类型')
          .fontSize(18)
          .fontWeight(FontWeight.Bold)
          .fontColor(Color.Black)

        Column({ space: 12 }) {
          // 图片选项
          Row({ space: 12 }) {
            Image($r('app.media.ic_image'))
              .width(24)
              .height(24)
              .fillColor(Color.Blue)
            
            Text('选择图片')
              .fontSize(16)
              .fontColor(Color.Black)
              .layoutWeight(1)
          }
          .width('100%')
          .height(48)
          .padding({ left: 16, right: 16 })
          .backgroundColor(Color.White)
          .borderRadius(8)
          .onClick(() => this.selectImage())

          // 音频选项
          Row({ space: 12 }) {
            Image($r('app.media.ic_audio'))
              .width(24)
              .height(24)
              .fillColor(Color.Green)
            
            Text('录制语音')
              .fontSize(16)
              .fontColor(Color.Black)
              .layoutWeight(1)
          }
          .width('100%')
          .height(48)
          .padding({ left: 16, right: 16 })
          .backgroundColor(Color.White)
          .borderRadius(8)
          .onClick(() => this.selectAudio())

          // 文件选项
          Row({ space: 12 }) {
            Image($r('app.media.ic_file'))
              .width(24)
              .height(24)
              .fillColor(Color.Orange)
            
            Text('选择文件')
              .fontSize(16)
              .fontColor(Color.Black)
              .layoutWeight(1)
          }
          .width('100%')
          .height(48)
          .padding({ left: 16, right: 16 })
          .backgroundColor(Color.White)
          .borderRadius(8)
          .onClick(() => this.selectFile())
        }

        // 取消按钮
        Button('取消')
          .width('100%')
          .height(44)
          .backgroundColor(Color.Gray)
          .fontColor(Color.White)
          .borderRadius(8)
          .onClick(() => this.cancel())
      }
      .width('80%')
      .padding(20)
      .backgroundColor(Color.White)
      .borderRadius(12)
      .shadow({
        radius: 10,
        color: Color.Black,
        offsetX: 0,
        offsetY: 2
      })
    }
    .width('100%')
    .height('100%')
    .position({ x: 0, y: 0 })
    .zIndex(1000)
  }
}