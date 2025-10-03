import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Switch,
  Select,
  InputNumber,
  Tabs,
  message,
  Space,
  Divider,
  Row,
  Col,
  Upload,
  Modal,
  Progress,
  Statistic,
  Alert
} from 'antd';
import {
  GlobalOutlined,
  DatabaseOutlined,
  SecurityScanOutlined,
  UploadOutlined,
  DeleteOutlined,
  ReloadOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';

const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;
const { confirm } = Modal;

interface SystemConfig {
  siteName: string;
  siteDescription: string;
  siteKeywords: string;
  siteLogo: string;
  favicon: string;
  contactEmail: string;
  icp: string;
  copyright: string;
  enableRegistration: boolean;
  enableComments: boolean;
  enableSearch: boolean;
  maxUploadSize: number;
  allowedFileTypes: string[];
  defaultLanguage: string;
  timezone: string;
}

interface SEOConfig {
  metaTitle: string;
  metaDescription: string;
  metaKeywords: string;
  ogTitle: string;
  ogDescription: string;
  ogImage: string;
  twitterCard: string;
  robotsTxt: string;
  sitemapEnabled: boolean;
  analyticsCode: string;
}

interface CacheConfig {
  enabled: boolean;
  ttl: number;
  maxSize: number;
  strategy: string;
}

interface CacheStats {
  hitRate: number;
  totalRequests: number;
  cacheSize: number;
  memoryUsage: number;
}

const SystemSettings: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [systemForm] = Form.useForm();
  const [seoForm] = Form.useForm();
  const [cacheForm] = Form.useForm();
  const [cacheStats, setCacheStats] = useState<CacheStats>({
    hitRate: 0,
    totalRequests: 0,
    cacheSize: 0,
    memoryUsage: 0
  });

  useEffect(() => {
    loadSystemConfig();
    loadSEOConfig();
    loadCacheConfig();
    loadCacheStats();
  }, []);

  const loadSystemConfig = async () => {
    try {
      const response = await apiClient.getSystemConfig();
      systemForm.setFieldsValue(response.data);
    } catch {
      message.error('加载系统配置失败');
    }
  };

  const loadSEOConfig = async () => {
    try {
      const response = await apiClient.getSEOConfig();
      seoForm.setFieldsValue(response.data);
    } catch {
      message.error('加载SEO配置失败');
    }
  };

  const loadCacheConfig = async () => {
    try {
      const response = await apiClient.getCacheConfig();
      cacheForm.setFieldsValue(response.data);
    } catch {
      message.error('加载缓存配置失败');
    }
  };

  const loadCacheStats = async () => {
    try {
      const response = await apiClient.getCacheStats();
      setCacheStats(response.data);
    } catch (error) {
      console.error('加载缓存统计失败:', error);
    }
  };

  const handleSystemConfigSave = async (values: SystemConfig) => {
    setLoading(true);
    try {
      await apiClient.updateSystemConfig(values);
      message.success('系统配置保存成功');
    } catch {
      message.error('系统配置保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSEOConfigSave = async (values: SEOConfig) => {
    setLoading(true);
    try {
      await apiClient.updateSEOConfig(values);
      message.success('SEO配置保存成功');
    } catch {
      message.error('SEO配置保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCacheConfigSave = async (values: CacheConfig) => {
    setLoading(true);
    try {
      await apiClient.updateCacheConfig(values);
      message.success('缓存配置保存成功');
      loadCacheStats();
    } catch {
      message.error('缓存配置保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClearCache = () => {
    confirm({
      title: '确认清空缓存',
      icon: <ExclamationCircleOutlined />,
      content: '清空缓存后可能会暂时影响网站性能，确定要继续吗？',
      onOk: async () => {
        try {
          await apiClient.clearCache();
          message.success('缓存清空成功');
          loadCacheStats();
        } catch {
          message.error('缓存清空失败');
        }
      }
    });
  };

  const handleTestEmail = async () => {
    try {
      await apiClient.testEmailConfig();
      message.success('邮件配置测试成功');
    } catch {
      message.error('邮件配置测试失败');
    }
  };

  const uploadProps = {
    name: 'file',
    action: '/api/upload',
    headers: {
      authorization: `Bearer ${localStorage.getItem('token')}`
    },
    onChange(info: { file: { status: string; name: string } }) {
      if (info.file.status === 'done') {
        message.success(`${info.file.name} 文件上传成功`);
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} 文件上传失败`);
      }
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <h1 style={{ fontSize: '24px', fontWeight: 'bold', margin: 0 }}>
          <SettingOutlined style={{ marginRight: '8px' }} />
          系统设置
        </h1>
        <p style={{ color: '#666', margin: '8px 0 0 0' }}>
          管理网站配置、SEO设置和缓存策略
        </p>
      </div>

      <Tabs defaultActiveKey="system" size="large">
        <TabPane
          tab={
            <span>
              <GlobalOutlined />
              网站配置
            </span>
          }
          key="system"
        >
          <Card>
            <Form
              form={systemForm}
              layout="vertical"
              onFinish={handleSystemConfigSave}
            >
              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="网站名称"
                    name="siteName"
                    rules={[{ required: true, message: '请输入网站名称' }]}
                  >
                    <Input placeholder="请输入网站名称" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="联系邮箱"
                    name="contactEmail"
                    rules={[
                      { required: true, message: '请输入联系邮箱' },
                      { type: 'email', message: '请输入有效的邮箱地址' }
                    ]}
                  >
                    <Input placeholder="请输入联系邮箱" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                label="网站描述"
                name="siteDescription"
                rules={[{ required: true, message: '请输入网站描述' }]}
              >
                <TextArea rows={3} placeholder="请输入网站描述" />
              </Form.Item>

              <Form.Item
                label="网站关键词"
                name="siteKeywords"
                help="多个关键词用逗号分隔"
              >
                <Input placeholder="请输入网站关键词" />
              </Form.Item>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item label="网站Logo" name="siteLogo">
                    <Upload {...uploadProps}>
                      <Button icon={<UploadOutlined />}>上传Logo</Button>
                    </Upload>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="网站图标" name="favicon">
                    <Upload {...uploadProps}>
                      <Button icon={<UploadOutlined />}>上传图标</Button>
                    </Upload>
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item label="ICP备案号" name="icp">
                    <Input placeholder="请输入ICP备案号" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="版权信息" name="copyright">
                    <Input placeholder="请输入版权信息" />
                  </Form.Item>
                </Col>
              </Row>

              <Divider>功能设置</Divider>

              <Row gutter={24}>
                <Col span={8}>
                  <Form.Item
                    label="允许用户注册"
                    name="enableRegistration"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    label="启用评论功能"
                    name="enableComments"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    label="启用搜索功能"
                    name="enableSearch"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="最大上传大小(MB)"
                    name="maxUploadSize"
                    rules={[{ required: true, message: '请输入最大上传大小' }]}
                  >
                    <InputNumber min={1} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="默认语言"
                    name="defaultLanguage"
                    rules={[{ required: true, message: '请选择默认语言' }]}
                  >
                    <Select placeholder="请选择默认语言">
                      <Option value="zh-CN">简体中文</Option>
                      <Option value="zh-TW">繁体中文</Option>
                      <Option value="en-US">English</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit" loading={loading}>
                    保存配置
                  </Button>
                  <Button onClick={handleTestEmail}>
                    测试邮件配置
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <SecurityScanOutlined />
              SEO设置
            </span>
          }
          key="seo"
        >
          <Card>
            <Form
              form={seoForm}
              layout="vertical"
              onFinish={handleSEOConfigSave}
            >
              <Alert
                message="SEO优化提示"
                description="合理的SEO设置可以提高网站在搜索引擎中的排名，建议定期更新和优化。"
                type="info"
                showIcon
                style={{ marginBottom: '24px' }}
              />

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="页面标题"
                    name="metaTitle"
                    rules={[{ required: true, message: '请输入页面标题' }]}
                  >
                    <Input placeholder="请输入页面标题" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="页面关键词"
                    name="metaKeywords"
                    help="多个关键词用逗号分隔"
                  >
                    <Input placeholder="请输入页面关键词" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item
                label="页面描述"
                name="metaDescription"
                rules={[{ required: true, message: '请输入页面描述' }]}
              >
                <TextArea rows={3} placeholder="请输入页面描述" />
              </Form.Item>

              <Divider>社交媒体优化</Divider>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item label="OG标题" name="ogTitle">
                    <Input placeholder="请输入OG标题" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="Twitter卡片类型" name="twitterCard">
                    <Select placeholder="请选择Twitter卡片类型">
                      <Option value="summary">摘要</Option>
                      <Option value="summary_large_image">大图摘要</Option>
                      <Option value="app">应用</Option>
                      <Option value="player">播放器</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item label="OG描述" name="ogDescription">
                <TextArea rows={2} placeholder="请输入OG描述" />
              </Form.Item>

              <Form.Item label="OG图片" name="ogImage">
                <Upload {...uploadProps}>
                  <Button icon={<UploadOutlined />}>上传OG图片</Button>
                </Upload>
              </Form.Item>

              <Divider>高级设置</Divider>

              <Form.Item
                label="Robots.txt"
                name="robotsTxt"
                help="控制搜索引擎爬虫的访问规则"
              >
                <TextArea
                  rows={6}
                  placeholder="User-agent: *&#10;Disallow: /admin/&#10;Allow: /"
                />
              </Form.Item>

              <Row gutter={24}>
                <Col span={12}>
                  <Form.Item
                    label="启用站点地图"
                    name="sitemapEnabled"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="统计代码" name="analyticsCode">
                    <Input placeholder="请输入Google Analytics等统计代码" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  保存SEO配置
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <DatabaseOutlined />
              缓存管理
            </span>
          }
          key="cache"
        >
          <Row gutter={24}>
            <Col span={16}>
              <Card title="缓存配置">
                <Form
                  form={cacheForm}
                  layout="vertical"
                  onFinish={handleCacheConfigSave}
                >
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label="启用缓存"
                        name="enabled"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label="缓存策略"
                        name="strategy"
                        rules={[{ required: true, message: '请选择缓存策略' }]}
                      >
                        <Select placeholder="请选择缓存策略">
                          <Option value="lru">LRU (最近最少使用)</Option>
                          <Option value="lfu">LFU (最少使用频率)</Option>
                          <Option value="fifo">FIFO (先进先出)</Option>
                          <Option value="ttl">TTL (生存时间)</Option>
                        </Select>
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label="缓存时间(秒)"
                        name="ttl"
                        rules={[{ required: true, message: '请输入缓存时间' }]}
                      >
                        <InputNumber
                          min={60}
                          max={86400}
                          style={{ width: '100%' }}
                          placeholder="3600"
                        />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label="最大缓存大小(MB)"
                        name="maxSize"
                        rules={[{ required: true, message: '请输入最大缓存大小' }]}
                      >
                        <InputNumber
                          min={10}
                          max={1024}
                          style={{ width: '100%' }}
                          placeholder="256"
                        />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item>
                    <Space>
                      <Button type="primary" htmlType="submit" loading={loading}>
                        保存配置
                      </Button>
                      <Button
                        danger
                        icon={<DeleteOutlined />}
                        onClick={handleClearCache}
                      >
                        清空缓存
                      </Button>
                      <Button
                        icon={<ReloadOutlined />}
                        onClick={loadCacheStats}
                      >
                        刷新统计
                      </Button>
                    </Space>
                  </Form.Item>
                </Form>
              </Card>
            </Col>

            <Col span={8}>
              <Card title="缓存统计">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Statistic
                    title="缓存命中率"
                    value={cacheStats.hitRate}
                    precision={2}
                    suffix="%"
                    valueStyle={{
                      color: cacheStats.hitRate > 80 ? '#3f8600' : '#cf1322'
                    }}
                  />
                  <Progress
                    percent={cacheStats.hitRate}
                    status={cacheStats.hitRate > 80 ? 'success' : 'exception'}
                    showInfo={false}
                  />

                  <Divider />

                  <Statistic
                    title="总请求数"
                    value={cacheStats.totalRequests}
                    prefix={<ClockCircleOutlined />}
                  />

                  <Statistic
                    title="缓存大小"
                    value={cacheStats.cacheSize}
                    suffix="MB"
                    prefix={<DatabaseOutlined />}
                  />

                  <Statistic
                    title="内存使用"
                    value={cacheStats.memoryUsage}
                    precision={2}
                    suffix="%"
                    prefix={<CheckCircleOutlined />}
                    valueStyle={{
                      color: cacheStats.memoryUsage > 80 ? '#cf1322' : '#3f8600'
                    }}
                  />
                </Space>
              </Card>
            </Col>
          </Row>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default SystemSettings;