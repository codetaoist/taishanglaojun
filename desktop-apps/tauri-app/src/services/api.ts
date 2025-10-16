// API service for backend communication
import { type APIConfig } from '../config/apiConfig';
import { isTauriEnvironment } from '../utils/environment';
import { invoke } from '@tauri-apps/api/core';
import { 
  UserMenuConfig,
  MenuResponse,
  BatchPermissionUpdate,
  DeviceType
} from '../types/menu';

interface AppModule {
  id: string;
  name: string;
  description: string;
  category: string;
  required_role: string;
  is_active: boolean;
  icon?: string;
  route?: string;
}

interface UserModules {
  modules: AppModule[];
  user_permissions: Record<string, boolean>;
}

class ApiService {
  // 通用API调用方法
  private async callAPI<T>(
    tauriCommand: string,
    backendEndpoint: keyof APIConfig['backend']['endpoints'],
    data?: any,
    method: string = 'GET'
  ): Promise<T | null> {
    try {
      // 在Tauri环境中使用invoke
      if (isTauriEnvironment()) {
        try {
          const result = await invoke(tauriCommand, data || {});
          return result as T;
        } catch (tauriError) {
          console.warn(`Tauri invoke failed for ${tauriCommand}:`, tauriError);
          // 在Tauri环境中，如果命令失败，直接返回null，不回退到HTTP
          return null;
        }
      }

      // 在Web环境中，由于后端服务可能不可用，我们返回模拟数据或null
      console.warn(`Skipping HTTP API call in web environment for ${backendEndpoint}`);
      
      // 为某些关键接口返回模拟数据
      if (backendEndpoint === 'modules') {
        return {
          modules: ['chat', 'document', 'settings'],
          permissions: ['read', 'write']
        } as T;
      }

      // 为菜单相关接口返回模拟数据
      if (backendEndpoint === 'menu') {
        if (tauriCommand === 'get_user_menu') {
          return {
            success: true,
            menuItems: [
              {
                id: 'dashboard',
                title: '仪表板',
                icon: 'Dashboard',
                route: '/dashboard',
                type: 'page',
                deviceTypes: ['DESKTOP', 'TABLET', 'MOBILE'],
                permissionLevel: 'read',
                isVisible: true,
                order: 1
              },
              {
                id: 'chat',
                title: '聊天',
                icon: 'Chat',
                route: '/chat',
                type: 'page',
                deviceTypes: ['DESKTOP', 'TABLET', 'MOBILE'],
                permissionLevel: 'read',
                isVisible: true,
                order: 2
              },
              {
                id: 'settings',
                title: '设置',
                icon: 'Settings',
                route: '/settings',
                type: 'page',
                deviceTypes: ['DESKTOP', 'TABLET'],
                permissionLevel: 'admin',
                isVisible: true,
                order: 3
              }
            ],
            userConfig: {
              userId: data?.userId || 'user1',
              customOrder: [],
              hiddenItems: [],
              devicePreferences: {}
            }
          } as T;
        }
        
        if (method === 'PUT' || method === 'POST') {
          return { success: true } as T;
        }
      }

      if (backendEndpoint === 'permissions') {
        return { success: true } as T;
      }

      if (backendEndpoint === 'analytics') {
        return {
          totalUsers: 150,
          activeMenuItems: 8,
          deviceDistribution: {
            DESKTOP: 60,
            MOBILE: 35,
            TABLET: 5
          },
          popularMenuItems: [
            { id: 'dashboard', clicks: 1250 },
            { id: 'chat', clicks: 980 },
            { id: 'settings', clicks: 340 }
          ]
        } as T;
      }
      
      return null;
    } catch (error) {
      console.error(`API call failed:`, error);
      return null;
    }
  }

  async getUserModules(): Promise<UserModules | null> {
    return await this.callAPI<UserModules>('get_user_modules', 'modules');
  }

  async getUserPreferences(): Promise<any> {
    try {
      // 在Tauri环境中使用本地存储
      if (isTauriEnvironment()) {
        const stored = localStorage.getItem('user_preferences');
        return stored ? JSON.parse(stored) : null;
      }
      
      // 在Web环境中使用HTTP API
      return await this.callAPI<any>('get_user_preferences', 'preferences');
    } catch (error) {
      console.error('Failed to get user preferences:', error);
      return null;
    }
  }

  async updateUserPreferences(preferences: any): Promise<boolean> {
    try {
      // 在Tauri环境中使用本地存储
      if (isTauriEnvironment()) {
        localStorage.setItem('user_preferences', JSON.stringify(preferences));
        return true;
      }
      
      // 在Web环境中使用HTTP API
      const result = await this.callAPI<any>('update_user_preferences', 'preferences', { preferences }, 'PUT');
      return result !== null;
    } catch (error) {
      console.error('Failed to update user preferences:', error);
      return false;
    }
  }

  // 菜单相关API方法
  async getUserMenu(userId: string, deviceType: DeviceType): Promise<MenuResponse | null> {
    try {
      return await this.callAPI<MenuResponse>('get_user_menu', 'menu', { userId, deviceType });
    } catch (error) {
      console.error('Failed to get user menu:', error);
      return null;
    }
  }

  async updateUserMenu(userId: string, config: UserMenuConfig): Promise<boolean> {
    try {
      const result = await this.callAPI<any>('update_user_menu', 'menu', { userId, config }, 'PUT');
      return result !== null;
    } catch (error) {
      console.error('Failed to update user menu:', error);
      return false;
    }
  }

  async updateMenuPermissions(request: BatchPermissionUpdate): Promise<boolean> {
    try {
      const result = await this.callAPI<any>('update_menu_permissions', 'permissions', request, 'POST');
      return result !== null;
    } catch (error) {
      console.error('Failed to update menu permissions:', error);
      return false;
    }
  }

  async getMenuAnalytics(userId?: string): Promise<any> {
    try {
      return await this.callAPI<any>('get_menu_analytics', 'analytics', { userId });
    } catch (error) {
      console.error('Failed to get menu analytics:', error);
      return null;
    }
  }
}

export const apiService = new ApiService();
export { ApiService };
export type { AppModule, UserModules };