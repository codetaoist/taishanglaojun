import { apiClient } from './api';
import type { AxiosResponse } from 'axios';

const API_BASE_URL = '/api/v1/enhanced-database';

// 增强的数据库管理服务类型定义
export interface BackupStatus {
  id: string;
  name: string;
  description?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  progress: number;
  file_path?: string;
  file_size?: number;
  created_at: string;
  completed_at?: string;
  created_by: string;
  error_message?: string;
  backup_type: 'full' | 'incremental' | 'differential';
  compression: boolean;
  encryption: boolean;
}

export interface DatabaseMetrics {
  timestamp: string;
  connections: {
    total: number;
    active: number;
    idle: number;
    waiting: number;
  };
  performance: {
    queries_per_second: number;
    avg_query_time: number;
    slow_queries: number;
    cache_hit_ratio: number;
  };
  storage: {
    total_size: number;
    used_size: number;
    free_size: number;
    table_count: number;
    index_size: number;
  };
  memory: {
    buffer_pool_size: number;
    buffer_pool_used: number;
    sort_buffer_size: number;
  };
}

export interface ConnectionInfo {
  id: string;
  name: string;
  type: string;
  host: string;
  port: number;
  database: string;
  username: string;
  status: 'connected' | 'disconnected' | 'error' | 'testing';
  last_ping: string;
  response_time: number;
  error_message?: string;
  connection_pool: {
    max_connections: number;
    active_connections: number;
    idle_connections: number;
    waiting_connections: number;
  };
  health_score: number;
  tags: string[];
  // 新增活跃连接相关字段
  process_id?: number;
  state?: string;
  query?: string;
  duration?: number;
  client_addr?: string;
  application_name?: string;
  backend_start?: string;
  query_start?: string;
  state_change?: string;
  waiting?: boolean;
  wait_event_type?: string;
  wait_event?: string;
}

export interface BackupListResponse {
  success: boolean;
  data: {
    backups: BackupStatus[];
    total: number;
    page: number;
    limit: number;
    pages: number;
  };
  message?: string;
}

export interface MetricsResponse {
  success: boolean;
  data: DatabaseMetrics;
  message?: string;
}

export interface ConnectionListResponse {
  success: boolean;
  data: {
    connections: ConnectionInfo[];
    total: number;
    page?: number;
    pageSize?: number;
    totalPages?: number;
    summary: {
      total_connections: number;
      healthy_connections: number;
      unhealthy_connections: number;
      avg_response_time: number;
      active_queries?: number;
      idle_connections?: number;
      waiting_connections?: number;
    };
  };
  message?: string;
}

// 错误类型定义
export interface DatabaseError {
  code: string;
  message: string;
  details?: any;
  timestamp: Date;
  retryable: boolean;
}

// 重试配置
interface RetryConfig {
  maxRetries: number;
  baseDelay: number;
  maxDelay: number;
  backoffFactor: number;
}

// 缓存配置
interface CacheConfig {
  ttl: number; // 缓存时间（毫秒）
  key: string;
}

// 缓存管理
class CacheManager {
  private cache = new Map<string, { data: any; timestamp: number; ttl: number }>();

  set(key: string, data: any, ttl: number = 30000): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  get(key: string): any | null {
    const item = this.cache.get(key);
    if (!item) return null;

    if (Date.now() - item.timestamp > item.ttl) {
      this.cache.delete(key);
      return null;
    }

    return item.data;
  }

  clear(): void {
    this.cache.clear();
  }
}

// 错误处理器
class ErrorHandler {
  static createError(error: any): DatabaseError {
    const timestamp = new Date();
    
    if (error.response) {
      // HTTP错误响应
      const status = error.response.status;
      const data = error.response.data;
      
      return {
        code: `HTTP_${status}`,
        message: data?.message || data?.error || `HTTP ${status} Error`,
        details: data,
        timestamp,
        retryable: this.isRetryableHttpError(status),
      };
    } else if (error.request) {
      // 网络错误
      return {
        code: 'NETWORK_ERROR',
        message: '网络连接失败，请检查网络连接',
        details: error.message,
        timestamp,
        retryable: true,
      };
    } else {
      // 其他错误
      return {
        code: 'UNKNOWN_ERROR',
        message: error.message || '未知错误',
        details: error,
        timestamp,
        retryable: false,
      };
    }
  }

  private static isRetryableHttpError(status: number): boolean {
    // 5xx服务器错误和429限流错误可重试
    return status >= 500 || status === 429;
  }
}

// 重试机制
class RetryManager {
  private static defaultConfig: RetryConfig = {
    maxRetries: 3,
    baseDelay: 1000,
    maxDelay: 10000,
    backoffFactor: 2,
  };

  static async executeWithRetry<T>(
    operation: () => Promise<T>,
    config: Partial<RetryConfig> = {}
  ): Promise<T> {
    const finalConfig = { ...this.defaultConfig, ...config };
    let lastError: DatabaseError;

    for (let attempt = 0; attempt <= finalConfig.maxRetries; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = ErrorHandler.createError(error);
        
        // 如果不可重试或已达到最大重试次数，直接抛出错误
        if (!lastError.retryable || attempt === finalConfig.maxRetries) {
          throw lastError;
        }

        // 计算延迟时间（指数退避）
        const delay = Math.min(
          finalConfig.baseDelay * Math.pow(finalConfig.backoffFactor, attempt),
          finalConfig.maxDelay
        );

        console.warn(`操作失败，${delay}ms后进行第${attempt + 1}次重试:`, lastError.message);
        await this.delay(delay);
      }
    }

    throw lastError!;
  }

  private static delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

class EnhancedDatabaseService {
  private cache = new CacheManager();
  private wsConnections = new Map<string, WebSocket>();

  // 通用请求方法，包含错误处理和重试
  private async request<T>(
    operation: () => Promise<AxiosResponse<T>>,
    cacheKey?: string,
    cacheTtl?: number,
    retryConfig?: Partial<RetryConfig>
  ): Promise<T> {
    // 检查缓存
    if (cacheKey) {
      const cached = this.cache.get(cacheKey);
      if (cached) {
        return cached;
      }
    }

    try {
      const response = await RetryManager.executeWithRetry(operation, retryConfig);
      const data = response.data;

      // 设置缓存
      if (cacheKey && cacheTtl) {
        this.cache.set(cacheKey, data, cacheTtl);
      }

      return data;
    } catch (error) {
      console.error('数据库服务请求失败:', error);
      throw error;
    }
  }

  private clearCache(pattern?: string): void {
    if (!pattern) {
      this.cache.clear();
      return;
    }

    // CacheManager doesn't have pattern clearing, so we'll clear all
    this.cache.clear();
  }

  // 辅助方法，用于向后兼容
  private async withRetry<T>(operation: () => Promise<T>, retryConfig?: Partial<RetryConfig>): Promise<T> {
    return RetryManager.executeWithRetry(operation, retryConfig);
  }

  private getCache(key: string): any {
    return this.cache.get(key);
  }

  private setCache(key: string, data: any, ttl: number): void {
    this.cache.set(key, data, ttl);
  }

  // 备份管理 - 增强版
  async createBackup(
    name: string,
    options: {
      description?: string;
      backup_type?: 'full' | 'incremental' | 'differential';
      compression?: boolean;
      encryption?: boolean;
      tables?: string[];
    } = {}
  ): Promise<{ success: boolean; backup: BackupStatus; message?: string }> {
    const result = await this.request(
      () => apiClient.post('/system/backup', {
        name,
        description: options.description,
        backup_type: options.backup_type || 'full',
        compression: options.compression ?? true,
        encryption: options.encryption ?? false,
        tables: options.tables
      }),
      undefined, // 不缓存创建操作
      undefined,
      { maxRetries: 2, baseDelay: 1000 }
    );

    this.clearCache('backups');
    return result;
  }

  async getBackups(
    params: {
      page?: number;
      limit?: number;
      status?: string;
      backup_type?: string;
      search?: string;
    } = {},
    useCache: boolean = true
  ): Promise<BackupListResponse> {
    const cacheKey = `backups_${JSON.stringify(params)}`;
    
    if (useCache) {
      const cached = this.getCache(cacheKey);
      if (cached) return cached;
    }

    return this.withRetry(async () => {
      const response = await apiClient.get('/system/backups', { params });
      const result = response.data;

      if (useCache && result.success) {
        this.setCache(cacheKey, result, 30000); // 30秒缓存
      }

      return result;
    });
  }

  async getBackupStatus(id: string): Promise<{ success: boolean; data: BackupStatus; message?: string }> {
    return this.withRetry(async () => {
      const response = await apiClient.get(`/system/backups/${id}`);
      return response.data;
    });
  }

  async cancelBackup(id: string): Promise<{ success: boolean; message?: string }> {
    return this.withRetry(async () => {
      const response = await apiClient.post(`/system/backups/${id}/cancel`);
      this.clearCache('backups');
      return response.data;
    });
  }

  async deleteBackup(id: string): Promise<{ success: boolean; message?: string }> {
    return this.withRetry(async () => {
      const response = await apiClient.delete(`/system/backups/${id}`);
      this.clearCache('backups');
      return response.data;
    });
  }

  async restoreBackup(
    id: string,
    options: {
      target_database?: string;
      overwrite?: boolean;
      verify_before_restore?: boolean;
    } = {}
  ): Promise<{ success: boolean; message?: string; restore_id?: string }> {
    return this.withRetry(async () => {
      const response = await apiClient.post(`/system/restore/${id}`, options);
      return response.data;
    });
  }

  async verifyBackup(id: string): Promise<{ success: boolean; valid: boolean; message?: string; details?: any }> {
    return this.withRetry(async () => {
      const response = await apiClient.post(`/system/backups/${id}/verify`);
      return response.data;
    });
  }

  // 实时数据库监控 - 增强版
  async getDatabaseMetrics(useCache: boolean = true): Promise<MetricsResponse> {
    return this.request(
      () => apiClient.get('/system/database/metrics'),
      useCache ? 'database_metrics' : undefined,
      useCache ? 5000 : undefined, // 5秒缓存
      { maxRetries: 2, baseDelay: 500 } // 快速重试配置
    );
  }

  async getDatabaseMetricsHistory(
    timeRange: '1h' | '6h' | '24h' | '7d' | '30d' = '1h',
    interval: '1m' | '5m' | '15m' | '1h' | '1d' = '5m'
  ): Promise<{ success: boolean; data: DatabaseMetrics[]; message?: string }> {
    return this.request(
      () => apiClient.get('/system/database/metrics/history', {
        params: { time_range: timeRange, interval }
      }),
      `metrics_history_${timeRange}_${interval}`,
      60000, // 1分钟缓存
      { maxRetries: 3 } // 标准重试配置
    );
  }

  async getDatabaseHealth(): Promise<{
    success: boolean;
    data: {
      overall_health: 'healthy' | 'warning' | 'critical';
      health_score: number;
      issues: Array<{
        type: 'performance' | 'storage' | 'connection' | 'security';
        severity: 'low' | 'medium' | 'high' | 'critical';
        message: string;
        recommendation?: string;
      }>;
      last_check: string;
    };
    message?: string;
  }> {
    return this.request(
      () => apiClient.get('/system/database/health'),
      'database_health',
      15000, // 15秒缓存
      { maxRetries: 2, baseDelay: 1000 } // 健康检查重试配置
    );
  }

  // 连接管理 - 增强版
  async getActiveConnections(options: {
    page?: number;
    pageSize?: number;
    sortBy?: string;
    sortOrder?: 'asc' | 'desc';
    user?: string;
    database?: string;
    state?: string;
    useCache?: boolean;
  } = {}): Promise<ConnectionListResponse> {
    const { useCache = true, ...params } = options;
    
    // 构建查询参数
    const queryParams: any = {};
    if (params.page) queryParams.page = params.page;
    if (params.pageSize) queryParams.page_size = params.pageSize;
    if (params.sortBy) queryParams.sort_by = params.sortBy;
    if (params.sortOrder) queryParams.sort_order = params.sortOrder;
    if (params.user) queryParams.user = params.user;
    if (params.database) queryParams.database = params.database;
    if (params.state) queryParams.state = params.state;
    
    const cacheKey = useCache ? `active_connections_${JSON.stringify(queryParams)}` : undefined;
    
    return this.request(
      () => apiClient.get('/admin/database-connections/active', { params: queryParams }),
      cacheKey,
      useCache ? 5000 : undefined, // 5秒缓存
      { maxRetries: 2, baseDelay: 500 }
    );
  }

  async testAllConnections(): Promise<{
    success: boolean;
    data: Array<{
      id: string;
      name: string;
      status: 'success' | 'failed';
      response_time?: number;
      error_message?: string;
    }>;
    message?: string;
  }> {
    return this.request(
      () => apiClient.post('/admin/database-connections/test-all'),
      undefined, // 不缓存测试结果
      undefined,
      { maxRetries: 1, baseDelay: 2000 } // 测试操作重试配置
    );
  }

  async getConnectionPoolStats(): Promise<{
    success: boolean;
    data: {
      total_pools: number;
      total_connections: number;
      active_connections: number;
      idle_connections: number;
      waiting_connections: number;
      pools: Array<{
        name: string;
        max_size: number;
        current_size: number;
        active: number;
        idle: number;
        waiting: number;
        health_score: number;
      }>;
    };
    message?: string;
  }> {
    return this.request(
      () => apiClient.get('/admin/database-connections/pool-stats'),
      'connection_pool_stats',
      5000, // 5秒缓存
      { maxRetries: 2, baseDelay: 500 }
    );
  }

  // 性能优化
  async optimizeDatabase(
    options: {
      analyze_tables?: boolean;
      rebuild_indexes?: boolean;
      cleanup_logs?: boolean;
      vacuum_tables?: boolean;
    } = {}
  ): Promise<{ success: boolean; message?: string; optimization_id?: string }> {
    return this.request(
      () => apiClient.post('/system/database/optimize', options),
      undefined, // 不缓存优化操作
      undefined,
      { maxRetries: 1, baseDelay: 1000 } // 优化操作重试配置
    );
  }

  async getOptimizationStatus(id: string): Promise<{
    success: boolean;
    data: {
      id: string;
      status: 'running' | 'completed' | 'failed';
      progress: number;
      steps: Array<{
        name: string;
        status: 'pending' | 'running' | 'completed' | 'failed';
        duration?: number;
        message?: string;
      }>;
      started_at: string;
      completed_at?: string;
    };
    message?: string;
  }> {
    return this.withRetry(async () => {
      const response = await apiClient.get(`/system/database/optimize/${id}`);
      return response.data;
    });
  }

  // 终止活跃连接
  async killConnection(connectionId: string): Promise<{ success: boolean; message?: string }> {
    return this.request(
      () => apiClient.post(`/admin/database-connections/kill/${connectionId}`),
      undefined, // 不缓存操作结果
      undefined,
      { maxRetries: 1, baseDelay: 1000 }
    );
  }

  // 批量终止连接
  async killMultipleConnections(connectionIds: string[]): Promise<{
    success: boolean;
    results: Array<{
      id: string;
      success: boolean;
      message?: string;
    }>;
    message?: string;
  }> {
    return this.request(
      () => apiClient.post('/admin/database-connections/kill-multiple', { connection_ids: connectionIds }),
      undefined, // 不缓存操作结果
      undefined,
      { maxRetries: 1, baseDelay: 1000 }
    );
  }

  // 实时监控订阅（WebSocket）
  subscribeToMetrics(callback: (metrics: DatabaseMetrics) => void): () => void {
    // 这里可以实现WebSocket连接来获取实时数据
    // 目前使用轮询作为fallback
    const interval = setInterval(async () => {
      try {
        const response = await this.getDatabaseMetrics(false);
        if (response.success) {
          callback(response.data);
        }
      } catch (error) {
        console.error('获取实时指标失败:', error);
      }
    }, 5000);

    return () => clearInterval(interval);
  }

  // 清理缓存
  clearAllCache(): void {
    this.clearCache();
  }
}

export const enhancedDatabaseService = new EnhancedDatabaseService();
export default enhancedDatabaseService;