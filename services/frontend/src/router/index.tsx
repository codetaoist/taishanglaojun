import React, { Suspense } from 'react';
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { Spin } from 'antd';
import MainLayout from '../layouts/MainLayout';
import AuthPage from '../pages/AuthPage';
import ProtectedRoute from '../components/ProtectedRoute';

// 懒加载页面组件
const Dashboard = React.lazy(() => import('../pages/Dashboard'));
const ConfigManagement = React.lazy(() => import('../pages/ConfigManagement'));
const PluginManagement = React.lazy(() => import('../pages/PluginManagement'));
const AuditLogs = React.lazy(() => import('../pages/AuditLogs'));
const ModelManagement = React.lazy(() => import('../pages/ModelManagement'));
const CollectionManagement = React.lazy(() => import('../pages/CollectionManagement'));
const VectorDataManagement = React.lazy(() => import('../pages/VectorDataManagement'));
const VectorMonitor = React.lazy(() => import('../pages/taishang/VectorMonitor'));
const TaskManagement = React.lazy(() => import('../pages/TaskManagement'));
const TokenManager = React.lazy(() => import('../pages/TokenManager'));

// 加载组件
const PageLoading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}>
    <Spin size="large" />
  </div>
);

// 未授权页面
const Unauthorized = () => (
  <div style={{ padding: '50px', textAlign: 'center' }}>
    <h1>403 - 未授权访问</h1>
    <p>您没有权限访问此页面</p>
  </div>
);

// 创建路由配置
const router = createBrowserRouter([
  {
    path: '/login',
    element: <AuthPage />,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <MainLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: (
          <Suspense fallback={<PageLoading />}>
            <Dashboard />
          </Suspense>
        ),
      },
      {
        path: 'token-manager',
        element: (
          <Suspense fallback={<PageLoading />}>
            <TokenManager />
          </Suspense>
        ),
      },
      {
        path: 'laojun',
        children: [
          {
            index: true,
            element: <Navigate to="/laojun/config" replace />,
          },
          {
            path: 'config',
            element: (
              <Suspense fallback={<PageLoading />}>
                <ConfigManagement />
              </Suspense>
            ),
          },
          {
            path: 'plugins',
            element: (
              <Suspense fallback={<PageLoading />}>
                <PluginManagement />
              </Suspense>
            ),
          },
          {
            path: 'audit-logs',
            element: (
              <ProtectedRoute requireAdmin>
                <Suspense fallback={<PageLoading />}>
                  <AuditLogs />
                </Suspense>
              </ProtectedRoute>
            ),
          },
        ],
      },
      {
        path: 'taishang',
        children: [
          {
            index: true,
            element: <Navigate to="/taishang/models" replace />,
          },
          {
            path: 'models',
            element: (
              <Suspense fallback={<PageLoading />}>
                <ModelManagement />
              </Suspense>
            ),
          },
          {
            path: 'collections',
            element: (
              <Suspense fallback={<PageLoading />}>
                <CollectionManagement />
              </Suspense>
            ),
          },
          {
            path: 'vector-data',
            element: (
              <Suspense fallback={<PageLoading />}>
                <VectorDataManagement />
              </Suspense>
            ),
          },
          {
            path: 'vector-monitor',
            element: (
              <Suspense fallback={<PageLoading />}>
                <VectorMonitor />
              </Suspense>
            ),
          },
          {
            path: 'tasks',
            element: (
              <Suspense fallback={<PageLoading />}>
                <TaskManagement />
              </Suspense>
            ),
          },
        ],
      },
    ],
  },
  {
    path: '/unauthorized',
    element: <Unauthorized />,
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

const AppRouter: React.FC = () => {
  return <RouterProvider router={router} />;
};

export default AppRouter;