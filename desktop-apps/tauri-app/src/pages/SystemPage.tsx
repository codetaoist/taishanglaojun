import { useState, useEffect } from 'react';
import { 
  Monitor, 
  Activity, 
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  Info,
  Database
} from 'lucide-react';
import { cn } from '../utils/cn';
import SystemMonitor from '../components/SystemMonitor';
import DataDashboard from '../components/DataDashboard';
import systemMonitor, { ProcessInfo } from '../services/systemMonitor';
import aiService from '../services/aiService';

export default function SystemPage() {
  const [activeTab, setActiveTab] = useState('overview');
  const [processes, setProcesses] = useState<ProcessInfo[]>([]);
  const [systemInfo, setSystemInfo] = useState<any>(null);
  const [aiMetrics, setAiMetrics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [processData, infoData, metricsData] = await Promise.all([
        systemMonitor.getProcessList(),
        systemMonitor.getSystemInfo(),
        aiService.getMetrics()
      ]);
      
      setProcesses(processData);
      setSystemInfo(infoData);
      setAiMetrics(metricsData);
    } catch (error) {
      console.error('Failed to load system data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await loadData();
    setRefreshing(false);
  };

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
      return `${days}天 ${hours}小时 ${minutes}分钟`;
    } else if (hours > 0) {
      return `${hours}小时 ${minutes}分钟`;
    } else {
      return `${minutes}分钟`;
    }
  };

  const getProcessStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'stopped':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
    }
  };

  const tabs = [
    { id: 'overview', label: '系统概览', icon: Monitor },
    { id: 'processes', label: '进程管理', icon: Activity },
    { id: 'ai-metrics', label: 'AI指标', icon: RefreshCw },
    { id: 'data-management', label: '数据管理', icon: Database },
    { id: 'system-info', label: '系统信息', icon: Info }
  ];

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题和刷新按钮 */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold">系统监控</h1>
          <p className="text-muted-foreground text-sm sm:text-base">实时监控系统状态和性能指标</p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className={cn(
            "flex items-center justify-center space-x-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors w-full sm:w-auto",
            refreshing && "opacity-50 cursor-not-allowed"
          )}
        >
          <RefreshCw className={cn("h-4 w-4", refreshing && "animate-spin")} />
          <span>刷新</span>
        </button>
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
        {activeTab === 'overview' && (
          <div className="space-y-6">
            <SystemMonitor />
            
            {/* 快速状态卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-green-100 rounded-lg">
                    <CheckCircle className="h-6 w-6 text-green-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">系统状态</div>
                    <div className="text-lg font-semibold text-green-600">正常运行</div>
                  </div>
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <Activity className="h-6 w-6 text-blue-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">活跃进程</div>
                    <div className="text-lg font-semibold">{processes.length}</div>
                  </div>
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-purple-100 rounded-lg">
                    <Clock className="h-6 w-6 text-purple-600" />
                  </div>
                  <div>
                    <div className="text-sm text-muted-foreground">运行时间</div>
                    <div className="text-lg font-semibold">
                      {systemInfo?.uptime ? formatUptime(systemInfo.uptime) : '未知'}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'processes' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">进程列表</h3>
              <span className="text-sm text-muted-foreground">
                共 {processes.length} 个进程
              </span>
            </div>

            <div className="bg-card rounded-lg border overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-muted/50">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium">进程名</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">PID</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">CPU使用率</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">内存使用</th>
                      <th className="px-4 py-3 text-left text-sm font-medium">状态</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {processes.map((process) => (
                      <tr key={process.pid} className="hover:bg-muted/30">
                        <td className="px-4 py-3">
                          <div className="flex items-center space-x-2">
                            {getProcessStatusIcon(process.status)}
                            <span className="font-medium">{process.name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 font-mono text-sm">{process.pid}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center space-x-2">
                            <div className="w-16 bg-gray-200 rounded-full h-2">
                              <div 
                                className="bg-blue-500 h-2 rounded-full"
                                style={{ width: `${Math.min(process.cpuUsage, 100)}%` }}
                              ></div>
                            </div>
                            <span className="text-sm font-mono">{process.cpuUsage.toFixed(1)}%</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 font-mono text-sm">
                          {process.memoryUsage.toFixed(1)} MB
                        </td>
                        <td className="px-4 py-3">
                          <span className={cn(
                            "px-2 py-1 rounded-full text-xs font-medium",
                            process.status === 'running' 
                              ? "bg-green-100 text-green-800" 
                              : "bg-red-100 text-red-800"
                          )}>
                            {process.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'ai-metrics' && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">AI服务指标</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="p-4 bg-card rounded-lg border">
                <div className="text-sm text-muted-foreground">总请求数</div>
                <div className="text-2xl font-bold">{aiMetrics?.totalRequests || 0}</div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="text-sm text-muted-foreground">成功请求</div>
                <div className="text-2xl font-bold text-green-600">
                  {aiMetrics?.successfulRequests || 0}
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="text-sm text-muted-foreground">失败请求</div>
                <div className="text-2xl font-bold text-red-600">
                  {aiMetrics?.failedRequests || 0}
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <div className="text-sm text-muted-foreground">成功率</div>
                <div className="text-2xl font-bold">
                  {((aiMetrics?.successRate || 0) * 100).toFixed(1)}%
                </div>
              </div>
            </div>

            <div className="p-4 bg-card rounded-lg border">
              <div className="text-sm text-muted-foreground mb-2">平均响应时间</div>
              <div className="text-3xl font-bold">
                {(aiMetrics?.averageResponseTime || 0).toFixed(0)} ms
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                <div 
                  className="bg-blue-500 h-2 rounded-full"
                  style={{ width: `${Math.min((aiMetrics?.averageResponseTime || 0) / 3000 * 100, 100)}%` }}
                ></div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'data-management' && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">数据管理</h3>
            <DataDashboard />
          </div>
        )}

        {activeTab === 'system-info' && (
          <div className="space-y-6">
            <h3 className="text-lg font-semibold">系统信息</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="p-4 bg-card rounded-lg border">
                <h4 className="font-medium mb-3">操作系统</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">系统:</span>
                    <span>{systemInfo?.os || '未知'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">版本:</span>
                    <span>{systemInfo?.version || '未知'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">架构:</span>
                    <span>{systemInfo?.architecture || '未知'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">主机名:</span>
                    <span>{systemInfo?.hostname || '未知'}</span>
                  </div>
                </div>
              </div>

              <div className="p-4 bg-card rounded-lg border">
                <h4 className="font-medium mb-3">运行时间</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">启动时间:</span>
                    <span>
                      {systemInfo?.bootTime 
                        ? new Date(systemInfo.bootTime).toLocaleString() 
                        : '未知'
                      }
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">运行时长:</span>
                    <span>
                      {systemInfo?.uptime ? formatUptime(systemInfo.uptime) : '未知'}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}