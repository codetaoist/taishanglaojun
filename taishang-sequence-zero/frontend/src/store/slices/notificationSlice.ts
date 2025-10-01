import { createSlice, PayloadAction } from '@reduxjs/toolkit';

// 通知类型
export type NotificationType = 'info' | 'success' | 'warning' | 'error' | 'consciousness' | 'cultural' | 'system';

// 通知优先级
export type NotificationPriority = 'low' | 'medium' | 'high' | 'urgent';

// 通知状态
export type NotificationStatus = 'unread' | 'read' | 'archived';

// 通知接口
export interface Notification {
  id: string;
  type: NotificationType;
  priority: NotificationPriority;
  status: NotificationStatus;
  title: string;
  message: string;
  description?: string;
  timestamp: number;
  expiresAt?: number;
  persistent: boolean;
  dismissible: boolean;
  actionable: boolean;
  actions?: NotificationAction[];
  metadata?: Record<string, any>;
  source: string;
  userId?: string;
  groupId?: string;
  tags: string[];
}

// 通知操作接口
export interface NotificationAction {
  id: string;
  label: string;
  type: 'primary' | 'secondary' | 'danger';
  action: string;
  payload?: any;
  icon?: string;
}

// 通知设置接口
export interface NotificationSettings {
  enabled: boolean;
  sound: boolean;
  desktop: boolean;
  email: boolean;
  push: boolean;
  types: Record<NotificationType, boolean>;
  priorities: Record<NotificationPriority, boolean>;
  quietHours: {
    enabled: boolean;
    start: string;
    end: string;
  };
  maxNotifications: number;
  autoArchive: boolean;
  archiveAfterDays: number;
}

// 通知统计接口
export interface NotificationStats {
  total: number;
  unread: number;
  byType: Record<NotificationType, number>;
  byPriority: Record<NotificationPriority, number>;
  todayCount: number;
  weekCount: number;
  monthCount: number;
}

// 通知状态接口
export interface NotificationSliceState {
  // 通知列表
  notifications: Notification[];
  
  // 过滤和排序
  filters: {
    type: NotificationType | 'all';
    priority: NotificationPriority | 'all';
    status: NotificationStatus | 'all';
    source: string | 'all';
    dateRange: [number, number] | null;
  };
  
  sortBy: 'timestamp' | 'priority' | 'type';
  sortOrder: 'asc' | 'desc';
  
  // 分页
  currentPage: number;
  pageSize: number;
  
  // 选中的通知
  selectedNotifications: string[];
  
  // 通知设置
  settings: NotificationSettings;
  
  // 统计数据
  stats: NotificationStats;
  
  // UI状态
  panelVisible: boolean;
  detailVisible: boolean;
  selectedNotification: Notification | null;
  
  // 实时通知
  realtimeEnabled: boolean;
  connectionStatus: 'connected' | 'disconnected' | 'connecting';
  
  // 加载状态
  loading: boolean;
  error: string | null;
}

// 初始状态
const initialState: NotificationSliceState = {
  notifications: [],
  filters: {
    type: 'all',
    priority: 'all',
    status: 'all',
    source: 'all',
    dateRange: null,
  },
  sortBy: 'timestamp',
  sortOrder: 'desc',
  currentPage: 1,
  pageSize: 20,
  selectedNotifications: [],
  settings: {
    enabled: true,
    sound: true,
    desktop: true,
    email: false,
    push: true,
    types: {
      info: true,
      success: true,
      warning: true,
      error: true,
      consciousness: true,
      cultural: true,
      system: true,
    },
    priorities: {
      low: true,
      medium: true,
      high: true,
      urgent: true,
    },
    quietHours: {
      enabled: false,
      start: '22:00',
      end: '08:00',
    },
    maxNotifications: 100,
    autoArchive: true,
    archiveAfterDays: 30,
  },
  stats: {
    total: 0,
    unread: 0,
    byType: {
      info: 0,
      success: 0,
      warning: 0,
      error: 0,
      consciousness: 0,
      cultural: 0,
      system: 0,
    },
    byPriority: {
      low: 0,
      medium: 0,
      high: 0,
      urgent: 0,
    },
    todayCount: 0,
    weekCount: 0,
    monthCount: 0,
  },
  panelVisible: false,
  detailVisible: false,
  selectedNotification: null,
  realtimeEnabled: true,
  connectionStatus: 'disconnected',
  loading: false,
  error: null,
};

// 工具函数：生成通知ID
const generateNotificationId = (): string => {
  return `notification_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
};

// 工具函数：更新统计数据
const updateStats = (state: NotificationSliceState) => {
  const now = Date.now();
  const today = new Date().setHours(0, 0, 0, 0);
  const weekAgo = now - 7 * 24 * 60 * 60 * 1000;
  const monthAgo = now - 30 * 24 * 60 * 60 * 1000;
  
  state.stats.total = state.notifications.length;
  state.stats.unread = state.notifications.filter(n => n.status === 'unread').length;
  
  // 按类型统计
  Object.keys(state.stats.byType).forEach(type => {
    state.stats.byType[type as NotificationType] = state.notifications.filter(n => n.type === type).length;
  });
  
  // 按优先级统计
  Object.keys(state.stats.byPriority).forEach(priority => {
    state.stats.byPriority[priority as NotificationPriority] = state.notifications.filter(n => n.priority === priority).length;
  });
  
  // 时间统计
  state.stats.todayCount = state.notifications.filter(n => n.timestamp >= today).length;
  state.stats.weekCount = state.notifications.filter(n => n.timestamp >= weekAgo).length;
  state.stats.monthCount = state.notifications.filter(n => n.timestamp >= monthAgo).length;
};

// 创建通知slice
const notificationSlice = createSlice({
  name: 'notification',
  initialState,
  reducers: {
    // 添加通知
    addNotification: (state, action: PayloadAction<Omit<Notification, 'id' | 'timestamp'>>) => {
      const notification: Notification = {
        ...action.payload,
        id: generateNotificationId(),
        timestamp: Date.now(),
      };
      
      // 检查是否超过最大通知数量
      if (state.notifications.length >= state.settings.maxNotifications) {
        // 移除最旧的通知
        state.notifications.pop();
      }
      
      state.notifications.unshift(notification);
      updateStats(state);
    },
    
    // 批量添加通知
    addNotifications: (state, action: PayloadAction<Omit<Notification, 'id' | 'timestamp'>[]>) => {
      const notifications = action.payload.map(n => ({
        ...n,
        id: generateNotificationId(),
        timestamp: Date.now(),
      }));
      
      state.notifications.unshift(...notifications);
      
      // 保持最大通知数量限制
      if (state.notifications.length > state.settings.maxNotifications) {
        state.notifications = state.notifications.slice(0, state.settings.maxNotifications);
      }
      
      updateStats(state);
    },
    
    // 标记通知为已读
    markAsRead: (state, action: PayloadAction<string>) => {
      const notification = state.notifications.find(n => n.id === action.payload);
      if (notification && notification.status === 'unread') {
        notification.status = 'read';
        updateStats(state);
      }
    },
    
    // 批量标记为已读
    markMultipleAsRead: (state, action: PayloadAction<string[]>) => {
      action.payload.forEach(id => {
        const notification = state.notifications.find(n => n.id === id);
        if (notification && notification.status === 'unread') {
          notification.status = 'read';
        }
      });
      updateStats(state);
    },
    
    // 标记所有通知为已读
    markAllAsRead: (state) => {
      state.notifications.forEach(notification => {
        if (notification.status === 'unread') {
          notification.status = 'read';
        }
      });
      updateStats(state);
    },
    
    // 归档通知
    archiveNotification: (state, action: PayloadAction<string>) => {
      const notification = state.notifications.find(n => n.id === action.payload);
      if (notification) {
        notification.status = 'archived';
        updateStats(state);
      }
    },
    
    // 删除通知
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(n => n.id !== action.payload);
      updateStats(state);
    },
    
    // 批量删除通知
    removeMultipleNotifications: (state, action: PayloadAction<string[]>) => {
      state.notifications = state.notifications.filter(n => !action.payload.includes(n.id));
      updateStats(state);
    },
    
    // 清除所有通知
    clearAllNotifications: (state) => {
      state.notifications = [];
      updateStats(state);
    },
    
    // 清除已读通知
    clearReadNotifications: (state) => {
      state.notifications = state.notifications.filter(n => n.status === 'unread');
      updateStats(state);
    },
    
    // 设置过滤器
    setFilters: (state, action: PayloadAction<Partial<NotificationSliceState['filters']>>) => {
      state.filters = { ...state.filters, ...action.payload };
      state.currentPage = 1; // 重置到第一页
    },
    
    // 设置排序
    setSorting: (state, action: PayloadAction<{ sortBy: NotificationSliceState['sortBy']; sortOrder: NotificationSliceState['sortOrder'] }>) => {
      state.sortBy = action.payload.sortBy;
      state.sortOrder = action.payload.sortOrder;
    },
    
    // 设置分页
    setPagination: (state, action: PayloadAction<{ currentPage: number; pageSize: number }>) => {
      state.currentPage = action.payload.currentPage;
      state.pageSize = action.payload.pageSize;
    },
    
    // 选择通知
    selectNotifications: (state, action: PayloadAction<string[]>) => {
      state.selectedNotifications = action.payload;
    },
    
    // 添加选中通知
    addSelectedNotification: (state, action: PayloadAction<string>) => {
      if (!state.selectedNotifications.includes(action.payload)) {
        state.selectedNotifications.push(action.payload);
      }
    },
    
    // 移除选中通知
    removeSelectedNotification: (state, action: PayloadAction<string>) => {
      state.selectedNotifications = state.selectedNotifications.filter(id => id !== action.payload);
    },
    
    // 清除选中通知
    clearSelectedNotifications: (state) => {
      state.selectedNotifications = [];
    },
    
    // 更新通知设置
    updateSettings: (state, action: PayloadAction<Partial<NotificationSettings>>) => {
      state.settings = { ...state.settings, ...action.payload };
    },
    
    // 显示/隐藏通知面板
    togglePanel: (state) => {
      state.panelVisible = !state.panelVisible;
    },
    
    setPanelVisible: (state, action: PayloadAction<boolean>) => {
      state.panelVisible = action.payload;
    },
    
    // 显示/隐藏通知详情
    showNotificationDetail: (state, action: PayloadAction<Notification>) => {
      state.selectedNotification = action.payload;
      state.detailVisible = true;
      // 自动标记为已读
      if (action.payload.status === 'unread') {
        const notification = state.notifications.find(n => n.id === action.payload.id);
        if (notification) {
          notification.status = 'read';
          updateStats(state);
        }
      }
    },
    
    hideNotificationDetail: (state) => {
      state.detailVisible = false;
      state.selectedNotification = null;
    },
    
    // 设置实时连接状态
    setConnectionStatus: (state, action: PayloadAction<NotificationSliceState['connectionStatus']>) => {
      state.connectionStatus = action.payload;
    },
    
    // 启用/禁用实时通知
    setRealtimeEnabled: (state, action: PayloadAction<boolean>) => {
      state.realtimeEnabled = action.payload;
    },
    
    // 自动清理过期通知
    cleanupExpiredNotifications: (state) => {
      const now = Date.now();
      state.notifications = state.notifications.filter(n => !n.expiresAt || n.expiresAt > now);
      updateStats(state);
    },
    
    // 自动归档旧通知
    autoArchiveOldNotifications: (state) => {
      if (state.settings.autoArchive) {
        const cutoffTime = Date.now() - (state.settings.archiveAfterDays * 24 * 60 * 60 * 1000);
        state.notifications.forEach(notification => {
          if (notification.timestamp < cutoffTime && notification.status === 'read') {
            notification.status = 'archived';
          }
        });
        updateStats(state);
      }
    },
    
    // 设置加载状态
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    
    // 设置错误信息
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    
    // 重置通知状态
    resetNotificationState: () => initialState,
  },
});

// 导出actions
export const {
  addNotification,
  addNotifications,
  markAsRead,
  markMultipleAsRead,
  markAllAsRead,
  archiveNotification,
  removeNotification,
  removeMultipleNotifications,
  clearAllNotifications,
  clearReadNotifications,
  setFilters,
  setSorting,
  setPagination,
  addSelectedNotification,
  removeSelectedNotification,
  clearSelectedNotifications,
  updateSettings,
  togglePanel,
  setPanelVisible,
  showNotificationDetail,
  hideNotificationDetail,
  setConnectionStatus,
  setRealtimeEnabled,
  cleanupExpiredNotifications,
  autoArchiveOldNotifications,
  setLoading,
  setError,
  resetNotificationState,
} = notificationSlice.actions;

// 选择器
export const selectNotifications = (state: { notification: NotificationSliceState }) => state.notification.notifications;
export const selectUnreadCount = (state: { notification: NotificationSliceState }) => state.notification.stats.unread;
export const selectNotificationStats = (state: { notification: NotificationSliceState }) => state.notification.stats;
export const selectNotificationSettings = (state: { notification: NotificationSliceState }) => state.notification.settings;
export const selectSelectedNotifications = (state: { notification: NotificationSliceState }) => state.notification.selectedNotifications;
export const selectNotificationFilters = (state: { notification: NotificationSliceState }) => state.notification.filters;

// 导出reducer
export default notificationSlice.reducer;