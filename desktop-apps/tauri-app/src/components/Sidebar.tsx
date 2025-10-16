import { NavLink } from 'react-router-dom';
import { 
  MessageCircle, 
  FileText, 
  Image, 
  Settings,
  Bot,
  TestTube,
  Share2,
  Monitor,
  Heart,
  X,
  Shield,
  Users,
  FolderOpen,
  Package,
  MessageSquare,
  LucideIcon
} from 'lucide-react';
import { cn } from '../utils/cn';
import { useModules } from '../contexts/ModuleContext';

interface SidebarProps {
  collapsed: boolean;
  onClose?: () => void;
}

// 图标映射函数
const getIconForModule = (moduleName: string): LucideIcon => {
  const iconMap: Record<string, LucideIcon> = {
    'AI对话': MessageCircle,
    '文档处理': FileText,
    '文件传输': Share2,
    '图像生成': Image,
    '桌面宠物': Heart,
    '系统监控': Monitor,
    '连接测试': TestTube,
    '用户认证': Shield,
    '好友管理': Users,
    '项目管理': FolderOpen,
    '应用管理': Package,
    '聊天管理': MessageSquare,
    '设置': Settings,
  };

  return iconMap[moduleName] || Bot;
};

// 路由映射函数
const getRouteForModule = (moduleName: string): string => {
  const routeMap: Record<string, string> = {
    'AI对话': '/chat',
    '文档处理': '/document',
    '文件传输': '/file-transfer',
    '图像生成': '/image',
    '桌面宠物': '/pet',
    '系统监控': '/system',
    '连接测试': '/test',
    '用户认证': '/auth-management',
    '好友管理': '/friend-management',
    '项目管理': '/project-management',
    '应用管理': '/app-management',
    '聊天管理': '/chat-management',
    '设置': '/settings',
  };

  return routeMap[moduleName] || '/';
};

export default function Sidebar({ collapsed, onClose }: SidebarProps) {
  const { modules, hasPermission } = useModules();

  // 过滤用户有权限的模块
  const availableModules = (modules || []).filter(module => 
    module && module.id && hasPermission(module.id)
  );

  return (
    <div className={cn(
      'sidebar transition-all duration-300 ease-in-out bg-gradient-to-b from-background to-background/95 border-r border-border/50 backdrop-blur-sm',
      collapsed ? 'w-16' : 'w-64'
    )}>
      {/* Logo */}
      <div className="flex items-center justify-between h-16 border-b border-border/50 px-4 bg-gradient-to-r from-primary/5 to-primary/10">
        <div className="flex items-center">
          <div className="relative">
            <Bot className="h-8 w-8 text-primary transition-transform duration-300 hover:scale-110" />
            <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
          </div>
          {!collapsed && (
            <span className="ml-3 text-xl font-bold text-foreground bg-gradient-to-r from-primary to-primary/80 bg-clip-text text-transparent">
              太上老君
            </span>
          )}
        </div>
        {/* 移动端关闭按钮 */}
        {onClose && (
          <button
            onClick={onClose}
            className="lg:hidden p-2 rounded-lg text-gray-600 hover:text-gray-900 hover:bg-white/80 transition-all duration-200 hover:scale-110"
          >
            <X className="h-5 w-5" />
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-6 space-y-1">
        {availableModules.map((module, index) => {
          const IconComponent = getIconForModule(module.name);
          const route = getRouteForModule(module.name);
          
          return (
            <NavLink
              key={module.id}
              to={route}
              className={({ isActive }) =>
                cn(
                  'group flex items-center px-3 py-3 text-sm font-medium rounded-xl transition-all duration-200 relative overflow-hidden',
                  'hover:scale-105 hover:shadow-md transform',
                  isActive
                    ? 'bg-gradient-to-r from-primary to-primary/90 text-primary-foreground shadow-lg scale-105'
                    : 'text-muted-foreground hover:bg-gradient-to-r hover:from-accent/50 hover:to-accent/30 hover:text-accent-foreground'
                )
              }
              style={{
                animationDelay: `${index * 50}ms`
              }}
            >
              <div className="relative">
                <IconComponent className={cn(
                  "h-5 w-5 flex-shrink-0 transition-all duration-200",
                  "group-hover:scale-110"
                )} />
                {!collapsed && (
                  <div className="absolute -top-1 -right-1 w-2 h-2 bg-primary/60 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200"></div>
                )}
              </div>
              {!collapsed && (
                <span className="ml-3 transition-all duration-200 group-hover:translate-x-1">
                  {module.name}
                </span>
              )}
              {/* 活跃状态指示器 */}
              <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-8 bg-primary-foreground rounded-r-full opacity-0 group-[.active]:opacity-100 transition-opacity duration-200"></div>
            </NavLink>
          );
        })}
      </nav>

      {/* User info */}
      <div className="border-t border-border/50 p-4 bg-gradient-to-r from-primary/5 to-primary/10">
        <div className="flex items-center group cursor-pointer hover:bg-white/10 rounded-xl p-2 transition-all duration-200">
          <div className="relative">
            <div className="h-10 w-10 rounded-full bg-gradient-to-br from-primary to-primary/80 flex items-center justify-center shadow-lg transition-transform duration-200 group-hover:scale-110">
              <span className="text-sm font-bold text-primary-foreground">
                用
              </span>
            </div>
            <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-green-500 rounded-full border-2 border-background animate-pulse"></div>
          </div>
          {!collapsed && (
            <div className="ml-3 transition-all duration-200 group-hover:translate-x-1">
              <p className="text-sm font-semibold text-foreground">用户</p>
              <div className="flex items-center space-x-1">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <p className="text-xs text-muted-foreground">在线</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}