// 太上老君AI平台全球化管理界面
import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Alert, Tabs, Button, Space, Tag, Progress, Spin } from 'antd';
import {
  GlobalOutlined,
  TranslationOutlined,
  ShieldCheckOutlined,
  ClockCircleOutlined,
  UserOutlined,
  DatabaseOutlined,
  SyncOutlined,
  SettingOutlined,
  MonitorOutlined,
  SafetyOutlined,
  CloudServerOutlined
} from '@ant-design/icons';
import RegionManagement from './components/RegionManagement';
import LocalizationConfig from './components/LocalizationConfig';
import ComplianceMonitoring from './components/ComplianceMonitoring';
import './index.less';

const { TabPane } = Tabs;

interface GlobalStats {
  totalRegions: number;
  activeRegions: number;
  supportedLanguages: number;
  complianceScore: number;
  totalUsers: number;
  dataVolume: string;
  uptime: number;
  lastSync: string;
}

const GlobalManagement: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<GlobalStats>({
    totalRegions: 0,
    activeRegions: 0,
    supportedLanguages: 0,
    complianceScore: 0,
    totalUsers: 0,
    dataVolume: '0 TB',
    uptime: 0,
    lastSync: ''
  });
  const [activeTab, setActiveTab] = useState('overview');

  useEffect(() => {
    fetchGlobalStats();
  }, []);

  const fetchGlobalStats = async () => {
    try {
      setLoading(true);
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setStats({
        totalRegions: 6,
        activeRegions: 5,
        supportedLanguages: 17,
        complianceScore: 95.8,
        totalUsers: 1250000,
        dataVolume: '2.5 TB',
        uptime: 99.9,
        lastSync: new Date().toLocaleString('zh-CN')
      });
    } catch (error) {
      console.error('Failed to fetch global stats:', error);
    } finally {
      setLoading(false);
    }
  };

  const refreshData = () => {
    fetchGlobalStats();
  };

  const renderOverview = () => (
    <div className="global-overview">
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={24}>
          <Alert
            message="全球化部署状态"
            description="太上老君AI平台已在全球6个区域部署，支持17种语言，合规性评分95.8%"
            type="success"
            showIcon
            style={{ marginBottom: 16 }}
          />
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="部署区域"
              value={stats.activeRegions}
              suffix={`/ ${stats.totalRegions}`}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="支持语言"
              value={stats.supportedLanguages}
              prefix={<TranslationOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="合规性评分"
              value={stats.complianceScore}
              suffix="%"
              prefix={<SafetyOutlined />}
              valueStyle={{ color: stats.complianceScore >= 90 ? '#3f8600' : '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="系统可用性"
              value={stats.uptime}
              suffix="%"
              prefix={<MonitorOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="全球用户数"
              value={stats.totalUsers}
              prefix={<DatabaseOutlined />}
              formatter={(value) => `${(Number(value) / 1000000).toFixed(1)}M`}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="数据存储量"
              value={stats.dataVolume}
              prefix={<CloudServerOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={12}>
          <Card>
            <Statistic
              title="最后同步时间"
              value={stats.lastSync}
              prefix={<MonitorOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card title="部署状态" extra={<Button icon={<SyncOutlined />}>同步状态</Button>}>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Card size="small">
                  <Statistic
                    title="亚太区域"
                    value="正常运行"
                    valueStyle={{ color: '#3f8600', fontSize: 16 }}
                  />
                  <Progress percent={98} size="small" status="active" />
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small">
                  <Statistic
                    title="欧洲区域"
                    value="正常运行"
                    valueStyle={{ color: '#3f8600', fontSize: 16 }}
                  />
                  <Progress percent={96} size="small" status="active" />
                </Card>
              </Col>
              <Col span={8}>
                <Card size="small">
                  <Statistic
                    title="北美区域"
                    value="正常运行"
                    valueStyle={{ color: '#3f8600', fontSize: 16 }}
                  />
                  <Progress percent={99} size="small" status="active" />
                </Card>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );

  return (
    <div className="global-management">
      <div className="page-header">
        <div className="header-content">
          <h1>
            <GlobalOutlined /> 全球化管理中心
          </h1>
          <p>管理多区域部署、本地化配置和合规性监控</p>
        </div>
        <div className="header-actions">
          <Space>
            <Button onClick={refreshData} loading={loading}>
              刷新数据
            </Button>
            <Button type="primary" icon={<SettingOutlined />}>
              全局设置
            </Button>
          </Space>
        </div>
      </div>

      <Spin spinning={loading}>
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          type="card"
          size="large"
        >
          <TabPane 
            tab={
              <span>
                <MonitorOutlined />
                概览
              </span>
            } 
            key="overview"
          >
            {renderOverview()}
          </TabPane>
          
          <TabPane 
            tab={
              <span>
                <GlobalOutlined />
                区域管理
              </span>
            } 
            key="regions"
          >
            <RegionManagement />
          </TabPane>
          
          <TabPane 
            tab={
              <span>
                <TranslationOutlined />
                本地化配置
              </span>
            } 
            key="localization"
          >
            <LocalizationConfig />
          </TabPane>
          
          <TabPane 
            tab={
              <span>
                <ShieldCheckOutlined />
                合规性监控
              </span>
            } 
            key="compliance"
          >
            <ComplianceMonitoring />
          </TabPane>
        </Tabs>
      </Spin>
    </div>
  );
};

export default GlobalManagement;