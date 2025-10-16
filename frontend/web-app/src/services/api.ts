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
  private inFlight: Map<string, Promise<AxiosResponse<any>>> = new Map();

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
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
      (response: AxiosResponse<ApiResponse>) => {
        // 可选：成功请求上报（仅非GET），通过 localStorage 开关控制
        try {
          const logSuccess = localStorage.getItem('log_success') === 'true';
          const method = String(response.config?.method || '').toLowerCase();
          if (logSuccess && method && method !== 'get') {
            const url = response.config?.url || '';
            const status = response.status;
            this.client.post('/system/logs', {
              level: 'info',
              message: `HTTP ${status} ${method.toUpperCase()} ${url}`,
              module: 'frontend/api',
              extra: { url, method, status },
            }, { headers: { 'x-log-skip': 'true' } }).catch(() => {});
          }
        } catch (_) {}
        return response;
      },
      (error) => {
        // 跳过由日志上报自身触发的拦截器处理，防止递归
        const skipLog = !!(error?.config?.headers && (error.config.headers as any)['x-log-skip']);

        // 将 HTTP 错误上报到系统日志（除非被显式跳过）
        if (!skipLog) {
          try {
            const cfg = error?.config || {};
            const method = String(cfg.method || '').toUpperCase();
            const status = error?.response?.status;
            const url = cfg.url || '';
            const serverMsg = (error?.response?.data && (error.response.data.message || error.response.data.error)) || '';
            const msg = `HTTP ${status ?? 'ERR'} ${method} ${url}: ${serverMsg || error.message || 'request failed'}`;
            // 通过专用头避免递归上报
            this.client.post('/system/logs', {
              level: 'error',
              message: msg,
              module: 'frontend/api',
              extra: {
                url,
                method,
                status,
                response: error?.response?.data,
              },
            }, { headers: { 'x-log-skip': 'true' } }).catch(() => {});
          } catch (_) {
            // 忽略上报错误
          }
        }
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

  // 公开的HTTP方法
  async get<T = any>(url: string, config?: any): Promise<AxiosResponse<T>> {
    const key = this.requestKey('GET', url, config);
    const existing = this.inFlight.get(key) as Promise<AxiosResponse<T>> | undefined;
    if (existing) return existing;
    const req = this.client.get<T>(url, config).finally(() => {
      this.inFlight.delete(key);
    });
    this.inFlight.set(key, req as any);
    return req;
  }

  async post<T = any>(url: string, data?: any, config?: any): Promise<AxiosResponse<T>> {
    const key = this.requestKey('POST', url, config, data);
    const existing = this.inFlight.get(key) as Promise<AxiosResponse<T>> | undefined;
    if (existing) return existing;
    const req = this.client.post<T>(url, data, config).finally(() => {
      this.inFlight.delete(key);
    });
    this.inFlight.set(key, req as any);
    return req;
  }

  async put<T = any>(url: string, data?: any, config?: any): Promise<AxiosResponse<T>> {
    const key = this.requestKey('PUT', url, config, data);
    const existing = this.inFlight.get(key) as Promise<AxiosResponse<T>> | undefined;
    if (existing) return existing;
    const req = this.client.put<T>(url, data, config).finally(() => {
      this.inFlight.delete(key);
    });
    this.inFlight.set(key, req as any);
    return req;
  }

  async delete<T = any>(url: string, config?: any): Promise<AxiosResponse<T>> {
    const key = this.requestKey('DELETE', url, config);
    const existing = this.inFlight.get(key) as Promise<AxiosResponse<T>> | undefined;
    if (existing) return existing;
    const req = this.client.delete<T>(url, config).finally(() => {
      this.inFlight.delete(key);
    });
    this.inFlight.set(key, req as any);
    return req;
  }

  async patch<T = any>(url: string, data?: any, config?: any): Promise<AxiosResponse<T>> {
    const key = this.requestKey('PATCH', url, config, data);
    const existing = this.inFlight.get(key) as Promise<AxiosResponse<T>> | undefined;
    if (existing) return existing;
    const req = this.client.patch<T>(url, data, config).finally(() => {
      this.inFlight.delete(key);
    });
    this.inFlight.set(key, req as any);
    return req;
  }

  private requestKey(method: string, url: string, config?: any, data?: any): string {
    const params = config?.params ? JSON.stringify(config.params) : '';
    const payload = data ? JSON.stringify(data) : '';
    // 使用 baseURL + 相对路径形成稳定 key
    const fullUrl = (this.client.defaults.baseURL || '') + url;
    return `${method}|${fullUrl}|${params}|${payload}`;
  }

  // 用户相关API
  async login(username: string, password: string): Promise<ApiResponse<{ user: User; token: string }>> {
    const response = await this.client.post('/auth/login', { username, password });
    
    // 后端返回格式: { data: { token, user_id, username, email, role, roles, permissions, ... }, success: true }
    if (response.data.success && response.data.data && response.data.data.token) {
      const userData = response.data.data;
      const userRole = userData.role || 'user';
      return {
        success: true,
        data: {
          user: {
            id: userData.user_id,
            username: userData.username,
            email: userData.email,
            role: userRole,
            isAdmin: userRole.toLowerCase() === 'admin' || userRole.toLowerCase() === 'super_admin' || userData.isAdmin === true,
            avatar: '', // 默认头像
            bio: '', // 默认简介
            created_at: userData.created_at || new Date().toISOString(),
            updated_at: userData.updated_at || new Date().toISOString(),
            roles: userData.roles || [userRole],
            permissions: userData.permissions || []
          } as User,
          token: userData.token
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
      const response = await this.client.get('/auth/me');
      
      // 处理后端返回的数据格式
      if (response.data.success && response.data.data) {
        const userData = response.data.data;
        const userRole = userData.role || 'user';
        return {
          success: true,
          data: {
            id: userData.id || userData.user_id,
            username: userData.username,
            email: userData.email,
            display_name: userData.display_name || userData.displayName || userData.username,
            role: userRole,
            roles: userData.roles || [userRole],
            permissions: userData.permissions || [],
            isAdmin: userRole.toLowerCase() === 'admin' || userRole.toLowerCase() === 'super_admin' || userData.isAdmin === true,
            avatar: userData.avatar || userData.avatar_url || '',
            bio: userData.bio || '',
            status: userData.status || 'active',
            created_at: userData.created_at || userData.createdAt,
            updated_at: userData.updated_at || userData.updatedAt,
            last_login: userData.last_login || userData.lastLogin
          } as User
        };
      } else {
        console.error('getCurrentUser API response error:', response.data);
        return {
          success: false,
          data: {} as User,
          error: response.data.message || '获取用户信息失败'
        };
      }
    } catch (error) {
      console.error('getCurrentUser API error:', error);
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
    const response = await this.client.get('/admin/users', { params });
    const body = response.data as ApiResponse<any> | any;

    // 后端可能返回形如 { success: true, data: { users: [...], pagination: {...} } }
    // 也可能直接返回 { success: true, data: [...] }
    // 统一规范为返回用户数组在 data 字段中，避免 Table 传入非数组导致报错（rawData.some 不是函数）。
    const usersArray = Array.isArray(body?.data)
      ? body.data
      : Array.isArray(body?.data?.users)
        ? body.data.users
        : Array.isArray(body?.users)
          ? body.users
          : null;

    if (usersArray) {
      return { success: true, data: usersArray } as ApiResponse<User[]>;
    }

    // 回退：保持原始结构（用于兼容不同返回格式），但这可能会导致调用方需要自行解包
    return response.data as ApiResponse<User[]>;
  }

  async getUserStats(): Promise<ApiResponse<UserStats>> {
    const response = await this.client.get('/admin/users/stats');
    const body = response.data as ApiResponse<any> | any;
    const raw = body?.data ?? body;
    const normalized: UserStats = {
      totalUsers: Number(raw?.totalUsers ?? raw?.total_users ?? 0),
      activeUsers: Number(raw?.activeUsers ?? raw?.active_users ?? 0),
      adminUsers: Number(raw?.adminUsers ?? raw?.admin_users ?? 0),
      newUsersToday: Number(raw?.newUsersToday ?? raw?.new_users_today ?? raw?.new_users ?? 0),
      onlineUsers: Number(raw?.onlineUsers ?? raw?.online_users ?? 0),
    };
    return { success: !!(body?.success ?? true), data: normalized };
  }

  async createUser(userData: UserData): Promise<ApiResponse<User>> {
    const response = await this.client.post('/admin/users', userData);
    return response.data;
  }

  async updateUser(userId: string, userData: UserData): Promise<ApiResponse<User>> {
    const response = await this.client.put(`/admin/users/${userId}`, userData);
    return response.data;
  }

  async deleteUser(userId: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/admin/users/${userId}`);
    return response.data;
  }

  async batchDeleteUsers(userIds: string[]): Promise<ApiResponse<{ success: number; failed: number }>> {
    const response = await this.client.post('/admin/users/batch-delete', { userIds });
    return response.data;
  }

  async updateUserStatus(userId: string, status: string): Promise<ApiResponse<User>> {
    const response = await this.client.put(`/admin/users/${userId}/status`, { status });
    return response.data;
  }

  async updateUserRole(userId: string, role: string): Promise<ApiResponse<User>> {
    const response = await this.client.put(`/admin/users/${userId}/role`, { role });
    return response.data;
  }

  // 数据分析相关 API
  async getAnalyticsData(params?: QueryParams): Promise<ApiResponse<AnalyticsData>> {
    const response = await this.client.get('/admin/analytics', { params });
    return response.data;
  }

  async exportAnalyticsReport(params?: QueryParams): Promise<ApiResponse<Blob>> {
    const response = await this.client.get('/admin/analytics/export', { params });
    return response.data;
  }

  // 系统设置相关 API
  // 辅助方法：将后端返回的配置数组映射为对象，并解析值
  private normalizeConfigKey(key: string): string {
    return (key || '').replace(/^(system[._]|seo[._]|cache[._])/, '');
  }

  private parseConfigValue(value: any): any {
    if (typeof value !== 'string') return value;
    const v = value.trim();
    if (v === 'true' || v === 'false') return v === 'true';
    if (/^-?\d+(?:\.\d+)?$/.test(v)) return Number(v);
    try {
      return JSON.parse(v);
    } catch {
      return v;
    }
  }

  private configsToObject(configs: Array<{ key: string; value: any }>): Record<string, any> {
    const obj: Record<string, any> = {};
    (configs || []).forEach((c) => {
      const k = this.normalizeConfigKey(c.key);
      obj[k] = this.parseConfigValue(c.value);
    });
    return obj;
  }

  async getSystemConfig(): Promise<ApiResponse<SystemConfig>> {
    const response = await this.client.get('/system/config', { params: { category: 'system' } });
    const data = (response.data as any) || {};
    const configs = data.configs || [];
    const mapped = this.configsToObject(configs);
    return { success: true, data: mapped } as ApiResponse<SystemConfig>;
  }

  async updateSystemConfig(data: SystemConfig): Promise<ApiResponse<SystemConfig>> {
    // 后端期望 { configs: [{ key, value }] }
    const entries = Object.entries(data || {}).filter(([, value]) => value !== undefined && value !== null);
    const payload = {
      configs: entries.map(([key, value]) => ({
        key: `system.${key}`,
        value: Array.isArray(value) || typeof value === 'object' ? JSON.stringify(value) : String(value)
      }))
    };
    const response = await this.client.put('/system/config', payload);
    return { success: true, data } as ApiResponse<SystemConfig>;
  }

  async getSEOConfig(): Promise<ApiResponse<SEOConfig>> {
    // 从系统配置中读取分类 seo 的键值
    const response = await this.client.get('/system/config', { params: { category: 'seo' } });
    const data = (response.data as any) || {};
    const configs = data.configs || [];
    const obj = this.configsToObject(configs);
    const seo: SEOConfig = {
      metaTitle: obj.metaTitle || obj.title || '',
      metaDescription: obj.metaDescription || obj.description || '',
      metaKeywords: obj.metaKeywords || obj.keywords || '',
      ogTitle: obj.ogTitle || '',
      ogDescription: obj.ogDescription || '',
      ogImage: obj.ogImage || '',
      twitterCard: obj.twitterCard || 'summary',
      robotsTxt: obj.robotsTxt || '',
      sitemapEnabled: obj.sitemapEnabled ?? false,
      analyticsCode: obj.analyticsCode || ''
    };
    return { success: true, data: seo };
  }

  async updateSEOConfig(data: SEOConfig): Promise<ApiResponse<SEOConfig>> {
    const entries = Object.entries(data || {}).filter(([, value]) => value !== undefined && value !== null);
    const payload = {
      configs: entries.map(([key, value]) => ({
        key: `seo.${key}`,
        value: Array.isArray(value) || typeof value === 'object' ? JSON.stringify(value) : String(value)
      }))
    };
    await this.client.put('/system/config', payload);
    return { success: true, data };
  }

  async getCacheConfig(): Promise<ApiResponse<CacheConfig>> {
    const response = await this.client.get('/system/config', { params: { category: 'cache' } });
    const data = (response.data as any) || {};
    const configs = data.configs || [];
    const obj = this.configsToObject(configs);
    const cache: CacheConfig = {
      enabled: obj.enabled ?? false,
      ttl: obj.ttl ?? 3600,
      maxSize: obj.maxSize ?? 256,
      strategy: obj.strategy || 'lru'
    };
    return { success: true, data: cache };
  }

  async updateCacheConfig(data: CacheConfig): Promise<ApiResponse<CacheConfig>> {
    const entries = Object.entries(data || {}).filter(([, value]) => value !== undefined && value !== null);
    const payload = {
      configs: entries.map(([key, value]) => ({
        key: `cache.${key}`,
        value: Array.isArray(value) || typeof value === 'object' ? JSON.stringify(value) : String(value)
      }))
    };
    await this.client.put('/system/config', payload);
    return { success: true, data };
  }

  async getCacheStats(): Promise<ApiResponse<Record<string, unknown>>> {
    const response = await this.client.get('/system/cache/stats');
    const data = (response.data as any) || {};
    const stats = data.cache_stats || data || {};
    // 标准化返回结构
    const parsed = {
      hitRate: Number(stats.hit_rate ?? stats.hitRate ?? 0),
      totalRequests: Number(stats.total_requests ?? stats.totalRequests ?? stats.total_keys ?? 0),
      cacheSize: Number((stats.memory_usage || '').toString().replace(/[^\d.]/g, '')) || Number(stats.cacheSize ?? 0),
      memoryUsage: Number(stats.memory_usage_percent ?? stats.memoryUsage ?? 0)
    };
    return { success: true, data: parsed };
  }

  async clearCache(): Promise<ApiResponse<void>> {
    await this.client.delete('/system/cache');
    return { success: true, data: undefined } as ApiResponse<void>;
  }

  // 系统日志 / 数据库 / 问题跟踪 API（新增）
  async getSystemLogs(params?: { level?: string; start?: string; end?: string; limit?: number; source?: string }): Promise<ApiResponse<Array<{ timestamp?: string; level?: string; source?: string; message?: string; user_id?: string; ip?: string; user_agent?: string; extra?: any }>>> {
    // 后端期望的参数为 page/limit/level/module，兼容前端传入的 source/start/end
    const queryParams: Record<string, any> = {
      page: 1,
      limit: params?.limit ?? 100,
    };
    if (params?.level) queryParams.level = params.level;
    if (params?.source) queryParams.module = params.source;
    // 后端当前未支持按时间范围筛选，保留 start/end 以便未来兼容
    if (params?.start) queryParams.start = params.start;
    if (params?.end) queryParams.end = params.end;

    const response = await this.client.get('/system/logs', { params: queryParams });
    const payload = (response.data as any) || {};
    const rawLogs: Array<Record<string, any>> = payload.logs || payload.data || payload || [];
    const parseExtra = (e: any) => {
      if (typeof e === 'string') {
        try { return JSON.parse(e); } catch { return e; }
      }
      return e ?? undefined;
    };
    const normalized = Array.isArray(rawLogs)
      ? rawLogs.map((l) => ({
          timestamp: l.timestamp || l.time || l.created_at || l.createdAt,
          level: (l.level || l.Level || '').toString().toUpperCase() || undefined,
          source: l.source || l.module,
          message: l.message,
          user_id: l.user_id || l.userId,
          ip: l.ip,
          user_agent: l.user_agent || l.userAgent,
          extra: parseExtra(l.extra),
        }))
      : [];
    return { success: true, data: normalized };
  }


  // 新增：前端/客户端写入系统日志
  async createSystemLog(entry: { level?: string; message: string; module?: string; extra?: any; timestamp?: string }): Promise<ApiResponse<{ created: boolean; log?: Record<string, any> }>> {
    const response = await this.client.post('/system/logs', entry, { headers: { 'x-log-skip': 'true' } });
    const data = (response.data as any) || {};
    const created = data.created === true || data.success === true;
    return { success: created, data: { created, log: data.log } };
  }

  async getSystemLogStats(): Promise<ApiResponse<Record<string, number>>> {
    const response = await this.client.get('/system/logs/stats');
    const data = (response.data as any) || {};
    const statsArr: any[] = Array.isArray(data.stats) ? data.stats : (Array.isArray(data.data) ? data.data : []);
    const result: Record<string, number> = { ERROR: 0, WARN: 0, INFO: 0, DEBUG: 0 };
    statsArr.forEach((s) => {
      const lvl = String(s.level || s.Level || '').toUpperCase();
      const cnt = Number(s.count || s.Count || 0);
      if (lvl) result[lvl] = (result[lvl] || 0) + cnt;
    });
    return { success: true, data: result };
  }

  // 系统备份管理API
  async createBackup(name: string, description?: string): Promise<ApiResponse<{ message: string; backup: Record<string, any> }>> {
    const response = await this.client.post('/system/backup', { name, description });
    return response.data;
  }

  async getBackups(params?: { page?: number; limit?: number }): Promise<ApiResponse<{
    backups: Array<Record<string, any>>;
    total: number;
    page: number;
    limit: number;
    pages: number;
  }>> {
    const response = await this.client.get('/system/backups', { params });
    return response.data;
  }

  async restoreBackup(id: number | string): Promise<ApiResponse<{ message: string; backup: Record<string, any> }>> {
    const response = await this.client.post(`/system/restore/${id}`);
    return response.data;
  }

  async getDatabaseStats(): Promise<ApiResponse<Record<string, any>>> {
    const response = await this.client.get('/system/database/stats');
    const data = (response.data as any) || {};
    const stats = data.database_stats || data.data || data;
    return { success: true, data: stats };
  }

  async optimizeDatabase(): Promise<ApiResponse<{ message: string }>> {
    const response = await this.client.post('/system/database/optimize', {});
    const data = (response.data as any) || {};
    return { success: true, data: { message: data.message || 'Database optimization started' } };
  }

  async listDatabaseTables(): Promise<ApiResponse<string[]>> {
    const response = await this.client.get('/system/database/tables');
    const payload = (response.data as any) || {};
    const raw = payload.tables || payload.data || [];
    const tables: string[] = Array.isArray(raw)
      ? raw.map((t: any) => {
          if (typeof t === 'string') return t;
          // 兼容后端返回 {schema, name}
          const schema = (t && (t.schema || t.table_schema)) || '';
          const name = (t && (t.name || t.table_name)) || '';
          if (schema && name) return `${schema}.${name}`;
          if (name) return name;
          // 兜底：字符串化
          try { return String(t); } catch { return ''; }
        })
      : [];
    return { success: true, data: tables.filter(Boolean) };
  }

  async listDatabaseTablesPaged(params?: { page?: number; limit?: number; schema?: string }): Promise<ApiResponse<{ items: string[]; total: number; page: number; limit: number; pages: number }>> {
    const response = await this.client.get('/system/database/tables', {
      params: {
        page: params?.page ?? 1,
        limit: params?.limit ?? 50,
        ...(params?.schema ? { schema: params.schema } : {}),
      },
    });
    const payload = (response.data as any) || {};
    const raw = payload.tables || payload.data || [];
    const items: string[] = Array.isArray(raw)
      ? raw.map((t: any) => {
          if (typeof t === 'string') return t;
          const schema = (t && (t.schema || t.table_schema)) || '';
          const name = (t && (t.name || t.table_name)) || '';
          if (schema && name) return `${schema}.${name}`;
          if (name) return name;
          try { return String(t); } catch { return ''; }
        }).filter(Boolean)
      : [];
    const total = Number(payload.total ?? items.length) || 0;
    const page = Number(payload.page ?? (params?.page ?? 1)) || 1;
    const limit = Number(payload.limit ?? (params?.limit ?? 50)) || 50;
    const pages = Number(payload.pages ?? Math.ceil(total / Math.max(limit, 1))) || Math.ceil(total / Math.max(limit, 1));
    return { success: true, data: { items, total, page, limit, pages } };
  }

  async listDatabaseSchemasPaged(params?: { page?: number; limit?: number }): Promise<ApiResponse<{ items: string[]; total: number; page: number; limit: number; pages: number }>> {
    const response = await this.client.get('/system/database/schemas', {
      params: {
        page: params?.page ?? 1,
        limit: params?.limit ?? 100,
      },
    });
    const payload = (response.data as any) || {};
    const raw = payload.schemas || payload.data || [];
    const items: string[] = Array.isArray(raw)
      ? raw.map((s: any) => (typeof s === 'string' ? s : String(s))).filter(Boolean)
      : [];
    const total = Number(payload.total ?? items.length) || 0;
    const page = Number(payload.page ?? (params?.page ?? 1)) || 1;
    const limit = Number(payload.limit ?? (params?.limit ?? 100)) || 100;
    const pages = Number(payload.pages ?? Math.ceil(total / Math.max(limit, 1))) || Math.ceil(total / Math.max(limit, 1));
    return { success: true, data: { items, total, page, limit, pages } };
  }

  async getTableColumns(tableName: string): Promise<ApiResponse<Array<{ name: string; type?: string; nullable?: boolean }>>> {
    // 支持传入 "schema.table"，自动拆分并传递 schema 查询参数（PostgreSQL 需要）
    let tbl = tableName;
    let schemaParam: string | undefined;
    const dotIdx = tableName.indexOf('.');
    if (dotIdx > 0) {
      schemaParam = tableName.slice(0, dotIdx);
      tbl = tableName.slice(dotIdx + 1);
    }

    const response = await this.client.get(`/system/database/tables/${encodeURIComponent(tbl)}/columns`, {
      params: schemaParam ? { schema: schemaParam } : undefined,
    });
    const data = (response.data as any) || {};
    const columns = data.columns || data.data || [];
    return { success: true, data: columns };
  }

  async runReadOnlyQuery(sql: string, options?: { maxRows?: number }): Promise<ApiResponse<{ columns: string[]; rows: Array<Record<string, any>> }>> {
    // 与后端约定字段：{ query: string, max_rows?: number }
    const response = await this.client.post('/system/database/query', { query: sql, ...(options?.maxRows ? { max_rows: options.maxRows } : {}) });
    const data = (response.data as any) || {};
    const rows = data.rows || data.data || [];
    const columns = data.columns || (rows[0] ? Object.keys(rows[0]) : []);
    return { success: true, data: { columns, rows } };
  }

  async detectIssues(payload?: { lookback_minutes?: number; severity?: string }): Promise<ApiResponse<Array<any>>> {
    // 后端路由为 GET /system/issues/detect，使用查询参数
    const response = await this.client.get('/system/issues/detect', { params: payload || {} });
    const data = (response.data as any) || {};
    const issues = data.issues || data.data || [];
    const normalized = Array.isArray(issues)
      ? issues.map((i: any) => ({
          ...i,
          timestamp: i.timestamp || i.last_seen || i.first_seen || i.time,
          source: i.source || i.category,
        }))
      : [];
    return { success: true, data: normalized };
  }

  async triggerIssueAlert(issueId: string, channels: string[] = ['internal']): Promise<ApiResponse<{ sent: boolean }>> {
    const response = await this.client.post('/system/issues/alert', { issue_id: issueId, channels });
    const data = (response.data as any) || {};
    return { success: true, data: { sent: !!(data.sent || data.success) } };
  }

  // 通知系统相关 API
  async getNotifications(params?: QueryParams): Promise<ApiResponse<Notification[]>> {
    const response = await this.client.get('/admin/notifications', { params });
    return response.data;
  }

  async createNotification(data: NotificationData): Promise<ApiResponse<Notification>> {
    const response = await this.client.post('/admin/notifications', data);
    return response.data;
  }

  async updateNotification(id: string, data: NotificationData): Promise<ApiResponse<Notification>> {
    const response = await this.client.put(`/admin/notifications/${id}`, data);
    return response.data;
  }

  async deleteNotification(id: string): Promise<ApiResponse<void>> {
    const response = await this.client.delete(`/admin/notifications/${id}`);
    return response.data;
  }

  async batchDeleteNotifications(ids: string[]): Promise<ApiResponse<void>> {
    const response = await this.client.post('/admin/notifications/batch-delete', { ids });
    return response.data;
  }

  async sendNotification(id: string): Promise<ApiResponse<{ success: boolean; message: string }>> {
    const response = await this.client.post(`/admin/notifications/${id}/send`);
    return response.data;
  }

  async getUserNotifications(params?: QueryParams): Promise<ApiResponse<Notification[]>> {
    const response = await this.client.get('/admin/notifications/users', { params });
    return response.data;
  }

  async getNotificationStats(): Promise<ApiResponse<Record<string, number>>> {
    const response = await this.client.get('/admin/notifications/stats');
    return response.data;
  }

  // 内容审核相关 API
  async getReviewItems(params?: QueryParams): Promise<ApiResponse<ReviewItem[]>> {
    const response = await this.client.get('/admin/review/items', { params });
    return response.data;
  }

  async getReviewStats(): Promise<ApiResponse<Record<string, number>>> {
    const response = await this.client.get('/admin/review/stats');
    return response.data;
  }

  async reviewContent(id: string, data: ReviewData): Promise<ApiResponse<ReviewItem>> {
    const response = await this.client.post(`/admin/review/items/${id}/review`, data);
    return response.data;
  }

  async batchReviewContent(data: BatchReviewData): Promise<ApiResponse<{ success: number; failed: number }>> {
    const response = await this.client.post('/admin/review/items/batch-review', data);
    return response.data;
  }

  async getReviewHistory(id: string): Promise<ApiResponse<ReviewItem[]>> {
    const response = await this.client.get(`/admin/review/items/${id}/history`);
    return response.data;
  }

  async exportReviewReport(params?: QueryParams): Promise<ApiResponse<Blob>> {
    const response = await this.client.get('/admin/review/export', { params });
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