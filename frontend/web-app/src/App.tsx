import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Provider } from 'react-redux';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { store } from './store';
import MainLayout from './components/layout/MainLayout';
import AdminLayout from './components/layout/AdminLayout';
import Home from './pages/Home';
import Dashboard from './pages/Dashboard';
import Chat from './pages/Chat';
import ApiTest from './pages/ApiTest';
import Wisdom from './pages/Wisdom';
import WisdomDetail from './pages/WisdomDetail';
import RecommendationCenter from './pages/RecommendationCenter';
import Profile from './pages/Profile';
import Community from './pages/Community';
import UserFavorites from './pages/UserFavorites';
import UserNotes from './pages/UserNotes';
import IntelligentLearning from './pages/IntelligentLearning';
import LearningAnalyticsDashboard from './pages/LearningAnalyticsDashboard';
import CourseCenter from './pages/CourseCenter';
import AbilityAssessment from './pages/AbilityAssessment';
import LearningPlan from './pages/LearningPlan';
import DailyCheckin from './pages/DailyCheckin';
import AchievementCenter from './pages/AchievementCenter';
import ProjectManagement from './pages/ProjectManagement';
import ProjectWorkspace from './pages/projects/ProjectWorkspace';
import TaskManagement from './pages/projects/TaskManagement';
import TeamCollaboration from './pages/projects/TeamCollaboration';
import ProjectAnalytics from './pages/projects/ProjectAnalytics';
import HealthManagement from './pages/HealthManagement';
import SecurityCenter from './pages/SecurityCenter';
import PermissionTest from './components/common/PermissionTest';
import Login from './pages/Login';
import TestLogin from './pages/TestLogin';
import EmailVerification from './pages/EmailVerification';
// 管理员页面
import AdminDashboard from './pages/admin/AdminDashboard';
import WisdomManagement from './pages/admin/WisdomManagement';
import CategoryManagement from './pages/admin/CategoryManagement';
import TagManagement from './pages/admin/TagManagement';
import UserManagement from './pages/admin/UserManagement';
import DataAnalytics from './pages/admin/DataAnalytics';
import SystemSettings from './pages/admin/SystemSettings';
import NotificationCenter from './pages/admin/NotificationCenter';
import ContentReview from './pages/admin/ContentReview';
import WisdomEditor from './pages/admin/WisdomEditor';
import { useAuth } from './hooks/useAuth';
import './index.css';

// 受保护的路由组件
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();
  
  if (isLoading) {
    return <div>Loading...</div>;
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

// 管理员路由组件
const AdminRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, user, isLoading } = useAuth();
  
  if (isLoading) {
    return <div>Loading...</div>;
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  // 检查是否有管理员权限
  if (!user?.isAdmin && user?.role !== 'admin') {
    return <Navigate to="/" replace />;
  }
  
  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <ConfigProvider 
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: '#1890ff',
          },
        }}
      >
        <Router>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/test-login" element={<TestLogin />} />
            <Route path="/verify-email" element={<EmailVerification />} />
            
            {/* 管理员路由 */}
            <Route path="/admin/*" element={
              <AdminRoute>
                <AdminLayout>
                  <Routes>
                    <Route path="/" element={<AdminDashboard />} />
                    <Route path="/wisdom" element={<WisdomManagement />} />
                    <Route path="/wisdom/add" element={<WisdomEditor />} />
                    <Route path="/wisdom/edit/:id" element={<WisdomEditor />} />
                    <Route path="/categories" element={<CategoryManagement />} />
                    <Route path="/tags" element={<TagManagement />} />
                    <Route path="/users" element={<UserManagement />} />
                    <Route path="/analytics" element={<DataAnalytics />} />
                    <Route path="/settings" element={<SystemSettings />} />
                    <Route path="/notifications" element={<NotificationCenter />} />
                    <Route path="/review" element={<ContentReview />} />
                  </Routes>
                </AdminLayout>
              </AdminRoute>
            } />
            
            {/* 普通用户路由 */}
            <Route path="/*" element={
              <ProtectedRoute>
                <MainLayout>
                  <Routes>
                    <Route path="/" element={<Home />} />
                    <Route path="/dashboard" element={<Dashboard />} />
                    <Route path="/chat" element={<Chat />} />
                    <Route path="/api-test" element={<ApiTest />} />
                    <Route path="/wisdom" element={<Wisdom />} />
                    <Route path="/wisdom/:id" element={<WisdomDetail />} />
                    <Route path="/recommendations" element={<RecommendationCenter />} />
                    <Route path="/community" element={<Community />} />
                    <Route path="/intelligent-learning" element={<IntelligentLearning />} />
                    <Route path="/learning/analytics-dashboard" element={<LearningAnalyticsDashboard />} />
                    <Route path="/course-center" element={<CourseCenter />} />
                    <Route path="/ability-assessment" element={<AbilityAssessment />} />
                <Route path="/learning-plan" element={<LearningPlan />} />
                <Route path="/daily-checkin" element={<DailyCheckin />} />
                <Route path="/achievement-center" element={<AchievementCenter />} />
                <Route path="/project-management" element={<ProjectManagement />} />
                    <Route path="/projects/workspace" element={<ProjectWorkspace />} />
                    <Route path="/projects/tasks" element={<TaskManagement />} />
                    <Route path="/projects/collaboration" element={<TeamCollaboration />} />
                    <Route path="/projects/analytics" element={<ProjectAnalytics />} />
                    <Route path="/health-management" element={<HealthManagement />} />
                    <Route path="/security" element={<SecurityCenter />} />
                    <Route path="/permission-test" element={<PermissionTest />} />
                    <Route path="/profile" element={<Profile />} />
                    <Route path="/favorites" element={<UserFavorites />} />
                    <Route path="/notes" element={<UserNotes />} />
                  </Routes>
                </MainLayout>
              </ProtectedRoute>
            } />
          </Routes>
        </Router>
      </ConfigProvider>
    </Provider>
  );
};

export default App;
