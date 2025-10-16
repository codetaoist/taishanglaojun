import { invoke } from '@tauri-apps/api/core';

export type PlatformType = 'windows' | 'macos' | 'linux';

// 检查是否在Tauri环境中
export const isTauriEnvironment = (): boolean => {
  return typeof window !== 'undefined' && '__TAURI__' in window;
};

export interface PlatformFeatures {
  supportsGlobalHotkeys: boolean;
  supportsSystemTray: boolean;
  supportsNotifications: boolean;
  supportsClipboardAccess: boolean;
  supportsFileSystemWatch: boolean;
  supportsAutoStart: boolean;
}

export class PlatformManager {
  private static instance: PlatformManager;
  private currentPlatform: PlatformType | null = null;
  private features: PlatformFeatures | null = null;

  private constructor() {}

  static getInstance(): PlatformManager {
    if (!PlatformManager.instance) {
      PlatformManager.instance = new PlatformManager();
    }
    return PlatformManager.instance;
  }

  async initialize(): Promise<void> {
    try {
      if (isTauriEnvironment()) {
        // 在Tauri环境中，动态导入platform插件
        const { platform } = await import('@tauri-apps/plugin-os');
        const platformName = await platform();
        if (platformName) {
          this.currentPlatform = this.mapPlatformName(platformName);
          this.features = this.getPlatformFeatures(this.currentPlatform);
        } else {
          throw new Error('Platform name is undefined');
        }
      } else {
        // 在浏览器环境中，使用navigator.platform检测
        this.currentPlatform = this.detectBrowserPlatform();
        this.features = this.getPlatformFeatures(this.currentPlatform);
      }
    } catch (error) {
      console.error('Failed to initialize platform manager:', error);
      // 默认为 Windows
      this.currentPlatform = 'windows';
      this.features = this.getPlatformFeatures('windows');
    }
  }

  private detectBrowserPlatform(): PlatformType {
    if (typeof navigator === 'undefined') {
      return 'windows'; // 默认
    }
    
    const platform = navigator.platform.toLowerCase();
    const userAgent = navigator.userAgent.toLowerCase();
    
    if (platform.includes('mac') || userAgent.includes('mac')) {
      return 'macos';
    } else if (platform.includes('linux') || userAgent.includes('linux')) {
      return 'linux';
    } else {
      return 'windows';
    }
  }

  private mapPlatformName(platformName: string): PlatformType {
    switch (platformName.toLowerCase()) {
      case 'win32':
      case 'windows':
        return 'windows';
      case 'darwin':
      case 'macos':
        return 'macos';
      case 'linux':
        return 'linux';
      default:
        return 'windows'; // 默认
    }
  }

  private getPlatformFeatures(platform: PlatformType): PlatformFeatures {
    const baseFeatures: PlatformFeatures = {
      supportsGlobalHotkeys: true,
      supportsSystemTray: true,
      supportsNotifications: true,
      supportsClipboardAccess: true,
      supportsFileSystemWatch: true,
      supportsAutoStart: true,
    };

    switch (platform) {
      case 'windows':
        return {
          ...baseFeatures,
          // Windows 特定功能
        };
      case 'macos':
        return {
          ...baseFeatures,
          // macOS 特定功能
        };
      case 'linux':
        return {
          ...baseFeatures,
          // Linux 特定功能可能有限制
          supportsAutoStart: false, // 某些 Linux 发行版可能不支持
        };
      default:
        return baseFeatures;
    }
  }

  getCurrentPlatform(): PlatformType {
    return this.currentPlatform || 'windows';
  }

  getFeatures(): PlatformFeatures {
    return this.features || this.getPlatformFeatures('windows');
  }

  isWindows(): boolean {
    return this.getCurrentPlatform() === 'windows';
  }

  isMacOS(): boolean {
    return this.getCurrentPlatform() === 'macos';
  }

  isLinux(): boolean {
    return this.getCurrentPlatform() === 'linux';
  }

  // 平台特定的输入优化功能
  async optimizeForPlatform(text: string): Promise<string> {
    try {
      const platform = this.getCurrentPlatform();
      
      switch (platform) {
        case 'windows':
          return await this.optimizeForWindows(text);
        case 'macos':
          return await this.optimizeForMacOS(text);
        case 'linux':
          return await this.optimizeForLinux(text);
        default:
          return text;
      }
    } catch (error) {
      console.error('Platform optimization failed:', error);
      return text;
    }
  }

  private async optimizeForWindows(text: string): Promise<string> {
    // Windows 特定的优化逻辑
    try {
      // 调用 Tauri 后端的 Windows 特定功能
      const result = await invoke('optimize_input_windows', { text });
      return result as string;
    } catch (error) {
      console.error('Windows optimization failed:', error);
      return text;
    }
  }

  private async optimizeForMacOS(text: string): Promise<string> {
    // macOS 特定的优化逻辑
    try {
      // 调用 Tauri 后端的 macOS 特定功能
      const result = await invoke('optimize_input_macos', { text });
      return result as string;
    } catch (error) {
      console.error('macOS optimization failed:', error);
      return text;
    }
  }

  private async optimizeForLinux(text: string): Promise<string> {
    // Linux 特定的优化逻辑
    try {
      // 调用 Tauri 后端的 Linux 特定功能
      const result = await invoke('optimize_input_linux', { text });
      return result as string;
    } catch (error) {
      console.error('Linux optimization failed:', error);
      return text;
    }
  }

  // 获取平台特定的快捷键
  getPlatformShortcuts(): Record<string, string> {
    const platform = this.getCurrentPlatform();
    
    switch (platform) {
      case 'windows':
        return {
          optimize: 'Ctrl+Shift+O',
          send: 'Ctrl+Enter',
          newChat: 'Ctrl+N',
          clearInput: 'Ctrl+L',
        };
      case 'macos':
        return {
          optimize: 'Cmd+Shift+O',
          send: 'Cmd+Enter',
          newChat: 'Cmd+N',
          clearInput: 'Cmd+L',
        };
      case 'linux':
        return {
          optimize: 'Ctrl+Shift+O',
          send: 'Ctrl+Enter',
          newChat: 'Ctrl+N',
          clearInput: 'Ctrl+L',
        };
      default:
        return {
          optimize: 'Ctrl+Shift+O',
          send: 'Ctrl+Enter',
          newChat: 'Ctrl+N',
          clearInput: 'Ctrl+L',
        };
    }
  }

  // 获取平台特定的样式类
  getPlatformStyles(): Record<string, string> {
    const platform = this.getCurrentPlatform();
    
    return {
      titlebar: platform === 'macos' ? 'titlebar-macos' : 'titlebar-default',
      window: platform === 'macos' ? 'window-macos' : 'window-default',
      button: platform === 'macos' ? 'button-macos' : 'button-default',
      input: platform === 'macos' ? 'input-macos' : 'input-default',
    };
  }

  // 注册全局快捷键
  async registerGlobalShortcuts(): Promise<void> {
    if (!isTauriEnvironment()) {
      console.warn('Global hotkeys only available in Tauri environment');
      return;
    }

    if (!this.getFeatures().supportsGlobalHotkeys) {
      console.warn('Global hotkeys not supported on this platform');
      return;
    }

    try {
      const shortcuts = this.getPlatformShortcuts();
      await invoke('register_global_shortcuts', { shortcuts });
    } catch (error) {
      console.error('Failed to register global shortcuts:', error);
    }
  }

  // 设置系统托盘
  async setupSystemTray(): Promise<void> {
    if (!isTauriEnvironment()) {
      console.warn('System tray only available in Tauri environment');
      return;
    }

    if (!this.getFeatures().supportsSystemTray) {
      console.warn('System tray not supported on this platform');
      return;
    }

    try {
      await invoke('setup_system_tray', { platform: this.getCurrentPlatform() });
    } catch (error) {
      console.error('Failed to setup system tray:', error);
    }
  }

  // 发送系统通知
  async sendNotification(title: string, body: string): Promise<void> {
    if (!isTauriEnvironment()) {
      // 在浏览器环境中使用Web Notifications API
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(title, { body });
      } else if ('Notification' in window && Notification.permission !== 'denied') {
        const permission = await Notification.requestPermission();
        if (permission === 'granted') {
          new Notification(title, { body });
        }
      }
      return;
    }

    if (!this.getFeatures().supportsNotifications) {
      console.warn('Notifications not supported on this platform');
      return;
    }

    try {
      await invoke('send_notification', { title, body, platform: this.getCurrentPlatform() });
    } catch (error) {
      console.error('Failed to send notification:', error);
    }
  }

  // 访问剪贴板
  async readClipboard(): Promise<string> {
    if (!isTauriEnvironment()) {
      // 在浏览器环境中使用Clipboard API
      if (navigator.clipboard && navigator.clipboard.readText) {
        try {
          return await navigator.clipboard.readText();
        } catch (error) {
          console.warn('Clipboard read access denied or not available');
          return '';
        }
      }
      return '';
    }

    if (!this.getFeatures().supportsClipboardAccess) {
      console.warn('Clipboard access not supported on this platform');
      return '';
    }

    try {
      const result = await invoke('read_clipboard');
      return result as string;
    } catch (error) {
      console.error('Failed to read clipboard:', error);
      return '';
    }
  }

  async writeClipboard(text: string): Promise<void> {
    if (!isTauriEnvironment()) {
      // 在浏览器环境中使用Clipboard API
      if (navigator.clipboard && navigator.clipboard.writeText) {
        try {
          await navigator.clipboard.writeText(text);
        } catch (error) {
          console.warn('Clipboard write access denied or not available');
        }
      }
      return;
    }

    if (!this.getFeatures().supportsClipboardAccess) {
      console.warn('Clipboard access not supported on this platform');
      return;
    }

    try {
      await invoke('write_clipboard', { text });
    } catch (error) {
      console.error('Failed to write clipboard:', error);
    }
  }
}

// 导出单例实例
export const platformManager = PlatformManager.getInstance();

// 工具函数
export const isPlatform = (targetPlatform: PlatformType): boolean => {
  return platformManager.getCurrentPlatform() === targetPlatform;
};

export const getPlatformModifierKey = (): string => {
  return platformManager.isMacOS() ? 'Cmd' : 'Ctrl';
};

export const formatShortcut = (shortcut: string): string => {
  const modifierKey = getPlatformModifierKey();
  return shortcut.replace(/Ctrl|Cmd/g, modifierKey);
};