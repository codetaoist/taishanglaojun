import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Statistic, 
  Button, 
  Space, 
  Typography, 
  Avatar, 
  List,
  Progress,
  Tag,
  Divider,
  notification
} from 'antd';
import {
  MessageOutlined,
  BookOutlined,
  RobotOutlined,
  TeamOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  ArrowRightOutlined,
  PlusOutlined,
  StarOutlined,
  FireOutlined,
  ThunderboltOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthContext as useAuth } from '../contexts/AuthContext';
import PageContainer from '../components/common/PageContainer';
import './HomePage.css';

const { Title, Paragraph, Text } = Typography;

const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [stats, setStats] = useState({
    totalChats: 156,
    wisdomQueries: 89,
    learningProgress: 75,
    communityPoints: 1240
  });

  // 快捷功能卡片
  const quickActions = [
    {
      title: 'AI智能对话',
      description: '与太上老君AI进行深度对话',
      icon: <RobotOutlined className="text-2xl text-blue-500" />,
      path: '/chat',
      color: 'blue',
      gradient: 'from-blue-500 to-cyan-500'
    },
    {
      title: '文化智慧',
      description: '探索中华传统文化精髓',
      icon: <BookOutlined className="text-2xl text-purple-500" />,
      path: '/wisdom',
      color: 'purple',
      gradient: 'from-purple-500 to-pink-500'
    },
    {
      title: '学习中心',
      description: '个性化学习路径规划',
      icon: <StarOutlined className="text-2xl text-yellow-500" />,
      path: '/learning',
      color: 'yellow',
      gradient: 'from-yellow-500 to-orange-500'
    },
    {
      title: '社区交流',
      description: '与志同道合的朋友交流',
      icon: <TeamOutlined className="text-2xl text-green-500" />,
      path: '/community',
      color: 'green',
      gradient: 'from-green-500 to-emerald-500'
    }
  ];

  // 最近活动
  const recentActivities = [
    {
      id: 1,
      type: 'chat',
      title: '完成了关于"道德经"的深度对话',
      time: '2小时前',
      icon: <MessageOutlined className="text-blue-500" />
    },
    {
      id: 2,
      type: 'wisdom',
      title: '学习了"中庸之道"的智慧',
      time: '1天前',
      icon: <BookOutlined className="text-purple-500" />
    },
    {
      id: 3,
      type: 'achievement',
      title: '获得了"智慧探索者"徽章',
      time: '2天前',
      icon: <TrophyOutlined className="text-yellow-500" />
    }
  ];

  // 推荐内容
  const recommendations = [
    {
      id: 1,
      title: '道德经第一章解读',
      description: '深入理解"道可道，非常道"的哲学内涵',
      category: '经典解读',
      hot: true
    },
    {
      id: 2,
      title: '中医养生智慧',
      description: '传统中医的养生理念与现代生活的结合',
      category: '养生智慧',
      new: true
    },
    {
      id: 3,
      title: '太极哲学思想',
      description: '阴阳平衡的智慧在现代管理中的应用',
      category: '哲学思辨',
      trending: true
    }
  ];

  const handleQuickAction = (path: string) => {
    navigate(path);
  };

  const handleViewMore = (type: string) => {
    switch (type) {
      case 'activities':
        navigate('/profile/activities');
        break;
      case 'recommendations':
        navigate('/wisdom/recommendations');
        break;
      default:
        break;
    }
  };

  return (
    <PageContainer ghost>
      <div className="homepage-container">
        <div className="space-y-6" style={{ position: 'relative', zIndex: 5, padding: '24px' }}>
          {/* 欢迎横幅 */}
          <Card className="welcome-banner-card">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
                <span className="text-white font-bold text-2xl">太</span>
              </div>
              <div>
                <Title level={2} className="mb-2 gradient-text">
                  欢迎回来，{user?.name || '道友'}！
                </Title>
                <Paragraph className="mb-0" style={{ color: 'rgba(255, 255, 255, 0.8)' }}>
                  今天是探索智慧的好日子，让我们一起开启新的学习之旅
                </Paragraph>
              </div>
            </div>
            <div className="hidden md:block">
              <Button 
                type="primary" 
                size="large" 
                icon={<PlusOutlined />}
                className="bg-gradient-to-r from-blue-500 to-purple-600 border-none shadow-lg hover:shadow-xl transition-all duration-300"
                onClick={() => navigate('/chat')}
              >
                开始对话
              </Button>
            </div>
          </div>
        </Card>

        {/* 数据统计 */}
        <Row gutter={[16, 16]}>
          <Col xs={12} sm={6}>
            <Card className="homepage-stat-card text-center">
              <Statistic
                title="对话次数"
                value={stats.totalChats}
                prefix={<MessageOutlined />}
                valueStyle={{ color: 'white' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card className="homepage-stat-card text-center">
              <Statistic
                title="智慧查询"
                value={stats.wisdomQueries}
                prefix={<BookOutlined />}
                valueStyle={{ color: 'white' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card className="homepage-stat-card text-center">
              <Statistic
                title="学习进度"
                value={stats.learningProgress}
                suffix="%"
                prefix={<StarOutlined />}
                valueStyle={{ color: 'white' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card className="homepage-stat-card text-center">
              <Statistic
                title="社区积分"
                value={stats.communityPoints}
                prefix={<TrophyOutlined />}
                valueStyle={{ color: 'white' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 快捷功能 */}
        <Card title="快捷功能" className="homepage-feature-card">
          <Row gutter={[16, 16]}>
            {quickActions.map((action, index) => (
              <Col xs={12} md={6} key={index}>
                <Card
                  hoverable
                  className="quick-action-card"
                  onClick={() => handleQuickAction(action.path)}
                >
                  <div className="text-center space-y-3">
                    <div className={`w-16 h-16 mx-auto bg-gradient-to-r ${action.gradient} rounded-2xl flex items-center justify-center shadow-lg`}>
                      {action.icon}
                    </div>
                    <div>
                      <Title level={4} className="mb-2">{action.title}</Title>
                      <Text type="secondary" className="text-sm">
                        {action.description}
                      </Text>
                    </div>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>
        </Card>

        <Row gutter={[16, 16]}>
          {/* 最近活动 */}
          <Col xs={24} lg={12}>
            <Card 
              title="最近活动" 
              className="activity-recommendation-card h-full"
              extra={
                <Button 
                  type="link" 
                  icon={<ArrowRightOutlined />}
                  onClick={() => handleViewMore('activities')}
                >
                  查看更多
                </Button>
              }
            >
              <List
                dataSource={recentActivities}
                renderItem={(item) => (
                  <List.Item className="border-none py-3">
                    <List.Item.Meta
                      avatar={
                        <div className="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center">
                          {item.icon}
                        </div>
                      }
                      title={<Text className="font-medium">{item.title}</Text>}
                      description={
                        <div className="flex items-center space-x-2 text-gray-500">
                          <ClockCircleOutlined />
                          <span>{item.time}</span>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          </Col>

          {/* 推荐内容 */}
          <Col xs={24} lg={12}>
            <Card 
              title="推荐内容" 
              className="activity-recommendation-card h-full"
              extra={
                <Button 
                  type="link" 
                  icon={<ArrowRightOutlined />}
                  onClick={() => handleViewMore('recommendations')}
                >
                  查看更多
                </Button>
              }
            >
              <List
                dataSource={recommendations}
                renderItem={(item) => (
                  <List.Item className="border-none py-3 cursor-pointer hover:bg-gray-50 rounded-lg px-2 transition-colors">
                    <List.Item.Meta
                      title={
                        <div className="flex items-center justify-between">
                          <Text className="font-medium">{item.title}</Text>
                          <div className="flex space-x-1">
                            {item.hot && <Tag color="red" icon={<FireOutlined />}>热门</Tag>}
                            {item.new && <Tag color="blue">新</Tag>}
                            {item.trending && <Tag color="orange" icon={<ThunderboltOutlined />}>趋势</Tag>}
                          </div>
                        </div>
                      }
                      description={
                        <div className="space-y-1">
                          <Text type="secondary" className="text-sm">{item.description}</Text>
                          <Tag size="small" color="geekblue">{item.category}</Tag>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          </Col>
        </Row>

        {/* 学习进度 */}
        <Card title="学习进度" className="learning-progress-card">
          <Row gutter={[16, 16]}>
            <Col xs={24} md={8}>
              <div className="text-center">
                <Progress
                  type="circle"
                  percent={stats.learningProgress}
                  strokeColor={{
                    '0%': '#108ee9',
                    '100%': '#87d068',
                  }}
                  size={120}
                />
                <div className="mt-4">
                  <Title level={4}>总体进度</Title>
                  <Text type="secondary">继续保持，你做得很棒！</Text>
                </div>
              </div>
            </Col>
            <Col xs={24} md={16}>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between mb-2">
                    <Text>道德经研读</Text>
                    <Text>85%</Text>
                  </div>
                  <Progress percent={85} strokeColor="#52c41a" />
                </div>
                <div>
                  <div className="flex justify-between mb-2">
                    <Text>中医养生</Text>
                    <Text>60%</Text>
                  </div>
                  <Progress percent={60} strokeColor="#1890ff" />
                </div>
                <div>
                  <div className="flex justify-between mb-2">
                    <Text>太极哲学</Text>
                    <Text>40%</Text>
                  </div>
                  <Progress percent={40} strokeColor="#faad14" />
                </div>
              </div>
            </Col>
          </Row>
        </Card>
        </div>
      </div>
    </PageContainer>
  );
};

export default HomePage;