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
  Tree,
  Row,
  Col,
  Statistic,
  Tooltip
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FolderOutlined,
  FileTextOutlined,
  EyeOutlined,
  HeartOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';
import { formatDate } from '../../utils/display';

const { Search } = Input;
const { Title } = Typography;
const { TextArea } = Input;

interface Category {
  id: string;
  name: string;
  description?: string;
  parent_id?: string;
  sort_order: number;
  is_active: boolean;
  wisdom_count?: number;
  total_views?: number;
  total_likes?: number;
  created_at: string;
  updated_at: string;
  children?: Category[];
}

interface CategoryFormData {
  name: string;
  description?: string;
  parent_id?: string;
  sort_order: number;
  is_active: boolean;
}

const CategoryManagement: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [form] = Form.useForm();
  const { t } = useTranslation();

  // 加载分类列表
  const loadCategories = async () => {
    setLoading(true);
    try {
      const response = await apiClient.getCategories();
      if (response.success && response.data) {
        setCategories(response.data);
      } else {
        message.error('获取分类列表失败');
      }
    } catch {
      console.error('加载分类列表失败');
      message.error('网络错误，请稍后重试');
    }
    setLoading(false);
  };

  // 创建分类
  const handleCreate = () => {
    setEditingCategory(null);
    form.resetFields();
    form.setFieldsValue({
      sort_order: 0,
      is_active: true
    });
    setModalVisible(true);
  };

  // 编辑分类
  const handleEdit = (category: Category) => {
    setEditingCategory(category);
    form.setFieldsValue({
      name: category.name,
      description: category.description,
      parent_id: category.parent_id,
      sort_order: category.sort_order,
      is_active: category.is_active
    });
    setModalVisible(true);
  };

  // 保存分类
  const handleSave = async (values: CategoryFormData) => {
    try {
      let response;
      if (editingCategory) {
        response = await apiClient.updateCategory(editingCategory.id, values);
      } else {
        response = await apiClient.createCategory(values);
      }

      if (response.success) {
        message.success(editingCategory ? '更新成功' : '创建成功');
        setModalVisible(false);
        loadCategories();
      } else {
        message.error(editingCategory ? '更新失败' : '创建失败');
      }
    } catch {
      console.error('保存失败');
      message.error('保存失败');
    }
  };

  // 删除分类
  const handleDelete = async (id: string) => {
    try {
      const response = await apiClient.deleteCategory(id);
      if (response.success) {
        message.success('删除成功');
        loadCategories();
      } else {
        message.error('删除失败');
      }
    } catch {
      console.error('删除失败');
      message.error('删除失败');
    }
  };

  // 搜索分类
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
  };

  // 过滤分类
  const filteredCategories = categories.filter(category =>
    category.name.toLowerCase().includes(searchKeyword.toLowerCase()) ||
    (category.description && category.description.toLowerCase().includes(searchKeyword.toLowerCase()))
  );

  useEffect(() => {
    loadCategories();
  }, []);

  // 构建树形数据
  const buildTreeData = (categories: Category[]): TreeNode[] => {
    const categoryMap = new Map();
    const roots: TreeNode[] = [];

    // 创建映射
    categories.forEach(category => {
      categoryMap.set(category.id, {
        key: category.id,
        title: (
          <div className="flex items-center justify-between">
            <span>{category.name}</span>
            <Space>
              <Tag color={category.is_active ? 'green' : 'red'}>
                {category.is_active ? '启用' : '禁用'}
              </Tag>
              <Tag color="blue">{category.wisdom_count || 0}</Tag>
            </Space>
          </div>
        ),
        children: [],
        data: category
      });
    });

    // 构建树形结构
    categories.forEach(category => {
      const node = categoryMap.get(category.id);
      if (category.parent_id && categoryMap.has(category.parent_id)) {
        categoryMap.get(category.parent_id).children.push(node);
      } else {
        roots.push(node);
      }
    });

    return roots;
  };

  // 表格列定义
  const columns = [
    {
      title: '分类名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Category) => (
        <Space>
          <FolderOutlined />
          <span>{text}</span>
          {!record.is_active && <Tag color="red">已禁用</Tag>}
        </Space>
      )
    },
    {
      title: '描述',
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
      title: '排序',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 80,
      sorter: (a: Category, b: Category) => a.sort_order - b.sort_order
    },
    {
      title: '智慧数量',
      dataIndex: 'wisdom_count',
      key: 'wisdom_count',
      width: 100,
      render: (count: number) => (
        <Tag color="blue" icon={<FileTextOutlined />}>
          {count || 0}
        </Tag>
      )
    },
    {
      title: '统计',
      key: 'stats',
      width: 120,
      render: (record: Category) => (
        <Space direction="vertical" size="small">
          <span>
            <EyeOutlined /> {record.total_views || 0}
          </span>
          <span>
            <HeartOutlined /> {record.total_likes || 0}
          </span>
        </Space>
      )
    },
    {
      title: t('adminCategory.table.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => formatDate(date)
    },
    {
      title: t('adminCategory.table.actions'),
      key: 'actions',
      width: 150,
      render: (record: Category) => (
        <Space>
          <Tooltip title={t('adminCategory.actions.edit')}>
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title={t('adminCategory.confirm.deleteTitle')}
            description={t('adminCategory.confirm.deleteDescription')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('adminCategory.confirm.ok')}
            cancelText={t('adminCategory.confirm.cancel')}
          >
            <Tooltip title={t('adminCategory.actions.delete')}>
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

  // 计算统计数据
  const totalCategories = categories.length;
  const activeCategories = categories.filter(cat => cat.is_active).length;
  const totalWisdom = categories.reduce((sum, cat) => sum + (cat.wisdom_count || 0), 0);
  const totalViews = categories.reduce((sum, cat) => sum + (cat.total_views || 0), 0);

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <Title level={2} className="mb-2">
              {t('adminCategory.title')}
            </Title>
            <p className="text-gray-600 mb-0">
              {t('adminCategory.subtitle')}
            </p>
          </div>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            {t('adminCategory.actions.addCategory')}
          </Button>
        </div>
      </Card>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminCategory.stats.totalCategories')}
              value={totalCategories}
              prefix={<FolderOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminCategory.stats.activeCategories')}
              value={activeCategories}
              prefix={<FolderOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminCategory.stats.totalWisdom')}
              value={totalWisdom}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={6}>
          <Card>
            <Statistic
              title={t('adminCategory.stats.totalViews')}
              value={totalViews}
              prefix={<EyeOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 分类树 */}
        <Col xs={24} lg={8}>
          <Card title={t('adminCategory.treeTitle')} className="h-96">
            <Tree
              treeData={buildTreeData(filteredCategories)}
              defaultExpandAll
              showLine
              showIcon
            />
          </Card>
        </Col>

        {/* 分类列表 */}
        <Col xs={24} lg={16}>
          <Card 
            title={t('adminCategory.listTitle')}
            extra={
              <Search
                placeholder={t('adminCategory.searchPlaceholder')}
                allowClear
                onSearch={handleSearch}
                style={{ width: 250 }}
              />
            }
          >
            <Table
              columns={columns}
              dataSource={filteredCategories}
              rowKey="id"
              loading={loading}
              pagination={{
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total, range) => 
                  t('adminCategory.pagination.total', { start: range[0], end: range[1], total })
              }}
            />
          </Card>
        </Col>
      </Row>

      {/* 编辑模态框 */}
      <Modal
        title={editingCategory ? t('adminCategory.modal.editTitle') : t('adminCategory.modal.addTitle')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
        >
          <Form.Item
            name="name"
            label="分类名称"
            rules={[{ required: true, message: '请输入分类名称' }]}
          >
            <Input placeholder="请输入分类名称" />
          </Form.Item>

          <Form.Item name="description" label="分类描述">
            <TextArea
              placeholder="请输入分类描述"
              rows={3}
              showCount
              maxLength={200}
            />
          </Form.Item>

          <Form.Item name="parent_id" label="父分类">
            <Select placeholder="选择父分类（可选）" allowClear>
              {categories
                .filter(cat => !editingCategory || cat.id !== editingCategory.id)
                .map(cat => (
                  <Select.Option key={cat.id} value={cat.id}>
                    {cat.name}
                  </Select.Option>
                ))}
            </Select>
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="sort_order"
                label="排序"
                rules={[{ required: true, message: '请输入排序值' }]}
              >
                <Input type="number" placeholder="排序值" />
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

export default CategoryManagement;

// 定义树节点类型
interface TreeNode {
  key: string;
  title: React.ReactNode;
  children?: TreeNode[];
}