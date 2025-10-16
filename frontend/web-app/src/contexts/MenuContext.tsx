import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { frontendMenuService, type UserPermissions } from '../services/frontendMenuService';
import { type MenuItem as FrontendMenuItem } from '../config/menuConfig';
import i18n from '../i18n';

interface MenuContextType {
  menuItems: FrontendMenuItem[];
  loading: boolean;
  error: string | null;
  loadMenuData: (userPermissions?: UserPermissions) => Promise<void>;
  refreshMenu: () => Promise<void>;
  clearMenu: () => void;
}

const MenuContext = createContext<MenuContextType | null>(null);

interface MenuProviderProps {
  children: React.ReactNode;
}

export const MenuProvider: React.FC<MenuProviderProps> = ({ children }) => {
  // 仅存储可序列化的菜单字段，避免 React 元素在本地存储中导致错误
  const sanitizeMenuItems = (items: any[]): any[] => {
    return (items || []).map(item => {
      const { icon, children, ...rest } = item || {};
      const sanitized: any = { ...rest };
      // 去除不可序列化的 React 元素
      sanitized.icon = undefined;
      if (Array.isArray(children) && children.length > 0) {
        sanitized.children = sanitizeMenuItems(children);
      }
      return sanitized;
    });
  };

  const [menuItems, setMenuItems] = useState<FrontendMenuItem[]>(() => {
    // 从localStorage恢复菜单状态（清理图标等不可序列化字段）
    try {
      const savedMenuItems = localStorage.getItem('menuItems');
      if (!savedMenuItems) return [];
      const parsed = JSON.parse(savedMenuItems);
      // 防御：如果解析到的 icon 是对象（可能是不合法的 React 元素快照），进行清理
      const cleaned = sanitizeMenuItems(parsed);
      // 额外清洗：若 label 看起来是键名（mainMenu.labels.*），回退为键尾段
      const normalizeLabels = (items: any[]): any[] => items.map(it => {
        const next: any = { ...it };
        const lk = next.labelKey as string | undefined;
        const lbl = next.label as string | undefined;
        const looksLikeKey = typeof lbl === 'string' && lbl.startsWith('mainMenu.labels.');
        if (lk && (looksLikeKey || !lbl)) {
          next.label = lk.startsWith('mainMenu.labels.') ? (lk.split('.').pop() || lk) : (lbl || lk);
        }
        if (Array.isArray(next.children)) {
          next.children = normalizeLabels(next.children);
        }
        return next;
      });
      return normalizeLabels(cleaned);
    } catch (error) {
      console.warn('Failed to load menu items from localStorage:', error);
      // 出错时清除损坏的缓存
      try { localStorage.removeItem('menuItems'); } catch {}
      return [];
    }
  });
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // 自动保存菜单状态到localStorage（仅保存可序列化字段）
  useEffect(() => {
    if (menuItems.length > 0) {
      try {
        const serializable = sanitizeMenuItems(menuItems as any);
        localStorage.setItem('menuItems', JSON.stringify(serializable));
      } catch (error) {
        console.warn('Failed to save menu items to localStorage:', error);
      }
    }
  }, [menuItems]);
  const [lastUserPermissions, setLastUserPermissions] = useState<UserPermissions | null>(null);

  const loadMenuData = useCallback(async (userPermissions?: UserPermissions) => {
    // 如果没有提供用户权限，且之前没有加载过，则不执行
    if (!userPermissions && !lastUserPermissions) {
      return;
    }

    const permissionsToUse = userPermissions || lastUserPermissions;
    if (!permissionsToUse) {
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      console.log('🔄 MenuContext: 开始加载菜单数据...', permissionsToUse);
      
      const menus = await frontendMenuService.getMainMenu(permissionsToUse);
      
      console.log('✅ MenuContext: 菜单数据加载成功', menus.length, '个菜单');
      
      setMenuItems(menus);
      setLastUserPermissions(permissionsToUse);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '加载菜单失败';
      console.error('❌ MenuContext: 菜单数据加载失败', errorMessage);
      setError(errorMessage);
      setMenuItems([]);
    } finally {
      setLoading(false);
    }
  }, []); // 移除lastUserPermissions依赖，避免无限循环

  const refreshMenu = useCallback(async () => {
    if (lastUserPermissions) {
      console.log('🔄 MenuContext: 刷新菜单数据...');
      // 清除缓存并重新加载
      frontendMenuService.clearCache();
      await loadMenuData(lastUserPermissions);
    }
  }, [loadMenuData]); // 只保留loadMenuData依赖

  const clearMenu = useCallback(() => {
    console.log('🗑️ MenuContext: 清除菜单数据');
    setMenuItems([]);
    setError(null);
    setLoading(false);
    setLastUserPermissions(null);
    frontendMenuService.clearCache();
    // 同时清除localStorage中的菜单数据
    try {
      localStorage.removeItem('menuItems');
    } catch (error) {
      console.warn('Failed to clear menu items from localStorage:', error);
    }
  }, []);

  // 监听语言切换，清理缓存并刷新菜单，以确保 labelKey 生效
  useEffect(() => {
    const handleLanguageChange = () => {
      try { localStorage.removeItem('menuItems'); } catch {}
      frontendMenuService.clearCache();
      // 使用最近一次权限上下文刷新菜单
      if (lastUserPermissions) {
        loadMenuData(lastUserPermissions).catch(() => {});
      } else {
        refreshMenu().catch(() => {});
      }
    };
    i18n.on('languageChanged', handleLanguageChange);
    return () => {
      i18n.off('languageChanged', handleLanguageChange);
    };
  }, [lastUserPermissions, loadMenuData, refreshMenu]);

  const contextValue: MenuContextType = {
    menuItems,
    loading,
    error,
    loadMenuData,
    refreshMenu,
    clearMenu,
  };

  return (
    <MenuContext.Provider value={contextValue}>
      {children}
    </MenuContext.Provider>
  );
};

export const useMenu = (): MenuContextType => {
  const context = useContext(MenuContext);
  if (!context) {
    throw new Error('useMenu must be used within a MenuProvider');
  }
  return context;
};

export default MenuContext;