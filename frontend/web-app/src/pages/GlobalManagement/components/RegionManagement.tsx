// 区域管理组件
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
  Statistic,
  Progress,
  Tooltip,
  message,
  Popconfirm
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  GlobalOutlined,
  MonitorOutlined,
  WarningOutlined
} from '@ant-design/icons';

const { Option } = Select;

interface Region {
  id: string;
  name: string;
  code: string;
  provider: string;
  location: string;
  status: 'active' | 'inactive' | 'maintenance' | 'error';
  endpoints: {
    api: string;
    cdn: string;
    database: string;
  };
  metrics: {
    cpu: number;
    memory: number;
    storage: number;
    latency: number;
    uptime: number;
  };
  compliance: string[];
  users: number;
  dataVolume: string;
  lastSync: string;
  createdAt: string;
}

const RegionManagement: React.FC = () => {
  const [regions, setRegions] = useState<Region[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRegion, setEditingRegion] = useState<Region | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchRegions();
  }, []);

  const fetchRegions = async () => {
    setLoading(true);
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const mockRegions: Region[] = [
        {
          id: 'ap-east-1',
          name: '亚太东部',
          code: 'APAC-HK',
          provider: 'AWS',
          location: '香港',
          status: 'active',
          endpoints: {
            api: 'https://api-hk.taishanglaojun.com',
            cdn: 'https://cdn-hk.taishanglaojun.com',
            database: 'postgres-hk.taishanglaojun.com'
          },
          metrics: {
            cpu: 65,
            memory: 72,
            storage: 58,
            latency: 45,
            uptime: 99.9
          },
          compliance: ['GDPR', 'PDPA'],
          users: 450000,
          dataVolume: '850 GB',
          lastSync: '2024-01-15 14:30:00',
          createdAt: '2023-06-01'
        },
        {
          id: 'eu-west-1',
          name: '欧洲西部',
          code: 'EMEA-IE',
          provider: 'AWS',
          location: '爱尔兰',
          status: 'active',
          endpoints: {
            api: 'https://api-eu.taishanglaojun.com',
            cdn: 'https://cdn-eu.taishanglaojun.com',
            database: 'postgres-eu.taishanglaojun.com'
          },
          metrics: {
            cpu: 58,
            memory: 68,
            storage: 62,
            latency: 38,
            uptime: 99.8
          },
          compliance: ['GDPR'],
          users: 320000,
          dataVolume: '620 GB',
          lastSync: '2024-01-15 14:28:00',
          createdAt: '2023-06-01'
        },
        {
          id: 'us-east-1',
          name: '北美东部',
          code: 'NA-VA',
          provider: 'AWS',
          location: '弗吉尼亚',
          status: 'active',
          endpoints: {
            api: 'https://api-us.taishanglaojun.com',
            cdn: 'https://cdn-us.taishanglaojun.com',
            database: 'postgres-us.taishanglaojun.com'
          },
          metrics: {
            cpu: 71,
            memory: 75,
            storage: 68,
            latency: 42,
            uptime: 99.7
          },
          compliance: ['CCPA', 'PIPEDA'],
          users: 380000,
          dataVolume: '720 GB',
          lastSync: '2024-01-15 14:32:00',
          createdAt: '2023-06-01'
        },
        {
          id: 'ap-southeast-1',
          name: '亚太东南',
          code: 'APAC-SG',
          provider: 'AWS',
          location: '新加坡',
          status: 'maintenance',
          endpoints: {
            api: 'https://api-sg.taishanglaojun.com',
            cdn: 'https://cdn-sg.taishanglaojun.com',
            database: 'postgres-sg.taishanglaojun.com'
          },
          metrics: {
            cpu: 45,
            memory: 52,
            storage: 48,
            latency: 35,
            uptime: 98.5
          },
          compliance: ['PDPA'],
          users: 180000,
          dataVolume: '340 GB',
          lastSync: '2024-01-15 12:15:00',
          createdAt: '2023-08-15'
        },
        {
          id: 'cn-north-1',
          name: '中国北部',
          code: 'CN-BJ',
          provider: 'Alibaba Cloud',
          location: '北京',
          status: 'active',
          endpoints: {
            api: 'https://api-cn.taishanglaojun.com',
            cdn: 'https://cdn-cn.taishanglaojun.com',
            database: 'postgres-cn.taishanglaojun.com'
          },
          metrics: {
            cpu: 68,
            memory: 71,
            storage: 65,
            latency: 28,
            uptime: 99.6
          },
          compliance: ['PIPL', 'Cybersecurity Law'],
          users: 520000,
          dataVolume: '980 GB',
          lastSync: '2024-01-15 14:35:00',
          createdAt: '2023-09-01'
        },
        {
          id: 'sa-east-1',
          name: '南美东部',
          code: 'SA-BR',
          provider: 'AWS',
          location: '圣保罗',
          status: 'inactive',
          endpoints: {
            api: 'https://api-br.taishanglaojun.com',
            cdn: 'https://cdn-br.taishanglaojun.com',
            database: 'postgres-br.taishanglaojun.com'
          },
          metrics: {
            cpu: 0,
            memory: 0,
            storage: 0,
            latency: 0,
            uptime: 0
          },
          compliance: ['LGPD'],
          users: 0,
          dataVolume: '0 GB',
          lastSync: '2024-01-10 10:00:00',
          createdAt: '2023-12-01'
        }
      ];
      
      setRegions(mockRegions);
    } catch (error) {
      message.error('获取区域列表失败');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'inactive': return 'red';
      case 'maintenance': return 'orange';
      case 'error': return 'red';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '运行中';
      case 'inactive': return '已停用';
      case 'maintenance': return '维护中';
      case 'error': return '错误';
      default: return '未知';
    }
  };

  const handleAddRegion = () => {
    setEditingRegion(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEditRegion = (region: Region) => {
    setEditingRegion(region);
    form.setFieldsValue({
      name: region.name,
      code: region.code,
      provider: region.provider,
      location: region.location,
      status: region.status,
      apiEndpoint: region.endpoints.api,
      cdnEndpoint: region.endpoints.cdn,
      databaseEndpoint: region.endpoints.database,
      compliance: region.compliance
    });
    setModalVisible(true);
  };

  const handleDeleteRegion = async (regionId: string) => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));
      setRegions(regions.filter(r => r.id !== regionId));
      message.success('区域删除成功');
    } catch (error) {
      message.error('删除区域失败');
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      if (editingRegion) {
        // 更新区域
        const updatedRegions = regions.map(r => 
          r.id === editingRegion.id 
            ? {
                ...r,
                name: values.name,
                code: values.code,
                provider: values.provider,
                location: values.location,
                status: values.status,
                endpoints: {
                  api: values.apiEndpoint,
                  cdn: values.cdnEndpoint,
                  database: values.databaseEndpoint
                },
                compliance: values.compliance
              }
            : r
        );
        setRegions(updatedRegions);
        message.success('区域更新成功');
      } else {
        // 添加新区域
        const newRegion: Region = {
          id: `region-${Date.now()}`,
          name: values.name,
          code: values.code,
          provider: values.provider,
          location: values.location,
          status: values.status,
          endpoints: {
            api: values.apiEndpoint,
            cdn: values.cdnEndpoint,
            database: values.databaseEndpoint
          },
          metrics: {
            cpu: 0,
            memory: 0,
            storage: 0,
            latency: 0,
            uptime: 0
          },
          compliance: values.compliance,
          users: 0,
          dataVolume: '0 GB',
          lastSync: new Date().toLocaleString(),
          createdAt: new Date().toISOString().split('T')[0]
        };
        setRegions([...regions, newRegion]);
        message.success('区域添加成功');
      }
      
      setModalVisible(false);
    } catch (error) {
      console.error('Form validation failed:', error);
    }
  };

  const columns = [
    {
      title: '区域信息',
      key: 'region',
      render: (record: Region) => (
        <div>
          <div style={{ fontWeight: 600, fontSize: 16 }}>{record.name}</div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.code} • {record.location}
          </div>
          <div style={{ color: '#666', fontSize: 12 }}>
            {record.provider}
          </div>
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      )
    },
    {
      title: '性能指标',
      key: 'metrics',
      render: (record: Region) => (
        <div style={{ minWidth: 200 }}>
          <Row gutter={8}>
            <Col span={12}>
              <div style={{ fontSize: 12, color: '#666' }}>CPU</div>
              <Progress 
                percent={record.metrics.cpu} 
                size="small" 
                status={record.metrics.cpu > 80 ? 'exception' : 'normal'}
              />
            </Col>
            <Col span={12}>
              <div style={{ fontSize: 12, color: '#666' }}>内存</div>
              <Progress 
                percent={record.metrics.memory} 
                size="small"
                status={record.metrics.memory > 80 ? 'exception' : 'normal'}
              />
            </Col>
          </Row>
          <Row gutter={8} style={{ marginTop: 8 }}>
            <Col span={12}>
              <div style={{ fontSize: 12, color: '#666' }}>延迟: {record.metrics.latency}ms</div>
            </Col>
            <Col span={12}>
              <div style={{ fontSize: 12, color: '#666' }}>可用性: {record.metrics.uptime}%</div>
            </Col>
          </Row>
        </div>
      )
    },
    {
      title: '合规性',
      dataIndex: 'compliance',
      key: 'compliance',
      render: (compliance: string[]) => (
        <div>
          {compliance.map(item => (
            <Tag key={item} color="blue" style={{ marginBottom: 4 }}>
              {item}
            </Tag>
          ))}
        </div>
      )
    },
    {
      title: '用户数据',
      key: 'userData',
      render: (record: Region) => (
        <div>
          <div style={{ fontSize: 14, fontWeight: 500 }}>
            {(record.users / 1000).toFixed(0)}K 用户
          </div>
          <div style={{ fontSize: 12, color: '#666' }}>
            {record.dataVolume}
          </div>
        </div>
      )
    },
    {
      title: '最后同步',
      dataIndex: 'lastSync',
      key: 'lastSync',
      render: (lastSync: string) => (
        <div style={{ fontSize: 12 }}>
          {lastSync}
        </div>
      )
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: Region) => (
        <Space>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />} 
              onClick={() => handleEditRegion(record)}
            />
          </Tooltip>
          <Tooltip title="监控">
            <Button 
              type="text" 
              icon={<MonitorOutlined />}
              onClick={() => message.info('跳转到监控面板')}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个区域吗？"
            onConfirm={() => handleDeleteRegion(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button 
                type="text" 
                danger 
                icon={<DeleteOutlined />}
                disabled={record.status === 'active'}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const activeRegions = regions.filter(r => r.status === 'active').length;
  const totalUsers = regions.reduce((sum, r) => sum + r.users, 0);
  const avgUptime = regions.length > 0 
    ? regions.reduce((sum, r) => sum + r.metrics.uptime, 0) / regions.length 
    : 0;

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="活跃区域"
              value={activeRegions}
              suffix={`/ ${regions.length}`}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="全球用户"
              value={totalUsers}
              prefix={<CloudServerOutlined />}
              formatter={(value) => `${(Number(value) / 1000000).toFixed(1)}M`}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="平均可用性"
              value={avgUptime}
              suffix="%"
              prefix={<MonitorOutlined />}
              precision={1}
              valueStyle={{ color: avgUptime >= 99 ? '#3f8600' : '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="区域列表"
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={fetchRegions}
              loading={loading}
            >
              刷新
            </Button>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={handleAddRegion}
            >
              添加区域
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={regions}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个区域`
          }}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title={editingRegion ? '编辑区域' : '添加区域'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={800}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            status: 'active',
            provider: 'AWS',
            compliance: []
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="区域名称"
                rules={[{ required: true, message: '请输入区域名称' }]}
              >
                <Input placeholder="如：亚太东部" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="code"
                label="区域代码"
                rules={[{ required: true, message: '请输入区域代码' }]}
              >
                <Input placeholder="如：APAC-HK" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="provider"
                label="云服务商"
                rules={[{ required: true, message: '请选择云服务商' }]}
              >
                <Select>
                  <Option value="AWS">Amazon Web Services</Option>
                  <Option value="Azure">Microsoft Azure</Option>
                  <Option value="GCP">Google Cloud Platform</Option>
                  <Option value="Alibaba Cloud">阿里云</Option>
                  <Option value="Tencent Cloud">腾讯云</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="location"
                label="地理位置"
                rules={[{ required: true, message: '请输入地理位置' }]}
              >
                <Input placeholder="如：香港" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              <Option value="active">运行中</Option>
              <Option value="inactive">已停用</Option>
              <Option value="maintenance">维护中</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="compliance"
            label="合规性要求"
            rules={[{ required: true, message: '请选择合规性要求' }]}
          >
            <Select mode="multiple" placeholder="选择适用的法规">
              <Option value="GDPR">GDPR</Option>
              <Option value="CCPA">CCPA</Option>
              <Option value="PIPEDA">PIPEDA</Option>
              <Option value="LGPD">LGPD</Option>
              <Option value="PDPA">PDPA</Option>
              <Option value="PIPL">PIPL</Option>
              <Option value="Cybersecurity Law">网络安全法</Option>
            </Select>
          </Form.Item>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                name="apiEndpoint"
                label="API端点"
                rules={[{ required: true, message: '请输入API端点' }]}
              >
                <Input placeholder="https://api-region.example.com" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="cdnEndpoint"
                label="CDN端点"
                rules={[{ required: true, message: '请输入CDN端点' }]}
              >
                <Input placeholder="https://cdn-region.example.com" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="databaseEndpoint"
                label="数据库端点"
                rules={[{ required: true, message: '请输入数据库端点' }]}
              >
                <Input placeholder="postgres-region.example.com" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default RegionManagement;