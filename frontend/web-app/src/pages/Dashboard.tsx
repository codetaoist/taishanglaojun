import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Statistic, 
  Button, 
  List, 
  Avatar, 
  Progress, 
  Typography, 
  Space, 
  Tag, 
  Divider,
  Badge,
  Timeline,
  Carousel,
  Alert,
  Tooltip,
  Spin
} from 'antd';
import {
  MessageOutlined,
  BookOutlined,
  TeamOutlined,
  HeartOutlined,
  TrophyOutlined,
  RocketOutlined,
  BulbOutlined,
  StarOutlined,
  ClockCircleOutlined,
  UserOutlined,
  EyeOutlined,
  LikeOutlined,
  SecurityScanOutlined,
  BarChartOutlined,
  ProjectOutlined,
  ReadOutlined,
  SettingOutlined,
  BellOutlined,
  FireOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  SyncOutlined,
  CalendarOutlined,
  LineChartOutlined,
  RiseOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const { Title, Text, Paragraph } = Typography;

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [systemStatus, setSystemStatus] = useState('normal'); // normal, warning, error

  // 模拟数据加载
  useEffect(() => {
    setLoading(true);
    setTimeout(() => setLoading(false), 1000);
  }, []);

  // 系统统计数据
  const systemStats = {
    totalWisdom: { value: 12580, growth: 8.5 },
    totalUsers: { value: 8964, growth: 12.3 },
    totalVisits: { value: 156789, growth: -2.1 },
    totalFavorites: { value: 3456, growth: 15.7 },
    aiInteractions: { value: 45678, growth: 23.4 },
    securityEvents: { value: 12, growth: -45.2 },
    projectsActive: { value: 89, growth: 6.8 },
    learningProgress: { value: 78.5, growth: 4.2 }
  };

  // 最近活动
  const recentActivities = [
    {
      id: 1,
      user: '张三',
      action: '完成了AI对话训练',
      target: '智能问答系统',
      time: '2分钟前',
      type: 'ai',
      avatar: 'https://api.dicebear.com/7.x/miniavs/svg?seed=1'
    },
    {
      id: 2,
      user: '李四',
      action: '发现安全漏洞',
      target: '用户认证模块',
      time: '5分钟前',
      type: 'security',
      avatar: 'https://api.dicebear.com/7.x/miniavs/svg?seed=2'
    },
    {
      id: 3,
      user: '王五',
      action: '创建了新项目',
      target: '智慧学习平台',
      time: '10分钟前',
      type: 'project',
      avatar: 'https://api.dicebear.com/7.x/miniavs/svg?seed=3'
    },
    {
      id: 4,
      user: '赵六',
      action: '收藏了智慧内容',
      target: '《道德经》第一章',
      time: '15分钟前',
      type: 'wisdom',
      avatar: 'https://api.dicebear.com/7.x/miniavs/svg?seed=4'
    }
  ];

  // 快捷操作
  const quickActions = [
    {
      title: 'AI智能对话',
      description: '与AI助手进行深度对话',
      icon: <MessageOutlined />,
      color: '#1890ff',
      path: '/chat',
      badge: '新功能'
    },
    {
      title: '安全中心',
      description: '查看系统安全状态',
      icon: <SecurityScanOutlined />,
      color: '#f5222d',
      path: '/security',
      badge: systemStats.securityEvents.value > 0 ? '有警告' : null
    },
    {
      title: '智慧库',
      description: '探索古今智慧宝库',
      icon: <BookOutlined />,
      color: '#52c41a',
      path: '/wisdom'
    },
    {
      title: '项目管理',
      description: '管理您的项目和任务',
      icon: <ProjectOutlined />,
      color: '#722ed1',
      path: '/projects/workspace'
    },
    {
      title: '学习中心',
      description: '智能学习和能力提升',
      icon: <ReadOutlined />,
      color: '#fa8c16',
      path: '/learning/courses'
    },
    {
      title: '系统设置',
      description: '配置系统参数',
      icon: <SettingOutlined />,
      color: '#13c2c2',
      path: '/admin/settings'
    }
  ];

  // 系统通知
  const systemNotifications = [
    {
      id: 1,
      type: 'info',
      title: '系统更新',
      message: 'AI模块已更新到v2.1.0版本',
      time: '1小时前'
    },
    {
      id: 2,
      type: 'warning',
      title: '安全提醒',
      message: '检测到异常登录尝试，请注意账户安全',
      time: '2小时前'
    },
    {
      id: 3,
      type: 'success',
      title: '备份完成',
      message: '数据库备份已成功完成',
      time: '3小时前'
    }
  ];

  // 学习进度数据
  const learningProgress = [
    { subject: 'AI技术', progress: 85, color: '#1890ff' },
    { subject: '安全防护', progress: 72, color: '#f5222d' },
    { subject: '项目管理', progress: 90, color: '#52c41a' },
    { subject: '文化智慧', progress: 68, color: '#722ed1' }
  ];

  // 获取统计数据的增长趋势图标
  const getTrendIcon = (growth: number) => {
    if (growth > 0) {
      return <ArrowUpOutlined style={{ color: '#52c41a' }} />;
    } else if (growth < 0) {
      return <ArrowDownOutlined style={{ color: '#f5222d' }} />;
    }
    return <SyncOutlined style={{ color: '#1890ff' }} />;
  };

  // 获取活动类型图标
  const getActivityIcon = (type: string) => {
    const iconMap = {
      ai: <MessageOutlined style={{ color: '#1890ff' }} />,
      security: <SecurityScanOutlined style={{ color: '#f5222d' }} />,
      project: <ProjectOutlined style={{ color: '#722ed1' }} />,
      wisdom: <BookOutlined style={{ color: '#52c41a' }} />
    };
    return iconMap[type as keyof typeof iconMap] || <InfoCircleOutlined />;
  };

  // 获取通知类型图标
  const getNotificationIcon = (type: string) => {
    const iconMap = {
      info: <InfoCircleOutlined style={{ color: '#1890ff' }} />,
      warning: <ExclamationCircleOutlined style={{ color: '#fa8c16' }} />,
      success: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
      error: <ExclamationCircleOutlined style={{ color: '#f5222d' }} />
    };
    return iconMap[type as keyof typeof iconMap] || <InfoCircleOutlined />;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <Spin size="large" tip="加载仪表板数据..." />
      </div>
    );
  }

  return (
    <div className="p-6 bg-gradient-to-br from-blue-50 to-indigo-100 min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* 欢迎区域和系统状态 */}
        <div className="mb-6">
          <div className="flex justify-between items-start">
            <div>
              <Title level={2} className="mb-2 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                欢迎回来，{user?.name || user?.username}！
              </Title>
              <Paragraph className="text-gray-600 text-lg">
                太上老君AI平台 - 智慧与科技的完美融合
              </Paragraph>
            </div>
            <div className="text-right">
              <Badge 
                status={systemStatus === 'normal' ? 'success' : systemStatus === 'warning' ? 'warning' : 'error'} 
                text={systemStatus === 'normal' ? '系统正常' : systemStatus === 'warning' ? '系统警告' : '系统异常'}
              />
              <div className="text-sm text-gray-500 mt-1">
                <CalendarOutlined /> {new Date().toLocaleDateString('zh-CN', { 
                  year: 'numeric', 
                  month: 'long', 
                  day: 'numeric',
                  weekday: 'long'
                })}
              </div>
            </div>
          </div>
        </div>

        {/* 系统通知 */}
        {systemNotifications.length > 0 && (
          <div className="mb-6">
            <Carousel autoplay dots={false} autoplaySpeed={5000}>
              {systemNotifications.map(notification => (
                <div key={notification.id}>
                  <Alert
                    message={notification.title}
                    description={notification.message}
                    type={notification.type as any}
                    showIcon
                    icon={getNotificationIcon(notification.type)}
                    action={
                      <Space>
                        <Text type="secondary" className="text-xs">{notification.time}</Text>
                        <Button size="small" type="text">详情</Button>
                      </Space>
                    }
                    closable
                  />
                </div>
              ))}
            </Carousel>
          </div>
        )}

        {/* 核心统计数据 */}
        <Row gutter={[16, 16]} className="mb-6">
          <Col xs={24} sm={12} lg={6}>
            <Card className="hover:shadow-lg transition-shadow">
              <Statistic
                title="智慧总数"
                value={systemStats.totalWisdom.value}
                prefix={<BookOutlined />}
                suffix={
                  <Tooltip title={`增长率: ${systemStats.totalWisdom.growth}%`}>
                    {getTrendIcon(systemStats.totalWisdom.growth)}
                  </Tooltip>
                }
                valueStyle={{ color: '#3f8600' }}
              />
              <div className="text-xs text-gray-500 mt-2">
                较上月 {systemStats.totalWisdom.growth > 0 ? '增长' : '下降'} {Math.abs(systemStats.totalWisdom.growth)}%
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card className="hover:shadow-lg transition-shadow">
              <Statistic
                title="AI交互次数"
                value={systemStats.aiInteractions.value}
                prefix={<MessageOutlined />}
                suffix={
                  <Tooltip title={`增长率: ${systemStats.aiInteractions.growth}%`}>
                    {getTrendIcon(systemStats.aiInteractions.growth)}
                  </Tooltip>
                }
                valueStyle={{ color: '#1890ff' }}
              />
              <div className="text-xs text-gray-500 mt-2">
                较上月增长 {systemStats.aiInteractions.growth}%
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card className="hover:shadow-lg transition-shadow">
              <Statistic
                title="活跃项目"
                value={systemStats.projectsActive.value}
                prefix={<ProjectOutlined />}
                suffix={
                  <Tooltip title={`增长率: ${systemStats.projectsActive.growth}%`}>
                    {getTrendIcon(systemStats.projectsActive.growth)}
                  </Tooltip>
                }
                valueStyle={{ color: '#722ed1' }}
              />
              <div className="text-xs text-gray-500 mt-2">
                较上月增长 {systemStats.projectsActive.growth}%
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card className="hover:shadow-lg transition-shadow">
              <Statistic
                title="安全事件"
                value={systemStats.securityEvents.value}
                prefix={<SecurityScanOutlined />}
                suffix={
                  <Tooltip title={`变化率: ${systemStats.securityEvents.growth}%`}>
                    {getTrendIcon(systemStats.securityEvents.growth)}
                  </Tooltip>
                }
                valueStyle={{ color: systemStats.securityEvents.value > 0 ? '#f5222d' : '#52c41a' }}
              />
              <div className="text-xs text-gray-500 mt-2">
                较上月下降 {Math.abs(systemStats.securityEvents.growth)}%
              </div>
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          {/* 快捷操作 */}
          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <RocketOutlined />
                  快捷操作
                </Space>
              } 
              className="h-full"
              extra={<Button type="link" size="small">更多</Button>}
            >
              <Row gutter={[12, 12]}>
                {quickActions.map((action, index) => (
                  <Col xs={12} sm={8} key={index}>
                    <Badge.Ribbon text={action.badge} color={action.badge ? 'red' : undefined}>
                      <Card
                        hoverable
                        className="text-center cursor-pointer h-full"
                        onClick={() => navigate(action.path)}
                        bodyStyle={{ padding: '16px 8px' }}
                      >
                        <div className="text-2xl mb-2" style={{ color: action.color }}>
                          {action.icon}
                        </div>
                        <div className="font-medium text-sm">{action.title}</div>
                        <div className="text-xs text-gray-500 mt-1">
                          {action.description}
                        </div>
                      </Card>
                    </Badge.Ribbon>
                  </Col>
                ))}
              </Row>
            </Card>
          </Col>

          {/* 最近活动 */}
          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <ClockCircleOutlined />
                  最近活动
                </Space>
              } 
              className="h-full"
              extra={<Button type="link" size="small">查看全部</Button>}
            >
              <List
                dataSource={recentActivities}
                renderItem={(item) => (
                  <List.Item className="hover:bg-gray-50 rounded p-2 transition-colors">
                    <List.Item.Meta
                      avatar={
                        <Badge dot={item.type === 'security'}>
                          <Avatar src={item.avatar} />
                        </Badge>
                      }
                      title={
                        <Space size="small">
                          {getActivityIcon(item.type)}
                          <Text strong>{item.user}</Text>
                          <Text type="secondary">{item.action}</Text>
                        </Space>
                      }
                      description={
                        <div>
                          <Text className="text-sm">{item.target}</Text>
                          <div className="text-xs text-gray-400 mt-1">
                            <ClockCircleOutlined /> {item.time}
                          </div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          </Col>
        </Row>

        {/* 学习进度和系统概览 */}
        <Row gutter={[16, 16]} className="mt-6">
          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <LineChartOutlined />
                  学习进度
                </Space>
              }
              extra={<Button type="link" size="small">详细报告</Button>}
            >
              <div className="space-y-4">
                {learningProgress.map((item, index) => (
                  <div key={index}>
                    <div className="flex justify-between mb-2">
                      <Text>{item.subject}</Text>
                      <Text strong>{item.progress}%</Text>
                    </div>
                    <Progress 
                      percent={item.progress} 
                      strokeColor={item.color}
                      trailColor="#f0f0f0"
                      strokeWidth={8}
                    />
                  </div>
                ))}
              </div>
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <BarChartOutlined />
                  系统概览
                </Space>
              }
              extra={<Button type="link" size="small">监控中心</Button>}
            >
              <Row gutter={[16, 16]}>
                <Col span={12}>
                  <Statistic
                    title="总用户数"
                    value={systemStats.totalUsers.value}
                    valueStyle={{ fontSize: '20px' }}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="总访问量"
                    value={systemStats.totalVisits.value}
                    valueStyle={{ fontSize: '20px' }}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="收藏总数"
                    value={systemStats.totalFavorites.value}
                    valueStyle={{ fontSize: '20px' }}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="学习进度"
                    value={systemStats.learningProgress.value}
                    suffix="%"
                    valueStyle={{ fontSize: '20px' }}
                  />
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>

        {/* 成就与推荐 */}
        <Row gutter={[16, 16]} className="mt-6">
          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <TrophyOutlined />
                  成就与荣誉
                </Space>
              }
            >
              <Timeline
                items={[
                  {
                    dot: <TrophyOutlined className="text-yellow-500" />,
                    children: (
                      <div>
                        <Text strong>智慧探索者</Text>
                        <div className="text-sm text-gray-500">已学习100篇经典文献</div>
                      </div>
                    ),
                  },
                  {
                    dot: <StarOutlined className="text-blue-500" />,
                    children: (
                      <div>
                        <Text strong>社区贡献者</Text>
                        <div className="text-sm text-gray-500">获得50个点赞</div>
                      </div>
                    ),
                  },
                  {
                    dot: <RocketOutlined className="text-green-500" />,
                    children: (
                      <div>
                        <Text strong>AI对话达人</Text>
                        <div className="text-sm text-gray-500">完成200次AI对话</div>
                      </div>
                    ),
                  },
                  {
                    dot: <FireOutlined className="text-red-500" />,
                    children: (
                      <div>
                        <Text strong>安全卫士</Text>
                        <div className="text-sm text-gray-500">发现并修复5个安全漏洞</div>
                      </div>
                    ),
                  },
                ]}
              />
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <BulbOutlined />
                  今日推荐
                </Space>
              }
            >
              <div className="space-y-4">
                <Card
                  size="small"
                  hoverable
                  className="cursor-pointer"
                  onClick={() => navigate('/wisdom')}
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-12 h-12 bg-gradient-to-r from-blue-400 to-purple-500 rounded-lg flex items-center justify-center">
                      <BookOutlined className="text-white text-lg" />
                    </div>
                    <div className="flex-1">
                      <Text strong>智慧精选</Text>
                      <div className="text-sm text-gray-500">《道德经》第一章：道可道，非常道</div>
                    </div>
                  </div>
                </Card>

                <Card
                  size="small"
                  hoverable
                  className="cursor-pointer"
                  onClick={() => navigate('/chat')}
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-12 h-12 bg-gradient-to-r from-green-400 to-blue-500 rounded-lg flex items-center justify-center">
                      <MessageOutlined className="text-white text-lg" />
                    </div>
                    <div className="flex-1">
                      <Text strong>AI对话推荐</Text>
                      <div className="text-sm text-gray-500">与AI探讨传统文化的现代价值</div>
                    </div>
                  </div>
                </Card>

                <Card
                  size="small"
                  hoverable
                  className="cursor-pointer"
                  onClick={() => navigate('/community')}
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-12 h-12 bg-gradient-to-r from-purple-400 to-pink-500 rounded-lg flex items-center justify-center">
                      <TeamOutlined className="text-white text-lg" />
                    </div>
                    <div className="flex-1">
                      <Text strong>社区热议</Text>
                      <div className="text-sm text-gray-500">传统文化在现代教育中的应用</div>
                    </div>
                  </div>
                </Card>
              </div>
            </Card>
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default Dashboard;