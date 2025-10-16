import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Typography, 
  Space, 
  Select, 
  DatePicker, 
  Progress, 
  Tag, 
  Alert,
  Tabs,
  List,
  Statistic,
  Tooltip,
  Badge,
  Timeline,
  Table
} from 'antd';
import { 
  BarChartOutlined, 
  RiseOutlined, 
  FallOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined,
  BulbOutlined,
  ExportOutlined,
  ReloadOutlined,
  EyeOutlined,
  HeartOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined
} from '@ant-design/icons';
import { Line, Column, Pie, Area } from '@ant-design/plots';

const { Title, Paragraph, Text } = Typography;
const { RangePicker } = DatePicker;

interface AnalysisReport {
  id: string;
  title: string;
  type: 'trend' | 'risk' | 'improvement' | 'alert';
  severity: 'low' | 'medium' | 'high';
  description: string;
  recommendations: string[];
  metrics: string[];
  timestamp: Date;
  confidence: number;
}

interface HealthTrend {
  metric: string;
  trend: 'improving' | 'stable' | 'declining';
  change: number;
  period: string;
  significance: 'high' | 'medium' | 'low';
}

interface RiskFactor {
  id: string;
  name: string;
  level: 'low' | 'medium' | 'high';
  probability: number;
  impact: string;
  prevention: string[];
}

const HealthAnalysis: React.FC = () => {
  const [analysisReports, setAnalysisReports] = useState<AnalysisReport[]>([]);
  const [healthTrends, setHealthTrends] = useState<HealthTrend[]>([]);
  const [riskFactors, setRiskFactors] = useState<RiskFactor[]>([]);
  const [selectedPeriod, setSelectedPeriod] = useState<string>('30d');
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadAnalysisData();
  }, [selectedPeriod]);

  const loadAnalysisData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      setAnalysisReports([
        {
          id: '1',
          title: '心血管健康趋势分析',
          type: 'trend',
          severity: 'medium',
          description: '最近30天心率和血压数据显示轻微上升趋势，建议关注心血管健康',
          recommendations: [
            '增加有氧运动，每周至少150分钟',
            '减少钠盐摄入，控制在每日6克以下',
            '保持规律作息，避免熬夜'
          ],
          metrics: ['心率', '血压', '运动量'],
          timestamp: new Date(),
          confidence: 85
        },
        {
          id: '2',
          title: '睡眠质量改善报告',
          type: 'improvement',
          severity: 'low',
          description: '睡眠质量在过去两周有显著改善，深度睡眠时间增加25%',
          recommendations: [
            '继续保持当前的睡眠习惯',
            '睡前1小时避免使用电子设备',
            '保持卧室温度在18-22°C'
          ],
          metrics: ['睡眠时长', '深度睡眠', '睡眠效率'],
          timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000),
          confidence: 92
        },
        {
          id: '3',
          title: '体重管理风险提醒',
          type: 'alert',
          severity: 'high',
          description: '体重在过去一个月持续上升，已超出健康范围，需要立即采取行动',
          recommendations: [
            '制定科学的减重计划',
            '控制每日热量摄入',
            '增加运动频率和强度',
            '考虑咨询营养师'
          ],
          metrics: ['体重', 'BMI', '体脂率'],
          timestamp: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          confidence: 78
        }
      ]);

      setHealthTrends([
        { metric: '心率', trend: 'stable', change: 2, period: '30天', significance: 'low' },
        { metric: '血压', trend: 'improving', change: -5, period: '30天', significance: 'medium' },
        { metric: '体重', trend: 'declining', change: 3.2, period: '30天', significance: 'high' },
        { metric: '睡眠质量', trend: 'improving', change: 15, period: '30天', significance: 'high' },
        { metric: '运动量', trend: 'stable', change: -2, period: '30天', significance: 'low' },
        { metric: '血糖', trend: 'stable', change: 0.1, period: '30天', significance: 'low' }
      ]);

      setRiskFactors([
        {
          id: '1',
          name: '心血管疾病',
          level: 'medium',
          probability: 25,
          impact: '可能导致心脏病、中风等严重后果',
          prevention: ['控制血压', '规律运动', '健康饮食', '戒烟限酒']
        },
        {
          id: '2',
          name: '糖尿病',
          level: 'low',
          probability: 12,
          impact: '可能导致血糖控制困难，影响多个器官',
          prevention: ['控制体重', '均衡饮食', '定期检查', '适量运动']
        },
        {
          id: '3',
          name: '肥胖症',
          level: 'high',
          probability: 45,
          impact: '增加多种慢性疾病风险',
          prevention: ['控制饮食', '增加运动', '改善生活方式', '定期监测']
        }
      ]);

      setLoading(false);
    }, 1000);
  };

  // 获取趋势颜色
  const getTrendColor = (trend: string) => {
    switch (trend) {
      case 'improving': return '#52c41a';
      case 'stable': return '#1890ff';
      case 'declining': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取趋势图标
  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'improving': return <RiseOutlined style={{ color: '#52c41a' }} />;
      case 'stable': return <RiseOutlined style={{ color: '#1890ff', transform: 'rotate(90deg)' }} />;
      case 'declining': return <FallOutlined style={{ color: '#ff4d4f' }} />;
      default: return null;
    }
  };

  // 获取严重程度颜色
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'low': return '#52c41a';
      case 'medium': return '#faad14';
      case 'high': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取风险等级颜色
  const getRiskColor = (level: string) => {
    switch (level) {
      case 'low': return '#52c41a';
      case 'medium': return '#faad14';
      case 'high': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 渲染分析报告卡片
  const renderAnalysisReport = (report: AnalysisReport) => {
    const getTypeIcon = () => {
      switch (report.type) {
        case 'trend': return <BarChartOutlined />;
        case 'risk': return <ExclamationCircleOutlined />;
        case 'improvement': return <CheckCircleOutlined />;
        case 'alert': return <WarningOutlined />;
        default: return <InfoCircleOutlined />;
      }
    };

    return (
      <Card
        key={report.id}
        style={{ marginBottom: '16px' }}
        title={
          <Space>
            {getTypeIcon()}
            <Text strong>{report.title}</Text>
            <Tag color={getSeverityColor(report.severity)}>
              {report.severity === 'low' ? '低' : report.severity === 'medium' ? '中' : '高'}
            </Tag>
          </Space>
        }
        extra={
          <Space>
            <Text type="secondary">置信度: {report.confidence}%</Text>
            <Progress 
              type="circle" 
              percent={report.confidence} 
              width={30} 
              strokeColor={getSeverityColor(report.severity)}
            />
          </Space>
        }
      >
        <Paragraph>{report.description}</Paragraph>
        
        <div style={{ marginBottom: '16px' }}>
          <Text strong>相关指标: </Text>
          <Space wrap>
            {report.metrics.map((metric, index) => (
              <Tag key={index} color="blue">{metric}</Tag>
            ))}
          </Space>
        </div>

        <div style={{ marginBottom: '16px' }}>
          <Text strong>建议措施:</Text>
          <List
            size="small"
            dataSource={report.recommendations}
            renderItem={(item, index) => (
              <List.Item style={{ padding: '4px 0' }}>
                <Text>• {item}</Text>
              </List.Item>
            )}
          />
        </div>

        <div style={{ textAlign: 'right' }}>
          <Text type="secondary" style={{ fontSize: '12px' }}>
            {report.timestamp.toLocaleString()}
          </Text>
        </div>
      </Card>
    );
  };

  // 渲染健康趋势图表
  const renderTrendChart = () => {
    const data = healthTrends.map(trend => ({
      metric: trend.metric,
      change: Math.abs(trend.change),
      trend: trend.trend,
      significance: trend.significance
    }));

    const config = {
      data,
      xField: 'metric',
      yField: 'change',
      seriesField: 'trend',
      color: ['#52c41a', '#1890ff', '#ff4d4f'],
      columnStyle: {
        radius: [4, 4, 0, 0],
      },
      legend: {
        position: 'top-left' as const,
      },
    };

    return <Column {...config} />;
  };

  // 渲染风险分布饼图
  const renderRiskChart = () => {
    const data = riskFactors.map(risk => ({
      type: risk.name,
      value: risk.probability,
      level: risk.level
    }));

    const config = {
      data,
      angleField: 'value',
      colorField: 'type',
      radius: 0.8,
      label: {
        type: 'outer',
        content: '{name} {percentage}',
      },
      interactions: [
        {
          type: 'element-active',
        },
      ],
    };

    return <Pie {...config} />;
  };

  // 渲染健康评分
  const renderHealthScore = () => {
    const overallScore = 78; // 模拟总体健康评分
    const scores = [
      { name: '心血管', score: 85, color: '#52c41a' },
      { name: '代谢', score: 72, color: '#faad14' },
      { name: '睡眠', score: 88, color: '#1890ff' },
      { name: '运动', score: 65, color: '#ff7a45' },
      { name: '营养', score: 75, color: '#722ed1' }
    ];

    return (
      <Card title="健康评分" extra={<Button icon={<EyeOutlined />} size="small">详细报告</Button>}>
        <div style={{ textAlign: 'center', marginBottom: '24px' }}>
          <Progress
            type="circle"
            percent={overallScore}
            width={120}
            strokeColor={{
              '0%': '#ff4d4f',
              '50%': '#faad14',
              '100%': '#52c41a',
            }}
            format={(percent) => (
              <div>
                <div style={{ fontSize: '24px', fontWeight: 'bold' }}>{percent}</div>
                <div style={{ fontSize: '12px', color: '#666' }}>总体评分</div>
              </div>
            )}
          />
        </div>

        <Space direction="vertical" style={{ width: '100%' }}>
          {scores.map((item, index) => (
            <div key={index} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Text>{item.name}</Text>
              <div style={{ flex: 1, margin: '0 16px' }}>
                <Progress 
                  percent={item.score} 
                  strokeColor={item.color}
                  size="small"
                  showInfo={false}
                />
              </div>
              <Text strong style={{ color: item.color }}>{item.score}</Text>
            </div>
          ))}
        </Space>
      </Card>
    );
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <BarChartOutlined style={{ color: '#52c41a', marginRight: '8px' }} />
          健康分析
        </Title>
        <Paragraph>
          基于AI智能分析，为您提供个性化的健康洞察和改善建议
        </Paragraph>
      </div>

      {/* 操作栏 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Select
                value={selectedPeriod}
                onChange={setSelectedPeriod}
                style={{ width: 120 }}
                options={[
                  { label: '最近7天', value: '7d' },
                  { label: '最近30天', value: '30d' },
                  { label: '最近3个月', value: '3m' },
                  { label: '最近1年', value: '1y' },
                ]}
              />
              <RangePicker />
            </Space>
          </Col>
          <Col>
            <Space>
              <Button icon={<ReloadOutlined />} onClick={loadAnalysisData}>
                重新分析
              </Button>
              <Button icon={<ExportOutlined />}>
                导出报告
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={[24, 24]}>
        {/* 左侧主要内容 */}
        <Col xs={24} lg={16}>
          <Tabs 
            activeKey={activeTab} 
            onChange={setActiveTab}
            items={[
              {
                key: 'overview',
                label: '分析概览',
                children: (
                  <Space direction="vertical" style={{ width: '100%' }} size="large">
                    {/* 健康趋势图表 */}
                    <Card title="健康趋势分析" extra={<Button icon={<EyeOutlined />} size="small">查看详情</Button>}>
                      <div style={{ height: '300px' }}>
                        {renderTrendChart()}
                      </div>
                    </Card>

                    {/* 趋势列表 */}
                    <Card title="指标变化趋势">
                      <Row gutter={[16, 16]}>
                        {healthTrends.map((trend, index) => (
                          <Col xs={24} sm={12} md={8} key={index}>
                            <Card size="small" style={{ textAlign: 'center' }}>
                              <Space direction="vertical">
                                <Text strong>{trend.metric}</Text>
                                {getTrendIcon(trend.trend)}
                                <Text style={{ color: getTrendColor(trend.trend) }}>
                                  {trend.trend === 'improving' ? '改善' : 
                                   trend.trend === 'stable' ? '稳定' : '下降'}
                                </Text>
                                <Text type="secondary">
                                  {trend.change > 0 ? '+' : ''}{trend.change}% ({trend.period})
                                </Text>
                                <Badge 
                                  color={trend.significance === 'high' ? '#ff4d4f' : 
                                         trend.significance === 'medium' ? '#faad14' : '#52c41a'}
                                  text={trend.significance === 'high' ? '显著' : 
                                        trend.significance === 'medium' ? '中等' : '轻微'}
                                />
                              </Space>
                            </Card>
                          </Col>
                        ))}
                      </Row>
                    </Card>
                  </Space>
                )
              },
              {
                key: 'risk',
                label: '风险评估',
                children: (
                  <Space direction="vertical" style={{ width: '100%' }} size="large">
                    {/* 风险分布图 */}
                    <Card title="健康风险分布">
                      <div style={{ height: '300px' }}>
                        {renderRiskChart()}
                      </div>
                    </Card>

                    {/* 风险因子详情 */}
                    <Card title="风险因子分析">
                      <Space direction="vertical" style={{ width: '100%' }}>
                        {riskFactors.map((risk) => (
                          <Card key={risk.id} size="small" style={{ marginBottom: '8px' }}>
                            <Row justify="space-between" align="middle">
                              <Col span={16}>
                                <Space direction="vertical" size="small">
                                  <Space>
                                    <Text strong>{risk.name}</Text>
                                    <Tag color={getRiskColor(risk.level)}>
                                      {risk.level === 'low' ? '低风险' : 
                                       risk.level === 'medium' ? '中风险' : '高风险'}
                                    </Tag>
                                  </Space>
                                  <Text type="secondary">{risk.impact}</Text>
                                  <div>
                                    <Text strong>预防措施: </Text>
                                    {risk.prevention.map((item, index) => (
                                      <Tag key={index} style={{ marginBottom: '4px' }}>{item}</Tag>
                                    ))}
                                  </div>
                                </Space>
                              </Col>
                              <Col span={8} style={{ textAlign: 'center' }}>
                                <Progress
                                  type="circle"
                                  percent={risk.probability}
                                  width={80}
                                  strokeColor={getRiskColor(risk.level)}
                                  format={(percent) => `${percent}%`}
                                />
                                <div style={{ marginTop: '8px' }}>
                                  <Text type="secondary" style={{ fontSize: '12px' }}>发生概率</Text>
                                </div>
                              </Col>
                            </Row>
                          </Card>
                        ))}
                      </Space>
                    </Card>
                  </Space>
                )
              },
              {
                key: 'reports',
                label: 'AI报告',
                children: (
                  <div>
                    {analysisReports.map(renderAnalysisReport)}
                  </div>
                )
              }
            ]}
          />
        </Col>

        {/* 右侧侧边栏 */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            {/* 健康评分 */}
            {renderHealthScore()}

            {/* 关键指标 */}
            <Card title="关键指标">
              <Space direction="vertical" style={{ width: '100%' }}>
                <Statistic
                  title="健康改善率"
                  value={68}
                  suffix="%"
                  valueStyle={{ color: '#52c41a' }}
                  prefix={<RiseOutlined />}
                />
                <Statistic
                  title="风险控制率"
                  value={85}
                  suffix="%"
                  valueStyle={{ color: '#1890ff' }}
                  prefix={<CheckCircleOutlined />}
                />
                <Statistic
                  title="目标达成率"
                  value={72}
                  suffix="%"
                  valueStyle={{ color: '#faad14' }}
                  prefix={<ThunderboltOutlined />}
                />
              </Space>
            </Card>

            {/* 最新动态 */}
            <Card title="分析动态">
              <Timeline size="small">
                <Timeline.Item color="green">
                  <Text>睡眠质量分析完成</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>2分钟前</Text>
                </Timeline.Item>
                <Timeline.Item color="blue">
                  <Text>心血管风险评估更新</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>1小时前</Text>
                </Timeline.Item>
                <Timeline.Item color="orange">
                  <Text>体重趋势分析完成</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>3小时前</Text>
                </Timeline.Item>
                <Timeline.Item>
                  <Text>生成月度健康报告</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>昨天</Text>
                </Timeline.Item>
              </Timeline>
            </Card>

            {/* 快速操作 */}
            <Card title="快速操作">
              <Space direction="vertical" style={{ width: '100%' }}>
                <Button block icon={<BarChartOutlined />}>
                  生成详细报告
                </Button>
                <Button block icon={<BulbOutlined />}>
                  获取改善建议
                </Button>
                <Button block icon={<ExportOutlined />}>
                  导出分析数据
                </Button>
                <Button block icon={<ClockCircleOutlined />}>
                  设置分析提醒
                </Button>
              </Space>
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
};

export default HealthAnalysis;