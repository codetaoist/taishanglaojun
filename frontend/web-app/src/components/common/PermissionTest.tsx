import React, { useState } from 'react';
import { Card, Button, Select, Space, Typography, Divider, Tag, Alert } from 'antd';
import { useAuth } from '../../hooks/useAuth';
import { 
  mainMenuConfig, 
  filterMenuByPermissions, 
  filterMenuByStatus
} from '../../config/menuConfig';

const { Title, Text } = Typography;
const { Option } = Select;

interface TestUser {
  id: string;
  username: string;
  email: string;
  role: 'user' | 'admin' | 'moderator';
  roles: string[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

const PermissionTest: React.FC = () => {
  const { user } = useAuth();
  const [testUser, setTestUser] = useState<TestUser | null>(null);

  // 预定义的测试用户
  const testUsers: TestUser[] = [
    {
      id: '1',
      username: 'regular_user',
      email: 'user@example.com',
      role: 'user',
      roles: ['user'],
      permissions: ['read:wisdom', 'create:note', 'read:community'],
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    },
    {
      id: '2',
      username: 'moderator_user',
      email: 'moderator@example.com',
      role: 'moderator',
      roles: ['user', 'moderator'],
      permissions: ['read:wisdom', 'create:note', 'read:community', 'moderate:content', 'manage:categories'],
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    },
    {
      id: '3',
      username: 'admin_user',
      email: 'admin@example.com',
      role: 'admin',
      roles: ['user', 'moderator', 'admin'],
      permissions: ['*'], // 管理员拥有所有权限
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    },
    {
      id: '4',
      username: 'security_user',
      email: 'security@example.com',
      role: 'admin',
      roles: ['user', 'admin', 'security'],
      permissions: ['read:wisdom', 'create:note', 'read:community', 'manage:security', 'scan:security'],
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z'
    }
  ];

  const handleUserChange = (userId: string) => {
    const selectedUser = testUsers.find(u => u.id === userId);
    setTestUser(selectedUser || null);
  };

  const getFilteredMenus = (testUserData: TestUser | null) => {
    if (!testUserData) return [];
    
    const userRoles = testUserData.roles || [];
    const userPermissions = testUserData.permissions || [];
    
    // 根据权限过滤菜单
    let filtered = filterMenuByPermissions(mainMenuConfig, userRoles, userPermissions);
    
    // 根据开发状态过滤
    filtered = filterMenuByStatus(filtered, ['completed', 'partial']);
    
    return filtered;
  };

  const renderMenuItem = (item: MenuItem, level = 0) => {
    const indent = level * 20;
    return (
      <div key={item.key} style={{ marginLeft: indent, marginBottom: 8 }}>
        <Space>
          {item.icon}
          <Text strong={level === 0}>{item.label}</Text>
          <Tag color={item.status === 'completed' ? 'green' : item.status === 'partial' ? 'orange' : 'blue'}>
            {item.status === 'completed' ? '已完成' : item.status === 'partial' ? '部分完成' : '规划中'}
          </Tag>
          {item.requiredRole && (
            <Tag color="purple">需要角色: {item.requiredRole.join(', ')}</Tag>
          )}
          {item.requiredPermission && (
            <Tag color="cyan">需要权限: {item.requiredPermission.join(', ')}</Tag>
          )}
        </Space>
        {item.children && (
          <div style={{ marginTop: 8 }}>
            {item.children.map(child => renderMenuItem(child, level + 1))}
          </div>
        )}
      </div>
    );
  };

  const currentUserMenus = user ? getFilteredMenus({
    id: user.id,
    username: user.username,
    email: user.email,
    role: user.role,
    roles: user.roles || [user.role],
    permissions: user.permissions || [],
    createdAt: user.createdAt,
    updatedAt: user.updatedAt
  }) : [];

  const testUserMenus = testUser ? getFilteredMenus(testUser) : [];

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>权限系统测试</Title>
      
      <Alert
        message="权限测试说明"
        description="此页面用于测试菜单权限控制系统。您可以选择不同的测试用户来查看他们能够访问的菜单项。"
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Card title="当前用户权限" style={{ marginBottom: 24 }}>
        {user ? (
          <div>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>用户名：</Text> {user.username}
              </div>
              <div>
                <Text strong>角色：</Text> {user.role}
                {user.roles && user.roles.length > 0 && (
                  <span> (多角色: {user.roles.join(', ')})</span>
                )}
              </div>
              <div>
                <Text strong>权限：</Text> 
                {user.permissions && user.permissions.length > 0 ? (
                  <div style={{ marginTop: 8 }}>
                    {user.permissions.map(permission => (
                      <Tag key={permission} color="blue">{permission}</Tag>
                    ))}
                  </div>
                ) : (
                  <Text type="secondary">无特殊权限</Text>
                )}
              </div>
            </Space>
            
            <Divider />
            
            <Title level={4}>可访问的菜单项：</Title>
            <div style={{ maxHeight: 400, overflowY: 'auto' }}>
              {currentUserMenus.map(item => renderMenuItem(item))}
            </div>
          </div>
        ) : (
          <Text type="secondary">未登录</Text>
        )}
      </Card>

      <Card title="测试用户权限">
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <Text strong>选择测试用户：</Text>
            <Select
              style={{ width: 300, marginLeft: 16 }}
              placeholder="选择一个测试用户"
              onChange={handleUserChange}
              allowClear
            >
              {testUsers.map(user => (
                <Option key={user.id} value={user.id}>
                  {user.username} ({user.role})
                </Option>
              ))}
            </Select>
          </div>

          {testUser && (
            <div>
              <Divider />
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Text strong>测试用户：</Text> {testUser.username}
                </div>
                <div>
                  <Text strong>角色：</Text> {testUser.role}
                  {testUser.roles && testUser.roles.length > 0 && (
                    <span> (多角色: {testUser.roles.join(', ')})</span>
                  )}
                </div>
                <div>
                  <Text strong>权限：</Text>
                  <div style={{ marginTop: 8 }}>
                    {testUser.permissions.map(permission => (
                      <Tag key={permission} color="green">{permission}</Tag>
                    ))}
                  </div>
                </div>
              </Space>
              
              <Divider />
              
              <Title level={4}>该用户可访问的菜单项：</Title>
              <div style={{ maxHeight: 400, overflowY: 'auto' }}>
                {testUserMenus.map(item => renderMenuItem(item))}
              </div>
            </div>
          )}
        </Space>
      </Card>
    </div>
  );
};

export default PermissionTest;