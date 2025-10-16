import React, { useState, useEffect } from 'react';
import { Card, Input, List, Tag, Space, Typography, Row, Col, Button, Select, Checkbox, Divider, Empty, Tooltip, AutoComplete } from 'antd';
import { SearchOutlined, ApiOutlined, FilterOutlined, BookOutlined, LinkOutlined, ClockCircleOutlined, StarOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;

interface SearchResult {
  id: string;
  name: string;
  path: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  description: string;
  module: string;
  status: 'development' | 'testing' | 'production';
  version: string;
  tags: string[];
  developer: string;
  lastUpdated: string;
  popularity: number;
  responseTime: number;
}

interface SearchHistory {
  id: string;
  query: string;
  timestamp: string;
  resultCount: number;
}

const APISearch: React.FC = () => {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedMethods, setSelectedMethods] = useState<string[]>([]);
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);
  const [selectedModules, setSelectedModules] = useState<string[]>([]);
  const [sortBy, setSortBy] = useState<string>('relevance');
  const [searchHistory, setSearchHistory] = useState<SearchHistory[]>([]);
  const [showAdvancedFilter, setShowAdvancedFilter] = useState(false);

  // 模拟搜索数据
  const [allAPIs] = useState<SearchResult[]>([
    {
      id: '1',
      name: '用户登录',
      path: '/api/auth/login',
      method: 'POST',
      description: '用户登录接口，支持邮箱和用户名登录，返回JWT令牌用于后续认证',
      module: '用户管理',
      status: 'production',
      version: 'v1.0',
      tags: ['认证', '登录', '安全', 'JWT'],
      developer: '张三',
      lastUpdated: '2024-01-15',
      popularity: 95,
      responseTime: 120
    },
    {
      id: '2',
      name: '用户注册',
      path: '/api/auth/register',
      method: 'POST',
      description: '用户注册接口，需要邮箱验证，支持手机号和邮箱注册',
      module: '用户管理',
      status: 'production',
      version: 'v1.0',
      tags: ['认证', '注册', '邮箱验证'],
      developer: '李四',
      lastUpdated: '2024-01-14',
      popularity: 88,
      responseTime: 150
    },
    {
      id: '3',
      name: '获取用户信息',
      path: '/api/users/profile',
      method: 'GET',
      description: '获取当前用户的详细信息，包括基本资料、权限等',
      module: '用户管理',
      status: 'production',
      version: 'v1.0',
      tags: ['用户', '个人信息', '权限'],
      developer: '张三',
      lastUpdated: '2024-01-16',
      popularity: 92,
      responseTime: 80
    },
    {
      id: '4',
      name: '创建项目',
      path: '/api/projects',
      method: 'POST',
      description: '创建新的项目，支持项目模板和自定义配置',
      module: '项目管理',
      status: 'testing',
      version: 'v1.1',
      tags: ['项目', '创建', '模板'],
      developer: '王五',
      lastUpdated: '2024-01-16',
      popularity: 75,
      responseTime: 200
    },
    {
      id: '5',
      name: '多模态分析',
      path: '/api/ai/multimodal',
      method: 'POST',
      description: '多模态内容分析接口，支持文本、图像、音频的综合分析',
      module: 'AI服务',
      status: 'development',
      version: 'v2.0',
      tags: ['AI', '多模态', '分析', '机器学习'],
      developer: '赵六',
      lastUpdated: '2024-01-17',
      popularity: 45,
      responseTime: 500
    },
    {
      id: '6',
      name: '图像生成',
      path: '/api/ai/image/generate',
      method: 'POST',
      description: '基于文本描述生成高质量图像，支持多种风格和尺寸',
      module: 'AI服务',
      status: 'testing',
      version: 'v1.5',
      tags: ['AI', '图像生成', 'AIGC', '创意'],
      developer: '钱七',
      lastUpdated: '2024-01-16',
      popularity: 68,
      responseTime: 800
    }
  ]);

  // 搜索建议
  const [searchSuggestions, setSearchSuggestions] = useState<string[]>([]);

  useEffect(() => {
    // 生成搜索建议
    const suggestions = Array.from(new Set([
      ...allAPIs.map(api => api.name),
      ...allAPIs.flatMap(api => api.tags),
      ...allAPIs.map(api => api.module)
    ]));
    setSearchSuggestions(suggestions);
  }, [allAPIs]);

  const performSearch = (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }

    setLoading(true);
    
    // 模拟搜索延迟
    setTimeout(() => {
      const filtered = allAPIs.filter(api => {
        const searchText = query.toLowerCase();
        const matchesQuery = 
          api.name.toLowerCase().includes(searchText) ||
          api.description.toLowerCase().includes(searchText) ||
          api.path.toLowerCase().includes(searchText) ||
          api.tags.some(tag => tag.toLowerCase().includes(searchText)) ||
          api.module.toLowerCase().includes(searchText);

        const matchesMethod = selectedMethods.length === 0 || selectedMethods.includes(api.method);
        const matchesStatus = selectedStatuses.length === 0 || selectedStatuses.includes(api.status);
        const matchesModule = selectedModules.length === 0 || selectedModules.includes(api.module);

        return matchesQuery && matchesMethod && matchesStatus && matchesModule;
      });

      // 排序
      const sorted = filtered.sort((a, b) => {
        switch (sortBy) {
          case 'popularity':
            return b.popularity - a.popularity;
          case 'updated':
            return new Date(b.lastUpdated).getTime() - new Date(a.lastUpdated).getTime();
          case 'performance':
            return a.responseTime - b.responseTime;
          case 'name':
            return a.name.localeCompare(b.name);
          default: // relevance
            return b.popularity - a.popularity;
        }
      });

      setSearchResults(sorted);
      setLoading(false);

      // 添加到搜索历史
      if (query.trim()) {
        const newHistory: SearchHistory = {
          id: Date.now().toString(),
          query: query.trim(),
          timestamp: new Date().toLocaleString(),
          resultCount: sorted.length
        };
        setSearchHistory(prev => [newHistory, ...prev.slice(0, 9)]); // 保留最近10条
      }
    }, 300);
  };

  const handleSearch = (value: string) => {
    setSearchQuery(value);
    performSearch(value);
  };

  const getMethodColor = (method: string) => {
    const colors = {
      GET: 'green',
      POST: 'blue',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple'
    };
    return colors[method as keyof typeof colors] || 'default';
  };

  const getStatusColor = (status: string) => {
    const colors = {
      development: 'orange',
      testing: 'blue',
      production: 'green'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusText = (status: string) => {
    const texts = {
      development: '开发中',
      testing: '测试中',
      production: '已上线'
    };
    return texts[status as keyof typeof texts] || status;
  };

  const highlightText = (text: string, query: string) => {
    if (!query) return text;
    const regex = new RegExp(`(${query})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
  };

  const clearFilters = () => {
    setSelectedMethods([]);
    setSelectedStatuses([]);
    setSelectedModules([]);
  };

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <SearchOutlined /> 快速检索
      </Title>
      <Paragraph type="secondary">
        支持关键词搜索，快速定位所需接口，提供智能筛选和排序功能
      </Paragraph>

      {/* 搜索框 */}
      <Card style={{ marginBottom: '16px' }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={16}>
            <AutoComplete
              style={{ width: '100%' }}
              options={searchSuggestions.map(suggestion => ({ value: suggestion }))}
              filterOption={(inputValue, option) =>
                option!.value.toLowerCase().includes(inputValue.toLowerCase())
              }
            >
              <Search
                placeholder="搜索接口名称、描述、路径、标签或模块..."
                allowClear
                enterButton={<SearchOutlined />}
                size="large"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onSearch={handleSearch}
                loading={loading}
              />
            </AutoComplete>
          </Col>
          <Col xs={24} lg={8}>
            <Space>
              <Button 
                icon={<FilterOutlined />}
                onClick={() => setShowAdvancedFilter(!showAdvancedFilter)}
              >
                高级筛选
              </Button>
              <Select
                style={{ width: 120 }}
                value={sortBy}
                onChange={setSortBy}
              >
                <Select.Option value="relevance">相关性</Select.Option>
                <Select.Option value="popularity">热度</Select.Option>
                <Select.Option value="updated">更新时间</Select.Option>
                <Select.Option value="performance">性能</Select.Option>
                <Select.Option value="name">名称</Select.Option>
              </Select>
            </Space>
          </Col>
        </Row>

        {/* 高级筛选 */}
        {showAdvancedFilter && (
          <div style={{ marginTop: '16px', padding: '16px', backgroundColor: '#fafafa', borderRadius: '6px' }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={8}>
                <Text strong>请求方法：</Text>
                <div style={{ marginTop: '8px' }}>
                  <Checkbox.Group
                    value={selectedMethods}
                    onChange={setSelectedMethods}
                  >
                    <Space direction="vertical">
                      <Checkbox value="GET">GET</Checkbox>
                      <Checkbox value="POST">POST</Checkbox>
                      <Checkbox value="PUT">PUT</Checkbox>
                      <Checkbox value="DELETE">DELETE</Checkbox>
                      <Checkbox value="PATCH">PATCH</Checkbox>
                    </Space>
                  </Checkbox.Group>
                </div>
              </Col>
              <Col xs={24} sm={8}>
                <Text strong>接口状态：</Text>
                <div style={{ marginTop: '8px' }}>
                  <Checkbox.Group
                    value={selectedStatuses}
                    onChange={setSelectedStatuses}
                  >
                    <Space direction="vertical">
                      <Checkbox value="production">已上线</Checkbox>
                      <Checkbox value="testing">测试中</Checkbox>
                      <Checkbox value="development">开发中</Checkbox>
                    </Space>
                  </Checkbox.Group>
                </div>
              </Col>
              <Col xs={24} sm={8}>
                <Text strong>功能模块：</Text>
                <div style={{ marginTop: '8px' }}>
                  <Checkbox.Group
                    value={selectedModules}
                    onChange={setSelectedModules}
                  >
                    <Space direction="vertical">
                      <Checkbox value="用户管理">用户管理</Checkbox>
                      <Checkbox value="项目管理">项目管理</Checkbox>
                      <Checkbox value="AI服务">AI服务</Checkbox>
                    </Space>
                  </Checkbox.Group>
                </div>
              </Col>
            </Row>
            <div style={{ marginTop: '16px' }}>
              <Button size="small" onClick={clearFilters}>
                清除筛选
              </Button>
            </div>
          </div>
        )}
      </Card>

      <Row gutter={[16, 16]}>
        {/* 搜索结果 */}
        <Col xs={24} lg={18}>
          <Card 
            title={
              <Space>
                <Text>搜索结果</Text>
                {searchResults.length > 0 && (
                  <Tag color="blue">{searchResults.length} 个结果</Tag>
                )}
              </Space>
            }
          >
            {searchResults.length === 0 ? (
              searchQuery ? (
                <Empty 
                  description="未找到相关接口"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              ) : (
                <Empty 
                  description="请输入关键词开始搜索"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              )
            ) : (
              <List
                dataSource={searchResults}
                renderItem={(item) => (
                  <List.Item
                    actions={[
                      <Button type="link" icon={<BookOutlined />}>
                        文档
                      </Button>,
                      <Button type="link" icon={<LinkOutlined />}>
                        测试
                      </Button>
                    ]}
                  >
                    <List.Item.Meta
                      avatar={<ApiOutlined style={{ fontSize: '24px', color: '#1890ff' }} />}
                      title={
                        <Space>
                          <Text 
                            strong 
                            dangerouslySetInnerHTML={{ 
                              __html: highlightText(item.name, searchQuery) 
                            }}
                          />
                          <Tag color={getMethodColor(item.method)}>{item.method}</Tag>
                          <Tag color={getStatusColor(item.status)}>
                            {getStatusText(item.status)}
                          </Tag>
                          <Tag color="cyan">{item.version}</Tag>
                        </Space>
                      }
                      description={
                        <Space direction="vertical" size="small" style={{ width: '100%' }}>
                          <Text code>{item.path}</Text>
                          <Text 
                            dangerouslySetInnerHTML={{ 
                              __html: highlightText(item.description, searchQuery) 
                            }}
                          />
                          <div>
                            <Space wrap>
                              <Text type="secondary">模块：{item.module}</Text>
                              <Text type="secondary">开发者：{item.developer}</Text>
                              <Text type="secondary">响应时间：{item.responseTime}ms</Text>
                              <Tooltip title="接口热度">
                                <Space>
                                  <StarOutlined />
                                  <Text type="secondary">{item.popularity}</Text>
                                </Space>
                              </Tooltip>
                            </Space>
                          </div>
                          <div>
                            <Space wrap>
                              {item.tags.map(tag => (
                                <Tag 
                                  key={tag} 
                                  size="small"
                                  style={{
                                    backgroundColor: tag.toLowerCase().includes(searchQuery.toLowerCase()) ? '#fff2e8' : undefined,
                                    borderColor: tag.toLowerCase().includes(searchQuery.toLowerCase()) ? '#ffbb96' : undefined
                                  }}
                                >
                                  {tag}
                                </Tag>
                              ))}
                            </Space>
                          </div>
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Col>

        {/* 搜索历史 */}
        <Col xs={24} lg={6}>
          <Card title="搜索历史" size="small">
            {searchHistory.length === 0 ? (
              <Empty 
                description="暂无搜索历史"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                style={{ margin: '20px 0' }}
              />
            ) : (
              <List
                size="small"
                dataSource={searchHistory}
                renderItem={(item) => (
                  <List.Item
                    style={{ cursor: 'pointer' }}
                    onClick={() => handleSearch(item.query)}
                  >
                    <List.Item.Meta
                      avatar={<ClockCircleOutlined />}
                      title={<Text ellipsis>{item.query}</Text>}
                      description={
                        <Space>
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            {item.resultCount} 个结果
                          </Text>
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            {item.timestamp.split(' ')[1]}
                          </Text>
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default APISearch;