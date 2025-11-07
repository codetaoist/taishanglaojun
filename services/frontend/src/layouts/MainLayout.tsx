import React, { useState } from 'react';
import { Layout, Menu, theme } from 'antd';
import { 
  DashboardOutlined, 
  SettingOutlined, 
  AppstoreOutlined, 
  FileTextOutlined,
  DatabaseOutlined,
  ExperimentOutlined,
  HistoryOutlined
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const {
    token: { colorBgContainer },
  } = theme.useToken();
  
  const navigate = useNavigate();
  const location = useLocation();

  // 菜单项配置
  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/laojun',
      icon: <SettingOutlined />,
      label: '老君域',
      children: [
        {
          key: '/laojun/config',
          icon: <SettingOutlined />,
          label: '配置管理',
        },
        {
          key: '/laojun/plugins',
          icon: <AppstoreOutlined />,
          label: '插件管理',
        },
        {
          key: '/laojun/audit-logs',
          icon: <HistoryOutlined />,
          label: '审计日志',
        },
      ],
    },
    {
      key: '/taishang',
      icon: <DatabaseOutlined />,
      label: '太上域',
      children: [
        {
          key: '/taishang/models',
          icon: <ExperimentOutlined />,
          label: '模型管理',
        },
        {
          key: '/taishang/collections',
          icon: <DatabaseOutlined />,
          label: '向量集合',
        },
        {
          key: '/taishang/tasks',
          icon: <FileTextOutlined />,
          label: '任务管理',
        },
      ],
    },
  ];

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    return [location.pathname];
  };

  // 获取当前展开的菜单项
  const getOpenKeys = () => {
    const path = location.pathname;
    if (path.startsWith('/laojun')) return ['/laojun'];
    if (path.startsWith('/taishang')) return ['/taishang'];
    return [];
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed}>
        <div 
          style={{
            height: 32,
            margin: 16,
            background: 'rgba(255, 255, 255, 0.2)',
            borderRadius: 6,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            fontWeight: 'bold'
          }}
        >
          {collapsed ? 'TL' : 'TaiShangLaoJun'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: 0,
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between'
          }}
        >
          <div style={{ padding: '0 24px' }}>
            {/* 这里可以添加面包屑导航 */}
          </div>
          <div style={{ padding: '0 24px' }}>
            {/* 这里可以添加用户信息、通知等 */}
          </div>
        </Header>
        <Content
          style={{
            margin: '24px 16px',
            padding: 24,
            minHeight: 280,
            background: colorBgContainer,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;