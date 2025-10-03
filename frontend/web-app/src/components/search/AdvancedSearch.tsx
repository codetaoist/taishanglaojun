import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Select,
  DatePicker,
  Button,
  Space,
  Row,
  Col,
  Collapse,
  Tag,
  Slider,
  Switch,
  Typography,
  Divider,
  message
} from 'antd';
import {
  SearchOutlined,
  FilterOutlined,
  ClearOutlined,
  DownOutlined,
  UpOutlined
} from '@ant-design/icons';
import { apiClient } from '../../services/api';
import behaviorService from '../../services/behaviorService';
import SearchSuggestions from './SearchSuggestions';
import type { Category, Tag as WisdomTag } from '../../types';

const { Option } = Select;
const { RangePicker } = DatePicker;
const { Panel } = Collapse;
const { Text } = Typography;

export interface AdvancedSearchFilters {
  keyword?: string;
  category?: string;
  school?: string;
  author?: string;
  tags?: string[];
  difficulty?: number[];
  dateFrom?: string;
  dateTo?: string;
  status?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface AdvancedSearchProps {
  onSearch: (filters: AdvancedSearchFilters) => void;
  loading?: boolean;
  initialFilters?: AdvancedSearchFilters;
}

const AdvancedSearch: React.FC<AdvancedSearchProps> = ({
  onSearch,
  loading = false,
  initialFilters = {}
}) => {
  const [form] = Form.useForm();
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<WisdomTag[]>([]);
  const [expanded, setExpanded] = useState(false);
  const [loadingData, setLoadingData] = useState(false);

  // 学派选项
  const schoolOptions = [
    { label: '儒家', value: '儒家' },
    { label: '道家', value: '道家' },
    { label: '佛家', value: '佛家' },
    { label: '法家', value: '法家' },
    { label: '墨家', value: '墨家' },
    { label: '兵家', value: '兵家' },
    { label: '纵横家', value: '纵横家' },
    { label: '阴阳家', value: '阴阳家' }
  ];

  // 难度等级选项
  const difficultyMarks = {
    1: '入门',
    2: '初级',
    3: '中级',
    4: '高级',
    5: '专家'
  };

  // 排序选项
  const sortOptions = [
    { label: '创建时间', value: 'created_at' },
    { label: '更新时间', value: 'updated_at' },
    { label: '浏览量', value: 'views' },
    { label: '点赞数', value: 'likes' },
    { label: '收藏数', value: 'favorites' },
    { label: '标题', value: 'title' },
    { label: '作者', value: 'author' }
  ];

  // 状态选项
  const statusOptions = [
    { label: '已发布', value: 'published' },
    { label: '草稿', value: 'draft' },
    { label: '审核中', value: 'pending' },
    { label: '已下架', value: 'archived' }
  ];

  // 加载分类和标签数据
  const loadMetadata = async () => {
    setLoadingData(true);
    try {
      const [categoriesRes, tagsRes] = await Promise.all([
        apiClient.getCategories(),
        apiClient.getTags()
      ]);

      if (categoriesRes.success && categoriesRes.data) {
        setCategories(categoriesRes.data);
      }

      if (tagsRes.success && tagsRes.data) {
        setTags(tagsRes.data);
      }
    } catch (error) {
      console.error('加载元数据失败:', error);
      message.error('加载分类和标签数据失败');
    }
    setLoadingData(false);
  };

  // 处理搜索
  const handleSearch = async () => {
    const values = form.getFieldsValue();
    const filters: AdvancedSearchFilters = {};

    // 处理关键词
    if (values.keyword?.trim()) {
      filters.keyword = values.keyword.trim();
    }

    // 处理分类
    if (values.category) {
      filters.category = values.category;
    }

    // 处理学派
    if (values.school) {
      filters.school = values.school;
    }

    // 处理作者
    if (values.author?.trim()) {
      filters.author = values.author.trim();
    }

    // 处理标签
    if (values.tags && values.tags.length > 0) {
      filters.tags = values.tags;
    }

    // 处理难度
    if (values.difficulty && values.difficulty.length === 2) {
      const [min, max] = values.difficulty;
      filters.difficulty = [];
      for (let i = min; i <= max; i++) {
        filters.difficulty.push(i);
      }
    }

    // 处理日期范围
    if (values.dateRange && values.dateRange.length === 2) {
      filters.dateFrom = values.dateRange[0].format('YYYY-MM-DD');
      filters.dateTo = values.dateRange[1].format('YYYY-MM-DD');
    }

    // 处理状态
    if (values.status) {
      filters.status = values.status;
    }

    // 处理排序
    if (values.sortBy) {
      filters.sortBy = values.sortBy;
      filters.sortOrder = values.sortOrder || 'desc';
    }

    // 记录搜索行为
    try {
      const searchQuery = filters.keyword || '高级搜索';
      await behaviorService.recordSearch(searchQuery, 0); // 结果数量将在搜索完成后更新
    } catch (error) {
      console.warn('记录搜索行为失败:', error);
    }

    onSearch(filters);
  };

  // 重置表单
  const handleReset = () => {
    form.resetFields();
    onSearch({});
  };

  // 组件挂载时加载数据
  useEffect(() => {
    loadMetadata();
  }, []);

  // 设置初始值
  useEffect(() => {
    if (initialFilters && Object.keys(initialFilters).length > 0) {
      const formValues: any = { ...initialFilters };
      
      // 处理日期范围
      if (initialFilters.dateFrom && initialFilters.dateTo) {
        formValues.dateRange = [
          initialFilters.dateFrom,
          initialFilters.dateTo
        ];
      }

      // 处理难度范围
      if (initialFilters.difficulty && initialFilters.difficulty.length > 0) {
        const min = Math.min(...initialFilters.difficulty);
        const max = Math.max(...initialFilters.difficulty);
        formValues.difficulty = [min, max];
      }

      form.setFieldsValue(formValues);
    }
  }, [initialFilters, form]);

  return (
    <Card className="advanced-search-card">
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSearch}
        className="advanced-search-form"
      >
        {/* 基础搜索 */}
        <Row gutter={[16, 16]}>
          <Col xs={24} md={16}>
            <Form.Item name="keyword" label="关键词搜索">
              <SearchSuggestions
                placeholder="搜索标题、内容、作者..."
                size="large"
                onSearch={(value) => {
                  form.setFieldsValue({ keyword: value });
                  handleSearch();
                }}
                onSelect={(value) => {
                  form.setFieldsValue({ keyword: value });
                }}
              />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item label=" " className="search-buttons">
              <Space>
                <Button
                  type="primary"
                  icon={<SearchOutlined />}
                  onClick={handleSearch}
                  loading={loading}
                  size="large"
                >
                  搜索
                </Button>
                <Button
                  icon={<ClearOutlined />}
                  onClick={handleReset}
                  size="large"
                >
                  重置
                </Button>
                <Button
                  type="text"
                  icon={expanded ? <UpOutlined /> : <DownOutlined />}
                  onClick={() => setExpanded(!expanded)}
                  size="large"
                >
                  {expanded ? '收起' : '展开'}筛选
                </Button>
              </Space>
            </Form.Item>
          </Col>
        </Row>

        {/* 高级筛选 */}
        <Collapse
          activeKey={expanded ? ['filters'] : []}
          ghost
          className="advanced-filters"
        >
          <Panel key="filters" header="" showArrow={false}>
            <Divider orientation="left">
              <Text strong>
                <FilterOutlined /> 高级筛选
              </Text>
            </Divider>

            <Row gutter={[16, 16]}>
              {/* 分类筛选 */}
              <Col xs={24} sm={12} md={8}>
                <Form.Item name="category" label="分类">
                  <Select
                    placeholder="选择分类"
                    allowClear
                    loading={loadingData}
                    showSearch
                    filterOption={(input, option) =>
                      (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
                    }
                  >
                    {categories.map(category => (
                      <Option key={category.id} value={category.id}>
                        {category.name}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              {/* 学派筛选 */}
              <Col xs={24} sm={12} md={8}>
                <Form.Item name="school" label="学派">
                  <Select placeholder="选择学派" allowClear>
                    {schoolOptions.map(school => (
                      <Option key={school.value} value={school.value}>
                        {school.label}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              {/* 作者筛选 */}
              <Col xs={24} sm={12} md={8}>
                <Form.Item name="author" label="作者">
                  <Input placeholder="输入作者名称" allowClear />
                </Form.Item>
              </Col>

              {/* 标签筛选 */}
              <Col xs={24} md={12}>
                <Form.Item name="tags" label="标签">
                  <Select
                    mode="multiple"
                    placeholder="选择标签"
                    allowClear
                    loading={loadingData}
                    showSearch
                    filterOption={(input, option) =>
                      (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
                    }
                    maxTagCount="responsive"
                  >
                    {tags.map(tag => (
                      <Option key={tag.id} value={tag.name}>
                        {tag.name}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              {/* 状态筛选 */}
              <Col xs={24} sm={12} md={6}>
                <Form.Item name="status" label="状态">
                  <Select placeholder="选择状态" allowClear>
                    {statusOptions.map(status => (
                      <Option key={status.value} value={status.value}>
                        {status.label}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              {/* 日期范围 */}
              <Col xs={24} md={12}>
                <Form.Item name="dateRange" label="创建日期">
                  <RangePicker
                    placeholder={['开始日期', '结束日期']}
                    className="w-full"
                  />
                </Form.Item>
              </Col>

              {/* 难度等级 */}
              <Col xs={24}>
                <Form.Item name="difficulty" label="难度等级">
                  <Slider
                    range
                    min={1}
                    max={5}
                    marks={difficultyMarks}
                    step={1}
                    tipFormatter={(value) => difficultyMarks[value as keyof typeof difficultyMarks]}
                  />
                </Form.Item>
              </Col>

              {/* 排序选项 */}
              <Col xs={24} sm={12} md={8}>
                <Form.Item name="sortBy" label="排序方式">
                  <Select placeholder="选择排序字段" allowClear>
                    {sortOptions.map(sort => (
                      <Option key={sort.value} value={sort.value}>
                        {sort.label}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              <Col xs={24} sm={12} md={8}>
                <Form.Item name="sortOrder" label="排序顺序" initialValue="desc">
                  <Select>
                    <Option value="desc">降序</Option>
                    <Option value="asc">升序</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
          </Panel>
        </Collapse>
      </Form>
    </Card>
  );
};

export default AdvancedSearch;