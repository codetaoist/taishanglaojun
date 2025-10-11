import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Typography, 
  Space, 
  Tag, 
  Rate, 
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
  message,
  Tabs,
  Alert,
  Divider,
  Collapse
} from 'antd';
import { 
  BulbOutlined, 
  HeartOutlined, 
  ThunderboltOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  StarOutlined,
  LikeOutlined,
  DislikeOutlined,
  ShareAltOutlined,
  BookOutlined,
  PlayCircleOutlined,
  FileTextOutlined,
  CalendarOutlined,
  SettingOutlined,
  PlusOutlined,
  EyeOutlined,
  TrophyOutlined,
  FireOutlined,
  MedicineBoxOutlined
} from '@ant-design/icons';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

const { Panel } = Collapse;

interface HealthAdvice {
  id: string;
  title: string;
  category: string;
  priority: 'high' | 'medium' | 'low';
  type: 'exercise' | 'diet' | 'lifestyle' | 'medical' | 'mental';
  description: string;
  benefits: string[];
  steps: string[];
  duration: string;
  difficulty: number;
  effectiveness: number;
  personalizedScore: number;
  tags: string[];
  isCompleted: boolean;
  isFavorited: boolean;
  rating: number;
  feedback?: string;
  createdAt: Date;
  dueDate?: Date;
}

interface AdviceCategory {
  key: string;
  name: string;
  icon: React.ReactNode;
  color: string;
  count: number;
}

const HealthAdvice: React.FC = () => {
  const [healthAdvices, setHealthAdvices] = useState<HealthAdvice[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedPriority, setSelectedPriority] = useState<string>('all');
  const [feedbackVisible, setFeedbackVisible] = useState(false);
  const [selectedAdvice, setSelectedAdvice] = useState<HealthAdvice | null>(null);
  const [loading, setLoading] = useState(true);
  const [form] = Form.useForm();

  // 建议分类
  const categories: AdviceCategory[] = [
    { key: 'all', name: '全部', icon: <BulbOutlined />, color: '#1890ff', count: 0 },
    { key: 'exercise', name: '运动健身', icon: <ThunderboltOutlined />, color: '#52c41a', count: 0 },
    { key: 'diet', name: '饮食营养', icon: <HeartOutlined />, color: '#faad14', count: 0 },
    { key: 'lifestyle', name: '生活方式', icon: <ClockCircleOutlined />, color: '#722ed1', count: 0 },
    { key: 'medical', name: '医疗保健', icon: <MedicineBoxOutlined />, color: '#ff4d4f', count: 0 },
    { key: 'mental', name: '心理健康', icon: <StarOutlined />, color: '#13c2c2', count: 0 }
  ];

  useEffect(() => {
    loadHealthAdvices();
  }, []);

  const loadHealthAdvices = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockAdvices: HealthAdvice[] = [
        {
          id: '1',
          title: '每日30分钟快走计划',
          category: 'exercise',
          priority: 'high',
          type: 'exercise',
          description: '基于您的心率数据分析，建议进行低强度有氧运动来改善心血管健康',
          benefits: ['改善心血管功能', '增强体质', '控制体重', '提升睡眠质量'],
          steps: [
            '选择合适的运动鞋和舒适的服装',
            '选择安全的步行路线',
            '以中等速度步行30分钟',
            '运动后进行5分钟拉伸',
            '记录运动数据和感受'
          ],
          duration: '30分钟/天',
          difficulty: 2,
          effectiveness: 85,
          personalizedScore: 92,
          tags: ['有氧运动', '心血管', '减重'],
          isCompleted: false,
          isFavorited: true,
          rating: 4.5,
          createdAt: new Date(),
          dueDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
        },
        {
          id: '2',
          title: '地中海饮食模式',
          category: 'diet',
          priority: 'high',
          type: 'diet',
          description: '根据您的血脂水平，推荐采用地中海饮食模式来改善血脂状况',
          benefits: ['降低胆固醇', '预防心血管疾病', '抗氧化', '控制血糖'],
          steps: [
            '增加橄榄油的使用',
            '每周至少吃2次鱼类',
            '多吃新鲜蔬菜和水果',
            '选择全谷物食品',
            '适量摄入坚果和豆类'
          ],
          duration: '长期坚持',
          difficulty: 3,
          effectiveness: 78,
          personalizedScore: 88,
          tags: ['营养均衡', '血脂', '抗氧化'],
          isCompleted: false,
          isFavorited: false,
          rating: 4.2,
          createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000)
        },
        {
          id: '3',
          title: '睡眠质量优化方案',
          category: 'lifestyle',
          priority: 'medium',
          type: 'lifestyle',
          description: '基于您的睡眠监测数据，制定个性化的睡眠改善计划',
          benefits: ['提高睡眠质量', '增强免疫力', '改善记忆力', '调节情绪'],
          steps: [
            '建立固定的睡眠时间',
            '睡前1小时避免电子设备',
            '保持卧室温度在18-22°C',
            '使用遮光窗帘',
            '睡前进行放松练习'
          ],
          duration: '每晚8小时',
          difficulty: 2,
          effectiveness: 82,
          personalizedScore: 85,
          tags: ['睡眠', '作息', '放松'],
          isCompleted: true,
          isFavorited: true,
          rating: 4.8,
          createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000)
        },
        {
          id: '4',
          title: '压力管理冥想练习',
          category: 'mental',
          priority: 'medium',
          type: 'mental',
          description: '通过正念冥想来缓解日常压力，改善心理健康状态',
          benefits: ['减轻压力', '提高专注力', '改善情绪', '增强自我意识'],
          steps: [
            '找一个安静的环境',
            '采用舒适的坐姿',
            '专注于呼吸',
            '观察思绪但不评判',
            '从5分钟开始逐渐延长'
          ],
          duration: '10-20分钟/天',
          difficulty: 1,
          effectiveness: 75,
          personalizedScore: 80,
          tags: ['冥想', '压力', '心理健康'],
          isCompleted: false,
          isFavorited: false,
          rating: 4.0,
          createdAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000)
        },
        {
          id: '5',
          title: '定期体检提醒',
          category: 'medical',
          priority: 'low',
          type: 'medical',
          description: '根据您的年龄和健康状况，建议定期进行相关健康检查',
          benefits: ['早期发现疾病', '预防保健', '健康监测', '安心保障'],
          steps: [
            '预约体检时间',
            '准备体检前注意事项',
            '完成各项检查',
            '获取体检报告',
            '咨询医生建议'
          ],
          duration: '每年1-2次',
          difficulty: 1,
          effectiveness: 90,
          personalizedScore: 75,
          tags: ['体检', '预防', '健康监测'],
          isCompleted: false,
          isFavorited: false,
          rating: 4.3,
          createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
        }
      ];

      setHealthAdvices(mockAdvices);
      setLoading(false);
    }, 1000);
  };

  // 过滤建议
  const filteredAdvices = healthAdvices.filter(advice => {
    const categoryMatch = selectedCategory === 'all' || advice.category === selectedCategory;
    const priorityMatch = selectedPriority === 'all' || advice.priority === selectedPriority;
    return categoryMatch && priorityMatch;
  });

  // 获取优先级颜色
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return '#ff4d4f';
      case 'medium': return '#faad14';
      case 'low': return '#52c41a';
      default: return '#d9d9d9';
    }
  };

  // 获取优先级文本
  const getPriorityText = (priority: string) => {
    switch (priority) {
      case 'high': return '高';
      case 'medium': return '中';
      case 'low': return '低';
      default: return '未知';
    }
  };

  // 切换收藏状态
  const toggleFavorite = (adviceId: string) => {
    setHealthAdvices(prev => 
      prev.map(advice => 
        advice.id === adviceId 
          ? { ...advice, isFavorited: !advice.isFavorited }
          : advice
      )
    );
  };

  // 标记完成
  const markCompleted = (adviceId: string) => {
    setHealthAdvices(prev => 
      prev.map(advice => 
        advice.id === adviceId 
          ? { ...advice, isCompleted: !advice.isCompleted }
          : advice
      )
    );
    message.success('建议状态已更新');
  };

  // 提交反馈
  const handleFeedback = async (values: any) => {
    try {
      // 这里应该调用API保存反馈
      console.log('提交反馈:', values);
      message.success('反馈提交成功');
      setFeedbackVisible(false);
      form.resetFields();
    } catch (error) {
      message.error('提交失败，请重试');
    }
  };

  // 渲染建议卡片
  const renderAdviceCard = (advice: HealthAdvice) => {
    const categoryInfo = categories.find(cat => cat.key === advice.category);
    
    return (
      <Card
        key={advice.id}
        style={{ marginBottom: '16px' }}
        title={
          <Space>
            {categoryInfo?.icon}
            <Text strong>{advice.title}</Text>
            <Tag color={getPriorityColor(advice.priority)}>
              {getPriorityText(advice.priority)}优先级
            </Tag>
            {advice.isCompleted && <Badge status="success" text="已完成" />}
          </Space>
        }
        extra={
          <Space>
            <Rate disabled value={advice.rating} style={{ fontSize: '14px' }} />
            <Text type="secondary">({advice.rating})</Text>
          </Space>
        }
        actions={[
          <Tooltip title={advice.isFavorited ? '取消收藏' : '收藏'}>
            <StarOutlined
              style={{ color: advice.isFavorited ? '#faad14' : '#999' }}
              onClick={() => toggleFavorite(advice.id)}
            />
          </Tooltip>,
          <Tooltip title={advice.isCompleted ? '标记未完成' : '标记完成'}>
            <CheckCircleOutlined
              style={{ color: advice.isCompleted ? '#52c41a' : '#999' }}
              onClick={() => markCompleted(advice.id)}
            />
          </Tooltip>,
          <Tooltip title="分享">
            <ShareAltOutlined />
          </Tooltip>,
          <Tooltip title="反馈">
            <LikeOutlined 
              onClick={() => {
                setSelectedAdvice(advice);
                setFeedbackVisible(true);
              }}
            />
          </Tooltip>
        ]}
      >
        <div style={{ marginBottom: '16px' }}>
          <Paragraph>{advice.description}</Paragraph>
        </div>

        <Row gutter={[16, 16]} style={{ marginBottom: '16px' }}>
          <Col span={8}>
            <Statistic
              title="个性化匹配"
              value={advice.personalizedScore}
              suffix="%"
              valueStyle={{ fontSize: '16px', color: '#1890ff' }}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="预期效果"
              value={advice.effectiveness}
              suffix="%"
              valueStyle={{ fontSize: '16px', color: '#52c41a' }}
            />
          </Col>
          <Col span={8}>
            <div>
              <Text type="secondary">难度等级</Text>
              <div>
                <Rate disabled value={advice.difficulty} count={5} style={{ fontSize: '14px' }} />
              </div>
            </div>
          </Col>
        </Row>

        <div style={{ marginBottom: '16px' }}>
          <Text strong>预期收益:</Text>
          <div style={{ marginTop: '8px' }}>
            <Space wrap>
              {advice.benefits.map((benefit, index) => (
                <Tag key={index} color="green">{benefit}</Tag>
              ))}
            </Space>
          </div>
        </div>

        <Collapse ghost>
          <Panel header="查看详细步骤" key="1">
            <Timeline size="small">
              {advice.steps.map((step, index) => (
                <Timeline.Item key={index}>
                  <Text>{step}</Text>
                </Timeline.Item>
              ))}
            </Timeline>
          </Panel>
        </Collapse>

        <div style={{ marginTop: '16px' }}>
          <Space wrap>
            <Text type="secondary">
              <ClockCircleOutlined /> {advice.duration}
            </Text>
            {advice.dueDate && (
              <Text type="secondary">
                <CalendarOutlined /> 建议完成时间: {advice.dueDate.toLocaleDateString()}
              </Text>
            )}
          </Space>
        </div>

        <div style={{ marginTop: '12px' }}>
          <Space wrap>
            {advice.tags.map((tag, index) => (
              <Tag key={index}>{tag}</Tag>
            ))}
          </Space>
        </div>
      </Card>
    );
  };

  // 渲染统计信息
  const renderStats = () => {
    const totalAdvices = healthAdvices.length;
    const completedAdvices = healthAdvices.filter(a => a.isCompleted).length;
    const highPriorityAdvices = healthAdvices.filter(a => a.priority === 'high').length;
    const favoritedAdvices = healthAdvices.filter(a => a.isFavorited).length;

    return (
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Statistic
            title="总建议数"
            value={totalAdvices}
            prefix={<BulbOutlined />}
            valueStyle={{ color: '#1890ff' }}
          />
        </Col>
        <Col span={6}>
          <Statistic
            title="已完成"
            value={completedAdvices}
            prefix={<CheckCircleOutlined />}
            valueStyle={{ color: '#52c41a' }}
          />
        </Col>
        <Col span={6}>
          <Statistic
            title="高优先级"
            value={highPriorityAdvices}
            prefix={<FireOutlined />}
            valueStyle={{ color: '#ff4d4f' }}
          />
        </Col>
        <Col span={6}>
          <Statistic
            title="已收藏"
            value={favoritedAdvices}
            prefix={<StarOutlined />}
            valueStyle={{ color: '#faad14' }}
          />
        </Col>
      </Row>
    );
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <BulbOutlined style={{ color: '#faad14', marginRight: '8px' }} />
          健康建议
        </Title>
        <Paragraph>
          基于您的健康数据和AI分析，为您提供个性化的健康改善建议
        </Paragraph>
      </div>

      {/* 统计信息 */}
      <Card style={{ marginBottom: '24px' }}>
        {renderStats()}
      </Card>

      {/* 筛选和操作栏 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Select
                value={selectedCategory}
                onChange={setSelectedCategory}
                style={{ width: 150 }}
                placeholder="选择分类"
              >
                {categories.map(category => (
                  <Select.Option key={category.key} value={category.key}>
                    <Space>
                      {category.icon}
                      {category.name}
                    </Space>
                  </Select.Option>
                ))}
              </Select>
              
              <Select
                value={selectedPriority}
                onChange={setSelectedPriority}
                style={{ width: 120 }}
                placeholder="选择优先级"
              >
                <Select.Option value="all">全部优先级</Select.Option>
                <Select.Option value="high">高优先级</Select.Option>
                <Select.Option value="medium">中优先级</Select.Option>
                <Select.Option value="low">低优先级</Select.Option>
              </Select>
            </Space>
          </Col>
          <Col>
            <Space>
              <Button icon={<PlusOutlined />} type="primary">
                自定义建议
              </Button>
              <Button icon={<SettingOutlined />}>
                偏好设置
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={[24, 24]}>
        {/* 左侧建议列表 */}
        <Col xs={24} lg={16}>
          <div>
            {filteredAdvices.length === 0 ? (
              <Card style={{ textAlign: 'center', padding: '60px 20px' }}>
                <BulbOutlined style={{ fontSize: '64px', color: '#d9d9d9', marginBottom: '16px' }} />
                <div>
                  <Text>暂无符合条件的建议</Text>
                  <br />
                  <Text type="secondary">尝试调整筛选条件或添加更多健康数据</Text>
                </div>
              </Card>
            ) : (
              filteredAdvices.map(renderAdviceCard)
            )}
          </div>
        </Col>

        {/* 右侧侧边栏 */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            {/* 分类统计 */}
            <Card title="分类统计">
              <Space direction="vertical" style={{ width: '100%' }}>
                {categories.slice(1).map(category => {
                  const count = healthAdvices.filter(a => a.category === category.key).length;
                  return (
                    <div key={category.key} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Space>
                        {category.icon}
                        <Text>{category.name}</Text>
                      </Space>
                      <Badge count={count} style={{ backgroundColor: category.color }} />
                    </div>
                  );
                })}
              </Space>
            </Card>

            {/* 完成进度 */}
            <Card title="完成进度">
              <div style={{ textAlign: 'center', marginBottom: '16px' }}>
                <Progress
                  type="circle"
                  percent={Math.round((healthAdvices.filter(a => a.isCompleted).length / healthAdvices.length) * 100)}
                  strokeColor="#52c41a"
                />
              </div>
              <Text type="secondary">
                已完成 {healthAdvices.filter(a => a.isCompleted).length} / {healthAdvices.length} 项建议
              </Text>
            </Card>

            {/* 最近活动 */}
            <Card title="最近活动">
              <Timeline size="small">
                <Timeline.Item color="green">
                  <Text>完成了睡眠质量优化方案</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>2小时前</Text>
                </Timeline.Item>
                <Timeline.Item color="blue">
                  <Text>收藏了地中海饮食模式</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>昨天</Text>
                </Timeline.Item>
                <Timeline.Item>
                  <Text>获得新的运动建议</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>2天前</Text>
                </Timeline.Item>
              </Timeline>
            </Card>

            {/* 成就徽章 */}
            <Card title="成就徽章">
              <Space wrap>
                <Tooltip title="连续7天完成运动建议">
                  <Badge.Ribbon text="NEW">
                    <Avatar icon={<TrophyOutlined />} style={{ backgroundColor: '#faad14' }} />
                  </Badge.Ribbon>
                </Tooltip>
                <Tooltip title="完成10项健康建议">
                  <Avatar icon={<StarOutlined />} style={{ backgroundColor: '#52c41a' }} />
                </Tooltip>
                <Tooltip title="坚持健康饮食30天">
                  <Avatar icon={<HeartOutlined />} style={{ backgroundColor: '#1890ff' }} />
                </Tooltip>
              </Space>
            </Card>
          </Space>
        </Col>
      </Row>

      {/* 反馈模态框 */}
      <Modal
        title="建议反馈"
        visible={feedbackVisible}
        onCancel={() => setFeedbackVisible(false)}
        onOk={() => form.submit()}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFeedback}
        >
          <Form.Item
            name="rating"
            label="评分"
            rules={[{ required: true, message: '请给出评分' }]}
          >
            <Rate />
          </Form.Item>

          <Form.Item
            name="helpful"
            label="是否有帮助"
            rules={[{ required: true, message: '请选择是否有帮助' }]}
          >
            <Select placeholder="请选择">
              <Select.Option value="yes">非常有帮助</Select.Option>
              <Select.Option value="somewhat">有一定帮助</Select.Option>
              <Select.Option value="no">没有帮助</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="feedback"
            label="详细反馈"
          >
            <TextArea
              rows={4}
              placeholder="请分享您的使用体验和建议..."
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default HealthAdvice;