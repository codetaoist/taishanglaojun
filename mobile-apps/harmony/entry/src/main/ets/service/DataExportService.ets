import fs from '@ohos.file.fs';
import { DataService } from './DataService';
import { ChatMessage, Conversation } from '../model/ChatMessage';
import { TrajectoryData, LocationPoint } from '../model/LocationData';

/**
 * 导出格式枚举
 */
export enum ExportFormat {
  JSON = 'json',
  CSV = 'csv',
  TXT = 'txt'
}

/**
 * 导出类型枚举
 */
export enum ExportType {
  CHAT_MESSAGES = 'chat_messages',
  CONVERSATIONS = 'conversations',
  TRAJECTORIES = 'trajectories',
  LOCATION_POINTS = 'location_points',
  ALL_DATA = 'all_data'
}

/**
 * 导出选项接口
 */
export interface ExportOptions {
  type: ExportType;
  format: ExportFormat;
  dateRange?: {
    startDate: number;
    endDate: number;
  };
  conversationIds?: string[];
  trajectoryIds?: string[];
  includeMetadata?: boolean;
  compressOutput?: boolean;
}

/**
 * 导出结果接口
 */
export interface ExportResult {
  success: boolean;
  filePath?: string;
  fileName?: string;
  fileSize?: number;
  recordCount?: number;
  error?: string;
}

/**
 * 导出进度回调接口
 */
export interface ExportProgressCallback {
  onProgress: (progress: number, message: string) => void;
  onComplete: (result: ExportResult) => void;
  onError: (error: string) => void;
}

/**
 * 数据导出服务
 */
export class DataExportService {
  private static instance: DataExportService;
  private dataService: DataService;
  private exportDirectory: string;

  private constructor() {
    this.dataService = DataService.getInstance();
    this.exportDirectory = '';
  }

  /**
   * 获取单例实例
   */
  static getInstance(): DataExportService {
    if (!DataExportService.instance) {
      DataExportService.instance = new DataExportService();
    }
    return DataExportService.instance;
  }

  /**
   * 初始化导出服务
   */
  async initialize(): Promise<void> {
    try {
      // 获取应用文档目录
      const context = getContext();
      this.exportDirectory = context.filesDir + '/exports';
      
      // 确保导出目录存在
      await this.ensureDirectoryExists(this.exportDirectory);
      
      console.log('DataExportService initialized, export directory:', this.exportDirectory);
    } catch (error) {
      console.error('Failed to initialize DataExportService:', error);
      throw error;
    }
  }

  /**
   * 导出数据
   */
  async exportData(
    options: ExportOptions, 
    progressCallback?: ExportProgressCallback
  ): Promise<ExportResult> {
    try {
      progressCallback?.onProgress(0, '开始导出数据...');

      let data: any;
      let recordCount = 0;

      switch (options.type) {
        case ExportType.CHAT_MESSAGES:
          data = await this.exportChatMessages(options, progressCallback);
          recordCount = Array.isArray(data) ? data.length : 0;
          break;
        case ExportType.CONVERSATIONS:
          data = await this.exportConversations(options, progressCallback);
          recordCount = Array.isArray(data) ? data.length : 0;
          break;
        case ExportType.TRAJECTORIES:
          data = await this.exportTrajectories(options, progressCallback);
          recordCount = Array.isArray(data) ? data.length : 0;
          break;
        case ExportType.LOCATION_POINTS:
          data = await this.exportLocationPoints(options, progressCallback);
          recordCount = Array.isArray(data) ? data.length : 0;
          break;
        case ExportType.ALL_DATA:
          data = await this.exportAllData(options, progressCallback);
          recordCount = this.countAllRecords(data);
          break;
        default:
          throw new Error(`Unsupported export type: ${options.type}`);
      }

      progressCallback?.onProgress(80, '格式化导出数据...');

      // 格式化数据
      const formattedData = this.formatData(data, options.format);

      progressCallback?.onProgress(90, '保存导出文件...');

      // 保存到文件
      const fileName = this.generateFileName(options);
      const filePath = `${this.exportDirectory}/${fileName}`;
      
      await this.saveToFile(filePath, formattedData);

      // 获取文件大小
      const fileSize = await this.getFileSize(filePath);

      progressCallback?.onProgress(100, '导出完成');

      const result: ExportResult = {
        success: true,
        filePath,
        fileName,
        fileSize,
        recordCount
      };

      progressCallback?.onComplete(result);
      return result;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '导出失败';
      console.error('Export failed:', error);
      
      const result: ExportResult = {
        success: false,
        error: errorMessage
      };

      progressCallback?.onError(errorMessage);
      return result;
    }
  }

  /**
   * 导出聊天消息
   */
  private async exportChatMessages(
    options: ExportOptions, 
    progressCallback?: ExportProgressCallback
  ): Promise<ChatMessage[]> {
    progressCallback?.onProgress(20, '获取聊天消息...');

    const messages = await this.dataService.getAllChatMessages({
      startDate: options.dateRange?.startDate,
      endDate: options.dateRange?.endDate,
      conversationIds: options.conversationIds
    });

    progressCallback?.onProgress(60, `获取到 ${messages.length} 条消息`);

    if (!options.includeMetadata) {
      // 移除元数据以减少文件大小
      return messages.map(msg => ({
        ...msg,
        metadata: undefined
      }));
    }

    return messages;
  }

  /**
   * 导出对话列表
   */
  private async exportConversations(
    options: ExportOptions, 
    progressCallback?: ExportProgressCallback
  ): Promise<Conversation[]> {
    progressCallback?.onProgress(20, '获取对话列表...');

    const conversations = await this.dataService.getConversations();

    progressCallback?.onProgress(60, `获取到 ${conversations.length} 个对话`);

    // 根据日期范围过滤
    let filteredConversations = conversations;
    if (options.dateRange) {
      filteredConversations = conversations.filter(conv => 
        conv.createdAt >= options.dateRange!.startDate && 
        conv.createdAt <= options.dateRange!.endDate
      );
    }

    return filteredConversations;
  }

  /**
   * 导出轨迹数据
   */
  private async exportTrajectories(
    options: ExportOptions, 
    progressCallback?: ExportProgressCallback
  ): Promise<TrajectoryData[]> {
    progressCallback?.onProgress(20, '获取轨迹数据...');

    const trajectories = await this.dataService.getTrajectories({
      startDate: options.dateRange?.startDate,
      endDate: options.dateRange?.endDate,
      trajectoryIds: options.trajectoryIds
    });

    progressCallback?.onProgress(60, `获取到 ${trajectories.length} 条轨迹`);

    return trajectories;
  }

  /**
   * 导出位置点数据
   */
  private async exportLocationPoints(
    options: ExportOptions, 
    progressCallback?: ExportProgressCallback
  ): Promise<LocationPoint[]> {
    progressCallback?.onProgress(20, '获取位置点数据...');

    const locationPoints = await this.dataService.getLocationPoints({
      startDate: options.dateRange?.startDate,
      endDate: options.dateRange?.endDate,
      trajectoryIds: options.trajectoryIds
    });

    progressCallback?.onProgress(60, `获取到 ${locationPoints.length} 个位置点`);

    return locationPoints;
  }

  /**
   * 导出所有数据
   */
  private async exportAllData(
    options: ExportOptions, 
    progressCallback?: ExportProgressCallback
  ): Promise<any> {
    const allData: any = {};

    progressCallback?.onProgress(10, '导出对话数据...');
    allData.conversations = await this.exportConversations(options, progressCallback);

    progressCallback?.onProgress(30, '导出聊天消息...');
    allData.chatMessages = await this.exportChatMessages(options, progressCallback);

    progressCallback?.onProgress(50, '导出轨迹数据...');
    allData.trajectories = await this.exportTrajectories(options, progressCallback);

    progressCallback?.onProgress(70, '导出位置点数据...');
    allData.locationPoints = await this.exportLocationPoints(options, progressCallback);

    // 添加导出元信息
    allData.exportInfo = {
      exportDate: Date.now(),
      exportType: options.type,
      format: options.format,
      dateRange: options.dateRange,
      version: '1.0.0'
    };

    return allData;
  }

  /**
   * 格式化数据
   */
  private formatData(data: any, format: ExportFormat): string {
    switch (format) {
      case ExportFormat.JSON:
        return JSON.stringify(data, null, 2);
      case ExportFormat.CSV:
        return this.convertToCSV(data);
      case ExportFormat.TXT:
        return this.convertToText(data);
      default:
        throw new Error(`Unsupported format: ${format}`);
    }
  }

  /**
   * 转换为CSV格式
   */
  private convertToCSV(data: any): string {
    if (Array.isArray(data) && data.length > 0) {
      const headers = Object.keys(data[0]);
      const csvHeaders = headers.join(',');
      const csvRows = data.map(item => 
        headers.map(header => {
          const value = item[header];
          if (typeof value === 'object') {
            return `"${JSON.stringify(value).replace(/"/g, '""')}"`;
          }
          return `"${String(value).replace(/"/g, '""')}"`;
        }).join(',')
      );
      return [csvHeaders, ...csvRows].join('\n');
    }
    return '';
  }

  /**
   * 转换为文本格式
   */
  private convertToText(data: any): string {
    if (typeof data === 'object') {
      return this.objectToText(data, 0);
    }
    return String(data);
  }

  /**
   * 对象转文本（递归）
   */
  private objectToText(obj: any, indent: number): string {
    const spaces = '  '.repeat(indent);
    let result = '';

    if (Array.isArray(obj)) {
      obj.forEach((item, index) => {
        result += `${spaces}[${index}]\n`;
        result += this.objectToText(item, indent + 1);
        result += '\n';
      });
    } else if (typeof obj === 'object' && obj !== null) {
      Object.entries(obj).forEach(([key, value]) => {
        result += `${spaces}${key}: `;
        if (typeof value === 'object') {
          result += '\n' + this.objectToText(value, indent + 1);
        } else {
          result += String(value) + '\n';
        }
      });
    } else {
      result += `${spaces}${String(obj)}\n`;
    }

    return result;
  }

  /**
   * 生成文件名
   */
  private generateFileName(options: ExportOptions): string {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const typeStr = options.type.replace(/_/g, '-');
    return `${typeStr}-export-${timestamp}.${options.format}`;
  }

  /**
   * 保存到文件
   */
  private async saveToFile(filePath: string, content: string): Promise<void> {
    try {
      const file = fs.openSync(filePath, fs.OpenMode.CREATE | fs.OpenMode.WRITE_ONLY);
      fs.writeSync(file.fd, content);
      fs.closeSync(file);
    } catch (error) {
      console.error('Failed to save file:', error);
      throw error;
    }
  }

  /**
   * 获取文件大小
   */
  private async getFileSize(filePath: string): Promise<number> {
    try {
      const stat = fs.statSync(filePath);
      return stat.size;
    } catch (error) {
      console.error('Failed to get file size:', error);
      return 0;
    }
  }

  /**
   * 确保目录存在
   */
  private async ensureDirectoryExists(dirPath: string): Promise<void> {
    try {
      const stat = fs.statSync(dirPath);
      if (!stat.isDirectory()) {
        throw new Error(`Path exists but is not a directory: ${dirPath}`);
      }
    } catch (error) {
      // 目录不存在，创建它
      fs.mkdirSync(dirPath, true);
    }
  }

  /**
   * 计算所有记录数量
   */
  private countAllRecords(data: any): number {
    let count = 0;
    if (data.conversations) count += data.conversations.length;
    if (data.chatMessages) count += data.chatMessages.length;
    if (data.trajectories) count += data.trajectories.length;
    if (data.locationPoints) count += data.locationPoints.length;
    return count;
  }

  /**
   * 获取导出历史
   */
  async getExportHistory(): Promise<string[]> {
    try {
      const files = fs.listFileSync(this.exportDirectory);
      return files.filter(file => 
        file.endsWith('.json') || 
        file.endsWith('.csv') || 
        file.endsWith('.txt')
      ).sort().reverse(); // 按时间倒序
    } catch (error) {
      console.error('Failed to get export history:', error);
      return [];
    }
  }

  /**
   * 删除导出文件
   */
  async deleteExportFile(fileName: string): Promise<boolean> {
    try {
      const filePath = `${this.exportDirectory}/${fileName}`;
      fs.unlinkSync(filePath);
      return true;
    } catch (error) {
      console.error('Failed to delete export file:', error);
      return false;
    }
  }

  /**
   * 清理过期的导出文件（超过30天）
   */
  async cleanupOldExports(): Promise<number> {
    try {
      const files = await this.getExportHistory();
      const thirtyDaysAgo = Date.now() - (30 * 24 * 60 * 60 * 1000);
      let deletedCount = 0;

      for (const fileName of files) {
        const filePath = `${this.exportDirectory}/${fileName}`;
        try {
          const stat = fs.statSync(filePath);
          if (stat.mtime < thirtyDaysAgo) {
            fs.unlinkSync(filePath);
            deletedCount++;
          }
        } catch (error) {
          console.error(`Failed to process file ${fileName}:`, error);
        }
      }

      return deletedCount;
    } catch (error) {
      console.error('Failed to cleanup old exports:', error);
      return 0;
    }
  }
}