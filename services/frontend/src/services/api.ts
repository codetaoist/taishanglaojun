import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

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
const createApiInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // 请求拦截器
  instance.interceptors.request.use(
    (config) => {
      // 在这里可以添加认证token等
      const token = localStorage.getItem('token');
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
    (error) => {
      // 统一错误处理
      if (error.response) {
        // 服务器返回了错误状态码
        const { status, data } = error.response;
        console.error(`API Error [${status}]:`, data);
        
        // 可以根据状态码进行特殊处理
        if (status === 401) {
          // 未授权，可以跳转到登录页
          localStorage.removeItem('token');
          window.location.href = '/login';
        }
      } else if (error.request) {
        // 请求已发出但没有收到响应
        console.error('Network Error:', error.message);
      } else {
        // 请求配置出错
        console.error('Request Error:', error.message);
      }
      
      return Promise.reject(error);
    }
  );

  return instance;
};

// 创建API实例
const api = createApiInstance();

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
export { api };