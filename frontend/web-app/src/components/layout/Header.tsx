import React, { useState } from 'react';
import { Layout, Button, Avatar, Dropdown, Badge, Space, Tooltip, Switch, type MenuProps } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SearchOutlined,
  BellOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  GlobalOutlined,
  SunOutlined,
  MoonOutlined,
  QuestionCircleOutlined,
  CustomerServiceOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useApp, useTheme, useLanguage, useNotifications, useGlobalSearch, useBreakpoint } from '../../hooks';
import { useTranslation } from 'react-i18next';
import { useAuthContext as useAuth } from '../../contexts/AuthContext';
import GlobalSearch from '../common/GlobalSearch';
import NotificationCenter from '../common/NotificationCenter';
import CustomerService from '../common/CustomerService';
import { supportedLanguages, changeLanguage } from '../../i18n';

const { Header: AntHeader } = Layout;

interface HeaderProps {
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

export const Header: React.FC<HeaderProps> = ({ collapsed, onCollapse }) => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const { language, toggleLanguage } = useLanguage();
  const { unreadCount } = useNotifications();
  const { setVisible: setSearchVisible } = useGlobalSearch();
  const { isMobile } = useBreakpoint();
  const { setLanguage } = useApp();
  const { t } = useTranslation();
  
  const [notificationVisible, setNotificationVisible] = useState(false);
  const [customerServiceVisible, setCustomerServiceVisible] = useState(false);

  // 语言切换菜单项
  const languageMenuItems: MenuProps['items'] = supportedLanguages.map(lang => ({
    key: lang.code,
    label: (
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }} onClick={() => handleLanguageChange(lang.code)}>
        <span>{lang.flag}</span>
        <span>{lang.name}</span>
      </div>
    )
  }));

  const handleLanguageChange = async (code: string) => {
    try {
      await changeLanguage(code);
      // 同步到Redux（如果使用）
      // 仅支持 zh-CN 与 en-US 的轻量切换，其它语言走 i18next 即可
      if (code === 'zh-CN' || code === 'en-US') {
        setLanguage(code as 'zh-CN' | 'en-US');
      }
      localStorage.setItem('language', code);
    } catch (e) {
      // 忽略切换错误，保持当前语言
    }
  };

  // 用户菜单项
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/profile')
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
      onClick: () => navigate('/settings')
    },
    {
      type: 'divider' as const
    },
    {
      key: 'help',
      icon: <QuestionCircleOutlined />,
      label: '帮助中心',
      onClick: () => navigate('/help')
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: logout,
      danger: true
    }
  ];

  // 处理搜索快捷键
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        setSearchVisible(true);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [setSearchVisible]);

  return (
    <>
      <AntHeader 
        style={{
          padding: '0 24px',
          background: theme === 'dark' ? '#001529' : '#ffffff',
          borderBottom: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: 48,
        }}
      >
        {/* 左侧区域 */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          {/* 折叠按钮 */}
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => onCollapse(!collapsed)}
            style={{
              fontSize: '16px',
              color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
              border: 'none',
              background: 'transparent',
            }}
          />

          {/* 移除顶部Logo与标题显示 */}
        </div>

        {/* 右侧区域 */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {/* 功能按钮组 */}
          <Space size="small">
            {/* 搜索按钮 */}
            <Tooltip title={t('header.search.tooltip', { shortcut: `${navigator.platform.includes('Mac') ? '⌘' : 'Ctrl'}+K` })}>
              <Button
                type="text"
                icon={<SearchOutlined />}
                onClick={() => setSearchVisible(true)}
                aria-label={t('header.search.aria')}
                style={{
                  color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                  border: 'none',
                }}
              />
            </Tooltip>

            {/* 通知中心 */}
            <Tooltip title={t('header.notifications.title')}>
              <Badge count={unreadCount} size="small">
                <Button
                  type="text"
                  icon={<BellOutlined />}
                  onClick={() => setNotificationVisible(true)}
                  aria-label={t('header.notifications.title')}
                  style={{
                    color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                    border: 'none',
                  }}
                />
              </Badge>
            </Tooltip>

            {/* 主题切换 */}
            <Tooltip title={theme === 'light' ? t('header.theme.switchToDark') : t('header.theme.switchToLight')}>
              <Button
                type="text"
                icon={theme === 'light' ? <MoonOutlined /> : <SunOutlined />}
                onClick={toggleTheme}
                aria-label={theme === 'light' ? t('header.theme.switchToDark') : t('header.theme.switchToLight')}
                style={{
                  color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                  border: 'none',
                }}
              />
            </Tooltip>

            {/* 帮助中心 */}
            <Tooltip title={t('header.help')}>
              <Button
                type="text"
                icon={<QuestionCircleOutlined />}
                onClick={() => navigate('/help')}
                aria-label={t('header.help')}
                style={{
                  color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                  border: 'none',
                }}
              />
            </Tooltip>

            {/* 语言切换 */}
            <Dropdown
              menu={{ items: languageMenuItems }}
              placement="bottomRight"
              trigger={["click", "hover"]}
            >
            <Button
              type="text"
              icon={<GlobalOutlined />}
              aria-label={t('header.language')}
              style={{
                color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                border: 'none',
              }}
            />
            </Dropdown>
          </Space>

          {/* 用户菜单 */}
          <Dropdown
            menu={{ items: userMenuItems }}
            placement="bottomRight"
            trigger={['click']}
          >
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              cursor: 'pointer',
              padding: '4px 8px',
              borderRadius: '6px',
              transition: 'background-color 0.3s',
            }}>
              <Avatar
                size={24}
                src={user?.avatar || user?.avatar_url}
                icon={<UserOutlined />}
                style={{ backgroundColor: '#1890ff' }}
              />
              {!isMobile && (
                <span style={{
                  fontSize: '14px',
                  color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                  maxWidth: '100px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}>
                  {user?.display_name || user?.first_name || user?.username || user?.email || '用户'}
                </span>
              )}
            </div>
          </Dropdown>
        </div>
      </AntHeader>

      {/* 全局搜索组件 */}
      <GlobalSearch />

      {/* 通知中心 */}
      <NotificationCenter
        visible={notificationVisible}
        onClose={() => setNotificationVisible(false)}
      />

      {/* 客服支持 */}
      <CustomerService
        visible={customerServiceVisible}
        onClose={() => setCustomerServiceVisible(false)}
      />
    </>
  );
};

export default Header;