import { useState, useEffect, createContext, useContext } from 'react';
import type { User } from '../types';
import { apiClient } from '../services/api';

interface ApiError {
  response?: {
    data?: {
      message?: string;
    };
  };
  message?: string;
}

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

interface AuthContextType extends AuthState {
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (userData: { username: string; email: string; password: string }) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
  updateProfile: (profileData: Partial<User>) => Promise<{ success: boolean; error?: string }>;
  checkAuthStatus: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    isLoading: true,
    isAuthenticated: false,
  });

  // 防止重复调用的标志
  const [isChecking, setIsChecking] = useState(false);

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    // 防止重复调用
    if (isChecking) {
      return;
    }
    
    setIsChecking(true);
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        setAuthState({
          user: null,
          isLoading: false,
          isAuthenticated: false,
        });
        return;
      }

      const response = await apiClient.getCurrentUser();
      if (response.success && response.data) {
        setAuthState({
          user: response.data,
          isLoading: false,
          isAuthenticated: true,
        });
      } else {
        throw new Error(response.error || 'Failed to get user data');
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      localStorage.removeItem('token');
      setAuthState({
        user: null,
        isLoading: false,
        isAuthenticated: false,
      });
    } finally {
      setIsChecking(false);
    }
  };

  const login = async (email: string, password: string) => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true }));
      
      const response = await apiClient.login(email, password);

      if (response.success) {
        const { token, user } = response.data;
        localStorage.setItem('token', token);
        
        // 立即更新认证状态
        setAuthState({
          user,
          isLoading: false,
          isAuthenticated: true,
        });

        // 登录成功后验证认证状态，确保用户信息是最新的
        await checkAuthStatus();

        return { success: true };
      } else {
        throw new Error(response.message || '登录失败');
      }
    } catch (error) {
      const apiError = error as ApiError;
      setAuthState(prev => ({ ...prev, isLoading: false }));
      return { 
        success: false, 
        error: apiError.response?.data?.message || apiError.message || '登录失败' 
      };
    }
  };

  const register = async (userData: {
    username: string;
    email: string;
    password: string;
  }) => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true }));
      
      const response = await apiClient.register(userData.username, userData.email, userData.password);
      
      if (response.success) {
        const { token, user } = response.data;
        
        localStorage.setItem('token', token);
        
        setAuthState({
          user,
          isLoading: false,
          isAuthenticated: true,
        });

        return { success: true };
      } else {
        throw new Error(response.message || '注册失败');
      }
    } catch (error) {
      const apiError = error as ApiError;
      setAuthState(prev => ({ ...prev, isLoading: false }));
      return { 
        success: false, 
        error: apiError.response?.data?.message || apiError.message || '注册失败' 
      };
    }
  };

  const logout = async () => {
    try {
      // 调用后端logout API
      await apiClient.logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
      // 即使API调用失败，也要清除本地状态
    } finally {
      // 无论API调用是否成功，都要清除本地token和状态
      localStorage.removeItem('token');
      setAuthState({
        user: null,
        isLoading: false,
        isAuthenticated: false,
      });
    }
  };

  const updateProfile = async (profileData: Partial<User>) => {
    try {
      const response = await apiClient.updateProfile(profileData);
      setAuthState(prev => ({
        ...prev,
        user: response,
      }));
      return { success: true };
    } catch (error) {
      const apiError = error as ApiError;
      return { 
        success: false, 
        error: apiError.response?.data?.message || '更新失败' 
      };
    }
  };

  const contextValue: AuthContextType = {
    ...authState,
    login,
    register,
    logout,
    updateProfile,
    checkAuthStatus,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};