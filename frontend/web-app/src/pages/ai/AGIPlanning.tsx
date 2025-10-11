import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Input,
  Select,
  Typography,
  Space,
  Spin,
  Alert,
  Steps,
  Progress,
  Tag,
  Divider,
  Tooltip,
  Badge,
  Timeline,
  Statistic,
  Table,
  Collapse,
  Tree,
  Modal,
} from 'antd';
import {
  ProjectOutlined,
  RocketOutlined,
  BulbOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  StarOutlined,
  WarningOutlined,
  DollarOutlined,
  TeamOutlined,
  CalendarOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import { advancedAIService } from '../../services/advancedAiService';
import type { PlanningRequest, PlanningResponse } from '../../services/advancedAiService';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;
const { Panel } = Collapse;

interface PlanningSession {
  id: string;
  goal: string;
  response?: PlanningResponse;
  status: 'pending' | 'processing' | 'completed' | 'error';
  startTime: Date;
  endTime?: Date;
  error?: string;
  constraints?: Record<string, any>;
  context?: Record<string, any>;
}

const AGIPlanning: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [goal, setGoal] = useState('');
  const [constraints, setConstraints] = useState('');
  const [context, setContext] = useState('');
  const [detailLevel, setDetailLevel] = useState<'low' | 'medium' | 'high'>('medium');
  const [includeRisks, setIncludeRisks] = useState(true);
  const [includeTimeline, setIncludeTimeline] = useState(true);
  const [priority, setPriority] = useState(2);
  const [sessions, setSessions] = useState<PlanningSession[]>([]);
  const [currentSession, setCurrentSession] = useState<PlanningSession | null>(null);
  const [selectedPlan, setSelectedPlan] = useState<PlanningResponse['plan'] | null>(null);
  const [planModalVisible, setPlanModalVisible] = useState(false);
  const [stats, setStats] = useState({
    totalPlans: 0,
    avgSuccessRate: 0,
    avgPlanningTime: 0,
    completedProjects: 0,
  });

  // 预设规划模板
  const planningTemplates = [
    {
      title: '产品开发',
      goal: '开发一个智能客服系统',
      constraints: '预算: 100万元, 时间: 6个月, 团队: 8人',
      context: '行业: 电商, 规模: 中型企业, 技术栈: React + Node.js',
    },
    {
      title: '市场推广',
      goal: '制定新产品市场推广策略',
      constraints: '预算: 50万元, 时间: 3个月, 目标用户: 年轻群体',
      context: '产品类型: 移动应用, 竞争激烈度: 高, 品牌知名度: 低',
    },
    {
      title: '团队建设',
      goal: '组建高效的AI研发团队',
      constraints: '预算: 200万元/年, 时间: 4个月, 目标规模: 15人',
      context: '公司阶段: 成长期, 技术要求: 深度学习, 地点: 北京',
    },
    {
      title: '数字化转型',
      goal: '传统制造企业数字化转型',
      constraints: '预算: 500万元, 时间: 12个月, 现有员工: 200人',
      context: '行业: 制造业, 现状: 传统流程, 目标: 智能制造',
    },
  ];

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = () => {
    setStats({
      totalPlans: 89,
      avgSuccessRate: 0.91,
      avgPlanningTime: 4.2,
      completedProjects: 67,
    });
  };

  const handleSubmit = async () => {
    if (!goal.trim()) return;

    const sessionId = `session_${Date.now()}`;
    const newSession: PlanningSession = {
      id: sessionId,
      goal: goal.trim(),
      status: 'processing',
      startTime: new Date(),
      constraints: constraints ? JSON.parse(`{${constraints}}`) : undefined,
      context: context ? JSON.parse(`{${context}}`) : undefined,
    };

    setSessions(prev => [newSession, ...prev]);
    setCurrentSession(newSession);
    setLoading(true);

    try {
      const request: PlanningRequest = {
        goal: goal.trim(),
        constraints: newSession.constraints,
        context: newSession.context,
        requirements: {
          detail_level: detailLevel,
          include_risks: includeRisks,
          include_timeline: includeTimeline,
        },
        priority,
        timeout: 120,
      };

      const response = await advancedAIService.planning(request);

      if (response.success && response.data) {
        const completedSession: PlanningSession = {
          ...newSession,
          response: response.data,
          status: 'completed',
          endTime: new Date(),
        };

        setSessions(prev => prev.map(s => s.id === sessionId ? completedSession : s));
        setCurrentSession(completedSession);
      } else {
        throw new Error(response.error?.message || '规划失败');
      }
    } catch (error) {
      const errorSession: PlanningSession = {
        ...newSession,
        status: 'error',
        endTime: new Date(),
        error: error instanceof Error ? error.message : '未知错误',
      };

      setSessions(prev => prev.map(s => s.id === sessionId ? errorSession : s));
      setCurrentSession(errorSession);
    } finally {
      setLoading(false);
    }
  };

  const handleTemplateClick = (template: typeof planningTemplates[0]) => {
    setGoal(template.goal);
    setConstraints(template.constraints);
    setContext(template.context);
  };

  const showPlanDetails = (plan: PlanningResponse['plan']) => {
    setSelectedPlan(plan);
    setPlanModalVisible(true);
  };

  const renderPlanSteps = (steps: PlanningResponse['plan']['steps']) => {
    const columns = [
      {
        title: '步骤',
        dataIndex: 'title',
        key: 'title',
        render: (text: string, record: any) => (
          <Space direction="vertical" size="small">
            <Text strong>{text}</Text>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              {record.description}
            </Text>
          </Space>
        ),
      },
      {
        title: '预估时间',
        dataIndex: 'estimated_time',
        key: 'estimated_time',
        width: 120,
        render: (time: string) => (
          <Tag icon={<ClockCircleOutlined />} color="blue">
            {time}
          </Tag>
        ),
      },
      {
        title: '所需资源',
        dataIndex: 'resources',
        key: 'resources',
        width: 200,
        render: (resources: string[]) => (
          <Space wrap>
            {resources.map((resource, index) => (
              <Tag key={index} color="green">
                {resource}
              </Tag>
            ))}
          </Space>
        ),
      },
      {
        title: '风险',
        dataIndex: 'risks',
        key: 'risks',
        width: 200,
        render: (risks: string[]) => (
          <Space wrap>
            {risks.map((risk, index) => (
              <Tag key={index} color="orange" icon={<WarningOutlined />}>
                {risk}
              </Tag>
            ))}
          </Space>
        ),
      },
    ];

    return (
      <Table
        dataSource={steps.map((step, index) => ({ ...step, key: index }))}
        columns={columns}
        pagination={false}
        size="small"
      />
    );
  };

  const renderPlanTree = (steps: PlanningResponse['plan']['steps']) => {
    const treeData = steps.map((step, index) => ({
      title: (
        <Space>
          <Text strong>{step.title}</Text>
          <Tag color="blue">{step.estimated_time}</Tag>
        </Space>
      ),
      key: step.id,
      children: step.dependencies.map(dep => ({
        title: <Text type="secondary">依赖: {dep}</Text>,
        key: `${step.id}_${dep}`,
      })),
    }));

    return <Tree treeData={treeData} defaultExpandAll />;
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '32px', textAlign: 'center' }}>
        <Title level={1} style={{ marginBottom: '8px' }}>
          <ProjectOutlined style={{ color: '#1890ff', marginRight: '12px' }} />
          AGI 智能规划
        </Title>
        <Paragraph style={{ fontSize: '16px', color: '#666', maxWidth: '600px', margin: '0 auto' }}>
          基于AGI技术的智能规划系统，帮助您制定详细的项目计划，优化资源配置，预测风险
        </Paragraph>
      </div>

      {/* 统计数据 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '32px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总规划数"
              value={stats.totalPlans}
              prefix={<ProjectOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="平均成功率"
              value={stats.avgSuccessRate}
              precision={1}
              suffix="%"
              prefix={<StarOutlined style={{ color: '#52c41a' }} />}
              formatter={(value) => `${((value as number) * 100).toFixed(1)}%`}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="平均规划时间"
              value={stats.avgPlanningTime}
              precision={1}
              suffix="秒"
              prefix={<ClockCircleOutlined style={{ color: '#fa8c16' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="已完成项目"
              value={stats.completedProjects}
              prefix={<CheckCircleOutlined style={{ color: '#f5222d' }} />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]}>
        {/* 规划输入区域 */}
        <Col xs={24} lg={12}>
          <Card title="规划目标" extra={<Badge count="智能" style={{ backgroundColor: '#52c41a' }} />}>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {/* 目标输入 */}
              <div>
                <Text strong style={{ marginBottom: '8px', display: 'block' }}>
                  规划目标：
                  <Tooltip title="请清晰描述您要实现的目标">
                    <InfoCircleOutlined style={{ marginLeft: '4px', color: '#1890ff' }} />
                  </Tooltip>
                </Text>
                <TextArea
                  value={goal}
                  onChange={(e) => setGoal(e.target.value)}
                  placeholder="请输入您要实现的目标..."
                  rows={3}
                  maxLength={500}
                  showCount
                />
              </div>

              {/* 约束条件 */}
              <div>
                <Text strong style={{ marginBottom: '8px', display: 'block' }}>
                  约束条件：
                  <Tooltip title="格式: 预算: 100万, 时间: 6个月, 团队: 10人">
                    <InfoCircleOutlined style={{ marginLeft: '4px', color: '#1890ff' }} />
                  </Tooltip>
                </Text>
                <TextArea
                  value={constraints}
                  onChange={(e) => setConstraints(e.target.value)}
                  placeholder="预算: 100万元, 时间: 6个月, 团队: 8人"
                  rows={2}
                  maxLength={300}
                  showCount
                />
              </div>

              {/* 上下文信息 */}
              <div>
                <Text strong style={{ marginBottom: '8px', display: 'block' }}>
                  上下文信息：
                </Text>
                <TextArea
                  value={context}
                  onChange={(e) => setContext(e.target.value)}
                  placeholder="行业: 电商, 规模: 中型企业, 技术栈: React"
                  rows={2}
                  maxLength={300}
                  showCount
                />
              </div>

              {/* 规划设置 */}
              <Row gutter={16}>
                <Col span={12}>
                  <Text strong style={{ marginBottom: '8px', display: 'block' }}>详细程度：</Text>
                  <Select
                    value={detailLevel}
                    onChange={setDetailLevel}
                    style={{ width: '100%' }}
                  >
                    <Option value="low">简要 - 概要规划</Option>
                    <Option value="medium">标准 - 详细规划</Option>
                    <Option value="high">详尽 - 深度规划</Option>
                  </Select>
                </Col>
                <Col span={12}>
                  <Text strong style={{ marginBottom: '8px', display: 'block' }}>优先级：</Text>
                  <Select
                    value={priority}
                    onChange={setPriority}
                    style={{ width: '100%' }}
                  >
                    <Option value={1}>低优先级</Option>
                    <Option value={2}>中优先级</Option>
                    <Option value={3}>高优先级</Option>
                    <Option value={4}>紧急</Option>
                  </Select>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={12}>
                  <Button
                    type={includeRisks ? 'primary' : 'default'}
                    onClick={() => setIncludeRisks(!includeRisks)}
                    icon={<WarningOutlined />}
                    block
                  >
                    {includeRisks ? '包含' : '不含'}风险分析
                  </Button>
                </Col>
                <Col span={12}>
                  <Button
                    type={includeTimeline ? 'primary' : 'default'}
                    onClick={() => setIncludeTimeline(!includeTimeline)}
                    icon={<CalendarOutlined />}
                    block
                  >
                    {includeTimeline ? '包含' : '不含'}时间线
                  </Button>
                </Col>
              </Row>

              {/* 提交按钮 */}
              <Button
                type="primary"
                size="large"
                block
                icon={<RocketOutlined />}
                loading={loading}
                onClick={handleSubmit}
                disabled={!goal.trim()}
              >
                {loading ? '规划中...' : '开始规划'}
              </Button>
            </Space>
          </Card>

          {/* 规划模板 */}
          <Card title="规划模板" style={{ marginTop: '24px' }}>
            <Row gutter={[16, 16]}>
              {planningTemplates.map((template, index) => (
                <Col xs={24} sm={12} key={index}>
                  <Card
                    size="small"
                    hoverable
                    onClick={() => handleTemplateClick(template)}
                    style={{ cursor: 'pointer' }}
                  >
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <Text strong>{template.title}</Text>
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {template.goal.substring(0, 40)}...
                      </Text>
                      <div>
                        <Tag color="blue" size="small">模板</Tag>
                        <Tag color="green" size="small">推荐</Tag>
                      </div>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>

        {/* 规划结果区域 */}
        <Col xs={24} lg={12}>
          {currentSession ? (
            <Card
              title={
                <Space>
                  <ProjectOutlined />
                  规划结果
                </Space>
              }
              extra={
                currentSession.status === 'processing' ? (
                  <Spin size="small" />
                ) : currentSession.status === 'completed' ? (
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : currentSession.status === 'error' ? (
                  <ExclamationCircleOutlined style={{ color: '#f5222d' }} />
                ) : null
              }
            >
              {currentSession.status === 'processing' && (
                <div style={{ textAlign: 'center', padding: '40px' }}>
                  <Spin size="large" />
                  <div style={{ marginTop: '16px' }}>
                    <Text>AGI正在制定详细规划...</Text>
                  </div>
                </div>
              )}

              {currentSession.status === 'error' && (
                <Alert
                  message="规划失败"
                  description={currentSession.error}
                  type="error"
                  showIcon
                />
              )}

              {currentSession.status === 'completed' && currentSession.response && (
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                  {/* 规划概览 */}
                  <Card size="small" style={{ backgroundColor: '#f6ffed', border: '1px solid #b7eb8f' }}>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Statistic
                          title="总时长"
                          value={currentSession.response.plan.timeline}
                          prefix={<CalendarOutlined />}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic
                          title="预估成本"
                          value={currentSession.response.plan.total_cost}
                          prefix={<DollarOutlined />}
                          suffix="万元"
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic
                          title="成功概率"
                          value={currentSession.response.plan.success_probability}
                          precision={1}
                          suffix="%"
                          prefix={<StarOutlined />}
                          formatter={(value) => `${((value as number) * 100).toFixed(1)}%`}
                        />
                      </Col>
                    </Row>
                  </Card>

                  {/* 规划步骤 */}
                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                      <Text strong>执行步骤：</Text>
                      <Button
                        type="link"
                        onClick={() => showPlanDetails(currentSession.response!.plan)}
                        icon={<BarChartOutlined />}
                      >
                        查看详情
                      </Button>
                    </div>
                    
                    <Collapse>
                      {currentSession.response.plan.steps.slice(0, 3).map((step, index) => (
                        <Panel
                          header={
                            <Space>
                              <Text strong>{step.title}</Text>
                              <Tag color="blue">{step.estimated_time}</Tag>
                            </Space>
                          }
                          key={index}
                        >
                          <Space direction="vertical" style={{ width: '100%' }}>
                            <Text>{step.description}</Text>
                            <div>
                              <Text strong>所需资源：</Text>
                              <Space wrap style={{ marginTop: '8px' }}>
                                {step.resources.map((resource, idx) => (
                                  <Tag key={idx} color="green">{resource}</Tag>
                                ))}
                              </Space>
                            </div>
                            {step.risks.length > 0 && (
                              <div>
                                <Text strong>潜在风险：</Text>
                                <Space wrap style={{ marginTop: '8px' }}>
                                  {step.risks.map((risk, idx) => (
                                    <Tag key={idx} color="orange" icon={<WarningOutlined />}>
                                      {risk}
                                    </Tag>
                                  ))}
                                </Space>
                              </div>
                            )}
                          </Space>
                        </Panel>
                      ))}
                    </Collapse>
                  </div>

                  {/* 成功概率进度条 */}
                  <div>
                    <Text strong style={{ marginBottom: '8px', display: 'block' }}>项目成功概率：</Text>
                    <Progress
                      percent={currentSession.response.plan.success_probability * 100}
                      status={currentSession.response.plan.success_probability > 0.8 ? 'success' : 
                             currentSession.response.plan.success_probability > 0.6 ? 'active' : 'exception'}
                      strokeColor={currentSession.response.plan.success_probability > 0.8 ? '#52c41a' : 
                                  currentSession.response.plan.success_probability > 0.6 ? '#1890ff' : '#f5222d'}
                    />
                  </div>
                </Space>
              )}
            </Card>
          ) : (
            <Card title="规划结果" style={{ textAlign: 'center', padding: '40px' }}>
              <ProjectOutlined style={{ fontSize: '48px', color: '#d9d9d9', marginBottom: '16px' }} />
              <Text type="secondary">请输入目标开始智能规划</Text>
            </Card>
          )}

          {/* 历史规划 */}
          {sessions.length > 0 && (
            <Card title="规划历史" style={{ marginTop: '24px' }}>
              <Timeline mode="left">
                {sessions.slice(0, 5).map((session) => (
                  <Timeline.Item
                    key={session.id}
                    color={session.status === 'completed' ? 'green' : session.status === 'error' ? 'red' : 'blue'}
                    dot={
                      session.status === 'completed' ? <CheckCircleOutlined /> :
                      session.status === 'error' ? <ExclamationCircleOutlined /> :
                      <ClockCircleOutlined />
                    }
                  >
                    <div style={{ cursor: 'pointer' }} onClick={() => setCurrentSession(session)}>
                      <Text strong>{session.goal.substring(0, 40)}...</Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {session.startTime.toLocaleString()}
                        {session.response && ` · 成功率: ${(session.response.plan.success_probability * 100).toFixed(1)}%`}
                      </Text>
                    </div>
                  </Timeline.Item>
                ))}
              </Timeline>
            </Card>
          )}
        </Col>
      </Row>

      {/* 规划详情模态框 */}
      <Modal
        title="规划详情"
        open={planModalVisible}
        onCancel={() => setPlanModalVisible(false)}
        width={1000}
        footer={null}
      >
        {selectedPlan && (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Card size="small">
              <Row gutter={16}>
                <Col span={6}>
                  <Statistic title="项目ID" value={selectedPlan.id} />
                </Col>
                <Col span={6}>
                  <Statistic title="总时长" value={selectedPlan.timeline} />
                </Col>
                <Col span={6}>
                  <Statistic title="预估成本" value={selectedPlan.total_cost} suffix="万元" />
                </Col>
                <Col span={6}>
                  <Statistic 
                    title="成功概率" 
                    value={selectedPlan.success_probability * 100} 
                    precision={1}
                    suffix="%" 
                  />
                </Col>
              </Row>
            </Card>
            
            <div>
              <Text strong style={{ marginBottom: '16px', display: 'block' }}>执行步骤详情：</Text>
              {renderPlanSteps(selectedPlan.steps)}
            </div>
            
            <div>
              <Text strong style={{ marginBottom: '16px', display: 'block' }}>依赖关系图：</Text>
              {renderPlanTree(selectedPlan.steps)}
            </div>
          </Space>
        )}
      </Modal>
    </div>
  );
};

export default AGIPlanning;