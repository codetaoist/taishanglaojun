import React from 'react';
import { Card, Row, Col, Statistic, Button, Typography, Space, Badge } from 'antd';
import { 
  BookOutlined, 
  MessageOutlined, 
  TeamOutlined, 
  RobotOutlined,
  BulbOutlined,
  BarChartOutlined,
  ArrowRightOutlined,
  StarOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Paragraph } = Typography;

const Home: React.FC = () => {
  const navigate = useNavigate();

  const features = [
    {
      icon: <BookOutlined className="text-4xl text-cultural-gold" />,
      title: '文化智慧库',
      description: '汇聚千年文化精髓，传承古圣先贤智慧',
      action: () => navigate('/wisdom'),
      stats: { value: 1000, suffix: '+', label: '智慧条目' }
    },
    {
      icon: <MessageOutlined className="text-4xl text-primary-500" />,
      title: 'AI智慧对话',
      description: '与AI助手深度交流，获得个性化指导',
      action: () => navigate('/chat'),
      stats: { value: 50000, suffix: '+', label: '对话次数' },
      badge: 'AI'
    },
    {
      icon: <TeamOutlined className="text-4xl text-cultural-jade" />,
      title: '修行社区',
      description: '与志同道合者交流，共同成长进步',
      action: () => navigate('/community'),
      stats: { value: 5000, suffix: '+', label: '活跃用户' }
    }
  ];

  const aiFeatures = [
    {
      icon: <BulbOutlined className="text-2xl text-yellow-500" />,
      title: 'AI智慧解读',
      description: '深度解析古典智慧的现代意义',
      count: '1000+'
    },
    {
      icon: <BarChartOutlined className="text-2xl text-blue-500" />,
      title: 'AI深度分析',
      description: '多维度分析智慧内容的实用价值',
      count: '500+'
    },
    {
      icon: <RobotOutlined className="text-2xl text-green-500" />,
      title: 'AI智能推荐',
      description: '基于个人兴趣推荐相关智慧内容',
      count: '10000+'
    }
  ];

  const recentWisdom = [
    {
      title: '道德经第一章：道可道，非常道',
      category: '道家经典',
      views: 1234,
      likes: 89
    },
    {
      title: '论语·学而：学而时习之，不亦说乎',
      category: '儒家经典',
      views: 987,
      likes: 76
    },
    {
      title: '心经：观自在菩萨，行深般若波罗蜜多时',
      category: '佛家经典',
      views: 1567,
      likes: 123
    }
  ];

  return (
    <div className="space-y-8">
      {/* 欢迎区域 */}
      <div className="bg-gradient-to-br from-cultural-gold/20 via-cultural-red/10 to-primary-100/30 rounded-2xl p-12 text-center shadow-xl border border-cultural-gold/20">
        <div className="max-w-4xl mx-auto">
          <Title level={1} className="mb-6 bg-gradient-to-r from-cultural-gold to-cultural-red bg-clip-text text-transparent">
            欢迎来到太上老君智慧平台
          </Title>
          <Paragraph className="text-xl text-slate-700 mb-8 leading-relaxed">
            融合传统文化与现代AI技术，为您提供个性化的智慧学习体验
          </Paragraph>
          <Space size="large">
            <Button 
              type="primary" 
              size="large" 
              icon={<BookOutlined />}
              onClick={() => navigate('/wisdom')}
              className="h-12 px-8 text-lg font-medium shadow-lg hover:shadow-xl transition-all duration-300"
            >
              开始探索智慧
            </Button>
            <Button 
              size="large" 
              icon={<MessageOutlined />}
              onClick={() => navigate('/chat')}
              className="h-12 px-8 text-lg font-medium shadow-lg hover:shadow-xl transition-all duration-300 border-cultural-gold text-cultural-gold hover:bg-cultural-gold hover:text-white"
            >
              AI智慧对话
            </Button>
          </Space>
        </div>
      </div>

      {/* 核心功能卡片 */}
      <Row gutter={[24, 24]}>
        {features.map((feature, index) => (
          <Col xs={24} md={8} key={index}>
            <Card 
              hoverable
              className="h-full shadow-lg hover:shadow-xl transition-all duration-300 border-0 bg-white/80 backdrop-blur-sm"
              actions={[
                <Button 
                  type="link" 
                  icon={<ArrowRightOutlined />}
                  onClick={feature.action}
                  className="text-primary-500 hover:text-primary-600 font-medium"
                >
                  立即体验
                </Button>
              ]}
            >
              <div className="text-center space-y-6 p-4">
                <div className="w-16 h-16 mx-auto bg-gradient-to-br from-slate-100 to-slate-200 rounded-2xl flex items-center justify-center shadow-md relative">
                  {feature.icon}
                  {feature.badge && (
                    <Badge 
                      count={feature.badge} 
                      className="absolute -top-2 -right-2"
                      style={{ backgroundColor: '#52c41a' }}
                    />
                  )}
                </div>
                <Title level={3} className="mb-3 text-slate-800">{feature.title}</Title>
                <Paragraph className="text-slate-600 text-base leading-relaxed">
                  {feature.description}
                </Paragraph>
                <div className="bg-gradient-to-r from-cultural-gold/10 to-cultural-red/10 rounded-xl p-4">
                  <Statistic 
                    value={feature.stats.value}
                    suffix={feature.stats.suffix}
                    title={feature.stats.label}
                    valueStyle={{ color: '#d4af37', fontSize: '24px', fontWeight: 'bold' }}

                  />
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* AI功能展示 */}
      <Card 
        title={
          <div className="flex items-center space-x-2">
            <RobotOutlined className="text-blue-500" />
            <span>AI智能功能</span>
            <Badge count="NEW" style={{ backgroundColor: '#52c41a' }} />
          </div>
        }
        className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200"
      >
        <Row gutter={[24, 24]}>
          {aiFeatures.map((feature, index) => (
            <Col xs={24} md={8} key={index}>
              <Card 
                size="small" 
                hoverable
                className="text-center h-full shadow-sm hover:shadow-md transition-all duration-300"
              >
                <div className="space-y-4 p-2">
                  <div className="w-12 h-12 mx-auto bg-gradient-to-br from-white to-gray-100 rounded-xl flex items-center justify-center shadow-sm">
                    {feature.icon}
                  </div>
                  <Title level={5} className="mb-2 text-gray-800">{feature.title}</Title>
                  <Paragraph className="text-gray-600 text-sm mb-3">
                    {feature.description}
                  </Paragraph>
                  <div className="bg-gradient-to-r from-blue-100 to-purple-100 rounded-lg p-2">
                    <span className="text-blue-600 font-semibold">{feature.count}</span>
                    <div className="text-xs text-gray-500 mt-1">已处理</div>
                  </div>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
        <div className="text-center mt-6">
          <Button 
            type="primary" 
            size="large"
            icon={<BookOutlined />}
            onClick={() => navigate('/wisdom')}
            className="bg-gradient-to-r from-blue-500 to-purple-500 border-0 shadow-lg hover:shadow-xl transition-all duration-300"
          >
            体验AI智慧功能
          </Button>
        </div>
      </Card>

      {/* 最新智慧内容 */}
      <Card 
        title={
          <div className="flex items-center justify-between">
            <span className="flex items-center space-x-2">
              <StarOutlined className="text-cultural-gold" />
              <span>热门智慧</span>
            </span>
            <Button 
              type="link" 
              onClick={() => navigate('/wisdom')}
            >
              查看更多
            </Button>
          </div>
        }
      >
        <Row gutter={[16, 16]}>
          {recentWisdom.map((item, index) => (
            <Col xs={24} md={8} key={index}>
              <Card 
                size="small" 
                hoverable
                className="cursor-pointer"
                onClick={() => navigate(`/wisdom/${index + 1}`)}
              >
                <div className="space-y-2">
                  <Title level={5} className="mb-2 line-clamp-2">
                    {item.title}
                  </Title>
                  <div className="flex items-center justify-between text-sm text-gray-500">
                    <span className="bg-primary-50 text-primary-600 px-2 py-1 rounded">
                      {item.category}
                    </span>
                    <Space>
                      <span>{item.views} 阅读</span>
                      <span>{item.likes} 点赞</span>
                    </Space>
                  </div>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 统计数据 */}
      <Card title="平台数据">
        <Row gutter={[24, 24]} className="text-center">
          <Col xs={12} md={6}>
            <Statistic 
              title="智慧条目" 
              value={1000} 
              suffix="+"
              valueStyle={{ color: '#d4af37' }}
            />
          </Col>
          <Col xs={12} md={6}>
            <Statistic 
              title="注册用户" 
              value={5000} 
              suffix="+"
              valueStyle={{ color: '#059669' }}
            />
          </Col>
          <Col xs={12} md={6}>
            <Statistic 
              title="AI对话" 
              value={50000} 
              suffix="+"
              valueStyle={{ color: '#0ea5e9' }}
            />
          </Col>
          <Col xs={12} md={6}>
            <Statistic 
              title="社区讨论" 
              value={10000} 
              suffix="+"
              valueStyle={{ color: '#dc2626' }}
            />
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default Home;