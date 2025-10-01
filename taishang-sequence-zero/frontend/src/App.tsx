import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import zhCN from 'antd/locale/zh_CN';
import enUS from 'antd/locale/en_US';

// Redux store
import { store, persistor } from './store';
import { useAppSelector, useAppDispatch } from './hooks/redux';
import { checkAuthStatus } from './store/slices/authSlice';

// Layouts
import MainLayout from './components/Layout/MainLayout';
import AuthLayout from './components/Layout/AuthLayout';

// Route Guards
import ProtectedRoute from './components/Auth/ProtectedRoute';
import AdminRoute from './components/Auth/AdminRoute';

// Pages
import LoginPage from './pages/Auth/LoginPage';
import RegisterPage from './pages/Auth/RegisterPage';
import ForgotPasswordPage from './pages/Auth/ForgotPasswordPage';
import DashboardPage from './pages/Dashboard/DashboardPage';
import ConsciousnessPage from './pages/Consciousness/ConsciousnessPage';
import CulturalPage from './pages/Cultural/CulturalPage';
import ProfilePage from './pages/Profile/ProfilePage';
import AdminPage from './pages/Admin/AdminPage';

// Global Styles
import './styles/index.css';
import './styles/themes.css';
import './styles/global.css';
import './styles/error.css';

// Loading component
const LoadingSpinner: React.FC = () => (
  <div className="loading-container">
    <div className="loading-spinner">
      <div className="spinner"></div>
      <p>加载中...</p>
    </div>
  </div>
);

// Error Boundary
class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error?: Error }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="error-boundary">
          <div className="error-content">
            <h1>出现了一些问题</h1>
            <p>应用程序遇到了意外错误，请刷新页面重试。</p>
            <button 
              onClick={() => window.location.reload()}
              className="error-button"
            >
              刷新页面
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// App Content Component (needs to be inside Redux Provider)
const AppContent: React.FC = () => {
  const dispatch = useAppDispatch();
  const { isAuthenticated, loading: authLoading } = useAppSelector(state => state.auth);
  const { theme, language } = useAppSelector(state => state.ui);

  // Check authentication status on app load
  useEffect(() => {
    dispatch(checkAuthStatus());
  }, [dispatch]);

  // Apply theme to document
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  // Get Ant Design locale
  const getAntdLocale = () => {
    return language === 'en-US' ? enUS : zhCN;
  };

  // Show loading spinner during initial auth check
  if (authLoading) {
    return <LoadingSpinner />;
  }
  return (
    <ConfigProvider 
      locale={getAntdLocale()}
      theme={{
        token: {
          colorPrimary: '#1890ff',
          borderRadius: 8,
          fontSize: 14,
        },
        algorithm: theme === 'dark' ? undefined : undefined, // Can add dark algorithm here
      }}
    >
      <AntdApp>
        <Router>
          <Routes>
            {/* Public Routes - Auth Pages */}
            <Route path="/" element={
              isAuthenticated ? <Navigate to="/dashboard" replace /> : <Navigate to="/login" replace />
            } />
            
            <Route path="/login" element={
              isAuthenticated ? <Navigate to="/dashboard" replace /> : (
                <AuthLayout>
                  <LoginPage />
                </AuthLayout>
              )
            } />
            
            <Route path="/register" element={
              isAuthenticated ? <Navigate to="/dashboard" replace /> : (
                <AuthLayout>
                  <RegisterPage />
                </AuthLayout>
              )
            } />
            
            <Route path="/forgot-password" element={
              isAuthenticated ? <Navigate to="/dashboard" replace /> : (
                <AuthLayout>
                  <ForgotPasswordPage />
                </AuthLayout>
              )
            } />

            {/* Protected Routes - Main App */}
            <Route path="/dashboard" element={
              <ProtectedRoute>
                <MainLayout>
                  <DashboardPage />
                </MainLayout>
              </ProtectedRoute>
            } />
            
            <Route path="/consciousness" element={
              <ProtectedRoute>
                <MainLayout>
                  <ConsciousnessPage />
                </MainLayout>
              </ProtectedRoute>
            } />
            
            <Route path="/cultural" element={
              <ProtectedRoute>
                <MainLayout>
                  <CulturalPage />
                </MainLayout>
              </ProtectedRoute>
            } />
            
            <Route path="/profile" element={
              <ProtectedRoute>
                <MainLayout>
                  <ProfilePage />
                </MainLayout>
              </ProtectedRoute>
            } />

            {/* Admin Routes */}
            <Route path="/admin" element={
              <AdminRoute>
                <MainLayout>
                  <AdminPage />
                </MainLayout>
              </AdminRoute>
            } />

            {/* Error Pages */}
            <Route path="/403" element={
              <div className="error-page">
                <div className="error-content">
                  <h1>403</h1>
                  <h2>访问被拒绝</h2>
                  <p>您没有权限访问此页面。</p>
                  <button onClick={() => window.history.back()} className="error-button">
                    返回上一页
                  </button>
                </div>
              </div>
            } />
            
            <Route path="/404" element={
              <div className="error-page">
                <div className="error-content">
                  <h1>404</h1>
                  <h2>页面未找到</h2>
                  <p>您访问的页面不存在。</p>
                  <button onClick={() => window.location.href = '/dashboard'} className="error-button">
                    返回首页
                  </button>
                </div>
              </div>
            } />

            {/* Catch all route - redirect to 404 */}
            <Route path="*" element={<Navigate to="/404" replace />} />
          </Routes>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
};

// Main App Component
const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <Provider store={store}>
        <PersistGate loading={<LoadingSpinner />} persistor={persistor}>
          <AppContent />
        </PersistGate>
      </Provider>
    </ErrorBoundary>
  );
};

export default App;