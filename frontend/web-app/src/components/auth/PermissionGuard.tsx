import React from 'react';
import { usePermissions } from '../../hooks/usePermissions';

interface PermissionGuardProps {
  children: React.ReactNode;
  permission?: string;
  permissions?: string[];
  role?: string;
  roles?: string[];
  requireAll?: boolean; // 是否需要所有权限/角色
  fallback?: React.ReactNode;
  inverse?: boolean; // 反向权限检查
}

export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  children,
  permission,
  permissions = [],
  role,
  roles = [],
  requireAll = false,
  fallback = null,
  inverse = false,
}) => {
  const {
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAllPermissions,
    hasAnyRole,
  } = usePermissions();

  let hasAccess = true;

  // 检查单个权限
  if (permission) {
    hasAccess = hasPermission(permission);
  }

  // 检查多个权限
  if (permissions.length > 0) {
    hasAccess = requireAll 
      ? hasAllPermissions(permissions)
      : hasAnyPermission(permissions);
  }

  // 检查单个角色
  if (role) {
    hasAccess = hasAccess && hasRole(role);
  }

  // 检查多个角色
  if (roles.length > 0) {
    const roleAccess = requireAll
      ? roles.every(r => hasRole(r))
      : hasAnyRole(roles);
    hasAccess = hasAccess && roleAccess;
  }

  // 反向权限检查
  if (inverse) {
    hasAccess = !hasAccess;
  }

  return hasAccess ? <>{children}</> : <>{fallback}</>;
};

// 高阶组件版本
export const withPermission = (
  Component: React.ComponentType<any>,
  permissionConfig: Omit<PermissionGuardProps, 'children'>
) => {
  return (props: any) => (
    <PermissionGuard {...permissionConfig}>
      <Component {...props} />
    </PermissionGuard>
  );
};

// 权限检查装饰器
export const requirePermission = (permission: string) => {
  return (Component: React.ComponentType<any>) => {
    return withPermission(Component, { permission });
  };
};

export const requireRole = (role: string) => {
  return (Component: React.ComponentType<any>) => {
    return withPermission(Component, { role });
  };
};

export const requireAdmin = (Component: React.ComponentType<any>) => {
  return withPermission(Component, { roles: ['admin', 'super_admin'] });
};