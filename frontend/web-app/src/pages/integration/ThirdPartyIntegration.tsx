import React, { useState } from 'react';
import { Card, Tabs, Row, Col, Statistic, Button, Space, Typography, Divider } from 'antd';
import { 
  ApiOutlined, 
  AppstoreOutlined, 
  LinkOutlined, 
  KeyOutlined,
  SettingOutlined,
  CloudOutlined,
  SecurityScanOutlined,
  BarChartOutlined
} from '@ant-design/icons';
import APIKeyManagement from './components/APIKeyManagement';
import PluginManagement from './components/PluginManagement';
import ServiceIntegration from './components/ServiceIntegration';
import WebhookManagement from './components/WebhookManagement';
import OAuthManagement from './components/OAuthManagement';

const { Title, Paragraph } = Typography;
const { TabPane } = Tabs;

interface IntegrationStats {
  totalAPIKeys: number;
  activePlugins: number;
  totalIntegrations: number;
  webhookEvents: number;
  oauthApps: number;
  monthlyRequests: number;
}

const ThirdPartyIntegration: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [stats] = useState<IntegrationStats>({
    totalAPIKeys: 12,
    activePlugins: 8,
    totalIntegrations: 15,
    webhookEvents: 234,
    oauthApps: 5,
    monthlyRequests: 45678
  });

  const renderOverview = () => (
    <div>
      <Row gutter={[24, 24]} style={{ marginBottom: 24 }}>
        <Col span={24}>
          <Card>
            <Title level={3}>
              <CloudOutlined style={{ marginRight: 8, color: '#1890ff' }} />
              第三方集成平台
            </Title>
            <Paragraph>
              强大的第三方集成平台，提供API密钥管理、插件系统、服务集成、Webhook处理和OAuth认证等功能。
              轻松连接外部服务，扩展系统能力。
            </Paragraph>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="API密钥"
              value={stats.totalAPIKeys}
              prefix={<KeyOutlined style={{ color: '#52c41a' }} />}
              suffix="个"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="活跃插件"
              value={stats.activePlugins}
              prefix={<AppstoreOutlined style={{ color: '#1890ff' }} />}
              suffix="个"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="集成服务"
              value={stats.totalIntegrations}
              prefix={<LinkOutlined style={{ color: '#722ed1' }} />}
              suffix="个"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Webhook事件"
              value={stats.webhookEvents}
              prefix={<BarChartOutlined style={{ color: '#fa8c16' }} />}
              suffix="次"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="OAuth应用"
              value={stats.oauthApps}
              prefix={<SecurityScanOutlined style={{ color: '#eb2f96' }} />}
              suffix="个"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="月度请求"
              value={stats.monthlyRequests}
              prefix={<ApiOutlined style={{ color: '#13c2c2' }} />}
              suffix="次"
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <KeyOutlined style={{ color: '#52c41a' }} />
                API密钥管理
              </Space>
            }
            extra={
              <Button 
                type="primary" 
                size="small"
                onClick={() => setActiveTab('api-keys')}
              >
                管理
              </Button>
            }
          >
            <Paragraph>
              创建和管理API密钥，控制访问权限，监控使用情况。支持细粒度权限控制和使用统计。
            </Paragraph>
            <Space>
              <Button type="link" onClick={() => setActiveTab('api-keys')}>
                查看所有密钥
              </Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <AppstoreOutlined style={{ color: '#1890ff' }} />
                插件系统
              </Space>
            }
            extra={
              <Button 
                type="primary" 
                size="small"
                onClick={() => setActiveTab('plugins')}
              >
                管理
              </Button>
            }
          >
            <Paragraph>
              安装、配置和管理插件，扩展系统功能。支持插件生命周期管理和沙箱运行环境。
            </Paragraph>
            <Space>
              <Button type="link" onClick={() => setActiveTab('plugins')}>
                浏览插件市场
              </Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <LinkOutlined style={{ color: '#722ed1' }} />
                服务集成
              </Space>
            }
            extra={
              <Button 
                type="primary" 
                size="small"
                onClick={() => setActiveTab('integrations')}
              >
                管理
              </Button>
            }
          >
            <Paragraph>
              连接第三方服务，配置数据同步，管理集成状态。支持多种集成类型和自动化流程。
            </Paragraph>
            <Space>
              <Button type="link" onClick={() => setActiveTab('integrations')}>
                添加新集成
              </Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <SecurityScanOutlined style={{ color: '#eb2f96' }} />
                OAuth认证
              </Space>
            }
            extra={
              <Button 
                type="primary" 
                size="small"
                onClick={() => setActiveTab('oauth')}
              >
                管理
              </Button>
            }
          >
            <Paragraph>
              管理OAuth应用和令牌，提供安全的第三方认证。支持标准OAuth 2.0流程。
            </Paragraph>
            <Space>
              <Button type="link" onClick={() => setActiveTab('oauth')}>
                创建OAuth应用
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );

  return (
    <div style={{ padding: '24px', background: '#f0f2f5', minHeight: '100vh' }}>
      <div style={{ maxWidth: 1200, margin: '0 auto' }}>
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          size="large"
          style={{ background: 'white', borderRadius: 8, padding: '0 24px' }}
        >
          <TabPane
            tab={
              <span>
                <BarChartOutlined />
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
                <KeyOutlined />
                API密钥
              </span>
            }
            key="api-keys"
          >
            <APIKeyManagement />
          </TabPane>
          
          <TabPane
            tab={
              <span>
                <AppstoreOutlined />
                插件管理
              </span>
            }
            key="plugins"
          >
            <PluginManagement />
          </TabPane>
          
          <TabPane
            tab={
              <span>
                <LinkOutlined />
                服务集成
              </span>
            }
            key="integrations"
          >
            <ServiceIntegration />
          </TabPane>
          
          <TabPane
            tab={
              <span>
                <SettingOutlined />
                Webhook
              </span>
            }
            key="webhooks"
          >
            <WebhookManagement />
          </TabPane>
          
          <TabPane
            tab={
              <span>
                <SecurityScanOutlined />
                OAuth
              </span>
            }
            key="oauth"
          >
            <OAuthManagement />
          </TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default ThirdPartyIntegration;