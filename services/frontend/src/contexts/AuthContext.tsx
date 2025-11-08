import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { authApi } from '../services/authApi';
import { tokenManager } from '../services/tokenManager';
import * as authApiTypes from '../services/authApi';
import type { LoginRequest, RegisterRequest, User } from '../services/authApi';

// 认证上下文类型
interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (credentials: authApiTypes.LoginRequest) => Promise<boolean>;
  register: (data: authApiTypes.RegisterRequest) => Promise<boolean>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<boolean>;
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>;
  revokeToken: (reason?: string) => Promise<boolean>;
  clearError: () => void;
}

// 创建认证上下文
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// 认证提供者组件
interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // 初始化时检查token
  useEffect(() => {
    const initAuth = async () => {
      try {
        // 使用tokenManager检查有效令牌
        const token = await tokenManager.getValidToken();
        if (token) {
          setToken(token);
          // 获取用户信息
          const userResponse = await authApi.getProfile();
          setUser(userResponse.data.data);
        }
      } catch (error) {
        console.error('初始化认证失败:', error);
        await tokenManager.clearToken();
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  // 登录函数
  const login = async (credentials: LoginRequest): Promise<boolean> => {
    try {
      setIsLoading(true);
      const response = await authApi.login(credentials);
      
      // 使用tokenManager存储令牌
      // 注意：当前API响应不包含refreshToken，只传递token和expires_at
      tokenManager.setToken({
        token: response.data.token,
        expiresAt: response.data.expires_at,
      });
      
      setUser(response.data.user);
      setToken(response.data.token);
      setError(null);
      return true;
    } catch (err: any) {
      setError(err.response?.data?.message || '登录失败');
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  // 注册函数
  const register = async (userData: RegisterRequest): Promise<boolean> => {
    try {
      setIsLoading(true);
      await authApi.register(userData);
      // 注册成功后不自动登录，需要用户手动登录
      setError(null);
      return true;
    } catch (err: any) {
      setError(err.response?.data?.message || '注册失败');
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  // 登出
  const logout = async () => {
    try {
      if (token) {
        await authApi.logout();
      }
    } catch (error) {
      console.error('登出失败:', error);
    } finally {
      await tokenManager.clearToken();
      setToken(null);
      setUser(null);
    }
  };

  // 刷新token函数
  const refreshToken = async (): Promise<boolean> => {
    try {
      const response = await authApi.refreshToken();
      const newToken = response.data.data.token;
      
      // 使用tokenManager更新令牌
      tokenManager.setToken(newToken);
      
      setToken(newToken);
      return true;
    } catch (error) {
      console.error('刷新token失败:', error);
      // 刷新失败，清除所有token并跳转到登录页
      await tokenManager.clearToken();
      setUser(null);
      setToken(null);
      return false;
    }
  };

  // 修改密码
  const changePassword = async (oldPassword: string, newPassword: string) => {
    try {
      await authApi.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      });
    } catch (error) {
      console.error('修改密码失败:', error);
      throw error;
    }
  };

  // 撤销令牌函数
  const revokeToken = async (reason?: string): Promise<boolean> => {
    try {
      await authApi.revokeToken(reason);
      // 撤销成功，清除本地令牌
      await tokenManager.clearToken();
      setUser(null);
      setToken(null);
      return true;
    } catch (error) {
      console.error('撤销令牌失败:', error);
      return false;
    }
  };

  // 清除错误信息
  const clearError = () => {
    setError(null);
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated: !!user,
    isLoading,
    error,
    login,
    register,
    logout,
    refreshToken,
    changePassword,
    revokeToken,
    clearError,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// 使用认证上下文的Hook
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth 必须在 AuthProvider 内部使用');
  }
  return context;
};