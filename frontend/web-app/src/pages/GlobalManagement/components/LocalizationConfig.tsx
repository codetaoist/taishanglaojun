// 本地化配置组件
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
  Switch,
  Row,
  Col,
  Tabs,
  Upload,
  message,
  Progress,
  Tooltip,
  Popconfirm,
  InputNumber,
  TimePicker,
  Checkbox
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  UploadOutlined,
  DownloadOutlined,
  TranslationOutlined,
  GlobalOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Option } = Select;
const { TabPane } = Tabs;
const { TextArea } = Input;

interface Language {
  code: string;
  name: string;
  nativeName: string;
  region: string;
  enabled: boolean;
  progress: number;
  totalKeys: number;
  translatedKeys: number;
  lastUpdated: string;
}

interface Currency {
  code: string;
  name: string;
  symbol: string;
  regions: string[];
  enabled: boolean;
  exchangeRate: number;
  lastUpdated: string;
}

interface Timezone {
  id: string;
  name: string;
  offset: string;
  regions: string[];
  enabled: boolean;
}

interface CultureConfig {
  region: string;
  dateFormat: string;
  timeFormat: string;
  numberFormat: string;
  addressFormat: string;
  nameFormat: string;
  phoneFormat: string;
  businessHours: {
    start: string;
    end: string;
    workDays: number[];
  };
  holidays: string[];
  colorMeanings: Record<string, string>;
  tabooTopics: string[];
}

const LocalizationConfig: React.FC = () => {
  const [activeTab, setActiveTab] = useState('languages');
  const [languages, setLanguages] = useState<Language[]>([]);
  const [currencies, setCurrencies] = useState<Currency[]>([]);
  const [timezones, setTimezones] = useState<Timezone[]>([]);
  const [cultures, setCultures] = useState<CultureConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingItem, setEditingItem] = useState<any>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchLocalizationData();
  }, []);

  const fetchLocalizationData = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 语言数据
      setLanguages([
        {
          code: 'zh-CN',
          name: 'Chinese (Simplified)',
          nativeName: '简体中文',
          region: 'CN',
          enabled: true,
          progress: 100,
          totalKeys: 2500,
          translatedKeys: 2500,
          lastUpdated: '2024-01-15 14:30:00'
        },
        {
          code: 'en-US',
          name: 'English (US)',
          nativeName: 'English',
          region: 'US',
          enabled: true,
          progress: 100,
          totalKeys: 2500,
          translatedKeys: 2500,
          lastUpdated: '2024-01-15 14:25:00'
        },
        {
          code: 'ja-JP',
          name: 'Japanese',
          nativeName: '日本語',
          region: 'JP',
          enabled: true,
          progress: 95,
          totalKeys: 2500,
          translatedKeys: 2375,
          lastUpdated: '2024-01-14 16:20:00'
        },
        {
          code: 'ko-KR',
          name: 'Korean',
          nativeName: '한국어',
          region: 'KR',
          enabled: true,
          progress: 92,
          totalKeys: 2500,
          translatedKeys: 2300,
          lastUpdated: '2024-01-14 15:45:00'
        },
        {
          code: 'fr-FR',
          name: 'French',
          nativeName: 'Français',
          region: 'FR',
          enabled: true,
          progress: 88,
          totalKeys: 2500,
          translatedKeys: 2200,
          lastUpdated: '2024-01-13 11:30:00'
        },
        {
          code: 'de-DE',
          name: 'German',
          nativeName: 'Deutsch',
          region: 'DE',
          enabled: true,
          progress: 85,
          totalKeys: 2500,
          translatedKeys: 2125,
          lastUpdated: '2024-01-13 10:15:00'
        },
        {
          code: 'es-ES',
          name: 'Spanish',
          nativeName: 'Español',
          region: 'ES',
          enabled: false,
          progress: 65,
          totalKeys: 2500,
          translatedKeys: 1625,
          lastUpdated: '2024-01-10 14:20:00'
        }
      ]);

      // 货币数据
      setCurrencies([
        {
          code: 'USD',
          name: 'US Dollar',
          symbol: '$',
          regions: ['US', 'CA'],
          enabled: true,
          exchangeRate: 1.0,
          lastUpdated: '2024-01-15 14:30:00'
        },
        {
          code: 'EUR',
          name: 'Euro',
          symbol: '€',
          regions: ['DE', 'FR', 'IT', 'ES'],
          enabled: true,
          exchangeRate: 0.85,
          lastUpdated: '2024-01-15 14:30:00'
        },
        {
          code: 'CNY',
          name: 'Chinese Yuan',
          symbol: '¥',
          regions: ['CN'],
          enabled: true,
          exchangeRate: 7.2,
          lastUpdated: '2024-01-15 14:30:00'
        },
        {
          code: 'JPY',
          name: 'Japanese Yen',
          symbol: '¥',
          regions: ['JP'],
          enabled: true,
          exchangeRate: 110.5,
          lastUpdated: '2024-01-15 14:30:00'
        },
        {
          code: 'KRW',
          name: 'Korean Won',
          symbol: '₩',
          regions: ['KR'],
          enabled: true,
          exchangeRate: 1200.0,
          lastUpdated: '2024-01-15 14:30:00'
        }
      ]);

      // 时区数据
      setTimezones([
        {
          id: 'Asia/Shanghai',
          name: 'China Standard Time',
          offset: '+08:00',
          regions: ['CN'],
          enabled: true
        },
        {
          id: 'America/New_York',
          name: 'Eastern Standard Time',
          offset: '-05:00',
          regions: ['US'],
          enabled: true
        },
        {
          id: 'Europe/London',
          name: 'Greenwich Mean Time',
          offset: '+00:00',
          regions: ['GB'],
          enabled: true
        },
        {
          id: 'Asia/Tokyo',
          name: 'Japan Standard Time',
          offset: '+09:00',
          regions: ['JP'],
          enabled: true
        },
        {
          id: 'Asia/Seoul',
          name: 'Korea Standard Time',
          offset: '+09:00',
          regions: ['KR'],
          enabled: true
        }
      ]);

      // 文化配置数据
      setCultures([
        {
          region: 'zh-CN',
          dateFormat: 'YYYY-MM-DD',
          timeFormat: 'HH:mm:ss',
          numberFormat: '1,234.56',
          addressFormat: '{country} {province} {city} {district} {street} {number}',
          nameFormat: '{lastName} {firstName}',
          phoneFormat: '+86 {area} {number}',
          businessHours: {
            start: '09:00',
            end: '18:00',
            workDays: [1, 2, 3, 4, 5]
          },
          holidays: ['春节', '清明节', '劳动节', '端午节', '中秋节', '国庆节'],
          colorMeanings: {
            red: '吉祥、喜庆',
            gold: '富贵、尊贵',
            white: '纯洁、哀悼'
          },
          tabooTopics: ['政治敏感话题', '个人隐私', '宗教争议']
        },
        {
          region: 'en-US',
          dateFormat: 'MM/DD/YYYY',
          timeFormat: 'h:mm:ss A',
          numberFormat: '1,234.56',
          addressFormat: '{number} {street}, {city}, {state} {zipCode}, {country}',
          nameFormat: '{firstName} {lastName}',
          phoneFormat: '+1 ({area}) {number}',
          businessHours: {
            start: '09:00',
            end: '17:00',
            workDays: [1, 2, 3, 4, 5]
          },
          holidays: ['New Year\'s Day', 'Independence Day', 'Thanksgiving', 'Christmas'],
          colorMeanings: {
            red: 'passion, danger',
            blue: 'trust, stability',
            green: 'nature, money'
          },
          tabooTopics: ['personal income', 'age', 'weight', 'political views']
        }
      ]);
    } catch (error) {
      message.error('获取本地化数据失败');
    } finally {
      setLoading(false);
    }
  };

  const languageColumns = [
    {
      title: '语言',
      key: 'language',
      render: (record: Language) => (
        <div>
          <div style={{ fontWeight: 600 }}>{record.nativeName}</div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.name} ({record.code})
          </div>
        </div>
      )
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      render: (region: string) => <Tag>{region}</Tag>
    },
    {
      title: '翻译进度',
      key: 'progress',
      render: (record: Language) => (
        <div style={{ minWidth: 150 }}>
          <Progress 
            percent={record.progress} 
            size="small"
            status={record.progress === 100 ? 'success' : 'active'}
          />
          <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
            {record.translatedKeys} / {record.totalKeys} 条
          </div>
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? '已启用' : '已禁用'}
        </Tag>
      )
    },
    {
      title: '最后更新',
      dataIndex: 'lastUpdated',
      key: 'lastUpdated',
      render: (date: string) => (
        <div style={{ fontSize: 12 }}>{date}</div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: Language) => (
        <Space>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => handleEditLanguage(record)}
            />
          </Tooltip>
          <Tooltip title="导出翻译">
            <Button 
              type="text" 
              icon={<DownloadOutlined />}
              onClick={() => message.info('导出翻译文件')}
            />
          </Tooltip>
          <Tooltip title="导入翻译">
            <Button 
              type="text" 
              icon={<UploadOutlined />}
              onClick={() => message.info('导入翻译文件')}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const currencyColumns = [
    {
      title: '货币',
      key: 'currency',
      render: (record: Currency) => (
        <div>
          <div style={{ fontWeight: 600 }}>
            {record.symbol} {record.name}
          </div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.code}
          </div>
        </div>
      )
    },
    {
      title: '适用区域',
      dataIndex: 'regions',
      key: 'regions',
      render: (regions: string[]) => (
        <div>
          {regions.map(region => (
            <Tag key={region}>{region}</Tag>
          ))}
        </div>
      )
    },
    {
      title: '汇率 (相对USD)',
      dataIndex: 'exchangeRate',
      key: 'exchangeRate',
      render: (rate: number) => (
        <div style={{ fontWeight: 500 }}>
          {rate === 1 ? '1.0000' : rate.toFixed(4)}
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? '已启用' : '已禁用'}
        </Tag>
      )
    },
    {
      title: '最后更新',
      dataIndex: 'lastUpdated',
      key: 'lastUpdated',
      render: (date: string) => (
        <div style={{ fontSize: 12 }}>{date}</div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: Currency) => (
        <Space>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => handleEditCurrency(record)}
            />
          </Tooltip>
          <Tooltip title="更新汇率">
            <Button 
              type="text" 
              icon={<ReloadOutlined />}
              onClick={() => message.info('更新汇率')}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const timezoneColumns = [
    {
      title: '时区',
      key: 'timezone',
      render: (record: Timezone) => (
        <div>
          <div style={{ fontWeight: 600 }}>{record.name}</div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.id}
          </div>
        </div>
      )
    },
    {
      title: '偏移量',
      dataIndex: 'offset',
      key: 'offset',
      render: (offset: string) => (
        <Tag color="blue">{offset}</Tag>
      )
    },
    {
      title: '适用区域',
      dataIndex: 'regions',
      key: 'regions',
      render: (regions: string[]) => (
        <div>
          {regions.map(region => (
            <Tag key={region}>{region}</Tag>
          ))}
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? '已启用' : '已禁用'}
        </Tag>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: Timezone) => (
        <Space>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => handleEditTimezone(record)}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const cultureColumns = [
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      render: (region: string) => (
        <Tag color="blue">{region}</Tag>
      )
    },
    {
      title: '日期格式',
      dataIndex: 'dateFormat',
      key: 'dateFormat'
    },
    {
      title: '时间格式',
      dataIndex: 'timeFormat',
      key: 'timeFormat'
    },
    {
      title: '工作时间',
      key: 'businessHours',
      render: (record: CultureConfig) => (
        <div>
          {record.businessHours.start} - {record.businessHours.end}
        </div>
      )
    },
    {
      title: '节假日数量',
      dataIndex: 'holidays',
      key: 'holidays',
      render: (holidays: string[]) => (
        <div>{holidays.length} 个</div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: CultureConfig) => (
        <Space>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => handleEditCulture(record)}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  const handleEditLanguage = (language: Language) => {
    setEditingItem(language);
    form.setFieldsValue(language);
    setModalVisible(true);
  };

  const handleEditCurrency = (currency: Currency) => {
    setEditingItem(currency);
    form.setFieldsValue(currency);
    setModalVisible(true);
  };

  const handleEditTimezone = (timezone: Timezone) => {
    setEditingItem(timezone);
    form.setFieldsValue(timezone);
    setModalVisible(true);
  };

  const handleEditCulture = (culture: CultureConfig) => {
    setEditingItem(culture);
    form.setFieldsValue({
      ...culture,
      businessHoursStart: dayjs(culture.businessHours.start, 'HH:mm'),
      businessHoursEnd: dayjs(culture.businessHours.end, 'HH:mm'),
      workDays: culture.businessHours.workDays
    });
    setModalVisible(true);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      if (activeTab === 'languages') {
        const updatedLanguages = languages.map(lang => 
          lang.code === editingItem.code ? { ...lang, ...values } : lang
        );
        setLanguages(updatedLanguages);
      } else if (activeTab === 'currencies') {
        const updatedCurrencies = currencies.map(curr => 
          curr.code === editingItem.code ? { ...curr, ...values } : curr
        );
        setCurrencies(updatedCurrencies);
      } else if (activeTab === 'timezones') {
        const updatedTimezones = timezones.map(tz => 
          tz.id === editingItem.id ? { ...tz, ...values } : tz
        );
        setTimezones(updatedTimezones);
      } else if (activeTab === 'cultures') {
        const updatedCultures = cultures.map(culture => 
          culture.region === editingItem.region 
            ? { 
                ...culture, 
                ...values,
                businessHours: {
                  start: values.businessHoursStart.format('HH:mm'),
                  end: values.businessHoursEnd.format('HH:mm'),
                  workDays: values.workDays
                }
              } 
            : culture
        );
        setCultures(updatedCultures);
      }
      
      message.success('更新成功');
      setModalVisible(false);
    } catch (error) {
      console.error('Form validation failed:', error);
    }
  };

  const renderModalContent = () => {
    if (activeTab === 'languages') {
      return (
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="name" label="语言名称">
                <Input disabled />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="code" label="语言代码">
                <Input disabled />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="enabled" label="启用状态" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      );
    } else if (activeTab === 'currencies') {
      return (
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="name" label="货币名称">
                <Input disabled />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="code" label="货币代码">
                <Input disabled />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="exchangeRate" label="汇率 (相对USD)">
                <InputNumber min={0} step={0.0001} precision={4} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="enabled" label="启用状态" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      );
    } else if (activeTab === 'timezones') {
      return (
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="时区名称">
            <Input />
          </Form.Item>
          <Form.Item name="offset" label="偏移量">
            <Input placeholder="+08:00" />
          </Form.Item>
          <Form.Item name="enabled" label="启用状态" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      );
    } else if (activeTab === 'cultures') {
      return (
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="dateFormat" label="日期格式">
                <Input placeholder="YYYY-MM-DD" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="timeFormat" label="时间格式">
                <Input placeholder="HH:mm:ss" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="numberFormat" label="数字格式">
                <Input placeholder="1,234.56" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="phoneFormat" label="电话格式">
                <Input placeholder="+86 {area} {number}" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="addressFormat" label="地址格式">
            <TextArea placeholder="{country} {province} {city} {district} {street} {number}" />
          </Form.Item>
          <Form.Item name="nameFormat" label="姓名格式">
            <Input placeholder="{lastName} {firstName}" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="businessHoursStart" label="工作开始时间">
                <TimePicker format="HH:mm" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="businessHoursEnd" label="工作结束时间">
                <TimePicker format="HH:mm" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="workDays" label="工作日">
            <Checkbox.Group>
              <Checkbox value={1}>周一</Checkbox>
              <Checkbox value={2}>周二</Checkbox>
              <Checkbox value={3}>周三</Checkbox>
              <Checkbox value={4}>周四</Checkbox>
              <Checkbox value={5}>周五</Checkbox>
              <Checkbox value={6}>周六</Checkbox>
              <Checkbox value={0}>周日</Checkbox>
            </Checkbox.Group>
          </Form.Item>
          <Form.Item name="holidays" label="节假日">
            <Select mode="tags" placeholder="输入节假日名称">
              {editingItem?.holidays?.map((holiday: string) => (
                <Option key={holiday} value={holiday}>{holiday}</Option>
              ))}
            </Select>
          </Form.Item>
        </Form>
      );
    }
    return null;
  };

  return (
    <div>
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane 
          tab={
            <span>
              <TranslationOutlined />
              语言管理
            </span>
          } 
          key="languages"
        >
          <Card
            title="支持的语言"
            extra={
              <Space>
                <Button 
                  icon={<ReloadOutlined />} 
                  onClick={fetchLocalizationData}
                  loading={loading}
                >
                  刷新
                </Button>
                <Button type="primary" icon={<PlusOutlined />}>
                  添加语言
                </Button>
              </Space>
            }
          >
            <Table
              columns={languageColumns}
              dataSource={languages}
              rowKey="code"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 种语言`
              }}
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <DollarOutlined />
              货币管理
            </span>
          } 
          key="currencies"
        >
          <Card
            title="支持的货币"
            extra={
              <Space>
                <Button 
                  icon={<ReloadOutlined />} 
                  onClick={() => message.info('更新所有汇率')}
                >
                  更新汇率
                </Button>
                <Button type="primary" icon={<PlusOutlined />}>
                  添加货币
                </Button>
              </Space>
            }
          >
            <Table
              columns={currencyColumns}
              dataSource={currencies}
              rowKey="code"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 种货币`
              }}
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <ClockCircleOutlined />
              时区管理
            </span>
          } 
          key="timezones"
        >
          <Card
            title="支持的时区"
            extra={
              <Button type="primary" icon={<PlusOutlined />}>
                添加时区
              </Button>
            }
          >
            <Table
              columns={timezoneColumns}
              dataSource={timezones}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 个时区`
              }}
            />
          </Card>
        </TabPane>

        <TabPane 
          tab={
            <span>
              <GlobalOutlined />
              文化配置
            </span>
          } 
          key="cultures"
        >
          <Card
            title="文化适配配置"
            extra={
              <Button type="primary" icon={<PlusOutlined />}>
                添加配置
              </Button>
            }
          >
            <Table
              columns={cultureColumns}
              dataSource={cultures}
              rowKey="region"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 个配置`
              }}
            />
          </Card>
        </TabPane>
      </Tabs>

      <Modal
        title={`编辑${activeTab === 'languages' ? '语言' : activeTab === 'currencies' ? '货币' : activeTab === 'timezones' ? '时区' : '文化配置'}`}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={800}
        destroyOnClose
      >
        {renderModalContent()}
      </Modal>
    </div>
  );
};

export default LocalizationConfig;