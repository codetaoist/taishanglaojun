import React, { Suspense, lazy, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useAuthContext as useAuth } from './contexts/AuthContext';
import { AdminRoute } from './components/auth/RouteGuard';
import { MenuProvider } from './contexts/MenuContext';
import GlobalErrorBoundary from './components/ErrorBoundary/GlobalErrorBoundary';
import { PageLoading } from './components/Loading/GlobalLoading';
import { useNotificationService } from './services/notificationService';
import { 
  smartLazy, 
  highPriorityConfig, 
  defaultLazyConfig, 
  lowPriorityConfig,
  RoutePreloader,
  preloadComponents,
  useSmartPreload
} from './utils/lazyLoadOptimization';
import './index.css';

// 布局组件 - 保持直接导入，因为它们是核心组件
import MainLayout from './components/layout/MainLayout';

// 认证相关页面 - 保持直接导入，因为它们是首屏必需的
import Login from './pages/Login';
import EmailVerification from './pages/EmailVerification';

// 智能懒加载页面组件
// 主要页面 - 高优先级预加载
const Dashboard = smartLazy(() => import('./pages/Dashboard'), { ...highPriorityConfig, chunkName: 'dashboard' });
const Chat = smartLazy(() => import('./pages/Chat'), { ...highPriorityConfig, chunkName: 'chat' });
const Profile = smartLazy(() => import('./pages/Profile'), { ...defaultLazyConfig, chunkName: 'profile' });
const Help = smartLazy(() => import('./pages/Help'), { ...lowPriorityConfig, chunkName: 'help' });

// 智慧相关页面
const Wisdom = smartLazy(() => import('./pages/Wisdom'), { ...defaultLazyConfig, chunkName: 'wisdom' });
const WisdomDetail = smartLazy(() => import('./pages/WisdomDetail'), { ...defaultLazyConfig, chunkName: 'wisdom-detail' });
const RecommendationCenter = smartLazy(() => import('./pages/RecommendationCenter'), { ...defaultLazyConfig, chunkName: 'recommendation' });

// 社区相关页面
const Community = smartLazy(() => import('./pages/Community'), { ...defaultLazyConfig, chunkName: 'community' });
const UserFavorites = smartLazy(() => import('./pages/UserFavorites'), { ...lowPriorityConfig, chunkName: 'favorites' });
const UserNotes = smartLazy(() => import('./pages/UserNotes'), { ...lowPriorityConfig, chunkName: 'notes' });

// 学习相关页面
const IntelligentLearning = smartLazy(() => import('./pages/IntelligentLearning'), { ...defaultLazyConfig, chunkName: 'intelligent-learning' });
const LearningAnalyticsDashboard = smartLazy(() => import('./pages/LearningAnalyticsDashboard'), { ...defaultLazyConfig, chunkName: 'learning-analytics' });
const CourseCenter = smartLazy(() => import('./pages/CourseCenter'), { ...defaultLazyConfig, chunkName: 'course-center' });
const AbilityAssessment = smartLazy(() => import('./pages/AbilityAssessment'), { ...defaultLazyConfig, chunkName: 'ability-assessment' });
const LearningPlan = smartLazy(() => import('./pages/LearningPlan'), { ...defaultLazyConfig, chunkName: 'learning-plan' });
const DailyCheckin = smartLazy(() => import('./pages/DailyCheckin'), { ...lowPriorityConfig, chunkName: 'daily-checkin' });
const AchievementCenter = smartLazy(() => import('./pages/AchievementCenter'), { ...lowPriorityConfig, chunkName: 'achievement' });
const LearningProgress = smartLazy(() => import('./pages/learning/LearningProgress'), { ...defaultLazyConfig, chunkName: 'learning-progress' });
const LearningCourses = smartLazy(() => import('./pages/learning/LearningCourses'), { ...defaultLazyConfig, chunkName: 'learning-courses' });

// 项目管理相关页面
const ProjectManagement = smartLazy(() => import('./pages/ProjectManagement'), { ...defaultLazyConfig, chunkName: 'project-management' });
const ProjectWorkspace = smartLazy(() => import('./pages/projects/ProjectWorkspace'), { ...defaultLazyConfig, chunkName: 'project-workspace' });
const TaskManagement = smartLazy(() => import('./pages/projects/TaskManagement'), { ...defaultLazyConfig, chunkName: 'task-management' });
const TeamCollaboration = smartLazy(() => import('./pages/projects/TeamCollaboration'), { ...defaultLazyConfig, chunkName: 'team-collaboration' });
const ProjectAnalytics = smartLazy(() => import('./pages/projects/ProjectAnalytics'), { ...lowPriorityConfig, chunkName: 'project-analytics' });

// 健康管理相关页面
const HealthManagement = smartLazy(() => import('./pages/HealthManagement'), { ...defaultLazyConfig, chunkName: 'health-management' });
const HealthAdvice = smartLazy(() => import('./pages/health/HealthAdvice'), { ...defaultLazyConfig, chunkName: 'health-advice' });
const HealthAnalysis = smartLazy(() => import('./pages/health/HealthAnalysis'), { ...defaultLazyConfig, chunkName: 'health-analysis' });
const HealthMonitoring = smartLazy(() => import('./pages/health/HealthMonitoring'), { ...defaultLazyConfig, chunkName: 'health-monitoring' });
const HealthRecords = smartLazy(() => import('./pages/health/HealthRecords'), { ...lowPriorityConfig, chunkName: 'health-records' });

// 安全相关页面
const SecurityCenter = smartLazy(() => import('./pages/SecurityCenter'), { ...lowPriorityConfig, chunkName: 'security' });

// AI功能相关页面
const AIMultimodal = smartLazy(() => import('./pages/ai/AIMultimodal'), { ...defaultLazyConfig, chunkName: 'ai-multimodal' });
const ImageGeneration = smartLazy(() => import('./pages/ai/ImageGeneration'), { ...defaultLazyConfig, chunkName: 'image-generation' });
const ImageAnalysis = smartLazy(() => import('./pages/ai/ImageAnalysis'), { ...defaultLazyConfig, chunkName: 'image-analysis' });
const AIAnalysis = smartLazy(() => import('./pages/ai/AIAnalysis'), { ...defaultLazyConfig, chunkName: 'ai-analysis' });
const AIGeneration = smartLazy(() => import('./pages/ai/AIGeneration'), { ...defaultLazyConfig, chunkName: 'ai-generation' });
const AGIReasoning = smartLazy(() => import('./pages/ai/AGIReasoning'), { ...lowPriorityConfig, chunkName: 'agi-reasoning' });
const AGIPlanning = smartLazy(() => import('./pages/ai/AGIPlanning'), { ...lowPriorityConfig, chunkName: 'agi-planning' });
const MetaLearning = smartLazy(() => import('./pages/ai/MetaLearning'), { ...lowPriorityConfig, chunkName: 'meta-learning' });
const SelfEvolution = smartLazy(() => import('./pages/ai/SelfEvolution'), { ...lowPriorityConfig, chunkName: 'self-evolution' });
const VideoProcessing = smartLazy(() => import('./pages/ai/VideoProcessing'), { ...lowPriorityConfig, chunkName: 'video-processing' });
const AudioProcessing = smartLazy(() => import('./pages/ai/AudioProcessing'), { ...lowPriorityConfig, chunkName: 'audio-processing' });

// 第三方集成相关页面
const ThirdPartyIntegration = smartLazy(() => import('./pages/integration/ThirdPartyIntegration'), { ...lowPriorityConfig, chunkName: 'third-party' });

// 接口文档相关页面
const APICatalog = smartLazy(() => import('./pages/api-docs/APICatalog'), { ...defaultLazyConfig, chunkName: 'api-catalog' });
const APIStatus = smartLazy(() => import('./pages/api-docs/APIStatus'), { ...defaultLazyConfig, chunkName: 'api-status' });
const APIVersions = smartLazy(() => import('./pages/api-docs/APIVersions'), { ...lowPriorityConfig, chunkName: 'api-versions' });
const APISearch = smartLazy(() => import('./pages/api-docs/APISearch'), { ...defaultLazyConfig, chunkName: 'api-search' });

// 管理员页面 - 按需加载
const WisdomManagement = smartLazy(() => import('./pages/admin/WisdomManagement'), { ...defaultLazyConfig, chunkName: 'wisdom-management' });
const WisdomEditor = smartLazy(() => import('./pages/admin/WisdomEditor'), { ...defaultLazyConfig, chunkName: 'wisdom-editor' });
const CategoryManagement = smartLazy(() => import('./pages/admin/CategoryManagement'), { ...lowPriorityConfig, chunkName: 'category-management' });
const TagManagement = smartLazy(() => import('./pages/admin/TagManagement'), { ...lowPriorityConfig, chunkName: 'tag-management' });
const UserManagement = smartLazy(() => import('./pages/admin/UserManagement'), { ...defaultLazyConfig, chunkName: 'user-management' });
const RoleManagement = smartLazy(() => import('./pages/admin/RoleManagement'), { ...lowPriorityConfig, chunkName: 'role-management' });
const PermissionManagement = smartLazy(() => import('./pages/admin/PermissionManagement'), { ...lowPriorityConfig, chunkName: 'permission-management' });
const MenuManagement = smartLazy(() => import('./pages/admin/MenuManagement'), { ...lowPriorityConfig, chunkName: 'menu-management' });
const IconSelectorPage = smartLazy(() => import('./pages/admin/IconSelector'), { ...lowPriorityConfig, chunkName: 'icon-selector' });
const DataAnalytics = smartLazy(() => import('./pages/admin/DataAnalytics'), { ...lowPriorityConfig, chunkName: 'data-analytics' });
const SystemSettings = smartLazy(() => import('./pages/admin/SystemSettings'), { ...lowPriorityConfig, chunkName: 'system-settings' });
  const NotificationCenter = smartLazy(() => import('./pages/admin/NotificationCenter'), { ...lowPriorityConfig, chunkName: 'notification-center' });
  const ContentReview = smartLazy(() => import('./pages/admin/ContentReview'), { ...lowPriorityConfig, chunkName: 'content-review' });
  const SystemMonitoring = smartLazy(() => import('./pages/admin/SystemMonitoring'), { ...lowPriorityConfig, chunkName: 'system-monitoring' });
const AdminDashboard = smartLazy(() => import('./pages/admin/AdminDashboard'), { ...defaultLazyConfig, chunkName: 'admin-dashboard' });
// 系统管理子模块页面（新增）
const DatabaseManagement = smartLazy(() => import('./pages/admin/DatabaseManagement'), { ...lowPriorityConfig, chunkName: 'admin-database' });
const LogsManagement = smartLazy(() => import('./pages/admin/LogsManagement'), { ...lowPriorityConfig, chunkName: 'admin-logs' });
const IssueTracking = smartLazy(() => import('./pages/admin/IssueTracking'), { ...lowPriorityConfig, chunkName: 'admin-issues' });

// 调试和测试页面
const APITest = smartLazy(() => import('./pages/debug/APITest'), { ...lowPriorityConfig, chunkName: 'api-test' });

// 加载组件
const LoadingFallback: React.FC = () => <PageLoading />;

// 权限保护组件
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();
  
  // 如果正在加载认证状态，显示加载页面
  if (isLoading) {
    return <PageLoading />;
  }
  
  // 如果未认证，跳转到登录页面
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

// AdminRoute 统一由 components/auth/RouteGuard 导出并实现

// 路由预加载组件
const RoutePreloadManager: React.FC = () => {
  const location = useLocation();
  const { preloadRelatedRoutes } = useSmartPreload();

  useEffect(() => {
    // 注册所有路由到预加载器
    const preloader = RoutePreloader.getInstance();
    
    // 注册主要路由
    preloader.registerRoute('/dashboard', () => import('./pages/Dashboard'));
    preloader.registerRoute('/chat', () => import('./pages/Chat'));
    preloader.registerRoute('/profile', () => import('./pages/Profile'));
    preloader.registerRoute('/help', () => import('./pages/Help'));
    
    // 注册智慧相关路由
    preloader.registerRoute('/wisdom', () => import('./pages/Wisdom'));
    preloader.registerRoute('/wisdom/:id', () => import('./pages/WisdomDetail'));
    preloader.registerRoute('/recommendation', () => import('./pages/RecommendationCenter'));
    
    // 注册学习相关路由
    preloader.registerRoute('/learning', () => import('./pages/IntelligentLearning'));
    preloader.registerRoute('/learning/analytics', () => import('./pages/LearningAnalyticsDashboard'));
    preloader.registerRoute('/learning/courses', () => import('./pages/CourseCenter'));
    
    // 注册管理员路由
    // 管理员仪表板预加载
    preloader.registerRoute('/admin/dashboard', () => import('./pages/admin/AdminDashboard'));
    preloader.registerRoute('/admin/wisdom', () => import('./pages/admin/WisdomManagement'));
    preloader.registerRoute('/admin/users', () => import('./pages/admin/UserManagement'));
    preloader.registerRoute('/admin/icon-selector', () => import('./pages/admin/IconSelector'));
    // 新增系统管理子路由预加载
    preloader.registerRoute('/admin/database', () => import('./pages/admin/DatabaseManagement'));
    preloader.registerRoute('/admin/logs', () => import('./pages/admin/LogsManagement'));
    preloader.registerRoute('/admin/issues', () => import('./pages/admin/IssueTracking'));
    
    // 初始化时预加载高优先级组件
    preloadComponents([
      { importFn: () => import('./pages/Dashboard'), priority: 'high', chunkName: 'dashboard' },
      { importFn: () => import('./pages/Chat'), priority: 'high', chunkName: 'chat' },
    ]);
  }, []);

  useEffect(() => {
    // 当路由变化时，预加载相关路由
    preloadRelatedRoutes(location.pathname);
  }, [location.pathname, preloadRelatedRoutes]);

  return null;
};

const AppContent: React.FC = () => {
  // 在AntdApp内部初始化notification服务
  useNotificationService();
  
  return (
    <Router>
      <RoutePreloadManager />
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/verify-email" element={<EmailVerification />} />
        
        {/* 管理员路由 - 使用MainLayout */}
        <Route path="/admin" element={
          <AdminRoute>
            <MainLayout />
          </AdminRoute>
        }>
          <Route index element={<Navigate to="dashboard" replace />} />
          <Route path="dashboard" element={
            <Suspense fallback={<LoadingFallback />}>
              <AdminDashboard />
            </Suspense>
          } />
          <Route path="wisdom" element={
            <Suspense fallback={<LoadingFallback />}>
              <WisdomManagement />
            </Suspense>
          } />
          <Route path="wisdom/add" element={
            <Suspense fallback={<LoadingFallback />}>
              <WisdomEditor />
            </Suspense>
          } />
          <Route path="wisdom/edit/:id" element={
            <Suspense fallback={<LoadingFallback />}>
              <WisdomEditor />
            </Suspense>
          } />
          <Route path="categories" element={
            <Suspense fallback={<LoadingFallback />}>
              <CategoryManagement />
            </Suspense>
          } />
          <Route path="tags" element={
            <Suspense fallback={<LoadingFallback />}>
              <TagManagement />
            </Suspense>
          } />
          <Route path="users" element={
            <Suspense fallback={<LoadingFallback />}>
              <UserManagement />
            </Suspense>
          } />
          <Route path="roles" element={
            <Suspense fallback={<LoadingFallback />}>
              <RoleManagement />
            </Suspense>
          } />
          <Route path="permissions" element={
            <Suspense fallback={<LoadingFallback />}>
              <PermissionManagement />
            </Suspense>
          } />
          <Route path="menus" element={
            <Suspense fallback={<LoadingFallback />}>
              <MenuManagement />
            </Suspense>
          } />
          <Route path="icon-selector" element={
            <Suspense fallback={<LoadingFallback />}>
              <IconSelectorPage />
            </Suspense>
          } />

          <Route path="analytics" element={
            <Suspense fallback={<LoadingFallback />}>
              <DataAnalytics />
            </Suspense>
          } />
          <Route path="settings" element={
            <Suspense fallback={<LoadingFallback />}>
              <SystemSettings />
            </Suspense>
          } />
          <Route path="notifications" element={
            <Suspense fallback={<LoadingFallback />}>
              <NotificationCenter />
            </Suspense>
          } />
          <Route path="review" element={
            <Suspense fallback={<LoadingFallback />}>
              <ContentReview />
            </Suspense>
          } />
          <Route path="monitoring" element={
            <Suspense fallback={<LoadingFallback />}>
              <SystemMonitoring />
            </Suspense>
          } />
          <Route path="database" element={
            <Suspense fallback={<LoadingFallback />}> 
              <DatabaseManagement />
            </Suspense>
          } />
          <Route path="logs" element={
            <Suspense fallback={<LoadingFallback />}> 
              <LogsManagement />
            </Suspense>
          } />
          <Route path="issues" element={
            <Suspense fallback={<LoadingFallback />}> 
              <IssueTracking />
            </Suspense>
          } />

          <Route path="api-test" element={
            <Suspense fallback={<LoadingFallback />}>
              <APITest />
            </Suspense>
          } />
        </Route>
        
        {/* 普通用户路由 - 使用MainLayout */}
        <Route path="/" element={
          <ProtectedRoute>
            <MainLayout />
          </ProtectedRoute>
        }>
          <Route index element={
            <Suspense fallback={<LoadingFallback />}>
              <Dashboard />
            </Suspense>
          } />
          <Route path="dashboard" element={
            <Suspense fallback={<LoadingFallback />}>
              <Dashboard />
            </Suspense>
          } />

          {/* 已移除 /home 路由，统一使用 /dashboard 作为入口 */}
          <Route path="chat" element={
            <Suspense fallback={<LoadingFallback />}>
              <Chat />
            </Suspense>
          } />
          <Route path="wisdom" element={
            <Suspense fallback={<LoadingFallback />}>
              <Wisdom />
            </Suspense>
          } />
          <Route path="wisdom/search" element={
            <Suspense fallback={<LoadingFallback />}>
              <Wisdom />
            </Suspense>
          } />
          <Route path="wisdom/categories" element={
            <Suspense fallback={<LoadingFallback />}>
              <Wisdom />
            </Suspense>
          } />
          <Route path="wisdom/:id" element={
            <Suspense fallback={<LoadingFallback />}>
              <WisdomDetail />
            </Suspense>
          } />
          <Route path="wisdom/detail" element={
            <Suspense fallback={<LoadingFallback />}>
              <WisdomDetail />
            </Suspense>
          } />
          <Route path="recommendation" element={
            <Suspense fallback={<LoadingFallback />}>
              <RecommendationCenter />
            </Suspense>
          } />
          <Route path="recommendations" element={
            <Suspense fallback={<LoadingFallback />}>
              <RecommendationCenter />
            </Suspense>
          } />
          <Route path="community" element={
            <Suspense fallback={<LoadingFallback />}>
              <Community />
            </Suspense>
          } />
          <Route path="community/chat" element={
            <Suspense fallback={<LoadingFallback />}>
              <Community />
            </Suspense>
          } />
          <Route path="community/groups" element={
            <Suspense fallback={<LoadingFallback />}>
              <Community />
            </Suspense>
          } />
          <Route path="community/events" element={
            <Suspense fallback={<LoadingFallback />}>
              <Community />
            </Suspense>
          } />
          <Route path="intelligent-learning" element={
            <Suspense fallback={<LoadingFallback />}>
              <IntelligentLearning />
            </Suspense>
          } />
          <Route path="learning/analytics-dashboard" element={
            <Suspense fallback={<LoadingFallback />}>
              <LearningAnalyticsDashboard />
            </Suspense>
          } />
          <Route path="course-center" element={
            <Suspense fallback={<LoadingFallback />}>
              <CourseCenter />
            </Suspense>
          } />
          <Route path="ability-assessment" element={
            <Suspense fallback={<LoadingFallback />}>
              <AbilityAssessment />
            </Suspense>
          } />
          <Route path="learning-plan" element={
            <Suspense fallback={<LoadingFallback />}>
              <LearningPlan />
            </Suspense>
          } />
          <Route path="daily-checkin" element={
            <Suspense fallback={<LoadingFallback />}>
              <DailyCheckin />
            </Suspense>
          } />
          <Route path="achievement-center" element={
            <Suspense fallback={<LoadingFallback />}>
              <AchievementCenter />
            </Suspense>
          } />
          <Route path="projects" element={
            <Suspense fallback={<LoadingFallback />}>
              <ProjectManagement />
            </Suspense>
          } />
          <Route path="projects/workspace" element={
            <Suspense fallback={<LoadingFallback />}>
              <ProjectWorkspace />
            </Suspense>
          } />
          <Route path="projects/tasks" element={
            <Suspense fallback={<LoadingFallback />}>
              <TaskManagement />
            </Suspense>
          } />
          <Route path="projects/collaboration" element={
            <Suspense fallback={<LoadingFallback />}>
              <TeamCollaboration />
            </Suspense>
          } />
          <Route path="projects/analytics" element={
            <Suspense fallback={<LoadingFallback />}>
              <ProjectAnalytics />
            </Suspense>
          } />
          <Route path="health" element={
            <Suspense fallback={<LoadingFallback />}>
              <HealthManagement />
            </Suspense>
          } />
          <Route path="health/advice" element={
            <Suspense fallback={<LoadingFallback />}>
              <HealthAdvice />
            </Suspense>
          } />
          <Route path="health/analysis" element={
            <Suspense fallback={<LoadingFallback />}>
              <HealthAnalysis />
            </Suspense>
          } />
          <Route path="health/monitoring" element={
            <Suspense fallback={<LoadingFallback />}>
              <HealthMonitoring />
            </Suspense>
          } />
          <Route path="health/monitor" element={
            <Suspense fallback={<LoadingFallback />}>
              <HealthMonitoring />
            </Suspense>
          } />
          <Route path="health/records" element={
            <Suspense fallback={<LoadingFallback />}>
              <HealthRecords />
            </Suspense>
          } />
          <Route path="learning/progress" element={
            <Suspense fallback={<LoadingFallback />}>
              <LearningProgress />
            </Suspense>
          } />
          <Route path="learning/courses" element={
            <Suspense fallback={<LoadingFallback />}>
              <LearningCourses />
            </Suspense>
          } />
          <Route path="learning/assessment" element={
            <Suspense fallback={<LoadingFallback />}>
              <AbilityAssessment />
            </Suspense>
          } />
          <Route path="learning/certificates" element={
            <Suspense fallback={<LoadingFallback />}>
              <AchievementCenter />
            </Suspense>
          } />
          <Route path="security" element={
            <Suspense fallback={<LoadingFallback />}>
              <SecurityCenter />
            </Suspense>
          } />
          


          {/* AI功能路由 */}
          <Route path="ai" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIMultimodal />
            </Suspense>
          } />
          <Route path="ai/multimodal" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIMultimodal />
            </Suspense>
          } />
          <Route path="ai/multimodal/image-generation" element={
            <Suspense fallback={<LoadingFallback />}>
              <ImageGeneration />
            </Suspense>
          } />
          <Route path="ai/multimodal/image-analysis" element={
            <Suspense fallback={<LoadingFallback />}>
              <ImageAnalysis />
            </Suspense>
          } />
          <Route path="ai/multimodal/video-processing" element={
            <Suspense fallback={<LoadingFallback />}>
              <VideoProcessing />
            </Suspense>
          } />
          <Route path="ai/multimodal/audio-processing" element={
            <Suspense fallback={<LoadingFallback />}>
              <AudioProcessing />
            </Suspense>
          } />
          <Route path="ai/analysis" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIAnalysis />
            </Suspense>
          } />
          <Route path="ai/analysis/data" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIAnalysis />
            </Suspense>
          } />
          <Route path="ai/analysis/trends" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIAnalysis />
            </Suspense>
          } />
          <Route path="ai/analysis/reports" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIAnalysis />
            </Suspense>
          } />
          <Route path="ai/generation" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIGeneration />
            </Suspense>
          } />
          <Route path="ai/generation/text" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIGeneration />
            </Suspense>
          } />
          <Route path="ai/generation/code" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIGeneration />
            </Suspense>
          } />
          <Route path="ai/generation/design" element={
            <Suspense fallback={<LoadingFallback />}>
              <AIGeneration />
            </Suspense>
          } />
          <Route path="ai/image-analysis" element={
            <Suspense fallback={<LoadingFallback />}>
              <ImageAnalysis />
            </Suspense>
          } />
          <Route path="ai/image-generation" element={
            <Suspense fallback={<LoadingFallback />}>
              <ImageGeneration />
            </Suspense>
          } />
          <Route path="ai/agi-reasoning" element={
            <Suspense fallback={<LoadingFallback />}>
              <AGIReasoning />
            </Suspense>
          } />
          <Route path="ai/agi-planning" element={
            <Suspense fallback={<LoadingFallback />}>
              <AGIPlanning />
            </Suspense>
          } />
          <Route path="ai/meta-learning" element={
            <Suspense fallback={<LoadingFallback />}>
              <MetaLearning />
            </Suspense>
          } />
          <Route path="ai/self-evolution" element={
            <Suspense fallback={<LoadingFallback />}>
              <SelfEvolution />
            </Suspense>
          } />
          {/* 第三方集成路由 */}
          <Route path="integration" element={
            <Suspense fallback={<LoadingFallback />}>
              <ThirdPartyIntegration />
            </Suspense>
          } />
          
          {/* 接口文档路由 */}
          <Route path="api-docs" element={<Navigate to="api-docs/catalog" replace />} />
          <Route path="api-docs/catalog" element={
            <Suspense fallback={<LoadingFallback />}>
              <APICatalog />
            </Suspense>
          } />
          <Route path="api-docs/status" element={
            <Suspense fallback={<LoadingFallback />}>
              <APIStatus />
            </Suspense>
          } />
          <Route path="api-docs/versions" element={
            <Suspense fallback={<LoadingFallback />}>
              <APIVersions />
            </Suspense>
          } />
          <Route path="api-docs/search" element={
            <Suspense fallback={<LoadingFallback />}>
              <APISearch />
            </Suspense>
          } />
          
          <Route path="profile" element={
            <Suspense fallback={<LoadingFallback />}>
              <Profile />
            </Suspense>
          } />
          <Route path="profile/settings" element={
            <Suspense fallback={<LoadingFallback />}>
              <Profile />
            </Suspense>
          } />
          <Route path="profile/security" element={
            <Suspense fallback={<LoadingFallback />}>
              <Profile />
            </Suspense>
          } />
          <Route path="profile/notifications" element={
            <Suspense fallback={<LoadingFallback />}>
              <Profile />
            </Suspense>
          } />
          <Route path="favorites" element={
            <Suspense fallback={<LoadingFallback />}>
              <UserFavorites />
            </Suspense>
          } />
          <Route path="notes" element={
            <Suspense fallback={<LoadingFallback />}>
              <UserNotes />
            </Suspense>
          } />
          <Route path="help" element={
            <Suspense fallback={<LoadingFallback />}>
              <Help />
            </Suspense>
          } />
        </Route>
      </Routes>
    </Router>
  );
};

const App: React.FC = () => {
  return (
    <GlobalErrorBoundary>
      <ConfigProvider 
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: '#1890ff',
          },
        }}
      >
        <AntdApp>
          <MenuProvider>
            <AppContent />
          </MenuProvider>
        </AntdApp>
      </ConfigProvider>
    </GlobalErrorBoundary>
  );
};

export default App;
