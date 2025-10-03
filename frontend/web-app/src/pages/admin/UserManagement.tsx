import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Modal,
  Form,
  message,
  Popconfirm,
  Typography,
  Tag,
  Row,
  Col,
  Statistic,
  Avatar,
  Select,
  DatePicker,
  Tooltip,
  Badge,
  Dropdown,
} from 'antd';
import type { MenuProps } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  TeamOutlined,
  CrownOutlined,
  EyeOutlined,
  MoreOutlined,
  LockOutlined,
  UnlockOutlined,
} from '@ant-design/icons';
import type { ColumnType } from 'antd/es/table';
import { apiClient } from '../../services/api';
import dayjs from 'dayjs';

const { Search } = Input;
const { Title } = Typography;
const { RangePicker } = DatePicker;

interface UserData {
  id: string;
  username: string;
  email: string;
  name?: string;
  avatar?: string;
  role: 'user' | 'admin' | 'moderator';
  isAdmin: boolean;
  status: 'active' | 'inactive' | 'banned';
  lastLogin?: string;
  createdAt: string;
  updatedAt: string;
  loginCount: number;
  wisdomCount: number;
  commentCount: number;
}

interface UserFormData {
  username: string;
  email: string;
  name?: string;
  role: 'user' | 'admin' | 'moderator';
  status: 'active' | 'inactive' | 'banned';
  password?: string;
}

interface UserStats {
  totalUsers: number;
  activeUsers: number;
  adminUsers: number;
  newUsersToday: number;
  onlineUsers: number;
}

const UserManagement: React.FC = () => {
  const [users, setUsers] = useState<UserData[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<UserData | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [searchText, setSearchText] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [stats, setStats] = useState<UserStats>({
    totalUsers: 0,
    activeUsers: 0,
    adminUsers: 0,
    newUsersToday: 0,
    onlineUsers: 0
  });
  const [form] = Form.useForm();

  // 获取用户列表
  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const params = {
        search: searchText,
        role: roleFilter,
        status: statusFilter,
        startDate: dateRange?.[0]?.format('YYYY-MM-DD'),
        endDate: dateRange?.[1]?.format('YYYY-MM-DD'),
      };
      const response = await apiClient.getUsers(params);
      setUsers(response.data);
    } catch {
      message.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  }, [searchText, roleFilter, statusFilter, dateRange]);

  // 获取用户统计
  const fetchStats = async () => {
    try {
      const response = await apiClient.getUserStats();
      setStats(response.data);
    } catch {
      message.error('获取用户统计失败');
    }
  };

  useEffect(() => {
    fetchUsers();
    fetchStats();
  }, [searchText, roleFilter, statusFilter, dateRange, fetchUsers]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
  };

  // 处理筛选重置
  const handleReset = () => {
    setSearchText('');
    setRoleFilter('');
    setStatusFilter('');
    setDateRange(null);
  };

  // 处理用户创建/编辑
  const handleSubmit = async (values: UserFormData) => {
    try {
      if (editingUser) {
        await apiClient.updateUser(editingUser.id, values);
        message.success('用户更新成功');
      } else {
        await apiClient.createUser(values);
        message.success('用户创建成功');
      }
      setModalVisible(false);
      setEditingUser(null);
      form.resetFields();
      fetchUsers();
      fetchStats();
    } catch {
      message.error(editingUser ? '用户更新失败' : '用户创建失败');
    }
  };

  // 处理用户删除
  const handleDelete = async (userId: string) => {
    try {
      await apiClient.deleteUser(userId);
      message.success('用户删除成功');
      fetchUsers();
      fetchStats();
    } catch {
      message.error('用户删除失败');
    }
  };

  // 处理批量删除
  const handleBatchDelete = async () => {
    try {
      await apiClient.batchDeleteUsers(selectedRowKeys as string[]);
      message.success('批量删除成功');
      setSelectedRowKeys([]);
      fetchUsers();
      fetchStats();
    } catch {
      message.error('批量删除失败');
    }
  };

  // 处理用户状态切换
  const handleStatusToggle = async (userId: string, status: string) => {
    try {
      await apiClient.updateUserStatus(userId, status);
      message.success('用户状态更新成功');
      fetchUsers();
    } catch {
      message.error('用户状态更新失败');
    }
  };

  // 处理角色变更
  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await apiClient.updateUserRole(userId, role);
      message.success('用户角色更新成功');
      fetchUsers();
      fetchStats();
    } catch {
      message.error('用户角色更新失败');
    }
  };

  // 获取角色标签颜色
  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin': return 'red';
      case 'moderator': return 'orange';
      default: return 'blue';
    }
  };

  // 获取状态标签颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'orange';
      case 'banned': return 'red';
      default: return 'default';
    }
  };

  // 用户操作菜单
  const getUserActionMenu = (user: UserData): MenuProps => ({
    items: [
      {
        key: 'edit',
        icon: <EditOutlined />,
        label: '编辑用户',
        onClick: () => {
          setEditingUser(user);
          form.setFieldsValue(user);
          setModalVisible(true);
        }
      },
      {
        key: 'status',
        icon: user.status === 'active' ? <LockOutlined /> : <UnlockOutlined />,
        label: user.status === 'active' ? '禁用用户' : '启用用户',
        onClick: () => handleStatusToggle(user.id, user.status === 'active' ? 'inactive' : 'active')
      },
      {
        key: 'role',
        icon: <CrownOutlined />,
        label: '角色管理',
        children: [
          {
            key: 'user',
            label: '普通用户',
            onClick: () => handleRoleChange(user.id, 'user')
          },
          {
            key: 'moderator',
            label: '版主',
            onClick: () => handleRoleChange(user.id, 'moderator')
          },
          {
            key: 'admin',
            label: '管理员',
            onClick: () => handleRoleChange(user.id, 'admin')
          }
        ]
      },
      {
        type: 'divider'
      },
      {
        key: 'delete',
        icon: <DeleteOutlined />,
        label: '删除用户',
        danger: true,
        onClick: () => {
          Modal.confirm({
            title: '确认删除',
            content: `确定要删除用户 "${user.username}" 吗？`,
            onOk: () => handleDelete(user.id)
          });
        }
      }
    ]
  });

  // 表格列定义
  const columns: ColumnType<UserData>[] = [
    {
      title: '用户信息',
      key: 'userInfo',
      width: 200,
      render: (_, record) => (
        <Space>
          <Avatar 
            src={record.avatar} 
            icon={<UserOutlined />}
            size="large"
          />
          <div>
            <div style={{ fontWeight: 'bold' }}>{record.name || record.username}</div>
            <div style={{ color: '#666', fontSize: '12px' }}>{record.email}</div>
          </div>
        </Space>
      )
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      width: 120,
      sorter: (a, b) => a.username.localeCompare(b.username)
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 100,
      render: (role: string) => (
        <Tag color={getRoleColor(role)} icon={role === 'admin' ? <CrownOutlined /> : <UserOutlined />}>
          {role === 'admin' ? '管理员' : role === 'moderator' ? '版主' : '用户'}
        </Tag>
      ),
      filters: [
        { text: '管理员', value: 'admin' },
        { text: '版主', value: 'moderator' },
        { text: '普通用户', value: 'user' }
      ]
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status === 'active' ? '正常' : status === 'inactive' ? '禁用' : '封禁'}
        </Tag>
      ),
      filters: [
        { text: '正常', value: 'active' },
        { text: '禁用', value: 'inactive' },
        { text: '封禁', value: 'banned' }
      ]
    },
    {
      title: '统计信息',
      key: 'stats',
      width: 150,
      render: (_, record) => (
        <Space direction="vertical" size="small">
          <span>登录: {record.loginCount}次</span>
          <span>智慧: {record.wisdomCount}篇</span>
          <span>评论: {record.commentCount}条</span>
        </Space>
      )
    },
    {
      title: '最后登录',
      dataIndex: 'lastLogin',
      key: 'lastLogin',
      width: 120,
      render: (date: string) => date ? dayjs(date).format('MM-DD HH:mm') : '从未登录',
      sorter: (a, b) => {
        if (!a.lastLogin && !b.lastLogin) return 0;
        if (!a.lastLogin) return 1;
        if (!b.lastLogin) return -1;
        return dayjs(a.lastLogin).unix() - dayjs(b.lastLogin).unix();
      }
    },
    {
      title: '注册时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 120,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
      sorter: (a, b) => dayjs(a.createdAt).unix() - dayjs(b.createdAt).unix()
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              size="small"
              onClick={() => {
                // 查看用户详情逻辑
                message.info('查看用户详情功能待实现');
              }}
            />
          </Tooltip>
          <Dropdown menu={getUserActionMenu(record)} trigger={['click']}>
            <Button type="text" icon={<MoreOutlined />} size="small" />
          </Dropdown>
        </Space>
      )
    }
  ];

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
    getCheckboxProps: (record: UserData) => ({
      disabled: record.role === 'admin' && record.id === 'current-user-id', // 防止删除当前管理员
    }),
  };

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>用户管理</Title>
      
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={8} md={4}>
          <Card>
            <Statistic
              title="总用户数"
              value={stats.totalUsers}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8} md={4}>
          <Card>
            <Statistic
              title="活跃用户"
              value={stats.activeUsers}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8} md={4}>
          <Card>
            <Statistic
              title="管理员"
              value={stats.adminUsers}
              prefix={<CrownOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8} md={4}>
          <Card>
            <Statistic
              title="今日新增"
              value={stats.newUsersToday}
              prefix={<PlusOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8} md={4}>
          <Card>
            <Statistic
              title="在线用户"
              value={stats.onlineUsers}
              prefix={<Badge status="processing" />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索和筛选 */}
      <Card style={{ marginBottom: '16px' }}>
        <Row gutter={16} align="middle">
          <Col xs={24} sm={8} md={6}>
            <Search
              placeholder="搜索用户名或邮箱"
              allowClear
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
          <Col xs={24} sm={8} md={4}>
            <Select
              placeholder="角色筛选"
              allowClear
              style={{ width: '100%' }}
              value={roleFilter}
              onChange={setRoleFilter}
            >
              <Select.Option value="admin">管理员</Select.Option>
              <Select.Option value="moderator">版主</Select.Option>
              <Select.Option value="user">普通用户</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={8} md={4}>
            <Select
              placeholder="状态筛选"
              allowClear
              style={{ width: '100%' }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              <Select.Option value="active">正常</Select.Option>
              <Select.Option value="inactive">禁用</Select.Option>
              <Select.Option value="banned">封禁</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <RangePicker
              placeholder={['开始日期', '结束日期']}
              style={{ width: '100%' }}
              value={dateRange}
              onChange={setDateRange}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Space>
              <Button onClick={handleReset}>重置</Button>
              <Button 
                type="primary" 
                icon={<PlusOutlined />}
                onClick={() => {
                  setEditingUser(null);
                  form.resetFields();
                  setModalVisible(true);
                }}
              >
                添加用户
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 批量操作 */}
      {selectedRowKeys.length > 0 && (
        <Card style={{ marginBottom: '16px' }}>
          <Space>
            <span>已选择 {selectedRowKeys.length} 项</span>
            <Popconfirm
              title="确定要批量删除选中的用户吗？"
              onConfirm={handleBatchDelete}
            >
              <Button danger icon={<DeleteOutlined />}>
                批量删除
              </Button>
            </Popconfirm>
            <Button onClick={() => setSelectedRowKeys([])}>
              取消选择
            </Button>
          </Space>
        </Card>
      )}

      {/* 用户表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          scroll={{ x: 1200 }}
          pagination={{
            total: users.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/总共 ${total} 条`,
          }}
        />
      </Card>

      {/* 用户编辑/创建模态框 */}
      <Modal
        title={editingUser ? '编辑用户' : '添加用户'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditingUser(null);
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="username"
                label="用户名"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input placeholder="请输入用户名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="email"
                label="邮箱"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' }
                ]}
              >
                <Input placeholder="请输入邮箱" />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="name"
            label="真实姓名"
          >
            <Input placeholder="请输入真实姓名" />
          </Form.Item>

          {!editingUser && (
            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="role"
                label="角色"
                rules={[{ required: true, message: '请选择角色' }]}
              >
                <Select placeholder="请选择角色">
                  <Select.Option value="user">普通用户</Select.Option>
                  <Select.Option value="moderator">版主</Select.Option>
                  <Select.Option value="admin">管理员</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="status"
                label="状态"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select placeholder="请选择状态">
                  <Select.Option value="active">正常</Select.Option>
                  <Select.Option value="inactive">禁用</Select.Option>
                  <Select.Option value="banned">封禁</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingUser ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserManagement;