import { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { DynamicSidebar } from './components/DynamicSidebar';
import Header from './components/Header';
import ChatPage from './pages/ChatPage';
import DocumentPage from './pages/DocumentPage';
import FileTransferPage from './pages/FileTransferPage';
import ImagePage from './pages/ImagePage';
import PetPage from './pages/PetPage';
import SettingsPage from './pages/SettingsPage';
import SystemPage from './pages/SystemPage';
import TestPage from './pages/TestPage';
import TauriTestPage from './pages/TauriTestPage';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import UserManagementPage from './pages/UserManagementPage';
import MenuAdaptationTestPage from './pages/MenuAdaptationTestPage';
import AuthManagementPage from './pages/AuthManagementPage';
import FriendManagementPage from './pages/FriendManagementPage';
import ProjectManagementPage from './pages/ProjectManagementPage';
import AppManagementPage from './pages/AppManagementPage';
import ChatManagementPage from './pages/ChatManagementPage';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ModuleProvider } from './contexts/ModuleContext';
import { DynamicMenuProvider } from './contexts/DynamicMenuContext';
import { ThemeProvider } from './contexts/ThemeContext';
import ErrorBoundary from './components/ErrorBoundary';

function AppContent() {
  const { isAuthenticated, loading } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gradient-to-br from-background to-background/95">
        <div className="text-center space-y-4">
          <div className="relative">
            <div className="animate-spin rounded-full h-16 w-16 border-4 border-primary/20 border-t-primary mx-auto"></div>
            <div className="absolute inset-0 rounded-full bg-gradient-to-r from-primary/10 to-primary/5 animate-pulse"></div>
          </div>
          <div className="space-y-2">
            <p className="text-lg font-medium text-foreground">太上老君AI助手</p>
            <p className="text-sm text-muted-foreground">正在启动中...</p>
          </div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return (
    <div className="h-screen flex bg-gradient-to-br from-background via-background/98 to-background/95 overflow-hidden">
      {/* 动态侧边栏 */}
      <DynamicSidebar 
        open={sidebarOpen || !sidebarCollapsed}
        onToggle={() => {
          if (window.innerWidth >= 1024) {
            setSidebarCollapsed(!sidebarCollapsed);
          } else {
            setSidebarOpen(!sidebarOpen);
          }
        }}
        variant={window.innerWidth >= 1024 ? 'persistent' : 'temporary'}
      />
      
      {/* 主内容区域 */}
      <div className="flex-1 flex flex-col overflow-hidden min-w-0">
        {/* 头部 */}
        <Header 
          onToggleSidebar={() => {
            if (window.innerWidth >= 1024) {
              setSidebarCollapsed(!sidebarCollapsed);
            } else {
              setSidebarOpen(!sidebarOpen);
            }
          }}
          sidebarCollapsed={sidebarCollapsed}
        />
        
        {/* 主内容 */}
        <main className="flex-1 overflow-auto bg-gradient-to-br from-background/50 to-background/30 backdrop-blur-sm">
          <div className="container mx-auto p-4 lg:p-6 xl:p-8 max-w-7xl">
            <div className="animate-in slide-in-from-bottom-4 duration-500">
              <Routes>
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/menu-test" element={<MenuAdaptationTestPage />} />
                <Route path="/chat" element={<ChatPage />} />
                <Route path="/document" element={<DocumentPage />} />
                <Route path="/file-transfer" element={<FileTransferPage />} />
                <Route path="/image" element={<ImagePage />} />
                <Route path="/pet" element={<PetPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/system" element={<SystemPage />} />
                <Route path="/test" element={<TestPage />} />
                <Route path="/tauri-test" element={<TauriTestPage />} />
                <Route path="/admin/users" element={<UserManagementPage />} />
                <Route path="/admin/permissions" element={<UserManagementPage />} />
                <Route path="/auth-management" element={<AuthManagementPage />} />
                <Route path="/friend-management" element={<FriendManagementPage />} />
                <Route path="/project-management" element={<ProjectManagementPage />} />
                <Route path="/app-management" element={<AppManagementPage />} />
                <Route path="/chat-management" element={<ChatManagementPage />} />
              </Routes>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <ThemeProvider>
          <AuthProvider>
            <ModuleProvider>
              <DynamicMenuProvider>
                <AppContent />
              </DynamicMenuProvider>
            </ModuleProvider>
          </AuthProvider>
        </ThemeProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;