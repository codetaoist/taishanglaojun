import { useState, useEffect } from 'react';
import { 
  Play, 
  Pause, 
  Settings, 
  Palette,
  RotateCcw,
  Heart,
  Zap,
  Coffee,
  Moon,
  MessageCircle,
  Monitor,
  Eye,
  EyeOff
} from 'lucide-react';
import { cn } from '../utils/cn';
import DesktopPet from '../components/DesktopPet';

interface PetSettings {
  enabled: boolean;
  soundEnabled: boolean;
  autoHide: boolean;
  transparency: number;
  size: 'small' | 'medium' | 'large';
  theme: 'default' | 'cute' | 'professional';
  position: 'free' | 'corner' | 'edge';
  interactions: {
    chat: boolean;
    feeding: boolean;
    playing: boolean;
    sleeping: boolean;
  };
}

interface PetStats {
  totalInteractions: number;
  chatMessages: number;
  feedingTimes: number;
  playingSessions: number;
  onlineTime: number;
  lastActive: Date;
}

export default function PetPage() {
  const [petVisible, setPetVisible] = useState(false);
  const [petMinimized, setPetMinimized] = useState(false);
  const [settings, setSettings] = useState<PetSettings>({
    enabled: false,
    soundEnabled: true,
    autoHide: false,
    transparency: 90,
    size: 'medium',
    theme: 'default',
    position: 'free',
    interactions: {
      chat: true,
      feeding: true,
      playing: true,
      sleeping: true
    }
  });

  const [stats, setStats] = useState<PetStats>({
    totalInteractions: 0,
    chatMessages: 0,
    feedingTimes: 0,
    playingSessions: 0,
    onlineTime: 0,
    lastActive: new Date()
  });

  const [activeTab, setActiveTab] = useState('control');

  // 从本地存储加载设置
  useEffect(() => {
    const savedSettings = localStorage.getItem('petSettings');
    if (savedSettings) {
      setSettings(JSON.parse(savedSettings));
    }

    const savedStats = localStorage.getItem('petStats');
    if (savedStats) {
      setStats(JSON.parse(savedStats));
    }
  }, []);

  // 保存设置到本地存储
  useEffect(() => {
    localStorage.setItem('petSettings', JSON.stringify(settings));
  }, [settings]);

  // 保存统计到本地存储
  useEffect(() => {
    localStorage.setItem('petStats', JSON.stringify(stats));
  }, [stats]);

  const handleTogglePet = () => {
    if (settings.enabled) {
      setPetVisible(!petVisible);
    }
  };

  const handleClosePet = () => {
    setPetVisible(false);
  };

  const handleMinimizePet = () => {
    setPetMinimized(true);
    setPetVisible(false);
  };

  const handleSettingChange = (key: keyof PetSettings, value: any) => {
    setSettings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleInteractionChange = (key: keyof PetSettings['interactions'], value: boolean) => {
    setSettings(prev => ({
      ...prev,
      interactions: {
        ...prev.interactions,
        [key]: value
      }
    }));
  };

  const resetStats = () => {
    setStats({
      totalInteractions: 0,
      chatMessages: 0,
      feedingTimes: 0,
      playingSessions: 0,
      onlineTime: 0,
      lastActive: new Date()
    });
  };

  const formatTime = (minutes: number): string => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${hours}小时${mins}分钟`;
  };

  const tabs = [
    { id: 'control', label: '桌宠控制', icon: Play },
    { id: 'settings', label: '桌宠设置', icon: Settings },
    { id: 'stats', label: '使用统计', icon: Monitor },
    { id: 'appearance', label: '外观设置', icon: Palette }
  ];

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div>
        <h1 className="text-xl sm:text-2xl font-bold">桌面宠物</h1>
        <p className="text-muted-foreground text-sm sm:text-base">管理你的智能桌面助手</p>
      </div>

      {/* 标签页导航 */}
      <div className="border-b border-border">
        <nav className="flex space-x-2 md:space-x-8 overflow-x-auto">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  "flex items-center space-x-2 py-2 px-2 md:px-1 border-b-2 font-medium text-sm transition-colors whitespace-nowrap",
                  activeTab === tab.id
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                )}
              >
                <Icon className="h-4 w-4" />
                <span className="hidden sm:inline">{tab.label}</span>
              </button>
            );
          })}
        </nav>
      </div>

      {/* 标签页内容 */}
      <div className="space-y-6">
        {activeTab === 'control' && (
          <div className="space-y-6">
            {/* 桌宠状态卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="p-6 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className={cn(
                    "p-3 rounded-lg",
                    settings.enabled ? "bg-green-100" : "bg-gray-100"
                  )}>
                    {settings.enabled ? (
                      <Heart className="h-6 w-6 text-green-600" />
                    ) : (
                      <Heart className="h-6 w-6 text-gray-400" />
                    )}
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">桌宠状态</div>
                    <div className={cn(
                      "text-lg font-semibold",
                      settings.enabled ? "text-green-600" : "text-gray-400"
                    )}>
                      {settings.enabled ? "已启用" : "已禁用"}
                    </div>
                  </div>
                </div>
              </div>

              <div className="p-6 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className={cn(
                    "p-3 rounded-lg",
                    petVisible ? "bg-blue-100" : "bg-gray-100"
                  )}>
                    {petVisible ? (
                      <Eye className="h-6 w-6 text-blue-600" />
                    ) : (
                      <EyeOff className="h-6 w-6 text-gray-400" />
                    )}
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">显示状态</div>
                    <div className={cn(
                      "text-lg font-semibold",
                      petVisible ? "text-blue-600" : "text-gray-400"
                    )}>
                      {petVisible ? "显示中" : "已隐藏"}
                    </div>
                  </div>
                </div>
              </div>

              <div className="p-6 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-3 bg-purple-100 rounded-lg">
                    <MessageCircle className="h-6 w-6 text-purple-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">今日互动</div>
                    <div className="text-lg font-semibold">{stats.totalInteractions}</div>
                  </div>
                </div>
              </div>
            </div>

            {/* 控制面板 */}
            <div className="p-6 bg-card rounded-lg border">
              <h3 className="text-lg font-semibold mb-4">桌宠控制</h3>
              
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium">启用桌宠</div>
                    <div className="text-sm text-muted-foreground">开启或关闭桌面宠物功能</div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.enabled}
                      onChange={(e) => handleSettingChange('enabled', e.target.checked)}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                {settings.enabled && (
                  <>
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">显示桌宠</div>
                        <div className="text-sm text-muted-foreground">在桌面上显示或隐藏桌宠</div>
                      </div>
                      <button
                        onClick={handleTogglePet}
                        className={cn(
                          "flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors",
                          petVisible
                            ? "bg-red-100 text-red-700 hover:bg-red-200"
                            : "bg-green-100 text-green-700 hover:bg-green-200"
                        )}
                      >
                        {petVisible ? (
                          <>
                            <Pause className="h-4 w-4" />
                            <span>隐藏桌宠</span>
                          </>
                        ) : (
                          <>
                            <Play className="h-4 w-4" />
                            <span>显示桌宠</span>
                          </>
                        )}
                      </button>
                    </div>

                    {petMinimized && (
                      <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-medium text-yellow-800">桌宠已最小化</div>
                            <div className="text-sm text-yellow-600">点击恢复显示桌宠</div>
                          </div>
                          <button
                            onClick={() => {
                              setPetMinimized(false);
                              setPetVisible(true);
                            }}
                            className="px-3 py-1 bg-yellow-200 text-yellow-800 rounded hover:bg-yellow-300 transition-colors"
                          >
                            恢复
                          </button>
                        </div>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="space-y-6">
            <div className="p-6 bg-card rounded-lg border">
              <h3 className="text-lg font-semibold mb-4">基础设置</h3>
              
              <div className="space-y-6">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium">声音效果</div>
                    <div className="text-sm text-muted-foreground">启用桌宠的声音反馈</div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.soundEnabled}
                      onChange={(e) => handleSettingChange('soundEnabled', e.target.checked)}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium">自动隐藏</div>
                    <div className="text-sm text-muted-foreground">长时间无互动时自动隐藏</div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.autoHide}
                      onChange={(e) => handleSettingChange('autoHide', e.target.checked)}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div>
                  <div className="flex items-center justify-between mb-2">
                    <div className="font-medium">透明度</div>
                    <span className="text-sm text-muted-foreground">{settings.transparency}%</span>
                  </div>
                  <input
                    type="range"
                    min="50"
                    max="100"
                    value={settings.transparency}
                    onChange={(e) => handleSettingChange('transparency', parseInt(e.target.value))}
                    className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
                  />
                </div>

                <div>
                  <div className="font-medium mb-2">桌宠大小</div>
                  <div className="flex space-x-2">
                    {['small', 'medium', 'large'].map((size) => (
                      <button
                        key={size}
                        onClick={() => handleSettingChange('size', size)}
                        className={cn(
                          "px-4 py-2 rounded-lg border transition-colors",
                          settings.size === size
                            ? "bg-primary text-primary-foreground border-primary"
                            : "bg-background border-border hover:bg-accent"
                        )}
                      >
                        {size === 'small' ? '小' : size === 'medium' ? '中' : '大'}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            <div className="p-6 bg-card rounded-lg border">
              <h3 className="text-lg font-semibold mb-4">交互功能</h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    id="chat"
                    checked={settings.interactions.chat}
                    onChange={(e) => handleInteractionChange('chat', e.target.checked)}
                    className="rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <label htmlFor="chat" className="flex items-center space-x-2 cursor-pointer">
                    <MessageCircle className="h-4 w-4" />
                    <span>聊天功能</span>
                  </label>
                </div>

                <div className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    id="feeding"
                    checked={settings.interactions.feeding}
                    onChange={(e) => handleInteractionChange('feeding', e.target.checked)}
                    className="rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <label htmlFor="feeding" className="flex items-center space-x-2 cursor-pointer">
                    <Coffee className="h-4 w-4" />
                    <span>喂食功能</span>
                  </label>
                </div>

                <div className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    id="playing"
                    checked={settings.interactions.playing}
                    onChange={(e) => handleInteractionChange('playing', e.target.checked)}
                    className="rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <label htmlFor="playing" className="flex items-center space-x-2 cursor-pointer">
                    <Zap className="h-4 w-4" />
                    <span>玩耍功能</span>
                  </label>
                </div>

                <div className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    id="sleeping"
                    checked={settings.interactions.sleeping}
                    onChange={(e) => handleInteractionChange('sleeping', e.target.checked)}
                    className="rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <label htmlFor="sleeping" className="flex items-center space-x-2 cursor-pointer">
                    <Moon className="h-4 w-4" />
                    <span>睡眠功能</span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'stats' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">使用统计</h3>
              <button
                onClick={resetStats}
                className="flex items-center space-x-2 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors"
              >
                <RotateCcw className="h-4 w-4" />
                <span>重置统计</span>
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <MessageCircle className="h-5 w-5 text-blue-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">聊天消息</div>
                    <div className="text-xl font-bold">{stats.chatMessages}</div>
                  </div>
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-green-100 rounded-lg">
                    <Coffee className="h-5 w-5 text-green-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">喂食次数</div>
                    <div className="text-xl font-bold">{stats.feedingTimes}</div>
                  </div>
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-yellow-100 rounded-lg">
                    <Zap className="h-5 w-5 text-yellow-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">玩耍次数</div>
                    <div className="text-xl font-bold">{stats.playingSessions}</div>
                  </div>
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-purple-100 rounded-lg">
                    <Heart className="h-5 w-5 text-purple-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">总互动</div>
                    <div className="text-xl font-bold">{stats.totalInteractions}</div>
                  </div>
                </div>
              </div>
            </div>

            <div className="p-6 bg-card rounded-lg border">
              <h4 className="font-medium mb-4">详细统计</h4>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">在线时长:</span>
                  <span className="font-medium">{formatTime(stats.onlineTime)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">最后活跃:</span>
                  <span className="font-medium">
                    {stats.lastActive.toLocaleString()}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">平均每日互动:</span>
                  <span className="font-medium">
                    {Math.round(stats.totalInteractions / Math.max(1, Math.ceil(stats.onlineTime / 1440)))}
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'appearance' && (
          <div className="space-y-6">
            <div className="p-6 bg-card rounded-lg border">
              <h3 className="text-lg font-semibold mb-4">主题设置</h3>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {['default', 'cute', 'professional'].map((theme) => (
                  <button
                    key={theme}
                    onClick={() => handleSettingChange('theme', theme)}
                    className={cn(
                      "p-4 rounded-lg border-2 transition-colors text-left",
                      settings.theme === theme
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-primary/50"
                    )}
                  >
                    <div className="font-medium mb-2">
                      {theme === 'default' ? '默认主题' : 
                       theme === 'cute' ? '可爱主题' : '专业主题'}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {theme === 'default' ? '经典的桌宠外观' : 
                       theme === 'cute' ? '更加可爱的表情和动画' : '简洁专业的商务风格'}
                    </div>
                  </button>
                ))}
              </div>
            </div>

            <div className="p-6 bg-card rounded-lg border">
              <h3 className="text-lg font-semibold mb-4">位置设置</h3>
              
              <div className="space-y-4">
                <div>
                  <div className="font-medium mb-2">桌宠位置</div>
                  <div className="flex space-x-2">
                    {[
                      { value: 'free', label: '自由拖拽' },
                      { value: 'corner', label: '固定角落' },
                      { value: 'edge', label: '贴边显示' }
                    ].map((position) => (
                      <button
                        key={position.value}
                        onClick={() => handleSettingChange('position', position.value)}
                        className={cn(
                          "px-4 py-2 rounded-lg border transition-colors",
                          settings.position === position.value
                            ? "bg-primary text-primary-foreground border-primary"
                            : "bg-background border-border hover:bg-accent"
                        )}
                      >
                        {position.label}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 桌宠组件 */}
      <DesktopPet
        visible={petVisible}
        onClose={handleClosePet}
        onMinimize={handleMinimizePet}
      />
    </div>
  );
}