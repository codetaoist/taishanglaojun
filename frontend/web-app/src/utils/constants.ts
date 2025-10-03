// API 配置
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8000/api',
  TIMEOUT: 10000,
  RETRY_ATTEMPTS: 3,
};

// 路由路径
export const ROUTES = {
  HOME: '/',
  WISDOM: '/wisdom',
  WISDOM_BROWSE: '/wisdom/browse',
  WISDOM_SEARCH: '/wisdom/search',
  WISDOM_FAVORITES: '/wisdom/favorites',
  CHAT: '/chat',
  CHAT_NEW: '/chat/new',
  CHAT_HISTORY: '/chat/history',
  COMMUNITY: '/community',
  COMMUNITY_DISCUSSIONS: '/community/discussions',
  COMMUNITY_EVENTS: '/community/events',
  PROFILE: '/profile',
  SETTINGS: '/settings',
  LOGIN: '/login',
  REGISTER: '/register',
} as const;

// 本地存储键名
export const STORAGE_KEYS = {
  TOKEN: 'token',
  USER: 'user',
  THEME: 'theme',
  LANGUAGE: 'language',
  CHAT_HISTORY: 'chat_history',
  PREFERENCES: 'preferences',
} as const;

// 分页配置
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: ['10', '20', '50', '100'],
  SHOW_SIZE_CHANGER: true,
  SHOW_QUICK_JUMPER: true,
} as const;

// 文件上传配置
export const UPLOAD_CONFIG = {
  MAX_FILE_SIZE: 10 * 1024 * 1024, // 10MB
  ALLOWED_IMAGE_TYPES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  ALLOWED_DOCUMENT_TYPES: ['application/pdf', 'text/plain', 'application/msword'],
} as const;

// 主题色彩
export const THEME_COLORS = {
  PRIMARY: '#0ea5e9',
  SUCCESS: '#059669',
  WARNING: '#d4af37',
  ERROR: '#dc2626',
  CULTURAL: {
    GOLD: '#d4af37',
    RED: '#dc143c',
    JADE: '#00a86b',
  },
} as const;

// 消息类型
export const MESSAGE_TYPES = {
  SUCCESS: 'success',
  ERROR: 'error',
  WARNING: 'warning',
  INFO: 'info',
} as const;

// 用户角色
export const USER_ROLES = {
  ADMIN: 'admin',
  USER: 'user',
  MODERATOR: 'moderator',
} as const;

// 智慧内容分类
export const WISDOM_CATEGORIES = {
  PHILOSOPHY: 'philosophy',
  LITERATURE: 'literature',
  HISTORY: 'history',
  CULTURE: 'culture',
  LIFE: 'life',
  MEDITATION: 'meditation',
} as const;

// 智慧内容分类标签
export const WISDOM_CATEGORY_LABELS = {
  [WISDOM_CATEGORIES.PHILOSOPHY]: '哲学思辨',
  [WISDOM_CATEGORIES.LITERATURE]: '文学经典',
  [WISDOM_CATEGORIES.HISTORY]: '历史智慧',
  [WISDOM_CATEGORIES.CULTURE]: '文化传承',
  [WISDOM_CATEGORIES.LIFE]: '人生感悟',
  [WISDOM_CATEGORIES.MEDITATION]: '修身养性',
} as const;

// 聊天消息状态
export const CHAT_MESSAGE_STATUS = {
  SENDING: 'sending',
  SENT: 'sent',
  DELIVERED: 'delivered',
  FAILED: 'failed',
} as const;

// 正则表达式
export const REGEX_PATTERNS = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  PHONE: /^1[3-9]\d{9}$/,
  USERNAME: /^[a-zA-Z0-9_\u4e00-\u9fa5]{2,20}$/,
  PASSWORD: /^(?=.*[a-zA-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{6,20}$/,
} as const;

// 错误消息
export const ERROR_MESSAGES = {
  NETWORK_ERROR: '网络连接失败，请检查网络设置',
  SERVER_ERROR: '服务器错误，请稍后重试',
  UNAUTHORIZED: '未授权访问，请重新登录',
  FORBIDDEN: '权限不足，无法访问',
  NOT_FOUND: '请求的资源不存在',
  VALIDATION_ERROR: '输入数据格式错误',
  UNKNOWN_ERROR: '未知错误，请联系管理员',
} as const;

// 成功消息
export const SUCCESS_MESSAGES = {
  LOGIN_SUCCESS: '登录成功',
  REGISTER_SUCCESS: '注册成功',
  LOGOUT_SUCCESS: '退出成功',
  SAVE_SUCCESS: '保存成功',
  DELETE_SUCCESS: '删除成功',
  UPDATE_SUCCESS: '更新成功',
} as const;