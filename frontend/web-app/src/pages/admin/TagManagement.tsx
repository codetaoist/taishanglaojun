import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Modal,
  Form,
  message,
  Popconfirm,
  Typography,
  Tag,
  Row,
  Col,
  Statistic,
  Progress,
  Tooltip,
  Select,
  Switch
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TagOutlined,
  FileTextOutlined,
  EyeOutlined,
  HeartOutlined,
  RiseOutlined,
  FireOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';

const { Search } = Input;
const { Title } = Typography;
const { TextArea } = Input;

interface TagData {
  id: string;
  name: string;
  description?: string;
  color?: string;
  usage_count: number;
  wisdom_count: number;
  total_views: number;
  total_likes: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface TagFormData {
  name: string;
  description?: string;
  color?: string;
  is_active: boolean;
}

const TagManagement: React.FC = () => {
  const [tags, setTags] = useState<TagData[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingTag, setEditingTag] = useState<TagData | null>(null);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [form] = Form.useForm();

  // 预定义颜色
  const tagColors = [
    'magenta', 'red', 'volcano', 'orange', 'gold', 
    'lime', 'green', 'cyan', 'blue', 'geekblue', 
    'purple', 'default'
  ];

  // 加载标签列表
  const loadTags = async () => {
    setLoading(true);
    try {
      const response = await apiClient.getTags();
      if (response.success && response.data) {
        setTags(response.data);
      } else {
        message.error('获取标签列表失败');
      }
    } catch (error) {
      console.error('加载标签列表失败:', error);
      message.error('网络错误，请稍后重试');
    }
    setLoading(false);
  };

  // 创建标签
  const handleCreate = () => {
    setEditingTag(null);
    form.resetFields();
    form.setFieldsValue({
      color: 'blue',
      is_active: true
    });
    setModalVisible(true);
  };

  // 编辑标签
  const handleEdit = (tag: TagData) => {
    setEditingTag(tag);
    form.setFieldsValue({
      name: tag.name,
      description: tag.description,
      color: tag.color || 'blue',
      is_active: tag.is_active
    });
    setModalVisible(true);
  };

  // 保存标签
  const handleSave = async (values: TagFormData) => {
    try {
      let response;
      if (editingTag) {
        response = await apiClient.updateTag(editingTag.id, values);
      } else {
        response = await apiClient.createTag(values);
      }

      if (response.success) {
        message.success(editingTag ? '更新成功' : '创建成功');
        setModalVisible(false);
        loadTags();
      } else {
        message.error(editingTag ? '更新失败' : '创建失败');
      }
    } catch {
      message.error('保存标签失败');
    }
  };

  // 删除标签
  const handleDelete = async (id: string) => {
    try {
      const response = await apiClient.deleteTag(id);
      if (response.success) {
        message.success('删除成功');
        loadTags();
      } else {
        message.error('删除标签失败');
      }
    } catch (error) {
      console.error('删除失败:', error);
      message.error('删除失败');
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的标签');
      return;
    }

    try {
      await Promise.all(
        selectedRowKeys.map(id => apiClient.deleteTag(id as string))
      );
      message.success('批量删除成功');
      setSelectedRowKeys([]);
      loadTags();
    } catch (error) {
      console.error('批量删除失败:', error);
      message.error('批量删除标签失败');
    }
  };

  // 搜索标签
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
  };

  // 过滤标签
  const filteredTags = tags.filter(tag =>
    tag.name.toLowerCase().includes(searchKeyword.toLowerCase()) ||
    (tag.description && tag.description.toLowerCase().includes(searchKeyword.toLowerCase()))
  );

  useEffect(() => {
    loadTags();
  }, []);

  // 表格列定义
  const columns = [
    {
      title: '标签名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: TagData) => (
        <Tag color={record.color || 'blue'} icon={<TagOutlined />}>
          {text}
        </Tag>
      )
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text: string) => text || '-'
    },
    {
      title: '使用次数',
      dataIndex: 'usage_count',
      key: 'usage_count',
      width: 100,
      sorter: (a: TagData, b: TagData) => a.usage_count - b.usage_count,
      render: (count: number) => (
        <Tag color="orange" icon={<RiseOutlined />}>
          {count || 0}
        </Tag>
      )
    },
    {
      title: '关联智慧',
      dataIndex: 'wisdom_count',
      key: 'wisdom_count',
      width: 100,
      sorter: (a: TagData, b: TagData) => a.wisdom_count - b.wisdom_count,
      render: (count: number) => (
        <Tag color="blue" icon={<FileTextOutlined />}>
          {count || 0}
        </Tag>
      )
    },
    {
      title: '热度',
      key: 'popularity',
      width: 120,
      render: (record: TagData) => {
        const maxViews = Math.max(...tags.map(t => t.total_views || 0));
        const popularity = maxViews > 0 ? (record.total_views || 0) / maxViews * 100 : 0;
        
        return (
          <div>
            <Progress 
              percent={Math.round(popularity)} 
              size="small" 
              strokeColor={popularity > 70 ? '#f5222d' : popularity > 40 ? '#faad14' : '#52c41a'}
            />
            <div className="text-xs text-gray-500 mt-1">
              <EyeOutlined /> {record.total_views || 0} 
              <HeartOutlined className="ml-2" /> {record.total_likes || 0}
            </div>
          </div>
        );
      }
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 80,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? '启用' : '禁用'}
        </Tag>
      )
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => new Date(date).toLocaleDateString()
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (record: TagData) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这个标签吗？"
            description="删除后相关的智慧内容将失去此标签"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="link"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      )
    }
  ];

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
    getCheckboxProps: (record: TagData) => ({
      disabled: false,
      name: record.name,
    }),
  };

  // 计算统计数据
  const totalTags = tags.length;
  const activeTags = tags.filter(tag => tag.is_active).length;
  const totalUsage = tags.reduce((sum, tag) => sum + (tag.usage_count || 0), 0);
  const totalViews = tags.reduce((sum, tag) => sum + (tag.total_views || 0), 0);

  // 获取热门标签
  const popularTags = [...tags]
    .sort((a, b) => (b.total_views || 0) - (a.total_views || 0))
    .slice(0, 10);

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <Title level={2} className="mb-2">
              标签管理
            </Title>
            <p className="text-gray-600 mb-0">
              管理智慧内容的标签体系，提升内容分类和检索效率
            </p>
          </div>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            添加标签
          </Button>
        </div>
      </Card>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="总标签数"
              value={totalTags}
              prefix={<TagOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="启用标签"
              value={activeTags}
              prefix={<TagOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="总使用次数"
              value={totalUsage}
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title="总阅读量"
              value={totalViews}
              prefix={<EyeOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 热门标签 */}
        <Col xs={24} lg={8}>
          <Card title={<><FireOutlined /> 热门标签</>} className="h-96">
            <div className="space-y-3">
              {popularTags.map((tag, index) => (
                <div key={tag.id} className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <span className="text-gray-500 w-6">#{index + 1}</span>
                    <Tag color={tag.color || 'blue'}>{tag.name}</Tag>
                  </div>
                  <div className="text-sm text-gray-500">
                    <EyeOutlined /> {tag.total_views || 0}
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </Col>

        {/* 标签列表 */}
        <Col xs={24} lg={16}>
          <Card 
            title="标签列表"
            extra={
              <Space>
                <Search
                  placeholder="搜索标签名称或描述"
                  allowClear
                  onSearch={handleSearch}
                  style={{ width: 250 }}
                />
                {selectedRowKeys.length > 0 && (
                  <Popconfirm
                    title={`确定要删除选中的 ${selectedRowKeys.length} 个标签吗？`}
                    onConfirm={handleBatchDelete}
                    okText="确定"
                    cancelText="取消"
                  >
                    <Button danger>
                      批量删除 ({selectedRowKeys.length})
                    </Button>
                  </Popconfirm>
                )}
              </Space>
            }
          >
            <Table
              columns={columns}
              dataSource={filteredTags}
              rowKey="id"
              loading={loading}
              rowSelection={rowSelection}
              pagination={{
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total, range) => 
                  `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
              }}
            />
          </Card>
        </Col>
      </Row>

      {/* 编辑模态框 */}
      <Modal
        title={editingTag ? '编辑标签' : '添加标签'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
        >
          <Form.Item
            name="name"
            label="标签名称"
            rules={[{ required: true, message: '请输入标签名称' }]}
          >
            <Input placeholder="请输入标签名称" />
          </Form.Item>

          <Form.Item name="description" label="标签描述">
            <TextArea
              placeholder="请输入标签描述"
              rows={3}
              showCount
              maxLength={200}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="color" label="标签颜色">
                <Select placeholder="选择标签颜色">
                  {tagColors.map(color => (
                    <Select.Option key={color} value={color}>
                      <Tag color={color}>{color}</Tag>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="is_active" label="状态" valuePropName="checked">
                <Switch checkedChildren="启用" unCheckedChildren="禁用" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default TagManagement;