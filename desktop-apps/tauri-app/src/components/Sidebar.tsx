import { NavLink } from 'react-router-dom';
import { 
  MessageCircle, 
  FileText, 
  Image, 
  Settings,
  Bot
} from 'lucide-react';
import { cn } from '../utils/cn';

interface SidebarProps {
  collapsed: boolean;
}

const navigation = [
  { name: 'AI对话', href: '/chat', icon: MessageCircle },
  { name: '文档处理', href: '/document', icon: FileText },
  { name: '图像生成', href: '/image', icon: Image },
  { name: '设置', href: '/settings', icon: Settings },
];

export default function Sidebar({ collapsed }: SidebarProps) {
  return (
    <div className={cn(
      'sidebar transition-all duration-300 ease-in-out',
      collapsed ? 'w-16' : 'w-64'
    )}>
      {/* Logo */}
      <div className="flex items-center h-16 px-4 border-b border-border">
        <Bot className="h-8 w-8 text-primary" />
        {!collapsed && (
          <span className="ml-3 text-lg font-semibold text-foreground">
            太上老君
          </span>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-2 py-4 space-y-2">
        {navigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            className={({ isActive }) =>
              cn(
                'group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors',
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              )
            }
          >
            <item.icon className="h-5 w-5 flex-shrink-0" />
            {!collapsed && (
              <span className="ml-3">{item.name}</span>
            )}
          </NavLink>
        ))}
      </nav>

      {/* User info */}
      <div className="border-t border-border p-4">
        <div className="flex items-center">
          <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center">
            <span className="text-sm font-medium text-primary-foreground">
              用
            </span>
          </div>
          {!collapsed && (
            <div className="ml-3">
              <p className="text-sm font-medium text-foreground">用户</p>
              <p className="text-xs text-muted-foreground">在线</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}