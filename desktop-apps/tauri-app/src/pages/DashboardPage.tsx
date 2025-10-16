import React, { useState, useEffect } from 'react';
import {
  LayoutDashboard as DashboardIcon,
  Users as People,
  MessageCircle as Chat,
  TrendingUp,
  Monitor as Devices,
  Shield as Security,
  Bell as Notifications,
  RefreshCw as Refresh
} from 'lucide-react';
import { useDynamicMenu } from '../contexts/DynamicMenuContext';
import { useAuth } from '../contexts/AuthContext';
import { DeviceType, PermissionLevel } from '../types/menu';

interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  totalMessages: number;
  todayMessages: number;
  deviceDistribution: {
    [key in DeviceType]: number;
  };
  permissionDistribution: {
    [key in PermissionLevel]: number;
  };
}

const DashboardPage: React.FC = () => {
  const { user } = useAuth();
  const { deviceInfo, getFilteredMenuItems } = useDynamicMenu();
  const [stats, setStats] = useState<DashboardStats>({
    totalUsers: 0,
    activeUsers: 0,
    totalMessages: 0,
    todayMessages: 0,
    deviceDistribution: {
      [DeviceType.DESKTOP]: 0,
      [DeviceType.MOBILE]: 0,
      [DeviceType.TABLET]: 0,
      [DeviceType.WATCH]: 0
    },
    permissionDistribution: {
      [PermissionLevel.GUEST]: 0,
      [PermissionLevel.USER]: 0,
      [PermissionLevel.MODERATOR]: 0,
      [PermissionLevel.ADMIN]: 0,
      [PermissionLevel.SUPER_ADMIN]: 0
    }
  });
  const [loading, setLoading] = useState(true);

  // 模拟数据加载
  useEffect(() => {
    const loadDashboardData = async () => {
      setLoading(true);
      
      // 模拟API调用延迟
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 模拟统计数据
      setStats({
        totalUsers: 1250,
        activeUsers: 342,
        totalMessages: 15680,
        todayMessages: 234,
        deviceDistribution: {
          [DeviceType.DESKTOP]: 45,
          [DeviceType.MOBILE]: 35,
          [DeviceType.TABLET]: 15,
          [DeviceType.WATCH]: 5
        },
        permissionDistribution: {
          [PermissionLevel.GUEST]: 10,
          [PermissionLevel.USER]: 70,
          [PermissionLevel.MODERATOR]: 15,
          [PermissionLevel.ADMIN]: 4,
          [PermissionLevel.SUPER_ADMIN]: 1
        }
      });
      
      setLoading(false);
    };

    loadDashboardData();
  }, []);

  const refreshData = () => {
    setLoading(true);
    setTimeout(() => setLoading(false), 1000);
  };

  const menuItems = getFilteredMenuItems();

  const StatCard: React.FC<{
    title: string;
    value: string | number;
    icon: React.ReactElement;
    color: string;
    subtitle?: string;
    trend?: number;
  }> = ({ title, value, icon, color, subtitle, trend }) => (
    <div className="bg-white rounded-lg shadow-md p-6 h-full">
      <div className="flex items-center mb-4">
        <div 
          className="w-12 h-12 rounded-full flex items-center justify-center mr-4"
          style={{ backgroundColor: color }}
        >
          <div className="text-white">
            {React.cloneElement(icon, { className: "h-6 w-6" })}
          </div>
        </div>
        <div className="flex-1">
          <div className="text-3xl font-bold text-gray-900">
            {loading ? '-' : value}
          </div>
          <div className="text-sm text-gray-500">
            {title}
          </div>
        </div>
        {trend && (
          <div className="bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs font-medium flex items-center">
            <TrendingUp className="h-3 w-3 mr-1" />
            +{trend}%
          </div>
        )}
      </div>
      {subtitle && (
        <div className="text-xs text-gray-400">
          {subtitle}
        </div>
      )}
      {loading && (
        <div className="w-full bg-gray-200 rounded-full h-1 mt-2">
          <div className="bg-blue-600 h-1 rounded-full animate-pulse" style={{ width: '60%' }}></div>
        </div>
      )}
    </div>
  );

  return (
    <div className="p-6">
      {/* 页面标题 */}
      <div className="flex items-center mb-6">
        <DashboardIcon className="mr-4 h-8 w-8 text-gray-700" />
        <h1 className="text-3xl font-bold text-gray-900">
          仪表板
        </h1>
        <div className="flex-1"></div>
        <button 
          onClick={refreshData} 
          disabled={loading}
          className="p-2 rounded-lg hover:bg-gray-100 transition-colors disabled:opacity-50"
        >
          <Refresh className={`h-5 w-5 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* 欢迎信息 */}
      <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg p-6 mb-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-2">
          欢迎回来，{user?.username || '用户'}！
        </h2>
        <p className="text-gray-600">
          当前设备：{deviceInfo.type} | 屏幕尺寸：{deviceInfo.screenSize.width}x{deviceInfo.screenSize.height}
        </p>
        <p className="text-sm text-gray-500 mt-1">
          您有 {menuItems.length} 个可用菜单项
        </p>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
        <StatCard
          title="总用户数"
          value={stats.totalUsers}
          icon={<People />}
          color="#3b82f6"
          trend={12}
        />
        <StatCard
          title="活跃用户"
          value={stats.activeUsers}
          icon={<TrendingUp />}
          color="#10b981"
          subtitle="过去24小时"
        />
        <StatCard
          title="总消息数"
          value={stats.totalMessages}
          icon={<Chat />}
          color="#06b6d4"
          trend={8}
        />
        <StatCard
          title="今日消息"
          value={stats.todayMessages}
          icon={<Notifications />}
          color="#f59e0b"
          subtitle="实时更新"
        />
      </div>

      {/* 详细信息 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 设备分布 */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center mb-4">
            <Devices className="h-5 w-5 mr-2 text-gray-600" />
            <h3 className="text-lg font-semibold text-gray-900">设备分布</h3>
          </div>
          <div className="space-y-4">
            {Object.entries(stats.deviceDistribution).map(([device, percentage]) => (
              <div key={device} className="flex items-center">
                <div className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center mr-4">
                  <Devices className="h-5 w-5 text-white" />
                </div>
                <div className="flex-1">
                  <div className="font-medium text-gray-900">
                    {device.toUpperCase()}
                  </div>
                  <div className="flex items-center mt-1">
                    <div className="flex-1 bg-gray-200 rounded-full h-2 mr-4">
                      <div 
                        className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                        style={{ width: `${percentage}%` }}
                      ></div>
                    </div>
                    <span className="text-sm text-gray-600">{percentage}%</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 权限分布 */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center mb-4">
            <Security className="h-5 w-5 mr-2 text-gray-600" />
            <h3 className="text-lg font-semibold text-gray-900">权限分布</h3>
          </div>
          <div className="space-y-4">
            {Object.entries(stats.permissionDistribution).map(([permission, percentage]) => (
              <div key={permission} className="flex items-center">
                <div className="w-10 h-10 bg-purple-600 rounded-full flex items-center justify-center mr-4">
                  <Security className="h-5 w-5 text-white" />
                </div>
                <div className="flex-1">
                  <div className="font-medium text-gray-900">
                    {permission.replace('_', ' ').toUpperCase()}
                  </div>
                  <div className="flex items-center mt-1">
                    <div className="flex-1 bg-gray-200 rounded-full h-2 mr-4">
                      <div 
                        className="bg-purple-600 h-2 rounded-full transition-all duration-300"
                        style={{ width: `${percentage}%` }}
                      ></div>
                    </div>
                    <span className="text-sm text-gray-600">{percentage}%</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;