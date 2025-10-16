import React, { useState, useEffect } from 'react';
import {
  Users,
  Edit,
  Trash2,
  Plus
} from 'lucide-react';
import { PermissionLevel, DeviceType } from '../types/menu';

interface User {
  id: string;
  username: string;
  email: string;
  permissions: PermissionLevel[];
  deviceType: DeviceType;
  isActive: boolean;
  lastLogin: string;
  createdAt: string;
}

const UserManagementPage: React.FC = () => {

  const [users, setUsers] = useState<User[]>([]);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // 模拟用户数据
  useEffect(() => {
    const loadUsers = async () => {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const mockUsers: User[] = [
        {
          id: '1',
          username: 'admin',
          email: 'admin@example.com',
          permissions: [PermissionLevel.ADMIN, PermissionLevel.USER],
          deviceType: DeviceType.DESKTOP,
          isActive: true,
          lastLogin: '2024-01-15T10:30:00Z',
          createdAt: '2024-01-01T00:00:00Z'
        },
        {
          id: '2',
          username: 'user1',
          email: 'user1@example.com',
          permissions: [PermissionLevel.USER],
          deviceType: DeviceType.MOBILE,
          isActive: true,
          lastLogin: '2024-01-15T09:15:00Z',
          createdAt: '2024-01-05T00:00:00Z'
        },
        {
          id: '3',
          username: 'moderator',
          email: 'mod@example.com',
          permissions: [PermissionLevel.MODERATOR, PermissionLevel.USER],
          deviceType: DeviceType.TABLET,
          isActive: true,
          lastLogin: '2024-01-14T16:45:00Z',
          createdAt: '2024-01-03T00:00:00Z'
        }
      ];
      
      setUsers(mockUsers);
    };

    loadUsers();
  }, []);

  const handleEditUser = (user: User) => {
    setSelectedUser({ ...user });
    setEditDialogOpen(true);
  };

  const handleDeleteUser = (user: User) => {
    setSelectedUser(user);
    setDeleteDialogOpen(true);
  };

  const handleSaveUser = async () => {
    if (!selectedUser) return;

    try {
      // 这里应该调用实际的API
      const userIndex = users.findIndex(u => u.id === selectedUser.id);
      if (userIndex !== -1) {
        const updatedUsers = [...users];
        updatedUsers[userIndex] = selectedUser;
        setUsers(updatedUsers);
        
        setMessage({ type: 'success', text: '用户信息已更新' });
      }
      
      setEditDialogOpen(false);
      setSelectedUser(null);
    } catch (error) {
      setMessage({ type: 'error', text: '更新用户信息失败' });
    }
  };

  const handleConfirmDelete = async () => {
    if (!selectedUser) return;

    try {
      // 这里应该调用实际的API
      setUsers(users.filter(u => u.id !== selectedUser.id));
      setMessage({ type: 'success', text: '用户已删除' });
      
      setDeleteDialogOpen(false);
      setSelectedUser(null);
    } catch (error) {
      setMessage({ type: 'error', text: '删除用户失败' });
    }
  };

  const handlePermissionChange = (permission: PermissionLevel, checked: boolean) => {
    if (!selectedUser) return;

    const newPermissions = checked
      ? [...selectedUser.permissions, permission]
      : selectedUser.permissions.filter(p => p !== permission);

    setSelectedUser({
      ...selectedUser,
      permissions: newPermissions
    });
  };



  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  return (
    <div className="p-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center">
          <Users className="mr-3 h-8 w-8 text-blue-600" />
          <h1 className="text-3xl font-bold text-gray-900">用户管理</h1>
        </div>
        <button
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          onClick={() => {
            setSelectedUser({
              id: '',
              username: '',
              email: '',
              permissions: [PermissionLevel.USER],
              deviceType: DeviceType.DESKTOP,
              isActive: true,
              lastLogin: '',
              createdAt: new Date().toISOString()
            });
            setEditDialogOpen(true);
          }}
        >
          <Plus className="mr-2 h-4 w-4" />
          添加用户
        </button>
      </div>

      {/* 消息提示 */}
      {message && (
        <div className={`mb-4 p-4 rounded-lg flex items-center justify-between ${
          message.type === 'success' ? 'bg-green-50 text-green-800 border border-green-200' :
          message.type === 'error' ? 'bg-red-50 text-red-800 border border-red-200' :
          'bg-blue-50 text-blue-800 border border-blue-200'
        }`}>
          <span>{message.text}</span>
          <button 
            onClick={() => setMessage(null)}
            className="ml-4 text-gray-400 hover:text-gray-600"
          >
            ×
          </button>
        </div>
      )}

      {/* 用户表格 */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">用户</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">权限</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">设备类型</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">最后登录</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {users.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="h-10 w-10 rounded-full bg-blue-500 flex items-center justify-center text-white font-medium">
                        {user.username.charAt(0).toUpperCase()}
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900">{user.username}</div>
                        <div className="text-sm text-gray-500">{user.email}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-wrap gap-1">
                      {user.permissions.map((permission) => (
                        <span
                          key={permission}
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            permission === PermissionLevel.SUPER_ADMIN ? 'bg-red-100 text-red-800' :
                            permission === PermissionLevel.ADMIN ? 'bg-yellow-100 text-yellow-800' :
                            permission === PermissionLevel.MODERATOR ? 'bg-blue-100 text-blue-800' :
                            'bg-green-100 text-green-800'
                          }`}
                        >
                          {permission}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                      {user.deviceType}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      user.isActive ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                    }`}>
                      {user.isActive ? '活跃' : '非活跃'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {user.lastLogin ? formatDate(user.lastLogin) : '从未登录'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <button
                      onClick={() => handleEditUser(user)}
                      className="text-blue-600 hover:text-blue-900 mr-3"
                      title="编辑用户"
                    >
                      <Edit className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleDeleteUser(user)}
                      className="text-red-600 hover:text-red-900"
                      title="删除用户"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* 编辑用户对话框 */}
      {editDialogOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-bold mb-4">
              {selectedUser?.id ? '编辑用户' : '添加用户'}
            </h2>
            
            {selectedUser && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">用户名</label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    value={selectedUser.username}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSelectedUser({ ...selectedUser, username: e.target.value })}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
                  <input
                    type="email"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    value={selectedUser.email}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSelectedUser({ ...selectedUser, email: e.target.value })}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">设备类型</label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    value={selectedUser.deviceType}
                    onChange={(e: any) => setSelectedUser({ 
                      ...selectedUser, 
                      deviceType: e.target.value as DeviceType 
                    })}
                  >
                    {Object.values(DeviceType).map((type) => (
                      <option key={type} value={type}>
                        {type.toUpperCase()}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <h3 className="text-sm font-medium text-gray-700 mb-2">权限设置</h3>
                  <div className="space-y-2">
                    {Object.values(PermissionLevel).map((permission) => (
                      <label key={permission} className="flex items-center">
                        <input
                          type="checkbox"
                          className="mr-2"
                          checked={selectedUser.permissions.includes(permission)}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => handlePermissionChange(permission, e.target.checked)}
                        />
                        <span className="text-sm">{permission.replace('_', ' ').toUpperCase()}</span>
                      </label>
                    ))}
                  </div>
                </div>

                <label className="flex items-center">
                  <input
                    type="checkbox"
                    className="mr-2"
                    checked={selectedUser.isActive}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSelectedUser({ 
                      ...selectedUser, 
                      isActive: e.target.checked 
                    })}
                  />
                  <span className="text-sm">用户活跃状态</span>
                </label>
              </div>
            )}
            
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => setEditDialogOpen(false)}
                className="px-4 py-2 text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={handleSaveUser}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                保存
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 删除确认对话框 */}
      {deleteDialogOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-sm mx-4">
            <h2 className="text-xl font-bold mb-4">确认删除</h2>
            <p className="text-gray-600 mb-6">
              确定要删除用户 "{selectedUser?.username}" 吗？此操作不可撤销。
            </p>
            <div className="flex justify-end space-x-3">
              <button
                onClick={() => setDeleteDialogOpen(false)}
                className="px-4 py-2 text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={handleConfirmDelete}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default UserManagementPage;