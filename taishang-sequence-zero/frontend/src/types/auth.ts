// 用户角色类型
export type UserRole = 'user' | 'admin' | 'super_admin';

// 用户状态类型
export type UserStatus = 'active' | 'inactive' | 'suspended' | 'pending';

// 用户接口
export interface User {
  id: string;
  username: string;
  email: string;
  fullName?: string;
  avatar?: string;
  role: UserRole;
  status: UserStatus;
  bio?: string;
  location?: string;
  website?: string;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
  emailVerified: boolean;
  preferences?: {
    theme: 'light' | 'dark';
    language: 'zh' | 'en';
    notifications: {
      email: boolean;
      push: boolean;
      sms: boolean;
    };
    privacy: {
      profileVisible: boolean;
      activityVisible: boolean;
    };
  };
}

// 登录凭据接口
export interface LoginCredentials {
  email: string;
  password: string;
  rememberMe?: boolean;
  captcha?: string;
}

// 注册数据接口
export interface RegisterData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  fullName?: string;
  agreeToTerms: boolean;
  captcha?: string;
}

// 忘记密码数据接口
export interface ForgotPasswordData {
  email: string;
  captcha?: string;
}

// 重置密码数据接口
export interface ResetPasswordData {
  email: string;
  verificationCode: string;
  newPassword: string;
  confirmPassword: string;
}

// 更改密码数据接口
export interface ChangePasswordData {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

// 认证响应接口
export interface AuthResponse {
  success: boolean;
  message: string;
  data?: {
    user: User;
    token: string;
    refreshToken?: string;
    expiresIn: number;
  };
  errors?: Record<string, string[]>;
}

// JWT Token 载荷接口
export interface JWTPayload {
  sub: string; // 用户ID
  email: string;
  role: UserRole;
  iat: number; // 签发时间
  exp: number; // 过期时间
}

// 认证错误类型
export interface AuthError {
  code: string;
  message: string;
  field?: string;
}

// 用户统计信息接口
export interface UserStats {
  totalSessions: number;
  wisdomPoints: number;
  completedQuests: number;
  currentStreak: number;
  totalStudyTime: number; // 分钟
  achievements: string[];
  level: number;
  experience: number;
}

// 用户活动记录接口
export interface UserActivity {
  id: string;
  type: 'login' | 'logout' | 'fusion_session' | 'wisdom_study' | 'quiz_completed' | 'achievement_unlocked';
  description: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

// 用户设置接口
export interface UserSettings {
  theme: 'light' | 'dark' | 'auto';
  language: 'zh' | 'en';
  notifications: {
    email: boolean;
    push: boolean;
    sms: boolean;
    types: {
      system: boolean;
      fusion: boolean;
      wisdom: boolean;
      social: boolean;
    };
  };
  privacy: {
    profileVisible: boolean;
    activityVisible: boolean;
    allowFriendRequests: boolean;
    showOnlineStatus: boolean;
  };
  preferences: {
    autoSave: boolean;
    soundEnabled: boolean;
    animationsEnabled: boolean;
    compactMode: boolean;
  };
}