import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Row,
  Col,
  Badge,
  Progress,
  Statistic,
  List,
  Avatar,
  Space,
  Tag,
  Button,
  Tooltip,
  Alert,
  Typography,
  Spin,
  Empty
} from 'antd';
import {
  DatabaseOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  SyncOutlined,
  ReloadOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined
} from '@ant-design/icons';
import { DatabaseConnectionService } from '../../services/databaseConnectionService';
import type { DatabaseConnectionStatus, DatabaseConnectionStats } from '../../types/database';
import { ConnectionStatus, DatabaseType, DATABASE_TYPE_CONFIGS } from '../../types/database';

const { Title, Text } = Typography;

interface ConnectionStatusMonitorProps {
  refreshInterval?: number; // 刷新间隔，毫秒
  autoRefresh?: boolean; // 是否自动刷新
  className?: string;
}

const ConnectionStatusMonitor: React.FC<ConnectionStatusMonitorProps> = ({
  refreshInterval = 30000, // 默认30秒刷新一次
  autoRefresh = true,
  className
}) => {
  const [connectionStatuses, setConnectionStatuses] = useState<DatabaseConnectionStatus[]>([]);
  const [stats, setStats] = useState<DatabaseConnectionStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [refreshing, setRefreshing] = useState<Set<string>>(new Set());

  // 加载连接状态
  const loadConnectionStatuses = useCallback(async () => {
    try {
      setLoading(true);
      const response = await DatabaseConnectionService.getConnectionsStatus();
      if (response.success && response.data) {
        setConnectionStatuses(response.data);
        setLastUpdated(new Date());
      }
    } catch (error) {
      console.error('加载连接状态失败:', error);
    } finally {
      setLoading(false);
    }
  }, []);

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

  // 刷新单个连接状态
  const refreshConnectionStatus = async (connectionId: string) => {
    setRefreshing(prev => new Set(prev).add(connectionId));
    try {
      const response = await DatabaseConnectionService.refreshConnectionStatus(connectionId);
      if (response.success && response.data) {
        setConnectionStatuses(prev => 
          prev.map(status => 
            status.id === connectionId ? response.data! : status
          )
        );
      }
    } catch (error) {
      console.error('刷新连接状态失败:', error);
    } finally {
      setRefreshing(prev => {
        const newSet = new Set(prev);
        newSet.delete(connectionId);
        return newSet;
      });
    }
  };

  // 初始化加载
  useEffect(() => {
    loadConnectionStatuses();
    loadStats();
  }, [loadConnectionStatuses, loadStats]);

  // 自动刷新
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      loadConnectionStatuses();
      loadStats();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, loadConnectionStatuses, loadStats]);

  // 获取状态颜色
  const getStatusColor = (status: ConnectionStatus) => {
    switch (status) {
      case ConnectionStatus.CONNECTED:
        return '#52c41a';
      case ConnectionStatus.DISCONNECTED:
        return '#8c8c8c';
      case ConnectionStatus.CONNECTING:
        return '#1890ff';
      case ConnectionStatus.ERROR:
        return '#ff4d4f';
      default:
        return '#faad14';
    }
  };

  // 获取状态图标
  const getStatusIcon = (status: ConnectionStatus) => {
    switch (status) {
      case ConnectionStatus.CONNECTED:
        return <CheckCircleOutlined style={{ color: getStatusColor(status) }} />;
      case ConnectionStatus.DISCONNECTED:
        return <CloseCircleOutlined style={{ color: getStatusColor(status) }} />;
      case ConnectionStatus.CONNECTING:
        return <SyncOutlined spin style={{ color: getStatusColor(status) }} />;
      case ConnectionStatus.ERROR:
        return <ExclamationCircleOutlined style={{ color: getStatusColor(status) }} />;
      default:
        return <WarningOutlined style={{ color: getStatusColor(status) }} />;
    }
  };

  // 获取状态文本
  const getStatusText = (status: ConnectionStatus) => {
    switch (status) {
      case ConnectionStatus.CONNECTED:
        return '已连接';
      case ConnectionStatus.DISCONNECTED:
        return '未连接';
      case ConnectionStatus.CONNECTING:
        return '连接中';
      case ConnectionStatus.ERROR:
        return '连接错误';
      default:
        return '未知状态';
    }
  };

  // 计算健康度
  const calculateHealthScore = () => {
    if (!stats || stats.totalConnections === 0) return 0;
    return Math.round((stats.activeConnections / stats.totalConnections) * 100);
  };

  // 获取健康度颜色
  const getHealthColor = (score: number) => {
    if (score >= 80) return '#52c41a';
    if (score >= 60) return '#faad14';
    return '#ff4d4f';
  };

  // 按状态分组连接
  const groupedConnections = connectionStatuses.reduce((acc, connection) => {
    if (!acc[connection.status]) {
      acc[connection.status] = [];
    }
    acc[connection.status].push(connection);
    return acc;
  }, {} as Record<ConnectionStatus, DatabaseConnectionStatus[]>);

  return (
    <div className={className}>
      {/* 总体状态概览 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="连接健康度"
              value={calculateHealthScore()}
              suffix="%"
              valueStyle={{ color: getHealthColor(calculateHealthScore()) }}
              prefix={<ThunderboltOutlined />}
            />
            <Progress
              percent={calculateHealthScore()}
              strokeColor={getHealthColor(calculateHealthScore())}
              showInfo={false}
              size="small"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃连接"
              value={stats?.activeConnections || 0}
              suffix={`/ ${stats?.totalConnections || 0}`}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="平均响应时间"
              value={stats?.averageResponseTime || 0}
              suffix="ms"
              valueStyle={{ 
                color: (stats?.averageResponseTime || 0) < 100 ? '#52c41a' : 
                       (stats?.averageResponseTime || 0) < 500 ? '#faad14' : '#ff4d4f'
              }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="最后更新"
              value={lastUpdated ? lastUpdated.toLocaleTimeString() : '-'}
              prefix={<ReloadOutlined />}
            />
            <Button
              type="link"
              size="small"
              icon={<ReloadOutlined />}
              onClick={() => {
                loadConnectionStatuses();
                loadStats();
              }}
              loading={loading}
            >
              立即刷新
            </Button>
          </Card>
        </Col>
      </Row>

      {/* 状态分布 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={24}>
            <Card title="连接状态分布">
              <Row gutter={16}>
                {Object.entries(stats.connectionsByStatus).map(([status, count]) => (
                  <Col span={6} key={status}>
                    <Card size="small">
                      <Space>
                        {getStatusIcon(status as ConnectionStatus)}
                        <div>
                          <div style={{ fontSize: '24px', fontWeight: 'bold' }}>
                            {count}
                          </div>
                          <div style={{ color: '#8c8c8c' }}>
                            {getStatusText(status as ConnectionStatus)}
                          </div>
                        </div>
                      </Space>
                    </Card>
                  </Col>
                ))}
              </Row>
            </Card>
          </Col>
        </Row>
      )}

      {/* 连接详细状态 */}
      <Card
        title={
          <Space>
            <DatabaseOutlined />
            <span>连接状态详情</span>
            <Badge count={connectionStatuses.length} showZero />
          </Space>
        }
        extra={
          <Space>
            <Text type="secondary">
              自动刷新: {autoRefresh ? `${refreshInterval / 1000}s` : '关闭'}
            </Text>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                loadConnectionStatuses();
                loadStats();
              }}
              loading={loading}
            >
              刷新全部
            </Button>
          </Space>
        }
      >
        {loading && connectionStatuses.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
          </div>
        ) : connectionStatuses.length === 0 ? (
          <Empty description="暂无数据库连接" />
        ) : (
          <List
            dataSource={connectionStatuses}
            renderItem={(connection) => (
              <List.Item
                actions={[
                  <Tooltip title="刷新状态">
                    <Button
                      type="text"
                      icon={<ReloadOutlined />}
                      loading={refreshing.has(connection.id)}
                      onClick={() => refreshConnectionStatus(connection.id)}
                    />
                  </Tooltip>
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <Avatar
                      icon={<DatabaseOutlined />}
                      style={{ backgroundColor: getStatusColor(connection.status) }}
                    />
                  }
                  title={
                    <Space>
                      <span>{connection.id}</span>
                      {getStatusIcon(connection.status)}
                      <Tag color={connection.status === ConnectionStatus.CONNECTED ? 'success' : 
                                 connection.status === ConnectionStatus.ERROR ? 'error' : 'default'}>
                        {getStatusText(connection.status)}
                      </Tag>
                    </Space>
                  }
                  description={
                    <Space direction="vertical" size="small" style={{ width: '100%' }}>
                      <Space>
                        <Text type="secondary">最后检查:</Text>
                        <Text>{new Date(connection.lastChecked).toLocaleString()}</Text>
                      </Space>
                      {connection.responseTime && (
                        <Space>
                          <Text type="secondary">响应时间:</Text>
                          <Text>{connection.responseTime}ms</Text>
                        </Space>
                      )}
                      {connection.serverVersion && (
                        <Space>
                          <Text type="secondary">服务器版本:</Text>
                          <Text code>{connection.serverVersion}</Text>
                        </Space>
                      )}
                      {connection.databaseSize && (
                        <Space>
                          <Text type="secondary">数据库大小:</Text>
                          <Text>{connection.databaseSize}</Text>
                        </Space>
                      )}
                      {connection.activeConnections !== undefined && (
                        <Space>
                          <Text type="secondary">活跃连接数:</Text>
                          <Text>{connection.activeConnections}</Text>
                        </Space>
                      )}
                      {connection.errorMessage && (
                        <Alert
                          message="连接错误"
                          description={connection.errorMessage}
                          type="error"
                          size="small"
                          showIcon
                        />
                      )}
                    </Space>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  );
};

export default ConnectionStatusMonitor;