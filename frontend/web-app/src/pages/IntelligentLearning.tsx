import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Typography, 
  Space, 
  Button, 
  Row, 
  Col, 
  Statistic, 
  Progress, 
  List, 
  Avatar, 
  Tag, 
  Timeline,
  Alert,
  Carousel,
  Badge,
  Tooltip,
  Divider,
  Empty
} from 'antd';
import { 
  BookOutlined, 
  BulbOutlined, 
  TrophyOutlined, 
  RocketOutlined,
  ClockCircleOutlined,
  StarOutlined,
  FireOutlined,
  ThunderboltOutlined,
  UserOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  LineChartOutlined,
  BarChartOutlined,
  CalendarOutlined,
  GiftOutlined,
  CrownOutlined,
  HeartOutlined,
  TeamOutlined,
  AimOutlined,
  EyeOutlined,
  ArrowRightOutlined
} from '@ant-design/icons';
import { Line, Column, Pie } from '@ant-design/plots';
import { useNavigate } from 'react-router-dom';
import moment from 'moment';

const { Title, Paragraph, Text } = Typography;

interface LearningStats {
  totalCourses: number;
  completedCourses: number;
  totalStudyTime: number; // 小时
  currentStreak: number; // 连续学习天数
  totalAssessments: number;
  passedAssessments: number;
  averageScore: number;
  level: string;
  experience: number;
  nextLevelExp: number;
}

interface RecentActivity {
  id: string;
  type: 'course' | 'assessment' | 'achievement';
  title: string;
  description: string;
  timestamp: Date;
  icon: string;
  color: string;
}

interface Recommendation {
  id: string;
  type: 'course' | 'assessment' | 'practice';
  title: string;
  description: string;
  reason: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  estimatedTime: number; // 分钟
  thumbnail: string;
  tags: string[];
}

interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: string;
  earnedAt: Date;
  rarity: 'common' | 'rare' | 'epic' | 'legendary';
}

const IntelligentLearning: React.FC = () => {
  const navigate = useNavigate();
  const [learningStats, setLearningStats] = useState<LearningStats | null>(null);
  const [recentActivities, setRecentActivities] = useState<RecentActivity[]>([]);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadLearningData();
  }, []);

  const loadLearningData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockStats: LearningStats = {
        totalCourses: 25,
        completedCourses: 18,
        totalStudyTime: 156,
        currentStreak: 12,
        totalAssessments: 15,
        passedAssessments: 13,
        averageScore: 85,
        level: '学者',
        experience: 2850,
        nextLevelExp: 3000
      };

      const mockActivities: RecentActivity[] = [
        {
          id: '1',
          type: 'course',
          title: '完成课程：道德经精读',
          description: '深入学习了道德经的核心思想',
          timestamp: new Date('2024-02-01T14:30:00'),
          icon: '📚',
          color: '#1890ff'
        },
        {
          id: '2',
          type: 'assessment',
          title: '通过评估：儒家思想基础',
          description: '获得85分，表现优秀',
          timestamp: new Date('2024-01-31T16:45:00'),
          icon: '🏆',
          color: '#52c41a'
        },
        {
          id: '3',
          type: 'achievement',
          title: '获得成就：连续学习达人',
          description: '连续学习12天',
          timestamp: new Date('2024-01-30T09:15:00'),
          icon: '🔥',
          color: '#faad14'
        },
        {
          id: '4',
          type: 'course',
          title: '开始学习：禅修入门',
          description: '开始探索佛家修行之道',
          timestamp: new Date('2024-01-29T20:00:00'),
          icon: '🧘',
          color: '#722ed1'
        }
      ];

      const mockRecommendations: Recommendation[] = [
        {
          id: '1',
          type: 'course',
          title: '庄子逍遥游深度解析',
          description: '基于您对道家思想的兴趣，推荐深入学习庄子哲学',
          reason: '您在道德经课程中表现优秀',
          difficulty: 'intermediate',
          estimatedTime: 120,
          thumbnail: '/api/placeholder/300/200',
          tags: ['庄子', '道家', '哲学']
        },
        {
          id: '2',
          type: 'assessment',
          title: '佛学基础能力评估',
          description: '测试您对佛学基本概念的掌握程度',
          reason: '您最近开始学习禅修课程',
          difficulty: 'beginner',
          estimatedTime: 30,
          thumbnail: '/api/placeholder/300/200',
          tags: ['佛学', '基础', '评估']
        },
        {
          id: '3',
          type: 'practice',
          title: '传统文化现代应用练习',
          description: '将传统智慧应用到现代生活场景中',
          reason: '提升实践应用能力',
          difficulty: 'advanced',
          estimatedTime: 45,
          thumbnail: '/api/placeholder/300/200',
          tags: ['实践', '应用', '现代']
        }
      ];

      const mockAchievements: Achievement[] = [
        {
          id: '1',
          title: '学习新手',
          description: '完成第一门课程',
          icon: '🌱',
          earnedAt: new Date('2024-01-15'),
          rarity: 'common'
        },
        {
          id: '2',
          title: '道家学者',
          description: '在道家思想评估中获得优秀成绩',
          icon: '🏆',
          earnedAt: new Date('2024-01-25'),
          rarity: 'rare'
        },
        {
          id: '3',
          title: '连续学习达人',
          description: '连续学习12天',
          icon: '🔥',
          earnedAt: new Date('2024-01-30'),
          rarity: 'epic'
        }
      ];

      setLearningStats(mockStats);
      setRecentActivities(mockActivities);
      setRecommendations(mockRecommendations);
      setAchievements(mockAchievements);
      setLoading(false);
    }, 1000);
  };

  // 获取学习进度数据
  const getProgressData = () => {
    if (!learningStats) return [];
    
    const days = [];
    for (let i = 6; i >= 0; i--) {
      const date = moment().subtract(i, 'days');
      days.push({
        date: date.format('MM-DD'),
        hours: Math.random() * 3 + 1 // 模拟数据
      });
    }
    return days;
  };

  // 获取分类学习数据
  const getCategoryData = () => [
    { category: '道家思想', value: 35, color: '#1890ff' },
    { category: '儒家思想', value: 28, color: '#52c41a' },
    { category: '佛家思想', value: 22, color: '#faad14' },
    { category: '现代应用', value: 15, color: '#722ed1' }
  ];

  // 获取难度颜色
  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case 'beginner': return 'green';
      case 'intermediate': return 'orange';
      case 'advanced': return 'red';
      default: return 'default';
    }
  };

  // 获取稀有度颜色
  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case 'common': return '#d9d9d9';
      case 'rare': return '#1890ff';
      case 'epic': return '#722ed1';
      case 'legendary': return '#faad14';
      default: return '#d9d9d9';
    }
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <BulbOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          智能学习系统
        </Title>
        <Paragraph>
          基于AI的个性化学习平台，为您提供智能化的学习体验和全方位的能力提升
        </Paragraph>
      </div>

      {/* 学习统计概览 */}
      {learningStats && (
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="学习进度"
                value={learningStats.completedCourses}
                suffix={`/ ${learningStats.totalCourses}`}
                prefix={<BookOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
              <Progress 
                percent={(learningStats.completedCourses / learningStats.totalCourses) * 100}
                size="small"
                style={{ marginTop: '8px' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="学习时长"
                value={learningStats.totalStudyTime}
                suffix="小时"
                prefix={<ClockCircleOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
              <Text type="secondary" style={{ fontSize: '12px' }}>
                本月已学习 {Math.floor(learningStats.totalStudyTime * 0.3)} 小时
              </Text>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="连续学习"
                value={learningStats.currentStreak}
                suffix="天"
                prefix={<FireOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
              <Text type="secondary" style={{ fontSize: '12px' }}>
                继续保持，冲击新纪录！
              </Text>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="平均分数"
                value={learningStats.averageScore}
                suffix="分"
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
              <Text type="secondary" style={{ fontSize: '12px' }}>
                {learningStats.passedAssessments}/{learningStats.totalAssessments} 次通过
              </Text>
            </Card>
          </Col>
        </Row>
      )}

      {/* 等级和经验 */}
      {learningStats && (
        <Card style={{ marginBottom: '24px' }}>
          <Row align="middle">
            <Col flex="auto">
              <Space align="center">
                <Avatar size={64} style={{ backgroundColor: '#1890ff' }}>
                  <CrownOutlined style={{ fontSize: '32px' }} />
                </Avatar>
                <div>
                  <Title level={4} style={{ margin: 0 }}>
                    {learningStats.level}
                  </Title>
                  <Text type="secondary">
                    经验值: {learningStats.experience} / {learningStats.nextLevelExp}
                  </Text>
                  <Progress 
                    percent={(learningStats.experience / learningStats.nextLevelExp) * 100}
                    size="small"
                    style={{ width: '200px', marginTop: '4px' }}
                  />
                </div>
              </Space>
            </Col>
            <Col>
              <Space>
                <Button 
                  type="primary" 
                  icon={<AimOutlined />}
                  onClick={() => navigate('/learning/learning-progress')}
                >
                  查看详细进度
                </Button>
                <Button 
                  icon={<GiftOutlined />}
                >
                  每日签到
                </Button>
              </Space>
            </Col>
          </Row>
        </Card>
      )}

      {/* 快速入口 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={8}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/learning/learning-progress')}
          >
            <div style={{ fontSize: '48px', color: '#1890ff', marginBottom: '16px' }}>
              <LineChartOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>学习进度</Title>
            <Text type="secondary">查看详细的学习进度和统计</Text>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/learning/course-center')}
          >
            <div style={{ fontSize: '48px', color: '#52c41a', marginBottom: '16px' }}>
              <BookOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>课程中心</Title>
            <Text type="secondary">浏览和学习各类精品课程</Text>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/learning/ability-assessment')}
          >
            <div style={{ fontSize: '48px', color: '#faad14', marginBottom: '16px' }}>
              <TrophyOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>能力评估</Title>
            <Text type="secondary">测试和评估您的学习成果</Text>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 学习数据可视化 */}
        <Col xs={24} lg={12}>
          <Card title="学习趋势" extra={<BarChartOutlined />} style={{ marginBottom: '16px' }}>
            <Line
              data={getProgressData()}
              xField="date"
              yField="hours"
              height={200}
              smooth
              point={{
                size: 3,
                shape: 'circle'
              }}
              color="#1890ff"
            />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="学习分布" extra={<BarChartOutlined />} style={{ marginBottom: '16px' }}>
            <Pie
              data={getCategoryData()}
              angleField="value"
              colorField="category"
              height={200}
              radius={0.8}
              label={{
                type: 'outer',
                content: '{name} {percentage}'
              }}
            />
          </Card>
        </Col>

        {/* 智能推荐 */}
        <Col xs={24} lg={12}>
          <Card 
            title="智能推荐" 
            extra={
              <Button type="link" onClick={() => navigate('/learning/course-center')}>
                查看更多 <ArrowRightOutlined />
              </Button>
            }
            style={{ marginBottom: '16px' }}
          >
            <List
              dataSource={recommendations}
              renderItem={item => (
                <List.Item
                  actions={[
                    <Button 
                      type="primary" 
                      size="small"
                      icon={item.type === 'course' ? <PlayCircleOutlined /> : <EyeOutlined />}
                    >
                      {item.type === 'course' ? '开始学习' : 
                       item.type === 'assessment' ? '开始评估' : '开始练习'}
                    </Button>
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <Avatar 
                        shape="square" 
                        size={48}
                        src={item.thumbnail}
                      />
                    }
                    title={
                      <div>
                        <Text strong>{item.title}</Text>
                        <div style={{ marginTop: '4px' }}>
                          <Space size="small">
                            <Tag size="small" color={getDifficultyColor(item.difficulty)}>
                              {item.difficulty === 'beginner' ? '初级' :
                               item.difficulty === 'intermediate' ? '中级' : '高级'}
                            </Tag>
                            <Tag size="small">
                              <ClockCircleOutlined /> {item.estimatedTime}分钟
                            </Tag>
                          </Space>
                        </div>
                      </div>
                    }
                    description={
                      <div>
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                          {item.description}
                        </Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: '11px', fontStyle: 'italic' }}>
                          推荐理由: {item.reason}
                        </Text>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* 最近活动 */}
        <Col xs={24} lg={12}>
          <Card title="最近活动" style={{ marginBottom: '16px' }}>
            <Timeline>
              {recentActivities.map(activity => (
                <Timeline.Item
                  key={activity.id}
                  color={activity.color}
                  dot={
                    <div style={{ 
                      fontSize: '16px', 
                      display: 'flex', 
                      alignItems: 'center', 
                      justifyContent: 'center',
                      width: '24px',
                      height: '24px'
                    }}>
                      {activity.icon}
                    </div>
                  }
                >
                  <div>
                    <Text strong>{activity.title}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {activity.description}
                    </Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: '11px' }}>
                      {moment(activity.timestamp).fromNow()}
                    </Text>
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </Card>
        </Col>

        {/* 成就展示 */}
        <Col xs={24}>
          <Card title="最新成就" extra={<TrophyOutlined />}>
            <Row gutter={[16, 16]}>
              {achievements.map(achievement => (
                <Col xs={12} sm={8} md={6} lg={4} key={achievement.id}>
                  <Tooltip title={achievement.description}>
                    <Card 
                      size="small" 
                      style={{ 
                        textAlign: 'center',
                        borderColor: getRarityColor(achievement.rarity),
                        borderWidth: '2px'
                      }}
                      bodyStyle={{ padding: '16px 8px' }}
                    >
                      <div style={{ fontSize: '32px', marginBottom: '8px' }}>
                        {achievement.icon}
                      </div>
                      <Text strong style={{ fontSize: '12px', display: 'block' }}>
                        {achievement.title}
                      </Text>
                      <Tag 
                        size="small" 
                        color={getRarityColor(achievement.rarity)}
                        style={{ marginTop: '4px', fontSize: '10px' }}
                      >
                        {achievement.rarity === 'common' ? '普通' :
                         achievement.rarity === 'rare' ? '稀有' :
                         achievement.rarity === 'epic' ? '史诗' : '传说'}
                      </Tag>
                      <div style={{ marginTop: '4px' }}>
                        <Text type="secondary" style={{ fontSize: '10px' }}>
                          {moment(achievement.earnedAt).format('MM-DD')}
                        </Text>
                      </div>
                    </Card>
                  </Tooltip>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 学习提醒 */}
      <Alert
        message="学习提醒"
        description="您今天还没有完成学习目标，建议花费30分钟继续学习以保持连续学习记录。"
        type="info"
        showIcon
        style={{ marginTop: '24px' }}
        action={
          <Button size="small" type="primary" onClick={() => navigate('/learning/course-center')}>
            立即学习
          </Button>
        }
      />
    </div>
  );
};

export default IntelligentLearning;