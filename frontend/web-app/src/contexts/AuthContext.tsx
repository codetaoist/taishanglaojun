import React, { createContext, useContext, useEffect, useState } from 'react'
import { apiManager } from '../services/apiManager'

interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  name?: string;
  role: string;
  roles?: string[];
  permissions?: string[];
  isAdmin?: boolean;
  created_at: string;
  updated_at: string;
}

interface AuthContextType {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<{success: boolean, error?: string}>;
  logout: () => Promise<void>
  register: (userData: any) => Promise<void>
  updateProfile: (userData: Partial<User>) => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const useAuthContext = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuthContext must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: React.ReactNode
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true) // 初始状态设为true，避免闪烁

  const isAuthenticated = !!user

  useEffect(() => {
    const initializeAuth = async () => {
      console.log('🔄 AuthContext - 开始初始化认证状态');
      const token = localStorage.getItem('token')
      const lastActivity = localStorage.getItem('lastActivity')
      
      console.log('🔑 Token存在:', !!token);
      console.log('⏰ 最后活动时间:', lastActivity);
      
      if (token) {
        // 检查token是否过期（24小时）
        const now = Date.now()
        const lastActivityTime = lastActivity ? parseInt(lastActivity) : 0
        const tokenExpired = now - lastActivityTime > 24 * 60 * 60 * 1000 // 24小时
        
        console.log('⏱️ 时间差:', now - lastActivityTime, '最大允许:', 24 * 60 * 60 * 1000);
        
        if (tokenExpired) {
          console.log('⚠️ Token已过期，清除认证状态');
          localStorage.removeItem('token')
          localStorage.removeItem('lastActivity')
          setUser(null)
        } else {
          try {
            console.log('🔄 尝试刷新用户信息...');
            // 验证token并获取用户信息
            await refreshUser()
            console.log('✅ 用户信息刷新成功');
          } catch (error) {
            // 如果token无效，清除本地存储
            console.error('❌ Token验证失败:', error)
            localStorage.removeItem('token')
            localStorage.removeItem('lastActivity')
            setUser(null)
          }
        }
      } else {
        console.log('❌ 没有token，跳过认证');
      }
      setIsLoading(false) // 无论成功失败都要设置loading为false
    }

    initializeAuth()
  }, [])

  const login = async (username: string, password: string) => {
    try {
      setIsLoading(true)
      // apiManager.login 在成功时返回数据本身，失败时抛出异常
      const data = await apiManager.login(username, password)
      
      if (data && data.token) {
        // 先保存token和活动时间
        localStorage.setItem('token', data.token)
        localStorage.setItem('lastActivity', Date.now().toString())
        // 再设置用户状态
        setUser(data.user)
        // 确保状态更新完成后再返回
        await new Promise(resolve => setTimeout(resolve, 100))
        return { success: true, user: data.user, token: data.token }
      } else {
        return { success: false, error: '登录失败，未获取到有效数据' }
      }
    } catch (error: any) {
      return { 
        success: false, 
        error: error.response?.data?.message || error.message || '登录失败，请稍后重试' 
      }
    } finally {
      setIsLoading(false)
    }
  }

  const logout = async () => {
    try {
      setIsLoading(true)
      await apiManager.logout()
    } catch (error) {
      // 即使API调用失败，也要继续清除本地状态
      console.warn('Logout API call failed:', error)
    } finally {
      // 无论API调用成功与否，都要清除本地状态
      localStorage.removeItem('token')
      localStorage.removeItem('lastActivity')
      localStorage.removeItem('rememberLogin')
      setUser(null)
      setIsLoading(false)
    }
  }

  const register = async (userData: any) => {
    try {
      setIsLoading(true)
      const response = await apiManager.register(userData)
      
      if (response.success && response.data && response.data.token) {
        localStorage.setItem('token', response.data.token)
        localStorage.setItem('lastActivity', Date.now().toString())
        setUser(response.data.user)
      } else {
        throw new Error(response.error || '注册失败')
      }
    } catch (error) {
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const updateProfile = async (userData: Partial<User>) => {
    try {
      setIsLoading(true)
      const updatedUser = await apiManager.updateProfile(userData)
      setUser(updatedUser)
    } catch (error) {
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const refreshUser = async () => {
    try {
      console.log('🔄 AuthContext - 开始刷新用户信息');
      const userData = await apiManager.getCurrentUser()
      console.log('📥 获取到的用户数据:', userData);
      
      if (userData) {
        setUser(userData)
        console.log('👤 用户状态已更新:', userData);
        // 更新最后活动时间
        localStorage.setItem('lastActivity', Date.now().toString())
      } else {
        console.error('❌ 获取用户信息失败: 没有返回用户数据')
        // 如果获取用户信息失败，清除认证状态
        localStorage.removeItem('token')
        localStorage.removeItem('lastActivity')
        setUser(null)
        throw new Error('获取用户信息失败')
      }
    } catch (error) {
      console.error('❌ 刷新用户信息失败:', error)
      // 如果获取用户信息失败，清除认证状态
      localStorage.removeItem('token')
      localStorage.removeItem('lastActivity')
      setUser(null)
      throw error
    }
  }

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    login,
    logout,
    register,
    updateProfile,
    refreshUser,
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export default AuthContext