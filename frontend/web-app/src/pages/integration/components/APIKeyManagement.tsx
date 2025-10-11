import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  Space,
  Tag,
  Tooltip,
  message,
  Popconfirm,
  Typography,
  Row,
  Col,
  Statistic,
  Alert,
  Checkbox
} from 'antd';
import {
  PlusOutlined,
  KeyOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  BarChartOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;

interface APIKey {
  id: number;
  name: string;
  description: string;
  key?: string;
  prefix: string;
  permissions: string[];
  status: 'active' | 'inactive' | 'expired';
  lastUsedAt?: string;
  expiresAt?: string;
  createdAt: string;
  updatedAt: string;
  usageCount?: number;
}

interface APIKeyUsage {
  totalRequests: number;
  successfulRequests: number;
  failedRequests: number;
  lastRequest?: string;
  dailyUsage: Array<{
    date: string;
    requests: number;
  }>;
}

const APIKeyManagement: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [usageModalVisible, setUsageModalVisible] = useState(false);
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null);
  const [keyUsage, setKeyUsage] = useState<APIKeyUsage | null>(null);
  const [newKeyVisible, setNewKeyVisible] = useState(false);
  const [newKeyValue, setNewKeyValue] = useState('');
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const permissions = [
    { label: '读取数据', value: 'read' },
    { label: '写入数据', value: 'write' },
    { label: '删除数据', value: 'delete' },
    { label: '管理用户', value: 'manage_users' },
    { label: '系统配置', value: 'system_config' },
    { label: '插件管理', value: 'plugin_management' },
    { label: 'Webhook管理', value: 'webhook_management' }
  ];

  useEffect(() => {
    fetchAPIKeys();
  }, []);

  const fetchAPIKeys = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockData: APIKey[] = [
        {
          id: 1,
          name: '生产环境密钥',
          description: '用于生产环境的API访问',
          prefix: 'sk_prod_',
          permissions: ['read', 'write'],
          status: 'active',
          lastUsedAt: '2024-01-15T10:30:00Z',
          expiresAt: '2024-12-31T23:59:59Z',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-15T10:30:00Z',
          usageCount: 1234
        },
        {
          id: 2,
          name: '测试环境密钥',
          description: '用于测试和开发',
          prefix: 'sk_test_',
          permissions: ['read'],
          status: 'active',
          lastUsedAt: '2024-01-14T15:20:00Z',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-14T15:20:00Z',
          usageCount: 567
        },
        {
          id: 3,
          name: '已过期密钥',
          description: '旧版本API密钥',
          prefix: 'sk_old_',
          permissions: ['read', 'write', 'delete'],
          status: 'expired',
          expiresAt: '2023-12-31T23:59:59Z',
          createdAt: '2023-06-01T00:00:00Z',
          updatedAt: '2023-12-31T23:59:59Z',
          usageCount: 89
        }
      ];
      setApiKeys(mockData);
    } catch (error) {
      message.error('获取API密钥列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateKey = async (values: any) => {
    try {
      // 模拟API调用
      const newKey = `sk_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      const newAPIKey: APIKey = {
        id: Date.now(),
        name: values.name,
        description: values.description,
        key: newKey,
        prefix: newKey.split('_').slice(0, 2).join('_') + '_',
        permissions: values.permissions,
        status: 'active',
        expiresAt: values.expiresAt?.format('YYYY-MM-DDTHH:mm:ssZ'),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        usageCount: 0
      };

      setApiKeys([newAPIKey, ...apiKeys]);
      setNewKeyValue(newKey);
      setNewKeyVisible(true);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('API密钥创建成功');
    } catch (error) {
      message.error('创建API密钥失败');
    }
  };

  const handleEditKey = async (values: any) => {
    if (!selectedKey) return;

    try {
      const updatedKeys = apiKeys.map(key =>
        key.id === selectedKey.id
          ? {
              ...key,
              name: values.name,
              description: values.description,
              permissions: values.permissions,
              status: values.status,
              updatedAt: new Date().toISOString()
            }
          : key
      );
      setApiKeys(updatedKeys);
      setEditModalVisible(false);
      setSelectedKey(null);
      editForm.resetFields();
      message.success('API密钥更新成功');
    } catch (error) {
      message.error('更新API密钥失败');
    }
  };

  const handleDeleteKey = async (id: number) => {
    try {
      setApiKeys(apiKeys.filter(key => key.id !== id));
      message.success('API密钥删除成功');
    } catch (error) {
      message.error('删除API密钥失败');
    }
  };

  const handleViewUsage = async (key: APIKey) => {
    setSelectedKey(key);
    setUsageModalVisible(true);
    
    // 模拟获取使用统计
    const mockUsage: APIKeyUsage = {
      totalRequests: key.usageCount || 0,
      successfulRequests: Math.floor((key.usageCount || 0) * 0.95),
      failedRequests: Math.floor((key.usageCount || 0) * 0.05),
      lastRequest: key.lastUsedAt,
      dailyUsage: Array.from({ length: 7 }, (_, i) => ({
        date: dayjs().subtract(6 - i, 'day').format('YYYY-MM-DD'),
        requests: Math.floor(Math.random() * 100)
      }))
    };
    setKeyUsage(mockUsage);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('已复制到剪贴板');
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      active: { color: 'green', text: '活跃' },
      inactive: { color: 'orange', text: '未激活' },
      expired: { color: 'red', text: '已过期' }
    };
    const config = statusConfig[status as keyof typeof statusConfig];
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns: ColumnsType<APIKey> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.description}
          </Text>
        </div>
      )
    },
    {
      title: '密钥前缀',
      dataIndex: 'prefix',
      key: 'prefix',
      render: (text) => (
        <Text code style={{ fontSize: 12 }}>
          {text}***
        </Text>
      )
    },
    {
      title: '权限',
      dataIndex: 'permissions',
      key: 'permissions',
      render: (permissions: string[]) => (
        <div>
          {permissions.slice(0, 2).map(permission => {
            const perm = permissions.find(p => p.value === permission);
            return (
              <Tag key={permission} size="small">
                {perm?.label || permission}
              </Tag>
            );
          })}
          {permissions.length > 2 && (
            <Tag size="small">+{permissions.length - 2}</Tag>
          )}
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: getStatusTag
    },
    {
      title: '使用次数',
      dataIndex: 'usageCount',
      key: 'usageCount',
      render: (count) => count?.toLocaleString() || 0
    },
    {
      title: '最后使用',
      dataIndex: 'lastUsedAt',
      key: 'lastUsedAt',
      render: (date) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '从未使用'
    },
    {
      title: '过期时间',
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      render: (date) => date ? dayjs(date).format('YYYY-MM-DD') : '永不过期'
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="查看使用统计">
            <Button
              type="text"
              icon={<BarChartOutlined />}
              onClick={() => handleViewUsage(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedKey(record);
                editForm.setFieldsValue({
                  name: record.name,
                  description: record.description,
                  permissions: record.permissions,
                  status: record.status
                });
                setEditModalVisible(true);
              }}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个API密钥吗？"
            onConfirm={() => handleDeleteKey(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const activeKeys = apiKeys.filter(key => key.status === 'active').length;
  const totalUsage = apiKeys.reduce((sum, key) => sum + (key.usageCount || 0), 0);

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="总密钥数"
              value={apiKeys.length}
              prefix={<KeyOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="活跃密钥"
              value={activeKeys}
              prefix={<KeyOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="总使用次数"
              value={totalUsage}
              prefix={<BarChartOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <KeyOutlined />
            API密钥管理
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建密钥
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={apiKeys}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`
          }}
        />
      </Card>

      {/* 创建API密钥模态框 */}
      <Modal
        title="创建API密钥"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateKey}
        >
          <Form.Item
            name="name"
            label="密钥名称"
            rules={[{ required: true, message: '请输入密钥名称' }]}
          >
            <Input placeholder="输入密钥名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea
              placeholder="输入密钥描述"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="permissions"
            label="权限"
            rules={[{ required: true, message: '请选择至少一个权限' }]}
          >
            <Checkbox.Group>
              <Row>
                {permissions.map(permission => (
                  <Col span={12} key={permission.value} style={{ marginBottom: 8 }}>
                    <Checkbox value={permission.value}>
                      {permission.label}
                    </Checkbox>
                  </Col>
                ))}
              </Row>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item
            name="expiresAt"
            label="过期时间"
          >
            <DatePicker
              style={{ width: '100%' }}
              placeholder="选择过期时间（可选）"
              disabledDate={(current) => current && current < dayjs().endOf('day')}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建密钥
              </Button>
              <Button onClick={() => {
                setCreateModalVisible(false);
                form.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑API密钥模态框 */}
      <Modal
        title="编辑API密钥"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setSelectedKey(null);
          editForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleEditKey}
        >
          <Form.Item
            name="name"
            label="密钥名称"
            rules={[{ required: true, message: '请输入密钥名称' }]}
          >
            <Input placeholder="输入密钥名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea
              placeholder="输入密钥描述"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="permissions"
            label="权限"
            rules={[{ required: true, message: '请选择至少一个权限' }]}
          >
            <Checkbox.Group>
              <Row>
                {permissions.map(permission => (
                  <Col span={12} key={permission.value} style={{ marginBottom: 8 }}>
                    <Checkbox value={permission.value}>
                      {permission.label}
                    </Checkbox>
                  </Col>
                ))}
              </Row>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              <Option value="active">活跃</Option>
              <Option value="inactive">未激活</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                更新密钥
              </Button>
              <Button onClick={() => {
                setEditModalVisible(false);
                setSelectedKey(null);
                editForm.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 新密钥显示模态框 */}
      <Modal
        title={
          <Space>
            <ExclamationCircleOutlined style={{ color: '#faad14' }} />
            API密钥创建成功
          </Space>
        }
        open={newKeyVisible}
        onCancel={() => {
          setNewKeyVisible(false);
          setNewKeyValue('');
        }}
        footer={[
          <Button
            key="copy"
            type="primary"
            icon={<CopyOutlined />}
            onClick={() => copyToClipboard(newKeyValue)}
          >
            复制密钥
          </Button>,
          <Button
            key="close"
            onClick={() => {
              setNewKeyVisible(false);
              setNewKeyValue('');
            }}
          >
            关闭
          </Button>
        ]}
      >
        <Alert
          message="重要提示"
          description="请立即复制并安全保存您的API密钥。出于安全考虑，密钥只会显示一次，关闭后将无法再次查看完整密钥。"
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Paragraph>
          您的新API密钥：
        </Paragraph>
        <Input.Password
          value={newKeyValue}
          readOnly
          iconRender={(visible) => (visible ? <EyeOutlined /> : <EyeInvisibleOutlined />)}
          style={{ fontFamily: 'monospace' }}
        />
      </Modal>

      {/* 使用统计模态框 */}
      <Modal
        title={`${selectedKey?.name} - 使用统计`}
        open={usageModalVisible}
        onCancel={() => {
          setUsageModalVisible(false);
          setSelectedKey(null);
          setKeyUsage(null);
        }}
        footer={[
          <Button
            key="close"
            onClick={() => {
              setUsageModalVisible(false);
              setSelectedKey(null);
              setKeyUsage(null);
            }}
          >
            关闭
          </Button>
        ]}
        width={800}
      >
        {keyUsage && (
          <div>
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              <Col span={8}>
                <Statistic
                  title="总请求数"
                  value={keyUsage.totalRequests}
                  prefix={<BarChartOutlined />}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="成功请求"
                  value={keyUsage.successfulRequests}
                  valueStyle={{ color: '#3f8600' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="失败请求"
                  value={keyUsage.failedRequests}
                  valueStyle={{ color: '#cf1322' }}
                />
              </Col>
            </Row>
            
            <Title level={5}>最近7天使用情况</Title>
            <Table
              size="small"
              dataSource={keyUsage.dailyUsage}
              pagination={false}
              columns={[
                {
                  title: '日期',
                  dataIndex: 'date',
                  key: 'date'
                },
                {
                  title: '请求次数',
                  dataIndex: 'requests',
                  key: 'requests'
                }
              ]}
            />
          </div>
        )}
      </Modal>
    </div>
  );
};

export default APIKeyManagement;