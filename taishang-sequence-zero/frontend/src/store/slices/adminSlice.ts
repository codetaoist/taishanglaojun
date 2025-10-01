import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { message } from 'antd';
import adminService from '../../services/adminService';
import { permissionService } from '../../services/permissionService';
import type { User } from '../../types/auth';

// 系统统计接口
export interface SystemStats {
  totalUsers: number;
  activeUsers: number;
  totalSessions: number;
  activeSessions: number;
  totalConsciousnessFusions: number;
  totalCulturalInteractions: number;
  systemUptime: number;
  memoryUsage: number;
  cpuUsage: number;
  diskUsage: number;
  networkTraffic: {
    incoming: number;
    outgoing: number;
  };
}

// 用户管理接口
export interface UserManagement {
  users: User[];
  totalCount: number;
  currentPage: number;
  pageSize: number;
}

// 系统日志接口
export interface SystemLog {
  id: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  source: string;
  userId?: string;
  metadata?: Record<string, any>;
}

// 系统配置接口
export interface SystemConfig {
  siteName: string;
  siteDescription: string;
  maxUsers: number;
  sessionTimeout: number;
  enableRegistration: boolean;
  enableGuestAccess: boolean;
}

// 权限管理接口
export interface Permission {
  id: number;
  name: string;
  code: string;
  description?: string;
  resource: string;
  action: string;
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: number;
  name: string;
  code: string;
  description?: string;
  permissions: Permission[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// 管理员状态接口
export interface AdminSliceState {
  // 系统统计
  systemStats: SystemStats | null;
  
  // 用户管理
  userManagement: UserManagement;
  selectedUsers: string[];
  
  // 系统日志
  systemLogs: SystemLog[];
  logFilters: {
    level: string;
    source: string;
    dateRange: [string, string] | null;
  };
  
  // 系统配置
  systemConfig: SystemConfig | null;
  
  // 权限和角色
  permissions: Permission[];
  roles: Role[];
  
  // 加载状态
  loading: boolean;
  statsLoading: boolean;
  usersLoading: boolean;
  logsLoading: boolean;
  configLoading: boolean;
  
  // 错误信息
  error: string | null;
  
  // 操作历史
  operationHistory: Array<{
    id: string;
    action: string;
    target: string;
    timestamp: string;
    adminId: string;
    adminName: string;
    result: 'success' | 'failed';
    details?: string;
  }>;
  
  // 实时监控
  realTimeData: {
    activeConnections: number;
    requestsPerSecond: number;
    errorRate: number;
    responseTime: number;
  };
}

// 初始状态
const initialState: AdminSliceState = {
  systemStats: null,
  userManagement: {
    users: [],
    totalCount: 0,
    currentPage: 1,
    pageSize: 10,
  },
  selectedUsers: [],
  systemLogs: [],
  logFilters: {
    level: 'all',
    source: 'all',
    dateRange: null,
  },
  systemConfig: null,
  permissions: [],
  roles: [],
  loading: false,
  statsLoading: false,
  usersLoading: false,
  logsLoading: false,
  configLoading: false,
  error: null,
  operationHistory: [],
  realTimeData: {
    activeConnections: 0,
    requestsPerSecond: 0,
    errorRate: 0,
    responseTime: 0,
  },
};

// 异步thunk：获取系统统计
export const fetchSystemStats = createAsyncThunk(
  'admin/fetchSystemStats',
  async (_, { rejectWithValue }) => {
    try {
      const response = await adminService.getSystemStats();
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取系统统计失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取用户列表
export const fetchUsers = createAsyncThunk(
  'admin/fetchUsers',
  async (params: { page?: number; limit?: number; search?: string; status?: string }, { rejectWithValue }) => {
    try {
      const response = await adminService.getUsers(params);
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取用户列表失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：更新用户状态
export const updateUserStatus = createAsyncThunk(
  'admin/updateUserStatus',
  async (params: { userId: string; status: string; reason?: string }, { rejectWithValue }) => {
    try {
      const response = await adminService.updateUserStatus(params.userId, params.status, params.reason);
      if (response.success) {
        message.success('用户状态更新成功');
      }
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '更新用户状态失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取系统日志
export const fetchSystemLogs = createAsyncThunk(
  'admin/fetchSystemLogs',
  async (params: { level?: string; source?: string; limit?: number; offset?: number }, { rejectWithValue }) => {
    try {
      const response = await adminService.getSystemLogs(params);
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取系统日志失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取系统配置
export const fetchSystemConfig = createAsyncThunk(
  'admin/fetchSystemConfig',
  async (_, { rejectWithValue }) => {
    try {
      const response = await adminService.getSystemConfig();
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取系统配置失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：更新系统配置
export const updateSystemConfig = createAsyncThunk(
  'admin/updateSystemConfig',
  async (config: Partial<SystemConfig>, { rejectWithValue }) => {
    try {
      const response = await adminService.updateSystemConfig(config);
      if (response.success) {
        message.success('系统配置更新成功');
      }
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '更新系统配置失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取权限列表
export const fetchPermissions = createAsyncThunk(
  'admin/fetchPermissions',
  async (_, { rejectWithValue }) => {
    try {
      const response = await permissionService.getPermissions();
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取权限列表失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：获取角色列表
export const fetchRoles = createAsyncThunk(
  'admin/fetchRoles',
  async (_, { rejectWithValue }) => {
    try {
      const response = await permissionService.getRoles();
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '获取角色列表失败';
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：创建用户
export const createUser = createAsyncThunk(
  'admin/createUser',
  async (userData: Partial<User>, { rejectWithValue }) => {
    try {
      const response = await adminService.createUser(userData);
      if (response.success) {
        message.success('用户创建成功');
      }
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '创建用户失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：删除用户
export const deleteUser = createAsyncThunk(
  'admin/deleteUser',
  async (userId: string, { rejectWithValue }) => {
    try {
      const response = await adminService.deleteUser(userId);
      if (response.success) {
        message.success('用户删除成功');
      }
      return { userId, ...response };
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '删除用户失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：创建角色
export const createRole = createAsyncThunk(
  'admin/createRole',
  async (roleData: Partial<Role>, { rejectWithValue }) => {
    try {
      const response = await permissionService.createRole(roleData);
      if (response.success) {
        message.success('角色创建成功');
      }
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '创建角色失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：更新角色
export const updateRole = createAsyncThunk(
  'admin/updateRole',
  async (params: { roleId: string; roleData: Partial<Role> }, { rejectWithValue }) => {
    try {
      const response = await permissionService.updateRole(params.roleId, params.roleData);
      if (response.success) {
        message.success('角色更新成功');
      }
      return response;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '更新角色失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 异步thunk：删除角色
export const deleteRole = createAsyncThunk(
  'admin/deleteRole',
  async (roleId: string, { rejectWithValue }) => {
    try {
      const response = await permissionService.deleteRole(roleId);
      if (response.success) {
        message.success('角色删除成功');
      }
      return { roleId, ...response };
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '删除角色失败';
      message.error(errorMessage);
      return rejectWithValue(errorMessage);
    }
  }
);

// 创建管理员slice
const adminSlice = createSlice({
  name: 'admin',
  initialState,
  reducers: {
    // 清除错误
    clearError: (state) => {
      state.error = null;
    },
    
    // 设置选中的用户
    setSelectedUsers: (state, action: PayloadAction<string[]>) => {
      state.selectedUsers = action.payload;
    },
    
    // 添加选中用户
    addSelectedUser: (state, action: PayloadAction<string>) => {
      if (!state.selectedUsers.includes(action.payload)) {
        state.selectedUsers.push(action.payload);
      }
    },
    
    // 移除选中用户
    removeSelectedUser: (state, action: PayloadAction<string>) => {
      state.selectedUsers = state.selectedUsers.filter(id => id !== action.payload);
    },
    
    // 清除选中用户
    clearSelectedUsers: (state) => {
      state.selectedUsers = [];
    },
    
    // 设置日志过滤器
    setLogFilters: (state, action: PayloadAction<Partial<AdminSliceState['logFilters']>>) => {
      state.logFilters = { ...state.logFilters, ...action.payload };
    },
    
    // 添加操作历史
    addOperationHistory: (state, action: PayloadAction<Omit<AdminSliceState['operationHistory'][0], 'id' | 'timestamp'>>) => {
      const operation = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: new Date().toISOString(),
      };
      state.operationHistory.unshift(operation);
      // 保持最近100条记录
      if (state.operationHistory.length > 100) {
        state.operationHistory = state.operationHistory.slice(0, 100);
      }
    },
    
    // 更新实时数据
    updateRealTimeData: (state, action: PayloadAction<Partial<AdminSliceState['realTimeData']>>) => {
      state.realTimeData = { ...state.realTimeData, ...action.payload };
    },
    
    // 重置管理员状态
    resetAdminState: () => initialState,
  },
  extraReducers: (builder) => {
    // 获取系统统计
    builder
      .addCase(fetchSystemStats.pending, (state) => {
        state.statsLoading = true;
        state.error = null;
      })
      .addCase(fetchSystemStats.fulfilled, (state, action) => {
        state.statsLoading = false;
        state.systemStats = {
          ...action.payload,
          activeSessions: action.payload.totalSessions || 0,
          totalConsciousnessFusions: 0,
          totalCulturalInteractions: 0,
          systemUptime: 0,
          cpuUsage: action.payload.systemLoad || 0,
          networkTraffic: {
            incoming: 0,
            outgoing: 0
          }
        };
      })
      .addCase(fetchSystemStats.rejected, (state, action) => {
        state.statsLoading = false;
        state.error = action.payload as string;
      })
      
    // 获取用户列表
    builder
      .addCase(fetchUsers.pending, (state) => {
        state.usersLoading = true;
        state.error = null;
      })
      .addCase(fetchUsers.fulfilled, (state, action) => {
        state.usersLoading = false;
        state.userManagement = {
          users: action.payload.users,
          totalCount: action.payload.totalCount || action.payload.users.length,
          currentPage: action.payload.currentPage || 1,
          pageSize: action.payload.pageSize || 10
        };
      })
      .addCase(fetchUsers.rejected, (state, action) => {
        state.usersLoading = false;
        state.error = action.payload as string;
      })
      
    // 更新用户状态
    builder
      .addCase(updateUserStatus.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateUserStatus.fulfilled, (state, action) => {
        state.loading = false;
        const userIndex = state.userManagement.users.findIndex(u => u.id === action.payload.id);
        if (userIndex > -1) {
          state.userManagement.users[userIndex] = action.payload;
        }
      })
      .addCase(updateUserStatus.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 获取系统日志
    builder
      .addCase(fetchSystemLogs.pending, (state) => {
        state.logsLoading = true;
        state.error = null;
      })
      .addCase(fetchSystemLogs.fulfilled, (state, action) => {
        state.logsLoading = false;
        state.systemLogs = action.payload.logs;
      })
      .addCase(fetchSystemLogs.rejected, (state, action) => {
        state.logsLoading = false;
        state.error = action.payload as string;
      })
      
    // 获取系统配置
    builder
      .addCase(fetchSystemConfig.pending, (state) => {
        state.configLoading = true;
        state.error = null;
      })
      .addCase(fetchSystemConfig.fulfilled, (state, action) => {
        state.configLoading = false;
        state.systemConfig = {
          siteName: action.payload.siteName || 'Taishang Sequence Zero',
          siteDescription: action.payload.siteDescription || '',
          maxUsers: action.payload.maxUsers ?? 100,
          sessionTimeout: action.payload.sessionTimeout ?? 3600,
          enableRegistration: action.payload.enableRegistration ?? true,
          enableGuestAccess: action.payload.enableGuestAccess ?? false
        };
      })
      .addCase(fetchSystemConfig.rejected, (state, action) => {
        state.configLoading = false;
        state.error = action.payload as string;
      })
      
    // 更新系统配置
    builder
      .addCase(updateSystemConfig.pending, (state) => {
        state.configLoading = true;
        state.error = null;
      })
      .addCase(updateSystemConfig.fulfilled, (state, action) => {
        state.configLoading = false;
        if (state.systemConfig) {
          Object.assign(state.systemConfig, action.payload);
        }
      })
      .addCase(updateSystemConfig.rejected, (state, action) => {
        state.configLoading = false;
        state.error = action.payload as string;
      })
      
    // 获取权限列表
    builder
      .addCase(fetchPermissions.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchPermissions.fulfilled, (state, action) => {
        state.loading = false;
        state.permissions = action.payload;
      })
      .addCase(fetchPermissions.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 获取角色列表
    builder
      .addCase(fetchRoles.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchRoles.fulfilled, (state, action) => {
        state.loading = false;
        state.roles = action.payload;
      })
      .addCase(fetchRoles.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 创建用户
    builder
      .addCase(createUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createUser.fulfilled, (state, action) => {
        state.loading = false;
        if (action.payload && state.userManagement?.users) {
          state.userManagement.users.push(action.payload);
          if (state.userManagement.totalCount !== undefined) {
            state.userManagement.totalCount += 1;
          }
        }
      })
      .addCase(createUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 删除用户
    builder
      .addCase(deleteUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteUser.fulfilled, (state, action) => {
        state.loading = false;
        if (state.userManagement?.users) {
          state.userManagement.users = state.userManagement.users.filter(
            user => user.id !== action.payload.userId
          );
          if (state.userManagement.totalCount !== undefined) {
            state.userManagement.totalCount -= 1;
          }
        }
      })
      .addCase(deleteUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 创建角色
    builder
      .addCase(createRole.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createRole.fulfilled, (state, action) => {
        state.loading = false;
        if (action.payload && state.roles) {
          state.roles.push(action.payload);
        }
      })
      .addCase(createRole.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 更新角色
    builder
      .addCase(updateRole.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateRole.fulfilled, (state, action) => {
        state.loading = false;
        if (action.payload && state.roles) {
          const roleIndex = state.roles.findIndex((role: Role) => role.id === action.payload.id);
          if (roleIndex !== -1) {
            state.roles[roleIndex] = action.payload;
          }
        }
      })
      .addCase(updateRole.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
    // 删除角色
    builder
      .addCase(deleteRole.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteRole.fulfilled, (state, action) => {
        state.loading = false;
        if (state.roles) {
          state.roles = state.roles.filter((role: Role) => role.id !== action.payload.roleId);
        }
      })
      .addCase(deleteRole.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

// 导出actions
export const {
  clearError,
  setSelectedUsers,
  addSelectedUser,
  removeSelectedUser,
  clearSelectedUsers,
  setLogFilters,
  addOperationHistory,
  updateRealTimeData,
  resetAdminState,
} = adminSlice.actions;

// 导出异步thunk函数
export {
  fetchSystemStats,
  fetchUsers,
  updateUserStatus,
  fetchSystemLogs,
  fetchSystemConfig,
  updateSystemConfig,
  fetchPermissions,
  fetchRoles,
  createUser,
  deleteUser,
  createRole,
  updateRole,
  deleteRole,
};

// 选择器
export const selectAdmin = (state: { admin: AdminSliceState }) => state.admin;
export const selectSystemStats = (state: { admin: AdminSliceState }) => state.admin.systemStats;
export const selectUserManagement = (state: { admin: AdminSliceState }) => state.admin.userManagement;
export const selectSystemLogs = (state: { admin: AdminSliceState }) => state.admin.systemLogs;
export const selectSystemConfig = (state: { admin: AdminSliceState }) => state.admin.systemConfig;
export const selectPermissions = (state: { admin: AdminSliceState }) => state.admin.permissions;
export const selectRoles = (state: { admin: AdminSliceState }) => state.admin.roles;

// 导出reducer
export default adminSlice.reducer;