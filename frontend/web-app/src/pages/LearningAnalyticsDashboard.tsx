import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Select,
  DatePicker,
  Space,
  Typography,
  Button,
  Badge,
  Tag,
  Tooltip,
  Alert,
  Spin,
  Empty,
  message,
  Divider,
  Avatar,
  List,
  Timeline
} from 'antd';
import {
  LineChartOutlined,
  BarChartOutlined,
  PieChartOutlined,
  TrophyOutlined,
  FireOutlined,
  ClockCircleOutlined,
  BookOutlined,
  UserOutlined,
  CalendarOutlined,
  ReloadOutlined,
  DownloadOutlined,
  ShareAltOutlined,
  SettingOutlined,
  BulbOutlined,
  RiseOutlined,
  FallOutlined,
  MinusOutlined,
  StarOutlined,
  HeartOutlined,
  ThunderboltOutlined,
  EyeOutlined,
  TeamOutlined
} from '@ant-design/icons';
import { Line, Column, Pie, Area, DualAxes } from '@ant-design/plots';
import dayjs from 'dayjs';
import { learningApi } from '../services/learningApi';

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;
const { Option } = Select;

interface AnalyticsData {
  totalStudyTime: number;
  coursesCompleted: number;
  currentStreak: number;
  averageScore: number;
  weeklyGoalProgress: number;
  monthlyGoalProgress: number;
  dailyGoalProgress: number;
  totalSessions: number;
  averageSessionTime: number;
  skillsImproved: number;
  achievementsEarned: number;
  studyEfficiency: number;
}

interface TimeSeriesData {
  date: string;
  studyTime: number;
  coursesCompleted: number;
  exercisesCompleted: number;
  score: number;
  efficiency: number;
}

interface SkillAnalytics {
  skillName: string;
  currentLevel: number;
  progress: number;
  timeSpent: number;
  improvement: number;
  trend: 'up' | 'down' | 'stable';
}

interface LearningPattern {
  timeOfDay: string;
  studyTime: number;
  efficiency: number;
  preferredSubjects: string[];
}

const LearningAnalyticsDashboard: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [analyticsData, setAnalyticsData] = useState<AnalyticsData | null>(null);
  const [timeSeriesData, setTimeSeriesData] = useState<TimeSeriesData[]>([]);
  const [skillAnalytics, setSkillAnalytics] = useState<SkillAnalytics[]>([]);
  const [learningPatterns, setLearningPatterns] = useState<LearningPattern[]>([]);
  const [selectedTimeRange, setSelectedTimeRange] = useState<string>('7days');
  const [selectedMetric, setSelectedMetric] = useState<string>('studyTime');
  const [dateRange, setDateRange] = useState<[moment.Moment, moment.Moment] | null>(null);

  // 模拟数据生成函数
  const generateMockAnalyticsData = (): AnalyticsData => ({
    totalStudyTime: 156.5,
    coursesCompleted: 12,
    currentStreak: 15,
    averageScore: 87.3,
    weeklyGoalProgress: 78,
    monthlyGoalProgress: 65,
    dailyGoalProgress: 85,
    totalSessions: 45,
    averageSessionTime: 3.5,
    skillsImproved: 8,
    achievementsEarned: 6,
    studyEfficiency: 92.1
  });

  const generateMockTimeSeriesData = (days: number): TimeSeriesData[] => {
    const data: TimeSeriesData[] = [];
    for (let i = days - 1; i >= 0; i--) {
      const date = dayjs().subtract(i, 'days');
      data.push({
        date: date.format('YYYY-MM-DD'),
        studyTime: Math.random() * 5 + 1,
        coursesCompleted: Math.floor(Math.random() * 3),
        exercisesCompleted: Math.floor(Math.random() * 8 + 2),
        score: Math.random() * 20 + 75,
        efficiency: Math.random() * 20 + 80
      });
    }
    return data;
  };

  const generateMockSkillAnalytics = (): SkillAnalytics[] => [
    { skillName: 'JavaScript', currentLevel: 8, progress: 85, timeSpent: 45.2, improvement: 12, trend: 'up' },
    { skillName: 'React', currentLevel: 7, progress: 78, timeSpent: 38.5, improvement: 8, trend: 'up' },
    { skillName: 'TypeScript', currentLevel: 6, progress: 65, timeSpent: 28.3, improvement: -2, trend: 'down' },
    { skillName: 'Node.js', currentLevel: 5, progress: 52, timeSpent: 22.1, improvement: 5, trend: 'stable' },
    { skillName: 'Python', currentLevel: 9, progress: 92, timeSpent: 52.8, improvement: 15, trend: 'up' },
    { skillName: 'SQL', currentLevel: 7, progress: 73, timeSpent: 31.4, improvement: 3, trend: 'stable' }
  ];

  const generateMockLearningPatterns = (): LearningPattern[] => [
    { timeOfDay: '09:00-12:00', studyTime: 2.5, efficiency: 95, preferredSubjects: ['编程', '算法'] },
    { timeOfDay: '14:00-17:00', studyTime: 1.8, efficiency: 88, preferredSubjects: ['设计', '理论'] },
    { timeOfDay: '19:00-22:00', studyTime: 2.2, efficiency: 82, preferredSubjects: ['复习', '练习'] },
    { timeOfDay: '22:00-24:00', studyTime: 1.2, efficiency: 75, preferredSubjects: ['阅读', '总结'] }
  ];

  const loadAnalyticsData = async () => {
    setLoading(true);
    try {
      // 这里应该调用真实的API，现在使用模拟数据
      const analytics = generateMockAnalyticsData();
      const timeSeries = generateMockTimeSeriesData(parseInt(selectedTimeRange.replace('days', '')));
      const skills = generateMockSkillAnalytics();
      const patterns = generateMockLearningPatterns();

      setAnalyticsData(analytics);
      setTimeSeriesData(timeSeries);
      setSkillAnalytics(skills);
      setLearningPatterns(patterns);
    } catch (error) {
      message.error('加载分析数据失败');
      console.error('Analytics data loading error:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAnalyticsData();
  }, [selectedTimeRange]);

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return <RiseOutlined style={{ color: '#52c41a' }} />;
      case 'down': return <FallOutlined style={{ color: '#ff4d4f' }} />;
      default: return <MinusOutlined style={{ color: '#faad14' }} />;
    }
  };

  const getEfficiencyColor = (efficiency: number) => {
    if (efficiency >= 90) return '#52c41a';
    if (efficiency >= 80) return '#1890ff';
    if (efficiency >= 70) return '#faad14';
    return '#ff4d4f';
  };

  const getStudyTimeChartData = () => {
    return timeSeriesData.map(item => ({
      date: dayjs(item.date).format('MM-DD'),
      value: item.studyTime,
      type: '学习时长'
    }));
  };

  const getPerformanceChartData = () => {
    return timeSeriesData.map(item => ({
      date: dayjs(item.date).format('MM-DD'),
      score: item.score,
      efficiency: item.efficiency
    }));
  };

  const getSkillDistributionData = () => {
    return skillAnalytics.map(skill => ({
      skill: skill.skillName,
      progress: skill.progress,
      level: skill.currentLevel
    }));
  };

  const getLearningPatternData = () => {
    return learningPatterns.map(pattern => ({
      time: pattern.timeOfDay,
      studyTime: pattern.studyTime,
      efficiency: pattern.efficiency
    }));
  };

  const exportData = () => {
    message.success('数据导出功能开发中...');
  };

  const shareReport = () => {
    message.success('分享报告功能开发中...');
  };

  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
        <div style={{ marginTop: '16px' }}>
          <Text>正在加载学习分析数据...</Text>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题和控制栏 */}
      <div style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space align="center">
              <LineChartOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
              <Title level={2} style={{ margin: 0 }}>
                学习分析仪表板
              </Title>
              <Badge count="实时" color="#52c41a" />
            </Space>
            <Text type="secondary" style={{ fontSize: '14px' }}>
              深入了解您的学习模式和进步轨迹
            </Text>
          </Col>
          <Col>
            <Space>
              <Select
                value={selectedTimeRange}
                onChange={setSelectedTimeRange}
                style={{ width: 120 }}
              >
                <Option value="7days">最近7天</Option>
                <Option value="14days">最近14天</Option>
                <Option value="30days">最近30天</Option>
                <Option value="90days">最近90天</Option>
              </Select>
              <RangePicker
                value={dateRange}
                onChange={setDateRange}
                style={{ width: 240 }}
              />
              <Button icon={<ReloadOutlined />} onClick={loadAnalyticsData}>
                刷新
              </Button>
              <Button icon={<DownloadOutlined />} onClick={exportData}>
                导出
              </Button>
              <Button icon={<ShareAltOutlined />} onClick={shareReport}>
                分享
              </Button>
            </Space>
          </Col>
        </Row>
      </div>

      {/* 核心指标概览 */}
      {analyticsData && (
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="总学习时长"
                value={analyticsData.totalStudyTime}
                suffix="小时"
                prefix={<ClockCircleOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
              <Progress 
                percent={analyticsData.weeklyGoalProgress} 
                size="small" 
                strokeColor="#1890ff"
                format={() => `周目标 ${analyticsData.weeklyGoalProgress}%`}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="学习效率"
                value={analyticsData.studyEfficiency}
                suffix="%"
                prefix={<ThunderboltOutlined />}
                valueStyle={{ color: getEfficiencyColor(analyticsData.studyEfficiency) }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  平均每次学习 {analyticsData.averageSessionTime} 小时
                </Text>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="连续学习"
                value={analyticsData.currentStreak}
                suffix="天"
                prefix={<FireOutlined />}
                valueStyle={{ color: '#faad14' }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  总共 {analyticsData.totalSessions} 次学习会话
                </Text>
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="平均分数"
                value={analyticsData.averageScore}
                suffix="分"
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  获得 {analyticsData.achievementsEarned} 个新成就
                </Text>
              </div>
            </Card>
          </Col>
        </Row>
      )}

      <Row gutter={[16, 16]}>
        {/* 学习时长趋势 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <LineChartOutlined />
                学习时长趋势
                <Tag color="blue">每日统计</Tag>
              </Space>
            }
            extra={
              <Select
                value={selectedMetric}
                onChange={setSelectedMetric}
                size="small"
                style={{ width: 100 }}
              >
                <Option value="studyTime">学习时长</Option>
                <Option value="courses">完成课程</Option>
                <Option value="exercises">练习数量</Option>
              </Select>
            }
          >
            {timeSeriesData.length > 0 ? (
              <Area
                data={getStudyTimeChartData()}
                xField="date"
                yField="value"
                height={280}
                smooth
                areaStyle={{
                  fill: 'l(270) 0:#ffffff 0.5:#7ec2f3 1:#1890ff',
                }}
                line={{
                  color: '#1890ff',
                  size: 2
                }}
                point={{
                  size: 4,
                  shape: 'circle',
                  style: {
                    fill: '#1890ff',
                    stroke: '#fff',
                    lineWidth: 2
                  }
                }}
                tooltip={{
                  formatter: (datum) => ({
                    name: '学习时长',
                    value: `${datum.value.toFixed(1)}小时`
                  })
                }}
              />
            ) : (
              <Empty description="暂无数据" />
            )}
          </Card>
        </Col>

        {/* 学习表现分析 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <BarChartOutlined />
                学习表现分析
                <Tag color="green">双轴图表</Tag>
              </Space>
            }
          >
            {timeSeriesData.length > 0 ? (
              <DualAxes
                data={[getPerformanceChartData(), getPerformanceChartData()]}
                xField="date"
                yField={['score', 'efficiency']}
                height={280}
                geometryOptions={[
                  {
                    geometry: 'column',
                    color: '#5B8FF9',
                    columnWidthRatio: 0.4,
                  },
                  {
                    geometry: 'line',
                    color: '#5AD8A6',
                    lineStyle: {
                      lineWidth: 2,
                    },
                    point: {
                      size: 4,
                      shape: 'circle',
                    },
                  },
                ]}
                tooltip={{
                  formatter: (datum, geometryType) => {
                    if (geometryType === 'column') {
                      return { name: '学习分数', value: `${datum.score.toFixed(1)}分` };
                    }
                    return { name: '学习效率', value: `${datum.efficiency.toFixed(1)}%` };
                  }
                }}
              />
            ) : (
              <Empty description="暂无数据" />
            )}
          </Card>
        </Col>

        {/* 技能进度分析 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <TrophyOutlined />
                技能进度分析
                <Badge count={skillAnalytics.length} color="#52c41a" />
              </Space>
            }
          >
            {skillAnalytics.length > 0 ? (
              <List
                dataSource={skillAnalytics}
                renderItem={(skill) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          style={{ 
                            backgroundColor: getEfficiencyColor(skill.progress),
                            fontSize: '12px'
                          }}
                        >
                          L{skill.currentLevel}
                        </Avatar>
                      }
                      title={
                        <Space>
                          <Text strong>{skill.skillName}</Text>
                          {getTrendIcon(skill.trend)}
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            {skill.improvement > 0 ? '+' : ''}{skill.improvement}%
                          </Text>
                        </Space>
                      }
                      description={
                        <div>
                          <Progress 
                            percent={skill.progress} 
                            size="small" 
                            strokeColor={getEfficiencyColor(skill.progress)}
                            format={() => `${skill.progress}%`}
                          />
                          <Text type="secondary" style={{ fontSize: '11px' }}>
                            学习时长: {skill.timeSpent}小时
                          </Text>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="暂无技能数据" />
            )}
          </Card>
        </Col>

        {/* 学习模式分析 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <PieChartOutlined />
                学习模式分析
                <Tag color="purple">时间分布</Tag>
              </Space>
            }
          >
            {learningPatterns.length > 0 ? (
              <Column
                data={getLearningPatternData()}
                xField="time"
                yField="studyTime"
                height={280}
                color="#722ed1"
                columnWidthRatio={0.6}
                tooltip={{
                  formatter: (datum) => [
                    { name: '学习时长', value: `${datum.studyTime}小时` },
                    { name: '学习效率', value: `${datum.efficiency}%` }
                  ]
                }}
                label={{
                  position: 'top',
                  formatter: (datum) => `${datum.studyTime}h`
                }}
              />
            ) : (
              <Empty description="暂无模式数据" />
            )}
          </Card>
        </Col>

        {/* 学习建议 */}
        <Col xs={24}>
          <Card 
            title={
              <Space>
                <BulbOutlined />
                智能学习建议
                <Badge count="AI推荐" color="#faad14" />
              </Space>
            }
          >
            <Row gutter={[16, 16]}>
              <Col xs={24} md={8}>
                <Alert
                  message="最佳学习时段"
                  description="根据您的学习数据，上午9-12点是您效率最高的时段，建议安排重要课程学习。"
                  type="success"
                  showIcon
                  icon={<ClockCircleOutlined />}
                />
              </Col>
              <Col xs={24} md={8}>
                <Alert
                  message="技能提升建议"
                  description="TypeScript技能进度有所下降，建议增加相关练习时间，巩固基础知识。"
                  type="warning"
                  showIcon
                  icon={<BookOutlined />}
                />
              </Col>
              <Col xs={24} md={8}>
                <Alert
                  message="学习目标调整"
                  description="当前学习强度适中，可以考虑适当增加每日学习目标，挑战更高水平。"
                  type="info"
                  showIcon
                  icon={<TrophyOutlined />}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default LearningAnalyticsDashboard;