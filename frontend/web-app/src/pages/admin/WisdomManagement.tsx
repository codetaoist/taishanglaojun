import React, { useState, useEffect, useCallback, useMemo } from 'react';
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
import { useDebounce, useVirtualScroll, useRenderCount } from '../../utils/performanceOptimization';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
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

  // 性能监控
  useRenderCount('WisdomManagement');

  // 防抖搜索
  const debouncedSearch = useDebounce((searchFilters: WisdomFilters) => {
    loadWisdomList(1, pageSize, searchFilters);
  }, 300);

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
        message.error(t('adminWisdom.messages.loadFailed'));
      }
    } catch {
      message.error(t('adminWisdom.messages.loadListFailed'));
    }
    setLoading(false);
  }, [currentPage, pageSize, filters]);

  // 加载分类和学派数据
  const loadFilterOptions = useCallback(async () => {
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
    } catch {
      message.error(t('adminWisdom.messages.loadFiltersFailed'));
    }
  }, []);

  // 缓存表格列配置
  const columns = useMemo(() => [
    {
      title: t('adminWisdom.table.title'),
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
      render: (text: string, record: CulturalWisdom) => (
        <Button 
          type="link" 
          onClick={() => handleView(record.id)}
          style={{ padding: 0, height: 'auto' }}
        >
          {text}
        </Button>
      ),
    },
    {
      title: t('adminWisdom.table.category'),
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (category: string) => (
        <Tag color="blue">{category}</Tag>
      ),
    },
    {
      title: t('adminWisdom.table.school'),
      dataIndex: 'school',
      key: 'school',
      width: 120,
      render: (school: string) => (
        <Tag color="green">{school}</Tag>
      ),
    },
    {
      title: t('adminWisdom.table.author'),
      dataIndex: 'author',
      key: 'author',
      width: 120,
      ellipsis: true,
    },
    {
      title: t('adminWisdom.table.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusConfig = {
          published: { color: 'success', text: t('adminWisdom.table.statusMap.published') },
          draft: { color: 'warning', text: t('adminWisdom.table.statusMap.draft') },
          archived: { color: 'default', text: t('adminWisdom.table.statusMap.archived') },
        };
        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.draft;
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: t('adminWisdom.table.createdAt'),
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: t('adminWisdom.table.actions'),
      key: 'action',
      width: 200,
      fixed: 'right' as const,
      render: (_: any, record: CulturalWisdom) => (
        <Space size="small">
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleView(record.id)}
            title={t('adminWisdom.table.tooltips.view')}
          />
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record.id)}
            title={t('adminWisdom.table.tooltips.edit')}
          />
          <Popconfirm
            title={t('adminWisdom.table.confirm.deleteTitle')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('adminWisdom.table.confirm.ok')}
            cancelText={t('adminWisdom.table.confirm.cancel')}
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title={t('adminWisdom.table.tooltips.delete')}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ], [t]);

  // 缓存行选择配置
  const rowSelection = useMemo(() => ({
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys);
    },
    getCheckboxProps: (record: CulturalWisdom) => ({
      disabled: record.status === 'archived',
      name: record.title,
    }),
  }), [selectedRowKeys]);

  // 缓存分页配置
  const paginationConfig = useMemo(() => ({
    current: currentPage,
    pageSize,
    total,
    showSizeChanger: true,
    showQuickJumper: true,
    showTotal: (total: number, range: [number, number]) =>
      t('adminWisdom.table.pagination.total', { start: range[0], end: range[1], total }),
    onChange: (page: number, size: number) => {
      setCurrentPage(page);
      setPageSize(size);
      loadWisdomList(page, size);
    },
  }), [currentPage, pageSize, total, loadWisdomList, t]);

  // 事件处理函数优化
  const handleView = useCallback((id: string) => {
    navigate(`/admin/wisdom/${id}`);
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(`/admin/wisdom/edit/${id}`);
  }, [navigate]);

  const handleDelete = useCallback(async (id: string) => {
    try {
      const response = await apiClient.deleteWisdom(id);
      if (response.success) {
        message.success(t('adminWisdom.messages.deleteSuccess'));
        loadWisdomList();
      } else {
        message.error(t('adminWisdom.messages.deleteFailed'));
      }
    } catch {
      message.error(t('adminWisdom.messages.deleteFailed'));
    }
  }, [loadWisdomList, t]);

  const handleBatchDelete = useCallback(async () => {
    if (selectedRowKeys.length === 0) {
      message.warning(t('adminWisdom.messages.selectItemsToDelete'));
      return;
    }

    try {
      const response = await apiClient.batchDeleteWisdom(selectedRowKeys as string[]);
      if (response.success) {
        message.success(t('adminWisdom.messages.batchDeleteSuccess'));
        setSelectedRowKeys([]);
        loadWisdomList();
      } else {
        message.error(t('adminWisdom.messages.batchDeleteFailed'));
      }
    } catch {
      message.error(t('adminWisdom.messages.batchDeleteFailed'));
    }
  }, [selectedRowKeys, loadWisdomList, t]);

  const handleSearch = useCallback((value: string) => {
    const newFilters = { ...filters, keyword: value };
    setFilters(newFilters);
    debouncedSearch(newFilters);
  }, [filters, debouncedSearch]);

  const handleFilterChange = useCallback((key: keyof WisdomFilters, value: any) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    debouncedSearch(newFilters);
  }, [filters, debouncedSearch]);

  const handleResetFilters = useCallback(() => {
    const resetFilters = {};
    setFilters(resetFilters);
    form.resetFields();
    loadWisdomList(1, pageSize, resetFilters);
  }, [form, pageSize, loadWisdomList]);

  const handleExport = useCallback(async () => {
    try {
      const response = await apiClient.exportWisdom(filters);
      if (response.success) {
        message.success(t('adminWisdom.messages.exportSuccess'));
        // 处理文件下载
      } else {
        message.error(t('adminWisdom.messages.exportFailed'));
      }
    } catch {
      message.error(t('adminWisdom.messages.exportFailed'));
    }
  }, [filters, t]);

  useEffect(() => {
    loadWisdomList();
    loadFilterOptions();
  }, [loadWisdomList]);




  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <Title level={2} className="mb-2">
              {t('adminWisdom.title')}
            </Title>
            <p className="text-gray-600 mb-0">
              {t('adminWisdom.subtitle')}
            </p>
          </div>
          <Space>
            <Button 
              icon={<ImportOutlined />}
              onClick={() => message.info(t('adminWisdom.messages.importWip'))}
            >
              {t('adminWisdom.actions.import')}
            </Button>
            <Button 
              icon={<ExportOutlined />}
              onClick={handleExport}
            >
              {t('adminWisdom.actions.export')}
            </Button>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={() => navigate('/admin/wisdom/create')}
            >
              {t('adminWisdom.actions.addWisdom')}
            </Button>
          </Space>
        </div>
      </Card>

      {/* 搜索和筛选 */}
      <Card>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} md={12}>
            <Search
              placeholder={t('adminWisdom.search.placeholder')}
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
                {t('adminWisdom.actions.advancedFilter')}
              </Button>
              {selectedRowKeys.length > 0 && (
                <Popconfirm
                  title={t('adminWisdom.confirm.batchDeleteTitle', { count: selectedRowKeys.length })}
                  onConfirm={handleBatchDelete}
                  okText={t('adminWisdom.confirm.ok')}
                  cancelText={t('adminWisdom.confirm.cancel')}
                >
                  <Button danger>
                    {t('adminWisdom.actions.batchDelete')} ({selectedRowKeys.length})
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
        title={t('adminWisdom.drawer.title')}
        placement="right"
        width={400}
        onClose={() => setFilterDrawerVisible(false)}
        open={filterDrawerVisible}
        extra={
          <Space>
            <Button onClick={handleResetFilters}>
              {t('adminWisdom.actions.reset')}
            </Button>
            <Button type="primary" onClick={() => form.submit()}>
              {t('adminWisdom.actions.apply')}
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
          <Form.Item name="keyword" label={t('adminWisdom.drawer.fields.keyword')}>
            <Input placeholder={t('adminWisdom.drawer.placeholders.keyword')} />
          </Form.Item>

          <Form.Item name="category" label={t('adminWisdom.drawer.fields.category')}>
            <Select placeholder={t('adminWisdom.drawer.placeholders.category')} allowClear>
              {categories.map(cat => (
                <Option key={cat.id} value={cat.name}>
                  {cat.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="school" label={t('adminWisdom.drawer.fields.school')}>
            <Select placeholder={t('adminWisdom.drawer.placeholders.school')} allowClear>
              {schools.map(school => (
                <Option key={school.id} value={school.name}>
                  {school.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="author" label={t('adminWisdom.drawer.fields.author')}>
            <Input placeholder={t('adminWisdom.drawer.placeholders.author')} />
          </Form.Item>

          <Form.Item name="dateRange" label={t('adminWisdom.drawer.fields.dateRange')}>
            <RangePicker className="w-full" />
          </Form.Item>

          <Form.Item name="status" label={t('adminWisdom.drawer.fields.status')}>
            <Select placeholder={t('adminWisdom.drawer.placeholders.status')} allowClear>
              <Option value="published">{t('adminWisdom.drawer.statusOptions.published')}</Option>
              <Option value="draft">{t('adminWisdom.drawer.statusOptions.draft')}</Option>
              <Option value="archived">{t('adminWisdom.drawer.statusOptions.archived')}</Option>
            </Select>
          </Form.Item>
        </Form>
      </Drawer>
    </div>
  );
};

export default WisdomManagement;