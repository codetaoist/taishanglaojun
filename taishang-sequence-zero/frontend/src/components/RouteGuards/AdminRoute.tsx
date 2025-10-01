import React from 'react';
import { Navigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { RootState } from '../../store';

interface AdminRouteProps {
  children: React.ReactNode;
}

const AdminRoute: React.FC<AdminRouteProps> = ({ children }) => {
  const { isAuthenticated, user } = useSelector((state: RootState) => state.auth);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // 检查用户是否有管理员权限
  if (!user || (user.role !== 'admin' && user.role !== 'super_admin')) {
    return <Navigate to="/403" replace />;
  }

  return <>{children}</>;
};

export default AdminRoute;