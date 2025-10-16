import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { 
  MenuItem, 
  UserMenuConfig, 
  MenuUpdateRequest, 
  PermissionLevel 
} from '../types/menu';
import { menuService } from '../services/menuService';
import { getDeviceInfo } from '../utils/deviceDetection';
import { useAuth } from './AuthContext';

interface DynamicMenuContextType {
  // 状态
  menuItems: MenuItem[];
  userConfig: UserMenuConfig | null;
  deviceInfo: ReturnType<typeof getDeviceInfo>;
  isLoading: boolean;
  error: string | null;

  // 操作方法
  refreshMenu: () => Promise<void>;
  updateUserMenu: (request: MenuUpdateRequest) => Promise<boolean>;
  hasMenuPermission: (menuId: string) => boolean;
  getFilteredMenuItems: () => MenuItem[];
  
  // 管理员操作
  grantPermission: (targetUserId: string, permissions: PermissionLevel[]) => Promise<boolean>;
  revokePermission: (targetUserId: string, permissions: PermissionLevel[]) => Promise<boolean>;
  updateMenuVisibility: (targetUserId: string, menuItemIds: string[], visible: boolean) => Promise<boolean>;
}

const DynamicMenuContext = createContext<DynamicMenuContextType | undefined>(undefined);

interface DynamicMenuProviderProps {
  children: React.ReactNode;
}

export const DynamicMenuProvider: React.FC<DynamicMenuProviderProps> = ({ children }) => {
  const { user, isAuthenticated } = useAuth();
  const [menuItems, setMenuItems] = useState<MenuItem[]>([]);
  const [userConfig, setUserConfig] = useState<UserMenuConfig | null>(null);
  const [deviceInfo] = useState(() => getDeviceInfo());
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 刷新菜单
  const refreshMenu = useCallback(async () => {
    if (!isAuthenticated || !user) {
      setMenuItems([]);
      setUserConfig(null);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await menuService.getUserMenu(user.id);
      
      if (response.success) {
        setMenuItems(response.data.menuItems);
        setUserConfig(response.data.userConfig);
      } else {
        setError(response.message || '获取菜单失败');
      }
    } catch (err) {
      setError('网络错误，无法获取菜单');
      console.error('Menu refresh error:', err);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, user]);

  // 更新用户菜单
  const updateUserMenu = useCallback(async (request: MenuUpdateRequest): Promise<boolean> => {
    try {
      const success = await menuService.updateUserMenu(request);
      
      if (success) {
        await refreshMenu(); // 刷新菜单以获取最新状态
      }
      
      return success;
    } catch (err) {
      console.error('Update user menu error:', err);
      return false;
    }
  }, [refreshMenu]);

  // 检查菜单权限
  const hasMenuPermission = useCallback((menuId: string): boolean => {
    if (!userConfig || !menuItems.length) {
      return false;
    }

    const menuItem = findMenuItemById(menuItems, menuId);
    if (!menuItem) {
      return false;
    }

    return menuService.filterMenuItems(
      [menuItem], 
      userConfig.userPermissions, 
      deviceInfo.type
    ).length > 0;
  }, [menuItems, userConfig, deviceInfo.type]);

  // 获取过滤后的菜单项
  const getFilteredMenuItems = useCallback((): MenuItem[] => {
    if (!userConfig || !menuItems.length) {
      return [];
    }

    let filteredItems = menuService.filterMenuItems(
      menuItems, 
      userConfig.userPermissions, 
      deviceInfo.type
    );

    // 应用用户自定义隐藏设置
    if (userConfig.hiddenMenuItems?.length) {
      filteredItems = filteredItems.filter(item => 
        !userConfig.hiddenMenuItems!.includes(item.id)
      );
    }

    // 应用用户自定义排序
    if (userConfig.menuOrder?.length) {
      filteredItems.sort((a, b) => {
        const aIndex = userConfig.menuOrder!.indexOf(a.id);
        const bIndex = userConfig.menuOrder!.indexOf(b.id);
        
        if (aIndex === -1 && bIndex === -1) {
          return a.order - b.order;
        }
        if (aIndex === -1) return 1;
        if (bIndex === -1) return -1;
        
        return aIndex - bIndex;
      });
    }

    return filteredItems;
  }, [menuItems, userConfig, deviceInfo.type]);

  // 管理员操作：授予权限
  const grantPermission = useCallback(async (
    targetUserId: string, 
    permissions: PermissionLevel[]
  ): Promise<boolean> => {
    if (!user || !userConfig?.userPermissions.includes(PermissionLevel.ADMIN)) {
      return false;
    }

    const request: MenuUpdateRequest = {
      userId: user.id,
      adminAction: {
        type: 'grant_permission',
        targetUserId,
        permissions
      }
    };

    return await updateUserMenu(request);
  }, [user, userConfig, updateUserMenu]);

  // 管理员操作：撤销权限
  const revokePermission = useCallback(async (
    targetUserId: string, 
    permissions: PermissionLevel[]
  ): Promise<boolean> => {
    if (!user || !userConfig?.userPermissions.includes(PermissionLevel.ADMIN)) {
      return false;
    }

    const request: MenuUpdateRequest = {
      userId: user.id,
      adminAction: {
        type: 'revoke_permission',
        targetUserId,
        permissions
      }
    };

    return await updateUserMenu(request);
  }, [user, userConfig, updateUserMenu]);

  // 管理员操作：更新菜单可见性
  const updateMenuVisibility = useCallback(async (
    targetUserId: string, 
    menuItemIds: string[], 
    visible: boolean
  ): Promise<boolean> => {
    if (!user || !userConfig?.userPermissions.includes(PermissionLevel.ADMIN)) {
      return false;
    }

    // 根据visible参数决定是显示还是隐藏菜单项
    const hiddenMenuItems = visible 
      ? userConfig?.hiddenMenuItems?.filter(id => !menuItemIds.includes(id)) || []
      : [...(userConfig?.hiddenMenuItems || []), ...menuItemIds];

    const request: MenuUpdateRequest = {
      userId: user.id,
      userConfig: {
        hiddenMenuItems
      },
      adminAction: {
        type: 'update_menu',
        targetUserId,
        menuItemIds
      }
    };

    return await updateUserMenu(request);
  }, [user, userConfig, updateUserMenu]);

  // 初始化和用户变化时刷新菜单
  useEffect(() => {
    refreshMenu();
  }, [refreshMenu]);

  // 监听设备方向变化
  useEffect(() => {
    const handleResize = () => {
      // 设备信息变化时可能需要重新适配菜单
      if (userConfig) {
        const newDeviceInfo = getDeviceInfo();
        if (newDeviceInfo.type !== deviceInfo.type) {
          refreshMenu();
        }
      }
    };

    window.addEventListener('resize', handleResize);
    window.addEventListener('orientationchange', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('orientationchange', handleResize);
    };
  }, [userConfig, deviceInfo.type, refreshMenu]);

  const contextValue: DynamicMenuContextType = {
    menuItems,
    userConfig,
    deviceInfo,
    isLoading,
    error,
    refreshMenu,
    updateUserMenu,
    hasMenuPermission,
    getFilteredMenuItems,
    grantPermission,
    revokePermission,
    updateMenuVisibility
  };

  return (
    <DynamicMenuContext.Provider value={contextValue}>
      {children}
    </DynamicMenuContext.Provider>
  );
};

// Hook for using the dynamic menu context
export const useDynamicMenu = (): DynamicMenuContextType => {
  const context = useContext(DynamicMenuContext);
  if (context === undefined) {
    throw new Error('useDynamicMenu must be used within a DynamicMenuProvider');
  }
  return context;
};

// 辅助函数：在菜单树中查找菜单项
function findMenuItemById(menuItems: MenuItem[], id: string): MenuItem | null {
  for (const item of menuItems) {
    if (item.id === id) {
      return item;
    }
    
    if (item.children) {
      const found = findMenuItemById(item.children, id);
      if (found) {
        return found;
      }
    }
  }
  
  return null;
}