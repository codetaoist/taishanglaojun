import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Row, Col, Statistic, Progress, List, Avatar, Tag, Spin, Alert, Button } from 'antd';
import { 
  UserOutlined, 
  ProjectOutlined, 
  CheckCircleOutlined, 
  ClockCircleOutlined,
  BarChartOutlined,
  SettingOutlined,
  PlusOutlined,
  UserAddOutlined,
  FileTextOutlined
} from '@ant-design/icons';
import { Line } from '@ant-design/plots';
import dashboardService from '../services/dashboardService';
import type { DashboardStats, SystemMetrics, ActivityData, TrendData } from '../services/dashboardService';
import { 
  useRenderCount, 
  useMemoryMonitor, 
  useDebounce
} from '../utils/performanceOptimization';
// 性能浮窗已移除

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  // 性能监控
  const renderCount = useRenderCount('Dashboard');
  const memoryInfo = useMemoryMonitor(80); // 80MB阈值
  
  // 简化状态管理，不使用批量状态更新
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics | null>(null);
  const [recentActivities, setRecentActivities] = useState<ActivityData[]>([]);
  const [trendData, setTrendData] = useState<TrendData[]>([]);
  const [quickActions, setQuickActions] = useState<any[]>([]);

  // 使用useMemo缓存图表配置
  const chartConfig = useMemo(() => ({
    data: trendData,
    xField: 'date',
    yField: 'value',
    seriesField: 'category',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    point: {
      size: 3,
      shape: 'circle',
    },
    legend: {
      position: 'top' as const,
    },
    tooltip: {
      showMarkers: true,
    },
  }), [trendData]);

  // 使用useMemo缓存统计卡片数据
  const statisticsCards = useMemo(() => {
    if (!stats) return [];
    
    return [
      {
        title: '总用户数',
        value: stats.totalUsers,
        icon: <UserOutlined />,
        color: '#1890ff',
        suffix: '人',
        precision: 0,
      },
      {
        title: '活跃项目',
        value: stats.activeProjects,
        icon: <ProjectOutlined />,
        color: '#52c41a',
        suffix: '个',
        precision: 0,
      },
      {
        title: '完成任务',
        value: stats.completedTasks,
        icon: <CheckCircleOutlined />,
        color: '#faad14',
        suffix: '项',
        precision: 0,
      },
      {
        title: '系统运行时间',
        value: stats.systemUptime,
        icon: <ClockCircleOutlined />,
        color: '#722ed1',
        suffix: '小时',
        precision: 1,
      },
    ];
  }, [stats]);

  // 使用useMemo缓存系统指标数据
  const systemMetricsData = useMemo(() => {
    if (!systemMetrics && !stats) return [];
    
    return [
      {
        label: 'CPU使用率',
        value: systemMetrics?.cpu.usage || stats?.cpuUsage || 0,
        color: '#1890ff',
      },
      {
        label: '内存使用率',
        value: systemMetrics?.memory.usage || stats?.memoryUsage || 0,
        color: '#52c41a',
      },
      {
        label: '磁盘使用率',
        value: systemMetrics?.disk.usage || stats?.diskUsage || 0,
        color: '#faad14',
      },
    ];
  }, [systemMetrics, stats]);

  // 使用useCallback缓存数据加载函数
  const loadDashboardData = useCallback(async () => {
    try {
      console.log('🔄 开始加载Dashboard数据...');
      setLoading(true);
      setError(null);

      // 并行加载所有数据
      const [
        statsResponse,
        metricsResponse,
        activitiesResponse,
        trendsResponse,
        actionsResponse
      ] = await Promise.all([
        dashboardService.getDashboardStats(),
        dashboardService.getSystemMetrics(),
        dashboardService.getRecentActivities(5),
        dashboardService.getTrendData(7),
        dashboardService.getQuickActions()
      ]);

      console.log('📊 API响应:', {
        stats: statsResponse,
        metrics: metricsResponse,
        activities: activitiesResponse,
        trends: trendsResponse,
        actions: actionsResponse
      });

      // 更新各个状态
      if (statsResponse.success) {
        setStats(statsResponse.data);
      }

      if (metricsResponse.success) {
        setSystemMetrics(metricsResponse.data);
      }

      if (activitiesResponse.success) {
        setRecentActivities(activitiesResponse.data);
      }

      if (trendsResponse.success) {
        setTrendData(trendsResponse.data);
      }

      if (actionsResponse.success) {
        setQuickActions(actionsResponse.data);
      }

      console.log('✅ Dashboard数据加载完成');
      setLoading(false);

    } catch (err) {
      console.error('❌ Dashboard数据加载失败:', err);
      setLoading(false);
      setError(t('dashboard.error.loadFailed'));
    }
  }, []);

  // 使用useCallback缓存刷新函数
  const handleRefresh = useCallback(() => {
    loadDashboardData();
  }, [loadDashboardData]);

  // 使用useCallback缓存快速操作处理函数
  const handleQuickAction = useCallback((action: any) => {
    console.log('Quick action:', action);
    // 处理快速操作逻辑
  }, []);

  useEffect(() => {
    loadDashboardData();
  }, [loadDashboardData]);

  // 获取活动类型对应的图标
  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'user_login':
        return <UserOutlined style={{ color: '#1890ff' }} />;
      case 'task_completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'project_created':
        return <ProjectOutlined style={{ color: '#722ed1' }} />;
      case 'system_alert':
        return <ClockCircleOutlined style={{ color: '#faad14' }} />;
      default:
        return <BarChartOutlined style={{ color: '#8c8c8c' }} />;
    }
  };

  // 获取活动严重程度对应的标签颜色
  const getSeverityColor = (severity?: string) => {
    switch (severity) {
      case 'success':
        return 'green';
      case 'warning':
        return 'orange';
      case 'error':
        return 'red';
      case 'info':
      default:
        return 'blue';
    }
  };

  // 格式化时间
  const formatTime = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diff = now.getTime() - time.getTime();
    const minutes = Math.floor(diff / (1000 * 60));
    
    if (minutes < 1) return t('dashboard.time.justNow');
    if (minutes < 60) return t('dashboard.time.minutesAgo', { count: minutes });
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return t('dashboard.time.hoursAgo', { count: hours });
    
    const days = Math.floor(hours / 24);
    return t('dashboard.time.daysAgo', { count: days });
  };

  // 趋势图配置
  const trendConfig = {
    data: trendData,
    xField: 'date',
    yField: 'users',
    seriesField: 'type',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    point: {
      size: 5,
      shape: 'diamond',
    },
    label: {
      style: {
        fill: '#aaa',
      },
    },
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px' }}>
        <div style={{ textAlign: 'center' }}>
          <Spin size="large" />
          <div style={{ marginTop: '8px', color: '#8c8c8c' }}>{t('dashboard.loading.text')}</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <Alert
        message={t('dashboard.error.title')}
        description={error}
        type="error"
        showIcon
        action={
          <Button size="small" onClick={loadDashboardData}>
            {t('dashboard.error.retry')}
          </Button>
        }
      />
    );
  }

  return (
    <>
      <div style={{ padding: '24px' }}>
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title={t('dashboard.statistics.totalUsers')}
                value={stats?.totalUsers || 0}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
            <Statistic
              title={t('dashboard.statistics.activeUsers')}
              value={stats?.activeUsers || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title={t('dashboard.statistics.totalProjects')}
                value={stats?.totalProjects || 0}
                prefix={<ProjectOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
            <Statistic
              title={t('dashboard.statistics.completedTasks')}
              value={stats?.completedTasks || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        {/* 数据趋势图表 */}
        <Col xs={24} lg={16}>
          <Card title={t('dashboard.trend.title')} extra={<Button type="link">{t('dashboard.trend.viewDetails')}</Button>}>
            {trendData.length > 0 ? (
              <Line {...trendConfig} />
            ) : (
              <div style={{ textAlign: 'center', padding: '40px' }}>
                <Spin />
                <div style={{ marginTop: '8px', color: '#8c8c8c' }}>{t('dashboard.trend.loading')}</div>
              </div>
            )}
          </Card>
        </Col>

        {/* 快捷操作 */}
        <Col xs={24} lg={8}>
          <Card title={t('dashboard.quickActions.title')}>
            <Row gutter={[8, 8]}>
              {quickActions.map((action) => (
                <Col span={12} key={action.id}>
                  <Card 
                    size="small" 
                    hoverable
                    style={{ textAlign: 'center', cursor: 'pointer' }}
                    bodyStyle={{ padding: '12px' }}
                    onClick={() => {
                      // 这里可以添加导航逻辑
                      console.log('Navigate to:', action.action);
                    }}
                  >
                    <div style={{ color: action.color, fontSize: '24px', marginBottom: '8px' }}>
                      {action.icon === 'plus' && <PlusOutlined />}
                      {action.icon === 'user-add' && <UserAddOutlined />}
                      {action.icon === 'setting' && <SettingOutlined />}
                      {action.icon === 'file-text' && <FileTextOutlined />}
                    </div>
                    <div style={{ fontSize: '12px', fontWeight: 'bold' }}>{action.title}</div>
                    <div style={{ fontSize: '10px', color: '#8c8c8c' }}>{action.description}</div>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 最近活动 */}
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.recentActivities.title')} extra={<Button type="link">{t('dashboard.recentActivities.viewAll')}</Button>}>
            <List
              itemLayout="horizontal"
              dataSource={recentActivities}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={
                      item.user?.avatar ? (
                        <Avatar src={item.user.avatar} />
                      ) : (
                        <Avatar icon={getActivityIcon(item.type)} />
                      )
                    }
                    title={
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span>{item.title}</span>
                        {item.severity && (
                          <Tag color={getSeverityColor(item.severity)} size="small">
                            {item.severity}
                          </Tag>
                        )}
                      </div>
                    }
                    description={
                      <div>
                        <div>{item.description}</div>
                        <div style={{ fontSize: '12px', color: '#8c8c8c', marginTop: '4px' }}>
                          {formatTime(item.timestamp)}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* 系统状态 */}
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.systemStatus.title')}>
            <div style={{ marginBottom: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span>{t('dashboard.systemStatus.cpu')}</span>
                <span>{systemMetrics?.cpu.usage.toFixed(1) || stats?.cpuUsage.toFixed(1) || 0}%</span>
              </div>
              <Progress 
                percent={systemMetrics?.cpu.usage || stats?.cpuUsage || 0} 
                strokeColor="#1890ff"
                size="small"
              />
            </div>

            <div style={{ marginBottom: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span>{t('dashboard.systemStatus.memory')}</span>
                <span>{systemMetrics?.memory.usage.toFixed(1) || stats?.memoryUsage.toFixed(1) || 0}%</span>
              </div>
              <Progress 
                percent={systemMetrics?.memory.usage || stats?.memoryUsage || 0} 
                strokeColor="#52c41a"
                size="small"
              />
            </div>

            <div style={{ marginBottom: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span>{t('dashboard.systemStatus.disk')}</span>
                <span>{systemMetrics?.disk.usage.toFixed(1) || stats?.diskUsage.toFixed(1) || 0}%</span>
              </div>
              <Progress 
                percent={systemMetrics?.disk.usage || stats?.diskUsage || 0} 
                strokeColor="#faad14"
                size="small"
              />
            </div>

            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <span>{t('dashboard.systemStatus.networkLatency')}</span>
                <span>{systemMetrics?.network.latency || stats?.networkLatency || 0}ms</span>
              </div>
              <Progress 
                percent={Math.min((systemMetrics?.network.latency || stats?.networkLatency || 0) / 100 * 100, 100)} 
                strokeColor="#722ed1"
                size="small"
              />
            </div>
          </Card>
        </Col>
      </Row>
      </div>
    </>
  );
};

export default Dashboard;