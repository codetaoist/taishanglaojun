import { authApi } from './authApi';

// API响应类型定义
interface AuthResponse {
  data: {
    token: string;
    expires_at: string;
    refresh_token?: string;
  };
}

// 令牌信息接口
export interface TokenInfo {
  token: string;
  expiresAt: string;
  refreshToken?: string;
}

// 令牌管理器类
export class TokenManager {
  private static instance: TokenManager;
  private tokenInfo: TokenInfo | null = null;
  private refreshPromise: Promise<void> | null = null;
  private readonly TOKEN_KEY = 'auth_token';
  private readonly EXPIRES_KEY = 'token_expires';
  private readonly REFRESH_TOKEN_KEY = 'refresh_token';
  private readonly DEFAULT_EXPIRY_HOURS = 1; // 默认过期时间（小时）

  // 单例模式
  static getInstance(): TokenManager {
    if (!TokenManager.instance) {
      TokenManager.instance = new TokenManager();
    }
    return TokenManager.instance;
  }

  // 私有构造函数
  private constructor() {
    this.loadTokenFromStorage();
  }

  // 从本地存储加载令牌
  private loadTokenFromStorage(): void {
    try {
      const token = localStorage.getItem(this.TOKEN_KEY);
      const expiresAt = localStorage.getItem(this.EXPIRES_KEY);
      const refreshToken = localStorage.getItem(this.REFRESH_TOKEN_KEY);

      if (token && expiresAt) {
        this.tokenInfo = {
          token,
          expiresAt,
          refreshToken: refreshToken || undefined,
        };
      }
    } catch (error) {
      console.error('加载令牌失败:', error);
      this.clearToken();
    }
  }

  // 保存令牌到本地存储
  private saveTokenToStorage(): void {
    if (!this.tokenInfo) return;

    try {
      localStorage.setItem(this.TOKEN_KEY, this.tokenInfo.token);
      localStorage.setItem(this.EXPIRES_KEY, this.tokenInfo.expiresAt);
      
      if (this.tokenInfo.refreshToken) {
        localStorage.setItem(this.REFRESH_TOKEN_KEY, this.tokenInfo.refreshToken);
      }
    } catch (error) {
      console.error('保存令牌失败:', error);
    }
  }

  // 设置令牌 - 支持多种参数形式
  setToken(tokenOrInfo: string | TokenInfo, refreshToken?: string, expiresInHours?: number): void {
    if (typeof tokenOrInfo === 'string') {
      // 如果第一个参数是字符串，则认为是token
      const expiresAt = new Date();
      expiresAt.setHours(expiresAt.getHours() + (expiresInHours || this.DEFAULT_EXPIRY_HOURS));
      
      this.tokenInfo = {
        token: tokenOrInfo,
        expiresAt: expiresAt.toISOString(),
        refreshToken,
      };
    } else {
      // 如果第一个参数是对象，则直接使用
      this.tokenInfo = tokenOrInfo;
    }
    
    this.saveTokenToStorage();
  }

  // 获取当前令牌
  getToken(): string | null {
    return this.tokenInfo?.token || null;
  }

  // 获取令牌信息
  getTokenInfo(): TokenInfo | null {
    return this.tokenInfo;
  }

  // 检查令牌是否即将过期（5分钟内）
  isTokenExpiringSoon(): boolean {
    if (!this.tokenInfo) return true;

    const now = new Date();
    const expiresAt = new Date(this.tokenInfo.expiresAt);
    const fiveMinutesFromNow = new Date(now.getTime() + 5 * 60 * 1000);

    return expiresAt <= fiveMinutesFromNow;
  }

  // 检查令牌是否已过期
  isTokenExpired(): boolean {
    if (!this.tokenInfo) return true;

    const now = new Date();
    const expiresAt = new Date(this.tokenInfo.expiresAt);

    return expiresAt <= now;
  }

  // 刷新令牌
  async refreshToken(): Promise<void> {
    // 如果已经在刷新中，返回现有的Promise
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    // 如果没有令牌，直接返回错误
    if (!this.tokenInfo) {
      throw new Error('没有可刷新的令牌');
    }

    // 创建刷新Promise
    this.refreshPromise = this.doRefreshToken();

    try {
      await this.refreshPromise;
    } finally {
      this.refreshPromise = null;
    }
  }

  // 实际执行令牌刷新
  private async doRefreshToken(): Promise<void> {
    try {
      const response = await authApi.refreshToken() as AuthResponse;
      this.setToken({
        token: response.data.token,
        expiresAt: response.data.expires_at,
        refreshToken: response.data.refresh_token || this.tokenInfo?.refreshToken,
      });
    } catch (error) {
      console.error('令牌刷新失败:', error);
      this.clearToken();
      throw error;
    }
  }

  // 获取有效的令牌（如果需要则自动刷新）
  async getValidToken(): Promise<string | null> {
    if (!this.tokenInfo) return null;

    // 如果令牌已过期，清除并返回null
    if (this.isTokenExpired()) {
      this.clearToken();
      return null;
    }

    // 如果令牌即将过期，尝试刷新
    if (this.isTokenExpiringSoon()) {
      try {
        await this.refreshToken();
      } catch (error) {
        console.error('自动刷新令牌失败:', error);
        this.clearToken();
        return null;
      }
    }

    return this.tokenInfo.token;
  }

  // 清除令牌
  clearToken(): void {
    this.tokenInfo = null;
    try {
      localStorage.removeItem(this.TOKEN_KEY);
      localStorage.removeItem(this.EXPIRES_KEY);
      localStorage.removeItem(this.REFRESH_TOKEN_KEY);
    } catch (error) {
      console.error('清除令牌失败:', error);
    }
  }

  // 检查是否有有效的令牌
  hasValidToken(): boolean {
    return !!this.tokenInfo && !this.isTokenExpired();
  }
}

// 导出单例实例
export const tokenManager = TokenManager.getInstance();