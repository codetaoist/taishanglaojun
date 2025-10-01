import axios, { AxiosInstance, AxiosResponse, AxiosError } from 'axios';

// API响应接口
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  errors?: Record<string, string[]>;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// 错误响应接口
export interface ApiError {
  message: string;
  status?: number;
  errors?: Record<string, string[]>;
}

// API配置
const API_CONFIG = {
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
};

// 创建axios实例
const api: AxiosInstance = axios.create(API_CONFIG);

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
  (config) => {
    // 从localStorage获取token
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    // 添加请求时间戳
    config.headers['X-Request-Time'] = new Date().toISOString();
    
    // 添加用户ID（如果存在）
    const userId = localStorage.getItem('userId');
    if (userId) {
      config.headers['X-User-ID'] = userId;
    }
    
    return config;
  },
  (error) => {
    console.error('Request interceptor error:', error);
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误和token刷新
api.interceptors.response.use(
  (response: AxiosResponse) => {
    // 检查响应中的新token
    const newToken = response.headers['x-new-token'];
    if (newToken) {
      localStorage.setItem('authToken', newToken);
    }
    
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as any;
    
    // 处理401未授权错误
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      // 尝试刷新token
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          const response = await axios.post(`${API_CONFIG.baseURL}/auth/refresh`, {
            refresh_token: refreshToken,
          });
          
          const { access_token, refresh_token: newRefreshToken } = response.data;
          localStorage.setItem('authToken', access_token);
          localStorage.setItem('refreshToken', newRefreshToken);
          
          // 重新发送原始请求
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        } catch (refreshError) {
          // 刷新失败，清除token并跳转到登录页
          localStorage.removeItem('authToken');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('userId');
          window.location.href = '/login';
          return Promise.reject(refreshError);
        }
      } else {
        // 没有refresh token，直接跳转到登录页
        localStorage.removeItem('authToken');
        localStorage.removeItem('userId');
        window.location.href = '/login';
      }
    }
    
    // 处理网络错误
    if (!error.response) {
      console.error('Network error:', error.message);
      return Promise.reject({
        message: '网络连接失败，请检查网络设置',
        status: 0,
      } as ApiError);
    }
    
    // 处理服务器错误
    const apiError: ApiError = {
      message: error.response.data?.message || '服务器错误',
      status: error.response.status,
      errors: error.response.data?.errors,
    };
    
    console.error('API error:', apiError);
    return Promise.reject(apiError);
  }
);

// API工具函数
export const apiUtils = {
  // 设置认证token
  setAuthToken: (token: string) => {
    localStorage.setItem('authToken', token);
  },
  
  // 清除认证token
  clearAuthToken: () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('userId');
  },
  
  // 获取当前token
  getAuthToken: () => {
    return localStorage.getItem('authToken');
  },
  
  // 检查是否已认证
  isAuthenticated: () => {
    return !!localStorage.getItem('authToken');
  },
  
  // 设置用户ID
  setUserId: (userId: string) => {
    localStorage.setItem('userId', userId);
  },
  
  // 获取用户ID
  getUserId: () => {
    return localStorage.getItem('userId');
  },
  
  // 处理API响应
  handleResponse: <T>(response: AxiosResponse<ApiResponse<T>>): T => {
    if (response.data.success) {
      return response.data.data as T;
    } else {
      throw new Error(response.data.message || '请求失败');
    }
  },
  
  // 处理API错误
  handleError: (error: any): ApiError => {
    if (error.response) {
      return {
        message: error.response.data?.message || '服务器错误',
        status: error.response.status,
        errors: error.response.data?.errors,
      };
    } else if (error.request) {
      return {
        message: '网络连接失败',
        status: 0,
      };
    } else {
      return {
        message: error.message || '未知错误',
      };
    }
  },
  
  // 构建查询参数
  buildQueryParams: (params: Record<string, any>): URLSearchParams => {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        if (Array.isArray(value)) {
          value.forEach(v => searchParams.append(key, v.toString()));
        } else {
          searchParams.append(key, value.toString());
        }
      }
    });
    return searchParams;
  },
  
  // 上传文件
  uploadFile: async (file: File, endpoint: string, onProgress?: (progress: number) => void) => {
    const formData = new FormData();
    formData.append('file', file);
    
    return api.post(endpoint, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  },
  
  // 下载文件
  downloadFile: async (endpoint: string, filename?: string) => {
    const response = await api.get(endpoint, {
      responseType: 'blob',
    });
    
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename || 'download');
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
};

// 健康检查
export const healthCheck = async (): Promise<boolean> => {
  try {
    const response = await api.get('/health');
    return response.status === 200;
  } catch (error) {
    console.error('Health check failed:', error);
    return false;
  }
};

// 服务状态检查
export const checkServiceStatus = async () => {
  try {
    const response = await api.get('/status');
    return response.data;
  } catch (error) {
    console.error('Service status check failed:', error);
    return null;
  }
};

// 导出axios实例
export { api };
export default api;

// 导出类型
export type { ApiResponse, ApiError };