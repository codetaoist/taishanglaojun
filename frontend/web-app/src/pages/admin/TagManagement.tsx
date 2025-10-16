import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
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
import { formatDate } from '../../utils/display';

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
  const { t } = useTranslation();

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
        message.error(t('adminTag.messages.fetchFailure'));
      }
    } catch (error) {
      console.error('加载标签列表失败:', error);
      message.error(t('adminTag.messages.networkError'));
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
        message.success(editingTag ? t('adminTag.messages.updateSuccess') : t('adminTag.messages.createSuccess'));
        setModalVisible(false);
        loadTags();
      } else {
        message.error(editingTag ? t('adminTag.messages.updateFailure') : t('adminTag.messages.createFailure'));
      }
    } catch {
      message.error(t('adminTag.messages.saveFailure'));
    }
  };

  // 删除标签
  const handleDelete = async (id: string) => {
    try {
      const response = await apiClient.deleteTag(id);
      if (response.success) {
        message.success(t('adminTag.messages.deleteSuccess'));
        loadTags();
      } else {
        message.error(t('adminTag.messages.deleteFailure'));
      }
    } catch (error) {
      console.error('删除失败:', error);
      message.error(t('adminTag.messages.deleteFailure'));
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning(t('adminTag.messages.batchDeleteNoneSelected'));
      return;
    }

    try {
      await Promise.all(
        selectedRowKeys.map(id => apiClient.deleteTag(id as string))
      );
      message.success(t('adminTag.messages.batchDeleteSuccess'));
      setSelectedRowKeys([]);
      loadTags();
    } catch (error) {
      console.error('批量删除失败:', error);
      message.error(t('adminTag.messages.batchDeleteFailure'));
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
      title: t('adminTag.table.name'),
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: TagData) => (
        <Tag color={record.color || 'blue'} icon={<TagOutlined />}>
          {text}
        </Tag>
      )
    },
    {
      title: t('adminTag.table.description'),
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text?: string) => (
        text ? (
          <Tooltip title={text}>
            <Typography.Text ellipsis style={{ maxWidth: 240 }}>{text}</Typography.Text>
          </Tooltip>
        ) : (
          <span>-</span>
        )
      )
    },
    {
      title: t('adminTag.table.usageCount'),
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
      title: t('adminTag.table.wisdomCount'),
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
      title: t('adminTag.table.popularity'),
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
      title: t('adminTag.table.status'),
      dataIndex: 'is_active',
      key: 'is_active',
      width: 80,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'red'}>
          {isActive ? t('adminTag.status.enabled') : t('adminTag.status.disabled')}
        </Tag>
      )
    },
    {
      title: t('adminTag.table.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => formatDate(date)
    },
    {
      title: t('adminTag.table.actions'),
      key: 'actions',
      width: 150,
      render: (record: TagData) => (
        <Space>
          <Tooltip title={t('adminTag.tooltips.edit')}>
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title={t('adminTag.confirm.deleteTitle')}
            description={t('adminTag.confirm.deleteDesc')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('adminTag.confirm.ok')}
            cancelText={t('adminTag.confirm.cancel')}
          >
            <Tooltip title={t('adminTag.tooltips.delete')}>
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
              {t('adminTag.title')}
            </Title>
            <p className="text-gray-600 mb-0">
              {t('adminTag.subtitle')}
            </p>
          </div>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            {t('adminTag.modal.addTitle')}
          </Button>
        </div>
      </Card>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminTag.stats.totalTags')}
              value={totalTags}
              prefix={<TagOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminTag.stats.activeTags')}
              value={activeTags}
              prefix={<TagOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminTag.stats.totalUsage')}
              value={totalUsage}
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminTag.stats.totalViews')}
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
          <Card title={<><FireOutlined /> {t('adminTag.cards.hotTags')}</>} className="h-96">
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
            title={t('adminTag.cards.tagList')}
            extra={
              <Space>
                <Search
                  placeholder={t('adminTag.searchPlaceholder')}
                  allowClear
                  onSearch={handleSearch}
                  style={{ width: 250 }}
                />
                {selectedRowKeys.length > 0 && (
                  <Popconfirm
                    title={t('adminTag.confirm.batchDeleteTitle', { count: selectedRowKeys.length })}
                    onConfirm={handleBatchDelete}
                    okText={t('adminTag.confirm.ok')}
                    cancelText={t('adminTag.confirm.cancel')}
                  >
                    <Button danger>
                      {t('adminTag.actions.batchDelete')} ({selectedRowKeys.length})
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
                  t('adminTag.pagination.total', { start: range[0], end: range[1], total })
              }}
            />
          </Card>
        </Col>
      </Row>

      {/* 编辑模态框 */}
      <Modal
        title={editingTag ? t('adminTag.modal.editTitle') : t('adminTag.modal.addTitle')}
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
            label={t('adminTag.form.nameLabel')}
            rules={[{ required: true, message: t('adminTag.form.nameRequired') }]}
          >
            <Input placeholder={t('adminTag.form.namePlaceholder')} />
          </Form.Item>

          <Form.Item name="description" label={t('adminTag.form.descriptionLabel')}>
            <TextArea
              placeholder={t('adminTag.form.descriptionPlaceholder')}
              rows={3}
              showCount
              maxLength={200}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="color" label={t('adminTag.form.colorLabel')}>
                <Select placeholder={t('adminTag.form.colorPlaceholder')}>
                  {tagColors.map(color => (
                    <Select.Option key={color} value={color}>
                      <Tag color={color}>{color}</Tag>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="is_active" label={t('adminTag.form.statusLabel')} valuePropName="checked">
                <Switch checkedChildren={t('adminTag.form.switchEnabled')} unCheckedChildren={t('adminTag.form.switchDisabled')} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default TagManagement;