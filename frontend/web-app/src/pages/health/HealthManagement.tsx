import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Typography, 
  Space, 
  Statistic, 
  Progress, 
  Timeline, 
  Avatar, 
  Tag, 
  Alert,
  Divider,
  List,
  Badge,
  Tooltip
} from 'antd';
import { 
  HeartOutlined, 
  DashboardOutlined, 
  BarChartOutlined, 
  BulbOutlined,
  FileTextOutlined,
  PlusOutlined,
  SettingOutlined,
  CalendarOutlined,
  TrophyOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  UserOutlined,
  MedicineBoxOutlined,
  ThunderboltOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Paragraph, Text } = Typography;

interface HealthMetric {
  id: string;
  name: string;
  value: number;
  unit: string;
  status: 'normal' | 'warning' | 'danger';
  trend: 'up' | 'down' | 'stable';
  lastUpdate: Date;
}

interface HealthAlert {
  id: string;
  type: 'warning' | 'info' | 'success';
  title: string;
  description: string;
  time: Date;
  read: boolean;
}

interface HealthGoal {
  id: string;
  title: string;
  target: number;
  current: number;
  unit: string;
  deadline: Date;
  category: string;
}

const HealthManagement: React.FC = () => {
  const navigate = useNavigate();
  const [healthMetrics, setHealthMetrics] = useState<HealthMetric[]>([]);
  const [healthAlerts, setHealthAlerts] = useState<HealthAlert[]>([]);
  const [healthGoals, setHealthGoals] = useState<HealthGoal[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 模拟数据加载
    setTimeout(() => {
      setHealthMetrics([
        {
          id: '1',
          name: '心率',
          value: 72,
          unit: 'bpm',
          status: 'normal',
          trend: 'stable',
          lastUpdate: new Date()
        },
        {
          id: '2',
          name: '血压',
          value: 120,
          unit: 'mmHg',
          status: 'normal',
          trend: 'down',
          lastUpdate: new Date()
        },
        {
          id: '3',
          name: '体重',
          value: 65.5,
          unit: 'kg',
          status: 'normal',
          trend: 'down',
          lastUpdate: new Date()
        },
        {
          id: '4',
          name: '睡眠质量',
          value: 85,
          unit: '%',
          status: 'normal',
          trend: 'up',
          lastUpdate: new Date()
        }
      ]);

      setHealthAlerts([
        {
          id: '1',
          type: 'warning',
          title: '运动量不足提醒',
          description: '今日步数仅3,200步，建议增加运动量',
          time: new Date(),
          read: false
        },
        {
          id: '2',
          type: 'info',
          title: '体检提醒',
          description: '距离下次体检还有7天，请及时预约',
          time: new Date(Date.now() - 2 * 60 * 60 * 1000),
          read: false
        },
        {
          id: '3',
          type: 'success',
          title: '目标达成',
          description: '恭喜！本周运动目标已完成',
          time: new Date(Date.now() - 24 * 60 * 60 * 1000),
          read: true
        }
      ]);

      setHealthGoals([
        {
          id: '1',
          title: '减重目标',
          target: 5,
          current: 2.5,
          unit: 'kg',
          deadline: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
          category: '体重管理'
        },
        {
          id: '2',
          title: '每日步数',
          target: 10000,
          current: 7500,
          unit: '步',
          deadline: new Date(),
          category: '运动健身'
        },
        {
          id: '3',
          title: '睡眠时长',
          target: 8,
          current: 7.2,
          unit: '小时',
          deadline: new Date(),
          category: '睡眠管理'
        }
      ]);

      setLoading(false);
    }, 1000);
  }, []);

  // 功能模块配置
  const modules = [
    {
      key: 'monitoring',
      title: '健康监测',
      description: '实时监测各项健康指标',
      icon: <DashboardOutlined style={{ fontSize: '32px', color: '#1890ff' }} />,
      path: '/health/monitoring',
      color: '#1890ff',
      stats: '4项指标'
    },
    {
      key: 'analysis',
      title: '健康分析',
      description: 'AI智能分析健康趋势',
      icon: <BarChartOutlined style={{ fontSize: '32px', color: '#52c41a' }} />,
      path: '/health/analysis',
      color: '#52c41a',
      stats: '7天趋势'
    },
    {
      key: 'advice',
      title: '健康建议',
      description: '个性化健康改善建议',
      icon: <BulbOutlined style={{ fontSize: '32px', color: '#faad14' }} />,
      path: '/health/advice',
      color: '#faad14',
      stats: '3条新建议'
    },
    {
      key: 'records',
      title: '健康档案',
      description: '完整的健康记录管理',
      icon: <FileTextOutlined style={{ fontSize: '32px', color: '#722ed1' }} />,
      path: '/health/records',
      color: '#722ed1',
      stats: '12条记录'
    }
  ];

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'normal': return '#52c41a';
      case 'warning': return '#faad14';
      case 'danger': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取趋势图标
  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return '↗️';
      case 'down': return '↘️';
      case 'stable': return '➡️';
      default: return '➡️';
    }
  };

  // 渲染健康指标卡片
  const renderHealthMetrics = () => (
    <Card title="健康指标概览" extra={<Button icon={<EyeOutlined />} onClick={() => navigate('/health/monitoring')}>查看详情</Button>}>
      <Row gutter={[16, 16]}>
        {healthMetrics.map((metric) => (
          <Col xs={12} sm={6} key={metric.id}>
            <Card size="small" style={{ textAlign: 'center' }}>
              <Statistic
                title={metric.name}
                value={metric.value}
                suffix={metric.unit}
                valueStyle={{ color: getStatusColor(metric.status) }}
              />
              <div style={{ marginTop: '8px' }}>
                <Space>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {getTrendIcon(metric.trend)}
                  </Text>
                  <Badge 
                    color={getStatusColor(metric.status)} 
                    text={metric.status === 'normal' ? '正常' : metric.status === 'warning' ? '注意' : '异常'}
                  />
                </Space>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </Card>
  );

  // 渲染功能模块
  const renderModules = () => (
    <Row gutter={[24, 24]}>
      {modules.map((module) => (
        <Col xs={24} sm={12} lg={6} key={module.key}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            onClick={() => navigate(module.path)}
          >
            <div style={{ marginBottom: '16px' }}>
              {module.icon}
            </div>
            <Title level={4} style={{ marginBottom: '8px' }}>
              {module.title}
            </Title>
            <Paragraph style={{ color: '#666', fontSize: '14px', marginBottom: '16px' }}>
              {module.description}
            </Paragraph>
            <Tag color={module.color}>{module.stats}</Tag>
          </Card>
        </Col>
      ))}
    </Row>
  );

  // 渲染健康提醒
  const renderHealthAlerts = () => (
    <Card 
      title="健康提醒" 
      extra={
        <Space>
          <Badge count={healthAlerts.filter(alert => !alert.read).length} />
          <Button size="small">全部标记已读</Button>
        </Space>
      }
    >
      <List
        size="small"
        dataSource={healthAlerts.slice(0, 3)}
        renderItem={(alert) => (
          <List.Item
            style={{ 
              opacity: alert.read ? 0.6 : 1,
              background: alert.read ? 'transparent' : '#fafafa',
              padding: '12px',
              borderRadius: '6px',
              marginBottom: '8px'
            }}
          >
            <List.Item.Meta
              avatar={
                <Avatar 
                  icon={
                    alert.type === 'warning' ? <WarningOutlined /> :
                    alert.type === 'success' ? <CheckCircleOutlined /> :
                    <ClockCircleOutlined />
                  }
                  style={{ 
                    backgroundColor: 
                      alert.type === 'warning' ? '#faad14' :
                      alert.type === 'success' ? '#52c41a' :
                      '#1890ff'
                  }}
                />
              }
              title={alert.title}
              description={
                <Space direction="vertical" size="small">
                  <Text type="secondary">{alert.description}</Text>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {alert.time.toLocaleString()}
                  </Text>
                </Space>
              }
            />
          </List.Item>
        )}
      />
      {healthAlerts.length > 3 && (
        <div style={{ textAlign: 'center', marginTop: '16px' }}>
          <Button type="link">查看全部提醒</Button>
        </div>
      )}
    </Card>
  );

  // 渲染健康目标
  const renderHealthGoals = () => (
    <Card title="健康目标" extra={<Button icon={<PlusOutlined />} size="small">添加目标</Button>}>
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        {healthGoals.map((goal) => {
          const progress = (goal.current / goal.target) * 100;
          return (
            <div key={goal.id}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <Space>
                  <Text strong>{goal.title}</Text>
                  <Tag size="small">{goal.category}</Tag>
                </Space>
                <Text type="secondary">
                  {goal.current}/{goal.target} {goal.unit}
                </Text>
              </div>
              <Progress 
                percent={Math.min(progress, 100)} 
                strokeColor={progress >= 100 ? '#52c41a' : progress >= 75 ? '#1890ff' : '#faad14'}
                size="small"
              />
            </div>
          );
        })}
      </Space>
    </Card>
  );

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <HeartOutlined style={{ color: '#ff4d4f', marginRight: '8px' }} />
          健康管理
        </Title>
        <Paragraph>
          全方位健康监测与管理，让科技守护您的健康生活
        </Paragraph>
      </div>

      {/* 快速操作栏 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space size="large">
              <Statistic title="健康评分" value={85} suffix="分" valueStyle={{ color: '#52c41a' }} />
              <Statistic title="连续监测" value={15} suffix="天" valueStyle={{ color: '#1890ff' }} />
              <Statistic title="目标完成" value={67} suffix="%" valueStyle={{ color: '#faad14' }} />
            </Space>
          </Col>
          <Col>
            <Space>
              <Button icon={<PlusOutlined />} type="primary">
                添加数据
              </Button>
              <Button icon={<CalendarOutlined />}>
                预约体检
              </Button>
              <Button icon={<SettingOutlined />}>
                设置
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 主要内容区域 */}
      <Row gutter={[24, 24]}>
        {/* 左侧内容 */}
        <Col xs={24} lg={16}>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            {/* 健康指标概览 */}
            {renderHealthMetrics()}

            {/* 功能模块 */}
            <Card title="功能模块">
              {renderModules()}
            </Card>
          </Space>
        </Col>

        {/* 右侧侧边栏 */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            {/* 健康提醒 */}
            {renderHealthAlerts()}

            {/* 健康目标 */}
            {renderHealthGoals()}

            {/* 今日建议 */}
            <Card title="今日建议" extra={<BulbOutlined />}>
              <Timeline size="small">
                <Timeline.Item color="green">
                  <Text>早晨：进行30分钟有氧运动</Text>
                </Timeline.Item>
                <Timeline.Item color="blue">
                  <Text>中午：补充维生素D</Text>
                </Timeline.Item>
                <Timeline.Item color="orange">
                  <Text>晚上：保证8小时充足睡眠</Text>
                </Timeline.Item>
              </Timeline>
              <Button type="link" style={{ padding: 0, marginTop: '8px' }} onClick={() => navigate('/health/advice')}>
                查看更多建议 →
              </Button>
            </Card>

            {/* 健康小贴士 */}
            <Alert
              message="健康小贴士"
              description="规律作息、均衡饮食、适量运动是维持健康的三大基石。建议每天至少进行30分钟的中等强度运动。"
              type="info"
              showIcon
              icon={<MedicineBoxOutlined />}
            />
          </Space>
        </Col>
      </Row>
    </div>
  );
};

export default HealthManagement;