import { api, apiUtils, ApiResponse } from './api';
import { User } from './authService';

// 权限接口
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

// 角色接口
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

// 用户权限接口
export interface UserPermissions {
  user_id: number;
  roles: Role[];
  permissions: Permission[];
  effective_permissions: string[];
}

// 权限检查请求接口
export interface PermissionCheckRequest {
  user_id?: number;
  resource: string;
  action: string;
  context?: Record<string, any>;
}

// 权限检查响应接口
export interface PermissionCheckResponse {
  allowed: boolean;
  reason?: string;
  required_permissions?: string[];
}

// 角色分配请求接口
export interface RoleAssignmentRequest {
  user_id: number;
  role_ids: number[];
}

// 权限分配请求接口
export interface PermissionAssignmentRequest {
  user_id: number;
  permission_ids: number[];
}

// 权限服务
export const permissionService = {
  // 获取所有权限
  getAllPermissions: async (): Promise<Permission[]> => {
    try {
      const response = await api.get('/permissions');
      return response.data.permissions;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取权限详情
  getPermission: async (id: number): Promise<Permission> => {
    try {
      const response = await api.get(`/permissions/${id}`);
      return response.data.permission;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 创建权限
  createPermission: async (data: Omit<Permission, 'id' | 'created_at' | 'updated_at'>): Promise<Permission> => {
    try {
      const response = await api.post('/permissions', data);
      return response.data.permission;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 更新权限
  updatePermission: async (id: number, data: Partial<Permission>): Promise<Permission> => {
    try {
      const response = await api.put(`/permissions/${id}`, data);
      return response.data.permission;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 删除权限
  deletePermission: async (id: number): Promise<void> => {
    try {
      await api.delete(`/permissions/${id}`);
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取所有角色
  getAllRoles: async (): Promise<Role[]> => {
    try {
      const response = await api.get('/roles');
      return response.data.roles;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取角色详情
  getRole: async (id: number): Promise<Role> => {
    try {
      const response = await api.get(`/roles/${id}`);
      return response.data.role;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 创建角色
  createRole: async (data: Omit<Role, 'id' | 'permissions' | 'created_at' | 'updated_at'>): Promise<Role> => {
    try {
      const response = await api.post('/roles', data);
      return response.data.role;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 更新角色
  updateRole: async (id: number, data: Partial<Role>): Promise<Role> => {
    try {
      const response = await api.put(`/roles/${id}`, data);
      return response.data.role;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 删除角色
  deleteRole: async (id: number): Promise<void> => {
    try {
      await api.delete(`/roles/${id}`);
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 为角色分配权限
  assignPermissionsToRole: async (roleId: number, permissionIds: number[]): Promise<Role> => {
    try {
      const response = await api.post(`/roles/${roleId}/permissions`, {
        permission_ids: permissionIds,
      });
      return response.data.role;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 从角色移除权限
  removePermissionsFromRole: async (roleId: number, permissionIds: number[]): Promise<Role> => {
    try {
      const response = await api.delete(`/roles/${roleId}/permissions`, {
        data: { permission_ids: permissionIds },
      });
      return response.data.role;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取用户权限
  getUserPermissions: async (userId: number): Promise<UserPermissions> => {
    try {
      const response = await api.get(`/users/${userId}/permissions`);
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 为用户分配角色
  assignRolesToUser: async (data: RoleAssignmentRequest): Promise<UserPermissions> => {
    try {
      const response = await api.post(`/users/${data.user_id}/roles`, {
        role_ids: data.role_ids,
      });
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 从用户移除角色
  removeRolesFromUser: async (userId: number, roleIds: number[]): Promise<UserPermissions> => {
    try {
      const response = await api.delete(`/users/${userId}/roles`, {
        data: { role_ids: roleIds },
      });
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 为用户分配权限
  assignPermissionsToUser: async (data: PermissionAssignmentRequest): Promise<UserPermissions> => {
    try {
      const response = await api.post(`/users/${data.user_id}/permissions`, {
        permission_ids: data.permission_ids,
      });
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 从用户移除权限
  removePermissionsFromUser: async (userId: number, permissionIds: number[]): Promise<UserPermissions> => {
    try {
      const response = await api.delete(`/users/${userId}/permissions`, {
        data: { permission_ids: permissionIds },
      });
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 检查权限
  checkPermission: async (data: PermissionCheckRequest): Promise<PermissionCheckResponse> => {
    try {
      const response = await api.post('/permissions/check', data);
      return response.data;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 批量检查权限
  checkMultiplePermissions: async (checks: PermissionCheckRequest[]): Promise<PermissionCheckResponse[]> => {
    try {
      const response = await api.post('/permissions/check-multiple', {
        checks,
      });
      return response.data.results;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取资源列表
  getResources: async (): Promise<string[]> => {
    try {
      const response = await api.get('/permissions/resources');
      return response.data.resources;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 获取操作列表
  getActions: async (resource?: string): Promise<string[]> => {
    try {
      const params = resource ? { resource } : {};
      const response = await api.get('/permissions/actions', { params });
      return response.data.actions;
    } catch (error: any) {
      throw apiUtils.handleError(error);
    }
  },

  // 本地权限检查工具函数
  hasPermission: (userPermissions: UserPermissions, resource: string, action: string): boolean => {
    const permissionCode = `${resource}:${action}`;
    return userPermissions.effective_permissions.includes(permissionCode);
  },

  // 检查用户是否有角色
  hasRole: (userPermissions: UserPermissions, roleCode: string): boolean => {
    return userPermissions.roles.some(role => role.code === roleCode);
  },

  // 检查用户是否有任一角色
  hasAnyRole: (userPermissions: UserPermissions, roleCodes: string[]): boolean => {
    return userPermissions.roles.some(role => roleCodes.includes(role.code));
  },

  // 检查用户是否有所有角色
  hasAllRoles: (userPermissions: UserPermissions, roleCodes: string[]): boolean => {
    const userRoleCodes = userPermissions.roles.map(role => role.code);
    return roleCodes.every(code => userRoleCodes.includes(code));
  },

  // 获取用户可访问的资源
  getAccessibleResources: (userPermissions: UserPermissions): string[] => {
    const resources = new Set<string>();
    userPermissions.effective_permissions.forEach(permission => {
      const [resource] = permission.split(':');
      if (resource) {
        resources.add(resource);
      }
    });
    return Array.from(resources);
  },

  // 获取用户对资源的可用操作
  getAvailableActions: (userPermissions: UserPermissions, resource: string): string[] => {
    const actions = new Set<string>();
    userPermissions.effective_permissions.forEach(permission => {
      const [permResource, action] = permission.split(':');
      if (permResource === resource && action) {
        actions.add(action);
      }
    });
    return Array.from(actions);
  },

  // 权限常量
  PERMISSIONS: {
    // 用户管理
    USER_VIEW: 'user:view',
    USER_CREATE: 'user:create',
    USER_UPDATE: 'user:update',
    USER_DELETE: 'user:delete',
    
    // 角色管理
    ROLE_VIEW: 'role:view',
    ROLE_CREATE: 'role:create',
    ROLE_UPDATE: 'role:update',
    ROLE_DELETE: 'role:delete',
    
    // 权限管理
    PERMISSION_VIEW: 'permission:view',
    PERMISSION_CREATE: 'permission:create',
    PERMISSION_UPDATE: 'permission:update',
    PERMISSION_DELETE: 'permission:delete',
    
    // 意识融合
    CONSCIOUSNESS_VIEW: 'consciousness:view',
    CONSCIOUSNESS_CREATE: 'consciousness:create',
    CONSCIOUSNESS_UPDATE: 'consciousness:update',
    CONSCIOUSNESS_DELETE: 'consciousness:delete',
    CONSCIOUSNESS_ANALYZE: 'consciousness:analyze',
    CONSCIOUSNESS_FUSE: 'consciousness:fuse',
    
    // 文化智慧
    CULTURAL_VIEW: 'cultural:view',
    CULTURAL_CREATE: 'cultural:create',
    CULTURAL_UPDATE: 'cultural:update',
    CULTURAL_DELETE: 'cultural:delete',
    CULTURAL_INQUIRY: 'cultural:inquiry',
    CULTURAL_LEARN: 'cultural:learn',
    
    // 系统管理
    SYSTEM_CONFIG: 'system:config',
    SYSTEM_MONITOR: 'system:monitor',
    SYSTEM_BACKUP: 'system:backup',
  } as const,

  // 角色常量
  ROLES: {
    ADMIN: 'admin',
    USER: 'user',
    MODERATOR: 'moderator',
    GUEST: 'guest',
  } as const,
};

// 导出类型
export type {
  Permission,
  Role,
  UserPermissions,
  PermissionCheckRequest,
  PermissionCheckResponse,
  RoleAssignmentRequest,
  PermissionAssignmentRequest,
};

// 默认导出
export default permissionService;