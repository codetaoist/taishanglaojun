import { invoke } from '@tauri-apps/api/core';

export interface SystemMetrics {
  cpu: {
    usage: number;
    cores: number;
    frequency: number;
  };
  memory: {
    total: number;
    used: number;
    available: number;
    usage: number;
  };
  disk: {
    total: number;
    used: number;
    available: number;
    usage: number;
  };
  network: {
    status: 'connected' | 'disconnected' | 'limited';
    downloadSpeed: number;
    uploadSpeed: number;
    latency: number;
  };
  gpu?: {
    name: string;
    usage: number;
    memory: number;
    temperature: number;
  };
}

export interface ProcessInfo {
  pid: number;
  name: string;
  cpuUsage: number;
  memoryUsage: number;
  status: string;
}

export interface ServiceStatus {
  aiService: 'running' | 'stopped' | 'error';
  database: 'connected' | 'disconnected' | 'error';
  fileTransfer: 'active' | 'inactive' | 'error';
  security: 'enabled' | 'disabled' | 'warning';
}

class SystemMonitorService {
  private updateInterval: number = 5000; // 5秒更新一次
  private isMonitoring: boolean = false;
  private intervalId: NodeJS.Timeout | null = null;
  private listeners: ((metrics: SystemMetrics) => void)[] = [];

  constructor() {
    this.startMonitoring();
  }

  // 开始监控
  startMonitoring(): void {
    if (this.isMonitoring) return;
    
    this.isMonitoring = true;
    this.intervalId = setInterval(async () => {
      try {
        const metrics = await this.getSystemMetrics();
        this.notifyListeners(metrics);
      } catch (error) {
        console.error('Failed to get system metrics:', error);
      }
    }, this.updateInterval);
  }

  // 停止监控
  stopMonitoring(): void {
    if (!this.isMonitoring) return;
    
    this.isMonitoring = false;
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  // 添加监听器
  addListener(callback: (metrics: SystemMetrics) => void): void {
    this.listeners.push(callback);
  }

  // 移除监听器
  removeListener(callback: (metrics: SystemMetrics) => void): void {
    const index = this.listeners.indexOf(callback);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  // 通知所有监听器
  private notifyListeners(metrics: SystemMetrics): void {
    this.listeners.forEach(callback => callback(metrics));
  }

  // 获取系统指标
  async getSystemMetrics(): Promise<SystemMetrics> {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        // 在 Tauri 环境中调用后端
        return await invoke('get_system_metrics');
      } else {
        // 模拟数据
        return this.generateMockMetrics();
      }
    } catch (error) {
      console.error('Failed to get system metrics:', error);
      return this.generateMockMetrics();
    }
  }

  // 获取进程信息
  async getProcessList(): Promise<ProcessInfo[]> {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        return await invoke('get_process_list');
      } else {
        return this.generateMockProcesses();
      }
    } catch (error) {
      console.error('Failed to get process list:', error);
      return this.generateMockProcesses();
    }
  }

  // 获取服务状态
  async getServiceStatus(): Promise<ServiceStatus> {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        return await invoke('get_service_status');
      } else {
        return {
          aiService: 'running',
          database: 'connected',
          fileTransfer: 'active',
          security: 'enabled'
        };
      }
    } catch (error) {
      console.error('Failed to get service status:', error);
      return {
        aiService: 'error',
        database: 'error',
        fileTransfer: 'error',
        security: 'warning'
      };
    }
  }

  // 生成模拟指标
  private generateMockMetrics(): SystemMetrics {
    return {
      cpu: {
        usage: Math.random() * 80 + 10,
        cores: 8,
        frequency: 3200
      },
      memory: {
        total: 16 * 1024 * 1024 * 1024, // 16GB
        used: Math.random() * 8 * 1024 * 1024 * 1024, // 随机使用量
        available: 0,
        usage: 0
      },
      disk: {
        total: 512 * 1024 * 1024 * 1024, // 512GB
        used: Math.random() * 256 * 1024 * 1024 * 1024, // 随机使用量
        available: 0,
        usage: 0
      },
      network: {
        status: 'connected',
        downloadSpeed: Math.random() * 100,
        uploadSpeed: Math.random() * 50,
        latency: Math.random() * 50 + 10
      },
      gpu: {
        name: 'NVIDIA RTX 4080',
        usage: Math.random() * 60,
        memory: Math.random() * 12 * 1024,
        temperature: Math.random() * 30 + 50
      }
    };
  }

  // 生成模拟进程
  private generateMockProcesses(): ProcessInfo[] {
    const processes = [
      'taishang-laojun-desktop.exe',
      'chrome.exe',
      'code.exe',
      'explorer.exe',
      'winlogon.exe'
    ];

    return processes.map((name, index) => ({
      pid: 1000 + index,
      name,
      cpuUsage: Math.random() * 20,
      memoryUsage: Math.random() * 500,
      status: 'running'
    }));
  }

  // 设置更新间隔
  setUpdateInterval(interval: number): void {
    this.updateInterval = interval;
    if (this.isMonitoring) {
      this.stopMonitoring();
      this.startMonitoring();
    }
  }

  // 获取系统信息
  async getSystemInfo(): Promise<any> {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        return await invoke('get_system_info');
      } else {
        return {
          os: 'Windows 11',
          version: '22H2',
          architecture: 'x64',
          hostname: 'DESKTOP-PC',
          uptime: Math.floor(Math.random() * 86400),
          bootTime: new Date(Date.now() - Math.random() * 86400000).toISOString()
        };
      }
    } catch (error) {
      console.error('Failed to get system info:', error);
      return null;
    }
  }

  // 清理资源
  cleanup(): void {
    this.stopMonitoring();
    this.listeners = [];
  }
}

// 导出单例实例
export const systemMonitor = new SystemMonitorService();
export default systemMonitor;