import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  ChevronUp,
  ChevronDown,
  Menu as MenuIcon,
  LayoutDashboard,
  MessageCircle,
  Settings,
  Shield,
  Users,
  Lock,
  Bell
} from 'lucide-react';
import { useDynamicMenu } from '../contexts/DynamicMenuContext';
import { MenuItem, MenuItemType, DeviceType } from '../types/menu';
import { getRecommendedUIConfig } from '../utils/deviceDetection';

interface DynamicSidebarProps {
  open: boolean;
  onToggle: () => void;
  variant?: 'permanent' | 'persistent' | 'temporary';
}

const iconMap: Record<string, React.ReactElement> = {
  dashboard: <LayoutDashboard className="h-5 w-5" />,
  chat: <MessageCircle className="h-5 w-5" />,
  settings: <Settings className="h-5 w-5" />,
  admin_panel_settings: <Shield className="h-5 w-5" />,
  people: <Users className="h-5 w-5" />,
  security: <Lock className="h-5 w-5" />,
  notifications: <Bell className="h-5 w-5" />,
  message: <MessageCircle className="h-5 w-5" />
};

export const DynamicSidebar: React.FC<DynamicSidebarProps> = ({ 
  open, 
  onToggle, 
  variant = 'temporary' 
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { 
    getFilteredMenuItems, 
    deviceInfo, 
    isLoading, 
    error 
  } = useDynamicMenu();

  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set());
  const [menuItems, setMenuItems] = useState<MenuItem[]>([]);

  // 响应式设计
  const isMobile = deviceInfo.type === DeviceType.MOBILE || deviceInfo.type === DeviceType.WATCH;
  
  // 获取设备特定的UI配置
  const uiConfig = getRecommendedUIConfig(deviceInfo.type);

  // 更新菜单项
  useEffect(() => {
    const filteredItems = getFilteredMenuItems();
    setMenuItems(filteredItems);
  }, [getFilteredMenuItems]);

  // 处理菜单项展开/收起
  const handleExpandClick = (itemId: string) => {
    setExpandedItems(prev => {
      const newSet = new Set(prev);
      if (newSet.has(itemId)) {
        newSet.delete(itemId);
      } else {
        newSet.add(itemId);
      }
      return newSet;
    });
  };

  // 处理菜单项点击
  const handleMenuClick = (item: MenuItem) => {
    if (item.type === MenuItemType.PAGE && item.route) {
      navigate(item.route);
      
      // 在移动设备上点击后关闭侧边栏
      if (isMobile) {
        onToggle();
      }
    } else if (item.type === MenuItemType.ACTION && item.action) {
      // 处理动作类型的菜单项
      console.log('Execute action:', item.action);
    } else if (item.type === MenuItemType.SUBMENU) {
      handleExpandClick(item.id);
    }
  };

  // 检查菜单项是否为当前活动项
  const isActiveItem = (item: MenuItem): boolean => {
    if (item.route) {
      return location.pathname === item.route;
    }
    return false;
  };

  // 渲染菜单图标
  const renderIcon = (item: MenuItem) => {
    const icon = item.icon ? iconMap[item.icon] : <LayoutDashboard className="h-5 w-5" />;
    
    if (uiConfig.showIcons) {
      return (
        <div 
          className={`${uiConfig.compactMode ? 'min-w-10' : 'min-w-14'} flex items-center justify-center ${
            isActiveItem(item) ? 'text-blue-600' : 'text-gray-600'
          }`}
        >
          {icon}
        </div>
      );
    }
    
    return null;
  };

  // 渲染菜单文本
  const renderText = (item: MenuItem) => {
    if (!uiConfig.showLabels && uiConfig.compactMode) {
      return null;
    }

    return (
      <div className="flex-1">
        <div 
          className={`${
            uiConfig.fontSize === 'small' ? 'text-sm' : 'text-base'
          } ${
            isActiveItem(item) ? 'font-semibold' : 'font-normal'
          }`}
        >
          {item.title}
        </div>
        {!uiConfig.compactMode && item.description && (
          <div className="text-xs text-gray-500 mt-1">
            {item.description}
          </div>
        )}
      </div>
    );
  };

  // 渲染单个菜单项
  const renderMenuItem = (item: MenuItem, level = 0) => {
    const hasChildren = item.children && item.children.length > 0;
    const isExpanded = expandedItems.has(item.id);
    const isActive = isActiveItem(item);

    if (item.type === MenuItemType.SEPARATOR) {
      return <div key={item.id} className="border-t border-gray-200 my-2" />;
    }

    const menuContent = (
      <div key={item.id} className="w-full">
        <button
          onClick={() => handleMenuClick(item)}
          className={`w-full flex items-center px-4 py-2 text-left transition-colors duration-200 ${
            uiConfig.compactMode ? 'min-h-10' : 'min-h-12'
          } ${
            isActive 
              ? 'bg-blue-50 border-r-4 border-blue-600 text-blue-700' 
              : 'hover:bg-gray-50 text-gray-700'
          }`}
          style={{ paddingLeft: `${level * 16 + 16}px` }}
        >
          {renderIcon(item)}
          {renderText(item)}
          
          {hasChildren && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleExpandClick(item.id);
              }}
              className="p-1 rounded hover:bg-gray-200 transition-colors"
            >
              {isExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </button>
          )}
        </button>
      </div>
    );

    // 如果是紧凑模式且不显示标签，使用title属性作为tooltip
    if (uiConfig.compactMode && !uiConfig.showLabels) {
      return (
        <div key={item.id} title={item.title}>
          {menuContent}
          {hasChildren && isExpanded && (
            <div className="bg-gray-50">
              {item.children!.map(child => renderMenuItem(child, level + 1))}
            </div>
          )}
        </div>
      );
    }

    return (
      <div key={item.id}>
        {menuContent}
        
        {hasChildren && isExpanded && (
          <div className="bg-gray-50">
            {item.children!.map(child => renderMenuItem(child, level + 1))}
          </div>
        )}
      </div>
    );
  };

  // 侧边栏内容
  const sidebarContent = (
    <div 
      className="h-full bg-white border-r border-gray-200 flex flex-col"
      style={{ width: uiConfig.sidebarWidth || 280 }}
    >
      {/* 头部 */}
      <div className="p-4 flex items-center justify-between border-b border-gray-200">
        {!uiConfig.compactMode && (
          <h2 className="text-lg font-semibold text-gray-800 truncate">
            动态菜单
          </h2>
        )}
        
        {(isMobile || variant === 'temporary') && (
          <button 
            onClick={onToggle}
            className="p-1 rounded hover:bg-gray-100 transition-colors"
          >
            <MenuIcon className="h-5 w-5" />
          </button>
        )}
      </div>

      {/* 菜单列表 */}
      <div className="flex-1 overflow-auto">
        {isLoading ? (
          <div className="p-4 text-center">
            <div className="text-sm text-gray-500">
              加载菜单中...
            </div>
          </div>
        ) : error ? (
          <div className="p-4 text-center">
            <div className="text-sm text-red-500">
              {error}
            </div>
          </div>
        ) : (
          <div>
            {menuItems.map(item => renderMenuItem(item))}
          </div>
        )}
      </div>

      {/* 底部信息 */}
      {!uiConfig.compactMode && (
        <div className="p-4 border-t border-gray-200 text-center">
          <div className="text-xs text-gray-500">
            设备: {deviceInfo.type}
          </div>
        </div>
      )}
    </div>
  );

  // 根据设备类型和配置决定侧边栏样式
  if (deviceInfo.type === DeviceType.MOBILE || deviceInfo.type === DeviceType.WATCH) {
    return (
      <>
        {/* 移动端遮罩层 */}
        {open && (
          <div 
            className="fixed inset-0 bg-black bg-opacity-50 z-40"
            onClick={onToggle}
          />
        )}
        
        {/* 移动端侧边栏 */}
        <div 
          className={`fixed top-0 left-0 h-full z-50 transform transition-transform duration-300 ease-in-out ${
            open ? 'translate-x-0' : '-translate-x-full'
          }`}
        >
          {sidebarContent}
        </div>
      </>
    );
  }

  // 桌面端侧边栏
  if (variant === 'temporary') {
    return (
      <>
        {/* 桌面端遮罩层 */}
        {open && (
          <div 
            className="fixed inset-0 bg-black bg-opacity-50 z-40"
            onClick={onToggle}
          />
        )}
        
        {/* 桌面端临时侧边栏 */}
        <div 
          className={`fixed top-0 left-0 h-full z-50 transform transition-transform duration-300 ease-in-out ${
            open ? 'translate-x-0' : '-translate-x-full'
          }`}
        >
          {sidebarContent}
        </div>
      </>
    );
  }

  // 持久化侧边栏
  return (
    <div 
      className={`h-full transition-all duration-300 ease-in-out ${
        open ? 'w-auto' : 'w-0 overflow-hidden'
      }`}
    >
      {sidebarContent}
    </div>
  );
};