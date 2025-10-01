import { api, apiUtils, ApiResponse } from './api';

// 用户接口
export interface User {
  id: number;
  username: string;
  email: string;
  full_name?: string;
  avatar_url?: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
  profile?: {
    bio?: string;
    location?: string;
    website?: string;
    preferences?: Record<string, any>;
  };
}

// 登录凭据接口
export interface LoginCredentials {
  username: string;
  password: string;
  remember_me?: boolean;
}

// 注册数据接口
export interface RegisterData {
  username: string;
  email: string;
  password: string;
  confirm_password: string;
  full_name?: string;
  terms_accepted: boolean;
}

// 忘记密码数据接口
export interface ForgotPasswordData {
  email: string;
}

// 重置密码数据接口
export interface ResetPasswordData {
  token: string;
  password: string;
  confirm_password: string;
}

// 修改密码数据接口
export interface ChangePasswordData {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

// 认证响应接口
export interface AuthResponse {
  success: boolean;
  message?: string;
  user?: User;
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
}

// 用户设置接口
export interface UserSettings {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  notifications: {
    email: boolean;
    push: boolean;
    consciousness_updates: boolean;
    cultural_reminders: boolean;
  };
  privacy: {
    profile_visibility: 'public' | 'private' | 'friends';
    show_activity: boolean;
    allow_messages: boolean;
  };
  consciousness: {
    default_session_duration: number;
    preferred_practices: string[];
    difficulty_level: number;
  };
  cultural: {
    preferred_categories: string[];
    learning_pace: 'slow' | 'medium' | 'fast';
    cultural_periods: string[];
  };
}

// 认证服务
export const authService = {
  // 用户登录
  login: async (credentials: LoginCredentials): Promise<AuthResponse> => {
    try {
      const response = await api.post('/auth/login', credentials);
      const authData = response.data;
      
      if (authData.success && authData.access_token) {
        // 存储认证信息
        apiUtils.setAuthToken(authData.access_token);
        if (authData.refresh_token) {
          localStorage.setItem('refreshToken', authData.refresh_token);
        }
        if (authData.user) {
          apiUtils.setUserId(authData.user.id.toString());
          localStorage.setItem('user', JSON.stringify(authData.user));
        }
      }
      
      return authData;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 用户注册
  register: async (data: RegisterData): Promise<AuthResponse> => {
    try {
      const response = await api.post('/auth/register', data);
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 用户登出
  logout: async (): Promise<void> => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // 清除本地存储
      apiUtils.clearAuthToken();
      localStorage.removeItem('user');
      localStorage.removeItem('userSettings');
    }
  },

  // 刷新token
  refreshToken: async (): Promise<AuthResponse> => {
    try {
      const refreshToken = localStorage.getItem('refreshToken');
      if (!refreshToken) {
        throw new Error('No refresh token available');
      }
      
      const response = await api.post('/auth/refresh', {
        refresh_token: refreshToken,
      });
      
      const authData = response.data;
      if (authData.success && authData.access_token) {
        apiUtils.setAuthToken(authData.access_token);
        if (authData.refresh_token) {
          localStorage.setItem('refreshToken', authData.refresh_token);
        }
      }
      
      return authData;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 验证token
  verifyToken: async (): Promise<{ valid: boolean; user?: User }> => {
    try {
      const response = await api.post('/auth/verify');
      return response.data;
    } catch (error: any) {
      return { valid: false };
    }
  },

  // 获取当前用户信息
  getCurrentUser: async (): Promise<User> => {
    try {
      const response = await api.get('/auth/me');
      const user = response.data.user;
      localStorage.setItem('user', JSON.stringify(user));
      return user;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 更新用户信息
  updateProfile: async (data: Partial<User>): Promise<User> => {
    try {
      const response = await api.put('/auth/profile', data);
      const user = response.data.user;
      localStorage.setItem('user', JSON.stringify(user));
      return user;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 修改密码
  changePassword: async (data: ChangePasswordData): Promise<void> => {
    try {
      await api.put('/auth/change-password', data);
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 忘记密码
  forgotPassword: async (data: ForgotPasswordData): Promise<void> => {
    try {
      await api.post('/auth/forgot-password', data);
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 重置密码
  resetPassword: async (data: ResetPasswordData): Promise<void> => {
    try {
      await api.post('/auth/reset-password', data);
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取用户设置
  getUserSettings: async (): Promise<UserSettings> => {
    try {
      const response = await api.get('/auth/settings');
      const settings = response.data.settings;
      localStorage.setItem('userSettings', JSON.stringify(settings));
      return settings;
    } catch (error: any) {
      // 返回默认设置
      return authService.getDefaultSettings();
    }
  },

  // 更新用户设置
  updateUserSettings: async (settings: Partial<UserSettings>): Promise<UserSettings> => {
    try {
      const response = await api.put('/auth/settings', settings);
      const updatedSettings = response.data.settings;
      localStorage.setItem('userSettings', JSON.stringify(updatedSettings));
      return updatedSettings;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 上传头像
  uploadAvatar: async (file: File): Promise<string> => {
    try {
      const response = await apiUtils.uploadFile(file, '/auth/avatar');
      return response.data.avatar_url;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 删除账户
  deleteAccount: async (password: string): Promise<void> => {
    try {
      await api.delete('/auth/account', {
        data: { password },
      });
      // 清除本地存储
      authService.logout();
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取登录历史
  getLoginHistory: async (limit?: number): Promise<any[]> => {
    try {
      const params = limit ? { limit } : {};
      const response = await api.get('/auth/login-history', { params });
      return response.data.history;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取活跃会话
  getActiveSessions: async (): Promise<any[]> => {
    try {
      const response = await api.get('/auth/sessions');
      return response.data.sessions;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 终止会话
  terminateSession: async (sessionId: string): Promise<void> => {
    try {
      await api.delete(`/auth/sessions/${sessionId}`);
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 本地存储工具函数
  getStoredUser: (): User | null => {
    try {
      const userStr = localStorage.getItem('user');
      return userStr ? JSON.parse(userStr) : null;
    } catch (error) {
      return null;
    }
  },

  getStoredSettings: (): UserSettings | null => {
    try {
      const settingsStr = localStorage.getItem('userSettings');
      return settingsStr ? JSON.parse(settingsStr) : null;
    } catch (error) {
      return null;
    }
  },

  isLoggedIn: (): boolean => {
    return apiUtils.isAuthenticated() && !!authService.getStoredUser();
  },

  // 获取默认设置
  getDefaultSettings: (): UserSettings => {
    return {
      theme: 'auto',
      language: 'zh-CN',
      notifications: {
        email: true,
        push: true,
        consciousness_updates: true,
        cultural_reminders: true,
      },
      privacy: {
        profile_visibility: 'public',
        show_activity: true,
        allow_messages: true,
      },
      consciousness: {
        default_session_duration: 20,
        preferred_practices: ['meditation', 'breathing'],
        difficulty_level: 1,
      },
      cultural: {
        preferred_categories: ['philosophy', 'meditation'],
        learning_pace: 'medium',
        cultural_periods: ['pre_qin', 'sui_tang'],
      },
    };
  },

  // 检查用户权限
  hasPermission: (permission: string): boolean => {
    const user = authService.getStoredUser();
    if (!user) return false;
    
    // 简单的权限检查逻辑
    const adminPermissions = ['admin', 'manage_users', 'manage_content'];
    const userPermissions = ['view_content', 'create_content', 'edit_profile'];
    
    if (user.role === 'admin') {
      return [...adminPermissions, ...userPermissions].includes(permission);
    } else if (user.role === 'user') {
      return userPermissions.includes(permission);
    }
    
    return false;
  },

  // 检查用户角色
  hasRole: (role: string): boolean => {
    const user = authService.getStoredUser();
    return user?.role === role;
  },
};

// 导出类型
export type {
  User,
  LoginCredentials,
  RegisterData,
  ForgotPasswordData,
  ResetPasswordData,
  ChangePasswordData,
  AuthResponse,
  UserSettings,
};

// 默认导出
export default authService;