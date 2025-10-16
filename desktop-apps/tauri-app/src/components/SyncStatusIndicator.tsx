import React, { useState, useEffect } from 'react';
import { multiDeviceSync } from '../services/multiDeviceSync';
import { SyncStatus, DeviceInfo } from '../services/multiDeviceSync';

interface SyncStatusIndicatorProps {
  className?: string;
  showDetails?: boolean;
}

const SyncStatusIndicator: React.FC<SyncStatusIndicatorProps> = ({
  className = '',
  showDetails = false
}) => {
  const [syncStatus, setSyncStatus] = useState<SyncStatus>(SyncStatus.Disconnected);
  const [onlineDevices, setOnlineDevices] = useState<DeviceInfo[]>([]);
  const [lastSyncTime, setLastSyncTime] = useState<Date | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);

  useEffect(() => {
    // 监听同步状态变化
    const handleStatusChange = (status: SyncStatus) => {
      setSyncStatus(status);
    };

    const handleDevicesUpdate = (devices: DeviceInfo[]) => {
      setOnlineDevices(devices);
    };

    const handleSyncComplete = () => {
      setLastSyncTime(new Date());
    };

    // 注册事件监听器
    multiDeviceSync.on('statusChange', handleStatusChange);
    multiDeviceSync.on('devicesUpdate', handleDevicesUpdate);
    multiDeviceSync.on('syncComplete', handleSyncComplete);

    // 初始化状态
    setSyncStatus(multiDeviceSync.getStatus());
    multiDeviceSync.getOnlineDevices().then(devices => {
      setOnlineDevices(devices);
    }).catch(error => {
      console.error('Failed to get online devices:', error);
    });

    return () => {
      // 清理事件监听器
      multiDeviceSync.off('statusChange');
      multiDeviceSync.off('devicesUpdate');
      multiDeviceSync.off('syncComplete');
    };
  }, []);

  const getStatusIcon = () => {
    switch (syncStatus) {
      case SyncStatus.Connected:
        return '🟢';
      case SyncStatus.Syncing:
        return '🔄';
      case SyncStatus.Connecting:
        return '🟡';
      case SyncStatus.Disconnected:
        return '🔴';
      case SyncStatus.Error:
        return '❌';
      default:
        return '⚪';
    }
  };

  const getStatusText = () => {
    switch (syncStatus) {
      case SyncStatus.Connected:
        return '已连接';
      case SyncStatus.Syncing:
        return '同步中';
      case SyncStatus.Connecting:
        return '连接中';
      case SyncStatus.Disconnected:
        return '已断开';
      case SyncStatus.Error:
        return '错误';
      default:
        return '未知';
    }
  };

  const getStatusColor = () => {
    switch (syncStatus) {
      case SyncStatus.Connected:
        return 'text-green-600';
      case SyncStatus.Syncing:
        return 'text-blue-600';
      case SyncStatus.Connecting:
        return 'text-yellow-600';
      case SyncStatus.Disconnected:
        return 'text-red-600';
      case SyncStatus.Error:
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  const formatLastSyncTime = () => {
    if (!lastSyncTime) return '从未同步';
    
    const now = new Date();
    const diff = now.getTime() - lastSyncTime.getTime();
    const minutes = Math.floor(diff / 60000);
    
    if (minutes < 1) return '刚刚同步';
    if (minutes < 60) return `${minutes}分钟前`;
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}小时前`;
    
    const days = Math.floor(hours / 24);
    return `${days}天前`;
  };

  const handleManualSync = async () => {
    try {
      await multiDeviceSync.syncIncremental();
    } catch (error) {
      console.error('手动同步失败:', error);
    }
  };

  return (
    <div className={`sync-status-indicator ${className}`}>
      <div 
        className="flex items-center space-x-2 cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <span className="text-lg">{getStatusIcon()}</span>
        <span className={`text-sm font-medium ${getStatusColor()}`}>
          {getStatusText()}
        </span>
        {showDetails && (
          <span className="text-xs text-gray-500">
            ({onlineDevices.length} 设备在线)
          </span>
        )}
        <svg 
          className={`w-4 h-4 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
          fill="none" 
          stroke="currentColor" 
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </div>

      {isExpanded && (
        <div className="mt-3 p-3 bg-gray-50 rounded-lg border">
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">最后同步:</span>
              <span className="text-sm font-medium">{formatLastSyncTime()}</span>
            </div>
            
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600">在线设备:</span>
              <span className="text-sm font-medium">{onlineDevices.length}</span>
            </div>

            {onlineDevices.length > 0 && (
              <div className="mt-2">
                <div className="text-xs text-gray-500 mb-1">设备列表:</div>
                <div className="space-y-1">
                  {onlineDevices.map((device) => (
                    <div key={device.device_id} className="flex items-center space-x-2 text-xs">
                      <span className="w-2 h-2 bg-green-400 rounded-full"></span>
                      <span className="font-medium">{device.device_name}</span>
                      <span className="text-gray-500">({device.device_type})</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="flex space-x-2 mt-3">
              <button
                onClick={handleManualSync}
                disabled={syncStatus === SyncStatus.Syncing}
                className="flex-1 px-3 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {syncStatus === SyncStatus.Syncing ? '同步中...' : '手动同步'}
              </button>
              
              <button
                onClick={() => setIsExpanded(false)}
                className="px-3 py-1 text-xs bg-gray-300 text-gray-700 rounded hover:bg-gray-400"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SyncStatusIndicator;