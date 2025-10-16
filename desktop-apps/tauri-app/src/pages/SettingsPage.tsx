import { useState, useEffect } from 'react';
import { 
  Settings, 
  Shield, 
  Palette, 
  Database,
  Save,
  RefreshCw
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { cn } from '../utils/cn';
import aiService from '../services/aiService';

interface SystemStatus {
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  network_status: string;
  ai_service_status: string;
  database_status: string;
}

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('general');
  const [systemStatus, setSystemStatus] = useState<SystemStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const { theme, setTheme } = useTheme();


  // Settings state
  const [settings, setSettings] = useState({
    language: 'zh-CN',
    autoSave: true,
    notifications: true,
    aiModel: 'gpt-4',
    maxTokens: 2048,
    temperature: 0.7,
    apiEndpoint: 'https://api.openai.com/v1',
    encryptionEnabled: true,
    backupEnabled: true,
    backupInterval: 24,
  });

  useEffect(() => {
    loadSystemStatus();
    loadSettings();
  }, []);

  const loadSystemStatus = async () => {
    try {
      // 获取AI服务状态和模拟系统状态
      const aiStatus = await aiService.getSystemStatus();
      const mockStatus: SystemStatus = {
        cpu_usage: Math.random() * 100,
        memory_usage: Math.random() * 100,
        disk_usage: Math.random() * 100,
        network_status: 'connected',
        ai_service_status: aiStatus.success ? 'running' : 'error',
        database_status: 'connected'
      };
      setSystemStatus(mockStatus);
    } catch (error) {
      console.error('Failed to load system status:', error);
    }
  };

  const loadSettings = async () => {
    try {
      // Load settings from backend
      // const savedSettings = await invoke('get_settings');
      // setSettings(savedSettings);
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  };

  const saveSettings = async () => {
    setLoading(true);
    try {
      // 保存设置到本地存储
      localStorage.setItem('app_settings', JSON.stringify(settings));
      console.log('Settings saved:', settings);
      // Show success message
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSettingChange = (key: string, value: any) => {
    setSettings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const tabs = [
    { id: 'general', label: '常规设置', icon: Settings },
    { id: 'ai', label: 'AI设置', icon: RefreshCw },
    { id: 'security', label: '安全设置', icon: Shield },
    { id: 'appearance', label: '外观设置', icon: Palette },
    { id: 'system', label: '系统状态', icon: Database },
  ];

  const formatPercentage = (value: number) => `${(value * 100).toFixed(1)}%`;

  return (
    <div className="h-full flex flex-col space-y-6 bg-gradient-to-br from-gray-50/50 to-blue-50/30 dark:from-gray-900/50 dark:to-blue-900/20 p-6">
      {/* Header */}
      <div className="fade-in">
        <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-800 to-blue-600 dark:from-white dark:to-blue-400 bg-clip-text text-transparent">设置</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2 slide-in-up" style={{ animationDelay: '0.1s' }}>
          配置应用程序设置和查看系统状态
        </p>
      </div>

      <div className="flex-1 grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar */}
        <div className="lg:col-span-1">
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-xl border border-gray-200/50 dark:border-gray-700/50 rounded-2xl p-6 shadow-lg slide-in-left" style={{ animationDelay: '0.2s' }}>
            <nav className="space-y-3">
              {tabs.map((tab, index) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={cn(
                    'w-full flex items-center space-x-3 px-4 py-3 text-sm font-medium rounded-xl transition-all duration-200 hover:scale-105 stagger-animation',
                    activeTab === tab.id
                      ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white shadow-lg'
                      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100/80 dark:hover:bg-gray-700/80 hover:text-gray-800 dark:hover:text-gray-200'
                  )}
                  style={{ animationDelay: `${0.3 + index * 0.1}s` }}
                >
                  <tab.icon className="h-5 w-5" />
                  <span>{tab.label}</span>
                </button>
              ))}
            </nav>
          </div>
        </div>

        {/* Content */}
        <div className="lg:col-span-3">
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-xl border border-gray-200/50 dark:border-gray-700/50 rounded-2xl p-8 shadow-lg slide-in-right" style={{ animationDelay: '0.3s' }}>
            {/* General Settings */}
            {activeTab === 'general' && (
              <div className="space-y-8 fade-in" style={{ animationDelay: '0.4s' }}>
                <h2 className="text-2xl font-semibold bg-gradient-to-r from-gray-800 to-blue-600 dark:from-white dark:to-blue-400 bg-clip-text text-transparent">常规设置</h2>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                  <div className="scale-in" style={{ animationDelay: '0.5s' }}>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">界面语言</label>
                    <select
                      value={settings.language}
                      onChange={(e) => handleSettingChange('language', e.target.value)}
                      className="w-full px-4 py-3 bg-white/70 dark:bg-gray-700/70 border border-gray-200/50 dark:border-gray-600/50 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-transparent backdrop-blur-sm transition-all duration-200 hover:bg-white/90 dark:hover:bg-gray-700/90"
                    >
                      <option value="zh-CN">简体中文</option>
                      <option value="zh-TW">繁体中文</option>
                      <option value="en-US">English</option>
                      <option value="ja-JP">日本語</option>
                    </select>
                  </div>

                  <div className="flex items-center justify-between p-4 bg-gray-50/50 dark:bg-gray-700/30 rounded-xl border border-gray-200/30 dark:border-gray-600/30 scale-in" style={{ animationDelay: '0.6s' }}>
                    <div>
                      <label className="text-sm font-medium text-gray-700 dark:text-gray-300">自动保存</label>
                      <p className="text-xs text-gray-500 dark:text-gray-400">自动保存对话和文档</p>
                    </div>
                    <button
                      onClick={() => handleSettingChange('autoSave', !settings.autoSave)}
                      className={cn(
                        'relative inline-flex h-7 w-12 items-center rounded-full transition-all duration-200 hover:scale-105',
                        settings.autoSave ? 'bg-gradient-to-r from-blue-500 to-purple-600' : 'bg-gray-300 dark:bg-gray-600'
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-5 w-5 transform rounded-full bg-white shadow-lg transition-transform duration-200',
                          settings.autoSave ? 'translate-x-6' : 'translate-x-1'
                        )}
                      />
                    </button>
                  </div>

                  <div className="flex items-center justify-between p-4 bg-gray-50/50 dark:bg-gray-700/30 rounded-xl border border-gray-200/30 dark:border-gray-600/30 scale-in" style={{ animationDelay: '0.7s' }}>
                    <div>
                      <label className="text-sm font-medium text-gray-700 dark:text-gray-300">桌面通知</label>
                      <p className="text-xs text-gray-500 dark:text-gray-400">接收系统通知</p>
                    </div>
                    <button
                      onClick={() => handleSettingChange('notifications', !settings.notifications)}
                      className={cn(
                        'relative inline-flex h-7 w-12 items-center rounded-full transition-all duration-200 hover:scale-105',
                        settings.notifications ? 'bg-gradient-to-r from-blue-500 to-purple-600' : 'bg-gray-300 dark:bg-gray-600'
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-5 w-5 transform rounded-full bg-white shadow-lg transition-transform duration-200',
                          settings.notifications ? 'translate-x-6' : 'translate-x-1'
                        )}
                      />
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* AI Settings */}
            {activeTab === 'ai' && (
              <div className="space-y-6">
                <h2 className="text-lg font-semibold">AI设置</h2>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium mb-2">AI模型</label>
                    <select
                      value={settings.aiModel}
                      onChange={(e) => handleSettingChange('aiModel', e.target.value)}
                      className="input w-full"
                    >
                      <option value="gpt-4">GPT-4</option>
                      <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                      <option value="claude-3">Claude 3</option>
                      <option value="gemini-pro">Gemini Pro</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">最大令牌数</label>
                    <input
                      type="number"
                      value={settings.maxTokens}
                      onChange={(e) => handleSettingChange('maxTokens', Number(e.target.value))}
                      className="input w-full"
                      min="512"
                      max="8192"
                      step="256"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      创造性 ({settings.temperature})
                    </label>
                    <input
                      type="range"
                      value={settings.temperature}
                      onChange={(e) => handleSettingChange('temperature', Number(e.target.value))}
                      className="w-full"
                      min="0"
                      max="1"
                      step="0.1"
                    />
                    <div className="flex justify-between text-xs text-muted-foreground mt-1">
                      <span>保守</span>
                      <span>创新</span>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">API端点</label>
                    <input
                      type="url"
                      value={settings.apiEndpoint}
                      onChange={(e) => handleSettingChange('apiEndpoint', e.target.value)}
                      className="input w-full"
                      placeholder="https://api.openai.com/v1"
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Security Settings */}
            {activeTab === 'security' && (
              <div className="space-y-6">
                <h2 className="text-lg font-semibold">安全设置</h2>
                
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 border border-border rounded-lg">
                    <div>
                      <label className="text-sm font-medium">数据加密</label>
                      <p className="text-xs text-muted-foreground">加密存储的对话和文档</p>
                    </div>
                    <button
                      onClick={() => handleSettingChange('encryptionEnabled', !settings.encryptionEnabled)}
                      className={cn(
                        'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                        settings.encryptionEnabled ? 'bg-primary' : 'bg-secondary'
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                          settings.encryptionEnabled ? 'translate-x-6' : 'translate-x-1'
                        )}
                      />
                    </button>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-border rounded-lg">
                    <div>
                      <label className="text-sm font-medium">自动备份</label>
                      <p className="text-xs text-muted-foreground">定期备份应用数据</p>
                    </div>
                    <button
                      onClick={() => handleSettingChange('backupEnabled', !settings.backupEnabled)}
                      className={cn(
                        'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                        settings.backupEnabled ? 'bg-primary' : 'bg-secondary'
                      )}
                    >
                      <span
                        className={cn(
                          'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                          settings.backupEnabled ? 'translate-x-6' : 'translate-x-1'
                        )}
                      />
                    </button>
                  </div>

                  {settings.backupEnabled && (
                    <div className="ml-4">
                      <label className="block text-sm font-medium mb-2">备份间隔（小时）</label>
                      <input
                        type="number"
                        value={settings.backupInterval}
                        onChange={(e) => handleSettingChange('backupInterval', Number(e.target.value))}
                        className="input w-32"
                        min="1"
                        max="168"
                      />
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Appearance Settings */}
            {activeTab === 'appearance' && (
              <div className="space-y-6">
                <h2 className="text-lg font-semibold">外观设置</h2>
                
                <div>
                  <label className="block text-sm font-medium mb-4">主题模式</label>
                  <div className="grid grid-cols-3 gap-4">
                    {[
                      { value: 'light', label: '浅色模式', icon: '☀️' },
                      { value: 'dark', label: '深色模式', icon: '🌙' },
                      { value: 'system', label: '跟随系统', icon: '💻' },
                    ].map((option) => (
                      <button
                        key={option.value}
                        onClick={() => setTheme(option.value as any)}
                        className={cn(
                          'p-4 border rounded-lg text-center transition-colors',
                          theme === option.value
                            ? 'border-primary bg-primary/10'
                            : 'border-border hover:border-primary/50'
                        )}
                      >
                        <div className="text-2xl mb-2">{option.icon}</div>
                        <p className="text-sm font-medium">{option.label}</p>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* System Status */}
            {activeTab === 'system' && (
              <div className="space-y-6">
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-semibold">系统状态</h2>
                  <button
                    onClick={loadSystemStatus}
                    className="btn-secondary"
                  >
                    <RefreshCw className="h-4 w-4 mr-2" />
                    刷新
                  </button>
                </div>
                
                {systemStatus ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="p-4 border border-border rounded-lg">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">CPU使用率</span>
                        <span className="text-sm text-muted-foreground">
                          {formatPercentage(systemStatus.cpu_usage)}
                        </span>
                      </div>
                      <div className="w-full bg-secondary rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full transition-all"
                          style={{ width: `${systemStatus.cpu_usage * 100}%` }}
                        />
                      </div>
                    </div>

                    <div className="p-4 border border-border rounded-lg">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">内存使用率</span>
                        <span className="text-sm text-muted-foreground">
                          {formatPercentage(systemStatus.memory_usage)}
                        </span>
                      </div>
                      <div className="w-full bg-secondary rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full transition-all"
                          style={{ width: `${systemStatus.memory_usage * 100}%` }}
                        />
                      </div>
                    </div>

                    <div className="p-4 border border-border rounded-lg">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">磁盘使用率</span>
                        <span className="text-sm text-muted-foreground">
                          {formatPercentage(systemStatus.disk_usage)}
                        </span>
                      </div>
                      <div className="w-full bg-secondary rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full transition-all"
                          style={{ width: `${systemStatus.disk_usage * 100}%` }}
                        />
                      </div>
                    </div>

                    <div className="p-4 border border-border rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">网络状态</span>
                        <span className={cn(
                          'text-sm px-2 py-1 rounded',
                          systemStatus.network_status === 'connected'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        )}>
                          {systemStatus.network_status === 'connected' ? '已连接' : '未连接'}
                        </span>
                      </div>
                    </div>

                    <div className="p-4 border border-border rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">AI服务状态</span>
                        <span className={cn(
                          'text-sm px-2 py-1 rounded',
                          systemStatus.ai_service_status === 'running'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        )}>
                          {systemStatus.ai_service_status === 'running' ? '运行中' : '已停止'}
                        </span>
                      </div>
                    </div>

                    <div className="p-4 border border-border rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">数据库状态</span>
                        <span className={cn(
                          'text-sm px-2 py-1 rounded',
                          systemStatus.database_status === 'connected'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        )}>
                          {systemStatus.database_status === 'connected' ? '已连接' : '未连接'}
                        </span>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-center h-32 text-muted-foreground">
                    <p>加载系统状态中...</p>
                  </div>
                )}
              </div>
            )}

            {/* Save button */}
            {activeTab !== 'system' && (
              <div className="pt-6 border-t border-border">
                <button
                  onClick={saveSettings}
                  disabled={loading}
                  className={cn(
                    'btn-primary',
                    loading && 'opacity-50 cursor-not-allowed'
                  )}
                >
                  {loading ? (
                    <>
                      <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                      保存中...
                    </>
                  ) : (
                    <>
                      <Save className="h-4 w-4 mr-2" />
                      保存设置
                    </>
                  )}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}