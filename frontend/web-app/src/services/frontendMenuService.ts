/**
 * 前端菜单渲染服务
 * 将后端菜单数据转换为前端组件可用的格式
 */

import React from 'react';
import { menuService, type MenuItem as BackendMenuItem } from './menuService';
import menuConfig, { type MenuItem as FrontendMenuItem } from '../config/menuConfig';
import {
  DashboardOutlined,
  UserOutlined,
  ProjectOutlined,
  DesktopOutlined,
  SettingOutlined,
  RobotOutlined,
  MessageOutlined,
  CameraOutlined,
  BarChartOutlined,
  EditOutlined,
  ApiOutlined,
  BulbOutlined,
  LockOutlined,
  BellOutlined,
  MenuOutlined,
} from '@ant-design/icons';
import { getIconNode, defaultIconName, isValidIconName } from '../ui/icons/iconRegistry';
import i18n from '../i18n';

// 规范化标签文本：去除空白，保证匹配稳定
function sanitizeLabel(s?: string): string | undefined {
  if (!s || typeof s !== 'string') return undefined;
  const t = s.trim();
  return t.length > 0 ? t : undefined;
}

export interface UserPermissions {
  roles: string[];
  permissions: string[];
}

class FrontendMenuService {
  private cache: Map<string, { data: any; timestamp: number }> = new Map();
  private readonly CACHE_DURATION = 5 * 60 * 1000; // 5分钟缓存

  // 基于静态菜单配置的路径/键到 labelKey 的索引，用于提升翻译命中率
  private readonly labelKeyByPath: Record<string, string> = {};
  private readonly labelKeyByKey: Record<string, string> = {};
  private readonly labelKeyByLabel: Record<string, string> = {};
  // 新增：中文标签到 labelKey 的索引（使用固定中文翻译构建），用于后端返回中文时的跨语言映射
  private readonly labelKeyByZhLabel: Record<string, string> = {};
  // 新增：针对若干未翻译菜单的手动中文映射
  private readonly manualLabelKeyMapZh: Record<string, string> = {
    '管理仪表板': 'adminDashboard.title',
    '管理员仪表板': 'adminDashboard.title',
    '帮助中心': 'header.help',
    '个人中心': 'mainMenu.labels.profile-overview',
    '推荐中心': 'mainMenu.labels.wisdom-recommend',
    '智慧推荐中心': 'mainMenu.labels.wisdom-recommend',
  };

  constructor() {
    // 建立索引：从静态菜单配置中收集所有 path 和 key 对应的 labelKey
    const indexItems = (items: FrontendMenuItem[]) => {
      items.forEach(item => {
        const lk = item.labelKey || `mainMenu.labels.${item.key}`;
        if (item.path) {
          this.labelKeyByPath[item.path] = lk;
        }
        if (item.key) {
          this.labelKeyByKey[String(item.key)] = lk;
        }
        // 记录静态配置中的显示文案到翻译键映射（用于无 path 的父级菜单）
        if (item.label && typeof item.label === 'string') {
          this.labelKeyByLabel[item.label] = lk;
        }
        // 中文标签索引：用固定中文翻译获取该键在中文下的显示文本
        try {
          const tZh = i18n.getFixedT('zh-CN');
          const zh = sanitizeLabel(tZh(lk, { defaultValue: item.label || '' }));
          if (zh) {
            this.labelKeyByZhLabel[zh] = lk;
          }
        } catch {}
        if (item.children && item.children.length > 0) {
          indexItems(item.children);
        }
      });
    };

    try {
      indexItems(menuConfig.mainMenuConfig || []);
      indexItems(menuConfig.profileMenuConfig || []);
      indexItems(menuConfig.quickActionsConfig || []);
    } catch (e) {
      // 若加载失败，不影响后续逻辑，只是无法使用索引增强翻译
      console.warn('菜单配置索引构建失败:', e);
    }
  }

  /**
   * 获取用户主菜单（转换为前端格式）
   */
  async getMainMenu(userPermissions?: UserPermissions): Promise<FrontendMenuItem[]> {
    try {
      console.log('🔄 FrontendMenuService - 开始获取主菜单');
      console.log('🔑 传入的用户权限:', userPermissions);
      
      // 尝试从缓存获取
      const cacheKey = `mainMenu_${JSON.stringify(userPermissions)}`;
      const cached = this.getFromCache(cacheKey);
      if (cached) {
        console.log('📦 从缓存获取菜单数据:', cached);
        return cached;
      }

      // 从后端获取菜单树
      console.log('🌐 从后端获取菜单数据...');
      const response = await menuService.getMenuTree();
      console.log('📥 后端返回的菜单数据:', response);
      
      // 处理后端返回的数据格式 {code: 200, data: [...]} 或 {items: [...]}
      const menuData = response.data || response.items || [];
      console.log('📋 解析后的菜单数据:', menuData);
      
      if (menuData && menuData.length > 0) {
        // 转换为前端格式
        console.log('🔄 转换为前端格式...');
        const frontendMenus = this.convertToFrontendFormat(menuData);
        console.log('✅ 转换后的前端菜单:', frontendMenus);
        
        // 根据权限过滤
        console.log('🔍 根据权限过滤菜单...');
        const filteredMenus = userPermissions 
          ? this.filterMenuByPermissions(frontendMenus, userPermissions)
          : frontendMenus;
        console.log('✅ 过滤后的菜单:', filteredMenus);

        // 缓存结果（严格使用后端数据，不做静态增强）
        this.setCache(cacheKey, filteredMenus);
        return filteredMenus;
      } else {
        console.warn('⚠️ 没有获取到菜单数据');
        throw new Error('没有获取到菜单数据');
      }
    } catch (error) {
      console.error('❌ 获取动态菜单失败:', error);
      return [];
    }
  }

  /**
   * 将后端菜单格式转换为前端格式
   */
  private convertToFrontendFormat(backendMenus: any[]): FrontendMenuItem[] {
    return backendMenus
      .filter(menu => {
        const visible = menu.is_visible !== false; // 默认为可见
        const enabled = menu.is_enabled !== false; // 默认为启用
        const activeByStatus = menu.status ? (menu.status !== 'inactive') : true; // 兼容已转换的后端item
        const activeFlag = (menu.isActive === undefined) ? true : !!menu.isActive; // 兼容 isActive 布尔标识
        return visible && enabled && activeByStatus && activeFlag;
      })
      .sort((a, b) => (b?.sort ?? 0) - (a?.sort ?? 0))
      .map(menu => this.convertMenuItem(menu));
  }

  /**
   * 转换单个菜单项
   */
  private convertMenuItem(backendMenu: any): FrontendMenuItem {
    console.log('🔄 转换菜单项:', {
      id: backendMenu.id,
      title: backendMenu.title,
      name: backendMenu.name,
      required_role: backendMenu.required_role,
      is_visible: backendMenu.is_visible,
      is_enabled: backendMenu.is_enabled
    });

    // 处理required_role字段，支持多种格式
    let requiredRole: string[] = [];
    if (backendMenu.required_role) {
      if (typeof backendMenu.required_role === 'string') {
        // 如果是字符串，分割成数组（支持逗号分隔）
        requiredRole = backendMenu.required_role.split(',').map((role: string) => role.trim().toLowerCase()).filter(Boolean);
      } else if (Array.isArray(backendMenu.required_role)) {
        // 如果已经是数组，直接使用
        requiredRole = backendMenu.required_role.map((role: string) => role.toLowerCase());
      }
    }

    // 计算 labelKey：优先使用后端下发的 label_key；其次使用静态配置映射与中文索引、手动映射；最后旧策略
    const backendKey = backendMenu.key || backendMenu.code || backendMenu.name || backendMenu.title;
    // 优先顺序：后端下发的 label_key -> 通过 path 映射到已知键 -> 通过 key 映射到已知键 -> 旧有回退策略
    let labelKey: string | undefined = backendMenu.label_key ? String(backendMenu.label_key) : undefined;
    if (!labelKey && backendMenu.path && this.labelKeyByPath[backendMenu.path]) {
      labelKey = this.labelKeyByPath[backendMenu.path];
    }
    if (!labelKey && backendKey && this.labelKeyByKey[String(backendKey)]) {
      labelKey = this.labelKeyByKey[String(backendKey)];
    }
    // 针对无 path 的目录型菜单，尝试通过标题/名称文本匹配
    const backendLabelText: string | undefined = (backendMenu.title || backendMenu.name);
    if (!labelKey && backendLabelText && this.labelKeyByLabel[backendLabelText]) {
      labelKey = this.labelKeyByLabel[backendLabelText];
    }
    // 中文标签匹配（后端通常返回中文标题/名称）
    if (!labelKey) {
      const zhCandidates = [backendMenu.title, backendMenu.name].map(v => sanitizeLabel(v)).filter(Boolean) as string[];
      for (const zh of zhCandidates) {
        if (this.labelKeyByZhLabel[zh]) {
          labelKey = this.labelKeyByZhLabel[zh];
          break;
        }
      }
    }
    // 手动中文映射，覆盖若干常见但不在静态配置中的菜单
    if (!labelKey) {
      const zhCandidates = [backendMenu.title, backendMenu.name].map(v => sanitizeLabel(v)).filter(Boolean) as string[];
      for (const zh of zhCandidates) {
        if (this.manualLabelKeyMapZh[zh]) {
          labelKey = this.manualLabelKeyMapZh[zh];
          break;
        }
      }
    }
    if (!labelKey && typeof backendKey === 'string') {
      labelKey = `mainMenu.labels.${backendKey}`;
    }

    const baseLabel = backendMenu.title || backendMenu.name;
    // 翻译优先；若结果与键相同则回退到原始标签
    const tt = labelKey ? i18n.t(labelKey) : undefined;
    const localizedLabel = tt && labelKey && tt !== labelKey ? tt : baseLabel;

    const frontendMenu: FrontendMenuItem = {
      key: backendMenu.id ?? backendKey ?? baseLabel,
      label: localizedLabel,
      labelKey: labelKey,
      path: backendMenu.path,
      icon: this.mapIcon(backendMenu.icon),
      status: 'completed', // 默认为已完成状态
      requiredRole: requiredRole,
      requiredPermission: [], // 暂时为空，后续可以根据需要添加
    };

    console.log('✅ 转换后的菜单项:', {
      key: frontendMenu.key,
      label: frontendMenu.label,
      requiredRole: frontendMenu.requiredRole
    });

    // 保持后端下发的路径不做统一重写，确保诸如 /admin/menus 等管理页正常导航

    // 处理子菜单
    if (backendMenu.children && backendMenu.children.length > 0) {
      frontendMenu.children = this.convertToFrontendFormat(backendMenu.children);
    }

    return frontendMenu;
  }

  /**
   * 增强动态菜单：若后端未提供“系统管理”下的静态新项，则补充
   * 目标项：数据库管理(/admin/database)、日志管理(/admin/logs)、问题跟踪(/admin/issues)
   */
  private augmentWithStaticAdminMenus(
    menus: FrontendMenuItem[],
    userPermissions: UserPermissions
  ): FrontendMenuItem[] {
    try {
      const staticMain = (menuConfig as any).mainMenuConfig as FrontendMenuItem[] | undefined;
      if (!staticMain || staticMain.length === 0) return menus;

      const staticAdmin = staticMain.find(it =>
        it.key === 'system-management' ||
        it.labelKey === 'mainMenu.labels.system-management' ||
        (typeof it.label === 'string' && it.label.includes('系统管理'))
      );
      if (!staticAdmin) return menus;

      const requiredKeys = new Set(['system-database', 'system-logs', 'system-issues']);
      const staticChildren = (staticAdmin.children || []).filter(ch => requiredKeys.has(String(ch.key)));
      if (staticChildren.length === 0) return menus;

      // 权限过滤静态子项，避免非管理员看到
      const eligibleStaticChildren = this.filterMenuByPermissions(staticChildren as FrontendMenuItem[], userPermissions);
      if (eligibleStaticChildren.length === 0) return menus;

      // 查找是否存在管理分组，以及是否已有对应路径
      const hasPath = (path: string): boolean => {
        const walk = (items: FrontendMenuItem[]): boolean => {
          for (const it of items) {
            if (it.path === path) return true;
            if (it.children && walk(it.children)) return true;
          }
          return false;
        };
        return walk(menus);
      };

      const findAdminGroup = (items: FrontendMenuItem[]): FrontendMenuItem | null => {
        for (const it of items) {
          const isAdminGroup = (it.path === '/admin') ||
            (it.labelKey === 'mainMenu.labels.system-management') ||
            (typeof it.label === 'string' && it.label.includes('系统管理'));
          if (isAdminGroup) return it;
          if (it.children) {
            const found = findAdminGroup(it.children);
            if (found) return found;
          }
        }
        return null;
      };

      const adminGroup = findAdminGroup(menus);
      const missingChildren = eligibleStaticChildren.filter(ch => ch.path && !hasPath(ch.path));
      if (missingChildren.length === 0) return menus;

      if (!adminGroup) {
        // 创建一个管理分组并追加缺失子项
        const newAdminGroup: FrontendMenuItem = {
          key: staticAdmin.key,
          label: staticAdmin.label,
          labelKey: staticAdmin.labelKey || 'mainMenu.labels.system-management',
          path: '/admin',
          icon: staticAdmin.icon,
          status: staticAdmin.status,
          requiredRole: staticAdmin.requiredRole,
          requiredPermission: staticAdmin.requiredPermission,
          children: missingChildren as FrontendMenuItem[]
        };
        return [...menus, newAdminGroup];
      }

      adminGroup.children = [...(adminGroup.children || []), ...(missingChildren as FrontendMenuItem[])];
      return menus;
    } catch (e) {
      console.warn('augmentWithStaticAdminMenus 失败:', e);
      return menus;
    }
  }

  /**
   * 将后端字符串图标名称映射为 Ant Design 图标节点
   */
  private mapIcon(iconName?: string): React.ReactNode | undefined {
    // 统一使用图标注册中心，并为未配置图标提供默认回退
    const name = (iconName && typeof iconName === 'string' && isValidIconName(iconName))
      ? iconName.trim()
      : defaultIconName;
    return getIconNode(name);
  }

  /**
   * 从权限中提取角色
   */
  private extractRoles(permissions: any[]): string[] {
    // 这里可以根据实际的权限结构来提取角色
    // 目前简化处理，可以根据权限代码推断角色
    const roles: string[] = [];
    
    permissions.forEach(permission => {
      if (permission.code.includes('manage')) {
        roles.push('admin');
      } else if (permission.code.includes('view')) {
        roles.push('user');
      }
    });

    return [...new Set(roles)]; // 去重
  }

  /**
   * 从权限中提取权限代码
   */
  private extractPermissions(permissions: any[]): string[] {
    return permissions.map(permission => permission.code);
  }

  /**
   * 根据用户权限过滤菜单
   */
  private filterMenuByPermissions(
    menuItems: FrontendMenuItem[],
    userPermissions: UserPermissions
  ): FrontendMenuItem[] {
    console.log('🔍 根据权限过滤菜单...');
    console.log('👤 用户权限信息:', userPermissions);
    console.log('📋 待过滤的菜单项数量:', menuItems.length);
    
    const filterItem = (item: FrontendMenuItem): FrontendMenuItem | null => {
      console.log(`🔍 检查菜单项: ${item.label}`, {
        requiredRole: item.requiredRole,
        requiredPermission: item.requiredPermission,
        userRoles: userPermissions.roles,
        userPermissions: userPermissions.permissions
      });

      // 如果用户是admin或administrator，允许访问所有菜单
      const isAdmin = userPermissions.roles?.some(role => 
        ['admin', 'administrator', 'super_admin'].includes(role.toLowerCase())
      );
      
      if (isAdmin) {
        console.log(`✅ ${item.label}: 管理员权限，允许访问`);
        // 递归处理子菜单，确保子菜单也被正确处理
        if (item.children && item.children.length > 0) {
          const processedChildren = item.children
            .map(child => filterItem(child))
            .filter((child): child is FrontendMenuItem => child !== null);
          
          return {
            ...item,
            children: processedChildren
          };
        }
        return item;
      }

      // 检查角色权限
      if (item.requiredRole && item.requiredRole.length > 0) {
        console.log(`🔍 ${item.label}: 开始角色检查`, {
          requiredRole: item.requiredRole,
          userRoles: userPermissions.roles
        });

        const hasRole = item.requiredRole.some(role => {
          const includes = userPermissions.roles?.some(userRole => 
            userRole.toLowerCase() === role.toLowerCase()
          );
          console.log(`🔍 检查角色 "${role}":`, {
            role: role,
            userRoles: userPermissions.roles,
            includes
          });
          return includes;
        });

        console.log(`🔑 ${item.label}: 角色检查结果`, {
          required: item.requiredRole,
          userRoles: userPermissions.roles,
          hasRole
        });
        if (!hasRole) {
          console.log(`❌ ${item.label}: 角色权限不足，拒绝访问`);
          return null;
        }
      } else {
        console.log(`✅ ${item.label}: 无角色要求，允许访问`);
      }

      // 检查具体权限
      if (item.requiredPermission && item.requiredPermission.length > 0) {
        const hasPermission = item.requiredPermission.some(permission => 
          userPermissions.permissions?.includes(permission)
        );
        console.log(`🔐 ${item.label}: 权限检查`, {
          required: item.requiredPermission,
          userPermissions: userPermissions.permissions,
          hasPermission
        });
        if (!hasPermission) {
          console.log(`❌ ${item.label}: 具体权限不足，拒绝访问`);
          return null;
        }
      } else {
        console.log(`✅ ${item.label}: 无具体权限要求，允许访问`);
      }

      // 递归过滤子菜单
      if (item.children && item.children.length > 0) {
        const filteredChildren = item.children
          .map(child => filterItem(child))
          .filter((child): child is FrontendMenuItem => child !== null);
        
        console.log(`👶 ${item.label}: 子菜单过滤结果`, {
          原始数量: item.children.length,
          过滤后数量: filteredChildren.length
        });

        return {
          ...item,
          children: filteredChildren
        };
      }

      console.log(`✅ ${item.label}: 通过所有权限检查`);
      return item;
    };

    const filteredMenu = menuItems
      .map(item => filterItem(item))
      .filter((item): item is FrontendMenuItem => item !== null);

    console.log('✅ 最终过滤后的菜单:', filteredMenu);
    console.log('📊 过滤统计:', {
      原始菜单数量: menuItems.length,
      过滤后菜单数量: filteredMenu.length
    });
    
    return filteredMenu;
  }

  /**
   * 获取默认菜单配置（fallback）
   */
  private getDefaultMenu(): FrontendMenuItem[] {
    console.warn('当前没有获取到菜单数据，使用临时测试菜单');
    
    // 返回临时测试菜单数据（带 labelKey 并应用 i18n）
    const items: FrontendMenuItem[] = [
      {
        key: 'dashboard',
        label: '仪表板',
        labelKey: 'mainMenu.labels.dashboard',
        path: '/dashboard',
        icon: 'DashboardOutlined',
        status: 'completed',
        requiredRole: ['user', 'admin'],
        requiredPermission: []
      },
      {
        key: 'profile',
        label: '个人资料',
        labelKey: 'mainMenu.labels.profile-overview',
        path: '/profile',
        icon: 'UserOutlined',
        status: 'completed',
        requiredRole: ['user', 'admin'],
        requiredPermission: []
      },
      {
        key: 'projects',
        label: '项目管理',
        labelKey: 'mainMenu.labels.project-management',
        path: '/projects',
        icon: 'ProjectOutlined',
        status: 'completed',
        requiredRole: ['user', 'admin'],
        requiredPermission: [],
        children: [
          {
            key: 'projects-workspace',
            label: '项目工作台',
            labelKey: 'mainMenu.labels.project-workspace',
            path: '/projects/workspace',
            icon: 'DesktopOutlined',
            status: 'completed',
            requiredRole: ['user', 'admin'],
            requiredPermission: []
          },
          {
            key: 'projects-management',
            label: '项目管理',
            labelKey: 'mainMenu.labels.project-management',
            path: '/projects/management',
            icon: 'SettingOutlined',
            status: 'completed',
            requiredRole: ['admin'],
            requiredPermission: []
          }
        ]
      },
      {
        key: 'learning',
        label: '学习中心',
        labelKey: 'mainMenu.labels.intelligent-learning',
        path: '/learning',
        icon: 'BookOutlined',
        status: 'completed',
        requiredRole: ['user', 'admin'],
        requiredPermission: [],
        children: [
          {
            key: 'learning-courses',
            label: '课程中心',
            labelKey: 'mainMenu.labels.learning-courses',
            path: '/learning/courses',
            icon: 'ReadOutlined',
            status: 'completed',
            requiredRole: ['user', 'admin'],
            requiredPermission: []
          },
          {
            key: 'learning-progress',
            label: '学习进度',
            labelKey: 'mainMenu.labels.learning-progress',
            path: '/learning/progress',
            icon: 'BarChartOutlined',
            status: 'completed',
            requiredRole: ['user', 'admin'],
            requiredPermission: []
          }
        ]
      },
      {
        key: 'admin',
        label: '系统管理',
        labelKey: 'mainMenu.labels.system-management',
        path: '/admin',
        icon: 'SettingOutlined',
        status: 'completed',
        requiredRole: ['admin'],
        requiredPermission: [],
        children: [
          {
            key: 'admin-users',
            label: '用户管理',
            labelKey: 'mainMenu.labels.system-users',
            path: '/admin/users',
            icon: 'TeamOutlined',
            status: 'completed',
            requiredRole: ['admin'],
            requiredPermission: []
          },
          {
            key: 'admin-settings',
            label: '系统设置',
            labelKey: 'mainMenu.labels.system-settings',
            path: '/admin/settings',
            icon: 'ControlOutlined',
            status: 'completed',
            requiredRole: ['admin'],
            requiredPermission: []
          }
        ]
      }
    ];

    const applyLocalization = (arr: FrontendMenuItem[]) => {
      arr.forEach(item => {
        const key = item.labelKey || `mainMenu.labels.${item.key}`;
        item.labelKey = key;
        item.label = i18n.t(key);
        if (item.children && item.children.length > 0) {
          applyLocalization(item.children);
        }
      });
    };

    applyLocalization(items);
    return items;
  }

  /**
   * 缓存管理
   */
  private getFromCache(key: string): any {
    const cached = this.cache.get(key);
    if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
      return cached.data;
    }
    return null;
  }

  private setCache(key: string, data: any): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now()
    });
  }

  /**
   * 清除缓存
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * 刷新菜单数据
   */
  async refreshMenu(userPermissions?: UserPermissions): Promise<FrontendMenuItem[]> {
    this.clearCache();
    return this.getMainMenu(userPermissions);
  }

  /**
   * 根据路径查找菜单项
   */
  findMenuByPath(menuItems: FrontendMenuItem[], path: string): FrontendMenuItem | null {
    for (const item of menuItems) {
      if (item.path === path) {
        return item;
      }
      if (item.children) {
        const found = this.findMenuByPath(item.children, path);
        if (found) return found;
      }
    }
    return null;
  }

  /**
   * 获取面包屑导航
   */
  getBreadcrumb(menuItems: FrontendMenuItem[], currentPath: string): FrontendMenuItem[] {
    const breadcrumb: FrontendMenuItem[] = [];
    
    const findPath = (items: FrontendMenuItem[], path: string, parents: FrontendMenuItem[] = []): boolean => {
      for (const item of items) {
        const currentParents = [...parents, item];
        
        if (item.path === path) {
          breadcrumb.push(...currentParents);
          return true;
        }
        
        if (item.children && findPath(item.children, path, currentParents)) {
          return true;
        }
      }
      return false;
    };

    findPath(menuItems, currentPath);
    return breadcrumb;
  }
}

// 导出单例实例
export const frontendMenuService = new FrontendMenuService();
export default frontendMenuService;