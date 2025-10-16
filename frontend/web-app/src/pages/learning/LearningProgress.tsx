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
  Badge,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  message,
  Tabs,
  Alert,
  Divider,
  Statistic,
  Empty,
  Calendar,
  Drawer,
  Table,
  Rate
} from 'antd';
import { 
  BookOutlined, 
  TrophyOutlined, 
  ClockCircleOutlined, 
  FireOutlined,
  CalendarOutlined,
  LineChartOutlined,
  BarChartOutlined,
  PieChartOutlined,
  CheckCircleOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StarOutlined,
  AimOutlined,
  RocketOutlined,
  BulbOutlined,
  HeartOutlined,
  ThunderboltOutlined,
  CrownOutlined,
  GiftOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
  SettingOutlined,
  ShareAltOutlined,
  DownloadOutlined
} from '@ant-design/icons';
import { Line, Column, Pie, Area } from '@ant-design/plots';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Paragraph, Text } = Typography;

const { RangePicker } = DatePicker;

interface Course {
  id: string;
  title: string;
  category: string;
  instructor: string;
  totalLessons: number;
  completedLessons: number;
  totalDuration: number; // 分钟
  completedDuration: number; // 分钟
  progress: number;
  status: 'not_started' | 'in_progress' | 'completed' | 'paused';
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  rating: number;
  enrollDate: Date;
  lastStudyDate?: Date;
  estimatedCompletion?: Date;
  certificate?: boolean;
  tags: string[];
}

interface StudySession {
  id: string;
  courseId: string;
  courseName: string;
  lessonName: string;
  duration: number; // 分钟
  date: Date;
  progress: number;
  notes?: string;
  rating?: number;
}

interface LearningGoal {
  id: string;
  title: string;
  description: string;
  targetDate: Date;
  progress: number;
  status: 'active' | 'completed' | 'paused' | 'overdue';
  priority: 'high' | 'medium' | 'low';
  category: string;
  milestones: {
    id: string;
    title: string;
    completed: boolean;
    completedDate?: Date;
  }[];
}

interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: string;
  category: string;
  earnedDate: Date;
  rarity: 'common' | 'rare' | 'epic' | 'legendary';
  points: number;
}

const LearningProgress: React.FC = () => {
  const [courses, setCourses] = useState<Course[]>([]);
  const [studySessions, setStudySessions] = useState<StudySession[]>([]);
  const [learningGoals, setLearningGoals] = useState<LearningGoal[]>([]);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [goalModalVisible, setGoalModalVisible] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');
  const [dateRange, setDateRange] = useState<[moment.Moment, moment.Moment] | null>(null);
  const [loading, setLoading] = useState(true);
  const [form] = Form.useForm();

  useEffect(() => {
    loadLearningData();
  }, []);

  const loadLearningData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockCourses: Course[] = [
        {
          id: '1',
          title: '道德经深度解读',
          category: '道家思想',
          instructor: '张教授',
          totalLessons: 24,
          completedLessons: 18,
          totalDuration: 720, // 12小时
          completedDuration: 540, // 9小时
          progress: 75,
          status: 'in_progress',
          difficulty: 'intermediate',
          rating: 4.8,
          enrollDate: new Date('2024-01-01'),
          lastStudyDate: new Date('2024-02-01'),
          estimatedCompletion: new Date('2024-02-15'),
          certificate: true,
          tags: ['道家', '哲学', '经典']
        },
        {
          id: '2',
          title: '论语精讲',
          category: '儒家思想',
          instructor: '李教授',
          totalLessons: 30,
          completedLessons: 30,
          totalDuration: 900, // 15小时
          completedDuration: 900,
          progress: 100,
          status: 'completed',
          difficulty: 'beginner',
          rating: 4.9,
          enrollDate: new Date('2023-12-01'),
          lastStudyDate: new Date('2024-01-20'),
          certificate: true,
          tags: ['儒家', '经典', '修身']
        },
        {
          id: '3',
          title: '心经禅修指导',
          category: '佛家思想',
          instructor: '王法师',
          totalLessons: 12,
          completedLessons: 5,
          totalDuration: 360, // 6小时
          completedDuration: 150, // 2.5小时
          progress: 42,
          status: 'in_progress',
          difficulty: 'advanced',
          rating: 4.7,
          enrollDate: new Date('2024-01-15'),
          lastStudyDate: new Date('2024-01-30'),
          estimatedCompletion: new Date('2024-03-01'),
          certificate: false,
          tags: ['佛家', '禅修', '心经']
        },
        {
          id: '4',
          title: 'AI与传统文化融合',
          category: '现代应用',
          instructor: '陈博士',
          totalLessons: 16,
          completedLessons: 0,
          totalDuration: 480, // 8小时
          completedDuration: 0,
          progress: 0,
          status: 'not_started',
          difficulty: 'intermediate',
          rating: 4.6,
          enrollDate: new Date('2024-02-01'),
          certificate: true,
          tags: ['AI', '创新', '应用']
        }
      ];

      const mockStudySessions: StudySession[] = [
        {
          id: '1',
          courseId: '1',
          courseName: '道德经深度解读',
          lessonName: '第十八章：大道废，有仁义',
          duration: 45,
          date: new Date('2024-02-01'),
          progress: 100,
          notes: '深入理解了道德经中关于仁义的论述',
          rating: 5
        },
        {
          id: '2',
          courseId: '1',
          courseName: '道德经深度解读',
          lessonName: '第十七章：太上，不知有之',
          duration: 30,
          date: new Date('2024-01-31'),
          progress: 100,
          rating: 4
        },
        {
          id: '3',
          courseId: '3',
          courseName: '心经禅修指导',
          lessonName: '观自在菩萨的含义',
          duration: 35,
          date: new Date('2024-01-30'),
          progress: 100,
          notes: '学习了观自在菩萨的深层含义',
          rating: 5
        }
      ];

      const mockGoals: LearningGoal[] = [
        {
          id: '1',
          title: '完成道德经课程',
          description: '在2月底前完成道德经深度解读课程的学习',
          targetDate: new Date('2024-02-28'),
          progress: 75,
          status: 'active',
          priority: 'high',
          category: '课程学习',
          milestones: [
            { id: '1', title: '完成前10章', completed: true, completedDate: new Date('2024-01-15') },
            { id: '2', title: '完成前20章', completed: true, completedDate: new Date('2024-01-25') },
            { id: '3', title: '完成全部课程', completed: false },
            { id: '4', title: '获得课程证书', completed: false }
          ]
        },
        {
          id: '2',
          title: '每日学习1小时',
          description: '保持每天至少1小时的学习时间',
          targetDate: new Date('2024-12-31'),
          progress: 85,
          status: 'active',
          priority: 'medium',
          category: '学习习惯',
          milestones: [
            { id: '1', title: '连续7天', completed: true, completedDate: new Date('2024-01-07') },
            { id: '2', title: '连续30天', completed: true, completedDate: new Date('2024-01-30') },
            { id: '3', title: '连续90天', completed: false },
            { id: '4', title: '连续365天', completed: false }
          ]
        }
      ];

      const mockAchievements: Achievement[] = [
        {
          id: '1',
          title: '初学者',
          description: '完成第一门课程',
          icon: '🎓',
          category: '学习成就',
          earnedDate: new Date('2024-01-20'),
          rarity: 'common',
          points: 100
        },
        {
          id: '2',
          title: '坚持不懈',
          description: '连续学习30天',
          icon: '🔥',
          category: '学习习惯',
          earnedDate: new Date('2024-01-30'),
          rarity: 'rare',
          points: 500
        },
        {
          id: '3',
          title: '道德经专家',
          description: '完成道德经相关课程并获得高分',
          icon: '👑',
          category: '专业成就',
          earnedDate: new Date('2024-01-25'),
          rarity: 'epic',
          points: 1000
        }
      ];

      setCourses(mockCourses);
      setStudySessions(mockStudySessions);
      setLearningGoals(mockGoals);
      setAchievements(mockAchievements);
      setLoading(false);
    }, 1000);
  };

  // 获取学习统计数据
  const getStudyStats = () => {
    const totalCourses = courses.length;
    const completedCourses = courses.filter(c => c.status === 'completed').length;
    const inProgressCourses = courses.filter(c => c.status === 'in_progress').length;
    const totalStudyTime = studySessions.reduce((sum, session) => sum + session.duration, 0);
    const averageRating = courses.reduce((sum, course) => sum + course.rating, 0) / courses.length;
    const totalAchievements = achievements.length;
    const totalPoints = achievements.reduce((sum, achievement) => sum + achievement.points, 0);

    return {
      totalCourses,
      completedCourses,
      inProgressCourses,
      totalStudyTime,
      averageRating,
      totalAchievements,
      totalPoints
    };
  };

  // 获取学习趋势数据
  const getStudyTrendData = () => {
    const last30Days = Array.from({ length: 30 }, (_, i) => {
      const date = dayjs().subtract(29 - i, 'days');
    const sessionsOnDate = studySessions.filter(session => 
      dayjs(session.date).isSame(date, 'day')
      );
      const totalDuration = sessionsOnDate.reduce((sum, session) => sum + session.duration, 0);
      
      return {
        date: date.format('MM-DD'),
        duration: totalDuration,
        sessions: sessionsOnDate.length
      };
    });

    return last30Days;
  };

  // 获取课程分类数据
  const getCategoryData = () => {
    const categoryStats = courses.reduce((acc, course) => {
      if (!acc[course.category]) {
        acc[course.category] = { count: 0, completed: 0, totalProgress: 0 };
      }
      acc[course.category].count++;
      acc[course.category].totalProgress += course.progress;
      if (course.status === 'completed') {
        acc[course.category].completed++;
      }
      return acc;
    }, {} as Record<string, { count: number; completed: number; totalProgress: number }>);

    return Object.entries(categoryStats).map(([category, stats]) => ({
      category,
      count: stats.count,
      completed: stats.completed,
      averageProgress: Math.round(stats.totalProgress / stats.count)
    }));
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return '#52c41a';
      case 'in_progress': return '#1890ff';
      case 'paused': return '#faad14';
      case 'not_started': return '#d9d9d9';
      case 'overdue': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取难度颜色
  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case 'beginner': return '#52c41a';
      case 'intermediate': return '#faad14';
      case 'advanced': return '#ff4d4f';
      default: return '#d9d9d9';
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

  // 课程表格列定义
  const courseColumns: ColumnsType<Course> = [
    {
      title: '课程名称',
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: Course) => (
        <Space direction="vertical" size="small">
          <Text strong>{title}</Text>
          <Space>
            <Tag color={getDifficultyColor(record.difficulty)}>
              {record.difficulty === 'beginner' ? '初级' : 
               record.difficulty === 'intermediate' ? '中级' : '高级'}
            </Tag>
            <Text type="secondary">{record.instructor}</Text>
          </Space>
        </Space>
      )
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      width: 200,
      render: (progress: number, record: Course) => (
        <Space direction="vertical" size="small" style={{ width: '100%' }}>
          <Progress percent={progress} size="small" />
          <Text type="secondary">
            {record.completedLessons}/{record.totalLessons} 课时
          </Text>
        </Space>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap = {
          'not_started': '未开始',
          'in_progress': '学习中',
          'completed': '已完成',
          'paused': '已暂停'
        };
        return (
          <Tag color={getStatusColor(status)}>
            {statusMap[status as keyof typeof statusMap]}
          </Tag>
        );
      }
    },
    {
      title: '评分',
      dataIndex: 'rating',
      key: 'rating',
      width: 120,
      render: (rating: number) => (
        <Space>
          <Rate disabled value={rating} style={{ fontSize: '14px' }} />
          <Text>({rating})</Text>
        </Space>
      )
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record: Course) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              onClick={() => {
                setSelectedCourse(record);
                setDetailDrawerVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="继续学习">
            <Button 
              type="text" 
              icon={<PlayCircleOutlined />}
              disabled={record.status === 'completed'}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const stats = getStudyStats();
  const trendData = getStudyTrendData();
  const categoryData = getCategoryData();

  // 渲染概览页面
  const renderOverview = () => (
    <div>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总课程数"
              value={stats.totalCourses}
              prefix={<BookOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="已完成"
              value={stats.completedCourses}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="学习时长"
              value={Math.round(stats.totalStudyTime / 60 * 10) / 10}
              suffix="小时"
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="成就点数"
              value={stats.totalPoints}
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 学习趋势图 */}
        <Col xs={24} lg={16}>
          <Card title="学习趋势" extra={<LineChartOutlined />}>
            <Line
              data={trendData}
              xField="date"
              yField="duration"
              height={300}
              smooth
              point={{
                size: 3,
                shape: 'circle'
              }}
              tooltip={{
                formatter: (datum) => ({
                  name: '学习时长',
                  value: `${datum.duration} 分钟`
                })
              }}
            />
          </Card>
        </Col>

        {/* 课程分类统计 */}
        <Col xs={24} lg={8}>
          <Card title="课程分类" extra={<PieChartOutlined />}>
            <Pie
              data={categoryData}
              angleField="count"
              colorField="category"
              radius={0.8}
              height={300}
              label={{
                type: 'outer',
                content: '{name}: {percentage}'
              }}
            />
          </Card>
        </Col>
      </Row>

      {/* 最近学习记录 */}
      <Card title="最近学习记录" style={{ marginTop: '16px' }}>
        <Timeline>
          {studySessions.slice(0, 5).map(session => (
            <Timeline.Item
              key={session.id}
              color={session.progress === 100 ? 'green' : 'blue'}
            >
              <div>
                <Text strong>{session.courseName}</Text>
                <br />
                <Text>{session.lessonName}</Text>
                <br />
                <Space>
                  <Text type="secondary">
                    <ClockCircleOutlined /> {session.duration}分钟
                  </Text>
                  <Text type="secondary">
                    {dayjs(session.date).format('YYYY-MM-DD HH:mm')}
                  </Text>
                  {session.rating && (
                    <Rate disabled value={session.rating} style={{ fontSize: '12px' }} />
                  )}
                </Space>
                {session.notes && (
                  <div style={{ marginTop: '8px' }}>
                    <Text type="secondary">{session.notes}</Text>
                  </div>
                )}
              </div>
            </Timeline.Item>
          ))}
        </Timeline>
      </Card>
    </div>
  );

  // 渲染课程页面
  const renderCourses = () => (
    <div>
      <Card>
        <Table
          columns={courseColumns}
          dataSource={courses}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 门课程`
          }}
        />
      </Card>
    </div>
  );

  // 渲染学习目标页面
  const renderGoals = () => (
    <div>
      <Card 
        title="学习目标"
        extra={
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setGoalModalVisible(true)}
          >
            添加目标
          </Button>
        }
      >
        <Row gutter={[16, 16]}>
          {learningGoals.map(goal => (
            <Col xs={24} lg={12} key={goal.id}>
              <Card size="small">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Text strong>{goal.title}</Text>
                    <Tag color={getStatusColor(goal.status)}>
                      {goal.status === 'active' ? '进行中' :
                       goal.status === 'completed' ? '已完成' :
                       goal.status === 'paused' ? '已暂停' : '已逾期'}
                    </Tag>
                  </div>
                  
                  <Text type="secondary">{goal.description}</Text>
                  
                  <Progress percent={goal.progress} />
                  
                  <div>
                    <Text type="secondary">
                      目标日期: {dayjs(goal.targetDate).format('YYYY-MM-DD')}
                    </Text>
                  </div>
                  
                  <div>
                    <Text strong>里程碑进度:</Text>
                    <div style={{ marginTop: '8px' }}>
                      {goal.milestones.map(milestone => (
                        <div key={milestone.id} style={{ marginBottom: '4px' }}>
                          <Space>
                            <CheckCircleOutlined 
                              style={{ 
                                color: milestone.completed ? '#52c41a' : '#d9d9d9' 
                              }} 
                            />
                            <Text 
                              style={{ 
                                textDecoration: milestone.completed ? 'line-through' : 'none',
                                color: milestone.completed ? '#999' : 'inherit'
                              }}
                            >
                              {milestone.title}
                            </Text>
                            {milestone.completedDate && (
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                ({dayjs(milestone.completedDate).format('MM-DD')})
                              </Text>
                            )}
                          </Space>
                        </div>
                      ))}
                    </div>
                  </div>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  );

  // 渲染成就页面
  const renderAchievements = () => (
    <div>
      <Card title="成就徽章">
        <Row gutter={[16, 16]}>
          {achievements.map(achievement => (
            <Col xs={24} sm={12} md={8} lg={6} key={achievement.id}>
              <Card 
                size="small" 
                hoverable
                style={{ 
                  textAlign: 'center',
                  borderColor: getRarityColor(achievement.rarity)
                }}
              >
                <div style={{ fontSize: '48px', marginBottom: '8px' }}>
                  {achievement.icon}
                </div>
                <Text strong>{achievement.title}</Text>
                <br />
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  {achievement.description}
                </Text>
                <br />
                <Space style={{ marginTop: '8px' }}>
                  <Tag color={getRarityColor(achievement.rarity)}>
                    {achievement.rarity === 'common' ? '普通' :
                     achievement.rarity === 'rare' ? '稀有' :
                     achievement.rarity === 'epic' ? '史诗' : '传说'}
                  </Tag>
                  <Text strong style={{ color: '#faad14' }}>
                    +{achievement.points}
                  </Text>
                </Space>
                <br />
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  {dayjs(achievement.earnedDate).format('YYYY-MM-DD')}
                </Text>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  );

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <LineChartOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          学习进度
        </Title>
        <Paragraph>
          跟踪您的学习进度，管理学习目标，查看学习成就
        </Paragraph>
      </div>

      {/* 主要内容 */}
      <Tabs 
        activeKey={activeTab} 
        onChange={setActiveTab}
        items={[
          {
            key: 'overview',
            label: '学习概览',
            children: renderOverview()
          },
          {
            key: 'courses',
            label: '我的课程',
            children: renderCourses()
          },
          {
            key: 'goals',
            label: '学习目标',
            children: renderGoals()
          },
          {
            key: 'achievements',
            label: '成就徽章',
            children: renderAchievements()
          }
        ]}
      />

      {/* 课程详情抽屉 */}
      <Drawer
        title="课程详情"
        placement="right"
        width={600}
        visible={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
      >
        {selectedCourse && (
          <div>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <div>
                <Title level={4}>{selectedCourse.title}</Title>
                <Space wrap>
                  <Tag color={getDifficultyColor(selectedCourse.difficulty)}>
                    {selectedCourse.difficulty === 'beginner' ? '初级' : 
                     selectedCourse.difficulty === 'intermediate' ? '中级' : '高级'}
                  </Tag>
                  <Tag>{selectedCourse.category}</Tag>
                  <Text type="secondary">讲师: {selectedCourse.instructor}</Text>
                </Space>
              </div>

              <div>
                <Text strong>学习进度</Text>
                <Progress 
                  percent={selectedCourse.progress} 
                  style={{ marginTop: '8px' }}
                />
                <div style={{ marginTop: '8px' }}>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Statistic
                        title="已完成课时"
                        value={selectedCourse.completedLessons}
                        suffix={`/ ${selectedCourse.totalLessons}`}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="学习时长"
                        value={Math.round(selectedCourse.completedDuration / 60 * 10) / 10}
                        suffix={`/ ${Math.round(selectedCourse.totalDuration / 60 * 10) / 10} 小时`}
                      />
                    </Col>
                  </Row>
                </div>
              </div>

              <div>
                <Text strong>课程信息</Text>
                <div style={{ marginTop: '8px' }}>
                  <p><Text type="secondary">报名日期:</Text> {dayjs(selectedCourse.enrollDate).format('YYYY-MM-DD')}</p>
              {selectedCourse.lastStudyDate && (
                <p><Text type="secondary">最后学习:</Text> {dayjs(selectedCourse.lastStudyDate).format('YYYY-MM-DD')}</p>
              )}
              {selectedCourse.estimatedCompletion && (
                <p><Text type="secondary">预计完成:</Text> {dayjs(selectedCourse.estimatedCompletion).format('YYYY-MM-DD')}</p>
              )}
                  <p><Text type="secondary">课程评分:</Text> <Rate disabled value={selectedCourse.rating} style={{ fontSize: '14px' }} /></p>
                  {selectedCourse.certificate && (
                    <p><Text type="secondary">证书:</Text> <Tag color="gold">可获得证书</Tag></p>
                  )}
                </div>
              </div>

              <div>
                <Text strong>标签</Text>
                <div style={{ marginTop: '8px' }}>
                  <Space wrap>
                    {selectedCourse.tags.map((tag, index) => (
                      <Tag key={index}>{tag}</Tag>
                    ))}
                  </Space>
                </div>
              </div>

              <div>
                <Space>
                  <Button type="primary" icon={<PlayCircleOutlined />}>
                    继续学习
                  </Button>
                  <Button icon={<ShareAltOutlined />}>
                    分享课程
                  </Button>
                  {selectedCourse.certificate && selectedCourse.status === 'completed' && (
                    <Button icon={<DownloadOutlined />}>
                      下载证书
                    </Button>
                  )}
                </Space>
              </div>
            </Space>
          </div>
        )}
      </Drawer>

      {/* 添加目标模态框 */}
      <Modal
        title="添加学习目标"
        visible={goalModalVisible}
        onCancel={() => setGoalModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => {
            console.log('新目标:', values);
            message.success('目标添加成功');
            setGoalModalVisible(false);
            form.resetFields();
          }}
        >
          <Form.Item
            name="title"
            label="目标标题"
            rules={[{ required: true, message: '请输入目标标题' }]}
          >
            <Input placeholder="请输入学习目标" />
          </Form.Item>

          <Form.Item
            name="description"
            label="目标描述"
            rules={[{ required: true, message: '请输入目标描述' }]}
          >
            <Input.TextArea rows={3} placeholder="详细描述您的学习目标" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="targetDate"
                label="目标日期"
                rules={[{ required: true, message: '请选择目标日期' }]}
              >
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="priority"
                label="优先级"
                rules={[{ required: true, message: '请选择优先级' }]}
              >
                <Select placeholder="选择优先级">
                  <Select.Option value="high">高</Select.Option>
                  <Select.Option value="medium">中</Select.Option>
                  <Select.Option value="low">低</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="category"
            label="目标分类"
            rules={[{ required: true, message: '请选择目标分类' }]}
          >
            <Select placeholder="选择目标分类">
              <Select.Option value="课程学习">课程学习</Select.Option>
              <Select.Option value="技能提升">技能提升</Select.Option>
              <Select.Option value="学习习惯">学习习惯</Select.Option>
              <Select.Option value="考试认证">考试认证</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default LearningProgress;