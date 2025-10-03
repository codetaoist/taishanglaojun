import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Button, 
  Tag, 
  Space, 
  Typography, 
  Pagination,
  Spin,
  message,
  Row,
  Col,
  Divider
} from 'antd';
import { 
  SearchOutlined, 
  BookOutlined, 
  EyeOutlined, 
  HeartOutlined,
  StarOutlined,
  FilterOutlined,
  AppstoreOutlined,
  BarsOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../services/api';
import behaviorService from '../services/behaviorService';
import AdvancedSearch from '../components/search/AdvancedSearch';
import type { AdvancedSearchFilters } from '../components/search/AdvancedSearch';
import type { CulturalWisdom } from '../types';

const { Title, Paragraph } = Typography;

const Wisdom: React.FC = () => {
  const navigate = useNavigate();
  const [wisdomList, setWisdomList] = useState<CulturalWisdom[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [currentFilters, setCurrentFilters] = useState<AdvancedSearchFilters>({});
  const [viewMode, setViewMode] = useState<'card' | 'list'>('card');

  // 加载智慧列表
  const loadWisdomList = async (page = 1, filters: AdvancedSearchFilters = {}) => {
    setLoading(true);
    try {
      // 使用高级搜索API
      const response = await apiClient.advancedSearchWisdom({
        ...filters,
        page,
        size: pageSize
      });
      
      if (response.success && response.data) {
        setWisdomList(response.data.items || []);
        setTotal(response.data.total || 0);
      } else {
        message.error('获取智慧列表失败');
      }
    } catch (error) {
      console.error('加载智慧列表失败:', error);
      message.error('网络错误，请稍后重试');
    }
    setLoading(false);
  };

  // 处理高级搜索
  const handleAdvancedSearch = async (filters: AdvancedSearchFilters) => {
    setCurrentFilters(filters);
    setCurrentPage(1);
    await loadWisdomList(1, filters);
    
    // 记录搜索行为（更新结果数量）
    try {
      const searchQuery = filters.keyword || '高级搜索';
      await behaviorService.recordSearch(searchQuery, total);
    } catch (error) {
      console.warn('更新搜索行为记录失败:', error);
    }
  };

  // 页面变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    loadWisdomList(page, currentFilters);
  };

  // 查看智慧详情
  const handleViewDetail = async (wisdom: CulturalWisdom) => {
    // 记录点击行为
    try {
      await behaviorService.recordClick('wisdom', wisdom.id.toString());
    } catch (error) {
      console.warn('记录点击行为失败:', error);
    }
    
    navigate(`/wisdom/${wisdom.id}`);
  };

  // 组件挂载时加载数据
  useEffect(() => {
    loadWisdomList();
  }, []);

  // 渲染智慧卡片
  const renderWisdomCard = (item: CulturalWisdom) => (
    <List.Item key={item.id}>
      <Card 
        hoverable
        className="w-full shadow-md hover:shadow-lg transition-all duration-300"
        onClick={() => handleViewDetail(item)}
      >
        <div className="space-y-4">
          {/* 标题和分类 */}
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <Title level={4} className="mb-2 line-clamp-2">
                {item.title}
              </Title>
              <div className="flex items-center space-x-2 mb-3">
                <Tag color="gold">{item.category}</Tag>
                {item.source && <Tag color="blue">{item.source}</Tag>}
                {item.dynasty && <Tag color="green">{item.dynasty}</Tag>}
              </div>
            </div>
          </div>

          {/* 内容预览 */}
          <Paragraph 
            className="text-gray-600 line-clamp-3"
            ellipsis={{ rows: 3, expandable: false }}
          >
            {item.content}
          </Paragraph>

          {/* 标签 */}
          {item.tags && item.tags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {item.tags.slice(0, 5).map((tag, index) => (
                <Tag key={index} color="processing" className="text-xs">
                  {tag}
                </Tag>
              ))}
              {item.tags.length > 5 && (
                <Tag color="default" className="text-xs">
                  +{item.tags.length - 5}
                </Tag>
              )}
            </div>
          )}

          <Divider className="my-3" />

          {/* 底部信息 */}
          <div className="flex items-center justify-between text-sm text-gray-500">
            <Space>
              <EyeOutlined />
              <span>{item.views || 0} 阅读</span>
              <HeartOutlined />
              <span>{item.likes || 0} 点赞</span>
            </Space>
            <Space>
              <StarOutlined />
              <span>收藏</span>
            </Space>
          </div>
        </div>
      </Card>
    </List.Item>
  );

  // 渲染智慧列表项
  const renderWisdomListItem = (item: CulturalWisdom) => (
    <List.Item
      key={item.id}
      className="hover:bg-gray-50 cursor-pointer p-4 border-b"
      onClick={() => handleViewDetail(item)}
      actions={[
        <Space key="stats">
          <EyeOutlined />
          <span>{item.views || 0}</span>
          <HeartOutlined />
          <span>{item.likes || 0}</span>
        </Space>,
        <Button key="favorite" type="text" icon={<StarOutlined />}>
          收藏
        </Button>
      ]}
    >
      <List.Item.Meta
        title={
          <div className="flex items-center space-x-2">
            <span className="text-lg font-medium">{item.title}</span>
            <Tag color="gold">{item.category}</Tag>
            {item.source && <Tag color="blue">{item.source}</Tag>}
          </div>
        }
        description={
          <div className="space-y-2">
            <Paragraph 
              className="text-gray-600"
              ellipsis={{ rows: 2, expandable: false }}
            >
              {item.content}
            </Paragraph>
            {item.tags && item.tags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {item.tags.slice(0, 3).map((tag, index) => (
                  <Tag key={index} color="processing" size="small">
                    {tag}
                  </Tag>
                ))}
                {item.tags.length > 3 && (
                  <Tag color="default" size="small">
                    +{item.tags.length - 3}
                  </Tag>
                )}
              </div>
            )}
          </div>
        }
      />
    </List.Item>
  );

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card className="bg-gradient-to-r from-cultural-gold/10 to-cultural-red/10 border-cultural-gold/20">
        <div className="text-center space-y-4">
          <Title level={2} className="flex items-center justify-center space-x-2 mb-4">
            <BookOutlined className="text-cultural-gold" />
            <span>文化智慧库</span>
          </Title>
          <Paragraph className="text-lg text-gray-600 max-w-2xl mx-auto">
            汇聚千年文化精髓，传承古圣先贤智慧。在这里探索中华文化的博大精深，感悟人生哲理。
          </Paragraph>
        </div>
      </Card>

      {/* 高级搜索 */}
      <AdvancedSearch
        onSearch={handleAdvancedSearch}
        loading={loading}
        initialFilters={currentFilters}
      />

      {/* 智慧列表 */}
      <Card 
        title={
          <div className="flex items-center justify-between">
            <span>智慧列表 ({total} 条)</span>
            <Space>
              {/* 视图切换 */}
              <Button.Group>
                <Button
                  type={viewMode === 'card' ? 'primary' : 'default'}
                  icon={<AppstoreOutlined />}
                  onClick={() => setViewMode('card')}
                >
                  卡片
                </Button>
                <Button
                  type={viewMode === 'list' ? 'primary' : 'default'}
                  icon={<BarsOutlined />}
                  onClick={() => setViewMode('list')}
                >
                  列表
                </Button>
              </Button.Group>
              
              {/* 刷新按钮 */}
              <Button 
                type="primary" 
                onClick={() => loadWisdomList(1, currentFilters)}
                loading={loading}
              >
                刷新
              </Button>
            </Space>
          </div>
        }
      >
        <Spin spinning={loading}>
          {wisdomList.length > 0 ? (
            <>
              <List
                grid={viewMode === 'card' ? { 
                  gutter: 16, 
                  xs: 1, 
                  sm: 1, 
                  md: 1, 
                  lg: 1, 
                  xl: 1, 
                  xxl: 1 
                } : false}
                dataSource={wisdomList}
                renderItem={viewMode === 'card' ? renderWisdomCard : renderWisdomListItem}
              />
              
              {/* 分页 */}
              <div className="flex justify-center mt-8">
                <Pagination
                  current={currentPage}
                  total={total}
                  pageSize={pageSize}
                  onChange={handlePageChange}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(total, range) => 
                    `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                  }
                />
              </div>
            </>
          ) : (
            <div className="text-center py-12">
              <BookOutlined className="text-6xl text-gray-300 mb-4" />
              <Title level={4} type="secondary">
                {Object.keys(currentFilters).length > 0 ? '未找到符合条件的智慧内容' : '暂无智慧内容'}
              </Title>
              <Paragraph type="secondary">
                {Object.keys(currentFilters).length > 0 ? '请尝试调整搜索条件' : '智慧内容正在加载中...'}
              </Paragraph>
            </div>
          )}
        </Spin>
      </Card>
    </div>
  );
};

export default Wisdom;