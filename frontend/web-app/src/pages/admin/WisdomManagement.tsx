import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Select,
  Tag,
  message,
  Popconfirm,
  Typography,
  Row,
  Col,
  DatePicker,
  Form,
  Drawer
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  EyeOutlined,
  HeartOutlined,
  FilterOutlined,
  ExportOutlined,
  ImportOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../services/api';
import type { CulturalWisdom } from '../../types';

const { Search } = Input;
const { Option } = Select;
const { Title } = Typography;
const { RangePicker } = DatePicker;

interface WisdomFilters {
  keyword?: string;
  category?: string;
  school?: string;
  author?: string;
  dateRange?: [string, string];
  status?: string;
}

const WisdomManagement: React.FC = () => {
  const navigate = useNavigate();
  const [wisdomList, setWisdomList] = useState<CulturalWisdom[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [filters, setFilters] = useState<WisdomFilters>({});
  const [categories, setCategories] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [schools, setSchools] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [filterDrawerVisible, setFilterDrawerVisible] = useState(false);
  const [form] = Form.useForm();

  // 加载智慧列表
  const loadWisdomList = useCallback(async (page = currentPage, size = pageSize, searchFilters = filters) => {
    setLoading(true);
    try {
      const params = {
        page,
        pageSize: size,
        ...searchFilters
      };

      const response = await apiClient.getWisdomList(params);
      
      if (response.success && response.data) {
        setWisdomList(response.data.items || []);
        setTotal(response.data.total || 0);
      } else {
        message.error('获取智慧列表失败');
      }
    } catch {
      message.error('加载智慧列表失败');
    }
    setLoading(false);
  }, [currentPage, pageSize, filters]);

  // 加载分类和学派数据
  const loadFilterOptions = async () => {
    try {
      const [categoriesRes, schoolsRes] = await Promise.all([
        apiClient.getCategories(),
        apiClient.getSchools()
      ]);

      if (categoriesRes.success) {
        setCategories(categoriesRes.data || []);
      }
      
      if (schoolsRes.success) {
        setSchools(schoolsRes.data || []);
      }
    } catch (error) {
      console.error('加载筛选选项失败:', error);
    }
  };

  // 删除智慧
  const handleDelete = async (id: string) => {
    try {
      const response = await apiClient.deleteWisdom(id);
      if (response.success) {
        message.success('删除成功');
        loadWisdomList();
      } else {
        message.error('删除智慧失败');
      }
    } catch (error) {
      console.error('删除失败:', error);
      message.error('删除失败');
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的项目');
      return;
    }

    try {
      await Promise.all(
        selectedRowKeys.map(id => apiClient.deleteWisdom(id as string))
      );
      message.success('批量删除成功');
      setSelectedRowKeys([]);
      loadWisdomList();
    } catch {
      message.error('批量删除失败');
    }
  };

  // 搜索
  const handleSearch = (value: string) => {
    const newFilters = { ...filters, keyword: value };
    setFilters(newFilters);
    setCurrentPage(1);
    loadWisdomList(1, pageSize, newFilters);
  };

  // 应用筛选
  const handleApplyFilters = (values: WisdomFilters) => {
    const newFilters: WisdomFilters = {
      keyword: values.keyword,
      category: values.category,
      school: values.school,
      author: values.author,
      status: values.status
    };

    if (values.dateRange) {
      newFilters.dateRange = [
        values.dateRange[0].format('YYYY-MM-DD'),
        values.dateRange[1].format('YYYY-MM-DD')
      ];
    }

    setFilters(newFilters);
    setCurrentPage(1);
    setFilterDrawerVisible(false);
    loadWisdomList(1, pageSize, newFilters);
  };

  // 重置筛选
  const handleResetFilters = () => {
    form.resetFields();
    setFilters({});
    setCurrentPage(1);
    setFilterDrawerVisible(false);
    loadWisdomList(1, pageSize, {});
  };

  // 导出数据
  const handleExport = async () => {
    try {
      // 这里应该调用导出API
      message.success('导出功能开发中...');
    } catch {
      message.error('导出失败');
    }
  };

  useEffect(() => {
    loadWisdomList();
    loadFilterOptions();
  }, [loadWisdomList]);

  // 表格列定义
  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
      render: (text: string, record: CulturalWisdom) => (
        <a onClick={() => navigate(`/wisdom/${record.id}`)}>
          {text}
        </a>
      )
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (category: string) => (
        <Tag color="blue">{category}</Tag>
      )
    },
    {
      title: '学派',
      dataIndex: 'school',
      key: 'school',
      width: 100,
      render: (school: string) => school && (
        <Tag color="green">{school}</Tag>
      )
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author',
      width: 120,
      ellipsis: true
    },
    {
      title: '朝代',
      dataIndex: 'dynasty',
      key: 'dynasty',
      width: 100,
      render: (dynasty: string) => dynasty && (
        <Tag color="orange">{dynasty}</Tag>
      )
    },
    {
      title: '统计',
      key: 'stats',
      width: 120,
      render: (record: CulturalWisdom) => (
        <Space direction="vertical" size="small">
          <span>
            <EyeOutlined /> {record.views || 0}
          </span>
          <span>
            <HeartOutlined /> {record.likes || 0}
          </span>
        </Space>
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
      fixed: 'right' as const,
      render: (record: CulturalWisdom) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/wisdom/${record.id}`)}
          >
            查看
          </Button>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => navigate(`/admin/wisdom/${record.id}/edit`)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这条智慧吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
    getCheckboxProps: (record: CulturalWisdom) => ({
      disabled: false,
      name: record.title,
    }),
  };

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <Title level={2} className="mb-2">
              智慧内容管理
            </Title>
            <p className="text-gray-600 mb-0">
              管理所有文化智慧内容，包括增加、编辑、删除和查看
            </p>
          </div>
          <Space>
            <Button 
              icon={<ImportOutlined />}
              onClick={() => message.info('导入功能开发中...')}
            >
              导入
            </Button>
            <Button 
              icon={<ExportOutlined />}
              onClick={handleExport}
            >
              导出
            </Button>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={() => navigate('/admin/wisdom/create')}
            >
              添加智慧
            </Button>
          </Space>
        </div>
      </Card>

      {/* 搜索和筛选 */}
      <Card>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} md={12}>
            <Search
              placeholder="搜索标题、内容、作者..."
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              className="w-full"
            />
          </Col>
          <Col xs={24} md={12}>
            <Space className="w-full justify-end">
              <Button 
                icon={<FilterOutlined />}
                onClick={() => setFilterDrawerVisible(true)}
              >
                高级筛选
              </Button>
              {selectedRowKeys.length > 0 && (
                <Popconfirm
                  title={`确定要删除选中的 ${selectedRowKeys.length} 项吗？`}
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
          </Col>
        </Row>
      </Card>

      {/* 数据表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={wisdomList}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          scroll={{ x: 1200 }}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size || 10);
              loadWisdomList(page, size || 10);
            }
          }}
        />
      </Card>

      {/* 高级筛选抽屉 */}
      <Drawer
        title="高级筛选"
        placement="right"
        width={400}
        onClose={() => setFilterDrawerVisible(false)}
        open={filterDrawerVisible}
        extra={
          <Space>
            <Button onClick={handleResetFilters}>
              重置
            </Button>
            <Button type="primary" onClick={() => form.submit()}>
              应用
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleApplyFilters}
          initialValues={filters}
        >
          <Form.Item name="keyword" label="关键词">
            <Input placeholder="搜索标题、内容、作者..." />
          </Form.Item>

          <Form.Item name="category" label="分类">
            <Select placeholder="选择分类" allowClear>
              {categories.map(cat => (
                <Option key={cat.id} value={cat.name}>
                  {cat.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="school" label="学派">
            <Select placeholder="选择学派" allowClear>
              {schools.map(school => (
                <Option key={school.id} value={school.name}>
                  {school.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="author" label="作者">
            <Input placeholder="输入作者名称" />
          </Form.Item>

          <Form.Item name="dateRange" label="创建时间">
            <RangePicker className="w-full" />
          </Form.Item>

          <Form.Item name="status" label="状态">
            <Select placeholder="选择状态" allowClear>
              <Option value="published">已发布</Option>
              <Option value="draft">草稿</Option>
              <Option value="archived">已归档</Option>
            </Select>
          </Form.Item>
        </Form>
      </Drawer>
    </div>
  );
};

export default WisdomManagement;