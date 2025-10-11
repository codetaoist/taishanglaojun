import React, { useState, useEffect } from 'react';
import { 
  Card, Row, Col, Tabs, Button, Table, Tag, Progress, Alert, Statistic, Space, Typography, 
  Divider, Timeline, Badge, Tooltip, Spin, List, Avatar, Switch, Drawer, notification,
  Carousel, Empty, Descriptions, Steps, Rate, Popover, Dropdown, Menu
} from 'antd';
import { 
  SafetyCertificateOutlined, 
  BugOutlined, 
  ExperimentOutlined, 
  BookOutlined, 
  AuditOutlined,
  WarningOutlined,
  EyeOutlined,
  SettingOutlined,
  ReloadOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  DashboardOutlined,
  FireOutlined,
  RadarChartOutlined,
  MonitorOutlined,
  AlertOutlined,
  LockOutlined,
  UnlockOutlined,
  ScanOutlined,
  GlobalOutlined,
  UserOutlined,
  TeamOutlined,
  HistoryOutlined,
  BarChartOutlined,
  LineChartOutlined,
  PieChartOutlined,
  ThunderboltOutlined,
  RocketOutlined,
  BellOutlined,
  MoreOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  FullscreenOutlined,
  DownloadOutlined,
  ShareAltOutlined,
  FilterOutlined,
  SearchOutlined,
  SyncOutlined,
  CloudOutlined,
  DatabaseOutlined,
  ApiOutlined,
  WifiOutlined,
  MobileOutlined,
  DesktopOutlined,
  TabletOutlined,
  CameraOutlined,
  VideoCameraOutlined,
  SoundOutlined,
  FileProtectOutlined,
  KeyOutlined,
  CrownOutlined,
  StarOutlined,
  HeartOutlined,
  LikeOutlined,
  DislikeOutlined,
  CommentOutlined,
  MessageOutlined,
  MailOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  CalendarOutlined,
  ClockCircleFilled,
  CheckCircleFilled,
  ExclamationCircleFilled,
  CloseCircleFilled,
  InfoCircleFilled,
  QuestionCircleFilled,
  PlusOutlined,
  MinusOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  ScissorOutlined,
  HighlightOutlined,
  FontSizeOutlined,
  BoldOutlined,
  ItalicOutlined,
  UnderlineOutlined,
  StrikethroughOutlined,
  AlignLeftOutlined,
  AlignCenterOutlined,
  AlignRightOutlined,
  OrderedListOutlined,
  UnorderedListOutlined,
  LinkOutlined,
  PictureOutlined,
  FileOutlined,
  FolderOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  RotateLeftOutlined,
  RotateRightOutlined,
  SwapOutlined,
  RollbackOutlined,
  ForwardOutlined,
  VerticalAlignTopOutlined,
  VerticalAlignBottomOutlined,
  MenuOutlined,
  AppstoreOutlined,
  BarsOutlined,
  BorderOutlined,
  TableOutlined,
  PartitionOutlined,
  LayoutOutlined,
  SlidersOutlined,
  ControlOutlined,
  BranchesOutlined,
  NodeIndexOutlined,
  SelectOutlined,
  GatewayOutlined,
  DeploymentUnitOutlined,
  ClusterOutlined,
  GroupOutlined,
  SendOutlined,
  ImportOutlined,
  ExportOutlined,
  InboxOutlined,
  RedoOutlined,
  UndoOutlined,
  CaretRightOutlined,
  CaretLeftOutlined,
  CaretUpOutlined,
  CaretDownOutlined,
  UpOutlined,
  DownOutlined,
  LeftOutlined,
  RightOutlined,
  DoubleRightOutlined,
  DoubleLeftOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  ArrowLeftOutlined,
  ArrowRightOutlined,
  ExpandOutlined,
  ExpandAltOutlined,
  ShrinkOutlined,
  ArrowsAltOutlined,
  ColumnWidthOutlined,
  ColumnHeightOutlined,
  AreaChartOutlined,
  DotChartOutlined,
  FundOutlined,
  SlackOutlined,
  BehanceOutlined,
  DribbbleOutlined,
  InstagramOutlined,
  YuqueOutlined,
  AlibabaOutlined,
  YahooOutlined,
  RedditOutlined,
  SkypeOutlined,
  CodeSandboxOutlined,
  ChromeOutlined,
  AmazonOutlined,
  CodepenOutlined,
  AlipayOutlined,
  AntDesignOutlined,
  AntCloudOutlined,
  AliyunOutlined,
  ZhihuOutlined,
  SlackSquareOutlined,
  BehanceSquareOutlined,
  DribbbleSquareOutlined,
  ChromeFilled
} from '@ant-design/icons';
import ThreatDetection from '../components/security/ThreatDetection';
import VulnerabilityManagement from '../components/security/VulnerabilityManagement';
import PenetrationTesting from '../components/security/PenetrationTesting';
import SecurityEducation from '../components/security/SecurityEducation';
import SecurityAudit from '../components/security/SecurityAudit';

const { Title, Text } = Typography;

interface SecurityStats {
  threatAlerts: number;
  vulnerabilities: number;
  activeTests: number;
  completedCourses: number;
  auditLogs: number;
  riskScore: number;
  blockedAttacks: number;
  activeConnections: number;
  systemUptime: number;
  lastScanTime: string;
  criticalAlerts: number;
  resolvedIncidents: number;
  securityEvents24h: number;
  complianceScore: number;
  firewallStatus: 'active' | 'inactive' | 'warning';
  antivirusStatus: 'active' | 'inactive' | 'warning';
  intrusionDetection: 'active' | 'inactive' | 'warning';
  dataEncryption: 'active' | 'inactive' | 'warning';
}

interface SecurityEvent {
  id: string;
  type: 'threat' | 'vulnerability' | 'audit' | 'compliance' | 'system';
  severity: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  timestamp: string;
  source: string;
  status: 'new' | 'investigating' | 'resolved' | 'ignored';
  assignee?: string;
}

interface SystemHealth {
  cpu: number;
  memory: number;
  disk: number;
  network: number;
  temperature: number;
  status: 'healthy' | 'warning' | 'critical';
}

const SecurityCenter: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [stats, setStats] = useState<SecurityStats>({
    threatAlerts: 0,
    vulnerabilities: 0,
    activeTests: 0,
    completedCourses: 0,
    auditLogs: 0,
    riskScore: 85,
    blockedAttacks: 0,
    activeConnections: 0,
    systemUptime: 0,
    lastScanTime: '',
    criticalAlerts: 0,
    resolvedIncidents: 0,
    securityEvents24h: 0,
    complianceScore: 0,
    firewallStatus: 'active',
    antivirusStatus: 'active',
    intrusionDetection: 'active',
    dataEncryption: 'active'
  });
  const [loading, setLoading] = useState(false);
  const [realTimeMode, setRealTimeMode] = useState(true);
  const [securityEvents, setSecurityEvents] = useState<SecurityEvent[]>([]);
  const [systemHealth, setSystemHealth] = useState<SystemHealth>({
    cpu: 0,
    memory: 0,
    disk: 0,
    network: 0,
    temperature: 0,
    status: 'healthy'
  });
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<SecurityEvent | null>(null);

  useEffect(() => {
    loadSecurityStats();
    loadSecurityEvents();
    loadSystemHealth();
    
    // 实时监控模式下定时更新数据
    let interval: NodeJS.Timeout;
    if (realTimeMode) {
      interval = setInterval(() => {
        loadSecurityStats();
        loadSystemHealth();
        if (Math.random() > 0.7) { // 30%概率生成新事件
          generateRandomEvent();
        }
      }, 5000); // 每5秒更新一次
    }
    
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [realTimeMode]);

  const loadSecurityStats = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      setTimeout(() => {
        const now = new Date();
        setStats({
          threatAlerts: Math.floor(Math.random() * 20) + 5,
          vulnerabilities: Math.floor(Math.random() * 15) + 3,
          activeTests: Math.floor(Math.random() * 8) + 1,
          completedCourses: Math.floor(Math.random() * 30) + 10,
          auditLogs: Math.floor(Math.random() * 2000) + 1000,
          riskScore: Math.floor(Math.random() * 30) + 70,
          blockedAttacks: Math.floor(Math.random() * 100) + 50,
          activeConnections: Math.floor(Math.random() * 500) + 200,
          systemUptime: Math.floor(Math.random() * 720) + 24, // 1-30天
          lastScanTime: now.toLocaleString(),
          criticalAlerts: Math.floor(Math.random() * 5),
          resolvedIncidents: Math.floor(Math.random() * 50) + 20,
          securityEvents24h: Math.floor(Math.random() * 200) + 100,
          complianceScore: Math.floor(Math.random() * 20) + 80,
          firewallStatus: Math.random() > 0.9 ? 'warning' : 'active',
          antivirusStatus: Math.random() > 0.95 ? 'warning' : 'active',
          intrusionDetection: Math.random() > 0.9 ? 'warning' : 'active',
          dataEncryption: 'active'
        });
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load security stats:', error);
      setLoading(false);
    }
  };

  const loadSecurityEvents = async () => {
    try {
      const events: SecurityEvent[] = [
        {
          id: '1',
          type: 'threat',
          severity: 'high',
          title: '检测到可疑登录尝试',
          description: '来自IP 192.168.1.100的异常登录尝试，已连续失败5次',
          timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
          source: '身份验证系统',
          status: 'investigating',
          assignee: '安全团队'
        },
        {
          id: '2',
          type: 'vulnerability',
          severity: 'critical',
          title: '发现SQL注入漏洞',
          description: '在用户管理模块发现潜在的SQL注入风险，需要立即修复',
          timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
          source: '漏洞扫描器',
          status: 'new'
        },
        {
          id: '3',
          type: 'system',
          severity: 'medium',
          title: '系统性能异常',
          description: 'CPU使用率持续超过80%，可能影响系统稳定性',
          timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
          source: '系统监控',
          status: 'resolved',
          assignee: '运维团队'
        }
      ];
      setSecurityEvents(events);
    } catch (error) {
      console.error('Failed to load security events:', error);
    }
  };

  const loadSystemHealth = async () => {
    try {
      const health: SystemHealth = {
        cpu: Math.floor(Math.random() * 40) + 30,
        memory: Math.floor(Math.random() * 30) + 50,
        disk: Math.floor(Math.random() * 20) + 60,
        network: Math.floor(Math.random() * 50) + 20,
        temperature: Math.floor(Math.random() * 20) + 45,
        status: 'healthy'
      };
      
      // 根据指标确定系统状态
      if (health.cpu > 80 || health.memory > 85 || health.temperature > 70) {
        health.status = 'critical';
      } else if (health.cpu > 60 || health.memory > 70 || health.temperature > 60) {
        health.status = 'warning';
      }
      
      setSystemHealth(health);
    } catch (error) {
      console.error('Failed to load system health:', error);
    }
  };

  const generateRandomEvent = () => {
    const eventTypes = ['threat', 'vulnerability', 'audit', 'compliance', 'system'] as const;
    const severities = ['low', 'medium', 'high', 'critical'] as const;
    const titles = [
      '检测到异常网络流量',
      '发现新的安全漏洞',
      '系统配置变更',
      '用户权限异常',
      '文件完整性检查失败',
      '防火墙规则触发',
      '恶意软件检测',
      '数据访问异常'
    ];
    
    const newEvent: SecurityEvent = {
      id: Date.now().toString(),
      type: eventTypes[Math.floor(Math.random() * eventTypes.length)],
      severity: severities[Math.floor(Math.random() * severities.length)],
      title: titles[Math.floor(Math.random() * titles.length)],
      description: '系统自动检测到的安全事件，需要进一步分析',
      timestamp: new Date().toISOString(),
      source: '实时监控系统',
      status: 'new'
    };
    
    setSecurityEvents(prev => [newEvent, ...prev.slice(0, 9)]); // 保持最新10条
    
    // 显示通知
    if (newEvent.severity === 'critical' || newEvent.severity === 'high') {
      notification.warning({
        message: '安全警报',
        description: newEvent.title,
        placement: 'topRight',
        duration: 5
      });
    }
  };

  const getRiskLevel = (score: number) => {
    if (score >= 90) return { level: 'low', color: 'green', text: '低风险' };
    if (score >= 70) return { level: 'medium', color: 'orange', text: '中等风险' };
    return { level: 'high', color: 'red', text: '高风险' };
  };

  const getStatusColor = (status: 'active' | 'inactive' | 'warning') => {
    switch (status) {
      case 'active': return 'green';
      case 'warning': return 'orange';
      case 'inactive': return 'red';
      default: return 'gray';
    }
  };

  const getStatusIcon = (status: 'active' | 'inactive' | 'warning') => {
    switch (status) {
      case 'active': return <CheckCircleFilled style={{ color: 'green' }} />;
      case 'warning': return <ExclamationCircleFilled style={{ color: 'orange' }} />;
      case 'inactive': return <CloseCircleFilled style={{ color: 'red' }} />;
      default: return <QuestionCircleFilled style={{ color: 'gray' }} />;
    }
  };

  const getSeverityColor = (severity: 'low' | 'medium' | 'high' | 'critical') => {
    switch (severity) {
      case 'low': return 'blue';
      case 'medium': return 'orange';
      case 'high': return 'red';
      case 'critical': return 'purple';
      default: return 'gray';
    }
  };

  const getEventTypeIcon = (type: 'threat' | 'vulnerability' | 'audit' | 'compliance' | 'system') => {
    switch (type) {
      case 'threat': return <FireOutlined />;
      case 'vulnerability': return <BugOutlined />;
      case 'audit': return <AuditOutlined />;
      case 'compliance': return <SafetyCertificateOutlined />;
      case 'system': return <MonitorOutlined />;
      default: return <InfoCircleFilled />;
    }
  };

  const formatUptime = (hours: number) => {
    const days = Math.floor(hours / 24);
    const remainingHours = hours % 24;
    return `${days}天 ${remainingHours}小时`;
  };

  const riskInfo = getRiskLevel(stats.riskScore);

  const renderOverview = () => (
    <div>
      {/* 实时监控控制面板 */}
      <Card 
        style={{ marginBottom: 24 }}
        bodyStyle={{ padding: '16px 24px' }}
      >
        <Row justify="space-between" align="middle">
          <Col>
            <Space size="large">
              <div>
                <Badge 
                  status={realTimeMode ? "processing" : "default"} 
                  text={realTimeMode ? "实时监控已启用" : "实时监控已关闭"}
                />
              </div>
              <div>
                <Text type="secondary">最后更新: {stats.lastScanTime}</Text>
              </div>
            </Space>
          </Col>
          <Col>
            <Space>
              <Tooltip title="切换实时监控模式">
                <Switch
                  checked={realTimeMode}
                  onChange={setRealTimeMode}
                  checkedChildren={<MonitorOutlined />}
                  unCheckedChildren={<PauseCircleOutlined />}
                />
              </Tooltip>
              <Button 
                icon={<ReloadOutlined />} 
                onClick={loadSecurityStats} 
                loading={loading}
                type="primary"
              >
                刷新数据
              </Button>
              <Dropdown
                overlay={
                  <Menu>
                    <Menu.Item key="export" icon={<DownloadOutlined />}>
                      导出报告
                    </Menu.Item>
                    <Menu.Item key="share" icon={<ShareAltOutlined />}>
                      分享仪表板
                    </Menu.Item>
                    <Menu.Item key="settings" icon={<SettingOutlined />}>
                      监控设置
                    </Menu.Item>
                  </Menu>
                }
              >
                <Button icon={<MoreOutlined />} />
              </Dropdown>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 核心安全指标 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable>
            <Statistic
              title="威胁告警"
              value={stats.threatAlerts}
              prefix={<WarningOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: stats.threatAlerts > 0 ? '#ff4d4f' : '#52c41a' }}
              suffix={
                <Tooltip title="24小时内检测到的威胁数量">
                  <Badge 
                    count={stats.criticalAlerts} 
                    style={{ backgroundColor: '#ff4d4f' }}
                    size="small"
                  />
                </Tooltip>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable>
            <Statistic
              title="漏洞数量"
              value={stats.vulnerabilities}
              prefix={<BugOutlined style={{ color: '#fa8c16' }} />}
              valueStyle={{ color: stats.vulnerabilities > 0 ? '#fa8c16' : '#52c41a' }}
              suffix={
                <Tooltip title="已发现但未修复的漏洞">
                  <Tag color="orange" size="small">待修复</Tag>
                </Tooltip>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable>
            <Statistic
              title="已阻止攻击"
              value={stats.blockedAttacks}
              prefix={<SafetyCertificateOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
              suffix={
                <Tooltip title="防火墙和入侵检测系统阻止的攻击次数">
                  <ThunderboltOutlined style={{ color: '#52c41a' }} />
                </Tooltip>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable>
            <Statistic
              title="活跃连接"
              value={stats.activeConnections}
              prefix={<GlobalOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
              suffix={
                <Tooltip title="当前活跃的网络连接数">
                  <WifiOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              }
            />
          </Card>
        </Col>
      </Row>

      {/* 系统状态和健康监控 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={8}>
          <Card title="系统健康状态" extra={
            <Badge 
              status={systemHealth.status === 'healthy' ? 'success' : systemHealth.status === 'warning' ? 'warning' : 'error'} 
              text={systemHealth.status === 'healthy' ? '正常' : systemHealth.status === 'warning' ? '警告' : '严重'}
            />
          }>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text>CPU使用率</Text>
                <Progress 
                  percent={systemHealth.cpu} 
                  size="small" 
                  status={systemHealth.cpu > 80 ? 'exception' : systemHealth.cpu > 60 ? 'active' : 'success'}
                />
              </div>
              <div>
                <Text>内存使用率</Text>
                <Progress 
                  percent={systemHealth.memory} 
                  size="small"
                  status={systemHealth.memory > 85 ? 'exception' : systemHealth.memory > 70 ? 'active' : 'success'}
                />
              </div>
              <div>
                <Text>磁盘使用率</Text>
                <Progress 
                  percent={systemHealth.disk} 
                  size="small"
                  status={systemHealth.disk > 90 ? 'exception' : systemHealth.disk > 75 ? 'active' : 'success'}
                />
              </div>
              <div>
                <Text>网络流量</Text>
                <Progress 
                  percent={systemHealth.network} 
                  size="small"
                  status="normal"
                />
              </div>
            </Space>
          </Card>
        </Col>
        
        <Col xs={24} lg={8}>
          <Card title="安全服务状态">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Row justify="space-between" align="middle">
                <Col>
                  <Space>
                    <FireOutlined />
                    <Text>防火墙</Text>
                  </Space>
                </Col>
                <Col>
                  <Space>
                    {getStatusIcon(stats.firewallStatus)}
                    <Tag color={getStatusColor(stats.firewallStatus)}>
                      {stats.firewallStatus === 'active' ? '运行中' : stats.firewallStatus === 'warning' ? '警告' : '已停止'}
                    </Tag>
                  </Space>
                </Col>
              </Row>
              
              <Row justify="space-between" align="middle">
                <Col>
                  <Space>
                    <BugOutlined />
                    <Text>防病毒</Text>
                  </Space>
                </Col>
                <Col>
                  <Space>
                    {getStatusIcon(stats.antivirusStatus)}
                    <Tag color={getStatusColor(stats.antivirusStatus)}>
                      {stats.antivirusStatus === 'active' ? '运行中' : stats.antivirusStatus === 'warning' ? '警告' : '已停止'}
                    </Tag>
                  </Space>
                </Col>
              </Row>
              
              <Row justify="space-between" align="middle">
                <Col>
                  <Space>
                    <RadarChartOutlined />
                    <Text>入侵检测</Text>
                  </Space>
                </Col>
                <Col>
                  <Space>
                    {getStatusIcon(stats.intrusionDetection)}
                    <Tag color={getStatusColor(stats.intrusionDetection)}>
                      {stats.intrusionDetection === 'active' ? '运行中' : stats.intrusionDetection === 'warning' ? '警告' : '已停止'}
                    </Tag>
                  </Space>
                </Col>
              </Row>
              
              <Row justify="space-between" align="middle">
                <Col>
                  <Space>
                    <LockOutlined />
                    <Text>数据加密</Text>
                  </Space>
                </Col>
                <Col>
                  <Space>
                    {getStatusIcon(stats.dataEncryption)}
                    <Tag color={getStatusColor(stats.dataEncryption)}>
                      {stats.dataEncryption === 'active' ? '运行中' : stats.dataEncryption === 'warning' ? '警告' : '已停止'}
                    </Tag>
                  </Space>
                </Col>
              </Row>
            </Space>
          </Card>
        </Col>
        
        <Col xs={24} lg={8}>
          <Card title="系统运行时间" extra={
            <Tooltip title="系统连续运行时间">
              <ClockCircleOutlined />
            </Tooltip>
          }>
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <Statistic
                value={formatUptime(stats.systemUptime)}
                valueStyle={{ fontSize: '24px', fontWeight: 'bold' }}
              />
              <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
                系统稳定运行
              </Text>
              <Progress
                type="circle"
                percent={Math.min((stats.systemUptime / 720) * 100, 100)}
                size={80}
                format={() => '稳定'}
                strokeColor="#52c41a"
                style={{ marginTop: 16 }}
              />
            </div>
          </Card>
        </Col>
      </Row>

      {/* 安全风险评分和合规性 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={12}>
          <Card 
            title="安全风险评分" 
            extra={
              <Space>
                <Tooltip title="风险评分说明">
                  <Button icon={<QuestionCircleFilled />} type="text" size="small" />
                </Tooltip>
                <Button icon={<ScanOutlined />} type="primary" size="small">
                  重新评估
                </Button>
              </Space>
            }
          >
            <Row gutter={24}>
              <Col span={12}>
                <div style={{ textAlign: 'center' }}>
                  <Progress
                    type="circle"
                    percent={stats.riskScore}
                    format={() => `${stats.riskScore}分`}
                    strokeColor={riskInfo.color}
                    size={120}
                    strokeWidth={8}
                  />
                  <div style={{ marginTop: 16 }}>
                    <Tag color={riskInfo.color} style={{ fontSize: 14, padding: '4px 12px' }}>
                      {riskInfo.text}
                    </Tag>
                  </div>
                </div>
              </Col>
              <Col span={12}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text strong>合规性评分</Text>
                    <Progress 
                      percent={stats.complianceScore} 
                      size="small"
                      status={stats.complianceScore > 90 ? 'success' : stats.complianceScore > 70 ? 'active' : 'exception'}
                    />
                  </div>
                  <div>
                    <Text strong>24小时事件</Text>
                    <div style={{ marginTop: 4 }}>
                      <Badge count={stats.securityEvents24h} style={{ backgroundColor: '#1890ff' }} />
                      <Text type="secondary" style={{ marginLeft: 8 }}>个安全事件</Text>
                    </div>
                  </div>
                  <div>
                    <Text strong>已解决事件</Text>
                    <div style={{ marginTop: 4 }}>
                      <Badge count={stats.resolvedIncidents} style={{ backgroundColor: '#52c41a' }} />
                      <Text type="secondary" style={{ marginLeft: 8 }}>个已处理</Text>
                    </div>
                  </div>
                </Space>
              </Col>
            </Row>
            <Divider />
            <Text type="secondary" style={{ fontSize: 12 }}>
              基于威胁检测、漏洞扫描、合规检查、系统配置等多维度综合评估
            </Text>
          </Card>
        </Col>
        
        <Col xs={24} lg={12}>
          <Card 
            title="最近安全事件" 
            extra={
              <Space>
                <Button 
                  icon={<EyeOutlined />} 
                  size="small"
                  onClick={() => setDrawerVisible(true)}
                >
                  查看全部
                </Button>
                <Button icon={<FilterOutlined />} size="small">
                  筛选
                </Button>
              </Space>
            }
          >
            <div style={{ height: 280, overflowY: 'auto' }}>
              {securityEvents.length > 0 ? (
                <List
                  size="small"
                  dataSource={securityEvents.slice(0, 5)}
                  renderItem={(event) => (
                    <List.Item
                      style={{ 
                        padding: '8px 0',
                        cursor: 'pointer',
                        borderRadius: 4,
                        marginBottom: 4
                      }}
                      onClick={() => {
                        setSelectedEvent(event);
                        setDrawerVisible(true);
                      }}
                    >
                      <List.Item.Meta
                        avatar={
                          <Avatar 
                            icon={getEventTypeIcon(event.type)} 
                            style={{ 
                              backgroundColor: getSeverityColor(event.severity) === 'blue' ? '#1890ff' : 
                                              getSeverityColor(event.severity) === 'orange' ? '#fa8c16' :
                                              getSeverityColor(event.severity) === 'red' ? '#ff4d4f' : '#722ed1'
                            }}
                          />
                        }
                        title={
                          <Space>
                            <Text strong style={{ fontSize: 13 }}>{event.title}</Text>
                            <Tag 
                              color={getSeverityColor(event.severity)} 
                              size="small"
                            >
                              {event.severity}
                            </Tag>
                          </Space>
                        }
                        description={
                          <div>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {event.description.length > 50 ? 
                                `${event.description.substring(0, 50)}...` : 
                                event.description
                              }
                            </Text>
                            <br />
                            <Text type="secondary" style={{ fontSize: 11 }}>
                              {new Date(event.timestamp).toLocaleString()} • {event.source}
                            </Text>
                          </div>
                        }
                      />
                      <div>
                        <Tag 
                          color={
                            event.status === 'new' ? 'red' :
                            event.status === 'investigating' ? 'orange' :
                            event.status === 'resolved' ? 'green' : 'gray'
                          }
                          size="small"
                        >
                          {event.status === 'new' ? '新' :
                           event.status === 'investigating' ? '处理中' :
                           event.status === 'resolved' ? '已解决' : '已忽略'}
                        </Tag>
                      </div>
                    </List.Item>
                  )}
                />
              ) : (
                <Empty 
                  description="暂无安全事件"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  style={{ marginTop: 40 }}
                />
              )}
            </div>
          </Card>
        </Col>
      </Row>

      {/* 快速操作面板 */}
      <Card 
        title="快速操作" 
        extra={
          <Tooltip title="自定义快速操作">
            <Button icon={<SettingOutlined />} type="text" size="small" />
          </Tooltip>
        }
      >
        <Row gutter={[12, 12]}>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card 
              hoverable 
              size="small"
              style={{ textAlign: 'center', height: 100 }}
              bodyStyle={{ padding: '12px 8px' }}
              onClick={() => setActiveTab('threat-detection')}
            >
              <ScanOutlined style={{ fontSize: 24, color: '#1890ff', marginBottom: 8 }} />
              <div style={{ fontSize: 12, fontWeight: 500 }}>威胁检测</div>
              <Badge count={stats.threatAlerts} size="small" style={{ marginTop: 4 }} />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card 
              hoverable 
              size="small"
              style={{ textAlign: 'center', height: 100 }}
              bodyStyle={{ padding: '12px 8px' }}
              onClick={() => setActiveTab('vulnerability')}
            >
              <BugOutlined style={{ fontSize: 24, color: '#fa8c16', marginBottom: 8 }} />
              <div style={{ fontSize: 12, fontWeight: 500 }}>漏洞管理</div>
              <Badge count={stats.vulnerabilities} size="small" style={{ marginTop: 4 }} />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card 
              hoverable 
              size="small"
              style={{ textAlign: 'center', height: 100 }}
              bodyStyle={{ padding: '12px 8px' }}
              onClick={() => setActiveTab('pentest')}
            >
              <ExperimentOutlined style={{ fontSize: 24, color: '#722ed1', marginBottom: 8 }} />
              <div style={{ fontSize: 12, fontWeight: 500 }}>渗透测试</div>
              <Badge count={stats.activeTests} size="small" style={{ marginTop: 4 }} />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card 
              hoverable 
              size="small"
              style={{ textAlign: 'center', height: 100 }}
              bodyStyle={{ padding: '12px 8px' }}
              onClick={() => setActiveTab('education')}
            >
              <BookOutlined style={{ fontSize: 24, color: '#52c41a', marginBottom: 8 }} />
              <div style={{ fontSize: 12, fontWeight: 500 }}>安全教育</div>
              <Badge count={stats.completedCourses} size="small" style={{ marginTop: 4 }} />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card 
              hoverable 
              size="small"
              style={{ textAlign: 'center', height: 100 }}
              bodyStyle={{ padding: '12px 8px' }}
              onClick={() => setActiveTab('audit')}
            >
              <AuditOutlined style={{ fontSize: 24, color: '#13c2c2', marginBottom: 8 }} />
              <div style={{ fontSize: 12, fontWeight: 500 }}>安全审计</div>
              <Badge count="新" size="small" style={{ marginTop: 4 }} />
            </Card>
          </Col>
          <Col xs={12} sm={8} md={6} lg={4}>
            <Card 
              hoverable 
              size="small"
              style={{ textAlign: 'center', height: 100 }}
              bodyStyle={{ padding: '12px 8px' }}
            >
              <FileProtectOutlined style={{ fontSize: 24, color: '#eb2f96', marginBottom: 8 }} />
              <div style={{ fontSize: 12, fontWeight: 500 }}>数据保护</div>
              <Badge count="Pro" size="small" style={{ marginTop: 4 }} />
            </Card>
          </Col>
        </Row>
      </Card>

      {/* 安全事件详情抽屉 */}
      <Drawer
        title="安全事件详情"
        placement="right"
        width={600}
        onClose={() => {
          setDrawerVisible(false);
          setSelectedEvent(null);
        }}
        open={drawerVisible}
      >
        {selectedEvent ? (
          <div>
            <Descriptions title={selectedEvent.title} bordered column={1}>
              <Descriptions.Item label="事件类型">
                <Space>
                  {getEventTypeIcon(selectedEvent.type)}
                  <Tag color={getSeverityColor(selectedEvent.severity)}>
                    {selectedEvent.type}
                  </Tag>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="严重程度">
                <Tag color={getSeverityColor(selectedEvent.severity)} style={{ fontSize: 14 }}>
                  {selectedEvent.severity.toUpperCase()}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag 
                  color={
                    selectedEvent.status === 'new' ? 'red' :
                    selectedEvent.status === 'investigating' ? 'orange' :
                    selectedEvent.status === 'resolved' ? 'green' : 'gray'
                  }
                >
                  {selectedEvent.status === 'new' ? '新事件' :
                   selectedEvent.status === 'investigating' ? '调查中' :
                   selectedEvent.status === 'resolved' ? '已解决' : '已忽略'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="发生时间">
                {new Date(selectedEvent.timestamp).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="事件源">
                {selectedEvent.source}
              </Descriptions.Item>
              {selectedEvent.assignee && (
                <Descriptions.Item label="负责人">
                  <Space>
                    <UserOutlined />
                    {selectedEvent.assignee}
                  </Space>
                </Descriptions.Item>
              )}
              <Descriptions.Item label="详细描述">
                {selectedEvent.description}
              </Descriptions.Item>
            </Descriptions>
            
            <Divider />
            
            <Space style={{ width: '100%', justifyContent: 'center' }}>
              <Button type="primary" icon={<CheckCircleOutlined />}>
                标记为已解决
              </Button>
              <Button icon={<UserOutlined />}>
                分配处理人
              </Button>
              <Button icon={<CommentOutlined />}>
                添加备注
              </Button>
              <Button danger icon={<CloseCircleFilled />}>
                忽略事件
              </Button>
            </Space>
          </div>
        ) : (
          <div>
            <Title level={4}>最近安全事件</Title>
            <List
              dataSource={securityEvents}
              renderItem={(event) => (
                <List.Item
                  style={{ cursor: 'pointer' }}
                  onClick={() => setSelectedEvent(event)}
                >
                  <List.Item.Meta
                    avatar={
                      <Avatar 
                        icon={getEventTypeIcon(event.type)} 
                        style={{ 
                          backgroundColor: getSeverityColor(event.severity) === 'blue' ? '#1890ff' : 
                                          getSeverityColor(event.severity) === 'orange' ? '#fa8c16' :
                                          getSeverityColor(event.severity) === 'red' ? '#ff4d4f' : '#722ed1'
                        }}
                      />
                    }
                    title={
                      <Space>
                        <Text strong>{event.title}</Text>
                        <Tag color={getSeverityColor(event.severity)} size="small">
                          {event.severity}
                        </Tag>
                      </Space>
                    }
                    description={
                      <div>
                        <Text type="secondary">{event.description}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {new Date(event.timestamp).toLocaleString()} • {event.source}
                        </Text>
                      </div>
                    }
                  />
                  <Tag 
                    color={
                      event.status === 'new' ? 'red' :
                      event.status === 'investigating' ? 'orange' :
                      event.status === 'resolved' ? 'green' : 'gray'
                    }
                  >
                    {event.status === 'new' ? '新' :
                     event.status === 'investigating' ? '处理中' :
                     event.status === 'resolved' ? '已解决' : '已忽略'}
                  </Tag>
                </List.Item>
              )}
            />
          </div>
        )}
      </Drawer>
    </div>
  );

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={2}>
          <SafetyCertificateOutlined style={{ marginRight: 8 }} />
          安全控制台
        </Title>
        <Text type="secondary">
          全面的安全管理平台，提供威胁检测、漏洞管理、渗透测试、安全教育和审计功能
        </Text>
      </div>

      <Tabs 
        activeKey={activeTab} 
        onChange={setActiveTab} 
        size="large"
        items={[
          {
            key: 'overview',
            label: (
              <span>
                <EyeOutlined />
                安全概览
              </span>
            ),
            children: renderOverview()
          },
          {
            key: 'threat-detection',
            label: (
              <span>
                <SafetyCertificateOutlined />
                威胁检测
              </span>
            ),
            children: <ThreatDetection />
          },
          {
            key: 'vulnerability',
            label: (
              <span>
                <BugOutlined />
                漏洞管理
              </span>
            ),
            children: <VulnerabilityManagement />
          },
          {
            key: 'pentest',
            label: (
              <span>
                <ExperimentOutlined />
                渗透测试
              </span>
            ),
            children: <PenetrationTesting />
          },
          {
            key: 'education',
            label: (
              <span>
                <BookOutlined />
                安全教育
              </span>
            ),
            children: <SecurityEducation />
          },
          {
            key: 'audit',
            label: (
              <span>
                <AuditOutlined />
                安全审计
              </span>
            ),
            children: <SecurityAudit />
          }
        ]}
      />
    </div>
  );
};

export default SecurityCenter;