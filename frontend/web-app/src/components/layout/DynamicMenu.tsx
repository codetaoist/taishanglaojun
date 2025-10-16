/**
 * 动态菜单组件
 * 根据用户权限和角色动态渲染菜单
 */

import React, { useMemo, useState, useEffect } from 'react';
import { Menu, Badge, Tooltip, Spin } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import { usePermissions } from '../../hooks/usePermissions';
import { useMenu } from '../../contexts/MenuContext';
import { 
  getStatusBadge,
  getPriorityColor,
  type MenuItem 
} from '../../config/menuConfig';
import { type UserPermissions } from '../../services/frontendMenuService';

interface DynamicMenuProps {
  mode?: 'horizontal' | 'vertical' | 'inline';
  theme?: 'light' | 'dark';
  collapsed?: boolean;
  includeStatuses?: ('completed' | 'partial' | 'planned')[];
  showBadges?: boolean;
  showTooltips?: boolean;
  className?: string;
  style?: React.CSSProperties;
  onMenuClick?: (key: string, item: MenuItem) => void;
}

export const DynamicMenu: React.FC<DynamicMenuProps> = ({
  mode = 'inline',
  theme = 'light',
  collapsed = false,
  includeStatuses = ['completed', 'partial'],
  showBadges = true,
  showTooltips = true,
  className,
  style,
  onMenuClick
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { userRoles, userPermissions } = usePermissions();
  const { menuItems: contextMenuItems, loading: contextLoading, loadMenuData } = useMenu();
  const [filteredMenuItems, setFilteredMenuItems] = useState<MenuItem[]>([]);

  // 加载菜单数据
  useEffect(() => {
    const loadMenu = async () => {
      if (!userRoles && !userPermissions) {
        console.log('🚫 DynamicMenu - 用户权限未加载，跳过菜单加载');
        return;
      }

      console.log('🎯 DynamicMenu - 开始加载菜单数据');
      console.log('👤 用户角色:', userRoles);
      console.log('🔑 用户权限:', userPermissions);
      
      const permissions: UserPermissions = {
        roles: ((userRoles && userRoles.length > 0) ? userRoles : ['user']).map(r => r.toLowerCase()),
        permissions: (userPermissions || []).map(p => p.toLowerCase())
      };
      console.log('📋 构建的权限对象:', permissions);
      
      await loadMenuData(permissions);
    };

    loadMenu();
  }, [userRoles?.join(','), userPermissions?.join(','), loadMenuData]); // 使用字符串化的权限避免引用变化

  // 根据状态过滤菜单
  useEffect(() => {
    console.log('🔍 DynamicMenu - 开始状态过滤');
    console.log('📥 原始菜单数据:', contextMenuItems);
    console.log('📊 包含的状态:', includeStatuses);
    
    const filteredByStatus = contextMenuItems.filter(item => 
      includeStatuses.includes(item.status as any)
    );
    
    console.log('🔍 状态过滤后的菜单:', filteredByStatus);
    setFilteredMenuItems(filteredByStatus);
  }, [contextMenuItems, includeStatuses]);

  // 转换菜单项为 Ant Design Menu 格式
  const convertToAntdMenuItems = (items: MenuItem[]): any[] => {
    return items.map(item => {
      const statusBadge = getStatusBadge(item.status);
      const priorityColor = getPriorityColor(item.priority);
      
      // 构建标签
      let label: React.ReactNode = (
        <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {item.icon}
          <span>{item.label}</span>
          {showBadges && statusBadge.text && (
            <Badge 
              count={statusBadge.text} 
              style={{ 
                backgroundColor: statusBadge.color,
                fontSize: '10px',
                height: '16px',
                lineHeight: '16px',
                minWidth: '16px'
              }} 
            />
          )}
          {item.badge && (
            <Badge 
              count={item.badge} 
              style={{ backgroundColor: priorityColor }} 
            />
          )}
        </span>
      );

      // 添加工具提示
      if (showTooltips && item.description && !collapsed) {
        label = (
          <Tooltip title={item.description} placement="right">
            {label}
          </Tooltip>
        );
      }

      const menuItem: any = {
        key: item.key,
        label,
        icon: collapsed ? item.icon : undefined,
        disabled: item.status === 'planned'
      };

      // 处理子菜单
      if (item.children && item.children.length > 0) {
        menuItem.children = convertToAntdMenuItems(item.children);
      }

      return menuItem;
    });
  };

  const antdMenuItems = useMemo(() => {
    if (contextLoading || filteredMenuItems.length === 0) return [];
    return convertToAntdMenuItems(filteredMenuItems);
  }, [filteredMenuItems, collapsed, showBadges, showTooltips, contextLoading]);

  // 获取当前选中的菜单项
  const selectedKeys = useMemo(() => {
    if (contextLoading || filteredMenuItems.length === 0) return [];
    
    const findSelectedKey = (items: MenuItem[], path: string): string[] => {
      for (const item of items) {
        if (item.path === path) {
          return [item.key];
        }
        if (item.children) {
          const found = findSelectedKey(item.children, path);
          if (found.length > 0) {
            return found;
          }
        }
      }
      return [];
    };

    return findSelectedKey(filteredMenuItems, location.pathname);
  }, [filteredMenuItems, location.pathname, contextLoading]);

  // 获取展开的菜单项
  const openKeys = useMemo(() => {
    if (contextLoading || filteredMenuItems.length === 0) return [];
    
    const findOpenKeys = (items: MenuItem[], path: string): string[] => {
      for (const item of items) {
        if (item.children) {
          const found = findOpenKeys(item.children, path);
          if (found.length > 0) {
            return [item.key, ...found];
          }
        }
        if (item.path === path) {
          return [];
        }
      }
      return [];
    };

    return findOpenKeys(filteredMenuItems, location.pathname);
  }, [filteredMenuItems, location.pathname, contextLoading]);

  // 查找菜单项
  const findMenuItem = (key: string, items: MenuItem[] = filteredMenuItems): MenuItem | null => {
    for (const item of items) {
      if (item.key === key) {
        return item;
      }
      if (item.children) {
        const found = findMenuItem(key, item.children);
        if (found) {
          return found;
        }
      }
    }
    return null;
  };

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    const menuItem = findMenuItem(key);
    if (menuItem) {
      // 调用自定义点击处理器
      onMenuClick?.(key, menuItem);
      
      // 如果有路径，则导航
      if (menuItem.path) {
        navigate(menuItem.path);
      }
    }
  };

  if (contextLoading) {
    return (
      <div style={{ 
        padding: '20px', 
        textAlign: 'center',
        ...style 
      }}>
        <Spin size="small" />
        <div style={{ marginTop: '8px', fontSize: '12px', opacity: 0.6 }}>
          加载菜单中...
        </div>
      </div>
    );
  }

  return (
    <Menu
      mode={mode}
      theme={theme}
      selectedKeys={selectedKeys}
      defaultOpenKeys={openKeys}
      items={antdMenuItems}
      onClick={handleMenuClick}
      className={className}
      style={style}
      inlineCollapsed={collapsed && mode === 'inline'}
    />
  );
};

/**
 * 侧边栏菜单组件
 */
export const SidebarMenu: React.FC<{
  collapsed?: boolean;
  onMenuClick?: (key: string, item: MenuItem) => void;
}> = ({ collapsed = false, onMenuClick }) => {
  return (
    <DynamicMenu
      mode="inline"
      theme="dark"
      collapsed={collapsed}
      onMenuClick={onMenuClick}
      style={{ height: '100%', borderRight: 0 }}
    />
  );
};

/**
 * 顶部菜单组件
 */
export const TopMenu: React.FC<{
  onMenuClick?: (key: string, item: MenuItem) => void;
}> = ({ onMenuClick }) => {
  return (
    <DynamicMenu
      mode="horizontal"
      theme="light"
      onMenuClick={onMenuClick}
      includeStatuses={['completed']}
      showBadges={false}
      showTooltips={false}
    />
  );
};

/**
 * 移动端菜单组件
 */
export const MobileMenu: React.FC<{
  onMenuClick?: (key: string, item: MenuItem) => void;
}> = ({ onMenuClick }) => {
  return (
    <DynamicMenu
      mode="inline"
      theme="light"
      onMenuClick={onMenuClick}
      showTooltips={false}
      style={{ border: 'none' }}
    />
  );
};

export default DynamicMenu;