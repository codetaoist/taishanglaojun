import { useState, useEffect } from 'react';
import { DeviceType, PermissionLevel } from '../types/menu';
import { useDynamicMenu } from '../contexts/DynamicMenuContext';
import { getDeviceInfo } from '../utils/deviceDetection';

interface TestUser {
  id: string;
  name: string;
  permissions: PermissionLevel[];
  deviceType: DeviceType;
}

const testUsers: TestUser[] = [
  {
    id: 'admin',
    name: '管理员',
    permissions: [PermissionLevel.ADMIN, PermissionLevel.MODERATOR, PermissionLevel.USER],
    deviceType: DeviceType.DESKTOP
  },
  {
    id: 'editor',
    name: '编辑者',
    permissions: [PermissionLevel.MODERATOR, PermissionLevel.USER],
    deviceType: DeviceType.TABLET
  },
  {
    id: 'viewer',
    name: '查看者',
    permissions: [PermissionLevel.USER],
    deviceType: DeviceType.MOBILE
  },
  {
    id: 'watch_user',
    name: '手表用户',
    permissions: [PermissionLevel.GUEST],
    deviceType: DeviceType.WATCH
  }
];

export default function MenuAdaptationTestPage() {
  const { menuItems } = useDynamicMenu();
  const [selectedUser, setSelectedUser] = useState<TestUser>(testUsers[0]);
  const [selectedDeviceType, setSelectedDeviceType] = useState<DeviceType>(DeviceType.DESKTOP);
  const [showHidden, setShowHidden] = useState(false);
  const [filteredMenus, setFilteredMenus] = useState<any[]>([]);
  
  const [mockDeviceInfo] = useState(getDeviceInfo());

  // 本地过滤函数
  const filterMenuItems = (deviceType: DeviceType, userPermissions: PermissionLevel[]) => {
    return menuItems.filter(item => {
      // 检查设备支持
      const deviceSupported = item.supportedDevices?.includes(deviceType);
      // 检查权限
      const hasPermission = item.requiredPermissions?.some(perm => userPermissions.includes(perm));
      return deviceSupported && hasPermission;
    });
  };

  useEffect(() => {
    // 根据选择的用户和设备类型过滤菜单
    const filtered = filterMenuItems(selectedDeviceType, selectedUser.permissions);
    setFilteredMenus(filtered);
  }, [selectedUser, selectedDeviceType, menuItems]);

  const handleUserChange = (userId: string) => {
    const user = testUsers.find(u => u.id === userId);
    if (user) {
      setSelectedUser(user);
      setSelectedDeviceType(user.deviceType);
    }
  };

  const getPermissionColor = (permission: PermissionLevel) => {
    switch (permission) {
      case PermissionLevel.ADMIN: return 'bg-red-100 text-red-800';
      case PermissionLevel.MODERATOR: return 'bg-yellow-100 text-yellow-800';
      case PermissionLevel.USER: return 'bg-green-100 text-green-800';
      default: return 'bg-gray-100 text-gray-600';
    }
  };

  const getDeviceTypeColor = (deviceType: DeviceType) => {
    switch (deviceType) {
      case DeviceType.DESKTOP: return 'bg-blue-100 text-blue-800';
      case DeviceType.TABLET: return 'bg-purple-100 text-purple-800';
      case DeviceType.MOBILE: return 'bg-cyan-100 text-cyan-800';
      case DeviceType.WATCH: return 'bg-orange-100 text-orange-800';
      default: return 'bg-gray-100 text-gray-600';
    }
  };

  const isMenuItemVisible = (item: any) => {
    // 检查设备类型支持
    if (!item.supportedDevices?.includes(selectedDeviceType)) {
      return false;
    }

    // 检查权限
    if (!item.requiredPermissions?.some((perm: any) => selectedUser.permissions.includes(perm))) {
      return false;
    }

    return item.isVisible;
  };

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">
        菜单适配测试
      </h1>
      
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
        <p className="text-blue-800">
          此页面用于测试基于用户权限和设备类型的动态菜单适配功能。
          选择不同的用户和设备类型来查看菜单如何自动适配。
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 控制面板 */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">
              测试控制
            </h2>
            
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                选择用户
              </label>
              <select
                value={selectedUser.id}
                onChange={(e) => handleUserChange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {testUsers.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                设备类型
              </label>
              <select
                value={selectedDeviceType}
                onChange={(e) => setSelectedDeviceType(e.target.value as DeviceType)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="DESKTOP">桌面端</option>
                <option value="TABLET">平板</option>
                <option value="MOBILE">手机</option>
                <option value="WATCH">手表</option>
              </select>
            </div>

            <div className="mb-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={showHidden}
                  onChange={(e) => setShowHidden(e.target.checked)}
                  className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <span className="text-sm text-gray-700">显示隐藏项目</span>
              </label>
            </div>

            <div className="border-t border-gray-200 pt-4 mb-4">
              <h3 className="text-sm font-medium text-gray-700 mb-2">
                当前用户权限:
              </h3>
              <div className="flex flex-wrap gap-2 mb-4">
                {selectedUser.permissions.map((permission) => (
                  <span
                    key={permission}
                    className={`px-2 py-1 rounded-full text-xs font-medium ${getPermissionColor(permission)}`}
                  >
                    {permission}
                  </span>
                ))}
              </div>

              <h3 className="text-sm font-medium text-gray-700 mb-2">
                设备信息:
              </h3>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="text-sm text-gray-600 mb-1">
                  类型: <span className={`px-2 py-1 rounded-full text-xs font-medium ${getDeviceTypeColor(selectedDeviceType)}`}>
                    {selectedDeviceType}
                  </span>
                </div>
                <div className="text-sm text-gray-600 mb-1">
                  屏幕: {mockDeviceInfo.screenSize.width} x {mockDeviceInfo.screenSize.height}
                </div>
                <div className="text-sm text-gray-600 mb-1">
                  触摸: {mockDeviceInfo.isTouchDevice ? '支持' : '不支持'}
                </div>
                <div className="text-sm text-gray-600">
                  移动设备: {mockDeviceInfo.type === DeviceType.MOBILE ? '是' : '否'}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 菜单显示 */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">
              适配后的菜单 ({filteredMenus.length} 项)
            </h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {menuItems.map((item) => {
                const isVisible = isMenuItemVisible(item);
                const isInFiltered = filteredMenus.some(f => f.id === item.id);
                
                if (!isVisible && !showHidden) return null;

                return (
                  <div
                    key={item.id}
                    className={`p-4 rounded-lg border-2 ${
                      isInFiltered ? 'border-blue-500' : 'border-gray-300'
                    } ${isVisible ? 'bg-white' : 'bg-gray-100 opacity-50'}`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="font-semibold text-gray-900 flex-1">
                        {item.title}
                      </h3>
                      {!isVisible && (
                        <span className="px-2 py-1 bg-red-100 text-red-800 text-xs font-medium rounded-full">
                          隐藏
                        </span>
                      )}
                    </div>
                    
                    <p className="text-sm text-gray-600 mb-2">
                      路由: {item.route}
                    </p>
                    
                    <div className="mb-2">
                       <span className="text-xs text-gray-500 block mb-1">
                         权限要求:
                       </span>
                       <div className="flex flex-wrap gap-1">
                         {item.requiredPermissions?.map((permission: any) => (
                           <span key={permission} className={`px-2 py-1 rounded-full text-xs font-medium ${getPermissionColor(permission)}`}>
                             {permission}
                           </span>
                         ))}
                       </div>
                     </div>
                    
                    <div className="mb-3">
                       <span className="text-xs text-gray-500 block mb-1">
                         支持设备:
                       </span>
                       <div className="flex flex-wrap gap-1">
                         {item.supportedDevices?.map((deviceType: any) => (
                           <span
                             key={deviceType}
                             className={`px-2 py-1 rounded-full text-xs font-medium ${
                               deviceType === selectedDeviceType 
                                 ? 'bg-blue-100 text-blue-800' 
                                 : 'bg-gray-100 text-gray-600'
                             }`}
                           >
                             {deviceType}
                           </span>
                         ))}
                       </div>
                     </div>

                    <div className="flex flex-wrap gap-2">
                       {/* 权限检查 */}
                       <span
                         className={`px-2 py-1 rounded-full text-xs font-medium ${
                           item.requiredPermissions?.some((perm: any) => selectedUser.permissions.includes(perm))
                             ? 'bg-green-100 text-green-800'
                             : 'bg-red-100 text-red-800'
                         }`}
                       >
                         {item.requiredPermissions?.some((perm: any) => selectedUser.permissions.includes(perm)) ? '权限✓' : '权限✗'}
                       </span>
                       
                       {/* 设备支持检查 */}
                       <span
                         className={`px-2 py-1 rounded-full text-xs font-medium ${
                           item.supportedDevices?.includes(selectedDeviceType)
                             ? 'bg-green-100 text-green-800'
                             : 'bg-red-100 text-red-800'
                         }`}
                       >
                         {item.supportedDevices?.includes(selectedDeviceType) ? '设备✓' : '设备✗'}
                       </span>
                     </div>
                  </div>
                );
              })}
            </div>

            {filteredMenus.length === 0 && (
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mt-4">
                <p className="text-yellow-800">
                  当前用户和设备类型组合下没有可用的菜单项。
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 统计信息 */}
      <div className="bg-white rounded-lg shadow-md p-6 mt-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">
          适配统计
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-gray-50 rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-blue-600">
              {menuItems.length}
            </div>
            <div className="text-sm text-gray-600">
              总菜单项
            </div>
          </div>
          <div className="bg-gray-50 rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-green-600">
              {filteredMenus.length}
            </div>
            <div className="text-sm text-gray-600">
              可用菜单项
            </div>
          </div>
          <div className="bg-gray-50 rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-orange-600">
              {menuItems.filter(item => !item.supportedDevices?.includes(selectedDeviceType)).length}
            </div>
            <div className="text-sm text-gray-600">
              设备不支持
            </div>
          </div>
          <div className="bg-gray-50 rounded-lg p-4 text-center">
            <div className="text-2xl font-bold text-red-600">
              {menuItems.filter(item => !item.requiredPermissions?.some((perm: any) => selectedUser.permissions.includes(perm))).length}
            </div>
            <div className="text-sm text-gray-600">
              权限不足
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}