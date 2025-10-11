import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
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
  Switch,
  Tabs,
  List,
  Badge,
  Divider,
  Descriptions,
  Timeline,
  Progress,
  QRCode
} from 'antd';
import {
  PlusOutlined,
  KeyOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SettingOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  LinkOutlined,
  CopyOutlined,
  ReloadOutlined,
  UserOutlined,
  AppstoreOutlined,
  SafetyOutlined,
  GlobalOutlined,
  QrcodeOutlined,
  HistoryOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

interface OAuthApp {
  id: number;
  name: string;
  clientId: string;
  clientSecret: string;
  redirectUris: string[];
  scopes: string[];
  grantTypes: string[];
  isActive: boolean;
  description?: string;
  website?: string;
  logoUrl?: string;
  status: 'active' | 'inactive' | 'suspended';
  tokenCount: number;
  lastUsed?: string;
  createdAt: string;
  updatedAt: string;
}

interface OAuthToken {
  id: number;
  appId: number;
  userId: number;
  accessToken: string;
  refreshToken?: string;
  scopes: string[];
  expiresAt: string;
  isRevoked: boolean;
  createdAt: string;
  lastUsed?: string;
  userInfo?: {
    id: number;
    email: string;
    name: string;
  };
}

const OAuthManagement: React.FC = () => {
  const [apps, setApps] = useState<OAuthApp[]>([]);
  const [tokens, setTokens] = useState<OAuthToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [secretModalVisible, setSecretModalVisible] = useState(false);
  const [tokensModalVisible, setTokensModalVisible] = useState(false);
  const [selectedApp, setSelectedApp] = useState<OAuthApp | null>(null);
  const [activeTab, setActiveTab] = useState('apps');
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const availableScopes = [
    'read',
    'write',
    'admin',
    'user:read',
    'user:write',
    'data:read',
    'data:write',
    'integration:read',
    'integration:write'
  ];

  const availableGrantTypes = [
    'authorization_code',
    'client_credentials',
    'refresh_token'
  ];

  useEffect(() => {
    fetchApps();
    fetchTokens();
  }, []);

  const fetchApps = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockData: OAuthApp[] = [
        {
          id: 1,
          name: '移动应用',
          clientId: 'app_1234567890abcdef',
          clientSecret: 'secret_abcdef1234567890',
          redirectUris: ['https://app.example.com/callback', 'myapp://oauth/callback'],
          scopes: ['read', 'user:read', 'data:read'],
          grantTypes: ['authorization_code', 'refresh_token'],
          isActive: true,
          description: '官方移动应用OAuth集成',
          website: 'https://app.example.com',
          status: 'active',
          tokenCount: 1250,
          lastUsed: '2024-01-16T10:30:00Z',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-16T10:30:00Z'
        },
        {
          id: 2,
          name: '第三方分析工具',
          clientId: 'analytics_9876543210fedcba',
          clientSecret: 'secret_fedcba0987654321',
          redirectUris: ['https://analytics.partner.com/oauth/callback'],
          scopes: ['read', 'data:read'],
          grantTypes: ['client_credentials'],
          isActive: true,
          description: '数据分析合作伙伴集成',
          website: 'https://analytics.partner.com',
          status: 'active',
          tokenCount: 45,
          lastUsed: '2024-01-16T08:15:00Z',
          createdAt: '2024-01-05T00:00:00Z',
          updatedAt: '2024-01-16T08:15:00Z'
        },
        {
          id: 3,
          name: '测试应用',
          clientId: 'test_abcd1234efgh5678',
          clientSecret: 'secret_test1234abcd5678',
          redirectUris: ['http://localhost:3000/callback'],
          scopes: ['read', 'write'],
          grantTypes: ['authorization_code'],
          isActive: false,
          description: '开发测试用OAuth应用',
          status: 'inactive',
          tokenCount: 0,
          createdAt: '2024-01-10T00:00:00Z',
          updatedAt: '2024-01-10T00:00:00Z'
        }
      ];
      setApps(mockData);
    } catch (error) {
      message.error('获取OAuth应用列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchTokens = async () => {
    try {
      // 模拟Token数据
      const mockTokens: OAuthToken[] = [
        {
          id: 1,
          appId: 1,
          userId: 101,
          accessToken: 'at_1234567890abcdef',
          refreshToken: 'rt_abcdef1234567890',
          scopes: ['read', 'user:read'],
          expiresAt: '2024-01-17T10:30:00Z',
          isRevoked: false,
          createdAt: '2024-01-16T10:30:00Z',
          lastUsed: '2024-01-16T15:20:00Z',
          userInfo: {
            id: 101,
            email: 'user@example.com',
            name: '张三'
          }
        },
        {
          id: 2,
          appId: 1,
          userId: 102,
          accessToken: 'at_fedcba0987654321',
          refreshToken: 'rt_0987654321fedcba',
          scopes: ['read', 'data:read'],
          expiresAt: '2024-01-18T08:15:00Z',
          isRevoked: false,
          createdAt: '2024-01-15T08:15:00Z',
          lastUsed: '2024-01-16T12:45:00Z',
          userInfo: {
            id: 102,
            email: 'user2@example.com',
            name: '李四'
          }
        },
        {
          id: 3,
          appId: 2,
          userId: 103,
          accessToken: 'at_service_token_123',
          scopes: ['data:read'],
          expiresAt: '2024-02-16T08:15:00Z',
          isRevoked: false,
          createdAt: '2024-01-16T08:15:00Z',
          lastUsed: '2024-01-16T08:15:00Z',
          userInfo: {
            id: 103,
            email: 'service@partner.com',
            name: '服务账户'
          }
        }
      ];
      setTokens(mockTokens);
    } catch (error) {
      message.error('获取Token列表失败');
    }
  };

  const handleCreateApp = async (values: any) => {
    try {
      const newApp: OAuthApp = {
        id: Date.now(),
        name: values.name,
        clientId: `app_${Math.random().toString(36).substr(2, 16)}`,
        clientSecret: `secret_${Math.random().toString(36).substr(2, 20)}`,
        redirectUris: values.redirectUris.split('\n').filter((uri: string) => uri.trim()),
        scopes: values.scopes,
        grantTypes: values.grantTypes,
        isActive: values.isActive !== false,
        description: values.description,
        website: values.website,
        status: 'active',
        tokenCount: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      setApps([newApp, ...apps]);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('OAuth应用创建成功');
    } catch (error) {
      message.error('创建OAuth应用失败');
    }
  };

  const handleUpdateApp = async (values: any) => {
    if (!selectedApp) return;

    try {
      setApps(apps.map(app =>
        app.id === selectedApp.id
          ? {
              ...app,
              ...values,
              redirectUris: values.redirectUris.split('\n').filter((uri: string) => uri.trim()),
              updatedAt: new Date().toISOString()
            }
          : app
      ));
      setEditModalVisible(false);
      setSelectedApp(null);
      editForm.resetFields();
      message.success('OAuth应用更新成功');
    } catch (error) {
      message.error('更新OAuth应用失败');
    }
  };

  const handleToggleApp = async (id: number, active: boolean) => {
    try {
      setApps(apps.map(app =>
        app.id === id
          ? { 
              ...app, 
              isActive: active, 
              status: active ? 'active' as const : 'inactive' as const,
              updatedAt: new Date().toISOString() 
            }
          : app
      ));
      message.success(active ? 'OAuth应用已启用' : 'OAuth应用已禁用');
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleDeleteApp = async (id: number) => {
    try {
      setApps(apps.filter(app => app.id !== id));
      setTokens(tokens.filter(token => token.appId !== id));
      message.success('OAuth应用删除成功');
    } catch (error) {
      message.error('删除OAuth应用失败');
    }
  };

  const handleRevokeToken = async (tokenId: number) => {
    try {
      setTokens(tokens.map(token =>
        token.id === tokenId
          ? { ...token, isRevoked: true }
          : token
      ));
      message.success('Token已撤销');
    } catch (error) {
      message.error('撤销Token失败');
    }
  };

  const handleRegenerateSecret = async (appId: number) => {
    try {
      const newSecret = `secret_${Math.random().toString(36).substr(2, 20)}`;
      setApps(apps.map(app =>
        app.id === appId
          ? { ...app, clientSecret: newSecret, updatedAt: new Date().toISOString() }
          : app
      ));
      message.success('Client Secret已重新生成');
    } catch (error) {
      message.error('重新生成Secret失败');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('已复制到剪贴板');
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      active: { color: 'green', text: '活跃', icon: <CheckCircleOutlined /> },
      inactive: { color: 'default', text: '未激活', icon: <ExclamationCircleOutlined /> },
      suspended: { color: 'red', text: '已暂停', icon: <ExclamationCircleOutlined /> }
    };
    const config = statusConfig[status as keyof typeof statusConfig];
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  const appColumns: ColumnsType<OAuthApp> = [
    {
      title: '应用信息',
      key: 'info',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 500, marginBottom: 4 }}>
            {record.name}
          </div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Client ID: {record.clientId}
          </Text>
          {record.description && (
            <div style={{ marginTop: 4 }}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {record.description}
              </Text>
            </div>
          )}
        </div>
      )
    },
    {
      title: '权限范围',
      dataIndex: 'scopes',
      key: 'scopes',
      render: (scopes) => (
        <div>
          {scopes.slice(0, 2).map((scope: string) => (
            <Tag key={scope} size="small">{scope}</Tag>
          ))}
          {scopes.length > 2 && (
            <Tag size="small">+{scopes.length - 2}</Tag>
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
      title: '启用状态',
      dataIndex: 'isActive',
      key: 'isActive',
      render: (active, record) => (
        <Switch
          checked={active}
          onChange={(checked) => handleToggleApp(record.id, checked)}
        />
      )
    },
    {
      title: 'Token数量',
      dataIndex: 'tokenCount',
      key: 'tokenCount',
      render: (count) => (
        <Badge count={count} showZero color="blue" />
      )
    },
    {
      title: '最后使用',
      dataIndex: 'lastUsed',
      key: 'lastUsed',
      render: (date) => date ? dayjs(date).format('MM-DD HH:mm') : '未使用'
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="查看密钥">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedApp(record);
                setSecretModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="查看Tokens">
            <Button
              type="text"
              icon={<KeyOutlined />}
              onClick={() => {
                setSelectedApp(record);
                setTokensModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedApp(record);
                editForm.setFieldsValue({
                  ...record,
                  redirectUris: record.redirectUris.join('\n')
                });
                setEditModalVisible(true);
              }}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个OAuth应用吗？这将撤销所有相关的Token。"
            onConfirm={() => handleDeleteApp(record.id)}
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

  const tokenColumns: ColumnsType<OAuthToken> = [
    {
      title: '用户信息',
      key: 'user',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>
            {record.userInfo?.name || '未知用户'}
          </div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.userInfo?.email}
          </Text>
        </div>
      )
    },
    {
      title: '应用',
      key: 'app',
      render: (_, record) => {
        const app = apps.find(a => a.id === record.appId);
        return app ? app.name : '未知应用';
      }
    },
    {
      title: '权限范围',
      dataIndex: 'scopes',
      key: 'scopes',
      render: (scopes) => (
        <div>
          {scopes.map((scope: string) => (
            <Tag key={scope} size="small">{scope}</Tag>
          ))}
        </div>
      )
    },
    {
      title: '状态',
      key: 'status',
      render: (_, record) => {
        const isExpired = dayjs(record.expiresAt).isBefore(dayjs());
        if (record.isRevoked) {
          return <Tag color="red">已撤销</Tag>;
        } else if (isExpired) {
          return <Tag color="orange">已过期</Tag>;
        } else {
          return <Tag color="green">有效</Tag>;
        }
      }
    },
    {
      title: '过期时间',
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      render: (date) => dayjs(date).format('YYYY-MM-DD HH:mm')
    },
    {
      title: '最后使用',
      dataIndex: 'lastUsed',
      key: 'lastUsed',
      render: (date) => date ? dayjs(date).format('MM-DD HH:mm') : '未使用'
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {!record.isRevoked && (
            <Popconfirm
              title="确定要撤销这个Token吗？"
              onConfirm={() => handleRevokeToken(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="text" danger size="small">
                撤销
              </Button>
            </Popconfirm>
          )}
        </Space>
      )
    }
  ];

  const activeApps = apps.filter(app => app.isActive).length;
  const totalTokens = tokens.filter(token => !token.isRevoked).length;
  const expiredTokens = tokens.filter(token => 
    !token.isRevoked && dayjs(token.expiresAt).isBefore(dayjs())
  ).length;

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃应用"
              value={activeApps}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="有效Tokens"
              value={totalTokens}
              prefix={<KeyOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="过期Tokens"
              value={expiredTokens}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总授权数"
              value={apps.reduce((sum, app) => sum + app.tokenCount, 0)}
              prefix={<SafetyOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <SafetyOutlined />
            OAuth管理
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建OAuth应用
          </Button>
        }
      >
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="OAuth应用" key="apps">
            <Table
              columns={appColumns}
              dataSource={apps}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 个应用`
              }}
            />
          </TabPane>
          
          <TabPane tab="访问令牌" key="tokens">
            <Table
              columns={tokenColumns}
              dataSource={tokens}
              rowKey="id"
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 个Token`
              }}
            />
          </TabPane>
        </Tabs>
      </Card>

      {/* 创建OAuth应用模态框 */}
      <Modal
        title="创建OAuth应用"
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
          onFinish={handleCreateApp}
        >
          <Form.Item
            name="name"
            label="应用名称"
            rules={[{ required: true, message: '请输入应用名称' }]}
          >
            <Input placeholder="输入OAuth应用名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="应用描述"
          >
            <TextArea
              placeholder="输入应用描述"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="website"
            label="应用网站"
          >
            <Input placeholder="https://example.com" />
          </Form.Item>

          <Form.Item
            name="redirectUris"
            label="重定向URI"
            rules={[{ required: true, message: '请输入至少一个重定向URI' }]}
          >
            <TextArea
              placeholder="每行一个URI，例如：&#10;https://example.com/callback&#10;myapp://oauth/callback"
              rows={4}
            />
          </Form.Item>

          <Form.Item
            name="scopes"
            label="权限范围"
            rules={[{ required: true, message: '请选择至少一个权限范围' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择权限范围"
              options={availableScopes.map(scope => ({ label: scope, value: scope }))}
            />
          </Form.Item>

          <Form.Item
            name="grantTypes"
            label="授权类型"
            rules={[{ required: true, message: '请选择至少一个授权类型' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择授权类型"
              options={availableGrantTypes.map(type => ({ label: type, value: type }))}
            />
          </Form.Item>

          <Form.Item
            name="isActive"
            valuePropName="checked"
            initialValue={true}
          >
            <Switch checkedChildren="启用" unCheckedChildren="禁用" />
            <span style={{ marginLeft: 8 }}>创建后立即启用</span>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建应用
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

      {/* 编辑OAuth应用模态框 */}
      <Modal
        title="编辑OAuth应用"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setSelectedApp(null);
          editForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdateApp}
        >
          <Form.Item
            name="name"
            label="应用名称"
            rules={[{ required: true, message: '请输入应用名称' }]}
          >
            <Input placeholder="输入OAuth应用名称" />
          </Form.Item>

          <Form.Item
            name="description"
            label="应用描述"
          >
            <TextArea
              placeholder="输入应用描述"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            name="website"
            label="应用网站"
          >
            <Input placeholder="https://example.com" />
          </Form.Item>

          <Form.Item
            name="redirectUris"
            label="重定向URI"
            rules={[{ required: true, message: '请输入至少一个重定向URI' }]}
          >
            <TextArea
              placeholder="每行一个URI"
              rows={4}
            />
          </Form.Item>

          <Form.Item
            name="scopes"
            label="权限范围"
            rules={[{ required: true, message: '请选择至少一个权限范围' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择权限范围"
              options={availableScopes.map(scope => ({ label: scope, value: scope }))}
            />
          </Form.Item>

          <Form.Item
            name="grantTypes"
            label="授权类型"
            rules={[{ required: true, message: '请选择至少一个授权类型' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择授权类型"
              options={availableGrantTypes.map(type => ({ label: type, value: type }))}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                更新应用
              </Button>
              <Button onClick={() => {
                setEditModalVisible(false);
                setSelectedApp(null);
                editForm.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 查看密钥模态框 */}
      <Modal
        title={`应用密钥 - ${selectedApp?.name}`}
        open={secretModalVisible}
        onCancel={() => {
          setSecretModalVisible(false);
          setSelectedApp(null);
        }}
        footer={
          <Button onClick={() => {
            setSecretModalVisible(false);
            setSelectedApp(null);
          }}>
            关闭
          </Button>
        }
        width={600}
      >
        {selectedApp && (
          <div>
            <Alert
              message="安全提醒"
              description="请妥善保管Client Secret，不要在客户端代码中暴露。"
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
            
            <Descriptions column={1} bordered>
              <Descriptions.Item 
                label="Client ID"
                extra={
                  <Button
                    type="text"
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(selectedApp.clientId)}
                  />
                }
              >
                <Text code>{selectedApp.clientId}</Text>
              </Descriptions.Item>
              
              <Descriptions.Item 
                label="Client Secret"
                extra={
                  <Space>
                    <Button
                      type="text"
                      icon={<CopyOutlined />}
                      onClick={() => copyToClipboard(selectedApp.clientSecret)}
                    />
                    <Popconfirm
                      title="确定要重新生成Client Secret吗？这将使现有的Secret失效。"
                      onConfirm={() => handleRegenerateSecret(selectedApp.id)}
                      okText="确定"
                      cancelText="取消"
                    >
                      <Button
                        type="text"
                        icon={<ReloadOutlined />}
                        danger
                      />
                    </Popconfirm>
                  </Space>
                }
              >
                <Text code>{selectedApp.clientSecret}</Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="重定向URI">
                {selectedApp.redirectUris.map((uri, index) => (
                  <div key={index}>
                    <Text code>{uri}</Text>
                    <Button
                      type="text"
                      size="small"
                      icon={<CopyOutlined />}
                      onClick={() => copyToClipboard(uri)}
                    />
                  </div>
                ))}
              </Descriptions.Item>
            </Descriptions>

            <Divider />
            
            <Title level={5}>授权URL示例</Title>
            <Text code style={{ wordBreak: 'break-all' }}>
              {`https://api.example.com/oauth/authorize?client_id=${selectedApp.clientId}&redirect_uri=${encodeURIComponent(selectedApp.redirectUris[0] || '')}&response_type=code&scope=${selectedApp.scopes.join('%20')}`}
            </Text>
            <Button
              type="text"
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(`https://api.example.com/oauth/authorize?client_id=${selectedApp.clientId}&redirect_uri=${encodeURIComponent(selectedApp.redirectUris[0] || '')}&response_type=code&scope=${selectedApp.scopes.join('%20')}`)}
              style={{ marginLeft: 8 }}
            />
          </div>
        )}
      </Modal>

      {/* 查看Tokens模态框 */}
      <Modal
        title={`访问令牌 - ${selectedApp?.name}`}
        open={tokensModalVisible}
        onCancel={() => {
          setTokensModalVisible(false);
          setSelectedApp(null);
        }}
        footer={
          <Button onClick={() => {
            setTokensModalVisible(false);
            setSelectedApp(null);
          }}>
            关闭
          </Button>
        }
        width={800}
      >
        <Table
          columns={tokenColumns}
          dataSource={tokens.filter(token => token.appId === selectedApp?.id)}
          rowKey="id"
          size="small"
          pagination={{
            pageSize: 5,
            showSizeChanger: false
          }}
        />
      </Modal>
    </div>
  );
};

export default OAuthManagement;