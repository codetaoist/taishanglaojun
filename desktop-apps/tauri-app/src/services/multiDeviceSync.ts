import { invoke } from '@tauri-apps/api/core';

// 设备类型枚举
export enum DeviceType {
  Desktop = 'Desktop',
  Mobile = 'Mobile',
  Watch = 'Watch',
  Tablet = 'Tablet',
  Web = 'Web',
}

// 同步数据类型
export enum SyncDataType {
  UserProfile = 'UserProfile',
  ChatMessage = 'ChatMessage',
  ChatSession = 'ChatSession',
  Friend = 'Friend',
  Project = 'Project',
  File = 'File',
  Settings = 'Settings',
}

// 同步操作类型
export enum SyncOperation {
  Create = 'Create',
  Update = 'Update',
  Delete = 'Delete',
  Read = 'Read',
}

// 设备信息接口
export interface DeviceInfo {
  device_id: string;
  device_type: DeviceType;
  device_name: string;
  platform: string;
  app_version: string;
  last_sync: string;
  is_online: boolean;
}

// 同步记录接口
export interface SyncRecord {
  id: string;
  user_id: string;
  device_id: string;
  data_type: SyncDataType;
  operation: SyncOperation;
  data_id: string;
  data_hash: string;
  timestamp: string;
  version: number;
  conflict_resolution?: string;
}

// 实时同步消息接口
export interface RealtimeSyncMessage {
  type: 'DeviceOnline' | 'DeviceOffline' | 'DataUpdate' | 'ChatMessage' | 'FriendStatusUpdate' | 'SyncRequest' | 'SyncResponse' | 'Heartbeat';
  data: any;
}

// 同步状态
export enum SyncStatus {
  Disconnected = 'Disconnected',
  Connecting = 'Connecting',
  Connected = 'Connected',
  Syncing = 'Syncing',
  Error = 'Error',
}

// 冲突解决策略
export enum ConflictResolution {
  LastWriteWins = 'LastWriteWins',
  FirstWriteWins = 'FirstWriteWins',
  MergeChanges = 'MergeChanges',
  UserChoice = 'UserChoice',
  DevicePriority = 'DevicePriority',
}

// 多设备同步管理器
export class MultiDeviceSyncManager {
  private static instance: MultiDeviceSyncManager;
  private websocket: WebSocket | null = null;
  private syncStatus: SyncStatus = SyncStatus.Disconnected;
  private deviceInfo: DeviceInfo | null = null;
  private lastSyncTime: Date = new Date(0);
  private syncListeners: Map<string, (data: any) => void> = new Map();
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // 1秒

  private constructor() {}

  public static getInstance(): MultiDeviceSyncManager {
    if (!MultiDeviceSyncManager.instance) {
      MultiDeviceSyncManager.instance = new MultiDeviceSyncManager();
    }
    return MultiDeviceSyncManager.instance;
  }

  // 初始化同步管理器
  public async initialize(userId: string): Promise<void> {
    try {
      // 获取设备信息
      this.deviceInfo = await this.getDeviceInfo();
      
      // 注册设备
      await invoke('register_device', { deviceInfo: this.deviceInfo });
      
      // 连接WebSocket
      await this.connectWebSocket();
      
      // 执行初始同步
      await this.performInitialSync(userId);
      
      console.log('多设备同步管理器初始化完成');
    } catch (error) {
      console.error('初始化多设备同步管理器失败:', error);
      throw error;
    }
  }

  // 连接WebSocket
  private async connectWebSocket(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.setSyncStatus(SyncStatus.Connecting);
        
        // 获取WebSocket服务器地址
        const wsUrl = 'ws://localhost:8080/sync'; // 可配置
        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
          console.log('WebSocket连接已建立');
          this.setSyncStatus(SyncStatus.Connected);
          this.reconnectAttempts = 0;
          
          // 发送设备上线消息
          this.sendMessage({
            type: 'DeviceOnline',
            data: {
              device_id: this.deviceInfo?.device_id,
              user_id: this.getCurrentUserId(),
            }
          });

          // 启动心跳
          this.startHeartbeat();
          
          resolve();
        };

        this.websocket.onmessage = (event) => {
          this.handleWebSocketMessage(event.data);
        };

        this.websocket.onclose = () => {
          console.log('WebSocket连接已关闭');
          this.setSyncStatus(SyncStatus.Disconnected);
          this.stopHeartbeat();
          
          // 尝试重连
          this.attemptReconnect();
        };

        this.websocket.onerror = (error) => {
          console.error('WebSocket错误:', error);
          this.setSyncStatus(SyncStatus.Error);
          reject(error);
        };

      } catch (error) {
        reject(error);
      }
    });
  }

  // 处理WebSocket消息
  private handleWebSocketMessage(data: string): void {
    try {
      const message: RealtimeSyncMessage = JSON.parse(data);
      
      switch (message.type) {
        case 'DataUpdate':
          this.handleDataUpdate(message.data);
          break;
        case 'ChatMessage':
          this.handleChatMessage(message.data);
          break;
        case 'FriendStatusUpdate':
          this.handleFriendStatusUpdate(message.data);
          break;
        case 'SyncResponse':
          this.handleSyncResponse(message.data);
          break;
        case 'DeviceOnline':
        case 'DeviceOffline':
          this.handleDeviceStatusChange(message.data);
          break;
        default:
          console.log('未知消息类型:', message.type);
      }

      // 通知监听器
      this.notifyListeners(message.type, message.data);
    } catch (error) {
      console.error('处理WebSocket消息失败:', error);
    }
  }

  // 发送消息
  private sendMessage(message: RealtimeSyncMessage): void {
    if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
      this.websocket.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket未连接，消息发送失败');
    }
  }

  // 执行初始同步
  private async performInitialSync(userId: string): Promise<void> {
    try {
      this.setSyncStatus(SyncStatus.Syncing);
      
      // 请求增量同步
      this.sendMessage({
        type: 'SyncRequest',
        data: {
          user_id: userId,
          device_id: this.deviceInfo?.device_id,
          last_sync: this.lastSyncTime.toISOString(),
        }
      });

      // 处理离线操作队列
      await invoke('process_offline_queue');
      
      this.lastSyncTime = new Date();
      this.setSyncStatus(SyncStatus.Connected);
      
      console.log('初始同步完成');
    } catch (error) {
      console.error('初始同步失败:', error);
      this.setSyncStatus(SyncStatus.Error);
    }
  }

  // 同步聊天消息
  public async syncChatMessage(message: any): Promise<void> {
    try {
      // 本地保存
      await invoke('save_chat_message', { message });
      
      // 创建同步记录
      await invoke('create_sync_record', {
        userId: this.getCurrentUserId(),
        deviceId: this.deviceInfo?.device_id,
        dataType: SyncDataType.ChatMessage,
        operation: SyncOperation.Create,
        dataId: message.id,
        data: JSON.stringify(message),
      });

      // 实时广播
      this.sendMessage({
        type: 'ChatMessage',
        data: message,
      });

    } catch (error) {
      console.error('同步聊天消息失败:', error);
      
      // 添加到离线队列
      await invoke('add_offline_operation', {
        userId: this.getCurrentUserId(),
        deviceId: this.deviceInfo?.device_id,
        operationType: SyncOperation.Create,
        dataType: SyncDataType.ChatMessage,
        dataId: message.id,
        dataPayload: JSON.stringify(message),
        priority: 'High',
      });
    }
  }

  // 同步好友数据
  public async syncFriend(friend: any, operation: SyncOperation): Promise<void> {
    try {
      // 本地操作
      switch (operation) {
        case SyncOperation.Create:
          await invoke('add_friend', { friend });
          break;
        case SyncOperation.Update:
          await invoke('update_friend', { friend });
          break;
        case SyncOperation.Delete:
          await invoke('delete_friend', { friendId: friend.id });
          break;
      }

      // 创建同步记录
      await invoke('create_sync_record', {
        userId: this.getCurrentUserId(),
        deviceId: this.deviceInfo?.device_id,
        dataType: SyncDataType.Friend,
        operation,
        dataId: friend.id,
        data: JSON.stringify(friend),
      });

      // 实时广播
      this.sendMessage({
        type: 'DataUpdate',
        data: {
          sync_record: {
            data_type: SyncDataType.Friend,
            operation,
            data_id: friend.id,
            timestamp: new Date().toISOString(),
          }
        }
      });

    } catch (error) {
      console.error('同步好友数据失败:', error);
      
      // 添加到离线队列
      await invoke('add_offline_operation', {
        userId: this.getCurrentUserId(),
        deviceId: this.deviceInfo?.device_id,
        operationType: operation,
        dataType: SyncDataType.Friend,
        dataId: friend.id,
        dataPayload: JSON.stringify(friend),
        priority: 'Normal',
      });
    }
  }

  // 获取在线设备列表
  public async getOnlineDevices(): Promise<DeviceInfo[]> {
    try {
      return await invoke('get_online_devices', {
        userId: this.getCurrentUserId(),
      });
    } catch (error) {
      console.error('获取在线设备失败:', error);
      return [];
    }
  }

  // 获取同步状态
  public getSyncStatus(): SyncStatus {
    return this.syncStatus;
  }

  public getStatus(): SyncStatus {
    return this.syncStatus;
  }



  public async syncIncremental(): Promise<void> {
    try {
      this.setSyncStatus(SyncStatus.Syncing);
      // 执行增量同步逻辑
      await this.manualSync();
    } catch (error) {
      console.error('Incremental sync failed:', error);
      this.setSyncStatus(SyncStatus.Error);
      throw error;
    }
  }

  public addSyncListener(type: string, callback: (data: any) => void): void {
    this.syncListeners.set(type, callback);
  }

  public removeSyncListener(type: string): void {
    this.syncListeners.delete(type);
  }

  // 别名方法，为了兼容性
  public on(type: string, callback: (data: any) => void): void {
    this.addSyncListener(type, callback);
  }

  public off(type: string): void {
    this.removeSyncListener(type);
  }

  // 手动触发同步
  public async manualSync(): Promise<void> {
    if (this.syncStatus === SyncStatus.Connected) {
      await this.performInitialSync(this.getCurrentUserId());
    } else {
      throw new Error('设备未连接，无法执行同步');
    }
  }

  // 断开连接
  public disconnect(): void {
    if (this.websocket) {
      this.websocket.close();
      this.websocket = null;
    }
    this.stopHeartbeat();
    this.setSyncStatus(SyncStatus.Disconnected);
  }

  // 私有方法
  private setSyncStatus(status: SyncStatus): void {
    this.syncStatus = status;
    this.notifyListeners('statusChange', { status });
  }

  private notifyListeners(type: string, data: any): void {
    const listener = this.syncListeners.get(type);
    if (listener) {
      listener(data);
    }
  }

  private startHeartbeat(): void {
    this.heartbeatInterval = setInterval(() => {
      this.sendMessage({
        type: 'Heartbeat',
        data: {
          device_id: this.deviceInfo?.device_id,
          timestamp: new Date().toISOString(),
        }
      });
    }, 30000); // 30秒心跳
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      
      console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})，${delay}ms后重试`);
      
      setTimeout(() => {
        this.connectWebSocket().catch(error => {
          console.error('重连失败:', error);
        });
      }, delay);
    } else {
      console.error('达到最大重连次数，停止重连');
      this.setSyncStatus(SyncStatus.Error);
    }
  }

  private async getDeviceInfo(): Promise<DeviceInfo> {
    // 获取设备信息
    const deviceId = await invoke('get_device_id') as string;
    const platform = await invoke('get_platform') as string;
    const appVersion = await invoke('get_app_version') as string;
    
    return {
      device_id: deviceId,
      device_type: this.getDeviceType(platform),
      device_name: await this.getDeviceName(),
      platform,
      app_version: appVersion,
      last_sync: new Date().toISOString(),
      is_online: true,
    };
  }

  private getDeviceType(platform: string): DeviceType {
    switch (platform.toLowerCase()) {
      case 'windows':
      case 'macos':
      case 'linux':
        return DeviceType.Desktop;
      case 'ios':
      case 'android':
        return DeviceType.Mobile;
      default:
        return DeviceType.Desktop;
    }
  }

  private async getDeviceName(): Promise<string> {
    try {
      return await invoke('get_device_name') as string;
    } catch {
      return `${this.getDeviceType(await invoke('get_platform') as string)} Device`;
    }
  }

  private getCurrentUserId(): string {
    // 从本地存储或状态管理中获取当前用户ID
    return localStorage.getItem('userId') || '';
  }

  // 消息处理方法
  private handleDataUpdate(data: any): void {
    console.log('收到数据更新:', data);
    // 处理数据更新逻辑
  }

  private handleChatMessage(data: any): void {
    console.log('收到聊天消息:', data);
    // 处理聊天消息逻辑
  }

  private handleFriendStatusUpdate(data: any): void {
    console.log('收到好友状态更新:', data);
    // 处理好友状态更新逻辑
  }

  private handleSyncResponse(data: any): void {
    console.log('收到同步响应:', data);
    // 处理同步响应逻辑
  }

  private handleDeviceStatusChange(data: any): void {
    console.log('设备状态变化:', data);
    // 处理设备状态变化逻辑
  }
}

// 导出单例实例
export const multiDeviceSync = MultiDeviceSyncManager.getInstance();