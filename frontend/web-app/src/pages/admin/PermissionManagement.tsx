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
  Select,
  Tree,
  Tooltip,
  Badge,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyOutlined,
  EyeOutlined,
  FolderOutlined,
  FileOutlined,
} from '@ant-design/icons';
import type { ColumnType } from 'antd/es/table';
import { permissionService, type Permission } from '../../services/permissionService';
import { PermissionGuard } from '../../components/auth/PermissionGuard';
import { PERMISSIONS } from '../../hooks/usePermissions';
import { useTranslation } from 'react-i18next';
import { formatDateTime } from '../../utils/display';

const { Search } = Input;
const { Title } = Typography;
const { Option } = Select;

interface PermissionFormData {
  name: string;
  resource: string;
  action: string;
  description?: string;
}

const PermissionManagement: React.FC = () => {
  const { t } = useTranslation();
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPermission, setEditingPermission] = useState<Permission | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [searchText, setSearchText] = useState('');
  const [resourceFilter, setResourceFilter] = useState<string>('');
  const [actionFilter, setActionFilter] = useState<string>('');
  const [viewMode, setViewMode] = useState<'table' | 'tree'>('table');
  const [form] = Form.useForm();

  // 获取权限列表
  const fetchPermissions = useCallback(async () => {
    setLoading(true);
    try {
      const params: any = {};
      if (searchText) params.search = searchText;
      if (resourceFilter) params.resource = resourceFilter;

      const response = await permissionService.getPermissions(params);
      setPermissions(response.data || []);
    } catch (error) {
      message.error(t('adminPermission.messages.fetchPermissionsFailure'));
      console.error('Failed to fetch permissions:', error);
    } finally {
      setLoading(false);
    }
  }, [searchText, resourceFilter]);

  useEffect(() => {
    fetchPermissions();
  }, [fetchPermissions]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
  };

  // 处理筛选
  const handleFilterChange = (type: string, value: string) => {
    if (type === 'resource') {
      setResourceFilter(value);
    } else if (type === 'action') {
      setActionFilter(value);
    }
  };

  // 打开新增/编辑模态框
  const handleOpenModal = (permission?: Permission) => {
    setEditingPermission(permission || null);
    if (permission) {
      const [resource, action] = permission.name.split(':');
      form.setFieldsValue({
        ...permission,
        resource,
        action,
      });
    } else {
      form.resetFields();
    }
    setModalVisible(true);
  };

  // 关闭模态框
  const handleCloseModal = () => {
    setModalVisible(false);
    setEditingPermission(null);
    form.resetFields();
  };

  // 保存权限
  const handleSavePermission = async () => {
    try {
      const values = await form.validateFields();
      const permissionData: PermissionFormData = {
        name: `${values.resource}:${values.action}`,
        resource: values.resource,
        action: values.action,
        description: values.description,
      };

      if (editingPermission) {
        await permissionService.updatePermission(editingPermission.id, permissionData);
        message.success(t('adminPermission.messages.updateSuccess'));
      } else {
        await permissionService.createPermission(permissionData);
        message.success(t('adminPermission.messages.createSuccess'));
      }

      handleCloseModal();
      fetchPermissions();
    } catch (error) {
      message.error(editingPermission ? t('adminPermission.messages.updateFailure') : t('adminPermission.messages.createFailure'));
      console.error('Failed to save permission:', error);
    }
  };

  // 删除权限
  const handleDeletePermission = async (permissionId: string) => {
    try {
      await permissionService.deletePermission(permissionId);
      message.success(t('adminPermission.messages.deleteSuccess'));
      fetchPermissions();
    } catch (error) {
      message.error(t('adminPermission.messages.deleteFailure'));
      console.error('Failed to delete permission:', error);
    }
  };

  // 批量删除权限
  const handleBatchDelete = async () => {
    try {
      await Promise.all(
        selectedRowKeys.map(id => permissionService.deletePermission(id as string))
      );
      message.success(t('adminPermission.messages.batchDeleteSuccess'));
      setSelectedRowKeys([]);
      fetchPermissions();
    } catch (error) {
      message.error(t('adminPermission.messages.batchDeleteFailure'));
      console.error('Failed to batch delete permissions:', error);
    }
  };

  // 获取资源列表
  const resources = Array.from(new Set(permissions.map(p => p.resource)));
  
  // 获取操作列表
  const actions = Array.from(new Set(permissions.map(p => p.action)));

  // 过滤权限
  const filteredPermissions = permissions.filter(permission => {
    if (actionFilter && permission.action !== actionFilter) return false;
    return true;
  });

  // 表格列定义
  const columns: ColumnType<Permission>[] = [
    {
      title: t('adminPermission.table.name'),
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <Space>
          <SafetyOutlined />
          <code>{text}</code>
        </Space>
      ),
    },
    {
      title: t('adminPermission.table.resource'),
      dataIndex: 'resource',
      key: 'resource',
      render: (resource) => <Tag color="blue">{resource}</Tag>,
    },
    {
      title: t('adminPermission.table.action'),
      dataIndex: 'action',
      key: 'action',
      render: (action) => <Tag color="green">{action}</Tag>,
    },
    {
      title: t('adminPermission.table.description'),
      dataIndex: 'description',
      key: 'description',
      render: (text?: string) => (
        text ? (
          <Tooltip title={text}>
            <Typography.Text ellipsis style={{ maxWidth: 240 }}>{text}</Typography.Text>
          </Tooltip>
        ) : (
          <span>-</span>
        )
      ),
    },
    {
      title: t('adminPermission.table.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => formatDateTime(date),
    },
    {
      title: t('adminPermission.table.actions'),
      key: 'actions',
      render: (_, record) => (
        <Space>
          <PermissionGuard permission={PERMISSIONS.PERMISSION_READ}>
            <Tooltip title={t('adminPermission.actions.viewDetails')}>
              <Button
                type="text"
                icon={<EyeOutlined />}
                onClick={() => handleOpenModal(record)}
              />
            </Tooltip>
          </PermissionGuard>
          
          <PermissionGuard permission={PERMISSIONS.PERMISSION_WRITE}>
            <Tooltip title={t('adminPermission.actions.editPermission')}>
              <Button
                type="text"
                icon={<EditOutlined />}
                onClick={() => handleOpenModal(record)}
              />
            </Tooltip>
          </PermissionGuard>

          <PermissionGuard permission={PERMISSIONS.PERMISSION_DELETE}>
            <Popconfirm
              title={t('adminPermission.confirm.deleteTitle')}
              onConfirm={() => handleDeletePermission(record.id)}
              okText={t('adminPermission.confirm.ok')}
              cancelText={t('adminPermission.confirm.cancel')}
            >
              <Tooltip title={t('adminPermission.actions.deletePermission')}>
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                />
              </Tooltip>
            </Popconfirm>
          </PermissionGuard>
        </Space>
      ),
    },
  ];

  // 权限树数据
  const permissionTreeData = resources.map(resource => ({
    title: (
      <Space>
        <FolderOutlined />
        <span>{resource}</span>
        <Badge count={permissions.filter(p => p.resource === resource).length} />
      </Space>
    ),
    key: resource,
    children: permissions
      .filter(p => p.resource === resource)
      .map(permission => ({
        title: (
          <Space>
            <FileOutlined />
            <span>{permission.action}</span>
            <Tag size="small">{permission.name}</Tag>
          </Space>
        ),
        key: permission.id,
        isLeaf: true,
      })),
  }));

  return (
    <div className="permission-management">
      <Card>
        <div className="mb-4">
          <Row justify="space-between" align="middle">
            <Col>
              <Title level={4}>
                <SafetyOutlined className="mr-2" />
                {t('adminPermission.title')}
              </Title>
            </Col>
            <Col>
              <Space>
                <Space.Compact>
                  <Button
                    type={viewMode === 'table' ? 'primary' : 'default'}
                    onClick={() => setViewMode('table')}
                  >
                    {t('adminPermission.actions.tableView')}
                  </Button>
                  <Button
                    type={viewMode === 'tree' ? 'primary' : 'default'}
                    onClick={() => setViewMode('tree')}
                  >
                    {t('adminPermission.actions.treeView')}
                  </Button>
                </Space.Compact>

                <PermissionGuard permission={PERMISSIONS.PERMISSION_WRITE}>
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => handleOpenModal()}
                  >
                    {t('adminPermission.actions.addPermission')}
                  </Button>
                </PermissionGuard>
                
                <PermissionGuard permission={PERMISSIONS.PERMISSION_DELETE}>
                  {selectedRowKeys.length > 0 && (
                    <Popconfirm
                      title={t('adminPermission.confirm.batchDeleteTitle', { count: selectedRowKeys.length })}
                      onConfirm={handleBatchDelete}
                      okText={t('adminPermission.confirm.ok')}
                      cancelText={t('adminPermission.confirm.cancel')}
                    >
                      <Button danger>
                        {t('adminPermission.actions.batchDeleteWithCount', { count: selectedRowKeys.length })}
                      </Button>
                    </Popconfirm>
                  )}
                </PermissionGuard>
              </Space>
            </Col>
          </Row>

          <Row gutter={16} className="mb-4">
            <Col span={6}>
              <Search
                placeholder={t('adminPermission.search.placeholder')}
                allowClear
                onSearch={handleSearch}
                style={{ width: '100%' }}
              />
            </Col>
            <Col span={4}>
              <Select
                placeholder={t('adminPermission.filters.resourcePlaceholder')}
                allowClear
                style={{ width: '100%' }}
                onChange={(value) => handleFilterChange('resource', value)}
              >
                {resources.map(resource => (
                  <Option key={resource} value={resource}>
                    {resource}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder={t('adminPermission.filters.actionPlaceholder')}
                allowClear
                style={{ width: '100%' }}
                onChange={(value) => handleFilterChange('action', value)}
              >
                {actions.map(action => (
                  <Option key={action} value={action}>
                    {action}
                  </Option>
                ))}
              </Select>
            </Col>
          </Row>
        </div>

        {viewMode === 'table' ? (
          <Table
            columns={columns}
            dataSource={filteredPermissions}
            rowKey="id"
            loading={loading}
            rowSelection={{
              selectedRowKeys,
              onChange: setSelectedRowKeys,
            }}
            pagination={{
              total: filteredPermissions.length,
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) =>
                t('adminPermission.pagination.total', { start: range[0], end: range[1], total }),
            }}
          />
        ) : (
          <Tree
            treeData={permissionTreeData}
            defaultExpandAll
            height={600}
            showLine
          />
        )}
      </Card>

      {/* 新增/编辑权限模态框 */}
      <Modal
        title={editingPermission ? t('adminPermission.modal.editTitle') : t('adminPermission.modal.addTitle')}
        open={modalVisible}
        onOk={handleSavePermission}
        onCancel={handleCloseModal}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="resource"
                label={t('adminPermission.form.resourceLabel')}
                rules={[{ required: true, message: t('adminPermission.form.resourceRequired') }]}
              >
                <Select
                  placeholder={t('adminPermission.form.resourceSelectPlaceholder')}
                  showSearch
                  allowClear
                  mode="combobox"
                >
                  {resources.map(resource => (
                    <Option key={resource} value={resource}>
                      {resource}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="action"
                label={t('adminPermission.form.actionLabel')}
                rules={[{ required: true, message: t('adminPermission.form.actionRequired') }]}
              >
                <Select
                  placeholder={t('adminPermission.form.actionSelectPlaceholder')}
                  showSearch
                  allowClear
                  mode="combobox"
                >
                  {actions.map(action => (
                    <Option key={action} value={action}>
                      {action}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label={t('adminPermission.form.descriptionLabel')}>
            <Input.TextArea
              placeholder={t('adminPermission.form.descriptionPlaceholder')}
              rows={3}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PermissionManagement;