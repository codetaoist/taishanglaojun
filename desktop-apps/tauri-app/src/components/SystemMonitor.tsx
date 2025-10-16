import { useState, useEffect } from 'react';
import { 
  Cpu, 
  HardDrive, 
  Wifi, 
  Activity,
  Server,
  Shield,
  Database,
  Zap
} from 'lucide-react';
import { cn } from '../utils/cn';
import systemMonitor, { SystemMetrics, ServiceStatus } from '../services/systemMonitor';

interface SystemMonitorProps {
  className?: string;
  compact?: boolean;
}

export default function SystemMonitor({ className, compact = false }: SystemMonitorProps) {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [serviceStatus, setServiceStatus] = useState<ServiceStatus | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 初始加载
    loadData();

    // 添加监听器
    const handleMetricsUpdate = (newMetrics: SystemMetrics) => {
      setMetrics(newMetrics);
      setLoading(false);
    };

    systemMonitor.addListener(handleMetricsUpdate);

    // 定期更新服务状态
    const statusInterval = setInterval(loadServiceStatus, 10000);

    return () => {
      systemMonitor.removeListener(handleMetricsUpdate);
      clearInterval(statusInterval);
    };
  }, []);

  const loadData = async () => {
    try {
      const [metricsData, statusData] = await Promise.all([
        systemMonitor.getSystemMetrics(),
        systemMonitor.getServiceStatus()
      ]);
      setMetrics(metricsData);
      setServiceStatus(statusData);
    } catch (error) {
      console.error('Failed to load system data:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadServiceStatus = async () => {
    try {
      const status = await systemMonitor.getServiceStatus();
      setServiceStatus(status);
    } catch (error) {
      console.error('Failed to load service status:', error);
    }
  };

  const formatBytes = (bytes: number): string => {
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'running':
      case 'connected':
      case 'active':
      case 'enabled':
        return 'text-green-500';
      case 'warning':
      case 'limited':
        return 'text-yellow-500';
      case 'error':
      case 'stopped':
      case 'disconnected':
      case 'inactive':
      case 'disabled':
        return 'text-red-500';
      default:
        return 'text-gray-500';
    }
  };

  const getUsageColor = (usage: number): string => {
    if (usage < 50) return 'bg-green-500';
    if (usage < 80) return 'bg-yellow-500';
    return 'bg-red-500';
  };

  if (loading) {
    return (
      <div className={cn("p-4 bg-card rounded-lg border", className)}>
        <div className="flex items-center justify-center h-32">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  if (compact) {
    return (
      <div className={cn("p-3 bg-card rounded-lg border", className)}>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <Cpu className="h-4 w-4 text-blue-500" />
              <span className="text-sm">{metrics?.cpu.usage.toFixed(1)}%</span>
            </div>
            <div className="flex items-center space-x-2">
              <Activity className="h-4 w-4 text-green-500" />
              <span className="text-sm">{metrics?.memory.usage.toFixed(1)}%</span>
            </div>
            <div className="flex items-center space-x-2">
              <HardDrive className="h-4 w-4 text-purple-500" />
              <span className="text-sm">{metrics?.disk.usage.toFixed(1)}%</span>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <div className={cn("w-2 h-2 rounded-full", 
              serviceStatus?.aiService === 'running' ? 'bg-green-500' : 'bg-red-500'
            )}></div>
            <span className="text-xs text-muted-foreground">AI服务</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={cn("p-6 bg-card rounded-lg border", className)}>
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-lg font-semibold">系统监控</h3>
        <div className="flex items-center space-x-2">
          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
          <span className="text-sm text-muted-foreground">实时监控</span>
        </div>
      </div>

      {/* 系统资源 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {/* CPU */}
        <div className="p-4 bg-background rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <Cpu className="h-5 w-5 text-blue-500" />
              <span className="font-medium">CPU</span>
            </div>
            <span className="text-sm font-mono">{metrics?.cpu.usage.toFixed(1)}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={cn("h-2 rounded-full transition-all duration-300", getUsageColor(metrics?.cpu.usage || 0))}
              style={{ width: `${metrics?.cpu.usage || 0}%` }}
            ></div>
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {metrics?.cpu.cores} 核心 @ {metrics?.cpu.frequency}MHz
          </div>
        </div>

        {/* 内存 */}
        <div className="p-4 bg-background rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <Activity className="h-5 w-5 text-green-500" />
              <span className="font-medium">内存</span>
            </div>
            <span className="text-sm font-mono">{metrics?.memory.usage.toFixed(1)}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={cn("h-2 rounded-full transition-all duration-300", getUsageColor(metrics?.memory.usage || 0))}
              style={{ width: `${metrics?.memory.usage || 0}%` }}
            ></div>
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {formatBytes(metrics?.memory.used || 0)} / {formatBytes(metrics?.memory.total || 0)}
          </div>
        </div>

        {/* 磁盘 */}
        <div className="p-4 bg-background rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <HardDrive className="h-5 w-5 text-purple-500" />
              <span className="font-medium">磁盘</span>
            </div>
            <span className="text-sm font-mono">{metrics?.disk.usage.toFixed(1)}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={cn("h-2 rounded-full transition-all duration-300", getUsageColor(metrics?.disk.usage || 0))}
              style={{ width: `${metrics?.disk.usage || 0}%` }}
            ></div>
          </div>
          <div className="text-xs text-muted-foreground mt-1">
            {formatBytes(metrics?.disk.used || 0)} / {formatBytes(metrics?.disk.total || 0)}
          </div>
        </div>

        {/* 网络 */}
        <div className="p-4 bg-background rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center space-x-2">
              <Wifi className={cn("h-5 w-5", getStatusColor(metrics?.network.status || 'disconnected'))} />
              <span className="font-medium">网络</span>
            </div>
            <span className={cn("text-sm", getStatusColor(metrics?.network.status || 'disconnected'))}>
              {metrics?.network.status}
            </span>
          </div>
          <div className="text-xs text-muted-foreground">
            <div>下载: {metrics?.network.downloadSpeed.toFixed(1)} MB/s</div>
            <div>上传: {metrics?.network.uploadSpeed.toFixed(1)} MB/s</div>
            <div>延迟: {metrics?.network.latency.toFixed(0)} ms</div>
          </div>
        </div>
      </div>

      {/* GPU 信息 */}
      {metrics?.gpu && (
        <div className="mb-6">
          <h4 className="text-md font-medium mb-3">GPU 状态</h4>
          <div className="p-4 bg-background rounded-lg">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center space-x-2">
                <Zap className="h-5 w-5 text-orange-500" />
                <span className="font-medium">{metrics.gpu.name}</span>
              </div>
              <span className="text-sm font-mono">{metrics.gpu.usage.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
              <div 
                className={cn("h-2 rounded-full transition-all duration-300", getUsageColor(metrics.gpu.usage))}
                style={{ width: `${metrics.gpu.usage}%` }}
              ></div>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>显存: {formatBytes(metrics.gpu.memory)}</span>
              <span>温度: {metrics.gpu.temperature.toFixed(0)}°C</span>
            </div>
          </div>
        </div>
      )}

      {/* 服务状态 */}
      <div>
        <h4 className="text-md font-medium mb-3">服务状态</h4>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div className="flex items-center space-x-2 p-3 bg-background rounded-lg">
            <Server className={cn("h-4 w-4", getStatusColor(serviceStatus?.aiService || 'error'))} />
            <div>
              <div className="text-sm font-medium">AI服务</div>
              <div className={cn("text-xs", getStatusColor(serviceStatus?.aiService || 'error'))}>
                {serviceStatus?.aiService || 'unknown'}
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-2 p-3 bg-background rounded-lg">
            <Database className={cn("h-4 w-4", getStatusColor(serviceStatus?.database || 'error'))} />
            <div>
              <div className="text-sm font-medium">数据库</div>
              <div className={cn("text-xs", getStatusColor(serviceStatus?.database || 'error'))}>
                {serviceStatus?.database || 'unknown'}
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-2 p-3 bg-background rounded-lg">
            <Activity className={cn("h-4 w-4", getStatusColor(serviceStatus?.fileTransfer || 'error'))} />
            <div>
              <div className="text-sm font-medium">文件传输</div>
              <div className={cn("text-xs", getStatusColor(serviceStatus?.fileTransfer || 'error'))}>
                {serviceStatus?.fileTransfer || 'unknown'}
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-2 p-3 bg-background rounded-lg">
            <Shield className={cn("h-4 w-4", getStatusColor(serviceStatus?.security || 'error'))} />
            <div>
              <div className="text-sm font-medium">安全服务</div>
              <div className={cn("text-xs", getStatusColor(serviceStatus?.security || 'error'))}>
                {serviceStatus?.security || 'unknown'}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}