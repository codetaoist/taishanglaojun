// 测试权限API调用的工具函数
import { apiClient } from '../services/api';

export const testPermissionAPI = async () => {
  console.log('🧪 开始测试权限API...');
  
  try {
    // 测试基础连接
    console.log('1️⃣ 测试基础API连接...');
    const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
    console.log('🌐 API基础URL:', baseURL);
    
    // 检查认证状态
    const token = localStorage.getItem('token');
    const userInfo = localStorage.getItem('userInfo');
    console.log('🔑 认证令牌:', token ? `已存在 (长度: ${token.length})` : '未找到');
    console.log('👤 用户信息:', userInfo ? JSON.parse(userInfo) : '未找到');
    
    if (!token) {
      console.warn('⚠️ 未找到认证令牌，API调用可能失败');
      return {
        success: false,
        error: '未找到认证令牌',
        needsAuth: true
      };
    }
    
    // 测试权限API调用
    console.log('2️⃣ 测试权限API调用...');
    console.log('📡 请求URL:', `${baseURL}/permissions`);
    console.log('📋 请求参数:', { page_size: 1000 });
    
    const response = await apiClient.get('/permissions', {
      params: { page_size: 1000 }
    });
    
    console.log('✅ API调用成功!');
    console.log('📊 响应状态:', response.status);
    console.log('📄 响应数据:', response.data);
    console.log('📊 响应头:', response.headers);
    
    return {
      success: true,
      data: response.data,
      status: response.status,
      headers: response.headers
    };
    
  } catch (error: any) {
    console.error('❌ API调用失败:', error);
    
    const errorDetails = {
      message: error.message,
      status: error.response?.status,
      statusText: error.response?.statusText,
      data: error.response?.data,
      headers: error.response?.headers,
      config: {
        url: error.config?.url,
        method: error.config?.method,
        baseURL: error.config?.baseURL,
        headers: error.config?.headers
      }
    };
    
    console.error('🔍 错误详情:', errorDetails);
    
    // 特殊处理认证错误
    if (error.response?.status === 401) {
      console.error('🚫 认证失败 - 令牌可能已过期或无效');
    } else if (error.response?.status === 403) {
      console.error('🚫 权限不足 - 用户没有访问权限API的权限');
    } else if (error.response?.status === 404) {
      console.error('🚫 API端点未找到 - 检查后端服务是否正常运行');
    }
    
    return {
      success: false,
      error: error.message,
      status: error.response?.status,
      data: error.response?.data,
      errorDetails
    };
  }
};

// 在浏览器控制台中可以调用的全局函数
(window as any).testPermissionAPI = testPermissionAPI;