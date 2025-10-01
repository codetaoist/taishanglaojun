import React, { useState, useEffect } from 'react';
import { Layout, Menu, Avatar, Dropdown, Badge, Button, Drawer, Switch } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  BankOutlined,
  BookOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  BellOutlined,
  MoonOutlined,
  SunOutlined,
  GlobalOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { logoutUser } from '../../store/slices/authSlice';
import { setTheme, setLanguage, setSidebarCollapsed } from '../../store/slices/uiSlice';
import { setPanelVisible } from '../../store/slices/notificationSlice';

const { Header, Sider, Content } = Layout;

interface MainLayoutProps {
  children?: React.ReactNode;
}

// 菜单项配置
const menuItems = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: '仪表板',
    labelEn: 'Dashboard',
  },
  {
    key: '/consciousness',
    icon: <BankOutlined />,
    label: '意识融合',
    labelEn: 'Consciousness',
  },
  {
    key: '/cultural',
    icon: <BookOutlined />,
    label: '文化智慧',
    labelEn: 'Cultural Wisdom',
  },
  {
    key: '/profile',
    icon: <UserOutlined />,
    label: '个人资料',
    labelEn: 'Profile',
  },
];

// 管理员菜单项
const adminMenuItems = [
  {
    key: '/admin',
    icon: <SettingOutlined />,
    label: '系统管理',
    labelEn: 'Admin',
  },
];

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useAppDispatch();
  
  const { user } = useAppSelector(state => state.auth);
  const { theme, language } = useAppSelector(state => state.ui);
  const [sidebarCollapsed, setSidebarCollapsedLocal] = useState(false);
  const { stats } = useAppSelector(state => state.notification);
  
  const [mobileDrawerVisible, setMobileDrawerVisible] = useState(false);
  const [isMobile, setIsMobile] = useState(false);

  // 检测移动设备
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // 处理菜单点击
  const handleMenuClick = (key: string) => {
    navigate(key);
    if (isMobile) {
      setMobileDrawerVisible(false);
    }
  };

  // 处理用户菜单点击
  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate('/profile');
        break;
      case 'settings':
        // 打开设置面板
        break;
      case 'logout':
        dispatch(logoutUser());
        break;
    }
  };

  // 切换主题
  const toggleTheme = () => {
    dispatch(setTheme(theme === 'light' ? 'dark' : 'light'));
  };

  // 切换语言
  const toggleLanguage = () => {
    dispatch(setLanguage(language === 'zh-CN' ? 'en-US' : 'zh-CN'));
  };

  // 切换侧边栏
  const toggleSidebar = () => {
    if (isMobile) {
      setMobileDrawerVisible(!mobileDrawerVisible);
    } else {
      setSidebarCollapsedLocal(!sidebarCollapsed);
    }
  };

  // 打开通知面板
  const openNotificationPanel = () => {
    dispatch(setPanelVisible(true));
  };

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    const pathname = location.pathname;
    if (pathname.startsWith('/admin')) return ['/admin'];
    if (pathname.startsWith('/consciousness')) return ['/consciousness'];
    if (pathname.startsWith('/cultural')) return ['/cultural'];
    if (pathname.startsWith('/profile')) return ['/profile'];
    return ['/dashboard'];
  };

  // 获取菜单标签
  const getMenuLabel = (item: typeof menuItems[0]) => {
    return language === 'en-US' ? item.labelEn : item.label;
  };

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: language === 'en-US' ? 'Profile' : '个人资料',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: language === 'en-US' ? 'Settings' : '设置',
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: language === 'en-US' ? 'Logout' : '退出登录',
      danger: true,
    },
  ];

  // 渲染菜单
  const renderMenu = () => {
    const allMenuItems = [...menuItems];
    
    // 如果是管理员，添加管理员菜单
    if (user?.role === 'admin') {
      allMenuItems.push(...adminMenuItems);
    }

    return (
      <Menu
        theme={theme === 'dark' ? 'dark' : 'light'}
        mode="inline"
        selectedKeys={getSelectedKeys()}
        items={allMenuItems.map(item => ({
          key: item.key,
          icon: item.icon,
          label: getMenuLabel(item),
          onClick: () => handleMenuClick(item.key),
        }))}
        style={{ borderRight: 0 }}
      />
    );
  };

  // 侧边栏内容
  const sidebarContent = (
    <div className="sidebar-content">
      {/* Logo */}
      <div className="sidebar-logo">
        <div className="logo-icon">🧙‍♂️</div>
        {!sidebarCollapsed && (
          <div className="logo-text">
            <div className="logo-title">
              {language === 'en-US' ? 'Taishang' : '太上老君'}
            </div>
            <div className="logo-subtitle">
              {language === 'en-US' ? 'Sequence Zero' : '序列零'}
            </div>
          </div>
        )}
      </div>
      
      {/* 菜单 */}
      {renderMenu()}
    </div>
  );

  return (
    <Layout className="main-layout">
      {/* 桌面端侧边栏 */}
      {!isMobile && (
        <Sider
          trigger={null}
          collapsible
          collapsed={sidebarCollapsed}
          width={240}
          collapsedWidth={80}
          theme={theme === 'dark' ? 'dark' : 'light'}
          className="main-sider"
        >
          {sidebarContent}
        </Sider>
      )}

      {/* 移动端抽屉 */}
      {isMobile && (
        <Drawer
          title={null}
          placement="left"
          closable={false}
          onClose={() => setMobileDrawerVisible(false)}
          open={mobileDrawerVisible}
          bodyStyle={{ padding: 0 }}
          width={240}
        >
          {sidebarContent}
        </Drawer>
      )}

      <Layout>
        {/* 头部 */}
        <Header className="main-header">
          <div className="header-left">
            <Button
              type="text"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={toggleSidebar}
              className="sidebar-trigger"
            />
          </div>

          <div className="header-right">
            {/* 语言切换 */}
            <Button
              type="text"
              icon={<GlobalOutlined />}
              onClick={toggleLanguage}
              title={language === 'en-US' ? '切换到中文' : 'Switch to English'}
            />

            {/* 主题切换 */}
            <Button
              type="text"
              icon={theme === 'light' ? <MoonOutlined /> : <SunOutlined />}
              onClick={toggleTheme}
              title={theme === 'light' ? (language === 'en-US' ? 'Dark Mode' : '深色模式') : (language === 'en-US' ? 'Light Mode' : '浅色模式')}
            />

            {/* 通知 */}
            <Badge count={stats.unread} size="small">
              <Button
                type="text"
                icon={<BellOutlined />}
                onClick={openNotificationPanel}
                title={language === 'en-US' ? 'Notifications' : '通知'}
              />
            </Badge>

            {/* 用户菜单 */}
            <Dropdown
              menu={{
                items: userMenuItems,
                onClick: handleUserMenuClick,
              }}
              placement="bottomRight"
              arrow
            >
              <div className="user-info">
                <Avatar
                  size="small"
                  src={user?.avatar}
                  icon={<UserOutlined />}
                />
                <span className="username">{user?.username || user?.email}</span>
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* 主内容区 */}
        <Content className="main-content">
          {children}
        </Content>
      </Layout>

      <style>{`
        .main-layout {
          height: 100vh;
        }

        .main-sider {
          box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
        }

        .sidebar-content {
          height: 100%;
          display: flex;
          flex-direction: column;
        }

        .sidebar-logo {
          display: flex;
          align-items: center;
          padding: 16px;
          border-bottom: 1px solid var(--border-light);
        }

        .logo-icon {
          font-size: 24px;
          margin-right: 12px;
        }

        .logo-text {
          flex: 1;
        }

        .logo-title {
          font-size: 16px;
          font-weight: 600;
          color: var(--text-primary);
          line-height: 1.2;
        }

        .logo-subtitle {
          font-size: 12px;
          color: var(--text-secondary);
          line-height: 1.2;
        }

        .main-header {
          background: var(--bg-primary);
          border-bottom: 1px solid var(--border-light);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0 24px;
          height: 64px;
        }

        .header-left {
          display: flex;
          align-items: center;
        }

        .header-right {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .sidebar-trigger {
          font-size: 18px;
          width: 40px;
          height: 40px;
        }

        .user-info {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          border-radius: 6px;
          cursor: pointer;
          transition: background-color 0.3s;
        }

        .user-info:hover {
          background-color: var(--bg-secondary);
        }

        .username {
          font-size: 14px;
          color: var(--text-primary);
          max-width: 120px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .main-content {
          background: var(--bg-secondary);
          overflow-y: auto;
        }

        @media (max-width: 768px) {
          .main-header {
            padding: 0 16px;
          }

          .username {
            display: none;
          }
        }
      `}</style>
    </Layout>
  );
};

export default MainLayout;