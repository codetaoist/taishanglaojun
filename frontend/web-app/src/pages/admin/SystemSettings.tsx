import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
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
  ClockCircleOutlined,
  SettingOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';
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
  const { t } = useTranslation();
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
      message.error(t('adminSystemSettings.messages.loadSiteFailed'));
    }
  };

  const loadSEOConfig = async () => {
    try {
      const response = await apiClient.getSEOConfig();
      seoForm.setFieldsValue(response.data);
    } catch {
      message.error(t('adminSystemSettings.messages.loadSEOFailed'));
    }
  };

  const loadCacheConfig = async () => {
    try {
      const response = await apiClient.getCacheConfig();
      cacheForm.setFieldsValue(response.data);
    } catch {
      message.error(t('adminSystemSettings.messages.loadCacheFailed'));
    }
  };

  const loadCacheStats = async () => {
    try {
      const response = await apiClient.getCacheStats();
      setCacheStats(response.data);
    } catch (error) {
      console.error(t('adminSystemSettings.messages.loadCacheStatsFailed'), error);
    }
  };

  const handleSystemConfigSave = async (values: SystemConfig) => {
    setLoading(true);
    try {
      await apiClient.updateSystemConfig(values);
      message.success(t('adminSystemSettings.messages.saveSiteSuccess'));
    } catch {
      message.error(t('adminSystemSettings.messages.saveSiteFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleSEOConfigSave = async (values: SEOConfig) => {
    setLoading(true);
    try {
      await apiClient.updateSEOConfig(values);
      message.success(t('adminSystemSettings.messages.saveSEOSuccess'));
    } catch {
      message.error(t('adminSystemSettings.messages.saveSEOFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleCacheConfigSave = async (values: CacheConfig) => {
    setLoading(true);
    try {
      await apiClient.updateCacheConfig(values);
      message.success(t('adminSystemSettings.messages.saveCacheSuccess'));
      loadCacheStats();
    } catch {
      message.error(t('adminSystemSettings.messages.saveCacheFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleClearCache = () => {
    confirm({
      title: t('adminSystemSettings.messages.clearCacheTitle'),
      icon: <ExclamationCircleOutlined />,
      content: t('adminSystemSettings.messages.clearCacheContent'),
      onOk: async () => {
        try {
          await apiClient.clearCache();
          message.success(t('adminSystemSettings.messages.clearCacheSuccess'));
          loadCacheStats();
        } catch {
          message.error(t('adminSystemSettings.messages.clearCacheFailed'));
        }
      }
    });
  };



  // 为避免 antd Upload 的 value 警告，不让表单直接控制 Upload；
  // 使用 onChange 解析返回的文件URL并写入对应的隐藏字段。
  const makeUploadProps = (form: any, fieldName: string) => ({
    name: 'file',
    action: '/api/upload',
    headers: {
      authorization: `Bearer ${localStorage.getItem('token')}`
    },
    onChange(info: any) {
      const status = info?.file?.status;
      const fileName = info?.file?.name;
      if (status === 'done') {
        const url = info?.file?.response?.url || info?.file?.response?.data?.url || info?.file?.url;
        if (url) {
          form.setFieldsValue({ [fieldName]: url });
        }
        message.success(t('adminSystemSettings.messages.uploadSuccess', { filename: fileName }));
      } else if (status === 'error') {
        message.error(t('adminSystemSettings.messages.uploadFailed', { filename: fileName }));
      }
    }
  });

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <h1 style={{ fontSize: '24px', fontWeight: 'bold', margin: 0 }}>
          <SettingOutlined style={{ marginRight: '8px' }} />
          {t('adminSystemSettings.title')}
        </h1>
        <p style={{ color: '#666', margin: '8px 0 0 0' }}>
          {t('adminSystemSettings.subtitle')}
        </p>
      </div>

      <Tabs 
        defaultActiveKey="system" 
        size="large"
        items={[
          {
            key: 'system',
            label: (
              <span>
                <GlobalOutlined />
                {t('adminSystemSettings.cards.siteConfig')}
              </span>
            ),
            children: (
              <Card>
                <Form
                  form={systemForm}
                  layout="vertical"
                  onFinish={handleSystemConfigSave}
                >
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.siteName')}
                        name="siteName"
                        rules={[{ required: true, message: t('adminSystemSettings.placeholders.siteName') }]}
                      >
                        <Input placeholder={t('adminSystemSettings.placeholders.siteName')} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.contactEmail')}
                        name="contactEmail"
                        rules={[
                          { required: true, message: t('adminSystemSettings.placeholders.contactEmail') },
                          { type: 'email', message: t('validation.email') }
                        ]}
                      >
                        <Input placeholder={t('adminSystemSettings.placeholders.contactEmail')} />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item
                    label={t('adminSystemSettings.fields.siteDescription')}
                    name="siteDescription"
                    rules={[{ required: true, message: t('adminSystemSettings.placeholders.siteDescription') }]}
                  >
                    <TextArea rows={3} placeholder={t('adminSystemSettings.placeholders.siteDescription')} />
                  </Form.Item>

                  <Form.Item
                    label={t('adminSystemSettings.fields.siteKeywords')}
                    name="siteKeywords"
                    help={t('adminSystemSettings.help.keywords')}
                  >
                    <Input placeholder={t('adminSystemSettings.placeholders.siteKeywords')} />
                  </Form.Item>

                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item label={t('adminSystemSettings.fields.siteLogo')}>
                        <Upload {...makeUploadProps(systemForm, 'siteLogo')}>
                          <Button icon={<UploadOutlined />}>{t('adminSystemSettings.buttons.uploadLogo')}</Button>
                        </Upload>
                      </Form.Item>
                      <Form.Item name="siteLogo" hidden>
                        <Input />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item label={t('adminSystemSettings.fields.favicon')}>
                        <Upload {...makeUploadProps(systemForm, 'favicon')}>
                          <Button icon={<UploadOutlined />}>{t('adminSystemSettings.buttons.uploadIcon')}</Button>
                        </Upload>
                      </Form.Item>
                      <Form.Item name="favicon" hidden>
                        <Input />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item label={t('adminSystemSettings.fields.icp')} name="icp">
                        <Input placeholder={t('adminSystemSettings.placeholders.icp')} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item label={t('adminSystemSettings.fields.copyright')} name="copyright">
                        <Input placeholder={t('adminSystemSettings.placeholders.copyright')} />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Divider>{t('adminSystemSettings.cards.feature')}</Divider>

                  <Row gutter={24}>
                    <Col span={8}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.allowRegister')}
                        name="enableRegistration"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.enableComments')}
                        name="enableComments"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.enableSearch')}
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
                        label={t('adminSystemSettings.fields.maxUploadMB')}
                        name="maxUploadSize"
                        rules={[{ required: true, message: t('validation.required') }]}
                      >
                        <InputNumber min={1} max={100} style={{ width: '100%' }} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.defaultLanguage')}
                        name="defaultLanguage"
                        rules={[{ required: true, message: t('adminSystemSettings.placeholders.defaultLanguage') }]}
                      >
                        <Select placeholder={t('adminSystemSettings.placeholders.defaultLanguage')}>
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
                        {t('adminSystemSettings.buttons.save')}
                      </Button>
                    </Space>
                  </Form.Item>
                </Form>
              </Card>
            )
          },
          {
            key: 'seo',
            label: (
              <span>
                <SecurityScanOutlined />
                {t('adminSystemSettings.cards.seo')}
              </span>
            ),
            children: (
              <Card>
                <Form
                  form={seoForm}
                  layout="vertical"
                  onFinish={handleSEOConfigSave}
                >
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.pageTitle')}
                        name="metaTitle"
                        rules={[{ required: true, message: t('adminSystemSettings.placeholders.pageTitle') }]}
                      >
                        <Input placeholder={t('adminSystemSettings.placeholders.pageTitle')} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.pageKeywords')}
                        name="metaKeywords"
                        help={t('adminSystemSettings.help.keywords')}
                      >
                        <Input placeholder={t('adminSystemSettings.placeholders.pageKeywords')} />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item
                    label={t('adminSystemSettings.fields.pageDescription')}
                    name="metaDescription"
                    rules={[{ required: true, message: t('adminSystemSettings.placeholders.pageDescription') }]}
                  >
                    <TextArea rows={3} placeholder={t('adminSystemSettings.placeholders.pageDescription')} />
                  </Form.Item>

                  <Divider>{t('adminSystemSettings.cards.openGraph')}</Divider>

                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.ogTitle')}
                        name="ogTitle"
                      >
                        <Input placeholder={t('adminSystemSettings.placeholders.ogTitle')} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.ogImage')}
                      >
                        <Upload {...makeUploadProps(seoForm, 'ogImage')}>
                          <Button icon={<UploadOutlined />}>{t('adminSystemSettings.buttons.uploadOgImage')}</Button>
                        </Upload>
                      </Form.Item>
                      <Form.Item name="ogImage" hidden>
                        <Input />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item
                    label={t('adminSystemSettings.fields.ogDescription')}
                    name="ogDescription"
                  >
                    <TextArea rows={2} placeholder={t('adminSystemSettings.placeholders.ogDescription')} />
                  </Form.Item>

                  <Divider>{t('adminSystemSettings.cards.other')}</Divider>

                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.twitterCardType')}
                        name="twitterCard"
                      >
                        <Select placeholder={t('adminSystemSettings.placeholders.twitterCardType')}>
                          <Option value="summary">{t('adminSystemSettings.twitterCardOptions.summary')}</Option>
                          <Option value="summary_large_image">{t('adminSystemSettings.twitterCardOptions.summary_large_image')}</Option>
                          <Option value="app">{t('adminSystemSettings.twitterCardOptions.app')}</Option>
                          <Option value="player">{t('adminSystemSettings.twitterCardOptions.player')}</Option>
                        </Select>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        label={t('adminSystemSettings.fields.enableSitemap')}
                        name="sitemapEnabled"
                        valuePropName="checked"
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item
                    label={t('adminSystemSettings.fields.robots')}
                    name="robotsTxt"
                    help={t('adminSystemSettings.help.robots')}
                  >
                    <TextArea rows={4} placeholder="User-agent: *&#10;Disallow:" />
                  </Form.Item>

                  <Form.Item
                    label={t('adminSystemSettings.fields.analytics')}
                    name="analyticsCode"
                    help={t('adminSystemSettings.help.analytics')}
                  >
                    <TextArea rows={3} placeholder={t('adminSystemSettings.placeholders.analytics')} />
                  </Form.Item>

                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={loading}>
                      {t('adminSystemSettings.buttons.saveSEO')}
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            )
          },
          {
            key: 'cache',
            label: (
              <span>
                <DatabaseOutlined />
                {t('adminSystemSettings.cards.cache')}
              </span>
            ),
            children: (
              <Row gutter={24}>
                <Col span={16}>
                  <Card title={t('adminSystemSettings.cards.cacheConfig')}>
                    <Form
                      form={cacheForm}
                      layout="vertical"
                      onFinish={handleCacheConfigSave}
                    >
                      <Row gutter={24}>
                        <Col span={12}>
                          <Form.Item
                            label={t('adminSystemSettings.fields.enableCache')}
                            name="enabled"
                            valuePropName="checked"
                          >
                            <Switch />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            label={t('adminSystemSettings.fields.cacheStrategy')}
                            name="strategy"
                            rules={[{ required: true, message: t('adminSystemSettings.placeholders.cacheStrategy') }]}
                          >
                            <Select placeholder={t('adminSystemSettings.placeholders.cacheStrategy')}>
                              <Option value="lru">{t('adminSystemSettings.cacheStrategies.lru')}</Option>
                              <Option value="lfu">{t('adminSystemSettings.cacheStrategies.lfu')}</Option>
                              <Option value="fifo">{t('adminSystemSettings.cacheStrategies.fifo')}</Option>
                              <Option value="ttl">{t('adminSystemSettings.cacheStrategies.ttl')}</Option>
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>

                      <Row gutter={24}>
                        <Col span={12}>
                          <Form.Item
                            label={t('adminSystemSettings.fields.cacheTTL')}
                            name="ttl"
                            rules={[{ required: true, message: t('validation.required') }]}
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
                            label={t('adminSystemSettings.fields.cacheMaxMB')}
                            name="maxSize"
                            rules={[{ required: true, message: t('validation.required') }]}
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
                            {t('adminSystemSettings.buttons.save')}
                          </Button>
                          <Button
                            danger
                            icon={<DeleteOutlined />}
                            onClick={handleClearCache}
                          >
                            {t('adminSystemSettings.buttons.clearCache')}
                          </Button>
                          <Button
                            icon={<ReloadOutlined />}
                            onClick={loadCacheStats}
                          >
                            {t('adminSystemSettings.buttons.refreshStats')}
                          </Button>
                        </Space>
                      </Form.Item>
                    </Form>
                  </Card>
                </Col>

                <Col span={8}>
                  <Card title={t('adminSystemSettings.cards.cacheStats')}>
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <Statistic
                        title={t('adminSystemSettings.stats.hitRate')}
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
                        title={t('adminSystemSettings.stats.totalRequests')}
                        value={cacheStats.totalRequests}
                        prefix={<ClockCircleOutlined />}
                      />

                      <Statistic
                        title={t('adminSystemSettings.stats.cacheSize')}
                        value={cacheStats.cacheSize}
                        suffix="MB"
                        prefix={<DatabaseOutlined />}
                      />

                      <Statistic
                        title={t('adminSystemSettings.stats.memoryUsage')}
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
            )
          }
        ]}
      />
    </div>
  );
};

export default SystemSettings;