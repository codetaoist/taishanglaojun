/**
 * 媒体服务
 * 处理图片、音频等多媒体文件的保存和管理
 */

import fs from '@ohos.file.fs';
import { BusinessError } from '@ohos.base';
import hilog from '@ohos.hilog';
import { Context } from '@ohos.abilityAccessCtrl';
import photoAccessHelper from '@ohos.file.photoAccessHelper';
import camera from '@ohos.multimedia.camera';
import media from '@ohos.multimedia.media';

// 媒体类型枚举
export enum MediaType {
  IMAGE = 'image',
  AUDIO = 'audio',
  VIDEO = 'video'
}

// 媒体文件信息接口
export interface MediaFileInfo {
  id: string;
  type: MediaType;
  fileName: string;
  filePath: string;
  fileSize: number;
  mimeType: string;
  duration?: number; // 音频/视频时长（毫秒）
  width?: number;    // 图片/视频宽度
  height?: number;   // 图片/视频高度
  createdAt: number;
}

// 媒体保存选项
export interface MediaSaveOptions {
  quality?: number;     // 图片质量 (0-100)
  maxWidth?: number;    // 最大宽度
  maxHeight?: number;   // 最大高度
  compress?: boolean;   // 是否压缩
}

/**
 * 媒体服务类
 */
export class MediaService {
  private static instance: MediaService;
  private context?: Context;
  private mediaDir: string = '';
  private isInitialized: boolean = false;

  private constructor() {}

  /**
   * 获取单例实例
   */
  static getInstance(): MediaService {
    if (!MediaService.instance) {
      MediaService.instance = new MediaService();
    }
    return MediaService.instance;
  }

  /**
   * 初始化媒体服务
   */
  async initialize(context: Context): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    try {
      this.context = context;
      this.mediaDir = context.filesDir + '/media';
      
      // 创建媒体目录
      await this.ensureDirectoryExists(this.mediaDir);
      await this.ensureDirectoryExists(this.mediaDir + '/images');
      await this.ensureDirectoryExists(this.mediaDir + '/audio');
      await this.ensureDirectoryExists(this.mediaDir + '/video');
      
      this.isInitialized = true;
      hilog.info(0x0000, 'MediaService', 'Media service initialized successfully');
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Media service initialization error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 确保目录存在
   */
  private async ensureDirectoryExists(dirPath: string): Promise<void> {
    try {
      const stat = await fs.stat(dirPath);
      if (!stat.isDirectory()) {
        throw new Error(`Path exists but is not a directory: ${dirPath}`);
      }
    } catch (error) {
      // 目录不存在，创建它
      await fs.mkdir(dirPath, true);
    }
  }

  /**
   * 生成唯一文件名
   */
  private generateFileName(type: MediaType, extension: string): string {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(2, 8);
    return `${type}_${timestamp}_${random}.${extension}`;
  }

  /**
   * 获取文件MIME类型
   */
  private getMimeType(extension: string): string {
    const mimeTypes: Record<string, string> = {
      'jpg': 'image/jpeg',
      'jpeg': 'image/jpeg',
      'png': 'image/png',
      'gif': 'image/gif',
      'webp': 'image/webp',
      'mp3': 'audio/mpeg',
      'wav': 'audio/wav',
      'aac': 'audio/aac',
      'mp4': 'video/mp4',
      'mov': 'video/quicktime',
      'avi': 'video/x-msvideo'
    };
    return mimeTypes[extension.toLowerCase()] || 'application/octet-stream';
  }

  /**
   * 保存图片文件
   */
  async saveImage(imageData: ArrayBuffer, options?: MediaSaveOptions): Promise<MediaFileInfo> {
    if (!this.isInitialized) {
      throw new Error('MediaService not initialized');
    }

    try {
      const fileName = this.generateFileName(MediaType.IMAGE, 'jpg');
      const filePath = `${this.mediaDir}/images/${fileName}`;
      
      // 保存文件
      const file = await fs.open(filePath, fs.OpenMode.CREATE | fs.OpenMode.WRITE_ONLY);
      await fs.write(file.fd, imageData);
      await fs.close(file.fd);
      
      // 获取文件信息
      const stat = await fs.stat(filePath);
      
      const mediaInfo: MediaFileInfo = {
        id: this.generateMediaId(),
        type: MediaType.IMAGE,
        fileName: fileName,
        filePath: filePath,
        fileSize: stat.size,
        mimeType: this.getMimeType('jpg'),
        createdAt: Date.now()
      };

      hilog.info(0x0000, 'MediaService', 'Image saved successfully: %{public}s', fileName);
      return mediaInfo;
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Save image error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 保存音频文件
   */
  async saveAudio(audioData: ArrayBuffer, extension: string = 'mp3'): Promise<MediaFileInfo> {
    if (!this.isInitialized) {
      throw new Error('MediaService not initialized');
    }

    try {
      const fileName = this.generateFileName(MediaType.AUDIO, extension);
      const filePath = `${this.mediaDir}/audio/${fileName}`;
      
      // 保存文件
      const file = await fs.open(filePath, fs.OpenMode.CREATE | fs.OpenMode.WRITE_ONLY);
      await fs.write(file.fd, audioData);
      await fs.close(file.fd);
      
      // 获取文件信息
      const stat = await fs.stat(filePath);
      
      const mediaInfo: MediaFileInfo = {
        id: this.generateMediaId(),
        type: MediaType.AUDIO,
        fileName: fileName,
        filePath: filePath,
        fileSize: stat.size,
        mimeType: this.getMimeType(extension),
        createdAt: Date.now()
      };

      hilog.info(0x0000, 'MediaService', 'Audio saved successfully: %{public}s', fileName);
      return mediaInfo;
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Save audio error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 保存视频文件
   */
  async saveVideo(videoData: ArrayBuffer, extension: string = 'mp4'): Promise<MediaFileInfo> {
    if (!this.isInitialized) {
      throw new Error('MediaService not initialized');
    }

    try {
      const fileName = this.generateFileName(MediaType.VIDEO, extension);
      const filePath = `${this.mediaDir}/video/${fileName}`;
      
      // 保存文件
      const file = await fs.open(filePath, fs.OpenMode.CREATE | fs.OpenMode.WRITE_ONLY);
      await fs.write(file.fd, videoData);
      await fs.close(file.fd);
      
      // 获取文件信息
      const stat = await fs.stat(filePath);
      
      const mediaInfo: MediaFileInfo = {
        id: this.generateMediaId(),
        type: MediaType.VIDEO,
        fileName: fileName,
        filePath: filePath,
        fileSize: stat.size,
        mimeType: this.getMimeType(extension),
        createdAt: Date.now()
      };

      hilog.info(0x0000, 'MediaService', 'Video saved successfully: %{public}s', fileName);
      return mediaInfo;
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Save video error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 删除媒体文件
   */
  async deleteMedia(filePath: string): Promise<void> {
    try {
      await fs.unlink(filePath);
      hilog.info(0x0000, 'MediaService', 'Media file deleted: %{public}s', filePath);
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Delete media error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取媒体文件信息
   */
  async getMediaInfo(filePath: string): Promise<MediaFileInfo | null> {
    try {
      const stat = await fs.stat(filePath);
      if (!stat.isFile()) {
        return null;
      }

      const fileName = filePath.split('/').pop() || '';
      const extension = fileName.split('.').pop() || '';
      
      let type: MediaType;
      if (['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(extension.toLowerCase())) {
        type = MediaType.IMAGE;
      } else if (['mp3', 'wav', 'aac'].includes(extension.toLowerCase())) {
        type = MediaType.AUDIO;
      } else if (['mp4', 'mov', 'avi'].includes(extension.toLowerCase())) {
        type = MediaType.VIDEO;
      } else {
        return null;
      }

      return {
        id: this.generateMediaId(),
        type: type,
        fileName: fileName,
        filePath: filePath,
        fileSize: stat.size,
        mimeType: this.getMimeType(extension),
        createdAt: stat.mtime
      };
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Get media info error: %{public}s', JSON.stringify(error));
      return null;
    }
  }

  /**
   * 清理过期的媒体文件
   */
  async cleanupExpiredMedia(maxAgeMs: number = 7 * 24 * 60 * 60 * 1000): Promise<void> {
    if (!this.isInitialized) {
      return;
    }

    try {
      const now = Date.now();
      const directories = ['images', 'audio', 'video'];
      
      for (const dir of directories) {
        const dirPath = `${this.mediaDir}/${dir}`;
        const files = await fs.listFile(dirPath);
        
        for (const fileName of files) {
          const filePath = `${dirPath}/${fileName}`;
          const stat = await fs.stat(filePath);
          
          if (now - stat.mtime > maxAgeMs) {
            await fs.unlink(filePath);
            hilog.info(0x0000, 'MediaService', 'Expired media file deleted: %{public}s', fileName);
          }
        }
      }
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Cleanup expired media error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 获取媒体目录大小
   */
  async getMediaDirectorySize(): Promise<number> {
    if (!this.isInitialized) {
      return 0;
    }

    try {
      let totalSize = 0;
      const directories = ['images', 'audio', 'video'];
      
      for (const dir of directories) {
        const dirPath = `${this.mediaDir}/${dir}`;
        const files = await fs.listFile(dirPath);
        
        for (const fileName of files) {
          const filePath = `${dirPath}/${fileName}`;
          const stat = await fs.stat(filePath);
          totalSize += stat.size;
        }
      }
      
      return totalSize;
    } catch (error) {
      hilog.error(0x0000, 'MediaService', 'Get media directory size error: %{public}s', JSON.stringify(error));
      return 0;
    }
  }

  /**
   * 生成媒体ID
   */
  private generateMediaId(): string {
    return `media_${Date.now()}_${Math.random().toString(36).substring(2, 8)}`;
  }

  /**
   * 获取媒体目录路径
   */
  getMediaDirectory(): string {
    return this.mediaDir;
  }

  /**
   * 检查服务是否已初始化
   */
  isServiceInitialized(): boolean {
    return this.isInitialized;
  }
}