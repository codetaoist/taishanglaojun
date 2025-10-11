// 合规性监控组件
import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Row,
  Col,
  Tabs,
  Progress,
  Alert,
  Statistic,
  Timeline,
  Badge,
  Tooltip,
  Popconfirm,
  DatePicker,
  message,
  Descriptions,
  List,
  Avatar,
  Divider
} from 'antd';
import {
  ShieldCheckOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  FileTextOutlined,
  AlertOutlined,
  SettingOutlined,
  DownloadOutlined,
  ReloadOutlined,
  BellOutlined,
  UserOutlined,
  GlobalOutlined,
  LockOutlined
} from '@ant-design/icons';
import moment from 'moment';

const { Option } = Select;
const { TabPane } = Tabs;
const { TextArea } = Input;
const { RangePicker } = DatePicker;

interface ComplianceStatus {
  region: string;
  regulation: string;
  status: 'compliant' | 'warning' | 'violation';
  score: number;
  lastAssessment: string;
  nextAssessment: string;
  issues: number;
  criticalIssues: number;
}

interface ComplianceViolation {
  id: string;
  region: string;
  regulation: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  detectedAt: string;
  status: 'open' | 'investigating' | 'resolved' | 'dismissed';
  assignee: string;
  dueDate: string;
}

interface DataSubjectRequest {
  id: string;
  type: 'access' | 'deletion' | 'portability' | 'rectification' | 'restriction' | 'objection';
  region: string;
  regulation: string;
  requestedAt: string;
  status: 'pending' | 'processing' | 'completed' | 'rejected';
  dueDate: string;
  requester: string;
  assignee: string;
}

interface ComplianceReport {
  id: string;
  type: 'audit' | 'assessment' | 'incident' | 'dpo';
  title: string;
  region: string;
  regulation: string;
  generatedAt: string;
  status: 'draft' | 'review' | 'approved' | 'published';
  author: string;
  size: string;
}

interface ComplianceAlert {
  id: string;
  type: 'violation' | 'deadline' | 'assessment' | 'system';
  severity: 'info' | 'warning' | 'error' | 'critical';
  title: string;
  description: string;
  region: string;
  regulation: string;
  triggeredAt: string;
  acknowledged: boolean;
  assignee: string;
}

const ComplianceMonitoring: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [complianceStatus, setComplianceStatus] = useState<ComplianceStatus[]>([]);
  const [violations, setViolations] = useState<ComplianceViolation[]>([]);
  const [dataRequests, setDataRequests] = useState<DataSubjectRequest[]>([]);
  const [reports, setReports] = useState<ComplianceReport[]>([]);
  const [alerts, setAlerts] = useState<ComplianceAlert[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [selectedItem, setSelectedItem] = useState<any>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchComplianceData();
  }, []);

  const fetchComplianceData = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 合规状态数据
      setComplianceStatus([
        {
          region: 'EU',
          regulation: 'GDPR',
          status: 'compliant',
          score: 95,
          lastAssessment: '2024-01-10',
          nextAssessment: '2024-04-10',
          issues: 2,
          criticalIssues: 0
        },
        {
          region: 'US',
          regulation: 'CCPA',
          status: 'warning',
          score: 78,
          lastAssessment: '2024-01-08',
          nextAssessment: '2024-04-08',
          issues: 5,
          criticalIssues: 1
        },
        {
          region: 'CA',
          regulation: 'PIPEDA',
          status: 'compliant',
          score: 88,
          lastAssessment: '2024-01-12',
          nextAssessment: '2024-04-12',
          issues: 3,
          criticalIssues: 0
        },
        {
          region: 'CN',
          regulation: 'PIPL',
          status: 'violation',
          score: 65,
          lastAssessment: '2024-01-05',
          nextAssessment: '2024-04-05',
          issues: 8,
          criticalIssues: 2
        }
      ]);

      // 违规记录数据
      setViolations([
        {
          id: 'V001',
          region: 'US',
          regulation: 'CCPA',
          type: 'Data Processing',
          severity: 'critical',
          description: '未经明确同意处理个人信息',
          detectedAt: '2024-01-14 10:30:00',
          status: 'investigating',
          assignee: 'John Smith',
          dueDate: '2024-01-21'
        },
        {
          id: 'V002',
          region: 'CN',
          regulation: 'PIPL',
          type: 'Data Transfer',
          severity: 'high',
          description: '跨境数据传输未完成安全评估',
          detectedAt: '2024-01-13 15:45:00',
          status: 'open',
          assignee: 'Li Wei',
          dueDate: '2024-01-20'
        },
        {
          id: 'V003',
          region: 'CN',
          regulation: 'PIPL',
          type: 'Consent Management',
          severity: 'medium',
          description: '用户同意记录不完整',
          detectedAt: '2024-01-12 09:15:00',
          status: 'resolved',
          assignee: 'Wang Ming',
          dueDate: '2024-01-19'
        }
      ]);

      // 数据主体请求数据
      setDataRequests([
        {
          id: 'DSR001',
          type: 'deletion',
          region: 'EU',
          regulation: 'GDPR',
          requestedAt: '2024-01-14 14:20:00',
          status: 'processing',
          dueDate: '2024-02-13',
          requester: 'user@example.com',
          assignee: 'Sarah Johnson'
        },
        {
          id: 'DSR002',
          type: 'access',
          region: 'US',
          regulation: 'CCPA',
          requestedAt: '2024-01-13 11:30:00',
          status: 'completed',
          dueDate: '2024-02-12',
          requester: 'customer@test.com',
          assignee: 'Mike Davis'
        },
        {
          id: 'DSR003',
          type: 'portability',
          region: 'EU',
          regulation: 'GDPR',
          requestedAt: '2024-01-12 16:45:00',
          status: 'pending',
          dueDate: '2024-02-11',
          requester: 'client@demo.org',
          assignee: 'Emma Wilson'
        }
      ]);

      // 合规报告数据
      setReports([
        {
          id: 'RPT001',
          type: 'audit',
          title: 'GDPR 合规审计报告 Q4 2023',
          region: 'EU',
          regulation: 'GDPR',
          generatedAt: '2024-01-10 10:00:00',
          status: 'approved',
          author: 'Compliance Team',
          size: '2.5 MB'
        },
        {
          id: 'RPT002',
          type: 'assessment',
          title: 'CCPA 风险评估报告',
          region: 'US',
          regulation: 'CCPA',
          generatedAt: '2024-01-08 14:30:00',
          status: 'review',
          author: 'Risk Assessment Team',
          size: '1.8 MB'
        },
        {
          id: 'RPT003',
          type: 'incident',
          title: 'PIPL 违规事件报告',
          region: 'CN',
          regulation: 'PIPL',
          generatedAt: '2024-01-05 09:15:00',
          status: 'draft',
          author: 'Incident Response Team',
          size: '950 KB'
        }
      ]);

      // 合规告警数据
      setAlerts([
        {
          id: 'ALT001',
          type: 'violation',
          severity: 'critical',
          title: 'CCPA 严重违规检测',
          description: '检测到未经授权的个人信息处理活动',
          region: 'US',
          regulation: 'CCPA',
          triggeredAt: '2024-01-14 10:30:00',
          acknowledged: false,
          assignee: 'John Smith'
        },
        {
          id: 'ALT002',
          type: 'deadline',
          severity: 'warning',
          title: 'GDPR 评估截止日期临近',
          description: '下次合规评估将在7天后到期',
          region: 'EU',
          regulation: 'GDPR',
          triggeredAt: '2024-01-14 08:00:00',
          acknowledged: true,
          assignee: 'Sarah Johnson'
        },
        {
          id: 'ALT003',
          type: 'system',
          severity: 'info',
          title: '合规监控系统更新',
          description: '合规监控规则已更新，新增PIPL相关检查项',
          region: 'CN',
          regulation: 'PIPL',
          triggeredAt: '2024-01-13 16:00:00',
          acknowledged: true,
          assignee: 'System Admin'
        }
      ]);
    } catch (error) {
      message.error('获取合规数据失败');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'compliant': return 'green';
      case 'warning': return 'orange';
      case 'violation': return 'red';
      case 'completed': return 'green';
      case 'processing': return 'blue';
      case 'pending': return 'orange';
      case 'open': return 'red';
      case 'investigating': return 'orange';
      case 'resolved': return 'green';
      default: return 'default';
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'yellow';
      case 'low': return 'blue';
      case 'error': return 'red';
      case 'warning': return 'orange';
      case 'info': return 'blue';
      default: return 'default';
    }
  };

  const complianceColumns = [
    {
      title: '区域/法规',
      key: 'regulation',
      render: (record: ComplianceStatus) => (
        <div>
          <div style={{ fontWeight: 600 }}>{record.regulation}</div>
          <div style={{ color: '#666', fontSize: 12 }}>{record.region}</div>
        </div>
      )
    },
    {
      title: '合规状态',
      key: 'status',
      render: (record: ComplianceStatus) => (
        <div>
          <Tag color={getStatusColor(record.status)}>
            {record.status === 'compliant' ? '合规' : 
             record.status === 'warning' ? '警告' : '违规'}
          </Tag>
          <div style={{ marginTop: 4 }}>
            <Progress 
              percent={record.score} 
              size="small"
              status={record.score >= 90 ? 'success' : record.score >= 70 ? 'active' : 'exception'}
            />
          </div>
        </div>
      )
    },
    {
      title: '问题统计',
      key: 'issues',
      render: (record: ComplianceStatus) => (
        <div>
          <div>总问题: {record.issues}</div>
          <div style={{ color: record.criticalIssues > 0 ? '#ff4d4f' : '#52c41a' }}>
            严重问题: {record.criticalIssues}
          </div>
        </div>
      )
    },
    {
      title: '评估时间',
      key: 'assessment',
      render: (record: ComplianceStatus) => (
        <div>
          <div style={{ fontSize: 12 }}>
            上次: {record.lastAssessment}
          </div>
          <div style={{ fontSize: 12, color: '#666' }}>
            下次: {record.nextAssessment}
          </div>
        </div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: ComplianceStatus) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />}
              onClick={() => handleViewDetails(record)}
            />
          </Tooltip>
          <Tooltip title="生成报告">
            <Button 
              type="text" 
              icon={<FileTextOutlined />}
              onClick={() => message.info('生成合规报告')}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const violationColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80
    },
    {
      title: '违规类型',
      key: 'violation',
      render: (record: ComplianceViolation) => (
        <div>
          <div style={{ fontWeight: 600 }}>{record.type}</div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.region} - {record.regulation}
          </div>
        </div>
      )
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => (
        <Tag color={getSeverityColor(severity)}>
          {severity === 'critical' ? '严重' :
           severity === 'high' ? '高' :
           severity === 'medium' ? '中' : '低'}
        </Tag>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status === 'open' ? '待处理' :
           status === 'investigating' ? '调查中' :
           status === 'resolved' ? '已解决' : '已忽略'}
        </Tag>
      )
    },
    {
      title: '负责人',
      dataIndex: 'assignee',
      key: 'assignee'
    },
    {
      title: '截止日期',
      dataIndex: 'dueDate',
      key: 'dueDate',
      render: (date: string) => {
        const isOverdue = moment(date).isBefore(moment());
        return (
          <div style={{ color: isOverdue ? '#ff4d4f' : undefined }}>
            {date}
            {isOverdue && <div style={{ fontSize: 12 }}>已逾期</div>}
          </div>
        );
      }
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: ComplianceViolation) => (
        <Space>
          <Button 
            type="text" 
            icon={<EyeOutlined />}
            onClick={() => handleViewViolation(record)}
          >
            查看
          </Button>
        </Space>
      )
    }
  ];

  const requestColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80
    },
    {
      title: '请求类型',
      key: 'request',
      render: (record: DataSubjectRequest) => (
        <div>
          <div style={{ fontWeight: 600 }}>
            {record.type === 'access' ? '数据访问' :
             record.type === 'deletion' ? '数据删除' :
             record.type === 'portability' ? '数据可携带' :
             record.type === 'rectification' ? '数据更正' :
             record.type === 'restriction' ? '限制处理' : '反对处理'}
          </div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.region} - {record.regulation}
          </div>
        </div>
      )
    },
    {
      title: '申请人',
      dataIndex: 'requester',
      key: 'requester'
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status === 'pending' ? '待处理' :
           status === 'processing' ? '处理中' :
           status === 'completed' ? '已完成' : '已拒绝'}
        </Tag>
      )
    },
    {
      title: '负责人',
      dataIndex: 'assignee',
      key: 'assignee'
    },
    {
      title: '截止日期',
      dataIndex: 'dueDate',
      key: 'dueDate',
      render: (date: string) => {
        const isOverdue = moment(date).isBefore(moment());
        return (
          <div style={{ color: isOverdue ? '#ff4d4f' : undefined }}>
            {date}
            {isOverdue && <div style={{ fontSize: 12 }}>已逾期</div>}
          </div>
        );
      }
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: DataSubjectRequest) => (
        <Space>
          <Button 
            type="text" 
            icon={<EyeOutlined />}
            onClick={() => handleViewRequest(record)}
          >
            查看
          </Button>
        </Space>
      )
    }
  ];

  const reportColumns = [
    {
      title: '报告标题',
      key: 'title',
      render: (record: ComplianceReport) => (
        <div>
          <div style={{ fontWeight: 600 }}>{record.title}</div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.region} - {record.regulation}
          </div>
        </div>
      )
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag>
          {type === 'audit' ? '审计报告' :
           type === 'assessment' ? '评估报告' :
           type === 'incident' ? '事件报告' : 'DPO报告'}
        </Tag>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {status === 'draft' ? '草稿' :
           status === 'review' ? '审核中' :
           status === 'approved' ? '已批准' : '已发布'}
        </Tag>
      )
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author'
    },
    {
      title: '生成时间',
      dataIndex: 'generatedAt',
      key: 'generatedAt',
      render: (date: string) => (
        <div style={{ fontSize: 12 }}>{date}</div>
      )
    },
    {
      title: '文件大小',
      dataIndex: 'size',
      key: 'size'
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: ComplianceReport) => (
        <Space>
          <Button 
            type="text" 
            icon={<EyeOutlined />}
            onClick={() => message.info('预览报告')}
          >
            预览
          </Button>
          <Button 
            type="text" 
            icon={<DownloadOutlined />}
            onClick={() => message.info('下载报告')}
          >
            下载
          </Button>
        </Space>
      )
    }
  ];

  const handleViewDetails = (record: ComplianceStatus) => {
    setSelectedItem(record);
    setModalVisible(true);
  };

  const handleViewViolation = (record: ComplianceViolation) => {
    setSelectedItem(record);
    setModalVisible(true);
  };

  const handleViewRequest = (record: DataSubjectRequest) => {
    setSelectedItem(record);
    setModalVisible(true);
  };

  const handleAcknowledgeAlert = (alertId: string) => {
    setAlerts(alerts.map(alert => 
      alert.id === alertId ? { ...alert, acknowledged: true } : alert
    ));
    message.success('告警已确认');
  };

  const renderOverview = () => (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总体合规分数"
              value={82}
              suffix="/ 100"
              valueStyle={{ color: '#1890ff' }}
              prefix={<ShieldCheckOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃违规"
              value={violations.filter(v => v.status !== 'resolved').length}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<ExclamationCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="待处理请求"
              value={dataRequests.filter(r => r.status === 'pending').length}
              valueStyle={{ color: '#faad14' }}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="未确认告警"
              value={alerts.filter(a => !a.acknowledged).length}
              valueStyle={{ color: '#ff7a45' }}
              prefix={<BellOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={16}>
          <Card title="合规状态概览" extra={<Button icon={<ReloadOutlined />} onClick={fetchComplianceData} />}>
            <Table
              columns={complianceColumns}
              dataSource={complianceStatus}
              rowKey="regulation"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="最新告警" extra={<Badge count={alerts.filter(a => !a.acknowledged).length} />}>
            <List
              dataSource={alerts.slice(0, 5)}
              renderItem={(alert) => (
                <List.Item
                  actions={[
                    !alert.acknowledged && (
                      <Button 
                        type="link" 
                        size="small"
                        onClick={() => handleAcknowledgeAlert(alert.id)}
                      >
                        确认
                      </Button>
                    )
                  ].filter(Boolean)}
                >
                  <List.Item.Meta
                    avatar={
                      <Avatar 
                        icon={<AlertOutlined />} 
                        style={{ 
                          backgroundColor: getSeverityColor(alert.severity) === 'red' ? '#ff4d4f' : 
                                           getSeverityColor(alert.severity) === 'orange' ? '#faad14' : '#1890ff'
                        }} 
                      />
                    }
                    title={
                      <div style={{ fontSize: 12 }}>
                        {alert.title}
                        {!alert.acknowledged && <Badge status="processing" style={{ marginLeft: 8 }} />}
                      </div>
                    }
                    description={
                      <div style={{ fontSize: 11, color: '#666' }}>
                        {alert.region} - {alert.regulation}
                        <br />
                        {moment(alert.triggeredAt).fromNow()}
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );

  return (
    <div>
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane 
          tab={
            <span>
              <ShieldCheckOutlined />
              合规概览
            </span>
          } 
          key="overview"
        >
          {renderOverview()}
        </TabPane>

        <TabPane 
          tab={
            <span>
              <ExclamationCircleOutlined />
              违规管理
              <Badge count={violations.filter(v => v.status !== 'resolved').length} style={{ marginLeft: 8 }} />
            </span>
          } 
          key="violations"
        >
          <Card
            title="违规记录"
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={fetchComplianceData}>
                  刷新
                </Button>
                <Button type="primary" icon={<AlertOutlined />}>
                  创建违规记录
                </Button>
              </Space>
            }
          >
            <Table
              columns={violationColumns}
              dataSource={violations}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 条违规记录`
              }}
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <UserOutlined />
              数据主体请求
              <Badge count={dataRequests.filter(r => r.status === 'pending').length} style={{ marginLeft: 8 }} />
            </span>
          } 
          key="requests"
        >
          <Card
            title="数据主体请求"
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={fetchComplianceData}>
                  刷新
                </Button>
                <Button type="primary" icon={<UserOutlined />}>
                  创建请求
                </Button>
              </Space>
            }
          >
            <Table
              columns={requestColumns}
              dataSource={dataRequests}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 个请求`
              }}
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <FileTextOutlined />
              合规报告
            </span>
          } 
          key="reports"
        >
          <Card
            title="合规报告"
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={fetchComplianceData}>
                  刷新
                </Button>
                <Button type="primary" icon={<FileTextOutlined />}>
                  生成报告
                </Button>
              </Space>
            }
          >
            <Table
              columns={reportColumns}
              dataSource={reports}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 个报告`
              }}
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <BellOutlined />
              告警管理
              <Badge count={alerts.filter(a => !a.acknowledged).length} style={{ marginLeft: 8 }} />
            </span>
          } 
          key="alerts"
        >
          <Card
            title="合规告警"
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={fetchComplianceData}>
                  刷新
                </Button>
                <Button type="primary" icon={<SettingOutlined />}>
                  告警设置
                </Button>
              </Space>
            }
          >
            <List
              dataSource={alerts}
              renderItem={(alert) => (
                <List.Item
                  actions={[
                    <Button 
                      type="link"
                      onClick={() => message.info('查看详情')}
                    >
                      查看
                    </Button>,
                    !alert.acknowledged && (
                      <Button 
                        type="primary"
                        size="small"
                        onClick={() => handleAcknowledgeAlert(alert.id)}
                      >
                        确认
                      </Button>
                    )
                  ].filter(Boolean)}
                >
                  <List.Item.Meta
                    avatar={
                      <Avatar 
                        icon={<AlertOutlined />}
                        style={{ 
                          backgroundColor: getSeverityColor(alert.severity) === 'red' ? '#ff4d4f' : 
                                           getSeverityColor(alert.severity) === 'orange' ? '#faad14' : '#1890ff'
                        }}
                      />
                    }
                    title={
                      <div>
                        {alert.title}
                        <Tag color={getSeverityColor(alert.severity)} style={{ marginLeft: 8 }}>
                          {alert.severity === 'critical' ? '严重' :
                           alert.severity === 'error' ? '错误' :
                           alert.severity === 'warning' ? '警告' : '信息'}
                        </Tag>
                        {!alert.acknowledged && <Badge status="processing" style={{ marginLeft: 8 }} />}
                      </div>
                    }
                    description={
                      <div>
                        <div>{alert.description}</div>
                        <div style={{ marginTop: 4, fontSize: 12, color: '#666' }}>
                          {alert.region} - {alert.regulation} | 负责人: {alert.assignee} | {moment(alert.triggeredAt).fromNow()}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </TabPane>
      </Tabs>

      <Modal
        title="详情"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={800}
      >
        {selectedItem && (
          <Descriptions column={2} bordered>
            {Object.entries(selectedItem).map(([key, value]) => (
              <Descriptions.Item label={key} key={key}>
                {typeof value === 'object' ? JSON.stringify(value) : String(value)}
              </Descriptions.Item>
            ))}
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default ComplianceMonitoring;