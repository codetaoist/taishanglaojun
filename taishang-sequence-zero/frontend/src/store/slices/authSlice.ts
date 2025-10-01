import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { message } from 'antd';
import { authService } from '../../services/authService';
import type { User, LoginCredentials, RegisterData, AuthResponse, UserSettings } from '../../services/authService';
import { permissionService } from '../../services/permissionService';
import type { UserPermissions } from '../../services/permissionService';

// 认证状态接口
export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  loading: boolean;
  error: string | null;
  permissions: UserPermissions | null;
  settings: UserSettings | null;
  lastLoginTime: string | null;
  sessionExpiry: number | null;
  loginAttempts: number;
  isLocked: boolean;
  lockExpiry: number | null;
}

// 初始状态
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null,
  refreshToken: null,
  loading: false,
  error: null,
  permissions: null,
  settings: null,
  lastLoginTime: null,
  sessionExpiry: null,
  loginAttempts: 0,
  isLocked: false,
  lockExpiry: null,
};

// 异步thunk：用户登录
export const loginUser = createAsyncThunk(
  'auth/loginUser',
  async (credentials: LoginCredentials, { rejectWithValue }) => {
    try {
      const response = await authService.login(credentials);
      
      if (response.success && response.access_token && response.user) {
        // 获取用户权限
        const permissionsResponse = await permissionService.getUserPermissions(response.user!.id);
        const permissions = permissionsResponse;
        
        // 获取用户设置
        const settingsResponse = await authService.getUserSettings();
        const settings = settingsResponse;
        
        return {
          user: response.user!,
          token: response.access_token!,
          refreshToken: response.refresh_token!,
          permissions,
          settings,
          expiresIn: response.expires_in!,
        };
      } else {
        return rejectWithValue(response.message || '登录失败');
      }
    } catch (error: any) {
      return rejectWithValue(error.message || '登录失败，请稍后重试');
    }
  }
);

// 异步thunk：用户注册
export const registerUser = createAsyncThunk(
  'auth/registerUser',
  async (data: RegisterData, { rejectWithValue }) => {
    try {
      const response = await authService.register(data);
      
      if (response.success) {
        return response.user!;
      } else {
        return rejectWithValue(response.message || '注册失败');
      }
    } catch (error: any) {
      return rejectWithValue(error.message || '注册失败，请稍后重试');
    }
  }
);

// 异步thunk：检查认证状态
export const checkAuthStatus = createAsyncThunk(
  'auth/checkAuthStatus',
  async (_, { rejectWithValue }) => {
    try {
      // 检查本地存储的token
      const token = localStorage.getItem('authToken');
      if (!token) {
        return rejectWithValue('未登录');
      }
      
      // 验证token
      const tokenValidation = await authService.verifyToken();
      
      if (!tokenValidation.valid) {
        // 尝试刷新token
        try {
          const refreshResponse = await authService.refreshToken();
          if (refreshResponse.success && refreshResponse.user) {
            const permissionsResponse = await permissionService.getUserPermissions(refreshResponse.user.id);
            const permissions = permissionsResponse;
            const settingsResponse = await authService.getUserSettings();
            const settings = settingsResponse;
            
            return {
              user: refreshResponse.user!,
              token: refreshResponse.access_token!,
              refreshToken: refreshResponse.refresh_token!,
              permissions,
              settings,
              expiresIn: refreshResponse.expires_in!,
            };
          }
        } catch (refreshError) {
          return rejectWithValue('认证已过期');
        }
      }
      
      // Token有效，获取最新用户信息
      const currentUserResponse = await authService.getCurrentUser();
      
      const permissionsResponse = await permissionService.getUserPermissions(currentUserResponse.id);
      const permissions = permissionsResponse;
      
      const settingsResponse = await authService.getUserSettings();
      const settings = settingsResponse;

      return {
        user: currentUserResponse,
        token: localStorage.getItem('authToken'),
        refreshToken: localStorage.getItem('refreshToken'),
        permissions,
        settings,
      };
    } catch (error: any) {
      return rejectWithValue(error.message || '认证检查失败');
    }
  }
);

// 异步thunk：刷新token
export const refreshAuthToken = createAsyncThunk(
  'auth/refreshToken',
  async (_, { rejectWithValue }) => {
    try {
      const response = await authService.refreshToken();
      
      if (response.success) {
        return {
          token: response.access_token!,
          refreshToken: response.refresh_token!,
          expiresIn: response.expires_in!,
        };
      } else {
        return rejectWithValue(response.message || 'Token刷新失败');
      }
    } catch (error: any) {
      return rejectWithValue(error.message || 'Token刷新失败');
    }
  }
);

// 异步thunk：用户登出
export const logoutUser = createAsyncThunk(
  'auth/logoutUser',
  async (_, { rejectWithValue }) => {
    try {
      await authService.logout();
      return true;
    } catch (error: any) {
      // 即使服务器登出失败，也要清除本地状态
      return true;
    }
  }
);

// 异步thunk：更新用户信息
export const updateUserProfile = createAsyncThunk(
  'auth/updateProfile',
  async (data: Partial<User>, { rejectWithValue }) => {
    try {
      const response = await authService.updateProfile(data);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新用户信息失败');
    }
  }
);

// 异步thunk：更新用户设置
export const updateUserSettings = createAsyncThunk(
  'auth/updateSettings',
  async (settings: Partial<UserSettings>, { rejectWithValue }) => {
    try {
      const response = await authService.updateUserSettings(settings);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新设置失败');
    }
  }
);

// 异步thunk：上传头像
export const uploadAvatar = createAsyncThunk(
  'auth/uploadAvatar',
  async (file: File, { rejectWithValue }) => {
    try {
      const response = await authService.uploadAvatar(file);
      return response;
    } catch (error: any) {
      return rejectWithValue(error.message || '上传头像失败');
    }
  }
);

// 创建slice
const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    // 清除错误
    clearError: (state) => {
      state.error = null;
    },
    
    // 设置加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    
    // 增加登录尝试次数
    incrementLoginAttempts: (state) => {
      state.loginAttempts += 1;
      if (state.loginAttempts >= 5) {
        state.isLocked = true;
        state.lockExpiry = Date.now() + 15 * 60 * 1000; // 锁定15分钟
      }
    },
    
    // 重置登录尝试次数
    resetLoginAttempts: (state) => {
      state.loginAttempts = 0;
      state.isLocked = false;
      state.lockExpiry = null;
    },
    
    // 检查锁定状态
    checkLockStatus: (state) => {
      if (state.isLocked && state.lockExpiry && Date.now() > state.lockExpiry) {
        state.isLocked = false;
        state.lockExpiry = null;
        state.loginAttempts = 0;
      }
    },
    
    // 更新权限
    updatePermissions: (state, action: PayloadAction<UserPermissions>) => {
      state.permissions = action.payload;
    },
    
    // 重置状态
    resetAuthState: () => initialState,
  },
  extraReducers: (builder) => {
    // 登录
    builder
      .addCase(loginUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.loading = false;
        state.isAuthenticated = true;
        state.user = action.payload.user;
        state.token = action.payload.token;
        state.refreshToken = action.payload.refreshToken;
        state.permissions = action.payload.permissions;
        state.settings = action.payload.settings;
        state.lastLoginTime = new Date().toISOString();
        state.sessionExpiry = action.payload.expiresIn ? Date.now() + action.payload.expiresIn * 1000 : null;
        state.loginAttempts = 0;
        state.isLocked = false;
        state.lockExpiry = null;
        state.error = null;
        
        message.success('登录成功');
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        state.loginAttempts += 1;
        
        if (state.loginAttempts >= 5) {
          state.isLocked = true;
          state.lockExpiry = Date.now() + 15 * 60 * 1000;
          message.error('登录失败次数过多，账户已被锁定15分钟');
        } else {
          message.error(action.payload as string);
        }
      });
    
    // 注册
    builder
      .addCase(registerUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(registerUser.fulfilled, (state) => {
        state.loading = false;
        state.error = null;
        message.success('注册成功，请登录');
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 检查认证状态
    builder
      .addCase(checkAuthStatus.pending, (state) => {
        state.loading = true;
      })
      .addCase(checkAuthStatus.fulfilled, (state, action) => {
        state.loading = false;
        state.isAuthenticated = true;
        state.user = action.payload.user;
        state.token = action.payload.token;
        state.refreshToken = action.payload.refreshToken;
        state.permissions = action.payload.permissions;
        state.settings = action.payload.settings;
        state.error = null;
      })
      .addCase(checkAuthStatus.rejected, (state) => {
        state.loading = false;
        state.isAuthenticated = false;
        state.user = null;
        state.token = null;
        state.refreshToken = null;
        state.permissions = null;
        state.settings = null;
      });
    
    // 刷新token
    builder
      .addCase(refreshAuthToken.fulfilled, (state, action) => {
        state.token = action.payload.token;
        state.refreshToken = action.payload.refreshToken;
        state.sessionExpiry = action.payload.expiresIn ? Date.now() + action.payload.expiresIn * 1000 : null;
      })
      .addCase(refreshAuthToken.rejected, (state) => {
        state.isAuthenticated = false;
        state.user = null;
        state.token = null;
        state.refreshToken = null;
        state.permissions = null;
        state.settings = null;
      });
    
    // 登出
    builder
      .addCase(logoutUser.fulfilled, (state) => {
        state.isAuthenticated = false;
        state.user = null;
        state.token = null;
        state.refreshToken = null;
        state.permissions = null;
        state.settings = null;
        state.lastLoginTime = null;
        state.sessionExpiry = null;
        state.error = null;
        message.success('已安全退出');
      });
    
    // 更新用户信息
    builder
      .addCase(updateUserProfile.pending, (state) => {
        state.loading = true;
      })
      .addCase(updateUserProfile.fulfilled, (state, action) => {
        state.loading = false;
        state.user = action.payload;
        state.error = null;
        message.success('用户信息更新成功');
      })
      .addCase(updateUserProfile.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 更新用户设置
    builder
      .addCase(updateUserSettings.pending, (state) => {
        state.loading = true;
      })
      .addCase(updateUserSettings.fulfilled, (state, action) => {
        state.loading = false;
        state.settings = action.payload;
        state.error = null;
        message.success('设置更新成功');
      })
      .addCase(updateUserSettings.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
    
    // 上传头像
    builder
      .addCase(uploadAvatar.pending, (state) => {
        state.loading = true;
      })
      .addCase(uploadAvatar.fulfilled, (state, action) => {
        state.loading = false;
        if (state.user) {
          state.user.avatar_url = action.payload;
        }
        state.error = null;
        message.success('头像上传成功');
      })
      .addCase(uploadAvatar.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        message.error(action.payload as string);
      });
  },
});

// 导出actions
export const {
  clearError,
  setLoading,
  incrementLoginAttempts,
  resetLoginAttempts,
  checkLockStatus,
  updatePermissions,
  resetAuthState,
} = authSlice.actions;

// 选择器
export const selectAuth = (state: { auth: AuthState }) => state.auth;
export const selectUser = (state: { auth: AuthState }) => state.auth.user;
export const selectIsAuthenticated = (state: { auth: AuthState }) => state.auth.isAuthenticated;
export const selectPermissions = (state: { auth: AuthState }) => state.auth.permissions;
export const selectSettings = (state: { auth: AuthState }) => state.auth.settings;
export const selectAuthLoading = (state: { auth: AuthState }) => state.auth.loading;
export const selectAuthError = (state: { auth: AuthState }) => state.auth.error;

// 权限检查选择器
export const selectHasPermission = (resource: string, action: string) => (state: { auth: AuthState }) => {
  const permissions = state.auth.permissions;
  return permissions ? permissionService.hasPermission(permissions, resource, action) : false;
};

export const selectHasRole = (roleCode: string) => (state: { auth: AuthState }) => {
  const permissions = state.auth.permissions;
  return permissions ? permissionService.hasRole(permissions, roleCode) : false;
};

export const selectHasAnyRole = (roleCodes: string[]) => (state: { auth: AuthState }) => {
  const permissions = state.auth.permissions;
  return permissions ? permissionService.hasAnyRole(permissions, roleCodes) : false;
};

// 导出reducer
export default authSlice.reducer;