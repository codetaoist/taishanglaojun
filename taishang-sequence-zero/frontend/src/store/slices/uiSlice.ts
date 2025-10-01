import { createSlice, PayloadAction } from '@reduxjs/toolkit';

// 主题类型
export type ThemeType = 'light' | 'dark' | 'auto';

// 语言类型
export type LanguageType = 'zh-CN' | 'en-US';

// 布局类型
export type LayoutType = 'side' | 'top' | 'mix';

// 侧边栏状态
export interface SidebarState {
  collapsed: boolean;
  width: number;
  collapsedWidth: number;
}

// 面包屑项
export interface BreadcrumbItem {
  title: string;
  path?: string;
  icon?: string;
}

// 标签页项
export interface TabItem {
  key: string;
  title: string;
  path: string;
  closable?: boolean;
  icon?: string;
}

// 通知项
export interface NotificationItem {
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

// UI状态接口
export interface UIState {
  // 主题设置
  theme: ThemeType;
  primaryColor: string;
  
  // 语言设置
  language: LanguageType;
  
  // 布局设置
  layout: LayoutType;
  sidebar: SidebarState;
  headerFixed: boolean;
  footerFixed: boolean;
  
  // 页面状态
  loading: boolean;
  pageLoading: boolean;
  
  // 导航状态
  breadcrumbs: BreadcrumbItem[];
  tabs: TabItem[];
  activeTab: string;
  
  // 通知状态
  notifications: NotificationItem[];
  unreadCount: number;
  
  // 模态框状态
  modals: Record<string, boolean>;
  
  // 抽屉状态
  drawers: Record<string, boolean>;
  
  // 全屏状态
  fullscreen: boolean;
  
  // 设备信息
  isMobile: boolean;
  screenSize: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xxl';
  
  // 页面设置
  pageSettings: {
    showBreadcrumb: boolean;
    showTabs: boolean;
    showFooter: boolean;
    contentPadding: number;
    animationEnabled: boolean;
  };
  
  // 用户偏好
  preferences: {
    autoSave: boolean;
    soundEnabled: boolean;
    notificationEnabled: boolean;
    compactMode: boolean;
  };
}

// 初始状态
const initialState: UIState = {
  // 主题设置
  theme: 'light',
  primaryColor: '#d4af37', // 太上老君金色
  
  // 语言设置
  language: 'zh-CN',
  
  // 布局设置
  layout: 'side',
  sidebar: {
    collapsed: false,
    width: 256,
    collapsedWidth: 80,
  },
  headerFixed: true,
  footerFixed: false,
  
  // 页面状态
  loading: false,
  pageLoading: false,
  
  // 导航状态
  breadcrumbs: [],
  tabs: [],
  activeTab: '',
  
  // 通知状态
  notifications: [],
  unreadCount: 0,
  
  // 模态框状态
  modals: {},
  
  // 抽屉状态
  drawers: {},
  
  // 全屏状态
  fullscreen: false,
  
  // 设备信息
  isMobile: false,
  screenSize: 'lg',
  
  // 页面设置
  pageSettings: {
    showBreadcrumb: true,
    showTabs: true,
    showFooter: true,
    contentPadding: 24,
    animationEnabled: true,
  },
  
  // 用户偏好
  preferences: {
    autoSave: true,
    soundEnabled: true,
    notificationEnabled: true,
    compactMode: false,
  },
};

// 创建UI slice
const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    // 主题设置
    setTheme: (state, action: PayloadAction<ThemeType>) => {
      state.theme = action.payload;
    },
    
    setPrimaryColor: (state, action: PayloadAction<string>) => {
      state.primaryColor = action.payload;
    },
    
    toggleTheme: (state) => {
      state.theme = state.theme === 'light' ? 'dark' : 'light';
    },
    
    // 语言设置
    setLanguage: (state, action: PayloadAction<LanguageType>) => {
      state.language = action.payload;
    },
    
    // 布局设置
    setLayout: (state, action: PayloadAction<LayoutType>) => {
      state.layout = action.payload;
    },
    
    // 侧边栏控制
    toggleSidebar: (state) => {
      state.sidebar.collapsed = !state.sidebar.collapsed;
    },
    
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebar.collapsed = action.payload;
    },
    
    setSidebarWidth: (state, action: PayloadAction<number>) => {
      state.sidebar.width = action.payload;
    },
    
    // 页面加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    
    setPageLoading: (state, action: PayloadAction<boolean>) => {
      state.pageLoading = action.payload;
    },
    
    // 面包屑管理
    setBreadcrumbs: (state, action: PayloadAction<BreadcrumbItem[]>) => {
      state.breadcrumbs = action.payload;
    },
    
    addBreadcrumb: (state, action: PayloadAction<BreadcrumbItem>) => {
      state.breadcrumbs.push(action.payload);
    },
    
    clearBreadcrumbs: (state) => {
      state.breadcrumbs = [];
    },
    
    // 标签页管理
    addTab: (state, action: PayloadAction<TabItem>) => {
      const existingTab = state.tabs.find(tab => tab.key === action.payload.key);
      if (!existingTab) {
        state.tabs.push(action.payload);
      }
      state.activeTab = action.payload.key;
    },
    
    removeTab: (state, action: PayloadAction<string>) => {
      const index = state.tabs.findIndex(tab => tab.key === action.payload);
      if (index > -1) {
        state.tabs.splice(index, 1);
        // 如果删除的是当前活跃标签，切换到前一个标签
        if (state.activeTab === action.payload && state.tabs.length > 0) {
          const newIndex = Math.max(0, index - 1);
          state.activeTab = state.tabs[newIndex]?.key || '';
        }
      }
    },
    
    setActiveTab: (state, action: PayloadAction<string>) => {
      state.activeTab = action.payload;
    },
    
    clearTabs: (state) => {
      state.tabs = [];
      state.activeTab = '';
    },
    
    // 通知管理
    addNotification: (state, action: PayloadAction<Omit<NotificationItem, 'id' | 'timestamp'>>) => {
      const notification: NotificationItem = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: Date.now(),
      };
      state.notifications.unshift(notification);
      if (!notification.read) {
        state.unreadCount += 1;
      }
    },
    
    markNotificationAsRead: (state, action: PayloadAction<string>) => {
      const notification = state.notifications.find(n => n.id === action.payload);
      if (notification && !notification.read) {
        notification.read = true;
        state.unreadCount = Math.max(0, state.unreadCount - 1);
      }
    },
    
    markAllNotificationsAsRead: (state) => {
      state.notifications.forEach(notification => {
        notification.read = true;
      });
      state.unreadCount = 0;
    },
    
    removeNotification: (state, action: PayloadAction<string>) => {
      const index = state.notifications.findIndex(n => n.id === action.payload);
      if (index > -1) {
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
    
    // 模态框管理
    openModal: (state, action: PayloadAction<string>) => {
      state.modals[action.payload] = true;
    },
    
    closeModal: (state, action: PayloadAction<string>) => {
      state.modals[action.payload] = false;
    },
    
    toggleModal: (state, action: PayloadAction<string>) => {
      state.modals[action.payload] = !state.modals[action.payload];
    },
    
    // 抽屉管理
    openDrawer: (state, action: PayloadAction<string>) => {
      state.drawers[action.payload] = true;
    },
    
    closeDrawer: (state, action: PayloadAction<string>) => {
      state.drawers[action.payload] = false;
    },
    
    toggleDrawer: (state, action: PayloadAction<string>) => {
      state.drawers[action.payload] = !state.drawers[action.payload];
    },
    
    // 全屏控制
    setFullscreen: (state, action: PayloadAction<boolean>) => {
      state.fullscreen = action.payload;
    },
    
    toggleFullscreen: (state) => {
      state.fullscreen = !state.fullscreen;
    },
    
    // 设备信息
    setIsMobile: (state, action: PayloadAction<boolean>) => {
      state.isMobile = action.payload;
    },
    
    setScreenSize: (state, action: PayloadAction<UIState['screenSize']>) => {
      state.screenSize = action.payload;
    },
    
    // 页面设置
    updatePageSettings: (state, action: PayloadAction<Partial<UIState['pageSettings']>>) => {
      state.pageSettings = { ...state.pageSettings, ...action.payload };
    },
    
    // 用户偏好
    updatePreferences: (state, action: PayloadAction<Partial<UIState['preferences']>>) => {
      state.preferences = { ...state.preferences, ...action.payload };
    },
    
    // 重置UI状态
    resetUIState: () => initialState,
  },
});

// 导出actions
export const {
  setTheme,
  setPrimaryColor,
  toggleTheme,
  setLanguage,
  setLayout,
  toggleSidebar,
  setSidebarCollapsed,
  setSidebarWidth,
  setLoading,
  setPageLoading,
  setBreadcrumbs,
  addBreadcrumb,
  clearBreadcrumbs,
  addTab,
  removeTab,
  setActiveTab,
  clearTabs,
  addNotification,
  markNotificationAsRead,
  markAllNotificationsAsRead,
  removeNotification,
  clearNotifications,
  openModal,
  closeModal,
  toggleModal,
  openDrawer,
  closeDrawer,
  toggleDrawer,
  setFullscreen,
  toggleFullscreen,
  setIsMobile,
  setScreenSize,
  updatePageSettings,
  updatePreferences,
  resetUIState,
} = uiSlice.actions;

// 选择器
export const selectUI = (state: { ui: UIState }) => state.ui;
export const selectTheme = (state: { ui: UIState }) => state.ui.theme;
export const selectSidebar = (state: { ui: UIState }) => state.ui.sidebar;
export const selectBreadcrumbs = (state: { ui: UIState }) => state.ui.breadcrumbs;
export const selectTabs = (state: { ui: UIState }) => state.ui.tabs;
export const selectNotifications = (state: { ui: UIState }) => state.ui.notifications;
export const selectUnreadCount = (state: { ui: UIState }) => state.ui.unreadCount;
export const selectPageSettings = (state: { ui: UIState }) => state.ui.pageSettings;
export const selectPreferences = (state: { ui: UIState }) => state.ui.preferences;

// 导出reducer
export default uiSlice.reducer;