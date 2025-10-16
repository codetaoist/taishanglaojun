import { useCallback, useMemo } from 'react';
import { useAuthContext } from '../contexts/AuthContext';

export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
}

export interface Role {
  id: string;
  name: string;
  code: string;
  description?: string;
  permissions: string[];
  level: number;
}

export const usePermissions = () => {
  const { user, isLoading } = useAuthContext();

  // 检查是否已认证
  const isAuthenticated = useMemo(() => {
    return !!user;
  }, [user]);

  // 检查用户是否有特定权限
  const hasPermission = useCallback((permission: string): boolean => {
    if (!user) return false;
    
    // 超级管理员拥有所有权限
    if (user.roles?.some(role => role.toLowerCase() === 'super_admin')) return true;
    
    // 检查用户权限列表
    return user.permissions?.includes(permission) || false;
  }, [user]);

  // 检查用户是否具有指定角色（忽略大小写）
  const hasRole = useCallback((role: string): boolean => {
    if (!user) return false;
    const normalizedRole = role.toLowerCase();
    const rolesLower = (user.roles ?? []).map(r => r.toLowerCase());
    const singleRoleLower = user.role?.toLowerCase();

    if (rolesLower.includes(normalizedRole)) return true;
    if (singleRoleLower === normalizedRole) return true;

    // 兼容 isAdmin 标记和历史别名 'administrator'
    if (['admin', 'super_admin', 'administrator'].includes(normalizedRole) && user.isAdmin) return true;

    return false;
  }, [user]);

  // 检查用户是否有任一权限
  const hasAnyPermission = useCallback((permissions: string[]): boolean => {
    if (!user) return false;
    if (user.roles?.some(role => role.toLowerCase() === 'super_admin')) return true;
    
    return permissions.some(permission => 
      user.permissions?.includes(permission)
    );
  }, [user]);

  // 检查用户是否有所有权限
  const hasAllPermissions = useCallback((permissions: string[]): boolean => {
    if (!user) return false;
    if (user.roles?.some(role => role.toLowerCase() === 'super_admin')) return true;
    
    return permissions.every(permission => 
      user.permissions?.includes(permission)
    );
  }, [user]);

  // 检查用户是否具有任意一个指定角色（忽略大小写）
  const hasAnyRole = useCallback((roles: string[]): boolean => {
    if (!user) return false;
    const normalizedRoles = roles.map(role => role.toLowerCase());
    const rolesLower = (user.roles ?? []).map(r => r.toLowerCase());
    const singleRoleLower = user.role?.toLowerCase();

    if (rolesLower.some(userRole => normalizedRoles.includes(userRole))) return true;
    if (singleRoleLower && normalizedRoles.includes(singleRoleLower)) return true;

    // 兼容 isAdmin 标记和历史别名 'administrator'
    if (user.isAdmin && normalizedRoles.some(r => ['admin', 'super_admin', 'administrator'].includes(r))) return true;

    return false;
  }, [user]);

  // 检查用户是否具有所有指定角色（忽略大小写）
  const hasAllRoles = useCallback((roles: string[]): boolean => {
    if (!user) return false;
    const normalizedRoles = roles.map(role => role.toLowerCase());
    const rolesLower = (user.roles ?? []).map(r => r.toLowerCase());
    const singleRoleLower = user.role?.toLowerCase();

    return normalizedRoles.every(role =>
      rolesLower.includes(role) ||
      singleRoleLower === role ||
      (user.isAdmin && ['admin', 'super_admin', 'administrator'].includes(role))
    );
  }, [user]);

  // 检查资源权限（资源:操作格式）
  const hasResourcePermission = useCallback((resource: string, action: string): boolean => {
    const permission = `${resource}:${action}`;
    return hasPermission(permission);
  }, [hasPermission]);

  // 获取用户所有权限
  const userPermissions = useMemo(() => {
    return user?.permissions || [];
  }, [user?.permissions]);

  // 获取用户所有角色
  const userRoles = useMemo(() => {
    return user?.roles || [];
  }, [user?.roles]);

  // 检查是否为管理员
  const isAdmin = useMemo(() => {
    return !!user && (user.isAdmin || hasAnyRole(['admin', 'super_admin', 'administrator']));
  }, [user, hasAnyRole]);

  // 检查是否为超级管理员
  const isSuperAdmin = useMemo(() => {
    return hasRole('super_admin');
  }, [hasRole]);

  return {
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAllPermissions,
    hasAnyRole,
    hasAllRoles,
    hasResourcePermission,
    userPermissions,
    userRoles,
    isAdmin,
    isSuperAdmin,
    isAuthenticated,
    isLoading,
  };
};

// 权限常量定义
export const PERMISSIONS = {
  // 用户管理
  USER_READ: 'user:read',
  USER_WRITE: 'user:write',
  USER_DELETE: 'user:delete',
  USER_MANAGE: 'user:manage',

  // 角色管理
  ROLE_READ: 'role:read',
  ROLE_WRITE: 'role:write',
  ROLE_DELETE: 'role:delete',
  ROLE_MANAGE: 'role:manage',

  // 权限管理
  PERMISSION_READ: 'permission:read',
  PERMISSION_WRITE: 'permission:write',
  PERMISSION_DELETE: 'permission:delete',
  PERMISSION_MANAGE: 'permission:manage',

  // 菜单管理
  MENU_READ: 'menu:read',
  MENU_WRITE: 'menu:write',
  MENU_DELETE: 'menu:delete',
  MENU_MANAGE: 'menu:manage',

  // 系统管理
  SYSTEM_READ: 'system:read',
  SYSTEM_WRITE: 'system:write',
  SYSTEM_MANAGE: 'system:manage',

  // AI功能
  AI_CHAT: 'ai:chat',
  AI_IMAGE: 'ai:image',
  AI_VIDEO: 'ai:video',
  AI_AUDIO: 'ai:audio',
  AI_ANALYSIS: 'ai:analysis',

  // 内容管理
  CONTENT_READ: 'content:read',
  CONTENT_WRITE: 'content:write',
  CONTENT_DELETE: 'content:delete',
  CONTENT_MANAGE: 'content:manage',

  // 智慧库
  WISDOM_READ: 'wisdom:read',
  WISDOM_WRITE: 'wisdom:write',
  WISDOM_DELETE: 'wisdom:delete',
  WISDOM_MANAGE: 'wisdom:manage',
} as const;

// 角色常量定义
export const ROLES = {
  SUPER_ADMIN: 'super_admin',
  ADMIN: 'admin',
  MODERATOR: 'moderator',
  USER: 'user',
  GUEST: 'guest',
} as const;