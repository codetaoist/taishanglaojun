import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { tokenManager } from './tokenManager';

// API响应的基础接口
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

// 错误信息接口
export interface ErrorInfo {
  code: string;
  message: string;
  details?: any;
}

// 创建axios实例
const createApiInstance = (baseURL: string): AxiosInstance => {
  const instance = axios.create({
    baseURL,
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // 请求拦截器
  instance.interceptors.request.use(
    async (config) => {
      // 获取有效的令牌（如果需要则自动刷新）
      const token = await tokenManager.getValidToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      return response;
    },
    async (error) => {
      // 处理401未授权错误
      if (error.response?.status === 401) {
        // 如果是登录请求，不尝试刷新令牌，直接返回错误
        if (error.config.url?.includes('/api/v1/auth/login')) {
          return Promise.reject(error);
        }
        
        // 尝试刷新令牌
        const refreshed = await tokenManager.refreshToken();
        
        // 如果刷新成功，重试原请求
        if (refreshed) {
          const originalRequest = error.config;
          const token = await tokenManager.getValidToken();
          if (token) {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return instance(originalRequest);
          }
        }
        
        // 刷新失败，清除令牌并重定向到登录页
        tokenManager.clearToken();
        window.location.href = '/login';
      }
      
      return Promise.reject(error);
    }
  );

  return instance;
};

// 创建API实例 - 通过网关连接到各个服务
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
const AUTH_API_BASE_URL = API_BASE_URL; // 认证服务也通过网关访问

const api = createApiInstance(API_BASE_URL);
const authApiInstance = createApiInstance(AUTH_API_BASE_URL); // 认证服务通过网关连接

// 封装GET请求
export const get = <T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return api.get(url, config).then(response => response.data);
};

// 封装POST请求
export const post = <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return api.post(url, data, config).then(response => response.data);
};

// 封装PUT请求
export const put = <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return api.put(url, data, config).then(response => response.data);
};

// 封装DELETE请求
export const del = <T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return api.delete(url, config).then(response => response.data);
};

// 导出API实例
export { api, authApiInstance as authApi };

// 认证API请求方法封装
export const authApiRequest = {
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    return authApiInstance.get(url, config).then(response => response.data);
  },
  
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    return authApiInstance.post(url, data, config).then(response => response.data);
  },
  
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    return authApiInstance.put(url, data, config).then(response => response.data);
  },
  
  del: <T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    return authApiInstance.delete(url, config).then(response => response.data);
  }
};

// 保持向后兼容性的单独导出
export const authGet = authApiRequest.get;
export const authPost = authApiRequest.post;
export const authPut = authApiRequest.put;
export const authDel = authApiRequest.del;