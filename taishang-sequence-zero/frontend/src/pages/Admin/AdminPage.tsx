import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Table, Button, Form, Input, Select, Modal, message, Tabs, Statistic, Tag, Avatar, Space, Popconfirm, Drawer, DatePicker, Switch, Alert, Progress, List, Typography } from 'antd';
import {
  UserOutlined,
  SettingOutlined,
  DashboardOutlined,
  TeamOutlined,
  FileTextOutlined,
  BarChartOutlined,
  SecurityScanOutlined,
  BellOutlined,
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  ExportOutlined,
  EyeOutlined,
  LockOutlined,
  UnlockOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { clearError } from '../../store/slices/adminSlice';
import { setLanguage } from '../../store/slices/uiSlice';

const { Option } = Select;
const { TabPane } = Tabs;
const { TextArea } = Input;
const { RangePicker } = DatePicker;

// 模拟用户数据
const mockUsers = [
  {
    id: '1',
    username: 'admin',
    email: 'admin@example.com',
    role: 'admin',
    status: 'active',
    createdAt: '2024-01-01',
    lastLogin: '2024-01-15'
  },
  {
    id: '2',
    username: 'user1',
    email: 'user1@example.com',
    role: 'user',
    status: 'active',
    createdAt: '2024-01-02',
    lastLogin: '2024-01-14'
  }
];
const { Title, Text } = Typography;

interface User {
  id: string;
  username: string;
  email: string;
  fullName: string;
  role: 'user' | 'admin' | 'super_admin';
  status: 'active' | 'inactive' | 'banned';
  createdAt: string;
  lastLoginAt: string;
  loginCount: number;
  avatar?: string;
}

interface SystemStats {
  totalUsers: number;
  activeUsers: number;
  newUsersToday: number;
  totalSessions: number;
  activeSessions: number;
  systemLoad: number;
  memoryUsage: number;
  diskUsage: number;
}

const AdminPage: React.FC = () => {
  const [userForm] = Form.useForm();
  const [settingsForm] = Form.useForm();
  
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);
  const { loading } = useAppSelector(state => state.admin);

  const [activeTab, setActiveTab] = useState('dashboard');
  const [userModalVisible, setUserModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [users, setUsers] = useState<any[]>([]);
  const [userDrawerVisible, setUserDrawerVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [searchText, setSearchText] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');

  // 模拟系统统计数据
  const [systemStats] = useState<SystemStats>({
    totalUsers: 1248,
    activeUsers: 892,
    newUsersToday: 23,
    totalSessions: 5647,
    activeSessions: 156,
    systemLoad: 45,
    memoryUsage: 68,
    diskUsage: 34,
  });

  // 模拟用户数据
  const [mockUsers] = useState<User[]>([
    {
      id: '1',
      username: 'admin',
      email: 'admin@example.com',
      fullName: '系统管理员',
      role: 'super_admin',
      status: 'active',
      createdAt: '2024-01-01',
      lastLoginAt: '2024-01-25 10:30:00',
      loginCount: 245,
    },
    {
      id: '2',
      username: 'user001',
      email: 'user001@example.com',
      fullName: '张三',
      role: 'user',
      status: 'active',
      createdAt: '2024-01-15',
      lastLoginAt: '2024-01-25 09:15:00',
      loginCount: 67,
    },
    {
      id: '3',
      username: 'moderator',
      email: 'mod@example.com',
      fullName: '李四',
      role: 'admin',
      status: 'active',
      createdAt: '2024-01-10',
      lastLoginAt: '2024-01-24 16:45:00',
      loginCount: 123,
    },
  ]);

  useEffect(() => {
    // TODO: 加载用户数据
    console.log('Load users data');
    setUsers(mockUsers);
  }, [dispatch]);

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 获取角色标签
  const getRoleTag = (role: string) => {
    const roleMap = {
      super_admin: { color: 'red', text: getText('超级管理员', 'Super Admin') },
      admin: { color: 'orange', text: getText('管理员', 'Admin') },
      user: { color: 'blue', text: getText('普通用户', 'User') },
    };
    const roleInfo = roleMap[role as keyof typeof roleMap] || roleMap.user;
    return <Tag color={roleInfo.color}>{roleInfo.text}</Tag>;
  };

  // 获取状态标签
  const getStatusTag = (status: string) => {
    const statusMap = {
      active: { color: 'green', text: getText('正常', 'Active'), icon: <CheckCircleOutlined /> },
      inactive: { color: 'default', text: getText('未激活', 'Inactive'), icon: <WarningOutlined /> },
      banned: { color: 'red', text: getText('已封禁', 'Banned'), icon: <CloseCircleOutlined /> },
    };
    const statusInfo = statusMap[status as keyof typeof statusMap] || statusMap.active;
    return (
      <Tag color={statusInfo.color} icon={statusInfo.icon}>
        {statusInfo.text}
      </Tag>
    );
  };

  // 处理用户操作
  const handleCreateUser = () => {
    setEditingUser(null);
    userForm.resetFields();
    setUserModalVisible(true);
  };

  const handleEditUser = (user: User) => {
    setEditingUser(user);
    userForm.setFieldsValue(user);
    setUserModalVisible(true);
  };

  const handleViewUser = (user: User) => {
    setSelectedUser(user);
    setUserDrawerVisible(true);
  };

  const handleSaveUser = async (values: any) => {
    try {
      if (editingUser) {
        // TODO: 实现更新用户功能
        console.log('Update user:', { id: editingUser.id, ...values });
        message.success(getText('用户更新成功', 'User updated successfully'));
      } else {
        // TODO: 实现创建用户功能
        console.log('Create user:', values);
        message.success(getText('用户创建成功', 'User created successfully'));
      }
      setUserModalVisible(false);
      userForm.resetFields();
    } catch (error) {
      message.error(getText('操作失败', 'Operation failed'));
    }
  };

  const handleDeleteUser = async (userId: string) => {
    try {
      // TODO: 实现删除用户功能
      console.log('Delete user:', userId);
      message.success(getText('用户删除成功', 'User deleted successfully'));
    } catch (error) {
      message.error(getText('删除失败', 'Delete failed'));
    }
  };

  const handleToggleUserStatus = async (user: User) => {
    const newStatus = user.status === 'active' ? 'banned' : 'active';
    try {
      // TODO: 实现用户状态切换功能
      console.log('Toggle user status:', { id: user.id, status: newStatus });
      message.success(getText('状态更新成功', 'Status updated successfully'));
    } catch (error) {
      message.error(getText('状态更新失败', 'Status update failed'));
    }
  };

  // 过滤用户数据
  const filteredUsers = users.filter(user => {
    const matchesSearch = !searchText || 
      user.username.toLowerCase().includes(searchText.toLowerCase()) ||
      user.email.toLowerCase().includes(searchText.toLowerCase()) ||
      user.fullName.toLowerCase().includes(searchText.toLowerCase());
    const matchesRole = !roleFilter || user.role === roleFilter;
    const matchesStatus = !statusFilter || user.status === statusFilter;
    return matchesSearch && matchesRole && matchesStatus;
  });

  // 用户表格列定义
  const userColumns = [
    {
      title: getText('用户', 'User'),
      key: 'user',
      render: (record: User) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Avatar src={record.avatar} icon={<UserOutlined />} />
          <div>
            <div style={{ fontWeight: 500 }}>{record.fullName || record.username}</div>
            <div style={{ fontSize: 12, color: '#666' }}>{record.email}</div>
          </div>
        </div>
      ),
    },
    {
      title: getText('角色', 'Role'),
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => getRoleTag(role),
    },
    {
      title: getText('状态', 'Status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: getText('注册时间', 'Created'),
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
    {
      title: getText('最后登录', 'Last Login'),
      dataIndex: 'lastLoginAt',
      key: 'lastLoginAt',
    },
    {
      title: getText('登录次数', 'Login Count'),
      dataIndex: 'loginCount',
      key: 'loginCount',
    },
    {
      title: getText('操作', 'Actions'),
      key: 'actions',
      render: (record: User) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewUser(record)}
            title={getText('查看详情', 'View Details')}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEditUser(record)}
            title={getText('编辑用户', 'Edit User')}
          />
          <Button
            type="text"
            icon={record.status === 'active' ? <LockOutlined /> : <UnlockOutlined />}
            onClick={() => handleToggleUserStatus(record)}
            title={record.status === 'active' ? getText('封禁用户', 'Ban User') : getText('解封用户', 'Unban User')}
          />
          <Popconfirm
            title={getText('确定要删除这个用户吗？', 'Are you sure to delete this user?')}
            onConfirm={() => handleDeleteUser(record.id)}
            okText={getText('确定', 'Yes')}
            cancelText={getText('取消', 'No')}
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title={getText('删除用户', 'Delete User')}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 模拟系统日志数据
  const systemLogs = [
    {
      id: 1,
      type: 'info',
      message: getText('用户 user001 登录系统', 'User user001 logged in'),
      timestamp: '2024-01-25 10:30:15',
    },
    {
      id: 2,
      type: 'warning',
      message: getText('系统内存使用率达到 80%', 'System memory usage reached 80%'),
      timestamp: '2024-01-25 10:25:32',
    },
    {
      id: 3,
      type: 'error',
      message: getText('数据库连接失败', 'Database connection failed'),
      timestamp: '2024-01-25 10:20:45',
    },
  ];

  return (
    <div className="admin-page">
      {/* 页面标题 */}
      <div className="page-header">
        <h1 className="page-title">
          <SettingOutlined style={{ marginRight: 12, color: '#1890ff' }} />
          {getText('系统管理', 'System Administration')}
        </h1>
        <p className="page-description">
          {getText(
            '管理用户、监控系统状态、配置系统设置',
            'Manage users, monitor system status, and configure system settings'
          )}
        </p>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab} size="large">
        {/* 仪表板 */}
        <TabPane 
          tab={
            <span>
              <DashboardOutlined />
              {getText('仪表板', 'Dashboard')}
            </span>
          } 
          key="dashboard"
        >
          <Row gutter={[24, 24]}>
            {/* 系统统计 */}
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title={getText('总用户数', 'Total Users')}
                  value={systemStats.totalUsers}
                  prefix={<UserOutlined />}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title={getText('活跃用户', 'Active Users')}
                  value={systemStats.activeUsers}
                  prefix={<TeamOutlined />}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title={getText('今日新增', 'New Today')}
                  value={systemStats.newUsersToday}
                  prefix={<PlusOutlined />}
                  valueStyle={{ color: '#faad14' }}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title={getText('活跃会话', 'Active Sessions')}
                  value={systemStats.activeSessions}
                  prefix={<BellOutlined />}
                  valueStyle={{ color: '#722ed1' }}
                />
              </Card>
            </Col>

            {/* 系统性能 */}
            <Col xs={24} lg={12}>
              <Card title={getText('系统性能', 'System Performance')}>
                <div className="performance-metrics">
                  <div className="metric-item">
                    <div className="metric-label">{getText('系统负载', 'System Load')}</div>
                    <Progress percent={systemStats.systemLoad} status={systemStats.systemLoad > 80 ? 'exception' : 'normal'} />
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">{getText('内存使用', 'Memory Usage')}</div>
                    <Progress percent={systemStats.memoryUsage} status={systemStats.memoryUsage > 80 ? 'exception' : 'normal'} />
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">{getText('磁盘使用', 'Disk Usage')}</div>
                    <Progress percent={systemStats.diskUsage} />
                  </div>
                </div>
              </Card>
            </Col>

            {/* 系统日志 */}
            <Col xs={24} lg={12}>
              <Card 
                title={getText('系统日志', 'System Logs')}
                extra={
                  <Button type="text" icon={<ReloadOutlined />} size="small">
                    {getText('刷新', 'Refresh')}
                  </Button>
                }
              >
                <List
                  size="small"
                  dataSource={systemLogs}
                  renderItem={log => (
                    <List.Item>
                      <div className="log-item">
                        <Tag 
                          color={
                            log.type === 'error' ? 'red' : 
                            log.type === 'warning' ? 'orange' : 'blue'
                          }
                        >
                          {log.type.toUpperCase()}
                        </Tag>
                        <span className="log-message">{log.message}</span>
                        <span className="log-time">{log.timestamp}</span>
                      </div>
                    </List.Item>
                  )}
                />
              </Card>
            </Col>
          </Row>
        </TabPane>

        {/* 用户管理 */}
        <TabPane 
          tab={
            <span>
              <TeamOutlined />
              {getText('用户管理', 'User Management')}
            </span>
          } 
          key="users"
        >
          <Card>
            {/* 搜索和过滤 */}
            <div className="table-toolbar">
              <div className="toolbar-left">
                <Input.Search
                  placeholder={getText('搜索用户...', 'Search users...')}
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                  style={{ width: 300 }}
                  allowClear
                />
                <Select
                  placeholder={getText('角色筛选', 'Filter by role')}
                  value={roleFilter}
                  onChange={setRoleFilter}
                  style={{ width: 150 }}
                  allowClear
                >
                  <Option value="user">{getText('普通用户', 'User')}</Option>
                  <Option value="admin">{getText('管理员', 'Admin')}</Option>
                  <Option value="super_admin">{getText('超级管理员', 'Super Admin')}</Option>
                </Select>
                <Select
                  placeholder={getText('状态筛选', 'Filter by status')}
                  value={statusFilter}
                  onChange={setStatusFilter}
                  style={{ width: 150 }}
                  allowClear
                >
                  <Option value="active">{getText('正常', 'Active')}</Option>
                  <Option value="inactive">{getText('未激活', 'Inactive')}</Option>
                  <Option value="banned">{getText('已封禁', 'Banned')}</Option>
                </Select>
              </div>
              <div className="toolbar-right">
                <Button icon={<ExportOutlined />}>
                  {getText('导出', 'Export')}
                </Button>
                <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateUser}>
                  {getText('新建用户', 'Create User')}
                </Button>
              </div>
            </div>

            {/* 用户表格 */}
            <Table
              columns={userColumns}
              dataSource={filteredUsers}
              rowKey="id"
              loading={loading}
              pagination={{
                total: filteredUsers.length,
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total, range) => 
                  getText(
                    `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
                    `${range[0]}-${range[1]} of ${total} items`
                  ),
              }}
            />
          </Card>
        </TabPane>

        {/* 系统设置 */}
        <TabPane 
          tab={
            <span>
              <SettingOutlined />
              {getText('系统设置', 'System Settings')}
            </span>
          } 
          key="settings"
        >
          <Row gutter={[24, 24]}>
            <Col xs={24} lg={12}>
              <Card title={getText('基本设置', 'Basic Settings')}>
                <Form
                  form={settingsForm}
                  layout="vertical"
                  initialValues={{
                    siteName: getText('太上序列零', 'Taishang Sequence Zero'),
                    siteDescription: getText('意识融合与文化智慧平台', 'Consciousness Fusion and Cultural Wisdom Platform'),
                    defaultLanguage: 'zh',
                    registrationEnabled: true,
                    emailVerificationRequired: true,
                  }}
                >
                  <Form.Item
                    name="siteName"
                    label={getText('站点名称', 'Site Name')}
                  >
                    <Input />
                  </Form.Item>
                  
                  <Form.Item
                    name="siteDescription"
                    label={getText('站点描述', 'Site Description')}
                  >
                    <TextArea rows={3} />
                  </Form.Item>
                  
                  <Form.Item
                    name="defaultLanguage"
                    label={getText('默认语言', 'Default Language')}
                  >
                    <Select>
                      <Option value="zh">中文</Option>
                      <Option value="en">English</Option>
                    </Select>
                  </Form.Item>
                  
                  <Form.Item
                    name="registrationEnabled"
                    label={getText('允许用户注册', 'Allow User Registration')}
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                  
                  <Form.Item
                    name="emailVerificationRequired"
                    label={getText('需要邮箱验证', 'Require Email Verification')}
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button type="primary">
                      {getText('保存设置', 'Save Settings')}
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
            
            <Col xs={24} lg={12}>
              <Card title={getText('安全设置', 'Security Settings')}>
                <Alert
                  message={getText('安全提醒', 'Security Notice')}
                  description={getText(
                    '请定期检查系统安全设置，确保系统安全运行。',
                    'Please regularly check system security settings to ensure secure operation.'
                  )}
                  type="info"
                  showIcon
                  style={{ marginBottom: 20 }}
                />
                
                <div className="security-options">
                  <div className="security-item">
                    <div className="security-info">
                      <SecurityScanOutlined className="security-icon" />
                      <div>
                        <div className="security-title">{getText('密码策略', 'Password Policy')}</div>
                        <div className="security-desc">{getText('设置密码复杂度要求', 'Set password complexity requirements')}</div>
                      </div>
                    </div>
                    <Button>{getText('配置', 'Configure')}</Button>
                  </div>
                  
                  <div className="security-item">
                    <div className="security-info">
                      <LockOutlined className="security-icon" />
                      <div>
                        <div className="security-title">{getText('登录限制', 'Login Restrictions')}</div>
                        <div className="security-desc">{getText('设置登录失败次数限制', 'Set login failure attempt limits')}</div>
                      </div>
                    </div>
                    <Button>{getText('配置', 'Configure')}</Button>
                  </div>
                  
                  <div className="security-item">
                    <div className="security-info">
                      <BellOutlined className="security-icon" />
                      <div>
                        <div className="security-title">{getText('安全通知', 'Security Notifications')}</div>
                        <div className="security-desc">{getText('配置安全事件通知', 'Configure security event notifications')}</div>
                      </div>
                    </div>
                    <Button>{getText('配置', 'Configure')}</Button>
                  </div>
                </div>
              </Card>
            </Col>
          </Row>
        </TabPane>
      </Tabs>

      {/* 用户编辑/创建模态框 */}
      <Modal
        title={editingUser ? getText('编辑用户', 'Edit User') : getText('创建用户', 'Create User')}
        open={userModalVisible}
        onCancel={() => setUserModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={userForm}
          layout="vertical"
          onFinish={handleSaveUser}
        >
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="username"
                label={getText('用户名', 'Username')}
                rules={[
                  { required: true, message: getText('请输入用户名', 'Please enter username') },
                  { min: 3, message: getText('用户名至少3位字符', 'Username must be at least 3 characters') },
                ]}
              >
                <Input disabled={!!editingUser} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="email"
                label={getText('邮箱地址', 'Email Address')}
                rules={[
                  { required: true, message: getText('请输入邮箱地址', 'Please enter email address') },
                  { type: 'email', message: getText('请输入有效的邮箱地址', 'Please enter a valid email address') },
                ]}
              >
                <Input />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="fullName"
                label={getText('真实姓名', 'Full Name')}
              >
                <Input />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="role"
                label={getText('用户角色', 'User Role')}
                rules={[
                  { required: true, message: getText('请选择用户角色', 'Please select user role') },
                ]}
              >
                <Select>
                  <Option value="user">{getText('普通用户', 'User')}</Option>
                  <Option value="admin">{getText('管理员', 'Admin')}</Option>
                  {user?.role === 'super_admin' && (
                    <Option value="super_admin">{getText('超级管理员', 'Super Admin')}</Option>
                  )}
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="status"
            label={getText('用户状态', 'User Status')}
            rules={[
              { required: true, message: getText('请选择用户状态', 'Please select user status') },
            ]}
          >
            <Select>
              <Option value="active">{getText('正常', 'Active')}</Option>
              <Option value="inactive">{getText('未激活', 'Inactive')}</Option>
              <Option value="banned">{getText('已封禁', 'Banned')}</Option>
            </Select>
          </Form.Item>
          
          {!editingUser && (
            <Form.Item
              name="password"
              label={getText('初始密码', 'Initial Password')}
              rules={[
                { required: true, message: getText('请输入初始密码', 'Please enter initial password') },
                { min: 6, message: getText('密码至少6位字符', 'Password must be at least 6 characters') },
              ]}
            >
              <Input.Password />
            </Form.Item>
          )}
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingUser ? getText('更新', 'Update') : getText('创建', 'Create')}
              </Button>
              <Button onClick={() => setUserModalVisible(false)}>
                {getText('取消', 'Cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 用户详情抽屉 */}
      <Drawer
        title={getText('用户详情', 'User Details')}
        placement="right"
        onClose={() => setUserDrawerVisible(false)}
        open={userDrawerVisible}
        width={400}
      >
        {selectedUser && (
          <div className="user-details">
            <div className="user-avatar-section">
              <Avatar size={80} src={selectedUser.avatar} icon={<UserOutlined />} />
              <div className="user-basic-info">
                <h3>{selectedUser.fullName || selectedUser.username}</h3>
                <p>{selectedUser.email}</p>
                {getRoleTag(selectedUser.role)}
                {getStatusTag(selectedUser.status)}
              </div>
            </div>
            
            <div className="user-stats">
              <div className="stat-row">
                <span className="stat-label">{getText('注册时间', 'Registration Date')}:</span>
                <span className="stat-value">{selectedUser.createdAt}</span>
              </div>
              <div className="stat-row">
                <span className="stat-label">{getText('最后登录', 'Last Login')}:</span>
                <span className="stat-value">{selectedUser.lastLoginAt}</span>
              </div>
              <div className="stat-row">
                <span className="stat-label">{getText('登录次数', 'Login Count')}:</span>
                <span className="stat-value">{selectedUser.loginCount}</span>
              </div>
            </div>
            
            <div className="user-actions">
              <Button 
                type="primary" 
                block 
                style={{ marginBottom: 8 }}
                onClick={() => {
                  setUserDrawerVisible(false);
                  handleEditUser(selectedUser);
                }}
              >
                {getText('编辑用户', 'Edit User')}
              </Button>
              <Button 
                block 
                onClick={() => handleToggleUserStatus(selectedUser)}
              >
                {selectedUser.status === 'active' ? getText('封禁用户', 'Ban User') : getText('解封用户', 'Unban User')}
              </Button>
            </div>
          </div>
        )}
      </Drawer>

      <style>{`
        .admin-page {
          padding: 0;
        }

        .page-header {
          margin-bottom: 24px;
        }

        .page-title {
          font-size: 28px;
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 8px;
          display: flex;
          align-items: center;
        }

        .page-description {
          font-size: 16px;
          color: var(--text-secondary);
          margin: 0;
          line-height: 1.6;
        }

        .performance-metrics {
          display: flex;
          flex-direction: column;
          gap: 16px;
        }

        .metric-item {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .metric-label {
          min-width: 80px;
          font-size: 14px;
          color: var(--text-secondary);
        }

        .log-item {
          display: flex;
          align-items: center;
          gap: 8px;
          width: 100%;
        }

        .log-message {
          flex: 1;
          font-size: 14px;
        }

        .log-time {
          font-size: 12px;
          color: var(--text-tertiary);
        }

        .table-toolbar {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          gap: 16px;
        }

        .toolbar-left {
          display: flex;
          gap: 12px;
          flex: 1;
        }

        .toolbar-right {
          display: flex;
          gap: 8px;
        }

        .security-options {
          display: flex;
          flex-direction: column;
          gap: 16px;
        }

        .security-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 16px;
          border: 1px solid var(--border-light);
          border-radius: 8px;
          background: var(--bg-secondary);
        }

        .security-info {
          display: flex;
          align-items: center;
          gap: 12px;
          flex: 1;
        }

        .security-icon {
          font-size: 20px;
          color: var(--primary-color);
        }

        .security-title {
          font-weight: 500;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .security-desc {
          font-size: 12px;
          color: var(--text-tertiary);
          margin: 0;
        }

        .user-details {
          display: flex;
          flex-direction: column;
          gap: 24px;
        }

        .user-avatar-section {
          display: flex;
          flex-direction: column;
          align-items: center;
          text-align: center;
          gap: 16px;
        }

        .user-basic-info h3 {
          margin-bottom: 8px;
          color: var(--text-primary);
        }

        .user-basic-info p {
          margin-bottom: 12px;
          color: var(--text-secondary);
        }

        .user-stats {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .stat-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 8px 0;
          border-bottom: 1px solid var(--border-light);
        }

        .stat-row:last-child {
          border-bottom: none;
        }

        .stat-label {
          font-size: 14px;
          color: var(--text-secondary);
        }

        .stat-value {
          font-size: 14px;
          color: var(--text-primary);
          font-weight: 500;
        }

        .user-actions {
          margin-top: 16px;
        }

        /* 响应式设计 */
        @media (max-width: 768px) {
          .page-title {
            font-size: 24px;
          }

          .page-description {
            font-size: 14px;
          }

          .table-toolbar {
            flex-direction: column;
            align-items: stretch;
          }

          .toolbar-left {
            flex-direction: column;
          }

          .toolbar-right {
            justify-content: center;
          }

          .security-item {
            flex-direction: column;
            align-items: flex-start;
            gap: 12px;
          }

          .performance-metrics {
            gap: 12px;
          }

          .metric-item {
            flex-direction: column;
            align-items: flex-start;
            gap: 8px;
          }

          .metric-label {
            min-width: auto;
          }
        }
      `}</style>
    </div>
  );
};

export default AdminPage;