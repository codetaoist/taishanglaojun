import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Spin } from 'antd';
import { useAppSelector } from '../../hooks/redux';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: string;
  requiredPermissions?: string[];
  fallbackPath?: string;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requiredRole,
  requiredPermissions = [],
  fallbackPath = '/login',
}) => {
  const location = useLocation();
  const { isAuthenticated, loading, user } = useAppSelector(state => state.auth);

  // 如果正在加载认证状态，显示加载器
  if (loading) {
    return (
      <div className="protected-route-loading">
        <Spin size="large" />
        <p>验证身份中...</p>
      </div>
    );
  }

  // 如果未认证，重定向到登录页面
  if (!isAuthenticated) {
    return (
      <Navigate 
        to={fallbackPath} 
        state={{ from: location }} 
        replace 
      />
    );
  }

  // 检查角色权限
  if (requiredRole && user) {
    const hasRequiredRole = user.role === requiredRole;
    if (!hasRequiredRole) {
      return (
        <Navigate 
          to="/dashboard" 
          state={{ 
            error: '您没有访问此页面的权限',
            from: location 
          }} 
          replace 
        />
      );
    }
  }

  // 检查具体权限
  if (requiredPermissions.length > 0 && user) {
    const hasAllPermissions = requiredPermissions.every(permission => 
      (user as any).permissions?.includes(permission)
    );
    
    if (!hasAllPermissions) {
      return (
        <Navigate 
          to="/dashboard" 
          state={{ 
            error: '您没有执行此操作的权限',
            from: location 
          }} 
          replace 
        />
      );
    }
  }

  // 如果所有检查都通过，渲染子组件
  return (
    <>
      {children}
      <style>{`
        .protected-route-loading {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100vh;
          background: var(--bg-secondary);
          color: var(--text-primary);
        }

        .protected-route-loading p {
          margin-top: 16px;
          font-size: 16px;
          color: var(--text-secondary);
        }
      `}</style>
    </>
  );
};

export default ProtectedRoute;