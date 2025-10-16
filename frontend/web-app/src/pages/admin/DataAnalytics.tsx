import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Typography,
  Select,
  DatePicker,
  Table,
  Progress,
  Space,
  Button,
  Tooltip,
  Empty,
  Spin
} from 'antd';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import {
  EyeOutlined,
  UserOutlined,
  FileTextOutlined,
  HeartOutlined,
  MessageOutlined,
  ShareAltOutlined,
  TrophyOutlined,
  RiseOutlined,
  FallOutlined,
  BarChartOutlined,
  PieChartOutlined,
  LineChartOutlined,
  DownloadOutlined
} from '@ant-design/icons';
import type { ColumnType } from 'antd/es/table';
import { apiClient } from '../../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

interface AnalyticsData {
  overview: {
    totalViews: number;
    totalUsers: number;
    totalWisdom: number;
    totalComments: number;
    avgSessionDuration: number;
    bounceRate: number;
    conversionRate: number;
    growthRate: number;
  };
  trafficTrends: Array<{
    date: string;
    views: number;
    users: number;
    sessions: number;
  }>;
  contentAnalytics: Array<{
    id: string;
    title: string;
    views: number;
    likes: number;
    comments: number;
    shares: number;
    category: string;
    publishDate: string;
  }>;
  userBehavior: {
    topPages: Array<{
      page: string;
      views: number;
      avgDuration: number;
      bounceRate: number;
    }>;
    deviceStats: Array<{
      device: string;
      count: number;
      percentage: number;
    }>;
    sourceStats: Array<{
      source: string;
      count: number;
      percentage: number;
    }>;
  };
  categoryStats: Array<{
    category: string;
    count: number;
    views: number;
    engagement: number;
  }>;
}

const DataAnalytics: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(30, 'day'),
    dayjs()
  ]);
  const [timeGranularity, setTimeGranularity] = useState<'day' | 'week' | 'month'>('day');
  const [analyticsData, setAnalyticsData] = useState<AnalyticsData | null>(null);

  // 获取分析数据
  const fetchAnalyticsData = async () => {
    setLoading(true);
    try {
      const params = {
        startDate: dateRange[0].format('YYYY-MM-DD'),
        endDate: dateRange[1].format('YYYY-MM-DD'),
        granularity: timeGranularity
      };
      const response = await apiClient.getAnalyticsData(params);
      setAnalyticsData(response.data);
    } catch {
      message.error('加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAnalyticsData();
  }, [dateRange, timeGranularity]);

  // 导出报告
  const handleExportReport = async () => {
    try {
      const params = {
        startDate: dateRange[0].format('YYYY-MM-DD'),
        endDate: dateRange[1].format('YYYY-MM-DD'),
        format: 'excel'
      };
      await apiClient.exportAnalyticsReport(params);
    } catch (error) {
      console.error('导出报告失败:', error);
    }
  };

  // 图表颜色配置
  const chartColors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#fa8c16'];

  // 内容分析表格列
  const contentColumns: ColumnType<{
    title: string;
    category: string;
    views: number;
    likes: number;
    comments: number;
    shares: number;
  }>[] = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
      render: (title: string) => (
        <Tooltip title={title}>
          <Text strong>{title}</Text>
        </Tooltip>
      )
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (category: string) => <Tag color="blue">{category}</Tag>
    },
    {
      title: '浏览量',
      dataIndex: 'views',
      key: 'views',
      width: 100,
      sorter: (a, b) => a.views - b.views,
      render: (views: number) => (
        <Space>
          <EyeOutlined />
          {views.toLocaleString()}
        </Space>
      )
    },
    {
      title: '点赞数',
      dataIndex: 'likes',
      key: 'likes',
      width: 100,
      sorter: (a, b) => a.likes - b.likes,
      render: (likes: number) => (
        <Space>
          <HeartOutlined style={{ color: '#f5222d' }} />
          {likes}
        </Space>
      )
    },
    {
      title: '评论数',
      dataIndex: 'comments',
      key: 'comments',
      width: 100,
      sorter: (a, b) => a.comments - b.comments,
      render: (comments: number) => (
        <Space>
          <MessageOutlined />
          {comments}
        </Space>
      )
    },
    {
      title: '分享数',
      dataIndex: 'shares',
      key: 'shares',
      width: 100,
      sorter: (a, b) => a.shares - b.shares,
      render: (shares: number) => (
        <Space>
          <ShareAltOutlined />
          {shares}
        </Space>
      )
    },
    {
      title: '发布时间',
      dataIndex: 'publishDate',
      key: 'publishDate',
      width: 120,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD')
    }
  ];

  // 页面访问表格列
  const pageColumns: ColumnType<{
    page: string;
    views: number;
    avgDuration: number;
    bounceRate: number;
  }>[] = [
    {
      title: '页面',
      dataIndex: 'page',
      key: 'page',
      ellipsis: true
    },
    {
      title: '浏览量',
      dataIndex: 'views',
      key: 'views',
      sorter: (a, b) => a.views - b.views,
      render: (views: number) => views.toLocaleString()
    },
    {
      title: '平均停留时间',
      dataIndex: 'avgDuration',
      key: 'avgDuration',
      render: (duration: number) => `${Math.round(duration / 60)}分${duration % 60}秒`
    },
    {
      title: '跳出率',
      dataIndex: 'bounceRate',
      key: 'bounceRate',
      render: (rate: number) => (
        <Progress
          percent={rate}
          size="small"
          status={rate > 70 ? 'exception' : rate > 50 ? 'normal' : 'success'}
        />
      )
    }
  ];

  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!analyticsData) {
    return (
      <div style={{ padding: '24px' }}>
        <Empty description="暂无数据" />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <Title level={2}>数据分析</Title>
        <Space>
          <RangePicker
            value={dateRange}
            onChange={(dates) => dates && setDateRange(dates)}
            format="YYYY-MM-DD"
          />
          <Select
            value={timeGranularity}
            onChange={setTimeGranularity}
            style={{ width: 100 }}
          >
            <Select.Option value="day">按天</Select.Option>
            <Select.Option value="week">按周</Select.Option>
            <Select.Option value="month">按月</Select.Option>
          </Select>
          <Button 
            type="primary" 
            icon={<DownloadOutlined />}
            onClick={handleExportReport}
          >
            导出报告
          </Button>
        </Space>
      </div>

      {/* 概览统计 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总浏览量"
              value={analyticsData.overview.totalViews}
              prefix={<EyeOutlined />}
              valueStyle={{ color: '#1890ff' }}
              suffix={
                <span style={{ fontSize: '12px', color: '#52c41a' }}>
                  <RiseOutlined /> {analyticsData.overview.growthRate}%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总用户数"
              value={analyticsData.overview.totalUsers}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="智慧内容"
              value={analyticsData.overview.totalWisdom}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总评论数"
              value={analyticsData.overview.totalComments}
              prefix={<MessageOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 关键指标 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="平均会话时长"
              value={Math.round(analyticsData.overview.avgSessionDuration / 60)}
              suffix="分钟"
              prefix={<BarChartOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="跳出率"
              value={analyticsData.overview.bounceRate}
              suffix="%"
              prefix={<FallOutlined />}
              valueStyle={{ 
                color: analyticsData.overview.bounceRate > 70 ? '#f5222d' : '#52c41a' 
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="转化率"
              value={analyticsData.overview.conversionRate}
              suffix="%"
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 流量趋势图 */}
      <Card title="流量趋势" style={{ marginBottom: '24px' }}>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={analyticsData.trafficTrends}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <RechartsTooltip />
            <Legend />
            <Line 
              type="monotone" 
              dataKey="views" 
              stroke="#1890ff" 
              name="浏览量"
              strokeWidth={2}
            />
            <Line 
              type="monotone" 
              dataKey="users" 
              stroke="#52c41a" 
              name="用户数"
              strokeWidth={2}
            />
            <Line 
              type="monotone" 
              dataKey="sessions" 
              stroke="#faad14" 
              name="会话数"
              strokeWidth={2}
            />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      <Row gutter={16} style={{ marginBottom: '24px' }}>
        {/* 设备统计 */}
        <Col xs={24} lg={12}>
          <Card title="设备分布" extra={<PieChartOutlined />}>
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={analyticsData.userBehavior.deviceStats}
                  cx="50%"
                  cy="50%"
                  outerRadius={80}
                  dataKey="count"
                  nameKey="device"
                  label={({ device, percentage }) => `${device} ${percentage}%`}
                >
                  {analyticsData.userBehavior.deviceStats.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={chartColors[index % chartColors.length]} />
                  ))}
                </Pie>
                <RechartsTooltip />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>

        {/* 来源统计 */}
        <Col xs={24} lg={12}>
          <Card title="流量来源" extra={<BarChartOutlined />}>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={analyticsData.userBehavior.sourceStats}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="source" />
                <YAxis />
                <RechartsTooltip />
                <Bar dataKey="count" fill="#1890ff" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      {/* 分类统计 */}
      <Card title="分类统计" style={{ marginBottom: '24px' }}>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={analyticsData.categoryStats}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="category" />
            <YAxis />
            <RechartsTooltip />
            <Legend />
            <Area 
              type="monotone" 
              dataKey="count" 
              stackId="1" 
              stroke="#1890ff" 
              fill="#1890ff" 
              name="内容数量"
            />
            <Area 
              type="monotone" 
              dataKey="views" 
              stackId="2" 
              stroke="#52c41a" 
              fill="#52c41a" 
              name="浏览量"
            />
          </AreaChart>
        </ResponsiveContainer>
      </Card>

      <Row gutter={16}>
        {/* 热门内容 */}
        <Col xs={24} lg={14}>
          <Card title="热门内容" extra={<LineChartOutlined />}>
            <Table
              columns={contentColumns}
              dataSource={analyticsData.contentAnalytics}
              rowKey="id"
              pagination={{ pageSize: 5 }}
              size="small"
            />
          </Card>
        </Col>

        {/* 热门页面 */}
        <Col xs={24} lg={10}>
          <Card title="热门页面">
            <Table
              columns={pageColumns}
              dataSource={analyticsData.userBehavior.topPages}
              rowKey="page"
              pagination={{ pageSize: 5 }}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default DataAnalytics;