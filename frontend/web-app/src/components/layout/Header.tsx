import React from 'react';
import { Layout, Menu, Avatar, Dropdown, Space, Button } from 'antd';
import { UserOutlined, LogoutOutlined, SettingOutlined, MenuOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth.tsx';

const { Header: AntHeader } = Layout;

interface HeaderProps {
  collapsed: boolean;
  onToggle: () => void;
}

const Header: React.FC<HeaderProps> = ({ onToggle }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  const mainMenuItems = [
    {
      key: '/',
      label: '首页',
    },
    {
      key: '/wisdom',
      label: '文化智慧',
    },
    {
      key: '/chat',
      label: 'AI对话',
    },
    {
      key: '/community',
      label: '社区',
    },
  ];

  return (
    <AntHeader className="bg-gradient-to-r from-slate-50 to-white shadow-lg border-b border-slate-200 px-6 flex items-center justify-between">
      <div className="flex items-center space-x-6">
        <Button
          type="text"
          icon={<MenuOutlined />}
          onClick={onToggle}
          className="lg:hidden hover:bg-slate-100 text-slate-600"
          size="large"
        />
        
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-gradient-to-r from-cultural-gold to-cultural-red rounded-xl flex items-center justify-center shadow-lg">
            <span className="text-white font-bold text-lg">太</span>
          </div>
          <h1 className="text-2xl font-bold bg-gradient-to-r from-slate-800 to-slate-600 bg-clip-text text-transparent hidden sm:block">
            太上老君
          </h1>
        </div>

        <Menu
          mode="horizontal"
          selectedKeys={[location.pathname]}
          items={mainMenuItems}
          className="border-none bg-transparent hidden md:flex ml-8"
          onClick={({ key }) => navigate(key)}
          style={{
            backgroundColor: 'transparent',
            borderBottom: 'none'
          }}
        />
      </div>

      <div className="flex items-center space-x-4">
        {user ? (
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Space className="cursor-pointer hover:bg-slate-100 px-3 py-2 rounded-lg transition-colors duration-200">
              <Avatar 
                src={user.avatar || undefined} 
                icon={<UserOutlined />}
                size="small"
              />
              <span className="text-gray-700 hidden sm:inline">
                {user.username}
              </span>
            </Space>
          </Dropdown>
        ) : (
          <Space>
            <Button type="text" onClick={() => navigate('/login')}>
              登录
            </Button>
            <Button type="primary" onClick={() => navigate('/register')}>
              注册
            </Button>
          </Space>
        )}
      </div>
    </AntHeader>
  );
};

export default Header;