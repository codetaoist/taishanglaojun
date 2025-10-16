import React, { useState, useEffect } from 'react';
import {
  LogIn,
  LogOut,
  UserPlus,
  RefreshCw,
  Settings,
  Eye,
  EyeOff,
  User,
  Mail,
  Lock,
  Shield,
  Key,
  CheckCircle,
  XCircle,
  AlertCircle,
  Copy,
  X,
} from 'lucide-react';
import { invoke } from '@tauri-apps/api/core';

interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  roles: string[];
  permissions: string[];
  last_login: string;
  created_at: string;
  updated_at: string;
}

interface AuthResponse {
  success: boolean;
  message: string;
  user?: User;
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
}

interface LoginRequest {
  username: string;
  password: string;
  remember_me: boolean;
}

interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  display_name: string;
}

interface RegisterFormData extends RegisterRequest {
  confirmPassword: string;
}

const AuthManagementPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'login' | 'register' | 'profile' | 'settings'>('login');
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState(false);
  const [serverUrl, setServerUrl] = useState('http://localhost:8082');
  const [showPassword, setShowPassword] = useState(false);
  const [showTokenDialog, setShowTokenDialog] = useState(false);

  // 登录表单状态
  const [loginForm, setLoginForm] = useState<LoginRequest>({
    username: '',
    password: '',
    remember_me: false,
  });

  // 注册表单状态
  const [registerForm, setRegisterForm] = useState<RegisterFormData>({
    username: '',
    email: '',
    password: '',
    display_name: '',
    confirmPassword: '',
  });

  useEffect(() => {
    checkLoginStatus();
    loadSettings();
  }, []);

  const checkLoginStatus = async () => {
    try {
      const loggedIn = await invoke<boolean>('auth_is_logged_in');
      setIsLoggedIn(loggedIn);
      
      if (loggedIn) {
        const user = await invoke<User>('auth_get_current_user');
        setCurrentUser(user);
        
        const token = await invoke<string>('auth_get_access_token');
        setAccessToken(token);
        setActiveTab('profile');
      }
    } catch (error) {
      console.error('检查登录状态失败:', error);
    }
  };

  const loadSettings = async () => {
    const savedServerUrl = localStorage.getItem('auth_server_url');
    if (savedServerUrl) {
      setServerUrl(savedServerUrl);
    }
    
    const savedAutoRefresh = localStorage.getItem('auth_auto_refresh');
    if (savedAutoRefresh) {
      setAutoRefreshEnabled(savedAutoRefresh === 'true');
    }
  };

  const handleLogin = async () => {
    if (!loginForm.username || !loginForm.password) {
      setMessage({ type: 'error', text: '请填写用户名和密码' });
      return;
    }

    setLoading(true);
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        // Tauri 环境
        const response = await invoke<AuthResponse>('auth_login', {
          request: loginForm,
        });

        if (response.success) {
          setMessage({ type: 'success', text: '登录成功' });
          setIsLoggedIn(true);
          setCurrentUser(response.user || null);
          setAccessToken(response.access_token || '');
          setLoginForm({ username: '', password: '', remember_me: false });
          setActiveTab('profile');
        } else {
          setMessage({ type: 'error', text: response.message });
        }
      } else {
        // 浏览器环境 - 使用模拟认证
        console.log('Browser mode: Using mock authentication');
        
        // Mock user database
        const mockUsers = {
          'admin': {
            password: 'Admin123!',
            userData: {
              id: '1',
              username: 'admin',
              email: 'admin@example.com',
              display_name: '管理员',
              avatar_url: '',
              roles: ['ADMIN'],
              permissions: ['ADMIN', 'USER', 'MODERATOR'],
              last_login: new Date().toISOString(),
              created_at: '2024-01-01T00:00:00Z',
              updated_at: new Date().toISOString(),
            }
          },
          'testuser': {
            password: 'Test123!',
            userData: {
              id: '2',
              username: 'testuser',
              email: 'testuser@example.com',
              display_name: '测试用户',
              avatar_url: '',
              roles: ['USER'],
              permissions: ['USER'],
              last_login: new Date().toISOString(),
              created_at: '2024-01-01T00:00:00Z',
              updated_at: new Date().toISOString(),
            }
          }
        };

        // Validate credentials
        const user = mockUsers[loginForm.username as keyof typeof mockUsers];
        if (!user || user.password !== loginForm.password) {
          setMessage({ type: 'error', text: '用户名或密码错误' });
          return;
        }

        // Generate mock token
        const mockToken = `mock_token_${loginForm.username}_${Date.now()}`;
        
        // Store token and user info
        localStorage.setItem('auth_token', mockToken);
        
        setMessage({ type: 'success', text: '登录成功' });
        setIsLoggedIn(true);
        setCurrentUser(user.userData);
        setAccessToken(mockToken);
        setLoginForm({ username: '', password: '', remember_me: false });
        setActiveTab('profile');
      }
    } catch (error) {
      setMessage({ type: 'error', text: `登录失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async () => {
    if (!registerForm.username || !registerForm.email || !registerForm.password || !registerForm.display_name) {
      setMessage({ type: 'error', text: '请填写所有必填字段' });
      return;
    }

    setLoading(true);
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        // Tauri 环境
        const response = await invoke<AuthResponse>('auth_register', {
          request: registerForm,
        });

        if (response.success) {
           setMessage({ type: 'success', text: '注册成功，请登录' });
           setRegisterForm({
             username: '',
             email: '',
             password: '',
             display_name: '',
             confirmPassword: '',
           });
           setActiveTab('login');
         } else {
           setMessage({ type: 'error', text: response.message });
         }
      } else {
        // 浏览器环境 - 直接调用 HTTP API
        const response = await fetch('http://localhost:8080/api/v1/auth/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username: registerForm.username,
            email: registerForm.email,
            password: registerForm.password,
            display_name: registerForm.display_name,
          }),
        });

        if (response.ok) {
          const data = await response.json();
          if (data.success) {
             setMessage({ type: 'success', text: '注册成功，请登录' });
             setRegisterForm({
               username: '',
               email: '',
               password: '',
               display_name: '',
               confirmPassword: '',
             });
             setActiveTab('login');
           } else {
             setMessage({ type: 'error', text: data.message || '注册失败' });
           }
        } else {
          setMessage({ type: 'error', text: '注册请求失败' });
        }
      }
    } catch (error) {
      setMessage({ type: 'error', text: `注册失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    setLoading(true);
    try {
      await invoke('auth_logout');
      setMessage({ type: 'success', text: '已退出登录' });
      setIsLoggedIn(false);
      setCurrentUser(null);
      setAccessToken('');
      setActiveTab('login');
    } catch (error) {
      setMessage({ type: 'error', text: `退出登录失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const handleRefreshToken = async () => {
    setLoading(true);
    try {
      const response = await invoke<AuthResponse>('auth_refresh_token');
      if (response.success) {
        setMessage({ type: 'success', text: '令牌刷新成功' });
        setAccessToken(response.access_token || '');
      } else {
        setMessage({ type: 'error', text: response.message });
      }
    } catch (error) {
      setMessage({ type: 'error', text: `刷新令牌失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setMessage({ type: 'success', text: '已复制到剪贴板' });
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  const getMessageIcon = (type: string) => {
    switch (type) {
      case 'success': return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'error': return <XCircle className="h-5 w-5 text-red-500" />;
      case 'info': return <AlertCircle className="h-5 w-5 text-blue-500" />;
      default: return <AlertCircle className="h-5 w-5 text-gray-500" />;
    }
  };

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-4xl mx-auto">
        {/* 头部 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">身份认证管理</h1>
          <p className="text-muted-foreground">管理用户登录、注册和身份验证设置</p>
        </div>

        {/* 消息提示 */}
        {message && (
          <div className={`mb-6 p-4 rounded-lg border flex items-center space-x-3 ${
            message.type === 'success' ? 'bg-green-50 border-green-200 text-green-800' :
            message.type === 'error' ? 'bg-red-50 border-red-200 text-red-800' :
            'bg-blue-50 border-blue-200 text-blue-800'
          }`}>
            {getMessageIcon(message.type)}
            <span>{message.text}</span>
            <button
              onClick={() => setMessage(null)}
              className="ml-auto text-current hover:opacity-70"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {/* 标签页导航 */}
        <div className="mb-6">
          <div className="border-b border-border">
            <nav className="flex space-x-8">
              {[
                { id: 'login', label: '登录', icon: LogIn, disabled: isLoggedIn },
                { id: 'register', label: '注册', icon: UserPlus, disabled: isLoggedIn },
                { id: 'profile', label: '用户信息', icon: User, disabled: !isLoggedIn },
                { id: 'settings', label: '设置', icon: Settings, disabled: false },
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => !tab.disabled && setActiveTab(tab.id as any)}
                  disabled={tab.disabled}
                  className={`flex items-center space-x-2 py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === tab.id
                      ? 'border-primary text-primary'
                      : tab.disabled
                      ? 'border-transparent text-muted-foreground cursor-not-allowed opacity-50'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300'
                  }`}
                >
                  <tab.icon className="h-4 w-4" />
                  <span>{tab.label}</span>
                </button>
              ))}
            </nav>
          </div>
        </div>

        {/* 标签页内容 */}
        <div className="bg-card border border-border rounded-lg p-6">
          {/* 登录标签页 */}
          {activeTab === 'login' && !isLoggedIn && (
            <div className="max-w-md mx-auto">
              <div className="text-center mb-6">
                <LogIn className="h-12 w-12 text-primary mx-auto mb-4" />
                <h2 className="text-2xl font-bold text-foreground">用户登录</h2>
                <p className="text-muted-foreground">请输入您的登录凭据</p>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    用户名
                  </label>
                  <div className="relative">
                    <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type="text"
                      value={loginForm.username}
                      onChange={(e) => setLoginForm({ ...loginForm, username: e.target.value })}
                      className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="输入用户名"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={loginForm.password}
                      onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })}
                      className="w-full pl-10 pr-12 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="输入密码"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="remember"
                    checked={loginForm.remember_me}
                    onChange={(e) => setLoginForm({ ...loginForm, remember_me: e.target.checked })}
                    className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                  />
                  <label htmlFor="remember" className="ml-2 text-sm text-foreground">
                    记住我
                  </label>
                </div>

                <button
                  onClick={handleLogin}
                  disabled={loading}
                  className="w-full bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
                >
                  {loading ? (
                    <RefreshCw className="h-4 w-4 animate-spin" />
                  ) : (
                    <LogIn className="h-4 w-4" />
                  )}
                  <span>{loading ? '登录中...' : '登录'}</span>
                </button>
              </div>
            </div>
          )}

          {/* 注册标签页 */}
          {activeTab === 'register' && !isLoggedIn && (
            <div className="max-w-md mx-auto">
              <div className="text-center mb-6">
                <UserPlus className="h-12 w-12 text-primary mx-auto mb-4" />
                <h2 className="text-2xl font-bold text-foreground">用户注册</h2>
                <p className="text-muted-foreground">创建新的用户账户</p>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    用户名
                  </label>
                  <div className="relative">
                    <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type="text"
                      value={registerForm.username}
                      onChange={(e) => setRegisterForm({ ...registerForm, username: e.target.value })}
                      className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="输入用户名"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    邮箱
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type="email"
                      value={registerForm.email}
                      onChange={(e) => setRegisterForm({ ...registerForm, email: e.target.value })}
                      className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="输入邮箱地址"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    显示名称
                  </label>
                  <input
                    type="text"
                    value={registerForm.display_name}
                    onChange={(e) => setRegisterForm({ ...registerForm, display_name: e.target.value })}
                    className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    placeholder="输入显示名称"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={registerForm.password}
                      onChange={(e) => setRegisterForm({ ...registerForm, password: e.target.value })}
                      className="w-full pl-10 pr-12 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                      placeholder="输入密码"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </div>

                <button
                  onClick={handleRegister}
                  disabled={loading}
                  className="w-full bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
                >
                  {loading ? (
                    <RefreshCw className="h-4 w-4 animate-spin" />
                  ) : (
                    <UserPlus className="h-4 w-4" />
                  )}
                  <span>{loading ? '注册中...' : '注册'}</span>
                </button>
              </div>
            </div>
          )}

          {/* 用户信息标签页 */}
          {activeTab === 'profile' && isLoggedIn && currentUser && (
            <div>
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-4">
                  <div className="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center">
                    <User className="h-8 w-8 text-primary" />
                  </div>
                  <div>
                    <h2 className="text-2xl font-bold text-foreground">{currentUser.display_name}</h2>
                    <p className="text-muted-foreground">@{currentUser.username}</p>
                  </div>
                </div>
                <button
                  onClick={handleLogout}
                  disabled={loading}
                  className="bg-red-500 text-white py-2 px-4 rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                >
                  <LogOut className="h-4 w-4" />
                  <span>退出登录</span>
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* 基本信息 */}
                <div className="bg-secondary/50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-foreground mb-4 flex items-center space-x-2">
                    <User className="h-5 w-5" />
                    <span>基本信息</span>
                  </h3>
                  <div className="space-y-3">
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">用户ID</label>
                      <p className="text-foreground font-mono text-sm">{currentUser.id}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">邮箱</label>
                      <p className="text-foreground">{currentUser.email}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">创建时间</label>
                      <p className="text-foreground">{formatDate(currentUser.created_at)}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">最后登录</label>
                      <p className="text-foreground">{formatDate(currentUser.last_login)}</p>
                    </div>
                  </div>
                </div>

                {/* 权限信息 */}
                <div className="bg-secondary/50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-foreground mb-4 flex items-center space-x-2">
                    <Shield className="h-5 w-5" />
                    <span>权限信息</span>
                  </h3>
                  <div className="space-y-3">
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">角色</label>
                      <div className="flex flex-wrap gap-2 mt-1">
                        {currentUser.roles.map((role, index) => (
                          <span
                            key={index}
                            className="px-2 py-1 bg-primary/10 text-primary text-xs rounded-full"
                          >
                            {role}
                          </span>
                        ))}
                      </div>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">权限</label>
                      <div className="flex flex-wrap gap-2 mt-1">
                        {currentUser.permissions.map((permission, index) => (
                          <span
                            key={index}
                            className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full"
                          >
                            {permission}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>

                {/* 访问令牌 */}
                <div className="md:col-span-2 bg-secondary/50 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-foreground flex items-center space-x-2">
                      <Key className="h-5 w-5" />
                      <span>访问令牌</span>
                    </h3>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => setShowTokenDialog(true)}
                        className="text-primary hover:text-primary/80 text-sm"
                      >
                        查看完整令牌
                      </button>
                      <button
                        onClick={handleRefreshToken}
                        disabled={loading}
                        className="bg-primary text-primary-foreground py-1 px-3 rounded text-sm hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center space-x-1"
                      >
                        <RefreshCw className={`h-3 w-3 ${loading ? 'animate-spin' : ''}`} />
                        <span>刷新</span>
                      </button>
                    </div>
                  </div>
                  <div className="bg-background border border-border rounded p-3 font-mono text-sm">
                    <p className="text-muted-foreground truncate">
                      {accessToken ? `${accessToken.substring(0, 50)}...` : '无令牌'}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* 设置标签页 */}
          {activeTab === 'settings' && (
            <div>
              <h2 className="text-2xl font-bold text-foreground mb-6">认证设置</h2>
              
              <div className="space-y-6">
                <div className="bg-secondary/50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-foreground mb-4">服务器配置</h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-foreground mb-2">
                        认证服务器地址
                      </label>
                      <input
                        type="url"
                        value={serverUrl}
                        onChange={(e) => setServerUrl(e.target.value)}
                        className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                        placeholder="http://localhost:8082"
                      />
                    </div>
                  </div>
                </div>

                <div className="bg-secondary/50 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-foreground mb-4">自动化设置</h3>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <label className="text-sm font-medium text-foreground">自动刷新令牌</label>
                        <p className="text-xs text-muted-foreground">在令牌即将过期时自动刷新</p>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={autoRefreshEnabled}
                          onChange={(e) => setAutoRefreshEnabled(e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                      </label>
                    </div>
                  </div>
                </div>

                <div className="flex justify-end">
                  <button
                    onClick={() => {
                      localStorage.setItem('auth_server_url', serverUrl);
                      localStorage.setItem('auth_auto_refresh', autoRefreshEnabled.toString());
                      setMessage({ type: 'success', text: '设置已保存' });
                    }}
                    className="bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors"
                  >
                    保存设置
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 令牌详情对话框 */}
        {showTokenDialog && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-card border border-border rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-foreground">访问令牌详情</h2>
                <button
                  onClick={() => setShowTokenDialog(false)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>
              
              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <label className="text-sm font-medium text-foreground">完整令牌</label>
                    <button
                      onClick={() => copyToClipboard(accessToken)}
                      className="text-primary hover:text-primary/80 text-sm flex items-center space-x-1"
                    >
                      <Copy className="h-3 w-3" />
                      <span>复制</span>
                    </button>
                  </div>
                  <div className="bg-background border border-border rounded p-3 font-mono text-sm break-all">
                    {accessToken || '无令牌'}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AuthManagementPage;