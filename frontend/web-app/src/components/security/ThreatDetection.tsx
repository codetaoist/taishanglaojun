import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Tag, Space, Modal, Form, Input, Select, DatePicker, Row, Col, Statistic, Alert } from 'antd';
import { SafetyCertificateOutlined, WarningOutlined, StopOutlined, PlayCircleOutlined, EyeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Option } = Select;
const { RangePicker } = DatePicker;

interface ThreatAlert {
  id: string;
  title: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  category: string;
  sourceIP: string;
  targetIP: string;
  status: 'open' | 'investigating' | 'resolved' | 'false_positive';
  createdAt: string;
}

interface DetectionRule {
  id: string;
  name: string;
  category: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  enabled: boolean;
  description: string;
}

const ThreatDetection: React.FC = () => {
  const [alerts, setAlerts] = useState<ThreatAlert[]>([]);
  const [rules, setRules] = useState<DetectionRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [selectedAlert, setSelectedAlert] = useState<ThreatAlert | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadThreatAlerts();
    loadDetectionRules();
  }, []);

  const loadThreatAlerts = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      setTimeout(() => {
        setAlerts([
          {
            id: '1',
            title: '可疑登录尝试',
            severity: 'high',
            category: '身份验证',
            sourceIP: '192.168.1.100',
            targetIP: '10.0.0.1',
            status: 'open',
            createdAt: '2024-01-15 10:30:00'
          },
          {
            id: '2',
            title: 'SQL注入攻击',
            severity: 'critical',
            category: 'Web攻击',
            sourceIP: '203.0.113.1',
            targetIP: '10.0.0.2',
            status: 'investigating',
            createdAt: '2024-01-15 09:15:00'
          },
          {
            id: '3',
            title: '异常网络流量',
            severity: 'medium',
            category: '网络异常',
            sourceIP: '198.51.100.1',
            targetIP: '10.0.0.3',
            status: 'resolved',
            createdAt: '2024-01-14 16:45:00'
          }
        ]);
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load threat alerts:', error);
      setLoading(false);
    }
  };

  const loadDetectionRules = async () => {
    try {
      // 模拟API调用
      setTimeout(() => {
        setRules([
          {
            id: '1',
            name: '暴力破解检测',
            category: '身份验证',
            severity: 'high',
            enabled: true,
            description: '检测短时间内多次失败的登录尝试'
          },
          {
            id: '2',
            name: 'SQL注入检测',
            category: 'Web攻击',
            severity: 'critical',
            enabled: true,
            description: '检测SQL注入攻击模式'
          },
          {
            id: '3',
            name: 'XSS攻击检测',
            category: 'Web攻击',
            severity: 'high',
            enabled: false,
            description: '检测跨站脚本攻击'
          }
        ]);
      }, 500);
    } catch (error) {
      console.error('Failed to load detection rules:', error);
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
      case 'open': return 'red';
      case 'investigating': return 'orange';
      case 'resolved': return 'green';
      case 'false_positive': return 'gray';
      default: return 'default';
    }
  };

  const alertColumns: ColumnsType<ThreatAlert> = [
    {
      title: '告警标题',
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
      title: '源IP',
      dataIndex: 'sourceIP',
      key: 'sourceIP',
    },
    {
      title: '目标IP',
      dataIndex: 'targetIP',
      key: 'targetIP',
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
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
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
              setSelectedAlert(record);
              setModalVisible(true);
            }}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  const ruleColumns: ColumnsType<DetectionRule> = [
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
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
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled) => (
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={record.enabled ? <StopOutlined /> : <PlayCircleOutlined />}
          >
            {record.enabled ? '禁用' : '启用'}
          </Button>
          <Button type="link">编辑</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 威胁检测统计 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="今日告警"
              value={12}
              prefix={<WarningOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="活跃规则"
              value={8}
              prefix={<SafetyCertificateOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="阻止攻击"
              value={156}
              prefix={<StopOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 威胁告警列表 */}
      <Card 
        title="威胁告警" 
        extra={
          <Space>
            <Button type="primary" onClick={loadThreatAlerts}>
              刷新
            </Button>
            <Button>导出</Button>
          </Space>
        }
        style={{ marginBottom: 24 }}
      >
        <Table
          columns={alertColumns}
          dataSource={alerts}
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

      {/* 检测规则列表 */}
      <Card 
        title="检测规则" 
        extra={
          <Space>
            <Button type="primary">新增规则</Button>
            <Button>批量操作</Button>
          </Space>
        }
      >
        <Table
          columns={ruleColumns}
          dataSource={rules}
          rowKey="id"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 告警详情模态框 */}
      <Modal
        title="告警详情"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setModalVisible(false)}>
            关闭
          </Button>,
          <Button key="resolve" type="primary">
            标记为已解决
          </Button>,
        ]}
        width={800}
      >
        {selectedAlert && (
          <div>
            <Alert
              message={selectedAlert.title}
              description={`来源IP: ${selectedAlert.sourceIP} → 目标IP: ${selectedAlert.targetIP}`}
              type={selectedAlert.severity === 'critical' ? 'error' : 'warning'}
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <strong>严重程度:</strong> 
                <Tag color={getSeverityColor(selectedAlert.severity)} style={{ marginLeft: 8 }}>
                  {selectedAlert.severity.toUpperCase()}
                </Tag>
              </Col>
              <Col span={12}>
                <strong>类别:</strong> {selectedAlert.category}
              </Col>
              <Col span={12}>
                <strong>状态:</strong> 
                <Tag color={getStatusColor(selectedAlert.status)} style={{ marginLeft: 8 }}>
                  {selectedAlert.status === 'open' ? '待处理' :
                   selectedAlert.status === 'investigating' ? '调查中' :
                   selectedAlert.status === 'resolved' ? '已解决' : '误报'}
                </Tag>
              </Col>
              <Col span={12}>
                <strong>创建时间:</strong> {selectedAlert.createdAt}
              </Col>
            </Row>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ThreatDetection;