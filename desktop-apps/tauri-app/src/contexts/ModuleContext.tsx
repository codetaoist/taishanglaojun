import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { apiService, AppModule } from '../services/api';
import { useAuth } from './AuthContext';

interface ModuleContextType {
  modules: AppModule[];
  userPermissions: Record<string, boolean>;
  loading: boolean;
  refreshModules: () => Promise<void>;
  hasPermission: (moduleId: string) => boolean;
  getModulesByCategory: (category: string) => AppModule[];
}

const ModuleContext = createContext<ModuleContextType | undefined>(undefined);

export function useModules() {
  const context = useContext(ModuleContext);
  if (context === undefined) {
    throw new Error('useModules must be used within a ModuleProvider');
  }
  return context;
}

interface ModuleProviderProps {
  children: ReactNode;
}

export function ModuleProvider({ children }: ModuleProviderProps) {
  const [modules, setModules] = useState<AppModule[]>([]);
  const [userPermissions, setUserPermissions] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const { isAuthenticated, user } = useAuth();

  // 默认模块配置
  const getDefaultModules = (): AppModule[] => [
    {
      id: 'chat',
      name: 'AI对话',
      description: '与AI进行智能对话',
      category: 'ai',
      required_role: 'USER',
      is_active: true,
      icon: 'MessageCircle',
      route: '/chat'
    },
    {
      id: 'document',
      name: '文档处理',
      description: '文档编辑和处理功能',
      category: 'productivity',
      required_role: 'USER',
      is_active: true,
      icon: 'FileText',
      route: '/document'
    },
    {
      id: 'file-transfer',
      name: '文件传输',
      description: '文件上传和下载',
      category: 'utility',
      required_role: 'USER',
      is_active: true,
      icon: 'Share2',
      route: '/file-transfer'
    },
    {
      id: 'image',
      name: '图像生成',
      description: 'AI图像生成和编辑',
      category: 'ai',
      required_role: 'USER',
      is_active: true,
      icon: 'Image',
      route: '/image'
    },
    {
      id: 'pet',
      name: '桌面宠物',
      description: '可爱的桌面宠物助手',
      category: 'entertainment',
      required_role: 'USER',
      is_active: true,
      icon: 'Heart',
      route: '/pet'
    },
    {
      id: 'system',
      name: '系统监控',
      description: '系统性能监控',
      category: 'system',
      required_role: 'ADMIN',
      is_active: true,
      icon: 'Monitor',
      route: '/system'
    },
    {
      id: 'friend-management',
      name: '好友管理',
      description: '管理好友和联系人',
      category: 'social',
      required_role: 'USER',
      is_active: true,
      icon: 'Users',
      route: '/friends'
    },
    {
      id: 'project-management',
      name: '项目管理',
      description: '项目和任务管理',
      category: 'productivity',
      required_role: 'USER',
      is_active: true,
      icon: 'FolderOpen',
      route: '/projects'
    },
    {
      id: 'settings',
      name: '设置',
      description: '应用设置和配置',
      category: 'system',
      required_role: 'USER',
      is_active: true,
      icon: 'Settings',
      route: '/settings'
    }
  ];

  const getDefaultPermissions = (modules: AppModule[]): Record<string, boolean> => {
    const permissions: Record<string, boolean> = {};
    modules.forEach(module => {
      permissions[module.id] = true;
    });
    return permissions;
  };

  useEffect(() => {
    if (isAuthenticated && user) {
      refreshModules();
    } else {
      // Clear modules when user is not authenticated
      setModules([]);
      setUserPermissions({});
    }
  }, [isAuthenticated, user]);

  const refreshModules = async () => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      
      // 尝试从API获取用户模块
      try {
        const userModules = await apiService.getUserModules();
        
        if (userModules) {
          setModules(userModules.modules);
          setUserPermissions(userModules.user_permissions);
          return;
        }
      } catch (apiError) {
        console.warn('API call failed, using default modules:', apiError);
      }
      
      // 如果API调用失败，使用默认模块配置
      const defaultModules = getDefaultModules();
      const defaultPermissions = getDefaultPermissions(defaultModules);
      
      setModules(defaultModules);
      setUserPermissions(defaultPermissions);
      
    } catch (error) {
      console.error('Failed to refresh modules:', error);
      
      // 即使出错也提供默认模块
      const defaultModules = getDefaultModules();
      const defaultPermissions = getDefaultPermissions(defaultModules);
      
      setModules(defaultModules);
      setUserPermissions(defaultPermissions);
    } finally {
      setLoading(false);
    }
  };

  const hasPermission = (moduleId: string): boolean => {
    // 如果 userPermissions 为空或者 moduleId 不存在，返回 false
    if (!userPermissions || typeof userPermissions !== 'object') {
      return false;
    }
    return userPermissions[moduleId] === true;
  };

  const getModulesByCategory = (category: string): AppModule[] => {
    return modules.filter(module => module.category === category && hasPermission(module.id));
  };

  const value: ModuleContextType = {
    modules,
    userPermissions,
    loading,
    refreshModules,
    hasPermission,
    getModulesByCategory,
  };

  return (
    <ModuleContext.Provider value={value}>
      {children}
    </ModuleContext.Provider>
  );
}