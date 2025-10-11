import React, { useState } from 'react';
import { Card, Row, Col, Button, Typography, Space, Statistic, Badge, Avatar, Divider } from 'antd';
import { 
  PictureOutlined, 
  EyeOutlined, 
  RobotOutlined, 
  ThunderboltOutlined,
  StarOutlined,
  FireOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  BulbOutlined,
  ProjectOutlined,
  ExperimentOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Paragraph, Text } = Typography;

const AIMultimodal: React.FC = () => {
  const navigate = useNavigate();
  const [hoveredCard, setHoveredCard] = useState<string | null>(null);

  // 模拟统计数据
  const stats = {
    totalGenerations: 15420,
    totalAnalyses: 8930,
    activeUsers: 2340,
    successRate: 98.5
  };

  // 功能卡片数据
  const features = [
    {
      key: 'image-generation',
      title: '图像生成',
      description: '基于文本描述生成高质量图像，支持多种艺术风格和创意表达',
      icon: <PictureOutlined />,
      color: '#1890ff',
      path: '/ai/image-generation',
      features: ['文本到图像', '风格转换', '图像编辑', '批量生成'],
      badge: 'HOT',
      stats: { generations: 12450, avgTime: '3.2s' }
    },
    {
      key: 'image-analysis',
      title: '图像分析',
      description: '智能识别图像内容，提供详细的分析报告和洞察',
      icon: <EyeOutlined />,
      color: '#52c41a',
      path: '/ai/image-analysis',
      features: ['内容识别', '场景分析', '情感检测', '相似度比较'],
      badge: 'NEW',
      stats: { analyses: 8930, accuracy: '96.8%' }
    },
    {
      key: 'agi-reasoning',
      title: 'AGI推理',
      description: '通用人工智能推理系统，支持演绎、归纳、溯因等多种推理模式',
      icon: <BulbOutlined />,
      color: '#722ed1',
      path: '/ai/agi-reasoning',
      features: ['逻辑推理', '因果分析', '假设验证', '知识推导'],
      badge: 'BETA',
      stats: { sessions: 3240, accuracy: '94.2%' }
    },
    {
      key: 'agi-planning',
      title: 'AGI规划',
      description: '智能规划系统，自动生成最优执行方案和策略',
      icon: <ProjectOutlined />,
      color: '#fa8c16',
      path: '/ai/agi-planning',
      features: ['目标分解', '路径规划', '资源优化', '风险评估'],
      badge: 'BETA',
      stats: { plans: 1850, success: '91.5%' }
    },
    {
      key: 'meta-learning',
      title: '元学习',
      description: '学会学习的AI系统，快速适应新任务和领域',
      icon: <ExperimentOutlined />,
      color: '#13c2c2',
      path: '/ai/meta-learning',
      features: ['快速适应', '知识迁移', '少样本学习', '持续学习'],
      badge: 'ALPHA',
      stats: { tasks: 680, adaptation: '2.3s' }
    },
    {
      key: 'self-evolution',
      title: '自进化',
      description: '自主进化的AI系统，持续优化性能和能力',
      icon: <ThunderboltOutlined />,
      color: '#f5222d',
      path: '/ai/self-evolution',
      features: ['性能优化', '架构进化', '能力扩展', '自主学习'],
      badge: 'ALPHA',
      stats: { generations: 156, improvement: '23.4%' }
    }
  ];

  // 最近活动数据
  const recentActivities = [
    { type: 'generation', user: '用户A', action: '生成了一幅山水画', time: '2分钟前' },
    { type: 'analysis', user: '用户B', action: '分析了产品图片', time: '5分钟前' },
    { type: 'generation', user: '用户C', action: '创建了logo设计', time: '8分钟前' },
    { type: 'analysis', user: '用户D', action: '识别了文档内容', time: '12分钟前' }
  ];

  const handleFeatureClick = (path: string) => {
    navigate(path);
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '32px', textAlign: 'center' }}>
        <Title level={1} style={{ marginBottom: '8px' }}>
          <RobotOutlined style={{ color: '#1890ff', marginRight: '12px' }} />
          AI多模态智能中心
        </Title>
        <Paragraph style={{ fontSize: '16px', color: '#666', maxWidth: '600px', margin: '0 auto' }}>
          融合最先进的AI技术，为您提供强大的图像生成和分析能力，释放创意潜能，洞察视觉世界
        </Paragraph>
      </div>

      {/* 统计数据 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '32px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总生成次数"
              value={stats.totalGenerations}
              prefix={<PictureOutlined style={{ color: '#1890ff' }} />}
              suffix="次"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总分析次数"
              value={stats.totalAnalyses}
              prefix={<EyeOutlined style={{ color: '#52c41a' }} />}
              suffix="次"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="活跃用户"
              value={stats.activeUsers}
              prefix={<FireOutlined style={{ color: '#fa8c16' }} />}
              suffix="人"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="成功率"
              value={stats.successRate}
              prefix={<TrophyOutlined style={{ color: '#f5222d' }} />}
              suffix="%"
              precision={1}
            />
          </Card>
        </Col>
      </Row>

      {/* 主要功能卡片 */}
      <Row gutter={[24, 24]} style={{ marginBottom: '32px' }}>
        {features.map((feature) => (
          <Col xs={24} md={12} lg={8} key={feature.key}>
            <Card
              hoverable
              style={{
                height: '100%',
                border: hoveredCard === feature.key ? `2px solid ${feature.color}` : '1px solid #d9d9d9',
                transition: 'all 0.3s ease',
                transform: hoveredCard === feature.key ? 'translateY(-4px)' : 'translateY(0)',
                boxShadow: hoveredCard === feature.key ? '0 8px 24px rgba(0,0,0,0.12)' : '0 2px 8px rgba(0,0,0,0.06)'
              }}
              onMouseEnter={() => setHoveredCard(feature.key)}
              onMouseLeave={() => setHoveredCard(null)}
              onClick={() => handleFeatureClick(feature.path)}
            >
              <div style={{ position: 'relative' }}>
                {feature.badge && (
                  <Badge.Ribbon 
                    text={feature.badge} 
                    color={
                      feature.badge === 'HOT' ? 'red' : 
                      feature.badge === 'NEW' ? 'blue' :
                      feature.badge === 'BETA' ? 'orange' :
                      feature.badge === 'ALPHA' ? 'purple' : 'blue'
                    }
                  >
                    <div />
                  </Badge.Ribbon>
                )}
                
                <div style={{ display: 'flex', alignItems: 'center', marginBottom: '16px' }}>
                  <Avatar
                    size={48}
                    style={{ backgroundColor: feature.color, marginRight: '16px' }}
                    icon={feature.icon}
                  />
                  <div>
                    <Title level={3} style={{ margin: 0, color: feature.color }}>
                      {feature.title}
                    </Title>
                    <Text type="secondary">
                      {feature.key === 'image-generation' ? 
                        `${feature.stats.generations}次生成 · 平均${feature.stats.avgTime}` :
                        `${feature.stats.analyses}次分析 · ${feature.stats.accuracy}准确率`
                      }
                    </Text>
                  </div>
                </div>

                <Paragraph style={{ marginBottom: '16px', color: '#666' }}>
                  {feature.description}
                </Paragraph>

                <div style={{ marginBottom: '20px' }}>
                  <Text strong style={{ marginBottom: '8px', display: 'block' }}>核心功能：</Text>
                  <Space wrap>
                    {feature.features.map((feat, index) => (
                      <Badge key={index} count={feat} style={{ backgroundColor: feature.color }} />
                    ))}
                  </Space>
                </div>

                <Button 
                  type="primary" 
                  size="large" 
                  block
                  style={{ backgroundColor: feature.color, borderColor: feature.color }}
                  icon={<ThunderboltOutlined />}
                >
                  立即体验
                </Button>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 最近活动和快速操作 */}
      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card title="最近活动" extra={<Button type="link">查看全部</Button>}>
            <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
              {recentActivities.map((activity, index) => (
                <div key={index} style={{ marginBottom: '16px', display: 'flex', alignItems: 'center' }}>
                  <Avatar
                    size="small"
                    style={{ 
                      backgroundColor: activity.type === 'generation' ? '#1890ff' : '#52c41a',
                      marginRight: '12px'
                    }}
                    icon={activity.type === 'generation' ? <PictureOutlined /> : <EyeOutlined />}
                  />
                  <div style={{ flex: 1 }}>
                    <Text strong>{activity.user}</Text>
                    <Text style={{ marginLeft: '8px' }}>{activity.action}</Text>
                    <div>
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        <ClockCircleOutlined style={{ marginRight: '4px' }} />
                        {activity.time}
                      </Text>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="快速操作">
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Button 
                type="primary" 
                block 
                size="large"
                icon={<PictureOutlined />}
                onClick={() => navigate('/ai/image-generation')}
              >
                快速生成图像
              </Button>
              <Button 
                block 
                size="large"
                icon={<EyeOutlined />}
                onClick={() => navigate('/ai/image-analysis')}
              >
                上传图像分析
              </Button>
              <Divider />
              <div style={{ textAlign: 'center' }}>
                <StarOutlined style={{ color: '#faad14', marginRight: '8px' }} />
                <Text type="secondary">每日免费额度：50次</Text>
              </div>
              <div style={{ textAlign: 'center' }}>
                <Text type="secondary">今日已使用：12次</Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default AIMultimodal;