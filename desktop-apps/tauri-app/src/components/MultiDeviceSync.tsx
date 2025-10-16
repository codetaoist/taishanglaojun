import React, { useState, useEffect } from 'react';
import { 
  multiDeviceSync, 
  DeviceInfo, 
  SyncStatus, 
  DeviceType 
} from '../services/multiDeviceSync';

interface MultiDeviceSyncProps {
  userId: string;
}

const MultiDeviceSync: React.FC<MultiDeviceSyncProps> = ({ userId }) => {
  const [syncStatus, setSyncStatus] = useState<SyncStatus>(SyncStatus.Disconnected);
  const [onlineDevices, setOnlineDevices] = useState<DeviceInfo[]>([]);
  const [lastSyncTime, setLastSyncTime] = useState<Date | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [syncProgress, setSyncProgress] = useState(0);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    initializeSync();
    
    // 添加同步状态监听器
    multiDeviceSync.addSyncListener('statusChange', handleStatusChange);
    multiDeviceSync.addSyncListener('DeviceOnline', handleDeviceOnline);
    multiDeviceSync.addSyncListener('DeviceOffline', handleDeviceOffline);
    multiDeviceSync.addSyncListener('DataUpdate', handleDataUpdate);

    // 定期更新在线设备列表
    const deviceUpdateInterval = setInterval(updateOnlineDevices, 30000);

    return () => {
      multiDeviceSync.removeSyncListener('statusChange');
      multiDeviceSync.removeSyncListener('DeviceOnline');
      multiDeviceSync.removeSyncListener('DeviceOffline');
      multiDeviceSync.removeSyncListener('DataUpdate');
      clearInterval(deviceUpdateInterval);
    };
  }, [userId]);

  const initializeSync = async () => {
    try {
      setErrorMessage(null);
      await multiDeviceSync.initialize(userId);
      setIsInitialized(true);
      await updateOnlineDevices();
    } catch (error) {
      console.error('初始化同步失败:', error);
      setErrorMessage('同步服务初始化失败');
      setIsInitialized(false);
    }
  };

  const updateOnlineDevices = async () => {
    try {
      const devices = await multiDeviceSync.getOnlineDevices();
      setOnlineDevices(devices);
    } catch (error) {
      console.error('获取在线设备失败:', error);
    }
  };

  const handleStatusChange = (data: { status: SyncStatus }) => {
    setSyncStatus(data.status);
    if (data.status === SyncStatus.Connected) {
      setLastSyncTime(new Date());
      setErrorMessage(null);
    }
  };

  const handleDeviceOnline = (data: any) => {
    console.log('设备上线:', data);
    updateOnlineDevices();
  };

  const handleDeviceOffline = (data: any) => {
    console.log('设备下线:', data);
    updateOnlineDevices();
  };

  const handleDataUpdate = (data: any) => {
    console.log('数据更新:', data);
    setLastSyncTime(new Date());
  };

  const handleManualSync = async () => {
    try {
      setErrorMessage(null);
      setSyncProgress(0);
      
      // 模拟同步进度
      const progressInterval = setInterval(() => {
        setSyncProgress(prev => {
          if (prev >= 90) {
            clearInterval(progressInterval);
            return 90;
          }
          return prev + 10;
        });
      }, 200);

      await multiDeviceSync.manualSync();
      
      clearInterval(progressInterval);
      setSyncProgress(100);
      setLastSyncTime(new Date());
      
      setTimeout(() => setSyncProgress(0), 2000);
    } catch (error) {
      console.error('手动同步失败:', error);
      setErrorMessage('同步失败，请检查网络连接');
      setSyncProgress(0);
    }
  };

  const getStatusColor = (status: SyncStatus): string => {
    switch (status) {
      case SyncStatus.Connected:
        return 'text-green-500';
      case SyncStatus.Connecting:
      case SyncStatus.Syncing:
        return 'text-yellow-500';
      case SyncStatus.Error:
        return 'text-red-500';
      default:
        return 'text-gray-500';
    }
  };

  const getStatusText = (status: SyncStatus): string => {
    switch (status) {
      case SyncStatus.Connected:
        return '已连接';
      case SyncStatus.Connecting:
        return '连接中...';
      case SyncStatus.Syncing:
        return '同步中...';
      case SyncStatus.Error:
        return '连接错误';
      case SyncStatus.Disconnected:
        return '未连接';
      default:
        return '未知状态';
    }
  };

  const getDeviceIcon = (deviceType: DeviceType): string => {
    switch (deviceType) {
      case DeviceType.Desktop:
        return '🖥️';
      case DeviceType.Mobile:
        return '📱';
      case DeviceType.Watch:
        return '⌚';
      case DeviceType.Tablet:
        return '📱';
      default:
        return '📱';
    }
  };

  const formatLastSync = (date: Date | null): string => {
    if (!date) return '从未同步';
    
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    
    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}小时前`;
    
    const days = Math.floor(hours / 24);
    return `${days}天前`;
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800">多设备同步</h2>
        <div className="flex items-center space-x-4">
          <div className={`flex items-center space-x-2 ${getStatusColor(syncStatus)}`}>
            <div className="w-2 h-2 rounded-full bg-current animate-pulse"></div>
            <span className="text-sm font-medium">{getStatusText(syncStatus)}</span>
          </div>
          {isInitialized && (
            <button
              onClick={handleManualSync}
              disabled={syncStatus === SyncStatus.Syncing || syncStatus === SyncStatus.Connecting}
              className="px-3 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
            >
              手动同步
            </button>
          )}
        </div>
      </div>

      {errorMessage && (
        <div className="mb-4 p-3 bg-red-100 border border-red-300 rounded-md">
          <p className="text-red-700 text-sm">{errorMessage}</p>
          <button
            onClick={initializeSync}
            className="mt-2 text-sm text-red-600 hover:text-red-800 underline"
          >
            重试连接
          </button>
        </div>
      )}

      {syncProgress > 0 && (
        <div className="mb-4">
          <div className="flex justify-between text-sm text-gray-600 mb-1">
            <span>同步进度</span>
            <span>{syncProgress}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-500 h-2 rounded-full transition-all duration-300"
              style={{ width: `${syncProgress}%` }}
            ></div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* 同步状态信息 */}
        <div className="space-y-4">
          <h3 className="text-lg font-medium text-gray-700">同步信息</h3>
          
          <div className="space-y-2">
            <div className="flex justify-between">
              <span className="text-gray-600">状态:</span>
              <span className={`font-medium ${getStatusColor(syncStatus)}`}>
                {getStatusText(syncStatus)}
              </span>
            </div>
            
            <div className="flex justify-between">
              <span className="text-gray-600">最后同步:</span>
              <span className="text-gray-800">{formatLastSync(lastSyncTime)}</span>
            </div>
            
            <div className="flex justify-between">
              <span className="text-gray-600">在线设备:</span>
              <span className="text-gray-800">{onlineDevices.length} 台</span>
            </div>
          </div>

          {!isInitialized && (
            <div className="mt-4">
              <button
                onClick={initializeSync}
                className="w-full px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                启动同步服务
              </button>
            </div>
          )}
        </div>

        {/* 在线设备列表 */}
        <div className="space-y-4">
          <h3 className="text-lg font-medium text-gray-700">在线设备</h3>
          
          {onlineDevices.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <p>暂无其他在线设备</p>
            </div>
          ) : (
            <div className="space-y-3">
              {onlineDevices.map((device) => (
                <div
                  key={device.device_id}
                  className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                >
                  <div className="flex items-center space-x-3">
                    <span className="text-2xl">{getDeviceIcon(device.device_type)}</span>
                    <div>
                      <p className="font-medium text-gray-800">{device.device_name}</p>
                      <p className="text-sm text-gray-600">
                        {device.platform} • {device.device_type}
                      </p>
                    </div>
                  </div>
                  
                  <div className="text-right">
                    <div className="flex items-center space-x-1">
                      <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                      <span className="text-sm text-green-600">在线</span>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                      {formatLastSync(new Date(device.last_sync))}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* 同步统计信息 */}
      <div className="mt-6 pt-6 border-t border-gray-200">
        <h3 className="text-lg font-medium text-gray-700 mb-4">同步统计</h3>
        
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center p-3 bg-blue-50 rounded-lg">
            <p className="text-2xl font-bold text-blue-600">0</p>
            <p className="text-sm text-gray-600">待同步消息</p>
          </div>
          
          <div className="text-center p-3 bg-green-50 rounded-lg">
            <p className="text-2xl font-bold text-green-600">0</p>
            <p className="text-sm text-gray-600">同步冲突</p>
          </div>
          
          <div className="text-center p-3 bg-yellow-50 rounded-lg">
            <p className="text-2xl font-bold text-yellow-600">0</p>
            <p className="text-sm text-gray-600">离线操作</p>
          </div>
          
          <div className="text-center p-3 bg-purple-50 rounded-lg">
            <p className="text-2xl font-bold text-purple-600">
              {onlineDevices.length}
            </p>
            <p className="text-sm text-gray-600">活跃设备</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MultiDeviceSync;