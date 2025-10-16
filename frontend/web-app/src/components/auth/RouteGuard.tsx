/**
 * 路由守卫组件
 * 用于保护需要特定权限或角色的路由
 */

import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Result, Button } from 'antd';
import { usePermissions } from '../../hooks/usePermissions';

interface RouteGuardProps {
  children: React.ReactNode;
  // 权限要求
  permission?: string;
  permissions?: string[];
  requireAllPermissions?: boolean;
  // 角色要求
  role?: string;
  roles?: string[];
  requireAllRoles?: boolean;
  // 资源权限
  resource?: string;
  action?: string;
  // 其他配置
  redirectTo?: string;
  fallback?: React.ReactNode;
  showFallback?: boolean;
}

export const RouteGuard: React.FC<RouteGuardProps> = ({
  children,
  permission,
  permissions,
  requireAllPermissions = false,
  role,
  roles,
  requireAllRoles = false,
  resource,
  action,
  redirectTo = '/login',
  fallback,
  showFallback = true
}) => {
  const location = useLocation();
  const {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    hasRole,
    hasAnyRole,
    hasAllRoles,
    hasResourcePermission,
    isAuthenticated,
    isLoading
  } = usePermissions();

  // 加载中状态
  if (isLoading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '200px' 
      }}>
        加载中...
      </div>
    );
  }

  // 未登录，重定向到登录页
  if (!isAuthenticated) {
    return <Navigate to={redirectTo} state={{ from: location }} replace />;
  }

  // 检查权限
  let hasRequiredPermission = true;

  // 单个权限检查
  if (permission) {
    hasRequiredPermission = hasPermission(permission);
  }

  // 多个权限检查
  if (permissions && permissions.length > 0) {
    if (requireAllPermissions) {
      hasRequiredPermission = hasAllPermissions(permissions);
    } else {
      hasRequiredPermission = hasAnyPermission(permissions);
    }
  }

  // 资源权限检查
  if (resource && action) {
    hasRequiredPermission = hasResourcePermission(resource, action);
  }

  // 检查角色
  let hasRequiredRole = true;

  // 单个角色检查
  if (role) {
    hasRequiredRole = hasRole(role);
  }

  // 多个角色检查
  if (roles && roles.length > 0) {
    if (requireAllRoles) {
      hasRequiredRole = hasAllRoles(roles);
    } else {
      hasRequiredRole = hasAnyRole(roles);
    }
  }

  // 权限或角色不足
  if (!hasRequiredPermission || !hasRequiredRole) {
    if (fallback) {
      return <>{fallback}</>;
    }

    if (showFallback) {
      return (
        <Result
          status="403"
          title="403"
          subTitle="抱歉，您没有权限访问此页面。"
          extra={
            <Button type="primary" onClick={() => window.history.back()}>
              返回上一页
            </Button>
          }
        />
      );
    }

    return <Navigate to="/403" replace />;
  }

  return <>{children}</>;
};

/**
 * 管理员路由守卫
 */
export const AdminRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <RouteGuard>
      {children}
    </RouteGuard>
  );
};

/**
 * 受保护的路由守卫（需要登录）
 */
export const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <RouteGuard>
      {children}
    </RouteGuard>
  );
};

/**
 * 高阶组件：为组件添加权限检查
 */
export const withRouteGuard = <P extends object>(
  Component: React.ComponentType<P>,
  guardProps: Omit<RouteGuardProps, 'children'>
) => {
  return (props: P) => (
    <RouteGuard {...guardProps}>
      <Component {...props} />
    </RouteGuard>
  );
};

/**
 * 权限检查装饰器
 */
export const requirePermission = (permission: string) => 
  <P extends object>(Component: React.ComponentType<P>) =>
    withRouteGuard(Component, { permission });

export const requireRole = (role: string) => 
  <P extends object>(Component: React.ComponentType<P>) =>
    withRouteGuard(Component, { role });

export const requireAdmin = <P extends object>(Component: React.ComponentType<P>) =>
  withRouteGuard(Component, { roles: ['admin', 'super_admin', 'administrator'] });

export default RouteGuard;