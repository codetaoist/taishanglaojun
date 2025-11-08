import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Badge,
  Button,
  Table,
  Tag,
  Space,
  Alert,
  Spin,
  Typography,
  Tabs,
  List,
  Descriptions,
  Divider,
  Progress,
} from 'antd';
import {
  DatabaseOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  InfoCircleOutlined,
  CloudServerOutlined,
} from '@ant-design/icons';
import { vectorMonitorApi } from '../../services/vectorMonitorApi';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

interface VectorDatabaseStatus {
  connected: boolean;
  lastChecked: string;
  error?: string;
}

interface VectorDatabaseInfo {
  type: string;
  version: string;
}

interface CollectionStats {
  count: number;
}

interface VectorCollectionInfo {
  name: string;
  description?: string;
  dimension?: number;
  metricType?: string;
  vectorCount?: number;
}

const VectorMonitor: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<VectorDatabaseStatus | null>(null);
  const [info, setInfo] = useState<VectorDatabaseInfo | null>(null);
  const [collections, setCollections] = useState<VectorCollectionInfo[]>([]);
  const [activeTab, setActiveTab] = useState('status');
  const [error, setError] = useState<string | null>(null);

  // 获取数据库状态
  const fetchStatus = async () => {
    try {
      setError(null);
      const statusData = await vectorMonitorApi.getStatus();
      setStatus(statusData);
    } catch (err: any) {
      setError(err.message || '获取状态失败');
      console.error('获取状态失败:', err);
    }
  };

  // 获取数据库信息
  const fetchInfo = async () => {
    try {
      setError(null);
      const infoData = await vectorMonitorApi.getInfo();
      setInfo(infoData);
    } catch (err: any) {
      setError(err.message || '获取信息失败');
      console.error('获取信息失败:', err);
    }
  };

  // 获取集合列表
  const fetchCollections = async () => {
    try {
      setError(null);
      const collectionsData = await vectorMonitorApi.listVectorCollections();
      setCollections(collectionsData);
    } catch (err: any) {
      setError(err.message || '获取集合列表失败');
      console.error('获取集合列表失败:', err);
    }
  };

  // 连接数据库
  const connectDatabase = async () => {
    try {
      setLoading(true);
      setError(null);
      await vectorMonitorApi.connect();
      // 连接成功后刷新状态
      await fetchStatus();
    } catch (err: any) {
      setError(err.message || '连接数据库失败');
      console.error('连接数据库失败:', err);
    } finally {
      setLoading(false);
    }
  };

  // 刷新所有数据
  const refreshAll = async () => {
    setLoading(true);
    try {
      await Promise.all([
        fetchStatus(),
        fetchInfo(),
        fetchCollections(),
      ]);
    } finally {
      setLoading(false);
    }
  };

  // 初始化数据
  useEffect(() => {
    refreshAll();
  }, []);

  // 集合表格列定义
  const collectionColumns: ColumnsType<VectorCollectionInfo> = [
    {
      title: '集合名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: '维度',
      dataIndex: 'dimension',
      key: 'dimension',
      render: (dimension: number) => dimension || <Text type="secondary">-</Text>,
    },
    {
      title: '向量数量',
      dataIndex: 'vectorCount',
      key: 'vectorCount',
      render: (count: number) => count || <Text type="secondary">-</Text>,
    },
    {
      title: '度量类型',
      dataIndex: 'metricType',
      key: 'metricType',
      render: (type: string) => type ? <Tag color="blue">{type}</Tag> : <Text type="secondary">-</Text>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (description: string) => description || <Text type="secondary">-</Text>,
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <DatabaseOutlined /> 向量数据库监控
      </Title>
      <Paragraph type="secondary">
        监控向量数据库的状态、性能和集合信息
      </Paragraph>

      {error && (
        <Alert
          message="错误"
          description={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}

      <div style={{ marginBottom: 16 }}>
        <Space>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={refreshAll}
            loading={loading}
          >
            刷新数据
          </Button>
          {status && !status.connected && (
            <Button
              type="primary"
              icon={<CloudServerOutlined />}
              onClick={connectDatabase}
              loading={loading}
            >
              连接数据库
            </Button>
          )}
        </Space>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="状态概览" key="status">
          <Row gutter={[16, 16]}>
            <Col span={8}>
              <Card>
                <Statistic
                  title="连接状态"
                  value={status?.connected ? '已连接' : '未连接'}
                  prefix={
                    status?.connected ? (
                      <CheckCircleOutlined style={{ color: '#52c41a' }} />
                    ) : (
                      <ExclamationCircleOutlined style={{ color: '#f5222d' }} />
                    )
                  }
                />
                {status?.lastChecked && (
                  <div style={{ marginTop: 8 }}>
                    <Text type="secondary">最后检查时间: {new Date(status.lastChecked).toLocaleString()}</Text>
                  </div>
                )}
                {status?.error && (
                  <div style={{ marginTop: 8 }}>
                    <Text type="danger">错误: {status.error}</Text>
                  </div>
                )}
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic
                  title="数据库类型"
                  value={info?.type || '未知'}
                  prefix={<DatabaseOutlined />}
                />
                {info?.version && (
                  <div style={{ marginTop: 8 }}>
                    <Text type="secondary">版本: {info.version}</Text>
                  </div>
                )}
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic
                  title="集合数量"
                  value={collections.length}
                  prefix={<InfoCircleOutlined />}
                />
              </Card>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab="集合列表" key="collections">
          <Card>
            <Table
              columns={collectionColumns}
              dataSource={collections}
              rowKey="name"
              loading={loading}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </TabPane>

        <TabPane tab="详细信息" key="details">
          <Row gutter={[16, 16]}>
            <Col span={12}>
              <Card title="数据库状态" bordered={false}>
                {status ? (
                  <Descriptions column={1}>
                    <Descriptions.Item label="连接状态">
                      <Badge
                        status={status.connected ? 'success' : 'error'}
                        text={status.connected ? '已连接' : '未连接'}
                      />
                    </Descriptions.Item>
                    <Descriptions.Item label="最后检查时间">
                      {new Date(status.lastChecked).toLocaleString()}
                    </Descriptions.Item>
                    {status.error && (
                      <Descriptions.Item label="错误信息">
                        <Text type="danger">{status.error}</Text>
                      </Descriptions.Item>
                    )}
                  </Descriptions>
                ) : (
                  <div style={{ textAlign: 'center', padding: '20px' }}>
                    <Spin size="large" />
                  </div>
                )}
              </Card>
            </Col>
            <Col span={12}>
              <Card title="数据库信息" bordered={false}>
                {info ? (
                  <Descriptions column={1}>
                    <Descriptions.Item label="数据库类型">
                      {info.type}
                    </Descriptions.Item>
                    <Descriptions.Item label="版本">
                      {info.version}
                    </Descriptions.Item>
                  </Descriptions>
                ) : (
                  <div style={{ textAlign: 'center', padding: '20px' }}>
                    <Spin size="large" />
                  </div>
                )}
              </Card>
            </Col>
          </Row>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default VectorMonitor;