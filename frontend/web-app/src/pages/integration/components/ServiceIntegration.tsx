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
  Avatar,
  Badge,
  Progress,
  Divider,
  Steps,
  Result
} from 'antd';
import {
  PlusOutlined,
  ApiOutlined,
  EditOutlined,
  DeleteOutlined,
  SyncOutlined,
  SettingOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  DisconnectOutlined,
  LinkOutlined,
  ExperimentOutlined,
  DatabaseOutlined,
  CloudOutlined,
  MailOutlined,
  MessageOutlined,
  FileTextOutlined,
  ShoppingCartOutlined,
  DollarOutlined,
  BarChartOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;
const { Step } = Steps;

interface Integration {
  id: number;
  name: string;
  type: 'api' | 'webhook' | 'oauth' | 'database' | 'file';
  service: string;
  status: 'connected' | 'disconnected' | 'error' | 'testing';
  isEnabled: boolean;
  config: Record<string, any>;
  settings: Record<string, any>;
  lastSync?: string;
  syncStatus?: 'success' | 'failed' | 'pending';
  createdAt: string;
  updatedAt: string;
  errorMessage?: string;
  syncCount?: number;
  dataVolume?: string;
}

interface ServiceTemplate {
  id: string;
  name: string;
  type: string;
  category: string;
  description: string;
  icon: React.ReactNode;
  popular: boolean;
  configFields: Array<{
    name: string;
    label: string;
    type: 'input' | 'password' | 'select' | 'textarea';
    required: boolean;
    options?: string[];
    placeholder?: string;
  }>;
}

const ServiceIntegration: React.FC = () => {
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [serviceTemplates, setServiceTemplates] = useState<ServiceTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [testModalVisible, setTestModalVisible] = useState(false);
  const [selectedIntegration, setSelectedIntegration] = useState<Integration | null>(null);
  const [selectedTemplate, setSelectedTemplate] = useState<ServiceTemplate | null>(null);
  const [currentStep, setCurrentStep] = useState(0);
  const [testResult, setTestResult] = useState<any>(null);
  const [form] = Form.useForm();
  const [configForm] = Form.useForm();

  useEffect(() => {
    fetchIntegrations();
    fetchServiceTemplates();
  }, []);

  const fetchIntegrations = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockData: Integration[] = [
        {
          id: 1,
          name: 'Slack通知',
          type: 'webhook',
          service: 'Slack',
          status: 'connected',
          isEnabled: true,
          config: {
            webhookUrl: 'https://hooks.slack.com/services/***',
            channel: '#general',
            username: 'AI助手'
          },
          settings: {
            notifyOnError: true,
            notifyOnSuccess: false
          },
          lastSync: '2024-01-16T10:30:00Z',
          syncStatus: 'success',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-16T10:30:00Z',
          syncCount: 1250,
          dataVolume: '2.5MB'
        },
        {
          id: 2,
          name: 'Google Drive备份',
          type: 'oauth',
          service: 'Google Drive',
          status: 'connected',
          isEnabled: true,
          config: {
            clientId: '***',
            clientSecret: '***',
            folderId: '1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms'
          },
          settings: {
            autoBackup: true,
            backupInterval: 'daily',
            retentionDays: 30
          },
          lastSync: '2024-01-16T08:00:00Z',
          syncStatus: 'success',
          createdAt: '2024-01-05T00:00:00Z',
          updatedAt: '2024-01-16T08:00:00Z',
          syncCount: 45,
          dataVolume: '128MB'
        },
        {
          id: 3,
          name: 'MySQL数据库',
          type: 'database',
          service: 'MySQL',
          status: 'error',
          isEnabled: false,
          config: {
            host: 'localhost',
            port: 3306,
            database: 'production',
            username: 'admin'
          },
          settings: {
            connectionTimeout: 30,
            maxConnections: 10
          },
          lastSync: '2024-01-15T20:15:00Z',
          syncStatus: 'failed',
          createdAt: '2024-01-10T00:00:00Z',
          updatedAt: '2024-01-15T20:15:00Z',
          errorMessage: '连接超时：无法连接到数据库服务器',
          syncCount: 0,
          dataVolume: '0B'
        },
        {
          id: 4,
          name: 'Stripe支付',
          type: 'api',
          service: 'Stripe',
          status: 'testing',
          isEnabled: false,
          config: {
            apiKey: 'sk_test_***',
            webhookSecret: 'whsec_***'
          },
          settings: {
            currency: 'USD',
            testMode: true
          },
          createdAt: '2024-01-16T12:00:00Z',
          updatedAt: '2024-01-16T12:00:00Z',
          syncCount: 0
        }
      ];
      setIntegrations(mockData);
    } catch (error) {
      message.error('获取集成列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchServiceTemplates = async () => {
    try {
      // 模拟服务模板数据
      const mockTemplates: ServiceTemplate[] = [
        {
          id: 'slack',
          name: 'Slack',
          type: 'webhook',
          category: 'Communication',
          description: '发送消息到Slack频道',
          icon: <MessageOutlined />,
          popular: true,
          configFields: [
            { name: 'webhookUrl', label: 'Webhook URL', type: 'input', required: true, placeholder: 'https://hooks.slack.com/services/...' },
            { name: 'channel', label: '频道', type: 'input', required: true, placeholder: '#general' },
            { name: 'username', label: '用户名', type: 'input', required: false, placeholder: 'Bot' }
          ]
        },
        {
          id: 'google-drive',
          name: 'Google Drive',
          type: 'oauth',
          category: 'Storage',
          description: '同步文件到Google Drive',
          icon: <CloudOutlined />,
          popular: true,
          configFields: [
            { name: 'clientId', label: 'Client ID', type: 'input', required: true },
            { name: 'clientSecret', label: 'Client Secret', type: 'password', required: true },
            { name: 'folderId', label: '文件夹ID', type: 'input', required: false }
          ]
        },
        {
          id: 'mysql',
          name: 'MySQL',
          type: 'database',
          category: 'Database',
          description: '连接MySQL数据库',
          icon: <DatabaseOutlined />,
          popular: false,
          configFields: [
            { name: 'host', label: '主机', type: 'input', required: true, placeholder: 'localhost' },
            { name: 'port', label: '端口', type: 'input', required: true, placeholder: '3306' },
            { name: 'database', label: '数据库', type: 'input', required: true },
            { name: 'username', label: '用户名', type: 'input', required: true },
            { name: 'password', label: '密码', type: 'password', required: true }
          ]
        },
        {
          id: 'stripe',
          name: 'Stripe',
          type: 'api',
          category: 'Payment',
          description: '处理在线支付',
          icon: <DollarOutlined />,
          popular: true,
          configFields: [
            { name: 'apiKey', label: 'API Key', type: 'password', required: true },
            { name: 'webhookSecret', label: 'Webhook Secret', type: 'password', required: false }
          ]
        },
        {
          id: 'sendgrid',
          name: 'SendGrid',
          type: 'api',
          category: 'Email',
          description: '发送邮件通知',
          icon: <MailOutlined />,
          popular: false,
          configFields: [
            { name: 'apiKey', label: 'API Key', type: 'password', required: true },
            { name: 'fromEmail', label: '发件人邮箱', type: 'input', required: true },
            { name: 'fromName', label: '发件人姓名', type: 'input', required: false }
          ]
        },
        {
          id: 'analytics',
          name: 'Google Analytics',
          type: 'oauth',
          category: 'Analytics',
          description: '获取网站分析数据',
          icon: <BarChartOutlined />,
          popular: false,
          configFields: [
            { name: 'clientId', label: 'Client ID', type: 'input', required: true },
            { name: 'clientSecret', label: 'Client Secret', type: 'password', required: true },
            { name: 'viewId', label: 'View ID', type: 'input', required: true }
          ]
        }
      ];
      setServiceTemplates(mockTemplates);
    } catch (error) {
      message.error('获取服务模板失败');
    }
  };

  const handleCreateIntegration = async (values: any) => {
    if (!selectedTemplate) return;

    try {
      const newIntegration: Integration = {
        id: Date.now(),
        name: values.name || selectedTemplate.name,
        type: selectedTemplate.type as any,
        service: selectedTemplate.name,
        status: 'testing',
        isEnabled: false,
        config: values,
        settings: {},
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        syncCount: 0
      };

      setIntegrations([newIntegration, ...integrations]);
      setCreateModalVisible(false);
      setCurrentStep(0);
      setSelectedTemplate(null);
      form.resetFields();
      
      // 模拟测试连接
      setTimeout(() => {
        setIntegrations(prev => prev.map(i => 
          i.id === newIntegration.id 
            ? { ...i, status: 'connected' as const }
            : i
        ));
        message.success('集成创建成功');
      }, 2000);
      
      message.info('正在测试连接...');
    } catch (error) {
      message.error('创建集成失败');
    }
  };

  const handleTestConnection = async (integration: Integration) => {
    setSelectedIntegration(integration);
    setTestModalVisible(true);
    
    try {
      // 模拟测试连接
      setTestResult({ status: 'testing' });
      
      setTimeout(() => {
        const success = Math.random() > 0.3; // 70% 成功率
        if (success) {
          setTestResult({
            status: 'success',
            message: '连接测试成功',
            details: {
              responseTime: '120ms',
              statusCode: 200,
              timestamp: new Date().toISOString()
            }
          });
          
          // 更新集成状态
          setIntegrations(prev => prev.map(i => 
            i.id === integration.id 
              ? { ...i, status: 'connected' as const, errorMessage: undefined }
              : i
          ));
        } else {
          setTestResult({
            status: 'error',
            message: '连接测试失败',
            error: '认证失败：API密钥无效'
          });
          
          // 更新集成状态
          setIntegrations(prev => prev.map(i => 
            i.id === integration.id 
              ? { ...i, status: 'error' as const, errorMessage: '认证失败：API密钥无效' }
              : i
          ));
        }
      }, 2000);
    } catch (error) {
      setTestResult({
        status: 'error',
        message: '测试连接时发生错误',
        error: error instanceof Error ? error.message : '未知错误'
      });
    }
  };

  const handleToggleIntegration = async (id: number, enabled: boolean) => {
    try {
      setIntegrations(integrations.map(integration =>
        integration.id === id
          ? { ...integration, isEnabled: enabled, updatedAt: new Date().toISOString() }
          : integration
      ));
      message.success(enabled ? '集成已启用' : '集成已禁用');
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleDeleteIntegration = async (id: number) => {
    try {
      setIntegrations(integrations.filter(integration => integration.id !== id));
      message.success('集成删除成功');
    } catch (error) {
      message.error('删除集成失败');
    }
  };

  const handleSyncNow = async (id: number) => {
    try {
      setIntegrations(integrations.map(integration =>
        integration.id === id
          ? { 
              ...integration, 
              syncStatus: 'pending' as const,
              lastSync: new Date().toISOString(),
              updatedAt: new Date().toISOString()
            }
          : integration
      ));
      
      // 模拟同步过程
      setTimeout(() => {
        setIntegrations(prev => prev.map(i => 
          i.id === id 
            ? { 
                ...i, 
                syncStatus: 'success' as const,
                syncCount: (i.syncCount || 0) + 1
              }
            : i
        ));
        message.success('同步完成');
      }, 3000);
      
      message.info('开始同步...');
    } catch (error) {
      message.error('同步失败');
    }
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      connected: { color: 'green', text: '已连接', icon: <CheckCircleOutlined /> },
      disconnected: { color: 'default', text: '未连接', icon: <DisconnectOutlined /> },
      error: { color: 'red', text: '错误', icon: <ExclamationCircleOutlined /> },
      testing: { color: 'blue', text: '测试中', icon: <ClockCircleOutlined /> }
    };
    const config = statusConfig[status as keyof typeof statusConfig];
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  const getSyncStatusTag = (status?: string) => {
    if (!status) return null;
    
    const statusConfig = {
      success: { color: 'green', text: '成功' },
      failed: { color: 'red', text: '失败' },
      pending: { color: 'blue', text: '进行中' }
    };
    const config = statusConfig[status as keyof typeof statusConfig];
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getServiceIcon = (service: string) => {
    const iconMap: Record<string, React.ReactNode> = {
      'Slack': <MessageOutlined />,
      'Google Drive': <CloudOutlined />,
      'MySQL': <DatabaseOutlined />,
      'Stripe': <DollarOutlined />,
      'SendGrid': <MailOutlined />,
      'Google Analytics': <BarChartOutlined />
    };
    return iconMap[service] || <ApiOutlined />;
  };

  const columns: ColumnsType<Integration> = [
    {
      title: '服务信息',
      key: 'service',
      render: (_, record) => (
        <Space>
          <Avatar icon={getServiceIcon(record.service)} />
          <div>
            <div style={{ fontWeight: 500 }}>{record.name}</div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {record.service} • {record.type.toUpperCase()}
            </Text>
          </div>
        </Space>
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
      dataIndex: 'isEnabled',
      key: 'isEnabled',
      render: (enabled, record) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggleIntegration(record.id, checked)}
          disabled={record.status !== 'connected'}
        />
      )
    },
    {
      title: '最后同步',
      key: 'sync',
      render: (_, record) => (
        <div>
          {record.lastSync ? (
            <>
              <div>{dayjs(record.lastSync).format('MM-DD HH:mm')}</div>
              <div>{getSyncStatusTag(record.syncStatus)}</div>
            </>
          ) : (
            <Text type="secondary">未同步</Text>
          )}
        </div>
      )
    },
    {
      title: '数据量',
      key: 'data',
      render: (_, record) => (
        <div>
          <div>同步次数: {record.syncCount || 0}</div>
          {record.dataVolume && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              数据量: {record.dataVolume}
            </Text>
          )}
        </div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="测试连接">
            <Button
              type="text"
              icon={<ExperimentOutlined />}
              onClick={() => handleTestConnection(record)}
            />
          </Tooltip>
          {record.status === 'connected' && record.isEnabled && (
            <Tooltip title="立即同步">
              <Button
                type="text"
                icon={<SyncOutlined />}
                onClick={() => handleSyncNow(record.id)}
                loading={record.syncStatus === 'pending'}
              />
            </Tooltip>
          )}
          <Tooltip title="配置">
            <Button
              type="text"
              icon={<SettingOutlined />}
              onClick={() => {
                setSelectedIntegration(record);
                configForm.setFieldsValue(record.config);
                setConfigModalVisible(true);
              }}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个集成吗？"
            onConfirm={() => handleDeleteIntegration(record.id)}
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

  const connectedCount = integrations.filter(i => i.status === 'connected').length;
  const enabledCount = integrations.filter(i => i.isEnabled).length;
  const totalSyncs = integrations.reduce((sum, i) => sum + (i.syncCount || 0), 0);
  const errorCount = integrations.filter(i => i.status === 'error').length;

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="已连接服务"
              value={connectedCount}
              prefix={<LinkOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已启用集成"
              value={enabledCount}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总同步次数"
              value={totalSyncs}
              prefix={<SyncOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="错误数量"
              value={errorCount}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: errorCount > 0 ? '#cf1322' : undefined }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <ApiOutlined />
            第三方服务集成
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            添加集成
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={integrations}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个集成`
          }}
        />
      </Card>

      {/* 创建集成模态框 */}
      <Modal
        title="添加第三方服务集成"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          setCurrentStep(0);
          setSelectedTemplate(null);
          form.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Steps current={currentStep} style={{ marginBottom: 24 }}>
          <Step title="选择服务" />
          <Step title="配置参数" />
          <Step title="完成" />
        </Steps>

        {currentStep === 0 && (
          <div>
            <Title level={5}>选择要集成的服务</Title>
            <Tabs defaultActiveKey="popular">
              <TabPane tab="热门服务" key="popular">
                <Row gutter={[16, 16]}>
                  {serviceTemplates.filter(t => t.popular).map(template => (
                    <Col span={8} key={template.id}>
                      <Card
                        hoverable
                        onClick={() => {
                          setSelectedTemplate(template);
                          setCurrentStep(1);
                        }}
                        style={{ 
                          border: selectedTemplate?.id === template.id ? '2px solid #1890ff' : undefined 
                        }}
                      >
                        <Card.Meta
                          avatar={<Avatar icon={template.icon} />}
                          title={template.name}
                          description={template.description}
                        />
                        <div style={{ marginTop: 8 }}>
                          <Tag color="blue">{template.category}</Tag>
                          <Tag>{template.type.toUpperCase()}</Tag>
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              </TabPane>
              <TabPane tab="所有服务" key="all">
                <Row gutter={[16, 16]}>
                  {serviceTemplates.map(template => (
                    <Col span={8} key={template.id}>
                      <Card
                        hoverable
                        onClick={() => {
                          setSelectedTemplate(template);
                          setCurrentStep(1);
                        }}
                        style={{ 
                          border: selectedTemplate?.id === template.id ? '2px solid #1890ff' : undefined 
                        }}
                      >
                        <Card.Meta
                          avatar={<Avatar icon={template.icon} />}
                          title={template.name}
                          description={template.description}
                        />
                        <div style={{ marginTop: 8 }}>
                          <Tag color="blue">{template.category}</Tag>
                          <Tag>{template.type.toUpperCase()}</Tag>
                          {template.popular && <Tag color="gold">热门</Tag>}
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              </TabPane>
            </Tabs>
          </div>
        )}

        {currentStep === 1 && selectedTemplate && (
          <div>
            <Title level={5}>配置 {selectedTemplate.name}</Title>
            <Form
              form={form}
              layout="vertical"
              onFinish={handleCreateIntegration}
            >
              <Form.Item
                name="name"
                label="集成名称"
                initialValue={selectedTemplate.name}
              >
                <Input placeholder="输入集成名称" />
              </Form.Item>

              {selectedTemplate.configFields.map(field => (
                <Form.Item
                  key={field.name}
                  name={field.name}
                  label={field.label}
                  rules={field.required ? [{ required: true, message: `请输入${field.label}` }] : []}
                >
                  {field.type === 'password' ? (
                    <Input.Password placeholder={field.placeholder} />
                  ) : field.type === 'textarea' ? (
                    <TextArea placeholder={field.placeholder} rows={3} />
                  ) : field.type === 'select' ? (
                    <Select placeholder={field.placeholder}>
                      {field.options?.map(option => (
                        <Option key={option} value={option}>{option}</Option>
                      ))}
                    </Select>
                  ) : (
                    <Input placeholder={field.placeholder} />
                  )}
                </Form.Item>
              ))}

              <Form.Item>
                <Space>
                  <Button onClick={() => setCurrentStep(0)}>
                    上一步
                  </Button>
                  <Button type="primary" htmlType="submit">
                    创建集成
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </div>
        )}
      </Modal>

      {/* 配置模态框 */}
      <Modal
        title={`配置 ${selectedIntegration?.name}`}
        open={configModalVisible}
        onCancel={() => {
          setConfigModalVisible(false);
          setSelectedIntegration(null);
          configForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={configForm}
          layout="vertical"
          onFinish={(values) => {
            if (!selectedIntegration) return;
            
            setIntegrations(integrations.map(integration =>
              integration.id === selectedIntegration.id
                ? { ...integration, config: values, updatedAt: new Date().toISOString() }
                : integration
            ));
            setConfigModalVisible(false);
            setSelectedIntegration(null);
            configForm.resetFields();
            message.success('配置更新成功');
          }}
        >
          {selectedIntegration && Object.keys(selectedIntegration.config).map(key => (
            <Form.Item
              key={key}
              name={key}
              label={key}
            >
              {key.toLowerCase().includes('password') || key.toLowerCase().includes('secret') || key.toLowerCase().includes('key') ? (
                <Input.Password placeholder={`输入 ${key}`} />
              ) : (
                <Input placeholder={`输入 ${key}`} />
              )}
            </Form.Item>
          ))}

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                保存配置
              </Button>
              <Button onClick={() => {
                setConfigModalVisible(false);
                setSelectedIntegration(null);
                configForm.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 测试连接模态框 */}
      <Modal
        title={`测试连接 - ${selectedIntegration?.name}`}
        open={testModalVisible}
        onCancel={() => {
          setTestModalVisible(false);
          setSelectedIntegration(null);
          setTestResult(null);
        }}
        footer={
          <Button onClick={() => {
            setTestModalVisible(false);
            setSelectedIntegration(null);
            setTestResult(null);
          }}>
            关闭
          </Button>
        }
        width={500}
      >
        {testResult ? (
          <div>
            {testResult.status === 'testing' && (
              <div style={{ textAlign: 'center', padding: '40px 0' }}>
                <Progress type="circle" percent={60} />
                <div style={{ marginTop: 16 }}>正在测试连接...</div>
              </div>
            )}
            
            {testResult.status === 'success' && (
              <Result
                status="success"
                title="连接测试成功"
                subTitle={testResult.message}
                extra={
                  testResult.details && (
                    <div style={{ textAlign: 'left' }}>
                      <Divider />
                      <div><strong>响应时间:</strong> {testResult.details.responseTime}</div>
                      <div><strong>状态码:</strong> {testResult.details.statusCode}</div>
                      <div><strong>测试时间:</strong> {dayjs(testResult.details.timestamp).format('YYYY-MM-DD HH:mm:ss')}</div>
                    </div>
                  )
                }
              />
            )}
            
            {testResult.status === 'error' && (
              <Result
                status="error"
                title="连接测试失败"
                subTitle={testResult.message}
                extra={
                  testResult.error && (
                    <Alert
                      message="错误详情"
                      description={testResult.error}
                      type="error"
                      showIcon
                    />
                  )
                }
              />
            )}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Button
              type="primary"
              icon={<ExperimentOutlined />}
              onClick={() => selectedIntegration && handleTestConnection(selectedIntegration)}
            >
              开始测试
            </Button>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ServiceIntegration;