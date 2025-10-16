import { useState, useEffect } from 'react';
import { Package, Eye } from 'lucide-react';
import { useModules } from '../contexts/ModuleContext';
import { ApiService } from '../services/api';

interface ModulePermission {
  module_name: string;
  permission_type: string;
  granted: boolean;
}

interface UserModulePreferences {
  user_id: number;
  module_permissions: ModulePermission[];
  ui_preferences: {
    theme: string;
    language: string;
    sidebar_collapsed: boolean;
  };
}

export default function AppManagementPage() {
  const { modules, hasPermission, refreshModules } = useModules();
  const [preferences, setPreferences] = useState<UserModulePreferences | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const apiService = new ApiService();

  useEffect(() => {
    loadUserPreferences();
  }, []);

  const loadUserPreferences = async () => {
    try {
      setLoading(true);
      const userPrefs = await apiService.getUserPreferences();
      setPreferences(userPrefs);
    } catch (error) {
      console.error('Failed to load user preferences:', error);
    } finally {
      setLoading(false);
    }
  };

  const updateModulePermission = async (moduleName: string, permissionType: string, granted: boolean) => {
    if (!preferences) return;

    try {
      setSaving(true);
      
      const updatedPermissions = preferences.module_permissions.map(perm => 
        perm.module_name === moduleName && perm.permission_type === permissionType
          ? { ...perm, granted }
          : perm
      );

      const updatedPreferences = {
        ...preferences,
        module_permissions: updatedPermissions
      };

      await apiService.updateUserPreferences(updatedPreferences);
      setPreferences(updatedPreferences);
      await refreshModules();
    } catch (error) {
      console.error('Failed to update module permission:', error);
    } finally {
      setSaving(false);
    }
  };



  const categories = ['all', ...new Set(modules.map(module => module.category))];
  const filteredModules = selectedCategory === 'all' 
    ? modules 
    : modules.filter(module => module.category === selectedCategory);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-foreground mb-2">应用模块管理</h1>
        <p className="text-muted-foreground">管理您的应用模块权限和设置</p>
      </div>

      {/* 分类筛选 */}
      <div className="mb-6">
        <div className="flex flex-wrap gap-2">
          {categories.map(category => (
            <button
              key={category}
              onClick={() => setSelectedCategory(category)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                selectedCategory === category
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
              }`}
            >
              {category === 'all' ? '全部' : category}
            </button>
          ))}
        </div>
      </div>

      {/* 模块列表 */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {filteredModules.map(module => {
          const hasModulePermission = hasPermission(module.id);

          return (
            <div key={module.id} className="bg-card rounded-lg border border-border p-6">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center">
                  <Package className="h-8 w-8 text-primary mr-3" />
                  <div>
                    <h3 className="font-semibold text-foreground">{module.name}</h3>
                    <p className="text-sm text-muted-foreground">{module.category}</p>
                  </div>
                </div>
                <div className={`px-2 py-1 rounded text-xs font-medium ${
                  module.is_active 
                    ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                    : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                }`}>
                  {module.is_active ? '活跃' : '非活跃'}
                </div>
              </div>

              <p className="text-sm text-muted-foreground mb-4">{module.description}</p>

              {/* 权限控制 */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <Eye className="h-4 w-4 text-muted-foreground mr-2" />
                    <span className="text-sm">模块权限</span>
                  </div>
                  <button
                    onClick={() => updateModulePermission(module.name, 'access', !hasModulePermission)}
                    disabled={saving}
                    className={`w-10 h-6 rounded-full transition-colors ${
                      hasModulePermission 
                        ? 'bg-primary' 
                        : 'bg-gray-300 dark:bg-gray-600'
                    } ${saving ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
                  >
                    <div className={`w-4 h-4 bg-white rounded-full transition-transform ${
                      hasModulePermission ? 'translate-x-5' : 'translate-x-1'
                    }`} />
                  </button>
                </div>
              </div>

              {/* 模块信息 */}
              <div className="mt-4 pt-4 border-t border-border">
                <div className="text-xs text-muted-foreground space-y-1">
                  <div>类别: {module.category}</div>
                  <div>权限要求: {module.required_role}</div>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {filteredModules.length === 0 && (
        <div className="text-center py-12">
          <Package className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-medium text-foreground mb-2">没有找到模块</h3>
          <p className="text-muted-foreground">当前分类下没有可用的模块</p>
        </div>
      )}
    </div>
  );
}