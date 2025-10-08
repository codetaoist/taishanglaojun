import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Typography,
  Calendar,
  Badge,
  Modal,
  List,
  Avatar,
  Space,
  Statistic,
  Progress,
  Alert,
  Divider,
  Tag,
  Tooltip,
  Empty,
  message,
  Spin,
  Timeline,
  Tabs,
  Grid,
} from 'antd';
import {
  CalendarOutlined,
  GiftOutlined,
  TrophyOutlined,
  FireOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  StarOutlined,
  CrownOutlined,
  ThunderboltOutlined,
  HeartOutlined,
  RocketOutlined,
  BulbOutlined,
} from '@ant-design/icons';
import { Column, Pie } from '@ant-design/plots';
import dayjs, { Dayjs } from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;
const { useBreakpoint } = Grid;

// 签到奖励配置
const checkinRewards = [
  { day: 1, type: 'points', amount: 10, icon: '⭐', name: '积分奖励' },
  { day: 2, type: 'points', amount: 15, icon: '⭐', name: '积分奖励' },
  { day: 3, type: 'badge', amount: 1, icon: '🏅', name: '学习徽章' },
  { day: 4, type: 'points', amount: 20, icon: '⭐', name: '积分奖励' },
  { day: 5, type: 'points', amount: 25, icon: '⭐', name: '积分奖励' },
  { day: 6, type: 'gift', amount: 1, icon: '🎁', name: '神秘礼包' },
  { day: 7, type: 'premium', amount: 1, icon: '👑', name: '会员体验' },
];

// 生成签到历史数据
const generateCheckinHistory = () => {
  const history = [];
  const today = dayjs();
  
  for (let i = 30; i >= 0; i--) {
    const date = today.subtract(i, 'day');
    const isChecked = Math.random() > 0.3; // 70%概率签到
    
    if (isChecked) {
      history.push({
        date: date.format('YYYY-MM-DD'),
        timestamp: date.valueOf(),
        reward: checkinRewards[Math.floor(Math.random() * checkinRewards.length)],
        streak: Math.floor(Math.random() * 10) + 1,
      });
    }
  }
  
  return history.sort((a, b) => b.timestamp - a.timestamp);
};

// 生成签到统计数据
const generateCheckinStats = (history: any[]) => {
  const totalDays = history.length;
  const thisMonth = history.filter(h => dayjs(h.date).month() === dayjs().month()).length;
  const maxStreak = Math.max(...history.map(h => h.streak), 0);
  const currentStreak = history.length > 0 && dayjs(history[0].date).isSame(dayjs(), 'day') 
    ? history[0].streak 
    : 0;
  
  return {
    totalDays,
    thisMonth,
    maxStreak,
    currentStreak,
    totalPoints: totalDays * 15,
    totalBadges: Math.floor(totalDays / 3),
  };
};

const DailyCheckin: React.FC = () => {
  const [checkinHistory, setCheckinHistory] = useState<any[]>([]);
  const [todayChecked, setTodayChecked] = useState(false);
  const [loading, setLoading] = useState(false);
  const [rewardModalVisible, setRewardModalVisible] = useState(false);
  const [todayReward, setTodayReward] = useState<any>(null);
  const [stats, setStats] = useState<any>({});
  const screens = useBreakpoint();

  useEffect(() => {
    const history = generateCheckinHistory();
    setCheckinHistory(history);
    setStats(generateCheckinStats(history));
    
    // 检查今天是否已签到
    const today = dayjs().format('YYYY-MM-DD');
    setTodayChecked(history.some(h => h.date === today));
  }, []);

  const handleCheckin = async () => {
    setLoading(true);
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const today = dayjs().format('YYYY-MM-DD');
    const currentStreak = stats.currentStreak + 1;
    const rewardIndex = (currentStreak - 1) % 7;
    const reward = checkinRewards[rewardIndex];
    
    const newCheckin = {
      date: today,
      timestamp: dayjs().valueOf(),
      reward,
      streak: currentStreak,
    };
    
    const newHistory = [newCheckin, ...checkinHistory];
    setCheckinHistory(newHistory);
    setStats(generateCheckinStats(newHistory));
    setTodayChecked(true);
    setTodayReward(reward);
    setRewardModalVisible(true);
    setLoading(false);
    
    message.success('签到成功！');
  };

  const getCalendarDateCellRender = (value: Dayjs) => {
    const dateStr = value.format('YYYY-MM-DD');
    const checkin = checkinHistory.find(h => h.date === dateStr);
    
    if (checkin) {
      return (
        <Tooltip title={`已签到 - 获得${checkin.reward.name}`}>
          <Badge status="success" />
        </Tooltip>
      );
    }
    
    return null;
  };

  const renderCheckinCard = () => (
    <Card
      style={{ 
        textAlign: 'center', 
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        color: 'white',
        border: 'none',
      }}
    >
      <div style={{ padding: '20px 0' }}>
        <div style={{ fontSize: '64px', marginBottom: '16px' }}>
          {todayChecked ? '✅' : '📅'}
        </div>
        <Title level={2} style={{ color: 'white', margin: '0 0 8px 0' }}>
          {todayChecked ? '今日已签到' : '每日签到'}
        </Title>
        <Text style={{ color: 'rgba(255,255,255,0.8)', fontSize: '16px' }}>
          {todayChecked 
            ? `连续签到 ${stats.currentStreak} 天` 
            : '坚持签到，获得丰厚奖励'
          }
        </Text>
        
        {!todayChecked && (
          <div style={{ marginTop: '24px' }}>
            <Button
              type="primary"
              size="large"
              icon={<CheckCircleOutlined />}
              loading={loading}
              onClick={handleCheckin}
              style={{
                background: 'rgba(255,255,255,0.2)',
                border: '1px solid rgba(255,255,255,0.3)',
                backdropFilter: 'blur(10px)',
              }}
            >
              立即签到
            </Button>
          </div>
        )}
        
        {todayChecked && (
          <div style={{ marginTop: '24px' }}>
            <Text style={{ color: 'rgba(255,255,255,0.8)' }}>
              明天再来签到吧！
            </Text>
          </div>
        )}
      </div>
    </Card>
  );

  const renderStatsCards = () => (
    <Row gutter={[16, 16]}>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="累计签到"
            value={stats.totalDays}
            suffix="天"
            prefix={<CalendarOutlined />}
            valueStyle={{ color: '#1890ff' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="连续签到"
            value={stats.currentStreak}
            suffix="天"
            prefix={<FireOutlined />}
            valueStyle={{ color: '#f5222d' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="最长连签"
            value={stats.maxStreak}
            suffix="天"
            prefix={<TrophyOutlined />}
            valueStyle={{ color: '#faad14' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="本月签到"
            value={stats.thisMonth}
            suffix="天"
            prefix={<StarOutlined />}
            valueStyle={{ color: '#52c41a' }}
          />
        </Card>
      </Col>
    </Row>
  );

  const renderRewardProgress = () => {
    const currentDay = (stats.currentStreak % 7) || 7;
    
    return (
      <Card title="签到奖励进度" extra={<GiftOutlined />}>
        <div style={{ marginBottom: '16px' }}>
          <Text>连续签到进度：{currentDay}/7 天</Text>
          <Progress 
            percent={Math.round((currentDay / 7) * 100)} 
            strokeColor={{
              '0%': '#108ee9',
              '100%': '#87d068',
            }}
          />
        </div>
        
        <Row gutter={[8, 8]}>
          {checkinRewards.map((reward, index) => {
            const isCompleted = currentDay > index + 1;
            const isCurrent = currentDay === index + 1;
            
            return (
              <Col key={index} span={24 / 7} style={{ textAlign: 'center' }}>
                <div
                  style={{
                    padding: '8px',
                    borderRadius: '8px',
                    background: isCompleted 
                      ? '#f6ffed' 
                      : isCurrent 
                        ? '#fff7e6' 
                        : '#fafafa',
                    border: isCurrent ? '2px solid #faad14' : '1px solid #d9d9d9',
                  }}
                >
                  <div style={{ fontSize: '20px', marginBottom: '4px' }}>
                    {reward.icon}
                  </div>
                  <Text style={{ fontSize: '12px' }}>
                    第{index + 1}天
                  </Text>
                  {isCompleted && (
                    <div>
                      <CheckCircleOutlined style={{ color: '#52c41a' }} />
                    </div>
                  )}
                </div>
              </Col>
            );
          })}
        </Row>
      </Card>
    );
  };

  const renderCheckinHistory = () => (
    <Card title="签到历史" extra={<ClockCircleOutlined />}>
      {checkinHistory.length > 0 ? (
        <List
          dataSource={checkinHistory.slice(0, 10)}
          renderItem={(item) => (
            <List.Item>
              <List.Item.Meta
                avatar={
                  <Avatar 
                    style={{ backgroundColor: '#52c41a' }}
                    icon={<CheckCircleOutlined />}
                  />
                }
                title={
                  <Space>
                    <Text>{dayjs(item.date).format('MM月DD日')}</Text>
                    <Tag color="blue">连续{item.streak}天</Tag>
                  </Space>
                }
                description={
                  <Space>
                    <Text type="secondary">获得奖励：</Text>
                    <Text>{item.reward.icon} {item.reward.name}</Text>
                  </Space>
                }
              />
            </List.Item>
          )}
        />
      ) : (
        <Empty description="暂无签到记录" />
      )}
    </Card>
  );

  const renderCheckinCalendar = () => (
    <Card title="签到日历" extra={<CalendarOutlined />}>
      <Calendar
        fullscreen={false}
        dateCellRender={getCalendarDateCellRender}
      />
    </Card>
  );

  const renderRewardModal = () => (
    <Modal
      title="签到成功"
      open={rewardModalVisible}
      onCancel={() => setRewardModalVisible(false)}
      footer={[
        <Button key="ok" type="primary" onClick={() => setRewardModalVisible(false)}>
          确定
        </Button>
      ]}
      centered
    >
      {todayReward && (
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          <div style={{ fontSize: '64px', marginBottom: '16px' }}>
            {todayReward.icon}
          </div>
          <Title level={3}>恭喜获得奖励！</Title>
          <Paragraph>
            <Text strong>{todayReward.name}</Text>
            {todayReward.type === 'points' && (
              <Text> +{todayReward.amount} 积分</Text>
            )}
          </Paragraph>
          <Alert
            message={`连续签到 ${stats.currentStreak} 天`}
            type="success"
            showIcon
          />
        </div>
      )}
    </Modal>
  );

  const renderMonthlyChart = () => {
    const monthlyData = [];
    for (let i = 0; i < 12; i++) {
      const month = dayjs().month(i);
      const count = checkinHistory.filter(h => 
        dayjs(h.date).month() === i && dayjs(h.date).year() === dayjs().year()
      ).length;
      
      monthlyData.push({
        month: month.format('MM月'),
        count,
      });
    }

    const config = {
      data: monthlyData,
      xField: 'month',
      yField: 'count',
      color: '#1890ff',
      meta: {
        count: {
          alias: '签到天数',
        },
      },
    };

    return (
      <Card title="月度签到统计">
        <Column {...config} />
      </Card>
    );
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <CalendarOutlined style={{ marginRight: '8px' }} />
          每日签到
        </Title>
        <Text type="secondary">坚持每日签到，获得学习奖励和成就</Text>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} lg={8}>
          {renderCheckinCard()}
        </Col>
        <Col xs={24} lg={16}>
          {renderStatsCards()}
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} lg={12}>
          {renderRewardProgress()}
        </Col>
        <Col xs={24} lg={12}>
          {renderMonthlyChart()}
        </Col>
      </Row>

      <Tabs defaultActiveKey="history">
        <TabPane tab="签到历史" key="history">
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              {renderCheckinHistory()}
            </Col>
            <Col xs={24} lg={12}>
              {renderCheckinCalendar()}
            </Col>
          </Row>
        </TabPane>
        
        <TabPane tab="奖励说明" key="rewards">
          <Card>
            <Title level={4}>签到奖励规则</Title>
            <List
              dataSource={checkinRewards}
              renderItem={(reward, index) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar>{reward.icon}</Avatar>}
                    title={`第${index + 1}天：${reward.name}`}
                    description={
                      reward.type === 'points' 
                        ? `获得 ${reward.amount} 积分`
                        : reward.type === 'badge'
                          ? '获得专属学习徽章'
                          : reward.type === 'gift'
                            ? '获得神秘礼包'
                            : '获得会员体验'
                    }
                  />
                </List.Item>
              )}
            />
            
            <Divider />
            
            <Alert
              message="温馨提示"
              description="连续签到7天后，奖励循环重置。中断签到会重新开始计算连续天数。"
              type="info"
              showIcon
            />
          </Card>
        </TabPane>
      </Tabs>

      {renderRewardModal()}
    </div>
  );
};

export default DailyCheckin;