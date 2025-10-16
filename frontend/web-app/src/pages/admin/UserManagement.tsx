import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
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
  Spin,
  Empty,
  Pagination,
  Checkbox,
  Divider,
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
  ReloadOutlined,
  MoreOutlined,
  LockOutlined,
  UnlockOutlined,
  SafetyOutlined,
  UserAddOutlined,
  LoginOutlined,
  BulbOutlined,
  CommentOutlined,
} from '@ant-design/icons';
import type { ColumnType } from 'antd/es/table';
import { apiClient } from '../../services/api';
import { permissionService } from '../../services/permissionService';
import dayjs from 'dayjs';
import { formatDate, formatDateTime, getRoleColor, getStatusColor } from '../../utils/display';
import { useTranslation } from 'react-i18next';


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

// 角色数据接口
interface RoleData {
  id: string;
  name: string;
  code: string;
  type: string;
  level: number;
  description?: string;
  is_active: boolean;
  is_system: boolean;
}

const UserManagement: React.FC = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState<UserData[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [viewModalVisible, setViewModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<UserData | null>(null);
  const [viewingUser, setViewingUser] = useState<UserData | null>(null);
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
  // 角色数据状态
  const [roles, setRoles] = useState<RoleData[]>([]);
  const [rolesLoading, setRolesLoading] = useState(false);
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });
  // 数据缓存
  const [dataCache, setDataCache] = useState<Map<string, UserData[]>>(new Map());
  const [form] = Form.useForm();

  // 模态框打开后再设置或重置表单，避免未挂载调用
  useEffect(() => {
    if (modalVisible) {
      if (editingUser) {
        form.setFieldsValue({
          username: editingUser.username,
          email: editingUser.email,
          name: editingUser.name,
          role: editingUser.role,
          status: editingUser.status,
        });
      } else {
        form.resetFields();
      }
    }
  }, [modalVisible, editingUser, form]);

  // 防抖搜索
  const debouncedSearch = useMemo(() => {
    let timeoutId: NodeJS.Timeout;
    return (value: string) => {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        setSearchText(value);
        setPagination(prev => ({ ...prev, current: 1 }));
      }, 300);
    };
  }, []);

  // 请求取消控制器
  const abortControllerRef = useRef<AbortController | null>(null);

  // 获取用户列表
  const fetchUsers = useCallback(async (page = 1, pageSize = 10) => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    
    // 创建新的取消控制器
    abortControllerRef.current = new AbortController();
    
    setLoading(true);
    try {
      const params = {
        search: searchText,
        role: roleFilter,
        status: statusFilter,
        startDate: dateRange?.[0]?.format('YYYY-MM-DD'),
        endDate: dateRange?.[1]?.format('YYYY-MM-DD'),
        page,
        pageSize
      };
      
      // 生成缓存键
      const cacheKey = JSON.stringify(params);
      
      // 检查缓存（仅在没有搜索条件时使用缓存）
      if (!searchText && dataCache.has(cacheKey)) {
        const cachedData = dataCache.get(cacheKey)!;
        setUsers(cachedData);
        setLoading(false);
        return;
      }
      
      const response = await apiClient.getUsers(params, {
        signal: abortControllerRef.current.signal
      });
      
      // 检查请求是否被取消
      if (abortControllerRef.current.signal.aborted) {
        return;
      }
      
      const userData = response.data?.data || response.data || [];
      const total = response.data?.total || userData.length;
      
      setUsers(userData);
      setPagination(prev => ({ ...prev, total, current: page, pageSize }));
      
      // 更新缓存（仅在没有搜索条件时缓存）
      if (!searchText) {
        setDataCache(prev => new Map(prev.set(cacheKey, userData)));
      }
    } catch (error: any) {
      // 忽略取消的请求
      if (error.name !== 'AbortError') {
        message.error(t('adminUser.messages.fetchUsersFailure'));
      }
    } finally {
      setLoading(false);
    }
  }, [searchText, roleFilter, statusFilter, dateRange, dataCache, t]);

  // 获取角色列表
  const fetchRoles = useCallback(async () => {
    setRolesLoading(true);
    try {
      const response = await permissionService.getRoles({
        page: 1,
        pageSize: 100, // 获取所有角色用于选择器
        search: '',
        type: '',
        active: 'true' // 只获取激活的角色
      });
      
      // 处理响应数据的多层兼容性
      let rolesList: RoleData[] = [];
      if (response?.data?.data) {
        rolesList = response.data.data;
      } else if (response?.data?.roles) {
        rolesList = response.data.roles;
      } else if (Array.isArray(response?.data)) {
        rolesList = response.data;
      } else if (Array.isArray(response)) {
        rolesList = response;
      }
      
      setRoles(rolesList);
    } catch (error) {
      console.error('获取角色列表失败:', error);
      message.error(t('adminUser.messages.fetchRolesFailure'));
    } finally {
      setRolesLoading(false);
    }
  }, [t]);

  // 获取用户统计
  const fetchStats = async () => {
    try {
      const response = await apiClient.getUserStats();
      setStats(response.data);
    } catch {
      message.error(t('adminUser.messages.fetchStatsFailure'));
    }
  };

  useEffect(() => {
    fetchUsers(pagination.current, pagination.pageSize);
  }, [searchText, roleFilter, statusFilter, dateRange]);

  useEffect(() => {
    fetchStats();
    fetchRoles();
  }, []);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      // 取消正在进行的请求
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      // 清理缓存
      setDataCache(new Map());
    };
  }, []);

  // 处理搜索
  const handleSearch = (value: string) => {
    debouncedSearch(value);
  };

  // 处理分页变化
  const handleTableChange = (paginationConfig: any) => {
    const { current, pageSize } = paginationConfig;
    fetchUsers(current, pageSize);
  };

  // 处理筛选重置
  const handleReset = () => {
    setSearchText('');
    setRoleFilter('');
    setStatusFilter('');
    setDateRange(null);
    setPagination(prev => ({ ...prev, current: 1 }));
    setDataCache(new Map()); // 清空缓存
  };

  // 处理用户创建/编辑
  const handleSubmit = async (values: UserFormData) => {
    try {
      if (editingUser) {
        await apiClient.updateUser(editingUser.id, values);
        message.success(t('adminUser.messages.updateSuccess'));
      } else {
        await apiClient.createUser(values);
        message.success(t('adminUser.messages.createSuccess'));
      }
      setModalVisible(false);
      setEditingUser(null);
      form.resetFields();
      fetchUsers(pagination.current, pagination.pageSize);
      fetchStats();
      setDataCache(new Map()); // 清空缓存
    } catch {
      message.error(editingUser ? t('adminUser.messages.updateFailure') : t('adminUser.messages.createFailure'));
    }
  };

  // 处理用户删除
  const handleDelete = async (userId: string) => {
    try {
      await apiClient.deleteUser(userId);
      message.success(t('adminUser.messages.deleteSuccess'));
      fetchUsers(pagination.current, pagination.pageSize);
      fetchStats();
      setDataCache(new Map()); // 清空缓存
    } catch {
      message.error(t('adminUser.messages.deleteFailure'));
    }
  };

  // 处理批量删除
  const handleBatchDelete = async () => {
    try {
      await apiClient.batchDeleteUsers(selectedRowKeys as string[]);
      message.success(t('adminUser.messages.batchDeleteSuccess'));
      setSelectedRowKeys([]);
      fetchUsers(pagination.current, pagination.pageSize);
      fetchStats();
      setDataCache(new Map()); // 清空缓存
    } catch {
      message.error(t('adminUser.messages.batchDeleteFailure'));
    }
  };

  // 处理用户状态切换
  const handleStatusToggle = async (userId: string, status: string) => {
    try {
      await apiClient.updateUserStatus(userId, status);
      message.success(t('adminUser.messages.statusUpdateSuccess'));
      fetchUsers(pagination.current, pagination.pageSize);
      fetchStats();
      setDataCache(new Map()); // 清空缓存
    } catch {
      message.error(t('adminUser.messages.statusUpdateFailure'));
    }
  };

  // 处理角色变更
  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await apiClient.updateUserRole(userId, role);
      message.success(t('adminUser.messages.roleUpdateSuccess'));
      fetchUsers(pagination.current, pagination.pageSize);
      fetchStats();
      setDataCache(new Map()); // 清空缓存
    } catch {
      message.error(t('adminUser.messages.roleUpdateFailure'));
    }
  };

  // 打开角色权限分配模态框
  const [rolePermissionModalVisible, setRolePermissionModalVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserData | null>(null);
  const [userRoles, setUserRoles] = useState<string[]>([]);
  const [userPermissions, setUserPermissions] = useState<string[]>([]);

  const handleOpenRolePermissionModal = async (user: UserData) => {
    setSelectedUser(user);
    try {
      // 获取用户当前角色和权限
      const rolesResponse = await permissionService.getUserRoles(user.id);
      const permissionsResponse = await permissionService.getUserPermissions(user.id);
      const assignedRoles = Array.isArray(rolesResponse?.data) ? rolesResponse.data : [];
      setUserRoles(assignedRoles.map((r: any) => r?.id).filter(Boolean));
      setUserPermissions(Array.isArray(permissionsResponse?.data) ? permissionsResponse.data : []);
      setRolePermissionModalVisible(true);
    } catch (error) {
      message.error(t('adminUser.messages.fetchUserRolePermFailure'));
    }
  };

  // 保存用户角色权限
  const handleSaveUserRolePermissions = async () => {
    if (!selectedUser) return;
    
    try {
      await permissionService.batchAssignRolesToUser(selectedUser.id, userRoles);
      message.success(t('adminUser.messages.userRolePermUpdateSuccess'));
      setRolePermissionModalVisible(false);
      fetchUsers(pagination.current, pagination.pageSize);
      fetchStats();
      setDataCache(new Map()); // 清空缓存
    } catch (error) {
      message.error(t('adminUser.messages.userRolePermUpdateFailure'));
    }
  };

  // 颜色与日期统一规范：从 utils/display 引入

  // 查看用户详情
  const handleViewUser = (user: UserData) => {
    setViewingUser(user);
    setViewModalVisible(true);
  };

  // 编辑用户
  const handleEditUser = (user: UserData) => {
    setEditingUser(user);
    setModalVisible(true);
  };

  // 用户操作菜单
  const getUserActionMenu = (user: UserData): MenuProps => ({
    items: [
      {
        key: 'edit',
        icon: <EditOutlined />,
        label: t('adminUser.actions.editUser'),
        onClick: () => handleEditUser(user)
      },
      {
        key: 'status',
        icon: user.status === 'active' ? <LockOutlined /> : <UnlockOutlined />,
        label: user.status === 'active' ? t('adminUser.actions.disableUser') : t('adminUser.actions.enableUser'),
        onClick: () => handleStatusToggle(user.id, user.status === 'active' ? 'inactive' : 'active')
      },
      {
        key: 'permissions',
        icon: <SafetyOutlined />,
        label: t('adminUser.actions.manageRolePermissions'),
        onClick: () => handleOpenRolePermissionModal(user)
      },
      {
        key: 'role',
        icon: <CrownOutlined />,
        label: t('adminUser.rolePermissionModal.roleSwitchQuickTitle'),
        children: roles.map(r => ({
          key: r.code,
          label: r.name,
          onClick: () => handleRoleChange(user.id, r.code)
        }))
      },
      {
        type: 'divider'
      },
      {
        key: 'delete',
        icon: <DeleteOutlined />,
        label: t('adminUser.actions.deleteUser'),
        danger: true,
        onClick: () => {
          Modal.confirm({
            title: t('adminUser.confirm.deleteTitle'),
            content: t('adminUser.confirm.deleteContent', { username: user.username }),
            okText: t('adminUser.confirm.ok'),
            cancelText: t('adminUser.confirm.cancel'),
            onOk: () => handleDelete(user.id)
          });
        }
      }
    ]
  });

  // 表格列定义（记忆化优化）
  const columns: ColumnType<UserData>[] = useMemo(() => [
    {
      title: t('adminUser.table.userInfo'),
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
            <Tooltip title={record.email}>
              <Typography.Text style={{ color: '#666', fontSize: '12px', maxWidth: 200 }} ellipsis>
                {record.email}
              </Typography.Text>
            </Tooltip>
          </div>
        </Space>
      )
    },
    {
      title: t('adminUser.table.username'),
      dataIndex: 'username',
      key: 'username',
      width: 120,
      sorter: (a, b) => a.username.localeCompare(b.username)
    },
    {
      title: t('adminUser.table.role'),
      dataIndex: 'role',
      key: 'role',
      width: 100,
      render: (role: string) => {
        // 根据角色代码查找角色名称
        const roleData = roles.find(r => r.code === role);
        const roleName = roleData?.name || role;
        const isAdmin = role === 'admin' || role === 'super_admin';
        
        return (
          <Tag color={getRoleColor(role)} icon={isAdmin ? <CrownOutlined /> : <UserOutlined />}>
            {roleName}
          </Tag>
        );
      },
      filters: roles.map(role => ({
        text: role.name,
        value: role.code
      }))
    },
    {
      title: t('adminUser.table.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status === 'active' ? t('adminUser.filters.statusOptions.active') : status === 'inactive' ? t('adminUser.filters.statusOptions.inactive') : t('adminUser.filters.statusOptions.banned')}
        </Tag>
      ),
      filters: [
        { text: t('adminUser.filters.statusOptions.active'), value: 'active' },
        { text: t('adminUser.filters.statusOptions.inactive'), value: 'inactive' },
        { text: t('adminUser.filters.statusOptions.banned'), value: 'banned' }
      ]
    },
    {
      title: t('adminUser.table.stats'),
      key: 'stats',
      width: 150,
      render: (_, record) => (
        <Space size="small" wrap>
          <Tag color="blue" icon={<LoginOutlined />}>{t('adminUser.labels.statItems.login', { count: record.loginCount })}</Tag>
          <Tag color="green" icon={<BulbOutlined />}>{t('adminUser.labels.statItems.wisdom', { count: record.wisdomCount })}</Tag>
          <Tag color="purple" icon={<CommentOutlined />}>{t('adminUser.labels.statItems.comment', { count: record.commentCount })}</Tag>
        </Space>
      )
    },
    {
      title: t('adminUser.table.lastLogin'),
      dataIndex: 'lastLogin',
      key: 'lastLogin',
      width: 120,
      render: (date: string) => date ? formatDateTime(date) : t('adminUser.labelsExtra.neverLogin'),
      sorter: (a, b) => {
        if (!a.lastLogin && !b.lastLogin) return 0;
        if (!a.lastLogin) return 1;
        if (!b.lastLogin) return -1;
        return dayjs(a.lastLogin).unix() - dayjs(b.lastLogin).unix();
      }
    },
    {
      title: t('adminUser.table.registeredAt'),
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 120,
      render: (date: string) => formatDate(date),
      sorter: (a, b) => dayjs(a.createdAt).unix() - dayjs(b.createdAt).unix()
    },
    {
      title: t('adminUser.table.actions'),
      key: 'actions',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Tooltip title={t('adminUser.actions.viewDetails')}>
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              size="small"
              onClick={() => handleViewUser(record)}
            />
          </Tooltip>
          <Tooltip title={t('adminUser.actions.editUser')}>
            <Button 
              type="text" 
              icon={<EditOutlined />} 
              size="small"
              onClick={() => handleEditUser(record)}
            />
          </Tooltip>
          <Dropdown menu={getUserActionMenu(record)} trigger={['click']}>
            <Button type="text" icon={<MoreOutlined />} size="small" />
          </Dropdown>
        </Space>
      )
    }
  ], [t, roles, handleStatusToggle, handleRoleChange, handleDelete]);

  // 行选择配置（记忆化优化）
  const rowSelection = useMemo(() => ({
    selectedRowKeys,
    onChange: setSelectedRowKeys,
    getCheckboxProps: (record: UserData) => ({
      disabled: record.role === 'admin' && record.id === 'current-user-id', // 防止删除当前管理员
    }),
  }), [selectedRowKeys]);

  return (
    <div style={{ padding: '24px', minHeight: '100vh', backgroundColor: '#f5f5f5' }}>
      <div style={{ maxWidth: '1400px', margin: '0 auto' }}>
        <Title level={2} style={{ marginBottom: '24px', color: '#262626' }}>
          {t('adminUser.title')}
        </Title>
        
        {/* 统计卡片 */}
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} md={6} lg={6} xl={4}>
            <Card 
              hoverable
              style={{ 
                borderRadius: '8px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                transition: 'all 0.3s ease'
              }}
            >
              <Statistic
                title={t('adminUser.stats.totalUsers')}
                value={stats.totalUsers}
                prefix={<TeamOutlined style={{ color: '#1890ff' }} />}
                valueStyle={{ color: '#1890ff', fontSize: '24px', fontWeight: 'bold' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6} lg={6} xl={4}>
            <Card 
              hoverable
              style={{ 
                borderRadius: '8px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                transition: 'all 0.3s ease'
              }}
            >
              <Statistic
                title={t('adminUser.stats.activeUsers')}
                value={stats.activeUsers}
                prefix={<UserOutlined style={{ color: '#52c41a' }} />}
                valueStyle={{ color: '#52c41a', fontSize: '24px', fontWeight: 'bold' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6} lg={6} xl={4}>
            <Card 
              hoverable
              style={{ 
                borderRadius: '8px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                transition: 'all 0.3s ease'
              }}
            >
              <Statistic
                title={t('adminUser.stats.admins')}
                value={stats.adminUsers}
                prefix={<CrownOutlined style={{ color: '#f5222d' }} />}
                valueStyle={{ color: '#f5222d', fontSize: '24px', fontWeight: 'bold' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6} lg={6} xl={4}>
            <Card 
              hoverable
              style={{ 
                borderRadius: '8px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                transition: 'all 0.3s ease'
              }}
            >
              <Statistic
                title={t('adminUser.stats.newToday')}
                value={stats.newUsersToday}
                prefix={<UserAddOutlined style={{ color: '#faad14' }} />}
                valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 'bold' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6} lg={6} xl={4}>
            <Card 
              hoverable
              style={{ 
                borderRadius: '8px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                transition: 'all 0.3s ease'
              }}
            >
              <Statistic
                title={t('adminUser.stats.onlineUsers')}
                value={stats.onlineUsers}
                prefix={<Badge status="processing" />}
                valueStyle={{ color: '#722ed1', fontSize: '24px', fontWeight: 'bold' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 搜索和筛选 */}
        <Card style={{ marginBottom: '16px', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} sm={12} md={8} lg={6}>
              <Search
                placeholder={t('adminUser.search.placeholder')}
                allowClear
                onSearch={handleSearch}
                style={{ width: '100%' }}
                size="middle"
              />
            </Col>
            <Col xs={24} sm={12} md={8} lg={4}>
              <Select
                placeholder={t('adminUser.filters.rolePlaceholder')}
                allowClear
                style={{ width: '100%' }}
                value={roleFilter}
                onChange={setRoleFilter}
                loading={rolesLoading}
                size="middle"
              >
                {roles.map(role => (
                  <Select.Option key={role.id} value={role.code}>
                    {role.name}
                  </Select.Option>
                ))}
              </Select>
            </Col>
            <Col xs={24} sm={12} md={8} lg={4}>
              <Select
                placeholder={t('adminUser.filters.statusPlaceholder')}
                allowClear
                style={{ width: '100%' }}
                value={statusFilter}
                onChange={setStatusFilter}
                size="middle"
              >
                <Select.Option value="active">{t('adminUser.filters.statusOptions.active')}</Select.Option>
                <Select.Option value="inactive">{t('adminUser.filters.statusOptions.inactive')}</Select.Option>
                <Select.Option value="banned">{t('adminUser.filters.statusOptions.banned')}</Select.Option>
              </Select>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <RangePicker
                placeholder={[t('adminUser.filters.dateRangePlaceholder.0'), t('adminUser.filters.dateRangePlaceholder.1')]}
                style={{ width: '100%' }}
                value={dateRange}
                onChange={setDateRange}
                size="middle"
              />
            </Col>
            <Col xs={24} sm={24} md={24} lg={4}>
              <Space wrap style={{ width: '100%', justifyContent: 'flex-end' }}>
                <Button 
                  onClick={handleReset}
                  icon={<ReloadOutlined />}
                  size="middle"
                >
                  {t('adminUser.actions.reset')}
                </Button>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />}
                  onClick={() => {
                    setEditingUser(null);
                    setModalVisible(true);
                  }}
                  size="middle"
                >
                  {t('adminUser.actions.addUser')}
                </Button>
              </Space>
            </Col>
          </Row>
        </Card>

        {/* 批量操作 */}
        {selectedRowKeys.length > 0 && (
          <Card style={{ marginBottom: '16px', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <Space>
            <span>{t('selection.selectedCount', { count: selectedRowKeys.length })}</span>
            <Popconfirm
              title={t('adminUser.confirm.batchDeleteTitle')}
              okText={t('adminUser.confirm.ok')}
              cancelText={t('adminUser.confirm.cancel')}
              onConfirm={handleBatchDelete}
            >
              <Button danger icon={<DeleteOutlined />}>
                {t('adminUser.actions.batchDelete')}
              </Button>
            </Popconfirm>
            <Button onClick={() => setSelectedRowKeys([])}>
              {t('adminUser.actions.resetSelection')}
            </Button>
          </Space>
          </Card>
        )}

        {/* 用户表格 */}
        <Card style={{ borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          scroll={{ x: 1200, y: 600 }}
          size="middle"
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => t('adminUser.pagination.total', { start: range?.[0] || 0, end: range?.[1] || 0, total }),
            pageSizeOptions: ['10', '20', '50', '100'],
            responsive: true,
            position: ['bottomCenter']
          }}
          onChange={handleTableChange}
          style={{ 
            borderRadius: '8px',
            overflow: 'hidden'
          }}
        />
        </Card>

        {/* 用户编辑/创建模态框 */}
        <Modal
        title={editingUser ? t('adminUser.modal.editTitle') : t('adminUser.modal.addTitle')}
        open={modalVisible}
        destroyOnClose
        forceRender
        onCancel={() => {
          setModalVisible(false);
          setEditingUser(null);
          form.resetFields();
        }}
        footer={null}
        width={600}
        style={{ top: 20 }}
        bodyStyle={{ 
          maxHeight: '70vh', 
          overflowY: 'auto',
          padding: '24px'
        }}
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
                label={t('adminUser.form.usernameLabel')}
                rules={[{ required: true, message: t('adminUser.form.usernameRequired') }]}
              >
                <Input placeholder={t('adminUser.form.usernameRequired')} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="email"
                label={t('adminUser.form.emailLabel')}
                rules={[
                  { required: true, message: t('adminUser.form.emailRequired') },
                  { type: 'email', message: t('adminUser.form.emailInvalid') }
                ]}
              >
                <Input placeholder={t('adminUser.form.emailRequired')} />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item
            name="name"
            label={t('adminUser.form.realNameLabel')}
          >
            <Input placeholder={t('adminUser.form.realNameLabel')} />
          </Form.Item>

          {!editingUser && (
            <Form.Item
              name="password"
              label={t('adminUser.form.passwordLabel')}
              rules={[{ required: true, message: t('adminUser.form.passwordRequired') }]}
            >
              <Input.Password placeholder={t('adminUser.form.passwordRequired')} />
            </Form.Item>
          )}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="role"
                label={t('adminUser.form.roleLabel')}
                rules={[{ required: true, message: t('adminUser.form.roleRequired') }]}
              >
                <Select placeholder={t('adminUser.form.roleRequired')} loading={rolesLoading}>
                  {roles.map(role => (
                    <Select.Option key={role.id} value={role.code}>
                      {role.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="status"
                label={t('adminUser.form.statusLabel')}
                rules={[{ required: true, message: t('adminUser.form.statusRequired') }]}
              >
                <Select placeholder={t('adminUser.form.statusRequired')}>
                  <Select.Option value="active">{t('adminUser.filters.statusOptions.active')}</Select.Option>
                  <Select.Option value="inactive">{t('adminUser.filters.statusOptions.inactive')}</Select.Option>
                  <Select.Option value="banned">{t('adminUser.filters.statusOptions.banned')}</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingUser ? t('adminUser.form.update') : t('adminUser.form.create')}
              </Button>
              <Button onClick={() => {
                setModalVisible(false);
                setEditingUser(null);
                form.resetFields();
              }}>
                {t('adminUser.form.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

        {/* 角色权限分配模态框 */}
        <Modal
        title={t('adminUser.rolePermissionModal.title', { username: selectedUser?.username ?? '' })}
        open={rolePermissionModalVisible}
        onOk={handleSaveUserRolePermissions}
        onCancel={() => {
          setRolePermissionModalVisible(false);
          setSelectedUser(null);
          setUserRoles([]);
          setUserPermissions([]);
        }}
        width={800}
        style={{ top: 20 }}
        bodyStyle={{ 
          maxHeight: '70vh', 
          overflowY: 'auto',
          padding: '24px'
        }}
        okText={t('adminUser.actions.save')}
        cancelText={t('adminUser.actions.cancel')}
      >
        <div style={{ marginBottom: 24 }}>
          <Title level={5}>{t('adminUser.rolePermissionModal.userInfoTitle')}</Title>
          <Space>
            <Avatar src={selectedUser?.avatar} icon={<UserOutlined />} />
            <div>
              <div><strong>{selectedUser?.name || selectedUser?.username}</strong></div>
              <div style={{ color: '#666' }}>{selectedUser?.email}</div>
            </div>
          </Space>
        </div>

        <Row gutter={24}>
          <Col span={12}>
          <Title level={5}>{t('adminUser.rolePermissionModal.roleAssignTitle')}</Title>
          <Select
            mode="multiple"
            placeholder={t('adminUser.rolePermissionModal.roleSelectPlaceholder')}
            style={{ width: '100%' }}
            value={userRoles}
            onChange={setUserRoles}
            loading={rolesLoading}
          >
            {roles.map(role => (
              <Select.Option key={role.id} value={role.id}>
                <Space>
                  {role.is_system && <Tag size="small" color="blue">系统</Tag>}
                  {role.name}
                  {role.description && (
                    <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
                      ({role.description})
                    </Typography.Text>
                  )}
                </Space>
              </Select.Option>
            ))}
          </Select>
          <div style={{ marginTop: 8, color: '#666', fontSize: '12px' }}>
              {t('adminUser.rolePermissionModal.roleAssignDescription')}
          </div>
          </Col>
          
          <Col span={12}>
            <Title level={5}>{t('adminUser.rolePermissionModal.permissionAssignTitle')}</Title>
            <Select
              mode="multiple"
              disabled
              placeholder={t('adminUser.rolePermissionModal.permissionSelectPlaceholder')}
              style={{ width: '100%' }}
              value={userPermissions}
            >
              <Select.OptGroup label={t('adminUser.permissions.groups.userManagement')}>
                <Select.Option value="user:read">{t('adminUser.permissions.items.read', { name: t('adminUser.permissions.resources.user') })}</Select.Option>
                <Select.Option value="user:write">{t('adminUser.permissions.items.write', { name: t('adminUser.permissions.resources.user') })}</Select.Option>
                <Select.Option value="user:delete">{t('adminUser.permissions.items.delete', { name: t('adminUser.permissions.resources.user') })}</Select.Option>
              </Select.OptGroup>
              <Select.OptGroup label={t('adminUser.permissions.groups.roleManagement')}>
                <Select.Option value="role:read">{t('adminUser.permissions.items.read', { name: t('adminUser.permissions.resources.role') })}</Select.Option>
                <Select.Option value="role:write">{t('adminUser.permissions.items.write', { name: t('adminUser.permissions.resources.role') })}</Select.Option>
                <Select.Option value="role:delete">{t('adminUser.permissions.items.delete', { name: t('adminUser.permissions.resources.role') })}</Select.Option>
              </Select.OptGroup>
              <Select.OptGroup label={t('adminUser.permissions.groups.permissionManagement')}>
                <Select.Option value="permission:read">{t('adminUser.permissions.items.read', { name: t('adminUser.permissions.resources.permission') })}</Select.Option>
                <Select.Option value="permission:write">{t('adminUser.permissions.items.write', { name: t('adminUser.permissions.resources.permission') })}</Select.Option>
                <Select.Option value="permission:delete">{t('adminUser.permissions.items.delete', { name: t('adminUser.permissions.resources.permission') })}</Select.Option>
              </Select.OptGroup>
              <Select.OptGroup label={t('adminUser.permissions.groups.menuManagement')}>
                <Select.Option value="menu:read">{t('adminUser.permissions.items.read', { name: t('adminUser.permissions.resources.menu') })}</Select.Option>
                <Select.Option value="menu:write">{t('adminUser.permissions.items.write', { name: t('adminUser.permissions.resources.menu') })}</Select.Option>
                <Select.Option value="menu:delete">{t('adminUser.permissions.items.delete', { name: t('adminUser.permissions.resources.menu') })}</Select.Option>
              </Select.OptGroup>
            </Select>
            <div style={{ marginTop: 8, color: '#666', fontSize: '12px' }}>
              {t('adminUser.rolePermissionModal.permissionNote')}
            </div>
          </Col>
        </Row>
        </Modal>

        {/* 用户详情查看模态框 */}
        <Modal
          title={t('adminUser.modal.viewTitle')}
          open={viewModalVisible}
          onCancel={() => {
            setViewModalVisible(false);
            setViewingUser(null);
          }}
          footer={[
            <Button key="close" onClick={() => {
              setViewModalVisible(false);
              setViewingUser(null);
            }}>
              {t('adminUser.actions.close')}
            </Button>,
            <Button 
              key="edit" 
              type="primary" 
              icon={<EditOutlined />}
              onClick={() => {
                if (viewingUser) {
                  setViewModalVisible(false);
                  handleEditUser(viewingUser);
                }
              }}
            >
              {t('adminUser.actions.editUser')}
            </Button>
          ]}
          width={600}
        >
          {viewingUser && (
            <div style={{ padding: '16px 0' }}>
              <Row gutter={[16, 16]}>
                <Col span={24} style={{ textAlign: 'center', marginBottom: '24px' }}>
                  <Avatar 
                    size={80} 
                    src={viewingUser.avatar} 
                    icon={<UserOutlined />}
                    style={{ marginBottom: '16px' }}
                  />
                  <div>
                    <Typography.Title level={4} style={{ margin: 0 }}>
                      {viewingUser.name || viewingUser.username}
                    </Typography.Title>
                    <Typography.Text type="secondary">
                      {viewingUser.email}
                    </Typography.Text>
                  </div>
                </Col>
                
                <Col span={12}>
                  <Typography.Text strong>{t('adminUser.form.usernameLabel')}:</Typography.Text>
                  <div style={{ marginBottom: '12px' }}>{viewingUser.username}</div>
                </Col>
                
                <Col span={12}>
                  <Typography.Text strong>{t('adminUser.form.emailLabel')}:</Typography.Text>
                  <div style={{ marginBottom: '12px' }}>{viewingUser.email}</div>
                </Col>
                
                <Col span={12}>
                  <Typography.Text strong>{t('adminUser.table.role')}:</Typography.Text>
                  <div style={{ marginBottom: '12px' }}>
                    <Tag color={getRoleColor(viewingUser.role)} icon={viewingUser.role === 'admin' || viewingUser.role === 'super_admin' ? <CrownOutlined /> : <UserOutlined />}>
                      {roles.find(r => r.code === viewingUser.role)?.name || viewingUser.role}
                    </Tag>
                  </div>
                </Col>
                
                <Col span={12}>
                  <Typography.Text strong>{t('adminUser.table.status')}:</Typography.Text>
                  <div style={{ marginBottom: '12px' }}>
                    <Tag color={getStatusColor(viewingUser.status)}>
                      {viewingUser.status === 'active' ? t('adminUser.filters.statusOptions.active') : 
                       viewingUser.status === 'inactive' ? t('adminUser.filters.statusOptions.inactive') : 
                       t('adminUser.filters.statusOptions.banned')}
                    </Tag>
                  </div>
                </Col>
                
                <Col span={12}>
                  <Typography.Text strong>{t('adminUser.table.lastLogin')}:</Typography.Text>
                  <div style={{ marginBottom: '12px' }}>
                    {viewingUser.lastLogin ? formatDateTime(viewingUser.lastLogin) : t('adminUser.labelsExtra.neverLogin')}
                  </div>
                </Col>
                
                <Col span={12}>
                  <Typography.Text strong>{t('adminUser.table.registeredAt')}:</Typography.Text>
                  <div style={{ marginBottom: '12px' }}>
                    {formatDate(viewingUser.createdAt)}
                  </div>
                </Col>
                
                <Col span={24}>
                  <Typography.Text strong>{t('adminUser.table.stats')}:</Typography.Text>
                  <div style={{ marginTop: '8px' }}>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Statistic 
                          title={t('adminUser.labels.statItems.login', { count: '' }).replace(/\d+/, '')}
                          value={viewingUser.loginCount || 0}
                          valueStyle={{ fontSize: '16px' }}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic 
                          title={t('adminUser.labels.statItems.wisdom', { count: '' }).replace(/\d+/, '')}
                          value={viewingUser.wisdomCount || 0}
                          valueStyle={{ fontSize: '16px' }}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic 
                          title={t('adminUser.labels.statItems.comment', { count: '' }).replace(/\d+/, '')}
                          value={viewingUser.commentCount || 0}
                          valueStyle={{ fontSize: '16px' }}
                        />
                      </Col>
                    </Row>
                  </div>
                </Col>
              </Row>
            </div>
          )}
        </Modal>
      </div>
    </div>
  );
};

export default UserManagement;