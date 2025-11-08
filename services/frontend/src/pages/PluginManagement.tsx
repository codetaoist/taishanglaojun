import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Typography, 
  Table, 
  Button, 
  Space, 
  Tag, 
  Modal, 
  Form, 
  Input, 
  message, 
  Popconfirm,
  Tooltip,
  Descriptions,
  Divider,
  Row,
  Col
} from 'antd';
import { 
  PlusOutlined, 
  PlayCircleOutlined, 
  PauseCircleOutlined, 
  UpOutlined, 
  DeleteOutlined,
  InfoCircleOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { pluginApi, Plugin, PluginStatus } from '../services/laojunApi';

const { Title } = Typography;
const { TextArea } = Input;

const PluginManagement: React.FC = () => {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [loading, setLoading] = useState(false);
  const [installModalVisible, setInstallModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);
  const [form] = Form.useForm();

  // 获取插件列表
  const fetchPlugins = async () => {
    setLoading(true);
    try {
      const response = await pluginApi.getAll();
      if (response.code === 200) {
        setPlugins(response.data || []);
      } else {
        message.error('获取插件列表失败');
      }
    } catch (error) {
      message.error('获取插件列表失败');
      console.error('获取插件列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载插件列表
  useEffect(() => {
    fetchPlugins();
  }, []);

  // 安装插件
  const handleInstallPlugin = async (values: any) => {
    try {
      const response = await pluginApi.install(values.id, values.source, values.version);
      if (response.code === 200) {
        message.success('插件安装成功');
        setInstallModalVisible(false);
        form.resetFields();
        fetchPlugins();
      } else {
        message.error(response.message || '插件安装失败');
      }
    } catch (error) {
      message.error('插件安装失败');
      console.error('插件安装失败:', error);
    }
  };

  // 启动插件
  const handleStartPlugin = async (id: string) => {
    try {
      const response = await pluginApi.start(id);
      if (response.code === 200) {
        message.success('插件启动成功');
        fetchPlugins();
      } else {
        message.error(response.message || '插件启动失败');
      }
    } catch (error) {
      message.error('插件启动失败');
      console.error('插件启动失败:', error);
    }
  };

  // 停止插件
  const handleStopPlugin = async (id: string) => {
    try {
      const response = await pluginApi.stop(id);
      if (response.code === 200) {
        message.success('插件停止成功');
        fetchPlugins();
      } else {
        message.error(response.message || '插件停止失败');
      }
    } catch (error) {
      message.error('插件停止失败');
      console.error('插件停止失败:', error);
    }
  };

  // 升级插件
  const handleUpgradePlugin = async (id: string, version?: string) => {
    try {
      const response = await pluginApi.upgrade(id, version);
      if (response.code === 200) {
        message.success('插件升级成功');
        fetchPlugins();
      } else {
        message.error(response.message || '插件升级失败');
      }
    } catch (error) {
      message.error('插件升级失败');
      console.error('插件升级失败:', error);
    }
  };

  // 卸载插件
  const handleUninstallPlugin = async (id: string) => {
    try {
      const response = await pluginApi.uninstall(id);
      if (response.code === 200) {
        message.success('插件卸载成功');
        fetchPlugins();
      } else {
        message.error(response.message || '插件卸载失败');
      }
    } catch (error) {
      message.error('插件卸载失败');
      console.error('插件卸载失败:', error);
    }
  };

  // 查看插件详情
  const handleViewPluginDetail = async (plugin: Plugin) => {
    try {
      const response = await pluginApi.get(plugin.id);
      if (response.code === 200) {
        setSelectedPlugin(response.data);
        setDetailModalVisible(true);
      } else {
        message.error('获取插件详情失败');
      }
    } catch (error) {
      message.error('获取插件详情失败');
      console.error('获取插件详情失败:', error);
    }
  };

  // 获取状态标签颜色
  const getStatusColor = (status: PluginStatus) => {
    switch (status) {
      case PluginStatus.Installed:
        return 'default';
      case PluginStatus.Running:
        return 'success';
      case PluginStatus.Stopped:
        return 'warning';
      case PluginStatus.Error:
        return 'error';
      default:
        return 'default';
    }
  };

  // 获取状态文本
  const getStatusText = (status: PluginStatus) => {
    switch (status) {
      case PluginStatus.Installed:
        return '已安装';
      case PluginStatus.Running:
        return '运行中';
      case PluginStatus.Stopped:
        return '已停止';
      case PluginStatus.Error:
        return '错误';
      default:
        return '未知';
    }
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: PluginStatus) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author',
      width: 150,
      render: (author: string) => author || '-',
    },
    {
      title: '安装时间',
      dataIndex: 'installedAt',
      key: 'installedAt',
      width: 180,
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 300,
      render: (_: any, record: Plugin) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<InfoCircleOutlined />} 
              onClick={() => handleViewPluginDetail(record)}
            />
          </Tooltip>
          
          {record.status === PluginStatus.Stopped && (
            <Tooltip title="启动">
              <Button 
                type="link" 
                icon={<PlayCircleOutlined />} 
                onClick={() => handleStartPlugin(record.id)}
              />
            </Tooltip>
          )}
          
          {record.status === PluginStatus.Running && (
            <Tooltip title="停止">
              <Button 
                type="link" 
                icon={<PauseCircleOutlined />} 
                onClick={() => handleStopPlugin(record.id)}
              />
            </Tooltip>
          )}
          
          <Tooltip title="升级">
            <Button 
              type="link" 
              icon={<UpOutlined />} 
              onClick={() => handleUpgradePlugin(record.id)}
            />
          </Tooltip>
          
          <Popconfirm
            title="确定要卸载此插件吗？"
            onConfirm={() => handleUninstallPlugin(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="卸载">
              <Button 
                type="link" 
                danger 
                icon={<DeleteOutlined />} 
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Title level={2}>插件管理</Title>
      
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <div>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={() => setInstallModalVisible(true)}
            >
              安装插件
            </Button>
          </div>
          
          <div>
            <Button 
              icon={<ReloadOutlined />}
              onClick={fetchPlugins}
              loading={loading}
            >
              刷新
            </Button>
          </div>
        </div>
        
        <Table
          columns={columns}
          dataSource={plugins}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个插件`,
          }}
        />
      </Card>

      {/* 安装插件模态框 */}
      <Modal
        title="安装插件"
        open={installModalVisible}
        onCancel={() => setInstallModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleInstallPlugin}
        >
          <Form.Item
            name="id"
            label="插件ID"
            rules={[{ required: true, message: '请输入插件ID' }]}
          >
            <Input placeholder="请输入插件ID" />
          </Form.Item>
          
          <Form.Item
            name="source"
            label="插件源"
            rules={[{ required: true, message: '请输入插件源' }]}
          >
            <TextArea rows={3} placeholder="请输入插件源地址" />
          </Form.Item>
          
          <Form.Item
            name="version"
            label="版本"
          >
            <Input placeholder="可选，留空安装最新版本" />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                安装
              </Button>
              <Button onClick={() => setInstallModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 插件详情模态框 */}
      <Modal
        title="插件详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={800}
      >
        {selectedPlugin && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="ID">{selectedPlugin.id}</Descriptions.Item>
              <Descriptions.Item label="名称">{selectedPlugin.name}</Descriptions.Item>
              <Descriptions.Item label="版本">{selectedPlugin.version}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedPlugin.status)}>
                  {getStatusText(selectedPlugin.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="作者">{selectedPlugin.author || '-'}</Descriptions.Item>
              <Descriptions.Item label="主页">
                {selectedPlugin.homepage ? (
                  <a href={selectedPlugin.homepage} target="_blank" rel="noopener noreferrer">
                    {selectedPlugin.homepage}
                  </a>
                ) : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="安装时间">
                {selectedPlugin.installedAt ? new Date(selectedPlugin.installedAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {selectedPlugin.updatedAt ? new Date(selectedPlugin.updatedAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>
                {selectedPlugin.description || '-'}
              </Descriptions.Item>
            </Descriptions>
            
            {selectedPlugin.manifest && (
              <>
                <Divider>插件清单</Divider>
                <Row>
                  <Col span={24}>
                    <pre style={{ 
                      background: '#f5f5f5', 
                      padding: '10px', 
                      borderRadius: '4px',
                      overflow: 'auto',
                      maxHeight: '300px'
                    }}>
                      {JSON.stringify(selectedPlugin.manifest, null, 2)}
                    </pre>
                  </Col>
                </Row>
              </>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default PluginManagement;