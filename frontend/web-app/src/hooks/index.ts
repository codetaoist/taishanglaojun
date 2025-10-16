import { useDispatch, useSelector } from 'react-redux';
import { useCallback, useEffect, useState } from 'react';
import type { TypedUseSelectorHook } from 'react-redux';
import type { RootState, AppDispatch } from '../store';
import { apiManager } from '../services/apiManager';
import {
  setLoading,
  setTheme,
  setLanguage,
  setUser,
  updateUser,
  setSidebarCollapsed,
  addNotification,
  markNotificationRead,
  markAllNotificationsRead,
  removeNotification,
  clearNotifications,
  setGlobalSearchVisible,
  setGlobalSearchQuery,
  setGlobalSearchResults,
  setGlobalSearchLoading,
  setBreadcrumbs,
  setPageTitle,
  addError,
  removeError,
  clearErrors,
  resetState,
  selectApp,
  selectUser,
  selectIsAuthenticated,
  selectTheme,
  selectLanguage,
  selectLoading,
  selectNotifications,
  selectUnreadCount,
  selectGlobalSearch,
  selectBreadcrumbs,
  selectPageTitle,
  selectErrors
} from '../store';

// 类型化的hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

// 应用状态hooks
export const useApp = () => {
  const dispatch = useAppDispatch();
  const app = useAppSelector(selectApp);

  return {
    ...app,
    setLoading: useCallback((loading: boolean) => dispatch(setLoading(loading)), [dispatch]),
    setTheme: useCallback((theme: 'light' | 'dark' | 'auto') => dispatch(setTheme(theme)), [dispatch]),
    setLanguage: useCallback((language: 'zh-CN' | 'en-US') => dispatch(setLanguage(language)), [dispatch]),
    setSidebarCollapsed: useCallback((collapsed: boolean) => dispatch(setSidebarCollapsed(collapsed)), [dispatch]),
    setPageTitle: useCallback((title: string) => dispatch(setPageTitle(title)), [dispatch]),
    setBreadcrumbs: useCallback((breadcrumbs: Array<{ title: string; path?: string }>) => 
      dispatch(setBreadcrumbs(breadcrumbs)), [dispatch]),
    resetState: useCallback(() => dispatch(resetState()), [dispatch])
  };
};

// 用户状态hooks
export const useAuth = () => {
  const dispatch = useAppDispatch();
  const user = useAppSelector(selectUser);
  const isAuthenticated = useAppSelector(selectIsAuthenticated);

  const login = useCallback(async (email: string, password: string) => {
    try {
      const result = await apiManager.login(email, password);
      dispatch(setUser(result.user));
      localStorage.setItem('token', result.token);
      return result;
    } catch (error) {
      throw error;
    }
  }, [dispatch]);

  const logout = useCallback(async () => {
    try {
      await apiManager.logout({ silent: true });
    } catch (error) {
      console.warn('Logout API call failed:', error);
    } finally {
      dispatch(resetState());
      localStorage.removeItem('token');
    }
  }, [dispatch]);

  const updateProfile = useCallback(async (profileData: any) => {
    try {
      const updatedUser = await apiManager.updateProfile(profileData);
      dispatch(updateUser(updatedUser));
      return updatedUser;
    } catch (error) {
      throw error;
    }
  }, [dispatch]);

  const getCurrentUser = useCallback(async () => {
    try {
      const userData = await apiManager.getCurrentUser({ silent: true });
      dispatch(setUser(userData));
      return userData;
    } catch (error) {
      // 如果获取用户信息失败，清除认证状态
      dispatch(resetState());
      localStorage.removeItem('token');
      throw error;
    }
  }, [dispatch]);

  return {
    user,
    isAuthenticated,
    login,
    logout,
    updateProfile,
    getCurrentUser,
    setUser: useCallback((user: any) => dispatch(setUser(user)), [dispatch]),
    updateUser: useCallback((userData: any) => dispatch(updateUser(userData)), [dispatch])
  };
};

// 通知系统hooks
export const useNotifications = () => {
  const dispatch = useAppDispatch();
  const notifications = useAppSelector(selectNotifications);
  const unreadCount = useAppSelector(selectUnreadCount);

  return {
    notifications,
    unreadCount,
    addNotification: useCallback((notification: any) => dispatch(addNotification(notification)), [dispatch]),
    markRead: useCallback((id: string) => dispatch(markNotificationRead(id)), [dispatch]),
    markAllRead: useCallback(() => dispatch(markAllNotificationsRead()), [dispatch]),
    remove: useCallback((id: string) => dispatch(removeNotification(id)), [dispatch]),
    clear: useCallback(() => dispatch(clearNotifications()), [dispatch])
  };
};

// 全局搜索hooks
export const useGlobalSearch = () => {
  const dispatch = useAppDispatch();
  const globalSearch = useAppSelector(selectGlobalSearch);

  const search = useCallback(async (query: string) => {
    if (!query.trim()) {
      dispatch(setGlobalSearchResults([]));
      return;
    }

    dispatch(setGlobalSearchLoading(true));
    try {
      // 这里可以调用实际的搜索API
      const results = await apiManager.searchWisdom(query, {}, { silent: true });
      dispatch(setGlobalSearchResults(results.items || []));
    } catch (error) {
      dispatch(setGlobalSearchResults([]));
    } finally {
      dispatch(setGlobalSearchLoading(false));
    }
  }, [dispatch]);

  return {
    ...globalSearch,
    setVisible: useCallback((visible: boolean) => dispatch(setGlobalSearchVisible(visible)), [dispatch]),
    setQuery: useCallback((query: string) => dispatch(setGlobalSearchQuery(query)), [dispatch]),
    search
  };
};

// 错误处理hooks
export const useErrors = () => {
  const dispatch = useAppDispatch();
  const errors = useAppSelector(selectErrors);

  return {
    errors,
    addError: useCallback((error: { message: string; stack?: string }) => 
      dispatch(addError(error)), [dispatch]),
    removeError: useCallback((id: string) => dispatch(removeError(id)), [dispatch]),
    clearErrors: useCallback(() => dispatch(clearErrors()), [dispatch])
  };
};

// API加载状态hooks
export const useApiLoading = () => {
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const handleLoadingChange = (event: CustomEvent) => {
      setLoading(event.detail.loading);
    };

    window.addEventListener('api-loading-change', handleLoadingChange as EventListener);
    return () => {
      window.removeEventListener('api-loading-change', handleLoadingChange as EventListener);
    };
  }, []);

  return loading;
};

// 主题hooks
export const useTheme = () => {
  const theme = useAppSelector(selectTheme);
  const dispatch = useAppDispatch();

  const setTheme = useCallback((newTheme: 'light' | 'dark' | 'auto') => {
    dispatch(setTheme(newTheme));
  }, [dispatch]);

  const toggleTheme = useCallback(() => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  }, [theme, setTheme]);

  return {
    theme,
    setTheme,
    toggleTheme,
    isDark: theme === 'dark'
  };
};

// 语言hooks
export const useLanguage = () => {
  const language = useAppSelector(selectLanguage);
  const dispatch = useAppDispatch();

  const setLanguage = useCallback((newLanguage: 'zh-CN' | 'en-US') => {
    dispatch(setLanguage(newLanguage));
  }, [dispatch]);

  const toggleLanguage = useCallback(() => {
    const newLanguage = language === 'zh-CN' ? 'en-US' : 'zh-CN';
    setLanguage(newLanguage);
  }, [language, setLanguage]);

  return {
    language,
    setLanguage,
    toggleLanguage,
    isEnglish: language === 'en-US'
  };
};

// 面包屑导航hooks
export const useBreadcrumbs = () => {
  const breadcrumbs = useAppSelector(selectBreadcrumbs);
  const dispatch = useAppDispatch();

  const setBreadcrumbs = useCallback((breadcrumbs: Array<{ title: string; path?: string }>) => {
    dispatch(setBreadcrumbs(breadcrumbs));
  }, [dispatch]);

  const addBreadcrumb = useCallback((breadcrumb: { title: string; path?: string }) => {
    dispatch(setBreadcrumbs([...breadcrumbs, breadcrumb]));
  }, [breadcrumbs, dispatch]);

  return {
    breadcrumbs,
    setBreadcrumbs,
    addBreadcrumb
  };
};

// 页面标题hooks
export const usePageTitle = () => {
  const pageTitle = useAppSelector(selectPageTitle);
  const dispatch = useAppDispatch();

  const setPageTitle = useCallback((title: string) => {
    dispatch(setPageTitle(title));
  }, [dispatch]);

  return {
    pageTitle,
    setPageTitle
  };
};

// 本地存储hooks
export const useLocalStorage = <T>(key: string, initialValue: T) => {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error);
      return initialValue;
    }
  });

  const setValue = useCallback((value: T | ((val: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      window.localStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.warn(`Error setting localStorage key "${key}":`, error);
    }
  }, [key, storedValue]);

  return [storedValue, setValue] as const;
};

// 防抖hooks
export const useDebounce = <T>(value: T, delay: number) => {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
};

// 窗口大小hooks
export const useWindowSize = () => {
  const [windowSize, setWindowSize] = useState({
    width: window.innerWidth,
    height: window.innerHeight,
  });

  useEffect(() => {
    const handleResize = () => {
      setWindowSize({
        width: window.innerWidth,
        height: window.innerHeight,
      });
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return windowSize;
};

// 响应式断点hooks
export const useBreakpoint = () => {
  const { width } = useWindowSize();

  return {
    isMobile: width < 768,
    isTablet: width >= 768 && width < 1024,
    isDesktop: width >= 1024,
    isLarge: width >= 1200,
    isXLarge: width >= 1600,
    width
  };
};