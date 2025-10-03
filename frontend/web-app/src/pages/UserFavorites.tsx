import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Typography, 
  Tag, 
  Space, 
  Button, 
  message,
  Spin,
  Empty,
  Pagination,
  Input,
  Row,
  Col
} from 'antd';
import { 
  StarOutlined,
  EyeOutlined,
  HeartOutlined,
  BookOutlined,
  UserOutlined,
  CalendarOutlined,
  SearchOutlined,
  StarFilled
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../services/api';

const { Title, Paragraph } = Typography;
const { Search } = Input;

interface FavoriteItem {
  wisdom_id: string;
  title: string;
  author: string;
  category: string;
  school: string;
  summary: string;
  created_at: string;
  favorited_at: string;
}

const UserFavorites: React.FC = () => {
  const navigate = useNavigate();
  const [favorites, setFavorites] = useState<FavoriteItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const [searchKeyword, setSearchKeyword] = useState('');

  // 加载收藏列表
  const loadFavorites = async (page = 1, search = '') => {
    setLoading(true);
    try {
      const response = await apiClient.getUserFavorites({
        page,
        limit: pageSize,
      });

      if (response.success && response.data) {
        let filteredFavorites = response.data.favorites;
        
        // 前端搜索过滤（如果后端不支持搜索）
        if (search) {
          filteredFavorites = filteredFavorites.filter(item =>
            item.title.toLowerCase().includes(search.toLowerCase()) ||
            item.author.toLowerCase().includes(search.toLowerCase()) ||
            item.category.toLowerCase().includes(search.toLowerCase())
          );
        }

        setFavorites(filteredFavorites);
        setTotal(response.data.total);
      } else {
        message.error('获取收藏列表失败');
      }
    } catch (error) {
      console.error('加载收藏列表失败:', error);
      message.error('网络错误，请稍后重试');
    }
    setLoading(false);
  };

  // 取消收藏
  const handleRemoveFavorite = async (wisdomId: string) => {
    try {
      await apiClient.removeFavorite(wisdomId);
      message.success('取消收藏成功');
      // 重新加载列表
      loadFavorites(currentPage, searchKeyword);
    } catch (error) {
      message.error('取消收藏失败');
    }
  };

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    setCurrentPage(1);
    loadFavorites(1, value);
  };

  // 页码变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    loadFavorites(page, searchKeyword);
  };

  useEffect(() => {
    loadFavorites();
  }, []);

  return (
    <div className="max-w-6xl mx-auto p-6">
      {/* 页面标题 */}
      <div className="mb-6">
        <Title level={2} className="mb-2">
          <StarFilled className="text-yellow-500 mr-2" />
          我的收藏
        </Title>
        <Paragraph className="text-gray-600">
          管理您收藏的智慧内容，随时回顾经典名句
        </Paragraph>
      </div>

      {/* 搜索栏 */}
      <Card className="mb-6">
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={18} md={20}>
            <Search
              placeholder="搜索收藏的智慧..."
              allowClear
              enterButton={<SearchOutlined />}
              size="large"
              onSearch={handleSearch}
              onChange={(e) => {
                if (!e.target.value) {
                  handleSearch('');
                }
              }}
            />
          </Col>
          <Col xs={24} sm={6} md={4}>
            <div className="text-right">
              <span className="text-gray-500">
                共 {total} 条收藏
              </span>
            </div>
          </Col>
        </Row>
      </Card>

      {/* 收藏列表 */}
      <Card>
        <Spin spinning={loading}>
          {favorites.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                searchKeyword ? '没有找到匹配的收藏' : '您还没有收藏任何智慧'
              }
            >
              {!searchKeyword && (
                <Button type="primary" onClick={() => navigate('/wisdom')}>
                  去发现智慧
                </Button>
              )}
            </Empty>
          ) : (
            <>
              <List
                itemLayout="vertical"
                dataSource={favorites}
                renderItem={(item) => (
                  <List.Item
                    key={item.wisdom_id}
                    className="hover:bg-gray-50 transition-colors duration-200 rounded-lg p-4"
                    actions={[
                      <Button
                        key="view"
                        type="link"
                        onClick={() => navigate(`/wisdom/${item.wisdom_id}`)}
                      >
                        查看详情
                      </Button>,
                      <Button
                        key="unfavorite"
                        type="link"
                        danger
                        icon={<StarOutlined />}
                        onClick={() => handleRemoveFavorite(item.wisdom_id)}
                      >
                        取消收藏
                      </Button>,
                    ]}
                  >
                    <List.Item.Meta
                      title={
                        <div className="flex items-center justify-between">
                          <span
                            className="text-lg font-semibold text-blue-600 hover:text-blue-800 cursor-pointer"
                            onClick={() => navigate(`/wisdom/${item.wisdom_id}`)}
                          >
                            {item.title}
                          </span>
                          <div className="flex items-center space-x-2">
                            <Tag color="gold" className="text-xs">
                              <BookOutlined className="mr-1" />
                              {item.category}
                            </Tag>
                            {item.school && (
                              <Tag color="blue" className="text-xs">
                                {item.school}
                              </Tag>
                            )}
                          </div>
                        </div>
                      }
                      description={
                        <div className="space-y-2">
                          <Paragraph
                            className="text-gray-700 mb-2"
                            ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}
                          >
                            {item.summary}
                          </Paragraph>
                          <div className="flex items-center justify-between text-sm text-gray-500">
                            <Space>
                              {item.author && (
                                <span>
                                  <UserOutlined className="mr-1" />
                                  {item.author}
                                </span>
                              )}
                              <span>
                                <CalendarOutlined className="mr-1" />
                                收藏于 {new Date(item.favorited_at).toLocaleDateString()}
                              </span>
                            </Space>
                          </div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />

              {/* 分页 */}
              {total > pageSize && (
                <div className="mt-6 text-center">
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
              )}
            </>
          )}
        </Spin>
      </Card>
    </div>
  );
};

export default UserFavorites;