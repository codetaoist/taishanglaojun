import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Progress,
  Typography,
  Steps,
  Radio,
  Checkbox,
  Input,
  Form,
  Modal,
  Result,
  Tabs,
  Tag,
  Space,
  Statistic,
  Timeline,
  Alert,
  Divider,
  List,
  Avatar,
  Badge,
  Tooltip,
  Empty,
} from 'antd';
import {
  TrophyOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  StarOutlined,
  FireOutlined,
  ThunderboltOutlined,
  BulbOutlined,
  AimOutlined,
  RocketOutlined,
  BookOutlined,
  LineChartOutlined,
  BarChartOutlined,
  PieChartOutlined,
} from '@ant-design/icons';
import { Radar, Column, Pie } from '@ant-design/plots';

const { Title, Text, Paragraph } = Typography;
const { Step } = Steps;
const { TabPane } = Tabs;
const { TextArea } = Input;

// 技能领域定义
const skillDomains = [
  { key: 'frontend', name: '前端开发', icon: '🎨', color: '#1890ff' },
  { key: 'backend', name: '后端开发', icon: '⚙️', color: '#52c41a' },
  { key: 'database', name: '数据库', icon: '🗄️', color: '#722ed1' },
  { key: 'devops', name: 'DevOps', icon: '🚀', color: '#fa8c16' },
  { key: 'mobile', name: '移动开发', icon: '📱', color: '#eb2f96' },
  { key: 'ai', name: '人工智能', icon: '🤖', color: '#13c2c2' },
  { key: 'design', name: '设计', icon: '🎭', color: '#f5222d' },
  { key: 'project', name: '项目管理', icon: '📊', color: '#a0d911' },
];

// 生成模拟测试题目
const generateQuestions = (domain: string) => {
  const questionTypes = ['single', 'multiple', 'text'];
  const difficulties = ['easy', 'medium', 'hard'];
  
  return Array.from({ length: 20 }, (_, index) => ({
    id: index + 1,
    type: questionTypes[index % questionTypes.length],
    difficulty: difficulties[index % difficulties.length],
    domain,
    question: `${domain}相关问题 ${index + 1}：这是一个${difficulties[index % difficulties.length]}难度的问题，请选择正确答案或填写您的理解。`,
    options: questionTypes[index % questionTypes.length] !== 'text' ? [
      '选项A：这是第一个选项',
      '选项B：这是第二个选项',
      '选项C：这是第三个选项',
      '选项D：这是第四个选项',
    ] : undefined,
    correctAnswer: questionTypes[index % questionTypes.length] === 'single' ? 0 : 
                   questionTypes[index % questionTypes.length] === 'multiple' ? [0, 2] : undefined,
    points: difficulties[index % difficulties.length] === 'easy' ? 1 : 
            difficulties[index % difficulties.length] === 'medium' ? 2 : 3,
  }));
};

// 生成评估报告数据
const generateAssessmentReport = (answers: any[], questions: any[]) => {
  const totalPoints = questions.reduce((sum, q) => sum + q.points, 0);
  const earnedPoints = Math.floor(totalPoints * (0.6 + Math.random() * 0.3));
  const percentage = Math.round((earnedPoints / totalPoints) * 100);
  
  const skillLevels = skillDomains.map(domain => ({
    skill: domain.name,
    level: 60 + Math.random() * 35,
    color: domain.color,
  }));

  const strengths = [
    '逻辑思维能力强',
    '学习能力突出',
    '实践经验丰富',
    '团队协作良好',
  ].slice(0, Math.floor(Math.random() * 2) + 2);

  const improvements = [
    '需要加强理论基础',
    '可以提升编程效率',
    '建议多做项目实践',
    '需要关注新技术趋势',
  ].slice(0, Math.floor(Math.random() * 2) + 1);

  const recommendations = [
    { type: 'course', title: '推荐课程：高级前端开发', reason: '提升核心技能' },
    { type: 'practice', title: '实战项目：电商系统开发', reason: '增强实践经验' },
    { type: 'reading', title: '推荐书籍：《代码整洁之道》', reason: '提升代码质量' },
  ];

  return {
    score: percentage,
    earnedPoints,
    totalPoints,
    level: percentage >= 90 ? '专家' : percentage >= 75 ? '高级' : percentage >= 60 ? '中级' : '初级',
    skillLevels,
    strengths,
    improvements,
    recommendations,
    completedAt: new Date().toLocaleString(),
  };
};

const AbilityAssessment: React.FC = () => {
  const [currentStep, setCurrentStep] = useState(0);
  const [selectedDomain, setSelectedDomain] = useState<string>('');
  const [questions, setQuestions] = useState<any[]>([]);
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [answers, setAnswers] = useState<any[]>([]);
  const [timeLeft, setTimeLeft] = useState(1800); // 30分钟
  const [isTestStarted, setIsTestStarted] = useState(false);
  const [isTestCompleted, setIsTestCompleted] = useState(false);
  const [assessmentReport, setAssessmentReport] = useState<any>(null);
  const [form] = Form.useForm();

  // 历史评估记录
  const [assessmentHistory] = useState([
    {
      id: 1,
      domain: '前端开发',
      score: 85,
      level: '高级',
      date: '2024-01-15',
      duration: '25分钟',
    },
    {
      id: 2,
      domain: '后端开发',
      score: 72,
      level: '中级',
      date: '2024-01-10',
      duration: '28分钟',
    },
    {
      id: 3,
      domain: '数据库',
      score: 68,
      level: '中级',
      date: '2024-01-05',
      duration: '22分钟',
    },
  ]);

  // 计时器
  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (isTestStarted && !isTestCompleted && timeLeft > 0) {
      timer = setTimeout(() => {
        setTimeLeft(timeLeft - 1);
      }, 1000);
    } else if (timeLeft === 0) {
      handleTestComplete();
    }
    return () => clearTimeout(timer);
  }, [isTestStarted, isTestCompleted, timeLeft]);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const handleDomainSelect = (domain: string) => {
    setSelectedDomain(domain);
    setCurrentStep(1);
  };

  const handleTestStart = () => {
    const testQuestions = generateQuestions(selectedDomain);
    setQuestions(testQuestions);
    setAnswers(new Array(testQuestions.length).fill(null));
    setIsTestStarted(true);
    setCurrentStep(2);
  };

  const handleAnswerChange = (value: any) => {
    const newAnswers = [...answers];
    newAnswers[currentQuestion] = value;
    setAnswers(newAnswers);
  };

  const handleNextQuestion = () => {
    if (currentQuestion < questions.length - 1) {
      setCurrentQuestion(currentQuestion + 1);
    } else {
      handleTestComplete();
    }
  };

  const handlePrevQuestion = () => {
    if (currentQuestion > 0) {
      setCurrentQuestion(currentQuestion - 1);
    }
  };

  const handleTestComplete = () => {
    setIsTestCompleted(true);
    const report = generateAssessmentReport(answers, questions);
    setAssessmentReport(report);
    setCurrentStep(3);
  };

  const resetTest = () => {
    setCurrentStep(0);
    setSelectedDomain('');
    setQuestions([]);
    setCurrentQuestion(0);
    setAnswers([]);
    setTimeLeft(1800);
    setIsTestStarted(false);
    setIsTestCompleted(false);
    setAssessmentReport(null);
    form.resetFields();
  };

  const renderDomainSelection = () => (
    <div>
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <Title level={2}>选择评估领域</Title>
        <Text type="secondary">请选择您想要评估的技能领域</Text>
      </div>
      <Row gutter={[16, 16]}>
        {skillDomains.map(domain => (
          <Col key={domain.key} xs={24} sm={12} md={8} lg={6}>
            <Card
              hoverable
              style={{ 
                textAlign: 'center',
                border: selectedDomain === domain.key ? `2px solid ${domain.color}` : undefined,
              }}
              onClick={() => handleDomainSelect(domain.key)}
            >
              <div style={{ fontSize: 48, marginBottom: 16 }}>
                {domain.icon}
              </div>
              <Title level={4} style={{ color: domain.color, margin: 0 }}>
                {domain.name}
              </Title>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );

  const renderTestPreparation = () => {
    const selectedDomainInfo = skillDomains.find(d => d.key === selectedDomain);
    return (
      <div style={{ maxWidth: 600, margin: '0 auto' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{ fontSize: 64, marginBottom: 16 }}>
            {selectedDomainInfo?.icon}
          </div>
          <Title level={2}>{selectedDomainInfo?.name}能力评估</Title>
          <Text type="secondary">准备开始您的技能评估测试</Text>
        </div>

        <Card>
          <Title level={4}>测试说明</Title>
          <List
            size="small"
            dataSource={[
              '测试时间：30分钟',
              '题目数量：20道题',
              '题型包括：单选题、多选题、简答题',
              '难度分布：简单、中等、困难',
              '评分标准：根据答题正确率和难度系数计算',
              '建议在安静的环境中完成测试',
            ]}
            renderItem={item => (
              <List.Item>
                <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                {item}
              </List.Item>
            )}
          />
          
          <Divider />
          
          <div style={{ textAlign: 'center' }}>
            <Space size="large">
              <Button onClick={() => setCurrentStep(0)}>
                重新选择领域
              </Button>
              <Button type="primary" size="large" onClick={handleTestStart}>
                开始测试
              </Button>
            </Space>
          </div>
        </Card>
      </div>
    );
  };

  const renderTestInterface = () => {
    const question = questions[currentQuestion];
    if (!question) return null;

    return (
      <div>
        {/* 测试头部 */}
        <Card style={{ marginBottom: 16 }}>
          <Row justify="space-between" align="middle">
            <Col>
              <Space>
                <Text strong>题目 {currentQuestion + 1} / {questions.length}</Text>
                <Tag color={question.difficulty === 'easy' ? 'green' : 
                           question.difficulty === 'medium' ? 'orange' : 'red'}>
                  {question.difficulty === 'easy' ? '简单' : 
                   question.difficulty === 'medium' ? '中等' : '困难'}
                </Tag>
                <Text type="secondary">分值: {question.points}</Text>
              </Space>
            </Col>
            <Col>
              <Space>
                <ClockCircleOutlined />
                <Text strong style={{ color: timeLeft < 300 ? '#ff4d4f' : undefined }}>
                  {formatTime(timeLeft)}
                </Text>
              </Space>
            </Col>
          </Row>
          <Progress 
            percent={Math.round(((currentQuestion + 1) / questions.length) * 100)} 
            style={{ marginTop: 16 }}
          />
        </Card>

        {/* 题目内容 */}
        <Card>
          <Title level={4}>{question.question}</Title>
          
          <Form form={form} layout="vertical">
            {question.type === 'single' && (
              <Form.Item>
                <Radio.Group
                  value={answers[currentQuestion]}
                  onChange={(e) => handleAnswerChange(e.target.value)}
                >
                  <Space direction="vertical" style={{ width: '100%' }}>
                    {question.options?.map((option: string, index: number) => (
                      <Radio key={index} value={index} style={{ padding: '8px 0' }}>
                        {option}
                      </Radio>
                    ))}
                  </Space>
                </Radio.Group>
              </Form.Item>
            )}

            {question.type === 'multiple' && (
              <Form.Item>
                <Checkbox.Group
                  value={answers[currentQuestion] || []}
                  onChange={handleAnswerChange}
                >
                  <Space direction="vertical" style={{ width: '100%' }}>
                    {question.options?.map((option: string, index: number) => (
                      <Checkbox key={index} value={index} style={{ padding: '8px 0' }}>
                        {option}
                      </Checkbox>
                    ))}
                  </Space>
                </Checkbox.Group>
              </Form.Item>
            )}

            {question.type === 'text' && (
              <Form.Item>
                <TextArea
                  rows={4}
                  placeholder="请输入您的答案..."
                  value={answers[currentQuestion] || ''}
                  onChange={(e) => handleAnswerChange(e.target.value)}
                />
              </Form.Item>
            )}
          </Form>

          <Divider />

          <Row justify="space-between">
            <Col>
              <Button 
                disabled={currentQuestion === 0}
                onClick={handlePrevQuestion}
              >
                上一题
              </Button>
            </Col>
            <Col>
              <Space>
                <Button onClick={handleTestComplete}>
                  提交测试
                </Button>
                <Button 
                  type="primary"
                  onClick={handleNextQuestion}
                >
                  {currentQuestion === questions.length - 1 ? '完成测试' : '下一题'}
                </Button>
              </Space>
            </Col>
          </Row>
        </Card>
      </div>
    );
  };

  const renderAssessmentResult = () => {
    if (!assessmentReport) return null;

    const radarData = assessmentReport.skillLevels.map((skill: any) => ({
      skill: skill.skill,
      value: skill.level,
    }));

    const radarConfig = {
      data: radarData,
      xField: 'skill',
      yField: 'value',
      area: {
        visible: true,
        style: {
          fillOpacity: 0.3,
        },
      },
      point: {
        visible: true,
        size: 4,
      },
      meta: {
        value: {
          alias: '能力值',
          min: 0,
          max: 100,
        },
      },
    };

    const columnData = assessmentReport.skillLevels.map((skill: any) => ({
      skill: skill.skill,
      value: skill.level,
    }));

    const columnConfig = {
      data: columnData,
      xField: 'skill',
      yField: 'value',
      color: '#1890ff',
      meta: {
        value: {
          alias: '能力值',
        },
      },
    };

    return (
      <div>
        <Result
          status="success"
          title="评估完成！"
          subTitle={`您的${skillDomains.find(d => d.key === selectedDomain)?.name}能力评估已完成`}
          extra={[
            <Button type="primary" key="view" onClick={() => {}}>
              查看详细报告
            </Button>,
            <Button key="retry" onClick={resetTest}>
              重新测试
            </Button>,
          ]}
        />

        <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
          <Col xs={24} sm={6}>
            <Card>
              <Statistic
                title="总分"
                value={assessmentReport.score}
                suffix="%"
                valueStyle={{ color: '#1890ff' }}
                prefix={<TrophyOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card>
              <Statistic
                title="能力等级"
                value={assessmentReport.level}
                valueStyle={{ color: '#52c41a' }}
                prefix={<StarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card>
              <Statistic
                title="得分"
                value={assessmentReport.earnedPoints}
                suffix={`/ ${assessmentReport.totalPoints}`}
                valueStyle={{ color: '#722ed1' }}
                prefix={<AimOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={6}>
            <Card>
              <Statistic
                title="完成时间"
                value={formatTime(1800 - timeLeft)}
                valueStyle={{ color: '#fa8c16' }}
                prefix={<ClockCircleOutlined />}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
          <Col xs={24} lg={12}>
            <Card title="能力雷达图">
              <Radar {...radarConfig} />
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="技能分布">
              <Column {...columnConfig} />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
          <Col xs={24} md={12}>
            <Card title="优势能力" extra={<FireOutlined style={{ color: '#52c41a' }} />}>
              <List
                size="small"
                dataSource={assessmentReport.strengths}
                renderItem={(item: string) => (
                  <List.Item>
                    <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                    {item}
                  </List.Item>
                )}
              />
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card title="改进建议" extra={<BulbOutlined style={{ color: '#fa8c16' }} />}>
              <List
                size="small"
                dataSource={assessmentReport.improvements}
                renderItem={(item: string) => (
                  <List.Item>
                    <ExclamationCircleOutlined style={{ color: '#fa8c16', marginRight: 8 }} />
                    {item}
                  </List.Item>
                )}
              />
            </Card>
          </Col>
        </Row>

        <Card title="学习推荐" style={{ marginTop: 16 }}>
          <List
            dataSource={assessmentReport.recommendations}
            renderItem={(item: any) => (
              <List.Item
                actions={[
                  <Button type="link" key="view">查看详情</Button>,
                  <Button type="link" key="start">开始学习</Button>,
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <Avatar 
                      icon={item.type === 'course' ? <BookOutlined /> : 
                            item.type === 'practice' ? <RocketOutlined /> : <BookOutlined />}
                      style={{ 
                        backgroundColor: item.type === 'course' ? '#1890ff' : 
                                         item.type === 'practice' ? '#52c41a' : '#722ed1'
                      }}
                    />
                  }
                  title={item.title}
                  description={item.reason}
                />
              </List.Item>
            )}
          />
        </Card>
      </div>
    );
  };

  const renderHistoryTab = () => (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="总评估次数"
              value={assessmentHistory.length}
              prefix={<BarChartOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="平均分数"
              value={Math.round(assessmentHistory.reduce((sum, item) => sum + item.score, 0) / assessmentHistory.length)}
              suffix="%"
              prefix={<LineChartOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="最高分数"
              value={Math.max(...assessmentHistory.map(item => item.score))}
              suffix="%"
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="评估历史">
        <Timeline>
          {assessmentHistory.map(item => (
            <Timeline.Item
              key={item.id}
              color={item.score >= 80 ? 'green' : item.score >= 60 ? 'blue' : 'red'}
            >
              <div>
                <Text strong>{item.domain}</Text>
                <Tag color={item.score >= 80 ? 'green' : item.score >= 60 ? 'blue' : 'red'} style={{ marginLeft: 8 }}>
                  {item.level}
                </Tag>
                <div style={{ marginTop: 4 }}>
                  <Space>
                    <Text type="secondary">分数: {item.score}%</Text>
                    <Text type="secondary">用时: {item.duration}</Text>
                    <Text type="secondary">{item.date}</Text>
                  </Space>
                </div>
              </div>
            </Timeline.Item>
          ))}
        </Timeline>
      </Card>
    </div>
  );

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>
          <TrophyOutlined style={{ marginRight: 8 }} />
          能力评估
        </Title>
        <Text type="secondary">全面评估您的技能水平，获得个性化学习建议</Text>
      </div>

      <Tabs defaultActiveKey="assessment">
        <TabPane tab="开始评估" key="assessment">
          <Steps current={currentStep} style={{ marginBottom: 32 }}>
            <Step title="选择领域" description="选择评估的技能领域" />
            <Step title="测试准备" description="了解测试规则和要求" />
            <Step title="进行测试" description="完成能力评估测试" />
            <Step title="查看结果" description="获得评估报告和建议" />
          </Steps>

          {currentStep === 0 && renderDomainSelection()}
          {currentStep === 1 && renderTestPreparation()}
          {currentStep === 2 && renderTestInterface()}
          {currentStep === 3 && renderAssessmentResult()}
        </TabPane>
        
        <TabPane tab="评估历史" key="history">
          {renderHistoryTab()}
        </TabPane>
      </Tabs>
    </div>
  );
};

export default AbilityAssessment;