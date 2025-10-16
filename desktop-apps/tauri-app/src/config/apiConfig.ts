// API配置管理 - 区分桌面端和总后台接口
import { isTauriEnvironment } from '../utils/environment';

export interface APIEndpoint {
  name: string;
  url: string;
  timeout: number;
  retries: number;
  headers?: Record<string, string>;
}

export interface APIConfig {
  // 桌面端本地API (Tauri后端)
  desktop: {
    baseUrl: string;
    endpoints: {
      auth: APIEndpoint;
      chat: APIEndpoint;
      fileTransfer: APIEndpoint;
      system: APIEndpoint;
      ai: APIEndpoint;
      project: APIEndpoint;
      friend: APIEndpoint;
      menu: APIEndpoint;
    };
  };
  
  // 总后台API (Web服务)
  backend: {
    baseUrl: string;
    endpoints: {
      user: APIEndpoint;
      modules: APIEndpoint;
      preferences: APIEndpoint;
      sync: APIEndpoint;
      analytics: APIEndpoint;
      menu: APIEndpoint;
      permissions: APIEndpoint;
    };
  };
  
  // 环境配置
  environment: 'development' | 'production' | 'testing';
  
  // 是否启用后端同步
  enableBackendSync: boolean;
  
  // 是否优先使用本地API
  preferLocalAPI: boolean;
}

// 默认配置
export const defaultAPIConfig: APIConfig = {
  desktop: {
    baseUrl: 'tauri://localhost',
    endpoints: {
      auth: {
        name: 'Desktop Auth',
        url: 'tauri://localhost/auth',
        timeout: 5000,
        retries: 2
      },
      chat: {
        name: 'Desktop Chat',
        url: 'tauri://localhost/chat',
        timeout: 30000,
        retries: 1
      },
      fileTransfer: {
        name: 'Desktop File Transfer',
        url: 'tauri://localhost/file-transfer',
        timeout: 60000,
        retries: 1
      },
      system: {
        name: 'Desktop System',
        url: 'tauri://localhost/system',
        timeout: 10000,
        retries: 2
      },
      ai: {
        name: 'Desktop AI',
        url: 'tauri://localhost/ai',
        timeout: 60000,
        retries: 1
      },
      project: {
        name: 'Desktop Project',
        url: 'tauri://localhost/project',
        timeout: 15000,
        retries: 2
      },
      friend: {
        name: 'Desktop Friend',
        url: 'tauri://localhost/friend',
        timeout: 10000,
        retries: 2
      },
      menu: {
        name: 'Desktop Menu',
        url: 'tauri://localhost/menu',
        timeout: 10000,
        retries: 2
      }
    }
  },
  
  backend: {
    baseUrl: process.env.NODE_ENV === 'production' 
      ? 'https://api.taishanglaojun.com'
      : 'http://localhost:3001',
    endpoints: {
      user: {
        name: 'Backend User',
        url: '/user',
        timeout: 10000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      },
      modules: {
        name: 'Backend Modules',
        url: '/modules',
        timeout: 10000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      },
      preferences: {
        name: 'Backend Preferences',
        url: '/preferences',
        timeout: 10000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      },
      sync: {
        name: 'Backend Sync',
        url: '/sync',
        timeout: 30000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      },
      analytics: {
        name: 'Backend Analytics',
        url: '/analytics',
        timeout: 15000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      },
      menu: {
        name: 'Backend Menu',
        url: '/menu',
        timeout: 10000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      },
      permissions: {
        name: 'Backend Permissions',
        url: '/permissions',
        timeout: 10000,
        retries: 2,
        headers: {
          'Content-Type': 'application/json'
        }
      }
    }
  },
  
  environment: process.env.NODE_ENV as 'development' | 'production' | 'testing' || 'development',
  enableBackendSync: true,
  preferLocalAPI: true
};

// 环境检测
export const isWebEnvironment = (): boolean => {
  return !isTauriEnvironment();
};

// API配置管理器
export class APIConfigManager {
  private config: APIConfig;
  
  constructor(config: APIConfig = defaultAPIConfig) {
    this.config = config;
  }
  
  // 获取配置
  getConfig(): APIConfig {
    return this.config;
  }
  
  // 更新配置
  updateConfig(newConfig: Partial<APIConfig>): void {
    this.config = { ...this.config, ...newConfig };
    this.saveConfig();
  }
  
  // 获取桌面端API端点
  getDesktopEndpoint(name: keyof APIConfig['desktop']['endpoints']): APIEndpoint {
    return this.config.desktop.endpoints[name];
  }
  
  // 获取后端API端点
  getBackendEndpoint(name: keyof APIConfig['backend']['endpoints']): APIEndpoint {
    const endpoint = this.config.backend.endpoints[name];
    return {
      ...endpoint,
      url: this.config.backend.baseUrl + endpoint.url
    };
  }
  
  // 根据环境和配置选择API端点
  selectEndpoint(
    desktopEndpoint: keyof APIConfig['desktop']['endpoints'],
    backendEndpoint: keyof APIConfig['backend']['endpoints']
  ): APIEndpoint {
    // 在Tauri环境中且优先使用本地API
    if (isTauriEnvironment() && this.config.preferLocalAPI) {
      return this.getDesktopEndpoint(desktopEndpoint);
    }
    
    // 在Web环境中或配置为使用后端API
    return this.getBackendEndpoint(backendEndpoint);
  }
  
  // 检查后端健康状态
  async checkBackendHealth(): Promise<boolean> {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        // 在Tauri环境中使用Tauri命令
        try {
          const { invoke } = await import('@tauri-apps/api/core');
          const health = await invoke('health_check');
          return health as boolean;
        } catch (tauriError) {
          console.warn('Tauri health check failed:', tauriError);
          // 在Tauri环境中，如果命令失败，假设服务不可用
          return false;
        }
      } else {
        // 在Web环境中，由于后端服务可能不可用，我们跳过健康检查
        // 或者返回true假设服务可用，让实际的API调用来处理错误
        console.warn('Skipping backend health check in web environment');
        return true;
      }
    } catch (error) {
      console.warn('Backend health check failed:', error);
      return false;
    }
  }
  
  // 自动选择可用的API
  async selectAvailableEndpoint(
    desktopEndpoint: keyof APIConfig['desktop']['endpoints'],
    backendEndpoint: keyof APIConfig['backend']['endpoints']
  ): Promise<APIEndpoint> {
    // 在Tauri环境中，直接使用桌面端API
    if (isTauriEnvironment()) {
      return this.getDesktopEndpoint(desktopEndpoint);
    }
    
    // 在Web环境中，使用后端API
    return this.getBackendEndpoint(backendEndpoint);
  }
  
  // 保存配置到本地存储
  private saveConfig(): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem('api-config', JSON.stringify(this.config));
    }
  }
  
  // 从本地存储加载配置
  static loadConfig(): APIConfig {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('api-config');
      if (saved) {
        try {
          return { ...defaultAPIConfig, ...JSON.parse(saved) };
        } catch (error) {
          console.warn('Failed to load API config:', error);
        }
      }
    }
    return defaultAPIConfig;
  }
}

// 全局API配置管理器实例
export const apiConfigManager = new APIConfigManager(APIConfigManager.loadConfig());

// 便捷函数
export const getAPIConfig = () => apiConfigManager.getConfig();
export const updateAPIConfig = (config: Partial<APIConfig>) => apiConfigManager.updateConfig(config);
export const selectEndpoint = (
  desktop: keyof APIConfig['desktop']['endpoints'],
  backend: keyof APIConfig['backend']['endpoints']
) => apiConfigManager.selectEndpoint(desktop, backend);
export const selectAvailableEndpoint = (
  desktop: keyof APIConfig['desktop']['endpoints'],
  backend: keyof APIConfig['backend']['endpoints']
) => apiConfigManager.selectAvailableEndpoint(desktop, backend);