// 用户相关类型
export interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  name?: string;
  role: 'user' | 'admin' | 'moderator';
  roles?: string[]; // 支持多角色
  permissions?: string[]; // 用户权限列表
  isAdmin?: boolean;
  createdAt: string;
  updatedAt: string;
}

// 文化智慧相关类型
export interface CulturalWisdom {
  id: string;
  title: string;
  content: string;
  category: string;
  tags: string[];
  author: string;
  source?: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  likes: number;
  views: number;
  favorites?: number;
  dynasty?: string;
  explanation?: string;
  createdAt: string;
  updatedAt: string;
}

// API响应类型
export interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
  error?: string;
}

// 分页类型
export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

// 搜索过滤器类型
export interface SearchFilters {
  keyword?: string;
  category?: string;
  tags?: string[];
  difficulty?: string;
  sortBy?: 'created_at' | 'updated_at' | 'likes' | 'views';
  sortOrder?: 'asc' | 'desc';
}

// 分类类型
export interface Category {
  id: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

// 标签类型
export interface Tag {
  id: string;
  name: string;
  color?: string;
  createdAt: string;
  updatedAt: string;
}

// 用户数据类型
export interface UserData {
  username?: string;
  email?: string;
  name?: string;
  avatar?: string;
  role?: 'user' | 'admin' | 'moderator';
}

// 智慧创建/更新数据类型
export interface WisdomData {
  title: string;
  content: string;
  category: string;
  tags: string[];
  author?: string;
  source?: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  dynasty?: string;
  explanation?: string;
}

// 分类数据类型
export interface CategoryData {
  name: string;
  description?: string;
}

// 用户统计类型
export interface UserStats {
  totalUsers: number;
  activeUsers: number;
  adminUsers: number;
  newUsersToday: number;
  onlineUsers: number;
}

// 用户查询参数类型
export interface UserQueryParams {
  page?: number;
  limit?: number;
  search?: string;
  role?: string;
  status?: string;
}

// 分析数据类型
export interface AnalyticsData {
  pageViews: number;
  uniqueVisitors: number;
  bounceRate: number;
  averageSessionDuration: number;
  topPages: Array<{ path: string; views: number }>;
  userGrowth: Array<{ date: string; count: number }>;
}

// 系统配置类型
export interface SystemConfig {
  siteName: string;
  siteDescription: string;
  maintenanceMode: boolean;
  allowRegistration: boolean;
  maxFileSize: number;
}

// SEO配置类型
export interface SEOConfig {
  title: string;
  description: string;
  keywords: string[];
  ogImage?: string;
}

// 缓存配置类型
export interface CacheConfig {
  enabled: boolean;
  ttl: number;
  maxSize: number;
  strategy: 'lru' | 'fifo' | 'lfu';
}

// 通知类型
export interface Notification {
  id: string;
  title: string;
  content: string;
  type: 'info' | 'success' | 'warning' | 'error';
  read: boolean;
  createdAt: string;
  updatedAt: string;
}

// 通知数据类型
export interface NotificationData {
  title: string;
  content: string;
  type: 'info' | 'success' | 'warning' | 'error';
  targetUsers?: string[];
}

// 审核项目类型
export interface ReviewItem {
  id: string;
  type: 'wisdom' | 'comment' | 'user';
  title: string;
  content: string;
  status: 'pending' | 'approved' | 'rejected';
  submittedBy: string;
  createdAt: string;
  updatedAt: string;
}

// 审核数据类型
export interface ReviewData {
  status: 'approved' | 'rejected';
  reason?: string;
}

// 批量审核数据类型
export interface BatchReviewData {
  ids: string[];
  status: 'approved' | 'rejected';
  reason?: string;
}

// 查询参数类型
export interface QueryParams {
  page?: number;
  limit?: number;
  search?: string;
  status?: string;
  type?: string;
  startDate?: string;
  endDate?: string;
}

// 标签数据类型
export interface TagData {
  name: string;
  color?: string;
}

// 聊天相关类型
export interface ChatMessage {
  id: string;
  content: string;
  role: 'user' | 'assistant';
  timestamp: string;
  metadata?: {
    sources?: string[];
    confidence?: number;
    wisdomId?: string;
    wisdomTitle?: string;
    category?: string;
  };
}

export interface ChatSession {
  id: string;
  title: string;
  messages: ChatMessage[];
  createdAt: string;
  updatedAt: string;
}

// 对话相关类型
export interface Conversation {
  id: string;
  title: string;
  messages: ChatMessage[];
  createdAt: string;
  updatedAt: string;
  sessionId?: string;
  isArchived: boolean;
  messageCount: number;
}

export interface ConversationSummary {
  id: string;
  title: string;
  lastMessage: string;
  messageCount: number;
  createdAt: string;
  updatedAt: string;
  isArchived: boolean;
}