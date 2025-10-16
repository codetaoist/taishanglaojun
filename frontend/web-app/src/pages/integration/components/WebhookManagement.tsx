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
  Progress,
  Divider,
  Timeline,
  Drawer,
  Descriptions
} from 'antd';
import {
  PlusOutlined,
  SendOutlined,
  EditOutlined,
  DeleteOutlined,
  HistoryOutlined,
  SettingOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  LinkOutlined,
  ExperimentOutlined,
  ReloadOutlined,
  EyeOutlined,
  CopyOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  BugOutlined,
  GlobalOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

interface Webhook {
  id: number;
  name: string;
  url: string;
  method: 'POST' | 'PUT' | 'PATCH';
  events: string[];
  isActive: boolean;
  secret?: string;
  headers: Record<string, string>;
  retryCount: number;
  timeout: number;
  lastTriggered?: string;
  status: 'active' | 'inactive' | 'error' | 'testing';
  successCount: number;
  failureCount: number;
  createdAt: string;
  updatedAt: string;
  errorMessage?: string;
}

interface WebhookLog {
  id: number;
  webhookId: number;
  event: string;
  status: 'success' | 'failed' | 'pending' | 'retrying';
  statusCode?: number;
  responseTime?: number;
  payload: any;
  response?: any;
  error?: string;
  attempt: number;
  maxAttempts: number;
  timestamp: string;
}

const WebhookManagement: React.FC = () => {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [webhookLogs, setWebhookLogs] = useState<WebhookLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [logsDrawerVisible, setLogsDrawerVisible] = useState(false);
  const [testModalVisible, setTestModalVisible] = useState(false);
  const [selectedWebhook, setSelectedWebhook] = useState<Webhook | null>(null);
  const [selectedLog, setSelectedLog] = useState<WebhookLog | null>(null);
  const [testResult, setTestResult] = useState<any>(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const availableEvents = [
    'user.created',
    'user.updated',
    'user.deleted',
    'order.created',
    'order.updated',
    'order.completed',
    'payment.success',
    'payment.failed',
    'system.error',
    'data.sync'
  ];

  useEffect(() => {
    fetchWebhooks();
    fetchWebhookLogs();
  }, []);

  const fetchWebhooks = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockData: Webhook[] = [
        {
          id: 1,
          name: 'Slack通知',
          url: 'https://example.com/webhook',
          method: 'POST',
          events: ['user.created', 'order.completed', 'system.error'],
          isActive: true,
          secret: 'webhook_secret_123',
          headers: {
            'Content-Type': 'application/json',
            'User-Agent': 'TaiShangLaoJun-Webhook/1.0'
          },
          retryCount: 3,
          timeout: 30,
          lastTriggered: '2024-01-16T10:30:00Z',
          status: 'active',
          successCount: 1250,
          failureCount: 15,
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-16T10:30:00Z'
        },
        {
          id: 2,
          name: '订单处理系统',
          url: 'https://api.example.com/webhooks/orders',
          method: 'POST',
          events: ['order.created', 'order.updated', 'payment.success'],
          isActive: true,
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer token123'
          },
          retryCount: 5,
          timeout: 60,
          lastTriggered: '2024-01-16T09:15:00Z',
          status: 'active',
          successCount: 890,
          failureCount: 8,
          createdAt: '2024-01-05T00:00:00Z',
          updatedAt: '2024-01-16T09:15:00Z'
        },
        {
          id: 3,
          name: '数据同步服务',
          url: 'https://sync.example.com/webhook',
          method: 'PUT',
          events: ['data.sync'],
          isActive: false,
          headers: {
            'Content-Type': 'application/json'
          },
          retryCount: 2,
          timeout: 45,
          lastTriggered: '2024-01-15T20:00:00Z',
          status: 'error',
          successCount: 156,
          failureCount: 45,
          createdAt: '2024-01-10T00:00:00Z',
          updatedAt: '2024-01-15T20:00:00Z',
          errorMessage: '连接超时：目标服务器无响应'
        }
      ];
      setWebhooks(mockData);
    } catch (error) {
      message.error('获取Webhook列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchWebhookLogs = async () => {
    try {
      // 模拟日志数据
      const mockLogs: WebhookLog[] = [
        {
          id: 1,
          webhookId: 1,
          event: 'user.created',
          status: 'success',
          statusCode: 200,
          responseTime: 120,
          payload: {
            event: 'user.created',
            data: {
              id: 123,
              email: 'user@example.com',
              name: '张三'
            },
            timestamp: '2024-01-16T10:30:00Z'
          },
          response: {
            message: 'ok'
          },
          attempt: 1,
          maxAttempts: 3,
          timestamp: '2024-01-16T10:30:00Z'
        },
        {
          id: 2,
          webhookId: 1,
          event: 'order.completed',
          status: 'failed',
          statusCode: 500,
          responseTime: 5000,
          payload: {
            event: 'order.completed',
            data: {
              orderId: 'ORD-001',
              amount: 99.99,
              status: 'completed'
            },
            timestamp: '2024-01-16T10:25:00Z'
          },
          error: 'Internal Server Error',
          attempt: 3,
          maxAttempts: 3,
          timestamp: '2024-01-16T10:25:00Z'
        },
        {
          id: 3,
          webhookId: 2,
          event: 'payment.success',
          status: 'retrying',
          statusCode: 429,
          responseTime: 200,
          payload: {
            event: 'payment.success',
            data: {
              paymentId: 'PAY-001',
              amount: 199.99,
              currency: 'USD'
            },
            timestamp: '2024-01-16T10:20:00Z'
          },
          error: 'Rate limit exceeded',
          attempt: 2,
          maxAttempts: 5,
          timestamp: '2024-01-16T10:20:00Z'
        }
      ];
      setWebhookLogs(mockLogs);
    } catch (error) {
      message.error('获取Webhook日志失败');
    }
  };

  const handleCreateWebhook = async (values: any) => {
    try {
      const newWebhook: Webhook = {
        id: Date.now(),
        name: values.name,
        url: values.url,
        method: values.method,
        events: values.events,
        isActive: values.isActive !== false,
        secret: values.secret,
        headers: values.headers ? JSON.parse(values.headers) : {},
        retryCount: values.retryCount || 3,
        timeout: values.timeout || 30,
        status: 'inactive',
        successCount: 0,
        failureCount: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      setWebhooks([newWebhook, ...webhooks]);
      setCreateModalVisible(false);
      form.resetFields();
      message.success('Webhook创建成功');
    } catch (error) {
      message.error('创建Webhook失败');
    }
  };

  const handleUpdateWebhook = async (values: any) => {
    if (!selectedWebhook) return;

    try {
      setWebhooks(webhooks.map(webhook =>
        webhook.id === selectedWebhook.id
          ? {
              ...webhook,
              ...values,
              headers: values.headers ? JSON.parse(values.headers) : webhook.headers,
              updatedAt: new Date().toISOString()
            }
          : webhook
      ));
      setEditModalVisible(false);
      setSelectedWebhook(null);
      editForm.resetFields();
      message.success('Webhook更新成功');
    } catch (error) {
      message.error('更新Webhook失败');
    }
  };

  const handleToggleWebhook = async (id: number, active: boolean) => {
    try {
      setWebhooks(webhooks.map(webhook =>
        webhook.id === id
          ? { 
              ...webhook, 
              isActive: active, 
              status: active ? 'active' as const : 'inactive' as const,
              updatedAt: new Date().toISOString() 
            }
          : webhook
      ));
      message.success(active ? 'Webhook已启用' : 'Webhook已禁用');
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleDeleteWebhook = async (id: number) => {
    try {
      setWebhooks(webhooks.filter(webhook => webhook.id !== id));
      message.success('Webhook删除成功');
    } catch (error) {
      message.error('删除Webhook失败');
    }
  };

  const handleTestWebhook = async (webhook: Webhook) => {
    setSelectedWebhook(webhook);
    setTestModalVisible(true);
    setTestResult({ status: 'testing' });

    try {
      // 模拟测试请求
      setTimeout(() => {
        const success = Math.random() > 0.2; // 80% 成功率
        if (success) {
          setTestResult({
            status: 'success',
            statusCode: 200,
            responseTime: 150,
            response: { message: 'Test webhook received successfully' },
            timestamp: new Date().toISOString()
          });
        } else {
          setTestResult({
            status: 'error',
            statusCode: 500,
            responseTime: 5000,
            error: 'Internal Server Error',
            timestamp: new Date().toISOString()
          });
        }
      }, 2000);
    } catch (error) {
      setTestResult({
        status: 'error',
        error: error instanceof Error ? error.message : '未知错误',
        timestamp: new Date().toISOString()
      });
    }
  };

  const handleRetryWebhook = async (logId: number) => {
    try {
      setWebhookLogs(webhookLogs.map(log =>
        log.id === logId
          ? { ...log, status: 'retrying' as const, attempt: log.attempt + 1 }
          : log
      ));
      
      // 模拟重试过程
      setTimeout(() => {
        const success = Math.random() > 0.3; // 70% 成功率
        setWebhookLogs(prev => prev.map(log =>
          log.id === logId
            ? { 
                ...log, 
                status: success ? 'success' as const : 'failed' as const,
                statusCode: success ? 200 : 500,
                responseTime: success ? 120 : 5000,
                error: success ? undefined : 'Retry failed'
              }
            : log
        ));
        message.success(success ? '重试成功' : '重试失败');
      }, 2000);
      
      message.info('正在重试...');
    } catch (error) {
      message.error('重试失败');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('已复制到剪贴板');
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      active: { color: 'green', text: '活跃', icon: <CheckCircleOutlined /> },
      inactive: { color: 'default', text: '未激活', icon: <PauseCircleOutlined /> },
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

  const getLogStatusTag = (status: string) => {
    const statusConfig = {
      success: { color: 'green', text: '成功' },
      failed: { color: 'red', text: '失败' },
      pending: { color: 'blue', text: '等待中' },
      retrying: { color: 'orange', text: '重试中' }
    };
    const config = statusConfig[status as keyof typeof statusConfig];
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns: ColumnsType<Webhook> = [
    {
      title: 'Webhook信息',
      key: 'info',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 500, marginBottom: 4 }}>
            {record.name}
          </div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.url}
          </Text>
          <div style={{ marginTop: 4 }}>
            <Tag color="blue">{record.method}</Tag>
            <Text type="secondary" style={{ fontSize: 11 }}>
              {record.events.length} 个事件
            </Text>
          </div>
        </div>
      )
    },
    {
      title: '事件',
      dataIndex: 'events',
      key: 'events',
      render: (events) => (
        <div>
          {events.slice(0, 2).map((event: string) => (
            <Tag key={event} size="small">{event}</Tag>
          ))}
          {events.length > 2 && (
            <Tag size="small">+{events.length - 2}</Tag>
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
          onChange={(checked) => handleToggleWebhook(record.id, checked)}
        />
      )
    },
    {
      title: '成功率',
      key: 'successRate',
      render: (_, record) => {
        const total = record.successCount + record.failureCount;
        const rate = total > 0 ? (record.successCount / total * 100).toFixed(1) : '0';
        return (
          <div>
            <div>{rate}%</div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {record.successCount}/{total}
            </Text>
          </div>
        );
      }
    },
    {
      title: '最后触发',
      dataIndex: 'lastTriggered',
      key: 'lastTriggered',
      render: (date) => date ? dayjs(date).format('MM-DD HH:mm') : '未触发'
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="测试">
            <Button
              type="text"
              icon={<ExperimentOutlined />}
              onClick={() => handleTestWebhook(record)}
            />
          </Tooltip>
          <Tooltip title="查看日志">
            <Button
              type="text"
              icon={<HistoryOutlined />}
              onClick={() => {
                setSelectedWebhook(record);
                setLogsDrawerVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedWebhook(record);
                editForm.setFieldsValue({
                  ...record,
                  headers: JSON.stringify(record.headers, null, 2)
                });
                setEditModalVisible(true);
              }}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个Webhook吗？"
            onConfirm={() => handleDeleteWebhook(record.id)}
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

  const activeWebhooks = webhooks.filter(w => w.isActive).length;
  const totalRequests = webhooks.reduce((sum, w) => sum + w.successCount + w.failureCount, 0);
  const totalSuccess = webhooks.reduce((sum, w) => sum + w.successCount, 0);
  const successRate = totalRequests > 0 ? (totalSuccess / totalRequests * 100).toFixed(1) : '0';

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃Webhooks"
              value={activeWebhooks}
              prefix={<SendOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总请求数"
              value={totalRequests}
              prefix={<GlobalOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功率"
              value={successRate}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="错误数量"
              value={webhooks.filter(w => w.status === 'error').length}
              prefix={<BugOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <SendOutlined />
            Webhook管理
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建Webhook
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={webhooks}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个Webhook`
          }}
        />
      </Card>

      {/* 创建Webhook模态框 */}
      <Modal
        title="创建Webhook"
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
          onFinish={handleCreateWebhook}
        >
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入Webhook名称' }]}
          >
            <Input placeholder="输入Webhook名称" />
          </Form.Item>

          <Form.Item
            name="url"
            label="URL"
            rules={[
              { required: true, message: '请输入Webhook URL' },
              { type: 'url', message: '请输入有效的URL' }
            ]}
          >
            <Input placeholder="https://example.com/webhook" />
          </Form.Item>

          <Form.Item
            name="method"
            label="HTTP方法"
            initialValue="POST"
            rules={[{ required: true, message: '请选择HTTP方法' }]}
          >
            <Select>
              <Option value="POST">POST</Option>
              <Option value="PUT">PUT</Option>
              <Option value="PATCH">PATCH</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="events"
            label="监听事件"
            rules={[{ required: true, message: '请选择至少一个事件' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择要监听的事件"
              options={availableEvents.map(event => ({ label: event, value: event }))}
            />
          </Form.Item>

          <Form.Item
            name="secret"
            label="签名密钥"
          >
            <Input.Password placeholder="用于验证请求签名（可选）" />
          </Form.Item>

          <Form.Item
            name="headers"
            label="自定义请求头"
          >
            <TextArea
              placeholder='{"Content-Type": "application/json"}'
              rows={3}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="retryCount"
                label="重试次数"
                initialValue={3}
              >
                <Input type="number" min={0} max={10} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="timeout"
                label="超时时间(秒)"
                initialValue={30}
              >
                <Input type="number" min={1} max={300} />
              </Form.Item>
            </Col>
          </Row>

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
                创建Webhook
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

      {/* 编辑Webhook模态框 */}
      <Modal
        title="编辑Webhook"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setSelectedWebhook(null);
          editForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdateWebhook}
        >
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入Webhook名称' }]}
          >
            <Input placeholder="输入Webhook名称" />
          </Form.Item>

          <Form.Item
            name="url"
            label="URL"
            rules={[
              { required: true, message: '请输入Webhook URL' },
              { type: 'url', message: '请输入有效的URL' }
            ]}
          >
            <Input placeholder="https://example.com/webhook" />
          </Form.Item>

          <Form.Item
            name="method"
            label="HTTP方法"
            rules={[{ required: true, message: '请选择HTTP方法' }]}
          >
            <Select>
              <Option value="POST">POST</Option>
              <Option value="PUT">PUT</Option>
              <Option value="PATCH">PATCH</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="events"
            label="监听事件"
            rules={[{ required: true, message: '请选择至少一个事件' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择要监听的事件"
              options={availableEvents.map(event => ({ label: event, value: event }))}
            />
          </Form.Item>

          <Form.Item
            name="secret"
            label="签名密钥"
          >
            <Input.Password placeholder="用于验证请求签名（可选）" />
          </Form.Item>

          <Form.Item
            name="headers"
            label="自定义请求头"
          >
            <TextArea
              placeholder='{"Content-Type": "application/json"}'
              rows={3}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="retryCount"
                label="重试次数"
              >
                <Input type="number" min={0} max={10} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="timeout"
                label="超时时间(秒)"
              >
                <Input type="number" min={1} max={300} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                更新Webhook
              </Button>
              <Button onClick={() => {
                setEditModalVisible(false);
                setSelectedWebhook(null);
                editForm.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 测试Webhook模态框 */}
      <Modal
        title={`测试Webhook - ${selectedWebhook?.name}`}
        open={testModalVisible}
        onCancel={() => {
          setTestModalVisible(false);
          setSelectedWebhook(null);
          setTestResult(null);
        }}
        footer={
          <Button onClick={() => {
            setTestModalVisible(false);
            setSelectedWebhook(null);
            setTestResult(null);
          }}>
            关闭
          </Button>
        }
        width={600}
      >
        {testResult ? (
          <div>
            {testResult.status === 'testing' && (
              <div style={{ textAlign: 'center', padding: '40px 0' }}>
                <Progress type="circle" percent={60} />
                <div style={{ marginTop: 16 }}>正在发送测试请求...</div>
              </div>
            )}
            
            {testResult.status === 'success' && (
              <div>
                <Alert
                  message="测试成功"
                  description="Webhook响应正常"
                  type="success"
                  showIcon
                  style={{ marginBottom: 16 }}
                />
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="状态码">{testResult.statusCode}</Descriptions.Item>
                  <Descriptions.Item label="响应时间">{testResult.responseTime}ms</Descriptions.Item>
                  <Descriptions.Item label="响应内容">
                    <pre style={{ margin: 0, fontSize: 12 }}>
                      {JSON.stringify(testResult.response, null, 2)}
                    </pre>
                  </Descriptions.Item>
                </Descriptions>
              </div>
            )}
            
            {testResult.status === 'error' && (
              <div>
                <Alert
                  message="测试失败"
                  description={testResult.error}
                  type="error"
                  showIcon
                  style={{ marginBottom: 16 }}
                />
                {testResult.statusCode && (
                  <Descriptions column={1} bordered size="small">
                    <Descriptions.Item label="状态码">{testResult.statusCode}</Descriptions.Item>
                    <Descriptions.Item label="响应时间">{testResult.responseTime}ms</Descriptions.Item>
                  </Descriptions>
                )}
              </div>
            )}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Button
              type="primary"
              icon={<ExperimentOutlined />}
              onClick={() => selectedWebhook && handleTestWebhook(selectedWebhook)}
            >
              发送测试请求
            </Button>
          </div>
        )}
      </Modal>

      {/* Webhook日志抽屉 */}
      <Drawer
        title={`Webhook日志 - ${selectedWebhook?.name}`}
        placement="right"
        onClose={() => {
          setLogsDrawerVisible(false);
          setSelectedWebhook(null);
          setSelectedLog(null);
        }}
        open={logsDrawerVisible}
        width={800}
      >
        <List
          dataSource={webhookLogs.filter(log => log.webhookId === selectedWebhook?.id)}
          renderItem={(log) => (
            <List.Item
              actions={[
                <Button
                  type="text"
                  icon={<EyeOutlined />}
                  onClick={() => setSelectedLog(log)}
                >
                  详情
                </Button>,
                log.status === 'failed' && log.attempt < log.maxAttempts && (
                  <Button
                    type="text"
                    icon={<ReloadOutlined />}
                    onClick={() => handleRetryWebhook(log.id)}
                  >
                    重试
                  </Button>
                )
              ].filter(Boolean)}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <Tag color="blue">{log.event}</Tag>
                    {getLogStatusTag(log.status)}
                    {log.statusCode && (
                      <Tag color={log.statusCode < 400 ? 'green' : 'red'}>
                        {log.statusCode}
                      </Tag>
                    )}
                  </Space>
                }
                description={
                  <div>
                    <div>时间: {dayjs(log.timestamp).format('YYYY-MM-DD HH:mm:ss')}</div>
                    {log.responseTime && <div>响应时间: {log.responseTime}ms</div>}
                    {log.error && <div style={{ color: '#ff4d4f' }}>错误: {log.error}</div>}
                    <div>尝试次数: {log.attempt}/{log.maxAttempts}</div>
                  </div>
                }
              />
            </List.Item>
          )}
        />

        {/* 日志详情模态框 */}
        <Modal
          title="日志详情"
          open={!!selectedLog}
          onCancel={() => setSelectedLog(null)}
          footer={
            <Button onClick={() => setSelectedLog(null)}>
              关闭
            </Button>
          }
          width={800}
        >
          {selectedLog && (
            <Tabs defaultActiveKey="payload">
              <TabPane tab="请求载荷" key="payload">
                <div style={{ marginBottom: 8 }}>
                  <Button
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(JSON.stringify(selectedLog.payload, null, 2))}
                  >
                    复制
                  </Button>
                </div>
                <pre style={{ 
                  background: '#f5f5f5', 
                  padding: 12, 
                  borderRadius: 4,
                  fontSize: 12,
                  maxHeight: 400,
                  overflow: 'auto'
                }}>
                  {JSON.stringify(selectedLog.payload, null, 2)}
                </pre>
              </TabPane>
              
              {selectedLog.response && (
                <TabPane tab="响应内容" key="response">
                  <div style={{ marginBottom: 8 }}>
                    <Button
                      size="small"
                      icon={<CopyOutlined />}
                      onClick={() => copyToClipboard(JSON.stringify(selectedLog.response, null, 2))}
                    >
                      复制
                    </Button>
                  </div>
                  <pre style={{ 
                    background: '#f5f5f5', 
                    padding: 12, 
                    borderRadius: 4,
                    fontSize: 12,
                    maxHeight: 400,
                    overflow: 'auto'
                  }}>
                    {JSON.stringify(selectedLog.response, null, 2)}
                  </pre>
                </TabPane>
              )}
              
              <TabPane tab="基本信息" key="info">
                <Descriptions column={1} bordered>
                  <Descriptions.Item label="事件">{selectedLog.event}</Descriptions.Item>
                  <Descriptions.Item label="状态">{getLogStatusTag(selectedLog.status)}</Descriptions.Item>
                  <Descriptions.Item label="状态码">{selectedLog.statusCode || '-'}</Descriptions.Item>
                  <Descriptions.Item label="响应时间">{selectedLog.responseTime ? `${selectedLog.responseTime}ms` : '-'}</Descriptions.Item>
                  <Descriptions.Item label="尝试次数">{selectedLog.attempt}/{selectedLog.maxAttempts}</Descriptions.Item>
                  <Descriptions.Item label="时间戳">{dayjs(selectedLog.timestamp).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
                  {selectedLog.error && (
                    <Descriptions.Item label="错误信息">
                      <Text type="danger">{selectedLog.error}</Text>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </TabPane>
            </Tabs>
          )}
        </Modal>
      </Drawer>
    </div>
  );
};

export default WebhookManagement;