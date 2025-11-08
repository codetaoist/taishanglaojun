import { authApiRequest } from './api';

// 认证相关的类型定义
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  role?: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: {
    id: number;
    username: string;
    email: string;
    role: string;
  };
}

export interface RefreshTokenRequest {
  token: string;
}

export interface RefreshTokenResponse {
  token: string;
  expires_at: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  status: string;
}

// 认证API服务
export const authApi = {
  // 登录
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    return authApiRequest.post<LoginResponse>('/api/v1/auth/login', data);
  },

  // 注册
  register: async (data: RegisterRequest): Promise<void> => {
    return authApiRequest.post<void>('/api/v1/auth/register', data);
  },

  // 刷新令牌
  refreshToken: async (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
    return authApiRequest.post<RefreshTokenResponse>('/api/v1/auth/refresh', data);
  },

  // 登出
  logout: async (): Promise<void> => {
    return authApiRequest.post<void>('/api/v1/auth/logout');
  },

  // 撤销令牌
  revokeToken: async (reason?: string) => {
    return authApiRequest.post('/api/v1/auth/revoke-token', { reason });
  },

  // 修改密码
  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    return authApiRequest.post<void>('/api/v1/auth/change-password', data);
  },

  // 获取用户资料
  getProfile: async (): Promise<User> => {
    return authApiRequest.get<User>('/api/v1/profile');
  },

  // 获取用户信息（管理员权限）
  getUser: async (id: number): Promise<User> => {
    return authApiRequest.get<User>(`/api/v1/admin/users/${id}`);
  },
};