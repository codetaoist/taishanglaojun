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
  Steps,
  Radio,
  Checkbox,
  Input,
  Slider,
  Rate,
  Modal,
  Form,
  Select,
  DatePicker,
  message,
  Tabs,
  List,
  Avatar,
  Badge,
  Tooltip,
  Divider,
  Empty,
  Spin,
  Alert,
  Timeline,
  Statistic,
  Table,
  Drawer,
  Result,
  Affix,
  BackTop
} from 'antd';
import { 
  TrophyOutlined, 
  StarOutlined, 
  ClockCircleOutlined, 
  CheckCircleOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  EyeOutlined,
  DownloadOutlined,
  ShareAltOutlined,
  BulbOutlined,
  ThunderboltOutlined,
  CrownOutlined,
  FireOutlined,
  RocketOutlined,
  HeartOutlined,
  BookOutlined,
  UserOutlined,
  TeamOutlined,
  LineChartOutlined,
  BarChartOutlined,
  PieChartOutlined,
  RadarChartOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  QuestionCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  SettingOutlined,
  CalendarOutlined,
  TagOutlined,
  FileTextOutlined,
  AimOutlined,
  GiftOutlined
} from '@ant-design/icons';
import { Radar, Column, Line, Pie } from '@ant-design/plots';
import type { ColumnsType } from 'antd/es/table';
import moment from 'moment';

const { Title, Paragraph, Text } = Typography;
const { Step } = Steps;
const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;

interface Question {
  id: string;
  type: 'single' | 'multiple' | 'text' | 'rating' | 'slider';
  category: string;
  difficulty: number; // 1-5
  question: string;
  description?: string;
  options?: string[];
  correctAnswer?: string | string[];
  points: number;
  timeLimit?: number; // 秒
  explanation?: string;
  tags: string[];
}

interface Assessment {
  id: string;
  title: string;
  description: string;
  category: string;
  type: 'knowledge' | 'skill' | 'comprehensive';
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  duration: number; // 分钟
  questionsCount: number;
  totalPoints: number;
  passingScore: number;
  attempts: number;
  maxAttempts: number;
  isPublic: boolean;
  isActive: boolean;
  createdBy: string;
  createdAt: Date;
  updatedAt: Date;
  tags: string[];
  prerequisites: string[];
  learningOutcomes: string[];
  questions: Question[];
  thumbnail: string;
  certificate: boolean;
}

interface AssessmentResult {
  id: string;
  assessmentId: string;
  assessmentTitle: string;
  userId: string;
  score: number;
  totalPoints: number;
  percentage: number;
  passed: boolean;
  timeSpent: number; // 分钟
  startTime: Date;
  endTime: Date;
  answers: {
    questionId: string;
    answer: string | string[];
    isCorrect: boolean;
    points: number;
    timeSpent: number;
  }[];
  categoryScores: {
    category: string;
    score: number;
    totalPoints: number;
    percentage: number;
  }[];
  strengths: string[];
  weaknesses: string[];
  recommendations: string[];
  certificate?: {
    id: string;
    url: string;
    issuedAt: Date;
  };
}

interface AbilityProfile {
  userId: string;
  overallScore: number;
  level: 'beginner' | 'intermediate' | 'advanced' | 'expert';
  categories: {
    category: string;
    score: number;
    level: string;
    assessmentsCount: number;
    lastAssessment: Date;
    trend: 'up' | 'down' | 'stable';
  }[];
  strengths: string[];
  weaknesses: string[];
  recommendations: string[];
  achievements: {
    id: string;
    title: string;
    description: string;
    icon: string;
    earnedAt: Date;
    rarity: 'common' | 'rare' | 'epic' | 'legendary';
  }[];
  totalAssessments: number;
  averageScore: number;
  improvementRate: number;
  lastUpdated: Date;
}

const AbilityAssessment: React.FC = () => {
  const [assessments, setAssessments] = useState<Assessment[]>([]);
  const [results, setResults] = useState<AssessmentResult[]>([]);
  const [abilityProfile, setAbilityProfile] = useState<AbilityProfile | null>(null);
  const [currentAssessment, setCurrentAssessment] = useState<Assessment | null>(null);
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, any>>({});
  const [assessmentStarted, setAssessmentStarted] = useState(false);
  const [assessmentCompleted, setAssessmentCompleted] = useState(false);
  const [currentResult, setCurrentResult] = useState<AssessmentResult | null>(null);
  const [timeRemaining, setTimeRemaining] = useState(0);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('available');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedDifficulty, setSelectedDifficulty] = useState<string>('all');
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [selectedResult, setSelectedResult] = useState<AssessmentResult | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadAssessmentData();
  }, []);

  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (assessmentStarted && timeRemaining > 0) {
      timer = setTimeout(() => {
        setTimeRemaining(timeRemaining - 1);
      }, 1000);
    } else if (assessmentStarted && timeRemaining === 0) {
      handleAssessmentSubmit();
    }
    return () => clearTimeout(timer);
  }, [assessmentStarted, timeRemaining]);

  const loadAssessmentData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockQuestions: Question[] = [
        {
          id: '1',
          type: 'single',
          category: '道家思想',
          difficulty: 2,
          question: '道德经中"道可道，非常道"的含义是什么？',
          options: [
            '可以言说的道不是永恒的道',
            '道路可以行走，但不是普通的道路',
            '道德可以传授，但不是常见的道德',
            '以上都不对'
          ],
          correctAnswer: '可以言说的道不是永恒的道',
          points: 10,
          timeLimit: 60,
          explanation: '这句话表达了道的超越性和不可言喻性，真正的道是超越语言文字的。',
          tags: ['道德经', '老子', '道家哲学']
        },
        {
          id: '2',
          type: 'multiple',
          category: '儒家思想',
          difficulty: 3,
          question: '孔子提出的"仁"的内涵包括哪些方面？',
          options: [
            '爱人',
            '克己复礼',
            '忠恕之道',
            '修身齐家',
            '温良恭俭让'
          ],
          correctAnswer: ['爱人', '克己复礼', '忠恕之道', '温良恭俭让'],
          points: 15,
          timeLimit: 90,
          explanation: '"仁"是儒家思想的核心，包含了爱人、自我修养、待人接物等多个层面。',
          tags: ['论语', '孔子', '仁学']
        },
        {
          id: '3',
          type: 'text',
          category: '佛家思想',
          difficulty: 4,
          question: '请简述"色即是空，空即是色"的哲学含义。',
          points: 20,
          timeLimit: 180,
          explanation: '这是《心经》中的核心思想，表达了现象与本质、有与无的辩证关系。',
          tags: ['心经', '般若', '空性']
        },
        {
          id: '4',
          type: 'rating',
          category: '修行实践',
          difficulty: 2,
          question: '您对冥想练习的熟练程度如何？',
          points: 5,
          tags: ['冥想', '修行', '实践']
        },
        {
          id: '5',
          type: 'slider',
          category: '现代应用',
          difficulty: 3,
          question: '您认为传统文化在现代社会中的重要性程度如何？（1-10分）',
          points: 8,
          tags: ['传统文化', '现代应用', '价值观']
        }
      ];

      const mockAssessments: Assessment[] = [
        {
          id: '1',
          title: '道家思想基础测评',
          description: '测试您对道家思想核心概念的理解程度，包括道德经、庄子等经典著作的基本内容',
          category: '道家思想',
          type: 'knowledge',
          difficulty: 'intermediate',
          duration: 30,
          questionsCount: 20,
          totalPoints: 200,
          passingScore: 120,
          attempts: 0,
          maxAttempts: 3,
          isPublic: true,
          isActive: true,
          createdBy: '系统管理员',
          createdAt: new Date('2024-01-01'),
          updatedAt: new Date('2024-02-01'),
          tags: ['道德经', '庄子', '道家哲学', '基础测评'],
          prerequisites: [],
          learningOutcomes: [
            '掌握道家思想核心概念',
            '理解道德经基本内容',
            '了解庄子哲学思想',
            '具备道家思维方式'
          ],
          questions: mockQuestions,
          thumbnail: '/api/placeholder/300/200',
          certificate: true
        },
        {
          id: '2',
          title: '儒家经典综合评估',
          description: '全面评估您对儒家经典的掌握程度，包括论语、孟子、大学、中庸等重要文献',
          category: '儒家思想',
          type: 'comprehensive',
          difficulty: 'advanced',
          duration: 45,
          questionsCount: 30,
          totalPoints: 300,
          passingScore: 180,
          attempts: 1,
          maxAttempts: 2,
          isPublic: true,
          isActive: true,
          createdBy: '李教授',
          createdAt: new Date('2024-01-15'),
          updatedAt: new Date('2024-02-01'),
          tags: ['论语', '孟子', '四书五经', '综合评估'],
          prerequisites: ['儒家思想基础'],
          learningOutcomes: [
            '深入理解儒家经典',
            '掌握儒家教育思想',
            '具备儒家修身理念',
            '能够运用儒家智慧'
          ],
          questions: mockQuestions,
          thumbnail: '/api/placeholder/300/200',
          certificate: true
        },
        {
          id: '3',
          title: '禅修能力评估',
          description: '评估您的禅修实践能力和对佛家思想的理解深度',
          category: '佛家思想',
          type: 'skill',
          difficulty: 'advanced',
          duration: 25,
          questionsCount: 15,
          totalPoints: 150,
          passingScore: 90,
          attempts: 0,
          maxAttempts: 5,
          isPublic: true,
          isActive: true,
          createdBy: '王法师',
          createdAt: new Date('2024-01-20'),
          updatedAt: new Date('2024-02-01'),
          tags: ['禅修', '冥想', '佛学', '实践能力'],
          prerequisites: ['佛学基础'],
          learningOutcomes: [
            '掌握禅修基本方法',
            '理解佛学核心思想',
            '具备内观能力',
            '培养正念意识'
          ],
          questions: mockQuestions,
          thumbnail: '/api/placeholder/300/200',
          certificate: false
        },
        {
          id: '4',
          title: '传统文化现代应用',
          description: '测试您将传统文化智慧应用于现代生活和工作的能力',
          category: '现代应用',
          type: 'comprehensive',
          duration: 40,
          questionsCount: 25,
          totalPoints: 250,
          passingScore: 150,
          attempts: 0,
          maxAttempts: 3,
          isPublic: true,
          isActive: true,
          createdBy: '陈博士',
          createdAt: new Date('2024-02-01'),
          updatedAt: new Date('2024-02-01'),
          tags: ['现代应用', '文化创新', '实践智慧'],
          prerequisites: ['传统文化基础'],
          learningOutcomes: [
            '掌握文化应用方法',
            '具备创新思维',
            '能够解决实际问题',
            '提升综合素养'
          ],
          questions: mockQuestions,
          thumbnail: '/api/placeholder/300/200',
          certificate: true
        }
      ];

      const mockResults: AssessmentResult[] = [
        {
          id: '1',
          assessmentId: '1',
          assessmentTitle: '道家思想基础测评',
          userId: 'user1',
          score: 165,
          totalPoints: 200,
          percentage: 82.5,
          passed: true,
          timeSpent: 25,
          startTime: new Date('2024-01-25T10:00:00'),
          endTime: new Date('2024-01-25T10:25:00'),
          answers: [],
          categoryScores: [
            { category: '道德经', score: 85, totalPoints: 100, percentage: 85 },
            { category: '庄子', score: 80, totalPoints: 100, percentage: 80 }
          ],
          strengths: ['道德经理解深入', '哲学思辨能力强'],
          weaknesses: ['庄子思想掌握不够'],
          recommendations: [
            '建议深入学习庄子的逍遥游思想',
            '多练习道家思维的实际应用',
            '参与相关讨论和交流'
          ],
          certificate: {
            id: 'cert1',
            url: '/certificates/cert1.pdf',
            issuedAt: new Date('2024-01-25T10:30:00')
          }
        }
      ];

      const mockAbilityProfile: AbilityProfile = {
        userId: 'user1',
        overallScore: 78,
        level: 'intermediate',
        categories: [
          {
            category: '道家思想',
            score: 82,
            level: 'advanced',
            assessmentsCount: 3,
            lastAssessment: new Date('2024-01-25'),
            trend: 'up'
          },
          {
            category: '儒家思想',
            score: 75,
            level: 'intermediate',
            assessmentsCount: 2,
            lastAssessment: new Date('2024-01-20'),
            trend: 'stable'
          },
          {
            category: '佛家思想',
            score: 68,
            level: 'intermediate',
            assessmentsCount: 1,
            lastAssessment: new Date('2024-01-15'),
            trend: 'up'
          },
          {
            category: '现代应用',
            score: 85,
            level: 'advanced',
            assessmentsCount: 2,
            lastAssessment: new Date('2024-01-30'),
            trend: 'up'
          }
        ],
        strengths: ['哲学思辨', '理论理解', '现代应用'],
        weaknesses: ['实践经验', '经典记忆'],
        recommendations: [
          '增加实践练习',
          '加强经典文献背诵',
          '参与更多讨论交流',
          '尝试跨领域应用'
        ],
        achievements: [
          {
            id: '1',
            title: '道家学者',
            description: '在道家思想评估中获得优秀成绩',
            icon: '🏆',
            earnedAt: new Date('2024-01-25'),
            rarity: 'rare'
          },
          {
            id: '2',
            title: '持续学习者',
            description: '连续完成5次评估',
            icon: '📚',
            earnedAt: new Date('2024-01-30'),
            rarity: 'common'
          }
        ],
        totalAssessments: 8,
        averageScore: 78,
        improvementRate: 15,
        lastUpdated: new Date('2024-02-01')
      };

      setAssessments(mockAssessments);
      setResults(mockResults);
      setAbilityProfile(mockAbilityProfile);
      setLoading(false);
    }, 1000);
  };

  // 开始评估
  const startAssessment = (assessment: Assessment) => {
    setCurrentAssessment(assessment);
    setCurrentQuestionIndex(0);
    setAnswers({});
    setAssessmentStarted(true);
    setAssessmentCompleted(false);
    setTimeRemaining(assessment.duration * 60); // 转换为秒
  };

  // 提交答案
  const handleAnswerChange = (questionId: string, answer: any) => {
    setAnswers(prev => ({
      ...prev,
      [questionId]: answer
    }));
  };

  // 下一题
  const nextQuestion = () => {
    if (currentAssessment && currentQuestionIndex < currentAssessment.questions.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1);
    }
  };

  // 上一题
  const prevQuestion = () => {
    if (currentQuestionIndex > 0) {
      setCurrentQuestionIndex(currentQuestionIndex - 1);
    }
  };

  // 提交评估
  const handleAssessmentSubmit = () => {
    if (!currentAssessment) return;

    // 计算分数
    let totalScore = 0;
    const categoryScores: Record<string, { score: number; total: number }> = {};

    currentAssessment.questions.forEach(question => {
      const userAnswer = answers[question.id];
      let isCorrect = false;
      let points = 0;

      if (question.type === 'single' && userAnswer === question.correctAnswer) {
        isCorrect = true;
        points = question.points;
      } else if (question.type === 'multiple' && Array.isArray(userAnswer) && Array.isArray(question.correctAnswer)) {
        const correctSet = new Set(question.correctAnswer);
        const userSet = new Set(userAnswer);
        if (correctSet.size === userSet.size && [...correctSet].every(x => userSet.has(x))) {
          isCorrect = true;
          points = question.points;
        }
      } else if (question.type === 'text' && userAnswer && userAnswer.trim().length > 0) {
        // 文本题简单评分（实际应该用AI评分）
        points = Math.floor(question.points * 0.8);
      } else if (question.type === 'rating' || question.type === 'slider') {
        points = question.points; // 主观题给满分
      }

      totalScore += points;

      // 分类统计
      if (!categoryScores[question.category]) {
        categoryScores[question.category] = { score: 0, total: 0 };
      }
      categoryScores[question.category].score += points;
      categoryScores[question.category].total += question.points;
    });

    const percentage = (totalScore / currentAssessment.totalPoints) * 100;
    const passed = totalScore >= currentAssessment.passingScore;

    const result: AssessmentResult = {
      id: Date.now().toString(),
      assessmentId: currentAssessment.id,
      assessmentTitle: currentAssessment.title,
      userId: 'user1',
      score: totalScore,
      totalPoints: currentAssessment.totalPoints,
      percentage: Math.round(percentage * 10) / 10,
      passed,
      timeSpent: Math.round((currentAssessment.duration * 60 - timeRemaining) / 60),
      startTime: new Date(Date.now() - (currentAssessment.duration * 60 - timeRemaining) * 1000),
      endTime: new Date(),
      answers: [],
      categoryScores: Object.entries(categoryScores).map(([category, scores]) => ({
        category,
        score: scores.score,
        totalPoints: scores.total,
        percentage: Math.round((scores.score / scores.total) * 100 * 10) / 10
      })),
      strengths: [],
      weaknesses: [],
      recommendations: []
    };

    setCurrentResult(result);
    setResults(prev => [result, ...prev]);
    setAssessmentStarted(false);
    setAssessmentCompleted(true);
    
    message.success(passed ? '恭喜您通过了评估！' : '评估完成，继续努力！');
  };

  // 格式化时间
  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  // 获取难度颜色
  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case 'beginner': return 'green';
      case 'intermediate': return 'orange';
      case 'advanced': return 'red';
      default: return 'default';
    }
  };

  // 获取等级颜色
  const getLevelColor = (level: string) => {
    switch (level) {
      case 'beginner': return '#52c41a';
      case 'intermediate': return '#faad14';
      case 'advanced': return '#ff4d4f';
      case 'expert': return '#722ed1';
      default: return '#d9d9d9';
    }
  };

  // 渲染问题
  const renderQuestion = (question: Question) => {
    const userAnswer = answers[question.id];

    switch (question.type) {
      case 'single':
        return (
          <Radio.Group
            value={userAnswer}
            onChange={(e) => handleAnswerChange(question.id, e.target.value)}
          >
            <Space direction="vertical">
              {question.options?.map((option, index) => (
                <Radio key={index} value={option}>
                  {option}
                </Radio>
              ))}
            </Space>
          </Radio.Group>
        );

      case 'multiple':
        return (
          <Checkbox.Group
            value={userAnswer || []}
            onChange={(values) => handleAnswerChange(question.id, values)}
          >
            <Space direction="vertical">
              {question.options?.map((option, index) => (
                <Checkbox key={index} value={option}>
                  {option}
                </Checkbox>
              ))}
            </Space>
          </Checkbox.Group>
        );

      case 'text':
        return (
          <TextArea
            rows={4}
            value={userAnswer || ''}
            onChange={(e) => handleAnswerChange(question.id, e.target.value)}
            placeholder="请输入您的答案..."
          />
        );

      case 'rating':
        return (
          <Rate
            value={userAnswer || 0}
            onChange={(value) => handleAnswerChange(question.id, value)}
          />
        );

      case 'slider':
        return (
          <Slider
            min={1}
            max={10}
            value={userAnswer || 5}
            onChange={(value) => handleAnswerChange(question.id, value)}
            marks={{
              1: '1',
              5: '5',
              10: '10'
            }}
          />
        );

      default:
        return null;
    }
  };

  // 渲染评估卡片
  const renderAssessmentCard = (assessment: Assessment) => (
    <Card
      key={assessment.id}
      hoverable
      style={{ marginBottom: '16px' }}
      cover={
        <div style={{ position: 'relative' }}>
          <img
            alt={assessment.title}
            src={assessment.thumbnail}
            style={{ height: '200px', objectFit: 'cover' }}
          />
          <div style={{ 
            position: 'absolute', 
            top: '8px', 
            left: '8px',
            display: 'flex',
            flexDirection: 'column',
            gap: '4px'
          }}>
            <Tag color={getDifficultyColor(assessment.difficulty)}>
              {assessment.difficulty === 'beginner' ? '初级' :
               assessment.difficulty === 'intermediate' ? '中级' : '高级'}
            </Tag>
            {assessment.certificate && (
              <Tag color="gold">
                <TrophyOutlined /> 证书
              </Tag>
            )}
          </div>
          <div style={{ 
            position: 'absolute', 
            bottom: '8px', 
            right: '8px',
            background: 'rgba(0,0,0,0.7)',
            color: 'white',
            padding: '4px 8px',
            borderRadius: '4px',
            fontSize: '12px'
          }}>
            <ClockCircleOutlined /> {assessment.duration}分钟
          </div>
        </div>
      }
      actions={[
        <Tooltip title="查看详情">
          <EyeOutlined 
            onClick={() => {
              setCurrentAssessment(assessment);
              setDetailDrawerVisible(true);
            }}
          />
        </Tooltip>,
        <Tooltip title="开始评估">
          <PlayCircleOutlined 
            onClick={() => startAssessment(assessment)}
          />
        </Tooltip>,
        <Tooltip title="分享">
          <ShareAltOutlined />
        </Tooltip>
      ]}
    >
      <Card.Meta
        title={
          <div>
            <Text strong>{assessment.title}</Text>
            <div style={{ marginTop: '4px' }}>
              <Space>
                <Tag size="small">{assessment.category}</Tag>
                <Tag size="small" color="blue">
                  {assessment.type === 'knowledge' ? '知识测评' :
                   assessment.type === 'skill' ? '技能测评' : '综合评估'}
                </Tag>
              </Space>
            </div>
          </div>
        }
        description={
          <div>
            <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: '8px' }}>
              {assessment.description}
            </Paragraph>
            
            <Space direction="vertical" style={{ width: '100%' }} size="small">
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Text type="secondary">
                  <BookOutlined /> {assessment.questionsCount} 题
                </Text>
                <Text type="secondary">
                  总分: {assessment.totalPoints}
                </Text>
              </div>
              
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <Text type="secondary">
                  及格分: {assessment.passingScore}
                </Text>
                <Text type="secondary">
                  已尝试: {assessment.attempts}/{assessment.maxAttempts}
                </Text>
              </div>

              <Progress 
                percent={(assessment.passingScore / assessment.totalPoints) * 100}
                size="small"
                showInfo={false}
                strokeColor="#52c41a"
              />

              <Button 
                type="primary" 
                block 
                icon={<PlayCircleOutlined />}
                disabled={assessment.attempts >= assessment.maxAttempts}
                onClick={() => startAssessment(assessment)}
              >
                {assessment.attempts >= assessment.maxAttempts ? '已达最大尝试次数' : '开始评估'}
              </Button>
            </Space>
          </div>
        }
      />
    </Card>
  );

  // 评估表格列定义
  const resultColumns: ColumnsType<AssessmentResult> = [
    {
      title: '评估名称',
      dataIndex: 'assessmentTitle',
      key: 'assessmentTitle',
      render: (title: string, record: AssessmentResult) => (
        <div>
          <Text strong>{title}</Text>
          <br />
          <Text type="secondary" style={{ fontSize: '12px' }}>
            {moment(record.endTime).format('YYYY-MM-DD HH:mm')}
          </Text>
        </div>
      )
    },
    {
      title: '得分',
      key: 'score',
      width: 120,
      render: (_, record: AssessmentResult) => (
        <div style={{ textAlign: 'center' }}>
          <Text strong style={{ fontSize: '16px', color: record.passed ? '#52c41a' : '#ff4d4f' }}>
            {record.score}
          </Text>
          <br />
          <Text type="secondary" style={{ fontSize: '12px' }}>
            / {record.totalPoints}
          </Text>
        </div>
      )
    },
    {
      title: '百分比',
      dataIndex: 'percentage',
      key: 'percentage',
      width: 100,
      render: (percentage: number, record: AssessmentResult) => (
        <div style={{ textAlign: 'center' }}>
          <Text strong style={{ color: record.passed ? '#52c41a' : '#ff4d4f' }}>
            {percentage}%
          </Text>
          <br />
          <Progress 
            percent={percentage} 
            size="small" 
            showInfo={false}
            strokeColor={record.passed ? '#52c41a' : '#ff4d4f'}
          />
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'passed',
      key: 'passed',
      width: 80,
      render: (passed: boolean) => (
        <Tag color={passed ? 'success' : 'error'}>
          {passed ? '通过' : '未通过'}
        </Tag>
      )
    },
    {
      title: '用时',
      dataIndex: 'timeSpent',
      key: 'timeSpent',
      width: 80,
      render: (timeSpent: number) => (
        <Text type="secondary">
          {timeSpent}分钟
        </Text>
      )
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record: AssessmentResult) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />}
              onClick={() => {
                setSelectedResult(record);
                setDetailDrawerVisible(true);
              }}
            />
          </Tooltip>
          {record.certificate && (
            <Tooltip title="下载证书">
              <Button 
                type="text" 
                icon={<DownloadOutlined />}
                onClick={() => window.open(record.certificate!.url)}
              />
            </Tooltip>
          )}
        </Space>
      )
    }
  ];

  // 获取雷达图数据
  const getRadarData = () => {
    if (!abilityProfile) return [];
    
    return abilityProfile.categories.map(category => ({
      category: category.category,
      score: category.score
    }));
  };

  // 获取趋势数据
  const getTrendData = () => {
    return results.slice(0, 10).reverse().map((result, index) => ({
      assessment: index + 1,
      score: result.percentage
    }));
  };

  const filteredAssessments = assessments.filter(assessment => {
    if (selectedCategory !== 'all' && assessment.category !== selectedCategory) return false;
    if (selectedDifficulty !== 'all' && assessment.difficulty !== selectedDifficulty) return false;
    return true;
  });

  const categories = [...new Set(assessments.map(a => a.category))];

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <TrophyOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          能力评估
        </Title>
        <Paragraph>
          智能化能力测评，全面了解您的学习水平和能力发展
        </Paragraph>
      </div>

      {/* 评估进行中 */}
      {assessmentStarted && currentAssessment && (
        <Card style={{ marginBottom: '24px' }}>
          <div style={{ textAlign: 'center', marginBottom: '24px' }}>
            <Title level={3}>{currentAssessment.title}</Title>
            <Space size="large">
              <Text>
                <ClockCircleOutlined /> 剩余时间: {formatTime(timeRemaining)}
              </Text>
              <Text>
                进度: {currentQuestionIndex + 1} / {currentAssessment.questions.length}
              </Text>
            </Space>
          </div>

          <Progress 
            percent={((currentQuestionIndex + 1) / currentAssessment.questions.length) * 100}
            style={{ marginBottom: '24px' }}
          />

          {currentAssessment.questions[currentQuestionIndex] && (
            <div>
              <Card>
                <div style={{ marginBottom: '16px' }}>
                  <Space>
                    <Tag color="blue">
                      第 {currentQuestionIndex + 1} 题
                    </Tag>
                    <Tag>
                      {currentAssessment.questions[currentQuestionIndex].category}
                    </Tag>
                    <Tag color="orange">
                      {currentAssessment.questions[currentQuestionIndex].points} 分
                    </Tag>
                    {currentAssessment.questions[currentQuestionIndex].timeLimit && (
                      <Tag color="red">
                        <ClockCircleOutlined /> {currentAssessment.questions[currentQuestionIndex].timeLimit}秒
                      </Tag>
                    )}
                  </Space>
                </div>

                <Title level={4} style={{ marginBottom: '16px' }}>
                  {currentAssessment.questions[currentQuestionIndex].question}
                </Title>

                {currentAssessment.questions[currentQuestionIndex].description && (
                  <Alert
                    message={currentAssessment.questions[currentQuestionIndex].description}
                    type="info"
                    style={{ marginBottom: '16px' }}
                  />
                )}

                <div style={{ marginBottom: '24px' }}>
                  {renderQuestion(currentAssessment.questions[currentQuestionIndex])}
                </div>

                <div style={{ textAlign: 'center' }}>
                  <Space>
                    <Button 
                      onClick={prevQuestion}
                      disabled={currentQuestionIndex === 0}
                    >
                      上一题
                    </Button>
                    {currentQuestionIndex === currentAssessment.questions.length - 1 ? (
                      <Button 
                        type="primary" 
                        onClick={handleAssessmentSubmit}
                      >
                        提交评估
                      </Button>
                    ) : (
                      <Button 
                        type="primary" 
                        onClick={nextQuestion}
                      >
                        下一题
                      </Button>
                    )}
                  </Space>
                </div>
              </Card>
            </div>
          )}
        </Card>
      )}

      {/* 评估完成结果 */}
      {assessmentCompleted && currentResult && (
        <Card style={{ marginBottom: '24px' }}>
          <Result
            status={currentResult.passed ? 'success' : 'warning'}
            title={currentResult.passed ? '恭喜您通过了评估！' : '评估完成，继续努力！'}
            subTitle={
              <div>
                <Text>
                  您的得分: {currentResult.score} / {currentResult.totalPoints} ({currentResult.percentage}%)
                </Text>
                <br />
                <Text type="secondary">
                  用时: {currentResult.timeSpent} 分钟
                </Text>
              </div>
            }
            extra={[
              <Button 
                type="primary" 
                key="detail"
                onClick={() => {
                  setSelectedResult(currentResult);
                  setDetailDrawerVisible(true);
                }}
              >
                查看详细报告
              </Button>,
              <Button 
                key="retry"
                onClick={() => {
                  setAssessmentCompleted(false);
                  setCurrentResult(null);
                }}
              >
                返回评估列表
              </Button>
            ]}
          />
        </Card>
      )}

      {/* 主要内容 */}
      {!assessmentStarted && !assessmentCompleted && (
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="可用评估" key="available">
            <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
              <Col span={12}>
                <Select
                  value={selectedCategory}
                  onChange={setSelectedCategory}
                  style={{ width: '100%' }}
                  placeholder="选择分类"
                >
                  <Option value="all">全部分类</Option>
                  {categories.map(category => (
                    <Option key={category} value={category}>
                      {category}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col span={12}>
                <Select
                  value={selectedDifficulty}
                  onChange={setSelectedDifficulty}
                  style={{ width: '100%' }}
                  placeholder="选择难度"
                >
                  <Option value="all">全部难度</Option>
                  <Option value="beginner">初级</Option>
                  <Option value="intermediate">中级</Option>
                  <Option value="advanced">高级</Option>
                </Select>
              </Col>
            </Row>

            <Row gutter={[16, 16]}>
              {filteredAssessments.map(assessment => (
                <Col xs={24} sm={12} lg={8} key={assessment.id}>
                  {renderAssessmentCard(assessment)}
                </Col>
              ))}
            </Row>
          </TabPane>

          <TabPane tab="我的成绩" key="results">
            <Card>
              <Table
                columns={resultColumns}
                dataSource={results}
                rowKey="id"
                pagination={{
                  pageSize: 10,
                  showSizeChanger: true,
                  showTotal: (total) => `共 ${total} 条记录`
                }}
              />
            </Card>
          </TabPane>

          <TabPane tab="能力档案" key="profile">
            {abilityProfile && (
              <div>
                {/* 总体能力概览 */}
                <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
                  <Col xs={24} sm={6}>
                    <Card>
                      <Statistic
                        title="总体评分"
                        value={abilityProfile.overallScore}
                        suffix="/ 100"
                        valueStyle={{ color: getLevelColor(abilityProfile.level) }}
                      />
                      <Tag color={getLevelColor(abilityProfile.level)} style={{ marginTop: '8px' }}>
                        {abilityProfile.level === 'beginner' ? '初学者' :
                         abilityProfile.level === 'intermediate' ? '中级' :
                         abilityProfile.level === 'advanced' ? '高级' : '专家'}
                      </Tag>
                    </Card>
                  </Col>
                  <Col xs={24} sm={6}>
                    <Card>
                      <Statistic
                        title="评估次数"
                        value={abilityProfile.totalAssessments}
                        prefix={<BookOutlined />}
                      />
                    </Card>
                  </Col>
                  <Col xs={24} sm={6}>
                    <Card>
                      <Statistic
                        title="平均分数"
                        value={abilityProfile.averageScore}
                        suffix="分"
                        prefix={<StarOutlined />}
                      />
                    </Card>
                  </Col>
                  <Col xs={24} sm={6}>
                    <Card>
                      <Statistic
                        title="提升幅度"
                        value={abilityProfile.improvementRate}
                        suffix="%"
                        prefix={<RocketOutlined />}
                        valueStyle={{ color: '#52c41a' }}
                      />
                    </Card>
                  </Col>
                </Row>

                <Row gutter={[16, 16]}>
                  {/* 能力雷达图 */}
                  <Col xs={24} lg={12}>
                    <Card title="能力雷达图" extra={<RadarChartOutlined />}>
                      <Radar
                        data={getRadarData()}
                        xField="category"
                        yField="score"
                        height={300}
                        area={{}}
                        point={{
                          size: 3
                        }}
                      />
                    </Card>
                  </Col>

                  {/* 成绩趋势 */}
                  <Col xs={24} lg={12}>
                    <Card title="成绩趋势" extra={<LineChartOutlined />}>
                      <Line
                        data={getTrendData()}
                        xField="assessment"
                        yField="score"
                        height={300}
                        smooth
                        point={{
                          size: 3,
                          shape: 'circle'
                        }}
                      />
                    </Card>
                  </Col>
                </Row>

                {/* 分类能力详情 */}
                <Card title="分类能力详情" style={{ marginTop: '16px' }}>
                  <Row gutter={[16, 16]}>
                    {abilityProfile.categories.map(category => (
                      <Col xs={24} sm={12} lg={6} key={category.category}>
                        <Card size="small">
                          <div style={{ textAlign: 'center' }}>
                            <Title level={5}>{category.category}</Title>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: getLevelColor(category.level) }}>
                              {category.score}
                            </div>
                            <Tag color={getLevelColor(category.level)} style={{ marginTop: '8px' }}>
                              {category.level === 'beginner' ? '初级' :
                               category.level === 'intermediate' ? '中级' :
                               category.level === 'advanced' ? '高级' : '专家'}
                            </Tag>
                            <div style={{ marginTop: '8px' }}>
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {category.assessmentsCount} 次评估
                              </Text>
                              <br />
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {moment(category.lastAssessment).format('MM-DD')}
                              </Text>
                            </div>
                          </div>
                        </Card>
                      </Col>
                    ))}
                  </Row>
                </Card>

                {/* 优势与建议 */}
                <Row gutter={[16, 16]} style={{ marginTop: '16px' }}>
                  <Col xs={24} lg={8}>
                    <Card title="优势领域" extra={<StarOutlined style={{ color: '#faad14' }} />}>
                      <List
                        dataSource={abilityProfile.strengths}
                        renderItem={item => (
                          <List.Item>
                            <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
                            {item}
                          </List.Item>
                        )}
                      />
                    </Card>
                  </Col>
                  <Col xs={24} lg={8}>
                    <Card title="待提升" extra={<ExclamationCircleOutlined style={{ color: '#faad14' }} />}>
                      <List
                        dataSource={abilityProfile.weaknesses}
                        renderItem={item => (
                          <List.Item>
                            <InfoCircleOutlined style={{ color: '#faad14', marginRight: '8px' }} />
                            {item}
                          </List.Item>
                        )}
                      />
                    </Card>
                  </Col>
                  <Col xs={24} lg={8}>
                    <Card title="学习建议" extra={<BulbOutlined style={{ color: '#1890ff' }} />}>
                      <List
                        dataSource={abilityProfile.recommendations}
                        renderItem={item => (
                          <List.Item>
                            <BulbOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
                            {item}
                          </List.Item>
                        )}
                      />
                    </Card>
                  </Col>
                </Row>

                {/* 成就徽章 */}
                <Card title="成就徽章" style={{ marginTop: '16px' }}>
                  <Row gutter={[16, 16]}>
                    {abilityProfile.achievements.map(achievement => (
                      <Col xs={12} sm={8} md={6} lg={4} key={achievement.id}>
                        <Card size="small" style={{ textAlign: 'center' }}>
                          <div style={{ fontSize: '32px', marginBottom: '8px' }}>
                            {achievement.icon}
                          </div>
                          <Text strong style={{ fontSize: '12px' }}>
                            {achievement.title}
                          </Text>
                          <br />
                          <Text type="secondary" style={{ fontSize: '10px' }}>
                            {achievement.description}
                          </Text>
                          <br />
                          <Tag 
                            size="small" 
                            color={
                              achievement.rarity === 'common' ? 'default' :
                              achievement.rarity === 'rare' ? 'blue' :
                              achievement.rarity === 'epic' ? 'purple' : 'gold'
                            }
                            style={{ marginTop: '4px' }}
                          >
                            {achievement.rarity === 'common' ? '普通' :
                             achievement.rarity === 'rare' ? '稀有' :
                             achievement.rarity === 'epic' ? '史诗' : '传说'}
                          </Tag>
                        </Card>
                      </Col>
                    ))}
                  </Row>
                </Card>
              </div>
            )}
          </TabPane>
        </Tabs>
      )}

      {/* 详情抽屉 */}
      <Drawer
        title={selectedResult ? "评估报告" : "评估详情"}
        placement="right"
        width={600}
        visible={detailDrawerVisible}
        onClose={() => {
          setDetailDrawerVisible(false);
          setSelectedResult(null);
          setCurrentAssessment(null);
        }}
      >
        {selectedResult ? (
          <div>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <div>
                <Title level={4}>{selectedResult.assessmentTitle}</Title>
                <Text type="secondary">
                  完成时间: {moment(selectedResult.endTime).format('YYYY-MM-DD HH:mm')}
                </Text>
              </div>

              <Card size="small">
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic
                      title="得分"
                      value={selectedResult.score}
                      suffix={`/ ${selectedResult.totalPoints}`}
                      valueStyle={{ color: selectedResult.passed ? '#52c41a' : '#ff4d4f' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="百分比"
                      value={selectedResult.percentage}
                      suffix="%"
                      valueStyle={{ color: selectedResult.passed ? '#52c41a' : '#ff4d4f' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="用时"
                      value={selectedResult.timeSpent}
                      suffix="分钟"
                    />
                  </Col>
                </Row>
              </Card>

              <div>
                <Title level={5}>分类得分</Title>
                <List
                  dataSource={selectedResult.categoryScores}
                  renderItem={category => (
                    <List.Item>
                      <div style={{ width: '100%' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                          <Text strong>{category.category}</Text>
                          <Text>{category.score} / {category.totalPoints}</Text>
                        </div>
                        <Progress percent={category.percentage} size="small" />
                      </div>
                    </List.Item>
                  )}
                />
              </div>

              {selectedResult.strengths.length > 0 && (
                <div>
                  <Title level={5}>优势表现</Title>
                  <List
                    dataSource={selectedResult.strengths}
                    renderItem={item => (
                      <List.Item>
                        <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
                        {item}
                      </List.Item>
                    )}
                  />
                </div>
              )}

              {selectedResult.weaknesses.length > 0 && (
                <div>
                  <Title level={5}>待改进</Title>
                  <List
                    dataSource={selectedResult.weaknesses}
                    renderItem={item => (
                      <List.Item>
                        <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: '8px' }} />
                        {item}
                      </List.Item>
                    )}
                  />
                </div>
              )}

              {selectedResult.recommendations.length > 0 && (
                <div>
                  <Title level={5}>学习建议</Title>
                  <List
                    dataSource={selectedResult.recommendations}
                    renderItem={item => (
                      <List.Item>
                        <BulbOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
                        {item}
                      </List.Item>
                    )}
                  />
                </div>
              )}

              {selectedResult.certificate && (
                <div>
                  <Title level={5}>证书</Title>
                  <Button 
                    type="primary" 
                    icon={<DownloadOutlined />}
                    onClick={() => window.open(selectedResult.certificate!.url)}
                  >
                    下载证书
                  </Button>
                </div>
              )}
            </Space>
          </div>
        ) : currentAssessment ? (
          <div>
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <div>
                <Title level={4}>{currentAssessment.title}</Title>
                <Text type="secondary">{currentAssessment.description}</Text>
              </div>

              <div>
                <Space wrap>
                  <Tag color={getDifficultyColor(currentAssessment.difficulty)}>
                    {currentAssessment.difficulty === 'beginner' ? '初级' :
                     currentAssessment.difficulty === 'intermediate' ? '中级' : '高级'}
                  </Tag>
                  <Tag>{currentAssessment.category}</Tag>
                  <Tag color="blue">
                    {currentAssessment.type === 'knowledge' ? '知识测评' :
                     currentAssessment.type === 'skill' ? '技能测评' : '综合评估'}
                  </Tag>
                  {currentAssessment.certificate && (
                    <Tag color="gold">
                      <TrophyOutlined /> 可获得证书
                    </Tag>
                  )}
                </Space>
              </div>

              <Card size="small">
                <Row gutter={16}>
                  <Col span={8}>
                    <Statistic
                      title="题目数量"
                      value={currentAssessment.questionsCount}
                      suffix="题"
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="考试时长"
                      value={currentAssessment.duration}
                      suffix="分钟"
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="总分"
                      value={currentAssessment.totalPoints}
                      suffix="分"
                    />
                  </Col>
                </Row>
              </Card>

              <div>
                <Text strong>及格分数: </Text>
                <Text>{currentAssessment.passingScore} 分</Text>
                <Progress 
                  percent={(currentAssessment.passingScore / currentAssessment.totalPoints) * 100}
                  size="small"
                  style={{ marginTop: '8px' }}
                />
              </div>

              <div>
                <Text strong>尝试次数: </Text>
                <Text>{currentAssessment.attempts} / {currentAssessment.maxAttempts}</Text>
              </div>

              {currentAssessment.prerequisites.length > 0 && (
                <div>
                  <Title level={5}>前置要求</Title>
                  <List
                    dataSource={currentAssessment.prerequisites}
                    renderItem={item => (
                      <List.Item>
                        <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: '8px' }} />
                        {item}
                      </List.Item>
                    )}
                  />
                </div>
              )}

              <div>
                <Title level={5}>学习成果</Title>
                <List
                  dataSource={currentAssessment.learningOutcomes}
                  renderItem={item => (
                    <List.Item>
                      <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
                      {item}
                    </List.Item>
                  )}
                />
              </div>

              <div>
                <Title level={5}>标签</Title>
                <Space wrap>
                  {currentAssessment.tags.map((tag, index) => (
                    <Tag key={index}>{tag}</Tag>
                  ))}
                </Space>
              </div>

              <Button 
                type="primary" 
                block 
                size="large"
                icon={<PlayCircleOutlined />}
                disabled={currentAssessment.attempts >= currentAssessment.maxAttempts}
                onClick={() => {
                  setDetailDrawerVisible(false);
                  startAssessment(currentAssessment);
                }}
              >
                {currentAssessment.attempts >= currentAssessment.maxAttempts ? '已达最大尝试次数' : '开始评估'}
              </Button>
            </Space>
          </div>
        ) : null}
      </Drawer>

      <BackTop />
    </div>
  );
};

export default AbilityAssessment;