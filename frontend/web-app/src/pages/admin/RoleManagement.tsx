import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
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
  Select,
  Switch,
  Tree,
  Tabs,
  Divider,
  Tooltip,
  Badge,
  Alert,
  Spin,
  Checkbox,
  Empty,
  Descriptions,
} from 'antd';
import {
  PlusOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  SettingOutlined,
  TeamOutlined,
  SafetyOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import type { ColumnType } from 'antd/es/table';
import { permissionService, type Role, type Permission } from '../../services/permissionService';
import { getStatusColor } from '../../utils/display';
import { PermissionGuard } from '../../components/auth/PermissionGuard';
import { PERMISSIONS } from '../../hooks/usePermissions';

const { Search } = Input;
const { Title } = Typography;
const { Option } = Select;
const { TabPane } = Tabs;

interface RoleFormData {
  name: string;
  code: string;
  description?: string;
  type: 'system' | 'custom' | 'functional' | 'data';
  level: number;
  parent_id?: string;
  permissions: string[];
  is_active: boolean;
}

const RoleManagement: React.FC = () => {
  const { t } = useTranslation();
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [permissionsLoading, setPermissionsLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [permissionModalVisible, setPermissionModalVisible] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [detailRole, setDetailRole] = useState<Role | null>(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [searchText, setSearchText] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [activeFilter, setActiveFilter] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [form] = Form.useForm();
  const [permissionForm] = Form.useForm();
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);

  // 检查角色代码是否重复
  const checkRoleCodeExists = useCallback(async (code: string, excludeId?: string) => {
    try {
      const response = await permissionService.getRoles({ code });
      const existingRoles = response.data || [];
      return existingRoles.some((role: Role) => role.code === code && role.id !== excludeId);
    } catch (error) {
      console.error('Failed to check role code:', error);
      return false;
    }
  }, []);

  // 检查角色名称是否重复
  const checkRoleNameExists = useCallback(async (name: string, excludeId?: string) => {
    try {
      const response = await permissionService.getRoles({ name });
      const existingRoles = response.data || [];
      return existingRoles.some((role: Role) => role.name === name && role.id !== excludeId);
    } catch (error) {
      console.error('Failed to check role name:', error);
      return false;
    }
  }, []);

  // 自定义验证器：角色代码唯一性
  const validateRoleCode = useCallback(async (_: any, value: string) => {
    if (!value) return Promise.resolve();
    
    const exists = await checkRoleCodeExists(value, editingRole?.id);
    if (exists) {
      return Promise.reject(new Error(t('adminRole.validation.codeExists')));
    }
    return Promise.resolve();
  }, [checkRoleCodeExists, editingRole?.id, t]);

  // 自定义验证器：角色名称唯一性
  const validateRoleName = useCallback(async (_: any, value: string) => {
    if (!value) return Promise.resolve();
    
    const exists = await checkRoleNameExists(value, editingRole?.id);
    if (exists) {
      return Promise.reject(new Error(t('adminRole.validation.nameExists')));
    }
    return Promise.resolve();
  }, [checkRoleNameExists, editingRole?.id, t]);

  // 自定义验证器：角色代码格式
  const validateRoleCodeFormat = useCallback((_: any, value: string) => {
    if (!value) return Promise.resolve();
    
    // 检查是否以字母开头
    if (!/^[a-z]/.test(value)) {
      return Promise.reject(new Error(t('adminRole.validation.codeStartWithLetter')));
    }
    
    // 检查是否包含连续下划线
    if (/__/.test(value)) {
      return Promise.reject(new Error(t('adminRole.validation.codeNoConsecutiveUnderscore')));
    }
    
    // 检查是否以下划线结尾
    if (/_$/.test(value)) {
      return Promise.reject(new Error(t('adminRole.validation.codeNoEndingUnderscore')));
    }
    
    return Promise.resolve();
  }, [t]);

  // 自定义验证器：描述内容质量
  const validateDescription = useCallback((_: any, value: string) => {
    if (!value) return Promise.resolve();
    
    // 检查是否只包含空格
    if (value.trim().length === 0) {
      return Promise.reject(new Error(t('adminRole.validation.descriptionNoSpaceOnly')));
    }
    
    // 检查最小有效长度
    if (value.trim().length < 5) {
      return Promise.reject(new Error(t('adminRole.validation.descriptionMinLength')));
    }
    
    return Promise.resolve();
  }, [t]);

  // 获取角色列表
  const fetchRoles = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params: any = {};
      if (searchText) params.search = searchText;
      if (typeFilter) params.type = typeFilter;
      if (activeFilter !== '') params.is_active = activeFilter === 'true';

      const response = await permissionService.getRoles(params);
      const rolesData = response.data || [];
      setRoles(rolesData);
      
      if (rolesData.length === 0 && (searchText || typeFilter || activeFilter)) {
        message.info(t('adminRole.messages.noDataFound'));
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.message || error?.message || t('adminRole.messages.fetchRolesFailure');
      setError(errorMsg);
      message.error(errorMsg);
      console.error('Failed to fetch roles:', error);
    } finally {
      setLoading(false);
    }
  }, [searchText, typeFilter, activeFilter, t]);

  // 获取权限列表
  const fetchPermissions = useCallback(async () => {
    setPermissionsLoading(true);
    try {
      const response = await permissionService.getPermissions();
      setPermissions(response.data || []);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.message || error?.message || t('adminRole.messages.fetchPermissionsFailure');
      message.error(errorMsg);
      console.error('Failed to fetch permissions:', error);
    } finally {
      setPermissionsLoading(false);
    }
  }, [t]);

  // 初始化数据
  useEffect(() => {
    fetchRoles();
    fetchPermissions();
  }, [fetchRoles, fetchPermissions]);

  // 查询条件变化时自动刷新（轻量防抖）
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchRoles();
    }, 250);
    return () => clearTimeout(timer);
  }, [searchText, typeFilter, activeFilter, fetchRoles]);

  // 手动刷新数据
  const handleRefresh = () => {
    fetchRoles();
    fetchPermissions();
  };

  // 防抖搜索函数
  const debouncedSearch = useMemo(() => {
    const timeoutRef = { current: null as NodeJS.Timeout | null };
    return (value: string) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => {
        setSearchText(value);
      }, 300);
    };
  }, []);

  // 处理搜索
  const handleSearch = useCallback((value: string) => {
    debouncedSearch(value);
  }, [debouncedSearch]);

  // 处理筛选
  const handleFilterChange = (type: string, value: string) => {
    if (type === 'type') {
      setTypeFilter(value);
    } else if (type === 'active') {
      setActiveFilter(value);
    }
  };

  // 清空筛选条件
  const handleClearFilters = () => {
    setSearchText('');
    setTypeFilter('');
    setActiveFilter('');
  };

  // 打开新增/编辑模态框
  const handleOpenModal = (role?: Role) => {
    setEditingRole(role || null);
    if (role) {
      form.setFieldsValue({
        ...role,
        permissions: role.permissions || [],
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        type: 'custom',
        level: 1,
        is_active: true,
        permissions: [],
      });
    }
    setModalVisible(true);
  };

  // 打开详情模态框
  const handleOpenDetailModal = (role: Role) => {
    setDetailRole(role);
    setDetailModalVisible(true);
  };

  const handleCloseDetailModal = () => {
    setDetailModalVisible(false);
    setDetailRole(null);
  };

  // 关闭模态框
  const handleCloseModal = () => {
    setModalVisible(false);
    setEditingRole(null);
    form.resetFields();
  };

  // 保存角色
  const handleSaveRole = async () => {
    try {
      const values = await form.validateFields();
      const roleData: RoleFormData = {
        ...values,
      };

      if (editingRole) {
        await permissionService.updateRole(editingRole.id, roleData);
        message.success(t('adminRole.messages.updateRoleSuccess'));
      } else {
        await permissionService.createRole(roleData);
        message.success(t('adminRole.messages.createRoleSuccess'));
      }

      handleCloseModal();
      fetchRoles();
    } catch (error: any) {
      const errorMsg = error?.response?.data?.message || error?.message || t('adminRole.messages.saveRoleFailure');
      message.error(errorMsg);
      console.error('Failed to save role:', error);
    }
  };

  // 删除角色
  const handleDeleteRole = async (roleId: string) => {
    try {
      setLoading(true);
      await permissionService.deleteRole(roleId);
      message.success(t('adminRole.messages.deleteSuccess'));
      await fetchRoles();
    } catch (error: any) {
      console.error('删除角色失败:', error);
      message.error(error.response?.data?.message || t('adminRole.messages.deleteError'));
    } finally {
      setLoading(false);
    }
  };

  // 批量删除角色
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning(t('adminRole.messages.selectRolesToDelete'));
      return;
    }

    Modal.confirm({
      title: t('adminRole.confirm.batchDeleteTitle'),
      content: (
        <div>
          <Alert
            message={t('adminRole.messages.warning')}
            description={t('adminRole.messages.batchDeleteWarning', { count: selectedRowKeys.length })}
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
          <Typography.Text type="secondary">
            {t('adminRole.messages.rolesToDelete')}：
          </Typography.Text>
          <div style={{ marginTop: 8, maxHeight: 200, overflowY: 'auto' }}>
            {selectedRowKeys.map(id => {
              const role = roles.find(r => r.id === id);
              return role ? (
                <Tag key={id} style={{ margin: '2px' }}>
                  {role.name} ({role.code})
                </Tag>
              ) : null;
            })}
          </div>
        </div>
      ),
      icon: <DeleteOutlined style={{ color: '#ff4d4f' }} />,
      okText: t('adminRole.confirm.ok'),
      okType: 'danger',
      cancelText: t('adminRole.confirm.cancel'),
      width: 500,
      onOk: async () => {
        try {
          setLoading(true);
          // 并发删除所有选中的角色
          await Promise.all(
            selectedRowKeys.map(id => permissionService.deleteRole(id as string))
          );
          message.success(t('adminRole.messages.batchDeleteSuccess', { count: selectedRowKeys.length }));
          setSelectedRowKeys([]);
          await fetchRoles();
        } catch (error: any) {
          console.error('批量删除失败:', error);
          message.error(error.response?.data?.message || t('adminRole.messages.batchDeleteFailure'));
        } finally {
          setLoading(false);
        }
      }
    });
  };

  // 打开权限配置模态框
  const handleOpenPermissionModal = async (role: Role) => {
    setSelectedRole(role);
    try {
      const response = await permissionService.getRolePermissions(role.id);
      const rolePermissions = response.data?.map((p: any) => p.id) || [];
      permissionForm.setFieldsValue({
        permissions: rolePermissions,
      });
    } catch (error: any) {
      const errorMsg = error?.response?.data?.message || error?.message || 'Failed to fetch role permissions';
      message.error(errorMsg);
      console.error('Failed to fetch role permissions:', error);
    }
    setPermissionModalVisible(true);
  };

  // 保存权限配置
  const handleSavePermissions = async () => {
    if (!selectedRole) return;

    try {
      const values = await permissionForm.validateFields();
      await permissionService.batchAssignPermissionsToRole(
        selectedRole.id,
        values.permissions
      );
      message.success(t('adminRole.messages.savePermissionsSuccess'));
      setPermissionModalVisible(false);
      fetchRoles();
    } catch (error: any) {
      const errorMsg = error?.response?.data?.message || error?.message || t('adminRole.messages.savePermissionsFailure');
      message.error(errorMsg);
      console.error('Failed to save permissions:', error);
    }
  };

  // 表格列定义（记忆化）
  const columns: ColumnType<Role>[] = useMemo(() => [
    {
      title: t('adminRole.table.name'),
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <Tooltip title={text}>
          <Typography.Text ellipsis style={{ maxWidth: 200, fontWeight: 500 }}>{text}</Typography.Text>
        </Tooltip>
      ),
    },
    {
      title: t('adminRole.table.code'),
      dataIndex: 'code',
      key: 'code',
      render: (text) => <code style={{ whiteSpace: 'nowrap' }}>{text}</code>,
    },
    {
      title: t('adminRole.table.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type) => {
        const typeMap = {
          system: { color: 'blue', text: '系统角色' },
          custom: { color: 'green', text: '自定义角色' },
          functional: { color: 'orange', text: '功能角色' },
          data: { color: 'purple', text: '数据角色' },
        } as const;
        const config = typeMap[type as keyof typeof typeMap];
        return <Tag color={config?.color}>{config?.text || type}</Tag>;
      },
    },
    {
      title: t('adminRole.table.level'),
      dataIndex: 'level',
      key: 'level',
      render: (level) => (
        <Space>
          <Badge count={level} color="blue" />
          <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
            {level === 1 && '基础权限'}
            {level === 2 && '一般权限'}
            {level === 3 && '中等权限'}
            {level === 4 && '高级权限'}
            {level === 5 && '最高权限'}
          </Typography.Text>
        </Space>
      ),
    },
    {
      title: t('adminRole.table.permissionsCount'),
      dataIndex: 'permissions',
      key: 'permissions',
      render: (permissions) => {
        const count = permissions?.length || 0;
        return (
          <Space>
            <Badge count={count} color="green" />
            <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
              {count === 0 ? '无权限' : count === 1 ? '1项权限' : `${count}项权限`}
            </Typography.Text>
          </Space>
        );
      },
    },
    {
      title: t('adminRole.table.status'),
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive) => (
        <Tag color={getStatusColor(isActive ? 'active' : 'inactive')}>
          {isActive ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: t('adminRole.table.description'),
      dataIndex: 'description',
      key: 'description',
      render: (text?: string) => (
        text ? (
          <Tooltip title={text}>
            <Typography.Text ellipsis style={{ maxWidth: 240 }}>{text}</Typography.Text>
          </Tooltip>
        ) : (
          <span style={{ color: '#999' }}>暂无描述</span>
        )
      ),
    },
    {
      title: t('adminRole.table.actions'),
      key: 'actions',
      render: (_, record) => (
        <Space>
          <PermissionGuard permission={PERMISSIONS.ROLE_READ}>
            <Tooltip title="查看详情">
              <Button
                type="text"
                icon={<EyeOutlined />}
                onClick={() => handleOpenDetailModal(record)}
              />
            </Tooltip>
          </PermissionGuard>
          
          <PermissionGuard permission={PERMISSIONS.ROLE_WRITE}>
            <Tooltip title="编辑角色">
              <Button
                type="text"
                icon={<EditOutlined />}
                onClick={() => handleOpenModal(record)}
              />
            </Tooltip>
          </PermissionGuard>

          <PermissionGuard permission={PERMISSIONS.PERMISSION_MANAGE}>
            <Tooltip title="配置权限">
              <Button
                type="text"
                icon={<SettingOutlined />}
                onClick={() => handleOpenPermissionModal(record)}
              />
            </Tooltip>
          </PermissionGuard>

          <PermissionGuard permission={PERMISSIONS.ROLE_DELETE}>
            {!record.is_system && (
              <Popconfirm
                title="确定要删除这个角色吗？"
                description="删除后将无法恢复，请谨慎操作"
                onConfirm={() => handleDeleteRole(record.id)}
                okText="确定"
                cancelText="取消"
              >
                <Tooltip title="删除角色">
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                  />
                </Tooltip>
              </Popconfirm>
            )}
          </PermissionGuard>
        </Space>
      ),
    },
  ], [t, handleOpenModal, handleOpenPermissionModal, handleOpenDetailModal, handleDeleteRole]);

  // 行选择记忆化
  const rowSelection = useMemo(() => ({
    selectedRowKeys,
    onChange: setSelectedRowKeys,
    getCheckboxProps: (record: Role) => ({
      disabled: record.is_system,
    }),
  }), [selectedRowKeys]);

  // 权限树构建（按资源分组）
  const buildPermissionTree = useCallback((perms: Permission[]) => {
    const groups: Record<string, { title: string; key: string; children: any[] }> = {};
    perms.forEach((p) => {
      const [resource, action] = p.name.split(':');
      if (!groups[resource]) {
        groups[resource] = { title: resource, key: resource, children: [] };
      }
      groups[resource].children.push({ title: `${action} (${p.description || p.name})`, key: p.id });
    });
    return Object.values(groups);
  }, []);

  // 权限树数据
  const permissionTreeData = permissions.reduce((acc, permission) => {
    const [resource, action] = permission.name.split(':');
    let resourceNode = acc.find(node => node.key === resource);
    
    if (!resourceNode) {
      resourceNode = {
        title: resource,
        key: resource,
        children: [],
      };
      acc.push(resourceNode);
    }

    resourceNode.children!.push({
      title: `${action} (${permission.description || permission.name})`,
      key: permission.id,
    });

    return acc;
  }, [] as any[]);

  return (
    <div className="role-management">
      <Card>
        <div className="mb-4">
          <Row justify="space-between" align="middle">
            <Col>
              <Title level={4}>
                <TeamOutlined className="mr-2" />
                角色管理
              </Title>
            </Col>
            <Col>
              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={handleRefresh}
                  loading={loading || permissionsLoading}
                >
                  刷新数据
                </Button>
                
                <PermissionGuard permission={PERMISSIONS.ROLE_WRITE}>
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => handleOpenModal()}
                  >
                    新增角色
                  </Button>
                </PermissionGuard>
                
                <PermissionGuard permission={PERMISSIONS.ROLE_DELETE}>
                  {selectedRowKeys.length > 0 && (
                    <Popconfirm
                      title={`确定要删除选中的 ${selectedRowKeys.length} 个角色吗？`}
                      description="删除后无法恢复，请谨慎操作"
                      onConfirm={handleBatchDelete}
                      okText="确定删除"
                      cancelText="取消"
                    >
                      <Button danger>
                        批量删除 ({selectedRowKeys.length})
                      </Button>
                    </Popconfirm>
                  )}
                </PermissionGuard>
              </Space>
            </Col>
          </Row>

          {error && (
            <Alert
              message="数据加载失败"
              description={error}
              type="error"
              showIcon
              closable
              onClose={() => setError('')}
              style={{ marginBottom: 16 }}
            />
          )}

          <Row gutter={16} className="mb-4">
            <Col span={8}>
              <Search
                placeholder="搜索角色名称或代码..."
                allowClear
                onSearch={handleSearch}
                style={{ width: '100%' }}
              />
            </Col>
            <Col span={4}>
              <Select
                placeholder="选择角色类型"
                allowClear
                style={{ width: '100%' }}
                value={typeFilter || undefined}
                onChange={(value) => handleFilterChange('type', value)}
              >
                <Option value="system">系统角色</Option>
                <Option value="custom">自定义角色</Option>
                <Option value="functional">功能角色</Option>
                <Option value="data">数据角色</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="选择启用状态"
                allowClear
                style={{ width: '100%' }}
                value={activeFilter || undefined}
                onChange={(value) => handleFilterChange('active', value)}
              >
                <Option value="true">启用</Option>
                <Option value="false">禁用</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Button onClick={handleClearFilters}>
                清空筛选
              </Button>
            </Col>
          </Row>
        </div>

        <Table
          columns={columns}
          dataSource={roles}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          size="middle"
          scroll={{ x: 1000 }}
          pagination={{
            total: roles.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            responsive: true,
            position: ['bottomCenter'],
            showTotal: (total, range) =>
              `显示第 ${range[0]}-${range[1]} 项，共 ${total} 项`,
          }}
        />
      </Card>

      {/* 新增/编辑角色模态框 */}
      <Modal
        title={editingRole ? '编辑角色' : '新增角色'}
        open={modalVisible}
        onOk={handleSaveRole}
        onCancel={handleCloseModal}
        width={600}
        destroyOnClose
        confirmLoading={loading}
        okText="保存"
        cancelText="取消"
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            type: 'custom',
            level: 1,
            is_active: true,
          }}
        >
          <Row gutter={16}>
             <Col span={12}>
               <Form.Item
                 name="name"
                 label="角色名称"
                 rules={[
                   { required: true, message: '请输入角色名称' },
                   { min: 2, max: 50, message: '角色名称长度应在2-50个字符之间' },
                   { 
                     pattern: /^[\u4e00-\u9fa5a-zA-Z0-9\s\-_]+$/, 
                     message: '角色名称只能包含中文、英文、数字、空格、连字符和下划线' 
                   },
                   { validator: validateRoleName }
                 ]}
                 hasFeedback
               >
                 <Input 
                   placeholder="请输入角色名称，如：内容管理员"
                   showCount
                   maxLength={50}
                 />
               </Form.Item>
             </Col>
             <Col span={12}>
               <Form.Item
                 name="code"
                 label="角色代码"
                 rules={[
                   { required: true, message: '请输入角色代码' },
                   { min: 2, max: 30, message: '角色代码长度应在2-30个字符之间' },
                   { pattern: /^[a-z][a-z0-9_]*$/, message: '只能包含小写字母、数字和下划线，且必须以字母开头' },
                   { validator: validateRoleCodeFormat },
                   { validator: validateRoleCode }
                 ]}
                 hasFeedback
                 extra="角色代码用于系统内部标识，建议使用英文，如：content_manager"
               >
                 <Input 
                   placeholder="请输入角色代码，如：content_manager"
                   showCount
                   maxLength={30}
                   style={{ fontFamily: 'monospace' }}
                   disabled={editingRole?.is_system}
                 />
               </Form.Item>
             </Col>
           </Row>

           <Row gutter={16}>
             <Col span={12}>
               <Form.Item
                 name="type"
                 label="角色类型"
                 rules={[{ required: true, message: '请选择角色类型' }]}
                 extra="选择角色类型以便更好地管理和分类"
               >
                 <Select placeholder="请选择角色类型">
                   <Option value="custom">
                     <Space>
                       <span>自定义角色</span>
                       <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
                         - 用户自定义的业务角色
                       </Typography.Text>
                     </Space>
                   </Option>
                   <Option value="functional">
                     <Space>
                       <span>功能角色</span>
                       <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
                         - 基于功能模块的角色
                       </Typography.Text>
                     </Space>
                   </Option>
                   <Option value="data">
                     <Space>
                       <span>数据角色</span>
                       <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
                         - 基于数据访问权限的角色
                       </Typography.Text>
                     </Space>
                   </Option>
                 </Select>
               </Form.Item>
             </Col>
             <Col span={12}>
               <Form.Item
                 name="level"
                 label="角色级别"
                 rules={[{ required: true, message: '请选择角色级别' }]}
                 extra="级别越高，权限范围越大"
               >
                 <Select placeholder="请选择角色级别">
                   {[1, 2, 3, 4, 5].map(level => (
                     <Option key={level} value={level}>
                       <Space>
                         <Badge count={level} color="blue" />
                         <span>级别 {level}</span>
                         <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
                           {level === 1 && '- 基础权限'}
                           {level === 2 && '- 一般权限'}
                           {level === 3 && '- 中等权限'}
                           {level === 4 && '- 高级权限'}
                           {level === 5 && '- 最高权限'}
                         </Typography.Text>
                       </Space>
                     </Option>
                   ))}
                 </Select>
               </Form.Item>
             </Col>
           </Row>

           <Form.Item 
             name="description" 
             label="角色描述"
             rules={[
               { max: 200, message: '描述长度不能超过200个字符' },
               { validator: validateDescription }
             ]}
             extra="详细描述角色的职责和用途，有助于后续管理"
           >
             <Input.TextArea
               placeholder="请详细描述该角色的职责和权限范围，例如：负责内容的创建、编辑和发布，可以管理文章分类和标签..."
               rows={4}
               showCount
               maxLength={200}
             />
           </Form.Item>

           <Form.Item 
             name="is_active" 
             label="启用状态"
             valuePropName="checked"
             extra="禁用的角色将无法分配给用户"
           >
             <Switch 
               checkedChildren="启用" 
               unCheckedChildren="禁用"
             />
           </Form.Item>
        </Form>
      </Modal>

      {/* 查看详情模态框 */}
      <Modal
        title="角色详情"
        open={detailModalVisible}
        onCancel={handleCloseDetailModal}
        footer={null}
        width={600}
        destroyOnClose
      >
        {detailRole && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="角色名称">{detailRole.name}</Descriptions.Item>
            <Descriptions.Item label="角色代码">
              <code style={{ whiteSpace: 'nowrap' }}>{detailRole.code}</code>
            </Descriptions.Item>
            <Descriptions.Item label="角色类型">
              {detailRole.type === 'system' ? '系统角色' :
               detailRole.type === 'custom' ? '自定义角色' :
               detailRole.type === 'functional' ? '功能角色' :
               detailRole.type === 'data' ? '数据角色' : detailRole.type}
            </Descriptions.Item>
            <Descriptions.Item label="角色级别">
              <Space>
                <Badge count={detailRole.level} color="blue" />
                <span>级别 {detailRole.level}</span>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="启用状态">
              <Tag color={getStatusColor(detailRole.is_active ? 'active' : 'inactive')}>
                {detailRole.is_active ? '启用' : '禁用'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="权限数量">
              <Space>
                <Badge count={detailRole.permissions?.length || 0} color="green" />
                <span>
                  {detailRole.permissions?.length === 0 ? '无权限' :
                   detailRole.permissions?.length === 1 ? '1项权限' :
                   `${detailRole.permissions?.length}项权限`}
                </span>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="角色描述">
              {detailRole.description || '暂无描述'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* 权限配置模态框 */}
      <Modal
        title={
          <Space>
            <SettingOutlined />
            <span>配置角色权限</span>
            {selectedRole && (
              <Tag color="blue">{selectedRole.name}</Tag>
            )}
          </Space>
        }
        open={permissionModalVisible}
        onCancel={() => {
          setPermissionModalVisible(false);
          setSelectedRole(null);
          setSelectedPermissions([]);
        }}
        onOk={handleSavePermissions}
        confirmLoading={loading}
        width={800}
        okText="保存权限"
        cancelText="取消"
        destroyOnClose
      >
        {selectedRole && (
          <div>
            <Alert
              message="权限配置说明"
              description={
                <div>
                  <Typography.Text>
                    正在为角色 <strong>{selectedRole.name}</strong> 配置权限。请选择该角色需要的权限项。
                  </Typography.Text>
                  <br />
                  <Typography.Text type="secondary" style={{ fontSize: '12px' }}>
                    提示：权限按资源分组显示，您可以选择整个资源组或单个权限项。
                  </Typography.Text>
                </div>
              }
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />

            <div style={{ marginBottom: 16 }}>
              <Space>
                <Button
                   size="small"
                   onClick={() => {
                     const allPermissionIds = permissions.map(p => p.id);
                     setSelectedPermissions(allPermissionIds);
                   }}
                 >
                   全选
                 </Button>
                 <Button
                   size="small"
                   onClick={() => setSelectedPermissions([])}
                 >
                   清空
                 </Button>
                <Typography.Text type="secondary">
                  已选择 {selectedPermissions.length} / {permissions.length} 项权限
                </Typography.Text>
              </Space>
            </div>

            <Spin spinning={permissionsLoading}>
              {permissions.length > 0 ? (
                <Tree
                  checkable
                  checkedKeys={selectedPermissions}
                  onCheck={(checkedKeys) => {
                    setSelectedPermissions(checkedKeys as string[]);
                  }}
                  treeData={buildPermissionTree(permissions)}
                  height={400}
                  style={{
                    border: '1px solid #d9d9d9',
                    borderRadius: '6px',
                    padding: '8px'
                  }}
                />
              ) : (
                <Empty
                  description="暂无可配置的权限"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              )}
            </Spin>

            {selectedPermissions.length > 0 && (
              <div style={{ marginTop: 16 }}>
                <Typography.Text strong>已选择的权限：</Typography.Text>
                <div style={{ 
                  marginTop: 8, 
                  maxHeight: 120, 
                  overflowY: 'auto',
                  border: '1px solid #f0f0f0',
                  borderRadius: '4px',
                  padding: '8px',
                  backgroundColor: '#fafafa'
                }}>
                  {selectedPermissions.map(permId => {
                    const permission = permissions.find(p => p.id === permId);
                    return permission ? (
                      <Tag 
                        key={permId} 
                        style={{ margin: '2px' }}
                        color="blue"
                      >
                        {permission.name}
                      </Tag>
                    ) : null;
                  })}
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default RoleManagement;