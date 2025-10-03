import axios from 'axios';

// 行为类型枚举
export enum ActionType {
  VIEW = 'view',
  LIKE = 'like',
  SHARE = 'share',
  COMMENT = 'comment',
  SEARCH = 'search',
  CLICK = 'click',
  DOWNLOAD = 'download',
  BOOKMARK = 'bookmark'
}

// 目标类型枚举
export enum TargetType {
  WISDOM = 'wisdom',
  CATEGORY = 'category',
  AUTHOR = 'author',
  TAG = 'tag',
  SEARCH_RESULT = 'search_result'
}

// 行为记录请求接口
export interface BehaviorRequest {
  userID: string;
  wisdomID?: string;
  actionType: ActionType;
  targetType?: TargetType;
  targetID?: string;
  actionValue?: number;
  sessionID?: string;
  ipAddress?: string;
  userAgent?: string;
  metadata?: Record<string, any>;
}

// 用户画像接口
export interface UserProfile {
  userID: string;
  totalBehaviors: number;
  categories: CategoryScore[];
  schools: SchoolScore[];
  authors: AuthorScore[];
  tags: TagScore[];
  preferences: {
    difficulty: string;
    readingSpeed: number;
  };
  lastActive: string;
  createdAt: string;
  updatedAt: string;
}

export interface CategoryScore {
  category: string;
  score: number;
  count: number;
}

export interface SchoolScore {
  school: string;
  score: number;
  count: number;
}

export interface AuthorScore {
  author: string;
  score: number;
  count: number;
}

export interface TagScore {
  tag: string;
  score: number;
  count: number;
}

// API响应接口
export interface ApiResponse<T> {
  code?: number;
  message: string;
  data?: T;
}

export interface UserProfileResponse extends ApiResponse<UserProfile> {}

export interface SimilarUsersResponse extends ApiResponse<string[]> {
  total: number;
}

class BehaviorService {
  private baseURL = '/api/v1/user-behavior';
  private sessionID: string;
  private userID: string | null = null;

  constructor() {
    // 生成会话ID
    this.sessionID = this.generateSessionID();
    
    // 从localStorage获取用户ID（如果已登录）
    this.userID = localStorage.getItem('userID');
  }

  // 生成会话ID
  private generateSessionID(): string {
    return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }

  // 设置用户ID
  setUserID(userID: string) {
    this.userID = userID;
    localStorage.setItem('userID', userID);
  }

  // 获取用户ID
  getUserID(): string | null {
    return this.userID;
  }

  // 记录用户行为
  async recordBehavior(request: Partial<BehaviorRequest>): Promise<void> {
    try {
      // 如果没有用户ID，使用访客ID
      const userID = this.userID || this.getOrCreateGuestID();
      
      const behaviorData: BehaviorRequest = {
        userID,
        sessionID: this.sessionID,
        actionType: request.actionType!,
        targetType: request.targetType || TargetType.WISDOM,
        wisdomID: request.wisdomID,
        targetID: request.targetID,
        actionValue: request.actionValue || 1,
        metadata: {
          timestamp: new Date().toISOString(),
          url: window.location.href,
          referrer: document.referrer,
          ...request.metadata
        }
      };

      await axios.post(`${this.baseURL}/record`, behaviorData);
    } catch (error) {
      console.warn('Failed to record behavior:', error);
      // 不抛出错误，避免影响用户体验
    }
  }

  // 获取或创建访客ID
  private getOrCreateGuestID(): string {
    let guestID = localStorage.getItem('guestID');
    if (!guestID) {
      guestID = 'guest_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
      localStorage.setItem('guestID', guestID);
    }
    return guestID;
  }

  // 获取用户画像
  async getUserProfile(userID?: string): Promise<UserProfile | null> {
    try {
      const targetUserID = userID || this.userID;
      if (!targetUserID) {
        return null;
      }

      const response = await axios.get<UserProfileResponse>(`${this.baseURL}/profile`, {
        params: { user_id: targetUserID }
      });

      return response.data.data || null;
    } catch (error) {
      console.error('Failed to get user profile:', error);
      return null;
    }
  }

  // 获取相似用户
  async getSimilarUsers(userID?: string, limit = 10): Promise<string[]> {
    try {
      const targetUserID = userID || this.userID;
      if (!targetUserID) {
        return [];
      }

      const response = await axios.get<SimilarUsersResponse>(`${this.baseURL}/similar-users`, {
        params: { 
          user_id: targetUserID,
          limit 
        }
      });

      return response.data.data || [];
    } catch (error) {
      console.error('Failed to get similar users:', error);
      return [];
    }
  }

  // 便捷方法：记录智慧浏览行为
  async recordWisdomView(wisdomID: string, metadata?: Record<string, any>) {
    await this.recordBehavior({
      actionType: ActionType.VIEW,
      targetType: TargetType.WISDOM,
      wisdomID,
      targetID: wisdomID,
      metadata
    });
  }

  // 便捷方法：记录智慧点赞行为
  async recordWisdomLike(wisdomID: string, isLike: boolean = true) {
    await this.recordBehavior({
      actionType: ActionType.LIKE,
      targetType: TargetType.WISDOM,
      wisdomID,
      targetID: wisdomID,
      actionValue: isLike ? 1 : -1,
      metadata: { isLike }
    });
  }

  // 便捷方法：记录智慧分享行为
  async recordWisdomShare(wisdomID: string, shareMethod?: string) {
    await this.recordBehavior({
      actionType: ActionType.SHARE,
      targetType: TargetType.WISDOM,
      wisdomID,
      targetID: wisdomID,
      metadata: { shareMethod }
    });
  }

  // 便捷方法：记录搜索行为
  async recordSearch(query: string, resultCount?: number) {
    await this.recordBehavior({
      actionType: ActionType.SEARCH,
      targetType: TargetType.SEARCH_RESULT,
      targetID: query,
      metadata: { 
        query, 
        resultCount,
        searchTime: new Date().toISOString()
      }
    });
  }

  // 便捷方法：记录点击行为
  async recordClick(targetType: TargetType, targetID: string, metadata?: Record<string, any>) {
    await this.recordBehavior({
      actionType: ActionType.CLICK,
      targetType,
      targetID,
      metadata
    });
  }

  // 便捷方法：记录收藏行为
  async recordBookmark(wisdomID: string, isBookmark: boolean = true) {
    await this.recordBehavior({
      actionType: ActionType.BOOKMARK,
      targetType: TargetType.WISDOM,
      wisdomID,
      targetID: wisdomID,
      actionValue: isBookmark ? 1 : -1,
      metadata: { isBookmark }
    });
  }
}

// 创建单例实例
export const behaviorService = new BehaviorService();

// 导出默认实例
export default behaviorService;