import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { invoke } from '@tauri-apps/api/core';

interface User {
  id: string;
  username: string;
  email: string;
  role?: string;
  isPremium?: boolean;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  isPremium: boolean;
  login: (username: string, password: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  togglePremium: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  const togglePremium = () => {
    if (user) {
      setUser({
        ...user,
        isPremium: !user.isPremium
      });
    }
  };

  useEffect(() => {
    // Check if user is already logged in
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      // Check if we're in a Tauri environment
      if (typeof window !== 'undefined' && window.__TAURI__) {
        // Use Tauri commands to check session validity
        try {
          const sessionValid = await invoke('validate_session');
          if (sessionValid) {
            // Get user info if session is valid
            const userInfo = await invoke('get_user_info');
            if (userInfo) {
              setUser(userInfo as User);
              setIsAuthenticated(true);
            } else {
              setUser(null);
              setIsAuthenticated(false);
            }
          } else {
            setUser(null);
            setIsAuthenticated(false);
          }
        } catch (tauriError) {
          console.warn('Tauri session validation failed:', tauriError);
          // 在Tauri环境中，如果命令失败，假设用户未登录
          setUser(null);
          setIsAuthenticated(false);
        }
      } else {
        // In browser environment, check for stored token
        const token = localStorage.getItem('auth_token');
        if (token) {
          try {
            // 在Web环境中，由于后端服务可能不可用，我们先检查token的基本有效性
            // 如果token存在且格式正确，暂时认为用户已登录
            // 实际的验证会在用户进行操作时进行
            const tokenParts = token.split('.');
            if (tokenParts.length === 3) {
              // 简单的JWT格式检查
              try {
                const payload = JSON.parse(atob(tokenParts[1]));
                if (payload.exp && payload.exp > Date.now() / 1000) {
                  // Token未过期，从localStorage恢复用户信息
                  const storedUser = localStorage.getItem('user_info');
                  if (storedUser) {
                    setUser(JSON.parse(storedUser));
                    setIsAuthenticated(true);
                  } else {
                    // 如果没有存储的用户信息，清除token
                    localStorage.removeItem('auth_token');
                    setUser(null);
                    setIsAuthenticated(false);
                  }
                } else {
                  // Token已过期
                  localStorage.removeItem('auth_token');
                  localStorage.removeItem('user_info');
                  setUser(null);
                  setIsAuthenticated(false);
                }
              } catch (parseError) {
                // Token格式错误
                localStorage.removeItem('auth_token');
                localStorage.removeItem('user_info');
                setUser(null);
                setIsAuthenticated(false);
              }
            } else {
              // Token格式错误
              localStorage.removeItem('auth_token');
              localStorage.removeItem('user_info');
              setUser(null);
              setIsAuthenticated(false);
            }
          } catch (error) {
            console.error('Token validation failed:', error);
            localStorage.removeItem('auth_token');
            localStorage.removeItem('user_info');
            setUser(null);
            setIsAuthenticated(false);
          }
        } else {
          setUser(null);
          setIsAuthenticated(false);
        }
      }
    } catch (error) {
      console.error('Auth check failed:', error);
      setUser(null);
      setIsAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  const login = async (username: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {

      if (typeof window !== 'undefined' && window.__TAURI__) {
        // Use Tauri invoke to call backend API
        const response = await invoke('auth_login', { 
          request: { username, password, remember_me: false }
        });
        
        if (response && (response as any).success) {
          const userInfo = await invoke('get_user_info');
          if (userInfo) {
            setUser(userInfo as User);
            setIsAuthenticated(true);
            return { success: true };
          }
        }
        return { success: false, error: (response as any)?.message || '用户名或密码错误' };
      } else {
        // Browser mode - use mock authentication since backend API is not available
        console.log('Browser mode: Using mock authentication');
        
        // Mock user database
        const mockUsers = {
          'admin': {
            password: 'Admin123!',
            userData: {
              id: '1',
              username: 'admin',
              email: 'admin@example.com',
              role: 'ADMIN',
              displayName: '管理员',
              permissions: ['ADMIN', 'USER', 'MODERATOR'],
              isPremium: true,
            }
          },
          'testuser': {
            password: 'Test123!',
            userData: {
              id: '2',
              username: 'testuser',
              email: 'testuser@example.com',
              role: 'USER',
              displayName: '测试用户',
              permissions: ['USER'],
              isPremium: false,
            }
          }
        };

        // Validate credentials
        const user = mockUsers[username as keyof typeof mockUsers];
        if (!user || user.password !== password) {
          return { success: false, error: '用户名或密码错误' };
        }

        // Generate mock token
        const mockToken = `mock_token_${username}_${Date.now()}`;
        
        // Store token and user info in localStorage
        localStorage.setItem('auth_token', mockToken);
        localStorage.setItem('user_info', JSON.stringify(user.userData));
        
        setUser(user.userData);
        setIsAuthenticated(true);
        return { success: true };
      }
    } catch (error) {
      return { success: false, error: '登录失败，请重试' };
    } finally {
      setLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        await invoke('auth_logout');
      } else {
        // Browser mode - call backend API and clear token
        const token = localStorage.getItem('auth_token');
        if (token) {
          try {
            await fetch('http://localhost:8080/api/v1/auth/logout', {
              method: 'POST',
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
              },
            });
          } catch (apiError) {
            console.error('API logout failed:', apiError);
          }
          localStorage.removeItem('auth_token');
          localStorage.removeItem('user_info');
        }
      }
      setUser(null);
      setIsAuthenticated(false);
    } catch (error) {
      console.error('Logout failed:', error);
      setUser(null);
      setIsAuthenticated(false);
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user_info');
    }
  };

  const register = async (username: string, email: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      setLoading(true);
      
      // Validate input
      if (!username.trim()) {
        return { success: false, error: '请输入用户名' };
      }
      if (!email.trim()) {
        return { success: false, error: '请输入邮箱' };
      }
      if (!password.trim()) {
        return { success: false, error: '请输入密码' };
      }
      if (password.length < 6) {
        return { success: false, error: '密码长度至少6位' };
      }
      
      if (typeof window !== 'undefined' && window.__TAURI__) {
        const response = await invoke('auth_register', { 
          request: { username, email, password, display_name: username }
        });
        if (response && (response as any).success) {
          return { success: true };
        } else {
          return { success: false, error: (response as any)?.message || '注册失败，请重试' };
        }
      } else {
        // Browser mode - call backend API directly
        try {
          const response = await fetch('http://localhost:8080/api/v1/auth/register', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, email, password }),
          });

          if (response.ok) {
            const data = await response.json();
            if (data.success) {
              return { success: true };
            } else {
              return { success: false, error: data.message || '注册失败' };
            }
          } else if (response.status === 409) {
            return { success: false, error: '用户名或邮箱已存在' };
          } else if (response.status === 400) {
            return { success: false, error: '输入信息格式不正确' };
          } else if (response.status >= 500) {
            return { success: false, error: '服务器错误，请稍后再试' };
          } else {
            return { success: false, error: '注册失败，请检查网络连接' };
          }
        } catch (apiError) {
          console.error('API registration failed:', apiError);
          return { success: false, error: '网络连接失败，请检查网络设置' };
        }
      }
    } catch (error) {
      console.error('Registration failed:', error);
      return { success: false, error: '注册过程中发生未知错误' };
    } finally {
      setLoading(false);
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated,
    loading,
    isPremium: user?.isPremium || false,
    login,
    logout,
    register,
    togglePremium,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}