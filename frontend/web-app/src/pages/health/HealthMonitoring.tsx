import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Typography, 
  Space, 
  Statistic, 
  Progress, 
  Select, 
  DatePicker, 
  Table, 
  Tag, 
  Alert,
  Divider,
  Tooltip,
  Badge,
  Modal,
  Form,
  Input,
  InputNumber,
  message
} from 'antd';
import { 
  HeartOutlined, 
  DashboardOutlined, 
  PlusOutlined, 
  SyncOutlined,
  SettingOutlined,
  ExportOutlined,
  ImportOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined,
  LineChartOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { Line, Area, Gauge } from '@ant-design/plots';

const { Title, Paragraph, Text } = Typography;
const { RangePicker } = DatePicker;

interface HealthData {
  id: string;
  type: string;
  value: number;
  unit: string;
  timestamp: Date;
  status: 'normal' | 'warning' | 'danger';
  source: string;
  notes?: string;
}

interface HealthMetric {
  id: string;
  name: string;
  type: string;
  currentValue: number;
  unit: string;
  status: 'normal' | 'warning' | 'danger';
  normalRange: { min: number; max: number };
  trend: 'up' | 'down' | 'stable';
  lastUpdate: Date;
  data: Array<{ time: string; value: number }>;
}

const HealthMonitoring: React.FC = () => {
  const [healthMetrics, setHealthMetrics] = useState<HealthMetric[]>([]);
  const [selectedMetric, setSelectedMetric] = useState<string>('heartRate');
  const [timeRange, setTimeRange] = useState<string>('7d');
  const [addDataVisible, setAddDataVisible] = useState(false);
  const [loading, setLoading] = useState(true);
  const [form] = Form.useForm();

  useEffect(() => {
    loadHealthData();
  }, []);

  const loadHealthData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockData: HealthMetric[] = [
        {
          id: 'heartRate',
          name: '心率',
          type: 'heartRate',
          currentValue: 72,
          unit: 'bpm',
          status: 'normal',
          normalRange: { min: 60, max: 100 },
          trend: 'stable',
          lastUpdate: new Date(),
          data: generateTimeSeriesData(72, 7)
        },
        {
          id: 'bloodPressure',
          name: '血压',
          type: 'bloodPressure',
          currentValue: 120,
          unit: 'mmHg',
          status: 'normal',
          normalRange: { min: 90, max: 140 },
          trend: 'down',
          lastUpdate: new Date(),
          data: generateTimeSeriesData(120, 7)
        },
        {
          id: 'weight',
          name: '体重',
          type: 'weight',
          currentValue: 65.5,
          unit: 'kg',
          status: 'normal',
          normalRange: { min: 60, max: 70 },
          trend: 'down',
          lastUpdate: new Date(),
          data: generateTimeSeriesData(65.5, 7)
        },
        {
          id: 'temperature',
          name: '体温',
          type: 'temperature',
          currentValue: 36.5,
          unit: '°C',
          status: 'normal',
          normalRange: { min: 36, max: 37.5 },
          trend: 'stable',
          lastUpdate: new Date(),
          data: generateTimeSeriesData(36.5, 7)
        },
        {
          id: 'bloodSugar',
          name: '血糖',
          type: 'bloodSugar',
          currentValue: 5.2,
          unit: 'mmol/L',
          status: 'normal',
          normalRange: { min: 3.9, max: 6.1 },
          trend: 'up',
          lastUpdate: new Date(),
          data: generateTimeSeriesData(5.2, 7)
        },
        {
          id: 'sleepQuality',
          name: '睡眠质量',
          type: 'sleepQuality',
          currentValue: 85,
          unit: '%',
          status: 'normal',
          normalRange: { min: 70, max: 100 },
          trend: 'up',
          lastUpdate: new Date(),
          data: generateTimeSeriesData(85, 7)
        }
      ];
      setHealthMetrics(mockData);
      setLoading(false);
    }, 1000);
  };

  // 生成时间序列数据
  const generateTimeSeriesData = (baseValue: number, days: number) => {
    const data = [];
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      const variation = (Math.random() - 0.5) * baseValue * 0.1;
      data.push({
        time: date.toISOString().split('T')[0],
        value: Math.round((baseValue + variation) * 10) / 10
      });
    }
    return data;
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'normal': return '#52c41a';
      case 'warning': return '#faad14';
      case 'danger': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'normal': return '正常';
      case 'warning': return '注意';
      case 'danger': return '异常';
      default: return '未知';
    }
  };

  // 获取趋势图标
  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return '↗️';
      case 'down': return '↘️';
      case 'stable': return '➡️';
      default: return '➡️';
    }
  };

  // 添加健康数据
  const handleAddData = async (values: any) => {
    try {
      // 这里应该调用API保存数据
      console.log('添加健康数据:', values);
      message.success('健康数据添加成功');
      setAddDataVisible(false);
      form.resetFields();
      loadHealthData(); // 重新加载数据
    } catch (error) {
      message.error('添加失败，请重试');
    }
  };

  // 渲染指标卡片
  const renderMetricCard = (metric: HealthMetric) => {
    const progressPercent = ((metric.currentValue - metric.normalRange.min) / 
      (metric.normalRange.max - metric.normalRange.min)) * 100;

    return (
      <Card
        key={metric.id}
        hoverable
        style={{ 
          border: selectedMetric === metric.id ? `2px solid ${getStatusColor(metric.status)}` : '1px solid #d9d9d9'
        }}
        onClick={() => setSelectedMetric(metric.id)}
      >
        <div style={{ textAlign: 'center' }}>
          <Statistic
            title={metric.name}
            value={metric.currentValue}
            suffix={metric.unit}
            valueStyle={{ 
              color: getStatusColor(metric.status),
              fontSize: '24px'
            }}
          />
          
          <div style={{ marginTop: '12px', marginBottom: '12px' }}>
            <Progress
              percent={Math.max(0, Math.min(100, progressPercent))}
              strokeColor={getStatusColor(metric.status)}
              size="small"
              showInfo={false}
            />
            <Text type="secondary" style={{ fontSize: '12px' }}>
              正常范围: {metric.normalRange.min}-{metric.normalRange.max} {metric.unit}
            </Text>
          </div>

          <Space>
            <Badge color={getStatusColor(metric.status)} text={getStatusText(metric.status)} />
            <Text type="secondary">{getTrendIcon(metric.trend)}</Text>
          </Space>
          
          <div style={{ marginTop: '8px' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              更新时间: {metric.lastUpdate.toLocaleTimeString()}
            </Text>
          </div>
        </div>
      </Card>
    );
  };

  // 渲染趋势图表
  const renderTrendChart = () => {
    const selectedMetricData = healthMetrics.find(m => m.id === selectedMetric);
    if (!selectedMetricData) return null;

    const config = {
      data: selectedMetricData.data,
      xField: 'time',
      yField: 'value',
      smooth: true,
      color: getStatusColor(selectedMetricData.status),
      point: {
        size: 4,
        shape: 'circle',
      },
      area: {
        style: {
          fill: `l(270) 0:${getStatusColor(selectedMetricData.status)}20 1:${getStatusColor(selectedMetricData.status)}05`,
        },
      },
      xAxis: {
        type: 'time',
        tickCount: 7,
      },
      yAxis: {
        title: {
          text: `${selectedMetricData.name} (${selectedMetricData.unit})`,
        },
      },
      annotations: [
        {
          type: 'line',
          start: ['min', selectedMetricData.normalRange.min],
          end: ['max', selectedMetricData.normalRange.min],
          style: {
            stroke: '#faad14',
            lineDash: [4, 4],
          },
        },
        {
          type: 'line',
          start: ['min', selectedMetricData.normalRange.max],
          end: ['max', selectedMetricData.normalRange.max],
          style: {
            stroke: '#faad14',
            lineDash: [4, 4],
          },
        },
      ],
    };

    return <Area {...config} />;
  };

  // 渲染仪表盘
  const renderGaugeChart = () => {
    const selectedMetricData = healthMetrics.find(m => m.id === selectedMetric);
    if (!selectedMetricData) return null;

    const { currentValue, normalRange } = selectedMetricData;
    const percent = ((currentValue - normalRange.min) / (normalRange.max - normalRange.min));

    const config = {
      percent: Math.max(0, Math.min(1, percent)),
      range: {
        color: getStatusColor(selectedMetricData.status),
      },
      indicator: {
        pointer: {
          style: {
            stroke: '#D0D0D0',
          },
        },
        pin: {
          style: {
            stroke: '#D0D0D0',
          },
        },
      },
      statistic: {
        content: {
          style: {
            fontSize: '24px',
            lineHeight: '24px',
            color: getStatusColor(selectedMetricData.status),
          },
          formatter: () => `${currentValue}${selectedMetricData.unit}`,
        },
      },
    };

    return <Gauge {...config} />;
  };

  // 数据表格列配置
  const tableColumns = [
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '数值',
      dataIndex: 'value',
      key: 'value',
      render: (value: number) => {
        const metric = healthMetrics.find(m => m.id === selectedMetric);
        return `${value} ${metric?.unit || ''}`;
      },
    },
    {
      title: '状态',
      key: 'status',
      render: (record: any) => {
        const metric = healthMetrics.find(m => m.id === selectedMetric);
        if (!metric) return null;
        
        const { normalRange } = metric;
        const status = record.value >= normalRange.min && record.value <= normalRange.max 
          ? 'normal' 
          : record.value < normalRange.min || record.value > normalRange.max 
          ? 'warning' 
          : 'danger';
        
        return <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>;
      },
    },
  ];

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <DashboardOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          健康监测
        </Title>
        <Paragraph>
          实时监测您的健康指标，及时发现健康变化趋势
        </Paragraph>
      </div>

      {/* 操作栏 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Select
                value={timeRange}
                onChange={setTimeRange}
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
              <Button icon={<PlusOutlined />} type="primary" onClick={() => setAddDataVisible(true)}>
                添加数据
              </Button>
              <Button icon={<SyncOutlined />} onClick={loadHealthData}>
                刷新
              </Button>
              <Button icon={<ExportOutlined />}>
                导出数据
              </Button>
              <Button icon={<SettingOutlined />}>
                设置
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 健康指标卡片 */}
      <Card title="健康指标" style={{ marginBottom: '24px' }}>
        <Row gutter={[16, 16]}>
          {healthMetrics.map((metric) => (
            <Col xs={24} sm={12} md={8} lg={4} key={metric.id}>
              {renderMetricCard(metric)}
            </Col>
          ))}
        </Row>
      </Card>

      {/* 详细分析 */}
      <Row gutter={[24, 24]}>
        {/* 趋势图表 */}
        <Col xs={24} lg={16}>
          <Card 
            title={`${healthMetrics.find(m => m.id === selectedMetric)?.name || ''} 趋势分析`}
            extra={
              <Space>
                <Button icon={<LineChartOutlined />} size="small">
                  切换图表
                </Button>
                <Button icon={<ExportOutlined />} size="small">
                  导出图表
                </Button>
              </Space>
            }
          >
            <div style={{ height: '300px' }}>
              {renderTrendChart()}
            </div>
          </Card>

          {/* 数据表格 */}
          <Card title="历史数据" style={{ marginTop: '24px' }}>
            <Table
              columns={tableColumns}
              dataSource={healthMetrics.find(m => m.id === selectedMetric)?.data || []}
              rowKey="time"
              size="small"
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </Col>

        {/* 右侧面板 */}
        <Col xs={24} lg={8}>
          {/* 仪表盘 */}
          <Card title="当前状态" style={{ marginBottom: '24px' }}>
            <div style={{ height: '200px' }}>
              {renderGaugeChart()}
            </div>
          </Card>

          {/* 健康提醒 */}
          <Card title="健康提醒" style={{ marginBottom: '24px' }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Alert
                message="血压偏高"
                description="最近3天血压持续偏高，建议减少盐分摄入"
                type="warning"
                showIcon
                icon={<WarningOutlined />}
              />
              <Alert
                message="睡眠质量良好"
                description="本周睡眠质量稳定，继续保持"
                type="success"
                showIcon
                icon={<CheckCircleOutlined />}
              />
              <Alert
                message="运动提醒"
                description="今日运动量不足，建议增加30分钟有氧运动"
                type="info"
                showIcon
                icon={<InfoCircleOutlined />}
              />
            </Space>
          </Card>

          {/* 快速操作 */}
          <Card title="快速操作">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Button block icon={<PlusOutlined />}>
                手动添加数据
              </Button>
              <Button block icon={<CalendarOutlined />}>
                设置提醒
              </Button>
              <Button block icon={<ExportOutlined />}>
                生成报告
              </Button>
              <Button block icon={<SettingOutlined />}>
                设备管理
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 添加数据模态框 */}
      <Modal
        title="添加健康数据"
        visible={addDataVisible}
        onCancel={() => setAddDataVisible(false)}
        onOk={() => form.submit()}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddData}
        >
          <Form.Item
            name="type"
            label="数据类型"
            rules={[{ required: true, message: '请选择数据类型' }]}
          >
            <Select
              placeholder="选择要添加的健康数据类型"
              options={healthMetrics.map(metric => ({
                label: metric.name,
                value: metric.type
              }))}
            />
          </Form.Item>

          <Form.Item
            name="value"
            label="数值"
            rules={[{ required: true, message: '请输入数值' }]}
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder="请输入测量数值"
              precision={1}
            />
          </Form.Item>

          <Form.Item
            name="timestamp"
            label="测量时间"
            rules={[{ required: true, message: '请选择测量时间' }]}
          >
            <DatePicker
              showTime
              style={{ width: '100%' }}
              placeholder="选择测量时间"
            />
          </Form.Item>

          <Form.Item
            name="source"
            label="数据来源"
          >
            <Select
              placeholder="选择数据来源"
              options={[
                { label: '手动输入', value: 'manual' },
                { label: '智能手环', value: 'smartband' },
                { label: '血压计', value: 'bloodpressure' },
                { label: '体重秤', value: 'scale' },
                { label: '血糖仪', value: 'glucometer' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="notes"
            label="备注"
          >
            <Input.TextArea
              rows={3}
              placeholder="添加备注信息（可选）"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default HealthMonitoring;