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
  Upload,
  Progress,
  Tabs,
  List,
  Avatar,
  Rate,
  Badge
} from 'antd';
import {
  PlusOutlined,
  AppstoreOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  UploadOutlined,
  DownloadOutlined,
  SettingOutlined,
  InfoCircleOutlined,
  StarOutlined,
  CloudDownloadOutlined,
  BugOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;
const { TabPane } = Tabs;

interface Plugin {
  id: number;
  name: string;
  version: string;
  description: string;
  author: string;
  status: 'installed' | 'installing' | 'failed' | 'updating';
  isEnabled: boolean;
  config: Record<string, any>;
  manifest: {
    permissions: string[];
    dependencies: string[];
    category: string;
    homepage?: string;
    repository?: string;
  };
  installedAt: string;
  updatedAt: string;
  size?: string;
  rating?: number;
  downloads?: number;
}

interface PluginMarketItem {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  category: string;
  rating: number;
  downloads: number;
  size: string;
  tags: string[];
  screenshots: string[];
  isInstalled: boolean;
  latestVersion: string;
}

const PluginManagement: React.FC = () => {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [marketPlugins, setMarketPlugins] = useState<PluginMarketItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [installModalVisible, setInstallModalVisible] = useState(false);
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [marketModalVisible, setMarketModalVisible] = useState(false);
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);
  const [selectedMarketPlugin, setSelectedMarketPlugin] = useState<PluginMarketItem | null>(null);
  const [activeTab, setActiveTab] = useState('installed');
  const [form] = Form.useForm();
  const [configForm] = Form.useForm();

  useEffect(() => {
    fetchPlugins();
    fetchMarketPlugins();
  }, []);

  const fetchPlugins = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      const mockData: Plugin[] = [
        {
          id: 1,
          name: 'AI助手插件',
          version: '1.2.0',
          description: '集成AI助手功能，提供智能对话和自动化处理',
          author: 'AI Team',
          status: 'installed',
          isEnabled: true,
          config: {
            apiKey: '***',
            model: 'gpt-4',
            maxTokens: 2048
          },
          manifest: {
            permissions: ['api_access', 'user_data'],
            dependencies: ['openai'],
            category: 'AI',
            homepage: 'https://example.com/ai-assistant'
          },
          installedAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-15T10:30:00Z',
          size: '2.5MB',
          rating: 4.8,
          downloads: 15420
        },
        {
          id: 2,
          name: '数据同步插件',
          version: '2.1.3',
          description: '自动同步数据到第三方服务',
          author: 'Data Team',
          status: 'installed',
          isEnabled: false,
          config: {
            syncInterval: 3600,
            targetService: 'webhook',
            retryCount: 3
          },
          manifest: {
            permissions: ['data_access', 'network'],
            dependencies: ['axios'],
            category: 'Integration',
            repository: 'https://github.com/example/data-sync'
          },
          installedAt: '2024-01-05T00:00:00Z',
          updatedAt: '2024-01-10T15:20:00Z',
          size: '1.8MB',
          rating: 4.5,
          downloads: 8930
        },
        {
          id: 3,
          name: '监控插件',
          version: '1.0.0',
          description: '系统性能监控和告警',
          author: 'Monitor Team',
          status: 'installing',
          isEnabled: false,
          config: {},
          manifest: {
            permissions: ['system_monitor'],
            dependencies: ['prometheus'],
            category: 'Monitoring'
          },
          installedAt: '2024-01-16T00:00:00Z',
          updatedAt: '2024-01-16T00:00:00Z',
          size: '3.2MB'
        }
      ];
      setPlugins(mockData);
    } catch (error) {
      message.error('获取插件列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchMarketPlugins = async () => {
    try {
      // 模拟插件市场数据
      const mockMarketData: PluginMarketItem[] = [
        {
          id: 'email-plugin',
          name: '邮件通知插件',
          version: '1.5.0',
          description: '发送邮件通知和自动化邮件处理',
          author: 'Email Team',
          category: 'Communication',
          rating: 4.7,
          downloads: 25680,
          size: '1.2MB',
          tags: ['email', 'notification', 'automation'],
          screenshots: [],
          isInstalled: false,
          latestVersion: '1.5.0'
        },
        {
          id: 'backup-plugin',
          name: '自动备份插件',
          version: '2.0.1',
          description: '定期备份数据到云存储',
          author: 'Backup Team',
          category: 'Utility',
          rating: 4.9,
          downloads: 18750,
          size: '2.8MB',
          tags: ['backup', 'cloud', 'automation'],
          screenshots: [],
          isInstalled: false,
          latestVersion: '2.0.1'
        },
        {
          id: 'analytics-plugin',
          name: '数据分析插件',
          version: '3.1.0',
          description: '高级数据分析和可视化',
          author: 'Analytics Team',
          category: 'Analytics',
          rating: 4.6,
          downloads: 12340,
          size: '4.5MB',
          tags: ['analytics', 'visualization', 'charts'],
          screenshots: [],
          isInstalled: false,
          latestVersion: '3.1.0'
        }
      ];
      setMarketPlugins(mockMarketData);
    } catch (error) {
      message.error('获取插件市场数据失败');
    }
  };

  const handleInstallPlugin = async (values: any) => {
    try {
      // 模拟安装过程
      const newPlugin: Plugin = {
        id: Date.now(),
        name: values.name,
        version: values.version || '1.0.0',
        description: values.description,
        author: values.author || 'Unknown',
        status: 'installing',
        isEnabled: false,
        config: {},
        manifest: {
          permissions: [],
          dependencies: [],
          category: 'Custom'
        },
        installedAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      setPlugins([newPlugin, ...plugins]);
      setInstallModalVisible(false);
      form.resetFields();
      
      // 模拟安装完成
      setTimeout(() => {
        setPlugins(prev => prev.map(p => 
          p.id === newPlugin.id 
            ? { ...p, status: 'installed' as const }
            : p
        ));
        message.success('插件安装成功');
      }, 2000);
      
      message.info('插件安装中...');
    } catch (error) {
      message.error('安装插件失败');
    }
  };

  const handleInstallFromMarket = async (marketPlugin: PluginMarketItem) => {
    try {
      const newPlugin: Plugin = {
        id: Date.now(),
        name: marketPlugin.name,
        version: marketPlugin.version,
        description: marketPlugin.description,
        author: marketPlugin.author,
        status: 'installing',
        isEnabled: false,
        config: {},
        manifest: {
          permissions: [],
          dependencies: [],
          category: marketPlugin.category
        },
        installedAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        size: marketPlugin.size,
        rating: marketPlugin.rating,
        downloads: marketPlugin.downloads
      };

      setPlugins([newPlugin, ...plugins]);
      setMarketModalVisible(false);
      
      // 模拟安装完成
      setTimeout(() => {
        setPlugins(prev => prev.map(p => 
          p.id === newPlugin.id 
            ? { ...p, status: 'installed' as const }
            : p
        ));
        message.success(`${marketPlugin.name} 安装成功`);
      }, 2000);
      
      message.info(`正在安装 ${marketPlugin.name}...`);
    } catch (error) {
      message.error('安装插件失败');
    }
  };

  const handleTogglePlugin = async (id: number, enabled: boolean) => {
    try {
      setPlugins(plugins.map(plugin =>
        plugin.id === id
          ? { ...plugin, isEnabled: enabled, updatedAt: new Date().toISOString() }
          : plugin
      ));
      message.success(enabled ? '插件已启用' : '插件已禁用');
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleDeletePlugin = async (id: number) => {
    try {
      setPlugins(plugins.filter(plugin => plugin.id !== id));
      message.success('插件卸载成功');
    } catch (error) {
      message.error('卸载插件失败');
    }
  };

  const handleUpdateConfig = async (values: any) => {
    if (!selectedPlugin) return;

    try {
      setPlugins(plugins.map(plugin =>
        plugin.id === selectedPlugin.id
          ? { ...plugin, config: values, updatedAt: new Date().toISOString() }
          : plugin
      ));
      setConfigModalVisible(false);
      setSelectedPlugin(null);
      configForm.resetFields();
      message.success('配置更新成功');
    } catch (error) {
      message.error('更新配置失败');
    }
  };

  const getStatusTag = (status: string) => {
    const statusConfig = {
      installed: { color: 'green', text: '已安装', icon: <CheckCircleOutlined /> },
      installing: { color: 'blue', text: '安装中', icon: <CloudDownloadOutlined /> },
      failed: { color: 'red', text: '安装失败', icon: <ExclamationCircleOutlined /> },
      updating: { color: 'orange', text: '更新中', icon: <CloudDownloadOutlined /> }
    };
    const config = statusConfig[status as keyof typeof statusConfig];
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  const columns: ColumnsType<Plugin> = [
    {
      title: '插件信息',
      key: 'info',
      render: (_, record) => (
        <div>
          <div style={{ fontWeight: 500, marginBottom: 4 }}>
            {record.name} v{record.version}
          </div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.description}
          </Text>
          <div style={{ marginTop: 4 }}>
            <Text type="secondary" style={{ fontSize: 11 }}>
              作者: {record.author}
            </Text>
            {record.size && (
              <Text type="secondary" style={{ fontSize: 11, marginLeft: 8 }}>
                大小: {record.size}
              </Text>
            )}
          </div>
        </div>
      )
    },
    {
      title: '类别',
      dataIndex: ['manifest', 'category'],
      key: 'category',
      render: (category) => (
        <Tag color="blue">{category}</Tag>
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
        <Badge
          status={enabled ? 'success' : 'default'}
          text={enabled ? '已启用' : '已禁用'}
        />
      )
    },
    {
      title: '评分',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating) => rating ? (
        <Space>
          <Rate disabled defaultValue={rating} style={{ fontSize: 12 }} />
          <Text style={{ fontSize: 12 }}>{rating}</Text>
        </Space>
      ) : '-'
    },
    {
      title: '安装时间',
      dataIndex: 'installedAt',
      key: 'installedAt',
      render: (date) => dayjs(date).format('YYYY-MM-DD HH:mm')
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {record.status === 'installed' && (
            <>
              <Tooltip title={record.isEnabled ? '禁用插件' : '启用插件'}>
                <Button
                  type="text"
                  icon={record.isEnabled ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                  onClick={() => handleTogglePlugin(record.id, !record.isEnabled)}
                />
              </Tooltip>
              <Tooltip title="配置">
                <Button
                  type="text"
                  icon={<SettingOutlined />}
                  onClick={() => {
                    setSelectedPlugin(record);
                    configForm.setFieldsValue(record.config);
                    setConfigModalVisible(true);
                  }}
                />
              </Tooltip>
            </>
          )}
          {record.status === 'installing' && (
            <Progress type="circle" size={24} percent={60} />
          )}
          <Popconfirm
            title="确定要卸载这个插件吗？"
            onConfirm={() => handleDeletePlugin(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="卸载">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                disabled={record.status === 'installing'}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const installedPlugins = plugins.filter(p => p.status === 'installed').length;
  const enabledPlugins = plugins.filter(p => p.isEnabled).length;
  const totalDownloads = plugins.reduce((sum, p) => sum + (p.downloads || 0), 0);

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="已安装插件"
              value={installedPlugins}
              prefix={<AppstoreOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已启用插件"
              value={enabledPlugins}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总下载量"
              value={totalDownloads}
              prefix={<DownloadOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="可用更新"
              value={2}
              prefix={<BugOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <Space>
            <AppstoreOutlined />
            插件管理
          </Space>
        }
        extra={
          <Space>
            <Button
              icon={<CloudDownloadOutlined />}
              onClick={() => setMarketModalVisible(true)}
            >
              插件市场
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setInstallModalVisible(true)}
            >
              安装插件
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={plugins}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个插件`
          }}
        />
      </Card>

      {/* 安装插件模态框 */}
      <Modal
        title="安装插件"
        open={installModalVisible}
        onCancel={() => {
          setInstallModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Tabs defaultActiveKey="upload">
          <TabPane tab="上传安装" key="upload">
            <Form
              form={form}
              layout="vertical"
              onFinish={handleInstallPlugin}
            >
              <Form.Item
                name="file"
                label="插件文件"
              >
                <Upload.Dragger
                  name="file"
                  multiple={false}
                  accept=".zip,.tar.gz"
                  beforeUpload={() => false}
                >
                  <p className="ant-upload-drag-icon">
                    <UploadOutlined />
                  </p>
                  <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
                  <p className="ant-upload-hint">
                    支持 .zip 和 .tar.gz 格式的插件包
                  </p>
                </Upload.Dragger>
              </Form.Item>

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit">
                    安装插件
                  </Button>
                  <Button onClick={() => {
                    setInstallModalVisible(false);
                    form.resetFields();
                  }}>
                    取消
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </TabPane>
          
          <TabPane tab="手动安装" key="manual">
            <Form
              form={form}
              layout="vertical"
              onFinish={handleInstallPlugin}
            >
              <Form.Item
                name="name"
                label="插件名称"
                rules={[{ required: true, message: '请输入插件名称' }]}
              >
                <Input placeholder="输入插件名称" />
              </Form.Item>

              <Form.Item
                name="version"
                label="版本"
              >
                <Input placeholder="输入版本号（可选）" />
              </Form.Item>

              <Form.Item
                name="description"
                label="描述"
              >
                <TextArea
                  placeholder="输入插件描述"
                  rows={3}
                />
              </Form.Item>

              <Form.Item
                name="author"
                label="作者"
              >
                <Input placeholder="输入作者名称" />
              </Form.Item>

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit">
                    安装插件
                  </Button>
                  <Button onClick={() => {
                    setInstallModalVisible(false);
                    form.resetFields();
                  }}>
                    取消
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </TabPane>
        </Tabs>
      </Modal>

      {/* 插件配置模态框 */}
      <Modal
        title={`配置 ${selectedPlugin?.name}`}
        open={configModalVisible}
        onCancel={() => {
          setConfigModalVisible(false);
          setSelectedPlugin(null);
          configForm.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={configForm}
          layout="vertical"
          onFinish={handleUpdateConfig}
        >
          {selectedPlugin && Object.keys(selectedPlugin.config).map(key => (
            <Form.Item
              key={key}
              name={key}
              label={key}
            >
              <Input placeholder={`输入 ${key}`} />
            </Form.Item>
          ))}

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                保存配置
              </Button>
              <Button onClick={() => {
                setConfigModalVisible(false);
                setSelectedPlugin(null);
                configForm.resetFields();
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 插件市场模态框 */}
      <Modal
        title="插件市场"
        open={marketModalVisible}
        onCancel={() => setMarketModalVisible(false)}
        footer={null}
        width={1000}
      >
        <List
          grid={{ gutter: 16, column: 2 }}
          dataSource={marketPlugins}
          renderItem={(item) => (
            <List.Item>
              <Card
                hoverable
                actions={[
                  <Button
                    type="primary"
                    icon={<DownloadOutlined />}
                    onClick={() => handleInstallFromMarket(item)}
                    disabled={item.isInstalled}
                  >
                    {item.isInstalled ? '已安装' : '安装'}
                  </Button>,
                  <Button
                    type="text"
                    icon={<InfoCircleOutlined />}
                    onClick={() => setSelectedMarketPlugin(item)}
                  >
                    详情
                  </Button>
                ]}
              >
                <Card.Meta
                  avatar={<Avatar icon={<AppstoreOutlined />} />}
                  title={
                    <Space>
                      {item.name}
                      <Tag color="blue">v{item.version}</Tag>
                    </Space>
                  }
                  description={
                    <div>
                      <Paragraph ellipsis={{ rows: 2 }}>
                        {item.description}
                      </Paragraph>
                      <Space>
                        <Rate disabled defaultValue={item.rating} style={{ fontSize: 12 }} />
                        <Text style={{ fontSize: 12 }}>{item.rating}</Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {item.downloads.toLocaleString()} 下载
                        </Text>
                      </Space>
                      <div style={{ marginTop: 8 }}>
                        {item.tags.map(tag => (
                          <Tag key={tag} size="small">{tag}</Tag>
                        ))}
                      </div>
                    </div>
                  }
                />
              </Card>
            </List.Item>
          )}
        />
      </Modal>
    </div>
  );
};

export default PluginManagement;