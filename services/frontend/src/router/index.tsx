import React from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '../layouts/MainLayout';
import Dashboard from '../pages/Dashboard';
import ConfigManagement from '../pages/ConfigManagement';

// 老君域页面组件（暂时使用占位符）
const PluginManagement = React.lazy(() => import('../pages/PluginManagement'));
const AuditLogs = React.lazy(() => import('../pages/AuditLogs'));

// 太上域页面组件（暂时使用占位符）
const ModelManagement = React.lazy(() => import('../pages/ModelManagement'));
const CollectionManagement = React.lazy(() => import('../pages/CollectionManagement'));
const TaskManagement = React.lazy(() => import('../pages/TaskManagement'));

const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        index: true,
        element: <Dashboard />,
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
            element: <ConfigManagement />,
          },
          {
            path: 'plugins',
            element: (
              <React.Suspense fallback={<div>加载中...</div>}>
                <PluginManagement />
              </React.Suspense>
            ),
          },
          {
            path: 'audit-logs',
            element: (
              <React.Suspense fallback={<div>加载中...</div>}>
                <AuditLogs />
              </React.Suspense>
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
              <React.Suspense fallback={<div>加载中...</div>}>
                <ModelManagement />
              </React.Suspense>
            ),
          },
          {
            path: 'collections',
            element: (
              <React.Suspense fallback={<div>加载中...</div>}>
                <CollectionManagement />
              </React.Suspense>
            ),
          },
          {
            path: 'tasks',
            element: (
              <React.Suspense fallback={<div>加载中...</div>}>
                <TaskManagement />
              </React.Suspense>
            ),
          },
        ],
      },
    ],
  },
]);

export default router;