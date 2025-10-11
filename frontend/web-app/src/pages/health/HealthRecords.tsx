import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Button, 
  Typography, 
  Space, 
  Tag, 
  Table, 
  Timeline, 
  List, 
  Avatar, 
  Tooltip,
  Badge,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  Upload,
  message,
  Tabs,
  Alert,
  Divider,
  Descriptions,
  Progress,
  Statistic,
  Empty,
  Drawer
} from 'antd';
import { 
  FileTextOutlined, 
  UserOutlined, 
  HeartOutlined, 
  MedicineBoxOutlined,
  CalendarOutlined,
  UploadOutlined,
  DownloadOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  SearchOutlined,
  FilterOutlined,
  ShareAltOutlined,
  PrinterOutlined,
  HistoryOutlined,
  SafetyCertificateOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  FileAddOutlined,
  FolderOutlined,
  StarOutlined,
  WarningOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { UploadProps } from 'antd';
import moment from 'moment';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { RangePicker } = DatePicker;

interface HealthRecord {
  id: string;
  type: 'medical' | 'examination' | 'medication' | 'allergy' | 'surgery' | 'vaccination' | 'family_history';
  title: string;
  description: string;
  date: Date;
  doctor?: string;
  hospital?: string;
  department?: string;
  diagnosis?: string;
  treatment?: string;
  medication?: string;
  dosage?: string;
  duration?: string;
  severity?: 'mild' | 'moderate' | 'severe';
  status: 'active' | 'resolved' | 'chronic' | 'monitoring';
  attachments?: string[];
  tags: string[];
  isImportant: boolean;
  privacy: 'private' | 'family' | 'doctor' | 'public';
  createdAt: Date;
  updatedAt: Date;
}

interface VitalSigns {
  date: Date;
  bloodPressure: { systolic: number; diastolic: number };
  heartRate: number;
  temperature: number;
  weight: number;
  height: number;
  bmi: number;
}

interface LabResult {
  id: string;
  testName: string;
  value: number;
  unit: string;
  referenceRange: string;
  status: 'normal' | 'high' | 'low' | 'critical';
  date: Date;
  lab: string;
}

const HealthRecords: React.FC = () => {
  const [healthRecords, setHealthRecords] = useState<HealthRecord[]>([]);
  const [vitalSigns, setVitalSigns] = useState<VitalSigns[]>([]);
  const [labResults, setLabResults] = useState<LabResult[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<HealthRecord | null>(null);
  const [recordModalVisible, setRecordModalVisible] = useState(false);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [activeTab, setActiveTab] = useState('records');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [dateRange, setDateRange] = useState<[moment.Moment, moment.Moment] | null>(null);
  const [loading, setLoading] = useState(true);
  const [form] = Form.useForm();

  // 记录类型配置
  const recordTypes = [
    { key: 'all', name: '全部记录', icon: <FileTextOutlined />, color: '#1890ff' },
    { key: 'medical', name: '就医记录', icon: <MedicineBoxOutlined />, color: '#52c41a' },
    { key: 'examination', name: '体检报告', icon: <SafetyCertificateOutlined />, color: '#faad14' },
    { key: 'medication', name: '用药记录', icon: <MedicineBoxOutlined />, color: '#722ed1' },
    { key: 'allergy', name: '过敏史', icon: <WarningOutlined />, color: '#ff4d4f' },
    { key: 'surgery', name: '手术记录', icon: <MedicineBoxOutlined />, color: '#13c2c2' },
    { key: 'vaccination', name: '疫苗接种', icon: <SafetyCertificateOutlined />, color: '#52c41a' },
    { key: 'family_history', name: '家族病史', icon: <UserOutlined />, color: '#fa8c16' }
  ];

  useEffect(() => {
    loadHealthData();
  }, []);

  const loadHealthData = () => {
    setLoading(true);
    // 模拟数据加载
    setTimeout(() => {
      const mockRecords: HealthRecord[] = [
        {
          id: '1',
          type: 'medical',
          title: '感冒就诊',
          description: '因发热、咳嗽就诊，诊断为普通感冒',
          date: new Date('2024-01-15'),
          doctor: '张医生',
          hospital: '市人民医院',
          department: '内科',
          diagnosis: '普通感冒',
          treatment: '对症治疗',
          medication: '感冒灵颗粒',
          dosage: '1袋/次，3次/日',
          duration: '7天',
          severity: 'mild',
          status: 'resolved',
          attachments: ['prescription.pdf'],
          tags: ['感冒', '发热', '咳嗽'],
          isImportant: false,
          privacy: 'private',
          createdAt: new Date('2024-01-15'),
          updatedAt: new Date('2024-01-15')
        },
        {
          id: '2',
          type: 'examination',
          title: '年度体检',
          description: '2024年度全面健康体检',
          date: new Date('2024-01-10'),
          doctor: '李医生',
          hospital: '体检中心',
          department: '体检科',
          diagnosis: '整体健康状况良好',
          status: 'resolved',
          attachments: ['checkup_report.pdf', 'blood_test.pdf'],
          tags: ['体检', '血常规', '心电图'],
          isImportant: true,
          privacy: 'family',
          createdAt: new Date('2024-01-10'),
          updatedAt: new Date('2024-01-10')
        },
        {
          id: '3',
          type: 'allergy',
          title: '青霉素过敏',
          description: '对青霉素类抗生素过敏，会出现皮疹',
          date: new Date('2020-05-20'),
          severity: 'moderate',
          status: 'chronic',
          tags: ['过敏', '青霉素', '皮疹'],
          isImportant: true,
          privacy: 'doctor',
          createdAt: new Date('2020-05-20'),
          updatedAt: new Date('2020-05-20')
        },
        {
          id: '4',
          type: 'vaccination',
          title: 'COVID-19疫苗接种',
          description: '新冠疫苗第三针加强针接种',
          date: new Date('2023-12-01'),
          hospital: '社区卫生服务中心',
          medication: '新冠疫苗(mRNA)',
          status: 'resolved',
          tags: ['疫苗', 'COVID-19', '加强针'],
          isImportant: true,
          privacy: 'public',
          createdAt: new Date('2023-12-01'),
          updatedAt: new Date('2023-12-01')
        },
        {
          id: '5',
          type: 'family_history',
          title: '家族高血压史',
          description: '父亲有高血压病史，需要定期监测血压',
          date: new Date('2024-01-01'),
          status: 'monitoring',
          tags: ['家族史', '高血压', '遗传'],
          isImportant: true,
          privacy: 'family',
          createdAt: new Date('2024-01-01'),
          updatedAt: new Date('2024-01-01')
        }
      ];

      const mockVitalSigns: VitalSigns[] = [
        {
          date: new Date('2024-01-15'),
          bloodPressure: { systolic: 120, diastolic: 80 },
          heartRate: 72,
          temperature: 36.5,
          weight: 70,
          height: 175,
          bmi: 22.9
        },
        {
          date: new Date('2024-01-10'),
          bloodPressure: { systolic: 118, diastolic: 78 },
          heartRate: 68,
          temperature: 36.3,
          weight: 69.5,
          height: 175,
          bmi: 22.7
        }
      ];

      const mockLabResults: LabResult[] = [
        {
          id: '1',
          testName: '血糖',
          value: 5.2,
          unit: 'mmol/L',
          referenceRange: '3.9-6.1',
          status: 'normal',
          date: new Date('2024-01-10'),
          lab: '市人民医院检验科'
        },
        {
          id: '2',
          testName: '总胆固醇',
          value: 4.8,
          unit: 'mmol/L',
          referenceRange: '<5.2',
          status: 'normal',
          date: new Date('2024-01-10'),
          lab: '市人民医院检验科'
        },
        {
          id: '3',
          testName: '白细胞计数',
          value: 6.5,
          unit: '×10⁹/L',
          referenceRange: '3.5-9.5',
          status: 'normal',
          date: new Date('2024-01-10'),
          lab: '市人民医院检验科'
        }
      ];

      setHealthRecords(mockRecords);
      setVitalSigns(mockVitalSigns);
      setLabResults(mockLabResults);
      setLoading(false);
    }, 1000);
  };

  // 过滤记录
  const filteredRecords = healthRecords.filter(record => {
    const typeMatch = selectedType === 'all' || record.type === selectedType;
    const dateMatch = !dateRange || (
      moment(record.date).isSameOrAfter(dateRange[0], 'day') &&
      moment(record.date).isSameOrBefore(dateRange[1], 'day')
    );
    return typeMatch && dateMatch;
  });

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return '#1890ff';
      case 'resolved': return '#52c41a';
      case 'chronic': return '#faad14';
      case 'monitoring': return '#722ed1';
      default: return '#d9d9d9';
    }
  };

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '进行中';
      case 'resolved': return '已解决';
      case 'chronic': return '慢性';
      case 'monitoring': return '监测中';
      default: return '未知';
    }
  };

  // 获取严重程度颜色
  const getSeverityColor = (severity?: string) => {
    switch (severity) {
      case 'mild': return '#52c41a';
      case 'moderate': return '#faad14';
      case 'severe': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取检验结果状态颜色
  const getLabStatusColor = (status: string) => {
    switch (status) {
      case 'normal': return '#52c41a';
      case 'high': return '#faad14';
      case 'low': return '#1890ff';
      case 'critical': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 表格列定义
  const recordColumns: ColumnsType<HealthRecord> = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => {
        const typeInfo = recordTypes.find(t => t.key === type);
        return (
          <Tag color={typeInfo?.color} icon={typeInfo?.icon}>
            {typeInfo?.name}
          </Tag>
        );
      }
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: HealthRecord) => (
        <Space>
          <Text strong>{title}</Text>
          {record.isImportant && <StarOutlined style={{ color: '#faad14' }} />}
        </Space>
      )
    },
    {
      title: '日期',
      dataIndex: 'date',
      key: 'date',
      width: 120,
      render: (date: Date) => moment(date).format('YYYY-MM-DD')
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      )
    },
    {
      title: '医院/医生',
      key: 'medical_info',
      width: 150,
      render: (_, record: HealthRecord) => (
        <div>
          {record.hospital && <div><Text type="secondary">{record.hospital}</Text></div>}
          {record.doctor && <div><Text>{record.doctor}</Text></div>}
        </div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record: HealthRecord) => (
        <Space>
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              onClick={() => {
                setSelectedRecord(record);
                setDetailDrawerVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedRecord(record);
                form.setFieldsValue({
                  ...record,
                  date: moment(record.date)
                });
                setRecordModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button 
              type="text" 
              danger 
              icon={<DeleteOutlined />}
              onClick={() => {
                Modal.confirm({
                  title: '确认删除',
                  content: '确定要删除这条记录吗？',
                  onOk: () => {
                    setHealthRecords(prev => prev.filter(r => r.id !== record.id));
                    message.success('记录已删除');
                  }
                });
              }}
            />
          </Tooltip>
        </Space>
      )
    }
  ];

  // 检验结果表格列
  const labColumns: ColumnsType<LabResult> = [
    {
      title: '检验项目',
      dataIndex: 'testName',
      key: 'testName'
    },
    {
      title: '结果',
      dataIndex: 'value',
      key: 'value',
      render: (value: number, record: LabResult) => (
        <Text strong style={{ color: getLabStatusColor(record.status) }}>
          {value} {record.unit}
        </Text>
      )
    },
    {
      title: '参考范围',
      dataIndex: 'referenceRange',
      key: 'referenceRange'
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getLabStatusColor(status)}>
          {status === 'normal' ? '正常' : 
           status === 'high' ? '偏高' : 
           status === 'low' ? '偏低' : '异常'}
        </Tag>
      )
    },
    {
      title: '检验日期',
      dataIndex: 'date',
      key: 'date',
      render: (date: Date) => moment(date).format('YYYY-MM-DD')
    }
  ];

  // 上传配置
  const uploadProps: UploadProps = {
    name: 'file',
    multiple: true,
    action: '/api/upload',
    onChange(info) {
      const { status } = info.file;
      if (status === 'done') {
        message.success(`${info.file.name} 文件上传成功`);
      } else if (status === 'error') {
        message.error(`${info.file.name} 文件上传失败`);
      }
    }
  };

  // 提交记录
  const handleSubmitRecord = async (values: any) => {
    try {
      const newRecord: HealthRecord = {
        id: selectedRecord?.id || Date.now().toString(),
        ...values,
        date: values.date.toDate(),
        tags: values.tags || [],
        attachments: values.attachments || [],
        createdAt: selectedRecord?.createdAt || new Date(),
        updatedAt: new Date()
      };

      if (selectedRecord) {
        setHealthRecords(prev => 
          prev.map(record => 
            record.id === selectedRecord.id ? newRecord : record
          )
        );
        message.success('记录更新成功');
      } else {
        setHealthRecords(prev => [newRecord, ...prev]);
        message.success('记录添加成功');
      }

      setRecordModalVisible(false);
      setSelectedRecord(null);
      form.resetFields();
    } catch (error) {
      message.error('操作失败，请重试');
    }
  };

  // 渲染统计信息
  const renderStats = () => {
    const totalRecords = healthRecords.length;
    const importantRecords = healthRecords.filter(r => r.isImportant).length;
    const activeRecords = healthRecords.filter(r => r.status === 'active').length;
    const recentRecords = healthRecords.filter(r => 
      moment().diff(moment(r.date), 'days') <= 30
    ).length;

    return (
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Statistic
            title="总记录数"
            value={totalRecords}
            prefix={<FileTextOutlined />}
            valueStyle={{ color: '#1890ff' }}
          />
        </Col>
        <Col span={6}>
          <Statistic
            title="重要记录"
            value={importantRecords}
            prefix={<StarOutlined />}
            valueStyle={{ color: '#faad14' }}
          />
        </Col>
        <Col span={6}>
          <Statistic
            title="进行中"
            value={activeRecords}
            prefix={<ClockCircleOutlined />}
            valueStyle={{ color: '#722ed1' }}
          />
        </Col>
        <Col span={6}>
          <Statistic
            title="近30天"
            value={recentRecords}
            prefix={<CalendarOutlined />}
            valueStyle={{ color: '#52c41a' }}
          />
        </Col>
      </Row>
    );
  };

  // 渲染生命体征
  const renderVitalSigns = () => {
    const latestVitals = vitalSigns[0];
    if (!latestVitals) return <Empty description="暂无生命体征数据" />;

    return (
      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card>
            <Statistic
              title="血压"
              value={`${latestVitals.bloodPressure.systolic}/${latestVitals.bloodPressure.diastolic}`}
              suffix="mmHg"
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="心率"
              value={latestVitals.heartRate}
              suffix="次/分"
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="体重"
              value={latestVitals.weight}
              suffix="kg"
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="体温"
              value={latestVitals.temperature}
              suffix="°C"
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="身高"
              value={latestVitals.height}
              suffix="cm"
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="BMI"
              value={latestVitals.bmi}
              precision={1}
              valueStyle={{ color: '#13c2c2' }}
            />
          </Card>
        </Col>
      </Row>
    );
  };

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>
          <FileTextOutlined style={{ color: '#1890ff', marginRight: '8px' }} />
          健康档案
        </Title>
        <Paragraph>
          完整记录和管理您的健康信息，建立个人健康档案
        </Paragraph>
      </div>

      {/* 统计信息 */}
      <Card style={{ marginBottom: '24px' }}>
        {renderStats()}
      </Card>

      {/* 主要内容 */}
      <Tabs 
        activeKey={activeTab} 
        onChange={setActiveTab}
        items={[
          {
            key: 'records',
            label: '健康记录',
            children: (
              <>
                {/* 筛选和操作栏 */}
                <Card style={{ marginBottom: '16px' }}>
                  <Row justify="space-between" align="middle">
                    <Col>
                      <Space>
                        <Select
                          value={selectedType}
                          onChange={setSelectedType}
                          style={{ width: 150 }}
                          placeholder="选择类型"
                        >
                          {recordTypes.map(type => (
                            <Select.Option key={type.key} value={type.key}>
                              <Space>
                                {type.icon}
                                {type.name}
                              </Space>
                            </Select.Option>
                          ))}
                        </Select>
                        
                        <RangePicker
                          value={dateRange}
                          onChange={setDateRange}
                          placeholder={['开始日期', '结束日期']}
                        />
                      </Space>
                    </Col>
                    <Col>
                      <Space>
                        <Button icon={<PlusOutlined />} type="primary" onClick={() => setRecordModalVisible(true)}>
                          添加记录
                        </Button>
                        <Button icon={<UploadOutlined />}>
                          导入记录
                        </Button>
                        <Button icon={<DownloadOutlined />}>
                          导出记录
                        </Button>
                      </Space>
                    </Col>
                  </Row>
                </Card>

                {/* 记录表格 */}
                <Card>
                  <Table
                    columns={recordColumns}
                    dataSource={filteredRecords}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                      pageSize: 10,
                      showSizeChanger: true,
                      showQuickJumper: true,
                      showTotal: (total) => `共 ${total} 条记录`
                    }}
                  />
                </Card>
              </>
            )
          },
          {
            key: 'vitals',
            label: '生命体征',
            children: (
              <Card title="最新生命体征">
                {renderVitalSigns()}
              </Card>
            )
          },
          {
            key: 'lab_results',
            label: '检验结果',
            children: (
              <Card>
                <Table
                  columns={labColumns}
                  dataSource={labResults}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    pageSize: 10,
                    showSizeChanger: true,
                    showTotal: (total) => `共 ${total} 条结果`
                  }}
                />
              </Card>
            )
          },
          {
            key: 'timeline',
            label: '健康时间线',
            children: (
              <Card>
                <Timeline mode="left">
                  {healthRecords
                    .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
                    .map(record => {
                      const typeInfo = recordTypes.find(t => t.key === record.type);
                      return (
                        <Timeline.Item
                          key={record.id}
                          color={typeInfo?.color}
                          dot={typeInfo?.icon}
                          label={moment(record.date).format('YYYY-MM-DD')}
                        >
                          <Card size="small" style={{ marginBottom: '8px' }}>
                            <Space direction="vertical" style={{ width: '100%' }}>
                              <div>
                                <Text strong>{record.title}</Text>
                                {record.isImportant && <StarOutlined style={{ color: '#faad14', marginLeft: '8px' }} />}
                              </div>
                              <Text type="secondary">{record.description}</Text>
                              {record.hospital && (
                                <Text type="secondary">
                                  <MedicineBoxOutlined /> {record.hospital}
                                  {record.doctor && ` - ${record.doctor}`}
                                </Text>
                              )}
                              <Space wrap>
                                {record.tags.map((tag, index) => (
                                  <Tag key={index} size="small">{tag}</Tag>
                                ))}
                              </Space>
                            </Space>
                          </Card>
                        </Timeline.Item>
                      );
                    })}
                </Timeline>
              </Card>
            )
          }
        ]}
      />

      {/* 添加/编辑记录模态框 */}
      <Modal
        title={selectedRecord ? '编辑记录' : '添加记录'}
        visible={recordModalVisible}
        onCancel={() => {
          setRecordModalVisible(false);
          setSelectedRecord(null);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmitRecord}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="type"
                label="记录类型"
                rules={[{ required: true, message: '请选择记录类型' }]}
              >
                <Select placeholder="请选择记录类型">
                  {recordTypes.slice(1).map(type => (
                    <Select.Option key={type.key} value={type.key}>
                      <Space>
                        {type.icon}
                        {type.name}
                      </Space>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="date"
                label="日期"
                rules={[{ required: true, message: '请选择日期' }]}
              >
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input placeholder="请输入记录标题" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入描述' }]}
          >
            <TextArea rows={3} placeholder="请输入详细描述" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="hospital" label="医院">
                <Input placeholder="请输入医院名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="doctor" label="医生">
                <Input placeholder="请输入医生姓名" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="diagnosis" label="诊断">
                <Input placeholder="请输入诊断结果" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="status" label="状态">
                <Select placeholder="请选择状态">
                  <Select.Option value="active">进行中</Select.Option>
                  <Select.Option value="resolved">已解决</Select.Option>
                  <Select.Option value="chronic">慢性</Select.Option>
                  <Select.Option value="monitoring">监测中</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="severity" label="严重程度">
                <Select placeholder="请选择严重程度">
                  <Select.Option value="mild">轻度</Select.Option>
                  <Select.Option value="moderate">中度</Select.Option>
                  <Select.Option value="severe">重度</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="privacy" label="隐私设置">
                <Select placeholder="请选择隐私设置" defaultValue="private">
                  <Select.Option value="private">仅自己</Select.Option>
                  <Select.Option value="family">家庭成员</Select.Option>
                  <Select.Option value="doctor">医生可见</Select.Option>
                  <Select.Option value="public">公开</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="isImportant" valuePropName="checked">
            <Space>
              <input type="checkbox" />
              <Text>标记为重要记录</Text>
            </Space>
          </Form.Item>

          <Form.Item name="attachments" label="附件">
            <Upload {...uploadProps}>
              <Button icon={<UploadOutlined />}>上传文件</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>

      {/* 记录详情抽屉 */}
      <Drawer
        title="记录详情"
        placement="right"
        width={600}
        visible={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
      >
        {selectedRecord && (
          <div>
            <Descriptions column={1} bordered>
              <Descriptions.Item label="标题">
                <Space>
                  {selectedRecord.title}
                  {selectedRecord.isImportant && <StarOutlined style={{ color: '#faad14' }} />}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="类型">
                {recordTypes.find(t => t.key === selectedRecord.type)?.name}
              </Descriptions.Item>
              <Descriptions.Item label="日期">
                {moment(selectedRecord.date).format('YYYY-MM-DD')}
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedRecord.status)}>
                  {getStatusText(selectedRecord.status)}
                </Tag>
              </Descriptions.Item>
              {selectedRecord.severity && (
                <Descriptions.Item label="严重程度">
                  <Tag color={getSeverityColor(selectedRecord.severity)}>
                    {selectedRecord.severity === 'mild' ? '轻度' :
                     selectedRecord.severity === 'moderate' ? '中度' : '重度'}
                  </Tag>
                </Descriptions.Item>
              )}
              <Descriptions.Item label="描述">
                {selectedRecord.description}
              </Descriptions.Item>
              {selectedRecord.hospital && (
                <Descriptions.Item label="医院">
                  {selectedRecord.hospital}
                </Descriptions.Item>
              )}
              {selectedRecord.doctor && (
                <Descriptions.Item label="医生">
                  {selectedRecord.doctor}
                </Descriptions.Item>
              )}
              {selectedRecord.diagnosis && (
                <Descriptions.Item label="诊断">
                  {selectedRecord.diagnosis}
                </Descriptions.Item>
              )}
              {selectedRecord.treatment && (
                <Descriptions.Item label="治疗">
                  {selectedRecord.treatment}
                </Descriptions.Item>
              )}
              {selectedRecord.medication && (
                <Descriptions.Item label="用药">
                  {selectedRecord.medication}
                  {selectedRecord.dosage && ` - ${selectedRecord.dosage}`}
                  {selectedRecord.duration && ` - ${selectedRecord.duration}`}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="标签">
                <Space wrap>
                  {selectedRecord.tags.map((tag, index) => (
                    <Tag key={index}>{tag}</Tag>
                  ))}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {moment(selectedRecord.createdAt).format('YYYY-MM-DD HH:mm')}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {moment(selectedRecord.updatedAt).format('YYYY-MM-DD HH:mm')}
              </Descriptions.Item>
            </Descriptions>

            {selectedRecord.attachments && selectedRecord.attachments.length > 0 && (
              <div style={{ marginTop: '24px' }}>
                <Title level={5}>附件</Title>
                <List
                  dataSource={selectedRecord.attachments}
                  renderItem={(attachment) => (
                    <List.Item
                      actions={[
                        <Button type="link" icon={<DownloadOutlined />}>下载</Button>,
                        <Button type="link" icon={<EyeOutlined />}>预览</Button>
                      ]}
                    >
                      <List.Item.Meta
                        avatar={<Avatar icon={<FileTextOutlined />} />}
                        title={attachment}
                      />
                    </List.Item>
                  )}
                />
              </div>
            )}
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default HealthRecords;