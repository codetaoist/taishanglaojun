import React, { useState, useEffect } from 'react';
import { Card, Avatar, Typography, Space, Button, Row, Col, Statistic } from 'antd';
import {
  UserOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  TrophyOutlined,
  FireOutlined,
  StarOutlined,
  ThunderboltOutlined
} from '@ant-design/icons';
import { useAuthContext as useAuth } from '../../contexts/AuthContext';

const { Title, Text } = Typography;

interface WelcomeSectionProps {
  className?: string;
}

const WelcomeSection: React.FC<WelcomeSectionProps> = ({ className }) => {
  const { user } = useAuth();
  const [currentTime, setCurrentTime] = useState(new Date());
  const [greeting, setGreeting] = useState('');

  // 更新时间
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  // 根据时间设置问候语
  useEffect(() => {
    const hour = currentTime.getHours();
    if (hour < 6) {
      setGreeting('夜深了，注意休息');
    } else if (hour < 12) {
      setGreeting('早上好');
    } else if (hour < 18) {
      setGreeting('下午好');
    } else {
      setGreeting('晚上好');
    }
  }, [currentTime]);

  // 模拟用户数据
  const userData = {
    level: 15,
    exp: 2850,
    maxExp: 3000,
    streak: 7, // 连续签到天数
    todayTasks: 3,
    completedTasks: 2,
    achievements: [
      { name: '智慧探索者', icon: <StarOutlined />, color: '#faad14' },
      { name: 'AI对话达人', icon: <ThunderboltOutlined />, color: '#1890ff' },
      { name: '社区贡献者', icon: <HeartOutlined />, color: '#f5222d' }
    ],
    recentActivity: {
      aiChats: 23,
      wisdomRead: 12,
      projectsCreated: 5
    }
  };

  const getTimeIcon = () => {
    const hour = currentTime.getHours();
    if (hour < 6 || hour >= 22) return '🌙';
    if (hour < 12) return '🌅';
    if (hour < 18) return '☀️';
    return '🌆';
  };

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'long'
    });
  };

  return (
    <Card className={className}>
      <Row gutter={[16, 16]} align="middle">
        {/* 左侧：用户信息和问候 */}
        <Col xs={24} lg={12}>
          <Space size="large" direction="horizontal" style={{ width: '100%' }}>
            <Avatar 
              size={{ xs: 48, sm: 56, md: 64 }}
              src={user?.avatar} 
              icon={<UserOutlined />}
            />
            <div style={{ flex: 1 }}>
              <Title level={3} style={{ margin: 0, fontSize: 'clamp(16px, 4vw, 20px)' }}>
                {greeting}，{user?.display_name || user?.first_name || user?.username}！
              </Title>
              <Space 
                className="mt-2" 
                size="small"
                style={{ 
                  display: 'flex',
                  flexWrap: 'wrap',
                  gap: '8px'
                }}
              >
                <Space size="small">
                  <CalendarOutlined />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {formatDate(currentTime)}
                  </Text>
                </Space>
                <Space size="small">
                  <ClockCircleOutlined />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {formatTime(currentTime)}
                  </Text>
                </Space>
              </Space>
            </div>
          </Space>
        </Col>

        {/* 右侧：统计数据 */}
        <Col xs={24} lg={12}>
          <Row gutter={[8, 8]}>
            <Col xs={8} sm={8} md={8}>
              <Statistic
                title="等级"
                value={userData.level}
                prefix={<TrophyOutlined />}
                valueStyle={{ color: '#faad14', fontSize: 'clamp(16px, 3vw, 24px)' }}
                style={{ textAlign: 'center' }}
              />
            </Col>
            <Col xs={8} sm={8} md={8}>
              <Statistic
                title="连续签到"
                value={userData.streak}
                suffix="天"
                prefix={<FireOutlined />}
                valueStyle={{ color: '#f5222d', fontSize: 'clamp(16px, 3vw, 24px)' }}
                style={{ textAlign: 'center' }}
              />
            </Col>
            <Col xs={8} sm={8} md={8}>
              <Statistic
                title="AI对话"
                value={userData.recentActivity.aiChats}
                prefix={<ThunderboltOutlined />}
                valueStyle={{ color: '#1890ff', fontSize: 'clamp(16px, 3vw, 24px)' }}
                style={{ textAlign: 'center' }}
              />
            </Col>
          </Row>
        </Col>
      </Row>
    </Card>
  );
};

export default WelcomeSection;