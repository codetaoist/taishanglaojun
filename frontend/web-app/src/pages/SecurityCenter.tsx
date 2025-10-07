import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Tabs, Button, Table, Tag, Progress, Alert, Statistic, Space, Typography, Divider, Timeline } from 'antd';
import { 
  SafetyCertificateOutlined, 
  BugOutlined, 
  ExperimentOutlined, 
  BookOutlined, 
  AuditOutlined,
  WarningOutlined,
  SafetyOutlined,
  EyeOutlined,
  SettingOutlined,
  ReloadOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import ThreatDetection from '../components/security/ThreatDetection';
import VulnerabilityManagement from '../components/security/VulnerabilityManagement';
import PenetrationTesting from '../components/security/PenetrationTesting';
import SecurityEducation from '../components/security/SecurityEducation';
import SecurityAudit from '../components/security/SecurityAudit';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

interface SecurityStats {
  threatAlerts: number;
  vulnerabilities: number;
  activeTests: number;
  completedCourses: number;
  auditLogs: number;
  riskScore: number;
}

const SecurityCenter: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [stats, setStats] = useState<SecurityStats>({
    threatAlerts: 0,
    vulnerabilities: 0,
    activeTests: 0,
    completedCourses: 0,
    auditLogs: 0,
    riskScore: 85
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadSecurityStats();
  }, []);

  const loadSecurityStats = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      setTimeout(() => {
        setStats({
          threatAlerts: 12,
          vulnerabilities: 8,
          activeTests: 3,
          completedCourses: 15,
          auditLogs: 1247,
          riskScore: 85
        });
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load security stats:', error);
      setLoading(false);
    }
  };

  const getRiskLevel = (score: number) => {
    if (score >= 90) return { level: 'low', color: 'green', text: '低风险' };
    if (score >= 70) return { level: 'medium', color: 'orange', text: '中等风险' };
    return { level: 'high', color: 'red', text: '高风险' };
  };

  const riskInfo = getRiskLevel(stats.riskScore);

  const renderOverview = () => (
    <div>
      {/* 安全概览卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="威胁告警"
              value={stats.threatAlerts}
              prefix={<WarningOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: stats.threatAlerts > 0 ? '#ff4d4f' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="漏洞数量"
              value={stats.vulnerabilities}
              prefix={<BugOutlined style={{ color: '#fa8c16' }} />}
              valueStyle={{ color: stats.vulnerabilities > 0 ? '#fa8c16' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="活跃测试"
              value={stats.activeTests}
              prefix={<ExperimentOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="完成课程"
              value={stats.completedCourses}
              prefix={<BookOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 安全风险评分 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} md={12}>
          <Card title="安全风险评分" extra={<Button icon={<ReloadOutlined />} onClick={loadSecurityStats} loading={loading}>刷新</Button>}>
            <div style={{ textAlign: 'center' }}>
              <Progress
                type="circle"
                percent={stats.riskScore}
                format={() => `${stats.riskScore}分`}
                strokeColor={riskInfo.color}
                size={120}
              />
              <div style={{ marginTop: 16 }}>
                <Tag color={riskInfo.color} style={{ fontSize: 14, padding: '4px 12px' }}>
                  {riskInfo.text}
                </Tag>
              </div>
              <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
                基于威胁检测、漏洞扫描、合规检查等多维度评估
              </Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="最近安全事件">
            <div style={{ height: 200, overflowY: 'auto' }}>
              <Alert
                message="检测到可疑登录尝试"
                description="来自IP 192.168.1.100的异常登录尝试"
                type="warning"
                showIcon
                style={{ marginBottom: 8 }}
              />
              <Alert
                message="发现SQL注入漏洞"
                description="在用户管理模块发现潜在的SQL注入风险"
                type="error"
                showIcon
                style={{ marginBottom: 8 }}
              />
              <Alert
                message="安全扫描完成"
                description="系统安全扫描已完成，发现3个中等风险漏洞"
                type="info"
                showIcon
                style={{ marginBottom: 8 }}
              />
            </div>
          </Card>
        </Col>
      </Row>

      {/* 快速操作 */}
      <Card title="快速操作">
        <Space wrap>
          <Button type="primary" icon={<SafetyCertificateOutlined />} onClick={() => setActiveTab('threat-detection')}>
                  启动威胁检测
                </Button>
          <Button icon={<BugOutlined />} onClick={() => setActiveTab('vulnerability')}>
            漏洞扫描
          </Button>
          <Button icon={<ExperimentOutlined />} onClick={() => setActiveTab('pentest')}>
            渗透测试
          </Button>
          <Button icon={<BookOutlined />} onClick={() => setActiveTab('education')}>
            安全培训
          </Button>
          <Button icon={<AuditOutlined />} onClick={() => setActiveTab('audit')}>
            安全审计
          </Button>
        </Space>
      </Card>
    </div>
  );

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>
          <SafetyOutlined style={{ marginRight: 8 }} />
          安全控制台
        </Title>
        <Text type="secondary">
          全面的安全管理平台，提供威胁检测、漏洞管理、渗透测试、安全教育和审计功能
        </Text>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab} size="large">
        <TabPane
          tab={
            <span>
              <EyeOutlined />
              安全概览
            </span>
          }
          key="overview"
        >
          {renderOverview()}
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <SafetyCertificateOutlined />
              威胁检测
            </span>
          }
          key="threat-detection"
        >
          <ThreatDetection />
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <BugOutlined />
              漏洞管理
            </span>
          }
          key="vulnerability"
        >
          <VulnerabilityManagement />
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <ExperimentOutlined />
              渗透测试
            </span>
          }
          key="pentest"
        >
          <PenetrationTesting />
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <BookOutlined />
              安全教育
            </span>
          }
          key="education"
        >
          <SecurityEducation />
        </TabPane>
        
        <TabPane
          tab={
            <span>
              <AuditOutlined />
              安全审计
            </span>
          }
          key="audit"
        >
          <SecurityAudit />
        </TabPane>
      </Tabs>
    </div>
  );
};

export default SecurityCenter;