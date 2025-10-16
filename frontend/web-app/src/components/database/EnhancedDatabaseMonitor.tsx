import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  message,
  Tag,
  Tooltip,
  Space,
  Alert,
  Tabs,
  Badge,
  Spin,
  Empty,
  Popconfirm,
  Timeline,
  Descriptions
} from 'antd';
import {
  DatabaseOutlined,
  CloudUploadOutlined,
  MonitorOutlined,
  LinkOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  WarningOutlined
} from '@ant-design/icons';
import { Line, Area } from '@ant-design/plots';
import enhancedDatabaseService, {
  type BackupStatus,
  type DatabaseMetrics,
  type ConnectionInfo
} from '../../services/enhancedDatabaseService';

const { TabPane } = Tabs;
const { Option } = Select;

interface EnhancedDatabaseMonitorProps {
  refreshInterval?: number;
}

const EnhancedDatabaseMonitor: React.FC<EnhancedDatabaseMonitorProps> = ({
  refreshInterval = 5000
}) => {
  // 状态管理
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('metrics');
  
  // 指标数据
  const [currentMetrics, setCurrentMetrics] = useState<DatabaseMetrics | null>(null);
  const [metricsHistory, setMetricsHistory] = useState<DatabaseMetrics[]>([]);
  const [healthStatus, setHealthStatus] = useState<any>(null);
  
  // 备份数据
  const [backups, setBackups] = useState<BackupStatus[]>([]);
  const [backupLoading, setBackupLoading] = useState(false);
  const [backupModalVisible, setBackupModalVisible] = useState(false);
  
  // 连接数据
  const [connections, setConnections] = useState<ConnectionInfo[]>([]);
  const [connectionStats, setConnectionStats] = useState<any>(null);
  
  // 表单
  const [backupForm] = Form.useForm();
  
  // 引用
  const metricsSubscription = useRef<(() => void) | null>(null);

  // 加载数据
  const loadMetrics = useCallback(async () => {
    try {
      const [metricsRes, healthRes] = await Promise.all([
        enhancedDatabaseService.getDatabaseMetrics(),
        enhancedDatabaseService.getDatabaseHealth()
      ]);

      if (metricsRes.success) {
        setCurrentMetrics(metricsRes.data);
      }

      if (healthRes.success) {
        setHealthStatus(healthRes.data);
      }
    } catch (error) {
      console.error('加载数据库指标失败:', error);
    }
  }, []);

  const loadMetricsHistory = useCallback(async () => {
    try {
      const response = await enhancedDatabaseService.getDatabaseMetricsHistory('1h', '5m');
      if (response.success) {
        setMetricsHistory(response.data);
      }
    } catch (error) {
      console.error('加载历史指标失败:', error);
    }
  }, []);

  const loadBackups = useCallback(async () => {
    setBackupLoading(true);
    try {
      const response = await enhancedDatabaseService.getBackups();
      if (response.success) {
        setBackups(response.data.backups);
      }
    } catch (error) {
      message.error('加载备份列表失败');
    } finally {
      setBackupLoading(false);
    }
  }, []);

  const loadConnections = useCallback(async () => {
    try {
      const [connectionsRes, statsRes] = await Promise.all([
        enhancedDatabaseService.getActiveConnections(),
        enhancedDatabaseService.getConnectionPoolStats()
      ]);

      if (connectionsRes.success) {
        setConnections(connectionsRes.data.connections);
      }

      if (statsRes.success) {
        setConnectionStats(statsRes.data);
      }
    } catch (error) {
      console.error('加载连接信息失败:', error);
    }
  }, []);

  // 备份操作
  const handleCreateBackup = async (values: any) => {
    try {
      setLoading(true);
      const response = await enhancedDatabaseService.createBackup(values.name, {
        description: values.description,
        backup_type: values.backup_type,
        compression: values.compression,
        encryption: values.encryption
      });

      if (response.success) {
        message.success('备份任务已启动');
        setBackupModalVisible(false);
        backupForm.resetFields();
        loadBackups();
      } else {
        message.error(response.message || '创建备份失败');
      }
    } catch (error) {
      message.error('创建备份失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteBackup = async (id: string) => {
    try {
      const response = await enhancedDatabaseService.deleteBackup(id);
      if (response.success) {
        message.success('删除备份成功');
        loadBackups();
      } else {
        message.error(response.message || '删除备份失败');
      }
    } catch (error) {
      message.error('删除备份失败');
    }
  };

  const handleRestoreBackup = async (id: string) => {
    try {
      const response = await enhancedDatabaseService.restoreBackup(id);
      if (response.success) {
        message.success('恢复任务已启动');
      } else {
        message.error(response.message || '恢复备份失败');
      }
    } catch (error) {
      message.error('恢复备份失败');
    }
  };

  // 连接操作
  const handleTestAllConnections = async () => {
    try {
      setLoading(true);
      const response = await enhancedDatabaseService.testAllConnections();
      if (response.success) {
        message.success('连接测试完成');
        loadConnections();
      } else {
        message.error(response.message || '连接测试失败');
      }
    } catch (error) {
      message.error('连接测试失败');
    } finally {
      setLoading(false);
    }
  };

  // 数据库优化
  const handleOptimizeDatabase = async () => {
    try {
      setLoading(true);
      const response = await enhancedDatabaseService.optimizeDatabase({
        analyze_tables: true,
        rebuild_indexes: true,
        cleanup_logs: true
      });

      if (response.success) {
        message.success('数据库优化已启动');
      } else {
        message.error(response.message || '数据库优化失败');
      }
    } catch (error) {
      message.error('数据库优化失败');
    } finally {
      setLoading(false);
    }
  };

  // 生命周期
  useEffect(() => {
    loadMetrics();
    loadMetricsHistory();
    loadBackups();
    loadConnections();

    // 订阅实时指标
    metricsSubscription.current = enhancedDatabaseService.subscribeToMetrics((metrics) => {
      setCurrentMetrics(metrics);
    });

    return () => {
      if (metricsSubscription.current) {
        metricsSubscription.current();
      }
    };
  }, [loadMetrics, loadMetricsHistory, loadBackups, loadConnections]);

  // 渲染健康状态
  const renderHealthStatus = () => {
    if (!healthStatus) return null;

    const { overall_health, health_score, issues } = healthStatus;
    
    let statusColor = 'success';
    let statusText = '健康';
    
    if (overall_health === 'warning') {
      statusColor = 'warning';
      statusText = '警告';
    } else if (overall_health === 'critical') {
      statusColor = 'error';
      statusText = '严重';
    }

    return (
      <Card title="数据库健康状态" size="small">
        <Row gutter={16}>
          <Col span={12}>
            <Statistic
              title="健康评分"
              value={health_score}
              suffix="/ 100"
              valueStyle={{ color: statusColor === 'success' ? '#3f8600' : statusColor === 'warning' ? '#cf1322' : '#cf1322' }}
            />
          </Col>
          <Col span={12}>
            <Statistic
              title="状态"
              value={statusText}
              valueStyle={{ color: statusColor === 'success' ? '#3f8600' : statusColor === 'warning' ? '#cf1322' : '#cf1322' }}
            />
          </Col>
        </Row>
        
        {issues && issues.length > 0 && (
          <div style={{ marginTop: 16 }}>
            <h4>发现的问题:</h4>
            {issues.map((issue, index) => (
              <Alert
                key={index}
                message={issue.message}
                description={issue.recommendation}
                type={issue.severity === 'critical' ? 'error' : issue.severity === 'high' ? 'warning' : 'info'}
                showIcon
                style={{ marginBottom: 8 }}
              />
            ))}
          </div>
        )}
      </Card>
    );
  };

  // 渲染性能指标
  const renderMetrics = () => {
    if (!currentMetrics) return <Spin size="large" />;

    const { connections, performance, storage, memory } = currentMetrics;

    return (
      <div>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃连接"
                value={connections.active}
                suffix={`/ ${connections.total}`}
                prefix={<DatabaseOutlined />}
              />
              <Progress
                percent={(connections.active / connections.total) * 100}
                size="small"
                status={connections.active / connections.total > 0.8 ? 'exception' : 'normal'}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="查询/秒"
                value={performance.queries_per_second}
                precision={1}
                prefix={<MonitorOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="缓存命中率"
                value={performance.cache_hit_ratio}
                suffix="%"
                precision={1}
                valueStyle={{ color: performance.cache_hit_ratio > 90 ? '#3f8600' : '#cf1322' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="存储使用率"
                value={(storage.used_size / storage.total_size) * 100}
                suffix="%"
                precision={1}
              />
              <Progress
                percent={(storage.used_size / storage.total_size) * 100}
                size="small"
                status={(storage.used_size / storage.total_size) > 0.8 ? 'exception' : 'normal'}
              />
            </Card>
          </Col>
        </Row>

        {renderHealthStatus()}

        {metricsHistory.length > 0 && (
          <Card title="性能趋势" style={{ marginTop: 16 }}>
            <Area
              data={metricsHistory.map(m => ({
                time: new Date(m.timestamp).toLocaleTimeString(),
                value: m.performance.queries_per_second,
                type: '查询/秒'
              }))}
              xField="time"
              yField="value"
              seriesField="type"
              height={200}
              smooth
            />
          </Card>
        )}
      </div>
    );
  };

  // 渲染备份管理
  const renderBackups = () => {
    const columns = [
      {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '类型',
        dataIndex: 'backup_type',
        key: 'backup_type',
        render: (type: string) => (
          <Tag color={type === 'full' ? 'blue' : type === 'incremental' ? 'green' : 'orange'}>
            {type === 'full' ? '完整' : type === 'incremental' ? '增量' : '差异'}
          </Tag>
        ),
      },
      {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        render: (status: string, record: BackupStatus) => {
          const statusConfig = {
            pending: { color: 'default', icon: <ClockCircleOutlined />, text: '等待中' },
            running: { color: 'processing', icon: <PlayCircleOutlined />, text: '进行中' },
            completed: { color: 'success', icon: <CheckCircleOutlined />, text: '已完成' },
            failed: { color: 'error', icon: <ExclamationCircleOutlined />, text: '失败' },
            cancelled: { color: 'default', icon: <PauseCircleOutlined />, text: '已取消' }
          };

          const config = statusConfig[status as keyof typeof statusConfig];
          
          return (
            <div>
              <Tag color={config.color} icon={config.icon}>
                {config.text}
              </Tag>
              {status === 'running' && (
                <Progress percent={record.progress} size="small" style={{ width: 100, marginLeft: 8 }} />
              )}
            </div>
          );
        },
      },
      {
        title: '大小',
        dataIndex: 'file_size',
        key: 'file_size',
        render: (size: number) => size ? `${(size / 1024 / 1024).toFixed(2)} MB` : '-',
      },
      {
        title: '创建时间',
        dataIndex: 'created_at',
        key: 'created_at',
        render: (time: string) => new Date(time).toLocaleString(),
      },
      {
        title: '操作',
        key: 'actions',
        render: (_, record: BackupStatus) => (
          <Space>
            {record.status === 'completed' && (
              <Button
                type="link"
                size="small"
                onClick={() => handleRestoreBackup(record.id)}
              >
                恢复
              </Button>
            )}
            <Popconfirm
              title="确定要删除这个备份吗？"
              onConfirm={() => handleDeleteBackup(record.id)}
            >
              <Button type="link" size="small" danger>
                删除
              </Button>
            </Popconfirm>
          </Space>
        ),
      },
    ];

    return (
      <div>
        <div style={{ marginBottom: 16 }}>
          <Space>
            <Button
              type="primary"
              icon={<CloudUploadOutlined />}
              onClick={() => setBackupModalVisible(true)}
            >
              创建备份
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={loadBackups}
              loading={backupLoading}
            >
              刷新
            </Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={backups}
          rowKey="id"
          loading={backupLoading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
        />
      </div>
    );
  };

  // 渲染连接监控
  const renderConnections = () => {
    const columns = [
      {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '类型',
        dataIndex: 'type',
        key: 'type',
        render: (type: string) => <Tag>{type.toUpperCase()}</Tag>,
      },
      {
        title: '地址',
        key: 'address',
        render: (_, record: ConnectionInfo) => `${record.host}:${record.port}`,
      },
      {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        render: (status: string) => {
          const statusConfig = {
            connected: { color: 'success', text: '已连接' },
            disconnected: { color: 'default', text: '未连接' },
            error: { color: 'error', text: '错误' },
            testing: { color: 'processing', text: '测试中' }
          };

          const config = statusConfig[status as keyof typeof statusConfig];
          return <Badge status={config.color as any} text={config.text} />;
        },
      },
      {
        title: '响应时间',
        dataIndex: 'response_time',
        key: 'response_time',
        render: (time: number) => `${time}ms`,
      },
      {
        title: '健康评分',
        dataIndex: 'health_score',
        key: 'health_score',
        render: (score: number) => (
          <Progress
            percent={score}
            size="small"
            status={score > 80 ? 'success' : score > 60 ? 'normal' : 'exception'}
            style={{ width: 100 }}
          />
        ),
      },
      {
        title: '连接池',
        key: 'pool',
        render: (_, record: ConnectionInfo) => (
          <Tooltip title={`活跃: ${record.connection_pool.active_connections}, 空闲: ${record.connection_pool.idle_connections}`}>
            <Progress
              percent={(record.connection_pool.active_connections / record.connection_pool.max_connections) * 100}
              size="small"
              style={{ width: 80 }}
            />
          </Tooltip>
        ),
      },
    ];

    return (
      <div>
        <div style={{ marginBottom: 16 }}>
          <Space>
            <Button
              type="primary"
              icon={<LinkOutlined />}
              onClick={handleTestAllConnections}
              loading={loading}
            >
              测试所有连接
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={loadConnections}
            >
              刷新
            </Button>
          </Space>
        </div>

        {connectionStats && (
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title="总连接数"
                  value={connectionStats.total_connections}
                  prefix={<DatabaseOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="活跃连接"
                  value={connectionStats.active_connections}
                  valueStyle={{ color: '#3f8600' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="空闲连接"
                  value={connectionStats.idle_connections}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title="等待连接"
                  value={connectionStats.waiting_connections}
                  valueStyle={{ color: connectionStats.waiting_connections > 0 ? '#cf1322' : '#3f8600' }}
                />
              </Card>
            </Col>
          </Row>
        )}

        <Table
          columns={columns}
          dataSource={connections}
          rowKey="id"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
          }}
        />
      </div>
    );
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Button
            type="primary"
            icon={<MonitorOutlined />}
            onClick={handleOptimizeDatabase}
            loading={loading}
          >
            优化数据库
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              loadMetrics();
              loadBackups();
              loadConnections();
            }}
          >
            刷新所有
          </Button>
        </Space>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="性能监控" key="metrics">
          {renderMetrics()}
        </TabPane>
        <TabPane tab="备份管理" key="backups">
          {renderBackups()}
        </TabPane>
        <TabPane tab="连接监控" key="connections">
          {renderConnections()}
        </TabPane>
      </Tabs>

      {/* 创建备份模态框 */}
      <Modal
        title="创建数据库备份"
        open={backupModalVisible}
        onCancel={() => setBackupModalVisible(false)}
        footer={null}
      >
        <Form
          form={backupForm}
          layout="vertical"
          onFinish={handleCreateBackup}
        >
          <Form.Item
            name="name"
            label="备份名称"
            rules={[{ required: true, message: '请输入备份名称' }]}
          >
            <Input placeholder="输入备份名称" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="输入备份描述" rows={3} />
          </Form.Item>

          <Form.Item
            name="backup_type"
            label="备份类型"
            initialValue="full"
          >
            <Select>
              <Option value="full">完整备份</Option>
              <Option value="incremental">增量备份</Option>
              <Option value="differential">差异备份</Option>
            </Select>
          </Form.Item>

          <Form.Item name="compression" label="启用压缩" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>

          <Form.Item name="encryption" label="启用加密" valuePropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                创建备份
              </Button>
              <Button onClick={() => setBackupModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default EnhancedDatabaseMonitor;