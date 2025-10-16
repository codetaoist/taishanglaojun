import React from 'react';
import { 
  Menu, 
  Sun, 
  Moon, 
  Monitor, 
  LogOut,
  Crown
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useAuth } from '../contexts/AuthContext';
import { cn } from '../utils/cn';

interface HeaderProps {
  onToggleSidebar: () => void;
  sidebarCollapsed: boolean;
}

export default function Header({ onToggleSidebar }: HeaderProps) {
  const { theme, setTheme, actualTheme } = useTheme();
  const { user, logout, isPremium, togglePremium } = useAuth();
  const [showUserMenu, setShowUserMenu] = React.useState(false);
  const [showThemeMenu, setShowThemeMenu] = React.useState(false);
  const [showUpgradeConfirm, setShowUpgradeConfirm] = React.useState(false);

  const themeOptions = [
    { value: 'light', label: '浅色', icon: Sun },
    { value: 'dark', label: '深色', icon: Moon },
    { value: 'system', label: '跟随系统', icon: Monitor },
  ];

  const handleLogout = async () => {
    await logout();
    setShowUserMenu(false);
  };

  const handleUpgradeClick = () => {
    setShowUpgradeConfirm(true);
    setShowUserMenu(false);
  };

  const handleUpgradeConfirm = () => {
    togglePremium();
    setShowUpgradeConfirm(false);
  };

  const handleUpgradeCancel = () => {
    setShowUpgradeConfirm(false);
  };

  return (
    <header className="h-16 border-b border-border/50 bg-gradient-to-r from-background/95 to-background/90 backdrop-blur-md supports-[backdrop-filter]:bg-background/60 shadow-sm relative z-40">
      <div className="flex items-center justify-between h-full px-6 relative z-10">
        {/* Left side */}
        <div className="flex items-center space-x-6">
          <button
            onClick={onToggleSidebar}
            className="p-2.5 rounded-xl hover:bg-accent/80 hover:text-accent-foreground transition-all duration-200 hover:scale-110 active:scale-95 shadow-sm hover:shadow-md"
          >
            <Menu className="h-5 w-5" />
          </button>
          
          <div className="flex items-center space-x-3">
            <div className="relative">
              <img 
                src="/src/assets/laojun-avatar.svg" 
                alt="老君头像" 
                className="w-8 h-8 rounded-full shadow-md transition-transform duration-300 hover:scale-110"
              />
              <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-green-500 rounded-full border-2 border-background animate-pulse"></div>
            </div>
            <div className="text-sm font-medium text-muted-foreground bg-gradient-to-r from-primary/60 to-primary/40 bg-clip-text text-transparent transition-all duration-500">
              欢迎使用{isPremium ? '太上' : '老君'}AI助手
            </div>
            {isPremium && (
              <div className="px-2 py-1 bg-gradient-to-r from-yellow-400/20 to-orange-400/20 border border-yellow-400/30 rounded-full text-xs font-medium text-yellow-600 dark:text-yellow-400 animate-pulse">
                高级版
              </div>
            )}
          </div>
        </div>

        {/* Right side */}
        <div className="flex items-center space-x-3 relative z-20">
          {/* Theme toggle */}
          <div className="relative">
            <button
              onClick={() => setShowThemeMenu(!showThemeMenu)}
              className="p-2.5 rounded-xl hover:bg-accent/80 hover:text-accent-foreground transition-all duration-200 hover:scale-110 active:scale-95 shadow-sm hover:shadow-md relative overflow-hidden"
            >
              <div className="relative z-10">
                {actualTheme === 'dark' ? (
                  <Moon className="h-5 w-5" />
                ) : (
                  <Sun className="h-5 w-5" />
                )}
              </div>
              <div className="absolute inset-0 bg-gradient-to-r from-yellow-400/20 to-orange-400/20 opacity-0 hover:opacity-100 transition-opacity duration-200"></div>
            </button>
            
            {showThemeMenu && (
              <div className="absolute right-0 mt-3 w-52 bg-popover/95 backdrop-blur-md border border-border/50 rounded-xl shadow-xl z-50 animate-in slide-in-from-top-2 duration-200">
                <div className="py-2">
                  {themeOptions.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => {
                        setTheme(option.value as any);
                        setShowThemeMenu(false);
                      }}
                      className={cn(
                        'flex items-center w-full px-4 py-3 text-sm hover:bg-accent/50 hover:text-accent-foreground transition-all duration-200 hover:scale-105 transform',
                        theme === option.value && 'bg-gradient-to-r from-primary/20 to-primary/10 text-primary font-medium'
                      )}
                    >
                      <option.icon className="h-4 w-4 mr-3" />
                      {option.label}
                      {theme === option.value && (
                        <div className="ml-auto w-2 h-2 bg-primary rounded-full"></div>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* User menu */}
          <div className="relative">
            <button
              onClick={() => setShowUserMenu(!showUserMenu)}
              className="flex items-center space-x-3 p-2 rounded-xl hover:bg-accent/80 hover:text-accent-foreground transition-all duration-200 hover:scale-105 active:scale-95 shadow-sm hover:shadow-md group"
            >
              <div className="relative">
                <div className="h-8 w-8 rounded-full overflow-hidden shadow-md transition-transform duration-200 group-hover:scale-110 border-2 border-primary/20">
                  <img 
                    src="/src/assets/laojun-avatar.svg" 
                    alt="用户头像" 
                    className="w-full h-full object-cover"
                  />
                </div>
                <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-green-500 rounded-full border-2 border-background animate-pulse"></div>
              </div>
              <span className="text-sm font-medium transition-all duration-200 group-hover:translate-x-1">
                {user?.username || '用户'}
              </span>
            </button>
            
            {showUserMenu && (
              <div className="absolute right-0 mt-3 w-56 bg-popover/95 backdrop-blur-md border border-border/50 rounded-xl shadow-xl z-50 animate-in slide-in-from-top-2 duration-200">
                <div className="py-2">
                  <div className="px-4 py-3 text-sm text-muted-foreground border-b border-border/50 bg-gradient-to-r from-primary/5 to-primary/10">
                    <div className="flex items-center space-x-2">
                      <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                      <span>{user?.email || 'user@example.com'}</span>
                    </div>
                  </div>
                  <button
                    onClick={handleUpgradeClick}
                    className={cn(
                      "flex items-center w-full px-4 py-3 text-sm transition-all duration-200 hover:scale-105 transform group",
                      isPremium 
                        ? "hover:bg-yellow-50 hover:text-yellow-600 bg-gradient-to-r from-yellow-400/10 to-orange-400/10" 
                        : "hover:bg-blue-50 hover:text-blue-600"
                    )}
                  >
                    <Crown className={cn(
                      "h-4 w-4 mr-3 transition-transform duration-200 group-hover:scale-110",
                      isPremium ? "text-yellow-500" : "text-gray-400"
                    )} />
                    <span className="transition-all duration-200 group-hover:translate-x-1">
                      {isPremium ? '切换到老君版' : '升级到太上版'}
                    </span>
                  </button>
                  <button
                    onClick={handleLogout}
                    className="flex items-center w-full px-4 py-3 text-sm hover:bg-red-50 hover:text-red-600 transition-all duration-200 hover:scale-105 transform group"
                  >
                    <LogOut className="h-4 w-4 mr-3 transition-transform duration-200 group-hover:scale-110" />
                    <span className="transition-all duration-200 group-hover:translate-x-1">退出登录</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 升级确认对话框 */}
      {showUpgradeConfirm && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-xl p-6 max-w-md w-full max-h-[90vh] overflow-y-auto shadow-2xl border border-gray-200 dark:border-gray-700 animate-in fade-in-0 zoom-in-95 duration-200">
            <div className="flex items-center space-x-3 mb-4">
              <Crown className="h-6 w-6 text-yellow-500 flex-shrink-0" />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                {isPremium ? '切换到老君版' : '升级到太上版'}
              </h3>
            </div>
            
            <p className="text-gray-600 dark:text-gray-300 mb-6 leading-relaxed">
              {isPremium 
                ? '您确定要切换到老君版吗？这将会降级您的AI功能。' 
                : '您确定要升级到太上版吗？这将为您提供更强大的AI功能。'
              }
            </p>
            
            <div className="flex space-x-3">
              <button
                onClick={handleUpgradeCancel}
                className="flex-1 px-4 py-2 text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-all duration-200 hover:scale-105 active:scale-95"
              >
                取消
              </button>
              <button
                onClick={handleUpgradeConfirm}
                className={cn(
                  "flex-1 px-4 py-2 text-white rounded-lg transition-all duration-200 hover:scale-105 active:scale-95",
                  isPremium 
                    ? "bg-gray-600 hover:bg-gray-700" 
                    : "bg-yellow-500 hover:bg-yellow-600"
                )}
              >
                确认{isPremium ? '切换' : '升级'}
              </button>
            </div>
          </div>
        </div>
      )}
    </header>
  );
}