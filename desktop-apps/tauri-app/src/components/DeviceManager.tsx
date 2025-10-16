import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { DeviceInfo, DeviceType } from '../services/multiDeviceSync';

interface DeviceManagerProps {
  userId: string;
  onDeviceSelect?: (device: DeviceInfo) => void;
}

const DeviceManager: React.FC<DeviceManagerProps> = ({ userId, onDeviceSelect }) => {
  const [devices, setDevices] = useState<DeviceInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddDevice, setShowAddDevice] = useState(false);
  const [newDevice, setNewDevice] = useState({
    device_name: '',
    device_type: DeviceType.Desktop,
    platform: '',
    version: ''
  });

  useEffect(() => {
    loadDevices();
  }, [userId]);

  const loadDevices = async () => {
    try {
      setLoading(true);
      setError(null);
      const userDevices = await invoke<DeviceInfo[]>('get_user_devices', { userId });
      setDevices(userDevices);
    } catch (err) {
      setError(err as string);
      console.error('加载设备列表失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleAddDevice = async () => {
    try {
      const deviceInfo: DeviceInfo = {
        device_id: '', // 后端会生成
        device_name: newDevice.device_name,
        device_type: newDevice.device_type,
        platform: newDevice.platform,
        app_version: newDevice.version,
        last_sync: new Date().toISOString(),
        is_online: false
      };

      await invoke('register_device', { deviceInfo });
      await loadDevices();
      setShowAddDevice(false);
      setNewDevice({
        device_name: '',
        device_type: DeviceType.Desktop,
        platform: '',
        version: ''
      });
    } catch (err) {
      setError(err as string);
      console.error('添加设备失败:', err);
    }
  };

  const handleRemoveDevice = async (deviceId: string) => {
    if (!confirm('确定要移除这个设备吗？')) return;

    try {
      // 这里需要实现移除设备的后端接口
      await invoke('remove_device', { deviceId });
      await loadDevices();
    } catch (err) {
      setError(err as string);
      console.error('移除设备失败:', err);
    }
  };

  const getDeviceIcon = (deviceType: DeviceType) => {
    switch (deviceType) {
      case DeviceType.Desktop:
        return '🖥️';
      case DeviceType.Mobile:
        return '📱';
      case DeviceType.Tablet:
        return '📱';
      case DeviceType.Watch:
        return '⌚';
      default:
        return '📱';
    }
  };

  const getDeviceTypeText = (deviceType: DeviceType) => {
    switch (deviceType) {
      case DeviceType.Desktop:
        return '桌面设备';
      case DeviceType.Mobile:
        return '手机';
      case DeviceType.Tablet:
        return '平板';
      case DeviceType.Watch:
        return '手表';
      default:
        return '未知设备';
    }
  };

  const formatLastSeen = (lastSeen: string) => {
    const date = new Date(lastSeen);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);

    if (minutes < 1) return '刚刚在线';
    if (minutes < 60) return `${minutes}分钟前`;
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}小时前`;
    
    const days = Math.floor(hours / 24);
    return `${days}天前`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span className="ml-2 text-gray-600">加载设备列表...</span>
      </div>
    );
  }

  return (
    <div className="device-manager">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-semibold text-gray-800">设备管理</h2>
        <button
          onClick={() => setShowAddDevice(true)}
          className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          添加设备
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {devices.map((device) => (
          <div
            key={device.device_id}
            className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow cursor-pointer"
            onClick={() => onDeviceSelect?.(device)}
          >
            <div className="flex items-start justify-between">
              <div className="flex items-center space-x-3">
                <span className="text-2xl">{getDeviceIcon(device.device_type)}</span>
                <div>
                  <h3 className="font-medium text-gray-900">{device.device_name}</h3>
                  <p className="text-sm text-gray-500">{getDeviceTypeText(device.device_type)}</p>
                </div>
              </div>
              
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${device.is_online ? 'bg-green-400' : 'bg-gray-300'}`}></div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRemoveDevice(device.device_id);
                  }}
                  className="text-gray-400 hover:text-red-500 transition-colors"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="mt-3 space-y-1">
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">平台:</span>
                <span className="text-gray-700">{device.platform}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">版本:</span>
                <span className="text-gray-700">{device.app_version}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">最后同步:</span>
                <span className="text-gray-700">{formatLastSeen(device.last_sync)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">在线状态:</span>
                <span className={`font-medium ${device.is_online ? 'text-green-600' : 'text-gray-500'}`}>
                  {device.is_online ? '在线' : '离线'}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {devices.length === 0 && !loading && (
        <div className="text-center py-8">
          <div className="text-gray-400 text-4xl mb-4">📱</div>
          <p className="text-gray-500">还没有注册的设备</p>
          <button
            onClick={() => setShowAddDevice(true)}
            className="mt-2 text-blue-500 hover:text-blue-600"
          >
            添加第一个设备
          </button>
        </div>
      )}

      {/* 添加设备模态框 */}
      {showAddDevice && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold mb-4">添加新设备</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  设备名称
                </label>
                <input
                  type="text"
                  value={newDevice.device_name}
                  onChange={(e) => setNewDevice({ ...newDevice, device_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如: 我的iPhone"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  设备类型
                </label>
                <select
                  value={newDevice.device_type}
                  onChange={(e) => setNewDevice({ ...newDevice, device_type: e.target.value as DeviceType })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value={DeviceType.Desktop}>桌面设备</option>
                  <option value={DeviceType.Mobile}>手机</option>
                  <option value={DeviceType.Tablet}>平板</option>
                  <option value={DeviceType.Watch}>手表</option>
                  <option value={DeviceType.Web}>网页</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  平台
                </label>
                <input
                  type="text"
                  value={newDevice.platform}
                  onChange={(e) => setNewDevice({ ...newDevice, platform: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如: iOS, Android, Windows"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  版本
                </label>
                <input
                  type="text"
                  value={newDevice.version}
                  onChange={(e) => setNewDevice({ ...newDevice, version: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="例如: 1.0.0"
                />
              </div>
            </div>

            <div className="flex space-x-3 mt-6">
              <button
                onClick={handleAddDevice}
                disabled={!newDevice.device_name.trim()}
                className="flex-1 px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                添加设备
              </button>
              <button
                onClick={() => setShowAddDevice(false)}
                className="flex-1 px-4 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default DeviceManager;