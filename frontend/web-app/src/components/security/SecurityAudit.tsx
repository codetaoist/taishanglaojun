import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Tag, Space, Modal, Form, Input, Select, DatePicker, Row, Col, Statistic, Descriptions, Timeline, Tabs } from 'antd';
import { AuditOutlined, FileTextOutlined, CheckCircleOutlined, ExclamationCircleOutlined, EyeOutlined, DownloadOutlined, SearchOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Option } = Select;
const { RangePicker } = DatePicker;
const { TabPane } = Tabs;

interface AuditLog {
  id: string;
  timestamp: string;
  userId: string;
  userName: string;
  action: string;
  resource: string;
  ipAddress: string;
  userAgent: string;
  status: 'success' | 'failed' | 'warning';
  details: string;
}

interface SecurityEvent {
  id: string;
  eventType: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  source: string;
  target: string;
  description: string;
  timestamp: string;
  status: 'open' | 'investigating' | 'resolved' | 'false_positive';
  assignee?: string;
}

interface ComplianceReport {
  id: string;
  name: string;
  framework: 'ISO27001' | 'SOX' | 'GDPR' | 'PCI-DSS' | 'HIPAA';
  reportDate: string;
  complianceScore: number;
  totalControls: number;
  passedControls: number;
  failedControls: number;
  status: 'draft' | 'final' | 'approved';
  auditor: string;
}

const SecurityAudit: React.FC = () => {
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [securityEvents, setSecurityEvents] = useState<SecurityEvent[]>([]);
  const [complianceReports, setComplianceReports] = useState<ComplianceReport[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [eventModalVisible, setEventModalVisible] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<SecurityEvent | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadAuditLogs();
    loadSecurityEvents();
    loadComplianceReports();
  }, []);

  const loadAuditLogs = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      setTimeout(() => {
        setAuditLogs([
          {
            id: '1',
            timestamp: '2024-01-15 10:30:25',
            userId: 'user001',
            userName: '张三',
            action: '登录系统',
            resource: '/auth/login',
            ipAddress: '192.168.1.100',
            userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
            status: 'success',
            details: '用户成功登录系统'
          },
          {
            id: '2',
            timestamp: '2024-01-15 10:25:15',
            userId: 'user002',
            userName: '李四',
            action: '访问敏感数据',
            resource: '/api/users/sensitive',
            ipAddress: '192.168.1.101',
            userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
            status: 'failed',
            details: '权限不足，访问被拒绝'
          },
          {
            id: '3',
            timestamp: '2024-01-15 10:20:45',
            userId: 'admin001',
            userName: '管理员',
            action: '修改用户权限',
            resource: '/admin/users/permissions',
            ipAddress: '192.168.1.10',
            userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
            status: 'success',
            details: '修改用户user003的权限级别'
          }
        ]);
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load audit logs:', error);
      setLoading(false);
    }
  };

  const loadSecurityEvents = async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        setSecurityEvents([
          {
            id: '1',
            eventType: '异常登录',
            severity: 'high',
            source: '192.168.1.200',
            target: 'auth-service',
            description: '检测到来自异常地理位置的登录尝试',
            timestamp: '2024-01-15 09:45:30',
            status: 'investigating',
            assignee: '安全团队'
          },
          {
            id: '2',
            eventType: '权限提升',
            severity: 'critical',
            source: 'user005',
            target: 'admin-panel',
            description: '用户尝试未授权的权限提升操作',
            timestamp: '2024-01-15 08:30:15',
            status: 'open'
          },
          {
            id: '3',
            eventType: '数据泄露',
            severity: 'critical',
            source: 'internal-system',
            target: 'database',
            description: '检测到大量敏感数据被异常访问',
            timestamp: '2024-01-14 22:15:45',
            status: 'resolved',
            assignee: '数据保护官'
          }
        ]);
      }, 500);
    } catch (error) {
      console.error('Failed to load security events:', error);
    }
  };

  const loadComplianceReports = async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        setComplianceReports([
          {
            id: '1',
            name: '2024年第一季度ISO27001合规报告',
            framework: 'ISO27001',
            reportDate: '2024-01-15',
            complianceScore: 85,
            totalControls: 114,
            passedControls: 97,
            failedControls: 17,
            status: 'final',
            auditor: '外部审计师'
          },
          {
            id: '2',
            name: 'GDPR数据保护合规评估',
            framework: 'GDPR',
            reportDate: '2024-01-10',
            complianceScore: 92,
            totalControls: 25,
            passedControls: 23,
            failedControls: 2,
            status: 'approved',
            auditor: '数据保护官'
          },
          {
            id: '3',
            name: 'PCI-DSS支付安全合规检查',
            framework: 'PCI-DSS',
            reportDate: '2024-01-08',
            complianceScore: 78,
            totalControls: 12,
            passedControls: 9,
            failedControls: 3,
            status: 'draft',
            auditor: '内部审计师'
          }
        ]);
      }, 300);
    } catch (error) {
      console.error('Failed to load compliance reports:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success': case 'resolved': case 'approved': return 'green';
      case 'failed': case 'open': return 'red';
      case 'warning': case 'investigating': case 'draft': return 'orange';
      case 'false_positive': case 'final': return 'blue';
      default: return 'default';
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

  const auditLogColumns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 150,
    },
    {
      title: '用户',
      dataIndex: 'userName',
      key: 'userName',
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
    },
    {
      title: 'IP地址',
      dataIndex: 'ipAddress',
      key: 'ipAddress',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'success' ? '成功' :
           status === 'failed' ? '失败' : '警告'}
        </Tag>
      ),
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
              setSelectedLog(record);
              setModalVisible(true);
            }}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  const securityEventColumns: ColumnsType<SecurityEvent> = [
    {
      title: '事件类型',
      dataIndex: 'eventType',
      key: 'eventType',
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
      title: '来源',
      dataIndex: 'source',
      key: 'source',
    },
    {
      title: '目标',
      dataIndex: 'target',
      key: 'target',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'open' ? '待处理' :
           status === 'investigating' ? '调查中' :
           status === 'resolved' ? '已解决' : '误报'}
        </Tag>
      ),
    },
    {
      title: '负责人',
      dataIndex: 'assignee',
      key: 'assignee',
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
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
              setSelectedEvent(record);
              setEventModalVisible(true);
            }}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  const complianceColumns: ColumnsType<ComplianceReport> = [
    {
      title: '报告名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '合规框架',
      dataIndex: 'framework',
      key: 'framework',
      render: (framework) => (
        <Tag color="blue">{framework}</Tag>
      ),
    },
    {
      title: '合规分数',
      dataIndex: 'complianceScore',
      key: 'complianceScore',
      render: (score) => (
        <span style={{ color: score >= 80 ? '#52c41a' : score >= 60 ? '#fa8c16' : '#ff4d4f' }}>
          {score}%
        </span>
      ),
    },
    {
      title: '通过/总计',
      key: 'controls',
      render: (_, record) => (
        <span>
          {record.passedControls}/{record.totalControls}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'draft' ? '草稿' :
           status === 'final' ? '最终' : '已批准'}
        </Tag>
      ),
    },
    {
      title: '审计师',
      dataIndex: 'auditor',
      key: 'auditor',
    },
    {
      title: '报告日期',
      dataIndex: 'reportDate',
      key: 'reportDate',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button type="link" icon={<EyeOutlined />}>查看</Button>
          <Button type="link" icon={<DownloadOutlined />}>下载</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 安全审计统计 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="今日审计日志"
              value={1247}
              prefix={<AuditOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="安全事件"
              value={8}
              prefix={<ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="合规报告"
              value={12}
              prefix={<FileTextOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="平均合规分数"
              value={85}
              suffix="%"
              prefix={<CheckCircleOutlined style={{ color: '#722ed1' }} />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Tabs defaultActiveKey="auditLogs">
        <TabPane tab="审计日志" key="auditLogs">
          <Card 
            title="系统审计日志" 
            extra={
              <Space>
                <Button icon={<SearchOutlined />}>高级搜索</Button>
                <Button icon={<DownloadOutlined />}>导出日志</Button>
                <Button onClick={loadAuditLogs}>刷新</Button>
              </Space>
            }
          >
            <Table
              columns={auditLogColumns}
              dataSource={auditLogs}
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
        </TabPane>

        <TabPane tab="安全事件" key="securityEvents">
          <Card 
            title="安全事件管理" 
            extra={
              <Space>
                <Button type="primary">新建事件</Button>
                <Button onClick={loadSecurityEvents}>刷新</Button>
              </Space>
            }
          >
            <Table
              columns={securityEventColumns}
              dataSource={securityEvents}
              rowKey="id"
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`,
              }}
            />
          </Card>
        </TabPane>

        <TabPane tab="合规报告" key="compliance">
          <Card 
            title="合规性报告" 
            extra={
              <Space>
                <Button type="primary" icon={<FileTextOutlined />}>
                  生成报告
                </Button>
                <Button onClick={loadComplianceReports}>刷新</Button>
              </Space>
            }
          >
            <Table
              columns={complianceColumns}
              dataSource={complianceReports}
              rowKey="id"
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条记录`,
              }}
            />
          </Card>
        </TabPane>
      </Tabs>

      {/* 审计日志详情模态框 */}
      <Modal
        title="审计日志详情"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={700}
      >
        {selectedLog && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="时间戳">
                {selectedLog.timestamp}
              </Descriptions.Item>
              <Descriptions.Item label="用户ID">
                {selectedLog.userId}
              </Descriptions.Item>
              <Descriptions.Item label="用户名">
                {selectedLog.userName}
              </Descriptions.Item>
              <Descriptions.Item label="操作">
                {selectedLog.action}
              </Descriptions.Item>
              <Descriptions.Item label="资源">
                {selectedLog.resource}
              </Descriptions.Item>
              <Descriptions.Item label="IP地址">
                {selectedLog.ipAddress}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedLog.status)}>
                  {selectedLog.status === 'success' ? '成功' :
                   selectedLog.status === 'failed' ? '失败' : '警告'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="User Agent" span={2}>
                {selectedLog.userAgent}
              </Descriptions.Item>
              <Descriptions.Item label="详细信息" span={2}>
                {selectedLog.details}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>

      {/* 安全事件详情模态框 */}
      <Modal
        title="安全事件详情"
        open={eventModalVisible}
        onCancel={() => setEventModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setEventModalVisible(false)}>
            关闭
          </Button>,
          <Button key="assign" type="primary">
            分配处理人
          </Button>,
        ]}
        width={700}
      >
        {selectedEvent && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="事件类型">
                {selectedEvent.eventType}
              </Descriptions.Item>
              <Descriptions.Item label="严重程度">
                <Tag color={getSeverityColor(selectedEvent.severity)}>
                  {selectedEvent.severity.toUpperCase()}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="来源">
                {selectedEvent.source}
              </Descriptions.Item>
              <Descriptions.Item label="目标">
                {selectedEvent.target}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedEvent.status)}>
                  {selectedEvent.status === 'open' ? '待处理' :
                   selectedEvent.status === 'investigating' ? '调查中' :
                   selectedEvent.status === 'resolved' ? '已解决' : '误报'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="负责人">
                {selectedEvent.assignee || '未分配'}
              </Descriptions.Item>
              <Descriptions.Item label="发生时间">
                {selectedEvent.timestamp}
              </Descriptions.Item>
              <Descriptions.Item label="事件描述" span={2}>
                {selectedEvent.description}
              </Descriptions.Item>
            </Descriptions>

            <Card title="处理时间线" size="small" style={{ marginTop: 16 }}>
              <Timeline>
                <Timeline.Item color="blue">
                  {selectedEvent.timestamp} - 事件被检测到
                </Timeline.Item>
                {selectedEvent.status !== 'open' && (
                  <Timeline.Item color="orange">
                    {selectedEvent.timestamp} - 开始调查处理
                  </Timeline.Item>
                )}
                {selectedEvent.status === 'resolved' && (
                  <Timeline.Item color="green">
                    {selectedEvent.timestamp} - 事件已解决
                  </Timeline.Item>
                )}
              </Timeline>
            </Card>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default SecurityAudit;