import { 
  MenuItem, 
  UserMenuConfig, 
  MenuResponse, 
  MenuUpdateRequest, 
  BatchPermissionUpdate,
  DeviceType, 
  PermissionLevel, 
  MenuItemType 
} from '../types/menu';
import { getDeviceInfo } from '../utils/deviceDetection';
import { ApiService } from './api';

class MenuService {
  private menuCache: Map<string, MenuItem[]> = new Map();
  private userConfigCache: Map<string, UserMenuConfig> = new Map();
  private apiService = new ApiService();

  // 获取用户菜单
  async getUserMenu(userId: string, forceRefresh = false): Promise<MenuResponse> {
    const cacheKey = `${userId}_${getDeviceInfo().type}`;
    
    if (!forceRefresh && this.menuCache.has(cacheKey)) {
      const cachedMenu = this.menuCache.get(cacheKey)!;
      const cachedConfig = this.userConfigCache.get(userId);
      
      if (cachedConfig) {
        return {
          success: true,
          data: {
            menuItems: cachedMenu,
            userConfig: cachedConfig,
            deviceInfo: getDeviceInfo()
          }
        };
      }
    }

    try {
      const response = await this.apiService.getUserMenu(userId, getDeviceInfo().type);

      if (response) {
        this.menuCache.set(cacheKey, response.data.menuItems);
        this.userConfigCache.set(userId, response.data.userConfig);
        return response;
      }

      // 如果API调用失败，返回默认菜单
      const defaultMenuItems = this.getDefaultMenu();
      const defaultConfig = this.getDefaultUserConfig(userId);
      
      return {
        success: true,
        data: {
          menuItems: defaultMenuItems,
          userConfig: defaultConfig,
          deviceInfo: getDeviceInfo()
        }
      };
    } catch (error) {
      console.error('Failed to fetch user menu:', error);
      
      // 返回默认菜单作为后备
      const defaultMenu = this.getDefaultMenu();
      const defaultConfig = this.getDefaultUserConfig(userId);
      
      return {
        success: true,
        data: {
          menuItems: defaultMenu,
          userConfig: defaultConfig,
          deviceInfo: getDeviceInfo()
        },
        message: 'Using default menu due to network error'
      };
    }
  }

  // 过滤菜单项基于权限和设备
  filterMenuItems(
    menuItems: MenuItem[], 
    userPermissions: PermissionLevel[], 
    deviceType: DeviceType
  ): MenuItem[] {
    return menuItems
      .filter(item => this.hasPermission(item, userPermissions))
      .filter(item => this.supportsDevice(item, deviceType))
      .map(item => this.adaptMenuItemForDevice(item, deviceType))
      .sort((a, b) => a.order - b.order);
  }

  // 检查用户是否有权限访问菜单项
  private hasPermission(item: MenuItem, userPermissions: PermissionLevel[]): boolean {
    // 检查必需权限
    const hasRequiredPermission = item.requiredPermissions.some(permission => 
      userPermissions.includes(permission)
    );

    if (!hasRequiredPermission) {
      return false;
    }

    // 检查排除权限
    if (item.excludedPermissions) {
      const hasExcludedPermission = item.excludedPermissions.some(permission => 
        userPermissions.includes(permission)
      );
      
      if (hasExcludedPermission) {
        return false;
      }
    }

    return true;
  }

  // 检查菜单项是否支持当前设备
  private supportsDevice(item: MenuItem, deviceType: DeviceType): boolean {
    return item.supportedDevices.includes(deviceType);
  }

  // 为特定设备适配菜单项
  private adaptMenuItemForDevice(item: MenuItem, deviceType: DeviceType): MenuItem {
    const deviceConfig = item.deviceSpecificConfig?.[deviceType];
    
    if (!deviceConfig) {
      return item;
    }

    return {
      ...item,
      title: deviceConfig.title || item.title,
      icon: deviceConfig.icon || item.icon,
      isVisible: item.isVisible && !deviceConfig.hidden,
      order: deviceConfig.order !== undefined ? deviceConfig.order : item.order,
      children: item.children ? 
        item.children.map(child => this.adaptMenuItemForDevice(child, deviceType)) : 
        undefined
    };
  }

  // 更新用户菜单配置
  async updateUserMenu(request: MenuUpdateRequest): Promise<boolean> {
    try {
      // 构建完整的用户配置对象
      const userConfig: UserMenuConfig = {
        userId: request.userId,
        deviceType: getDeviceInfo().type,
        userPermissions: [],
        lastUpdated: new Date().toISOString(),
        ...request.userConfig
      };

      const response = await this.apiService.updateUserMenu(request.userId, userConfig);

      if (response) {
        // 清除相关缓存
        this.clearUserCache(request.userId);
        if (request.adminAction?.targetUserId) {
          this.clearUserCache(request.adminAction.targetUserId);
        }
      }

      return response;
    } catch (error) {
      console.error('Failed to update user menu:', error);
      return false;
    }
  }

  // 批量更新权限
  async batchUpdatePermissions(update: BatchPermissionUpdate): Promise<boolean> {
    try {
      const response = await this.apiService.updateMenuPermissions(update);

      if (response) {
        // 清除所有受影响用户的缓存
        update.userIds.forEach(userId => this.clearUserCache(userId));
      }

      return response;
    } catch (error) {
      console.error('Failed to batch update permissions:', error);
      return false;
    }
  }

  // 获取默认菜单
  private getDefaultMenu(): MenuItem[] {
    const deviceType = getDeviceInfo().type;
    
    const baseMenuItems: MenuItem[] = [
      {
        id: 'dashboard',
        title: '仪表板',
        description: '系统概览和统计信息',
        icon: 'dashboard',
        type: MenuItemType.PAGE,
        route: '/dashboard',
        requiredPermissions: [PermissionLevel.USER],
        supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET, DeviceType.MOBILE],
        order: 1,
        isActive: true,
        isVisible: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        createdBy: 'system'
      },
      {
        id: 'chat',
        title: '聊天',
        description: '实时通讯功能',
        icon: 'chat',
        type: MenuItemType.PAGE,
        route: '/chat',
        requiredPermissions: [PermissionLevel.USER],
        supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET, DeviceType.MOBILE, DeviceType.WATCH],
        deviceSpecificConfig: {
          [DeviceType.WATCH]: {
            title: '消息',
            icon: 'message'
          }
        },
        order: 2,
        isActive: true,
        isVisible: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        createdBy: 'system'
      },
      {
        id: 'menu-test',
        title: '菜单测试',
        description: '菜单功能测试页面',
        icon: 'science',
        type: MenuItemType.PAGE,
        route: '/menu-test',
        requiredPermissions: [PermissionLevel.USER],
        supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET],
        order: 3,
        isActive: true,
        isVisible: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        createdBy: 'system'
      },
      {
        id: 'settings',
        title: '设置',
        description: '系统设置和配置',
        icon: 'settings',
        type: MenuItemType.PAGE,
        route: '/settings',
        requiredPermissions: [PermissionLevel.USER],
        supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET, DeviceType.MOBILE],
        order: 10,
        isActive: true,
        isVisible: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        createdBy: 'system'
      },
      {
        id: 'admin',
        title: '管理面板',
        description: '系统管理功能',
        icon: 'admin_panel_settings',
        type: MenuItemType.SUBMENU,
        requiredPermissions: [PermissionLevel.ADMIN],
        supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET],
        children: [
          {
            id: 'user-management',
            title: '用户管理',
            icon: 'people',
            type: MenuItemType.PAGE,
            route: '/admin/users',
            requiredPermissions: [PermissionLevel.ADMIN],
            supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET],
            order: 1,
            isActive: true,
            isVisible: true,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            createdBy: 'system'
          },
          {
            id: 'permission-management',
            title: '权限管理',
            icon: 'security',
            type: MenuItemType.PAGE,
            route: '/admin/permissions',
            requiredPermissions: [PermissionLevel.ADMIN],
            supportedDevices: [DeviceType.DESKTOP, DeviceType.TABLET],
            order: 2,
            isActive: true,
            isVisible: true,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            createdBy: 'system'
          }
        ],
        order: 20,
        isActive: true,
        isVisible: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        createdBy: 'system'
      }
    ];

    // 根据设备类型过滤菜单
    return baseMenuItems.filter(item => item.supportedDevices.includes(deviceType));
  }

  // 获取默认用户配置
  private getDefaultUserConfig(userId: string): UserMenuConfig {
    return {
      userId,
      deviceType: getDeviceInfo().type,
      userPermissions: [PermissionLevel.USER],
      lastUpdated: new Date().toISOString()
    };
  }

  // 清除用户缓存
  private clearUserCache(userId: string): void {
    const deviceTypes = Object.values(DeviceType);
    deviceTypes.forEach(deviceType => {
      const cacheKey = `${userId}_${deviceType}`;
      this.menuCache.delete(cacheKey);
    });
    this.userConfigCache.delete(userId);
  }

  // 清除所有缓存
  clearAllCache(): void {
    this.menuCache.clear();
    this.userConfigCache.clear();
  }
}

// 导出单例实例
export const menuService = new MenuService();