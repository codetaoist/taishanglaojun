import React from 'react';
import { Card, Row, Col, Statistic, Typography, Space, Button, List, Avatar, Tag } from 'antd';
import { 
  UserOutlined, 
  BookOutlined, 
  MessageOutlined, 
  HeartOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  RiseOutlined,
  TeamOutlined
} from '@ant-design/icons';
import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';

const { Title, Paragraph } = Typography;

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

  // 模拟数据
  const stats = {
    totalWisdom: 1248,
    totalUsers: 3567,
    todayVisits: 892,
    myFavorites: 23
  };

  const recentActivities = [
    {
      id: '1',
      type: 'wisdom',
      title: '新增智慧：《道德经》第一章解读',
      time: '2小时前',
      icon: <BookOutlined />
    },
    {
      id: '2',
      type: 'community',
      title: '参与讨论：传统文化在现代社会的价值',
      time: '4小时前',
      icon: <MessageOutlined />
    },
    {
      id: '3',
      type: 'favorite',
      title: '收藏了：《论语》学而篇精选',
      time: '1天前',
      icon: <HeartOutlined />
    }
  ];

  const quickActions = [
    {
      title: 'AI智能对话',
      description: '与AI助手探讨文化智慧',
      path: '/chat',
      icon: '🤖'
    },
    {
      title: '浏览智慧',
      description: '探索传统文化宝库',
      path: '/wisdom',
      icon: '📚'
    },
    {
      title: '社区讨论',
      description: '与同道中人交流心得',
      path: '/community',
      icon: '👥'
    },
    {
      title: '我的收藏',
      description: '查看收藏的智慧内容',
      path: '/favorites',
      icon: '❤️'
    }
  ];

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* 欢迎区域 */}
        <div className="mb-6">
          <Title level={2}>
            欢迎回来，{user?.name || user?.username}！
          </Title>
          <Paragraph className="text-gray-600">
            今天是探索传统文化智慧的好日子，让我们一起开始学习之旅吧。
          </Paragraph>
        </div>

        {/* 统计卡片 */}
        <Row gutter={[16, 16]} className="mb-6">
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="智慧总数"
                value={stats.totalWisdom}
                prefix={<BookOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="用户总数"
                value={stats.totalUsers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="今日访问"
                value={stats.todayVisits}
                prefix={<RiseOutlined />}
                valueStyle={{ color: '#cf1322' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="我的收藏"
                value={stats.myFavorites}
                prefix={<HeartOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          {/* 快速操作 */}
          <Col xs={24} lg={12}>
            <Card title="快速操作" className="h-full">
              <Row gutter={[16, 16]}>
                {quickActions.map((action, index) => (
                  <Col xs={12} key={index}>
                    <Card 
                      hoverable
                      className="text-center cursor-pointer"
                      onClick={() => navigate(action.path)}
                    >
                      <div className="text-2xl mb-2">{action.icon}</div>
                      <div className="font-medium">{action.title}</div>
                      <div className="text-gray-500 text-sm mt-1">
                        {action.description}
                      </div>
                    </Card>
                  </Col>
                ))}
              </Row>
            </Card>
          </Col>

          {/* 最近活动 */}
          <Col xs={24} lg={12}>
            <Card title="最近活动" className="h-full">
              <List
                dataSource={recentActivities}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={item.icon} />}
                      title={item.title}
                      description={
                        <Space>
                          <ClockCircleOutlined />
                          <span>{item.time}</span>
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            </Card>
          </Col>
        </Row>

        {/* 学习进度和成就 */}
        <Row gutter={[16, 16]} className="mt-6">
          <Col xs={24} lg={8}>
            <Card title="学习进度" className="h-full">
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between mb-1">
                    <span>道德经</span>
                    <span>75%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-blue-600 h-2 rounded-full" style={{ width: '75%' }}></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <span>论语</span>
                    <span>45%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-green-600 h-2 rounded-full" style={{ width: '45%' }}></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <span>孟子</span>
                    <span>20%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-yellow-600 h-2 rounded-full" style={{ width: '20%' }}></div>
                  </div>
                </div>
              </div>
            </Card>
          </Col>

          <Col xs={24} lg={8}>
            <Card title="学习成就" className="h-full">
              <Space direction="vertical" className="w-full">
                <div className="flex items-center space-x-2">
                  <TrophyOutlined className="text-yellow-500" />
                  <span>连续学习7天</span>
                  <Tag color="gold">已达成</Tag>
                </div>
                <div className="flex items-center space-x-2">
                  <BookOutlined className="text-blue-500" />
                  <span>阅读100篇智慧</span>
                  <Tag color="blue">已达成</Tag>
                </div>
                <div className="flex items-center space-x-2">
                  <HeartOutlined className="text-red-500" />
                  <span>收藏50篇内容</span>
                  <Tag color="orange">进行中</Tag>
                </div>
                <div className="flex items-center space-x-2">
                  <TeamOutlined className="text-green-500" />
                  <span>参与10次讨论</span>
                  <Tag color="green">进行中</Tag>
                </div>
              </Space>
            </Card>
          </Col>

          <Col xs={24} lg={8}>
            <Card title="今日推荐" className="h-full">
              <div className="space-y-3">
                <div className="p-3 bg-blue-50 rounded-lg">
                  <div className="font-medium text-blue-800">每日一句</div>
                  <div className="text-sm text-blue-600 mt-1">
                    "学而时习之，不亦说乎？"
                  </div>
                  <div className="text-xs text-blue-500 mt-1">
                    —— 《论语·学而》
                  </div>
                </div>
                <Button 
                  type="primary" 
                  block 
                  onClick={() => navigate('/recommendations')}
                >
                  查看更多推荐
                </Button>
              </div>
            </Card>
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default Dashboard;