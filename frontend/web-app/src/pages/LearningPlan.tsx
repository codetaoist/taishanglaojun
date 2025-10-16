import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Progress,
  Typography,
  Steps,
  Form,
  Input,
  Select,
  DatePicker,
  TimePicker,
  Modal,
  List,
  Tag,
  Space,
  Statistic,
  Timeline,
  Alert,
  Divider,
  Avatar,
  Badge,
  Tooltip,
  Empty,
  Tabs,
  Checkbox,
  Slider,
  Radio,
  Calendar,
  Popover,
} from 'antd';
import {
  CalendarOutlined,
  ClockCircleOutlined,
  BookOutlined,
  TrophyOutlined,
  AimOutlined,
  FireOutlined,
  ThunderboltOutlined,
  BulbOutlined,
  RocketOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StarOutlined,
  LineChartOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import { Line, Column, Pie } from '@ant-design/plots';
import dayjs, { Dayjs } from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Step } = Steps;

const { Option } = Select;
const { RangePicker } = DatePicker;

// 学习目标类型
const goalTypes = [
  { key: 'skill', name: '技能提升', icon: '🎯', color: '#1890ff' },
  { key: 'certification', name: '认证考试', icon: '🏆', color: '#52c41a' },
  { key: 'project', name: '项目实战', icon: '🚀', color: '#722ed1' },
  { key: 'career', name: '职业发展', icon: '💼', color: '#fa8c16' },
];

// 学习方式
const learningMethods = [
  { key: 'video', name: '视频课程', icon: '📹' },
  { key: 'reading', name: '阅读材料', icon: '📚' },
  { key: 'practice', name: '实践练习', icon: '💻' },
  { key: 'project', name: '项目实战', icon: '🛠️' },
];

// 生成模拟学习计划数据
const generateMockPlans = () => {
  return [
    {
      id: 1,
      title: '前端开发进阶计划',
      description: '从基础到高级的前端开发技能提升计划',
      type: 'skill',
      status: 'active',
      progress: 65,
      startDate: '2024-01-01',
      endDate: '2024-03-31',
      totalHours: 120,
      completedHours: 78,
      courses: ['React进阶', 'TypeScript实战', '前端工程化'],
      milestones: [
        { name: '基础掌握', completed: true, date: '2024-01-15' },
        { name: '进阶学习', completed: true, date: '2024-02-15' },
        { name: '项目实战', completed: false, date: '2024-03-15' },
        { name: '技能认证', completed: false, date: '2024-03-31' },
      ],
    },
    {
      id: 2,
      title: 'AWS云架构师认证',
      description: '准备AWS解决方案架构师认证考试',
      type: 'certification',
      status: 'planning',
      progress: 0,
      startDate: '2024-02-01',
      endDate: '2024-05-31',
      totalHours: 200,
      completedHours: 0,
      courses: ['AWS基础', '云架构设计', '安全最佳实践'],
      milestones: [
        { name: '基础知识', completed: false, date: '2024-02-28' },
        { name: '架构设计', completed: false, date: '2024-03-31' },
        { name: '实践项目', completed: false, date: '2024-04-30' },
        { name: '认证考试', completed: false, date: '2024-05-31' },
      ],
    },
    {
      id: 3,
      title: '全栈电商项目',
      description: '完整的电商系统开发项目实战',
      type: 'project',
      status: 'completed',
      progress: 100,
      startDate: '2023-10-01',
      endDate: '2023-12-31',
      totalHours: 150,
      completedHours: 150,
      courses: ['前端开发', '后端API', '数据库设计', '部署运维'],
      milestones: [
        { name: '需求分析', completed: true, date: '2023-10-15' },
        { name: '前端开发', completed: true, date: '2023-11-15' },
        { name: '后端开发', completed: true, date: '2023-12-01' },
        { name: '项目部署', completed: true, date: '2023-12-31' },
      ],
    },
  ];
};

// 生成学习活动数据
const generateLearningActivity = () => {
  const today = dayjs();
  const activities = [];
  
  for (let i = 29; i >= 0; i--) {
    const date = today.subtract(i, 'day');
    activities.push({
      date: date.format('YYYY-MM-DD'),
      hours: Math.random() * 4,
      courses: Math.floor(Math.random() * 3),
      exercises: Math.floor(Math.random() * 5),
    });
  }
  
  return activities;
};

const LearningPlan: React.FC = () => {
  const [plans, setPlans] = useState<any[]>([]);
  const [selectedPlan, setSelectedPlan] = useState<any>(null);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [learningActivity] = useState(generateLearningActivity());
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  useEffect(() => {
    setPlans(generateMockPlans());
  }, []);

  const handleCreatePlan = (values: any) => {
    const newPlan = {
      id: Date.now(),
      ...values,
      status: 'planning',
      progress: 0,
      completedHours: 0,
      startDate: values.dateRange[0].format('YYYY-MM-DD'),
      endDate: values.dateRange[1].format('YYYY-MM-DD'),
      milestones: values.milestones?.map((name: string, index: number) => ({
        name,
        completed: false,
        date: dayjs(values.dateRange[0]).add((index + 1) * 30, 'day').format('YYYY-MM-DD'),
      })) || [],
    };
    
    setPlans([...plans, newPlan]);
    setCreateModalVisible(false);
    form.resetFields();
  };

  const handleEditPlan = (values: any) => {
    const updatedPlans = plans.map(plan =>
      plan.id === selectedPlan.id
        ? {
            ...plan,
            ...values,
            startDate: values.dateRange[0].format('YYYY-MM-DD'),
            endDate: values.dateRange[1].format('YYYY-MM-DD'),
          }
        : plan
    );
    
    setPlans(updatedPlans);
    setEditModalVisible(false);
    setSelectedPlan(null);
    editForm.resetFields();
  };

  const handleDeletePlan = (planId: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个学习计划吗？',
      onOk: () => {
        setPlans(plans.filter(plan => plan.id !== planId));
      },
    });
  };

  const handleStartPlan = (planId: number) => {
    setPlans(plans.map(plan =>
      plan.id === planId ? { ...plan, status: 'active' } : plan
    ));
  };

  const handlePausePlan = (planId: number) => {
    setPlans(plans.map(plan =>
      plan.id === planId ? { ...plan, status: 'paused' } : plan
    ));
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'planning': return 'blue';
      case 'paused': return 'orange';
      case 'completed': return 'purple';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '进行中';
      case 'planning': return '计划中';
      case 'paused': return '已暂停';
      case 'completed': return '已完成';
      default: return '未知';
    }
  };

  const renderPlanCard = (plan: any) => (
    <Card
      key={plan.id}
      title={
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Text strong>{plan.title}</Text>
            <Tag color={getStatusColor(plan.status)}>
              {getStatusText(plan.status)}
            </Tag>
          </Space>
          <Space>
            {plan.status === 'planning' && (
              <Button
                type="primary"
                size="small"
                icon={<PlayCircleOutlined />}
                onClick={() => handleStartPlan(plan.id)}
              >
                开始
              </Button>
            )}
            {plan.status === 'active' && (
              <Button
                size="small"
                icon={<PauseCircleOutlined />}
                onClick={() => handlePausePlan(plan.id)}
              >
                暂停
              </Button>
            )}
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedPlan(plan);
                editForm.setFieldsValue({
                  ...plan,
                  dateRange: [dayjs(plan.startDate), dayjs(plan.endDate)],
                });
                setEditModalVisible(true);
              }}
            />
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeletePlan(plan.id)}
            />
          </Space>
        </div>
      }
      extra={
        <div style={{ textAlign: 'right' }}>
          <Text type="secondary">{plan.startDate} ~ {plan.endDate}</Text>
        </div>
      }
    >
      <Paragraph ellipsis={{ rows: 2 }}>{plan.description}</Paragraph>
      
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={12}>
          <Statistic
            title="学习进度"
            value={plan.progress}
            suffix="%"
            valueStyle={{ fontSize: 16 }}
          />
          <Progress percent={plan.progress} size="small" />
        </Col>
        <Col span={12}>
          <Statistic
            title="学习时长"
            value={plan.completedHours}
            suffix={`/ ${plan.totalHours}h`}
            valueStyle={{ fontSize: 16 }}
          />
          <Progress 
            percent={Math.round((plan.completedHours / plan.totalHours) * 100)} 
            size="small" 
          />
        </Col>
      </Row>

      <div style={{ marginBottom: 16 }}>
        <Text strong>课程内容：</Text>
        <div style={{ marginTop: 8 }}>
          {plan.courses.map((course: string) => (
            <Tag key={course} color="blue" style={{ marginBottom: 4 }}>
              {course}
            </Tag>
          ))}
        </div>
      </div>

      <div>
        <Text strong>学习里程碑：</Text>
        <Timeline size="small" style={{ marginTop: 8 }}>
          {plan.milestones.map((milestone: any, index: number) => (
            <Timeline.Item
              key={index}
              color={milestone.completed ? 'green' : 'gray'}
              dot={milestone.completed ? <CheckCircleOutlined /> : <ClockCircleOutlined />}
            >
              <div>
                <Text strong={milestone.completed}>{milestone.name}</Text>
                <div>
                  <Text type="secondary">{milestone.date}</Text>
                </div>
              </div>
            </Timeline.Item>
          ))}
        </Timeline>
      </div>
    </Card>
  );

  const renderCreateModal = () => (
    <Modal
      title="创建学习计划"
      open={createModalVisible}
      onCancel={() => {
        setCreateModalVisible(false);
        form.resetFields();
      }}
      footer={null}
      width={600}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleCreatePlan}
      >
        <Form.Item
          name="title"
          label="计划标题"
          rules={[{ required: true, message: '请输入计划标题' }]}
        >
          <Input placeholder="输入学习计划标题" />
        </Form.Item>

        <Form.Item
          name="description"
          label="计划描述"
          rules={[{ required: true, message: '请输入计划描述' }]}
        >
          <Input.TextArea rows={3} placeholder="描述您的学习目标和计划" />
        </Form.Item>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="type"
              label="计划类型"
              rules={[{ required: true, message: '请选择计划类型' }]}
            >
              <Select placeholder="选择计划类型">
                {goalTypes.map(type => (
                  <Option key={type.key} value={type.key}>
                    {type.icon} {type.name}
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="totalHours"
              label="预计学习时长(小时)"
              rules={[{ required: true, message: '请输入预计学习时长' }]}
            >
              <Input type="number" placeholder="100" />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item
          name="dateRange"
          label="学习周期"
          rules={[{ required: true, message: '请选择学习周期' }]}
        >
          <RangePicker style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item
          name="courses"
          label="课程内容"
        >
          <Select
            mode="tags"
            placeholder="输入课程名称，按回车添加"
            style={{ width: '100%' }}
          />
        </Form.Item>

        <Form.Item
          name="milestones"
          label="学习里程碑"
        >
          <Select
            mode="tags"
            placeholder="输入里程碑名称，按回车添加"
            style={{ width: '100%' }}
          />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button onClick={() => {
              setCreateModalVisible(false);
              form.resetFields();
            }}>
              取消
            </Button>
            <Button type="primary" htmlType="submit">
              创建计划
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );

  const renderEditModal = () => (
    <Modal
      title="编辑学习计划"
      open={editModalVisible}
      onCancel={() => {
        setEditModalVisible(false);
        setSelectedPlan(null);
        editForm.resetFields();
      }}
      footer={null}
      width={600}
    >
      <Form
        form={editForm}
        layout="vertical"
        onFinish={handleEditPlan}
      >
        <Form.Item
          name="title"
          label="计划标题"
          rules={[{ required: true, message: '请输入计划标题' }]}
        >
          <Input placeholder="输入学习计划标题" />
        </Form.Item>

        <Form.Item
          name="description"
          label="计划描述"
          rules={[{ required: true, message: '请输入计划描述' }]}
        >
          <Input.TextArea rows={3} placeholder="描述您的学习目标和计划" />
        </Form.Item>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="type"
              label="计划类型"
              rules={[{ required: true, message: '请选择计划类型' }]}
            >
              <Select placeholder="选择计划类型">
                {goalTypes.map(type => (
                  <Option key={type.key} value={type.key}>
                    {type.icon} {type.name}
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="totalHours"
              label="预计学习时长(小时)"
              rules={[{ required: true, message: '请输入预计学习时长' }]}
            >
              <Input type="number" placeholder="100" />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item
          name="dateRange"
          label="学习周期"
          rules={[{ required: true, message: '请选择学习周期' }]}
        >
          <RangePicker style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button onClick={() => {
              setEditModalVisible(false);
              setSelectedPlan(null);
              editForm.resetFields();
            }}>
              取消
            </Button>
            <Button type="primary" htmlType="submit">
              保存修改
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );

  const renderStatistics = () => {
    const activePlans = plans.filter(p => p.status === 'active').length;
    const completedPlans = plans.filter(p => p.status === 'completed').length;
    const totalHours = plans.reduce((sum, p) => sum + p.completedHours, 0);
    const avgProgress = plans.length > 0 ? Math.round(plans.reduce((sum, p) => sum + p.progress, 0) / plans.length) : 0;

    return (
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="进行中计划"
              value={activePlans}
              prefix={<RocketOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="已完成计划"
              value={completedPlans}
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="总学习时长"
              value={totalHours}
              suffix="小时"
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="平均进度"
              value={avgProgress}
              suffix="%"
              prefix={<AimOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>
    );
  };

  const renderLearningChart = () => {
    const chartData = learningActivity.map(item => ({
      date: item.date,
      hours: item.hours,
    }));

    const config = {
      data: chartData,
      xField: 'date',
      yField: 'hours',
      smooth: true,
      color: '#1890ff',
      point: {
        size: 3,
        shape: 'circle',
      },
      meta: {
        hours: {
          alias: '学习时长(小时)',
        },
        date: {
          alias: '日期',
        },
      },
    };

    return (
      <Card title="学习活动趋势" style={{ marginBottom: 16 }}>
        <Line {...config} />
      </Card>
    );
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>
          <CalendarOutlined style={{ marginRight: 8 }} />
          学习计划
        </Title>
        <Text type="secondary">制定个性化学习路径，跟踪学习进度</Text>
      </div>

      {renderStatistics()}

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={16}>
          {renderLearningChart()}
        </Col>
        <Col xs={24} lg={8}>
          <Card title="快速操作">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                block
                onClick={() => setCreateModalVisible(true)}
              >
                创建新计划
              </Button>
              <Button
                icon={<LineChartOutlined />}
                block
                onClick={() => {}}
              >
                学习分析
              </Button>
              <Button
                icon={<AimOutlined />}
                block
                onClick={() => {}}
              >
                目标设置
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      <Tabs 
        defaultActiveKey="all"
        items={[
          {
            key: "all",
            label: "全部计划",
            children: plans.length > 0 ? (
              <Row gutter={[16, 16]}>
                {plans.map(plan => (
                  <Col key={plan.id} xs={24} lg={12}>
                    {renderPlanCard(plan)}
                  </Col>
                ))}
              </Row>
            ) : (
              <Empty
                description="暂无学习计划"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              >
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => setCreateModalVisible(true)}
                >
                  创建第一个学习计划
                </Button>
              </Empty>
            )
          },
          {
            key: "active",
            label: "进行中",
            children: (
              <Row gutter={[16, 16]}>
                {plans.filter(p => p.status === 'active').map(plan => (
                  <Col key={plan.id} xs={24} lg={12}>
                    {renderPlanCard(plan)}
                  </Col>
                ))}
              </Row>
            )
          },
          {
            key: "completed",
            label: "已完成",
            children: (
              <Row gutter={[16, 16]}>
                {plans.filter(p => p.status === 'completed').map(plan => (
                  <Col key={plan.id} xs={24} lg={12}>
                    {renderPlanCard(plan)}
                  </Col>
                ))}
              </Row>
            )
          }
        ]}
      />

      {renderCreateModal()}
      {renderEditModal()}
    </div>
  );
};

export default LearningPlan;