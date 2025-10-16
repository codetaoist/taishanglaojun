// 设备类型枚举
export enum DeviceType {
  MOBILE = 'mobile',
  DESKTOP = 'desktop',
  WATCH = 'watch',
  TABLET = 'tablet'
}

// 权限级别枚举
export enum PermissionLevel {
  GUEST = 'guest',
  USER = 'user',
  MODERATOR = 'moderator',
  ADMIN = 'admin',
  SUPER_ADMIN = 'super_admin'
}

// 菜单项类型
export enum MenuItemType {
  PAGE = 'page',
  ACTION = 'action',
  SUBMENU = 'submenu',
  SEPARATOR = 'separator'
}

// 菜单项配置接口
export interface MenuItem {
  id: string;
  title: string;
  description?: string;
  icon?: string;
  type: MenuItemType;
  route?: string;
  action?: string;
  children?: MenuItem[];
  
  // 权限控制
  requiredPermissions: PermissionLevel[];
  excludedPermissions?: PermissionLevel[];
  
  // 设备适配
  supportedDevices: DeviceType[];
  deviceSpecificConfig?: {
    [key in DeviceType]?: {
      title?: string;
      icon?: string;
      hidden?: boolean;
      order?: number;
    };
  };
  
  // 显示控制
  order: number;
  isActive: boolean;
  isVisible: boolean;
  
  // 元数据
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

// 用户菜单配置
export interface UserMenuConfig {
  userId: string;
  deviceType: DeviceType;
  userPermissions: PermissionLevel[];
  customMenuItems?: MenuItem[];
  hiddenMenuItems?: string[]; // 菜单项ID列表
  menuOrder?: string[]; // 自定义菜单顺序
  lastUpdated: string;
}

// 菜单响应接口
export interface MenuResponse {
  success: boolean;
  data: {
    menuItems: MenuItem[];
    userConfig: UserMenuConfig;
    deviceInfo: {
      type: DeviceType;
      screenSize: {
        width: number;
        height: number;
      };
      capabilities: string[];
    };
  };
  message?: string;
}

// 菜单更新请求
export interface MenuUpdateRequest {
  userId: string;
  menuItems?: MenuItem[];
  userConfig?: Partial<UserMenuConfig>;
  adminAction?: {
    type: 'grant_permission' | 'revoke_permission' | 'update_menu' | 'reset_menu';
    targetUserId?: string;
    permissions?: PermissionLevel[];
    menuItemIds?: string[];
  };
}

// 批量权限管理
export interface BatchPermissionUpdate {
  userIds: string[];
  permissions: {
    grant?: PermissionLevel[];
    revoke?: PermissionLevel[];
  };
  menuItems?: {
    show?: string[];
    hide?: string[];
  };
  effectiveDate?: string;
  expiryDate?: string;
}

// 菜单统计信息
export interface MenuAnalytics {
  totalMenuItems: number;
  activeUsers: number;
  deviceDistribution: {
    [key in DeviceType]: number;
  };
  popularMenuItems: {
    id: string;
    title: string;
    clickCount: number;
    deviceBreakdown: {
      [key in DeviceType]: number;
    };
  }[];
  permissionDistribution: {
    [key in PermissionLevel]: number;
  };
}