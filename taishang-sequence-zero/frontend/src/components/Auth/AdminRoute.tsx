import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Result, Button } from 'antd';
import { useAppSelector } from '../../hooks/redux';
import { useNavigate } from 'react-router-dom';

interface AdminRouteProps {
  children: React.ReactNode;
  requiredPermissions?: string[];
}

const AdminRoute: React.FC<AdminRouteProps> = ({
  children,
  requiredPermissions = [],
}) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { isAuthenticated, user, loading } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);

  // 如果正在加载，不做任何处理
  if (loading) {
    return null;
  }

  // 如果未认证，重定向到登录页面
  if (!isAuthenticated) {
    return (
      <Navigate 
        to="/login" 
        state={{ from: location }} 
        replace 
      />
    );
  }

  // 检查是否是管理员
  const isAdmin = user?.role === 'admin' || user?.role === 'super_admin';
  
  if (!isAdmin) {
    return (
      <div className="admin-route-error">
        <Result
          status="403"
          title="403"
          subTitle={
            language === 'en-US'
              ? "Sorry, you don't have permission to access this page."
              : "抱歉，您没有权限访问此页面。"
          }
          extra={
            <Button 
              type="primary" 
              onClick={() => navigate('/dashboard')}
            >
              {language === 'en-US' ? 'Back to Dashboard' : '返回仪表板'}
            </Button>
          }
        />
        <style>{`
          .admin-route-error {
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 60vh;
            background: var(--bg-primary);
          }
        `}</style>
      </div>
    );
  }

  // 检查具体的管理员权限
  if (requiredPermissions.length > 0) {
    const hasAllPermissions = requiredPermissions.every(permission => 
      (user as any)?.permissions?.includes(permission)
    );
    
    if (!hasAllPermissions) {
      return (
        <div className="admin-route-error">
          <Result
            status="403"
            title="403"
            subTitle={
              language === 'en-US'
                ? "You don't have the required permissions for this operation."
                : "您没有执行此操作所需的权限。"
            }
            extra={
              <Button 
                type="primary" 
                onClick={() => navigate('/admin')}
              >
                {language === 'en-US' ? 'Back to Admin' : '返回管理页面'}
              </Button>
            }
          />
          <style>{`
            .admin-route-error {
              display: flex;
              align-items: center;
              justify-content: center;
              min-height: 60vh;
              background: var(--bg-primary);
            }
          `}</style>
        </div>
      );
    }
  }

  // 如果所有检查都通过，渲染子组件
  return <>{children}</>;
};

export default AdminRoute;