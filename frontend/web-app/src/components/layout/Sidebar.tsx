import React, { useMemo } from 'react';
import { Layout, Menu, Badge, Tooltip } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { 
  mainMenuConfig, 
  profileMenuConfig,
  filterMenuByPermissions,
  filterMenuByStatus,
  getStatusBadge,
  type MenuItem 
} from '../../config/menuConfig.tsx';

const { Sider } = Layout;

interface SidebarProps {
  collapsed: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();

  // 转换菜单配置为Ant Design Menu格式
  const convertToAntdMenuItems = (items: MenuItem[]): any[] => {
    return items.map(item => {
      const statusBadge = getStatusBadge(item.status);
      
      const menuItem: any = {
        key: item.path || item.key,
        icon: item.icon,
        label: (
          <div className="flex items-center justify-between w-full">
            <span>{item.label}</span>
            {!collapsed && statusBadge.text && (
              <Tooltip title={`状态: ${item.status === 'completed' ? '已完成' : item.status === 'partial' ? '部分完成' : '规划中'}`}>
                <span style={{ fontSize: '12px' }}>{statusBadge.text}</span>
              </Tooltip>
            )}
          </div>
        ),
        title: item.description,
      };

      if (item.children && item.children.length > 0) {
        menuItem.children = convertToAntdMenuItems(item.children);
      }

      return menuItem;
    });
  };

  // 根据用户权限和开发状态过滤菜单
  const filteredMenuItems = useMemo(() => {
    const userRoles = user?.roles || [];
    const userPermissions = user?.permissions || [];
    
    // 首先根据权限过滤
    let filtered = filterMenuByPermissions(mainMenuConfig, userRoles, userPermissions);
    
    // 然后根据开发状态过滤（只显示已完成和部分完成的功能）
    filtered = filterMenuByStatus(filtered, ['completed', 'partial']);
    
    return filtered;
  }, [user]);

  // 个人中心菜单
  const profileMenuItems = useMemo(() => {
    return convertToAntdMenuItems(profileMenuConfig);
  }, []);

  // 主菜单项
  const mainMenuItems = useMemo(() => {
    return convertToAntdMenuItems(filteredMenuItems);
  }, [filteredMenuItems]);

  // 合并所有菜单项
  const allMenuItems = useMemo(() => {
    return [
      ...mainMenuItems,
      {
        type: 'divider' as const,
      },
      ...profileMenuItems,
      {
        key: '/api-test',
        icon: <span>🔧</span>,
        label: 'API测试',
        title: '开发调试工具'
      }
    ];
  }, [mainMenuItems, profileMenuItems]);

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  return (
    <Sider 
      trigger={null} 
      collapsible 
      collapsed={collapsed}
      className="bg-white shadow-md"
      width={240}
    >
      <div className="h-16 flex items-center justify-center border-b border-gray-200">
        <h1 className={`font-bold text-lg ${collapsed ? 'hidden' : 'block'}`}>
          太上老君
        </h1>
        {collapsed && <span className="text-xl">太</span>}
      </div>
      <Menu
        theme="light"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={allMenuItems}
        onClick={handleMenuClick}
        className="border-r-0"
        style={{
          height: 'calc(100vh - 64px)',
          overflowY: 'auto',
        }}
      />
      
      {/* 添加自定义样式 */}
      <style jsx>{`
        .ant-menu-item .flex {
          width: 100%;
        }
        .ant-menu-submenu-title .flex {
          width: 100%;
        }
        .ant-menu-item-selected {
          background-color: #e6f7ff !important;
        }
        .ant-menu-submenu-selected > .ant-menu-submenu-title {
          color: #1890ff !important;
        }
      `}</style>
    </Sider>
  );
};

export default Sidebar;