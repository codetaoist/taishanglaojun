import { invoke } from '@tauri-apps/api/core';

// 数据类型定义
export interface User {
  id: string;
  username: string;
  email: string;
  display_name?: string;
  avatar_url?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ChatSessionWithMessages {
  session: {
    id: string;
    title: string;
    chat_type: string;
    created_at: string;
    updated_at: string;
    message_count: number;
  };
  messages: Array<{
    id: string;
    session_id: string;
    role: string;
    content: string;
    message_type: string;
    metadata?: string;
    created_at: string;
  }>;
  user_info?: User;
}

export interface ProjectWithFiles {
  id: string;
  name: string;
  description?: string;
  owner: User;
  files: Array<{
    path: string;
    name: string;
    size: number;
    created_at: string;
    modified_at: string;
    file_type?: string;
    hash?: string;
  }>;
  created_at: string;
}

export interface DatabaseHealthStatus {
  main_db_healthy: boolean;
  chat_db_healthy: boolean;
  storage_db_healthy: boolean;
  main_db_error?: string;
  chat_db_error?: string;
  storage_db_error?: string;
}

export interface SearchResults {
  users: Array<{
    id: string;
    username: string;
    email: string;
    display_name?: string;
  }>;
  chat_sessions: Array<{
    id: string;
    title: string;
    chat_type: string;
  }>;
  files: Array<{
    path: string;
    name: string;
    file_type?: string;
  }>;
}

export interface DatabaseStats {
  main_db: {
    users: number;
    projects: number;
  };
  chat_db: {
    sessions: number;
    messages: number;
  };
  storage_db: {
    files: number;
    total_size: number;
  };
}

/**
 * 数据管理器类 - 统一管理跨数据库的数据访问
 */
export class DataManager {
  /**
   * 获取用户详细信息（包含统计数据）
   */
  static async getUserWithStats(userId: string): Promise<User | null> {
    try {
      return await invoke<User | null>('get_user_with_stats', { userId });
    } catch (error) {
      console.error('Failed to get user with stats:', error);
      throw error;
    }
  }

  /**
   * 跨数据库搜索
   */
  static async searchAllData(query: string): Promise<SearchResults> {
    try {
      const results = await invoke<Record<string, any[]>>('search_all_data', { query });
      return {
        users: results.users || [],
        chat_sessions: results.chat_sessions || [],
        files: results.files || [],
      };
    } catch (error) {
      console.error('Failed to search all data:', error);
      throw error;
    }
  }

  /**
   * 获取数据库统计信息
   */
  static async getDatabaseStatistics(): Promise<DatabaseStats> {
    try {
      return await invoke<DatabaseStats>('get_database_statistics');
    } catch (error) {
      console.error('Failed to get database statistics:', error);
      throw error;
    }
  }

  /**
   * 检查数据库健康状态
   */
  static async checkDatabaseHealth(): Promise<DatabaseHealthStatus> {
    try {
      return await invoke<DatabaseHealthStatus>('check_database_health');
    } catch (error) {
      console.error('Failed to check database health:', error);
      throw error;
    }
  }

  /**
   * 获取聊天会话及相关信息
   */
  static async getChatSessionWithContext(sessionId: string): Promise<ChatSessionWithMessages | null> {
    try {
      // 这里可能需要调用后端的相应命令，目前先返回null
      // 可以根据需要添加相应的Tauri命令
      console.warn('getChatSessionWithContext not implemented yet for session:', sessionId);
      return null;
    } catch (error) {
      console.error('Failed to get chat session with context:', error);
      throw error;
    }
  }

  /**
   * 获取项目及其文件信息
   */
  static async getProjectWithFiles(projectId: string): Promise<ProjectWithFiles | null> {
    try {
      // 这里可能需要调用后端的相应命令，目前先返回null
      // 可以根据需要添加相应的Tauri命令
      console.warn('getProjectWithFiles not implemented yet for project:', projectId);
      return null;
    } catch (error) {
      console.error('Failed to get project with files:', error);
      throw error;
    }
  }
}

/**
 * 数据同步状态管理
 */
export class DataSyncStatus {
  private static listeners: Array<(status: any) => void> = [];

  /**
   * 添加同步状态监听器
   */
  static addListener(callback: (status: any) => void): void {
    this.listeners.push(callback);
  }

  /**
   * 移除同步状态监听器
   */
  static removeListener(callback: (status: any) => void): void {
    const index = this.listeners.indexOf(callback);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  /**
   * 通知所有监听器
   */
  private static notifyListeners(status: any): void {
    this.listeners.forEach(callback => callback(status));
  }

  /**
   * 检查数据同步状态
   */
  static async checkSyncStatus(): Promise<void> {
    try {
      const health = await DataManager.checkDatabaseHealth();
      const stats = await DataManager.getDatabaseStatistics();
      
      const status = {
        healthy: health.main_db_healthy && health.chat_db_healthy && health.storage_db_healthy,
        errors: [
          health.main_db_error,
          health.chat_db_error,
          health.storage_db_error
        ].filter(Boolean),
        stats,
        timestamp: new Date().toISOString(),
      };

      this.notifyListeners(status);
    } catch (error) {
      console.error('Failed to check sync status:', error);
      this.notifyListeners({
        healthy: false,
        errors: [error instanceof Error ? error.message : 'Unknown error'],
        stats: null,
        timestamp: new Date().toISOString(),
      });
    }
  }
}

/**
 * 数据缓存管理
 */
export class DataCache {
  private static cache = new Map<string, { data: any; timestamp: number; ttl: number }>();
  private static readonly DEFAULT_TTL = 5 * 60 * 1000; // 5分钟

  /**
   * 设置缓存
   */
  static set(key: string, data: any, ttl: number = this.DEFAULT_TTL): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  /**
   * 获取缓存
   */
  static get<T>(key: string): T | null {
    const item = this.cache.get(key);
    if (!item) return null;

    if (Date.now() - item.timestamp > item.ttl) {
      this.cache.delete(key);
      return null;
    }

    return item.data as T;
  }

  /**
   * 删除缓存
   */
  static delete(key: string): void {
    this.cache.delete(key);
  }

  /**
   * 清空缓存
   */
  static clear(): void {
    this.cache.clear();
  }

  /**
   * 获取缓存统计
   */
  static getStats(): { size: number; keys: string[] } {
    return {
      size: this.cache.size,
      keys: Array.from(this.cache.keys()),
    };
  }
}

/**
 * 数据管理工具函数
 */
export const DataUtils = {
  /**
   * 格式化文件大小
   */
  formatFileSize(bytes: number): string {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(2)} ${units[unitIndex]}`;
  },

  /**
   * 格式化日期
   */
  formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  },

  /**
   * 生成缓存键
   */
  generateCacheKey(prefix: string, ...params: (string | number)[]): string {
    return `${prefix}:${params.join(':')}`;
  },

  /**
   * 防抖函数
   */
  debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): (...args: Parameters<T>) => void {
    let timeout: NodeJS.Timeout | undefined;
    return (...args: Parameters<T>) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(this, args), wait);
    };
  },

  /**
   * 节流函数
   */
  throttle<T extends (...args: any[]) => any>(
    func: T,
    limit: number
  ): (...args: Parameters<T>) => void {
    let inThrottle: boolean;
    return (...args: Parameters<T>) => {
      if (!inThrottle) {
        func.apply(this, args);
        inThrottle = true;
        setTimeout(() => (inThrottle = false), limit);
      }
    };
  },
};

export default DataManager;