/**
 * 数据同步服务
 * 负责在后台同步位置数据、轨迹数据和聊天数据
 */

import { backgroundTaskManager } from '@kit.BackgroundTasksKit';
import { connection } from '@kit.NetworkKit';
import { LocationPoint } from '../data/LocationPoint';
import { Trajectory } from '../data/Trajectory';
import { SecureApiClient, ApiResponse } from '../network/SecureApiClient';
import { DataService } from './DataService';

/**
 * 同步状态枚举
 */
export enum SyncState {
  IDLE = 'idle',           // 空闲
  RUNNING = 'running',     // 运行中
  SUCCESS = 'success',     // 成功
  ERROR = 'error',         // 错误
  WAITING_NETWORK = 'waiting_network' // 等待网络
}

/**
 * 同步统计数据
 */
export interface SyncStats {
  lastSyncTime: number;
  totalSyncCount: number;
  locationSyncCount: number;
  trajectorySyncCount: number;
  chatSyncCount: number;
  errorCount: number;
  lastSyncDuration: number;
}

/**
 * 数据同步服务类
 */
export class DataSyncService {
  private static instance: DataSyncService;
  
  // 服务状态
  private syncState: SyncState = SyncState.IDLE;
  private syncStats: SyncStats = {
    lastSyncTime: 0,
    totalSyncCount: 0,
    locationSyncCount: 0,
    trajectorySyncCount: 0,
    chatSyncCount: 0,
    errorCount: 0,
    lastSyncDuration: 0
  };
  
  // 依赖服务
  private apiClient: SecureApiClient;
  private dataService: DataService;
  
  // 同步配置
  private readonly SYNC_INTERVAL = 30000; // 30秒
  private readonly RETRY_DELAY = 60000; // 重试延迟60秒
  private readonly MAX_RETRY_COUNT = 3;
  
  // 定时器和任务ID
  private syncTimer: number | null = null;
  private backgroundTaskId: number | null = null;
  
  private constructor() {
    this.apiClient = SecureApiClient.getInstance();
    this.dataService = DataService.getInstance();
    this.setupNetworkMonitoring();
  }
  
  /**
   * 获取单例实例
   */
  public static getInstance(): DataSyncService {
    if (!DataSyncService.instance) {
      DataSyncService.instance = new DataSyncService();
    }
    return DataSyncService.instance;
  }
  
  /**
   * 设置网络监听
   */
  private setupNetworkMonitoring(): void {
    connection.getDefaultNet().then((netHandle) => {
      connection.getNetCapabilities(netHandle).then((netCapabilities) => {
        if (netCapabilities.hasCapability(connection.NetCap.NET_CAPABILITY_INTERNET)) {
          if (this.syncState === SyncState.WAITING_NETWORK) {
            this.performSync();
          }
        }
      }).catch((error) => {
        console.error('获取网络能力失败:', error);
      });
    }).catch((error) => {
      console.error('获取默认网络失败:', error);
    });
  }
  
  /**
   * 开始周期性同步
   */
  public async startPeriodicSync(): Promise<void> {
    if (this.syncTimer !== null) {
      console.info('🔄 同步任务已在运行');
      return;
    }
    
    console.info('🚀 开始周期性数据同步');
    this.syncState = SyncState.RUNNING;
    
    // 申请后台任务
    try {
      this.backgroundTaskId = await backgroundTaskManager.requestSuspendDelay(
        backgroundTaskManager.DelaySuspendInfo.NORMAL,
        () => {
          console.info('后台任务即将到期');
          this.stopPeriodicSync();
        }
      );
    } catch (error) {
      console.error('申请后台任务失败:', error);
    }
    
    // 立即执行一次同步
    await this.performSync();
    
    // 设置定时器
    this.syncTimer = setInterval(async () => {
      await this.performSync();
    }, this.SYNC_INTERVAL);
  }
  
  /**
   * 停止周期性同步
   */
  public stopPeriodicSync(): void {
    console.info('🛑 停止周期性数据同步');
    
    if (this.syncTimer !== null) {
      clearInterval(this.syncTimer);
      this.syncTimer = null;
    }
    
    if (this.backgroundTaskId !== null) {
      backgroundTaskManager.cancelSuspendDelay(this.backgroundTaskId);
      this.backgroundTaskId = null;
    }
    
    this.syncState = SyncState.IDLE;
  }
  
  /**
   * 执行强制同步
   */
  public async forceSync(): Promise<void> {
    console.info('⚡ 执行强制同步');
    await this.performSync();
  }
  
  /**
   * 执行数据同步
   */
  private async performSync(): Promise<void> {
    // 检查网络连接
    const isNetworkAvailable = await this.checkNetworkAvailability();
    if (!isNetworkAvailable) {
      console.warn('⚠️ 网络不可用，等待网络连接');
      this.syncState = SyncState.WAITING_NETWORK;
      return;
    }
    
    this.syncState = SyncState.RUNNING;
    const startTime = Date.now();
    
    let locationSyncCount = 0;
    let trajectorySyncCount = 0;
    let chatSyncCount = 0;
    let errorCount = 0;
    
    try {
      console.info('📊 开始执行数据同步');
      
      // 同步位置数据
      locationSyncCount = await this.syncLocationData();
      
      // 同步轨迹数据
      trajectorySyncCount = await this.syncTrajectoryData();
      
      // 同步聊天数据
      chatSyncCount = await this.syncChatData();
      
      this.syncState = SyncState.SUCCESS;
      
    } catch (error) {
      console.error('❌ 数据同步失败:', error);
      errorCount++;
      this.syncState = SyncState.ERROR;
    }
    
    // 更新同步统计
    const duration = Date.now() - startTime;
    this.syncStats = {
      lastSyncTime: Date.now(),
      totalSyncCount: this.syncStats.totalSyncCount + 1,
      locationSyncCount: this.syncStats.locationSyncCount + locationSyncCount,
      trajectorySyncCount: this.syncStats.trajectorySyncCount + trajectorySyncCount,
      chatSyncCount: this.syncStats.chatSyncCount + chatSyncCount,
      errorCount: this.syncStats.errorCount + errorCount,
      lastSyncDuration: duration
    };
    
    console.info(`✅ 数据同步完成 - 位置: ${locationSyncCount}, 轨迹: ${trajectorySyncCount}, 聊天: ${chatSyncCount}, 耗时: ${duration}ms`);
  }
  
  /**
   * 检查网络可用性
   */
  private async checkNetworkAvailability(): Promise<boolean> {
    try {
      const netHandle = await connection.getDefaultNet();
      const netCapabilities = await connection.getNetCapabilities(netHandle);
      return netCapabilities.hasCapability(connection.NetCap.NET_CAPABILITY_INTERNET);
    } catch (error) {
      console.error('检查网络可用性失败:', error);
      return false;
    }
  }
  
  /**
   * 同步位置数据
   */
  private async syncLocationData(): Promise<number> {
    try {
      const unsyncedPoints = await this.dataService.getUnsyncedLocationPoints();
      
      if (unsyncedPoints.length === 0) {
        console.info('📍 没有未同步的位置数据');
        return 0;
      }
      
      console.info(`📍 开始同步 ${unsyncedPoints.length} 个位置点`);
      
      const response = await this.apiClient.uploadLocationPoints(unsyncedPoints);
      
      if (response.success) {
        // 标记为已同步
        await this.dataService.markLocationPointsAsSynced(unsyncedPoints.map(point => point.id));
        console.info('✅ 位置数据同步成功');
        return unsyncedPoints.length;
      } else {
        console.error(`❌ 位置数据同步失败: ${response.message}`);
        return 0;
      }
      
    } catch (error) {
      console.error('❌ 位置数据同步异常:', error);
      return 0;
    }
  }
  
  /**
   * 同步轨迹数据
   */
  private async syncTrajectoryData(): Promise<number> {
    try {
      const unsyncedTrajectories = await this.dataService.getUnsyncedTrajectories();
      
      if (unsyncedTrajectories.length === 0) {
        console.info('🛤️ 没有未同步的轨迹数据');
        return 0;
      }
      
      console.info(`🛤️ 开始同步 ${unsyncedTrajectories.length} 条轨迹`);
      
      let syncCount = 0;
      
      for (const trajectory of unsyncedTrajectories) {
        const response = await this.apiClient.uploadTrajectory(trajectory);
        
        if (response.success) {
          // 标记为已同步
          const syncedTrajectory = { ...trajectory, synced: true };
          await this.dataService.saveTrajectory(syncedTrajectory);
          syncCount++;
        } else {
          console.error(`❌ 轨迹同步失败: ${response.message}`);
        }
      }
      
      console.info(`✅ 轨迹数据同步完成，成功: ${syncCount}/${unsyncedTrajectories.length}`);
      return syncCount;
      
    } catch (error) {
      console.error('❌ 轨迹数据同步异常:', error);
      return 0;
    }
  }
  
  /**
   * 同步聊天数据
   */
  private async syncChatData(): Promise<number> {
    try {
      let totalSyncCount = 0;
      
      // 同步对话数据
      const conversationSyncCount = await this.syncConversations();
      totalSyncCount += conversationSyncCount;
      
      // 同步消息数据
      const messageSyncCount = await this.syncMessages();
      totalSyncCount += messageSyncCount;
      
      console.info(`💬 聊天数据同步完成，对话: ${conversationSyncCount}, 消息: ${messageSyncCount}`);
      return totalSyncCount;
      
    } catch (error) {
      console.error('❌ 聊天数据同步异常:', error);
      return 0;
    }
  }
  
  /**
   * 同步对话数据
   */
  private async syncConversations(): Promise<number> {
    try {
      const unsyncedConversations = await this.dataService.getUnsyncedConversations();
      
      if (unsyncedConversations.length === 0) {
        console.info('💬 没有未同步的对话数据');
        return 0;
      }
      
      console.info(`💬 开始同步 ${unsyncedConversations.length} 个对话`);
      
      let syncCount = 0;
      
      for (const conversation of unsyncedConversations) {
        try {
          const response = await this.apiClient.uploadConversation(conversation);
          
          if (response.success) {
            // 标记为已同步
            await this.dataService.markConversationAsSynced(conversation.id);
            syncCount++;
          } else {
            console.error(`❌ 对话同步失败: ${response.message}`);
          }
        } catch (error) {
          console.error(`❌ 对话 ${conversation.id} 同步异常:`, error);
        }
      }
      
      console.info(`✅ 对话数据同步完成，成功: ${syncCount}/${unsyncedConversations.length}`);
      return syncCount;
      
    } catch (error) {
      console.error('❌ 对话数据同步异常:', error);
      return 0;
    }
  }
  
  /**
   * 同步消息数据
   */
  private async syncMessages(): Promise<number> {
    try {
      const unsyncedMessages = await this.dataService.getUnsyncedMessages();
      
      if (unsyncedMessages.length === 0) {
        console.info('💬 没有未同步的消息数据');
        return 0;
      }
      
      console.info(`💬 开始同步 ${unsyncedMessages.length} 条消息`);
      
      let syncCount = 0;
      
      for (const message of unsyncedMessages) {
        try {
          const response = await this.apiClient.uploadMessage(message);
          
          if (response.success) {
            // 标记为已同步
            await this.dataService.markMessageAsSynced(message.id);
            syncCount++;
          } else {
            console.error(`❌ 消息同步失败: ${response.message}`);
          }
        } catch (error) {
          console.error(`❌ 消息 ${message.id} 同步异常:`, error);
        }
      }
      
      console.info(`✅ 消息数据同步完成，成功: ${syncCount}/${unsyncedMessages.length}`);
      return syncCount;
      
    } catch (error) {
      console.error('❌ 消息数据同步异常:', error);
      return 0;
    }
  }
  
  /**
   * 获取同步状态
   */
  public getSyncState(): SyncState {
    return this.syncState;
  }
  
  /**
   * 获取同步统计
   */
  public getSyncStats(): SyncStats {
    return this.syncStats;
  }
  
  /**
   * 获取格式化的最后同步时间
   */
  public getFormattedLastSyncTime(): string {
    if (this.syncStats.lastSyncTime === 0) {
      return '从未同步';
    }
    
    const date = new Date(this.syncStats.lastSyncTime);
    return date.toLocaleString('zh-CN');
  }
  
  /**
   * 获取格式化的同步耗时
   */
  public getFormattedDuration(): string {
    return `${(this.syncStats.lastSyncDuration / 1000).toFixed(2)}s`;
  }
}