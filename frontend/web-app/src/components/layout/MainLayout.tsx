import React, { useState, useEffect } from 'react';
import { Layout, FloatButton, ConfigProvider, theme as antdTheme, Breadcrumb } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { 
  CustomerServiceOutlined, 
  QuestionCircleOutlined,
  UpOutlined,
  HomeOutlined 
} from '@ant-design/icons';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { Footer } from './Footer';
import { useTheme, useBreakpoint, useApp } from '../../hooks';
import Loading from '../common/Loading';
import CustomerService from '../common/CustomerService';
import { useTranslation } from 'react-i18next';


const { Content, Sider } = Layout;

export const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileDrawerVisible, setMobileDrawerVisible] = useState(false);
  const [customerServiceVisible, setCustomerServiceVisible] = useState(false);
  const { theme } = useTheme();
  const { isMobile, isTablet } = useBreakpoint();
  const { loading, setLoading } = useApp();
  const location = useLocation();
  const navigate = useNavigate();
  const { t } = useTranslation();

  // MainLayout不再管理全局loading状态

  // 面包屑路径映射（使用 i18n 动态翻译，包含回退文案）
  const breadcrumbNameMap: Record<string, string> = {
    '/': t('navigation.home', { defaultValue: '首页' }),
    '/dashboard': t('mainMenu.labels.dashboard', { defaultValue: '仪表板' }),
    '/dashboard/analysis': t('dashboard.analysis', { defaultValue: '分析页' }),
    '/dashboard/monitor': t('dashboard.monitor', { defaultValue: '监控页' }),
    '/dashboard/workspace': t('dashboard.workspace', { defaultValue: '工作台' }),
    '/admin': t('mainMenu.labels.system-management', { defaultValue: '系统管理' }),
    '/admin/dashboard': t('mainMenu.labels.admin-dashboard', { defaultValue: '管理员仪表板' }),
    '/admin/users': t('mainMenu.labels.system-users', { defaultValue: '用户管理' }),
    '/admin/settings': t('mainMenu.labels.system-settings', { defaultValue: '系统设置' }),
    '/profile': t('mainMenu.labels.profile-overview', { defaultValue: '个人中心' }),
    '/settings': t('header.settings', { defaultValue: '设置' }),
    '/help': t('header.help', { defaultValue: '帮助中心' }),
    '/ai': t('mainMenu.labels.ai-services', { defaultValue: 'AI 服务' }),
    '/ai/analysis': t('mainMenu.labels.ai-analysis', { defaultValue: 'AI 分析' }),
    '/ai/generation': t('mainMenu.labels.ai-generation', { defaultValue: 'AI 生成' }),
    '/learning': t('mainMenu.labels.intelligent-learning', { defaultValue: '学习中心' }),
    '/learning/courses': t('mainMenu.labels.learning-courses', { defaultValue: '课程中心' }),
    '/learning/progress': t('mainMenu.labels.learning-progress', { defaultValue: '学习进度' }),
    '/projects': t('mainMenu.labels.project-management', { defaultValue: '项目管理' }),
    '/projects/workspace': t('mainMenu.labels.project-workspace', { defaultValue: '项目工作台' }),
    '/projects/analytics': t('mainMenu.labels.project-analytics', { defaultValue: '项目分析' }),
    '/health': t('mainMenu.labels.health-management', { defaultValue: '健康管理' }),
    '/community': t('mainMenu.labels.community', { defaultValue: '社区' }),
    '/wisdom': t('mainMenu.labels.wisdom', { defaultValue: '智慧库' }),
    '/api-docs': t('mainMenu.labels.api-documentation', { defaultValue: '接口文档' }),
    '/api-docs/catalog': t('mainMenu.labels.api-catalog', { defaultValue: '接口目录' }),
    '/api-docs/status': t('mainMenu.labels.api-status', { defaultValue: '接口状态' }),
    '/api-docs/versions': t('mainMenu.labels.api-versions', { defaultValue: '版本管理' }),
    '/api-docs/search': t('mainMenu.labels.api-search', { defaultValue: '快速检索' }),
  };

  // 生成面包屑项目
  const getBreadcrumbItems = () => {
    const pathSnippets = location.pathname.split('/').filter(i => i);
    const breadcrumbItems = [
      {
        title: (
          <span onClick={() => navigate('/')} className="cursor-pointer hover:text-primary">
            <HomeOutlined className="mr-1" />
            {t('navigation.home', { defaultValue: '首页' })}
          </span>
        ),
      },
    ];

    pathSnippets.forEach((_, index) => {
      const url = `/${pathSnippets.slice(0, index + 1).join('/')}`;
      const breadcrumbName = breadcrumbNameMap[url];
      if (breadcrumbName) {
        breadcrumbItems.push({
          title: index === pathSnippets.length - 1 ? (
            <span className="text-gray-500">{breadcrumbName}</span>
          ) : (
            <span 
              onClick={() => navigate(url)} 
              className="cursor-pointer hover:text-primary"
            >
              {breadcrumbName}
            </span>
          ),
        });
      }
    });

    return breadcrumbItems;
  };

  // 移动端自动折叠侧边栏
  useEffect(() => {
    if (isMobile) {
      setCollapsed(true);
    } else if (isTablet) {
      setCollapsed(false);
    }
  }, [isMobile, isTablet]);

  // 处理侧边栏折叠
  const handleCollapse = (collapsed: boolean) => {
    if (isMobile) {
      setMobileDrawerVisible(!mobileDrawerVisible);
    } else {
      setCollapsed(collapsed);
    }
  };

  // 移动端点击遮罩关闭抽屉
  const handleMobileDrawerClose = () => {
    setMobileDrawerVisible(false);
  };

  // Ant Design Pro 主题配置
  const antdThemeConfig = {
    algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: {
      colorPrimary: '#1890ff', // Pro 标准蓝色主题
      colorSuccess: '#52c41a',
      colorWarning: '#faad14',
      colorError: '#ff4d4f',
      colorInfo: '#1890ff',
      borderRadius: 6,
      borderRadiusLG: 8,
      borderRadiusSM: 4,
      wireframe: false,
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif',
      fontSize: 14,
      controlHeight: 32,
      controlHeightLG: 40,
      controlHeightSM: 24,
      boxShadow: theme === 'dark' 
        ? '0 2px 8px rgba(0, 0, 0, 0.15)'
        : '0 2px 8px rgba(0, 0, 0, 0.15)',
      motionDurationMid: '0.2s',
      // Pro 风格的间距和尺寸
      marginXS: 8,
      marginSM: 12,
      margin: 16,
      marginMD: 20,
      marginLG: 24,
      marginXL: 32,
      paddingXS: 8,
      paddingSM: 12,
      padding: 16,
      paddingMD: 20,
      paddingLG: 24,
      paddingXL: 32,
    },
    components: {
      Layout: {
        headerBg: theme === 'dark' ? '#001529' : '#ffffff',
        headerHeight: 48,
        headerPadding: '0 24px',
        siderBg: theme === 'dark' ? '#001529' : '#ffffff',
        bodyBg: theme === 'dark' ? '#000000' : '#f0f2f5',
        footerBg: theme === 'dark' ? '#001529' : '#ffffff',
        triggerBg: theme === 'dark' ? '#002140' : '#ffffff',
        triggerColor: theme === 'dark' ? '#ffffff' : 'rgba(0, 0, 0, 0.65)',
      },
      Menu: {
        itemBg: 'transparent',
        subMenuItemBg: 'transparent',
        itemSelectedBg: theme === 'dark' 
          ? '#1890ff' 
          : '#e6f7ff',
        itemHoverBg: theme === 'dark' 
          ? 'rgba(255, 255, 255, 0.08)' 
          : 'rgba(24, 144, 255, 0.06)',
        itemSelectedColor: theme === 'dark' ? '#ffffff' : '#1890ff',
        itemColor: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
        iconSize: 14,
        itemHeight: 40,
        itemMarginInline: 4,
        itemBorderRadius: 6,
        subMenuItemBorderRadius: 6,
      },
      Button: {
        borderRadius: 6,
        controlHeight: 32,
        paddingContentHorizontal: 15,
        fontWeight: 400,
      },
      Card: {
        borderRadius: 6,
        paddingLG: 24,
        boxShadowTertiary: theme === 'dark'
          ? '0 1px 2px 0 rgba(0, 0, 0, 0.03)'
          : '0 1px 2px 0 rgba(0, 0, 0, 0.03)',
      },
      Input: {
        borderRadius: 6,
        controlHeight: 32,
        paddingInline: 11,
      },
      Select: {
        borderRadius: 6,
        controlHeight: 32,
      },
      Dropdown: {
        borderRadius: 6,
        boxShadowSecondary: theme === 'dark'
          ? '0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05)'
          : '0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
      },
      Modal: {
        borderRadius: 8,
        paddingLG: 24,
      },
      Drawer: {
        borderRadius: 0,
        paddingLG: 24,
      },
      Tooltip: {
        borderRadius: 6,
      },
      Popover: {
        borderRadius: 6,
      },
      Breadcrumb: {
        fontSize: 14,
        itemColor: theme === 'dark' ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.45)',
        lastItemColor: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
        linkColor: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
        linkHoverColor: theme === 'dark' ? '#ffffff' : '#1890ff',
        separatorColor: theme === 'dark' ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.45)',
      },
    },
  };

  return (
    <ConfigProvider theme={antdThemeConfig}>
      <Layout style={{ minHeight: '100vh' }}>
        {/* 桌面端侧边栏 */}
        {!isMobile && (
          <Sider
            trigger={null}
            collapsible
            collapsed={collapsed}
            width={256}
            collapsedWidth={64}
            style={{
              position: 'fixed',
              left: 0,
              top: 0,
              bottom: 0,
              zIndex: 50,
              transition: 'all 0.3s',
              background: theme === 'dark' 
                ? 'linear-gradient(to bottom, #111827, #1f2937, #111827)' 
                : 'linear-gradient(to bottom, #ffffff, #f9fafb, #ffffff)',
              borderRight: theme === 'dark' ? 'none' : '1px solid rgba(229, 231, 235, 0.5)',
              boxShadow: theme === 'dark' 
                ? '4px 0 24px rgba(0, 0, 0, 0.4), 2px 0 8px rgba(0, 0, 0, 0.3)' 
                : '4px 0 24px rgba(0, 0, 0, 0.08), 2px 0 8px rgba(0, 0, 0, 0.06)',
              backdropFilter: 'blur(10px)',
            }}
          >
            <Sidebar 
              collapsed={collapsed}
              isMobile={false}
              onClose={() => {}}
            />
          </Sider>
        )}

        {/* 移动端抽屉式侧边栏 */}
        {isMobile && (
          <>
            {/* 遮罩层 */}
            {mobileDrawerVisible && (
              <div
                style={{
                  position: 'fixed',
                  top: 0,
                  left: 0,
                  right: 0,
                  bottom: 0,
                  backgroundColor: 'rgba(0, 0, 0, 0.5)',
                  zIndex: 40,
                  transition: 'opacity 0.3s',
                }}
                onClick={handleMobileDrawerClose}
              />
            )}
            
            {/* 抽屉 */}
            <div
              style={{
                position: 'fixed',
                left: 0,
                top: 0,
                bottom: 0,
                width: 256,
                zIndex: 50,
                transition: 'transform 0.3s',
                transform: mobileDrawerVisible ? 'translateX(0)' : 'translateX(-100%)',
                background: theme === 'dark' ? '#111827' : '#ffffff',
                boxShadow: mobileDrawerVisible 
                  ? (theme === 'dark' ? '4px 0 12px rgba(0, 0, 0, 0.4)' : '4px 0 12px rgba(0, 0, 0, 0.15)')
                  : 'none'
              }}
            >
              <Sidebar 
                collapsed={false}
                isMobile={true}
                onClose={handleMobileDrawerClose}
              />
            </div>
          </>
        )}

        {/* 主布局容器 */}
        <Layout style={{ 
          marginLeft: !isMobile ? (collapsed ? 64 : 256) : 0,
          transition: 'margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
        }}>
          {/* 头部 */}
          <Header 
            collapsed={collapsed} 
            onCollapse={handleCollapse}
          />

          {/* 面包屑导航 */}
          <div style={{
            padding: '12px 24px',
            background: theme === 'dark' ? '#141414' : '#fafafa',
            borderBottom: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
          }}>
            <Breadcrumb items={getBreadcrumbItems()} />
          </div>

          {/* 主要内容区域 */}
          <Content style={{ 
            margin: '24px 24px 0',
            padding: 0,
            background: theme === 'dark' ? '#141414' : '#ffffff',
            minHeight: 'calc(100vh - 157px)', // 48px header + 49px breadcrumb + 24px margin + 36px footer
            borderRadius: '6px',
            border: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
          }}>
            <Outlet />
          </Content>

          {/* 底部栏 */}
          <Footer />
        </Layout>

        {/* 回到顶部按钮 */}
        <FloatButton.BackTop
          style={{
            right: 24,
            bottom: 80,
          }}
        />

        {/* 浮动按钮组 */}
        <FloatButton.Group
          trigger="hover"
          type="primary"
          style={{ right: 24, bottom: 24 }}
          icon={<QuestionCircleOutlined />}
        >
          <FloatButton
            icon={<QuestionCircleOutlined />}
            tooltip="帮助中心"
            onClick={() => navigate('/help')}
          />
          <FloatButton
            icon={<CustomerServiceOutlined />}
            tooltip="客服支持"
            onClick={() => setCustomerServiceVisible(true)}
          />
        </FloatButton.Group>

        {/* 客服支持 */}
        <CustomerService
          visible={customerServiceVisible}
          onClose={() => setCustomerServiceVisible(false)}
        />


      </Layout>

      </ConfigProvider>
    );
};

export default MainLayout;