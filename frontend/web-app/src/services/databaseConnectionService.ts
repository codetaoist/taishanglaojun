import { apiClient } from './api';
import type {
  DatabaseConnectionForm,
  DatabaseConnectionListItem,
  DatabaseConnectionStatus,
  DatabaseConnectionTest,
  DatabaseConnectionStats,
  DatabaseConnectionQuery,
  DatabaseConnectionResponse,
  DatabaseConnectionListResponse,
  DatabaseConnectionStatusResponse,
  DatabaseConnectionTestResponse,
  DatabaseConnectionStatsResponse
} from '../types/database';
import { DatabaseType } from '../types/database';

export class DatabaseConnectionService {
  private static readonly BASE_URL = '/admin/database/connections';

  /**
   * 获取数据库连接列表
   */
  static async getConnections(query?: DatabaseConnectionQuery): Promise<DatabaseConnectionListResponse> {
    try {
      const params = new URLSearchParams();
      
      if (query?.page) params.append('page', query.page.toString());
      if (query?.pageSize) params.append('page_size', query.pageSize.toString());
      if (query?.search) params.append('search', query.search);
      if (query?.type) params.append('type', query.type);
      if (query?.status) params.append('status', query.status);
      if (query?.tags?.length) params.append('tags', query.tags.join(','));
      if (query?.sortBy) params.append('sort_by', query.sortBy);
      if (query?.sortOrder) params.append('sort_order', query.sortOrder);

      const response = await apiClient.get(`${this.BASE_URL}?${params.toString()}`);
      return response.data;
    } catch (error) {
      console.error('获取数据库连接列表失败:', error);
      throw error;
    }
  }

  /**
   * 获取单个数据库连接配置
   */
  static async getConnection(id: string): Promise<DatabaseConnectionResponse> {
    try {
      const response = await apiClient.get(`${this.BASE_URL}/${id}`);
      return response.data;
    } catch (error) {
      console.error('获取数据库连接配置失败:', error);
      throw error;
    }
  }

  /**
   * 创建数据库连接配置
   */
  static async createConnection(config: DatabaseConnectionForm): Promise<DatabaseConnectionResponse> {
    try {
      const response = await apiClient.post(this.BASE_URL, config);
      return response.data;
    } catch (error) {
      console.error('创建数据库连接配置失败:', error);
      throw error;
    }
  }

  /**
   * 更新数据库连接配置
   */
  static async updateConnection(id: string, config: Partial<DatabaseConnectionForm>): Promise<DatabaseConnectionResponse> {
    try {
      const response = await apiClient.put(`${this.BASE_URL}/${id}`, config);
      return response.data;
    } catch (error) {
      console.error('更新数据库连接配置失败:', error);
      throw error;
    }
  }

  /**
   * 删除数据库连接配置
   */
  static async deleteConnection(id: string): Promise<{ success: boolean; message?: string }> {
    try {
      const response = await apiClient.delete(`${this.BASE_URL}/${id}`);
      return response.data;
    } catch (error) {
      console.error('删除数据库连接配置失败:', error);
      throw error;
    }
  }

  /**
   * 测试数据库连接
   */
  static async testConnection(config: DatabaseConnectionForm): Promise<DatabaseConnectionTestResponse> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/test`, config);
      return response.data;
    } catch (error) {
      console.error('测试数据库连接失败:', error);
      throw error;
    }
  }

  /**
   * 测试已保存的数据库连接
   */
  static async testSavedConnection(id: string): Promise<DatabaseConnectionTestResponse> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/${id}/test`);
      return response.data;
    } catch (error) {
      console.error('测试已保存的数据库连接失败:', error);
      throw error;
    }
  }

  /**
   * 获取所有数据库连接状态
   */
  static async getConnectionsStatus(): Promise<DatabaseConnectionStatusResponse> {
    try {
      const response = await apiClient.get(`${this.BASE_URL}/status`);
      return response.data;
    } catch (error) {
      console.error('获取数据库连接状态失败:', error);
      throw error;
    }
  }

  /**
   * 获取单个数据库连接状态
   */
  static async getConnectionStatus(id: string): Promise<{ success: boolean; data?: DatabaseConnectionStatus; message?: string }> {
    try {
      const response = await apiClient.get(`${this.BASE_URL}/${id}/status`);
      return response.data;
    } catch (error) {
      console.error('获取数据库连接状态失败:', error);
      throw error;
    }
  }

  /**
   * 刷新数据库连接状态
   */
  static async refreshConnectionStatus(id: string): Promise<{ success: boolean; data?: DatabaseConnectionStatus; message?: string }> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/${id}/status`);
      return response.data;
    } catch (error) {
      console.error('刷新数据库连接状态失败:', error);
      throw error;
    }
  }

  /**
   * 连接到数据库
   */
  static async connectToDatabase(id: string): Promise<{ success: boolean; message?: string }> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/${id}/connect`);
      return response.data;
    } catch (error) {
      console.error('连接数据库失败:', error);
      throw error;
    }
  }

  /**
   * 断开数据库连接
   */
  static async disconnectFromDatabase(id: string): Promise<{ success: boolean; message?: string }> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/${id}/disconnect`);
      return response.data;
    } catch (error) {
      console.error('断开数据库连接失败:', error);
      throw error;
    }
  }

  /**
   * 获取数据库连接统计信息
   */
  static async getConnectionStats(): Promise<DatabaseConnectionStatsResponse> {
    try {
      const response = await apiClient.get(`${this.BASE_URL}/stats`);
      return response.data;
    } catch (error) {
      console.error('获取数据库连接统计信息失败:', error);
      throw error;
    }
  }

  /**
   * 批量操作数据库连接
   */
  static async batchOperation(
    operation: 'connect' | 'disconnect' | 'test' | 'delete',
    connectionIds: string[]
  ): Promise<{ success: boolean; results: Array<{ id: string; success: boolean; message?: string }>; message?: string }> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/batch`, {
        operation,
        connectionIds
      });
      return response.data;
    } catch (error: any) {
      // 若后端未实现 /batch 端点，使用前端降级逻辑处理部分操作
      const status = error?.response?.status;
      if (status === 404 || status === 405 || status === 501) {
        const results: Array<{ id: string; success: boolean; message?: string }> = [];
        for (const id of connectionIds) {
          try {
            if (operation === 'test') {
              const res = await this.testSavedConnection(id);
              results.push({ id, success: !!res.success, message: res.message });
            } else if (operation === 'delete') {
              const res = await this.deleteConnection(id);
              results.push({ id, success: !!res.success, message: res.message });
            } else {
              results.push({ id, success: false, message: '该操作不支持降级批量处理' });
            }
          } catch (e: any) {
            results.push({ id, success: false, message: e?.message || '操作失败' });
          }
        }
        return { success: true, results, message: '已使用前端降级逻辑完成批量操作（仅 test/delete）' };
      }
      console.error('批量操作数据库连接失败:', error);
      throw error;
    }
  }

  /**
   * 导出数据库连接配置
   */
  static async exportConnections(connectionIds?: string[]): Promise<Blob> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/export`, 
        { connectionIds },
        { responseType: 'blob' }
      );
      return response.data;
    } catch (error: any) {
      // 若后端未实现 /export，前端生成 JSON 导出
      const status = error?.response?.status;
      if (status === 404 || status === 405 || status === 501) {
        try {
          // 收集需要导出的连接 ID
          let ids: string[] = connectionIds || [];
          if (!ids.length) {
            const list = await this.getConnections({ page: 1, pageSize: 1000 });
            ids = list.data?.connections?.map(c => c.id) || [];
          }

          const configs: any[] = [];
          for (const id of ids) {
            try {
              const res = await this.getConnection(id);
              if (res?.data) configs.push(res.data);
            } catch {}
          }

          const payload = {
            success: true,
            exportedAt: new Date().toISOString(),
            count: configs.length,
            connections: configs
          };
          const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
          return blob;
        } catch (e) {
          console.error('前端生成导出数据失败:', e);
          throw e;
        }
      }
      console.error('导出数据库连接配置失败:', error);
      throw error;
    }
  }

  /**
   * 导入数据库连接配置
   */
  static async importConnections(file: File): Promise<{ success: boolean; imported: number; failed: number; message?: string }> {
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await apiClient.post(`${this.BASE_URL}/import`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });
      return response.data;
    } catch (error: any) {
      // 若后端未实现 /import，前端解析文件并逐条创建
      const status = error?.response?.status;
      if (status === 404 || status === 405 || status === 501) {
        try {
          const text = await file.text();
          const data = JSON.parse(text);
          const items: any[] = Array.isArray(data) ? data : (data?.connections || []);
          let imported = 0;
          let failed = 0;
          for (const item of items) {
            const form: DatabaseConnectionForm = {
              name: item.name,
              type: item.type,
              host: item.host,
              port: item.port,
              database: item.database,
              username: item.username,
              password: item.password || '',
              ssl: item.ssl,
              connectionTimeout: item.connectionTimeout,
              maxConnections: item.maxConnections,
              description: item.description,
              tags: item.tags,
              isDefault: item.isDefault,
            };
            try {
              const res = await this.createConnection(form);
              if (res?.success) imported++; else failed++;
            } catch {
              failed++;
            }
          }
          return { success: true, imported, failed, message: '已使用前端降级逻辑完成导入' };
        } catch (e) {
          console.error('前端导入解析失败:', e);
          throw e;
        }
      }
      console.error('导入数据库连接配置失败:', error);
      throw error;
    }
  }

  /**
   * 获取支持的数据库类型
   */
  static async getSupportedDatabaseTypes(): Promise<{ success: boolean; data?: DatabaseType[]; message?: string }> {
    try {
      const response = await apiClient.get(`${this.BASE_URL}/supported-types`);
      return response.data;
    } catch (error) {
      console.error('获取支持的数据库类型失败:', error);
      throw error;
    }
  }

  /**
   * 获取数据库连接模板
   */
  static async getConnectionTemplate(type: DatabaseType): Promise<{ success: boolean; data?: Partial<DatabaseConnectionForm>; message?: string }> {
    try {
      const response = await apiClient.get(`${this.BASE_URL}/template/${type}`);
      return response.data;
    } catch (error) {
      console.error('获取数据库连接模板失败:', error);
      throw error;
    }
  }

  /**
   * 验证数据库连接配置
   */
  static async validateConnection(config: DatabaseConnectionForm): Promise<{ success: boolean; errors?: string[]; message?: string }> {
    try {
      const response = await apiClient.post(`${this.BASE_URL}/validate`, config);
      return response.data;
    } catch (error) {
      console.error('验证数据库连接配置失败:', error);
      throw error;
    }
  }
}