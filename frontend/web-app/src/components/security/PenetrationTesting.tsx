import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Tag, Space, Modal, Form, Input, Select, DatePicker, Row, Col, Statistic, Descriptions, Progress, Steps, Timeline } from 'antd';
import { ExperimentOutlined, PlayCircleOutlined, PauseCircleOutlined, CheckCircleOutlined, ClockCircleOutlined, EyeOutlined, FileTextOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Option } = Select;
const { RangePicker } = DatePicker;
const { Step } = Steps;

interface PentestProject {
  id: string;
  name: string;
  target: string;
  scope: string;
  methodology: 'owasp' | 'nist' | 'ptes' | 'custom';
  status: 'planning' | 'in_progress' | 'completed' | 'paused' | 'cancelled';
  progress: number;
  startDate: string;
  endDate?: string;
  tester: string;
  findings: number;
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
}

interface PentestResult {
  id: string;
  projectId: string;
  title: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  category: string;
  description: string;
  impact: string;
  recommendation: string;
  status: 'open' | 'fixed' | 'accepted' | 'mitigated';
  discoveredAt: string;
}

const PenetrationTesting: React.FC = () => {
  const [projects, setProjects] = useState<PentestProject[]>([]);
  const [results, setResults] = useState<PentestResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [projectModalVisible, setProjectModalVisible] = useState(false);
  const [selectedProject, setSelectedProject] = useState<PentestProject | null>(null);
  const [selectedResult, setSelectedResult] = useState<PentestResult | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadProjects();
    loadResults();
  }, []);

  const loadProjects = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      setTimeout(() => {
        setProjects([
          {
            id: '1',
            name: 'Web应用渗透测试',
            target: 'https://app.example.com',
            scope: 'Web应用程序及API接口',
            methodology: 'owasp',
            status: 'in_progress',
            progress: 65,
            startDate: '2024-01-10',
            tester: '张三',
            findings: 8,
            riskLevel: 'high'
          },
          {
            id: '2',
            name: '内网渗透测试',
            target: '192.168.1.0/24',
            scope: '内部网络基础设施',
            methodology: 'ptes',
            status: 'completed',
            progress: 100,
            startDate: '2024-01-01',
            endDate: '2024-01-08',
            tester: '李四',
            findings: 12,
            riskLevel: 'critical'
          },
          {
            id: '3',
            name: '移动应用安全测试',
            target: 'Mobile App v2.1',
            scope: 'Android/iOS移动应用',
            methodology: 'custom',
            status: 'planning',
            progress: 0,
            startDate: '2024-01-20',
            tester: '王五',
            findings: 0,
            riskLevel: 'medium'
          }
        ]);
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load projects:', error);
      setLoading(false);
    }
  };

  const loadResults = async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        setResults([
          {
            id: '1',
            projectId: '1',
            title: 'SQL注入漏洞',
            severity: 'critical',
            category: '注入攻击',
            description: '在登录页面发现SQL注入漏洞，可绕过身份验证',
            impact: '攻击者可获取数据库中的敏感信息，包括用户凭据',
            recommendation: '使用参数化查询，实施输入验证',
            status: 'open',
            discoveredAt: '2024-01-12 14:30:00'
          },
          {
            id: '2',
            projectId: '1',
            title: '跨站脚本攻击(XSS)',
            severity: 'high',
            category: '客户端攻击',
            description: '用户输入未经过滤直接输出到页面',
            impact: '可能导致用户会话劫持和敏感信息泄露',
            recommendation: '对用户输入进行HTML编码，实施CSP策略',
            status: 'fixed',
            discoveredAt: '2024-01-11 10:15:00'
          },
          {
            id: '3',
            projectId: '2',
            title: '弱密码策略',
            severity: 'medium',
            category: '身份验证',
            description: '系统允许使用简单密码',
            impact: '增加暴力破解攻击的成功率',
            recommendation: '实施强密码策略，启用账户锁定机制',
            status: 'mitigated',
            discoveredAt: '2024-01-05 16:20:00'
          }
        ]);
      }, 500);
    } catch (error) {
      console.error('Failed to load results:', error);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'yellow';
      case 'low': return 'blue';
      default: return 'default';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'planning': return 'blue';
      case 'in_progress': return 'orange';
      case 'completed': return 'green';
      case 'paused': return 'yellow';
      case 'cancelled': return 'red';
      default: return 'default';
    }
  };

  const getResultStatusColor = (status: string) => {
    switch (status) {
      case 'open': return 'red';
      case 'fixed': return 'green';
      case 'accepted': return 'gray';
      case 'mitigated': return 'orange';
      default: return 'default';
    }
  };

  const projectColumns: ColumnsType<PentestProject> = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '测试目标',
      dataIndex: 'target',
      key: 'target',
    },
    {
      title: '测试方法',
      dataIndex: 'methodology',
      key: 'methodology',
      render: (methodology) => (
        <Tag>
          {methodology.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status, record) => (
        <Space direction="vertical" size="small">
          <Tag color={getStatusColor(status)}>
            {status === 'planning' ? '计划中' :
             status === 'in_progress' ? '进行中' :
             status === 'completed' ? '已完成' :
             status === 'paused' ? '已暂停' : '已取消'}
          </Tag>
          {status === 'in_progress' && (
            <Progress percent={record.progress} size="small" />
          )}
        </Space>
      ),
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      render: (riskLevel) => (
        <Tag color={getSeverityColor(riskLevel)}>
          {riskLevel.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '发现问题',
      dataIndex: 'findings',
      key: 'findings',
      render: (findings) => (
        <Tag color={findings > 0 ? 'red' : 'green'}>
          {findings} 个
        </Tag>
      ),
    },
    {
      title: '测试人员',
      dataIndex: 'tester',
      key: 'tester',
    },
    {
      title: '开始日期',
      dataIndex: 'startDate',
      key: 'startDate',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<EyeOutlined />}
            onClick={() => {
              setSelectedProject(record);
              setProjectModalVisible(true);
            }}
          >
            查看详情
          </Button>
          <Button 
            type="link" 
            icon={<FileTextOutlined />}
          >
            生成报告
          </Button>
        </Space>
      ),
    },
  ];

  const resultColumns: ColumnsType<PentestResult> = [
    {
      title: '问题标题',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity) => (
        <Tag color={getSeverityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getResultStatusColor(status)}>
          {status === 'open' ? '待修复' :
           status === 'fixed' ? '已修复' :
           status === 'accepted' ? '已接受' : '已缓解'}
        </Tag>
      ),
    },
    {
      title: '发现时间',
      dataIndex: 'discoveredAt',
      key: 'discoveredAt',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<EyeOutlined />}
            onClick={() => {
              setSelectedResult(record);
              setModalVisible(true);
            }}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  const handleCreateProject = (values: any) => {
    console.log('Creating project with values:', values);
    setProjectModalVisible(false);
    form.resetFields();
    // 这里会调用API创建项目
  };

  const getProjectSteps = (status: string, progress: number) => {
    const steps = [
      { title: '项目规划', status: 'finish' },
      { title: '信息收集', status: progress > 20 ? 'finish' : 'wait' },
      { title: '漏洞发现', status: progress > 40 ? 'finish' : progress > 20 ? 'process' : 'wait' },
      { title: '漏洞利用', status: progress > 60 ? 'finish' : progress > 40 ? 'process' : 'wait' },
      { title: '报告生成', status: progress === 100 ? 'finish' : progress > 80 ? 'process' : 'wait' }
    ];
    return steps;
  };

  return (
    <div>
      {/* 渗透测试统计 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="进行中项目"
              value={2}
              prefix={<ExperimentOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="已完成项目"
              value={8}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="发现漏洞"
              value={45}
              prefix={<ExperimentOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="平均用时"
              value={7}
              suffix="天"
              prefix={<ClockCircleOutlined style={{ color: '#fa8c16' }} />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 渗透测试项目列表 */}
      <Card 
        title="渗透测试项目" 
        extra={
          <Space>
            <Button 
              type="primary" 
              icon={<ExperimentOutlined />}
              onClick={() => setProjectModalVisible(true)}
            >
              新建项目
            </Button>
            <Button onClick={loadProjects}>刷新</Button>
          </Space>
        }
        style={{ marginBottom: 24 }}
      >
        <Table
          columns={projectColumns}
          dataSource={projects}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 测试结果列表 */}
      <Card 
        title="测试结果" 
        extra={
          <Space>
            <Button>导出报告</Button>
            <Button onClick={loadResults}>刷新</Button>
          </Space>
        }
      >
        <Table
          columns={resultColumns}
          dataSource={results}
          rowKey="id"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 项目详情模态框 */}
      <Modal
        title="项目详情"
        open={projectModalVisible && selectedProject !== null}
        onCancel={() => setProjectModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setProjectModalVisible(false)}>
            关闭
          </Button>,
          <Button key="report" type="primary" icon={<FileTextOutlined />}>
            生成报告
          </Button>,
        ]}
        width={900}
      >
        {selectedProject && (
          <div>
            <Descriptions bordered column={2} style={{ marginBottom: 24 }}>
              <Descriptions.Item label="项目名称" span={2}>
                {selectedProject.name}
              </Descriptions.Item>
              <Descriptions.Item label="测试目标">
                {selectedProject.target}
              </Descriptions.Item>
              <Descriptions.Item label="测试范围">
                {selectedProject.scope}
              </Descriptions.Item>
              <Descriptions.Item label="测试方法">
                <Tag>{selectedProject.methodology.toUpperCase()}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="风险等级">
                <Tag color={getSeverityColor(selectedProject.riskLevel)}>
                  {selectedProject.riskLevel.toUpperCase()}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="测试人员">
                {selectedProject.tester}
              </Descriptions.Item>
              <Descriptions.Item label="开始日期">
                {selectedProject.startDate}
              </Descriptions.Item>
              <Descriptions.Item label="发现问题">
                <Tag color={selectedProject.findings > 0 ? 'red' : 'green'}>
                  {selectedProject.findings} 个
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="完成进度">
                <Progress percent={selectedProject.progress} />
              </Descriptions.Item>
            </Descriptions>

            <Card title="测试进度" size="small">
              <Steps 
                current={Math.floor(selectedProject.progress / 20)}
                items={getProjectSteps(selectedProject.status, selectedProject.progress)}
              />
            </Card>
          </div>
        )}
      </Modal>

      {/* 测试结果详情模态框 */}
      <Modal
        title="测试结果详情"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setModalVisible(false)}>
            关闭
          </Button>,
          <Button key="fix" type="primary">
            标记为已修复
          </Button>,
        ]}
        width={800}
      >
        {selectedResult && (
          <div>
            <Descriptions bordered column={1}>
              <Descriptions.Item label="问题标题">
                {selectedResult.title}
              </Descriptions.Item>
              <Descriptions.Item label="严重程度">
                <Tag color={getSeverityColor(selectedResult.severity)}>
                  {selectedResult.severity.toUpperCase()}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="类别">
                {selectedResult.category}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getResultStatusColor(selectedResult.status)}>
                  {selectedResult.status === 'open' ? '待修复' :
                   selectedResult.status === 'fixed' ? '已修复' :
                   selectedResult.status === 'accepted' ? '已接受' : '已缓解'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="问题描述">
                {selectedResult.description}
              </Descriptions.Item>
              <Descriptions.Item label="影响分析">
                {selectedResult.impact}
              </Descriptions.Item>
              <Descriptions.Item label="修复建议">
                {selectedResult.recommendation}
              </Descriptions.Item>
              <Descriptions.Item label="发现时间">
                {selectedResult.discoveredAt}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      {/* 新建项目模态框 */}
      <Modal
        title="新建渗透测试项目"
        open={projectModalVisible && selectedProject === null}
        onCancel={() => setProjectModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateProject}
        >
          <Form.Item
            name="name"
            label="项目名称"
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>
          
          <Form.Item
            name="target"
            label="测试目标"
            rules={[{ required: true, message: '请输入测试目标' }]}
          >
            <Input placeholder="URL、IP地址或应用名称" />
          </Form.Item>

          <Form.Item
            name="scope"
            label="测试范围"
            rules={[{ required: true, message: '请输入测试范围' }]}
          >
            <Input.TextArea rows={3} placeholder="详细描述测试范围和边界" />
          </Form.Item>

          <Form.Item
            name="methodology"
            label="测试方法"
            rules={[{ required: true, message: '请选择测试方法' }]}
          >
            <Select placeholder="请选择测试方法">
              <Option value="owasp">OWASP Testing Guide</Option>
              <Option value="nist">NIST SP 800-115</Option>
              <Option value="ptes">PTES</Option>
              <Option value="custom">自定义方法</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="tester"
            label="测试人员"
            rules={[{ required: true, message: '请输入测试人员' }]}
          >
            <Input placeholder="请输入测试人员姓名" />
          </Form.Item>

          <Form.Item
            name="dateRange"
            label="测试时间"
            rules={[{ required: true, message: '请选择测试时间' }]}
          >
            <RangePicker style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PenetrationTesting;