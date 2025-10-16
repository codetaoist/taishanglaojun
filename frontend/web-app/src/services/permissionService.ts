import { apiClient } from './api';

export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: string;
  name: string;
  code: string;
  description?: string;
  type: 'system' | 'custom' | 'functional' | 'data';
  level: number;
  parent_id?: string;
  permissions: string[];
  is_active: boolean;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserRole {
  user_id: string;
  role_id: string;
  assigned_at: string;
  assigned_by: string;
}

export interface PermissionCheckRequest {
  user_id: string;
  resource: string;
  action: string;
  resource_id?: string;
  context?: Record<string, any>;
}

export interface PermissionCheckResponse {
  allowed: boolean;
  reason: string;
  context?: Record<string, any>;
}

class PermissionService {
  // 权限管理
  async getPermissions(params?: {
    page?: number;
    page_size?: number;
    search?: string;
    resource?: string;
  }) {
    const response = await apiClient.get('/permissions', { params });
    const raw = response.data;
    const list = Array.isArray(raw)
      ? raw
      : Array.isArray(raw?.permissions)
        ? raw.permissions
        : Array.isArray(raw?.data?.permissions)
          ? raw.data.permissions
          : Array.isArray(raw?.data)
            ? raw.data
            : Array.isArray(raw?.items)
              ? raw.items
              : [];

    return { data: list };
  }

  async createPermission(data: {
    name: string;
    resource: string;
    action: string;
    description?: string;
  }) {
    const response = await apiClient.post('/permissions', data);
    return response.data;
  }

  async updatePermission(id: string, data: Partial<Permission>) {
    const response = await apiClient.put(`/permissions/${id}`, data);
    return response.data;
  }

  async deletePermission(id: string) {
    const response = await apiClient.delete(`/permissions/${id}`);
    return response.data;
  }

  // 角色管理
  async getRoles(params?: {
    page?: number;
    page_size?: number;
    search?: string;
    type?: string;
    is_active?: boolean;
    parent_id?: string;
  }) {
    const response = await apiClient.get('/roles', { params });
    const raw = response.data;
    const list = Array.isArray(raw)
      ? raw
      : Array.isArray(raw?.roles)
        ? raw.roles
        : Array.isArray(raw?.data?.roles)
          ? raw.data.roles
          : Array.isArray(raw?.data)
            ? raw.data
            : Array.isArray(raw?.items)
              ? raw.items
              : [];

    return { data: list };
  }

  async getRole(id: string) {
    const response = await apiClient.get(`/roles/${id}`);
    return response.data;
  }

  async createRole(data: {
    name: string;
    code: string;
    description?: string;
    type?: string;
    level?: number;
    parent_id?: string;
    permissions?: string[];
  }) {
    const response = await apiClient.post('/roles', data);
    return response.data;
  }

  async updateRole(id: string, data: Partial<Role>) {
    const response = await apiClient.put(`/roles/${id}`, data);
    return response.data;
  }

  async deleteRole(id: string) {
    const response = await apiClient.delete(`/roles/${id}`);
    return response.data;
  }

  // 角色权限管理
  async getRolePermissions(roleId: string) {
    const response = await apiClient.get(`/roles/${roleId}/permissions`);
    const raw = response.data;
    const list = Array.isArray(raw)
      ? raw
      : Array.isArray(raw?.permissions)
        ? raw.permissions
        : Array.isArray(raw?.data?.permissions)
          ? raw.data.permissions
          : Array.isArray(raw?.data)
            ? raw.data
            : Array.isArray(raw?.items)
              ? raw.items
              : [];

    return { data: list };
  }

  async assignPermissionToRole(roleId: string, permissionId: string) {
    const response = await apiClient.post(`/roles/${roleId}/permissions`, {
      permission_id: permissionId,
    });
    return response.data;
  }

  async revokePermissionFromRole(roleId: string, permissionId: string) {
    const response = await apiClient.delete(`/roles/${roleId}/permissions/${permissionId}`);
    return response.data;
  }

  async batchAssignPermissionsToRole(roleId: string, permissionIds: string[]) {
    const response = await apiClient.post(`/roles/${roleId}/permissions/batch`, {
      permission_ids: permissionIds,
    });
    return response.data;
  }

  // 用户角色管理
  async getUserRoles(userId: string) {
    const response = await apiClient.get(`/user-roles/${userId}/roles`);
    const raw = response.data;
    const list = Array.isArray((raw as any)?.roles)
      ? (raw as any).roles
      : Array.isArray((raw as any)?.data?.roles)
        ? (raw as any).data.roles
        : Array.isArray(raw)
          ? (raw as any)
          : Array.isArray((raw as any)?.data)
            ? (raw as any).data
            : [];
    return { data: list };
  }

  async assignRoleToUser(userId: string, roleId: string) {
    const response = await apiClient.post(`/user-roles/${userId}/roles`, {
      role_ids: [roleId],
    });
    return response.data;
  }

  async revokeRoleFromUser(userId: string, roleId: string) {
    const response = await apiClient.delete(`/user-roles/${userId}/roles/${roleId}`);
    return response.data;
  }

  async batchAssignRolesToUser(userId: string, roleIds: string[]) {
    const response = await apiClient.post(`/user-roles/${userId}/roles`, {
      role_ids: roleIds,
    });
    return response.data;
  }

  // 用户权限管理
  async getUserPermissions(userId: string) {
    // 后端未提供直接的用户权限列表接口；
    // 兼容处理：通过用户角色获取权限并映射为 resource:action 代码数组
    const rolesResp = await this.getUserRoles(userId);
    const roles = Array.isArray(rolesResp?.data) ? rolesResp.data : [];
    const permissions: string[] = [];
    roles.forEach((role: any) => {
      const perms = Array.isArray(role?.permissions) ? role.permissions : [];
      perms.forEach((p: any) => {
        const res = p?.resource || p?.Resource;
        const act = p?.action || p?.Action;
        if (res && act) {
          permissions.push(`${res}:${act}`);
        }
      });
    });
    return { data: permissions };
  }

  async batchAssignPermissionsToUser(userId: string, permissionIds: string[]) {
    // 后端当前不支持直接分配用户权限；前端应通过角色进行权限管理。
    // 该方法保留以兼容旧调用，但不执行写操作。
    return { message: 'Direct user permission assignment is not supported. Use roles.' } as any;
  }

  async checkPermission(request: PermissionCheckRequest): Promise<PermissionCheckResponse> {
    const response = await apiClient.post('/permissions/check', request);
    return response.data;
  }

  async batchCheckPermissions(requests: PermissionCheckRequest[]) {
    const response = await apiClient.post('/permissions/check/batch', {
      requests,
    });
    return response.data;
  }

  // 获取当前用户权限
  async getCurrentUserPermissions() {
    // 兼容实现：通过 /auth/me 获取当前用户，再查询其角色并计算权限
    try {
      const meResp = await apiClient.get('/auth/me');
      const meRaw: any = (meResp as any)?.data || {};
      const meData = (meRaw?.data ?? meRaw) as any;
      const userId: string = meData?.user_id || meData?.id;
      if (!userId) return { data: [] };
      const permsResp = await this.getUserPermissions(userId);
      return permsResp;
    } catch {
      return { data: [] };
    }
  }

  async checkCurrentUserPermission(resource: string, action: string, resourceId?: string) {
    const response = await apiClient.post('/users/me/permissions/check', {
      resource,
      action,
      resource_id: resourceId,
    });
    return response.data;
  }

  // 权限树结构
  async getPermissionTree() {
    const response = await apiClient.get('/permissions/tree');
    return response.data;
  }

  async getRoleTree() {
    const response = await apiClient.get('/roles/tree');
    return response.data;
  }
}

export const permissionService = new PermissionService();