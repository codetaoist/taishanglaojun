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
} from 'antd';
import {
  BulbOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  RocketOutlined,
  StarOutlined,
} from '@ant-design/icons';
import { advancedAIService } from '../../services/advancedAiService';
import type { ReasoningRequest, ReasoningResponse } from '../../services/advancedAiService';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;
const { Step } = Steps;

interface ReasoningSession {
  id: string;
  query: string;
  response?: ReasoningResponse;
  status: 'pending' | 'processing' | 'completed' | 'error';
  startTime: Date;
  endTime?: Date;
  error?: string;
}

const AGIReasoning: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [query, setQuery] = useState('');
  const [context, setContext] = useState('');
  const [reasoningType, setReasoningType] = useState<'deductive' | 'inductive' | 'abductive'>('deductive');
  const [maxSteps, setMaxSteps] = useState(10);
  const [confidenceThreshold, setConfidenceThreshold] = useState(0.8);
  const [sessions, setSessions] = useState<ReasoningSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ReasoningSession | null>(null);
  const [stats, setStats] = useState({
    totalQueries: 0,
    avgConfidence: 0,
    avgProcessingTime: 0,
    successRate: 0,
  });

  // 预设问题示例
  const exampleQueries = [
    {
      title: '逻辑推理',
      query: '如果所有的鸟都会飞，企鹅是鸟，那么企鹅会飞吗？请分析这个逻辑问题。',
      type: 'deductive' as const,
    },
    {
      title: '因果分析',
      query: '公司销售额下降了20%，可能的原因有哪些？请进行系统性分析。',
      type: 'abductive' as const,
    },
    {
      title: '趋势预测',
      query: '基于过去5年的数据，人工智能行业的发展趋势如何？',
      type: 'inductive' as const,
    },
    {
      title: '问题解决',
      query: '如何设计一个高效的在线学习系统？请提供详细的解决方案。',
      type: 'deductive' as const,
    },
  ];

  useEffect(() => {
    // 加载历史会话统计
    loadStats();
  }, []);

  const loadStats = () => {
    // 模拟统计数据
    setStats({
      totalQueries: 156,
      avgConfidence: 0.87,
      avgProcessingTime: 2.3,
      successRate: 94.2,
    });
  };

  const handleSubmit = async () => {
    if (!query.trim()) return;

    const sessionId = `session_${Date.now()}`;
    const newSession: ReasoningSession = {
      id: sessionId,
      query: query.trim(),
      status: 'processing',
      startTime: new Date(),
    };

    setSessions(prev => [newSession, ...prev]);
    setCurrentSession(newSession);
    setLoading(true);

    try {
      const request: ReasoningRequest = {
        query: query.trim(),
        context: context.trim() || undefined,
        reasoning_type: reasoningType,
        max_steps: maxSteps,
        confidence_threshold: confidenceThreshold,
      };

      const response = await advancedAIService.reasoning(request);

      if (response.success && response.data) {
        const completedSession: ReasoningSession = {
          ...newSession,
          response: response.data,
          status: 'completed',
          endTime: new Date(),
        };

        setSessions(prev => prev.map(s => s.id === sessionId ? completedSession : s));
        setCurrentSession(completedSession);
      } else {
        throw new Error(response.error?.message || '推理失败');
      }
    } catch (error) {
      const errorSession: ReasoningSession = {
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

  const handleExampleClick = (example: typeof exampleQueries[0]) => {
    setQuery(example.query);
    setReasoningType(example.type);
  };

  const getReasoningTypeColor = (type: string) => {
    switch (type) {
      case 'deductive': return 'blue';
      case 'inductive': return 'green';
      case 'abductive': return 'orange';
      default: return 'default';
    }
  };

  const getReasoningTypeIcon = (type: string) => {
    switch (type) {
      case 'deductive': return <ThunderboltOutlined />;
      case 'inductive': return <BulbOutlined />;
      case 'abductive': return <BulbOutlined />;
      default: return <InfoCircleOutlined />;
    }
  };

  const renderReasoningSteps = (steps: ReasoningResponse['reasoning_steps']) => {
    return (
      <Steps direction="vertical" size="small">
        {steps.map((step, index) => (
          <Step
            key={index}
            title={`步骤 ${step.step}`}
            description={step.description}
            status={step.confidence > 0.8 ? 'finish' : step.confidence > 0.6 ? 'process' : 'wait'}
            icon={
              step.confidence > 0.8 ? <CheckCircleOutlined /> :
              step.confidence > 0.6 ? <ClockCircleOutlined /> :
              <ExclamationCircleOutlined />
            }
            subTitle={
              <Tag color={step.confidence > 0.8 ? 'green' : step.confidence > 0.6 ? 'orange' : 'red'}>
                置信度: {(step.confidence * 100).toFixed(1)}%
              </Tag>
            }
          />
        ))}
      </Steps>
    );
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '32px', textAlign: 'center' }}>
        <Title level={1} style={{ marginBottom: '8px' }}>
          <BulbOutlined style={{ color: '#1890ff', marginRight: '12px' }} />
          AGI 智能推理
        </Title>
        <Paragraph style={{ fontSize: '16px', color: '#666', maxWidth: '600px', margin: '0 auto' }}>
          基于先进的AGI技术，提供多种推理模式，帮助您分析复杂问题，获得深度洞察
        </Paragraph>
      </div>

      {/* 统计数据 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '32px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总查询次数"
              value={stats.totalQueries}
              prefix={<BulbOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="平均置信度"
              value={stats.avgConfidence}
              precision={2}
              suffix="%"
              prefix={<StarOutlined style={{ color: '#52c41a' }} />}
              formatter={(value) => `${((value as number) * 100).toFixed(1)}%`}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="平均处理时间"
              value={stats.avgProcessingTime}
              precision={1}
              suffix="秒"
              prefix={<ClockCircleOutlined style={{ color: '#fa8c16' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="成功率"
              value={stats.successRate}
              precision={1}
              suffix="%"
              prefix={<CheckCircleOutlined style={{ color: '#f5222d' }} />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]}>
        {/* 推理输入区域 */}
        <Col xs={24} lg={12}>
          <Card title="推理查询" extra={<Badge count="智能" style={{ backgroundColor: '#52c41a' }} />}>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {/* 推理类型选择 */}
              <div>
                <Text strong style={{ marginBottom: '8px', display: 'block' }}>推理类型：</Text>
                <Select
                  value={reasoningType}
                  onChange={setReasoningType}
                  style={{ width: '100%' }}
                  size="large"
                >
                  <Option value="deductive">
                    <Space>
                      <ThunderboltOutlined />
                      演绎推理 - 从一般到特殊
                    </Space>
                  </Option>
                  <Option value="inductive">
                    <Space>
                      <BulbOutlined />
                      归纳推理 - 从特殊到一般
                    </Space>
                  </Option>
                  <Option value="abductive">
                    <Space>
                      <BulbOutlined />
                      溯因推理 - 最佳解释推理
                    </Space>
                  </Option>
                </Select>
              </div>

              {/* 问题输入 */}
              <div>
                <Text strong style={{ marginBottom: '8px', display: 'block' }}>
                  推理问题：
                  <Tooltip title="请详细描述您需要推理分析的问题">
                    <InfoCircleOutlined style={{ marginLeft: '4px', color: '#1890ff' }} />
                  </Tooltip>
                </Text>
                <TextArea
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="请输入您需要推理分析的问题..."
                  rows={4}
                  maxLength={1000}
                  showCount
                />
              </div>

              {/* 上下文信息 */}
              <div>
                <Text strong style={{ marginBottom: '8px', display: 'block' }}>
                  上下文信息（可选）：
                </Text>
                <TextArea
                  value={context}
                  onChange={(e) => setContext(e.target.value)}
                  placeholder="提供相关的背景信息或约束条件..."
                  rows={3}
                  maxLength={500}
                  showCount
                />
              </div>

              {/* 高级设置 */}
              <Row gutter={16}>
                <Col span={12}>
                  <Text strong style={{ marginBottom: '8px', display: 'block' }}>最大推理步数：</Text>
                  <Select
                    value={maxSteps}
                    onChange={setMaxSteps}
                    style={{ width: '100%' }}
                  >
                    <Option value={5}>5步 - 快速</Option>
                    <Option value={10}>10步 - 标准</Option>
                    <Option value={15}>15步 - 深度</Option>
                    <Option value={20}>20步 - 极致</Option>
                  </Select>
                </Col>
                <Col span={12}>
                  <Text strong style={{ marginBottom: '8px', display: 'block' }}>置信度阈值：</Text>
                  <Select
                    value={confidenceThreshold}
                    onChange={setConfidenceThreshold}
                    style={{ width: '100%' }}
                  >
                    <Option value={0.6}>60% - 宽松</Option>
                    <Option value={0.7}>70% - 平衡</Option>
                    <Option value={0.8}>80% - 严格</Option>
                    <Option value={0.9}>90% - 极严</Option>
                  </Select>
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
                disabled={!query.trim()}
              >
                {loading ? '推理中...' : '开始推理'}
              </Button>
            </Space>
          </Card>

          {/* 示例问题 */}
          <Card title="示例问题" style={{ marginTop: '24px' }}>
            <Row gutter={[16, 16]}>
              {exampleQueries.map((example, index) => (
                <Col xs={24} sm={12} key={index}>
                  <Card
                    size="small"
                    hoverable
                    onClick={() => handleExampleClick(example)}
                    style={{ cursor: 'pointer' }}
                  >
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Text strong>{example.title}</Text>
                        <Tag color={getReasoningTypeColor(example.type)} icon={getReasoningTypeIcon(example.type)}>
                          {example.type}
                        </Tag>
                      </div>
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {example.query.substring(0, 60)}...
                      </Text>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>

        {/* 推理结果区域 */}
        <Col xs={24} lg={12}>
          {currentSession ? (
            <Card
              title={
                <Space>
                  <Tag color={getReasoningTypeColor(reasoningType)} icon={getReasoningTypeIcon(reasoningType)}>
                    {reasoningType}
                  </Tag>
                  推理结果
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
                    <Text>AGI正在深度思考中...</Text>
                  </div>
                </div>
              )}

              {currentSession.status === 'error' && (
                <Alert
                  message="推理失败"
                  description={currentSession.error}
                  type="error"
                  showIcon
                />
              )}

              {currentSession.status === 'completed' && currentSession.response && (
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                  {/* 推理结果概览 */}
                  <div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                      <Text strong>推理结果：</Text>
                      <Space>
                        <Tag color="blue">
                          置信度: {(currentSession.response.confidence * 100).toFixed(1)}%
                        </Tag>
                        <Tag color="green">
                          {currentSession.endTime && currentSession.startTime ? 
                            `${((currentSession.endTime.getTime() - currentSession.startTime.getTime()) / 1000).toFixed(1)}秒` :
                            '计算中'
                          }
                        </Tag>
                      </Space>
                    </div>
                    <Card size="small" style={{ backgroundColor: '#f6ffed', border: '1px solid #b7eb8f' }}>
                      <Paragraph>{currentSession.response.solution}</Paragraph>
                    </Card>
                  </div>

                  {/* 推理步骤 */}
                  {currentSession.response.reasoning_steps.length > 0 && (
                    <div>
                      <Text strong style={{ marginBottom: '16px', display: 'block' }}>推理过程：</Text>
                      {renderReasoningSteps(currentSession.response.reasoning_steps)}
                    </div>
                  )}

                  {/* 备选方案 */}
                  {currentSession.response.alternatives && currentSession.response.alternatives.length > 0 && (
                    <div>
                      <Text strong style={{ marginBottom: '16px', display: 'block' }}>备选方案：</Text>
                      <Space direction="vertical" style={{ width: '100%' }}>
                        {currentSession.response.alternatives.map((alt, index) => (
                          <Card key={index} size="small" style={{ backgroundColor: '#fff7e6', border: '1px solid #ffd591' }}>
                            <Text>{alt}</Text>
                          </Card>
                        ))}
                      </Space>
                    </div>
                  )}

                  {/* 置信度进度条 */}
                  <div>
                    <Text strong style={{ marginBottom: '8px', display: 'block' }}>整体置信度：</Text>
                    <Progress
                      percent={currentSession.response.confidence * 100}
                      status={currentSession.response.confidence > 0.8 ? 'success' : currentSession.response.confidence > 0.6 ? 'active' : 'exception'}
                      strokeColor={currentSession.response.confidence > 0.8 ? '#52c41a' : currentSession.response.confidence > 0.6 ? '#1890ff' : '#f5222d'}
                    />
                  </div>
                </Space>
              )}
            </Card>
          ) : (
            <Card title="推理结果" style={{ textAlign: 'center', padding: '40px' }}>
              <BulbOutlined style={{ fontSize: '48px', color: '#d9d9d9', marginBottom: '16px' }} />
              <Text type="secondary">请输入问题开始推理分析</Text>
            </Card>
          )}

          {/* 历史会话 */}
          {sessions.length > 0 && (
            <Card title="推理历史" style={{ marginTop: '24px' }}>
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
                      <Text strong>{session.query.substring(0, 50)}...</Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {session.startTime.toLocaleString()}
                        {session.response && ` · 置信度: ${(session.response.confidence * 100).toFixed(1)}%`}
                      </Text>
                    </div>
                  </Timeline.Item>
                ))}
              </Timeline>
            </Card>
          )}
        </Col>
      </Row>
    </div>
  );
};

export default AGIReasoning;