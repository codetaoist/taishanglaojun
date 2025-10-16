import React, { useState, useEffect, useMemo } from 'react';
import { Layout, Menu, Button, Avatar, Badge, Divider, Tooltip, Typography, Space, Progress, Popover } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  CloseOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  CrownOutlined,
  ExperimentOutlined,
  BugOutlined,
  FireOutlined,
  StarOutlined,
  ThunderboltOutlined,
  GiftOutlined,
  TrophyOutlined
} from '@ant-design/icons';
import { useTheme, useBreakpoint } from '../../hooks';
import { useAuthContext as useAuth } from '../../contexts/AuthContext';
import { useMenu } from '../../contexts/MenuContext';
import { type MenuItem } from '../../config/menuConfig';
import { type UserPermissions } from '../../services/frontendMenuService';

const { Text } = Typography;
const { Sider } = Layout;

interface SidebarProps {
  collapsed: boolean;
  isMobile: boolean;
  onClose: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, isMobile, onClose }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { theme } = useTheme();
  const { menuItems, loading, loadMenuData } = useMenu();
  const [openKeys, setOpenKeys] = useState<string[]>([]);
  const { t } = useTranslation();

  // 获取动态菜单数据
  useEffect(() => {
    const loadMenu = async () => {
      if (!user) {
        console.log('🚫 Sidebar - 用户未登录，跳过菜单加载');
        return;
      }

      console.log('🚀 Sidebar - useEffect 触发，开始加载菜单');
      console.log('👤 当前用户:', user);
      console.log('🔍 用户详细信息:', {
        id: user.id,
        username: user.username,
        role: user.role,
        roles: user.roles,
        permissions: user.permissions,
        isAdmin: user.isAdmin
      });
      
      const userPermissions: UserPermissions = {
        roles: (user.roles && user.roles.length > 0) ? user.roles : ['user'],
        permissions: user.permissions || []
      };
      
      console.log('🔑 构建的权限对象:', userPermissions);
      console.log('🔍 权限对象详细信息:', {
        roles类型: typeof userPermissions.roles,
        roles内容: userPermissions.roles,
        roles长度: userPermissions.roles?.length,
        permissions类型: typeof userPermissions.permissions,
        permissions内容: userPermissions.permissions,
        permissions长度: userPermissions.permissions?.length
      });
      console.log('📞 Sidebar - 调用 MenuContext.loadMenuData...');
      
      await loadMenuData(userPermissions);
    };

    loadMenu();
  }, [user?.id, loadMenuData]); // 只依赖用户ID和loadMenuData函数，避免无限循环

  // 渲染状态标签
  const renderStatusBadge = (status?: string) => {
    if (!status || status === 'completed') return null;

    const statusConfig = {
      beta: { color: '#ff7a00', text: t('sidebar.status.beta'), icon: <ExperimentOutlined /> },
      new: { color: '#52c41a', text: t('sidebar.status.new'), icon: <FireOutlined /> },
      experimental: { color: '#722ed1', text: t('sidebar.status.experimental'), icon: <BugOutlined /> },
      partial: { color: '#1890ff', text: t('sidebar.status.partial'), icon: <SettingOutlined /> },
      development: { color: '#f5222d', text: t('sidebar.status.development'), icon: <CrownOutlined /> }
    } as const;

    const config = statusConfig[status as keyof typeof statusConfig];
    if (!config) return null;

    return (
      <Tooltip title={t('sidebar.statusTooltip', { text: config.text })}>
        <Badge
          count={config.text}
          style={{
            backgroundColor: config.color,
            fontSize: '10px',
            height: '18px',
            lineHeight: '18px',
            borderRadius: '9px',
            marginLeft: '8px',
            boxShadow: `0 2px 4px ${config.color}40`
          }}
        />
      </Tooltip>
    );
  };

  // 转换菜单配置为Ant Design Menu格式
  const convertToMenuItems = (items: MenuItem[]): any[] => {
    return items.map(item => {
      const hasChildren = !!(item.children && item.children.length > 0);
      // 统一的菜单标签翻译与回退：优先使用 labelKey 的翻译；无翻译则回退到 label；若 label 为空且为 mainMenu.labels.* 键，则提取最后一段作为回退
      const fallbackFromKey = (k?: string) => {
        if (!k) return '';
        const last = k.split('.').pop() || '';
        return k.startsWith('mainMenu.labels.') ? last : k;
      };
      const getLabel = (menuItem: MenuItem) => {
        const key = (menuItem as any).labelKey as string | undefined;
        const base = (menuItem as any).label as string | undefined;
        const safeBase = base && !String(base).startsWith('mainMenu.labels.') ? base : undefined;
        return key ? t(key, { defaultValue: safeBase ?? fallbackFromKey(key) }) : (safeBase ?? '');
      };
      const tooltipTitle = (
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <span style={{ fontWeight: 600 }}>{getLabel(item)}</span>
          {hasChildren && (
            <span style={{ marginTop: 4, opacity: 0.85 }}>
              {item.children!.map(child => getLabel(child)).join('、')}
            </span>
          )}
        </div>
      );

      const iconNode = (collapsed && !isMobile)
        ? hasChildren
          ? (
            <Popover
              placement="right"
              trigger={["hover", "focus"]}
              overlayStyle={{ padding: 0 }}
              content={
                <div
                  role="menu"
                  aria-label={getLabel(item)}
                  style={{ minWidth: 180 }}
                >
                  {item.children!.map(child => (
                    <div
                      key={child.key}
                      role="menuitem"
                      tabIndex={0}
                      aria-label={getLabel(child)}
                      aria-disabled={!child.path}
                      style={{
                        padding: '8px 12px',
                        cursor: child.path ? 'pointer' : 'default',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        gap: 10
                      }}
                      onClick={() => {
                        if (child.path) {
                          navigate(child.path);
                          if (isMobile) onClose();
                        }
                      }}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          if (child.path) {
                            navigate(child.path);
                            if (isMobile) onClose();
                          }
                        }
                      }}
                    >
                      <span style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          {child.icon && <span style={{ fontSize: 14, lineHeight: 0 }}>{child.icon}</span>}
                          <span title={getLabel(child)}>{getLabel(child)}</span>
                        </span>
                        {child.description && (
                          <span style={{
                            fontSize: 12,
                            color: theme === 'dark' ? 'rgba(255,255,255,0.55)' : 'rgba(0,0,0,0.45)'
                          }} title={child.description}>{child.description}</span>
                        )}
                      </span>
                      {renderStatusBadge(child.status)}
                    </div>
                  ))}
                </div>
              }
            >
              <span
                role="button"
                tabIndex={0}
                aria-label={getLabel(item)}
                aria-haspopup="menu"
                style={{ outline: 'none' }}
              >
                {item.icon}
              </span>
            </Popover>
          )
          : (
            <Tooltip title={tooltipTitle} placement="right" trigger={["hover", "focus"]}>
              <span
                role="button"
                tabIndex={0}
                aria-label={getLabel(item)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    if (item.path) {
                      navigate(item.path);
                      if (isMobile) onClose();
                    }
                  }
                }}
                style={{ outline: 'none' }}
              >
                {item.icon}
              </span>
            </Tooltip>
          )
        : item.icon;

      const displayLabel = getLabel(item);
      const menuItem: any = {
        key: item.key,
        icon: iconNode,
        label: collapsed && !isMobile ? (
          // 收缩状态下隐藏标签，仅在悬停图标时显示Tooltip（含子菜单信息）
          <span style={{ display: 'none' }}>{displayLabel}</span>
        ) : (
          // 展开状态下，显示标签和状态徽章
          <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span>{displayLabel}</span>
            {renderStatusBadge(item.status)}
          </span>
        ),
        onClick: item.path ? () => {
          navigate(item.path!);
          if (isMobile) {
            onClose();
          }
        } : undefined
      };

      if (item.children && item.children.length > 0) {
        menuItem.children = convertToMenuItems(item.children);
      }

      return menuItem;
    });
  };

  const antdMenuItems = convertToMenuItems(menuItems);

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    if (loading || menuItems.length === 0) return [];
    
    const findSelectedKey = (items: MenuItem[], path: string): string | null => {
      for (const item of items) {
        if (item.path === path) {
          return item.key;
        }
        if (item.children) {
          const childKey = findSelectedKey(item.children, path);
          if (childKey) return childKey;
        }
      }
      return null;
    };

    const selectedKey = findSelectedKey(menuItems, location.pathname);
    return selectedKey ? [selectedKey] : [];
  };

  // 获取默认展开的菜单项
  const getDefaultOpenKeys = () => {
    if (loading || menuItems.length === 0) return [];
    
    const findParentKeys = (items: MenuItem[], targetPath: string, parentKey?: string): string[] => {
      for (const item of items) {
        if (item.path === targetPath && parentKey) {
          return [parentKey];
        }
        if (item.children) {
          const result = findParentKeys(item.children, targetPath, item.key);
          if (result.length > 0) {
            return parentKey ? [parentKey, ...result] : result;
          }
        }
      }
      return [];
    };

    return findParentKeys(menuItems, location.pathname);
  };

  // 处理子菜单展开/收起
  const handleOpenChange = (keys: string[]) => {
    setOpenKeys(keys);
  };

  return (
    <div style={{
      height: '100%',
      background: theme === 'dark' ? '#001529' : '#ffffff',
      borderRight: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
    }}>
      {/* 顶部Logo区域（使用 SVG Logo，文案“太上老君”，点击跳转 /dashboard） */}
      {(!collapsed || isMobile) && (
        <div style={{
          padding: '16px 24px',
          borderBottom: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <div
            style={{ display: 'flex', alignItems: 'center', gap: '12px', cursor: 'pointer' }}
            onClick={() => navigate('/dashboard')}
            title={t('sidebar.logo.title')}
          >
            <img
              src="/laojun-avatar.svg"
              alt={t('sidebar.logo.name') + ' Logo'}
              style={{ width: 32, height: 32, borderRadius: 8, boxShadow: '0 2px 8px rgba(24, 144, 255, 0.15)' }}
            />
            <div style={{
              fontSize: '16px',
              fontWeight: 600,
              color: theme === 'dark' ? '#ffffff' : 'rgba(0, 0, 0, 0.85)',
            }}>
              {t('sidebar.logo.name')}
            </div>
          </div>
          
          {isMobile && (
            <Button
              type="text"
              icon={<CloseOutlined />}
              onClick={onClose}
              style={{
                color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)',
                border: 'none',
              }}
            />
          )}
        </div>
      )}

      {/* 收缩状态下的Logo（使用 SVG，提示与点击跳转 /dashboard） */}
      {collapsed && !isMobile && (
        <div style={{
          padding: '16px',
          borderBottom: `1px solid ${theme === 'dark' ? '#303030' : '#f0f0f0'}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}>
          <Tooltip title={t('sidebar.logo.name')} placement="right">
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: 8,
                cursor: 'pointer',
              }}
              onClick={() => navigate('/dashboard')}
            >
              <img
                src="/laojun-avatar.svg"
                alt={t('sidebar.logo.name') + ' Logo'}
                style={{ width: 32, height: 32, borderRadius: 8, boxShadow: '0 2px 8px rgba(24, 144, 255, 0.15)' }}
              />
            </div>
          </Tooltip>
        </div>
      )}

      {/* 菜单区域 */}
      <div style={{ 
        height: collapsed && !isMobile ? 'calc(100% - 65px)' : 'calc(100% - 65px)', 
        overflow: 'auto',
        // 自定义滚动条样式
        scrollbarWidth: 'thin',
        scrollbarColor: theme === 'dark' ? '#434343 #1f1f1f' : '#c1c1c1 #f1f1f1',
      }}>
        {loading ? (
          <div style={{ 
            padding: '20px', 
            textAlign: 'center',
            color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)'
          }}>
            <div>{t('sidebar.loading')}</div>
            <Progress percent={50} showInfo={false} size="small" style={{ marginTop: '8px' }} />
          </div>
        ) : menuItems.length === 0 ? (
          <div style={{ 
            padding: '20px', 
            textAlign: 'center',
            color: theme === 'dark' ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.65)'
          }}>
            <div>{t('sidebar.noData.title')}</div>
            <div style={{ fontSize: '12px', marginTop: '8px', opacity: 0.7 }}>
              {user ? t('sidebar.noData.noteLogged') : t('sidebar.noData.noteGuest')}
            </div>
            {user && (
              <Button 
                type="link" 
                size="small" 
                onClick={() => {
                  const userPermissions: UserPermissions = {
                    roles: ((user.roles && user.roles.length > 0) ? user.roles : ['user']).map(r => r.toLowerCase()),
                    permissions: (user.permissions || []).map(p => p.toLowerCase())
                  };
                  loadMenuData(userPermissions);
                }}
                style={{ marginTop: '8px' }}
              >
                {t('sidebar.reload')}
              </Button>
            )}
          </div>
        ) : (
          <>
            {/* 移除开发模式下的菜单统计信息显示 */}
            <Menu
              mode="inline"
              theme={theme === 'dark' ? 'dark' : 'light'}
              inlineCollapsed={collapsed && !isMobile}
              selectedKeys={getSelectedKeys()}
              defaultOpenKeys={getDefaultOpenKeys()}
              openKeys={collapsed && !isMobile ? [] : openKeys}
              onOpenChange={handleOpenChange}
              onClick={({ key }) => {
                // 查找菜单项并处理点击
                const findMenuItem = (items: MenuItem[], targetKey: string): MenuItem | null => {
                  for (const item of items) {
                    if (item.key === targetKey) {
                      return item;
                    }
                    if (item.children) {
                      const found = findMenuItem(item.children, targetKey);
                      if (found) return found;
                    }
                  }
                  return null;
                };
                
                const menuItem = findMenuItem(menuItems, key);
                if (menuItem && menuItem.path) {
                  navigate(menuItem.path);
                  if (isMobile) {
                    onClose();
                  }
                }
              }}
              items={antdMenuItems}
              style={{
                border: 'none',
                background: 'transparent',
                height: '100%',
                // 确保菜单项图标在收缩状态下居中显示
                ...(collapsed && !isMobile && {
                  '.ant-menu-item': {
                    paddingLeft: '24px !important',
                    textAlign: 'center',
                  },
                  '.ant-menu-item .ant-menu-item-icon': {
                    fontSize: '16px',
                    marginInlineEnd: '0 !important',
                  },
                  '.ant-menu-submenu-title': {
                    paddingLeft: '24px !important',
                    textAlign: 'center',
                  },
                  '.ant-menu-submenu-title .ant-menu-item-icon': {
                    fontSize: '16px',
                    marginInlineEnd: '0 !important',
                  }
                })
              }}
            />
          </>
        )}
      </div>
    </div>
  );
};

export default Sidebar;