import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Typography, Row, Col, Statistic, Progress, Button, Select, Input, DatePicker, Tooltip } from 'antd';
import { MonitorOutlined, CheckCircleOutlined, ClockCircleOutlined, ExclamationCircleOutlined, ReloadOutlined, FilterOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { ColumnsType } from 'antd/es/table';

const { Title, Text } = Typography;
const { Search } = Input;
const { RangePicker } = DatePicker;

interface APIStatusItem {
  id: string;
  name: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  path: string;
  status: 'development' | 'testing' | 'production';
  version: string;
  module: string;
  lastUpdated: string;
  developer: string;
  testCoverage?: number;
  uptime?: number;
  responseTime?: number;
}

const APIStatus: React.FC = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [moduleFilter, setModuleFilter] = useState<string>('all');

  // 模拟API状态数据
  const [apiStatusData] = useState<APIStatusItem[]>([
    {
      id: '1',
      name: '用户登录',
      method: 'POST',
      path: '/api/auth/login',
      status: 'production',
      version: 'v1.0',
      module: '用户管理',
      lastUpdated: '2024-01-15',
      developer: '张三',
      testCoverage: 95,
      uptime: 99.9,
      responseTime: 120
    },
    {
      id: '2',
      name: '用户注册',
      method: 'POST',
      path: '/api/auth/register',
      status: 'production',
      version: 'v1.0',
      module: '用户管理',
      lastUpdated: '2024-01-14',
      developer: '李四',
      testCoverage: 88,
      uptime: 99.8,
      responseTime: 150
    },
    {
      id: '3',
      name: '创建项目',
      method: 'POST',
      path: '/api/projects',
      status: 'testing',
      version: 'v1.1',
      module: '项目管理',
      lastUpdated: '2024-01-16',
      developer: '王五',
      testCoverage: 75,
      uptime: 98.5,
      responseTime: 200
    },
    {
      id: '4',
      name: '多模态分析',
      method: 'POST',
      path: '/api/ai/multimodal',
      status: 'development',
      version: 'v2.0',
      module: 'AI服务',
      lastUpdated: '2024-01-17',
      developer: '赵六',
      testCoverage: 45,
      uptime: 85.0,
      responseTime: 500
    },
    {
      id: '5',
      name: '图像生成',
      method: 'POST',
      path: '/api/ai/image/generate',
      status: 'testing',
      version: 'v1.5',
      module: 'AI服务',
      lastUpdated: '2024-01-16',
      developer: '钱七',
      testCoverage: 68,
      uptime: 96.2,
      responseTime: 800
    }
  ]);

  const getStatusColor = (status: string) => {
    const colors = {
      development: 'orange',
      testing: 'blue',
      production: 'green'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusText = (status: string) => {
    const texts = {
      development: '开发中',
      testing: '测试中',
      production: '已上线'
    };
    return texts[status as keyof typeof texts] || status;
  };

  const getStatusIcon = (status: string) => {
    const icons = {
      development: <ClockCircleOutlined />,
      testing: <ExclamationCircleOutlined />,
      production: <CheckCircleOutlined />
    };
    return icons[status as keyof typeof icons] || null;
  };

  const getMethodColor = (method: string) => {
    const colors = {
      GET: 'green',
      POST: 'blue',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple'
    };
    return colors[method as keyof typeof colors] || 'default';
  };

  const columns: ColumnsType<APIStatusItem> = [
    {
      title: '接口名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text, record) => (
        <Space direction="vertical" size="small">
          <Text strong>{text}</Text>
          <Text code style={{ fontSize: '12px' }}>{record.path}</Text>
        </Space>
      )
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      render: (method) => (
        <Tag color={getMethodColor(method)}>{method}</Tag>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => (
        <Tag color={getStatusColor(status)} icon={getStatusIcon(status)}>
          {getStatusText(status)}
        </Tag>
      )
    },
    {
      title: '模块',
      dataIndex: 'module',
      key: 'module',
      width: 120
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 80,
      render: (version) => <Tag color="cyan">{version}</Tag>
    },
    {
      title: '测试覆盖率',
      dataIndex: 'testCoverage',
      key: 'testCoverage',
      width: 120,
      render: (coverage) => (
        <Tooltip title={`测试覆盖率: ${coverage}%`}>
          <Progress 
            percent={coverage} 
            size="small" 
            status={coverage >= 80 ? 'success' : coverage >= 60 ? 'active' : 'exception'}
            showInfo={false}
          />
          <Text style={{ fontSize: '12px', marginLeft: '8px' }}>{coverage}%</Text>
        </Tooltip>
      )
    },
    {
      title: '可用性',
      dataIndex: 'uptime',
      key: 'uptime',
      width: 100,
      render: (uptime) => (
        <Text style={{ color: uptime >= 99 ? '#52c41a' : uptime >= 95 ? '#faad14' : '#ff4d4f' }}>
          {uptime}%
        </Text>
      )
    },
    {
      title: '响应时间',
      dataIndex: 'responseTime',
      key: 'responseTime',
      width: 100,
      render: (time) => (
        <Text style={{ color: time <= 200 ? '#52c41a' : time <= 500 ? '#faad14' : '#ff4d4f' }}>
          {time}ms
        </Text>
      )
    },
    {
      title: '开发者',
      dataIndex: 'developer',
      key: 'developer',
      width: 100
    },
    {
      title: '最后更新',
      dataIndex: 'lastUpdated',
      key: 'lastUpdated',
      width: 120
    }
  ];

  // 统计数据
  const statusStats = {
    total: apiStatusData.length,
    production: apiStatusData.filter(item => item.status === 'production').length,
    testing: apiStatusData.filter(item => item.status === 'testing').length,
    development: apiStatusData.filter(item => item.status === 'development').length
  };

  // 过滤数据
  const filteredData = apiStatusData.filter(item => {
    const matchesSearch = item.name.toLowerCase().includes(searchText.toLowerCase()) ||
                         item.path.toLowerCase().includes(searchText.toLowerCase());
    const matchesStatus = statusFilter === 'all' || item.status === statusFilter;
    const matchesModule = moduleFilter === 'all' || item.module === moduleFilter;
    
    return matchesSearch && matchesStatus && matchesModule;
  });

  const handleRefresh = () => {
    setLoading(true);
    // 模拟刷新
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  };

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <MonitorOutlined /> 接口状态
      </Title>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="总接口数"
              value={statusStats.total}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="已上线"
              value={statusStats.production}
              valueStyle={{ color: '#52c41a' }}
              suffix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="测试中"
              value={statusStats.testing}
              valueStyle={{ color: '#1890ff' }}
              suffix={<ExclamationCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <Statistic
              title="开发中"
              value={statusStats.development}
              valueStyle={{ color: '#faad14' }}
              suffix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 过滤器 */}
      <Card size="small" style={{ marginBottom: '16px' }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={8}>
            <Search
              placeholder="搜索接口名称或路径..."
              allowClear
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
            />
          </Col>
          <Col xs={12} sm={4}>
            <Select
              style={{ width: '100%' }}
              placeholder="状态筛选"
              value={statusFilter}
              onChange={setStatusFilter}
            >
              <Select.Option value="all">全部状态</Select.Option>
              <Select.Option value="production">已上线</Select.Option>
              <Select.Option value="testing">测试中</Select.Option>
              <Select.Option value="development">开发中</Select.Option>
            </Select>
          </Col>
          <Col xs={12} sm={4}>
            <Select
              style={{ width: '100%' }}
              placeholder="模块筛选"
              value={moduleFilter}
              onChange={setModuleFilter}
            >
              <Select.Option value="all">全部模块</Select.Option>
              <Select.Option value="用户管理">用户管理</Select.Option>
              <Select.Option value="项目管理">项目管理</Select.Option>
              <Select.Option value="AI服务">AI服务</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={8}>
            <Space>
              <Button 
                icon={<ReloadOutlined />} 
                onClick={handleRefresh}
                loading={loading}
              >
                刷新
              </Button>
              <Button icon={<FilterOutlined />}>
                高级筛选
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 接口状态表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={filteredData}
          rowKey="id"
          loading={loading}
          pagination={{
            total: filteredData.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default APIStatus;