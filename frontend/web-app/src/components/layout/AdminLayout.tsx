import React, { useState } from 'react';
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Typography,
  Badge,
  Button,
  Drawer,
  Grid
} from 'antd';
import {
  DashboardOutlined,
  BookOutlined,
  TagsOutlined,
  FolderOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  BellOutlined,
  MenuOutlined,
  HomeOutlined,
  BarChartOutlined,
  TeamOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

const { Header, Sider, Content } = Layout;
const { Title, Text } = Typography;
const { useBreakpoint } = Grid;

interface AdminLayoutProps {
  children?: React.ReactNode;
}

const AdminLayout: React.FC<AdminLayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const screens = useBreakpoint();
  
  const [collapsed, setCollapsed] = useState(false);
  const [mobileDrawerVisible, setMobileDrawerVisible] = useState(false);
  const isMobile = !screens.lg;

  // 菜单项配置
  const menuItems = [
    {
      key: '/admin',
      icon: <DashboardOutlined />,
      label: '仪表板',
      path: '/admin'
    },
    {
      key: '/admin/wisdom',
      icon: <BookOutlined />,
      label: '智慧内容',
      path: '/admin/wisdom',
      children: [
        {
          key: '/admin/wisdom/list',
          label: '内容列表',
          path: '/admin/wisdom'
        },
        {
          key: '/admin/wisdom/add',
          label: '添加内容',
          path: '/admin/wisdom/add'
        }
      ]
    },
    {
      key: '/admin/categories',
      icon: <FolderOutlined />,
      label: '分类管理',
      path: '/admin/categories'
    },
    {
      key: '/admin/tags',
      icon: <TagsOutlined />,
      label: '标签管理',
      path: '/admin/tags'
    },
    {
      key: '/admin/analytics',
      icon: <BarChartOutlined />,
      label: '数据分析',
      path: '/admin/analytics'
    },
    {
      key: '/admin/users',
      icon: <TeamOutlined />,
      label: '用户管理',
      path: '/admin/users'
    },
    {
      key: '/admin/settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      path: '/admin/settings'
    }
  ];

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料'
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '账户设置'
    },
    {
      type: 'divider'
    },
    {
      key: 'home',
      icon: <HomeOutlined />,
      label: '返回首页'
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true
    }
  ];

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    const item = findMenuItem(menuItems, key);
    if (item?.path) {
      navigate(item.path);
      if (isMobile) {
        setMobileDrawerVisible(false);
      }
    }
  };

  // 查找菜单项
  interface MenuItem {
    key: string;
    icon?: React.ReactNode;
    label: string;
    path: string;
    children?: MenuItem[];
  }
  
  const findMenuItem = (items: MenuItem[], key: string): MenuItem | null => {
    for (const item of items) {
      if (item.key === key) return item;
      if (item.children) {
        const found = findMenuItem(item.children, key);
        if (found) return found;
      }
    }
    return null;
  };

  // 处理用户菜单点击
  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate('/admin/profile');
        break;
      case 'settings':
        navigate('/admin/account-settings');
        break;
      case 'home':
        navigate('/');
        break;
      case 'logout':
        logout();
        navigate('/login');
        break;
    }
  };

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    const path = location.pathname;
    
    // 精确匹配
    if (path === '/admin') return ['/admin'];
    
    // 智慧内容相关路径
    if (path.startsWith('/admin/wisdom')) {
      if (path === '/admin/wisdom' || path === '/admin/wisdom/') {
        return ['/admin/wisdom/list'];
      }
      if (path.includes('/add') || path.includes('/edit')) {
        return ['/admin/wisdom/add'];
      }
      return ['/admin/wisdom'];
    }
    
    // 其他路径
    for (const item of menuItems) {
      if (item.path && path.startsWith(item.path) && item.path !== '/admin') {
        return [item.key];
      }
    }
    
    return [path];
  };

  // 获取展开的菜单项
  const getOpenKeys = () => {
    const path = location.pathname;
    const openKeys: string[] = [];
    
    if (path.startsWith('/admin/wisdom')) {
      openKeys.push('/admin/wisdom');
    }
    
    return openKeys;
  };

  // 侧边栏内容
  const sidebarContent = (
    <div className="h-full flex flex-col">
      {/* Logo 区域 */}
      <div className="h-16 flex items-center justify-center border-b border-gray-200">
        {collapsed && !isMobile ? (
          <div className="text-2xl font-bold text-blue-600">太</div>
        ) : (
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold">太</span>
            </div>
            <span className="text-lg font-bold text-gray-800">管理后台</span>
          </div>
        )}
      </div>

      {/* 菜单 */}
      <div className="flex-1 overflow-y-auto">
        <Menu
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
          className="border-none"
        />
      </div>

      {/* 底部信息 */}
      {!collapsed && !isMobile && (
        <div className="p-4 border-t border-gray-200">
          <div className="text-xs text-gray-500 text-center">
            <div>太上老君文化智慧平台</div>
            <div className="mt-1">管理系统 v1.0</div>
          </div>
        </div>
      )}
    </div>
  );

  return (
    <Layout className="min-h-screen">
      {/* 桌面端侧边栏 */}
      {!isMobile && (
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          width={240}
          collapsedWidth={80}
          className="bg-white shadow-sm"
        >
          {sidebarContent}
        </Sider>
      )}

      {/* 移动端抽屉 */}
      {isMobile && (
        <Drawer
          title="管理后台"
          placement="left"
          onClose={() => setMobileDrawerVisible(false)}
          open={mobileDrawerVisible}
          bodyStyle={{ padding: 0 }}
          width={240}
        >
          {sidebarContent}
        </Drawer>
      )}

      <Layout>
        {/* 顶部导航栏 */}
        <Header className="bg-white shadow-sm px-4 flex items-center justify-between">
          <div className="flex items-center space-x-4">
            {isMobile && (
              <Button
                type="text"
                icon={<MenuOutlined />}
                onClick={() => setMobileDrawerVisible(true)}
              />
            )}
            
            <div>
              <Title level={4} className="mb-0">
                {(() => {
                  const path = location.pathname;
                  if (path === '/admin') return '仪表板';
                  if (path.startsWith('/admin/wisdom')) return '智慧内容管理';
                  if (path.startsWith('/admin/categories')) return '分类管理';
                  if (path.startsWith('/admin/tags')) return '标签管理';
                  if (path.startsWith('/admin/analytics')) return '数据分析';
                  if (path.startsWith('/admin/users')) return '用户管理';
                  if (path.startsWith('/admin/settings')) return '系统设置';
                  return '管理后台';
                })()}
              </Title>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            {/* 通知 */}
            <Badge count={3} size="small">
              <Button
                type="text"
                icon={<BellOutlined />}
                className="flex items-center justify-center"
              />
            </Badge>

            {/* 用户信息 */}
            <Dropdown
              menu={{
                items: userMenuItems,
                onClick: handleUserMenuClick
              }}
              placement="bottomRight"
            >
              <div className="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 px-2 py-1 rounded">
                <Avatar
                  size="small"
                  icon={<UserOutlined />}
                  src={user?.avatar}
                />
                {!isMobile && (
                  <div className="flex flex-col">
                    <Text className="text-sm font-medium">
                      {user?.name || '管理员'}
                    </Text>
                    <Text className="text-xs text-gray-500">
                      {user?.role || '系统管理员'}
                    </Text>
                  </div>
                )}
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* 主要内容区域 */}
        <Content className="p-6 bg-gray-50">
          <div className="max-w-full">
            {children || <Outlet />}
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default AdminLayout;