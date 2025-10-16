import { configureStore, createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';
import { persistStore, persistReducer } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import { combineReducers } from '@reduxjs/toolkit';

// 用户信息类型
export interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  role: 'user' | 'admin' | 'moderator';
  permissions: string[];
  profile?: {
    nickname?: string;
    bio?: string;
    location?: string;
    website?: string;
  };
  preferences?: {
    theme: 'light' | 'dark' | 'auto';
    language: 'zh-CN' | 'en-US';
    notifications: boolean;
  };
}

// 通知类型
export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: number;
  read: boolean;
  actions?: Array<{
    label: string;
    action: string;
  }>;
}

// 应用状态接口
export interface AppState {
  loading: boolean;
  theme: 'light' | 'dark' | 'auto';
  language: 'zh-CN' | 'en-US';
  user: User | null;
  isAuthenticated: boolean;
  sidebarCollapsed: boolean;
  notifications: Notification[];
  unreadCount: number;
  globalSearch: {
    visible: boolean;
    query: string;
    results: any[];
    loading: boolean;
  };
  breadcrumbs: Array<{
    title: string;
    path?: string;
  }>;
  pageTitle: string;
  errors: Array<{
    id: string;
    message: string;
    timestamp: number;
    stack?: string;
  }>;
}

// 初始状态
const initialState: AppState = {
  loading: false,
  theme: 'light',
  language: 'zh-CN',
  user: null,
  isAuthenticated: false,
  sidebarCollapsed: false,
  notifications: [],
  unreadCount: 0,
  globalSearch: {
    visible: false,
    query: '',
    results: [],
    loading: false
  },
  breadcrumbs: [],
  pageTitle: '太上老君AI平台',
  errors: []
};

// 应用切片
const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    // 加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    
    // 主题设置
    setTheme: (state, action: PayloadAction<'light' | 'dark' | 'auto'>) => {
      state.theme = action.payload;
      if (state.user) {
        state.user.preferences = {
          ...state.user.preferences,
          theme: action.payload
        };
      }
    },
    
    // 语言设置
    setLanguage: (state, action: PayloadAction<'zh-CN' | 'en-US'>) => {
      state.language = action.payload;
      if (state.user) {
        state.user.preferences = {
          ...state.user.preferences,
          language: action.payload
        };
      }
    },
    
    // 用户信息
    setUser: (state, action: PayloadAction<User | null>) => {
      state.user = action.payload;
      state.isAuthenticated = !!action.payload;
      
      // 同步用户偏好设置
      if (action.payload?.preferences) {
        state.theme = action.payload.preferences.theme || 'light';
        state.language = action.payload.preferences.language || 'zh-CN';
      }
    },
    
    // 更新用户信息
    updateUser: (state, action: PayloadAction<Partial<User>>) => {
      if (state.user) {
        state.user = { ...state.user, ...action.payload };
      }
    },
    
    // 侧边栏折叠
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebarCollapsed = action.payload;
    },
    
    // 通知管理
    addNotification: (state, action: PayloadAction<Omit<Notification, 'id' | 'timestamp'>>) => {
      const notification: Notification = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: Date.now(),
        read: false
      };
      state.notifications.unshift(notification);
      state.unreadCount += 1;
    },
    
    markNotificationRead: (state, action: PayloadAction<string>) => {
      const notification = state.notifications.find(n => n.id === action.payload);
      if (notification && !notification.read) {
        notification.read = true;
        state.unreadCount = Math.max(0, state.unreadCount - 1);
      }
    },
    
    markAllNotificationsRead: (state) => {
      state.notifications.forEach(n => n.read = true);
      state.unreadCount = 0;
    },
    
    removeNotification: (state, action: PayloadAction<string>) => {
      const index = state.notifications.findIndex(n => n.id === action.payload);
      if (index !== -1) {
        const notification = state.notifications[index];
        if (!notification.read) {
          state.unreadCount = Math.max(0, state.unreadCount - 1);
        }
        state.notifications.splice(index, 1);
      }
    },
    
    clearNotifications: (state) => {
      state.notifications = [];
      state.unreadCount = 0;
    },
    
    // 全局搜索
    setGlobalSearchVisible: (state, action: PayloadAction<boolean>) => {
      state.globalSearch.visible = action.payload;
      if (!action.payload) {
        state.globalSearch.query = '';
        state.globalSearch.results = [];
      }
    },
    
    setGlobalSearchQuery: (state, action: PayloadAction<string>) => {
      state.globalSearch.query = action.payload;
    },
    
    setGlobalSearchResults: (state, action: PayloadAction<any[]>) => {
      state.globalSearch.results = action.payload;
    },
    
    setGlobalSearchLoading: (state, action: PayloadAction<boolean>) => {
      state.globalSearch.loading = action.payload;
    },
    
    // 面包屑导航
    setBreadcrumbs: (state, action: PayloadAction<Array<{ title: string; path?: string }>>) => {
      state.breadcrumbs = action.payload;
    },
    
    // 页面标题
    setPageTitle: (state, action: PayloadAction<string>) => {
      state.pageTitle = action.payload;
      document.title = `${action.payload} - 太上老君AI平台`;
    },
    
    // 错误管理
    addError: (state, action: PayloadAction<{ message: string; stack?: string }>) => {
      const error = {
        id: Date.now().toString(),
        message: action.payload.message,
        stack: action.payload.stack,
        timestamp: Date.now()
      };
      state.errors.unshift(error);
      
      // 只保留最近的10个错误
      if (state.errors.length > 10) {
        state.errors = state.errors.slice(0, 10);
      }
    },
    
    removeError: (state, action: PayloadAction<string>) => {
      state.errors = state.errors.filter(error => error.id !== action.payload);
    },
    
    clearErrors: (state) => {
      state.errors = [];
    },
    
    // 重置状态（用于登出）
    resetState: () => {
      return {
        ...initialState,
        theme: initialState.theme, // 保留主题设置
        language: initialState.language // 保留语言设置
      };
    }
  }
});

// 导出actions
export const {
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
  resetState
} = appSlice.actions;

// 持久化配置
const persistConfig = {
  key: 'taishanglaojun-app',
  storage,
  whitelist: ['theme', 'language', 'user', 'isAuthenticated', 'sidebarCollapsed'] // 只持久化这些字段
};

// 根reducer
const rootReducer = combineReducers({
  app: persistReducer(persistConfig, appSlice.reducer)
});

// 配置store
export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE']
      }
    }),
  devTools: process.env.NODE_ENV !== 'production'
});

// 持久化store
export const persistor = persistStore(store);

// 类型定义
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// 选择器
export const selectApp = (state: RootState) => state.app;
export const selectUser = (state: RootState) => state.app.user;
export const selectIsAuthenticated = (state: RootState) => state.app.isAuthenticated;
export const selectTheme = (state: RootState) => state.app.theme;
export const selectLanguage = (state: RootState) => state.app.language;
export const selectLoading = (state: RootState) => state.app.loading;
export const selectNotifications = (state: RootState) => state.app.notifications;
export const selectUnreadCount = (state: RootState) => state.app.unreadCount;
export const selectGlobalSearch = (state: RootState) => state.app.globalSearch;
export const selectBreadcrumbs = (state: RootState) => state.app.breadcrumbs;
export const selectPageTitle = (state: RootState) => state.app.pageTitle;
export const selectErrors = (state: RootState) => state.app.errors;

export default store;