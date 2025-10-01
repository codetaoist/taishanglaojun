import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Progress, List, Avatar, Tag, Button, Divider, Timeline, Alert } from 'antd';
import {
  UserOutlined,
  HeartOutlined,
  BookOutlined,
  TrophyOutlined,
  RiseOutlined,
  ClockCircleOutlined,
  FireOutlined,
  StarOutlined,
  BulbOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { useAppSelector, useAppDispatch } from '../../hooks/redux';
import { initializeConsciousness, fetchInsights } from '../../store/slices/consciousnessSlice';
import { fetchWisdomItems } from '../../store/slices/culturalSlice';

const DashboardPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector(state => state.auth);
  const { language } = useAppSelector(state => state.ui);
  const {
    currentEntity,
    insights,
    sessions,
    loading: consciousnessLoading 
  } = useAppSelector(state => state.consciousness);
  const {
    learningProgress,
    wisdomItems,
    loading: culturalLoading
  } = useAppSelector(state => state.cultural);

  const [greeting, setGreeting] = useState('');

  // 获取问候语
  useEffect(() => {
    const hour = new Date().getHours();
    let greetingText = '';
    
    if (language === 'en-US') {
      if (hour < 12) greetingText = 'Good Morning';
      else if (hour < 18) greetingText = 'Good Afternoon';
      else greetingText = 'Good Evening';
    } else {
      if (hour < 12) greetingText = '早上好';
      else if (hour < 18) greetingText = '下午好';
      else greetingText = '晚上好';
    }
    
    setGreeting(greetingText);
  }, [language]);

  // 加载数据
  useEffect(() => {
    dispatch(initializeConsciousness());
    dispatch(fetchInsights({ limit: 5 }));
    dispatch(fetchWisdomItems({ category: 'philosophy' }));
  }, [dispatch]);

  // 获取文本内容
  const getText = (zhText: string, enText: string) => {
    return language === 'en-US' ? enText : zhText;
  };

  // 模拟数据（实际应该从后端获取）
  const mockStats = {
    totalSessions: sessions?.length || 12,
    wisdomPoints: (learningProgress as any)?.totalPoints || 2580,
    completedQuests: (learningProgress as any)?.completedQuests || 8,
    currentStreak: (learningProgress as any)?.currentStreak || 15,
  };

  const mockActivities = [
    {
      id: 1,
      type: 'consciousness',
      title: getText('完成意识融合会话', 'Completed consciousness fusion session'),
      time: '2小时前',
      icon: <HeartOutlined style={{ color: '#ff4d4f' }} />,
    },
    {
      id: 2,
      type: 'cultural',
      title: getText('学习《道德经》第一章', 'Studied Tao Te Ching Chapter 1'),
      time: '4小时前',
      icon: <BookOutlined style={{ color: '#1890ff' }} />,
    },
    {
      id: 3,
      type: 'achievement',
      title: getText('获得"智慧探索者"徽章', 'Earned "Wisdom Explorer" badge'),
      time: '1天前',
      icon: <TrophyOutlined style={{ color: '#faad14' }} />,
    },
    {
      id: 4,
      type: 'insight',
      title: getText('获得新的智慧洞察', 'Gained new wisdom insight'),
      time: '2天前',
      icon: <BulbOutlined style={{ color: '#52c41a' }} />,
    },
  ];

  const mockRecommendations = [
    {
      id: 1,
      title: getText('探索庄子的逍遥游', 'Explore Zhuangzi\'s Free and Easy Wandering'),
      description: getText('深入理解道家的自由精神', 'Understand the Taoist spirit of freedom'),
      category: 'cultural',
      difficulty: getText('中级', 'Intermediate'),
    },
    {
      id: 2,
      title: getText('意识融合进阶训练', 'Advanced Consciousness Fusion Training'),
      description: getText('提升意识融合的深度和质量', 'Enhance the depth and quality of consciousness fusion'),
      category: 'consciousness',
      difficulty: getText('高级', 'Advanced'),
    },
    {
      id: 3,
      title: getText('儒家修身之道', 'Confucian Self-Cultivation'),
      description: getText('学习儒家的修身养性理念', 'Learn Confucian concepts of self-cultivation'),
      category: 'cultural',
      difficulty: getText('初级', 'Beginner'),
    },
  ];

  return (
    <div className="dashboard-page">
      {/* 欢迎区域 */}
      <div className="welcome-section">
        <div className="welcome-content">
          <h1 className="welcome-title">
            {greeting}, {user?.username || getText('用户', 'User')}!
          </h1>
          <p className="welcome-subtitle">
            {getText(
              '欢迎回到太上老君序列零，继续您的智慧探索之旅',
              'Welcome back to Taishang Sequence Zero, continue your wisdom exploration journey'
            )}
          </p>
        </div>
        <div className="welcome-stats">
          <div className="stat-item">
            <div className="stat-number">{mockStats.currentStreak}</div>
            <div className="stat-label">{getText('连续天数', 'Day Streak')}</div>
          </div>
          <div className="stat-item">
            <div className="stat-number">{mockStats.wisdomPoints}</div>
            <div className="stat-label">{getText('智慧积分', 'Wisdom Points')}</div>
          </div>
        </div>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className="stats-row">
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title={getText('融合会话', 'Fusion Sessions')}
              value={mockStats.totalSessions}
              prefix={<HeartOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title={getText('智慧积分', 'Wisdom Points')}
              value={mockStats.wisdomPoints}
              prefix={<StarOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title={getText('完成任务', 'Completed Quests')}
              value={mockStats.completedQuests}
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title={getText('连续天数', 'Day Streak')}
              value={mockStats.currentStreak}
              prefix={<FireOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="content-row">
        {/* 学习进度 */}
        <Col xs={24} lg={12}>
          <Card 
            title={getText('学习进度', 'Learning Progress')}
            extra={<RiseOutlined />}
          >
            <div className="progress-section">
              <div className="progress-item">
                <div className="progress-header">
                  <span>{getText('意识融合', 'Consciousness Fusion')}</span>
                  <span className="progress-percent">75%</span>
                </div>
                <Progress percent={75} strokeColor="#ff4d4f" />
              </div>
              
              <div className="progress-item">
                <div className="progress-header">
                  <span>{getText('文化智慧', 'Cultural Wisdom')}</span>
                  <span className="progress-percent">60%</span>
                </div>
                <Progress percent={60} strokeColor="#1890ff" />
              </div>
              
              <div className="progress-item">
                <div className="progress-header">
                  <span>{getText('哲学思辨', 'Philosophical Thinking')}</span>
                  <span className="progress-percent">45%</span>
                </div>
                <Progress percent={45} strokeColor="#52c41a" />
              </div>
              
              <div className="progress-item">
                <div className="progress-header">
                  <span>{getText('修身养性', 'Self-Cultivation')}</span>
                  <span className="progress-percent">80%</span>
                </div>
                <Progress percent={80} strokeColor="#faad14" />
              </div>
            </div>
          </Card>
        </Col>

        {/* 最近活动 */}
        <Col xs={24} lg={12}>
          <Card 
            title={getText('最近活动', 'Recent Activities')}
            extra={<ClockCircleOutlined />}
          >
            <Timeline>
              {mockActivities.map(activity => (
                <Timeline.Item key={activity.id} dot={activity.icon}>
                  <div className="activity-item">
                    <div className="activity-title">{activity.title}</div>
                    <div className="activity-time">{activity.time}</div>
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="content-row">
        {/* 智慧洞察 */}
        <Col xs={24} lg={12}>
          <Card 
            title={getText('智慧洞察', 'Wisdom Insights')}
            extra={<BulbOutlined />}
          >
            {insights && insights.length > 0 ? (
              <List
                dataSource={insights.slice(0, 3)}
                renderItem={insight => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<BulbOutlined />} style={{ backgroundColor: '#faad14' }} />}
                      title={insight.title}
                      description={insight.content.substring(0, 100) + '...'}
                    />
                    <Tag color="blue">{insight.category}</Tag>
                  </List.Item>
                )}
              />
            ) : (
              <div className="empty-state">
                <BulbOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
                <p>{getText('暂无智慧洞察', 'No wisdom insights yet')}</p>
              </div>
            )}
          </Card>
        </Col>

        {/* 推荐内容 */}
        <Col xs={24} lg={12}>
          <Card 
            title={getText('推荐内容', 'Recommendations')}
            extra={<TeamOutlined />}
          >
            <List
              dataSource={mockRecommendations}
              renderItem={item => (
                <List.Item
                  actions={[
                    <Button type="link" size="small">
                      {getText('开始学习', 'Start Learning')}
                    </Button>
                  ]}
                >
                  <List.Item.Meta
                    title={item.title}
                    description={item.description}
                  />
                  <div className="recommendation-tags">
                    <Tag color={item.category === 'cultural' ? 'blue' : 'red'}>
                      {item.category === 'cultural' ? getText('文化', 'Cultural') : getText('意识', 'Consciousness')}
                    </Tag>
                    <Tag>{item.difficulty}</Tag>
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 每日提醒 */}
      <Row gutter={[16, 16]} className="content-row">
        <Col span={24}>
          <Alert
            message={getText('每日智慧', 'Daily Wisdom')}
            description={getText(
              '"知者不言，言者不知。" - 老子。真正的智慧往往存在于沉默之中，而非言语之间。',
              '"Those who know do not speak. Those who speak do not know." - Laozi. True wisdom often exists in silence, not in words.'
            )}
            type="info"
            showIcon
            icon={<BookOutlined />}
            className="daily-wisdom"
          />
        </Col>
      </Row>

      <style>{`
        .dashboard-page {
          padding: 0;
        }

        .welcome-section {
          background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-hover) 100%);
          border-radius: 12px;
          padding: 32px;
          margin-bottom: 24px;
          color: white;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .welcome-title {
          font-size: 28px;
          font-weight: 600;
          margin-bottom: 8px;
          color: white;
        }

        .welcome-subtitle {
          font-size: 16px;
          opacity: 0.9;
          margin: 0;
          line-height: 1.5;
        }

        .welcome-stats {
          display: flex;
          gap: 32px;
        }

        .stat-item {
          text-align: center;
        }

        .stat-number {
          font-size: 32px;
          font-weight: 700;
          line-height: 1;
          margin-bottom: 4px;
        }

        .stat-label {
          font-size: 14px;
          opacity: 0.8;
        }

        .stats-row {
          margin-bottom: 24px;
        }

        .content-row {
          margin-bottom: 24px;
        }

        .progress-section {
          display: flex;
          flex-direction: column;
          gap: 20px;
        }

        .progress-item {
          width: 100%;
        }

        .progress-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;
        }

        .progress-percent {
          font-weight: 600;
          color: var(--text-secondary);
        }

        .activity-item {
          display: flex;
          flex-direction: column;
        }

        .activity-title {
          font-weight: 500;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .activity-time {
          font-size: 12px;
          color: var(--text-tertiary);
        }

        .empty-state {
          text-align: center;
          padding: 40px 20px;
        }

        .empty-state p {
          margin-top: 16px;
          color: var(--text-tertiary);
        }

        .recommendation-tags {
          display: flex;
          gap: 4px;
          margin-top: 8px;
        }

        .daily-wisdom {
          border-radius: 8px;
        }

        /* 响应式设计 */
        @media (max-width: 768px) {
          .welcome-section {
            flex-direction: column;
            text-align: center;
            gap: 24px;
          }

          .welcome-stats {
            gap: 24px;
          }

          .welcome-title {
            font-size: 24px;
          }

          .welcome-subtitle {
            font-size: 14px;
          }

          .stat-number {
            font-size: 24px;
          }
        }

        @media (max-width: 480px) {
          .welcome-section {
            padding: 24px 20px;
          }

          .welcome-stats {
            gap: 16px;
          }

          .stat-number {
            font-size: 20px;
          }

          .stat-label {
            font-size: 12px;
          }
        }
      `}</style>
    </div>
  );
};

export default DashboardPage;