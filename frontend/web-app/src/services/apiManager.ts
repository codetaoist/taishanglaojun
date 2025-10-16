import { apiClient } from './api';
import { getNotificationInstance } from './notificationService';
import type { ApiResponse } from '../types';

export interface RequestConfig {
  showLoading?: boolean;
  showError?: boolean;
  showSuccess?: boolean;
  successMessage?: string;
  errorMessage?: string;
  silent?: boolean; // 静默模式，不显示任何提示
}

class ApiManager {
  private apiClient = apiClient;
  private loadingRequests: Set<string> = new Set();

  // 生成请求唯一标识
  private generateRequestId(method: string, url: string): string {
    return `${method}:${url}`;
  }

  // 设置加载状态
  private setLoading(requestId: string, loading: boolean) {
    if (loading) {
      this.loadingRequests.add(requestId);
    } else {
      this.loadingRequests.delete(requestId);
    }
    
    // 触发全局加载状态更新
    window.dispatchEvent(new CustomEvent('api-loading-change', {
      detail: { loading: this.loadingRequests.size > 0 }
    }));
  }

  // 处理API响应
  private handleResponse<T>(
    response: ApiResponse<T>,
    config: RequestConfig = {}
  ): T {
    const {
      showError = true,
      showSuccess = false,
      successMessage,
      errorMessage,
      silent = false
    } = config;

    const notification = getNotificationInstance();
    
    if (response.success) {
      if (showSuccess && successMessage && !silent) {
        notification.success({
          message: '操作成功',
          description: successMessage,
          duration: 3
        });
      }
      return response.data;
    } else {
      if (showError && !silent) {
        notification.error({
          message: '操作失败',
          description: errorMessage || response.error || '请求失败，请稍后重试',
          duration: 4
        });
      }
      throw new Error(response.error || '请求失败');
    }
  }

  // 通用请求方法
  async request<T>(
    apiMethod: () => Promise<ApiResponse<T>>,
    config: RequestConfig = {}
  ): Promise<T> {
    const requestId = Math.random().toString(36).substr(2, 9);
    
    try {
      if (config.showLoading) {
        this.setLoading(requestId, true);
      }

      const response = await apiMethod();
      return this.handleResponse(response, config);
    } catch (error) {
      if (config.showError && !config.silent) {
        const errorMsg = error instanceof Error ? error.message : '未知错误';
        const notification = getNotificationInstance();
        notification.error({
          message: '请求失败',
          description: config.errorMessage || errorMsg,
          duration: 4
        });
      }
      throw error;
    } finally {
      if (config.showLoading) {
        this.setLoading(requestId, false);
      }
    }
  }

  // 用户相关API
  async login(username: string, password: string, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.login(username, password),
      {
        showLoading: true,
        showSuccess: true,
        successMessage: '登录成功',
        errorMessage: '登录失败，请检查用户名和密码',
        ...config
      }
    );
  }

  async register(username: string, email: string, password: string, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.register(username, email, password),
      {
        showLoading: true,
        showSuccess: true,
        successMessage: '注册成功',
        errorMessage: '注册失败，请检查输入信息',
        ...config
      }
    );
  }

  async getCurrentUser(config?: RequestConfig) {
    return this.request(
      () => this.apiClient.getCurrentUser(),
      {
        showLoading: false,
        showError: false, // 用户信息获取失败通常不需要显示错误
        ...config
      }
    );
  }

  async logout(config?: RequestConfig) {
    return this.request(
      () => this.apiClient.logout(),
      {
        showLoading: true,
        showSuccess: true,
        successMessage: '已安全退出',
        ...config
      }
    );
  }

  async updateProfile(profileData: any, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.updateProfile(profileData),
      {
        showLoading: true,
        showSuccess: true,
        successMessage: '个人信息更新成功',
        errorMessage: '更新失败，请稍后重试',
        ...config
      }
    );
  }

  // 文化智慧相关API
  async getWisdomList(filters?: any, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.getWisdomList(filters),
      {
        showLoading: true,
        showError: true,
        errorMessage: '获取智慧内容失败',
        ...config
      }
    );
  }

  async getWisdomById(id: string, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.getWisdomById(id),
      {
        showLoading: true,
        showError: true,
        errorMessage: '获取智慧详情失败',
        ...config
      }
    );
  }

  async searchWisdom(query: string, filters?: any, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.searchWisdom(query, filters),
      {
        showLoading: true,
        showError: true,
        errorMessage: '搜索失败，请稍后重试',
        ...config
      }
    );
  }

  // AI对话相关API
  async sendChatMessage(message: string, conversationId?: string, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.sendChatMessage(message, conversationId),
      {
        showLoading: true,
        showError: true,
        errorMessage: '发送消息失败',
        ...config
      }
    );
  }

  async getChatHistory(conversationId?: string, config?: RequestConfig) {
    return this.request(
      () => this.apiClient.getChatHistory(conversationId),
      {
        showLoading: false,
        showError: true,
        errorMessage: '获取聊天记录失败',
        ...config
      }
    );
  }

  async getConversations(config?: RequestConfig) {
    return this.request(
      () => this.apiClient.getConversations(),
      {
        showLoading: false,
        showError: true,
        errorMessage: '获取对话列表失败',
        ...config
      }
    );
  }

  // 获取当前加载状态
  get isLoading(): boolean {
    return this.loadingRequests.size > 0;
  }

  // 获取加载中的请求数量
  get loadingCount(): number {
    return this.loadingRequests.size;
  }

  // 清除所有加载状态（用于组件卸载时）
  clearAllLoading(): void {
    this.loadingRequests.clear();
    window.dispatchEvent(new CustomEvent('api-loading-change', {
      detail: { loading: false }
    }));
  }
}

// 创建单例实例
export const apiManager = new ApiManager();
export default apiManager;