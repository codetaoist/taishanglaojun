import React from 'react';
import { Layout, Menu } from 'antd';
import { 
  HomeOutlined, 
  BookOutlined, 
  MessageOutlined, 
  TeamOutlined,
  SettingOutlined,
  UserOutlined,
  HeartOutlined,
  SearchOutlined,
  RobotOutlined,
  EditOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';

const { Sider } = Layout;

interface SidebarProps {
  collapsed: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed }) => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/',
      icon: <HomeOutlined />,
      label: '首页',
    },
    {
      key: '/chat',
      icon: <MessageOutlined />,
      label: 'AI对话',
    },
    {
      key: '/api-test',
      icon: <SettingOutlined />,
      label: 'API测试',
    },
    {
      key: '/wisdom',
      icon: <BookOutlined />,
      label: '文化智慧',
      children: [
        {
          key: '/wisdom/browse',
          label: '浏览智慧',
        },
        {
          key: '/wisdom/search',
          icon: <SearchOutlined />,
          label: '智慧搜索',
        },
        {
          key: '/recommendations',
          icon: <RobotOutlined />,
          label: '智慧推荐',
        },
        {
          key: '/favorites',
          icon: <HeartOutlined />,
          label: '我的收藏',
        },
        {
          key: '/notes',
          icon: <EditOutlined />,
          label: '我的笔记',
        },
      ],
    },
    {
      key: '/community',
      icon: <TeamOutlined />,
      label: '社区',
      children: [
        {
          key: '/community/discussions',
          label: '讨论区',
        },
        {
          key: '/community/events',
          label: '活动',
        },
      ],
    },
    {
      type: 'divider' as const,
    },
    {
      key: '/profile',
      icon: <UserOutlined />,
      label: '个人中心',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={collapsed}
      className="bg-gradient-to-b from-slate-900 to-slate-800 shadow-2xl"
      width={240}
      collapsedWidth={80}
    >
      <div className="h-16 flex items-center justify-center border-b border-slate-700">
        {!collapsed ? (
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 bg-gradient-to-r from-cultural-gold to-cultural-red rounded-xl flex items-center justify-center shadow-lg">
              <span className="text-white font-bold text-lg">太</span>
            </div>
            <span className="text-xl font-bold text-white tracking-wide">太上老君</span>
          </div>
        ) : (
          <div className="w-10 h-10 bg-gradient-to-r from-cultural-gold to-cultural-red rounded-xl flex items-center justify-center shadow-lg">
            <span className="text-white font-bold text-lg">太</span>
          </div>
        )}
      </div>
      
      <Menu
        mode="inline"
        selectedKeys={[location.pathname]}
        defaultOpenKeys={['/wisdom', '/chat', '/community']}
        items={menuItems}
        onClick={handleMenuClick}
        className="border-none bg-transparent mt-4"
        style={{ 
          height: 'calc(100vh - 80px)', 
          borderRight: 0,
          backgroundColor: 'transparent'
        }}
        theme="dark"
      />
    </Sider>
  );
};

export default Sidebar;