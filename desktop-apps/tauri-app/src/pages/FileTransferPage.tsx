import { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { open } from '@tauri-apps/plugin-dialog';
import {
  Wifi,
  WifiOff,
  Send,
  Pause,
  Play,
  Square,
  Monitor,
  Shield,
  User,
  Folder,
  Upload
} from 'lucide-react';

interface DeviceInfo {
  device_id: string;
  device_name: string;
  device_type: string;
  ip_address: string;
  port: number;
  protocol: string;
  is_trusted: boolean;
  last_seen: string;
}

interface TransferTask {
  task_id: string;
  file_name: string;
  file_size: number;
  source_device: string;
  target_device: string;
  status: 'pending' | 'transferring' | 'completed' | 'failed' | 'paused';
  progress: number;
  speed: number;
  created_at: string;
}

interface AccountInfo {
  account_id: string;
  username: string;
  device_name: string;
  is_online: boolean;
}

export default function FileTransferPage() {
  const [tabValue, setTabValue] = useState(0);
  const [isDiscovering, setIsDiscovering] = useState(false);
  const [devices, setDevices] = useState<DeviceInfo[]>([]);
  const [transferTasks, setTransferTasks] = useState<TransferTask[]>([]);
  const [currentAccount, setCurrentAccount] = useState<AccountInfo | null>(null);
  const [selectedFile, setSelectedFile] = useState<string>('');
  const [transferDialog, setTransferDialog] = useState(false);
  const [selectedDevice, setSelectedDevice] = useState<DeviceInfo | null>(null);
  const [trustDialog, setTrustDialog] = useState(false);
  const [deviceToTrust, setDeviceToTrust] = useState<DeviceInfo | null>(null);
  const [alert, setAlert] = useState<{ type: 'success' | 'error' | 'info'; message: string } | null>(null);

  useEffect(() => {
    loadAccountInfo();
    loadTransferTasks();
  }, []);

  useEffect(() => {
    const interval = setInterval(() => {
      if (isDiscovering) {
        refreshDeviceList();
      }
      refreshTransferTasks();
    }, 3000);

    return () => clearInterval(interval);
  }, [isDiscovering]);

  const handleTabChange = (newValue: number) => {
    setTabValue(newValue);
  };

  const loadAccountInfo = async () => {
    try {
      const account = await invoke<AccountInfo>('get_account_info');
      setCurrentAccount(account);
    } catch (error) {
      console.error('Failed to load account info:', error);
      // 模拟数据
      setCurrentAccount({
        account_id: 'user_001',
        username: '用户001',
        device_name: '我的设备',
        is_online: true
      });
    }
  };

  const loadTransferTasks = async () => {
    try {
      const tasks = await invoke<TransferTask[]>('get_transfer_tasks');
      setTransferTasks(tasks);
    } catch (error) {
      console.error('Failed to load transfer tasks:', error);
      // 模拟数据
      setTransferTasks([
        {
          task_id: 'task_001',
          file_name: 'document.pdf',
          file_size: 2048000,
          source_device: '我的设备',
          target_device: '设备A',
          status: 'completed',
          progress: 100,
          speed: 0,
          created_at: new Date().toISOString()
        }
      ]);
    }
  };

  const startDeviceDiscovery = async () => {
    setIsDiscovering(true);
    try {
      await invoke('start_device_discovery');
      refreshDeviceList();
    } catch (error) {
      console.error('Failed to start device discovery:', error);
      setAlert({ type: 'error', message: '启动设备搜索失败' });
    }
  };

  const stopDeviceDiscovery = async () => {
    setIsDiscovering(false);
    try {
      await invoke('stop_device_discovery');
    } catch (error) {
      console.error('Failed to stop device discovery:', error);
    }
  };

  const refreshDeviceList = async () => {
    try {
      const deviceList = await invoke<DeviceInfo[]>('get_discovered_devices');
      setDevices(deviceList);
    } catch (error) {
      console.error('Failed to refresh device list:', error);
      // 模拟数据
      setDevices([
        {
          device_id: 'device_001',
          device_name: '设备A',
          device_type: 'desktop',
          ip_address: '192.168.1.100',
          port: 8080,
          protocol: 'tcp',
          is_trusted: true,
          last_seen: new Date().toISOString()
        },
        {
          device_id: 'device_002',
          device_name: '设备B',
          device_type: 'mobile',
          ip_address: '192.168.1.101',
          port: 8080,
          protocol: 'tcp',
          is_trusted: false,
          last_seen: new Date().toISOString()
        }
      ]);
    }
  };

  const refreshTransferTasks = async () => {
    try {
      const tasks = await invoke<TransferTask[]>('get_transfer_tasks');
      setTransferTasks(tasks);
    } catch (error) {
      console.error('Failed to refresh transfer tasks:', error);
    }
  };

  const selectFile = async () => {
    try {
      const selected = await open({
        multiple: false,
        filters: [{
          name: 'All Files',
          extensions: ['*']
        }]
      });
      
      if (selected && typeof selected === 'string') {
        setSelectedFile(selected);
      }
    } catch (error) {
      console.error('Failed to select file:', error);
    }
  };

  const sendFile = async () => {
    if (!selectedDevice || !selectedFile) return;

    try {
      await invoke('send_file', {
        deviceId: selectedDevice.device_id,
        filePath: selectedFile
      });
      setAlert({ type: 'success', message: '文件发送已开始' });
      setTransferDialog(false);
      refreshTransferTasks();
    } catch (error) {
      console.error('Failed to send file:', error);
      setAlert({ type: 'error', message: '文件发送失败' });
    }
  };

  const addTrustedDevice = async () => {
    if (!deviceToTrust) return;

    try {
      await invoke('add_trusted_device', {
        deviceId: deviceToTrust.device_id
      });
      setAlert({ type: 'success', message: '设备已添加到信任列表' });
      setTrustDialog(false);
      refreshDeviceList();
    } catch (error) {
      console.error('Failed to add trusted device:', error);
      setAlert({ type: 'error', message: '添加信任设备失败' });
    }
  };

  const removeTrustedDevice = async (deviceId: string) => {
    try {
      await invoke('remove_trusted_device', { deviceId });
      setAlert({ type: 'success', message: '设备已从信任列表移除' });
      refreshDeviceList();
    } catch (error) {
      console.error('Failed to remove trusted device:', error);
      setAlert({ type: 'error', message: '移除信任设备失败' });
    }
  };

  const pauseTransfer = async (taskId: string) => {
    try {
      await invoke('pause_transfer', { taskId });
      refreshTransferTasks();
    } catch (error) {
      console.error('Failed to pause transfer:', error);
    }
  };

  const resumeTransfer = async (taskId: string) => {
    try {
      await invoke('resume_transfer', { taskId });
      refreshTransferTasks();
    } catch (error) {
      console.error('Failed to resume transfer:', error);
    }
  };

  const cancelTransfer = async (taskId: string) => {
    try {
      await invoke('cancel_transfer', { taskId });
      refreshTransferTasks();
    } catch (error) {
      console.error('Failed to cancel transfer:', error);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatSpeed = (bytesPerSecond: number): string => {
    return formatFileSize(bytesPerSecond) + '/s';
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">文件传输</h1>

      {/* 警告信息 */}
      {alert && (
        <div className={`mb-4 p-4 rounded-lg ${
          alert.type === 'success' ? 'bg-green-100 text-green-800' :
          alert.type === 'error' ? 'bg-red-100 text-red-800' :
          'bg-blue-100 text-blue-800'
        }`}>
          {alert.message}
          <button 
            onClick={() => setAlert(null)}
            className="ml-4 text-sm underline"
          >
            关闭
          </button>
        </div>
      )}

      {/* 账号信息 */}
      {currentAccount && (
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <div className="flex items-center space-x-4">
            <div className="w-12 h-12 bg-blue-500 rounded-full flex items-center justify-center">
              <User className="w-6 h-6 text-white" />
            </div>
            <div>
              <h3 className="text-lg font-semibold">{currentAccount.username}</h3>
              <p className="text-gray-600">设备: {currentAccount.device_name}</p>
              <p className="text-gray-600">账号ID: {currentAccount.account_id}</p>
            </div>
          </div>
        </div>
      )}

      {/* 标签页 */}
      <div className="mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            {[
              { icon: Monitor, label: '设备发现', index: 0 },
              { icon: Upload, label: '传输任务', index: 1 },
              { icon: Shield, label: '安全设置', index: 2 }
            ].map((tab) => (
              <button
                key={tab.index}
                onClick={() => handleTabChange(tab.index)}
                className={`flex items-center space-x-2 py-2 px-1 border-b-2 font-medium text-sm ${
                  tabValue === tab.index
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <tab.icon className="w-5 h-5" />
                <span>{tab.label}</span>
              </button>
            ))}
          </nav>
        </div>
      </div>

      {/* 设备发现标签页 */}
      {tabValue === 0 && (
        <div>
          <div className="mb-6 flex flex-wrap gap-4 items-center">
            <button
              onClick={isDiscovering ? stopDeviceDiscovery : startDeviceDiscovery}
              className={`flex items-center space-x-2 px-4 py-2 rounded-lg font-medium ${
                isDiscovering
                  ? 'bg-red-500 hover:bg-red-600 text-white'
                  : 'bg-blue-500 hover:bg-blue-600 text-white'
              }`}
            >
              {isDiscovering ? <WifiOff className="w-5 h-5" /> : <Wifi className="w-5 h-5" />}
              <span>{isDiscovering ? '停止搜索' : '搜索设备'}</span>
            </button>
            
            <button
              onClick={selectFile}
              className="flex items-center space-x-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              <Folder className="w-5 h-5" />
              <span>选择文件</span>
            </button>
            
            {selectedFile && (
              <span className="text-gray-600">
                已选择: {selectedFile.split('/').pop()}
              </span>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {devices.map((device) => (
              <div key={device.device_id} className="bg-white rounded-lg shadow-md p-6">
                <div className="flex justify-between items-start mb-4">
                  <h3 className="text-lg font-semibold">{device.device_name}</h3>
                  {device.is_trusted && (
                    <span className="bg-green-100 text-green-800 text-xs px-2 py-1 rounded-full">
                      已信任
                    </span>
                  )}
                </div>
                
                <div className="space-y-2 text-sm text-gray-600 mb-4">
                  <p>类型: {device.device_type}</p>
                  <p>地址: {device.ip_address}:{device.port}</p>
                  <p>协议: {device.protocol}</p>
                  <p>最后发现: {new Date(device.last_seen).toLocaleString()}</p>
                </div>
                
                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      setSelectedDevice(device);
                      setTransferDialog(true);
                    }}
                    disabled={!selectedFile}
                    className="flex items-center space-x-1 px-3 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed text-sm"
                  >
                    <Send className="w-4 h-4" />
                    <span>发送文件</span>
                  </button>
                  
                  {device.is_trusted ? (
                    <button
                      onClick={() => removeTrustedDevice(device.device_id)}
                      className="px-3 py-2 border border-red-300 text-red-600 rounded hover:bg-red-50 text-sm"
                    >
                      取消信任
                    </button>
                  ) : (
                    <button
                      onClick={() => {
                        setDeviceToTrust(device);
                        setTrustDialog(true);
                      }}
                      className="px-3 py-2 border border-gray-300 text-gray-600 rounded hover:bg-gray-50 text-sm"
                    >
                      添加信任
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 传输任务标签页 */}
      {tabValue === 1 && (
        <div>
          <div className="bg-white rounded-lg shadow-md">
            <div className="p-6 border-b border-gray-200">
              <h3 className="text-lg font-semibold">传输任务</h3>
            </div>
            
            <div className="divide-y divide-gray-200">
              {transferTasks.map((task) => (
                <div key={task.task_id} className="p-6">
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <h4 className="font-medium">{task.file_name}</h4>
                      <p className="text-sm text-gray-600">
                        {formatFileSize(task.file_size)} • {task.source_device} → {task.target_device}
                      </p>
                    </div>
                    
                    <span className={`px-2 py-1 rounded-full text-xs ${
                      task.status === 'completed' ? 'bg-green-100 text-green-800' :
                      task.status === 'transferring' ? 'bg-blue-100 text-blue-800' :
                      task.status === 'failed' ? 'bg-red-100 text-red-800' :
                      task.status === 'paused' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {task.status === 'completed' ? '已完成' :
                       task.status === 'transferring' ? '传输中' :
                       task.status === 'failed' ? '失败' :
                       task.status === 'paused' ? '已暂停' : '等待中'}
                    </span>
                  </div>
                  
                  {task.status === 'transferring' && (
                    <div className="mb-4">
                      <div className="flex justify-between text-sm text-gray-600 mb-1">
                        <span>{task.progress}%</span>
                        <span>{formatSpeed(task.speed)}</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                          className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                          style={{ width: `${task.progress}%` }}
                        />
                      </div>
                    </div>
                  )}
                  
                  <div className="flex gap-2">
                    {task.status === 'transferring' && (
                      <button
                        onClick={() => pauseTransfer(task.task_id)}
                        className="flex items-center space-x-1 px-3 py-1 border border-gray-300 rounded hover:bg-gray-50 text-sm"
                      >
                        <Pause className="w-4 h-4" />
                        <span>暂停</span>
                      </button>
                    )}
                    
                    {task.status === 'paused' && (
                      <button
                        onClick={() => resumeTransfer(task.task_id)}
                        className="flex items-center space-x-1 px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
                      >
                        <Play className="w-4 h-4" />
                        <span>继续</span>
                      </button>
                    )}
                    
                    {(task.status === 'transferring' || task.status === 'paused' || task.status === 'pending') && (
                      <button
                        onClick={() => cancelTransfer(task.task_id)}
                        className="flex items-center space-x-1 px-3 py-1 border border-red-300 text-red-600 rounded hover:bg-red-50 text-sm"
                      >
                        <Square className="w-4 h-4" />
                        <span>取消</span>
                      </button>
                    )}
                  </div>
                </div>
              ))}
              
              {transferTasks.length === 0 && (
                <div className="p-12 text-center text-gray-500">
                  <Upload className="w-12 h-12 mx-auto mb-4 opacity-50" />
                  <p>暂无传输任务</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 安全设置标签页 */}
      {tabValue === 2 && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <h3 className="text-lg font-semibold mb-6">安全设置</h3>
          
          <div className="space-y-6">
            <div>
              <h4 className="font-medium mb-2">信任设备管理</h4>
              <p className="text-gray-600 text-sm mb-4">
                管理可以与此设备进行文件传输的信任设备列表
              </p>
              
              <div className="space-y-2">
                {devices.filter(d => d.is_trusted).map((device) => (
                  <div key={device.device_id} className="flex justify-between items-center p-3 border border-gray-200 rounded">
                    <div>
                      <span className="font-medium">{device.device_name}</span>
                      <span className="text-gray-600 text-sm ml-2">({device.device_type})</span>
                    </div>
                    <button
                      onClick={() => removeTrustedDevice(device.device_id)}
                      className="text-red-600 hover:text-red-800 text-sm"
                    >
                      移除
                    </button>
                  </div>
                ))}
                
                {devices.filter(d => d.is_trusted).length === 0 && (
                  <p className="text-gray-500 text-sm">暂无信任设备</p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 文件传输对话框 */}
      {transferDialog && selectedDevice && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold mb-4">确认文件传输</h3>
            
            <div className="space-y-3 mb-6">
              <p><strong>目标设备:</strong> {selectedDevice.device_name}</p>
              <p><strong>文件:</strong> {selectedFile.split('/').pop()}</p>
              <p><strong>设备地址:</strong> {selectedDevice.ip_address}:{selectedDevice.port}</p>
            </div>
            
            <div className="flex gap-3">
              <button
                onClick={sendFile}
                className="flex-1 bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600"
              >
                开始传输
              </button>
              <button
                onClick={() => setTransferDialog(false)}
                className="flex-1 border border-gray-300 py-2 px-4 rounded hover:bg-gray-50"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 信任设备对话框 */}
      {trustDialog && deviceToTrust && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold mb-4">添加信任设备</h3>
            
            <div className="space-y-3 mb-6">
              <p><strong>设备名称:</strong> {deviceToTrust.device_name}</p>
              <p><strong>设备类型:</strong> {deviceToTrust.device_type}</p>
              <p><strong>设备地址:</strong> {deviceToTrust.ip_address}:{deviceToTrust.port}</p>
              <p className="text-sm text-gray-600">
                添加到信任列表后，该设备将可以与您的设备进行文件传输。
              </p>
            </div>
            
            <div className="flex gap-3">
              <button
                onClick={addTrustedDevice}
                className="flex-1 bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600"
              >
                添加信任
              </button>
              <button
                onClick={() => setTrustDialog(false)}
                className="flex-1 border border-gray-300 py-2 px-4 rounded hover:bg-gray-50"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}