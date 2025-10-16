import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Typography,
  Progress,
  Modal,
  List,
  Avatar,
  Space,
  Statistic,
  Badge,
  Tag,
  Tooltip,
  Empty,
  message,
  Tabs,
  Timeline,
  Divider,
  Alert,
  Image,
  Popover,
  Grid,
  Input,
  Select,
  Drawer,
} from 'antd';
import {
  TrophyOutlined,
  StarOutlined,
  CrownOutlined,
  FireOutlined,
  ThunderboltOutlined,
  HeartOutlined,
  RocketOutlined,
  BulbOutlined,
  ShareAltOutlined,
  EyeOutlined,
  LockOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  FilterOutlined,
  SearchOutlined,
  GiftOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import { Pie, Column } from '@ant-design/plots';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;

const { useBreakpoint } = Grid;
const { Option } = Select;

// 成就类型配置
const achievementTypes = [
  { key: 'learning', name: '学习成就', icon: '📚', color: '#1890ff' },
  { key: 'streak', name: '连续学习', icon: '🔥', color: '#f5222d' },
  { key: 'skill', name: '技能掌握', icon: '🎯', color: '#52c41a' },
  { key: 'social', name: '社交互动', icon: '👥', color: '#722ed1' },
  { key: 'milestone', name: '里程碑', icon: '🏆', color: '#faad14' },
  { key: 'special', name: '特殊成就', icon: '⭐', color: '#13c2c2' },
];

// 成就稀有度配置
const rarityConfig = {
  common: { name: '普通', color: '#8c8c8c', icon: '🥉' },
  rare: { name: '稀有', color: '#1890ff', icon: '🥈' },
  epic: { name: '史诗', color: '#722ed1', icon: '🥇' },
  legendary: { name: '传说', color: '#faad14', icon: '👑' },
  mythic: { name: '神话', color: '#f5222d', icon: '💎' },
};

// 生成模拟成就数据
const generateMockAchievements = () => {
  return [
    {
      id: 1,
      title: '学习新手',
      description: '完成第一次学习任务',
      type: 'learning',
      rarity: 'common',
      icon: '🎓',
      progress: 100,
      maxProgress: 1,
      unlocked: true,
      unlockedAt: '2024-01-15',
      points: 10,
      requirements: ['完成任意一个学习任务'],
      tips: '这是您学习之旅的第一步！',
    },
    {
      id: 2,
      title: '连续学习者',
      description: '连续学习7天',
      type: 'streak',
      rarity: 'rare',
      icon: '🔥',
      progress: 5,
      maxProgress: 7,
      unlocked: false,
      points: 50,
      requirements: ['连续7天进行学习活动'],
      tips: '坚持就是胜利！每天学习一点点。',
    },
    {
      id: 3,
      title: '技能大师',
      description: '掌握5项核心技能',
      type: 'skill',
      rarity: 'epic',
      icon: '🎯',
      progress: 3,
      maxProgress: 5,
      unlocked: false,
      points: 100,
      requirements: ['完成5个不同技能的学习路径'],
      tips: '多元化学习让您更加全面！',
    },
    {
      id: 4,
      title: '社交达人',
      description: '与50位学习伙伴互动',
      type: 'social',
      rarity: 'rare',
      icon: '👥',
      progress: 23,
      maxProgress: 50,
      unlocked: false,
      points: 75,
      requirements: ['与50位不同的用户进行学习交流'],
      tips: '学习路上不孤单，一起进步更快乐！',
    },
    {
      id: 5,
      title: '百日学者',
      description: '累计学习100天',
      type: 'milestone',
      rarity: 'legendary',
      icon: '📖',
      progress: 67,
      maxProgress: 100,
      unlocked: false,
      points: 200,
      requirements: ['累计学习天数达到100天'],
      tips: '百日筑基，学习成就非凡！',
    },
    {
      id: 6,
      title: '完美主义者',
      description: '连续30次测试满分',
      type: 'special',
      rarity: 'mythic',
      icon: '💯',
      progress: 12,
      maxProgress: 30,
      unlocked: false,
      points: 500,
      requirements: ['连续30次测试获得满分'],
      tips: '追求完美的您值得最高荣誉！',
    },
    {
      id: 7,
      title: '早起鸟儿',
      description: '连续30天早上6点前学习',
      type: 'special',
      rarity: 'epic',
      icon: '🌅',
      progress: 8,
      maxProgress: 30,
      unlocked: false,
      points: 150,
      requirements: ['连续30天在早上6点前开始学习'],
      tips: '早起的鸟儿有虫吃，早学的您有成就！',
    },
    {
      id: 8,
      title: '知识探索者',
      description: '学习10个不同领域的课程',
      type: 'learning',
      rarity: 'epic',
      icon: '🔍',
      progress: 6,
      maxProgress: 10,
      unlocked: false,
      points: 120,
      requirements: ['完成10个不同学科领域的课程'],
      tips: '广泛的知识面让您更有竞争力！',
    },
  ];
};

// 生成成就统计数据
const generateAchievementStats = (achievements: any[]) => {
  const unlocked = achievements.filter(a => a.unlocked).length;
  const total = achievements.length;
  const totalPoints = achievements.filter(a => a.unlocked).reduce((sum, a) => sum + a.points, 0);
  const rarityStats = Object.keys(rarityConfig).map(rarity => ({
    rarity,
    count: achievements.filter(a => a.rarity === rarity && a.unlocked).length,
    total: achievements.filter(a => a.rarity === rarity).length,
  }));

  return {
    unlocked,
    total,
    totalPoints,
    rarityStats,
    completionRate: Math.round((unlocked / total) * 100),
  };
};

const AchievementCenter: React.FC = () => {
  const [achievements, setAchievements] = useState<any[]>([]);
  const [filteredAchievements, setFilteredAchievements] = useState<any[]>([]);
  const [selectedAchievement, setSelectedAchievement] = useState<any>(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [shareModalVisible, setShareModalVisible] = useState(false);
  const [stats, setStats] = useState<any>({});
  const [filterType, setFilterType] = useState('all');
  const [filterStatus, setFilterStatus] = useState('all');
  const [searchText, setSearchText] = useState('');
  const screens = useBreakpoint();

  useEffect(() => {
    const mockAchievements = generateMockAchievements();
    setAchievements(mockAchievements);
    setFilteredAchievements(mockAchievements);
    setStats(generateAchievementStats(mockAchievements));
  }, []);

  useEffect(() => {
    let filtered = achievements;

    // 按类型筛选
    if (filterType !== 'all') {
      filtered = filtered.filter(a => a.type === filterType);
    }

    // 按状态筛选
    if (filterStatus === 'unlocked') {
      filtered = filtered.filter(a => a.unlocked);
    } else if (filterStatus === 'locked') {
      filtered = filtered.filter(a => !a.unlocked);
    }

    // 按搜索文本筛选
    if (searchText) {
      filtered = filtered.filter(a => 
        a.title.toLowerCase().includes(searchText.toLowerCase()) ||
        a.description.toLowerCase().includes(searchText.toLowerCase())
      );
    }

    setFilteredAchievements(filtered);
  }, [achievements, filterType, filterStatus, searchText]);

  const handleAchievementClick = (achievement: any) => {
    setSelectedAchievement(achievement);
    setDetailModalVisible(true);
  };

  const handleShare = (achievement: any) => {
    setSelectedAchievement(achievement);
    setShareModalVisible(true);
  };

  const copyShareLink = () => {
    const shareText = `我在太上老君学习平台获得了"${selectedAchievement.title}"成就！🎉`;
    navigator.clipboard.writeText(shareText);
    message.success('分享内容已复制到剪贴板');
  };

  const renderAchievementCard = (achievement: any) => {
    const rarity = rarityConfig[achievement.rarity as keyof typeof rarityConfig];
    const progressPercent = Math.round((achievement.progress / achievement.maxProgress) * 100);

    return (
      <Card
        key={achievement.id}
        hoverable
        className={`achievement-card ${achievement.unlocked ? 'unlocked' : 'locked'}`}
        style={{
          position: 'relative',
          opacity: achievement.unlocked ? 1 : 0.7,
          border: achievement.unlocked ? `2px solid ${rarity.color}` : '1px solid #d9d9d9',
          background: achievement.unlocked 
            ? `linear-gradient(135deg, ${rarity.color}15 0%, ${rarity.color}05 100%)`
            : '#fafafa',
        }}
        styles={{ body: { padding: '16px' } }}
        onClick={() => handleAchievementClick(achievement)}
      >
        {achievement.unlocked && (
          <div
            style={{
              position: 'absolute',
              top: '8px',
              right: '8px',
              background: rarity.color,
              color: 'white',
              padding: '2px 6px',
              borderRadius: '4px',
              fontSize: '10px',
              fontWeight: 'bold',
            }}
          >
            {rarity.name}
          </div>
        )}

        <div style={{ textAlign: 'center', marginBottom: '12px' }}>
          <div
            style={{
              fontSize: '48px',
              marginBottom: '8px',
              filter: achievement.unlocked ? 'none' : 'grayscale(100%)',
            }}
          >
            {achievement.icon}
          </div>
          <Title level={5} style={{ margin: 0, color: achievement.unlocked ? rarity.color : '#8c8c8c' }}>
            {achievement.title}
          </Title>
        </div>

        <Paragraph
          ellipsis={{ rows: 2 }}
          style={{ 
            textAlign: 'center', 
            marginBottom: '12px',
            color: achievement.unlocked ? 'inherit' : '#8c8c8c',
          }}
        >
          {achievement.description}
        </Paragraph>

        {!achievement.unlocked && (
          <div style={{ marginBottom: '12px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
              <Text type="secondary">进度</Text>
              <Text type="secondary">{achievement.progress}/{achievement.maxProgress}</Text>
            </div>
            <Progress 
              percent={progressPercent} 
              size="small"
              strokeColor={rarity.color}
            />
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Tag color={achievementTypes.find(t => t.key === achievement.type)?.color}>
              {achievementTypes.find(t => t.key === achievement.type)?.name}
            </Tag>
            <Text strong style={{ color: rarity.color }}>
              +{achievement.points}分
            </Text>
          </Space>
          
          {achievement.unlocked && (
            <Space>
              <Tooltip title="查看详情">
                <Button size="small" icon={<EyeOutlined />} />
              </Tooltip>
              <Tooltip title="分享成就">
                <Button 
                  size="small" 
                  icon={<ShareAltOutlined />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleShare(achievement);
                  }}
                />
              </Tooltip>
            </Space>
          )}
        </div>

        {achievement.unlocked && achievement.unlockedAt && (
          <div style={{ marginTop: '8px', textAlign: 'center' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              获得于 {dayjs(achievement.unlockedAt).format('YYYY年MM月DD日')}
            </Text>
          </div>
        )}
      </Card>
    );
  };

  const renderStatsCards = () => (
    <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="已获得成就"
            value={stats.unlocked}
            suffix={`/ ${stats.total}`}
            prefix={<TrophyOutlined />}
            valueStyle={{ color: '#1890ff' }}
          />
          <Progress 
            percent={stats.completionRate} 
            size="small" 
            strokeColor="#1890ff"
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="成就积分"
            value={stats.totalPoints}
            prefix={<StarOutlined />}
            valueStyle={{ color: '#faad14' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="稀有成就"
            value={stats.rarityStats?.filter((r: any) => ['epic', 'legendary', 'mythic'].includes(r.rarity) && r.count > 0).length || 0}
            prefix={<CrownOutlined />}
            valueStyle={{ color: '#722ed1' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="完成率"
            value={stats.completionRate}
            suffix="%"
            prefix={<CheckCircleOutlined />}
            valueStyle={{ color: '#52c41a' }}
          />
        </Card>
      </Col>
    </Row>
  );

  const renderFilters = () => (
    <Card style={{ marginBottom: '16px' }}>
      <Row gutter={[16, 16]} align="middle">
        <Col xs={24} sm={8} md={6}>
          <Input
            placeholder="搜索成就..."
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
          />
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Select
            placeholder="选择类型"
            style={{ width: '100%' }}
            value={filterType}
            onChange={setFilterType}
          >
            <Option value="all">全部类型</Option>
            {achievementTypes.map(type => (
              <Option key={type.key} value={type.key}>
                {type.icon} {type.name}
              </Option>
            ))}
          </Select>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Select
            placeholder="选择状态"
            style={{ width: '100%' }}
            value={filterStatus}
            onChange={setFilterStatus}
          >
            <Option value="all">全部状态</Option>
            <Option value="unlocked">已获得</Option>
            <Option value="locked">未获得</Option>
          </Select>
        </Col>
      </Row>
    </Card>
  );

  const renderDetailModal = () => (
    <Modal
      title={null}
      open={detailModalVisible}
      onCancel={() => setDetailModalVisible(false)}
      footer={null}
      width={600}
      centered
    >
      {selectedAchievement && (
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          <div style={{ fontSize: '80px', marginBottom: '16px' }}>
            {selectedAchievement.icon}
          </div>
          
          <Title level={2} style={{ 
            color: rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].color,
            marginBottom: '8px',
          }}>
            {selectedAchievement.title}
          </Title>
          
          <Tag 
            color={rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].color}
            style={{ marginBottom: '16px' }}
          >
            {rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].icon} {rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].name}
          </Tag>
          
          <Paragraph style={{ fontSize: '16px', marginBottom: '24px' }}>
            {selectedAchievement.description}
          </Paragraph>

          {!selectedAchievement.unlocked && (
            <div style={{ marginBottom: '24px' }}>
              <Title level={4}>完成进度</Title>
              <Progress 
                percent={Math.round((selectedAchievement.progress / selectedAchievement.maxProgress) * 100)}
                strokeColor={rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].color}
                style={{ marginBottom: '8px' }}
              />
              <Text>{selectedAchievement.progress} / {selectedAchievement.maxProgress}</Text>
            </div>
          )}

          <div style={{ marginBottom: '24px' }}>
            <Title level={4}>获得条件</Title>
            <List
              size="small"
              dataSource={selectedAchievement.requirements}
              renderItem={(item: string) => (
                <List.Item>
                  <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
                  {item}
                </List.Item>
              )}
            />
          </div>

          {selectedAchievement.tips && (
            <Alert
              message="小贴士"
              description={selectedAchievement.tips}
              type="info"
              showIcon
              style={{ marginBottom: '24px' }}
            />
          )}

          <div style={{ display: 'flex', justifyContent: 'center', gap: '16px' }}>
            <Statistic
              title="奖励积分"
              value={selectedAchievement.points}
              prefix={<StarOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
            {selectedAchievement.unlocked && selectedAchievement.unlockedAt && (
              <Statistic
                title="获得时间"
                value={dayjs(selectedAchievement.unlockedAt).format('YYYY-MM-DD')}
                prefix={<CalendarOutlined />}
              />
            )}
          </div>

          {selectedAchievement.unlocked && (
            <div style={{ marginTop: '24px' }}>
              <Button 
                type="primary" 
                icon={<ShareAltOutlined />}
                onClick={() => {
                  setDetailModalVisible(false);
                  handleShare(selectedAchievement);
                }}
              >
                分享成就
              </Button>
            </div>
          )}
        </div>
      )}
    </Modal>
  );

  const renderShareModal = () => (
    <Modal
      title="分享成就"
      open={shareModalVisible}
      onCancel={() => setShareModalVisible(false)}
      footer={[
        <Button key="copy" type="primary" onClick={copyShareLink}>
          复制分享内容
        </Button>,
        <Button key="cancel" onClick={() => setShareModalVisible(false)}>
          取消
        </Button>
      ]}
      centered
    >
      {selectedAchievement && (
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          <div style={{ fontSize: '64px', marginBottom: '16px' }}>
            {selectedAchievement.icon}
          </div>
          <Title level={3}>{selectedAchievement.title}</Title>
          <Paragraph>
            我在太上老君学习平台获得了"{selectedAchievement.title}"成就！🎉
          </Paragraph>
          <Tag color={rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].color}>
            {rarityConfig[selectedAchievement.rarity as keyof typeof rarityConfig].name}成就
          </Tag>
        </div>
      )}
    </Modal>
  );

  const renderRarityChart = () => {
    const chartData = stats.rarityStats?.map((item: any) => ({
      rarity: rarityConfig[item.rarity as keyof typeof rarityConfig].name,
      count: item.count,
      total: item.total,
    })) || [];

    const config = {
      data: chartData,
      angleField: 'count',
      colorField: 'rarity',
      radius: 0.8,
      label: {
        type: 'outer',
        content: '{name} {percentage}',
      },
      interactions: [{ type: 'element-active' }],
    };

    return (
      <Card title="成就稀有度分布" style={{ marginBottom: '16px' }}>
        <Pie {...config} />
      </Card>
    );
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <TrophyOutlined style={{ marginRight: '8px' }} />
          成就中心
        </Title>
        <Text type="secondary">展示您的学习成就，分享您的进步历程</Text>
      </div>

      {renderStatsCards()}

      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} lg={16}>
          {renderFilters()}
        </Col>
        <Col xs={24} lg={8}>
          {renderRarityChart()}
        </Col>
      </Row>

      <Tabs 
        defaultActiveKey="all"
        items={[
          {
            key: 'all',
            label: `全部成就 (${filteredAchievements.length})`,
            children: filteredAchievements.length > 0 ? (
              <Row gutter={[16, 16]}>
                {filteredAchievements.map(achievement => (
                  <Col key={achievement.id} xs={24} sm={12} lg={8} xl={6}>
                    {renderAchievementCard(achievement)}
                  </Col>
                ))}
              </Row>
            ) : (
              <Empty description="没有找到匹配的成就" />
            )
          },
          {
            key: 'unlocked',
            label: `已获得 (${achievements.filter(a => a.unlocked).length})`,
            children: (
              <Row gutter={[16, 16]}>
                {achievements.filter(a => a.unlocked).map(achievement => (
                  <Col key={achievement.id} xs={24} sm={12} lg={8} xl={6}>
                    {renderAchievementCard(achievement)}
                  </Col>
                ))}
              </Row>
            )
          },
          {
            key: 'progress',
            label: `进行中 (${achievements.filter(a => !a.unlocked && a.progress > 0).length})`,
            children: (
              <Row gutter={[16, 16]}>
                {achievements.filter(a => !a.unlocked && a.progress > 0).map(achievement => (
                  <Col key={achievement.id} xs={24} sm={12} lg={8} xl={6}>
                    {renderAchievementCard(achievement)}
                  </Col>
                ))}
              </Row>
            )
          }
        ]}
      />

      {renderDetailModal()}
      {renderShareModal()}
    </div>
  );
};

export default AchievementCenter;