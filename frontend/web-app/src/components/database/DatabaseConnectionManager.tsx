import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  Switch,
  message,
  Popconfirm,
  Tooltip,
  Badge,
  Row,
  Col,
  Statistic,
  Divider,
  Upload,
  Progress,
  Typography
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  ExportOutlined,
  ImportOutlined,
  DatabaseOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  SyncOutlined,
  DownloadOutlined,
  UploadOutlined
} from '@ant-design/icons';
import { DatabaseConnectionService } from '../../services/databaseConnectionService';
import type {
  DatabaseConnectionListItem,
  DatabaseConnectionForm,
  DatabaseConnectionStats,
  DatabaseConnectionStatus
} from '../../types/database';
import { DatabaseType, ConnectionStatus, DATABASE_TYPE_CONFIGS } from '../../types/database';

const { Title, Text } = Typography;
const { Option } = Select;

interface DatabaseConnectionManagerProps {
  className?: string;
}

const DatabaseConnectionManager: React.FC<DatabaseConnectionManagerProps> = ({ className }) => {
  // 状态管理
  const [connections, setConnections] = useState<DatabaseConnectionListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<DatabaseConnectionStats | null>(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  
  // 分页和搜索
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });
  const [searchText, setSearchText] = useState('');
  const [filterType, setFilterType] = useState<DatabaseType | undefined>();
  const [filterStatus, setFilterStatus] = useState<ConnectionStatus | undefined>();

  // 模态框状态
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingConnection, setEditingConnection] = useState<DatabaseConnectionListItem | null>(null);
  const [testingConnections, setTestingConnections] = useState<Set<string>>(new Set());
  
  // 表单
  const [form] = Form.useForm<DatabaseConnectionForm>();

  // 加载数据
  const loadConnections = useCallback(async () => {
    try {
      setLoading(true);
      const response = await DatabaseConnectionService.getConnections({
        page: pagination.current,
        pageSize: pagination.pageSize,
        search: searchText || undefined,
        type: filterType,
        status: filterStatus
      });

      if (response.success && response.data) {
        setConnections(response.data.connections);
        setPagination(prev => ({
          ...prev,
          total: response.data!.total
        }));
      }
    } catch (error) {
      message.error('加载数据库连接列表失败');
    } finally {
      setLoading(false);
    }
  }, [pagination.current, pagination.pageSize, searchText, filterType, filterStatus]);

  // 加载统计信息
  const loadStats = useCallback(async () => {
    try {
      const response = await DatabaseConnectionService.getConnectionStats();
      if (response.success && response.data) {
        setStats(response.data);
      }
    } catch (error) {
      console.error('加载统计信息失败:', error);
    }
  }, []);

  // 初始化加载
  useEffect(() => {
    loadConnections();
    loadStats();
  }, [loadConnections, loadStats]);

  // 处理表单提交
  const handleSubmit = async (values: DatabaseConnectionForm) => {
    try {
      if (editingConnection) {
        await DatabaseConnectionService.updateConnection(editingConnection.id, values);
        message.success('更新数据库连接成功');
      } else {
        await DatabaseConnectionService.createConnection(values);
        message.success('创建数据库连接成功');
      }
      
      setIsModalVisible(false);
      setEditingConnection(null);
      form.resetFields();
      loadConnections();
      loadStats();
    } catch (error) {
      message.error(editingConnection ? '更新数据库连接失败' : '创建数据库连接失败');
    }
  };

  // 测试连接
  const handleTestConnection = async (connection: DatabaseConnectionListItem) => {
    setTestingConnections(prev => new Set(prev).add(connection.id));
    try {
      const response = await DatabaseConnectionService.testSavedConnection(connection.id);
      if (response.success && response.data) {
        message.success(`连接测试成功 (${response.data.responseTime}ms)`);
      } else {
        message.error(response.message || '连接测试失败');
      }
    } catch (error) {
      message.error('连接测试失败');
    } finally {
      setTestingConnections(prev => {
        const newSet = new Set(prev);
        newSet.delete(connection.id);
        return newSet;
      });
    }
  };

  // 删除连接
  const handleDelete = async (id: string) => {
    try {
      await DatabaseConnectionService.deleteConnection(id);
      message.success('删除数据库连接成功');
      loadConnections();
      loadStats();
    } catch (error) {
      message.error('删除数据库连接失败');
    }
  };

  // 批量操作
  const handleBatchOperation = async (operation: 'connect' | 'disconnect' | 'test' | 'delete') => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要操作的连接');
      return;
    }

    try {
      const response = await DatabaseConnectionService.batchOperation(operation, selectedRowKeys);
      if (response.success) {
        message.success(`批量${operation === 'delete' ? '删除' : operation === 'test' ? '测试' : operation === 'connect' ? '连接' : '断开'}成功`);
        setSelectedRowKeys([]);
        loadConnections();
        loadStats();
      } else {
        message.error(response.message || '批量操作失败');
      }
    } catch (error) {
      message.error('批量操作失败');
    }
  };

  // 导出配置
  const handleExport = async () => {
    try {
      const blob = await DatabaseConnectionService.exportConnections(
        selectedRowKeys.length > 0 ? selectedRowKeys : undefined
      );
      
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `database-connections-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
      message.success('导出成功');
    } catch (error) {
      message.error('导出失败');
    }
  };

  // 导入配置
  const handleImport = async (file: File) => {
    try {
      const response = await DatabaseConnectionService.importConnections(file);
      if (response.success) {
        message.success(`导入成功: ${response.imported} 个连接，${response.failed} 个失败`);
        loadConnections();
        loadStats();
      } else {
        message.error(response.message || '导入失败');
      }
    } catch (error) {
      message.error('导入失败');
    }
    return false; // 阻止默认上传行为
  };

  // 打开编辑模态框
  const handleEdit = (connection: DatabaseConnectionListItem) => {
    setEditingConnection(connection);
    form.setFieldsValue({
      name: connection.name,
      type: connection.type,
      host: connection.host,
      port: connection.port,
      database: connection.database,
      description: connection.description,
      tags: connection.tags,
      isDefault: connection.isDefault
    });
    setIsModalVisible(true);
  };

  // 获取状态图标
  const getStatusIcon = (status: ConnectionStatus) => {
    switch (status) {
      case ConnectionStatus.CONNECTED:
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case ConnectionStatus.DISCONNECTED:
        return <CloseCircleOutlined style={{ color: '#8c8c8c' }} />;
      case ConnectionStatus.CONNECTING:
        return <SyncOutlined spin style={{ color: '#1890ff' }} />;
      case ConnectionStatus.ERROR:
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
      default:
        return <ExclamationCircleOutlined style={{ color: '#faad14' }} />;
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: ConnectionStatus) => {
    switch (status) {
      case ConnectionStatus.CONNECTED:
        return 'success';
      case ConnectionStatus.DISCONNECTED:
        return 'default';
      case ConnectionStatus.CONNECTING:
        return 'processing';
      case ConnectionStatus.ERROR:
        return 'error';
      default:
        return 'warning';
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: DatabaseConnectionListItem) => (
        <Space>
          <DatabaseOutlined />
          <span>{text}</span>
          {record.isDefault && <Tag color="blue">默认</Tag>}
        </Space>
      )
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: DatabaseType) => {
        const config = DATABASE_TYPE_CONFIGS[type];
        return (
          <Tag color="geekblue">
            {config?.name || type}
          </Tag>
        );
      }
    },
    {
      title: '主机',
      dataIndex: 'host',
      key: 'host',
      render: (host: string, record: DatabaseConnectionListItem) => (
        <Text code>{host}:{record.port}</Text>
      )
    },
    {
      title: '数据库',
      dataIndex: 'database',
      key: 'database',
      render: (text: string) => <Text code>{text}</Text>
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: ConnectionStatus) => (
        <Space>
          {getStatusIcon(status)}
          <Tag color={getStatusColor(status)}>
            {status === ConnectionStatus.CONNECTED ? '已连接' :
             status === ConnectionStatus.DISCONNECTED ? '未连接' :
             status === ConnectionStatus.CONNECTING ? '连接中' :
             status === ConnectionStatus.ERROR ? '错误' : '未知'}
          </Tag>
        </Space>
      )
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => (
        <>
          {tags?.map(tag => (
            <Tag key={tag} color="cyan">{tag}</Tag>
          ))}
        </>
      )
    },
    {
      title: '最后连接',
      dataIndex: 'lastConnectedAt',
      key: 'lastConnectedAt',
      render: (text: string) => text ? new Date(text).toLocaleString() : '-'
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record: DatabaseConnectionListItem) => (
        <Space>
          <Tooltip title="测试连接">
            <Button
              type="text"
              icon={<PlayCircleOutlined />}
              loading={testingConnections.has(record.id)}
              onClick={() => handleTestConnection(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个数据库连接吗？"
            onConfirm={() => handleDelete(record.id)}
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

  return (
    <div className={className}>
      {/* 统计信息 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总连接数"
                value={stats.totalConnections}
                prefix={<DatabaseOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃连接"
                value={stats.activeConnections}
                prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="平均响应时间"
                value={stats.averageResponseTime}
                suffix="ms"
                prefix={<SyncOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="最后更新"
                value={new Date(stats.lastUpdated).toLocaleString()}
                prefix={<ReloadOutlined />}
              />
            </Card>
          </Col>
        </Row>
      )}

      <Card
        title={
          <Space>
            <DatabaseOutlined />
            <span>数据库连接管理</span>
          </Space>
        }
        extra={
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditingConnection(null);
                form.resetFields();
                setIsModalVisible(true);
              }}
            >
              新建连接
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                loadConnections();
                loadStats();
              }}
            >
              刷新
            </Button>
            <Button
              icon={<ExportOutlined />}
              onClick={handleExport}
            >
              导出
            </Button>
            <Upload
              accept=".json"
              showUploadList={false}
              beforeUpload={handleImport}
            >
              <Button icon={<ImportOutlined />}>
                导入
              </Button>
            </Upload>
          </Space>
        }
      >
        {/* 搜索和筛选 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={8}>
            <Input
              placeholder="搜索连接名称或主机"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              onPressEnter={loadConnections}
            />
          </Col>
          <Col span={4}>
            <Select
              placeholder="数据库类型"
              value={filterType}
              onChange={setFilterType}
              allowClear
              style={{ width: '100%' }}
            >
              {Object.values(DatabaseType).map(type => (
                <Option key={type} value={type}>
                  {DATABASE_TYPE_CONFIGS[type]?.name || type}
                </Option>
              ))}
            </Select>
          </Col>
          <Col span={4}>
            <Select
              placeholder="连接状态"
              value={filterStatus}
              onChange={setFilterStatus}
              allowClear
              style={{ width: '100%' }}
            >
              <Option value={ConnectionStatus.CONNECTED}>已连接</Option>
              <Option value={ConnectionStatus.DISCONNECTED}>未连接</Option>
              <Option value={ConnectionStatus.CONNECTING}>连接中</Option>
              <Option value={ConnectionStatus.ERROR}>错误</Option>
            </Select>
          </Col>
          <Col span={8}>
            <Space>
              <Button onClick={loadConnections}>搜索</Button>
              {selectedRowKeys.length > 0 && (
                <>
                  <Button onClick={() => handleBatchOperation('test')}>
                    批量测试
                  </Button>
                  <Popconfirm
                    title={`确定要删除选中的 ${selectedRowKeys.length} 个连接吗？`}
                    onConfirm={() => handleBatchOperation('delete')}
                  >
                    <Button danger>批量删除</Button>
                  </Popconfirm>
                </>
              )}
            </Space>
          </Col>
        </Row>

        {/* 数据表格 */}
        <Table
          rowSelection={{
            selectedRowKeys,
            onChange: setSelectedRowKeys
          }}
          columns={columns}
          dataSource={connections}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: (page, pageSize) => {
              setPagination(prev => ({ ...prev, current: page, pageSize: pageSize || 10 }));
            }
          }}
        />
      </Card>

      {/* 新建/编辑模态框 */}
      <Modal
        title={editingConnection ? '编辑数据库连接' : '新建数据库连接'}
        open={isModalVisible}
        onCancel={() => {
          setIsModalVisible(false);
          setEditingConnection(null);
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            type: DatabaseType.MYSQL,
            port: 3306,
            ssl: false,
            connectionTimeout: 30,
            maxConnections: 10,
            isDefault: false
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="连接名称"
                rules={[{ required: true, message: '请输入连接名称' }]}
              >
                <Input placeholder="输入连接名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="type"
                label="数据库类型"
                rules={[{ required: true, message: '请选择数据库类型' }]}
              >
                <Select
                  placeholder="选择数据库类型"
                  onChange={(value: DatabaseType) => {
                    const config = DATABASE_TYPE_CONFIGS[value];
                    if (config) {
                      form.setFieldsValue({ port: config.defaultPort });
                    }
                  }}
                >
                  {Object.values(DatabaseType).map(type => (
                    <Option key={type} value={type}>
                      {DATABASE_TYPE_CONFIGS[type]?.name || type}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={16}>
              <Form.Item
                name="host"
                label="主机地址"
                rules={[{ required: true, message: '请输入主机地址' }]}
              >
                <Input placeholder="localhost" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="port"
                label="端口"
                rules={[{ required: true, message: '请输入端口' }]}
              >
                <InputNumber
                  min={1}
                  max={65535}
                  style={{ width: '100%' }}
                  placeholder="3306"
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="database"
                label="数据库名"
                rules={[{ required: true, message: '请输入数据库名' }]}
              >
                <Input placeholder="数据库名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="username"
                label="用户名"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input placeholder="用户名" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="password"
            label="密码"
            rules={[{ required: !editingConnection, message: '请输入密码' }]}
          >
            <Input.Password placeholder={editingConnection ? '留空表示不修改密码' : '密码'} />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={3} placeholder="连接描述（可选）" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="tags"
                label="标签"
              >
                <Select
                  mode="tags"
                  placeholder="添加标签"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="isDefault"
                label="设为默认连接"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="connectionTimeout"
                label="连接超时(秒)"
              >
                <InputNumber
                  min={1}
                  max={300}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="maxConnections"
                label="最大连接数"
              >
                <InputNumber
                  min={1}
                  max={100}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="ssl"
            label="启用SSL"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Divider />

          <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
            <Button
              onClick={() => {
                setIsModalVisible(false);
                setEditingConnection(null);
                form.resetFields();
              }}
            >
              取消
            </Button>
            <Button
              type="default"
              onClick={async () => {
                try {
                  const values = await form.validateFields();
                  const response = await DatabaseConnectionService.testConnection(values);
                  if (response.success && response.data) {
                    message.success(`连接测试成功 (${response.data.responseTime}ms)`);
                  } else {
                    message.error(response.message || '连接测试失败');
                  }
                } catch (error) {
                  message.error('请先完善连接信息');
                }
              }}
            >
              测试连接
            </Button>
            <Button type="primary" htmlType="submit">
              {editingConnection ? '更新' : '创建'}
            </Button>
          </Space>
        </Form>
      </Modal>
    </div>
  );
};

export default DatabaseConnectionManager;