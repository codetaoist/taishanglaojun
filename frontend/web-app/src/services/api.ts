import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';
import type { 
  User, CulturalWisdom, SearchFilters, ChatMessage, Category, Tag, UserData, WisdomData, CategoryData, TagData, 
  UserStats, UserQueryParams, AnalyticsData, SystemConfig, SEOConfig, CacheConfig, Notification, NotificationData,
  ReviewItem, ReviewData, BatchReviewData, QueryParams
} from '../types';

// 定义API响应类型
interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
  error?: string;
}

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081',
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => response,
      (error) => {
        // 移除导致页面刷新的强制跳转逻辑
        // 让各个组件自己处理401错误，避免全局页面刷新
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          // 不再使用 window.location.href，让组件自己处理导航
        }
        return Promise.reject(error);
      }
    );
  }

  // 用户相关API
  async login(email: string, password: string): Promise<ApiResponse<{ user: User; token: string }>> {
    const response = await this.client.post('/auth/login', { email, password });
    
    // 后端直接返回用户数据和令牌，需要包装成前端期望的格式
    if (response.data.user && response.data.access_token) {
      return {
        success: true,
        data: {
          user: response.data.user,
          token: response.data.access_token
        }
      };
    } else {
      return {
        success: false,
        data: {} as { user: User; token: string },
        error: response.data.message || '登录失败'
      };
    }
  }

  async register(username: string, email: string, password: string): Promise<ApiResponse<{ user: User; token: string }>> {
    const response = await this.client.post('/auth/register', { username, email, password });
    return response.data;
  }

  async verifyEmail(data: { token: string }): Promise<ApiResponse<void>> {
    const response = await this.client.post('/auth/verify-email', data);
    return response.data;
  }

  async resendVerification(data: { email: string }): Promise<ApiResponse<void>> {
    const response = await this.client.post('/auth/resend-verification', data);
    return response.data;
  }

  async getCurrentUser(): Promise<ApiResponse<User>> {
    try {
      const response = await this.client.get('/user/me');
      
      // 后端直接返回用户数据，需要包装成前端期望的格式
      if (response.data && response.data.id) {
        return {
          success: true,
          data: response.data
        };
      } else {
        return {
          success: false,
          data: {} as User,
          error: '获取用户信息失败'
        };
      }
    } catch (error) {
      return {
        success: false,
        data: {} as User,
        error: (error as { response?: { data?: { message?: string } }; message?: string }).response?.data?.message || 
               (error as Error).message || '获取用户信息失败'
      };
    }
  }

  async logout(): Promise<ApiResponse<void>> {
    try {
      await this.client.post('/auth/logout');
      return {
        success: true,
        data: undefined
      };
    } catch (error) {
      // 即使API调用失败，也应该清除本地token
      return {
        success: false,
        data: undefined,
        error: (error as { response?: { data?: { message?: string } }; message?: string }).response?.data?.message || 
               (error as Error).message || '退出登录失败'
      };
    }
  }

  async updateProfile(profileData: Partial<User>): Promise<User> {
    const response = await this.client.put('/auth/profile', profileData);
    return response.data.data;
  }

  // 文化智慧相关API
  async getWisdomList(filters?: SearchFilters & { page?: number; pageSize?: number }): Promise<ApiResponse<{ items: CulturalWisdom[]; total: number; page: number; pageSize: number }>> {
    const response = await this.client.get('/cultural-wisdom/list', { params: filters });
    return response.data;
  }

  async getWisdomById(id: string): Promise<ApiResponse<CulturalWisdom>> {
    const response = await this.client.get(`/cultural-wisdom/${id}`);
    return response.data;
  }

  async searchWisdom(keyword: string, filters?: SearchFilters): Promise<ApiResponse<{ items: CulturalWisdom[]; total: number }>> {
    const response = await this.client.get('/cultural-wisdom/search', { 
      params: { q: keyword, ...filters } 
    });
    return response.data;
  }

  // AI聊天相关API
  async sendChatMessage(message: string, sessionId?: string, provider?: string, model?: string): Promise<ApiResponse<{ 
    session_id: string; 
    message_id: number; 
    content: string; 
    token_used: number; 
    provider: string; 
    model: string; 
  }>> {
    const response = await this.client.post('/ai/chat', { 
      message, 
      session_id: sessionId,
      provider: provider || 'openai',
      model: model || 'gpt-3.5-turbo'
    });
    return response.data;
  }

  async getChatHistory(sessionId: string): Promise<ApiResponse<ChatMessage[]>> {
    const response = await this.client.get(`/ai/sessions/${sessionId}/messages`);
    return response.data;
  }

  async getChatSessions(): Promise<ApiResponse<{ id: string; title: string; updated_at: string }[]>> {
    const response = await this.client.get('/ai/sessions');
    return response.data;
  }

  async deleteChatSession(sessionId: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/ai/sessions/${sessionId}`);
    return response.data;
  }

  // AI提供商相关API
  async getAIProviders(): Promise<ApiResponse<{ providers: string[] }>> {
    const response = await this.client.get('/ai/providers');
    return response.data;
  }

  async getAIModels(): Promise<ApiResponse<{ providers: Record<string, string[]> }>> {
    const response = await this.client.get('/ai/models');
    return response.data;
  }

  async checkAIHealth(): Promise<ApiResponse<{ status: string }>> {
    const response = await this.client.get('/ai/health');
    return response.data;
  }

  // AI智慧功能相关API
  async getWisdomInterpretation(wisdomId: string): Promise<ApiResponse<{ interpretation: string }>> {
    const response = await this.client.post(`/cultural-wisdom/ai/${wisdomId}/interpret`);
    return response.data;
  }

  // 新的推荐API方法
  async getRecommendations(wisdomId: string, params?: {
    limit?: number;
    algorithm?: string;
    categories?: string;
    schools?: string;
    authors?: string;
  }): Promise<ApiResponse<Array<{
    wisdom_id: string;
    title: string;
    author: string;
    category: string;
    school: string;
    summary: string;
    score: number;
    reason: string;
    view_count: number;
    like_count: number;
    created_at: string;
  }>>> {
    const response = await this.client.get(`/cultural-wisdom/${wisdomId}/recommendations`, { params });
    return response.data;
  }

  async getSimilarWisdoms(wisdomId: string, limit?: number): Promise<ApiResponse<Array<{
    wisdom_id: string;
    title: string;
    author: string;
    category: string;
    school: string;
    summary: string;
    score: number;
    reason: string;
    view_count: number;
    like_count: number;
    created_at: string;
  }>>> {
    const response = await this.client.get(`/cultural-wisdom/${wisdomId}/similar`, { 
      params: { limit } 
    });
    return response.data;
  }

  // 搜索建议API
  async getSearchSuggestions(query: string, limit?: number): Promise<ApiResponse<{
    suggestions: string[];
    count: number;
  }>> {
    const response = await this.client.get('/cultural-wisdom/search/suggestions', {
      params: { q: query, limit }
    });
    return response.data;
  }

  // 热门搜索API
  async getPopularSearches(limit?: number): Promise<ApiResponse<{
    searches: string[];
    count: number;
  }>> {
    const response = await this.client.get('/cultural-wisdom/search/popular', {
      params: { limit }
    });
    return response.data;
  }

  async getBatchRecommendations(wisdomIds: string[], params?: {
    limit?: number;
    algorithm?: string;
  }): Promise<ApiResponse<{
    [wisdomId: string]: Array<{
      wisdom_id: string;
      title: string;
      author: string;
      category: string;
      school: string;
      summary: string;
      score: number;
      reason: string;
      view_count: number;
      like_count: number;
      created_at: string;
    }>;
  }>> {
    const response = await this.client.post('/cultural-wisdom/recommendations/batch', {
      wisdom_ids: wisdomIds,
      ...params
    });
    return response.data;
  }

  // 保留旧的推荐API以保持兼容性
  async getWisdomRecommendations(wisdomId: string): Promise<ApiResponse<{
    recommendations: Array<{
      wisdom_id: string;
      title: string;
      author: string;
      category: string;
      summary: string;
      relevance: number;
      reason: string;
    }>;
  }>> {
    const response = await this.client.get(`/cultural-wisdom/ai/${wisdomId}/recommend`);
    return response.data;
  }

  // 智慧内容管理API
  async getWisdomDetail(id: string): Promise<ApiResponse<CulturalWisdom>> {
    const response = await this.client.get(`/cultural-wisdom/${id}`);
    return response.data;
  }

  async createWisdom(data: WisdomData): Promise<ApiResponse<CulturalWisdom>> {
    const response = await this.client.post('/cultural-wisdom', data);
    return response.data;
  }

  async updateWisdom(id: string, data: WisdomData): Promise<ApiResponse<CulturalWisdom>> {
    const response = await this.client.put(`/cultural-wisdom/${id}`, data);
    return response.data;
  }

  async deleteWisdom(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/cultural-wisdom/${id}`);
    return response.data;
  }

  // 批量操作API
  async batchDeleteWisdom(ids: string[]): Promise<ApiResponse<void>> {
    const response = await this.client.post('/cultural-wisdom/batch-delete', { ids });
    return response.data;
  }

  // 智慧统计API
  async getWisdomStats(): Promise<ApiResponse<{
    totalCount: number;
    publishedCount: number;
    draftCount: number;
    categoryStats: Record<string, number>;
    schoolStats: Record<string, number>;
    difficultyStats: Record<number, number>;
    monthlyStats: Array<{ year: number; month: number; count: number }>;
  }>> {
    const response = await this.client.get('/cultural-wisdom/stats');
    return response.data;
  }

  // 高级搜索API
  async advancedSearchWisdom(params: {
    keyword?: string;
    category?: string;
    school?: string;
    author?: string;
    tags?: string[];
    difficulty?: string[];
    dateRange?: [string, string];
    status?: string;
    page?: number;
    pageSize?: number;
    sortBy?: string;
    sortOrder?: 'asc' | 'desc';
  }): Promise<ApiResponse<{ items: CulturalWisdom[]; total: number; page: number; pageSize: number }>> {
    const response = await this.client.get('/cultural-wisdom/advanced-search', { params });
    return response.data;
  }

  // 分类管理API
  async getCategories(): Promise<ApiResponse<Category[]>> {
    const response = await this.client.get('/cultural-wisdom/categories');
    return response.data;
  }

  async getCategoryById(id: string): Promise<ApiResponse<Category>> {
    const response = await this.client.get(`/cultural-wisdom/categories/${id}`);
    return response.data;
  }

  async createCategory(data: CategoryData): Promise<ApiResponse<Category>> {
    const response = await this.client.post('/cultural-wisdom/categories', data);
    return response.data;
  }

  async updateCategory(id: string, data: CategoryData): Promise<ApiResponse<Category>> {
    const response = await this.client.put(`/cultural-wisdom/categories/${id}`, data);
    return response.data;
  }

  async deleteCategory(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/cultural-wisdom/categories/${id}`);
    return response.data;
  }

  // 标签管理API
  async getTags(): Promise<ApiResponse<Tag[]>> {
    const response = await this.client.get('/cultural-wisdom/tags');
    return response.data;
  }

  async getTagById(id: string): Promise<ApiResponse<Tag>> {
    const response = await this.client.get(`/cultural-wisdom/tags/${id}`);
    return response.data;
  }

  async createTag(data: TagData): Promise<ApiResponse<Tag>> {
    const response = await this.client.post('/cultural-wisdom/tags', data);
    return response.data;
  }

  async updateTag(id: string, data: TagData): Promise<ApiResponse<Tag>> {
    const response = await this.client.put(`/cultural-wisdom/tags/${id}`, data);
    return response.data;
  }

  async deleteTag(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/cultural-wisdom/tags/${id}`);
    return response.data;
  }

  // 学派管理API
  async getSchools(): Promise<ApiResponse<Array<{ id: string; name: string; description?: string }>>> {
    const response = await this.client.get('/cultural-wisdom/schools');
    return response.data;
  }

  // 用户管理相关 API
  async getUsers(params?: UserQueryParams): Promise<ApiResponse<User[]>> {
    const response = await this.client.get('/api/admin/users', { params });
    return response.data;
  }

  async getUserStats(): Promise<ApiResponse<UserStats>> {
    const response = await this.client.get('/api/admin/users/stats');
    return response.data;
  }

  async createUser(userData: UserData): Promise<ApiResponse<User>> {
    const response = await this.client.post('/api/admin/users', userData);
    return response.data;
  }

  async updateUser(userId: string, userData: UserData): Promise<ApiResponse<User>> {
    const response = await this.client.put(`/api/admin/users/${userId}`, userData);
    return response.data;
  }

  async deleteUser(userId: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/api/admin/users/${userId}`);
    return response.data;
  }

  async batchDeleteUsers(userIds: string[]): Promise<ApiResponse<{ success: number; failed: number }>> {
    const response = await this.client.post('/api/admin/users/batch-delete', { userIds });
    return response.data;
  }

  async updateUserStatus(userId: string, status: string): Promise<ApiResponse<User>> {
    const response = await this.client.put(`/api/admin/users/${userId}/status`, { status });
    return response.data;
  }

  async updateUserRole(userId: string, role: string): Promise<ApiResponse<User>> {
    const response = await this.client.put(`/api/admin/users/${userId}/role`, { role });
    return response.data;
  }

  // 数据分析相关 API
  async getAnalyticsData(params?: QueryParams): Promise<ApiResponse<AnalyticsData>> {
    const response = await this.client.get('/api/admin/analytics', { params });
    return response.data;
  }

  async exportAnalyticsReport(params?: QueryParams): Promise<ApiResponse<Blob>> {
    const response = await this.client.get('/api/admin/analytics/export', { params });
    return response.data;
  }

  // 系统设置相关 API
  async getSystemConfig(): Promise<ApiResponse<SystemConfig>> {
    const response = await this.client.get('/api/admin/system/config');
    return response.data;
  }

  async updateSystemConfig(data: SystemConfig): Promise<ApiResponse<SystemConfig>> {
    const response = await this.client.put('/api/admin/system/config', data);
    return response.data;
  }

  async getSEOConfig(): Promise<ApiResponse<SEOConfig>> {
    const response = await this.client.get('/api/admin/system/seo');
    return response.data;
  }

  async updateSEOConfig(data: SEOConfig): Promise<ApiResponse<SEOConfig>> {
    const response = await this.client.put('/api/admin/system/seo', data);
    return response.data;
  }

  async getCacheConfig(): Promise<ApiResponse<CacheConfig>> {
    const response = await this.client.get('/api/admin/system/cache/config');
    return response.data;
  }

  async updateCacheConfig(data: CacheConfig): Promise<ApiResponse<CacheConfig>> {
    const response = await this.client.put('/api/admin/system/cache/config', data);
    return response.data;
  }

  async getCacheStats(): Promise<ApiResponse<Record<string, unknown>>> {
    const response = await this.client.get('/api/admin/system/cache/stats');
    return response.data;
  }

  async clearCache(): Promise<ApiResponse<void>> {
    const response = await this.client.delete('/api/admin/system/cache');
    return response.data;
  }

  async testEmailConfig(): Promise<ApiResponse<{ success: boolean; message: string }>> {
    const response = await this.client.post('/api/admin/system/test-email');
    return response.data;
  }

  // 通知系统相关 API
  async getNotifications(params?: QueryParams): Promise<ApiResponse<Notification[]>> {
    const response = await this.client.get('/api/admin/notifications', { params });
    return response.data;
  }

  async createNotification(data: NotificationData): Promise<ApiResponse<Notification>> {
    const response = await this.client.post('/api/admin/notifications', data);
    return response.data;
  }

  async updateNotification(id: string, data: NotificationData): Promise<ApiResponse<Notification>> {
    const response = await this.client.put(`/api/admin/notifications/${id}`, data);
    return response.data;
  }

  async deleteNotification(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/api/admin/notifications/${id}`);
    return response.data;
  }

  async batchDeleteNotifications(ids: string[]): Promise<ApiResponse<void>> {
    const response = await this.client.post('/api/admin/notifications/batch-delete', { ids });
    return response.data;
  }

  async sendNotification(id: string): Promise<ApiResponse<{ success: boolean; message: string }>> {
    const response = await this.client.post(`/api/admin/notifications/${id}/send`);
    return response.data;
  }

  async getUserNotifications(params?: QueryParams): Promise<ApiResponse<Notification[]>> {
    const response = await this.client.get('/api/admin/notifications/users', { params });
    return response.data;
  }

  async getNotificationStats(): Promise<ApiResponse<Record<string, number>>> {
    const response = await this.client.get('/api/admin/notifications/stats');
    return response.data;
  }

  // 内容审核相关 API
  async getReviewItems(params?: QueryParams): Promise<ApiResponse<ReviewItem[]>> {
    const response = await this.client.get('/api/admin/review/items', { params });
    return response.data;
  }

  async getReviewStats(): Promise<ApiResponse<Record<string, number>>> {
    const response = await this.client.get('/api/admin/review/stats');
    return response.data;
  }

  async reviewContent(id: string, data: ReviewData): Promise<ApiResponse<ReviewItem>> {
    const response = await this.client.post(`/api/admin/review/items/${id}/review`, data);
    return response.data;
  }

  async batchReviewContent(data: BatchReviewData): Promise<ApiResponse<{ success: number; failed: number }>> {
    const response = await this.client.post('/api/admin/review/items/batch-review', data);
    return response.data;
  }

  async getReviewHistory(id: string): Promise<ApiResponse<ReviewItem[]>> {
    const response = await this.client.get(`/api/admin/review/items/${id}/history`);
    return response.data;
  }

  async exportReviewReport(params?: QueryParams): Promise<ApiResponse<Blob>> {
    const response = await this.client.get('/api/admin/review/export', { params });
    return response.data;
  }

  // 收藏功能API
  async addFavorite(wisdomId: string): Promise<ApiResponse<{ message: string }>> {
    const response = await this.client.post('/cultural-wisdom/favorites', {
      wisdom_id: wisdomId
    });
    return response.data;
  }

  async removeFavorite(wisdomId: string): Promise<ApiResponse<{ message: string }>> {
    const response = await this.client.delete(`/cultural-wisdom/favorites/${wisdomId}`);
    return response.data;
  }

  async getUserFavorites(params?: {
    page?: number;
    limit?: number;
  }): Promise<ApiResponse<{
    favorites: Array<{
      wisdom_id: string;
      title: string;
      author: string;
      category: string;
      school: string;
      summary: string;
      created_at: string;
      favorited_at: string;
    }>;
    total: number;
    page: number;
    limit: number;
  }>> {
    const response = await this.client.get('/cultural-wisdom/favorites', { params });
    return response.data;
  }

  async checkFavoriteStatus(wisdomId: string): Promise<ApiResponse<{ is_favorited: boolean }>> {
    const response = await this.client.get(`/cultural-wisdom/favorites/${wisdomId}/status`);
    return response.data;
  }

  // 笔记功能API
  async createNote(wisdomId: string, data: {
    title?: string;
    content: string;
    is_private?: boolean;
    tags?: string[];
  }): Promise<ApiResponse<{
    id: string;
    wisdom_id: string;
    title: string;
    content: string;
    is_private: boolean;
    tags: string[];
    created_at: string;
    updated_at: string;
  }>> {
    const response = await this.client.post('/cultural-wisdom/notes', {
      wisdom_id: wisdomId,
      ...data
    });
    return response.data;
  }

  async updateNote(wisdomId: string, data: {
    title?: string;
    content?: string;
    is_private?: boolean;
    tags?: string[];
  }): Promise<ApiResponse<{
    id: string;
    wisdom_id: string;
    title: string;
    content: string;
    is_private: boolean;
    tags: string[];
    created_at: string;
    updated_at: string;
  }>> {
    const response = await this.client.put(`/cultural-wisdom/notes/${wisdomId}`, data);
    return response.data;
  }

  async getNote(wisdomId: string): Promise<ApiResponse<{
    id: string;
    wisdom_id: string;
    title: string;
    content: string;
    is_private: boolean;
    tags: string[];
    created_at: string;
    updated_at: string;
  }>> {
    const response = await this.client.get(`/cultural-wisdom/notes/${wisdomId}`);
    return response.data;
  }

  async getUserNotes(params?: {
    page?: number;
    limit?: number;
    search?: string;
    tags?: string[];
  }): Promise<ApiResponse<{
    notes: Array<{
      id: string;
      wisdom_id: string;
      wisdom_title: string;
      title: string;
      content: string;
      is_private: boolean;
      tags: string[];
      created_at: string;
      updated_at: string;
    }>;
    total: number;
    page: number;
    limit: number;
  }>> {
    const response = await this.client.get('/cultural-wisdom/notes', { params });
    return response.data;
  }

  async deleteNote(wisdomId: string): Promise<ApiResponse<{ message: string }>> {
    const response = await this.client.delete(`/cultural-wisdom/notes/${wisdomId}`);
    return response.data;
  }
}

export const apiClient = new ApiClient();
export const authAPI = apiClient; // 为了向后兼容
export default apiClient;