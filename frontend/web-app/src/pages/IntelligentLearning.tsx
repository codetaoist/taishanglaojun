import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Typography,
  Space,
  Tag,
  Progress,
  Timeline,
  List,
  Avatar,
  Tooltip,
  Statistic,
  Alert,
  Spin,
  message,
  Badge,
  Divider,
  Empty,
  Skeleton
} from 'antd';
import {
  BookOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  FireOutlined,
  LineChartOutlined,
  BarChartOutlined,
  BulbOutlined,
  PlayCircleOutlined,
  EyeOutlined,
  ArrowRightOutlined,
  AimOutlined,
  GiftOutlined,
  CrownOutlined,
  StarOutlined,
  ThunderboltOutlined,
  RocketOutlined,
  HeartOutlined,
  CalendarOutlined,
  UserOutlined,
  SettingOutlined,
  ReloadOutlined,
  RiseOutlined,
  FallOutlined,
  MinusOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { Line, Pie, Column } from '@ant-design/plots';
import moment from 'moment';
import { learningApi } from '../services/learningApi';
import type { 
  LearnerProfile, 
  LearningAnalytics, 
  Recommendation, 
  Achievement,
  ActivityData,
  SkillProgress
} from '../services/learningApi';

const { Title, Text, Paragraph } = Typography;

const IntelligentLearning: React.FC = () => {
  const navigate = useNavigate();
  
  // 状态管理
  const [learnerProfile, setLearnerProfile] = useState<LearnerProfile | null>(null);
  const [learningAnalytics, setLearningAnalytics] = useState<LearningAnalytics | null>(null);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [skillProgress, setSkillProgress] = useState<SkillProgress[]>([]);
  const [weeklyActivity, setWeeklyActivity] = useState<ActivityData[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    loadLearningData();
  }, []);

  // 加载学习数据
  const loadLearningData = async () => {
    try {
      setLoading(true);
      
      // 并行加载所有数据
      const [
        profileResponse,
        analyticsResponse,
        recommendationsResponse,
        achievementsResponse,
        skillsResponse,
        weeklyActivityResponse
      ] = await Promise.all([
        learningApi.getLearnerProfile().catch(() => ({ success: false, data: null })),
        learningApi.getLearningAnalytics('current', '7d').catch(() => ({ success: false, data: null })),
        learningApi.getRecommendations('current').catch(() => ({ success: false, data: [] })),
        learningApi.getAchievements('current').catch(() => ({ success: false, data: [] })),
        learningApi.getSkillProgress('current').catch(() => ({ success: false, data: [] })),
        learningApi.getWeeklyActivity().catch(() => ({ success: false, data: [] }))
      ]);

      // 设置学习者档案
      if (profileResponse.success && profileResponse.data) {
        setLearnerProfile(profileResponse.data);
      } else {
        // 使用模拟数据
        setLearnerProfile({
          id: 'mock-user',
          userId: 'mock-user',
          name: '智慧学者',
          email: 'learner@example.com',
          level: '学者',
          experience: 2850,
          nextLevelExp: 3000,
          learningGoals: ['道家思想', '儒家思想', '佛家思想'],
          preferences: {
            learningStyle: 'visual',
            difficulty: 'intermediate',
            topics: ['哲学', '传统文化', '现代应用'],
            studyTime: 60,
            reminderEnabled: true,
            reminderTime: '20:00'
          },
          skills: [],
          achievements: [],
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-02-01T00:00:00Z'
        });
      }

      // 设置学习分析数据
      if (analyticsResponse.success && analyticsResponse.data) {
        setLearningAnalytics(analyticsResponse.data);
        setWeeklyActivity(analyticsResponse.data.weeklyActivity || []);
      } else {
        // 使用模拟数据
        const mockAnalytics: LearningAnalytics = {
          learnerId: 'mock-user',
          totalStudyTime: 156,
          coursesCompleted: 18,
          currentStreak: 12,
          averageScore: 85,
          skillProgress: [],
          weeklyActivity: generateMockWeeklyActivity(),
          monthlyProgress: [],
          recommendations: []
        };
        setLearningAnalytics(mockAnalytics);
        setWeeklyActivity(mockAnalytics.weeklyActivity);
      }

      // 设置推荐数据
      if (recommendationsResponse.success && recommendationsResponse.data) {
        setRecommendations(recommendationsResponse.data);
      } else {
        // 使用模拟推荐数据
        setRecommendations([
          {
            id: '1',
            type: 'course',
            title: '庄子逍遥游深度解析',
            description: '基于您对道家思想的兴趣，推荐深入学习庄子哲学',
            reason: '您在道德经课程中表现优秀',
            confidence: 95,
            priority: 'high',
            difficulty: 'intermediate',
            estimatedTime: 120,
            tags: ['庄子', '道家', '哲学'],
            thumbnail: '/api/placeholder/300/200'
          },
          {
            id: '2',
            type: 'assessment',
            title: '佛学基础能力评估',
            description: '测试您对佛学基本概念的掌握程度',
            reason: '您最近开始学习禅修课程',
            confidence: 88,
            priority: 'medium',
            difficulty: 'beginner',
            estimatedTime: 30,
            tags: ['佛学', '基础', '评估']
          },
          {
            id: '3',
            type: 'skill_practice',
            title: '传统文化现代应用练习',
            description: '将传统智慧应用到现代生活场景中',
            reason: '提升实践应用能力',
            confidence: 82,
            priority: 'medium',
            difficulty: 'advanced',
            estimatedTime: 45,
            tags: ['实践', '应用', '现代']
          }
        ]);
      }

      // 设置成就数据
      if (achievementsResponse.success && achievementsResponse.data) {
        setAchievements(achievementsResponse.data);
      } else {
        // 使用模拟成就数据
        setAchievements([
          {
            id: '1',
            title: '学习新手',
            description: '完成第一门课程',
            icon: '🌱',
            earnedAt: '2024-01-15T00:00:00Z',
            rarity: 'common'
          },
          {
            id: '2',
            title: '道家学者',
            description: '在道家思想评估中获得优秀成绩',
            icon: '🏆',
            earnedAt: '2024-01-25T00:00:00Z',
            rarity: 'rare'
          },
          {
            id: '3',
            title: '连续学习达人',
            description: '连续学习12天',
            icon: '🔥',
            earnedAt: '2024-01-30T00:00:00Z',
            rarity: 'epic'
          }
        ]);
      }

      // 设置技能进度数据
      if (skillsResponse.success && skillsResponse.data) {
        setSkillProgress(skillsResponse.data);
      } else {
        // 使用模拟技能数据
        setSkillProgress([
          { skillId: '1', skillName: '道家思想', currentLevel: 3, progress: 75, trend: 'up' },
          { skillId: '2', skillName: '儒家思想', currentLevel: 2, progress: 60, trend: 'up' },
          { skillId: '3', skillName: '佛家思想', currentLevel: 2, progress: 45, trend: 'stable' },
          { skillId: '4', skillName: '现代应用', currentLevel: 1, progress: 30, trend: 'up' }
        ]);
      }

      // 设置周活动数据
      if (weeklyActivityResponse.success && weeklyActivityResponse.data) {
        setWeeklyActivity(weeklyActivityResponse.data);
      }

    } catch (error) {
      console.error('加载学习数据失败:', error);
      message.error('加载学习数据失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 刷新数据
  const refreshData = async () => {
    setRefreshing(true);
    await loadLearningData();
    setRefreshing(false);
    message.success('数据已刷新');
  };

  // 生成模拟周活动数据
  const generateMockWeeklyActivity = (): ActivityData[] => {
    const days = [];
    for (let i = 6; i >= 0; i--) {
      const date = moment().subtract(i, 'days');
      days.push({
        date: date.format('YYYY-MM-DD'),
        studyTime: Math.floor(Math.random() * 180 + 30), // 30-210分钟
        coursesCompleted: Math.floor(Math.random() * 3),
        exercisesCompleted: Math.floor(Math.random() * 5 + 1)
      });
    }
    return days;
  };

  // 获取学习进度数据（用于图表）
  const getProgressChartData = () => {
    return weeklyActivity.map(activity => ({
      date: moment(activity.date).format('MM-DD'),
      studyTime: Math.round(activity.studyTime / 60 * 10) / 10, // 转换为小时，保留1位小数
      coursesCompleted: activity.coursesCompleted
    }));
  };

  // 获取技能分布数据
  const getSkillDistributionData = () => {
    return skillProgress.map(skill => ({
      skill: skill.skillName,
      progress: skill.progress,
      level: skill.currentLevel
    }));
  };

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

  // 获取优先级颜色
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return '#ff4d4f';
      case 'medium': return '#faad14';
      case 'low': return '#52c41a';
      default: return '#d9d9d9';
    }
  };

  // 获取趋势图标
  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return <RiseOutlined style={{ color: '#52c41a' }} />;
      case 'down': return <FallOutlined style={{ color: '#ff4d4f' }} />;
      case 'stable': return <MinusOutlined style={{ color: '#faad14' }} />;
      default: return null;
    }
  };

  // 格式化学习时间
  const formatStudyTime = (minutes: number) => {
    if (minutes < 60) {
      return `${minutes}分钟`;
    }
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    return remainingMinutes > 0 ? `${hours}小时${remainingMinutes}分钟` : `${hours}小时`;
  };

  if (loading) {
    return (
      <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
        <Skeleton active />
        <Row gutter={[16, 16]} style={{ marginTop: '24px' }}>
          {[1, 2, 3, 4].map(i => (
            <Col xs={24} sm={12} lg={6} key={i}>
              <Card>
                <Skeleton.Input style={{ width: '100%' }} active />
              </Card>
            </Col>
          ))}
        </Row>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <Title level={2} style={{ margin: 0 }}>
            <BulbOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
            智能学习系统
          </Title>
          <Paragraph style={{ margin: '8px 0 0 0', color: '#666' }}>
            基于AI的个性化学习平台，为您提供智能化的学习体验和全方位的能力提升
          </Paragraph>
        </div>
        <Space>
          <Button 
            icon={<SettingOutlined />} 
            onClick={() => navigate('/profile')}
          >
            设置
          </Button>
          <Button 
            icon={<ReloadOutlined />} 
            loading={refreshing}
            onClick={refreshData}
          >
            刷新
          </Button>
        </Space>
      </div>

      {/* 学习统计概览 */}
      {learningAnalytics && (
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} lg={6}>
            <Card hoverable>
              <Statistic
                title="完成课程"
                value={learningAnalytics.coursesCompleted}
                prefix={<BookOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  本月新增 {Math.floor(learningAnalytics.coursesCompleted * 0.2)} 门
                </Text>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card hoverable>
              <Statistic
                title="学习时长"
                value={learningAnalytics.totalStudyTime}
                suffix="小时"
                prefix={<ClockCircleOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  本周已学习 {Math.floor(weeklyActivity.reduce((sum, day) => sum + day.studyTime, 0) / 60)} 小时
                </Text>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card hoverable>
              <Statistic
                title="连续学习"
                value={learningAnalytics.currentStreak}
                suffix="天"
                prefix={<FireOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  继续保持，冲击新纪录！
                </Text>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card hoverable>
              <Statistic
                title="平均分数"
                value={learningAnalytics.averageScore}
                suffix="分"
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  超过 85% 的学习者
                </Text>
              </div>
            </Card>
          </Col>
        </Row>
      )}

      {/* 学习者档案和等级 */}
      {learnerProfile && (
        <Card style={{ marginBottom: '24px' }}>
          <Row align="middle">
            <Col flex="auto">
              <Space align="center" size="large">
                <Badge count={achievements.length} showZero={false} offset={[-8, 8]}>
                  <Avatar size={64} style={{ backgroundColor: '#1890ff' }} src={learnerProfile.avatar}>
                    {learnerProfile.avatar ? null : <UserOutlined style={{ fontSize: '32px' }} />}
                  </Avatar>
                </Badge>
                <div>
                  <Space align="center">
                    <Title level={4} style={{ margin: 0 }}>
                      {learnerProfile.name}
                    </Title>
                    <Tag color="gold" icon={<CrownOutlined />}>
                      {learnerProfile.level}
                    </Tag>
                  </Space>
                  <div style={{ marginTop: '8px' }}>
                    <Text type="secondary">
                      经验值: {learnerProfile.experience} / {learnerProfile.nextLevelExp}
                    </Text>
                    <Progress 
                      percent={(learnerProfile.experience / learnerProfile.nextLevelExp) * 100}
                      size="small"
                      style={{ width: '300px', marginTop: '4px' }}
                      strokeColor={{
                        '0%': '#108ee9',
                        '100%': '#87d068',
                      }}
                    />
                  </div>
                  <div style={{ marginTop: '8px' }}>
                    <Space size="small">
                      {learnerProfile.learningGoals.slice(0, 3).map((goal, index) => (
                        <Tag key={index} size="small" color="blue">
                          <AimOutlined style={{ marginRight: '4px' }} />
                          {goal}
                        </Tag>
                      ))}
                    </Space>
                  </div>
                </div>
              </Space>
            </Col>
            <Col>
              <Space direction="vertical" size="small">
                <Button 
                  type="primary" 
                  icon={<AimOutlined />}
                  onClick={() => navigate('/learning/learning-progress')}
                  block
                >
                  查看详细进度
                </Button>
                <Button 
                  icon={<LineChartOutlined />}
                  onClick={() => navigate('/learning/analytics-dashboard')}
                  block
                >
                  分析仪表板
                </Button>
                <Button 
                  icon={<GiftOutlined />}
                  block
                >
                  每日签到
                </Button>
                <Button 
                  icon={<CalendarOutlined />}
                  onClick={() => navigate('/course-center')}
                  block
                >
                  学习计划
                </Button>
              </Space>
            </Col>
          </Row>
        </Card>
      )}

      {/* 快速入口 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={8} xl={6}>
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
        <Col xs={24} sm={12} lg={8} xl={6}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/course-center')}
          >
            <div style={{ fontSize: '48px', color: '#52c41a', marginBottom: '16px' }}>
              <BookOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>课程中心</Title>
            <Text type="secondary">浏览和学习各类精品课程</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8} xl={6}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/learning-plan')}
          >
            <div style={{ fontSize: '48px', color: '#722ed1', marginBottom: '16px' }}>
              <CalendarOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>学习计划</Title>
            <Text type="secondary">制定个性化学习路径</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8} xl={6}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/ability-assessment')}
          >
            <div style={{ fontSize: '48px', color: '#faad14', marginBottom: '16px' }}>
              <TrophyOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>能力评估</Title>
            <Text type="secondary">测试和评估您的学习成果</Text>
          </Card>
        </Col>
      </Row>
      
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={8} xl={6}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/daily-checkin')}
          >
            <div style={{ fontSize: '48px', color: '#f5222d', marginBottom: '16px' }}>
              <CheckCircleOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>每日签到</Title>
            <Text type="secondary">坚持签到获得奖励</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8} xl={6}>
          <Card
            hoverable
            style={{ textAlign: 'center', height: '200px' }}
            bodyStyle={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
            onClick={() => navigate('/achievement-center')}
          >
            <div style={{ fontSize: '48px', color: '#722ed1', marginBottom: '16px' }}>
              <TrophyOutlined />
            </div>
            <Title level={4} style={{ margin: 0 }}>成就中心</Title>
            <Text type="secondary">查看学习成就和荣誉</Text>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 学习数据可视化 */}
        <Col xs={24} lg={12}>
          <Card 
            title="学习趋势" 
            extra={
              <Space>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  最近7天
                </Text>
                <LineChartOutlined />
              </Space>
            } 
            style={{ marginBottom: '16px' }}
          >
            {weeklyActivity.length > 0 ? (
              <Line
                data={getProgressChartData()}
                xField="date"
                yField="studyTime"
                height={200}
                smooth
                point={{
                  size: 4,
                  shape: 'circle'
                }}
                color="#1890ff"
                tooltip={{
                  formatter: (datum) => ({
                    name: '学习时长',
                    value: `${datum.studyTime}小时`
                  })
                }}
                annotations={[
                  {
                    type: 'line',
                    start: ['min', 'mean'],
                    end: ['max', 'mean'],
                    style: {
                      stroke: '#faad14',
                      lineDash: [4, 4],
                    },
                  },
                ]}
              />
            ) : (
              <Empty description="暂无学习数据" />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card 
            title="技能进度" 
            extra={
              <Space>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  当前掌握情况
                </Text>
                <BarChartOutlined />
              </Space>
            } 
            style={{ marginBottom: '16px' }}
          >
            {skillProgress.length > 0 ? (
              <Column
                data={getSkillDistributionData()}
                xField="skill"
                yField="progress"
                height={200}
                color="#52c41a"
                columnWidthRatio={0.6}
                tooltip={{
                  formatter: (datum) => ({
                    name: '掌握程度',
                    value: `${datum.progress}% (等级${datum.level})`
                  })
                }}
                label={{
                  position: 'top',
                  formatter: (datum) => `${datum.progress}%`
                }}
              />
            ) : (
              <Empty description="暂无技能数据" />
            )}
          </Card>
        </Col>

        {/* 智能推荐 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <BulbOutlined />
                智能推荐
                <Badge count={recommendations.length} showZero color="#52c41a" />
              </Space>
            }
            extra={
              <Button 
                type="link" 
                size="small" 
                icon={<ReloadOutlined />}
                onClick={refreshData}
              >
                刷新推荐
              </Button>
            }
            style={{ marginBottom: '16px' }}
          >
            {recommendations.length > 0 ? (
              <List
                dataSource={recommendations}
                renderItem={(item) => (
                  <List.Item
                    actions={[
                      <Button 
                        type="primary" 
                        size="small" 
                        icon={<PlayCircleOutlined />}
                        style={{ borderRadius: '6px' }}
                      >
                        开始学习
                      </Button>,
                      <Button 
                        type="text" 
                        size="small" 
                        icon={<StarOutlined />}
                        style={{ color: '#faad14' }}
                      >
                        收藏
                      </Button>
                    ]}
                  >
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          src={item.thumbnail} 
                          size={56}
                          icon={<BookOutlined />}
                          style={{ 
                            backgroundColor: '#f0f2f5',
                            border: '2px solid #fff',
                            boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
                          }}
                        />
                      }
                      title={
                        <Space wrap>
                          <Text strong style={{ fontSize: '14px' }}>
                            {item.title}
                          </Text>
                          <Tag color={getPriorityColor(item.priority)}>
                            {item.priority === 'high' ? '高优先级' : 
                             item.priority === 'medium' ? '中优先级' : '低优先级'}
                          </Tag>
                          <Tag color="blue">
                            匹配度 {Math.round(item.confidence)}%
                          </Tag>
                        </Space>
                      }
                      description={
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Text type="secondary" style={{ fontSize: '13px' }}>
                            {item.description}
                          </Text>
                          <Space wrap>
                            <Text type="secondary" style={{ fontSize: '12px' }}>
                              <ClockCircleOutlined /> 预计 {item.estimatedTime} 分钟
                            </Text>
                            <Text type="secondary" style={{ fontSize: '12px' }}>
                              <StarOutlined /> 推荐指数 {item.confidence}
                            </Text>
                            <Text type="secondary" style={{ fontSize: '12px' }}>
                              <BookOutlined /> {item.type === 'course' ? '课程' : 
                                                item.type === 'assessment' ? '评估' : '练习'}
                            </Text>
                          </Space>
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            ) : (
              <Empty 
                description="暂无推荐内容"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Card>
        </Col>

        {/* 最近活动 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <ClockCircleOutlined />
                最近活动
                <Badge 
                  count={weeklyActivity.reduce((sum, day) => sum + day.coursesCompleted + day.exercisesCompleted, 0)} 
                  showZero 
                  color="#1890ff" 
                />
              </Space>
            }
            extra={
              <Button 
                type="link" 
                size="small"
                onClick={() => navigate('/learning/learning-progress')}
              >
                查看全部 <ArrowRightOutlined />
              </Button>
            }
            style={{ marginBottom: '16px' }}
          >
            {weeklyActivity.length > 0 ? (
              <Timeline>
                {weeklyActivity.slice(0, 5).map((day, index) => (
                  <Timeline.Item
                    key={index}
                    dot={
                      <Avatar 
                        size="small" 
                        icon={
                          day.studyTime > 120 ? <FireOutlined /> :
                          day.studyTime > 60 ? <BookOutlined /> :
                          <ClockCircleOutlined />
                        }
                        style={{
                          backgroundColor: 
                            day.studyTime > 120 ? '#52c41a' :
                            day.studyTime > 60 ? '#1890ff' :
                            '#faad14'
                        }}
                      />
                    }
                    color={
                      day.studyTime > 120 ? 'green' :
                      day.studyTime > 60 ? 'blue' :
                      'orange'
                    }
                  >
                    <div>
                      <Space>
                        <Text strong>{moment(day.date).format('MM月DD日')}</Text>
                        <Tag color={day.studyTime > 120 ? 'success' : day.studyTime > 60 ? 'processing' : 'warning'}>
                          {formatStudyTime(day.studyTime)}
                        </Tag>
                      </Space>
                      <br />
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        完成 {day.coursesCompleted} 门课程，{day.exercisesCompleted} 个练习
                      </Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: '11px' }}>
                        {moment(day.date).fromNow()}
                      </Text>
                    </div>
                  </Timeline.Item>
                ))}
              </Timeline>
            ) : (
              <Empty 
                description="暂无活动记录"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Card>
        </Col>

        {/* 成就展示 */}
        <Col xs={24}>
          <Card 
            title={
              <Space>
                <TrophyOutlined />
                最新成就
                <Badge count={achievements.length} showZero color="#faad14" />
              </Space>
            }
            extra={
              <Button 
                type="link" 
                size="small"
                onClick={() => navigate('/learning/achievements')}
              >
                查看全部 <ArrowRightOutlined />
              </Button>
            }
            style={{ marginBottom: '16px' }}
          >
            {achievements.length > 0 ? (
              <Row gutter={[16, 16]}>
                {achievements.slice(0, 8).map((achievement) => (
                  <Col xs={12} sm={8} md={6} lg={4} xl={3} key={achievement.id}>
                    <Tooltip 
                      title={
                        <div>
                          <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>
                            {achievement.title}
                          </div>
                          <div style={{ marginBottom: '8px' }}>
                            {achievement.description}
                          </div>
                          <div style={{ fontSize: '12px', opacity: 0.8 }}>
                            获得时间: {moment(achievement.earnedAt).format('YYYY-MM-DD HH:mm')}
                          </div>
                        </div>
                      }
                      placement="top"
                    >
                      <Card
                        hoverable
                        size="small"
                        style={{
                          textAlign: 'center',
                          borderColor: achievement.rarity === 'legendary' ? '#faad14' :
                                      achievement.rarity === 'epic' ? '#722ed1' :
                                      achievement.rarity === 'rare' ? '#1890ff' : '#52c41a',
                          borderWidth: '2px',
                          background: achievement.rarity === 'legendary' ? 
                            'linear-gradient(135deg, #fff7e6 0%, #fff1b8 100%)' :
                            achievement.rarity === 'epic' ?
                            'linear-gradient(135deg, #f9f0ff 0%, #efdbff 100%)' :
                            'linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%)',
                          position: 'relative',
                          overflow: 'hidden'
                        }}
                        bodyStyle={{ padding: '16px 8px' }}
                      >
                        {achievement.rarity === 'legendary' && (
                          <div style={{
                            position: 'absolute',
                            top: '-10px',
                            right: '-10px',
                            width: '20px',
                            height: '20px',
                            background: 'linear-gradient(45deg, #faad14, #ffd666)',
                            borderRadius: '50%',
                            animation: 'pulse 2s infinite'
                          }} />
                        )}
                        <div style={{ 
                          fontSize: '32px', 
                          marginBottom: '8px',
                          filter: achievement.rarity === 'legendary' ? 'drop-shadow(0 0 8px #faad14)' : 'none'
                        }}>
                          {achievement.icon}
                        </div>
                        <Text strong style={{ 
                          fontSize: '11px', 
                          display: 'block',
                          marginBottom: '4px'
                        }}>
                          {achievement.title}
                        </Text>
                        <Tag 
                          size="small" 
                          color={
                            achievement.rarity === 'legendary' ? 'gold' :
                            achievement.rarity === 'epic' ? 'purple' :
                            achievement.rarity === 'rare' ? 'blue' : 'green'
                          }
                        >
                          {achievement.rarity === 'common' ? '普通' :
                           achievement.rarity === 'rare' ? '稀有' :
                           achievement.rarity === 'epic' ? '史诗' : '传说'}
                        </Tag>
                        <div style={{ 
                          fontSize: '10px', 
                          color: '#666', 
                          marginTop: '4px' 
                        }}>
                          {moment(achievement.earnedAt).format('MM-DD')}
                        </div>
                      </Card>
                    </Tooltip>
                  </Col>
                ))}
              </Row>
            ) : (
              <Empty 
                description="暂无成就记录"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* 学习提醒 */}
      {learningAnalytics && (
        <Alert
          message={
            <Space>
              <BulbOutlined />
              智能学习提醒
            </Space>
          }
          description={
            <div>
              {learningAnalytics.dailyGoalProgress < 100 ? (
                <div>
                  <Text>
                    您今天的学习进度为 <Text strong>{learningAnalytics.dailyGoalProgress}%</Text>，
                    还需要学习 <Text strong>{Math.ceil((100 - learningAnalytics.dailyGoalProgress) * 0.6)} 分钟</Text> 
                    即可完成今日目标。
                  </Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    建议优先学习推荐课程以获得更好的学习效果
                  </Text>
                </div>
              ) : (
                <div>
                  <Text>
                    🎉 恭喜！您已完成今日学习目标，当前连续学习 <Text strong>{learningAnalytics.currentStreak}</Text> 天。
                  </Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    继续保持，向更高的学习目标挑战！
                  </Text>
                </div>
              )}
            </div>
          }
          type={learningAnalytics.dailyGoalProgress < 100 ? "warning" : "success"}
          showIcon
          style={{ marginTop: '24px' }}
          action={
            <Space>
              {learningAnalytics.dailyGoalProgress < 100 ? (
                <Button 
                  size="small" 
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={() => navigate('/learning/course-center')}
                >
                  开始学习
                </Button>
              ) : (
                <Button 
                  size="small" 
                  type="primary"
                  icon={<TrophyOutlined />}
                  onClick={() => navigate('/learning/achievements')}
                >
                  查看成就
                </Button>
              )}
              <Button 
                size="small"
                icon={<SettingOutlined />}
                onClick={() => message.info('学习设置功能开发中...')}
              >
                设置目标
              </Button>
            </Space>
          }
          closable
        />
      )}
    </div>
  );
};

export default IntelligentLearning;